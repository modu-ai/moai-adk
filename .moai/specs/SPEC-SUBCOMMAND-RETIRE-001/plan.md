# Plan — SPEC-SUBCOMMAND-RETIRE-001

## §A. Context

Tier **L** removal SPEC. Surface spans ~60 files across two trees (local `.claude/` +
`internal/template/templates/.claude/`) plus catalog, CI-guard tests, and docs-site
(4-locale). Cycle type: DDD-style (PRESERVE the CI-guard tests as the safety net; IMPROVE
by removing artifacts with continuous `go test ./internal/template/...` runs). All
implementation is delegated to `manager-develop` per milestone; this plan is the
delegation map. See research.md for the verified inventory and design.md for the four
scope/sequencing decisions.

## §B. Known Issues / Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| CI-guard count cascade (skills 35→28, total 42→35) goes RED | HIGH | M2 atomic: catalog entries + dirs + constants + build in one milestone (design.md §D) |
| `gen-catalog-hashes.go --all` errors on a removed skill's missing SKILL.md | HIGH | Delete catalog entries BEFORE/with the dirs, then `make build` (generator is hash-only) |
| FROZEN `design/constitution.md` / zone-registry mirror edited by mistake | HIGH | Exempted in REQ-SCR-008 scope; AP-1 in design.md §F |
| Template-only or local-only residue after removal | MED | M6 dual-tree grep gate (REQ-SCR-006) |
| Orphaned cross-ref in retained `moai-domain-humanize` | MED | M3 rewrite (REQ-SCR-010) |
| docs-site 4-locale parity defect | MED | M5 symmetric removal across en/ko/ja/zh |
| Security capability gap | MED | M3 router note + replacement path (REQ-SCR-007) |

## §C. Pre-flight (run-phase entry checks)

1. `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` (multi-session race).
2. `go test ./internal/template/...` GREEN on the pre-removal tree (baseline — confirms the safety net is green before any change).
3. `make build` succeeds on the pre-removal tree (baseline build sanity).

## §D. Constraints

- Template-First Rule (CLAUDE.local.md §2): every removal hits the template tree; `make build` regenerates `embedded.go`.
- Both-tree cleanliness (REQ-SCR-006/012): the 7 removed basenames absent from both trees; counts asymmetric by design (28 template / 30 local, the 2 extra = user-owned `harness-*`); no raw full-tree diff.
- No FROZEN edits (design.md §A).
- Re-derive catalog counts by recount, not from this plan's numbers (verification-claim-integrity).

## §E. Self-Verification (plan-phase audit-ready signal → progress.md §E.1)

- [x] SPEC ID regex self-check printed (`→ PASS`).
- [x] 12-field frontmatter validated (status: draft, P2, ISO dates, comma-tag string).
- [x] Out of Scope section present with `### Out of Scope —` H3 sub-headings + `-` bullets.
- [x] Inventory verified by file existence + grep (research.md), not assumed.
- [x] CI-guard test list enumerated with exact constants (research.md §C).
- [x] progress.md §E skeleton emitted (placeholders only).

## §F. Milestones (priority-ordered, no time estimates)

### M1 — Remove command + workflow files + SKILL.md router cleanup
**Files (both trees):**
- Delete `.claude/commands/moai/{brain,design,e2e,coverage,security}.md` (template: `brain.md`/`design.md`/`security.md` plain; `coverage.md.tmpl`/`e2e.md.tmpl` templated).
- Delete `.claude/skills/moai/workflows/{brain,design,e2e,coverage,security}.md`.
- Edit `.claude/skills/moai/SKILL.md` (both trees): drop the 5 from the header subcommand list (lines 5-7); remove Priority 1 routing lines (58, 62, 71, 72, 74); reroute Priority 3 "Security language → security" (line 87) to review/Agent security scope; delete Quick Reference blocks for brain (105-111), security (143-148), coverage (192-197), e2e (199-204), design (206-212).
**Verify:** `grep -rn 'brain\|/moai design\|/moai e2e\|/moai coverage\|/moai security'` in SKILL.md (both trees) → only retained-context matches; router no longer routes the 5.

### M2 — Remove 7 skills + catalog.yaml entries + count-constant reconciliation + build (ATOMIC)
**Files:**
- Edit `internal/template/catalog.yaml`: delete 7 entries — `moai-domain-ideation` (lines ~91-94) + `moai-domain-research` (~96-99) from `core.skills`; `moai-domain-brand-design`/`moai-domain-copywriting`/`moai-domain-design-handoff`/`moai-workflow-design`/`moai-workflow-gan-loop` from `optional_packs.design` (169-196 region) — **preserve `moai-domain-humanize`**.
- Delete 7 skill directories in BOTH trees (`.claude/skills/<name>/` + `internal/template/templates/.claude/skills/<name>/`).
- Update CI-guard count constants (re-derive by recount; expected 28 / 35):
  - `internal/template/catalog_tier_audit_test.go`: `expectedSkillCount` 35 → 28 (+ history comment).
  - `internal/template/catalog_loader_test.go`: `expectedTotal` 42 → 35 (+ history comment).
  - `internal/template/embed_catalog_test.go`: `wantTotal` 42 → 35 (+ history comment).
- `internal/template/commands_audit_test.go`: delete `TestBrainCommandThinPattern` (hardcodes removed `brain.md`).
- `internal/template/agentless_audit_test.go`: **THREE package-level path-list edits required** (each list is read via `fs.ReadFile`→`t.Fatalf`, so a stale entry pointing at a deleted file goes RED):
  1. Remove `.claude/skills/moai/workflows/coverage.md` from `var utilitySkillPaths` (L31) — consumed by `TestAgentlessUtilityNoLLMControlFlow` (L77) + `TestUtilitySkillsContainModeFlagIgnoredSentinel` (L129) = 2 subtests.
  2. Remove `.claude/skills/moai/workflows/design.md` from `var implementationSkillPaths` (L44) — consumed by `TestImplementationSkillsContainPipelineRejectionSentinel` (L159) = 1 subtest.
  3. Remove `.claude/skills/moai/workflows/design.md` from `runDesignSkillPaths` in `TestRunDesignSkillsContainModeUnknownSentinel` (L186) — keep `run.md`; rename func if appropriate.
  - **Stale-comment hygiene (D3, same milestone):** update L28 "lists the 5 utility skill files"→4; the `@MX:ANCHOR fan_in=5` comment (≈L46) →4; L38 "lists the 4 implementation skill files"→3. Also refresh `catalog_slim_audit_test.go` comments "25 (17 optional…)" (L47) and "40 (20 core…)" (L80) for accuracy — those asserts are dynamic (no test break), comment-only.
- Run `make build` (regenerates `embedded.go`, refreshes remaining hashes).
**Verify:** `go test ./internal/template/...` GREEN; `make build` idempotent (re-run produces no diff).

### M3 — Cross-reference cleanup + security replacement path
**Files (both trees unless noted):**
- `.claude/output-styles/moai/moai.md`: line 121 `… or /moai security` → reroute to `/moai review --security` / `Agent(general-purpose)` security scope.
- `CLAUDE.md`: remove the "Unified Skill: /moai design" section (§3) + drop brain/design/coverage/e2e/security from the §3 subcommand list.
- `.claude/rules/moai/workflow/spec-workflow.md`: remove dangling subcommand refs.
- `.claude/rules/moai/core/glm-web-tooling.md`: drop the `moai-domain-research/SKILL.md` cross-ref token (line 12).
- `.claude/skills/moai-domain-humanize/SKILL.md`: rewrite line 149 to remove the `moai-domain-copywriting` dependency (REQ-SCR-010) — note copywriting is retired, keep humanize self-contained.
**Verify:** repo-wide grep (scratch/spec/backup excluded, FROZEN design/constitution.md + zone-registry exempted) → 0 dangling references to the 5 subcommands / 7 skills.

### M4 — `make build` + dual-tree removed-basename absence confirmation
- Re-run `make build`; confirm `embedded.go` regenerated and no stale embedded references to removed artifacts.
- **Dual-tree clean (asymmetric by design — REQ-SCR-012):** verify the SEVEN removed skill basenames are absent from BOTH trees; verify `commands/moai/` + `workflows/` free of the 5 in both trees. Expected counts: **template = 28 skills, local = 30 skills** (28 + 2 user-owned `harness-moaiadk-best-practices` / `harness-moaiadk-patterns`). Do NOT assert "28 each" and do NOT run a raw full-tree `diff` expecting identity.
  - [HARD §24 carve-out] The two `harness-*` skills are user-owned (§24) and MUST be preserved — never deleted to "make the trees match". Verification command:
    ```bash
    for n in moai-domain-ideation moai-domain-research moai-domain-design-handoff \
             moai-domain-brand-design moai-domain-copywriting moai-workflow-design moai-workflow-gan-loop; do
      for t in ".claude/skills/$n" "internal/template/templates/.claude/skills/$n"; do
        [ -e "$t" ] && echo "RESIDUE: $t"   # per-tree independent check — catches a skill left in EXACTLY ONE tree
      done
    done   # expect no RESIDUE lines
    ls -d .claude/skills/harness-*/   # MUST still list the 2 user-owned harness skills
    ```
**Verify:** REQ-SCR-006 (seven basenames absent from both trees) + REQ-SCR-012 (harness-* preserved; 28 template / 30 local).

### M5 — docs-site cleanup (4-locale parity)
**Files:**
- Delete `docs-site/content/{en,ko,ja,zh}/quality-commands/moai-coverage.md` + `moai-e2e.md` (8 files).
- `docs-site/data/menu/main.yaml`: remove `coverage` (lines ~276-280) + `e2e` (~282-286) entries.
- Clean cross-links in `quality-commands/{_index.md,_meta.yaml,moai-review.md,moai-codemaps.md}` + ko `getting-started/introduction.md` + `core-concepts/what-is-moai-adk.md`.
**Verify:** `hugo --cleanDestinationDir` build succeeds; no menu/page for coverage/e2e in any locale; en/ko/ja/zh symmetric (§17). (`security` has no page; `brain`/`design` already gone.)

### M6 — Final verification gate
- `go test ./internal/template/...` GREEN (full).
- `go test ./...` GREEN (catch cascading failures per CLAUDE.local.md §6).
- `golangci-lint run` clean.
- Template-neutrality CI guard passes (REQ-SCR-009).
- Repo-wide 0-dangling-ref grep (in-scope surfaces).
- `make build` idempotent.

> **Sequencing note**: M1 and M2 may both touch SKILL.md/catalog adjacency — M1 first
> (router) then M2 (catalog+skills+tests+build) keeps the catalog edit and constant updates
> in one atomic milestone. M3-M5 are independent and may parallelize across distinct files.

## §G. Anti-Patterns
See design.md §F (AP-1..AP-6). Most load-bearing: keep M2 atomic (no RED commit between
dir-removal and constant-update); never edit FROZEN design/constitution.md.

## §H. Cross-References
- research.md — verified inventory, empirical data, CI-guard analysis, exclusivity grep.
- design.md — four scope/sequencing/replacement decisions.
- CLAUDE.local.md §2 (Template-First), §17 (docs-site 4-locale parity), §25 (template neutrality).
- Precedent: `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001`, `SPEC-V3R6-SEQ-THINKING-RETIRE-001`, `SPEC-V3R5-CORE-SLIM-001`.
