package consulapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

type client struct {
	consulAddr  string
	consulIndex int64
	hostAddr    string
}

func (c *client) makeURL(path string) string {
	var index string
	if v := atomic.LoadInt64(&c.consulIndex); v > 0 {
		if strings.Contains(path, "?") {
			index = "&index=" + strconv.FormatInt(v, 10)
		} else {
			index = "?index=" + strconv.FormatInt(v, 10)
		}
	}
	return "http://" + c.consulAddr + path + index
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if str := resp.Header.Get("X-Consul-Index"); str != "" {
		if v, err := strconv.ParseInt(str, 10, 64); err == nil {
			atomic.StoreInt64(&c.consulIndex, v)
		}
	}

	return resp, nil
}

func (c *client) Request(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(data)
	}

	u := c.makeURL(path)
	req, err := http.NewRequest(method, u, r)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	return c.Do(req)
}

type Options struct {
	hostAddr   string
	consulAddr string
}

type Option func(*Options)

func WithConsulAddr(addr string) Option {
	return func(o *Options) {
		o.consulAddr = addr
	}
}

func newClient(opts ...Option) *client {
	options := Options{
		consulAddr: "localhost:8500",
	}
	if addr, err := lookupHostAddr(); err == nil {
		options.hostAddr = addr.String()
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &client{
		consulAddr: options.consulAddr,
		hostAddr:   options.hostAddr,
	}
}

func lookupHostAddr() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
