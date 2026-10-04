---
id: SPEC-USER-ASSET-INSTALL-001
title: "decision-index.md — decision register (gate on)"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Decision Index — SPEC-USER-ASSET-INSTALL-001

Stateless register: this artifact carries no `status:` field; the SPEC's
lifecycle lives in `spec.md` alone. Identifier convention (iter1 repair D1):
the six gate questions are D-Q1..D-Q6 — the same identifiers spec.md §5,
plan.md, design.md, research.md, and acceptance.md cite. The operator
premises keep the spec.md §1 labels P1-P6.

## Operator-resolved premises (card t1509)

### P1 (D4): Does the install model use a plugin/marketplace carrier?
Label: DECIDED
Authority anchor: operator decision record delivered with card t1509 (gtd
queue, 2026-10-05, "[v3.2 운영자 결정 D3·D4]"); durable anchor becomes this
SPEC's HISTORY row at completion. Caveat recorded honestly: the card is the
operator's written instruction, not yet a committed-tree artifact.
Why unresolved: n/a — recorded as a premise per the dispatch.
Operator verdict: NO plugin. moai copies all common skills + agents into the
user folders (D4).

### P2 (D3): Is factory (multi-lane operation) part of the L0 core bundle?
Label: DECIDED
Authority anchor: same operator decision record as P1 (D3 clause: "D3:
핵심(L0)에 factory(여러 레인 운영) 포함").
Why unresolved: n/a — settled premise.
Operator verdict: YES — L0 = plan·run·sync + 핵심 에이전트 5종 + 훅 + factory;
everything else is opt-in bundles.

### P3: Are user-created files with colliding names overwritten?
Label: DECIDED
Authority anchor: same operator decision record (card clause: "사용자가 만든
같은 이름 파일은 덮어쓰지 않고 보고").
Why unresolved: n/a — settled premise.
Operator verdict: NO — never overwritten; reported.

## Constraint-forced resolutions (closed at plan phase, iter1 repair D7)

### D-Q3: How do CLAUDE_CONFIG_DIR profile sessions see the user-level assets?
Label: POLICY-COVERED
Authority anchor: this SPEC's hard rules — spec.md REQ-002 (a profile
directory is never an install target) + C2 (four-root confinement), which
codify operator premise P4 ("per-profile settings folders stay per-profile",
card t1509). Same caveat as P1: the codifying constraints are this plan
phase's committed artifacts.
Why unresolved: n/a — closed. Both install-into-profile options (per-profile
symlinks, per-profile copies) write into `CLAUDE_CONFIG_DIR` profiles and
violate REQ-002, C2, and AC-019; the only surviving option is the declared
limitation, recorded as premise P6 in spec.md §1.
Operator verdict: closed by constraint — profile sessions do not see the
shared user assets in v1 (declared limitation, doc-visible per acceptance
§D.7). M2 ships without profile provisioning; M8 documents the limitation.

### D-Q6: Hard delete or deprecation window for the retired plugin surfaces?
Label: POLICY-COVERED
Authority anchor: this SPEC's hard rules — spec.md REQ-016 (shall no longer
generate, commit, or ship the plugin artifacts) + C5 (atomic retirement, no
orphaned drift check), which codify card decision D4 (no carrier, "처분
포함"). Same caveat as P1.
Why unresolved: n/a — closed. The deprecation-window option keeps the carrier
generating, committing, or shipping and violates REQ-016, C5, and AC-013;
hard delete is the only surviving option, recorded as premise P5 in spec.md
§1.
Operator verdict: closed by constraint — hard delete (M6 disposition list;
release-chain gates included per iter1-D4 — the iter1 audit defect list's
release-gate finding, NOT operator decision D4; identifier prefixed iter2
D21).

## Open founder questions (operator verdict pending)

### D-Q1: Which five agents are the "핵심 에이전트 5종" of L0?
Label: FOUNDER
Authority anchor: n/a — no committed artifact enumerates the five.
Why unresolved: catalog core tier carries 11 agents (sync-auditor,
manager-develop, manager-docs, manager-git, manager-lead, manager-spec,
plan-auditor, super-advisor, manager-todo, manager-design, e2e-tester; plus
builder-harness as harness-generated — research V14). Candidate reading A:
manager-spec, manager-develop, manager-docs, plan-auditor, sync-auditor (the
plan→run→sync chain with its two auditors). Candidate readings B/C include
manager-git or manager-lead. The card says "핵심 에이전트 5종" without naming
them. BLOCKS M0/M1.
Operator verdict:

### D-Q2: Where does the per-user manifest live, and what is it called?
Label: FOUNDER
Authority anchor: n/a — the location is new surface.
Why unresolved: candidates with different trade-offs: `~/.moai/user-assets.json`
(leading — `~/.moai/` is already moai's user-level state home per
`internal/paths/paths.go:91`), `~/.claude/moai-manifest.json` (next to the
assets it tracks, but Claude-only), per-root split files (no single source).
BLOCKS M0/M1.
Operator verdict:

### D-Q4: Does "plan·run·sync" in the L0 definition name the published command skills, the workflow skills, or both?
Label: FOUNDER
Authority anchor: n/a — the card's wording is not disambiguated by any
committed artifact.
Why unresolved: reading (a) `moai-plan`/`moai-run`/`moai-sync` published
command skills (literal adjacency to agents/hooks/factory in the card);
reading (b) `moai-workflow-spec`/`-tdd`/`-ddd` workflow skills; reading (c)
both. Research V15. Feeds M0.
Operator verdict:

### D-Q5: What is the bundle granularity for non-L0 assets?
Label: FOUNDER
Authority anchor: n/a — the card says only "나머지는 선택 묶음".
Why unresolved: (a) the six existing optional packs stand as-is and
current-`core` remainders re-bundle by theme, (b) remainders join existing
packs where topical, (c) one single "extended" bundle. Affects catalog shape
and the update removal path. Feeds M0.
Operator verdict:
