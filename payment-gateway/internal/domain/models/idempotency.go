package models

import "time"

// IdempotencyClaimRequest describes a request slot owned by one user and scope.
// Request is serialized only to derive a deterministic request hash.
type IdempotencyClaimRequest struct {
	Scope     string
	UserID    uint64
	Key       string
	Request   any
	ExpiresIn time.Duration
}

type IdempotencyClaimResult struct {
	Record       *IdempotencyKeyDTO
	Created      bool
	ResponseBody []byte
}

type IdempotencyKeyDTO struct {
	ID uint64
}
