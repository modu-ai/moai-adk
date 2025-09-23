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

### MoAI 핵심 5개 (워크플로우 에이전트)

| 에이전트         | 역할             | 자동화 기능                         |
| ---------------- | ---------------- | ----------------------------------- |
| **spec-builder** | EARS 명세 작성   | 브랜치/PR 생성을 지원(환경 의존)    |
| **code-builder** | TDD 구현         | Red-Green-Refactor 커밋 패턴 권장   |
| **doc-syncer**   | 문서 동기화      | PR 상태 전환/라벨링 지원(환경 의존) |
| **cc-manager**   | Claude Code 관리 | 설정 최적화/권한 문제 해결          |
| **debug-helper** | 오류 진단        | 일반 오류 분석 + Constitution 검사  |

### Awesome 전문 2개 (고급 기능 에이전트)

| 에이전트          | 역할                                                       | 사용 CLI                                               |
| ----------------- | ---------------------------------------------------------- | ------------------------------------------------------ |
| **codex-bridge**  | Codex CLI headless 호출, 브레인스토밍/디버깅 아이디어 수집 | `codex exec --model gpt-5-codex`                       |
| **gemini-bridge** | Gemini CLI headless 호출, 구조화된 분석/리뷰 결과 수집     | `gemini -m gemini-2.5-pro -p ... --output-format json` |

`brainstorming.enabled = true` 에서만 브리지 에이전트가 호출되며, CLI 설치·로그인은 사용자 동의 하에 진행합니다.

## 디버깅 시스템

### `/moai:debug` 명령어

- **일반 오류 디버깅**: `/moai:debug "오류내용"`
- **Constitution 위반 검사**: `/moai:debug --constitution-check`
- **진단 전문**: debug-helper 에이전트가 문제 분석과 해결책 제시
- **위임 원칙**: 실제 수정은 해당 전담 에이전트(code-builder, git-manager 등)에게 위임

## Git 작업 관리

### 자동 Git 처리 (권장)

모든 워크플로우 명령어에서 Git 작업이 자동으로 처리됩니다:

- `/moai:1-spec`: spec-builder + git-manager (브랜치 생성, 커밋)
- `/moai:2-build`: code-builder + git-manager (RED/GREEN/REFACTOR 커밋)
- `/moai:3-sync`: doc-syncer + git-manager (문서 동기화, PR 관리)

### 직접 Git 작업이 필요한 경우

**방법 1: git-manager 에이전트 직접 호출**

```
@agent-git-manager "체크포인트 생성"
@agent-git-manager "브랜치 생성: feature/new-feature"
@agent-git-manager "개인 모드 롤백"
```

**방법 2: 표준 Git 명령어 사용**

- 긴급상황이나 특수한 Git 작업이 필요한 경우만 사용
- 일반적인 워크플로우에서는 권장하지 않음

### Git 작업 원칙

- **99% 케이스**: 워크플로우 명령어 사용 (자동 처리)
- **1% 특수 케이스**: @agent-git-manager 직접 호출
- **긴급상황**: 표준 git 명령어

## TDD 커밋 단계(권장 패턴)

**SPEC 단계(예시 4단계)**

1. 명세 작성
2. User Stories 추가
3. 수락 기준 정의
4. 명세 완성

**BUILD 단계(예시 3단계)** 5. RED: 실패 테스트 작성 6. GREEN: 최소 구현 7. REFACTOR: 품질 개선

## 16-Core @TAG 시스템(추적성 강화 용도)

```markdown
@REQ:USER-AUTH-001 → @DESIGN:JWT-001 → @TASK:API-001 → @TEST:UNIT-001
```

**카테고리별 태그**

- **SPEC**: REQ, DESIGN, TASK
- **PROJECT**: VISION, STRUCT, TECH, ADR
- **IMPLEMENTATION**: FEATURE, API, TEST, DATA
- **QUALITY**: PERF, SEC, DEBT, TODO

## TRUST 5원칙 (개발 가이드)

**T** - **Test First** (테스트 우선): 코드 전에 테스트를 작성하라
**R** - **Readable** (읽기 쉽게): 미래의 나를 위해 명확하게 작성하라
**U** - **Unified** (통합 설계): 계층을 나누고 책임을 분리하라
**S** - **Secured** (안전하게): 로그를 남기고 검증하라
**T** - **Trackable** (추적 가능): 버전과 태그로 히스토리를 관리하라

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

## 검색 명령어 권장사항

MoAI-ADK에서는 효율적인 코드 검색을 위해 다음 가이드라인을 권장합니다:

**권장 검색 도구**

- **파일 내용 검색**: `rg` (ripgrep) 권장 - 빠르고 .gitignore 자동 준수
- **파일 이름 검색**: `rg --files -g "패턴"` 권장
- **정확한 패턴 매칭**: `rg -F "정확한문자열"` (고정 문자열 검색)

**사용 가능한 기존 명령어**

- `grep`, `find` 명령어도 여전히 사용 가능
- 단, 성능과 편의성 측면에서 `rg` 사용을 권장
- 복잡한 검색의 경우 Grep 도구 활용

**예시**

```bash
# 권장: ripgrep 사용
rg "function_name" --type py
rg --files -g "*.md" | rg "TODO"

# 사용 가능: 기존 명령어
grep -r "pattern" src/
find . -name "*.py" -exec grep "pattern" {} \;
```

## 메모리 전략

MoAI-ADK는 **TRUST 5원칙**을 핵심 메모리로 사용합니다:

- **MoAI 개발 가이드**: @.moai/memory/development-guide.md - TRUST 원칙 + 16-Core TAG 시스템
- **프로젝트 가이드**: @CLAUDE.md - 워크플로우 + 에이전트 시스템

**원칙**:

- TRUST 5원칙이 모든 개발 규칙을 정의
- CLAUDE.md가 실행 가이드 역할
- 핵심 지침만 메모리에 저장 (읽기 쉬운 원칙)
- 세부사항은 문서 링크로 참조
- 민감정보 저장 금지
