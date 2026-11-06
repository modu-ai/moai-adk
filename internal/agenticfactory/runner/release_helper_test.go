package runner

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// releaseHeld releases the holder's current claim of the Issue, whatever its epoch.
func releaseHeld(t *testing.T, r *Runner, i core.IssueRef, holder string) error {
	t.Helper()
	rec, err := r.Store.Read(i)
	if err != nil {
		return err
	}
	return r.Store.Release(i, holder, rec.Epoch)
}
