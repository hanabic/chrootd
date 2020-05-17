package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/osamingo/jsonrpc"
	"github.com/smallnest/rpcx/share"
)

type Gateway struct {
	cli Client
}

func NewGateway(network, addr string) (*Gateway, error) {
	cli, err := newRpcxClient(network, addr)
	if err != nil {
		return nil, err
	}
	return &Gateway{cli: cli}, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rpcs, batch, err := jsonrpc.ParseRequest(r)
	if err != nil {
		err := jsonrpc.SendResponse(w, []*jsonrpc.Response{
			{
				Version: jsonrpc.Version,
				Error:   err,
			},
		}, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	resp := make([]*jsonrpc.Response, len(rpcs))

	for i := range rpcs {
		var qs []interface{}
		var rq, rs interface{}
		var e error
		ctx := r.Context()

		_path := strings.SplitN(rpcs[i].Method, ".", 2)
		if len(_path) < 2 {
			goto exit
		}

		e = json.Unmarshal(*rpcs[i].Params, &qs)
		if e != nil {
			goto exit
		}

		if len(qs) == 1 {
			rq = qs[0]
		} else if len(qs) >= 2 {
			if meta, ok := qs[0].(map[string]interface{}); ok {
				m := make(map[string]string)
				for k, v := range meta {
					m[k] = fmt.Sprint(v)
				}
				ctx = context.WithValue(ctx, share.ReqMetaDataKey, m)
				if len(qs) == 2 {
					rq = qs[1]
				} else {
					rq = qs[1:]
				}
			}
		} else {
			rq = qs
		}

		e = g.cli.Call(ctx, _path[0], _path[1], rq, &rs)
		if e != nil {
			goto exit
		}

		resp[i] = &jsonrpc.Response{
			Version: jsonrpc.Version,
			Result:  rs,
		}
		continue

	exit:
		resp[i] = &jsonrpc.Response{
			Version: jsonrpc.Version,
			Error:   jsonrpc.ErrInternal(),
		}
	}

	if err := jsonrpc.SendResponse(w, resp, batch); err != nil {
		fmt.Fprint(w, "Failed to encode result objects")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (g *Gateway) Close() error {
	return g.cli.Close()
}
