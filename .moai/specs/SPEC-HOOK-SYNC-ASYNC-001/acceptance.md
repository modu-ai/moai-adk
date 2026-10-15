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
| AC-HSA-006 | Given the gate enabled and tree scope, When the wrapper runs, Then the shell receipt probe exists in BOTH wrapper surfaces and resolves ALLOW states without the synchronous handler invocation | REQ-HSA-006 | release-blocking | E-L06 | M3 |
| AC-HSA-007 | Given the gate enabled, tree scope, no skip policy, and NO receipt on a reviewable tree (an ALLOW state), When the wrapper runs, Then it completes with exit 0 WITHOUT a synchronous `hook codex-review-entry` invocation (argv-probe count 0 for that argv; the cache-miss background kick `verify codex-review` is allowed and expected on this state); given a fresh-FAIL receipt or any probe failure, When the wrapper runs, Then the synchronous handler is reached (preservation clause — probed in the M3 matrix; see E-L16's classification note) | REQ-HSA-006, REQ-HSA-007 | release-blocking | E-L07 | M3 |
| AC-HSA-008 | Given a fixture tree, When the shell key computation is compared with Go `verify.Key`, Then the outputs are byte-equal (parity test green on `shasum` and `openssl` paths) | REQ-HSA-008 | release-blocking | E-L08 | M3 |
| AC-HSA-009 | Given the keep-sync table, When the settings are re-inventoried after M7, Then the async-true count is exactly baseline 11 + 3 = 14 and no keep-sync registration changed sync state or decision semantics | REQ-HSA-009 | release-blocking | E-L09 | M7 |
| AC-HSA-010 | Given the live `hooks-system.md`, When the timeout section is read, Then 30s is documented as the official DEFAULT with no documented per-event maximum, and the ceiling bullet no longer states a 30s cap | REQ-HSA-010 | release-blocking | E-L10 | M6 |
| AC-HSA-011 | Given the template mirror `hooks-system.md`, When the same section is read, Then it matches AC-HSA-010's corrected wording | REQ-HSA-010 | release-blocking | E-L11 | M6 |
| AC-HSA-012 | Given both settings surfaces, When the two UserPromptSubmit entries are inspected, Then each carries `timeout: 30` | REQ-HSA-011 | release-blocking | E-L12 | M6 |
| AC-HSA-013 | Given M1 executed, When `.moai/reports/t1780/hook-sync-timings.md` is read, Then it carries 10 per-run timings for every synchronous registration and the fixed prescription-order statement | REQ-HSA-012 | release-blocking | E-L13 | M1 |
| AC-HSA-014 | Given real multi-lane load, When UserPromptSubmit hooks run, Then the hard-timeout count observed in the run log is 0 | REQ-HSA-012 | **regression-guard** | E-L14 | M7 observation |
| AC-HSA-015 | Given the wrapper runs on a cache-MISS state (no receipt or stale receipt), When the Stop arm later runs for the same tree state within the SAME kick bucket (marker bucket = epoch/900), Then exactly one background review was kicked (shared kick marker honored; observed via the argv-recording stub's `verify codex-review` line count across entry+Stop for the same tree state and bucket); and Given a fresh pass receipt, When the wrapper runs, Then NO kick occurs | REQ-HSA-006(a), REQ-HSA-006(b) | release-blocking | E-L15 | M3 |
| AC-HSA-016 | Given a card-scope session (`WT-` branch), When the wrapper runs with the gate ON, Then it retains today's behavior — the shell probe is not consulted (no key computation, no snapshot read, no kick) and the Go handler is reached exactly once per prompt | REQ-HSA-007, plan D5 | regression-guard | E-L16 | M3 probe matrix |
| AC-HSA-017 | Given the review gate OFF (or unreadable), When the wrapper runs, Then it exits 0 with no binary resolution | REQ-HSA-005 | regression-guard | E-L18 | M3 probe matrix |

## §D.1 Severity classification

- **S1 (release-blocking, fail the sync gate)**: AC-HSA-005, AC-HSA-006, AC-HSA-007, AC-HSA-008,
  AC-HSA-015. Rationale: a wrong output shape silently drops
  the failure message (async delivers only additionalContext); a wrong probe direction either
  loses REQ-GBN-002's enforcement (ALLOW on a fresh FAIL), kicks where the Go design suppresses
  (fresh pass / skip-policy states), or pays the cold start anyway; a
  near-miss key silently kills the kick arm; a double-kick wastes a 900s review budget.
- **S2 (release-blocking, functional)**: AC-HSA-001..004, AC-HSA-009..013.
- **Regression-guard (per §2.1)**: AC-HSA-014 (RED not re-executable — historical incident);
  AC-HSA-016 (preservation: the criterion pins behavior M3 must not change, so no red delta can
  exist pre-M3 — audit r1 D4's vacuous-red route); AC-HSA-017 (added at audit r1 per D6, same
  preservation shape — its four-element BASELINE cell is pinned green-now and the M3 rewrite
  must keep it green).
- **S3 (advisory)**: none.

## §D.2 Traceability

REQ-HSA-001→AC-HSA-001/002; REQ-HSA-002→AC-HSA-003; REQ-HSA-003→AC-HSA-004/005;
REQ-HSA-004→AC-HSA-009 (the count sweep asserts the entry's sync state unchanged);
REQ-HSA-005→AC-HSA-017 (dedicated AC, added at audit r1 per D6); REQ-HSA-006→AC-HSA-006/007/015;
REQ-HSA-007→AC-HSA-007/016; REQ-HSA-008→AC-HSA-008; REQ-HSA-009→AC-HSA-009;
REQ-HSA-010→AC-HSA-010/011; REQ-HSA-011→AC-HSA-012; REQ-HSA-012→AC-HSA-013/014. Every REQ has at
least one AC; no AC lacks a REQ. AC-HSA-017 is the only id-set addition (audit r1 D6); every
round-0 REQ/AC id is unchanged (delta re-audit eligibility).

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

- Snapshot file present but truncated/unparseable → receipt-ABSENT by the shared store's own
  reading (`verify.LoadReceipt`, `receipt.go:113-117` — every read failure is absence, never an
  error or a valid/failing verdict) → the REQ-HSA-006(a) MISS arm (ALLOW + kick, kick-region
  preconditions permitting) — aligned with REQ-HSA-007 at audit r2-D2.
- `RecordedAt` exactly at the 24h boundary → FRESH (CheckReceipt's comparison is strict `>`,
  `receipt.go:92-94` — measure-zero, wording corrected at audit r3-O3) → arm (b)/(c) by verdict;
  beyond 24h → stale → MISS arm.
- Kick-marker bucket roll between entry and Stop (the epoch/900 boundary crossed) → a second
  review for the same tree state is DESIGNED behavior: the bucket path changes, the exclusive
  create succeeds, and the dead-review re-kick within one budget is the marker's own r7 remedy
  (`codex_review_delay.go:82-86`). AC-HSA-015's exactly-one bound is scoped to one bucket.
- Clean tree or config-only-only tree (nothing-reviewable precondition) → exit 0 WITHOUT the
  early kick; the Stop gate's Go path owns the kick decision (REQ-GBN-002 preserved).
- Binary decably behind the tree HEAD (binary-age precondition) → exit 0 WITHOUT the early kick;
  undecidable age (no commit metadata, non-ancestor) → the early kick proceeds per binlag's
  leniency contract.
- Both sha256 tools absent → probe fails → fall through (REQ-HSA-007); no new shared failure mode.
- Gate enabled but `workflow.yaml` moves → the existing gate-OFF parse yields OFF → exit 0
  (unchanged behavior).
- Empty `codex-review` entry verdict on a fresh receipt → not "fail" → arm (b) (exit 0, no kick).

## §D.7 Quality gates

- `go test ./internal/hook ./internal/cli ./internal/template -count=1` green (E-level evidence
  in the run-phase §E.2).
- Template-registration tests assert the three async flags, the two UPS timeouts, and the
  keep-sync invariance.
- The M3 behavioral probe matrix (§F M3 list) green on macOS; the `openssl` fallback leg green.

## §E Evidence Ledger (RED-now, observed 2026-10-15, tree `e04ebd5ca6bc441b380c1072d5a1dff8431459ba`)

Each entry: command (single-invocation read-only form), verbatim stdout, exit code. The document
tree SHA above binds every entry; a criterion-level re-pin re-measures rather than re-cites.
The two fixture probes (E-L07, E-L18) exercise the wrapper against a `/tmp` fixture project and
a stub binary — they touch no repo-tree state; the fixture build steps are stated inside the
entries so each command is re-executable as cited.

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
229:            "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; d=\"${0%/.claude/hooks/moai/*}\"; mkdir -p \"$d/.moai/logs\" 2>/dev/null; printf \"%s %s\\n\" \"$(date -u +%FT%TZ 2>/dev/null || /bin/date -u +%FT%TZ 2>/dev/null || printf unknown-time)\" \"hook missing: $0\" >> \"$d/.moai/logs/hook-missing.log\" 2>/dev/null; exit 0", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/chain-event.sh"],
230-            "timeout": 5,
231-            "type": "command"
exit=0
```
The template mirror (same command on `settings.json.tmpl`) returned byte-identical lines 229-231,
`exit=0` (the `-c` wrapper args line is byte-identical to E-L01's — only the script basename
differs).

**E-L04** (AC-HSA-004, PostToolUseFailure lacks async in both surfaces):

```
$ grep -n -A 8 "handle-post-tool-failure" .claude/settings.json
241:            "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; d=\"${0%/.claude/hooks/moai/*}\"; mkdir -p \"$d/.moai/logs\" 2>/dev/null; printf \"%s %s\\n\" \"$(date -u +%FT%TZ 2>/dev/null || /bin/date -u +%FT%TZ 2>/dev/null || printf unknown-time)\" \"hook missing: $0\" >> \"$d/.moai/logs/hook-missing.log\" 2>/dev/null; exit 0", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-post-tool-failure.sh"],
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

**E-L07** (AC-HSA-007, RED for the genuinely-red post-M3 delta: an ALLOW state completes WITHOUT
a synchronous handler invocation. Fixture `/tmp/t1780-fp-red/` (reworked at audit r2-D1):
`proj-on/` is a git repo (branch `main` @ 7d70f69 — non-`WT-`, so the M3 key computation can
succeed and the fixture is a reviewable non-card tree) carrying
`.moai/config/sections/workflow.yaml` with `workflow.codex.review_gate.enabled: true` AND
`primary_scope: review` (REQ-CGSC-004 — an explicit review restores reviewability on a
primary-checkout tree, so the post-M3 fast path is eligible here) and no snapshot store (miss
state); `bin/moai` is an ARGV-RECORDING stub — it appends `"$*"` to `bin/moai.argv` and exits 0
— so the probe distinguishes the synchronous handler call (`hook codex-review-entry`, forbidden
on ALLOW states post-M3) from the background kick (`verify codex-review`, allowed and expected
there). Prep: `chmod +x` the stub, `moai.argv` cleared. Cited commands — single invocations,
literal PATH, stdin is the runner's closed stdin):

```
$ env CLAUDE_PROJECT_DIR=/tmp/t1780-fp-red/proj-on PATH="/tmp/t1780-fp-red/bin:/usr/bin:/bin:/usr/sbin:/sbin" bash .claude/hooks/moai/handle-codex-review-entry.sh
(no stdout)
exit=0
$ cat /tmp/t1780-fp-red/bin/moai.argv
hook codex-review-entry
exit=0
$ grep -c "hook codex-review-entry" /tmp/t1780-fp-red/bin/moai.argv
1
exit=0
```

Red because the argv record shows the synchronous handler call on the reviewable no-receipt
(ALLOW) state — today's line-84 unconditional forward. Post-M3 the same probe must show ZERO
`hook codex-review-entry` lines on this state (at most one `verify codex-review` kick line,
which AC-HSA-015 counts). The r1 marker-based form of this entry is superseded: an
all-invocations marker could not distinguish the permitted kick from the forbidden handler call,
and its git-less fixture could not compute the M3 key (both r2-D1 defects — the replacement
cures both). Provenance of the superseded r1 observation retained under Gaps.

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

**E-L16** (AC-HSA-016, regression-guard per §2.1 — audit r1 D4's required disposition). The
criterion is a preservation guard: it pins behavior M3 must NOT change (card-scope sessions
reach the Go handler exactly once per prompt and consult no shell probe state). No red delta can
exist pre-M3 — today the wrapper probes nothing (E-L06) and forwards unconditionally (line 84),
which is precisely the behavior to preserve; a "falls through" RED would be vacuous (D4),
because the fall-through is already true. Baseline evidence: E-L06 + E-L07's invocation probe +
the plan-phase script read showing no `WT-` branch check. Green path: the M3 probe matrix's
`WT-` branch case asserts the preserved behavior on the rewritten wrapper.

**E-L18** (AC-HSA-017, four-element BASELINE cell — regression-guard classification, added at
audit r1 per D6; re-observed at r2 with the argv-recording stub). The criterion preserves the
existing gate-OFF fast path: green by design pre-M3, and the M3 rewrite must keep it green, so
the cell pins the baseline the rewrite is guarded against (not recorded as a pass for the
implementation). Fixture: same `/tmp/t1780-fp-red/` tree, `proj-off/` carrying `enabled: false`
(no git repo needed — the gate parse exits before any git or binary work); `moai.argv` cleared
first.

```
$ env CLAUDE_PROJECT_DIR=/tmp/t1780-fp-red/proj-off PATH="/tmp/t1780-fp-red/bin:/usr/bin:/bin:/usr/sbin:/sbin" bash .claude/hooks/moai/handle-codex-review-entry.sh
(no stdout)
exit=0
$ ls /tmp/t1780-fp-red/bin/moai.argv
ls: /tmp/t1780-fp-red/bin/moai.argv: No such file or directory
exit=1
```

Elements: command (single invocation, literal PATH) / stdout (empty) / exit 0 (wrapper run) with
the argv-file-absence observation (`ls` exit 1) proving no binary resolution / tree
`e04ebd5ca6bc441b380c1072d5a1dff8431459ba` (document pin). The M3 probe matrix re-asserts this
behavior on the rewritten wrapper.

## Gaps (plan-phase observations NOT made)

- Probe R1 attempt 1 was REFUSED by the worktree-isolation guard (`env` with runtime-expanded
  `$PATH`); the observation of record is the literal-PATH retry in E-L07 (refusal recorded per
  `verification-claim-integrity.md` §3.1).
- The r1-form E-L07/E-L18 observations (marker stub `moai.invoked`, git-less fixture) are
  SUPERSEDED by the r2 fixture rework (argv stub, git-init'd reviewable fixture, explicit
  `primary_scope: review`) per audit r2-D1; their outputs are retained in the r1-authored
  version of this file in git history, not re-cited.
- The local primary checkout's settings (leader-raised 30) was taken from the card directive, not
  re-measured from the primary checkout (this session is worktree-isolated). M2 re-observes all
  three surfaces on the absorbed tree.

## Definition of Done

1. All ACs GREEN (AC-HSA-014 per its regression-guard closure rule) with ledger citations
   re-measured on the landing tree.
2. M1 measurement table committed at `.moai/reports/t1780/hook-sync-timings.md`.
3. Live/template pairs land together (C2); t1775 gate recorded before M3 (C1).
4. Run-phase §E evidence complete (E1-E7 of plan.md §E).
