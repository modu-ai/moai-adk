# progress.md — SPEC-DESIGN-DOCSV2-001

> Plan-phase skeleton only. Run-phase (manager-develop) populates §E.2/§E.3; sync-phase (manager-docs) populates §E.4.

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID: SPEC-DESIGN-DOCSV2-001
- Status: draft (plan-phase artifact set complete)
- Files authored: spec.md, plan.md, acceptance.md, research.md, design.md, progress.md
- Tier: L | era: V3R6 | harness: thorough
- Plan-phase self-check: SPEC ID regex PASS (`SPEC-DESIGN-DOCSV2-001`); 12-canonical-field frontmatter validated on spec.md; Out of Scope section present (6 `### Out of Scope — <topic>` H3 sub-headings); moai-brand.css unfreeze authorization recorded (spec.md §G).
- Ready for: plan-auditor independent audit → Implementation Kickoff Approval (plan→run HUMAN GATE).

## §E.2 Run-phase Evidence

### M1 — Token Unification (commit 19c40097d, PUSHED to main)

**AC Matrix (AC-TOK-001..007 — M1 scope):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-TOK-001 | PASS | `grep -rn '#faf9f5\|#ecefee' docs-site/static/ docs-site/layouts/` | 0 matches; `--color-bg: #f4f4f4` confirmed (brand.css L15) |
| AC-TOK-002 | PASS | `grep -n '^\s*--color-primary' docs-site/static/moai-brand.css` | `#3d7d5f` / `#316750` / `#265240` confirmed (L11-13) |
| AC-TOK-003 | PASS | `grep -rn '#211A14' docs-site/static/ docs-site/layouts/` | 0 matches; `--color-ink: #060606` confirmed (brand.css L14) |
| AC-TOK-004 | PASS | `grep -rn 'linear-gradient(135deg, #3d7d5f' docs-site/static/ docs-site/layouts/` | 0 matches; `--gradient-signature: #3d7d5f` solid (brand.css L51) |
| AC-TOK-005 | PASS | `grep -rn '#000000' brand.css design.css layouts/` (comment-excluded) | 0 matches |
| AC-TOK-006 | N/A (M1) | structural CSS inspection | Not regressed by M1 — token-value repoint only; no new selector blocks with gradient+shadow simultaneity introduced |
| AC-TOK-007 | PASS | `grep -n '^\s*--neutral-' docs-site/static/moai-brand.css` | All achromatic: #f7f7f7 → #060606 scale (L19-29) |
| AC-BLD-001 | PASS | `cd docs-site && hugo --minify --gc` | exit 0, 0 WARN/ERROR; KO 153p / EN 150p / JA 139p / ZH 150p; 2506ms |

**Files touched (7):**
- `docs-site/static/moai-brand.css` — :root rewritten to v2 SSOT raw tokens + 4 literal sweeps + `--color-primary-hover` corrected `#31684f→#316750` (AC-TOK-002)
- `docs-site/static/moai-docs-tokens.css` — clay/cream/ink scales repointed to brand.css v2 raw tokens (cycle-safe, one-directional); semantic aliases repointed; MaruBuri @font-face preserved (M2 typography)
- `docs-site/static/moai-docs-theme.css` — :root rewritten to pure v2 alias layer (drops ALL `--color-*`/`--neutral-*`/`--fg-*`/`--border-*` overrides — brand.css SSOT stands; CSS custom-property cycle broken); component literals swept
- `docs-site/static/moai-design.css` — `#faf9f5→var(--color-bg)` (3 occ), `#181715→var(--neutral-900)` (2 occ)
- `docs-site/layouts/partials/foot.html` — mermaid themeVariables `#faf9f5→#f4f4f4` (AC-TOK-001 `docs-site/layouts/` scope)
- `docs-site/assets/css/moai-brand.scss` — `git rm` (stale uncompiled SCSS, not referenced by head/custom.html)
- `.moai/specs/SPEC-DESIGN-DOCSV2-001/spec.md` — frontmatter `status: draft → in-progress`

**PRESERVED (unchanged):** `head/custom.html` (FNV32a content-hash cache busting auto-regenerates on `hugo --minify`), `moai-docs-theme.js` (TOC scroll-spy only — no theme-remap code), MaruBuri `@font-face` blocks in tokens.css (M2 typography scope).

**Deferred debt (sync-phase / M4 / out-of-scope):**
- `design.css` `#252320` (code-card gradient midpoint — warm but not AC-checked)
- `design.css`/`theme.css` `[data-theme="dark"]` dead-code blocks containing raw `#3d7d5f` (dark theme is dead per CLAUDE.local.md §17.1 — light-only single theme)
- `tokens.css` `--surface-dark: var(--neutral-900)` / `--charcoal: var(--neutral-900)` (warm-neutral scale residual, now repointed to neutral-900)
- `foot.html` mermaid warm literals `#141413` / `#d6ebde` / `#efe9de` (AC-MER-001 M4 scope — full mermaid v2 palette migration)
- PWA `manifest.json` / `site.webmanifest` `#000000` (outside AC-TOK-005 grep scope: `docs-site/static/moai-*.css docs-site/layouts/`)
- Physical 4-CSS-file collapse into single `tokens.css` (deferred to sync-phase per lean M1 approach — functional v2 effect achieved via `:root` repoint + literal sweep, not file consolidation)

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase)_

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase)_

## §F Phase 4 Mode Selection

- Tier: L | era: V3R6 | harness: thorough
- Input params: scope ~12 files (CSS + Hugo layouts + partials + mascots + i18n); domain count 6 (tokens / typography / layout / mermaid / mascot / i18n); file mix CSS+HTML+JS+YAML; concurrency benefit LOW (coding-heavy implementation).
- Mode evaluation: Mode 1 (trivial) — no; Mode 2 (background) — no (write work); Mode 3 (agent-team) — RETIRED; Mode 4 (parallel) — no (coding-heavy per Anthropic coding-task parallelism caveat, not research-heavy); Mode 6 (workflow) — no (semantic multi-rule transform with inter-file dependencies, not mechanical-uniform); Mode 5 (sub-agent) — selected.
- Decision: sub-agent
- Justification: coding-heavy CSS/layout migration with tight inter-file coupling (the unified token layer is consumed by every layout + component override). Anthropic coding-task parallelism caveat → sequential single sub-agent. Tier L → full Section A-E delegation template; manager-develop owns the milestone sequence with per-milestone commits, Route A main-direct.
- Run-phase: manager-develop, cycle_type=ddd (existing-CSS/layout behavior-preserving refactor; characterization baseline = current Hugo build 0-warnings + token-literal grep state + 4-locale rendered snapshot, per plan.md §G).
- Implementation Kickoff Approval: GRANTED (user approved run-phase entry 2026-07-16). plan-auditor independent audit did NOT produce a verdict (technical stall, 600s no-progress); orchestrator verified REQ/AC inventory directly (44 REQ / 9 groups / 30 AC, internally consistent) and proceeded per user direction. Resolved clarifications: mono-font = two-token split (--font-mono JetBrains + --font-code Goorm Sans Code); mermaid = stay v10 + v2 themeVariables only.
