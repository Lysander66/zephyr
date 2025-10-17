package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// JSON-RPC	https://www.jsonrpc.org/specification
const version2 = "2.0"

type RequestID interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Request struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      any    `json:"id"`
}

type Notification struct {
	Version string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type Response struct {
	Version string `json:"jsonrpc"`
	Result  any    `json:"result"`
	Error   *Error `json:"error,omitempty"`
	ID      any    `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return strconv.Itoa(e.Code) + " " + e.Message
}

func (r *Response) GetString() (string, error) {
	val, ok := r.Result.(string)
	if !ok {
		return "", fmt.Errorf("couldn't parse string from %s", r.Result)
	}

	return val, nil
}

func (r *Response) GetAny(v any) error {
	data, err := json.Marshal(r.Result)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func NewRequest[T RequestID](method string, params any, id T) *Request {
	return &Request{Version: version2, Method: method, Params: params, ID: id}
}

type Client struct {
	endpoint   string
	httpClient *http.Client
	timeout    time.Duration
}

type Option func(o *Client)

func HttpClient(hc *http.Client) Option {
	return func(o *Client) { o.httpClient = hc }
}

func Timeout(timeout time.Duration) Option {
	return func(o *Client) { o.timeout = timeout }
}

func NewClient(endpoint string, opts ...Option) *Client {
	c := &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{},
		timeout:    30 * time.Second, // default 30 seconds
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Call(ctx context.Context, request *Request) (*Response, error) {
	var payload bytes.Buffer
	if err := json.NewEncoder(&payload).Encode(request); err != nil {
		return nil, err
	}

	// If the passed ctx has no timeout, use the default timeout
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, &payload)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResponse, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Check HTTP status code
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %v for method %v", httpResponse.StatusCode, request.Method)
	}

	var rpcResponse *Response
	err = json.NewDecoder(httpResponse.Body).Decode(&rpcResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON response for %v: %v", request.Method, err)
	}

	if rpcResponse == nil {
		return nil, fmt.Errorf("empty response for %v", request.Method)
	}

	return rpcResponse, nil
}
