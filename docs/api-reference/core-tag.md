# @DOC:API-TAG-001 | Chain: @SPEC:DOCS-003 -> @DOC:API-001

# Core: TAG System

TAG 시스템 핵심 모듈입니다.

MoAI-ADK의 TAG 시스템은 주석 기반으로 구현되어 있으며, 별도의 Python 모듈이 아닌 코드 주석 패턴입니다. TAG는 `@SPEC:ID`, `@TEST:ID`, `@CODE:ID`, `@DOC:ID` 형식으로 소스 코드와 문서에 직접 작성됩니다.

자세한 TAG 시스템 사용법은 [TAG System Guide](../guides/concepts/tag-system.md)를 참조하세요.

## TAG 체인 구조

### @SPEC → @TEST → @CODE → @DOC
MoAI-ADK의 TAG 시스템은 코드 추적성을 보장하는 핵심 메커니즘입니다.

### TAG 형식
```
@SPEC:ID | @TEST:ID | @CODE:ID | @DOC:ID
```

### 관련 에이전트
- **tag-agent**: TAG 무결성 검증 및 체인 분석
- **doc-syncer**: TAG 기반 문서 동기화
