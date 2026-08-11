# research.md — SPEC-DESIGN-DOCS-V31-001

> Read-only investigation base. This file does NOT dictate implementation; it provides the verified evidence on which spec.md / plan.md / acceptance.md rest.

---

## §A. Handoff analysis (Claude Design export)

### §A.1 Handoff README — tone, copy, color, type, motion

The handoff ships a 14KB README at `_ds/ai-design-system-.../README.md` codifying the brand voice and visual system. Load-bearing findings (all verified by direct read at plan time):

> **Provenance note (v0.2.0 D6 fix).** The handoff README opens with "모두의AI (mo.ai.kr) 는 한국의 AI 사용자가 모이는 커뮤니티 플랙폼" and lists community-platform product surfaces (`/projects`, `/beta`, `/news`, `/academy`) and 6 upcoming consumer services. This is the **community-platform** voice — wider than the docs-site surface. The brand-wide rules (tone, color, type, motion, banned-vocabulary) ARE applicable across both surfaces, and REQ-KO-004 (renumbered from REQ-KO-005 at v0.1.0) applies the SUBSET relevant to docs-site: the voice/register rules and the banned vocabulary are imported; the community-platform-specific product copy ("모두의 사주", "바닐라 바게트", the `/beta` mention) is NOT imported into docs-site content. M0 records this scoping decision in the `_meta.yaml` rationale.

- **Tone.** "신뢰감 있고 전문적이며, 동시에 따뜻하고 포용적". The word "모두의" appears at least once per page. Beta-tester addressed as "동료/주체", not passive "user".
- **Register.** Body `~합니다`; CTA / speech-bubble / event copy `~해보세요 / ~할까요?` permitted.
- **First-mention jargon expansion.** `에이전트(스스로 일하는 AI)` — expanded once, abbreviated after.
- **No body emoji.** Mascot is the emotional anchor. Lucide icons replace decorative emoji.
- **Banned vocabulary.** 혁신적인, leverage, 솔루션, Game-changing, Cutting-edge, Next-level, Disruptive, 절대로, 유일한, 최고의, "지금 안 하면 평생 후회", 핵개꿀. (Codified into spec.md REQ-KO-005.)
- **Color.** Single core `#3d7d5f` (mascot sweater green). Ink `#060606`, mid-gray `#9fa0a0`, light `#e6e6e6`. Magenta/purple/orange gradients explicitly forbidden. `#000000` forbidden (use `#09110f`). Max 4 color tokens simultaneously on screen.
- **Type.** Pretendard single Korean font + Inter (Latin) + JetBrains Mono (code). 9 weights self-hosted. Letter-spacing: titles `-0.05em ~ -0.075em`, body `-0.025em ~ -0.05em`.
- **Motion.** Default `150–250ms cubic-bezier(0.4,0,0.2,1)`. Mascot-only bounce `cubic-bezier(0.34,1.56,0.64,1)`. Page transitions `cubic-bezier(0.16,1,0.3,1)` 600ms. `prefers-reduced-motion: reduce` degrades all transitions to 1ms.
- **Layout.** 12-column grid (≥1024px), 4-column (<768px). Container max 1440px. Content 1024px. Vertical rhythm 80/64/48px section gaps. Sticky header 64px.
- **Cards.** Surface `#ffffff`, `radius 16px`, `padding 24px`, `shadow.sm`, `border 1px #d4d4d4`. Hover `translateY(-2px)` + `shadow.md`.
- **Icons.** Lucide. Stroke-width 1.75 default, 2 emphasized. `currentColor` by default; `--color-primary` for emphasis only.

### §A.2 Screen 01 — Docs Home

The home page prototype defines the IA surface. Key structural elements:

- **Hero.** Left column: section label "MOAI-ADK · AGENTIC DEVELOPMENT KIT" → headline ("토크노믹스를 위해 설계된 하네스", 56px black weight, -0.05em tracking) → sub-headline (18px, 1.6 line-height) → dual CTA pill ("시작하기 →" primary green, "빠른 시작 가이드" outline) → install-command dark card (`#141414` bg, monospace, copy affordance). Right column: mascot `MoAI-Mascot-Explaining.png` 260px + caption "SINGLE BINARY · GO · APACHE-2.0".
- **3-card value grid.** Three cards: 01 TOKENOMICS, 02 SELF-LEARNING, 03 HARNESS. Each card has a monospace section label (green), an h3 Korean title, a 13.5px description, and inline links.
- **Section grid ("문서 구조").** 12-card grid (NOT 13). Each card has a Korean title + monospace English caption + 12.5px description. Right-aligned monospace label: "12 SECTIONS · KO / EN / JA / ZH".
- **Book CTA card.** Large card: mascot `MoAI-Mascot-Coffee.png` 120px + book metadata ("OFFICIAL BOOK · 488P") + title + description + outline pill "book.mo.ai.kr →".

**IA reconciliation finding.** The live `ko` tree has 13 sections (`_meta.yaml`); the handoff home grid shows 12. The 12-vs-13 delta is a design-time vs runtime mismatch. **[RESOLVED at v0.2.0 per user decision: 12 sections is canonical. The live 13th section `changelog` MOVES TO THE SITE FOOTER (becomes a footer link, not a sidebar section); `cost-optimization` RETAINS its own card per the handoff home grid. M0 implements this exact 12-section shape; no longer an open decision.]**

### §A.3 Screen 02 — Getting Started (section index)

The "섹션 인덱스 — 시작하기" prototype is the canonical section-index template. Structure:

- **Breadcrumb.** `문서 / 시작하기`.
- **Section header.** Monospace section label "SECTION 01 · GETTING STARTED" (green, 0.14em tracking) → h1 40px black weight → 16px introduction prose → info callout (green-tinted bg) with inline cross-links.
- **권장 읽기 순서.** Numbered card list (1–9), each card: number pill + title + description + chevron-right.
- **학습 흐름.** Horizontal flow diagram (box-and-arrow), 6 nodes: 소개 → 설치 → moai init → 빠른 시작 → 업데이트 → CLI·FAQ. Rendered as styled HTML boxes (NOT Mermaid — a static infographic).
- **다음 단계 card.** Mascot `MoAI-Mascot-Pointing.png` 64px + prose with inline link.
- **Footer nav.** Prev / Next pill cards.

This prototype is the visual target for ALL section index pages (not just getting-started).

### §A.4 Screens 03–06

- **03 Doc Detail** — the canonical article page. Header + body + TOC rail + footer nav. The body uses h2/h3 headings with `-0.05em` tracking; code blocks carry the macOS dark-card styling established in `render-codeblock.html`.
- **04 CLI Reference** — CLI command reference page. Monospace command headings, flag tables, example blocks.
- **05 Search** — search results page. The visual styling of the search affordance is ported to `site-header.html`; the search backend remains geekdoc built-in.
- **06 Not Found** — 404 page. Mascot `MoAI-Mascot-Searching.png` + friendly copy + home link.

### §A.5 Dark-mode divergence

The handoff prototypes include a dark-mode toggle button (`aria-label="다크 모드"`) in the header and a `[data-theme="dark"]` token block in `colors_and_type.css`. CLAUDE.local.md §17.1 mandates light-only. The divergence is explicit and user-approved:

- The dark-mode toggle button is NOT ported to `site-header.html`.
- The `[data-theme="dark"]` CSS block is NOT ported to production CSS.
- REQ-DS-003 codifies this; research documents it for traceability.

### §A.6 `_ds_bundle.js` — runtime viewer, not production

The handoff `_ds/` directory contains `_ds_bundle.js` (132KB), `_ds_manifest.json` (20KB), `_adherence.oxlintrc.json` (9KB), and `support.js` (66KB at the project root, loaded by every `.dc.html`). These are the prototype viewer's runtime — they render `<x-dc>` tags, `<sc-for>` loops, `style-hover` attributes, and `<script type="text/x-dc">` component classes inside the design-tool preview. They are NOT part of the design system; they are the tool that displays the design system.

Production target: Hugo + hugo-geekdoc. The Hugo templates use Go template syntax, partials, and shortcodes — none of which require `_ds_bundle.js`. The token CSS (`colors_and_type.css`) is the load-bearing artifact; the runtime is not. REQ-BL-003 forbids shipping the runtime.

---

## §B. Current docs-site content baseline

### §B.1 Section inventory (ko, verified at plan time)

| Section | Pages | Weight | Notes |
|---|---:|---:|---|
| getting-started | 8 | 10 | Onboarding path. Highest reader impact. |
| core-concepts | 7 | 20 | Three-pillars mental model. |
| workflow-commands | 7 | 30 | `/moai plan|run|sync` etc. |
| utility-commands | 12 | 40 | |
| cli-reference | 21 | 45 | Largest reference surface. |
| claude-code | 30 | 55 | Largest section. |
| multi-llm | 3 | 60 | |
| cost-optimization | 2 | 70 | Smallest content section. |
| guides | 3 | 85 | |
| worktree | 4 | 90 | |
| advanced | 24 | 100 | Second-largest. Target for v3.1 features. |
| contributing | 1 | 110 | |
| changelog | 1 | 120 | |
| **Total** | **124** | | |

### §B.2 Content depth sample (3 pages)

Sampled at plan time to establish the pre-SPEC baseline:

1. **`getting-started/_index.md`** — ~250 words of body prose. Opens with a callout (emoji + 소속 가값), then one paragraph, then a callout, then a Mermaid `flowchart TD`, then a 9-row table. Verdict: **reference-grade** — table-and-callout-heavy, concept exposition thin. Fails the §A.1 pillar 1 (concept-first exposition — the first content is a callout, not prose).
2. **`core-concepts/_index.md`** — ~180 words of body prose. Opens with a callout, an image, one paragraph, a callout, a table, a Mermaid diagram. Verdict: **reference-grade** — same pattern.
3. **Spot-check of a workflow-commands page** — similar pattern: callout → short prose → table.

**Baseline finding.** The current Korean content is uniformly reference-grade. The M4 rewrite is a genuine depth lift, not a polish. The 600/700/800-word floors in acceptance.md §A.2 are calibrated against this baseline — most pages currently sit well below.

### §B.3 Design baseline

The live `moai-brand.css` header confirms (verified by direct read):

```
FROZEN (re-stamped at sync-phase close, SPEC-DESIGN-DOCSV2-001, 2026-07-16):
  the v2 token vocabulary below (neutral canvas #f4f4f4 + solid green accent
  #3d7d5f) is now the frozen baseline.
```

The `:root` block carries `--color-primary: #3d7d5f`, `--color-ink: #060606`, `--color-bg: #f4f4f4`. These match the handoff's primary tokens exactly — the PRIMARY/INK/BG trio is stable across the v2 → v2-renewal transition. The neutral ramp is the delta surface (see §D).

---

## §C. v3.1 feature catalog (verified)

Each row verified by reading the SPEC's `spec.md` frontmatter `status:` and the CHANGELOG `[Unreleased]` section. Verification method: `grep -m1 "^status:" .moai/specs/<SPEC-ID>/spec.md`.

| Feature | SPEC | Status | Verified user-facing? | NEW page required? |
|---|---|---|---|---|
| `/moai goal` infinite-duration | SPEC-INFINITE-GOAL-001 | completed | YES (user-facing `moai goal arm` flags) | YES (workflow-commands) |
| Factory Mode | SPEC-FACTORY-MODE-001 | completed | YES (`-f` entry switch, user-facing) | YES (advanced) |
| BAS Navigator 3-tier sync | SPEC-PROJECT-NAVIGATOR-001/002/003 | completed | YES (`moai codemaps` surface, `--audit`) | YES (advanced) |
| `manager-lead` hierarchical team | SPEC-HIERARCHICAL-TEAM-001 | completed | YES (12th agent, user-facing delegation) | YES (advanced) |
| Multi-model audit | SPEC-AUDIT-MULTI-MODEL-001 | completed | YES (`audit_model: multi` config) | YES (advanced) |
| `MOAI_AUTONOMY_TIER` | SPEC-AUTONOMY-TIERS-001 | completed | YES (env var, user-facing) | YES (advanced) |
| `profile` matrix | SPEC-MODEL-PROFILE-MATRIX-001 | completed | YES (`--profile` flag, `moai model profile`) | NO (rewrite multi-llm + cli-reference) |
| Per-agent model enforcement | SPEC-AGENT-MODEL-ENFORCE-001 | in-progress | PARTIAL (pre-server hook exists; UI wiring debt) | NO (mention only until SPEC completes) |
| Stop-chain / per-edit consolidation | SPEC-STOPCHAIN-TRIM-001 | completed | YES (behavioral; mode-aware hooks) | NO (covered by autonomy-tier page) |
| Agent body diet + parallel batching | SPEC-AGENT-PARALLEL-OPT-001 | completed | YES (indirect — faster agent runs) | NO (claude-code mention) |
| CC 2.1.219 upstream alignment | SPEC-CC2219-UPSTREAM-ALIGN-001 | completed | YES (subagent nesting, mode-token deprecation) | NO (claude-code rewrite) |
| Dynamic Workflows / ultracode | (accumulated) | completed | YES | NO (rewrite existing advanced/ultracode-workflows) |

**Catalog decision.** 9 features require NEW pages. 6 features are rewrites/mentions of existing pages. The total new-page count for M4 is approximately 9 new ko pages × 4 locales = 36 new pages, plus the 124 existing ko pages rewritten = 124 × 4 = 496 page rewrites. M4 is the dominant cost.

**Excluded.** `moai web` console redesign (internal, per mission brief). Harness-learning LSEL pipeline internals (the *user-facing* `harness learning` concept is documented; the LSEL PROPOSE→APPLY mechanics are not). SPECs not at `completed` by 2026-08-11 are excluded from the NEW-badge catalog (except SPEC-AGENT-MODEL-ENFORCE-001, listed as "partial").

---

## §D. Design-token DIFF (`colors_and_type.css` vs current `moai-brand.css`)

Direct token-by-token comparison. The PRIMARY/INK/BG trio is identical (no port work). The neutral ramp and several secondary tokens differ.

### §D.1 Identical (no port work)

| Token | Current (`moai-brand.css`) | Handoff (`colors_and_type.css`) | Match? |
|---|---|---|---|
| `--color-primary` | `#3d7d5f` | `#3d7d5f` | YES |
| `--color-primary-hover` | `#316750` | `#316750` | YES |
| `--color-primary-active` | `#265240` | `#265240` | YES |
| `--color-ink` | `#060606` | `#060606` | YES |
| `--color-bg` | `#f4f4f4` | `#f4f4f4` | YES |
| `--color-surface` | `#ffffff` | `#ffffff` | YES |

### §D.2 Differs (port surface)

| Token | Current | Handoff v2-renewal | Note |
|---|---|---|---|
| `--neutral-50` | `#f7f7f7` | `#f4f4f4` (= bg) | Handoff collapses 50 → bg |
| `--neutral-100` | `#f0f0f0` | `#e6e6e6` | Mascot-light gray |
| `--neutral-200` | `#e4e4e4` | `#d1d1d1` | |
| `--neutral-300` | `#d4d4d4` | `#b5b5b5` | |
| `--neutral-400` | `#a3a3a3` | `#9fa0a0` | Mascot-mid gray (the namesake) |
| `--neutral-500` | `#737373` | `#757575` | |
| `--neutral-600` | `#525252` | `#565656` | |
| `--neutral-700` | `#262626` | `#3a3a3a` | |
| `--neutral-800` | `#171717` | `#242424` | |
| `--neutral-900` | `#141414` | `#141414` | match |
| `--neutral-950` | `#060606` | `#060606` | match (= ink) |

**Interpretation.** The current ramp is a pure-achromatic Tailwind-aligned scale. The handoff ramp is mascot-derived — `#9fa0a0` carries a barely-perceptible green hue (the mascot's body gray), and the whole ramp is shifted to feel slightly warmer/grayer. Visually subtle; structurally a clean token-port.

### §D.3 Net-new tokens (current has no equivalent)

The handoff introduces tokens the current `moai-brand.css` does not carry:

- `--gradient-signature: #3d7d5f` and soft/dark variants (the handoff collapses the prior 135deg gradient to a solid color — the current site's `moai-brand.css` may still carry the warm-cream gradient remnants from v1; verify in M1).
- `--tracking-display-tight: -0.075em` (displaytighter tracking).
- `--text-display: clamp(2.25rem, 4.5vw, 4rem)` (fluid display size).
- `--shadow-signature: 0 8px 32px rgba(61,125,95,0.20)` (green-tinted hover glow).
- `--easing-bounce: cubic-bezier(0.34,1.56,0.64,1)` and `--easing-smooth: cubic-bezier(0.16,1,0.3,1)`.
- `--radius-pill: 32px` and `--radius-full: 9999px` (the current site may carry only `--radius-lg`).

### §D.4 Removed/forbidden tokens

- The v1 warm-cream tokens (`--color-bg: #faf9f5`, the 135deg gradient) are already noted as superseded in the current FROZEN header. The handoff does not reintroduce them.
- The `[data-theme="dark"]` block is in the handoff but is NOT ported (REQ-DS-003, light-only).

### §D.5 Port risk assessment

The port is LOW-to-MEDIUM risk. The PRIMARY/INK/BG trio is stable (the dominant visual load-bearers). The neutral-ramp shift is visually subtle and structurally clean. The net-new tokens are additive (new `--gradient-signature`, `--shadow-signature`, easings) — they don't replace existing references, so existing layouts don't break. The risk surface is hardcoded hex values in `moai-docs-layout.css` / `moai-docs-theme.css` / `moai-design.css` that should be token references but aren't; M1's audit step converts these.

---

## §E. Risks / unknowns (plan-phase markers — ALL RESOLVED at v0.2.0)

- `[RESOLVED: 12-vs-13 section reconciliation]` — The handoff home grid shows 12 cards; the live tree has 13 sections. **User decision (v0.2.0): 12 sections is canonical; `changelog` moves to the site footer (a footer link, not a sidebar section); `cost-optimization` retains its own card per the handoff home grid.** M0 implements the 12-section shape. This was D1-a of the plan-auditor's iteration-1 FAIL; resolved via orchestrator AskUserQuestion round.
- `[RESOLVED: book CTA target]` — The Docs Home prototype shows a "book.mo.ai.kr" CTA card. **The URL `https://book.mo.ai.kr` is verified live (HTTP 200) at v0.2.0 and is already referenced in the live `data/menu/main.yaml`. The book CTA card stays in the home hero.** This was D1-b of the plan-auditor's iteration-1 FAIL; resolved via orchestrator URL verification.
- `[RESOLVED: mascot asset licensing]` — The 6 mascot PNGs and 3 logo variants are **covered by the project Apache-2.0 license** (confirmed by `© 2026 modu-ai · Apache-2.0` in the handoff home screen footer). No separate attribution footer is required. This was D1-c of the plan-auditor's iteration-1 FAIL; resolved via orchestrator license verification.

**Zero `[NEEDS CLARIFICATION]` markers remain** — verifiable by `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-DESIGN-DOCS-V31-001/` returning only the iteration-1 audit.md and the (renamed) §A.4 references inside acceptance.md, neither of which is an open marker.
