package proxy

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	ctyp "github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
)

type CntrService struct {
	addr        *utils.Addr
	reg         *api.AgentServiceRegistration
	cli         *api.Client
	mgr         ctyp.Manager
	tok         *cache.Cache
	activeConn  map[net.Conn]struct{}
	mu          sync.Mutex
	QueryLimits int
}

func NewCntrService(mgr ctyp.Manager, cli *api.Client, svcname string, rpcAddr, attachAddr *utils.Addr) (*CntrService, error) {
	svc := &CntrService{
		cli:         cli,
		mgr:         mgr,
		addr:        rpcAddr,
		tok:         cache.New(time.Minute, 10*time.Minute),
		activeConn:  make(map[net.Conn]struct{}),
		QueryLimits: 64,
	}

	if cli != nil {
		id, err := mgr.ID()
		if err != nil {
			return nil, err
		}

		svc.reg = &api.AgentServiceRegistration{
			ID:      ksuid.New().String(),
			Name:    svcname,
			Address: svc.addr.Addr(),
			Port:    svc.addr.Port(),
			Meta: map[string]string{
				"attachNetwork": attachAddr.Network(),
				"attach":        attachAddr.String(),
			},
			Tags: []string{id},
			Checks: api.AgentServiceChecks{
				&api.AgentServiceCheck{
					DeregisterCriticalServiceAfter: "1h",
					Interval:                       "1m",
					Timeout:                        "30s",
					TCP:                            svc.addr.String(),
				},
			},
		}

		err = cli.Agent().ServiceRegister(svc.reg)
		if err != nil {
			return nil, err
		}
	}

	return svc, nil
}

func (s *CntrService) ID(ctx context.Context, req struct{}, res *string) error {
	var err error
	*res, err = s.mgr.ID()
	return err
}

func (s *CntrService) Create(ctx context.Context, req *ctyp.Cntrinfo, res *string) error {
	id, err := s.mgr.Create(req)
	if err == nil {
		*res = id
	}
	return err
}

func (s *CntrService) Delete(ctx context.Context, cid string, res *struct{}) error {
	return s.mgr.Delete(cid)
}

func (s *CntrService) List(ctx context.Context, req string, res *[]ctyp.Cntrinfo) error {
	cnt := 0
	return s.mgr.List(req, func(cmeta *ctyp.Cntrinfo) error {
		*res = append(*res, *cmeta)
		cnt++
		if cnt > s.QueryLimits {
			return errors.New("exceed the query limits")
		}
		return nil
	})
}

func (s *CntrService) CntrMeta(ctx context.Context, req string, res *ctyp.Cntrinfo) error {
	cntr, err := s.mgr.Get(req)
	if err != nil {
		return err
	}

	meta, err := cntr.Meta()
	if err != nil {
		return err
	}

	*res = *meta
	return nil
}

type CntrStartReq struct {
	Id   string
	Info *ctyp.Taskinfo
}

func (s *CntrService) CntrStart(ctx context.Context, req *CntrStartReq, res *string) error {
	cntr, err := s.mgr.Get(req.Id)
	if err != nil {
		return err
	}

	*res, err = cntr.Start(req.Info)
	return err
}

type CntrStopReq struct {
	Id     string
	TaskId string
	Kill   bool
}

func (s *CntrService) CntrStop(ctx context.Context, req *CntrStopReq, res *struct{}) error {
	cntr, err := s.mgr.Get(req.Id)
	if err != nil {
		return err
	}

	return cntr.Stop(req.TaskId, req.Kill)
}

type CntrStopAllReq struct {
	Id   string
	Kill bool
}

func (s *CntrService) CntrStopAll(ctx context.Context, req *CntrStopAllReq, res *struct{}) error {
	cntr, err := s.mgr.Get(req.Id)
	if err != nil {
		return err
	}

	return cntr.StopAll(req.Kill)
}

func (s *CntrService) CntrWait(ctx context.Context, req string, res *struct{}) error {
	cntr, err := s.mgr.Get(req)
	if err != nil {
		return err
	}

	cntr.Wait()
	return nil
}

func (s *CntrService) CntrList(ctx context.Context, req string, res *[]string) error {
	cntr, err := s.mgr.Get(req)
	if err != nil {
		return err
	}

	cnt := 0
	return cntr.List(func(tid string) error {
		*res = append(*res, tid)
		cnt++
		if cnt > s.QueryLimits {
			return errors.New("exceed the query limits")
		}
		return nil
	})
}

type CntrAttachReq struct {
	Id     string
	TaskId string
}

func (s *CntrService) CntrAttach(ctx context.Context, req *CntrAttachReq, res *[]byte) error {
	_, err := s.mgr.Get(req.Id)
	if err != nil {
		return err
	}

	*res = ksuid.New().Bytes()
	return s.tok.Add(string(*res), req, cache.DefaultExpiration)
}

func (s *CntrService) ServeListener(ln net.Listener) error {
	defer ln.Close()

	for {
		c, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(100 * time.Microsecond)
				continue
			}
			return err
		}

		s.mu.Lock()
		s.activeConn[c] = struct{}{}
		s.mu.Unlock()

		go func(conn net.Conn) (err error) {
			defer func() {
				if err != nil {
					conn.Write([]byte("internal error\n"))
				}
				conn.Close()
			}()

			var tok ksuid.KSUID

			_, err = io.ReadFull(conn, tok[:])
			if err != nil {
				return err
			}

			tmp, _ := s.tok.Get(string(tok.Bytes()))
			reqt, ok := tmp.(*CntrAttachReq)
			if !ok {
				return errors.New("invalid token")
			}

			var cntr ctyp.Cntr
			cntr, err = s.mgr.Get(reqt.Id)
			if err != nil {
				return err
			}

			var rw ctyp.Attacher
			rw, err = cntr.Attach(reqt.TaskId)
			if err != nil {
				return err
			}

			ch := make(chan bool, 1)

			go func() {
				io.Copy(rw, conn)
				rw.CloseWrite()
				ch <- true
			}()

			buf := make([]byte, 128)
			lcond := true
			for lcond {
				select {
				case <-ch:
					lcond = false
				default:
				}

				n, err := rw.Read(buf)
				if err != nil {
					if err != io.EOF {
						conn.Close()
					}
					lcond = false
				}

				_, err = io.Copy(conn, bytes.NewReader(buf[:n]))
				if err != nil {
					conn.Close()
				}
			}

			s.mu.Lock()
			delete(s.activeConn, c)
			s.mu.Unlock()

			return nil
		}(c)
	}
}

func (s *CntrService) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	for c := range s.activeConn {
		err = c.Close()
	}
	return err
}
