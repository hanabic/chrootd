package test

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/pkg/errors"
	ctyp "github.com/xhebox/chrootd/cntr"
	mtyp "github.com/xhebox/chrootd/meta"
)

func TestCntrInstanceMeta(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
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

	cntr, err := cmgr.Get(cid)
	if err != nil {
		t.Fatal(err)
	}

	mmeta, err := cntr.Meta()
	if err != nil {
		t.Fatal(err)
	}

	if meta.Name != mmeta.Meta.Name {
		t.Fatal("except same name")
	}
}

func TestCntrInstanceStart(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
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

	cntr, err := cmgr.Get(cid)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cntr.Start(&ctyp.Taskinfo{
		Args: []string{"/bin/ls"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCntrInstanceWait(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
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

	cntr, err := cmgr.Get(cid)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cntr.Start(&ctyp.Taskinfo{
		Args: []string{"/bin/ls"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cntr.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCntrInstanceStop(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
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

	cntr, err := cmgr.Get(cid)
	if err != nil {
		t.Fatal(err)
	}

	tid, err := cntr.Start(&ctyp.Taskinfo{
		Args: []string{"/bin/sh"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cntr.Stop(tid, false)
	if err != nil {
		t.Fatal(err)
	}

	err = cntr.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCntrInstanceStopAll(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
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

	cntr, err := cmgr.Get(cid)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cntr.Start(&ctyp.Taskinfo{
		Args: []string{"/bin/sh"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cntr.StopAll(false)
	if err != nil {
		t.Fatal(err)
	}

	err = cntr.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCntrInstanceAttach(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
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

	cntr, err := cmgr.Get(cid)
	if err != nil {
		t.Fatal(err)
	}

	tid, err := cntr.Start(&ctyp.Taskinfo{
		Args: []string{"/bin/ls"},
	})
	if err != nil {
		t.Fatal(err)
	}

	rw, err := cntr.Attach(tid)
	if err != nil {
		t.Fatal(err)
	}
	defer rw.Close()

	b, err := ioutil.ReadAll(rw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(b), "bin") {
		t.Fatalf("too short content, failed: %s", b)
	}

	t.Logf("content %s\n", b)

	// case 2: interactive example
	tid, err = cntr.Start(&ctyp.Taskinfo{
		Args: []string{"/bin/sh"},
	})
	if err != nil {
		t.Fatal(err)
	}

	rw, err = cntr.Attach(tid)
	if err != nil {
		t.Fatal(err)
	}
	defer rw.Close()

	_, err = io.Copy(rw, strings.NewReader("ls;exit;"))
	if err != nil {
		t.Fatal(err)
	}

	rw.CloseWrite()

	b, err = ioutil.ReadAll(rw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(b), "bin") {
		t.Fatalf("too short content, failed: %s", b)
	}

	t.Logf("content %s\n", b)

	// case3: attach wrong task id
	rw, err = cntr.Attach("34")
	if err == nil {
		// service case
		defer rw.Close()

		b, err = ioutil.ReadAll(rw)
		if !strings.HasPrefix(string(b), "internal error") {
			t.Fatal("except wrong res")
		}
	}
}

func TestCntrInstanceList(mmgr mtyp.Manager, cmgr ctyp.Manager, t *testing.T) {
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

	cntr, err := cmgr.Get(cid)
	if err != nil {
		t.Fatal(err)
	}

	tid, err := cntr.Start(&ctyp.Taskinfo{
		Args: []string{"/bin/sh"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cntr.List(func(k string) error {
		if k != tid {
			return errors.New("unexpected task id")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cntr.Stop(tid, false)
	if err != nil {
		t.Fatal(err)
	}

	err = cntr.Wait()
	if err != nil {
		t.Fatal(err)
	}
}
