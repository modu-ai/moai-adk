---
id: SPEC-V3R6-DOCS-RC2-DOCSITE-001
title: "v3.0.0-rc2 docs-site version SSOT + lifecycle/era + harness-namespace — implementation plan"
version: "0.2.0"
status: draft
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "docs-site"
lifecycle: spec-anchored
tags: "docs-site, version-ssot, 4-locale, lifecycle, era, harness-namespace, i18n"
era: V3R6
---

# Implementation Plan — SPEC-V3R6-DOCS-RC2-DOCSITE-001

## §A. Context

This plan implements the docs-site portion of the v3.0.0-rc2 release cohort. The repo is at `v3.0.0-rc2` (no stable `v3.0.0` tag yet); the docs-site is stuck at `0.2.0` in `hugo.toml`, has mutually inconsistent hardcoded install versions across locales (`2.9.0` in ko vs `2.0.0` in en/ja/zh), and is missing two entire concept areas (lifecycle/era and harness-namespace) that are core to how MoAI is operated.

The user decision driving this plan: **ko-first authoring workflow**. Korean is the default locale (weight 1). The user wants the Korean text of each new section authored and finalized FIRST, then translated to en/ja/zh from the finalized Korean. This is encoded as a milestone ordering constraint (M3 → M4 for lifecycle/era; M5 ko → M5 translate for harness-namespace). The plan MUST NOT structure the work as "do all 4 locales in parallel from scratch" — ko is the source.

The hard invariant is **4-locale page-set parity** (currently 105 `.md` × 4 = 420, byte-identical page sets). Every new section MUST land in all four locales in a single commit, or parity breaks at that commit boundary. ko-first authoring is a QUALITY decision for in-progress text; the COMMIT is always 4-locale atomic.

This is a **plan-phase only** artifact set. Run-phase edits to docs-site are NOT in scope for this delegation.

## §B. Known Issues (baseline, reproduced 2026-06-19)

| Issue | Evidence | Affected REQ |
|-------|----------|--------------|
| hugo.toml version stale | `hugo.toml:55 version = "0.2.0"` | REQ-VLN-001 |
| hugo.toml releaseDate stale | `hugo.toml:56 releaseDate = "2026-05-21"` | REQ-VLN-001 |
| `{{< version >}}` shortcode dead | `grep -rln '{{< version >}}' docs-site/content/ \| wc -l` → `0` | REQ-VLN-002 |
| ko install version wrong | `content/ko/getting-started/installation.md:11` (`v2.9.0` Apache-2.0 prose), `:91` (`--version 2.9.0`), `:157` (`moai v2.9.0` output example); L16 (`v2.0.0` historical-license) PRESERVED | REQ-VLN-003 |
| en/ja/zh install version wrong | `content/{en,ja,zh}/getting-started/installation.md:81` (`--version 2.0.0`) | REQ-VLN-003 |
| `/Users/` baseline (NOT a defect) | `grep -rln '/Users/' docs-site/content/ \| wc -l` → `21` (4 locales × 5 file-types + session-memo; legitimate WSL/worktree examples; this SPEC adds zero new) | REQ-NTR-001 (baseline-aware) |
| lifecycle/era concept absent | `grep -rl 'grandfather\|era_final\|ClassifyEra\|drift detection\|Mx-phase\|sync_commit_sha' docs-site/content/ \| wc -l` → `0` each | REQ-LCY-001/002 |
| harness-namespace policy absent from page | `content/<locale>/core-concepts/harness-engineering.md` lacks namespace section | REQ-HNS-001/002 |
| SUPERSEDED SPEC may be cited | `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` is SUPERSEDED; risk of miscitation | REQ-HNS-003 |

## §C. Pre-flight (before M1)

1. **Re-verify baseline parity** at run-phase start: `for loc in ko en ja zh; do find docs-site/content/$loc -name '*.md' | wc -l; done` → expect `105 105 105 105`. If parity is already broken, stop and report — the "preserve parity" AC cannot be evaluated against a broken baseline.
2. **Confirm rc2 tag date** for the `releaseDate` field: `git log -1 --format=%ci v3.0.0-rc2` → `2026-06-03 08:00:59 +0900` → use `"2026-06-03"` verbatim. **M1 MUST land `hugo.toml` `version` + `releaseDate` in ONE shot** (no separate releaseDate follow-up commit), so the AC-VLN-004 pickaxe (`-S '--version 2.0.0'` on the installation page) binds unambiguously to the version-fix commit. A split version/releaseDate commit would make the hugo.toml pickaxe ambiguous.
3. **Confirm shortcode contract**: `cat docs-site/layouts/shortcodes/version.html` → confirm it renders `{{ .Site.Params.version }}` so wiring `{{< version >}}` actually displays the SSOT value.
4. **Read SSOT sources** for the new sections: `lifecycle-sync-gate.md` (era table + grandfather clause), `.moai/docs/harness-namespace-doctrine.md` (namespace matrix). Do NOT paraphrase the 5-bucket era table or the template-managed/user-owned split — copy the structure verbatim.

## §D. Constraints (carried from spec.md §E)

- 4-locale parity is a hard invariant (atomic 4-locale commits).
- ko-first authoring (M3 ko → M4 translate; M5 ko → M5 translate).
- Version wording exactly `"v3.0.0-rc2"` (never `"v3.0.0 stable"`).
- Do NOT touch already-correct model/agent content (§B.2 of spec.md).
- Do NOT cite SUPERSEDED `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`.
- No internal SPEC IDs (other than the two named originating SPECs), no commit SHAs, no `/Users/` paths in user-facing content.
- Mermaid diagrams in new sections must stay under hugo-geekdoc render ceilings.

## §E. Self-Verification (run-phase deliverable, E1–E7 skeleton)

This section is a skeleton; the manager-develop run-phase populates E1–E7 with observed evidence per the 5-Section Evidence-Bearing Report format.

### §E.1 Plan-phase Audit-Ready Signal
- _Plan-phase artifacts (spec.md + plan.md + acceptance.md) authored. Status: draft. Era: V3R6._

### §E.2 Run-phase Evidence
_<pending run-phase>_

### §E.3 Run-phase Audit-Ready Signal
_<pending run-phase>_

### §E.4 Sync-phase Audit-Ready Signal
_<pending sync-phase>_

## §F. Milestones (6 milestones, ko-first ordering)

> **Milestone philosophy**: every milestone that adds or modifies 4-locale content commits **atomically across all four locales**. The ko-first constraint governs the AUTHORING SEQUENCE inside a milestone (M3 authors ko first, then M4 translates), NOT the commit boundary (the commit still includes all four locales once M4 lands).
>
> Per the time-estimation rule, milestones carry **priority labels** (High/Medium/Low), not time estimates. Sequencing is enforced by milestone dependency arrows (M2 depends on M1; M4 depends on M3; M5 has internal ko→translate sub-steps; M6 depends on all prior).

### M1 — hugo.toml version SSOT + shortcode decision (Priority High)

**Scope**:
- `docs-site/hugo.toml`: `params.version` `"0.2.0"` → `"v3.0.0-rc2"` (L55); `releaseDate` → rc2 tag date from `git log -1 --format=%ci v3.0.0-rc2` (L56).
- **Decision recorded in this plan (already made)**: WIRE the `{{< version >}}` shortcode into intro + installation pages (preferred over removing the dead shortcode), so the SSOT becomes self-maintaining. Rationale: removing the shortcode saves nothing (it is 94 bytes and harmless); wiring it makes future version bumps touch `hugo.toml` only.

**Files touched**: `docs-site/hugo.toml`.

**Verification**: `grep -n 'version = ' docs-site/hugo.toml` → `v3.0.0-rc2`; `hugo` (or `hugo --renderToMemory`) does not error on the toml.

**AC binding**: AC-VLN-001, AC-VLN-002.

### M2 — 4-locale installation version fix + shortcode wiring (atomic) (Priority High, blocked-by M1)

**Scope**:
- 4-locale installation pages `content/{ko,en,ja,zh}/getting-started/installation.md`:
  - Replace the hardcoded install-command version with `v3.0.0-rc2` (or wire `{{< version >}}` if the install command line tolerates a shortcode — record which in the commit body).
  - ko: remove `2.9.0` literals at L11 (Apache-2.0 prose), L91 (`--version 2.9.0` install command), L157 (`moai v2.9.0` output example). PRESERVE L16 (`v2.0.0부터 Go 언어로 재작성` — grandfathered historical-license context).
  - en/ja/zh: remove `2.0.0` literal at L81 (install command).
  - Verify all four present the IDENTICAL version (no residual old literals).
- 4-locale intro pages `content/{ko,en,ja,zh}/_index.md`: if the intro mentions a version, wire `{{< version >}}` there too (per REQ-VLN-002).

**Atomic constraint**: all 4 locales + intro pages in ONE commit. Parity MUST NOT break at this commit.

**Verification (run-phase)**:
- `grep -rn '2\.9\.0\|2\.0\.0' docs-site/content/*/getting-started/installation.md` → exactly 1 hit (ko L16 grandfathered historical-license `v2.0.0부터 Go 언어로 재작성`); all other stale literals (ko L11/L91/L157, en/ja/zh L81) removed. This matches AC-VLN-003's widened scope (all stale version literals in installation pages, not just install-command lines).
- Per-locale install command lines are byte-identical modulo localized surrounding prose.

**AC binding**: AC-VLN-003, AC-VLN-004.

### M3 — lifecycle/era section: Korean author + finalize (Priority High)

**Scope**:
- Author the lifecycle/era concept section in KOREAN ONLY first, added to `content/ko/core-concepts/spec-based-dev.md`.
- Content MUST cover (per REQ-LCY-001):
  1. 5-bucket era classification table (`V2.x`, `V3R2-R4`, `V3R5`, `V3R6`, `unclassified`) with periods.
  2. Grandfather clause: V2.x / V3R2-R4 / V3R5 = `era_final: true`, protected, no drift detection.
  3. 3-phase close: `plan` → `run` → `sync`, with `sync_commit_sha` recorded in `§E.4 Sync-phase Audit-Ready Signal`.
  4. Mx-phase (the retired 4th phase) retirement per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`.
  5. Drift-detection scope: V3R6-only; `modernEraThreshold = "2026-04-01"`; `IsModern()` true only for V3R6.
- **Finalize the Korean text**: review for clarity, accuracy vs `lifecycle-sync-gate.md`, and prose quality. The finalized Korean is the SOURCE for M4.

**Do NOT touch en/ja/zh in this milestone.** Parity is temporarily broken at the ko-only commit and restored at M4. This is intentional and encoded in REQ-LCY-003 (ko-first workflow). The parity AC (AC-PAR-001) is evaluated **end-of-SPEC at M6**, NOT at every intermediate commit — this is resolved authoritatively in acceptance.md §E.2 (edge case 2: "Interim ko-only commit parity"). Intermediate ko-only commits are allowed inside the run-phase branch; only the atomic-4-locale translation commits merge to main.

**Files touched**: `content/ko/core-concepts/spec-based-dev.md`.

**AC binding**: AC-LCY-001.

### M4 — lifecycle/era section: translate ko → en/ja/zh (atomic) (Priority High, blocked-by M3)

**Scope**:
- Translate the finalized Korean section from M3 into en, ja, zh.
- Add the structurally identical section to `content/{en,ja,zh}/core-concepts/spec-based-dev.md`.
- "Structurally identical" = same headings, same table rows, same concept ordering. Localized prose, NOT mechanical word-by-word.
- Preserve any Mermaid diagram structure across locales (translate node labels only).

**Atomic constraint**: all 3 translations land in ONE commit (with the ko section already from M3, this restores 4-locale parity for the lifecycle/era section).

**Verification (run-phase)**:
- Per-locale `grep -c 'grandfather\|era_final\|sync_commit_sha\|3-phase\|plan.*run.*sync'` returns ≥1 hit each.
- The new section heading appears in all 4 locales.

**AC binding**: AC-LCY-002, AC-LCY-003, AC-PAR-001.

### M5 — harness-namespace section: ko author → 4-locale translate (atomic) (Priority Medium)

**Scope** (single milestone, ko-first sub-steps inside):
- **M5a (ko)**: Author the harness-namespace policy section in Korean, added to `content/ko/core-concepts/harness-engineering.md`. Content MUST cover (per REQ-HNS-001):
  1. Template-managed namespaces: `moai-*`, `moai-harness-*`, `moai-meta-harness` — `moai update` overwrites these on sync.
  2. User-owned namespaces: `harness-*` — `moai update` MUST preserve (back up, never delete).
  3. Implication for users authoring custom skills: put your custom skills under `harness-*` to survive `moai update`.
  4. Cite `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` (NOT the SUPERSEDED `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`).
- **M5b (translate)**: Translate the finalized Korean into en/ja/zh, add to `content/{en,ja,zh}/core-concepts/harness-engineering.md`. Atomic commit restores parity.

**Files touched**: `content/{ko,en,ja,zh}/core-concepts/harness-engineering.md` (4 files, one atomic commit for the translations; ko may land separately as M5a per the same ko-first interim-parity convention as M3/M4).

**Verification (run-phase)**:
- Per-locale `grep -c 'harness-\*\|moai-harness\|template-managed\|user-owned'` returns ≥1 hit each.
- The SUPERSEDED SPEC ID does NOT appear: `grep -rn 'HARNESS-MOAI-NAMESPACE' docs-site/content/` → 0 hits.

**AC binding**: AC-HNS-001, AC-HNS-002, AC-HNS-003.

### M6 — parity preservation + self-verify (Priority High, blocked-by M1–M5)

**Scope**:
- Confirm 4-locale page-set parity is intact after all additions: `for loc in ko en ja zh; do find docs-site/content/$loc -name '*.md' | wc -l; done` → identical counts across locales (105 + 0 new pages, since both additions are sections inside EXISTING files; if a new page was needed instead, +1 × 4 atomic).
- Confirm no regression on the already-correct content (§B.2 of spec.md): `grep -rln 'glm-5\.2\[1m\]' docs-site/content/ko/ | wc -l` ≥ 2; 8-agent content intact.
- Confirm no forbidden content leaked: `grep -rn 'v3\.0\.0 stable\|HARNESS-MOAI-NAMESPACE\|/Users/' docs-site/content/` → 0 hits.
- Run `hugo` build (or Vercel preview) to confirm the new sections render and Mermaid diagrams (if any) stay under render ceilings.
- Populate §E.2 + §E.3 in this plan with observed evidence (the 5-Section Evidence-Bearing Report format).

**Verification (run-phase, end-of-SPEC gate)**:
- 4-locale page counts identical.
- All AC-* in acceptance.md independently reproducible by the run-phase verifier.

**AC binding**: AC-PAR-001, AC-PAR-002, AC-NTR-001, AC-NTR-002, AC-VWD-001.

## §G. Anti-Patterns

- **AP-1 — Parallel-from-scratch 4-locale authoring**: structuring M3/M5 as "do all 4 locales in parallel from scratch". This violates REQ-LCY-003 / the ko-first decision. Korean MUST be finalized first; en/ja/zh translate from the finalized Korean.
- **AP-2 — Non-atomic 4-locale commit**: landing ko in one commit, en in another, ja in a third, zh in a fourth. This breaks parity at three commit boundaries. The atomic-4-locale-commit rule (REQ-VLN-004, REQ-PAR-002) binds the COMMIT, not the authoring sequence.
- **AP-3 — Removing the `{{< version >}}` shortcode instead of wiring it**: the decision is to WIRE it (self-maintaining SSOT), not remove it. Removing saves nothing and loses the self-maintaining property.
- **AP-4 — Citing the SUPERSEDED SPEC**: naming `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` anywhere in user-facing content. Use `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` only.
- **AP-5 — "v3.0.0 stable" wording**: no stable `v3.0.0` tag exists. Always `v3.0.0-rc2`.
- **AP-6 — Touching already-correct content**: modifying glm-5.2[1m], 8-agent, ultracode, CG-mode, or dynamic-workflows content. These are CURRENT; re-touching introduces regression.
- **AP-7 — Paraphrasing the era table**: the 5-bucket era table and the template-managed/user-owned namespace split are SSOTs. Copy the structure verbatim from `lifecycle-sync-gate.md` and `harness-namespace-doctrine.md`; do not invent a different taxonomy.

## §H. Cross-References

- **spec.md**: `.moai/specs/SPEC-V3R6-DOCS-RC2-DOCSITE-001/spec.md` (GEARS requirements REQ-VLN/REQ-LCY/REQ-HNS/REQ-PAR/REQ-NTR/REQ-VWD).
- **acceptance.md**: `.moai/specs/SPEC-V3R6-DOCS-RC2-DOCSITE-001/acceptance.md` (AC-VLN/AC-LCY/AC-HNS/AC-PAR/AC-NTR/AC-VWD).
- **Sibling SPEC (complementary)**: `SPEC-V3R6-DOCS-RC2-README-001` (README + CHANGELOG — not modified by this SPEC).
- **Authoritative SSOTs (read-only, not modified)**:
  - `.moai/config/sections/system.yaml` (version).
  - `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (era table + grandfather clause).
  - `internal/spec/era.go` (classification implementation).
  - `.moai/docs/harness-namespace-doctrine.md` + `CLAUDE.local.md §24` (namespace policy).
- **Originating SPECs cited in user-facing content**: `SPEC-V3R6-LIFECYCLE-REDESIGN-001` (Mx-phase retirement), `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` (final namespace doctrine).
