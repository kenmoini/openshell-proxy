package httperr

import (
	"errors"
	"fmt"
	"testing"
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestFromSDK_Table(t *testing.T) {
	d1500 := 1500 * time.Millisecond
	tests := []struct {
		name               string
		in                 error
		ifMatch            bool
		wantStatus         int
		wantReason         string
		wantRetryAfter     string
		wantOutcomeUnknown bool
		wantAuth           bool
	}{
		{name: "invalid argument", in: &v1.StatusError{Code: v1.ErrorInvalidArgument, GRPCCode: int32(codes.InvalidArgument), FieldViolations: []v1.FieldViolation{{Field: "spec.image", Description: "required"}}}, wantStatus: 400},
		{name: "unauthenticated", in: &v1.StatusError{Code: v1.ErrorUnauthenticated, GRPCCode: int32(codes.Unauthenticated)}, wantStatus: 401, wantAuth: true},
		{name: "permission denied", in: &v1.StatusError{Code: v1.ErrorPermissionDenied, GRPCCode: int32(codes.PermissionDenied)}, wantStatus: 403},
		{name: "not found", in: &v1.StatusError{Code: v1.ErrorNotFound, GRPCCode: int32(codes.NotFound)}, wantStatus: 404},
		{name: "already exists", in: &v1.StatusError{Code: v1.ErrorAlreadyExists, GRPCCode: int32(codes.AlreadyExists)}, wantStatus: 409},
		{name: "aborted", in: &v1.StatusError{Code: v1.ErrorConflict, GRPCCode: int32(codes.Aborted)}, wantStatus: 409},
		{name: "aborted ifMatch", in: &v1.StatusError{Code: v1.ErrorConflict, GRPCCode: int32(codes.Aborted)}, ifMatch: true, wantStatus: 412},
		{name: "failed precondition uncertain", in: &v1.StatusError{Code: v1.ErrorConflict, GRPCCode: int32(codes.FailedPrecondition), ErrorInfo: &v1.ErrorInfo{Reason: "REQUEST_OUTCOME_UNCERTAIN"}}, wantStatus: 409, wantReason: "REQUEST_OUTCOME_UNCERTAIN"},
		{name: "failed precondition other", in: &v1.StatusError{Code: v1.ErrorConflict, GRPCCode: int32(codes.FailedPrecondition)}, wantStatus: 422},
		{name: "failed precondition ifMatch", in: &v1.StatusError{Code: v1.ErrorConflict, GRPCCode: int32(codes.FailedPrecondition)}, ifMatch: true, wantStatus: 412},
		{name: "resource exhausted", in: &v1.StatusError{Code: v1.ErrorUnavailable, GRPCCode: int32(codes.ResourceExhausted), RetryDelay: &d1500}, wantStatus: 429, wantRetryAfter: "2"},
		{name: "unavailable", in: &v1.StatusError{Code: v1.ErrorUnavailable, GRPCCode: int32(codes.Unavailable)}, wantStatus: 503, wantRetryAfter: "1"},
		{name: "deadline exceeded", in: &v1.StatusError{Code: v1.ErrorDeadlineExceeded, GRPCCode: int32(codes.DeadlineExceeded)}, wantStatus: 504, wantOutcomeUnknown: true},
		{name: "cancelled", in: &v1.StatusError{Code: v1.ErrorCancelled, GRPCCode: int32(codes.Canceled)}, wantStatus: 504, wantOutcomeUnknown: true},
		{name: "unimplemented", in: &v1.StatusError{Code: v1.ErrorUnimplemented, GRPCCode: int32(codes.Unimplemented)}, wantStatus: 501},
		{name: "internal", in: &v1.StatusError{Code: v1.ErrorInternal, GRPCCode: int32(codes.Internal)}, wantStatus: 502},
		{name: "plain error", in: errors.New("boom"), wantStatus: 502},
		{name: "wrapped", in: fmt.Errorf("wrap: %w", &v1.StatusError{Code: v1.ErrorNotFound, GRPCCode: int32(codes.NotFound)}), wantStatus: 404},
		{name: "terminal phase", in: &v1.StatusError{Code: v1.ErrorInternal, Message: "sandbox in terminal phase"}, wantStatus: 409, wantReason: "SANDBOX_TERMINAL_PHASE"},
		{name: "code only not found", in: &v1.StatusError{Code: v1.ErrorNotFound}, wantStatus: 404},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := FromSDK(tc.in, Options{IfMatch: tc.ifMatch})
			require.NotNil(t, p)
			assert.Equal(t, tc.wantStatus, p.Status)
			assert.Equal(t, tc.wantStatus, p.GetStatus())
			assert.Equal(t, tc.wantReason, p.Reason)
			assert.Equal(t, tc.wantOutcomeUnknown, p.OutcomeUnknown)
			h := p.GetHeaders()
			assert.Equal(t, tc.wantRetryAfter, h.Get("Retry-After"))
			assert.Equal(t, tc.wantRetryAfter != "", p.RetryAfterSeconds != nil)
			if tc.wantAuth {
				assert.Equal(t, "Bearer", h.Get("WWW-Authenticate"))
			} else {
				assert.Empty(t, h.Get("WWW-Authenticate"))
			}
		})
	}

	t.Run("nil", func(t *testing.T) { assert.Nil(t, FromSDK(nil, Options{})) })

	t.Run("field violations", func(t *testing.T) {
		p := FromSDK(&v1.StatusError{Code: v1.ErrorInvalidArgument, GRPCCode: int32(codes.InvalidArgument), FieldViolations: []v1.FieldViolation{{Field: "spec.image", Description: "required"}}}, Options{})
		require.Len(t, p.Errors, 1)
		assert.Equal(t, "body.spec.image", p.Errors[0].Location)
		assert.Equal(t, "required", p.Errors[0].Message)
		assert.Equal(t, "InvalidArgument", p.GRPCCode)
		assert.Equal(t, "Bad Request", p.Title)
	})
}
