package test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	ctyp "github.com/xhebox/chrootd/cntr"
	mtyp "github.com/xhebox/chrootd/meta"
)

func TestCntrManagerID(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
	mid, err := mmgr.ID()
	if err != nil {
		t.Fatal(err)
	}

	cid, err := cmgr.ID()
	if err != nil {
		t.Fatal(err)
	}

	if mid != cid {
		t.Fatal("id different")
	}
}

func TestCntrManagerCreate(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
	mid, err := mmgr.Create(&mtyp.Metainfo{
		Name:           "test",
		Image:          "busybox",
		ImageReference: "latest",
	})
	if err != nil {
		t.Fatal(err)
	}

	meta, err := mmgr.Get(mid)
	if err != nil {
		t.Fatal(err)
	}

	rid, err := mmgr.ImageUnpack(context.Background(), mid)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cmgr.Create(&ctyp.Cntrinfo{
		Rootfs: rid,
		Meta: meta,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCntrManagerGet(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
	mid, err := mmgr.Create(&mtyp.Metainfo{
		Name:           "test",
		Image:          "busybox",
		ImageReference: "latest",
	})
	if err != nil {
		t.Fatal(err)
	}

	meta, err := mmgr.Get(mid)
	if err != nil {
		t.Fatal(err)
	}

	rid, err := mmgr.ImageUnpack(context.Background(), mid)
	if err != nil {
		t.Fatal(err)
	}

	cid, err := cmgr.Create(&ctyp.Cntrinfo{
		Rootfs: rid,
		Meta: meta,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = cmgr.Get(cid)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCntrManagerDelete(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
	mid, err := mmgr.Create(&mtyp.Metainfo{
		Name:           "test",
		Image:          "busybox",
		ImageReference: "latest",
	})
	if err != nil {
		t.Fatal(err)
	}

	meta, err := mmgr.Get(mid)
	if err != nil {
		t.Fatal(err)
	}

	rid, err := mmgr.ImageUnpack(context.Background(), mid)
	if err != nil {
		t.Fatal(err)
	}

	cid, err := cmgr.Create(&ctyp.Cntrinfo{
		Rootfs: rid,
		Meta: meta,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cmgr.Delete(cid)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCntrManagerList(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
	mid, err := mmgr.Create(&mtyp.Metainfo{
		Name:           "test",
		Image:          "busybox",
		ImageReference: "latest",
	})
	if err != nil {
		t.Fatal(err)
	}

	meta, err := mmgr.Get(mid)
	if err != nil {
		t.Fatal(err)
	}

	rid, err := mmgr.ImageUnpack(context.Background(), mid)
	if err != nil {
		t.Fatal(err)
	}

	cid, err := cmgr.Create(&ctyp.Cntrinfo{
		Rootfs: rid,
		Meta: meta,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cmgr.List("", func(m *ctyp.Cntrinfo) error {
		if m.Id != cid {
			return errors.New("unexpected container id")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
