# acceptance.md — SPEC-CI-VERDICT-INTEGRITY-001

Tier M verification layer. Tree pin for every RED cell: **`a158b4b5f`** (document-level pin; binds all criteria; measured files verified unchanged through `861a3ba56`). The evidence ledger (command / verbatim stdout / exit / tree per entry, E1-E24, with fenced verbatim-stdout blocks L-E2..L-E24) lives in `plan.md` §B and is cited here by id — the ledger is the carrier per verification-completeness §2.1. The workflow probes (E2/E4/E7) are committed `repro/` scripts that EXTRACT the step body from the live workflow at run time and substitute the declared dependency values — the substitution values are simulated (no runner execution), the script logic under test is the live file's.

Classification: **release-blocking (RB)** — RED re-executable on this tree, flips with this SPEC's work; **keep-set (KS)** — operator-executed, process criterion; probes that are template-substitution simulations of a gate script (E2, E4, E7) are marked as such inside the ledger and are treated as faithful RED evidence of the script logic, not of runner execution.

## §B. Acceptance criteria

### AC-CI-001 — Release gate fails on a cancelled or timed-out matrix (RB; maps REQ-CI-001)

- **Given** a release PR whose full matrix ended `cancelled` (or `timed_out`) while detect succeeded,
- **When** the `Release PR Multi-OS Gate` job evaluates its dependencies,
- **Then** the gate exits non-zero and emits an error annotation naming the non-success conclusion.
- RED: E1 (file contains no `cancelled` handling, exit 1) + E2 (faithful substitution of the current gate script with `full-matrix-test.result = "cancelled"` → PASSED, exit 0).
- Green path: M1 — gate requires `success` from detect AND matrix; re-running E2's substitution against the repaired script shape yields non-zero. Paired with AC-CI-002 so an always-fail mutant cannot satisfy this criterion alone.

### AC-CI-002 — Intentional exclusions produce a distinct, reason-bearing output (RB; maps REQ-CI-001, REQ-CI-002)

- **Given** a non-release PR (detect `skipped`) or a docs-only release PR (matrix skipped by a positively observed filter output),
- **When** the gate passes,
- **Then** it emits a distinct output naming the exclusion and its reason, and the unconditional "verification PASSED" wording fires only on a genuine success; a detect hard `failure` still fails the gate.
- RED: E2 — the single PASSED line currently serves every case (pass, docs-only skip, non-release skip, and — before E1's repair — cancelled).
- Green path: M1 — separate output branches; verified by reading the repaired gate step plus actionlint clean (E18 stays exit 0).

### AC-CI-003 — Merge is pinned to the CI-verified head (RB; maps REQ-CI-003)

- **Given** the auto-merge workflow verified head SHA S at guard time, and any push lands before the merge call,
- **When** the merge step runs,
- **Then** the merge command carries S as a hard match precondition, the step re-reads the PR head immediately before merging, and a mismatch or unreadable head withholds the merge (PR left open).
- RED: E3 (`match-head-commit` count 0, exit 1) with the merge command at auto-merge.yml:275 carrying no SHA condition (read evidence, plan §B).
- Green path: M1 — merge call passes the pin (flag exists: E17) and adds the immediate re-read; verified by reading the repaired step.

### AC-CI-004 — A failed, empty, or incomplete required-checks lookup withholds the merge (RB; maps REQ-CI-004)

- **Given** the checks lookup fails, returns no checks, or returns fewer checks than the target branch's protection requires,
- **When** the auto-merge checks step concludes,
- **Then** it outputs `should_merge=false` and records the incomplete observation; only a positively complete, successful lookup with zero pending can yield `should_merge=true`.
- RED: E4 (stub lookup failure with a real-shaped message → `should_merge=true`, exit 0).
- Green path: M1 — re-run E4's stub against the repaired loop → withholds.

### AC-CI-005 — One overall deadline, re-judged at merge time, immune to the merge-after-deadline mutant (RB; maps REQ-CI-005)

- **Given** the auto-merge job's timeout and its internal waits,
- **When** the workflow is authored,
- **Then** the checks wait and the review wait are both bounded by a single overall deadline strictly smaller than `timeout-minutes`, AND the merge step re-evaluates the current deadline and the current head commit at merge time; an implementation that compares only state captured before the waits does not satisfy this criterion.
- RED: E5 (job 20 min at :25 vs 10-min checks wait at :96 + 15-min review wait at :159 = 25 min) plus the auditor's iteration-1 mutant probe, recorded as the shape this criterion must catch: with the current tree's absent deadline logic, a simulated `declared_deadline=1140, merge_at=1141` merge SUCCEEDS (exit 0) — merge-after-deadline is never withheld today.
- Green path: M1 — the flip evidence is a mutant-style probe against the repaired workflow: a simulated merge arriving at `deadline + 1` must be WITHHELD (non-merge outcome with a notice), and a merge inside the deadline must proceed; both observations recorded in run-phase §E.2. A repair whose green shows only the in-deadline case has not flipped this criterion.

### AC-CI-006 — Install summary requires success from every needed job, parity included (RB; maps REQ-CI-006)

- **Given** the test-install summary job and its dependency set,
- **When** any needed job — `install-script-parity` included — concludes `cancelled`, `timed_out`, or `skipped`,
- **Then** the summary exits non-zero; it prints "All tests passed" only when every need concluded `success`.
- RED: E6 (no `cancelled` handling, exit 1) + E7 (all-`cancelled` substitution → `✅ All tests passed!`, exit 0) + E8 (needs list at :338 omits the job defined at :48).
- Green path: M1 — E7's substitution on the repaired summary exits non-zero; E8's grep shows parity in the needs list.

### AC-CI-007 — A missing required install feature fails the compatibility step (RB; maps REQ-CI-007)

- **Given** the compatibility-check feature probes, each targeting a feature its script implements (install.sh: `--install-dir`, `--help`, darwin/linux detection; install.ps1: `IsWindows`, `GetTempPath`, `MOAI_INSTALL_DIR`),
- **When** a probe's target is absent from its install script,
- **Then** the step exits non-zero; printing a "missing" marker with exit 0 is not a verdict; and the probe set carries no stale entry — a probe targeting a feature the script does not implement is realigned to the script's real surface within this milestone, not hardened into a guaranteed failure.
- RED: E9 (lines 309/317 — `grep -q … && echo ✓ || echo ✗` shape can only ever exit 0) + E19 (the :309 install.sh probe targets `MOAI_INSTALL_DIR`, which install.sh does not implement — 0 hits, exit 1; a hard-check repair without the probe realignment would fail every install, the impossible direction the two-cell rule forbids). E20 confirms the :317 install.ps1 probe is live (3 hits, exit 0).
- Green path: M1 — required probes become hard checks AND the stale :309 probe is realigned to install.sh's real surface (`--install-dir`); E9's shape no longer terminates in `\|\| echo` for required probes and no probe targets an unimplemented feature.

### AC-CI-008 — SSoT lists only published check names, per branch key (RB; maps REQ-CI-008)

- **Given** the live check-run names the workflows deterministically publish on main-targeting PRs (E13 refined: `Analyze (Go) (go)`, `Release PR Multi-OS Gate`, `Test (ubuntu-latest)`, Build × 5, `Lint` — the `CodeQL` and bare `Analyze (Go)` names in E13 come from non-PR-trigger and skipped-matrix runs) and the trigger facts (E21: all three PR triggers are `branches: [main]`, so release-targeting PRs publish none of them),
- **When** the corrected SSoT is read,
- **Then** the `main` key contains exactly names from that deterministically-published set — with `Test (macos-latest)`, `Test (windows-latest)`, and `CodeQL` absent, and `Release PR Multi-OS Gate` + `Analyze (Go) (go)` present — and the `release/*` key carries only what release-targeting PRs actually publish: an empty `contexts:` list under the current triggers (decision-index Q1, amendment-2 resolution; trigger expansion is out via the t1536 boundary).
- RED: E13 (live unique names vs the current file's three phantom contexts and one missing context — read evidence in plan §B corroboration) + E21 (trigger restriction facts).
- Green path: M2 — corrected file per the revised Q1; the operator GET re-verification re-confirms both keys against live check runs and live protection.

### AC-CI-009 — Protection apply is operator keep-set with GET re-verification (KS; maps REQ-CI-009)

- **Given** the corrected SSoT and the run-phase keep-set package (corrected file + rendered payload + apply command + pre-apply live-diff command + post-apply GET command, delivered at `.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md`),
- **When** the correction reaches live branch protection,
- **Then** the apply is executed by the operator/leader only, and the post-apply GET read-back of `repos/modu-ai/moai-adk/branches/main/protection` — matching the corrected contexts — is the only accepted evidence of live state. No run-phase agent performs a protection write.
- RED: E23 — the apply package does not exist on this tree (`test -e .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md` → exit 1, file absent); corroborated by the packaging surfaces it builds on being absent (E24's phantom silent-pass, E10's absent parser rejection). Flips when the run-phase package is delivered.
- Green path: M2 keep-set subsection (plan §F). Evidence list: (1) the delivered `apply-package.md`; (2) the operator-pasted TRANSCRIPT of the executed apply commands AND the GET read-back output into the card evidence path `.moai/reports/t1534/` — the transcript is the non-execution + live-state proof, because `git status` cannot observe an API call; (3) run-phase E6 scope evidence as a supporting check only.

### AC-CI-010 — Validator rejects parser absence and malformed YAML (RB; maps REQ-CI-010)

- **Given** the SSoT validator,
- **When** yq is unavailable, the SSoT fails to parse, or the validator runs against the repo's real, healthy SSoT,
- **Then** BOTH directions hold: on each failure shape the validator exits non-zero with a message naming the parser failure and an empty parse result never produces a passing verdict; on the healthy input the validator exits 0 with every dimension executing on real content. An always-fail mutant (prints "parser failure", exits 1 unconditionally) fails the healthy direction; an always-pass mutant fails the failure direction — the criterion is two-directional.
- RED: E10 (yq absent → vacuous `✅ All validations passed`, exit 0) + E11 (malformed YAML → identical vacuous pass, exit 0) for the failure direction; E22 (healthy SSoT + yq → every dimension runs, `✅ All validations passed`, exit 0) for the healthy direction, measured pre-repair so its post-repair green is interpretable.
- Green path: M2 — E10 and E11 re-run on the repaired validator → both non-zero; E22 re-run → still exit 0 with the dimensions executing (a repaired validator that broke the healthy path is a regression, not a fix).

### AC-CI-011 — Validator checks required-context publishability (RB; maps REQ-CI-010, REQ-CI-011)

- **Given** the corrected SSoT's required contexts,
- **When** the validator runs,
- **Then** it verifies each required context against the check names the pull_request-triggered workflows deterministically publish (derived from workflow `name:` fields and matrix values), and fails naming any context that no workflow can publish on that branch's PRs.
- RED: E24 — a VALID-YAML SSoT carrying a phantom required context (`Test (windows-latest)`) passes silently, exit 0 (committed fixture `repro/phantom/`): no dimension inspects `branches.*.contexts` for publishability, so this observation is red because the publishability dimension itself is missing — a parser-handling-only repair (fixing E10/E11) leaves E24 red, which is exactly why the phantom is the right RED for this criterion and E10/E11 are not. (E10/E11 remain AC-CI-010's parser-direction RED.)
- Green path: M2 — E24 re-run on the repaired validator → non-zero naming the phantom context; E22 (healthy SSoT) → still exit 0.

### AC-CI-012 — Detect filter covers the parity test's input + correspondence guard (RB; maps REQ-CI-011)

- **Given** the branch-protection parity test reading `.github/branch-protection.json.gtmpl` (branch_protection_parity_test.go:41,57),
- **When** a PR touches only that file,
- **Then** ci.yml's detect filter matches it and the parity test runs; and a correspondence guard fails CI when any repo-side test's read input is absent from the filter.
- RED: E14 (count 0, exit 1) — a parity-drift edit to the local file alone skips every test job.
- Green path: M2 — filter gains the path; the guard test (new, under `internal/template/`) red-verified by temporarily removing the entry in a fixture.

### AC-CI-013 — ci-watch requests supported fields and classifies by bucket (RB; maps REQ-CI-012)

- **Given** the installed gh CLI's published `pr checks` field set (E12: `name, state, bucket, link, …`),
- **When** the ci-watch loop polls,
- **Then** it requests only supported fields — the repaired poll command is `gh pr checks <PR> --json name,state,bucket,link`, which validates against the CLI (no field error, unlike E12's rejected `name,status,conclusion,detailsUrl` list) — processes the response as one JSON array, classifies each check by `bucket`, and exits non-zero on any gh failure — a fetch failure is never read as all-pass.
- RED: E12 (`Unknown JSON field: "status"` on stderr, stdout empty, exit 1) — every poll aborts before any classification.
- Green path: M3 — the repaired field list `name,state,bucket,link` validates against the CLI (E12's rerun with THAT list is the flip probe: no field error, exit per the PR's real state) and the loop classifies a real PR read-only. The unrepaired list still exits 1 — the criterion flips only when the repair lands.

### AC-CI-014 — ci-watch counts absent required checks as pending and splits no names (RB; maps REQ-CI-013)

- **Given** the SSoT required list for the PR's base branch and a poll response,
- **When** the loop classifies the tick,
- **Then** required checks are determined via `is_required`, an expected required check absent from the response counts as pending (never as pass), check names are iterated without whitespace splitting, and `gh pr checks` is invoked without `--required` ambiguity — classification comes from the SSoT, not from name substrings.
- RED: E15 (`is_required` unused, exit 1) + E16 (unquoted substitution at :134) + the structural read (the loop iterates only names present in the JSON, so absent required checks are invisible).
- Green path: M3 — run_test.sh extended: a fixture JSON missing one SSoT-required check yields `pending ≥ 1` and no exit 0; a spaced-name fixture classifies as one check.

## §C. Edge cases

1. Detect job hard `failure` (not skipped) — gate must fail (release-pr-multi-os.yml already documents this intent; the repaired gate preserves it).
2. `skipped` as a matrix conclusion from an intentional exclusion vs from a cancelled sibling — the gate distinguishes via the filter's positive output, not via the conclusion string alone.
3. `--delete-branch` combined with the merge pin — verify the repaired merge call keeps branch deletion behavior while pinning the head (gh accepts both; E17).
4. SSoT `auxiliary:` entries — M2's publishability dimension must not start requiring auxiliary checks (they stay advisory; ci-watch-protocol).
5. A required check that legitimately does not run on a PR type (e.g. the gate on a workflow_dispatch-only run) — the missing-required-pending rule (AC-CI-014) must key off the PR's base branch context, not blindly on all runs.
6. gh version drift — the repaired field list is validated against the CLI at pre-flight; a future CLI that drops a field fails loudly (exit 1), which is the intended fail-closed direction.
7. A probe target the script does not implement (found pre-implementation: install.sh `MOAI_INSTALL_DIR`, E19) — M1-5 realigns the probe list to the scripts' real surface; a hard-check repair alone would fail every install. The docs-site mirror copy (`docs-site/static/install.sh`) matches install.sh (0 hits), so no divergence case arises from it.

## §D. Quality gates (TRUST 5)

- **Tested**: `go test -timeout 30m ./internal/template/...` green (parity + guard test, ≥ 85% on new Go test code); `scripts/ci-watch/test/run_test.sh` green with the new fixtures; validator E10/E11 probe pairs flipped; gate probes E2/E7 re-run red→green.
- **Readable/Unified**: shell edits pass `actionlint` (E18 stays clean) and keep the surrounding style (existing comment conventions in the workflows); no reformatting of untouched hunks.
- **Secured**: no secrets introduced; the keep-set package carries commands, not tokens; protection writes stay operator-held.
- **Trackable**: Conventional Commits per milestone with `card t1534` and the `🗿 MoAI` trailer; AC matrix traceable REQ↔AC (13 REQs ↔ 14 ACs).

## §E. Definition of Done

1. AC-CI-001..008 and AC-CI-010..014 all green with their RED cells flipped and evidence recorded (run-phase §E.2).
2. AC-CI-009 package delivered to the operator; live apply + GET read-back recorded in card evidence `.moai/reports/t1534/` by the operator/leader.
3. Scope proof: only in-scope files changed; `internal/template/templates/**` untouched (AC-CI-012's mirror file is read, not written — the parity test compares the local file against the existing mirror).
4. `actionlint` clean; affected-package tests green; no local full-suite runs.
