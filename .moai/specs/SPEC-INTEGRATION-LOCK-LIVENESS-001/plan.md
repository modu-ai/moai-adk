# SPEC-INTEGRATION-LOCK-LIVENESS-001 — Implementation Plan (card t298)

Tier: **M** (justification in §A). Status: draft. Baseline tree: `d29b8942e` @
`WT-integration-lock`.

## §A Context and Tier Verdict

**Tier M.** The touch set is small — one Go package (`internal/kanban`), CLI wiring
(`internal/cli/integration.go`), one hook consumer's test surface, and two local-only docs — but
Tier S is ruled out by: (1) the red test is cross-process (child binary exec + parent
assertion), a harness shape the in-process `runIntegration` helper cannot express; (2) two
consumer surfaces (CLI status + PreToolUse guard) plus two platforms (unix ancestry walk vs
Windows conservative degradation) must be covered by distinct ACs; (3) an on-disk schema
backward-compatibility contract is in play; (4) two documentation rewrites are scheduled as AC
rows. A single-file inline-AC Tier S shape cannot carry those obligations; Tier L is excessive
(no ≥3-milestone multi-package coordination beyond what M5 covers, no research.md needed — the
verified-premises section of spec.md §B already carries the measured baseline).

Design decisions are ordered below by decision-reversibility: the schema field and anchor
semantics first (hardest to reverse once records are on disk), the resolution-chain wiring
second, mechanical test/doc work last.

## §B Design-Space Evaluation

### Option A — registry-authoritative staleness (rejected as primary)
`Stale()` consults the session work registry (`active-sessions.json`) for the recorded
`SessionID`: presence/freshness of the entry replaces the bare pid probe; recorded pid kept as
secondary evidence.

- Pros: registry entries already carry exactly `{session_id, pid, host, cwd, last_heartbeat}`;
  `Register` already records the true session pid via `resolveSessionPID()`.
- Cons (decisive): **freshness is the wrong oracle.** `DefaultStaleMinutes = 30` means a lane
  that is alive but quiet (paused mid-resolution, human-paced conflict work) reads stale after
  30 minutes — the false-"stale" direction whose cost the lock's own doctrine comment
  (integration_lock.go:84-90) calls the failure this lock exists to prevent. Heartbeats are
  hook-driven and best-effort; a lane working in a plain shell never registers at all, so its
  registry absence is indistinguishable from death. Using registry ABSENCE as "gone" also
  inherits the purge window (slow crash detection) and couples the lock to a store the lane
  may not maintain.
- Disposition: rejected as the decision input. Registry data remains available as advisory
  context for humans, nothing more.

### Option B — owner-pid anchoring at acquire time (CHOSEN)
The acquire verb resolves the calling session's owner pid using the same chain
`internal/session/session_pid.go` established for the identical defect shape:
`MOAI_SESSION_PID` (env stamp, only when it names a live process) → nearest non-wrapper
ancestor → **unresolved**. The resolved pid is recorded in the lock's existing `PID` field
plus an additive marker field (e.g. `pid_source: "session-owner"`); `Stale()`'s probe logic is
unchanged — it now simply probes a pid whose lifetime equals the session's lifetime.

- Pros: smallest behavioral diff with the exact reuse the codebase already validated for this
  failure ("the registry is written from a subprocess that exits within milliseconds" —
  session_pid.go's opening comment is this SPEC's problem statement, already solved once);
  immediate, correct crash detection (pid probe, not a 30-minute window); preserves the
  false-live/false-stale asymmetry verbatim; Windows behavior derived, not special-cased.
- Cons: needs the resolution chain reachable from the acquire path (the unexported
  `resolveSessionPID` lives in `internal/session`); the removed `os.Getpid()` fallback means
  unresolvable owners become conservative pid-0 (see C); ancestry depends on the wrapper-name
  set being accurate.

### Option C — indeterminate treated-live (COMPOSED as fallback tier, not standalone)
Record pid 0 (owner-anchor marker, no pid): `Stale()` already treats `PID <= 0` as live
(integration_lock.go:95-97). Standalone it would make every window live-forever-until-`--force`
— serialization without reclamation. Composed as Option B's fallback tier it is exactly the
documented conservative asymmetry: when owner identity cannot be established, prefer one
operator `--force` over a silent double merge.

### Verdict
**B, with C composed as the explicit unresolvable fallback; A rejected as decision input.**
One-line rationale: the codebase already built and tested the owner-pid resolution chain for
the same dead-on-arrival-subprocess defect, so anchoring to it is the smallest diff that makes
liveness both immediate and correctly asymmetric, while any freshness-based oracle (A) would
reintroduce the catastrophic false-"stale" direction on a timer.

### Chosen-shape consequences (the decisions M2 must implement)
1. **Schema (additive-only):** new optional field on `IntegrationLock`, e.g.
   `PIDSource string \`json:"pid_source,omitempty"\`` with the single value `"session-owner"`
   written by the anchored acquire path. Absent field = legacy record. No other field changes.
   Old records keep today's semantics (dead recorded pid → reclaimable); no wedge at upgrade.
2. **Resolution placement:** the owner-pid resolution runs in the **acquire CLI path**
   (`internal/cli/integration.go`), which already imports both `internal/kanban` and
   `internal/session` — so `internal/kanban` gains no new import and stays pure. Export the
   minimal seam from `internal/session` (e.g. a wrapper over `resolveSessionPID` that RETURNS
   the unresolved sentinel instead of falling back to `os.Getpid()` — the registry's fallback
   is its own cost profile; the lock must not inherit it).
3. **Defect-line removal:** `if want.PID == 0 { want.PID = os.Getpid() }`
   (integration_lock.go:174-176) is deleted — an unset pid from the anchored path is the
   legitimate conservative pid-0, not a value to fill with the CLI's own pid. Explicit
   `--pid`-style override is NOT added (no caller needs it; YAGNI).
4. **`Stale()` semantics:** unchanged for legacy records; for `pid_source == "session-owner"`
   records the existing probe already behaves correctly (pid>0 → probe; pid==0 → live).

## §C Pre-flight (before M1)

- Confirm the tree: `git rev-parse --short HEAD` → `d29b8942e`; branch `WT-integration-lock`.
- Confirm no other lane is mid-flight on these files (read `git status --short`).
- Confirm `internal/session` does not import `internal/kanban` (no cycle when exporting the
  seam): `grep -rn 'internal/kanban' internal/session/` → 0 hits expected.

## §D Constraints (carried from spec.md §D)

- TREAT-AS-LIVE on every uncertainty path (the file's own doctrine; asymmetry "not close").
- Additive-only JSON schema; legacy records behave as today.
- No heartbeat/freshness oracle may decide `Stale()`.
- All tests under `t.TempDir()`; `CLAUDE_PROJECT_DIR` + `GIT_CEILING_DIRECTORIES` pinned; the
  real primary-checkout `.moai/state/integration-lock.json` / `active-sessions.json` are never
  touched. Build the test child binary with `go build -o`, never `go run` (an intermediate
  `go` process is not in the ancestry walk's wrapper set and exits early, breaking the walk).
- Verification discipline: affected packages only (`go test ./internal/cli/... ./internal/kanban/... ./internal/session/... ./internal/hook/...`),
  then push and read CI for the full-suite verdict. No local `go test ./...`.
- Code, comments, godoc in English; `GOOS=windows go vet ./...` (or build) for parity compile
  checks — cross-build does not compile tests, so Windows behavior is verified through the
  conservative-path unit tests that are platform-neutral.

## §E Self-Verification (milestone exit evidence)

Each milestone's exit is a command plus observed output recorded in progress.md §E.2:
- M1 exit: the new cross-process test FAILS on the unmodified tree (observed failure text
  names "reclaimable" where HELD was asserted) — the defect made mechanical.
- M2 exit: the same test PASSES; `go test ./internal/cli/ ./internal/kanban/` green.
- M3 exit: regression-pair tests green (legit-stale, foreign-release, force-reporting,
  legacy-record compat).
- M4 exit: guard-consumer tests green; `GOOS=windows go vet ./internal/...` clean.
- M5 exit: grep shows the caveat text gone from both docs and the replacement text present.

## §F Milestones

### M1 — RED: cross-process reproduction test (Priority: High, FIRST)
Author `internal/cli/integration_lock_owner_liveness_test.go`:
- Helper `buildMoaiBinary(t)` (once per run): `go build -o <t.TempDir()>/moai ./cmd/moai`.
- Helper `runIntegrationChild(t, root, env, args...)`: exec the built binary with
  `CLAUDE_PROJECT_DIR=<temp root>` and `GIT_CEILING_DIRECTORIES=<temp parent>` pinned
  (same pinning contract as `runIntegration`, integration_lock_cli_test.go:26-43).
- **AC-INL-001 test:** child runs `integration acquire --session <id>` (no env stamp —
  exercises the ancestry walk), child EXITS, parent (the go test process, alive) runs
  `integration status --json` on the same temp root, assert `stale == false` / held. Runs on
  the unmodified tree → FAILS (reads reclaimable; observed output is the evidence).
- **AC-INL-002 test (env-stamp variant):** same shape but child env carries
  `MOAI_SESSION_PID=<parent pid>` → assert held/not-reclaimable after child exit. Also fails
  today (current code ignores the env var and records the child pid). This variant is the
  platform-neutral one (no ancestry walk) and doubles as the Windows-behavior proxy.
- Verification: `go test ./internal/cli/ -run 'TestIntegrationOwnerLiveness' -v` → both FAIL
  with the reclaimable-where-held assertion; record verbatim output in progress.md §E.2.

### M2 — GREEN: owner-pid anchor lands (Priority: High)
- `internal/session`: export the seam (e.g. `ResolveOwnerPID() (pid int, resolved bool)`)
  wrapping the existing chain WITHOUT the `os.Getpid()` fallback; unit-test the precedence
  (env-stamp-live → ancestry → unresolved) next to the existing session_pid tests.
- `internal/cli/integration.go` acquire path: resolve owner pid, record `PID` +
  `pid_source: "session-owner"` (resolved or 0).
- `internal/kanban/integration_lock.go`: add the additive `PIDSource` field; DELETE the
  `os.Getpid()` fallback (integration_lock.go:174-176); `Stale()` semantics per plan §B.4
  (anchored records: pid 0 → live; legacy records: unchanged).
- Exit: M1's two tests flip green; `go test ./internal/cli/ ./internal/kanban/ ./internal/session/` green.

### M3 — Regression pairs and backward compatibility (Priority: Medium)
- Legit-staleness pair (AC-INL-003): acquire with `MOAI_SESSION_PID` pointing at a spawned
  dummy child that is waited-on and exits → status reads reclaimable WITHOUT `--force`, and a
  foreign acquire succeeds reporting the displaced holder. This is the mutant guard against
  the "always record pid 0" shortcut that would satisfy AC-INL-001 vacuously.
- Conservative pair (AC-INL-004): anchored record with pid 0 → held/live (protects the
  conservative path against a "never conservative" mutant that would reintroduce
  `os.Getpid()`-style behavior under another name).
- Foreign-release refusal (AC-INL-005) and `--force` takeover displaced-reporting (AC-INL-006)
  — extend the existing in-process tests' style; takeover trace must never be silently
  dropped.
- Legacy-record compatibility (AC-INL-007): hand-write an old-format JSON (no `pid_source`,
  dead pid) into a temp root → parses, reads reclaimable; also assert re-acquire self-refresh
  still works.

### M4 — Guard consumer + platform parity (Priority: Medium)
- `internal/hook/integration_lock_guard_test.go`: anchored-live holder → deny with sentinel;
  anchored-dead holder → allow + advisory; anchored pid-0 holder → deny (conservative);
  legacy record → unchanged behavior.
- `GOOS=windows GOARCH=amd64 go vet ./internal/...` (compile parity); confirm the
  env-stamp variant (M1) covers Windows-behavior semantics via the platform-neutral path.
- Confirm `factory_alive_windows.go` untouched (`git diff --stat` shows no hit).

### M5 — Documentation rewrite (Priority: Medium)
- `.claude/rules/local/gitflow-lane-protocol.md` §3: replace the two-line caveat blockquote
  (lines 42-43, "직렬화를 보장하지 못한다 — 카드 t298 …") with the restored guarantee: the lock
  now anchors liveness to the session process; legacy pre-fix windows read reclaimable —
  re-acquire after upgrade; lead-notice remains the doctrine's first layer, the record is the
  mechanical layer under it.
- `CLAUDE.local.md` §4.1/§4.1.2: update the lead-notice-only serialization prose to the
  restored two-layer description.
- Both are local-only files — no template mirror, no `make build`.

## §G Anti-Patterns (do not)

- Do NOT decide staleness by heartbeat freshness or registry absence (Option A's failure).
- Do NOT fill an unresolved owner pid with `os.Getpid()` — that is the defect, in any guise
  (including "as a fallback so the field is never empty").
- Do NOT rename/retype existing JSON fields; the schema change is one additive optional field.
- Do NOT run the red test through `go run` or the in-process cobra harness — neither
  simulates "the CLI process exits".
- Do NOT touch the live primary-checkout state files during development or testing.
- Do NOT widen the guard's merge pattern or flip its enablement default.

## §H Cross-References

- spec.md §B premises (line-pinned baseline), §C REQ-INL-001..010, §D constraints.
- acceptance.md AC-INL-001..011 (two-cell discipline; each row names its flipping milestone).
- progress.md §E skeleton (this file's §E defines the per-milestone exit evidence format).
- Upstream doctrine: `.claude/rules/moai/workflow/kanban-dispatch.md` § "Integration into the
  release branch is self-served"; local operating rule:
  `.claude/rules/local/gitflow-lane-protocol.md` §3.
- Owner-pid lineage: `internal/session/session_pid.go`, `internal/cli/launch_session_pid.go`.
