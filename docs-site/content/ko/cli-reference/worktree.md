---
title: moai worktree 워크트리
weight: 25
draft: false
---

`moai worktree` (별칭 `moai wt`) 는 병렬 SPEC 개발을 위한 Git 워크트리를 관리합니다. 워크트리를 만들고, 나열하고, 전환하고, 동기화하고, 제거하고, 정리하는 하위 명령어를 제공합니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai worktree new [branch]` | 새 워크트리 생성 |
| `moai worktree list` | 활성 워크트리 나열 |
| `moai worktree status` | 워크트리 상태 표시 |
| `moai worktree switch [branch]` | 워크트리로 전환 |
| `moai worktree go [branch]` | 셸 이동용 워크트리 경로 출력 |
| `moai worktree sync [branch]` | 베이스 브랜치와 워크트리 동기화 |
| `moai worktree done [branch]` | 워크트리 완료 및 정리 |
| `moai worktree remove [path]` | 워크트리 제거 |
| `moai worktree clean` | stale 워크트리 참조 정리 |
| `moai worktree recover` | 워크트리 레지스트리 복구 |

## moai worktree new

```bash
moai worktree new [branch-name]
```

| 플래그 | 설명 |
|--------|------|
| `--path <dir>` | 워크트리 경로 지정 (기본: SPEC ID는 `.moai/worktrees/<SPEC-ID>`, 그 외 `../<branch-name>`) |
| `--base <branch>` | 베이스 브랜치 (기본: `origin/main`, 자동 fetch). `--base main` 은 로컬 전용 커밋용 |
| `--from-current` | 현재 HEAD를 워크트리 베이스로 사용 (`git fetch origin main` 생략) |
| `--tmux` | 워크트리 생성 후 tmux 세션 생성 |
| `--team` | 새 워크트리에서 Claude/GLM 세션 스폰 (tmux+CG → `moai glm` 창, tmux+CC → `moai cc` 창, no-tmux → 인프로세스, no-flag → 핸드오프 안내) |

## moai worktree done

```bash
moai worktree done [branch-name]
```

| 플래그 | 설명 |
|--------|------|
| `--force` | 미커밋 변경이 있어도 강제 제거 |
| `--delete-branch` | 워크트리 제거 후 브랜치 삭제 |
| `--auto` | 자동 모드 — 자동화용 무출력 실행 (예: PR 머지 후) |

## 예시

```bash
# SPEC용 워크트리 생성 (origin/main 베이스)
moai worktree new SPEC-AUTH-001

# 현재 HEAD 기준 로컬 워크트리
moai worktree new feature-x --from-current

# 활성 워크트리 나열
moai worktree list

# 셸에서 워크트리로 이동
cd "$(moai worktree go SPEC-AUTH-001)"

# 완료 후 정리 + 브랜치 삭제
moai worktree done SPEC-AUTH-001 --delete-branch
```

## 관련 문서

- [워크트리 워크플로우](/ko/advanced/autonomous-loops) — 병렬 개발 패턴
- [CG 모드](/ko/multi-llm/cg-mode) — `--team` 하이브리드 실행
- [CLI 개요](/ko/getting-started/cli)
