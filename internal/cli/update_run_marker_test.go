package cli

// update_run_marker_test.go — D9 (SPEC-UPDATE-REPAIR-RC30-001, REQ-RC30-010):
// the normal update writes the run marker into the config backup it creates, so
// retention counts that directory as part of one run.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateBackupCarriesRunMarker(t *testing.T) {
	root := makeLLMPreserveFixture(t, false, false)
	runTemplateSyncAt(t, root)

	backupRoot := filepath.Join(root, ".moai-backups")
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		t.Fatalf("the update created no config backup under %s: %v", backupRoot, err)
	}
	marked := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(backupRoot, entry.Name(), ".moai-run"))
		if readErr == nil && strings.HasSuffix(string(data), "Z\n") {
			marked++
		}
	}
	if marked == 0 {
		t.Errorf("the update's config backup under %s carries no .moai-run marker; retention would count it as its own run", backupRoot)
	}
}
