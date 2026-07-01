# Implementation Plan — SPEC-V3R6-DOCS-V3-REBUILD-001

## §A Context

Full rebuild of `docs-site/` (adk.mo.ai.kr) to v3.0.0-rc4 accuracy. Content-only on the existing `hugo-geekdoc` theme. 380 content files (95/locale × 4), plus README drift correction and `data/menu/main.yaml` IA rebuild. Canonical locale is `ko`; propagation chain is ko → en → ja/zh (§17.3).

Tier: **L** (thorough). Justification: IA redesign spanning 13 content sections; 380-file rewrite; a research-backed 112-file CC-mirror refresh requiring external Claude Code documentation; 16 net-new pages; cross-cutting 4-locale parity obligation. This exceeds Tier M envelope on every axis (domains, files, external-research dependency, cross-locale coordination).

## §B Known Issues (drift inventory feeding the rebuild)

| Surface | Drift | Target REQ |
|---------|-------|-----------|
| `hugo.toml` `params.version` | `v3.0.0-rc2` | REQ-DVR-001 |
| `content/ko/_index.md` line 10 | "릴리스 후보 2" | REQ-DVR-001 |
| `getting-started/introduction.md` L131 | "34,220줄 Go 코드, 32개 패키지" | REQ-DVR-001, -004 |
| `getting-started/introduction.md` L133/156/163 | "32개 스킬" | REQ-DVR-004 |
| `getting-started/installation.md` L11/91/157 | `v3.0.0-rc2` examples | REQ-DVR-001 |
| `README.md` / `README.ko.md` L40/62/64 | "30 `moai-*` skills" | REQ-DVR-013 |
| `README.md` / `README.ko.md` L309/319 | "12 total" but lists 13 | REQ-DVR-013 |
| `README.md` / `README.ko.md` L584/585 | `coverage` / `e2e` rows | REQ-DVR-013 |
| `README.ja.md` / `README.zh.md` (worse — verified by grep) | retired `/moai design` Design System section ~L926-1061; `/agency → /moai design` v2.12.0 refs L57/L63; `/moai coverage` in active workflow chains L623/L633; `coverage`/`e2e` rows L563/564; stale `v2.6.0`-as-current | REQ-DVR-013 |
| `advanced/builder-agents.md` | obsolete 3-builder model + "32개 스킬" | REQ-DVR-010 |
| `advanced/agent-guide.md` L124 | garbled `manager-develop` ×6 + stray archived refs | REQ-DVR-011 |
| `workflow-commands/moai-harness.md` | v3-learning vs v4-Builder reconcile | REQ-DVR-012 |
| `claude-code/` mirror (112 files) | predates latest CC feature set | REQ-DVR-008 |
| `data/menu/main.yaml` | omits `cost-optimization`; empty `contributing` sub | REQ-DVR-007, -016 |

## §C Pre-flight (run-phase entry preconditions)

1. IA blueprint (design.md §A) approved by user at the M0 gate.
2. v3.0 fact-sheet (research.md §1) confirmed against live codebase.
3. Skill count reconciled from template source and locked as the single doc-facing number.
4. Latest Claude Code research completed (M0 research task) — its output is a precondition for M2 CC-mirror rewrite.

## §D Constraints

- Content-only; no theme/CSS/layout change (REQ-DVR-014).
- 4-locale parity mandatory, same-milestone (REQ-DVR-015).
- Mermaid TD only; emphasis-paren spacing; no emoji; glossary untranslated; forbidden-URL blacklist (NFR-DVR-001..005).
- `latest` only, no version snapshot (NFR-DVR-007).
- Vercel binding unchanged (NFR-DVR-006).

## §E Self-Verification

Plan-phase self-verification is recorded in `progress.md` §E.1. Run-phase and sync-phase evidence are populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4).

## §F Milestones

The rebuild is sequenced across 5 milestones (M0-M4). Priority-based ordering; no time estimates.

### M0 · Foundation (Priority: High — gate before all rewrite work)

Establish the factual and structural baseline.

- **M0.1** Extract the v3.0 fact-sheet from the live codebase (commands, agents, skills, version, lifecycle, scale) — record in research.md §1.
- **M0.2** Reconcile the skill count from the template source `internal/template/templates/.claude/skills/` (derive the single doc-facing number = 27 `moai-*`). Lock it as the value used everywhere.
- **M0.3** Research the latest Claude Code feature set (execute the research.md §3 PLAN — target URLs at `code.claude.com/docs` + release notes; note the dev-only `harness-release-update` pattern as an optional acquisition path). Output: a per-CC-page delta note driving M2.
- **M0.4** Design the new IA blueprint / sitemap (design.md §A) reconciling `data/menu/main.yaml` with actual content sections.
- **M0.5** Fix README drift across ALL 4 README files (`README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`): remove retired subcommand rows (coverage/e2e/design), correct 12 → 13 command count, skill count 30 → 27. `README.ja.md` / `README.zh.md` additionally require removal of the retired `/moai design` Design System section (~L926-1061), the `/agency → /moai design` v2.12.0 references (L57/L63), `/moai coverage` in active workflow chains (L623/L633), and reconciliation of stale `v2.6.0`-as-current references (REQ-DVR-013; 4-locale parity per REQ-DVR-015 — ja/zh MUST NOT be deferred).
- **M0.6** Update the version SSOT: `docs-site/hugo.toml` L55 `params.version` → `v3.0.0-rc4` + L56 `releaseDate`. This is the single source the `{{< version >}}` shortcode binds to (design.md §B.2); M1.1 ko content and M4.1 verification both depend on this edit landing here. This bullet OWNS the hugo.toml version surface of REQ-DVR-001 (M4.1 only verifies it).
- **Gate**: IA blueprint + fact-sheet approved (design.md §A + research.md §1 signed off). CC research delta note present.

### M1 · ko core rewrite (Priority: High)

Rewrite the MoAI-ADK-specific ko pages (validate-then-rewrite) + author the 4 new pages.

- **M1.1** Validate-then-rewrite ~67 ko MoAI-ADK-specific pages (getting-started, core-concepts, workflow-commands, utility-commands, quality-commands, multi-llm, db, guides, worktree, cost-optimization, advanced, contributing). Sections flagged accurate reuse their content as the draft baseline.
- **M1.2** Author 4 new ko pages: Decision Memory (`moai preference`), `moai inventory`, Harness v4 Builder, `/effort ultracode` dynamic workflows.
- **M1.3** Rewrite the 3 ko drift pages: `advanced/builder-agents.md` (Harness v4), `advanced/agent-guide.md` (fix garble + dynamic-team spawning), `workflow-commands/moai-harness.md` (v3/v4 reconcile).
- **Gate**: all ko pages pass local fact-check against the M0 fact-sheet; ko `hugo` build clean for the ko locale.

### M2 · ko CC-mirror rewrite (Priority: Medium)

- **M2.1** Refresh the ~28 ko `claude-code/` pages to the latest Claude Code, using the M0.3 research delta note as the source of truth.
- **M2.2** Ensure every CC page cites the canonical Claude Code doc surface it mirrors (traceability back to research.md §3).
- **Gate**: ko CC-mirror pages reflect research-backed current CC behavior; no fabricated feature claims (verification-claim-integrity).

### M3 · Multi-locale propagation (Priority: Medium — largest-volume)

Propagate ko → en → ja/zh per §17.3.

- **M3.1** Propagate M1+M2 ko changes to `en` (translation + fact parity).
- **M3.2** Propagate to `ja` and `zh` in parallel.
- **M3.3** Rebuild `data/menu/main.yaml` 4-locale entries for all new/moved pages (16 new-page entries + IA reorg).
- **Gate**: file count/path parity across 4 locales; menu entries resolve to existing pages in every locale.

### M4 · Build + parity + deploy verify (Priority: High — closure gate)

- **M4.1** `cd docs-site && hugo --minify` → zero warnings; `public/sitemap.xml` generated; `v3.0.0-rc4` version string present.
- **M4.2** `scripts/docs-i18n-check.sh` → pass (4-locale parity, non-empty titles, H1 present, glossary preserved).
- **M4.3** §17.6 deploy verification: Mermaid TD render, language switcher navigation, Vercel preview 4-locale home pages, robots/llms domain match.
- **M4.4** Forbidden-URL scan (NFR-DVR-001) across `docs-site/` → zero matches.
- **Gate**: all verification gates green; Definition of Done (acceptance.md) satisfied.

## §G Anti-Patterns

- **AP-DVR-001** — Blank-page rewrite of already-accurate sections (wastes effort; violates REQ-DVR-006 validate-then-rewrite).
- **AP-DVR-002** — ko-only merge without en/ja/zh in the same milestone (violates REQ-DVR-015; §17.3 forbids ko-only merge).
- **AP-DVR-003** — Deriving the skill count from the local dev project (27 moai-* + 2 harness-*) instead of the template source (violates REQ-DVR-004; the 2 `harness-*` are user-owned and NOT shipped).
- **AP-DVR-004** — Claiming a Claude Code feature in the CC mirror without research-backed evidence (violates verification-claim-integrity; fabricated feature claims).
- **AP-DVR-005** — Touching theme/CSS/layout "while I'm here" (violates REQ-DVR-014 content-only scope).
- **AP-DVR-006** — Creating a `content/{locale}/v3/` snapshot (violates NFR-DVR-007; snapshot deferred to GA).
- **AP-DVR-007** — Introducing a forbidden URL or Mermaid `LR` diagram during rewrite (violates NFR-DVR-001/002).
- **AP-DVR-008** — Performing the CC research during plan-phase (violates REQ-DVR-020; plan-phase captures the PLAN only).

## §H Cross-References

- `spec.md` — requirements (REQ-DVR-001..020, NFR-DVR-001..007).
- `acceptance.md` — per-REQ acceptance criteria + Given-When-Then.
- `design.md` — IA blueprint, content-model, validate-then-rewrite strategy, 4-locale propagation design.
- `research.md` — v3.0 fact-sheet, drift inventory, CC-latest research PLAN.
- `.moai/docs/docs-site-i18n-rules.md` §17 — parity/URL/Mermaid/build/deploy doctrine (SSOT).
