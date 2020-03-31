package main

import (
	"sync"

	"github.com/segmentio/ksuid"
)

type taskPool struct {
	sync.Mutex
	pool map[ksuid.KSUID]*task
}

func newTaskPool() *taskPool {
	return &taskPool{
		pool: make(map[ksuid.KSUID]*task),
	}
}

func (p *taskPool) Add(o *task) ksuid.KSUID {
	p.Lock()
	defer p.Unlock()

	if o == nil {
		return ksuid.Nil
	}

	for i := 0; i < 3; i++ {
		newid := ksuid.New()

		_, ok := p.pool[newid]
		if ok {
			continue
		}

		o.SetId(newid)
		p.pool[newid] = o
		return newid
	}

	return ksuid.Nil
}

func (p *taskPool) Get(id ksuid.KSUID) *task {
	return p.pool[id]
}

func (p *taskPool) Del(id ksuid.KSUID) {
	p.Lock()
	defer p.Unlock()

	delete(p.pool, id)
}

func (p *taskPool) Range(f func(key ksuid.KSUID, val *task) bool) {
	for k, v := range p.pool {
		if !f(k, v) {
			break
		}
	}
}
