package test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	. "github.com/xhebox/chrootd/meta"
)

func TestMetaManagerID(mgr Manager, t *testing.T) {
	_, err := mgr.ID()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMetaManagerCreate(mgr Manager, t *testing.T) {
	id1, err := mgr.Create(&Metainfo{})
	if err != nil {
		t.Fatal(err)
	}

	id2, err := mgr.Create(&Metainfo{})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(id1, id2)
}

func TestMetaManagerGet(mgr Manager, t *testing.T) {
	id, err := mgr.Create(&Metainfo{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}

	meta, err := mgr.Get(id)
	if err != nil {
		t.Fatal(err)
	}

	if meta.Name != "test" {
		t.Fatal("name different")
	}
}

func TestMetaManagerUpdate(mgr Manager, t *testing.T) {
	id, err := mgr.Create(&Metainfo{})
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.Update(&Metainfo{Id: id})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMetaManagerDelete(mgr Manager, t *testing.T) {
	id, err := mgr.Create(&Metainfo{})
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.Delete(id)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgr.Get(id)
	if err == nil {
		t.Fatal("deleted, but still found")
	}
}

func TestMetaManagerQuery(mgr Manager, t *testing.T) {
	_, err := mgr.Create(&Metainfo{Name: "test1"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgr.Create(&Metainfo{Name: "test2"})
	if err != nil {
		t.Fatal(err)
	}

	var retMeta *Metainfo
	err = mgr.Query(`[@this].#(name=="test1")`, func(v *Metainfo) error {
		retMeta = v
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if retMeta == nil || retMeta.Name != "test1" {
		t.Fatal("fail to query the meta just created")
	}
}

func TestMetaManagerImageUnpack(mgr Manager, t *testing.T) {
	id, err := mgr.Create(&Metainfo{
		Name:           "test",
		Image:          "alpine",
		ImageReference: "latest",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgr.ImageUnpack(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMetaManagerImageDelete(mgr Manager, t *testing.T) {
	id, err := mgr.Create(&Metainfo{
		Name:           "test",
		Image:          "alpine",
		ImageReference: "latest",
	})
	if err != nil {
		t.Fatal(err)
	}

	rid, err := mgr.ImageUnpack(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.ImageDelete(id, rid)
	if err != nil {
		t.Fatal(err)
	}

	meta, err := mgr.Get(id)
	if err != nil {
		t.Fatal(err)
	}

	if len(meta.RootfsIds) != 0 {
		t.Fatal("fail to delete unpacked rootfs")
	}
}

func TestMetaManagerImageList(mgr Manager, t *testing.T) {
	id, err := mgr.Create(&Metainfo{
		Name:           "test",
		Image:          "alpine",
		ImageReference: "latest",
	})
	if err != nil {
		t.Fatal(err)
	}

	rid, err := mgr.ImageUnpack(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.ImageList(id, func(k string) error {
		if k != rid {
			return errors.New("unexpected rootfs id")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMetaManagerImageAvailable(mgr Manager, t *testing.T) {
	err := mgr.ImageAvailable(context.Background(), func(id string, name string, refs []string) error {
		t.Logf("id: %s, name: %s, refs: %v\n", id, name, refs)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
