---
id: SPEC-AGENCY-ABSORB-001
version: 0.2.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-20
author: GOOS
priority: High
labels: [agency, migration, design, hybrid, absorption]
---

# SPEC-AGENCY-ABSORB-001: Implementation Plan

## HISTORY

- 2026-04-20 v0.1.0: 최초 작성
- 2026-04-20 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정. YAML frontmatter 필수 필드 추가(D2), REQ 개수 62개로 재집계, M2에 REQ-MIGRATE-012a/b 플랫폼 분기 및 REQ-MIGRATE-013 SIGINT 처리 반영, M6 스코프 명확화(M5에서 사용자 `.claude/rules/agency/` 조건부 삭제).

---

## 1. 전체 전략 (High-Level Strategy)

Agency 흡수는 **5개 마일스톤(M1–M5)**로 나눠 진행한다. 순서는 "기반 마련 → 변환 → 제거 → 통합 → 정리". 각 마일스톤은 독립적으로 롤백 가능하도록 설계하며, 마일스톤 간 의존은 단방향.

### 원칙
- **No Big Bang**: 에이전트/파일 삭제는 맨 마지막(M5)에서만 수행
- **Template-First**: `internal/template/templates/` 반영 후 `make build` → 로컬 동기화
- **TDD 필수**: `internal/cli/migrate_agency_test.go`는 구현 전 작성
- **Thin Command Pattern**: `/moai design`은 라우팅만 담당, 로직은 `.claude/skills/moai/workflows/design.md`에 위치
- **언어 중립성**: 템플릿 계층은 16개 언어 동등 취급 (CLAUDE.local.md §15)

---

## 2. 마일스톤 (Milestones)

### M1: 기반 산출물 템플릿화 (Foundation Templates)

**목적**: 신규 스킬/설정/constitution을 템플릿 계층에 배치하여 `moai init` 신규 사용자부터 자동 제공.

**REQ 커버리지**: REQ-SKILL-001 ~ REQ-SKILL-014, REQ-SKILL-012a, REQ-SKILL-015(화이트리스트 스키마만 M1에서 정의), REQ-CONST-001 ~ REQ-CONST-004, REQ-DIR-003, REQ-BRIEF-001 ~ REQ-BRIEF-003

**작업 항목**:
- `internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md` 작성
- `internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md` 작성
- `internal/template/templates/.claude/skills/moai-workflow-design-import/SKILL.md` 작성
- `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md` 작성
- `internal/template/templates/.claude/skills/moai/workflows/design.md` 작성 (서브커맨드 로직)
- `internal/template/templates/.moai/project/brand/brand-voice.md` 스켈레톤
- `internal/template/templates/.moai/project/brand/visual-identity.md` 스켈레톤
- `internal/template/templates/.moai/project/brand/target-audience.md` 스켈레톤
- `internal/template/templates/.moai/config/sections/design.yaml` (기존 `.agency/config.yaml` 스키마 참조)
- `internal/template/templates/.claude/rules/moai/design/constitution.md` (agency constitution 재배치)
- `manager-spec` 에이전트에 BRIEF 섹션 확장 패치
- `make build` 실행 및 embedded.go 갱신

**우선순위**: High  
**선후관계**: 없음 (진입점)

---

### M2: 마이그레이션 커맨드 구현 (Migration Command, TDD)

**목적**: `moai migrate agency` CLI를 TDD로 구현. 기존 사용자의 무손실 이전을 가장 중요한 가치로 제공.

**REQ 커버리지**: REQ-MIGRATE-001 ~ REQ-MIGRATE-011, REQ-MIGRATE-012a, REQ-MIGRATE-012b, REQ-MIGRATE-013, REQ-DIR-001, REQ-DIR-002

**작업 항목**:
1. **RED**: `internal/cli/migrate_agency_test.go` 작성
   - 성공 케이스: dry-run, 실행, force 모드
   - 실패 케이스: 원본 부재, 대상 존재(force 없음), archive 존재, 디스크 부족
   - 롤백 케이스: 중간 단계 실패 시 `.agency/` 복원
   - SIGINT/SIGTERM 케이스: 시그널 전송 후 체크포인트 저장, `--resume` 재개 검증
   - 플랫폼별 테스트 분기:
     - `internal/cli/migrate_agency_posix_test.go` (build tag `//go:build !windows`): `os.Chmod` 검증
     - `internal/cli/migrate_agency_windows_test.go` (build tag `//go:build windows`): no-op 검증 + stderr 메시지 확인
   - 매핑 검증: REQ-DIR-001 표 한 줄 한 줄 테스트
   - `tech-preferences.md` ↔ `tech.md` 병합 충돌 케이스
2. **GREEN**: `internal/cli/migrate_agency.go` 구현
   - 트랜잭션 로그 구조체 (`migrationTx`, `~/.moai/.migrate-tx-<timestamp>.json`)
   - 순서 보장 (복사 → 이동 → 변환)
   - 원자성: 실패 시 전체 롤백
   - 권한 보존: POSIX는 `os.Chmod` 적용, Windows는 no-op (build tag로 분기)
   - `signal.Notify`로 SIGINT/SIGTERM 수신, 체크포인트 flush
   - `--resume <tx-id>` 플래그 지원
3. **REFACTOR**: 공통 파일 연산 유틸 추출 (`internal/fsutil/atomic.go` 등 기존 활용 검토)

**커버리지 목표**: 90%+ (critical package)  
**우선순위**: High  
**선후관계**: M1 완료 후 시작

---

### M3: `/moai design` 라우팅 구현 (Route Integration)

**목적**: `/moai design` 서브커맨드가 하이브리드 경로(A/B)를 제공하고 AskUserQuestion으로 사용자 선택을 수집.

**REQ 커버리지**: REQ-ROUTE-001 ~ REQ-ROUTE-008, REQ-FALLBACK-001 ~ REQ-FALLBACK-003, REQ-DETECT-003, REQ-SKILL-015(bundle 버전 화이트리스트 파싱 로직)

**작업 항목**:
- `internal/template/templates/.claude/commands/moai.md` 또는 서브커맨드 라우팅 테이블 업데이트
- `.claude/skills/moai/workflows/design.md`에 경로 분기 로직 명세
- 브랜드 컨텍스트 존재 여부 체크 로직
- AskUserQuestion 옵션 스펙 (경로 A 첫 번째 + (권장) 접미사)
- `expert-frontend` 위임 프롬프트 템플릿
- GAN 루프 연계 (`moai-workflow-gan-loop` skill 호출)
- Claude Design 미가용 사용자 감지/안내 UX

**우선순위**: High  
**선후관계**: M1 완료 후 시작 (M2와 병렬 가능)

---

### M4: 감지 및 Deprecation 라이프사이클 (Detection + Deprecation)

**목적**: 기존 `.agency/` 사용자 자동 감지, `/agency` 래퍼 제공.

**REQ 커버리지**: REQ-DETECT-001 ~ REQ-DETECT-003, REQ-DEPRECATE-001 ~ REQ-DEPRECATE-004

**작업 항목**:
- `moai doctor` 체크리스트에 `.agency/` 감지 추가
- `moai update` 실행 시 마이그레이션 권장 메시지
- SessionStart 훅 (`internal/hook/session_start.go`)에 일회성 공지 추가
- `.claude/commands/agency.md`를 deprecation wrapper로 변환
  - thin command pattern 유지 (< 20 LOC)
  - stderr 경고 1회 출력
  - `/moai design`으로 위임
- 미지원 서브커맨드(`learn`, `evolve`) 오류 처리
- CHANGELOG.md에 deprecation 항목 명시 (REQ-DEPRECATE-004 위반 시 CI 차단)

**우선순위**: Medium  
**선후관계**: M3 완료 후 시작

---

### M5: Agency 에이전트/스킬 제거 (Removal + Final Cleanup)

**목적**: 중복 에이전트 물리적 제거, 참조 정리, CI 감사 도입.

**REQ 커버리지**: REQ-REMOVE-001 ~ REQ-REMOVE-004

**작업 항목**:
- `.claude/agents/agency/planner.md` 삭제
- `.claude/agents/agency/designer.md` 삭제
- `.claude/agents/agency/builder.md` 삭제
- `.claude/agents/agency/evaluator.md` 삭제
- `.claude/agents/agency/copywriter.md` 삭제 (M1에서 skill로 흡수됨 확인 후)
- `.claude/agents/agency/learner.md` 삭제 (M1에서 moai-workflow-research로 흡수됨 확인 후)
- `.claude/skills/agency-frontend-patterns/` → `moai-domain-frontend/`로 병합
- `CLAUDE.md` §4 Agent Catalog 업데이트 (Agency Agents 섹션 축소/제거)
- 템플릿에서도 동일하게 제거 후 `make build`
- `internal/template/commands_audit_test.go`에 agency 에이전트 참조 금지 검증 추가
- 저장소 전체 `grep` 감사 스크립트 (`scripts/audit-no-agency-refs.sh`)

**우선순위**: Medium  
**선후관계**: M1 ~ M4 모두 완료 후

---

## 3. 기술적 접근 (Technical Approach)

### 3.1 마이그레이션 원자성 설계

```
Phase 1 (backup): .agency/ → .agency.archived/
Phase 2 (prepare): 대상 디렉터리 생성 (.moai/project/brand/, ...)
Phase 3 (transfer): 매핑표에 따라 복사/변환
Phase 4 (verify): 체크섬 비교 + 파일 개수 일치 확인
Phase 5 (rules-relocate): constitution.md 재배치
```

각 Phase는 트랜잭션 로그(`~/.moai/.migrate-tx-<timestamp>.json`)에 기록되고, 실패 시 역순으로 롤백.

### 3.2 `design.yaml` 스키마 변환

`.agency/config.yaml`의 `agency.*` 키를 `design.*`으로 재매핑. 예:
- `agency.gan_loop.*` → `design.gan_loop.*`
- `agency.evolution.*` → `design.evolution.*`
- `agency.pipeline.*` → `design.pipeline.*` (단, `planner`는 제거되어 `manager-spec` 위임으로 문서화)

### 3.3 Thin Command Pattern 유지

`/moai design`은 `.claude/commands/moai.md`의 서브커맨드 라우팅 테이블에 1줄 추가, 로직은 모두 `.claude/skills/moai/workflows/design.md`에 위치.

### 3.4 Claude Design handoff bundle 파서

- 1차 지원: ZIP (HTML + assets), HTML single-file
- 검증: zip-slip 공격 방지, 심볼릭 링크 거부, 실행 파일 탐지
- **버전 화이트리스트(REQ-SKILL-015)**: `.moai/config/sections/design.yaml`의 `supported_bundle_versions: ["v1", "v1.1"]` 키에서 읽어 비교. 불일치 시 `DESIGN_IMPORT_UNSUPPORTED_VERSION` 반환 및 경로 B 폴백 안내
- 출력: `.moai/design/tokens.json`, `.moai/design/components.json`
- 2차 로드맵: DOCX/PPTX/PDF/Canva 링크

### 3.5 GAN 루프 추출

`.claude/skills/moai-workflow-gan-loop/SKILL.md`는 기존 `.claude/rules/agency/constitution.md` Section 11/12 내용을 그대로 구현. Sprint Contract Protocol, Evaluator Leniency Prevention 5가지 메커니즘 포함.

### 3.6 언어 중립성 확인

`.moai/project/brand/` 템플릿은 16개 언어 동등 예시 제공 또는 언어 중립 스켈레톤만 포함. Go 코드 주석은 영어, SPEC 문서는 한글/영어 자유.

---

## 4. 에이전트 위임 계획 (Agent Delegation)

| 작업 단위 | 담당 에이전트 |
|---|---|
| SPEC 작성 (본 문서 포함) | `manager-spec` |
| M1 스킬/템플릿 작성 | `builder-skill` |
| M2 Go 구현 (TDD) | `manager-tdd` + `expert-backend` |
| M3 라우팅/스킬 | `builder-skill` |
| M4 훅/deprecation | `expert-backend` |
| M5 파일 제거/감사 | `manager-git` + `expert-refactoring` |
| 전체 리뷰 | `evaluator-active` |

---

## 5. 검증 전략 (Verification Strategy)

### 5.1 자동 검증 (CI)
- `go test ./...` 전체 통과
- `go test -race ./...` concurrency 검증
- `golangci-lint run`
- `internal/template/commands_audit_test.go` thin command 검증
- `scripts/audit-no-agency-refs.sh` 에이전트 참조 감사
- `make build` 성공

### 5.2 수동 검증 (QA)
- `.agency/` 샘플 프로젝트 마이그레이션 시연 (dry-run → 실제 → 롤백)
- `/moai design` 경로 A/B 실사용
- Pro 이하 사용자 시뮬레이션 (수동 플래그)
- Claude Design handoff bundle 실제 파일로 파싱 검증

### 5.3 회귀 방지
- `.agency/` 유지하는 기존 프로젝트 1개 이상을 fixture로 보존
- `internal/cli/testdata/agency_fixture/` 생성하여 마이그레이션 end-to-end 테스트

---

## 6. 리스크 및 완화 (Risks & Mitigations)

`spec.md` §8 참조. 추가로:

- **리스크**: M2 TDD 중 `os.Chmod`가 Windows에서 다르게 동작
  - **완화**: REQ-MIGRATE-012a(POSIX) / REQ-MIGRATE-012b(Windows) 플랫폼 분리, build tag로 테스트 분기, Windows CI 별도 검증 (lessons.md #7 참조)
- **리스크**: M2 TDD 중 SIGINT/SIGTERM이 마이그레이션 중간에 수신
  - **완화**: REQ-MIGRATE-013 체크포인트 트랜잭션 로그 + `--resume` 플래그로 재개
- **리스크**: M5 제거 시 사용자 프로젝트 `.claude/`에 복사본 잔존
  - **완화**: `moai update` 실행 시 자동 제거 대신 경고 출력 (사용자 자료 보호 원칙)
- **리스크**: Claude Design handoff bundle 스키마가 비공식
  - **완화**: REQ-SKILL-015 버전 감지 + `supported_bundle_versions` 화이트리스트 처리

---

## 7. 배포 계획 (Release Plan)

### Release vN (흡수 릴리스)
- M1 ~ M4 완료
- `/agency`는 deprecation wrapper로 유지
- CHANGELOG에 주요 경고 + 마이그레이션 가이드 링크

### Release vN+1
- M5 완료 (agency 파일 제거)
- Constitution FROZEN zone 검증 통과

### Release vN+2
- `/agency` 래퍼 완전 제거
- `.agency.archived/`는 사용자 결정에 따라 유지/삭제

---

## 8. 성공 지표 (Success Metrics)

- 마이그레이션 성공률 100% (테스트 픽스처 기준)
- `/moai design` 경로 A/B 분기 정상 동작 (수동 QA)
- agency 에이전트 참조 제거 이후 `grep` 결과 0건
- 테스트 커버리지: `internal/cli/migrate_agency.go` 90%+
- `make build` 실행 시간 회귀 없음 (±10% 이내)

---

## 9. Out-of-Scope (다시 한 번 확인)

`spec.md` §2.2와 동일. 특히:
- Claude Design과의 자동 통합 (API 미공개)
- `.agency.archived/` 자동 삭제
- 자가진화 Learner의 Go 포팅
- Figma 플러그인 직접 구현 (Phase 2 로드맵)

---

End of PLAN.
