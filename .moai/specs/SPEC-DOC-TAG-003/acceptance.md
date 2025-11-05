# Acceptance Criteria: @SPEC:DOC-TAG-003

> **Phase 3: 배치 마이그레이션 - 33개 미태깅 파일 자동 TAG 생성**
>
> **목표**: 42.3% 갭 해소, 78/78 파일 완전 태깅 달성 (100%)

---

## 📋 Overview

Phase 3 배치 마이그레이션의 성공 여부는 다음 기준으로 판단합니다:

1. **완전성**: 33개 미태깅 파일 모두 TAG 삽입 완료
2. **품질**: TAG 포맷 표준 준수, 중복 없음, Chain 무결성
3. **안전성**: 백업 및 롤백 시스템 정상 작동
4. **추적성**: TAG 인벤토리 및 체인 참조 완전 추적
5. **사용자 경험**: 승인 모델 작동, 진행 상태 명확

---

## 🧪 Test Scenarios

### Batch 1: Quick Wins

#### Scenario 1.1: Quick Wins 파일 TAG 생성

**Given**: Phase 1/2 완료, 5개 Quick Wins 파일 미태깅 상태

**When**: `/alfred:3-sync` 실행, Batch 1 승인

**Then**:
- ✅ 5개 파일 모두 TAG 삽입 성공:
  - `CLAUDE-AGENTS-GUIDE.md` → `@DOC:GUIDE-AGENT-001`
  - `CLAUDE-PRACTICES.md` → `@DOC:GUIDE-PRACTICE-001`
  - `CLAUDE-RULES.md` → `@DOC:GUIDE-RULES-001`
  - `CHANGELOG.md` → `@DOC:STATUS-CHANGELOG-001`
  - `README.md` → `@DOC:STATUS-README-001`
- ✅ 백업 파일 5개 생성 (`.moai/backups/batch-1/`)
- ✅ TAG 인벤토리 업데이트 (50/78 files tagged)
- ✅ 신규 도메인 2개 등록 (`@DOC:GUIDE-*`, `@DOC:STATUS-*`)
- ✅ Chain 참조 포함 (`@SPEC:DOC-TAG-003 -> @DOC:GUIDE-AGENT-001`)

#### Scenario 1.2: Batch 1 거부 시나리오

**Given**: Phase 1.5에서 Batch 1 제안 표시

**When**: 사용자가 Batch 1 거부 (`AskUserQuestion` → "n")

**Then**:
- ✅ Batch 1 파일 수정 없음
- ✅ Batch 2 제안 표시
- ✅ TAG 인벤토리 변경 없음 (45/78 files tagged)

#### Scenario 1.3: README.md 백업 검증

**Given**: `README.md` 원본 내용 존재

**When**: Batch 1 실행, TAG 삽입 전 백업 생성

**Then**:
- ✅ `.moai/backups/batch-1/README.md.backup` 파일 존재
- ✅ 백업 파일 내용 = 원본 내용 (TAG 삽입 전)
- ✅ TAG 삽입 후 원본 파일 내용 ≠ 백업 파일 내용

---

### Batch 2: Skills System

#### Scenario 2.1: Foundation Skill TAG 생성

**Given**: Batch 1 완료, 5개 Foundation Skill 미태깅 상태

**When**: Batch 2 승인

**Then**:
- ✅ 5개 Foundation Skill 모두 TAG 삽입:
  - `moai-foundation-ears/SKILL.md` → `@DOC:SKILL-EARS-001`
  - `moai-foundation-specs/SKILL.md` → `@DOC:SKILL-SPECS-001`
  - `moai-foundation-tags/SKILL.md` → `@DOC:SKILL-TAGS-001`
  - `moai-foundation-trust/SKILL.md` → `@DOC:SKILL-TRUST-001`
  - `moai-foundation-hooks/SKILL.md` → `@DOC:SKILL-HOOKS-001`
- ✅ TAG 인벤토리 업데이트 (55/78 files tagged)
- ✅ 신규 도메인 1개 등록 (`@DOC:SKILL-*`)

#### Scenario 2.2: moai-foundation-tags 자기 참조

**Given**: `moai-foundation-tags/SKILL.md` TAG 삽입

**When**: TAG 삽입 후 파일 내용 검증

**Then**:
- ✅ 헤더에 `@DOC:SKILL-TAGS-001` 포함
- ✅ Chain 참조 포함 (`@SPEC:DOC-TAG-001 -> @DOC:SKILL-TAGS-001`)
- ✅ 파일 내용에 자기 참조 TAG 언급 가능 (순환 위험 없음)

---

### Batch 3: Architecture

#### Scenario 3.1: Architecture Skill TAG 생성 (SPEC 매핑 있음)

**Given**: Batch 2 완료, `@SPEC:PROJECT-001` 존재

**When**: Batch 3 승인

**Then**:
- ✅ 3개 Architecture Skill TAG 삽입:
  - `moai-foundation-structure/SKILL.md` → `@DOC:SKILL-STRUCTURE-001`
  - `moai-foundation-product/SKILL.md` → `@DOC:SKILL-PRODUCT-001`
  - `moai-foundation-tech/SKILL.md` → `@DOC:SKILL-TECH-001`
- ✅ Chain 참조 포함 (`@SPEC:PROJECT-001 -> @DOC:SKILL-STRUCTURE-001`)
- ✅ TAG 인벤토리 업데이트 (58/78 files tagged)

#### Scenario 3.2: Architecture Skill TAG 생성 (SPEC 매핑 없음)

**Given**: `@SPEC:PROJECT-001` 미존재 또는 신뢰도 < 0.5

**When**: Batch 3 실행

**Then**:
- ✅ TAG 삽입 (Chain 참조 생략):
  ```markdown
  # @DOC:SKILL-STRUCTURE-001

  # Project Structure Guide
  ```
- ✅ 사용자에게 수동 검토 요청 표시
- ✅ TAG 인벤토리 업데이트 (58/78 files tagged)

---

### Batch 4: Concepts

#### Scenario 4.1: 혼합 신뢰도 Skill TAG 생성

**Given**: Batch 3 완료, 5개 Concept Skill (HIGH + MEDIUM 신뢰도)

**When**: Batch 4 승인 (개별 검토 옵션 선택)

**Then**:
- ✅ 5개 Concept Skill 모두 TAG 삽입
- ✅ HIGH 신뢰도 파일: Chain 참조 포함
- ✅ MEDIUM 신뢰도 파일: 사용자 검토 후 Chain 결정
- ✅ TAG 인벤토리 업데이트 (63/78 files tagged)

#### Scenario 4.2: Skill Tier 일관성 검증

**Given**: Batch 4 완료

**When**: Skill Tier 일관성 검증 실행

**Then**:
- ✅ `moai-essentials-*` 파일: `@DOC:SKILL-*` 도메인 사용
- ✅ `moai-alfred-*` 파일: `@DOC:SKILL-*` 도메인 사용
- ✅ Tier 간 명명 규칙 일관성 유지

---

### Batch 5: Workflows

#### Scenario 5.1: Workflow Skill TAG 생성 (Chain 참조 포함)

**Given**: Batch 4 완료, 6개 Workflow Skill (대부분 HIGH 신뢰도)

**When**: Batch 5 승인

**Then**:
- ✅ 6개 Workflow Skill 모두 TAG 삽입
- ✅ 5개 파일 Chain 참조 포함:
  - `@SPEC:PLAN-001 -> @DOC:SKILL-PLAN-WF-001`
  - `@SPEC:RUN-001 -> @DOC:SKILL-RUN-WF-001`
  - `@SPEC:SYNC-001 -> @DOC:SKILL-SYNC-WF-001`
  - `@SPEC:PROJECT-001 -> @DOC:SKILL-PROJECT-WF-001`
  - `@SPEC:TRUST-001 -> @DOC:SKILL-TRUST-VAL-001`
- ✅ 1개 파일 Chain 생략 (`@DOC:SKILL-INTERACTIVE-001`, 신뢰도 MEDIUM)
- ✅ TAG 인벤토리 업데이트 (69/78 files tagged)

---

### Batch 6: Tutorials

#### Scenario 6.1: Language Skill TAG 생성 (LOW 신뢰도)

**Given**: Batch 5 완료, 3개 Language Skill (LOW 신뢰도)

**When**: Batch 6 승인 (수동 검토 권장)

**Then**:
- ✅ 3개 Language Skill TAG 삽입:
  - `@DOC:SKILL-KOREAN-001`
  - `@DOC:SKILL-JAPANESE-001`
  - `@DOC:SKILL-SPANISH-001`
- ✅ Chain 참조 생략 (신뢰도 < 0.5)
- ✅ 사용자 수동 검토 완료 표시

#### Scenario 6.2: Domain/Ops Skill TAG 생성

**Given**: Batch 5 완료, 4개 Domain/Ops Skill (MEDIUM 신뢰도)

**When**: Batch 6 승인

**Then**:
- ✅ 4개 Domain/Ops Skill TAG 삽입:
  - `@DOC:SKILL-PYTHON-001`
  - `@DOC:SKILL-TYPESCRIPT-001`
  - `@DOC:SKILL-GIT-001`
  - `@DOC:SKILL-CICD-001`
- ✅ TAG 인벤토리 업데이트 (76/78 files tagged)

#### Scenario 6.3: Skill Tier 전체 일관성 검증

**Given**: Batch 6 완료 (26개 Skill 모두 태깅)

**When**: Skill Tier 전체 검증 실행

**Then**:
- ✅ Foundation Tier (5개): `@DOC:SKILL-EARS-001` ~ `@DOC:SKILL-HOOKS-001`
- ✅ Architecture Tier (3개): `@DOC:SKILL-STRUCTURE-001` ~ `@DOC:SKILL-TECH-001`
- ✅ Essentials Tier (2개): `@DOC:SKILL-CONTEXT-001`, `@DOC:SKILL-WORKFLOW-001`
- ✅ Alfred Tier (9개): `@DOC:SKILL-EARS-AUTHOR-001` ~ `@DOC:SKILL-INTERACTIVE-001`
- ✅ Domain Tier (2개): `@DOC:SKILL-PYTHON-001`, `@DOC:SKILL-TYPESCRIPT-001`
- ✅ Language Tier (3개): `@DOC:SKILL-KOREAN-001` ~ `@DOC:SKILL-SPANISH-001`
- ✅ Ops Tier (2개): `@DOC:SKILL-GIT-001`, `@DOC:SKILL-CICD-001`

---

### Batch 7: Polish

#### Scenario 7.1: 프로젝트 메타 문서 TAG 생성 (최종 배치)

**Given**: Batch 6 완료, 2개 프로젝트 메타 문서 미태깅

**When**: Batch 7 승인 (최종 배치)

**Then**:
- ✅ 2개 프로젝트 문서 TAG 삽입:
  - `.moai/project/structure.md` → `@DOC:PROJECT-STRUCTURE-001`
  - `.moai/project/tech.md` → `@DOC:PROJECT-TECH-001`
- ✅ Chain 참조 포함 (`@SPEC:PROJECT-001 -> @DOC:PROJECT-STRUCTURE-001`)
- ✅ TAG 인벤토리 업데이트 (**78/78 files tagged - 100%**)

#### Scenario 7.2: Phase 3 최종 검증

**Given**: Batch 7 완료

**When**: Phase 3 최종 검증 실행

**Then**:
- ✅ **78/78 파일 모두 태깅 완료 (100%)**
- ✅ TAG ID 전역 고유성 검증 (중복 0개)
- ✅ 도메인 명명 규칙 일관성 검증:
  - `@DOC:GUIDE-*` (3개)
  - `@DOC:SKILL-*` (26개)
  - `@DOC:STATUS-*` (2개)
  - `@DOC:PROJECT-*` (2개)
- ✅ Chain 참조 전체 추적 가능
- ✅ TAG 인벤토리 최종 업데이트 (`.moai/memory/tag-registry.json`)
- ✅ 백업 파일 보관 (`.moai/backups/batch-1~7/`)

---

## ✅ Verification Checklist

### 배치별 검증 체크리스트

#### Batch 1: Quick Wins

- ✅ 5개 파일 TAG 삽입 성공
- ✅ TAG ID 형식: `@DOC:GUIDE-*`, `@DOC:STATUS-*`
- ✅ TAG 포맷 표준 준수 (헤더 첫 줄)
- ✅ Chain 참조 포함 (GUIDE 도메인만)
- ✅ 백업 파일 5개 생성
- ✅ TAG 인벤토리 업데이트 (50/78)
- ✅ 신규 도메인 2개 등록

#### Batch 2: Skills System

- ✅ 5개 Foundation Skill TAG 삽입
- ✅ TAG ID 형식: `@DOC:SKILL-*`
- ✅ `moai-foundation-tags` 자기 참조 TAG 포함
- ✅ 백업 파일 5개 생성
- ✅ TAG 인벤토리 업데이트 (55/78)
- ✅ 신규 도메인 1개 등록 (`@DOC:SKILL-*`)

#### Batch 3: Architecture

- ✅ 3개 Architecture Skill TAG 삽입
- ✅ SPEC 매핑 신뢰도 검증 (HIGH/MEDIUM)
- ✅ Chain 참조 포함 (신뢰도 ≥ 0.5)
- ✅ 백업 파일 3개 생성
- ✅ TAG 인벤토리 업데이트 (58/78)

#### Batch 4: Concepts

- ✅ 5개 Concept Skill TAG 삽입
- ✅ Skill Tier 일관성 (moai-essentials-*, moai-alfred-*)
- ✅ 혼합 신뢰도 처리 (HIGH + MEDIUM)
- ✅ 백업 파일 5개 생성
- ✅ TAG 인벤토리 업데이트 (63/78)

#### Batch 5: Workflows

- ✅ 6개 Workflow Skill TAG 삽입
- ✅ Chain 참조 5개 포함
- ✅ 워크플로우 도메인 명명 규칙 준수
- ✅ 백업 파일 6개 생성
- ✅ TAG 인벤토리 업데이트 (69/78)

#### Batch 6: Tutorials

- ✅ 7개 Tutorial Skill TAG 삽입
- ✅ Language Tier 일관성 (moai-language-*)
- ✅ Domain/Ops Tier 일관성 (moai-domain-*, moai-ops-*)
- ✅ LOW 신뢰도 처리 (수동 검토)
- ✅ 백업 파일 7개 생성
- ✅ TAG 인벤토리 업데이트 (76/78)

#### Batch 7: Polish

- ✅ 2개 프로젝트 문서 TAG 삽입
- ✅ Chain 참조 포함
- ✅ 백업 파일 2개 생성
- ✅ TAG 인벤토리 최종 업데이트 (**78/78**)

---

### Phase 3 전체 검증 체크리스트

#### 완전성 검증

- ✅ **78/78 파일 모두 태깅 완료 (100%)**
- ✅ 미태깅 파일 0개 (0%)
- ✅ 33개 신규 TAG 생성
- ✅ 7개 배치 모두 완료

#### 품질 검증

- ✅ TAG ID 전역 고유성 (중복 0개)
- ✅ TAG 포맷 표준 준수:
  ```markdown
  # @DOC:DOMAIN-NNN | Chain: @SPEC:SOURCE-ID -> @DOC:DOMAIN-NNN
  ```
- ✅ 도메인 명명 규칙 일관성:
  - `@DOC:GUIDE-TOPIC-NNN`
  - `@DOC:SKILL-SKILL_NAME-NNN`
  - `@DOC:STATUS-TYPE-NNN`
  - `@DOC:PROJECT-ASPECT-NNN`
- ✅ Chain 참조 무결성 (신뢰도 ≥ 0.5)

#### 안전성 검증

- ✅ 백업 파일 33개 생성 (`.moai/backups/batch-1~7/`)
- ✅ 롤백 가능 상태 유지
- ✅ 백업 파일 7일 보관 정책 적용

#### 추적성 검증

- ✅ TAG 인벤토리 최종 업데이트 (`.moai/memory/tag-registry.json`)
- ✅ TAG 체인 전체 추적 가능:
  ```
  @SPEC:DOC-TAG-001 → @DOC:SKILL-EARS-001
  @SPEC:DOC-TAG-002 → @DOC:SKILL-TAGS-001
  @SPEC:DOC-TAG-003 → @DOC:GUIDE-AGENT-001
  @SPEC:PROJECT-001 → @DOC:SKILL-STRUCTURE-001
  ```
- ✅ 도메인별 TAG 집계 정확:
  - GUIDE: 3개
  - SKILL: 26개
  - STATUS: 2개
  - PROJECT: 2개

---

## 🔒 Quality Gates

### Gate 1: 배치 실행 전 (Pre-execution)

**조건**:
- ✅ Phase 1/2 완료 확인
- ✅ 백업 시스템 정상 작동 확인
- ✅ TAG 인벤토리 최신 상태
- ✅ 사용자 승인 모델 작동 확인

**검증 방법**:
```bash
# Phase 1 라이브러리 테스트 실행
pytest tests/core/doc_tag/ -v --cov

# 백업 디렉토리 접근 권한 확인
ls -la .moai/backups/

# TAG 인벤토리 존재 확인
cat .moai/memory/tag-registry.json
```

---

### Gate 2: 배치 실행 후 (Post-execution)

**조건**:
- ✅ 모든 파일 TAG 삽입 성공 (에러 0개)
- ✅ 백업 파일 생성 완료
- ✅ TAG ID 중복 검증 통과
- ✅ TAG 인벤토리 업데이트 완료

**검증 방법**:
```bash
# TAG 삽입 검증 (Batch 1 예시)
rg "^# @DOC:GUIDE-" CLAUDE-AGENTS-GUIDE.md
rg "^# @DOC:STATUS-" README.md CHANGELOG.md

# 백업 파일 존재 확인
ls -la .moai/backups/batch-1/

# TAG ID 중복 검증
rg "@DOC:GUIDE-AGENT-001" .moai/specs/ docs/ --count-matches
# 결과: 1 (중복 없음)

# TAG 인벤토리 업데이트 확인
jq '.domains["@DOC:GUIDE-*"] | length' .moai/memory/tag-registry.json
# 결과: 3
```

---

### Gate 3: Phase 3 완료 후 (Final)

**조건**:
- ✅ **78/78 파일 모두 태깅 (100%)**
- ✅ TAG 체인 무결성 검증 통과
- ✅ 도메인 일관성 검증 통과
- ✅ TRUST 5 원칙 준수

**검증 방법**:
```bash
# 전체 파일 TAG 카운트
rg "^# @DOC:" docs/ .claude/ .moai/project/ CLAUDE-*.md README.md CHANGELOG.md --count-matches
# 결과: 78

# TAG ID 전역 고유성 검증
rg "@DOC:" . -o | sort | uniq -d
# 결과: (빈 출력 = 중복 없음)

# 도메인별 TAG 집계
rg "@DOC:GUIDE-" . --count
rg "@DOC:SKILL-" . --count
rg "@DOC:STATUS-" . --count
rg "@DOC:PROJECT-" . --count

# Chain 참조 무결성 검증
rg "Chain: @SPEC:.* -> @DOC:" . --count
```

---

## 🔄 Rollback Strategy

### 롤백 시나리오

#### Scenario R1: 배치 실행 중 에러 발생

**Given**: Batch 3 실행 중, 파일 2/3 TAG 삽입 후 에러 발생

**When**: 자동 롤백 트리거

**Then**:
- ✅ 백업 파일 복원 (3개 파일 모두)
- ✅ TAG 인벤토리 롤백 (58/78 → 55/78)
- ✅ 에러 로그 저장 (`.moai/logs/batch-3-error.log`)
- ✅ 사용자에게 에러 원인 표시
- ✅ Batch 3 상태: "rolled_back"

#### Scenario R2: 사용자 수동 롤백 요청

**Given**: Batch 5 완료, 사용자가 결과 불만족

**When**: `/alfred:3-sync --rollback batch-5` 실행

**Then**:
- ✅ Batch 5 백업 파일 복원 (6개 파일)
- ✅ TAG 인벤토리 롤백 (69/78 → 63/78)
- ✅ Batch 5 TAG 제거 (6개)
- ✅ `migration-state.json` 업데이트 (Batch 5: "rolled_back")
- ✅ 다음 배치 제안 표시 (Batch 5 재시도 또는 Batch 6 진행)

#### Scenario R3: 최종 검증 실패 시 전체 롤백

**Given**: Batch 7 완료, 최종 검증에서 TAG ID 중복 발견

**When**: Phase 3 전체 롤백 실행

**Then**:
- ✅ 모든 배치 백업 파일 복원 (33개 파일)
- ✅ TAG 인벤토리 초기화 (78/78 → 45/78)
- ✅ Phase 3 TAG 모두 제거 (33개)
- ✅ `migration-state.json` 초기화
- ✅ 사용자에게 실패 원인 및 재시도 권장 표시

---

## 📊 User Acceptance Scenarios

### Scenario UA1: 전체 배치 순차 실행 (Happy Path)

**Given**: Phase 1/2 완료, 33개 파일 미태깅

**When**:
1. `/alfred:3-sync` 실행
2. Batch 1~7 모두 승인
3. 각 배치 품질 검증 통과

**Then**:
- ✅ 7개 배치 모두 완료
- ✅ 78/78 파일 태깅 (100%)
- ✅ TAG 체인 무결성 검증 통과
- ✅ 사용자에게 Phase 3 완료 리포트 표시:
  ```
  Phase 3 Migration Complete! 🎉
  ├─ Total Files: 78/78 (100%)
  ├─ New TAGs: 33
  ├─ New Domains: 3 (GUIDE, SKILL, STATUS)
  ├─ Batches: 7/7 completed
  └─ Next: Phase 4 (CLI & Automation)
  ```

---

### Scenario UA2: 일부 배치 건너뛰기

**Given**: Batch 3 실행 중, 사용자가 Batch 4 건너뛰기 요청

**When**:
1. Batch 3 승인 및 완료
2. Batch 4 제안 표시
3. 사용자 거부 (`AskUserQuestion` → "n")
4. Batch 5 제안 표시
5. 사용자 승인

**Then**:
- ✅ Batch 3 완료 (58/78 files)
- ✅ Batch 4 건너뜀 (0개 파일 수정)
- ✅ Batch 5 완료 (64/78 files, Batch 4의 5개 파일 제외)
- ✅ 최종 상태: 73/78 files (93.6%)
- ✅ 사용자에게 미태깅 파일 5개 알림

---

### Scenario UA3: 배치 병합 실행 (선택사항)

**Given**: Batch 1, 2 제안 표시

**When**: 사용자가 배치 병합 요청 (`AskUserQuestion` → "Merge Batch 1+2")

**Then**:
- ✅ 10개 파일 동시 스캔 (Batch 1 + Batch 2)
- ✅ 병합된 TAG 제안 표시
- ✅ 승인 시 10개 파일 동시 TAG 삽입
- ✅ 백업 디렉토리: `.moai/backups/batch-1-2/`
- ✅ TAG 인벤토리 업데이트 (55/78 files)

---

### Scenario UA4: 수동 검토 후 승인

**Given**: Batch 6 제안 표시 (Language Skill, LOW 신뢰도)

**When**:
1. 사용자가 수동 검토 선택 (`AskUserQuestion` → "Review")
2. 상세 리포트 표시 (신뢰도 점수, SPEC 매핑)
3. 사용자가 일부 파일만 승인 (4/7 파일)

**Then**:
- ✅ 4개 파일만 TAG 삽입:
  - `@DOC:SKILL-PYTHON-001`
  - `@DOC:SKILL-TYPESCRIPT-001`
  - `@DOC:SKILL-GIT-001`
  - `@DOC:SKILL-CICD-001`
- ✅ 3개 Language Skill 건너뜀
- ✅ TAG 인벤토리 업데이트 (73/78 files)
- ✅ 사용자에게 미태깅 파일 8개 알림 (Batch 6의 3개 + 이전 건너뛴 파일)

---

## 🎯 Definition of Done

Phase 3는 다음 조건을 **모두 만족**할 때 완료됩니다:

### 필수 조건 (Must Have)

- ✅ **78/78 파일 모두 TAG 삽입 완료 (100%)**
- ✅ TAG ID 전역 고유성 검증 (중복 0개)
- ✅ TAG 포맷 표준 준수 (100%)
- ✅ 백업 파일 33개 생성 및 보관
- ✅ TAG 인벤토리 최종 업데이트 (`.moai/memory/tag-registry.json`)
- ✅ 도메인 명명 규칙 일관성 검증 (GUIDE, SKILL, STATUS, PROJECT)
- ✅ Phase 3 완료 리포트 생성

### 선택 조건 (Should Have)

- ✅ Chain 참조 포함 (신뢰도 ≥ 0.5인 파일)
- ✅ Skill Tier 전체 일관성 검증
- ✅ 사용자 피드백 수집 (배치별 만족도)
- ✅ CHANGELOG.md 업데이트 (Phase 3 완료 기록)

### 검증 방법

**정량적 검증**:
```bash
# 78/78 파일 태깅 확인
rg "^# @DOC:" . --count-matches
# 결과: 78

# TAG ID 중복 검증
rg "@DOC:" . -o | sort | uniq -d | wc -l
# 결과: 0 (중복 없음)

# 도메인별 TAG 집계
jq '.domains' .moai/memory/tag-registry.json
# 결과:
# {
#   "@DOC:GUIDE-*": 3,
#   "@DOC:SKILL-*": 26,
#   "@DOC:STATUS-*": 2,
#   "@DOC:PROJECT-*": 2
# }
```

**정성적 검증**:
- ✅ TAG 포맷 육안 검증 (샘플 10개 파일)
- ✅ Chain 참조 추적 가능성 확인 (샘플 5개 파일)
- ✅ Skill Tier 명명 규칙 일관성 검증

---

## 📈 Success Metrics

### 핵심 지표 (KPI)

| 지표 | 목표 | 측정 방법 |
|------|------|----------|
| 파일 태깅 완료율 | 100% (78/78) | `rg "^# @DOC:" . --count-matches` |
| TAG ID 중복률 | 0% | `rg "@DOC:" . -o \| sort \| uniq -d \| wc -l` |
| 백업 파일 생성률 | 100% (33/33) | `ls .moai/backups/batch-*/*.backup \| wc -l` |
| Chain 참조 포함률 | ≥ 60% | `rg "Chain: @SPEC:" . --count-matches` |
| 배치 성공률 | 100% (7/7) | 수동 확인 (migration-state.json) |

### 품질 지표 (Quality)

| 지표 | 목표 | 측정 방법 |
|------|------|----------|
| TAG 포맷 준수율 | 100% | 육안 검증 (샘플 10개) |
| 도메인 일관성 | 100% | Skill Tier별 명명 규칙 검증 |
| 롤백 성공률 | 100% | 롤백 시나리오 테스트 (3개) |
| 사용자 승인 정확도 | ≥ 95% | 사용자 피드백 (배치별) |

### 안전성 지표 (Safety)

| 지표 | 목표 | 측정 방법 |
|------|------|----------|
| 백업 파일 무결성 | 100% | 백업 파일 내용 = 원본 내용 (TAG 삽입 전) |
| 롤백 가능 시간 | 7일 | 백업 파일 보관 기간 |
| 에러 발생 시 자동 롤백 | 100% | 롤백 시나리오 테스트 (Scenario R1) |

---

## 🔍 Regression Testing

### Regression Test 1: Phase 1 라이브러리 호환성

**Given**: Phase 3 완료

**When**: Phase 1 라이브러리 테스트 재실행

**Then**:
- ✅ 90.5% 테스트 커버리지 유지 (Phase 1 baseline)
- ✅ 모든 테스트 통과 (0 failures)
- ✅ `DocTagGenerator`, `TagInserter`, `TagRegistry` 정상 작동

**검증 방법**:
```bash
pytest tests/core/doc_tag/ -v --cov --cov-report=term-missing
```

---

### Regression Test 2: Phase 2 워크플로우 호환성

**Given**: Phase 3 완료

**When**: Phase 2 워크플로우 테스트 재실행 (`/alfred:3-sync`)

**Then**:
- ✅ Phase 1.5 TAG 할당 체크 정상 작동
- ✅ Phase 2.5 자동 TAG 생성 정상 작동
- ✅ 사용자 승인 모델 (`AskUserQuestion`) 정상 작동
- ✅ 백업 관리 시스템 정상 작동

**검증 방법**:
```bash
# 새로운 파일 생성 후 /alfred:3-sync 테스트
echo "# Test Document" > test-doc.md
/alfred:3-sync
# 결과: TAG 제안 표시, 승인 시 TAG 삽입
```

---

### Regression Test 3: TAG 체인 무결성

**Given**: Phase 3 완료

**When**: TAG 체인 검증 실행

**Then**:
- ✅ 모든 Chain 참조 추적 가능:
  ```
  @SPEC:DOC-TAG-001 → @DOC:SKILL-EARS-001
  @SPEC:DOC-TAG-002 → @DOC:SKILL-TAGS-001
  @SPEC:DOC-TAG-003 → @DOC:GUIDE-AGENT-001
  @SPEC:PROJECT-001 → @DOC:SKILL-STRUCTURE-001
  @SPEC:PLAN-001 → @DOC:SKILL-PLAN-WF-001
  ```
- ✅ 고아 TAG 0개 (Chain 참조 누락 없음)
- ✅ 순환 참조 0개

**검증 방법**:
```bash
# Chain 참조 추출
rg "Chain: (@SPEC:\S+) -> (@DOC:\S+)" . -o --no-filename | sort | uniq

# 고아 TAG 검색 (Chain 참조 없는 @DOC)
rg "^# @DOC:" . | rg -v "Chain:" | wc -l
# 결과: 신뢰도 < 0.5인 파일 수 (예상)
```

---

## 📝 Documentation Requirements

### Phase 3 완료 시 생성할 문서

1. **Phase 3 완료 리포트** (`.moai/reports/phase-3-complete.md`):
   - 배치별 실행 결과
   - TAG 생성 통계 (도메인별, Tier별)
   - 품질 검증 결과
   - 사용자 피드백

2. **TAG 시스템 사용자 가이드 업데이트** (`.claude/skills/moai-foundation-tags/SKILL.md`):
   - Phase 3 신규 도메인 추가 (GUIDE, SKILL, STATUS)
   - 배치 마이그레이션 예제 추가
   - TAG 체인 참조 예제 업데이트

3. **CHANGELOG.md 업데이트**:
   ```markdown
   ## v0.8.0 (2025-10-30)

   ### Added
   - Phase 3: 33개 파일 배치 마이그레이션 완료 (78/78 files, 100%)
   - 신규 도메인 3개: @DOC:GUIDE-*, @DOC:SKILL-*, @DOC:STATUS-*
   - TAG 시스템 완전 커버리지 달성

   ### Changed
   - TAG 인벤토리 최종 업데이트 (33개 신규 TAG)
   - Skill Tier 전체 TAG 일관성 확보
   ```

4. **Phase 4 계획 문서** (`.moai/specs/SPEC-DOC-TAG-004/spec.md`):
   - CLI 유틸리티 요구사항
   - Pre-commit Hook 통합 계획
   - 자동 커밋 워크플로우 설계

---

**END OF ACCEPTANCE CRITERIA**
