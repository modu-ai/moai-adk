# plan.md — SPEC-MEMORY-FOLD-BUDGET-001

Memory hygiene pass 1: card-done fold, byte-aware index budget, link repair. Tier M. Baseline tree: HEAD `2f492df19`. Priority labels and phase ordering only; no time estimates.

## §A Context

- Card t1502 (Class C, plan → run → sync), operator directive v3.2 redesign step 1. Epic reference: the v3.2 redesign series; this is its first SPEC.
- Read first: `spec.md` §1.2 (measured premises P1-P9), §1.4 (the budget-unit premise), §1.5 (reachability model — the primary invariant), §1.6 (store resolution).
- Tier justification (Tier M): 15 requirements and 15 acceptance criteria (ceilings are 16 and 16), roughly 12-14 files changed (Tier M band 5-15), new code in the hundreds of lines (band 300-1000), not constitutional. Not Tier S: three subsystems (CLI, hook, config). Not Tier L: no design-level unknown needs `design.md` or `research.md`; the codebase research is §B and §F below. The REQ and AC counts sit one below the ceiling, so §I names the split line if plan-audit asks for scope reduction.
- Run-phase shape: one `manager-develop` (TDD, serial), inside the card worktree, per `quality.yaml` `constitution.development_mode`.

### §A.1 Open decisions

Ordered by likelihood of change (user-facing semantics first). Each has a recommended default; the requirement text in `spec.md` states that default and cites the OD, so reversing one means editing only the rows named in the last column. The decision index carries the questions without preference (`decision-index.md`).

| ID | Question | Options | Recommended default | Reason | If reversed, change |
|---|---|---|---|---|---|
| OD-1 | Is the card-close wiring default-on or gated? | (a) gated, default off; (b) default on | **(a) gated, compiled default off** | Wiring makes every `todo done` by any session mutate shared state outside the repository, a store the host's own memory subsystem also writes (C-4). The card says apply to the real store only after the leader confirms. Fail-open covers errors, not a wrong-but-successful fold. Flip the default after the operator has run `fold` by hand on the real store and read the result. | REQ-MFB-007, `DefaultMemoryFoldOnDone`, AC-MFB-008 (i) |
| OD-2 | How is the gate expressed? | (a) environment variable + `defaults.go` constant; (b) a YAML key in a config section | **(a)** `config.EnvMemoryFoldOnDone` (`MOAI_MEMORY_FOLD_ON_DONE`) read at the call site, with `config.DefaultMemoryFoldOnDone = false` | A YAML key needs a typed struct, a loader, a template mirror and the loader-completeness test (`internal/config/audit_loader_completeness_test.go`); the `MOAI_MEMORY_AUDIT` kill switch (`internal/hook/post_tool.go`, `session_start.go`) is the precedent for env gating in this subsystem. The operator persists it in the settings `env` block. YAML stays a follow-up. | REQ-MFB-007, M1, M4 |
| OD-3 | What does the 80 % warning key on? | (a) raw bytes and lines; (b) characters; (c) loaded-content characters | **(a)** | Bytes ≥ characters ≥ loaded characters, so (a) warns earliest; the loader's unit is unconfirmed (§1.4), so the safe error is a false alarm. All four measures are still reported. | REQ-MFB-009, AC-MFB-010, AC-MFB-011, REQ-MFB-011 |
| OD-4 | Threshold values. | byte cap 25,000 vs 25,600; warn 80 %; line cap 200 | **byte cap 25,000, warn 80 %, line cap 200 (existing)** | 25,000 is the smaller reading of "25KB" (conservative); 80 % is the card's figure; 200 is the existing constant (`defaults.go:500`). Advisory only, so the choice moves when a warning appears, not what is lost. Basis stated in each finding. | `defaults.go` constants only |
| OD-5 | Which archive index does fold file into, and what if none fits? | (a) greatest-named existing `project_card_archive_<YYYY>_<MM>.md` linked from `MEMORY.md`, never auto-create; (b) month of the fold time, create when missing | **(a)** | P7: a freshly created index with fewer than three resolvable links is invisible to the orphan audit, so (b) would make doctor report the folded files as orphans, and a created file also needs a new `MEMORY.md` pointer line. Fail closed and name the file to create. Month rollover is a manual step (Out of Scope). | REQ-MFB-003, REQ-MFB-006, AC-MFB-004, AC-MFB-006 |
| OD-6 | Dry-run default? | (a) preview by default, `--yes` applies; (b) apply by default | **(a)** | The leader asked for consistency with `drain`/`archive`; measured: `drain` previews by default and takes `--yes`, `archive` applies immediately (P8). The fold follows `drain`, the safer of the two precedents. | REQ-MFB-001, AC-MFB-001 |
| OD-7 | How does fold decide a line belongs to a card? | (a) conjunctive: title-leading card id AND a card-identifying link target; (b) disjunctive; (c) text mention | **(a)** | Discipline lines carry a trailing `(tNNNN …)` provenance citation and must not leave `MEMORY.md` at that card's close (Admission keeps general discipline always-loaded). Requiring both signals keeps those as MENTION. The card says "by link target AND by text naming the card id" — read as a conjunction. AMBIGUOUS lines are reported, never moved. | REQ-MFB-002, AC-MFB-002 |
| OD-8 | Is link repair in this card, and what similarity counts as "unambiguous"? | (a) in, token-set Jaccard ≥ 0.75 with exactly one candidate at or above it; (b) split to a follow-up SPEC | **(a)**, constant `DefaultMemoryLinkRepairMinSimilarity = 0.75` | The card lists link repair as item (3). Basis for 0.75: one measured pair, the dangling name from the investigation note against its on-disk sibling, scores 0.778 by file name alone (computed from the two strings; the real store was not read), so 0.75 admits it; requiring a single candidate protects the ambiguous cases. One data point is thin (decision index Q8). §I names (b) as the contingency. | REQ-MFB-013, REQ-MFB-014, M6 |
| OD-9 | SessionStart surface and store derivation. | (a) `additionalContext` only, a new hook-side home-join site with an allowlist row and a real-home guard test; (b) also `systemMessage`; (c) extract one shared resolver | **(a)** | Binary-lag and guard-liveness use `additionalContext` only and the guard-liveness comment records why a second channel splits one concern. (c) would edit `resolveMemoryDir`, which doctrine says not to change (`internal/hook/session_end.go` comment, `.moai/docs/memory-dir-resolution-doctrine.md`). Reach, stated honestly: the orchestrator model sees the line each session; the operator sees it only if relayed. | REQ-MFB-011, AC-MFB-012, M5 |
| OD-10 | What does a repo-relative link become? | (a) its absolute path when the file exists under the project root, otherwise untouched and reported; (b) always absolute; (c) move the line to the archive index | **(a)** | A rewrite to a path that does not exist adds a second dead target; moving a line is the fold's job, not relink's. Never deletes. | REQ-MFB-014, AC-MFB-014 |

## §B Known issues (injected checklist, filtered to relevance)

- **B1 Cross-platform.** New code uses `filepath` only; verify `GOOS=windows GOARCH=amd64 go build ./...`.
- **B2 Cross-SPEC conflict scan.** `SPEC-MEMORY-STORE-RECONCILE-001` REQ-MSR-008 (no constant encoding the unconfirmed cut in the token guard) was read: the guard (`internal/config/token_budget_guard.go`) is not edited here. Its "index dieting" exclusion does not apply (no entry is shortened beyond the fold's description replacement; verification here counts unique targets file-wide). Run `grep -r "Retired\|superseded" internal/hook/memo internal/cli/memory*.go` before editing.
- **B3 Subagent boundary.** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/memory*.go internal/hook/session_start_memory_budget.go` must stay empty.
- **B4 Frontmatter schema.** Canonical 12 fields; `phase` is the release label `"v3.2.0 target"`.
- **B5 spec-lint conventions that gate a SPEC created on or after 2026-09-27.** (i) every `go test -run` pattern in a decision-rule artifact is anchored `^…$` at every alternation branch, and an outcome assertion is written `--- PASS: <name> ` with the trailing space (`VacuousTestAssertion`, `internal/spec/lint_vacuous_assertion.go`); (ii) `### Out of Scope — …` H3 headings; (iii) no uppercase `IF/THEN`; (iv) no bare `origin/<branch>` token in a git-command context (`lint_movingref.go`).
- **B6 `TestHomeJoinSiteCountIsPinned`.** A new file that joins the home directory with a `.claude/projects` slug in `internal/hook` turns this red until `homeJoinSiteAllowlist` gains its row and a guard test that proves the derivation stays under a temporary home (P9, OD-9).
- **B7 Hook `Data` is `json:"-"`.** A SessionStart test must assert on the **serialized** hook output, not on a struct field, or a line placed in `Data` would pass while reaching nobody (P5).
- **B8 Doctor test churn.** `linkage.go` keys targets by base name (`:99`, `:134`); making classification class-aware changes the dangling messages the existing taxonomy tests may assert. Read `internal/hook/memo/taxonomy/*_test.go` and `internal/cli/memory_test.go` before editing and record the changed expectations as deliberate in the RED step.
- **B9 `internal/cli` suite is a heavy run** (`.claude/rules/local/gitflow-lane-protocol.md` §8): take a `moai slot` lease before running the whole package and otherwise scope to the touched tests.
- **B10 Premise correction recorded for the leader.** The brief reads `drain`/`archive` as one dry-run convention; measured, `archive` has none (P8). The fold follows `drain`.
- **B11 Tool provenance.** A plain `go build` in this worktree stamps the Go VCS revision of another checkout (`c8f245c2c9a5`, not an ancestor of HEAD), so the stamp is not trusted; the lint and RED-now binary was stamped through `-ldflags` (see §H and `progress.md` §E.1).
- **B12 Guard refusals met while authoring.** The worktree-isolation guard refused two compound Bash forms (a `printf` with a multi-line payload; a `python3 -c` program reading a shell variable). Use plain separate commands, the Write tool for file payloads, and literal paths.

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

Binding text is `spec.md` §3 (C-1 … C-7). Restated in one line each for the delegation prompt: the real store is never touched; advisory paths fail open and never call `AskUserQuestion`; thresholds and the env name are constants; the SHA re-check narrows but does not close the host-write race; Windows build passes; doctrine edits are mirrored byte-identically; no time estimates. PRESERVE: `internal/config/token_budget_guard.go`, `resolveMemoryDir` and `projectSlug` (`internal/hook/session_end.go`), `dropMemoryIndexLines` and `archiveMemoryFiles` behavior, the topic-file cap semantics, `moai memory drain`.

## §E Self-verification deliverables (for manager-develop, per `manager-develop-prompt-template.md` §E)

E1 AC matrix with command and verbatim output per AC · E2 `go build ./...` and the Windows cross-build · E3 coverage ≥ 85 % for `internal/hook/memo/taxonomy` and the new `internal/cli` files · E4 the subagent-boundary grep (B3) · E5 lint, new versus baseline · E6 branch, HEAD and commit SHAs (no push) · E7 blocker report if any · E8 verbatim RED output captured **before** each GREEN, per milestone. Every item carries command, observed output, and the tree it was measured on.

## §F Milestones (ordered by dependency)

Run-phase order is fixed by dependency; the highest-change-likelihood decisions (M3 semantics, M4 default) sit after the shared core they need.

**M1 — Shared core (priority High).** No new verb.
- `internal/config/defaults.go`: `DefaultMemoryIndexByteCap`, `DefaultMemoryIndexWarnPercent`, `DefaultMemoryLinkRepairMinSimilarity`, `DefaultMemoryFoldOnDone`.
- `internal/config/envkeys.go`: `EnvMemoryFoldOnDone`.
- `internal/hook/memo/taxonomy/budget.go` (new): measure bytes, characters, loaded-content characters, lines; budget audit with the two new finding codes.
- `internal/hook/memo/taxonomy/links.go` (new): link classification, full-text target extraction, nearest-name (token-set Jaccard).
- `internal/hook/memo/taxonomy/reach.go` (new): store snapshot, R/T/I of §1.5, and the invariant checker (a)(b)(c) — used by every apply path as a pre-write assertion on the in-memory result, and by every test.
- Tests: `budget_test.go`, `links_test.go`, `reach_test.go` (AC-MFB-015 checker self-test: correct fold PASS, four lossy mutants FAIL).

**M2 — Doctor (priority High).** AC-MFB-009, -010, -011, -013.
- `internal/cli/memory.go`: report fields `index_bytes`, `index_chars`, `index_loaded_chars`, budget fields; flags `--byte-cap`, `--line-cap`, `--warn-percent`; text render.
- `internal/hook/memo/taxonomy/linkage.go`: class-aware dangling, `MEMORY_REPO_RELATIVE_LINK`, nearest-name suggestion.
- Tests: `internal/cli/memory_budget_test.go`, updates to existing doctor tests (B8).

**M3 — Fold core (priority High).** AC-MFB-001 … -007.
- `internal/cli/memory_fold.go` (new): command, store resolution, STRONG/AMBIGUOUS/MENTION classifier, plan, atomic apply with the SHA re-check, idempotence.
- `internal/cli/memory.go`: register the verb in `newMemoryCmd`.
- Tests: `internal/cli/memory_fold_test.go` (table-driven over fixture copies and generated variants; a test-only seam to inject a failure between apply steps).

**M4 — Card-close wiring (priority High).** AC-MFB-008.
- `internal/cli/memory_fold.go` (or a sibling): `foldClosedCardMemory(cardID string)` — gate read, store resolution, fold apply, one stderr line, all errors swallowed to that line.
- `internal/cli/todo.go` (near the `recordFactoryCardState` call), `internal/cli/todo_autodone.go` (after `applyAutoDoneCloses`, per closed card), `internal/cli/todo_auto.go` (after the `Mutate` returns nil): one call each.
- Tests: extend the todo test harness (`MOAI_HOME` sandbox) with isolated `HOME`/`USERPROFILE`/`CLAUDE_CONFIG_DIR`, golden stdout recordings for the disabled path.

**M5 — SessionStart advisory (priority Medium).** AC-MFB-012.
- `internal/hook/session_start_memory_budget.go` (new): store resolution (profile key then default key, session working directory), bounded read, one-line text; shaped like `binaryLagAdvisory` (join bound, async flag, recover).
- `internal/hook/session_start.go`: one `appendAdditionalContext` call after the guard-liveness advisory.
- `internal/hook/home_isolation_test.go`: allowlist row (B6) plus a real-home guard test.
- Tests: `internal/hook/session_start_memory_budget_test.go`.

**M6 — Relink (priority Medium).** AC-MFB-014.
- `internal/cli/memory_relink.go` (new) with preview/`--yes`/`--dir`/`--json`; register in `newMemoryCmd`.
- Tests: `internal/cli/memory_relink_test.go`.

**M7 — Doctrine pointer (priority Low).**
- `.claude/rules/moai/workflow/moai-memory.md` § MEMORY.md Index Budget and its byte-identical mirror under `internal/template/templates/`: one paragraph naming `moai memory doctor` byte/character/line report, `moai memory fold`, `moai memory relink`, in the re-measuring-command-first form, no machine-specific value. Verify the pair with `cmp`.
- Sync-phase (not run-phase): docs-site CLI reference for the new verbs in all four locales, per `hns-oss-docs-i18n-rules`.

TDD order inside every milestone: write the AC's test, run it and capture the verbatim RED (the intended assertion failing, not a compile error — `tdd-result-contract.md`), implement to GREEN, refactor. A test selector that sweeps nothing exits 0; every GREEN cell therefore also records the swept count (`go test -list`).

## §G Risks

| Risk | Mitigation | Residual |
|---|---|---|
| Lost update: the host's native subsystem writes `MEMORY.md` between the fold's read and rename. | SHA-256 re-check immediately before each rename; abort without writing on mismatch (REQ-MFB-004). | A write landing between the check and the rename still wins. Not closeable without a lock the host does not honor. |
| Doctor semantics change breaks existing assertions or machine consumers of the JSON. | New keys are additive; dangling message text changes only for repo-relative targets; B8 read-first. | A consumer parsing the old dangling text for repo-relative targets. |
| Wrong store in a linked worktree (the repository's doctrine and code disagree on the loaded key). | Always print the store; `--dir` override; wiring acts on one store only. | Which store the host loads in a worktree stays unmeasured. |
| CJK-heavy index over-warns on bytes. | Intended (OD-3); all four measures shown. | Operator annoyance; tune with `--byte-cap`. |
| 580-line archive index append cost. | Append-only temp+rename of one file. | None measured; the real store is not read. |
| SessionStart latency. | One stat, one read, join bound, `recover`, same shape as `binaryLagAdvisory`. | Join-bound value unmeasured on a slow filesystem. |
| Fold failure inside an unattended `--auto` cycle is visible only as one stderr line. | Named in `acceptance.md` check spec. | No durable record. |

## §H Tool provenance (SPEC lint and RED-now measurements)

`moai spec lint` and the RED-now commands in `acceptance.md` were run with a binary **built from this tree** and invoked by path (`./bin/moai`), not the installed `moai`, which is far behind the integration branch: `go build -buildvcs=false -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=2f492df19 -X github.com/modu-ai/moai-adk/pkg/version.Date=2026-10-04" -o bin/moai ./cmd/moai`. Its `version` line prints commit `2f492df19`; the tree HEAD is `2f492df19` (tracked tree clean at build; only this SPEC directory untracked). The build commit is declared by stamp because the Go VCS stamp was unreliable here (B11). The verbatim lint result is recorded in `progress.md` §E.1.

## §I Contingency — split line

If plan-audit reaches its Tier M ceiling over scope, split off REQ-MFB-013 and REQ-MFB-014 (link repair, M6, OD-8, OD-10) into a follow-up SPEC; M1 keeps the link-class helper only as far as REQ-MFB-009 needs. If it still exceeds, split REQ-MFB-011/012 (SessionStart, M5) next. The fold, wiring and doctor budget stay together: they share the reachability checker.

## §J Anti-patterns for this SPEC

- Treating a size decrease as a pass anywhere. Counting links by anchored line or by base name.
- Creating an archive index from fold. Running fold inside the queue lock.
- Literal thresholds or a literal env name in a check. Placing the SessionStart line in the `Data` map.
- Conditioning the SessionStart check on the session source.
- Touching the operator's real store from any test, command or step.
- An unanchored `go test -run` in an acceptance cell.

## §K Cross-references

`spec.md` · `acceptance.md` · `decision-index.md` · `.claude/rules/moai/workflow/moai-memory.md` · `.claude/rules/moai/development/verification-completeness.md` · `.claude/rules/moai/workflow/tdd-result-contract.md` · `internal/cli/memory.go` · `internal/hook/memo/taxonomy/linkage.go`.
