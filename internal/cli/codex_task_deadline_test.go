package cli

import (
	"context"
	"errors"
	"testing"
)

func TestCodexBackgroundDeadlinePreservesSuccessfulTurn(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errCodexTaskSessionDeadline)
	out := ReviewOutput{Summary: "completed before deadline"}

	got, err := codexBackgroundDeadlineResult(ctx, out, nil)
	if err != nil || got.Summary != out.Summary {
		t.Fatalf("successful turn was replaced after deadline: output=%q error=%v", got.Summary, err)
	}
}

func TestCodexBackgroundDeadlineOverridesFailedTurn(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errCodexTaskSessionDeadline)

	got, err := codexBackgroundDeadlineResult(ctx, ReviewOutput{Summary: "stdout closed"}, errors.New("EOF"))
	if err == nil || got.Summary != codexTaskTimeoutMessage() {
		t.Fatalf("failed turn did not report deadline: output=%q error=%v", got.Summary, err)
	}
}
