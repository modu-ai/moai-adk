# acceptance.md — SPEC-AUTONOMY-CLOSURE-001

> Verification layer. Each AC is a binary Given-When-Then scenario bound to one Go test. Requirements
> live in spec.md §D; this file does not restate them.

## §A. Conventions

- **Test naming (binding on run phase):** every AC has exactly one top-level Go test named
  `TestAC_CLOSURE_<NNN>` (subtests allowed) in the owning package of the §B matrix.
- **Pass check (every AC):** from the worktree root, with `<NNN>` and `<pkg>` from the matrix row:

  ```
  go test -v -count=1 -run '^TestAC_CLOSURE_<NNN>' <pkg> > ac.out 2>&1; echo "exit=$?"
  grep -E '^--- PASS: TestAC_CLOSURE_<NNN> ' ac.out
  grep -E '^--- FAIL|\[no tests to run\]' ac.out; echo "bad=$?"
  ```

  The AC passes only when `exit=0`, the second command prints exactly one `--- PASS:` line for the
  named test, and the third prints nothing and `bad=1`. `[no tests to run]` is a FAIL.
- **Fixture:** a throwaway project under `t.TempDir()` with a git repository (`user.name` /
  `user.email` set), a queue store holding card `c1` → `SPEC-FIXTURE-001`,
  `.moai/specs/SPEC-FIXTURE-001/{spec.md,acceptance.md,contract.yaml,progress.md}` (contract signed
  through A1's signing seam), and `.moai/reports/c1/`. No AC reads or writes the real `.moai/`.
- **Report checks** read `closure-report.json`; "section X shows Y" means the JSON value at the
  section's key equals Y and the Markdown file contains the same text under the section heading.
- **Refusal checks:** "refuses" means the command exits with the stated code and every fixture file
  is byte-identical to before (SHA-256 before and after).

## §B. AC Matrix

| AC | Requirement | Owning package `<pkg>` |
|---|---|---|
| AC-CLOSURE-001 | REQ-CLOSURE-001 | `./internal/cli/` |
| AC-CLOSURE-002 | REQ-CLOSURE-002 | `./internal/closure/` |
| AC-CLOSURE-003 | REQ-CLOSURE-003 | `./internal/closure/` |
| AC-CLOSURE-004 | REQ-CLOSURE-004 | `./internal/closure/` |
| AC-CLOSURE-005 | REQ-CLOSURE-005 | `./internal/closure/` |
| AC-CLOSURE-006 | REQ-CLOSURE-006 | `./internal/closure/` |
| AC-CLOSURE-007 | REQ-CLOSURE-007 | `./internal/closure/` |
| AC-CLOSURE-008 | REQ-CLOSURE-008 | `./internal/closure/` |
| AC-CLOSURE-009 | REQ-CLOSURE-009 | `./internal/closure/` |
| AC-CLOSURE-010 | REQ-CLOSURE-010 | `./internal/closure/` |
| AC-CLOSURE-011 | REQ-CLOSURE-011 | `./internal/closure/` |
| AC-CLOSURE-012 | REQ-CLOSURE-012 | `./internal/cli/` |
| AC-CLOSURE-013 | REQ-CLOSURE-013 | `./internal/closure/` |
| AC-CLOSURE-014 | REQ-CLOSURE-014 | `./internal/closure/` |
| AC-CLOSURE-015 | REQ-CLOSURE-015 | `./internal/hook/` |
| AC-CLOSURE-016 | REQ-CLOSURE-016 | `./internal/closure/` |
| AC-CLOSURE-017 | REQ-CLOSURE-017 | `./internal/hook/` |
| AC-CLOSURE-018 | REQ-CLOSURE-018 | `./internal/closure/` |
| AC-CLOSURE-019 | REQ-CLOSURE-019 | `./internal/cli/` |
| AC-CLOSURE-020 | REQ-CLOSURE-020 | `./internal/cli/` |
| AC-CLOSURE-021 | REQ-CLOSURE-021 | `./internal/hook/` |
| AC-CLOSURE-022 | REQ-CLOSURE-022 | `./internal/closure/` |
| AC-CLOSURE-023 | REQ-CLOSURE-023 | `./internal/hook/` |
| AC-CLOSURE-024 | REQ-CLOSURE-024 | `./internal/closure/` |
| AC-CLOSURE-025 | REQ-CLOSURE-025 | `./internal/template/` |
| AC-CLOSURE-026 | REQ-CLOSURE-002, §E determinism | `./internal/closure/` |

## §C. Criteria

### AC-CLOSURE-001 — Report command writes the pair or refuses

- **Given** the fixture, **When** `moai contract report c1` runs, **Then** it exits 0, prints the path
  of `.moai/reports/c1/closure-report.md`, and both `closure-report.md` and `closure-report.json`
  exist.
- **Given** a card `c2` absent from the queue, a card `c3` with no SPEC ID, and a card `c4` whose SPEC
  has no `contract.yaml`, **When** the command runs for each, **Then** each exits 2 naming the missing
  input and no file is created under `.moai/reports/`.

### AC-CLOSURE-002 — Section order and form parity

- **Given** a generated report, **When** its JSON top-level keys are read in order, **Then** the
  section keys appear in the order Summary, Kickoff, Reconciliation, Invariants, Ownership, New APIs,
  Escalations, First Verdict, Second Verdict, Plan-Audit Binding, Not Performed, Residual Risk, Human
  Verdict, `schema_version` equals 1, and the Markdown headings appear in the same order with the same
  values.

### AC-CLOSURE-003 — Missing inputs are listed, never passed

- **Given** the fixture with `progress.md`, the escalation directory, `second-review.jsonl`, and
  `closure-verdict.jsonl` removed, **When** the report is built, **Then** no section renders a pass
  state, every affected row renders `not observed` or `not recorded`, and `not_performed` contains
  `progress-missing`, `second-review-not-performed`, `escalation-detector-not-armed`, and
  `human-verdict-none`.

### AC-CLOSURE-004 — Contract reconciliation per AC

- **Given** a fixture `acceptance.md` with three live criterion IDs (first, second, third) and a
  `progress.md` §E.2 matrix with rows for the first ID (PASS) and for a fourth ID absent from
  `acceptance.md` (PASS), **When** the report is built, **Then** Reconciliation shows the first ID
  PASS with its evidence text verbatim, the second and third as `not reported`, the fourth as
  `unknown`, and the recorded versus measured acceptance hash and AC count with the A1 verify state.

### AC-CLOSURE-005 — Invariant results

- **Given** a contract with invariants `constitution:X-*`, `frozen-files`, `go test ./pkg/...`, and
  `make check`, an open `invariant-violation` record pointing at `go test ./pkg/...`, a record listing
  `make check` as not observed, and an armed card state file, **When** the report is built, **Then**
  the four invariants show `not observed`, `no violation recorded`, `violation recorded (open)`, and
  `not observed` respectively.
- **Given** the same contract with the card state file absent, **When** the report is built, **Then**
  every invariant without a record shows `not observed`.

### AC-CLOSURE-006 — Ownership count

- **Given** two `ownership-move` records (one open, one resolved) and an armed, not-disarmed card
  state file, **When** the report is built, **Then** Ownership lists both records and the count is 2.
- **Given** no `ownership-move` record and an armed state file, **Then** the count is 0.
- **Given** no record and a disarmed state file, **Then** the count is `not observed`.

### AC-CLOSURE-007 — New APIs

- **Given** a comparison seam returning two additions (`package internal/newpkg`,
  `exported func Foo`), **When** the report is built, **Then** New APIs lists both by kind and path and
  the escalation directory is byte-identical before and after.
- **Given** a comparison seam returning an error, **When** the report is built, **Then** New APIs
  states `not observed` and `not_performed` contains `new-api-comparison-unavailable`.

### AC-CLOSURE-008 — Escalations and needs-decision

- **Given** records of kind `contract` (open), `operational` (resolved), `revoke` (resolved), and one
  file with malformed frontmatter, **When** the report is built, **Then** the three parsed records are
  listed with kind, class, status, occurrences, contract reference, and decider, needs-decision is
  `yes`, and the malformed file is listed as `unreadable` with `escalation-unreadable` in
  `not_performed`.
- **Given** only the `operational` (resolved) and `revoke` records, **Then** needs-decision is `no`.

### AC-CLOSURE-009 — First verdict

- **Given** a §E.3 block with `run_status: complete`, `ac_pass_count: 2`, `ac_fail_count: 0`,
  `run_commit_sha: abc1234` and a measured AC count of 3, **When** the report is built, **Then** First
  Verdict shows the four values verbatim and a count mismatch flag.
- **Given** a §E.3 block without `run_commit_sha`, **Then** that field shows `not recorded`.

### AC-CLOSURE-010 — Kickoff display is tolerant

- **Given** four receipt fixtures — (a) decider `llm` with one answer; (b) decider `llm+jev` with two
  answers (llm, jev); (c) requested `jev`, signer `llm`, `fallback_reason: jev_low_confidence`, a
  `jev_attempt` with confidence 0.30; (d) receipt (a) plus an unknown top-level field — **When** the
  report is built for each, **Then** Kickoff shows the decider token verbatim, every answer with
  decider, answer, and confidence, fallback `yes` with its reason for (c) and `no` otherwise, the A1
  verify status, and (d) renders the same as (a).
- **Given** a contract signed with method `receipt` and no receipt file, **Then** Kickoff states
  `missing`; **Given** a contract signed on the human path, **Then** Kickoff shows method
  `interactive-tty`, signer kind `human`, and no receipt.

### AC-CLOSURE-011 — Plan-audit binding

- **Given** contract `plan_audit.verdict: PASS` and `.moai/reports/plan-audit/SPEC-FIXTURE-001-review-2.md`
  ending with `AUDIT-VERDICT: PASS spec=SPEC-FIXTURE-001 receipts=none` (and a `-review-1.md` ending
  with FAIL), **When** the report is built, **Then** the binding is `bound` and names `-review-2.md`.
- **Given** `-review-2.md` ending with FAIL, **Then** the binding is `mismatch`.
- **Given** no plan-audit report, **Then** the binding is `self-reported`.
- **Given** a receipt whose `inputs.plan_audit_report` names `-review-1.md` with a hash differing from
  the file, **Then** the binding is `mismatch`.

### AC-CLOSURE-012 — `audit_multi` second-review record

- **Given** backends stubbed to return claude `pass`, codex `fail`, glm `inconclusive`, **When**
  `audit_multi` is called with `card_id: c1` and `project_root` set to the fixture, **Then**
  `.moai/reports/c1/second-review.jsonl` gains exactly one line whose `card`, `spec_id`,
  `contract_sha256`, `head_sha`, `backends`, `participant_count`, `disagreement_flag`, and
  `recorded_at` match the fixture and the returned result.
- **Given** the same stubs and no `card_id`, **When** `audit_multi` is called, **Then** the returned
  JSON is byte-identical to the result produced by the pre-change handler for the same input (golden
  file) and no file exists under `.moai/reports/`.

### AC-CLOSURE-013 — Performed determination

- **Given** a table of `second-review.jsonl` fixtures against report HEAD `H` and digest `D`, **When**
  the state is computed, **Then**: no file → `not performed / no-record`; latest line head ≠ `H` →
  `not performed / stale-head`; digest ≠ `D` → `not performed / contract-changed`; only claude counted
  → `not performed / no-second-model`; digest `""` → `not performed / unbound`; codex `pass` + glm
  `inconclusive` → `performed / pass`; codex `pass` + glm `fail` → `performed / fail`.

### AC-CLOSURE-014 — Second verdict rendering

- **Given** a not-performed state with cause `no-record`, **When** the report is rendered, **Then** the
  Markdown Second Verdict section contains the literal `NOT PERFORMED` followed by `no-record`.
- **Given** a performed review by `glm` while the contract names `codex`, **Then** the section shows
  the backend, its verdict, the contract's `second_model`, and a substitute marker.

### AC-CLOSURE-015 — Push stop in the hook

- **Given** `mode: contract`, `second_review: required`, an in-range contract listing `push-develop`
  with a closure report and no second-review record, **When** the PreToolUse hook receives the Bash
  call `git push origin develop`, **Then** the decision is deny and the reason begins
  `CLOSURE_PUSH_STOP:` and contains `second_review_not_performed`.
- **Given** the same fixture with a performed `pass` record for the current HEAD and digest, **Then**
  the hook does not deny on A4's account.
- **Given** the not-ready fixture and the call `git push origin WT-feature`, **Then** no readiness
  evaluation runs (evaluation counter seam stays 0).

### AC-CLOSURE-016 — Readiness code set

- **Given** one fixture per code, **When** the evaluator runs, **Then** each of `contract_invalid`,
  `closure_report_missing`, `second_review_not_performed`, `second_review_stale`,
  `second_review_failed`, `push_check_undetermined`, and the human-verdict codes fixed by OQ-1 is
  produced by its fixture, and the evaluator's declared code list equals exactly this set.
- **Given** a fixture raising two codes, **Then** both are reported, sorted and de-duplicated.

### AC-CLOSURE-017 — Undetermined denies

- **Given** `mode: contract` and a git seam returning an error for the range listing, **When** the hook
  receives `git push origin develop`, **Then** the decision is deny with `push_check_undetermined`.
- **Given** an in-range contract whose card has no worktree and no primary-checkout evidence
  directory, **Then** the decision is deny with `push_check_undetermined`.

### AC-CLOSURE-018 — Advisory and off

- **Given** the not-performed fixture with `second_review: advisory`, **When** the evaluator runs,
  **Then** no second-review code is reported, and the report renders `NOT PERFORMED` plus a warning
  line.
- **Given** `second_review: off`, **Then** no second-review code is reported and Second Verdict states
  `not required`.

### AC-CLOSURE-019 — `push-check` CLI

- **Given** a ready fixture, **When** `moai contract push-check` runs, **Then** it exits 0 and prints
  `ready`.
- **Given** the not-ready fixture, **Then** it exits 1 and prints `SPEC-FIXTURE-001` with its codes.
- **Given** an unknown flag, **Then** it exits 2.
- **Given** `mode: guided`, **Then** it exits 0 and prints one line containing `inactive`.

### AC-CLOSURE-020 — Human verdict recording

- **Given** each of: an agent-environment marker set; a non-terminal standard input; a typed
  confirmation that differs; no closure report; **When** `moai contract verdict c1 accept` runs,
  **Then** each refuses with exit 1 and `closure-verdict.jsonl` does not exist.
- **Given** a terminal seam, markers cleared, and the typed token `accept c1`, **Then** it exits 0
  and `closure-verdict.jsonl` gains one line with `verdict: accept`, operator name and email from git
  configuration, `method: interactive-tty`, and `report_sha256` equal to the SHA-256 of the current
  `closure-report.json`.

### AC-CLOSURE-021 — Verdict command denied to agents

- **Given** `mode: guided` and then `mode: contract`, **When** the PreToolUse hook receives the Bash
  call `moai contract verdict c1 accept`, **Then** both decisions are deny with a reason beginning
  `CLOSURE_VERDICT_HUMAN_ONLY:`.
- **Shell check:** `grep -rn 'closure-verdict.jsonl' internal/ --include='*.go' | grep -v _test.go`
  lists only files under `internal/closure/` (the decoder and its path constant) and
  `internal/cli/contract_verdict.go`.

### AC-CLOSURE-022 — Human verdict section

- **Given** a verdict line whose `report_sha256` equals the report being rendered, **Then** Human
  Verdict shows the verdict, operator, time, and `current`.
- **Given** the report regenerated after the verdict with a different hash, **Then** it shows `stale`
  and `not_performed` contains `human-verdict-stale`.
- **Given** two verdict lines, **Then** the later one is shown.

### AC-CLOSURE-023 — Guided mode is a no-op

- **Given** `mode: guided` (and, separately, the `autonomy` section absent and `mode: bogus`), **When**
  the PreToolUse hook receives a corpus of Bash calls including `git push origin develop`,
  `git status`, and `go test ./...`, **Then** each output is byte-identical to the output of the hook
  built without A4's checks (golden), and the subprocess seam records zero A4 invocations.

### AC-CLOSURE-024 — Evidence resolution order

- **Given** `second-review.jsonl` present in both the card worktree and the primary checkout with
  different content, and `closure-verdict.jsonl` only in the primary checkout, **When** the report is
  built, **Then** the second review is read from the card worktree, the verdict from the primary
  checkout, and `sources` names both paths.

### AC-CLOSURE-025 — Template neutrality

- **Given** the template copies of the `moai-ref-cross-model-audit` skill and the `sync-auditor`
  agent, **When** the test reads them, **Then** each contains an instruction to pass the card argument
  to `audit_multi` in contract mode, and the text between the A4 markers contains no match for
  `SPEC-[A-Z]`, `\bt[0-9]{3,5}\b`, `20[0-9]{2}-[0-9]{2}-[0-9]{2}`, `A-Q[0-9]`, or `\b[0-9a-f]{9,40}\b`.
- **Shell check:** `make agents-emit-check` exits 0.

### AC-CLOSURE-026 — Determinism

- **Given** the fixture, **When** the report is built twice, **Then** the two JSON outputs are
  byte-identical after removing `generated_at`.

## §D. Edge Cases

- A card with two queue entries for the same SPEC: evidence read per card; the push evaluator
  reports `push_check_undetermined` when none carries a closure report (AC-CLOSURE-017).
- A terminal SPEC's contract never enters the push set (design.md §C.2).
- A `second-review.jsonl` line with an unknown `schema_version` is skipped and listed under Not
  Performed; it never counts as performed.

## §E. Definition of Done

- All 26 ACs pass under the §A pass check, verbatim output in progress.md §E.2.
- plan.md §E E1-E6 hold.
- OQ-1 and OQ-2 resolved and reflected in spec.md / design.md before run-phase M2.
- No write path to `verdict.md` (review of the diff).
