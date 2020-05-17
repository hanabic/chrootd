package local

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	ctyp "github.com/xhebox/chrootd/cntr"
	mtyp "github.com/xhebox/chrootd/meta"
	mloc "github.com/xhebox/chrootd/meta/local"
	ctest "github.com/xhebox/chrootd/cntr/test"
	"github.com/xhebox/chrootd/store"
)

func init() {
	if len(os.Args) < 2 || os.Args[1] != "___init" {
		return
	}

	InitLibcontainer()
}

type TestCntrManager struct {
	dir  string
	Meta mtyp.Manager
	Cntr ctyp.Manager
}

func NewTestCntrManager() (*TestCntrManager, error) {
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

	mgr2, err := NewCntrManager(dir, image, s)
	if err != nil {
		return nil, err
	}

	return &TestCntrManager{dir: dir, Meta: mgr1, Cntr: mgr2}, nil
}

func (t *TestCntrManager) Close() error {
	t.Meta.Close()
	t.Cntr.Close()
	return os.RemoveAll(t.dir)
}

func TestCntrManagerID(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerID(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerCreate(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerCreate(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerGet(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerGet(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerDelete(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerDelete(mgr.Meta, mgr.Cntr, t)
}

func TestCntrManagerList(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrManagerList(mgr.Meta, mgr.Cntr, t)
}
