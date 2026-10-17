package dto

import "time"

// PolicyChunk is a proposed policy change awaiting review.
type PolicyChunk struct {
	ID                           string             `json:"id"`
	Status                       string             `json:"status,omitempty"`
	RuleName                     string             `json:"ruleName,omitempty"`
	ProposedRule                 *NetworkPolicyRule `json:"proposedRule,omitempty"`
	Rationale                    string             `json:"rationale,omitempty"`
	SecurityNotes                string             `json:"securityNotes,omitempty"`
	Confidence                   float32            `json:"confidence"`
	DenialSummaryIDs             []string           `json:"denialSummaryIds"`
	CreatedAt                    *time.Time         `json:"createdAt,omitempty"`
	DecidedAt                    *time.Time         `json:"decidedAt,omitempty"`
	Stage                        string             `json:"stage,omitempty"`
	SupersedesChunkID            string             `json:"supersedesChunkId,omitempty"`
	HitCount                     int32              `json:"hitCount"`
	FirstSeen                    *time.Time         `json:"firstSeen,omitempty"`
	LastSeen                     *time.Time         `json:"lastSeen,omitempty"`
	Binary                       string             `json:"binary,omitempty"`
	ValidationResult             string             `json:"validationResult,omitempty"`
	RejectionReason              string             `json:"rejectionReason,omitempty"`
	ApplicationError             string             `json:"applicationError,omitempty"`
	ReviewToken                  string             `json:"reviewToken,omitempty"`
	CurrentEffectivePolicyHash   string             `json:"currentEffectivePolicyHash,omitempty"`
	CandidateEffectivePolicyHash string             `json:"candidateEffectivePolicyHash,omitempty"`
	CurrentEffectivePolicy       *SandboxPolicy     `json:"currentEffectivePolicy,omitempty"`
	CandidateEffectivePolicy     *SandboxPolicy     `json:"candidateEffectivePolicy,omitempty"`
}

// DraftPolicy is the full draft policy state.
type DraftPolicy struct {
	Chunks         []PolicyChunk `json:"chunks"`
	RollingSummary string        `json:"rollingSummary,omitempty"`
	DraftVersion   Uint64String  `json:"draftVersion"`
	LastAnalyzedAt *time.Time    `json:"lastAnalyzedAt,omitempty"`
}

// DraftHistoryEntry is one event in the draft history.
type DraftHistoryEntry struct {
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	EventType   string     `json:"eventType,omitempty"`
	Description string     `json:"description,omitempty"`
	ChunkID     string     `json:"chunkId,omitempty"`
}

// DraftHistory is the draft history list.
type DraftHistory struct {
	Items []DraftHistoryEntry `json:"items"`
}

// SandboxPolicyRevision is a versioned policy revision.
type SandboxPolicyRevision struct {
	Version    uint32            `json:"version"`
	PolicyHash string            `json:"policyHash,omitempty"`
	Status     string            `json:"status" doc:"Unspecified, Pending, Loaded, Failed or Superseded."`
	LoadError  string            `json:"loadError,omitempty"`
	CreatedAt  *time.Time        `json:"createdAt,omitempty"`
	LoadedAt   *time.Time        `json:"loadedAt,omitempty"`
	Policy     *SandboxPolicy    `json:"policy,omitempty"`
	Provenance map[string]string `json:"provenance,omitempty"`
}

// PolicyStatusResult is a revision plus the currently active version.
type PolicyStatusResult struct {
	Revision      SandboxPolicyRevision `json:"revision"`
	ActiveVersion uint32                `json:"activeVersion"`
}

// ApproveRequest approves one chunk, bound to its review token.
type ApproveRequest struct {
	ReviewToken string `json:"reviewToken" minLength:"1"`
}

// RejectRequest rejects one chunk.
type RejectRequest struct {
	Reason string `json:"reason,omitempty"`
}

// DraftChunkApproval binds a chunk to the review token that was inspected.
type DraftChunkApproval struct {
	ChunkID     string `json:"chunkId"`
	ReviewToken string `json:"reviewToken"`
}

// ApproveAllRequest approves many chunks.
type ApproveAllRequest struct {
	IncludeSecurityFlagged bool                 `json:"includeSecurityFlagged,omitempty"`
	Approvals              []DraftChunkApproval `json:"approvals,omitempty"`
}

// ApproveResult is the result of approving a chunk.
type ApproveResult struct {
	PolicyVersion uint32 `json:"policyVersion"`
	PolicyHash    string `json:"policyHash,omitempty"`
}

// ApproveAllResult is the result of a bulk approval.
type ApproveAllResult struct {
	PolicyVersion  uint32 `json:"policyVersion"`
	PolicyHash     string `json:"policyHash,omitempty"`
	ChunksApproved uint32 `json:"chunksApproved"`
	ChunksSkipped  uint32 `json:"chunksSkipped"`
}

// UndoResult is the result of undoing a chunk approval.
type UndoResult struct {
	PolicyVersion uint32 `json:"policyVersion"`
	PolicyHash    string `json:"policyHash,omitempty"`
}

// ClearResult is the result of clearing draft chunks.
type ClearResult struct {
	ChunksCleared uint32 `json:"chunksCleared"`
}
