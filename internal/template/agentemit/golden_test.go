// golden_test.go — SPEC-CODEX-DUAL-AGENTS-001 MS3 golden guards.
//
// These tests run the emitter against the REAL 11 template .md sources and
// pin the committed artifacts under templates/.codex/agents/moai/. They are
// the drift guard: a hand-edited .toml or a behavior change in the emitter
// or manifest fails here until regenerated via:
//
//	AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/...
//
// (the `make agents-emit` target wraps this). The .md side is never written
// by anything in this package — the regression ban is by construction, and
// these tests make it observable.
package agentemit_test

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// templatesDir is the template tree root relative to this package's dir.
const templatesDir = "../templates"

// agentMDRoot is the neutral-layer source root inside the template tree.
const agentMDRoot = ".claude/agents/moai"

// expectedMCPCarriers is the AC-007 inventory: exactly these 7 agents carry
// mcp__moai__* tokens (plan.md §A.2, verified against the template tree).
var expectedMCPCarriers = map[string]bool{
	"manager-develop": true, "manager-docs": true, "manager-lead": true,
	"manager-spec": true, "plan-auditor": true, "super-advisor": true,
	"sync-auditor": true,
}

// expectedEffort is the AC-008 inventory from the template frontmatter.
var expectedEffort = map[string]string{
	"builder-harness": "medium", "e2e-tester": "low", "manager-design": "medium",
	"manager-develop": "high", "manager-docs": "low", "manager-git": "low",
	"manager-lead": "xhigh", "manager-spec": "high", "plan-auditor": "high",
	"super-advisor": "high", "sync-auditor": "high",
}

// emitRealSet runs the emitter over the committed template tree.
func emitRealSet(t *testing.T) *agentemit.Publication {
	t.Helper()
	man, err := agentemit.LoadManifest()
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	fsys := os.DirFS(templatesDir)
	pub, err := agentemit.EmitAll(fsys, agentMDRoot, man)
	if err != nil {
		t.Fatalf("EmitAll over real template set: %v", err)
	}
	if len(pub.CodexTOML) != 11 {
		t.Fatalf("emitted %d TOMLs, want 11", len(pub.CodexTOML))
	}
	return pub
}

// committedTOMLPath resolves the committed artifact path for one emitted path.
func committedTOMLPath(emitted string) string {
	return filepath.Join(templatesDir, filepath.FromSlash(emitted))
}

// TestGoldenCommittedArtifactsMatchEmission is the AC-001/AC-004 drift guard:
// every emitted TOML must be byte-identical (sha256) to the committed
// artifact, and every published .md byte-identical to the committed .md.
// With AGENTEMIT_UPDATE=1 it (re)writes the committed .toml artifacts instead
// (the maintainer regeneration path).
func TestGoldenCommittedArtifactsMatchEmission(t *testing.T) {
	pub := emitRealSet(t)
	update := os.Getenv("AGENTEMIT_UPDATE") == "1"

	emittedPaths := make([]string, 0, len(pub.CodexTOML))
	for p := range pub.CodexTOML {
		emittedPaths = append(emittedPaths, p)
	}
	sort.Strings(emittedPaths)

	for _, p := range emittedPaths {
		data := pub.CodexTOML[p]
		sum := fmt.Sprintf("%x", sha256.Sum256(data))
		if update {
			if err := os.MkdirAll(filepath.Dir(committedTOMLPath(p)), 0o755); err != nil {
				t.Fatalf("update mkdir: %v", err)
			}
			if err := os.WriteFile(committedTOMLPath(p), data, 0o644); err != nil {
				t.Fatalf("update write %s: %v", p, err)
			}
			t.Logf("updated %s (sha256 %s)", p, sum[:12])
			continue
		}
		committed, err := os.ReadFile(committedTOMLPath(p))
		if err != nil {
			t.Errorf("%s: committed artifact missing (%v) — regenerate with AGENTEMIT_UPDATE=1", p, err)
			continue
		}
		if fmt.Sprintf("%x", sha256.Sum256(committed)) != sum {
			t.Errorf("%s: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing", p)
		}
	}

	// AC-001's .md face: the markdown publication is byte-identical to the
	// committed .md files (identity pass-through, never re-rendered).
	for p, data := range pub.Markdown {
		committed, err := os.ReadFile(filepath.Join(templatesDir, filepath.FromSlash(p)))
		if err != nil {
			t.Fatalf("%s: committed .md unreadable: %v", p, err)
		}
		if string(committed) != string(data) {
			t.Errorf("%s: markdown publication differs from committed .md", p)
		}
	}
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
	if len(out) != 11 {
		t.Fatalf("expected 11 .md sources, found %d", len(out))
	}
	return out
}

// TestRealSetCodexShape pins the AC-007/AC-008/AC-009 (+ sandbox) shape over
// the real 11: exactly the 7 inventory carriers declare mcp_servers, every
// agent carries its manifest-mapped model_reasoning_effort, zero carry a
// model key, and all carry the P-01-confirmed sandbox_mode.
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

		// AC-008: effort mapping per manifest (identity, P-02-locked).
		if got, _ := doc["model_reasoning_effort"].(string); got != expectedEffort[name] {
			t.Errorf("%s (%s): model_reasoning_effort = %q, want %q", path, name, got, expectedEffort[name])
		}

		// AC-009: model omitted everywhere (manager-git sonnet = documented drop).
		if _, has := doc["model"]; has {
			t.Errorf("%s (%s): model key must be omitted", path, name)
		}

		// P-01: sandbox_mode = workspace-write everywhere.
		if got, _ := doc["sandbox_mode"].(string); got != "workspace-write" {
			t.Errorf("%s (%s): sandbox_mode = %q, want workspace-write", path, name, got)
		}

		// R-005: developer_instructions decodes non-empty (byte-equality to
		// the .md body is asserted via the round-trip against the source).
		if body, _ := doc["developer_instructions"].(string); body == "" {
			t.Errorf("%s (%s): empty developer_instructions", path, name)
		}
	}
}

// TestRealSetBodiesByteEqual verifies AC-003/R-005 against the REAL sources:
// every emitted developer_instructions decodes byte-equal to the .md body of
// its agent, and name equals the frontmatter name, 11 of 11.
func TestRealSetBodiesByteEqual(t *testing.T) {
	pub := emitRealSet(t)
	for path, data := range pub.CodexTOML {
		doc, err := decodeTOML(string(data))
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		name, _ := doc["name"].(string)
		mdPath := filepath.Join(templatesDir, filepath.FromSlash(agentMDRoot), name+".md")
		raw, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("read source %s: %v", mdPath, err)
		}
		parsed, err := agentemit.ParseAgentDoc(name+".md", raw)
		if err != nil {
			t.Fatalf("parse source %s: %v", mdPath, err)
		}
		if parsed.Name != name {
			t.Errorf("%s: TOML name %q != frontmatter name %q", path, name, parsed.Name)
		}
		if got, _ := doc["developer_instructions"].(string); got != string(parsed.Body) {
			t.Errorf("%s: developer_instructions not byte-equal to .md body", path)
		}
	}
}

// TestEmbedFSPresenceAndByteEquality is AC-010's embed half: the embedded
// template FS (all:templates — dot-dirs included) exposes all 11 .codex TOML
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
	if count != 11 {
		t.Errorf("committed .codex/agents/moai carries %d TOMLs, want 11", count)
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
		raw, err := os.ReadFile(filepath.Join(templatesDir, filepath.FromSlash(agentMDRoot), name+".md"))
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

// TestPublicationPathHygiene pins the regeneration path's blast radius: the
// codex publication contains only .toml paths under .codex/ and the markdown
// publication only the .md source paths — update mode therefore has no .md
// write target (AC-002 for the one write surface this package owns).
func TestPublicationPathHygiene(t *testing.T) {
	pub := emitRealSet(t)
	for p := range pub.CodexTOML {
		if !strings.HasPrefix(p, ".codex/agents/") || !strings.HasSuffix(p, ".toml") {
			t.Errorf("codex publication path outside .codex/agents/*.toml: %s", p)
		}
	}
	for p := range pub.Markdown {
		if !strings.HasPrefix(p, agentMDRoot+"/") || !strings.HasSuffix(p, ".md") {
			t.Errorf("markdown publication path outside the .md source root: %s", p)
		}
	}
}
