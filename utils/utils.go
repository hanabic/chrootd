package utils

import (
	"io"
	"os"
)

func PathExist(path string) bool {
	_, e := os.Stat(path)
	return !os.IsNotExist(e)
}

type NopReadCloser struct {
	io.Reader
	closed bool
}

func (n *NopReadCloser) Read(p []byte) (int, error) {
	if n.closed {
		return 0, io.EOF
	}
	return n.Reader.Read(p)
}

func (n *NopReadCloser) Close() error {
	n.closed = true
	return nil
}

func NewNopReadCloser(rd io.Reader) io.ReadCloser {
	return &NopReadCloser{Reader: rd, closed: false}
}
