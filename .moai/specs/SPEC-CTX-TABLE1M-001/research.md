---
id: SPEC-CTX-TABLE1M-001
title: "Research — upstream CC 2.1.285/2.1.287 facts and local file state"
created: 2026-10-02
---

# Research — SPEC-CTX-TABLE1M-001

> Stateless artifact (no `status:` field, per spec-frontmatter-schema.md §
> Artifact Statelessness). Added at repair iteration 1 (2026-10-02) per
> coordinator instruction; consolidates facts already cited in spec.md §1.2
> and measured in acceptance.md §Evidence Ledger — nothing here was newly
> fetched during plan phase.

## 1. Upstream facts (provisioned; not re-fetched in plan phase)

Source: https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md
— fetched 2026-10-02 by the leader's release-update sweep U2 and provided via
operator card t1415. Verbatim bullets:

- **CC 2.1.285** — "Changed sessions behind a custom `ANTHROPIC_BASE_URL` to
  use the 1M context window of models that have one (Opus 4.7+, Sonnet 5+,
  Fable); run `/autocompact 200k` if your gateway stops at 200K"
- **CC 2.1.287** — "Changed Opus 4.7+ and Fable to use a 1M context window by
  default on Bedrock, Vertex, Foundry and the Claude apps gateway, with no
  `[1m]` suffix (`CLAUDE_CODE_DISABLE_1M_CONTEXT=1` keeps 200K)"

Model-list fidelity note: the gateway bullet qualifies Sonnet 5+ / Opus 4.7+ /
Fable; the provider bullet qualifies Opus 4.7+ and Fable only. Wording that
merges the two into one model list across both surfaces would overstate the
provider case for Sonnet and is avoided in the C1/P3 drafts (audit D3).

## 2. Local file state (measured, tree `c50da9c2f`, worktree t1415)

- Stale inventory: 5 lines / 6 claims (spec.md §1.2 table). Both live/mirror
  pairs byte-identical at measurement (`cmp` exit 0 ×2 — acceptance.md E17/E18).
- Sizes: cwm live = mirror = 6,999 B; model-policy live = mirror = 27,051 B
  (`wc -c` — acceptance.md E20 and ledger header).
- Inventory closure independently re-verified by plan-audit iteration 1
  (`.moai/reports/t1415/plan-audit.md` §Evidence E-C): `budgets 200K` →
  exactly 4 hits (model-policy 22/90, live+mirror); `running with a 200K
  window` → exactly 2 hits (cwm 16, live+mirror); beyond the edit targets
  only model-policy line 139 remains (capability-companion text, not a
  window-default claim — correctly out of scope).
- RED-now ledger: acceptance.md §Evidence Ledger E1–E28; the auditor
  re-executed all 14 release-blocking cells verbatim, all reproduced (E-A).

## 3. SPEC-ID shape constraint (discovered this run)

The dispatch-minted ID carried a digit-initial `1M` middle segment and is
rejected by the enforced canonical shape `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` at
three sites: `internal/spec/lint.go:1301` (FrontmatterInvalid, SeverityError),
`internal/cli/spec_lint.go:369` (CLI-argument refusal), and
`internal/cli/specid/specid.go:37` (`CanonicalSpecIDShapeLiteral`). Renamed to
`SPEC-CTX-TABLE1M-001` by operator decision (repair iteration 1); the old ID's
provenance is recorded in spec.md HISTORY.

## 4. Byte arithmetic (repair iteration 1)

Old cwm line 16 = 382 B (`LC_ALL=C awk` length). Revised C1 draft = 481 B
(`printf %s | wc -c` on the verbatim draft). Delta +99 B → projected cwm file
7,098 B against the 7,099 B budget AC (REQ-CTM-006). The iteration-0 draft
(396 B, +14) grew because iteration 1 folds in the `/autocompact 200k`
gateway-cap caveat (D2) and the provider-surface note (D3); the merged
single-list phrasing suggested for D3 was rejected as unfaithful to the two
changelog bullets (see §1 fidelity note).
