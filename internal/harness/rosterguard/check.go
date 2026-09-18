package rosterguard

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// countCache memoises compiled CountPattern regexps across sites.
var countCache = map[string]*regexp.Regexp{}

// AgentDefinitionNames returns the AxisDefinitionFiles population, derived from
// the .claude/agents/moai/*.md filenames under root.
//
// It is derived and never hand-typed. A hand-typed roster is precisely how the
// defects this package guards are produced: while this card was being measured,
// a hand-extraction whose alphabet was [A-Za-z-] silently dropped `e2e-tester`,
// because that name carries a digit.
func AgentDefinitionNames(root string) ([]string, error) {
	dir := filepath.Join(root, ".claude", "agents", "moai")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read agent definition dir %s: %w", dir, err)
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		out = append(out, strings.TrimSuffix(e.Name(), ".md"))
	}
	sort.Strings(out)
	return out, nil
}

// AxisMembers returns the canonical membership of axis.
//
// retained is the canonical retained roster (template.ProfileMatrixAgents());
// definitions is the definition-file population. Both are passed in rather than
// imported here so the same function is exercised by the control probe with
// deliberately wrong input.
func AxisMembers(axis Axis, retained, definitions []string) ([]string, error) {
	switch axis {
	case AxisRetainedRoster:
		return append([]string(nil), retained...), nil
	case AxisDefinitionFiles:
		return append([]string(nil), definitions...), nil
	case AxisSubsetByDesign:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown axis %q", axis)
	}
}

// ExtractBlock returns the lines of body occupied by the site's claim region.
//
// When BlockEnd is empty the region is the single line containing BlockStart —
// the right scope for a prose sentence that enumerates inline, where taking the
// whole file would let an unrelated later mention satisfy the assertion.
func ExtractBlock(body string, s Site) ([]string, error) {
	lines := strings.Split(body, "\n")
	if s.BlockStart == "" {
		return lines, nil
	}
	start := -1
	for i, ln := range lines {
		if strings.Contains(ln, s.BlockStart) {
			start = i
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("site %s: block anchor %q not found in %s (the anchor moved or the block was deleted)", s.ID, s.BlockStart, s.Path)
	}
	if s.BlockEnd == "" {
		return lines[start : start+1], nil
	}
	for i := start + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], s.BlockEnd) {
			return lines[start : i+1], nil
		}
	}
	return nil, fmt.Errorf("site %s: block terminator %q not found after anchor in %s", s.ID, s.BlockEnd, s.Path)
}

// NamesIn returns the members of universe that appear anywhere in lines,
// sorted.
//
// The match is the FULL name bounded on both sides by a non-identifier
// character. Plain containment is not enough, and the difference is not
// theoretical: a mutant that renamed `manager-git` to `manager-gitXX` in a
// registered site passed a containment-based check, because the old name is
// still a substring of the new one. Under containment, every rename and every
// typo is invisible to this guard.
//
// The boundary alphabet deliberately includes `-`, so `manager-git` does NOT
// match inside `manager-git-two`. No character alphabet is used to EXTRACT
// names — extraction by alphabet is how a hand-measurement of this very card
// silently dropped `e2e-tester`, whose name carries a digit — the names come
// from the definition filenames and only their boundaries are tested here.
func NamesIn(lines []string, universe []string) []string {
	joined := strings.Join(lines, "\n")
	var out []string
	for _, n := range universe {
		if containsWholeName(joined, n) {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// containsWholeName reports whether name occurs in s bounded by non-identifier
// characters on both sides.
func containsWholeName(s, name string) bool {
	for off := 0; ; {
		i := strings.Index(s[off:], name)
		if i < 0 {
			return false
		}
		i += off
		beforeOK := i == 0 || !isNameChar(rune(s[i-1]))
		end := i + len(name)
		afterOK := end == len(s) || !isNameChar(rune(s[end]))
		if beforeOK && afterOK {
			return true
		}
		off = i + 1
	}
}

func isNameChar(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	case r == '-', r == '_':
		return true
	}
	return false
}

// DeclaredCount returns the number captured by the site's CountPattern.
//
// The pattern must match exactly once. Two matches make it ambiguous which
// claim is being asserted; zero matches mean the pattern has drifted off the
// text it was written against — and a silent zero-match would turn the whole
// count assertion into a vacuous pass, which is the failure mode this package
// exists to prevent.
func DeclaredCount(body string, s Site) (int, error) {
	re, ok := countCache[s.CountPattern]
	if !ok {
		var err error
		re, err = regexp.Compile(s.CountPattern)
		if err != nil {
			return 0, fmt.Errorf("site %s: CountPattern %q does not compile: %w", s.ID, s.CountPattern, err)
		}
		if re.NumSubexp() != 1 {
			return 0, fmt.Errorf("site %s: CountPattern %q must carry exactly one capture group, has %d", s.ID, s.CountPattern, re.NumSubexp())
		}
		countCache[s.CountPattern] = re
	}
	m := re.FindAllStringSubmatch(body, -1)
	if len(m) != 1 {
		return 0, fmt.Errorf("site %s: CountPattern %q matched %d times in %s, want exactly 1 (0 = the anchor moved; >1 = ambiguous which claim is asserted)", s.ID, s.CountPattern, len(m), s.Path)
	}
	n, err := strconv.Atoi(m[0][1])
	if err != nil {
		return 0, fmt.Errorf("site %s: captured %q is not an integer: %w", s.ID, m[0][1], err)
	}
	return n, nil
}

// CheckSite asserts one site against its declared axis and returns every
// violation it observes. An empty slice is a pass.
//
// body is the file content; retained and definitions are the two canonical
// populations (see AxisMembers).
func CheckSite(s Site, body string, retained, definitions []string) []string {
	var bad []string

	// A site may legitimately be registered as a subset — but only when it
	// asserts no membership. Declaring both is a contradiction in the registry
	// itself, not a finding about the file.
	if s.Axis == AxisSubsetByDesign && s.Claims.Has(ClaimMembership) {
		return []string{fmt.Sprintf("site %s: registry contradiction — %s carries no membership assertion, so ClaimMembership must not be declared", s.ID, AxisSubsetByDesign)}
	}

	members, err := AxisMembers(s.Axis, retained, definitions)
	if err != nil {
		return []string{fmt.Sprintf("site %s: %v", s.ID, err)}
	}

	if s.Claims.Has(ClaimMembership) {
		block, err := ExtractBlock(body, s)
		if err != nil {
			bad = append(bad, err.Error())
		} else {
			// The universe is the union of both populations, so a name that
			// belongs to neither the expected set nor the axis is still
			// observed rather than silently ignored.
			universe := union(retained, definitions)
			found := NamesIn(block, universe)
			want := expectedMembership(members, s.KnownStale)
			if missing, extra := diff(want, found); len(missing) > 0 || len(extra) > 0 {
				bad = append(bad, membershipMessage(s, missing, extra))
			}
		}
		// A KnownStale marker that claims a gap must actually have one.
		if s.KnownStale != nil && len(s.KnownStale.MissingNames) == 0 && len(s.KnownStale.ExtraNames) == 0 {
			bad = append(bad, fmt.Sprintf("site %s: KnownStale declares no MissingNames and no ExtraNames — a marker that records no gap is a mute, not a record; delete it or name the gap", s.ID))
		}
	}

	if s.Claims.Has(ClaimCount) {
		got, err := DeclaredCount(body, s)
		if err != nil {
			bad = append(bad, err.Error())
		} else {
			want := len(members)
			if s.Axis == AxisSubsetByDesign {
				// A subset site's count claim is about the retained roster it
				// cites, not about its own (legitimately partial) content.
				want = len(retained)
			}
			if s.KnownStale != nil && s.KnownStale.DeclaredCount != 0 {
				switch {
				case s.KnownStale.DeclaredCount == want:
					bad = append(bad, fmt.Sprintf("site %s: KnownStale.DeclaredCount %d now equals the %s population — the site is no longer stale on this claim; delete the marker (follow-up: %s)", s.ID, want, s.Axis, s.KnownStale.FollowUp))
				case got != s.KnownStale.DeclaredCount:
					bad = append(bad, fmt.Sprintf("site %s: %s declares %d but the registry recorded the stale value as %d — the site moved; re-measure and update the KnownStale marker", s.ID, s.Path, got, s.KnownStale.DeclaredCount))
				}
			} else if got != want {
				bad = append(bad, fmt.Sprintf("site %s: %s declares %d on axis %s, which currently has %d; repair the site or declare a KnownStale marker naming the gap", s.ID, s.Path, got, s.Axis, want))
			}
		}
		if s.KnownStale != nil && s.KnownStale.DeclaredCount == 0 && !s.Claims.Has(ClaimMembership) {
			bad = append(bad, fmt.Sprintf("site %s: KnownStale on a count-only site must set DeclaredCount", s.ID))
		}
	}

	return bad
}

func membershipMessage(s Site, missing, extra []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "site %s: %s (axis %s) membership disagrees with the canonical set", s.ID, s.Path, s.Axis)
	if len(missing) > 0 {
		fmt.Fprintf(&b, "; absent from the site: %s", strings.Join(missing, ", "))
	}
	if len(extra) > 0 {
		fmt.Fprintf(&b, "; present at the site but not on the axis: %s", strings.Join(extra, ", "))
	}
	if s.KnownStale != nil {
		fmt.Fprintf(&b, ". This site carries a KnownStale marker (%s, follow-up %s) declaring absent=[%s]; the observed gap no longer matches the declared one — re-measure and update or delete the marker",
			s.KnownStale.Reason, s.KnownStale.FollowUp, strings.Join(s.KnownStale.MissingNames, ", "))
	} else {
		b.WriteString(". Repair the site, or declare a KnownStale marker naming the gap and its follow-up card")
	}
	return b.String()
}

// expectedMembership applies a KnownStale marker to the canonical set, so the
// assertion compares the observed gap against the DECLARED gap rather than
// simply skipping the site.
func expectedMembership(members []string, st *Staleness) []string {
	if st == nil {
		return members
	}
	drop := map[string]bool{}
	for _, n := range st.MissingNames {
		drop[n] = true
	}
	out := make([]string, 0, len(members))
	for _, n := range members {
		if !drop[n] {
			out = append(out, n)
		}
	}
	out = append(out, st.ExtraNames...)
	sort.Strings(out)
	return out
}

func union(a, b []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range append(append([]string(nil), a...), b...) {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// diff returns the members of want absent from got, and the members of got
// absent from want.
func diff(want, got []string) (missing, extra []string) {
	inGot := map[string]bool{}
	for _, s := range got {
		inGot[s] = true
	}
	inWant := map[string]bool{}
	for _, s := range want {
		inWant[s] = true
	}
	for _, s := range want {
		if !inGot[s] {
			missing = append(missing, s)
		}
	}
	for _, s := range got {
		if !inWant[s] {
			extra = append(extra, s)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

// sweepSkipPrefixes are the trees the anti-vacuity sweep does not walk.
//
// Each exclusion is a decision, not housekeeping: docs-site is owned by a
// concurrent card; SPECs, reports and release notes are dated records that
// describe the roster AS IT WAS and must not be rewritten to match today; a
// worktrees dir is another card's checkout.
var sweepSkipPrefixes = []string{
	"docs-site/",
	".moai/specs/",
	".moai/reports/",
	".moai/release-notes/",
	".claude/worktrees/",
	".git/",
	"node_modules/",
	"vendor/",
	"dist/",
	"bin/",
}

// sweepSkipFiles are single files excluded for the same reason as the dated
// trees above.
var sweepSkipFiles = map[string]bool{
	"CHANGELOG.md": true,
}

// SweepThreshold is the number of distinct agent names that makes a file a
// roster listing for the purposes of the anti-vacuity sweep.
//
// It is deliberately below the full roster size: a listing that has gone stale
// by two names is exactly the shape this package exists to find, and a
// threshold set at the full roster size would exclude every stale site — the
// sweep would then be vacuous on its own subject matter.
const SweepThreshold = 10

// maxSweptFileBytes bounds what the sweep reads, so a large binary or vendored
// blob cannot dominate the walk.
const maxSweptFileBytes = 4 << 20

// Sweep walks root and returns every repo-root-relative path that enumerates at
// least SweepThreshold distinct members of universe.
//
// This is the load-bearing half of the guard. The per-site assertions can only
// check sites someone remembered to register; the sweep is what makes a NEW
// listing impossible to add silently — an unregistered one fails the test with
// an instruction to declare its axis.
func Sweep(root string, universe []string) ([]string, error) {
	var found []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			// An unreadable entry is reported, never silently treated as
			// absent: "the sweep found nothing here" and "the sweep could not
			// look here" are different facts.
			return err
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "." {
				return nil
			}
			for _, skip := range sweepSkipPrefixes {
				if strings.HasPrefix(rel+"/", skip) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if sweepSkipFiles[rel] {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil || info.Size() > maxSweptFileBytes || !info.Mode().IsRegular() {
			return nil
		}
		b, readErr := os.ReadFile(p)
		if readErr != nil {
			return nil
		}
		s := string(b)
		n := 0
		for _, name := range universe {
			if containsWholeName(s, name) {
				n++
			}
		}
		if n >= SweepThreshold {
			found = append(found, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(found)
	return found, nil
}
