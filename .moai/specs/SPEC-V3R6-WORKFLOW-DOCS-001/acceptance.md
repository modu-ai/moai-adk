# Acceptance Criteria — SPEC-V3R6-WORKFLOW-DOCS-001

Tier M · 11 ACs (ceiling 16). **AC discipline (binding)**: every AC below was adopted ONLY after observing RED on the pre-implementation tree by actually running its command. Each AC carries two paired facts: RED-NOW (exact command + observed output + why it fails) and GREEN-PATH (which milestone flips it + passing output). All RED observations were measured on 2026-08-25 in worktree t273, branch `WT-workflow-docs`, HEAD `db1362739`.

## §D AC Matrix

### AC-WFD-001 — card-class section on kanban-mode.md, 4 locales (REQ-WFD-001)

**Given** the docs-site kanban-mode page **When** a reader looks for card classes **Then** all four locales carry a section with the normative heading token (ko `카드 클래스` / en `Card Classes` / ja `カードクラス` / zh `卡片类别`) presenting A/B/C semantics.

- RED-NOW: `grep -rn -E "카드 클래스|Class A|클래스 A|类别 A|卡片类别" docs-site/content/{ko,en,ja,zh}/advanced/kanban-mode.md README.ko.md README.md README.ja.md README.zh.md` → **no output (exit 1)** — zero matches in all 8 files; the concept is absent from every public surface.
- GREEN-PATH: M3 → `grep -c "<locale token>" docs-site/content/<locale>/advanced/kanban-mode.md` returns ≥ 1 in each of the 4 locales. Missing-in-one = FAIL (both directions counted).

### AC-WFD-002 — card-class table in README kanban section, 4 files (REQ-WFD-002)

**Given** the README 4-file set **When** the kanban section is read **Then** each file carries the compact card-class table (ko canonical, en/ja/zh derived).

- RED-NOW: same 8-file grep as AC-WFD-001 — the 4 README legs returned 0 matches (included in the measured no-output result).
- GREEN-PATH: M3 → `grep -c "카드 클래스" README.ko.md` ≥ 1; `grep -c "Card Classes" README.md` ≥ 1; same for ja/zh tokens in `README.ja.md` / `README.zh.md`.

### AC-WFD-003 — factory-mode.md exists in 4 locales + workers.json anchor (REQ-WFD-003)

**Given** the advanced section **When** Factory Mode is sought **Then** `advanced/factory-mode.md` exists in all 4 locales and names the slot registry.

- RED-NOW: `find docs-site/content -name "*factory*" -o -name "*lifecycle*"` → **no output** (0 files); `ls docs-site/content/ko/advanced/` lists 34 md files, no `factory-mode.md`.
- GREEN-PATH: M2 → `ls docs-site/content/{ko,en,ja,zh}/advanced/factory-mode.md` lists 4 files; `grep -c "workers.json" docs-site/content/ko/advanced/factory-mode.md` ≥ 1.

### AC-WFD-004 — per-lane cap anchor on factory-mode.md (REQ-WFD-003 / REQ-WFD-011)

**Given** the factory page **When** the concurrency model is documented **Then** the launcher-injected cap constant appears (grounded: `README.ko.md:80`, `advanced/kanban-mode.md:239`).

- RED-NOW: page absent — `grep -c "CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS" docs-site/content/ko/advanced/factory-mode.md` → **"No such file or directory" (exit 2)**.
- GREEN-PATH: M2 → grep returns ≥ 1 (constant locale-verbatim).

### AC-WFD-005 — kanban-mode.md keeps factory summary + link, 4 locales (REQ-WFD-004)

**Given** the dedicated factory page exists **When** kanban-mode.md's factory section is read **Then** it summarizes and links to the new page in each locale.

- RED-NOW: `grep -c "factory-mode" docs-site/content/ko/advanced/kanban-mode.md` → **0** (measured).
- GREEN-PATH: M2 → `grep -c "factory-mode" docs-site/content/<locale>/advanced/kanban-mode.md` ≥ 1 ×4 locales (`factory-mode` is a locale-verbatim slug).

### AC-WFD-006 — spec-lifecycle.md exists in 4 locales (REQ-WFD-005)

**Given** core-concepts **When** the lifecycle flow is sought **Then** `core-concepts/spec-lifecycle.md` exists in all 4 locales.

- RED-NOW: `find docs-site/content -name "*lifecycle*"` → **no output**; `ls docs-site/content/ko/core-concepts/` lists 11 md files, no `spec-lifecycle.md`.
- GREEN-PATH: M1 → `ls docs-site/content/{ko,en,ja,zh}/core-concepts/spec-lifecycle.md` lists 4 files.

### AC-WFD-007 — three gates + tier thresholds on spec-lifecycle.md (REQ-WFD-005 / REQ-WFD-011)

**Given** the lifecycle page **When** the gates section is read **Then** Implementation Kickoff Approval is named and the per-tier PASS thresholds appear (numbers locale-verbatim).

- RED-NOW: page absent — `grep -c "Implementation Kickoff" docs-site/content/ko/core-concepts/spec-lifecycle.md` → **"No such file or directory" (exit 2)**.
- GREEN-PATH: M1 → `grep -c "Implementation Kickoff" <page>` ≥ 1; `grep -c "0.85" <page>` ≥ 1 (implies the 0.75/0.80/0.85 row); ×4 locales for both greps.

### AC-WFD-008 — bidirectional cross-link spec-lifecycle ↔ spec-based-dev (REQ-WFD-006)

**Given** both pages exist **When** either is read **Then** each links the other with the division-of-labor statement.

- RED-NOW: `grep -c "spec-lifecycle" docs-site/content/ko/core-concepts/spec-based-dev.md docs-site/data/menu/main.yaml docs-site/content/ko/core-concepts/_meta.yaml` → **0 / 0 / 0** (measured).
- GREEN-PATH: M1 → `grep -c "spec-lifecycle" docs-site/content/<l>/core-concepts/spec-based-dev.md` ≥ 1 and `grep -c "spec-based-dev" docs-site/content/<l>/core-concepts/spec-lifecycle.md` ≥ 1, ×4 locales.

### AC-WFD-009 — nav registration (gated on lead approval) (REQ-WFD-007)

**Given** the lead approved the nav proposal (M0) **When** navigation is inspected **Then** both pages are registered in main.yaml and the per-locale `_meta.yaml` files.

- RED-NOW: `grep -c "factory-mode" docs-site/data/menu/main.yaml docs-site/content/ko/advanced/_meta.yaml` → **0 / 0**; `grep -c "spec-lifecycle" docs-site/data/menu/main.yaml docs-site/content/ko/core-concepts/_meta.yaml` → **0 / 0** (measured). Icon SVG cases verified present for reuse: `grep -n "school|flash_on" docs-site/layouts/partials/menu.html` → `:58 flash_on`, `:65 school`.
- GREEN-PATH: M4 (only after M0 approval) → all four greps ≥ 1 (main.yaml carries 4-locale name maps + `ref: /advanced/factory-mode` and `ref: /core-concepts/spec-lifecycle`; `_meta.yaml` ×4 locales each list both slugs).

### AC-WFD-010 — 4-locale page-count parity 150 → 152 (REQ-WFD-010)

**Given** the measured baseline **When** the change lands **Then** every locale counts exactly 152 pages (not 131→132 as the delegation draft quoted — see plan.md §B.1).

- RED-NOW (baseline): `find docs-site/content -name '*.md' | cut -d/ -f3 | sort | uniq -c` → **"150 en / 150 ja / 150 ko / 150 zh"** — at 150 the two new pages do not exist, so the target state is unreachable.
- GREEN-PATH: after M1+M2 → same command returns 152 for all four locales. Any locale ≠ 152 = FAIL.

### AC-WFD-011 — sync-auditor gate named on spec-lifecycle.md (REQ-WFD-011)

**Given** the lifecycle page **When** the third gate is documented **Then** the sync-auditor agent is named (4-dimension quality scoring owner).

- RED-NOW: page absent — `grep -c "sync-auditor" docs-site/content/ko/core-concepts/spec-lifecycle.md` → **"No such file or directory" (exit 2)**.
- GREEN-PATH: M1 → grep ≥ 1, ×4 locales (agent identifier locale-verbatim).

## §D.1 Severity

All 11 ACs are **MUST-PASS** (closure gates). AC-WFD-009 additionally carries the process gate: executing it before M0 approval is a violation even if the grep would pass.

## §D.2 Traceability

| REQ | ACs | Verification method |
|---|---|---|
| REQ-WFD-001 | AC-WFD-001 | grep, 4 locales |
| REQ-WFD-002 | AC-WFD-002 | grep, 4 files |
| REQ-WFD-003 | AC-WFD-003, AC-WFD-004 | file existence + grep anchors |
| REQ-WFD-004 | AC-WFD-005 | grep ×4 |
| REQ-WFD-005 | AC-WFD-006, AC-WFD-007, AC-WFD-011 | file existence + grep anchors |
| REQ-WFD-006 | AC-WFD-008 | grep, bidirectional |
| REQ-WFD-007 | AC-WFD-009 | grep, gated milestone |
| REQ-WFD-008 | — (process) | same-PR review at sync; locale-parity gate |
| REQ-WFD-009 | — (regression) | verify recipe §2/§3/§5 (plan §E.2) |
| REQ-WFD-010 | AC-WFD-010 | page count + ratchet `comm -23` empty (plan §E.2/§E.3) |
| REQ-WFD-011 | AC-WFD-004, AC-WFD-007, AC-WFD-011 + canon spot-check | grep anchors + cross-read (plan §E.4) |
| REQ-WFD-012 | — (regression) | full verify recipe (plan §E.2) |

## §D.3 RED evidence table (verbatim, tree `db1362739`, 2026-08-25)

| # | Command | Observed (RED) |
|---|---|---|
| E1 | `grep -rn -E "카드 클래스\|Class A\|클래스 A\|类别 A\|卡片类别" <kanban-mode.md ×4 + README ×4>` | no output (exit 1) |
| E2 | `grep -c "factory-mode" docs-site/content/ko/advanced/kanban-mode.md docs-site/data/menu/main.yaml docs-site/content/ko/advanced/_meta.yaml` | 0 / 0 / 0 |
| E3 | `grep -c "spec-lifecycle" docs-site/content/ko/core-concepts/spec-based-dev.md docs-site/data/menu/main.yaml docs-site/content/ko/core-concepts/_meta.yaml` | 0 / 0 / 0 |
| E4 | `find docs-site/content -name "*lifecycle*" -o -name "*factory*"` | no output |
| E5 | `find docs-site/content -name '*.md' \| cut -d/ -f3 \| sort \| uniq -c` | 150 en / 150 ja / 150 ko / 150 zh |
| E6 | `grep -c "^## " README.md README.ko.md README.ja.md README.zh.md` | 12 / 12 / 12 / 12 |
| E7 | `ls docs-site/content/ko/advanced/` | 34 md files, no factory-mode.md |
| E8 | `ls docs-site/content/ko/core-concepts/` | 11 md files, no spec-lifecycle.md |
| E9 | `grep -n "kanban-mode\|school\|flash_on" docs-site/data/menu/main.yaml docs-site/layouts/partials/menu.html` | leaf refs :484/:716; SVG cases flash_on menu.html:58, school menu.html:65 |
| E10 | `wc -l docs-site/.locale-parity-baseline` | 58 lines (ratchet baseline exists) |
| E11 | `grep -rn "workers.json" README.ko.md internal/` | README.ko.md:80; internal/kanban/factory_slots.go:48 |
| E12 | `grep -rn "최대 10" README.ko.md docs-site/content/ko/advanced/kanban-mode.md` | README.ko.md:80; kanban-mode.md:239 |

## §D.4 Indirect verification

Canon values that lack a distinctive grep anchor (Tier artifact sets 2/3/5 files, REQ/AC ceilings 8/16/25, Class B evidence-to-progress-record nuance, /clear strategy, Route A/B summary) are verified by plan §E.4 cross-read against the canonical rule files during M5, plus plan-auditor/sync-auditor review.

## §D.5 Closure gates

All 11 ACs GREEN + plan §E regression gates green + no NEW `.locale-parity-baseline` entries + sync-auditor verdict — then and only then does the card leave run.

## §D.6 Disqualified criteria (already passing today — regression gates, NOT ACs)

Per the AC discipline, the following checks pass on the current tree and no planned change is meant to flip them; they are guarded as regression gates in plan §E.2 instead: warning-free hugo build; sitemap existence; URL blacklist 0; Mermaid LR/RL 0; README H2 parity 12/12/12/12 (E6); body-emoji scan; version-string sync.

## §D.7 Forward-looking

After this SPEC lands, the docs-site covers card classes, Factory Mode, and the integrated lifecycle. Residual known gaps (deliberately out of scope, per gap map): session-handoff/context-window documentation; Origin-Trail internals; multi-llm operating-procedure restructure.
