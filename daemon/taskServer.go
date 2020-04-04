package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"runtime"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/cgroups/systemd"
	"github.com/opencontainers/runc/libcontainer/configs"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/segmentio/ksuid"
	"github.com/xhebox/chrootd/api"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()

		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			os.Exit(1)
		}
	}
}

type taskServer struct {
	api.UnimplementedTaskServer
	user    *User
	cntrs   *cntrPool
	tasks   *taskPool
	factory libcontainer.Factory
}

func newTaskServer(u *User, p *cntrPool, t *taskPool) (*taskServer, error) {
	cgroupMgr := libcontainer.Cgroupfs
	if systemd.UseSystemd() {
		cgroupMgr = libcontainer.SystemdCgroups
	}
	// TODO: rootless should be an option
	if true {
		cgroupMgr = libcontainer.RootlessCgroupfs
	}

	factory, err := libcontainer.New(u.RunPath, cgroupMgr, libcontainer.InitArgs(os.Args[0], "init"),
		// without suid/guid or corressponding caps, extern mapping tools are needed to run rootless(with correct configuration)
		// TODO: try to find binary by PATH, and then fallback
		libcontainer.NewuidmapPath("/bin/newuidmap"),
		libcontainer.NewgidmapPath("/bin/newgidmap"))
	if err != nil {
		return nil, err
	}

	return &taskServer{user: u, cntrs: p, tasks: t, factory: factory}, nil
}

func (s *taskServer) Close() {
	s.tasks.Range(func(key ksuid.KSUID, t *task) bool {
		t.Destroy()
		return true
	})
}

func (s *taskServer) Start(c context.Context, req *api.StartReq) (*api.StartRes, error) {
	s.user.Logger.Info().Msgf("request to start task[%p]", req)

	meta, err := s.cntrs.StartMeta(req.CntrId)
	if err != nil {
		return &api.StartRes{Id: nil, Reason: "no container of such id"}, nil
	}

	task := newTask(meta)

	tid := s.tasks.Add(task)
	if tid == ksuid.Nil {
		return &api.StartRes{Id: nil, Reason: "can not alloc uid"}, nil
	}

	cfg := &meta.Config

	// TODO: not all options can be passed from cli
	// need an permission filter

	// allow rootless
	cfg.RootlessEUID = true
	cfg.RootlessCgroups = true
	cfg.UidMappings = []configs.IDMap{
		{
			ContainerID: 0,
			HostID:      os.Getuid(),
			Size:        65536,
		},
	}
	cfg.GidMappings = []configs.IDMap{
		{
			ContainerID: 0,
			HostID:      os.Getgid(),
			Size:        65536,
		},
	}

	if cfg.Rootfs == "/" {
		// if you are mounting the root of host
		// you surely do not want it to be writable
		cfg.Readonlyfs = true
	}

	if err := task.Init(s.factory, cfg); err != nil {
		s.tasks.Del(tid)
		s.user.Logger.Warn().Err(err).Msgf("fail to create task[%p]", req)
		return &api.StartRes{Id: nil, Reason: "can not alloc resource for container"}, nil
	}

	uid, _ := cfg.HostRootUID()
	gid, _ := cfg.HostRootGID()
	s.user.Logger.Debug().Msgf("start task[%p]: %v, %v, %v", req, uid, gid, meta)
	return &api.StartRes{Id: tid.Bytes()}, nil
}

func (s *taskServer) Stop(c context.Context, req *api.StopReq) (*api.StopRes, error) {
	s.user.Logger.Info().Msgf("request to stop task[%p] %v", req, req)

	id, err := ksuid.FromBytes(req.Id)
	if err != nil {
		return &api.StopRes{Reason: "invalid id"}, nil
	}

	task := s.tasks.Get(id)
	if task == nil {
		return &api.StopRes{Reason: "no container of such id"}, nil
	}

	if err := task.Destroy(); err != nil {
		if err.Error() == libcontainer.SystemError.String() {
			return nil, err
		}

		return &api.StopRes{Reason: err.Error()}, nil
	}

	s.cntrs.StopMeta(task.meta.Id)
	s.tasks.Del(id)

	return &api.StopRes{}, nil
}

func (s *taskServer) List(req *api.ListReq, srv api.Task_ListServer) error {
	s.user.Logger.Info().Msgf("request to list tasks[%p]", req)

	var err error

	s.tasks.Range(func(key ksuid.KSUID, t *task) bool {
		r := true

		for _, filter := range req.Filters {
			switch filter.Key {
			// TODO: more filters
			default:
				r = false
				return false
			}
		}

		if r {
			if e := srv.Send(&api.ListRes{Id: key.Bytes()}); e != nil {
				err = e
				return false
			}
		}

		return true
	})

	return err
}

func infoFromProc(proc *libcontainer.Process) *api.Proc {
	r := &api.Proc{
		Args:          proc.Args,
		Env:           proc.Env,
		User:          proc.User,
		Groups:        proc.AdditionalGroups,
		Cwd:           proc.Cwd,
		ConsoleWidth:  uint32(proc.ConsoleWidth),
		ConsoleHeight: uint32(proc.ConsoleHeight),
		// handle rlimits/caps/io
	}

	id, e := proc.Pid()
	if e != nil {
		r.Pid = -1
	}
	r.Pid = int64(id)

	return r
}

func (s *taskServer) ListProc(req *api.ListProcReq, srv api.Task_ListProcServer) error {
	s.user.Logger.Info().Msgf("request to list task process[%p]", req)

	var err error

	id, err := ksuid.FromBytes(req.Id)
	if err != nil {
		return srv.Send(&api.ListProcRes{Reason: "invalid id"})
	}

	task := s.tasks.Get(id)
	if task == nil {
		return srv.Send(&api.ListProcRes{Reason: "no task of such id"})
	}

	task.RangeProc(func(proc *process) bool {
		if e := srv.Send(&api.ListProcRes{
			Info: infoFromProc(proc.Process),
		}); e != nil {
			err = e
			return false
		}

		return true
	})

	return err
}

func (s *taskServer) Exec(ctx context.Context, req *api.ExecReq) (*api.ExecRes, error) {
	s.user.Logger.Info().Msgf("request to run process[%p]", req)

	id, err := ksuid.FromBytes(req.Id)
	if err != nil {
		return &api.ExecRes{Reason: "invalid id"}, nil
	}

	task := s.tasks.Get(id)
	if task == nil {
		return &api.ExecRes{Reason: "no task of such id"}, nil
	}

	info := req.Prog

	proc := &libcontainer.Process{
		Args:             info.Args,
		Env:              info.Env,
		User:             info.User,
		AdditionalGroups: info.Groups,
		Cwd:              info.Cwd,
		ConsoleWidth:     uint16(info.ConsoleWidth),
		ConsoleHeight:    uint16(info.ConsoleHeight),
		Init:             task.isInit,
		// TODO: handle rlimits/caps
	}

	if err := task.Exec(proc, req.Attach); err != nil {
		s.user.Logger.Warn().Err(err).Msgf("fail to run process[%p]", req)
		return nil, err
	}

	if task.isInit {
		task.isInit = false
	}

	return &api.ExecRes{Info: infoFromProc(proc)}, nil
}

func (s *taskServer) IO(srv api.Task_IOServer) error {
	s.user.Logger.Debug().Msgf("request to io[%p]", srv)

	first, err := srv.Recv()
	if err != nil {
		return err
	}

	id, err := ksuid.FromBytes(first.Id)
	if err != nil {
		return srv.Send(&api.IORes{Str: "invalid id"})
	}

	task := s.tasks.Get(id)
	if task == nil {
		return srv.Send(&api.IORes{Str: "no task of such id"})
	}

	pid := first.Pid
	if pid <= 0 {
		return srv.Send(&api.IORes{Str: "no process of such pid"})
	}

	var p *process

	task.RangeProc(func(proc *process) bool {
		if pid == proc.pid {
			p = proc
			return false
		}
		return true
	})

	if p == nil {
		return srv.Send(&api.IORes{Str: "no process of such pid"})
	}

	err = srv.Send(&api.IORes{Str: "handshaked"})
	if err != nil {
		return err
	}

	handle := func(file *os.File, cat string, srv api.Task_IOServer) error {
		buf := make([]byte, 256)
	loop:
		for {
			select {
			case <-srv.Context().Done():
				break loop
			default:
				n, e := file.Read(buf)
				if e != nil {
					if e == io.EOF {
						break loop
					}
					return e
				}

				e = srv.Send(&api.IORes{
					Str: cat,
					D:   buf[:n],
				})
				if e != nil {
					return e
				}
			}
		}
		return nil
	}

	go handle(p.stdout, "stdout", srv)
	go handle(p.stderr, "stderr", srv)

	// TODO: should terminate when process is terminated
loop:
	for {
		select {
		case <-srv.Context().Done():
			break loop
		default:
			pkt, err := srv.Recv()
			if err != nil {
				if err == io.EOF {
					break loop
				}
				return err
			}

			data := pkt.D
			if data == nil {
				return srv.Send(&api.IORes{Str: "empty data"})
			}

			if _, err := io.Copy(p.stdin, bytes.NewReader(data)); err != nil {
				return nil
			}
		}
	}

	return nil
}
