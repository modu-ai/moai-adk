package cli

// SPEC-CLI-STATE-DIR-BOUND-001 REQ-7 — the chain ledger has one canonical
// location, and events left at the legacy one are never silently orphaned.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// chainProject builds an owned project tree with a state dir and returns it.
func chainProject(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	root := mustMkdirAll(t, filepath.Join(home, "project"))
	mustMkdirAll(t, filepath.Join(root, ".moai", "state"))
	// The home boundary is the test's own, so discovery cannot wander into
	// whatever the machine has above the temp directory.
	t.Setenv("HOME", home)
	return root
}

// TestChainDirIsCanonicalUnderBothBranches pins REQ-7: naming the project
// through the environment and letting it be discovered must land on the same
// directory. They used not to — the environment branch built
// .moai/state/chain and the discovery branch built .moai/chain — so one
// project could accumulate two ledgers depending on how it was invoked.
func TestChainDirIsCanonicalUnderBothBranches(t *testing.T) {
	root := chainProject(t)
	sub := mustMkdirAll(t, filepath.Join(root, "sub"))
	// Normalize the root that exists, then join: EvalSymlinks on a path that
	// has not been created yet would hand back the raw form.
	want := filepath.Join(normPath(t, root), ".moai", "state", "chain")
	legacy := filepath.Join(normPath(t, root), ".moai", "chain")

	var out bytes.Buffer

	t.Setenv(config.EnvClaudeProjectDir, root)
	fromEnv, err := resolveChainDirWith(&out)
	if err != nil {
		t.Fatalf("resolveChainDirWith (env branch): %v", err)
	}

	t.Setenv(config.EnvClaudeProjectDir, "")
	t.Chdir(sub)
	fromWalk, err := resolveChainDirWith(&out)
	if err != nil {
		t.Fatalf("resolveChainDirWith (discovery branch): %v", err)
	}

	if fromEnv != fromWalk {
		t.Errorf("branches disagree: env %q vs discovery %q", fromEnv, fromWalk)
	}
	for _, got := range []string{fromEnv, fromWalk} {
		if got != want {
			t.Errorf("got %q, want the canonical %q", got, want)
		}
		if got == legacy {
			t.Errorf("got the legacy location %q", legacy)
		}
	}
}

// writeLedger creates an event ledger with recognizable content.
func writeLedger(t *testing.T, dir, content string) string {
	t.Helper()
	mustMkdirAll(t, dir)
	path := filepath.Join(dir, ChainEventsFile)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// TestChainLegacyEventsRelocation covers all four dispositions of a legacy
// ledger. The "both present" case is not an exotic corner: the hook writer has
// always written the canonical path while the CLI wrote the legacy one, so any
// machine that used both surfaces already has two files.
func TestChainLegacyEventsRelocation(t *testing.T) {
	const body = "{\"event\":\"spawn\"}\n{\"event\":\"exit\"}\n"

	t.Run("legacy only is relocated once", func(t *testing.T) {
		root := t.TempDir()
		canonical := filepath.Join(root, ".moai", "state", "chain")
		legacy := filepath.Join(root, ".moai", "chain")
		legacyFile := writeLedger(t, legacy, body)

		var out bytes.Buffer
		got := disposeLegacyChainDir(canonical, legacy, &out)

		if got != canonical {
			t.Errorf("got %q, want the canonical %q", got, canonical)
		}
		moved, err := os.ReadFile(filepath.Join(canonical, ChainEventsFile))
		if err != nil {
			t.Fatalf("canonical ledger missing: %v", err)
		}
		if string(moved) != body {
			t.Errorf("relocated content differs:\ngot  %q\nwant %q", moved, body)
		}
		if _, err := os.Stat(legacyFile); !os.IsNotExist(err) {
			t.Errorf("legacy ledger still present; stat err = %v", err)
		}
	})

	t.Run("both present keeps the legacy file and warns", func(t *testing.T) {
		root := t.TempDir()
		canonical := filepath.Join(root, ".moai", "state", "chain")
		legacy := filepath.Join(root, ".moai", "chain")
		legacyFile := writeLedger(t, legacy, body)
		writeLedger(t, canonical, "{\"event\":\"hook\"}\n")

		var out bytes.Buffer
		got := disposeLegacyChainDir(canonical, legacy, &out)

		if got != canonical {
			t.Errorf("got %q, want the canonical %q", got, canonical)
		}
		kept, err := os.ReadFile(legacyFile)
		if err != nil {
			t.Fatalf("legacy ledger was touched: %v", err)
		}
		if string(kept) != body {
			t.Errorf("legacy content changed:\ngot  %q\nwant %q", kept, body)
		}
		if !strings.Contains(out.String(), "chain: legacy events retained at") {
			t.Errorf("no retention warning in %q", out.String())
		}
		if !strings.Contains(out.String(), legacyFile) {
			t.Errorf("warning %q does not name the legacy path %q", out.String(), legacyFile)
		}
	})

	t.Run("canonical only is silent", func(t *testing.T) {
		root := t.TempDir()
		canonical := filepath.Join(root, ".moai", "state", "chain")
		legacy := filepath.Join(root, ".moai", "chain")
		writeLedger(t, canonical, body)

		var out bytes.Buffer
		got := disposeLegacyChainDir(canonical, legacy, &out)

		if got != canonical {
			t.Errorf("got %q, want the canonical %q", got, canonical)
		}
		if out.Len() != 0 {
			t.Errorf("expected no output, got %q", out.String())
		}
	})

	t.Run("failed relocation keeps using the legacy path", func(t *testing.T) {
		root := t.TempDir()
		canonical := filepath.Join(root, ".moai", "state", "chain")
		legacy := filepath.Join(root, ".moai", "chain")
		legacyFile := writeLedger(t, legacy, body)

		// Injected rather than provoked through permissions: the permission
		// model differs on Windows, and skipping the case there would leave
		// the riskiest branch of REQ-7 unverified on a platform that ships.
		original := migrateChainEvents
		migrateChainEvents = func(string, string) error { return os.ErrPermission }
		t.Cleanup(func() { migrateChainEvents = original })

		var out bytes.Buffer
		got := disposeLegacyChainDir(canonical, legacy, &out)

		if got != legacy {
			t.Errorf("got %q, want the legacy path %q to stay in use", got, legacy)
		}
		kept, err := os.ReadFile(legacyFile)
		if err != nil {
			t.Fatalf("legacy ledger lost after a failed relocation: %v", err)
		}
		if string(kept) != body {
			t.Errorf("legacy content changed:\ngot  %q\nwant %q", kept, body)
		}
		if !strings.Contains(out.String(), "chain: migration failed, using legacy path") {
			t.Errorf("no failure warning in %q", out.String())
		}
		if !strings.Contains(out.String(), legacyFile) {
			t.Errorf("warning %q does not name the legacy path %q", out.String(), legacyFile)
		}
	})
}
