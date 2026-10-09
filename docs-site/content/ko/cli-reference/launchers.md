---
title: moai cc / glm 런처
weight: 15
draft: false
---

`moai cc`와 `moai glm`은 선택한 백엔드로 Claude Code를 실행합니다. 기존 CG 설정은 실행 전에 이전해야 합니다.

## 런처 비교

| 런처 | 백엔드 | 용도 |
|------|--------|------|
| `moai cc` | Claude 전용 | 표준 실행 — 모든 에이전트가 Claude 모델 사용 |
| `moai glm` | GLM 전용 | 모든 에이전트가 Z.AI 프록시 경유 GLM 모델 사용 |

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
| `--chrome` / `--no-chrome` | Claude Code 에 그대로 전달합니다. 런처가 스스로 붙이지 않으므로 `--no-chrome` 을 넘기지 않는 한 `/chrome` 으로 연결할 수 있습니다 |
| `-f, --factory` | **팩토리 리더**로 진입합니다. 인자를 받지 않습니다. 리더는 운영자가 고른 카드를 교차 세션 메시지로 빈 레인에 통째로 배분하고, 레인은 `-l`로 합류시킵니다 |
| `-l, --lane` | 실행 중인 팩토리에 **레인**으로 합류해 다음 `lane-<n>` 번호(살아 있는 레인 가운데 가장 큰 번호의 다음)를 자동으로 받습니다. 인자를 받지 않으며, 실행 중인 팩토리가 없으면 거부됩니다. `moai glm -l`과 `moai codex -l`도 같은 동작입니다 |
| `--leader <name>` | `-l` 또는 `--lane`과 함께만 씁니다. 합류할 리더 세션을 지정합니다(기본값 `leader`, 옛 철자 `lead`는 거부). 실행 기록이 없거나 은퇴했는데 살아 있는 리더가 있으면, 합류는 그 리더를 검증(pid + 프로세스 시작)하고 그 실행을 복원합니다 |
| `--factory-run <run-id>` | `-l`과 함께: 합류할 실행을 id로 지정합니다. `--leader`와는 함께 쓸 수 없습니다 |
| `--clear-policy <value>` | `moai cc -l`·`moai glm -l`과 함께: 카드를 끝낸 뒤 컨텍스트를 비우는 방식입니다(`clear-each` 기본값, `clear-when-full`, `relaunch`) |
| `--no-auto-dispatch` | `moai cc -l`·`moai glm -l`과 함께: 레인을 수동 모드로 띄웁니다. 기본은 큐의 다음 카드를 스스로 임대하는 자가 배차 레인입니다 |

{{< callout type="info" >}} 진입 토큰은 `-f`(리더)와 `-l`(레인) 둘뿐이며 어느 쪽도 인자를 받지 않습니다. 한 번의 실행에는 토큰이 하나만 붙으므로 `-f`와 `-l`을 함께 쓰면 에러이고, `-f <값>`, `-l lane-2`처럼 값이 붙은 형태와 Codex에서 리더를 요구하는 `moai codex`의 `-f`는 모두 한 줄 오류로 거부됩니다. 폐기된 `-k` 진입도 같은 방식으로 `-f`와 `-l`을 안내하며 거부됩니다. 자세한 계약은 [팩토리 모드](/ko/advanced/factory-mode)와 [manager-lead 리더 코디네이터](/ko/advanced/manager-lead)를 참고하세요. {{< /callout >}}

카드는 통째로 레인 하나에 들어가, 그 레인 안에서 `plan → run → sync` 세 단계를 순서대로 지나갑니다. 단계마다 그 세션이 `Agent()` 서브에이전트를 띄우며, 쓰기 작업을 맡는 스폰은 `isolation: "worktree"`로 격리합니다. 레인 하나가 동시에 띄울 수 있는 서브에이전트는 최대 10개이고, 런처가 레인 세션에 `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`로 이 값을 심어 두므로 레인 N개가 머신 용량을 나눠 쓰는 구조는 운영자의 자제가 아니라 설정으로 보장됩니다. 레인은 한꺼번에 켜지 마세요. 첫 레인을 먼저 올리고, 실제로 출력이 나오기 시작한 것을 확인한 뒤에 나머지를 띄웁니다.

리더와 레인은 서로 다른 백엔드로 띄울 수 있습니다. 백엔드 조합은 토큰 여력을 먼저 보고 정하며, 판단이 무거운 자리에만 Opus를 두고 구현 중심 레인은 GLM으로 돌리는 방식이 한 가지 출발점입니다. 다른 조합을 쓰거나 한쪽 백엔드로 통일하는 것도 똑같이 괜찮습니다.

권한 모드는 `default`, `acceptEdits`(`moai init` 기본값), `plan`, `auto`, `bypassPermissions`, `dontAsk` 중 하나입니다. `auto` 모드에서는 백그라운드 분류기가 동작을 검사합니다. 지원하는 플랜과 모델은 [Claude Code 권한 모드 문서](https://code.claude.com/docs/en/permission-modes)에서 확인할 수 있습니다.

## moai glm — GLM 백엔드

```bash
moai glm setup <api-key>   # API 키 저장 (최초 1회)
moai glm --key <api-key>   # API 키 저장 (플래그 형태 — setup과 같은 저장소)
moai glm                   # GLM 백엔드로 실행
moai glm -p work           # 'work' 프로필로 실행
moai glm status            # 자격증명 상태 확인
```

`~/.moai/.env.glm` 에서 GLM 자격증명을 읽어 `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL` 등 환경 변수를 주입한 뒤 Claude Code를 실행합니다.

| 하위 명령어 | 설명 |
|-------------|------|
| `moai glm setup [api-key]` | GLM API 키 저장 |
| `moai glm --key <api-key>` | GLM API 키 저장 (플래그 형태 — 두 형태가 같은 파일에 기록하며 나중에 저장한 값이 남습니다) |
| `moai glm status` | 현재 GLM 자격증명 상태 표시 |

Jev(TypeSafe) 자격증명은 `moai jev --key <credential>` 로 저장합니다 — `~/.moai/.env.typesafe`(권한 0600)에 기록하며 `moai doctor` 와 웹 콘솔이 같은 파일을 읽습니다. `--key` 없이 `moai jev` 를 실행하면 도움말만 인쇄합니다.

{{< callout type="warning" >}}
GLM은 `auto` 권한 모드를 지원하지 않습니다. 이 모드는 사용 조건을 충족하는 Claude 세션에서 선택하세요. 폐기된 CG는 동시 실행의 대안이 아닙니다.
{{< /callout >}}

## CG 폐기와 설정 이전

Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다. [CG 폐기와 설정 이전](/ko/multi-llm/cg-mode/) CG는 폐기되었습니다. `moai migrate cg`로 이전 선택지를 먼저 확인하세요.

```bash
moai migrate cg
moai migrate cg --target claude-only --apply --accept-role-change
```

`llm.team_mode: claude`, `llm.gateway.teammate_mode: in-process`, `llm.gateway.teammate_provider: inherit`을 저장합니다. 기존 혼합 역할 배정을 없애는 변경이며, Claude 리더(CG 리더)와 GLM 팀원 창의 분업을 보존하는 이전이 아닙니다.

`claude-glm` 대상은 Claude 리더와 tmux의 GLM 팀원 구성을 뜻합니다. 현재 TEAMMATE 통합 검증을 통과하지 않아 적용과 실행은 사용할 수 없으며 미리보기만 가능합니다. tmux를 설치하거나 `verified: true`를 적어도 이 제한은 해제되지 않습니다.

## 프로필 (`-p` 플래그)

두 런처 모두 `-p <name>` 으로 명명된 프로필을 지정하면 `CLAUDE_CONFIG_DIR` 이 `~/.moai/claude-profiles/<name>/` 로 설정됩니다. 여러 계정·설정 세트를 분리해 운용할 때 사용합니다.

## 격리 worktree (`-w` 플래그)

두 런처 모두 `-w [name]` 으로 격리된 git worktree 안에서 세션을 시작할 수 있습니다. `cd` 로 디렉터리를 옮기고 다시 실행하던 두 단계가 한 명령으로 줄어듭니다.

```bash
moai cc -w feat-login    # .claude/worktrees/feat-login/ 에서 시작
moai cc -w               # 이름 자동 생성
moai glm -w feat-login   # GLM 백엔드도 동일
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

- [CG 폐기와 설정 이전](/ko/multi-llm/cg-mode/)
- [프로필 관리](/ko/cli-reference/profile)
- [보안 노트](/ko/advanced/security-notes) — GLM 자격증명 경로 보안 모델
- [CLI 개요](/ko/getting-started/cli)
