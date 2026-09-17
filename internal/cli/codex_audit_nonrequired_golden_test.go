package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// updateCodexAuditNonRequiredGolden regenerates the non-required codex_audit
// goldens. The goldens are captured from the handler BEFORE the required-gate
// verdict change lands, so a later byte diff proves the non-required path moved.
var updateCodexAuditNonRequiredGolden = os.Getenv("UPDATE_CODEX_AUDIT_GOLDEN") == "1"

// codexAuditGoldenPlaceholders are the fixed values the build identity fields
// are normalized to: build_commit and build_lag depend on the binary running
// the test, not on the audit behavior under comparison.
var codexAuditGoldenPlaceholders = map[string]string{
	"build_commit": "<build_commit>",
	"build_lag":    "<build_lag>",
}

// normalizeCodexAuditResult serializes a codex_audit tool result into a stable
// byte form: isError, the text content decoded and re-marshalled with sorted
// keys, and the structured content likewise, with the build identity fields
// pinned to placeholders.
func normalizeCodexAuditResult(t *testing.T, res *mcp.CallToolResult) []byte {
	t.Helper()
	normalize := func(raw []byte) map[string]any {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("decode result JSON %s: %v", raw, err)
		}
		for k, v := range codexAuditGoldenPlaceholders {
			m[k] = v
		}
		return m
	}
	var text string
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			text = tc.Text
		}
	}
	structured, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	out, err := json.MarshalIndent(map[string]any{
		"isError":    res.IsError,
		"text":       normalize([]byte(text)),
		"structured": normalize(structured),
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal normalized result: %v", err)
	}
	return append(out, '\n')
}

// TestCodexAudit_NonRequiredGateGoldenByteIdentical pins the codex_audit result
// for every non-required gate state (off, advisory, key absent, config file
// absent, corrupt YAML) with codex absent. A required-gate change must leave
// all five byte-identical after build identity normalization.
func TestCodexAudit_NonRequiredGateGoldenByteIdentical(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, root string)
	}{
		{"off", func(t *testing.T, root string) { writeCodexAuditGate(t, root, "off") }},
		{"advisory", func(t *testing.T, root string) { writeCodexAuditGate(t, root, "advisory") }},
		{"key-absent", func(t *testing.T, root string) {
			writeRawCodexGateWorkflow(t, root, "workflow:\n  audit:\n    gates:\n      glm: advisory\n")
		}},
		{"file-absent", func(t *testing.T, root string) {}},
		{"corrupt-yaml", func(t *testing.T, root string) {
			writeRawCodexGateWorkflow(t, root, "workflow:\n  audit: [unclosed\n    gates: {codex: required\n")
		}},
		{"required-with-trailing-space", func(t *testing.T, root string) {
			writeRawCodexGateWorkflow(t, root, "workflow:\n  audit:\n    gates:\n      codex: \"required \"\n")
		}},
		{"required-uppercase", func(t *testing.T, root string) { writeCodexAuditGate(t, root, "REQUIRED") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newProbeProject(t, "SPEC-GOLDEN-001")
			tc.setup(t, root)
			withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

			got := normalizeCodexAuditResult(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
			path := filepath.Join("testdata", "codex-audit-nonrequired", tc.name+".golden")
			if updateCodexAuditNonRequiredGolden {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(path, got, 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden %s: %v (UPDATE_CODEX_AUDIT_GOLDEN=1 regenerates)", path, err)
			}
			if string(got) != string(want) {
				t.Errorf("codex_audit result drifted for gate state %q\ngot:\n%s\nwant:\n%s", tc.name, got, want)
			}
		})
	}
}

// writeRawCodexGateWorkflow writes a raw workflow.yaml body under root.
func writeRawCodexGateWorkflow(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
