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

// sectionRefPattern matches `<file>.md § <heading>` references. The heading
// ends at the first closing delimiter (backtick, paren, bracket, table pipe)
// or line end; trailing punctuation is trimmed by the caller.
var sectionRefPattern = regexp.MustCompile(`([A-Za-z0-9][A-Za-z0-9._/-]*\.md) § ([^\n` + "`" + `)|\]]+)`)

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

// headingSet extracts the normalized heading texts of a markdown file.
func headingSet(content string) map[string]bool {
	out := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		if !isHeadingLine(line) {
			continue
		}
		t := strings.TrimSpace(line)
		t = strings.TrimLeft(t, "#")
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		out[normHeading(t)] = true
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
		headings[f] = headingSet(string(data))
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
					if headings[tgt][want] {
						ok = true
						break
					}
					for _, fam := range companionFamilyOf(base, dir, files) {
						if headings[fam][want] {
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
