# @DOC:API-TAG-001 | Chain: @SPEC:DOCS-003 -> @DOC:API-001

# Core: TAG System

TAG 시스템 핵심 모듈입니다.

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
