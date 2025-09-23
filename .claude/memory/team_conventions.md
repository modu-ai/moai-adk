# MoAI-ADK Team Conventions

## 개발 원칙

### TRUST 5원칙 기반 개발

- **Test First**: RED-GREEN-REFACTOR 사이클 엄격 준수
- **Readable**: 코드는 미래의 나를 위해 명확하게 작성
- **Unified**: 계층 분리와 책임 분산으로 통합 설계
- **Secured**: 구조화된 로깅과 입력 검증 필수
- **Trackable**: 16-Core @TAG 시스템으로 완전한 추적성

## 코딩 스타일

### Python (주 언어)

- **포매터**: ruff format (black 호환)
- **린터**: ruff check (flake8, isort, pyupgrade 통합)
- **타입 체킹**: mypy strict 모드
- **보안 스캔**: bandit

### 명명 규칙

- **함수/변수**: snake_case
- **클래스**: PascalCase
- **상수**: UPPER_SNAKE_CASE
- **프라이빗**: \_leading_underscore

### 파일 구조

- **최대 라인 수**: 파일 ≤ 300 LOC, 함수 ≤ 50 LOC
- **복잡도**: Cyclomatic Complexity ≤ 10
- **매개변수**: 함수당 ≤ 5개

## Git 규칙

### 브랜치 전략

- **메인**: develop (기본 브랜치)
- **기능**: feature/SPEC-XXX-{feature-name}
- **핫픽스**: hotfix/{issue-number}
- **릴리스**: release/v{version}

### 커밋 메시지

```
🎯 [SPEC-001] Add user authentication

- RED: 실패하는 테스트 작성
- GREEN: 최소 구현으로 테스트 통과
- REFACTOR: 코드 품질 개선

@REQ:AUTH-001 @TEST:UNIT-001
```

### PR 규칙

- **Draft**: 개발 중 상태
- **Ready**: 리뷰 준비 완료
- **필수 확인**: TRUST 5원칙 + 테스트 커버리지 85%+

## 테스트 전략

### TDD 사이클

1. **RED**: 실패하는 테스트 먼저 작성
2. **GREEN**: 최소한의 코드로 테스트 통과
3. **REFACTOR**: 품질 개선 (테스트 통과 상태 유지)

### 커버리지 정책

- **단위 테스트**: ≥ 90%
- **통합 테스트**: ≥ 80%
- **E2E 테스트**: 주요 시나리오 100%

### 테스트 구조

```
tests/
├── unit/           # 단위 테스트
├── integration/    # 통합 테스트
├── e2e/           # E2E 테스트
└── fixtures/      # 테스트 데이터
```

## 16-Core @TAG 시스템

### TAG 카테고리

- **SPEC**: @REQ, @DESIGN, @TASK
- **PROJECT**: @VISION, @STRUCT, @TECH, @ADR
- **IMPLEMENTATION**: @FEATURE, @API, @TEST, @DATA
- **QUALITY**: @PERF, @SEC, @DEBT, @TODO

### TAG 연결 규칙

```
@REQ:USER-AUTH-001 → @DESIGN:JWT-001 → @TASK:API-001 → @TEST:UNIT-001
```

### TAG 검증

- **필수**: 모든 SPEC에 연결 체인 완성
- **금지**: 순환 참조, 고아 태그
- **권장**: 의미있는 설명과 컨텍스트

## 문서화 규칙

### Living Document 원칙

- **코드 변경 시**: 관련 문서 동시 업데이트
- **동기화 도구**: `/moai:3-sync` 명령어 활용
- **검증**: 문서-코드 일관성 100%

### 문서 구조

```
docs/
├── api/           # API 문서 (자동 생성)
├── architecture/  # 아키텍처 다이어그램
├── guides/        # 사용자 가이드
└── specs/         # 기술 명세
```

## 보안 정책

### 민감 정보 관리

- **시크릿**: ~/.moai/secrets/ 디렉토리 사용
- **환경 변수**: .env 파일 Git 제외
- **로깅**: 민감 정보 자동 마스킹 (`***redacted***`)

### 권한 관리

- **최소 권한**: 필요한 권한만 허용
- **정기 검토**: 월 1회 권한 검토
- **감사 로그**: 모든 권한 변경 기록

## 코드 리뷰

### 리뷰 체크리스트

- [ ] TRUST 5원칙 준수
- [ ] 테스트 커버리지 85% 이상
- [ ] @TAG 추적성 완성
- [ ] 보안 취약점 없음
- [ ] 성능 영향 검토

### 리뷰어 할당

- **자동**: GitHub CODEOWNERS 활용
- **수동**: 도메인 전문가 지정
- **필수**: 최소 1명 승인 후 머지

---

이 규칙들은 MoAI-ADK 프로젝트의 품질과 일관성을 보장하기 위한 필수 가이드라인입니다.
