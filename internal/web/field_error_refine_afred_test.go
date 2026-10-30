//go:build afred

// field_error_refine_afred_test.go: B14 follow-up RED, committed before the
// repair. handleSave puts config.ValidateLLMYAMLSection's error text into the
// inline field errors as is (handlers.go, fieldErrs[field] = llmTypeErr.Error()).
// The error quotes the llm.yaml file's path, and for a type mismatch the value, so
// an address, home path, token or number reaches the page through the
// "llm.yaml" slot. The case driven here is an unreadable llm.yaml (a directory in
// its place): the error is a plain read error, keyed "llm.yaml", and carries the
// full path of the file. The project root is placed under a directory named
// with the payload, so the payload sits in that path. Excluded from the default
// run by the afred build tag; run it with:
//
//	go test -tags afred -run '^TestInlineFieldErrorIsRefined$' ./internal/web/
//
// The assertion reads only the err_llm.yaml element: the shell prints the project
// path elsewhere on purpose. Sentinel on failure: AF_FIELD_ERROR_REFINE.
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
