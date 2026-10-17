package httperr

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"google.golang.org/grpc/codes"
)

// Options tunes FromSDK for the calling route.
type Options struct {
	// IfMatch is true when the request carried an If-Match precondition, so
	// Aborted and FailedPrecondition mean 412 rather than 409/422.
	IfMatch bool
}

// FromSDK maps an SDK (or other) error to a Problem. A nil error yields nil.
func FromSDK(err error, opts Options) *Problem {
	if err == nil {
		return nil
	}
	var se *v1.StatusError
	if !errors.As(err, &se) {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			p := New(http.StatusGatewayTimeout, "upstream call timed out or was cancelled; outcome unknown")
			p.OutcomeUnknown = true
			return p
		}
		return New(http.StatusBadGateway, "upstream error")
	}

	status := statusFor(se, opts)
	p := New(status, se.Message)
	if se.GRPCCode != 0 {
		p.GRPCCode = codes.Code(se.GRPCCode).String()
	}
	if se.ErrorInfo != nil {
		p.Reason = se.ErrorInfo.Reason
		p.Domain = se.ErrorInfo.Domain
		p.Metadata = se.ErrorInfo.Metadata
	}
	if isTerminalPhase(se) {
		p.Reason = "SANDBOX_TERMINAL_PHASE"
	}
	for _, fv := range se.FieldViolations {
		p.Errors = append(p.Errors, &huma.ErrorDetail{Location: "body." + fv.Field, Message: fv.Description})
	}

	switch {
	case se.RetryDelay != nil:
		p.setRetryAfter(ceilSeconds(*se.RetryDelay))
	case status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable:
		p.setRetryAfter(1)
	}
	if status == http.StatusUnauthorized {
		p.setHeader("WWW-Authenticate", "Bearer")
	}
	if status == http.StatusGatewayTimeout {
		p.OutcomeUnknown = true
	}
	return p
}

func isTerminalPhase(se *v1.StatusError) bool {
	return se.Code == v1.ErrorInternal && se.GRPCCode == 0 && se.Cause == nil
}

func statusFor(se *v1.StatusError, opts Options) int {
	if isTerminalPhase(se) {
		return http.StatusConflict
	}
	if se.GRPCCode != 0 {
		return statusFromGRPC(codes.Code(se.GRPCCode), se, opts)
	}
	switch se.Code {
	case v1.ErrorInvalidArgument:
		return http.StatusBadRequest
	case v1.ErrorUnauthenticated:
		return http.StatusUnauthorized
	case v1.ErrorPermissionDenied:
		return http.StatusForbidden
	case v1.ErrorNotFound:
		return http.StatusNotFound
	case v1.ErrorAlreadyExists:
		return http.StatusConflict
	case v1.ErrorConflict:
		if opts.IfMatch {
			return http.StatusPreconditionFailed
		}
		return http.StatusConflict
	case v1.ErrorUnavailable:
		return http.StatusServiceUnavailable
	case v1.ErrorDeadlineExceeded, v1.ErrorCancelled:
		return http.StatusGatewayTimeout
	case v1.ErrorUnimplemented:
		return http.StatusNotImplemented
	}
	return http.StatusBadGateway
}

func statusFromGRPC(c codes.Code, se *v1.StatusError, opts Options) int {
	switch c {
	case codes.InvalidArgument, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Aborted:
		if opts.IfMatch {
			return http.StatusPreconditionFailed
		}
		return http.StatusConflict
	case codes.FailedPrecondition:
		if opts.IfMatch {
			return http.StatusPreconditionFailed
		}
		if se.ErrorInfo != nil && strings.HasPrefix(se.ErrorInfo.Reason, "REQUEST_") {
			return http.StatusConflict
		}
		return http.StatusUnprocessableEntity
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded, codes.Canceled:
		return http.StatusGatewayTimeout
	case codes.Unimplemented:
		return http.StatusNotImplemented
	}
	return http.StatusBadGateway
}

func ceilSeconds(d time.Duration) int {
	s := int(math.Ceil(d.Seconds()))
	if s < 1 {
		return 1
	}
	return s
}
