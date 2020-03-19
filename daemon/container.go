package main

import (
	"sync"

	"github.com/segmentio/ksuid"
	"github.com/xhebox/chrootd/api"
)

type container struct {
	sync.Mutex
	*api.Container

	busy bool
	id   ksuid.KSUID
}

func newCntr(meta *api.Container) *container {
	return &container{
		id: ksuid.Nil,
		Container: meta,
	}
}

func (s *container) SetId(id ksuid.KSUID) {
	s.Lock()
	defer s.Unlock()

	if s.id != ksuid.Nil {
		return
	}

	s.id = id
	s.Id = id.Bytes()
}

func (s *container) UpdateMeta(meta *api.Container) {
	s.Lock()
	defer s.Unlock()

	// TODO: Copy & overwrite as needed
	// TODO: add more until metadata struct definition is freezed
	s.Name = meta.Name
}

func (s *container) IsBusy() bool {
	s.Lock()
	defer s.Unlock()

	return s.busy
}

func (s *container) Destroy() {
	// TODO: it's a stub
}
