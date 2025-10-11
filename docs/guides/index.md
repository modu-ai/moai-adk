# Guides

MoAI-ADK 개발 가이드입니다.

## 목차

- [Development Workflow](development-workflow.md) - 개발 워크플로우
- [SPEC Writing](spec-writing.md) - SPEC 작성 가이드
- [TDD Implementation](tdd-implementation.md) - TDD 구현 가이드
- [Context Engineering](context-engineering.md) - 컨텍스트 엔지니어링

## 3단계 개발 워크플로우

```bash
/alfred:1-spec     # SPEC 작성
/alfred:2-build    # TDD 구현
/alfred:3-sync     # 문서 동기화
```

## 핵심 원칙

1. **SPEC-First**: 명세 없이는 코드 없음
2. **TDD-First**: 테스트 없이는 구현 없음
3. **TRUST 5원칙**: 모든 코드는 TRUST 원칙 준수
4. **@TAG 추적성**: 코드와 문서의 완벽한 연결
