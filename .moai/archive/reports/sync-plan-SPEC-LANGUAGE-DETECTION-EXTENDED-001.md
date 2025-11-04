# 문서 동기화 계획: SPEC-LANGUAGE-DETECTION-EXTENDED-001

**마지막 업데이트**: 2025-10-31
**상태**: Ready for Execution
**PR 참조**: #135 (병합됨)
**SPEC 상태**: `draft` → `completed`로 변경 필요

---

## 개요 (Executive Summary)

### 상황 분석

- **병합된 PR**: #135 (SPEC-LANGUAGE-DETECTION-EXTENDED-001)
- **현재 브랜치**: `develop` (main 기준 6 커밋 앞서감)
- **변경 내용**: 15개 언어 전담 CI/CD 워크플로우 지원 확대 (4개 → 15개)
- **신규 기능**:
  - 11개 새로운 언어 감지 (Ruby, PHP, Java, Rust, Dart, Swift, Kotlin, C#, C, C++, Shell)
  - 4개 새로운 메서드 추가 (detector.py)
  - 11개 새로운 워크플로우 템플릿 (GitHub Actions)
  - 34개 신규 단위 테스트 (100% 통과)

### 동기화 범위

| 항목 | 상태 | 우선순위 | 설명 |
|------|------|---------|------|
| CHANGELOG.md | ✅ 완료됨 | CRITICAL | v0.11.1 항목이 이미 추가됨 |
| README.md | ⚠️ 검토 필요 | HIGH | 15개 언어 지원 정보 명시 필요 |
| 언어 감지 가이드 | ⚠️ 업데이트 필요 | HIGH | `.moai/docs/language-detection-guide.md` 확대 |
| SPEC 메타데이터 | ⚠️ 업데이트 필요 | HIGH | status: `draft` → `completed` |
| 템플릿 동기화 | ⚠️ 검증 필요 | MEDIUM | 로컬 vs 템플릿 폴더 일관성 확인 |
| TAG 무결성 | ✅ PASS | MEDIUM | 939개 TAGs, 4.2/5.0 health score |

---

## Phase 1: 정적 문서 업데이트 (Status: Ready)

### 1.1 CHANGELOG.md

**현재 상태**: ✅ **이미 업데이트됨**

**확인 사항**:
- [x] v0.11.1 섹션이 존재함 (2025-10-31)
- [x] 11개 언어별 워크플로우 템플릿 나열됨
- [x] 4개 새로운 메서드 문서화됨
- [x] 34개 테스트 항목 기재됨
- [x] @CODE 태그 참조 포함됨

**추가 필요 사항**: 없음 (완료됨)

---

### 1.2 README.md

**현재 상태**: ⚠️ **검토 필요**

**위치**: `/Users/goos/MoAI/MoAI-ADK/README.md`

**필요한 업데이트**:

1. **언어 지원 표 추가 또는 업데이트**
   - 현재: 언어 감지 기능에 대한 일반적 설명만 제공
   - 필요: 15개 지원 언어 명시적 목록

2. **Quick Start 섹션 강화** (선택적)
   - Ruby, Java, Rust 예시 추가
   - 각 언어별 워크플로우 자동 생성 예시

3. **기술 스택 섹션 업데이트** (선택적)
   - 15개 언어 감지 능력 명시
   - CI/CD 자동화 범위 확대 강조

**예상 변경**:
- 약 30-50줄 추가
- 표 1개 추가 (선택적)

---

### 1.3 언어 감지 가이드 문서

**파일**: `/Users/goos/MoAI/MoAI-ADK/.moai/docs/language-detection-guide.md`

**현재 상태**: ⚠️ **기존 문서 확인 후 업데이트**

**필요한 업데이트**:

1. **언어 감지 우선순위 확대**
   - 기존: Python, JavaScript, TypeScript, Go
   - 신규: Rust, Dart, Swift, Kotlin, C#, Java, Ruby, PHP (우선순위 순서)

2. **빌드 도구 감지 문서화**
   ```
   - detect_package_manager(path): 패키지 매니저 자동 감지
   - detect_build_tool(path, language): 빌드 도구 자동 감지
   - get_supported_languages_for_workflows(): 15개 언어 목록 반환
   ```

3. **신규 메서드 API 문서**
   - @CODE:LDE-WORKFLOW-PATH-001
   - @CODE:LDE-PKG-MGR-001
   - @CODE:LDE-BUILD-TOOL-001
   - @CODE:LDE-SUPPORTED-LANGS-001

**예상 변경**:
- 약 100-150줄 추가

---

## Phase 2: SPEC 메타데이터 업데이트

### 2.1 SPEC 파일 상태 변경

**파일**: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-LANGUAGE-DETECTION-EXTENDED-001/spec.md`

**현재 상태**:
```yaml
status: draft
version: 0.0.1
```

**필요한 변경**:
```yaml
status: completed
version: 1.0.0
updated: 2025-10-31
```

**HISTORY 섹션 추가**:
```
### v1.0.0 (2025-10-31) - COMPLETED
- **작성자**: @GoosLab
- **상태**: SPEC 구현 완료 및 마스터 브랜치 병합 완료
- **커밋**: PR #135 병합
```

---

### 2.2 수락 기준 (Acceptance Criteria) 검증

**파일**: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-LANGUAGE-DETECTION-EXTENDED-001/acceptance.md`

**현재 상태**: ✅ **모든 30개 시나리오 통과**

**확인된 항목**:
- 11개 언어 감지 시나리오: ✅ 통과
- 5개 빌드 도구 감지 시나리오: ✅ 통과
- 4개 우선순위 충돌 해결 시나리오: ✅ 통과
- 3개 오류 처리 시나리오: ✅ 통과
- 4개 하위 호환성 시나리오: ✅ 통과
- 3개 통합 시나리오: ✅ 통과

**테스트 실행**: 34개 단위 테스트, 100% 통과

---

## Phase 3: 템플릿 디렉토리 동기화

### 3.1 현재 동기화 상태

**수정된 파일** (git status에서):
```
M .claude/hooks/alfred/shared/core/project.py
M src/moai_adk/templates/.claude/hooks/alfred/shared/core/project.py
```

**상황**:
- 로컬 프로젝트 hook 파일과 템플릿 폴더의 hook 파일이 다를 가능성 있음
- CLAUDE.md 규칙: "패키지 템플릿이 가장 우선이다"

### 3.2 템플릿 폴더 우선 정책

**원칙**:
```
템플릿 폴더 (src/moai_adk/templates/) → 로컬 프로젝트 폴더
```

**확인할 항목**:

1. **Hook 파일 동기화**
   - 소스: `src/moai_adk/templates/.claude/hooks/alfred/shared/core/project.py`
   - 대상: `.claude/hooks/alfred/shared/core/project.py`
   - 상태: 검증 필요

2. **CI/CD 워크플로우 템플릿**
   - 소스: `src/moai_adk/templates/.github/workflows/*.yml`
   - 대상: `.github/workflows/*.yml`
   - 추가된 파일: 11개 (ruby, php, java, rust, dart, swift, kotlin, csharp, c, cpp, shell)
   - 상태: 이미 병합됨 (PR #135)

3. **SPEC 문서**
   - 소스: `src/moai_adk/templates/.moai/specs/`
   - 대상: `.moai/specs/`
   - 상태: 검증 필요 (가능하면 미동기화)

---

## Phase 4: TAG 무결성 검증

### 4.1 TAG 체인 검증

**현재 상태**: ✅ **정상**

| 지표 | 값 | 상태 |
|------|-----|------|
| 전체 TAGs | 939개 | ✅ |
| Health Score | 4.2/5.0 | ✅ |
| SPEC TAGs (src/) | N/A | ✅ |
| CODE TAGs (src/) | 58개 | ✅ |
| TEST TAGs (tests/) | N/A | ✅ |

### 4.2 신규 TAG 추적

**추가된 TAG들**:

| TAG ID | 파일 | 타입 | 설명 |
|--------|------|------|------|
| @CODE:LDE-WORKFLOW-PATH-001 | detector.py | CODE | 15개 언어 워크플로우 경로 반환 메서드 |
| @CODE:LDE-PKG-MGR-001 | detector.py | CODE | 패키지 매니저 자동 감지 메서드 |
| @CODE:LDE-BUILD-TOOL-001 | detector.py | CODE | 빌드 도구 자동 감지 메서드 |
| @CODE:LDE-SUPPORTED-LANGS-001 | detector.py | CODE | 15개 지원 언어 목록 반환 메서드 |
| @TEST:LDE-EXTENDED-001 | test_language_detector_extended.py | TEST | 34개 단위 테스트 모음 |
| @DOC:LANGUAGE-DETECTION-EXTENDED-001 | CHANGELOG.md | DOC | v0.11.1 릴리즈 문서 |

### 4.3 TAG 체인 검증 결과

```
SPEC → CODE → TEST → DOC 체인 확인
├─ @SPEC:LANGUAGE-DETECTION-EXTENDED-001 ✅
├─ @CODE:LDE-WORKFLOW-PATH-001 ✅
├─ @CODE:LDE-PKG-MGR-001 ✅
├─ @CODE:LDE-BUILD-TOOL-001 ✅
├─ @CODE:LDE-SUPPORTED-LANGS-001 ✅
├─ @TEST:LDE-EXTENDED-001 ✅
└─ @DOC:LANGUAGE-DETECTION-EXTENDED-001 ✅
```

---

## 구현 계획 (실행 순서)

### 단계 1: 정적 문서 업데이트 (2-3시간)

**항목**:
1. SPEC 메타데이터 업데이트 (status: draft → completed)
2. language-detection-guide.md 확대
3. README.md 검토 및 필요시 업데이트

**예상 영향**:
- 파일 3개 수정
- 약 150-200줄 추가
- TAG 무결성: 유지

**우선순위**: ⭐⭐⭐ CRITICAL

---

### 단계 2: 템플릿 동기화 검증 (1-2시간)

**항목**:
1. Hook 파일 비교
2. 필요시 로컬 파일 업데이트
3. CI/CD 워크플로우 신규 파일 확인

**예상 영향**:
- Hook 파일 1개 가능성 있음
- 신규 워크플로우 11개 (이미 병합됨)

**우선순위**: ⭐⭐ MEDIUM

---

### 단계 3: TAG 검증 및 보고서 (1시간)

**항목**:
1. TAG 체인 검증 실행
2. 동기화 보고서 생성
3. 깃 커밋 생성

**예상 영향**:
- 보고서 파일 1개 생성
- 커밋 1-2개

**우선순위**: ⭐⭐ MEDIUM

---

## 품질 게이트 (Quality Gates)

### 동기화 완료 확인 체크리스트

- [ ] SPEC 상태: `draft` → `completed`로 변경
- [ ] README.md: 15개 언어 지원 명시
- [ ] language-detection-guide.md: 신규 메서드 문서화
- [ ] TAG 체인: 938개 이상 유지
- [ ] 템플릿 동기화: 로컬과 템플릿 폴더 일치
- [ ] 깃 커밋: 동기화 작업 기록

### PR 준비도 확인

**현재 상태**:
- [x] 코드 구현: 완료 (PR #135 병합)
- [x] 테스트: 완료 (34/34 통과)
- [x] CHANGELOG: 업데이트됨
- [ ] SPEC 메타데이터: 업데이트 필요
- [ ] 문서: 부분 업데이트 필요

**PR 마스터 브랜치 병합 전 필수 사항**:
- [x] 모든 테스트 통과
- [x] 코드 리뷰 완료
- [x] CHANGELOG 업데이트
- [ ] 문서 동기화 완료
- [ ] SPEC 상태 최종화

---

## 예상 소요 시간

| 단계 | 예상 시간 | 실제 | 상태 |
|------|---------|------|------|
| Phase 1: 문서 업데이트 | 2-3시간 | TBD | 준비 완료 |
| Phase 2: 템플릿 동기화 | 1-2시간 | TBD | 준비 완료 |
| Phase 3: TAG 검증 | 1시간 | TBD | 준비 완료 |
| **전체 소요 시간** | **4-6시간** | **TBD** | **준비 완료** |

---

## 다음 단계 (Next Steps)

### 즉시 실행 항목

1. **SPEC 메타데이터 확정**
   - status: `completed`로 변경
   - version: `1.0.0`으로 업데이트
   - HISTORY 섹션 추가

2. **문서 검증 및 업데이트**
   - README.md 검토 (15개 언어 명시)
   - language-detection-guide.md 확대

3. **템플릿 동기화 확인**
   - Hook 파일 비교
   - 필요시 로컬 파일 업데이트

### 이후 단계

4. **TAG 검증 및 보고서 생성**
   - TAG 체인 무결성 확인
   - 동기화 보고서 작성

5. **깃 커밋 생성**
   - 동기화 작업 기록
   - PR 마스터 브랜치 병합 준비

6. **릴리즈 준비** (선택적)
   - 릴리즈 노트 작성
   - 버전 태그 생성

---

## 주의사항 (Warnings)

⚠️ **CRITICAL**:
- SPEC 상태 변경 전 모든 인수 기준(30개 시나리오) 재확인 필수
- 템플릿 폴더 우선 정책 준수 (패키지 템플릿이 우선)

⚠️ **IMPORTANT**:
- 로컬 hook 파일과 템플릿 폴더 파일 불일치 가능성 있음
- 메인 브랜치 병합 전 문서 동기화 완료 필수

---

## 참고 문서

| 문서 | 경로 | 설명 |
|------|------|------|
| SPEC | `.moai/specs/SPEC-LANGUAGE-DETECTION-EXTENDED-001/spec.md` | 15개 언어 확대 요구사항 |
| 인수 기준 | `.moai/specs/SPEC-LANGUAGE-DETECTION-EXTENDED-001/acceptance.md` | 30개 시나리오 및 검증 방법 |
| 구현 코드 | `src/moai_adk/core/project/detector.py` | LanguageDetector 클래스 확대 |
| 테스트 | `tests/unit/test_language_detector_extended.py` | 34개 단위 테스트 |
| CHANGELOG | `CHANGELOG.md` | v0.11.1 변경사항 (이미 업데이트됨) |

---

**문서 작성**: doc-syncer
**작성 일시**: 2025-10-31
**상태**: 실행 준비 완료
