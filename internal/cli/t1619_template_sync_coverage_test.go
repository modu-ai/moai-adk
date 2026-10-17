package cli

// t1619_template_sync_coverage_test.go — card t1619, bundle B2: the unexecuted changed line of
// update_template_sync.go (the confirmation preview's deployer construction failure). Family:
// TestTemplateSync.

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/template"
)

// TestTemplateSyncDeployerConstructionFailureIsReturned — when the preview cannot construct
// the run's deployer, the sync returns that failure to its caller and does not skip.
func TestTemplateSyncDeployerConstructionFailureIsReturned(t *testing.T) {
	t.Chdir(t.TempDir())
	orig := newTemplateSyncDeployer
	t.Cleanup(func() { newTemplateSyncDeployer = orig })
	cause := errors.New("deployer unavailable for test")
	newTemplateSyncDeployer = func(_ fs.FS, _ template.DeployMode) (template.Deployer, error) {
		return nil, cause
	}

	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("no-plugin", false, "")
	var out strings.Builder
	cmd.SetOut(&out)

	skipped, err := runTemplateSyncWithProgress(cmd, false)
	if skipped {
		t.Fatal("a failed deployer construction must not be reported as a skipped sync")
	}
	if err == nil || !errors.Is(err, cause) || !strings.Contains(err.Error(), "construct harness deployer") {
		t.Fatalf("the construction failure must be returned to the caller, got %v", err)
	}
}
