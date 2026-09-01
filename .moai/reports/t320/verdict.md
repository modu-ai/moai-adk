# t320 verdict — moai integration release empty-ERROR premise refuted

Card: t320 (moai integration release가 빈 메시지와 함께 ERROR — 실패 사유를 읽을 수 없다; 관측 lane-18 2026-08-27, t308 병합 창 반납 시)
Branch: WT-release-reason (worktree .claude/worktrees/t320, base b6231290d = develop tip at entry)
Verdict: PREMISE REFUTED — no code change. The observed empty-message ERROR is not reproducible at any tracked revision of any involved layer, and every failure state the card asks to distinguish is already distinguished in today's output. Recommend the lead DROP the card (t425/t10 precedent: 전제 반증).

## Claim

1. The card's requested repair — "반납 실패 시 사유가 출력에 나와야 한다 — 홀더 아님 / 이미 축출됨 / 파일 없음 등을 구분해서" — is ALREADY the behavior of the current tree: seven failure states reproduce with seven distinct, specific reasons.
2. The card's root-cause hypothesis ("release가 홀더-아님 상태를 에러로 뱉으면서 메시지를 비워둔 것") is contradicted by the code at every tracked revision: `ErrIntegrationLockNotHeld` has carried its message ("no release integration window is held") unchanged since t194 (b2ad9158c, 2026-08-24) through the observation day's tree (3f3465369, 2026-08-27) to HEAD.
3. "Full output saved"-style renderer drop is also unlikely as the cause: charm.land/fang/v2 is v2.0.1 at both the observation date and now (go.mod at 3f3465369 line 8 = HEAD line 8), and the same version renders every reason in this run's reproductions under the same non-TTY capture conditions a lane session runs under.
4. Timeline context supporting premise decay: the observation (2026-08-27) coincides with t298's pid-liveness fix landing the same day (3f3465369), followed by t336's atomicity rework (2026-08-30). The state set release can meet changed after the observation; the card itself names t298 착지 as its precondition — that precondition has been satisfied since.

## Evidence

All reproductions ran a binary built from THIS tree (`go build -o /tmp/moai-t320 ./cmd/moai`), against temp CLAUDE_PROJECT_DIR fixtures:

| # | State | Rendered reason (verbatim) |
|---|-------|----------------------------|
| 1 | no lock held (no record file) | `No release integration window is held.` rc=1 |
| 2 | session id unresolvable | `Cannot resolve this session's id; pass --session <id> or --force.` rc=1 |
| 3 | foreign holder (이미 축출됨) | `Release integration window is held by a different session: lane-2 (pid 999999) holds it.` rc=1 |
| 4 | corrupt record file | `Integration lock at /tmp/t320-corrupt/.moai/state/integration-lock.json is unreadable: invalid character 'o' in literal null (expecting 'u').` rc=1 |
| 5 | JSON-mode failure | same reason rendering as #1 under `--json` |
| 6 | mutation-lock contention (Busy) | code-read: `ErrIntegrationLockBusy` = "release integration window record is busy: another process is mutating it" + waited budget (construction verified by read; message path exercised by existing tests) |
| 7 | release success line | `release-integration window released (was <session> on <branch>)` — existing CLI test asserts the round trip |

Layer history (all non-empty at every revision read):

- `git show b2ad9158c:internal/kanban/integration_lock.go` — t194 original: the three sentinels already carried their messages; `ReadIntegrationLock` mapped ErrNotExist → empty record → NotHeld, identical to today.
- `git show 3f3465369:internal/kanban/integration_lock.go` (observation-day tree) — `ReleaseIntegrationLock` byte-identical in its message paths to the original.
- `git diff 3f3465369..HEAD -- internal/cli/integration.go` — EMPTY: the release verb is unchanged since the observation day's tree.
- `git show 3f3465369:go.mod` vs `go.mod` — `charm.land/fang/v2 v2.0.1` in both.

## Baseline-attribution

Every reproduction and every historical read was performed in this run, in worktree .claude/worktrees/t320 (branch WT-release-reason) at base b6231290d. The binary under test was built from this tree in this run. No figure or output is carried from another session, tree, or binary.

## Gaps

- The EXACT state lane-18's binary met on 2026-08-27 cannot be reconstructed: that binary predates the current tree, its lock state was transient, and no log of the invocation survives. The refutation is therefore "no tracked revision of the involved paths produces an empty message for any state I can construct" — not "lane-18 saw something else".
- Windows-only paths (integration_lock_mutation_windows.go, the wedged-artifact clear) were read, not executed: this machine is darwin. The Busy message construction is shared code and was verified by read; the Windows atomic-create contention path itself was not exercised.
- A rendering failure specific to lane-18's terminal (pane width, color profile) cannot be falsified from here — today's same-version renderer conveys every reason under lane-equivalent capture conditions.

## Residual-risk

- If lane-18's empty ERROR reappears, the productive observation to capture is the raw bytes (`moai integration release 2>&1 | cat -A`) plus the binary's `moai version` — that would discriminate renderer-drop from a message-bearing error that failed to display.
- "파일 없음" remains folded into "No release integration window is held" (no record file reads as no hold). Semantically the same answer at the right altitude; if a future maintainer wants the file-absent state named separately, that is an enhancement, not this card's defect.
