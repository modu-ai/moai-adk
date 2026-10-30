// field_error_refine_test.go: B14 follow-up. handleSave wrote
// config.ValidateLLMYAMLSection's error into the inline field errors as is. With
// llm.yaml unreadable the error is a read error keyed "llm.yaml" that names the
// file's full path, so an address, home path, token or number in that path reached
// the err_llm.yaml element. Written first as the afred RED (four failing subtests),
// green with the message passing through refineText. The project root sits under
// a directory named with the payload; the assertion reads only that element
// because the shell prints the project path elsewhere on purpose.
//
// Sentinel on failure: AF_FIELD_ERROR_REFINE.
package web

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestInlineFieldErrorIsRefined(t *testing.T) {
	for cat, payload := range bannerPayloads {
		if strings.ContainsAny(payload, "\n") {
			continue // a directory name cannot hold a stack trace
		}
		if cat == "home-path" {
			// A real home path cannot be made the start of a temp directory's
			// path, and the sanitizer masks /Users/<name>/... only where it starts
			// an absolute path; the home-path category is covered where the message
			// is a string (TestErrorBannersAreRefined).
			continue
		}
		t.Run(cat, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), payload)
			a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
			a.recordLastProfile = func(string) error { return nil }
			var calls []saveStep
			recordingSeams(a, &calls, "")
			if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"), 0o700); err != nil {
				t.Skipf("the platform refuses this directory name: %v", err)
			}
			body := servePost(t, a.routes(), "/save", reproForm()).Body.String()
			m := regexp.MustCompile(`(?s)<p class="field__err" id="err_llm\.yaml">(.*?)</p>`).FindStringSubmatch(body)
			if m == nil {
				t.Fatalf("AF_FIELD_ERROR_REFINE: the llm.yaml inline error was not rendered, so this case proves nothing")
			}
			if strings.Contains(m[1], payload) {
				t.Errorf("AF_FIELD_ERROR_REFINE: the %s reached the inline error: %.200q", cat, m[1])
			}
		})
	}
}
