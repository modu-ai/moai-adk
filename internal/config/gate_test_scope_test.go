package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// gate_test_scope_test.go — gate.pre_commit.test_scope.mode schema tests
// (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 M1, REQ-PAS-011 / AC-PAS-011).
//
// The pre-commit context's Go test step runs the affected packages by
// default. The key accepts "affected" or "full" and rejects anything else
// with a message naming the key and the accepted values.

// loadGateScopeMode loads a gate.yaml body through the shared loader and
// returns the config together with the validation verdict for it.
func loadGateScopeMode(t *testing.T, body string) (*Config, error) {
	t.Helper()
	moaiDir := writeGateYAML(t, body)
	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg, Validate(cfg, loader.LoadedSections())
}

// preCommitScopeBlock returns the text between the pre_commit.enabled line and
// the ast_grep_gate section of a gate.yaml — the region the scope block lives
// in. The enabled line itself is excluded: the two shipped copies differ there
// on purpose (this repository opts in, the template does not).
func preCommitScopeBlock(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	s := string(data)
	const head = "  pre_commit:\n    enabled: "
	start := strings.Index(s, head)
	if start < 0 {
		t.Fatalf("%s: no pre_commit block", path)
	}
	afterEnabled := strings.Index(s[start+len(head):], "\n")
	if afterEnabled < 0 {
		t.Fatalf("%s: pre_commit.enabled line is unterminated", path)
	}
	bodyStart := start + len(head) + afterEnabled + 1
	// bodyStart sits at a line start, so the first "  ast_grep_gate:" at or
	// after it is the section key itself (it is the next 2-space key).
	end := strings.Index(s[bodyStart:], "  ast_grep_gate:")
	if end < 0 {
		t.Fatalf("%s: no ast_grep_gate anchor after pre_commit", path)
	}
	return s[bodyStart : bodyStart+end]
}

// TestGatePreCommitTestScope is AC-PAS-011: default, accepted values,
// rejection, and agreement of the two shipped gate.yaml copies.
func TestGatePreCommitTestScope(t *testing.T) {
	t.Run("absent key defaults to affected", func(t *testing.T) {
		cfg, err := loadGateScopeMode(t, "gate:\n  enabled: true\n")
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if got := cfg.Gate.PreCommit.TestScope.Mode; got != "affected" {
			t.Errorf("default test_scope.mode = %q, want %q", got, "affected")
		}
	})

	t.Run("pre_commit without test_scope defaults to affected", func(t *testing.T) {
		cfg, err := loadGateScopeMode(t, "gate:\n  pre_commit:\n    enabled: true\n")
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if got := cfg.Gate.PreCommit.TestScope.Mode; got != "affected" {
			t.Errorf("test_scope.mode = %q, want %q", got, "affected")
		}
	})

	for _, mode := range []string{"affected", "full"} {
		t.Run("accepts "+mode, func(t *testing.T) {
			cfg, err := loadGateScopeMode(t, "gate:\n  pre_commit:\n    test_scope:\n      mode: "+mode+"\n")
			if err != nil {
				t.Fatalf("Validate rejected %q: %v", mode, err)
			}
			if got := cfg.Gate.PreCommit.TestScope.Mode; got != mode {
				t.Errorf("test_scope.mode = %q, want %q", got, mode)
			}
		})
	}

	t.Run("rejects any other value naming the key and the accepted values", func(t *testing.T) {
		_, err := loadGateScopeMode(t, "gate:\n  pre_commit:\n    test_scope:\n      mode: bogus\n")
		if err == nil {
			t.Fatal("Validate accepted test_scope.mode: bogus, want a rejection")
		}
		msg := err.Error()
		for _, want := range []string{"gate.pre_commit.test_scope.mode", "affected", "full", "bogus"} {
			if !strings.Contains(msg, want) {
				t.Errorf("rejection message %q does not contain %q", msg, want)
			}
		}
	})

	t.Run("shipped gate.yaml copies agree on the scope block", func(t *testing.T) {
		local := preCommitScopeBlock(t, filepath.Join("..", "..", ".moai", "config", "sections", "gate.yaml"))
		tmpl := preCommitScopeBlock(t, filepath.Join("..", "template", "templates", ".moai", "config", "sections", "gate.yaml"))
		if !strings.Contains(local, "test_scope:") || !strings.Contains(local, "mode: affected") {
			t.Errorf("local gate.yaml carries no test_scope block with mode: affected:\n%s", local)
		}
		if local != tmpl {
			t.Errorf("scope block differs between the two gate.yaml copies:\nlocal:\n%s\ntemplate:\n%s", local, tmpl)
		}
	})
}
