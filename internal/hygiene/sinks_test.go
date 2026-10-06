package hygiene

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestSinkRegistryContent — the closed registry carries the eight measured
// audit sinks, the hygiene audit sink itself (REQ-HYG-003, REQ-HYG-004),
// and the eleven run-phase additions the completeness guard verified in the
// real tree.
func TestSinkRegistryContent(t *testing.T) {
	want := []string{
		"rule-load-audit.jsonl",
		"agent-model-audit.jsonl",
		"preedit-session-guard.log",
		"codex-adapter.jsonl",
		"status-transition-audit.log",
		"subagent-write-guard.log",
		"agent-stop-audit.jsonl",
		"hook-runtime.log",
		"hygiene-audit.jsonl",
		"anchor-relocation-audit.jsonl",
		"anchor-trace.jsonl",
		"askuser-observations.jsonl",
		"autonomy-downgrade.log",
		"hook-skip.log",
		"lifecycle-close.log",
		"migrations.log",
		"navigator-sync.log",
		"permission.log",
		"slot-lease-audit.jsonl",
		"task-metrics.jsonl",
	}
	got := SinkRegistry()
	if len(got) != len(want) {
		t.Fatalf("registry size = %d, want %d", len(got), len(want))
	}
	seen := map[string]bool{}
	for _, name := range got {
		seen[name] = true
	}
	for _, name := range want {
		if !seen[name] {
			t.Fatalf("registry missing sink %s", name)
		}
		if !IsRegisteredSink(name) {
			t.Fatalf("IsRegisteredSink(%s) = false", name)
		}
	}
	if IsRegisteredSink("rogue-sink.jsonl") {
		t.Fatalf("unregistered name accepted")
	}
}

// TestSinkRegistryCompleteness — AC-HYG-005 (L-005). The guard fails naming
// an unregistered writer in a fixture source tree, and passes over the real
// internal/ tree where every current writer resolves to a registry entry.
func TestSinkRegistryCompleteness(t *testing.T) {
	t.Run("fixture tree with an unregistered writer", func(t *testing.T) {
		src := t.TempDir()
		fixture := "package fixture\n\n" +
			"import (\n\"os\"\n\"path/filepath\"\n)\n\n" +
			"var roguePath = filepath.Join(\".moai\", \"logs\", \"rogue-sink.jsonl\")\n\n" +
			"func writeRogue() error { return os.WriteFile(roguePath, nil, 0o644) }\n"
		if err := os.WriteFile(filepath.Join(src, "writer.go"), []byte(fixture), 0o644); err != nil {
			t.Fatalf("seed fixture: %v", err)
		}
		unregistered, err := ScanSourceSinks(src)
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		if len(unregistered) != 1 || unregistered[0] != "rogue-sink.jsonl" {
			t.Fatalf("unregistered = %v, want [rogue-sink.jsonl]", unregistered)
		}
	})

	t.Run("fixture tree with only registered writers", func(t *testing.T) {
		src := t.TempDir()
		fixture := "package fixture\n\n" +
			"import (\n\"os\"\n\"path/filepath\"\n)\n\n" +
			"var okPath = filepath.Join(\".moai\", \"logs\", \"rule-load-audit.jsonl\")\n\n" +
			"func writeOk() error { return os.WriteFile(okPath, nil, 0o644) }\n"
		if err := os.WriteFile(filepath.Join(src, "writer.go"), []byte(fixture), 0o644); err != nil {
			t.Fatalf("seed fixture: %v", err)
		}
		unregistered, err := ScanSourceSinks(src)
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		if len(unregistered) != 0 {
			t.Fatalf("unregistered = %v, want empty", unregistered)
		}
	})

	t.Run("fixture tree with a reader-only reference", func(t *testing.T) {
		src := t.TempDir()
		fixture := "package fixture\n\n" +
			"import (\n\"os\"\n\"path/filepath\"\n)\n\n" +
			"var roguePath = filepath.Join(\".moai\", \"logs\", \"rogue-sink.jsonl\")\n\n" +
			"func readRogue() ([]byte, error) { return os.ReadFile(roguePath) }\n"
		if err := os.WriteFile(filepath.Join(src, "reader.go"), []byte(fixture), 0o644); err != nil {
			t.Fatalf("seed fixture: %v", err)
		}
		unregistered, err := ScanSourceSinks(src)
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		if len(unregistered) != 0 {
			t.Fatalf("reader-only reference flagged as writer: %v", unregistered)
		}
	})

	t.Run("real internal tree resolves every writer", func(t *testing.T) {
		_, thisFile, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatalf("runtime.Caller failed")
		}
		pkgDir := filepath.Dir(thisFile)
		repoInternal := filepath.Join(pkgDir, "..")
		unregistered, err := ScanSourceSinks(repoInternal)
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		if len(unregistered) != 0 {
			t.Fatalf("unregistered sink writers in the real tree: %s",
				strings.Join(unregistered, ", "))
		}
	})
}

// TestRootGuardRefusesNonTempRoots — AC-HYG-016 (runtime arm, M1 share).
// With testing.Testing() true, an entry point handed a root outside the
// test's registered temp directory refuses to run.
func TestRootGuardRefusesNonTempRoots(t *testing.T) {
	root := t.TempDir()
	registerTestRoot(root)
	if err := validateRoot(root); err != nil {
		t.Fatalf("registered temp root refused: %v", err)
	}
	outside := filepath.Join(string(os.PathSeparator), "not", "a", "test", "root")
	if err := validateRoot(outside); err == nil {
		t.Fatalf("root outside the registered temp dirs was accepted")
	}
}
