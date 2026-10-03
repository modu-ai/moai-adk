# Plan — SPEC-TEST-ENV-HERMETIC-001

Milestones are ordered by decision-reversibility: the guard and the scrub-set decisions (the
choices most likely to change after the measurements) come first; mechanical steps come last.
Priorities are High / Medium / Low; no time estimates.

## §A Context

- **Worktree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1356`, branch
  `WT-test-env-hermetic-sweep`, plan-phase measurement tree `2de0a2cb6` (re-read `git rev-parse
  --short HEAD` and `git branch --show-current` before any commit).
- **Card**: t1356, Class C (plan → plan-audit → run → sync). Git-flow lane protocol: no card PR;
  integration into local `develop` via the integration window; CI on `origin/develop` is the
  full-suite verdict surface. The card id `t1356` appears in every commit message.
- **Artifacts**: spec.md + plan.md + acceptance.md (Tier M) + progress.md + decision-index.md
  (the decision gate is on in `.moai/config/sections/interview.yaml`).
- **Development mode**: TDD — the RED guard commit precedes the fix commits (REQ-THE-006).
- **Evidence status**: the five observed reds and the hook one-axis arms are re-measured and carried
  in acceptance.md §D.0 (ledger E-1..E-5). `.moai/reports/t1356/baseline.md` is local-only and
  gitignored; never `git add -f` it and never re-add a gitignore negation.
- **PRESERVE**: every production file; `internal/cli/ptycaptest/` (own guard); the six hook and ten
  cli helpers' bodies (no delegation in this SPEC); `factoryEnvPinnedEnv` semantics; all
  currently-green assertions.
- **Existing infrastructure to EXTEND**: cli `factoryAmbientEnvKeys` + `clearFactoryAmbientEnv`
  (called from `main_test.go:356`); hook `TestMain` (`main_test.go:67`, same file that already
  scrubs `CLAUDE_PROJECT_DIR`, card t1165).

## §B Known Issues

- **B1 — `t.Setenv` panics under `t.Parallel`.** Any M4 pin must drop `t.Parallel()` from that
  test or the test is handled differently; audit the touched test for parallel markers first.
- **B2 — a scrubbed arm needs `unset … && go test …` as ONE compound invocation.** A separate
  `unset` does not reach the next command (each Bash call is a fresh process).
- **B3 — `env -u … go test` is refused by the worktree guard** (measured at plan time: "cannot be
  shown not to be git"); `unset <VARS> && go test …` and `env NAME=value go test …` are accepted.
  Use those forms. A refusal is recorded as a Gap, never silently substituted
  (`verification-claim-integrity.md` §3.1).
- **B4 — `env NAME=value go test` does not remove other ambient axes.** A "lane arm" must set the
  whole lane env explicitly and a "scrubbed arm" must unset the whole family; a partial arm
  measures neither. Plan-phase partial arms were discarded for this reason (progress.md §E.1).
- **B5 — go test caching.** Every evidence-bearing run uses `-count=1`.
- **B6 — empty sweep.** A `-run` selector that matches nothing prints `[no tests to run]` and exits
  0; every narrow-run AC counts `--- PASS` lines, not exit codes alone (`verification-completeness.md` §1.1).
- **B7 — cross-platform.** Guard path logic uses `filepath`; `GOOS=windows GOARCH=amd64 go vet` on
  the two packages type-checks the test files for the Windows build.
- **B8 — scope discipline on a shared machine.** Stage by explicit pathspec only; never `git add
  -A`; re-read `git status --short` immediately before staging.
- **B9 — hook child re-exec sites.** `internal/hook` re-executes its own test binary in at least
  four files (`slot_lease_guard_test.go`, `session_start_drift_fill_test.go`,
  `session_start_drift_fill_burst_test.go`, `factory_handoff_race_test.go` / `_bind_test.go`); a
  `TestMain` scrub also reaches those children.

## §C Pre-Flight (run-phase session, before any commit; record outputs in progress.md §E.2)

```bash
git rev-parse --short HEAD
git branch --show-current
# expect: 2de0a2cb6 or a descendant containing only this SPEC's plan-phase commit / WT-test-env-hermetic-sweep
```

Then, one command per call:

1. Re-establish the five observed reds with the RED-now commands of acceptance.md §D.0 (E-1, E-2) —
   same commands, same tree lineage; record verbatim output and the tree SHA.
2. List the family from `internal/config/envkeys.go` (constants whose value starts `MOAI_FACTORY_`
   or `MOAI_KANBAN`, plus `MOAI_AUTONOMY_TIER`) — this list builds both arms' env; do not copy it
   from this plan.
3. Hook child re-exec census: for each `os.Args[0]` site in `internal/hook/*_test.go`, record
   whether its `cmd.Env` carries a family axis as payload (B9). Expected: none do; if one does,
   M3 uses the pin-marker pattern.
4. `GOOS=windows GOARCH=amd64 go vet ./internal/cli ./internal/hook` and `golangci-lint run
   --timeout=2m` baselines (distinguish NEW from pre-existing findings later).
5. Acquire the lease before any whole-package run: `moai slot status --resource heavy-test`, then
   `moai slot acquire --resource heavy-test --max-duration 20m`; release with `moai slot release
   --resource heavy-test` immediately after each arm. Exit 3 = held by another session: report and
   wait, never `--force`.

## §D Constraints (Hard)

1. **Test-first ordering (REQ-THE-006).** Commit order is fixed: (c1) baseline record → (c2) guard
   commit → (c3) cli fix → (c4) hook fix → (c5+) contingent pins → evidence/sync. No fix commit
   before c2 exists and its red is recorded.
2. **Test files only (REQ-THE-008).** `git diff --name-only 2de0a2cb6..HEAD` contains only
   `*_test.go` and `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/*`.
3. **No skip / delete / blacklist (REQ-THE-004).**
4. **Heavy runs leased and serial** (§C step 5); never `go test ./...`.
5. **English code comments**; env names via `internal/config/envkeys.go` constants; no inline
   `"MOAI_*"` literal in a test file (the existing `TestFactoryRoleEnvConstant` forbids the
   literal `"MOAI_FACTORY_ROLE"` in `internal/hook`).
6. **Conventional commits** with card id `t1356`; `🗿 MoAI` trailer per repo convention.
7. **MX tags**: new guard helpers are test code; no `@MX:ANCHOR` is expected; none added unless the
   scan reveals fan_in ≥ 3.

## §E Self-Verification (design decisions)

- **D1 — Option A + guard (assessment, not an operator decision).** See spec.md §D; the lane lead
  decides from audit evidence. The plan below is written for A; M4 is the Option-B contingency.
- **D2 — guard anchors on the judgment target.** The guard compares two sets read from the same
  tree (production references vs scrub set ∪ reasoned exemptions); it does not mock the env and
  does not depend on the ambient env.
- **D3 — guard logic is a pure comparison over inputs**, so it is exercised on synthetic inputs
  (uncovered axis, exempted axis with reason, exemption with empty reason, empty reference set),
  and the same comparison runs once against the real tree. The synthetic cases are the
  re-executable RED once the guard is committed.
- **D4 — guard test is an ordinary test function, not `TestMain`** (spec.md §H O4). Liveness: the
  guard asserts its own swept count (non-zero references found, non-empty scrub set).
- **D5 — hook scrub-set variable is declared empty in c2** (so the tree compiles and the guard is
  red for the right reason: the set is empty against referenced axes) and filled in c4.
- **D6 — the whole-package arms are real arms.** The lane arm sets the full lane env explicitly;
  the scrubbed arm unsets the full family; both run the whole package with no `-run` selector, and
  the number of tests swept (`-json` rows) is recorded for both so an arm cannot pass vacuously.

## §F Milestones

### M1 — Baseline record, then the RED recurrence guard — Priority High
Files (exact):
- c1 (record only): `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/progress.md` — §E.2 baseline: the
  whole-package pairs of `internal/cli` and `internal/hook` on the pre-guard tree (4 leased runs),
  each with command, exit code, failing-test list, swept count, lease acquire/release lines; the
  hook child census; the discovery narrow pair (lane vs scrubbed).
- c2 (guard): `internal/cli/factory_env_axes_test.go` (new — the cli guard; reads `factoryAmbientEnvKeys`
  from `factory_test.go`); `internal/hook/lane_env_axes_test.go` (new — the hook guard, the
  **empty** scrub-set declaration, and the exemption table skeleton).
Steps: pre-flight (§C); record c1; write c2; run each guard under its package selector and record
the **red** output naming the uncovered axes, on a plain shell and under the lane env (it must be red
in both — it is ambient-independent); record the tree SHA of c2.
Exit condition: both guards red for the stated reason; c1 and c2 are distinct commits, c1 first.

### M2 — internal/cli scrub-set fix + measured diff — Priority High
Files: `internal/cli/factory_test.go` (extend `factoryAmbientEnvKeys` with `config.EnvFactoryRole`
and, per measurement and the guard, any of `EnvAutonomyTier`, `EnvFactoryClearPolicy`,
`EnvFactoryAutoDispatch`, `EnvMoaiFactoryManaged`; any axis left out is added to the guard's
exemption table with a reason in `internal/cli/factory_env_axes_test.go`).
Steps: apply; run the narrow AC command of AC-THE-001 under the explicit lane env → green with 3
`--- PASS`; guard green; the existing pin/clear tests (`TestFactoryAmbientEnvClearedInTestMain`,
`TestFactoryEnvPinnedSkipsTestMainClear`) green. Whole-package pairs are measured once at M4 on
the final tree; M2 does not run them.

### M3 — internal/hook fix — Priority High
Files: `internal/hook/lane_env_axes_test.go` (fill the scrub-set declaration; finalize the
exemption table), `internal/hook/main_test.go` (`TestMain` scrubs the set before the first test,
beside the existing `CLAUDE_PROJECT_DIR` scrub; a re-executed child composing env is honored through
the pin-marker pattern if the §C census found one).
Steps (first): isolate the responsible axis for the two StaleRunNotice tests by a one-axis-at-a-time
measurement **on the then-current tree** — five single-axis arms of the form `unset <AXIS> && go
test ./internal/hook -count=1 -v -run '^(TestStaleRunNoticeLegacyLeaderSpelling|TestStaleRunNoticeLegacySessionRecord)$'`
(plan-phase result, to be re-observed: `MOAI_KANBAN_ID` and `MOAI_FACTORY_WORKERS` each alone flip
the verdict; ROLE/WORKER/BACKEND do not) — record all five before choosing the set. Then apply and
run the narrow AC command of AC-THE-002 under the explicit lane env → green with 2 `--- PASS`; the
positive control and the guard green.

### M4 — Contingent per-test pins, remaining survey, whole-package pairs — Priority Medium
Files: only those measured red by the M1 baseline pairs that M2/M3 did not already fix (named in
progress.md §E.2 before editing); plus the final measurement record.
Steps: (1) final whole-package pairs on the post-M3 tree under the lease — cli lane arm, cli
scrubbed arm, hook lane arm, hook scrubbed arm — record failing sets and the difference (AC-THE-003);
(2) for any test still differing: RED/GREEN pair, Option-B pin with `t.Setenv` of every axis the
code path reads, `t.Parallel()` dropped where B1 applies, no skip; (3) `go vet` + Windows
`GOOS=windows GOARCH=amd64 go vet ./internal/cli ./internal/hook` + `golangci-lint run` delta vs
the §C baselines; (4) the AC-THE-005 / AC-THE-006 checks from `git log`; (5) populate
progress.md §E.2/§E.3.

### M5 — Sync — Priority Low
manager-docs owns `progress.md` §E.4 and the `implemented → completed` transition on the single
sync commit. Test-only change; no CHANGELOG entry is expected unless the sync phase judges one is
owed. The leader integrates per the git-flow lane protocol and pushes in batch.

## §G Anti-Patterns

- **AP-1** — skipping, deleting, or listing a flip-prone test (REQ-THE-004).
- **AP-2** — a "lane arm" that is only a partial env, or a "scrubbed arm" that unsets a subset
  (B4); both measure nothing.
- **AP-3** — declaring the sweep done from the static survey's list.
- **AP-4** — writing the fix first and "confirming" the guard red afterwards (REQ-THE-006).
- **AP-5** — citing a `-run` selector result as the whole-package verdict (AC-THE-003 requires the
  unselected package).
- **AP-6** — local `go test ./...`.
- **AP-7** — a guard whose reference scan matches nothing and passes (D4).
- **AP-8** — an exemption without a reason.
- **AP-9** — adding an axis to the scrub set because it is convenient, without checking that a
  test does not rely on its ambient value (spec.md §D precondition (i)).

## §H Cross-References

- spec.md §A evidence, §C REQ-THE-001..008, §D options, §E guard spec, §F ordering, §H open questions.
- acceptance.md §D — AC matrix and evidence ledger.
- `.claude/rules/moai/development/verification-completeness.md` §1.2, §1.3, §2, §2.1;
  `.claude/rules/moai/core/verification-claim-integrity.md` §2.3, §3.1;
  `.claude/rules/local/gitflow-lane-protocol.md` §8 (slot lease, verification scope).
- SPEC-CLI-TEST-CWD-ISOLATION-001 plan.md — structure and RED-first precedent.
