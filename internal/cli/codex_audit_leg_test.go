package cli

import (
	"context"
	"os"
	"slices"
	"testing"
)

// The marker reaches only the codex child of an audit leg (t1788 brain ruling,
// condition a). Ordinary codex paths keep the inherited environment.
func TestCodexLegChildEnvMarksOnlyAuditLeg(t *testing.T) {
	if env := codexLegChildEnv(context.Background()); env != nil {
		t.Fatalf("ordinary codex child must inherit the environment (nil), got %d entries", len(env))
	}
	env := codexLegChildEnv(withCodexAuditLeg(context.Background()))
	if !slices.Contains(env, codexAuditLegEnv+"=1") {
		t.Fatalf("audit-leg child environment lacks %s=1", codexAuditLegEnv)
	}
}

// Building a leg child environment must not mark the spawning process (condition a).
func TestCodexLegMarkerDoesNotLeakToParent(t *testing.T) {
	t.Setenv(codexAuditLegEnv, "")
	_ = codexLegChildEnv(withCodexAuditLeg(context.Background()))
	if got := os.Getenv(codexAuditLegEnv); got != "" {
		t.Fatalf("building a leg child environment changed the parent: %s=%q", codexAuditLegEnv, got)
	}
}

// The codex child command carries the marker only when a leg spawns it.
func TestCodexChildCmdCarriesMarkerOnlyInLeg(t *testing.T) {
	unsetenvForTest(t, codexAuditLegEnv) // hermetic: ambient values are covered in codex_child_orca_env_test.go
	ordinary := codexChildCmd(context.Background(), "codex", []string{"app-server"})
	// t1869: the ordinary child carries the parent environment minus the Orca
	// pane-owned keys, so Env is set; what stays true is that it has no marker.
	if slices.Contains(ordinary.Env, codexAuditLegEnv+"=1") {
		t.Fatalf("ordinary codex child must not carry %s=1", codexAuditLegEnv)
	}
	leg := codexChildCmd(withCodexAuditLeg(context.Background()), "codex", []string{"app-server"})
	if !slices.Contains(leg.Env, codexAuditLegEnv+"=1") {
		t.Fatalf("audit-leg codex child lacks %s=1", codexAuditLegEnv)
	}
}

func TestCodexAuditLegMarkerIsReadFromEnv(t *testing.T) {
	t.Setenv(codexAuditLegEnv, "")
	if inCodexAuditLeg() {
		t.Fatal("an unset marker must not count as an audit leg")
	}
	t.Setenv(codexAuditLegEnv, "1")
	if !inCodexAuditLeg() {
		t.Fatal("the marker set to 1 must count as an audit leg")
	}
}
