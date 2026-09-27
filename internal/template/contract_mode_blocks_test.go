// contract_mode_blocks_test.go — structural guards for the contract-mode
// marker blocks and the contract-autonomy SSOT rule.
//
// A contract-mode block is an additive instruction that applies only when
// `workflow.autonomy.mode: contract`. It sits between two whole-line HTML
// comment markers:
//
//	<!-- moai:contract-mode-start id="<slug>" -->
//	Where `workflow.autonomy.mode: contract` — <instruction>.
//	<!-- moai:contract-mode-end -->
//
// The guards here need no base ref: they read the working tree (local copies
// under .claude/ and CLAUDE.md, template copies under
// internal/template/templates/) and check pairing, nesting, evolvable-zone
// placement, forbidden internal-content classes, block size, local↔template
// parity, the SSOT section layout, and the content of named blocks.
//
// Every checker returns its findings instead of failing the test directly,
// so the falsifier subtests can feed it a known-bad fixture and observe a
// non-empty finding list (the RED half of each guard).
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// grTemplatePrefix is the template mirror root, relative to the repository.
const grTemplatePrefix = "internal/template/templates/"

// grSSOTPath is the contract-autonomy SSOT rule (local path).
const grSSOTPath = ".claude/rules/moai/workflow/contract-autonomy.md"

// grTarget is one edited document and the block ids it must carry.
type grTarget struct {
	path         string   // local path, repository-relative
	ids          []string // block ids the document must carry
	alwaysLoaded bool     // loaded into every session (500-char block cap)
	emitter      bool     // a Kickoff emitter site (research §1.2 class E)
}

// grTargets are the block-bearing documents (the SSOT is handled apart).
var grTargets = []grTarget{
	{path: "CLAUDE.md", ids: []string{"contract-signing-pipeline", "contract-safe-dev"}, alwaysLoaded: true, emitter: true},
	{path: ".claude/rules/moai/core/askuser-protocol.md", ids: []string{"contract-ambiguity"}, alwaysLoaded: true, emitter: true},
	{path: ".claude/rules/moai/workflow/goal-directive.md", ids: []string{"contract-signing-goal"}, alwaysLoaded: true, emitter: true},
	{path: ".claude/rules/moai/workflow/orchestration-mode-selection.md", ids: []string{"contract-signing"}, emitter: true},
	{path: ".claude/skills/moai/SKILL.md", ids: []string{"contract-signing-router"}, emitter: true},
	{path: ".claude/skills/moai/workflows/moai.md", ids: []string{"contract-pipeline-gates", "contract-merged-round"}, emitter: true},
	{path: ".claude/skills/moai/workflows/plan.md", ids: []string{"contract-clarification"}, emitter: true},
	{path: ".claude/skills/moai/workflows/plan/spec-assembly.md", ids: []string{"contract-draft", "contract-signing-review", "contract-audit-retry", "contract-quality-gate"}, emitter: true},
	{path: ".claude/skills/moai/workflows/run.md", ids: []string{"contract-signing-run", "contract-lifecycle-run"}, emitter: true},
	{path: ".claude/skills/moai/workflows/goal.md", ids: []string{"contract-progression"}, emitter: true},
	{path: ".claude/skills/moai/workflows/sync.md", ids: []string{"contract-sync-gates"}},
	{path: ".claude/skills/moai/workflows/sync/doc-execution.md", ids: []string{"contract-doc-scope"}},
	{path: ".claude/skills/moai/workflows/sync/delivery.md", ids: []string{"contract-next-steps", "contract-error-flow"}},
}

// Block size caps in characters (runes) of the block body between markers.
const (
	grAlwaysLoadedBlockCap = 500
	grOtherBlockCap        = 900
)

var (
	grStartRe = regexp.MustCompile(`^<!-- moai:contract-mode-start id="([^"]*)" -->$`)
	grEndRe   = regexp.MustCompile(`^<!-- moai:contract-mode-end -->$`)
	grSlugRe  = regexp.MustCompile(`^[a-z0-9-]+$`)
	// grCondition is the mandatory first non-empty line prefix of a block.
	grCondition = "Where `workflow.autonomy.mode: contract`"
)

// grForbidden are the internal-content classes a block must not carry.
var grForbidden = []struct {
	name string
	re   *regexp.Regexp
}{
	{"spec-id", regexp.MustCompile(`\bSPEC-[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*-[0-9]{3}\b`)},
	{"req-ac-token", regexp.MustCompile(`\b(?:REQ|AC)-[A-Z0-9]+(?:-[A-Z0-9]+)*-[0-9]+\b`)},
	{"card-id", regexp.MustCompile(`\bt[0-9]{3,5}\b`)},
	{"date", regexp.MustCompile(`\b20[0-9]{2}-[01][0-9]-[0-3][0-9]\b`)},
	{"gate-number", regexp.MustCompile(`\bG[0-9]{1,2}\b`)},
	{"commit-sha", regexp.MustCompile(`\b[0-9a-f]{7,40}\b`)},
}

// grBlock is one parsed contract-mode block.
type grBlock struct {
	id         string
	startLine  int // 1-based line of the start marker
	endLine    int // 1-based line of the end marker
	text       string
	body       string // lines strictly between the markers
	firstLine  string // first non-empty body line
	evolvable  bool   // the block lies inside an evolvable zone
	nestedFail bool
}

// grParse splits text into blocks and reports structural findings: an end
// marker without a start, a start inside a block (nesting), a start without
// an end, a malformed or duplicate slug, and a block inside an evolvable zone.
func grParse(name, text string) ([]grBlock, []string) {
	var (
		blocks   []grBlock
		findings []string
		cur      *grBlock
		inEvo    bool
		body     []string
		seen     = map[string]bool{}
	)
	lines := strings.Split(text, "\n")
	for i, raw := range lines {
		ln := i + 1
		line := strings.TrimRight(raw, "\r")
		switch {
		case strings.HasPrefix(line, "<!-- moai:evolvable-start"):
			inEvo = true
		case strings.HasPrefix(line, "<!-- moai:evolvable-end"):
			inEvo = false
		}
		if m := grStartRe.FindStringSubmatch(line); m != nil {
			if cur != nil {
				findings = append(findings, fmt.Sprintf("%s:%d: nested start marker inside block %q", name, ln, cur.id))
				continue
			}
			id := m[1]
			if !grSlugRe.MatchString(id) {
				findings = append(findings, fmt.Sprintf("%s:%d: block id %q is not [a-z0-9-]+", name, ln, id))
			}
			if seen[id] {
				findings = append(findings, fmt.Sprintf("%s:%d: duplicate block id %q", name, ln, id))
			}
			seen[id] = true
			cur = &grBlock{id: id, startLine: ln, evolvable: inEvo}
			body = nil
			continue
		}
		if strings.Contains(line, "moai:contract-mode-start") {
			findings = append(findings, fmt.Sprintf("%s:%d: start marker does not occupy a whole line", name, ln))
		}
		if grEndRe.MatchString(line) {
			if cur == nil {
				findings = append(findings, fmt.Sprintf("%s:%d: end marker without a start marker", name, ln))
				continue
			}
			cur.endLine = ln
			cur.body = strings.Join(body, "\n")
			cur.text = strings.Join(lines[cur.startLine-1:ln], "\n")
			for _, b := range body {
				if strings.TrimSpace(b) != "" {
					cur.firstLine = strings.TrimSpace(b)
					break
				}
			}
			if cur.evolvable || inEvo {
				cur.evolvable = true
				findings = append(findings, fmt.Sprintf("%s:%d: block %q lies inside an evolvable zone", name, cur.startLine, cur.id))
			}
			blocks = append(blocks, *cur)
			cur = nil
			continue
		}
		if strings.Contains(line, "moai:contract-mode-end") {
			findings = append(findings, fmt.Sprintf("%s:%d: end marker does not occupy a whole line", name, ln))
		}
		if cur != nil {
			body = append(body, line)
		}
	}
	if cur != nil {
		findings = append(findings, fmt.Sprintf("%s:%d: start marker of block %q has no end marker", name, cur.startLine, cur.id))
	}
	return blocks, findings
}

// grStrip removes every block, marker lines included (design.md §1.1).
func grStrip(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	in := false
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		if !in && grStartRe.MatchString(line) {
			in = true
			continue
		}
		if in {
			if grEndRe.MatchString(line) {
				in = false
			}
			continue
		}
		out = append(out, raw)
	}
	return strings.Join(out, "\n")
}

// grCheckContent reports forbidden internal-content classes, a missing
// condition line, and an over-cap block body.
func grCheckContent(name string, blocks []grBlock, capRunes int) []string {
	var findings []string
	for _, b := range blocks {
		for _, f := range grForbidden {
			for _, m := range f.re.FindAllString(b.body, -1) {
				// A commit SHA mixes digits and hex letters; a run of one kind
				// only (a plain number, a word like "deadbeef") is not one.
				if f.name == "commit-sha" && (!strings.ContainsAny(m, "0123456789") || !strings.ContainsAny(m, "abcdef")) {
					continue
				}
				findings = append(findings, fmt.Sprintf("%s: block %q carries forbidden class %s (%q)", name, b.id, f.name, m))
				break
			}
		}
		if !strings.HasPrefix(b.firstLine, grCondition) {
			findings = append(findings, fmt.Sprintf("%s: block %q first line %q does not start with %s", name, b.id, b.firstLine, grCondition))
		}
		if n := utf8.RuneCountInString(b.body); n > capRunes {
			findings = append(findings, fmt.Sprintf("%s: block %q body is %d characters (cap %d)", name, b.id, n, capRunes))
		}
	}
	return findings
}

// grCheckDocument runs every structural and content check on one document.
func grCheckDocument(name, text string, capRunes int) ([]grBlock, []string) {
	blocks, findings := grParse(name, text)
	findings = append(findings, grCheckContent(name, blocks, capRunes)...)
	return blocks, findings
}

// grRoot returns the repository root (the directory holding go.mod).
func grRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

func grRead(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func grCapFor(alwaysLoaded bool) int {
	if alwaysLoaded {
		return grAlwaysLoadedBlockCap
	}
	return grOtherBlockCap
}

// grAlwaysLoaded reports whether a repository-relative path (either copy) is
// one of the always-loaded targets.
func grAlwaysLoaded(rel string) bool {
	rel = strings.TrimPrefix(filepath.ToSlash(rel), grTemplatePrefix)
	for _, tg := range grTargets {
		if tg.path == rel {
			return tg.alwaysLoaded
		}
	}
	return false
}

// grWellFormedTree walks the template tree and returns the block count and
// every finding. An empty sweep (zero blocks) is itself a finding.
func grWellFormedTree(root string) (int, []string, error) {
	count := 0
	var findings []string
	base := filepath.Join(root, filepath.FromSlash(grTemplatePrefix))
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		blocks, f := grCheckDocument(rel, string(data), grCapFor(grAlwaysLoaded(rel)))
		count += len(blocks)
		findings = append(findings, f...)
		return nil
	})
	if count == 0 {
		findings = append(findings, "template tree carries zero contract-mode blocks — the guard swept nothing")
	}
	return count, findings, err
}

func grBlockText(id, body string) string {
	return "<!-- moai:contract-mode-start id=\"" + id + "\" -->\n" + body + "\n<!-- moai:contract-mode-end -->"
}

// TestContractModeBlocksWellFormed is the block guard (AC-GR-008, AC-GR-009).
// Its falsifier subtests feed known-bad fixtures and require findings; the
// final subtest checks the real template tree and requires none.
func TestContractModeBlocksWellFormed(t *testing.T) {
	ok := "Where `workflow.autonomy.mode: contract` — sign instead. See the SSOT."
	falsifiers := []struct {
		name string
		text string
		cap  int
	}{
		{"unpaired-start", "intro\n<!-- moai:contract-mode-start id=\"a\" -->\n" + ok + "\n", 900},
		{"inside-evolvable", "<!-- moai:evolvable-start id=\"z\" -->\n" + grBlockText("a", ok) + "\n<!-- moai:evolvable-end -->\n", 900},
		{"forbidden-internal-token", grBlockText("a", ok+" Tracked by SPEC-AUTONOMY-GATE-REWIRE-001, card t1236 on 2026-09-26."), 900},
		{"over-cap", grBlockText("a", ok+" "+strings.Repeat("x", 600)), 500},
	}
	for _, f := range falsifiers {
		t.Run("falsifier/"+f.name, func(t *testing.T) {
			_, findings := grCheckDocument("fixture.md", f.text, f.cap)
			if len(findings) == 0 {
				t.Fatalf("guard accepted a known-bad fixture %q", f.name)
			}
			t.Logf("observed findings: %v", findings)
		})
	}
	t.Run("falsifier/zero-blocks", func(t *testing.T) {
		dir := t.TempDir()
		tpl := filepath.Join(dir, filepath.FromSlash(grTemplatePrefix))
		if err := os.MkdirAll(tpl, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tpl, "plain.md"), []byte("no blocks here\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		n, findings, err := grWellFormedTree(dir)
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 || len(findings) == 0 {
			t.Fatalf("empty sweep not reported: count=%d findings=%v", n, findings)
		}
	})
	t.Run("template-tree", func(t *testing.T) {
		root := grRoot(t)
		n, findings, err := grWellFormedTree(root)
		if err != nil {
			t.Fatalf("walk template tree: %v", err)
		}
		t.Logf("template tree: %d contract-mode blocks", n)
		for _, f := range findings {
			t.Error(f)
		}
	})
}

// TestContractModeLocalTemplateParity requires byte-identical blocks in the
// local and template copies, the expected block ids per document, and a
// byte-identical SSOT (AC-GR-002 half).
func TestContractModeLocalTemplateParity(t *testing.T) {
	root := grRoot(t)
	local := grRead(t, root, grSSOTPath)
	tmpl := grRead(t, root, grTemplatePrefix+grSSOTPath)
	if local != tmpl {
		t.Errorf("SSOT %s differs between the local and template copies", grSSOTPath)
	}
	for _, tg := range grTargets {
		lb, lf := grParse(tg.path, grRead(t, root, tg.path))
		tb, tf := grParse(grTemplatePrefix+tg.path, grRead(t, root, grTemplatePrefix+tg.path))
		for _, f := range append(lf, tf...) {
			t.Error(f)
		}
		lm, tm := map[string]string{}, map[string]string{}
		for _, b := range lb {
			lm[b.id] = b.text
		}
		for _, b := range tb {
			tm[b.id] = b.text
		}
		for _, id := range tg.ids {
			if _, ok := lm[id]; !ok {
				t.Errorf("%s: missing block %q", tg.path, id)
			}
		}
		if len(lm) != len(tm) {
			t.Errorf("%s: local carries %d blocks, template %d", tg.path, len(lm), len(tm))
		}
		for id, text := range lm {
			if tm[id] != text {
				t.Errorf("%s: block %q differs between the local and template copies", tg.path, id)
			}
		}
	}
}

// grSSOTSections are the SSOT's required section headings, in order.
var grSSOTSections = []string{
	"## Scope and activation",
	"## The signing gate",
	"## Equivalence clause (human signature only)",
	"## Gate disposition",
	"## Gates a contract never touches",
	"## Escalation routing",
	"## One-pass lifecycle",
	"## Guided mode",
	"## Autonomous Kickoff",
	"## Revocation",
}

// grStages are the seven lifecycle stages, in order.
var grStages = []string{"Discovery", "RED", "GREEN", "Qualification", "Closure", "Integration", "Push"}

// grSection returns the text of the `## ` section whose heading line equals
// heading, up to the next `## ` heading. ok is false when absent.
func grSection(text, heading string) (string, bool) {
	lines := strings.Split(text, "\n")
	start := -1
	for i, l := range lines {
		if strings.TrimRight(l, "\r") == heading {
			start = i
			break
		}
	}
	if start < 0 {
		return "", false
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n"), true
}

// grSSOTFindings checks the SSOT layout and the per-section tokens
// (design.md §5.1). It is a pure function so the falsifier can feed it a
// damaged copy.
func grSSOTFindings(text string) []string {
	var findings []string
	tokens := map[string][]string{
		"## The signing gate":                          {"kickoff-check", "AskUserQuestion", "interactive terminal", "reject", "human decision required"},
		"## Equivalence clause (human signature only)": {"signer_kind: human", "interactive-tty"},
		"## Gate disposition":                          {"Socratic", "approach", "assumption", "plan-audit", "quality gate", "gate-sync-2"},
		"## Gates a contract never touches":            {"question-channel", "sync-auditor", "/clear"},
		"## Escalation routing":                        {"Report-Before-Ask", "escalation/"},
		"## One-pass lifecycle":                        append(append([]string{}, grStages...), "verdict.md", "audit_multi", "mutation"),
		"## Revocation":                                {"moai contract revoke", "next stage boundary"},
	}
	prev := -1
	for _, h := range grSSOTSections {
		sec, ok := grSection(text, h)
		if !ok {
			findings = append(findings, "SSOT section missing: "+h)
			continue
		}
		idx := strings.Index(text, sec)
		if idx < prev {
			findings = append(findings, "SSOT section out of order: "+h)
		}
		prev = idx
		for _, tok := range tokens[h] {
			if !strings.Contains(sec, tok) {
				findings = append(findings, fmt.Sprintf("SSOT section %q lacks token %q", h, tok))
			}
		}
	}
	if eq, ok := grSection(text, "## Equivalence clause (human signature only)"); ok {
		for _, bad := range []string{"llm+jev", "signer_kind: llm"} {
			if strings.Contains(eq, bad) {
				findings = append(findings, fmt.Sprintf("Equivalence clause names a non-human signature (%q)", bad))
			}
		}
	}
	if ak, ok := grSection(text, "## Autonomous Kickoff"); ok {
		if strings.Contains(ak, `id="contract-autonomous-kickoff"`) {
			for _, tok := range []string{"decide", "llm+jev", "fallback", "jev_min_confidence", "author-decider-conflict"} {
				if !strings.Contains(ak, tok) {
					findings = append(findings, fmt.Sprintf("Autonomous Kickoff block lacks token %q", tok))
				}
			}
		} else if strings.Contains(ak, "llm+jev") {
			findings = append(findings, "Autonomous Kickoff section names llm+jev before its activation block exists")
		}
	}
	return findings
}

// TestContractModeSSOTSections checks the SSOT's section layout and the
// tokens each section must carry (AC-GR-005).
func TestContractModeSSOTSections(t *testing.T) {
	t.Run("falsifier/missing-section-and-token", func(t *testing.T) {
		bad := "## Scope and activation\n\n## The signing gate\nkickoff-check only\n"
		if f := grSSOTFindings(bad); len(f) == 0 {
			t.Fatal("SSOT checker accepted a damaged copy")
		} else {
			t.Logf("observed %d findings on the damaged copy", len(f))
		}
	})
	t.Run("ssot", func(t *testing.T) {
		root := grRoot(t)
		for _, f := range grSSOTFindings(grRead(t, root, grSSOTPath)) {
			t.Error(f)
		}
	})
}

// grBlocksOf returns the blocks of one document keyed by id.
func grBlocksOf(t *testing.T, root, rel string) map[string]grBlock {
	t.Helper()
	blocks, findings := grParse(rel, grRead(t, root, rel))
	for _, f := range findings {
		t.Error(f)
	}
	m := map[string]grBlock{}
	for _, b := range blocks {
		m[b.id] = b
	}
	return m
}

var grWordRe = map[string]*regexp.Regexp{}

func grWordIndex(text, word string) int {
	re, ok := grWordRe[word]
	if !ok {
		re = regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
		grWordRe[word] = re
	}
	loc := re.FindStringIndex(text)
	if loc == nil {
		return -1
	}
	return loc[0]
}

// grOrderFindings reports stages absent from text or out of order.
func grOrderFindings(label, text string, stages []string) []string {
	var findings []string
	prev := -1
	for _, s := range stages {
		i := grWordIndex(text, s)
		if i < 0 {
			findings = append(findings, fmt.Sprintf("%s: stage %s absent", label, s))
			continue
		}
		if i < prev {
			findings = append(findings, fmt.Sprintf("%s: stage %s out of order", label, s))
		}
		prev = i
	}
	return findings
}

// TestContractModeLifecycleOrder checks the stage order inside the run block,
// the sync block, and the SSOT lifecycle section only (AC-GR-006).
func TestContractModeLifecycleOrder(t *testing.T) {
	t.Run("falsifier/swapped", func(t *testing.T) {
		if f := grOrderFindings("fixture", "GREEN then RED then Discovery then Qualification", grStages[:4]); len(f) == 0 {
			t.Fatal("order checker accepted a swapped order")
		}
	})
	t.Run("tree", func(t *testing.T) {
		root := grRoot(t)
		run := grBlocksOf(t, root, ".claude/skills/moai/workflows/run.md")["contract-lifecycle-run"]
		sync := grBlocksOf(t, root, ".claude/skills/moai/workflows/sync.md")["contract-sync-gates"]
		sec, ok := grSection(grRead(t, root, grSSOTPath), "## One-pass lifecycle")
		if !ok {
			t.Fatal("SSOT lifecycle section missing")
		}
		var findings []string
		findings = append(findings, grOrderFindings("run block", run.body, grStages[:4])...)
		findings = append(findings, grOrderFindings("sync block", sync.body, grStages[4:])...)
		findings = append(findings, grOrderFindings("SSOT lifecycle", sec, grStages)...)
		for _, f := range findings {
			t.Error(f)
		}
	})
}

// grLifecycleRows parses the SSOT lifecycle table: stage → cells.
func grLifecycleRows(sec string) map[string][]string {
	rows := map[string][]string{}
	for _, l := range strings.Split(sec, "\n") {
		if !strings.HasPrefix(l, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(strings.TrimSpace(l), "|"), "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		if len(cells) < 4 {
			continue
		}
		for _, s := range grStages {
			if strings.Trim(cells[0], "*`") == s {
				rows[s] = cells
			}
		}
	}
	return rows
}

// grEvidenceFindings checks the evidence and advance columns (AC-GR-007).
func grEvidenceFindings(sec string) []string {
	want := map[string][]string{
		"Discovery":     {"reobserve"},
		"RED":           {"commit", "failing"},
		"GREEN":         {"passing"},
		"Qualification": {"lint", "coverage", "mutation", "audit_multi"},
		"Closure":       {"verdict.md"},
		"Integration":   {"merged tree"},
	}
	rows := grLifecycleRows(sec)
	var findings []string
	for _, s := range grStages {
		cells, ok := rows[s]
		if !ok {
			findings = append(findings, "lifecycle table row missing: "+s)
			continue
		}
		evidence, advance := cells[2], cells[len(cells)-1]
		for _, tok := range want[s] {
			if !strings.Contains(evidence, tok) {
				findings = append(findings, fmt.Sprintf("row %s evidence lacks %q", s, tok))
			}
		}
		if !strings.Contains(advance, "open escalation record") {
			findings = append(findings, fmt.Sprintf("row %s advance condition lacks %q", s, "open escalation record"))
		}
	}
	return findings
}

// TestContractModeLifecycleEvidence checks the per-stage evidence duty.
func TestContractModeLifecycleEvidence(t *testing.T) {
	t.Run("falsifier/missing-evidence", func(t *testing.T) {
		bad := "| Stage | Input | Evidence | Advance |\n|---|---|---|---|\n| Discovery | x | nothing | ok |\n"
		if f := grEvidenceFindings(bad); len(f) == 0 {
			t.Fatal("evidence checker accepted an empty table")
		}
	})
	t.Run("ssot", func(t *testing.T) {
		root := grRoot(t)
		sec, ok := grSection(grRead(t, root, grSSOTPath), "## One-pass lifecycle")
		if !ok {
			t.Fatal("SSOT lifecycle section missing")
		}
		for _, f := range grEvidenceFindings(sec) {
			t.Error(f)
		}
	})
}

// TestContractModeBlockCondition checks every block (both copies) opens with
// the contract-mode condition (AC-GR-011).
func TestContractModeBlockCondition(t *testing.T) {
	t.Run("falsifier/no-condition", func(t *testing.T) {
		blocks, _ := grParse("fixture.md", grBlockText("a", "Always sign first."))
		if f := grCheckContent("fixture.md", blocks, 900); len(f) == 0 {
			t.Fatal("condition check accepted a block without the condition line")
		}
	})
	t.Run("tree", func(t *testing.T) {
		root := grRoot(t)
		n := 0
		for _, tg := range grTargets {
			for _, rel := range []string{tg.path, grTemplatePrefix + tg.path} {
				for _, b := range grBlocksOf(t, root, rel) {
					n++
					if !strings.HasPrefix(b.firstLine, grCondition) {
						t.Errorf("%s: block %q first line %q", rel, b.id, b.firstLine)
					}
				}
			}
		}
		if n == 0 {
			t.Fatal("no blocks found — empty sweep")
		}
		t.Logf("checked %d blocks", n)
	})
}

// grRequireTokens reports tokens absent from text and forbidden ones present.
func grRequireTokens(label, text string, need, forbid []string) []string {
	var findings []string
	for _, tok := range need {
		if !strings.Contains(text, tok) {
			findings = append(findings, fmt.Sprintf("%s lacks %q", label, tok))
		}
	}
	for _, tok := range forbid {
		if strings.Contains(text, tok) {
			findings = append(findings, fmt.Sprintf("%s carries forbidden %q", label, tok))
		}
	}
	return findings
}

// TestContractModeAuditRetryBlocks checks the plan-audit retry blocks
// (AC-GR-012).
func TestContractModeAuditRetryBlocks(t *testing.T) {
	need, forbid := []string{"audit_retries", "budget_default"}, []string{"AskUserQuestion"}
	t.Run("falsifier/asks", func(t *testing.T) {
		if f := grRequireTokens("fixture", "audit_retries budget_default then AskUserQuestion", need, forbid); len(f) == 0 {
			t.Fatal("audit-retry check accepted a block that asks")
		}
	})
	t.Run("tree", func(t *testing.T) {
		root := grRoot(t)
		m := grBlocksOf(t, root, ".claude/skills/moai/workflows/plan/spec-assembly.md")
		for _, id := range []string{"contract-audit-retry", "contract-quality-gate"} {
			b, ok := m[id]
			if !ok {
				t.Errorf("block %q missing", id)
				continue
			}
			for _, f := range grRequireTokens(id, b.body, need, forbid) {
				t.Error(f)
			}
		}
	})
}

// TestContractModeSyncBlocks checks the sync blocks name the removed
// questions and the failure decision points, route to escalation, and
// neither ask nor touch the CI auto-fix loop (AC-GR-013).
func TestContractModeSyncBlocks(t *testing.T) {
	need := []string{"gate-sync-2", "Phase 1", "Phase 3", "Phase 6", "Phase 7", "Phase 8", "Phase 13", "escalat"}
	forbid := []string{"AskUserQuestion", "ci-autofix"}
	t.Run("falsifier/incomplete", func(t *testing.T) {
		if f := grRequireTokens("fixture", "gate-sync-2 only", need, forbid); len(f) == 0 {
			t.Fatal("sync check accepted an incomplete block set")
		}
	})
	t.Run("tree", func(t *testing.T) {
		root := grRoot(t)
		var all strings.Builder
		for _, rel := range []string{
			".claude/skills/moai/workflows/sync.md",
			".claude/skills/moai/workflows/sync/doc-execution.md",
			".claude/skills/moai/workflows/sync/delivery.md",
		} {
			bs := grBlocksOf(t, root, rel)
			if len(bs) == 0 {
				t.Errorf("%s carries no block", rel)
			}
			for _, b := range bs {
				all.WriteString(b.body + "\n")
			}
		}
		for _, f := range grRequireTokens("sync blocks", all.String(), need, forbid) {
			t.Error(f)
		}
	})
}

// grSentenceWith reports whether one sentence of text carries every word.
func grSentenceWith(text string, words ...string) bool {
	for _, s := range regexp.MustCompile(`[.!?]\s`).Split(text, -1) {
		all := true
		for _, w := range words {
			if grWordIndex(s, w) < 0 {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

// TestContractModeSigningBlocks checks the signing-gate and draft blocks
// (AC-GR-014).
func TestContractModeSigningBlocks(t *testing.T) {
	t.Run("falsifier/split-sentence", func(t *testing.T) {
		if grSentenceWith("An outcome of reject stops. A human signs.", "reject", "human") {
			t.Fatal("sentence check joined two sentences")
		}
	})
	t.Run("tree", func(t *testing.T) {
		root := grRoot(t)
		run := grBlocksOf(t, root, ".claude/skills/moai/workflows/run.md")["contract-signing-run"]
		if !strings.Contains(run.body, "moai contract kickoff-check") {
			t.Error("contract-signing-run lacks `moai contract kickoff-check`")
		}
		if !grSentenceWith(run.body, "reject", "human") {
			t.Error("contract-signing-run has no sentence naming both reject and human")
		}
		sa := grBlocksOf(t, root, ".claude/skills/moai/workflows/plan/spec-assembly.md")
		if !strings.Contains(sa["contract-signing-review"].body, "moai contract sign") {
			t.Error("contract-signing-review lacks `moai contract sign`")
		}
		if !strings.Contains(sa["contract-draft"].body, "contract.yaml") {
			t.Error("contract-draft lacks `contract.yaml`")
		}
	})
}
