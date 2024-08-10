package znet

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	hdrUserAgentKey   = http.CanonicalHeaderKey("User-Agent")
	hdrUserAgentValue = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"
)

type (
	// RequestMiddleware type is for request middleware, called before a request is sent
	RequestMiddleware func(*Client, *Request) error

	// ResponseMiddleware type is for response middleware, called after a response has been received
	ResponseMiddleware func(*Client, *Response) error
)

type Client struct {
	BaseURL       string
	Header        http.Header
	scheme        string
	httpClient    *http.Client
	proxyURL      *url.URL
	beforeRequest []RequestMiddleware
	afterResponse []ResponseMiddleware
}

func (c *Client) SetHeader(header, value string) *Client {
	c.Header.Set(header, value)
	return c
}

func (c *Client) SetHeaders(headers map[string]string) *Client {
	for h, v := range headers {
		c.Header.Set(h, v)
	}
	return c
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

func New() *Client {
	return createClient(&http.Client{})
}

func NewWithClient(hc *http.Client) *Client {
	return createClient(hc)
}

// NewWithLocalAddr method creates a new client with given Local Address to dial from.
func NewWithLocalAddr(localAddr net.Addr) *Client {
	return createClient(&http.Client{
		Transport: createTransport(localAddr),
	})
}

func (c *Client) R() *Request {
	r := &Request{
		QueryParam: url.Values{},
		FormData:   url.Values{},
		Header:     http.Header{},
		client:     c,
	}
	return r
}

// Executes method executes the given `Request` object and returns response error.
func (c *Client) execute(req *Request) (response *Response, err error) {
	// request middlewares
	for _, f := range c.beforeRequest {
		if err = f(c, req); err != nil {
			return nil, err
		}
	}

	req.Time = time.Now()
	resp, err := c.httpClient.Do(req.RawRequest)

	response = &Response{
		Request:     req,
		RawResponse: resp,
	}
	response.setReceivedAt()

	if err != nil || req.notParseResponse {
		return nil, err
	}
	defer resp.Body.Close()

	response.body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Apply Response middleware
	for _, f := range c.afterResponse {
		if err = f(c, response); err != nil {
			break
		}
	}
	return
}

func createClient(hc *http.Client) *Client {
	if hc.Transport == nil {
		hc.Transport = createTransport(nil)
	}

	c := &Client{
		Header:     http.Header{},
		httpClient: hc,
	}

	// default before request middlewares
	c.beforeRequest = []RequestMiddleware{
		parseRequestURL,
		parseRequestHeader,
		createHTTPRequest,
	}

	return c
}
