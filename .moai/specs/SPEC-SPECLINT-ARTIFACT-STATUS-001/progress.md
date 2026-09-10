# progress.md — SPEC-SPECLINT-ARTIFACT-STATUS-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
status: draft
tier: S
artifacts: [spec.md, plan.md]
spec_id: SPEC-SPECLINT-ARTIFACT-STATUS-001
card: t490
baseline: 615d18c1f
red_baseline: .moai/reports/t490/red-baseline.md
measured: 2026-09-06
```

## §E.2 Run-phase Evidence

M2 edit applied (2026-09-07, by run-phase delegation, worktree `t490`, branch `WT-speclint-status-transition`, base HEAD `a957d8308`):

- `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md` — deleted frontmatter line 5 (`status: completed`). All other frontmatter fields (`id`, `title`, `version`, `created`, `updated`, `author`, `tier`) remain byte-identical (REQ-002).
- `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/acceptance.md` — deleted frontmatter line 5 (`status: completed`). Same field-preservation guarantee (REQ-002).

Rationale (Artifact Statelessness, `spec-frontmatter-schema.md` § Artifact Statelessness): non-spec.md SPEC artifacts are stateless on the status axis; the lifecycle status lives in `spec.md` alone. The lint error message (`ArtifactStatusFieldForbidden`) prescribes exactly this removal.

Pre-flight: blob pins at edit time matched SPEC §4 (`plan.md` = `15d5ef6ae478b97fc301f9c778e472aa7bd6c8d8`, `acceptance.md` = `13d830137cd34b4fcb3616a92f0099e4880f2d54`, both via `git rev-parse HEAD:<path>`). Post-edit `git diff --stat`: 2 files changed, 2 deletions(-) — scope discipline confirmed (REQ-003; the lane re-verifies per AC-SCOPE).

Intended verification command (lane owns execution; verdict on `^ERROR` match count, never exit code — REQ-006):

```
go run ./cmd/moai spec lint --strict > <output-file>   # exit code captured separately
grep -c '^ERROR' <output-file>                          # expect 0 (AC-GREEN)
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
phase: run
run_status: PASS
spec_id: SPEC-SPECLINT-ARTIFACT-STATUS-001
card: t490
measured_at_tree: a75041e37   # the fix commit
verdict_artifact: .moai/reports/t490/verdict.md
measurer: lane-15 session — performed ALL measurements (Runs A-D) and owns all commits; this agent performed only the two-line edit (§E.2)
ac_pass_count: 5
ac_fail_count: 0
```

AC verdicts (each judged in `.moai/reports/t490/verdict.md`; all measured by lane-15):

- **AC-GREEN** — Run B on the fixed tree: `^ERROR` count 0 (4,333 warnings remain; exit 1 EXPECTED, not the verdict).
- **AC-MUTANT** — Run C with the two base blobs restored: exactly 2 ERRORs, both `ArtifactStatusFieldForbidden`; Run D after restore: 0.
- **AC-ORDERING** — RED baseline commit `a957d8308` precedes fix commit `a75041e37`.
- **AC-SCOPE** — fix commit carries exactly the two artifacts, two deletions total.
- **AC-EVIDENCE** — evidence committed under `.moai/reports/t490/`; every branch commit carries `t490`.

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: b58000784
sync_status: completed

Close summary: two-line repair shipped — removed the forbidden `status: completed` field from the `SPEC-CODEX-E2E-MEASURE-001` sibling artifacts (`plan.md` and `acceptance.md` frontmatter), the only two ERROR-severity spec-lint findings (`ArtifactStatusFieldForbidden`). Single-step `draft → completed` transition carried by this close commit (the run phase never committed an intermediate state — intended shape for this card). Evidence: `.moai/reports/t490/verdict.md`, progress.md §E.2/§E.3, plan-audit verdict `.moai/reports/t490/plan-audit.md` (PASS 0.94, delta-confirmed under operator exception). `sync_commit_sha` is the canonical `pending-backfill-sync` placeholder per the D3 backfill exemption (spec-frontmatter-schema.md § SHA placeholder backfill exemption); the lane backfills the real SHA in a follow-up commit.

CHANGELOG skip rationale: no CHANGELOG entry is emitted for this SPEC. The repair is a two-line internal SPEC-metadata fix (frontmatter field removal in a sibling SPEC's plan/acceptance artifacts) with no user-facing behavior change, so it does not meet the CHANGELOG emission bar. This note is the decision record (lane judgment, precedent t487/t472/t489).
