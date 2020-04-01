package main

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/segmentio/ksuid"
	"github.com/xhebox/chrootd/api"
)

type process struct {
	*libcontainer.Process
	stdin  *os.File
	stdout *os.File
	stderr *os.File
	pid    int64
}

// TODO: should clear dead process by some interval
type task struct {
	sync.Mutex
	id     ksuid.KSUID
	meta   *api.Metainfo
	cntr   libcontainer.Container
	procs  []*process
	isInit bool
}

func newTask(meta *api.Metainfo) *task {
	return &task{id: ksuid.Nil, meta: meta, isInit: false}
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

	if attach {
		p.Stdin, proc.stdin, err = os.Pipe()
		if err != nil {
			return err
		}

		proc.stdout, p.Stdout, err = os.Pipe()
		if err != nil {
			return err
		}

		proc.stderr, p.Stderr, err = os.Pipe()
		if err != nil {
			return err
		}
	} else {
		p.Stdin = strings.NewReader("")
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

	t.procs = append(t.procs, proc)

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

	for _, proc := range t.procs {
		_, err := proc.Wait()
		if err != nil {
			return err
		}
		proc.stdin.Close()
		proc.stdout.Close()
		proc.stderr.Close()
	}

	return t.cntr.Destroy()
}
