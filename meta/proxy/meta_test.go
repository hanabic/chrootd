package proxy

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/smallnest/rpcx/server"
	"github.com/xhebox/chrootd/client"
	mtyp "github.com/xhebox/chrootd/meta"
	mloc "github.com/xhebox/chrootd/meta/local"
	mtest "github.com/xhebox/chrootd/meta/test"
	"github.com/xhebox/chrootd/store"
	"github.com/xhebox/chrootd/utils"
)

func newTestMetaProxy(consul bool, t *testing.T) (*TestMetaProxy, error) {
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

	loc1, err := mloc.NewMetaManager(dir, image, s1)
	if err != nil {
		return nil, err
	}

	loc2, err := mloc.NewMetaManager(filepath.Join(dir, "l1"), image, s2)
	if err != nil {
		return nil, err
	}

	addr := utils.NewAddrFree()

	lnRPC, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	addr2 := utils.NewAddrFree()

	lnRPC2, err := net.Listen(addr2.Network(), addr2.String())
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
			Address: tsrv.HTTPAddr,
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

	svc1, err := NewMetaService(loc1, con, "s", addr)
	if err != nil {
		return nil, err
	}

	err = srv1.RegisterName("s", svc1, "")
	if err != nil {
		return nil, err
	}

	go srv1.ServeListener(addr.Network(), lnRPC)

	srv2 := server.NewServer(
		func(srv *server.Server) {
			srv.DisableHTTPGateway = true
			srv.DisableJSONRPC = true
		},
	)

	svc2, err := NewMetaService(loc2, con, "s", addr2)
	if err != nil {
		return nil, err
	}

	err = srv2.RegisterName("s", svc2, "")
	if err != nil {
		return nil, err
	}

	go srv2.ServeListener(addr2.Network(), lnRPC2)

	var mgr mtyp.Manager
	if consul {
		mgr, err = NewMetaProxy("s", con)
		if err != nil {
			return nil, err
		}
	} else {
		cli, err := client.NewClient(addr.Network(), addr.String())
		if err != nil {
			return nil, err
		}

		mgr, err = NewMetaProxy("s", cli)
		if err != nil {
			return nil, err
		}
	}

	return &TestMetaProxy{dir: dir,
		tsrv:    tsrv,
		srv1:    srv1,
		srv2:    srv2,
		svc1:    svc1,
		svc2:    svc2,
		loc1:    loc1,
		loc2:    loc2,
		Manager: mgr}, nil
}

func NewTestMetaProxy(t *testing.T) (*TestMetaProxy, error) {
	return newTestMetaProxy(false, t)
}

func NewTestMetaProxyConsul(t *testing.T) (*TestMetaProxy, error) {
	return newTestMetaProxy(true, t)
}

type TestMetaProxy struct {
	dir  string
	tsrv *testutil.TestServer
	loc1 mtyp.Manager
	svc1 *MetaService
	srv1 *server.Server
	loc2 mtyp.Manager
	svc2 *MetaService
	srv2 *server.Server
	mtyp.Manager
}

func (t *TestMetaProxy) Close() error {
	t.Manager.Close()
	t.srv2.Shutdown(context.Background())
	t.loc2.Close()
	t.srv1.Shutdown(context.Background())
	t.loc1.Close()
	if t.tsrv != nil {
		t.tsrv.Stop()
	}
	return os.RemoveAll(t.dir)
}

func TestMetaManagerID(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerID(mgr, t)
}

func TestMetaManagerCreate(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerCreate(mgr, t)
}

func TestMetaManagerGet(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerGet(mgr, t)
}

func TestMetaManagerUpdate(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerUpdate(mgr, t)
}

func TestMetaManagerDelete(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerDelete(mgr, t)
}

func TestMetaManagerQuery(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerQuery(mgr, t)
}

func TestMetaManagerImageUnpack(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageUnpack(mgr, t)
}

func TestMetaManagerImageDelete(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageDelete(mgr, t)
}

func TestMetaManagerImageList(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageList(mgr, t)
}

func TestMetaManagerImageAvailable(t *testing.T) {
	mgr, err := NewTestMetaProxy(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageAvailable(mgr, t)
}

func TestMetaManagerConsulID(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	time.Sleep(100 * time.Millisecond)

	mtest.TestMetaManagerID(mgr, t)
}

func TestMetaManagerConsulCreate(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerCreate(mgr, t)
}

func TestMetaManagerConsulGet(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerGet(mgr, t)
}

func TestMetaManagerConsulUpdate(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerUpdate(mgr, t)
}

func TestMetaManagerConsulDelete(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerDelete(mgr, t)
}

func TestMetaManagerConsulQuery(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerQuery(mgr, t)
}

func TestMetaManagerConsulImageUnpack(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageUnpack(mgr, t)
}

func TestMetaManagerConsulImageDelete(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageDelete(mgr, t)
}

func TestMetaManagerConsulImageList(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageList(mgr, t)
}

func TestMetaManagerConsulImageAvailable(t *testing.T) {
	mgr, err := NewTestMetaProxyConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageAvailable(mgr, t)
}
