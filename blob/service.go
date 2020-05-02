package blob

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/go-units"
)

type Service struct {
	mgr *Manager
}

func NewService(mgr *Manager) *Service {
	return &Service{mgr: mgr}
}

type BlobHdr struct {
	Read  bool
	Token string
}

func (s *Service) Serve(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token := r.Header.Get("token")

	if r.Method == "POST" {
		length := r.ContentLength
		if length > 3*units.GB {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fd, err := s.mgr.Write(token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = io.Copy(fd, r.Body)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")

		fd, err := s.mgr.Read(token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer fd.Close()

		info, err := fd.Stat()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Length", fmt.Sprint(info.Size()))

		_, err = io.Copy(w, fd)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
}

type BlobReadReq struct {
	Id string
}

type BlobReadRes struct {
	Token string
}

func (s *Service) Read(ctx context.Context, req *BlobReadReq, res *BlobReadRes) error {
	id, err := s.mgr.ReadToken(req.Id)
	res.Token = id
	return err
}

type BlobWriteReq struct {
	Meta string
}

type BlobWriteRes struct {
	Token string
}

func (s *Service) Write(ctx context.Context, req *BlobWriteReq, res *BlobWriteRes) error {
	id, err := s.mgr.WriteToken(req.Meta)
	res.Token = id
	return err
}

type BlobUpdateReq struct {
	Id   string
	Meta string
}

type BlobUpdateRes struct {
}

func (s *Service) Update(ctx context.Context, req *BlobUpdateReq, res *BlobUpdateRes) error {
	return s.mgr.Update(req.Id, req.Meta)
}

type BlobListReq struct {
	Query string
}

type BlobListRes struct {
	Blobs []Blob
}

func (s *Service) List(ctx context.Context, req *BlobListReq, res *BlobListRes) error {
	r, err := s.mgr.List(req.Query)
	if err != nil {
		return err
	}

	res.Blobs = r
	return nil
}

type BlobGetReq struct {
	Id string
}

type BlobGetRes struct {
	Blob
}

func (s *Service) Get(ctx context.Context, req *BlobGetReq, res *BlobGetRes) error {
	meta, err := s.mgr.GetMeta(req.Id)
	if err != nil {
		return err
	}

	res.Id = req.Id
	res.Meta = meta
	return nil
}

type BlobDeleteReq struct {
	Id string
}

type BlobDeleteRes struct {
}

func (s *Service) Delete(ctx context.Context, req *BlobDeleteReq, res *BlobDeleteRes) error {
	err := s.mgr.Delete(req.Id)
	if err != nil {
		return err
	}

	return nil
}
