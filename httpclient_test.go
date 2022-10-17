package httpclient_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/halimath/expect-go"
	"github.com/halimath/httpclient"
	"github.com/mccutchen/go-httpbin/v2/httpbin"
)

var errNotFound = errors.New("not found")

func TestPackage(t *testing.T) {
	app := httpbin.New()
	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	client := httpclient.New(
		httpclient.WithURLPrefix(testServer.URL),
		httpclient.WithResponseInterceptor(httpclient.ResponseInterceptorFunc(func(r *http.Response) (*http.Response, error) {
			if r.StatusCode == http.StatusNotFound {
				return nil, errNotFound
			}
			return r, nil
		})),
	)

	t.Run("GetJSON_notFound", func(t *testing.T) {
		ctx := context.Background()
		_, err := client.Get(ctx, "/status/404")
		ExpectThat(t, err).Is(Error(errNotFound))
	})

	t.Run("GetJSON_success", func(t *testing.T) {
		var body httpbinMethodResponse

		ctx := context.Background()
		_, err := client.Get(ctx, "/get",
			httpclient.ExpectedStatusCode(http.StatusOK),
			httpclient.ForJSON(&body),
		)

		t.Log(err)

		ExpectThat(t, err).Is(NoError())
		ExpectThat(t, body).Is(DeepEqual(httpbinMethodResponse{
			Args: make(map[string]any),
			Header: map[string][]string{
				"Accept":          {"application/json"},
				"Accept-Encoding": {"gzip"},
				"Host":            {strings.Replace(testServer.URL, "http://", "", -1)},
				"User-Agent":      {"Go-http-client/1.1"},
			},
			URL: testServer.URL + "/get",
		}))
	})

	t.Run("PostJSON_success", func(t *testing.T) {
		var body httpbinMethodResponse
		ctx := context.Background()
		_, err := client.Post(ctx, "/post",
			httpclient.WithJSON("hello, world"),
			httpclient.ForJSON(&body),
		)
		ExpectThat(t, err).Is(NoError())
		ExpectThat(t, body).Is(DeepEqual(httpbinMethodResponse{
			Args: make(map[string]any),
			Data: `"hello, world"`,
			Header: map[string][]string{
				"Accept":          {"application/json"},
				"Accept-Encoding": {"gzip"},
				"Host":            {strings.Replace(testServer.URL, "http://", "", -1)},
				"User-Agent":      {"Go-http-client/1.1"},
				"Content-Type":    {"application/json"},
				"Content-Length":  {"14"},
			},
			URL: testServer.URL + "/post",
		}))

	})
}

type httpbinMethodResponse struct {
	Args   map[string]any      `json:"args"`
	Data   string              `json:"data"`
	File   map[string]any      `json:"files"`
	Form   map[string]any      `json:"form"`
	Header map[string][]string `json:"headers"`
	// "json": null,
	URL string `json:"url"`
}
