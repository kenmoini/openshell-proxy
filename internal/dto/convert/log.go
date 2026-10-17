package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// LogResultFromSDK maps an SDK log result; Lines is never nil.
func LogResultFromSDK(r *v1.LogResult) dto.LogResult {
	out := dto.LogResult{Lines: []dto.LogLine{}}
	if r == nil {
		return out
	}
	out.BufferTotal = r.BufferTotal
	for _, l := range r.Lines {
		out.Lines = append(out.Lines, dto.LogLine{
			Timestamp: l.Timestamp,
			Level:     l.Level,
			Target:    l.Target,
			Message:   l.Message,
			Source:    l.Source,
			Fields:    l.Fields,
		})
	}
	return out
}
