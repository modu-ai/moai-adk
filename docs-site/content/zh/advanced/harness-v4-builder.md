---
title: Harness v4 Builder 深入指南
weight: 45
draft: false
---

如果说 [构建器智能体指南](/zh/advanced/builder-agents) 是 Harness v4 Builder 的概览，那么本文就是设计图 — 涵盖 4-phase 工作流各阶段的产出物、完整的 Manifest 模式，以及 Runner 原语的运行规则。

{{< callout type="info" >}}
**一句话总结**：Harness v4 Builder 通过苏格拉底式访谈把握所需的专业能力，并用基于 manifest 的 Runner 运营动态团队。哪个成员用什么模型工作，不是由代码而是由 manifest 声明决定的。
{{< /callout >}}

## 4-Phase Workflow 详解

### Phase 1: ANALYZE（分析）

分析当前项目的技术栈与需求。这一阶段的目标是用数据回答"这个项目缺少哪些专业能力"。

#### 分析对象

- **项目结构**：目录层级、识别核心包
- **所用语言**：检测 Go、Python、TypeScript、Java 等
- **框架**：识别 REST API、gRPC、FastAPI、Django 等
- **既有智能体**：`.claude/agents/` 中的既有定义目录
- **项目规模**：基于文件数、代码行数估算
- **依赖**：分析 `go.mod`、`package.json`、`pyproject.toml`

#### 产出物

```yaml
analysis_result:
  languages:
    - go (primary)
    - shell (build scripts)
  frameworks:
    - REST API (net/http)
    - PostgreSQL ORM (sqlc)
  scale: "100~300 files, ~50K LOC"
  existing_agents: 0
  expertise_gaps:
    - Database schema design
    - API error handling patterns
    - Test coverage automation
```

### Phase 2: PLAN（计划）

基于 ANALYZE 的结果设计团队组成。从团队规模到按角色的模型分配，所有影响成本的决定都在此阶段做出。

#### 计划决策事项

| 项目 | 决定方式 | 示例 |
|------|---------|------|
| **团队规模** | 项目复杂度 × 所需专业能力 | 3~5 名 |
| **角色档案** | Anthropic role_profiles（researcher/architect/implementer/tester/designer/reviewer） | architect, implementer, tester |
| **Worktree 隔离** | 并行成员冲突可能性 | L1_optional（可选隔离） |
| **模型选择** | 按角色的推理复杂度 | architect: inherit, tester: haiku |
| **技能预加载** | 角色专业能力所需技能 | moai-foundation-core, moai-domain-backend |

按角色选择模型是代币经济学的核心 — 把需要深度推理的设计交给强模型，把重复性的测试编写交给便宜的模型。

#### 计划确认

生成前会向用户确认。没有经过审批门禁，文件绝不会被创建。

```
计划的团队组成:
- 团队名: Backend Development Team
- 成员 3 名:
  ① architect (model: inherit)
  ② implementer (model: inherit)
  ③ tester (model: haiku)
- Worktree 隔离: L1_optional
- Manifest: .moai/harness/manifest.json

以这个组成继续吗?
```

### Phase 3: GENERATE（生成）

PLAN 获批后，生成实际的智能体文件与 manifest。

#### 生成产物

**1. 智能体定义文件**

```
.claude/agents/harness/
├── architect.md
├── implementer.md
└── tester.md
```

每个文件以 YAML 提示词定义。

```yaml
---
name: architect
description: API 架构设计专家
tools: Read, Write, Edit, Grep, Glob, Bash
model: inherit
---

你是本项目的 API 架构专家。
[按角色的详细指令]
```

**2. Manifest 文件**

```
.moai/harness/manifest.json
```

包含 Phase 与 Teammate 定义的 JSON（模式见 § Manifest 模式）。

#### 生成校验

生成后可以立即直接确认文件存在与定义正确性。

```bash
ls .claude/agents/harness/
# 确认 architect.md, implementer.md, tester.md

ls .moai/harness/
# 确认 manifest.json

grep -c "\"name\": \"architect\"" .moai/harness/manifest.json
# 确认 phase 定义是否正确
```

### Phase 4: ACTIVATE（激活）

注册生成的 Harness 并使其立即可用。

#### 激活步骤

1. **智能体校验**：检查各智能体文件的语法
2. **Manifest 校验**：JSON 模式与字段校验
3. **命令注册**：启用 `/harness:backend-team` 命令
4. **Runner 初始化**：准备启动基于 Manifest 的 Runner
5. **Worktree 创建**（可选）：设置 L1 隔离的启用条件

#### 激活确认

```bash
/harness list
# 显示 backend-team

/harness:backend-team status
# 确认 3 名成员、模型、状态
```

## Manifest 模式

### 顶层字段

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `spec_id` | string | 是 | `HARNESS-{DOMAIN}-{NUM}` 格式 |
| `name` | string | 是 | 团队显示名称 |
| `version` | string | 是 | 语义化版本 `X.Y.Z` |
| `created_at` | string | 是 | ISO 8601 时间戳 |
| `worktree_isolation` | enum | 是 | `L1_optional` \| `none` |
| `phases` | array | 是 | Phase 对象数组 |

### Phase 对象

```json
{
  "name": "run",
  "description": "实现阶段",
  "teammates": [...]
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | `plan` \| `run` \| `sync` |
| `description` | string | Phase 目标说明 |
| `teammates` | array | Teammate 对象数组 |

### Teammate 对象

```json
{
  "name": "api-developer",
  "role": "REST API 端点开发",
  "model": "inherit",
  "mode": "acceptEdits",
  "skills": ["moai-foundation-core"],
  "isolation": "worktree_optional"
}
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `name` | 必需 | 成员 ID（使用连字符，不含空格） |
| `role` | 必需 | 角色描述（自由文本） |
| `model` | `inherit` | `inherit`、`haiku`、`sonnet`、`opus` |
| `mode` | `acceptEdits` | 权限模式（`acceptEdits`、`default`、`bypassPermissions`） |
| `skills` | `[]` | 预加载技能数组（例：`["moai-foundation-core"]`） |
| `isolation` | 无 | `worktree_optional`（按条件启用 worktree 隔离） |

### 完整示例

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
      "description": "架构设计与 SPEC 编写",
      "teammates": [
        {
          "name": "architect",
          "role": "API 架构专家",
          "model": "inherit",
          "mode": "acceptEdits",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "description": "实际实现",
      "teammates": [
        {
          "name": "db-engineer",
          "role": "DB 设计与迁移",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        },
        {
          "name": "api-developer",
          "role": "REST API 端点实现",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        },
        {
          "name": "test-engineer",
          "role": "单元测试与集成测试",
          "model": "haiku",
          "mode": "acceptEdits"
        }
      ]
    }
  ]
}
```

## Runner 原语

基于 Manifest 的 Runner 负责执行生成的团队。

### Runner 生命周期

```
Team Spawn
  ↓
[Phase 1: plan]
  → 生成并委派 Teammate(architect)
  → 收集结果
  ↓
[Phase 2: run]
  → 并行生成 Teammate(db-engineer)
  → 并行生成 Teammate(api-developer)
  → 顺序生成 Teammate(test-engineer)
  → 收集并整合结果
  ↓
[Phase 3: sync]
  → 执行默认 manager-docs
  ↓
Team Teardown
```

### Runner 配置

Runner 的行为由 manifest 的字段控制。

| 配置 | 含义 |
|------|------|
| `worktree_isolation: "L1_optional"` | 检测到冲突时自动应用隔离 |
| `worktree_isolation: "none"` | 禁用隔离 |
| `model: "inherit"` | 继承父会话模型 |
| `model: "haiku"` | 强制 Haiku 模型（成本最优） |
| `skills: ["..."]` | 预加载技能 |

## Worktree 隔离规则

### L1_optional 的运行

```
Runner 创建时:
├── 成员 1: 主项目根目录
├── 成员 2: 主项目根目录
└── 检测到冲突时
    ├── 成员 2 → 切换到 L1 worktree
    └── 成员 1 保持在主目录 (或成员 1 也切换)

结果:
└── 规避文件冲突 ✓
```

### 隔离条件

以下任一为真时启用隔离。

1. **同一文件并行编辑**：两名成员同时修改同一文件
2. **递归目录写入**：多名成员在同一目录生成多个文件
3. **依赖竞争**：成员 A 的输出是成员 B 的输入（顺序重要）

### 选择不隔离 (none) 时

```
所有成员在主项目中工作
优点: 内存最小、并行更快
缺点: 存在冲突可能
```

## 相关文档

- [Harness v4 Builder 使用指南](/zh/workflow-commands/moai-harness) - 命令参考
- [智能体指南](/zh/advanced/agent-guide) - 智能体定义格式
- [基于 SPEC 的开发](/zh/workflow-commands/moai-plan) - Harness 与 SPEC 集成

{{< callout type="info" >}}
**提示**：Manifest 生成后可随时用 `/harness:team-name edit` 修改。添加成员、更换技能、调整隔离策略均可。
{{< /callout >}}
