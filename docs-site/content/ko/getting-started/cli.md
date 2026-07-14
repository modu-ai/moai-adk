---
title: CLI 레퍼런스
weight: 90
draft: false
---

터미널에서 실행하는 `moai` (Go 바이너리) 의 모든 명령어와 플래그를 참조합니다. Claude Code 대화창에서 입력하는 `/moai` (슬래시 서브커맨드) 와는 완전히 다른 도구입니다 — 이 페이지는 터미널 CLI만 다룹니다.

## 명령어 트리

```bash
moai --help
```

`moai` CLI는 세 그룹으로 나뉩니다.

| 그룹 | 명령어 | 설명 |
|------|--------|------|
| **Launch** | `moai cc` · `moai cg` · `moai glm` | Claude Code 세션 시작 (백엔드 선택) |
| **Project** | `moai init` · `moai update` · `moai doctor` · `moai status` | 프로젝트 초기화, 업데이트, 진단, 상태 조회 |
| **Tools** | `moai profile` · `moai inventory` · `moai hook` · `moai worktree` · `moai spec` · `moai harness` · ... | 설정, 인벤토리, 훅, 워크트리 등 도구 |

`moai version` 으로 현재 설치된 버전을 확인합니다.

```bash
moai version
# 출력 예시: moai <버전> (commit: <해시>, built: <빌드 날짜>)
```

---

## moai init

프로젝트를 초기화합니다. 대화형 마법사가 언어, Git 자동화, 모델 정책, 하네스 프로필 등을 설정합니다.

```bash
moai init [project-name] [OPTIONS]
```

### 플래그

| 플래그 | 설명 |
|--------|------|
| `--non-interactive` | 대화형 마법사 건너뛰기 (플래그와 기본값 사용) |
| `--force` | 기존 프로젝트 강제 재초기화 (현재 `.moai/` 백업) |
| `--no-hooks` | Git 훅 설치 건너뛰기 |
| `--all` | 카탈로그 전체 항목 배포 (core + optional packs + harness-generated) |
| `--standard` | Phase 1 질문 표시 (project mode, harness profile, LSP, quality gates, design) |
| `--advanced` | Phase 1 + Phase 2 질문 표시 (`--standard` 포함; Phase 2는 선행 조건 충족 시만) |
| `--project-mode <personal\|team>` | 프로젝트 모드 (기본값: personal) |
| `--harness-profile <profile>` | 하네스 평가자 프로필 (default, strict, lenient, frontend) |
| `--enable-lsp` | LSP 연동 활성화 (기본값: false) |
| `--enforce-quality` | 품질 게이트 강제 (기본값: true) |
| `--enable-design` | 디자인 워크플로우 활성화 (기본값: true) |
| `--model-policy <max\|medium\|low>` | 성능 티어 — `llm.yaml` `performance_tier` 에 저장 |
| `--plan-type <api\|subscription>` | 요금제 유형 — `llm.yaml` `plan_type` 에 저장 |
| `--high` | **삭제 예정** `--model-policy max` 의 별칭 |

### 예시

```bash
# 새 프로젝트 초기화 (대화형 마법사)
moai init my-project

# 기존 폴더에 설치
cd my-existing-project
moai init

# 비대화형 (CI/CD)
moai init --non-interactive --project-mode personal --model-policy medium

# Phase 1 질문까지 표시
moai init my-project --standard
```

자세한 마법사 단계는 [초기 설정](./init-wizard) 페이지를 참조하세요.

---

## moai update

MoAI-ADK를 최신 버전으로 업데이트합니다. 플래그 없이 실행하면 바이너리와 템플릿을 함께 갱신하며, 사용자 커스텀 자산은 자동 보존됩니다.

```bash
moai update [OPTIONS]
```

### 플래그

| 플래그 | 설명 |
|--------|------|
| `--check` | 새 버전이 있는지만 확인 (업데이트 안 함) |
| `-c, --config` | 설정 마법사 다시 실행 (템플릿 동기화 안 함) |
| `--force` | 강제 업데이트 (버전 일치 스킵, 백업+병합 강제, 아카이브 드리프트 덮어쓰기) |
| `--yes` | 모든 확인 자동 승인 (CI/CD 모드) |
| `--templates-only` | 바이너리 업데이트 건너뛰고 템플릿만 동기화 |
| `--binary` | 템플릿 동기화 건너뛰고 바이너리만 업데이트 |
| `--dry-run` | 파일시스템 변경 없이 계획된 작업만 표시 |
| `--no-hooks` | Git 훅 설치 건너뛰기 |
| `--verbose` | 모든 경고 표시 (진단 모드) |
| `--shell-env` | Claude Code 용 셸 환경변수 구성 |
| `--plan-type <api\|subscription>` | 요금제 유형 덮어쓰기 (`llm.yaml` `plan_type` 및 티어 프로필 재적용) |

### 예시

```bash
# 기본 업데이트 (바이너리 + 템플릿)
moai update

# 새 버전이 있는지 확인만
moai update --check

# 설정 마법사 다시 실행
moai update -c

# 템플릿만 동기화
moai update --templates-only
```

자세한 업데이트 절차는 [업데이트](./update) 페이지를 참조하세요.

---

## moai doctor

시스템 진단을 실행합니다. Git, 프로젝트 구조, 설정 파일, 언어별 개발 도구를 검사합니다.

```bash
moai doctor [OPTIONS]
```

### 플래그

| 플래그 | 설명 |
|--------|------|
| `-v, --verbose` | 상세 도구 버전 및 언어 감지 결과 표시 |
| `--fix` | 누락 도구 수정 제안 |
| `--export <path>` | 진단 결과를 JSON 파일로 내보내기 |
| `--check <tool>` | 특정 도구만 확인 (예: git, go, config) |

### 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai doctor sandbox` | 샌드박스 환경 진단 |
| `moai doctor permission` | 권한 설정 진단 |
| `moai doctor hook` | 훅 로딩 문제 진단 |
| `moai doctor config dump` | 현재 설정을 JSON으로 덤프 |
| `moai doctor config diff` | 로컬 설정과 템플릿 기본값 비교 |

### 예시

```bash
# 전체 진단
moai doctor

# 상세 진단
moai doctor --verbose

# 진단 결과 내보내기
moai doctor --export diagnostics.json
```

---

## moai status

프로젝트 상태를 한눈에 조회합니다. 초기화 여부, SPEC 개수, 설정 파일 수를 표시합니다.

```bash
moai status
```

플래그가 없는 읽기 전용 명령어입니다. 자세한 출력 내용은 [프로젝트 상태](./status) 페이지를 참조하세요.

---

## moai inventory

활성 세션, 워크트리, 하네스를 통합 조회하는 읽기 전용 명령어입니다.

```bash
moai inventory [OPTIONS]
```

### 플래그

| 플래그 | 설명 |
|--------|------|
| `--json` | 구조화된 JSON 출력 |
| `--project-root <path>` | 프로젝트 루트 경로 (기본값: 현재 디렉토리) |

자세한 JSON 스키마와 활용 예시는 [moai inventory](./inventory) 페이지를 참조하세요.

---

## moai profile

Claude Code 설정 프로필을 관리합니다. 프로필별로 독립적인 모델, 언어, 표시 설정을 유지할 수 있습니다.

```bash
moai profile [COMMAND]
```

### 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai profile list` | 사용 가능한 모든 프로필 표시 |
| `moai profile setup` | 대화형 설정 마법사 실행 |
| `moai profile current` | 현재 활성 프로필 표시 |
| `moai profile delete <name>` | 지정된 프로필 삭제 |

프로필 실행은 `-p` 플래그로 지정합니다:

```bash
moai cc -p work       # work 프로필로 Claude 실행
moai glm -p cost-save # cost-save 프로필로 GLM 실행
moai cg -p team       # team 프로필로 CG 모드 실행
```

자세한 내용은 [프로필 관리](./profile) 페이지를 참조하세요.

---

## moai hook

Claude Code 훅 이벤트를 처리하는 디스패처입니다. `settings.json` 의 훅 설정에서 `moai hook <event>` 형태로 호출됩니다.

```bash
moai hook <event>
```

### 지원 이벤트 (26개)

모든 이벤트 이름은 kebab-case 입니다.

| 이벤트 | 설명 |
|-------|------|
| `session-start` | 세션 시작 |
| `session-end` | 세션 종료 |
| `pre-tool` | 도구 실행 전 (PreToolUse) |
| `post-tool` | 도구 실행 후 (PostToolUse) |
| `post-tool-failure` | 도구 실행 실패 후 |
| `stop` | 세션 정지 |
| `stop-failure` | 정지 실패 |
| `compact` | 컨텍스트 압축 전 (PreCompact) |
| `post-compact` | 컨텍스트 압축 후 |
| `notification` | 시스템 알림 |
| `subagent-start` | 서브에이전트 시작 |
| `subagent-stop` | 서브에이전트 종료 |
| `user-prompt-submit` | 사용자 프롬프트 제출 |
| `permission-request` | 권한 요청 |
| `permission-denied` | 권한 거부 |
| `teammate-idle` | 팀원 유휴 상태 |
| `task-completed` | 작업 완료 |
| `task-created` | 작업 생성 |
| `worktree-create` | 워크트리 생성 |
| `worktree-remove` | 워크트리 제거 |
| `instructions-loaded` | 인스트럭션 로드 완료 |
| `config-change` | 설정 변경 |
| `cwd-changed` | 작업 디렉토리 변경 |
| `file-changed` | 파일 변경 |
| `elicitation` | MCP elicitation 요청 |
| `elicitation-result` | MCP elicitation 결과 |

훅은 사용자가 직접 실행하지 않습니다 — Claude Code의 `settings.json` 이 자동으로 호출합니다.

---

## moai worktree

Git worktree를 관리하여 병렬 SPEC 개발을 수행합니다.

```bash
moai worktree <COMMAND> [ARGS]...
```

### 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai worktree new <SPEC_ID>` | 새 worktree 생성 |
| `moai worktree list` | 활성 worktree 목록 |
| `moai worktree go <SPEC_ID>` | worktree 디렉토리로 이동 |
| `moai worktree remove <SPEC_ID>` | worktree 제거 |
| `moai worktree clean` | 오래된 worktree 정리 |
| `moai worktree recover` | 기존 디렉토리에서 복구 |
| `moai worktree status` | worktree 상태 조회 |

---

## moai cc / moai cg / moai glm

Claude Code를 시작하면서 백엔드를 선택하는 런치 명령어입니다. 세 명령어 모두 `-p <profile>` 플래그로 프로필을 지정할 수 있고, `--` 이후의 인자를 Claude Code에 그대로 전달합니다.

```bash
moai cc [-p profile] [-- claude-args...]
moai cg [-p profile] [-- claude-args...]
moai glm [-p profile] [-- claude-args...]
```

| 명령어 | 리더 | 워커 | tmux 필수 | 용도 |
|--------|------|------|-----------|------|
| `moai cc` | Claude | Claude | 아니오 | 최고 품질 (단일 백엔드) |
| `moai glm` | GLM | GLM | 권장 | 비용 최적화 (GLM 단독) |
| `moai cg` | Claude | GLM | 필수 | 품질 + 비용 균형 (하이브리드) |

`moai cg` 는 CG 모드 (Claude 리더 + GLM 팀원) 를 활성화합니다. tmux 세션 내에서 실행해야 하며, GLM 환경변수를 tmux 세션에 주입하고 리더 창은 Claude API를 사용합니다.

```bash
# 1. GLM API 키 저장 (최초 1회)
moai glm sk-your-glm-api-key

# 2. CG 모드 활성화 (tmux 내에서 실행)
moai cg

# 3. 같은 창에서 Claude Code 시작
claude
```

자세한 CG 모드 안내는 [소개 — GLM으로 토큰 절약](./introduction#glm으로-토큰-절약-5070) 을 참조하세요.

---

## moai version

버전, 커밋 해시, 빌드 날짜를 표시합니다.

```bash
moai version
moai --version    # 동일
```

---

## 모델 정책 (성능 티어)

MoAI-ADK는 에이전트에 최적의 AI 모델을 할당하는 성능 티어 시스템을 제공합니다 — 토크노믹스의 출발점입니다. `llm.yaml` 의 `performance_tier` 필드로 설정하며, `--model-policy` 플래그 또는 초기화 마법사에서 선택합니다.

| 티어 | 특징 |
|------|------|
| **max** | 최고 품질 — 계획·감사에 Opus 배정, 최대 추론 깊이 |
| **medium** (기본값) | 품질과 비용의 균형 |
| **low** | 경제적 — Sonnet 중심 배분 |

```bash
# 초기화 시 설정
moai init my-project --model-policy max

# 기존 프로젝트에서 재설정
moai update -c
```

요금제 유형(`plan_type`: api 또는 subscription) 은 별도로 설정하여, 같은 티어라도 과금 방식에 따라 모델 배정이 달라집니다. 자세한 모델-티어 매핑은 [모델 정책](/ko/multi-llm/model-policy) 페이지를 참조하세요.

---

## 참고

- [빠른 시작](./quickstart)
- [설치](./installation)
- [업데이트](./update)
- [초기 설정](./init-wizard)
- [프로필 관리](./profile)
- [프로젝트 상태](./status)
