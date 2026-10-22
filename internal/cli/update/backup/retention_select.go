package backup

import (
	"sort"

	"github.com/modu-ai/moai-adk/internal/config"
)

// retention_select.go — the selection step of backup retention
// (SPEC-UPDATE-BACKUP-RETENTION-001, design D2). It decides which excess runs a
// prune retains and which it removes, as a pure function of a census: it takes no
// filesystem handle, so the budget and cover rules are table-testable.

// Retention report vocabulary (REQ-UBR-011, design D8). A prune reports each run
// it removes and each excess run it retains; a retained run names its reason.
const (
	// RetentionActionRemoved and RetentionActionRetained are the two outcomes a
	// prune reports for an excess run.
	RetentionActionRemoved  = "removed"
	RetentionActionRetained = "retained"

	// RetentionReasonProtectsVersions: the run holds a version the per-file budget
	// protects and no other retained run holds it.
	RetentionReasonProtectsVersions = "protects-versions"
	// RetentionReasonUnclassifiable: the run holds an entry the pruner cannot read
	// or classify (REQ-UBR-008).
	RetentionReasonUnclassifiable = "unclassifiable-entry"
	// RetentionReasonRemovalFailed: the selection removed the run but the removal
	// failed; the pruner assigns this reason after the fact (design D2 step 5).
	RetentionReasonRemovalFailed = "removal-failed"
)

// versionKey identifies one backed-up file version: the file's path inside its
// backup directory together with the SHA-256 of its bytes (REQ-UBR-002).
type versionKey struct {
	path string
	sum  string
}

// censusRun is the census view of one backup run. id is opaque to the selection.
type censusRun struct {
	id string
	// keys are the regular-file versions the run holds, bookkeeping and the
	// .template-defaults render excluded.
	keys []versionKey
	// exempt marks a run holding an entry that cannot be read or classified,
	// including an entry that is not a regular file (REQ-UBR-008).
	exempt bool
	// exemptDetail names the first such entry, relative to its backup directory.
	exemptDetail string
}

// retainedRun is one excess run the selection keeps.
type retainedRun struct {
	id string
	// reason is one of the RetentionReason constants.
	reason string
	// protects counts the protected versions the run holds (reason
	// RetentionReasonProtectsVersions).
	protects int
	// detail names the unclassifiable entry (reason RetentionReasonUnclassifiable).
	detail string
}

// retentionPlan is the outcome of the selection for the excess runs, in the
// order they were given (newest first).
type retentionPlan struct {
	remove []string
	retain []retainedRun
	// protected is the number of distinct protected versions: the bound on the
	// number of non-exempt retained excess runs (REQ-UBR-010).
	protected int
}

// selectRetention returns the plan for runs, given newest first, with keepCount
// runs in the window and pastKeep past versions protected per file. A pastKeep
// below 1 reads as config.DefaultBackupPastVersionsKeep, so the protection of
// REQ-UBR-006 cannot be switched off by a configured value (REQ-UBR-007).
//
// The window is the first keepCount runs and is never decided here. Among the
// excess runs the selection retains, in order: every exempt run (REQ-UBR-008);
// then a greedy cover of the protected versions no exempt run holds, taking the
// run that covers the most uncovered versions and, on a tie, the newest; then it
// drops any chosen run whose protected versions the other retained runs already
// hold (REQ-UBR-005). Every other excess run is removed. keepCount 0 removes
// every run and keepCount below 0 decides nothing (REQ-UBR-014).
//
// @MX:NOTE: [AUTO] The budget (REQ-UBR-003/004), the cover and irredundancy
// rules (REQ-UBR-005) and the exempt-run disposition (REQ-UBR-008) are decided
// here and nowhere else; the pruner only gathers the census and applies the plan.
// @MX:SPEC: SPEC-UPDATE-BACKUP-RETENTION-001
func selectRetention(runs []censusRun, keepCount, pastKeep int) retentionPlan {
	if keepCount < 0 || len(runs) <= keepCount {
		return retentionPlan{}
	}
	excess := runs[keepCount:]
	if keepCount == 0 {
		plan := retentionPlan{remove: make([]string, 0, len(excess))}
		for _, r := range excess {
			plan.remove = append(plan.remove, r.id)
		}
		return plan
	}
	if pastKeep < 1 {
		pastKeep = config.DefaultBackupPastVersionsKeep
	}

	// held[j] is the set of distinct versions excess run j holds.
	held := make([]map[versionKey]struct{}, len(excess))
	for j, r := range excess {
		held[j] = make(map[versionKey]struct{}, len(r.keys))
		for _, k := range r.keys {
			held[j][k] = struct{}{}
		}
	}

	// Past versions: held by an excess run and by no window run (REQ-UBR-004),
	// each ranked by its newest excess holder.
	inWindow := make(map[versionKey]struct{})
	for _, r := range runs[:keepCount] {
		for _, k := range r.keys {
			inWindow[k] = struct{}{}
		}
	}
	newest := make(map[versionKey]int)
	for j := range excess {
		for k := range held[j] {
			if _, current := inWindow[k]; current {
				continue
			}
			if at, seen := newest[k]; !seen || j < at {
				newest[k] = j
			}
		}
	}

	// The pastKeep most recent past versions of each path are protected
	// (REQ-UBR-003); equal recency breaks on the digest so the result is stable.
	byPath := make(map[string][]versionKey)
	for k := range newest {
		byPath[k.path] = append(byPath[k.path], k)
	}
	protected := make(map[versionKey]struct{})
	for _, ks := range byPath {
		sort.Slice(ks, func(a, b int) bool {
			if newest[ks[a]] != newest[ks[b]] {
				return newest[ks[a]] < newest[ks[b]]
			}
			return ks[a].sum < ks[b].sum
		})
		for i := 0; i < len(ks) && i < pastKeep; i++ {
			protected[ks[i]] = struct{}{}
		}
	}

	// Exempt runs are retained first and hold the keys they hold.
	retain := make([]bool, len(excess))
	holders := make(map[versionKey]int) // protected version -> retained holders
	hold := func(j int) {
		for k := range held[j] {
			if _, ok := protected[k]; ok {
				holders[k]++
			}
		}
	}
	for j, r := range excess {
		if r.exempt {
			retain[j] = true
			hold(j)
		}
	}

	// Greedy cover of what no retained run holds yet; strict improvement only, so
	// a tie goes to the lowest index, the newest run.
	var chosen []int
	for {
		best, bestGain := -1, 0
		for j := range excess {
			if retain[j] {
				continue
			}
			gain := 0
			for k := range held[j] {
				if _, ok := protected[k]; ok && holders[k] == 0 {
					gain++
				}
			}
			if gain > bestGain {
				best, bestGain = j, gain
			}
		}
		if best < 0 {
			break
		}
		retain[best] = true
		chosen = append(chosen, best)
		hold(best)
	}

	// Irredundancy: drop a chosen run whose protected versions other retained
	// runs hold. Oldest first, so the newer of two equivalent runs stays.
	sort.Sort(sort.Reverse(sort.IntSlice(chosen)))
	for _, j := range chosen {
		redundant := true
		for k := range held[j] {
			if _, ok := protected[k]; ok && holders[k] < 2 {
				redundant = false
				break
			}
		}
		if !redundant {
			continue
		}
		retain[j] = false
		for k := range held[j] {
			if _, ok := protected[k]; ok {
				holders[k]--
			}
		}
	}

	plan := retentionPlan{protected: len(protected)}
	for j, r := range excess {
		if !retain[j] {
			plan.remove = append(plan.remove, r.id)
			continue
		}
		kept := retainedRun{id: r.id, reason: RetentionReasonProtectsVersions}
		for k := range held[j] {
			if _, ok := protected[k]; ok {
				kept.protects++
			}
		}
		if r.exempt {
			kept.reason = RetentionReasonUnclassifiable
			kept.detail = r.exemptDetail
		}
		plan.retain = append(plan.retain, kept)
	}
	return plan
}
