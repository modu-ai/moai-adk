# t114 — Always-Loaded Token Budget Trim — Evidence

Card: t114 (release blocker, v3.1.1 round, kanban run tjv7iy)
Branch: `WT-t114` (base `origin/release/v3.1.1` @ `5ed668566` — one commit past the card's stated tip `49697cfb4`; the extra commit `5ed668566` touches only `internal/cli/{glm,kanban,kanban_bootstrap_test}.go`, no rule overlap)

## Claim

The always-loaded surface was trimmed from **76,680 → 75,026 tokens** (17 entries) — a net **-1,654 tokens** against the required ≥880 — leaving **974 tokens of headroom** under the 76,000 budget. `TestAlwaysLoadedTokenBudget` passes. The companion `session-handoff-examples.md` was brought from 40,974 to 39,751 bytes (under the 40,000 instruction-file cap, coordinator-added scope).

## Evidence (commands + verbatim outputs)

Baseline reproduction (before any edit, at branch point):

```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v
--- FAIL: TestAlwaysLoadedTokenBudget (0.01s)
    token_budget_guard_test.go:67: always-loaded surface = 76680 tokens, exceeds budget 76000 (overflow 680); surface has 17 entries — trim always-loaded rules or raise AlwaysLoadedTokenBudget with justification
```

After trim (verbatim tail, `budget-test-tail.txt`):

```
=== RUN   TestAlwaysLoadedTokenBudget
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
=== RUN   TestAlwaysLoadedTokenBudget_OverBudgetFails
--- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails (0.00s)
    --- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails/over-budget (0.00s)
    --- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails/under-budget (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.435s
```

Surface total (per-file measurement, `measure_tool.go` replicating `alwaysLoadedSurface` + `measureAlwaysLoaded`; see `after-per-file.txt`): `TOTAL: 75026 tokens, 17 entries` → headroom = 76,000 − 75,026 = **974 tokens**.

Template STRICT (verbatim, `template-strict-tail.txt`):

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/template	29.676s
```

Config suite: `ok github.com/modu-ai/moai-adk/internal/config 2.925s` (`config-suite-tail.txt`).
Vet: `go vet ./internal/config/ ./internal/template/` → exit 0.
`make build` → clean rebuild; `internal/template/catalog.yaml` recomputed with no content change (no mirror-commit hazard).
Mirror parity: `diff` local↔template byte-identical for all 5 modified/new rule files (`PARITY OK`).

## Baseline-attribution

Measured in this run, against this tree (WT-t114 @ base 5ed668566), with the same char/4 estimator the guard uses. Per-file before/after (tokens = bytes/4):

| File | Before (B / tok) | After (B / tok) | Delta |
|---|---|---|---|
| `.claude/rules/moai/workflow/kanban-dispatch.md` | 27,764 / 6,941 | 22,460 / 5,615 | −5,304 B / **−1,326 tok** |
| `.claude/rules/moai/core/agent-common-protocol.md` | 39,455 / 9,863 | 38,140 / 9,535 | −1,315 B / **−328 tok** |
| New: `kanban-dispatch-detail.md` (paths-gated) | 0 | 9,400 | not on surface |
| `agent-common-protocol-reference.md` (paths-gated) | 13,059 | 14,273 | not on surface |
| `session-handoff-examples.md` (paths-gated) | 40,974 | 39,751 | not on surface; under 40K cap |
| **Always-loaded total** | **76,680** | **75,026** | **−1,654 tok** |

Full per-file lists: `before-per-file.txt`, `after-per-file.txt`.

## What moved where (stub ↔ companion mapping)

### kanban-dispatch.md → kanban-dispatch-detail.md (new, `paths: "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-kanban.md,**/.claude/skills/moai/workflows/todo.md"`)

| Stub section (kept) | Moved to companion section |
|---|---|
| The board (2-line summary) | § The board (column→role table) |
| Entry into the board (operator-act paragraph + `/moai todo`) | (mechanism bullets folded into stub sentence; table context in companion § The board) |
| Card classes (class table inline-summarized; Class A [HARD], Class B, WIP verbatim) | § Card classes (full table, ceremony-cost rationale, parallelism inversion) |
| The dispatch cycle (one-line flow + pointer-not-copy) | § The dispatch cycle (diagram, `<role>-<run-id>` naming rules, walkthrough) |
| Dispatch language ([HARD] + verbatim-addresses list) | § Dispatch language — classification rationale (the two classification bullets) |
| Dispatch format ([HARD] + block + 3 field rules verbatim; release-only section, absent on main) | § Dispatch format — rationale (the "made mechanical" paragraph) |
| CodeRabbit predicate ([HARD] + both conditions + bash snippet verbatim) | § CodeRabbit endpoint measurement (combined-endpoint narrative, 5-entry measurement, branch-protection paragraph) |
| Review lens selection (pointer + `--deep --patch` opt-in prohibition) | § Review lens selection (lens table) |
| `/clear` handoff ([HARD] + structure summary) | § The `/clear` handoff between phases — message structure (3-item list detail) |
| Isolation ([HARD] ×5 + forms table verbatim) | § Isolation rationale (shared-checkout two properties) |
| Verification load ([HARD] ×2 verbatim) | § Verification load incident record (load-413 narrative, spin-loop story, contention↔flakiness loop) |
| Integration into release branch, Boundaries, Cross-references | kept verbatim (externally cross-referenced headings) |

External pointers verified still resolving: `worktree-integration.md` → "§ Integration into the release branch is self-served" (kept); `manager-kanban.md` → "§ Boundaries" (kept); CLAUDE.md §15 → file-level reference (kept).

### agent-common-protocol.md → agent-common-protocol-reference.md (existing sidecar, extended)

| Stub section (kept) | Moved to sidecar section |
|---|---|
| § Pre-Spawn Sync Check ([HARD] + commands + matrices + exemption) | § Pre-Spawn Sync Check rationale and incident record (Rationale + Origin paragraphs, verbatim) |
| § Parallel Execution → Attributable diff-check doctrinal switch (heading + 3 match conditions + 4 mismatch sentinels + never-silent-skip kept) | narrative paragraphs compressed in place (not moved — condensed; normative lists intact, (a)/(b) markers preserved for the manager-develop-prompt-template triple reference) |
| § Ledger Closure ([HARD] + clauses (a)-(d) verbatim) | opening provenance narrative condensed in place |
| § Background Agent Execution ([HARD] + rules verbatim) | supersession-history sentence condensed in place |

### session-handoff-examples.md (in-file tightening, no doctrine moved)

Removed/condensed: redundant self-references (2), duplicate anti-pattern pointer line, cross-pollination-history stub, `/cd` paragraph prose, intro blockquote, localization postscript + fallback-rule paragraph (binding text remains in `session-handoff.md` § Localization Table — pointer left). No [ZONE]/[HARD] content altered.

## Constraint compliance

- Every [HARD] rule statement stayed in an always-loaded body: kanban-dispatch 15/15 (grep count 16 = 15 rules + 1 meta-mention in the companion pointer note), agent-common-protocol 15/15. [ZONE] tags: agent-common-protocol 16/16. No [ZONE:Frozen] text touched (subagent-prohibition Frozen sections untouched).
- No rule demoted to a skill; both companions are `paths:`-gated rule files in `.claude/rules/moai/` (same mechanism as `goal-directive-detail.md`).
- Template-First: all 5 files mirrored byte-identical + `make build`.
- Neutrality incident during the work: copying local→mirror re-introduced `PR #1526` + internal dates that the mirror had been scrubbed of (pre-existing local↔mirror drift in `agent-common-protocol-reference.md`, NOT caused by this card). Fixed by adopting the mirror's neutral wording into local ("In the recorded incident", "a pull request"); STRICT suite green after fix.

## Gaps (not verified)

- `golangci-lint` not run (touched only `.md` + template mirrors; no Go source changed — vet on both touched packages exit 0).
- Full `go test ./...` not run locally (lane-local discipline per kanban-dispatch § Verification load; CI on the batch release PR owns the full matrix).
- Token counts are the guard's char/4 estimate, not a real tokenizer (same estimator as the binding test — the test is the criterion, and it passes).
- The mirror-scrub drift fix (PR#1526/dates) was verified via the STRICT suite only; the separate CI neutrality workflow (`template-neutrality-check.yaml`) will re-check on the release PR.

## Residual risk

- Headroom is 974 tokens (~1.28%). The surface historically regrows (5→14 files in 3 months per the session-cost analysis); remaining v3.1.1 lanes (t60 settings/web, t99 internal/cli) are code-focused, but any late rule addition could consume headroom. If the budget trips again before the release PR merges, the next candidates are `askuser-protocol.md` (7,538 tok, sidecar exists) and `context-window-management.md` (3,252 tok, would need a new companion).
- Content that moved to the kanban companion is now paths-gated: a lead session that never touches kanban rule/agent/todo files loads only the stub. The stub's summaries were written to keep every operational decision reachable via one hop (each moved block has an inline pointer). Risk: a lead following only the stub misses the `<role>-<run-id>` naming detail and the lens table without opening the companion — the stub's pointers name the exact companion section.
- The A9 compression kept the (a)/(b) markers and all sentinel names; the full elaboration remains in `verification-batch-pattern.md` § Attributable diff-check pattern (unchanged).


---

## Finalization pass — t110 rebase resolution + channel rider (2026-08-17)

Executed by the t114 finalizer agent. All content work and verification below ran
against this worktree (WT-t114 @ d9600065e + the edits described here). Git
bookkeeping (rebase / commit / release merge / push) is handed off verbatim via
`resolved/HANDOFF.md` — this session was spawned with a cwd override into a
different worktree and the worktree-session guard refuses cross-tree git, so the
sequence was left staged rather than guessed around.

### Rider (from hub review)

`.claude/rules/moai/workflow/cross-session-messaging.md` (local + template
mirror, byte-identical). The Availability-constraints opener said the kanban
dispatch cycle "rides entirely on this channel", contradicting t110's principle ①
(the queue on disk is the delegation channel; cross-session messages are nudge
only). One-sentence rewrite + equal-or-greater trim in the same file:

- Rewrite: `"Four constraints bound where it exists at all — and because Kanban Mode delegates through the queue on disk, using this channel only to nudge companions, they bound where its nudges reach:"` (+21 B)
- Compensating trim (wording-only): `"It is not free: a delivered message counts toward the recipient's usage exactly as a typed prompt does — what is saved is the user's attention, not tokens."` (−19 B)
- Net file delta +2 B → 14,286 B; token count unchanged at 3,571.

### Rebase-conflict resolution (WT-t114 ← origin/release/v3.1.1 @ 6b44bdd2e)

t110 (5c9a9715c) edited the full-size kanban-dispatch.md; t114 restructured it
into stub + companion. Every t110 change survives; new [HARD] rules stay in the
always-loaded stub, narrative in the stub only where t110 placed rule statements:

| t110 change | Landed in | Form |
|---|---|---|
| Scope: "nudge delivery rides on cross-session messaging … the queue itself keeps working without it" | stub § Scope | verbatim |
| [HARD] The lead is the queue's sole producer (with `moai todo add` mechanism) | stub § Entry | verbatim |
| [HARD] Promotion is the operator's act, always | stub § Entry | verbatim |
| § The delegation channel is the queue ([HARD] queue-on-disk + nudge paragraph) | stub, new § between The dispatch cycle and Dispatch language | verbatim |
| The final PASS/FAIL verdict is the lead's (structural division) | stub § Completion is read, never trusted | verbatim |
| Bare-role companion naming, no run id | companion § The dispatch cycle | verbatim — also fixes t114's stale `<role>-<run-id>` text (bare-role naming predates t110) |
| Wording-only compressions (WT- prefix [HARD], /clear ×2, isolation narrative, dispatch-this, mechanism paragraph, integration completion signal, boundaries bullet) | stub | t110's compressed wording; normative content unchanged |

Not taken from t110 (t114 stub/companion already equivalent or shorter): Class-A
parenthetical, CodeRabbit condensed predicate, dispatch-language condensed
rationale (companion keeps the two-bullet archive), verification-load pointer
form, one-integration-surface bullet (already compressed), env-isolated section.

### Budget impact (this pass)

| File | Before pass | After pass | Delta |
|---|---|---|---|
| kanban-dispatch.md (stub) | 22,460 B / 5,615 tok | 23,544 B / 5,886 tok | +271 tok |
| cross-session-messaging.md | 14,284 B / 3,571 tok | 14,286 B / 3,571 tok | 0 tok |
| kanban-dispatch-detail.md (paths-gated, off-surface) | 9,400 B | 9,504 B | — |
| **Always-loaded total** | **75,026 tok** | **75,297 tok** | **+271 tok** |

Final headroom: 76,000 − 75,297 = **703 tokens** (≥ 200 required). [HARD] in the
stub: 16 → 19 occurrences (18 rules + 1 companion-pointer mention); this file
carries no [ZONE] tags before or after. agent-common-protocol.md (15 [HARD] /
16 [ZONE]) untouched by this pass.

### Verification (this pass, verbatim — fresh `-count=1` runs, artifacts under `resolved/artifacts/`)

```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v -count=1
=== RUN   TestAlwaysLoadedTokenBudget
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.269s

$ go run .moai/reports/t114/measure_tool.go .
TOTAL: 75297 tokens, 17 entries

$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	50.790s

$ diff local↔mirror × 3 (kanban-dispatch, kanban-dispatch-detail, cross-session-messaging) → identical
$ make build → clean rebuild; internal/template/catalog.yaml byte-identical to d9600065e (no mirror-commit hazard)
$ go vet ./internal/config/ ./internal/template/ → exit 0
```

### Gaps (this pass)

- Rebase / rider commit / release merge / push NOT executed here (session
  worktree-pinned; guard refuses cross-tree git). Exact sequence staged in
  `resolved/HANDOFF.md`; all resolved content under `resolved/` (untracked).
  Post-rebase the budget test must be re-run once by the finishing session
  (release-side changes touch no surface file, so 75,297 is the expected value).
- `golangci-lint` not run (no Go source changed; vet on both touched packages exit 0).
- Full `go test ./...` not run locally (lane-local discipline; CI on the batch
  release PR owns the full matrix).
