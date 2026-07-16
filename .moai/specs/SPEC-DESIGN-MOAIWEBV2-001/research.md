# research.md — SPEC-DESIGN-MOAIWEBV2-001

> Codebase investigation findings gathered during plan-phase. All findings are tool-observed (grep / file-listing / rendered-source inspection), per the verification-claim-integrity doctrine — a defect claim (e.g. "orphan panel") requires tool evidence, cited below.

## §A. Current Console Token Layer

- **File**: `internal/web/assets/console.css` (~33 KB, 223 `--var` definitions).
- **Origin**: authored under `SPEC-WEB-CONSOLE-004` (visual restyle) + `SPEC-WEB-CONSOLE-005` (CJK subset), from an EARLIER "from-claude-design" export — a DIFFERENT (older) token generation than the current v2 canon.
- **Offline-safe invariant (must preserve)**: the header comment documents `REQ-WC4-001/002` — Google Fonts `@import` REMOVED; Pretendard loaded as self-host woff2 subset (relative `/static/fonts/*` paths); Inter/JetBrains Mono → system fallback; `SPEC-WEB-CONSOLE-005` adds Noto Sans CJK SC/JP woff2 subsets for ja/zh. Zero remote font/style fetch.

### §A.1 Current `:root` color/token head (observed)

```
--color-primary: #3d7d5f; --color-primary-hover: #0e3835; --color-primary-active: #0a2825;
--color-ink: #09110f;   (teal-tinted)     --color-bg: #f3f3f3;
--neutral-50: #f3f3f3 / -100 #eaeaea / -200 #d4d4d4 / -300 #bcbcbc / -400 #959595 /
  -500 #6e6e6e / -600 #4c4c4c / -700 #2e2e2e / -800 #1a1f1d / -900 #0e1513 / -950 #09110f (teal)
--color-success #1c7c70   --fg-2 #4c4c4c   --fg-3 #6e6e6e
--border-1 #d4d4d4 / -2 #eaeaea / -strong #bcbcbc   --border-focus-ring rgba(61,125,95,0.12)
--gradient-signature: linear-gradient(135deg, #3d7d5f 0%, #09110f 100%)   (a GRADIENT)
--shadow-* base rgba(9,17,15,…)  (teal-tinted)
```

## §B. v2 Design Canon (bundle)

- **File**: `.moai/state/ai-design-system/project/colors_and_type.css` (v2 — "achromatic neutrals extracted from mascot + point-green sole accent").
- **Canon values**: `--color-ink #060606`, `--color-bg #f4f4f4`, achromatic hue-0 ramp (`#f4f4f4`/`#e6e6e6`/`#d1d1d1`/`#b5b5b5`/`#9fa0a0`/`#757575`/`#565656`/`#3a3a3a`/`#242424`/`#141414`/`#060606`), `--color-success #2e8a63`, `--fg-2 #565656`, `--fg-3 #757575`, borders `#d1d1d1`/`#e6e6e6`/`#b5b5b5`, focus-ring `rgba(61,125,95,0.16)`, shadow base `rgba(6,6,6,…)`, `--gradient-signature: #3d7d5f` (a SOLID color, NOT a gradient), primary-hover/active `#316750`/`#265240`.
- **Also present**: full type/tracking/size/line-height/spacing(4px)/radius/motion/container scales + a `[data-theme="dark"]` block.
- **Font caveat**: the v2 canon has a Google Fonts `@import` + OTF `@font-face` — these MUST NOT be ported (they would break the console offline-safe invariant). Only the `:root` VALUES are ported.

## §B.1 Current → v2 Token Delta Summary

| Token | Current console | v2 canon | Action |
|-------|-----------------|----------|--------|
| `--color-primary-hover` | `#0e3835` | `#316750` | update |
| `--color-primary-active` | `#0a2825` | `#265240` | update |
| `--color-ink` | `#09110f` (teal) | `#060606` | de-tint |
| `--color-bg` / `--neutral-50` | `#f3f3f3` | `#f4f4f4` | update |
| `--neutral-100` | `#eaeaea` | `#e6e6e6` | update |
| `--neutral-200` | `#d4d4d4` | `#d1d1d1` | update |
| `--neutral-300` | `#bcbcbc` | `#b5b5b5` | update |
| `--neutral-400` | `#959595` | `#9fa0a0` | update |
| `--neutral-500` | `#6e6e6e` | `#757575` | update |
| `--neutral-600` | `#4c4c4c` | `#565656` | update |
| `--neutral-700` | `#2e2e2e` | `#3a3a3a` | update |
| `--neutral-800` | `#1a1f1d` (teal) | `#242424` | de-tint |
| `--neutral-900` | `#0e1513` (teal) | `#141414` | de-tint |
| `--neutral-950` | `#09110f` (teal) | `#060606` | de-tint |
| `--color-success` | `#1c7c70` | `#2e8a63` | update |
| `--fg-2` | `#4c4c4c` | `#565656` | update |
| `--fg-3` | `#6e6e6e` | `#757575` | update |
| `--border-1/2/strong` | `#d4d4d4`/`#eaeaea`/`#bcbcbc` | `#d1d1d1`/`#e6e6e6`/`#b5b5b5` | update |
| `--border-focus-ring` | `rgba(…,0.12)` | `rgba(…,0.16)` | update |
| `--gradient-signature` | `linear-gradient(…#09110f)` | `#3d7d5f` (solid) | adopt v2 solid (confirmed via AskUserQuestion 2026-07-16) |
| `--shadow-*` base | `rgba(9,17,15,…)` | `rgba(6,6,6,…)` | de-tint |

Token NAMES are identical across both files (only values differ) → value-only substitution keeps every `var(--token)` consumer resolving. Type/tracking/size/spacing/radius/motion scales already match (both derive from the same design family); the primary delta is the color/neutral/shadow de-tint + signature solidification.

## §C. Orphan `project` Panel — Reachability Evidence

### §C.1 Tool evidence (commands run + observed output)

```
$ grep -rn 'data-panel="project"' internal/web/*.templ
internal/web/root.templ:129:  <div class="tabpanel" data-panel="project">   → @fieldsetProject(view)

$ (consoleTabs() body — internal/web/schemaform.go:33)
returns 6 tabs: identity, language, launch, llm, agentfm, report
   → NO "project" entry

$ grep -rn 'data-tab.*project|"project"' internal/web/assets/app.js
(no output — no data-tab="project" anywhere)
```

### §C.2 Interpretation

- The console renders 7 panels: `identity`, `language`, `launch`, **`project`**, `llm`, `report`, `agentfm`.
- `consoleTabs()` renders 6 tab buttons: `identity`, `language`, `launch`, `llm`, `agentfm`, `report`.
- Exactly ONE panel — `project` — has NO corresponding tab button. The tab-switch JS (`app.js` `wireTabs`, line 271-293) activates a panel only when a `data-tab` button with the matching id is clicked. With no `data-tab="project"` button, the `project` panel can never receive `is-active` → it is permanently `display:none` and **unreachable via the nav** = ORPHAN (verified, not inferred).
- **Content is real + tested**: `fieldsetProject` (`fieldsets.templ:159`) renders `development_mode`, `git_convention` + nested auto-detection (confidence/sample-size/enabled/enforce-on-push), and `quality.*` (coverage target, enforce-quality, min-coverage-per-commit). It is asserted by `projectnested_parse_test.go:151/177` (render) and used as a boundary marker by `schemaform_test.go:78` (`TestConsoleRendersReportTab` locates the `launch`→`project` span).
- **History**: heavy tab churn — `cca120c70` (WEBCONF-SIMPLIFY removed 11 tabs), `ec576d087` (model-policy merged into agentfm), `4145e5d8a` (report → own tab), `d900d992b` (M5-b introduced the tab nav). The `project` tab was most plausibly dropped during this consolidation, leaving the panel as dead DOM.

### §C.3 Consequence of removal (Option A)

Removing `fieldsetProject` drops the ONLY web-console editing surface for `development_mode` / `git_convention` / `quality.*` (grep confirms these fields render nowhere else — `fieldsets.templ` lines 170-178 are their sole occurrence). They remain editable via config files / CLI. The user accepted this functional loss and confirmed REMOVE via AskUserQuestion 2026-07-16 (plan.md §B).

## §D. Mascot Inventory

### §D.1 Current (observed)

```
$ ls internal/web/assets/mascots/
mascot-coding.png    (9309 bytes)  — USED (header brand badge, board.templ:106 + root.templ:162)
mascot-talking.png   (5120 bytes)  — UNUSED (grep -rn 'mascot-talking' internal/web/ → 0 matches)
```

Current count = **2 files**, only **1 wired** (`mascot-coding.png` in the header badge). `mascot-talking.png` is dead weight.

### §D.2 v2 6-pose target (bundle `assets/characters/`)

```
MoAI-Mascot-Coffee.png       → mascot-coffee.png
MoAI-Mascot-Explaining.png   → mascot-explaining.png
MoAI-Mascot-Pointing.png     → mascot-pointing.png
MoAI-Mascot-Searching.png    → mascot-searching.png
MoAI-Mascot-Teaching.png     → mascot-teaching.png
MoAI-Mascot-Thinking.png     → mascot-thinking.png
```

Naming: bundle uses `MoAI-Mascot-<Pose>.png` (capitalized); current console uses lowercase-kebab (`mascot-coding.png`). Decision = adopt lowercase-kebab to match the existing console convention (design.md §F). The v2 set has NO "coding" pose → the header badge remaps to `mascot-thinking.png` (confirmed via AskUserQuestion 2026-07-16).

### §D.3 Serving mechanism

`internal/web/assets.go` declares `//go:embed assets/console.css assets/app.js assets/i18n.js assets/htmx.min.js assets/fonts assets/mascots`. The `assets/mascots` glob captures ANY file in the dir → adding the 6 poses requires NO `assets.go` edit. `app.go:127` serves `/static/` from `staticFS()` (the `assets/` subtree) → poses reachable at `/static/mascots/mascot-<pose>.png`.

## §E. SPEC-1 Precedent (reuse)

`SPEC-DESIGN-DOCSV2-001` (closed 4981f160c) established for the docs-site:
- **§A token unification**: single v2-native token file, layering collapsed, grep-based token-parity ACs. → This SPEC mirrors the value-only realignment approach (§B.1 above).
- **§F mascot placement architecture**: pose→surface map + lowercase-kebab shortcode param vs Capitalized filename mapping + mascot-only bounce easing + `prefers-reduced-motion` respect. → This SPEC reuses the naming decision and adapts the placement map to the console's surfaces (design.md §F).
- **§I casing risk**: `MoAI-Mascot-Thinking.png` (capital) vs `thinking` (lowercase param) — the same casing bridge applies here (bundle Capitalized → console lowercase-kebab file).

## §F. Constraints Confirmed

- `internal/web/` = Go source (NOT template distribution) → Template-First N/A, but `make build` required for `templ generate` + embed (spec.md §A.1).
- Offline-safe invariant (`REQ-WC4-001/002`, `REQ-WC5-006/011`) → do NOT port the v2 canon `@import`/OTF `@font-face`.
- Server-contract preservation → visual + asset + dead-UI only; handler/route/seam untouched.

## §G. Resolved Decisions (mirror of plan.md §B — all confirmed via AskUserQuestion 2026-07-16)

- Project-panel disposition — **REMOVE**. Delete the orphan panel + related code; dev-mode/git-convention/quality settings remain editable via yaml/CLI (user accepted the editability loss).
- Signature accent — **v2 solid `#3d7d5f`** (drop the linear-gradient).
- Header brand pose — **`mascot-thinking.png`** replaces `mascot-coding.png`.
- Mascot placement — **library + header only**: embed all 6 poses; live UI swap only at the header brand badge.
