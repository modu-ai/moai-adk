package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// --dry-run promises that nothing on disk changes. --binary and --config used to
// run housekeeping before the dry-run branch, and --binary also reached the binary
// install; --config reached the interactive wizard, which writes the config
// sections on confirm. Both combinations are refused before any of that runs.
func TestRunUpdate_DryRunWithBinaryOrConfigChangesNothing(t *testing.T) {
	for _, flag := range []string{"binary", "config"} {
		t.Run(flag, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			markProject(t, root)
			writeStateFile(t, root, "todo/backlog.db", "", 48*time.Hour)
			shippedEmpty := filepath.Join(root, ".claude", "skills", "moai-lane-watchdog")
			if err := os.MkdirAll(shippedEmpty, 0o755); err != nil {
				t.Fatal(err)
			}

			// The wizard reads the terminal; a test must never wait on it.
			devNull, err := os.Open(os.DevNull)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = devNull.Close() }()
			origStdin := os.Stdin
			os.Stdin = devNull
			defer func() { os.Stdin = origStdin }()

			origDeps := deps
			deps = &Dependencies{}
			defer func() { deps = origDeps }()

			cmd := &cobra.Command{Use: "update-test"}
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.Flags().String("version", "", "")
			cmd.Flags().Bool("check", false, "")
			cmd.Flags().Bool("shell-env", false, "")
			cmd.Flags().Bool("config", false, "")
			cmd.Flags().Bool("binary", false, "")
			cmd.Flags().Bool("templates-only", false, "")
			cmd.Flags().Bool("verbose", false, "")
			cmd.Flags().Bool("dry-run", false, "")
			if err := cmd.Flags().Set(flag, "true"); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set("dry-run", "true"); err != nil {
				t.Fatal(err)
			}

			updateLedger.reset()
			runErr := runUpdate(cmd, nil)

			if _, err := os.Stat(filepath.Join(root, ".moai", "state", "todo", "backlog.db")); err != nil {
				t.Fatalf("--dry-run with --%s removed the state stub: %v", flag, err)
			}
			if _, err := os.Stat(shippedEmpty); err != nil {
				t.Fatalf("--dry-run with --%s removed the skill folder: %v", flag, err)
			}
			if runErr == nil || !strings.Contains(runErr.Error(), "--dry-run") {
				t.Fatalf("--dry-run with --%s: want a refusal naming --dry-run, got err=%v output=%q", flag, runErr, buf.String())
			}
		})
	}
}
