// Package auth extracts the caller's bearer token and carries it in the request context.
package auth

import "context"

type tokenKey struct{}

// WithToken returns ctx carrying the caller's bearer token.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}

// TokenFrom returns the bearer token stored by WithToken.
func TokenFrom(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(tokenKey{}).(string)
	return t, ok && t != ""
}
