# SPEC Review Report: SPEC-INTEGRATION-LOCK-LIVENESS-001 (card t298)

Iteration: 2/2 (Tier M ceiling — see § Audit-integrity note)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.92** (harmonic mean; Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. Every premise below was re-measured in this
run against this tree.

## Attribution

| Item | Value |
|---|---|
| Worktree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t298` (`git rev-parse --show-toplevel`) |
| Branch / HEAD | `WT-integration-lock` / `c67a6ea64` |
| Base | `afde6ebb3` is an ancestor (`git merge-base --is-ancestor` → exit 0) |
| Verdict pinned to | `spec.md` `0b1e2a17784da00cdd20099b52d3a83c760cee00`, `acceptance.md` `ecf4fb692582ebf05ab17869e6cc271f434314db`, `plan.md` `5fafad9d825bf33322802f6ab851c122e50d6d0c`, `progress.md` `191c2081de8398c264a7c04136982ebcc294e4dc` (2026-08-27T10:49:52Z) |
| SPEC version audited | **v0.1.2** |

**This verdict is void if any of the four hashes above differs at the time it is consumed.**

## Audit-integrity note (read first)

Two premises of the delegating instruction were false, and both bear on how this verdict is used.

1. **"No prior audit exists — this is iter-1" is false.** `.moai/reports/t298/plan-audit-iter1.md`
   already existed at audit start (17,711 bytes), carrying `Verdict: FAIL / Overall Score: 0.795`
   and defects D1-D9. This report is therefore **iteration 2**, and the Tier M ceiling
   (`plan_audit_tier_ceilings`, M=2) is now **consumed**. No further plan-phase audit iteration
   is available for this SPEC.
2. **The audit subject was rewritten four times during the audit window.** Artifact hashes moved
   under measurement at ~19:45, ~19:46, ~19:48 and ~19:49 local, and the SPEC version advanced
   `0.1.0 → 0.1.1 → 0.1.2` mid-read. `agent-common-protocol.md` § Background Agent Execution
   binds an actively audited worktree to exactly one writer for the window between the opening
   measurement and the landed verdict; observing a foreign write is named there as a process
   defect to report rather than absorb. Reported here as **D5**. Practical consequence: my
   opening reads were of v0.1.0; every finding below was re-confirmed against the pinned v0.1.2
   hashes before being written down, and findings that v0.1.1/v0.1.2 repaired were withdrawn.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-INL-[0-9]*' spec.md | sort -u` →
  `REQ-INL-001 … REQ-INL-010`, sequential, no gaps, no duplicates, uniform 3-digit padding.
  `grep -c '^- \*\*REQ-INL-' spec.md` → `10`. (The bare `REQ-INL-` token in the same sort is the
  numbering-convention sentence `REQ numbering: REQ-INL-NNN`, spec.md §C, not an entry.)
- **[PASS] MP-2 GEARS format compliance — judged against the REQUIREMENT layer (`REQ-XXX` in
  spec.md) only.** All ten entries match a GEARS pattern: Ubiquitous (001); Event-driven "When"
  (002, 003, 005, 009); State-driven "While" (004); Where / capability-gate (006, 007, 008);
  Unwanted "shall not" (010). No IF/THEN modality anywhere in §C. The Given-When-Then entries in
  `acceptance.md` are the **verification layer** and were graded under Group 4, never here.
  (v0.1.1 repaired the non-canonical `(Event-detected)` label on REQ-INL-003 → `(Event-driven)`.)
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with canonical
  names (`created` / `updated` / `tags`; no `created_at` / `updated_at` / `labels` / `spec_id`
  aliases), plus optional `tier: M`. `version: "0.1.2"` quoted semver; `status: draft` ∈ enum;
  `priority: P1` ∈ enum; `lifecycle: spec-anchored` ∈ enum; `phase: "v3.1.4 target"` is a release
  label, not one of the prohibited lifecycle stages (`plan`/`run`/`sync`/`mx`). See D1 for a
  `module:` **completeness** defect that does not breach the schema (the field is non-empty and
  path-like, so MP-3 passes). Tool-verified rather than read-only:
  `CLAUDE_PROJECT_DIR=$PWD moai spec lint` → `0 error(s), 64 warning(s)`, exit 0, and
  `grep -i 'INTEGRATION-LOCK-LIVENESS' <output>` → no match (`rc=1`) — **zero findings name this
  SPEC**. (Run against the binary built from this tree at `c67a6ea64`.)
- **[N/A] MP-4 Section 22 language neutrality** — single-language (Go) scope; no template-bound
  or multi-language-tooling content. Both documentation targets (`CLAUDE.local.md`,
  `.claude/rules/local/gitflow-lane-protocol.md`) are local-only with no template mirror, stated
  in spec.md §D and plan.md M5.
- **[N/A] MP-5 D7 cross-SPEC reconciliation** — verb executed:
  `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → a single hit,
  `SPEC-INTEGRATION-LOCK-LIVENESS-001` (self). No foreign SPEC is referenced, so there is no
  status to reconcile and no BLOCKING finding is emitted. Predecessors are named as **cards**
  (t194, t181), not SPEC IDs, and spec.md §F states that explicitly.
- **[PASS] MP-6 D8 cross-platform discipline** — verb executed:
  `grep -rn 'syscall' .moai/specs/SPEC-INTEGRATION-LOCK-LIVENESS-001/` → no match (`rc=1`).
  D8-4 auto-PASS. Windows parity is nonetheless handled explicitly (REQ-INL-007, AC-INL-009,
  `factory_alive_windows.go` untouched).
- **[PASS] MP-7 clarification gate** — verb executed:
  `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-INTEGRATION-LOCK-LIVENESS-001/` → no match
  (`rc=1`). `research.md` does not exist (correct for Tier M).

No must-pass failure. No D7 or D8 BLOCKING finding.

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.95 | 1.0-band, minus one | Every REQ has a single reading; no weasel words (`awk '/^## §C Requirements/,/^## §D/' spec.md \| grep -e ' should ' -e ' may ' -e 'reasonable' -e 'appropriate'` → none). §A now separates **identity-of-record** from **atomicity** explicitly (spec.md:L84-92), which is the distinction a reader would otherwise conflate. Deduction: `module:` (spec.md:L11) contradicts the plan's own touch set — D1. |
| Completeness | 0.90 | 1.0-band, minus one | HISTORY (§H, 3 rows), WHY (§A), WHAT (§C), HOW (plan.md §B), REQUIREMENTS (§C), ACCEPTANCE (acceptance.md, 12 rows), and four `### Out of Scope — <topic>` H3 sub-headings each with `-` bullets (`grep -c '^### Out of Scope — ' spec.md` → 4). §G now carries six risks incl. the read-modify-write hazard, PID reuse, and the absent host field. Deductions: D1 and D4. |
| Testability | 0.85 | 0.75/1.0 boundary | Strong: cross-process RED via a real built binary; on-disk-record assertions (AC-INL-001 clause (b), AC-INL-012 clause (a)) rather than printed-string assertions; explicit mutant-guard pairing (001↔003, 004); a measured coverage-absence probe (`grep -rln "exec.Command" …` → `rc=1`). Deductions: the M5 `[HARD]` wording bound has **no** acceptance criterion (D2), and AC-INL-010's green cell is prose-judged where its sibling AC-INL-011 names exact greps (D3). |
| Traceability | 1.00 | 1.0-band | All 10 REQs covered by ≥1 AC; all 12 ACs reference REQs that exist; §D matrix uses full REQ IDs (v0.1.1 repaired the abbreviated `REQ-INL-004, 008` form); every AC row names its flipping milestone and every milestone in plan.md §F names the ACs it flips; AC-INL-012 is traced to existing REQ-INL-004 + REQ-INL-010 with **no new REQ invented**. Budgets: 10 REQ ≤ 16, 12 AC ≤ 16 (Tier M ceilings). |

Harmonic mean = `4 / (1/0.95 + 1/0.90 + 1/0.85 + 1/1.00)` = **0.9216** → 0.92.

## §1 — Premise re-verification table (spec.md §B, re-measured on `c67a6ea64`)

Context first, because it decides the whole table: `internal/kanban` is **byte-unchanged**
between the pinned baseline and this tree.

```
$ git merge-base --is-ancestor d29b8942e HEAD && echo ancestor
ancestor
$ git diff --name-only d29b8942e..HEAD -- internal/kanban internal/cli internal/hook internal/session
internal/cli/codex_auth_ladder_test.go
internal/cli/codex_launcher.go
internal/cli/codex_launcher_cross_test.go
internal/cli/codex_launcher_guards_test.go
internal/cli/codex_launcher_readout_test.go
internal/cli/codex_readiness.go
internal/cli/codex_readiness_test.go
internal/cli/help.go
internal/cli/mcp_codex.go
internal/cli/mcp_codex_test.go
internal/cli/spawn.go
internal/cli/testdata/…
```

Every changed path is codex-launcher work in `internal/cli`; **no premise file appears**.
`internal/kanban`, `internal/hook`, `internal/session`, `internal/cli/integration.go`,
`internal/cli/integration_lock_cli_test.go`, `CLAUDE.local.md`, and
`.claude/rules/local/gitflow-lane-protocol.md` are all absent from the list — so the §B line pins
are measured on an unchanged surface. Each was still re-measured individually:

| # | Premise (as stated) | Verdict on `c67a6ea64` | Measurement |
|---|---|---|---|
| 1 | `integration_lock.go:71-79` — `IntegrationLock{SessionID, SessionName, PID, Branch, Worktree, AcquiredAt, Card}` | **HOLDS** | `grep -n "type IntegrationLock struct" -A 10` → struct opens at `:71`, closes `:79`, exactly those 7 fields |
| 2 | `integration_lock.go:174-176` — `if want.PID == 0 { want.PID = os.Getpid() }` is the defect | **HOLDS** | `grep -n "os.Getpid()" internal/kanban/integration_lock.go` → single hit at `:175`; `if` at `:174`, `}` at `:176` |
| 3 | `integration_lock.go:91-99` — `Stale()`: `Held()` gate → `PID <= 0` false → `!FactoryProcessAlive` | **HOLDS** | function spans `:91`-`:99` verbatim, `PID <= 0` at `:95`, probe at `:98` |
| 4 | `integration_lock.go:84-90` — the asymmetry doctrine comment | **HOLDS** | `grep -n "asymmetry is not close"` → `:90`; comment block starts `:84` |
| 5 | `integration.go:80-92` — `integrationSessionID` chain (flag → `CLAUDE_CODE_SESSION_ID` → side-channel file) | **HOLDS** | function spans `:80`-`:92`; env read `:84`, file read `:87` |
| 6 | `integration.go:148-152` — status prints "held by a session that is gone (reclaimable)" | **HOLDS** | `grep -n "reclaimable" internal/cli/integration.go` → `:149`, inside the `if lock.Stale()` at `:148`; the `Fprintf` runs `:151`-`:152` |
| 7 | Non-test consumers: `integration.go:131,138,148` + `integration_lock_guard.go:74,95`; guard stale branch allows with advisory at `:95-97` | **HOLDS** | `grep -rn -e ReadIntegrationLock -e 'Stale()' internal/ --include "*.go" \| grep -v _test.go` → exactly those five consumer lines (plus definitions in `integration_lock.go` and unrelated `Stale` identifiers in `internal/statusline`, `internal/mx`, which are different symbols); guard advisory verified at `:95-97` |
| 8 | `factory_alive_unix.go` + `factory_alive_windows.go` both exist; Windows parity to preserve | **HOLDS** | `ls internal/kanban/factory_alive_*.go` → both present |
| 9 | `registry.go` — `DefaultRegistryPath` `:35-39`, `DefaultStaleMinutes = 30` `:62-65`, recorded PID is `resolveSessionPID()` with the "NOT os.Getpid()" comment | **HOLDS** | `grep -n` → `:39` const path, `:65` `= 30`, `:191` comment "The session PID, NOT os.Getpid()", `:195` `PID: resolveSessionPID()` |
| 10 | `session_pid.go` — chain `MOAI_SESSION_PID`(live) → non-wrapper ancestor → `os.Getpid()` fallback; wrapper set `sh/bash/zsh/dash/ksh/fish/csh/tcsh/env/moai` | **HOLDS** | `resolveSessionPID()` at `:71`-`:79` is literally those three steps; `wrapperProcessNames` at `:40`-`:44` is exactly that set |
| 11 | `launch_session_pid.go` — `withSessionPID` stamps on POSIX; doc constraint "A hook must never set it" | **HOLDS** | `grep -n` → constraint at `:13`, `withSessionPID` at `:32` |
| 12 | `integration_lock_cli_test.go:26-43` — `runIntegration` drives cobra **in-process** with `CLAUDE_PROJECT_DIR` + `GIT_CEILING_DIRECTORIES` pinned | **HOLDS** | function spans `:26`-`:43`; `t.Setenv` pins at `:28`-`:29`; `cmd.Execute()` in-process at `:41` |
| 13 (v0.1.2 form) | `gitflow-lane-protocol.md` §3 **enshrines** the defect (blockquote L42-43); `CLAUDE.local.md` §4.1/§4.1.2 is **silent** on it | **HOLDS** | `grep -n '직렬화를 보장하지 못한다' .claude/rules/local/gitflow-lane-protocol.md` → `:42`; the workaround line "직렬화는 리드 공지로 한다" → `:43`. `grep -n '직렬화' CLAUDE.local.md` → no match; `§4.1.1` at `:291`, `§4.1.2` at `:302` is a bash block whose only lock lines are `:312` `moai integration acquire --name <lane>` and `:317` `moai integration release` |
| 13 (v0.1.0 form) | *"`CLAUDE.local.md` §4.1/§4.1.2 … enshrine the defect as a workaround"* | **WAS FALSE — repaired in v0.1.1** | Independently measured before reading the prior report; the file carries no workaround prose. v0.1.1 restated item 13 and AC-INL-011 against the measurement. Recorded here as convergent confirmation, not as an open defect |

**13/13 premises hold as stated in v0.1.2.** One premise (item 13, v0.1.0 form) was false and was
already repaired by the author before this verdict; my independent measurement agrees with the
repair.

## §2 — Root-cause soundness

**The card's hypothesis is correct, and I reproduced it directly rather than inferring it.**

The recorded identity does come from the acquiring CLI process: `newIntegrationAcquireCmd`
(`internal/cli/integration.go:178-183`) constructs `kanban.IntegrationLock{SessionID, SessionName,
Branch, Worktree, Card}` and **never sets `PID`**, so `AcquireIntegrationLock` fills the zero value
with `os.Getpid()` at `integration_lock.go:175` — the pid of the one-shot `moai` process.

Reproduction on this tree, with `CLAUDE_PROJECT_DIR` pinned to a scratch root (the real lock was
never touched — see §6):

```
$ go build -o "$SP/moai-t298" ./cmd/moai          # rc=0
$ export CLAUDE_PROJECT_DIR="$SP/red/root1" GIT_CEILING_DIRECTORIES="$SP/red"
$ echo "parent shell pid=$$"
parent shell pid=6740
$ "$SP/moai-t298" integration acquire --session sess-audit-1 --name probe-audit
release-integration window acquired by sess-audit-1 on WT-integration-lock
$ "$SP/moai-t298" integration status
release-integration window: held by a session that is gone (reclaimable)
  holder:   sess-audit-1 (pid 6747)
  branch:   WT-integration-lock
  ...
$ cat "$CLAUDE_PROJECT_DIR/.moai/state/integration-lock.json"
{ "session_id": "sess-audit-1", "session_name": "probe-audit", "pid": 6747, ... }
```

The recorded pid `6747` is neither the still-live parent shell (`6740`) nor any surviving process —
it is the exited acquire child. **AC-INL-001's RED cell is confirmed red, for the stated reason.**

AC-INL-002's RED is likewise confirmed: with the env stamp present, the recorded pid is still the
child's, so the stamp is genuinely ignored today:

```
$ echo "parent shell pid=$$"
parent shell pid=10434
$ MOAI_SESSION_PID=$$ "$SP/moai-t298" integration acquire --session sess-audit-2 --name probe-env
$ "$SP/moai-t298" integration status --json
{"held":true,"lock":{"session_id":"sess-audit-2",...,"pid":10441,...},"stale":true}
```

`10441 ≠ 10434` — the stamp naming the live parent was discarded.

And the downstream harm, reproduced end to end:

```
$ "$SP/moai-t298" integration acquire --session sess-foreign --name probe-foreign
release-integration window acquired by sess-foreign on WT-integration-lock
  displaced: sess-audit-2 (pid 10441), held since 2026-08-27T10:45:10Z
foreign-acquire rc=0
```

A bare foreign acquire — **no `--force`** — took the window and reported a displacement. This is
exactly the field harm spec.md §A observation 1 records, and exactly what AC-INL-012 now gates.

**Implementability of "record the session's identity instead":** yes. The CLI already imports both
`internal/kanban` and `internal/session` (`integration.go:25-27`), so plan.md §B consequence 2's
placement — resolve in the CLI, keep `internal/kanban` import-pure — is available with no new
dependency. The cycle risk is measured absent: `grep -rn '"github.com/modu-ai/moai-adk/internal/kanban"'
internal/session/` → 0 hits (`rc=1`); the single unscoped textual hit is a comment at
`session_pid.go:49`. Dropping `resolveSessionPID`'s step-3 `os.Getpid()` fallback for the lock's
seam (plan.md §B consequence 2) is correct and necessary — inheriting it would reintroduce the
defect under a new name.

**Independent second cause — found, and the SPEC now carries it.** `AcquireIntegrationLock`
(`:146`-`:181`) performs `ReadIntegrationLock` (`:155`) → decide (`:159`-`:169`) →
`writeIntegrationLock` (`:177`) with no interposed lock, and the write stages through a **fixed**
`path + ".tmp"` (`:229`) shared by every concurrent writer. `grep -n -e Flock -e flock -e LockFile
internal/kanban/integration_lock.go internal/cli/integration.go` returns 5 lines and **not one call
site**: `:15` and `:19` are `flock` prose in the package header, `:38`/`:39`/`:106` match only the
`IntegrationLockFileName` substring. So two simultaneous acquires can both read "free" and both
write — last write wins, silently, in the false-"live-for-me" direction the §D asymmetry names as
the catastrophic one. **This is correctly diagnosed and correctly scoped out in v0.1.1/v0.1.2**:
spec.md §G records it as a RETAINED residual hazard with the measurement inline, §A separates it
from the identity-of-record layer, and plan.md M5 bounds the doc wording so the fix is not
described as mutual exclusion. That disposition is right — widening this card to cover atomicity
would be scope creep. But see **D2**: the bound has no verification instrument.

Third cause considered and NOT found blocking: `integrationSessionID`'s side-channel fallback
(`integration.go:87`) reads a file that, by the code's own comment (`:71`-`:74`), "names whichever
session started last", so a lock can be recorded under a foreign session id — which makes
`SessionID`, the field REQ-INL-004 and the guard both depend on, independently stale-able. This is
pre-existing, out of scope, and acknowledged in code — but it is absent from §G. Recorded as **D6**
(optional).

Staleness-threshold / clock cause: **correctly ruled out.** `Stale()` uses a pid probe, not a
window, so there is no clock dependency and no TOCTOU beyond the write race above. plan.md §B
Option A rejects the registry's 30-minute `DefaultStaleMinutes` as an oracle for exactly the right
reason (it produces false-"stale" on a quiet-but-alive lane), and `DefaultStaleMinutes = 30` is
verified at `registry.go:65`.

File-format migration cause: covered by REQ-INL-006 + AC-INL-007, and cheap by construction —
`Stale()` needs no marker-conditional branch, because `PID <= 0 → live` and `PID > 0 → probe`
already yield the correct answer for both record shapes (v0.1.1 restated REQ-INL-006 to say this
explicitly rather than implying a second evaluation path).

## §3 — REQ ↔ AC ↔ milestone traceability

`grep -o 'REQ-INL-[0-9]*' spec.md | sort -u` → 001-010. `grep -c '^### AC-INL-' acceptance.md` → 12.

| REQ | Covered by | Mechanically checkable? |
|---|---|---|
| REQ-INL-001 | AC-INL-001 | yes — `status --json` `stale` field + on-disk record |
| REQ-INL-002 | AC-INL-001(b), AC-INL-002 | yes — recorded pid equality + `pid_source` marker read back from disk |
| REQ-INL-003 | AC-INL-004 | yes — `Stale()` false on a pid-0 anchored record |
| REQ-INL-004 | AC-INL-001 (status leg), AC-INL-008 (guard leg), AC-INL-012 (acquire-refusal leg) | yes — all three consumer legs now observed |
| REQ-INL-005 | AC-INL-003 | yes — exit status of a bare foreign acquire |
| REQ-INL-006 | AC-INL-007 | yes — hand-written legacy record parses + reads reclaimable |
| REQ-INL-007 | AC-INL-009 | partly — the compile leg is `GOOS=windows … go vet` (mechanical); the behavioral leg is proxied through AC-INL-002/004 and stated as such |
| REQ-INL-008 | AC-INL-008 | yes — `INTEGRATION_LOCK_VIOLATION` sentinel / advisory / conservative-deny, three legs one run |
| REQ-INL-009 | AC-INL-010, AC-INL-011 | **partly — D2/D3** |
| REQ-INL-010 | AC-INL-005, AC-INL-006, AC-INL-012(b) | yes — foreign-release sentinel; `displaced:` line present on takeover, absent on refusal |

**No uncovered REQ. No orphaned AC.** Every AC-INL-0NN row names a flipping milestone, and every
milestone in plan.md §F names the ACs it flips — including M1's three-test list and M2's
"M1's three tests flip green (AC-INL-001, AC-INL-002, AC-INL-012)" exit line, so the cross-layer
sweep for the v0.1.2 addition was actually performed (`grep -n "AC-INL-012" plan.md` → `:127`,
`:155`, `:179`).

ACs requiring prose judgment: **AC-INL-010 only** ("the replacement wording is present") — D3.

## §4 — Failing-test-first discipline

M1 is genuinely the failing-test-first milestone and its RED is genuinely red on this tree.

- **The command that demonstrates it:** `go test ./internal/cli/ -run 'TestIntegrationOwnerLiveness' -v`
  (plan.md §F M1 verification line), expected to fail on all three tests before M2.
- **Independent confirmation that it will fail** — I reproduced each test's assertion by hand with
  the real built binary (transcripts in §2 above): AC-INL-001 shape yields `stale: true` where
  `stale == false` is asserted; AC-INL-002 shape records `10441` where the parent's `10434` is
  asserted; AC-INL-012 shape yields a successful bare acquire reporting `displaced: sess-audit-2`
  where a non-zero exit and an untransferred record are asserted. All three assertions fail today,
  each for the reason its RED cell states.
- **The green path names the flipping milestone** in every case: M2 (`owner-pid anchor`), and M2's
  own exit line closes the loop by naming the three tests.
- **The harness constraint is right and non-obvious:** `runIntegration`
  (`integration_lock_cli_test.go:26-43`) executes cobra **in-process**, so the recorded pid is the
  live `go test` process — which is why the pre-existing refusal tests
  (`TestIntegrationAcquire_RefusesASecondLane` `:106`; `TestAcquireIntegrationLock_RefusesASecondLiveSession`
  `internal/kanban/integration_lock_test.go:60`) are green **vacuously** with respect to the field
  shape. The SPEC measures that absence rather than asserting it:
  `grep -rln "exec.Command" internal/cli/integration_lock_cli_test.go internal/kanban/integration_lock_test.go`
  → no output, `rc=1` (re-run by me; confirmed). The `go build -o` / not-`go run` constraint
  (plan.md §D) is also correct: an intermediate `go` process is not in `wrapperProcessNames`
  (`session_pid.go:40-44`) and exits early, which would corrupt the ancestry walk.

## §5 — Anti-mutant strength

| Mutant | Killed by | Verdict |
|---|---|---|
| (a) records a different but still short-lived pid | AC-INL-001 — the parent stays alive and `stale == false` is asserted after the child exits; any pid belonging to a process that dies with the command fails it. AC-INL-002 pins the value to the parent's pid exactly | **KILLED** |
| (b) `status` never reports reclaimable (symptom hidden) | AC-INL-003 — a genuinely-exited owner must read reclaimable and a bare foreign acquire must succeed. A never-stale mutant fails both halves | **KILLED** |
| (c) widens a staleness window instead of anchoring identity | AC-INL-003 again — its takeover is immediate, so any grace window (e.g. "records younger than N minutes are live") refuses the takeover and fails. AC-INL-001 alone would *not* catch it; the pair does | **KILLED (by the pair, not by 001)** |
| (d) works for fresh locks, misreads pre-existing records | AC-INL-007 — a hand-written old-format record (no `pid_source`, dead pid) must parse, read reclaimable, and be taken over; its stated mutants are (i) rename/retype breaking the parse and (ii) legacy-lives-forever upgrade wedge | **KILLED** |
| (e) asserts on the CLI's printed string, not record state | AC-INL-001 clause (b) reads `pid_source` back from `.moai/state/integration-lock.json` on disk; AC-INL-012 clause (a) asserts `session_id == "sess-a"` in that same file. (v0.1.1 added clause (b) precisely for this) | **KILLED** |
| (f) *anchors correctly for `Stale()`/`status` but leaves `acquire` willing to displace a live holder* | AC-INL-012 — added in v0.1.2. Before it, this mutant passed AC-INL-001..011 in full, and it is the **observed field harm**, not a hypothetical | **KILLED (only since v0.1.2)** |
| (g) M5 doc prose overclaims serialization | **nothing** | **NOT KILLED — D2** |
| (h) `--force`/stale takeover silently drops the `replaced` trace | AC-INL-006 (`--force` leg) + AC-INL-012(b) (refusal must produce no trace) | **KILLED** |

Mutant (g) is the one live gap: plan.md M5 carries a `[HARD]` bound listing forbidden phrases
("직렬화를 보장한다", "동시 acquire는 불가능하다", capability-boundary wording), and no criterion
tests for them. See D2.

## §6 — Shared-state hazard in the verification design

**Not a defect. The design avoids mutating the repository-wide lock, and it does so structurally.**

`integrationLockRoot()` (`integration.go:46-65`) consults `CLAUDE_PROJECT_DIR` **first**, before the
git fallback. Every test in the plan pins `CLAUDE_PROJECT_DIR=<t.TempDir()>` plus
`GIT_CEILING_DIRECTORIES=<temp parent>` (plan.md §D; spec.md §D `[HARD]` test-isolation clause;
acceptance.md preamble), so both the child binary and the in-process helper resolve to the temp
root and the real `.moai/state/integration-lock.json` is never on the path. I verified the
mechanism by using it: my own reproductions in §2 wrote only to
`$SP/red/root{1,2}/.moai/state/integration-lock.json`.

Two further points in the SPEC's favour:

- AC-INL-012 explicitly **declines** to produce a hand-run RED transcript, stating the reason:
  other lanes hold real merge windows right now and the verb mutates the shared record, so the
  failing observation is produced by the Go test under `t.TempDir()` instead. That is the correct
  call, and stating it beats silently omitting the transcript.
- AC-INL-003's "dummy process that has exited" is a spawned child in the test's own control, not a
  real lane's pid.

The plan nowhere prescribes running the real CLI against the real lock. **No BLOCKING defect on
this axis.** Residual note (not a finding): `go build -o ./cmd/moai` inside a test adds build cost
to `./internal/cli/`; bounded, and the plan already restricts verification to affected packages.

## §7 — Scope discipline

- **`internal/hook` is legitimately in scope, but as a test surface only.** M4 adds cases to
  `integration_lock_guard_test.go`; no production change to `integration_lock_guard.go` is
  proposed, and none is needed — the guard consumes `Stale()` and inherits the fix. plan.md §A
  says this ("one hook consumer's test surface"). Not scope creep.
- **`internal/session` is in scope and is missing from `module:`** — see **D1**. M2 exports a new
  seam from that package and unit-tests it; `grep -c 'internal/session' plan.md` → 13.
- **No Go production change exceeds the REQs.** The proposed production delta is: one additive
  struct field, one deleted `os.Getpid()` fallback, one exported resolver in `internal/session`,
  and the acquire-path wiring. plan.md §B consequence 3 explicitly declines a `--pid` override on
  YAGNI grounds. §G Anti-Patterns forbids widening the guard's merge pattern and flipping its
  enablement default. This is disciplined.

## Defects Found (structured defect-list)

D1. **MODULE-FIELD-INCOMPLETE** — `spec.md:L11` — `module: "internal/kanban, internal/cli, internal/hook"` omits `internal/session`, which plan.md M2 modifies with production code (export a `ResolveOwnerPID`-style seam wrapping the existing chain **without** the `os.Getpid()` fallback) plus new unit tests; `grep -c 'internal/session' plan.md` → 13, and plan.md §C's pre-flight is itself a cycle check on that package. The frontmatter therefore understates the touch set a reviewer or a later drift check would read. — Severity: **SHOULD-FIX** — Class: **blocking** (internal inconsistency between frontmatter and the plan's own touch set) — Required fix: set `module: "internal/kanban, internal/cli, internal/session, internal/hook"` and bump `updated:`. This is a non-transition frontmatter correction owned by `manager-spec` (`spec-frontmatter-schema.md` § Non-transition frontmatter corrections); no amendment procedure is triggered.

D2. **UNVERIFIED-HARD-WORDING-BOUND** — `plan.md` M5 (`[HARD]` wording bound) vs `acceptance.md` AC-INL-010 / AC-INL-011 — plan.md M5 states a `[HARD]` constraint with an explicit forbidden-phrase list ("직렬화를 보장한다" and any equivalent unqualified guarantee claim, "동시 acquire는 불가능하다", capability-boundary wording), and **no acceptance criterion tests for it**. Measured: `grep -n "보장\|wording bound\|forbidden" acceptance.md` returns exactly one hit, `:148`, which is AC-INL-010 *quoting the caveat being deleted* — not a check. A doc rewrite that deletes the §3 caveat, adds "세션 프로세스" and "재획득" to `CLAUDE.local.md` §4.1, **and also** asserts "이제 직렬화를 보장한다" satisfies AC-INL-010 and both of AC-INL-011's greps while violating the `[HARD]` bound. That mutant is writable, so the bound is unadopted under `verification-completeness.md` §2 (mutant probe) and §1.2(b) (a check specification names the input that turns it red). The harm is concrete rather than stylistic: the atomicity hazard in §G is real and unfixed, so a lane reading an unqualified guarantee would drop the lead announcement — the first serialization layer — on the strength of a mechanical layer that does not provide mutual exclusion. — Severity: **SHOULD-FIX** — Class: **blocking** (no instrument for a stated `[HARD]` constraint) — Required fix: add a negative-assertion clause to AC-INL-010 (and mirror it in AC-INL-011) naming the exact forbidden strings and asserting zero hits after M5, e.g. `grep -c '직렬화를 보장한다' .claude/rules/local/gitflow-lane-protocol.md CLAUDE.local.md` → **0** and `grep -c '동시 acquire는 불가능' …` → **0**, with the RED-now cell stating that both are trivially 0 today and that the criterion's purpose is to hold them at 0 across M5 (a preserved-invariant cell, paired with AC-INL-010's positive leg).

D3. **PROSE-JUDGED-GREEN-CELL** — `acceptance.md` AC-INL-010 green path ("the same grep returns 0 hits and **the replacement wording is present**") — the deletion leg is mechanical (grep → 0) but the replacement leg is judged by reading. Its sibling AC-INL-011, revised in v0.1.1, names two exact greps with measured baselines (both `0` today), which is the right shape — the asymmetry is within one milestone and one requirement. — Severity: **MINOR** — Class: **optional** — Required fix: give AC-INL-010's positive leg the same treatment, e.g. an `awk`-bounded §3 extraction plus `grep -c` on two required tokens, with the measured 0-baseline stated.

D4. **RISK-DIRECTION-SPLIT** — `spec.md` §G (ancestry bullet) vs `acceptance.md` §D.2 (edge-case bullet 4) — §G's ancestry risk covers only the "wrong **long-lived** process" direction (which fails toward TREAT-AS-LIVE, the cheap direction). The opposite direction — the walk anchoring to a **shorter-lived** process, so the holder reads stale **sooner than truth** — appears only as an acceptance.md edge case, described as "bounded by the wrapper set's accuracy". That is the false-"stale" direction which spec.md §D quotes the code as calling "the failure this lock exists to prevent", so the two directions of one mechanism are documented at different severities in different files, and the dangerous one is the one that is not in Risks. No AC observes it (a wrapper-set-inaccuracy case is genuinely hard to test, which is a reason to state the risk, not to omit it). — Severity: **MINOR** — Class: **optional** — Required fix: move the shorter-lived-anchor direction into §G alongside its sibling, keeping the existing mitigation (identical exposure to the session registry's already-shipped walk) and stating explicitly that it is untested-by-construction.

D5. **AUDIT-WINDOW-CONCURRENT-WRITER** — `.moai/specs/SPEC-INTEGRATION-LOCK-LIVENESS-001/*.md` (all four) — the audit subject was rewritten four times during this audit window: `shasum` taken at 10:46Z, 10:48Z and 10:49Z each returned different digests for `spec.md`, `acceptance.md`, `progress.md` and finally `plan.md`, and the SPEC version advanced `0.1.0 → 0.1.1 → 0.1.2` between my first read and this verdict. `agent-common-protocol.md` § Background Agent Execution binds an actively-audited worktree to one writer from the opening measurement to the landed verdict, and names an observed foreign write a process defect to report rather than absorb. Compounding it, the delegating instruction asserted "No prior audit exists — this is iter-1" while `.moai/reports/t298/plan-audit-iter1.md` (FAIL 0.795) already existed, so the Tier M 2-iteration ceiling was consumed before this audit began. — Severity: **SHOULD-FIX** — Class: **blocking** (process, not SPEC content) — Required fix: for the next audit of any card, freeze the artifact set before dispatching the auditor and state the true iteration number and prior-report path in the dispatch. For **this** verdict: it is pinned to the four hashes in § Attribution and is void against any other state.

D6. **UNLISTED-IDENTITY-STALENESS** — `spec.md` §G (absent); mechanism at `internal/cli/integration.go:71-74, 87` — the record's *session identity* has its own staleness path that §G does not list: `integrationSessionID` falls back to the SessionStart side-channel file, which by the code's own comment "is one slot per project and names whichever session started last", so a lock can be recorded under a foreign session id — and `SessionID` is exactly the field the guard compares (`integration_lock_guard.go:85`) and that REQ-INL-004 makes load-bearing. This SPEC anchors the *pid* and does not touch identity resolution, which is the correct scope; the gap is that §G, which now enumerates PID reuse, the absent host field, and the write race as inherited hazards, omits this one. — Severity: **MINOR** — Class: **optional** — Required fix: add a §G bullet recording it as inherited and out of scope, cross-referencing the code comment that already documents it.

No BLOCKING (must-pass-equivalent) defect. D1, D2 and D5 are blocking-class under M6 and are the
routing set; D3, D4 and D6 are optional and are the orchestrator's discretion.

## Regression Check (prior iteration)

Against `.moai/reports/t298/plan-audit-iter1.md` (iter-1, FAIL 0.795), all nine defects
re-verified on the pinned v0.1.2 state:

| Prior | Status | Evidence |
|---|---|---|
| D1 AC-GAP-PIDSOURCE | **RESOLVED** | AC-INL-001's **Then** now has clause (b) asserting `pid_source == "session-owner"` read back from the on-disk record; its RED-now (b) cell cites `grep -n 'PIDSource\|pid_source' internal/kanban/integration_lock.go` → no match, which I re-ran and confirmed (`rc=1`) |
| D2 FALSE-RED-PREMISE | **RESOLVED** | AC-INL-011 rewritten around the measured absence with two named greps and a stated 0-baseline; §B item 13 restated as a two-bullet distinction (enshrines vs silent). I independently reached the same measurement before reading the prior report |
| D3 PREFLIGHT-EXPECTS-WRONG-OUTPUT | **RESOLVED** | plan.md §C now scopes the cycle check to the import form (`grep -rn '"github.com/…/internal/kanban"' internal/session/` → 0, `rc=1`) and names the 1 unscoped hit as a comment at `session_pid.go:49`. Both re-measured, both agree |
| D4 UNADDRESSED-CONCURRENCY | **RESOLVED** | spec.md §G carries the read-modify-write hazard with the measurement inline; §A adds the identity-vs-atomicity layer distinction; plan.md M5 carries the `[HARD]` wording bound. I re-ran the `Flock\|flock\|LockFile` grep and confirm 5 hits, no call site. *Note:* the bound's own verification is D2 in this report |
| D5 UNATTRIBUTED-VERIFICATION-CLAIM | **RESOLVED** | progress.md §E.1 now quotes the exact regex command and pattern and explains why the multi-segment form applies |
| D6 ABBREVIATED-TRACE-TOKEN | **RESOLVED** | §D matrix row now reads `REQ-INL-004, REQ-INL-008`; `grep -c 'REQ-INL-008' acceptance.md` → ≥1 |
| D7 NON-CANONICAL-PATTERN-LABEL | **RESOLVED** | REQ-INL-003 now labeled `(Event-driven)` |
| D8 RISK-ENUMERATION-GAP | **RESOLVED** | §G gains PID-reuse and absent-host bullets; I confirmed `IntegrationLock` (`:71-79`) has no `Host` field while `session.Entry` does |
| D9 REQ-VS-PLAN-BRANCH-AMBIGUITY | **RESOLVED** | REQ-INL-006 reworded as a preserved invariant ("no marker-conditional branch is added"), matching plan.md §B consequence 4 |

**9/9 resolved.** No stagnation. Beyond the repair set, v0.1.2 added AC-INL-012 — closing a gap
neither iter-1 nor my own pre-report analysis had enumerated, traced to existing REQs without
inventing one, and grounded in a production occurrence rather than a hypothetical. That is the
strongest single item in this SPEC.

## Recommendation

**PASS-WITH-DEBT.** Score 0.92 clears the Tier M threshold of 0.80; all seven must-pass criteria
are PASS or N/A; no D7 or D8 BLOCKING finding exists. The root cause is correct and independently
reproduced, the RED-first discipline is real and verified red on this tree, traceability is
complete, and the verification design does not touch shared state. Three blocking-class findings
remain open (D1, D2, D5), none of which trips a must-pass — hence debt rather than FAIL.

The Tier M iteration ceiling (2) is **consumed**, so these are carried into run-phase rather than
into a third plan-phase audit:

1. **D1 — before M1.** One-line frontmatter repair, `manager-spec`: add `internal/session` to
   `module:`, bump `updated:`. No amendment procedure.
2. **D2 — before M5, ideally now.** Add the negative-assertion clause naming the forbidden strings
   to AC-INL-010/011 so the `[HARD]` wording bound has an instrument. Without it, M5 can ship prose
   that reinstates the operational hazard this card exists to reduce, and every AC will still be
   green. This is the one finding I would push back on shipping without.
3. **D5 — process, next card.** Freeze the artifact set before dispatching an auditor, and state
   the true iteration number and prior-report path in the dispatch.
4. **D3, D4, D6 — optional.** Route at the orchestrator's discretion; none affects correctness.

Nothing here blocks Implementation Kickoff Approval, which remains a separate mandatory gate this
verdict neither satisfies nor bypasses.

## Gaps (explicitly NOT checked)

- **`go test ./internal/kanban/` was not run.** The dispatch permitted it; I did not need it —
  no existing test's behavior is at issue in a plan-phase audit, and the RED claims were verified
  more directly by driving the real built binary. Existing-suite health on this tree is therefore
  unobserved by me.
- **`moai spec lint` covered the whole `.moai/specs/` tree, not this SPEC alone.** It completed
  (exit 0, `0 error(s), 64 warning(s)`) with zero findings naming this SPEC, which is what MP-3
  consumes; the 64 warnings belong to other, mostly grandfathered SPECs and were not triaged.
- **`GOOS=windows GOARCH=amd64 go vet ./internal/...` was not run.** AC-INL-009's compile leg is
  an M4 obligation, not a plan-phase precondition.
- **The M1/M2/M3/M4 tests do not exist yet**, so no test was executed — RED was verified by
  reproducing each assertion by hand against the built binary, which is evidence that the
  assertions fail today, not that the authored tests will fail for the same reason.
- **Windows behavior was not observed.** Only the code-path argument was read. spec.md §G already
  flags Windows launcher stamp propagation as unverified.
- **`design.md` / `research.md` were not read** — correctly absent at Tier M.

## Residual risk (what could still be wrong despite the above)

- **The subject may move again.** D5's writer was still active at 10:49Z. If any of the four pinned
  hashes differs when this is read, the verdict does not describe the artifacts in the tree.
- **Ancestry-walk behavior under the real Claude Code process tree is untested.** My reproductions
  ran under a plain shell whose parent chain differs from a lane session's
  (`moai` → `sh` → `claude`). The walk could anchor to a process whose lifetime differs from the
  session's in either direction (D4). Both directions are argued, neither is measured, and no AC
  observes either.
- **`pid_source` has no reader.** Under plan.md §B consequence 4 the marker is written and never
  branched on, because `Stale()` already handles both record shapes. That is the simplest correct
  design, but it means the field's only consumers are AC-INL-001(b) and human forensics — a future
  reader could mistake it for a semantic discriminator. Cheap to mitigate with a comment at the
  field; not worth a finding.
- **The atomicity hazard remains open after this card lands.** Post-fix, serialization becomes real
  everywhere except a concurrent-acquire window that is correctly scoped out. If D2 is not fixed,
  the documentation may not say so.
- **PID reuse and the single-host assumption** are recorded in §G but unmitigated and untested;
  both fail toward TREAT-AS-LIVE, so the residual cost is an operator `--force`, not a double merge.
