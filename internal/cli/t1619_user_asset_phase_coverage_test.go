package cli

// t1619_user_asset_phase_coverage_test.go — card t1619, bundle B2: the unexecuted changed line of
// user_asset_phase.go (announceUnselectedOptionalPacks records the selection on a plugin-mode
// project). Family: TestBundle.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBundleAnnounceUnselectedPacksOmitsSelectedPack — a plugin-mode project is told about each
// optional pack its recorded selection lacks, and never about a pack it selected.
func TestBundleAnnounceUnselectedPacksOmitsSelectedPack(t *testing.T) {
	projectRoot := t.TempDir()
	sections := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("llm:\n  deployment_mode: plugin\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	names := bundleNames()
	if len(names) < 2 {
		t.Fatalf("the catalog must carry at least two optional packs, got %v", names)
	}

	var out strings.Builder
	announceUnselectedOptionalPacks(projectRoot, []string{names[0]}, &out)
	got := out.String()
	if strings.Contains(got, "optional pack not installed: "+names[0]+" ") {
		t.Fatalf("a selected pack must not be announced as missing, got %q", got)
	}
	if !strings.Contains(got, "optional pack not installed: "+names[1]+" ") {
		t.Fatalf("an unselected pack must be announced, got %q", got)
	}
}
