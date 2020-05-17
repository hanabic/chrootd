package local

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/xhebox/chrootd/meta"
	mtest "github.com/xhebox/chrootd/meta/test"
	"github.com/xhebox/chrootd/store"
)

type TestMetaManager struct {
	dir string
	Manager
}

func NewTestMetaManager() (*TestMetaManager, error) {
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

	mgr, err := NewMetaManager(dir, image, s)
	if err != nil {
		return nil, err
	}

	return &TestMetaManager{dir: dir, Manager: mgr}, nil
}

func (t *TestMetaManager) Close() error {
	t.Manager.Close()
	return os.RemoveAll(t.dir)
}

func TestMetaManagerID(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerID(mgr, t)
}

func TestMetaManagerCreate(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerCreate(mgr, t)
}

func TestMetaManagerGet(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerGet(mgr, t)
}

func TestMetaManagerUpdate(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerUpdate(mgr, t)
}

func TestMetaManagerDelete(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerDelete(mgr, t)
}

func TestMetaManagerQuery(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerQuery(mgr, t)
}

func TestMetaManagerImageUnpack(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageUnpack(mgr, t)
}

func TestMetaManagerImageDelete(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageDelete(mgr, t)
}

func TestMetaManagerImageList(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageList(mgr, t)
}

func TestMetaManagerImageAvailable(t *testing.T) {
	mgr, err := NewTestMetaManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mtest.TestMetaManagerImageAvailable(mgr, t)
}
