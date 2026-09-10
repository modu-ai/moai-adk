// Package cli — update_llm_preserve_test.go
//
// End-to-end preservation contract for .moai/config/sections/llm.yaml
// (SPEC-LLMCFG-PRESERVE-001, card t239). llm.yaml is gitignored,
// user-editable, machine-expanded at runtime by `moai glm`/`cg`, and sits
// inside the .moai/config root that CleanMoaiManagedPaths deletes wholesale
// before template redeploy. Its only protection is the update pipeline's
// Backup → Clean Managed Paths → Deploy Templates → Restore Settings cycle
// (3-way node merge, SPEC-UPDATE-YAML-PRESERVE-001) — until these tests, no
// named test drove the REAL embedded template llm.yaml through that cycle.
//
// Contract posture (spec.md §D.1, "keep + pin"): the pipeline already
// preserves user divergence while delivering new template keys. These tests
// pin that behavior; they do not add preservation machinery. Every assertion
// is at parsed-key + comment-presence granularity — byte equality is asserted
// ONLY in TestUpdateLLMYAMLFirstDeployCalm, where nothing was merged and the
// verbatim template bytes ARE the contract (acceptance.md AC-LCP-003).
//
// Fixture honesty (spec.md §E "Real-template fixture obligation"): the user
// fixture is derived from the REAL embedded template bytes
// (template.EmbeddedTemplates), never a hand-copied miniature. Each fixture
// edit applies with must-replace semantics — a template drift that removes
// the edited line fails the fixture builder instead of silently producing a
// vacuous green.

package cli

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ---- fixture constants (the user-edit pattern) ----

const (
	// llmUserPinnedHigh is the user's replacement for the template default
	// glm.models.high value ("glm-5.3-flash"). Divergent from base AND from
	// the default so the preservation assertion cannot pass by accident.
	llmUserPinnedHigh = "glm-user-pinned-model"

	// llmMarkerComment is the user's marker comment line. Placement is
	// load-bearing: the merge walks the NEW template's key nodes, so a comment
	// attached to a key the template carries is dropped with the template's
	// own key comment (measured 2026-09-02, backup node merge). The marker
	// therefore rides a USER-ADDED key, whose key node is carried through the
	// merge's old-only retention pass.
	llmMarkerComment = "# T239-LLM-PRESERVE-MARKER: user note pinned above a user-added key"

	// llmUserMarkerKey is the user-added key under llmMarkerComment. It is
	// absent from the template, so the merge retains it (and the attached
	// marker comment) as an old-only key.
	llmUserMarkerKey = "t239_user_marker_key"

	// llmMarkerKeyValue is the value written under llmUserMarkerKey.
	llmUserMarkerKeyValue = "observed"
)

// readEmbeddedLLMYAML returns the REAL embedded template llm.yaml bytes.
// This is the file the update cycle deploys and merges against — fixtures
// derive from it so assertions bind to the shipped 247-line comment-dense
// document (issue #1243 class), not a synthetic miniature.
func readEmbeddedLLMYAML() ([]byte, error) {
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(fsys, ".moai/config/sections/llm.yaml")
}

// embeddedLLMYAMLBytes is the *testing.T wrapper over readEmbeddedLLMYAML.
func embeddedLLMYAMLBytes(t *testing.T) []byte {
	t.Helper()
	data, err := readEmbeddedLLMYAML()
	if err != nil {
		t.Fatalf("read embedded llm.yaml: %v", err)
	}
	return data
}

// replaceOnce applies strings.Replace and fails the test when the needle was
// not found exactly once — the must-replace semantics that keep a template
// drift from turning a fixture edit into silent no-op (a silently unapplied
// edit would make AC-LCP-002's delivery assertion vacuously green).
func replaceOnce(t *testing.T, s, old, replacement string) string {
	t.Helper()
	if strings.Count(s, old) != 1 {
		t.Fatalf("fixture edit: needle %q appears %d times in template llm.yaml, want exactly 1 — the template drifted from the contract fixture", old, strings.Count(s, old))
	}
	return strings.Replace(s, old, replacement, 1)
}

// buildLLMUserYAML derives a user-edited llm.yaml from the embedded template
// bytes: glm.models.high re-pinned by the user, an agent_overrides entry the
// template default ({}) does not contain, and the marker comment + user-added
// key inserted under the template's agent_overrides comment block. With
// dropPerformanceTier, the template-carried llm.performance_tier line is
// removed — the delivery fixture for AC-LCP-002.
func buildLLMUserYAML(t *testing.T, base []byte, dropPerformanceTier bool) []byte {
	t.Helper()
	s := string(base)

	s = replaceOnce(t, s, `high: "glm-5.3-flash"`, `high: "`+llmUserPinnedHigh+`"`)
	s = replaceOnce(t, s, "  agent_overrides: {}",
		"  agent_overrides:\n"+
			"    manager-develop: { model: opus, effort: xhigh }\n"+
			llmMarkerComment+"\n"+
			"  "+llmUserMarkerKey+": "+llmUserMarkerKeyValue)

	if dropPerformanceTier {
		s = replaceOnce(t, s, "  performance_tier: \"medium\"\n", "")
	}
	return []byte(s)
}

// makeLLMPreserveFixture builds a t.TempDir() project that drives the
// template-sync update cycle (version mismatch trigger + manifest), with an
// optional user-edited llm.yaml.
func makeLLMPreserveFixture(t *testing.T, withUserLLM, dropPerformanceTier bool) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"system:\n  template_version: \"0.0.0\"\n")
	writeTestFile(t, root, ".moai/manifest.json", "{}\n")
	if withUserLLM {
		writeTestFile(t, root, ".moai/config/sections/llm.yaml",
			string(buildLLMUserYAML(t, embeddedLLMYAMLBytes(t), dropPerformanceTier)))
	}
	return root
}

// runTemplateSyncAt chdirs into the fixture and drives the REAL template-sync
// update cycle (runTemplateSyncWithReporter, skipConfirm) — Backup → Clean
// Managed Paths → Deploy Templates → Restore Settings — returning the stdout
// buffer for advisory assertions. Chdir makes this test family serial (no
// t.Parallel), the same trade the seam guards already make.
func runTemplateSyncAt(t *testing.T, root string) string {
	t.Helper()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if chErr := os.Chdir(root); chErr != nil {
		t.Fatalf("chdir to fixture: %v", chErr)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("config", false, "")
	_ = cmd.Flags().Set("yes", "true")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetContext(context.Background())

	if syncErr := runTemplateSyncWithReporter(cmd, nil, true); syncErr != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v\noutput: %s", syncErr, buf.String())
	}
	return buf.String()
}

// llmYAMLDoc reads the fixture's llm.yaml and parses it into a map.
func llmYAMLDoc(t *testing.T, root string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai/config/sections/llm.yaml"))
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal llm.yaml: %v", err)
	}
	return doc
}

// llmYAMLString resolves a dotted path (arbitrary depth) in the fixture's
// llm.yaml and returns its string value — the parsed-key granularity the
// contract asserts at.
func llmYAMLString(t *testing.T, root string, path ...string) string {
	t.Helper()
	var cur any = llmYAMLDoc(t, root)
	for _, key := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("llm.yaml: path %v: reached a non-mapping at %q", path, key)
		}
		cur, ok = m[key]
		if !ok {
			t.Fatalf("llm.yaml: key %q missing (path %v)", key, path)
		}
	}
	s, ok := cur.(string)
	if !ok {
		t.Fatalf("llm.yaml: path %v is not a string (got %T)", path, cur)
	}
	return s
}

// readLLMYAML returns the fixture's llm.yaml raw bytes (comment assertions).
func readLLMYAML(t *testing.T, root string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai/config/sections/llm.yaml"))
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	return data
}

// TestUpdateLLMYAMLPreserveTemplateSync covers AC-LCP-001: the template-sync
// update cycle preserves the user's divergent llm.yaml values (re-pinned
// glm.models.high, a user agent_overrides entry) and the marker comment, all
// against the REAL embedded template.
func TestUpdateLLMYAMLPreserveTemplateSync(t *testing.T) {
	root := makeLLMPreserveFixture(t, true, false)
	runTemplateSyncAt(t, root)

	if got := llmYAMLString(t, root, "llm", "glm", "models", "high"); got != llmUserPinnedHigh {
		t.Errorf("llm.glm.models.high = %q; want user pin %q (template default reset the user's value?)", got, llmUserPinnedHigh)
	}
	if got := llmYAMLString(t, root, "llm", "agent_overrides", "manager-develop", "model"); got != "opus" {
		t.Errorf("llm.agent_overrides.manager-develop.model = %q; want opus (user's agent_overrides entry lost?)", got)
	}
	if got := llmYAMLString(t, root, "llm", llmUserMarkerKey); got != llmUserMarkerKeyValue {
		t.Errorf("llm.%s = %q; want %q (user-added key lost?)", llmUserMarkerKey, got, llmUserMarkerKeyValue)
	}
	content := string(readLLMYAML(t, root))
	if !strings.Contains(content, llmMarkerComment) {
		t.Errorf("marker comment did not survive the update cycle; llm.yaml:\n%s", content)
	}
}

// TestUpdateLLMYAMLNewKeyDelivery covers AC-LCP-002: a template-carried key
// the user's file lacks (performance_tier removed from the fixture) is
// delivered back with its template default — preservation must never
// fossilize the file — and the delivery does not cost the user's divergent
// values (both hold in the same update pass).
func TestUpdateLLMYAMLNewKeyDelivery(t *testing.T) {
	root := makeLLMPreserveFixture(t, true, true)
	runTemplateSyncAt(t, root)

	if got := llmYAMLString(t, root, "llm", "performance_tier"); got != "medium" {
		t.Errorf("llm.performance_tier = %q; want template default \"medium\" (removed key not delivered?)", got)
	}
	if got := llmYAMLString(t, root, "llm", "glm", "models", "high"); got != llmUserPinnedHigh {
		t.Errorf("delivery cost preservation: llm.glm.models.high = %q; want user pin %q", got, llmUserPinnedHigh)
	}
	if got := llmYAMLString(t, root, "llm", llmUserMarkerKey); got != llmUserMarkerKeyValue {
		t.Errorf("delivery cost preservation: llm.%s = %q; want %q", llmUserMarkerKey, got, llmUserMarkerKeyValue)
	}
}

// TestUpdateLLMYAMLFirstDeployCalm covers AC-LCP-003: for a project with NO
// llm.yaml, the update deploys the template default verbatim — byte-identical
// (nothing was merged, so the byte-equality ban does not apply here) — and
// the update output carries no preservation advisory for llm.yaml.
func TestUpdateLLMYAMLFirstDeployCalm(t *testing.T) {
	root := makeLLMPreserveFixture(t, false, false)
	out := runTemplateSyncAt(t, root)

	disk, err := os.ReadFile(filepath.Join(root, ".moai/config/sections/llm.yaml"))
	if err != nil {
		t.Fatalf("llm.yaml absent after update on a project that had none: %v", err)
	}
	embedded := embeddedLLMYAMLBytes(t)
	if !bytes.Equal(disk, embedded) {
		t.Errorf("first deploy must ship the template llm.yaml verbatim; disk %d bytes vs embedded %d bytes", len(disk), len(embedded))
	}
	if strings.Contains(out, "llm.yaml") {
		t.Errorf("update output reports a preservation advisory for llm.yaml on a fresh install; output:\n%s", out)
	}
}

// TestUpdateLLMYAMLCommentsSurvive covers AC-LCP-004: the merged llm.yaml
// retains the template's comment documentation — the Profile matrix block
// header and the GLM reasoning-effort collapse comment — the issue #1243
// comment-preservation class, asserted against REAL template bytes.
func TestUpdateLLMYAMLCommentsSurvive(t *testing.T) {
	root := makeLLMPreserveFixture(t, true, false)
	runTemplateSyncAt(t, root)

	content := string(readLLMYAML(t, root))
	for _, sentinel := range []string{
		"Profile matrix (transparency + editability)",
		"GLM reasoning-effort mapping",
		llmMarkerComment,
	} {
		if !strings.Contains(content, sentinel) {
			t.Errorf("comment sentinel %q did not survive the merge; llm.yaml:\n%s", sentinel, content)
		}
	}
}
