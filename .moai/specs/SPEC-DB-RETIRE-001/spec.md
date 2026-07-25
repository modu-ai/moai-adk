---
id: SPEC-DB-RETIRE-001
title: "DB 문서화 서브시스템 전면 제거"
version: "0.1.0"
status: completed
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P2
phase: "v3.0.x maintenance"
module: "internal/hook/dbsync + internal/cli + internal/template + guidance skills"
lifecycle: spec-anchored
tags: "cleanup, removal, db, tech-debt, template"
tier: M
related_specs: [SPEC-DB-SYNC-001, SPEC-DB-SYNC-RELOC-001, SPEC-DEPRECATEDPATHS-RECONCILE-001]
---

# SPEC-DB-RETIRE-001 — DB 문서화 서브시스템 전면 제거

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | 최초 작성 (plan-phase draft). 정적 분석 보고서 `.moai/reports/db-scaffold-sync-logic-20260725.md`의 결함 5건 확정 + 사용자 전면 제거 승인 근거. |

## §A. 배경 (Context) — WHY

독립 정적 분석(`.moai/reports/db-scaffold-sync-logic-20260725.md`)이 DB 문서화 서브시스템이 **연결되지 않은 vaporware**임을 5건의 결함으로 확정했다.

1. **훅 휴면(dormant)**: `db-schema-sync`는 CLI 서브커맨드로 등록만 되어 있고 어떤 훅 배선에도 연결되지 않았다 — `.claude/settings.json`·`settings.json.tmpl`·`handle-*.sh` 래퍼 모두 부재(실측으로 재확인: 로컬/템플릿 어디에도 `handle-db-*.sh` 래퍼 없음, settings에 `db-schema-sync` 배선 0건). 광고된 "이후 `/moai sync`가 자동으로 db/를 갱신" 체인은 기계적 트리거가 없다.
2. **`db.enabled` 소비 코드 0건**: 어떤 Go 코드도 `db.enabled`를 읽지 않는다(grep 0 hits). 게이팅은 지침 서술에만 존재. `loadMigrationPatterns`는 `migration_patterns`만 읽는다.
3. **파서 부재(vaporware)**: `internal/db/parser/` 패키지가 존재하지 않는다. 스텁은 원본 파일 내용을 그대로 반환한다.
4. **라이터 부재**: `moai init` 이후 `.moai/project/db/*.md`를 갱신하는 Go 코드가 없다. 훅은 cache에 proposal.json만 쓰고, 실제 문서 재작성은 배선되지 않은 오케스트레이터 흐름에 위임된다.
5. **지침-코드 불일치**: `doc-generation.md`는 db/ 템플릿이 "첫 `/moai sync`에 생성"된다고 주장하나 실제 생성은 `moai init`(Logic A)이며, 훅은 파일을 생성·갱신하지 않는다.

사용자는 이 서브시스템의 **전면 제거**를 승인했다(project memory `project_db_retire_removal_approved.md`, 2026-07-24). 수리(gate/wire/parser/writer/narrative)가 아니라 제거를 택한 근거: 5건 모두 "구축은 됐으나 연결되지 않은" 상태이고, 완성에 드는 비용(파서 구현 + 트리거 배선 + enabled-게이트 소비 + proposal→approval→rewrite 오케스트레이터 흐름)이 이 기능의 실제 가치를 초과한다.

## §B. 요구사항 (Requirements) — GEARS

> 코드 식별자·파일 경로·심볼은 영어, 서술은 한국어. 모든 REQ는 acceptance.md의 1개 이상 AC로 추적 가능.

### B.1 그룹 1 — Go: dbsync 패키지 + CLI 배선 제거

- **REQ-DBR-001** (Unwanted): The `moai` binary **shall not** register a `db-schema-sync` hook subcommand.
- **REQ-DBR-002** (Unwanted): The source tree **shall not** contain the `internal/hook/dbsync/` package (3 files).
- **REQ-DBR-003** (Ubiquitous): The `internal/cli/hook.go` file **shall** be free of the `dbsync` import, the `runDBSchemaSync` handler, the `loadMigrationPatterns` function, the `defaultMigrationPatterns` variable, and the now-orphaned `splitLines` / `trimSpace` helpers whose sole caller was `loadMigrationPatterns`.
- **REQ-DBR-004** (Ubiquitous): The `internal/cli/hook_e2e_test.go` `utilitySubcmds` map **shall** be free of the `"db-schema-sync"` entry.

### B.2 그룹 2 — 참조 해제 (정밀 편집, 통삭제 금지)

- **REQ-DBR-005** (Ubiquitous): The `internal/defs/dirs.go` **comments** **shall not** describe `db.yaml` as a live v3 config read by `loadMigrationPatterns`. This is a **comment-only** correction: at the current baseline `db.yaml` is not yet a `DeprecatedPaths` slice entry, so the comment edit alone changes no count. The authoritative slice change — ADDING `.moai/config/sections/db.yaml` as a new `DeprecatedPaths` entry and the resulting count transition (live baseline **39 → 40**, verified `internal/defs/dirs_test.go:32` `const want = 39`) plus the atomic `dirs_test.go` update — is owned **solely by REQ-DBR-022**. REQ-005 (comment scope) and REQ-022 (slice + count scope) do not overlap and do not contradict.
- **REQ-DBR-006** (Ubiquitous): The `internal/cli/v2_detection.go` comment **shall not** reference `db.yaml` as an un-deprecated live config.
- **REQ-DBR-007** (Unwanted): Neither the `internal/config/audit_registry.go` `yamlAuditExceptions` map (real-tree audit) nor the `internal/config/audit_loader_completeness_test.go` `acknowledgedUnloadedSections` slice (template-tree audit) **shall** contain the `"db"` entry. The `yamlAuditExceptions["db"]` removal **shall** be coordinated with the local `db.yaml` deletion (REQ-DBR-016; `TestAuditParity` scans the real tree), and the `acknowledgedUnloadedSections["db"]` removal **shall** be coordinated with the template `db.yaml` deletion (REQ-DBR-011; `TestAuditLoaderCompleteness` scans the template sections dir), so that no on-disk `db.yaml` is left orphaned without a registry entry on either side.
- **REQ-DBR-008** (Unwanted): The `internal/template/internal_content_leak_test.go` C7 "real restricted-package path" list **shall not** reference the deleted `internal/hook/dbsync/db_schema_sync.go`; it **shall** cite a real existing `internal/**` Go path instead.
- **REQ-DBR-009** (Ubiquitous): The `internal/template/settings_test.go` `TestRender_DbSchemaChangeHook_Removed` comment **shall not** claim that the `internal/hook/dbsync` package and the `moai hook db-schema-sync` CLI "remain intact".

### B.3 그룹 3 — 템플릿 제거

- **REQ-DBR-010** (Unwanted): The `internal/template/templates/.moai/project/db/` directory (7 files: `README.md`, `schema.md`, `erd.mmd`, `migrations.md`, `queries.md`, `rls-policies.md`, `seed-data.md`) **shall not** exist.
- **REQ-DBR-011** (Unwanted): The `internal/template/templates/.moai/config/sections/db.yaml` file **shall not** exist.
- **REQ-DBR-012** (Event-driven): **When** the binary is rebuilt via `make build`, the embedded template FS **shall not** carry the db scaffold or `db.yaml`.

### B.4 그룹 4 — 지침 제거 (로컬 + 템플릿 미러, 동일 커밋)

- **REQ-DBR-013** (Ubiquitous): Both the local `.claude/skills/moai/workflows/project/doc-generation.md` and its template mirror **shall** be free of the DB Detection phase, the DB detection-keyword reference block, the `db-schema-sync` / `db.yaml` / `detected_db` references, and the completion-phase DB branch (Branch A/B).
- **REQ-DBR-014** (Ubiquitous): Both the local `.claude/skills/moai/workflows/sync/quality-gates-context.md` and its template mirror **shall** be free of the "DB Schema Doc Check" phase and its `db-schema-sync` / `db.yaml` references.
- **REQ-DBR-015** (Ubiquitous): The template copies of the two guidance files **shall** remain byte-identical to their local copies **and shall not** cite the `SPEC-DB-RETIRE-001` identifier or any internal work-tracking token (§25 template neutrality).

### B.5 그룹 5 — 로컬 dogfood 제거

- **REQ-DBR-016** (Unwanted): The working tree **shall not** track `.moai/project/db/` (7 files) nor `.moai/config/sections/db.yaml` (removed via `git rm` in run-phase, owned by the orchestrator).

### B.6 그룹 6 — 배포 사용자 처리 (능동 삭제 + preserve-root 보존) + CHANGELOG

- **REQ-DBR-017** (State-driven): **While** an existing deployed user project carries a `.moai/project/db/` scaffold under the `.moai/project` preserve root (실측 `internal/cli/update_preserve_inventory.go:68` — preserve root 이므로 `moai update`가 삭제 불가), the removal **shall not** delete those docs; a CHANGELOG entry **shall** provide manual-deletion guidance for `.moai/project/db/`. (`.moai/config/sections/db.yaml`의 능동 삭제는 REQ-DBR-022가 담당 — preserve root가 아니므로 update가 제거함.)

### B.8 그룹 8 — db.yaml 능동 정리 (DeprecatedPaths, clarification RESOLVED)

> 사용자 결정(2026-07-25): `moai update` 가 배포 사용자의 orphaned `.moai/config/sections/db.yaml` 을 **능동 삭제**한다.

- **REQ-DBR-022** (Event-driven): **When** `moai update` (clean-reinstall path) runs on a user project carrying an orphaned `.moai/config/sections/db.yaml`, the update **shall** remove it via the `defs.DeprecatedPaths` mechanism (`internal/defs/dirs.go` `var DeprecatedPaths`; consumed by `scanDeprecatedPaths` at `internal/cli/update_cleanup.go`). The removal **shall** register a `DeprecatedPathEntry{Path: ".moai/config/sections/db.yaml", DeprecatedBy: "SPEC-DB-RETIRE-001", ...}` and **shall** update `internal/defs/dirs_test.go` counts atomically (total 39→40; the category-derivation comments — a D3 finding — are corrected in the same edit). **While** the template still ships `db.yaml`, this registration **shall not** be added (the `TestDeprecatedPaths_NoTemplateCollision` guard at `internal/cli/deprecated_paths_collision_test.go` requires the template `db.yaml` deletion of REQ-DBR-011 to land first — 순서 의존).

### B.7 그룹 7 — 제거 안전성 / 빌드 게이트

- **REQ-DBR-018** (Ubiquitous): After removal, `go build ./...` **shall** exit 0.
- **REQ-DBR-019** (Ubiquitous): After removal, `go test ./...` **shall** exit 0.
- **REQ-DBR-020** (Event-driven): **When** `make build` runs post-removal, it **shall** complete cleanly.
- **REQ-DBR-021** (Ubiquitous): No package outside the deleted set (`internal/cli/hook.go` and the deleted `internal/hook/dbsync/` test files) **shall** import `dbsync` symbols — verified before deletion.

## §C. 스코프 제외 (Exclusions)

이 SPEC의 스코프에 포함되지 않는 것 (out of scope). 각 항목은 별도 위치로 귀속되거나 의도적으로 보존된다.

### Out of Scope — MCP 백엔드-DB 프로비저닝

- `doc-generation.md` Phase 11 (MCP Server Provisioning)의 `backend-db` / "DB server" 참조는 MCP 서버 추천 로직이며 db 문서화 서브시스템과 별개다 — 보존한다.
- `doc-generation.md` Phase 8 interview 스키마의 `external_systems` (DB / APIs / services) 필드는 인터뷰 축이며 제거 대상이 아니다 — 보존한다.

### Out of Scope — 배포 준비 마이그레이션 점검

- `quality-gates-context.md` Phase 3 "Deployment Readiness Check"의 "Migration Check"(database schema changes / migrations detected / migrations_needed) 는 배포 준비 판정 로직이며 db-schema-sync 서브시스템과 무관하다 — 보존한다.

### Out of Scope — settings 패키지 테스트 픽스처

- `internal/settings/testdata/sections/db.yaml` 는 `internal/settings` 테스트 픽스처(yamlpatch 형상 검증용)로, DB 문서화 서브시스템의 산출물이 아니다 — 보존한다. 제거 시 settings 테스트가 깨질 수 있다.

### Out of Scope — 배포 사용자 프로젝트의 `.moai/project/db/` 문서 자동 삭제

- 이미 배포된 사용자 프로젝트의 `.moai/project/db/` 스캐폴드는 **preserve 루트**(`.moai/project`, 실측 `update_preserve_inventory.go:68`)이므로 `moai update`가 구조적으로 삭제할 수 없다 — 의도적 잔존, CHANGELOG 수동 삭제 안내로 대체(REQ-DBR-017). **주의**: `.moai/config/sections/db.yaml`은 preserve 루트가 **아니므로** 능동 삭제 대상이며 이 제외에 해당하지 않는다(REQ-DBR-022, 스코프 내).

### Out of Scope — `session.PhaseDB` (`/moai db` 슬래시 커맨드 잔재)

- `internal/session/phase.go:14` 의 `PhaseDB Phase = "db"` (참조: `phase.go:27` `Valid()` switch, `phase_test.go:18,34`) 는 **이미 retired 된 `/moai db` 슬래시 커맨드**(Bundle A, 2026-05-16)의 세션-phase 열거 잔재이며, 본 SPEC이 제거하는 DB **문서화** 서브시스템(db-schema-sync 훅 / db.yaml / `.moai/project/db/`)과 무관하다. 죽은 심볼일 가능성이 있으나 세션-phase 인프라 소관이므로 본 SPEC에서 건드리지 않는다 — 별도 후속 정리 소관.

### Out of Scope — DB 문서화 기능의 재설계 / 재구현

- 게이팅·트리거 배선·파서 구현·라이터 구현·proposal→approval→rewrite 흐름 중 어떤 것도 이 SPEC에서 구현하지 않는다. 이 SPEC은 순수 제거(removal-only)이며 대체 기능을 만들지 않는다.

## §D. 성공 기준 (Success Criteria)

- 22개 REQ 전부가 acceptance.md의 기계 검증 가능 AC로 매핑된다.
- `grep -rn "db-schema-sync\|dbsync\|db\.enabled" internal/ cmd/ | grep -v "internal/settings/"` 가 (OOS-보존 `internal/settings/` 및 이 SPEC 문서 제외하고) 0건.
- `internal/hook/dbsync/` 및 두 템플릿 db 자산이 부재.
- `.moai/config/sections/db.yaml` 이 `defs.DeprecatedPaths`에 등록되어 `moai update`가 배포 사용자 사본을 능동 삭제.
- `go build ./...` / `go test ./...` 종료코드 0, `make build` clean.
- CHANGELOG에 (a) `db.yaml`은 update가 자동 삭제(수동 조치 불필요), (b) `.moai/project/db/` 문서는 수동 삭제 안내 — 두 문구 존재.
- 두 지침 파일의 로컬-템플릿 byte-parity 유지.
