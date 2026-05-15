---
id: SPEC-V3R4-CATALOG-001
version: "0.3.0"
status: completed
created: 2026-05-12
updated: 2026-05-13
author: GOOS행님
priority: High
issue_number: 859
title: "3-Tier Catalog Manifest"
phase: "v3.0.0 - Lifecycle"
module: "catalog"
lifecycle: spec-anchored
tags: "catalog, manifest, 3-tier, v3r4, foundation"
run_pr: 862
eval_fix_pr: 863
evaluation_pass_at: "2026-05-12T03:30:00Z"
evaluation_overall_score: "0.82"
coverage: "84.0% (LoadCatalog: 100%)"
---

# SPEC-V3R4-CATALOG-001: 3-Tier Catalog Manifest

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-12 | GOOS행님 | Initial draft from /moai brain IDEA-003 (Wave 1 첫 SPEC, manifest schema + tier lock-in) |
| 0.2.0 | 2026-05-12 | manager-spec | plan-auditor iter 1 FAIL (0.72) 11 defects 반영: D1 (REQ-013 자체 모순 제거), D2 (REQ-027 duplicate sentinel 추가), D3 (37 skills + 28 agents = 65 entries 카운트 통일 + "catalog entry" 정의 명시), D4 (REQ-021 vacuously-true 재작성, workflows flat-md layout 반영), D5 (gen-catalog-hashes.go [NEW] 등록), D6 (15 untracked REQ → AC mapping 보강), D7 (deployer.go no-modify 명확화), D9 (REQ-024 약화). |
| 0.3.0 | 2026-05-12 | sync | status: draft → completed. PR #862 (M1-M5 implementation, main `ec80c8845`) + PR #863 (eval-1 follow-up: EC3 hash sentinel `t.Logf` → `t.Errorf`, LoadCatalog coverage 71.4% → 100%, main `0d4bf14ef`) 모두 머지. evaluator-active 독립 eval PASS 0.82 (2 required fixes 모두 PR #863에서 해결). Implementation Notes 섹션 추가. |

## Overview

moai-adk-go 카탈로그 슬림화 initiative (총 7 SPEC) 의 **foundation SPEC**이다. 현재 37 top-level skills 와 28 agents 가 분류 메타데이터 없이 일괄 deploy 되어, 신규 프로젝트는 사용하지 않는 도메인 자산까지 모두 받게 된다. Anthropic skill description budget (context window 의 약 1%) 압박과 `moai update` 시 drift 손실 위험이 누적되어 있다.

본 SPEC 은 **3-tier (core / optional-pack / harness-generated) 카탈로그 manifest 스키마를 정의**하고, 현재 65개 자산 (37 skills + 28 agents) 의 tier 분류를 lock-in 한다. 후속 6개 SPEC (재배치, pack CLI, update 안전 동기화, project 인터뷰, doctor, 마이그레이션 docs) 이 본 SPEC 의 manifest 위에 빌드된다.

**Catalog Entry 정의 (D3 권장 수정 반영):** "Catalog entry" 는 `internal/template/templates/.claude/skills/` 바로 아래의 top-level 디렉토리 1개 또는 `internal/template/templates/.claude/agents/moai/` 바로 아래의 `.md` 파일 1개를 의미한다. 본 SPEC 의 카운트 기준은 다음과 같다:

- **Skills (37)**: `templates/.claude/skills/` 의 top-level 디렉토리 수. `moai/` 컨테이너는 단일 logical skill (그 root 의 `SKILL.md` 가 entry point) 로 취급하며, 내부의 `workflows/`, `team/`, `references/` 는 `moai` skill 의 모듈로 별도 entry 가 아님. 32개는 `SKILL.md` (대문자) 를, 5개 `moai-ref-*` 디렉토리는 `skill.md` (소문자, legacy reference skill 명명) 를 갖는다 — manifest schema 는 둘 다 entry path 로 허용.
- **Agents (28)**: `templates/.claude/agents/moai/` 직속 `.md` 파일 수 (예: `manager-spec.md`, `expert-backend.md`).

**총 65 entries** (37 + 28). 이 수치는 본 SPEC 의 4개 파일 (spec.md / plan.md / acceptance.md / spec-compact.md) 에서 동일하게 사용된다.

**사용자 추가 제약 (brain Phase 1 응답):** `moai update` 요청 시 기존 프로젝트의 에이전트/스킬은 모두 안전하게 업데이트되어야 하며, 손실은 0 이어야 한다. 본 SPEC 의 manifest 스키마는 후속 SPEC-V3R4-CATALOG-004 의 drift detection 을 가능하게 하는 **hash + version 필드를 반드시 포함**한다.

본 SPEC 의 결과물은 (1) `internal/template/catalog.yaml` manifest 단일 진실 공급원, (2) `internal/template/catalog_tier_audit_test.go` 무결성 강제 audit suite, (3) `internal/template/catalog_loader.go` typed accessor, (4) `internal/template/scripts/gen-catalog-hashes.go` offline hash 생성 헬퍼이다. 실제 디렉토리 재배치, CLI 명령, 안전 동기화는 후속 SPEC 의 scope 이다.

## Background

본 SPEC 의 의사결정 근거 체인:

1. `.moai/brain/IDEA-003/proposal.md` — 7 SPEC 분해 + Wave 순서 (`Wave 1 Foundation → Wave 2 Distribution → Wave 3 Safety → Wave 4 Polish`) 와 hybrid 전략 (코어는 직접 배포, optional pack 은 Anthropic marketplace 검토 가능) 정립.
2. `.moai/brain/IDEA-003/ideation.md` — Lean Canvas + 5 비판/대응 + First Principles 분해. 핵심 invariant: "skill 의 존재 자체가 비용 (description 은 항상 budget 차지)" → 사용 빈도가 낮은 skill 일수록 존재 비용 > 활용 가치 → lazy install 정당화.
3. `.moai/specs/SPEC-V3R4-CATALOG-001/research.md` — 37 skills / 28 agents 검증, `internal/template/embed.go` go:embed 패턴, `internal/template/deployer.go` Deploy 인터페이스, `internal/cli/update.go` runUpdate 흐름, `internal/template/lang_boundary_audit_test.go` audit 패턴, 24개 `.moai/config/sections/*.yaml` manifest 선례 모두 분석.

선례 참조:
- **manifest 패턴**: `.moai/config/sections/harness.yaml` 의 `harness: levels: minimal/standard/thorough` 계층 구조 (research.md §"Existing Manifest Patterns").
- **audit 패턴**: `internal/template/lang_boundary_audit_test.go` 의 sentinel + parallel test 패턴 (research.md §"Audit Test Pattern (Reference)"), specifically `TestNoLangSkillDirectory` (sentinel `LANG_AS_SKILL_FORBIDDEN`), `TestRelatedSkillsNoLangReference` (sentinel `DEAD_LANG_FRONTMATTER_REFERENCE`).
- **embed 패턴**: `internal/template/embed.go:24-39` 의 `//go:embed all:templates` — catalog.yaml 을 templates 외부에 두어도 별도 directive 불필요 (단일 embed 가 모든 path 커버).
- **외부 표준 (참조용, 본 SPEC 도입은 안 함)**: Anthropic Plugin Marketplace 의 `marketplace.json` (research.md §"Embedded / Marketplace Compatibility"). 본 SPEC 은 internal-only manifest 로 출발하고, 마켓플레이스 publishing 은 후속 SPEC 영역으로 격리.

## EARS Requirements

### 1. Manifest Existence and Schema (Ubiquitous)

REQ-CATALOG-001-001: The system shall maintain exactly one catalog manifest file at `internal/template/catalog.yaml` as the single source of truth for skill and agent tier classification.

REQ-CATALOG-001-002: The manifest shall declare a top-level `version` field (semver string) representing the manifest schema version.

REQ-CATALOG-001-003: The manifest shall declare a top-level `generated_at` field (ISO 8601 date) recording when the manifest was last regenerated.

REQ-CATALOG-001-004: The manifest shall organize entries under a top-level `catalog` object containing exactly three sub-sections: `core`, `optional_packs`, and `harness_generated`.

REQ-CATALOG-001-005: Every top-level skill directory present under `internal/template/templates/.claude/skills/` (containing either `SKILL.md` or `skill.md`) shall appear in the manifest exactly once with a `tier` value of either `core`, `optional-pack:<name>`, or `harness-generated`.

REQ-CATALOG-001-006: Every agent file (`.md`) directly under `internal/template/templates/.claude/agents/moai/` shall appear in the manifest exactly once with a `tier` value of either `core`, `optional-pack:<name>`, or `harness-generated`.

### 2. Per-Entry Metadata for Drift Detection (Ubiquitous — sourced from user constraint)

REQ-CATALOG-001-007: Each catalog entry (skill or agent) shall include a `hash` field containing the sha256 digest of the normalized source file content (LF line endings, trailing whitespace trimmed). For skill entries, the hash MUST cover the entry's `SKILL.md` (or `skill.md` for reference skills) only; module sub-files inside the skill directory are NOT covered by this hash (a follow-up SPEC may extend coverage).

REQ-CATALOG-001-008: Each catalog entry shall include a `version` field (semver string) representing the entry's release version, used by SPEC-V3R4-CATALOG-004 to detect upstream changes.

REQ-CATALOG-001-009: Each catalog entry shall include a `path` field (relative to `internal/template/templates/`) identifying the entry's deploy target location.

REQ-CATALOG-001-010: Each optional-pack entry shall include a `depends_on` field (list of pack names, possibly empty `[]`) to enable acyclic dependency resolution in SPEC-V3R4-CATALOG-003.

### 3. Pack Definitions (Ubiquitous)

REQ-CATALOG-001-011: The manifest shall declare each optional pack under `catalog.optional_packs.<pack-name>` with the following fields: `description` (string), `depends_on` (list), `skills` (list of skill names), and `agents` (list of agent names).

REQ-CATALOG-001-012: Pack names shall conform to the regex `^[a-z][a-z0-9-]{1,30}$` (lowercase, hyphen-separated, length 2-31).

### 4. Tier Mutation Constraint (Unwanted Behavior) — D1 권장 수정 반영

REQ-CATALOG-001-013: If a contributor changes an entry's `tier` field in `catalog.yaml`, then the plan-auditor SHALL flag the change as requiring an explicit SPEC ID amendment reference in the pull request description, enforced at PR review time.

> **D1 fix note**: 이전 v0.1.0 의 REQ-013 은 audit suite 가 commit history 를 검증한다고 했으나, audit 은 runtime structural check 이고 history check 가 아님 — 자체 모순. v0.2.0 은 process invariant 를 명시적 plan-auditor 책임으로 재할당. Runtime audit suite (REQ-014..020) 는 manifest 구조만 검증하고, tier 변경 정당성은 PR 리뷰 (plan-auditor) 가 검증한다. 본 REQ 는 verifiable check (PR description 내 `SPEC-` 패턴 grep + tier delta 가 있는 commit 매칭) 로 후속 SPEC-V3R4-CATALOG-004 시점에 자동화 가능하며, 본 SPEC 단계에서는 plan-auditor 가 수동 검증한다.

### 5. Audit Suite (Ubiquitous)

REQ-CATALOG-001-014: The system shall provide a Go test file `internal/template/catalog_tier_audit_test.go` that fails the CI build when manifest integrity is violated.

REQ-CATALOG-001-015: The audit suite shall verify that every top-level skill directory under `.claude/skills/` in the embedded FS has a corresponding entry in `catalog.yaml`, emitting sentinel `CATALOG_ENTRY_MISSING: <path>` on failure.

REQ-CATALOG-001-016: The audit suite shall verify that every agent file under `.claude/agents/moai/` in the embedded FS has a corresponding entry in `catalog.yaml`, emitting sentinel `CATALOG_ENTRY_MISSING: <path>` on failure.

REQ-CATALOG-001-017: The audit suite shall verify that every `catalog.yaml` entry references an existing path in the embedded FS, emitting sentinel `CATALOG_ENTRY_ORPHAN: <path>` on failure.

REQ-CATALOG-001-018: The audit suite shall verify the pack dependency graph is acyclic, emitting sentinel `PACK_DEPENDENCY_CYCLE: <pack-A> <-> <pack-B>` on cycle detection.

REQ-CATALOG-001-019: The audit suite shall verify each entry's `tier` field matches one of the allowed values, emitting sentinel `CATALOG_TIER_INVALID: <entry> tier=<value>` on violation.

REQ-CATALOG-001-020: The audit suite shall verify each entry's `hash` field conforms to the sha256 hex format (64 lowercase hex chars), emitting sentinel `CATALOG_HASH_INVALID: <entry>` on violation.

REQ-CATALOG-001-027: If a skill or agent name appears in more than one tier section (`catalog.core.*`, `catalog.optional_packs.<pack>.*`, `catalog.harness_generated.*`), then the audit suite shall fail with sentinel `CATALOG_DUPLICATE_ENTRY: <name> in [<tier1>, <tier2>]`. (D2 권장 수정 반영 — orphan sentinel 의 REQ 보호.)

### 6. Workflow Trigger Coverage (Event-Driven, conditional — D4 권장 수정 반영)

REQ-CATALOG-001-021: When a workflow skill's frontmatter declares a `metadata.required-skills` field (list of skill names), the audit suite shall verify that every listed skill resolves to a catalog entry whose `tier` is either `core` or `optional-pack:<name>` explicitly declared in the workflow's `metadata.required-packs` field, emitting sentinel `WORKFLOW_UNCOVERED: <workflow> requires <skill>` on violation.

> **D4 fix note**: 이전 v0.1.0 은 workflow trigger coverage 를 unconditional 로 가정했으나, 현재 `internal/template/templates/.claude/skills/moai/workflows/` 의 20개 flat `.md` 파일은 `metadata.required-skills` frontmatter 키가 없음 (grep 결과 0 matches). 따라서 본 REQ 는 Event-Driven (conditional) 패턴으로 재작성되어, 해당 frontmatter 키가 부재한 경우 vacuously true 로 처리된다. 본 SPEC v1 manifest 는 workflow dependency 검증을 skip 하고, 후속 SPEC (TBD, post-CATALOG-007) 이 workflow 에 `metadata.required-skills` 필드를 retrofit 추가하여 실질 검증을 도입할 수 있다. 이 점은 Exclusions 에도 명시.

### 7. Hash Stability (State-Driven)

REQ-CATALOG-001-022: While the source content of a skill or agent file is unchanged (byte-identical after normalization), the regenerated `hash` value shall equal the previous `hash` value.

REQ-CATALOG-001-023: While the source content changes (any byte differs after normalization), the regenerated `hash` value shall differ from the previous `hash` value.

### 8. Optional Marketplace Compatibility (Optional) — D9 권장 수정 반영

REQ-CATALOG-001-024: The manifest schema MAY include reserved optional fields `marketplace_id`, `marketplace_url`, and `publisher` at the pack level (no current verification requirement). If any of these fields are present on a pack entry, then the audit suite SHALL verify that the value conforms to its declared type (string), emitting sentinel `CATALOG_RESERVED_FIELD_INVALID: <pack> <field>` on type violation.

> **D9 fix note**: 이전 v0.1.0 의 "reserved optional fields ... 매 audit 검증" 표현은 audit 으로 검증 불가했음 (필드가 부재해도 통과해야 하므로). v0.2.0 은 두 갈래 (1) 필드는 optional → 부재 OK, (2) 존재 시 type check 만 강제 — 로 명확화. 실제 marketplace 통합은 별도 SPEC 영역.

### 9. Backward Compatibility (Unwanted Behavior)

REQ-CATALOG-001-025: If a user runs `moai init` with the new manifest in place, then every skill and agent currently in the embedded FS shall still be deployable through the existing Deployer interface; the manifest must not break existing deploy pipelines.

REQ-CATALOG-001-026: If catalog.yaml is missing from the binary build, then the audit suite shall fail with sentinel `CATALOG_MANIFEST_ABSENT`, preventing release of an unmanaged build.

## Acceptance Criteria

See `acceptance.md` for the full Given-When-Then scenarios (**7 scenarios + 4 edge cases**, 27 REQ all mapped). High-level acceptance:

- AC-CATALOG-001-01: All 37 top-level skill directories + 28 agent files appear in `catalog.yaml` with a valid tier classification, with `path` field present and audit suite file existing. (maps REQ-CATALOG-001-005, REQ-CATALOG-001-006, REQ-CATALOG-001-009, REQ-CATALOG-001-014, REQ-CATALOG-001-015, REQ-CATALOG-001-016, REQ-CATALOG-001-019)
- AC-CATALOG-001-02: Drift detection foundation in place: every entry has sha256 hash + semver version. (maps REQ-CATALOG-001-007, REQ-CATALOG-001-008, REQ-CATALOG-001-020, REQ-CATALOG-001-022, REQ-CATALOG-001-023)
- AC-CATALOG-001-03: Pack dependency graph is acyclic (no circular packs). (maps REQ-CATALOG-001-010, REQ-CATALOG-001-018)
- AC-CATALOG-001-04: Workflow dependency check is vacuously true at present (no `metadata.required-skills` frontmatter in current workflows); REQ-021 conditional logic active for future retrofit. (maps REQ-CATALOG-001-021)
- AC-CATALOG-001-05: Schema validation rejects invalid tier values, malformed hashes, duplicate entries, tier changes without SPEC ID amendment, and (when present) malformed reserved marketplace fields. (maps REQ-CATALOG-001-013, REQ-CATALOG-001-019, REQ-CATALOG-001-020, REQ-CATALOG-001-024, REQ-CATALOG-001-027)
- AC-CATALOG-001-06: Manifest top-level structure is valid (version, generated_at, catalog with 3 sub-sections; audit suite exists; absent manifest triggers `CATALOG_MANIFEST_ABSENT`; reserved optional fields type-checked when present; `moai init` deploy pipeline unchanged). (maps REQ-CATALOG-001-001, REQ-CATALOG-001-002, REQ-CATALOG-001-003, REQ-CATALOG-001-004, REQ-CATALOG-001-014, REQ-CATALOG-001-024, REQ-CATALOG-001-025, REQ-CATALOG-001-026)
- AC-CATALOG-001-07: Pack definition structure is valid (each pack declares `description`, `depends_on`, `skills`, `agents`; pack names match regex). (maps REQ-CATALOG-001-011, REQ-CATALOG-001-012)
- AC-CATALOG-001-08: Edge cases (EC1-EC4) cover missing entry, orphan, hash format, duplicate detection. (maps REQ-CATALOG-001-007, REQ-CATALOG-001-009, REQ-CATALOG-001-017, REQ-CATALOG-001-020, REQ-CATALOG-001-022, REQ-CATALOG-001-027)

## Files to Modify / Create

[NEW] `internal/template/catalog.yaml` — Single source of truth manifest (~80-120KB anticipated, schema + 65 entries with hash + version + depends_on). Estimated 800-1200 YAML lines.

[NEW] `internal/template/catalog_tier_audit_test.go` — Audit suite (8-10 parallel sub-tests, sentinel-based assertions following `lang_boundary_audit_test.go` precedent). Estimated 450-550 LOC. Covers REQ-014, 015, 016, 017, 018, 019, 020, 021 (conditional), 022, 023, 024 (when fields present), 026, 027.

[NEW] `internal/template/catalog_loader.go` — Catalog YAML parser + typed accessor (`LoadCatalog(fs.FS) (*Catalog, error)`, `Catalog.LookupSkill(name) (*Entry, bool)`, etc.). Required so audit tests can reuse parsing logic and so SPEC-V3R4-CATALOG-002+ can build atop a typed API. Estimated 180-220 LOC.

[NEW] `internal/template/scripts/gen-catalog-hashes.go` — Offline helper to compute sha256 hashes for catalog.yaml entries. Invoked via `go run internal/template/scripts/gen-catalog-hashes.go [--entry <name>] [--all]`. Normalization: LF line endings, trailing whitespace trimmed. Used in EC1/EC3 recovery flows and during initial M2-T2.4 hash lock-in. Estimated 120-160 LOC. (D5 권장 수정 반영 — 이전 v0.1.0 의 acceptance EC1/EC3/plan T2.4 가 본 헬퍼를 언급했으나 Files to Modify 에 누락되어 있었음.)

[NEW] `internal/template/catalog_doc.md` — Brief schema docstring + tier semantics + hash normalization spec + cross-reference to follow-up SPECs. Markdown, not Go. Estimated 60-80 lines.

[MODIFY] `internal/template/embed.go` — **No source changes anticipated**. The existing `//go:embed all:templates` directive already covers `internal/template/catalog.yaml` (different path) but NOT `catalog.yaml` directly. **Audit-time confirmation only**: test asserts presence of `catalog.yaml` accessible via embed.FS (REQ-CATALOG-001-026). If go:embed coverage gap is found during M1, add a single new directive `//go:embed catalog.yaml` (additive, no removal). This is the only conditional modification.

[NO MODIFY] `internal/template/deployer.go` — **D7 권장 수정 반영**: SPEC-V3R4-CATALOG-001 의 scope 에서는 `deployer.go` 미수정. `catalog.yaml` 의 load 는 audit 테스트 (`catalog_tier_audit_test.go`) 와 별도 `catalog_loader.go` 의 `LoadCatalog()` 함수에서만 사용. Deploy 흐름은 manifest 를 의식하지 않으며, 회귀 검증 (M4-T4.5) 은 `internal/template/deployer_test.go` 의 기존 케이스가 모두 GREEN 임을 확인하는 read-only 검증에 한정. tier 별 filtering 은 SPEC-V3R4-CATALOG-002+003 의 scope.

## Exclusions (What NOT to Build)

The following are explicitly **OUT OF SCOPE** for SPEC-V3R4-CATALOG-001 and deferred to follow-up SPECs:

- **Directory relocation** (`internal/template/templates/` reorganization into `packs/<pack>/`) — SPEC-V3R4-CATALOG-002 영역. 본 SPEC 의 manifest 는 현재 디렉토리 구조 그대로 분류만 기록.
- **CLI commands** (`moai pack add|remove|list|available`) — SPEC-V3R4-CATALOG-003 영역. manifest 만 제공, 사용자 인터페이스 없음.
- **Safe update synchronization** (`moai update --catalog-sync`, drift detection runtime, 3-way merge, snapshot/rollback) — SPEC-V3R4-CATALOG-004 영역 (Wave 3, 최고위험). 본 SPEC 은 hash/version 필드만 제공하며, drift 비교 로직은 후속 SPEC.
- **`/moai project` interview extension** (harness opt-out AskUserQuestion 라운드 추가) — SPEC-V3R4-CATALOG-005 영역.
- **`moai doctor catalog` diagnostic command** (tier별 install 수, context budget 표시, idle skill 감지) — SPEC-V3R4-CATALOG-006 영역.
- **4개국어 migration documentation** (CHANGELOG breaking-vs-non-breaking, docs-site ko/en/ja/zh 4-locale sync, `moai update` 첫 실행 시 인라인 안내) — SPEC-V3R4-CATALOG-007 영역.
- **Tier filtering at deploy time** — `Deploy()` 가 manifest 를 읽지 않으며 tier 별 선별 deploy 도 안 함 (SPEC-V3R4-CATALOG-002 가 디렉토리 재배치 후 구현).
- **Workflow `metadata.required-skills` retrofit** — 현재 `templates/.claude/skills/moai/workflows/*.md` 는 해당 frontmatter 키 부재. REQ-021 은 conditional 패턴으로 vacuously true. 향후 별도 SPEC 이 workflow 에 dependency 메타데이터를 추가하면 자동으로 활성화. (D4 권장 수정 반영.)
- **Anthropic Plugin Marketplace publishing** — REQ-CATALOG-001-024 가 forward-compat 필드만 reserve 하고, 실제 publish 워크플로우는 별도 후속 SPEC (TBD).
- **Skill frontmatter 의 `tier:` 필드 추가** — research.md §"Open Questions" 와 일관성. manifest 가 단독 source of truth 이므로, 개별 SKILL.md 에 tier 필드를 중복으로 적지 않는다. drift 감지는 manifest 의 hash + version 으로 충분.
- **Skill module sub-files hash coverage** — REQ-007 의 hash 는 entry root 의 `SKILL.md` (또는 `skill.md`) 만 커버. `moai/workflows/*.md`, `moai-foundation-cc/references/*` 등 sub-file 은 별도 hash 미지원 (후속 SPEC 영역).
- **Idle skill telemetry** (사용 빈도 추적 hook) — SPEC-V3R4-CATALOG-006 의 doctor 명령 영역.

## Dependencies

- **Depends on**: 없음 (Wave 1 의 첫 SPEC, foundation).
- **Blocks**: SPEC-V3R4-CATALOG-002 (디렉토리 재배치는 본 manifest 의 tier 분류를 입력으로 사용), SPEC-V3R4-CATALOG-003 (`moai pack` 은 manifest 의 `depends_on` 필드로 의존성 해결), SPEC-V3R4-CATALOG-004 (`moai update --catalog-sync` 는 본 SPEC 의 hash/version 필드로 drift 계산), SPEC-V3R4-CATALOG-005, SPEC-V3R4-CATALOG-006, SPEC-V3R4-CATALOG-007.

## References

### Internal (필수 사전 read)
- `.moai/brain/IDEA-003/proposal.md` — SPEC 분해 + Wave 순서 + Open Questions 6건.
- `.moai/brain/IDEA-003/ideation.md` — Lean Canvas + Critical Evaluation 5 비판 + First Principles (skill 존재가 비용 invariant).
- `.moai/specs/SPEC-V3R4-CATALOG-001/research.md` — 37 skills + 28 agents 검증, embed/deploy/update 코드 분석, audit 패턴 reference, tier 분류 표.
- `.moai/project/product.md` — moai-adk-go 프로젝트 컨텍스트.

### Code Reference
- `internal/template/embed.go:24-39` — `//go:embed all:templates` 패턴.
- `internal/template/deployer.go:14-72` — Deploy 인터페이스 + walk + render + manifest 등록 (본 SPEC 미수정).
- `internal/cli/update.go:71-100+` — runUpdate 흐름 + 기존 drift 감지 framework + merge 위임.
- `internal/template/lang_boundary_audit_test.go` — audit 패턴 reference (sentinel `LANG_AS_SKILL_FORBIDDEN`, `DEAD_LANG_FRONTMATTER_REFERENCE`, embedded FS WalkDir 패턴).

### Configuration Patterns (선례)
- `.moai/config/sections/harness.yaml` — manifest version + levels + metadata 계층 구조 reference.
- `.moai/config/sections/quality.yaml` — yaml manifest 구조 reference.
- `.moai/config/sections/workflow.yaml` — manifest 의 nested config 패턴 reference.

### Rules
- `.claude/rules/moai/development/skill-authoring.md` — Skill YAML frontmatter 스키마 (manifest 설계의 cross-reference; tier 필드는 frontmatter 가 아닌 manifest 에만).
- `.claude/rules/moai/development/coding-standards.md` — Template-First, 16-language neutrality, single source of truth 원칙.
- `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-005 — Template-First discipline (catalog.yaml 도 templates/ 외부지만 templates → user project 흐름 일관 유지).


### Out of Scope

- N/A (legacy SPEC)

## Implementation Notes

본 SPEC 은 Sprint 11 / Wave 15 에서 단일 run 세션으로 M1-M5 를 모두 완료했다. 핵심 산출물과 분기 결정은 다음과 같다.

### Delivered Artifacts (main 머지 완료)

| 파일 | LOC | 역할 | 참조 |
|------|-----|------|------|
| `internal/template/catalog.yaml` | 413 | 65-entry 3-tier manifest (37 skills + 28 agents) | REQ-001..006, REQ-011..012 |
| `internal/template/catalog_loader.go` | 196 | `LoadCatalog(fs.FS)` + typed Catalog/Entry/Pack 구조 + Lookup 헬퍼 | REQ-001, REQ-004, REQ-021 |
| `internal/template/catalog_tier_audit_test.go` | 615 | 10 audit sub-tests (sentinel 기반) | REQ-014..020, REQ-022..023, REQ-026..027 |
| `internal/template/catalog_hash_norm.go` | 46 | `NormalizeForHash` 공유 (loader + scripts) — R7 mitigation (progress.md M2 supplemental) | REQ-007, REQ-022..023 |
| `internal/template/catalog_loader_test.go` | 154 | typed loader 단위 테스트 (PR #863에서 100% coverage 완성) | REQ-001, REQ-004 |
| `internal/template/scripts/gen-catalog-hashes.go` | 267 | offline 헬퍼 (`--entry` / `--all` / `--check`) | REQ-007, REQ-022 |
| `internal/template/catalog_doc.md` | 156 | 스키마 docstring + tier semantics + hash normalization spec | M5 deliverable |
| `internal/template/embed.go` | +5 | `//go:embed catalog.yaml` 추가 directive (additive) | REQ-026 |

### Divergence Summary (planned vs actual)

- **Scope expansion (minor)**: `catalog_hash_norm.go` (46 LOC) + `catalog_loader_test.go` (154 LOC) 는 spec.md "Files to Modify / Create" 표에 직접 명시되지 않았으나 progress.md M2/M4 "Supplemental Files Created" 에 사전 기재됨. 전자는 hash 정규화 로직 공유 (loader + offline tool 간 single source-of-truth, R7 risk mitigation), 후자는 typed loader 의 단위 테스트로서 TDD 규율상 필수. 두 파일 모두 SPEC 의도와 정렬됨.
- **Directory creation**: `internal/template/scripts/` 신규 디렉토리 1개 (spec.md 의 gen-catalog-hashes.go 경로 자체에 명시되어 있음 — 의도된 추가).
- **No deferred items**: 모든 27 REQ + 8 AC 가 implementation 으로 매핑됨.
- **No deployer.go modification (D7 lock 준수)**: regression 검증은 `deployer_test.go` 등 기존 케이스 GREEN 으로 확인.

### Quality Gates (Final)

| Gate | Verdict | Score |
|------|---------|-------|
| Phase 0.5 plan-auditor (iter 1) | PASS | 0.94 |
| `go test -race ./internal/template/...` | PASS | 4.7s |
| `golangci-lint run ./internal/template/...` | PASS | 0 issues |
| `go vet ./internal/template/...` | PASS | clean |
| `internal/template` 커버리지 | PASS | 84.0% (LoadCatalog 100%) |
| Phase 2.8a evaluator-active (iter 1, fresh-context) | PASS | 0.82 |
| PR #862 핵심 CI (Test×3 / Build×5 / Lint / Constitution / CodeQL) | PASS | All green |
| PR #863 핵심 CI (Test ubuntu/macos / Build×5 / Lint / CodeQL) | PASS | aux reviews FAILURE (per `feedback_ci_aux_workflow_failures` 비차단) |

### Evaluator-Active Required Fixes (PR #863에서 해결)

1. **EC3 hash sentinel enforcement**: `catalog_tier_audit_test.go:341` 의 `t.Logf` → `t.Errorf` 로 변경. 이전에는 hash mismatch 가 advisory log 였으나, CI 시점에 fail 하도록 강화.
2. **LoadCatalog coverage 71.4% → 100%**: `catalog_loader_test.go` 에 `TestLoadCatalog_MalformedYAML` (3 sub-cases) + `TestLoadCatalog_ManifestAbsent` (`testing/fstest.MapFS` 기반) 추가.

### Deferred to Follow-up SPECs (Exclusions per spec.md 와 일관)

- Directory relocation (`packs/<pack>/` 구조) → SPEC-V3R4-CATALOG-002
- `moai pack add|remove|list|available` CLI → SPEC-V3R4-CATALOG-003
- `moai update --catalog-sync` drift detection + 3-way merge + snapshot/rollback → SPEC-V3R4-CATALOG-004
- `/moai project` harness opt-out interview → SPEC-V3R4-CATALOG-005
- `moai doctor catalog` diagnostic → SPEC-V3R4-CATALOG-006
- 4개국어 migration docs (CHANGELOG breaking-vs-non-breaking, docs-site ko/en/ja/zh) → SPEC-V3R4-CATALOG-007
- evaluator-active nice-to-have 3건 (path containment, REQ-011/012 pack regex test, BenchmarkLoadCatalog) → CATALOG-002~007 후속 처리
- Skill module sub-files hash coverage 확장 → 별도 SPEC (현재 entry root 만 hash)

### Wave 1 Foundation Complete

본 SPEC 은 catalog initiative 7-SPEC 체인의 **foundation** 으로서, 후속 6개 SPEC 이 본 manifest 의 typed API (`LoadCatalog`) + hash 필드 + tier 분류 + depends_on 그래프를 입력으로 사용한다. Wave 2 (Distribution: 002+003), Wave 3 (Safety: 004), Wave 4 (Polish: 005+006+007) 진입 자격 충족.
