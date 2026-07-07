package template

import (
	"io/fs"
	"path"
	"strings"
	"testing"
)

// TestEmbeddedTemplatesNoRuntimeLogs closes a blind-spot between .gitignore
// and Go's //go:embed directive.
//
// .gitignore in templates/ has a "logs/" rule that prevents git from tracking
// runtime trace files under .moai/logs/. But Go's `//go:embed all:templates`
// directive in embed.go does NOT read .gitignore — it embeds every file on
// disk under templates/. Any trace-*.jsonl that lands in templates/.moai/logs/
// (or anywhere else under templates/) at build time is therefore compiled into
// the moai binary and shipped to every user via `moai init` / `moai update`.
//
// The 2026-07-07 incident: maintainer session trace-*.jsonl files (carrying
// InstructionsLoaded events + session_id values) were embedded into the
// distributed binary despite being git-ignored. The existing
// internal_content_leak_test.go did not catch them because it matches SPEC-ID
// / REQ-token patterns, not runtime-event JSON.
//
// This test fails if:
//  1. Any *.jsonl appears anywhere in the embedded tree (trace/task-metrics/etc
//     are runtime artifacts, never template content), OR
//  2. A non-.gitkeep file appears under the embedded .moai/logs/ tree (the
//     directory must be empty placeholder-only, matching .moai/state/).
//
// When this test fails: the offending files are on disk under
// internal/template/templates/ and got embedded. Remove them from disk and
// re-run `make build` — do NOT relax this test.
func TestEmbeddedTemplatesNoRuntimeLogs(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var offenders []string
	walkErr := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// (1) No .jsonl anywhere in the embedded template tree — runtime
		// trace/metrics serialization format, never template content.
		if strings.HasSuffix(p, ".jsonl") {
			offenders = append(offenders, p)
			return nil
		}
		// (2) Under .moai/logs/, only .gitkeep is permitted (matches the
		// .moai/state/ placeholder-only convention).
		if strings.HasPrefix(p, ".moai/logs/") && path.Base(p) != ".gitkeep" {
			offenders = append(offenders, p)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk embedded fs: %v", walkErr)
	}
	if len(offenders) > 0 {
		t.Errorf("embedded template contamination: runtime artifacts leaked into "+
			"the binary via //go:embed all:templates bypassing .gitignore — "+
			"remove from internal/template/templates/ and re-run `make build`: %v",
			offenders)
	}
}
