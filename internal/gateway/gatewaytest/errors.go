// Package gatewaytest provides a fake-backed v1.ClientInterface whose
// sub-interfaces can be replaced by function-field stubs.
package gatewaytest

import (
	"context"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"google.golang.org/grpc/codes"
)

// ErrUnimplemented is returned by every stub function that is not set.
var ErrUnimplemented = &v1.StatusError{
	Code:     v1.ErrorUnimplemented,
	GRPCCode: int32(codes.Unimplemented),
	Message:  "gatewaytest: stub function not set",
}

// StatusErr builds a v1.StatusError, folding the gRPC code the way the SDK does.
func StatusErr(code codes.Code, msg string) *v1.StatusError {
	var ec v1.ErrorCode
	switch code {
	case codes.NotFound:
		ec = v1.ErrorNotFound
	case codes.AlreadyExists:
		ec = v1.ErrorAlreadyExists
	case codes.Unavailable, codes.ResourceExhausted:
		ec = v1.ErrorUnavailable
	case codes.PermissionDenied:
		ec = v1.ErrorPermissionDenied
	case codes.InvalidArgument, codes.OutOfRange:
		ec = v1.ErrorInvalidArgument
	case codes.DeadlineExceeded:
		ec = v1.ErrorDeadlineExceeded
	case codes.Canceled:
		ec = v1.ErrorCancelled
	case codes.Unimplemented:
		ec = v1.ErrorUnimplemented
	case codes.Aborted, codes.FailedPrecondition:
		ec = v1.ErrorConflict
	case codes.Unauthenticated:
		ec = v1.ErrorUnauthenticated
	default:
		ec = v1.ErrorInternal
	}
	return &v1.StatusError{Code: ec, GRPCCode: int32(code), Message: msg}
}

// StaticPager returns a pager yielding one page of items with the given next
// token. If next is non-empty the pager follows it like a real one, and the
// follow-up page is empty and terminal.
func StaticPager[T any](items []T, next string) *v1.Pager[T] {
	return v1.NewPager("", func(_ context.Context, token string) (*v1.Page[T], error) {
		if token != "" {
			return &v1.Page[T]{}, nil
		}
		return &v1.Page[T]{Items: items, NextPageToken: next}, nil
	})
}
