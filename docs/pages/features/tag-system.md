---
---



## 🎯 시스템 개요


### 추적성 체인

```
    ↓
    ↓
    ↓
```

## 🏷️ TAG 카테고리

MoAI-ADK는 13개의 TAG 카테고리를 지원하여 모든 종류의 개발 작업을 체계적으로 분류합니다:

| 카테고리 | 설명 | 사용 예시 |
|---------|------|-----------|

## 🔧 TAG 정책 설정

### 강제 모드

```json
{
  "tags": {
    "policy": {
      "enforcement_mode": "strict",
      "require_spec_before_code": true,
      "require_test_for_code": true,
      "enforce_chains": true
    }
  }
}
```

### 필수 디렉토리

- `src/` - 소스 코드
- `tests/` - 테스트 코드
- `.moai/specs/` - SPEC 문서

### 선택적 디렉토리

- `CLAUDE.md`, `README.md` - 프로젝트 문서
- `.claude/` - 에이전트 설정
- `.moai/docs/` - 생성된 문서
- `docs/` - 사용자 문서

## 🚀 자동 기능

### 1. 언어별 디렉토리 감지

18개 프로그래밍 언어에 대한 자동 디렉토리 감지:

- **지원 언어**: Python, TypeScript, JavaScript, Go, Rust, Java, Kotlin, Swift, Dart, PHP, Ruby, C, C++, C#, Scala, R, SQL, Shell
- **감지 모드**: auto (언어 기반), manual (사용자 정의), hybrid (혼합)
- **제외 패턴**: tests/, docs/, node_modules/ 등 자동 제외

```json
{
  "code_directories": {
    "detection_mode": "auto",
    "auto_detect_from_language": true,
    "exclude_patterns": [
      "tests/",
      "node_modules/",
      "dist/",
      "build/"
    ]
  }
}
```

### 2. 자동 수정

3단계 위험도별 자동 수정 시스템:

- **SAFE** (자동 수정): 중복 TAG 제거, 형식 오류 수정
- **MEDIUM_RISK** (승인 필요): 복잡한 TAG 구조 수정
- **HIGH_RISK** (차단): 위험한 수정 요구

```json
{
  "auto_correction": {
    "enabled": true,
    "confidence_threshold": 0.8,
    "remove_duplicates": true,
    "backup_before_fix": true,
    "auto_fix_levels": {
      "safe": true,
      "medium_risk": false,
      "high_risk": false
    }
  }
}
```

### 3. SPEC 자동 생성

코드 분석을 통한 SPEC 템플릿 자동 생성:

- **신뢰도 계산**: 구조 30%, 도메인 40%, 문서화 30%
- **EARS 형식**: 표준 SPEC 구조 자동 생성
- **사용자 편집**: 생성된 템플릿 사용자 편집 요구

```json
{
  "auto_spec_generation": {
    "enabled": true,
    "mode": "template",
    "confidence_threshold": 0.6,
    "require_user_edit": true,
    "open_in_editor": true
  }
}
```

## 🔍 실시간 검증

### 실시간 TAG 검증

```bash
# 실시간 검증 활성화
"realtime_validation": {
  "enabled": true,
  "validation_timeout": 5,
  "enforce_chains": true,
  "quick_scan_max_files": 30
}
```

### 코드 스캔 정책

```bash
# TAG 검색 명령어

# 스캔 도구
- rg (ripgrep) - 기본
- grep - 대체
```

## 📊 연구 TAG 시스템

고급 연구 활동을 위한 특수 TAG 카테고리:

### 연구 카테고리

- **RESEARCH**: 연구 활동 및 조사
- **ANALYSIS**: 분석 및 평가
- **KNOWLEDGE**: 지식 축적 및 패턴
- **INSIGHT**: 인사이트 및 혁신

### 자동 탐지 패턴

```json
{
  "research_patterns": {
  }
}
```

## 💡 사용 예시

### 1. 기본 TAG 사용

```python
# src/auth_service.py
class AuthService:

    def login(self, username: str, password: str) -> str:
        # JWT 토큰 생성 및 반환
        pass
```

```python
# tests/test_auth.py
def test_login_success():
    # 테스트 구현
    pass
```

### 2. 연구 TAG 사용

```python
# performance_analysis.py
"""
API 성능 병목 현상 분석 결과

주요 발견:
- 데이터베이스 쿼리 최적화 필요
- 캐싱 전략 개선 제안

캐싱 패턴 모범 사례 정리
"""

def analyze_performance():
    pass
```

### 3. 추적성 체인

```markdown
# .moai/specs/SPEC-AUTH-001/spec.md

## 추적성

```

## 🛡️ 훅 통합

### SessionStart Hook

세션 시작 시 자동 TAG 검증:

```bash
📋 Configuration Health Check:
✅ TAG policy: Enforced
✅ Auto-correction: Enabled
✅ Real-time validation: Active
✅ Research tags: Configured
```

### PreToolUse Hook

도구 사용 전 TAG 유효성 검사:

- 파일 변경 전 TAG 연결성 확인
- 체인 누락 경고
- 자동 TAG 제안

## 📈 성능 메트릭

| 기능 | 성능 지표 |
|------|----------|
| **실시간 검증** | 5초 내 완료 |
| **자동 수정** | 80%+ 정확도 |
| **SPEC 생성** | 60% 신뢰도 임계값 |
| **디렉토리 감지** | 95%+ 정확도 |
| **연구 TAG 탐지** | 자동 패턴 매칭 |

## 🎯 모범 사례

### 1. TAG 명명 규칙

- **일관성**: 프로젝트 전체 동일한 규칙 사용

### 2. 체인 유지

```python
# 좋은 예시
def user_registration():
    # 구현
    pass

# 나쁜 예시
def user_registration():
    # TAG 없음 - 추적 불가
    pass
```

### 3. 문서화

```markdown
## 추적성 맵

| 구성요소 | TAG | 상태 |
|---------|-----|------|
```

## 🔄 통합 워크플로우

5. **검증**: 실시간 TAG 연결성 확인

