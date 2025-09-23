# MoAI-ADK (MoAI Agentic Development Kit)

**Spec-First TDD 개발 가이드**

## 핵심 철학
- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **GitFlow 지원**: Git 작업을 가이드/스크립트로 보조
- **Living Document 지향**: 코드·문서 동기화를 절차로 지원
- **Traceability 강화**: 16-Core @TAG로 추적성 향상(보장 아님)

## 4단계 개발 워크플로우

```bash
# 0. 프로젝트 문서 초기화 (product/structure/tech)
/moai:0-project

# 1. 명세 작성 (auto: 프로젝트 문서 기반 제안)
/moai:1-spec

# 2. TDD 구현(RED→GREEN→REFACTOR 권장 패턴)
/moai:2-build

# 3. 문서 동기화(+옵션: PR 상태 전환 지원)
/moai:3-sync
```

## 핵심 에이전트

### MoAI 핵심 4개 (워크플로우 에이전트)

| 에이전트 | 역할 | 자동화 기능 |
|----------|------|------------|
| **spec-builder** | EARS 명세 작성 | 브랜치/PR 생성을 지원(환경 의존) |
| **code-builder** | TDD 구현 | Red-Green-Refactor 커밋 패턴 권장 |
| **doc-syncer** | 문서 동기화 | PR 상태 전환/라벨링 지원(환경 의존) |
| **cc-manager** | Claude Code 관리 | 설정 최적화/권한 문제 해결 |

### Awesome 전문 2개 (고급 기능 에이전트)

| 에이전트 | 역할 | 사용 CLI |
|----------|------|-----------|
| **codex-bridge** | Codex CLI headless 호출, 브레인스토밍/디버깅 아이디어 수집 | `codex exec --model gpt-5-codex` |
| **gemini-bridge** | Gemini CLI headless 호출, 구조화된 분석/리뷰 결과 수집 | `gemini -m gemini-2.5-pro -p ... --output-format json` |

`brainstorming.enabled = true` 에서만 브리지 에이전트가 호출되며, CLI 설치·로그인은 사용자 동의 하에 진행합니다.

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
- **PROJECT**: VISION, STRUCT, TECH, ADR
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
│   ├── agents/            # 에이전트 정의
│   │   └── moai/          # MoAI 핵심 에이전트
│   ├── hooks/moai/        # 자동화 훅
│   ├── memory/            # 공유 메모리
│   └── settings.json      # 권한 설정
├── .moai/                 # MoAI 시스템
│   ├── project/           # 프로젝트 기본 문서 (product/structure/tech)
│   ├── specs/             # SPEC 문서
│   ├── memory/            # Constitution & 메모리 템플릿
│   ├── scripts/           # 검증/자동화 스크립트
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
