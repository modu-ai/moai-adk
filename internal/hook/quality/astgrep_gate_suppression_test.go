package quality

// astgrep_gate_suppression_test.go covers AC-UTIL-002-10/11/12/20.
//
// Tests for checkSuppressionPairing and its integration with RunAstGrepGateV2.
// All tests use t.TempDir() for isolation per TRUST 5 Tested pillar.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSuppressionPairing covers three scenarios per AC-UTIL-002-20.
func TestSuppressionPairing(t *testing.T) {
	t.Parallel()

	// Scenario 1: correctly paired suppression → 0 violations (AC-UTIL-002-10)
	t.Run("paired-correctly", func(t *testing.T) {
		t.Parallel()

		src := `package main

// ast-grep-ignore
// @MX:REASON generated stub, pattern false positive
buf := getUnsafeBuffer()
`
		tmpDir := t.TempDir()
		fp := filepath.Join(tmpDir, "paired.go")
		if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		violations := checkSuppressionPairing(fp)
		if len(violations) != 0 {
			t.Errorf("expected 0 violations for correctly paired suppression, got %d: %+v",
				len(violations), violations)
		}
	})

	// Scenario 2: unpaired suppression → 1 SUPPRESSION_WITHOUT_REASON (AC-UTIL-002-11)
	t.Run("unpaired", func(t *testing.T) {
		t.Parallel()

		src := `package main

// ast-grep-ignore
buf := getUnsafeBuffer()
`
		tmpDir := t.TempDir()
		fp := filepath.Join(tmpDir, "unpaired.go")
		if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		violations := checkSuppressionPairing(fp)
		if len(violations) != 1 {
			t.Fatalf("expected 1 violation for unpaired suppression, got %d: %+v",
				len(violations), violations)
		}

		v := violations[0]
		if v.Type != "SUPPRESSION_WITHOUT_REASON" {
			t.Errorf("violation.Type = %q, want SUPPRESSION_WITHOUT_REASON", v.Type)
		}
		if v.File != fp {
			t.Errorf("violation.File = %q, want %q", v.File, fp)
		}
		if v.Line <= 0 {
			t.Errorf("violation.Line = %d; must be positive (1-indexed)", v.Line)
		}
		if !strings.HasPrefix(v.Message, "ast-grep suppression at") {
			t.Errorf("violation.Message = %q; must start with 'ast-grep suppression at'", v.Message)
		}
	})

	// Scenario 3: suppression separated by intervening non-blank comment → 1 violation (AC-UTIL-002-20)
	t.Run("separated-by-unrelated-comment", func(t *testing.T) {
		t.Parallel()

		src := `package main

// ast-grep-ignore
// some unrelated comment here
// @MX:REASON this reason is too far
buf := getUnsafeBuffer()
`
		tmpDir := t.TempDir()
		fp := filepath.Join(tmpDir, "separated.go")
		if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		violations := checkSuppressionPairing(fp)
		if len(violations) != 1 {
			t.Fatalf("expected 1 violation when suppression separated by non-blank comment, got %d: %+v",
				len(violations), violations)
		}
		if violations[0].Type != "SUPPRESSION_WITHOUT_REASON" {
			t.Errorf("violation.Type = %q, want SUPPRESSION_WITHOUT_REASON", violations[0].Type)
		}
	})

	// Scenario 4: suppression with exactly 1 blank line between ignore and reason → 0 violations
	t.Run("one-blank-line-allowed", func(t *testing.T) {
		t.Parallel()

		src := `package main

// ast-grep-ignore

// @MX:REASON one blank line is allowed
buf := getUnsafeBuffer()
`
		tmpDir := t.TempDir()
		fp := filepath.Join(tmpDir, "one_blank.go")
		if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		violations := checkSuppressionPairing(fp)
		if len(violations) != 0 {
			t.Errorf("expected 0 violations with 1 blank line between ignore and reason, got %d: %+v",
				len(violations), violations)
		}
	})

	// Scenario 5: suppression with 2 blank lines between ignore and reason → 1 violation
	t.Run("two-blank-lines-too-far", func(t *testing.T) {
		t.Parallel()

		src := `package main

// ast-grep-ignore


// @MX:REASON two blank lines is too far
buf := getUnsafeBuffer()
`
		tmpDir := t.TempDir()
		fp := filepath.Join(tmpDir, "two_blank.go")
		if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		violations := checkSuppressionPairing(fp)
		if len(violations) != 1 {
			t.Fatalf("expected 1 violation with 2 blank lines (too far), got %d: %+v",
				len(violations), violations)
		}
	})

	// Scenario 6: Ruby/Elixir prefix # ast-grep-ignore → correctly paired → 0 violations
	t.Run("ruby-prefix-paired", func(t *testing.T) {
		t.Parallel()

		src := `# Ruby code
x = nil
# ast-grep-ignore
# @MX:REASON test fixture, nil assignment intentional
y = nil
`
		tmpDir := t.TempDir()
		fp := filepath.Join(tmpDir, "test.rb")
		if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		violations := checkSuppressionPairing(fp)
		if len(violations) != 0 {
			t.Errorf("expected 0 violations for correctly paired Ruby suppression, got %d: %+v",
				len(violations), violations)
		}
	})

	// Scenario 7: Ruby prefix unpaired → 1 violation
	t.Run("ruby-prefix-unpaired", func(t *testing.T) {
		t.Parallel()

		src := `# Ruby code
# ast-grep-ignore
x = nil
`
		tmpDir := t.TempDir()
		fp := filepath.Join(tmpDir, "test.rb")
		if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		violations := checkSuppressionPairing(fp)
		if len(violations) != 1 {
			t.Fatalf("expected 1 violation for unpaired Ruby suppression, got %d: %+v",
				len(violations), violations)
		}
	})
}

// TestSuppressionPairing_MultipleViolations verifies that multiple unpaired suppressions
// are all reported.
func TestSuppressionPairing_MultipleViolations(t *testing.T) {
	t.Parallel()

	src := `package main

// ast-grep-ignore
buf1 := getUnsafeBuffer()

// ast-grep-ignore
buf2 := getUnsafeBuffer()
`
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "multi.go")
	if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations := checkSuppressionPairing(fp)
	if len(violations) != 2 {
		t.Errorf("expected 2 violations for 2 unpaired suppressions, got %d: %+v",
			len(violations), violations)
	}
}

// TestSuppressionPairing_EmptyMXReason verifies that an empty @MX:REASON text fails.
func TestSuppressionPairing_EmptyMXReason(t *testing.T) {
	t.Parallel()

	src := `package main

// ast-grep-ignore
// @MX:REASON
buf := getUnsafeBuffer()
`
	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "empty_reason.go")
	if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations := checkSuppressionPairing(fp)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for empty @MX:REASON text, got %d: %+v",
			len(violations), violations)
	}
}

// TestRunAstGrepGateV2_SuppressionViolation verifies that RunAstGrepGateV2
// returns (false, output) containing SUPPRESSION_WITHOUT_REASON on unpaired ignore.
// AC-UTIL-002-12
func TestRunAstGrepGateV2_SuppressionViolation(t *testing.T) {
	// Note: t.Setenv is not compatible with t.Parallel()
	// Disable PATH to prevent sg from running (graceful degradation);
	// the suppression lint runs independently of sg.
	t.Setenv("PATH", "")

	projectDir := t.TempDir()
	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: true,
		WarnOnlyMode: false,
	}

	// Create a Go file with an unpaired suppression comment.
	srcDir := filepath.Join(projectDir, "pkg")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	goSrc := `package pkg

// ast-grep-ignore
x := dangerousOp()
`
	if err := os.WriteFile(filepath.Join(srcDir, "bad.go"), []byte(goSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	passed, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)
	if passed {
		t.Error("RunAstGrepGateV2 should return false when unpaired suppression is found")
	}
	if !strings.Contains(output, "SUPPRESSION_WITHOUT_REASON") {
		t.Errorf("output should contain SUPPRESSION_WITHOUT_REASON, got: %q", output)
	}
}

// TestRunAstGrepGateV2_PairedSuppressionPasses verifies that a correctly paired
// suppression does not cause the gate to fail.
func TestRunAstGrepGateV2_PairedSuppressionPasses(t *testing.T) {
	t.Setenv("PATH", "")

	projectDir := t.TempDir()
	cfg := &AstGrepGateConfig{
		Enabled:      true,
		RulesDir:     ".moai/config/astgrep-rules",
		BlockOnError: true,
		WarnOnlyMode: false,
	}

	srcDir := filepath.Join(projectDir, "pkg")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	goSrc := `package pkg

// ast-grep-ignore
// @MX:REASON legitimate false positive, pattern matches safe usage here
x := dangerousOp()
`
	if err := os.WriteFile(filepath.Join(srcDir, "ok.go"), []byte(goSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	passed, _ := RunAstGrepGateV2(context.Background(), projectDir, cfg)
	if !passed {
		t.Error("RunAstGrepGateV2 should return true when all suppressions are correctly paired")
	}
}

// TestCheckSuppressionPairing_UnsupportedExtension verifies that files with
// unsupported extensions return 0 violations (no false positives).
func TestCheckSuppressionPairing_UnsupportedExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fp := filepath.Join(tmpDir, "data.json")
	src := `{"key": "// ast-grep-ignore"}`
	if err := os.WriteFile(fp, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	violations := checkSuppressionPairing(fp)
	if len(violations) != 0 {
		t.Errorf("unsupported extension should produce 0 violations, got %d", len(violations))
	}
}

// TestCheckSuppressionPairing_NonExistentFile verifies that a missing file returns 0 violations.
func TestCheckSuppressionPairing_NonExistentFile(t *testing.T) {
	t.Parallel()
	violations := checkSuppressionPairing("/path/that/does/not/exist/file.go")
	if len(violations) != 0 {
		t.Errorf("non-existent file should return 0 violations, got %d", len(violations))
	}
}
