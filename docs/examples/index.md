# Examples

MoAI-ADK 실전 예제 모음입니다.

## 목차

- [**Todo App 풀스택**](todo-app-fullstack.md) - ⭐ MoAI-ADK 전체 워크플로우 실전 예제
- [Auth System](auth-system.md) - 인증 시스템 구축 예제
- [API Endpoint](api-endpoint.md) - REST API 엔드포인트 예제

## 예제 구성

각 예제는 다음 구조로 제공됩니다:

1. **SPEC 작성**: 요구사항 명세 (EARS 방식)
2. **TDD 구현**: RED-GREEN-REFACTOR 사이클
3. **문서 동기화**: Living Document 생성
4. **TAG 추적성**: @TAG 체인 검증

## 시작하기

```bash
# 프로젝트 초기화
npx moai-adk init example-project
cd example-project

# Alfred로 개발 시작
/alfred:1-spec "JWT 인증 시스템"
/alfred:2-build SPEC-AUTH-001
/alfred:3-sync
```

## 더 많은 예제

GitHub 저장소의 [examples](https://github.com/modu-ai/moai-adk/tree/main/examples) 디렉토리를 참조하세요.
