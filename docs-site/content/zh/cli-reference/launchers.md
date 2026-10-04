---
title: moai cc / glm 启动器
weight: 15
draft: false
---

`moai cc` 和 `moai glm` 使用明确选择的后端启动 Claude Code。旧 CG 配置必须先迁移才能启动。

## 受支持的启动器对比

| 启动器 | 后端 | 用途 |
|------|--------|------|
| `moai cc` | 仅 Claude | 标准执行 —— 所有 agent 使用 Claude 模型 |
| `moai glm` | 仅 GLM | 所有 agent 经 Z.AI 代理使用 GLM 模型 |

## moai cc —— Claude 后端

```bash
moai cc [-p profile] [-w [name]] [-- claude-args...]
```

从 `.claude/settings.local.json` 中移除 GLM 专用环境变量,若 team 模式曾开启则将其重置,然后启动 Claude Code。

| 标志 | 说明 |
|--------|------|
| `-p, --profile <name>` | 使用命名的 Claude 配置(`~/.moai/claude-profiles/<name>/`) |
| `--permission-mode <mode>` | 指定权限模式 |
| `-b, --bypass` | `--permission-mode bypassPermissions` 的简写 |
| `-c, --continue` | 继续上一个会话 |
| `-m, --model <model>` | 覆盖模型选择 |
| `-w, --worktree [name]` | 在隔离的 git worktree(`.claude/worktrees/<name>/`)中启动 —— 省略名称时自动生成 |
| `--chrome` / `--no-chrome` | 原样传递给 Claude Code。启动器不会自行添加任一标志，因此除非传入 `--no-chrome`，否则可通过 `/chrome` 连接 |
| `-f, --factory` | 以**工厂主导**身份进入，不带参数。主导会话把运维者挑好的卡片通过跨会话消息整张分配给空闲 lane，lane 用 `-l` 加入 |
| `-l, --lane` | 以 **lane** 身份加入正在运行的工厂，自动领取下一个 `lane-<n>` 编号（存活 lane 中最大编号的下一个）。不带参数，没有正在运行的工厂时会被拒绝。`moai glm -l` 与 `moai codex -l` 行为相同 |
| `--leader <name>` | 只能与 `-l` 或 `--lane` 同用，指定要加入的主导会话（默认 `leader`，旧拼写 `lead` 会被拒绝）。当运行记录缺失或已退役而存活的主导会话仍在时，加入会验证该主导会话（pid + 进程启动）并恢复它的运行 |
| `--factory-run <run-id>` | 与 `-l` 同用：按 id 指定要加入的运行，不能与 `--leader` 同时使用 |
| `--clear-policy <value>` | 与 `moai cc -l` / `moai glm -l` 同用：lane 做完卡片后清理上下文的方式（默认 `clear-each`，另有 `clear-when-full`、`relaunch`） |
| `--no-auto-dispatch` | 与 `moai cc -l` / `moai glm -l` 同用：以手动模式启动 lane。默认是自主派单的 lane，会自己租用队列里的下一张卡片 |

{{< callout type="info" >}} 进入令牌只有 `-f`（主导）和 `-l`（lane）两个，都不带参数。一次启动只能带一个令牌，所以 `-f` 与 `-l` 同时出现会报错；`-f <值>`、`-l lane-2` 这类带值的形式，以及在 Codex 上请求主导（`moai codex` 上的 `-f`），都会被一行错误拒绝。已停用的 `-k` 入口同样被拒绝，并提示改用 `-f` 和 `-l`。完整约定见[工厂模式](/zh/advanced/factory-mode)与 [manager-lead 主导协调者](/zh/advanced/manager-lead)。 {{< /callout >}}

卡片整张进入一条 lane，并在其中按顺序走完 `plan → run → sync` 三个阶段。每个阶段由该会话拉起 `Agent()` 子智能体，有写权限的拉起用 `isolation: "worktree"` 隔离。一条 lane 同时最多运行 10 个子智能体，启动器会在 lane 会话中以 `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` 设置这个值，所以 N 条 lane 分摊机器容量，靠的是配置的保证，而不是运维者自觉。不要一次把所有 lane 全部启动：先起第一条，确认它确实开始产出，再启动其余的。

主导会话与 lane 可以使用不同的后端。后端组合先看 token 余量再定，一个可用的起点是只在需要重判断的位置用 Opus，以实现为主的 lane 跑在 GLM 上。换一种组合，或统一用同一个后端，同样没有问题。

权限模式为 `default`、`acceptEdits`（`moai init` 的默认值）、`plan`、`auto`、`bypassPermissions`、`dontAsk` 之一。`auto` 模式由后台分类器审查操作。支持的方案和模型请参阅 [Claude Code 权限模式文档](https://code.claude.com/docs/en/permission-modes)。

## moai glm —— GLM 后端

```bash
moai glm setup <api-key>   # 保存 API 密钥(首次一次)
moai glm                   # 以 GLM 后端启动
moai glm -p work           # 以 'work' 配置启动
moai glm status            # 检查凭据状态
```

从 `~/.moai/.env.glm` 读取 GLM 凭据,注入 `ANTHROPIC_AUTH_TOKEN`、`ANTHROPIC_BASE_URL` 等环境变量后启动 Claude Code。

| 子命令 | 说明 |
|-------------|------|
| `moai glm setup [api-key]` | 保存 GLM API 密钥 |
| `moai glm status` | 显示当前 GLM 凭据状态 |

{{< callout type="warning" >}}
GLM 不支持 `auto` 权限模式。请在符合条件的 Claude 会话中选择该模式。已停用的 CG 不能作为并发执行的替代方案。
{{< /callout >}}

## CG 停用与配置迁移

它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。 项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。 [CG 停用与配置迁移](/zh/multi-llm/cg-mode/) CG 已停用，请用 `moai migrate cg` 预览迁移选项。

```bash
moai migrate cg
moai migrate cg --target claude-only --apply --accept-role-change
```

迁移会写入 `llm.team_mode: claude`、`llm.gateway.teammate_mode: in-process` 和 `llm.gateway.teammate_provider: inherit`。这会取消原有混合角色分配，并不会保留 Claude 领队与 GLM 队友窗格的分工。

`claude-glm` 表示 Claude 领队搭配 tmux 中的 GLM 队友。目前 TEAMMATE 集成验证尚未通过，因此不能应用或启动该方案，只能预览。安装 tmux 或设置 `verified: true` 都不能解除限制。

## 配置文件(`-p` 标志)

受支持的启动器都可用 `-p <name>` 指定命名配置,此时 `CLAUDE_CONFIG_DIR` 会设为 `~/.moai/claude-profiles/<name>/`。用于分离运营多个账户·设置集。

## 隔离 worktree(`-w` 标志)

受支持的启动器都可用 `-w [name]` 在隔离的 git worktree 内启动会话,把先 `cd` 再启动的两步合并为一条命令。

```bash
moai cc -w feat-login    # 在 .claude/worktrees/feat-login/ 中启动
moai cc -w               # 自动生成名称
moai glm -w feat-login   # GLM 后端同理
```

行为规则:

- worktree 路径为 `.claude/worktrees/<name>/`。`<name>` 是 **worktree 名称**,既不是分支名也不是 SPEC ID。
- 若同名 worktree 已存在,则**复用而不重新创建**。因此它也可作为回到上一个会话工作树的再入路径。
- 省略名称时由 Claude Code 自动命名。
- `-w=name`、`--worktree name`、`--worktree=name` 三种写法含义相同,均被接受。
- `--` 之后的参数原样传递给 Claude Code,不受此改写影响。

{{< callout type="info" >}}
在会话交接中把 worktree 名称取成与 SPEC ID 相同(`moai cc -w SPEC-XXX-001`),下一个会话即可用一行命令回到同一工作树。
{{< /callout >}}

## 相关文档

- [CG 停用与配置迁移](/zh/multi-llm/cg-mode/)
- [配置文件管理](/zh/cli-reference/profile)
- [安全说明](/zh/advanced/security-notes) —— GLM 凭据路径安全模型
- [CLI 概览](/zh/getting-started/cli)
