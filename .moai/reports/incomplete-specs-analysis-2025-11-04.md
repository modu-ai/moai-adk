# 📋 미완료 SPEC 상세 보고서

**작성일**: 2025-11-04
**분석 대상**: 6개 미완료 SPEC (in-progress 5개 + planning 1개)
**총 구현 완료율**: 88% (48/54)

---

## 📊 전체 현황

| 상태 | 개수 | 비율 | 예상 시간 |
|------|------|------|----------|
| **in-progress** | 5 | 9.3% | 3-4주 |
| **planning** | 1 | 1.9% | 1-2주 |
| **closed** | 25 | 46.3% | ✅ 완료 |
| **implementation-complete** | 23 | 42.6% | ✅ 완료 |

---

# 🟡 IN-PROGRESS SPEC (5개)

## 1. SPEC-DOC-TAG-003: @DOC 태그 자동 생성 - 배치 마이그레이션 (Phase 3)

### 📌 기본 정보
- **상태**: in-progress
- **우선순위**: High
- **버전**: v0.0.1
- **작성자**: @Goos
- **의존성**: SPEC-DOC-TAG-001, SPEC-DOC-TAG-002

### 🎯 목표
Phase 1-2를 통해 완성된 @DOC TAG 자동 생성 라이브러리와 워크플로우를 이용하여, **33개의 미태깅 파일을 7개 배치로 나누어 점진적으로 태깅하는 대규모 마이그레이션** 실행.

### 📋 주요 내용

#### 현재 상태 (2025-10-29)
- **총 마크다운 파일**: 78개
- **태깅 완료**: 45개 (57.7%)
- **미태깅 파일**: 33개 (42.3%)
- **기존 도메인**: `@DOC:AUTH-*`, `@DOC:INSTALLER-*`, `@DOC:PLAN-*` 등 8개
- **신규 도메인 (예상)**: `@DOC:GUIDE-*`, `@DOC:SKILL-*`, `@DOC:STATUS-*`

#### 7개 배치 전략

| Batch | 카테고리 | 파일 수 | 예상 시간 | 설명 |
|-------|---------|--------|----------|------|
| 1 | Quick Wins | 5개 | 6.5h | 프로젝트 최상위 문서 (README, CHANGELOG) |
| 2 | Skills System | 5개 | 5.5h | Foundation Tier Skill 문서 |
| 3 | Architecture | 3개 | 10h | 프로젝트 핵심 아키텍처 Skill |
| 4 | Concepts | 5개 | 17.5h | 개념 설명 Skill |
| 5 | Workflows | 6개 | 19h | 워크플로우 관련 Skill |
| 6 | Tutorials | 7개 | 26h | 튜토리얼 및 고급 Skill |
| 7 | Polish | 2개 | 3h | 프로젝트 메타 문서 (최종 완료) |

**총 예상 시간**: 87.5시간 (약 2주)

#### 신규 도메인 규칙

```yaml
@DOC:GUIDE-*      # 사용자 가이드 (예: @DOC:GUIDE-AGENT-001)
@DOC:SKILL-*      # Skill 시스템 (예: @DOC:SKILL-EARS-001)
@DOC:STATUS-*     # 프로젝트 상태 (예: @DOC:STATUS-README-001)
@DOC:PROJECT-*    # 프로젝트 메타 (예: @DOC:PROJECT-STRUCTURE-001)
```

#### 품질 검증

각 배치 완료 후:
- ✅ 모든 파일에 TAG 삽입 성공
- ✅ TAG ID 중복 검사
- ✅ Chain 참조 무결성
- ✅ TAG 인벤토리 업데이트

최종 검증:
- ✅ **78/78 파일 100% 태깅 완료**
- ✅ TAG 체인 무결성
- ✅ 도메인 명명 규칙 일관성

### 🚀 다음 단계
1. Phase 2 (`doc-syncer`) 정상 작동 확인
2. 백업 시스템 테스트
3. Batch 1 (Quick Wins) 시작
4. 사용자 승인 기반 점진적 진행

---

## 2. SPEC-DOC-TAG-004: @DOC TAG 자동 검증 및 품질 게이트 (Phase 4)

### 📌 기본 정보
- **상태**: in-progress
- **우선순위**: High
- **버전**: v0.0.1
- **작성자**: @Goos
- **의존성**: SPEC-DOC-TAG-001, SPEC-DOC-TAG-002, SPEC-DOC-TAG-003

### 🎯 목표
Phase 1-3을 통해 생성된 **82개 파일의 @DOC TAG를 자동으로 검증**하고, 중복/누락/고아 TAG를 감지하는 **CI/CD 품질 게이트** 구축.

### 📋 주요 내용

#### 현재 문제점
- ❌ TAG 무결성 자동 검증 부재
- ❌ 커밋 시 TAG 중복/누락 감지 불가
- ❌ CI/CD 파이프라인 미통합
- ❌ TAG 인벤토리 자동 업데이트 부재
- ❌ 품질 게이트 미적용

#### 4개 핵심 컴포넌트

| Component | 목적 | 구현 시간 | 기대 효과 |
|-----------|------|----------|----------|
| **Pre-commit Hooks** | 로컬 커밋 시 빠른 검증 | 8h | 조기 문제 발견 (70% 실패율 감소) |
| **CI/CD Pipeline** | GitHub Actions 전체 검증 | 12h | PR 머지 전 100% 검증 |
| **Validation System** | 복잡한 TAG 체인/중복/고아 감지 | 15h | 정밀한 검증 + 재사용 가능한 라이브러리 |
| **Documentation** | TAG 인벤토리/매트릭스 자동 생성 | 8h | 투명성 확보 + 추적 가능 |

**총 구현 시간**: 43시간 (약 1주)

#### 검증 기능

**중복 TAG 감지**:
- 동일한 TAG ID가 여러 파일에 존재하는지 자동 확인
- 문서 참조는 중복 허용 (예: plan.md에서 `@SPEC:AUTH-001` 참조)

**고아 TAG 검증**:
- TAG 체인이 끊긴 TAG 발견 (예: `@TEST:AUTH-001` 존재, `@SPEC:AUTH-001` 부재)
- 체인 복구 제안 자동 생성

**체인 무결성 검증**:
```
@SPEC:DOC-TAG-001
    ↓ (의존)
@SPEC:DOC-TAG-002
    ↓ (의존)
@SPEC:DOC-TAG-003
    ↓ (의존)
@SPEC:DOC-TAG-004 (현재)
```

#### 성공 기준
- ✅ Pre-commit Hook 실행 시간 3초 이내
- ✅ CI/CD 파이프라인 5분 이내
- ✅ TAG 중복 감지율 100%
- ✅ 고아 TAG 감지율 95%+
- ✅ False Positive 비율 5% 이하

### 🚀 다음 단계
1. Pre-commit Hook 설계 (`.moai/hooks/pre-commit.sh`)
2. TAG Validator 클래스 구현 (`src/moai_adk/core/tags/validator.py`)
3. CI/CD 워크플로우 통합 (`.github/workflows/tag-validation.yml`)
4. TAG 인벤토리 자동 생성 (`.moai/reports/tag-inventory.md`)

---

## 3. SPEC-BUGFIX-001: Windows Compatibility - Cross-Platform Timeout Handler

### 📌 기본 정보
- **상태**: in-progress
- **우선순위**: CRITICAL
- **버전**: v0.11.0
- **작성자**: debug-helper
- **이슈**: #129

### 🎯 목표
**Windows 사용자를 위해 Unix 전용 `signal.SIGALRM`을 `threading.Timer`로 대체**하여 모든 OS에서 MoAI-ADK Hook 시스템이 정상 작동하도록 수정.

### 📋 주요 내용

#### 문제
```
❌ AttributeError: module 'signal' has no attribute 'SIGALRM'
```

**영향 범위**: 11개 파일
- `alfred_hooks.py` (메인)
- `notification__handle_events.py`
- `post_tool__log_changes.py`
- `pre_tool__auto_checkpoint.py`
- `session_end__cleanup.py`
- `session_start__show_project_info.py`
- `stop__handle_interrupt.py`
- `subagent_stop__handle_subagent_end.py`
- `user_prompt__jit_load_docs.py`
- `core/project.py`
- `shared/core/project.py`

**심각도**: 🔴 CRITICAL (Windows 사용자 100% 기능 차단)

#### 해결 방안

**새로운 Cross-Platform Timeout 모듈 구현**:

```python
class CrossPlatformTimeout:
    """Windows는 threading.Timer, Unix는 signal.SIGALRM 사용"""

    def start(self):
        if platform.system() == "Windows":
            self.timer = threading.Timer(timeout_seconds, callback)
            self.timer.start()
        else:
            signal.signal(signal.SIGALRM, handler)
            signal.alarm(timeout_seconds)
```

#### 구현 계획

| Phase | 작업 | 기간 |
|-------|------|------|
| 1 | `utils/timeout.py` 모듈 생성 + 단위 테스트 | Day 1 |
| 2 | 11개 Hook 파일 업데이트 | Day 2-3 |
| 3 | CI/CD 통합 (Windows 테스트 매트릭스) | Day 4 |
| 4 | 문서화 업데이트 | Day 5 |
| 5 | Release (v0.11.0) | Day 6 |

#### 성공 기준
- ✅ Windows 10/11에서 Hook 정상 작동
- ✅ macOS, Linux 호환성 유지
- ✅ 모든 플랫폼에서 단위/통합 테스트 통과
- ✅ 타임아웃 오버헤드 < 10ms
- ✅ CI/CD 모든 OS에서 성공

### 🚀 다음 단계
1. `utils/timeout.py` 작성
2. 단위 테스트 작성 (Windows/Unix 플랫폼 분기)
3. `alfred_hooks.py` 업데이트
4. 10개 핸들러 파일 업데이트
5. GitHub Actions 테스트 매트릭스 확장

---

## 4. SPEC-CHECKPOINT-EVENT-001: Event-Driven Checkpoint 시스템

### 📌 기본 정보
- **상태**: in-progress
- **우선순위**: High
- **버전**: v0.1.0
- **작성자**: @Goos
- **카테고리**: Safety Net / Git

### 🎯 목표
**시간 기반 checkpoint (5분 간격) → 이벤트 기반 checkpoint**로 전환하여, 위험한 작업 전에만 자동으로 local branch를 생성하는 안전망 시스템 구축.

### 📋 주요 내용

#### 핵심 개선
- **이전**: 5분마다 무조건 checkpoint → 불필요한 tag 생성
- **이후**: 위험한 작업 전에만 checkpoint → 명확한 의도 표현

#### 감지할 위험 작업

| 위험 작업 | 트리거 조건 | 예시 |
|----------|-----------|------|
| 대규모 파일 삭제 | 10개 이상 파일 삭제 | `git rm` 명령, `rm -rf` |
| 복잡한 리팩토링 | 클래스 분리, 함수 추출 | 10개 이상 파일 이름 변경 |
| 병합 작업 | `git merge` 충돌 해결 | feature → main merge |
| 외부 스크립트 | 사용자 정의 스크립트 | Bash tool 복잡한 명령 |
| 중요 파일 수정 | config.json, CLAUDE.md 수정 | `.moai/memory/` 파일 변경 |

#### 구현 파일 구조
```python
src/moai_adk/core/git/
├── checkpoint.py        # Checkpoint 생성/복구/관리
├── event_detector.py    # 위험 작업 감지
└── branch_manager.py    # Local branch 관리 (최대 10개)

tests/unit/
├── test_checkpoint.py
├── test_event_detector.py
└── test_branch_manager.py (coverage: 85%)
```

#### 성공 기준
- ✅ 10개 이상 파일 삭제 시 자동 checkpoint 생성
- ✅ 대규모 리팩토링 시 자동 checkpoint 생성
- ✅ Checkpoint는 local branch로만 존재 (원격 push 차단)
- ✅ 최대 10개 checkpoint 유지 (FIFO)
- ✅ Checkpoint 복구 기능 정상 작동
- ✅ 테스트 커버리지 85%+

### 📝 상세 현황
**v0.1.0 (2025-10-16) 완료**:
- checkpoint.py 구현 ✅
- event_detector.py 구현 ✅
- branch_manager.py 구현 ✅
- 테스트 작성 ✅
- 테스트 커버리지 85% 달성 ✅

**왜 여전히 in-progress인가?**
- 로컬 브랜치 관리 정책 최종 검증 진행 중
- 원격 저장소와의 동기화 방식 결정 대기

### 🚀 다음 단계
1. 로컬 전용 checkpoint 정책 최종 확정
2. `.moai/checkpoints.log` 포맷 정의
3. Alfred 워크플로우와 통합
4. 최종 테스트 및 문서화

---

## 5. SPEC-HOOKS-003: TRUST 원칙 자동 검증 (PostToolUse 통합)

### 📌 기본 정보
- **상태**: in-progress
- **우선순위**: High
- **버전**: v0.1.0
- **작성자**: @Goos
- **의존성**: HOOKS-001, TRUST-001

### 🎯 목표
**/alfred:2-run 완료 후 TRUST 5 원칙 검증을 PostToolUse Hook에서 자동으로 실행**하여, 품질 기준을 충족하지 않은 코드가 repository에 병합되는 것을 방지.

### 📋 주요 내용

#### TRUST 5 원칙
```
T (Test)         - 85% 이상 테스트 커버리지
R (Readable)     - 코드 가독성 (Ruff 린트)
U (Unified)      - 일관된 스타일 (Black 포맷)
S (Secured)      - 보안 검사 (Bandit, pip-audit)
T (Trackable)    - TAG 추적성 (@TAG 존재)
```

#### PostToolUse Hook 실행 흐름

```
도구 실행 완료
    ↓
PostToolUse Hook 트리거
    ├─ 비동기 실행 (100ms 제약)
    └─ subprocess.Popen()로 검증 백그라운드 시작
    ↓
Git 커밋 로그 분석 (최근 5개)
    ├─ 🟢 GREEN: 또는 ♻️ REFACTOR: 감지
    └─ TDD 흐름 확인 (RED → GREEN → REFACTOR)
    ↓
TRUST-001 검증 도구 실행
    └─ scripts/validate_trust.py
    ↓
검증 결과 (별도 notification)
    ├─ 성공: "✅ TRUST 5 원칙 통과"
    └─ 실패: "❌ 테스트 커버리지 부족 (78%)"
```

#### 성능 제약
- Git 로그 파싱: < 10ms
- 검증 프로세스 시작: < 50ms
- **전체 동기 실행: < 100ms** (PostToolUse 제약)

#### 구현 상황
**v0.1.0 (2025-10-18) 완료**:
- TDD 단계 감지 로직 구현 ✅
- PostToolUse 핸들러 통합 ✅
- 비동기 실행 메커니즘 ✅
- 구현 검증 완료 ✅

**왜 여전히 in-progress인가?**
- 검증 실패 시 자동 복구 로직 설계 중
- notification 메시지 포맷 최적화 진행 중

### 🚀 다음 단계
1. TRUST-001 검증 도구 최종 버전 확정
2. notification 메시지 포맷 정의
3. GitHub Actions CI/CD 통합
4. 사용자 승인 플로우 설계

---

# 🔴 PLANNING SPEC (1개)

## 6. SPEC-FEATURE-001: Dynamic Document Reference System

### 📌 기본 정보
- **상태**: planning
- **우선순위**: Medium
- **버전**: v0.1.0 (예정)
- **작성자**: (미정)
- **논의**: GitHub Discussion #130

### 🎯 목표
**spec-builder 에이전트가 하드코딩된 3개 파일 대신 `.moai/project/` 디렉토리의 모든 markdown 파일을 자동으로 발견하고 참조**하여, 더 유연한 프로젝트 문서 구조 지원.

### 📋 주요 내용

#### 현재 제한사항
```python
# ❌ 하드코딩된 3개 파일만 참조
document_files = [
    ".moai/project/product.md",
    ".moai/project/structure.md",
    ".moai/project/tech.md"
]
```

#### 개선 후
```yaml
.moai/project/
├── product.md              # 비즈니스 요구사항
├── structure.md            # 시스템 아키텍처
├── tech.md                 # 기술 스택
├── ux-journey.md          # UX/UI 요구사항 (NEW)
├── testing-strategy.md    # QA 요구사항 (NEW)
├── api-spec.md            # API 계약 (NEW)
└── deployment-plan.md     # DevOps 요구사항 (NEW)
```

#### 설계 원칙
1. **Convention over Configuration**: 자동 발견
2. **Prioritization**: 기존 3개 파일 우선 처리
3. **Extensibility**: 무제한 문서 추가 가능
4. **Backward Compatibility**: 기존 3개 파일 구조 유지

#### 구현 범위
- `src/moai_adk/templates/.claude/agents/alfred/spec-builder.md` 업데이트
- 자동 스캔 로직 추가
- 문서 우선순위 정책 정의
- 사용 예시 및 문서화

### 📊 예상 규모
- **구현 시간**: 4-6시간
- **테스트 시간**: 2-3시간
- **총 예상**: 1주 이내

### 🚀 다음 단계
1. SPEC 상세 설계 작성
2. 자동 스캔 로직 설계
3. 문서 우선순위 정책 정의
4. 구현 및 테스트
5. 문서화

---

# 📈 예상 타임라인

```
현재: 2025-11-04

Week 1 (11월 4-10일)
├─ SPEC-BUGFIX-001: Cross-platform timeout (3-4일)
├─ SPEC-DOC-TAG-004: Phase 4 시작 (1-2일)
└─ SPEC-FEATURE-001: SPEC 작성 (1-2일)

Week 2 (11월 11-17일)
├─ SPEC-DOC-TAG-003: Batch 1-2 마이그레이션 (5일)
├─ SPEC-CHECKPOINT-EVENT-001: 정책 최종 확정 (1-2일)
└─ SPEC-HOOKS-003: 자동 복구 로직 (2-3일)

Week 3-4 (11월 18-12월 1일)
├─ SPEC-DOC-TAG-003: Batch 3-7 진행 (10일)
└─ SPEC-DOC-TAG-004: Pre-commit/CI-CD 완성 (5일)

예상 완료: **2025-12-05 (약 4-5주)**
```

---

# 🎯 권장 우선순위

### CRITICAL (즉시 시작)
1. **SPEC-BUGFIX-001** (Windows 호환성) - 기능 차단
2. **SPEC-DOC-TAG-003** (Phase 3) - Phase 4의 의존성

### HIGH (1주일 내)
3. **SPEC-DOC-TAG-004** (Phase 4) - 품질 게이트 필수
4. **SPEC-CHECKPOINT-EVENT-001** - 정책 최종 확정

### MEDIUM (2주일 내)
5. **SPEC-HOOKS-003** - 기능 개선
6. **SPEC-FEATURE-001** - 사용성 개선

---

**문서 종료**
