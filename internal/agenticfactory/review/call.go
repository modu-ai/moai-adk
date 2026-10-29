package review

import (
	"context"
	"errors"
	"fmt"
)

// BlockReason is the recorded reason a provider call stopped (REQ-AF-014).
type BlockReason string

const (
	BlockNone             BlockReason = ""
	BlockAuthFailed       BlockReason = "auth-failed"
	BlockPermissionDenied BlockReason = "permission-denied"
	BlockRateLimited      BlockReason = "rate-limited"
	BlockNotConfigured    BlockReason = "not-configured"
)

// StatusError is a provider failure that carries an HTTP-style status code.
type StatusError struct{ Code int }

func (e *StatusError) Error() string { return fmt.Sprintf("review: provider status %d", e.Code) }

// classify maps a provider error to a block reason, or BlockNone when the
// error is not one of the recorded kinds.
func classify(err error) BlockReason {
	var se *StatusError
	switch {
	case errors.Is(err, ErrNotConfigured), errors.Is(err, ErrBinaryAbsent):
		return BlockNotConfigured
	case errors.As(err, &se):
		switch se.Code {
		case 401:
			return BlockAuthFailed
		case 403:
			return BlockPermissionDenied
		case 429:
			return BlockRateLimited
		}
	}
	return BlockNone
}

// CallOnce calls the provider exactly once. It never retries and never
// acquires another key or capacity: an authentication failure, a permission
// denial, or a 429 comes back as a block reason with the original error, for
// the caller to record against the operation that failed.
func CallOnce(ctx context.Context, p Provider, req Request) (Result, BlockReason, error) {
	res, err := p.Review(ctx, req)
	if err != nil {
		return Result{}, classify(err), err
	}
	return res, BlockNone, nil
}
