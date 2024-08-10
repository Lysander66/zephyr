package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
}

type Option func(o *Client)

func HttpClient(hc *http.Client) Option {
	return func(o *Client) { o.httpClient = hc }
}

func NewClient(endpoint string, opts ...Option) *Client {
	c := &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Call(request *Request) (*Response, error) {
	var payload bytes.Buffer
	if err := json.NewEncoder(&payload).Encode(request); err != nil {
		return nil, err
	}
	httpResponse, err := c.httpClient.Post(c.endpoint, "application/json", &payload)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	var rpcResponse *Response
	err = json.NewDecoder(httpResponse.Body).Decode(&rpcResponse)
	if err != nil {
		return nil, err
	}

	if rpcResponse == nil {
		return nil, fmt.Errorf("rpc call %v() status code: %v. rpc response missing", request.Method, httpResponse.StatusCode)
	}
	if rpcResponse.Error != nil {
		return nil, fmt.Errorf("rpc call %v() status code: %v. rpc response error: %v", request.Method, httpResponse.StatusCode, rpcResponse.Error)
	}

	return rpcResponse, nil
}
