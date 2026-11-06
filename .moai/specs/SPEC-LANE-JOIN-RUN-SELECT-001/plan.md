# SPEC-LANE-JOIN-RUN-SELECT-001 — Implementation Plan

## A. Context

Card t1853 (Tier M, Class C; the A-part of the split of card t1848). Tree `fb749d372e61a6b10c53091362c76b4412827427`, branch `WT-10-17-class-9`. Requirements and root cause are in `spec.md` (§C was re-read at this tree). Line numbers below are coordinates read at that tree and drift with edits; the names are the contract.

Files this SPEC changes:

| File | Role |
|---|---|
| `internal/homestate/factory_run_select.go` (new) | `RunFacts` and `(*FactoryDB).ActiveRunFacts` — the read path (REQ-LJR-001) |
| `internal/factorymsg/factory_run_select.go` (new) | `FormatKST` (REQ-LJR-002) |
| `internal/factorymsg/store.go` | `describeRunOwners` (`:415`) renders the start time through `FormatKST` |
| `internal/cli/factory_run_select.go` (new) | seams, row builder, renderer, text chooser, the selection loop, the Korean catalogue |
| `internal/cli/factory.go` | `enterFactoryLaneRun` (`:597`) calls the selection loop on an interactive ambiguity; the `joinCancelledSentinel` constant beside `noActiveFactorySentinel` (`:521`) |
| `internal/cli/main_test.go` | `TestMain` (`:372`) sets the new interactive seam to false so no test reads a real terminal |
| `internal/cli/init_update_notice.go` (`isInteractiveStdin`, `:69`), `internal/cli/glamour_style.go` (`writerIsTerminal`, `:137`) | read only — reused, not duplicated |
| `docs-site/content/{en,ko,ja,zh}/advanced/factory-mode.md` | line 67 of each describes the ambiguity refusal; sync-phase obligation (§I) |

**Lane-join call sites (all four, non-test).** `enterFactoryLaneRun` is called from `cc.go:240`, `glm.go:299`, `codex_factory.go:103` (the lane branch of `enterCodexFactory`) and `codex_launcher.go:959` (`enterCodexRelaunchJoin`, the relaunch loop's join step). The chooser lives inside the gate, so no call site changes. The inventory is a drift guard: `grep -rn "enterFactoryLaneRun(" internal --include='*.go'` filtered to non-test files must list those four plus the definition; a fifth caller needs review.

Existing tests that must stay green: `internal/cli/factory_ambiguous_join_test.go` (`TestAmbiguousRefusalCarriesResolutionGuidance`, `TestExplicitFactoryRunJoinsNamedRunAmongTwoLive`, `TestAmbiguousLeaderJoinSelectsNamedLeaderRun`; observed PASS at this tree, ledger E4), `internal/cli/factory_join_discovery_test.go`, `internal/cli/codex_factory_retire_test.go`, `internal/homestate/factory_run_retire_test.go`.

## B. Decisions, ordered by likelihood of change (reversibility order)

1. **The selection loop (§C.4) and its N2 resolution** — new flow, cheapest to change before M3. The shape is dispatch-fixed (spec.md §E).
2. **Prompt grammar and the Korean catalogue (§C.3, §E)** — new UX, cheap to change before M2.
3. **Q1 — whether a re-presentation re-runs the reconciler** (default applied: read only; `decision-index.md`). Reversible by one added call before the re-read.
4. **Facts shape (`RunFacts`, §C.1)** — a new type consumed only by the chooser.
5. Mechanical steps last: `FormatKST`, the `describeRunOwners` edit, the `TestMain` seam.

## C. Design

### C.1 The read path (`internal/homestate/factory_run_select.go`)

- `type RunFacts struct { RunOwner; LastActivity string; InProgress int }`. `LastActivity` is the original stored string of the latest parsed time, empty when none.
- `(*FactoryDB).ActiveRunFacts(ctx, opts ReconcileOptions) ([]RunFacts, error)`: calls `classifyRuns(ctx, opts, "active")` (the same call `ReconcileActiveRuns` makes first, `factory_run_retire.go:340`; its body issues SELECTs, identity probes and boot-proof queries and no write — read in full at this tree), then, after that cursor is drained, one `SELECT COUNT(*)` over `cards WHERE run_id=? AND state NOT IN ` + `sqlTerminalStates` per run, and the times query below. The single-connection pool deadlocks if a second statement is issued while a row cursor is open (the hazard `classifyRuns` documents at `:199`), so every query is drained before the next starts. The function opens no transaction and writes nothing.
- `LastActivity` is computed in Go: read `workers.heartbeat_at` (rows whose `run_id` matches), `cards.heartbeat_at` and `cards.updated_at` for the run, `time.Parse(time.RFC3339Nano, ...)` each, keep the maximum `time.Time`, skip empty and unparseable values, and return the **original string** of the winner. Never SQL `MAX` or string order: stored times are `time.Now().UTC().Format(time.RFC3339Nano)`, which trims trailing zeros, so `...10:00:00.1Z` sorts after `...10:00:00.11Z` as text, and `...10:00:00Z` after `...10:00:00.5Z`.
- Naming discipline: this file defines none of the identifiers SPEC-FACTORY-RETIRE-GUARD-001 defines (`InProgressCard`, `rowsQueryer`, `inProgressCards`, `InProgressCards`, `ErrRunHasInProgressCards`, `RunInProgressCardsError`, `retireAfterCardRead`); the count is a private helper under a different name.

### C.2 `FormatKST` (`internal/factorymsg/factory_run_select.go`)

`FormatKST(raw string) string` parses with `time.Parse(time.RFC3339Nano, raw)` (an offset input parses too), converts to a fixed zone `time.FixedZone("KST", 9*3600)` — not `LoadLocation("Asia/Seoul")` and not `time.Local`, so a host without tzdata or with another `TZ` renders the same bytes — and formats `2006-01-02 15:04:05` plus ` KST`; a parse failure returns `raw` unchanged. `describeRunOwners` changes `started %s` to use it; `internal/cli` already imports `factorymsg`.

### C.3 Presentation seams (`internal/cli/factory_run_select.go`)

```go
type runChoiceKind int // choiceJoin, choiceCleanupThenJoin, choiceCancel
type runChoice struct { Kind runChoiceKind; RunID string }

// laneRunRow is one listed run as DISPLAY CELLS; the offered choices ride beside them.
type laneRunRow struct {
    RunID, Liveness, LeaderName, Start, LastActivity string // times are FormatKST text; "-" when empty
    InProgress int
    CanCleanup bool
}
func buildLaneRows(facts []homestate.RunFacts, names map[int]string) []laneRunRow
func renderLaneTable(rows []laneRunRow) string
const laneColSep = " | "

type runChooser interface { ChooseLaneRun(rows []laneRunRow) (runChoice, error) }
var runSelectionInteractive = func() bool { return isInteractiveStdin() && writerIsTerminal(os.Stderr) }
var activeRunChooser runChooser = textRunChooser{in: os.Stdin, out: os.Stderr}
```

- Each rendered line is the seven cells `<n>`, run id, liveness, leader name, start, last activity, count joined by `laneColSep`; a test reads a line with `strings.Split(line, laneColSep)` and asserts each cell by equality, so a dropped column changes the cell count and shifts every later cell.
- `CanCleanup = classification == dead && InProgress == 0 && len(rows) >= 2` — computed in the one pure function `buildLaneRows`. The `len(rows) >= 2` clause is the one-row rule of REQ-LJR-006: a single listed run has no "remaining run" to join after a cleanup, so it never offers one. `RetireRunIfDead` re-enforces owner liveness whatever a chooser returns.
- `textRunChooser` renders the table and reads one line **byte-wise up to the newline**, because the launcher goes on to hand the terminal to the backend and a buffered reader that over-reads would swallow the child's first keystrokes. Grammar: `<n>` joins listed run n; `r<n>` cleanup-then-join (accepted only when row n has `CanCleanup`); `q`, an empty line or end of input cancels. **No default**: an empty line is cancel. An invalid entry prints the catalogue's one-line reason and asks again.
- The leader-name column comes from `leaderNamesByPID(root)` (`factory.go:551`), keyed by the run's lead pid; `-` when none. A dead leader's name may already be pruned from the registry (spec.md §H).
- Tests inject a scripted spy chooser, call the pure functions directly, and drive `textRunChooser` with `strings.Reader`/`bytes.Buffer` (including a reader that yields one byte per call and an end-of-input before any newline). `TestMain` sets `runSelectionInteractive = func() bool { return false }`.

### C.4 The selection loop (`enterFactoryLaneRun`) and the N2 resolution

In `enterFactoryLaneRun`, after the existing `ambiguousJoinGuidance` wrap, when `isAmbiguous && explicit == "" && leadTarget == "" && runSelectionInteractive()`, the gate returns `chooseLaneRun(root, timing)`. A named `--leader` still reaches discovery first through the unchanged `discoveryMayAnswer` expression.

**N2 resolution (the unresolved defect of the predecessor's plan §C.5 step 2).** The predecessor re-resolved after every retire attempt with `enterSelectedFactoryRun(root, "", true, timing)`. That call returns the lone survivor's id to `os.Setenv(config.EnvFactoryRunID, ...)` at `factory.go:503` and returns no id, so "present again, no automatic join" was unsatisfiable, and for one survivor no `AmbiguousRunsError` exists to supply rows. Here:

1. **Every presentation reads the store through the read path of §C.1.** `readLaneRunFacts(root)` opens the factory database, calls `ActiveRunFacts(ctx, factorymsg.ReconcileOptionsFor(root))`, closes it, and pairs the result with `leaderNamesByPID(root)`. It writes no `runs` or `events` row, runs no reconciler, and calls neither `ResolveActiveRun` nor `enterSelectedFactoryRun`. The first presentation uses the same call, not the `AmbiguousRunsError` owners the gate received, so there is one source of facts for every pass.
2. **`enterSelectedFactoryRun` is called at exactly two places in the loop, both with an explicit run id**: (a) the operator's `join` choice; (b) the join half of `cleanup then join`, reached only when the retirement succeeded and the re-read shows exactly one active run. No call precedes an operator choice that names the run, so the environment variable is unset at every presentation entry.
3. **The loop** (state: `afterSuccess`, false initially):
   - Read the facts. 0 active rows -> return the existing `NO_ACTIVE_FACTORY` sentinel error. Exactly 1 active row and `afterSuccess` -> join it (2b). Exactly 1 active row and not `afterSuccess` -> present the one-row table (join/cancel only). 2 or more -> present all rows.
   - `join` -> (2a), return its restore function and error as the gate's result (an explicit-join failure is returned as for `--factory-run`, with no re-presentation).
   - `cleanup then join` -> `retireChosenRun(root, runID)` opens the factory handle and calls `RetireRunIfDead(ctx, runID, factorymsg.ReconcileOptionsFor(root))` — the production options, no injected classifier. Success sets `afterSuccess = true`; any non-nil error prints the refusal (catalogue line plus the error text) and sets `afterSuccess = false`. Loop.
   - `cancel` -> return the `FACTORY_JOIN_CANCELLED` sentinel error (constant beside `noActiveFactorySentinel`). The cancel path itself writes nothing and retires nothing.
4. **Loop bound.** Every pass after the first needs a fresh operator entry (end of input cancels); a successful retirement strictly shrinks the active set; a refused attempt writes nothing. The loop therefore ends at the operator's cancel, at a join, at a successful cleanup that leaves one run, or when no run is active. After a refused retirement a one-run set is presented, not ended (corrects the predecessor's O8).
5. **Why a read path and not `ResolveActiveRun` plus a restore.** The audit's alternative (keep the call, call its returned restore before presenting) would still run `ReconcileActiveRuns` on every pass (a write the operator did not choose) and still need a facts source for the single-survivor case. Q1 records the choice.

`RetireRunIfDead` is called unchanged. After SPEC-FACTORY-RETIRE-GUARD-001 lands it also refuses a run holding non-terminal cards (typed error) and flips only on the classified owner stamp; this loop treats any non-nil error as a refusal and does not import those symbols (§J).

### C.5 Non-interactive path

No path change except `describeRunOwners` rendering the start time in KST. No chooser seam is consulted when `runSelectionInteractive()` is false.

## D. Milestones (priority order; no time estimates)

Each milestone is RED-first: the named tests are written and observed failing for the stated reason before the code, then turned green. Milestones follow dependency order; within each, the interface (type or seam) is written before its mechanical consumers.

### M1 — Facts read path and KST (Priority High; REQ-LJR-001, REQ-LJR-002)

- Files: new `internal/homestate/factory_run_select.go` + `factory_run_select_test.go`; new `internal/factorymsg/factory_run_select.go` + `factory_run_select_test.go`; edit `internal/factorymsg/store.go`.
- Tests: `TestActiveRunFactsCountsNonTerminalCards`, `TestActiveRunFactsLastActivityUsesParsedTimes`, `TestActiveRunFactsIsReadOnly` (AC-LJR-001); `TestFormatKST`, `TestAmbiguousRunsErrorRendersStartInKST` (AC-LJR-002).
- Done when: those pass with the swept-count rule met; `GOOS=windows GOARCH=amd64 go build ./...` exits 0.

### M2 — Chooser seams (Priority High; REQ-LJR-003)

- Files: new `internal/cli/factory_run_select.go` (rows, renderer, parser, catalogue) + `factory_run_select_test.go`.
- Tests: `TestLaneJoinChooserRendersFacts` (AC-LJR-003); `TestLaneJoinChooserOptionEligibility`, `TestRunChooserParsesInMemoryStreams` (AC-LJR-004).
- Done when: those pass; the Korean catalogue source contains no `\u` escape.

### M3 — Gate wiring and the loop (Priority High; REQ-LJR-003..007)

- Files: edit `internal/cli/factory.go`, `internal/cli/main_test.go`; the loop in `internal/cli/factory_run_select.go`.
- Tests: `TestLaneJoinTTYAmbiguityInvokesChooserWithFacts` (AC-LJR-003); `TestLaneJoinChooserJoinAndCancel` (AC-LJR-005); `TestLaneJoinChooserRetireThenJoin` (AC-LJR-006); `TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining` (AC-LJR-007); `TestLaneJoinNonTTYKeepsGuidedRefusalInKST`, `TestLaneJoinInteractiveSkipsChooserForExplicitSelectors` (AC-LJR-008).
- Done when: those pass; `factory_ambiguous_join_test.go`, `factory_join_discovery_test.go` and `codex_factory_retire_test.go` stay green; the call-site inventory (§A) re-run lists the same four callers.

## E. Korean string catalogue (chooser; native written register, UTF-8, one constant block so tests assert on the constants)

| Constant | Text |
|---|---|
| prompt header (two or more rows) | `활성 factory run이 2개 이상입니다. 합류할 run을 고르세요.` |
| prompt header (one row) | `활성 factory run이 1개 남았습니다. 합류할 run을 고르거나 취소하세요.` |
| column labels | `번호`, `run id`, `리더 생존`, `리더 이름`, `시작 (KST)`, `마지막 레인 활동 (KST)`, `진행 중 카드` |
| liveness words | `live (생존)`, `dead (종료)`, `indeterminate (판정 불가)` — the three tokens stay verbatim, with a gloss |
| choice legend (two or more rows) | `번호: 해당 run에 합류 / r번호: 정리한 뒤 남은 run에 합류 / q 또는 빈 줄: 취소` |
| choice legend (one row) | `번호: 해당 run에 합류 / q 또는 빈 줄: 취소` |
| ineligible cleanup | `선택 r%d: 정리할 수 없습니다. 리더가 종료되고 진행 중 카드가 없는 run만 정리할 수 있습니다.` (no particle after the numeral, so the sentence is correct for every number) |
| invalid entry | `입력을 이해하지 못했습니다. 번호, r번호, q 중 하나를 입력하세요.` |
| cleanup refused | `run %s 정리에 실패했습니다 (%s). 저장소에서 다시 읽은 최신 상태로 목록을 보여 드립니다.` |
| cancelled | `합류를 취소했습니다. 이 선택으로 추가된 변경은 없습니다.` (it does not claim nothing was ever recorded: the automatic reconciliation or an earlier chooser cleanup may already have changed run state) |
| empty cell | `-` |

The non-interactive refusal keeps its existing English guidance (REQ-LJR-007 changes only the time rendering).

## F. Self-verification (manager-develop reports, Tier M template §E)

- Scoped runs only: `go test -cover` for `internal/homestate`, `internal/factorymsg` and `internal/cli` limited to the new test names plus the named existing tests. Do not run `go test ./...` or the full `internal/cli` suite locally; a package-wide run takes the single-slot lease first (`moai slot acquire --resource go-test-heavy`). CI runs the full suite.
- `GOOS=windows GOARCH=amd64 go build ./...`.
- Disjointness smoke: `git diff --name-only <base>..HEAD` against SPEC-FACTORY-RETIRE-GUARD-001's file list (§J) prints no shared path.
- Call-site inventory (§A) re-run after M3.
- Test isolation: factory databases live under the home state directory; fixtures follow `discoveryTestRoot` / `clearFactoryTestEnv` / `factorySandbox` (`HOME`, `MOAI_HOME` scrubbed into `t.TempDir()`); M1 confirms no test writes the real `~/.moai` database.
- Template-First: no file under `internal/template/templates/` is touched, so `make build` is not required by this SPEC; if run-phase edits reach a template, run `make build` and commit the regenerated embed.

## G. Risks

- **Prompt in a launcher that owns the terminal.** The chooser runs before the hand-off to the backend and reads one line byte-wise; a codex launcher that supervises with piped stdio fails `isInteractiveStdin` and takes the non-interactive path.
- **The read path does not reconcile (Q1).** A records-free dead run that appears between two presentations is listed as a dead zero-card row with cleanup offered, not auto-retired.
- **Window between render and cleanup.** A card can land in a run after its row was rendered. Before SPEC-FACTORY-RETIRE-GUARD-001 lands, `RetireRunIfDead` does not check cards, so that window is open; after it lands the in-transaction guard closes it. No acceptance criterion of this SPEC depends on that window.
- **Worker rows keep a dead run "live".** Cleanup applies to dead-owner runs whose only live records are worker rows; retiring such a run leaves its worker rows in place. Whether that residue matters was not measured.
- **Leader name for a dead leader may be `-`** (registry pruned).
- **Time-zone display only.** Stored values stay UTC; only rendering changes.
- **A first presentation with one active row** (another process retired a run between the gate's resolution and the read) shows the one-row table instead of joining; nothing is chosen silently.

## H. Anti-patterns to avoid

- Calling `enterSelectedFactoryRun` before the operator has named the run, or re-resolving with an empty selector (the N2 defect).
- A re-read through `ResolveActiveRun` (it runs the reconciler for two or more rows and joins by side effect for one).
- Preselecting a choice, or treating an empty line as a choice (REQ-LJR-003).
- Last-activity by string `MAX` or by string order (REQ-LJR-001).
- A second retirement predicate, or a chooser-side status flip instead of `RetireRunIfDead` (REQ-LJR-005).
- Reading stdin past the newline (the child session's keystrokes are lost).

## I. Sync-phase obligations (no acceptance criterion pins these; see spec.md §H)

- `docs-site/content/{en,ko,ja,zh}/advanced/factory-mode.md` line 67 (the ambiguity-refusal sentence) is updated in all four locales in the same change (the oss-docs 4-locale same-PR obligation): an interactive `-l` shows a chooser on two or more active runs; the refusal text now shows KST times. The sync phase reads all four pages before editing.

## J. Cross-references and the dependency order with SPEC-FACTORY-RETIRE-GUARD-001 (card t1852)

Its plan touches `internal/homestate/factory_run_retire.go`, a new `internal/homestate/factory_run_cards.go`, `internal/cli/factory_handoff_recover.go`, a new `internal/cli/factory_retire_guard_text.go`, `internal/cli/factory_relaunch.go` and their tests (read in the sibling tree `.moai/worktrees/t1852/.moai/specs/SPEC-FACTORY-RETIRE-GUARD-001/plan.md`, §A). This SPEC touches none of them; the two file sets are disjoint. The one shared function is `RetireRunIfDead`, which this SPEC only calls.

- **Compile-time:** none. This SPEC imports no symbol the other defines and defines none of the identifiers listed in §C.1.
- **Behavioral order:** either lands first. If this SPEC lands first, the chooser's eligibility (`CanCleanup`, computed from the freshly read count) is the only card check and the render-to-cleanup window stays open (§G). If SPEC-FACTORY-RETIRE-GUARD-001 lands first, `RetireRunIfDead` closes that window and the chooser prints its typed refusal as ordinary error text.
- **Merge:** the later landing rebases over a disjoint file set; the only textual overlap candidate is `docs-site/.../factory-mode.md`, where its sync phase edits the `--retire` sentence and this SPEC's sync phase edits the ambiguity sentence of the same line 67 — the second sync commit re-reads the line before editing.

Other references: SPEC-FACTORY-STALE-RUN-HEAL-001, SPEC-FACTORY-RUN-RETIRE-001 (the single retirement gate), SPEC-FACTORY-LANE-JOIN-SOCKET-001 (the shared lane-join gate), `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell adoption).
