---
title: moai worktree 워크트리
weight: 25
draft: false
---

`moai worktree` (별칭 `moai wt`) 는 병렬 SPEC 개발에 쓰는 Git 워크트리를 관리합니다. 동기화, 완료 처리, 제거, 정리, 레지스트리 복구, 그리고 격리된 에이전트 실행을 감싸는 상태 가드까지 여덟 개의 하위 명령어를 제공합니다.

## 워크트리 진입과 조회는 이 명령어의 일이 아닙니다

`moai worktree` 는 워크트리를 **관리**할 뿐, 그 안으로 들어가거나 목록을 보여 주지 않습니다.

| 하려는 일 | 쓰는 명령어 |
|-----------|-------------|
| 워크트리 안에서 작업 시작 | `moai cc -w <name>` (또는 `moai glm -w` / `moai cg -w`) |
| 현재 세션은 두고 새 tmux 창에서 열기 | `moai cc -w <name> --spawn` |
| 워크트리 목록 확인 | `git worktree list` |
| 워크트리 새로 만들기 | `moai cc -w <name>` (`.claude/worktrees/<name>/` 자동 생성) 또는 `git worktree add` |

`-w` 에 짧은 이름을 주면 `.claude/worktrees/<name>/` 아래에서 해석되고, 없으면 새로 만들어집니다. 절대 경로를 주면 `~/.moai/worktrees/` 또는 `<프로젝트>/.claude/worktrees/` 아래의 기존 워크트리로 다시 들어갑니다. 그 밖의 절대 경로는 거부됩니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai worktree sync [branch-name]` | 베이스 브랜치의 변경을 워크트리로 반영 |
| `moai worktree done <branch-name>` | 브랜치에 딸린 워크트리 제거, 선택적으로 브랜치까지 삭제 |
| `moai worktree remove <path>` | 지정한 경로의 워크트리 제거 |
| `moai worktree clean` | stale 참조 정리, 병합된/버려진 워크트리 정리 |
| `moai worktree recover` | 워크트리 레지스트리 복구 |
| `moai worktree snapshot` | 작업 트리 상태를 스냅샷으로 캡처 |
| `moai worktree verify` | 현재 작업 트리를 스냅샷과 대조 |
| `moai worktree restore` | 스냅샷 HEAD 상태로 작업 트리 되돌리기 |

## moai worktree sync

```bash
moai worktree sync [branch-name]
```

브랜치 이름을 주면 그 브랜치의 워크트리를, 생략하면 현재 디렉터리의 워크트리를 동기화합니다.

| 플래그 | 설명 |
|--------|------|
| `--base <branch>` | 기준 브랜치 (기본: `main`) |
| `--strategy <mode>` | `merge` (기본) 또는 `rebase` |

## moai worktree done

```bash
moai worktree done <branch-name>
```

브랜치 이름은 필수입니다. 해당 브랜치를 쓰는 워크트리를 찾아 제거하고, 원하면 브랜치까지 삭제합니다. **병합은 하지 않습니다** — 베이스 브랜치 병합은 `git merge` 나 PR 로 따로 끝내세요.

| 플래그 | 설명 |
|--------|------|
| `--force` | 미커밋 변경이 있어도 강제 제거 |
| `--delete-branch` | 워크트리 제거 후 브랜치 삭제 |
| `--auto` | 자동화용 무출력 모드 (예: PR 머지 후 정리). 워크트리를 못 찾아도 오류로 끝내지 않습니다 |

## moai worktree remove

```bash
moai worktree remove <path>
```

인자는 브랜치 이름이 아니라 **파일 시스템 경로**입니다.

| 플래그 | 설명 |
|--------|------|
| `--force` | 미커밋 변경이 있어도 강제 제거 |

## moai worktree clean

```bash
moai worktree clean [--merged-only | --stale] [--yes] [--base <branch>]
```

플래그 없이 실행하면 stale 워크트리 참조만 prune 합니다.

| 플래그 | 설명 |
|--------|------|
| `--merged-only` | 브랜치가 베이스에 병합된 워크트리만 제거 |
| `--stale` | 잃을 것이 없는 방치된 워크트리를 쓸어 담기 (기본은 미리보기) |
| `--yes` | `--stale` 미리보기 대신 실제 제거 수행 |
| `--base <branch>` | `--merged-only` · `--stale` 판정 기준 브랜치 (기본: `main`) |

`--stale` 과 `--merged-only` 는 함께 쓸 수 없습니다.

### --stale 안전 규칙

워크트리는 다음 두 조건을 **모두** 만족할 때만 제거 대상이 됩니다.

1. 작업 트리가 깨끗하다 — 미커밋 변경도, untracked 파일도 없다
2. 브랜치에 베이스를 넘어서는 고유 커밋이 없다

하나라도 어긋나면 그 워크트리는 유지되고, 유지한 이유가 함께 출력됩니다. **브랜치는 절대 삭제하지 않으므로** 워크트리 디렉터리가 사라져도 커밋은 브랜치 이름으로 그대로 남습니다. 메인 체크아웃과 명령을 실행 중인 워크트리는 항상 보호 대상입니다.

`--stale` 은 기본이 미리보기입니다. 실제로 지우려면 `--yes` 를 붙이세요.

## moai worktree recover

```bash
moai worktree recover
```

`git worktree repair` 로 워크트리 관리 파일을 고친 뒤 stale 참조를 prune 하고, 최종적으로 인식된 워크트리 목록을 출력합니다. 플래그는 없습니다.

## moai worktree snapshot

```bash
moai worktree snapshot
```

HEAD, 브랜치, porcelain 상태, `.moai/specs/` 아래 untracked 파일을 캡처해 `.moai/state/` 에 JSON 으로 기록합니다. 격리된 에이전트를 호출하기 직전에 찍어 두는 용도입니다.

| 플래그 | 설명 |
|--------|------|
| `--out <path>` | 스냅샷 저장 경로 (기본: `.moai/state/worktree-snapshot-<id>.json`) |
| `--agent-name <name>` | 에이전트 이름 기록 (이후 verify 단계에서 참조) |

## moai worktree verify

```bash
moai worktree verify --snapshot <path>
```

현재 작업 트리를 스냅샷과 대조합니다. `--snapshot` 은 **필수**입니다.

| 플래그 | 설명 |
|--------|------|
| `--snapshot <path>` | 사전 스냅샷 JSON 경로 (필수) |
| `--agent-response <path>` | 에이전트 응답 JSON — 빈 `worktreePath` 탐지용 |
| `--agent-name <name>` | divergence · suspect 로그에 기록할 에이전트 이름 |

| 종료 코드 | 의미 |
|-----------|------|
| `0` | clean |
| `1` | divergence 감지 |
| `2` | suspect (빈 `worktreePath`) |
| `3` | 둘 다 |

## moai worktree restore

```bash
moai worktree restore --snapshot <path>
```

`git restore --source=<스냅샷 HEAD> --staged --worktree :/` 를 실행해 추적 중인 파일을 스냅샷 HEAD 상태로 되돌립니다. **untracked 파일은 git 으로 되살릴 수 없어** 경로만 나열되며, 직접 다시 만들어야 합니다.

| 플래그 | 설명 |
|--------|------|
| `--snapshot <path>` | 스냅샷 JSON 경로 (필수) |
| `--dry-run` | 실행 없이 수행할 git 명령만 출력 |

## 예시

```bash
# 워크트리 만들면서 바로 들어가기 (.claude/worktrees/feat-auth/)
moai cc -w feat-auth

# 현재 세션은 유지한 채 새 tmux 창에서 GLM 팀원 띄우기
moai cg -w feat-auth --spawn

# 워크트리 목록
git worktree list

# 현재 워크트리를 main 과 동기화 (merge)
moai worktree sync

# 특정 워크트리를 rebase 로 동기화
moai worktree sync feature/SPEC-AUTH-001 --strategy rebase

# 방치된 워크트리 먼저 미리보기, 확인 뒤 실제 제거
moai worktree clean --stale
moai worktree clean --stale --yes

# 병합 끝난 뒤 워크트리 정리 + 브랜치 삭제
moai worktree done feature/SPEC-AUTH-001 --delete-branch
```

## 관련 문서

- [Git Worktree 개요](/ko/worktree/) — 개념과 워크플로우
- [완벽 가이드](/ko/worktree/guide) — 명령어별 상세 사용법
- [CG 모드](/ko/multi-llm/cg-mode) — Claude 리더 + GLM 팀원 하이브리드
- [CLI 개요](/ko/getting-started/cli)
