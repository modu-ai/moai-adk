# MoAI-ADK Output Styles 시스템

## 🎨 5개 최적화된 스타일 구성

MoAI-ADK는 다양한 개발 상황에 맞는 5개의 맞춤형 출력 스타일을 제공합니다.

### 스타일 개요

| 스타일 | 대상 | 특징 | 사용 시기 |
|--------|------|------|----------|
| **expert** | 숙련된 개발자 | 간결하고 효율적 | 빠른 개발 |
| **beginner** | 초보 개발자 | 상세한 설명과 단계별 안내 | 학습 과정 |
| **study** | 심화 학습자 | 깊이 있는 원리와 심화 학습 | 기술 연구 |
| **mentor** | 1:1 멘토링 | 페어 프로그래밍과 멘토링 | 코드 리뷰 |
| **audit** | 품질 관리 | 코드 품질 지속적 검증 개선 | 품질 검증 |

## 스타일별 상세 기능

### 1. expert.md - 전문가 모드
**특징**: 간결하고 효율적인 전문가 모드

```markdown
# 전문가 모드 특징
- 핵심 정보만 간결하게 제공
- 코드 중심의 설명
- 고급 패턴과 최적화 기법
- 빠른 의사결정 지원
```

**출력 예시**:
```
✅ JWT auth implemented
📊 Coverage: 95% (+15%)
🔧 Next: Rate limiting
```

### 2. beginner.md - 초보자 모드
**특징**: 상세한 설명과 단계별 안내

```markdown
# 초보자 모드 특징
- 각 단계의 상세한 설명
- 왜 그렇게 하는지 이유 설명
- 일반적인 실수와 해결법
- 용어 설명과 참고 자료
```

**출력 예시**:
```
🎯 단계 1: JWT 라이브러리 설치
JWT는 JSON Web Token의 약자로, 사용자 인증에 사용됩니다.

📝 설치 명령어:
npm install jsonwebtoken

💡 설명: 이 라이브러리는...
⚠️ 주의: 비밀키는 환경변수에...
```

### 3. study.md - 심화 학습 모드
**특징**: 깊이 있는 원리와 심화 학습

```markdown
# 심화 학습 모드 특징
- 기술의 원리와 내부 동작 설명
- 다양한 접근 방법 비교
- 성능과 보안 관점 분석
- 실제 프로덕션 사례
```

**출력 예시**:
```
🔬 JWT 토큰 구조 분석

Header.Payload.Signature 형태로 구성:
- Header: 알고리즘 정보 (HMAC SHA256)
- Payload: 클레임 정보 (사용자 ID, 권한)
- Signature: 변조 방지 서명

🔐 보안 고려사항:
1. 토큰 만료 시간 설정
2. Refresh Token 전략
3. XSS/CSRF 방어
```

### 4. mentor.md - 멘토링 모드
**특징**: 1:1 멘토링과 페어 프로그래밍

```markdown
# 멘토링 모드 특징
- 대화형 설명과 질문
- 단계별 코드 리뷰
- 모범 사례와 안티패턴
- 개별 맞춤 조언
```

**출력 예시**:
```
👨‍🏫 멘토 조언

현재 코드를 보니 JWT 구현이 잘 되어있네요!
몇 가지 개선점을 제안해드릴게요:

1. 토큰 저장 위치
   현재: localStorage
   제안: httpOnly 쿠키
   이유: XSS 공격 방어

💬 질문: 왜 localStorage가 위험한지 알고 계신가요?
```

### 5. audit.md - 품질 검증 모드
**특징**: 코드 품질 지속적 검증 개선

```markdown
# 품질 검증 모드 특징
- 품질 지표 중심 분석
- 잠재적 이슈 식별
- 개선 우선순위 제시
- 메트릭 기반 평가
```

**출력 예시**:
```
📊 코드 품질 분석 보고서

✅ 통과 항목:
- 테스트 커버리지: 85% (기준: 80%)
- 복잡도: 평균 3.2 (기준: < 10)

⚠️ 개선 필요:
- 함수 길이: authenticate() 45줄 (권장: < 20줄)
- 순환 복잡도: validateToken() 12 (기준: < 10)

🎯 개선 제안:
1. authenticate() 함수 분리
2. 에러 핸들링 통합
```

## 스타일 설치 및 사용

### 자동 설치
모든 스타일은 `moai init` 시 자동으로 설치됩니다:

```
.claude/output-styles/
├── expert.md
├── beginner.md
├── study.md
├── mentor.md
└── audit.md
```

### 스타일 전환
```bash
# Claude Code에서 스타일 변경
@use expert     # 전문가 모드
@use beginner   # 초보자 모드
@use study      # 심화 학습 모드
@use mentor     # 멘토링 모드
@use audit      # 품질 검증 모드
```

## 스타일 간 자동 전환

### 상황별 자동 전환
```python
# 자동 전환 로직
def auto_switch_style(context):
    if context.is_debugging():
        return "expert"
    elif context.is_learning():
        return "study"
    elif context.is_reviewing():
        return "audit"
    elif context.has_junior_dev():
        return "mentor"
    else:
        return "beginner"
```

### Hook 연동
```python
# session_start_notice.py에서 자동 감지
if detect_user_level() == "expert":
    suggest_style("expert")
elif current_task_type() == "code_review":
    suggest_style("audit")
```

## 팀 협업 지원

### 팀별 기본 스타일 설정
```json
// .claude/settings.json
{
  "team_settings": {
    "default_style": "mentor",
    "pair_programming_style": "mentor",
    "code_review_style": "audit",
    "onboarding_style": "beginner"
  }
}
```

### 프로젝트 단계별 스타일
```json
{
  "pipeline_styles": {
    "SPECIFY": "study",
    "PLAN": "expert",
    "TASKS": "beginner",
    "IMPLEMENT": "mentor"
  }
}
```

Output Styles 시스템은 **개발자의 경험 수준**과 **현재 작업 상황**에 최적화된 맞춤형 지원을 제공합니다.