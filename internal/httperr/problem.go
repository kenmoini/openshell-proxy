// Package httperr turns SDK and validation errors into RFC 9457 problem responses.
package httperr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
)

// Problem is an RFC 9457 problem document with OpenShell extension members.
// It implements huma.StatusError, huma.HeadersError and huma.ContentTypeFilter.
type Problem struct {
	Type              string              `json:"type,omitempty"`
	Title             string              `json:"title"`
	Status            int                 `json:"status"`
	Detail            string              `json:"detail,omitempty"`
	Instance          string              `json:"instance,omitempty"`
	Errors            []*huma.ErrorDetail `json:"errors,omitempty"`
	Reason            string              `json:"reason,omitempty"`
	Domain            string              `json:"domain,omitempty"`
	Metadata          map[string]string   `json:"metadata,omitempty"`
	GRPCCode          string              `json:"grpcCode,omitempty"`
	RetryAfterSeconds *int                `json:"retryAfterSeconds,omitempty"`
	OutcomeUnknown    bool                `json:"outcomeUnknown,omitempty"`
	RequestID         string              `json:"requestId,omitempty"`

	headers http.Header
}

// New builds a Problem whose title is the standard status text.
func New(status int, detail string) *Problem {
	return &Problem{Title: http.StatusText(status), Status: status, Detail: detail}
}

// Error implements error.
func (p *Problem) Error() string {
	if p.Detail == "" {
		return fmt.Sprintf("%d %s", p.Status, p.Title)
	}
	return fmt.Sprintf("%d %s: %s", p.Status, p.Title, p.Detail)
}

// GetStatus implements huma.StatusError.
func (p *Problem) GetStatus() int { return p.Status }

// GetHeaders implements huma.HeadersError.
func (p *Problem) GetHeaders() http.Header { return p.headers }

// ContentType implements huma.ContentTypeFilter.
func (p *Problem) ContentType(string) string { return "application/problem+json" }

func (p *Problem) setHeader(k, v string) {
	if p.headers == nil {
		p.headers = http.Header{}
	}
	p.headers.Set(k, v)
}

// WithHeader sets a response header on p and returns p for chaining.
func (p *Problem) WithHeader(k, v string) *Problem {
	p.setHeader(k, v)
	return p
}

func (p *Problem) setRetryAfter(seconds int) {
	p.RetryAfterSeconds = &seconds
	p.setHeader("Retry-After", strconv.Itoa(seconds))
}

// Write renders p as an application/problem+json response.
func Write(w http.ResponseWriter, _ *http.Request, p *Problem) {
	if p.RequestID == "" {
		p.RequestID = w.Header().Get("X-Request-ID")
	}
	for k, vs := range p.GetHeaders() {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}
