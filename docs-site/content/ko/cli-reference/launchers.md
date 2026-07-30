---
title: moai cc / cg / glm 런처
weight: 15
draft: false
---

`moai cc`, `moai cg`, `moai glm` 은 Claude Code를 서로 다른 백엔드 구성으로 실행하는 세 가지 런처입니다. 세 명령 모두 설정을 조정한 뒤 `exec` 로 현재 프로세스를 Claude Code로 대체합니다. 어떤 모델이 어떤 일을 맡느냐가 곧 비용을 결정하므로, 런처 선택이 비용을 줄이는 첫 단추입니다.

## 세 런처 비교

| 런처 | 백엔드 | 용도 |
|------|--------|------|
| `moai cc` | Claude 전용 | 표준 실행 — 모든 에이전트가 Claude 모델 사용 |
| `moai glm` | GLM 전용 | 모든 에이전트가 Z.AI 프록시 경유 GLM 모델 사용 |
| `moai cg` | Claude + GLM 하이브리드 | 리더는 Claude, 팀원은 GLM (60-70% 비용 절감) |

## moai cc — Claude 백엔드

```bash
moai cc [-p profile] [-w [name]] [-- claude-args...]
```

`.claude/settings.local.json` 에서 GLM 전용 환경 변수를 제거하고, team 모드가 켜져 있었다면 초기화한 뒤 Claude Code를 실행합니다.

| 플래그 | 설명 |
|--------|------|
| `-p, --profile <name>` | 명명된 Claude 프로필 사용 (`~/.moai/claude-profiles/<name>/`) |
| `--permission-mode <mode>` | 권한 모드 지정 |
| `-b, --bypass` | `--permission-mode bypassPermissions` 단축형 |
| `-c, --continue` | 이전 세션 이어서 시작 |
| `-m, --model <model>` | 모델 선택 재정의 |
| `-w, --worktree [name]` | 격리된 git worktree(`.claude/worktrees/<name>/`)에서 실행 — 이름 생략 시 자동 생성 |
| `--chrome` / `--no-chrome` | Chrome MCP 토글 |

권한 모드는 `default`, `acceptEdits`(프로젝트 기본), `plan`, `auto`, `bypassPermissions`, `dontAsk` 중 하나입니다. `auto` 모드에서는 백그라운드 분류기가 동작을 검사하며, Team 플랜과 Sonnet/Opus 4.6 이상이 필요합니다.

## moai glm — GLM 백엔드

```bash
moai glm setup <api-key>   # API 키 저장 (최초 1회)
moai glm                   # GLM 백엔드로 실행
moai glm -p work           # 'work' 프로필로 실행
moai glm status            # 자격증명 상태 확인
```

`~/.moai/.env.glm` 에서 GLM 자격증명을 읽어 `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL` 등 환경 변수를 주입한 뒤 Claude Code를 실행합니다.

| 하위 명령어 | 설명 |
|-------------|------|
| `moai glm setup [api-key]` | GLM API 키 저장 |
| `moai glm status` | 현재 GLM 자격증명 상태 표시 |

{{< callout type="warning" >}}
GLM은 `auto` 권한 모드를 지원하지 않습니다 (서드파티 제공자). `auto` 가 필요하면 `moai cc` 또는 `moai cg` 를 사용하세요. 또한 Z.AI는 동시 요청 한도가 낮으므로(유료 티어 1-3 in-flight), 다중 에이전트 병렬 실행은 `moai cg` 하이브리드 모드가 더 안정적입니다.
{{< /callout >}}

## moai cg — Claude + GLM 하이브리드

```bash
moai cg [-p profile]
```

CG는 "Claude + GLM"의 약자로, 비용 최적화 팀 구성입니다.

- **리더** (현재 tmux pane): Claude 모델 사용 (opus/sonnet)
- **팀원** (새 tmux pane): Z.AI 프록시 경유 GLM 모델 사용

실행하면 먼저 tmux 세션을 검증합니다. 이어서 리더 pane에서는 GLM 환경을 제거해 Claude로 두고, tmux 세션 쪽에는 GLM 환경을 주입해 팀원이 GLM을 쓰게 한 뒤, `teammateMode=tmux` 와 `team_mode: cg` 를 설정합니다.

**전제 조건**:

1. `moai glm setup <api-key>` 로 GLM API 키 설정
2. pane 단위 환경 격리를 위한 tmux 세션 내부 실행

## 프로필 (`-p` 플래그)

세 런처 모두 `-p <name>` 으로 명명된 프로필을 지정하면 `CLAUDE_CONFIG_DIR` 이 `~/.moai/claude-profiles/<name>/` 로 설정됩니다. 여러 계정·설정 세트를 분리해 운용할 때 사용합니다.

## 격리 worktree (`-w` 플래그)

세 런처 모두 `-w [name]` 으로 격리된 git worktree 안에서 세션을 시작할 수 있습니다. `cd` 로 디렉터리를 옮기고 다시 실행하던 두 단계가 한 명령으로 줄어듭니다.

```bash
moai cc -w feat-login    # .claude/worktrees/feat-login/ 에서 시작
moai cc -w               # 이름 자동 생성
moai glm -w feat-login   # GLM 백엔드도 동일
moai cg -w feat-login    # 하이브리드도 동일
```

동작 규칙:

- worktree 경로는 `.claude/worktrees/<name>/` 입니다. `<name>` 은 **worktree 이름**이며 브랜치명이나 SPEC ID가 아닙니다.
- 같은 이름의 worktree가 이미 있으면 **새로 만들지 않고 재사용**합니다. 그래서 이전 세션이 작업하던 트리로 다시 들어가는 재진입 경로로도 쓸 수 있습니다.
- 이름을 생략하면 Claude Code가 자동으로 짓습니다.
- `-w=name`, `--worktree name`, `--worktree=name` 표기도 모두 같은 의미로 받습니다.
- `--` 뒤의 인자는 그대로 Claude Code에 전달되며 이 재작성의 영향을 받지 않습니다.

{{< callout type="info" >}}
세션 인수인계에서 worktree 이름을 SPEC ID와 같게 지어 두면(`moai cc -w SPEC-XXX-001`) 다음 세션이 한 줄로 같은 작업 트리에 복귀할 수 있습니다.
{{< /callout >}}

## 관련 문서

- [CG 모드 (Claude + GLM)](/ko/multi-llm/cg-mode)
- [프로필 관리](/ko/cli-reference/profile)
- [보안 노트](/ko/advanced/security-notes) — GLM 자격증명 경로 보안 모델
- [CLI 개요](/ko/getting-started/cli)
