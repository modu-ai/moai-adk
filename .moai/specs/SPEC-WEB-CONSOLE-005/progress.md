# Progress — SPEC-WEB-CONSOLE-005

## Plan Phase

- **Tier**: M (3 artifacts: spec.md + plan.md + acceptance.md, + progress.md)
- **cycle_type (run-phase)**: tdd (per `development_mode: tdd`)
- **status**: draft
- **Cohort**: web-console-v3 **final** member (005, cohort-internal label "S3" — the cohort terminator); siblings 001 (모태) / 002 S1 / 003 S2a / 004 visual-restyle — **all completed**
- **Primary source**: `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` (§5.3 i18n decision / §6 cohort decomposition / §7 font strategy + offline-CDN discovery / §9 M5) + the design `i18n.js` dictionary (derivative source)
- **REQ count**: 12 (REQ-WC5-001 .. REQ-WC5-012)
- **AC count**: 13 logical (15 rows incl. AC-008a/008b + AC-010a/010b split)
- **MUST-PASS invariants**:
  1. **interface language ≠ content language** (REQ-WC5-008 / AC-WC5-008a) — the appbar langpick mutates NO server-submitted field; POST round-trip byte-identical regardless of interface language
  2. **server contract / validate.go byte-unchanged** (REQ-WC5-010 / AC-WC5-010a)
  3. **offline / CDN-free** (REQ-WC5-006/011 / AC-WC5-011) — all font + dictionary assets via `go:embed`; zero external `https://` font/style/script URL

### Scope (two parts)

- **E.1 — Web interface i18n** (greenfield; handoff §9 M5): `data-i18n` attribute wiring on the page chrome + a new client-side `internal/web/assets/i18n.js` dictionary (en/ko/ja/zh, STATIC, no server round-trip) + an appbar `langpick` `<select>` + `localStorage("moai-console-lang")` persistence + load-time apply. `rv.*` design-review keys EXCLUDED.
- **E.2 — CJK self-host webfont coverage** (handoff §7): the 004-embedded Pretendard ko+Latin subset does NOT cover ja hiragana/katakana/kanji or zh hanzi; add a self-host woff2 subset (Option (c) — glyph-subset to exactly the **shipped** ja/zh dictionary string set, a low-hundreds glyph set; estimate ~290, tokenizer-dependent ~284–287 band — run-phase measures the actual shipped-dictionary count and records it here) for the ja/zh interface glyphs.

### Plan-phase artifacts created

- spec.md — 12 GEARS REQs (REQ-WC5-001..012) + 13 AC index + §4 Exclusions (E.1 server-side i18n→out / E.2 content-language fieldset→untouched / E.3 nested config→S2b out / E.4 rv.* + .review aside / E.5 anti-patterns). §1.5 carries the estimated glyph demand (low-hundreds, ~290 estimate — run-phase re-measures against the shipped dictionary) grounding the font decision. §1.4 carries the interface≠content-language invariant rationale. REQ-WC5-003 fixes the appbar picker id to `uiLangSelect` (non-colliding with the live `langSelect` content-language helper template).
- plan.md — Tier M justification + §A Context + §B Known Issues (R1 server-contract-leak-via-langpick HIGHEST + R2 offline CDN + R3 CJK bloat + R4 004-test-guard contradiction + R5 rv.* leak + R6 data-i18n/dictionary drift + R7 html-lang/font interaction) + §C Pre-flight + §D Constraints + §E Self-Verification + **§F Font Strategy Research & Proposal** (3-option trade-off table; recommends Option (c) glyph-subset-to-exact-dictionary) + §G Milestones (M1 CJK font → M2 dictionary + data-i18n → M3 appbar langpick + client apply [server-contract gate] → M4 test reconciliation + a11y + closure) + §H Anti-Patterns + §I Cross-References
- acceptance.md — full Given-When-Then for all 13 AC (15 rows) + §C traceability + §D edge cases (EC-1..EC-10) + §E Definition of Done + §F forward-looking cohort-closure checks
- progress.md — this file

### Tier determination

Tier **M** — multi-file cross-asset change (page.html.tmpl + new i18n.js + app.js + console.css + new CJK woff2 subset font + assets.go go:embed + restyle_test.go reconciliation), little-to-no production Go change (interface language is client-only), no constitutional change, < ~15 files. Exceeds Tier S (<5 files / <300 LOC) due to the new client-i18n surface + new offline font pipeline + server-contract regression surface + sibling-test reconciliation. Not Tier L (no new persistence model, no nested-config redesign).

### Font strategy decision (plan.md §F)

- **Recommended: Option (c)** — glyph-subset the CJK font to EXACTLY the **shipped** ja/zh interface dictionary string set (a low-hundreds glyph set) via `pyftsubset --text=` against `internal/web/assets/i18n.js`, sourced from Noto Sans SC (zh) + Noto Sans JP (ja) under OFL-1.1. Appended to `--font-sans` after Pretendard so en/ko stay on Pretendard (no regression).
- **Why (c) over (a)/(b)**: the dictionary is a FIXED known glyph set in the low hundreds → exact subset is tens of KB (vs Option (b) full Noto multi-MB bloat). Option (a) single-Pretendard rejected because Pretendard has no simplified-SC coverage for zh. Same `pyftsubset` toolchain 004 already used — no new build dependency. The exact glyph count is measured at run-phase against the shipped dictionary (the strategy does not depend on the precise number — auditor D2: the design-file count is provisional + tokenizer-dependent).

### Critical run-phase note (sibling test contradiction)

004's `internal/web/restyle_test.go::TestAppbarRendered` (lines 207-216) asserts the rendered-body literals `class="langpick"`, `id="langSelect"`, and `data-i18n` are ABSENT (the 004 S3-exclusion guard). 005 INTENTIONALLY lands `class="langpick"` + `data-i18n` + the NEW `id="uiLangSelect"` → the `class="langpick"`/`data-i18n` assertions WILL FAIL unless inverted. The run-phase MUST reconcile this test (move `class="langpick"` + `data-i18n` to the expected block AND add an EXPECT for `id="uiLangSelect"`) — it is the planned guard reversal (REQ-WC5-012 / AC-WC5-012), NOT a regression. CRITICAL: do NOT invert the stale `id="langSelect"` forbidden-string into an EXPECT — it referenced the never-landed original id; the appbar picker uses `uiLangSelect` because `langSelect` is the live `{{define "langSelect"}}` content-language helper. This is an AC that inverts an existing passing test — the run-phase TDD RED must account for it explicitly.

plan_complete_at: 2026-06-03
plan_status: audit-ready

---

## §E — Phase 0.95 Mode Selection (orchestrator-logged per REQ-ATR-008)

**Decision: sub-agent (Mode 5)** — coding-heavy single-package `internal/web`; sequential manager-develop (cycle_type=tdd) per milestone M1-M4.

### Input parameters
- **tier**: M
- **scope (file count)**: ~7-8 files (page.html.tmpl, i18n.js NEW, app.js, console.css, fonts/ CJK subset NEW, assets.go, restyle_test.go) — all within `internal/web/`
- **domain count**: 1 (single package `internal/web`; template + client assets + Go test, not cross-domain)
- **file language mix**: HTML template + JS + CSS + woff2 binary + minimal Go (test + go:embed) — coding-heavy, NOT research-heavy
- **concurrency benefit**: LOW (tightly-coupled single-package edits; Finding A4 caveat)
- **Agent Teams prereqs (REQ-ATR-013)**: harness thorough? NO (default) · workflow.team.enabled? NO · CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1? NO → gate fails

### Mode evaluation
| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | multi-file semantic change (new i18n surface + font pipeline) |
| 2 background | no | run-phase writes files (CONST-V3R2-020 forbids background Write/Edit) |
| 3 agent-team | no | REQ-ATR-013 capability gate fails (all 3 unset) + single domain |
| 4 parallel | no | coding-heavy + single domain (Finding A4 — parallel multi-spawn is research-heavy multi-domain) |
| 5 sub-agent | **YES** | coding-heavy default; sequential manager-develop (tdd) per milestone M1-M4 |
| 6 workflow | no | not ≥30-file mechanical-uniform transform; semantic new-code work |

### Justification
Coding-heavy, single-package (`internal/web`) cross-asset implementation with tightly-coupled milestones (font → dictionary → langpick server-contract gate → test reconciliation). Per Anthropic Finding A4 ("most coding tasks involve fewer truly parallelizable tasks than research"), the sequential sub-agent path is the correct default. Agent Teams (Mode 3) capability gate fails on all three REQ-ATR-013 conditions. A single `manager-develop` spawn (cycle_type=tdd) handles M1-M4 with RED-GREEN-REFACTOR; the M3 server-contract gate (langpick must not leak into POST) is the critical milestone. Matches cohort precedent (004 = Mode 5 sub-agent manager-develop tdd).

### Gate provenance
- **GATE-2**: PASSED (user approved run-phase entry via AskUserQuestion; score-independent per REQ-ATR-015)
- **Phase 0.5 (Plan Audit Gate)**: satisfied in-session — plan-auditor iter-1 PASS-WITH-DEBT 0.84 (Tier M threshold 0.80) + D1/D2/D3 resolved + orchestrator independent verification 7/7; no re-execution (audit fresh + debt cleared)

mode_selection: sub-agent

---

## Run-phase Evidence (manager-develop, cycle_type=tdd, Mode 5)

- **run_commit_shas** (cherry-picked onto docs/glm-webtool-routing-m1-m5; L1 worktree auto-materialized base 398f882f5, conflict-free cherry-pick per L_l1_worktree_cherrypick; worktree-origin a19ad50ab..8d9cfdf1b):
  - M1 `a569058fa` — CJK woff2 subset font layer (offline-safe foundation)
  - M2 `66bb5c167` — i18n dictionary + data-i18n wiring (chrome-translation layer)
  - M3 `aca719270` — appbar langpick + client apply/persist (server-contract gate)
  - M4 `5657e403d` — 004 TestAppbarRendered guard inversion + a11y + closure
  - reconcile (this) — spec.md draft→in-progress + run evidence + mapsloop cleanup
- **AC matrix**: 13/13 AC PASS (acceptance.md SSOT). 3 must-pass verified: AC-WC5-008a (interface≠content — POST byte-identical), AC-WC5-010a (validate.go byte-unchanged), AC-WC5-011 (offline / CDN-free).
- **orchestrator independent verification (Trust-but-verify) 7/7 PASS**: cross-platform build host+windows exit 0; closure gate go test ./internal/web/... ./internal/cli/... ./internal/config/... ok; offline grep (served assets 0 external URL); rv.=0; `git diff --exit-code internal/web/validate.go`=0; langpick (`uiLangSelect`, no `name=`, L59) outside `<form>` (L106); data-i18n=36 (≥25 floor); locales=4; Noto CJK 6 subset ~235KB + OFL.
- **E7 — measured shipped-dictionary CJK glyph count: 279** (197 CJK ideographs + 44 katakana + 29 hiragana + 9 CJK punctuation; SC+JP subsets each cmap=279 — 100% coverage, 0 over-coverage). Supersedes the §1.5 provisional ~284-287 estimate (auditor D2: design-file count was tokenizer-dependent).
- **Font deliverable** (plan §F Option c): Noto Sans CJK SC (zh) + JP (ja), 3 weights each (Regular/Medium/Bold), 279-glyph pyftsubset (same toolchain as 004), ~235KB total, OFL-1.1 (OFL-NotoSansCJK.txt). Pretendard ko/Latin (004) preserved; CJK activated via `html[lang="ja|zh"]` font-stack override (en/ko stay Pretendard).

run_status: implementation-complete (sync pending)
