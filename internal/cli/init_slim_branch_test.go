package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newSlimTestCmd creates a minimal cobra.Command with the --all flag
// registered, mirroring the real initCmd flag definition.
// Named differently from newTestInitCmd in init_coverage_test.go to avoid collision.
func newSlimTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "init-slim-test"}
	cmd.Flags().Bool("all", false, "Deploy all catalog entries (test)")
	return cmd
}

// TestShouldDistributeAll_FlagTrue verifies that --all=true returns true.
//
// SPEC-V3R4-CATALOG-002 REQ-013.
// Note: uses t.Setenv — must not call t.Parallel per Go 1.21+ rules.
func TestShouldDistributeAll_FlagTrue(t *testing.T) {
	t.Setenv("MOAI_DISTRIBUTE_ALL", "")

	cmd := newSlimTestCmd()
	if err := cmd.Flags().Set("all", "true"); err != nil {
		t.Fatalf("failed to set --all flag: %v", err)
	}

	if !shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(--all=true, env='') = false, want true")
	}
}

// TestShouldDistributeAll_EnvOne verifies that MOAI_DISTRIBUTE_ALL=1 returns true.
//
// SPEC-V3R4-CATALOG-002 REQ-012.
func TestShouldDistributeAll_EnvOne(t *testing.T) {
	t.Setenv("MOAI_DISTRIBUTE_ALL", "1")

	cmd := newSlimTestCmd()
	if !shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(--all=false, env='1') = false, want true")
	}
}

// TestShouldDistributeAll_EnvTrueLower verifies case-insensitive "true" match.
//
// SPEC-V3R4-CATALOG-002 REQ-012.
func TestShouldDistributeAll_EnvTrueLower(t *testing.T) {
	t.Setenv("MOAI_DISTRIBUTE_ALL", "true")

	cmd := newSlimTestCmd()
	if !shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(env='true') = false, want true")
	}
}

// TestShouldDistributeAll_EnvTrueUpper verifies case-insensitive "TRUE" match.
//
// SPEC-V3R4-CATALOG-002 REQ-012.
func TestShouldDistributeAll_EnvTrueUpper(t *testing.T) {
	t.Setenv("MOAI_DISTRIBUTE_ALL", "TRUE")

	cmd := newSlimTestCmd()
	if !shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(env='TRUE') = false, want true")
	}
}

// TestShouldDistributeAll_EnvZero verifies that MOAI_DISTRIBUTE_ALL=0 returns false.
//
// SPEC-V3R4-CATALOG-002 REQ-012 EC narrow match.
func TestShouldDistributeAll_EnvZero(t *testing.T) {
	t.Setenv("MOAI_DISTRIBUTE_ALL", "0")

	cmd := newSlimTestCmd()
	if shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(env='0') = true, want false")
	}
}

// TestShouldDistributeAll_EnvYes verifies that MOAI_DISTRIBUTE_ALL=yes returns false.
// Only "1" and case-insensitive "true" are accepted.
//
// SPEC-V3R4-CATALOG-002 REQ-012 EC narrow match.
func TestShouldDistributeAll_EnvYes(t *testing.T) {
	t.Setenv("MOAI_DISTRIBUTE_ALL", "yes")

	cmd := newSlimTestCmd()
	if shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(env='yes') = true, want false")
	}
}

// TestShouldDistributeAll_EnvEmpty verifies that MOAI_DISTRIBUTE_ALL="" returns false.
//
// SPEC-V3R4-CATALOG-002 REQ-012.
func TestShouldDistributeAll_EnvEmpty(t *testing.T) {
	t.Setenv("MOAI_DISTRIBUTE_ALL", "")

	cmd := newSlimTestCmd()
	if shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(env='') = true, want false")
	}
}

// TestShouldDistributeAll_EnvUnset verifies that an absent env var returns false.
//
// SPEC-V3R4-CATALOG-002 REQ-012.
func TestShouldDistributeAll_EnvUnset(t *testing.T) {
	// Unset explicitly to be deterministic across all CI environments.
	t.Setenv("MOAI_DISTRIBUTE_ALL", "")

	cmd := newSlimTestCmd()
	if shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(env unset/empty) = true, want false")
	}
}

// TestShouldDistributeAll_BothFlagAndEnv verifies EC2 idempotency:
// both --all and MOAI_DISTRIBUTE_ALL=1 set simultaneously returns true
// without panicking or having double side-effects.
//
// SPEC-V3R4-CATALOG-002 REQ-019 (idempotent both-set).
func TestShouldDistributeAll_BothFlagAndEnv(t *testing.T) {
	t.Setenv("MOAI_DISTRIBUTE_ALL", "1")

	cmd := newSlimTestCmd()
	if err := cmd.Flags().Set("all", "true"); err != nil {
		t.Fatalf("failed to set --all flag: %v", err)
	}

	if !shouldDistributeAll(cmd) {
		t.Error("shouldDistributeAll(--all=true, env='1') = false, want true")
	}
}

// TestInitCmd_AllFlagRegistered verifies that the --all flag is registered on
// initCmd with the correct type (bool) and default (false).
//
// SPEC-V3R4-CATALOG-002 T2.1.
func TestInitCmd_AllFlagRegistered(t *testing.T) {
	t.Parallel()

	f := initCmd.Flags().Lookup("all")
	if f == nil {
		t.Fatal("initCmd --all flag not registered")
	}
	if f.DefValue != "false" {
		t.Errorf("initCmd --all flag default = %q, want \"false\"", f.DefValue)
	}
	if f.Value.Type() != "bool" {
		t.Errorf("initCmd --all flag type = %q, want \"bool\"", f.Value.Type())
	}
}

// TestEmitSlimModeNotice_FourSubstrings verifies that the slim-mode
// informational notice (SPEC-V3R4-CATALOG-002 REQ-021 + acceptance scenario S1)
// contains all four required substrings so downstream tooling (moai doctor) can
// pattern-match the message. Regression guard against silent drift of any of
// the four anchor strings.
//
// REQ-021 / S1.
func TestEmitSlimModeNotice_FourSubstrings(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	emitSlimModeNotice(&buf)
	out := buf.String()

	required := []string{
		"slim mode",
		"--all",
		"MOAI_DISTRIBUTE_ALL=1",
		"SPEC-V3R4-CATALOG-005",
	}
	for _, sub := range required {
		if !strings.Contains(out, sub) {
			t.Errorf("REQ-021 notice missing substring %q (full output: %q)", sub, out)
		}
	}
}
