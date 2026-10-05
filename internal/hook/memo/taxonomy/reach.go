// reach.go — the reachability model of SPEC-MEMORY-FOLD-BUDGET-001 §1.5.
//
// Link extraction and classification, the store snapshot, the index set I(S),
// the reachable set R(S), the target set T(S), the archive-index selection
// A(S), and the fold invariant checker (a)(b)(c)(d). The checker is the
// single judge every apply path and every test of the SPEC shares: (a)-(c)
// are inclusion checks over link targets, and (d) is an exact equality of the
// store with the plan — the only invariant that sees a deleted link-free
// line, a stray edit, or a wrong destination (spec.md P12).
//
// The checker takes snapshots, never paths into the operator's store.
package taxonomy

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// LinkClass classifies one markdown link target (§1.5).
type LinkClass string

// Link target classes (§1.5): a bare name (optionally ./-prefixed) is
// store-local, a /-prefixed path is absolute, anything else carrying a path
// separator is repo-relative.
const (
	LinkStoreLocal   LinkClass = "store-local"
	LinkAbsolute     LinkClass = "absolute"
	LinkRepoRelative LinkClass = "repo-relative"
)

// archiveIndexPattern matches the card-archive index names A(S) is chosen
// among. The definition of A(S) lives here and nowhere else (plan.md §J):
// every requirement that names the archive index refers to this selection.
var archiveIndexPattern = regexp.MustCompile(`^project_card_archive_[0-9]{4}_[0-9]{2}\.md$`)

// StoreSnapshot is an immutable, in-memory image of one memory store: every
// store-root .md file keyed by base name, byte for byte. The checker takes
// snapshots, never paths into the operator's store.
type StoreSnapshot map[string][]byte

// FoldPlan is a fold's plan: the line texts removed from MEMORY.md and the
// line texts appended to the archive index, in original order. Invariant (d)
// rebuilds the expected store from the plan alone, so the plan is the only
// description of what the apply may change.
type FoldPlan struct {
	Removed  []string
	Appended []string
}

// SnapshotStore reads dir into a snapshot: every store-root .md file, byte
// for byte, keyed by base name. Subdirectories (the _archive/ remedy
// directory among them) are never read, and symlinks are skipped — the same
// refusal ParseFile applies.
func SnapshotStore(dir string) (StoreSnapshot, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: read store %s: %w", dir, err)
	}
	snap := StoreSnapshot{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("taxonomy: read %s: %w", filepath.Join(dir, e.Name()), err)
		}
		snap[e.Name()] = data
	}
	return snap, nil
}

// ExtractLinkTargets returns every markdown link target ending in .md, in
// order of appearance, duplicates preserved, on any line shape (grouped lines
// and mid-line links included — never only lines beginning `- [`). The full
// target text is returned, never its base name (P6): distinct repo-relative
// targets sharing a base name stay distinct here.
func ExtractLinkTargets(content string) []string {
	matches := markdownLinkTarget.FindAllStringSubmatch(content, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

// ClassifyLinkTarget classifies one link target per §1.5.
func ClassifyLinkTarget(target string) LinkClass {
	if strings.HasPrefix(target, "/") {
		return LinkAbsolute
	}
	if rest, ok := strings.CutPrefix(target, "./"); ok &&
		!strings.ContainsAny(rest, `/\`) {
		return LinkStoreLocal
	}
	if !strings.ContainsAny(target, `/\`) {
		return LinkStoreLocal
	}
	return LinkRepoRelative
}

// storeLocalName is the store file name a store-local target names, or ""
// when the target is not store-local.
func storeLocalName(target string) string {
	if ClassifyLinkTarget(target) != LinkStoreLocal {
		return ""
	}
	return strings.TrimPrefix(target, "./")
}

// resolvedTargets returns the distinct existing store files one file's
// store-local targets name — the file itself and MEMORY.md excluded. This is
// the checker's link-resolution rule, which counts store-local targets only;
// the doctor resolves by base name and the one shape the two can disagree on
// (a repo-relative target whose base name equals a store file) is P6 of
// spec.md — there the checker is the stricter reading.
func (s StoreSnapshot) resolvedTargets(name string) map[string]bool {
	out := map[string]bool{}
	content, ok := s[name]
	if !ok {
		return out
	}
	for _, t := range ExtractLinkTargets(string(content)) {
		n := storeLocalName(t)
		if n == "" || n == name || n == indexFileName {
			continue
		}
		if _, exists := s[n]; !exists {
			continue
		}
		out[n] = true
	}
	return out
}

// ResolvedLinkCount returns the number of distinct existing store files the
// store file name resolves links to — the resolved links of file f in §1.5.
// The fold reads the archive index's count after the fold through this so
// the under-threshold refusal (REQ-MFB-006) judges with the checker's own
// resolution rule, never a second one.
func (s StoreSnapshot) ResolvedLinkCount(name string) int {
	return len(s.resolvedTargets(name))
}

// LineTargetSetKey exposes the comparable key of one line's link-target set —
// the identity REQ-MFB-004's step-2 re-read and REQ-MFB-005's differing-text
// check compare lines by. The checker uses it internally for invariant (c);
// exporting it keeps one statement of the identity.
func LineTargetSetKey(line string) string {
	return lineTargetSetKey(line)
}

// IsArchiveIndexName reports whether name matches the archive-index shape
// project_card_archive_<YYYY>_<MM>.md — the shape A(S) is selected among
// (§1.5). The selection itself stays ArchiveIndexName, the single statement
// of the criterion; this exposes the shape for the fold's information
// listing of present-but-unlinked candidates.
func IsArchiveIndexName(name string) bool {
	return archiveIndexPattern.MatchString(name)
}

// IndexSet returns I(S) (§1.5): MEMORY.md itself plus every store file
// carrying at least SecondaryIndexLinkThreshold() resolved links. The
// threshold is read through the accessor, so it cannot drift between the
// doctor and the checker.
func (s StoreSnapshot) IndexSet() map[string]bool {
	out := map[string]bool{}
	if _, ok := s[indexFileName]; ok {
		out[indexFileName] = true
	}
	th := SecondaryIndexLinkThreshold()
	for name := range s {
		if name == indexFileName {
			continue
		}
		if len(s.resolvedTargets(name)) >= th {
			out[name] = true
		}
	}
	return out
}

// ReachableSet returns R(S) (§1.5): the store files that appear as a
// store-local target in any member of I(S).
func (s StoreSnapshot) ReachableSet() map[string]bool {
	out := map[string]bool{}
	for member := range s.IndexSet() {
		for _, t := range ExtractLinkTargets(string(s[member])) {
			n := storeLocalName(t)
			if n == "" {
				continue
			}
			if _, exists := s[n]; exists {
				out[n] = true
			}
		}
	}
	return out
}

// TargetSet returns T(S) (§1.5): the distinct full link-target strings
// occurring in any member of I(S). Counted by full text, never by base name.
func (s StoreSnapshot) TargetSet() map[string]bool {
	out := map[string]bool{}
	for member := range s.IndexSet() {
		for _, t := range ExtractLinkTargets(string(s[member])) {
			out[t] = true
		}
	}
	return out
}

// ArchiveIndexName returns A(S) (§1.5): among the store-root files whose
// names match project_card_archive_<YYYY>_<MM>.md, those MEMORY.md links as a
// store-local target and that exist, the greatest name (byte-wise name
// order). A matching file MEMORY.md does not link is never A(S), however late
// its month; when no matching file is both linked and present, A(S) is empty.
// A(S) is never created by any operation of this SPEC.
func (s StoreSnapshot) ArchiveIndexName() string {
	idx, ok := s[indexFileName]
	if !ok {
		return ""
	}
	linked := map[string]bool{}
	for _, t := range ExtractLinkTargets(string(idx)) {
		if n := storeLocalName(t); n != "" {
			linked[n] = true
		}
	}
	best := ""
	for name := range s {
		if name == indexFileName || !archiveIndexPattern.MatchString(name) || !linked[name] {
			continue
		}
		if best == "" || name > best {
			best = name
		}
	}
	return best
}

// CheckFoldInvariants evaluates invariants (a)-(d) of §1.5 between the before
// and after snapshots of one fold, rebuilding the expected MEMORY.md and
// archive index from the plan alone (preserving line endings and the
// trailing-newline state) and comparing byte for byte. It returns the
// violated invariant identifiers in fixed order — a, b, c, d1, d2, d3 — and
// an empty result means every invariant holds.
func CheckFoldInvariants(before, after StoreSnapshot, plan FoldPlan) []string {
	var violated []string
	if !supersetOf(after.ReachableSet(), before.ReachableSet()) {
		violated = append(violated, "a")
	}
	if !supersetOf(after.TargetSet(), before.TargetSet()) {
		violated = append(violated, "b")
	}
	if !removedLinesRehomed(after, plan.Removed) {
		violated = append(violated, "c")
	}
	if !memoryMatchesPlan(before, after, plan) {
		violated = append(violated, "d1")
	}
	if !archiveMatchesPlan(before, after, plan) {
		violated = append(violated, "d2")
	}
	if !otherFilesUnchanged(before, after) {
		violated = append(violated, "d3")
	}
	return violated
}

func supersetOf(super, sub map[string]bool) bool {
	for k := range sub {
		if !super[k] {
			return false
		}
	}
	return true
}

// removedLinesRehomed is invariant (c): every line removed from MEMORY.md has
// an equal-target line — the same set of link targets on one line — in
// another member of I(after), i.e. a member other than MEMORY.md, the file
// the lines were removed from.
func removedLinesRehomed(after StoreSnapshot, removed []string) bool {
	sets := map[string]bool{}
	for member := range after.IndexSet() {
		if member == indexFileName {
			continue
		}
		for _, line := range strings.Split(string(after[member]), "\n") {
			sets[lineTargetSetKey(line)] = true
		}
	}
	for _, line := range removed {
		if !sets[lineTargetSetKey(line)] {
			return false
		}
	}
	return true
}

// lineTargetSetKey is the comparable key of one line's link-target set.
func lineTargetSetKey(line string) string {
	seen := map[string]bool{}
	set := make([]string, 0, 2)
	for _, t := range ExtractLinkTargets(line) {
		if !seen[t] {
			seen[t] = true
			set = append(set, t)
		}
	}
	sort.Strings(set)
	return strings.Join(set, "\x00")
}

// memoryMatchesPlan is invariant (d1): MEMORY.md equals its prior bytes minus
// exactly the removed lines — every other line, heading, blockquote, link
// free line, blank line and line ending, and the trailing-newline state,
// preserved byte for byte and in order.
func memoryMatchesPlan(before, after StoreSnapshot, plan FoldPlan) bool {
	beforeIdx, okBefore := before[indexFileName]
	afterIdx, okAfter := after[indexFileName]
	if !okBefore {
		return !okAfter
	}
	if !okAfter {
		return false
	}
	return bytes.Equal(removePlannedLines(beforeIdx, plan.Removed), afterIdx)
}

// removePlannedLines rebuilds content without the removed lines. Endings
// travel with their lines, so the surviving bytes and the trailing-newline
// state are exactly the original ones; a plan entry consumes matching lines
// in order, one instance per entry.
func removePlannedLines(content []byte, removed []string) []byte {
	remaining := make(map[string]int, len(removed))
	for _, l := range removed {
		remaining[l]++
	}
	var out []byte
	for _, seg := range splitLinesWithEndings(content) {
		if remaining[seg.text] > 0 {
			remaining[seg.text]--
			continue
		}
		out = append(out, seg.text...)
		out = append(out, seg.ending...)
	}
	return out
}

type lineSegment struct {
	text   string
	ending string // "" only for a final line without a terminator
}

// splitLinesWithEndings splits content into (line, ending) segments whose
// concatenation equals content byte for byte.
func splitLinesWithEndings(content []byte) []lineSegment {
	var segs []lineSegment
	for len(content) > 0 {
		idx := bytes.IndexByte(content, '\n')
		if idx < 0 {
			segs = append(segs, lineSegment{text: string(content)})
			break
		}
		text := content[:idx]
		ending := "\n"
		if idx > 0 && content[idx-1] == '\r' {
			text = text[:idx-1]
			ending = "\r\n"
		}
		segs = append(segs, lineSegment{text: string(text), ending: ending})
		content = content[idx+1:]
	}
	return segs
}

// archiveMatchesPlan is invariant (d2): A(S) equals its prior content,
// followed — after one line terminator when the prior content lacked one — by
// exactly the appended lines, verbatim and in their original order, using the
// file's existing line ending. The destination is the before snapshot's A(S):
// the plan named it, and it is not re-chosen after the fact. A plan with
// appended lines and no qualifying A(S) violates (d2) by construction — the
// fold refuses that input before the checker ever runs.
func archiveMatchesPlan(before, after StoreSnapshot, plan FoldPlan) bool {
	dest := before.ArchiveIndexName()
	if dest == "" {
		return len(plan.Appended) == 0
	}
	prior, ok := before[dest]
	if !ok {
		return false
	}
	got, ok := after[dest]
	if !ok {
		return false
	}
	return bytes.Equal(appendPlannedLines(prior, plan.Appended), got)
}

// appendPlannedLines rebuilds the archive index from its prior bytes and the
// appended lines.
func appendPlannedLines(prior []byte, appended []string) []byte {
	out := append([]byte(nil), prior...)
	if len(appended) == 0 {
		return out
	}
	ending := "\n"
	if bytes.Contains(prior, []byte("\r\n")) {
		ending = "\r\n"
	}
	if len(prior) > 0 && prior[len(prior)-1] != '\n' {
		out = append(out, ending...)
	}
	for _, l := range appended {
		out = append(out, l...)
		out = append(out, ending...)
	}
	return out
}

// otherFilesUnchanged is invariant (d3): every file other than MEMORY.md and
// A(S) is byte-identical, and no file is created, renamed or deleted.
func otherFilesUnchanged(before, after StoreSnapshot) bool {
	dest := before.ArchiveIndexName()
	if len(before) != len(after) {
		return false
	}
	for name, beforeBytes := range before {
		afterBytes, ok := after[name]
		if !ok {
			return false
		}
		if name == indexFileName || name == dest {
			continue
		}
		if !bytes.Equal(beforeBytes, afterBytes) {
			return false
		}
	}
	return true
}
