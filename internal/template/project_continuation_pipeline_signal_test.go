// project_continuation_pipeline_signal_test.go — card t447, audit finding F1
// on SPEC-PROJECT-CONTINUATION-KEY-001: under `pipeline`, an unhonoured
// carry degrades silently into default `card` behaviour, so the prose
// contract in doc-generation.md carries the same report-line pattern the
// unmatched-value path already has — a carry-commitment line in the Step 4.2
// summary and a stop-short naming inside the `pipeline` branch. These tests
// pin the fragments in the template source, assert the dogfood mirror is
// byte-identical to it, and assert the embedded copy the binary carries
// matches its source.
package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pipelineCarrySignalDoc is the prose contract under test, as a
// deployment-relative path under templates/.
const pipelineCarrySignalDoc = ".claude/skills/moai/workflows/project/doc-generation.md"

// TestPipelineCarrySignal_TemplateMirror asserts the three report-line
// fragments in the template source, one subtest each, so a RED run names
// exactly which fragment is missing rather than failing as one blob.
func TestPipelineCarrySignal_TemplateMirror(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("templates", pipelineCarrySignalDoc))
	if err != nil {
		t.Fatalf("read template copy: %v", err)
	}
	doc := string(raw)

	for _, tc := range []struct {
		name     string
		fragment string
	}{
		{"unmatched-value report line present", "is not one of none | card | pipeline"},
		{"pipeline carry-commitment line present", "this session is expected to carry past /moai plan"},
		{"stop-short naming clause present", "the carry stopped before the Implementation Kickoff Approval gate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(doc, tc.fragment) {
				t.Errorf("template copy of %s lacks the %s fragment: %q",
					pipelineCarrySignalDoc, tc.name, tc.fragment)
			}
		})
	}
}

// TestPipelineCarrySignal_LocalMirrorMatches asserts the dogfood mirror at
// the repository root carries the same edit byte-for-byte.
func TestPipelineCarrySignal_LocalMirrorMatches(t *testing.T) {
	local, err := os.ReadFile(filepath.Join("..", "..", pipelineCarrySignalDoc))
	if err != nil {
		t.Fatalf("read local copy: %v", err)
	}
	source, err := os.ReadFile(filepath.Join("templates", pipelineCarrySignalDoc))
	if err != nil {
		t.Fatalf("read template copy: %v", err)
	}
	if string(local) != string(source) {
		t.Errorf("%s: the local mirror differs from the template source — both copies must carry the same edit",
			pipelineCarrySignalDoc)
	}
}

// TestPipelineCarrySignal_EmbeddedMatchesSource asserts the embedded copy of
// the doc is byte-identical to its templates/ source.
func TestPipelineCarrySignal_EmbeddedMatchesSource(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	embedded, err := fs.ReadFile(fsys, pipelineCarrySignalDoc)
	if err != nil {
		t.Fatalf("embedded %s: %v", pipelineCarrySignalDoc, err)
	}
	source, err := os.ReadFile(filepath.Join("templates", pipelineCarrySignalDoc))
	if err != nil {
		t.Fatalf("source %s: %v", pipelineCarrySignalDoc, err)
	}
	if string(embedded) != string(source) {
		t.Errorf("%s: the embedded copy differs from its template source — the binary predates the edit (run `make build`)",
			pipelineCarrySignalDoc)
	}
}
