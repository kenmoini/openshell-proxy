package convert

import (
	"time"

	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// TimePtr returns nil for the zero time so it is omitted from JSON.
func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// Deletion maps an SDK deletion result; nil is treated as Unspecified.
func Deletion(r *v1.DeletionResult) dto.DeletionResult {
	if r == nil {
		return dto.DeletionResult{Outcome: "Unspecified"}
	}
	out := "Unspecified"
	switch r.Outcome {
	case v1.DeletionCompleted:
		out = "Completed"
	case v1.DeletionAccepted:
		out = "Accepted"
	case v1.DeletionAlreadyAbsent:
		out = "AlreadyAbsent"
	}
	return dto.DeletionResult{Outcome: out, SandboxID: r.SandboxID}
}

// DeletionStatus is 202 when deletion was accepted asynchronously, else 200.
func DeletionStatus(r *v1.DeletionResult) int {
	if r != nil && r.Outcome == v1.DeletionAccepted {
		return 202
	}
	return 200
}
