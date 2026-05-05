# SPEC-V3R2-WF-005 Task Breakdown

> Granular task decomposition of M1-M5 milestones from `plan.md` §2.
> Companion to `spec.md` v0.2.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 1B)     | Initial task breakdown — 23 tasks (T-WF005-01..23) across M1-M5         |

---

## Task ID Convention

- ID format: `T-WF005-NN`
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd`, `manager-docs`, `expert-backend` (Go test), `manager-git` (commit/PR boundary)
- Dependencies: explicit task ID list; tasks with no deps may run in parallel within their milestone
- DDD/TDD alignment: per `.moai/config/sections/quality.yaml` `development_mode: tdd`, M1 (RED) precedes M2-M5 (GREEN/REFACTOR)

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies only.

---

## M1: Test Scaffolding (RED phase)

Goal: Create the audit test that will fail until M3-M4 cleanup runs. Per `spec-workflow.md` TDD: write failing test first.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF005-01 | Create `internal/template/lang_boundary_audit_test.go` skeleton (package, imports, no test bodies yet) | expert-backend | `internal/template/lang_boundary_audit_test.go` (new file, ~15 LOC scaffold) | none | 1 file (create) | RED setup |
| T-WF005-02 | Implement `TestNoLangSkillDirectory` walking `.claude/skills/` in embedded FS, asserting no path matches `^moai-lang-[a-z-]+$` directory pattern. Sentinel: `LANG_AS_SKILL_FORBIDDEN`. | expert-backend | `internal/template/lang_boundary_audit_test.go` (test func, ~30 LOC) | T-WF005-01 | 1 file (extend) | RED — test must compile and PASS at baseline (no `moai-lang-*` directories exist today per research.md §3) |
| T-WF005-03 | Implement `TestRelatedSkillsNoLangReference` walking all `.claude/skills/**/SKILL.md` in embedded FS, parsing frontmatter, asserting `related-skills:` value contains no `moai-lang-` substring. Sentinel: `DEAD_LANG_SKILL_REFERENCE`. | expert-backend | `internal/template/lang_boundary_audit_test.go` (test func, ~45 LOC) | T-WF005-01 | 1 file (extend) | RED — must compile and FAIL initially (3 SKILL.md files have `moai-lang-*` per research.md §3) |
| T-WF005-04 | Implement `TestLanguageNeutrality` walking all skill/rule body content, scanning for primacy phrases (regex set per research.md §6.2), excluding fenced code blocks. Sentinel: `LANG_NEUTRALITY_VIOLATION`. | expert-backend | `internal/template/lang_boundary_audit_test.go` (test func, ~50 LOC) | T-WF005-01 | 1 file (extend) | RED — must compile and PASS at baseline (no obvious primacy violations per research.md inspection) |
| T-WF005-05 | Run `go test ./internal/template/ -run "TestNoLangSkillDirectory|TestRelatedSkillsNoLangReference|TestLanguageNeutrality"` and confirm RED state for `TestRelatedSkillsNoLangReference` (3 frontmatter violations); GREEN for the other 2 tests at baseline. | manager-tdd | n/a (verification only) | T-WF005-02, T-WF005-03, T-WF005-04 | 0 files | RED gate verification |

**M1 priority: P0** — blocks all subsequent milestones. M1 must complete before M2/M3/M4/M5 begin (TDD discipline).

---

## M2: "Language Guidance Lives in Rules" principle in skill-authoring.md (GREEN, part 1)

Goal: Publish the principle that establishes the canonical decision per REQ-WF005-001..004, REQ-WF005-009.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF005-06 | Append new section `## Language Guidance Lives in Rules, Not Skills` to `.claude/rules/moai/development/skill-authoring.md` (after "## Tool Permissions by Category"). Body per plan.md §2 M2 template. Include: principle declaration, canonical location pointer, paths-loading explanation, 17th-language gate, atomic reversal gate (REQ-WF005-012), `LANG_AS_SKILL_FORBIDDEN` CI key reference, cross-link to CLAUDE.local.md §15. | manager-docs | `.claude/rules/moai/development/skill-authoring.md` (append ~30 lines at end) | T-WF005-05 | 1 file (edit) | GREEN |
| T-WF005-07 | Mirror T-WF005-06 edit into `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` | manager-docs | template file | T-WF005-06 | 1 file (edit, parity) | Embedded-template parity |
| T-WF005-08 | Run `make build` to regenerate `internal/template/embedded.go`. Verify diff is exactly the principle section addition. | manager-docs | `internal/template/embedded.go` (regenerated) | T-WF005-07 | 1 file (regenerated) | Build verification |

**M2 priority: P0** — blocks M3/M4/M5 (the cleanup milestones reference the principle as the canonical justification).

---

## M3: Frontmatter `related-skills:` cleanup in 3 SKILL.md files (GREEN, part 2)

Goal: Make `TestRelatedSkillsNoLangReference` pass.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF005-09 | Edit `.claude/skills/moai-platform-auth/SKILL.md:20`: remove `moai-lang-typescript` token from `related-skills:` field. Other tokens (`moai-platform-supabase`, `moai-platform-vercel`, `moai-domain-backend`, `moai-expert-security`) preserved. | manager-docs | 1 SKILL.md frontmatter | T-WF005-08 | 1 file (edit) | GREEN |
| T-WF005-10 | Edit `.claude/skills/moai-framework-electron/SKILL.md:19`: remove `moai-lang-typescript` and `moai-lang-javascript` tokens from `related-skills:` field. `moai-domain-frontend` preserved. | manager-docs | 1 SKILL.md frontmatter | T-WF005-08 | 1 file (edit) | GREEN |
| T-WF005-11 | Edit `.claude/skills/moai-platform-chrome-extension/SKILL.md:19`: remove `moai-lang-typescript` and `moai-lang-javascript` tokens from `related-skills:` field. `moai-domain-frontend` preserved. | manager-docs | 1 SKILL.md frontmatter | T-WF005-08 | 1 file (edit) | GREEN |
| T-WF005-12 | Mirror T-WF005-09..11 edits into `internal/template/templates/.claude/skills/{moai-platform-auth,moai-framework-electron,moai-platform-chrome-extension}/SKILL.md` (3 files) | manager-docs | 3 template files | T-WF005-09, 10, 11 | 3 files (edit, parity) | Embedded-template parity |
| T-WF005-13 | Run `make build` + `go test ./internal/template/ -run TestRelatedSkillsNoLangReference` and confirm PASS (3 of 3 frontmatter subtests GREEN). | manager-tdd | n/a (verification only) | T-WF005-12 | 1 file (regenerated) | GREEN gate part 1 |

**M3 priority: P0** — required for `TestRelatedSkillsNoLangReference` to turn GREEN. Independent of M4 (body cleanup) which uses a different audit path (manual grep).

T-WF005-09, T-WF005-10, T-WF005-11 may execute in parallel — they touch independent files. T-WF005-12 must wait for all 3 source-of-truth edits.

---

## M4: Body cleanup of dead `moai-lang-*` references (GREEN, part 3)

Goal: Substitute or remove all body-level `moai-lang-*` mentions per REQ-WF005-005, REQ-WF005-008.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF005-14 | Replace `moai-lang-<name>` references in `.claude/skills/moai/workflows/run.md:338-352,979` with rule-path references to `.claude/rules/moai/languages/<name>.md`. Mapping target preserved (file marker → language); only the right-hand-side pointer changes. | manager-docs | 1 skill body, ~17 line edits | T-WF005-13 | 1 file (edit) | GREEN |
| T-WF005-15 | Replace `moai-lang-*` references in `.claude/skills/moai-foundation-core/modules/agents-reference.md:268,278-282,292` (catalog table). Substitute with rule-path references or remove rows that no longer apply. | manager-docs | 1 file, ~7 line edits | T-WF005-13 | 1 file (edit) | GREEN |
| T-WF005-16 | Replace body-level `moai-lang-*` references in 8 SKILL.md and reference files: `moai-framework-electron/SKILL.md:228,230`; `moai-platform-chrome-extension/SKILL.md:278,279`; `moai-platform-auth/SKILL.md:225`; `moai-platform-deployment/SKILL.md:398,399`; `moai-workflow-loop/SKILL.md:148,149`; `moai-domain-frontend/SKILL.md:119`; `moai-foundation-cc/reference/skill-examples.md:237,427`; `moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md:201,392`; `moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:30,439,936`; `moai-foundation-cc/reference/skill-formatting-guide.md:113`; `moai-foundation-cc/references/examples.md:175`. | manager-docs | ~12 files, ~17 line edits | T-WF005-13 | ~12 files (edit) | GREEN |
| T-WF005-17 | Replace `moai-lang-go` example in `.claude/rules/moai/development/agent-authoring.md:165` with `moai-domain-backend` (or another existing skill ID). Preserve YAML array syntax demonstration. | manager-docs | 1 rule body, 1 line edit | T-WF005-13 | 1 file (edit) | GREEN |
| T-WF005-18 | Replace cross-language `moai-lang-*` references inside language rules: `.claude/rules/moai/languages/scala.md:131` (`moai-lang-java`), `.claude/rules/moai/languages/flutter.md:94,95` (`moai-lang-swift`, `moai-lang-kotlin`), `.claude/rules/moai/languages/cpp.md:100` (`moai-lang-rust`). Substitute with `.claude/rules/moai/languages/<name>.md` references. | manager-docs | 3 rule bodies, ~4 line edits | T-WF005-13 | 3 files (edit) | GREEN |
| T-WF005-19 | Mirror T-WF005-14..18 edits into `internal/template/templates/.claude/...` for all affected files (~17 files). | manager-docs | ~17 template files | T-WF005-14, 15, 16, 17, 18 | ~17 files (edit, parity) | Embedded-template parity |
| T-WF005-20 | Run `make build` + manual `grep -rn "moai-lang-" .claude/skills .claude/rules` and confirm zero non-test matches. | manager-tdd | n/a (verification only) | T-WF005-19 | 1 file (regenerated) | GREEN gate part 2 |

**M4 priority: P0** — required for `spec.md` §2.1 In Scope contract.

[HARD] T-WF005-14 must NOT modify the language detection mapping logic itself; only the right-hand-side pointer (skill ID → rule path). Phase 0.9's behavioral contract is preserved.

[HARD] T-WF005-18 must NOT modify the language rule body content beyond the specific cross-language reference lines. The 16 rules' Tooling/Conventions/Best Practices sections are preserved per spec.md §1.2.

T-WF005-14 through T-WF005-18 may execute in parallel — they touch independent files (with the exception of `flutter.md` which is also touched in M5a/M5c, requiring sequential ordering within that file).

---

## M5: Other dead-skill-ID cleanup + CHANGELOG + MX tags (GREEN, part 4 + REFACTOR + Trackable)

Goal: TRUST 5 Trackable + complete dead-skill-ID cleanup per REQ-WF005-015.

### M5a: `moai-essentials-debug` substitution (9 files)

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF005-21 | Substitute `moai-essentials-debug` with "delegate to expert-debug agent for AI-powered debugging" (or context-appropriate equivalent) in 9 files: `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:621`; `.claude/rules/moai/languages/{r,typescript,flutter,cpp,csharp,elixir,javascript,ruby}.md` (8 files at lines per research.md §4.1). Mirror to template tree. | manager-docs | 9 files, 9 line edits + 9 template mirrors | T-WF005-20 | 18 files (edit + parity) | GREEN/REFACTOR |

### M5b: `moai-quality-testing` substitution (1 file)

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF005-22 | Substitute `moai-quality-testing` with "moai-foundation-quality + moai-ref-testing-pyramid" in `.claude/rules/moai/languages/kotlin.md:109`. Mirror to template tree. | manager-docs | 1 file, 1 line edit + 1 template mirror | T-WF005-20 | 2 files (edit + parity) | GREEN/REFACTOR |

### M5c: `moai-quality-security` substitution (3 files)

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF005-23 | Substitute `moai-quality-security` with "moai-foundation-quality + moai-ref-owasp-checklist" in 3 files: `.claude/skills/moai-domain-backend/SKILL.md:110`; `.claude/skills/moai-workflow-project/references/overview.md:120`; `.claude/rules/moai/languages/flutter.md:97`. Mirror to template tree. | manager-docs | 3 files, 3 line edits + 3 template mirrors | T-WF005-20 | 6 files (edit + parity) | GREEN/REFACTOR |

### M5d: `moai-infra-docker` removal (2 files) + final wrap-up

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF005-24 | Remove `moai-infra-docker` line entirely (no substitute) in 2 files: `.claude/rules/moai/languages/kotlin.md:110`; `.claude/rules/moai/languages/java.md:121`. Mirror to template tree. Verify: `grep -rn "moai-infra-docker\|moai-essentials-debug\|moai-quality-testing\|moai-quality-security" .claude/skills .claude/rules` returns zero matches. | manager-docs | 2 files, 2 line edits + 2 template mirrors | T-WF005-21, T-WF005-22, T-WF005-23 | 4 files (edit + parity) | GREEN/REFACTOR |
| T-WF005-25 | Add CHANGELOG entry under `## [Unreleased]` per plan.md §2 M5e. Insert MX tags (2 ANCHOR + 2 NOTE + 2 WARN = 6 total) per plan.md §6 in 5 distinct file locations. Run `make build` + final `go test ./...` from repo root. Verify ALL 3 audit tests GREEN + 0 cascading failures. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete`. | manager-docs | `CHANGELOG.md` + 5 MX-tagged files + `progress.md` + `internal/template/embedded.go` (regenerated) | T-WF005-24 | ~8 files | REFACTOR / Trackable |

**M5 priority: P1** — required for full spec.md §2.1 In Scope coverage and TRUST 5 Trackable.

[HARD] T-WF005-25's final verification step is the gating point. No tests may regress (per CLAUDE.local.md §6 HARD).

---

## Aggregate Statistics

- **Total tasks**: 25 (T-WF005-01 through T-WF005-25)
- **Total milestones**: 5 (M1: 5 tasks, M2: 3 tasks, M3: 5 tasks, M4: 7 tasks, M5: 5 tasks)
- **Files created**: 1 (`internal/template/lang_boundary_audit_test.go`)
- **Files modified (source-of-truth)**: ~28 unique files
  - 1 `skill-authoring.md` (principle section)
  - 3 SKILL.md files (frontmatter cleanup, M3)
  - ~12 skills + 1 dev rule (body cleanup, M4)
  - 3 language rules (cross-language refs, M4)
  - 9 files (`moai-essentials-debug`, M5a)
  - 1 file (`moai-quality-testing`, M5b)
  - 3 files (`moai-quality-security`, M5c)
  - 2 files (`moai-infra-docker` removal, M5d)
  - 1 `CHANGELOG.md`
- **Files modified (embedded-template parity)**: ~28 mirror files
- **MX tag insertions**: 6 across 5 distinct file locations (per plan.md §6)
- **Owner-role distribution**:
  - `expert-backend`: 4 tasks (T-WF005-01..04, all Go test work)
  - `manager-tdd`: 4 tasks (T-WF005-05, 13, 20, partial 25 — TDD gate verification)
  - `manager-docs`: 17 tasks (all content authoring + template parity + final sync)

---

## Owner-Role Rationale

- **Go test work** (`expert-backend`): the audit test is Go code; needs Go expertise (regex, fs.WalkDir, embedded FS, YAML frontmatter parsing).
- **TDD gate verification** (`manager-tdd`): the project's `quality.yaml` declares `development_mode: tdd`, so manager-tdd is the methodology owner. Each gate (RED, GREEN parts 1/2/3, final) is a manager-tdd checkpoint.
- **Content authoring** (`manager-docs`): all skill/rule/CHANGELOG edits are documentation per `.claude/rules/moai/development/coding-standards.md` (skills are config-as-code documents). manager-docs is the owner per `CLAUDE.md` §4 Manager Agents catalog.

---

## Parallel Execution Opportunities

These task groups have no inter-dependencies and may run in parallel within their milestone:

- **M1 parallel**: T-WF005-02, T-WF005-03, T-WF005-04 (extend the same file; cannot truly parallelize without merge conflict — execute sequentially in same edit pass).
- **M3 parallel**: T-WF005-09, T-WF005-10, T-WF005-11 (touch 3 independent files — true parallelism possible, but per `CLAUDE.md` §14 Multi-File Decomposition HARD rule with 3+ files, sequencing recommended).
- **M4 parallel**: T-WF005-14, T-WF005-15, T-WF005-16, T-WF005-17, T-WF005-18 (touch independent files except `flutter.md` cross-references — true parallelism possible with care).
- **M5a-d**: tasks T-WF005-21 / T-WF005-22 / T-WF005-23 are independent and may run in parallel; T-WF005-24 sequences after them; T-WF005-25 sequences last.

Per `CLAUDE.md` §1 HARD rule "Multi-File Decomposition: Split work when modifying 3+ files," every milestone with multi-file scope already encodes the per-file or per-category decomposition.

---

## Cross-Reference Map

Each task references which REQ(s) and which AC(s) it advances toward DoD:

| Task ID | REQ coverage | AC coverage |
|---------|--------------|-------------|
| T-WF005-01 | (scaffold) | (scaffold) |
| T-WF005-02 | REQ-WF005-002, 007, 009 | AC-WF005-01, 05, 09 |
| T-WF005-03 | REQ-WF005-005, 013 | AC-WF005-06, 12 |
| T-WF005-04 | REQ-WF005-014 | AC-WF005-11 |
| T-WF005-05 | (gate) | (gate) |
| T-WF005-06 | REQ-WF005-001, 002, 003, 004, 009, 011, 012 | AC-WF005-02, 03, 09 |
| T-WF005-07 | (parity) | (parity) |
| T-WF005-08 | (build) | (build) |
| T-WF005-09 | REQ-WF005-005, 008 | AC-WF005-06 |
| T-WF005-10 | REQ-WF005-005, 008 | AC-WF005-06 |
| T-WF005-11 | REQ-WF005-005, 008 | AC-WF005-06 |
| T-WF005-12 | (parity) | (parity) |
| T-WF005-13 | (gate) | AC-WF005-06 GREEN |
| T-WF005-14 | REQ-WF005-008, 010 | AC-WF005-12 |
| T-WF005-15 | REQ-WF005-005, 008 | AC-WF005-12 |
| T-WF005-16 | REQ-WF005-005, 008 | AC-WF005-12 |
| T-WF005-17 | REQ-WF005-005 | AC-WF005-12 |
| T-WF005-18 | REQ-WF005-005, 008 | AC-WF005-10, 12 |
| T-WF005-19 | (parity) | (parity) |
| T-WF005-20 | (gate) | AC-WF005-12 final |
| T-WF005-21 | REQ-WF005-015 | AC-WF005-08 |
| T-WF005-22 | REQ-WF005-011, 015 | AC-WF005-07 |
| T-WF005-23 | REQ-WF005-011, 015 | AC-WF005-07 |
| T-WF005-24 | REQ-WF005-015 | (final cleanup) |
| T-WF005-25 | (Trackable / MX) | (DoD steps 9-11) |

REQ coverage summary: 15 REQs, all referenced by at least one task (verified by transitive lookup against `spec.md` §5).

AC coverage summary: 12 ACs, all advanced toward DoD by tasks. AC-WF005-04 (Python project auto-loads `python.md`) is implicit — verified by the existing rule loader behavior plus research.md §2 baseline confirmation, and does not require a new task.

REQ Categories balance:
- Ubiquitous (REQ-001..005): T-WF005-02, 06, 09-11, 14-18, 21-24
- Event-Driven (REQ-006..008): T-WF005-14-18 (auto-load behavior + cleanup commit + CI rejection)
- State-Driven (REQ-009..010): T-WF005-02, 06, 18
- Optional (REQ-011..012): T-WF005-06, 22-23
- Complex (REQ-013..015): T-WF005-03, 04, 21-24

Every category has at least one task. Task-to-REQ mapping covers all 15 REQs.

---

## Final Verification Pass (subset of T-WF005-25)

Before marking SPEC implementation complete, execute these checks in order:

1. **Audit tests green**: `go test ./internal/template/ -run "TestNoLangSkillDirectory|TestRelatedSkillsNoLangReference|TestLanguageNeutrality" -v` — 3 funcs, all PASS.
2. **Full repo green**: `go test ./...` — 0 failures (per `CLAUDE.local.md` §6 HARD).
3. **Lint clean**: `golangci-lint run` — 0 errors, ≤10 warnings (per `quality.yaml` `lsp_quality_gates.sync`).
4. **Vet clean**: `go vet ./...` — 0 issues.
5. **Embedded parity**: diff `internal/template/embedded.go` after `make build` shows only the expected file content changes (no spurious diffs).
6. **Manual sweep**: `grep -rn "moai-lang-\|moai-essentials-debug\|moai-quality-testing\|moai-quality-security\|moai-infra-docker" .claude/skills .claude/rules` returns zero matches (excluding the new audit test which legitimately mentions sentinel strings inside Go code).
7. **DoD checklist**: all 11 items in `acceptance.md` §"Definition of Done" checked.

If any step fails, return to the appropriate milestone for remediation. Do NOT advance to `/moai sync` until all 7 checks pass.

---

End of tasks.md.
