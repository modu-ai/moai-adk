package spec

import (
	"strings"
	"testing"
)

// Card t1111 — SyncStatusDrift must not fire on a SPEC that is legitimately under
// in-place amendment. The frontmatter SSOT sanctions `completed → in-progress
// (amendment)`, declared by `amendment_of:` plus a HISTORY Amendments record. Such
// a SPEC still carries the prior close's §E.2 + §E.4 + sync_commit_sha, so the
// sync-evidence predicate alone cannot tell it from genuine drift.

// amendmentProgressMD carries the full sync-complete evidence the prior close left.
const amendmentProgressMD = "## §E.2 Run-phase Evidence\n\nrun rows\n\n" +
	"## §E.4 Sync-phase Audit-Ready Signal\n\nsync_commit_sha: 0e2377323\n"

// makeAmendmentSpecMD builds a spec.md shaped like a live amendment. amendmentOf
// empty omits the field; heading empty omits the HISTORY Amendments record.
func makeAmendmentSpecMD(id, status, amendmentOf, heading string) string {
	fm := "---\nid: " + id + "\n" +
		"title: \"" + id + " title\"\n" +
		"version: \"0.2.0\"\n" +
		"status: " + status + "\n" +
		"created: 2026-09-22\n" +
		"updated: 2026-09-23\n" +
		"author: manager-spec\n" +
		"priority: P2\n" +
		"phase: \"v3.2.0 target\"\n" +
		"module: \"internal/web\"\n" +
		"lifecycle: spec-anchored\n" +
		"tags: \"web\"\n" +
		"era: V3R6\n" +
		"tier: M\n"
	if amendmentOf != "" {
		fm += "amendment_of: " + amendmentOf + "\n"
	}
	fm += "---\n\n# " + id + "\n\n## HISTORY\n\n" +
		"| 날짜 | 버전 | 변경 | 주체 |\n|---|---|---|---|\n" +
		"| 2026-09-22 | 0.1.0 | plan | manager-spec |\n" +
		"| 2026-09-23 | 0.2.0 | in-place amendment | manager-spec |\n\n"
	// heading empty still writes the record rows (citing the sync SHA) but with no
	// Amendments heading — a record the SSOT does not recognise, so the negative
	// controls stay discriminating even for the SHA-citation leg.
	if heading != "" {
		fm += heading + "\n\n"
	}
	fm += "| 항목 | 값 |\n|---|---|\n" +
		"| prior completed version | 0.1.0 |\n" +
		"| prior_completed_sha | `0e2377323` |\n" +
		"| rationale | extend the guard |\n" +
		"| scope | probe + driver only |\n\n"
	fm += "---\n\n## §A 배경\n\nbody\n"
	return fm
}

func syncStatusDriftFor(t *testing.T, id, specMD string) *DriftFinding {
	t.Helper()
	baseDir := buildAuditFixture(t, []auditFixtureSpec{{id: id, specMD: specMD, progressMD: amendmentProgressMD}})
	result, err := Audit(AuditOptions{BaseDir: baseDir, FilterEra: "V3R6"})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}
	for i := range result.DriftFindings {
		f := result.DriftFindings[i]
		if f.SpecID == id && f.FindingType == FindingSyncStatusDrift {
			return &f
		}
	}
	return nil
}

// A valid amendment (status in-progress + amendment_of + HISTORY Amendments
// record) must NOT be reported as SyncStatusDrift. Both heading depths are
// covered: the SSOT names `## Amendments`; the real instance
// (SPEC-APPJS-FIRE-GUARD-001) nests it as `### Amendments` under HISTORY.
func TestAudit_SyncStatusDrift_ValidAmendmentExempt(t *testing.T) {
	t.Parallel()
	for _, heading := range []string{"### Amendments", "## Amendments"} {
		heading := heading
		t.Run(heading, func(t *testing.T) {
			t.Parallel()
			id := "SPEC-V3R6-AMEND-001"
			if f := syncStatusDriftFor(t, id, makeAmendmentSpecMD(id, "in-progress", id, heading)); f != nil {
				t.Errorf("SyncStatusDrift fired on a valid in-place amendment (remediation %q would revert it): %+v",
					f.Remediation, *f)
			}
		})
	}
}

// Negative controls: the exemption must not swallow genuine drift. Each case
// carries the same sync-complete evidence and must still be MUST-FIX.
func TestAudit_SyncStatusDrift_NotAValidAmendmentStillFires(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, status, amendmentOf, heading string
	}{
		{"in-progress without amendment_of", "in-progress", "", "### Amendments"},
		{"stray amendment_of on implemented", "implemented", "SPEC-V3R6-DRIFT-AM-001", "### Amendments"},
		{"amendment_of without Amendments record", "in-progress", "SPEC-V3R6-DRIFT-AM-001", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			id := "SPEC-V3R6-DRIFT-AM-001"
			f := syncStatusDriftFor(t, id, makeAmendmentSpecMD(id, tc.status, tc.amendmentOf, tc.heading))
			if f == nil {
				t.Fatalf("SyncStatusDrift not emitted for %s — the amendment exemption swallowed genuine drift", tc.name)
			}
			if f.Severity != "MUST-FIX" {
				t.Errorf("severity = %q, want MUST-FIX", f.Severity)
			}
		})
	}
}

// progressWithSyncSHA carries sync-complete evidence whose §E.4 sync_commit_sha
// is the given value.
func progressWithSyncSHA(sha string) string {
	return "## §E.2 Run-phase Evidence\n\nrun rows\n\n" +
		"## §E.4 Sync-phase Audit-Ready Signal\n\nsync_commit_sha: " + sha + "\n"
}

func syncStatusDriftForProgress(t *testing.T, id, specMD, progressMD string) *DriftFinding {
	t.Helper()
	baseDir := buildAuditFixture(t, []auditFixtureSpec{{id: id, specMD: specMD, progressMD: progressMD}})
	result, err := Audit(AuditOptions{BaseDir: baseDir, FilterEra: "V3R6"})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}
	for i := range result.DriftFindings {
		f := result.DriftFindings[i]
		if f.SpecID == id && f.FindingType == FindingSyncStatusDrift {
			return &f
		}
	}
	return nil
}

func requireMustFixSyncStatusDrift(t *testing.T, f *DriftFinding, what string) {
	t.Helper()
	if f == nil {
		t.Fatalf("SyncStatusDrift not emitted for %s — the amendment exemption swallowed genuine drift", what)
	}
	if f.Severity != "MUST-FIX" {
		t.Errorf("severity = %q, want MUST-FIX", f.Severity)
	}
}

// F1/P2 — an amendment that was itself re-synced (new sync_commit_sha in §E.4)
// but whose status transition was skipped is genuine drift: the §E.4 evidence is
// no longer the prior close's, and amendment_of + the Amendments record persist
// after an amendment closes, so they alone must not exempt it.
func TestAudit_SyncStatusDrift_AmendmentResyncedStillFires(t *testing.T) {
	t.Parallel()
	id := "SPEC-V3R6-AMEND-RESYNC-001"
	// The Amendments section cites prior_completed_sha 0e2377323; §E.4 carries a new SHA.
	specMD := makeAmendmentSpecMD(id, "in-progress", id, "### Amendments")
	f := syncStatusDriftForProgress(t, id, specMD, progressWithSyncSHA("7f3c1a9b2"))
	requireMustFixSyncStatusDrift(t, f, "an amendment re-synced with a new sync_commit_sha")
}

// F1/P1 — a successor amendment (amendment_of names the parent) has no prior
// close of its own; its Amendments record cites the PARENT's sha. Sync complete
// + in-progress is therefore always genuine drift.
func TestAudit_SyncStatusDrift_SuccessorAmendmentStillFires(t *testing.T) {
	t.Parallel()
	id := "SPEC-V3R6-AMEND-SUCC-002"
	specMD := makeAmendmentSpecMD(id, "in-progress", "SPEC-V3R6-AMEND-SUCC-001", "### Amendments")
	f := syncStatusDriftForProgress(t, id, specMD, progressWithSyncSHA("a1b2c3d4e"))
	requireMustFixSyncStatusDrift(t, f, "a successor amendment with its own completed sync")
}

// F1 section bound — the new sync SHA cited OUTSIDE the Amendments section (a
// HISTORY row, a later section) must not satisfy the citation check.
func TestAudit_SyncStatusDrift_SyncSHACitedOutsideAmendmentsStillFires(t *testing.T) {
	t.Parallel()
	id := "SPEC-V3R6-AMEND-BOUND-001"
	specMD := makeAmendmentSpecMD(id, "in-progress", id, "### Amendments")
	specMD = strings.Replace(specMD, "| in-place amendment | manager-spec |",
		"| in-place amendment, re-synced at `7f3c1a9b2` | manager-spec |", 1)
	specMD += "\n## §G Notes\n\nre-sync commit 7f3c1a9b2\n"
	f := syncStatusDriftForProgress(t, id, specMD, progressWithSyncSHA("7f3c1a9b2"))
	requireMustFixSyncStatusDrift(t, f, "a sync SHA cited only outside the Amendments section")
}

// F2 — an Amendments heading that exists only inside a fenced code block is not
// an Amendments record, even when the fence carries the matching SHA.
func TestAudit_SyncStatusDrift_FencedAmendmentsHeadingStillFires(t *testing.T) {
	t.Parallel()
	id := "SPEC-V3R6-AMEND-FENCE-001"
	specMD := makeAmendmentSpecMD(id, "in-progress", id, "")
	specMD += "\n## §H Example\n\n```markdown\n### Amendments\n\n| prior_completed_sha | `0e2377323` |\n```\n"
	f := syncStatusDriftForProgress(t, id, specMD, progressWithSyncSHA("0e2377323"))
	requireMustFixSyncStatusDrift(t, f, "an Amendments heading only inside a fenced code block")
}
