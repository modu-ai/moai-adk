# progress.md — SPEC-FACTORY-MANAGED-CARD-CHILD-001

> 단계별 진행 기록. `§E.1` 만 plan 단계(manager-spec)가 채우고, `§E.2`·`§E.3` 은 run 단계(manager-develop), `§E.4` 는 sync 단계(manager-docs)가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-after-plan-audit-iter1 (0.2.0, 재감사 대기)
plan_complete_at: 2026-10-03
artifacts: spec.md (REQ 14) · plan.md · acceptance.md (AC 13) · design.md (D-1..D-6) · decision-index.md · progress.md
tier: M
measured_tree: 2b9e4a4d0
open_clarifications: 0 (Q1-Q3 은 리더가 기본값으로 결정 — decision-index.md; mission contract 11c79e1a)

### 감사 이력과 처분

| 회차 | 판정 | 점수 | audited_sha | 처분 |
|---|---|---|---|---|
| 1 | FAIL | 0.81 | `d26215f25` | MP-7(D1)·MP-6(D2) 차단, 주요 D3–D6, 경미 D7–D12. 개정 0.2.0: D1 결정 기록, D2 `cross-platform exemption` 문장, D3 AC-CC-013 정확히 88, D4 AC-CC-002 `*exec.Cmd` 전체·stdio·펌프 미시작, D5 어댑터 재설계(줄 단위·종료 토큰 인식)와 REQ-CC-010 재서술, D6 어댑터를 레인 전용 이음새로 한정, D7 R-1 관측 RED(정리 접두 한 호출), D8 store.go 인용 정정, D9 REQ-CC-007 분리(007/013/014), D10 REQ-CC-004/005 구현 용어 축소, D11 §F 문구, D12 운영 모드 명시 |

## Residual-risk (plan 단계)

- **무인 레인은 첫 카드 이후로 진행하지 않는다(Q1).** 관리 카드 자식은 우선 턴 뒤 유휴이고 카드 시작 프롬프트를 주입하지 않으므로 운영자 입력이나 브로커 메시지 없이는 카드 작업을 시작하지 않는다. 이 SPEC 의 이득은 운영자가 붙은 관리 레인에서만 닿는다. 문서 known-limits 불릿(AC-CC-012)에 같은 문장을 적는다.
- **연속 시작 실패 상한 없음(Q3).** 관리 시작이 계속 실패하면(예: `app-server` 하위 명령 부재, 핸드셰이크 시간 초과) 루프가 큐의 카드를 연속으로 lease 하며 소진할 수 있다. 카드는 lease 만료까지 묶인다. 현행 직접 문 루프도 상한이 없다.
- **stdin EOF 레인(Q2)**: 세션이 `/exit`·`/quit`·치명 오류 없이는 끝나지 않는다.
- **입력 보장 범위**: 종료 토큰 `/exit`·`/quit` 이후 입력만 보장한다. 종료 토큰 없이 끝난 세션이 이미 받은 줄은 되찾지 않는다(design.md D-6 잔여). 어댑터는 드라이버의 토큰 두 값을 복제한다.
- **미관측 전제**: P-1(POSIX 직접 문의 `syscall.Exec` 교체)·P-2(임대 루트와 브로커 루트의 동일성)·P-4(개발자 지침 쌍의 스레드 적용)는 읽기만 했거나 관측 불가다. M0 에서 측정한다.
- **도구 출처**: plan-audit 보고서가 설치된 `moai` 빌드(`0732cc699`)의 지연을 기록했다. 이 plan 의 `moai spec lint` 결과도 같은 빌드에서 얻었다.

## 후속 카드 문안 (리더가 큐에 올릴 준비된 한 단락, 한국어)

Codex 관리 카드 자식에 카드 시작 프롬프트 주입: t1440 으로 `moai codex -f lane` 의 카드별 자식이 옵트인(`MOAI_FACTORY_MANAGED`) 아래 관리 Codex 소유자로 뜨게 됐지만, 관리 소유자의 드라이버는 우선 턴("준비 완료라고 한 줄로 답해, 아직 작업은 시작하지 마") 뒤 유휴로 들어가 운영자 입력이나 브로커 메시지가 올 때까지 카드 작업을 시작하지 않는다. 그래서 운영자가 붙지 않은 무인 레인은 첫 카드 이후로 진행하지 않는다. 레인 루프가 우선 턴 직후 리스한 카드의 id와 시작 지시를 첫 작업 턴으로 주입하는 방법(주입 위치, 부모 REQ-MS-014의 "카드 처분 판정 금지" 경계, 재시작·재배달 시 중복 주입 방지, 헤드리스라 운영자가 보지 못하는 점)을 SPEC으로 정하고 구현하는 카드다. 기본 꺼짐(옵트인)은 유지하고, t1408(TUI 부착)·t1459(시그널·Start/Close) 범위는 건드리지 않는다.

## Run-phase first obligations (plan-audit r2 debt)

- **D2** — AC-CC-002 의 env 절: 시험은 통제된 환경을 고정하고 env 와 args 를 **리터럴 기대값**과 비교한다(호스트에 따라 달라지는 키 집합을 `red-baseline.md` 에 기록해 쓰지 않고, 변경 대상 함수에서 도출하지도 않는다). RED 는 **값만 바꾸는 변이**로 확인한다.
- **D3** — 마일스톤 M1 은 먼저 새 이음새(`managedCodexCardLaunchFunc` 등)를 비활성 선언으로 도입하고 패키지가 컴파일됨을 확인한 **뒤에야** 기준 동작에서 AC-CC-002 가 초록이라고 주장한다.

## §E.2 Run-phase Evidence

Run-phase implementer: manager-develop, cycle_type tdd, card t1440. Every figure below was observed in this run; outputs are verbatim but bounded (`WARN config` noise lines dropped). Verification commands ran under the lane scrub in one compound invocation, `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test ...` (the prefix is abbreviated to `<scrub> &&` below).

**Baseline-attribution.** Tree: worktree `.claude/worktrees/t1440`, branch `WT-codex-card-managed-path`, base `e1f790d7e`; run commits `409ced12c` (M1 RED + inert declarations), `a184aa89c` (M2), `8ad13b2e8` (M3); results below were measured at HEAD `8ad13b2e8` (the progress/spec-status commit that records them follows it and touches no code). Judging build: no installed `moai` binary judged anything — `go test`/`go vet` compile the tree under test; `golangci-lint` is the installed `v2.1.6` (the CI version). `go version go1.26.8 darwin/arm64`.

### M0 — premises

- **P-1 (H-1) TRUE.** The production default `defaultCodexDirectLaunch` (`internal/cli/codex_direct_posix.go:53`, `syscall.Exec`) replaces the launcher process, so the legacy direct-door lane loop does not survive its first card on POSIX. Measured by a re-exec scratch experiment (deleted before any commit; full record in `.moai/reports/t1440/p1-syscall-exec.md`): mode `real` — fake codex pid `74982` equals the child test process pid `74982`, the child printed `=== RUN` and never `child loop returned` / `--- PASS`, one invocation (`card=t1`); positive control mode `stub` — two invocations (`card=t1`, `card=t2`) with distinct pids and `child loop returned: <nil>`. This is a **separate defect candidate (Q4)**, reported and not fixed here: it lives on the switch-off direct door, which REQ-CC-002 freezes. The managed path keeps the launcher as parent, so it does not have it. The loop comment at `codex_launcher.go:911-913` ("no process replacement happens on this path") is wrong on POSIX for the direct door.
- **P-2 (lease root vs broker root).** Measured with a scratch test (deleted): in the fixture, `factoryCardRoot()` = `/private/var/folders/.../002` and `launchProjectRoot()` = `/var/folders/.../002` — lexically different on macOS temp dirs (symlink spelling), the same directory (`filepath.EvalSymlinks(root)` = `/private/var/...`). For a linked worktree base the loop refuses first (`factoryAssertParentCheckout`: `refused — this verb runs from the parent checkout`). Consequence: the two roots differ only in path spelling; AC-CC-007 (`sequential_cards_rebind`) exercises the real chain end to end (run recorded under the lease root spelling, owner registering under the launch root spelling) and is green, so the spelling difference does not break registration here. Not shown: a project whose `CLAUDE_PROJECT_DIR` points at a non-primary checkout while the lease root resolves to the primary one (the loop refuses that shape before the owner runs).
- **P-3 (same pid, same label, second registration replaces the bound row) TRUE.** `TestManagedCardChildSecondCardRebinds/sequential_cards_rebind` PASS: one lane endpoint, `BindingBound`, session `fake-thread-1`, two `method thread/start` lines. `store.go` not touched.
- **P-4 (does the developer-instruction pair reach the thread) NOT MEASURED** — needs a real codex; the fake App Server cannot observe it. The SPEC states it as an unobserved limit for the docs (spec §1.2 item 6).

### D3 — compile first, then green on baseline (leader-mandated order)

M1 introduced the seams as inert declarations before the tests (`managedCodexCardLaunchFunc` in `codex_launcher.go`; `managedLaneOperatorSource`, `managedOperatorPumpsCreated`, a stub pump and a stub `defaultManagedCodexCardLaunch` in `managed_operator_input.go`), then:

```
$ go build ./internal/cli          -> build_exit=0
$ go vet ./internal/cli            -> vet_exit=0   (with both new test files present)
```

and only then ran AC-CC-002 on the baseline behavior: `ok .../internal/cli 72.946s` (`-count=3`, 4 subtests each, all PASS). RED/GREEN record: `red-baseline.md` (committed in `409ced12c`, an ancestor of both implementation commits).

### D2 — controlled environment, literal values, value-only RED

`TestManagedCardChildSwitchOffKeepsDirectDoor` clears the whole process environment (`controlProcessEnv`) and sets a fixed ordered list, then compares the direct door's `*exec.Cmd` to LITERALS: `Path`, ordered `Args` (with a committed `AGENTS.local.md` so the `-c developer_instructions=<json>` pair is present; the JSON literal is written out by hand), `Dir`, the ordered `Env` slice, and stdio pointer identity, plus zero source reads and zero pumps. Value-only mutations (each reverted): stamp value `MOAI_KANBAN_BACKEND=gpt-MUTANT` in `codexCardLaunchEnv` -> FAIL at the Env comparison; every inherited entry suffixed `-MUTANT` in `codexChildEnv` -> FAIL at the Env comparison; `-C` operand `wt-MUTANT` -> FAIL at the Args comparison. Details in `red-baseline.md`.

### AC-by-AC (commands, observed output, tree HEAD `8ad13b2e8`)

| AC | Command (abbreviated) | Observed |
|---|---|---|
| AC-CC-001 | `<scrub> && go test ./internal/cli -run '^TestManagedCardChildLaneLoopUsesManagedOwner$' -count=1 -v` | `--- PASS ... (9.00s)`, subtests `switch_1`, `switch_TRUE_padded` PASS, `ok ... 10.132s` |
| AC-CC-002 | `-run '^TestManagedCardChildSwitchOffKeepsDirectDoor$' -count=1 -v` | `--- PASS ... (20.80s)`, 4 subtests `unset`,`empty`,`zero`,`yes` PASS. Existing direct-door tests: `-list` printed exactly `TestSD_AC003_CodexRelaunchPerCard`, `TestSD_AC004_CodexOtherFactoryShapesRefused`; `-run` -> both PASS |
| AC-CC-003 | `-list` of the 6 names printed exactly 6 names; `-run` -> `--- PASS` x6 (`TestSD_AC003_CodexRelaunchPerCard`, `...DivertsFactorySession`, `...CarriesLaunchDir`, `...KeepsSpawnDoor`, `...RequiresOptIn`, `...SkipsLaterDebugSteps`), `ok 5.052s`. `grep -c 'runManagedFactoryCodex(bin, args, env, dir, os.Stdin)' internal/cli/codex_launcher.go` -> `1`; same on `managed_codex_factory.go` (positive-control form) -> `0`. `git diff -U0` `os.Stdin` deletion check: see Gaps (the guard refused the `develop...HEAD` form; `git diff --stat e1f790d7e HEAD -- internal/cli/codex_launcher.go` shows 27 insertions, 3 deletions, and the plain seam line is unchanged — grep count 1) |
| AC-CC-004 | `-run '^TestManagedCardChildLaunchShape$' -count=1 -v` | PASS; subtests `no_dash_C_and_pair_and_dir`, `oversize_instruction_skips_owner`, `owner_refuses_dash_C` PASS |
| AC-CC-005 | `-run '^TestManagedCardChildEnvCarriesIdentity$' -count=1 -v` | `--- PASS ... (2.24s)` |
| AC-CC-006 | `-run '^TestManagedCardChildKeepsLauncherClaim$' -count=1 -v` | `--- PASS ... (4.40s)` |
| AC-CC-007 | `-run '^TestManagedCardChildSecondCardRebinds$' -count=1 -v` (and `-race -count=3`) | `--- PASS ... (13.68s)`; subtests `sequential_cards_rebind`, `start_failure_leaves_no_pending_row`, `positive_control_pending_row_visible` PASS; `-race -count=3`: PASS x3, no `DATA RACE` |
| AC-CC-008 | `git diff --stat e1f790d7e HEAD -- internal/cli/managed_codex_factory.go internal/cli/managed_factory_session.go internal/factorymsg/store.go` -> **no output**; positive control `git diff --stat e1f790d7e HEAD -- internal/cli/codex_launcher.go` -> `27 insertions(+), 3 deletions(-)`. `-race -run '^TestManagedCodexServerRequestPolicy$' -count=1` -> `ok 2.589s`; `-race -run '^TestManagedDriverIsolatesTurnFailure$' -count=1` -> `ok 2.013s` (`-list` printed both names). Loop delivery `TestManagedCardChildDeliversInboxThroughLoop` (`-race`, and `-race -count=3`) -> PASS | |
| AC-CC-009 | `-run '^TestManagedCardChildSessionEndContinuesLoop$' -count=1 -v` | PASS; `owner_error_logged_and_loop_continues`, `same_line_form_as_direct_door` PASS |
| AC-CC-010 | `-list` printed exactly `TestManagedCardChildOperatorInputReachesNextSession`, `TestManagedOperatorInputPumpDetachesEndedSession`; `-race -run '^(...)$' -count=1 -v` | both PASS: pump subtests (7) + loop subtests `exit_exit`, `quit_quit`, `exit_quit` PASS; no `DATA RACE`; `-race -count=3` of the pump + loop tests: PASS x3, `ok 105.531s` |
| AC-CC-011 | `GOOS=windows GOARCH=amd64 go build ./...` -> `winbuild_exit=0`; `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` -> `winvet_exit=0`; `grep -n 'syscall\.' internal/cli/managed_operator_input.go internal/cli/managed_operator_input_test.go internal/cli/managed_card_child_test.go` -> no output, `grep_exit=1`; positive control `grep -c 'syscall\.' internal/cli/codex_direct_posix.go` -> `2` | |
| AC-CC-012 | not run — sync-phase (manager-docs) | |
| AC-CC-013 | `<scrub> && go test ./internal/cli -list '^.*(Managed|managed).*$'` piped to `grep -c '^Test'` -> `88`; the equivalent `-list anaged` -> `88`; `-list 'TestManagedSwitchDoesNotReachCodexLaneLoop'` printed no `Test` name (old test gone). `-run '^.*(Managed|managed).*$' -count=1` -> `ok ... 94.580s`. `gofmt -l internal/cli` -> `internal/cli/todo_classify_llm_test.go` (**pre-existing**, a file this card does not touch: absent from `git diff --stat e1f790d7e HEAD`; none of the files this card changed is listed). `go vet ./internal/cli` -> `vet_exit=0`. `golangci-lint run ./internal/cli/...` -> `0 issues.` (golangci-lint `v2.1.6`, the CI version). `git diff --stat e1f790d7e HEAD -- .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001` -> no output | |

Wider scoped regression (card-adjacent): `<scrub> && go test ./internal/cli -run '^Test(SD_|Codex)' -count=1` -> `ok ... 216.327s`. The `cli-suite` slot was held for the heavy runs (`moai slot acquire --resource cli-suite --max-duration 30m` exit 0, released exit 0).

### Mutation results (applied, observed red, reverted)

| id | mutation | observed |
|---|---|---|
| mu1 | managed branch ignores the switch (`factoryLaunchEnabled` only) | AC-CC-002 `unset` FAIL: `direct door calls=0 lane seam calls=2 plain seam calls=0, want 2/0/0` |
| mu2 | `-C <wt>` kept in the managed argv | AC-CC-004 `no_dash_C_and_pair_and_dir` FAIL (argv carries `-C`, argv != [bin pair]) |
| mu3 | run id not added to the managed env | AC-CC-005 FAIL (`env MOAI_KANBAN_ID = "", want "run-cli"`); AC-CC-007 `sequential_cards_rebind` FAIL (`roster lanes = []`) |
| mu4 | loop `return err` after the session error line | AC-CC-009 `owner_error_logged_and_loop_continues` FAIL (`codex lane: boom`) |
| mu5 | `codexWorktreeAnchorLock` call added to the managed branch | AC-CC-006 FAIL (`anchor lock placed 2 times ... want 0`) |
| mu6 | managed branch calls the owner AND falls through to the direct door | AC-CC-001 `switch_1` FAIL — through the second stub's own `stage t1 run: illegal card transition` fatal (both stubs advance the card), i.e. the direct door was reached; the explicit `direct 0` assertion did not get to speak (right direction, indirect reason) |
| mu7 | lane seam default passes `os.Stdin` instead of an adapter | AC-CC-010 loop `exit_exit` FAIL: `lane loop did not finish within 30s` |
| mu8 | source reader returns raw chunks instead of single lines | AC-CC-010 `single_line_per_read`, `multi_line_chunk_after_exit`, `quit_token`, `buffer_saturation`, `source_eof` FAIL (`Read = "a\n/exit\nnext\n"`) |
| mu9 | end-token check drops `/quit` | AC-CC-010 `quit_token` FAIL (`Read = "next\n", want "", io.EOF`) |
| mu10 | plain seam default's `os.Stdin` replaced by an adapter | AC-CC-003 grep count `1` -> `0` (reverted -> `1`) |
| mu11 | pump built eagerly at loop start (switch-off path too) | AC-CC-002 `unset` FAIL: `1 operator-input pump(s) created on the switch-off path, want 0` |
| extra | adapter `Close` does not broadcast | AC-CC-010 `close_unblocks_read` FAIL: `Close did not unblock the Read within 1s` |

### Design deviations recorded

- The loop's managed branch lives inside `launchCodexCardSession` (after the size check, before the direct door code), with `runID` added to its parameter list; the direct-door lines are unchanged apart from that signature.
- The pump is reached through a package-level holder (`managedLanePump`), created lazily by the lane seam's default body and cleared by `defer endManagedLanePump()` in `runCodexFactoryLane`, because the seam keeps the plain seam's four-argument shape (no reader parameter). The counter `managedOperatorPumpsCreated` is production code, read only by tests.
- `TestManagedSwitchDoesNotReachCodexLaneLoop` was deleted (not renamed) and its assertion lives inverted in `TestManagedCardChildLaneLoopUsesManagedOwner`; `managed_optin_test.go` lost the `kanban` import and its header comment now describes the new routing.

### Gaps (not observed in this run)

- AC-CC-012 (docs, CHANGELOG): sync-phase work, not done here.
- AC-CC-003's `git diff -U0 develop...HEAD -- internal/cli/codex_launcher.go` form: the worktree guard refused the compound `git diff` forms with a trailing `; echo`; I substituted the single-command `git diff --stat e1f790d7e HEAD -- ...` forms and the grep count. The `os.Stdin`-deletion-line check was not run as such.
- The AC-CC-013 literal list command (`-list '^.*(Managed|managed).*$' | grep -c '^Test'`) was run in that form and returned `88`; the `-run` of the full managed selection ran once (`ok 94.580s`), not under `-count=3`.
- Mutations were not run for the `oversize_instruction_skips_owner` guard (managed branch moved ahead of the size check) nor for the REQ-CC-014 rollback (the rollback is the unchanged owner's, shown by the `start_failure_leaves_no_pending_row` positive-control pair only).
- Windows: the fake App Server tests skip on Windows by design; the cross build and cross vet are the only Windows evidence.
- Real codex (P-4), a headed terminal, and an unattended-lane run were not exercised.

### Residual-risk

- Unattended lanes still do not advance past the first card (Q1, decided; docs carry it in sync).
- A fatal-ended session whose reader goroutine is blocked on the full 8-slot channel leaks that goroutine until process exit (plan-audit r2 D1); no input is stolen by it, and the owner/driver are unchanged by constraint. The pump's own source-reading goroutine is likewise uninterruptible while blocked inside the source's `Read` (stdin) and ends with the process.
- The pump copies the driver's two end-token values; a driver change that adds a token would drift silently except through the loop-level test, which only runs `/exit` and `/quit`.
- P-1: the legacy direct-door loop replaces the launcher process on POSIX (separate defect candidate, switch-off path, left unchanged by REQ-CC-002).
- Test timings are load-sensitive (the fixtures run a real git + factory store per card, 3-8 s each on this machine); the 30 s watchdogs are the SPEC's own bound.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready (pending sync-phase for AC-CC-012)
run_complete_at: 2026-10-03
run_head: see the progress/spec-status commit that follows `8ad13b2e8`
commits: `409ced12c` (M1 RED + inert declarations + red-baseline.md), `a184aa89c` (M2 branch and launch shape), `8ad13b2e8` (M3 operator-input pump), plus this progress/status commit
ac_status: AC-CC-001..011 and 013 observed PASS (table above); AC-CC-012 deferred to sync
test_count_managed_selection: 88 (79 - 1 + 10), observed with `-list` in both selector forms
lint: golangci-lint v2.1.6 `0 issues.`; go vet exit 0; gofmt only the pre-existing `todo_classify_llm_test.go`
mutations: mu1..mu11 all observed red (mu6 indirectly), one extra Close mutant
open_items_for_leader: P-1 defect candidate (direct door `syscall.Exec` ends the legacy lane loop after card 1 on POSIX); follow-up card text for the card-start prompt gap is in this file above

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
