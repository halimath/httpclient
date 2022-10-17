package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RequestInterceptor defines an interface for types that can intercept a
// http.Request to make modifications or otherwise apply some kind of
// treatment. A request may be modified (i.e. by adding a header or setting
// the request's body), replace (by returning a new request) or reject by
// returning an error.
type RequestInterceptor interface {
	// InterceptRequest intercepts r and returns either r or a new request or
	// an error. If the returned error is not nil, the returned request is
	// ignored and the request will be aborted.
	InterceptRequest(r *http.Request) (*http.Request, error)
}

// RequestInterceptorFunc defines a convenience type to implement a
// RequestInterceptor based on a bare function.
type RequestInterceptorFunc func(r *http.Request) (*http.Request, error)

func (f RequestInterceptorFunc) InterceptRequest(r *http.Request) (*http.Request, error) {
	return f(r)
}

// RequestInterceptorOption defines a type implementing both ClientOption and
// RequestOption used to pass a RequestInterceptor to either a Client or one
// of the request making functions.
type RequestInterceptorOption struct {
	RequestInterceptor
}

func (RequestInterceptorOption) clientOpt() {}
func (RequestInterceptorOption) reqOpt()    {}

// WithRequestInterceptor creates a RequestInterceptorOption holding r.
func WithRequestInterceptor(r RequestInterceptor) RequestInterceptorOption {
	return RequestInterceptorOption{r}
}

// WithRequestInterceptorFunc creates a RequestInterceptorOption holding f
// converted to a WithRequestInterceptorFunc.
func WithRequestInterceptorFunc(f func(*http.Request) (*http.Request, error)) RequestInterceptorOption {
	return RequestInterceptorOption{RequestInterceptorFunc(f)}
}

// WithRequestHeader creates a RequestInterceptor that adds header with value
// to the given request and returns that as a RequestInterceptorOption.
func WithRequestHeader(header, value string) RequestInterceptorOption {
	return WithRequestInterceptorFunc(func(r *http.Request) (*http.Request, error) {
		r.Header.Set(header, value)
		return r, nil
	})
}

type readCloser struct {
	r io.Reader
}

func (rc *readCloser) Close() error                     { return nil }
func (rc *readCloser) Read(p []byte) (n int, err error) { return rc.r.Read(p) }

func withBody(r io.Reader, contentType string, length int64) RequestInterceptorFunc {
	return func(req *http.Request) (*http.Request, error) {
		if req.Body != nil {
			if err := req.Body.Close(); err != nil {
				return req, err
			}
		}

		if c, ok := r.(io.ReadCloser); ok {
			req.Body = c
		} else {
			req.Body = &readCloser{r}
		}

		req.Header.Set("Content-Type", contentType)
		req.ContentLength = length

		return req, nil
	}
}

func WithBody(r io.Reader, contentType string, length int64) RequestInterceptorOption {
	return WithRequestInterceptorFunc(withBody(r, contentType, length))
}

// WithJSON uses value as a JSON encoded request body. It returns a
// RequestInterceptor wrapped in a RequestInterceptorOption that marshals value
// and sets it as the request`s Body. If the request had a previous non-nil
// Body this value is closed before. The interceptor also sets the
// Content-Type request header as well as the Content-Length header.
// Any error produced by json.Marshal or a previous request body's Close method
// is returned and aborts the request.
func WithJSON(value any) RequestInterceptorOption {
	return WithRequestInterceptorFunc(func(r *http.Request) (*http.Request, error) {
		b, err := json.Marshal(value)
		if err != nil {
			return r, err
		}

		return withBody(bytes.NewReader(b), "application/json", int64(len(b))).InterceptRequest(r)
	})
}

// ResponseInterceptor defines an interface for types that can intercept a
// http.Response and validate certain elements (such as a response status or a
// the presence of a response header), modify it (such as picking a cached
// response) or reject it (such as handling a 404 status code).
type ResponseInterceptor interface {
	// InterceptResponse intercepts r. If the returned error is non-nil, the
	// roundtrip is aborted and the error is returned to the caller. Otherwise
	// the returned response is pushed downstream.
	InterceptResponse(r *http.Response) (*http.Response, error)
}

// ResponseInterceptorFunc is a convenience type implementing
// ResponseInterceptor as a bare function.
type ResponseInterceptorFunc func(r *http.Response) (*http.Response, error)

func (f ResponseInterceptorFunc) InterceptResponse(r *http.Response) (*http.Response, error) {
	return f(r)
}

// ResponseInterceptorOption implements both ClientOption and RequestOption to
// wrap a ResponseInterceptor and use it is part of the request roundtrip.
type ResponseInterceptorOption struct {
	ResponseInterceptor
}

func (ResponseInterceptorOption) clientOpt() {}
func (ResponseInterceptorOption) reqOpt()    {}

// WithResponseInterceptor wraps r in a ResponseInterceptorOption.
func WithResponseInterceptor(r ResponseInterceptor) ResponseInterceptorOption {
	return ResponseInterceptorOption{r}
}

// WithResponseInterceptorFunc wraps f in a ResponseInterceptorOption.
func WithResponseInterceptorFunc(f func(*http.Response) (*http.Response, error)) ResponseInterceptorOption {
	return ResponseInterceptorOption{ResponseInterceptorFunc(f)}
}

// ExpectedStatusCode creates a ResponseInterceptor wrapped in a
// ResponseInterceptorOption that expects the resonse' status code to be any
// of the expectedStatusCodes. If the status code matches, the response is
// returned as is with a nil error. If the status code matches neither of the
// given status codes, an error is returned.
func ExpectedStatusCode(expectedStatusCodes ...int) ResponseInterceptorOption {
	return WithResponseInterceptorFunc(func(r *http.Response) (*http.Response, error) {
		for _, statusCode := range expectedStatusCodes {
			if r.StatusCode == statusCode {
				return r, nil
			}
		}
		return r, fmt.Errorf("unexpected status code: %d", r.StatusCode)
	})
}

// forJSON is both a RequestInterceptor and a ResponseInterceptor that is
// used to handle a JSON response body. During request interception, this type
// adds an Accept request header accepting application/json. In the response
// interception this type expects the content type to be application/json and
// then unmarshals the response body to the given value. If the returned
// content type is not application/json an error is returned. Any error that
// occurs while unmarshaling the response body is also returned.
type forJSON struct {
	value any
}

func (*forJSON) clientOpt() {}
func (*forJSON) reqOpt()    {}

func (*forJSON) InterceptRequest(r *http.Request) (*http.Request, error) {
	r.Header.Add("Accept", "application/json")
	return r, nil
}

func (jr *forJSON) InterceptResponse(r *http.Response) (*http.Response, error) {
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		return r, fmt.Errorf("expected JSON response but got %s", ct)
	}

	d, err := io.ReadAll(r.Body)
	if err != nil {
		return r, err
	}

	return r, json.Unmarshal(d, jr.value)
}

// ForJSON creates a RequestOption that captures the response body JSON data
// and unmarshals the data into value.
// The returned option is both a RequestInterceptor and a ResponseInterceptor.
// During request interception, this type
// adds an Accept request header accepting application/json. In the response
// interception this type expects the content type to be application/json and
// then unmarshals the response body to the given value. If the returned
// content type is not application/json an error is returned. Any error that
// occurs while unmarshaling the response body is also returned.
func ForJSON(value any) RequestOption {
	return &forJSON{value}
}

// WithURLPrefix creates a RequestInterceptorOption that applies a common URL
// prefix to requests not starting with either http:// or https://.
// prefix must be a syntactically valid HTTP(s) URL.
//
// Any error produced by applying the prefix to a request's URL will be
// returned when the request is executed.
func WithURLPrefix(prefix string) RequestInterceptorOption {
	return WithRequestInterceptorFunc(func(r *http.Request) (*http.Request, error) {
		if !r.URL.IsAbs() {
			u, err := url.Parse(prefix + r.URL.String())
			if err != nil {
				return r, err
			}
			r.URL = u
		}

		return r, nil
	})
}
