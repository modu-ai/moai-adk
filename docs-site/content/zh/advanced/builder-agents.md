---
title: 构建者代理与 Harness v4
weight: 40
draft: false
---

详细介绍扩展 MoAI-ADK 的 Harness v4 Builder。

{{< callout type="info" >}}
  **一句话总结**: Harness v4 Builder 通过自然语言请求动态生成项目特定的专家团队。包括 4 阶段工作流 (ANALYZE → PLAN → GENERATE → ACTIVATE) 和 manifest 基础 Runner。
{{< /callout >}}

## Harness v4 Builder 是什么?

Harness v4 Builder 通过 `/moai:harness <自然语言请求>` **动态生成项目特定的专家团队**。

### 与前版本的区别

| 分类 | 前版 (v3/静态模型) | 现版 (v4 Builder) |
|------|-----|-----------|
| 生成方式 | 3 个构建器代理 (构建器-技能、构建器-代理、构建器-插件) | 单一 Harness v4 Builder (动态生成) |
| 工作流 | 用户定义结构 | 4 阶段 ANALYZE → PLAN → GENERATE → ACTIVATE |
| 执行方式 | 各自独立 | Manifest 基础 Runner (可选 worktree 隔离) |
| 可扩展性 | 受限 | 项目上下文自动检测 |

## Harness v4 Builder 4 阶段工作流

### 1. ANALYZE (分析阶段)

分析当前项目并识别所需的专业性。

- 源代码结构分析
- 使用语言和框架检测
- 现有代理/技能清单调查
- 项目规模估计

### 2. PLAN (规划阶段)

定义所需专家团队的组成和角色。

- 团队规模决定 (3~5 成员)
- 每个团队成员的角色档案定义
- worktree 隔离必要性判断
- Manifest 架构设计

### 3. GENERATE (生成阶段)

生成实际的代理定义和配置。

- 在 `.claude/agents/harness/` 下生成代理文件
- 生成 `.moai/harness/manifest.json` (Runner 配置)
- 撰写角色特定的系统提示
- 定义技能预加载列表

### 4. ACTIVATE (激活阶段)

使生成的 harness 可立即使用。

- 代理注册和验证
- Manifest Runner 初始化
- 可选 worktree 生成和隔离设置
- 启用团队成员自动委托规则

## Manifest 基础 Runner

Harness v4 使用 **Manifest 基础 Runner** 来操作生成的团队。

### manifest.json 结构

```json
{
  "spec_id": "HARNESS-PROJECT-001",
  "name": "My Project Custom Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "phases": [
    {
      "name": "plan",
      "teammates": [
        {
          "name": "researcher",
          "model": "haiku",
          "mode": "plan",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "teammates": [
        {
          "name": "implementer",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        }
      ]
    }
  ],
  "worktree_isolation": "L1_optional"
}
```

### Runner 动作

1. **阶段进入**: 按照 manifest 的阶段序列进行
2. **团队成员生成**: 动态生成每个阶段的团队成员
3. **隔离应用**: 应用条件 worktree 隔离
4. **结果聚合**: 聚合每个团队成员的结果

## Harness 生命周期命令

由 Harness v4 Builder 生成的 harness 通过 `/harness:<name>` 命令管理。

### 可用命令

```bash
# 生成的 harness 列表
/harness list

# 特定 harness 的状态
/harness:my-project-team status

# 编辑 harness 配置
/harness:my-project-team edit

# 删除 harness
/harness:my-project-team remove

# 使用 Harness v4 Builder 创建新 harness
/moai:harness <自然语言请求>
```

## 通过自然语言请求创建 harness

### 基本用法

```bash
> 为我们的后端项目创建专家团队。
> 我们需要处理 API 端点、DB 架构、测试的团队。
```

### Builder 的动作流程

1. ANALYZE: 分析项目结构 (Go, PostgreSQL, REST API)
2. PLAN: 决定 3 人团队 (API Designer, DB Specialist, Test Engineer)
3. GENERATE: 生成每个代理定义和 manifest.json
4. ACTIVATE: 激活团队并注册 `/harness:backend-team` 命令

### 生成结果位置

- 代理定义: `.claude/agents/harness/api-designer.md`, `db-specialist.md`, ...
- Manifest: `.moai/harness/manifest.json`
- 可选 worktree: `~/.moai/worktrees/<project>/` (用户可选)

## Worktree 隔离 (可选)

Harness v4 支持条件 worktree 隔离。

### L1 隔离 (Optional)

Claude Code 运行时为每个代理生成 L1 worktree。

- **何时使用**: 并行团队成员编辑同一文件时
- **隔离范围**: 每个团队成员的文件写入发生在独立 worktree 中
- **成本**: 额外内存 + 并列收益相抵

### 禁用

在 manifest 中设置 `"worktree_isolation": "none"` 以跳过 L1 隔离。

## 相关文档

- [Harness v4 Builder 进阶指南](/zh/advanced/harness-v4-builder) - Builder 4 阶段详情和 manifest 架构
- [代理指南](/zh/advanced/agent-guide) - 8 个核心代理目录
- [动态工作流](/zh/advanced/ultracode-workflows) - `/effort ultracode` 并列执行

{{< callout type="info" >}}
**提示**: Harness v4 Builder 为每个项目**仅生成一次自定义团队**，之后该团队会在所有后续工作中自动使用。初始生成后，可通过 `/harness:team-name` 随时重复使用。
{{< /callout >}}
