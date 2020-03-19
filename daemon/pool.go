package main

import (
	"sync"

	"github.com/segmentio/ksuid"
)

type poolobj interface {
	SetId(ksuid.KSUID)
}

type pool struct {
	sync.Mutex
	p map[ksuid.KSUID]poolobj
}

func newPool() *pool {
	return &pool{
		p: make(map[ksuid.KSUID]poolobj),
	}
}

func (pool *pool) Add(o poolobj) ksuid.KSUID {
	pool.Lock()
	defer pool.Unlock()

	if o == nil {
		return ksuid.Nil
	}

	for i := 0; i < 3; i++ {
		newid := ksuid.New()

		_, ok := pool.p[newid]
		if ok {
			continue
		}

		o.SetId(newid)
		pool.p[newid] = o
		return newid
	}

	return ksuid.Nil
}

func (pool *pool) Get(id ksuid.KSUID) poolobj {
	pool.Lock()
	defer pool.Unlock()

	return pool.p[id]
}

func (pool *pool) Has(id ksuid.KSUID) bool {
	return pool.Get(id) != nil
}

func (pool *pool) Del(id ksuid.KSUID) {
	pool.Lock()
	defer pool.Unlock()

	delete(pool.p, id)
}

func (pool *pool) Range(f func(key ksuid.KSUID, val poolobj) bool) {
	for k, v := range pool.p {
		if !f(k, v) {
			break
		}
	}
}

type cntrPool struct {
	*pool
}

func newCntrPool() *cntrPool {
	return &cntrPool{pool: newPool()}
}

func (p *cntrPool) Add(i *container) ksuid.KSUID {
	return p.pool.Add(i)
}

func (p *cntrPool) Get(id ksuid.KSUID) *container {
	r := p.pool.Get(id)
	if r != nil {
		return r.(*container)
	}
	return nil
}

func (p *cntrPool) Del(id ksuid.KSUID) {
	p.pool.Del(id)
}

func (p *cntrPool) Range(f func(key ksuid.KSUID, val *container) bool) {
	p.pool.Range(func(key ksuid.KSUID, val poolobj) bool {
		return f(key, val.(*container))
	})
}

type taskPool struct {
	*pool
}

func newTaskPool() *taskPool {
	return &taskPool{pool: newPool()}
}

func (p *taskPool) Add(i *task) ksuid.KSUID {
	return p.pool.Add(i)
}

func (p *taskPool) Get(id ksuid.KSUID) *task {
	r := p.pool.Get(id)
	if r != nil {
		return r.(*task)
	}
	return nil
}

func (p *taskPool) Del(id ksuid.KSUID) {
	p.pool.Del(id)
}

func (p *taskPool) Range(f func(key ksuid.KSUID, val *task) bool) {
	p.pool.Range(func(key ksuid.KSUID, val poolobj) bool {
		return f(key, val.(*task))
	})
}
