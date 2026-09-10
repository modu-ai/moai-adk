---
id: SPEC-SYNC-GATE-FAILSTATE-001
title: "Sync-phase quality gate — failure state survives re-invocation on the same HEAD"
version: "0.1.0"
status: in-progress
created: 2026-09-10
updated: 2026-09-10
author: manager-spec (card t624)
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude/hooks/moai; internal/template/templates/.claude/skills/moai/workflows/sync; internal/hook"
lifecycle: spec-anchored
tags: "hook, stop-hook, sync-gate, quality-gate, state, docs-accuracy"
tier: M
---

# SPEC — Sync-phase quality gate: failure state survives re-invocation

## HISTORY

- 2026-09-10: plan-phase authored on card t624. Defect H01 reproduced on tree `d5dc42959`
  (`.moai/reports/t624/h01-repro-develop.md`). Every target file is byte-identical between
  `d5dc42959` and the plan-phase tree `fa96fe644` (acceptance.md §D.0 L-01), so that
  reproduction describes the current tree. Direction D1-D6 is lead-approved and encoded
  here as decided (§2), not re-opened.
- 2026-09-10: lead rulings on the plan-phase open decisions applied (plan.md §B), all in the
  "never looser than today" direction. B1: after the single allowed stale re-run also fails
  to complete, later turns emit a non-blocking notice instead of staying silent (REQ-009;
  AC-006c, AC-015). B2: mode is resolved again at re-delivery, and an advisory first run is
  never retroactively blocked (REQ-006). B3: the advisory warning is emitted once (REQ-006).
  B4: behavioral RED cells are adopted at the M1 test-only commit (acceptance.md legend).
  B5: record only. B6: the doc line 43 claim moved out of this SPEC (§5).
- 2026-09-10: plan-audit round 1 findings applied (`.moai/reports/t624/plan-audit.md`; 11
  blocking, 7 optional), inside the existing decisions and in the "never looser" direction.
  Two structural changes:
  - **Auxiliary state.** A payload file and a retry marker are now named, each a separate file
    beside the closed-token record (REQ-001, REQ-002). Every check-running invocation
    invalidates the payload.
  - **Payload generalized.** D1's persisted block now covers the output of every failing run,
    advisory included. This lets the audit's write-ordering choice (a `fail` record without a
    payload re-gates, REQ-007) hold without breaking B2 or B3.

  REQ-010 is restated as an invariant over call history. REQ-013 is anchored on content rather
  than line numbers.
- 2026-09-10: lead-requested additions made before audit round 2.
  - B7 approved as a D1 state-representation decision: every failing run stores its payload
    (block or advisory message) before the `fail` record.
  - Torn-write rows added to AC-005 (payload present with record not `fail`; neither written),
    with mutants M23 and M24.
  - The mtime-unreadable fallback recorded as an explicit gap.

  This revision was audited in round 2 (`.moai/reports/t624/plan-audit-r2.md`, FAIL 0.86).
- 2026-09-10: plan-audit round 2 findings applied (N-1..N-8, O2), inside the existing decisions.
  - AC-015 N2 split into N2a (mtime-only refresh) and N2b (full record rewrite), with M26.
  - AC-004 F4 tab form added, with M25. REQ-005 narrowed: no top-level claim; nested keys are an
    explicit gap.
  - REQ-008/009 boundary defined (stale only when the age is strictly greater than 60 s); b61
    row and M27 added; the exact-60 s point recorded as an explicit gap.
  - Interrupted-run harness: per-run marker lifecycle and non-fatal assertions.
  - U4 read step added; L-18 regex extended to `then` / `do`; L-21 pending ledger row added.

  This revision was audited in round 3 (`.moai/reports/t624/plan-audit-r3.md`, FAIL 0.86, 1
  blocking).
- 2026-09-10: operator decision after round 3 — PASS-with-debt. Implementation Kickoff was
  approved on condition that NEW-1 is resolved before M1. The debt is paid in this revision, the
  first run-phase change set:
  - NEW-1: the AC-015 N2a/N2b stub-count deltas are stated as check runs with their invocation
    counts.
  - NEW-2: the L-21 selector runs verbatim, with a coverage contingency.
  - NEW-3: a word boundary before `then` / `do` in the L-18 regex.
  - NEW-4: stale cross-references updated.
  - NEW-5: HISTORY entries and plan §A verdict pointers.

  Status stays `draft`; the `draft → in-progress` transition belongs to the run-phase owner.

## §1 Background and problem statement

### §1.1 H01 — a failure on a HEAD disappears on the next turn

The Stop hook `sync-phase-quality-gate.sh` exists in two copies that must stay byte-identical:
`.claude/hooks/moai/sync-phase-quality-gate.sh` and
`internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` (contract:
`internal/hook/wrapper_copies_contract_test.go`, entry `sync-phase-quality-gate.sh`). No
`.sh.tmpl` exists for this hook.

Reproduced in an isolated git fixture (Go project, sync-phase commit subject; evidence
`.moai/reports/t624/h01-repro-develop.md` and the five `h01-develop-*.out` files):

| Call | HEAD | Code state | stdout | Log line |
|---|---|---|---|---|
| call1 | `0a9aa6d` | build fails | 237-byte block JSON | `decision=block` |
| call2 | `0a9aa6d` (same) | same | **0 bytes** | **none** |
| control A | `124e484` | build fails | 237-byte block JSON | `decision=block` |
| control B1 | `c08b6e4` | build passes | 0 bytes | `decision=allow` |
| control B2 | `c08b6e4` (same) | same | 0 bytes | none |

Root cause (hook lines 176-186 on `fa96fe644`): the hook writes **only** the HEAD SHA to
`.moai/state/sync-quality-gate.last`, and it does so **before** the checks run. Lines 179-182
then exit 0 whenever the recorded content equals the HEAD SHA, whatever the earlier run's
outcome was. A failing sync commit therefore blocks exactly one turn, and every later turn on
that HEAD passes silently with no check, no log line, and no output.

Control A shows that re-gating on a new HEAD works; control B shows that a silent re-call
after a pass is correct and must stay that way.

### §1.2 H03 — the sync workflow document promises a scan the hook does not run

`quality-gates-quality.md` (both copies) says, at line 145, that a dependency vulnerability
scan "runs automatically via the Stop hook". At line 142 it says the audit covers every
manifest present at the project root, and at line 116 it calls this "the only check for that
drift". The hook's comments call the step a "dependency manifest audit", once in the header
(line 3) and again above the manifest step (line 287).

What the hook actually does (lines 287-314 on `fa96fe644`) is run `git diff` over the
detected language's manifests for the HEAD commit and set `deps_modified=1` when that diff is
non-empty. The value appears only in the output message and the log; it never drives the
decision. No vulnerability scan runs, and manifests the HEAD commit did not change are never
examined.

### §1.3 SX-R05 — two severity rules in one document

In the same document, Step 0.5.4 (line 71) states the sync-auditor Security rule: any
Critical/High finding makes the result FAIL. Phase 8 (lines 136-137 and 156) states that only
CRITICAL blocks and HIGH is a warning. The canonical rubric is
`.claude/agents/moai/sync-auditor.md` (Evaluation Dimensions, Security row: "Any Critical/High
finding"). The document never says how the two rules relate, so a reader can conclude that a
HIGH finding which the sync-auditor already failed is cleared by Phase 8.

## §2 Decisions encoded (lead-approved — not re-opened here)

- **D1 — outcome record.** The existing state path holds `<sha> <running|pass|fail>`. On
  `fail` the emitted payload is persisted beside it and re-emitted verbatim when the same HEAD
  is gated again, but only when that payload is a block (REQ-003, REQ-006). A legacy record
  holding only a SHA means the outcome is unknown, and the gate runs once more.
- **D2 — in-progress.** A `running` record for the same HEAD whose state-file mtime is older
  than the hook timeout (60 s, `internal/template/templates/.claude/settings.json.tmpl`, entry
  for `sync-phase-quality-gate.sh`) triggers one more run. A fresher `running` record is not
  re-run; the gate emits a non-blocking notice instead. **No path may be looser than the
  current hook**: every situation in which today's hook blocks must still block (REQ-010).
- **D3 — `stop_hook_active`.** When stdin carries `stop_hook_active` set to `true`, a stored
  failure is not re-emitted as a block, so the runtime Stop-hook block cap is not spent.
  Re-delivery happens on the next fresh turn. Detection is jq-free.
- **D4 — explicit retry.** There is no new flag or environment variable. Deleting the state
  file forces a new gate run, and the hook documents this.
- **D5 — H03 wording.** Doc lines 116 and 140-146 (both copies) and the hook's manifest
  comments (both copies) are corrected to describe manifest-change observation. No gate
  behavior is added or removed.
- **D6 — SX-R05 wording.** The document states that the sync-auditor rubric is canonical and
  that Phase 8 is an additional lens whose CRITICAL-only stop gate never clears an earlier
  sync-auditor FAIL. Wording only.

## §3 GEARS requirements

- **REQ-001 (Ubiquitous):** The gate shall keep, for the HEAD it gates, a single-line record
  `<sha> <outcome>` at `.moai/state/sync-quality-gate.last`. `<outcome>` shall be exactly one
  token from the closed set `running`, `pass`, `fail`; the gate shall write no other token and
  no further field. The record shall read `running` before any check starts. When the checks
  finish it shall read `fail` if any check failed — whether the resolved mode then blocked or
  only advised — and `pass` otherwise.

- **REQ-002 (Ubiquitous/When):** The gate shall keep two auxiliary files under `.moai/state/`,
  each bound to its HEAD and separate from the record:
  - a **payload file** holding the exact stdout bytes a failing run emitted, together with
    whether those bytes were a block or an advisory message;
  - a **retry marker** recording that the one stale re-run of REQ-009 has been used for that
    HEAD.

  **When** an invocation runs the checks, the gate shall, before any check starts, atomically
  remove the payload file and set the retry marker. The marker shall be present if and only if
  this invocation is that one stale re-run; otherwise it shall be removed. **When** a run
  finishes with outcome `fail`, the gate shall atomically write the payload file first, and
  only then write the `fail` record.

- **REQ-003 (When):** **When** the gate is invoked on a HEAD whose record is `fail`, whose
  payload file holds a block for that HEAD, and for which the mode resolved at this invocation
  is blocking, the gate shall not run the checks and shall write the stored block to stdout
  byte-identical — unless REQ-005 applies.

- **REQ-004 (When):** **When** the record for the current HEAD is `pass`, the gate shall not
  run the checks and shall write nothing to stdout. **When** there is no record, or the
  recorded SHA differs from HEAD, the gate shall run the checks as it does today.

- **REQ-005 (When):** **When** stdin carries a `stop_hook_active` key set to `true` and the gate
  would re-deliver a stored block under REQ-003, the gate shall emit no output carrying a
  `decision` field, and shall leave the record and the payload file unchanged so the next
  invocation without the flag re-delivers the block. Detection shall not use `jq`. It shall
  recognize the key in object-key position whatever whitespace separates the key, the colon,
  and the value: none, one or more spaces, or tabs. A `stop_hook_active` literal appearing
  escaped inside a JSON string value (such as `last_assistant_message`) shall not count as the
  key. Detection is not required to tell a top-level key from a nested one (acceptance.md §D.0,
  explicit gap). This requirement shall not suppress a block from a run that executes the
  checks.

- **REQ-006 (When):** **When** the record for the current HEAD is `fail` and its payload file
  holds a block, the gate shall resolve the blocking-versus-advisory mode at this invocation,
  from the same inputs and rules a check run uses (`MOAI_SYNC_GATE_BLOCKING`,
  `MOAI_AUTONOMY_TIER`, and the stored failed-check composition). If that mode is advisory,
  the gate shall neither run the checks nor write anything to stdout. **When** the record is
  `fail` and its payload file holds an advisory message, the gate shall neither run the checks
  nor write anything to stdout, whatever the mode resolves to now: an advisory first run is
  never retroactively blocked, and the advisory warning is written once, by the run that
  executed the checks, as it is today.

- **REQ-007 (When):** **When** the record file is empty or unreadable, or names the current
  HEAD and is a bare SHA with no outcome token (the legacy format), carries a token outside the
  closed set or any additional field, or is `fail` with no payload file for that HEAD, the gate
  shall treat the outcome as unknown, run the checks, and rewrite the record per REQ-001.

- **REQ-008 (While/When):** **While** the record for the current HEAD is `running` and the
  record file's age is at most the stale window, **when** the gate is invoked, it shall not run
  the checks and shall emit only a non-blocking `systemMessage` saying that the previous gate
  run for this HEAD has not completed. The stale window shall equal the timeout the shipped
  settings template registers for this hook (60 s), and the age comparison shall use that
  value. A record is stale only when its age is strictly greater than the window; an age of
  exactly 60 s is fresh and gets the notice.

- **REQ-009 (When):** **When** the record for the current HEAD is `running` and its age is
  strictly greater than the stale window, the gate shall run the checks. That re-run shall happen at most once per
  HEAD. **When** that single re-run also leaves a stale `running` record, every later
  invocation for that HEAD shall neither run the checks nor stay silent: it shall emit only a
  non-blocking `systemMessage` saying that this HEAD's gate run has not completed and that
  deleting the state file forces a new gate run. That notice shall never carry a `decision`
  field, so repeated notices never count toward the runtime Stop-hook block cap.

- **REQ-010 (Ubiquitous):** The gate shall block on every invocation on which the hook at
  `fa96fe644`, given the same history, would block. A history is any sequence of invocations —
  each with its own HEAD, environment, and stdin — that starts from no state record or from a
  legacy bare-SHA record. Within it, state files change only through the gate itself, through
  removal of state files (REQ-011), or through an interrupted gate run.

  > **Why this holds (equivalence argument, non-normative).** On `fa96fe644` an invocation
  > blocks only if all four hold:
  > 1. it passes the early exits — sync-phase subject, recognized language marker, non-zero
  >    code delta;
  > 2. the record content differs from the bare HEAD SHA;
  > 3. a check fails;
  > 4. the mode resolves to blocking.
  >
  > In such a history the record is always absent or the bare SHA of the last gated HEAD, so
  > condition 2 means "no record, or a record for another HEAD". The new gate writes its record
  > for the same last gated HEAD, so in exactly those cases it also sees no record or another
  > HEAD, and REQ-004 makes it run the checks. The early exits, check commands, and mode
  > resolution are unchanged and run in the same order (REQ-012), so conditions 3 and 4
  > produce the same block. The new state logic only decides whether the checks run, and it
  > runs after the early exits, so each remaining input splits into independent classes: the
  > subject pattern, the language branch, the failed-check composition, the mode inputs, the
  > stdin form, and the starting state. acceptance.md §D.7 takes one representative per class.

- **REQ-011 (Ubiquitous):** The gate shall run the checks on the first invocation after
  `.moai/state/sync-quality-gate.last`, or `.moai/state` as a whole, has been removed. Its
  header comment shall document that removal as the way to force a re-gate. The gate shall add
  no flag or environment variable for retrying, and shall keep every file it introduces under
  `.moai/state/`.

- **REQ-012 (Ubiquitous):** The hook shall exit 0 on every path and block only through stdout
  JSON. It shall invoke no `jq`. Its early exits, per-language detection, check commands, mode
  resolution, and `--skip-hook` path shall stay unchanged for all 16 supported programming
  languages. Its comments shall describe the outcome record and auxiliary state (REQ-001
  through REQ-011) instead of the once-per-commit sentinel they describe on `fa96fe644`.
  Template content shall carry no SPEC ID, card id, internal date, or commit SHA.

- **REQ-013 (Ubiquitous):** Both copies of the hook shall be byte-identical. The Phase 9
  anchor line `Purpose: Ensure code has appropriate @MX annotations for AI agent context.
  Supports all 16 MoAI-ADK languages.` shall appear exactly once in each copy of
  `quality-gates-quality.md`. The two copies shall be identical from their first line through
  that anchor line, and the text after it in each copy shall be unchanged from `fa96fe644`.

- **REQ-014 (Ubiquitous):** Doc lines 116 and 140-146 (both copies), the hook header's
  Purpose line, and the hook's comment above the manifest step (both hook copies) shall
  describe manifest-change observation: the hook sets `deps_modified` when a dependency
  manifest of the detected language changed in the HEAD commit; the value is informational,
  never drives the decision, and is not a vulnerability scan. None of them shall claim that the
  hook scans vulnerabilities, audits every manifest at the project root, or detects
  transitive-vulnerability drift. Neither hook copy shall contain the phrase `manifest audit`.

- **REQ-015 (Ubiquitous):** `quality-gates-quality.md` (both copies) shall state that the
  sync-auditor Security rule (a Critical/High finding makes the result FAIL) is canonical, and
  that Phase 8's CRITICAL-only stop gate is an additional lens which never clears a sync-auditor
  FAIL. The Step 0.5.4 HARD THRESHOLD sentence shall stay as it is, and no gate behavior shall
  change.

## §4 Constraints

- **Test resets must keep working.** `internal/hook/stopchain_ac004_006b_test.go` and
  `internal/hook/stopchain_trim_test.go` reset the gate by deleting `.moai/state` as a whole.
  All new state lives under `.moai/state/`.
- **Compliance scan.** `internal/template/hook_official_compliance_test.go`
  (`TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput`) inspects the **first**
  line containing `printf`, `decision`, and `block`. No re-emission line may become that first
  match unless it carries the `hookSpecificOutput` wrapper and `hookEventName` `Stop`.
- **Autonomy tiers.** `internal/hook/stopchain_ac004_006b_test.go` pins advisory-versus-blocking
  per tier; a re-delivered failure must never block under an advisory resolution (REQ-006).
- **Stays dependency-free.** No `jq` invocation
  (`.claude/rules/moai/development/hook-independence.md` §4, row d) and no dependency on the
  `moai` binary.
- **Always exit 0.** Blocking goes through stdout JSON only.
- **Verification scope.** Targeted `go test ./internal/hook/ ./internal/template/` runs. No
  local full suite and no `make build` in run-phase verification (dispatch constraint; plan.md
  §B, B5). Template-First editing order is a process note (plan.md §D), not a verifiable
  requirement.

## §5 Out of Scope

### Out of Scope — a real dependency vulnerability scanner

- Running `govulncheck` or any other scanner from the gate. That is another card's work. This
  SPEC only corrects the wording that claims such a scan exists (REQ-014).
- Adding any check for transitive-vulnerability drift across manifests the HEAD commit did not
  change.

### Out of Scope — Phase 8 severity behavior

- Making Phase 8 block on HIGH findings. REQ-015 reconciles the wording only.
- Changing the sync-auditor rubric or its must-pass firewall.

### Out of Scope — other claims in the same document

- Observed during plan-phase and handled elsewhere: line 43's claim that the Stop hook reads the shared diagnostic snapshot.
- Everything after the Phase 9 anchor line (REQ-013), where the two document copies already
  differ.

### Out of Scope — new controls and neighbouring mechanisms

- Any new flag, environment variable, or configuration key for retrying or disabling the
  outcome record (D4).
- Parsing `stopReason` to recognize recovery turns (runtime-recovery doctrine §4 remains
  documentation-only).
- The Go Stop handler (`internal/hook/stop.go`), other hook scripts, and their wrappers.
- Cross-session locking of `.moai/state/` when two sessions share one project root. Atomic
  writes are required (REQ-002); a lock is not.
- Changing the hook's 60-second timeout in the settings template.
