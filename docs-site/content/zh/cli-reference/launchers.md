---
title: moai cc / cg / glm 启动器
weight: 15
draft: false
---

`moai cc`、`moai cg`、`moai glm` 是以不同后端配置启动 Claude Code 的三个启动器。三个命令都会先调整设置,再用 `exec` 将当前进程替换为 Claude Code。哪个模型承担哪种工作直接决定成本,因此启动器的选择是省下成本的第一步。

## 三个启动器对比

| 启动器 | 后端 | 用途 |
|------|--------|------|
| `moai cc` | 仅 Claude | 标准执行 —— 所有 agent 使用 Claude 模型 |
| `moai glm` | 仅 GLM | 所有 agent 经 Z.AI 代理使用 GLM 模型 |
| `moai cg` | Claude + GLM 混合 | 领导者用 Claude,队员用 GLM(节省 60-70% 成本) |

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
| `--chrome` / `--no-chrome` | 切换 Chrome MCP |
| `-k, --kanban [SPEC-ID]` | 进入看板主控 —— 把 `plan → run → sync` 链种进本会话。附上 SPEC-ID 时以该 SPEC 为目标 |
| `-k --name <role>` | 作为伴随会话加入已打开的看板 run。角色为 `plan` · `run` · `sync`。同一角色名已被活着的会话占用时取下一个编号 (`plan-1`, `plan-2`, …) |
| `-k <N>` | 以**编号 worker 运行**的主控进入 —— N 个编号 worker（`worker-1`…`worker-N`）的看板 run。主控把积压区里的卡片分给空闲的 worker |
| `-k <N> --name worker-<i>` | 以编号 worker `<i>` 进入。编号与活着的会话冲突时取下一个编号。不带 N 只用 `--name worker-<i>` 时默认 8 个 |

{{< callout type="info" >}}
同一个 `-k` 标记有三种解释 —— 不带参数 / 带 SPEC-ID 是看板主控，`--name <角色>` 是看板伴随会话，数字是编号 worker 运行。混合后端启动器 `moai cg` 下两者都被拒绝。改名前的旧名字 `-f`/`--factory` 旗标已退役，现在会得到明确的报错 —— 详细契约见[看板模式](/zh/advanced/kanban-mode)。
{{< /callout >}}

权限模式为 `default`、`acceptEdits`(项目默认)、`plan`、`auto`、`bypassPermissions`、`dontAsk` 之一。`auto` 模式由后台分类器检查动作,需要 Team 方案 + Sonnet/Opus 4.6 及以上。

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
GLM 不支持 `auto` 权限模式(第三方提供商)。若需要 `auto`,请使用 `moai cc` 或 `moai cg`。此外 Z.AI 的并发请求上限较低(付费层 1-3 in-flight),因此多 agent 并行执行用 `moai cg` 混合模式更稳定。
{{< /callout >}}

## moai cg —— Claude + GLM 混合

```bash
moai cg [-p profile]
```

CG 是 "Claude + GLM" 的缩写,是成本优化的团队组合。

- **领导者** (当前 tmux pane):使用 Claude 模型(opus/sonnet)
- **队员** (新 tmux pane):经 Z.AI 代理使用 GLM 模型

执行时会验证 tmux 会话,在领导者 pane 中移除 GLM 环境(Claude),向 tmux 会话注入 GLM 环境(队员),并设置 `teammateMode=tmux` 与 `team_mode: cg`。

**前置条件**:

1. 用 `moai glm setup <api-key>` 设置 GLM API 密钥
2. 在 tmux 会话内部执行以实现 pane 级环境隔离

## 配置文件(`-p` 标志)

三个启动器都可用 `-p <name>` 指定命名配置,此时 `CLAUDE_CONFIG_DIR` 会设为 `~/.moai/claude-profiles/<name>/`。用于分离运营多个账户·设置集。

## 隔离 worktree(`-w` 标志)

三个启动器都可用 `-w [name]` 在隔离的 git worktree 内启动会话,把先 `cd` 再启动的两步合并为一条命令。

```bash
moai cc -w feat-login    # 在 .claude/worktrees/feat-login/ 中启动
moai cc -w               # 自动生成名称
moai glm -w feat-login   # GLM 后端同理
moai cg -w feat-login    # 混合模式同理
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

- [CG 模式(Claude + GLM)](/zh/multi-llm/cg-mode)
- [配置文件管理](/zh/cli-reference/profile)
- [安全说明](/zh/advanced/security-notes) —— GLM 凭据路径安全模型
- [CLI 概览](/zh/getting-started/cli)
