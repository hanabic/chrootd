package local

import (
	"testing"

	ctest "github.com/xhebox/chrootd/cntr/test"
)

func TestCntrInstanceMeta(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceMeta(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceStart(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStart(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceWait(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceWait(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceStop(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStop(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceStopAll(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStopAll(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceAttach(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceAttach(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceList(t *testing.T) {
	mgr, err := NewTestCntrManager()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceList(mgr.Meta, mgr.Cntr, t)
}
