package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/xhebox/chrootd/api"
)

type imageServer struct {
	api.UnimplementedImageServer
	imagePath string
}

func newImageServer(path string) *imageServer {
	return &imageServer{imagePath: path}
}

func (c *imageServer) Upload(srv api.Image_UploadServer) error {
	var name string
	var req *api.ImageUploadReq
	var meta, rootfs *os.File
	var err error

	res := &api.ImageUploadRes{}
	defer srv.SendAndClose(res)

	for {
		req, err = srv.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		switch v := req.Pkt.(type) {
		case *api.ImageUploadReq_Name:
			// TODO: log error
			name, err = filepath.Abs(path.Join(c.imagePath, path.Base(v.Name)))
			if err != nil {
				res.Reason = "invalid path"
				return nil
			}

			err = os.RemoveAll(name)
			if err != nil {
				res.Reason = "fail to create image"
				return nil
			}

			err = os.MkdirAll(name, 0755)
			if err != nil {
				res.Reason = "fail to create image"
				return nil
			}

			meta, err = os.OpenFile(path.Join(name, "meta"), os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				res.Reason = "fail to open meta file"
				return nil
			}
			defer meta.Close()

			rootfs, err = os.OpenFile(path.Join(name, "rootfs"), os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				res.Reason = "fail to open rootfs file"
				return nil
			}
			defer rootfs.Close()
		case *api.ImageUploadReq_MetadataJson:
			if meta == nil {
				res.Reason = "did not set name"
				return nil
			}

			_, err = io.Copy(meta, bytes.NewReader(v.MetadataJson))
			if err != nil {
				res.Reason = "fail to write meta"
				return nil
			}
		case *api.ImageUploadReq_RootfsTarball:
			if rootfs == nil {
				res.Reason = "did not set name"
				return nil
			}

			_, err = io.Copy(rootfs, bytes.NewReader(v.RootfsTarball))
			if err != nil {
				res.Reason = "fail to write rootfs"
				return nil
			}
		}
	}
}

func (c *imageServer) Delete(ctx context.Context, req *api.ImageDeleteReq) (*api.ImageDeleteRes, error) {
	res := &api.ImageDeleteRes{}

	name, err := filepath.Abs(path.Join(c.imagePath, path.Base(req.Name)))
	if err != nil {
		return res, nil
	}

	err = os.RemoveAll(name)
	if err != nil {
		return res, nil
	}

	return res, nil
}

func (c *imageServer) List(req *api.ImageListReq, srv api.Image_ListServer) error {
	dir, err := os.Open(c.imagePath)
	if err != nil {
		return nil
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil
	}

	for i := range names {
		res := &api.ImageListRes{Name: names[i]}

		name, err := filepath.Abs(path.Join(c.imagePath, names[i], "meta"))
		if err == nil {
			res.MetadataJson, _ = ioutil.ReadFile(name)
		}

		if err := srv.Send(res); err != nil {
			return err
		}
	}

	return nil
}
