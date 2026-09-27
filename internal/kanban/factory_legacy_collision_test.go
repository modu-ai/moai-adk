package kanban

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// seedFactoryRegistry writes live claims (the probe below treats every pid
// as alive) into root's registry.
func seedFactoryRegistry(t *testing.T, root string, labels ...string) {
	t.Helper()
	reg := map[string]FactoryWorkerEntry{}
	for i, label := range labels {
		reg[label] = FactoryWorkerEntry{PID: 1000 + i}
	}
	if err := SaveFactoryRegistry(FactoryRegistryPath(root), reg); err != nil {
		t.Fatalf("seed registry: %v", err)
	}
}

// TestClaimFactoryWorkerExplicitLegacyCollisionNamesTheRow: an operator-typed
// number whose number is held by a live legacy row is refused with an error
// naming that legacy label — never silently moved to another number.
func TestClaimFactoryWorkerExplicitLegacyCollisionNamesTheRow(t *testing.T) {
	t.Parallel()
	alive := func(int) bool { return true }

	cases := []struct{ held, request string }{
		{"agent-3", "worker-3"}, // canonical request vs legacy agent row
		{"lane-3", "lane-3"},    // legacy alias request vs legacy lane row
		{"lane-2", "agent-2"},   // cross-shape legacy collision
	}
	for _, c := range cases {
		root := t.TempDir()
		seedFactoryRegistry(t, root, c.held)
		_, err := ClaimFactoryWorker(root, c.request, false, 4242, alive)
		var collision *FactoryLegacyCollisionError
		if !errors.As(err, &collision) {
			t.Errorf("explicit %s over live %s: err = %v, want *FactoryLegacyCollisionError", c.request, c.held, err)
			continue
		}
		if collision.Held != c.held || !strings.Contains(err.Error(), c.held) {
			t.Errorf("explicit %s over live %s: error %q does not name the legacy row", c.request, c.held, err)
		}
		if _, claimed := LoadFactoryRegistry(FactoryRegistryPath(root))[collision.Requested]; claimed {
			t.Errorf("a refused claim must record nothing, registry holds %s", collision.Requested)
		}
	}
}

// TestClaimFactoryWorkerReportsSkippedLegacyRows: a number the launcher
// chose (auto) lands past live legacy rows and names them; an explicit
// request bumped past a canonical holder names the legacy rows it hops.
func TestClaimFactoryWorkerReportsSkippedLegacyRows(t *testing.T) {
	t.Parallel()
	alive := func(int) bool { return true }

	t.Run("auto", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		seedFactoryRegistry(t, root, "worker-1", "lane-2", "agent-4")
		// `-f worker` asks for one past the highest live claim of any shape.
		got, err := ClaimFactoryWorker(root, "worker-5", true, 4242, alive)
		if err != nil || got.Label != "worker-5" {
			t.Fatalf("auto claim = (%+v, %v), want worker-5", got, err)
		}
		if !slices.Equal(got.SkippedLegacy, []string{"lane-2", "agent-4"}) {
			t.Errorf("auto claim skipped %v, want [lane-2 agent-4]", got.SkippedLegacy)
		}
	})

	t.Run("explicit bump", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		seedFactoryRegistry(t, root, "worker-3", "lane-4")
		got, err := ClaimFactoryWorker(root, "worker-3", false, 4242, alive)
		if err != nil || got.Label != "worker-5" {
			t.Fatalf("explicit claim = (%+v, %v), want worker-5", got, err)
		}
		if !slices.Equal(got.SkippedLegacy, []string{"lane-4"}) {
			t.Errorf("explicit bump skipped %v, want [lane-4]", got.SkippedLegacy)
		}
	})

	t.Run("no legacy rows, no report", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		seedFactoryRegistry(t, root, "worker-1")
		got, err := ClaimFactoryWorker(root, "worker-2", true, 4242, alive)
		if err != nil || got.Label != "worker-2" || len(got.SkippedLegacy) != 0 {
			t.Errorf("clean auto claim = (%+v, %v), want worker-2 with nothing skipped", got, err)
		}
	})
}
