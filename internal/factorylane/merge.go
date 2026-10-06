package factorylane

// merge.go — the lane-direct merge condition triple (SPEC-FACTORY-LANE-AUTONOMY-001
// fragment 3, design.md D3): the check sequence a lane executes BEFORE entering
// the integration window. The triple is t1241's card-text discipline, re-pinned
// at M3 pre-flight from the live queue card verbatim:
//
//	"병합 창 자동화(자율 모드: sync-audit PASS·충돌 없음·HEAD^{tree}=HEAD^2^{tree}
//	 면 로컬 develop 병합)"
//
// It is a check sequence over the EXISTING `moai integration acquire`/`release`
// window — no new serialization mechanism (REQ-FLA-010/011), no F3 controller
// machinery (spec.md §F exclusion): no write-ahead start events, no trial
// merges, no tick loop.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// The three condition names of the re-pinned triple. They are the record's
// vocabulary: a refusal names the failing condition by one of these.
const (
	// CheckSyncAudit is condition (a): the card's sync phase record
	// (progress.md §E.4) reads sync_status complete — the sync phase, whose
	// entry is gated on the sync audit, has closed.
	CheckSyncAudit = "sync-audit"
	// CheckConflictFree is condition (b): the merge carries no unresolved
	// conflict, probed mechanically with `git merge-tree --write-tree` — a
	// dry-run form, never an actual merge.
	CheckConflictFree = "conflict-free"
	// CheckTreeIdentity is condition (c): t1241's tree identity. On a merge
	// commit it is the literal `HEAD^{tree} == HEAD^2^{tree}`; before the
	// merge exists it is the same predicate evaluated on the would-be merge
	// tree (`git merge-tree --write-tree` result == the card branch's tree —
	// a clean merge commit's tree is exactly the merge-tree result, and
	// HEAD^2 is the merged branch, so the two forms coincide).
	CheckTreeIdentity = "tree-identity"
)

// GitExitError reports a git invocation that exited non-zero, carrying the
// exit code and both streams. `git merge-tree` uses exit 1 to report
// conflicts — that is a check RESULT the triple reads, not a tool failure —
// so the error must expose the code rather than collapse into a message.
type GitExitError struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Error renders the exit code with git's own stderr for the caller's log.
func (e *GitExitError) Error() string {
	msg := strings.TrimSpace(e.Stderr)
	if msg == "" {
		msg = strings.TrimSpace(e.Stdout)
	}
	if msg == "" {
		return fmt.Sprintf("git exited %d", e.ExitCode)
	}
	return fmt.Sprintf("git exited %d: %s", e.ExitCode, msg)
}

// GitRunner executes one git invocation in the repository the check probes
// and returns its standard output. It is the seam that keeps the triple
// evaluation hermetic: unit tests script it, the CLI wires the real runner.
type GitRunner interface {
	Git(args ...string) (string, error)
}

// ExecGitRunner is the real GitRunner: it runs git in Dir.
type ExecGitRunner struct {
	Dir string
}

// Git implements GitRunner with os/exec. A non-zero exit becomes a
// *GitExitError carrying both streams; other failures pass through verbatim.
func (r ExecGitRunner) Git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		code := 1
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		}
		return stdout.String(), &GitExitError{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}
	}
	return stdout.String(), nil
}

// MergeTripleInput carries what one triple evaluation needs: the card's SPEC
// directory (the sync phase record lives in its progress.md), the card branch
// and the integration branch the merge lands on, the repository the probes
// run in, and — when one already exists — the merge commit the literal
// tree-identity form evaluates.
type MergeTripleInput struct {
	Lane        string
	Card        string
	SpecDir     string // the card's .moai/specs/<SPEC-ID> directory; empty = a SPEC-less card, which reads RepoDir's .moai/reports/<Card>/verdict.md instead
	Branch      string // the card branch (merge source)
	Develop     string // the integration branch (merge target)
	RepoDir     string // the git repository the probes run in
	MergeCommit string // optional: a prepared merge commit for the literal tree-identity form
}

// MergeCheck is one condition's recorded verdict: the condition name, whether
// it passed, and the check output verbatim (or the named failure).
type MergeCheck struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

// MergeCheckRun is one recorded run of the triple. One file per run, under
// the factory-fallback store's merge-checks/<lane>/ directory, following the
// one-record-per-file store convention. The checked-at stamp is what
// AC-FLA-009's second half verifies against the integration acquire
// timestamp: the recorded check output for all three exists BEFORE the
// window was taken.
type MergeCheckRun struct {
	Lane            string       `json:"lane"`
	Card            string       `json:"card"`
	Branch          string       `json:"branch"`
	Develop         string       `json:"develop"`
	Checks          []MergeCheck `json:"checks"`
	AllPassed       bool         `json:"all_passed"`
	FailedCondition string       `json:"failed_condition,omitempty"`
	CheckedAt       time.Time    `json:"checked_at"`
}

// EvaluateMergeTriple executes the check sequence and returns the run
// record — every condition evaluated and its output recorded, all_passed set,
// and failed_condition naming the FIRST failing condition when any failed.
// It never performs a merge: the conflict probe is the merge-tree dry-run
// form, and every git invocation here is read-only.
//
// @MX:ANCHOR: [AUTO] the lane-direct merge condition triple (AC-FLA-009) funnels through here
// @MX:REASON: the merge-readiness verb, the gate verb, and the M5 cross-fragment tests all consume this sequence; a condition silently skipped here would let an unready merge reach the window.
// @MX:SPEC: SPEC-FACTORY-LANE-AUTONOMY-001
func EvaluateMergeTriple(in MergeTripleInput, git GitRunner) (MergeCheckRun, error) {
	run := MergeCheckRun{Lane: in.Lane, Card: in.Card, Branch: in.Branch, Develop: in.Develop}

	// (a) sync-audit: the card's sync phase record reads closed.
	// A SPEC-less card (empty SpecDir) closes on its verdict file instead
	// (merge_specless.go); a card with a SPEC reads its progress.md as before.
	var syncDetail string
	var syncOK bool
	if in.SpecDir == "" {
		syncDetail, syncOK = checkSyncAuditVerdictFile(git, in.RepoDir, in.Card)
	} else {
		syncDetail, syncOK = checkSyncAudit(in.SpecDir)
	}
	run.Checks = append(run.Checks, MergeCheck{Name: CheckSyncAudit, Passed: syncOK, Detail: syncDetail})

	// (b) conflict-free and (c) tree-identity share one merge-tree probe in
	// the pre-merge form; a prepared merge commit switches (c) to t1241's
	// literal rev-parse form instead.
	conflictDetail, conflictOK := checkConflictFree(git, in.Develop, in.Branch)
	run.Checks = append(run.Checks, MergeCheck{Name: CheckConflictFree, Passed: conflictOK, Detail: conflictDetail})

	var treeDetail string
	var treeOK bool
	if in.MergeCommit != "" {
		treeOK, treeDetail = checkTreeIdentityOnMergeCommit(git, in.MergeCommit)
	} else {
		treeOK, treeDetail = checkTreeIdentityPreMerge(git, in.Develop, in.Branch)
	}
	run.Checks = append(run.Checks, MergeCheck{Name: CheckTreeIdentity, Passed: treeOK, Detail: treeDetail})

	run.AllPassed = syncOK && conflictOK && treeOK
	for _, chk := range run.Checks {
		if !chk.Passed {
			run.FailedCondition = chk.Name
			break
		}
	}
	return run, nil
}

// checkSyncAudit reads the card's sync phase record: progress.md §E.4's
// sync_status value. The first whitespace token decides — the annotated
// `complete (3-phase close …)` spelling passes, `audit-ready` does not. A
// missing record is a failure, not a pass: the sync phase has not closed.
func checkSyncAudit(specDir string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(specDir, "progress.md"))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("no progress.md under %s — the sync phase record does not exist", specDir), false
		}
		return fmt.Sprintf("read progress.md: %v", err), false
	}
	status, found := syncStatusFromE4(string(data))
	if !found {
		return "progress.md §E.4 Sync-phase Audit-Ready Signal not found — the sync phase has not closed", false
	}
	first := ""
	if fields := strings.Fields(status); len(fields) > 0 {
		first = fields[0]
	}
	if first != "complete" {
		return fmt.Sprintf("§E.4 sync_status is %q, want complete — the sync phase record does not read closed", status), false
	}
	return fmt.Sprintf("§E.4 sync_status: %s — the sync phase record reads closed", status), true
}

// syncStatusFromE4 extracts the sync_status line from the §E.4 section of a
// progress.md, scoped to the section so a stray occurrence elsewhere never
// answers the question.
func syncStatusFromE4(content string) (string, bool) {
	inSection := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			inSection = strings.HasPrefix(trimmed, "## §E.4")
			continue
		}
		if inSection && strings.HasPrefix(trimmed, "sync_status:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "sync_status:")), true
		}
	}
	return "", false
}

// checkConflictFree evaluates condition (b): the merge-tree dry-run. Exit 0
// is clean; exit 1 is git's conflict verdict — a check RESULT, recorded with
// the conflicted paths; any other failure refuses fail-closed, because a
// check that cannot run is not a pass.
func checkConflictFree(git GitRunner, developRef, branch string) (string, bool) {
	out, err := git.Git("merge-tree", "--write-tree", developRef, branch)
	if err == nil {
		return fmt.Sprintf("merge-tree %s + %s clean; result tree %s", developRef, branch, strings.TrimSpace(out)), true
	}
	var exitErr *GitExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode == 1 {
		return fmt.Sprintf("merge-tree %s + %s reports unresolved conflicts:\n%s", developRef, branch, strings.TrimSpace(exitErr.Stdout)), false
	}
	return fmt.Sprintf("merge-tree %s + %s could not run: %v — fail-closed", developRef, branch, err), false
}

// checkTreeIdentityPreMerge evaluates condition (c) before the merge exists:
// the merge-tree result tree equals the card branch's tree. This is the
// pre-merge form of t1241's literal HEAD^{tree} == HEAD^2^{tree} — a clean
// merge commit's tree is exactly the merge-tree result, and HEAD^2 is the
// merged branch, so the two forms decide identically (verified on a real
// repository in TestMergeTriple_PreMergeFormMatchesLiteralPostMergeForm).
func checkTreeIdentityPreMerge(git GitRunner, developRef, branch string) (bool, string) {
	out, err := git.Git("merge-tree", "--write-tree", developRef, branch)
	if err != nil {
		return false, fmt.Sprintf("merge-tree %s + %s could not run: %v — tree identity cannot be evaluated (fail-closed)", developRef, branch, err)
	}
	mergeTree := strings.TrimSpace(out)
	branchOut, err := git.Git("rev-parse", branch+"^{tree}")
	if err != nil {
		return false, fmt.Sprintf("rev-parse %s^{{tree}} failed: %v", branch, err)
	}
	branchTree := strings.TrimSpace(branchOut)
	if mergeTree != branchTree {
		return false, fmt.Sprintf("merge result tree %s != card branch tree %s — the merge would carry changes beyond the branch (pre-merge form of HEAD^{{tree}} == HEAD^2^{{tree}})", mergeTree, branchTree)
	}
	return true, fmt.Sprintf("merge result tree %s == card branch tree %s (HEAD^{{tree}} == HEAD^2^{{tree}} holds)", mergeTree, branchTree)
}

// checkTreeIdentityOnMergeCommit evaluates condition (c) in t1241's literal
// form on an existing merge commit, through rev-parse: HEAD^{tree} vs
// HEAD^2^{tree}.
func checkTreeIdentityOnMergeCommit(git GitRunner, mergeCommit string) (bool, string) {
	headOut, err := git.Git("rev-parse", mergeCommit+"^{tree}")
	if err != nil {
		return false, fmt.Sprintf("rev-parse %s^{{tree}} failed: %v", mergeCommit, err)
	}
	secondOut, err := git.Git("rev-parse", mergeCommit+"^2^{tree}")
	if err != nil {
		return false, fmt.Sprintf("%s has no second parent — not a merge commit; the literal tree-identity form cannot be evaluated: %v", mergeCommit, err)
	}
	headTree, secondTree := strings.TrimSpace(headOut), strings.TrimSpace(secondOut)
	if headTree != secondTree {
		return false, fmt.Sprintf("HEAD^{tree} %s != HEAD^2^{tree} %s on %s — the merge commit carries changes beyond the merged branch", headTree, secondTree, mergeCommit)
	}
	return true, fmt.Sprintf("HEAD^{tree} %s == HEAD^2^{tree} on %s", headTree, mergeCommit)
}

// VerifyRunBeforeAcquire reports whether a recorded run satisfies AC-FLA-009's
// second half: all three checks passed, and the run was recorded BEFORE the
// integration acquire timestamp. The refusal names which half is missing.
func VerifyRunBeforeAcquire(run *MergeCheckRun, acquireAt time.Time) (bool, string) {
	if run == nil {
		return false, "no recorded merge check run for this lane and card — run the check sequence first"
	}
	if !run.AllPassed {
		return false, fmt.Sprintf("the recorded run failed condition %q", run.FailedCondition)
	}
	if run.CheckedAt.IsZero() {
		return false, "the recorded run carries no timestamp"
	}
	if !run.CheckedAt.Before(acquireAt) {
		return false, fmt.Sprintf("recorded at %s, not before the window acquire at %s",
			run.CheckedAt.Format(time.RFC3339), acquireAt.Format(time.RFC3339))
	}
	return true, fmt.Sprintf("three checks recorded at %s, before the window acquire at %s",
		run.CheckedAt.Format(time.RFC3339), acquireAt.Format(time.RFC3339))
}

// VerifyWhyText is the cleared verdict's standing explanation: what the
// recorded run plus the window's acquire stamp together prove.
const VerifyWhyText = "the recorded check output for all three conditions exists before the integration acquire timestamp (AC-FLA-009)"

// WindowSnapshot is one read of the integration window record, flattened to
// what the AC-FLA-011 predicate decides on. The CLI flattens the
// factory.IntegrationLock into it; tests forge it directly.
type WindowSnapshot struct {
	Held       bool // a holder is recorded at all
	Live       bool // the recorded holder's session is alive
	HolderName string
	AcquiredAt time.Time
	KnownAt    bool // the record carried a parseable acquire timestamp
}

// WindowCoversMerge is REQ-FLA-011 as a predicate: a lane-direct merge may
// proceed only when a LIVE acquire record for this lane covers the merge
// moment — the window taken before the merge and not yet released. The
// refusal names which property is missing, so the negative case is always
// attributable. This is a checked property of the EXISTING window — the
// second serialization mechanism the SPEC forbids stays unbuilt.
func WindowCoversMerge(w WindowSnapshot, lane string, at time.Time) (bool, string) {
	switch {
	case !w.Held:
		return false, "no integration acquire record exists — the window was never taken"
	case !w.Live:
		return false, "the window record names a holder whose session is gone (stale) — not a live hold"
	case lane != "" && w.HolderName != "" && w.HolderName != lane:
		return false, fmt.Sprintf("the window is held by %s, not this lane", w.HolderName)
	case !w.KnownAt:
		return false, "the acquire record carries no parseable acquired-at timestamp — it cannot cover any moment"
	case w.AcquiredAt.After(at):
		return false, fmt.Sprintf("the window was acquired at %s, after the merge moment %s",
			w.AcquiredAt.Format(time.RFC3339), at.Format(time.RFC3339))
	}
	return true, fmt.Sprintf("a live acquire record for lane %s covers the moment (held since %s)",
		lane, w.AcquiredAt.Format(time.RFC3339))
}

// mergeCheckDir is the directory holding one lane's recorded triple runs.
func (s *Store) mergeCheckDir(lane string) string {
	return filepath.Join(s.root, "merge-checks", lane)
}

// RecordMergeCheckRun persists one triple run — one file per run, stamped
// with the store clock. The filename stamp bumps on a collision (the
// transitions precedent); the record's CheckedAt always carries the true
// check instant.
//
// @MX:NOTE: [AUTO] the pre-acquire timestamp on this record is what AC-FLA-009's second half verifies against the integration lock's acquired-at — deleting or rewriting runs would erase that proof.
// @MX:SPEC: SPEC-FACTORY-LANE-AUTONOMY-001
func (s *Store) RecordMergeCheckRun(run MergeCheckRun) (MergeCheckRun, error) {
	run.CheckedAt = s.clock.Now().UTC()
	data, err := marshalRecord(run)
	if err != nil {
		return MergeCheckRun{}, err
	}
	dir := s.mergeCheckDir(run.Lane)
	stamp := run.CheckedAt
	for attempt := 0; attempt < 8; attempt++ {
		path := filepath.Join(dir, fmt.Sprintf("chk-%019d.json", stamp.UnixNano()))
		if _, statErr := os.Stat(path); statErr == nil {
			stamp = stamp.Add(time.Nanosecond)
			continue
		} else if !os.IsNotExist(statErr) {
			return MergeCheckRun{}, fmt.Errorf("factorylane: stat merge-check %s: %w", path, statErr)
		}
		return run, writeFileAtomic(path, data)
	}
	return MergeCheckRun{}, fmt.Errorf("factorylane: merge-check filename collisions exhausted for lane %s", run.Lane)
}

// LatestMergeCheckRun returns the newest recorded run for lane and card, or
// nil when none exists. An unreadable record is an error, never a silent
// skip: a run nobody can read cannot vouch for a merge.
func (s *Store) LatestMergeCheckRun(lane, card string) (*MergeCheckRun, error) {
	dir := s.mergeCheckDir(lane)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("factorylane: read merge-checks dir %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "chk-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, e.Name())
	}
	// Newest first: the zero-padded instant orders lexically.
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("factorylane: read merge-check %s: %w", name, err)
		}
		var run MergeCheckRun
		if err := json.Unmarshal(data, &run); err != nil {
			return nil, fmt.Errorf("factorylane: parse merge-check %s: %w", name, err)
		}
		if run.Card == card {
			return &run, nil
		}
	}
	return nil, nil
}
