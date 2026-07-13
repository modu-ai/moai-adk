---
title: 构建器智能体与 Harness v4
weight: 40
draft: false
---

智能体 Harness 的最后一块拼图是递归 — Harness 制造 Harness。Harness v4 Builder 就是这个递归结构的入口：用一句自然语言请求生成项目专属的专家团队。

{{< callout type="info" >}}
**一句话总结**：Harness v4 Builder 用自然语言请求动态生成项目专属的专家团队。它由 4 阶段工作流（ANALYZE → PLAN → GENERATE → ACTIVATE）和基于 manifest 的 Runner 组成。
{{< /callout >}}

## 什么是 Harness v4 Builder？

Harness v4 Builder 通过 `/moai:harness <自然语言请求>` **动态生成项目专属的专家团队**。

通用智能体目录（11 个）对所有项目通用，而 Builder 生成的 Harness 是只存在于你项目中的定制团队。

### 与旧版本的差异

| 区分 | 旧版（v3/静态模型） | 现在（v4 Builder） |
|------|-----|-----------|
| 生成方式 | 3 种构建器智能体（builder-skill、builder-agent、builder-plugin） | 单一 Harness v4 Builder（动态生成） |
| 工作流 | 用户自定义结构 | 4-phase ANALYZE → PLAN → GENERATE → ACTIVATE |
| 执行方式 | 各自独立 | 基于 Manifest 的 Runner（可选 worktree 隔离） |
| 扩展性 | 有限 | 自动感知项目上下文 |

## Harness v4 Builder 4-Phase Workflow

### 1. ANALYZE（分析阶段）

分析当前项目，把握所需的专业能力。

- 分析源代码结构
- 检测所用语言与框架
- 盘点既有智能体/技能
- 估算项目规模

### 2. PLAN（计划阶段）

定义所需专家团队的组成与角色。

- 决定团队规模（3~5 名成员）
- 定义每名成员的角色档案
- 判断是否需要 worktree 隔离
- 设计 Manifest 模式

### 3. GENERATE（生成阶段）

生成实际的智能体定义与配置。

- 在 `.claude/agents/harness/` 下生成智能体文件
- 生成 `.moai/harness/manifest.json`（Runner 配置）
- 编写按角色的系统提示词
- 定义技能预加载列表

### 4. ACTIVATE（激活阶段）

激活生成的 Harness，使其立即可用。

- 注册并校验智能体
- 初始化 Manifest Runner
- 可选的 worktree 创建与隔离设置
- 启用团队成员自动委派规则

## 基于 Manifest 的 Runner

Harness v4 使用 **基于 Manifest 的 Runner** 运营生成的团队。哪个 phase 投入哪个成员、使用什么模型与权限模式，都声明在一份 manifest 文件中 — 用声明式管理模型分配的代币经济学原则同样适用于此。

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

### Runner 的运行

1. **Phase 进入**：按 manifest 的 phase 序列推进
2. **Teammate Spawn**：动态生成每个 phase 的 teammates
3. **Isolation 应用**：按条件应用 worktree 隔离
4. **Result Aggregation**：整合每个 teammate 的结果

## Harness Lifecycle Commands

用 Harness v4 Builder 生成的 Harness 通过 `/harness:<name>` 命令管理。

### 可用命令

```bash
# 查看已生成的 Harness 列表
/harness list

# 查看特定 Harness 状态
/harness:my-project-team status

# 编辑 Harness 配置
/harness:my-project-team edit

# 删除 Harness
/harness:my-project-team remove

# 用 Harness v4 Builder 生成新 Harness
/moai:harness <自然语言请求>
```

## 用自然语言请求生成 Harness

### 基本用法

```bash
> 给我们的后端项目组建一支合适的专家团队。
> 需要一个负责 API 设计、DB 模式和测试的团队。
```

### Builder 的运行流程

1. ANALYZE：分析项目结构（Go、PostgreSQL、REST API）
2. PLAN：决定 3 人团队（API Designer、DB Specialist、Test Engineer）
3. GENERATE：生成各智能体定义与 manifest.json
4. ACTIVATE：激活团队并注册 `/harness:backend-team` 命令

### 生成结果的位置

- 智能体定义：`.claude/agents/harness/api-designer.md`、`db-specialist.md`、...
- Manifest：`.moai/harness/manifest.json`
- 可选 worktree：`~/.moai/worktrees/<project>/`（用户 opt-in 时）

## Worktree 隔离（可选）

Harness v4 支持按条件的 worktree 隔离。

### L1 隔离 (Optional)

Claude Code 运行时会为每个智能体创建 L1 worktree。

- **使用时机**：并行成员编辑同一文件时
- **隔离范围**：每名成员的文件写入发生在独立的 worktree 中
- **成本**：额外内存 + 抵消部分并行收益

### 禁用

在 manifest 中设置 `"worktree_isolation": "none"` 即可跳过 L1 隔离。

## 相关文档

- [Harness v4 Builder 深入指南](/zh/advanced/harness-v4-builder) - Builder 4-phase 详解与 manifest 模式
- [智能体指南](/zh/advanced/agent-guide) - 11 个核心智能体目录
- [动态工作流](/zh/advanced/ultracode-workflows) - `/effort ultracode` 并行执行

{{< callout type="info" >}}
**提示**：用 Harness v4 Builder 为每个项目 **只需生成一次定制团队**，之后所有任务都会自动委派给该团队。首次生成后，随时可用 `/harness:team-name` 重复使用。
{{< /callout >}}
