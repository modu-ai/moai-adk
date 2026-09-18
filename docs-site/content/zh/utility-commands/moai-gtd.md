---
title: /moai gtd
weight: 29
draft: false
new: true
---

# /moai gtd

`/moai gtd` 与 `moai gtd` 是正式的 GTD 入口：先收集工作、判断是否可执行，再把获准的工作接入现有开发队列。原有 `todo` SQLite 数据库、卡片 ID、顺序、`queued`/`picked`/`dropped` 状态以及归档和恢复语义均不改变。`todo` 作为共用同一命令树的兼容名称继续保留。

```bash
moai gtd add "整理认证错误路径"
moai gtd list
moai gtd next t1 --spec SPEC-AUTH-001
moai gtd done t1 --expect "认证"
```

现有 `moai todo ...` 调用也会得到相同结果。完整的队列动词与参数见 [todo 兼容命令参考](/zh/utility-commands/moai-todo)。

五个 GTD 专用动词会沿用同一条 SQLite 记录。

```bash
moai gtd capture "整理认证错误路径" --event inbox-42 --source user --sensitivity private
moai gtd clarify <gtd-id> --disposition action --outcome "已合并" --evidence "测试与合并 SHA" --authority "queue,commit" --trusted
moai gtd organize <gtd-id> --class action --context computer --depends-on <gtd-id>
moai gtd reflect --rebuild-projection --json
moai gtd engage <gtd-id> --approve --fresh --dependencies-ready --resources --pick --dispatch --lane lane-10 --run-id mission-42
```

`capture` 必须提供稳定的 `--event`，以便重试时去重。`engage` 要执行调度，必须同时带上 `--pick`、`--lane` 与 `--run-id`；批准、新鲜度、依赖或资源确认缺少任一项都不会产生效果。操作 receipt 会先写入 SQLite；中断恢复时先读取同一操作 ID 的权威结果，再决定是否重试。

需要明确 opt-in 的 GTD v2 导出/导入格式，不仅保存条目和关系，也保存封存合同、任务、事件与操作 receipt，并在导入时校验完整性。卡片 archive/reopen 保持同一 GTD 身份；`reflect` 会重新读取已归档完成、重开、取消和 source revision 变化，从而标出陈旧证据与受阻后续项。

## 五个步骤

GTD 是**整理并选择工作的流程**，不是开发看板的新列。

| 步骤 | 判断与保存结果 |
|---|---|
| Capture | 记录来源、敏感级别和去重标识，但不创建开发卡片。 |
| Clarify | 检查目标结果、完成证据、权限与来源可信度；任一项未确定便暂缓发布。 |
| Organize | 保存项目、行动、参考或暂缓分类，以及执行上下文、复查时间和关系；拒绝循环依赖。 |
| Reflect | 重新审视证据变化、归档、重开、取消和定期复查事件；取消不算前置工作完成。 |
| Engage | 只有通过审批范围、证据新鲜度、依赖、泳道所有权和资源上限的条目，才会被建议或接入队列。 |

开发仍沿用 `backlog → plan → run → sync → done`。Capture 至 Engage 不会取代这些阶段。

## 关系与私有图谱

只有 `depends_on` 会阻塞执行。`part_of`、`supported_by`、`related_to` 都是非阻塞关系；原有 `contains`、`absorbs`、`replaces`、`conflicts` 的含义也继续保留。GTD 图谱是队列数据库旁的私有派生物，默认不进入仓库、日志、遥测、导出或备份。元数据或来源 revision 不一致时会被判为陈旧。

## 自主运行与安全边界

LLM 与 `mission-governor` 只生成结构化提案，不直接改动文件、Git、队列或调度状态。普通代码会先检查用户批准并封存的目标、范围、允许行为、完成证据、资源上限、停止条件与当前 snapshot，之后才准备操作。扩大范围、证据过期或未经授权的行为不会获得默认批准，而会停在 `blocked`。

当前实现提供负责确定性策略、恢复、调度、显式路径提交和 local develop `--no-ff` 合并的 owner adapter。提交必须持有与当前 HEAD 对应、位于仓库内且权限为 `0600` 的测试 receipt；本地合并会重新检查 manager-git 角色、`WT-*` 分支、基准 SHA 和 `.git` 下的 `0600` lease。备份、恢复与导出只有在明确 opt-in 时才包含 GTD 扩展，private projection 则从 SQLite revision 重建。在真实供应方证明启动、重连、替换、凭证和进程身份能力全部可用之前，不承诺会话结束后的持续运行，也不承诺远程 push、PR 或合并完成。

相关：[`/moai goal --auto`](/zh/utility-commands/moai-goal#auto-任务模式) · [看板模式](/zh/advanced/kanban-mode)
