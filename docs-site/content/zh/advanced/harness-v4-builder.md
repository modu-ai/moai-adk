---
title: Harness v4 Builder 深掘指南
weight: 45
draft: false
---

详细指导 Harness v4 Builder 的 4 阶段工作流、Manifest 架构和 Runner 原语。

{{< callout type="info" >}}
**一行总结**: Harness v4 Builder 通过 Socratic 访谈识别所需专业性，并通过基于 manifest 的 Runner 操作动态团队。
{{< /callout >}}

## 4-Phase Workflow 详解

### Phase 1: ANALYZE（分析）

分析当前项目的技术栈和要求。

#### 分析对象

- **项目结构**: 目录层次、核心包识别
- **使用语言**: Go、Python、TypeScript、Java 等检测
- **框架**: REST API、gRPC、FastAPI、Django 等识别
- **现有代理**: `.claude/agents/` 现有定义目录
- **项目规模**: 基于文件数、代码行数估算
- **依赖**: `go.mod`、`package.json`、`pyproject.toml` 分析

#### 产出

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

### Phase 2: PLAN（规划）

基于 ANALYZE 结果设计团队构成。

#### 规划决策

| 项目 | 决策方式 | 示例 |
|------|--------|------|
| **团队规模** | 项目复杂度 × 所需专业性 | 3~5 人 |
| **角色档案** | Anthropic role_profiles（researcher/architect/implementer/tester/designer/reviewer） | architect、implementer、tester |
| **Worktree 隔离** | 并行团队成员冲突可能性 | L1_optional（可选隔离） |
| **模型选择** | 按角色推理复杂度 | architect: inherit、tester: haiku |
| **技能预加载** | 角色专业性所需技能 | moai-foundation-core、moai-domain-backend |

#### 规划验证

生成前向用户确认:

```
计划的团队构成:
- 团队名: Backend Development Team
- 3 名团队成员:
  ① architect (model: inherit)
  ② implementer (model: inherit)
  ③ tester (model: haiku)
- Worktree 隔离: L1_optional
- Manifest: .moai/harness/manifest.json

这个构成继续吗?
```

### Phase 3: GENERATE（生成）

PLAN 批准后生成实际代理文件和 manifest。

#### 生成产出

**1. 代理定义文件**

```
.claude/agents/harness/
├── architect.md
├── implementer.md
└── tester.md
```

每个文件通过 YAML 提示定义:

```yaml
---
name: architect
description: API 架构设计专家
tools: Read, Write, Edit, Grep, Glob, Bash
model: inherit
---

你是这个项目的 API 架构专家。
[按角色详细指导]
```

**2. Manifest 文件**

```
.moai/harness/manifest.json
```

包含 Phase 和 Teammate 定义的 JSON（架构见 § Manifest 架构）。

#### 生成验证

```bash
ls .claude/agents/harness/
# 确认 architect.md、implementer.md、tester.md

ls .moai/harness/
# 确认 manifest.json

grep -c "\"name\": \"architect\"" .moai/harness/manifest.json
# 验证 phase 定义是否准确
```

### Phase 4: ACTIVATE（激活）

注册生成的工具并使其立即可用。

#### 激活步骤

1. **代理验证**: 各代理文件语法检查
2. **Manifest 验证**: JSON 架构和字段验证
3. **命令注册**: `/harness:backend-team` 命令激活
4. **Runner 初始化**: Manifest 基础 Runner 启动准备
5. **Worktree 生成**（可选）: L1 隔离激活条件设置

#### 激活确认

```bash
/harness list
# 显示 backend-team

/harness:backend-team status
# 确认 3 名团队成员、模型、状态
```

## Manifest 架构

### 顶级字段

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `spec_id` | string | 是 | `HARNESS-{DOMAIN}-{NUM}` 格式 |
| `name` | string | 是 | 团队显示名称 |
| `version` | string | 是 | 语义版本化 `X.Y.Z` |
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
| `name` | 必需 | 团队成员 ID（使用连字符，无空格） |
| `role` | 必需 | 角色说明（自由文本） |
| `model` | `inherit` | `inherit`、`haiku`、`sonnet`、`opus` |
| `mode` | `acceptEdits` | 权限模式（`acceptEdits`、`default`、`bypassPermissions`） |
| `skills` | `[]` | 预加载技能数组（例: `["moai-foundation-core"]`） |
| `isolation` | 无 | `worktree_optional`（worktree 隔离条件激活） |

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
      "description": "架构设计和 SPEC 编写",
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
          "role": "DB 设计和迁移",
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
          "role": "单元测试和集成测试",
          "model": "haiku",
          "mode": "acceptEdits"
        }
      ]
    }
  ]
}
```

## Runner 原语

基于 Manifest 的 Runner 执行生成的团队。

### Runner 生命周期

```
Team Spawn
  ↓
[Phase 1: plan]
  → Teammate(architect) 生成和委托
  → 收集结果
  ↓
[Phase 2: run]
  → Teammate(db-engineer) 并行生成
  → Teammate(api-developer) 并行生成
  → Teammate(test-engineer) 顺序生成
  → 收集和集成结果
  ↓
[Phase 3: sync]
  → 运行基础 manager-docs
  ↓
Team Teardown
```

### Runner 配置

Runner 的行为由 manifest 的字段控制:

| 配置 | 意义 |
|------|------|
| `worktree_isolation: "L1_optional"` | 冲突检测时自动应用隔离 |
| `worktree_isolation: "none"` | 禁用隔离 |
| `model: "inherit"` | 继承父会话模型 |
| `model: "haiku"` | 强制 Haiku 模型（成本优化） |
| `skills: ["..."]` | 预加载技能 |

## Worktree 隔离规则

### L1_optional 行为

```
Runner 生成时:
├── 团队成员 1: 主项目根
├── 团队成员 2: 主项目根
└── 冲突检测时
    ├── 团队成员 2 → 切换到 L1 worktree
    └── 团队成员 1 保持主（或也切换）

结果:
└── 文件冲突避免 ✓
```

### 隔离条件

满足以下任一条件时激活隔离:

1. **同一文件并行编辑**: 两个团队成员同时修改同一文件
2. **递归目录写入**: 团队成员在同一目录创建多个文件
3. **依赖冲突**: 团队成员 A 的输出是团队成员 B 的输入（顺序重要）

### 非隔离（none）选择时

```
所有团队成员在主项目中工作
优点: 最小内存、快速并行
缺点: 冲突可能性
```

## 相关文档

- [Harness v4 Builder 使用指南](/workflow-commands/moai-harness) - 命令参考
- [代理指南](/advanced/agent-guide) - 代理定义格式
- [基于 SPEC 的开发](/workflow-commands/moai-plan) - Harness 和 SPEC 集成

{{< callout type="info" >}}
**提示**: Manifest 生成后可以随时通过 `/harness:team-name edit` 进行修改。添加团队成员、更改技能、调整隔离策略都可以。
{{< /callout >}}
