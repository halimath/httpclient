// Package httpclient provides a convenient client API to send HTTP requests
// and handle the reponse. It is based on http.Client and adds an abstraction
// layer that provides convenient capabilities often needed when handling HTTP
// requests.
//
// httpclient supports all options offered by http.Client with the exception of
// a client-global timeout (httpclient uses context.Context for this).
package httpclient
