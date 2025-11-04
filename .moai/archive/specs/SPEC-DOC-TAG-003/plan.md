# Implementation Plan: @SPEC:DOC-TAG-003

> **Phase 3: 배치 마이그레이션 - 33개 미태깅 파일 자동 TAG 생성**
>
> **목표**: 42.3% 갭 해소, 78/78 파일 완전 태깅 달성 (100%)

---

## 📋 Implementation Overview

### 현재 상태 (Baseline)

- **총 마크다운 파일**: 78개
- **태깅 완료**: 45개 (57.7%)
- **미태깅 파일**: 33개 (42.3%)
- **Phase 1/2 상태**: 완료 및 프로덕션 준비 완료
- **워크플로우**: Phase 1.5/2.5 통합 완료, 사용자 승인 모델 작동

### 목표 상태 (Target)

- **총 마크다운 파일**: 78개
- **태깅 완료**: 78개 (100%)
- **미태깅 파일**: 0개 (0%)
- **신규 도메인**: `@DOC:GUIDE-*`, `@DOC:SKILL-*`, `@DOC:STATUS-*` (3개)
- **TAG 체인 무결성**: 100%

### 전략

Phase 3는 **7개 배치**로 구성되며, 각 배치는 독립적으로 실행/승인/거부 가능:

1. **Batch 1: Quick Wins** (5개 파일) - 프로젝트 최상위 문서 (README, CHANGELOG 등)
2. **Batch 2: Skills System** (5개 파일) - Foundation Tier Skill 문서
3. **Batch 3: Architecture** (3개 파일) - 아키텍처 핵심 Skill (Structure, Product, Tech)
4. **Batch 4: Concepts** (5개 파일) - 개념 설명 Skill (Context, Workflow 등)
5. **Batch 5: Workflows** (6개 파일) - 워크플로우 관련 Skill (Plan, Run, Sync 등)
6. **Batch 6: Tutorials** (7개 파일) - 언어별 Skill (Python, TypeScript 등)
7. **Batch 7: Polish** (2개 파일) - 프로젝트 메타 문서 (structure.md, tech.md)

### 우선순위 기준

- **가시성**: 사용자에게 보이는 문서 우선 (README, CHANGELOG)
- **난이도**: 간단한 문서부터 시작 (Quick Wins → 복잡한 Skill)
- **의존성**: 도메인 정의가 필요한 문서는 후순위
- **영향도**: 프로젝트 메타 문서는 최종 단계 (Polish)

---

## 🚀 Batch-by-Batch Execution Plan

### Batch 1: Quick Wins

**목표**: 프로젝트 최상위 문서 태깅

**파일 목록** (5개):

| 파일 | TAG ID | 도메인 | 신뢰도 | Chain 참조 |
|------|--------|--------|--------|-----------|
| `CLAUDE-AGENTS-GUIDE.md` | `@DOC:GUIDE-AGENT-001` | GUIDE | HIGH | `@SPEC:DOC-TAG-003 -> @DOC:GUIDE-AGENT-001` |
| `CLAUDE-PRACTICES.md` | `@DOC:GUIDE-PRACTICE-001` | GUIDE | HIGH | `@SPEC:DOC-TAG-003 -> @DOC:GUIDE-PRACTICE-001` |
| `CLAUDE-RULES.md` | `@DOC:GUIDE-RULES-001` | GUIDE | HIGH | `@SPEC:DOC-TAG-003 -> @DOC:GUIDE-RULES-001` |
| `CHANGELOG.md` | `@DOC:STATUS-CHANGELOG-001` | STATUS | HIGH | - |
| `README.md` | `@DOC:STATUS-README-001` | STATUS | HIGH | - |

**신규 도메인**:
- `@DOC:GUIDE-*`: 사용자 가이드 문서 (Agent, Practice, Rules)
- `@DOC:STATUS-*`: 프로젝트 상태 문서 (README, CHANGELOG)

**실행 단계**:

1. **Phase 1.5 준비** (`.claude/commands/alfred/3-sync.md`):
   - 5개 파일 스캔
   - TAG 제안 생성 (신뢰도 HIGH)
   - 제안 목록 표시:
     ```
     Batch 1: Quick Wins (5 files)
     ✅ CLAUDE-AGENTS-GUIDE.md → @DOC:GUIDE-AGENT-001 (HIGH)
     ✅ CLAUDE-PRACTICES.md → @DOC:GUIDE-PRACTICE-001 (HIGH)
     ✅ CLAUDE-RULES.md → @DOC:GUIDE-RULES-001 (HIGH)
     ✅ CHANGELOG.md → @DOC:STATUS-CHANGELOG-001 (HIGH)
     ✅ README.md → @DOC:STATUS-README-001 (HIGH)

     Proceed with Batch 1? [Y/n]
     ```

2. **사용자 승인** (`AskUserQuestion`):
   - 승인 → Phase 2.5 진행
   - 거부 → Batch 2 제안

3. **Phase 2.5 실행** (`.claude/agents/alfred/doc-syncer.md`):
   - 백업 생성: `.moai/backups/batch-1/*.backup`
   - TAG 삽입:
     ```markdown
     # @DOC:GUIDE-AGENT-001 | Chain: @SPEC:DOC-TAG-003 -> @DOC:GUIDE-AGENT-001

     # MoAI-ADK Agent System Guide
     ```
   - TAG 인벤토리 업데이트

4. **품질 검증**:
   - ✅ 5개 파일 모두 TAG 삽입 성공
   - ✅ TAG ID 중복 없음
   - ✅ TAG 포맷 표준 준수
   - ✅ Chain 참조 무결성
   - ✅ 백업 파일 존재 확인

5. **완료 리포트**:
   ```
   Batch 1 Complete ✅
   ├─ Success: 5 files
   ├─ Failed: 0 files
   ├─ Tagged: 50/78 files (64.1%)
   └─ Remaining: 28 files (35.9%)

   Next: Batch 2 (Skills System)
   ```

**예상 작업량**: 6.5시간 (파일당 1.3시간 평균)

**위험 요소**:
- ⚠️ README.md는 자주 수정됨 → 백업 필수
- ⚠️ CHANGELOG.md는 버전 관리 중요 → 신중한 TAG 위치 선택

**완료 조건**:
- ✅ 5개 파일 모두 TAG 삽입
- ✅ 신규 도메인 2개 (GUIDE, STATUS) 인벤토리 등록
- ✅ 백업 파일 5개 생성

---

### Batch 2: Skills System

**목표**: Foundation Tier Skill 문서 태깅

**파일 목록** (5개):

| 파일 | TAG ID | 도메인 | 신뢰도 | Chain 참조 |
|------|--------|--------|--------|-----------|
| `moai-foundation-ears/SKILL.md` | `@DOC:SKILL-EARS-001` | SKILL | HIGH | `@SPEC:DOC-TAG-001 -> @DOC:SKILL-EARS-001` |
| `moai-foundation-specs/SKILL.md` | `@DOC:SKILL-SPECS-001` | SKILL | HIGH | `@SPEC:DOC-TAG-001 -> @DOC:SKILL-SPECS-001` |
| `moai-foundation-tags/SKILL.md` | `@DOC:SKILL-TAGS-001` | SKILL | HIGH | `@SPEC:DOC-TAG-001 -> @DOC:SKILL-TAGS-001` |
| `moai-foundation-trust/SKILL.md` | `@DOC:SKILL-TRUST-001` | SKILL | HIGH | `@SPEC:DOC-TAG-001 -> @DOC:SKILL-TRUST-001` |
| `moai-foundation-hooks/SKILL.md` | `@DOC:SKILL-HOOKS-001` | SKILL | HIGH | - |

**신규 도메인**:
- `@DOC:SKILL-*`: Skill 시스템 문서 (26개 파일 중 첫 5개)

**실행 단계**:

1. **Phase 1.5 준비**:
   - Batch 1 완료 확인
   - 5개 Foundation Skill 스캔
   - SPEC-DOC 매핑 확인 (`@SPEC:DOC-TAG-001`)

2. **사용자 승인**:
   - 제안 목록 표시
   - `AskUserQuestion`: "Batch 2 (Skills System) 실행? [Y/n]"

3. **Phase 2.5 실행**:
   - 백업 생성: `.moai/backups/batch-2/*.backup`
   - TAG 삽입 (Skill 명명 규칙 준수):
     ```markdown
     # @DOC:SKILL-EARS-001 | Chain: @SPEC:DOC-TAG-001 -> @DOC:SKILL-EARS-001

     # EARS Specification Methodology
     ```
   - `moai-foundation-tags/SKILL.md` 자기 참조 업데이트:
     ```markdown
     # @DOC:SKILL-TAGS-001 | Chain: @SPEC:DOC-TAG-001 -> @DOC:SKILL-TAGS-001

     ## TAG System Guide

     This document (@DOC:SKILL-TAGS-001) describes the MoAI-ADK TAG system...
     ```

4. **품질 검증**:
   - ✅ 5개 Skill 파일 모두 TAG 삽입
   - ✅ `moai-foundation-tags` 자기 참조 TAG 포함
   - ✅ Skill Tier 일관성 검증

5. **완료 리포트**:
   ```
   Batch 2 Complete ✅
   ├─ Success: 5 files
   ├─ Failed: 0 files
   ├─ Tagged: 55/78 files (70.5%)
   └─ Remaining: 23 files (29.5%)

   Next: Batch 3 (Architecture)
   ```

**예상 작업량**: 5.5시간 (파일당 1.1시간 평균)

**위험 요소**:
- ⚠️ `moai-foundation-tags` 자기 참조 순환 위험 → 신중한 TAG 삽입 위치 선택

**완료 조건**:
- ✅ 5개 Foundation Skill 모두 TAG 삽입
- ✅ 신규 도메인 1개 (SKILL) 인벤토리 등록
- ✅ `moai-foundation-tags` 자기 참조 TAG 업데이트

---

### Batch 3: Architecture

**목표**: 프로젝트 아키텍처 핵심 Skill 태깅

**파일 목록** (3개):

| 파일 | TAG ID | 도메인 | 신뢰도 | Chain 참조 |
|------|--------|--------|--------|-----------|
| `moai-foundation-structure/SKILL.md` | `@DOC:SKILL-STRUCTURE-001` | SKILL | MEDIUM | `@SPEC:PROJECT-001 -> @DOC:SKILL-STRUCTURE-001` |
| `moai-foundation-product/SKILL.md` | `@DOC:SKILL-PRODUCT-001` | SKILL | MEDIUM | `@SPEC:PROJECT-001 -> @DOC:SKILL-PRODUCT-001` |
| `moai-foundation-tech/SKILL.md` | `@DOC:SKILL-TECH-001` | SKILL | MEDIUM | `@SPEC:PROJECT-001 -> @DOC:SKILL-TECH-001` |

**실행 단계**:

1. **Phase 1.5 준비**:
   - Batch 2 완료 확인
   - 3개 Architecture Skill 스캔
   - SPEC-DOC 매핑 확인 (`@SPEC:PROJECT-001` 존재 여부)

2. **사용자 승인**:
   - 신뢰도 MEDIUM → 수동 검토 권장
   - `AskUserQuestion`: "Batch 3 (Architecture) 실행? [Y/n/Review]"

3. **Phase 2.5 실행**:
   - 백업 생성: `.moai/backups/batch-3/*.backup`
   - TAG 삽입:
     - SPEC 매핑 있으면 Chain 참조 포함
     - SPEC 매핑 없으면 단순 TAG만 삽입

4. **품질 검증**:
   - ✅ 3개 Architecture Skill 모두 TAG 삽입
   - ✅ Chain 참조 무결성 (SPEC 매핑 있는 경우)
   - ✅ 아키텍처 도메인 일관성

5. **완료 리포트**:
   ```
   Batch 3 Complete ✅
   ├─ Success: 3 files
   ├─ Failed: 0 files
   ├─ Tagged: 58/78 files (74.4%)
   └─ Remaining: 20 files (25.6%)

   Next: Batch 4 (Concepts)
   ```

**예상 작업량**: 10시간 (파일당 3.3시간 평균, 복잡도 높음)

**위험 요소**:
- ⚠️ SPEC-DOC 매핑 신뢰도 MEDIUM → 사용자 검토 필요
- ⚠️ `@SPEC:PROJECT-001` 미존재 가능성 → Chain 생략

**완료 조건**:
- ✅ 3개 Architecture Skill 모두 TAG 삽입
- ✅ SPEC 매핑 신뢰도 검증
- ✅ 백업 파일 3개 생성

---

### Batch 4: Concepts

**목표**: 개념 설명 Skill 태깅

**파일 목록** (5개):

| 파일 | TAG ID | 도메인 | 신뢰도 |
|------|--------|--------|--------|
| `moai-essentials-context/SKILL.md` | `@DOC:SKILL-CONTEXT-001` | SKILL | MEDIUM |
| `moai-essentials-workflow/SKILL.md` | `@DOC:SKILL-WORKFLOW-001` | SKILL | MEDIUM |
| `moai-alfred-ears-authoring/SKILL.md` | `@DOC:SKILL-EARS-AUTHOR-001` | SKILL | HIGH |
| `moai-alfred-spec-metadata-validation/SKILL.md` | `@DOC:SKILL-SPEC-META-001` | SKILL | HIGH |
| `moai-alfred-tag-scanning/SKILL.md` | `@DOC:SKILL-TAG-SCAN-001` | SKILL | HIGH |

**실행 단계**:

1. **Phase 1.5 준비**:
   - Batch 3 완료 확인
   - 5개 Concept Skill 스캔
   - Skill Tier 간 일관성 확인 (Essentials vs Alfred)

2. **사용자 승인**:
   - 혼합 신뢰도 (HIGH + MEDIUM) → 개별 검토 옵션 제공
   - `AskUserQuestion`: "Batch 4 (Concepts) 실행? [Y/n/Review]"

3. **Phase 2.5 실행**:
   - 백업 생성: `.moai/backups/batch-4/*.backup`
   - TAG 삽입 (Skill Tier 명명 규칙 준수)

4. **품질 검증**:
   - ✅ 5개 Concept Skill 모두 TAG 삽입
   - ✅ Skill Tier 일관성 (moai-essentials-*, moai-alfred-*)
   - ✅ TAG 명명 규칙 준수

5. **완료 리포트**:
   ```
   Batch 4 Complete ✅
   ├─ Success: 5 files
   ├─ Failed: 0 files
   ├─ Tagged: 63/78 files (80.8%)
   └─ Remaining: 15 files (19.2%)

   Next: Batch 5 (Workflows)
   ```

**예상 작업량**: 17.5시간 (파일당 3.5시간 평균)

**위험 요소**:
- ⚠️ Skill Tier 간 명명 규칙 차이 → 일관성 검증 필수

**완료 조건**:
- ✅ 5개 Concept Skill 모두 TAG 삽입
- ✅ Skill Tier 일관성 검증 통과
- ✅ 백업 파일 5개 생성

---

### Batch 5: Workflows

**목표**: 워크플로우 관련 Skill 태깅

**파일 목록** (6개):

| 파일 | TAG ID | 도메인 | 신뢰도 | Chain 참조 |
|------|--------|--------|--------|-----------|
| `moai-alfred-plan-workflow/SKILL.md` | `@DOC:SKILL-PLAN-WF-001` | SKILL | HIGH | `@SPEC:PLAN-001 -> @DOC:SKILL-PLAN-WF-001` |
| `moai-alfred-run-workflow/SKILL.md` | `@DOC:SKILL-RUN-WF-001` | SKILL | HIGH | `@SPEC:RUN-001 -> @DOC:SKILL-RUN-WF-001` |
| `moai-alfred-sync-workflow/SKILL.md` | `@DOC:SKILL-SYNC-WF-001` | SKILL | HIGH | `@SPEC:SYNC-001 -> @DOC:SKILL-SYNC-WF-001` |
| `moai-alfred-project-workflow/SKILL.md` | `@DOC:SKILL-PROJECT-WF-001` | SKILL | HIGH | `@SPEC:PROJECT-001 -> @DOC:SKILL-PROJECT-WF-001` |
| `moai-alfred-trust-validation/SKILL.md` | `@DOC:SKILL-TRUST-VAL-001` | SKILL | HIGH | `@SPEC:TRUST-001 -> @DOC:SKILL-TRUST-VAL-001` |
| `moai-alfred-ask-user-questions/SKILL.md` | `@DOC:SKILL-INTERACTIVE-001` | SKILL | MEDIUM | - |

**실행 단계**:

1. **Phase 1.5 준비**:
   - Batch 4 완료 확인
   - 6개 Workflow Skill 스캔
   - SPEC-DOC 매핑 확인 (대부분 HIGH 신뢰도)

2. **사용자 승인**:
   - 대부분 HIGH 신뢰도 → 자동 제안
   - `AskUserQuestion`: "Batch 5 (Workflows) 실행? [Y/n]"

3. **Phase 2.5 실행**:
   - 백업 생성: `.moai/backups/batch-5/*.backup`
   - TAG 삽입 (워크플로우 도메인 명명 규칙):
     ```markdown
     # @DOC:SKILL-PLAN-WF-001 | Chain: @SPEC:PLAN-001 -> @DOC:SKILL-PLAN-WF-001

     # Plan Workflow Guide
     ```

4. **품질 검증**:
   - ✅ 6개 Workflow Skill 모두 TAG 삽입
   - ✅ Chain 참조 무결성 (5개 파일)
   - ✅ 워크플로우 도메인 일관성

5. **완료 리포트**:
   ```
   Batch 5 Complete ✅
   ├─ Success: 6 files
   ├─ Failed: 0 files
   ├─ Tagged: 69/78 files (88.5%)
   └─ Remaining: 9 files (11.5%)

   Next: Batch 6 (Tutorials)
   ```

**예상 작업량**: 19시간 (파일당 3.2시간 평균)

**위험 요소**:
- ⚠️ SPEC 매핑 복잡도 높음 → Chain 참조 검증 필수

**완료 조건**:
- ✅ 6개 Workflow Skill 모두 TAG 삽입
- ✅ SPEC-DOC 매핑 5개 성공
- ✅ Chain 참조 무결성 검증 통과

---

### Batch 6: Tutorials

**목표**: 튜토리얼 및 고급 Skill 태깅

**파일 목록** (7개):

| 파일 | TAG ID | 도메인 | 신뢰도 |
|------|--------|--------|--------|
| `moai-domain-python/SKILL.md` | `@DOC:SKILL-PYTHON-001` | SKILL | MEDIUM |
| `moai-domain-typescript/SKILL.md` | `@DOC:SKILL-TYPESCRIPT-001` | SKILL | MEDIUM |
| `moai-ops-git/SKILL.md` | `@DOC:SKILL-GIT-001` | SKILL | MEDIUM |
| `moai-ops-ci-cd/SKILL.md` | `@DOC:SKILL-CICD-001` | SKILL | MEDIUM |
| `moai-language-korean/SKILL.md` | `@DOC:SKILL-KOREAN-001` | SKILL | LOW |
| `moai-language-japanese/SKILL.md` | `@DOC:SKILL-JAPANESE-001` | SKILL | LOW |
| `moai-language-spanish/SKILL.md` | `@DOC:SKILL-SPANISH-001` | SKILL | LOW |

**실행 단계**:

1. **Phase 1.5 준비**:
   - Batch 5 완료 확인
   - 7개 Tutorial Skill 스캔
   - Language Tier 명명 규칙 확인

2. **사용자 승인**:
   - 혼합 신뢰도 (MEDIUM + LOW) → 개별 검토 권장
   - `AskUserQuestion`: "Batch 6 (Tutorials) 실행? [Y/n/Review]"

3. **Phase 2.5 실행**:
   - 백업 생성: `.moai/backups/batch-6/*.backup`
   - TAG 삽입 (Language Tier 일관성 유지)

4. **품질 검증**:
   - ✅ 7개 Tutorial Skill 모두 TAG 삽입
   - ✅ Language Tier 일관성 (moai-language-*)
   - ✅ Domain/Ops Tier 일관성 (moai-domain-*, moai-ops-*)

5. **완료 리포트**:
   ```
   Batch 6 Complete ✅
   ├─ Success: 7 files
   ├─ Failed: 0 files
   ├─ Tagged: 76/78 files (97.4%)
   └─ Remaining: 2 files (2.6%)

   Next: Batch 7 (Polish)
   ```

**예상 작업량**: 26시간 (파일당 3.7시간 평균, Language Skill 복잡도 높음)

**위험 요소**:
- ⚠️ Language Skill 신뢰도 LOW → 수동 검토 필수
- ⚠️ Domain Tier 분류 모호 → 도메인 명명 규칙 재확인

**완료 조건**:
- ✅ 7개 Tutorial Skill 모두 TAG 삽입
- ✅ Skill Tier 전체 일관성 검증 (Foundation → Essentials → Alfred → Domain → Language → Ops)
- ✅ 백업 파일 7개 생성

---

### Batch 7: Polish

**목표**: 프로젝트 메타 문서 태깅 (Phase 3 완료)

**파일 목록** (2개):

| 파일 | TAG ID | 도메인 | 신뢰도 | Chain 참조 |
|------|--------|--------|--------|-----------|
| `.moai/project/structure.md` | `@DOC:PROJECT-STRUCTURE-001` | PROJECT | HIGH | `@SPEC:PROJECT-001 -> @DOC:PROJECT-STRUCTURE-001` |
| `.moai/project/tech.md` | `@DOC:PROJECT-TECH-001` | PROJECT | HIGH | `@SPEC:PROJECT-001 -> @DOC:PROJECT-TECH-001` |

**실행 단계**:

1. **Phase 1.5 준비**:
   - Batch 6 완료 확인
   - 2개 프로젝트 메타 문서 스캔
   - SPEC-DOC 매핑 확인 (`@SPEC:PROJECT-001`)

2. **사용자 승인**:
   - 최종 배치 → 신중한 확인
   - `AskUserQuestion`: "Batch 7 (Polish - FINAL) 실행? [Y/n]"

3. **Phase 2.5 실행**:
   - 백업 생성: `.moai/backups/batch-7/*.backup`
   - TAG 삽입:
     ```markdown
     # @DOC:PROJECT-STRUCTURE-001 | Chain: @SPEC:PROJECT-001 -> @DOC:PROJECT-STRUCTURE-001

     # Project Structure
     ```

4. **품질 검증**:
   - ✅ 2개 프로젝트 문서 모두 TAG 삽입
   - ✅ Chain 참조 무결성
   - ✅ **최종 검증**: 78/78 파일 모두 태깅 완료 (100%)

5. **완료 리포트**:
   ```
   Batch 7 Complete ✅
   ├─ Success: 2 files
   ├─ Failed: 0 files
   ├─ Tagged: 78/78 files (100%) ← COMPLETE!
   └─ Remaining: 0 files (0%)

   Phase 3 Migration Complete! 🎉
   ```

6. **최종 검증 체크리스트**:
   - ✅ 78/78 파일 모두 태깅 완료
   - ✅ TAG ID 전역 고유성 검증
   - ✅ 도메인 명명 규칙 일관성 검증
   - ✅ Chain 참조 전체 추적 가능
   - ✅ TAG 인벤토리 최종 업데이트
   - ✅ `.moai/memory/tag-registry.json` 업데이트

**예상 작업량**: 3시간 (파일당 1.5시간 평균)

**위험 요소**:
- ⚠️ 프로젝트 메타 문서 중요도 높음 → 백업 필수
- ⚠️ 최종 검증 실패 시 Phase 3 전체 롤백 가능성

**완료 조건**:
- ✅ 2개 프로젝트 문서 모두 TAG 삽입
- ✅ **Phase 3 완료**: 78/78 파일 100% 태깅 달성
- ✅ Phase 4 (CLI & Automation) 준비 완료

---

## 📊 Implementation Summary

### 총 작업량 예상

| Batch | 파일 수 | 예상 시간 | 누적 진행률 |
|-------|--------|----------|------------|
| Batch 1: Quick Wins | 5 | 6.5h | 64.1% |
| Batch 2: Skills System | 5 | 5.5h | 70.5% |
| Batch 3: Architecture | 3 | 10h | 74.4% |
| Batch 4: Concepts | 5 | 17.5h | 80.8% |
| Batch 5: Workflows | 6 | 19h | 88.5% |
| Batch 6: Tutorials | 7 | 26h | 97.4% |
| Batch 7: Polish | 2 | 3h | 100% |
| **합계** | **33** | **87.5h** | **100%** |

**참고**: 시간 예상은 참고용이며, 실제 실행은 사용자 승인 및 품질 검증 시간에 따라 달라짐.

### 도메인 분포

| 도메인 | 파일 수 | 예시 |
|--------|--------|------|
| `@DOC:SKILL-*` | 26 | `@DOC:SKILL-EARS-001`, `@DOC:SKILL-PLAN-WF-001` |
| `@DOC:GUIDE-*` | 3 | `@DOC:GUIDE-AGENT-001`, `@DOC:GUIDE-PRACTICE-001` |
| `@DOC:STATUS-*` | 2 | `@DOC:STATUS-README-001`, `@DOC:STATUS-CHANGELOG-001` |
| `@DOC:PROJECT-*` | 2 | `@DOC:PROJECT-STRUCTURE-001`, `@DOC:PROJECT-TECH-001` |
| **합계** | **33** | |

---

## 🔧 Implementation Strategy

### 의존성 관리

**Phase 1/2 완료 확인**:
- ✅ Phase 1 라이브러리 (90.5% 테스트 커버리지)
- ✅ Phase 2 워크플로우 (Phase 1.5/2.5 통합)
- ✅ 사용자 승인 모델 (`AskUserQuestion`)
- ✅ 백업 관리 시스템

**배치 간 의존성**:
- Batch 1 → Batch 2: 신규 도메인 (GUIDE, STATUS) 정의 필요
- Batch 2 → Batch 3: Skill 도메인 명명 규칙 확정
- Batch 6 → Batch 7: Skill Tier 전체 일관성 확인 후 프로젝트 메타 문서 태깅

### 기술 접근

**Phase 1.5 (`/alfred:3-sync`)**:
- 배치 단위 파일 스캔
- TAG 제안 생성 (신뢰도 점수 포함)
- 사용자 승인/거부 수집 (`AskUserQuestion`)

**Phase 2.5 (`doc-syncer`)**:
- 백업 생성 (배치 단위)
- TAG 삽입 (마크다운 헤더 첫 줄)
- TAG 인벤토리 업데이트
- 품질 검증 (TRUST 5 원칙)

**백업 관리**:
- 배치별 백업 디렉토리 (`. moai/backups/batch-N/`)
- 롤백 시 백업 파일 복원
- 성공 시 백업 보관 (7일)

### 아키텍처 설계

```
/alfred:3-sync
    ↓
Phase 0: 프로젝트 분석
    ↓
Phase 1.5: TAG 할당 체크
    ├─ Batch 1 스캔
    ├─ TAG 제안 (5개 파일)
    ├─ 신뢰도 점수 계산
    └─ AskUserQuestion: "Batch 1 실행?"
    ↓
사용자 승인
    ↓
Phase 1: doc-syncer 호출
    ↓
Phase 2.5: @DOC TAG 자동 생성
    ├─ 백업 생성
    ├─ TAG 삽입
    ├─ 인벤토리 업데이트
    └─ 품질 검증
    ↓
Batch 1 완료 리포트
    └─ AskUserQuestion: "Batch 2 진행?"
    ↓
반복 (Batch 2 → Batch 7)
    ↓
최종 검증
    ├─ 78/78 파일 태깅 확인
    ├─ TAG 체인 무결성 검증
    └─ 도메인 일관성 검증
    ↓
Phase 3 완료 (100%)
```

---

## ⚠️ Risk Assessment

### 식별된 위험 요소

| 위험 | 영향도 | 확률 | 완화 전략 |
|------|--------|------|-----------|
| 백업 생성 실패 | 높음 | 낮음 | 자동 롤백, 사용자 알림 |
| TAG ID 중복 | 중간 | 중간 | 자동 증분, TAG 인벤토리 검증 |
| SPEC 매핑 신뢰도 낮음 | 중간 | 중간 | 사용자 수동 검토 요청, Chain 생략 |
| 배치 실행 중 에러 | 높음 | 낮음 | 배치별 롤백, 에러 로그 저장 |
| 도메인 명명 규칙 불일치 | 낮음 | 중간 | Skill Tier별 일관성 검증 |
| 사용자 승인 거부 | 낮음 | 낮음 | 다음 배치 제안, 중단 옵션 제공 |
| 최종 검증 실패 | 높음 | 낮음 | Phase 3 전체 롤백, 품질 게이트 재실행 |

### 완화 전략

**백업 및 롤백**:
- 모든 배치 실행 전 백업 생성
- 에러 발생 시 자동 롤백
- 백업 파일 7일 보관

**품질 검증**:
- 각 배치 완료 후 TAG 형식 검증
- TAG ID 전역 고유성 검증
- Chain 참조 무결성 검증
- Skill Tier 일관성 검증

**사용자 상호작용**:
- 신뢰도 점수 표시
- 배치별 승인/거부 옵션
- 수동 검토 권장 (신뢰도 < 0.5)

---

## ✅ Success Criteria

### 배치별 성공 조건

**각 배치 완료 시**:
- ✅ 모든 파일 TAG 삽입 성공
- ✅ TAG ID 중복 없음
- ✅ TAG 포맷 표준 준수
- ✅ Chain 참조 무결성 (SPEC 매핑 있는 경우)
- ✅ 백업 파일 존재 확인
- ✅ TAG 인벤토리 업데이트

### Phase 3 전체 완료 조건

- ✅ **78/78 파일 모두 태깅 완료 (100%)**
- ✅ TAG ID 전역 고유성 검증
- ✅ 도메인 명명 규칙 일관성 검증 (GUIDE, SKILL, STATUS, PROJECT)
- ✅ Chain 참조 전체 추적 가능
- ✅ TAG 인벤토리 최종 검증
- ✅ `.moai/memory/tag-registry.json` 업데이트
- ✅ 백업 파일 보관 (`.moai/backups/batch-1~7/`)
- ✅ Phase 4 (CLI & Automation) 준비 완료

### TRUST 5 원칙 검증

**Test First**:
- ✅ Phase 1 라이브러리 테스트 커버리지 90.5% 유지
- ✅ 각 배치 품질 검증 통과

**Readable**:
- ✅ TAG 명명 규칙 일관성
- ✅ Chain 참조 명확성

**Unified**:
- ✅ 도메인 분류 일관성 (GUIDE, SKILL, STATUS, PROJECT)
- ✅ Skill Tier 일관성 (Foundation → Essentials → Alfred → Domain → Language → Ops)

**Secured**:
- ✅ 백업 관리 시스템 작동
- ✅ 롤백 가능성 보장

**Trackable**:
- ✅ TAG 인벤토리 실시간 업데이트
- ✅ 배치별 진행 상태 추적
- ✅ Chain 참조 전체 추적

---

## 🎯 Completion Conditions

### Phase 3 완료 정의

**정량적 지표**:
- ✅ 78/78 파일 태깅 완료 (100%)
- ✅ 33개 신규 TAG 생성
- ✅ 3개 신규 도메인 정의 (GUIDE, SKILL, STATUS)
- ✅ 2개 확장 도메인 (PROJECT)
- ✅ 7개 배치 모두 성공

**정성적 지표**:
- ✅ TAG 체계 일관성 유지
- ✅ 사용자 승인 모델 작동
- ✅ 백업 및 롤백 시스템 작동
- ✅ TRUST 5 원칙 준수

### Next Steps (Phase 4)

Phase 3 완료 후 다음 단계:

1. **Phase 4 계획** (`@SPEC:DOC-TAG-004`):
   - CLI 유틸리티 개발 (`moai-adk tag-generate`, `tag-validate`, `tag-map`)
   - Pre-commit Hook 통합
   - 자동 커밋 워크플로우

2. **문서화**:
   - Phase 3 완료 리포트 작성
   - TAG 시스템 사용자 가이드 업데이트
   - CHANGELOG.md 업데이트 (v0.8.0)

3. **최종 검증**:
   - 전체 TAG 체계 검토
   - 사용자 피드백 수집
   - Phase 4 우선순위 확정

---

**END OF PLAN**
