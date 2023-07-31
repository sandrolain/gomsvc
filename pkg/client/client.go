package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/sandrolain/gomscv/pkg/body"
)

type ClientOptions[T any] struct {
	Method      string
	Url         string
	ContentType string
	Request     *Request[T]
}

type H [][2]string
type P [][2]string
type Q [][2]string

type Request[T any] struct {
	Headers H
	Params  P
	Query   Q
	Body    *T
}

func CreateClient[T any, R any](opts ClientOptions[T]) func(...Request[T]) (R, *http.Response, error) {
	return func(request ...Request[T]) (resData R, res *http.Response, err error) {
		reqOpts := getFirst(request)
		client := &http.Client{}
		headers := reqOpts.Headers
		if opts.Request != nil && len(opts.Request.Headers) > 0 {
			headers = mergeCouples(headers, opts.Request.Headers)
		}
		params := reqOpts.Params
		if opts.Request != nil && len(opts.Request.Params) > 0 {
			params = mergeCouples(params, opts.Request.Params)
		}
		query := reqOpts.Query
		if opts.Request != nil && len(opts.Request.Query) > 0 {
			query = mergeCouples(query, opts.Request.Query)
		}
		reqUrl := opts.Url
		if len(params) > 0 {
			reqUrl, err = replaceParams(reqUrl, params)
			if err != nil {
				return
			}
		}
		if len(query) > 0 {
			reqUrl, err = applyQuery(reqUrl, query)
			if err != nil {
				return
			}
		}
		method := opts.Method
		if method == "" {
			method = "GET"
		}
		var bodyReader *bytes.Reader
		if reqOpts.Body != nil {
			var data []byte
			data, err = body.MarshalBody(opts.ContentType, reqOpts.Body)
			if err != nil {
				return
			}
			if len(opts.ContentType) > 0 {
				headers = append(headers, [2]string{"Content-Type", opts.ContentType})
			}
			bodyReader = bytes.NewReader(data)
		}
		req, err := http.NewRequest(method, reqUrl, bodyReader)
		if err != nil {
			return
		}
		if len(headers) > 0 {
			applyHeaders(req, headers)
		}
		res, err = client.Do(req)
		if err != nil {
			return
		}
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			var resBody []byte
			resBody, err = streamToByte(res.Body)
			if err != nil {
				return
			}
			resData, err = body.UnmarshalBody[R](res.Header.Get("content-type"), resBody)
		}
		if res.StatusCode >= 400 {
			err = &ResponseError{StatusCode: res.StatusCode}
		}
		return
	}
}

func streamToByte(stream io.Reader) (data []byte, err error) {
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(stream)
	if err == nil {
		data = buf.Bytes()
	}
	return
}

func getFirst[T any](list []T) (first T) {
	if len(list) > 0 {
		first = list[0]
	}
	return
}

func replaceParams(url string, params P) (res string, err error) {
	var re *regexp.Regexp
	for _, value := range params {
		re, err = regexp.Compile(":" + regexp.QuoteMeta(value[0]) + "([^a-zA-Z0-9]|$)")
		if err != nil {
			return
		}
		url = re.ReplaceAllString(url, value[1]+"$1")
	}
	return
}

func applyQuery(reqUrl string, query Q) (res string, err error) {
	var u *url.URL
	u, err = url.Parse(reqUrl)
	if err != nil {
		return
	}
	q := u.Query()
	for _, value := range query {
		q.Add(value[0], value[1])
	}
	u.RawQuery = q.Encode()
	res = u.String()
	return
}

func applyHeaders(req *http.Request, headers H) {
	h := &req.Header
	for _, value := range headers {
		h.Add(value[0], value[1])
	}
}

func mergeCouples(m1 [][2]string, m2 [][2]string) [][2]string {
	merged := make([][2]string, len(m1)+len(m2))
	i := 0
	for _, v := range m1 {
		merged[i] = v
		i++
	}
	for _, v := range m2 {
		merged[i] = v
		i++
	}
	return merged
}

func mergeMaps(m1 map[string]string, m2 map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}

type ResponseError struct {
	StatusCode int
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("HTTP Error %d", e.StatusCode)
}
