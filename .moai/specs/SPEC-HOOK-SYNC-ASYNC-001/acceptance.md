---
id: SPEC-HOOK-SYNC-ASYNC-001
title: "Acceptance criteria and evidence ledger — hook sync/async cleanup"
version: "1.0.0"
created: 2026-10-15
author: manager-spec
tier: M
---

# acceptance.md — SPEC-HOOK-SYNC-ASYNC-001

AC format: every criterion is a labeled `AC-HSA-XXX` Given-When-Then, adopted as a two-cell pair
per `.claude/rules/moai/development/verification-completeness.md` §2 — a **RED-now** cell
(observed red on the pre-implementation tree, pinned to tree
`e04ebd5ca6bc441b380c1072d5a1dff8431459ba`, document-level pin binding every criterion without
its own pin) and a **green path** cell naming the milestone that flips it. Release-blocking
criteria carry four-element RED-now cells (command / verbatim stdout / exit code / tree SHA) via
the evidence ledger in §E; the ledger id is cited from the AC row.

## §D AC Matrix

| AC | Criterion (Given / When / Then) | REQ | Class | RED ledger | Green path |
|----|----------------------------------|-----|-------|------------|------------|
| AC-HSA-001 | Given the template settings, When the ConfigChange entry is inspected, Then it carries `"async": true` | REQ-HSA-001 | release-blocking | E-L01 | M5 |
| AC-HSA-002 | Given the live settings, When the ConfigChange entry is inspected, Then it carries `"async": true` | REQ-HSA-001 | release-blocking | E-L02 | M5 |
| AC-HSA-003 | Given both settings surfaces, When the SubagentStop chain-event entry is inspected, Then it carries `"async": true` | REQ-HSA-002 | release-blocking | E-L03 | M5 |
| AC-HSA-004 | Given both settings surfaces, When the PostToolUseFailure entry is inspected, Then it carries `"async": true` | REQ-HSA-003 | release-blocking | E-L04 | M4 |
| AC-HSA-005 | Given the Go handler, When a PostToolUseFailure event is handled, Then the output carries `additionalContext` and no `systemMessage` | REQ-HSA-003 | release-blocking | E-L05 | M4 |
| AC-HSA-006 | Given the gate enabled and tree scope, When the wrapper runs, Then the shell receipt probe exists in BOTH wrapper surfaces and resolves ALLOW states without the binary | REQ-HSA-006 | release-blocking | E-L06 | M3 |
| AC-HSA-007 | Given a fresh-FAIL receipt, When the wrapper runs, Then it falls through to `moai hook codex-review-entry`; given any probe failure, When the wrapper runs, Then it degrades to the current behavior (fall-through or exit 0), never a block | REQ-HSA-006, REQ-HSA-007 | release-blocking | E-L07 | M3 |
| AC-HSA-008 | Given a fixture tree, When the shell key computation is compared with Go `verify.Key`, Then the outputs are byte-equal (parity test green on `shasum` and `openssl` paths) | REQ-HSA-008 | release-blocking | E-L08 | M3 |
| AC-HSA-009 | Given the keep-sync table, When the settings are re-inventoried after M7, Then the async-true count is exactly baseline 11 + 3 = 14 and no keep-sync registration changed sync state or decision semantics | REQ-HSA-009 | release-blocking | E-L09 | M7 |
| AC-HSA-010 | Given the live `hooks-system.md`, When the timeout section is read, Then 30s is documented as the official DEFAULT with no documented per-event maximum, and the ceiling bullet no longer states a 30s cap | REQ-HSA-010 | release-blocking | E-L10 | M6 |
| AC-HSA-011 | Given the template mirror `hooks-system.md`, When the same section is read, Then it matches AC-HSA-010's corrected wording | REQ-HSA-010 | release-blocking | E-L11 | M6 |
| AC-HSA-012 | Given both settings surfaces, When the two UserPromptSubmit entries are inspected, Then each carries `timeout: 30` | REQ-HSA-011 | release-blocking | E-L12 | M6 |
| AC-HSA-013 | Given M1 executed, When `.moai/reports/t1780/hook-sync-timings.md` is read, Then it carries 10 per-run timings for every synchronous registration and the fixed prescription-order statement | REQ-HSA-012 | release-blocking | E-L13 | M1 |
| AC-HSA-014 | Given real multi-lane load, When UserPromptSubmit hooks run, Then the hard-timeout count observed in the run log is 0 | REQ-HSA-012 | **regression-guard** | E-L14 | M7 observation |
| AC-HSA-015 | Given the wrapper runs on an ALLOW state, When the Stop arm later runs for the same tree state, Then exactly one background review was kicked (shared kick marker honored) | REQ-HSA-006(a) | release-blocking | E-L15 | M3 |
| AC-HSA-016 | Given a card-scope session (`WT-` branch), When the wrapper runs with the gate ON, Then it falls through to the Go handler (no shell probe result is trusted for card scope) | REQ-HSA-007, D5 | release-blocking | E-L16 | M3 |

## §D.1 Severity classification

- **S1 (release-blocking, fail the sync gate)**: AC-HSA-005, AC-HSA-006, AC-HSA-007, AC-HSA-008,
  AC-HSA-014's underlying evidence, AC-HSA-015. Rationale: a wrong output shape silently drops
  the failure message (async delivers only additionalContext); a wrong probe direction either
  loses REQ-GBN-002's enforcement (ALLOW on a fresh FAIL) or pays the cold start anyway; a
  near-miss key silently kills the kick arm; a double-kick wastes a 900s review budget; a prompt
  block regression is user-facing.
- **S2 (release-blocking, functional)**: AC-HSA-001..004, AC-HSA-009..013, AC-HSA-016.
- **S3 (advisory)**: none.

## §D.2 Traceability

REQ-HSA-001→AC-001/002; REQ-HSA-002→AC-003; REQ-HSA-003→AC-004/005; REQ-HSA-004→kept by
construction (AC-009 sweep asserts the entry's sync state unchanged); REQ-HSA-005→covered by the
existing gate-OFF tests, re-asserted in the M3 probe matrix; REQ-HSA-006→AC-006/007/015;
REQ-HSA-007→AC-007/016; REQ-HSA-008→AC-008; REQ-HSA-009→AC-009; REQ-HSA-010→AC-010/011;
REQ-HSA-011→AC-012; REQ-HSA-012→AC-013/014. Every REQ has at least one AC; no AC lacks a REQ.

## §D.3 Indirect verification

- The kick marker's shared use is verified indirectly: AC-HSA-015 observes one review per tree
  state across entry+Stop triggers rather than parsing marker contents.
- The async delivery contract is upstream; AC-HSA-005 verifies the emitted JSON shape (the part
  this codebase owns), and the M4 completion report names the upstream delivery semantics as a
  documented dependency, not a locally verifiable claim.

## §D.4 Closure gates

- All S1 ACs GREEN with ledger citations re-measured on the landing tree (re-pinned if the absorb
  moved the base — §4 re-measure rule).
- AC-HSA-014 remains regression-guard: it is NOT recorded as a pass from a single observation;
  its green path is the M7 observation under real multi-lane load, recorded with the session log
  excerpt.

## §D.5 Forward-looking checks

- The M1 measurement table may surface further async/timeout candidates — these become follow-up
  cards citing the table row; they do not reopen this SPEC's scope.

## §D.6 Edge cases

- Snapshot file present but truncated/unparseable → treated as receipt-absent (ALLOW arm), never
  an error exit.
- `RecordedAt` exactly at the 24h boundary → stale (CheckReceipt uses `>` ttl; mirror the same
  comparison).
- Both sha256 tools absent → probe fails → fall through (REQ-HSA-007); no new shared failure mode.
- Gate enabled but `workflow.yaml` moves → the existing gate-OFF parse yields OFF → exit 0
  (unchanged behavior).
- Empty `codex-review` entry verdict → not "fail" → ALLOW arm.

## §D.7 Quality gates

- `go test ./internal/hook ./internal/cli ./internal/template -count=1` green (E-level evidence
  in the run-phase §E.2).
- Template-registration tests assert the three async flags, the two UPS timeouts, and the
  keep-sync invariance.
- The M3 behavioral probe matrix (§F M3 list) green on macOS; the `openssl` fallback leg green.

## §E Evidence Ledger (RED-now, observed 2026-10-15, tree `e04ebd5ca6bc441b380c1072d5a1dff8431459ba`)

Each entry: command (single-invocation read-only form), verbatim stdout, exit code. The document
tree SHA above binds every entry; a criterion-level re-pin re-measures rather than re-cites.

**E-L01** (AC-HSA-001, template ConfigChange lacks async — red because the printed entry's fields
end at `timeout`/`type` with no `async` key):

```
$ grep -n -A 8 "handle-config-change" internal/template/templates/.claude/settings.json.tmpl
315:            "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; d=\"${0%/.claude/hooks/moai/*}\"; mkdir -p \"$d/.moai/logs\" 2>/dev/null; printf \"%s %s\\n\" \"$(date -u +%FT%TZ 2>/dev/null || /bin/date -u +%FT%TZ 2>/dev/null || printf unknown-time)\" \"hook missing: $0\" >> \"$d/.moai/logs/hook-missing.log\" 2>/dev/null; exit 0", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-config-change.sh"],
316-            "timeout": 5,
317-            "type": "command"
exit=0
```

**E-L02** (AC-HSA-002, live ConfigChange lacks async): same command form on
`.claude/settings.json` — verbatim output identical to E-L01 (lines 315-317, entry at 315,
`timeout: 5`, `type: command`, no `async`), `exit=0`.

**E-L03** (AC-HSA-003, chain-event lacks async in both surfaces — red same shape):

```
$ grep -n -A 8 "chain-event" .claude/settings.json
229:            "args": [...same -c wrapper... , "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/chain-event.sh"],
230-            "timeout": 5,
231-            "type": "command"
exit=0
```
The template mirror (same command on `settings.json.tmpl`) returned byte-identical lines 229-231,
`exit=0`.

**E-L04** (AC-HSA-004, PostToolUseFailure lacks async in both surfaces):

```
$ grep -n -A 8 "handle-post-tool-failure" .claude/settings.json
241:            "args": [...same -c wrapper..., "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-post-tool-failure.sh"],
242-            "timeout": 5,
243-            "type": "command"
exit=0
```
Template mirror: byte-identical lines 241-243, `exit=0`.

**E-L05** (AC-HSA-005, handler emits systemMessage today — red because the criterion requires
additionalContext and the source returns SystemMessage; the current shape is test-pinned, so the
M4 change must update those tests in the same commit):

```
$ grep -n "SystemMessage: message" internal/hook/post_tool_failure.go
109:		SystemMessage: message,
exit=0
$ go test ./internal/hook -run "TestPostToolUseFailure" -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	1.672s
exit=0
```

**E-L06** (AC-HSA-006, no shell receipt probe in either wrapper — empty stdout + non-zero exit is
the complete observation):

```
$ grep -n "snapshots" .claude/hooks/moai/handle-codex-review-entry.sh
exit=1
$ grep -n "snapshots" internal/template/templates/.claude/hooks/moai/handle-codex-review-entry.sh
exit=1
```

**E-L07** (AC-HSA-007, the wrapper forwards unconditionally today — every gate-ON prompt pays the
binary; no probe, no fall-through logic): E-L06's absence plus the wrapper's line 84
(`printf '%s' "$INPUT" | "$MOAI_BIN" hook codex-review-entry` being the only gate-ON path,
observed in the script read at plan phase).

**E-L08** (AC-HSA-008, no key-parity test exists — the fixture and test are M3 deliverables):

```
$ grep -rn "parity" internal/cli/codex_review_entry_fastpath_test.go
exit=2  (No such file or directory)
```

**E-L09** (AC-HSA-009, async baseline the sweep protects — count 11 today; target 14):

```
$ grep -c "async.: true" .claude/settings.json
11
exit=0
```

**E-L10** (AC-HSA-010, live doctrine states a cap — red because the official semantics is
default-not-cap):

```
$ grep -n "reduced max\|15s" .claude/rules/moai/core/hooks-system.md
368:| UserPromptSubmit | 15s | 30s | Blocks user interaction; reduced max. Raised from 5s: the factory-messaging drain in handle-user-prompt-submit can exceed 5s under multi-lane load, and the hard-timeout error surfaced to the user |
exit=0
$ grep -n "lowers" .claude/rules/moai/core/hooks-system.md
374:- **`UserPromptSubmit`** lowers `command` / `http` / `mcp_tool` hooks to **30s**, because the hook blocks the user's own input.
375:- **`MessageDisplay`** lowers them to **10s**, because it runs while assistant text is being displayed.
exit=0
```

**E-L11** (AC-HSA-011, template doctrine identical): same two command forms on
`internal/template/templates/.claude/rules/moai/core/hooks-system.md` — byte-identical lines 368
and 374-375, `exit=0` both.

**E-L12** (AC-HSA-012, template UPS timeouts at 5; this worktree's live UPS-1 at 15 — red because
the criterion requires 30):

```
$ grep -n -A 6 "handle-user-prompt-submit.sh" internal/template/templates/.claude/settings.json.tmpl
266-            "timeout": 5,
$ grep -n -A 6 "handle-user-prompt-submit.sh" .claude/settings.json
266-            "timeout": 15,
exit=0
```
(the codex-review-entry UPS entry: template `timeout: 5` at line 272; this worktree's live value
observed 5 in the same read.)

**E-L13** (AC-HSA-013, measurement table absent):

```
$ ls .moai/reports/t1780/hook-sync-timings.md
ls: .moai/reports/t1780/hook-sync-timings.md: No such file or directory
exit=1
```

**E-L14** (AC-HSA-014, regression-guard classification). The RED cannot be re-executed on the
current tree: the motivating incident is historical (the hard-timeout error observed by the
operator under multi-lane load, recorded in the hooks-system.md table row note — E-L10's line
368 text), and the card forbids artificial load. Per verification-completeness §2.1 this
criterion loses release-blocking RED-now eligibility and is classified regression-guard; it is
NOT recorded as a pass at plan phase. Green path: the M7 observation under real multi-lane load.

**E-L15** (AC-HSA-015, no kick logic exists in the wrapper today — the only "kick" occurrence is
a comment):

```
$ grep -n "kick" .claude/hooks/moai/handle-codex-review-entry.sh
7:# and kicks a background review; this wrapper is the entry-side half: it reads
exit=0
```

**E-L16** (AC-HSA-016, no scope pre-check exists in the wrapper — the fall-through-for-card-scope
behavior is new; RED evidenced by E-L06's probe absence plus the script read showing no `WT-`
branch check).

## Gaps (plan-phase observations NOT made)

- The local primary checkout's settings (leader-raised 30) was taken from the card directive, not
  re-measured from the primary checkout (this session is worktree-isolated). M2 re-observes all
  three surfaces on the absorbed tree.

## Definition of Done

1. All ACs GREEN (AC-HSA-014 per its regression-guard closure rule) with ledger citations
   re-measured on the landing tree.
2. M1 measurement table committed at `.moai/reports/t1780/hook-sync-timings.md`.
3. Live/template pairs land together (C2); t1775 gate recorded before M3 (C1).
4. Run-phase §E evidence complete (E1-E7 of plan.md §E).
