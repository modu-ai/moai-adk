# MoAI-ADK (MoAI Agentic Development Kit)

**Spec-First TDD 자동화 개발 시스템**

## 핵심 철학
- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **GitFlow 투명성**: Git을 몰라도 프로페셔널 워크플로우
- **Living Document**: 문서와 코드는 항상 동기화
- **Full Traceability**: 16-Core @TAG 시스템으로 추적

## 3단계 개발 워크플로우

```bash
# 1. 명세 작성 + 브랜치 + Draft PR
/moai:1-spec "JWT 기반 사용자 인증"

# 2. TDD 구현 + 자동 커밋
/moai:2-build

# 3. 문서 동기화 + PR Ready
/moai:3-sync
```

## 핵심 에이전트

| 에이전트 | 역할 | 자동화 기능 |
|----------|------|------------|
| **spec-builder** | EARS 명세 작성 | 브랜치 생성, Draft PR |
| **code-builder** | TDD 구현 | Red-Green-Refactor 커밋 |
| **doc-syncer** | 문서 동기화 | PR Ready, 리뷰어 할당 |

## 7단계 자동 커밋 시스템

**SPEC 단계 (4단계)**
1. 명세 작성
2. User Stories 추가
3. 수락 기준 정의
4. 명세 완성

**BUILD 단계 (3단계)**
5. RED: 실패 테스트 작성
6. GREEN: 최소 구현
7. REFACTOR: 품질 개선

## 16-Core @TAG 시스템

```markdown
@REQ:USER-AUTH-001 → @DESIGN:JWT-001 → @TASK:API-001 → @TEST:UNIT-001
```

**카테고리별 태그**
- **SPEC**: REQ, DESIGN, TASK
- **STEERING**: VISION, STRUCT, TECH, ADR
- **IMPLEMENTATION**: FEATURE, API, TEST, DATA
- **QUALITY**: PERF, SEC, DEBT, TODO

## Constitution 5원칙

1. **Simplicity**: 프로젝트 복잡도 ≤ 3
2. **Architecture**: 라이브러리 기반 설계
3. **Testing**: TDD 필수
4. **Observability**: 구조화 로깅
5. **Versioning**: 시맨틱 버전

## 언어 자동 감지

세션 시작 시 자동 감지되는 언어별 권장 도구:
- **Python**: pytest, ruff, black
- **JavaScript/TypeScript**: npm test, eslint, prettier
- **Go**: go test, gofmt
- **Rust**: cargo test, rustfmt
- **Java**: gradle/maven test
- **.NET**: dotnet test

감지 스크립트: `.claude/hooks/moai/language_detector.py`

## 프로젝트 구조

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
