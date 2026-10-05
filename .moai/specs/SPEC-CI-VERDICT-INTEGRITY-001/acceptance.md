# acceptance.md — SPEC-CI-VERDICT-INTEGRITY-001

Tier M verification layer. Tree pin for every RED cell: **`a158b4b5f`** (document-level pin; binds all criteria). The evidence ledger (command / verbatim stdout / exit / tree per entry, E1-E18) lives in `plan.md` §B and is cited here by id — the ledger is the carrier per verification-completeness §2.1.

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

### AC-CI-005 — One overall deadline, strictly inside the job budget (RB; maps REQ-CI-005)

- **Given** the auto-merge job's timeout and its internal waits,
- **When** the workflow is authored,
- **Then** the checks wait and the review wait are both bounded by a single overall deadline that is strictly smaller than `timeout-minutes`, and the merge step is unreachable past it.
- RED: E5 (job 20 min at :25 vs 10-min checks wait at :96 + 15-min review wait at :159 = 25 min).
- Green path: M1 — E5's three-line read on the repaired file shows deadline ≤ 19 min total.

### AC-CI-006 — Install summary requires success from every needed job, parity included (RB; maps REQ-CI-006)

- **Given** the test-install summary job and its dependency set,
- **When** any needed job — `install-script-parity` included — concludes `cancelled`, `timed_out`, or `skipped`,
- **Then** the summary exits non-zero; it prints "All tests passed" only when every need concluded `success`.
- RED: E6 (no `cancelled` handling, exit 1) + E7 (all-`cancelled` substitution → `✅ All tests passed!`, exit 0) + E8 (needs list at :338 omits the job defined at :48).
- Green path: M1 — E7's substitution on the repaired summary exits non-zero; E8's grep shows parity in the needs list.

### AC-CI-007 — A missing required install feature fails the compatibility step (RB; maps REQ-CI-007)

- **Given** the compatibility-check feature probes (e.g. `MOAI_INSTALL_DIR`, `--help`, platform detection, `IsWindows`, `GetTempPath`),
- **When** a probe's target is absent from the install script,
- **Then** the step exits non-zero; printing a "missing" marker with exit 0 is not a verdict.
- RED: E9 (lines 309/317 — `grep -q … && echo ✓ || echo ✗` shape can only ever exit 0).
- Green path: M1 — required probes become hard checks; E9's shape no longer terminates in `\|\| echo` for required features.

### AC-CI-008 — SSoT lists only published check names (RB; maps REQ-CI-008)

- **Given** the live check-run names the workflows publish on main-bound PRs (E13: `Analyze (Go) (go)`, `Release PR Multi-OS Gate`, `Test (ubuntu-latest)`, Build × 5, `Lint`),
- **When** the corrected SSoT is read,
- **Then** `main` (and `release/*`, per decision-index Q1) contains exactly names from that published set, with `Test (macos-latest)`, `Test (windows-latest)`, and `CodeQL` absent, and `Release PR Multi-OS Gate` + `Analyze (Go) (go)` present.
- RED: E13 live JSON vs the current file values (three phantom contexts, one missing context — read evidence in plan §B corroboration).
- Green path: M2 — corrected file; re-read E13 after the next PR bearing the change and confirm no new phantom contexts (the check names in E13 come from the workflows' own `name:` fields + matrix values).

### AC-CI-009 — Protection apply is operator keep-set with GET re-verification (KS; maps REQ-CI-009)

- **Given** the corrected SSoT and the run-phase keep-set package (payload + apply command + pre-apply live-diff command + post-apply GET command),
- **When** the correction reaches live branch protection,
- **Then** the apply is executed by the operator/leader only, the pre-apply diff of live vs SSoT is recorded, and the post-apply GET read-back of `repos/modu-ai/moai-adk/branches/main/protection` — matching the corrected contexts — is the only accepted evidence of live state. No run-phase agent performs a protection write.
- RED: n/a — process gate; the prohibition is verified in run close by E6-scope evidence (no protection API calls in the agent transcript) and at card level by the recorded GET output.
- Green path: M2 keep-set subsection (plan §F).

### AC-CI-010 — Validator rejects parser absence and malformed YAML (RB; maps REQ-CI-010)

- **Given** the SSoT validator,
- **When** yq is unavailable, or the SSoT fails to parse,
- **Then** the validator exits non-zero with a message naming the parser failure; an empty parse result never produces a passing verdict.
- RED: E10 (yq absent → `✅ All validations passed`, exit 0) + E11 (malformed YAML → identical vacuous pass, exit 0).
- Green path: M2 — E10 and E11 re-run on the repaired validator → both non-zero.

### AC-CI-011 — Validator checks required-context publishability (RB; maps REQ-CI-010, REQ-CI-011)

- **Given** the corrected SSoT's required contexts,
- **When** the validator runs,
- **Then** it verifies each required context against the check names the workflows publish (derived from workflow `name:` fields and matrix values), and fails naming any context that no workflow can publish.
- RED: E10/E11 (no dimension inspects `branches.*.contexts` for publishability — only auxiliary mapping and ∩auxiliary overlap; an empty required list passes silently).
- Green path: M2 — a mutant SSoT carrying one phantom context (e.g. `Test (windows-latest)`) makes the validator exit non-zero.

### AC-CI-012 — Detect filter covers the parity test's input + correspondence guard (RB; maps REQ-CI-011)

- **Given** the branch-protection parity test reading `.github/branch-protection.json.gtmpl` (branch_protection_parity_test.go:41,57),
- **When** a PR touches only that file,
- **Then** ci.yml's detect filter matches it and the parity test runs; and a correspondence guard fails CI when any repo-side test's read input is absent from the filter.
- RED: E14 (count 0, exit 1) — a parity-drift edit to the local file alone skips every test job.
- Green path: M2 — filter gains the path; the guard test (new, under `internal/template/`) red-verified by temporarily removing the entry in a fixture.

### AC-CI-013 — ci-watch requests supported fields and classifies by bucket (RB; maps REQ-CI-012)

- **Given** the installed gh CLI's published `pr checks` field set (E12: `name, state, bucket, link, …`),
- **When** the ci-watch loop polls,
- **Then** it requests only supported fields, processes the response as one JSON array, classifies each check by `bucket`, and exits non-zero on any gh failure — a fetch failure is never read as all-pass.
- RED: E12 (`Unknown JSON field: "status"`, exit 1) — every poll aborts before any classification.
- Green path: M3 — E12's exact field list validates against the CLI (no field error) and the loop classifies a real PR read-only.

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
