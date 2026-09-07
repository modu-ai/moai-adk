# Card Verdict — t513 (SPEC-STATUS-DRYRUN-001)

> Tree: worktree `.claude/worktrees/t513` · HEAD at writing: `c106f620d` (branch `WT-spec-status-dryrun`, base `0b1e27877`, 8 commits)
> Delivers: GitHub #1692 + #1693 — SPEC status locator anchored to YAML frontmatter (#1693, false DRIFT eliminated) and `spec status --sync-git` honoring `--dry-run` (#1692, data-loss write eliminated), plus the write-amplifier body-clobbering repair and a loud non-enum skip guard.

## Claim

1. Both reported defects are fixed and regression-guarded two-directionally (frontmatter reads normally; body `Status`/`Notes` cells — header-row AND backticked-prose variants — are never read as status, mutation-proven).
2. `--dry-run` on `--sync-git` writes nothing (byte-identical tree); the real run still updates the frontmatter; non-enum parsed statuses are skipped loudly.
3. The SPEC closed through the 3-phase lifecycle: plan-audit PASS 0.875 (iter 2/2), sync-audit PASS 95/100 (zero blocking findings).

## Evidence (this run, this tree — orchestrator's independent re-measurement; agent claims were not trusted)

- **Tests**: `go test ./internal/spec/` — all green except pre-existing `TestCatalogHashParity` (see Attribution); `go test ./internal/cli/` full package **exit 0** (measured twice: agent E3 + orchestrator background run); AC-scoped `go test -run 'TestParseStatus|TestUpdateStatus|TestValidStatuses|TestSyncGitSpecStatuses|TestSpecStatus' ./internal/spec/ ./internal/cli/` → `ok …internal/spec 0.452s` + `ok …internal/cli 3.277s` (snapshot key `4088754a4:spec-status-ac-scoped`).
- **Builds/lint**: `go build ./...` + `GOOS=windows GOARCH=amd64` exit 0; `golangci-lint run` affected pkgs → `0 issues.`
- **End-to-end (AC-009)**: fresh binary (`go build -o /tmp/t513-repro-bin/moai-fixed ./cmd/moai`) + `repro.sh` re-run → `--list` reads completed/completed/draft; drift cache has zero `"FrontmatterStatus": "Notes"`; `Summary: 2/3 SPECs have status drift` persisted (true positives by design, plan-audit settled); dry-run prints `Summary: dry-run, nothing written…` with the R4 guard counters; fixture spec.md hashes `739d7126…/7927ee00…/02b0c23f…` byte-identical pre/post; fixture `git status` empty; corruption diff section empty. Output preserved: `repro-output-fixed-orchestrator-rerun.txt`.
- **Scope**: branch diff vs own base `0b1e27877` = 15 files (13 at run-phase close + 2 evidence files added after) — the declared envelope (`internal/spec/status.go`, `internal/cli/spec_status.go`, + their tests, SPEC artifacts, reports) plus the compiler-forced 3-line `exitcode_streams_test.go` signature cascade. Zero template/catalog/kanban files.
- **Close shape**: sync commit `166ec90ef` carries CHANGELOG + `status: in-progress → completed` frontmatter + §E.4; backfill `c106f620d` writes the real SHA; `git diff --name-only 166ec90ef..HEAD` contains zero `.go` files (sync close is the last code write).

## Baseline-attribution

All figures above: this run, against `WT-spec-status-dryrun` @ `c106f620d` unless explicitly pinned otherwise (repro fixture verdicts: freshly built binary from this HEAD; plan-audit 0.875: iter-2 verdict tree, artifact-hash stable since). Agent-reported figures used in this card were re-measured by the orchestrator or the sync-auditor independently (sync-auditor re-ran the AC sweep — 38 PASS / 0 FAIL, swept set counted — and its own binary repro, diffing behavioral lines against the checked-in fixed output: zero divergence).

## Gaps (explicitly not observed)

- Full test suite: NOT run locally (machine-load discipline); CI on the integration branch owns it.
- `GOOS=linux` build not separately measured (pure Go string/regex change; darwin native + windows/amd64 verified).
- The AC-002 mutant was executed by the run phase (recorded RED) and verified analytically by the auditor — not re-executed by a third party (file-modification boundary).
- Single-shot repro runs (deterministic fixture asserts; agent + auditor + orchestrator each ran the harness once — three independent executions agree).

## Residual-risk / carried items

- **`TestCatalogHashParity` red on THIS tree** — attributed: names 4 template agent files (`manager-develop/manager-lead/manager-design/e2e-tester.md`) untouched by this branch; root cause is t497's template revisions vs late catalog refresh, **already fixed on develop by t526 (`072bc9f60`)**; lead-measured GREEN at develop tip `6a46c0edb`. Expected green after develop absorption at the integration window; if still red there, that is a NEW fact and will be reported.
- `internal/cli` package coverage 81.0% vs 90% convention — pre-existing (package untouched except `spec_status.go` at 66.1% with documented unreachable seams).
- sync-audit optional findings F1-F5 (Low/Info, non-blocking): duplicated comment block `status.go:282-284`; `TestValidStatuses` missing 2 enum values (pre-existing); indented-`status:` tolerance (speculative, shared predicate = no split-brain); scope-count drift 13→15 (evidence files, noted); hardcoded `→ implemented` dry-run label (consistent today). Recommended for future touching cards, not this one.
- **em-dash git-implied hypothesis** (reporter's secondary): investigated, NOT reproduced (fixture with identical em-dash close subject behind `--no-ff` merge classifies `completed`); recorded spec.md §E; reporter to be invited to re-test on 3.2.0.
- **Issue reply**: draft at `issue-reply-draft.md`; posting is the LEAD's act after remote landing with operator approval (lane does not post).

## Window request (follows)

Formal request sent after this commit: card t513 · `WT-spec-status-dryrun` @ HEAD · unpushed commit count · evidence path (this file) · re-measurement scope as above. Absorb target: LOCAL develop (lead directive 2026-09-07; re-read at window entry — the two waiting lanes' merges will have advanced it).
