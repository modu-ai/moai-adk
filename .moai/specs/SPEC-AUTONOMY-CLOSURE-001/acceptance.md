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
- **Fixture:** a throwaway repository under `t.TempDir()` (`user.name` / `user.email` set) with an
  integration branch `develop`, a bare remote holding `origin/develop`, a linked worktree directory
  named `c1` on a card branch, a queue store mapping card `c1` → `SPEC-FIXTURE-001`, and
  `.moai/specs/SPEC-FIXTURE-001/{spec.md,acceptance.md,contract.yaml,progress.md}` whose contract
  carries `card: c1`, lists `push-develop`, has `ownership.write: [src/**, .moai/specs/SPEC-FIXTURE-001/**]`,
  and is signed through A1's signing seam. The card evidence directory is `<c1 worktree>/.moai/reports/c1/`.
  No AC reads or writes the real `.moai/`.
- **Report checks** read `closure-report.json`; "section X shows Y" means the JSON value at the
  section's key equals Y and the Markdown file contains the same text under the section heading.
- **Refusal checks:** "refuses" means the stated exit code and every fixture file byte-identical to
  before (SHA-256 before and after).

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
| AC-CLOSURE-024 | REQ-CLOSURE-024 | `./internal/cli/` |
| AC-CLOSURE-025 | REQ-CLOSURE-025 | `./internal/template/` |

## §C. Criteria

### AC-CLOSURE-001 — Report command writes the pair or refuses

- **Given** the fixture, **When** `moai contract report c1` runs, **Then** it exits 0, prints the path
  of `<c1 worktree>/.moai/reports/c1/closure-report.md`, and both files of the pair exist there.
- **Given** a card `c2` absent from the queue, a card `c3` with no SPEC ID, a card `c4` whose SPEC has
  no `contract.yaml`, and a card `c5` mapped to a SPEC whose contract carries `card: c1`, **When** the
  command runs for each, **Then** each exits 2 naming the cause and no file is created under any
  `.moai/reports/`.

### AC-CLOSURE-002 — Section order, form parity, determinism

- **Given** a generated report, **When** its JSON top-level keys are read in order, **Then** the
  section keys appear in the order Summary, Kickoff, Reconciliation, Invariants, Ownership, New APIs,
  Escalations, First Verdict, Second Verdict, Plan-Audit Binding, Not Performed, Residual Risk, Human
  Verdict, `schema_version` equals 1, and the Markdown headings appear in the same order with the same
  values.
- **Given** the fixture, **When** the report is built twice, **Then** the two JSON outputs are
  byte-identical after removing `generated_at`.

### AC-CLOSURE-003 — Missing inputs are listed, never passed

- **Given** the fixture with `progress.md`, the escalation directory, the A2 card state file,
  `second-review.jsonl`, and `closure-verdict.jsonl` absent, **When** the report is built, **Then** no
  section renders a pass state, every affected row renders `not observed` or `not recorded`, and
  `not_performed` contains `progress-missing`, `second-review-not-performed`,
  `escalation-detector-not-armed`, and `human-verdict-none`.

### AC-CLOSURE-004 — Contract reconciliation per AC

- **Given** a fixture `acceptance.md` with three live criterion IDs (first, second, third) and a
  `progress.md` §E.2 matrix with rows for the first ID (PASS) and for a fourth ID absent from
  `acceptance.md` (PASS), **When** the report is built, **Then** Reconciliation shows the first ID
  PASS with its evidence text verbatim, the second and third as `not reported`, the fourth as
  `unknown`, and the recorded versus measured acceptance hash and AC count with the A1 verify state.

### AC-CLOSURE-005 — Invariant results never claim an unobserved pass

- **Given** invariants `constitution:X-*`, `frozen-files`, `go test ./pkg/...`, and `make check`, an
  open `invariant-violation` record naming `go test ./pkg/...`, and an armed, not-disarmed A2 card
  state file for the contract in force, **When** the report is built, **Then** the four invariants
  show `not observed`, `no violation recorded`, `violation recorded (open)`, and `not observed`.
- **Given** the same contract with the card state file disarmed, **Then** `frozen-files` shows
  `not observed` and no invariant shows `no violation recorded`.

### AC-CLOSURE-006 — Ownership count

- **Given** two `ownership-move` records (one open, one resolved) and an armed, not-disarmed card state
  file, **Then** Ownership lists both and the count is 2; **Given** no record and an armed state file,
  **Then** the count is 0; **Given** no record and a disarmed state file, **Then** the count is
  `not observed`.

### AC-CLOSURE-007 — New APIs

- **Given** a comparison seam returning two additions, **When** the report is built, **Then** New APIs
  lists both by kind and path and the escalation directory is byte-identical before and after.
- **Given** a comparison seam returning an error, **Then** New APIs states `not observed` and
  `not_performed` contains `new-api-comparison-unavailable`.

### AC-CLOSURE-008 — Escalations and needs-decision

- **Given** records of kind `contract` (open), `operational` (resolved), `revoke` (resolved), and one
  file with malformed frontmatter, **Then** the three parsed records are listed with kind, class,
  status, occurrences, contract reference, and decider, needs-decision is `yes`, and the malformed file
  is listed as `unreadable` with `escalation-unreadable` in `not_performed`.
- **Given** only the `operational` (resolved) and `revoke` records, **Then** needs-decision is `no`.

### AC-CLOSURE-009 — First verdict

- **Given** a §E.3 block with `run_status: complete`, `ac_pass_count: 2`, `ac_fail_count: 0`,
  `run_commit_sha: abc1234` and a measured AC count of 3, **Then** First Verdict shows the four values
  verbatim and a count mismatch flag; **Given** a block without `run_commit_sha`, **Then** that field
  shows `not recorded`.

### AC-CLOSURE-010 — Kickoff display (A1 v0.5.1 receipt)

- **Given** five receipt fixtures — (a) requested and effective `llm`, `fallback.applied: false`,
  `llm_answer` approve 0.82, `outcome: approve`; (b) requested and effective `llm+jev` with both
  answers, `outcome: human`; (c) requested `llm+jev`, effective `llm`, `fallback: {applied: true,
  reason: jev_low_confidence}`, no `jev_answer`, `outcome: approve`; (d) receipt (a) plus an unknown
  top-level field `extra_answer`; (e) a contract signed with method `receipt` and no receipt file.
  Fixture (a) is the receipt the contract was signed with; (b), (c), and (d) are written over it after
  signing (A1 v0.5.1 refuses to sign (b) and (d)), so their A1 verify status is `signed-invalid` —
  **When** the report is built for each, **Then** Kickoff shows for (a)-(c) both deciders, each present
  answer with answer and confidence, the fallback flag and reason, the outcome, and the A1 verify
  status (`signed-valid` for (a), `signed-invalid` for (b)-(d)); (d) renders the fields of (a) plus a
  `receipt-field-unrecognized` entry naming `extra_answer`; (e) states `missing`.
- **Given** a contract signed on the human path, **Then** Kickoff shows method `interactive-tty`,
  signer kind `human`, and no receipt.

### AC-CLOSURE-011 — Plan-audit binding (both report streams)

- **Given** contract `plan_audit.verdict: PASS`, no receipt, and in the card evidence home a
  `.moai/reports/c1/plan-audit-iter2.md` ending with `AUDIT-VERDICT: PASS spec=SPEC-FIXTURE-001 receipts=none`
  plus a `.moai/reports/plan-audit/SPEC-FIXTURE-001-review-1.md` ending with FAIL, **Then** the binding
  is `bound` and names `plan-audit-iter2.md`.
- **Given** only `.moai/reports/plan-audit/SPEC-FIXTURE-001-review-2.md` ending with PASS, **Then**
  `bound` naming that file; ending with FAIL, **Then** `mismatch`.
- **Given** no plan-audit report, **Then** `self-reported`; **Given** a receipt whose
  `inputs.plan_audit_report` names a file whose hash differs from the receipt, **Then** `mismatch`.

### AC-CLOSURE-012 — `audit_multi` second-review record

- **Given** backends stubbed to claude `pass`, codex `fail`, glm `inconclusive`, **When** `audit_multi`
  is called with `card_id: c1`, `target: baseBranch`, and `project_root` set to the c1 worktree,
  **Then** `second-review.jsonl` gains exactly one line whose `card`, `contract_card`, `spec_id`,
  `contract_sha256`, `head_sha`, `target`, `scope` (non-empty base, head equal to `head_sha`,
  `changed_files` ≥ 1), `backends`, `participant_count`, `disagreement_flag`, and `recorded_at` match
  the fixture and the returned result.
- **Given** `card_id: c9` unknown to the queue, **Then** a line is written with `spec_id: ""`; **Given**
  the evidence directory made unwritable, **Then** the result carries a non-empty
  `second_review_record_error` and its other fields equal those of a run without `card_id`.
- **Given** the same stubs and no `card_id`, **Then** the returned JSON is byte-identical to the
  pre-change handler's output for the same input (golden file) and no file exists under any
  `.moai/reports/`.

### AC-CLOSURE-013 — Second-review selection and causes

- **Given** evaluation commit `P` (the card HEAD) and contract digest `D`, **When** the state is
  computed for each `second-review.jsonl` fixture, **Then**: no file → `not performed / no-record`;
  only a line whose `card` ≠ `contract_card` → `unbound`; only a line with digest ≠ `D` →
  `contract-changed`; only a line with `target: uncommittedChanges` or `changed_files: 0` →
  `scope-not-covered`; only a line whose counted backends are claude alone → `no-second-model`; only a
  line whose audited commit is not an ancestor of `P` → `not-in-history`; a valid line at `P` with codex
  `pass` and glm `inconclusive` → `performed / pass`; with codex `pass` and glm `fail` →
  `performed / fail`.
- **Given** a valid line at commit `R` and a later non-merge commit changing `src/a.go`, **Then**
  `stale`; **Given** instead a later commit changing only `.moai/specs/SPEC-FIXTURE-001/progress.md`
  and `spec.md`, **Then** `performed` (the sync commit and SHA backfill do not stale the review).
- **Given** an older valid line at `P` and a newer line with `scope-not-covered`, **Then** the older
  line is selected and the state is `performed`.

### AC-CLOSURE-014 — Second verdict rendering

- **Given** states `not performed / no-record`, `stale`, and `performed / fail`, **When** rendered,
  **Then** the Markdown Second Verdict section contains `NOT PERFORMED` followed by `no-record`,
  `STALE` followed by the superseding commit, and `FAILED` respectively.
- **Given** a performed review by `glm` while the contract names `codex`, **Then** the section shows the
  backend, its verdict, the contract's `second_model`, and a substitute marker.

### AC-CLOSURE-015 — Push stop evaluates the pushed ref

- **Given** `mode: contract`, `second_review: required`, the c1 card branch merged `--no-ff` into local
  `develop` with a closure report but no second-review record, **When** the PreToolUse hook receives
  each of `git push origin develop` issued from a tree whose checked-out branch is `main`,
  `git push origin HEAD:develop` from the develop worktree, `git push origin +develop`,
  `git push origin develop:refs/heads/develop`, and `git -C <develop worktree> push origin develop`,
  **Then** every decision is deny with a reason beginning `CLOSURE_PUSH_STOP:` containing
  `SPEC-FIXTURE-001=` and `second_review_not_performed`.
- **Given** the same fixture with a performed `pass` record at the card HEAD, **Then** the hook does not
  deny on A4's account for `git push origin develop`.
- **Given** the not-ready fixture whose `spec.md` carries `status: completed` in the pushed commit (the
  sync commit landed before the merge), **When** the hook receives `git push origin develop`, **Then**
  the decision is deny with `second_review_not_performed`.
- **Given** the not-ready fixture and `git push origin WT-feature`, **Then** no readiness evaluation runs
  (evaluation counter seam stays 0).
- **Given** a second contract `SPEC-FIXTURE-002` (card `c0`, `status: completed`, `ownership.write:
  [src/**]`, `push-develop`, no second-review record and no closure report) already present on
  `origin/develop`, and a range whose only change under `src/` is `src/b.go` from card c1's work, with
  c1 ready (performed `pass` review, current closure report), **When** the hook receives
  `git push origin develop`, **Then** the decision is not a deny on A4's account and the evaluation seam
  records `SPEC-FIXTURE-001` only (an out-of-range contract is not a candidate even when paths overlap).
- **Given** the same `SPEC-FIXTURE-002` fixture (`status: completed`, already on `origin/develop`, no
  closure report and no second-review record), and one additional non-merge commit in the pushed range
  from ready card c1 that edits `.moai/specs/SPEC-FIXTURE-002/spec.md` (one supersession line),
  **When** the hook receives `git push origin develop`, **Then** the evaluation seam records
  `SPEC-FIXTURE-002` (it is a candidate on the path-only SPEC-directory branch of REQ-CLOSURE-015) and
  the decision is deny with a reason beginning `CLOSURE_PUSH_STOP:` containing `SPEC-FIXTURE-002=` with
  `closure_report_missing` and `second_review_not_performed` — the named accepted residual of spec.md
  §H (another card's SPEC-directory edit re-admits a closed card).

### AC-CLOSURE-016 — Readiness code set

- **Given** one fixture per code, **When** the evaluator runs, **Then** each of `contract_invalid`,
  `closure_report_missing`, `closure_report_stale`, `human_verdict_reject`,
  `human_verdict_amend_contract`, `second_review_not_performed`, `second_review_stale`,
  `second_review_failed`, and `push_check_undetermined` is produced by its fixture, and the evaluator's
  declared code list equals exactly these nine codes.
- **Given** a performed `fail` review, **Then** `second_review_failed` is reported and
  `second_review_not_performed` is not.
- **Given** a ready fixture whose latest verdict line is `accept` and one with no verdict line,
  **Then** both are ready (no human-verdict code).
- **Given** a fixture raising two codes, **Then** both are reported, sorted and de-duplicated.

### AC-CLOSURE-017 — Undetermined denies

- **Given** `mode: contract` and a ready fixture (performed `pass` review, current closure report) with
  a local `develop` present, **When** the hook receives each of `git push --all origin`,
  `git push --mirror origin`, `sh -c "git push origin develop"`, `git push origin "$BR"`,
  `git push origin $(git branch --show-current)`, and a bare `git push` in a tree whose current branch
  has no upstream, **Then** every decision is deny with `push_check_undetermined`.
- **Given** a git seam returning an error for the range listing, or a missing `origin/develop` ref,
  **Then** deny with `push_check_undetermined`.

### AC-CLOSURE-018 — Advisory and off

- **Given** the not-performed fixture with `second_review: advisory`, **Then** no second-review code is
  reported and the report renders `NOT PERFORMED` plus a warning line; with `off`, **Then** no
  second-review code and Second Verdict states `not required`.

### AC-CLOSURE-019 — `push-check` CLI

- **Given** a ready fixture, **When** `moai contract push-check` runs, **Then** exit 0 and `ready`;
  **Given** the not-ready fixture, **Then** exit 1 printing `SPEC-FIXTURE-001` with its codes; **Given**
  `moai contract push-check origin HEAD:develop` from a tree on `main`, **Then** the same result as the
  hook for that push; **Given** `moai contract push-check --all origin` or a missing `origin/develop`
  ref, **Then** exit 1 printing `push_check_undetermined`; **Given** an unknown flag, **Then** exit 2;
  **Given** `mode: guided`, **Then**
  exit 0 and one line containing `inactive`.

### AC-CLOSURE-020 — Human verdict recording

- **Given** each of: an agent-environment marker set; a non-terminal standard input; a typed
  confirmation that differs; no closure report; **When** `moai contract verdict c1 accept` runs,
  **Then** each exits 1 and `closure-verdict.jsonl` does not exist.
- **Given** a terminal seam, markers cleared, and the typed token `accept c1`, **Then** exit 0 and one
  line with `verdict: accept`, operator name and email from git configuration,
  `method: interactive-tty`, and `report_sha256` equal to the SHA-256 of the current
  `closure-report.json`.

### AC-CLOSURE-021 — Verdict command reserved to humans

- **Given** `mode: guided` and then `mode: contract`, **When** the hook receives
  `moai contract verdict c1 accept`, **Then** both decisions are deny with a reason beginning
  `CLOSURE_VERDICT_HUMAN_ONLY:`.
- **Given** the fixture, **When** the test runs `moai contract report c1`, `moai contract push-check`,
  and `audit_multi` with `card_id: c1`, **Then** no `closure-verdict.jsonl` exists in any tree
  afterwards.

### AC-CLOSURE-022 — Human verdict section

- **Given** a verdict line whose `report_sha256` equals the report being rendered, **Then** Human
  Verdict shows the verdict, operator, time, and `current`; **Given** the report regenerated with a
  different hash, **Then** `stale` and `human-verdict-stale` in `not_performed`; **Given** two lines,
  **Then** the later one is shown.

### AC-CLOSURE-023 — Guided mode is a no-op

- **Given** `mode: guided` (and, separately, the `autonomy` section absent and `mode: bogus`), **When**
  the hook receives a corpus of Bash calls including `git push origin develop`, `git status`, and
  `go test ./...`, **Then** each output is byte-identical to the hook built without A4's checks
  (golden), and the subprocess seam records zero A4 invocations.

### AC-CLOSURE-024 — One evidence home for writers and readers

- **Given** the c1 worktree exists, **When** `moai contract report c1` runs from the primary checkout
  and from the develop worktree, and `audit_multi` runs with `card_id: c1` and `project_root` set to the
  develop worktree, **Then** the report pair and the second-review line land only in
  `<c1 worktree>/.moai/reports/c1/`, the report reads them from there, and `sources` names those paths.
- **Given** the c1 worktree removed, **When** the report runs, **Then** it writes into
  `<primary checkout>/.moai/reports/c1/`.

### AC-CLOSURE-025 — Template neutrality

- **Given** the template copies of the `moai-ref-cross-model-audit` skill and the `sync-auditor` agent,
  **When** the test reads the lines between `<!-- moai:closure-second-review:start -->` and
  `<!-- moai:closure-second-review:end -->`, **Then** each copy has exactly one such block, the block
  instructs passing the card argument with target `baseBranch` after the last commit changing the
  governed paths, and it contains no match for `SPEC-[A-Z]`, `\bt[0-9]{3,5}\b`,
  `20[0-9]{2}-[0-9]{2}-[0-9]{2}`, `A-Q[0-9]`, or `\b[0-9a-f]{9,40}\b`.
- **Shell check:** `make agents-emit-check` exits 0.

## §D. Edge Cases

- Two queue entries map to the same SPEC: the contract's `card` field decides the card; the other
  entry is ignored.
- A `completed` SPEC's contract is a candidate when a range commit changes its own SPEC directory (the
  card closes in this push; a later card's edit of the closed SPEC's own directory also re-admits it —
  spec.md §H); a contract terminal on the remote whose SPEC directory is unchanged in the range is not
  a candidate even when its governed paths overlap the pushed change (AC-CLOSURE-015).
- A `second-review.jsonl` line with an unknown `schema_version` is skipped and listed under Not
  Performed; it never counts as performed.

## §E. Definition of Done

- All 25 ACs pass under the §A pass check, verbatim output in progress.md §E.2.
- plan.md §E E1-E6 hold.
- No write path to `verdict.md` (review of the diff).
