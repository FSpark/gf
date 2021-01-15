package ghttp

import (
	"github.com/gogf/gf/errors/gerror"
	"net/http"
)

const gfHTTPClientMiddlewareKey = "__gfHttpClientMiddlewareKey"

// Use Add middleware to client
func (c *Client) Use(handlers ...ClientHandlerFunc) *Client {
	newClient := c
	if c.parent == nil {
		newClient = c.Clone()
	}

	newClient.middlewareHandler = append(newClient.middlewareHandler, handlers...)
	return newClient
}

// MiddlewareNext call next middleware
// this is should only be call in ClientHandlerFunc
func (c *Client) MiddlewareNext(req *http.Request) (*ClientResponse, error) {
	m, ok := req.Context().Value(gfHTTPClientMiddlewareKey).(*clientMiddleware)
	if ok {
		resp, err := m.Next(req)
		return resp, err
	}
	return c.callRequest(req)
}

// MiddlewareAbort stop call after all middleware, so it will not send http request
// this is should only be call in ClientHandlerFunc
func (c *Client) MiddlewareAbort(req *http.Request) (*ClientResponse, error) {
	m, ok := req.Context().Value(gfHTTPClientMiddlewareKey).(*clientMiddleware)
	if ok {
		m.Abort()
		return m.resp, m.err
	}
	return nil, gerror.New("http request abort")
}

// ClientHandlerFunc middleware handler func
type ClientHandlerFunc = func(c *Client, r *http.Request) (*ClientResponse, error)

// clientMiddleware is the plugin for http client request workflow management.
type clientMiddleware struct {
	client       *Client             // http client
	handlers     []ClientHandlerFunc // mdl handlers
	handlerIndex int                 // current handler index
	abort        bool                // abort call after handlers
	resp         *ClientResponse     // save resp
	err          error               // save err
}

// Next call next middleware handler, if abort,
func (m *clientMiddleware) Next(req *http.Request) (resp *ClientResponse, err error) {
	if m.abort || m.err != nil {
		return m.resp, m.err
	}
	if m.handlerIndex < len(m.handlers) {
		m.handlerIndex++
		resp, err = m.handlers[m.handlerIndex](m.client, req)
		m.resp = resp
		m.err = err
	}
	return
}

func (m *clientMiddleware) Abort() {
	m.abort = true
	if m.err == nil {
		m.err = gerror.New("http request abort")
	}
}
