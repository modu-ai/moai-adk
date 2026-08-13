package merge

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/defs"
	mrg "github.com/modu-ai/moai-adk/internal/merge"
)

// mergeWithDerivedBase runs the real strategy-aware engine over a derived base,
// which is the pairing MergeUserFiles performs. Asserting through the engine
// rather than on the base alone is deliberate: the base is only meaningful by
// the merge outcome it produces.
func mergeWithDerivedBase(t *testing.T, path, current, updated string) map[string]any {
	t.Helper()

	base, ok := deriveTemplateBase(path, []byte(current), []byte(updated))
	if !ok {
		t.Fatalf("no base derived for %s", path)
	}
	result, err := mrg.NewEngine().MergeFile(context.Background(), path, base, []byte(current), []byte(updated))
	if err != nil {
		t.Fatalf("MergeFile: %v", err)
	}
	var merged map[string]any
	if err := json.Unmarshal(result.Content, &merged); err != nil {
		t.Fatalf("unmarshal merged: %v\n%s", err, result.Content)
	}
	return merged
}

// TestMergeAddsTemplateEntryUserNeverHad is the regression case for the reported
// defect: a project whose .mcp.json predates the moai server entry never
// received it, because a base identical to the new template made the entry's
// absence read as a deletion the user had made.
func TestMergeAddsTemplateEntryUserNeverHad(t *testing.T) {
	const userFile = `{
  "mcpServers": {
    "context7": {"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y @upstash/context7-mcp@latest"]}
  }
}`
	const template = `{
  "mcpServers": {
    "moai": {"command": "moai", "args": ["mcp-server"]}
  },
  "staggeredStartup": {"enabled": true, "delayMs": 500}
}`

	merged := mergeWithDerivedBase(t, ".mcp.json", userFile, template)
	servers, _ := merged["mcpServers"].(map[string]any)

	if _, ok := servers["moai"]; !ok {
		t.Errorf("template entry mcpServers.moai was not added:\n%v", merged)
	}
	if _, ok := servers["context7"]; !ok {
		t.Errorf("user entry mcpServers.context7 was dropped:\n%v", merged)
	}
	if _, ok := merged["staggeredStartup"]; !ok {
		t.Errorf("top-level template key staggeredStartup was not added:\n%v", merged)
	}
}

// TestMergeKeepsUserEditToSharedKey guards the other direction: introducing
// template keys must not come at the cost of overwriting a value the user
// deliberately changed.
func TestMergeKeepsUserEditToSharedKey(t *testing.T) {
	const userFile = `{"model": "opus", "mcpServers": {"context7": {"command": "/bin/bash"}}}`
	const template = `{"model": "sonnet", "mcpServers": {"moai": {"command": "moai"}}}`

	merged := mergeWithDerivedBase(t, ".mcp.json", userFile, template)

	if got := merged["model"]; got != "opus" {
		t.Errorf("model = %v, want the user's value %q", got, "opus")
	}
	servers, _ := merged["mcpServers"].(map[string]any)
	if _, ok := servers["moai"]; !ok {
		t.Error("template entry was not added alongside the preserved user edit")
	}
	if _, ok := servers["context7"]; !ok {
		t.Error("user entry was dropped")
	}
}

// TestMergeAddsNestedTemplateKey pins the recursion. Narrowing that stopped at
// the container would leave a nested addition looking like a user deletion,
// which is exactly the mcpServers.moai shape.
func TestMergeAddsNestedTemplateKey(t *testing.T) {
	const userFile = `{"permissions": {"allow": ["Read"]}}`
	const template = `{"permissions": {"allow": ["Read"], "deny": ["Bash"]}}`

	merged := mergeWithDerivedBase(t, ".claude/settings.json", userFile, template)
	perms, _ := merged["permissions"].(map[string]any)

	if _, ok := perms["deny"]; !ok {
		t.Errorf("nested template key permissions.deny was not added:\n%v", merged)
	}
	if _, ok := perms["allow"]; !ok {
		t.Errorf("nested user key permissions.allow was dropped:\n%v", merged)
	}
}

// TestDeriveTemplateBaseDeclinesUnstructured covers the files that carry no key
// structure to narrow. Reporting no base keeps the caller on the
// preserve-user-content path it already used for them — a rendered shell script
// must not start line-merging against a fabricated base.
func TestDeriveTemplateBaseDeclinesUnstructured(t *testing.T) {
	for _, path := range []string{".moai/status_line.sh", "README.md", "Makefile", "logo.png"} {
		if _, ok := deriveTemplateBase(path, []byte("a\n"), []byte("b\n")); ok {
			t.Errorf("%s: derived a base for an unstructured file", path)
		}
	}
}

// TestDeriveTemplateBaseDeclinesUnparseable asserts a malformed or non-object
// side yields no base rather than a wrong one.
func TestDeriveTemplateBaseDeclinesUnparseable(t *testing.T) {
	cases := map[string][2]string{
		"current is malformed": {`{"a":`, `{"a": 1}`},
		"updated is malformed": {`{"a": 1}`, `{"a":`},
		"current is an array":  {`[1, 2]`, `{"a": 1}`},
		"updated is null":      {`{"a": 1}`, `null`},
		"current is empty":     {``, `{"a": 1}`},
	}
	for name, pair := range cases {
		t.Run(name, func(t *testing.T) {
			if _, ok := deriveTemplateBase(".mcp.json", []byte(pair[0]), []byte(pair[1])); ok {
				t.Error("derived a base from an unusable pair")
			}
		})
	}
}

// TestTemplateManaged covers the discriminator that keeps a user's own file out
// of the merge. The .tmpl form must count, because a rendered template lands at
// the path without that suffix — missing it would exclude .claude/settings.json,
// which is exactly the file that needs merging.
func TestTemplateManaged(t *testing.T) {
	embedded := fstest.MapFS{
		".mcp.json":                  {Data: []byte("{}")},
		".claude/settings.json.tmpl": {Data: []byte("{}")},
	}
	managed := map[string]bool{
		".mcp.json":                   true,
		".claude/settings.json":       true, // shipped as .tmpl
		".claude/settings.local.json": false,
		".moai/config/test.yaml":      false,
	}
	for path, want := range managed {
		if got := templateManaged(embedded, path); got != want {
			t.Errorf("templateManaged(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestMergeUserFilesAddsTemplateEntry drives the reported defect through the
// exported entry point rather than the helper: a .mcp.json that predates the
// moai server entry must come out of an update carrying it.
func TestMergeUserFilesAddsTemplateEntry(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".moai", "manifest.json"),
		[]byte(`{"version":"1","files":{}}`), defs.FilePerm); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// The deployed template is what sits at the path when the merge runs.
	deployed := `{"mcpServers": {"moai": {"command": "moai", "args": ["mcp-server"]}}}`
	if err := os.WriteFile(filepath.Join(tmpDir, ".mcp.json"), []byte(deployed), defs.FilePerm); err != nil {
		t.Fatalf("write deployed: %v", err)
	}

	userFile := `{"mcpServers": {"context7": {"command": "/bin/bash"}}}`
	var out strings.Builder
	if err := MergeUserFiles(tmpDir, []FileBackup{{Path: ".mcp.json", Data: []byte(userFile)}}, &out); err != nil {
		t.Fatalf("MergeUserFiles: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read merged: %v", err)
	}
	var merged map[string]any
	if err := json.Unmarshal(content, &merged); err != nil {
		t.Fatalf("unmarshal merged: %v\n%s", err, content)
	}
	servers, _ := merged["mcpServers"].(map[string]any)
	if _, ok := servers["moai"]; !ok {
		t.Errorf("moai entry missing after update:\n%s", content)
	}
	if _, ok := servers["context7"]; !ok {
		t.Errorf("user entry dropped after update:\n%s", content)
	}
}

// TestDeriveTemplateBaseSupportsYAML keeps the helper honest for the other
// structured format the merge engine compares by key, so a future mergeable
// YAML file does not silently fall through to the unstructured path.
func TestDeriveTemplateBaseSupportsYAML(t *testing.T) {
	base, ok := deriveTemplateBase("cfg.yaml", []byte("kept: 1\n"), []byte("kept: 2\nfresh: 3\n"))
	if !ok {
		t.Fatal("no base derived for a YAML pair")
	}
	if strings.Contains(string(base), "fresh") {
		t.Errorf("base carries a key the user's file lacks, which would read as a deletion:\n%s", base)
	}
	if !strings.Contains(string(base), "kept") {
		t.Errorf("base dropped a shared key:\n%s", base)
	}
}
