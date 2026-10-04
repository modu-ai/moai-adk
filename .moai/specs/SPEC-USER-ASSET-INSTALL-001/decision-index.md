---
id: SPEC-USER-ASSET-INSTALL-001
title: "decision-index.md — decision register (gate on)"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Decision Index — SPEC-USER-ASSET-INSTALL-001

Stateless register: this artifact carries no `status:` field; the SPEC's
lifecycle lives in `spec.md` alone.

## Operator-resolved premises (card t1509)

### Q1: Does the install model use a plugin/marketplace carrier?
Label: DECIDED
Authority anchor: operator decision record delivered with card t1509 (gtd
queue, 2026-10-05, "[v3.2 운영자 결정 D3·D4]"); durable anchor becomes this
SPEC's HISTORY row at completion. Caveat recorded honestly: the card is the
operator's written instruction, not yet a committed-tree artifact.
Why unresolved: n/a — recorded as a premise per the dispatch.
Operator verdict: NO plugin. moai copies all common skills + agents into the
user folders (D4).

### Q2: Is factory (multi-lane operation) part of the L0 core bundle?
Label: DECIDED
Authority anchor: same operator decision record as Q1 (D3 clause: "D3:
핵심(L0)에 factory(여러 레인 운영) 포함").
Why unresolved: n/a — settled premise.
Operator verdict: YES — L0 = plan·run·sync + 핵심 에이전트 5종 + 훅 + factory;
everything else is opt-in bundles.

### Q3: Are user-created files with colliding names overwritten?
Label: DECIDED
Authority anchor: same operator decision record (card clause: "사용자가 만든
같은 이름 파일은 덮어쓰지 않고 보고").
Why unresolved: n/a — settled premise.
Operator verdict: NO — never overwritten; reported.

## Open founder questions (operator verdict pending)

### Q4: Which five agents are the "핵심 에이전트 5종" of L0?
Label: FOUNDER
Authority anchor: n/a — no committed artifact enumerates the five.
Why unresolved: catalog core tier carries 11 agents (sync-auditor,
manager-develop, manager-docs, manager-git, manager-lead, manager-spec,
plan-auditor, super-advisor, manager-todo, manager-design, e2e-tester; plus
builder-harness as harness-generated — research V14). Candidate reading A:
manager-spec, manager-develop, manager-docs, plan-auditor, sync-auditor (the
plan→run→sync chain with its two auditors). Candidate readings B/C include
manager-git or manager-lead. The card says "핵심 에이전트 5종" without naming
them.
Operator verdict:

### Q5: Where does the per-user manifest live, and what is it called?
Label: FOUNDER
Authority anchor: n/a — the location is new surface.
Why unresolved: candidates with different trade-offs: `~/.moai/user-assets.json`
(leading — `~/.moai/` is already moai's user-level state home per
`internal/paths/paths.go:91`), `~/.claude/moai-manifest.json` (next to the
assets it tracks, but Claude-only), per-root split files (no single source).
Operator verdict:

### Q6: How do CLAUDE_CONFIG_DIR profile sessions see the user-level assets?
Label: FOUNDER
Authority anchor: n/a — no committed artifact settles profile asset
visibility.
Why unresolved: a profile is an isolated `CLAUDE_CONFIG_DIR`
(`internal/cli/profile.go:20-22`); whether a profile session reads
`~/.claude/skills` at all was NOT measured (research §3). Options: (a)
provision per-profile symlinks to the user folders, (b) per-profile copies
(manifest must then track per-profile), (c) declare the limitation for v1 —
profiles keep their own settings and simply do not see shared assets.
Operator verdict:

### Q7: Does "plan·run·sync" in the L0 definition name the published command skills, the workflow skills, or both?
Label: FOUNDER
Authority anchor: n/a — the card's wording is not disambiguated by any
committed artifact.
Why unresolved: reading (a) `moai-plan`/`moai-run`/`moai-sync` published
command skills (literal adjacency to agents/hooks/factory in the card);
reading (b) `moai-workflow-spec`/`-tdd`/`-ddd` workflow skills; reading (c)
both. Research V15.
Operator verdict:

### Q8: What is the bundle granularity for non-L0 assets?
Label: FOUNDER
Authority anchor: n/a — the card says only "나머지는 선택 묶음".
Why unresolved: (a) the six existing optional packs stand as-is and
current-`core` remainders re-bundle by theme, (b) remainders join existing
packs where topical, (c) one single "extended" bundle. Affects catalog shape
and the update removal path.
Operator verdict:

### Q9: Hard delete or deprecation window for the retired plugin surfaces?
Label: FOUNDER
Authority anchor: SPEC-PLUGIN-MARKETPLACE-001 is `completed` with golden
tests and doctor checks pinned to the carrier; no committed policy states a
deprecation-window norm for retired carriers.
Why unresolved: hard delete satisfies C5 (atomic retirement) and the card's
"처분 포함"; a window preserves rollback for users on the plugin path. Affects
M6 scope.
Operator verdict:
