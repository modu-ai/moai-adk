---
title: "Codex 双 harness — AGENTS.md、代理双重发布、钩子适配器"
weight: 31
draft: false
added_in: "v3.1.3"
description: "让 codex-cli 与 Claude Code 并行使用 MoAI-ADK 的共用表面与各 harness 个人指令。"
---

MoAI-ADK 的第一 harness(实际驱动代理的运行环境)是 Claude Code,但从 v3.1.3 起它带上了一层 **codex-cli 也能读的二元表面**。共用规则与代理定义会按 Codex 寻找的位置和形式一并发布,个人指令则按 harness 分开。本文说明这些表面各自解决什么问题。

## 根 AGENTS.md —— harness 通用的 standing contract

仓库根目录的 `AGENTS.md` 不是 Claude 专属文件,而是**无论哪个代理 harness 驱动一回合都适用的 standing contract**(常时契约)。它以单一文件存在,原因在 codex 的读取方式: codex 在字节上限内读项目指示,**超出的尾部会被静默丢弃 —— 没有警告,退出码还是 0**。装不进上限的契约会以"完整"的姿态汇报。所以"装得进上限"本身就是需求,由 build guard(构建时检查该文件是否在上限之内的机制)把守。

为了腾出空间,11 份常驻加载文档降级成了指向 8 份惰性伴随文档(lazy companion,按需才读的详述文档)的存根(简短摘要)。**搬走的是解释义务的文字,从来不是义务本身** —— `AGENTS.md` 是各 harness 共用契约的基准,`.claude/rules/moai/**` 和 `CLAUDE.md` 补充 Claude 专用机制。

{{< callout type="info" >}}
个人的 `~/.codex/AGENTS.md` 会加入同一条合并链,并在本文件**之前**被消费,压缩项目契约能承载的宽度。溢出从尾部开始静默丢弃 —— 这正是本文件的条款按最重要在前排序的原因。
{{< /callout >}}

## 代理双重发布 —— 十一份 TOML

保留的 11 个代理以两种形式发布: Claude Code 用的 `.claude/agents/moai/*.md`(原件),和 codex 读的 `.codex/agents/moai/*.toml`(派生)。TOML 不是手写的 —— `internal/template/agentemit` 从 markdown 原件**确定性地**(同样输入永远同样输出)生成,生成文件的开头钉着一句 "regenerate, do not edit"(重新生成,别直接改)。

原件与派生之间的漂移由三层护栏挡住: golden 文件比对(与期望输出对照)、嵌入校验(与编进二进制的模板对照)、部署校验(与落到用户仓库的结果对照)。改 markdown,TOML 跟着来;只改 TOML,护栏会抓住。

## `.agents/skills` —— 技能镜像

codex-cli 不读 Claude Code 的 `.claude/skills/`,所以技能以**镜像**(复写副本)形式部署到 `.agents/skills` 下。镜像清单不是手工维护的,而是在部署执行时从实际的技能集合导出 —— 技能增减,清单不会过期。这个目录是面向**用户仓库之外**的部署产物,不进 git;优先符号链接,无法创建链接的环境退回复制部署(`moai init`、`moai update` 的完成摘要会说明这一点 —— 详见 [moai update](/zh/cli-reference/update/) 文档)。

## 各 harness 的个人指令

`AGENTS.local.md` 仅供 Codex 使用。本地 `moai codex` 启动器从项目根目录读取它,并把原始内容作为 Codex 会话的 `developer_instructions` 覆盖项传入;共享文件不会用 `@` 导入它。`CLAUDE.local.md`、`.claude/settings.local.json` 和 Claude 自动生成的 `MEMORY.md` 仍仅供 Claude 使用。Codex Web 会话不经过本地启动器,因此不会收到这项注入。

## `internal/codexadapter` —— 钩子适配器库

两个 harness 的钩子表面几乎相同,但不完全相同。实测(以 codex-cli 0.153.4 为准)发现分歧有三处: harness 传入的**事件名**、codex 声明了却不会响应的**三个输出键**(`systemMessage`、`continue`、`stopReason`),以及 **PreToolUse 决定契约** — codex 解析器会拒绝没有 `updatedInput` 的 `permissionDecision:allow`,并全面拒绝 `permissionDecision:ask`。`internal/codexadapter` 是坐在分发器**前面**的薄翻译层(`internal/hook` 不被触碰);被拒绝的决定形态(allow、ask、defer)降级为无意见 `{}`,交给 codex 自己的审批流程,每次降级都通过 discard sink 通报,无理由的 deny 会补上默认理由。

### 12 事件表

| Codex 事件 | MoAI 分发器参数 | 本里程碑适配? |
|---|---|---|
| PreToolUse | `pre-tool` | 是 |
| PostToolUse | `post-tool` | 是 |
| SessionStart | `session-start` | 是 |
| SessionEnd | `session-end` | 是 |
| Stop | `stop` | 是 |
| UserPromptSubmit | `user-prompt-submit` | 是 |
| PreCompact | `compact` | 否 —— 非交互运行中压缩从未触发 |
| PostCompact | `post-compact` | 否 —— 非交互运行中压缩从未触发 |
| PermissionRequest | `permission-request` | 否 —— 非交互运行中审批请求从未出现 |
| SubagentStart | `subagent-start` | 是 |
| SubagentStop | `subagent-stop` | 是 |
| Interrupt | —— (无对应物) | 否 —— 在 SIGINT 时触发;适配需要新的分发器子命令(后续卡片) |

共覆盖 12 个事件: 11 个有分发器对应物,官方文档中的第 12 个事件 `Interrupt` 在对应物出现之前,会以一条明确的"无对应物"消息被识别后拒绝。在 codex-cli 0.153.4 实测中,`SubagentStart` 与 `SubagentStop` **确认触发**(推翻了此前 0.147.0 下 SubagentStop 不触发的观测),现已适配 —— `RenderHooks` 会把 `moai hook subagent-start --harness codex` 和 `moai hook subagent-stop --harness codex` 两行写入用户的 `.codex/hooks.json`。compact、permission 系列是基于诚实证据而非假设的保留: 非交互的 `codex exec` 运行从未达到压缩条件(实测输入上限 1,048,576 字符,最好成绩 264,808 输入 token),也从未引发审批请求 —— 记录为**触发未达成**,绝不记作"不触发"。

未适配事件不会被静默无视,而是被**拒绝**。未知事件(拼错)和被识别但不处理的事件(范围决定,或像 `Interrupt` 那样缺少对应物)返回不同的错误,运营者能把失误和决定区分开。配置校验器不停在第一个,而是**收齐所有**未知键违规后一次展示。

### 现在的调用方

`RenderHooks` 会把 8 个已适配事件的命令写入用户的 `.codex/hooks.json`。`moai init --llm codex|both` 创建这套配线;已有项目可用 `moai tool enable codex` 添加或刷新。

## 下一步

- [多模型审计收敛](/zh/advanced/multi-model-audit/) —— codex 后端如今已经参与审计的路径
- [moai update](/zh/cli-reference/update/) —— 技能镜像的 symlink·复制部署及其通知
- [代理指南](/zh/advanced/agent-guide/) —— 被双重发布的 11 个代理的角色
