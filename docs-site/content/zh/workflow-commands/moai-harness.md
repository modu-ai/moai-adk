---
title: /moai harness 命令
weight: 55
draft: false
---

通过 Harness v4 Builder 创建和管理项目特定的动态专家团队。

{{< callout type="info" >}}
**斜杠命令**: 在 Claude Code 中输入 `/moai:harness <自然语言请求>` 即可直接执行该命令。
{{< /callout >}}

## 概述

`/moai:harness` 执行 MoAI-ADK 的 **Harness v4 Builder**，生成与项目需求相匹配的动态专家团队。

### Harness v4 Builder 是什么?

Harness v4 Builder 通过基于 Socratic 访谈的 4 阶段工作流 (ANALYZE → PLAN → GENERATE → ACTIVATE) 来构建团队。

| 阶段 | 说明 |
|------|------|
| ANALYZE | 分析项目结构、使用语言、现有代理清单 |
| PLAN | 确定所需团队规模 (3~5 人)、角色定义、worktree 隔离策略 |
| GENERATE | 生成 `.claude/agents/harness/` 代理文件、`.moai/harness/manifest.json` |
| ACTIVATE | 注册团队并启用 `/harness:<name>` 命令 |

## 使用方法

### 1 步: 用自然语言请求创建团队

```bash
> /moai:harness <自然语言请求>
```

**示例:**
```
为我们的 Go 后端项目创建专家团队。
我们需要分别处理 DB 迁移、REST API 端点、单元测试的团队。
```

### 2 步: Builder 自动处理

Builder 自动执行 4 阶段:

1. **ANALYZE**: 检测 Go, PostgreSQL, REST API 技术栈
2. **PLAN**: 决定 3 人团队 (DB Engineer, API Developer, Test Engineer)
3. **GENERATE**: 
   - `.claude/agents/harness/db-engineer.md`
   - `.claude/agents/harness/api-developer.md`
   - `.claude/agents/harness/test-engineer.md`
   - `.moai/harness/manifest.json` 生成
4. **ACTIVATE**: 注册 `/harness:backend-team` 命令

### 3 步: 使用生成的团队

生成后，所有后续工作中自动使用该团队:

```bash
/moai run SPEC-BACKEND-001
/moai run --team SPEC-BACKEND-001    # 强制团队模式
```

MoAI 分析 SPEC 复杂度并按照 manifest 的阶段序列自动委托团队成员。

## Harness 管理命令

### harness list

查看生成的所有 harness:

```bash
/harness list
```

### harness:<name> status

查看特定 harness 的详细信息:

```bash
/harness:backend-team status
```

输出信息:
- 团队成员列表和角色
- 使用的模型 (inherit, haiku, sonnet, opus)
- 可选 worktree 隔离设置
- Manifest 版本和生成日期

### harness:<name> edit

编辑 manifest.json 和代理定义:

```bash
/harness:backend-team edit
```

可修改项:
- 添加/删除团队成员
- 技能预加载列表
- Worktree 隔离策略
- 角色特定提示

### harness:<name> remove

删除 harness 和关联文件:

```bash
/harness:backend-team remove
```

删除项:
- `.claude/agents/harness/` 代理定义
- `.moai/harness/manifest.json`
- 注册的 `/harness:<name>` 命令
- worktree 隔离策略

## Manifest 结构

Harness v4 通过 **manifest.json** 定义团队组成。

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
          "role": "DB 设计和迁移",
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

### 阶段字段

| 字段 | 说明 |
|------|------|
| `name` | 阶段名称 (`plan`, `run`, `sync`) |
| `teammates` | 该阶段参与的团队成员数组 |

### 团队成员字段

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `name` | 必需 | 团队成员唯一标识 |
| `role` | 必需 | 团队成员的角色描述 |
| `model` | `inherit` | 模型选择 (`inherit`, `haiku`, `sonnet`, `opus`) |
| `skills` | `[]` | 预加载的技能列表 |

## Worktree 隔离

Harness v4 支持可选的 worktree 隔离。

### L1_optional (默认)

```json
"worktree_isolation": "L1_optional"
```

Claude Code 在并行团队成员间检测到冲突时自动创建 L1 worktree。

- **可选**: 仅在冲突时启用隔离
- **自动**: 运行时在冲突后自动生成
- **成本**: worktree 隔离时内存增加

### none

```json
"worktree_isolation": "none"
```

所有团队成员在项目根目录工作 (最小内存使用)。

## 团队委托工作流

Harness 激活后，MoAI 自动利用该团队。

### SPEC 执行时的团队委托

```bash
> /moai run SPEC-BACKEND-001
```

**MoAI 的自动判断:**
1. 估计 SPEC 复杂度 (文件数、代码行数)
2. 选择合适的 harness
3. 按 manifest 阶段顺序顺序/并列委托团队成员

### 基于阶段的委托示例

```
PLAN 阶段:
  → architect 团队成员负责架构设计

RUN 阶段:
  → db-engineer、api-developer 并列委托
  → test-engineer 顺序委托 (测试)

SYNC 阶段:
  → 文档生成和 PR 撰写 (默认 manager-docs)
```

## 自然语言请求的力量

Harness v4 Builder 通过 Socratic 访谈方式理解需求。

### 有效请求示例

```
我们的团队正在开发 Python FastAPI 后端。
我们需要擅长 API 端点、数据验证、错误处理的团队。
```

Builder 自动:
- 检测 Python、FastAPI、asyncio 技术栈
- 决定 3~5 人团队规模
- 设置每个团队成员的特化领域
- 预加载必要技能

### 模糊请求由 Builder 澄清

```
我需要一个团队。

→ Builder: 项目的主要技术是? (语言、框架)
→ Builder: 团队应关注哪个领域? (后端、前端、全栈)
→ Builder: 特别需要什么专业性?
```

## 相关文档

- [Harness v4 Builder 指南](/zh/advanced/builder-agents) - Builder 4 阶段详情
- [代理指南](/zh/advanced/agent-guide) - 8 个核心代理理解
- [SPEC 基础开发](/zh/workflow-commands/moai-plan) - SPEC 工作流概览

{{< callout type="info" >}}
**提示**: Harness 创建一次后，所有后续工作中该团队会自动使用。可通过 `/harness:team-name` 命令随时重复使用。
{{< /callout >}}
