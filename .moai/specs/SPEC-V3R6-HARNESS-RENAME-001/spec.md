---
id: SPEC-V3R6-HARNESS-RENAME-001
version: "0.2.0"
status: implemented
created_at: 2026-05-22
updated_at: 2026-05-22
author: manager-spec
priority: Medium
labels: [harness, rename, foundation, wave-2, v3.0.0]
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-HARNESS-LEARNER-FIX-001, SPEC-V3R6-CATALOG-SSOT-001, SPEC-V3R6-AGENT-FOLDER-SPLIT-001, SPEC-V3R6-META-HARNESS-PATH-001]
---

# SPEC-V3R6-HARNESS-RENAME-001: Harness Folder + Skill Prefix Rename

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — Wave 2 첫 SPEC, my-harness/ → harness/ + my-harness-* → moai-harness-* 정렬 (Tier S, 64+34=98 refs estimated) |
| 0.2.0 | 2026-05-22 | manager-develop | Implementation complete (orchestrator-direct). Scope reduced per user decision 2026-05-22: internal/ Go 31 files (~100 refs) descoped due to cross-SPEC tension with REQ-PH-009 user-area preservation system. In-scope changes: 5 dir renames + 8 frontmatter (Option A: moai-harness-X-specialist) + 5 cross-ref skill body edits + 8 template mirror files (4 new agents + 4 new skills) + 5 template mirror overwrites + 8 new catalog entries + 3 test count adjustments. AC matrix: 6/7 strict PASS, 1 semantic PASS (AC-HRN-004 outcome count 18 vs baseline 19 due to TestManifestHashFormat FAIL→PASS removing `--- FAIL:` line, no test deletion). |

## 1. Goal

`.claude/agents/my-harness/` 디렉토리와 `.claude/skills/my-harness-*` skill prefix를 v3.0.0 redesign blueprint § Wave 2 결정에 맞춰 일관된 이름 체계로 정렬한다. `moai-harness-learner` skill이 이미 따르고 있는 `moai-*` prefix 규약을 동일 카테고리의 나머지 4개 skill과 4개 agent에도 적용함으로써 catalog SSoT 일관성을 회복하고 v3.0.0 Wave 3 (slim-down) 진입 전 폴더 구조를 확정한다.

## 2. Why

- **Inconsistency**: `.claude/skills/moai-harness-learner/`는 이미 `moai-*` prefix를 사용하나, 같은 카테고리의 `my-harness-{cli-template,hook-ci,quality,workflow}/`는 `my-*` prefix 잔존 → catalog SSoT 검증 시 혼란 + 신규 사용자 학습 비용 증가
- **Folder name convention**: v3.0.0 blueprint § Wave 1 결정 "`agents/{core,expert,meta,harness}/`" — Wave 2 Agent-Folder-Split SPEC이 진입하기 전에 `harness/` 이름을 미리 확정 (현재 `my-harness/`는 v2 잔재)
- **Template-First Rule 위반 가시화**: `.claude/agents/my-harness/` 4 files + `.claude/skills/my-harness-*/` 4 dirs는 local-only로 `internal/template/templates/`에 미존재 → 본 SPEC을 통해 Template-First mirror도 동시 수립
- **Wave 2 dependency**: Wave 2의 후속 SPEC (AGENT-FOLDER-SPLIT-001, META-HARNESS-PATH-001) 은 안정된 `harness/` 폴더 이름을 전제로 함

## 3. Requirements (EARS)

### Ubiquitous

- REQ-HRN-001: The system shall use `harness/` (not `my-harness/`) as the agent folder name under `.claude/agents/`.
- REQ-HRN-002: The system shall use `moai-harness-*` prefix (not `my-harness-*`) for all harness-category skills under `.claude/skills/`.
- REQ-HRN-003: The system shall maintain a Template-First mirror of all harness agents and skills under `internal/template/templates/.claude/`.

### State-Driven

- REQ-HRN-004: While the rename is in progress, the system shall preserve every existing reference target by updating all 98+ references (64 my-harness + 34 moai-harness verification crosscheck) in a single atomic milestone.
- REQ-HRN-005: While preserving backward-compat for in-flight SPECs (e.g., SPEC-V3R6-HARNESS-LEARNER-FIX-001 PR #1037 merged), the system shall not break references in completed/implemented SPECs by using one-way rename (no aliases needed since completed SPECs are read-only artifacts).

### Event-Driven

- REQ-HRN-006: When `go build ./...` runs on darwin and `GOOS=windows GOARCH=amd64 go build ./...`, the build shall succeed with exit code 0 (no broken string references in Go code).
- REQ-HRN-007: When test files referencing `my-harness` or `moai-harness` literals execute (`go test ./internal/cli/... ./internal/template/... ./internal/design/...`), the test count shall not decrease.

### Unwanted Behavior

- REQ-HRN-008: The system shall NOT rename `.moai/harness/` (runtime config directory containing README.md/main.md/seeds/usage-log.jsonl) — that path is a stable runtime contract.
- REQ-HRN-009: The system shall NOT introduce backward-compat symlinks or alias files for renamed paths — completed SPECs are read-only historical artifacts that do not require live references.

## 4. Scope

### In Scope

1. `.claude/agents/my-harness/` directory rename to `.claude/agents/harness/` (4 files: cli-template-specialist, hook-ci-specialist, quality-specialist, workflow-specialist).
2. `.claude/skills/my-harness-{cli-template,hook-ci,quality,workflow}/` rename to `.claude/skills/moai-harness-{cli-template,hook-ci,quality,workflow}/` (4 directories).
3. Agent frontmatter `name:` field rename inside each .md file (e.g., `my-harness-workflow-specialist` → user-chosen prefix per Plan §1.3 decision matrix).
4. Skill `skills:` array references in agent definitions (e.g., `skills: [my-harness-workflow]` → `skills: [moai-harness-workflow]`).
5. Cross-file references in `.claude/rules/`, `.claude/skills/`, `.moai/research/`, `CLAUDE.md`, `CLAUDE.local.md` — body text + code-block references.
6. Template-First mirror creation under `internal/template/templates/.claude/agents/harness/` (4 files new) and `internal/template/templates/.claude/skills/moai-harness-{cli-template,hook-ci,quality,workflow}/` (4 dirs new).
7. Go test files containing string literals `my-harness` or `moai-harness` (10 files identified — update string literals only; do not delete tests).
8. Go source files in `internal/cli/`, `internal/harness/`, `internal/template/`, `internal/design/pipeline/` — string-literal-only changes (no symbol or API surface changes).
9. `internal/template/catalog.yaml` + `internal/template/catalog_doc.md` SSoT updates (skill entries + hash regeneration via `gen-catalog-hashes.go --all`).

### Out of Scope

1. `.moai/harness/` runtime config directory (REQ-HRN-008 prohibits).
2. Active SPEC frontmatter rewrites — historical artifact integrity (e.g., SPEC-V3R6-HARNESS-LEARNER-FIX-001 `id:` field stays as-is even if body content is updated).
3. Backward-compat aliasing or symlink creation (REQ-HRN-009 prohibits).
4. `internal/template/templates/.claude/agents/moai/*` mass rename (out of scope — that belongs to Wave 2 SPEC-V3R6-AGENT-FOLDER-SPLIT-001).
5. Skill content body restructure (out of scope — that belongs to Wave 3 SPEC-V3R6-SKILL-SLIM-001).
6. docs-site/ updates (verified 0 refs — no impact).
7. Symbol-level Go API renames (Public functions `harness.CreateLearning(...)` etc. — string literals only).

## 5. Exclusions

- `.moai/harness/` directory and its contents are explicitly excluded (REQ-HRN-008).
- Backward-compatibility shims, deprecation warnings, or alias mappings are explicitly excluded (REQ-HRN-009).
- Multi-locale docs-site/ propagation is excluded (0 refs detected — no propagation required).
- Live SPEC reference rewrites for already-completed/merged SPECs are excluded (historical artifact integrity).
- AGENT-FOLDER-SPLIT (moai/ → core/expert/meta/) is excluded — separate Wave 2 SPEC handles it.

## 6. Non-Goals

- This SPEC does not refactor any harness business logic in `internal/harness/`.
- This SPEC does not add new harness functionality.
- This SPEC does not change the runtime contract of `moai cg`, `moai cc`, `moai glm` commands.
- This SPEC does not modify the agent invocation pattern at runtime — only identifier strings change.

## 7. Acceptance Criteria

See [acceptance.md](./acceptance.md). 7 binary ACs covering rename completeness, Template-First mirror, build sanity, test count preservation, catalog SSoT, agent-name self-reference consistency, and exclusion compliance.

## 8. Risks

| Risk ID | Description | Severity | Mitigation |
|---------|-------------|----------|------------|
| R-HRN-001 | Hidden `my-harness` literal in a rule/skill body breaks agent lookup at runtime | High | grep-driven exhaustive search + AC-HRN-001 binary verification |
| R-HRN-002 | Agent frontmatter `name:` field naming convention ambiguity (3 options A/B/C in plan §1.3) causes inconsistency | Medium | Plan §1.3 decision matrix; user confirms at run-phase via AskUserQuestion or default to Option A |
| R-HRN-003 | Template-First mirror creation skipped by accident (local edits only) | High | AC-HRN-002 binary check + Section D constraint in delegation prompt |
| R-HRN-004 | `internal/template/catalog.yaml` hash mismatch after skill rename (SHA256 of moved file content) | Medium | `gen-catalog-hashes.go --all` invocation per SPEC-V3R6-CATALOG-SSOT-001 self-healing gate |
| R-HRN-005 | Go test file rename (e.g., `update_preserve_my_harness_test.go`) creates filename-content drift | Low | Update string literals inside; defer filename rename if drift detected |

## 9. Dependencies

- SPEC-V3R6-CATALOG-SSOT-001 (M1 cycle, status=implemented v0.2.0) — provides `gen-catalog-hashes.go --all` self-healing gate that this SPEC relies on for catalog hash regen after skill rename.
- SPEC-V3R6-HARNESS-LEARNER-FIX-001 (PR #1037 merged) — confirms `moai-harness-learner` prefix is accepted; this SPEC extends the same pattern to remaining 4 skills.

## 10. Success Metrics

- 0 occurrences of `my-harness` literal in `.claude/`, `.moai/research/`, `CLAUDE.md`, `CLAUDE.local.md`, `internal/` (except historical SPECs that are out-of-scope per Exclusions).
- 64+34 = 98+ references successfully renamed or verified.
- Cross-platform build PASS (darwin + windows amd64).
- Test count baseline preserved (no test deletions, only string literal updates).
- Template-First mirror present: 4 agents + 4 skills under `internal/template/templates/.claude/`.
- `TestManifestHashFormat` CI gate PASS (catalog.yaml hashes regenerated correctly).
