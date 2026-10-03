---
id: SPEC-FACTORY-MANAGED-TUI-001
title: "design.md — TUI attach design decisions"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# design.md — decisions D-1 .. D-10

> Stateless on the status axis (`spec-frontmatter-schema.md` § Artifact Statelessness); the lifecycle lives in `spec.md`. Identifiers (function, field and interface names) are working names for explanation and the run phase may rename them. Grades: **measured** = a command ran in this plan run on tree `2b9e4a4d0`; **inferred** = read from a schema or document; **not observed** = no automated check observes it. Premise numbers P1..P10 are the table in `spec.md` §A.1.

## D-1 — A new SPEC, not an amendment

**Decision.** A new SPEC, `SPEC-FACTORY-MANAGED-TUI-001`, with `related_specs` and `depends_on` pointing at the two completed managed-session SPECs. Neither parent directory is edited.

**Why.** (1) SPEC-FACTORY-MANAGED-SESSION-001 is `completed` and was amended twice already; a third amendment would return it to `in-progress`, invalidate its cached plan-audit PASS, and force a delta audit and a second re-close for a change that adds a whole new surface (a second owned child process, terminal ownership, two new environment names). (2) The parent's own text does not need to change to stay true: REQ-MS-009 says the launcher's automatic approval arguments go only to the App Server it owns, and that remains the behavior (REQ-MT-009 keeps the TUI free of them); its parenthesis "the TUI is not delivered by this SPEC" stays accurate as a statement about that SPEC. Known debt 1 is paid by a later SPEC, which is what the debt text asks for ("owed to a follow-up card"). (3) The SPEC-FACTORY-MANAGED-HARDEN-001 precedent did the same for F3/F4: parent untouched, link by `related_specs`, disclosure in the operator document and CHANGELOG (REQ-MT-013).

**Rejected.** (a) *Amend SESSION-001* — cost above, and it would reopen REQ-MS-009 and AC-MS-012, whose test `TestMoAIMCPApprovalArgsOnlyTargetMoAI` is a PRESERVE target. (b) *Amend HARDEN-001* — its scope is F3/F4 and its §F already names TUI attach as card t1408's; it is also `completed` at 0.5.2. (c) *Fold into the t1459 SPEC* — that SPEC is a draft on a held branch with a different subject (signals and `Start`/`Close`); merging two open subjects would couple their audits.

**Debt-payoff record.** The new SPEC's HISTORY names Known debt 1 and audit finding F2; the sync phase updates the operator document (the two sentences that say the screen shows nothing) and adds the CHANGELOG entry; the sync-phase close note cites this SPEC against the parent's Known debt 1.

## D-2 — Who owns the terminal, and what happens to input priority

**Decision.** The TUI child inherits the launcher's real `stdin`, `stdout` and `stderr` (the three `*os.File` values, not pipes). From the moment the TUI is attached the launcher's driver does not start its stdin reader. Operator keystrokes therefore go terminal → TUI → App Server and never enter the launcher's turn queue. To make that possible, attach is possible only when the launcher's stdin is a real `*os.File` that is a terminal: an `io.Reader` that is not a file would make `os/exec` interpose a pipe, and the TUI refuses a non-terminal (P3). That also keeps every existing managed Codex test, which passes pipe readers, on the headless path untouched.

**Input priority, stated precisely.** Today `AC-MS-008` promises only arrival-order FIFO between operator lines and broker batches (parent Amendment 1, F10). With the TUI attached the two no longer share a queue:

- The launcher's queue holds broker batches only. The driver claims a batch only while no turn is active (D-7) and delivers it at once.
- An operator turn starts in the TUI whenever the operator submits. If the thread is idle it starts immediately; if a launcher-injected turn is running, what the TUI does with typed input (queue, steer) is the TUI's, and is not observed (AC-MT-015).
- So operator input is never delayed by the launcher's queue, and broker delivery never preempts an active turn. This is a deferral rule, not a strict priority: when the thread is idle and both arrive within the race window (known debt 4), the server decides the order.
- `AC-MS-008` itself is unchanged and still governs the stream owners and the headless Codex path. In attached mode it is inapplicable to operator input, not violated. `/exit` and `/quit` typed at the terminal are no longer launcher commands while attached; the TUI's own commands apply (known debt 6).
- The managed `turn/start` hazard that makes the deferral necessary is P5: a `turn/start` on a thread with an active turn most likely steers that turn, so an ungated launcher would inject the broker metadata prompt into the middle of an operator's turn and then wait on the operator's turn for up to the 10-minute turn timeout, which is session-fatal.

**Rejected.** (a) *A launcher-owned pseudo-terminal proxy* so the launcher keeps seeing keystrokes and can queue operator lines itself — needs terminal-mode system calls and a pty dependency, breaks the zero-`syscall` property and the single code path on Windows (parent D-4), and gives the launcher a screen-scraping job it does not need. (b) *The launcher keeps reading stdin and forwards to the TUI over a pipe* — the TUI will not run without a terminal (P3). (c) *Keep the FIFO queue for operator lines and send them as launcher turns* — that is today's headless design with a TUI that shows nothing typed in it. (d) *Strict operator priority by interrupting launcher turns* — needs `turn/interrupt`, which is out of scope and would drop a claimed batch's turn.

## D-3 — Attach sequence, thread identity, token delivery

**Sequence.** The existing order already binds before priming, so the new step slots in after the priming turn:

1. `Start`: spawn App Server, `/readyz`, dial, `initialize`, `thread/start`, `thread/name/set` (existing).
2. Open the broker store and bind the launch-pending endpoint to the thread id (existing).
3. Driver: priming turn (existing) — this persists a rollout for the thread.
4. **New:** attach the TUI.
5. Driver loop with busy-gated claims and no stdin reader.

**Why after priming and after the bind.** (a) P8: the prior art observed that a remote TUI needs a persisted rollout before it can resume a thread; the real priming turn is already that persistence, so no extra call is needed. (b) Binding first means the TUI never takes the terminal for a session that failed to register; a bind failure or a failed priming turn is session-fatal today and ends the launcher with a plain error before any terminal takeover. (c) Typed characters during `Start` and priming wait in the tty buffer and reach the TUI once it runs, which is better than today's behavior where they are dropped until the driver's stdin reader starts.

**Thread identity.** The positional argument to `codex resume` is the thread UUID `thread/start` returned — the same string the launcher passed to `BindLaunchPending` as the endpoint's session UUID. `codex resume` accepts a session id or a session name, and UUIDs take precedence; the lane label set by `thread/name/set` is reused across runs, so the id is the unambiguous choice. The TUI child's working directory is the same directory the thread was started in (`s.dir`, or the process cwd when empty), because the lookup may be cwd-filtered (P9, not observed). The command line is:

```
codex resume --remote ws://127.0.0.1:<port> --remote-auth-token-env MOAI_FACTORY_APP_SERVER_TOKEN [operator -c/-m ...] <thread-uuid>
```

**Token.** The environment variable name is the prior art's (`MOAI_FACTORY_APP_SERVER_TOKEN`), defined as a constant in `internal/config/envkeys.go`. The value is read from the 0600 token file at attach time (`<tokenDir>/token`) rather than captured in `Start`, so this SPEC adds no line to the `Start` region that t1410 and t1459 edit. The child environment is the launcher's launch environment plus that one variable (the last occurrence wins, so a stale value inherited from the operator's shell is overridden). The App Server keeps `--ws-token-file`; its command line is unchanged.

**Rejected.** (a) `thread/inject_items` to persist a rollout without a model turn (the prior art) — the effect on `resume` in 0.160.0 is unobserved and the priming turn already does the job. (b) Attach before priming — races the TUI's resume against an unpersisted rollout. (c) Resume by name — collisions across runs. (d) Token in argv (`--remote-auth-token` style) — visible to any local process listing; the env-var-name form is the documented option. (e) A TUI child with its own token — the App Server accepts one capability token.

## D-4 — Capability probe and fallback

**Decision.** Feature-detect with `codex resume --help` and require both option names in the output, within a bounded probe timeout (`DefaultManagedCodexProbeTimeout`, working name, in `defaults.go`). The probe runs in the owner entry before `Start`, because the decision also fixes where the App Server's stderr goes (D-6). It runs only inside the existing opt-in gate (REQ-MT-001), so an ungated launch never spawns it. Measured cost on this machine: `codex resume --help` took 0.040 s wall; the bound is set orders of magnitude above that and is UNMEASURED beyond it.

Measured behaviors that shape the preconditions (P3, tree `2b9e4a4d0`, codex-cli 0.160.0, stdin `/dev/null`):

| Run | Exit | Output |
|---|---|---|
| `codex resume --remote ws://127.0.0.1:1 --remote-auth-token-env NOPE_TOKEN <uuid>`, variable unset | 1 | `Error: environment variable NOPE_TOKEN is not set` |
| same, `NOPE_TOKEN=x` set | 1 | `Error: stdin is not a terminal` |
| `codex resume --bogus-flag-xyz` | 2 | `error: unexpected argument '--bogus-flag-xyz' found` plus the usage line |

**Preconditions (REQ-MT-003).** (a) stdin is an `*os.File` that is a terminal and stdout is a terminal (the package already uses `isatty.IsTerminal(os.Stdin.Fd())`; no new dependency). (b) the probe names both options. (c) no operator opt-out. Any false precondition, or a `cmd.Start` failure on the TUI itself, falls back to headless with exactly one stderr line naming the reason, written before the terminal could be taken over.

**Why feature detection, not a version compare.** A version list goes stale and a codex fork or rename passes a version test while lacking the option; the help text is what the binary actually offers. **Why not attempt-and-fall-back on a quick exit.** Exit 2 could also be a user quitting; the terminal would already have been disturbed; and a TUI that connected and then died is indistinguishable from an operator who quit.

**The opt-out (`MOAI_FACTORY_MANAGED_TUI=0|false|off`).** Included as the escape hatch for the one thing this plan cannot observe: a real TUI (P6, P7, P9). Without it, a codex that passes the probe but cannot resume the thread would end every managed lane session at once, and the operator could not use managed mode at all until a fix lands. The opt-out restores today's headless behavior inside the existing gate; it is not a second opt-in. Open question 2 in the plan-phase report asks the operator to confirm or strike it.

**Rejected.** (a) Version probe. (b) Attempt-and-fall-back on quick exit. (c) Always-attach with no fallback (turns every non-tty launch into an error). (d) Headless fallback after the TUI has started and failed (hides the failure; the operator would see a blank terminal again).

## D-5 — Approval arguments and which requests the launcher still answers

**Approval arguments (REQ-MT-009).** `factoryMoAIMCPApprovalArgs()` stays on the App Server command line only. The TUI child gets none of the launcher-generated `-c mcp_servers.moai.*` overrides, and the project config's `default_tools_approval_mode = "writes"` is untouched. This is not a loss for TUI-originated turns: MCP tool approval is evaluated by the App Server that runs the thread, so the server-side pre-approval of the two broker tools applies to every turn on that server (inferred; not observed). The operator's own `-c`/`-m` are forwarded to the TUI because the operator typed them at a program they now see; they are the operator's explicit act, unlike the generated arguments.

**Request scoping (REQ-MT-008).** SPEC-FACTORY-MANAGED-HARDEN-001 makes the launcher's connection reader answer every server request with a refusal, because the session was unattended. With a human attached that is wrong for the human's turns: if the server delivers an approval request to the launcher's connection, the launcher would decline it before the operator could read it. The rule is therefore scoped by turn ownership, and it needs no knowledge of how the server routes (P7):

- While attached, a request that carries a `turnId` naming a turn the owner did not start is left unanswered by the launcher.
- A request with no `turnId` (four kinds, P4; an elicitation whose `turnId` is null) keeps today's answer. This is a deliberate narrow residue: an unscoped request cannot be attributed, and leaving it unanswered would hang a server that routed it only to the launcher.
- A request whose `turnId` the owner started, or that arrives while the owner's turn window is open and its turn id is not yet known, keeps today's answer (the launcher's own turns stay unattended-safe; the human may be away from the screen).
- While detached (headless, or after the TUI ended) the rule is off and HARDEN-001's behavior is exact.

"A turn the owner started" is the set of turn ids returned by the owner's own `turn/start` calls plus the priming turn, kept for the session (one short string per turn). The existing HARDEN-001 attribution (`prevTurnID`, the open window) is unchanged and sits underneath.

**Under either routing.** If the server sends a request only to the connection that started the turn, the launcher never sees operator-turn requests and the rule is idle. If it broadcasts, the rule is what keeps the launcher from declining the operator's approvals. If it routes requests of unknown provenance to the launcher alone, the narrow residue above answers them as before.

**Rejected.** (a) Keep declining everything — the operator could never approve anything in the TUI. (b) Stop answering everything while attached — launcher-injected broker turns would hang on any request with nobody at the screen, then time out session-fatal. (c) Accept the operator's requests from the launcher — no. REQ-MH-003 forbids every accepting answer. (d) Teach the launcher to detect whether a human is watching — unobservable.

## D-6 — Keeping the terminal clean

The TUI draws on the terminal; any other writer corrupts the screen. Three writers exist today (P10, measured): `managedLogf` defaulting to `os.Stderr`, the App Server child with `cmd.Stderr = os.Stderr`, and one direct `fmt.Fprintln(os.Stderr, "Factory inbox:", err)` in the driver. REQ-MT-006 moves all three to a session log file while attached.

- **File.** `.moai/logs/factory-managed-<run-id>-<label>.log` under the project root, mode 0600, append. `.moai/logs/` is gitignored (`.gitignore:399`, measured). The path is printed once on stderr before the TUI starts. The file is kept after the session for post-mortem.
- **Mechanism.** The owner entry already has the single log seam `managedLogOutput` (HARDEN-001's only test seam); production stores a synchronized writer there while attached and clears it when the TUI has ended. The driver's direct `Factory inbox:` line moves onto `managedLogf`. The App Server child's stderr is chosen before `Start`, which is why the attach decision (D-4) is made before `Start`: the `cmd.Stderr` line is the only edit in the `Start` region (D-9).
- **Why a file and not `/dev/null`.** The HARDEN-001 log lines are the only evidence of a declined MoAI elicitation, which is the silent redelivery loop that SPEC was written to expose. Discarding them would re-open that gap. They are not on the screen any more, and the documentation must say where they are (REQ-MT-013, known debt 7).
- **Writer safety.** The sink is a mutex-guarded writer (HARDEN-001's rule that the atomic pointer protects the load only).

**Rejected.** (a) Leave logs on stderr — corrupts the TUI. (b) Print them into the TUI — no channel. (c) `/dev/null`. (d) A new log subsystem — one file and the existing seam suffice.

## D-7 — Busy tracking and the connection reader

**Problem.** The reader pushes every `turn/started` and `turn/completed` to a 32-slot event channel that only `call` and `waitTurn` drain. In headless mode that is fine, because the launcher is the only turn source. With a TUI, every operator turn emits a pair of frames with no consumer; after about 16 operator turns between launcher turns the reader blocks on the channel, stops reading, stops answering server requests, and the next launcher turn times out session-fatal. This is a defect of the attach, not of HARDEN-001.

**Decision.**
- The reader keeps a set of active turn ids from `turn/started` and `turn/completed` of every turn, and exposes `busy := len(active) > 0`. Frames of turns the owner did not start never enter the event channel; frames of the owner's own armed window do, exactly as now.
- A stale-busy ceiling: when the thread has been continuously busy for longer than `DefaultManagedCodexTurnTimeout` (the existing 10-minute bound), busy reads as false and one log line is written, so a lost `turn/completed` frame cannot starve delivery. No new constant.
- The driver reaches `busy` through the optional capability interface of D-8 and skips its claim step while busy. Claims therefore happen only at idle and are delivered at once, which also keeps the 2-minute claim lease from expiring while a batch waits.
- A deferred batch is not claimed, so nothing is lost; the next idle tick claims it.

**Premise.** P6: this works only if the launcher's connection receives the TUI-started turns' frames. If it does not, `busy` stays false and delivery follows P5 (steer into the active turn) — degraded, not deadlocked, and the manual check observes which case holds (known debt 3). `thread/status/changed` (idle/active) exists in the 0.160.0 notification schema and would be a second signal; it is not used because it adds a second unobserved premise.

**Rejected.** (a) Poll `thread/read` for status — a new RPC and a new failure mode. (b) Let the event channel grow unbounded — hides a leak. (c) Gate on the TUI process state — says nothing about turns.

## D-8 — Lifecycle, exit status, teardown order

**Optional capability interface.** The driver finds the new behavior through a second interface the Codex owner implements and the stream owners do not (working name `managedOperatorSurface`: `AttachOperator` and `Busy`). The `managedSession` interface and the driver signature do not change, which also keeps t1459's refusal of an interface change intact. After the priming turn the driver asks for the surface; if it attaches, the driver does not start the stdin reader, adds the TUI-exit channel to its select, and gates claims on `Busy`. If it does not (every stream owner, every headless Codex), the driver is exactly today's.

**End paths.**

| Path | What happens | Result |
|---|---|---|
| TUI exits on its own | the wait goroutine records the status; the driver returns; deferred `Close` runs | launcher exit 0 for status 0, the TUI's code for 1..255, 1 when a signal ended it |
| App Server connection closes (reader exits) | the monitor stops the TUI; the driver returns the connection error | launcher exit 1, error names the closed connection |
| Driver ends the session (3 consecutive turn failures, a session-fatal error) | deferred `Close` stops the TUI first | the driver's own error stands |

**Precedence.** The TUI's exit status is reported only if the TUI ended before the owner asked it to stop. An owner-initiated stop must not turn into the TUI's interrupt-induced status. A `DeliverTurn` blocked inside the owner when the TUI exits is released by a TUI-exit channel the client selects on (a small addition next to `done`), and the driver then returns the TUI's result; the owner never calls `Close` from a second goroutine, which stays unsafe until t1459 makes it concurrency-safe.

**Teardown order (one fixed order).** (1) stop the TUI: `os.Interrupt` to the process, wait up to `DefaultManagedCodexTUIStopGrace` (working name, `defaults.go`, UNMEASURED), then `Kill`, then `Wait`; on Windows `Signal(os.Interrupt)` is unsupported and the owner goes straight to `Kill`; (2) shut the WS client down; (3) kill and wait for the App Server; (4) remove the token directory; (5) clear the log sink. The TUI goes first so it does not render a dying server and so the terminal returns before the launcher exits. `stopTUI` is idempotent behind its own small mutex, because the monitor goroutine and `Close` can both reach it.

**Why interrupt before kill.** A killed TUI cannot restore the terminal; an interrupted one may. Whether it does is not observed (known debt 5).

**Exit-status mechanics.** The package already carries `exitCodeError` (an `ExitCode()` the CLI entry maps). The TUI status becomes that error when nonzero; zero returns nil. Go's `ProcessState.ExitCode()` is -1 for a signaled process, which maps to 1; no `syscall` is needed.

**Rejected.** (a) Add `Context` to `DeliverTurn` or change `managedSession` — t1459 rejected the same change on a measured blast radius (`DeliverTurn(` at 13 sites in 6 files). (b) Kill the TUI first and ignore grace — terminal damage. (c) Let the launcher outlive the TUI as a headless owner — a lane session that keeps running with no screen after the operator quit is the original problem.

## D-9 — Composition with t1459 and t1410, and the merge-order interaction

Measured surfaces (tree `2b9e4a4d0`; t1459's draft read from `git show 2e41b007c:.moai/specs/SPEC-FACTORY-MANAGED-SIGNAL-001/...`, not in this tree):

| This SPEC edits | File | Overlaps |
|---|---|---|
| New TUI code (attach, stop, probe, log sink, optional interface impl) | new `internal/cli/managed_codex_tui.go` | none |
| One struct field on `managedCodexSession` | `managed_codex_factory.go` struct | t1459 adds lock/closed fields in the same struct |
| One line `s.cmd.Stderr = os.Stderr` → the session's chosen writer | `Start`, line 689 | t1410 F8 edits the `cmd.Start` failure branch at lines 690-696 directly below; t1459 rewrites `Start` |
| One call `s.stopTUI()` as the first step of `Close` | `Close` | t1459 rewrites `Close` (its O17–O19) |
| Reader changes (busy set, drop non-owned lifecycle frames, request scoping) | `read()` and the turn-window helpers | none in t1459/t1410 (HARDEN-001 owns them and is landed) |
| Driver: optional-interface block, `Factory inbox:` line to `managedLogf` | `managed_factory_session.go` | none |
| Constants | `defaults.go` after line 130; `envkeys.go` | t1410 F9 changes the value at `:117`; adding after `:130` avoids the text conflict |
| Tests | new `managed_codex_tui_test.go` with its own fake codex role | t1410 shares `serveFakeRPC`; this SPEC does not edit it |

**t1459 (F5, held).** t1459's draft treats the session as two published resources (child process, token directory) behind a lock `L`, with a fixed teardown table (O19). The TUI is a third. Recommended landing order is **t1408 first**: t1459 is on hold, and its plan then gains two deltas — O19 gets "stop TUI" as its first step, and the signal-delivered rule R-E must win over the TUI exit status (a signal that tore the TUI down reports the interruption, not the TUI's interrupt-induced status). If t1459 lands first, this SPEC's run phase rebases onto its `Close`: `stopTUI` becomes the first step inside the O19 body, and `AttachOperator` publishes the TUI field under `L` at attach time. Either order works because the TUI code is isolated in one file and meets `Close` at one line. This is design intent, not a verified merge; the run-phase merge will re-measure the three overlap lines. The open question for the leader is whether to confirm t1408-first.

**t1410 (F8/F9/F13, queued).** F8 edits token-directory cleanup on `Start` failure paths; the attach happens after a successful `Start` and reads the token file only then, so the two do not interact. F13 touches `managedCodexAppReady`, which this SPEC does not touch. F9 changes `DefaultManagedCodexReadyTimeout`'s value; the dial `HandshakeTimeout` still uses it and the probe has its own constant. The shared fake in `managed_codex_factory_test.go` is why this SPEC puts its fake codex in a new file.

**t1440 (lane-17).** It edits `codex_launcher.go` and creates `managed_operator_input.go` and `managed_card_child_test.go`; it leaves `managed_codex_factory.go` and `managed_factory_session.go` alone. This SPEC touches only those two files plus new files (`managed_codex_tui.go`, `managed_codex_tui_test.go`) and the config constants, and plans nothing in `codex_launcher.go`. The doc bullets in `.moai/docs/factory-managed-session.md` are a shared surface (minimal separate bullet edit).

**Do not absorb.** This SPEC does not add signal handling, restructure `Start`/`Close`, change token-directory cleanup, change the handshake budget or add the readyz redirect check.

## D-10 — Portability and test seams

- **Zero `syscall`.** The new file uses `os/exec`, `os.Interrupt`, `os.Process`, `isatty` (already in `go.mod` and used in `internal/cli`). No `syscall.` text appears in any `managed_*` file, test files included (the parent grep covers `managed_*.go` and so `managed_*_test.go`). Test fixtures check process liveness through the owner's recorded wait result or `exec.Command("kill","-0",…)`, not `syscall.Signal(0)`.
- **Windows.** `Signal(os.Interrupt)` returns an error on Windows; the stop goes straight to `Kill`. The re-exec fixtures are POSIX-shell shims and skip on Windows, as the existing managed tests do; the cross build is the Windows proof.
- **Seams (package-private, production defaults real).** (1) terminal predicate `func(*os.File) bool`, default `isatty`; (2) the TUI's stdout/stderr destinations as session fields, default the process streams; (3) the existing `managedLogOutput`. The fake codex is a re-exec of the test binary, one helper with an App Server role and a TUI role, so the "program" the owner launches for both `app-server` and `resume` is the same shim, as in production where both are the one codex binary. The helper's log is a single append-only file, so cross-process ordering assertions read one sequence.
- **Hardcoding.** `MOAI_FACTORY_APP_SERVER_TOKEN` and `MOAI_FACTORY_MANAGED_TUI` are `envkeys.go` constants; the probe timeout and stop grace are `defaults.go` constants with UNMEASURED comments, like the sibling `DefaultManaged*` values.

## Premises that rest on prior art or inference

Each is labeled where it is used; none is observed. P5, P6, P7 and P9 are the ones whose failure would change a decision: P5/P6 decide how delivery behaves while an operator turn runs (D-7), P7 decides how often the narrow request residue of D-5 matters, and P9 decides whether D-3's identity and D-8's stop behave on a real terminal. AC-MT-015 is the single place a real run answers them.
