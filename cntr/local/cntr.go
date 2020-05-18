package local

import (
	"bytes"
	"fmt"
	"syscall"
	"os"
	"sync"
	"time"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/pkg/errors"
	. "github.com/xhebox/chrootd/cntr"
	mtyp "github.com/xhebox/chrootd/meta"
)

type task struct {
	*libcontainer.Process
	mu       sync.Mutex
	Inr, Inw *os.File
	Out      *bytes.Buffer
	closed   bool
}

func (t *task) Read(buf []byte) (int, error) {
	if !t.closed && t.Out.Len() == 0 {
		time.Sleep(100 * time.Millisecond)
		return 0, nil
	}
	return t.Out.Read(buf)
}

func (t *task) Write(buf []byte) (n int, err error) {
	tbuf := bytes.NewBuffer(nil)
	tbuf.Grow(len(buf))
	for _, b := range buf {
		switch b {
		case 0x03:
			t.Signal(syscall.SIGTERM)
			return t.Inw.Write(tbuf.Bytes())
		default:
			err = tbuf.WriteByte(b)
			if err != nil {
				return
			}
		}
	}
	return t.Inw.Write(tbuf.Bytes())
}

func (t *task) CloseWrite() error {
	t.Inw.Close()
	t.Inr.Close()
	return nil
}

func (t *task) Close() error {
	t.closed = true
	t.Inr.Close()
	t.Inw.Close()
	return nil
}

type cntr struct {
	meta  *mtyp.Metainfo
	cntr  libcontainer.Container
	tasks map[string]*task
	rwmux sync.RWMutex
	wg    sync.WaitGroup
}

func newCntr(c libcontainer.Container, meta *mtyp.Metainfo) *cntr {
	return &cntr{
		meta:  meta,
		cntr:  c,
		tasks: make(map[string]*task),
	}
}

func (c *cntr) getTask(id string) (*task, bool) {
	c.rwmux.RLock()
	t, ok := c.tasks[id]
	c.rwmux.RUnlock()
	return t, ok
}

func (c *cntr) Meta() (*mtyp.Metainfo, error) {
	return c.meta, nil
}

func (c *cntr) Start(rt *Taskinfo) (string, error) {
	if len(rt.Args) == 0 {
		return "", errors.New("empty args, should have at least one argument")
	}

	t := &task{
		Process: &libcontainer.Process{
			Cwd:           "/",
			Args:          rt.Args,
			Env:           rt.Env,
			Capabilities:  &rt.Capabilities,
			Rlimits:       spec2runcRlimits(rt.Rlimits),
			Init:          len(c.tasks) == 0,
			ConsoleHeight: rt.TermHeight,
			ConsoleWidth:  rt.TermWidth,
		},
		Out: bytes.NewBufferString(""),
	}

	var err error
	t.Inr, t.Inw, err = os.Pipe()
	if err != nil {
		return "", err
	}
	t.Process.Stdin = t.Inr
	t.Process.Stdout = t.Out
	t.Process.Stderr = t.Out

	err = c.cntr.Run(t.Process)
	if err != nil {
		return "", err
	}

	pid, _ := t.Pid()
	id := fmt.Sprint(pid)

	c.rwmux.Lock()
	oldt, ok := c.tasks[id]
	c.tasks[id] = t
	c.rwmux.Unlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		t.Wait()

		c.rwmux.Lock()
		t, ok := c.tasks[id]
		if ok {
			t.Close()
		}
		delete(c.tasks, id)
		c.rwmux.Unlock()
	}()

	if ok {
		oldt.Close()
	}

	return id, nil
}

func (c *cntr) Stop(id string, kill bool) error {
	sig := syscall.SIGTERM
	if kill {
		sig = syscall.SIGKILL
	}

	t, ok := c.getTask(id)
	if ok {
		t.Close()
		return t.Signal(sig)
	}
	return nil
}

func (c *cntr) StopAll(kill bool) error {
	sig := syscall.SIGTERM
	if kill {
		sig = syscall.SIGKILL
	}
	c.cntr.Signal(sig, true)

	c.rwmux.RLock()

	var err error
	for _, t := range c.tasks {
		err = t.Close()
	}

	c.rwmux.RUnlock()

	return err
}

func (c *cntr) Attach(id string) (Attacher, error) {
	t, ok := c.getTask(id)
	if !ok {
		return nil, errors.New("can not find task")
	}

	return t, nil
}

func (c *cntr) List(f func(string) error) error {
	c.rwmux.Lock()
	defer c.rwmux.Unlock()
	for k := range c.tasks {
		if err := f(k); err != nil {
			return err
		}
	}
	return nil
}

func (c *cntr) Wait() error {
	c.wg.Wait()
	return nil
}

func (c *cntr) Destroy() error {
	c.StopAll(true)

	c.wg.Wait()

	return c.cntr.Destroy()
}
