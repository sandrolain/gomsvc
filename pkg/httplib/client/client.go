package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Common errors
var (
	ErrNilContext       = errors.New("context cannot be nil")
	ErrEmptyURL        = errors.New("URL cannot be empty")
	ErrInvalidResponse = errors.New("invalid response")
	ErrRequestFailed   = errors.New("request failed")
	ErrInvalidTimeout  = errors.New("timeout cannot be negative")
	ErrInvalidRetryCount = errors.New("retry count cannot be negative")
	ErrInvalidRetryWait = errors.New("retry wait cannot be negative")
)

type Init struct {
	Params      map[string]string
	Query       map[string]string
	Headers     map[string]string
	FormData    map[string]string
	Files       map[string]string
	Body        interface{}
	Timeout     time.Duration
	RetryCount  int
	RetryWait   time.Duration
	BaseURL     string
}

type Response[T any] struct {
	StatusCode int
	Headers    http.Header
	Resty     *resty.Response
	Body      T
}

func validateInit(init *Init) error {
	if init == nil {
		return nil
	}
	if init.Timeout < 0 {
		return ErrInvalidTimeout
	}
	if init.RetryCount < 0 {
		return ErrInvalidRetryCount
	}
	if init.RetryWait < 0 {
		return ErrInvalidRetryWait
	}
	return nil
}

func validateRequest(ctx context.Context, url string) error {
	if ctx == nil {
		return ErrNilContext
	}
	if url == "" {
		return ErrEmptyURL
	}
	return nil
}

func applyInit(ctx context.Context, init *Init) (*resty.Request, error) {
	if err := validateRequest(ctx, init.BaseURL); err != nil {
		return nil, err
	}
	if err := validateInit(init); err != nil {
		return nil, err
	}

	client := resty.New()
	if init != nil {
		if init.Timeout > 0 {
			client.SetTimeout(init.Timeout)
		}
		if init.RetryCount > 0 {
			client.SetRetryCount(init.RetryCount)
			if init.RetryWait > 0 {
				client.SetRetryWaitTime(init.RetryWait)
			}
		}
		if init.BaseURL != "" {
			client.SetBaseURL(init.BaseURL)
		}
	}

	r := client.R().SetContext(ctx)

	if init == nil {
		return r, nil
	}

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

	return r, nil
}

func processResponse[T any](resp *resty.Response, err error) (Response[T], error) {
	var result Response[T]
	if err != nil {
		return result, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	if resp == nil {
		return result, fmt.Errorf("%w: response is nil", ErrInvalidResponse)
	}

	result.StatusCode = resp.StatusCode()
	result.Headers = resp.Header()
	result.Resty = resp

	// Check for server errors (5xx)
	if resp.StatusCode() >= 500 {
		return result, fmt.Errorf("%w: server error %d: %s", ErrRequestFailed, resp.StatusCode(), resp.String())
	}

	// Check for client errors (4xx)
	if resp.StatusCode() >= 400 {
		return result, fmt.Errorf("%w: client error %d: %s", ErrRequestFailed, resp.StatusCode(), resp.String())
	}

	return result, nil
}

func GetJSON[R any](ctx context.Context, url string, init Init) (Response[*R], error) {
	req, err := applyInit(ctx, &init)
	if err != nil {
		return Response[*R]{}, err
	}

	resp, err := req.Get(init.BaseURL + url)
	result, err := processResponse[*R](resp, err)
	if err != nil {
		return result, err
	}

	if resp == nil || len(resp.Body()) == 0 {
		return result, fmt.Errorf("%w: empty response body", ErrInvalidResponse)
	}

	var body R
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return result, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	result.Body = &body
	return result, nil
}

func GetBytes(ctx context.Context, url string, init Init) (Response[[]byte], error) {
	req, err := applyInit(ctx, &init)
	if err != nil {
		return Response[[]byte]{}, err
	}

	resp, err := req.Get(init.BaseURL + url)
	result, err := processResponse[[]byte](resp, err)
	if err != nil {
		return result, err
	}

	if resp == nil || len(resp.Body()) == 0 {
		return result, fmt.Errorf("%w: empty response body", ErrInvalidResponse)
	}

	result.Body = resp.Body()
	return result, nil
}

func PostJSON[R any](ctx context.Context, url string, init Init) (Response[*R], error) {
	req, err := applyInit(ctx, &init)
	if err != nil {
		return Response[*R]{}, err
	}

	resp, err := req.Post(init.BaseURL + url)
	result, err := processResponse[*R](resp, err)
	if err != nil {
		return result, err
	}

	if resp == nil || len(resp.Body()) == 0 {
		return result, fmt.Errorf("%w: empty response body", ErrInvalidResponse)
	}

	var body R
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return result, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	result.Body = &body
	return result, nil
}

func PostBytes(ctx context.Context, url string, init Init) (Response[[]byte], error) {
	req, err := applyInit(ctx, &init)
	if err != nil {
		return Response[[]byte]{}, err
	}

	resp, err := req.Post(init.BaseURL + url)
	result, err := processResponse[[]byte](resp, err)
	if err != nil {
		return result, err
	}

	if resp == nil || len(resp.Body()) == 0 {
		return result, fmt.Errorf("%w: empty response body", ErrInvalidResponse)
	}

	result.Body = resp.Body()
	return result, nil
}

func PutJSON[R any](ctx context.Context, url string, init Init) (Response[*R], error) {
	req, err := applyInit(ctx, &init)
	if err != nil {
		return Response[*R]{}, err
	}

	resp, err := req.Put(init.BaseURL + url)
	result, err := processResponse[*R](resp, err)
	if err != nil {
		return result, err
	}

	if resp == nil || len(resp.Body()) == 0 {
		return result, fmt.Errorf("%w: empty response body", ErrInvalidResponse)
	}

	var body R
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return result, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	result.Body = &body
	return result, nil
}

func DeleteJSON[R any](ctx context.Context, url string, init Init) (Response[*R], error) {
	req, err := applyInit(ctx, &init)
	if err != nil {
		return Response[*R]{}, err
	}

	resp, err := req.Delete(init.BaseURL + url)
	result, err := processResponse[*R](resp, err)
	if err != nil {
		return result, err
	}

	if resp == nil || len(resp.Body()) == 0 {
		return result, fmt.Errorf("%w: empty response body", ErrInvalidResponse)
	}

	var body R
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return result, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	result.Body = &body
	return result, nil
}

func PatchJSON[R any](ctx context.Context, url string, init Init) (Response[*R], error) {
	req, err := applyInit(ctx, &init)
	if err != nil {
		return Response[*R]{}, err
	}

	resp, err := req.Patch(init.BaseURL + url)
	result, err := processResponse[*R](resp, err)
	if err != nil {
		return result, err
	}

	if resp == nil || len(resp.Body()) == 0 {
		return result, fmt.Errorf("%w: empty response body", ErrInvalidResponse)
	}

	var body R
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return result, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	result.Body = &body
	return result, nil
}
