---
name: moai:debug
description: 일반 오류 디버깅 및 Constitution 위반 검사. 구조화된 진단과 해결책 제시를 담당합니다.
argument-hint: "오류내용" | --constitution-check
allowed-tools: Read, Write, Edit, MultiEdit, Grep, Glob, Bash, TodoWrite
model: sonnet
---

# /moai:debug — 통합 디버깅 시스템

## 기능
MoAI-ADK의 통합 디버깅 시스템으로 두 가지 핵심 모드를 제공합니다:
1. **일반 오류 디버깅**: 코드/Git/설정 오류 분석
2. **Constitution 위반 검사**: 5원칙 준수도 검증

## 에이전트 협업 구조
- **전담 에이전트**: `debug-helper`가 모든 디버깅 작업을 담당합니다
- **단일 책임 원칙**: debug-helper는 진단만 수행하고 실제 수정은 전담 에이전트에게 위임
- **구조화된 출력**: 일관된 포맷으로 문제 분석과 해결책 제시

## 사용법

### 일반 오류 디버깅
```bash
# 코드 오류
/moai:debug "TypeError: 'NoneType' object has no attribute 'name'"

# Git 오류
/moai:debug "fatal: refusing to merge unrelated histories"

# 설정 오류
/moai:debug "PermissionError: [Errno 13] Permission denied"

# Import 오류
/moai:debug "ImportError: No module named 'requests'"
```

### Constitution 위반 검사
```bash
# 전체 Constitution 5원칙 검사
/moai:debug --constitution-check
```

## 처리 흐름

### 1. 일반 오류 디버깅 모드
```yaml
입력 분석:
- 오류 메시지 파싱
- 오류 유형 분류 (코드/Git/설정)
- 관련 파일/컨텍스트 검색

진단 과정:
- 오류 패턴 매칭
- 유사 사례 검색
- 영향도 평가
- 근본 원인 분석

출력 제공:
- 구조화된 분석 결과
- 단계별 해결 방안
- 전담 에이전트 추천
```

### 2. Constitution 위반 검사 모드
```yaml
검사 대상:
- Simplicity: 파일/함수 크기, 복잡도
- Architecture: 계층 분리, 의존성 방향
- Testing: 커버리지, 테스트 구조
- Observability: 로깅, 오류 처리
- Versioning: 시맨틱 버전, Git 태그

검사 과정:
- 프로젝트 파일 스캔
- 원칙별 준수도 측정
- 위반 사항 목록화
- 개선 우선순위 결정

출력 제공:
- 원칙별 준수율
- 위반 사항 상세 분석
- 개선 방안 제시
```

## 예상 출력 포맷

### 일반 오류 디버깅 결과
```markdown
🐛 디버그 분석 결과
━━━━━━━━━━━━━━━━━━━
📍 오류 위치: src/auth/login.py:45
🔍 오류 유형: TypeError
📝 오류 내용: 'NoneType' object has no attribute 'name'

🔬 원인 분석:
- 직접 원인: user 객체가 None 상태
- 근본 원인: 인증 실패 시 예외 처리 누락
- 영향 범위: 로그인 플로우 전체

🛠️ 해결 방안:
1. 즉시 조치: None 체크 추가 (user and user.name)
2. 권장 수정: Optional chaining 활용 (user?.name)
3. 예방 대책: 인증 실패 예외 처리 강화

🎯 다음 단계:
→ code-builder 호출 권장
→ 예상 명령: /moai:2-build (코드 수정)
```

### Constitution 위반 검사 결과
```markdown
🏛️ Constitution 검사 결과
━━━━━━━━━━━━━━━━━━━━━
📊 전체 준수율: 85%

❌ 위반 사항:
1. Simplicity (파일 크기)
   - 현재: src/core/analyzer.py 420줄 (목표: ≤300줄)
   - 권장: 모듈 분리 (Parser, Validator, Formatter)

2. Testing (커버리지)
   - 현재: 72% (목표: ≥85%)
   - 파일: src/auth/*.py 테스트 누락
   - 권장: 누락 테스트 추가

3. Architecture (순환 의존성)
   - 문제: auth ↔ utils 순환 참조
   - 권장: 공통 인터페이스 분리

✅ 준수 사항:
- Observability: 구조화 로깅 100% ✓
- Versioning: 시맨틱 버전 체계 ✓

🎯 개선 우선순위:
1. 테스트 커버리지 향상 (영향도: 높음)
2. 큰 파일 분리 (영향도: 중간)
3. 순환 의존성 해결 (영향도: 중간)

🔄 권장 다음 단계:
→ /moai:2-build (테스트 코드 추가)
→ /moai:1-spec (아키텍처 개선 명세)
```

## 에이전트 위임 규칙

debug-helper는 진단만 수행하고 실제 수정은 해당 전담 에이전트에게 위임합니다:

```yaml
코드 관련 문제:
- 대상: TypeError, ImportError, 테스트 실패 등
- 위임: code-builder 에이전트
- 명령: /moai:2-build

Git 관련 문제:
- 대상: merge conflict, push rejected 등
- 위임: git-manager 에이전트
- 명령: git 작업 관련 스크립트

설정 관련 문제:
- 대상: Permission denied, Hook 실패 등
- 위임: cc-manager 에이전트
- 명령: 설정 검토 및 수정

문서 관련 문제:
- 대상: TAG 불일치, 링크 오류 등
- 위임: doc-syncer 에이전트
- 명령: /moai:3-sync

복합 문제:
- 대상: 다중 에이전트 관련 오류
- 위임: 해당 커맨드 실행 권장
- 명령: /moai:1-spec, /moai:2-build, /moai:3-sync
```

## 제약사항

### debug-helper가 수행하지 않는 작업
- **코드 수정**: 실제 파일 편집 금지
- **Git 조작**: Git 명령 실행 금지
- **설정 변경**: Claude Code 설정 수정 금지
- **문서 갱신**: 문서 동기화 작업 금지

### 단일 책임 원칙
- **진단 전담**: 문제 분석과 해결책 제시만
- **위임 원칙**: 실제 수정은 전담 에이전트에게
- **구조화 출력**: 일관된 포맷으로 결과 제공

## 성능 지표

### 목표 성능
- **진단 정확도**: 95% 이상
- **해결책 유효성**: 90% 이상
- **응답 시간**: 30초 이내
- **적절한 에이전트 추천**: 95% 이상

### 품질 보장
- 구조화된 출력 포맷 유지
- Constitution 5원칙 기반 검사
- 16-Core TAG 시스템 활용
- 명확한 다음 단계 제시

## 다음 단계

debug-helper의 진단 결과에 따라 적절한 전담 에이전트를 호출하여 실제 문제를 해결하세요:

```bash
# 진단 후 코드 수정이 필요한 경우
/moai:2-build

# 진단 후 문서 동기화가 필요한 경우
/moai:3-sync

# 진단 후 새로운 명세가 필요한 경우
/moai:1-spec
```

**debug-helper는 MoAI-ADK의 통합 진단 센터로서 모든 문제를 체계적으로 분석하고 올바른 해결 방향을 제시합니다.**