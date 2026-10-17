package convert

import (
	v1 "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	"github.com/kenmoini/openshell-proxy/internal/dto"
)

// PolicyChunkFromSDK maps an SDK policy chunk.
func PolicyChunkFromSDK(c v1.PolicyChunk) dto.PolicyChunk {
	ids := c.DenialSummaryIDs
	if ids == nil {
		ids = []string{}
	}
	return dto.PolicyChunk{
		ID:                           c.ID,
		Status:                       c.Status,
		RuleName:                     c.RuleName,
		ProposedRule:                 NetworkRuleFromSDK(c.ProposedRule),
		Rationale:                    c.Rationale,
		SecurityNotes:                c.SecurityNotes,
		Confidence:                   c.Confidence,
		DenialSummaryIDs:             ids,
		CreatedAt:                    TimePtr(c.CreatedAt),
		DecidedAt:                    TimePtr(c.DecidedAt),
		Stage:                        c.Stage,
		SupersedesChunkID:            c.SupersedesChunkID,
		HitCount:                     c.HitCount,
		FirstSeen:                    TimePtr(c.FirstSeen),
		LastSeen:                     TimePtr(c.LastSeen),
		Binary:                       c.Binary,
		ValidationResult:             c.ValidationResult,
		RejectionReason:              c.RejectionReason,
		ApplicationError:             c.ApplicationError,
		ReviewToken:                  c.ReviewToken,
		CurrentEffectivePolicyHash:   c.CurrentEffectivePolicyHash,
		CandidateEffectivePolicyHash: c.CandidateEffectivePolicyHash,
		CurrentEffectivePolicy:       PolicyFromSDK(c.CurrentEffectivePolicy),
		CandidateEffectivePolicy:     PolicyFromSDK(c.CandidateEffectivePolicy),
	}
}

// DraftPolicyFromSDK maps an SDK draft policy; Chunks is never nil.
func DraftPolicyFromSDK(p *v1.DraftPolicy) dto.DraftPolicy {
	chunks := make([]dto.PolicyChunk, 0, len(p.Chunks))
	for _, c := range p.Chunks {
		chunks = append(chunks, PolicyChunkFromSDK(c))
	}
	return dto.DraftPolicy{
		Chunks:         chunks,
		RollingSummary: p.RollingSummary,
		DraftVersion:   dto.Uint64String(p.DraftVersion),
		LastAnalyzedAt: TimePtr(p.LastAnalyzedAt),
	}
}

// DraftHistoryFromSDK maps history entries; Items is never nil.
func DraftHistoryFromSDK(es []v1.DraftHistoryEntry) dto.DraftHistory {
	items := make([]dto.DraftHistoryEntry, 0, len(es))
	for _, e := range es {
		items = append(items, dto.DraftHistoryEntry{
			Timestamp: TimePtr(e.Timestamp), EventType: e.EventType,
			Description: e.Description, ChunkID: e.ChunkID,
		})
	}
	return dto.DraftHistory{Items: items}
}

// PolicyRevisionFromSDK maps an SDK policy revision.
func PolicyRevisionFromSDK(r v1.SandboxPolicyRevision) dto.SandboxPolicyRevision {
	return dto.SandboxPolicyRevision{
		Version:    r.Version,
		PolicyHash: r.PolicyHash,
		Status:     r.Status.String(),
		LoadError:  r.LoadError,
		CreatedAt:  TimePtr(r.CreatedAt),
		LoadedAt:   TimePtr(r.LoadedAt),
		Policy:     PolicyFromSDK(r.Policy),
		Provenance: r.Provenance,
	}
}

// PolicyStatusFromSDK maps an SDK policy status result.
func PolicyStatusFromSDK(r *v1.PolicyStatusResult) dto.PolicyStatusResult {
	return dto.PolicyStatusResult{Revision: PolicyRevisionFromSDK(r.Revision), ActiveVersion: r.ActiveVersion}
}

// ApproveResultFromSDK maps an approve result.
func ApproveResultFromSDK(r *v1.ApproveResult) dto.ApproveResult {
	return dto.ApproveResult{PolicyVersion: r.PolicyVersion, PolicyHash: r.PolicyHash}
}

// ApproveAllResultFromSDK maps a bulk approve result.
func ApproveAllResultFromSDK(r *v1.ApproveAllResult) dto.ApproveAllResult {
	return dto.ApproveAllResult{PolicyVersion: r.PolicyVersion, PolicyHash: r.PolicyHash,
		ChunksApproved: r.ChunksApproved, ChunksSkipped: r.ChunksSkipped}
}

// UndoResultFromSDK maps an undo result.
func UndoResultFromSDK(r *v1.UndoResult) dto.UndoResult {
	return dto.UndoResult{PolicyVersion: r.PolicyVersion, PolicyHash: r.PolicyHash}
}

// ClearResultFromSDK maps a clear result.
func ClearResultFromSDK(r *v1.ClearResult) dto.ClearResult {
	return dto.ClearResult{ChunksCleared: r.ChunksCleared}
}

// DraftApprovalsToSDK maps approvals to SDK values.
func DraftApprovalsToSDK(in []dto.DraftChunkApproval) []v1.DraftChunkApproval {
	out := make([]v1.DraftChunkApproval, 0, len(in))
	for _, a := range in {
		out = append(out, v1.DraftChunkApproval{ChunkID: a.ChunkID, ReviewToken: a.ReviewToken})
	}
	return out
}
