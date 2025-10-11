# Getting Started

MoAI-ADK에 오신 것을 환영합니다! 이 가이드는 5분 안에 MoAI-ADK를 시작할 수 있도록 도와드립니다.

## What is MoAI-ADK?

MoAI-ADK (Modu-AI Agentic Development Kit)는 SPEC-First TDD 방법론을 기반으로 한 범용 개발 도구입니다.

### 핵심 특징

- **SPEC-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **Alfred SuperAgent**: 9개의 전문 에이전트를 조율하는 중앙 오케스트레이터
- **Universal Language Support**: Python, TypeScript, Java, Go, Rust, Ruby, Dart, Swift, Kotlin 등 모든 주요 언어 지원
- **TAG Traceability**: `@SPEC → @TEST → @CODE → @DOC` 완벽한 추적성

---

## Prerequisites

MoAI-ADK를 사용하기 전에 다음 도구들이 설치되어 있어야 합니다:

### 필수 요구사항

| 도구             | 최소 버전 | 권장 버전  | 확인 명령어      |
| ---------------- | --------- | ---------- | ---------------- |
| **Node.js**      | 18.0.0+   | 20.0.0+    | `node --version` |
| **npm/pnpm/bun** | -         | bun 1.2.0+ | `bun --version`  |
| **Git**          | 2.0+      | 최신       | `git --version`  |
| **Claude Code**  | -         | 최신       | VSCode 확장      |

### 선택 요구사항

- **GitHub CLI**: PR 자동 생성 및 관리 (`gh` 명령어)

---

## Quick Installation

### Global Installation (권장)

::: code-group

```bash [bun]
bun add -g moai-adk
```

```bash [npm]
npm install -g moai-adk
```

```bash [pnpm]
pnpm add -g moai-adk
```

```bash [yarn]
yarn global add moai-adk
```

:::

### Local Installation (프로젝트별)

::: code-group

```bash [bun]
bun add -D moai-adk
```

```bash [npm]
npm install --save-dev moai-adk
```

```bash [pnpm]
pnpm add -D moai-adk
```

```bash [yarn]
yarn add -D moai-adk
```

:::

---

## Verify Installation

설치가 완료되면 다음 명령어로 확인하세요:

```bash
moai --version
# Expected output: v0.2.x
```

도움말 보기:

```bash
moai help
```

---

## Your First MoAI Project

### 1. 프로젝트 초기화

MoAI-ADK는 두 가지 방법으로 초기화할 수 있습니다:

::: code-group

```bash [새 프로젝트]
# 새 프로젝트 생성 (디렉토리 자동 생성)
moai init my-moai-project
```

```bash [기존 프로젝트]
# 기존 디렉토리에서 초기화
cd existing-project
moai init .
```

:::

#### 초기화 프로세스

`moai init my-moai-project` 실행 시 다음과 같이 진행됩니다:

**Welcome Screen** (시작 화면)

```
███╗   ███╗          █████╗ ██╗      █████╗ ██████╗ ██╗  ██╗
████╗ ████║ ██████╗ ██╔══██╗██║     ██╔══██╗██╔══██╗██║ ██╔╝
██╔████╔██║██╔═══██ ███████║██║ ███ ███████║██║  ██║█████╔╝
██║╚██╔╝██║██║   ██║██╔══██║██║ ╚══ ██╔══██║██║  ██║██╔═██╗
██║ ╚═╝ ██║╚██████╔╝██║  ██║██║     ██║  ██║██████╔╝██║  ██╗
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝
════════════════════════════════════════════════════════════

🗿 MoAI-ADK: Modu-AI's Agentic Development kit (v0.2.17) 🚀

copyleft 2024, Modu-AI / 모두의AI (https://mo.ai.kr)

🚀 Initializing my-moai-project project...
```

**Step 1: System Verification** (시스템 검증)

```
────────────────────────────────────────────────────────────
📋 Step 1: System Verification
────────────────────────────────────────────────────────────

🔍 Checking system requirements...

  ⚙️  Runtime:
    ✅ Git (2.50.1)
    ✅ Node.js (20.19.4)

  🛠️  Development:
    ✅ npm (10.8.2)

  📦 Optional:
    ✅ Git LFS (3.7.0)

────────────────────────────────────────────────────────────
  📊 Summary:
     Checks: 4 total
     Status: 4 passed
────────────────────────────────────────────────────────────

✅ All requirements satisfied!
```

**Step 2: Interactive Configuration** (대화형 설정)

```
────────────────────────────────────────────────────────────
⚙️  Step 2: Interactive Configuration
────────────────────────────────────────────────────────────

  Let's set up your project with a few questions...
  You can change these settings later in .moai/config.json



❓ Question [1/4]
→ Language Selection / 언어 선택
✔ Choose CLI language / CLI 언어를 선택하세요: 한국어
  💡 한국어가 선택되었습니다. 이후 모든 메시지가 한국어로 표시됩니다.

❓ Question [2/4]
→ 프로젝트 정보
✔ 프로젝트 이름: my-moai-project
  💡 폴더 이름과 프로젝트 식별자로 사용됩니다

❓ Question [3/4]
→ 개발 모드
✔ 모드 선택: Personal
  💡 Personal 모드: SPEC 파일이 로컬에 저장되며, 단순한 워크플로우

❓ Question [4/4]
→ 버전 관리
  Existing Git repository detected
    • Commits: 0
    • Branch:
  💡 Git repository detected. Will use existing configuration.

✅ Configuration Complete!

📋 Summary:
────────────────────────────────────────────────────────────
  Project Name:  my-moai-project
  Mode:          🧑 Personal
  Git:           ✓ Enabled
────────────────────────────────────────────────────────────



✅ Configuration saved
```

::: tip 개발 모드 선택

- **Personal**: SPEC 파일이 로컬에 저장, 단순한 워크플로우
- **Team**: 공유 저장소 연동, 협업 기능 활성화
:::

**Step 3: Installation** (설치)

5단계 진행률이 실시간으로 표시됩니다:

```
────────────────────────────────────────────────────────────
📦 Step 3: Installation
────────────────────────────────────────────────────────────

[░░░░░░░░░░░░░░░░░░░░] 0% - Phase 1: Preparation and backup...
[████░░░░░░░░░░░░░░░░] 20% - Phase 2: Creating directory structure...
[████████░░░░░░░░░░░░] 40% - Phase 3: Installing resources...
[████████████░░░░░░░░] 60% - Phase 4: Generating configurations...
[████████████████░░░░] 80% - Phase 5: Validation and finalization...
[████████████████████] 100% - Installation complete!
info: Installation succeeded
```

최종 완료 메시지:

```
────────────────────────────────────────────────────────────
✅ Initialization Completed Successfully!
────────────────────────────────────────────────────────────

📊 Summary:
  📁 Location:  /Users/your-path/my-moai-project
  📄 Files:     35 created
  ⏱️  Duration:  57ms

🚀 Next Steps:
  1. cd my-moai-project
  2. 💡 Tip: Run "claude" to start development with Claude Code

────────────────────────────────────────────────────────────
```

### 2. 프로젝트 구조 확인

초기화가 완료되면 **35개 파일**이 다음과 같은 구조로 생성됩니다:

```
my-moai-project/
├── .claude/                          # Claude Code 설정
│   ├── agents/alfred/                # 9개 전문 에이전트
│   │   ├── cc-manager.md
│   │   ├── code-builder.md
│   │   ├── debug-helper.md
│   │   ├── doc-syncer.md
│   │   ├── git-manager.md
│   │   ├── project-manager.md
│   │   ├── spec-builder.md
│   │   ├── tag-agent.md
│   │   └── trust-checker.md
│   ├── commands/alfred/              # 5개 Alfred 커맨드
│   │   ├── 0-project.md              # 프로젝트 초기화
│   │   ├── 1-spec.md                 # SPEC 작성
│   │   ├── 2-build.md                # TDD 구현
│   │   ├── 3-sync.md                 # 문서 동기화
│   │   └── 9-update.md               # 업데이트
│   ├── hooks/alfred/                 # 4개 Git 훅
│   │   ├── policy-block.cjs          # 정책 차단
│   │   ├── pre-write-guard.cjs       # 쓰기 전 검증
│   │   ├── session-notice.cjs        # 세션 알림
│   │   └── tag-enforcer.cjs          # TAG 강제
│   ├── logs/                         # 로그 디렉토리
│   ├── output-styles/alfred/         # 4개 출력 스타일
│   │   ├── alfred-pro.md
│   │   ├── beginner-learning.md
│   │   ├── pair-collab.md
│   │   └── study-deep.md
│   └── settings.json                 # Claude Code 설정
│
├── .moai/                            # MoAI-ADK 설정
│   ├── config.json                   # 프로젝트 설정
│   ├── memory/                       # 개발 가이드
│   │   ├── development-guide.md      # 개발 규칙
│   │   └── spec-metadata.md          # SPEC 메타데이터
│   ├── project/                      # 프로젝트 정보
│   │   ├── product.md                # 제품 정보
│   │   ├── structure.md              # 프로젝트 구조
│   │   └── tech.md                   # 기술 스택
│   ├── reports/                      # 동기화 리포트
│   │   └── .gitkeep
│   └── specs/                        # SPEC 문서 저장소
│       └── .gitkeep
│
└── CLAUDE.md                         # 프로젝트 지침 (루트)
```

#### 주요 디렉토리 설명

**`.claude/`** - Claude Code 통합

- `agents/`: 9개의 전문 에이전트 (spec-builder, code-builder, doc-syncer 등)
- `commands/`: 5개 Alfred 커맨드 (0-project, 1-spec, 2-build, 3-sync, 9-update)
- `hooks/`: 4개 자동화 훅 (정책 차단, 쓰기 검증, TAG 강제 등)
- `output-styles/`: 4개 출력 스타일 (프로, 초보자, 협업, 심층 학습)

**`.moai/`** - MoAI-ADK 코어

- `config.json`: 프로젝트 설정 (모드, 언어, Git 등)
- `memory/`: 개발 가이드 및 SPEC 메타데이터 표준
- `project/`: 제품 정보, 구조, 기술 스택 문서
- `specs/`: SPEC 문서가 저장되는 위치 (SPEC-XXX-001/)
- `reports/`: 동기화 리포트 및 TAG 검증 결과

### 3. 시스템 진단

프로젝트가 올바르게 설정되었는지 확인합니다:

```bash
moai doctor
```

실제 출력:

```
🔍 Checking system requirements...

  ⚙️  Runtime:
    ✅ Git (2.50.1)
    ✅ Node.js (20.19.4)

  🛠️  Development:
    ✅ npm (10.8.2)

  📦 Optional:
    ✅ Git LFS (3.7.0)

────────────────────────────────────────────────────────────
  📊 Summary:
     Checks: 4 total
     Status: 4 passed
────────────────────────────────────────────────────────────

✅ All requirements satisfied!
```

### 4. Post-Installation Setup

초기화가 완료되면 다음 단계를 진행합니다:

#### Step 1: 프로젝트 디렉토리로 이동

```bash
cd my-moai-project
```

#### Step 2: Claude Code 시작

VSCode에서 Claude Code를 실행합니다:

1. **방법 1: 터미널에서**

   ```bash
   claude
   ```

2. **방법 2: VSCode 커맨드 팔레트**
   - `Cmd+Shift+P` (Mac) 또는 `Ctrl+Shift+P` (Windows/Linux)
   - "Claude Code" 입력
   - "Open Claude Code" 선택

3. **방법 3: VSCode 하단 상태바**
   - 하단 상태바의 "Claude" 아이콘 클릭

#### Step 3: 프로젝트 초기화

Claude Code에서 프로젝트 정보를 설정합니다:

```bash
# Claude Code 채팅에서
/alfred:0-project
```

이 커맨드는 다음 파일들을 생성/업데이트합니다:

- **product.md**: 제품 개요, 목표, 사용자 정의
- **structure.md**: 디렉토리 구조, 아키텍처 설계
- **tech.md**: 기술 스택, 도구, 라이브러리 선택

::: tip 프로젝트 정보 작성
Alfred가 대화형으로 프로젝트 정보를 질문합니다:

1. 제품명과 설명
2. 주요 기능 목록
3. 기술 스택 선택
4. 프로젝트 구조 확인
:::

## 3-Stage Development Workflow

MoAI-ADK의 핵심 개발 사이클을 체험해보세요:

### Stage 1: SPEC 작성

```bash
# Claude Code에서 실행
/alfred:1-spec "사용자 인증 기능"
```

**결과**:

- `.moai/specs/SPEC-AUTH-001/spec.md` 생성
- `feature/SPEC-AUTH-001` 브랜치 생성
- Draft PR 생성

### Stage 2: TDD 구현

```bash
/alfred:2-build SPEC-AUTH-001
```

**결과**:

- `@TEST:AUTH-001` 테스트 작성 (RED)
- `@CODE:AUTH-001` 구현 (GREEN)
- 리팩토링 (REFACTOR)

### Stage 3: 문서 동기화

```bash
/alfred:3-sync
```

**결과**:

- Living Document 자동 생성
- TAG 체인 검증 (`@SPEC → @TEST → @CODE → @DOC`)
- PR Ready 전환

---

## What's Next?

축하합니다! 이제 MoAI-ADK를 사용할 준비가 되었습니다. 🎉

### 추천 학습 경로

1. **[Installation Guide](/guides/installation)** - 상세 설치 가이드
2. **[Quick Start Tutorial](/guides/quick-start)** - 실습 튜토리얼
3. **[SPEC-First TDD](/guides/concepts/spec-first-tdd)** - 핵심 개념 이해
4. **[EARS Requirements](/guides/concepts/ears-guide)** - 요구사항 작성법
5. **[TAG System](/guides/concepts/tag-system)** - 추적성 시스템
6. **[TRUST Principles](/guides/concepts/trust-principles)** - 품질 원칙

### 유용한 링크

- [API Reference](/api/index.html)
- [GitHub Repository](https://github.com/modu-ai/moai-adk)
- [Issue Tracker](https://github.com/modu-ai/moai-adk/issues)
- [Changelog](https://github.com/modu-ai/moai-adk/releases)

---

## Need Help?

문제가 발생하면 다음 리소스를 활용하세요:

### Troubleshooting

```bash
# 시스템 진단
moai doctor

# 프로젝트 상태 확인
moai status

# 상세 로그 보기
moai doctor --verbose
```

### Community

- **[GitHub Issues](https://github.com/modu-ai/moai-adk/issues)**: 버그 리포트 및 기능 요청
- **[Discussions](https://github.com/modu-ai/moai-adk/discussions)**: 질문 및 아이디어 공유

---

<div style="text-align: center; margin-top: 40px;">
  <p><strong>Happy Coding with MoAI-ADK!</strong> 🗿</p>
</div>
