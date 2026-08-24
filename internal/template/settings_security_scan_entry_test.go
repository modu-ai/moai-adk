package template

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRenderedSettingsHasNoSecurityScanPostToolEntry — AC-SSS-015.
// The distributed settings template must not register handle-security-scan.sh
// on the Write|Edit|MultiEdit PostToolUse matcher: the guardian buffer scan now
// runs inside the post-tool handler's own process. The wrapper script itself is
// retained, so a user project whose settings still name it keeps working.
//
// The template is Go-templated and is not valid JSON before rendering, so this
// is asserted through a render rather than through a raw-text scan.
func TestRenderedSettingsHasNoSecurityScanPostToolEntry(t *testing.T) {
	t.Parallel()

	for _, platform := range []string{"darwin", "linux", "windows"} {
		t.Run(platform, func(t *testing.T) {
			rendered := strings.TrimSpace(renderTemplate(t, ".claude/settings.json.tmpl", testContext(platform)))

			var settings struct {
				Hooks struct {
					PostToolUse []struct {
						Matcher string `json:"matcher"`
						Hooks   []struct {
							Args []string `json:"args"`
						} `json:"hooks"`
					} `json:"PostToolUse"`
				} `json:"hooks"`
			}
			if err := json.Unmarshal([]byte(rendered), &settings); err != nil {
				t.Fatalf("rendered settings.json is not valid JSON for %s: %v", platform, err)
			}

			found := 0
			for _, group := range settings.Hooks.PostToolUse {
				if group.Matcher != "Write|Edit|MultiEdit" {
					continue
				}
				for _, h := range group.Hooks {
					for _, arg := range h.Args {
						if strings.Contains(arg, "handle-security-scan") {
							found++
						}
					}
				}
			}
			if found != 0 {
				t.Errorf("rendered settings still registers handle-security-scan on Write|Edit|MultiEdit (%d arg matches)", found)
			}
		})
	}
}
