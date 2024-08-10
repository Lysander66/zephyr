package znet

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	URL              string
	Method           string
	QueryParam       url.Values
	FormData         url.Values
	Header           http.Header
	Time             time.Time
	Attempt          int
	RawRequest       *http.Request
	ctx              context.Context
	client           *Client
	notParseResponse bool
}

func (r *Request) SetContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

// SetDoNotParseResponse
//
// Note: Response middlewares are not applicable, if you use this option.
func (r *Request) SetDoNotParseResponse(parse bool) *Request {
	r.notParseResponse = parse
	return r
}

func (r *Request) SetHeader(header, value string) *Request {
	r.Header.Set(header, value)
	return r
}

func (r *Request) SetHeaders(headers map[string]string) *Request {
	for h, v := range headers {
		r.SetHeader(h, v)
	}
	return r
}

// SetQueryParam method sets single parameter and its value in the current request.
// It will be formed as query string for the request.
//
// For Example: `search=kitchen%20papers&size=large` in the URL after `?` mark.
//
//	client.R().
//		SetQueryParam("search", "kitchen papers").
//		SetQueryParam("size", "large")
//
// Also you can override query params value, which was set at client instance level.
func (r *Request) SetQueryParam(param, value string) *Request {
	r.QueryParam.Set(param, value)
	return r
}

func (r *Request) SetQueryParams(params map[string]string) *Request {
	for p, v := range params {
		r.SetQueryParam(p, v)
	}
	return r
}

func (r *Request) SetFormData(data map[string]string) *Request {
	for k, v := range data {
		r.FormData.Set(k, v)
	}
	return r
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// HTTP verb method starts here
//_______________________________________________________________________

// Get method does GET HTTP request. It's defined in section 4.3.1 of RFC7231.
func (r *Request) Get(url string) (*Response, error) {
	return r.Execute(http.MethodGet, url)
}

// Head method does HEAD HTTP request. It's defined in section 4.3.2 of RFC7231.
func (r *Request) Head(url string) (*Response, error) {
	return r.Execute(http.MethodHead, url)
}

func (r *Request) Send() (*Response, error) {
	return r.Execute(r.Method, r.URL)
}

func (r *Request) GetWithRetries(url string, options ...Option) (*Response, error) {
	return r.ExecuteWithRetries(http.MethodGet, url, options...)
}

// Execute method performs the HTTP request with given HTTP method and URL for current `Request`.
func (r *Request) Execute(method, url string) (resp *Response, err error) {
	r.Method = method
	r.URL = url
	return r.client.execute(r)
}

func (r *Request) ExecuteWithRetries(method, url string, options ...Option) (resp *Response, err error) {
	r.Method = method
	r.URL = url

	err = Backoff(
		func() (*Response, error) {
			r.Attempt++

			resp, err = r.client.execute(r)
			if err != nil {
				slog.Error("Backoff", "err", err, "Attempt", r.Attempt)
			}

			return resp, err
		},
		options...,
	)

	if err != nil {
		slog.Error("ExecuteWithRetries", "err", err)
	}

	return resp, err
}
