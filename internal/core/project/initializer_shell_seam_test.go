package project

// Shell-config step seam tests (SPEC-INIT-QUIET-WIZARD-001 M1, AC-IQW-004 /
// AC-IQW-016). Step 6 of Init writes the user's shell rc files; these tests
// swap the exported seam for a counting spy so the step can be observed
// without writing any real rc file.
//
// Seam-swapping discipline (AC-IQW-016): never t.Parallel; call t.Setenv
// before swapping the seam (Go testing panics if t.Parallel is then added);
// restore the original seam value with t.Cleanup.

import (
	"context"
	"log/slog"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/shell"
)

// TestConfigureShellEnvFn_DefaultIsProductionFunc reads the seam's default
// value directly: outside tests it must point at the production shell-config
// writer. It compares function identity only, not the function body.
func TestConfigureShellEnvFn_DefaultIsProductionFunc(t *testing.T) {
	got := reflect.ValueOf(ConfigureShellEnvFn).Pointer()
	want := reflect.ValueOf(defaultConfigureShellEnv).Pointer()
	if got != want {
		t.Errorf("shell-config seam default points at %#x, want defaultConfigureShellEnv %#x "+
			"(a previous test may not have restored the seam)", got, want)
	}
}

// TestInitializer_ShellConfigSeamGate runs the real Init with a counting spy
// in the shell-config seam and checks that the SkipShellConfig gate decides
// whether Step 6 reaches the seam: true bypasses it, false calls it once.
func TestInitializer_ShellConfigSeamGate(t *testing.T) {
	cases := []struct {
		name            string
		skipShellConfig bool
		wantCalls       int32
	}{
		{name: "skip_shell_config_true_bypasses_step6", skipShellConfig: true, wantCalls: 0},
		{name: "skip_shell_config_false_reaches_step6", skipShellConfig: false, wantCalls: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Redirect the private MoAI home before touching the seam. This
			// t.Setenv also makes any t.Parallel on this test panic.
			t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))

			var calls atomic.Int32
			origSeam := ConfigureShellEnvFn
			ConfigureShellEnvFn = func(_ *slog.Logger) (*shell.ConfigResult, error) {
				calls.Add(1)
				// Report "already configured" and never call the real writer.
				return &shell.ConfigResult{Skipped: true}, nil
			}
			t.Cleanup(func() { ConfigureShellEnvFn = origSeam })

			opts := InitOptions{
				ProjectRoot:     t.TempDir(),
				ProjectName:     "shell-seam-gate",
				Language:        "Go",
				UserName:        "tester",
				ConvLang:        "en",
				DevelopmentMode: "tdd",
				SkipShellConfig: tc.skipShellConfig,
			}

			if _, err := NewInitializer(nil, manifest.NewManager(), nil).Init(context.Background(), opts); err != nil {
				t.Fatalf("Init() error = %v", err)
			}

			if got := calls.Load(); got != tc.wantCalls {
				t.Errorf("shell-config seam calls = %d, want %d (SkipShellConfig=%t)",
					got, tc.wantCalls, tc.skipShellConfig)
			}
		})
	}
}
