package httpclient

import (
	"context"
	"fmt"
	"net/http"
)

// RequestOption defines an interface for types that can be passed to requests
// to customize request execution and processing.
type RequestOption interface {
	reqOpt()
}

// ClientOption defines an interface for types that can be passed when
// constructing a Client to customize its behaviour.
type ClientOption interface {
	clientOpt()
}

// HTTPClientOption is a ClientOption that customizes the http.Client in use.
type HTTPClientOption func(*http.Client)

func (HTTPClientOption) clientOpt() {}

// WithTransport creates a ClientOption using t for the Client to be created.
func WithTransport(t http.RoundTripper) ClientOption {
	return HTTPClientOption(func(c *http.Client) {
		c.Transport = t
	})
}

// Client implements a convenient HTTP client.
type Client struct {
	c               *http.Client
	reqInterceptors []RequestInterceptor
	resInterceptors []ResponseInterceptor
}

// New create a new Client using the given opts to customize the client.
// Calling New() with no options creates a fully usable Client using defaults.
func New(opts ...ClientOption) *Client {
	c := &Client{
		c: new(http.Client),
	}

	for _, opt := range opts {
		switch o := opt.(type) {
		case HTTPClientOption:
			o(c.c)
		case RequestInterceptor:
			c.reqInterceptors = append(c.reqInterceptors, o)
		case ResponseInterceptor:
			c.resInterceptors = append(c.resInterceptors, o)
		default:
			panic(fmt.Sprintf("unexpected option: %v", opt))
		}
	}

	return c
}

// Get executes a HTTP GET request for url using ctx and opts. It returns the
// received and (potentially processed) response as well as any error received.
func (c *Client) Get(ctx context.Context, url string, opts ...RequestOption) (*http.Response, error) {
	return c.Execute(ctx, http.MethodGet, url, opts...)
}

// Post executes a HTTP POST request for url using ctx and opts. In contrast to
// the standard lib's Post function this method sets the request body using a
// RequestInterceptor.
func (c *Client) Post(ctx context.Context, url string, opts ...RequestOption) (*http.Response, error) {
	return c.Execute(ctx, http.MethodPost, url, opts...)
}

func (c *Client) Execute(ctx context.Context, method string, url string, opts ...RequestOption) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, opts...)
}

// Do executes req applying any opts and returns the received response as well
// as any error.
func (c *Client) Do(req *http.Request, opts ...RequestOption) (*http.Response, error) {
	var err error

	for _, i := range c.reqInterceptors {
		req, err = i.InterceptRequest(req)
		if err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		if i, ok := opt.(RequestInterceptor); ok {
			req, err = i.InterceptRequest(req)
			if err != nil {
				return nil, err
			}
		}
	}

	res, err := c.c.Do(req)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	for _, opt := range opts {
		if i, ok := opt.(ResponseInterceptor); ok {
			res, err = i.InterceptResponse(res)
			if err != nil {
				return res, err
			}
		}
	}

	for _, i := range c.resInterceptors {
		res, err = i.InterceptResponse(res)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}

// func (c *Client) decorateURL(u string) string {
// 	if len(c.urlPrefix) == 0 {
// 		return u
// 	}

// 	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
// 		return c.urlPrefix + u
// 	}

// 	return u
// }
