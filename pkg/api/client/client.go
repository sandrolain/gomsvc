package client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"
)

type Init struct {
	Params   map[string]string
	Query    map[string]string
	Headers  map[string]string
	FormData map[string]string
	Files    map[string]string
	Body     interface{}
	Timeout  time.Duration
}

type Response[T any] struct {
	Resty *resty.Response
	Body  T
}

func applyInit(ctx context.Context, init *Init) *resty.Request {
	client := resty.New()
	if init.Timeout > 0 {
		client.SetTimeout(init.Timeout)
	}

	r := client.R().SetContext(ctx)
	// EnableTrace().

	if init.Headers != nil && len(init.Headers) > 0 {
		r.SetHeaders(init.Headers)
	}
	if init.Query != nil && len(init.Query) > 0 {
		r.SetQueryParams(init.Query)
	}
	if init.Params != nil && len(init.Params) > 0 {
		r.SetPathParams(init.Params)
	}
	if init.FormData != nil && len(init.FormData) > 0 {
		r.SetFormData(init.FormData)
	}
	if init.Files != nil && len(init.Files) > 0 {
		r.SetFiles(init.Files)
	}
	if init.Body != nil {
		r.SetBody(init.Body)
	}

	return r
}

func GetJSON[R any](ctx context.Context, url string, init Init) (res Response[R], err error) {
	req := applyInit(ctx, &init)
	if resp, err := req.Get(url); err == nil {
		var body R
		err = json.Unmarshal(resp.Body(), &body)
		res.Resty = resp
		res.Body = body
	}
	return
}

func GetBytes(ctx context.Context, url string, init Init) (res Response[[]byte], err error) {
	req := applyInit(ctx, &init)
	if resp, err := req.Get(url); err == nil {
		res.Resty = resp
		res.Body = resp.Body()
	}
	return
}

func PostBytes(ctx context.Context, url string, init Init) (res Response[[]byte], err error) {
	req := applyInit(ctx, &init)
	if resp, err := req.Post(url); err == nil {
		res.Resty = resp
		res.Body = resp.Body()
	}
	return
}
