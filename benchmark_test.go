package httpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/httpclient"
	"github.com/mccutchen/go-httpbin/v2/httpbin"
)

func BenchmarkStdLib(b *testing.B) {
	app := httpbin.New()
	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	b.ResetTimer()

	c := http.Client{}

	for i := 0; i < b.N; i++ {
		res, err := c.Get(testServer.URL + "/get")
		if err != nil {
			b.Fatal(err)
		}

		if res.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status code: %d", res.StatusCode)
		}
	}
}

func BenchmarkHTTPClient(b *testing.B) {
	app := httpbin.New()
	testServer := httptest.NewServer(app.Handler())
	defer testServer.Close()

	b.ResetTimer()

	c := httpclient.New(
		httpclient.WithURLPrefix(testServer.URL),
		httpclient.ExpectedStatusCode(http.StatusOK),
	)

	for i := 0; i < b.N; i++ {

		_, err := c.Get(context.Background(), "/get")
		if err != nil {
			b.Fatal(err)
		}
	}
}
