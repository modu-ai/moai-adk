---
id: SPEC-V3R6-DOCS-RC2-DOCSITE-001
title: "v3.0.0-rc2 docs-site version SSOT + 4-locale lifecycle/era + harness-namespace pages"
version: "0.2.0"
status: in-progress
created: 2026-06-19
updated: 2026-06-22
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "docs-site"
lifecycle: spec-anchored
tags: "docs-site, version-ssot, 4-locale, lifecycle, era, harness-namespace, i18n"
era: V3R6
---

# SPEC-V3R6-DOCS-RC2-DOCSITE-001 — v3.0.0-rc2 docs-site version SSOT + lifecycle/era + harness-namespace pages

## §A. Problem Statement

The docs-site (`https://adk.mo.ai.kr/`) presents three classes of drift against the authoritative repo state as of v3.0.0-rc2:

1. **Version SSOT drift.** `docs-site/hugo.toml` `params.version` is `"0.2.0"` (L55) and `releaseDate` is `"2026-05-21"` (L56), while the authoritative repo version is `v3.0.0-rc2` (`system.yaml` `version: v3.0.0-rc2`, git tag `v3.0.0-rc2`). The Hugo `{{< version >}}` shortcode (`docs-site/layouts/shortcodes/version.html`) exists but is referenced by **zero** pages (dead config), so the SSOT is not self-maintaining. Per-locale installation pages hardcode mutually inconsistent old versions: `ko/getting-started/installation.md` L11 + L91 say `2.9.0`, while `en`/`ja`/`zh` `getting-started/installation.md` L81 say `2.0.0`. None reflect v3.0.0-rc2.

2. **Lifecycle/era concept gap.** The MoAI SPEC lifecycle and era-classification system are the backbone of how the project is managed (3-phase close `plan→run→sync` with `sync_commit_sha` in `§E.4`; 5-bucket era classification `V2.x`/`V3R2-R4`/`V3R5`/`V3R6`/`unclassified`; grandfather clause; drift detection scoped to V3R6-only). None of these concepts are documented for end users. Site-wide grep returns **0 hits** for `grandfather`, `era_final`, `ClassifyEra`, `drift detection`, `Mx-phase`, `sync_commit_sha`. The Mx-phase (the retired 4th phase) retirement is invisible to readers.

3. **Harness-namespace policy gap.** Users authoring custom skills need to know which namespaces `moai update` preserves versus overwrites. The doctrine (`moai-*` / `moai-harness-*` / `moai-meta-harness` = template-managed; `harness-*` = user-owned) exists internally (`CLAUDE.local.md §24` + `.moai/docs/harness-namespace-doctrine.md`, originating SPEC `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`, status: completed) but is not surfaced to docs-site readers. The `harness-engineering.md` page (present in all 4 locales) does not yet carry the namespace policy.

## §B. Background & Ground Truth

### §B.1 docs-site stack

- Hugo + hugo-geekdoc theme (NOT Nextra/Next).
- `baseURL = "https://adk.mo.ai.kr/"` (`hugo.toml` L1).
- Vercel deploy (`docs-site/vercel.json`).
- `defaultContentLanguage = "ko"` (L4); `hasCJKLanguage = true` (L6); `enableGitInfo = true` (L7).
- 4-locale structure: `ko` (default, weight 1) / `en` (2) / `ja` (3) / `zh` (4), each under `content/<locale>/`.
- **Exact page parity verified**: 105 `.md` files per locale (×4 = 420 total). Page sets are byte-identical across locales (modulo localized text).
- i18n bundles: `docs-site/i18n/{en,ja,ko,zh-cn}.yaml`.
- Version SSOT: `hugo.toml` `[params]` `version` (L55) + `releaseDate` (L56), consumed by `layouts/shortcodes/version.html`.

### §B.2 already-correct content (DO NOT re-touch)

The following are CURRENT and MUST NOT be modified by this SPEC:

- 8 retained agents + 12 archived (agent-guide.md, all locales).
- `glm-5.2[1m]` model references (2 hits/locale, no stale `glm-5.1`).
- `ultracode` keyword references (3 hits/locale).
- CG mode (`content/<locale>/multi-llm/cg-mode.md`).
- Dynamic workflows (`content/<locale>/claude-code/agentic/workflows.md`).

The model references are current. Re-touching them would introduce regression.

### §B.3 authoritative repo state

- Version: `v3.0.0-rc2` (`system.yaml` `version: v3.0.0-rc2` L48; git tag `v3.0.0-rc2`).
- **NO `v3.0.0` stable tag exists.** User decision: docs-site wording MUST say `"v3.0.0-rc2"`, never `"v3.0.0 stable"`.
- Lifecycle: 3-phase close (`plan` → `run` → `sync`); the 4th "Mx-phase" is **RETIRED** per `SPEC-V3R6-LIFECYCLE-REDESIGN-001` (run-phase code landed). `sync_commit_sha` lives in `§E.4 Sync-phase Audit-Ready Signal`.
- Era system: 5 buckets. `modernEraThreshold = "2026-04-01"`. `IsModern()` returns `true` ONLY for `V3R6`. Grandfather clause: V2.x / V3R2-R4 / V3R5 = `era_final: true`, no drift detection. SSOT: `.claude/rules/moai/workflow/lifecycle-sync-gate.md` + `internal/spec/era.go`.
- Harness namespace: `moai-*` / `moai-harness-*` / `moai-meta-harness` = template-managed (sync overwrites); `harness-*` = user-owned (`moai update` MUST preserve, back up, never delete). Doctrine SSOT: `CLAUDE.local.md §24` + `.moai/docs/harness-namespace-doctrine.md`. Originating SPEC: **`SPEC-V3R6-HARNESS-NAMESPACE-V2-001`** (status: completed). The earlier `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` (my-harness-* rename) is **SUPERSEDED** and MUST NOT be cited.

### §B.4 drift evidence (live grep)

| Claim | Command | Observed |
|-------|---------|----------|
| hugo.toml version stale | `grep -n 'version' docs-site/hugo.toml` | L55 `version = "0.2.0"` |
| hugo.toml releaseDate stale | `grep -n 'releaseDate' docs-site/hugo.toml` | L56 `releaseDate = "2026-05-21"` |
| `{{< version >}}` shortcode dead | `grep -rln '{{< version >}}' docs-site/content/ \| wc -l` | `0` |
| version.html shortcode exists | `ls docs-site/layouts/shortcodes/version.html` | exists (94 bytes) |
| ko install version hardcoded | `sed -n '11p;91p' docs-site/content/ko/getting-started/installation.md` | L11 `v2.9.0`, L91 `--version 2.9.0` |
| en/ja/zh install version hardcoded | `sed -n '81p' docs-site/content/{en,ja,zh}/getting-started/installation.md` | `--version 2.0.0` (each) |
| grandfather absent | `grep -rl 'grandfather' docs-site/content/ \| wc -l` | `0` |
| era_final absent | `grep -rl 'era_final' docs-site/content/ \| wc -l` | `0` |
| ClassifyEra absent | `grep -rl 'ClassifyEra' docs-site/content/ \| wc -l` | `0` |
| drift detection absent | `grep -rl 'drift detection' docs-site/content/ \| wc -l` | `0` case-sensitive hits (1 case-insensitive pre-existing hit in `en/db/_index.md` — `**Drift Detection** — Use /moai db verify to detect ...`, a database-schema drift concept unrelated to SPEC lifecycle drift; this SPEC does NOT remove it) |
| sync_commit_sha absent | `grep -rl 'sync_commit_sha' docs-site/content/ \| wc -l` | `0` |
| Mx-phase absent | `grep -rl 'Mx-phase' docs-site/content/ \| wc -l` | `0` |
| 4-locale parity (page count) | `for loc in ko en ja zh; do find docs-site/content/$loc -name '*.md' \| wc -l; done` | `105 105 105 105` |

## §C. Goals

1. Bring `docs-site/hugo.toml` `params.version` into alignment with the authoritative repo version `v3.0.0-rc2`.
2. Reduce duplicate version literals in PROSE surfaces by wiring the `{{< version >}}` shortcode into at least the intro (`_index.md`) and the prose portions of the 4-locale installation pages. The fenced install-command literal remains a known second touch-point (Hugo shortcodes do not render inside fenced code blocks), so a future bump still requires editing `hugo.toml` + the install-command literal ×4 locales; the shortcode eliminates the prose duplicates.
3. Align all 4-locale installation pages on `v3.0.0-rc2` (identical across ko/en/ja/zh), eliminating the current `2.9.0` (ko) vs `2.0.0` (en/ja/zh) inconsistency.
4. Add a lifecycle/era concept section to `content/<locale>/core-concepts/spec-based-dev.md` (all 4 locales) covering 5-bucket era classification, grandfather clause, 3-phase close with `sync_commit_sha` in `§E.4`, Mx-phase retirement, and V3R6-only drift detection scope.
5. Add a harness-namespace policy section to `content/<locale>/core-concepts/harness-engineering.md` (all 4 locales) covering template-managed vs user-owned separation.
6. Preserve exact 4-locale page-set parity (105 × 4) after all additions land.

## §D. Requirements (GEARS)

### REQ-VLN-001 — hugo.toml version SSOT alignment (Ubiquitous)
The docs-site `hugo.toml` `[params]` block **shall** carry `version = "v3.0.0-rc2"` and `releaseDate = "2026-06-03"` (the `v3.0.0-rc2` git tag date, verified via `git log -1 --format=%ci v3.0.0-rc2` → `2026-06-03 08:00:59 +0900`), so the Hugo-rendered site reports the authoritative repo version.

### REQ-VLN-002 — version shortcode wired into intro + installation prose (Capability gate)
**Where** the `{{< version >}}` shortcode exists in `docs-site/layouts/shortcodes/version.html`, the docs-site **shall** reference that shortcode from at least the 4-locale intro pages (`content/<locale>/_index.md`) and the prose (non-fenced) surfaces of the 4-locale installation pages (`content/<locale>/getting-started/installation.md`), so a future version bump reduces duplicate version literals in PROSE surfaces. Rationale: Hugo shortcodes do NOT render inside fenced code blocks, so the install command (`curl ... | bash -s -- --version <VER>`) keeps its literal version as a known second touch-point; the shortcode eliminates duplicate version literals in PROSE (intro + installation intro text), and a future bump still requires editing `hugo.toml` + the fenced install-command literal (×4 locales). This second touch-point is documented in acceptance.md §E.1.

### REQ-VLN-003 — 4-locale installation version identical, stale literals removed (Ubiquitous)
The 4-locale installation pages (`content/{ko,en,ja,zh}/getting-started/installation.md`) **shall** present the identical current version `v3.0.0-rc2` in the install command AND remove all stale `2.9.0` / `2.0.0` version literals across the page — including install-command lines (ko L91 `--version 2.9.0`; en/ja/zh L81 `--version 2.0.0`), prose version mentions (ko L11 `MoAI-ADK v2.9.0 이상은 ... Apache-2.0 라이선스`), and output-example blocks (ko L157 `moai v2.9.0 ...`). The single exception is the grandfathered historical-license context at ko L16 (`v2.0.0부터 Go 언어로 재작성` — a legitimate historical reference to the Python→Go license change, NOT an install command or current-version claim); that one mention is preserved verbatim. The AC grep covers ALL `2.9.0|2.0.0` literals in the four installation pages and matches this REQ's scope exactly.

### REQ-VLN-004 — atomic 4-locale commit (State-driven)
**While** the version SSOT and installation fixes are being applied, the docs-site **shall** land all 4 locales plus `hugo.toml` in a single atomic commit, so parity is never broken at any commit boundary.

### REQ-LCY-001 — lifecycle/era concept section authored in Korean (Ubiquitous)
The docs-site **shall** add a lifecycle/era concept section to `content/ko/core-concepts/spec-based-dev.md` covering: the 5-bucket era classification (`V2.x`, `V3R2-R4`, `V3R5`, `V3R6`, `unclassified`); the grandfather clause (V2.x/V3R2-R4/V3R5 = protected, `era_final: true`, no drift detection); the 3-phase close (`plan` → `run` → `sync`) with `sync_commit_sha` recorded in `§E.4`; the retirement of the 4th Mx-phase per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`; and drift-detection scope (V3R6-only, `modernEraThreshold = "2026-04-01"`, `IsModern()` true only for V3R6).

### REQ-LCY-002 — lifecycle/era section translated to en/ja/zh (Ubiquitous)
The docs-site **shall** carry a structurally identical translation of the lifecycle/era concept section from REQ-LCY-001 in `content/{en,ja,zh}/core-concepts/spec-based-dev.md`, so all four locales expose the same concept set.

### REQ-LCY-003 — Korean authored before translation (State-driven)
**While** the lifecycle/era section is being produced, the docs-site **shall** finalize the Korean text first and then translate the finalized Korean into en/ja/zh, so the Korean version is the single source for the translations (ko-first workflow).

### REQ-HNS-001 — harness-namespace policy section authored in Korean (Ubiquitous)
The docs-site **shall** add a harness-namespace policy section to `content/ko/core-concepts/harness-engineering.md` covering: template-managed namespaces (`moai-*`, `moai-harness-*`, `moai-meta-harness`) which `moai update` overwrites; user-owned namespaces (`harness-*`) which `moai update` MUST preserve (back up, never delete); and the implication for users authoring custom skills.

### REQ-HNS-002 — harness-namespace section translated to en/ja/zh (Ubiquitous)
The docs-site **shall** carry a structurally identical translation of the harness-namespace policy section from REQ-HNS-001 in `content/{en,ja,zh}/core-concepts/harness-engineering.md`.

### REQ-HNS-003 — cite final doctrine only (Unwanted)
The docs-site **shall not** cite the SUPERSEDED `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`; where the originating SPEC is named, the docs-site **shall** cite `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` (the final doctrine).

### REQ-PAR-001 — 4-locale page-set parity preserved (Ubiquitous)
The docs-site **shall** preserve exact page-set parity across the four locales after all additions land, so the post-change per-locale `.md` file count remains identical across ko/en/ja/zh.

### REQ-PAR-002 — atomic commit for additive content (State-driven)
**While** the lifecycle/era and harness-namespace sections are being added, the docs-site **shall** land all four locales of each section together, so no commit breaks parity.

### REQ-NTR-001 — template-neutrality respected (Unwanted)
The docs-site content **shall not** embed moai-adk internal development artifacts (internal SPEC IDs other than the two named originating SPECs in §B.3, commit SHAs, `/Users/` paths, or internal-audit citations), so the user-facing docs stay generic. (Note: `docs-site/` is project-owned, not under `internal/template/templates/`, so the §25 isolation CI guard does not bind, but the content discipline still applies.)

### REQ-NTR-002 — model/agent content untouched (Unwanted)
The docs-site **shall not** modify the already-correct `glm-5.2[1m]`, 8-retained-agent, 12-archived, `ultracode`, CG-mode, or dynamic-workflows content listed in §B.2.

### REQ-VWD-001 — version wording precision (Unwanted)
The docs-site **shall not** use the phrase `"v3.0.0 stable"` or imply a stable `v3.0.0` tag exists; the version wording **shall** be exactly `"v3.0.0-rc2"` everywhere it appears.

## §E. Constraints

- **4-locale parity is a hard invariant.** The site currently has EXACT parity (105 `.md` × 4). Any new section MUST land in all four locales atomically (same commit), or parity breaks at that commit boundary.
- **ko-first authoring workflow** (user decision): author and finalize the Korean version FIRST, then translate the finalized Korean into en/ja/zh. This is a QUALITY decision (Korean is the default locale, weight 1) and is encoded in plan.md milestone ordering (M3 ko → M4 translate). It is NOT a parallel-from-scratch workflow.
- **Template neutrality (§25)**: `docs-site/` is project-owned (NOT under `internal/template/templates/`), so the isolation CI guard does not bind. But content stays generic/user-facing — no internal SPEC IDs (other than the two named originating SPECs in §B.3), no commit SHAs, no `/Users/` paths.
- **Version wording**: ALWAYS `"v3.0.0-rc2"` — never `"v3.0.0 stable"` (that tag does not exist).
- **Do NOT touch** the already-correct `glm-5.2[1m]` / 8-agent / ultracode / CG-mode / dynamic-workflows content.
- **Do NOT cite** SUPERSEDED `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`; cite `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`.
- **Mermaid node limits**: hugo-geekdoc + diagrams must stay under render ceilings (no oversized graphs in the new sections).
- **Evidence**: every drift claim in §B.4 cites the exact grep command and the observed value, so the plan-phase audit can reproduce the baseline independently.

## §F. Out of Scope

### Out of Scope — README and CHANGELOG version updates
- README.md and README.ko.md version-line updates belong to the sibling SPEC `SPEC-V3R6-DOCS-RC2-README-001`, NOT this SPEC.
- CHANGELOG.md entry for the v3.0.0-rc2 docs cohort belongs to the sibling README SPEC (or the cohort sync-phase), NOT this SPEC.

### Out of Scope — v3.0.0 stable release
- Cutting a stable `v3.0.0` git tag and updating docs-site to `"v3.0.0"` (dropping the `-rc2` suffix) is a FUTURE SPEC. This SPEC targets `v3.0.0-rc2` only.
- Any work that depends on a stable `v3.0.0` tag existing is out of scope.

### Out of Scope — model/agent content re-touch
- The already-correct content (§B.2: glm-5.2[1m], 8 retained agents, 12 archived, ultracode, CG-mode, dynamic-workflows) is NOT modified by this SPEC.

### Out of Scope — internal rule-file / Go-code changes
- `internal/spec/era.go`, `.claude/rules/moai/workflow/lifecycle-sync-gate.md`, and `CLAUDE.local.md §24` are the authoritative SSOTs and are NOT modified by this SPEC. This SPEC only surfaces their content to docs-site readers.
- No Go code changes (`go.mod`, `internal/...`, `cmd/...`).

### Out of Scope — i18n bundle YAML changes
- The `docs-site/i18n/{en,ja,ko,zh-cn}.yaml` UI-string bundles are NOT modified by this SPEC unless the new sections require a new UI string (in which case the addition is 4-locale atomic).

### Out of Scope — my-harness-* rename
- The SUPERSEDED `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` (my-harness-* → moai-harness-* rename) is NOT re-opened by this SPEC. The harness-namespace SECTION added here cites the final doctrine only (`SPEC-V3R6-HARNESS-NAMESPACE-V2-001`).

## §G. Dependencies

- **Authoritative version**: `.moai/config/sections/system.yaml` (`version: v3.0.0-rc2`).
- **Lifecycle/era SSOT**: `.claude/rules/moai/workflow/lifecycle-sync-gate.md` + `internal/spec/era.go`.
- **Harness-namespace doctrine**: `.moai/docs/harness-namespace-doctrine.md` + `CLAUDE.local.md §24`.
- **Sibling SPEC (non-blocking, complementary scope)**: `SPEC-V3R6-DOCS-RC2-README-001` owns README + CHANGELOG; the two SPECs together form the v3.0.0-rc2 docs cohort.
- **Originating SPECs (cited in user-facing content, not modified)**: `SPEC-V3R6-LIFECYCLE-REDESIGN-001`, `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`.

## §H. Success Criteria

1. `hugo.toml` `params.version == "v3.0.0-rc2"` and `releaseDate == "2026-06-03"` (the rc2 tag date).
2. `{{< version >}}` shortcode is referenced by ≥1 page in each of the 4 locales (intro and/or installation prose).
3. All 4 installation pages present the identical version `v3.0.0-rc2`; zero residual `2.9.0` or `2.0.0` literals in installation pages (except the grandfathered historical-license mention at ko L16).
4. Each of the 4 locales has a lifecycle/era section mentioning `grandfather`, `era_final`, 3-phase close, `sync_commit_sha`, and Mx-phase retirement.
5. Each of the 4 locales has a harness-namespace section distinguishing template-managed (`moai-*`) from user-owned (`harness-*`).
6. 4-locale page-set parity preserved (per-locale `.md` count identical across ko/en/ja/zh after all additions).
7. No content mentions `"v3.0.0 stable"` or cites the SUPERSEDED `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`.
