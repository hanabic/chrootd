package cntr

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"github.com/docker/libkv/store"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/cgroups/systemd"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/ryanuber/go-glob"
	"github.com/segmentio/ksuid"
	"github.com/smallnest/rpcx/server"
	"github.com/xhebox/chrootd/registry"
)

func init() {
	if len(os.Args) == 2 && os.Args[1] == "___init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()

		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			os.Exit(1)
		}
	}
}

type Attach struct {
	Prog       *libcontainer.Process
	Inr, Inw   *os.File
	Outr, Outw *os.File
}

func (pty *Attach) Close() {
	pty.Inr.Close()
	pty.Inw.Close()
	pty.Outr.Close()
	pty.Outw.Close()
}

type Server struct {
	registered  bool
	serviceAddr string
	attachAddr  string
	cntrPath    string
	registry    registry.Registry
	factory     libcontainer.Factory
	states      store.Store

	procs   map[string]*Attach
	procsMu sync.Mutex

	Context    context.Context
	ProcLimits int
	Rootless   bool
	// Apply default secure config automatically
	Secure bool
}

type Opt func(*Server) error

func NewServer(cntrPath string, serviceAddr string, attachAddr string, states store.Store, registry registry.Registry, opts ...Opt) (*Server, error) {
	res := &Server{
		registered:  false,
		cntrPath:    cntrPath,
		registry:    registry,
		states:      states,
		serviceAddr: serviceAddr,
		attachAddr:  attachAddr,
		procs:       make(map[string]*Attach),
		ProcLimits:  64,
		Rootless:    true,
		Secure:      true,
		Context:     context.Background(),
	}

	for _, f := range opts {
		if err := f(res); err != nil {
			return nil, err
		}
	}

	cgroupMgr := libcontainer.Cgroupfs
	if systemd.UseSystemd() {
		cgroupMgr = libcontainer.SystemdCgroups
	}
	if res.Rootless {
		cgroupMgr = libcontainer.RootlessCgroupfs
	}

	uidPath, err := exec.LookPath("newuidmap")
	if err != nil {
		uidPath = "/bin/newuidmap"
	}

	gidPath, err := exec.LookPath("newgidmap")
	if err != nil {
		gidPath = "/bin/newgidmap"
	}

	err = os.RemoveAll(filepath.Join(cntrPath, "factory"))
	if err != nil {
		return nil, err
	}

	res.factory, err = libcontainer.New(cntrPath,
		cgroupMgr,
		libcontainer.InitArgs(os.Args[0], "___init"),
		// without suid/guid or corressponding caps, extern mapping tools are needed to run rootless(with correct configuration)
		libcontainer.NewuidmapPath(uidPath),
		libcontainer.NewgidmapPath(gidPath),
	)
	if err != nil {
		return nil, err
	}

	ok, _ := res.states.Exists("id")
	if !ok {
		err = res.states.Put("id", []byte(ksuid.New().String()), &store.WriteOptions{})
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (s *Server) Register(rpcx *server.Server, servicePath string) error {
	var err error

	if s.registered {
		return fmt.Errorf("registered")
	}
	s.registered = true

	// metadata related rpc
	err = rpcx.RegisterFunctionName(servicePath, "List", s.List, "")
	if err != nil {
		return err
	}

	err = rpcx.RegisterFunctionName(servicePath, "Create", s.Create, "")
	if err != nil {
		return err
	}

	err = rpcx.RegisterFunctionName(servicePath, "Config", s.Config, "")
	if err != nil {
		return err
	}

	err = rpcx.RegisterFunctionName(servicePath, "Delete", s.Delete, "")
	if err != nil {
		return err
	}

	// runtime container related rpc
	err = rpcx.RegisterFunctionName(servicePath, "Start", s.Start, "")
	if err != nil {
		return err
	}

	err = rpcx.RegisterFunctionName(servicePath, "Exec", s.Exec, "")
	if err != nil {
		return err
	}

	err = rpcx.RegisterFunctionName(servicePath, "Pause", s.Pause, "")
	if err != nil {
		return err
	}

	err = rpcx.RegisterFunctionName(servicePath, "Resume", s.Resume, "")
	if err != nil {
		return err
	}

	err = rpcx.RegisterFunctionName(servicePath, "Stop", s.Stop, "")
	if err != nil {
		return err
	}

	err = rpcx.RegisterFunctionName(servicePath, "Status", s.Status, "")
	if err != nil {
		return err
	}

	if s.registry != nil {
		kvs, err := s.states.Get("id")
		if err != nil {
			return err
		}

		err = s.registry.Put(string(kvs.Value), []byte(s.serviceAddr))
		if err != nil {
			return err
		}
	}

	return nil
}

type CreateReq struct {
	Meta Metainfo
}

type CreateRes struct {
	Id   string
	Meta Metainfo
}

func (s *Server) Create(ctx context.Context, req *CreateReq, res *CreateRes) (err error) {
	id := ksuid.New().String()

	meta := req.Meta

	if s.Secure {
		err = meta.Default()
		if err != nil {
			return err
		}
		meta.Cgroups.Name = id
	}

	mbytes, err := meta.ToBytes()
	if err != nil {
		return err
	}

	ok, _, err := s.states.AtomicPut(id, mbytes, nil, &store.WriteOptions{})
	if !ok {
		return err
	}

	res.Id = id
	res.Meta = meta

	return nil
}

type ConfigReq struct {
	Id   string `json:"cntrid"`
	Meta Metainfo
}

type ConfigRes struct {
	Meta Metainfo
}

func (s *Server) Config(ctx context.Context, req *ConfigReq, res *ConfigRes) error {
	kvs, err := s.states.Get(req.Id)
	if err != nil {
		return err
	}

	meta, err := NewMetaFromBytes(kvs.Value)
	if err != nil {
		return err
	}

	err = meta.Merge(&req.Meta)
	if err != nil {
		return err
	}

	mbytes, err := meta.ToBytes()
	if err != nil {
		return err
	}

	ok, _, err := s.states.AtomicPut(req.Id, mbytes, kvs, &store.WriteOptions{})
	if !ok {
		return err
	}

	res.Meta = *meta

	return nil
}

type DeleteReq struct {
	Id string `json:"cntrid"`
}

type DeleteRes struct {
}

func (s *Server) Delete(ctx context.Context, req *DeleteReq, res *DeleteRes) error {
	cntr, err := s.factory.Load(req.Id)
	if err == nil {
		err = cntr.Destroy()
		if err != nil {
			return err
		}
	}

	err = s.states.DeleteTree(req.Id)
	if err != nil {
		return fmt.Errorf("container not exist")
	}

	return nil
}

type ListFilter struct {
	Key string
	Val string
}

type ListReq struct {
	Filters []ListFilter
}

type ListCntr struct {
	Id   string
	Addr string
}

type ListRes struct {
	CntrIds []ListCntr
}

func (s *Server) List(ctx context.Context, req *ListReq, res *ListRes) error {
	cntrs, err := s.states.List("")
	if err != nil {
		return err
	}

	for _, cntr := range cntrs {
		meta, err := NewMetaFromBytes(cntr.Value)
		if err != nil {
			return err
		}

		r := true

		for _, filter := range req.Filters {
			switch filter.Key {
			case "hostname":
				r = glob.Glob(filter.Val, meta.Hostname)
			}
			if !r {
				break
			}
		}

		if r {
			res.CntrIds = append(res.CntrIds, ListCntr{
				Id:   cntr.Key,
				Addr: s.serviceAddr,
			})
		}
	}

	return nil
}

type ExecReq struct {
	Id string `json:"cntrid"`
}

type ExecRes struct {
}

func (s *Server) Exec(ctx context.Context, req *ExecReq, res *ExecRes) error {
	cntr, err := s.factory.Load(req.Id)
	if err != nil {
		return err
	}

	return cntr.Exec()
}

type PauseReq struct {
	Id string `json:"cntrid"`
}

type PauseRes struct {
}

func (s *Server) Pause(ctx context.Context, req *PauseReq, res *PauseRes) error {
	cntr, err := s.factory.Load(req.Id)
	if err != nil {
		return err
	}

	return cntr.Pause()
}

type ResumeReq struct {
	Id string `json:"cntrid"`
}

type ResumeRes struct {
}

func (s *Server) Resume(ctx context.Context, req *ResumeReq, res *ResumeRes) error {
	cntr, err := s.factory.Load(req.Id)
	if err != nil {
		return err
	}

	return cntr.Resume()
}

type StartReq struct {
	Id string
	// run immd
	Attach bool
	Prog   Proc
}

type StartRes struct {
	AttachId   string
	AttachAddr string
}

func (s *Server) Start(ctx context.Context, req *StartReq, res *StartRes) error {
	cntr, err := s.factory.Load(req.Id)
	if err != nil {
		E, ok := err.(libcontainer.Error)
		if !ok {
			return err
		}

		if E.Code() != libcontainer.ContainerNotExists {
			return err
		}

		id, err := ksuid.Parse(req.Id)
		if err != nil {
			return err
		}

		m, err := s.states.Get(req.Id)
		if err != nil {
			return err
		}

		meta, err := NewMetaFromBytes(m.Value)
		if err != nil {
			return err
		}

		cfg := meta.ToConfig()

		cfg.RootlessEUID = s.Rootless
		cfg.RootlessCgroups = s.Rootless

		cntr, err = s.factory.Create(id.String(), &cfg)
		if err != nil {
			return err
		}
	}

	var st libcontainer.Status

	st, err = cntr.Status()
	if err != nil {
		return err
	}

	prog := req.Prog.ToProc()

	prog.Init = st == libcontainer.Stopped || st == libcontainer.Created

	if req.Attach {
		pty := &Attach{}
		pid := ksuid.New().String()

		pty.Inr, pty.Inw, err = os.Pipe()
		if err != nil {
			return err
		}

		pty.Outr, pty.Outw, err = os.Pipe()
		if err != nil {
			return err
		}

		prog.Stdin = pty.Inr
		prog.Stdout = pty.Outw
		prog.Stderr = pty.Outw

		res.AttachId = pid
		res.AttachAddr = s.attachAddr

		err = cntr.Start(prog)
		if err != nil {
			return err
		}

		s.procsMu.Lock()
		if len(s.procs) >= s.ProcLimits {
			s.procsMu.Unlock()
			return fmt.Errorf("can not attach it")
		}
		s.procs[pid] = pty
		s.procsMu.Unlock()

		go func() {
			prog.Wait()
			s.procsMu.Lock()
			delete(s.procs, pid)
			s.procsMu.Unlock()
			pty.Close()
		}()
	} else {
		err = cntr.Start(prog)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) ServeAttach(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	rd := bufio.NewReader(conn)

	id, ok, err := rd.ReadLine()
	if err != nil || ok {
		return
	}

	s.procsMu.Lock()
	c, ok := s.procs[string(id)]
	if !ok {
		s.procsMu.Unlock()
		return
	}
	s.procsMu.Unlock()

	ch := make(chan bool, 2)

	go func() {
		io.Copy(c.Inw, rd)
		c.Close()
		ch <- true
	}()
	go func() {
		io.Copy(conn, c.Outr)
		conn.Close()
		ch <- true
	}()

	// no leaking goroutines
	for i := 0; i < 2; i++ {
		select {
		case <-ctx.Done():
			c.Prog.Signal(syscall.SIGKILL)
			c.Prog.Wait()
			return
		case <-ch:
		}
	}
}

type StopReq struct {
	Id string
}

type StopRes struct {
}

func (s *Server) stop(id string) error {
	cntr, err := s.factory.Load(id)
	if err == nil {
		st, err := cntr.Status()
		if err != nil {
			return err
		}

		switch st {
		case libcontainer.Paused:
			err = cntr.Resume()
			if err != nil {
				return err
			}
			fallthrough
		case libcontainer.Running:
			err = cntr.Signal(syscall.SIGKILL, true)
			if err != nil {
				return err
			}
		case libcontainer.Pausing:
			return fmt.Errorf("container is pausing, try later")
		}

		return cntr.Destroy()
	}

	return nil
}

func (s *Server) Stop(ctx context.Context, req *StopReq, res *StopRes) error {
	return s.stop(req.Id)
}

type StatusReq struct {
	Id string
}

type StatusRes struct {
	Status     libcontainer.Status
	Statistics *libcontainer.Stats
}

func (s *Server) Status(ctx context.Context, req *StatusReq, res *StatusRes) error {
	cntr, err := s.factory.Load(req.Id)
	if err != nil {
		return err
	}

	s1, err := cntr.Status()
	if err != nil {
		return err
	}
	res.Status = s1

	s2, err := cntr.Stats()
	if err != nil {
		return err
	}
	res.Statistics = s2

	return nil
}

func (s *Server) Close() error {
	var err error

	cntrs, err := s.states.List("")
	if err != nil {
		return err
	}

	for _, cntr := range cntrs {
		if cntr.Key == "id" {
			continue
		}
		e := s.stop(cntr.Key)
		if e != nil {
			err = e
		}
	}

	return err
}
