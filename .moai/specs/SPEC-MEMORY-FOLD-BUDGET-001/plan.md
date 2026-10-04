# plan.md — SPEC-MEMORY-FOLD-BUDGET-001

Memory hygiene pass 1: card-done fold, byte-aware index budget, link-class report. Tier M. Baseline tree: HEAD `2f492df19`. Priority labels and phase ordering only; no time estimates. Revision 0.2.0 (plan delta for audit iteration 1; the link-repair half is split out).

## §A Context

- Card t1502 (Class C, plan → run → sync), operator directive v3.2 redesign step 1. Epic reference: the v3.2 redesign series; this is its first SPEC.
- Read first: `spec.md` §1.2 (measured premises P1-P11), §1.4 (the budget-unit premise), §1.5 (reachability model on the doctor's own index rule — the primary invariant), §1.6 (store resolution).
- **Counts.** 14 requirements and 14 acceptance criteria (Tier M ceilings are 16 and 16). Not Tier S: three subsystems (CLI, hook, config). Not Tier L by design need: no unknown calls for `design.md` or `research.md`; the codebase research is §B and §F below.
- **File recount (same method as the audit: non-test source and doc files the milestones name, plus named test files).** Certain count: **20** files — 12 non-test source files (`internal/config/defaults.go`, `internal/config/envkeys.go`, `internal/hook/memo/taxonomy/budget.go`, `internal/hook/memo/taxonomy/reach.go`, `internal/hook/memo/taxonomy/linkage.go`, `internal/cli/memory.go`, `internal/cli/memory_fold.go`, `internal/cli/todo.go`, `internal/cli/todo_autodone.go`, `internal/cli/todo_auto.go`, `internal/hook/session_start_memory_budget.go`, `internal/hook/session_start.go`) and 8 named test files (`reach_test.go`, `budget_test.go`, `internal/cli/memory_budget_test.go`, `memory_fold_test.go`, `memory_fold_wiring_test.go`, `internal/hook/session_start_memory_budget_test.go`, `internal/hook/home_isolation_test.go`, `internal/hook/main_test.go`). Up to **23** if the existing doctor tests need changed expectations (B8: `internal/hook/memo/taxonomy/linkage_test.go`, `internal/cli/memory_test.go`) or the todo test harness needs an extension (`internal/cli/todo_autodone_test.go`). The audit counted about 24 for v0.1.0; the split and the removal of the doctrine pair and `links.go` bring it to 20.
- **Band verdict (stated plainly, not hidden): the plan is still above the Tier M band.** `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier gives Tier M "5 - 15 files affected" and Tier L "> 15 files". The certain count 20 exceeds the band by 5; the upper bound 23 by 8. Per the leader ruling Tier M is kept and the tier is not changed here; the REQ and AC counts (14, 14) and the single-run-session shape are Tier M. LOC is unmeasured at plan phase (the LOC column is guidance, not a gate). The second split line is in §I; the leader decides whether to take it before run phase.
- Run-phase shape: one `manager-develop` (TDD, serial), inside the card worktree, per `quality.yaml` `constitution.development_mode`.

### §A.1 Decisions

Ordered by likelihood of change (user-facing semantics first). Requirement text states behavior; the default each decision picked is recorded here and in `decision-index.md`, so reversing one means editing only the rows named in the last column. Where the leader ruled (`.moai/reports/t1502/leader-disposition.md`) the row says so.

| ID | Question | Options | Default picked | Reason | If reversed, change |
|---|---|---|---|---|---|
| OD-1 | Is the card-close wiring default-on or gated? | (a) gated, default off; (b) default on | **(a) gated, compiled default off** — *leader ruling* | Wiring makes every `todo done` by any session mutate shared state outside the repository, a store the host's own memory subsystem also writes (C-4). Fail-open covers errors, not a wrong-but-successful fold. Flip the default after the operator has run `fold` by hand on the real store and read the result. | REQ-MFB-007, `DefaultMemoryFoldOnDone`, AC-MFB-008 (i) |
| OD-2 | How is the gate expressed? | (a) environment variable + `defaults.go` constant; (b) a YAML key in a config section | **(a)** `config.EnvMemoryFoldOnDone` (`MOAI_MEMORY_FOLD_ON_DONE`) read at the call site, with `config.DefaultMemoryFoldOnDone = false` | A YAML key needs a typed struct, a loader, a template mirror and the loader-completeness test (`internal/config/audit_loader_completeness_test.go`); the `MOAI_MEMORY_AUDIT` kill switch (`internal/hook/post_tool.go`, `session_start.go`) is the precedent for env gating in this subsystem. The leader ruling names an environment-variable gate for Q1. YAML stays a follow-up. | REQ-MFB-007, M1, M4 |
| OD-3 | What does the 80 % warning key on? | (a) raw bytes and lines; (b) characters; (c) loaded-content characters | **(a)** | Bytes ≥ characters ≥ loaded characters, so (a) warns earliest; the loader's unit is unconfirmed (§1.4), so the safe error is a false alarm. All four measures are still reported. | REQ-MFB-009, AC-MFB-010, AC-MFB-011, REQ-MFB-011 |
| OD-4 | Threshold values. | byte cap 25,000 vs 25,600; warn 80 %; line cap 200 | **byte cap 25,000, warn 80 %, line cap 200 (existing)** | 25,000 is the smaller reading of "25KB" (conservative); 80 % is the card's figure; 200 is the existing constant (`defaults.go:500`). Advisory only, so the choice moves when a warning appears, not what is lost. Basis stated in each finding. | `defaults.go` constants only |
| OD-5 | Which archive index does fold file into, and what if none fits or it is too small? | (a) greatest-named existing `project_card_archive_<YYYY>_<MM>.md` linked from `MEMORY.md`, never auto-create, refuse when the index would carry fewer than the doctor's threshold of resolved links after the fold; (b) month of the fold time, create when missing; (c) accept an under-threshold index and let the doctor flag it afterwards | **(a)** | P7: an index under the threshold is invisible to the orphan audit, so (b) and (c) both make the doctor report folded files as orphans. Refusing costs a manual step (add links to the archive index first); accepting would let an unattended path cause a doctor-visible regression. Month rollover stays manual (Out of Scope). | REQ-MFB-003, REQ-MFB-006, AC-MFB-004, AC-MFB-006 |
| OD-6 | Dry-run default? | (a) preview by default, `--yes` applies; (b) apply by default | **(a)** | The leader asked for consistency with `drain`/`archive`; measured: `drain` previews by default and takes `--yes`, `archive` applies immediately (P8). The fold follows `drain`, the safer of the two precedents. | REQ-MFB-001, AC-MFB-001 |
| OD-7 | How does fold decide a line belongs to a card? | (a) conjunctive: title-leading card id AND a card-identifying link target; (b) disjunctive; (c) text mention | **(a)** | Discipline lines carry a trailing `(tNNNN …)` provenance citation and must not leave `MEMORY.md` at that card's close (Admission keeps general discipline always-loaded). Requiring both signals keeps those as MENTION. The card says "by link target AND by text naming the card id" — read as a conjunction. AMBIGUOUS lines are reported, never moved. | REQ-MFB-002, AC-MFB-002 |
| OD-9 | SessionStart surface and store derivation. | (a) `additionalContext` only, a new hook-side home-join site with an allowlist row and a real-home guard test; (b) also `systemMessage`; (c) extract one shared resolver | **(a)** — *leader ruling (additionalContext only, no per-session screen output)* | Binary-lag and guard-liveness use `additionalContext` only and the guard-liveness comment records why a second channel splits one concern. Reach, stated honestly: the orchestrator model sees the line each session; the operator sees it only if relayed. Why not (c), recorded for audit finding D18: `internal/cli` imports `internal/hook` (for example `internal/cli/binary_lag_test.go`) and the reverse import does not occur (a search of `internal/hook/*.go` for `internal/cli` matches nothing), so a shared resolver would have to live in a new neutral package that both sides migrate onto, re-pointing `memoryCandidateStores` (whose tests the doctor pins) and `resolveMemoryDir` (doctrine says do not change, `internal/hook/session_end.go`, `.moai/docs/memory-dir-resolution-doctrine.md`) — more files than this card's band can carry — and it would still not settle which key the host loads (spec §1.6). The advisory instead derives the same two keys in the same order, prints the path it derived, and a test asserts its derivation equals the independently computed profile-key path. Residual: the two derivations can drift (§G). | REQ-MFB-011, AC-MFB-012, M5 |

Split out to the follow-up card (not decided here): **OD-8** (is link repair in the card, and what similarity counts as unambiguous) and **OD-10** (what a repo-relative link becomes). Their `decision-index.md` rows (Q6, Q8) record that disposition.

## §B Known issues (injected checklist, filtered to relevance)

- **B1 Cross-platform.** New code uses `filepath` only; verify `GOOS=windows GOARCH=amd64 go build ./...`. A FIFO-based test (AC-MFB-008 (vii)) is Unix-only: it carries a build tag or a runtime skip on Windows, recorded as a skip with its reason, never a silent pass.
- **B2 Cross-SPEC conflict scan.** `SPEC-MEMORY-STORE-RECONCILE-001` REQ-MSR-008 (no constant encoding the unconfirmed cut in the token guard) was read: the guard (`internal/config/token_budget_guard.go`) is not edited here. Its "index dieting" exclusion does not apply (no entry is shortened; the fold moves lines verbatim; verification counts unique targets file-wide). Run `grep -r "Retired\|superseded" internal/hook/memo internal/cli/memory*.go` before editing.
- **B3 Subagent boundary.** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/memory*.go internal/hook/session_start_memory_budget.go` must stay empty.
- **B4 Frontmatter schema.** Canonical 12 fields; `phase` is the release label `"v3.2.0 target"`.
- **B5 spec-lint conventions that gate a SPEC created on or after 2026-09-27.** (i) every `go test -run` pattern in a decision-rule artifact is anchored `^…$` at every alternation branch, and an outcome assertion is written `--- PASS: <name> ` with the trailing space (`VacuousTestAssertion`, `internal/spec/lint_vacuous_assertion.go`); (ii) `### Out of Scope — …` H3 headings; (iii) no uppercase conditional keyword pair; (iv) no bare remote-tracking branch token in a git-command context (`lint_movingref.go`).
- **B6 `TestHomeJoinSiteCountIsPinned`.** A new file that joins the home directory with a `.claude/projects` slug in `internal/hook` turns this red until `homeJoinSiteAllowlist` gains its row and a guard test that proves the derivation stays under a temporary home (P9, OD-9).
- **B7 Hook `Data` is `json:"-"`.** A SessionStart test must assert on the **serialized** hook output, not on a struct field, or a line placed in `Data` would pass while reaching nobody (P5).
- **B8 Doctor test churn.** `linkage.go` keys targets by base name (`:99`, `:134`); making classification class-aware changes the dangling messages the existing taxonomy tests may assert. Read `internal/hook/memo/taxonomy/*_test.go` and `internal/cli/memory_test.go` before editing and record the changed expectations as deliberate in the RED step.
- **B9 `internal/cli` suite is a heavy run** (`.claude/rules/local/gitflow-lane-protocol.md` §8): take a `moai slot` lease before running the whole package and otherwise scope to the touched tests.
- **B10 Premise correction recorded for the leader.** The brief reads `drain`/`archive` as one dry-run convention; measured, `archive` has none (P8). The fold follows `drain`.
- **B11 Tool provenance.** A plain `go build` in this worktree stamps the Go VCS revision of another checkout (`c8f245c2c9a5`, not an ancestor of HEAD), so the stamp is not trusted; the lint and RED-now binary was stamped through `-ldflags` (see §H and `progress.md` §E.1).
- **B12 Guard refusals met while authoring.** The worktree-isolation guard refused compound Bash forms (a `printf` with a multi-line payload; a `python3 -c` program reading a shell variable; `awk -f` programs). Use plain separate commands, the Write tool for file payloads, and literal paths.
- **B13 Hook-package home isolation.** Every existing test that drives the SessionStart handler now runs the advisory (P10). M5 therefore edits `internal/hook/main_test.go` (home sandbox) and installs a path-recording read seam with a containment assertion; without both, the real-home safety of those 22-odd test files rests on nothing.
- **B14 Blocking reads.** P11: reading a FIFO named `MEMORY.md` blocks. Both the fold-on-done step and the SessionStart advisory run their read on a bounded wait (named constants, M1) and abandon with no write.

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

Binding text is `spec.md` §3 (C-1 … C-7). Restated in one line each for the delegation prompt: the real store is never touched, including by pre-existing hook tests; advisory paths fail open, are bounded and never call `AskUserQuestion`; thresholds, bounds and the env name are constants; the content re-check narrows but does not close the host-write race; Windows build passes; no rule or template file is edited; no time estimates. PRESERVE: `internal/config/token_budget_guard.go`, `resolveMemoryDir` and `projectSlug` (`internal/hook/session_end.go`), `dropMemoryIndexLines` and `archiveMemoryFiles` behavior, the topic-file cap semantics, `moai memory drain`, and the doctrine pair `.claude/rules/moai/workflow/moai-memory.md` with its template mirror.

## §E Self-verification deliverables (for manager-develop, per `manager-develop-prompt-template.md` §E)

E1 AC matrix with command and verbatim output per AC · E2 `go build ./...` and the Windows cross-build · E3 coverage ≥ 85 % for `internal/hook/memo/taxonomy` and the new `internal/cli` files · E4 the subagent-boundary grep (B3) · E5 lint, new versus baseline · E6 branch, HEAD and commit SHAs (no push) · E7 blocker report if any · E8 verbatim RED output captured **before** each GREEN, per milestone. Every item carries command, observed output, and the tree it was measured on.

## §F Milestones (ordered by dependency)

Run-phase order is fixed by dependency; the highest-change-likelihood decisions (M3 semantics, M4 default) sit after the shared core they need.

**M1 — Shared core (priority High).** No new verb.
- `internal/config/defaults.go`: `DefaultMemoryIndexByteCap`, `DefaultMemoryIndexWarnPercent`, `DefaultMemoryFoldOnDone`, `DefaultMemoryFoldOnDoneBound` (execution bound of REQ-MFB-007), `DefaultMemoryAdvisoryJoinBound` (join bound of REQ-MFB-011).
- `internal/config/envkeys.go`: `EnvMemoryFoldOnDone`.
- `internal/hook/memo/taxonomy/linkage.go`: expose the existing threshold through one accessor (for example `SecondaryIndexLinkThreshold()`), so no second literal exists; the doctor's own use at `:147` is unchanged.
- `internal/hook/memo/taxonomy/budget.go` (new): measure bytes, characters, loaded-content characters, lines; budget audit with the two new finding codes.
- `internal/hook/memo/taxonomy/reach.go` (new): link extraction and classification (store-local, absolute, repo-relative, full-text targets), store snapshot, I(S) / R(S) / T(S) of §1.5 with the threshold read from the same source as `linkage.go`, and the invariant checker (a)(b)(c) — used by every apply path as a pre-write assertion on the in-memory result, and by every test. The checker takes snapshots, never paths into the operator's store.
- Tests: `budget_test.go`, `reach_test.go` (AC-MFB-014 checker self-test: correct fold PASS, five lossy mutants FAIL; a boundary test builds files with threshold − 1 and threshold resolved links using the accessor and asserts index membership flips there).

**M2 — Doctor (priority High).** AC-MFB-009, -010, -011, -013.
- `internal/cli/memory.go`: report fields `index_bytes`, `index_chars`, `index_loaded_chars`, `index_link_targets`, budget fields; flags `--byte-cap`, `--line-cap` (also passed to `AuditIndex`, so one line cap governs both the budget axis and `MEMORY_INDEX_OVERFLOW`), `--warn-percent`; text render.
- `internal/hook/memo/taxonomy/linkage.go`: class-aware dangling, `MEMORY_REPO_RELATIVE_LINK`; the secondary-index qualification rule itself is unchanged.
- Tests: `internal/cli/memory_budget_test.go`, updates to existing doctor tests (B8).

**M3 — Fold core (priority High).** AC-MFB-001 … -007.
- `internal/cli/memory_fold.go` (new): command, store resolution, STRONG/AMBIGUOUS/MENTION classifier, plan, verbatim filing, atomic apply with the content re-check (SHA-256 comparison of the two files; temp file in the same directory, then rename) and the pre-write checker call, refusal on an under-threshold archive index, idempotence.
- `internal/cli/memory.go`: register the verb in `newMemoryCmd`.
- Tests: `internal/cli/memory_fold_test.go` (table-driven over fixture copies and generated variants; a test-only seam to inject a failure between apply steps).

**M4 — Card-close wiring (priority High).** AC-MFB-008. Single milestone: AC-MFB-008 (i) is a differential assertion needing no pre-existing artifact (D10), so no golden recording is scheduled ahead of the wiring commit.
- `internal/cli/memory_fold.go`: `foldClosedCardMemory(cardID string)` — gate read, store resolution, fold apply on a bounded wait (`DefaultMemoryFoldOnDoneBound`; the bound is a test-overridable variable), a recover wrapper, one stderr line, all errors swallowed to that line; the cancelled step begins no further write.
- `internal/cli/todo.go` (near the `recordFactoryCardState` call), `internal/cli/todo_autodone.go` (after `applyAutoDoneCloses`, per closed card), `internal/cli/todo_auto.go` (after the `Mutate` returns nil): one call each.
- Tests: `internal/cli/memory_fold_wiring_test.go` builds its own isolated environment (`t.Setenv` for `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR`, `MOAI_HOME`); it reuses the existing todo helpers without editing them where possible (a needed helper edit adds `todo_autodone_test.go` to the count, §A). Cells: disabled differential, enabled fold, fail-open with an absent archive index, seeded panic, blocked read on a FIFO named `MEMORY.md`, ordering after the queue write, three close paths.

**M5 — SessionStart advisory (priority Medium).** AC-MFB-012.
- `internal/hook/session_start_memory_budget.go` (new): store resolution (profile key then default key, session working directory), bounded read through a package-level read seam, one-line text; shaped like `binaryLagAdvisory` (join bound, async flag, recover).
- `internal/hook/session_start.go`: one `appendAdditionalContext` call after the guard-liveness advisory.
- `internal/hook/main_test.go`: `TestMain` sandboxes `HOME` and `USERPROFILE` to a temporary directory (restored at exit) and installs the read seam as a recorder that returns not-exist for any path outside the sandbox root, so no pre-existing handler test can open a file outside it (P10, B13).
- `internal/hook/home_isolation_test.go`: allowlist row (B6) plus a real-home guard test.
- Tests: `internal/hook/session_start_memory_budget_test.go`.

TDD order inside every milestone: write the AC's test, run it and capture the verbatim RED (the intended assertion failing, not a compile error — `tdd-result-contract.md`), implement to GREEN, refactor. A test selector that sweeps nothing exits 0; every GREEN cell therefore also records the swept count (`go test -list`).

## §G Risks

| Risk | Mitigation | Residual |
|---|---|---|
| Lost update: the host's native subsystem writes `MEMORY.md` between the fold's read and rename. | Content re-check immediately before each rename; abort without writing on mismatch (REQ-MFB-004). | A write landing between the check and the rename still wins. Not closeable without a lock the host does not honor. |
| Doctor semantics change breaks existing assertions or machine consumers of the JSON. | New keys are additive; dangling message text changes only for repo-relative targets; B8 read-first. | A consumer parsing the old dangling text for repo-relative targets. |
| Wrong store in a linked worktree (the repository's doctrine and code disagree on the loaded key). | Always print the store; `--dir` override; wiring acts on one store only. | Which store the host loads in a worktree stays unmeasured. |
| The SessionStart advisory's store derivation and `memoryCandidateStores` drift apart (OD-9, D18). | Same two keys in the same order; the line prints the derived path; an independent-expectation test. | A future edit to one site not mirrored in the other. |
| The checker and the doctor disagree on a repo-relative target whose base name equals a store file (P6: the doctor keys by base name, the checker by full text). | The checker is the stricter reading; the fixture has no such target and the case is a named edge. | Fold could be refused or passed differently from the doctor's view in that one shape; not measured. |
| The doctor's threshold constant changes. | The checker reads it through the accessor (M1), so both follow; the boundary test is written against the accessor. | None known. |
| CJK-heavy index over-warns on bytes. | Intended (OD-3); all four measures shown. | Operator annoyance; tune with `--byte-cap`. |
| A blocked read leaves an abandoned goroutine in the CLI process (REQ-MFB-007). | The process ends with the close command; the test releases the FIFO at teardown and checks for leaks. | In a long-lived host process the goroutine would persist until the file unblocks; the CLI is not one. |
| 580-line archive index append cost. | Append-only temp+rename of one file. | None measured; the real store is not read. |
| SessionStart latency. | One stat, one read, bounded wait, `recover`, same shape as `binaryLagAdvisory`. | Join-bound production value unmeasured on a slow filesystem (named constant, set in M1 by measurement). |
| Fold failure inside an unattended `--auto` cycle is visible only as one stderr line. | Named in `acceptance.md` check spec. | No durable record. |

## §H Tool provenance (SPEC lint and RED-now measurements)

`moai spec lint` and the RED-now commands in `acceptance.md` were run with a binary **built from this tree** and invoked by path (`./bin/moai`), not the installed `moai`, which is far behind the integration branch: `go build -buildvcs=false -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=2f492df19 -X github.com/modu-ai/moai-adk/pkg/version.Date=2026-10-04" -o bin/moai ./cmd/moai`. Its `version` line prints commit `2f492df19`; the code tree is that of `2f492df19` (the plan commit and this delta touch only the SPEC directory). The build commit is declared by stamp because the Go VCS stamp was unreliable here (B11). The verbatim lint results are recorded in `progress.md` §E.1.

## §I Split lines

**First split — applied in this revision (leader ruling, option A).** The link-repair half left the card: `moai memory relink`, the nearest-name suggestion, the repo-relative rewrite and its decision, the similarity threshold (old REQ-MFB-013's suggestion clause, old REQ-MFB-014, old M6, OD-8, OD-10) — now a follow-up card (spec Out of Scope). That follow-up must also carry the card-id-token guard (audit D11).

**Second split line — not applied; flagged for the leader because the certain file count (20) is still above the Tier M band (15).** Split off the SessionStart advisory (REQ-MFB-011/012, M5, AC-MFB-012, OD-9): removes `session_start_memory_budget.go`, `session_start.go`, `session_start_memory_budget_test.go`, `home_isolation_test.go`, `internal/hook/main_test.go` (5 files: certain count 15, inside the band; 12 REQ and 13 AC remain, because AC-MFB-012 is the one criterion covering both requirements) and the whole hook-side isolation work. The alternative cut is the card-close wiring (REQ-MFB-007, M4, AC-MFB-008, OD-1/OD-2): removes `todo.go`, `todo_autodone.go`, `todo_auto.go`, `memory_fold_wiring_test.go` (4 files, plus the gate and bound constants inside already-counted config files; certain count 16, still over; 13 REQ and 13 AC remain). Taking the SessionStart cut alone gets the plan inside the band; the fold, doctor budget and link report stay together because they share the reachability checker.

## §J Anti-patterns for this SPEC

- Treating a size decrease as a pass anywhere. Counting links by anchored line or by base name.
- Defining the checker's index set by anything but the doctor's own secondary-index rule (a files-MEMORY.md-links definition admitted a lossy mutant).
- Creating an archive index from fold. Running fold inside the queue lock. An unbounded read on a close path or at session start.
- Literal thresholds, bounds or env name in a check. Placing the SessionStart line in the `Data` map.
- Conditioning the SessionStart check on the session source.
- Touching the operator's real store from any test, command or step, including a pre-existing hook test that now runs the advisory.
- An unanchored `go test -run` in an acceptance cell.

## §K Cross-references

`spec.md` · `acceptance.md` · `decision-index.md` · `.claude/rules/moai/development/verification-completeness.md` · `.claude/rules/moai/workflow/tdd-result-contract.md` · `internal/cli/memory.go` · `internal/hook/memo/taxonomy/linkage.go`.

## §L Disposition of the audit's optional findings (iteration 1)

| ID | Disposition | Where |
|---|---|---|
| D13 | Fixed: the two bounds are named constants (M1: `DefaultMemoryAdvisoryJoinBound`, `DefaultMemoryFoldOnDoneBound`), the tests override them through variables, and the slack is a number: the join-bound cell sets the bound to 50 ms and asserts return in under 250 ms; the fold-bound cells set it to 200 ms and assert each close path returns in under 400 ms. | M1, M4, M5, AC-MFB-008 (vi, vii), AC-MFB-012 |
| D14 | Fixed: AC-MFB-010 writes `--dir <temp store>` and no cap flag, and states that `--line-cap` also feeds `AuditIndex`, with the lines-at-cap and lines-over-cap outcomes. | AC-MFB-010, REQ-MFB-009, M2 |
| D15 | Fixed in part: requirement text no longer names SHA-256, temp-file-then-rename, `internal/config/defaults.go`, token-set Jaccard or the "Open decision" sentences; those live here. REQ-MFB-006, -007, -009 still bundle several observable behaviors because each is one command's response table. | spec §2 |
| D16 | Fixed: "fires unasked and unconditionally" reworded to "no source, path or branch condition"; AC-MFB-012 asserts the percentage, the keyed measure and the `moai memory doctor` pointer. | acceptance §3 AC-MFB-012, §4.1 |
| D17 | Fixed: acceptance §1 lists the fixture's 14 files with sizes and the SHA-256 of the two index files; the tests assert the list. | acceptance §1 |
| D18 | Fixed by recording the reason a shared resolver was not chosen and the mitigation. | §A.1 OD-9, §G |
| D19 | Fixed: renamed `MEMORY_INDEX_BUDGET_AT_CAP`; the basis sentence stays mandatory in AC-MFB-010. | spec REQ-MFB-009, §1.4 |
| D20 | Fixed for the cells that remain: E5a and E6a keep their stated limit (identifier-keyed, the `G-*` tests decide); the E4 cell moved with the `relink` verb, and its true reason (`Unknown flag: --dir` on the parent command) is recorded in the follow-up pointer. | acceptance §2 |
