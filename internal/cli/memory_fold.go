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
	"sync"
	"sync/atomic"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
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

// foldKeptReasonArchiveLink is the entry-link keep reason (card-review
// round-1 P2): the STRONG line carries the archive's own entry link and is
// MEMORY.md's only link to it — removing the line would unlink the fold's
// own destination, so the line stays.
const foldKeptReasonArchiveLink = "removing the line would unlink the archive index"

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
	// mutateDuringWrite is invoked inside atomicWriteFoldFile after the temp
	// file is prepared and before the pre-rename re-check — the window a
	// concurrent writer races (gate overlay cell iii-b).
	mutateDuringWrite func(dir, name string)
	// orderProbe records the pre-rename check order of atomicWriteFoldFile
	// ("effective-start" then "bytes-done") — the codex-review round-2
	// ordering regression asserts on this sequence.
	orderProbe func(stage string)
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
			if err := foldIndexGuard(store.Dir); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if !jsonOutput {
				_, _ = fmt.Fprintf(out, "store: %s (%s)\n", store.Dir, store.Origin)
			}
			// REQ-DISPATCH-008: the whole transaction — snapshot read, plan,
			// and both index writes — runs inside the store's cross-process
			// lock, so a concurrent fold's process waits and re-plans
			// against the post-transaction store. The preview and no-fold
			// paths release the lock with nothing written, exactly as they
			// returned nothing before.
			return withFoldStoreLock(store.Dir, func() error {
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
				if err := applyFold(store.Dir, before, comp, nil, nil); err != nil {
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
			})
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
			// The lines THIS fold appends join the dedupe sets: a second
			// identical copy is removed without a second append, and a
			// same-targets different wording is kept with its reason
			// (codex-review round 2 — the same line twice in MEMORY.md
			// must not file twice).
			exact[line] = true
			keys[taxonomy.LineTargetSetKey(line)] = true
		}
	}
	if len(comp.plan.Removed) == 0 {
		return comp, nil // every STRONG line kept under REQ-MFB-005
	}

	// The entry-link guard (card-review round-1 P2): a STRONG line that
	// also links the archive cannot be removed when the removal would take
	// MEMORY.md's LAST link to the fold's own destination — the archive
	// would lose its entry, ArchiveIndexName() would return empty, and
	// every later fold would refuse. Those lines are kept with the reason
	// (their appended twins are dropped with them).
	remainingArchiveLinks := 0
	removedSet := make(map[string]bool, len(comp.plan.Removed))
	for _, line := range comp.plan.Removed {
		removedSet[line] = true
	}
	linksArchive := func(line string) bool {
		for _, tgt := range taxonomy.ExtractLinkTargets(line) {
			if taxonomy.ClassifyLinkTarget(tgt) == taxonomy.LinkStoreLocal && strings.TrimPrefix(tgt, "./") == comp.archive {
				return true
			}
		}
		return false
	}
	for _, seg := range memSegs {
		if removedSet[seg.text] {
			continue
		}
		if linksArchive(seg.text) {
			remainingArchiveLinks++
		}
	}
	if remainingArchiveLinks == 0 {
		var removed, appended []string
		for _, line := range comp.plan.Removed {
			if linksArchive(line) {
				comp.kept = append(comp.kept, foldKeptLine{Line: line, Class: string(foldStrong), Reason: foldKeptReasonArchiveLink})
				continue
			}
			removed = append(removed, line)
		}
		for _, line := range comp.plan.Appended {
			if !linksArchive(line) {
				appended = append(appended, line)
			}
		}
		comp.plan.Removed, comp.plan.Appended = removed, appended
		if len(comp.plan.Removed) == 0 {
			return comp, nil // every STRONG line kept under the entry-link guard
		}
	}

	// Plan-phase target existence (card-review round-2 P2): a removed line
	// whose store-local target was NEVER in the store is a clean refusal
	// BEFORE the archive gains anything — filing the line and only then
	// refusing (the apply-time reachability check) would leave no archive
	// write undone and block every retry on the same broken link. A
	// pre-existing broken link is the store's own state; the fold names it
	// and touches nothing.
	for _, line := range comp.plan.Removed {
		for _, tgt := range taxonomy.ExtractLinkTargets(line) {
			if taxonomy.ClassifyLinkTarget(tgt) != taxonomy.LinkStoreLocal {
				continue
			}
			if _, ok := before[strings.TrimPrefix(tgt, "./")]; !ok {
				return comp, fmt.Errorf("memory fold: the line's target %s is not in the store — refusing before any write (fix or remove the link first)", tgt)
			}
		}
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
func applyFold(dir string, before taxonomy.StoreSnapshot, comp foldComputed, writesForbidden func() bool, run *foldOnDoneRun) error {
	archive := comp.archive

	// An abandoned close-path step (the card-close bound expired) must not
	// begin a write it already reported as never started.
	if writesForbidden != nil && writesForbidden() {
		return fmt.Errorf("memory fold: the step was abandoned — not writing")
	}

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
		if err := atomicWriteFoldFile(dir, archive, comp.planned[archive], before[archive], nil, writesForbidden, run); err != nil {
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

	// (2.5) Re-run qualification and reachability against the CURRENT store
	// before preparing the deletion write (card-review round-1 P2): a
	// concurrent mover may have renamed a referenced topic file — the
	// indexes' bytes still match the plan while the archive's effective
	// resolved-link count dropped below the threshold, or a moved line's
	// target no longer resolves. The SAME verification re-runs at the
	// pre-rename position (guard.check below), so the window between this
	// check and the rename is covered too (codex-review round 2).
	if err := verifyArchiveEffectiveState(dir, before, comp); err != nil {
		return err
	}

	// (3) rewrite MEMORY.md.
	if memoryFoldSeam.failAt == "rewrite" {
		return fmt.Errorf("memory fold: writing %s failed", memoryIndexName)
	}
	// The archive is re-verified against its EXPECTED POST-APPLY bytes in
	// the MEMORY.md write's pre-rename position (guard below): planned equals
	// before on a retry that appends nothing, so the guard reads the same on
	// every path (REQ-MFB-004; gate-overlay P1). The guard's check closure
	// re-runs the effective-state verification at that same position — a
	// move landing during the temp-file preparation invalidates the link
	// count the earlier check measured (codex-review round 2).
	return atomicWriteFoldFile(dir, memoryIndexName, comp.planned[memoryIndexName], before[memoryIndexName],
		&foldRenameGuard{
			name:  archive,
			want:  comp.planned[archive],
			check: func(dir string) error { return verifyArchiveEffectiveState(dir, before, comp) },
		}, writesForbidden, run)
}

// verifyArchiveEffectiveState re-runs the archive's qualification and the
// moved lines' reachability against the CURRENT disk state (card-review
// round-1 P2, codex-review round 2): a concurrent mover can rename a
// referenced topic file at any point up to the last rename — the indexes'
// bytes still match the plan while the archive's effective resolved-link
// count dropped below the threshold, or a moved line's target vanished.
// Deleting the original line in that state strands the card; the caller
// refuses and keeps it. Only a target that EXISTED when the plan was
// computed and is gone NOW is a mid-run move: a never-existed target is
// the plan phase's clean refusal, not a concurrency signal.
func verifyArchiveEffectiveState(dir string, before taxonomy.StoreSnapshot, comp foldComputed) error {
	current, err := taxonomy.SnapshotStore(dir)
	if err != nil {
		return err
	}
	if n := current.ResolvedLinkCount(comp.archive); n < taxonomy.SecondaryIndexLinkThreshold() {
		return fmt.Errorf("memory fold: %s now carries %d resolved links, under the secondary-index threshold of %d — keeping the line in %s", comp.archive, n, taxonomy.SecondaryIndexLinkThreshold(), memoryIndexName)
	}
	for _, line := range comp.plan.Removed {
		for _, tgt := range taxonomy.ExtractLinkTargets(line) {
			if taxonomy.ClassifyLinkTarget(tgt) != taxonomy.LinkStoreLocal {
				continue
			}
			name := strings.TrimPrefix(tgt, "./")
			_, inBefore := before[name]
			_, inCurrent := current[name]
			if inBefore && !inCurrent {
				return fmt.Errorf("memory fold: the moved line's target %s vanished while the fold ran — keeping the line in %s", tgt, memoryIndexName)
			}
		}
	}
	return nil
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

// foldRenameGuard is an additional pre-rename re-check: guard.name must
// still hold guard.want on disk when dir/name is about to be replaced. The
// MEMORY.md rename guards the archive against its EXPECTED POST-APPLY bytes,
// so a concurrent author who edits the archive while the MEMORY.md temp file
// is being prepared aborts the rename instead of being published over
// (REQ-MFB-004; gate-overlay P1 — the line must survive in at least one of
// the two files).
type foldRenameGuard struct {
	name string
	want []byte
	// check, when set, re-verifies the store's EFFECTIVE state at the same
	// pre-rename position — a byte comparison cannot see a referenced file
	// that moved, which invalidates the link count an earlier check
	// measured (codex-review round 2).
	check func(dir string) error
}

// atomicWriteFoldFile replaces dir/name with want: a temp file in the same
// directory carries the new content and one rename publishes it, so no
// reader observes a partial file. Immediately before the rename the content
// re-check runs (it is the last step before the rename), and a non-nil
// guard re-verifies the other file the same way. A non-nil writesForbidden
// check refuses this write when the close-path step was abandoned while it
// ran — an expired step begins no new write. No temporary file remains on
// any path.
func atomicWriteFoldFile(dir, name string, want, planTime []byte, guard *foldRenameGuard, writesForbidden func() bool, run *foldOnDoneRun) error {
	if writesForbidden != nil && writesForbidden() {
		return fmt.Errorf("memory fold: %s: the step was abandoned — not writing", name)
	}
	tmp, err := os.CreateTemp(dir, ".moai-fold-*.tmp")
	if err != nil {
		return fmt.Errorf("memory fold: temp file for %s: %w", name, err)
	}
	tmpName := tmp.Name()
	// Recorded into THIS worker's own state so the card-close caller's
	// timeout branch can recover the temp file when the one-shot CLI
	// process would exit before the worker's deferred removal runs
	// (codex-review round 2). Per-worker with occupancy semantics: a late
	// worker's cleanup can never clear a newer card's recorded path. The
	// fold verb passes no run and records nothing.
	if run != nil {
		run.temp.set(tmpName)
	}
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
			if run != nil {
				run.temp.clearIf(tmpName) // the worker removed it itself
			}
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
	if memoryFoldSeam.mutateDuringWrite != nil {
		memoryFoldSeam.mutateDuringWrite(dir, name)
	}
	// Order (codex-review round 2, TestReviewArchiveUpdateDuringEffectiveScan):
	// the FULL-STORE effective check runs FIRST — it re-reads every topic
	// file, which takes time, and a MEMORY.md or archive update landing
	// DURING it would sail past an already-passed byte comparison. The
	// two-file byte comparison is therefore the LAST step, immediately
	// before the rename (REQ-MFB-004; plan.md §G "immediately before each
	// rename"): whatever the effective check's own window let through, the
	// bytes are judged at the last observable moment.
	if guard != nil {
		if guard.check != nil {
			if memoryFoldSeam.orderProbe != nil {
				memoryFoldSeam.orderProbe("effective-start")
			}
			if err := guard.check(dir); err != nil {
				return err
			}
		}
	}
	if err := checkFoldUnchanged(dir, name, planTime); err != nil {
		return err
	}
	if guard != nil {
		if err := checkFoldUnchanged(dir, guard.name, guard.want); err != nil {
			return err
		}
	}
	if memoryFoldSeam.orderProbe != nil {
		memoryFoldSeam.orderProbe("bytes-done")
	}
	if writesForbidden != nil && writesForbidden() {
		return fmt.Errorf("memory fold: %s: the step was abandoned — not writing", name)
	}
	// The FINAL byte comparison (REQ-DISPATCH-008, AC-DI-009; plan-audit
	// D9): a byte comparison FOLLOWS the probe and is the last check before
	// the rename. Whatever a non-cooperating concurrent author wrote in the
	// probe's window — into THIS file or into the guard file — is detected
	// here and refused, never published over. The irreducible tail between
	// this comparison and the rename is AC-DI-009's stated residual risk
	// against a writer that bypasses every protocol; the store lock
	// (withFoldStoreLock) closes the window for cooperating writers.
	if err := checkFoldUnchanged(dir, name, planTime); err != nil {
		return err
	}
	if guard != nil {
		if err := checkFoldUnchanged(dir, guard.name, guard.want); err != nil {
			return err
		}
	}
	if err := os.Rename(tmpName, filepath.Join(dir, name)); err != nil {
		return fmt.Errorf("memory fold: rename %s: %w", name, err)
	}
	if run != nil {
		run.temp.clearIf(tmpName) // published — there is nothing to recover
	}
	tmpName = ""
	return nil
}

// foldLockName is the store's cross-process lock file (REQ-DISPATCH-008,
// SPEC-DISPATCH-INTEGRITY-001): a dot-prefixed non-.md resident BESIDE the
// store's index. The path must derive from the store alone — never from a
// process-local environment value such as TMPDIR — or two processes
// addressing one store under different environments would resolve different
// lock files and stop serializing (run-gate finding 1). The lock is durable
// named infrastructure of the store: the store-shape helpers
// (storeHashes, requireNoTempFiles) exempt it by name.
const foldLockName = ".moai-store-lock"

// foldLockPath is the lock file's path for one store.
func foldLockPath(dir string) string {
	return filepath.Join(dir, foldLockName)
}

// withFoldStoreLock runs fn inside the store's cross-process lock
// (REQ-DISPATCH-008): the lock spans the fold's whole transaction — the
// snapshot read, the plan, and both index writes — so a second fold's
// PROCESS waits at the lock and re-plans against the post-transaction store
// instead of interleaving its writes with the holder's renames. Two
// `moai memory fold` invocations are separate OS processes; an in-process
// mutex cannot serialize them (constraint C5).
//
// The bounded on-done step keeps its own discipline under the lock: the
// wait for the lock happens inside acquire, and the step's forbidden()
// checks still refuse the write of an abandoned step after the wait.
//
// @MX:WARN: [AUTO] Cross-process file lock — while held it blocks every
// other fold process on this store until released or the holder exits.
// @MX:REASON: a concurrent fold's process must serialize its whole
// transaction (read, plan, both index writes) against the holder's; without
// it a fold completing inside the holder's window lost its line to the
// holder's pre-window snapshot rename (EL-009/EL-002).
func withFoldStoreLock(dir string, fn func() error) error {
	release, err := acquireFoldStoreLock(dir)
	if err != nil {
		return err
	}
	defer release()
	return fn()
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

// ── Card-close wiring (plan M4, REQ-MFB-007, AC-MFB-008) ──────────────────
//
// Every queue close path (todo done, todo auto-done, the todo --auto cycle)
// offers one bounded, fail-open fold of the closed card's index lines. The
// step is gated (OD-1/OD-2: compiled default OFF), bounded per card
// (OD-11), recover-wrapped, and reports at most ONE stderr line — it never
// blocks, never fails, and never prompts.

// foldIndexGuard refuses a MEMORY.md that is missing or a symbolic link
// (REQ-MFB-001/§6). Shared by the fold verb and the card-close step.
func foldIndexGuard(dir string) error {
	indexPath := filepath.Join(dir, memoryIndexName)
	if info, err := os.Lstat(indexPath); err != nil {
		return fmt.Errorf("memory fold: %s: %w", memoryIndexName, err)
	} else if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("memory fold: %s is a symbolic link — refusing", memoryIndexName)
	}
	return nil
}

// foldOnDoneStderr carries the wiring's single stderr line; a variable so
// tests can capture it without swapping os.Stderr process-wide.
var foldOnDoneStderr io.Writer = os.Stderr

// memoryFoldOnDoneBound bounds ONE card's fold-on-done step; a variable so
// the tests can shorten it. The production value is
// config.DefaultMemoryFoldOnDoneBound, whose value and ceiling
// AC-MFB-008 (x) asserts by name.
var memoryFoldOnDoneBound = config.DefaultMemoryFoldOnDoneBound

// foldClosedCardMemoryFn is the seam the close paths call. Production points
// it at foldClosedCardMemory; the tests replace it for the
// disabled-by-construction arm of AC-MFB-008 (i).
var foldClosedCardMemoryFn = foldClosedCardMemory

// foldOnDoneStepFn is the bounded step body, replaceable so tests can seed a
// panic inside the recover wrapper (AC-MFB-008 (vi)) without touching a
// store. The second argument is the run that owns the step: its abandoned
// flag expires the step's writes, its temp reference lets the caller's
// timeout branch recover the in-flight temp file.
var foldOnDoneStepFn = foldOnDoneStep

// foldOnDoneRecorder records the candidate store directories the step
// resolved and the store files it opened (AC-MFB-008 (v), (viii), (xi)).
// Production leaves the pointer nil; only the wiring tests set it. The
// granularity is the wiring's own file operations at this layer — the gate
// is checked before ANY of them, so a gate-off close records nothing at all.
type foldOnDoneRecorder struct {
	mu     sync.Mutex
	stores []string
	opens  []string
}

func (r *foldOnDoneRecorder) storeDir(dir string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stores = append(r.stores, dir)
}

func (r *foldOnDoneRecorder) fileOpen(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.opens = append(r.opens, path)
}

func (r *foldOnDoneRecorder) snapshot() (stores, opens []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.stores...), append([]string(nil), r.opens...)
}

// memoryFoldOnDoneRec is the active recorder, nil outside the wiring tests.
var memoryFoldOnDoneRec *foldOnDoneRecorder

// foldOnDoneExit, when non-nil, is closed by the bounded step's goroutine
// when it returns — including after a deadline-expired abandonment — so a
// test can synchronize on the worker's exit before restoring test globals
// the step reads (memoryFoldSeam). Production leaves it nil.
var foldOnDoneExit chan struct{}

// foldOnDoneGateOpen reads config.EnvMemoryFoldOnDone with OD-2's accepted
// values — "1" and "true", case-insensitively, trimmed; everything else
// (unset, empty, "0", "false", "yes") is off. The same vocabulary as the
// repository's existing environment-flag parse (internal/session envTruthy).
func foldOnDoneGateOpen() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(config.EnvMemoryFoldOnDone))) {
	case "1", "true":
		return true
	}
	return false
}

// foldClosedCardMemory folds one closed card's STRONG index lines (REQ-MFB-007).
// The gate is read FIRST — a closed gate touches no store and opens no file
// (AC-MFB-008 (viii), plan §J). With the gate open the fold runs on the
// bounded wait of memoryFoldOnDoneBound: a step that does not finish in time
// is abandoned with no further write (B14/P11), a panicking step is
// recovered, and every outcome is reported as at most ONE stderr line.
func foldClosedCardMemory(cardID string) {
	if !foldOnDoneGateOpen() {
		return
	}
	// run is THIS card's own state: abandoned flips when the bound expires
	// (the running step checks it before every phase and the write path
	// before every write), and temp carries the in-flight temp file's path
	// for the timeout branch to recover.
	run := &foldOnDoneRun{}
	done := make(chan foldOnDoneOutcome, 1)
	exit := foldOnDoneExit // captured at call time: the NEXT fixture may replace the global while this worker still runs
	go func() {
		// Registered first, so it runs LAST: the outcome lands on done
		// before the exit signal closes — a waiter on the captured exit
		// channel sees every step read complete.
		defer func() {
			if exit != nil {
				close(exit)
			}
		}()
		summary, err := func() (summary string, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("the fold step panicked: %v", r)
				}
			}()
			return foldOnDoneStepFn(cardID, run)
		}()
		done <- foldOnDoneOutcome{summary: summary, err: err}
	}()
	select {
	case out := <-done:
		foldOnDoneReport(cardID, out)
	case <-time.After(memoryFoldOnDoneBound):
		run.abandoned.Store(true)
		// Recover the in-flight temp file (codex-review round 2): the
		// abandoned worker may be parked on a read while holding a
		// fully-written .moai-fold-*.tmp, and a one-shot CLI process exits
		// before the worker's deferred removal runs. The path is taken
		// from THIS run's own state — a later card's worker registering
		// its own temp file can never clobber an earlier run's recovery
		// (the sequential-ownership regression). A remove that races a
		// rename is a harmless ENOENT, and removing a temp file can never
		// corrupt the store — no reader path touches it.
		if p := run.temp.take(); p != "" {
			_ = os.Remove(p)
		}
		foldOnDoneReport(cardID, foldOnDoneOutcome{
			err: fmt.Errorf("abandoned after %s — the store did not answer in time; the step will begin no write", memoryFoldOnDoneBound),
		})
	}
}

// foldOnDoneRun is one bounded fold step's own state: the abandonment flag
// the caller's timeout sets and the step checks before every write, and the
// per-worker temp-file reference the timeout branch recovers. Per-worker —
// never a shared global — because an auto-done close of N cards runs N
// workers whose lifetimes overlap: a late finisher clearing a shared path
// would clobber the next card's recoverable temp (codex-review round 2,
// sequential-ownership regression).
type foldOnDoneRun struct {
	abandoned atomic.Bool
	temp      foldTempRef
}

// forbidden reports whether the run's bound has expired.
func (r *foldOnDoneRun) forbidden() bool { return r.abandoned.Load() }

// foldTempRef holds at most one in-flight temp file path with occupancy
// semantics: take() atomically empties it, clearIf only clears a path it
// still holds.
type foldTempRef struct {
	mu   sync.Mutex
	path string
}

func (r *foldTempRef) set(p string) {
	r.mu.Lock()
	r.path = p
	r.mu.Unlock()
}

// take atomically returns and empties the reference.
func (r *foldTempRef) take() string {
	r.mu.Lock()
	p := r.path
	r.path = ""
	r.mu.Unlock()
	return p
}

// get returns the currently held path, if any.
func (r *foldTempRef) get() string {
	r.mu.Lock()
	p := r.path
	r.mu.Unlock()
	return p
}

// clearIf empties the reference only when it still holds p — a late
// worker's cleanup must never clear a newer worker's path.
func (r *foldTempRef) clearIf(p string) {
	r.mu.Lock()
	if r.path == p {
		r.path = ""
	}
	r.mu.Unlock()
}

// foldOnDoneOutcome is one completed (or recovered) step's report.
type foldOnDoneOutcome struct {
	summary string
	err     error
}

// foldOnDoneReport emits the wiring's ONE stderr line for the step's
// outcome: the failure reason, or the success summary naming the store, the
// archive file and the line count (AC-MFB-008 (ii)). A step with nothing to
// fold stays silent.
func foldOnDoneReport(cardID string, out foldOnDoneOutcome) {
	switch {
	case out.err != nil:
		_, _ = fmt.Fprintf(foldOnDoneStderr, "memory fold-on-done: %s: %v\n", cardID, out.err)
	case out.summary != "":
		_, _ = fmt.Fprintf(foldOnDoneStderr, "memory fold-on-done: %s\n", out.summary)
	}
}

// foldOnDoneStep resolves the one store, folds cardID's lines through the
// M3 apply path (never a re-implementation), and returns the success summary
// line. An empty summary means there was nothing to fold. The recorder, when
// present, sees every candidate directory and every store-file open. The
// run's abandoned flag is checked at every phase boundary and handed to the
// apply path, which refuses each individual write once the bound has
// expired; the run's temp reference records this worker's in-flight temp
// file for the caller's timeout recovery.
func foldOnDoneStep(cardID string, run *foldOnDoneRun) (string, error) {
	forbidden := run.forbidden
	if forbidden() {
		return "", fmt.Errorf("the step was abandoned before it started")
	}
	rec := memoryFoldOnDoneRec
	candidates, err := memoryCandidateStores(resolveProjectDir())
	if err != nil {
		return "", err
	}
	var chosen *memoryStore
	for i := range candidates {
		if rec != nil {
			rec.storeDir(candidates[i].Dir)
		}
		idx := filepath.Join(candidates[i].Dir, memoryIndexName)
		if info, statErr := os.Stat(idx); statErr == nil && !info.IsDir() {
			if rec != nil {
				rec.fileOpen(idx)
			}
			chosen = &candidates[i]
			break
		}
	}
	if chosen == nil {
		return "", fmt.Errorf("no memory store with %s resolved", memoryIndexName)
	}
	if err := foldIndexGuard(chosen.Dir); err != nil {
		return "", err
	}
	if rec != nil {
		rec.fileOpen(filepath.Join(chosen.Dir, memoryIndexName))
	}
	// REQ-DISPATCH-008: the on-done step's transaction takes the store's
	// cross-process lock like the verb. The wait happens inside the
	// acquire; an abandoned step's forbidden() checks still refuse the
	// write after the wait — and the read itself is abandonment-polled
	// (run-gate finding 2): a read that blocks past the bound releases the
	// lock instead of pinning it for every follow-up fold.
	var summary string
	err = withFoldStoreLock(chosen.Dir, func() error {
		before, err := snapshotStoreBounded(chosen.Dir, forbidden)
		if err != nil {
			return err
		}
		comp, err := buildFoldPlan(before, cardID)
		if err != nil {
			return err
		}
		if len(comp.plan.Removed) == 0 {
			return nil
		}
		if forbidden() {
			return fmt.Errorf("the step was abandoned before its write")
		}
		if err := applyFold(chosen.Dir, before, comp, forbidden, run); err != nil {
			return err
		}
		summary = fmt.Sprintf("filed %d line(s) of %s into %s (store: %s)",
			len(comp.plan.Removed), cardID, comp.archive, chosen.Dir)
		return nil
	})
	if err != nil {
		return "", err
	}
	return summary, nil
}

// foldAbandonedReadPoll is how often a bounded step's blocking read checks
// the abandonment flag while it waits.
const foldAbandonedReadPoll = 100 * time.Millisecond

// snapshotStoreBounded runs SnapshotStore for a bounded on-done step,
// polling the step's abandonment flag while the read runs: a read that
// blocks — a FIFO standing in for the index, a hung filesystem — must not
// pin the store lock past the step's own bound. On abandonment the caller
// releases the lock and returns; the read's own goroutine (still blocked)
// holds nothing and its result is discarded.
func snapshotStoreBounded(dir string, forbidden func() bool) (taxonomy.StoreSnapshot, error) {
	type foldReadResult struct {
		snap taxonomy.StoreSnapshot
		err  error
	}
	ch := make(chan foldReadResult, 1)
	go func() {
		snap, err := taxonomy.SnapshotStore(dir)
		ch <- foldReadResult{snap, err}
	}()
	tick := time.NewTimer(foldAbandonedReadPoll)
	defer tick.Stop()
	for {
		select {
		case res := <-ch:
			return res.snap, res.err
		case <-tick.C:
			if forbidden() {
				return nil, fmt.Errorf("memory fold: the step was abandoned while reading the store")
			}
			tick.Reset(foldAbandonedReadPoll)
		}
	}
}
