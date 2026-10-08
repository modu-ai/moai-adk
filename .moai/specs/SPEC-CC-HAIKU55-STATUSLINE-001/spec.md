---
id: "SPEC-CC-HAIKU55-STATUSLINE-001"
title: "Haiku 5.5 documentation sync (GD-1) + subagentStatusLine agentType badge"
version: "0.1.0"
status: draft
created: 2026-10-08
updated: 2026-10-08
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "docs-sync,haiku-5-5,statusline,agenttype,claude-code-2-1-293,i18n"
tier: M
---

# SPEC-CC-HAIKU55-STATUSLINE-001

## 1. Background and Intent

Claude Code 2.1.293 introduced two upstream changes that touch MoAI-ADK surfaces:

1. **Haiku 5.5 (`claude-haiku-5-5`)** — the default Haiku on the Anthropic API is promoted to
   Haiku 5.5: 1M context on all plans (no `[1m]` suffix), requires CC v2.1.293+, auto-compact
   default ~967K. Alias resolution stays provider-split: Anthropic API → Haiku 5.5;
   AWS Bedrock / GCP Agent Platform / Microsoft Foundry → Haiku 4.5 (200K kept). Effort:
   `low/medium/high/xhigh/max` all supported, default `medium` (same as Opus 5.5 / Sonnet 5.5);
   thinking cannot be disabled (always adaptive reasoning). Rates per Mtok: input $0.10 /
   output $0.50; prompts over 100K tokens: input $0.50 / output $2.50.
2. **`subagentStatusLine.agentType`** — the `subagentStatusLine` `tasks[]` payload gained an
   `agentType` field (the custom subagent's type name; distinct from the existing `type`
   field). Official statusline docs do not yet list `agentType` (docs-lag measured 2026-10-08);
   the changelog is the canon.

Design is frozen; this SPEC absorbs the measured coordinates from the two design-canon reports
(`.moai/reports/release-update-20261008/upstream-update-20261008.md` § GD-1, and
`.moai/worktrees/t1538/.moai/reports/release-update/2026-10-08-haiku55-subagent-agenttype.html`)
and does not re-research them. All line-number coordinates in this SPEC were measured on the
t1538 tree @ `65e649d5f`; **every coordinate MUST be re-grepped at edit time** — values may
have drifted.

Go model configuration is UNCHANGED: alias-based resolution (`internal/config/defaults.go:155`
`DefaultSpeedModel = "haiku"`, `envkeys.go` `ANTHROPIC_DEFAULT_HAIKU_MODEL`) stays valid — no
config-code change is required or permitted for the Haiku axis.

## 2. HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-10-08 | 0.1.0 | Initial draft — plan-phase artifacts authored (Tier M: spec/plan/acceptance/progress) from the 2026-10-08 release-update canon. |

## 3. Requirements (GEARS)

### D1 — Haiku 5.5 documentation sync

- REQ-CC-HAIKU55-001: The MoAI documentation surfaces shall state the Haiku 5.5 facts as
  measured in §1 (model id `claude-haiku-5-5`, 1M context on all plans, CC v2.1.293+,
  auto-compact ~967K, provider-split alias resolution, effort set `low/medium/high/xhigh/max`
  with default `medium`, always-on adaptive thinking, rates $0.10/$0.50 per Mtok and
  $0.50/$2.50 above a 100K prompt).

- REQ-CC-HAIKU55-002: The context-window threshold table in
  `.claude/rules/moai/workflow/context-window-management.md` (row at ~line 17) shall carry a
  **Haiku 5.5 (1M) | 1,000,000 tokens | 50% | ~500,000 tokens** row coexisting with the
  existing **Haiku (200K) | 200,000 tokens | 90% | ~180,000 tokens** row, plus a
  provider-interpretation note consistent with the table's existing rule that "a session that
  matches both a 1M row and the 200K-sessions row takes the 200K row" (AWS-lineage Haiku 4.5
  sessions keep 200K/90%). The new row composes with that rule; it never bypasses it.

- REQ-CC-HAIKU55-003: `.claude/rules/moai/development/model-policy.md` shall be updated at the
  measured anchors — the `haiku` alias description lines (~:18 and ~:26), the "Haiku 4.5 still
  ships 200K" sentence (~:83), and the effort-support sentence (~:161, which gains Haiku 5.5:
  xhigh and max supported, default medium) — **while preserving the No-Haiku policy phrasing**
  ("retired from MoAI agent routing per the No-Haiku policy") and the HaikuResidualRule lint
  scope (agent definitions + the `claude_models` block) untouched.

- REQ-CC-HAIKU55-004: `.claude/rules/moai/development/prompting-best-practices.md:7` shall name
  the official guide family (Opus 5/5.5, Sonnet 5.5, Haiku 5.5) in place of the stale
  "(Opus 4.8/4.7, Sonnet (current generation), Haiku 4.5)" list.

- REQ-CC-HAIKU55-005: For each of the three rules files above, the template mirror under
  `internal/template/templates/.claude/rules/...` shall be byte-identical to the live file
  after the edit, and `make build` shall succeed (templates are `go:embed`-ed).

- REQ-CC-HAIKU55-006: The docs-site pages named by the canon — `multi-llm/model-policy.md`,
  `multi-llm/_index.md`, `advanced/token-budget.md`, `cost-optimization/prompt-caching.md`,
  `claude-code/context-memory/context-window.md`, `claude-code/foundations/commands.md`,
  `claude-code/foundations/how-claude-code-works.md`, `claude-code/_index.md` — shall carry a
  Haiku 5.5 entry (`claude-haiku-5-5`, 1M, provider-split alias, v2.1.293+) in every locale
  (ko/en/ja/zh) where the page carries its Haiku-bearing row or table, following the
  ko-canonical chain (ko → en → ja/zh), and `scripts/docs-i18n-check.sh` shall exit 0.

- REQ-CC-HAIKU55-007: The README 4-locale set (ko canonical) shall gain the Haiku 5.5 row in
  every model table that names a Haiku model generation, per the `hns-oss-docs-readme-sync`
  synchronization contract.

- REQ-CC-HAIKU55-008: **When** the Haiku 5.5 docs refresh lands, the No-Haiku policy surface
  shall remain coherent: no agent definition and no `claude_models` config block gains a haiku
  reference, and the model-policy "retired from routing" statements remain factually true
  against the refreshed lineup.

### D2 — subagentStatusLine agentType

- REQ-CC-HAIKU55-009: The template `settings.json`
  (`internal/template/templates/.claude/settings.json.tmpl`, `statusLine` block at ~line 405)
  shall wire a `subagentStatusLine` block alongside the existing `statusLine` block, routing
  the subagent statusline event to the same MoAI statusline entry point
  (`.moai/status_line.sh`). Measured 2026-10-08 (tree t1538 @ `65e649d5f` and this tree):
  no subagent statusline wiring exists — the template carries only the single `statusLine`
  block — so the edit is unconditional (re-grep at edit time to re-confirm the anchor, not
  to condition the requirement).

- REQ-CC-HAIKU55-010: The `internal/statusline` package shall parse the subagentStatusLine
  `tasks[]` payload and render an agentType badge per task row, following the package's
  existing stdin-parsing conventions (`StdinData` in `internal/statusline/types.go`,
  orchestration in `internal/statusline/builder.go`).

- REQ-CC-HAIKU55-011: **When** the `agentType` field is absent or null in a tasks[] row, the
  statusline shall degrade silently — render the row without the badge, matching the package's
  existing missing-field conventions (pointer-nil optional fields) — and shall not error,
  panic, or drop the row.

## 4. Non-Functional Constraints

- 4-locale docs discipline (ko-canonical chains, no body emoji in docs-site, per-locale
  emphasis-marker spacing, Mermaid TD-only, URL whitelist `adk.mo.ai.kr`) — the run-phase
  worker loads the `hns-oss-docs-i18n-rules` skill before any docs-site or README edit.
- TRUST 5: new Go code carries unit tests (agentType present / absent / null); coverage on
  `internal/statusline` stays at or above the package's pre-change level.
- No time estimates anywhere; milestones are priority-ordered.

## 5. Exclusions

### Out of Scope — Go model configuration

- No change to `internal/config/defaults.go` (`DefaultSpeedModel = "haiku"` stays), to
  `envkeys.go`, or to any alias-resolution code — the alias layer already resolves correctly.

### Out of Scope — Codex axis and other upstream deltas

- Codex 0.161.0 conformance re-measurement (CARD-4 proposal), BP mapping cards (CARD-2/3
  amendments), and every other 2.1.293/2.1.294 delta outside the two deliverables named in §1.

### Out of Scope — Haiku 5.5 unmeasured axes

- Max output tokens, knowledge cutoff, Batch/Prompt-caching rates, beta-header behavior, and
  live API alias-resolution verification — unmeasured in the canon (GLM backend); not encoded.

### Out of Scope — leader-owned settlement

- Queue record settlement, card `done`, merge, and push — leader-owned (lane context: this
  card was leased without a factory lease — the serial slot was wedged by the ownerless-lease
  assigned row t1595; leader-dispatched; t1498 precedent applies).

## 6. Cross-References

- Design canon: `.moai/reports/release-update-20261008/upstream-update-20261008.md` § GD-1
  (binding coordinates) · `.moai/worktrees/t1538/.moai/reports/release-update/2026-10-08-haiku55-subagent-agenttype.html`
  (docs-lag observations)
- Skills contract for run phase: `hns-oss-docs-i18n-rules`, `hns-oss-docs-readme-sync`,
  `hns-oss-docs-verify`
- Related: SPEC-AGENT-ARCH-V2-001 §D (No-Haiku policy), t1600 (worker fanout — a future
  consumer of the agentType badge; not a dependency of this SPEC)
