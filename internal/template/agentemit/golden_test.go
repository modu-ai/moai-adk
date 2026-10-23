// golden_test.go — SPEC-CODEX-DUAL-AGENTS-001 MS3 golden guards, retargeted by
// SPEC-AGENT-CORE4-001 (REQ-AC4-011).
//
// These tests run the emitter against the REAL agent sources under
// internal/template/agentemit/source and pin the committed artifacts: the
// Codex TOML set under templates/.codex/agents/moai and the Markdown set under
// templates/.claude/agents/moai. They are the drift guard: a hand-edited copy
// or a behavior change in the emitter or manifest fails here until the copies
// are regenerated with `make agents-emit` (the agent source is the only edit
// surface). Nothing in this package writes a copy during a test run; the
// regeneration path is the agents-emit command.
package agentemit_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// templatesDir is the template tree root relative to this package's dir.
const templatesDir = "../templates"

// agentMDRoot is the rendered Claude definition set (C2) inside the template tree.
const agentMDRoot = ".claude/agents/moai"

// sourceRoot is the agent source directory, relative to the repository root.
const sourceRoot = "internal/template/agentemit/source"

// expectedMCPCarriers is the AC-007 inventory: exactly these 7 agents carry
// mcp__moai__* tokens (plan.md §A.2, verified against the template tree).
var expectedMCPCarriers = map[string]bool{
	"manager-develop": true, "manager-docs": true, "manager-lead": true,
	"manager-spec": true, "plan-auditor": true, "super-advisor": true,
	"sync-auditor": true,
}

// emitRealSet runs the emitter over the committed agent sources.
func emitRealSet(t *testing.T) *agentemit.Publication {
	t.Helper()
	man, err := agentemit.LoadManifest()
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	fsys := os.DirFS(repoRoot)
	pub, err := agentemit.EmitAll(fsys, sourceRoot, man)
	if err != nil {
		t.Fatalf("EmitAll over real agent sources: %v", err)
	}
	if len(pub.CodexTOML) != 12 {
		t.Fatalf("emitted %d TOMLs, want 12", len(pub.CodexTOML))
	}
	return pub
}

// committedTOMLPath resolves the committed artifact path for one emitted path.
func committedTOMLPath(emitted string) string {
	return filepath.Join(templatesDir, filepath.FromSlash(emitted))
}

// committedMarkdownPath resolves the committed C2 definition rendered from the
// agent source at sourcePath.
func committedMarkdownPath(sourcePath string) string {
	return filepath.Join(templatesDir, filepath.FromSlash(agentMDRoot), path.Base(sourcePath))
}

// markdownDiffs returns the sorted source paths whose rendered bytes differ
// from the committed C2 definition. Both the equality test and the mismatch
// test use it, so a comparison that cannot fail is caught by the second.
func markdownDiffs(t *testing.T, rendered map[string][]byte) []string {
	t.Helper()
	var diffs []string
	for p, data := range rendered {
		committed, err := os.ReadFile(committedMarkdownPath(p))
		if err != nil {
			t.Fatalf("%s: committed .md unreadable: %v", p, err)
		}
		if !bytes.Equal(committed, data) {
			diffs = append(diffs, p)
		}
	}
	sort.Strings(diffs)
	return diffs
}

// TestGoldenCommittedArtifactsMatchEmission is the AC-001/AC-004 drift guard:
// every emitted TOML must be byte-identical (sha256) to the committed
// artifact, and every rendered Claude definition equal to the committed .md.
// The copies are regenerated with `make agents-emit`.
func TestGoldenCommittedArtifactsMatchEmission(t *testing.T) {
	pub := emitRealSet(t)

	emittedPaths := make([]string, 0, len(pub.CodexTOML))
	for p := range pub.CodexTOML {
		emittedPaths = append(emittedPaths, p)
	}
	sort.Strings(emittedPaths)

	for _, p := range emittedPaths {
		data := pub.CodexTOML[p]
		sum := fmt.Sprintf("%x", sha256.Sum256(data))
		committed, err := os.ReadFile(committedTOMLPath(p))
		if err != nil {
			t.Errorf("%s: committed artifact missing (%v) — regenerate with `make agents-emit`", p, err)
			continue
		}
		if fmt.Sprintf("%x", sha256.Sum256(committed)) != sum {
			t.Errorf("%s: committed artifact differs from emission (sha256 mismatch) — edit the agent source and run `make agents-emit`", p)
		}
	}

	// AC-001's .md face: the Claude render of every agent source equals the
	// committed .md definition (REQ-AC4-011: equality with a fresh render
	// replaces the former identity pass-through).
	t.Run("markdown_equals_fresh_render", func(t *testing.T) {
		for _, p := range markdownDiffs(t, pub.Markdown) {
			t.Errorf("%s: committed .md differs from a fresh render of its agent source — edit the source and run `make agents-emit`", p)
		}
	})
}

// TestGoldenMarkdownMismatchDetected closes the mutants of the comparison above
// (`if false`, comparing a value with itself): one byte of a rendered
// definition is flipped in memory and the same comparison must report exactly
// that file.
func TestGoldenMarkdownMismatchDetected(t *testing.T) {
	pub := emitRealSet(t)
	if diffs := markdownDiffs(t, pub.Markdown); len(diffs) != 0 {
		t.Fatalf("precondition: committed .md must equal a fresh render, differs: %v", diffs)
	}
	t.Run("injected_mismatch_is_red", func(t *testing.T) {
		var victim string
		for p := range pub.Markdown {
			if victim == "" || p < victim {
				victim = p
			}
		}
		flipped := map[string][]byte{}
		for p, data := range pub.Markdown {
			flipped[p] = append([]byte(nil), data...)
		}
		flipped[victim][len(flipped[victim])/2] ^= 0x01
		diffs := markdownDiffs(t, flipped)
		if len(diffs) != 1 || diffs[0] != victim {
			t.Errorf("flipping one byte of %s must make the comparison report exactly that file, got %v", victim, diffs)
		}
	})
}

// TestRealSetMarkdownTreeUnmodified is AC-002: running the emitter (success
// or failure paths included — the failure face is covered by the MS1
// fail-closed tests, which never touch disk) leaves the committed .md tree
// byte-identical. The emitter holds no .md write path; this test makes that
// observable rather than assumed.
func TestRealSetMarkdownTreeUnmodified(t *testing.T) {
	before := hashMDTree(t)
	emitRealSet(t) // full success-path emission over the real tree
	after := hashMDTree(t)
	for name, sum := range before {
		if after[name] != sum {
			t.Errorf("%s: .md tree modified by emission (sha256 %s -> %s)", name, sum, after[name])
		}
	}
	if len(after) != len(before) {
		t.Errorf(".md file count changed: %d -> %d", len(before), len(after))
	}
}

func hashMDTree(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	entries, err := os.ReadDir(filepath.Join(templatesDir, filepath.FromSlash(agentMDRoot)))
	if err != nil {
		t.Fatalf("read md tree: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(templatesDir, filepath.FromSlash(agentMDRoot), e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		out[e.Name()] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	if len(out) != 12 {
		t.Fatalf("expected 12 .md sources, found %d", len(out))
	}
	return out
}

// realSetCodexModelEffort is the per-role Codex model and reasoning effort the
// emitter derives from the agent frontmatter (card t1808): opus and sonnet map
// to the sol model, haiku to luna; luna roles run at high effort, and
// manager-develop is fixed at medium on Codex. manager-lead declares neither,
// so both keys stay omitted (inherit the parent).
var realSetCodexModelEffort = map[string][2]string{
	"manager-spec":    {config.CodexModelSol, "medium"},
	"plan-auditor":    {config.CodexModelSol, "high"},
	"sync-auditor":    {config.CodexModelSol, "high"},
	"super-advisor":   {config.CodexModelSol, "high"},
	"manager-develop": {config.CodexModelSol, "medium"},
	"manager-docs":    {config.CodexModelSol, "medium"},
	"e2e-tester":      {config.CodexModelSol, "medium"},
	"builder-harness": {config.CodexModelSol, "medium"},
	"manager-design":  {config.CodexModelSol, "medium"},
	"manager-git":     {config.CodexModelLuna, "high"},
	"manager-todo":    {config.CodexModelLuna, "high"},
}

// TestRealSetCodexShape pins the AC-007 shape over the real 12 — exactly the 7
// inventory carriers declare mcp_servers — plus the card t1808 model/effort
// table, and that no role file carries sandbox_mode (Codex ignores the field
// for a spawned role, which inherits the parent session's sandbox).
func TestRealSetCodexShape(t *testing.T) {
	pub := emitRealSet(t)
	for path, data := range pub.CodexTOML {
		doc, err := decodeTOML(string(data))
		if err != nil {
			t.Fatalf("independent decode of %s failed: %v\n%s", path, err, data)
		}
		name, _ := doc["name"].(string)
		if name == "" {
			t.Fatalf("%s: no name", path)
		}

		// AC-007: MCP server mapping on exactly the 7 carriers — map shape
		// (the run-phase-measured form; the array form is rejected by codex).
		servers, hasMCP := doc["mcp_servers"].(map[string]any)
		if expectedMCPCarriers[name] && !hasMCP {
			t.Errorf("%s (%s): inventory carrier must declare mcp_servers", path, name)
		}
		if !expectedMCPCarriers[name] && hasMCP {
			t.Errorf("%s (%s): non-carrier must not declare mcp_servers", path, name)
		}
		if hasMCP {
			moai, ok := servers["moai"].(map[string]any)
			if !ok || moai["command"] != "moai" {
				t.Errorf("%s (%s): mcp_servers.moai must carry the server definition (command=moai)", path, name)
			}
		}

		// card t1808: model and model_reasoning_effort follow the role table;
		// a role absent from the table declares neither key.
		want, inTable := realSetCodexModelEffort[name]
		model, hasModel := doc["model"].(string)
		effort, hasEffort := doc["model_reasoning_effort"].(string)
		switch {
		case inTable && (model != want[0] || effort != want[1]):
			t.Errorf("%s (%s): model/effort = %q/%q, want %q/%q", path, name, model, effort, want[0], want[1])
		case !inTable && (hasModel || hasEffort):
			t.Errorf("%s (%s): model/effort = %q/%q, want both keys omitted (no frontmatter values)", path, name, model, effort)
		}

		// sandbox_mode is not emitted: Codex does not apply it to a spawned
		// role (measured, codex-cli 0.156.1 — permission contract sandbox row).
		if v, has := doc["sandbox_mode"]; has {
			t.Errorf("%s (%s): sandbox_mode = %v, want the key omitted", path, name, v)
		}

		// R-005: developer_instructions decodes non-empty (byte-equality to
		// the .md body is asserted via the round-trip against the source).
		if body, _ := doc["developer_instructions"].(string); body == "" {
			t.Errorf("%s (%s): empty developer_instructions", path, name)
		}
	}
}

// TestRealSetBodiesByteEqual verifies AC-003/R-005 against the REAL sources:
// every emitted developer_instructions decodes byte-equal to the Codex-view
// body of its agent source, and name equals the frontmatter name, 12 of 12.
func TestRealSetBodiesByteEqual(t *testing.T) {
	pub := emitRealSet(t)
	man, err := agentemit.LoadManifest()
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	for path, data := range pub.CodexTOML {
		doc, err := decodeTOML(string(data))
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		name, _ := doc["name"].(string)
		srcPath := filepath.Join(repoRoot, filepath.FromSlash(sourceRoot), name+".md")
		raw, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("read source %s: %v", srcPath, err)
		}
		view, err := agentemit.RenderView(name+".md", raw, agentemit.ViewCodex)
		if err != nil {
			t.Fatalf("render Codex view of %s: %v", srcPath, err)
		}
		parsed, err := agentemit.ParseAgentDoc(name+".md", view)
		if err != nil {
			t.Fatalf("parse source %s: %v", srcPath, err)
		}
		if parsed.Name != name {
			t.Errorf("%s: TOML name %q != frontmatter name %q", path, name, parsed.Name)
		}
		// The body is carried verbatim; a role with a Codex-only addendum
		// gets exactly that addendum after it, and nothing else changes.
		got, _ := doc["developer_instructions"].(string)
		want := string(parsed.Body)
		if add, ok := man.CodexRoleAddenda[name]; ok {
			want += "\n" + strings.TrimRight(add, "\n") + "\n"
		}
		if !strings.HasPrefix(got, string(parsed.Body)) {
			t.Errorf("%s: developer_instructions does not start with the verbatim .md body", path)
		}
		if got != want {
			t.Errorf("%s: developer_instructions is not the .md body plus its manifest addendum", path)
		}
	}
}

// TestEmbedFSPresenceAndByteEquality is AC-010's embed half: the embedded
// template FS (all:templates — dot-dirs included) exposes all 12 .codex TOML
// paths byte-equal to the committed sources.
func TestEmbedFSPresenceAndByteEquality(t *testing.T) {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(templatesDir, ".codex/agents/moai"))
	if err != nil {
		t.Fatalf("committed .codex tree missing (%v) — regenerate first", err)
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		count++
		rel := ".codex/agents/moai/" + e.Name()
		embedData, err := fs.ReadFile(embedded, rel)
		if err != nil {
			t.Errorf("%s: missing from embedded FS: %v", rel, err)
			continue
		}
		committed, err := os.ReadFile(filepath.Join(templatesDir, rel))
		if err != nil {
			t.Fatalf("read committed %s: %v", rel, err)
		}
		if string(embedData) != string(committed) {
			t.Errorf("%s: embedded bytes differ from committed (run make build)", rel)
		}
	}
	if count != 12 {
		t.Errorf("committed .codex/agents/moai carries %d TOMLs, want 12", count)
	}
}

// neutralityPattern is the SPEC AC-011 leak pattern (broad form): SPEC-IDs,
// ISO dates, and sha-like hex runs.
var neutralityPattern = regexp.MustCompile(`SPEC-[A-Z][A-Z0-9-]*|20[0-9]{2}-[0-9]{2}-[0-9]{2}|[0-9a-f]{9}`)

// TestNeutralityByInheritance is AC-011's emitter-side face. The committed
// .md sources carry legitimate pedagogical tokens (SPEC-XXX placeholders,
// regex-walkthrough examples) that ride verbatim into the TOML bodies —
// R-005 forbids altering them. The neutrality obligation this SPEC can
// enforce is that the EMITTER introduces no new token: every pattern match
// in an emitted TOML must already occur in its .md source. (The CI
// neutrality workflows do not scan .toml — verified: both guards' extension
// sets exclude it — so this test is the .toml-side guard.)
func TestNeutralityByInheritance(t *testing.T) {
	pub := emitRealSet(t)
	for path, data := range pub.CodexTOML {
		doc, err := decodeTOML(string(data))
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		name, _ := doc["name"].(string)
		raw, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(sourceRoot), name+".md"))
		if err != nil {
			t.Fatalf("read source for %s: %v", path, err)
		}
		md := string(raw)
		for _, match := range neutralityPattern.FindAllString(string(data), -1) {
			if !strings.Contains(md, match) {
				t.Errorf("%s: emitted content introduces new neutrality-pattern token %q not present in its .md source (emitter leak)", path, match)
			}
		}
	}
}

// TestPublicationPathHygiene pins the emission's blast radius: the codex
// publication contains only .toml paths under .codex/ and the markdown
// publication only the .md files of the agent source directory.
func TestPublicationPathHygiene(t *testing.T) {
	pub := emitRealSet(t)
	for p := range pub.CodexTOML {
		if !strings.HasPrefix(p, ".codex/agents/") || !strings.HasSuffix(p, ".toml") {
			t.Errorf("codex publication path outside .codex/agents/*.toml: %s", p)
		}
	}
	for p := range pub.Markdown {
		if !strings.HasPrefix(p, sourceRoot+"/") || !strings.HasSuffix(p, ".md") {
			t.Errorf("markdown publication path outside the agent source root: %s", p)
		}
	}
}
