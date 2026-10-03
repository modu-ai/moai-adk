// golden_test.go — SPEC-PLUGIN-MARKETPLACE-001 M1 golden guards.
//
// TestManifestsGolden pins the SHAPE of the four manifests against hand-
// reviewed files under testdata/golden, using a synthetic version and a
// synthetic template tree (so it moves with the generator, never with a
// release bump). TestGoldenCommittedArtifactsMatchEmission is the drift
// guard: the emitter run over the REAL template tree and the real version
// SSOT must equal the files committed at the repository root. Both are
// switched into regeneration mode by PLUGIN_EMIT_UPDATE=1:
//
//	PLUGIN_EMIT_UPDATE=1 go test ./internal/template/pluginemit/... -run 'Test(ManifestsGolden|GoldenCommittedArtifactsMatchEmission)$'
//
// (the `make plugin-emit` target wraps this).
package pluginemit_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/pluginemit"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// rawTemplateDir is the raw embed layout root (catalog.yaml beside templates/)
// relative to this package's dir.
const rawTemplateDir = ".."

// goldenFiles maps each emitted path to its shape-pin file under testdata: the
// four manifests and the payload .mcp.json (the one payload file a tree with no
// skills or commands still emits).
var goldenFiles = map[string]string{
	claudeMarketplacePath:    "claude-marketplace.json",
	codexMarketplacePath:     "codex-marketplace.json",
	claudePluginPath:         "claude-plugin.json",
	codexPluginPath:          "codex-plugin.json",
	"plugins/moai/.mcp.json": "payload-mcp.json",
}

func updateMode() bool { return os.Getenv(pluginemit.EnvUpdate) == "1" }

// TestManifestsGolden compares the four manifests and the payload .mcp.json, emitted at version
// v1.2.3 over a synthetic tree, with the hand-reviewed shape pins.
func TestManifestsGolden(t *testing.T) {
	orig := version.Version
	version.Version = "v1.2.3"
	t.Cleanup(func() { version.Version = orig })

	pub := emitSynthetic(t, defaultMCPJSON)

	paths := make([]string, 0, len(goldenFiles))
	for p := range goldenFiles {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		golden := filepath.Join("testdata", "golden", goldenFiles[p])
		got, ok := pub.Files[p]
		if !ok {
			t.Errorf("%s: not emitted", p)
			continue
		}
		if updateMode() {
			if err := os.WriteFile(golden, got, 0o644); err != nil {
				t.Fatalf("update write %s: %v", golden, err)
			}
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("read golden %s: %v", golden, err)
		}
		if string(got) != string(want) {
			t.Errorf("%s differs from %s\n--- got ---\n%s\n--- want ---\n%s", p, golden, got, want)
		}
	}
	if len(pub.Files) != len(goldenFiles) {
		t.Errorf("emitted %d files, want exactly the %d pinned files", len(pub.Files), len(goldenFiles))
	}
}

// TestGoldenCommittedArtifactsMatchEmission is the drift guard: the emitter run
// over the real template tree must equal the committed tree at the repository
// root — bytes, modes, missing files and extra files (REQ-009). With
// PLUGIN_EMIT_UPDATE=1 it regenerates the committed tree instead.
func TestGoldenCommittedArtifactsMatchEmission(t *testing.T) {
	pub, err := pluginemit.Emit(os.DirFS(rawTemplateDir), pluginemit.DefaultOptions())
	if err != nil {
		t.Fatalf("Emit over the real template tree: %v", err)
	}
	if len(pub.Files) == 0 {
		t.Fatal("emitted set is empty — nothing was compared")
	}

	if updateMode() {
		if err := pluginemit.Write(pub, repoRoot); err != nil {
			t.Fatalf("update write: %v", err)
		}
		t.Logf("regenerated %d files", len(pub.Files))
		return
	}

	diffs, err := pluginemit.Drift(pub, repoRoot)
	if err != nil {
		t.Fatalf("Drift: %v", err)
	}
	for _, d := range diffs {
		t.Errorf("%s: %s — run `make plugin-emit` or stop hand-editing", d.Path, d.Kind)
	}
}
