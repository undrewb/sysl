// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/anz-bank/sysl/pkg/lsp/framework/jsonrpc2"
	"github.com/anz-bank/sysl/pkg/lsp/framework/xcontext"
	errors "golang.org/x/xerrors"
)

var (
	// RequestCancelledError should be used when a request is cancelled early.
	RequestCancelledError = jsonrpc2.NewError(-32800, "JSON RPC cancelled")
)

type ClientCloser interface {
	Client
	io.Closer
}

type connSender interface {
	io.Closer

	Notify(ctx context.Context, method string, params interface{}) error
	Call(ctx context.Context, method string, params, result interface{}) error
}

type clientDispatcher struct {
	sender connSender
}

func (c *clientDispatcher) Close() error {
	return c.sender.Close()
}

// ClientDispatcher returns a Client that dispatches LSP requests across the
// given jsonrpc2 connection.
func ClientDispatcher(conn jsonrpc2.Conn) ClientCloser {
	return &clientDispatcher{sender: clientConn{conn}}
}

type clientConn struct {
	conn jsonrpc2.Conn
}

func (c clientConn) Close() error {
	return c.conn.Close()
}

func (c clientConn) Notify(ctx context.Context, method string, params interface{}) error {
	return c.conn.Notify(ctx, method, params)
}

func (c clientConn) Call(ctx context.Context, method string, params interface{}, result interface{}) error {
	id, err := c.conn.Call(ctx, method, params, result)
	if ctx.Err() != nil {
		cancelCall(ctx, c, id)
	}
	return err
}

// ServerDispatcher returns a Server that dispatches LSP requests across the
// given jsonrpc2 connection.
func ServerDispatcher(conn jsonrpc2.Conn) Server {
	return &serverDispatcher{sender: clientConn{conn}}
}

type serverDispatcher struct {
	sender connSender
}

func ClientHandler(client Client, handler jsonrpc2.Handler) jsonrpc2.Handler {
	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		if ctx.Err() != nil {
			ctx := xcontext.Detach(ctx)
			return reply(ctx, nil, RequestCancelledError)
		}
		handled, err := clientDispatch(ctx, client, reply, req)
		if handled || err != nil {
			return err
		}
		return handler(ctx, reply, req)
	}
}

func ServerHandler(server Server, handler jsonrpc2.Handler) jsonrpc2.Handler {
	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		if ctx.Err() != nil {
			ctx := xcontext.Detach(ctx)
			return reply(ctx, nil, RequestCancelledError)
		}
		handled, err := serverDispatch(ctx, server, reply, req)
		if handled || err != nil {
			return err
		}
		//TODO: This code is wrong, it ignores handler and assumes non standard
		// request handles everything
		// non standard request should just be a layered handler.
		var params interface{}
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return sendParseError(ctx, reply, err)
		}
		resp, err := server.NonstandardRequest(ctx, req.Method(), params)
		return reply(ctx, resp, err)

	}
}

func Handlers(handler jsonrpc2.Handler) jsonrpc2.Handler {
	return CancelHandler(
		jsonrpc2.AsyncHandler(
			jsonrpc2.MustReplyHandler(handler)))
}

func CancelHandler(handler jsonrpc2.Handler) jsonrpc2.Handler {
	handler, canceller := jsonrpc2.CancelHandler(handler)
	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		if req.Method() != "$/cancelRequest" {
			// TODO(iancottrell): See if we can generate a reply for the request to be cancelled
			// at the point of cancellation rather than waiting for gopls to naturally reply.
			// To do that, we need to keep track of whether a reply has been sent already and
			// be careful about racing between the two paths.
			// TODO(iancottrell): Add a test that watches the stream and verifies the response
			// for the cancelled request flows.
			replyWithDetachedContext := func(ctx context.Context, resp interface{}, err error) error {
				// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#cancelRequest
				if ctx.Err() != nil && err == nil {
					err = RequestCancelledError
				}
				ctx = xcontext.Detach(ctx)
				return reply(ctx, resp, err)
			}
			return handler(ctx, replyWithDetachedContext, req)
		}
		var params CancelParams
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return sendParseError(ctx, reply, err)
		}
		if n, ok := params.ID.(float64); ok {
			canceller(jsonrpc2.NewIntID(int64(n)))
		} else if s, ok := params.ID.(string); ok {
			canceller(jsonrpc2.NewStringID(s))
		} else {
			return sendParseError(ctx, reply, fmt.Errorf("request ID %v malformed", params.ID))
		}
		return reply(ctx, nil, nil)
	}
}

func Call(ctx context.Context, conn jsonrpc2.Conn, method string, params interface{}, result interface{}) error {
	id, err := conn.Call(ctx, method, params, result)
	if ctx.Err() != nil {
		cancelCall(ctx, clientConn{conn}, id)
	}
	return err
}

func cancelCall(ctx context.Context, sender connSender, id jsonrpc2.ID) {
	ctx = xcontext.Detach(ctx)
	// Note that only *jsonrpc2.ID implements json.Marshaler.
	sender.Notify(ctx, "$/cancelRequest", &CancelParams{ID: &id})
}

func sendParseError(ctx context.Context, reply jsonrpc2.Replier, err error) error {
	return reply(ctx, nil, errors.Errorf("%w: %s", jsonrpc2.ErrParse, err))
}
