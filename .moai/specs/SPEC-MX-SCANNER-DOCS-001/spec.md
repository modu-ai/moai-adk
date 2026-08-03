---
id: SPEC-MX-SCANNER-DOCS-001
title: "Document MX scanner advanced features (rotRisk, LSP fan-in, CGO complexity, scan automation)"
version: 0.1.0
status: completed
created: 2026-08-04
updated: 2026-08-04
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: docs
lifecycle: spec-anchored
tags: "mx, scanner, docs, i18n, rot-risk, fan-in, cgo, complexity, automation"
tier: M
---

# SPEC-MX-SCANNER-DOCS-001 — Document MX scanner advanced features

## HISTORY

- 2026-08-04 — Created (manager-spec, plan-phase). P2-3 item from the MX activation queue (`mx-system-analysis-activation-20260804.html`). Documentation-only: the scanner already implements all 4 features; this SPEC captures HOW each works so docs-site (4 locales) and the README FAQ can describe them accurately. Prior queue items landed: SPEC-MX-ACTIVATION-001 (P0-1), the P2-1 SPEC, SPEC-MX-ASSOCIATION-001 (P1-3, PR #1321).

## §A. User Story

As a MoAI user reading `adk.mo.ai.kr` or the project README, I want the four non-obvious scanner behaviors — rotRisk scoring on `@MX:DEBT` tags, the LSP-backed fan-in condition for `@MX:ANCHOR`, the CGO-gated complexity measurement path, and the automatic scan timing (when the sidecar index is rebuilt without my running `moai mx scan`) — to be documented in my locale with accurate diagrams, so that I can interpret query output (`rotRisk: "no-trigger"`, `fan_in_method: "lsp"` vs `"textual"`, `Supported: false` on complexity) and reason about when my index is fresh without reading Go source.

## §B. Scope

In scope:

- Documentation of four EXISTING scanner behaviors (no code or behavior change to the scanner — this SPEC is documentation-only):
  1. `rotRisk` scoring on `@MX:DEBT` tags (value `"no-trigger"` when `@MX:UPGRADE` is absent; empty when present).
  2. The LSP fan-in condition for `@MX:ANCHOR` — LSP-primary with textual fallback, plus the strict-mode failure shape.
  3. The CGO-gated complexity measurement path — why CGO and non-CGO builds differ, and what `Supported: false` means in each case.
  4. Scan automation timing — which lifecycle hooks (SessionStart cold-start, PostToolUse, SessionEnd, sync) trigger or consult the scan, and the fail-open ceilings.
- docs-site coverage in 4 locales (ko/en/ja/zh) on a new dedicated page, with cross-reference links from the existing `advanced/mx-tags.md` and `utility-commands/moai-mx.md`.
- README FAQ entries in the 4 README files (en primary; ko/ja/zh derived).

Out of scope — see §G.

## §C. Functional Requirements (GEARS)

### REQ-MSD-001 — rotRisk documentation (Ubiquitous)
The docs-site scanner-internals page and the README FAQ shall describe `rotRisk` as a property of `@MX:DEBT` tags whose value is `"no-trigger"` when the tag lacks an `@MX:UPGRADE` sub-line and empty (omitted) when an `@MX:UPGRADE` sub-line is present, with the explicit note that an absent `@MX:CEILING` is a quality note rather than the rot gate.

### REQ-MSD-002 — LSP fan-in condition documentation (Ubiquitous)
The docs-site scanner-internals page shall document that `@MX:ANCHOR` fan-in is computed LSP-primary (`textDocument/references`) with a textual grep fallback, that the `fan_in_method` field discloses which path produced the count (`"lsp"` or `"textual"`), and that under `MOAI_MX_QUERY_STRICT=1` an unavailable LSP server returns `LSPRequiredError` instead of falling back.

### REQ-MSD-003 — CGO complexity path documentation (Ubiquitous)
The docs-site scanner-internals page shall document that cyclomatic complexity measurement is a CGO-gated tree-sitter implementation, that a non-CGO build emits `Result{Supported: false}` for every language as a hard stub, and that even on a CGO build the measurement returns `Supported: false` for scaffolded (unseeded) languages, files larger than 1 MiB, and any parse or query error.

### REQ-MSD-004 — Scan automation timing documentation (Ubiquitous)
The docs-site scanner-internals page and the README FAQ shall document the lifecycle points that trigger or consult an MX scan: the explicit `moai mx scan` CLI, the SessionStart deferred cold-start scan (time-boxed, fail-open), the PostToolUse MX validation hook, the SessionEnd batch validation, and the sync-workflow enforcement gate, with the fail-open semantics made explicit.

### REQ-MSD-005 — Mermaid diagrams TD-only (Ubiquitous)
**Where** the docs-site scanner-internals page includes a sequence or flow diagram, the docs shall render it as a `flowchart TD` or `graph TB` Mermaid block, and the diagram source shall be identical across all 4 locales.

### REQ-MSD-006 — Icon shortcodes over emoji (Ubiquitous)
The docs-site scanner-internals page body text shall use the `{{</* icon <name> [variant] */>}}` shortcode for any decorative callout marker and shall NOT introduce body-text emoji; typographic arrows (`→ ← ↓ ✓ ✗`) and branding emoji inside MoAI-orchestrator example code blocks are permitted per the docs-site i18n rules.

### REQ-MSD-007 — 4-locale same-PR parity (Ubiquitous)
The docs-site scanner-internals page and any cross-reference link added to `mx-tags.md` / `moai-mx.md` shall land in all 4 locales (ko/en/ja/zh) in the same PR, with section structure and diagram source preserved verbatim across locales (locale-parity threshold 1.0).

### REQ-MSD-008 — README FAQ 4-file parity (Ubiquitous)
The README FAQ shall gain entries covering rotRisk, LSP fan-in fallback, the CGO complexity stub, and scan automation timing, authored in `README.md` (English canonical) and derived into `README.ko.md`, `README.ja.md`, and `README.zh.md` in the same PR with heading and section-order parity.

### REQ-MSD-009 — No scanner behavior change (Unwanted)
**When** the documentation is authored, the scanner source under `internal/mx/`, `internal/hook/mx/`, and `internal/cli/mx_scan.go` shall not be modified by this SPEC; this SPEC produces documentation artifacts only.

### REQ-MSD-010 — URL whitelist (Ubiquitous)
The docs-site scanner-internals page and the README FAQ entries shall link only to URLs on the `adk.mo.ai.kr` docs-site domain; no blacklisted or alternative docs-site domain (e.g. `docs.moai-ai.dev`, `adk.moai.com`) shall appear in any of the 4 locales' page or any of the 4 README files.

## §D. Acceptance Criteria Summary

See `acceptance.md` for the full Given-When-Then matrix. Headline criteria: 4 locales of the new docs-site page exist with section parity; README 4-file FAQ parity holds; Mermaid diagrams are TD-only; no body-text emoji outside the permitted classes; the four behaviors are described accurately against the source (rotRisk values, `fan_in_method` values, `Supported: false` triggers, scan lifecycle points); zero source files under `internal/mx/` or `internal/hook/mx/` are touched.

## §E. Non-Functional Constraints

- 4-locale i18n parity (ko canonical for docs-site; en canonical for README) — `.moai/docs/docs-site-i18n-rules.md` + CLAUDE.local.md §17.1.
- Mermaid diagrams TD-only; no LR/RL directions.
- Body content uses `{{</* icon */>}}` shortcodes; no decorative body emoji.
- Light single-theme only; no dark-theme branches.
- URLs restricted to the `adk.mo.ai.kr` whitelist; no blacklisted domains.
- README 4-file heading and section-order parity; English is canonical.

## §F. Risks

- **Accuracy drift**: the four behaviors are subtle (strict-mode failure shape, scaffolded-language stub). A mistranslation could assert a behavior the scanner does not have. Mitigation: plan.md §C research findings are the single factual basis; locale derivations translate verbatim against the en/ko canonical.
- **Page-proliferation**: adding a 5th MX-related docs page could fragment discoverability. Mitigation: cross-reference links from `mx-tags.md` and `moai-mx.md` are in scope (REQ-MSD-007) so the new page is reachable from the two existing entry points.
- **README bloat**: four new FAQ entries could push the FAQ past a comfortable length. Mitigation: each entry stays to 2–4 sentences with a cross-link to the docs-site page.

## §G. Exclusions (Out of Scope)

### Out of Scope — Scanner source changes
- This SPEC MUST NOT modify `internal/mx/scanner.go`, `tag.go`, `resolver_query.go`, `fanin.go`, `fanin_lsp.go`, `danger_category.go`, `internal/hook/mx/{config,validator,types}.go`, `internal/hook/mx/complexity/*.go`, or `internal/cli/mx_scan.go`. The scanner is the subject of the documentation, not the target of a change.

### Out of Scope — New tag types or sub-lines
- Introducing new `@MX:*` tag types or sub-lines, or changing the semantics of existing ones, is out of scope. The docs describe the current tag taxonomy as defined in `mx-tag-protocol.md`.

### Out of Scope — docs-site design system changes
- Touching `moai-brand.css`, `moai-design.css`, the icon shortcode registry, or the Hugo render hooks is out of scope. The page uses the existing design system as-is.

### Out of Scope — Localization tooling
- Building or modifying `docs-i18n-check.sh`, `gen_menu.py`, or any automated locale-parity linter is out of scope. The 4-locale parity obligation (REQ-MSD-007/008) is satisfied by manual verification per the `hns-oss-docs-verify` recipe.

### Out of Scope — Menu / sidebar restructure
- Adding the new page to `data/menu/main.yaml` and the per-locale `_meta.yaml` IS in scope (necessary for discoverability); broader sidebar restructure or icon additions are out of scope.

## §H. Cross-References

- Prior MX SPECs: `SPEC-MX-001` (TAG system), `SPEC-MX-002` (auto-validation), `SPEC-MX-ASSOCIATION-001` (@MX:SPEC sub-line → spec_associations, PR #1321).
- docs-site i18n rules: `.moai/docs/docs-site-i18n-rules.md` + CLAUDE.local.md §17.1.
- MX tag protocol: `.claude/rules/moai/workflow/mx-tag-protocol.md`.
- Research basis: `plan.md` §C (per-feature findings with file/line citations).
- Acceptance matrix: `acceptance.md`.
