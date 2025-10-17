# @DOC:SEC-OVERVIEW-001 | Chain: @SPEC:DOCS-003 -> @DOC:SEC-001

# Security Overview

MoAI-ADK 보안 개요입니다.

## 보안 원칙

MoAI-ADK는 다음 보안 원칙을 따릅니다:

### 1. 최소 권한 원칙 (Principle of Least Privilege)

에이전트는 작업 수행에 필요한 최소한의 권한만 부여받습니다.

- 파일 시스템 접근 제한
- 네트워크 요청 최소화
- 민감한 데이터 접근 차단

### 2. 보안 기본값 (Secure by Default)

모든 기본 설정이 보안을 우선합니다.

- 템플릿 보안 검증 필수
- 자격증명 자동 제외 (.env, secrets)
- Hook 기반 보안 검증 활성화

### 3. 정기 감사 (Regular Audits)

보안 상태를 지속적으로 모니터링합니다.

- 의존성 취약점 스캔 (pip-audit)
- 코드 보안 분석 (bandit)
- 템플릿 무결성 검증

## 보안 계층

MoAI-ADK의 보안은 4계층으로 구성됩니다:

```
┌─────────────────────────────────────┐
│  1. 템플릿 보안 (Template Security) │
│     - 악성 코드 탐지                │
│     - 무결성 검증                   │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│  2. Hook 보안 (Hook Security)       │
│     - PreToolUseHook 검증           │
│     - 민감 파일 접근 차단           │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│  3. 런타임 보안 (Runtime Security)  │
│     - 자격증명 제외                 │
│     - 안전한 명령 실행              │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│  4. 의존성 보안 (Dependency Security)│
│     - pip-audit 스캔                │
│     - 버전 고정                     │
└─────────────────────────────────────┘
```

## 보안 기능

### 템플릿 보안 검증

모든 템플릿은 설치 전 보안 검증을 거칩니다.

```python
from moai_adk.core.template_security import validate_template

result = validate_template(template_path)
if result.is_safe:
    install_template(template_path)
else:
    print(f"⚠️ Security risk: {result.issues}")
```

### 자격증명 자동 제외

민감한 파일은 자동으로 Git에서 제외됩니다.

```
.env
.env.*
**/secrets/**
credentials.json
*.key
*.pem
```

### 보안 Hook

PreToolUseHook을 통해 위험한 작업을 차단합니다.

```python
class SecurityHook(PreToolUseHook):
    def execute(self, tool_name, tool_params):
        # .env 파일 쓰기 차단
        if tool_name == "Write":
            file_path = tool_params.get("file_path")
            if ".env" in file_path:
                return {"block": True}
        return {"block": False}
```

## 보안 체크리스트

프로젝트 보안을 위한 체크리스트:

- [ ] 템플릿 보안 검증 활성화
- [ ] 민감한 파일 .gitignore 추가
- [ ] Hook 기반 보안 검증 설정
- [ ] 의존성 정기 스캔 (pip-audit)
- [ ] 보안 패치 적용
- [ ] 자격증명 환경변수 사용
- [ ] 로그에 민감 정보 제외
- [ ] 정기 보안 감사

상세 내용은 [보안 체크리스트](checklist.md)를 참조하세요.

## 취약점 보고

보안 취약점 발견 시:

1. **즉시 보고**: security@moai.dev (비공개)
2. **재현 정보 제공**: 취약점 재현 방법
3. **패치 대기**: 24시간 내 응답
4. **공개 보류**: 패치 후 공개

## 관련 문서

- [Template Security](template-security.md)
- [Best Practices](best-practices.md)
- [Security Checklist](checklist.md)
