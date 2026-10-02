---
id: SPEC-CTX-TABLE1M-001
title: "Context-window 1M-default correction — CC 2.1.285/2.1.287 gateway and cloud-provider defaults"
version: "0.1.1"
status: completed
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P2
phase: "v3.2.0"
module: ".claude/rules/moai"
lifecycle: spec-anchored
tier: S
tags: "context-window, 1m-context, gateway, bedrock, vertex, foundry, documentation, template-mirror, claude-code-2-1-285"
---

# SPEC-CTX-TABLE1M-001 — Context-window 1M-default correction (CC 2.1.285/2.1.287)

## HISTORY

| Version | Date       | Author       | Change |
|---------|------------|--------------|--------|
| 0.1.0   | 2026-10-02 | manager-spec | Initial draft. Card t1415 (operator-approved dispatch). Docs-only correction across one live/mirror rule pair. |
| 0.1.1   | 2026-10-02 | manager-spec | Repair iteration 1 (plan-audit FAIL 0.88/1.00, MP-3 id-shape; report `.moai/reports/t1415/plan-audit.md`). SPEC ID renamed SPEC-CTX-TABLE-1M-001 → SPEC-CTX-TABLE1M-001 per operator decision: the dispatch-minted ID carried a digit-initial `1M` middle segment and is machine-refused by the enforced canonical shape `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` at `internal/spec/lint.go:1301` (FrontmatterInvalid, SeverityError), `internal/cli/spec_lint.go:369` (CLI-argument refusal), and `internal/cli/specid/specid.go:37` (CanonicalSpecIDShapeLiteral); the old ID is retained in this row as provenance only. Same iteration: AC-CTM-008 green cell de-ordering-sensitized (D1); REQ-CTM-008 added so AC-CTM-007/008 carry REQ back-references (D7); cwm row draft extended with the `/autocompact 200k` gateway-cap caveat (D2) and the CC 2.1.287 provider-surface note with per-surface model lists kept changelog-faithful (D3); P3 names the fourth 2.1.287 surface, the Claude apps gateway (D4); never-`moai update` added to plan §E (D6); `[1m]`-vs-flag precedence recorded as residual risk (D5); research.md and decision-index.md added. |

## 1. Overview

### 1.1 Goal

Correct the two rule documents that still describe 200K context windows as the
default expectation on custom LLM gateways and on Bedrock / Google Cloud /
Microsoft Foundry, where Claude Code 2.1.285 and 2.1.287 now default the 1M
window for the qualifying model generations. Five stale lines carrying six
stale claims (each line in a live copy and a byte-identical template mirror)
are rewritten; every other byte of the two files is preserved.

### 1.2 Problem

**Factual basis (cited, not re-verified).** The card's factual source is the
leader's release-update sweep U2, covering CC 2.1.285 and 2.1.287:

- **CC 2.1.285** — on custom `ANTHROPIC_BASE_URL` gateways (non-Anthropic),
  the 1M context window now defaults for Sonnet 5+ · Opus 4.7+ · Fable.
- **CC 2.1.287** — on Bedrock / Vertex / Foundry, the 1M context window now
  defaults for Opus 4.7+ · Fable.
- Remaining 200K paths: `CLAUDE_CODE_DISABLE_1M_CONTEXT=1` · Opus 4.6 (and
  Sonnet 4.6 without `[1m]`) · older models (Sonnet 4.5 / Opus 4.5 and
  earlier) · non-qualifying older models behind a gateway (Sonnet 4.x and
  older).

This SPEC does NOT restate those upstream facts as independently re-verified;
plan phase measured only the local rule text against them.

**Locally measured defect state** (grep + full-file reads, this run, worktree
t1415, tree `c50da9c2f`, working tree clean). The stale text lives in exactly
five lines carrying six stale claims (line 16 of the cwm rule holds two), all
inside the four edit-target files:

| # | File | Line | Stale claim (verbatim fragment) |
|---|------|------|--------------------------------|
| D1 | `.claude/rules/moai/workflow/context-window-management.md` | 16 | "Opus 4.8+ running with a 200K window (e.g. on Bedrock / Google Cloud / Foundry)" |
| D2 | same | 16 | "`sonnet` behind an LLM gateway (non-Anthropic `ANTHROPIC_BASE_URL`) unless `sonnet[1m]` is selected" |
| D3 | `.claude/rules/moai/development/model-policy.md` | 22 | "Behind an LLM gateway or with `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, `sonnet` budgets 200K." |
| D4 | same | 28 | "Opus 4.8 and later run with a 200K window on some providers, such as Amazon Bedrock, Google Cloud, and Microsoft Foundry." |
| D5 | same | 66 | "`sonnet` behind an LLM gateway or under `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, and Opus on a 200K provider such as Amazon Bedrock" |
| D6 | same | 90 | "behind an LLM gateway (`ANTHROPIC_BASE_URL` non-Anthropic) or with `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, `sonnet` budgets 200K" |

The dispatch named D1/D2 (line 16) and D6 (line 90). D3, D4, and D5 were
discovered by the plan-phase sweeps: D3 is the identical gateway-sonnet claim
in the same file (the dispatch's file-level stale grep cannot reach zero
without it); D4 is the provider-window paragraph that D5 cross-references
("see ... the provider-window paragraph at the top of this file"), so
correcting D5 while leaving D4 would point a corrected line at a stale target;
D5 is the same gateway/Bedrock enumeration in the breaker-resolution scoping
sentence. All three are the same defect claim (gateway/provider 200K defaults)
in the dispatched file, so they are in scope. One adjacent instance was
reviewed and deliberately left: line 43's "gateway-selected Sonnet paths" is a
credit-entitlement claim, not a window-default claim — outside this defect
class (recorded in §3.1).

Each defect instance exists twice — live rule + template mirror
(`internal/template/templates/...`). Both pairs are currently byte-identical
(`cmp` exit 0, both pairs, measured this run): the mirrors carry the stale
text, so fixing only the live side would leave deployed templates stale — the
recurring one-side-only defect this card family repairs.

### 1.3 Why now

- The rule tree is the always-loaded context surface: `context-window-management.md`
  (6,999 B, no `paths:` frontmatter) is re-injected every turn and tells the
  orchestrator which handoff threshold to apply. A session on a gateway or on
  Bedrock/Vertex/Foundry running Sonnet 5+/Opus 4.7+/Fable is now a 1M
  session by default; the current row tells that session to `/clear` at 90%
  (~180K) instead of 50% (~500K) — roughly a 5x-premature context reset on
  every such session until this lands.
- `model-policy.md` (27,051 B, `paths:`-scoped to `**/.claude/agents/**`) is
  the `[1m]`-doctrine SSOT; its gateway and provider claims feed the
  alias-entry and breaker-resolution sections that future model-matrix edits
  will build on. Stale premises there compound.
- The defect instances were found during the leader's release-update sweep
  (U2); the correction is mechanical once scoped — the card is queued now so
  the sweep's findings land before the next model-matrix card reuses them.

### 1.4 Impact and honest tier note

- **Modified files**: 4 (two live/mirror pairs). Docs-only; no Go source.
- **LOC delta**: 5 line replacements; measured draft deltas (repair iteration
  1): cwm line 16 382→481 bytes (+99); model-policy lines 22/28/66/90 →
  +637 chars total (P3 extended with the fourth 2.1.287 surface). Net cwm
  growth +99 B — within the +100 B noise budget (REQ-CTM-006) and well under
  the 1,000 B single-edit duty threshold of rule-authoring.md (no statement
  duty fires).
- **Tier**: LOC and file counts are Tier S range. The dispatch mandates the
  three-artifact set (spec + plan + acceptance), i.e. the Tier M artifact
  structure on Tier S scope — the same reconciliation SPEC-WORKTREE-GUARD-
  HEREDOC-DOC-001 recorded; acceptance-criteria ceiling taken as Tier M's 16.
  Repair iteration 1 adds research.md (coordinator instruction) and
  decision-index.md (`interview.decision_gate: on`), so the directory now
  carries six files; REQ/AC counts stand at 8/8, inside the Tier S ceilings.
- **User impact**: deployed users receive the corrected rules via `moai
  update` (the mirror is the `go:embed` source). Behavior of the `moai`
  binary itself is unchanged.

## 2. Goals

1. The 200K-sessions row of the context-window threshold table names exactly
   the surviving 200K paths per CC 2.1.285/2.1.287 and asserts no
   provider/gateway 200K default that those releases removed.
2. `model-policy.md` splits the flag path (still 200K, all native-1M models,
   CC 2.1.223 blast radius) from the gateway path (200K now only below the
   default-1M line) and stops describing Bedrock/Google Cloud/Foundry as
   200K-default providers.
3. Both template mirrors carry the corrected text byte-identically in the
   same commit (`cmp` exit 0 on both pairs post-edit).
4. The always-loaded cost of the corrected cwm rule does not grow beyond
   noise (≤ +100 B against the 6,999 B baseline).
5. The embedded templates are refreshed (`make build`) after the mirror
   edits, so the built binary carries the corrected text.

## 3. Non-Goals

### 3.1 Out of Scope

- **No Go code changes** — the card is docs-only. `internal/` and `pkg/` Go
  sources are untouched; `make build` runs only to refresh the embedded
  templates.
- **CHANGELOG / README / docs-site sync** — belongs to the sync phase
  (manager-docs), not this SPEC's run phase.
- **Line 43 of model-policy.md ("gateway-selected Sonnet paths")** — reviewed
  and left: it is a `[1m]` credit-entitlement claim (where the spawn-time
  credit mismatch can occur), not a context-window default claim. Different
  defect class; changing it would exceed the card's factual basis.
- **Line 68 of model-policy.md** — its "200K-budget paths named above" and
  legacy-model enumeration stay accurate once lines 22/28/66/90 are
  corrected; no edit needed (coherence verified in plan §E review).
- **The threshold values themselves** (50% / 90%, window sizes) — unchanged;
  only the model-class membership of the 200K row is corrected.
- **Other rules mentioning 200K** (e.g. `session-handoff.md` Trigger #1, which
  consumes this table by reference) — they read from this table; correcting
  the source corrects them.
- **Upstream verification** — the CC 2.1.285/2.1.287 release-note facts are
  cited from the leader's sweep U2, not re-fetched.

## 4. GEARS Requirements

> Notation: GEARS 5-state (Ubiquitous / Event-driven / State-driven / Where /
> Event-detected). All requirements here are Ubiquitous: they are static
> content-correctness rules over documentation, with no trigger event.

### REQ-CTM-001 — 200K-sessions row correction (Ubiquitous)

The 200K-sessions row of `.claude/rules/moai/workflow/context-window-management.md`
shall list as 200K paths exactly: (a) any native-1M model under
`CLAUDE_CODE_DISABLE_1M_CONTEXT=1`; (b) Sonnet 4.6 / Opus 4.6 without `[1m]`;
(c) models below the 1M-default line behind an LLM gateway (non-Anthropic
`ANTHROPIC_BASE_URL`), the row noting both upstream defaults — CC 2.1.285
defaults gateways to 1M for Sonnet 5+ / Opus 4.7+ / Fable, and CC 2.1.287
defaults Bedrock / Vertex / Foundry to 1M for Opus 4.7+ and Fable — so the
gateway 200K path is Sonnet 4.x and older; (d) a 200K-capped gateway, the row
carrying the `/autocompact 200k` remedy; and (e) Sonnet 4.5 / Opus 4.5 and
earlier. The row shall not assert that Opus 4.8+
runs with a 200K window on Bedrock / Google Cloud / Foundry (CC 2.1.287
removed that default expectation), and shall not assert that gateway `sonnet`
budgets 200K unless `sonnet[1m]` is selected.

근거 (Rationale): D1/D2. The threshold table is the always-loaded handoff
signal; a wrong row misclassifies the session's `/clear` threshold for every
gateway and cloud-provider session.

### REQ-CTM-002 — Row-precedence coherence preserved (Ubiquitous)

The row-precedence sentence following the threshold table ("A session that
matches both a 1M row and the 200K-sessions row takes the 200K row ...") shall
remain present and semantically valid against the corrected row: under
`CLAUDE_CODE_DISABLE_1M_CONTEXT=1` a Sonnet 5.5 session still matches both a
1M row and the 200K row, and the 200K row wins. The sentence is not edited;
this requirement verifies coherence, not change.

근거: dispatch instruction ("keep line 19 coherent"); the precedence rule is
what makes the flag path correct without per-flag rows.

### REQ-CTM-003 — model-policy gateway-200K claim correction (Ubiquitous)

`.claude/rules/moai/development/model-policy.md` shall not claim that `sonnet`
behind an LLM gateway budgets 200K by default. The sonnet alias entry (line
22) and the `[1m]`-doctrine surviving-paths bullet (line 90) shall scope the
gateway 200K budget to models below the default-1M line, citing CC 2.1.285
(gateways default 1M for Sonnet 5+ / Opus 4.7+ / Fable). The
`CLAUDE_CODE_DISABLE_1M_CONTEXT=1` binding — under the flag `sonnet` budgets
200K, and per CC 2.1.223 the flag holds every native-1M Claude model (Opus 5.5
included) to 200K — shall be preserved in meaning in both lines; the flag's
blast-radius sentence in line 90 is preserved verbatim.

근거: D3/D6. The flag path is the load-bearing 200K path and must not be
weakened while the gateway half of the old conflation is removed.

### REQ-CTM-004 — model-policy provider-200K-default claim correction (Ubiquitous)

`.claude/rules/moai/development/model-policy.md` shall not describe Bedrock /
Google Cloud / Microsoft Foundry as providers where Opus 4.8+ runs with a 200K
window by default. The provider-window paragraph (line 28) shall state that
since CC 2.1.287 the 1M window also defaults there for Opus 4.7+ and Fable,
with a 200K window on those providers now the opt-in; the
breaker-resolution scoping sentence (line 66) shall enumerate the surviving
200K-budget paths consistently with the corrected paragraph it
cross-references. The unchanged remainder of both lines (the breaker history,
the cross-reference parenthetical, the Anthropic-API resolution) shall be
preserved.

근거: D4/D5. Correcting the referrer (line 66) while leaving its named target
(line 28) stale would leave the document internally contradictory.

### REQ-CTM-005 — Template mirror byte parity (Ubiquitous)

Both template mirrors —
`internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`
and `internal/template/templates/.claude/rules/moai/development/model-policy.md`
— shall carry byte-identical copies of the corrected live rules in the same
commit: `cmp` exit 0 on each live/mirror pair post-edit. Each line edit is
applied to the mirror first, then replicated byte-identically to the live
copy.

근거: the mirror is the `go:embed` deploy source; a one-side-only fix ships
stale rules to every `moai update` user. Both pairs are byte-identical today
(measured, tree `c50da9c2f`) and must remain so after the edit.

### REQ-CTM-006 — Always-loaded byte budget (Ubiquitous)

The post-edit size of `.claude/rules/moai/workflow/context-window-management.md`
(the always-loaded copy) shall not exceed 7,099 bytes (baseline 6,999 B +
100 B tolerance); the expected value is ≈7,098 B (+99 B from the row
replacement, byte-measured at repair iteration 1 — the C1 draft grew when the
gateway-cap caveat and the provider-surface note were folded in). No single-edit growth exceeds the 1,000 B threshold of
`.claude/rules/moai/development/rule-authoring.md` § statement duty, so no
cost-statement duty fires. `model-policy.md` is `paths:`-scoped (frontmatter
`paths: "**/.claude/agents/**"`) and outside the always-loaded surface; its
projected growth (~+610 B) is recorded but carries no statement duty.

근거: rule-authoring.md byte-budget discipline; the dispatch constraint
"net-shrink or stay neutral on the always-loaded file".

### REQ-CTM-007 — Embed refresh (Ubiquitous)

After both mirror edits land, `make build` shall run and exit 0, so the
rebuilt binary embeds the corrected templates. No Go source changes are made;
the build's role here is embed refresh only (its preceding agents-emit-check
runs read-only and is unaffected — this card touches no agent definitions).

근거: AGENTS.local.md §2 Template-First cycle (edit templates → `make
build`); `internal/template/embed.go` compiles the templates into the binary.

### REQ-CTM-008 — Template neutrality and docs-only scope (Ubiquitous)

The corrected rule texts shall introduce no card id, no SPEC ID, no internal
dates, and no report paths into the mirrored rules (template neutrality, plan
§E item 6), and the run phase shall modify no file outside the four edit
targets of §5 — docs-only, zero `.go` paths.

근거: plan-audit D7 — the AC-CTM-007/008 obligations previously lived in plan
§E item 6 / spec §5 outside the REQ layer; recorded here as the eighth
requirement (Tier S ceiling 8) so both ACs carry a REQ back-reference.

## 5. Affected Files

### MODIFY (4 — two byte-identical live/mirror pairs)

| File | Lines | Est. delta | Change |
|------|-------|-----------|--------|
| `.claude/rules/moai/workflow/context-window-management.md` | 16 | 382→481 bytes | REQ-CTM-001/002: 200K row rewritten (dispatch-named; iteration 1 adds the gateway-cap caveat and the provider-surface note) |
| `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | 16 | byte-identical | REQ-CTM-005: mirror of the above |
| `.claude/rules/moai/development/model-policy.md` | 22, 28, 66, 90 | ~+637 chars total | REQ-CTM-003/004: four stale lines corrected (line 90 dispatch-named; 22/28/66 plan-discovered, §1.2) |
| `internal/template/templates/.claude/rules/moai/development/model-policy.md` | 22, 28, 66, 90 | byte-identical | REQ-CTM-005: mirror of the above |

### PRESERVE (not touched)

- The threshold values (50% / 90%, 1,000,000 / 200,000 / ~500,000 / ~180,000)
  and every other table row of `context-window-management.md`, including line
  19 (REQ-CTM-002 verifies, does not edit).
- model-policy.md line 43 (credit-entitlement claim — out of scope, §3.1) and
  line 68 (coherent once lines 22/28/66 are corrected).
- The `[1m]`-doctrine flag blast-radius sentence in line 90 (CC 2.1.223) —
  preserved verbatim inside the rewritten line (REQ-CTM-003).
- All Go sources, hooks, CI workflows, agent definitions, skill bodies.
- `.moai/specs/SPEC-CTX-TABLE1M-001/**` artifacts (this SPEC set).

## 6. References

- Card: t1415 (operator-approved dispatch; factual source: leader's
  release-update sweep U2, CC 2.1.285 + 2.1.287)
- Edit targets and verbatim old/new line text: `plan.md` §B (authoritative
  edit specification)
- RED-now evidence ledger: `acceptance.md` §Evidence Ledger (tree `c50da9c2f`)
- Two-cell discipline + swept-set rule: `.claude/rules/moai/development/verification-completeness.md` §2, §1.1
- Byte-budget discipline: `.claude/rules/moai/development/rule-authoring.md`
- Template-First cycle: `AGENTS.local.md` §2; mirror-parity context:
  `internal/template/rule_template_mirror_test.go`
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Plan-audit iteration 1: `.moai/reports/t1415/plan-audit.md` (FAIL 0.88/1.00,
  MP-3 id-shape — the scope of this repair iteration)
