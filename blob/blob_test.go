package blob

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhebox/chrootd/store"
)

func setupManager() (*Manager, error) {
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

	return New(dir, s)
}

func write(mgr *Manager, meta, data string) (string, error) {
	writeToken, err := mgr.WriteToken(meta)
	if err != nil {
		return "", err
	}

	fd, err := mgr.Write(writeToken)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	_, err = io.Copy(fd, strings.NewReader(data))
	if err != nil {
		return "", err
	}

	return writeToken, err
}

func read(mgr *Manager, token string) (string, error) {
	readToken, err := mgr.ReadToken(token)
	if err != nil {
		return "", err
	}

	fd, err := mgr.Read(readToken)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, fd)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func TestManagerWrite(t *testing.T) {
	mgr, err := setupManager()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)

	_, err = write(mgr, `{
	"name": "test",
}`, "data")
	if err == nil {
		t.Fatal("invalid json passed test")
	}

	_, err = write(mgr, `{
	"name": "test"
}`, "data")
	if err != nil {
		t.Fatal(err)
	}
}

func TestManagerRead(t *testing.T) {
	mgr, err := setupManager()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)

	writeToken, err := write(mgr, `{
	"name": "test"
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	res, err := read(mgr, writeToken)
	if err != nil {
		t.Fatal(err)
	}

	if res != "data" {
		t.Fatal("readwrite different")
	}
}

func TestManagerDelete(t *testing.T) {
	mgr, err := setupManager()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)

	writeToken, err := write(mgr, `{
	"name": "test"
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.Delete(writeToken)
	if err != nil {
		t.Fatal(err)
	}

	_, err = read(mgr, writeToken)
	if err == nil {
		t.Fatal("deleted, but still readable")
	}
}

func TestManagerUpdate(t *testing.T) {
	mgr, err := setupManager()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)

	writeToken, err := write(mgr, `{
	"name": "test"
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.Update(writeToken, "{}")
	if err != nil {
		t.Fatal(err)
	}

	newm, err := mgr.GetMeta(writeToken)
	if err != nil {
		t.Fatal(err)
	}

	if newm != "{}" {
		t.Fatal("updated, but unexpected content")
	}
}

func TestManagerList(t *testing.T) {
	mgr, err := setupManager()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)

	_, err = write(mgr, `{
	"name": "test",
	"age": 37
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	res, err := mgr.List(`[@this].#(name=="test")`)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Fatal("fail to list")
	}
}
