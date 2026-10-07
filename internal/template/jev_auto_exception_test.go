// jev_auto_exception_test.go — linkage and wording guards for the linked
// amendment that lets the `todo --auto` cycle's own candidate ranking take a
// Jev answer, for selection order only.
//
// TestJevAutoExceptionLinkage requires every amended surface to carry the
// universal token, plus one arming constant in this file, all-or-none and
// first appearing in one commit. TestJevAutoExceptionWording requires every
// amended passage to carry both exception literals in one paragraph, its
// closed targets, and no accuracy claim or over-reach phrasing.
//
// Two rules keep the guard honest about its own source:
//
//   - The arming token is assembled from two pieces (jaeArmName and
//     jaeArmTail). The presence check is a plain substring read of the file
//     and the first-commit check is a pickaxe search, so spelling the token
//     as one literal anywhere in this file would make the arming marker
//     present, and first-appearing, in the guard's own commit. Fixtures
//     build it from the same two pieces.
//   - The arming constant is a standalone const line, not a member of a
//     const block, where gofmt would align the equals sign and break the
//     substring the linked commit flips.
//
// The first amendment's guards (activation_test.go in internal/contract/
// kickoff and the content guards in contract_mode_blocks_test.go) are not
// edited: appending to their marker list fails by design, because their first
// commit differs from this amendment's.
package template_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	// jaeGuardFile is this file, the carrier of the arming marker.
	jaeGuardFile = "internal/template/jev_auto_exception_test.go"
	// jaeMirrorRoot prefixes the template mirror of a live path.
	jaeMirrorRoot = "internal/template/templates/"

	// jaeToken is the universal marker: it is template-neutral (no SPEC id,
	// requirement token, date or hash), so it may sit in mirrors.
	jaeToken = "auto-scoped ranking exception"
	// jaeBound is the literal that bounds the exception, in the same
	// paragraph as jaeToken.
	jaeBound = "selection order only"
	// jaeGate is the default-off setting a long-form passage names.
	jaeGate = "workflow.jev.enabled"
	// jaeFilterEN and jaeFilterKO state that mechanical filters fix the
	// candidate set before any answer is read.
	jaeFilterEN = "mechanical filters"
	jaeFilterKO = "기계적 필터"

	// jaeArmName and jaeArmTail are the two pieces of the arming token.
	jaeArmName = "jevAutoExceptionAmended"
	jaeArmTail = " = true"
)

// jevAutoExceptionAmended arms the guard: false while no surface carries the
// amendment, flipped to true in the single linked commit.
const jevAutoExceptionAmended = true

// jaeMarker is one location the amendment touches and the token that shows it
// amended.
type jaeMarker struct{ path, token string }

// jaeAnchors are the landed documents that already carry the token. They are
// presence-only: they predate the amendment, so they are kept out of the
// first-commit comparison.
var jaeAnchors = []string{
	".claude/rules/moai/workflow/factory-dispatch.md",
	".claude/skills/moai/workflows/gtd.md",
	".claude/agents/moai/manager-todo.md",
}

// jaeRule states what one passage must carry.
type jaeRule struct {
	filter string   // non-empty marks a long-form passage: also jaeGate and this filter literal
	needs  []string // further literals the passage carries
	closed []string // closed targets the passage keeps
	forbid []string // tokens the passage must not carry
}

// jaePassage is one passage a rule applies to.
type jaePassage struct {
	label string
	text  string
	rule  jaeRule
}

// jaeSurface is one amended file (and, when mirror is set, its template
// mirror) with the locator that extracts its passages.
type jaeSurface struct {
	group  string
	path   string
	mirror bool
	locate func(text string) []jaePassage
}

// jaeGroups lists the surface groups and how many passages each must locate,
// so a locator that silently finds nothing cannot pass for a clean group.
var jaeGroups = []struct {
	name string
	want int
}{
	{"go-comments", 2},
	{"config-comment", 2},
	{"catalogue-rows", 4},
	{"spec-core", 4},
	{"spec-manager-todo", 2},
	{"local-guide", 1},
	{"extension-rows", 6},
}

// jaeSurfaces is the confirmed surface list; the marker registry derives from
// it, so cutting a surface removes its rows and nothing else.
func jaeSurfaces() []jaeSurface {
	long := jaeRule{filter: jaeFilterEN}
	kickoffTargets := []string{"contract-mode Kickoff", "llm+jev"}
	return []jaeSurface{
		{group: "go-comments", path: "internal/jev/jev.go", locate: func(s string) []jaePassage {
			return []jaePassage{{"package comment", jaeCommentBlock(s, "display-only"), long}}
		}},
		{group: "go-comments", path: "internal/cli/mcp_jev.go", locate: func(s string) []jaePassage {
			r := jaeRule{
				filter: jaeFilterEN,
				needs:  []string{"this tool"},
				closed: []string{"completion verdict", "merge approval", "queue mutation"},
			}
			return []jaePassage{{"file comment", jaeCommentBlock(s, "display-only"), r}}
		}},
		{group: "config-comment", path: ".moai/config/sections/workflow.yaml", mirror: true, locate: func(s string) []jaePassage {
			r := jaeRule{
				filter: jaeFilterEN,
				closed: []string{"completion verdict", "merge", "queue mutation", "never decides alone"},
			}
			return []jaePassage{{"jev comment", grJevYAMLComment(s), r}}
		}},
		{group: "catalogue-rows", path: ".claude/rules/moai/core/moai-mcp-tools-catalogue.md", mirror: true, locate: func(s string) []jaePassage {
			r := jaeRule{
				needs:  []string{"never through this tool"},
				closed: []string{"completion predicate", "merge approval", "queue mutation", "never decides alone"},
			}
			return []jaePassage{
				{"jev_ask row", grLineWith(s, "| `mcp__moai__jev_ask` |"), r},
				{"Judgment (gated) row", grLineWith(s, "| Judgment (gated) |"), r},
			}
		}},
		{group: "spec-core", path: ".moai/specs/SPEC-JEV-CORE-001/spec.md", locate: func(s string) []jaePassage {
			req := jaeRule{filter: jaeFilterEN, needs: []string{"v0.4.0"}, closed: kickoffTargets}
			bullet := jaeRule{closed: kickoffTargets}
			ps := []jaePassage{
				{"REQ-JEVC-011", grReqBody(s, "REQ-JEVC-011"), req},
				{"REQ-JEVC-012", grReqBody(s, "REQ-JEVC-012"), req},
			}
			if sec, ok := grSection(s, "### Out of Scope — authority"); ok {
				n := 0
				for _, line := range strings.Split(sec, "\n") {
					if strings.HasPrefix(line, "- ") {
						n++
						ps = append(ps, jaePassage{fmt.Sprintf("authority bullet %d", n), line, bullet})
					}
				}
			}
			return ps
		}},
		{group: "spec-manager-todo", path: ".moai/specs/SPEC-MANAGER-TODO-001/spec.md", locate: func(s string) []jaePassage {
			return []jaePassage{
				{"REQ-MT-014", jaeListItem(s, "REQ-MT-014"), jaeRule{closed: []string{"display-only", "never as authority"}}},
				{"REQ-MT-015", jaeListItem(s, "REQ-MT-015"), jaeRule{closed: []string{"queue mutation"}}},
			}
		}},
		{group: "local-guide", path: ".moai/docs/jev-local-operations.md", locate: func(s string) []jaePassage {
			// The pinned paragraph ("예외는 한 곳뿐이다") stays verbatim (the
			// first amendment's guard owns that); the amendment is the
			// paragraph that follows it.
			r := jaeRule{filter: jaeFilterKO}
			return []jaePassage{{"paragraph after the pinned one", jaeBlockAfter(s, "예외는 한 곳뿐이다"), r}}
		}},
		{group: "extension-rows", path: ".claude/rules/moai/development/agent-authoring.md", mirror: true, locate: func(s string) []jaePassage {
			return []jaePassage{{"manager-todo line", grLineWith(s, "consults Jev as a display-only signal"), jaeRule{closed: []string{"display-only"}}}}
		}},
		{group: "extension-rows", path: ".claude/skills/moai/SKILL.md", mirror: true, locate: func(s string) []jaePassage {
			return []jaePassage{{"pick line", grLineWith(s, "never reorder by inferred priority"), jaeRule{closed: []string{"never reorder by inferred priority"}}}}
		}},
		{group: "extension-rows", path: ".claude/skills/moai-ref-jev-question-design/SKILL.md", mirror: true, locate: func(s string) []jaePassage {
			r := jaeRule{
				closed: []string{"completion predicate", "merge approval", "queue mutation"},
				forbid: []string{"internal/jev", "mcp__moai__jev", "jev_ask", "moai jev"},
			}
			return []jaePassage{{"capability paragraph", jaeBlockWith(s, "a labelled model signal a person reads"), r}}
		}},
	}
}

// jaeMarkers is the marker registry: the universal token in every surface file
// (live and mirror), plus the arming token in this file — built from two
// pieces, never spelled as one literal.
func jaeMarkers() []jaeMarker {
	var ms []jaeMarker
	for _, s := range jaeSurfaces() {
		ms = append(ms, jaeMarker{s.path, jaeToken})
		if s.mirror {
			ms = append(ms, jaeMarker{jaeMirrorRoot + s.path, jaeToken})
		}
	}
	return append(ms, jaeMarker{jaeGuardFile, jaeArmName + jaeArmTail})
}

// --- locators ---------------------------------------------------------------

// jaeCommentBlock returns the contiguous `//` block holding the first comment
// line that contains marker ("" when absent).
func jaeCommentBlock(text, marker string) string {
	lines := strings.Split(text, "\n")
	at := -1
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "//") && strings.Contains(l, marker) {
			at = i
			break
		}
	}
	if at < 0 {
		return ""
	}
	lo, hi := at, at
	for lo > 0 && strings.HasPrefix(strings.TrimSpace(lines[lo-1]), "//") {
		lo--
	}
	for hi+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[hi+1]), "//") {
		hi++
	}
	return strings.Join(lines[lo:hi+1], "\n")
}

// jaeListItem returns the list item whose line carries the bold id, with its
// continuation lines, up to a blank line, the next item or a heading.
func jaeListItem(text, id string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if !strings.Contains(l, "**"+id+"**") {
			continue
		}
		end := i + 1
		for end < len(lines) {
			n := lines[end]
			if strings.TrimSpace(n) == "" || strings.HasPrefix(n, "- ") || strings.HasPrefix(n, "#") {
				break
			}
			end++
		}
		return strings.Join(lines[i:end], "\n")
	}
	return ""
}

// jaeBlocks splits text into blank-line separated blocks, lines kept raw.
func jaeBlocks(text string) []string {
	var blocks, cur []string
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) == "" {
			if len(cur) > 0 {
				blocks = append(blocks, strings.Join(cur, "\n"))
				cur = nil
			}
			continue
		}
		cur = append(cur, l)
	}
	if len(cur) > 0 {
		blocks = append(blocks, strings.Join(cur, "\n"))
	}
	return blocks
}

// jaeBlockWith returns the first block containing marker ("" when absent).
func jaeBlockWith(text, marker string) string {
	for _, b := range jaeBlocks(text) {
		if strings.Contains(b, marker) {
			return b
		}
	}
	return ""
}

// jaeBlockAfter returns the block that follows the first block containing
// marker ("" when absent or last).
func jaeBlockAfter(text, marker string) string {
	blocks := jaeBlocks(text)
	for i, b := range blocks {
		if strings.Contains(b, marker) && i+1 < len(blocks) {
			return blocks[i+1]
		}
	}
	return ""
}

// --- checkers under test ---------------------------------------------------------

// jaeHas reports whether the file at rel under repo contains token.
func jaeHas(repo, rel, token string) bool {
	data, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(rel)))
	return err == nil && strings.Contains(string(data), token)
}

// jaeFirstCommit is the oldest commit whose diff adds or removes token in path
// ("" when none, or when git cannot answer — a source tree without history).
func jaeFirstCommit(repo, token, path string) string {
	cmd := exec.Command("git", "log", "--reverse", "--format=%H", "-S"+token, "--", path)
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	first, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return first
}

// autoExceptionLinkageFindings requires the markers to be all present or all
// absent, every landed anchor to carry the token once any marker is present,
// and a complete marker set to first appear in one commit. Anchors are
// presence-only: they predate the amendment, so they are left out of the
// first-commit comparison. With the arming constant false no marker may be
// present; with it true all must be, because the arming token is itself a
// marker.
func autoExceptionLinkageFindings(t *testing.T, repo string, markers []jaeMarker, anchors []string) []string {
	t.Helper()
	var present, absent []string
	for _, m := range markers {
		if jaeHas(repo, m.path, m.token) {
			present = append(present, m.path)
		} else {
			absent = append(absent, m.path)
		}
	}
	if len(present) == 0 {
		return nil
	}
	var f []string
	if len(absent) > 0 {
		f = append(f, "partial amendment: present in "+strings.Join(present, ", ")+"; absent in "+strings.Join(absent, ", "))
	}
	for _, a := range anchors {
		if !jaeHas(repo, a, jaeToken) {
			f = append(f, "dangling amendment: the landed anchor "+a+" does not carry the token while markers are present")
		}
	}
	if len(absent) > 0 {
		// The commit comparison is meaningful for a complete set only.
		return f
	}
	first := ""
	for _, m := range markers {
		c := jaeFirstCommit(repo, m.token, m.path)
		switch {
		case c == "":
			continue
		case first == "":
			first = c
		case c != first:
			f = append(f, "marker in "+m.path+" first appears in "+c+", not "+first)
		}
	}
	return f
}

// jaeParagraphs splits a passage into paragraphs — blank lines, once comment
// markers are stripped, separate them — and collapses each to single-spaced
// text, so a literal wrapped across lines still matches.
func jaeParagraphs(passage string) []string {
	var paras, cur []string
	flush := func() {
		if len(cur) > 0 {
			paras = append(paras, strings.Join(cur, " "))
			cur = nil
		}
	}
	for _, line := range strings.Split(passage, "\n") {
		s := strings.TrimSpace(strings.TrimLeft(strings.TrimPrefix(strings.TrimSpace(line), "//"), "#"))
		if s == "" {
			flush()
			continue
		}
		cur = append(cur, strings.Join(strings.Fields(s), " "))
	}
	flush()
	return paras
}

// The claim set and the over-reach set. The first four claim expressions and
// the over-reach expression are the plan's reference expressions; the two
// exemptions below them narrow the false positives that a validation sentence
// ("the answer set is validated as a whole") and a negated scope sentence
// ("does not apply to any moai todo pick outside the cycle") would trip.
var (
	jaeClaimRes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(?:is|are|was|were) (?:accurate|measured|reliable|validated)\b`),
		regexp.MustCompile(`(?i)\bbeats\b`),
		regexp.MustCompile(`(?i)\boutperform`),
		regexp.MustCompile(`(?i)\bmeasured accuracy`),
		regexp.MustCompile(`(?i)\baccuracy of \d`),
	}
	jaeOverreachRe = regexp.MustCompile("(?i)\\b(?:every|any|all) (?:`?moai todo|picks?)\\b")
	jaeNegationRe  = regexp.MustCompile(`(?i)\b(?:not|never|no|nothing|nor)\b`)
)

// jaeLead returns the last 30 characters of prefix's final sentence.
func jaeLead(prefix string) string {
	if i := strings.LastIndexAny(prefix, ".;"); i >= 0 {
		prefix = prefix[i+1:]
	}
	if len(prefix) > 30 {
		prefix = prefix[len(prefix)-30:]
	}
	return prefix
}

// jaeClaims returns the accuracy-claim phrases in text. "<answer> is validated
// as a whole" describes checking an answer set, not the ordering, and is not a
// claim.
func jaeClaims(text string) []string {
	var out []string
	for _, re := range jaeClaimRes {
		for _, loc := range re.FindAllStringIndex(text, -1) {
			m := text[loc[0]:loc[1]]
			if strings.HasSuffix(strings.ToLower(m), "validated") &&
				strings.HasPrefix(text[loc[1]:], " as a whole") &&
				strings.Contains(strings.ToLower(jaeLead(text[:loc[0]])), "answer") {
				continue
			}
			out = append(out, m)
		}
	}
	return out
}

// jaeOverreach returns the phrases that extend the exception to every, any or
// all moai todo picks. A negation just before the phrase narrows the scope
// instead of widening it.
func jaeOverreach(text string) []string {
	var out []string
	for _, loc := range jaeOverreachRe.FindAllStringIndex(text, -1) {
		if jaeNegationRe.MatchString(jaeLead(text[:loc[0]])) {
			continue
		}
		out = append(out, text[loc[0]:loc[1]])
	}
	return out
}

// jaeWordingFindings checks one passage against its rule: the token and its
// bound in one paragraph; the long-form literals, the rule's own literals and
// closed targets anywhere in the passage; none of the forbidden tokens; and
// neither a claim nor an over-reach phrasing.
func jaeWordingFindings(p jaePassage) []string {
	paras := jaeParagraphs(p.text)
	whole := strings.Join(paras, " ")
	var f []string
	named, bounded := false, false
	for _, para := range paras {
		if strings.Contains(para, jaeToken) {
			named = true
			bounded = bounded || strings.Contains(para, jaeBound)
		}
	}
	switch {
	case !named:
		f = append(f, fmt.Sprintf("%s: does not name the %q", p.label, jaeToken))
	case !bounded:
		f = append(f, fmt.Sprintf("%s: the paragraph naming the exception does not carry %q", p.label, jaeBound))
	}
	need := append([]string(nil), p.rule.needs...)
	if p.rule.filter != "" {
		need = append(need, jaeGate, p.rule.filter)
	}
	for _, n := range need {
		if !strings.Contains(whole, n) {
			f = append(f, fmt.Sprintf("%s: lacks the required literal %q", p.label, n))
		}
	}
	for _, c := range p.rule.closed {
		if !strings.Contains(whole, c) {
			f = append(f, fmt.Sprintf("%s: dropped the closed target %q", p.label, c))
		}
	}
	for _, bad := range p.rule.forbid {
		if strings.Contains(whole, bad) {
			f = append(f, fmt.Sprintf("%s: carries the forbidden token %q", p.label, bad))
		}
	}
	for _, c := range jaeClaims(whole) {
		f = append(f, fmt.Sprintf("%s: claims accuracy (%q)", p.label, c))
	}
	for _, o := range jaeOverreach(whole) {
		f = append(f, fmt.Sprintf("%s: over-reaches beyond the --auto cycle (%q)", p.label, o))
	}
	return f
}

// jaeTokenParagraphs returns the paragraphs of a passage that name the token.
func jaeTokenParagraphs(passage string) []string {
	var out []string
	for _, p := range jaeParagraphs(passage) {
		if strings.Contains(p, jaeToken) {
			out = append(out, p)
		}
	}
	return out
}

// jaeParityFindings compares the paragraphs naming the token in a live passage
// and its mirror. Only that block is compared: the pairs differ elsewhere by
// design (a live setting value, a sanitised line).
func jaeParityFindings(label, live, mirror string) []string {
	a, b := jaeTokenParagraphs(live), jaeTokenParagraphs(mirror)
	if len(a) != len(b) {
		return []string{fmt.Sprintf("%s: %d live paragraphs name the exception, %d mirror paragraphs do", label, len(a), len(b))}
	}
	var f []string
	for i := range a {
		if a[i] != b[i] {
			f = append(f, fmt.Sprintf("%s: mirror paragraph %d differs from the live copy", label, i+1))
		}
	}
	return f
}

// jaeLoad extracts the passages of one group from the tree, live and mirror.
// A locator that finds nothing contributes no passage, so the group's count
// check catches it.
func jaeLoad(t *testing.T, root, group string) []jaePassage {
	t.Helper()
	var out []jaePassage
	for _, s := range jaeSurfaces() {
		if s.group != group {
			continue
		}
		paths := []string{s.path}
		if s.mirror {
			paths = append(paths, jaeMirrorRoot+s.path)
		}
		for _, p := range paths {
			for _, ps := range s.locate(strings.ReplaceAll(grRead(t, root, p), "\r\n", "\n")) {
				if strings.TrimSpace(ps.text) == "" {
					continue
				}
				ps.label = p + " " + ps.label
				out = append(out, ps)
			}
		}
	}
	return out
}

// --- fixtures ----------------------------------------------------------------

// jaeRepo creates a throwaway repository whose git configuration cannot reach
// the caller's.
func jaeRepo(t *testing.T) string {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	repo := t.TempDir()
	grGit(t, repo, "init", "-q")
	return repo
}

// jaeCommit writes files into the throwaway repo and commits them as one commit.
func jaeCommit(t *testing.T, repo string, files map[string]string, msg string) {
	t.Helper()
	for rel, body := range files {
		p := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	grGit(t, repo, "add", "-A")
	grGit(t, repo, "-c", "user.name=f", "-c", "user.email=f@example.com", "commit", "-q", "-m", msg)
}

// jaeMarkerFiles gives every marker file a base body, plus its token unless
// skip says otherwise.
func jaeMarkerFiles(markers []jaeMarker, skip func(jaeMarker) bool) map[string]string {
	files := map[string]string{}
	for _, m := range markers {
		body := "base\n"
		if !skip(m) {
			body += m.token + "\n"
		}
		files[m.path] = body
	}
	return files
}

// jaeAnchorFiles gives every anchor the token, except the one at index bare
// (-1 for none).
func jaeAnchorFiles(bare int) map[string]string {
	files := map[string]string{}
	for i, a := range jaeAnchors {
		body := "landed\n"
		if i != bare {
			body += jaeToken + "\n"
		}
		files[a] = body
	}
	return files
}

// jaeExpect fails unless findings is non-empty and mentions want.
func jaeExpect(t *testing.T, what string, findings []string, want string) {
	t.Helper()
	if len(findings) == 0 {
		t.Fatalf("checker accepted %s", what)
	}
	if joined := strings.Join(findings, "\n"); !strings.Contains(joined, want) {
		t.Fatalf("checker rejected %s without naming %q: %v", what, want, findings)
	}
	t.Logf("observed: %v", findings)
}

// TestJevAutoExceptionLinkage: the amendment markers and the arming constant
// are all present or all absent, and first appear in one commit.
func TestJevAutoExceptionLinkage(t *testing.T) {
	markers := jaeMarkers()
	none := func(jaeMarker) bool { return false }
	armingRow := func(m jaeMarker) bool { return m.path == jaeGuardFile }

	for i, row := range markers {
		t.Run(fmt.Sprintf("falsifier/partial/%d", i), func(t *testing.T) {
			repo := jaeRepo(t)
			jaeCommit(t, repo, jaeAnchorFiles(-1), "anchors landed first")
			jaeCommit(t, repo, jaeMarkerFiles(markers, func(m jaeMarker) bool { return m == row }), "linked but one row")
			jaeExpect(t, "a registry missing "+row.path, autoExceptionLinkageFindings(t, repo, markers, jaeAnchors), "absent in "+row.path)
		})
	}
	t.Run("falsifier/arming-only", func(t *testing.T) {
		repo := jaeRepo(t)
		jaeCommit(t, repo, jaeAnchorFiles(-1), "anchors landed first")
		jaeCommit(t, repo, jaeMarkerFiles(markers, func(m jaeMarker) bool { return !armingRow(m) }), "arming constant alone")
		jaeExpect(t, "the arming constant without the markers", autoExceptionLinkageFindings(t, repo, markers, jaeAnchors), "partial amendment")
	})
	t.Run("falsifier/self-match", func(t *testing.T) {
		// The guard file spells the arming token beside the false constant: the
		// checker must report that marker present, so the tree subtest fails
		// at armed=false instead of letting the marker first appear in G.
		repo := jaeRepo(t)
		files := jaeMarkerFiles(markers, func(jaeMarker) bool { return true })
		files[jaeGuardFile] = "const " + jaeArmName + " = false\n// " + jaeArmName + jaeArmTail + "\n"
		jaeCommit(t, repo, jaeAnchorFiles(-1), "anchors landed first")
		jaeCommit(t, repo, files, "guard spelling its own arming token")
		jaeExpect(t, "a guard that spells its arming token", autoExceptionLinkageFindings(t, repo, markers, jaeAnchors), "present in "+jaeGuardFile)
	})
	t.Run("falsifier/split-commits", func(t *testing.T) {
		repo := jaeRepo(t)
		jaeCommit(t, repo, jaeAnchorFiles(-1), "anchors landed first")
		last := markers[len(markers)-1]
		jaeCommit(t, repo, jaeMarkerFiles(markers, func(m jaeMarker) bool { return m == last }), "markers without the last row")
		jaeCommit(t, repo, map[string]string{last.path: "base\n" + last.token + "\n"}, "the last row alone")
		jaeExpect(t, "markers split across commits", autoExceptionLinkageFindings(t, repo, markers, jaeAnchors), "first appears in")
	})
	t.Run("falsifier/dangling-anchor", func(t *testing.T) {
		repo := jaeRepo(t)
		jaeCommit(t, repo, jaeAnchorFiles(0), "anchors landed first, one without the token")
		jaeCommit(t, repo, jaeMarkerFiles(markers, none), "linked")
		jaeExpect(t, "markers over an anchor that lacks the token", autoExceptionLinkageFindings(t, repo, markers, jaeAnchors), "dangling")
	})
	t.Run("all-in-one-commit", func(t *testing.T) {
		// Anchors land first and stay out of the first-commit comparison.
		repo := jaeRepo(t)
		jaeCommit(t, repo, jaeAnchorFiles(-1), "anchors landed first")
		jaeCommit(t, repo, jaeMarkerFiles(markers, none), "linked")
		if f := autoExceptionLinkageFindings(t, repo, markers, jaeAnchors); len(f) != 0 {
			t.Fatalf("complete linkage rejected: %v", f)
		}
	})
	t.Run("tree", func(t *testing.T) {
		root := grRoot(t)
		t.Logf("armed=%v", jevAutoExceptionAmended)
		for _, a := range jaeAnchors {
			if !jaeHas(root, a, jaeToken) {
				t.Errorf("landed anchor %s does not carry the token", a)
			}
		}
		for _, m := range markers {
			if has := jaeHas(root, m.path, m.token); has != jevAutoExceptionAmended {
				t.Errorf("armed=%v but the marker in %s is present=%v", jevAutoExceptionAmended, m.path, has)
			}
		}
		for _, f := range autoExceptionLinkageFindings(t, root, markers, jaeAnchors) {
			t.Error(f)
		}
	})
}

// --- wording fixtures ----------------------------------------------------------

// jaeFixtureRule is the rule of the wording fixtures: a long-form passage.
var jaeFixtureRule = jaeRule{
	filter: jaeFilterEN,
	closed: []string{"completion verdict", "merge approval", "queue mutation"},
}

// jaeGood is a long-form fixture that satisfies jaeFixtureRule.
func jaeGood() string {
	return strings.Join([]string{
		"// The capability is display-only: an answer is a signal for a reader.",
		"//",
		"// A second exception is the todo --auto cycle's own candidate ranking, the " + jaeToken + ".",
		"// Behind the default-off " + jaeGate + " gate an in-process caller may supply the key",
		"// that sets " + jaeBound + "; " + jaeFilterEN + " fix the candidate set before any",
		"// answer is read. It is never a completion verdict, a merge approval or a queue mutation.",
	}, "\n")
}

func jaeFixturePassage(text string) jaePassage {
	return jaePassage{label: "fixture", text: text, rule: jaeFixtureRule}
}

// jaeBeside appends a sentence to the exception paragraph of jaeGood.
func jaeBeside(sentence string) string {
	return jaeGood() + "\n// " + sentence
}

// jaeRejects fails unless the checker flags bad (and accepted the control).
func jaeRejects(t *testing.T, what, bad, want string) {
	t.Helper()
	if f := jaeWordingFindings(jaeFixturePassage(jaeGood())); len(f) != 0 {
		t.Fatalf("control passage rejected: %v", f)
	}
	jaeExpect(t, what, jaeWordingFindings(jaeFixturePassage(bad)), want)
}

// TestJevAutoExceptionWording: every passage that restates the principle also
// names the auto-scoped exception, bounded, keeps its closed targets, and
// makes no accuracy claim.
func TestJevAutoExceptionWording(t *testing.T) {
	root := grRoot(t)

	t.Run("falsifier/literal-missing", func(t *testing.T) {
		jaeRejects(t, "a passage without the exception literal",
			strings.Replace(jaeGood(), jaeToken, "ranking exception", 1), "does not name")
	})
	t.Run("falsifier/closed-target-dropped", func(t *testing.T) {
		jaeRejects(t, "a passage that dropped a closed target",
			strings.Replace(jaeGood(), "a queue mutation", "anything else", 1), `closed target "queue mutation"`)
	})
	t.Run("falsifier/bound-in-other-paragraph", func(t *testing.T) {
		bad := strings.Replace(jaeGood(), "sets "+jaeBound, "sets the order", 1) + "\n//\n// Scope: " + jaeBound + "."
		jaeRejects(t, "a passage whose bound sits in another paragraph", bad, "does not carry")
	})
	t.Run("falsifier/long-form-literal-missing", func(t *testing.T) {
		jaeRejects(t, "a long-form passage without the gate literal",
			strings.Replace(jaeGood(), jaeGate, "the", 1), jaeGate)
		jaeRejects(t, "a long-form passage without the filter literal",
			strings.Replace(jaeGood(), jaeFilterEN, "some checks", 1), jaeFilterEN)
	})
	t.Run("falsifier/claim-phrasing", func(t *testing.T) {
		for _, phrase := range []string{
			"It sets selection order only, and the ordering is accurate.",
			"It sets selection order only, and the ordering beats the fallback.",
			"It sets selection order only, with a measured accuracy of 80%.",
			"The ordering is validated as a whole.",
		} {
			jaeRejects(t, "the claim "+phrase, jaeBeside(phrase), "claims accuracy")
		}
	})
	t.Run("falsifier/over-reach-phrasing", func(t *testing.T) {
		for _, phrase := range []string{
			"It sets selection order only, and it applies to every moai todo pick.",
			"It applies to every moai todo pick outside the --auto cycle.",
		} {
			jaeRejects(t, "the over-reach "+phrase, jaeBeside(phrase), "over-reaches")
		}
	})
	t.Run("disclaimer-not-flagged", func(t *testing.T) {
		for _, phrase := range []string{
			"No ordering accuracy is claimed.",
			"It is not claimed to be accurate.",
			"It is never the basis of a completion verdict, a merge approval or any other decision that is hard to undo.",
			"Nothing here is any other decision that is hard to undo.",
			"The answer set is validated as a whole before it is used.",
			"An answer that is validated as a whole may set selection order only.",
			"It does not apply to any moai todo pick outside the --auto cycle.",
			"Nothing here changes all picks outside the cycle.",
		} {
			if f := jaeWordingFindings(jaeFixturePassage(jaeBeside(phrase))); len(f) != 0 {
				t.Errorf("checker flagged the legitimate sentence %q: %v", phrase, f)
			}
		}
	})
	t.Run("mirror-parity", func(t *testing.T) {
		same := jaeGood()
		if f := jaeParityFindings("fixture", same, same); len(f) != 0 {
			t.Fatalf("identical blocks rejected: %v", f)
		}
		drift := strings.Replace(same, jaeFilterEN, "some checks", 1)
		jaeExpect(t, "a mirror that drifted", jaeParityFindings("fixture", same, drift), "differs")
		if !jevAutoExceptionAmended {
			t.Log("tree comparison deferred: the surfaces carry no exception block while armed=false")
			return
		}
		for _, s := range jaeSurfaces() {
			if !s.mirror {
				continue
			}
			live := s.locate(strings.ReplaceAll(grRead(t, root, s.path), "\r\n", "\n"))
			mirror := s.locate(strings.ReplaceAll(grRead(t, root, jaeMirrorRoot+s.path), "\r\n", "\n"))
			if len(live) != len(mirror) {
				t.Errorf("%s: %d live passages, %d mirror passages", s.path, len(live), len(mirror))
				continue
			}
			for i := range live {
				for _, f := range jaeParityFindings(s.path+" "+live[i].label, live[i].text, mirror[i].text) {
					t.Error(f)
				}
			}
		}
	})

	for _, g := range jaeGroups {
		t.Run(g.name, func(t *testing.T) {
			passages := jaeLoad(t, root, g.name)
			if len(passages) != g.want {
				t.Errorf("located %d passages, want %d", len(passages), g.want)
			}
			if !jevAutoExceptionAmended {
				t.Skipf("surfaces unamended: armed=false; %d passages located, wording checks run once armed", len(passages))
			}
			for _, p := range passages {
				for _, f := range jaeWordingFindings(p) {
					t.Error(f)
				}
			}
		})
	}
}
