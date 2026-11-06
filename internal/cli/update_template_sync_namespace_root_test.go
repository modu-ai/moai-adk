package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// Card t1799 (codex P3): the template-sync pruning step must report a refused
// namespace backup root the way the clean-reinstall step does. A symlinked
// .moai/backups was refused silently because only the old backup root was
// consulted for the warning.
func TestTemplateSync_SymlinkedNamespaceBackupRootReportsSkippedPruning(t *testing.T) {
	projectDir := initIdentityProject(t)
	outside := t.TempDir()
	keep := filepath.Join(outside, "keep.txt")
	if err := os.WriteFile(keep, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(projectDir, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	if err := os.Symlink(outside, link); err != nil {
		t.Skip(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("config", false, "")
	_ = cmd.Flags().Set("force", "true")
	_ = cmd.Flags().Set("yes", "true")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())
	if syncErr := runTemplateSyncWithReporter(cmd, nil, true, nil); syncErr != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v\nstderr: %s", syncErr, errBuf.String())
	}

	if !strings.Contains(errBuf.String(), "skipped old backup pruning") {
		t.Errorf("a symlinked namespace backup root was refused silently; stderr:\n%s", errBuf.String())
	}
	if got, rerr := os.ReadFile(keep); rerr != nil || string(got) != "outside" {
		t.Errorf("file outside the project changed: %q (err %v)", got, rerr)
	}
}
