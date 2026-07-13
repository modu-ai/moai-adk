---
title: /moai harness
weight: 55
draft: false
---

创建项目专属的动态专家团队(挽具),并管理挽具学习生命周期。

{{< callout type="info" >}}
**斜杠命令**: 在 Claude Code 中输入 `/moai:harness <自然语言请求>` 即可直接执行此命令。
{{< /callout >}}

## 概述

`/moai:harness` 运行 MoAI-ADK 的 **Harness v4 Builder**,自动生成契合项目需求的动态专家团队。

这条命令可以让您直接体验 v3 的第三大支柱 — **智能体挽具**:挽具创造挽具的递归结构。当通用智能体目录无法覆盖项目特有领域(例如特定的数据库迁移流程、公司内部 API 规范)时,只需一句自然语言,即可为该领域搭建一支专家团队。生成的挽具还与 **递归式自我学习** 子系统相连 — 随着使用观察的积累,挽具会自行生成改进建议,并在通过用户批准门禁后使指令持续演化。

### 什么是 Harness v4 Builder?

Harness v4 Builder 通过基于苏格拉底式访谈的 4-phase 工作流(ANALYZE → PLAN → GENERATE → ACTIVATE)组建团队。

| 阶段 | 说明 |
|------|------|
| ANALYZE | 分析项目结构、使用语言、现有智能体清单 |
| PLAN | 决定所需团队规模(3~5 人)、每位成员的角色、是否使用 worktree 隔离 |
| GENERATE | 生成 `.claude/agents/harness/` 智能体文件、`.moai/harness/manifest.json` |
| ACTIVATE | 注册团队并激活 `/harness:<name>` 命令 |

## 使用方法

### 第 1 步: 用自然语言请求创建团队

```bash
> /moai:harness <自然语言请求>
```

**示例:**
```
请为我们的 Go 后端项目组建一支合适的专家团队。
需要分别负责数据库迁移、REST API 端点、单元测试的团队。
```

### 第 2 步: Builder 自动处理

Builder 自动执行 4-phase:

1. **ANALYZE**: 检测 Go、PostgreSQL、REST API 技术栈
2. **PLAN**: 决定组建 DB Engineer、API Developer、Test Engineer 3 人团队
3. **GENERATE**:
   - `.claude/agents/harness/db-engineer.md`
   - `.claude/agents/harness/api-developer.md`
   - `.claude/agents/harness/test-engineer.md`
   - 生成 `.moai/harness/manifest.json`
4. **ACTIVATE**: 注册 `/harness:backend-team` 命令

### 第 3 步: 使用生成的团队

生成后,所有工作都会自动使用该团队:

```bash
/moai run SPEC-BACKEND-001
/moai run --team SPEC-BACKEND-001    # 强制团队模式
```

MoAI 分析 SPEC 复杂度后,按照 manifest 的 phase 顺序自动向团队成员委派工作。

## Harness 管理命令

### harness list

查看所有已创建的挽具列表:

```bash
/harness list
```

### harness:<name> status

查看特定挽具的详细信息:

```bash
/harness:backend-team status
```

输出信息:
- 团队成员列表及角色
- 使用的模型 (inherit, haiku, sonnet, opus)
- 可选的 worktree 隔离设置
- Manifest 版本及创建日期

### harness:<name> edit

编辑 manifest.json 与智能体定义:

```bash
/harness:backend-team edit
```

可修改的项目:
- 添加/移除团队成员
- 技能预加载列表
- Worktree 隔离策略
- 各角色的提示词

### harness:<name> remove

删除挽具及关联文件:

```bash
/harness:backend-team remove
```

被删除的项目:
- `.claude/agents/harness/` 智能体定义
- `.moai/harness/manifest.json`
- 已注册的 `/harness:<name>` 命令
- Worktree 隔离策略

## 挽具学习生命周期 — 递归式自我学习

挽具不是创建完就结束的静态产物。通过 `/moai harness` 子命令管理 **学习子系统** 的生命周期。

| 命令 | 说明 |
|--------|------|
| `moai harness status` | 查看学习状态(观察数、模式、建议) |
| `moai harness apply` | 应用建议(需通过用户批准门禁) |
| `moai harness rollback` | 回滚上一次应用 |
| `moai harness disable` | 禁用学习 |
| `moai harness list` (v4) | 列出所有学习规则 |
| `moai harness edit` (v4) | 直接编辑规则 |
| `moai harness remove` (v4) | 删除规则 |
| `moai harness doctor` (v4) | 诊断学习系统 |

**4 层学习阶梯** — 观察越积累,学习层级越高:

| Tier | 观察数 | 行为 |
|------|---------|------|
| TierObservation | ≥1 | 单纯记录 |
| TierHeuristic | ≥3 | 模式识别 |
| TierRule | ≥5 | 形成规则 |
| TierAutoUpdate | ≥10 | 自动更新(必须经用户批准) |

**产出物**: `.moai/harness/` 目录 (usage-log.jsonl, learned-rules.yaml)

{{< callout type="warning" >}}
自动演化始终只在 **用户批准门禁** 之下应用。评估者与批准权限位于演化回路之外,并且随时可以用 `moai harness rollback` 恢复。
{{< /callout >}}

## Manifest 结构

Harness v4 通过 **manifest.json** 定义团队构成。

### manifest.json 示例

```json
{
  "spec_id": "HARNESS-BACKEND-001",
  "name": "Backend Development Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "worktree_isolation": "L1_optional",
  
  "phases": [
    {
      "name": "plan",
      "teammates": [
        {
          "name": "architect",
          "role": "API 架构专家",
          "model": "inherit",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "teammates": [
        {
          "name": "db-engineer",
          "role": "数据库设计与迁移",
          "model": "inherit"
        },
        {
          "name": "api-developer",
          "role": "REST API 端点",
          "model": "inherit"
        },
        {
          "name": "test-engineer",
          "role": "单元测试",
          "model": "haiku"
        }
      ]
    }
  ]
}
```

### Phase 字段

| 字段 | 说明 |
|------|------|
| `name` | 阶段名称 (`plan`, `run`, `sync`) |
| `teammates` | 参与此阶段的团队成员数组 |

### Teammate 字段

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `name` | 必填 | 团队成员唯一标识符 |
| `role` | 必填 | 团队成员的角色描述 |
| `model` | `inherit` | 模型选择 (`inherit`, `haiku`, `sonnet`, `opus`) |
| `skills` | `[]` | 需要预加载的技能列表 |

可以为每位成员指定不同的模型(`model` 字段),这是令牌经济学设计的延伸 — 像架构决策这样推理繁重的角色,与像重复编写测试这样轻量的角色,没有理由使用同一个模型。

## Worktree 隔离

Harness v4 支持可选的 worktree 隔离。

### L1_optional (默认值)

```json
"worktree_isolation": "L1_optional"
```

当 Claude Code 检测到并行团队成员之间的冲突时,会自动创建 L1 worktree。

- **可选**: 仅在冲突时应用隔离
- **自动**: 运行时检测到冲突后自动创建
- **成本**: worktree 隔离时内存占用增加

### none

```json
"worktree_isolation": "none"
```

所有团队成员都在项目根目录工作(内存占用最小)。

## 团队委派工作流

挽具激活后,MoAI 会自动使用相应的团队。

### 执行 SPEC 时的团队委派

```bash
> /moai run SPEC-BACKEND-001
```

**MoAI 的自动判断:**
1. 估算 SPEC 复杂度(文件数、代码行数)
2. 选择合适的挽具
3. 按照 manifest 的 phase 顺序,顺序/并行委派团队成员

### 基于 Phase 的委派示例

```
PLAN Phase:
  → architect 成员负责架构设计

RUN Phase:
  → db-engineer、api-developer 并行委派
  → test-engineer 顺序委派(测试)

SYNC Phase:
  → 生成文档与撰写 PR(默认 manager-docs)
```

## 自然语言请求的力量

Harness v4 Builder 以苏格拉底式访谈的方式了解需求。

### 高效请求示例

```
我们团队正在开发 Python FastAPI 后端。
需要一支擅长 API 端点、数据校验、错误处理的团队。
```

Builder 会自动:
- 检测 Python、FastAPI、asyncio 技术栈
- 决定 3~5 人的团队规模
- 设定每位成员的专长领域
- 预加载所需技能

### 请求不明确时 Builder 会主动提问

```
我需要一个团队。

→ Builder: 项目的主要技术是什么?(语言、框架)
→ Builder: 团队应聚焦的领域是?(后端、前端、全部)
→ Builder: 是否有特别需要的专长?
```

## 相关文档

- [Harness v4 Builder 指南](/advanced/builder-agents) - Builder 4-phase 详解
- [智能体指南](/advanced/agent-guide) - 理解 10 个智能体目录
- [基于 SPEC 的开发](/workflow-commands/moai-plan) - SPEC 工作流概览

{{< callout type="info" >}}
**提示**: 挽具创建一次后,所有后续工作都会自动使用该团队。也可以随时通过 `/harness:team-name` 命令复用。
{{< /callout >}}
