# t1833 progress (run phase, Class B, no SPEC)

Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833`, branch `WT-10-17-class-7`, base `fb749d372` (local main). Commits: `c78e02e68` (bisect verb), `44bff97a6` (release CI gate), `774c166ab` (runbook). Nothing pushed. No remote contacted, no tag touched, `scripts/release.sh` never run for real (tests use `--dry-run` against scratch repositories with a stub `gh`).

## 1. Cause-establishing evidence (Class B record)

Card (a): is a landing one `--no-ff` merge carrying the card id?
Code: `internal/factory/integration_merge_step.go:399` `mergeMsg := fmt.Sprintf("Merge %s into %s (card %s, integration merge)", cardBranch, in.IntegrationBranch, in.CardID)`; line 609 `git("merge", "--no-ff", "--no-autostash", "--no-overwrite-ignore", "-q", "-m", mergeMsg, pinned)`.
Observed (command `git log --first-parent --format='%P|%s' main -n N`, classified by parent count and `\(card t[0-9]+` in the subject, tree fb749d372):

| window | two parents + card id | two parents, no card id | one parent |
|---|---|---|---|
| 40 | 40 | 0 | 0 |
| 100 (back to 2026-10-07) | 64 | 4 | 32 |

Examples, two parents + card id: `fb749d372 d2d9ac08f c3d2a5d4c | Merge WT-10-15-release-3 into main (card t1760, integration merge)`; `7faf75fdf e806cdd6c e6ca1d560 | Merge WT-audit-gate-revert into main (card t1845, integration merge)`.
Two parents, no card id: `ad002d3a0 ... | Merge WT-codex-moai-home-add-dir into main (codex --add-dir moai home, integration merge)`; `2aab5f797 ... | Merge pull request #1816 from modu-ai/WT-always-loaded-char-budget`.
One parent: `842160157 763193861 | chore(config): auditor and planner models per operator role table (opus high)` (3 such chore commits); `db0c514d3 d3f51d0d9 | fix(factorymsg): set synchronous(NORMAL) on broker open paths (t1592) (#1815)` (PR-era single-parent commits, 28 of the 32 carry a `tNNNN` token).
First-parent positions 1-42 (2026-10-14 .. 2026-10-17) are all card merges. Older than the 100 window: 0 further card merges (`awk` over the rest of first-parent history; 138 of 3738 first-parent commits are two-parent).

Batch shape (origin/main, read-only local ref): `M=6dde1c1c1` "Merge pull request #1826 from modu-ai/WT-batch-landing-1017": `git merge-base --is-ancestor M^1 M^2` true; `M^{tree}` equals `M^2^{tree}`; `git log --first-parent M^1..M^2` yields 63 units, including two non-card units (`merge: absorb origin/main into the batch landing branch`, `fix: repair three main-side CI failures blocking the batch landing PR`). So origin/main first-parent = one step per batch; card granularity lives on the `M^2` chain.

Result: the claim "each landing carries the card id" holds for the code path and for the observed log since 2026-10-14; it does not hold for direct commits and batch-assembly commits, which a bisect reports as `card: none`.

Card (b), pre-change defect (RED, scratch repo, stub gh): `TestReleaseScriptCIGate` all 7 rows FAIL on the unchanged script; `TestReleaseScriptCIGateWithoutGh` on the unchanged script: exit 0, output `[WARN] gh CLI not available — skipping CI status check` (fail-open), override run `[WARN] CI check skipped (--skip-ci-check)` with no summary record. Old read: `gh api /repos/modu-ai/moai-adk/commits/$LOCAL_SHA/status` (combined status), `pending` accepted via y/N prompt, unknown only warned.

## 2. Design decision

- Verb vs recipe: a recipe alone is `git bisect start --first-parent <bad> <good>` + `git bisect run sh -c '<cmd>'` + `git bisect reset`. A verb is warranted only for what the recipe cannot enforce: bisect moves HEAD, so the primary checkout must be refused (shared by several sessions); a dirty tree, an in-progress bisect and an unresolvable ref fail late and confusingly; a forgotten `reset` leaves a tree mid-bisect; and the useful output is the first-parent merge subject with the card id, which `git bisect run` prints only as a SHA. Verb kept minimal (one file, ~130 lines, reuses `runScrubbedGit`, `identifyPrimaryCheckout`, `factorySameTree`, `shellJoinArgs`, `factory.ShortSHA`); `--dry-run` prints the three git commands. Not built: endpoint verification, a batch-aware `--batch <M>` shorthand, an inner-card mode (documented as 3 git lines in the runbook section 2.4).
- Release gate: in `scripts/release.sh` validation 9 (single tag-placement point). Check runs, not combined status. Strict (all check runs) rather than required-only: the tag is irreversible. Override stays `--skip-ci-check`, recorded in output and summary. GoReleaser: `.goreleaser.yml` consumes an already-pushed tag and has no point at which a tag is placed or refused; `release.yml` runs after the tag exists. Not changed; named as a follow-up candidate in the runbook.
- Template mirror: none. The runbook is repo-local (`scripts/release.sh`, `AGENTS.local.md`); `internal/template/templates/.moai/docs/` mirrors 6 generic docs only, and no test lists `.moai/docs` membership (moai-docs sweep floor 33, repo has 41 files).

## 3. Tests (all with TMPDIR=GOTMPDIR=/private/tmp/lane3-bisect, no -race, one package per run)

| command | exit | result |
|---|---|---|
| `go test ./internal/cli -run '^TestBisect' -count=1 -v` | 0 | 8 tests PASS, `ok ... 6.876s` (prefix selector, 8 tests, 7 s; broader than a single name) |
| `go test ./internal/template -run '^TestReleaseScriptCIGate$' -count=1` | 0 | ok (7 rows) |
| `go test ./internal/template -run '^TestReleaseScriptCIGateWithoutGh$' -count=1` | 0 | ok (2 rows) |
| `go test ./internal/template -run '^TestReleaseScriptCharacterization$' -count=1` | 0 | ok, also ok before the change |
| `go test ./internal/template -run '^TestReleaseScriptMatrixGate$' -count=1` | 0 | ok |
| `go test ./internal/template -run '^TestReleaseScriptHeadAdmission$' -count=1` | 0 | ok |
| `go test ./internal/template -run '^TestReleaseScriptRcChangelogSkip$' -count=1` | 0 | ok |
| `bash -n scripts/release.sh` | 0 | syntax ok |
| `go vet ./internal/cli`, `go vet ./internal/template` | 0, 0 | clean |
| `gofmt -l` on the 5 Go files touched | 0 | no output |

Mutants (each applied by `sed`, run, then restored; `cmp` against the saved original confirmed restoration):
- bisect verb: no primary refusal -> `TestBisectRefusesThePrimaryCheckout` FAIL; no reset -> `TestBisectNamesTheCulpritCard` and `TestBisectReportsAnAbortedSearchAndStillResets` FAIL; no `--first-parent` -> `TestBisectNamesTheCulpritCard` and `TestBisectDryRunPrintsTheGitCommandsAndMovesNothing` FAIL; no dirty check -> `TestBisectRefusesADirtyTree` FAIL (first attempt did not compile, reapplied with `&& false`).
- release.sh: pending branch disabled -> `.../pending_is_refused` FAIL; gh-missing made non-fatal -> `.../no_override` FAIL; failed branch disabled -> `.../failed_is_refused` and `.../failed_beats_pending` FAIL.

## 4. Files changed

`internal/cli/integration_bisect.go` (new), `internal/cli/integration_bisect_test.go` (new), `internal/cli/integration.go` (register verb), `scripts/release.sh`, `internal/template/release_fixture_test.go` (stub answers check-runs, default green), `internal/template/release_ci_gate_test.go` (new), `.moai/docs/batch-failure-bisect-runbook.md` (new), this file.

## 5. Gaps

- The verb was not run against the real repository history and `moai` was not rebuilt (`make build` barred); its behavior is evidenced on synthetic repositories with real git only.
- The jq expression and the `check-runs` response shape were exercised against the stub only; no live GitHub call (remote contact barred). The claim that the combined-status endpoint omits Actions check runs is GitHub's documented definition, not observed here.
- `golangci-lint` was not run (package-wide lint judged heavy); `go vet` and `gofmt` were.
- The sweep guard (`github_flow_sweep_guard_test.go`) is env-armed and was not run against the new doc.
- Whole-package `go test ./internal/cli` and `./internal/template` not run (barred); only the named tests above.
- `.github/workflows/release.yml` (post-tag provenance gate) deliberately untouched.
- Local main and origin/main currently diverge (`git rev-list --count --left-right origin/main...main` gave `10 131`); noted only, not acted on.

## 6. Residual risk

A strict all-check-runs gate can block a release over an auxiliary failing check; the override exists and is recorded. A hand-made tag bypasses `release.sh` entirely.

## WAIT STATE: sync-audit iteration 1 FAIL; the fix spawn is denied by the audit receipt guard (lane-3, 2026-10-17)
- Verdict: .moai/reports/t1833/sync-audit-iter1.md, FAIL 0.62 at audited_sha 8f1873c58d32e8f09fd9655b0b835399b2ff8a48 (committed 2ca38cdcf); codex required gate `fail`, receipts rcpt-12da61c41fa93963600ea2a4 and rcpt-5abefcca08d731b67723a1c5. Blocking: F1-F5 (bisect culprit read from mixed output; GIT_* env bypasses the primary-checkout refusal; failed `git bisect reset` returns success; failed `bisect start` leaves HEAD moved; `git tag -a` names no commit). Optional but P1 by codex: F6 (gate passes on any single green check run). Low: F7-F9.
- What was attempted: `Agent(subagent_type: manager-develop)` for fix round 1 (Reproduction-First, one test per defect). Denied by the PreToolUse hook, verbatim: "AUDIT_RECEIPT_VIOLATION: manager-develop cannot be spawned while an audit PASS stands unproven in /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833. Outstanding: sync-auditor / unknown-spec / verdict line missing. Re-run the audit through codex_audit or audit_multi and let the auditor end with a verdict line citing the receipt."
- Observed record (read-only, `.moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json`, gitignored machine-local state): agent_type sync-auditor, spec_id unknown-spec, agent_id a0c5e4ec206274aa7 (the t1833 sync auditor), cause "verdict line missing", rejected_at 2026-10-17T14:34:51Z, reentry_warned true. It was written when the auditor stopped once without a parseable verdict line and was blocked, then continued; the auditor's final message did end with `AUDIT-VERDICT: FAIL spec=t1833 receipts=...`.
- Read from internal/hook/audit_receipt_guard.go (not tested by running it): a FAIL verdict line returns without clearing rejections ("A FAIL needs no receipt"); only a proven PASS by the same role in this tree clears receipt-kind rejections (ClearRejectionsForRoleInTreeKind); phase-entry spawns (manager-develop, manager-docs, manager-git) are denied while any such record exists; no CLI verb or rule documents a manual clear (grep over .claude/rules, .moai/docs, AGENTS.local.md found no mention of the sentinel). Consequence: the only sanctioned way to clear it is a PASS, a PASS needs F1-F5 fixed, and the fixes are routed to the specialist the guard denies.
- Not done on purpose: the record was not deleted or edited (it is the guard's record, not the lane's); no substitute agent type was spawned to make the fixes (that routes around the deny); the lane did not edit the Go code or scripts itself (delegation duty). No push, no tag.
- Who is waited on: the leader (operator-delegated), who decides how to proceed; the lane asked by message.
- Resume condition: a decision that lets the fix round run. Then: fix round 1 as scoped above -> commit -> `moai factory stage t1833 sync <sha>` is not needed (stage is already sync) -> sync-audit iteration 2 on the new HEAD with the codex gate -> T11 `sync-audit` with `<audited_sha>:<tracked file>` -> T13 `merge-ready` -> slot, re-measure, window, `factory complete`.
- State now: factory record t1833 = sync v6 (run tn1p4e), lease kept alive by the session-local renew cron; card branch WT-10-17-class-7 tip is this record's commit on top of 2ca38cdcf.
