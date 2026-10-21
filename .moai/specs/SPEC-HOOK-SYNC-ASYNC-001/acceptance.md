---
id: SPEC-HOOK-SYNC-ASYNC-001
title: "Acceptance criteria and evidence ledger — hook sync/async cleanup"
version: "1.0.0"
created: 2026-10-15
updated: 2026-10-22
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
| AC-HSA-007 | Given the gate enabled, tree scope, no skip policy, and NO receipt on a reviewable tree (an ALLOW state), When the wrapper runs, Then it completes with exit 0 WITHOUT a synchronous `hook codex-review-entry` invocation (argv-probe count 0 for that argv; the cache-miss background kick `verify codex-review` is allowed and expected on this state); given a fresh-FAIL receipt or any probe failure, When the wrapper runs, Then the synchronous handler is reached (preservation clause — probed in the M3 matrix; see E-L16's classification note) | REQ-HSA-006, REQ-HSA-007, REQ-HSA-008 | release-blocking | E-L07 | M3 |
| AC-HSA-008 | Given a fixture tree, When the shell key computation is compared with Go `verify.Key`, Then the outputs are byte-equal (parity test green on `shasum` and `openssl` paths) | REQ-HSA-008 | release-blocking | E-L08 | M3 |
| AC-HSA-009 | Given the keep-sync table, When the settings are re-inventoried after M7, Then the async-true count is exactly baseline 11 + 3 = 14 and no keep-sync registration changed sync state or decision semantics | REQ-HSA-009 | release-blocking | E-L09 | M7 |
| AC-HSA-010 | Given the live `hooks-system.md`, When the timeout section is read, Then 30s is documented as the official DEFAULT with no documented per-event maximum, and the ceiling bullet no longer states a 30s cap | REQ-HSA-010 | release-blocking | E-L10 | M6 |
| AC-HSA-011 | Given the template mirror `hooks-system.md`, When the same section is read, Then it matches AC-HSA-010's corrected wording | REQ-HSA-010 | release-blocking | E-L11 | M6 |
| AC-HSA-012 | Given both settings surfaces, When the two UserPromptSubmit entries are inspected, Then each carries `timeout: 30` | REQ-HSA-011 | release-blocking | E-L12 | M6 |
| AC-HSA-013 | Given M1 executed, When `.moai/reports/t1780/hook-sync-timings.md` is read, Then it carries 10 per-run timings for every synchronous registration and the fixed prescription-order statement | REQ-HSA-012 | release-blocking | E-L13 | M1 |
| AC-HSA-014 | Given real multi-lane load, When UserPromptSubmit hooks run, Then the hard-timeout count observed in the run log is 0 | REQ-HSA-012 | **regression-guard** | E-L14 | M7 observation |
| AC-HSA-015 | Given the wrapper runs on a cache-MISS state (no receipt or stale receipt), When the Stop arm later runs for the same tree state within the SAME kick bucket (marker bucket = epoch/900), Then exactly one background review was kicked (shared kick marker honored; observed via the argv-recording stub's `verify codex-review` line count across entry+Stop for the same tree state and bucket); and Given a fresh pass receipt, When the wrapper runs, Then NO kick occurs | REQ-HSA-006(a), REQ-HSA-006(b), REQ-HSA-008 | release-blocking | E-L15 | M3 |
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
REQ-HSA-007→AC-HSA-007/016; REQ-HSA-008→AC-HSA-007/008/015; REQ-HSA-009→AC-HSA-009;
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

- Clock portability (AC-HSA-008/015): Given an awk implementation whose `srand()` seed is
  nanosecond-scale, When the wrapper reads an actual fresh receipt or an existing Stop-arm
  marker, Then it uses validated system Unix seconds and retains the existing fresh-verdict
  behavior and exactly-one kick bound within the same epoch/900 bucket. The random seed
  cannot make a fresh receipt stale or create a different bucket.
- Clock unavailable (AC-HSA-007): Given the system `date +%s` probe is missing, fails, or
  returns a non-decimal value, When the gate-ON tree-scope wrapper runs, Then the Go handler
  is reached under REQ-HSA-007; the shell emits no synthetic verdict and does not infer time
  from an awk seed. These cases extend the probe-failure preservation clause, not the set of
  states allowed to skip enforcement.
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

**GREEN cells (M7 close, added 2026-10-22).** Every `GREEN` block below carries its command, the
verbatim output observed in the M7 run, the exit code, and the landing observation tree
`703654daca` (the M7 doctrine touch-up commit; the ledger evidence commit follows it). Test
commands ran under the env scrub
`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && TMPDIR=/tmp`.
The single-invocation command form is a RED-cell constraint (§2.1); GREEN cells cite the
command exactly as run.

**E-L01** (AC-HSA-001, template ConfigChange lacks async — red because the printed entry's fields
end at `timeout`/`type` with no `async` key):

```
$ grep -n -A 8 "handle-config-change" internal/template/templates/.claude/settings.json.tmpl
315:            "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; d=\"${0%/.claude/hooks/moai/*}\"; mkdir -p \"$d/.moai/logs\" 2>/dev/null; printf \"%s %s\\n\" \"$(date -u +%FT%TZ 2>/dev/null || /bin/date -u +%FT%TZ 2>/dev/null || printf unknown-time)\" \"hook missing: $0\" >> \"$d/.moai/logs/hook-missing.log\" 2>/dev/null; exit 0", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-config-change.sh"],
316-            "timeout": 5,
317-            "type": "command"
exit=0
```

**E-L01 GREEN** (flag landed at M5, commit 4a0fa991ed; re-pinned at M7, tree `703654daca` — the
template ConfigChange entry now ends in `async: true`; the entry anchor is the line-317 args
line, the same line the RED cell above pinned):

```
$ awk 'NR>=318 && NR<=320 {print NR": "$0}' internal/template/templates/.claude/settings.json.tmpl
318:             "timeout": 5,
319:             "type": "command",
320:             "async": true
exit=0
```

**E-L02** (AC-HSA-002, live ConfigChange lacks async): same command form on
`.claude/settings.json` — verbatim output identical to E-L01 (lines 315-317, entry at 315,
`timeout: 5`, `type: command`, no `async`), `exit=0`.

**E-L02 GREEN** (M5; re-pinned at M7, tree `703654daca` — the live ConfigChange entry mirrors the
template byte-for-byte at the same line numbers):

```
$ awk 'NR>=318 && NR<=320 {print NR": "$0}' .claude/settings.json
318:             "timeout": 5,
319:             "type": "command",
320:             "async": true
exit=0
```

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

**E-L03 GREEN** (M5; re-pinned at M7, tree `703654daca` — chain-event carries `async: true` in
BOTH surfaces at identical lines 230-232; `exit=0` both):

```
$ awk 'NR>=230 && NR<=232 {print NR": "$0}' .claude/settings.json
230:             "timeout": 5,
231:             "type": "command",
232:             "async": true
$ awk 'NR>=230 && NR<=232 {print NR": "$0}' internal/template/templates/.claude/settings.json.tmpl
230:             "timeout": 5,
231:             "type": "command",
232:             "async": true
```

**E-L04** (AC-HSA-004, PostToolUseFailure lacks async in both surfaces):

```
$ grep -n -A 8 "handle-post-tool-failure" .claude/settings.json
241:            "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; d=\"${0%/.claude/hooks/moai/*}\"; mkdir -p \"$d/.moai/logs\" 2>/dev/null; printf \"%s %s\\n\" \"$(date -u +%FT%TZ 2>/dev/null || /bin/date -u +%FT%TZ 2>/dev/null || printf unknown-time)\" \"hook missing: $0\" >> \"$d/.moai/logs/hook-missing.log\" 2>/dev/null; exit 0", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-post-tool-failure.sh"],
242-            "timeout": 5,
243-            "type": "command"
exit=0
```
Template mirror: byte-identical lines 241-243, `exit=0`.

**E-L04 GREEN** (M4, commit a29a17cbeb; re-pinned at M7, tree `703654daca` — PostToolUseFailure
carries `async: true` in BOTH surfaces at identical lines 243-245; `exit=0` both):

```
$ awk 'NR>=243 && NR<=245 {print NR": "$0}' .claude/settings.json
243:             "timeout": 5,
244:             "type": "command",
245:             "async": true
$ awk 'NR>=243 && NR<=245 {print NR": "$0}' internal/template/templates/.claude/settings.json.tmpl
243:             "timeout": 5,
244:             "type": "command",
245:             "async": true
```

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

**E-L05 GREEN** (M4 flipped the emission shape — commit a29a17cbeb; re-pinned at M7, tree
`703654daca`):

```
$ go test ./internal/hook/ -run 'TestPostToolUseFailureHandler_AdditionalContextNotSystemMessage' -count=1 -v
--- PASS: TestPostToolUseFailureHandler_AdditionalContextNotSystemMessage (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.504s
exit=0
```
(the pin asserts additionalContext present with the category prefix, systemMessage absent,
hookEventName PostToolUseFailure; the full TestPostToolUseFailure family also green —
`ok ... 0.639s`.)

**E-L06** (AC-HSA-006, no shell receipt probe in either wrapper — empty stdout + non-zero exit is
the complete observation):

```
$ grep -n "snapshots" .claude/hooks/moai/handle-codex-review-entry.sh
exit=1
$ grep -n "snapshots" internal/template/templates/.claude/hooks/moai/handle-codex-review-entry.sh
exit=1
```

**E-L06 GREEN** (M3 — commits 772b9a1590+3483d65726; re-pinned at M7, tree `703654daca` — the
probe exists in BOTH wrapper surfaces and live↔template byte equality is test-pinned):

```
$ go test ./internal/cli/ -run 'TestCodexEntryFastpath|TestHookEntryParity' -count=1 -v
--- PASS: TestCodexEntryFastpath_BothSurfacesCarryProbe (0.00s)
--- PASS: TestHookEntryParity_NoDivergenceEitherDirection (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	17.932s
exit=0
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

**E-L07 GREEN** (M3; re-pinned at M7, tree `703654daca` — the no-receipt ALLOW state now completes
WITHOUT the synchronous handler: `forward=0 kick=1`, the exact inverse of the RED observation
above):

```
$ go test ./internal/cli/ -run 'TestCodexEntryFastpath|TestHookEntryParity' -count=1 -v
=== RUN   TestCodexEntryFastpath_ProbeMatrix/no_receipt_on_a_reviewable_tree:_ALLOW_with_one_kick,_no_forward
--- PASS: TestCodexEntryFastpath_ProbeMatrix (13.10s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	17.932s
exit=0
```
(the ProbeMatrix ran all 18 states green in this M7 run — gate off, no receipt, stale receipt,
unparseable snapshot, fresh pass, fresh inconclusive, fresh FAIL, clean tree, config-only-only,
behind binary, HEAD-match binary, no-commit-metadata binary, missing git, missing shasum, both
sha tools absent, WT- branch, tree_scope skip, primary-checkout session.)

**E-L08** (AC-HSA-008, no key-parity test exists — the fixture and test are M3 deliverables):

```
$ grep -rn "parity" internal/cli/codex_review_entry_fastpath_test.go
exit=2  (No such file or directory)
```

**E-L08 GREEN** (M3; re-pinned at M7, tree `703654daca` — shell digest byte-equal to `verify.Key`
on both the `shasum` and the `openssl` pipeline):

```
$ go test ./internal/cli/ -run 'TestCodexEntryFastpath|TestHookEntryParity' -count=1 -v
--- PASS: TestCodexEntryFastpath_KeyParity (0.60s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	17.932s
exit=0
```

**E-L09** (AC-HSA-009, async baseline the sweep protects — count 11 today; target 14):

```
$ grep -c "async.: true" .claude/settings.json
11
exit=0
```

**E-L09 GREEN** (M7 — this sweep IS the AC-HSA-009 observation, tree `703654daca`. Count 14/14,
exactly baseline 11 + 3):

```
$ grep -c "async.: true" .claude/settings.json
14
exit=0
$ grep -c "async.: true" internal/template/templates/.claude/settings.json.tmpl
14
exit=0
```

Diff-based no-keep-sync-change evidence against the card directive's baseline `8249bff8c3` — the
settings delta is EXACTLY the three async additions and nothing else (3 hunks, 6 insertions,
3 deletions; every added line is a `"type": "command",` / `"async": true` pair, so no keep-sync
registration changed sync state, timeout, or command):

```
$ git diff 8249bff8c3..HEAD --stat -- .claude/settings.json
 .claude/settings.json | 9 ++++++---
 1 file changed, 6 insertions(+), 3 deletions(-)
$ git diff 8249bff8c3..HEAD -- .claude/settings.json | grep -c "^@@"
3
```

Async-entry inventory (both surfaces, identical multisets — 14 entries across 12 wrappers,
status-transition-ownership.sh carrying 3 PostToolUse entries):

```
$ grep -B 4 '"async": true' .claude/settings.json | grep -o '[a-z0-9-]*\.sh' | sort | uniq -c
   1 chain-event.sh
   1 handle-config-change.sh
   1 handle-harness-observe-stop.sh
   1 handle-harness-observe-subagent-stop.sh
   1 handle-harness-observe-user-prompt-submit.sh
   1 handle-harness-observe.sh
   1 handle-instructions-loaded.sh
   1 handle-post-tool-failure.sh
   1 handle-post-tool.sh
   1 handle-security-commit.sh
   1 handle-security-turn.sh
   3 status-transition-ownership.sh
```

Reading against the REQ-HSA-009 keep-sync table: baseline 11 = PostToolUse 4 (handle-post-tool +
status-transition-ownership ×3) + harness-observe 4 + Stop security 2 + InstructionsLoaded 1;
the additions are exactly the three conversions (chain-event.sh, handle-config-change.sh,
handle-post-tool-failure.sh). NO keep-sync-table wrapper (handle-pre-tool, permission handlers,
handle-stop, sync-phase-quality-gate, handle-subagent-stop, handle-user-prompt-submit,
session-start pair, handle-compact, handle-post-compact, handle-stop-failure, cwd/file-changed,
session-end, teammate-idle/task-completed) appears in the async list on either surface. Base
archaeology: `git show 8249bff8c3:.claude/settings.json | grep -c "async.: true"` = 11 — the
baseline the sweep protects, observed this run.

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

**E-L10 GREEN** (M6 content rode the inherited base — `git show 8249bff8c3:.claude/rules/moai/core/hooks-system.md`
already carries the corrected row, observed this run; re-pinned at M7, tree `703654daca` — the
live doctrine states default-not-cap, with the ceiling bullet restated as default):

```
$ sed -n '369p' .claude/rules/moai/core/hooks-system.md
| UserPromptSubmit | 30s | not documented | Blocks user interaction. Default 30s for this event (Claude Code's default; the official hooks reference documents no separate maximum). Raised from 5s: the factory-messaging drain in handle-user-prompt-submit can exceed 5s under multi-lane load, and the hard-timeout error surfaced to the user |
exit=0
$ sed -n '373,375p' .claude/rules/moai/core/hooks-system.md
Two events have a shorter timeout than the 600s maximum. For `MessageDisplay` that value is a runtime ceiling and is not overridable upward; for `UserPromptSubmit` 30s is the Claude Code default, not a ceiling:

- **`UserPromptSubmit`** command, `http`, and `mcp_tool` hooks default to **30s**, because the hook blocks the user's own input.
exit=0
```

**E-L11** (AC-HSA-011, template doctrine identical): same two command forms on
`internal/template/templates/.claude/rules/moai/core/hooks-system.md` — byte-identical lines 368
and 374-375, `exit=0` both.

**E-L11 GREEN** (re-pinned at M7, tree `703654daca` — the template mirror carries the identical
corrected wording at the same lines 369 and 373-375, observed this run; the whole live↔template
pair is additionally byte-equal via `cmp` exit 0 after the M7 touch-up commit):

```
$ sed -n '369p' internal/template/templates/.claude/rules/moai/core/hooks-system.md
| UserPromptSubmit | 30s | not documented | Blocks user interaction. Default 30s for this event (Claude Code's default; the official hooks reference documents no separate maximum). Raised from 5s: the factory-messaging drain in handle-user-prompt-submit can exceed 5s under multi-lane load, and the hard-timeout error surfaced to the user |
exit=0
$ sed -n '373,375p' internal/template/templates/.claude/rules/moai/core/hooks-system.md
Two events have a shorter timeout than the 600s maximum. For `MessageDisplay` that value is a runtime ceiling and is not overridable upward; for `UserPromptSubmit` 30s is the Claude Code default, not a ceiling:

- **`UserPromptSubmit`** command, `http`, and `mcp_tool` hooks default to **30s**, because the hook blocks the user's own input.
exit=0
```

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

**E-L12 GREEN** (M6 alignment rode the inherited base — `8249bff8c3` already carried both UPS
entries at 30, observed via `git show` this run; re-pinned at M7, tree `703654daca` — both UPS
registrations carry `timeout: 30` in BOTH surfaces. Entry anchors from the M7 grep: args at
lines 267/273 (`handle-user-prompt-submit.sh` / `handle-codex-review-entry.sh`), timeouts at
268/274):

```
$ awk 'NR==268 || NR==274 {print NR": "$0}' .claude/settings.json
268:             "timeout": 30,
274:             "timeout": 30,
$ awk 'NR==268 || NR==274 {print NR": "$0}' internal/template/templates/.claude/settings.json.tmpl
268:             "timeout": 30,
274:             "timeout": 30,
```
(exit=0 both)

**E-L13** (AC-HSA-013, measurement table absent):

```
$ ls .moai/reports/t1780/hook-sync-timings.md
ls: .moai/reports/t1780/hook-sync-timings.md: No such file or directory
exit=1
```

**E-L13 GREEN** (M1 executed; table finalized at M7 with the prescription-order statement, tree
`703654daca`. The path is gitignored (`.gitignore:235` — `.moai/reports/*`) and is CITED, never
committed, per the card directive):

```
$ ls .moai/reports/t1780/hook-sync-timings.md
.moai/reports/t1780/hook-sync-timings.md
exit=0
$ awk -F'|' 'NR>=12 && NR<=40 {print gsub(/exit:/,"exit:")}' .moai/reports/t1780/hook-sync-timings.md | sort | uniq -c
  29 10
$ grep -c "fast-path first, (2) timeout raise" .moai/reports/t1780/hook-sync-timings.md
1
exit=0
```
(29 registration rows × exactly 10 per-run timings each = the table's own 290-invocation footer;
the fixed prescription-order statement (fast-path → timeout raise → async, per REQ-HSA-012) is
present once in the M7 finalization section. Note: rows 19/20 record the M1-era UPS timeouts
(15/5) — a faithful pre-M6 baseline record at the M1 tree `ec6db3d821`, superseded on the landing
tree by E-L12 GREEN's 30/30.)

**E-L14** (AC-HSA-014, regression-guard classification). The RED cannot be re-executed on the
current tree: the motivating incident is historical (the hard-timeout error observed by the
operator under multi-lane load, recorded in the hooks-system.md table row note — E-L10's line
368 text), and the card forbids artificial load. Per verification-completeness §2.1 this
criterion loses release-blocking RED-now eligibility and is classified regression-guard; it is
NOT recorded as a pass at plan phase. Green path: the M7 observation under real multi-lane load.

**E-L14 GREEN** (M7 observation — regression-guard per §D.4, recorded as an observation NOT a
release-blocking pass, tree `703654daca`. Method: this worktree's hook execution trace log
`.moai/logs/trace-*.jsonl` — the moai hook dispatcher's per-invocation records of the form
`{"event":"UserPromptSubmit","handler":"*hook.userPromptSubmitHandler","duration_ms":...}` —
covering this card's real sessions 2026-10-21 00:14 → M7 close):

```
$ grep '"event":"UserPromptSubmit"' .moai/logs/trace-37a7b7e9-b2f1-47ad-84d9-40e7b2ff35b1.jsonl .moai/logs/trace-fd7e2cd0-25c5-4af2-8ec7-63ad5050a764.jsonl | grep -c "duration_ms"
119
$ grep -o '"event":"UserPromptSubmit","handler":"[^"]*","duration_ms":[0-9]*' .moai/logs/trace-37a7b7e9-b2f1-47ad-84d9-40e7b2ff35b1.jsonl .moai/logs/trace-fd7e2cd0-25c5-4af2-8ec7-63ad5050a764.jsonl | grep -o '[0-9]*$' | sort -n | tail -1
281
$ grep -c "timed out\|hard.timeout\|timeout_error" .moai/logs/trace-37a7b7e9-b2f1-47ad-84d9-40e7b2ff35b1.jsonl .moai/logs/trace-fd7e2cd0-25c5-4af2-8ec7-63ad5050a764.jsonl
.moai/logs/trace-37a7b7e9-b2f1-47ad-84d9-40e7b2ff35b1.jsonl:0
.moai/logs/trace-fd7e2cd0-25c5-4af2-8ec7-63ad5050a764.jsonl:0
```

Observation: 119 UserPromptSubmit handler executions, max duration 281ms against the 30s
registration, zero timeout or error records. The count is a snapshot — the trace files grow live
with the session (an earlier pass in the same M7 run counted 79; the max was 281ms in both).
Gap (coverage boundary): these traces cover THIS worktree's sessions only; the incident-period
multi-lane load ran in the primary checkout and other card trees, whose logs are outside this
tree and uncitable from here. Per the documented dependency (the regression-guard classification
above and plan §D D6) the zero count MUST NOT be read as certifying the multi-lane population.

**E-L15** (AC-HSA-015, no kick logic exists in the wrapper today — the only "kick" occurrence is
a comment):

```
$ grep -n "kick" .claude/hooks/moai/handle-codex-review-entry.sh
7:# and kicks a background review; this wrapper is the entry-side half: it reads
exit=0
```

**E-L15 GREEN** (M3; re-pinned at M7, tree `703654daca` — the wrapper now carries the detached
background kick the RED cell found absent; the kick exec lives at lines 490/492, marker-guarded,
with SingleKickPerBucket pinning the exactly-one-per-bucket bound in the same M7 run):

```
$ grep -n "set -o noclobber\|verify codex-review" .claude/hooks/moai/handle-codex-review-entry.sh
303:	receipt=$(awk -v cmd="moai verify codex-review" -v now="$now" -v ttl=86400 '
485:	# The kick: the same detached `moai verify codex-review --project-root
490:		"$moai_bin" verify codex-review --project-root "$root" >>"$log_dir/codex-review-bg.log" 2>&1 &
492:		"$moai_bin" verify codex-review --project-root "$root" >/dev/null 2>&1 &
exit=0
$ go test ./internal/cli/ -run 'TestCodexEntryFastpath|TestHookEntryParity' -count=1 -v
--- PASS: TestCodexEntryFastpath_SingleKickPerBucket (3.11s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	17.932s
exit=0
```
(SingleKickPerBucket subtests: two entries one bucket → kick count 1; a pre-existing Stop-arm
marker → 0; fresh pass → 0.)

**E-L16** (AC-HSA-016, regression-guard per §2.1 — audit r1 D4's required disposition). The
criterion is a preservation guard: it pins behavior M3 must NOT change (card-scope sessions
reach the Go handler exactly once per prompt and consult no shell probe state). No red delta can
exist pre-M3 — today the wrapper probes nothing (E-L06) and forwards unconditionally (line 84),
which is precisely the behavior to preserve; a "falls through" RED would be vacuous (D4),
because the fall-through is already true. Baseline evidence: E-L06 + E-L07's invocation probe +
the plan-phase script read showing no `WT-` branch check. Green path: the M3 probe matrix's
`WT-` branch case asserts the preserved behavior on the rewritten wrapper.

**E-L16 GREEN** (M3; re-pinned at M7, tree `703654daca` — the WT- branch preservation case: the
Go handler reached exactly once, no shell probe state consulted):

```
$ go test ./internal/cli/ -run 'TestCodexEntryFastpath|TestHookEntryParity' -count=1 -v
=== RUN   TestCodexEntryFastpath_ProbeMatrix/WT-_branch:_card_scope_reaches_the_Go_handler_once
--- PASS: TestCodexEntryFastpath_ProbeMatrix (13.10s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	17.932s
exit=0
```
(the subtest asserts `forward=1 kick=0` exit 0 — the preserved behavior E-L16 pins, on the
rewritten wrapper.)

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

**E-L18 GREEN** (M3; re-pinned at M7, tree `703654daca` — the gate-OFF fast path preserved on the
rewritten wrapper: exit 0, stub never invoked):

```
$ go test ./internal/cli/ -run 'TestCodexEntryFastpath|TestHookEntryParity' -count=1 -v
=== RUN   TestCodexEntryFastpath_ProbeMatrix/gate_off_exits_0_with_no_binary_resolution
--- PASS: TestCodexEntryFastpath_ProbeMatrix (13.10s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	17.932s
exit=0
```

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

M7-run additions (2026-10-22, tree `703654daca`):

- The M7 `golangci-lint run --timeout=2m` surfaced 2 staticcheck findings:
  `internal/cli/paths.go:172 ST1008` (named in the M3/M4+M5 baselines) AND
  `internal/cli/factory_card.go:1995 QF1001` (not named in any prior baseline of this SPEC).
  Both files are byte-unchanged since the M5 tip — `git diff --name-only
  f04b507365..703654daca -- internal/cli/factory_card.go internal/cli/paths.go` prints 0 files —
  and neither is in this SPEC's run-phase file set, so the delta is not attributable to this
  SPEC's commits; what produced the changed finding set (tool build vs earlier tree state) is
  unmeasured here. Zero lint findings on SPEC-touched files.
- Plan §F M7's "measurement table finalized + committed" is executed as finalize + cite: the
  table path `.moai/reports/t1780/hook-sync-timings.md` is gitignored (`.gitignore:235`) and the
  card directive supersedes the commit leg ("cite it, never commit it"). The finalization
  (prescription-order statement) and its observations are recorded in E-L13 GREEN.
- AC-HSA-014's coverage boundary is recorded inside E-L14 GREEN (this tree's traces only; the
  multi-lane population's logs are outside this tree).

## Definition of Done

1. All ACs GREEN (AC-HSA-014 per its regression-guard closure rule) with ledger citations
   re-measured on the landing tree.
2. M1 measurement table committed at `.moai/reports/t1780/hook-sync-timings.md`.
3. Live/template pairs land together (C2); t1775 gate recorded before M3 (C1).
4. Run-phase §E evidence complete (E1-E7 of plan.md §E).
