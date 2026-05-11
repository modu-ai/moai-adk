# Implementation Plan — SPEC-V3R4-CATALOG-001

> **plan.md v0.2.0 (D3/D4/D8/D11 권장 수정 반영)**: 카운트 통일 (37 skills + 28 agents = 65 entries), workflows flat-md layout 반영 (T3.9 재작성), Estimated Complexity LOC 재계산, T1.1 schema 필드 분류 명확화.

## Goal

3-tier (core / optional-pack / harness-generated) 카탈로그 manifest 스키마를 `internal/template/catalog.yaml` 로 도입하고, **37 skills + 28 agents = 65 entries** 의 tier 분류를 lock-in 하며, `catalog_tier_audit_test.go` 로 무결성을 강제한다. 후속 6개 SPEC 의 foundation 을 확립한다.

> **Catalog Entry 정의 (spec.md §Overview 와 일관성)**: "Catalog entry" 는 `internal/template/templates/.claude/skills/` 바로 아래 top-level 디렉토리 1개 (37개; `moai/` 컨테이너는 단일 logical skill 로 취급) 또는 `internal/template/templates/.claude/agents/moai/` 바로 아래 `.md` 파일 1개 (28개) 를 의미한다. `moai/workflows/*.md`, `moai/team/*.md`, `moai/references/*.md` 등 sub-file 은 `moai` skill 의 모듈로 별도 entry 가 아니다.

## Approach

본 SPEC 은 **데이터 + 검증 + 최소 통합** 의 3-layer 전략을 따른다:

1. **Schema-first**: catalog.yaml 스키마를 먼저 확정 (M1). 후속 6개 SPEC 의 인터페이스를 미리 정의하므로 schema 결정이 가장 중요.
2. **Data lock-in**: 현재 자산 **65 entries (37 skills + 28 agents)** 를 manifest 에 등록 (M2). research.md 의 Tier Classification Map 을 신뢰할 수 있는 기준으로 채택.
3. **Audit-driven invariants**: `lang_boundary_audit_test.go` 패턴을 차용한 audit suite 로 schema 위반을 CI 단계에서 차단 (M3). manifest 와 disk 의 drift, pack DAG 순환, tier 오타, hash 형식 모두 자동 검출.
4. **Minimal integration**: catalog.yaml 로딩은 audit test 와 향후 SPEC 의 입구 (catalog_loader.go) 로만 통합. Deploy() 의 actual filtering 은 SPEC-002 영역으로 격리하여 회귀 위험 최소화 (M4).
5. **Documentation**: schema docstring + 후속 SPEC cross-reference (M5).

본 SPEC 의 변경은 **현재 deploy 동작에 영향이 0 이어야 한다** (`moai init` 이 여전히 모든 자산을 deploy). manifest 는 selective deploy 의 기반만 제공하며, 실제 selective 동작은 SPEC-002+003 에서 도입.

## Task Decomposition

전체 task ~16개. development_mode 는 quality.yaml 기준 (TDD 가정). 단, 본 SPEC 은 데이터 중심이라 audit test 작성 (M3) 이 TDD-style RED 단계로 적합하지만, manifest 데이터 작성 (M2) 은 절차적 작업이라 GREEN 으로 묶음.

### M1 — Schema 설계 + 결정 lock-in

- T1.1: `catalog.yaml` schema 초안 작성. (D11 권장 수정 반영) 필드 분류:
  - **Top-level (3)**: `version` (semver string), `generated_at` (ISO 8601), `catalog` (object).
  - **Catalog sub-sections (3)**: `catalog.core`, `catalog.optional_packs` (map[pack-name]→Pack), `catalog.harness_generated`.
  - **Per-entry (6)**: `name`, `path`, `tier`, `hash`, `version`, optional `depends_on` (optional-pack 만).
  - **Per-pack (4 required + 3 reserved optional)**: `description`, `depends_on`, `skills`, `agents`; optional `marketplace_id`, `marketplace_url`, `publisher` (REQ-024).
- T1.2: Open Question 1-6 권장안을 사용자에게 confirm (Decision Point 1). 결정 결과를 plan.md 본 섹션 끝에 기록.
- T1.3: 기존 `harness.yaml`/`quality.yaml` manifest 패턴과 schema 호환성 검토 (YAML indentation, version 필드 위치, comment 스타일).
- T1.4: schema 의 forward-compatibility 검토 — REQ-CATALOG-001-024 (marketplace 필드 reserve) 가 schema 에 자연스럽게 통합되는지 확인.

### M2 — Tier Data Lock-in

- T2.1: research.md §"Tier Classification Map" 의 18개 core skills + 15개 core agents 분류를 catalog.yaml 의 `catalog.core` 섹션에 입력.
- T2.2: research.md 의 optional-pack 후보 9개 (backend, frontend, mobile, chrome-extension, auth, deployment, design, devops, testing) 를 `catalog.optional_packs` 에 입력.
- T2.3: harness-generated 자산 (`builder-harness` 등) 을 `catalog.harness_generated` 에 입력.
- T2.4: 각 entry 의 `hash` 필드를 sha256 으로 계산 (LF endings + trailing whitespace 정규화). 초기 빌드 시 `internal/template/scripts/gen-catalog-hashes.go` (offline 헬퍼) 로 생성.
- T2.5: 각 entry 의 `version` 필드 초기값 `"1.0.0"` 으로 lock-in. 향후 entry 단위 변경 시 semver bump.
- T2.6: 9개 pack 의 `depends_on` 그래프 작성 (예: design → frontend, deployment → backend). DAG 검증 (audit M3 가 잡지만, 수동으로 한 번 확인).
- T2.7: `catalog.version` 필드 = `"1.0.0"`, `generated_at` = `2026-05-12` 로 설정.

### M3 — Audit Suite (TDD RED → GREEN)

- T3.1: `catalog_tier_audit_test.go` 골격 작성 — package 선언, import (`io/fs`, `regexp`, `strings`, `testing`, sha256 패키지), `EmbeddedTemplates()` 호출 패턴 reuse.
- T3.2: `TestCatalogManifestPresent` — REQ-CATALOG-001-026 강제. embedded FS 에 `internal/template/catalog.yaml` 존재 검증. Sentinel: `CATALOG_MANIFEST_ABSENT`.
- T3.3: `TestAllSkillsInCatalog` — REQ-CATALOG-001-005 강제. `.claude/skills/` walk + catalog.yaml 의 모든 tier 합집합 비교. Sentinel: `CATALOG_ENTRY_MISSING: <skill-path>`.
- T3.4: `TestAllAgentsInCatalog` — REQ-CATALOG-001-006 강제. `.claude/agents/moai/` walk + catalog.yaml 합집합 비교. Sentinel: `CATALOG_ENTRY_MISSING: <agent-path>`.
- T3.5: `TestCatalogReferencesValid` — REQ-CATALOG-001-017 강제. catalog.yaml 의 모든 entry path 가 embedded FS 에 실제 존재하는지 (orphan 검출). Sentinel: `CATALOG_ENTRY_ORPHAN: <path>`.
- T3.6: `TestCatalogTierValid` — REQ-CATALOG-001-019 강제. tier 값이 `core | optional-pack:<name> | harness-generated` regex 매칭. Sentinel: `CATALOG_TIER_INVALID: <entry> tier=<value>`.
- T3.7: `TestPackDependencyDAG` — REQ-CATALOG-001-018 강제. 위상정렬 또는 DFS-based 순환 검출. Sentinel: `PACK_DEPENDENCY_CYCLE: <pack-A> <-> <pack-B>`.
- T3.8: `TestManifestHashStability` — REQ-CATALOG-001-020, 022, 023 강제. 각 entry 의 hash 가 sha256 hex (64 lowercase chars) 형식. 추가: 동일 file content 로 hash 재계산해도 결과 일치 (golden test). Sentinel: `CATALOG_HASH_INVALID: <entry>`, `CATALOG_HASH_UNSTABLE`.
- T3.9: `TestWorkflowTriggerCoverage` — REQ-CATALOG-001-021 강제 (D4 권장 수정 반영). 현재 `.claude/skills/moai/workflows/` 는 20개 flat `.md` 파일 layout (subdirectory + SKILL.md 아님) 이며, `metadata.required-skills` frontmatter 키는 0건 (`grep -r 'required-skills' .claude/skills/moai/workflows/` → 0 matches). 따라서 본 테스트는 **conditional vacuously-true** 패턴으로 작성: 각 workflow `.md` 파일의 frontmatter 파싱 → `metadata.required-skills` 키 부재 시 통과 (vacuously true) → 키 존재 시 각 의존성을 catalog 에서 resolve 검증 + `metadata.required-packs` 와 cross-check. v1 manifest 단계에서는 모든 workflows 가 vacuously true 통과. 후속 SPEC 이 retrofit 시 자동 활성화. Sentinel: `WORKFLOW_UNCOVERED: <workflow> requires <skill>`.
- T3.10: `TestCatalogNoDuplicateEntries` — REQ-CATALOG-001-027 강제 (D2 권장 수정 반영, 신규 REQ). skill 또는 agent 이름이 `catalog.core`, `catalog.optional_packs.<pack>`, `catalog.harness_generated` 의 2개 이상 tier section 에 동시 등장하는지 검출. 구현: tier 별 entry 이름 슬라이스 합집합 vs 합계 비교. Sentinel: `CATALOG_DUPLICATE_ENTRY: <name> in [<tier1>, <tier2>]`.

### M4 — Loader + Minimal Integration

- T4.1: `internal/template/catalog_loader.go` 작성. `LoadCatalog(fsys fs.FS) (*Catalog, error)` 함수. yaml.Unmarshal 로 catalog.yaml 파싱.
- T4.2: `Catalog` struct 정의 — Version, GeneratedAt, Core, OptionalPacks (map[string]*Pack), HarnessGenerated 필드.
- T4.3: `Catalog.LookupSkill(name string) (*Entry, bool)` + `Catalog.LookupAgent(name string) (*Entry, bool)` accessor 추가 (audit test + 후속 SPEC 의 사용자).
- T4.4: audit test M3 를 catalog_loader 로 refactor — direct YAML parse 가 아닌 LoadCatalog() 호출. 단일 진실 공급원 유지.
- T4.5: **현재 Deploy() 흐름 영향 검증** — catalog.yaml 추가가 기존 deploy 결과를 변경하지 않음을 통합 테스트로 확인 (`internal/template/deployer_test.go` 의 기존 케이스가 모두 GREEN 유지).

### M5 — Documentation + Schema Notes

- T5.1: `internal/template/catalog_doc.md` 작성 — schema 필드별 설명, tier 의미, hash 알고리즘 (sha256 정규화 절차), 후속 SPEC cross-reference (`SPEC-V3R4-CATALOG-002/003/004/005/006/007`).
- T5.2: `catalog.yaml` 내 YAML 주석 — 각 섹션 (core / optional_packs / harness_generated) 시작에 짧은 설명 + IDEA-003 proposal.md 링크.

## Risks & Mitigations

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | catalog.yaml schema 가 후속 SPEC-002~007 와 충돌 (예: SPEC-004 가 hash 형식을 다르게 가정) | 중 | 중 | M1 에서 후속 6 SPEC scope 모두 사전 검토. hash 알고리즘 / depends_on 형식을 OQ 권장안에 명시하여 사용자 confirm. |
| R2 | hash 알고리즘 선택 (sha256 vs sha1 vs blake3) 로 인한 향후 호환성 부담 | 낮 | 중 | OQ6 권장안 sha256 채택. Go stdlib `crypto/sha256` 사용, 외부 의존성 0. |
| R3 | audit test 가 너무 strict 해서 dev 마찰 증가 (예: tier 필드 변경마다 다중 sentinel) | 중 | 낮 | sentinel 메시지에 명확한 fix 가이드 포함 (e.g., "add entry to catalog.yaml under tier=<X>"). 초기에는 hard fail, 후속 SPEC 에서 warning option 검토. |
| R4 | go:embed 변경이 binary 크기 + build 속도 영향 | 낮 | 낮 | catalog.yaml 약 80-120KB 추정 (65 entries × 6 fields). 현재 templates/ 약 1.2MB 대비 6.7-10% 증가. 무시 가능. |
| R5 | 65 entries (37+28) 의 hash 계산이 build 시간에 영향 | 낮 | 낮 | hash 는 offline 헬퍼 (`internal/template/scripts/gen-catalog-hashes.go`) 에서 1회 계산 후 catalog.yaml 에 정적 저장. build 시 재계산 안 함 (audit test 만 read-only 검증). |
| R6 | research.md 의 tier 분류 표가 잘못된 자산 매핑 (예: moai-foundation-thinking 을 core 로 분류했으나 실제 의존성은 optional-pack 에 적합) | 중 | 중 | T2.1-T2.6 작성 시 각 자산의 workflow trigger + dependency 를 catalog_doc.md 에 명시. plan-auditor 가 검증. 사용자 검토 (Decision Point 2) 로 최종 confirm. |
| R7 | catalog.yaml 의 sha256 hash 가 OS 별 line ending 차이 (LF vs CRLF) 로 불일치 | 중 | 낮 | T2.4 의 normalization 절차 (LF endings, trailing whitespace trim) 를 catalog_doc.md 에 명시. audit test (T3.8) 가 hash stability 검증. Windows CI 에서도 GREEN 보장. |
| R8 | `moai update` 가 본 SPEC 이후에도 모든 자산을 강제 동기화하여 사용자 손실 발생 | 중 | 높음 | **사용자 추가 제약**. 본 SPEC 은 hash 필드만 제공 (drift 감지 입력). 실제 안전 동기화는 SPEC-V3R4-CATALOG-004 에서 구현. 본 SPEC 단독 머지 시 동작 변경 0 (Deploy 흐름 영향 0). |
| R9 | 신규 skill 추가 시 catalog.yaml 등록 누락 → CI 실패 | 중 | 낮 (의도된 동작) | audit test 의 sentinel 메시지에 정확한 추가 위치 + tier 선택 가이드 포함. README / CONTRIBUTING 에 catalog.yaml 갱신 의무 명시. |

## Open Questions — Recommended Defaults (Decision Point 1)

research.md 의 6 Open Questions 각각에 권장안 제시. 사용자가 Decision Point 1 에서 confirm.

### OQ1: Manifest 파일 위치

- **권장 (옵션 A)**: `internal/template/catalog.yaml` — code-driven, embed 가능, build-time artifact.
- 대안 B: `.moai/catalogs/core.yaml` + `.moai/catalogs/packs.yaml` (tier 별 분할). 가독성 ↑ but 단일 source-of-truth 위배.
- 대안 C: `internal/template/templates/.moai/config/sections/catalog.yaml` — 사용자 프로젝트의 `.moai/config/` 에 deploy 됨. 사용자가 직접 편집 가능 but 의도와 충돌 (catalog 는 build metadata).
- **Rationale**: catalog 는 distribution metadata (사용자 편집 대상 아님). `internal/template/` 외부 단일 파일로 두는 게 single source of truth + Template-First (CLAUDE.local.md §2) 원칙 일관.

### OQ2: Pack 개수 및 이름

- **권장**: 9 packs — `backend`, `frontend`, `mobile`, `chrome-extension`, `auth`, `deployment`, `design`, `devops`, `testing`.
- 대안: 7 packs (mobile + chrome-extension 통합, devops + deployment 통합).
- **Rationale**: brain IDEA-003 proposal.md 와 research.md tier 분류표가 9 packs 로 정렬. 추후 사용 빈도 telemetry (SPEC-006) 에서 통합/분할 가능.

### OQ3: Harness auto-activation 기본값

- **권장**: opt-out (`harness_auto_activate: true` 가 manifest 기본값). brain Phase 1 사용자 답변 직접 반영.
- 대안: opt-in (`harness_auto_activate: false`). 안전성 우선이나 사용자 명시 답변과 상충.
- **Note**: SPEC-V3R4-CATALOG-005 의 `/moai project` 인터뷰 영역이지만, manifest 에 기본값 필드를 추가하여 후속 SPEC 의 입력으로 제공.

### OQ4: CLI 문법 (manifest 영향 받는 영역만)

- **권장**: `moai pack {add|remove|list|available}` (Git-style 동사 + 객체).
- **Note**: SPEC-V3R4-CATALOG-003 영역. manifest 자체는 CLI-agnostic 으로 설계 (pack 이름 + depends_on 만 노출).

### OQ5: 기존 프로젝트 migration 전략

- **권장**: non-breaking by default. `moai update` 단독 실행 시 catalog.yaml 만 동기화 (사용자 자산 보존). `--catalog-sync` 명시적 opt-in 으로만 tier-based 재정렬 수행.
- 대안: aggressive (자동 tier 재배치). 위험: 데이터 손실.
- **Note**: SPEC-V3R4-CATALOG-004 영역. 본 SPEC 의 manifest 에 `legacy_compat: true` 기본 필드는 두지 않음 (SPEC-004 가 자체 flag 정의).

### OQ6: Hash 알고리즘

- **권장**: **sha256** (Go stdlib `crypto/sha256`, 64 lowercase hex chars 출력).
- 대안 A: git blob hash (sha1). git 통합 자연스러우나 sha1 collision 우려.
- 대안 B: blake3. 빠르나 외부 의존성 (`github.com/zeebo/blake3`).
- **Rationale**: 보안 표준 + stdlib + audit test (T3.8) regex 단순화 (`^[0-9a-f]{64}$`).

## MX Tag Plan

본 SPEC 의 산출물은 catalog.yaml (데이터) + audit test (Go) + loader (Go) + doc (Markdown).

**MX 적용 대상**:
- `internal/template/catalog_tier_audit_test.go` 의 Sentinel 발신 함수 — @MX:ANCHOR 부착. `lang_boundary_audit_test.go` 선례와 동일. high fan_in 무결성 강제.
- `internal/template/catalog_loader.go` 의 `LoadCatalog()` — @MX:NOTE 부착. intent 명시 (single source of truth, hash 검증 위임 안 함 — loader 는 parse only).
- `internal/template/catalog.yaml` — 데이터 파일, MX 태그 불요 (YAML 주석으로 대체).

**MX 코멘트 언어**: `.moai/config/sections/language.yaml` `code_comments: ko` 이지만, SPEC-V3R3-MX-INJECT-001 의 `feedback_mx_tag_language.md` 기준 (2026-05-05): MX 태그 설명은 영문. 단, NOTE/REASON 본문은 ko/en 혼용 가능.

## Estimated Complexity

(D8 권장 수정 반영 — 이전 v0.1.0 의 LOC 추정이 optimistic 했음. catalog.yaml ~600 lines 가 spec.md 의 ~80KB anticipated 와 불일치, audit test ~350 LOC 가 `lang_boundary_audit_test.go` ~250 LOC (9 sub-tests) 대비 underestimate. v0.2.0 은 sentinel discipline + 10 sub-test 기반 재계산.)

- **Task count**: ~28 (M1: 4, M2: 7, M3: 10 (T3.10 신규), M4: 5, M5: 2). Grouped into M1-M5 milestones; leaf tasks ~28 including sub-steps.
- **LOC delta**:
  - `catalog.yaml`: **800-1200 lines (~80-120KB)** — 65 entries × ~6 fields × YAML indentation + 9개 pack 정의 + comments. spec.md §"Files to Modify" 와 일관.
  - `catalog_tier_audit_test.go`: **~450-550 LOC** — 10 sub-tests (T3.1-T3.10), sentinel-based assertions, parallel test 패턴, helper functions (walkSkills, walkAgents, parseManifest, dagCycleDetect). `lang_boundary_audit_test.go` ~250 LOC (9 sub-tests) 대비 entry 65개의 walk + hash 검증 + DAG 추가.
  - `catalog_loader.go`: ~150-200 LOC — Catalog struct + LoadCatalog + 2 LookupX accessors + yaml.Unmarshal helpers.
  - `gen-catalog-hashes.go`: ~80-120 LOC — file walker + sha256 + LF normalization + CLI flag parser.
  - `catalog_doc.md`: ~60-80 lines — markdown documentation.
  - **Total: ~1500-1900 LOC** (Go + YAML + Markdown).
- **Risk level**: **Low-Medium** (Wave 1 foundation, deploy 영향 0, audit test 만 새로 fail). 회귀 위험은 R6 (tier 분류 오류) 와 R7 (hash 불안정) 두 곳에 집중되어 audit test 로 자동 검출. LOC 증가에 따라 review 부담은 medium.
- **사용자 검토 필요 지점**: Decision Point 1 (OQ 6건 권장안 confirm), Decision Point 2 (M2 tier 분류 표 final review).
