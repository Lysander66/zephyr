package znet

import (
	"net/http"
	"net/url"
	"strings"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Request Middleware(s)
//_______________________________________________________________________

func parseRequestURL(c *Client, r *Request) error {
	// Parsing request URL
	reqURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	// If Request.URL is relative path
	if !reqURL.IsAbs() {
		r.URL = reqURL.String()
		if len(r.URL) > 0 && r.URL[0] != '/' {
			r.URL = "/" + r.URL
		}

		baseURL := c.BaseURL
		reqURL, err = url.Parse(baseURL + r.URL)
		if err != nil {
			return err
		}
	}

	if reqURL.Scheme == "" && len(c.scheme) > 0 {
		reqURL.Scheme = c.scheme
	}

	// Adding Query Param
	query := make(url.Values)
	for k, v := range r.QueryParam {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	if len(query) > 0 {
		if IsStringEmpty(reqURL.RawQuery) {
			reqURL.RawQuery = query.Encode()
		} else {
			reqURL.RawQuery = reqURL.RawQuery + "&" + query.Encode()
		}
	}

	r.URL = reqURL.String()

	return nil
}

func parseRequestHeader(c *Client, r *Request) error {
	for k, v := range c.Header {
		if _, ok := r.Header[k]; ok {
			continue
		}
		r.Header[k] = v[:]
	}

	if IsStringEmpty(r.Header.Get(hdrUserAgentKey)) {
		r.Header.Set(hdrUserAgentKey, hdrUserAgentValue)
	}

	return nil
}

func createHTTPRequest(c *Client, r *Request) (err error) {
	r.RawRequest, err = http.NewRequest(r.Method, r.URL, nil)

	if err != nil {
		return
	}

	// Add headers into http request
	r.RawRequest.Header = r.Header

	// Use context if it was specified
	if r.ctx != nil {
		r.RawRequest = r.RawRequest.WithContext(r.ctx)
	}

	return
}

func IsStringEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}
