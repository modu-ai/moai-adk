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
| `-k, --kanban [SPEC-ID]` | 进入看板主控 —— 把 `plan → run → sync` 链种进本会话。附上 SPEC-ID 时以该 SPEC 为目标 |
| `-k --name <role>` | 作为伴随会话加入已打开的看板 run。角色为 `plan` · `run` · `sync`。同一角色名已被活着的会话占用时取下一个编号 (`plan-1`, `plan-2`, …) |
| `-f, --factory` | 进入**工厂主控** —— 打开工厂 run，一名工作者（`worker-1`）。主控通过跨会话消息把操作者选中的卡片分给空闲工作者 |
| `-f worker` | 让一名工作者自动加入下一个空号，连到正在运行的工厂主控套接字 |
| `-f worker-<n>` | 精确启动那个编号（`worker-<n>`）的工作者。编号与存活的正规工作者冲突时顺延到下一个空号；与存活的旧式工作者（`agent-<n>`/`lane-<n>`）冲突则会被点名拒绝。`moai glm -f worker` / `-f worker-<n>` 在 GLM 后端上行为相同 |
| `-k <N>` / `-k <N> --name worker-<i>` | v1.2.0 的统一形式，至今仍然有效 —— `-k <N>` 是 N 名工作者 run 的主控，`-k <N> --name worker-<i>` 是其中的工作者 `<i>`。不带 N 只用 `-k --name worker-<i>` 时默认 8 名工作者 |
| `-f agent` · `-f lane-<n>` · `--name lane-<n>` | **已弃用别名。** 各自与 `-f worker` / `-f worker-<n>` / `--name worker-<n>` 行为完全相同，但每次启动都会提示改用规范拼写 |

{{< callout type="info" >}} `-k` 是看板链的标记，`-f` 是**工厂模式** (Factory Mode) 的专用进入标记。 `-k` 一个标记有三种解释这点没变 —— 不带参数 / 带 SPEC-ID 是看板主控，`--name <角色>` 是看板伴随会话，数字是工作者 run。 一次启动只能带一个进入标记，所以 `-k` 和 `-f` 同时给出会报错。 详细契约见[看板模式](/zh/advanced/kanban-mode)和 [manager-lead 领导协调者](/zh/advanced/manager-lead)。 {{< /callout >}}

卡片流转的方式与看板不同。看板里一张卡片在 `plan → run → sync` 各列之间移动，而工厂里一张卡片整个交给一名工作者，在该工作者内部按顺序走完三个阶段。每个阶段都由该会话启动 `Agent()` 子智能体来跑，其中承担写入的生成用 `isolation: "worktree"` 隔离。一名工作者最多同时启动 10 个子智能体，启动器会把这个值以 `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` 注入工作者与伴随会话，所以 N 名工作者分摊机器容量的结构由配置保证，而不是靠操作者自制。工作者不要一次全开 —— 先把第一名工作者拉起来，确认它真的开始产出之后再启动其余工作者。

后端组合先看 token 余量再定。一个可用的起点是：主控用 GLM、plan 用 Claude（Opus）、run 用 GLM、sync 用 Claude（Opus），只把 Opus 放在判断吃重的阶段。换别的组合、或统一到一个后端，同样没有问题。

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
