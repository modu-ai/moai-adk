// manifest_test.go — SPEC-PLUGIN-MARKETPLACE-001 M1 version guards (AC-003).
package pluginemit_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template/pluginemit"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// repoRoot is the repository root relative to this package's dir; the
// committed marketplace and manifests live there.
const repoRoot = "../../.."

// The four generated manifests, as repository-root-relative paths.
const (
	claudeMarketplacePath = ".claude-plugin/marketplace.json"
	codexMarketplacePath  = ".agents/plugins/marketplace.json"
	claudePluginPath      = "plugins/moai/.claude-plugin/plugin.json"
	codexPluginPath       = "plugins/moai/.codex-plugin/plugin.json"
)

// emptyCatalog is a catalog.yaml with no entries.
const emptyCatalog = "version: 1.0.0\ncatalog:\n  core:\n    skills: []\n"

// syntheticRaw builds the raw embed layout the generator reads: catalog.yaml at
// the root and the template tree files under templates/.
func syntheticRaw(catalog string, tree map[string]string) fs.FS {
	m := fstest.MapFS{"catalog.yaml": &fstest.MapFile{Data: []byte(catalog)}}
	for p, content := range tree {
		m["templates/"+p] = &fstest.MapFile{Data: []byte(content)}
	}
	return m
}

// syntheticTemplate builds a raw layout whose template tree holds only the
// given .mcp.json.
func syntheticTemplate(mcpJSON string) fs.FS {
	return syntheticRaw(emptyCatalog, map[string]string{".mcp.json": mcpJSON})
}

const defaultMCPJSON = `{
  "mcpServers": {
    "moai": {"command": "moai", "args": ["mcp-server"]},
    "context7": {"command": "npx", "args": ["-y", "@upstash/context7-mcp@latest"]}
  }
}`

// emitSynthetic emits over a synthetic tree with the default options (the
// version comes from the SSOT at call time).
func emitSynthetic(t *testing.T, mcpJSON string) *pluginemit.Publication {
	t.Helper()
	pub, err := pluginemit.Emit(syntheticTemplate(mcpJSON), pluginemit.DefaultOptions())
	if err != nil {
		t.Fatalf("Emit over synthetic tree: %v", err)
	}
	return pub
}

// decodeFile decodes one emitted file as a JSON object.
func decodeFile(t *testing.T, pub *pluginemit.Publication, path string) map[string]any {
	t.Helper()
	data, ok := pub.Files[path]
	if !ok {
		t.Fatalf("emitted set lacks %s (have %d files)", path, len(pub.Files))
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("%s is not a JSON object: %v", path, err)
	}
	return m
}

// versionFields returns the four version-carrying fields of the emitted set.
func versionFields(t *testing.T, get func(path string) map[string]any) map[string]any {
	t.Helper()
	cm := get(claudeMarketplacePath)
	metadata, _ := cm["metadata"].(map[string]any)
	plugins, _ := cm["plugins"].([]any)
	if metadata == nil || len(plugins) != 1 {
		t.Fatalf("%s: want metadata and exactly one plugin entry, got %v", claudeMarketplacePath, cm)
	}
	entry, _ := plugins[0].(map[string]any)
	return map[string]any{
		"claude marketplace metadata.version":   metadata["version"],
		"claude marketplace plugins[0].version": entry["version"],
		"claude plugin version":                 get(claudePluginPath)["version"],
		"codex plugin version":                  get(codexPluginPath)["version"],
	}
}

// TestVersionStampedFromSSOT is AC-003 (d): with version.Version set to a
// synthetic value, all four version-carrying fields carry it without the
// leading v, and the Codex marketplace carries no version field.
func TestVersionStampedFromSSOT(t *testing.T) {
	orig := version.Version
	version.Version = "v9.8.7-rc.1"
	t.Cleanup(func() { version.Version = orig })

	pub := emitSynthetic(t, defaultMCPJSON)
	get := func(path string) map[string]any { return decodeFile(t, pub, path) }

	for name, got := range versionFields(t, get) {
		if got != "9.8.7-rc.1" {
			t.Errorf("%s = %v, want %q (SSOT minus the leading v)", name, got, "9.8.7-rc.1")
		}
	}

	cx := get(codexMarketplacePath)
	if _, has := cx["version"]; has {
		t.Errorf("%s carries a marketplace-level version; the Codex shape has none (P-35)", codexMarketplacePath)
	}
	plugins, _ := cx["plugins"].([]any)
	if len(plugins) != 1 {
		t.Fatalf("%s: want exactly one plugin entry, got %d", codexMarketplacePath, len(plugins))
	}
	if entry, _ := plugins[0].(map[string]any); entry == nil {
		t.Errorf("%s: plugin entry is not an object", codexMarketplacePath)
	} else if _, has := entry["version"]; has {
		t.Errorf("%s carries an entry-level version; the Codex shape has none (P-35)", codexMarketplacePath)
	}
}

// TestCommittedVersionMatchesSSOT is AC-003 (e): the committed manifests
// carry the fallback value of the version SSOT minus its leading v.
func TestCommittedVersionMatchesSSOT(t *testing.T) {
	want := strings.TrimPrefix(version.Version, "v")
	get := func(path string) map[string]any {
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("committed manifest missing: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("%s is not a JSON object: %v", path, err)
		}
		return m
	}
	for name, got := range versionFields(t, get) {
		if got != want {
			t.Errorf("committed %s = %v, want %q — run `make plugin-emit`", name, got, want)
		}
	}
}
