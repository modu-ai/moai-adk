// bundle_announce_test.go — U-03 (AC-RC30-04, REQ-RC30-005): the update output
// of a former plugin-mode project names the exact `moai bundle add <name>`
// command for each optional pack the project has not installed.
package cli

import (
	"strings"
	"testing"
)

// TestBundleAnnounceForPluginMode: the user-asset update phase names every
// optional pack the recorded selection lacks for a plugin-mode project, and
// names none for a local-mode project.
func TestBundleAnnounceForPluginMode(t *testing.T) {
	t.Run("plugin_mode_names_each_uninstalled_pack", func(t *testing.T) {
		output := runUserAssetPhaseInMode(t, "plugin")
		var missing []string
		for _, name := range bundleNames() {
			if !strings.Contains(output, "moai bundle add "+name) {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			t.Errorf("update output lacks the add command for %d optional pack(s) %v:\n%s", len(missing), missing, output)
		}
	})
	t.Run("local_mode_names_no_pack", func(t *testing.T) {
		output := runUserAssetPhaseInMode(t, "local")
		if strings.Contains(output, "moai bundle add") {
			t.Errorf("local-mode project got a bundle announce:\n%s", output)
		}
	})
}

// runUserAssetPhaseInMode runs the user-asset update phase from a project whose
// recorded deployment_mode is mode, with an empty user selection, and returns
// the phase output.
func runUserAssetPhaseInMode(t *testing.T, mode string) string {
	t.Helper()
	project := t.TempDir()
	writeFixtureFile(t, project, ".moai/config/sections/llm.yaml", "llm:\n  harness: claude\n  deployment_mode: "+mode+"\n")
	t.Chdir(project)
	var out strings.Builder
	if err := runUserAssetUpdatePhase(t.TempDir(), &out); err != nil {
		t.Fatal(err)
	}
	return out.String()
}
