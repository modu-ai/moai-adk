---
title: TDD 개발 가이드
description: Test-Driven Development 완전 가이드 - RED, GREEN, REFACTOR 사이클로 안정적인 코드 작성
status: stable
---

# TDD (Test-Driven Development) 개발 가이드

**TDD (Test-Driven Development)**는 MoAI-ADK의 핵심 원칙입니다. 이 가이드에서는 RED-GREEN-REFACTOR 사이클을 통해 테스트 우선 개발을 실현하는 방법을 배웁니다.

## <span class="material-icons">article</span> **TDD란?**

Test-Driven Development는 다음 순서로 진행됩니다:

1. **RED**: 실패하는 테스트 작성
2. **GREEN**: 테스트를 통과시키는 최소한의 코드 작성
3. **REFACTOR**: 코드 품질 개선

이 사이클을 반복하면서 요구사항을 만족하는 안정적인 코드를 작성합니다.

## <span class="material-icons">list</span> **각 단계별 가이드**

### [RED 단계](red.md)
- 실패하는 테스트 작성
- 테스트 케이스 설계
- 경계값 및 예외 처리

### [GREEN 단계](green.md)
- 최소 구현 (YAGNI 원칙)
- 빠른 테스트 통과
- 성능 vs 기능 균형

### [REFACTOR 단계](refactor.md)
- 코드 정리 및 최적화
- SOLID 원칙 적용
- 가독성 향상

## <span class="material-icons">sync</span> **Alfred와 함께하는 TDD**

Alfred SuperAgent는 TDD 사이클을 자동화합니다:

- `/alfred:2-run SPEC-ID`: RED-GREEN-REFACTOR 자동 실행
- 각 단계별 자동 검증
- Git 커밋 자동화

[Alfred 워크플로우로 TDD 시작하기](../alfred/2-run.md)

## <span class="material-icons">analytics</span> **TDD의 이점**

| 항목 | 효과 |
|------|------|
| **테스트 커버리지** | 87%+ 자동 달성 |
| **버그 조기 발견** | 개발 중 95% 이상 검출 |
| **리팩토링 안전성** | 테스트로 인한 완벽한 보호 |
| **문서화** | 테스트 자체가 실행 가능한 문서 |
| **설계 개선** | 테스트 가능한 설계의 자동 형성 |

## <span class="material-icons">navigate_next</span> **다음 단계**

- [RED: 실패하는 테스트 작성](red.md)
- [GREEN: 최소 구현으로 통과](green.md)
- [REFACTOR: 코드 개선](refactor.md)
- [Alfred 2-run 워크플로우](../alfred/2-run.md)

---

**Learn more**: MoAI-ADK의 TDD 원칙은 SPEC-First 개발 철학의 핵심입니다. SPEC 정의 후 TDD로 구현하면 완벽하게 요구사항을 만족하는 코드가 완성됩니다.
