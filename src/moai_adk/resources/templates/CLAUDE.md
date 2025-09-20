# MoAI-ADK (MoAI Agentic Development Kit)

**Spec-First TDD 개발 가이드 (사실·증빙 우선)**

## 핵심 철학
- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **GitFlow 지원**: Git 작업을 가이드/스크립트로 보조
- **Living Document 지향**: 코드·문서 동기화를 절차로 지원
- **Traceability 강화**: 16-Core @TAG로 추적성 향상(보장 아님)

## 사실성 · 증빙 · 안전한 표현 규칙

- 검증 없는 수치·효과·ROI·비교 금지. 측정 로그/출처 없으면 언급하지 않는다.
- 외부 프로젝트 비교(예: spec-kit vs MoAI-ADK)는 공신력 있는 출처나 재현 가능한 실험 로그가 있을 때만 요약한다.
- 절대 표현 금지: "100%", "완전 자동", "보장", "강제" → "기본 목표", "시도", "지원", "권장"으로 완화한다.
- "자동"이라는 표현은 실제 자동 실행 결과(명령/로그/커밋 등) 증빙을 함께 제시할 때만 사용한다. 그렇지 않으면 "자동화를 지원"이라고 적는다.
- 테스트 커버리지는 "기본 목표값"으로 표기하고, 측정 명령과 결과 요약(예: pytest/coverage 리포트)을 함께 제시한다.
- Git/PR/CI 관련 기능은 환경 의존(gh CLI, 권한 등)을 명시하고, 실패 시 수동 절차를 안내한다.
- 가정·제약·미확인은 반드시 별도 블록에 드러낸다.

## 출력 형식 표준(할루시네이션 방지)

출력은 아래 구조를 따른다. 각 항목은 생략하지 않는다.
- Results: 실제 수행 결과만(파일 경로, 커밋 해시, 테스트 통과/실패 요약)
- Evidence: 실행 명령과 핵심 로그 스니펫(필요 시 3~5줄 발췌)과 파일 참조
- Plan: 다음 단계(검증 가능한 작업 단위)
- Assumptions/Constraints: 전제, 환경 의존성(예: gh, pytest), 제한
- Notes: 미확인/추정 사항은 명시적으로 표시하고 결론에 포함하지 않는다.

## 3단계 개발 워크플로우

```bash
# 1. 명세 작성 (+옵션: 브랜치/PR 생성 지원)
/moai:1-spec "JWT 기반 사용자 인증"

# 2. TDD 구현(RED→GREEN→REFACTOR 권장 패턴)
/moai:2-build

# 3. 문서 동기화(+옵션: PR 상태 전환 지원)
/moai:3-sync
```

## 핵심 에이전트

| 에이전트 | 역할 | 자동화 기능 |
|----------|------|------------|
| **spec-builder** | EARS 명세 작성 | 브랜치/PR 생성을 지원(환경 의존) |
| **code-builder** | TDD 구현 | Red-Green-Refactor 커밋 패턴 권장 |
| **doc-syncer** | 문서 동기화 | PR 상태 전환/라벨링 지원(환경 의존) |

## TDD 커밋 단계(권장 패턴)

**SPEC 단계(예시 4단계)**
1. 명세 작성
2. User Stories 추가
3. 수락 기준 정의
4. 명세 완성

**BUILD 단계(예시 3단계)**
5. RED: 실패 테스트 작성
6. GREEN: 최소 구현
7. REFACTOR: 품질 개선

## 16-Core @TAG 시스템(추적성 강화 용도)

```markdown
@REQ:USER-AUTH-001 → @DESIGN:JWT-001 → @TASK:API-001 → @TEST:UNIT-001
```

**카테고리별 태그**
- **SPEC**: REQ, DESIGN, TASK
- **STEERING**: VISION, STRUCT, TECH, ADR
- **IMPLEMENTATION**: FEATURE, API, TEST, DATA
- **QUALITY**: PERF, SEC, DEBT, TODO

## Constitution 5원칙(목표/원칙)

1. **Simplicity**: 프로젝트 복잡도 ≤ 3
2. **Architecture**: 라이브러리 기반 설계
3. **Testing**: TDD 필수
4. **Observability**: 구조화 로깅
5. **Versioning**: 시맨틱 버전

## 언어 자동 감지(시도)

세션 시작 시 자동 감지되는 언어별 권장 도구:
- **Python**: pytest, ruff, black
- **JavaScript/TypeScript**: npm test, eslint, prettier
- **Go**: go test, gofmt
- **Rust**: cargo test, rustfmt
- **Java**: gradle/maven test
- **Kotlin**: junit, detekt, ktlint
- **.NET**: dotnet test
- **Swift**: XCTest, swift-format, SwiftLint
- **Dart/Flutter**: flutter test, flutter analyze, dart format
- **React Native**: jest, detox, eslint

감지 스크립트: `.claude/hooks/moai/language_detector.py` (감지 실패 시 수동 선택 안내)

## 프로젝트 구조(예시)

```
프로젝트/
├── .claude/               # Claude Code 설정
│   ├── commands/moai/     # 커스텀 명령어
│   ├── agents/moai/       # 에이전트 정의
│   ├── hooks/moai/        # 자동화 훅
│   ├── memory/            # 공유 메모리
│   └── settings.json      # 권한 설정
├── .moai/                 # MoAI 시스템
│   ├── specs/             # SPEC 문서
│   ├── memory/            # Constitution
│   ├── scripts/           # 검증 스크립트
│   └── config.json        # 설정
└── CLAUDE.md              # 이 파일
```

## 코드 작성 규칙

**크기 제한**
- 파일 ≤ 300 LOC
- 함수 ≤ 50 LOC
- 매개변수 ≤ 5
- 복잡도 ≤ 10

**품질 원칙**
- 명시적 코드 작성
- 섣부른 추상화 금지
- 의도를 드러내는 이름
- 가드절 우선
- 부수효과 격리
- 구조화 로깅
- 입력 검증

**테스트 규칙**
- 새 코드 = 새 테스트
- 버그 수정 = 회귀 테스트
- 테스트는 독립적, 결정적
- 성공/실패 경로 포함

추가: 커버리지 목표는 프로젝트 설정에 따르며(예: 80~85%), 실제 수치는 측정 결과로만 보고한다.

## 금지/주의 표현 체크리스트

- 금지: 근거 없는 수치/비율/ROI/효과/비교/보장/절대어
- 금지: 실행하지 않은 과정/결과를 했다고 서술
- 주의: "자동", "완전", "항상" → 증빙 동반/표현 완화
- 주의: 외부 도구/권한(gh, CI, 권한) 의존성 명시

## 검증 방법 예시(증빙 템플릿)

- 테스트: `pytest -q` → 결과 요약, 실패 시나리오 포함
- 커버리지: `coverage run -m pytest && coverage report` → 수치 인용
- Git/PR(선택): `git status`, `git log -1 --oneline`, `gh pr status` → 사용 시 출력 스니펫
- 파일 증빙: `README.md`, `.moai/specs/SPEC-XXX/spec.md`, `tests/...` 경로를 결과에 명시

## 메모리 전략

MoAI-ADK는 **Constitution 5원칙**을 핵심 메모리로 사용합니다:

- **Constitution**: @.moai/memory/constitution.md - 5원칙 + 16-Core TAG 시스템
- **프로젝트 가이드**: @CLAUDE.md - 워크플로우 + 에이전트 시스템

**원칙**:

- Constitution 5원칙이 모든 개발 규칙을 정의
- CLAUDE.md가 실행 가이드 역할
- 핵심 지침만 메모리에 저장 (단순성 원칙)
- 세부사항은 문서 링크로 참조
- 민감정보 저장 금지
