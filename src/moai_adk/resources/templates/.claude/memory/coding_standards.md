# 코딩 기준(요약)

언어/프레임워크별 세부 규칙은 `@.moai/memory/engineering-standards.md`에 통합되어 있습니다. `.claude/memory`에는 핵심 요약만 제공합니다.

## 핵심 원칙
- 작은 단위: 파일 ≤ 300 LOC, 함수 ≤ 50 LOC, 파라미터 ≤ 5, 순환 복잡도 < 10.
- 구조: 입력 → 처리 → 반환, 가드절 우선, 부수효과는 경계 계층에 격리.
- 명시성: 의미 있는 네이밍, 상수 심볼화, 구조화 로깅(JSON) + 상관관계 ID.
- 테스트/보안: `@.claude/memory/tdd_guidelines.md`, `@.claude/memory/security_rules.md`, `@.claude/memory/shared_checklists.md` 준수.

### 참조
- 엔지니어링 표준 전문: `@.moai/memory/engineering-standards.md`
- 운영 메모: `@.moai/memory/operations.md`
