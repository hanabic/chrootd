package main

import (
	"errors"
	"fmt"

	//"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/segmentio/ksuid"
	"github.com/xhebox/chrootd/api"
)

type process struct {
	*libcontainer.Process
	rstdin  *os.File
	wstdin  *os.File
	rstdout *os.File
	wstdout *os.File
	rstderr *os.File
	wstderr *os.File
	pid     int64
}

// TODO: should clear dead process by some interval
type task struct {
	sync.Mutex
	id    ksuid.KSUID
	meta  *api.Metainfo
	cntr  libcontainer.Container
	procs map[int64]*process
}

func newTask(meta *api.Metainfo) *task {
	return &task{id: ksuid.Nil, meta: meta, procs: make(map[int64]*process)}
}

func (t *task) SetId(id ksuid.KSUID) {
	t.Lock()
	defer t.Unlock()

	if t.id != ksuid.Nil {
		return
	}

	t.id = id
}

func (t *task) Init(factory libcontainer.Factory, config *configs.Config) (err error) {
	t.Lock()
	defer t.Unlock()

	if t.id == ksuid.Nil {
		return errors.New("id empty, something goes wrong")
	}

	t.cntr, err = factory.Create(t.id.String(), config)
	return
}

func (t *task) Exec(p *libcontainer.Process, attach bool) error {
	t.Lock()
	defer t.Unlock()

	proc := &process{
		Process: p,
	}

	var err error

	status, err := t.cntr.Status()
	if err != nil {
		return err
	}

	p.Init = status == libcontainer.Stopped || status == libcontainer.Created

	if attach {
		proc.rstdin, proc.wstdin, err = os.Pipe()
		if err != nil {
			return err
		}
		p.Stdin = proc.rstdin

		proc.rstdout, proc.wstdout, err = os.Pipe()
		if err != nil {
			return err
		}
		p.Stdout = proc.wstdout

		proc.rstderr, proc.wstderr, err = os.Pipe()
		if err != nil {
			return err
		}
		p.Stderr = proc.wstderr
	} else {
		p.Stdin = os.Stdin
		p.Stdout = os.Stdout
		p.Stderr = os.Stderr
	}

	if err := t.cntr.Run(p); err != nil {
		return err
	}

	pid, err := p.Pid()
	if err != nil {
		return err
	}
	proc.pid = int64(pid)

	t.procs[proc.pid] = proc

	go func() {
		// TODO: handle exit status
		e, err := proc.Wait()
		fmt.Println(e, err)
		if proc.rstdin != nil {
			proc.rstdin.Close()
			proc.wstdout.Close()
			proc.wstderr.Close()
			proc.wstdin.Close()
			proc.wstdout.Close()
			proc.wstderr.Close()
		}
		delete(t.procs, proc.pid)
	}()

	return nil
}

func (t *task) RangeProc(f func(*process) bool) {
	t.Lock()
	defer t.Unlock()

	for _, proc := range t.procs {
		if !f(proc) {
			break
		}
	}
}

func (t *task) Destroy() error {
	t.Lock()
	defer t.Unlock()

	t.cntr.Signal(syscall.SIGTERM, true)

	for _, proc := range t.procs {
		_, err := proc.Wait()
		if err != nil {
			// TODO: log it
			_ = err
		}
		if proc.rstdin != nil {
			proc.rstdin.Close()
			proc.rstdout.Close()
			proc.rstderr.Close()
			proc.wstdin.Close()
			proc.wstdout.Close()
			proc.wstderr.Close()
		}
	}

	return t.cntr.Destroy()
}
