---
title: moai worktree 工作树
weight: 25
draft: false
---

`moai worktree`(别名 `moai wt`)管理用于并行 SPEC 开发的 Git 工作树。它提供八个子命令:同步、收尾、移除、清理、注册表恢复,以及包裹隔离智能体运行的状态守卫。

## 进入和列出工作树不归这条命令管

`moai worktree` 只**管理**工作树,既不带你进去,也不替你列出它们。

| 你想做的事 | 该用的命令 |
|-----------|-------------|
| 在工作树里开始工作 | `moai cc -w <name>` (或 `moai glm -w` / `moai cg -w`) |
| 保留当前会话,在新 tmux 窗口中打开 | `moai cc -w <name> --spawn` |
| 查看工作树列表 | `git worktree list` |
| 新建工作树 | `moai cc -w <name>` (自动创建 `.claude/worktrees/<name>/`) 或 `git worktree add` |

传给 `-w` 的短名称会在 `.claude/worktrees/<name>/` 下解析,不存在就地创建。给出绝对路径时,会重新进入 `~/.moai/worktrees/` 或 `<项目>/.claude/worktrees/` 下的既有工作树。这两个前缀之外的绝对路径一律被拒绝。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai worktree sync [branch-name]` | 把基础分支的变更带进工作树 |
| `moai worktree done <branch-name>` | 移除分支所属的工作树,可选连分支一起删除 |
| `moai worktree remove <path>` | 移除指定路径上的工作树 |
| `moai worktree clean` | 清理 stale 引用,收拾已合并或已废弃的工作树 |
| `moai worktree recover` | 修复工作树注册表 |
| `moai worktree snapshot` | 把工作树状态捕获为快照 |
| `moai worktree verify` | 将当前工作树与快照对照 |
| `moai worktree restore` | 把工作树回退到快照 HEAD 状态 |

## moai worktree sync

```bash
moai worktree sync [branch-name]
```

给出分支名就同步该分支的工作树,省略则同步当前目录所在的工作树。

| 标志 | 说明 |
|--------|------|
| `--base <branch>` | 基准分支 (默认: `main`) |
| `--strategy <mode>` | `merge` (默认) 或 `rebase` |

## moai worktree done

```bash
moai worktree done <branch-name>
```

分支名是必填的。它会找到使用该分支的工作树并移除,需要的话连分支一起删掉。**它不做合并** —— 到基础分支的合并请用 `git merge` 或 PR 另行完成。

| 标志 | 说明 |
|--------|------|
| `--force` | 即使有未提交变更也强制移除 |
| `--delete-branch` | 移除工作树后删除分支 |
| `--auto` | 供自动化使用的静默模式 (例如 PR 合并后的清理)。找不到工作树也不会以错误结束 |

## moai worktree remove

```bash
moai worktree remove <path>
```

参数是**文件系统路径**,不是分支名。

| 标志 | 说明 |
|--------|------|
| `--force` | 即使有未提交变更也强制移除 |

## moai worktree clean

```bash
moai worktree clean [--merged-only | --stale] [--yes] [--json] [--base <branch>]
```

不带标志运行时,只 prune stale 的工作树引用。

| 标志 | 说明 |
|--------|------|
| `--merged-only` | 只移除分支已合并进基础分支的工作树 |
| `--stale` | 清扫没有任何东西可失去的废弃工作树 (默认只预览) |
| `--yes` | 不再预览,真正执行 `--stale` 的移除 |
| `--json` | 与 `--stale` 搭配：以 JSON 输出所有非保护工作树及其保留理由、dirty、合并与锚定状态。不会移除任何东西，并且优先于 `--yes` |
| `--base <branch>` | 判定 `--merged-only` 与 `--stale` 所依据的基础分支 (默认: `origin/main`) |

`--stale` 与 `--merged-only` 不能一起使用。

### --stale 的安全规则

工作树只有**同时**满足下面两个条件才会成为移除候选。

1. 工作树是干净的 —— 既没有未提交变更,也没有 untracked 文件
2. 分支上没有超出基础分支的独有提交

只要有一条不满足,该工作树就会被保留,并一并打印保留的原因。**分支绝不会被删除**,所以即使工作树目录消失了,提交仍然可以通过分支名找回。主检出以及正在运行该命令的工作树始终受保护。

`--stale` 默认只预览。真要删除请加上 `--yes`。

## moai worktree recover

```bash
moai worktree recover
```

用 `git worktree repair` 修复工作树管理文件,随后 prune stale 引用,最后打印识别到的工作树列表。它没有标志。

## moai worktree snapshot

```bash
moai worktree snapshot
```

捕获 HEAD、分支、porcelain 状态以及 `.moai/specs/` 下的 untracked 文件,并以 JSON 记录到 `.moai/state/`。它的用途是在调用隔离智能体之前先留下一份读数。

| 标志 | 说明 |
|--------|------|
| `--out <path>` | 快照输出路径 (默认: `.moai/state/worktree-snapshot-<id>.json`) |
| `--agent-name <name>` | 记录智能体名称 (之后在 verify 阶段引用) |

## moai worktree verify

```bash
moai worktree verify --snapshot <path>
```

将当前工作树与快照对照。`--snapshot` 是**必填**的。

| 标志 | 说明 |
|--------|------|
| `--snapshot <path>` | 事前快照 JSON 的路径 (必填) |
| `--agent-response <path>` | 智能体响应 JSON —— 用于检测空的 `worktreePath` |
| `--agent-name <name>` | 记入 divergence 与 suspect 日志的智能体名称 |

| 退出码 | 含义 |
|-----------|------|
| `0` | clean |
| `1` | 检测到 divergence |
| `2` | suspect (空的 `worktreePath`) |
| `3` | 两者皆有 |

## moai worktree restore

```bash
moai worktree restore --snapshot <path>
```

执行 `git restore --source=<快照 HEAD> --staged --worktree :/`,把被跟踪的文件回退到快照 HEAD 状态。**untracked 文件无法由 git 找回**,因此只会列出它们的路径,需要你自己重新创建。

| 标志 | 说明 |
|--------|------|
| `--snapshot <path>` | 快照 JSON 的路径 (必填) |
| `--dry-run` | 只打印将要执行的 git 命令,不实际执行 |

## 示例

```bash
# 创建工作树并立刻进入 (.claude/worktrees/feat-auth/)
moai cc -w feat-auth

# 保留当前会话,在新 tmux 窗口中启动 GLM 队友
moai cg -w feat-auth --spawn

# 查看工作树列表
git worktree list

# 把当前工作树与 main 同步 (merge)
moai worktree sync

# 用 rebase 同步指定的工作树
moai worktree sync feature/SPEC-AUTH-001 --strategy rebase

# 先预览废弃的工作树,确认后再真正移除
moai worktree clean --stale
moai worktree clean --stale --yes

# 合并完成后清理工作树并删除分支
moai worktree done feature/SPEC-AUTH-001 --delete-branch
```

## 相关文档

- [Git Worktree 概述](/zh/worktree/) —— 概念与工作流
- [完整指南](/zh/worktree/guide) —— 每条命令的详细用法
- [CG 模式](/zh/multi-llm/cg-mode) —— Claude 领导 + GLM 队友混合
- [CLI 概览](/zh/getting-started/cli)
