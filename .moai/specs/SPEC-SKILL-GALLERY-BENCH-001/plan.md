# plan.md — SPEC-SKILL-GALLERY-BENCH-001

Derived implementation plan. The SPEC (`spec.md`) owns the requirements; this file
owns sequencing, touch points, and risks. Ordered by decision-reversibility: the
benchmark protocol decisions (§F M1) are the most likely to need rework and come
first; docs placement (M2) second; mechanical verification (M3) last.

## §A Context

- Kanban card t272, Class C (design change spanning benchmark + docs), Tier M.
- Worktree `.claude/worktrees/t272`, branch `WT-skillstead-gallery`, base `origin/main`.
- All paths below are worktree-relative (CWD = worktree root).
- Benchmark subject: `.claude/skills/moai-domain-svg-infographic/` —
  SKILL.md (workflow, dials, gates), `references/archetypes.md` (A1-A4 skeletons and
  per-archetype ceilings), `references/authoring.md` (formula set), 
  `scripts/check-svg.mjs` (lint), `scripts/render.mjs` (2x PNG).

## §B Known Issues

1. **Four archetypes vs nine forms**: `archetypes.md` closes with "a request that
   resists all four is usually two diagrams — split it rather than inventing a fifth
   shape". Several TypePack forms (notably `before-after`, `cards-kpi-grid`,
   `roadmap-timeline`, `nested-scope`) plausibly resist all four. This is expected
   and is precisely what the failure taxonomy exists to classify — a `NOT-PRODUCIBLE`
   verdict with a `structural-limit` classification is a valid, deliverable outcome,
   not a benchmark failure.
2. **Chromium availability**: `render.mjs` resolves a browser from `CHROME_PATH`,
   well-known install locations, then `PATH`. Exit 2 (no browser) is the degradation
   signal; per REQ-SGB-013 that is a blocker, not a partial pass.
3. **hugo.toml version lag**: hugo.toml reads `v3.1.2` while release v3.1.3 is in
   flight elsewhere. The version-sync check compares displays against hugo.toml as
   SSOT — the emphasis content must not introduce version displays at all, which
   sidesteps the race.

## §C Pre-flight (before any artifact)

1. `node --version` → >= 18.
2. Browser probe: run `render.mjs` over one bundled fixture
   (`scripts/fixtures/a11y-present.svg`) into `/tmp` → exit 0, browser + version
   printed. This proves G3 is measurable before nine forms are authored.
3. Confirm evidence dir target `.moai/reports/t272/` is absent (run-phase creates it).
4. `git status --short` in the worktree → clean baseline before generation.

## §D Constraints (restated from spec.md §E — binding)

- No edits to the skill directory or its template mirror (REQ-SGB-008).
- oss-docs HARD i18n rules for M2 (Skill `hns-oss-docs-i18n-rules` is the digest;
  `.moai/docs/docs-site-i18n-rules.md` is the SSOT).
- No fabricated gate results (skill's degradation contract).
- English for SPEC/verdict artifacts; ko authored first for docs content.

## §E Self-Verification (run-phase obligation)

- Every gate claim in `verdict.md` cites a log file under `logs/` whose verbatim
  content shows the command and its exit status (verification-claim-integrity:
  claim → evidence → baseline-attribution).
- Docs claims cite artifact paths that resolve on the branch.
- `git diff --stat` over the skill dirs is empty at every commit point.

## §F Milestones

### M1 — 9-form benchmark execution + verdict table (High)

Per form, in the pinned order of spec.md §D.2:

1. Invoke Skill("moai-domain-svg-infographic"); route per its Step 0 table (all nine
   briefs are image-deliverable freeform infographics — the SVG route has an unopposed
   reason by construction of the briefs).
2. Frame + settle the four dials (defaults pinned in spec.md §D.2; deviations
   recorded).
3. Run the numeric layout pass; pick an archetype (or document that none fits —
   that observation IS the `structural-limit` evidence candidate).
4. Author `artifacts/<form>.svg`; run G2 (`check-svg.mjs`, output to
   `logs/<form>-lint.txt`); run G3 (`render.mjs`, output to
   `logs/<form>-render.txt`, PNG to `artifacts/<form>.png`).
5. Record the verdict row (verdict, archetype used, G1-G3, evidence paths,
   deviation sentence, taxonomy classification when not `PRODUCIBLE`).

Deliverable: `.moai/reports/t272/verdict.md` complete with 9 rows + evidence tree.

### M2 — docs-site + README 4-locale emphasis (Medium)

Exact touch points (measured on the worktree):

| Surface | File(s) | Change |
|---------|---------|--------|
| docs-site | `docs-site/content/ko/advanced/skill-guide.md` (Domain skill table, current row: `moai-domain-svg-infographic` at ~line 158) | Expand the single table row into a `### SVG 인포그래픽 — 생성 가능한 다이어그램 종류` subsection (or equivalent placement inside the existing Domain section) listing the measured `PRODUCIBLE` forms, each citing its evidence path |
| docs-site | `docs-site/content/{en,ja,zh}/advanced/skill-guide.md` | Derived translations of the same subsection, same PR |
| README | `README.ko.md` — host section `## 핵심 기능` (line 306); the svg-infographic mentions today are the skill-list lines ~406 and ~739 | Add the emphasis to `핵심 기능` (ko authored first), citing measured forms |
| README | `README.md`, `README.ja.md`, `README.zh.md` | Minimal derivation of the touched section only |

Rules: no new H2 in any README (keeps 4-file H2 parity trivially intact); docs-site
H3 addition must land in all 4 locales the same PR; no body emoji; no version
displays; no URLs outside `adk.mo.ai.kr`.

Placement decision latitude: the exact anchor (expand table row vs add H3 under the
Domain section) is the run-phase author's call within the files named above; any
placement that adds a page or moves one triggers the vercel.json redirect rule and is
out of scope (spec.md §G).

### M3 — verification + PR readiness (Medium)

1. Execute Skill("hns-oss-docs-verify") in full; record per-check results.
2. README parity checklist from Skill("hns-oss-docs-readme-sync") (H2 counts, H3
   counts in touched sections, switcher header, blacklist grep).
3. Confirm `git diff` empty on both skill dirs.
4. PR body: card id t272 in the title (card-delivering PR contract), summary of the
   verdict table, evidence path pointer.

## §G Anti-Patterns

| Anti-pattern | Correct approach |
|---|---|
| Listing a diagram type in docs because the archetype "should" cover it | Only `PRODUCIBLE`-verdict forms are listed, each with its artifact path (REQ-SGB-009) |
| Fixing a failing form by editing the skill mid-run | Record the failure + taxonomy; skill edits are out of scope (REQ-SGB-008) |
| Reporting a lint/render verdict from memory or a prior form's run | Each form's gates re-run and logged individually |
| Downgrading a `NOT-PRODUCIBLE` to `PARTIAL` to make docs look better | The verdict vocabulary is defined in spec.md §B; degradation of verdicts corrupts the taxonomy |
| Authoring docs content first, benchmark later | M1 precedes M2 strictly; the emphasis is downstream of evidence |
| Translating docs by rewriting untouched prose | Minimal-diff derivation of touched sections only |

## §H Cross-References

- `spec.md` (this SPEC) — requirements, pinned briefs, gates, taxonomy.
- `.claude/skills/moai-domain-svg-infographic/SKILL.md` — the benchmark subject and
  the gate definitions (G2/G3 commands, degradation contract).
- `.claude/skills/moai-domain-svg-infographic/references/archetypes.md` — A1-A4 and
  the "resists all four" rule the taxonomy leans on.
- `.claude/skills/hns-oss-docs-i18n-rules/`, `hns-oss-docs-readme-sync/`,
  `hns-oss-docs-verify/` — the M2/M3 rule digests.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the no-unobserved-claim
  invariant behind §E of this plan.
- Kanban card t272 (queue: `moai todo`); lead's analysis report 2026-08-25 for
  SkillStead provenance.
