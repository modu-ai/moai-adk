---
title: moai worktree 工作树
weight: 25
draft: false
---

`moai worktree`(别名 `moai wt`)管理用于并行 SPEC 开发的 Git 工作树。它提供创建、列出、切换、同步、移除、清理工作树的子命令。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai worktree new [branch]` | 创建新工作树 |
| `moai worktree list` | 列出活动工作树 |
| `moai worktree status` | 显示工作树状态 |
| `moai worktree switch [branch]` | 切换到工作树 |
| `moai worktree go [branch]` | 输出用于 shell 跳转的工作树路径 |
| `moai worktree sync [branch]` | 将工作树与基础分支同步 |
| `moai worktree done [branch]` | 完成并清理工作树 |
| `moai worktree remove [path]` | 移除工作树 |
| `moai worktree clean` | 清理 stale 工作树引用 |
| `moai worktree recover` | 恢复工作树注册表 |

## moai worktree new

```bash
moai worktree new [branch-name]
```

| 标志 | 说明 |
|--------|------|
| `--path <dir>` | 指定工作树路径(默认:SPEC ID 用 `.moai/worktrees/<SPEC-ID>`,其他用 `../<branch-name>`) |
| `--base <branch>` | 基础分支(默认:`origin/main`,自动 fetch)。`--base main` 用于纯本地提交 |
| `--from-current` | 以当前 HEAD 作为工作树基础(跳过 `git fetch origin main`) |
| `--tmux` | 创建工作树后创建 tmux 会话 |
| `--team` | 在新工作树中 spawn Claude/GLM 会话(tmux+CG → `moai glm` 窗口,tmux+CC → `moai cc` 窗口,no-tmux → 进程内,no-flag → 交接引导) |

## moai worktree done

```bash
moai worktree done [branch-name]
```

| 标志 | 说明 |
|--------|------|
| `--force` | 即使有未提交更改也强制移除 |
| `--delete-branch` | 移除工作树后删除分支 |
| `--auto` | 自动模式 —— 用于自动化的无输出执行(例如 PR 合并后) |

## 示例

```bash
# 为 SPEC 创建工作树(基于 origin/main)
moai worktree new SPEC-AUTH-001

# 基于当前 HEAD 的本地工作树
moai worktree new feature-x --from-current

# 列出活动工作树
moai worktree list

# 在 shell 中跳转到工作树
cd "$(moai worktree go SPEC-AUTH-001)"

# 完成后清理 + 删除分支
moai worktree done SPEC-AUTH-001 --delete-branch
```

## 相关文档

- [工作树工作流](/zh/advanced/autonomous-loops) —— 并行开发模式
- [CG 模式](/zh/multi-llm/cg-mode) —— `--team` 混合执行
- [CLI 概览](/zh/getting-started/cli)
