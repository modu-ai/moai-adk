---
id: SPEC-DB-RETIRE-001
title: "DB 문서화 서브시스템 전면 제거 — 구현 계획"
version: "0.1.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
---

# SPEC-DB-RETIRE-001 — 구현 계획 (plan.md)

## §A. 컨텍스트

순수 제거(removal-only) SPEC. 새 기능 로직 없음. development_mode = `tdd`(실측: `.moai/config/sections/quality.yaml:2`) 이므로 run-phase는 **제거 안전성 검증**(no-importer 사전 확인 → `go build` → `go test` → `make build`)을 RED-GREEN의 등가물로 사용한다. 대부분의 변경은 파일 삭제이고, 핵심 난이도는 6~7개 Go 참조 지점의 정밀 해제와 2개 지침 파일의 서술 편집(로컬+템플릿 미러)에 있다.

## §B. 실측 검증된 코드 위치 (Known Issues / Verified Locations)

> [HARD] 보고서의 라인 번호를 그대로 신뢰하지 않고 콘텐츠 앵커(함수명·문자열 리터럴)로 재-grep 해 확정했다. 현재 브랜치는 `feat/SPEC-AGENT-PARALLEL-OPT-001`(공유 체크아웃)이며 라인은 드리프트할 수 있으므로, run-phase는 아래 **콘텐츠 앵커**로 재확인 후 편집한다.

### B.1 그룹 1 — Go: dbsync + CLI 배선 (REQ-DBR-001..004)

| 대상 | 위치(실측 2026-07-25) | 콘텐츠 앵커 | 조치 |
|------|----------------------|-------------|------|
| dbsync 패키지 | `internal/hook/dbsync/` (3파일: `db_schema_sync.go`, `db_schema_sync_test.go`, `db_schema_sync_internal_test.go`) | `package dbsync` | 디렉터리 통삭제 |
| 서브커맨드 등록 | `internal/cli/hook.go:136-144` | `Use:   "db-schema-sync"` / `dbSchemaSyncCmd` | 블록 삭제 |
| import | `internal/cli/hook.go:27` | `"github.com/modu-ai/moai-adk/internal/hook/dbsync"` | import 삭제(미삭제 시 컴파일 에러) |
| 핸들러 | `internal/cli/hook.go:409-445` | `func runDBSchemaSync` | 함수 삭제 |
| 기본 패턴 | `internal/cli/hook.go:501-509` | `var defaultMigrationPatterns` | 변수 삭제 |
| 로더 | `internal/cli/hook.go:511-548` | `func loadMigrationPatterns` | 함수 삭제 |
| 죽은 헬퍼 | `internal/cli/hook.go:550-575` | `func splitLines` / `func trimSpace` | **삭제 필수** — 유일 호출자가 `loadMigrationPatterns`(실측: grep 결과 두 함수는 `loadMigrationPatterns` 내부에서만 사용). 방치 시 golangci-lint `unused` 실패 |
| 공유 헬퍼 주석 | `internal/cli/hook.go:666-667` | `resolveHookProjectRoot`의 "...+ db-schema-sync + harness-classify handlers..." | 주석에서 `db-schema-sync` 문구만 제거(**함수 자체는 보존** — 다수 핸들러 공유) |
| e2e 유틸 맵 | `internal/cli/hook_e2e_test.go:363` | `"db-schema-sync":   true,` | 항목 삭제(잔존해도 기계적 통과하나 스코프상 제거) |

주의: `resolveHookProjectRoot`(hook.go:676)는 8개 이상 핸들러가 공유(실측 grep 12곳) — 제거 금지, 주석만 정정.

### B.2 그룹 2 — 참조 해제(정밀, 통삭제 금지) (REQ-DBR-005..009)

| 대상 | 위치 | 성격 | 조치 |
|------|------|------|------|
| `internal/defs/dirs.go:44-47, 238-241` | 주석만 | db.yaml을 `loadMigrationPatterns`가 읽는 live config로 서술 | **주석만** 정정(REQ-005 범위). 현재 baseline `DeprecatedPaths` = **39**(실측 `dirs_test.go:32` `const want = 39`, `len` 39); db.yaml은 아직 슬라이스 엔트리 아님 → 주석 편집 자체는 count 무변. **슬라이스 엔트리 추가 + count 39→40 + `dirs_test.go` 원자 갱신은 REQ-022(M2.5) 소관** — REQ-005와 비중복. 종전 "40 불변" 기재는 오류(40은 CONFIG-AUDIT-REPAIR 이전 값)로 정정 |
| `internal/cli/v2_detection.go:33-34` | 주석만 | "un-deprecated design.yaml + db.yaml" 서술 | 주석에서 db.yaml 언급 정정. 검출 로직은 `defs.DeprecatedPaths` 사용 → 동작 무변화 |
| `internal/config/audit_registry.go:74` | `yamlAuditExceptions["db"]` 실제 항목 | db.yaml을 hook line-scan 소비로 등록 | 항목 삭제. **순서 의존**: 로컬 db.yaml 삭제(REQ-016)와 함께(`TestAuditParity`는 real tree `.moai/config/sections` 스캔, 실측 audit_test.go:28). `TestAuditParity_ExceptionsRespected`의 `expectedExceptions`={constitution,context,interview,design,harness} 에 "db" 없음(실측 audit_test.go:156) → 삭제 안전 |
| **[D2] `internal/config/audit_loader_completeness_test.go:15-16`** | `acknowledgedUnloadedSections` slice의 `"db"` 항목 + `hook.go migration_patterns` 인용 주석 | yamlAuditExceptions의 sibling(template-tree 감사) | 항목 삭제. **순서 의존**: 템플릿 db.yaml 삭제(REQ-011)와 함께(`TestAuditLoaderCompleteness`는 **template** sections dir 스캔, 실측 audit_loader_completeness_test.go:66-69). 템플릿 db.yaml이 남아있는데 "db"만 지우면 `YAML_SECTION_NO_LOADER: db` 실패 → 삭제 순서 필수. REQ-DBR-007 확장으로 편입 |
| `internal/template/internal_content_leak_test.go:980` | 정규식 테스트 문자열 | `internal/hook/dbsync/db_schema_sync.go`를 "real restricted-package path"로 사용 | 실존 경로(예: `internal/hook/security/scan.go` 등)로 교체. **기계적으로는 비파괴**(문자열 정규식 매칭, 파일 stat 안 함)이나 삭제된 경로 참조라 정정 |
| `internal/template/settings_test.go:1047-1055` | `TestRender_DbSchemaChangeHook_Removed` 주석 | "internal/hook/dbsync 패키지와 CLI가 intact 유지" 서술 | 주석 정정. **테스트 본문은 dbsync 심볼 미참조**(settings.json 렌더 문자열 검사만) → 삭제 후에도 컴파일·통과 |

### B.3 그룹 3 — 템플릿 (REQ-DBR-010..012)

- `internal/template/templates/.moai/project/db/` — 7파일(실측: `README.md`, `schema.md`, `erd.mmd`, `migrations.md`, `queries.md`, `rls-policies.md`, `seed-data.md`) 통삭제.
- `internal/template/templates/.moai/config/sections/db.yaml` — 삭제.
- 삭제 후 `make build` 필요(embed FS 재컴파일). **run-phase 조치**(이 plan 단계에서는 실행 금지).

### B.4 그룹 4 — 지침 (로컬 + 템플릿 미러, byte-parity) (REQ-DBR-013..015)

두 파일은 로컬-템플릿 **byte-identical**(실측 `diff -q` → IDENTICAL). rule/skill 미러-parity 자동 테스트에는 **미등록**(실측: `rule_template_mirror_test.go`는 explicit 리스트 방식, 두 파일 부재) → 테스트 강제는 없으나 §2 Template-First 규율상 동일 편집 필수. 편집 후 `diff -q`로 byte-parity 수동 검증.

`doc-generation.md`(로컬 + `internal/template/templates/.claude/.../doc-generation.md`) 제거 대상:
- frontmatter `description`(L2)·`phase`(L6)의 "DB detection" / "4.1a" 표기.
- "Detection Keywords Reference" 절(L292-321, Phase 13 전용 지원 블록).
- "Phase 13: DB Detection" 전체(L321-370).
- "Phase 14: Completion" Step 4.2의 3-way 분기(Branch A/B는 DB, Branch C는 non-DB). → `detected_db` 기반 분기를 제거하고 non-DB 옵션(Create SPEC / Review and Edit / Generate harness / Done)을 무조건 제시.
- "Agent Chain Summary"의 Phase 13 항목(L456).
- **보존**(§C 제외): Phase 11 MCP `backend-db` 참조, Phase 8 `external_systems` 필드.

`quality-gates-context.md`(로컬 + 템플릿) 제거 대상:
- "Phase 2: DB Schema Doc Check (Conditional)" 전체(L152-200, Steps 0.08.1~0.08.4).
- **보존**: Phase 3 "Deployment Readiness Check"의 Migration Check(L215-240, 배포 준비 판정 — db-schema-sync 무관).

Phase 번호 재부여(gap vs renumber)는 run-phase 판단. 권장 기본: gap 없이 후속 Phase 재번호 시 다른 파일의 "Phase N" 교차참조 cascade를 grep 확인 후 결정. 단순 안전책은 DB Phase 내용만 제거하고 번호 gap 허용.

[HARD] 템플릿 지침 편집은 `SPEC-DB-RETIRE-001` 식별자나 내부 작업 토큰을 **인용하지 않는다**(§25 neutrality). 제거 사유 주석을 템플릿에 남기지 말 것.

### B.5 그룹 5 — 로컬 dogfood (REQ-DBR-016)

- `.moai/project/db/`(7파일, git-tracked 실측) + `.moai/config/sections/db.yaml`(git-tracked 실측) → run-phase에 `git rm`(오케스트레이터 소유).

### B.6 그룹 6 — CHANGELOG (REQ-DBR-017)

- `CHANGELOG.md`(존재 실측)에 제거 항목 + 배포 사용자 수동 삭제 안내(`.moai/project/db/` 및 `.moai/config/sections/db.yaml`).

### B.7 그룹 7 — 제거 안전성 (REQ-DBR-018..021)

- dbsync 실제 import는 `internal/cli/hook.go:27` 1곳 뿐(실측: settings_test.go/internal_content_leak_test.go의 매칭은 import가 아닌 문자열 참조; `import` 블록 grep으로 확인). 패키지 자체 테스트 2파일은 패키지와 함께 삭제.

### B.8 그룹 8 — db.yaml 능동 정리 (DeprecatedPaths, REQ-DBR-022) — clarification RESOLVED

실측 검증된 메커니즘:

| 요소 | 위치(실측 2026-07-25) | 조치 |
|------|----------------------|------|
| SSOT 슬라이스 | `internal/defs/dirs.go:61` `var DeprecatedPaths = []DeprecatedPathEntry{...}` | `.moai/config/sections/db.yaml` 엔트리 추가(`DeprecatedBy: "SPEC-DB-RETIRE-001"`; dirs.go는 템플릿 아님 → §25 무관, SPEC-ID 인용 허용, 기존 엔트리 관례와 일치) |
| 엔트리 구조 | `DeprecatedPathEntry{Path, DeprecatedSince, DeprecatedBy, RemovalSchedule}` (샘플: `dirs.go:243,249`) | 4필드 전부 채움(`TestDeprecatedPathsRequiredFields`) |
| 삭제 소비자 | `internal/cli/update_cleanup.go:126` `scanDeprecatedPaths(projectRoot)` → `internal/cli/update_clean_install.go` Step 4 REMOVE | 자동 소비(추가 배선 불요 — 슬래시 등록만으로 clean-reinstall이 제거) |
| **충돌 가드(순서 결정)** | `internal/cli/deprecated_paths_collision_test.go:53` `TestDeprecatedPaths_NoTemplateCollision` — DeprecatedPaths ∩ 템플릿 FS = ∅ 강제 | **템플릿 db.yaml 삭제(REQ-011)가 반드시 선행**. 안 하면 `db.yaml`이 슬라이스+템플릿 양쪽 → 충돌 가드 빌드 실패(#1084 재발 방지) |
| 카운트 테스트(D3 포함) | `internal/defs/dirs_test.go` — content-anchor `const want = 39`(실측 L32), `wantCategoryB = 27`(실측 L54), `TestDeprecatedPathsCategoryBExpectedEntries` 목록, 파생 주석(L5-11) — 라인 드리프트 시 content-anchor 우선 | 원자적 갱신: total 39→40, Category B 27→28(`DeprecatedSince` 값에 따라), 기대-엔트리 목록에 db.yaml 추가, 파생 주석 정정(**D3 = 여기 흡수**). @MX:ANCHOR(`SSOT for v.2.x → v3 cleanup targets`, dirs.go:53)가 "슬라이스+테스트 원자 갱신" 강제 |

**카테고리 선택 주의**: `TestDeprecatedPathsCategorySplit`은 3개 `DeprecatedSince` 버킷(AGENCY-ABSORB / V2-V3-CLEAN-REINSTALL / AGENT-FOLDER-SPLIT)만 카운트한다. db.yaml에 새 `DeprecatedSince`(예: `SPEC-DB-RETIRE-001`)를 쓰면 4번째 버킷이 되어 split 테스트가 total과 불일치할 수 있으므로, run-phase는 (a) 기존 버킷 중 하나 재사용 vs (b) split 테스트에 새 버킷 추가 중 택1 후 dirs_test.go를 정합화한다. 권장: `DeprecatedSince: "SPEC-DB-RETIRE-001"` + split 테스트에 4번째 버킷 케이스 추가(가장 정직).

## §C. 사전 점검 (Pre-flight)

1. **제거 안전성 사전 확인(RED 등가)**: `grep -rn '"github.com/modu-ai/moai-adk/internal/hook/dbsync"' internal/ cmd/ pkg/` → `internal/cli/hook.go` 1곳만 나와야 함(패키지 자체 테스트 제외). 다른 importer 발견 시 blocker 보고.
2. 콘텐츠 앵커 재-grep으로 현재 라인 확정(§B 표의 앵커 사용).
3. `.moai/project/db/` / `db.yaml` git-tracked 상태 재확인(`git ls-files`).
4. 두 지침 파일 로컬-템플릿 byte-parity 사전 확인(`diff -q`).
5. **순서 의존 3건 확인(빌드/테스트 파손 방지)**:
   - (a) 템플릿 `db.yaml` 삭제(REQ-011) → **선행 필수** → 그 후 DeprecatedPaths 등록(REQ-022, 충돌 가드) **및** `acknowledgedUnloadedSections["db"]` 삭제(REQ-007, loader-completeness는 template 스캔).
   - (b) 로컬 `db.yaml` 삭제(REQ-016) → 그와 함께 `yamlAuditExceptions["db"]` 삭제(REQ-007, TestAuditParity는 real tree 스캔).
   - (c) 한쪽만 하면: 템플릿 db.yaml 잔존+"db" allowlist 삭제 → `YAML_SECTION_NO_LOADER` 실패 / db.yaml 슬라이스+템플릿 공존 → 충돌 가드 실패.
6. **closure grep 사전 검산(D1)**: `grep -rn "db-schema-sync\|dbsync\|db\.enabled" internal/ cmd/ | grep -v "internal/settings/"` 를 현재 트리에서 실행 → 매칭이 전부 제거 대상 집합(hook.go/dbsync/템플릿/e2e/leak-test/settings-comment)임을 확인. `internal/settings/schema_sections_test.go:292`의 `"db.enabled"`(OOS key-path 테스트)는 `internal/settings/` 제외로 걸러짐을 확인.

## §D. 제약 (Constraints)

- **race-safety**: 공유 체크아웃 병렬 세션 존재 가능 — run-phase는 pre-spawn sync 점검 후 진행(오케스트레이터 소유). **plan-phase(현재)는 SPEC 디렉터리 밖 파일을 절대 건드리지 않음**.
- **정밀 편집 원칙**: 그룹 2 참조는 통삭제 금지 — 주변 로직·주석 문맥 보존하며 db 참조만 제거.
- **byte-parity**: 그룹 4 로컬/템플릿 동일 편집.
- **template neutrality(§25)**: 템플릿 편집에 내부 식별자 인용 금지.
- **공유 헬퍼 보존**: `resolveHookProjectRoot`는 제거 금지(다수 핸들러 공유).
- **Go build 규율**: import 제거 + 죽은 헬퍼(`splitLines`/`trimSpace`) 제거를 dbsync 삭제와 원자적으로.

## §E. 자체 검증 (Self-Verification, plan-phase audit-ready signal)

- [x] 22개 REQ 전부 acceptance.md의 ≥1 AC로 매핑(요구사항→AC 커버리지 22/22, 총 24 AC).
- [x] 모든 코드 위치를 콘텐츠 앵커로 실측 재확인(DeprecatedPaths·audit_loader·PhaseDB·schema_sections_test 추가 실측 포함).
- [x] 보고서 스코프 밖 추가 참조 발견·편입: audit_registry.go, internal_content_leak_test.go, settings_test.go 주석, hook.go 공유 헬퍼 주석, 죽은 헬퍼 splitLines/trimSpace, **audit_loader_completeness_test.go(D2)**, **dirs_test.go 파생 주석(D3)**.
- [x] plan-auditor D1-D5 반영: D1(closure grep→`internal/settings/` 제외), D2(acknowledgedUnloadedSections["db"] 편입), D3(dirs_test 주석→DeprecatedPaths REQ 흡수), D4(AC-017 grep 앵커화), D5(PhaseDB→명시 OOS).
- [x] clarification RESOLVED → REQ-DBR-022 능동 삭제 신설, 미해소 clarification 마커(콜론-정규형) **0건**.
- [x] out-of-scope 5항목 명시(MCP backend-db, deployment migration check, settings testdata db.yaml, `.moai/project/db/` 문서(preserve-root 잔존), PhaseDB 잔재).
- [x] 순서 의존 3건 명시(template-first → collision guard + loader-completeness; local-first → audit parity).
- [x] 제거 안전성: dbsync 실제 importer 1곳(hook.go)만 실측 확인.
- [x] SPEC ID pre-write self-check PASS (`SPEC ✓ | DB ✓ | RETIRE ✓ | 001 ✓ → PASS`).

## §F. 마일스톤 (변경 가능성 내림차순 — 리뷰 집중도 순)

> 리뷰 초점을 가장 바뀔 가능성이 높은 결정에 두기 위해, 사용자-노출 지침 편집(가장 주관적·가역적)을 먼저, 기계적 삭제(가장 확정적)를 뒤로 배치. run-phase 실행 순서는 각 마일스톤의 "실행 의존" 주석 참조.

### M1 — 지침 콘텐츠 재설계 (사용자-노출, 최고 변경가능성) [REQ-DBR-013..015]
- `doc-generation.md` DB Detection phase + detection-keyword 블록 + completion 분기 제거, non-DB 옵션 무조건화.
- `quality-gates-context.md` "DB Schema Doc Check" phase 제거.
- 로컬 + 템플릿 미러 동일 편집 → `diff -q` byte-parity 검증. 템플릿에 내부 식별자 인용 금지(§25).
- Phase 재번호 여부 결정(권장: cascade grep 후 판단).

### M2 — 참조 해제 (정밀 편집, 판단 필요) [REQ-DBR-005..009]
- `dirs.go` / `v2_detection.go` 주석 정정(주석-only db.yaml 참조; DeprecatedPaths 슬라이스 count는 M2.5에서 변경).
- `audit_registry.go` `yamlAuditExceptions["db"]` 삭제(로컬 db.yaml 삭제와 조율 — real tree).
- **`audit_loader_completeness_test.go` `acknowledgedUnloadedSections["db"]` 삭제(D2)**(템플릿 db.yaml 삭제와 조율 — template tree).
- `internal_content_leak_test.go` 실존 경로로 교체, `settings_test.go` 주석 정정.

### M2.5 — db.yaml 능동 삭제 등록 (사용자-노출 결정, DeprecatedPaths) [REQ-DBR-022]
- 실행 의존: **M3의 템플릿 db.yaml 삭제 이후**(충돌 가드 `TestDeprecatedPaths_NoTemplateCollision`).
- `internal/defs/dirs.go` `DeprecatedPaths` 에 `.moai/config/sections/db.yaml` 엔트리 추가(`DeprecatedBy: "SPEC-DB-RETIRE-001"`).
- `internal/defs/dirs_test.go` 원자 갱신: total 39→40, Category split(4번째 버킷 or 기존 재사용), 기대-엔트리 목록, 파생 주석(D3 흡수).
- 검증: clean-reinstall 테스트로 seed된 db.yaml 제거 확인.

### M3 — 기계적 제거 (확정적) [REQ-DBR-001..004, 010, 011, 016]
- 실행 의존: **M3 착수 전 §C 제거 안전성 사전 확인 통과 필수**.
- `internal/hook/dbsync/` 통삭제.
- `hook.go`: import(27) + 서브커맨드(136-144) + `runDBSchemaSync` + `defaultMigrationPatterns` + `loadMigrationPatterns` + 죽은 헬퍼 `splitLines`/`trimSpace` + 공유 헬퍼 주석 정정.
- `hook_e2e_test.go` 유틸 맵 항목 삭제.
- 템플릿 db/ 7파일 + db.yaml 삭제.
- 로컬 dogfood `.moai/project/db/` + `db.yaml` `git rm`.

### M4 — 빌드/테스트 검증 게이트 (GREEN 등가) [REQ-DBR-012, 018..021]
- 실행 의존: M2·M3 완료 후.
- `go build ./...`(0) → `go test ./...`(0) → `make build`(clean) → 재-grep 잔여 0 확인.

### M5 — CHANGELOG [REQ-DBR-017]
- 제거 항목 + 배포 사용자 수동 삭제 안내 기재.

## §G. 안티패턴 (Anti-Patterns)

- **AP-1**: 그룹 2 참조를 통파일 삭제 → 주변 로직 손상. (정밀 편집만)
- **AP-2**: `resolveHookProjectRoot` 를 db-schema-sync 전용으로 오인해 삭제 → 다수 핸들러 붕괴.
- **AP-3**: `splitLines`/`trimSpace` 미삭제 → golangci-lint `unused` 실패(go build는 통과하나 lint 게이트 실패).
- **AP-4**: 템플릿에만 편집하고 로컬 미러 누락(또는 반대) → byte-parity 붕괴.
- **AP-5**: 템플릿 지침에 "removed by SPEC-DB-RETIRE-001" 주석 삽입 → §25 neutrality 위반.
- **AP-6**: 로컬 db.yaml만 삭제하고 audit_registry.go "db" 예외는 남김(또는 반대) → 순서 불일치 시 `TestAuditParity` orphan 실패 위험.
- **AP-7**: MCP `backend-db`(Phase 11) / deployment migration check(Phase 3) / settings testdata db.yaml / `session.PhaseDB` / schema_sections_test.go `db.enabled` key-path 를 db 문서화 서브시스템으로 오인 제거 → 무관 기능 회귀.
- **AP-8**: 보고서 라인 번호 verbatim 신뢰 → 드리프트로 오편집. (콘텐츠 앵커 재-grep 필수)
- **AP-9**: 템플릿 db.yaml 삭제 **전에** DeprecatedPaths에 db.yaml 등록 → `TestDeprecatedPaths_NoTemplateCollision` 빌드 실패(#1084 재발 가드). 반드시 template-first.
- **AP-10**: 템플릿 db.yaml 잔존 상태에서 `acknowledgedUnloadedSections["db"]` 삭제 → `YAML_SECTION_NO_LOADER: db` 실패. template-first.
- **AP-11**: DeprecatedPaths 슬라이스만 갱신하고 `dirs_test.go` count/버킷/기대-목록 미갱신 → 카운트 테스트 실패. @MX:ANCHOR 관례상 원자 갱신.
- **AP-12**: closure grep에서 `internal/settings/` 를 제외하지 않아 `schema_sections_test.go:292 "db.enabled"`(OOS 보존)이 false-fail → AC-DBR-022 오탐.

## §H. 교차 참조 (Cross-References)

- 근거 보고서: `.moai/reports/db-scaffold-sync-logic-20260725.md`
- 승인 기록: project memory `project_db_retire_removal_approved.md`
- 관련 SPEC: `SPEC-DB-SYNC-001`(원 구현), `SPEC-DB-SYNC-RELOC-001`(PostToolUse 훅 relocation — 이미 래퍼 제거됨), `SPEC-DEPRECATEDPATHS-RECONCILE-001`(db.yaml un-deprecate 이력)
- 템플릿 규율: CLAUDE.local.md §2(Template-First), §15(언어 중립성), §25(내부 콘텐츠 격리)

## §I. Clarification RESOLVED (열린 항목 0건)

- **[RESOLVED 2026-07-25] db.yaml DeprecatedPaths 능동 정리** — 사용자 결정: `moai update` 가 배포 사용자의 orphaned `.moai/config/sections/db.yaml` 을 **능동 삭제**한다(수동 잔존 아님). 반영: REQ-DBR-022 신설 + AC-DBR-023/024 + §B.8 메커니즘 위치 확정. `.moai/project/db/` **문서**는 별개 결정 — `.moai/project` 가 preserve 루트(실측 `update_preserve_inventory.go:68`)라 update가 구조적으로 삭제 불가 → 잔존 + CHANGELOG 수동 안내(REQ-DBR-017). 남은 미해소 clarification 항목(콜론-정규형 마커): **0건**.
