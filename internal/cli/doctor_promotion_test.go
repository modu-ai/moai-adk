package cli

// SPEC-DOCTOR-PROMOTION-001: table-driven spec for the "Plugin Deployment"
// doctor check (REQ-DP-001..004) — marker detection in both version: key
// shapes, promotion suggestion via 'moai init', zero noise, graceful degradation.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/pkg/version"
)

func TestCheckPluginDeployment(t *testing.T) {
	binary := version.GetVersion()

	tests := []struct {
		name         string
		content      string            // fixture system.yaml body; "" = no file
		unreadable   bool              // create system.yaml as a directory (read error)
		wantStatus   uikit.CheckStatus // "" = assert non-Fail only (graceful rows)
		wantInDetail []string
	}{
		{
			name:         "moai-rooted plugin-deployed marker warns",
			content:      "moai:\n  version: \"plugin-deployed v1.2.3\"\n  template_version: \"plugin-deployed v1.2.3\"\n",
			wantStatus:   uikit.CheckWarn,
			wantInDetail: []string{"moai init", "v1.2.3", binary},
		},
		{
			name:         "top-level plugin-deployed marker warns",
			content:      "version: plugin-deployed v2.0.0\n",
			wantStatus:   uikit.CheckWarn,
			wantInDetail: []string{"moai init", "v2.0.0", binary},
		},
		{
			name:       "plain semver is OK (zero noise)",
			content:    "moai:\n  version: \"v3.0.0-rc11\"\n  template_version: \"v3.0.0-rc11\"\n",
			wantStatus: uikit.CheckOK,
		},
		{name: "missing system.yaml is OK (zero noise)", wantStatus: uikit.CheckOK},
		{name: "malformed yaml without version key degrades gracefully", content: "\x00\x01 not: [valid: yaml\n\t###\n"},
		{name: "unreadable system.yaml (directory) degrades gracefully", unreadable: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			sysPath := filepath.Join(dir, defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML)
			if tt.unreadable {
				if err := os.MkdirAll(sysPath, 0o755); err != nil {
					t.Fatalf("mkdir system.yaml dir: %v", err)
				}
			} else if tt.content != "" {
				if err := os.MkdirAll(filepath.Dir(sysPath), 0o755); err != nil {
					t.Fatalf("mkdir sections dir: %v", err)
				}
				if err := os.WriteFile(sysPath, []byte(tt.content), 0o644); err != nil {
					t.Fatalf("write system.yaml fixture: %v", err)
				}
			}

			check := checkPluginDeployment(dir, true)

			if check.Name != "Plugin Deployment" {
				t.Errorf("Name = %q, want %q", check.Name, "Plugin Deployment")
			}
			if check.Status == uikit.CheckFail { // REQ-DP-004: never Fail
				t.Fatalf("checkPluginDeployment must never return Fail (REQ-DP-004), got %+v", check)
			}
			if tt.wantStatus != "" && check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", check.Status, tt.wantStatus)
			}
			for _, sub := range tt.wantInDetail {
				if !strings.Contains(check.Detail, sub) {
					t.Errorf("Detail = %q, should contain %q", check.Detail, sub)
				}
			}
			if tt.wantStatus == uikit.CheckWarn && !strings.Contains(check.Message, binary) {
				t.Errorf("Warn Message = %q, should name binary version %q", check.Message, binary)
			}
		})
	}
}
