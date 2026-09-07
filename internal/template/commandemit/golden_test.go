// golden_test.go — SPEC-CODEX-COMMAND-SKILLS-001 golden guards.
//
// These tests run the emitter against the REAL 16 template command sources
// and pin the committed artifacts under templates/.agents/skills/. They are
// the drift guard: a hand-edited SKILL.md, a command-source edit, or a
// behavior change in the emitter fails here until regenerated via:
//
//	COMMAND_EMIT_UPDATE=1 go test ./internal/template/commandemit/... -run TestGoldenCommittedArtifactsMatchEmission
//
// (the `make commands-emit` target wraps this). The command sources are
// never written by anything in this package — the regression ban is by
// construction, and these tests make it observable.
package commandemit_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/commandemit"
)

// templatesDir is the template tree root relative to this package's dir.
const templatesDir = "../templates"

// expectedCommandCount is the published-set size: one per command source
// (AC-001).
const expectedCommandCount = 16

// emitRealSet runs the emitter over the committed template tree.
func emitRealSet(t *testing.T) *commandemit.Publication {
	t.Helper()
	fsys := os.DirFS(templatesDir)
	pub, err := commandemit.Emit(fsys, commandemit.DefaultOptions())
	if err != nil {
		t.Fatalf("Emit over real command set: %v", err)
	}
	if len(pub.Skills) != expectedCommandCount {
		t.Fatalf("emitted %d skills, want %d", len(pub.Skills), expectedCommandCount)
	}
	return pub
}

// committedSkillPath resolves the committed artifact path for one emitted path.
func committedSkillPath(emitted string) string {
	return filepath.Join(templatesDir, filepath.FromSlash(emitted))
}

// TestGoldenCommittedArtifactsMatchEmission is the AC-001/AC-011 drift
// guard: every emitted SKILL.md must be byte-identical (sha256) to the
// committed artifact. With COMMAND_EMIT_UPDATE=1 it (re)writes the committed
// artifacts instead (the maintainer regeneration path).
func TestGoldenCommittedArtifactsMatchEmission(t *testing.T) {
	pub := emitRealSet(t)
	update := os.Getenv(commandemit.EnvUpdate) == "1"

	emittedPaths := make([]string, 0, len(pub.Skills))
	for p := range pub.Skills {
		emittedPaths = append(emittedPaths, p)
	}
	sort.Strings(emittedPaths)

	for _, p := range emittedPaths {
		data := pub.Skills[p]
		sum := fmt.Sprintf("%x", sha256.Sum256(data))
		if update {
			if err := os.MkdirAll(filepath.Dir(committedSkillPath(p)), 0o755); err != nil {
				t.Fatalf("update mkdir: %v", err)
			}
			if err := os.WriteFile(committedSkillPath(p), data, 0o644); err != nil {
				t.Fatalf("update write %s: %v", p, err)
			}
			t.Logf("updated %s (sha256 %s)", p, sum[:12])
			continue
		}
		committed, err := os.ReadFile(committedSkillPath(p))
		if err != nil {
			t.Errorf("%s: committed artifact missing (%v) — regenerate with COMMAND_EMIT_UPDATE=1", p, err)
			continue
		}
		if fmt.Sprintf("%x", sha256.Sum256(committed)) != sum {
			t.Errorf("%s: committed artifact differs from emission (sha256 mismatch) — regenerate via `make commands-emit` or stop hand-editing", p)
		}
	}
}

// TestCommandSourcesUnmodified is AC-002: running the emitter leaves the
// command source tree byte-identical. The emitter holds no command-source
// write path; this test makes that observable rather than assumed.
func TestCommandSourcesUnmodified(t *testing.T) {
	before := hashCommandTree(t)
	emitRealSet(t) // full success-path emission over the real tree
	after := hashCommandTree(t)
	for name, sum := range before {
		if after[name] != sum {
			t.Errorf("%s: command source modified by emission (sha256 %s -> %s)", name, sum, after[name])
		}
	}
	if len(after) != len(before) {
		t.Errorf("command source file count changed: %d -> %d", len(before), len(after))
	}
}

func hashCommandTree(t *testing.T) map[string]string {
	t.Helper()
	root := filepath.Join(templatesDir, ".claude", "commands", "moai")
	out := map[string]string{}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read command tree: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || (!strings.HasSuffix(e.Name(), ".md") && !strings.HasSuffix(e.Name(), ".md.tmpl")) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		out[e.Name()] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	if len(out) != expectedCommandCount {
		t.Fatalf("expected %d command sources, found %d", expectedCommandCount, len(out))
	}
	return out
}

// TestPublishedIdentitySet is AC-003: the published directory set is exactly
// the 16 command names with the moai- prefix, and every frontmatter name
// equals its directory-derived identity.
func TestPublishedIdentitySet(t *testing.T) {
	pub := emitRealSet(t)

	want := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(templatesDir, ".claude", "commands", "moai"))
	if err != nil {
		t.Fatalf("read command sources: %v", err)
	}
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md.tmpl")
		name = strings.TrimSuffix(name, ".md")
		want["moai-"+name] = true
	}
	if len(want) != expectedCommandCount {
		t.Fatalf("derived %d identities from sources, want %d", len(want), expectedCommandCount)
	}

	got := map[string]bool{}
	for p, data := range pub.Skills {
		dir := strings.TrimSuffix(strings.TrimPrefix(p, ".agents/skills/"), "/SKILL.md")
		got[dir] = true
		if !strings.Contains(string(data), "name: "+dir+"\n") {
			t.Errorf("%s: frontmatter name does not equal derived identity %q", p, dir)
		}
	}
	for name := range want {
		if !got[name] {
			t.Errorf("published set missing derived identity %q", name)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("published set carries unexpected identity %q", name)
		}
	}
}

// TestDescriptionsEnglishNoTemplateSyntax is AC-004's content face: every
// published description equals the emitter's English extraction for its
// source, and no emitted file carries Go template syntax (AC-004's grep is
// a drift net over the same property).
func TestDescriptionsEnglishNoTemplateSyntax(t *testing.T) {
	pub := emitRealSet(t)
	for _, rep := range pub.Report {
		data, ok := pub.Skills[rep.Path]
		if !ok {
			t.Fatalf("report path %s not in publication", rep.Path)
		}
		if !strings.Contains(string(data), "description: \""+escapeForGrep(rep.Description)+"\"") {
			t.Errorf("%s: published description does not equal the English extraction %q", rep.Path, rep.Description)
		}
	}
	for p, data := range pub.Skills {
		if strings.Contains(string(data), "{{") {
			t.Errorf("%s: emitted artifact carries Go template syntax", p)
		}
	}
}

// escapeForGrep re-renders a description the way quoteYAMLScalar escapes it,
// so the containment check reads the emitted form.
func escapeForGrep(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	return strings.ReplaceAll(v, `"`, `\"`)
}

// TestBodiesByteEqual is AC-006: for every published skill, the body bytes
// after the frontmatter delimiter equal the source body bytes, 16 of 16.
func TestBodiesByteEqual(t *testing.T) {
	pub := emitRealSet(t)
	for _, rep := range pub.Report {
		emitted, err := os.ReadFile(committedSkillPath(rep.Path))
		if err != nil {
			t.Fatalf("read committed %s: %v", rep.Path, err)
		}
		emittedBody := bodyAfterFrontmatter(t, rep.Path, emitted)
		srcPath := filepath.Join(templatesDir, ".claude", "commands", "moai", sourceName(rep.Command))
		src, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("read source %s: %v", srcPath, err)
		}
		srcBody := bodyAfterFrontmatter(t, rep.Command, src)
		if string(emittedBody) != string(srcBody) {
			t.Errorf("%s: published body not byte-identical to source body", rep.Path)
		}
	}
}

// sourceName maps a command name back to its source file name.
func sourceName(command string) string {
	// The sources are <name>.md.tmpl except todo.md; try both shapes.
	tmpl := command + ".md.tmpl"
	if _, err := os.Stat(filepath.Join(templatesDir, ".claude", "commands", "moai", tmpl)); err == nil {
		return tmpl
	}
	return command + ".md"
}

// bodyAfterFrontmatter returns the bytes after the closing "---" delimiter
// line of a SKILL.md-shaped document.
func bodyAfterFrontmatter(t *testing.T, label string, data []byte) []byte {
	t.Helper()
	s := string(data)
	const delim = "---\n"
	if !strings.HasPrefix(s, delim) {
		t.Fatalf("%s: no opening delimiter", label)
	}
	idx := strings.Index(s[len(delim):], "\n---\n")
	if idx < 0 {
		t.Fatalf("%s: no closing delimiter", label)
	}
	return []byte(s[len(delim)+idx+len("\n---\n"):])
}

// TestBoundaryFlagsRecorded is AC-007's emitter half: every report entry
// records the boundary flag, and every published body still carries the
// verbatim Claude-only dispatcher line (proving no repair happened).
func TestBoundaryFlagsRecorded(t *testing.T) {
	pub := emitRealSet(t)
	if len(pub.Report) != expectedCommandCount {
		t.Fatalf("report carries %d entries, want %d", len(pub.Report), expectedCommandCount)
	}
	for _, rep := range pub.Report {
		if rep.BoundaryFlag == "" {
			t.Errorf("%s: report entry missing boundary flag", rep.Path)
		}
		data, ok := pub.Skills[rep.Path]
		if !ok {
			t.Fatalf("report path %s not in publication", rep.Path)
		}
		if !strings.Contains(string(data), `Use Skill("moai")`) {
			t.Errorf("%s: published body lost the verbatim Claude-only dispatcher line", rep.Path)
		}
	}
}

// TestPublicationPathHygiene pins the regeneration path's blast radius: the
// publication contains only SKILL.md paths under .agents/skills/ with the
// moai- prefix — update mode therefore has no command-source write target.
func TestPublicationPathHygiene(t *testing.T) {
	pub := emitRealSet(t)
	for p := range pub.Skills {
		if !strings.HasPrefix(p, ".agents/skills/moai-") || !strings.HasSuffix(p, "/SKILL.md") {
			t.Errorf("publication path outside .agents/skills/moai-*/SKILL.md: %s", p)
		}
	}
}

// TestGitignoreCarriesEveryPublishedName guards the distributed
// templates/.gitignore: the mirror rule `.agents/skills/moai*` would ignore
// the published skills too, so each published entry must carry an explicit
// re-inclusion negation. A new command whose negation line was forgotten is
// a silently-untracked artifact — this makes that staleness visible.
func TestGitignoreCarriesEveryPublishedName(t *testing.T) {
	pub := emitRealSet(t)
	raw, err := os.ReadFile(filepath.Join(templatesDir, ".gitignore"))
	if err != nil {
		t.Fatalf("read templates/.gitignore: %v", err)
	}
	for _, rep := range pub.Report {
		want := "!.agents/skills/" + rep.Skill + "/"
		if !strings.Contains(string(raw), want) {
			t.Errorf("templates/.gitignore missing re-inclusion %q — the published skill would be silently ignored", want)
		}
	}
}
