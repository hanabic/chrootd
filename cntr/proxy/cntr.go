package proxy

import (
	"bytes"
	"io"
	"net"

	"github.com/xhebox/chrootd/client"
	ctyp "github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
)

type cntr struct {
	cid string
	*CntrProxy
}

func (m *cntr) Meta() (*ctyp.Cntrinfo, error) {
	res := &ctyp.Cntrinfo{}
	err := m.Call(m.cid, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "CntrMeta", m.cid, res)
	})
	return res,err
}

func (m *cntr) Start(task *ctyp.Taskinfo) (string, error) {
	res := ""
	return res, m.Call(m.cid, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "CntrStart", &CntrStartReq{
			Id:   m.cid,
			Info: task,
		}, &res)
	})
}

func (m *cntr) Stop(tid string, kill bool) error {
	return m.Call(m.cid, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "CntrStop", &CntrStopReq{
			Id:     m.cid,
			TaskId: tid,
			Kill:   kill,
		}, nil)
	})
}

func (m *cntr) StopAll(kill bool) error {
	return m.Call(m.cid, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "CntrStopAll", &CntrStopAllReq{
			Id:   m.cid,
			Kill: kill,
		}, nil)
	})
}

func (m *cntr) Wait() error {
	return m.Call(m.cid, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "CntrWait", m.cid, nil)
	})
}

func (m *cntr) List(f func(string) error) error {
	return m.Call(m.cid, func(cli client.Client, svc map[string]string) error {
		res := []string{}

		err := cli.Call(m.Context, m.svc, "CntrList", m.cid, &res)
		if err != nil {
			return err
		}

		for _, v := range res {
			if err := f(v); err != nil {
				return err
			}
		}

		return nil

	})
}

func (m *cntr) Attach(tid string) (ctyp.Attacher, error) {
	var attachAddr *utils.Addr
	tok := []byte{}
	err := m.Call(m.cid, func(cli client.Client, svc map[string]string) error {
		attachAddr = utils.NewAddrString(svc["attachNetwork"], svc["attach"])
		return cli.Call(m.Context, m.svc, "CntrAttach", &CntrAttachReq{
			Id:     m.cid,
			TaskId: tid,
		}, &tok)
	})
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr(attachAddr.Network(), attachAddr.String())
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP(attachAddr.Network(), nil, addr)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(conn, bytes.NewReader(tok))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
