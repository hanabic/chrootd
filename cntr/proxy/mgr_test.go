package proxy

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
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

	image, err := filepath.Abs("../../images")
	if err != nil {
		return nil, err
	}

	mgr1, err := mloc.NewMetaManager(dir, image, s)
	if err != nil {
		return nil, err
	}

	mgr2, err := cloc.NewCntrManager(dir, image, s)
	if err != nil {
		return nil, err
	}

	addr := utils.NewAddrFree()
	attachAddr := utils.NewAddrFree()

	lnRPC, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	lnAttach, err := net.Listen(attachAddr.Network(), attachAddr.String())
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
			HttpClient: tsrv.HTTPClient,
		})
		if err != nil {
			return nil, err
		}
	}

	srv := server.NewServer(
		func(srv *server.Server) {
			srv.DisableHTTPGateway = true
			srv.DisableJSONRPC = true
		},
	)

	msvc, err := mpro.NewMetaService(mgr1, con, "meta", addr)
	if err != nil {
		return nil, err
	}

	err = srv.RegisterName("meta", msvc, "")
	if err != nil {
		return nil, err
	}

	csvc, err := NewCntrService(mgr2, con, "cntr", addr, attachAddr)
	if err != nil {
		return nil, err
	}

	err = srv.RegisterName("cntr", csvc, "")
	if err != nil {
		return nil, err
	}

	go srv.ServeListener(addr.Network(), lnRPC)
	go csvc.ServeListener(lnAttach)

	cli, err := client.NewClient(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	mpr1, err := mpro.NewMetaProxy("meta", cli)
	if err != nil {
		return nil, err
	}

	mpr2, err := NewCntrProxy("cntr", cli, attachAddr)
	if err != nil {
		return nil, err
	}

	return &TestCntrManager{dir: dir, tsrv: tsrv, csvc: csvc, srv: srv, meta: mgr1, cntr: mgr2, Meta: mpr1, Cntr: mpr2}, nil
}

func NewTestCntrManager(t *testing.T) (*TestCntrManager, error) {
	return newTestCntrManager(false, t)
}

func NewTestCntrManagerConsul(t *testing.T) (*TestCntrManager, error) {
	return newTestCntrManager(true, t)
}

type TestCntrManager struct {
	dir  string
	csvc *CntrService
	tsrv *testutil.TestServer
	srv  *server.Server
	meta mtyp.Manager
	cntr ctyp.Manager
	Meta mtyp.Manager
	Cntr ctyp.Manager
}

func (t *TestCntrManager) Close() error {
	t.Meta.Close()
	t.Cntr.Close()
	t.srv.Shutdown(context.Background())
	t.csvc.Shutdown()
	t.meta.Close()
	t.cntr.Close()
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

	runtime.Gosched()

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
