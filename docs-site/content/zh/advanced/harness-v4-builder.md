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
| **Specialist 数量** | 项目复杂度 × 所需专业能力（HARD 上限 3~7） | 3 个 specialist |
| **执行原语（primitive）** | 各 specialist 的执行形态 | sub-agent、adversarial-fan-out |
| **隔离（isolation）** | 并行 specialist 冲突可能性 | none \| worktree |
| **模型·effort 分配** | 各 specialist 的推理复杂度（目的驱动） | content-author: opus/high, translator: sonnet/medium |
| **companion 技能** | specialist 专业能力所需技能 | hns-oss-docs-i18n-rules |

按 specialist 选择模型·effort 是代币经济学的核心 — 把需要深度推理的写作交给高阶模型·high effort，把重复性的派生工作交给便宜的模型·medium effort。用户审批门禁在 PLAN→GENERATE 边界通过 `AskUserQuestion` 进行。

#### 计划确认

生成前会向用户确认。没有经过审批门禁，文件绝不会被创建。

```
计划的 Harness 组成:
- 名称: backend-team
- specialist 3 个:
  ① architect (primitive: sub-agent, model: opus, effort: high)
  ② implementer (primitive: sub-agent, model: inherit, effort: high)
  ③ tester (primitive: sub-agent, model: sonnet, effort: medium)
- entry 命令: /harness:backend-team

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
3. **命令注册**：启用 `/harness:backend-team` entry 命令
4. **Runner 初始化**：准备启动基于 Manifest 的 Runner
5. **Worktree 创建**（可选）：设置 specialist 隔离的启用条件

#### 激活确认

```bash
moai harness list
# 显示 backend-team（名称 + 领域 + entry 命令）

moai harness doctor
# 引用完整性冒烟门禁（specialist·skill·workflow 引用校验）
```

## Manifest 模式

### 顶层字段

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `name` | string | 是 | Harness 名称（用于 entry 命令） |
| `domain` | string | 是 | Harness 领域说明 |
| `patterns` | array | 是 | 执行模式（`Pipeline`、`Fan-out/Fan-in`、`Producer-Reviewer` 等） |
| `specialists` | array | 是 | Specialist 对象数组（3~7 个 HARD 上限） |
| `sprint_contract` | object | 是 | 质量维度·阈值·must_pass 门禁 |
| `companion_skills` | array | — | Harness 专用 companion 技能列表 |
| `entry_command` | string | 是 | `/harness:<name>` entry 命令 |
| `runner_workflow` | string | 是 | Runner 工作流脚本文件 |
| `schedule` | object | — | （可选）重复执行调度 — `mode: discovery-only` 等 |

### Specialist 对象

```json
{
  "role": "content-author",
  "description": "canonical-locale 原文写作",
  "agent_file": ".claude/agents/harness/hns-oss-docs-content-author-specialist.md",
  "primitive": "sub-agent",
  "isolation": "none",
  "effort": "high",
  "model": "opus"
}
```

| 字段 | 说明 |
|------|------|
| `role` | specialist 角色（连字符/英文） |
| `description` | 角色说明（自由文本） |
| `agent_file` | specialist 智能体文件路径（`.claude/agents/harness/`） |
| `primitive` | 执行原语（`sub-agent`、`adversarial-fan-out` 等） |
| `isolation` | 隔离级别（`none`、`worktree`） |
| `effort` | 推理强度（`low`、`medium`、`high`、`xhigh`）— 目的驱动 |
| `model` | 模型档位（`opus`、`sonnet`、`haiku`、`inherit`）— 目的驱动 |

### Sprint Contract

```json
{
  "dimensions": ["locale-parity", "build-clean", "style-compliance", "content-fidelity"],
  "thresholds": { "locale-parity": 1.0, "build-clean": 1.0, "style-compliance": 0.95 },
  "must_pass": ["locale-parity", "build-clean"]
}
```

`dimensions` 是评分维度，`thresholds` 是各维度的通过阈值，`must_pass` 定义必须通过的门禁。

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
| `isolation: "worktree"` | 对 specialist 应用 worktree 隔离 |
| `isolation: "none"` | 禁用隔离 |
| `model: "inherit"` | 继承父会话模型 |
| `model: "sonnet"` | 派生/重复工作的低成本档位 |
| `effort: "high"` \| `"medium"` | 各 specialist 的推理强度（目的驱动） |
| `companion_skills: ["..."]` | Harness 专用 companion 技能 |

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
**提示**：Manifest 生成后可随时用 `moai harness edit <name>` 确认编辑路径并修改。添加 specialist、更换技能、调整隔离策略均可。
{{< /callout >}}
