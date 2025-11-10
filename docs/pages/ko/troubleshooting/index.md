# 문제 해결 가이드

MoAI-ADK 사용 중 발생하는 문제들의 해결 방법입니다.

## 문제별 솔루션 찾기

<span class="material-icons">search</span> **문제 유형별 해결책**

### 설치 및 초기화 문제

- [설치 오류](https://adk.mo.ai.kr/troubleshooting/installation)
- [초기화 실패](https://adk.mo.ai.kr/troubleshooting/initialization)
- [환경 설정](https://adk.mo.ai.kr/troubleshooting/environment)

### Alfred 명령어 문제

- [명령 인식 불가](https://adk.mo.ai.kr/troubleshooting/command-not-found)
- [SPEC 생성 실패](https://adk.mo.ai.kr/troubleshooting/spec-creation)
- [TDD 사이클 오류](https://adk.mo.ai.kr/troubleshooting/tdd-errors)

### 개발 및 빌드 문제

- [테스트 실패](https://adk.mo.ai.kr/troubleshooting/test-failures)
- [의존성 오류](https://adk.mo.ai.kr/troubleshooting/dependency-errors)
- [빌드 오류](https://adk.mo.ai.kr/troubleshooting/build-errors)

### Git 및 배포 문제

- [Git 충돌](https://adk.mo.ai.kr/troubleshooting/git-conflicts)
- [배포 실패](https://adk.mo.ai.kr/troubleshooting/deployment-errors)
- [CI/CD 문제](https://adk.mo.ai.kr/troubleshooting/cicd-issues)

______________________________________________________________________

## 자주 묻는 질문 (FAQ)

<span class="material-icons">help</span> **FAQ 모음**

### 기본 사용법

**Q: MoAI-ADK를 처음 시작하려면 어떻게 해야 하나요?** A: [빠른 시작 가이드](getting-started/quick-start.md)를 참고하세요. 3분
안에 기본 설정을 마칠 수 있습니다.

**Q: SPEC-First가 무엇인가요?** A: [기본 개념](getting-started/concepts.md)에서 상세히 설명합니다. 간단히 말해, 코드를 작성하기
전에 명세서를 먼저 작성하는 방식입니다.

**Q: Alfred는 어떤 역할을 하나요?** A: [Alfred 워크플로우](guides/alfred/index.md)에서 확인하세요. Alfred는 19명의 AI 전문가
팀을 조율하는 슈퍼에이전트입니다.

### TDD 관련

**Q: TDD의 RED-GREEN-REFACTOR는 무엇인가요?** A: [TDD 가이드](guides/tdd/index.md)에서 각 단계를 상세히 설명합니다.

**Q: 테스트 커버리지는 얼마나 되어야 하나요?** A: MoAI-ADK는 **85% 이상의 테스트 커버리지**를 권장합니다.

### TAG 시스템

**Q: @TAG 시스템이 왜 필요한가요?** A: [TAG 시스템](guides/specs/tags.md)을 통해 SPEC, TEST, CODE, DOC을 모두 연결하여
완전한 추적성을 제공합니다.

______________________________________________________________________

## 일반적인 오류 메시지

<span class="material-icons">error</span> **자주 발생하는 오류**

### "Command not found: /alfred:1-plan"

**원인**: Claude Code가 Alfred 명령을 인식하지 못함

**해결책**:

```bash
# 1. Claude Code 재시작
exit
claude

# 2. 디렉토리 확인
ls .claude/commands/

# 3. 설정 새로고침
/alfred:0-project
```

### "SPEC file not found"

**원인**: SPEC 파일이 올바른 위치에 생성되지 않음

**해결책**:

```bash
# 프로젝트 상태 확인
moai-adk doctor

# .moai/ 디렉토리 권한 확인
ls -la .moai/

# 재초기화
rm -rf .moai
/alfred:0-project
```

### "Test coverage below 85%"

**원인**: 테스트 커버리지가 부족함

**해결책**:

```bash
# 현재 커버리지 확인
pytest --cov=src tests/

# 누락된 테스트 추가
# tests/test_*.py에 테스트 케이스 추가

# 다시 실행
pytest --cov=src tests/
```

______________________________________________________________________

## 🔧 시스템 진단

### 진단 도구 실행

```bash
# 전체 시스템 상태 확인
moai-adk doctor

# 상세 출력
moai-adk doctor --verbose
```

### 확인되는 항목

- Python 버전 및 의존성
- Git 설정 및 권한
- .moai/ 디렉토리 구조
- .claude/ 설정 파일
- 필수 도구 설치 여부

______________________________________________________________________

## 💬 추가 도움

### 커뮤니티

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) - 질문하고 아이디어 공유
- [Issue Tracker](https://github.com/modu-ai/moai-adk/issues) - 버그 보고

### 문서

- [온라인 문서](https://adk.mo.ai.kr) - 최신 정보
- [로컬 문서](index.md) - 오프라인 참고

### 피드백

```bash
# 문제 보고 (GitHub Issue 자동 생성)
/alfred:9-feedback
```

______________________________________________________________________

**도움이 되었나요?** [더 많은 질문](https://github.com/modu-ai/moai-adk/discussions)을 환영합니다!
