package template

// TestDeployedRuleSectionRefsResolve — the § cross-reference resolution and
// migrated-section pointer guard of SPEC-ALWAYS-LOADED-BUDGET-001
// (REQ-ALB-016, REQ-ALB-017 pointer half, AC-ALB-023).
//
// Two assertions over the deployed rules tree:
//
//  1. Every `<file>.md § <heading>` cross-reference inside the deployed rules
//     tree resolves to an existing heading — in the named file, or in a
//     same-directory companion-family file (`<stem>-*.md`), since the M3 split
//     legitimately relocates sections into such companions.
//  2. Every section the binding ledger relocates to a companion (companion:
//     rows, grouped by source file + heading) keeps at least one pointer line
//     in its original core file — a line naming the companion file and the
//     section heading (REQ-ALB-017: one-line pointer per relocated section).
//
// Two exclusion classes are counted and named, never silent: references to
// files that are not part of the deployed tree at all (operator-side artifacts
// the split cannot break, e.g. a SPEC progress file), and placeholder
// headings written as `<...>` (format examples, not references).

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// sectionRefPattern matches `<file>.md § <heading>` references in both the
// bare form (`file.md § Heading`) and the backtick-quoted form
// (`file.md` § Heading — the form project-root-relative citations use, e.g.
// context-window-management.md's askuser-protocol.md SSOT pointer; before
// the backtick allowance that reference went UNMATCHED, so a broken
// reference there passed the sweep unseen). The heading ends at the first
// closing delimiter (backtick, paren, bracket, table pipe) or line end;
// trailing punctuation is trimmed by the caller.
var sectionRefPattern = regexp.MustCompile(`([A-Za-z0-9][A-Za-z0-9._/-]*\.md)` + "`?" + ` § ([^\n` + "`" + `)|\]]+)`)

// placeholderHeading reports a format placeholder like `<Doctrine Section>`.
func placeholderHeading(h string) bool {
	t := strings.TrimSpace(h)
	return strings.HasPrefix(t, "<") && strings.HasSuffix(t, ">")
}

// normHeading normalizes a heading for comparison: strip markdown emphasis and
// code formatting, collapse whitespace, trim.
func normHeading(s string) string {
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.Join(strings.Fields(s), " ")
}

// commonPrefixLen returns the length of the longest common prefix of a and b.
func commonPrefixLen(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}

// capturedHeadingMatches decides whether a captured § reference text names a
// real heading of the target file. The pointer convention varies: a capture
// can truncate a real heading ("§ Attributable diff-check" for "Attributable
// diff-check detail"), run past it into prose ("§ Parallel Execution applies
// to all commands"), stop at a list separator ("§ Design intent, § The
// leader…"), or sit inside a sentence whose common prefix with the heading
// ends at a word boundary in both strings. The capture names the heading when
// the common prefix covers the whole capture (its trailing separators
// trimmed), covers the whole heading, or spans two or more words ending at a
// non-alphanumeric boundary in both strings; a capture sharing no such
// prefix is the broken case the sweep exists to catch.
func capturedHeadingMatches(headings map[string]bool, want string) bool {
	if headings[want] {
		return true
	}
	wt := strings.Trim(want, " \t.,;:()[]|·→-/\\\"'`*")
	if wt == "" {
		return false
	}
	for h := range headings {
		n := commonPrefixLen(wt, h)
		if n == 0 {
			continue
		}
		if n == len(wt) || n == len(h) {
			return true
		}
		// Two or more shared words name the same section even when the
		// capture runs on into prose whose next character is a letter
		// ("… § Pre-Spawn Sync Check to keep the always-loaded file…").
		if len(strings.Fields(wt[:n])) >= 2 {
			return true
		}
	}
	return false
}

// trimmedRefHeading trims trailing punctuation the regex cannot exclude.
func trimmedRefHeading(h string) string {
	h = strings.TrimSpace(h)
	for len(h) > 0 {
		switch h[len(h)-1] {
		case '.', ',', ';', ')', '*', '`', '"', '\'':
			h = strings.TrimSpace(h[:len(h)-1])
		default:
			return h
		}
	}
	return h
}

// headingSetForRefs extracts the normalized heading texts of a markdown file
// for the §-reference sweep: markdown headings PLUS the bold lead-ins the
// pointer convention also addresses ("**Hoist a tree's evidence before
// disposing it.**" as a paragraph opener, "> **SHA placeholder backfill
// exemption (D3)** —" as a blockquote lead) — real § targets that no #
// line carries. (The shared headingSet stays untouched: the binding-ledger
// checks read it too, and its shape is ledger-versioned.)
func headingSetForRefs(content string) map[string]bool {
	out := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		if !isHeadingLine(line) {
			continue
		}
		t := strings.TrimSpace(line)
		t = strings.TrimLeft(t, "#")
		t = strings.TrimSpace(t)
		if t != "" {
			out[normHeading(t)] = true
		}
	}
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		t = strings.TrimPrefix(t, ">")
		t = strings.TrimSpace(t)
		// A [HARD]-tagged paragraph opens with the tag before its bold
		// lead-in ("[HARD] **Hoist a tree's evidence before disposing it.**").
		t = strings.TrimPrefix(t, "[HARD] ")
		t = strings.TrimSpace(t)
		if !strings.HasPrefix(t, "**") {
			continue
		}
		end := strings.Index(t[2:], "**")
		if end <= 0 {
			continue
		}
		inner := strings.TrimSpace(t[2 : 2+end])
		inner = strings.TrimRight(inner, ".:—-")
		if inner == "" {
			continue
		}
		out[normHeading(inner)] = true
	}
	return out
}

// rulesTreeFiles walks the deployed rules tree and returns slash paths of all
// markdown files.
func rulesTreeFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	_ = filepath.Walk(filepath.Join(root, ".claude", "rules"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil //nolint:nilerr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil //nolint:nilerr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	return out
}

// companionFamilyOf returns the deployed rules-tree files of the family of
// baseName: the file itself plus same-directory siblings whose name extends
// the same stem with "-" (factory-dispatch.md -> factory-dispatch-*.md).
func companionFamilyOf(baseName, dir string, all []string) []string {
	stem := strings.TrimSuffix(baseName, ".md")
	var out []string
	for _, f := range all {
		d := filepath.ToSlash(filepath.Dir(f))
		b := filepath.Base(f)
		if d != filepath.ToSlash(dir) {
			continue
		}
		if b == baseName || strings.HasPrefix(b, stem+"-") {
			out = append(out, f)
		}
	}
	return out
}

func TestDeployedRuleSectionRefsResolve(t *testing.T) {
	root := deployEmbeddedTemplatesForTest(t)
	led := loadBindingLedger(t)

	files := rulesTreeFiles(t, root)
	if len(files) == 0 {
		t.Fatal("deployed rules tree swept no markdown files — empty sweep is a failure, not a pass")
	}
	byBase := map[string][]string{} // basename -> slash paths
	contents := map[string]string{}
	headings := map[string]map[string]bool{}
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f)))
		if err != nil {
			t.Fatalf("read deployed rule %s: %v", f, err)
		}
		contents[f] = string(data)
		headings[f] = headingSetForRefs(string(data))
		byBase[filepath.Base(f)] = append(byBase[filepath.Base(f)], f)
	}

	// ---- Half 1: every § reference resolves -------------------------------
	checked, resolved, placeholder, notDeployed := 0, 0, 0, 0
	var unresolved []string
	for _, f := range files {
		dir := filepath.Dir(f)
		for i, line := range strings.Split(contents[f], "\n") {
			for _, m := range sectionRefPattern.FindAllStringSubmatch(line, -1) {
				base, head := m[1], trimmedRefHeading(m[2])
				if placeholderHeading(head) {
					placeholder++
					continue
				}
				checked++
				targets := byBase[base]
				if len(targets) == 0 {
					// The referenced file is not in the deployed rules tree at
					// all: outside the split's reach (named, counted).
					notDeployed++
					t.Logf("[section-ref] target not deployed (excluded, named): %s § %s (in %s:%d)", base, head, f, i+1)
					continue
				}
				want := normHeading(head)
				ok := false
				for _, tgt := range targets {
					if capturedHeadingMatches(headings[tgt], want) {
						ok = true
						break
					}
					for _, fam := range companionFamilyOf(base, dir, files) {
						if capturedHeadingMatches(headings[fam], want) {
							ok = true
							break
						}
					}
					if ok {
						break
					}
				}
				if ok {
					resolved++
				} else {
					unresolved = append(unresolved, albRefStr(base, head, f, i+1))
				}
			}
		}
	}
	t.Logf("[section-ref] swept files=%d refs checked=%d resolved=%d placeholder=%d not-deployed-excluded=%d unresolved=%d",
		len(files), checked, resolved, placeholder, notDeployed, len(unresolved))
	if checked == 0 {
		t.Errorf("section-ref check swept 0 references — empty sweep is a failure, not a pass")
	}
	sort.Strings(unresolved)
	for _, u := range unresolved {
		t.Errorf("unresolved section reference: %s", u)
	}

	// ---- Half 2: migrated sections keep a pointer line --------------------
	// Group companion: rows by (member file, heading); each group needs >=1
	// line in the member's deployed file naming the companion basename and
	// the heading.
	type migKey struct{ member, heading string }
	groups := map[migKey]map[string]bool{} // key -> companion basenames
	for i := range led.Rows {
		r := &led.Rows[i]
		if !strings.HasPrefix(r.Location, "companion:") {
			continue
		}
		member := templatePathToMember(r.Source.File)
		compBase := filepath.Base(strings.TrimPrefix(r.Location, "companion:"))
		k := migKey{member, r.Source.Heading}
		if groups[k] == nil {
			groups[k] = map[string]bool{}
		}
		groups[k][compBase] = true
	}
	if len(groups) == 0 {
		t.Errorf("pointer check swept 0 companion: rows — empty sweep is a failure, not a pass")
	}
	pointerSeen, missing := 0, 0
	var missingList []string
	for k, comps := range groups {
		rel := ".claude/rules/moai/" + k.member
		content, ok := contents[rel]
		if !ok {
			missingList = append(missingList, albPointerStr(k.member, k.heading, "core file absent from deployed tree"))
			missing++
			continue
		}
		found := false
		for _, line := range strings.Split(content, "\n") {
			for base := range comps {
				if strings.Contains(line, base) && strings.Contains(normLine(line), normLine(k.heading)) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			pointerSeen++
		} else {
			missingList = append(missingList, albPointerStr(k.member, k.heading, "no line names the companion "+strings.Join(keysOf(comps), ", ")))
			missing++
		}
	}
	t.Logf("[migrated-section pointers] swept migrated sections=%d with-pointer=%d missing=%d", len(groups), pointerSeen, missing)
	for _, m := range missingList {
		t.Errorf("migrated section missing its pointer line: %s", m)
	}
}

// normLine normalizes a content line for pointer matching (same classes as
// normHeading, so emphasis does not hide a pointer).
func normLine(s string) string {
	return normHeading(s)
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func albRefStr(base, head, file string, line int) string {
	return "`" + base + "` § " + head + " (at " + file + ":" + albItoa(line) + ")"
}

func albPointerStr(member, heading, why string) string {
	return member + " § " + heading + " — " + why
}

func albItoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
