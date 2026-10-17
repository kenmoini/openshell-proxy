package dto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Uint64String is a uint64 that travels as a quoted decimal string in JSON so
// that values above 2^53 survive JavaScript clients. It accepts either a JSON
// integer or a quoted decimal string on input.
type Uint64String uint64

// MarshalJSON renders the value as a quoted decimal string.
func (u Uint64String) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatUint(uint64(u), 10) + `"`), nil
}

// UnmarshalJSON accepts 123 or "123"; negatives, floats and non-numeric
// strings are errors.
func (u *Uint64String) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if string(b) == "null" {
		return nil
	}
	s := string(b)
	if len(b) >= 2 && b[0] == '"' {
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid unsigned integer %q", s)
	}
	*u = Uint64String(n)
	return nil
}

// ObjectMeta is the common metadata block of API resources.
type ObjectMeta struct {
	ID                string            `json:"id,omitempty"`
	Name              string            `json:"name"`
	Workspace         string            `json:"workspace,omitempty"`
	CreatedAt         *time.Time        `json:"createdAt,omitempty"`
	DeletionTimestamp *time.Time        `json:"deletionTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	ResourceVersion   Uint64String      `json:"resourceVersion,omitempty"`
}

// Page is one page of a list response. Items must be a non-nil slice so an
// empty page marshals as [] rather than null; callers and converters ensure
// this.
type Page[T any] struct {
	Items         []T    `json:"items"`
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// DeletionResult reports the outcome of a delete call.
type DeletionResult struct {
	Outcome   string `json:"outcome"` // "Completed" | "Accepted" | "AlreadyAbsent" | "Unspecified"
	SandboxID string `json:"sandboxId,omitempty"`
}
