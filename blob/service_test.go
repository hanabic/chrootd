package blob

import (
	"bytes"
	"context"
	"strings"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/smallnest/rpcx/server"
	"github.com/xhebox/chrootd/client"
	"github.com/xhebox/chrootd/utils"
)

func setupServer(ctx context.Context) (*Manager, client.Client, error) {
	mgr, err := setupManager()
	if err != nil {
		return nil, nil, err
	}

	addr := utils.NewAddrFree()

	lnc := net.ListenConfig{}

	ln, err := lnc.Listen(ctx, addr.Network(), addr.Addr())
	if err != nil {
		return nil, nil, err
	}

	srv := server.NewServer()

	svc := NewService(mgr)

	rpcx := client.NewHTTPListener()

	mux := http.NewServeMux()

	mux.HandleFunc("/", rpcx.ServeHTTP)
	mux.HandleFunc("/blob", svc.Serve)

	err = srv.RegisterName("s", svc, "")
	if err != nil {
		return nil, nil, err
	}

	go http.Serve(ln, mux)
	go srv.ServeListener("tcp", rpcx)

	cli, err := client.NewClient("http", fmt.Sprintf("http://%s/", addr.Addr()))
	if err != nil {
		return nil, nil, err
	}

	return mgr, cli, nil
}

func TestServiceList(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr, cli, err := setupServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)
	defer cli.Close()

	_, err = write(mgr, `{
	"name": "test",
	"age": 37
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	res := &BlobListRes{}
	err = cli.Call(ctx, "s", "List", BlobListReq{
		Query: `[@this].#(name=="test")`,
	}, res)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Blobs) != 1 {
		t.Fatal("fail to list")
	}
}

func TestServiceGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr, cli, err := setupServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)
	defer cli.Close()

	writeToken, err := write(mgr, `{
	"name": "test",
	"age": 37
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	res := &BlobGetRes{}
	err = cli.Call(ctx, "s", "Get", BlobGetReq{
		Id: writeToken,
	}, res)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(res)
}

func TestServiceDelete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr, cli, err := setupServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)
	defer cli.Close()

	writeToken, err := write(mgr, `{
	"name": "test",
	"age": 37
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	res1 := &BlobDeleteRes{}
	err = cli.Call(ctx, "s", "Delete", BlobDeleteReq{
		Id: writeToken,
	}, res1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgr.GetMeta(writeToken)
	if err == nil {
		t.Fatal("deleted, still exists")
	}
}

func TestServiceUpdate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr, cli, err := setupServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)
	defer cli.Close()

	writeToken, err := write(mgr, `{
	"name": "test",
	"age": 37
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	res1 := &BlobUpdateRes{}
	err = cli.Call(ctx, "s", "Update", BlobUpdateReq{
		Id:   writeToken,
		Meta: "{}",
	}, res1)
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

func TestServiceRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr, cli, err := setupServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)
	defer cli.Close()

	writeToken, err := write(mgr, `{
	"name": "test",
	"age": 37
}`, "data")
	if err != nil {
		t.Fatal(err)
	}

	res1 := &BlobReadRes{}
	err = cli.Call(ctx, "s", "Read", BlobReadReq{
		Id: writeToken,
	}, res1)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/blob", cli.RemoteAddr().Addr()), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("token", res1.Token)

	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()

	var buf bytes.Buffer
	io.Copy(&buf, res2.Body)

	if buf.String() != "data" {
		t.Fatalf("wrong data [%s]", buf.String())
	}
}

func TestServiceWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr, cli, err := setupServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mgr.path)
	defer cli.Close()

	res1 := &BlobWriteRes{}
	err = cli.Call(ctx, "s", "Write", BlobWriteReq{
		Meta: "{}",
	}, res1)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/blob", cli.RemoteAddr().Addr()), strings.NewReader("test"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("token", res1.Token)

	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()

	var buf bytes.Buffer
	io.Copy(&buf, res2.Body)

	if buf.String() != "" {
		t.Fatalf("fail to write [%s]", buf.String())
	}
}
