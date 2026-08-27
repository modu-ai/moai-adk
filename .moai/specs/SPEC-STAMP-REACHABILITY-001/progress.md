# progress.md — SPEC-STAMP-REACHABILITY-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier M canonical set (spec.md, plan.md, acceptance.md) + research.md + progress.md, written in worktree `.claude/worktrees/t291` (branch `WT-stamp-orphan-guard`, base `da791eb0a`), 2026-08-27.
- Requirements: 10 REQ (GEARS, pattern-annotated: REQ-SR-001..010); 10 AC (Given-When-Then) each carrying a RED-now baseline measured at base; REQ↔AC traceability both directions complete (acceptance.md §D.1). Counts below Tier M ceilings.
- Premise re-measurement (this run, verbatim outcomes): `git merge-base --is-ancestor 0d15864ae90b origin/main` rc≠0 (`NOT_ANCESTOR` — orphan fixture live); provenance `.commit_sha` read as `410da655f39d…` via jq; `git merge-base --is-ancestor 410da655f39d1d3731abf2f247e8e5353a9d0de5 origin/main` rc=0 (`CURRENT_STAMP_ANCESTOR` — guard-green baseline live). SPEC ID regex self-check executed as Bash: `PASS`; catalog uniqueness probe count 0.
- Countermeasure adjudication per lead dispatch: option 1 (CI pre-merge reachability guard) and option 2 (explicit stamp-commit mode) BOTH selected as in-scope milestones M2/M1; option 3 alternatives rejected-and-recorded only. Validation split recorded as a boundary decision: resolvability at stamp time, survivability enforced by the CI gate where it is decidable.
- Scope fences: #1666 recurrence classification stays with t278; provenance schema evolution; historical-orphan retrofit; mx/edges layers — all fenced in spec.md §E with H3 Out-of-Scope headings.
- Constraint inheritance: predecessor REQ-GFR-014 (never restamp vs branch-local HEAD) binds this SPEC's own delivery; described-source churn below threshold means the committed stamp carries unchanged **provided its verification-time `.commit_sha` remains main-reachable** — at-or-above threshold accepts the pre-merge red until a POST-MERGE main restamp outside delivery scope (spec.md §D).
- Delta round (2026-08-27, post plan-audit iter-2 PASS-WITH-DEBT 0.94): v0.2.1 wording remediation N2/N3/N4/N5 all applied; N3 authoritative numeral re-measured first-hand (`grep -n` single hit → check.go**:213** at HEAD 016dc0b8c; iter-2's :214 was an authoring mis-count); HISTORY row appended; version field bumped 0.2.0 → 0.2.1 for traceability across frontmatter/HISTORY/progress.
- Stamp-health moving-target observation, recorded as dated fact (audit-cycle context): on 2026-08-27 after upstream #1668 landed through HEAD 0364217e9+ lineage, the tracked `.moai/project/codemaps/provenance.json` names `commit_sha` = `a995e58fa69b40f3d76a76c0a5ca56d0594fb528` — locally present (`git cat-file -t` → commit) and NOT an ancestor of origin/main (`git merge-base --is-ancestor … origin/main` rc=1): a live SECOND occurrence of the F5 orphan class on main. Instance repair routes to lead triage OUTSIDE this card's scope; no requirement text changed on its account. Live-probe corroboration with the §A.1 canonical script returned rc=1 with the ancestry annotation naming that sha — the planned M2 guard would surface exactly this state red pre-merge.
- Constraint inheritance note superseded in part by the observation above: iter-2's "committed stamp expected to carry unchanged" premise was TRUE of its measurement moment (sha `410da655f39d…`, ancestor-verified rc=0, 2026-08-27 pre-#1668) and is retained as a date-pinned historical observation only (acceptance.md baseline-attribution, AC-SP-002(ii)).
- Evidence paths cited across artifacts resolve within the worktree tree (triage-table.md tracked on base; code anchors at graph_stamp.go/provenance.go/check.go/graph_check.go line-cited in plan.md §H).

## §E.2 Run-phase Evidence

Run-phase executed 2026-08-27 in this worktree, branch `WT-stamp-orphan-guard`, TDD (RED-GREEN-REFACTOR) per milestone. Evidence artifacts persist under `.moai/state/verify/t291/` (dryrun.log, red-mx.txt, red-cli.txt, green-cli.txt, anatomy.txt, coverage.txt, docs-parity-check.sh, guard-ref.sh, guard-mutant.sh, dryrun-driver.sh).

### Pre-flight (plan.md §C 1-5, verbatim outcomes)

1. Worktree identity: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t291`; branch `WT-stamp-orphan-guard`; HEAD `cbd33980a` at pre-flight (plan-phase base da791eb0a advanced via absorbed merge e514e695f — expected).
2. Tracked provenance `.commit_sha` = `a995e58fa69b40f3d76a76c0a5ca56d0594fb528`; `git merge-base --is-ancestor <sha> origin/main` → **rc=1**; `git cat-file -t <sha>` → `commit` (locally present). Date/tree pin: 2026-08-27, tree cbd33980a. This is the known live second F5-class instance recorded in §E.1 — known-state correction acknowledged, NOT treated as a planning regression, NO repair attempted, NOTHING restamped (REQ-GFR-014). NOTE: plan.md §C item 2's "still rc=0" wording predates the upstream #1668 replacement and is superseded by the dispatch's known-state correction and §E.1's date-pinned observation.
3. Orphan fixture: `git cat-file -e 0d15864ae90b^{commit}` → rc=0 (fixture live in this worktree's history).
4. Affected-package baseline BEFORE edits: `go test ./internal/cli/... ./internal/mx/... ./internal/graph/...` → all `ok` (20 packages, zero failures).
5. Divergence: `git rev-list --count --left-right origin/main...HEAD` after `git fetch origin main` → `0 6` (local ahead only; clean).

### M1 — Explicit stamp-commit mode (commit cd71c701e)

**RED evidence (TDD invariant i)** — verbatim pre-GREEN failing output: mx-level `go test ./internal/mx/ -run 'TestResolveCommit|TestStampCodemaps' -count=1` → `undefined: ResolveCommit` + `too many arguments in call to StampCodemaps` → `[build failed] FAIL` (.moai/state/verify/t291/red-mx.txt). CLI-level (implementation reverted before test authoring per invariant ii): `go test ./internal/cli/ -run 'TestGraphStampCmd_...' -count=1` → `not enough arguments in call to mx.StampCodemaps` + `undefined: commitRev` → `[build failed] FAIL` (red-cli.txt).

**GREEN**: all six new CLI tests + four new mx tests pass (green-cli.txt); full targeted set green.

| AC | Status | Verification command + verbatim outcome |
|----|--------|------------------------------------------|
| AC-SP-005 | PASS | `go test ./internal/cli/ -run TestGraphStampCmd_ExplicitCommitRecordsResolvedSHA` → `--- PASS` — named parent sha recorded verbatim (not HEAD), dirty:false, schema 1, described_roots `[internal cmd pkg]` |
| AC-SP-006 | PASS | `go test ./internal/cli/ -run TestGraphStampCmd_UnresolvableCommitRejected` → `--- PASS` — `deadbeef42` exits non-zero, error names the revision, path-free, NO provenance file and NO temp written |
| AC-SP-007 | PASS | CLI `TestGraphStampCmd_CommitWithDirtyTreeRejectedPreWrite` (no file, no temp, tree untouched) + mx `TestStampCodemaps_ExplicitCommitDirtyRejected` (entry rejects pre-install) → both `--- PASS` |
| AC-SP-008 | PASS | `TestGraphStampCmd_FlaglessCharacterization` + `TestStampCodemaps_DefaultPathUnchanged` → `--- PASS` — clean→HEAD anchor + exact Describe line; dirty→64-hex fingerprint, commit_sha key omitted; `--commit HEAD` legal no-op |
| (edge) short-rev / ref-name | PASS | `TestGraphStampCmd_ExplicitCommitShortRevResolvesFull`, mx `TestResolveCommit` subtests (full sha / short rev / ref name / named parent) → `--- PASS` |

Files: `internal/cli/graph_stamp.go` (flag + validation order + Long recipe text), `internal/cli/graph_stamp_commit_test.go` (new), `internal/mx/provenance.go` (`ResolveCommit` new; `StampCodemaps(projectRoot, commitSHA)` signature, mixed-anchor rejection), `internal/mx/provenance_stamp_test.go` (new).

### M2 — CI pre-merge reachability guard (commit e4eb15ea4)

Shipped step matches acceptance.md §A.1 canonical text modulo env-plumbing lines (`REPO`/`PROV` assignments, `env: GITHUB_BASE_REF: ${{ github.base_ref }}`). Trigger block untouched. One anatomy self-catch during the run: the step comment initially contained the literal token `continue-on-error` ("no continue-on-error"), tripping AC-SP-011(b)'s `grep -c` → 1; reworded → 0.

**Dry-run (E3)** — every §A.1 branch executed as local shell via `.moai/state/verify/t291/dryrun-driver.sh` against `.moai/state/verify/t291/guard-ref.sh` (canonical text) and `guard-mutant.sh`; full log at `.moai/state/verify/t291/dryrun.log`:

| Cell | Outcome (verbatim key line) | rc |
|------|------------------------------|----|
| A non-ancestor (AC-SP-001) | `::error::codemaps provenance stamp 0d15864ae90b is NOT an ancestor of PR base origin/main — orphan-bound stamp (a squash merge would strand it)` | 1 |
| B ancestor GREEN (AC-SP-002(i), ancestor read fresh: `git merge-base HEAD origin/main` → 3abde70530c4…) | `guard: stamp 3abde70530c4f1c8ebc5f1bc9c4b8f5912380271 reachable from origin/main` | 0 |
| C absence premise (AC-SP-003 setup) | `fatal: Not a valid object name 0d15864ae90b^{commit}` | 128 |
| C missing-object (AC-SP-003) | `::error::codemaps provenance names commit 0d15864ae90b which does not exist in this checkout's history` | 1 |
| D no-anchor (AC-SP-004) | `guard: provenance carries no commit anchor (dirty/content-fingerprint anchor) — reachability not judgeable, skipping` | 0 |
| E push-event (AC-SP-011(a)) | `guard: push event (no base ref) — object presence verified only` (zero ancestry output) | 0 |
| MUTANT (emptiness test dropped) | `::error::…NOT an ancestor of PR base base origin/` — annotation naming origin/ | 1 |
| EDGE corrupt provenance (§E edge row) | `jq: parse error: …` — step fails, never passes | 5 |
| LIVE probe (AC-SP-002(iii), record-only) | `::error::codemaps provenance stamp a995e58fa69b40f3d76a76c0a5ca56d0594fb528 is NOT an ancestor of PR base origin/main …` | 1 |

Live-probe interpretation (fixed per acceptance.md): red against the tracked tree is CORRECT instrument output — the guard surfacing the recorded live orphan instance; instrument validity rests on cells B/A.

| AC | Status | Verification command + verbatim outcome |
|----|--------|------------------------------------------|
| AC-SP-001 | PASS | Cell A above (rc=1, annotation names sha + base ref) |
| AC-SP-002 | PASS | Cell B above (rc=0, pass line); live probe recorded, not asserted (rc=1 as §E.1 pins) |
| AC-SP-003 | PASS | Absence premise rc=128 measured first; guard cell rc=1 with the DISTINCT missing-object message |
| AC-SP-004 | PASS | Cell D rc=0 with greppable `no commit anchor` token |
| AC-SP-011 | PASS | (a) cell E rc=0 presence-only + mutant cell rc=1 with `::error::` naming origin/; (b) `grep -c 'continue-on-error'`→0, `- name:.*[Rr]eachab`→1, `GITHUB_BASE_REF:`→2 (≥1), `\[ -z "\${GITHUB_BASE_REF:-}" \]`→1; (c) awk-range `grep -c '^      if:'`→0 (anatomy.txt, executed post-M2 on the landed YAML) |
| AC-SP-012 | PASS | Shipped step body: forbidden-token grep (`-mtime|-atime|-ctime| find |newermt|newer-than|date +%|time.Now|os.Stat`) zero matches (rc=1); positive allowance `cat-file -e`/`merge-base --is-ancestor`/`jq -r` each 1 |

YAML syntax validated: `python3 -c "import yaml; yaml.safe_load(...)"` → OK. **NOT claimed executed-live**: the workflow's first real CI run on the delivering PR is the live witness — **prediction (stated, never pre-claimed)**: the guard step fires RED on this PR (missing-object clause naming `a995e58f…` — measured NOT an ancestor of origin/main AND NOT an ancestor of this branch HEAD, both rc=1, 2026-08-27 tree cbd33980a), and `graph check` exits 2 not-comparable for the same recorded live instance. Both are correct instrument output for the live orphan; instance repair routes to lead triage OUTSIDE this SPEC.

### M3 — Regression lock + operator docs (commit: see §E.3 run_commit_sha)

Coverage sweep measured THREE gaps (plan.md expected none — measured, not assumed): reverted-churn-zero, described-roots diff scoping, one-below-threshold boundary. All three pinned by new characterization tests in `internal/graph/check_regression_lock_test.go` (each names the mutant it discriminates; characterization of existing behavior, green on arrival by design). Already covered (selectors located, all passing): dirty-fingerprint anchoring (`TestCheckFreshness_DirtyGenerationAnchor`), at-threshold stale (`TestCheckFreshness_AgedLayerFails`, exactly 40 files), not-comparable system error (`TestCheckFreshness_NotComparableIsSystemError`) + CLI exit 2 (`TestGraphCheckCmd_NotComparableExitsTwo`, `TestGraphCheckCmd_MalformedGateYamlExitsTwo`), endpoint-diff counting (`TestCheckFreshness_PerLayerMetrics`). Threshold authoritative pin re-verified: `check.go:213` = `if count >= th.CodemapsChangedFiles {`.

| AC | Status | Verification command + verbatim outcome |
|----|--------|------------------------------------------|
| AC-SP-009 | PASS | `go test ./internal/cli/... ./internal/mx/... ./internal/graph/... -count=1` → 20× `ok`, zero failures (baseline also green pre-edit — no inherited-red masking). New pins: `TestCheckFreshness_RevertedChurnCountsZero` / `_DescribedRootsScopeFidelity` / `_ThresholdBoundaryOneBelowFresh` all `--- PASS` |
| AC-SP-010 | PASS | `.moai/state/verify/t291/docs-parity-check.sh` → `AC-SP-010 OK: all four locales carry recipe + cmd + prohibition` (rc=0); `hugo -s docs-site --minify --gc` rc=0, warning-free (`grep -ciE 'warn|error'` on full output → 0), 187 pages |

Docs files: `docs-site/content/{en,ko,ja,zh}/cli-reference/graph.md` — new `###` subsection under `moai graph stamp codemaps` (en: "Naming a merge-surviving commit (`--commit`)"), code spans untranslated, prohibition prose native idiom per locale.

### E1-E5 self-verification (plan.md §E)

- **E1** AC matrix: 12/12 PASS (table above; AC-SP-001..012).
- **E2** Package gates: targeted `go test` 20× ok; `go vet ./internal/cli/... ./internal/mx/... ./internal/graph/...` → rc=0; `gofmt -l` on all 5 touched Go files → clean (one self-catch: regression-lock file initially flagged, fixed before commit).
- **E3** Guard dry-run: 9 cells quoted above (dryrun.log).
- **E4** Boundary grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/graph_stamp.go internal/mx/provenance.go` (non-test, non-comment) → zero matches (rc=1); path-free stderr asserted by tests (AC-SP-006/007 test bodies).
- **E5** Docs parity: 4/4 locales, same change-set, hugo warning-free.
- **E8 (TDD)** RED outputs: red-mx.txt + red-cli.txt (verbatim, captured pre-GREEN).

Stamp-state final check (REQ-GFR-014 inheritance): NOTHING was restamped during run-phase — `git status` shows `.moai/project/codemaps/provenance.json` unmodified throughout; described-source churn vs the tracked stamp measures 29 changed + 1 untracked = 30 < 40 threshold (`git diff --name-only a995e58f… -- internal cmd pkg | wc -l`; `git ls-files --others -- internal cmd pkg | wc -l`, 2026-08-27) — the committed stamp carries unchanged; its `.commit_sha` verification-time reachability is the recorded lead-triage item, not an in-SPEC repair.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-08-27"
run_commit_sha: "2378bc14c033652313e7b62519caf3de25e3442e"
run_status: "run-phase-complete"
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE-list file modified beyond plan scope
l44_pre_commit_fetch: true        # git fetch origin main executed pre-flight & pre-commit
l44_post_push_fetch: false        # no push in run-phase (integration deferred to orchestrator)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: "not-run (targeted-tests-only discipline; go build compiled via test builds)"
  windows_amd64: "not-run (no build-tag or syscall change; touched files are pure Go + shell-in-YAML)"
total_run_phase_files: 14         # 5 Go (2 new tests, 2 impl, 1 new graph test) + 1 YAML + 4 docs + 4 SPEC/evidence
m1_to_mN_commit_strategy: "one commit per milestone (M1 cd71c701e, M2 e4eb15ea4, M3 <this commit>) + §E.3 backfill if needed"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Input parameters: tier=M · scope ≈6-8 files (internal/cli/graph_stamp.go+tests, internal/mx/provenance.go+tests, .github/workflows/graph-freshness.yml, command Long text, docs-site cli-reference ×4 locales) · domains=3 (Go CLI / CI YAML / docs markdown) · language mix Go+YAML+MD · concurrency benefit LOW (coding-heavy, inter-milestone dependency M1→M2→M3 sequential in plan.md) · Agent Teams prereqs not requested.
- Mode evaluation: direct — not selected (multi-file semantic change) · fanout — not selected (Anthropic coding-task parallelism caveat: implementation is coding-heavy with milestone ordering; parallel spawns would race shared files) · sweep — not selected (<30 files, non-mechanical multi-rule edits, Workflow overhead unwarranted) · **serial — selected**.
- Decision: `serial` (one `manager-develop` sub-agent, milestones in plan.md order, TDD cycles per milestone).
- Justification: coding-heavy work plus strict M1→M2→M3 dependency ordering makes sequential single-agent execution both the safe default and the cache-economical one; a single §E evidence stream per milestone lands contiguously in §E.2. Recorded after Implementation Kickoff Approval (operator-approved 2026-08-27, relayed by lead session) and pre-spawn sync checks (divergence absorbed through merge `e514e695f`; same-SPEC active sessions `[]`).
