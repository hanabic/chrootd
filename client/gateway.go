package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
)

type HTTPListener struct {
	conn chan *requestConn
}

func NewHTTPListener() *HTTPListener {
	return &HTTPListener{conn: make(chan *requestConn)}
}

func (ln *HTTPListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	args, err := server.HTTPRequest2RpcxRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	rp, wp := io.Pipe()
	conn := &requestConn{rd: bytes.NewReader(args.Encode()), wt: wp}
	ln.conn <- conn

	reply, err := protocol.Read(rp)
	if err != nil {
		w.WriteHeader(http.StatusUnavailableForLegalReasons)
	}

	wh := w.Header()
	wh.Set(server.XMessageType, fmt.Sprint(reply.Header.MessageType()))
	wh.Set(server.XMessageID, fmt.Sprint(reply.Header.Seq()))
	wh.Set(server.XMessageStatusType, fmt.Sprint(reply.Header.MessageStatusType()))
	wh.Set(server.XSerializeType, fmt.Sprint(reply.Header.SerializeType()))
	io.Copy(w, bytes.NewReader(reply.Payload))
}

func (ln *HTTPListener) Accept() (net.Conn, error) {
	c, ok := <-ln.conn
	if !ok {
		return nil, errors.New("listener closed")
	}
	return c, nil
}

func (ln *HTTPListener) Close() error {
	close(ln.conn)
	return nil
}

func (ln *HTTPListener) Addr() net.Addr {
	return &net.IPAddr{}
}

type requestConn struct {
	rd io.Reader
	wt io.Writer
}

func (c *requestConn) Read(b []byte) (n int, err error) {
	return c.rd.Read(b)
}

func (c *requestConn) Write(b []byte) (n int, err error) {
	return c.wt.Write(b)
}

func (c *requestConn) Close() error {
	return nil
}

func (c *requestConn) LocalAddr() net.Addr {
	return &net.IPAddr{}
}

func (c *requestConn) RemoteAddr() net.Addr {
	return &net.IPAddr{}
}

func (c *requestConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *requestConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *requestConn) SetWriteDeadline(t time.Time) error {
	return nil
}
