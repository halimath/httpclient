package httpclient_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/halimath/httpclient"
)

func Example_simpleGetRequest() {
	c := httpclient.New(httpclient.WithURLPrefix("https://httpbin.org"))
	ctx := context.Background()
	res, err := c.Get(ctx, "/status/204")
	if err != nil {
		panic(err)
	}

	fmt.Println(res.StatusCode)

	// Output: 204
}

func Example_jsonResponse() {
	c := httpclient.New(
		httpclient.WithURLPrefix("https://httpbin.org"),
		httpclient.ExpectedStatusCode(http.StatusOK),
	)

	var userAgent struct {
		UserAgent string `json:"user-agent"`
	}

	ctx := context.Background()
	_, err := c.Get(ctx, "/user-agent",
		httpclient.WithRequestHeader("User-Agent", "httpclient/1.0"),
		httpclient.ForJSON(&userAgent),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(userAgent.UserAgent)

	// Output: httpclient/1.0
}
