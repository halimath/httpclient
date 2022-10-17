# httpclient

![CI Status][ci-img-url] [![Go Report Card][go-report-card-img-url]][go-report-card-url]
[![Package Doc][package-doc-img-url]][package-doc-url] [![Releases][release-img-url]][release-url]

[ci-img-url]: https://github.com/halimath/httpclient/workflows/CI/badge.svg
[go-report-card-img-url]: https://goreportcard.com/badge/github.com/halimath/httpclient
[go-report-card-url]: https://goreportcard.com/report/github.com/halimath/httpclient
[package-doc-img-url]: https://img.shields.io/badge/GoDoc-Reference-blue.svg
[package-doc-url]: https://pkg.go.dev/github.com/halimath/httpclient
[release-img-url]: https://img.shields.io/github/v/release/halimath/httpclient.svg
[release-url]: https://github.com/halimath/httpclient/releases

`httpclient` provides a convenient http client library for the go programming language which is completely
based and compatible with the standard lib's `net/http` package but adds commonly used functionality in a
simple API.

# Usage

## Installation

`httpclient` uses go modules and requires Go 1.18 or greater.

```
$ go get -u github.com/halimath/httpclient
```
## Making requests

Making simple requests requires a `httpclient.Client`. This package provides no package level functions.

```go
c := httpclient.New(httpclient.WithURLPrefix("https://httpbin.org"))
ctx := context.Background()
res, err := c.Get(ctx, "/status/204")
```

A central piece of `httpclient` is the use of _interceptors_ to handle requests and 
responses. This allows you to add additional information to a request or "spy" on 
response values. `httpclient` provides a couple of common interceptors. Adding request
headers and handline response bodies becomes very easy:

```go
c := httpclient.New(
	httpclient.WithURLPrefix("https://httpbin.org"),
	httpclient.ExpectedStatusCode(http.StatusOK),
)

var userAgent struct {
	UserAgent string `json:"user-agent"`
}

ctx := context.Background()
res, err := c.Get(ctx, "/user-agent",
	httpclient.WithRequestHeader("User-Agent", "httpclient/1.0"),
	httpclient.ForJSON(&userAgent),
)
```

As you can add interceptors both on the request as well as on the client level, common
things (such as handling error status codes) can easily be defined globally.

You can also provide your own interceptors by implementing either 
`httpclient.RequestInterceptor` or `httpclient.ResponseInterceptor`.

# Changelog

## 0.1.0
* Initial release

# License

```
Copyright 2022 Alexander Metzner.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
