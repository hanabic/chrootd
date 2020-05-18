package proxy

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/smallnest/rpcx/server"
	"github.com/xhebox/chrootd/client"
	ctyp "github.com/xhebox/chrootd/cntr"
	cloc "github.com/xhebox/chrootd/cntr/local"
	ctest "github.com/xhebox/chrootd/cntr/test"
	mtyp "github.com/xhebox/chrootd/meta"
	mloc "github.com/xhebox/chrootd/meta/local"
	mpro "github.com/xhebox/chrootd/meta/proxy"
	"github.com/xhebox/chrootd/store"
	"github.com/xhebox/chrootd/utils"
)

func init() {
	if len(os.Args) < 2 || os.Args[1] != "___init" {
		return
	}

	cloc.InitLibcontainer()
}

func newTestCntrManager(consul bool, t *testing.T) (*TestCntrManager, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "temp")
	if err != nil {
		return nil, err
	}

	os.MkdirAll(dir, 0755)

	s, err := store.NewBolt(
		filepath.Join(dir, "s"),
		"test",
	)
	if err != nil {
		return nil, err
	}

	s1, err := store.NewWrapStore("l1", s)
	if err != nil {
		return nil, err
	}

	s2, err := store.NewWrapStore("l2", s)
	if err != nil {
		return nil, err
	}

	image, err := filepath.Abs("../../images")
	if err != nil {
		return nil, err
	}

	mmgr1, err := mloc.NewMetaManager(filepath.Join(dir, "l1"), image, s1)
	if err != nil {
		return nil, err
	}

	cmgr1, err := cloc.NewCntrManager(filepath.Join(dir, "l1"), image, s1)
	if err != nil {
		return nil, err
	}

	mmgr2, err := mloc.NewMetaManager(filepath.Join(dir, "l2"), image, s2)
	if err != nil {
		return nil, err
	}

	cmgr2, err := cloc.NewCntrManager(filepath.Join(dir, "l2"), image, s2)
	if err != nil {
		return nil, err
	}

	addr1 := utils.NewAddrFree()
	attachAddr1 := utils.NewAddrFree()

	lnRPC1, err := net.Listen(addr1.Network(), addr1.String())
	if err != nil {
		return nil, err
	}

	lnAttach1, err := net.Listen(attachAddr1.Network(), attachAddr1.String())
	if err != nil {
		return nil, err
	}

	addr2 := utils.NewAddrFree()
	attachAddr2 := utils.NewAddrFree()

	lnRPC2, err := net.Listen(addr2.Network(), addr2.String())
	if err != nil {
		return nil, err
	}

	lnAttach2, err := net.Listen(attachAddr2.Network(), attachAddr2.String())
	if err != nil {
		return nil, err
	}

	var con *api.Client
	var tsrv *testutil.TestServer
	if consul {
		tsrv, err = testutil.NewTestServerConfigT(t, func(cfg *testutil.TestServerConfig) {
			cfg.LogLevel = "err"
		})
		if err != nil {
			return nil, err
		}

		con, err = api.NewClient(&api.Config{
			Address:    tsrv.HTTPAddr,
		})
		if err != nil {
			return nil, err
		}
	}

	srv1 := server.NewServer(
		func(srv *server.Server) {
			srv.DisableHTTPGateway = true
			srv.DisableJSONRPC = true
		},
	)

	msvc1, err := mpro.NewMetaService(mmgr1, con, "meta", addr1)
	if err != nil {
		return nil, err
	}

	err = srv1.RegisterName("meta", msvc1, "")
	if err != nil {
		return nil, err
	}

	csvc1, err := NewCntrService(cmgr1, con, "cntr", addr1, attachAddr1)
	if err != nil {
		return nil, err
	}

	err = srv1.RegisterName("cntr", csvc1, "")
	if err != nil {
		return nil, err
	}

	go srv1.ServeListener(addr1.Network(), lnRPC1)
	go csvc1.ServeListener(lnAttach1)

	srv2 := server.NewServer(
		func(srv *server.Server) {
			srv.DisableHTTPGateway = true
			srv.DisableJSONRPC = true
		},
	)

	msvc2, err := mpro.NewMetaService(mmgr2, con, "meta", addr2)
	if err != nil {
		return nil, err
	}

	err = srv2.RegisterName("meta", msvc2, "")
	if err != nil {
		return nil, err
	}

	csvc2, err := NewCntrService(cmgr2, con, "cntr", addr2, attachAddr2)
	if err != nil {
		return nil, err
	}

	err = srv2.RegisterName("cntr", csvc2, "")
	if err != nil {
		return nil, err
	}

	go srv2.ServeListener(addr2.Network(), lnRPC2)
	go csvc2.ServeListener(lnAttach2)

	var mpr1 mtyp.Manager
	var mpr2 ctyp.Manager
	if consul {
		mpr1, err = mpro.NewMetaProxy("meta", con)
		if err != nil {
			return nil, err
		}

		mpr2, err = NewCntrProxy("cntr", con, nil)
		if err != nil {
			return nil, err
		}
	} else {
		cli, err := client.NewClient(addr1.Network(), addr1.String())
		if err != nil {
			return nil, err
		}

		mpr1, err = mpro.NewMetaProxy("meta", cli)
		if err != nil {
			return nil, err
		}

		mpr2, err = NewCntrProxy("cntr", cli, attachAddr1)
		if err != nil {
			return nil, err
		}
	}

	return &TestCntrManager{
		dir:   dir,
		tsrv:  tsrv,
		csvc1: csvc1,
		srv1:  srv1,
		meta1: mmgr1,
		cntr1: cmgr1,
		csvc2: csvc2,
		srv2:  srv2,
		meta2: mmgr2,
		cntr2: cmgr2,
		Meta:  mpr1,
		Cntr:  mpr2,
	}, nil
}

func NewTestCntrManager(t *testing.T) (*TestCntrManager, error) {
	return newTestCntrManager(false, t)
}

func NewTestCntrManagerConsul(t *testing.T) (*TestCntrManager, error) {
	return newTestCntrManager(true, t)
}

type TestCntrManager struct {
	dir  string
	tsrv *testutil.TestServer

	csvc1 *CntrService
	meta1 mtyp.Manager
	cntr1 ctyp.Manager
	srv1  *server.Server

	csvc2 *CntrService
	meta2 mtyp.Manager
	cntr2 ctyp.Manager
	srv2  *server.Server

	Meta mtyp.Manager
	Cntr ctyp.Manager
}

func (t *TestCntrManager) Close() error {
	t.Meta.Close()
	t.Cntr.Close()

	t.srv1.Shutdown(context.Background())
	t.csvc1.Shutdown()
	t.meta1.Close()
	t.cntr1.Close()

	t.srv2.Shutdown(context.Background())
	t.csvc2.Shutdown()
	t.meta2.Close()
	t.cntr2.Close()
	if t.tsrv != nil {
		t.tsrv.Stop()
	}
	return os.RemoveAll(t.dir)
}

func TestCntrManagerID(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerID(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerCreate(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerCreate(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerGet(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerGet(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerDelete(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerDelete(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerList(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerList(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerConsulID(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	time.Sleep(100 * time.Millisecond)

	ctest.TestCntrManagerID(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerConsulCreate(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerCreate(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerConsulGet(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerGet(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerConsulDelete(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerDelete(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerConsulList(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerList(mgr.Meta, mgr.Cntr, t)
}
