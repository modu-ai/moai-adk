package quality

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestQualityGate_Run_AstGrepGuardReachable proves the ast-grep guard branch
// inside QualityGate.Run is reachable when AstGrepGate.Enabled is true.
//
// The project uses a Zig marker (build.zig) because the Zig toolchain has no
// vet/lint steps, so Run reaches the ast-grep step immediately; the test step
// comes after the ast-grep step and is never reached on failure. The observable
// proof is the pure-Go suppression-pairing check inside RunAstGrepGateV2: an
// unpaired ast-grep-ignore comment makes Run return (false, violations) — which
// can only happen if the guard branch executed. No sg binary is required.
func TestQualityGate_Run_AstGrepGuardReachable(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "build.zig"), []byte("// zig build marker\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	// Unpaired suppression: ast-grep-ignore without an adjacent @MX:REASON line.
	src := "// ast-grep-ignore\nconst x = 1;\n"
	if err := os.WriteFile(filepath.Join(projectDir, "main.ts"), []byte(src), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	gate := NewQualityGate(&GateConfig{
		Enabled:    true,
		SkipTests:  true,
		ProjectDir: projectDir,
		VetTimeout: 5 * time.Second,
		AstGrepGate: &AstGrepGateConfig{
			Enabled:      true,
			RulesDir:     ".moai/config/astgrep-rules",
			BlockOnError: true,
		},
	})

	passed, output := gate.Run(context.Background())
	if passed {
		t.Fatal("Run: want failure via the ast-grep guard branch, got pass")
	}
	if !strings.Contains(output, "suppression policy violations") {
		t.Errorf("output should carry the ast-grep suppression report, got: %q", output)
	}
}

// TestQualityGate_Run_AstGrepGuardDisabled is the default-OFF twin: with
// AstGrepGate.Enabled=false the same project passes (the guard is skipped).
func TestQualityGate_Run_AstGrepGuardDisabled(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "build.zig"), []byte("// zig build marker\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	src := "// ast-grep-ignore\nconst x = 1;\n"
	if err := os.WriteFile(filepath.Join(projectDir, "main.ts"), []byte(src), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	gate := NewQualityGate(&GateConfig{
		Enabled:     true,
		SkipTests:   true,
		ProjectDir:  projectDir,
		AstGrepGate: &AstGrepGateConfig{Enabled: false},
	})

	passed, _ := gate.Run(context.Background())
	if !passed {
		t.Error("Run: want pass when the ast-grep sub-gate is disabled")
	}
}
