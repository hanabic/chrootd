package main

import (
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

	cfg := meta.Config

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

	if err := task.Init(s.factory, &cfg); err != nil {
		s.tasks.Del(tid)
		s.user.Logger.Warn().Err(err).Msgf("fail to create task[%p]", req)
		return &api.StartRes{Id: nil, Reason: "can not alloc resource for container"}, nil
	}

	uid, _ := cfg.HostRootUID()
	gid, _ := cfg.HostRootGID()
	s.user.Logger.Debug().Msgf("start task[%p]: %v, %v, %+v", req, uid, gid, meta)
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
		// TODO: handle rlimits/caps
	}

	if err := task.Exec(proc, req.Attach); err != nil {
		s.user.Logger.Warn().Err(err).Msgf("fail to run process[%p]", req)
		return nil, err
	}

	return &api.ExecRes{Info: infoFromProc(proc)}, nil
}

func (s *taskServer) Read(ctx context.Context, req *api.ReadReq) (*api.ReadRes, error) {
	s.user.Logger.Debug().Msgf("request to read[%p]: %+v", req, req)

	id, err := ksuid.FromBytes(req.Id)
	if err != nil {
		return &api.ReadRes{Reason: "invalid task id"}, nil
	}

	task := s.tasks.Get(id)
	if task == nil {
		return &api.ReadRes{Reason: "no task of such id"}, nil
	}

	p := task.procs[req.Pid]
	if p == nil {
		return &api.ReadRes{Reason: "no process of such pid"}, nil
	}

	d := make([]byte, 1024)
	n := 0

	switch req.Type {
	case "stdout":
		n, err = p.rstdout.Read(d)
		if err == io.EOF {
			p.rstdout.Close()
		}
	case "stderr":
		n, err = p.rstderr.Read(d)
		if err == io.EOF {
			p.rstderr.Close()
		}
	}

	s.user.Logger.Debug().Msgf("request to read[%p] ---- %d, %s", req, n, err)
	if err != nil {
		if err == io.EOF {
			return &api.ReadRes{Reason: "eof"}, nil
		}
		// TODO: log it
		return &api.ReadRes{Reason: "unknown error"}, nil
	}

	return &api.ReadRes{D: d[:n]}, nil
}

func (s *taskServer) Write(ctx context.Context, req *api.WriteReq) (*api.WriteRes, error) {
	s.user.Logger.Debug().Msgf("request to write[%p]", req)

	id, err := ksuid.FromBytes(req.Id)
	if err != nil {
		return &api.WriteRes{Reason: "invalid task id"}, nil
	}

	task := s.tasks.Get(id)
	if task == nil {
		return &api.WriteRes{Reason: "no task of such id"}, nil
	}

	p := task.procs[req.Pid]
	if p == nil {
		return &api.WriteRes{Reason: "no process of such pid"}, nil
	}

	for k := 0; k < len(req.D); {
		n, err := p.wstdin.Write(req.D[k:])
		k += n
		if err != nil {
			if err == io.ErrClosedPipe {
				p.rstdin.Close()
			}
			return &api.WriteRes{Reason: "write error"}, nil
		}
	}

	return &api.WriteRes{}, nil
}
