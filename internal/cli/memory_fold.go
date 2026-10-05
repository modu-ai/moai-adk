// memory_fold.go — `moai memory fold` (SPEC-MEMORY-FOLD-BUDGET-001, plan M3):
// move a closed card's STRONG index lines from MEMORY.md into the card
// archive index A(S), verbatim, under the reachability invariants (a)-(d) of
// spec.md §1.5.
//
// The command previews by default; --yes applies (plan.md OD-6, following
// `drain`). The plan is the only description of what an apply may change:
// the pre-write checker (taxonomy.CheckFoldInvariants) rebuilds the expected
// store from the plan alone and refuses to write when the in-memory result
// diverges, and every rename is preceded by a byte comparison against the
// content the plan was computed from (REQ-MFB-004).
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
	"github.com/spf13/cobra"
)

// foldLineClass is one MEMORY.md line's verdict against a card id
// (REQ-MFB-002). Only STRONG lines move; AMBIGUOUS and MENTION lines are
// listed as kept, each with its reason.
type foldLineClass string

const (
	foldStrong    foldLineClass = "STRONG"
	foldAmbiguous foldLineClass = "AMBIGUOUS"
	foldMention   foldLineClass = "MENTION"
)

// foldKeptReasonDifferent is REQ-MFB-005's differing-text report: a STRONG
// line whose link-target set the archive already carries on another wording.
const foldKeptReasonDifferent = "archive carries a different line for the same targets"

var (
	foldCardTokenPattern = regexp.MustCompile(`t[0-9]+`)
	foldFirstLinkPattern = regexp.MustCompile(`\[([^]]*)\]\(`)
	foldDigitsPattern    = regexp.MustCompile(`^[0-9]+$`)
)

// foldTestSeam carries the test-only injection points of the apply path
// (AC-MFB-007). Production leaves it zero; only memory_fold_test.go sets it,
// and never while another fold apply runs.
type foldTestSeam struct {
	// mutateResult diverges the in-memory result from the plan (cell iv:
	// one extra line removed, or one stray byte in a non-index file).
	mutateResult func(taxonomy.StoreSnapshot) taxonomy.StoreSnapshot
	// mutateDisk writes to the store between the plan and the apply
	// (cell iii: MEMORY.md mutated under the fold).
	mutateDisk func(dir string)
	// failAt injects a write failure: "append" (cell i, before anything
	// lands) or "rewrite" (cell ii, after the archive append landed).
	failAt string
}

var memoryFoldSeam foldTestSeam

// normalizeFoldCardID validates --card per REQ-MFB-001: t<digits> or bare
// digits normalized to t<digits>. Every other value is refused before the
// store is read.
func normalizeFoldCardID(raw string) (string, error) {
	if foldDigitsPattern.MatchString(raw) {
		return "t" + raw, nil
	}
	if len(raw) > 1 && raw[0] == 't' && foldDigitsPattern.MatchString(raw[1:]) {
		return raw, nil
	}
	return "", fmt.Errorf("memory fold: --card %q is not a card id (t<digits> or bare digits)", raw)
}

// newMemoryFoldCmd — `moai memory fold --card <id> [--yes] [--json]
// [--dir PATH]`.
func newMemoryFoldCmd() *cobra.Command {
	var yes, jsonOutput bool
	var dirOverride, card string

	cmd := &cobra.Command{
		Use:   "fold --card <id>",
		Short: "Fold a closed card's index lines into the card archive index",
		Args:  cobra.NoArgs,
		// Refusals and errors carry their message on stderr; the usage
		// block must never reach stdout, whose first line names the store
		// and whose refusal runs must stay empty (AC-MFB-001, REQ-MFB-006).
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cardID, err := normalizeFoldCardID(card)
			if err != nil {
				return err
			}
			store, err := resolveFoldStore(dirOverride)
			if err != nil {
				return err
			}
			indexPath := filepath.Join(store.Dir, memoryIndexName)
			if info, statErr := os.Lstat(indexPath); statErr != nil {
				return fmt.Errorf("memory fold: %s: %w", memoryIndexName, statErr)
			} else if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("memory fold: %s is a symbolic link — refusing", memoryIndexName)
			}

			out := cmd.OutOrStdout()
			if !jsonOutput {
				_, _ = fmt.Fprintf(out, "store: %s (%s)\n", store.Dir, store.Origin)
			}
			before, err := taxonomy.SnapshotStore(store.Dir)
			if err != nil {
				return err
			}
			comp, err := buildFoldPlan(before, cardID)
			if err != nil {
				return err
			}
			plan := memoryFoldPlan{
				Store:    store,
				Card:     cardID,
				Archive:  comp.archive,
				Removed:  comp.plan.Removed,
				Appended: comp.plan.Appended,
				Kept:     comp.kept,
				Unlinked: comp.unlinked,
			}

			if len(comp.plan.Removed) == 0 {
				// Nothing to fold: no line names the card, or every STRONG
				// line is kept under REQ-MFB-005 (REQ-MFB-006).
				if jsonOutput {
					return renderFoldPlanJSON(out, plan)
				}
				_, _ = fmt.Fprintf(out, "no line to fold for %s — nothing remains to fold\n", cardID)
				renderFoldKept(out, comp.kept)
				return nil
			}
			if !yes {
				if jsonOutput {
					return renderFoldPlanJSON(out, plan)
				}
				renderFoldPreview(out, plan)
				return nil
			}
			if memoryFoldSeam.mutateDisk != nil {
				// Test seam (AC-MFB-007 iii): a concurrent writer lands
				// between the plan and the apply; the content re-checks
				// below must catch it.
				memoryFoldSeam.mutateDisk(store.Dir)
			}
			if err := applyFold(store.Dir, before, comp); err != nil {
				return err
			}
			plan.Applied = true
			if jsonOutput {
				return renderFoldPlanJSON(out, plan)
			}
			_, _ = fmt.Fprintf(out, "filed %d line(s) into %s\n", len(comp.plan.Removed), comp.archive)
			renderFoldUnlinked(out, comp.unlinked)
			renderFoldKept(out, comp.kept)
			return nil
		},
	}
	cmd.Flags().StringVar(&card, "card", "", "Card id whose open-work lines to fold (t<digits> or bare digits)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Apply the fold (the default is a preview that writes nothing)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit the plan as one JSON object on stdout")
	cmd.Flags().StringVar(&dirOverride, "dir", "", "Fold within this memory directory instead of the resolved store")
	_ = cmd.MarkFlagRequired("card")
	return cmd
}

// memoryFoldPlan is the plan `--json` emits (REQ-MFB-001): one JSON object
// describing exactly what the apply may change.
type memoryFoldPlan struct {
	Store    memoryStore    `json:"store"`
	Card     string         `json:"card"`
	Archive  string         `json:"archive_index"`
	Removed  []string       `json:"removed"`
	Appended []string       `json:"appended"`
	Kept     []foldKeptLine `json:"kept"`
	// Unlinked lists the archive-pattern files present but not linked from
	// MEMORY.md — information only, never a destination (§1.5; the
	// month-rollover listing of REQ-MFB-006).
	Unlinked []string `json:"unlinked_archive_files"`
	Applied  bool     `json:"applied"`
}

// foldKeptLine is one line the fold keeps in MEMORY.md, with its class and
// reason (REQ-MFB-002, REQ-MFB-005).
type foldKeptLine struct {
	Line   string `json:"line"`
	Class  string `json:"class"`
	Reason string `json:"reason"`
}

// foldComputed carries everything one fold derives from the before
// snapshot: the plan, the in-memory result the plan produces, and the
// listing fields.
type foldComputed struct {
	plan     taxonomy.FoldPlan
	planned  taxonomy.StoreSnapshot
	archive  string
	kept     []foldKeptLine
	unlinked []string
}

// buildFoldPlan classifies MEMORY.md, applies the REQ-MFB-005 dedupe, and
// evaluates every refusal of REQ-MFB-006 against the before snapshot — so a
// refused fold never reaches a write. The returned planned snapshot is the
// in-memory result the plan produces; the pre-write checker judges it.
func buildFoldPlan(before taxonomy.StoreSnapshot, cardID string) (foldComputed, error) {
	var comp foldComputed
	mem, ok := before[memoryIndexName]
	if !ok {
		return comp, fmt.Errorf("memory fold: %s is missing from the store", memoryIndexName)
	}
	memSegs := splitFoldLines(mem)

	var strong []string
	for _, seg := range memSegs {
		class, reason := classifyFoldLine(seg.text, cardID)
		switch class {
		case foldStrong:
			strong = append(strong, seg.text)
		case foldAmbiguous, foldMention:
			comp.kept = append(comp.kept, foldKeptLine{Line: seg.text, Class: string(class), Reason: reason})
		}
	}
	if len(strong) == 0 {
		return comp, nil
	}

	comp.archive = before.ArchiveIndexName()
	// The unlinked listing is information only: the archive-pattern files
	// that are present but that MEMORY.md does not link (§1.5 — never a
	// destination, however late their month).
	linkedNames := map[string]bool{}
	for _, t := range taxonomy.ExtractLinkTargets(string(mem)) {
		if taxonomy.ClassifyLinkTarget(t) == taxonomy.LinkStoreLocal {
			linkedNames[strings.TrimPrefix(t, "./")] = true
		}
	}
	for name := range before {
		if name != memoryIndexName && taxonomy.IsArchiveIndexName(name) && !linkedNames[name] {
			comp.unlinked = append(comp.unlinked, name)
		}
	}
	sort.Strings(comp.unlinked)

	if comp.archive == "" {
		// REQ-MFB-006: name the file to create or link, naming every
		// matching file present but unlinked as a file to link.
		if len(comp.unlinked) > 0 {
			return comp, fmt.Errorf("memory fold: no archive index is linked from %s — link one of %s (or create project_card_archive_<YYYY>_<MM>.md and link it)",
				memoryIndexName, strings.Join(comp.unlinked, ", "))
		}
		return comp, fmt.Errorf("memory fold: no archive index to fold into — create project_card_archive_<YYYY>_<MM>.md and link it from %s", memoryIndexName)
	}

	// REQ-MFB-005: a STRONG line already filed byte for byte is removed
	// without a second append; one whose link-target set the archive already
	// carries on another wording is kept with its reason.
	exact := map[string]bool{}
	keys := map[string]bool{}
	for _, seg := range splitFoldLines(before[comp.archive]) {
		exact[seg.text] = true
		keys[taxonomy.LineTargetSetKey(seg.text)] = true
	}
	for _, line := range strong {
		switch {
		case exact[line]:
			comp.plan.Removed = append(comp.plan.Removed, line)
		case keys[taxonomy.LineTargetSetKey(line)]:
			comp.kept = append(comp.kept, foldKeptLine{Line: line, Class: string(foldStrong), Reason: foldKeptReasonDifferent})
		default:
			comp.plan.Removed = append(comp.plan.Removed, line)
			comp.plan.Appended = append(comp.plan.Appended, line)
		}
	}
	if len(comp.plan.Removed) == 0 {
		return comp, nil // every STRONG line kept under REQ-MFB-005
	}

	// The in-memory result the plan produces — what the apply is allowed to
	// write, byte for byte, and what the checker judges.
	comp.planned = make(taxonomy.StoreSnapshot, len(before))
	for name, data := range before {
		comp.planned[name] = append([]byte(nil), data...)
	}
	comp.planned[memoryIndexName] = removeFoldLines(mem, comp.plan.Removed)
	if len(comp.plan.Appended) > 0 {
		comp.planned[comp.archive] = appendFoldLines(before[comp.archive], comp.plan.Appended)
	}

	// REQ-MFB-006: an archive index that would carry fewer resolved links
	// than the doctor's threshold after the fold is refused, with the count
	// and the threshold named. The count is read through the checker's own
	// resolution rule, never a second one.
	if n := comp.planned.ResolvedLinkCount(comp.archive); n < taxonomy.SecondaryIndexLinkThreshold() {
		return comp, fmt.Errorf("memory fold: archive index %s would carry %d resolved links after the fold, under the secondary-index threshold of %d",
			comp.archive, n, taxonomy.SecondaryIndexLinkThreshold())
	}

	// The pre-write checker (REQ-MFB-004): invariants (a)-(d) of §1.5 must
	// hold on the in-memory result before anything is written.
	if memoryFoldSeam.mutateResult != nil {
		comp.planned = memoryFoldSeam.mutateResult(comp.planned)
	}
	if violated := taxonomy.CheckFoldInvariants(before, comp.planned, comp.plan); len(violated) > 0 {
		return comp, fmt.Errorf("memory fold: the planned result violates invariant(s) %v of the reachability model — aborting without writing", violated)
	}
	return comp, nil
}

// applyFold applies the plan in REQ-MFB-004's order: (1) append the lines to
// the archive index, (2) re-read the archive and confirm each moved line's
// link-target set is present on one line, (3) rewrite MEMORY.md without the
// removed lines. Each file is replaced atomically, and every rename is
// preceded by a comparison of the on-disk bytes against the content the plan
// was computed from.
func applyFold(dir string, before taxonomy.StoreSnapshot, comp foldComputed) error {
	archive := comp.archive

	// A change that landed between the plan and the apply is caught here,
	// before the archive gains a line (AC-MFB-007 iii).
	if err := checkFoldUnchanged(dir, memoryIndexName, before[memoryIndexName]); err != nil {
		return err
	}

	// The archive is compared whole against the bytes the plan was computed
	// from on EVERY path — including a retry that appends nothing. The
	// in-memory checker judged the plan-time archive; the deletion of the
	// original lines is only safe when the disk still holds it. A concurrent
	// author who removed the moved line, or the archive's other links (so it
	// loses index qualification even though the moved line survives), makes
	// the fold refuse here, with MEMORY.md untouched (REQ-MFB-004; gate
	// overlay regression).
	if err := checkFoldUnchanged(dir, archive, before[archive]); err != nil {
		return err
	}

	if len(comp.plan.Appended) > 0 {
		if memoryFoldSeam.failAt == "append" {
			return fmt.Errorf("memory fold: writing %s failed", archive)
		}
		if err := atomicWriteFoldFile(dir, archive, comp.planned[archive], before[archive]); err != nil {
			return err
		}
	}
	// (2) re-read the archive from disk and confirm every moved line's
	// link-target set is present on one line — regardless of whether this
	// attempt appended, so the deletion from MEMORY.md never outruns the
	// archive's observed copy (REQ-MFB-004 step 2).
	data, err := os.ReadFile(filepath.Join(dir, archive))
	if err != nil {
		return fmt.Errorf("memory fold: re-read %s: %w", archive, err)
	}
	keys := map[string]bool{}
	for _, seg := range splitFoldLines(data) {
		keys[taxonomy.LineTargetSetKey(seg.text)] = true
	}
	for _, line := range comp.plan.Removed {
		if !keys[taxonomy.LineTargetSetKey(line)] {
			return fmt.Errorf("memory fold: %s does not carry a line with the targets of the moved line %q — keeping the line in %s", archive, line, memoryIndexName)
		}
	}

	// (3) rewrite MEMORY.md.
	if memoryFoldSeam.failAt == "rewrite" {
		return fmt.Errorf("memory fold: writing %s failed", memoryIndexName)
	}
	return atomicWriteFoldFile(dir, memoryIndexName, comp.planned[memoryIndexName], before[memoryIndexName])
}

// checkFoldUnchanged aborts when the on-disk file no longer holds the bytes
// the plan was computed from (REQ-MFB-004; the host-write race of plan.md
// §G narrows here).
func checkFoldUnchanged(dir, name string, planTime []byte) error {
	current, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("memory fold: re-read %s: %w", name, err)
	}
	if !bytes.Equal(current, planTime) {
		return fmt.Errorf("memory fold: %s changed since the plan was computed — aborting without writing", name)
	}
	return nil
}

// atomicWriteFoldFile replaces dir/name with want: the on-disk bytes must
// still equal planTime (the content re-check), then a temp file in the same
// directory carries the new content and one rename publishes it, so no
// reader observes a partial file. No temporary file remains on any path.
func atomicWriteFoldFile(dir, name string, want, planTime []byte) error {
	if err := checkFoldUnchanged(dir, name, planTime); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".moai-fold-*.tmp")
	if err != nil {
		return fmt.Errorf("memory fold: temp file for %s: %w", name, err)
	}
	tmpName := tmp.Name()
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(want); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("memory fold: write %s: %w", name, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("memory fold: write %s: %w", name, err)
	}
	// Preserve the original file's permission bits on replace — the
	// replacement never widens an existing 0600 to 0644. The 0644 fallback
	// covers a genuinely new file.
	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(filepath.Join(dir, name)); statErr == nil {
		mode = info.Mode().Perm()
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return fmt.Errorf("memory fold: chmod %s: %w", name, err)
	}
	if err := os.Rename(tmpName, filepath.Join(dir, name)); err != nil {
		return fmt.Errorf("memory fold: rename %s: %w", name, err)
	}
	tmpName = ""
	return nil
}

// resolveFoldStore resolves the one store a fold acts on (§1.6): --dir, or
// the first candidate from memoryCandidateStores whose directory carries
// MEMORY.md.
func resolveFoldStore(dirOverride string) (memoryStore, error) {
	if dirOverride != "" {
		abs, err := filepath.Abs(dirOverride)
		if err != nil {
			return memoryStore{}, fmt.Errorf("memory fold: resolve --dir: %w", err)
		}
		return memoryStore{Dir: abs, Origin: "--dir"}, nil
	}
	stores, err := memoryCandidateStores(resolveProjectDir())
	if err != nil {
		return memoryStore{}, err
	}
	for _, s := range stores {
		if info, statErr := os.Stat(filepath.Join(s.Dir, memoryIndexName)); statErr == nil && !info.IsDir() {
			return s, nil
		}
	}
	return memoryStore{}, fmt.Errorf("memory fold: no memory store with %s resolved", memoryIndexName)
}

// classifyFoldLine is REQ-MFB-002. A line is STRONG when it begins `- [`,
// its first link title begins with the card id as a whole token, and at
// least one of its link targets contains the card id as a whole token while
// no target contains a different card id; exactly one of the two conditions
// is AMBIGUOUS; the card id only elsewhere in the text is a MENTION.
func classifyFoldLine(line, cardID string) (foldLineClass, string) {
	titleHit := false
	if strings.HasPrefix(line, "- [") {
		if m := foldFirstLinkPattern.FindStringSubmatch(line); m != nil && foldTitleHasCard(m[1], cardID) {
			titleHit = true
		}
	}
	targetHit, conflict := false, false
	for _, target := range taxonomy.ExtractLinkTargets(line) {
		for _, id := range foldCardTokens(target) {
			if id == cardID {
				targetHit = true
			} else {
				conflict = true
			}
		}
	}
	switch {
	case titleHit && targetHit && !conflict:
		return foldStrong, ""
	case titleHit && targetHit:
		return foldAmbiguous, "targets name more than one card id"
	case titleHit:
		return foldAmbiguous, "title names the card, no target does"
	case targetHit:
		return foldAmbiguous, "a target names the card, the title does not"
	case foldContainsWholeToken(line, cardID):
		return foldMention, "mentions the card id in its text"
	}
	return "", ""
}

// foldTitleHasCard reports whether the link title begins with the card id as
// a whole token.
func foldTitleHasCard(title, card string) bool {
	if !strings.HasPrefix(title, card) {
		return false
	}
	rest := title[len(card):]
	return rest == "" || !foldIsAlnum(rest[0])
}

// foldCardTokens returns the distinct card-id whole tokens of s, in order.
func foldCardTokens(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, loc := range foldCardTokenPattern.FindAllStringIndex(s, -1) {
		if loc[0] > 0 && foldIsAlnum(s[loc[0]-1]) {
			continue
		}
		if loc[1] < len(s) && foldIsAlnum(s[loc[1]]) {
			continue
		}
		tok := s[loc[0]:loc[1]]
		if !seen[tok] {
			seen[tok] = true
			out = append(out, tok)
		}
	}
	return out
}

// foldContainsWholeToken reports whether s carries token bounded by the
// string edge or a character outside [0-9A-Za-z] (REQ-MFB-002).
func foldContainsWholeToken(s, token string) bool {
	for i := 0; i+len(token) <= len(s); i++ {
		if s[i:i+len(token)] != token {
			continue
		}
		if i > 0 && foldIsAlnum(s[i-1]) {
			continue
		}
		if i+len(token) < len(s) && foldIsAlnum(s[i+len(token)]) {
			continue
		}
		return true
	}
	return false
}

// foldIsAlnum reports whether b is inside the [0-9A-Za-z] token boundary set.
func foldIsAlnum(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z'
}

// foldLineSegment is one line of a store file: its text without the
// terminator plus the terminator it carries.
type foldLineSegment struct {
	text   string
	ending string
}

// splitFoldLines splits content into (text, ending) segments whose
// concatenation equals content byte for byte, so removal preserves the
// surviving bytes and the trailing-newline state exactly (§1.5 d1).
func splitFoldLines(content []byte) []foldLineSegment {
	var segs []foldLineSegment
	for len(content) > 0 {
		idx := bytes.IndexByte(content, '\n')
		if idx < 0 {
			segs = append(segs, foldLineSegment{text: string(content)})
			break
		}
		text := string(content[:idx])
		ending := "\n"
		if idx > 0 && content[idx-1] == '\r' {
			text = string(content[:idx-1])
			ending = "\r\n"
		}
		segs = append(segs, foldLineSegment{text: text, ending: ending})
		content = content[idx+1:]
	}
	return segs
}

// removeFoldLines rebuilds content without one instance of each removed
// line; endings travel with their lines.
func removeFoldLines(content []byte, removed []string) []byte {
	remaining := make(map[string]int, len(removed))
	for _, l := range removed {
		remaining[l]++
	}
	var out []byte
	for _, seg := range splitFoldLines(content) {
		if remaining[seg.text] > 0 {
			remaining[seg.text]--
			continue
		}
		out = append(out, seg.text...)
		out = append(out, seg.ending...)
	}
	return out
}

// appendFoldLines rebuilds the archive index from its prior bytes plus the
// appended lines (§1.5 d2): one terminator supplied when the prior content
// lacked one, each line verbatim in order with the file's existing line
// ending.
func appendFoldLines(prior []byte, appended []string) []byte {
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

// renderFoldPreview prints the plan of a dry run (REQ-MFB-001): the STRONG
// lines it would file, the archive index it would file them into, and every
// kept line with its reason.
func renderFoldPreview(out io.Writer, plan memoryFoldPlan) {
	_, _ = fmt.Fprintf(out, "card: %s\n", plan.Card)
	if plan.Archive != "" {
		_, _ = fmt.Fprintf(out, "archive index: %s\n", plan.Archive)
	}
	renderFoldUnlinked(out, plan.Unlinked)
	if len(plan.Removed) > 0 {
		_, _ = fmt.Fprintf(out, "would file %d line(s):\n", len(plan.Removed))
		for _, l := range plan.Removed {
			_, _ = fmt.Fprintf(out, "  %s\n", l)
		}
	}
	renderFoldKept(out, plan.Kept)
	_, _ = fmt.Fprintln(out, "nothing is written without --yes")
}

// renderFoldUnlinked lists the present-but-unlinked archive-pattern files
// as skipped information (REQ-MFB-006, the month-rollover listing).
func renderFoldUnlinked(out io.Writer, unlinked []string) {
	for _, u := range unlinked {
		_, _ = fmt.Fprintf(out, "unlinked archive index (skipped): %s\n", u)
	}
}

// renderFoldKept lists the kept lines, each with its class and reason
// (REQ-MFB-002).
func renderFoldKept(out io.Writer, kept []foldKeptLine) {
	if len(kept) == 0 {
		return
	}
	_, _ = fmt.Fprintf(out, "kept %d line(s):\n", len(kept))
	for _, k := range kept {
		_, _ = fmt.Fprintf(out, "  [%s] %s\n    reason: %s\n", k.Class, k.Line, k.Reason)
	}
}

// renderFoldPlanJSON emits the plan as one JSON object (REQ-MFB-001).
func renderFoldPlanJSON(out io.Writer, plan memoryFoldPlan) error {
	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, string(data))
	return nil
}
