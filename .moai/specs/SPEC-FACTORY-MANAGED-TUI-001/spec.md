---
id: SPEC-FACTORY-MANAGED-TUI-001
title: "Factory managed Codex session — operator TUI attach and the headless-to-interactive transition (SPEC-FACTORY-MANAGED-SESSION-001 Known debt 1, sync audit F2)"
version: "0.3.3"
status: completed
created: 2026-10-03
updated: 2026-10-04
author: GOOS (manager-spec)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-FACTORY-MANAGED-SESSION-001, SPEC-FACTORY-MANAGED-HARDEN-001]
depends_on: [SPEC-FACTORY-MANAGED-SESSION-001, SPEC-FACTORY-MANAGED-HARDEN-001]
tags: "factory, managed-session, codex, tui, app-server, terminal-ownership, headless-to-interactive"
amendment_of: SPEC-FACTORY-MANAGED-TUI-001
---

# SPEC-FACTORY-MANAGED-TUI-001 — Managed Codex session: operator TUI attach

## HISTORY

### Amendments

Amendment 2 (version 0.3.3):

| Field | Value |
|---|---|
| prior_completed_version | 0.3.2 |
| prior_completed_sha | 985cedfb788dad89bbffb5e010e79e96ae95f960 (the re-close sync commit; progress.md §E.4 `sync_commit_sha` carries its short form) |
| prior_completed_record | progress.md §E.4 sync_commit_sha |
| rationale | The delta sync audit (`.moai/reports/t1408/sync-audit-iter2.md`, local-only) reproduced two bypasses of the safe session-log open that known debt 15 did not name, so it understated the limit; the operator document (`tui-log-race`) and the CHANGELOG already carry the fuller wording. The frontmatter record pointer of Amendment 1 also named the renamed key `sync_commit_sha`. |
| scope | spec.md (§H item 15, the Amendment 1 `prior_completed_record` cell, frontmatter, this HISTORY) + acceptance.md/design.md version lines only. REQ count stays 14; AC count stays 16; no requirement or criterion text changes. progress.md, CHANGELOG and the operator document are the sync owner's and are not touched. |
| weakening note | None: the limit is stated more broadly, not less. |
| re_close_path | SPEC returns to `completed` on a later sync commit owned by manager-docs. This amendment does not set `completed`. |

Amendment 1 (version 0.3.2):

| Field | Value |
|---|---|
| prior_completed_version | 0.3.1 |
| prior_completed_sha | 40aa3aedf390a28beac616d53c247d63e66703c2 (the sync commit; progress.md §E.4 `sync_commit_sha` carries its short form, backfilled by `3ca47fbe3`) |
| prior_completed_record | progress.md §E.4 superseded_sync_commit_sha (the key that now holds the first close; the field was renamed when the SPEC was re-closed) |
| rationale | The independent sync audit (`.moai/reports/t1408/sync-audit.md`, FAIL 76.2) found two SPEC-text defects and two code defects. F1: AC-MT-010 names `TestManagedSwitchDoesNotReachCodexLaneLoop`, which card t1440 removed upstream (commit `a184aa89c`); the name is gone from `internal/cli` (measured: the grep finds nothing). F7: AC-MT-013 demanded that the old "Codex 관리 세션은 화면에 아무것도 보여 주지 않는다" sentence be absent, while the completed SPEC-FACTORY-MANAGED-CARD-CHILD-001 AC-CC-012 demands that same sentence be present. The code repairs F5 (connection lost before attach) and F6 (safe session-log open) landed in `42a952661` with tests `TestManagedCodexConnectionLostBeforeAttachStopsTUI` and `TestManagedTUILogFileIsSafe`; F4 was deliberately not fixed and is recorded as known debt 14. |
| scope | acceptance.md (AC-MT-010 test names and positive-control line numbers, its §1.1 block, AC-MT-013 grep rows and a record of observed counts, the coverage statement, an edge case for F4) + spec.md (frontmatter, this HISTORY, §H debts 14 and 15) + design.md (version, one D-5 note). REQ count stays 14; AC count stays 16; no requirement text changes. progress.md, CHANGELOG and the operator document are the sync owner's and are not touched. |
| weakening note | AC-MT-010 loses one named test (the lane-loop one, removed upstream by t1440) and gains `TestManagedCardChildSwitchOffKeepsDirectDoor`. The intent (no managed branch without the opt-in) is carried by the three surviving opt-in tests, that replacement, and AC-MT-016. AC-MT-013 no longer constrains the count of the old "does not show anything" sentence on the Codex-session line: that sentence is AC-CC-012's, not this SPEC's, and the headless fact it carried is now a card-child bullet this AC checks directly. |
| re_close_path | SPEC returns to `completed` on a later sync commit owned by manager-docs. This amendment does not set `completed`. |

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.3.3 | 2026-10-04 | manager-spec | Completed-SPEC amendment (card t1408) — known debt 15 restated to the full limit (hard link, symlinked parent directory, final-check race); Amendment 1 record pointer corrected. See Amendments. |
| 0.3.2 | 2026-10-04 | manager-spec | Completed-SPEC amendment (card t1408) — sync audit F1, F7; known debts 14, 15 (F4 and the log-open race). See Amendments. |
| 0.3.1 | 2026-10-03 | manager-spec | Final delta-audit wording fixes B1, B2, O1, O2, O8 (no new audit; PASS-WITH-DEBT convergence). |
| 0.3.0 | 2026-10-03 | manager-spec | Plan-audit iteration 2 delta revision (FAIL 0.86; scoped to D1-D5 and D7). Blocked-call release redesigned (WebSocket close, write deadline); AC-MT-015 made operator-held; t1408-before-t1459 decided; AC-MT-014 split and made checkable. See progress.md §E.1. |
| 0.2.0 | 2026-10-03 | manager-spec | Plan-audit iteration 1 revision (FAIL 0.63, defects D1-D19; local report `.moai/reports/t1408/plan-audit-iter1.md`). Changed: REQ-MT-004/006/007/008/013/014 wording; AC-MT-016 added (AC 16); the stale-busy rule (deferral never self-clears); log-sink lifetime; D-9 merge-order text corrected against the held t1459 draft; M1 compile-stub RED strategy; milestone exits re-gated; EXCL-syscall clauses. Disposition table in progress.md §E.1. |
| 0.1.0 | 2026-10-03 | manager-spec | Initial SPEC (card t1408). Pays Known debt 1 of the completed SPEC-FACTORY-MANAGED-SESSION-001 (independent sync audit `.moai/reports/t1375/sync-audit.md`, finding F2; operator decision 2026-10-02). The parent SPEC directories are not modified (design.md D-1). Hard-dependent on the landed SPEC-FACTORY-MANAGED-HARDEN-001 (server-request answering, per-turn failure isolation). |

## §A. Overview

SPEC-FACTORY-MANAGED-SESSION-001 (status `completed`, amended twice) states in its Amendment 1 that the managed Codex launcher is a **headless App Server owner**: it spawns `codex app-server`, drives it over JSON-RPC from the launcher process, and starts no TUI. Its Known debt 1 records the consequence, by the sync-audit F2 reading of the source and not by a run: a factory-environment launch shows no model output, and the operator's keyboard reaches the model only through the launcher's own stdin line reader. TUI attach was deferred to a follow-up card. The operator decided on 2026-10-02 to pay that debt (card t1408).

This SPEC attaches the Codex TUI as a client of the owned App Server, on the same thread the broker endpoint is bound to, and fixes everything the attach changes: who owns the terminal, how operator turns and broker-injected turns share one thread, which server requests the launcher may still answer on the operator's behalf, how the session ends, and what happens when a TUI cannot run.

### A.1 Observation grades

"Measured" means a command was run in this plan run against tree `2b9e4a4d0` (the card base, a clean worktree). "Inferred" means read from a document or schema and not run. "Not observed" means no check in this SPEC's automated suite can observe it.

| # | Fact | Grade | How |
|---|------|-------|-----|
| P1 | The launcher starts no TUI. Searching `internal/cli` for the literal `remote-auth-token-env` matches nothing (exit 1). | Measured | `grep -rln --include='*.go' 'remote-auth-token-env' internal/cli` |
| P2 | codex-cli 0.160.0 offers `codex resume --remote <ADDR>` and `--remote-auth-token-env <ENV_VAR>`; `codex resume --help` took 0.040 s. | Measured | `codex --version`, `codex resume --help`, `time` |
| P3 | `codex resume --remote ws://127.0.0.1:1 --remote-auth-token-env NOPE_TOKEN <uuid>` with stdin `/dev/null`: exit 1 `environment variable NOPE_TOKEN is not set` when the variable is unset; exit 1 `stdin is not a terminal` when it is set; an unknown flag exits 2 `unexpected argument`. So a TUI refuses to run without a terminal before it connects, and a codex without the option fails with a usage error. | Measured | three runs, outputs in design.md D-4 |
| P4 | In the generated 0.160.0 server-request schemas, six request kinds carry `turnId` (`item/commandExecution/requestApproval`, `item/fileChange/requestApproval`, `item/tool/requestUserInput`, `mcpServer/elicitation/request` where it is optional, `item/permissions/requestApproval`, `item/tool/call`) and four do not (`account/chatgptAuthTokens/refresh`, `attestation/generate`, `applyPatchApproval`, `execCommandApproval`). `thread/resume` rejoins a running thread by id. | Measured (schema) | `codex app-server generate-json-schema` |
| P5 | `turn/start` against a thread with an active turn steers that turn instead of starting a second one. | Inferred | schema text: `turnTrigger` is "ignored when this request steers an already-active turn" |
| P6 | The launcher's own connection receives the turn lifecycle frames of turns another client (the TUI) starts. | Not observed | the prior art (commit `0e3c66454`, not in this tree) gated on it; no run here |
| P7 | When two connections are subscribed to one thread, a server request goes to the connection that started the turn only, or to both. | Not observed | the app-server README states nothing on it (fetched this run; a summarizing model read it) |
| P8 | A remote TUI needs a persisted rollout before it can resume a thread. | Prior art, version unknown | comment in `0e3c66454:internal/cli/managed_codex_factory.go`; the existing priming turn persists one |
| P9 | `codex resume --remote <thread-uuid>` finds a loaded thread without a cwd filter; the TUI renders turns the launcher started; `/quit` exits 0; the terminal is restored after an interrupt. | Not observed | AC-MT-015 (manual) |
| P10 | The Codex path writes to the terminal from three places: the owner's log default (`:187`), the App Server child's stderr (`:689`), and one direct `Factory inbox:` line in the driver (`managed_factory_session.go:375`). The same grep also prints `managed_factory_session.go:179` (the Claude/GLM stream owner's child stderr), which is out of scope. | Measured | `grep -n 'os\.Stderr' internal/cli/managed_codex_factory.go internal/cli/managed_factory_session.go` |

## §B. Scope

Four changes inside the managed Codex owner and the delivery driver; the Claude/GLM stream owners are untouched.

1. **Attach.** After the priming turn, the owner starts `codex resume --remote …` as a child process on the owned App Server's thread, with the token in an environment variable only.
2. **Terminal and output ownership.** The TUI inherits the real terminal; the launcher stops reading stdin and stops writing to the terminal, diverting its own log lines and the App Server's stderr to a session log file.
3. **A shared thread.** Operator turns start inside the TUI; broker deliveries start from the launcher. The launcher defers delivery while any turn is active, keeps its connection reader alive, and stops auto-declining requests that belong to operator-started turns.
4. **Lifecycle and fallback.** TUI exit ends the session and sets the launcher's exit status; session end and App Server death stop the TUI; a missing terminal, a codex without `--remote`, or an operator opt-out keeps today's headless behavior with one notice.

## §C. Requirements (GEARS)

### C.1 Attach and identity

- **REQ-MT-001** — **Where** the managed-session opt-in is active (`MOAI_FACTORY_MANAGED` set to `1` or `true` together with the factory stamps — the gate of SPEC-FACTORY-MANAGED-SESSION-001, unchanged), **when** the Codex session owner has completed the priming turn on a thread whose broker endpoint is already bound to that thread's id and the attach preconditions of REQ-MT-003 hold, the owner shall start the Codex TUI as a child process of the launcher, connected to the owned App Server and resuming that same thread id; outside the opt-in gate the owner shall neither run the capability probe of REQ-MT-003 nor start a TUI.
- **REQ-MT-002** — The owner shall hand the App Server capability token to the TUI child only as the value of an environment variable whose name, and never whose value, appears on the TUI command line; the token value shall appear in no command-line argument of any child process, and the App Server shall keep receiving it through its token file.
- **REQ-MT-003** — The owner shall attach the TUI only where all three preconditions hold: (a) the stream the owner would give the TUI as stdin is a terminal file and its stdout is a terminal; (b) the `resume --help` output of the codex binary names both the `--remote` and the `--remote-auth-token-env` option within the probe timeout set in `internal/config/defaults.go`; (c) the operator opt-out variable `MOAI_FACTORY_MANAGED_TUI` is not set to `0`, `false` or `off` (case-insensitive, trimmed).
- **REQ-MT-004** — **When** an attach precondition does not hold, or the TUI child cannot be started, the owner shall continue as the headless session does today — operator stdin lines and claimed broker messages through the serial delivery loop — and shall write exactly one line to the launcher's stderr naming the reason before the session continues; the fallback shall not be an error and shall not change the launcher's exit status. One exception to "as today": the attach decision is made before the App Server starts, so when the TUI fails to start after that, the App Server's stderr stays in the session log file of REQ-MT-006 and the launcher prints that file's path once.

### C.2 Terminal ownership and the shared thread

- **REQ-MT-005** — **While** the TUI is attached, the TUI child shall inherit the launcher's stdin, stdout and stderr, and the delivery driver shall not read the operator's stdin or interpret `/exit` and `/quit` lines; operator turns are started by the TUI directly on the App Server and never pass through the launcher's turn queue.
- **REQ-MT-006** — **While** the TUI is attached, the launcher shall write nothing to the terminal: the operator-facing log lines of the owner and of the delivery driver, and the App Server child's stderr, shall go to a session log file under the project's `.moai/logs/` directory whose path the launcher prints once on stderr before the TUI starts, and **when** the TUI child has been reaped the owner's and the driver's log lines, teardown lines included, shall resume on stderr (the App Server child keeps writing to the file, because its stderr is fixed when it starts).
- **REQ-MT-007** — **While** the TUI is attached, the owner shall track whether the thread has an active turn from the turn lifecycle frames of every turn whichever client started it, shall keep its connection reader running between the owner's own turns without letting operator-started turns back up behind the owner's event channel, and the delivery driver shall claim broker messages only while no turn is active, so that an inbox batch is deferred, not dropped, while an operator turn runs; deferral shall never end by elapsed time: a turn that stays active longer than `DefaultManagedCodexTurnTimeout` shall only produce one log line per elapsed interval, so a legitimate long operator turn is never steered into and a lost completion frame is visible in the log instead of silent (known debt 13).
- **REQ-MT-008** — **While** the TUI is attached, the owner shall not answer a server request whose `turnId` names a turn the owner did not start, so that the operator answers it in the TUI, except a request that arrives while the owner's own `turn/start` is outstanding and its turn id is not yet known, which the owner answers as before (known debt 9); every other server request, and every request while no TUI is attached, shall keep the answer fixed by SPEC-FACTORY-MANAGED-HARDEN-001, and no answer shall be an accepting decision.
- **REQ-MT-009** — The TUI command line shall carry none of the MoAI MCP approval arguments the launcher generates for the App Server, shall carry the operator's own `-c`/`--config` and `-m`/`--model` values unchanged, and the project-level capability-based approval mode (`default_tools_approval_mode = "writes"`) shall stay unchanged.

### C.3 Lifecycle

- **REQ-MT-010** — **When** the TUI child ends on its own, the delivery driver shall stop, the owner shall run its existing teardown, and the launcher shall exit with status 0 when the TUI exited 0 and with the TUI's exit status (status 1 when a signal ended it) otherwise, unless the session had already ended with its own error, in which case that error stands.
- **REQ-MT-011** — **When** the App Server connection closes while the TUI runs, or the delivery driver ends the session for any other reason, the owner shall stop the TUI — interrupt, wait the grace period set in `internal/config/defaults.go`, then kill — and wait for it before the launcher returns, as a step of the owner's session teardown, so that no TUI child outlives the launcher.

### C.4 Portability and disclosure

- **REQ-MT-012** — The delivery shall add no `syscall` reference to any `managed_*` source or test file (the parent SPEC's AC-MS-014 grep stays at 0 lines), shall keep the `GOOS=windows GOARCH=amd64` cross build passing, and shall leave `internal/factorymsg/store.go`, the directories of SPEC-FACTORY-MANAGED-SESSION-001 and SPEC-FACTORY-MANAGED-HARDEN-001, and the App Server approval arguments unchanged.
- **REQ-MT-013** — The operator documentation `.moai/docs/factory-managed-session.md` and this SPEC's CHANGELOG entry shall state that Known debt 1 of SPEC-FACTORY-MANAGED-SESSION-001 is addressed and that real-TUI behavior is unverified until the operator has run the operator-held check AC-MT-015 and recorded steps 1, 2, 3 and 6 as observed (only then may the word "resolved" be used), shall state the opt-out, the log file location and the changed `/exit` and `/quit` handling, shall state the signal-handling gap as still open and owned by card t1459, and shall name every remaining limit without claiming behavior that was not observed; the remaining limits include the premises of §A.1 and the known debts of §H that no automated check observes.
- **REQ-MT-014** — The run phase shall first land one behavior-neutral commit of compile stubs (the symbols the tests need, with the unchanged headless behavior, so the existing suites stay green), and shall then land each reproduction test's RED baseline, as a named failing assertion per acceptance criterion, as the tracked file `.moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md` together with the reproduction tests, in a commit that precedes every commit carrying a corresponding fix, so the commit graph witnesses the ordering (the stub commit carries no fix); `.moai/reports/t1408/` holds local copies only.

EXCL-syscall: this SPEC introduces no `syscall` reference anywhere; the Windows cross build of AC-MT-012 is the proof, and the one place a platform primitive differs (`os.Interrupt` is unsupported on Windows) is handled by a kill-only path that needs no build constraint.

## §D. Constraints

- **Gate unchanged.** `MOAI_FACTORY_MANAGED=1|true` plus the factory stamps stays the only way into the managed layer, default off. No second opt-in is added; the opt-out of REQ-MT-003(c) only turns the new behavior off inside the existing gate.
- **Delivery-only boundary.** The parent design D-1 stands: no card disposition, merge window, re-alert or completion judgment from this layer (the parent AC-MS-010 grep stays at 0 lines).
- **No approval widening.** Every answer the launcher gives stays a refusal or a grant-nothing result (SPEC-FACTORY-MANAGED-HARDEN-001 REQ-MH-003). The MoAI broker tool pre-approval stays on the App Server command line only.
- **Broker untouched.** `internal/factorymsg/store.go` is not modified.
- **No hardcoding.** New numeric bounds (probe timeout, TUI stop grace) go in `internal/config/defaults.go`; new environment names (`MOAI_FACTORY_APP_SERVER_TOKEN`, `MOAI_FACTORY_MANAGED_TUI`) go in `internal/config/envkeys.go`.
- **Interface stability.** This SPEC leaves the three methods of the `managedSession` interface and the signature of `driveManagedFactorySession` as they are; new behavior reaches the driver through an optional capability interface the owner implements (design.md D-8). The held t1459 draft changes that function's signature (a leading `ctx`); design.md D-9 states how the two compose.
- **Windows.** `GOOS=windows GOARCH=amd64 go build ./...` stays green; the new code is zero-`syscall` and terminal detection uses the dependency the package already uses.
- **Verification load.** Lane-local verification runs only the tests this change can affect; the `internal/cli` package suite does not run locally. Minute-scale suites run only inside a `moai slot` lease (`.claude/rules/local/gitflow-lane-protocol.md` §8).
- **Plan-audit gate.** A passing plan audit is the run-phase entry condition.
- **EXCL-syscall.** No `syscall` reference is introduced; the cross build of AC-MT-012 is the proof.

## §E. Success criteria summary

All of AC-MT-001..014 and 016 in `acceptance.md` (AC-MT-015 is operator-held and not release-blocking). In particular: (1) the reproduction tests are RED on the card base and GREEN after the fix, with the RED baseline in a tracked file in an earlier commit; (2) a loopback test with a fake codex shim (App Server role and TUI role in one re-exec helper) runs the whole attach → inbox delivery → receipt → TUI exit path with no real codex; (3) the parent suites for the managed layer and the four opt-in tests stay green; (4) the cross build and the zero-`syscall` grep pass; (5) one explicit manual check, AC-MT-015, is labeled as not run in CI and is the only observation of a real interactive TUI. EXCL-syscall: the zero-`syscall` property is proved by AC-MT-012's grep and cross build.

## §F. Exclusions

### Out of Scope — F5 signal handling and the Start/Close lifecycle (card t1459)

- Launcher handling of SIGINT, SIGTERM and SIGHUP and the redesign of `Start` and `Close` (concurrency-safe single-run teardown, publication order) belong to card t1459 (on hold). This SPEC adds the TUI child to the existing teardown as one step and one field; it does not restructure `Start` or `Close`. design.md D-9 states the merge-order interaction.

### Out of Scope — minor items F8, F9, F13 (card t1410)

- Token directory residue on failed starts (F8), the 10-second handshake budget (F9) and the `/readyz` redirect re-check (F13) stay with t1410. This SPEC reads the token file at attach time so that its only edit in the `Start` region is one line (the App Server child's stderr destination, `:689`, directly above F8's failure branch); design.md D-9 lists it.

### Out of Scope — Codex lane child sessions (card t1440)

- `moai codex -f lane` card children are not managed sessions and stay on the direct door; wiring them is t1440's.

### Out of Scope — file ownership boundaries with card t1440 (lane-17)

- Card t1440 edits `internal/cli/codex_launcher.go` (the `runCodexFactoryLane` / `launchCodexCardSession` loop, the `managedFactoryCodexLaunchFunc` seam default, and the comment near the managed divert) and creates `internal/cli/managed_operator_input.go` and `internal/cli/managed_card_child_test.go`. This SPEC plans no edit in `codex_launcher.go` and uses neither new file name. Its production edits are confined to `internal/cli/managed_codex_factory.go`, `internal/cli/managed_factory_session.go`, new files with distinct names (`managed_codex_tui.go`, `managed_codex_tui_test.go`), and the two config constant files.
- The attach decision is made inside the owner entry `runManagedFactoryCodex`, so no launcher-side wiring change is needed. If one proves unavoidable in the run phase it is returned to the leader as a blocker, not made.
- `.moai/docs/factory-managed-session.md` "알려진 한계" bullets are touched by both cards: this SPEC changes them as one minimal, separate bullet edit (replace the "screen shows nothing" bullet; add new bullets after it) so the later merger can reconcile.

### Out of Scope — Claude and GLM managed owners

- The stream-json owners and their stdin handling are unchanged; the arrival-order FIFO of AC-MS-008 still governs them and the headless Codex path.

### Out of Scope — pseudo-terminal proxying

- The launcher does not allocate a pseudo-terminal, relay keystrokes, or multiplex its own output into the TUI. That needs terminal-mode system calls and breaks the zero-`syscall` property (design.md D-2). EXCL-syscall: this exclusion exists to keep `syscall` out; AC-MT-012's cross build is the proof.

### Out of Scope — `turn/steer`, `turn/interrupt` and `thread/inject_items`

- The launcher starts turns with `turn/start` only, and only while no turn is active; it never steers or interrupts an operator turn and does not use `thread/inject_items`.

### Out of Scope — broker changes

- No attempt cap, nack call or claim release. A deferred inbox batch simply waits until the thread is idle; redelivery stays the broker's claim-lease policy.

### Out of Scope — Windows runtime certification and live Codex gating in CI

- The Windows cross build and the fake-codex tests are in scope; running a real TUI on Windows is not. No automated test starts a real codex; AC-MT-015 is the only real-TUI observation and it is manual.

### Out of Scope — parent SPEC edits

- SPEC-FACTORY-MANAGED-SESSION-001 and SPEC-FACTORY-MANAGED-HARDEN-001 (body, frontmatter, HISTORY) are not modified. The link is `related_specs`, `depends_on` and this text.

## §G. Cross-references

- `plan.md` — milestones M1..M6, pre-flight measurements, risks.
- `acceptance.md` — AC-MT-001..016, RED-now evidence ledger, mutant probe, edge cases, the manual check procedure.
- `design.md` — D-1 SPEC form, D-2 terminal ownership and input priority, D-3 attach sequence, identity and token, D-4 capability probe and fallback, D-5 approvals and request scoping, D-6 terminal-clean output, D-7 busy tracking, D-8 lifecycle and exit status, D-9 composition with t1459 and t1410, D-10 portability and test seams.
- `.moai/reports/t1408/verdict.md` — the card premise measurement (local copy; `.moai/reports/*` is gitignored).
- SPEC-FACTORY-MANAGED-SESSION-001 — parent; Known debt 1, REQ-MS-009, AC-MS-008, AC-MS-012. SPEC-FACTORY-MANAGED-HARDEN-001 — F3/F4 policy this SPEC scopes.

## §H. Known debts carried by this SPEC

These are stated at plan time and are not closed by it. REQ-MT-013 requires the documentation to repeat them.

1. **No real interactive TUI run is observed in CI.** P6, P7, P9 and the TUI's rendering of launcher-started turns rest on AC-MT-015 (manual) and on inference. A green CI says the owner behaves correctly against a fake codex, not that a real TUI behaves as designed.
2. **Server-request routing with two subscribed connections is not observed (P7).** The scoping of REQ-MT-008 is built to hold under either routing, but the four request kinds without `turnId` (and an elicitation whose `turnId` is null) keep today's automatic answer and could race the operator's answer in the TUI if the server broadcasts them.
3. **Busy tracking depends on P6.** If the launcher's connection never receives lifecycle frames of operator turns, the launcher cannot tell the thread is busy; delivery then follows P5 (steering into the active turn) instead of waiting. Not a deadlock, but not the designed deferral.
4. **A race window remains** between the idle check and `turn/start`; a turn the operator starts in that window is steered into, not preempted.
5. **A forced TUI stop may leave the terminal in a TUI mode.** The owner interrupts, waits, then kills; terminal restoration is the TUI's job and is not observed.
6. **`/exit` and `/quit` typed at the terminal no longer reach the launcher** while the TUI is attached; the TUI's own commands apply.
7. **The operator-visible traces of HARDEN-001 (declined requests, turn failures) move to the session log file while the TUI runs.** They are not on the screen.
8. **Terminal detection cannot be tested with a real tty in CI**; the predicate is a test seam and the production default is only exercised by AC-MT-015.
9. **A request from an operator-started turn that arrives while the owner's own `turn/start` is outstanding and its turn id is not yet known is declined** (benefit of the doubt goes to the owner's turn; REQ-MT-008's exception).
10. **Signals (t1459) are still open.** A SIGHUP or SIGINT delivered to the launcher's process group ends it by the default action and can orphan the App Server.
11. **The token value is in the TUI child's environment**, visible to same-user process inspection on platforms that expose it.
12. **Windows runtime and the codex binary version matrix are not certified**; the probe is feature detection, not a version list.
13. **A lost `turn/completed` frame starves delivery.** Busy never clears by time (REQ-MT-007), so the broker batch stays unclaimed and the only signal is a repeated line in the session log file, which is off-screen. Releasing after the ceiling would steer broker prompts into a legitimate long operator turn; this SPEC picks visible starvation over silent steering.
14. **An operator turn that starts and completes inside the owner's armed window (before the owner's `turn/start` response) is attributed to the owner; the owner's own lifecycle frames are then filtered and the owner waits for its turn until the turn timeout.** The fix needs response-id attribution with frame holding and would touch the REQ-MT-008 exception. Reproduced at method level by `TestManagedOperatorTurnInsideArmedWindowKnownDebt`, which pins the current behavior. Not fixed (sync audit F4).
15. **The session-log open is only partly hardened.** Replacing the log path's leaf with a symlink and an over-open file mode are closed by `TestManagedTUILogFileIsSafe`. A **hard link** at the log path (the same-file check passes; appends and the mode tightening hit the original file) and a **symlinked parent directory** (`.moai/logs` pointing outside the project) remain open known limits, and the race between the final check and the use of the file also remains. All of them need the same user's write access to the log location. Reproduced by the delta sync auditor (`.moai/reports/t1408/sync-audit-iter2.md`, local-only). The operator document anchor `tui-log-race` and the CHANGELOG carry the same wording.
