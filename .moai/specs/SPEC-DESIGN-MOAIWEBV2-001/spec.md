---
id: SPEC-DESIGN-MOAIWEBV2-001
title: "moai web console v2 design-system alignment"
version: "0.1.0"
status: completed
created: 2026-07-16
updated: 2026-07-17
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "design, web-console, tokens, mascot, templ"
tier: M
related_specs: [SPEC-DESIGN-DOCSV2-001]
---

# SPEC-DESIGN-MOAIWEBV2-001 — moai web console v2 design-system alignment

> Epic DESIGN-V2, SPEC 2 of 2. SPEC-1 (`SPEC-DESIGN-DOCSV2-001`, closed at commit 4981f160c) migrated the docs-site (adk.mo.ai.kr) to the v2 design system. This SPEC applies the SAME v2 design canon to the `moai web` console (the `internal/web/` Templ-based settings/board UI served on loopback), expands its mascot inventory to the v2 6-pose set, and cleans up the unreachable ("orphan") `project` config panel.

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-16 | manager-spec | Initial plan-phase authoring (draft). Tier M. GEARS requirements for v2 token alignment, mascot 6-pose expansion, orphan project-panel cleanup. |

---

## §A. Context & Intent

The `moai web` console (`internal/web/`) is a self-contained, offline-safe (loopback-only, zero-network) settings + board UI rendered by compiled-in Templ components. Its design token layer (`internal/web/assets/console.css` `:root` block, 223 `--var` definitions) was authored under `SPEC-WEB-CONSOLE-004/005` from an EARLIER "from-claude-design" export. That earlier token generation diverges from the current v2 canon (`.moai/state/ai-design-system/project/colors_and_type.css`) on ink hue (teal-tinted `#09110f` vs achromatic `#060606`), the neutral ramp, semantic colors, borders, focus-ring alpha, shadow base, and the signature accent (a linear-gradient vs the v2 solid point-green `#3d7d5f`).

This SPEC re-aligns the console token layer to the v2 canon WHILE preserving the offline-safe font layer (self-hosted Pretendard + Noto Sans CJK woff2 subsets; zero remote font/style fetch — the `REQ-WC4-001/002` invariant), expands the mascot library from the current 2 files (only 1 wired) to the v2 6-pose set, and removes the unreachable `project` config panel.

### §A.1 Template-First rule does NOT apply (scope note)

[HARD] The `internal/web/` tree is **Go application source**, NOT template distribution content under `internal/template/templates/`. The CLAUDE.local.md §2 Template-First rule (add to `internal/template/templates/` first, then `make build`) does **not** apply to `internal/web/` edits. However, `make build` IS still required after every edit because `internal/web/*.templ` sources compile to `internal/web/*_templ.go` via `templ generate` and the `console.css` / mascot assets are compiled into the binary via `//go:embed all:templates` — no, via `//go:embed assets/...` in `internal/web/assets.go`. Every milestone's verification therefore includes `make build` (templ generate + embed + Go compile). This SPEC touches `internal/web/` ONLY; it does NOT touch `docs-site/`, `internal/template/templates/`, or any `.claude/settings*.json`.

### §A.2 Design SSOT

The v2 design canon is `.moai/state/ai-design-system/project/` (a read-only Claude Design export). The authoritative token file is `colors_and_type.css`; the canonical mascot poses live in `assets/characters/` (`MoAI-Mascot-{Coffee,Explaining,Pointing,Searching,Teaching,Thinking}.png`). This SPEC CONSUMES the bundle; it does not edit bundle files.

---

## §B. Requirements (GEARS)

### §B.1 v2 Token Alignment

- **REQ-MWV2-001** (Ubiquitous): The console token layer (`console.css` `:root`) shall define color, neutral-ramp, semantic, foreground, border, radius, shadow, spacing, type, tracking, line-height, motion, and container tokens whose VALUES match the v2 canon (`colors_and_type.css`) for every token present in both files.
- **REQ-MWV2-002** (Ubiquitous): The neutral ramp (`--neutral-50` … `--neutral-950`) shall be pure-achromatic (hue-0) per the v2 canon — the console shall not retain the earlier teal-tinted ink/neutral values (`#09110f`, `#1a1f1d`, `#0e1513`).
- **REQ-MWV2-003** (Ubiquitous): The console `--color-ink` shall be `#060606` and `--color-bg` shall be `#f4f4f4`, matching the v2 achromatic canon.
- **REQ-MWV2-004** (Capability gate): Where a v2 token has no equivalent in the current console set, the console shall adopt the v2 token verbatim; where a console token has no v2 equivalent, it shall be preserved.
- **REQ-MWV2-005** (State-driven): While a component consumes the signature accent token, the console shall render it per the design.md §C signature-accent decision (v2 solid point-green vs retained gradient) — a single, consistent treatment across all accent surfaces.

### §B.2 Mascot 6-Pose Expansion

- **REQ-MWV2-010** (Ubiquitous): The console mascot library (`internal/web/assets/mascots/`) shall contain the v2 canonical 6-pose set (Coffee, Explaining, Pointing, Searching, Teaching, Thinking) sourced from the design bundle `assets/characters/`.
- **REQ-MWV2-011** (Ubiquitous): Each of the 6 poses shall be embedded into the binary via the existing `internal/web/assets.go` `//go:embed assets/mascots` directive and served under `/static/mascots/`.
- **REQ-MWV2-012** (Event-driven): When the console header brand badge renders, the console shall display a v2 mascot pose per the design.md §F placement map (replacing the current `mascot-coding.png` reference).
- **REQ-MWV2-013** (Unwanted behavior): The console shall not retain the unused `mascot-talking.png` asset (0 references confirmed) after the 6-pose set is adopted.
- **REQ-MWV2-014** (Ubiquitous): The mascot filenames shall follow the existing lowercase-kebab convention (`mascot-<pose>.png`) per the design.md §F naming decision.

### §B.3 Orphan `project` Panel Cleanup

- **REQ-MWV2-020** (Ubiquitous): The console shall not render an unreachable tab panel — every `data-panel="<id>"` in the rendered console HTML shall have a corresponding navigable `data-tab="<id>"` button (or shall be removed).
- **REQ-MWV2-021** (Event-driven): When the console page renders, the `project` config panel shall be resolved per the design.md §E disposition decision (remove the orphan, or restore reachability via a `project` tab) so that no orphan panel remains.
- **REQ-MWV2-022** (State-driven): While the orphan-panel disposition is "remove", the console shall preserve the atomic-Save submission contract for all remaining panels — no in-DOM field that another panel depends on shall be dropped without its owning fields being relocated or explicitly retired.

### §B.4 Offline-Safe & Zero-Contract Preservation (unwanted behavior)

- **REQ-MWV2-030** (Unwanted behavior): The console shall not introduce any remote font or stylesheet fetch (no `@import url("https://…")`, no CDN `@font-face src`) — the self-hosted Pretendard + Noto Sans CJK woff2 subset layer (`REQ-WC4-001/002`, `REQ-WC5-006/011`) shall be preserved verbatim.
- **REQ-MWV2-031** (Unwanted behavior): The console shall not change any server-side handler contract, route, form field name, or persistence seam — this SPEC is a visual + asset + dead-UI change only (zero server-contract change, mirroring the `SPEC-WEB-CONSOLE-004` restyle precedent).
- **REQ-MWV2-032** (Ubiquitous): The console shall preserve the light-only theme posture; the imported v2 `[data-theme="dark"]` token block, if adopted, shall remain inert dead code (no dark toggle wired).

### §B.5 Build & Verification Gate

- **REQ-MWV2-040** (Event-driven): When any `internal/web/*.templ`, `console.css`, or mascot asset is edited, the milestone shall regenerate `*_templ.go` and re-embed assets via `make build` before the milestone is considered complete.
- **REQ-MWV2-041** (Event-driven): When `go build ./...` or `go test ./internal/web/...` returns a non-zero exit after an edit, the milestone shall halt and the defect shall be resolved before proceeding.

---

## §C. Success Criteria

1. `console.css` `:root` token values match the v2 canon for every shared token (grep-verifiable per acceptance.md §D.1).
2. 6 mascot pose files exist under `internal/web/assets/mascots/` and are embedded + served (file-existence + embed-directive grep).
3. No orphan panel remains — `data-panel` set ⊆ navigable `data-tab` set in rendered HTML (rendered-HTML boundary test).
4. Zero remote font/stylesheet fetch reintroduced (absence grep for `@import url("http` / CDN `src`).
5. `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0, `go test ./internal/web/...` exit 0.
6. `make build` regenerates `*_templ.go` cleanly.

---

## §D. Constraints

- Scope: `internal/web/` ONLY. Do NOT touch `docs-site/`, `internal/template/templates/`, `.claude/settings*.json`, `.moai/state/ai-design-system/` (read-only bundle).
- Preserve the offline-safe font layer and all server-side contracts.
- Follow the existing file/naming conventions of the files being modified (lowercase-kebab mascot names, Templ component style).
- `make build` after every asset/templ/css edit; never hand-edit `*_templ.go` (regenerated).

---

## §E. Out of Scope

### Out of Scope — docs-site (SPEC-1)

- The docs-site v2 migration is `SPEC-DESIGN-DOCSV2-001` (Epic DESIGN-V2 SPEC-1, already closed). This SPEC touches `internal/web/` only; `docs-site/` is untouched.

### Out of Scope — v2 bundle authoring / edits

- The design bundle at `.moai/state/ai-design-system/project/` is a read-only handoff. This SPEC consumes `colors_and_type.css` and the mascot poses; it does not edit `colors_and_type.css`, `README.md`, or any bundle asset. Bundle edits are a separate brand-SSOT workflow.

### Out of Scope — Template distribution / Template-First

- `internal/web/` is Go source, not template distribution. No `internal/template/templates/` mirror is created; the CLAUDE.local.md §2 Template-First rule does not apply (see §A.1). `make build` remains required for `templ generate` + embed, but that is a Go build step, not a template sync.

### Out of Scope — Server-contract / behavior changes

- No route, handler, form-field name, validation rule, or persistence seam changes. This is a visual + asset + dead-UI cleanup only. Adding new editable config fields, new tabs of new functionality, or new persistence paths is out of scope (the orphan-panel cleanup only removes-or-restores an EXISTING panel — it adds no new capability).

### Out of Scope — Dark mode

- The console is light-only. If the v2 `[data-theme="dark"]` token block is imported, it stays inert dead code; no dark toggle is wired.

### Out of Scope — i18n dictionary content

- No new interface-language strings are authored. If the orphan-panel cleanup removes a panel, its now-dead `sec.project.*` i18n keys MAY be left in place (inert) or pruned as a housekeeping sub-task, but net-new 4-locale translation work is out of scope.

---

## §F. References

- v2 design canon: `.moai/state/ai-design-system/project/colors_and_type.css`, `assets/characters/`, `README.md`, `round3/02-tokens.json`.
- SPEC-1 precedent: `.moai/specs/SPEC-DESIGN-DOCSV2-001/{spec,design}.md` (token-unification + mascot-placement architecture).
- Current console: `internal/web/assets/console.css`, `board.templ`, `root.templ`, `fieldsets.templ`, `schemaform.go`, `assets.go`, `app.go`.
- Offline-safe invariant: `SPEC-WEB-CONSOLE-004` (REQ-WC4-001/002), `SPEC-WEB-CONSOLE-005` (REQ-WC5-006/011).
