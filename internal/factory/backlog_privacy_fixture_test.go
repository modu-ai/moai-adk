package factory

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
)

// Explicitly own a managed child without modifying the unrelated TempDir.
// A Windows token's default owner can differ from its current user SID.
func privateBacklogFixtureDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "private-queue")
	if err := privatefile.Mkdir(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}
