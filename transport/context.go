package transport

import "context"

type responseStatusKey struct{}

// ContextWithResponseStatus returns a derived context that associates statusCode
// with outgoing HTTP requests made through this package's client. When a response
// is received, the client stores resp.StatusCode in *statusCode. If no response
// is received (for example, a network failure before headers), *statusCode is not
// updated and usually remains its initial value (commonly 0).
func ContextWithResponseStatus(parent context.Context, statusCode *int) context.Context {
	return context.WithValue(parent, responseStatusKey{}, statusCode)
}
