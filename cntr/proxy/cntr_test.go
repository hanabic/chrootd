package proxy

import (
	"testing"

	ctest "github.com/xhebox/chrootd/cntr/test"
)

func TestCntrInstanceMeta(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceMeta(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceStart(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStart(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceWait(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceWait(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceStop(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStop(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceStopAll(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStopAll(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceAttach(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceAttach(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceList(t *testing.T) {
	mgr, err := NewTestCntrManager(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceList(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceConsulMeta(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceMeta(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceConsulStart(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStart(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceConsulWait(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceWait(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceConsulStop(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStop(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceConsulStopAll(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceStopAll(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceConsulAttach(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceAttach(mgr.Meta, mgr.Cntr, t)
}

func TestCntrInstanceConsulList(t *testing.T) {
	mgr, err := NewTestCntrManagerConsul(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	ctest.TestCntrInstanceList(mgr.Meta, mgr.Cntr, t)
}
