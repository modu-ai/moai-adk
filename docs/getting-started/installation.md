# 설치

## 사전 요구사항

- Node.js 18.0 이상
- Bun 1.2.0 이상 (권장)
- Git

## npm으로 설치

```bash
npm install -g moai-adk
```

## Bun으로 설치 (권장)

```bash
bun install -g moai-adk
```

## 설치 확인

```bash
moai --version
moai doctor
```

`moai doctor` 명령어는 시스템 요구사항을 확인하고 모든 것이 올바르게 설정되었는지 검증합니다.

## 시스템 진단

MoAI-ADK는 프로젝트 언어를 자동으로 감지하고 필요한 개발 도구를 추천합니다:

```bash
moai doctor
```

**지능형 언어 감지:**
- JavaScript/TypeScript: npm, TypeScript, ESLint
- Python: pytest, ruff, black
- Java: Maven/Gradle, JUnit
- Go: go test, golint

## 문제 해결

설치 중 문제가 발생하면:

1. Node.js/Bun 버전 확인
2. 전역 설치 권한 확인
3. `moai doctor`로 시스템 진단 실행
4. GitHub Issues에 문제 보고

## 다음 단계

- [빠른 시작](/getting-started/quick-start) 가이드 확인
- [3단계 워크플로우](/guide/workflow) 학습