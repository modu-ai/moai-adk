package hygiene

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Decision is one candidate's classified outcome for the run report.
type Decision struct {
	Class   TargetClass       `json:"class"`
	Key     string            `json:"key"`
	Path    string            `json:"path"`
	Verdict string            `json:"verdict,omitempty"`
	Outcome string            `json:"outcome"`
	Reason  string            `json:"reason,omitempty"`
	Signal  map[string]string `json:"signal,omitempty"`
}

// GC removes finished-session state residue under the .moai root's closed
// target registry (REQ-HYG-005), gated by the fail-closed liveness verdict
// (REQ-HYG-007/008), content-recorded dating (REQ-HYG-009), symlink
// refusal scoped below the resolved root with anchored-handle actions
// (REQ-HYG-006, D27), and an action-time re-judge (D28). The unit fails
// independently of the rotator (REQ-HYG-012).
type GC struct {
	// MoaiRoot is the fully-resolved absolute path to <projectRoot>/.moai.
	MoaiRoot string
	// TranscriptRoots / RegistryPath feed the liveness evaluator.
	TranscriptRoots []string
	RegistryPath    string
	// MinAge is the minimum-age floor (production default from
	// internal/config; never a call-site literal).
	MinAge time.Duration
	// TranscriptWindow / HeartbeatWindow feed the evaluator.
	TranscriptWindow time.Duration
	HeartbeatWindow  time.Duration

	now func() time.Time
	// Liveness, when non-nil, replaces the internally built evaluator —
	// the verdict-matrix seam the GC tests drive; production leaves it
	// nil.
	Liveness *Liveness
	// preRejudge, when non-nil, runs between enumeration and the D28
	// action-time re-judge — the writer-race seam (a writer replacing the
	// file with fresh state between the two steps).
	preRejudge func()
	// preAction, when non-nil, runs between the immediately-before-action
	// component check and the anchored action — the D27 swap-after-check
	// seam.
	preAction func()
	// rootHandle, when non-nil, replaces os.OpenRoot (anchored-action seam).
	rootHandle func(string) (*os.Root, error)
}

// gnow returns the GC's clock.
func (g *GC) gnow() time.Time {
	if g.now != nil {
		return g.now()
	}
	return time.Now()
}

// Report is one GC pass's classified decisions.
type GCReport struct {
	Decisions []Decision
	Counts    map[string]int
}

// Run executes one GC pass. Report mode classifies and reports without
// touching anything (one summary row to the audit sink); apply mode
// deletes the eligible set through the anchored root handle (REQ-HYG-009,
// REQ-HYG-010).
//
// @MX:ANCHOR: [AUTO] GC.Run — the fail-closed deletion gate for state GC
// @MX:REASON: every state-file deletion passes through here; the
// DEAD+datable+aged+re-judged conjunction and the symlink/refusal rules
// are the safety property (REQ-HYG-005/006/008/009), and the report-mode
// branch must stay byte-neutral (REQ-HYG-010).
func (g *GC) Run(mode Mode) (*GCReport, error) {
	if err := validateRoot(g.MoaiRoot); err != nil {
		return nil, err
	}
	if g.MinAge <= 0 {
		return nil, fmt.Errorf("gc: non-positive minimum age (config-invalid)")
	}

	report := &GCReport{Counts: map[string]int{}}
	candidates, lockHits, unresolvable, err := enumerateCandidates(g.MoaiRoot)
	if err != nil {
		return nil, err
	}

	live := g.Liveness
	if live == nil {
		live = &Liveness{
			TranscriptWindow: g.TranscriptWindow,
			HeartbeatWindow:  g.HeartbeatWindow,
			TranscriptRoots:  g.TranscriptRoots,
			RegistryPath:     g.RegistryPath,
			now:              g.now,
		}
	}

	// Lock-class scan hits: reported excluded, never touched (REQ-HYG-011).
	for _, rel := range lockHits {
		report.record(Decision{Class: "lock", Path: rel,
			Outcome: string(OutcomeLockClassExcluded),
			Reason:  "spec-close lock class excluded from GC scope (REQ-HYG-011)"})
	}
	// Unresolvable keys: spared.
	for _, c := range unresolvable {
		report.record(Decision{Class: c.Class, Path: c.Paths[0],
			Outcome: string(OutcomeKept),
			Reason:  "name carries no session key (out of scope)"})
	}

	var anchor *os.Root
	if mode == ModeApply {
		opener := g.rootHandle
		if opener == nil {
			opener = os.OpenRoot
		}
		anchor, err = opener(g.MoaiRoot)
		if err != nil {
			return nil, fmt.Errorf("gc: anchored root handle: %w", err)
		}
		defer anchor.Close()
	}

	for i := range candidates {
		c := candidates[i]
		verdict, unmeasured, evidence := live.Evaluate(c.Key)
		_ = unmeasured

		if verdict == VerdictLive {
			report.record(g.decision(c, verdict, OutcomeKept, "live session", evidence))
			continue
		}
		if verdict == VerdictIndeterminate {
			report.record(g.decision(c, verdict, OutcomeKept,
				"indeterminate liveness — kept with reason (fail-closed)", evidence))
			continue
		}

		// DEAD: dating decides. The goal triple is gated by its .json
		// member; the verify class is dated per entry.
		if c.Class == ClassVerifyScratch {
			g.runVerifyEntries(anchor, mode, c, verdict, evidence, report)
			continue
		}
		dateTs, datable := candidateDate(g.MoaiRoot, c)
		if !datable {
			report.record(g.decision(c, verdict, OutcomeKept,
				"content-undatable — no recorded timestamp (mtime never a deletion datum)", evidence))
			continue
		}
		age := g.gnow().Sub(dateTs)
		evidence["content_date_age"] = age.String()
		if age < g.MinAge {
			report.record(g.decision(c, verdict, OutcomeKept,
				"younger than the minimum-age floor", evidence))
			continue
		}
		if mode == ModeReport {
			report.record(g.decision(c, verdict, OutcomeKept,
				"deletion eligible; report mode does not mutate", evidence))
			continue
		}
		g.applyDelete(anchor, c, verdict, evidence, report)
	}

	report.finish()
	g.writeAuditRows(mode, report)
	return report, nil
}

// decision builds a Decision for one candidate.
func (g *GC) decision(c Candidate, verdict Verdict, outcome Outcome, reason string, evidence map[string]string) Decision {
	return Decision{
		Class: c.Class, Key: c.Key, Path: c.Paths[0],
		Verdict: verdict.String(), Outcome: string(outcome),
		Reason: reason, Signal: evidence,
	}
}

// applyDelete deletes one candidate's paths through the anchored root
// handle: component check, then action, per path; the D28 action-time
// re-judge re-verifies the dating member immediately before its removal.
func (g *GC) applyDelete(anchor *os.Root, c Candidate, verdict Verdict, evidence map[string]string, report *GCReport) {
	for i, rel := range c.Paths {
		if err := g.checkComponents(rel); err != nil {
			report.record(g.decision(c, verdict, OutcomeSymlinkRefused, err.Error(), evidence))
			return
		}
		isDating := c.DatingPath != "" && rel == c.DatingPath
		if isDating {
			if g.preRejudge != nil {
				g.preRejudge()
			}
			// D28 action-time re-judge: re-stat + re-read + re-verify.
			if outcome, reason := g.rejudge(c, evidence); outcome != "" {
				report.record(g.decision(c, verdict, outcome, reason, evidence))
				return
			}
		}
		if g.preAction != nil {
			g.preAction()
		}
		if err := anchor.Remove(rel); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				report.record(g.decision(c, verdict, OutcomeAlreadyGone, rel, evidence))
				continue
			}
			report.record(g.decision(c, verdict, OutcomeError,
				fmt.Sprintf("remove %s: %v", rel, err), evidence))
			return
		}
		_ = i
	}
	report.record(g.decision(c, verdict, OutcomeDeleted, strings.Join(c.Paths, ", "), evidence))
}

// rejudge re-verifies the dating member immediately before its removal:
// the file is re-read and the datum re-parsed; an absent file is
// already-gone (skip silently by returning empty); a body whose datum no
// longer parses aborts; a datum now younger than the floor aborts — the
// writer's refreshed state survives (D28). The inode comparison rides the
// re-stat the read implies: a replaced file whose new body fails the
// predicate is kept, and a replacement whose body still reads
// aged-and-dead is the audit-sanctioned residual window (§C residual 4).
func (g *GC) rejudge(c Candidate, evidence map[string]string) (Outcome, string) {
	abs := filepath.Join(g.MoaiRoot, filepath.FromSlash(c.DatingPath))
	if _, err := os.Stat(abs); errors.Is(err, fs.ErrNotExist) {
		return OutcomeAlreadyGone, c.DatingPath
	}
	ts, ok := candidateDate(g.MoaiRoot, c)
	if !ok {
		return OutcomeRejudgedKeep, "action-time re-read: datum no longer parseable — kept"
	}
	age := g.gnow().Sub(ts)
	evidence["rejudge_content_date_age"] = age.String()
	if age < g.MinAge {
		return OutcomeRejudgedKeep, "action-time re-read: refreshed state younger than the floor — kept"
	}
	return "", ""
}

// runVerifyEntries deletes (or reports) the verify candidate's entries
// entry-by-entry; the directory itself is removed only when empty
// (REQ-HYG-005). An entry without its own recorded_at is undatable ⇒
// spared. Each mutation row carries the entry's own content-date age as
// signal evidence (REQ-HYG-009).
func (g *GC) runVerifyEntries(anchor *os.Root, mode Mode, c Candidate, verdict Verdict, evidence map[string]string, report *GCReport) {
	deletedAny := false
	for _, rel := range c.Paths {
		ts, datable := entryDate(g.MoaiRoot, rel)
		if !datable {
			report.record(g.decision(c, verdict, OutcomeKept,
				"entry content-undatable — spared ("+rel+")", evidence))
			continue
		}
		age := g.gnow().Sub(ts)
		if age < g.MinAge {
			report.record(g.decision(c, verdict, OutcomeKept,
				"entry younger than the floor ("+rel+")", evidence))
			continue
		}
		if mode == ModeReport {
			report.record(g.decision(c, verdict, OutcomeKept,
				"entry deletion eligible; report mode does not mutate ("+rel+")", evidence))
			continue
		}
		if err := g.checkComponents(rel); err != nil {
			report.record(g.decision(c, verdict, OutcomeSymlinkRefused, err.Error()+" ("+rel+")", evidence))
			continue
		}
		// D28 per-entry re-judge: re-read the entry's datum immediately
		// before removal.
		ts2, ok := entryDate(g.MoaiRoot, rel)
		if !ok {
			report.record(g.decision(c, verdict, OutcomeRejudgedKeep,
				"action-time re-read: entry datum gone or unparseable ("+rel+")", evidence))
			continue
		}
		rejudgeAge := g.gnow().Sub(ts2)
		if rejudgeAge < g.MinAge {
			report.record(g.decision(c, verdict, OutcomeRejudgedKeep,
				"action-time re-read: entry refreshed ("+rel+")", evidence))
			continue
		}
		entryEvidence := map[string]string{}
		for k, v := range evidence {
			entryEvidence[k] = v
		}
		entryEvidence["content_date_age"] = age.String()
		entryEvidence["rejudge_content_date_age"] = rejudgeAge.String()
		if g.preAction != nil {
			g.preAction()
		}
		if err := anchor.Remove(rel); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				report.record(g.decision(c, verdict, OutcomeAlreadyGone, rel, entryEvidence))
				continue
			}
			report.record(g.decision(c, verdict, OutcomeError,
				fmt.Sprintf("remove %s: %v", rel, err), entryEvidence))
			continue
		}
		deletedAny = true
		report.record(Decision{
			Class: c.Class, Key: c.Key, Path: rel,
			Verdict: verdict.String(), Outcome: string(OutcomeDeleted),
			Reason: "verify scratch entry", Signal: entryEvidence,
		})
	}
	if mode == ModeApply && deletedAny {
		// Remove the directory only when empty.
		dirRel := filepath.Dir(c.Paths[0])
		entries, err := os.ReadDir(filepath.Join(g.MoaiRoot, filepath.FromSlash(dirRel)))
		if err == nil && len(entries) == 0 {
			if err := g.checkComponents(dirRel); err == nil {
				if g.preAction != nil {
					g.preAction()
				}
				if err := anchor.Remove(dirRel); err == nil {
					report.record(g.decision(c, verdict, OutcomeDeleted,
						"empty directory removed: "+dirRel, evidence))
				}
			}
		}
	}
}

// checkComponents refuses any symlinked component strictly below the
// resolved .moai root (REQ-HYG-006): each component of the relative path
// is Lstat'd; a symlink reads symlink-refused. A missing FINAL component
// is not a refusal — the anchored action reports it as already-gone — but
// a missing INTERMEDIATE component is (a vanished parent is exactly the
// swap shape the check exists to catch). The anchored action that follows
// closes the swap window for symlink swaps (D27).
func (g *GC) checkComponents(rel string) error {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	cur := g.MoaiRoot
	for i, part := range parts {
		cur = filepath.Join(cur, filepath.FromSlash(part))
		info, err := os.Lstat(cur)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				if i == len(parts)-1 {
					return nil // final component absent: already-gone territory
				}
				return fmt.Errorf("component vanished during check: %s", rel)
			}
			return fmt.Errorf("component check failed: %s: %v", rel, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinked component below the resolved root: %s", rel)
		}
	}
	return nil
}

// writeAuditRows persists the pass's rows in the mode-appropriate
// granularity (REQ-HYG-004): report mode exactly one summary row; apply
// mode one row per action or skip.
func (g *GC) writeAuditRows(mode Mode, report *GCReport) {
	sink := auditSinkPath(filepath.Join(g.MoaiRoot, "logs"))
	var rows []AuditRow
	if mode == ModeReport {
		rows = append(rows, AuditRow{
			Unit: "gc", Mode: string(ModeReport), Outcome: OutcomeSummary,
			Counts: report.Counts,
		})
	} else {
		for _, d := range report.Decisions {
			rows = append(rows, AuditRow{
				Unit: "gc", Mode: string(ModeApply), Outcome: Outcome(d.Outcome),
				Path: d.Path, Reason: d.Reason, Signal: d.Signal,
			})
		}
	}
	if len(rows) == 0 {
		return
	}
	for i := range rows {
		if rows[i].TS == "" {
			rows[i].TS = g.gnow().UTC().Format(time.RFC3339)
		}
	}
	_ = appendAuditRowsTo(sink, rows)
}

// record appends one decision and bumps its outcome count.
func (r *GCReport) record(d Decision) {
	r.Decisions = append(r.Decisions, d)
	r.Counts[d.Outcome]++
}

// finish sorts decisions deterministically for stable reports.
func (r *GCReport) finish() {
	sort.SliceStable(r.Decisions, func(i, j int) bool {
		if r.Decisions[i].Class != r.Decisions[j].Class {
			return r.Decisions[i].Class < r.Decisions[j].Class
		}
		return r.Decisions[i].Path < r.Decisions[j].Path
	})
}

// appendAuditRowsTo writes rows to an absolute audit-sink path.
func appendAuditRowsTo(sink string, rows []AuditRow) error {
	for _, row := range rows {
		blob, err := json.Marshal(row)
		if err != nil {
			return err
		}
		if err := appendText(sink, string(blob)+"\n"); err != nil {
			return err
		}
	}
	return nil
}
