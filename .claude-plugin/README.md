# @DOC:PLUGIN-001 | SPEC: SPEC-PLUGIN-001.md

# MoAI-ADK Claude Code Plugin

🗿 **SPEC-First TDD Development Kit with Alfred SuperAgent**

MoAI-ADK의 공식 Claude Code 플러그인입니다. Alfred SuperAgent와 9개의 전문 에이전트가 체계적인 개발 워크플로우를 오케스트레이션합니다.

---

## 📦 설치 방법

### 마켓플레이스 추가

```bash
/plugin marketplace add modu-ai/moai-adk
```

또는 Git URL 사용:

```bash
/plugin marketplace add https://github.com/modu-ai/moai-adk.git
```

### 플러그인 활성화

```bash
/plugin install moai-adk@moai-adk
```

### 설치 확인

```bash
/plugin list
```

---

## 📂 플러그인 구조

```
.claude-plugin/
├── plugin.json           # 플러그인 매니페스트 (필수)
├── marketplace.json      # 마켓플레이스 정의
└── README.md             # 이 문서

hooks/
├── hooks.json            # 후크 설정 (PostToolUse, PreToolUse 등)
└── scripts/              # 후크 실행 스크립트
    ├── pre-write-guard.cjs   # 파일 쓰기 전 검증
    ├── tag-enforcer.cjs      # @TAG 시스템 강제
    ├── policy-block.cjs      # Bash 명령 정책 검사
    └── session-notice.cjs    # 세션 시작 알림

commands/alfred/          # Alfred 커맨드 (5개)
├── 1-spec.md            # SPEC 작성
├── 2-build.md           # TDD 구현
├── 3-sync.md            # 문서 동기화
├── 8-project.md         # 프로젝트 초기화
└── 9-update.md          # 플러그인 업데이트

agents/alfred/            # Alfred 에이전트 (9개)
├── spec-builder.md      # 🏗️ SPEC 작성 전문가
├── code-builder.md      # 💎 TDD 구현 전문가
├── doc-syncer.md        # 📖 문서 동기화 전문가
├── tag-agent.md         # 🏷️ TAG 시스템 관리
├── git-manager.md       # 🚀 Git 워크플로우
├── debug-helper.md      # 🔬 오류 진단
├── trust-checker.md     # ✅ TRUST 검증
├── cc-manager.md        # 🛠️ Claude Code 설정
└── project-manager.md   # 📋 프로젝트 관리

templates/                # 프로젝트 템플릿
├── .moai/               # MoAI-ADK 설정
│   ├── config.json      # 프로젝트 설정
│   ├── memory/          # 개발 가이드, SPEC 메타데이터
│   └── project/         # 프로젝트 정보
├── CLAUDE.md            # Claude Code 프로젝트 지침
└── .gitignore           # Git 무시 목록
```

---

## 🔑 핵심 개념

### 환경변수

- **${CLAUDE_PLUGIN_ROOT}**: 플러그인 루트 경로 (자동 설정)
  - Claude Code가 플러그인 로드 시 자동으로 설정
  - hooks.json의 모든 스크립트 경로에서 사용

### 후크 시스템

MoAI-ADK는 5가지 후크 타입을 활용합니다:

| 후크 타입           | 실행 시점          | 주요 용도                   |
|---------------------|-------------------|----------------------------|
| **PostToolUse**     | 도구 사용 후       | 결과 검증, 후처리           |
| **PreToolUse**      | 도구 사용 전       | 입력 검증, 정책 적용        |
| **SessionStart**    | 세션 시작         | 초기화, 상태 표시           |
| **UserPromptSubmit**| 프롬프트 제출     | 요청 전처리                |
| **SessionEnd**      | 세션 종료         | 정리, 요약 생성             |

**활성화된 후크 예시**:
- `PreToolUse`: 파일 쓰기 전 검증 (`Edit|Write|MultiEdit`)
- `PreToolUse`: Bash 명령 정책 검사 (`Bash`)
- `SessionStart`: 프로젝트 상태 알림 (`*`)

### TAG 시스템

MoAI-ADK의 핵심 추적 시스템:

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

- **@SPEC:ID**: EARS 방식 요구사항 명세 (.moai/specs/)
- **@TEST:ID**: TDD RED 단계 테스트 (tests/)
- **@CODE:ID**: TDD GREEN + REFACTOR 구현 (src/)
- **@DOC:ID**: Living Document 문서화 (docs/)

---

## 🚀 3단계 개발 워크플로우

### 1️⃣ SPEC 작성 (`/alfred:1-spec`)

**명세 없이는 코드 없다**

```bash
/alfred:1-spec "JWT 인증 시스템"
```

- EARS 구문으로 요구사항 작성
- SPEC 문서 생성 (.moai/specs/SPEC-{ID}/)
- feature/{SPEC-ID} 브랜치 생성
- Draft PR 생성 (Team 모드)

### 2️⃣ TDD 구현 (`/alfred:2-build`)

**테스트 없이는 구현 없다**

```bash
/alfred:2-build SPEC-AUTH-001
```

- 🔴 RED: 실패하는 테스트 작성 (@TEST:ID)
- 🟢 GREEN: 테스트 통과 구현 (@CODE:ID)
- ♻️ REFACTOR: 코드 품질 개선

### 3️⃣ 문서 동기화 (`/alfred:3-sync`)

**추적성 없이는 완성 없다**

```bash
/alfred:3-sync --auto-merge
```

- TAG 체인 검증 (@SPEC → @TEST → @CODE → @DOC)
- Living Document 생성
- PR Ready 전환 + 자동 머지 (Team 모드)

---

## 🛠️ Alfred 에이전트 생태계

Alfred는 9명의 전문 에이전트를 조율합니다:

| 에이전트            | 역할          | 전문 영역               | 호출 방법            |
|--------------------|--------------|------------------------|---------------------|
| **spec-builder** 🏗️ | 시스템 아키텍트 | SPEC 작성, EARS 명세     | `/alfred:1-spec`    |
| **code-builder** 💎 | 수석 개발자    | TDD 구현, 코드 품질      | `/alfred:2-build`   |
| **doc-syncer** 📖   | 테크니컬 라이터 | 문서 동기화, Living Doc | `/alfred:3-sync`    |
| **tag-agent** 🏷️    | 지식 관리자    | TAG 시스템, 추적성       | `@agent-tag-agent`  |
| **git-manager** 🚀  | 릴리스 엔지니어 | Git 워크플로우, 배포     | `@agent-git-manager`|
| **debug-helper** 🔬 | 트러블슈팅 전문가| 오류 진단, 해결         | `@agent-debug-helper`|
| **trust-checker** ✅| 품질 보증 리드 | TRUST 검증, 성능/보안    | `@agent-trust-checker`|
| **cc-manager** 🛠️   | 데브옵스 엔지니어| Claude Code 설정        | `@agent-cc-manager` |
| **project-manager** 📋| 프로젝트 매니저 | 프로젝트 초기화         | `/alfred:8-project` |

---

## 📖 TRUST 5원칙

Alfred가 모든 코드에 적용하는 품질 기준:

- **T**est First: 언어별 최적 도구 (Jest/Vitest, pytest, JUnit 등)
- **R**eadable: 언어별 린터 (ESLint/Biome, ruff, golint 등)
- **U**nified: 타입 안전성 또는 런타임 검증
- **S**ecured: 언어별 보안 도구 및 정적 분석
- **T**rackable: CODE-FIRST @TAG 시스템 (코드 직접 스캔)

---

## 🌐 지원 언어

MoAI-ADK는 범용 개발 프레임워크입니다:

**백엔드**:
- TypeScript, Python, Java, Go, Rust
- C#, PHP, Ruby, Kotlin

**모바일**:
- Flutter (Dart)
- React Native (TypeScript)
- iOS (Swift)
- Android (Kotlin)

**기타**:
- Shell Script, SQL, YAML, JSON

---

## 🔧 설치 및 업데이트

### 플러그인 업데이트 확인

```bash
/alfred:9-update --check
```

### 자동 업데이트

```bash
/alfred:9-update --force
```

### 품질 검사

```bash
/alfred:9-update --check-quality
```

---

## 📚 참고 문서

- **SPEC-PLUGIN-001**: 이 플러그인 구조 설계 명세
- **development-guide.md**: MoAI-ADK 개발 가이드
- **spec-metadata.md**: SPEC 메타데이터 표준
- **CLAUDE.md**: Claude Code 프로젝트 지침

---

## 📝 라이선스

MIT © MoAI Team

## 🔗 링크

- **GitHub**: https://github.com/modu-ai/moai-adk
- **홈페이지**: https://moai-adk.vercel.app
- **문서**: https://moai-adk.vercel.app/docs
