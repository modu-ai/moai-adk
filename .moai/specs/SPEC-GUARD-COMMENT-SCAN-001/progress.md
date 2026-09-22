# Progress — SPEC-GUARD-COMMENT-SCAN-001 (card t1056)

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **S** — one preprocessing step in one existing file plus one new test file; 8 REQ / 6 AC,
  both inside the 8/8 Tier S ceiling.
- Artifact set: `spec.md` + `plan.md` + `acceptance.md` + this `progress.md`. The separate
  `acceptance.md` departs from the Tier S convention (which inlines AC in `spec.md`) at explicit
  dispatch instruction; the deviation is recorded in `plan.md` §H rather than silently taken.
- Base measured against: HEAD `3dfae918a`, branch `WT-guard-prose`, worktree
  `.claude/worktrees/t1056`.
- Defect class: the third axis of "data is not a command" in the BranchGuard scan-preprocessing
  pipeline. Quoted arguments (`substituteQuotedArguments`) and heredoc bodies
  (`substituteHeredocBodies`) are already collapsed; shell comments are not.
- Narrowing change, so **both** mutation arms are mandatory acceptance criteria — Arm A
  (comment prose no longer matches) and Arm B (the guard is not blunted). Neither alone
  distinguishes a correct narrowing from a disabled guard.
- Discriminant recorded in `spec.md` §A: a refusal without the `BRANCH_GUARD_VIOLATION:` sentinel
  was not produced by this repository's guard. This reversed the card's original premise (E3), and
  E3 is excluded in `spec.md` §E with its measurement.
- Known Gap carried forward, not papered over: the over-match is reproduced at the matcher level
  only; hook-path reachability is unmeasured, and `plan.md` §D decides to keep it a Gap and names
  the file that would close it.
- Evidence file `.moai/reports/t1056/reproduction.md` is gitignored (`.gitignore:227`) and
  untracked — this worktree holds the only copy (`plan.md` §E).
- Plan-phase produced artifacts only: no implementation code, no branch creation, no push, no CI,
  no verification load.

### Plan repair, iteration 1 (plan-audit FAIL 0.625 → repaired)

The first plan-audit returned FAIL at 0.625 against the Tier S threshold 0.75
(`.moai/reports/t1056/verdict.md`, untracked — this worktree holds the only copy). Repaired at HEAD
`5bc42a304`, in repair order D4 → D1 → D2 → D3 → D5, because D4's decision determines D1's wording:

- **D4 (major)** — the manufactured-`#` residual is **ACCEPTED**, with its direction corrected from
  "under-match / fail-open" to **blinding**, and its reach corrected from one token to end-of-line.
  Recorded on the requirement surface (`spec.md` §F) as well as in `plan.md` §A, with both rejected
  repairs and a follow-up coordinate in `substituteQuotedArguments`.
- **D1 (critical)** — AC-GCS-002 gains rows 3-4, placing a comment and a must-survive branch-state
  command on the **same line**. Nothing in the pre-repair set fixed the *direction* of the elision.
  REQ-GCS-004 now also defines "comment run" and states that preceding text survives.
- **D2 (critical)** — AC-GCS-004 gains rows 3-4 with **non-alphanumeric** preceding characters
  (`/`, `.`), row 3 carrying a real command later on the same line. `spec.md` §B.1 carries the bash
  measurement plus a positive control.
- **D3 (major)** — the continuation-line case is recorded as a **blinding** residual in `spec.md`
  §F, and `plan.md` §C's inverted rationale is corrected (handling continuations would *narrow*,
  not widen, the elision). Not re-measured here: every probe form was refused by the Claude Code
  runtime worktree-isolation guard, so the figure is cited to the verdict with that status inline.
- **D5 (major)** — the third §A positive control is re-measured: **112**, not 3. The block's
  conclusion survives; only the attribution was broken.
- **D6 / D7 / D8 (optional)** — D6 recorded as an over-match residual (cited, not re-measured, same
  guard refusal); D7 partially taken (REQ-GCS-007 restated as an outcome, REQ-GCS-004 de-identified
  — REQ-GCS-008 left in place to avoid renumbering); D8 taken ("measurably" → "demonstrably").

The requirement ↔ criterion map in `acceptance.md` was **re-derived from the criteria as they now
read**, not edited incrementally: REQ-GCS-002 now maps to AC-GCS-001 + AC-GCS-004 (positive and
negative halves) and REQ-GCS-004 to AC-GCS-005 + AC-GCS-002, and the map carries an explicit
exclusion table naming the two pass-but-wrong implementations the set now rejects.

### Plan repair, iteration 2 (plan-audit PASS-WITH-DEBT 0.8125 — [HARD] debt closed here)

The second plan-audit returned **PASS-WITH-DEBT at 0.8125** against the Tier S threshold 0.75
(`.moai/reports/t1056/verdict-iter2.md`, untracked — this worktree holds the only copy alongside
the first verdict). Four of the five iteration-1 blocking defects were confirmed resolved and D2
partial; **four new defects** were raised, two of them blocking, with a [HARD] condition that N1
and N2 land in `acceptance.md` **before M2 (test authoring) starts** and that the map be
re-verified afterwards. Repaired at HEAD `9d822d826`:

- **N1 (major, blocking)** — the criterion set still admitted a blinding implementation. The
  discriminating property of a row is **not** the character preceding the `#`; it is whether a
  must-survive command sits **after** the `#` on the same line. AC-GCS-004 row 4 carried the `.`
  character but its `git merge` sat before the hash, so eliding to end-of-line still left a
  matching remnant and all ten measured candidates passed it. Row 4 was rewritten into the
  surviving-command shape and rows 5-6 added, giving **one falsifying row per character** §B.1
  measured as literal (`/` `.` `=` `-`). The rule itself is now written into AC-GCS-004 as a
  [HARD] block, with the measured instance of the mistake named, and the preamble points every
  future row-adder at it.
- **N2 (major, blocking)** — the re-derived map's REQ-GCS-002 cell was false twice: it claimed a
  `;`-preceded row that did not exist (AC-GCS-001 row 3 is whitespace-preceded), and it claimed
  AC-GCS-004 rows 1-4 carried the negative half when only row 3 did. Repaired by **making the
  claims true rather than by softening them**: AC-GCS-001 gains row 4 (`ls ;# …`, the separator
  case with no space, which excludes the two whitespace-only candidates) and row 5 (a second `#`
  inside an open run, which excludes the last-hash candidate and pins where the run *starts*). The
  `&` `|` `(` separators remain unfalsified and are now **named as a residual in the map** rather
  than implied to be covered.
- **N3 (minor)** — `spec.md` §A's `112 (of 282)` drew numerator and denominator from different
  populations. Corrected to **112 of 412 (recursive)** with all four commands and their observed
  output recorded.
- **N4 (minor)** — `spec.md` §F's continuation-line entry had dropped the condition its cited
  source carried (the continuation is literal only where **no whitespace precedes the backslash**).
  Restored. The correction remains **cited, not re-measured** — the probe was refused by the
  Claude Code runtime guard in this run as it was in both prior ones, and routing around the guard
  via a script file was available and deliberately not taken.

**Honest record of a claim that did not hold.** The iteration-1 section above states the map was
"re-derived from the criteria as they now read". The audit found two cells of it false. The
re-derivation in *this* iteration was performed row by row over all nineteen rows and is reported
in the return; it carries the same standing as any other claim here — checkable, not privileged.
The map now opens with a [HARD] notice recording that it has been false twice and why no automated
signal catches it.

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-21
plan_repaired_at: 2026-09-21
plan_repair_iteration: 2
plan_audit_debt: closed   # N1, N2 landed in acceptance.md before M2; N3, N4 taken in the same pass
tier: S
req_count: 8
ac_count: 6
ac_row_count: 19          # was 15 at iteration 1
```

## §E.2 Run-phase Evidence

All figures below are **this run** (2026-09-22), worktree `.claude/worktrees/t1056`, branch
`WT-guard-prose`, measured against HEAD `e1b90a3cb` (M1 GREEN; M2 RED commit `6c7de0d0a`).
Commands ran package-scoped only (`./internal/hook/`) — the full-suite ban was honored; CI owns
the full verdict.

### E8 — verbatim RED (pre-GREEN, TDD evidence)

Command: `go test ./internal/hook/ -count=1 -v -run '^TestBranchStatePatterns_(ShellCommentIsNotACommand|CommentCollapseDoesNotBlindTheGuard)$'`
→ **exit 1**. Captured BEFORE the M1 implementation existed; the RED commit `6c7de0d0a` carries
the test file alone. Pre-file control on the same selector had printed `ok … [no tests to run]`,
so the non-empty sweep below is meaningful. Verbatim failing lines:

```text
    branch_guard_comment_test.go:53: matchBranchStateCommand("# align with git merge --ff-only develop") = ("git merge", true), want false — a shell comment is text the shell never executes
    branch_guard_comment_test.go:53: matchBranchStateCommand("moai todo list  # then git switch main") = ("git switch", true), want false — a shell comment is text the shell never executes
    branch_guard_comment_test.go:53: matchBranchStateCommand("ls ; # git reset --hard HEAD") = ("git reset --hard", true), want false — a shell comment is text the shell never executes
    branch_guard_comment_test.go:53: matchBranchStateCommand("ls ;# git reset --hard HEAD") = ("git reset --hard", true), want false — a shell comment is text the shell never executes
--- FAIL: TestBranchStatePatterns_ShellCommentIsNotACommand (0.00s)
    branch_guard_comment_test.go:182: matchBranchStateCommand("git checkout -b # x") matched = true (suffix "git checkout <branch/-b>"), want match false
--- FAIL: TestBranchStatePatterns_CommentCollapseDoesNotBlindTheGuard (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.655s
```

5 subtests failed RED (AC-GCS-001 r1-r4 + AC-GCS-005); 14 passed (the preservation arm is green
pre-fix — the expected RED shape for a narrowing change). AC-GCS-001 r1 returned
`("git merge", true)` exactly as acceptance.md's RED-now note predicted — no discrepancy there.

### E1 — AC matrix (GREEN, 19/19)

Same command → **exit 0**; `=== RUN` count **21** (2 parents + 19 subtests, `grep -c '^=== RUN'`);
`--- PASS` subtests **19**, `--- FAIL` **0**. Every acceptance row landed as written:

```text
--- PASS: TestBranchStatePatterns_ShellCommentIsNotACommand (0.00s)
    --- PASS: …/whole-line_comment_naming_git_merge_(AC-GCS-001_r1) (0.00s)
    --- PASS: …/trailing_comment_after_a_harmless_command_(AC-GCS-001_r2) (0.00s)
    --- PASS: …/comment_after_a_separator_and_a_space_(AC-GCS-001_r3) (0.00s)
    --- PASS: …/comment_opened_by_a_separator_with_no_space_(AC-GCS-001_r4) (0.00s)
    --- PASS: …/second_#_inside_an_already-open_comment_run_(AC-GCS-001_r5) (0.00s)
--- PASS: TestBranchStatePatterns_CommentCollapseDoesNotBlindTheGuard (0.00s)
    --- PASS: …/no_comment_at_all:_git_merge_(AC-GCS-002_r1) (0.00s)
    --- PASS: …/no_comment_at_all:_git_switch_(AC-GCS-002_r2) (0.00s)
    --- PASS: …/trailing_comment,_command_survives_on_the_same_line_(AC-GCS-002_r3) (0.00s)
    --- PASS: …/trailing_comment,_git_switch_survives_on_the_same_line_(AC-GCS-002_r4) (0.00s)
    --- PASS: …/comment_line,_then_real_command_on_the_next_line_(AC-GCS-003) (0.00s)
    --- PASS: …/alphanumeric-preceded_#_in_a_branch_name_(AC-GCS-004_r1,_documentary) (0.00s)
    --- PASS: …/alphanumeric-preceded_#_in_a_merge_topic_(AC-GCS-004_r2,_documentary) (0.00s)
    --- PASS: …/slash-preceded_mid-word_#,_command_after_it_survives_(AC-GCS-004_r3) (0.00s)
    --- PASS: …/dotted_version_operand,_command_after_it_survives_(AC-GCS-004_r4) (0.00s)
    --- PASS: …/equals-preceded_mid-word_#,_command_after_it_survives_(AC-GCS-004_r5) (0.00s)
    --- PASS: …/hyphen-preceded_mid-word_#,_command_after_it_survives_(AC-GCS-004_r6) (0.00s)
    --- PASS: …/comment_after_-b_elides:_no_operand_presented_(AC-GCS-005) (0.00s)
    --- PASS: …/quoted_#_opens_no_comment_(AC-GCS-006) (0.00s)
    --- PASS: …/heredoc-body_#_opens_no_comment_(AC-GCS-006_companion) (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook	0.687s
```

(Subtest lines above are the run's verbatim `--- PASS` lines with the common parent prefix
elided to `…` for width; the captured full output carries them unshortened.)
Per-AC: AC-GCS-001 = 5 subtests · AC-GCS-002 = 4 · AC-GCS-003 = 1 · AC-GCS-004 = 6 ·
AC-GCS-005 = 1 (the only no-match row) · AC-GCS-006 = 2. **All six AC PASS.**

### E2 — cross-platform build

```
$ go build ./...                           → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### E3 — coverage (package-scoped)

```
$ go test -cover ./internal/hook/   (env scrubbed: unset MOAI_KANBAN_BACKEND — CI-like)
→ exit 0; ok github.com/modu-ai/moai-adk/internal/hook 194.256s coverage: 85.7% of statements
```

85.7% ≥ the 85% package target. The same command WITHOUT the scrub measured the identical
85.7% (exit 1 from the two environmental failures below) — coverage is unaffected by the env.

### E4 — subagent boundary grep

```
$ grep -rn 'AskUserQuestion' internal/hook/ --include='*.go' | grep -v _test.go | grep -v '// '
→ 1 hit: internal/hook/pre_tool.go:676: if input.ToolName == "AskUserQuestion" {
```

**Measured, not 0**: that hit is pre-existing baseline — `pre_tool.go` is absent from this run's
diff (E6 file list proves it), and the line is a `ToolName` string comparison in the PostToolUse
capture routing, not an AskUserQuestion call. **This diff contributes 0 hits.** Positive control:
the same grep WITHOUT the exclusions hits 58 lines, so the grep form works.

### E5 — lint status

```
$ gofmt -l internal/hook/        → empty (exit 0)
$ go vet ./internal/hook/...     → exit 0
$ golangci-lint run ./internal/hook/... → "0 issues." (exit 0)
```

No NEW lint findings; baseline-equivalent clean.

### E6 — commit range, diff scope, protected symbol

```
$ git log --format='%h %s' e756204d0..HEAD
e1b90a3cb fix(SPEC-GUARD-COMMENT-SCAN-001): M1 elide shell comments in the scan pipeline
6c7de0d0a fix(SPEC-GUARD-COMMENT-SCAN-001): M2 RED two-armed comment-scan test matrix
$ git diff --stat e756204d0..HEAD
 .moai/specs/SPEC-GUARD-COMMENT-SCAN-001/spec.md |   4 +-
 internal/hook/branch_guard.go                   |  84 ++++++++++-
 internal/hook/branch_guard_comment_test.go      | 189 ++++++++++++++++++++++++
 3 files changed, 271 insertions(+), 6 deletions(-)
$ git diff e756204d0..HEAD | grep -c branchStatePatterns → 1
```

The single `branchStatePatterns` occurrence is a **context line** (leading space:
` for _, p := range branchStatePatterns {`) displayed around the one-line pipeline rewiring —
**0 `+`/`-` lines touch it**. Positive control: the same diff grep for `substituteShellComments`
hits 4 lines, so the form can match change lines.

### E7 — discrepancies (no blockers)

1. **AC-GCS-001 r5 passes RED** (observed above): the prose `prose about reset --hard HEAD`
   carries no `git` token, so even the last-hash candidate (elide from the LAST word-start `#`)
   leaves text `\bgit\s+reset\s+--hard\b` cannot match — the row excludes nothing as written.
   acceptance.md's row-5 rationale ("re-matches `git reset --hard`") does not hold for this
   string. Row implemented as written per the [HARD] rule; surfaced for manager-spec via the lead.
2. **AC-GCS-004 r4 cell/string mismatch**: the "Character preceding the `#`" cell says `.`, but
   in `v=rel-1.0#rc2` the byte before `#` is `0`. A `.`-widened word-start set (candidate J)
   therefore passes the row as written; the per-character exclusivity claim for `.` is not
   delivered by this row. acceptance.md itself marks the exclusion column "derived, not
   measured — a reader who finds a cell wrong should trust the row over the cell"; the row passes
   with the conformant implementation.
3. **E4 baseline**: 1 pre-existing hit vs the delegation's "expect 0" (attribution in E4).
4. **Ambient-env suite note**: this session's environment carries `MOAI_KANBAN_BACKEND=glm`,
   which fails 2 factory-notice tests (`TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide`,
   `TestFactoryLeadNoticeWorkerCountDrivesLineCount`) — `clearKanbanEnv`
   (session_start_kanban_test.go:19) does not cover `EnvMoaiKanbanBackend`, which
   `factoryLaunchEntry` (session_start_factory.go:83) falls back to. Scrubbed runs exit 0
   (plain suite `ok … 208.888s`; cover run `ok … coverage: 85.7%`). Probe: `unset
   MOAI_KANBAN_BACKEND && go test … -run '^TestFactoryLeadNotice(...)$'` → both PASS. Zero
   call-path overlap with this change (`matchBranchStateCommand` appears 0 times in
   session_start_factory.go / session_start_record.go). Pre-existing test-isolation gap; CI (no
   such env var) is unaffected; candidate follow-up card, outside this SPEC's 2-file scope.

### Known Gap restated (spec.md §F — carried, not dropped)

The over-match is reproduced at the **matcher level only**. Whether a comment-borne false
positive reaches a user-visible deny through the full hook path — the opt-in flag
`Workflow.BranchGuard.Enabled`, primary-checkout discrimination, and the exemption axes — was
NOT measured (plan.md §D decision). Coordinate for whoever closes it:
`internal/hook/pre_tool_branch_guard_integration_test.go`.

### Definition of Done (acceptance.md)

- [x] All six criteria verified with recorded commands and verbatim output (`--- PASS:` lines +
      non-zero `=== RUN` count) — E1/E8.
- [x] `go test ./internal/hook/ -count=1` passes whole-package (scrubbed run `ok … 208.888s`;
      the two unscrubbed failures are environmental, E7 item 4) — package-scoped, full suite left
      to CI.
- [x] `gofmt -l internal/hook/` empty; `go vet ./internal/hook/...` exit 0 — E5.
- [x] `git diff --stat` names the in-scope files; `branchStatePatterns` appears in no change
      line — E6.
- [x] The Known Gap is restated above rather than quietly dropped.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: e72fd90d9   # backfilled per the D3 placeholder convention — the M3 evidence commit landed after this file wrote
run_status: complete
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0        # branchStatePatterns: 0 +/- diff lines (1 context-line occurrence only, E6)
l44_pre_commit_fetch: n/a-worktree-local   # HEAD re-read before each commit (staleness rule); card-local branch, no shared-checkout writes
l44_post_push_fetch: n/a-no-push           # [HARD] local-only — lanes never push; the lead batch-pushes develop
new_warnings_or_lints_introduced: 0     # E5: gofmt empty, vet 0, golangci-lint "0 issues."
cross_platform_build:
  darwin_arm64: pass                    # go build ./... exit 0
  windows_amd64: pass                   # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 4                # branch_guard.go, branch_guard_comment_test.go, spec.md (frontmatter), progress.md
m1_to_mN_commit_strategy: M2-first RED commit (6c7de0d0a) then M1 GREEN (e1b90a3cb) then M3 evidence; TDD RED-before-GREEN history
```

## §E.4 Sync-phase Audit-Ready Signal

### README / docs-site assessment (no doc change needed)

Checked whether any shipped user-facing doc describes the guard's command-text
pattern-matching behavior. Surfaces checked: `README.md`, `README.ko.md`, `README.ja.md`,
`README.zh.md`, and `docs-site/content/{en,ja,ko,zh}/advanced/{config-sections,autonomous-loops}.md`.
Result: every mention is either a capability-table row naming the guard's existence
(`README.md:197`, `README.ko.md:197`), the `workflow.yaml — branch_guard` config section
documenting the `enabled` key, scope (primary checkout vs worktree), exemptions, and fail-open
direction (`advanced/config-sections.md`, all 4 locales), or a cross-reference to the BranchGuard
pattern family from the multi-review-gate page (`advanced/autonomous-loops.md`, en/ja/zh — the ko
page carries no such mention). No shipped doc states which patterns match or how the scanned
command text is preprocessed, so the comment-elision change makes nothing stale. `BRANCH_GUARD_VIOLATION`
appears in none of these files. Per minimal-change discipline, **no README or docs-site edit was
made** — inventing documentation the guard never had was not done.

### CHANGELOG B12 self-tests (run before emission)

- **(a) pre-emission grep** — `grep -c 'SPEC-GUARD-COMMENT-SCAN-001' CHANGELOG.md` → `0`
  (exit 1, no matches). No duplicate entry; emission proceeds.
- **(b) AC count match** — `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u` → 6 ids
  (AC-GCS-001..006). The CHANGELOG entry states "6 acceptance criteria (AC-GCS-001..006), all
  PASS per `progress.md` §E.2" — counts agree, and the dispatch ground truth also reads 6/6.
- **(c) file path verification** — every path named in the CHANGELOG entry checked with `ls`:
  `internal/hook/branch_guard.go`, `internal/hook/branch_guard_comment_test.go`,
  `.moai/specs/SPEC-GUARD-COMMENT-SCAN-001/spec.md` — all exist.

### Frontmatter transition note — plan.md / acceptance.md carry no status field

The dispatch named `spec.md`, `plan.md`, and `acceptance.md` for the status transition, but
`plan.md` and `acceptance.md` carry no YAML frontmatter at all, and the schema SSOT
(`.claude/rules/moai/development/spec-frontmatter-schema.md` § Artifact Statelessness) states the
sibling artifacts are stateless on the status axis — they MUST NOT carry a `status:` field; the
SPEC's lifecycle state lives in exactly one place, `spec.md`. House practice agrees: every
neighboring sync close in `CHANGELOG.md` records only the `spec.md` frontmatter transition. The
transition was therefore applied to `spec.md` only; adding a status field to the two stateless
artifacts would have been a modification outside the allowed frontmatter scope, so it was not
done and is recorded here instead.

```yaml
sync_complete_at: 2026-09-22
sync_commit_sha: 11d197b30   # backfilled per the D3 placeholder convention — the sync commit cannot cite its own hash
sync_status: complete
b12_self_test_a: pass (grep count 0)
b12_self_test_b: pass (6 == 6)
b12_self_test_c: pass (3/3 paths verified via ls)
changelog_entry_position: "top of [Unreleased] > Added"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "n/a — stateless on the status axis (spec-frontmatter-schema.md § Artifact Statelessness); carries no status field"
  acceptance_md: "n/a — stateless on the status axis (spec-frontmatter-schema.md § Artifact Statelessness); carries no status field"
canary_compliance_check: not-applicable — this SPEC does not define a forward-looking policy
```

No SPEC body content (`spec.md` / `plan.md` / `acceptance.md`) was modified in this sync commit —
only `spec.md` frontmatter `status:` (`updated` already read the sync date `2026-09-22`), this
`progress.md` §E.4 section, and `CHANGELOG.md`. Verification cited from run-phase §E.2 (code
unchanged during sync; whole-package suite not re-run per dispatch): two-function matrix 19/19
subtests PASS, whole-package scrubbed run `ok … 208.888s`, coverage 85.7%, gofmt/vet/golangci-lint
clean, darwin+windows builds exit 0.

🗿 MoAI

## §F Phase 4 Mode Selection

Decision: serial

- Inputs: tier S · 2 code files · 1 domain (Go, `internal/hook`) · 100% Go · concurrency benefit LOW (coding-heavy) · Agent Teams not requested.
- Evaluation: direct not selected (2-file semantic change, not a single-line fix); fanout not selected (single domain, no independent slices); sweep not selected (under 30 files, semantic work); **serial selected** — coding-heavy single-package work per the Anthropic coding-task caveat; one manager-develop spawn (opus, cycle_type=tdd) carries M1→M2→M3.
- Justification: the run-phase diff is one preprocessing function plus one two-armed test file in one existing package, with the 19-row test matrix fully authored in `acceptance.md`; the blunted-guard hazard is controlled by the SPEC's own two-armed criteria, not by extra spawns.
- Gate record: Implementation Kickoff Approval was presented via `AskUserQuestion` at run entry (2026-09-22); the round timed out with the user away. Proceeding on the operator's own prior instruction — the pasted handoff names `/moai run SPEC-GUARD-COMMENT-SCAN-001` as the primary action — with preconditions 4/4 verified and plan-audit debt closed (`e756204d0`). All actions stay local to this worktree: commits only, no push, no merge.
