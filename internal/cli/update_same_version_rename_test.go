package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// A same-version update without --force returns at the outer version-match check,
// before the template-sync confirmation. The held-back skill.md rename must still run
// there, so a project that kept a lowercase skill.md is repaired by an ordinary update
// instead of never being repaired.
func TestSameVersionUpdate_AppliesDeferredSkillRename(t *testing.T) {
	projectDir := initIdentityProject(t)
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	dir := filepath.Join(projectDir, ".claude", "skills", "moai-domain-backend")
	if err := os.Remove(filepath.Join(dir, "SKILL.md")); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.md"), []byte("lower body\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("config", false, "")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(cmd.Context())

	skipped, err := runTemplateSyncWithProgress(cmd, false)
	if err != nil {
		t.Fatalf("runTemplateSyncWithProgress: %v\noutput: %s", err, buf.String())
	}
	if !skipped {
		t.Fatalf("fixture is not version-matched; the same-version path was not exercised\noutput: %s", buf.String())
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var sawUpper, sawLower bool
	for _, e := range entries {
		switch e.Name() {
		case "SKILL.md":
			sawUpper = true
		case "skill.md":
			sawLower = true
		}
	}
	if !sawUpper || sawLower {
		t.Errorf("same-version update left the lowercase skill.md (entries = %v); want SKILL.md", entries)
	}
}
