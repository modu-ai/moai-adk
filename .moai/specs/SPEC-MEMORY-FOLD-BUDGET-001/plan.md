# plan.md — SPEC-MEMORY-FOLD-BUDGET-001

Memory hygiene pass 1: card-done fold, byte-aware index budget, link-class report. Tier M. Baseline tree: HEAD `2f492df19`. Priority labels and phase ordering only; no time estimates. Revision 0.4.0 (run-entry delta for audit iteration 3 debt D34-D40: the Definition-of-Done path clause, the OD-11 decision row, the dangling-assertion count, the multi-line fold order, the checker/doctor sentence, and the run-mandatory test-containment step M0). Revision 0.3.0 had split the SessionStart half out as well as the link-repair half.

## §A Context

- Card t1502 (Class C, plan → run → sync), operator directive v3.2 redesign step 1. Epic reference: the v3.2 redesign series; this is its first SPEC.
- Read first: `spec.md` §1.2 (measured premises P1-P4, P6-P8, P11, P12), §1.4 (the budget-unit premise), §1.5 (reachability model on the doctor's own index rule, the archive-index definition and invariants (a)-(d) — the primary invariant), §1.6 (store resolution).
- **Counts.** 12 requirements and 13 acceptance criteria (Tier M ceilings are 16 and 16). Not Tier S: two subsystems with a shared checker (CLI plus the taxonomy package, and config). Not Tier L by design need: no unknown calls for `design.md` or `research.md`; the codebase research is §B and §F below.
- **File recount (same method as the previous audits: the non-test source files the milestones in §F name, plus the test files they name; the list was re-extracted from §F with a grep, `progress.md` §E.1).** **Certain count: 15** — 10 non-test source files (`internal/config/defaults.go`, `internal/config/envkeys.go`, `internal/hook/memo/taxonomy/budget.go`, `internal/hook/memo/taxonomy/reach.go`, `internal/hook/memo/taxonomy/linkage.go`, `internal/cli/memory.go`, `internal/cli/memory_fold.go`, `internal/cli/todo.go`, `internal/cli/todo_autodone.go`, `internal/cli/todo_auto.go`) and 5 named test files (`reach_test.go`, `budget_test.go`, `internal/cli/memory_budget_test.go`, `internal/cli/memory_fold_test.go`, `internal/cli/memory_fold_wiring_test.go`). **Decided existing-test edits: none** — `linkage_test.go` carries three dangling count assertions, at `:99` (`TestAuditLinkageFindsDanglingLinks`: one finding over the store-local missing file `feedback_gone.md`), `:377` and `:392` (`TestAuditLinkageSecondaryIndexDanglingIsReported`: three findings each over the store-local names `ghost_a.md`, `ghost_b.md`, `ghost_c.md`), plus a carrier-path check at `:396` (`grep -n WarnDanglingIndexLink internal/hook/memo/taxonomy/linkage_test.go` on tree `b2c254d44` prints lines 99, 377, 392 and 396); every target in them is a bare store-local name, which class-aware classification leaves dangling, so none is expected to change — confirm at the RED step; `memory_test.go` carries no assertion on the doctor JSON keys or dangling text (a grep of both files for `verdict.md`, `index_lines` and the dangling message text finds nothing); the wiring test reuses the same-package helpers `seedCard`, `seedArchivedCard` and `writeSpecFixture` of `todo_autodone_test.go` (lines 85, 103, 138) without editing them; the class-aware unit tests go into `reach_test.go`, so `linkage_test.go` is not touched. **Contingent, named, not planned:** if a changed expectation surfaces in `linkage_test.go` or `memory_test.go`, or the wiring test needs a helper the existing file lacks (`todo_autodone_test.go`), the run phase records the edit as deliberate; each is one more file, so the worst case is **18**.
- **Band verdict, stated plainly.** `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier gives Tier M "5 - 15 files affected" and Tier L "> 15 files". The certain and planned count is **15, which sits on the upper edge of the Tier M band, inside it by the rule's own inclusive range** (the iteration-2 count of 20 was above it). It is not comfortably inside: any one of the three named contingencies takes the card to 16, above the band, and all three to 18. The tier is not changed here and the risk is not hidden; if the run phase hits a contingency it reports the count as a deviation with the file named, and the leader decides between accepting it and a third split (the card-close wiring, §I). LOC is unmeasured at plan phase (the LOC column is guidance, not a gate).
- Run-phase shape: one `manager-develop` (TDD, serial), inside the card worktree, per `quality.yaml` `constitution.development_mode`.

### §A.1 Decisions

Ordered by likelihood of change (user-facing semantics first). Requirement text states behavior; the default each decision picked is recorded here and in `decision-index.md`, so reversing one means editing only the rows named in the last column. Where the leader or the operator ruled the row says so.

| ID | Question | Options | Default picked | Reason | If reversed, change |
|---|---|---|---|---|---|
| OD-1 | Is the card-close wiring default-on or gated? | (a) gated, default off; (b) default on | **(a) gated, compiled default off** — *operator confirmed 2026-10-04 (via the leader's question channel)* | Wiring makes every `todo done` by any session mutate shared state outside the repository, a store the host's own memory subsystem also writes (C-4). Fail-open covers errors, not a wrong-but-successful fold. Flip the default after the operator has run `fold` by hand on the real store and read the result. | REQ-MFB-007, `DefaultMemoryFoldOnDone`, AC-MFB-008 (i) |
| OD-2 | How is the gate expressed, and which values enable it? | (a) environment variable + `defaults.go` constant; (b) a YAML key in a config section | **(a)** `config.EnvMemoryFoldOnDone` (`MOAI_MEMORY_FOLD_ON_DONE`) read at the call site, with `config.DefaultMemoryFoldOnDone = false`; accepted enabling values `1` and `true` (case-insensitive, trimmed), every other value disabled | A YAML key needs a typed struct, a loader, a template mirror and the loader-completeness test (`internal/config/audit_loader_completeness_test.go`); the `MOAI_MEMORY_AUDIT` kill switch (`internal/hook/post_tool.go`, `session_start.go`) is the precedent for env gating in this subsystem. The accepted values reuse the repository's existing environment-flag parse, `envTruthy` at `internal/session/anchor_trace.go:46` (`"1"` and `"true"` case-insensitively, everything else including the empty string falsy) rather than a second vocabulary (audit D29). YAML stays a follow-up. | REQ-MFB-007, M1, M4 |
| OD-3 | What does the 80 % warning key on? | (a) raw bytes and lines; (b) characters; (c) loaded-content characters | **(a)** | Bytes ≥ characters ≥ loaded characters, so (a) warns earliest; the loader's unit is unconfirmed (§1.4), so the safe error is a false alarm. All four measures are still reported. | REQ-MFB-009, AC-MFB-010, AC-MFB-011 |
| OD-4 | Threshold values. | byte cap 25,000 vs 25,600; warn 80 %; line cap 200 | **byte cap 25,000, warn 80 %, line cap 200 (existing)** | 25,000 is the smaller reading of "25KB" (conservative); 80 % is the card's figure; 200 is the existing constant (`defaults.go:500`). Advisory only, so the choice moves when a warning appears, not what is lost. Basis stated in each finding. | `defaults.go` constants only |
| OD-5 | Which archive index does fold file into, and what if none fits or it is too small? | (a) one definition (spec §1.5): among the `project_card_archive_<YYYY>_<MM>.md` files that `MEMORY.md` links and that exist, the greatest name; never auto-create; refuse when none qualifies or when it would carry fewer than the doctor's threshold of resolved links after the fold; (b) the greatest-named archive-pattern file in the store, then test whether `MEMORY.md` links it; (c) month of the fold time, create when missing; (d) accept an under-threshold index and let the doctor flag it afterwards | **(a)** — the single criterion REQ-MFB-003, REQ-MFB-006, AC-MFB-004, AC-MFB-006 and decision-index Q4 all state by reference to §1.5 (audit D22) | (a) is the reading REQ-MFB-003 and Q4 already carried; (b) was the reading of the old REQ-MFB-006 and AC-MFB-006, and for a month rollover (an older archive linked, a newer one created but not yet linked) it refuses where (a) files into the linked, reachable archive — a fold into an unlinked file would make the folded lines unreachable (P7), so (a) is also the reachability-safe reading. P7: an index under the threshold is invisible to the orphan audit, so (c) and (d) both make the doctor report folded files as orphans. Refusing costs a manual step (link the new archive from `MEMORY.md` or add links to a short one); creating or accepting would let an unattended path cause a doctor-visible regression. Month rollover stays manual (Out of Scope); the unlinked newer archive is listed as information so the rollover is not silent. | REQ-MFB-003, REQ-MFB-006, spec §1.5, AC-MFB-004, AC-MFB-006 |
| OD-6 | Dry-run default? | (a) preview by default, `--yes` applies; (b) apply by default | **(a)** | The leader asked for consistency with `drain`/`archive`; measured: `drain` previews by default and takes `--yes`, `archive` applies immediately (P8). The fold follows `drain`, the safer of the two precedents. | REQ-MFB-001, AC-MFB-001 |
| OD-7 | How does fold decide a line belongs to a card? | (a) conjunctive: title-leading card id AND a card-identifying link target; (b) disjunctive; (c) text mention | **(a)** | Discipline lines carry a trailing `(tNNNN …)` provenance citation and must not leave `MEMORY.md` at that card's close (Admission keeps general discipline always-loaded). Requiring both signals keeps those as MENTION. The card says "by link target AND by text naming the card id" — read as a conjunction. AMBIGUOUS lines are reported, never moved. | REQ-MFB-002, AC-MFB-002 |
| OD-11 | Value and ceiling of the card-close execution bound. | (a) 2 s value, 5 s ceiling; (b) a value in minutes; (c) no bound | **(a)** value `DefaultMemoryFoldOnDoneBound = 2 * time.Second`, ceiling 5 s asserted in the test | The close paths are one-shot CLI processes. The repository's own calibration for advisory side work that must not stall a one-shot process is 2 seconds, three times over: `DefaultHookAsyncJoinTimeout` (`internal/config/defaults.go:288`, "how long a one-shot hook process may linger at exit"), `DefaultTraceFlushTimeout` (`:262`) and `DefaultSessionStartDriftTimeout` (`:547`, an advisory that "must never block"); this value reuses it rather than inventing a second calibration (as the first of those comments says of itself). The work bounded here — a file read, one append, one rename — is orders of magnitude below 2 s. The 5 s ceiling is this SPEC's choice, not a measured repository value: 2.5 times the value, and the same size as the short probe and grace bounds in the file (`DefaultManagedCodexProbeTimeout` `:137` and `DefaultManagedCodexTUIStopGrace` `:143`, both 5 s; the file also carries far longer bounds for unrelated long-running jobs, which are not the comparison class), a point past which an interactive `todo done` would read as hung. The ceiling is asserted by name (AC-MFB-008 (x)) and the production value is exercised, not overridden, in AC-MFB-008 (viii) and (ix), so a constant of minutes cannot pass. (A second constant for the SessionStart join bound left with that half.) | `defaults.go` constant, REQ-MFB-007, AC-MFB-008 (ix), (x) |

Split out to follow-up cards (not decided here): **OD-8** (is link repair in the card, and what similarity counts as unambiguous) and **OD-10** (what a repo-relative link becomes) — the link-repair card; **OD-9** (SessionStart surface and store derivation) — the SessionStart card. Their `decision-index.md` rows (Q6, Q8, Q7) record the disposition.

## §B Known issues (injected checklist, filtered to relevance)

B6, B7 and B13 of v0.2.0 described the SessionStart half (the pinned home-join allowlist, the hook `Data` map, the hook-package home isolation); they moved with it to the follow-up card (`spec.md` §4). The identifiers are not renumbered, so the audit reports keep resolving.

- **B1 Cross-platform.** New code uses `filepath` only; verify `GOOS=windows GOARCH=amd64 go build ./...`. The FIFO-based cells (AC-MFB-008 (vii), (viii), (ix)) are Unix-only: they carry a build tag or a runtime skip on Windows, recorded as a skip with its reason, never a silent pass.
- **B2 Cross-SPEC conflict scan.** `SPEC-MEMORY-STORE-RECONCILE-001` REQ-MSR-008 (no constant encoding the unconfirmed cut in the token guard) was read: the guard (`internal/config/token_budget_guard.go`) is not edited here. Its "index dieting" exclusion does not apply (no entry is shortened; the fold moves lines verbatim; verification counts unique targets file-wide and compares the store to the plan exactly). Run `grep -r "Retired\|superseded" internal/hook/memo internal/cli/memory*.go` before editing.
- **B3 Subagent boundary.** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/memory*.go` must stay empty.
- **B4 Frontmatter schema.** Canonical 12 fields; `phase` is the release label `"v3.2.0 target"`.
- **B5 spec-lint conventions that gate a SPEC created on or after 2026-09-27.** (i) every `go test -run` pattern in a decision-rule artifact is anchored `^…$` at every alternation branch, and an outcome assertion is written with the trailing space convention (`VacuousTestAssertion`, `internal/spec/lint_vacuous_assertion.go`); (ii) `### Out of Scope — …` H3 headings; (iii) no uppercase conditional keyword pair; (iv) no bare remote-tracking branch token in a git-command context (`lint_movingref.go`).
- **B8 Doctor test churn.** `linkage.go` keys targets by base name (`:99`, `:134`); making classification class-aware changes the dangling messages the existing taxonomy tests may assert. Measured at plan time: the three dangling count assertions (`linkage_test.go:99`, `:377`, `:392`) each count store-local names, which class-aware classification leaves unchanged, and `memory_test.go` has no assertion on the changed output, so no existing-test edit is planned (§A; confirm at the RED step). Read `internal/hook/memo/taxonomy/*_test.go` and `internal/cli/memory_test.go` before editing and record any changed expectation as deliberate in the RED step.
- **B9 `internal/cli` suite is a heavy run** (`.claude/rules/local/gitflow-lane-protocol.md` §8): take a `moai slot` lease before running the whole package and otherwise scope to the touched tests.
- **B10 Premise correction recorded for the leader.** The brief reads `drain`/`archive` as one dry-run convention; measured, `archive` has none (P8). The fold follows `drain`.
- **B11 Tool provenance.** A plain `go build` in this worktree stamps the Go VCS revision of another checkout (`c8f245c2c9a5`, not an ancestor of HEAD), so the stamp is not trusted; the lint and RED-now binary was stamped through `-ldflags` (see §H and `progress.md` §E.1).
- **B12 Guard refusals met while authoring.** The worktree-isolation guard refused compound Bash forms (a `printf` with a multi-line payload; a `python3 -c` program reading a shell variable; `awk -f` programs; a `python3 <script> <path>` whose path came from a shell variable). Use plain separate commands, the Write tool for file payloads, and literal paths.
- **B14 Blocking reads.** P11: reading a FIFO named `MEMORY.md` blocks. The fold-on-done step runs its read on a bounded wait (the named constant of OD-11, M1) and abandons with no write; with the gate off it opens nothing (AC-MFB-008 (viii)).

## §C Pre-flight (run before the first code change)

```bash
git branch --show-current
git rev-parse --short HEAD
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
go test ./internal/hook/memo/taxonomy/... -count=1
golangci-lint run --timeout=2m
```

Record the baseline pass counts; later red/green is read against them, not against remembered figures.

## §D Constraints

Binding text is `spec.md` §3 (C-1 … C-7). Restated in one line each for the delegation prompt: the real store is never touched; the card-close wiring fails open, is bounded and never calls `AskUserQuestion`; thresholds, the bound and the env name are constants; the content re-check narrows but does not close the host-write race; Windows build passes; no rule or template file is edited; nothing under `internal/hook/` other than `internal/hook/memo/taxonomy/` is edited (the five taxonomy files of M1 and M2 are the only `internal/hook/` paths this card touches; the SessionStart hook side moved to the follow-up card; verifying command and positive control: `acceptance.md` §7); no time estimates. PRESERVE: `internal/config/token_budget_guard.go`, `resolveMemoryDir` and `projectSlug` (`internal/hook/session_end.go`), `dropMemoryIndexLines` and `archiveMemoryFiles` behavior, the topic-file cap semantics, `moai memory drain`, and the doctrine pair `.claude/rules/moai/workflow/moai-memory.md` with its template mirror.

## §E Self-verification deliverables (for manager-develop, per `manager-develop-prompt-template.md` §E)

E1 AC matrix with command and verbatim output per AC · E2 `go build ./...` and the Windows cross-build · E3 coverage ≥ 85 % for `internal/hook/memo/taxonomy` and the new `internal/cli` files · E4 the subagent-boundary grep (B3) · E5 lint, new versus baseline · E6 branch, HEAD and commit SHAs (no push) · E7 blocker report if any · E8 verbatim RED output captured **before** each GREEN, per milestone. Every item carries command, observed output, and the tree it was measured on.

## §F Milestones (ordered by dependency)

Run-phase order is fixed by dependency; M0 (the test-containment repair) comes first, and the highest-change-likelihood decisions (M3 semantics, M4 default) sit after the shared core they need.

**M0 — Test-containment repair (priority High, run-mandatory, FIRST code step; audit finding D35).** Lands before M1 so that every later `go test ./internal/cli` run in this card is already contained. The gap: the fold helper reaches the real home through `memoryCandidateStores` (`internal/cli/memory.go:123`), which calls `userHomeDir()` — a straight delegate to `paths.Home()` (`internal/cli/homedir.go:18-20`) — and not the `userHomeDirFn` seam (`internal/cli/glm_tools.go:124`) that `TestMain` wraps with its home sandbox (`internal/cli/main_test.go`, "HOME SANDBOX", which itself records that production sites calling `paths.Home()` directly are not covered). The gate is an environment variable the operator will set in their own shell, so every pre-existing test that drives a close path (`ArchiveCard` at `todo.go:1104`, `todo_autodone.go:397`, `todo_auto.go:331`; 18 `internal/cli/*_test.go` files reference them) can run with the gate on and resolve stores under the real home. This plan takes the first of the two options the audit named — option (i), the seam — and not the second (a `TestMain` scrub of `config.EnvMemoryFoldOnDone`, which would add a sixteenth certain file):
- `internal/cli/memory.go` (already counted): the home lookup at `:123` goes through `userHomeDirFn()` instead of `userHomeDir()`. One line, no new file, no edit to `main_test.go`, `homedir.go` or any pre-existing test. Behavior outside tests is unchanged (`userHomeDirFn` is initialized to `userHomeDir`).
- Containment cell: AC-MFB-008 (xi), test `TestMemoryFoldOnDone_ExistingClosePathsContained` in `internal/cli/memory_fold_wiring_test.go` (an already-counted file). It runs in the `TestMain`-sandboxed environment exactly as the pre-existing close-path tests do — no per-test `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR` or `MOAI_HOME` override — with the gate set to `1`, drives the three close paths through the unedited helpers, and asserts on the recorder's resolved candidate list.
- RED-now: ledger E9a with control E9b (acceptance.md §2). The RED to record in `progress.md` §E.2 before the one-line edit is the (xi) cell itself failing, its assertion naming a recorded candidate store directory beneath the real home (the cell is written first, per the TDD order below). M0 edits only already-counted files (`memory.go`, `memory_fold_wiring_test.go`), so the certain file count stays 15 (§A).
- Order inside M0: write the (xi) cell, capture its verbatim RED, make the one-line edit, capture GREEN with the swept count (`go test -list`).

**M1 — Shared core (priority High).** No new verb.
- `internal/config/defaults.go`: `DefaultMemoryIndexByteCap`, `DefaultMemoryIndexWarnPercent`, `DefaultMemoryFoldOnDone`, `DefaultMemoryFoldOnDoneBound` (2 s, OD-11; the ceiling of 5 s is asserted in the test, not carried as a second constant).
- `internal/config/envkeys.go`: `EnvMemoryFoldOnDone`.
- `internal/hook/memo/taxonomy/linkage.go`: expose the existing threshold through one accessor (for example `SecondaryIndexLinkThreshold()`), so no second literal exists; the doctor's own use at `:147` is unchanged.
- `internal/hook/memo/taxonomy/budget.go` (new): measure bytes, characters, loaded-content characters, lines; budget audit with the two new finding codes.
- `internal/hook/memo/taxonomy/reach.go` (new): link extraction and classification (store-local, absolute, repo-relative, full-text targets), store snapshot, I(S) / R(S) / T(S) of §1.5 with the threshold read from the same source as `linkage.go`, the **single implementation of the archive-index selection A(S)** (§1.5: linked, present, greatest name) that both the fold and the checker call, and the invariant checker (a)(b)(c)(d): (a)-(c) over snapshots as before, and **(d) by exact comparison — the checker takes the before snapshot, the after snapshot and the plan's `removed` and `appended` line lists, rebuilds the expected `MEMORY.md` and archive index from the plan alone (preserving line endings and the trailing-newline state) and compares them, and every other file, byte for byte**. It is used by every apply path as a pre-write assertion on the in-memory result, and by every test. The checker takes snapshots, never paths into the operator's store.
- Tests: `budget_test.go`, `reach_test.go` (AC-MFB-013 checker self-test: correct fold PASS, eight lossy mutants FAIL — five by (a)-(c), three by (d) alone; a boundary test builds files with threshold − 1 and threshold resolved links using the accessor and asserts index membership flips there; the class-aware link-classification unit tests for the doctor live here too, §A).

**M2 — Doctor (priority High).** AC-MFB-009, -010, -011, -012.
- `internal/cli/memory.go`: report fields `index_bytes`, `index_chars`, `index_loaded_chars`, `index_link_targets`, budget fields; flags `--byte-cap`, `--line-cap` (also passed to `AuditIndex`, so one line cap governs both the budget axis and `MEMORY_INDEX_OVERFLOW`), `--warn-percent`; text render.
- `internal/hook/memo/taxonomy/linkage.go`: class-aware dangling, `MEMORY_REPO_RELATIVE_LINK`; the secondary-index qualification rule itself is unchanged.
- Tests: `internal/cli/memory_budget_test.go`.

**M3 — Fold core (priority High).** AC-MFB-001 … -007.
- `internal/cli/memory_fold.go` (new): command, store resolution, STRONG/AMBIGUOUS/MENTION classifier, plan (`removed` and `appended` lists, REQ-MFB-005 dedupe by whole-line equality with the differing-text report), archive-index selection through the M1 function, verbatim filing, atomic apply with the content re-check (SHA-256 comparison of the two files; temp file in the same directory, then rename) and the pre-write checker call over (a)-(d), refusal on an empty A(S) or an under-threshold one, listing of skipped unlinked archive-pattern files, idempotence.
- `internal/cli/memory.go`: register the verb in `newMemoryCmd`.
- Tests: `internal/cli/memory_fold_test.go` (table-driven over fixture copies and generated variants; a test-only seam to inject a failure between apply steps and a seam that makes the in-memory result diverge from the plan for the (d) abort cell).

**M4 — Card-close wiring (priority High).** AC-MFB-008. Single milestone: AC-MFB-008 (i) is a differential assertion needing no pre-existing artifact, so no golden recording is scheduled ahead of the wiring commit.
- `internal/cli/memory_fold.go`: `foldClosedCardMemory(cardID string)` — gate read (OD-2 accepted values), store resolution only after the gate is open, fold apply on a bounded wait (`DefaultMemoryFoldOnDoneBound`; the bound is a test-overridable variable), a recover wrapper, one stderr line, all errors swallowed to that line; the cancelled step begins no further write; a path-recording seam on its file opens for the (viii) cell, which also records every candidate store directory the helper resolved (for the (xi) cell).
- `internal/cli/todo.go` (near the `recordFactoryCardState` call), `internal/cli/todo_autodone.go` (after `applyAutoDoneCloses`, per closed card), `internal/cli/todo_auto.go` (after the `Mutate` returns nil): one call each.
- Tests: `internal/cli/memory_fold_wiring_test.go` builds its own isolated environment (`t.Setenv` for `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR`, `MOAI_HOME`); it reuses the existing todo helpers without editing them (§A). Cells: disabled differential and accepted-values table, enabled fold, fail-open with an absent archive index, seeded panic, blocked read on a FIFO named `MEMORY.md` (bound overridden), ordering after the queue write, three close paths, **gate-off FIFO never opened (production bound, under 1 s, empty recorder)**, **production bound effective (not overridden, under 3 s)**, **bound constant equals 2 s and is at or below 5 s**.

**M5 — moved.** The SessionStart advisory (v0.2.0 M5) belongs to the follow-up card; no milestone here edits anything under `internal/hook/` except `internal/hook/memo/taxonomy/` (M1 and M2 create and edit its five files).

TDD order inside every milestone: write the AC's test, run it and capture the verbatim RED (the intended assertion failing, not a compile error — `tdd-result-contract.md`), implement to GREEN, refactor. A test selector that sweeps nothing exits 0; every GREEN cell therefore also records the swept count (`go test -list`).

## §G Risks

| Risk | Mitigation | Residual |
|---|---|---|
| Lost update: the host's native subsystem writes `MEMORY.md` between the fold's read and rename. | Content re-check immediately before each rename; abort without writing on mismatch (REQ-MFB-004). | A write landing between the check and the rename still wins. Not closeable without a lock the host does not honor. |
| Doctor semantics change breaks existing assertions or machine consumers of the JSON. | New keys are additive; dangling message text changes only for repo-relative targets; B8 read-first. | A consumer parsing the old dangling text for repo-relative targets. |
| Wrong store in a linked worktree (the repository's doctrine and code disagree on the loaded key). | Always print the store; `--dir` override; wiring acts on one store only. | Which store the host loads in a worktree stays unmeasured. |
| The checker and the doctor disagree on a repo-relative target whose base name equals a store file (P6: the doctor keys by base name, the checker by full text). | The checker is the stricter reading; the fixture has no such target and the case is a named edge. | Fold could be refused or passed differently from the doctor's view in that one shape; not measured. |
| The doctor's threshold constant changes. | The checker reads it through the accessor (M1), so both follow; the boundary test is written against the accessor. | None known. |
| CJK-heavy index over-warns on bytes. | Intended (OD-3); all four measures shown. | Operator annoyance; tune with `--byte-cap`. |
| A blocked read leaves an abandoned goroutine in the CLI process (REQ-MFB-007). | The process ends with the close command; the test releases the FIFO at teardown and checks for leaks. | In a long-lived host process the goroutine would persist until the file unblocks; the CLI is not one. |
| A command closing N cards (`auto-done`, the `--auto` cycle) can wait up to N times the per-card bound when every store read blocks. | The bound is 2 s and per card; a blocked read is an operator-visible fault of the store. | Not tested beyond one blocked card per path; worst case N × 2 s. |
| The exact-plan checker (d) refuses a legitimate fold whose planned lines differ from what the apply wrote (an implementation bug, not a data risk). | The refusal aborts before any write (AC-MFB-007 (iv)); the operator sees the invariant named. | A false refusal costs a rerun, never data. |
| 580-line archive index append cost. | Append-only temp+rename of one file. | None measured; the real store is not read. |
| Fold failure inside an unattended `--auto` cycle is visible only as one stderr line. | Named in `acceptance.md` check spec. | No durable record. |
| No unasked surface for the budget in this card (the SessionStart warning moved). | The doctor is the on-demand surface; the follow-up card adds the unasked one. | Until that card lands the budget is visible only to whoever runs the doctor (acceptance §4.1). |

## §H Tool provenance (SPEC lint and RED-now measurements)

`moai spec lint` and the RED-now commands in `acceptance.md` were run with a binary **built from this tree** and invoked by path (`./bin/moai`), not the installed `moai`, which is far behind the integration branch: `go build -buildvcs=false -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=2f492df19 -X github.com/modu-ai/moai-adk/pkg/version.Date=2026-10-04" -o bin/moai ./cmd/moai`. Its `version` line prints commit `2f492df19`; the code tree is that of `2f492df19` (the plan commits and this delta touch only the SPEC directory). The build commit is declared by stamp because the Go VCS stamp was unreliable here (B11). The verbatim lint results are recorded in `progress.md` §E.1.

## §I Split lines

**First split — applied in revision 0.2.0 (leader ruling, option A).** The link-repair half left the card: `moai memory relink`, the nearest-name suggestion, the repo-relative rewrite and its decision, the similarity threshold (old REQ-MFB-013's suggestion clause, old REQ-MFB-014, old M6, OD-8, OD-10) — a follow-up card (spec Out of Scope). That follow-up must also carry the card-id-token guard (audit D11).

**Second split — applied in this revision (leader ruling of 2026-10-04, option A).** The SessionStart advisory left the card: v0.2.0 REQ-MFB-011 and REQ-MFB-012, M5, AC-MFB-012, OD-9, and the hook-side home-directory isolation (audit D5, D25) — `internal/hook/main_test.go`, `internal/hook/home_isolation_test.go`, `internal/hook/session_start_memory_budget.go` and its test, the one-line edit in `internal/hook/session_start.go`, the hook read seam and the advisory join-bound constant. The leader issues that follow-up card; it inherits decision-index Q7 (operator-confirmed: `additionalContext` only) and findings D5, D25, D23 (join bound), D31, D32. Effect on the counts: certain files 20 → 15, requirements 14 → 12, criteria 14 → 13.

**Remaining split line, not applied.** The card-close wiring (REQ-MFB-007, M4, AC-MFB-008, OD-1/OD-2/OD-11) removes `todo.go`, `todo_autodone.go`, `todo_auto.go` and `memory_fold_wiring_test.go` (4 files; the constants stay inside already-counted config files): certain count 11, 11 requirements, 12 criteria. It is the cut to take if the run phase reaches one of the named contingencies and the count goes above 15; the fold, doctor budget and link report stay together because they share the reachability checker.

## §J Anti-patterns for this SPEC

- Treating a size decrease as a pass anywhere. Counting links by anchored line or by base name.
- Closing the checker at the inclusion invariants (a)-(c): three measured transformations pass them and lose lines, edit another file or choose another destination (spec P12). The exact-plan equality (d) is part of the checker, never optional.
- Defining the checker's index set by anything but the doctor's own secondary-index rule (a files-MEMORY.md-links definition admitted a lossy mutant).
- Two statements of the archive-index criterion. The definition is spec §1.5; every other place refers to it.
- Creating an archive index from fold. Running fold inside the queue lock. An unbounded read on a close path.
- Literal thresholds, bounds or env name in a check. A bound constant with no value or ceiling asserted by name.
- Reading any memory file on a close path before the gate is checked.
- Touching the operator's real store from any test, command or step.
- An unanchored `go test -run` in an acceptance cell.

## §K Cross-references

`spec.md` · `acceptance.md` · `decision-index.md` · `.claude/rules/moai/development/verification-completeness.md` · `.claude/rules/moai/workflow/tdd-result-contract.md` · `internal/cli/memory.go` · `internal/hook/memo/taxonomy/linkage.go`.

## §L Disposition of the audit's optional findings (iteration 1)

Rows that named the SessionStart half (D13's join-bound slack, D16, D18) moved with it to the follow-up card; the remaining content is as recorded.

| ID | Disposition | Where |
|---|---|---|
| D13 | Fixed for the fold-on-done bound: a named constant (M1) with a stated value and ceiling (OD-11), overridden by the tests through a variable, the slack numeric in AC-MFB-008 (vi), (vii) (200 ms bound, 400 ms assertion) and the production cells (viii), (ix) (1 s and 3 s). The join-bound half moved with the SessionStart advisory. | M1, M4, OD-11, AC-MFB-008 |
| D14 | Fixed: AC-MFB-010 writes `--dir <temp store>` and no cap flag, and states that `--line-cap` also feeds `AuditIndex`, with the lines-at-cap and lines-over-cap outcomes. | AC-MFB-010, REQ-MFB-009, M2 |
| D15 | Fixed in part: requirement text no longer names SHA-256, temp-file-then-rename, `internal/config/defaults.go`, token-set Jaccard or the "Open decision" sentences; those live here. REQ-MFB-006, -007, -009 still bundle several observable behaviors because each is one command's response table. | spec §2 |
| D16 | Moved with the SessionStart half. | spec §4 |
| D17 | Fixed: acceptance §1 lists the fixture's 14 files with sizes and the SHA-256 of the two index files; the tests assert the list. | acceptance §1 |
| D18 | Moved with the SessionStart half; the corrected import fact is recorded in spec §4. | spec §4 |
| D19 | Fixed: renamed `MEMORY_INDEX_BUDGET_AT_CAP`; the basis sentence stays mandatory in AC-MFB-010. | spec REQ-MFB-009, §1.4 |
| D20 | Fixed for the cells that remain: E6a keeps its stated limit (identifier-keyed, the `G-*` tests decide); the E4 cell moved with the `relink` verb, and its true reason (`Unknown flag: --dir` on the parent command) is recorded in acceptance §2. | acceptance §2 |

## §M Disposition of the audit's findings (iteration 2)

| ID | Disposition | Where |
|---|---|---|
| D8 | Moved: second split (leader ruling of 2026-10-04); certain count 20 → 15, 12 requirements, 13 criteria. The count sits on the band's edge with three named contingencies, stated plainly in §A. | §A, §I, spec §4 |
| D21 | Fixed: invariant (d) in spec §1.5, REQ-MFB-004, REQ-MFB-012; the checker compares full snapshots against the plan (M1 `reach.go`); AC-MFB-003, AC-MFB-007 (iv), AC-MFB-013; the sixth mutant executed and killed, plus the seventh and a stray-edit mutant (acceptance §5, P12). | spec §1.5, acceptance §3, §5, M1 |
| D22 | Fixed: one archive-index definition (spec §1.5) shared by REQ-MFB-003 and REQ-MFB-006, AC-MFB-004, AC-MFB-006 with the month-rollover cell, decision-index Q4 (OD-5). | spec §1.5, acceptance §3, OD-5 |
| D23 | Fixed for the fold bound: value 2 s, ceiling 5 s, grounded in the repository's own bounds (OD-11); asserted by name and exercised at its production value (AC-MFB-008 (ix), (x)). The join-bound half moved with the SessionStart advisory. | OD-11, spec REQ-MFB-007, acceptance AC-MFB-008 |
| D24 | Fixed: AC-MFB-008 (viii) — gate unset, blocking FIFO, production bound, under 1 s and an empty path recorder. | acceptance AC-MFB-008 |
| D25 | Moved: OD-9 left with the SessionStart half; the false sentence is deleted from this SPEC and the corrected fact is recorded in spec §4. | spec §4 |
| D26 | Fixed: Q1 and Q7 carry the operator's confirmation of 2026-10-04 relayed through the leader's channel; no verdict cites a gitignored report. | decision-index |
| D27 | Fixed: killed by invariant (d) and the seventh mutant (`m7-hub`, acceptance §5). | spec §1.5, acceptance §5 |
| D28 | Fixed: equal target set with different text is kept and reported (REQ-MFB-005, AC-MFB-005). | spec REQ-MFB-005 |
| D29 | Fixed: accepted gate values stated (REQ-MFB-007, OD-2 cites the repository's existing parse); the bound is per card and the worst case for N cards is named (§G). | spec REQ-MFB-007, §G |
| D30 | Fixed: each `Default:` names its ranking step (decision-index). | decision-index |
| D31 | Moved with the SessionStart half. | spec §4 |
| D32 | Fixed for what remains: the binary-lag signal is scoped to checkouts that carry a comparison commit (acceptance §4.1); the SessionStart half moved. | acceptance §4.1 |
| D33 | Fixed: AC-MFB-003 and AC-MFB-012 also map REQ-MFB-012; the requirement-to-criterion table is in acceptance §3. | acceptance §3 |

## §N Disposition of the audit's findings (iteration 3, accepted as run-entry debt)

| ID | Disposition | Where |
|---|---|---|
| D34 | Fixed: the three statements (`acceptance.md` §7, §D, the M5 line) now name the path-exact set actually touched under `internal/hook/`, with a verifying command and a positive control. | acceptance §7, §D, §F M5 line |
| D35 | Recorded as the run-mandatory first code step M0 with its containment cell AC-MFB-008 (xi) and ledger cells E9a/E9b; the code change (one line in `memory.go`) is the run phase's, not made here. Option (i) of the audit is taken; option (ii) is not (it would add a sixteenth certain file). | §F M0, acceptance AC-MFB-008 (xi), §2 E9a/E9b |
| D36 | Fixed: `decision-index.md` row Q10 for OD-11 (FOUNDER, implementation-level, default stamped by the published rule), header sentence updated. | decision-index Q10 |
| D37 | Fixed in the same pass: Q2, Q4, Q5, Q8, Q9 re-stamped. | decision-index |
| D38 | Fixed: three dangling count assertions (`:99`, `:377`, `:392`) plus the carrier-path check at `:396`, all over store-local names; the count stays 15. | §A recount, B8 |
| D39 | Fixed: AC-MFB-004 gains a multi-line reordering variant; AC-MFB-003 and AC-MFB-004 state that the expected `removed` and `appended` lists are fixed in the test, never read back from the command under test. The optional ninth checker mutant (a reordering transformation) is not added. | acceptance AC-MFB-003, AC-MFB-004 |
| D40 | Fixed: `spec.md` §1.5 narrowed to the shared threshold and names the P6-shape disagreement; AC-MFB-003 limits "the checker and the doctor agree" to stores with no such target. No closing cell for the P6-shape store is added. | spec §1.5, acceptance AC-MFB-003 |
| D41 | Open, no SPEC edit applies: Q3 and Q6 keep empty verdicts, which route the Kickoff to the operator by the published rule. | decision-index Q3, Q6 |
| D42 | Optional, not addressed. | — |
