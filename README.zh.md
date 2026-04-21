<p align="center">
  <img src="./assets/images/moai-adk-og.png" alt="MoAI-ADK" width="100%">
</p>

<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>Claude Code 的 Agentic Development Kit</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <a href="./README.ja.md">日本語</a> ·
  <a href="./README.zh.md">中文</a>
</p>

<p align="center">
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
  <a href="https://codecov.io/gh/modu-ai/moai-adk"><img src="https://codecov.io/gh/modu-ai/moai-adk/branch/main/graph/badge.svg" alt="Codecov"></a>
  <br>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/github/v/release/modu-ai/moai-adk?sort=semver" alt="Release"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
</p>

<p align="center">
  <a href="https://adk.mo.ai.kr"><strong>官方文档</strong></a>
</p>

---

> 📚 **[官方文档](https://adk.mo.ai.kr)**

---

> **"氛围编程的目的不是追求速度，而是代码质量。"**

MoAI-ADK 是专为 Claude Code 打造的**高性能 AI 开发环境**。26 个专业 AI 智能体与 47 个技能协同工作，助力产出高质量代码。新项目和功能开发默认采用 TDD，覆盖率低于 10% 的现有项目自动采用 DDD，并支持 Sub-Agent 与 Agent Teams 双执行模式。

使用 Go 编写的单一可执行文件 -- 零依赖，全平台即刻运行。

---

## v2.12.0 新特性

MoAI-ADK v2.12.0 引入了设计系统、Claude Code 原生集成和 Opus 4.7 支持的重大升级。

### 主要里程碑

| 版本 | 亮点 |
|------|------|
| **v2.9.0** | Claude Code v2.1.89-90 原生技能集成 (Opus 4.6) |
| **v2.10.x** | LSP 套件扩展、SPEC-CC297-001 `permissionMode` 属性支持、Opus 4.7 预览 |
| **v2.11.x** | 自学习系统集成、多源文档加载、增强的内存管理 |
| **v2.12.0** | **[SPEC-AGENCY-ABSORB-001]** /agency → /moai design 吸收、Opus 4.7 完整支持、Adaptive Thinking 原生集成 |

### 主要变化

**设计系统吸收 (SPEC-AGENCY-ABSORB-001)**

旧有的 `/agency` 命令已完全融入 `/moai design`。现有的 `/agency/` 项目可通过以下命令自动迁移：

```bash
moai migrate agency
```

优势：
- 单一统一的设计工作流，而非双重的 `/moai` + `/agency` 命令
- 与 MoAI 核心的改进集成（品牌上下文、质量门、SPEC 驱动工作流）
- 增强的文档 [adk.mo.ai.kr](https://adk.mo.ai.kr)

**Opus 4.7 原生支持**

MoAI-ADK 现已支持 Claude Opus 4.7 的原生 Adaptive Thinking：

- 推理的自动动态令牌分配（无固定预算）
- 通过简洁提示表述实现更快推理
- 复杂任务的更优成本效益

**自学习和内存演进**

v2.11+ 自学习系统现已与智能体学习集成：

- 智能体自动从更正中捕获教训
- 内存系统跨会话持久化 (`.claude/agent-memory/`)
- 文档根据任务上下文实时加载

---

## 为什么选择 MoAI-ADK？

我们将基于 Python 的 MoAI-ADK（约 73,000 行）用 Go 完全重写。

| 项目 | Python 版 | Go 版 |
|------|-----------|-------|
| 部署 | pip + venv + 依赖管理 | **单一可执行文件**，零依赖 |
| 启动时间 | ~800ms 解释器启动 | **~5ms** 原生执行 |
| 并发性 | asyncio / threading | **原生 goroutines** |
| 类型安全 | 运行时（mypy 可选） | **编译时强制** |
| 跨平台 | 需要 Python 运行时 | **预构建二进制**（macOS、Linux、Windows） |
| 钩子执行 | Shell 包装器 + Python | **编译后的二进制**，JSON 协议 |

### 核心数据

- **38,700+ 行** Go 代码，**38 个**包
- **85-100%** 测试覆盖率
- **26 个**专业 AI 智能体 + **47 个**技能
- **18 种**编程语言支持
- **27 个** Claude Code 钩子事件

---

## 硬件工程架构

MoAI-ADK 实现了**硬件工程（Harness Engineering）**范式 — 不直接编写代码，而是为 AI 智能体设计运行环境。

| 组件 | 说明 | 命令 |
|------|------|------|
| **自我验证循环** | 智能体自主执行代码编写→测试→失败→修复→通过的循环 | `/moai loop` |
| **上下文地图** | 代码库架构图和文档始终可供智能体参考 | `/moai codemaps` |
| **会话持久化** | `progress.md` 追踪已完成阶段，中断的执行自动恢复 | `/moai run SPEC-XXX` |
| **失败清单** | 所有验收标准在运行开始时注册为 pending 任务，实现后标记为 completed | `/moai run SPEC-XXX` |
| **语言无关** | 支持 16 种语言：自动检测语言，选择正确的 LSP/代码检查/测试/覆盖率工具 | 全部工作流 |
| **垃圾回收** | 定期扫描并清除死代码、AI Slop 和未使用的导入 | `/moai clean` |
| **脚手架优先** | 在实现前创建空文件桩，防止代码熵增 | `/moai run SPEC-XXX` |

> "人类负责把舵（Steer），智能体负责执行（Execute）。" — 工程师的角色从编写代码转变为设计硬件：SPEC、质量门禁和反馈循环。

---

## 系统要求

| 平台 | 支持环境 | 备注 |
|------|---------|------|
| macOS | Terminal, iTerm2 | 完全支持 |
| Linux | Bash, Zsh | 完全支持 |
| Windows | **WSL（推荐）**, PowerShell 7.x+ | 原生 cmd.exe 不支持 |

**前置条件：**
- 所有平台必须安装 **Git**
- **Windows 用户**：必须安装 [Git for Windows](https://gitforwindows.org/)（包含 Git Bash）
  - 推荐使用 **WSL**（适用于 Linux 的 Windows 子系统）
  - 也支持 PowerShell 7.x 及以上版本
  - 旧版 Windows PowerShell 5.x 和 cmd.exe **不受支持**

---

## 快速入门

### 1. 安装

#### macOS / Linux / WSL

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

#### Windows（PowerShell 7.x+）

> **推荐**：为获得最佳体验，请在 WSL 中使用上面的 Linux 安装命令。

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

> 需要先安装 [Git for Windows](https://gitforwindows.org/)。

#### 从源码构建（Go 1.26+）

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk && make build
```

> 预构建二进制可在 [Releases](https://github.com/modu-ai/moai-adk/releases) 页面下载。

### 2. Windows 特定问题

#### 韩文用户名路径错误

如果 Windows 用户名包含非 ASCII 字符（韩文、中文等），
可能会因 Windows 8.3 短文件名转换而遇到 `EINVAL` 错误。

**解决方案 1：** 设置替代临时目录：

```bash
# 命令提示符
set MOAI_TEMP_DIR=C:\temp
mkdir C:\temp 2>nul

# PowerShell
$env:MOAI_TEMP_DIR="C:\temp"
New-Item -ItemType Directory -Path "C:\temp" -Force
```

**解决方案 2：** 禁用 8.3 文件名生成（需要管理员权限）：

```bash
fsutil 8dot3name set 1
```

**解决方案 3：** 使用仅包含 ASCII 字符的用户名创建新 Windows 用户帐户。

### 3. 初始化项目

```bash
moai init my-project
```

交互式向导将自动检测语言、框架和方法论，并生成 Claude Code 集成文件。

### 4. 在 Claude Code 中开始开发

```bash
# 启动 Claude Code 后
/moai project                            # 生成项目文档（product.md, structure.md, tech.md）
/moai plan "添加用户认证功能"              # 生成 SPEC 文档
/moai run SPEC-AUTH-001                   # DDD/TDD 实现
/moai sync SPEC-AUTH-001                  # 文档同步 & 创建 PR
/moai github issues                      # GitHub issue自动化 (Agent Teams)
/moai github pr 123                       # PR多角度审查 (multi-perspective)
```

```mermaid
graph LR
    A["🔍 /moai project"] --> B["📋 /moai plan"]
    B -->|"SPEC 文档"| C["🔨 /moai run"]
    C -->|"实现完成"| D["📄 /moai sync"]
    D -->|"创建 PR"| E["✅ Done"]
```

---

## MoAI 开发方法论

MoAI-ADK 根据项目状态自动选择最优的开发方法论。

```mermaid
flowchart TD
    A["🔍 项目分析"] --> B{"新项目或<br/>10%+ 测试覆盖率？"}
    B -->|"Yes"| C["TDD（默认）"]
    B -->|"No"| D{"现有项目<br/>< 10% 覆盖率？"}
    D -->|"Yes"| E["DDD"]
    C --> F["RED → GREEN → REFACTOR"]
    E --> G["ANALYZE → PRESERVE → IMPROVE"]

    style C fill:#4CAF50,color:#fff
    style E fill:#2196F3,color:#fff
```

### TDD 方法论（默认）

新项目和功能开发推荐的默认方法论。先写测试，再实现代码。

| 阶段 | 说明 |
|------|------|
| **RED** | 编写定义期望行为的失败测试 |
| **GREEN** | 编写使测试通过的最小代码 |
| **REFACTOR** | 在保持测试绿色的同时提升代码质量。REFACTOR 完成后 `/simplify` 自动运行。 |

对于棕地项目（现有代码库），TDD 增加了 **RED 前分析步骤**：在编写测试前先阅读现有代码，了解当前行为。

### DDD 方法论（覆盖率 < 10% 的现有项目）

专为测试覆盖率较低的现有项目安全重构而设计的方法论。

```
ANALYZE   → 分析现有代码和依赖关系，识别领域边界
PRESERVE  → 编写特征测试，捕获当前行为快照
IMPROVE   → 在测试保护下渐进改进。IMPROVE 完成后 /simplify 自动运行。
```

> 方法论在 `moai init` 时自动选定（`--mode <ddd|tdd>`，默认: tdd），可在 `.moai/config/sections/quality.yaml` 的 `development_mode` 中修改。
>
> **注意**：MoAI-ADK v2.5.0+ 采用二元方法论选择（仅 TDD 或 DDD）。混合模式已移除，以确保清晰性和一致性。

### 自动质量 & 横向扩展层

MoAI-ADK v2.6.0+ 集成了两个 MoAI **自主**调用的 Claude Code 原生技能 — 无需任何标志或手动命令。

| 技能 | 作用 | 触发时机 |
|------|------|---------|
| `/simplify` | 质量强化 | TDD REFACTOR 和 DDD IMPROVE 阶段完成后**始终**运行 |
| `/batch` | 横向扩展执行 | 任务复杂度超过阈值时自动触发 |

**`/simplify` — 自动质量通道**

使用并行智能体从复用机会、质量问题、效率以及 CLAUDE.md 合规性等维度审查变更代码，并自动修复发现的问题。无需配置，每个实现周期后由 MoAI 直接调用。

**`/batch` — 并行横向扩展**

在隔离的 git worktree 中启动数十个智能体，处理大规模并行工作。每个智能体运行测试并报告结果，MoAI 负责合并。各工作流的自动触发条件：

| 工作流 | 触发条件 |
|--------|---------|
| `run` | 任务数 ≥ 5，或预测文件变更数 ≥ 10，或独立任务数 ≥ 3 |
| `mx` | 源文件数 ≥ 50 |
| `coverage` | P1+P2 覆盖率缺口 ≥ 10 |
| `clean` | 确认的死代码项目 ≥ 20 |

---

## AI 智能体编排

MoAI 是一个**战略编排器**。它不直接编写代码，而是将任务委托给 27 个专业智能体。

```mermaid
graph LR
    U["👤 用户请求"] --> M["🗿 MoAI Orchestrator"]

    M --> MG["📋 Manager (8)"]
    M --> EX["⚡ Expert (8)"]
    M --> BL["🔧 Builder (3)"]
    M --> EV["🔍 Evaluator (2)"]
    M --> AG["🎨 Agency (6)"]

    MG --> MG1["spec · ddd · tdd · docs<br/>quality · project · strategy · git"]
    EX --> EX1["backend · frontend · security · devops<br/>performance · debug · testing · refactoring"]
    BL --> BL1["agent · skill · plugin"]
    EV --> EV1["evaluator-active · plan-auditor"]
    AG --> AG1["planner · copywriter · designer<br/>builder · evaluator · learner"]

    style M fill:#FF6B35,color:#fff
    style MG fill:#4CAF50,color:#fff
    style EX fill:#2196F3,color:#fff
    style BL fill:#9C27B0,color:#fff
    style EV fill:#FF5722,color:#fff
    style AG fill:#FF9800,color:#fff
```

### 智能体分类

| 分类 | 数量 | 智能体 | 职责 |
|------|------|--------|------|
| **Manager** | 8 | spec, ddd, tdd, docs, quality, project, strategy, git | 工作流编排、SPEC 生成、质量管理 |
| **Expert** | 8 | backend, frontend, security, devops, performance, debug, testing, refactoring | 领域专业实现、分析、优化 |
| **Builder** | 3 | agent, skill, plugin | 创建新的 MoAI 组件 |
| **Evaluator** | 2 | evaluator-active, plan-auditor | 独立质量评估、计划阶段文档审计 |
| **Agency** | 6 | planner, copywriter, designer, builder, evaluator, learner | 创意生产流水线 |

### 47 个技能（渐进式披露）

为优化 Token 效率，采用三级渐进式披露系统管理：

| 分类 | 技能数 | 示例 |
|------|--------|------|
| **Foundation** | 5 | core, claude, philosopher, quality, context |
| **Workflow** | 11 | spec, project, ddd, tdd, testing, worktree, thinking... |
| **Domain** | 5 | backend, frontend, database, uiux, data-formats |
| **Language** | 18 | Go, Python, TypeScript, Rust, Java, Kotlin, Swift, C++... |
| **Platform** | 9 | Vercel, Supabase, Firebase, Auth0, Clerk, Railway... |
| **Library** | 3 | shadcn, nextra, mermaid |
| **Tool** | 2 | ast-grep, svg |
| **Specialist** | 10 | Figma, Flutter, Electron, Pencil... |

---

## 模型策略（Token 优化）

MoAI-ADK 根据您的 Claude Code 订阅方案，为 26 个智能体分配最优 AI 模型。在方案的速率限制内最大化质量。

| 策略 | 方案 | 🟣 Opus | 🔵 Sonnet | 🟡 Haiku | 适用场景 |
|------|------|------|--------|-------|----------|
| **High** | Max $200/月 | 22 | 1 | 4 | 最高质量，最大吞吐量 |
| **Medium** | Max $100/月 | 4 | 18 | 5 | 质量与成本平衡 |
| **Low** | Plus $20/月 | 0 | 11 | 16 | 经济实惠，无 Opus |

> **为什么重要？** Plus $20 方案不包含 Opus。设置为 `Low` 后，所有智能体仅使用 Sonnet 和 Haiku，避免速率限制错误。更高级方案可在关键智能体（安全、策略、架构）上使用 Opus，常规任务使用 Sonnet/Haiku。

### 各档位智能体模型分配

#### Manager Agents

| 智能体 | High | Medium | Low |
|--------|------|--------|-----|
| manager-spec | 🟣 opus | 🟣 opus | 🔵 sonnet |
| manager-strategy | 🟣 opus | 🟣 opus | 🔵 sonnet |
| manager-ddd | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| manager-tdd | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| manager-project | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| manager-docs | 🔵 sonnet | 🟡 haiku | 🟡 haiku |
| manager-quality | 🟡 haiku | 🟡 haiku | 🟡 haiku |
| manager-git | 🟡 haiku | 🟡 haiku | 🟡 haiku |

#### Expert Agents

| 智能体 | High | Medium | Low |
|--------|------|--------|-----|
| expert-backend | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| expert-frontend | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| expert-security | 🟣 opus | 🟣 opus | 🔵 sonnet |
| expert-debug | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| expert-refactoring | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| expert-devops | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| expert-performance | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| expert-testing | 🟣 opus | 🔵 sonnet | 🟡 haiku |

#### Builder Agents

| 智能体 | High | Medium | Low |
|--------|------|--------|-----|
| builder-agent | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| builder-skill | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| builder-plugin | 🟣 opus | 🔵 sonnet | 🟡 haiku |

#### Team Agents

| 智能体 | High | Medium | Low |
|--------|------|--------|-----|
| team-reader | 🟡 haiku | 🟡 haiku | 🟡 haiku |
| team-coder | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| team-tester | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| team-designer | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| team-validator | 🟡 haiku | 🟡 haiku | 🟡 haiku |

### 配置方法

```bash
# 项目初始化时
moai init my-project          # 交互式向导中选择模型策略

# 重新配置现有项目
moai update                   # 交互式提示每个配置步骤
```

在 `moai update` 期间，您会被询问：
- **重置模型策略？** (y/n) - 重新运行模型策略配置向导
- **更新 GLM 设置？** (y/n) - 在 settings.local.json 中配置 GLM 环境变量

> 默认策略为 `High`。GLM 设置隔离在 `settings.local.json`（不提交到 Git）。

---

## 双执行模式

MoAI-ADK 同时提供 Claude Code 支持的 **Sub-Agent** 和 **Agent Teams** 两种执行模式。

```mermaid
graph TD
    A["🗿 MoAI Orchestrator"] --> B{"选择执行模式"}
    B -->|"--solo"| C["Sub-Agent 模式"]
    B -->|"--team"| D["Agent Teams 模式"]
    B -->|"默认（自动）"| E["自动选择"]

    C --> F["顺序专家委托<br/>Task() → Expert Agent"]
    D --> G["并行团队协作<br/>TeamCreate → SendMessage"]
    E -->|"高复杂度"| D
    E -->|"低复杂度"| C

    style C fill:#2196F3,color:#fff
    style D fill:#FF9800,color:#fff
    style E fill:#4CAF50,color:#fff
```

### Agent Teams 模式（默认）

MoAI-ADK 自动分析项目复杂度并选择最优执行模式：

| 条件 | 选择模式 | 原因 |
|------|----------|------|
| 3+ 个领域 | Agent Teams | 多领域协调 |
| 10+ 个受影响文件 | Agent Teams | 大规模变更 |
| 复杂度评分 7+ | Agent Teams | 高度复杂 |
| 其他 | Sub-Agent | 简单、可预测的工作流 |

**Agent Teams 模式**采用并行团队开发：

- 多个智能体同时工作，通过共享任务列表协作
- 通过 `TeamCreate`、`SendMessage` 和 `TaskList` 实现实时协调
- 最适合大规模功能开发和多领域任务

```bash
/moai plan "大型功能"          # 自动：researcher + analyst + architect 并行
/moai run SPEC-XXX             # 自动：backend-dev + frontend-dev + tester 并行
/moai run SPEC-XXX --team      # 强制 Agent Teams 模式
```

**Agent Teams 质量钩子：**
- **TeammateIdle Hook**：在智能体空闲前验证 LSP 质量门禁（错误、类型错误、Lint 错误）
- **TaskCompleted Hook**：验证任务引用 SPEC-XXX 模式时 SPEC 文档存在
- 所有验证使用优雅降级 - 警告已记录但工作继续

### Sub-Agent 模式（`--solo`）

利用 Claude Code 的 `Task()` API 实现顺序智能体委托方式。

- 将任务委托给一个专业智能体并获取结果
- 按步骤依次执行 Manager → Expert → Quality
- 适用于简单且可预测的工作流

```bash
/moai run SPEC-AUTH-001 --solo    # 强制 Sub-Agent 模式
```

---

## MoAI 工作流

### Plan → Run → Sync 流水线

MoAI 的核心工作流由三个阶段组成：

```mermaid
graph TB
    subgraph Plan ["📋 Plan 阶段"]
        P1["代码库探索"] --> P2["需求分析"]
        P2 --> P3["生成 SPEC 文档（EARS 格式）"]
    end

    subgraph Run ["🔨 Run 阶段"]
        R1["SPEC 分析 & 执行计划"] --> R2["DDD/TDD 实现"]
        R2 --> R3["TRUST 5 质量验证"]
    end

    subgraph Sync ["📄 Sync 阶段"]
        S1["生成文档"] --> S2["更新 README/CHANGELOG"]
        S2 --> S3["创建 Pull Request"]
    end

    Plan --> Run
    Run --> Sync

    style Plan fill:#E3F2FD,stroke:#1565C0
    style Run fill:#E8F5E9,stroke:#2E7D32
    style Sync fill:#FFF3E0,stroke:#E65100
```

#### 执行模式选择门

从 Plan 阶段过渡到 Run 阶段时，MoAI 会自动检测当前执行环境（cc/glm/cg），并在实施开始前显示选择 UI，让用户确认或更改模式。

```mermaid
graph LR
    A["Plan 完成"] --> B["环境检测"]
    B --> C{"模式选择 UI"}
    C -->|"CC"| D["仅 Claude 执行"]
    C -->|"GLM"| E["仅 GLM 执行"]
    C -->|"CG"| F["Claude 领导者 + GLM 工作者"]
```

此门确保无论环境状态如何都使用正确的执行模式，防止实施过程中的模式不匹配。

### /moai 子命令

所有子命令在 Claude Code 中以 `/moai <subcommand>` 方式调用。

#### 核心工作流

| 子命令 | 别名 | 目的 | 关键标志 |
|--------|------|------|----------|
| `plan` | `spec` | 创建 SPEC 文档（EARS 格式） | `--worktree`, `--branch`, `--resume SPEC-XXX`, `--team` |
| `run` | `impl` | SPEC 的 DDD/TDD 实现 | `--resume SPEC-XXX`, `--team` |
| `sync` | `docs`, `pr` | 文档同步、代码地图和创建 PR | `--merge`, `--skip-mx` |

#### 质量与测试

| 子命令 | 别名 | 目的 | 关键标志 |
|--------|------|------|----------|
| `fix` | — | 自动修复 LSP 错误、Lint、类型错误（单次执行） | `--dry`, `--seq`, `--level N`, `--resume`, `--team` |
| `loop` | — | 迭代自动修复直至完成（最多 100 次） | `--max N`, `--auto-fix`, `--seq` |
| `review` | `code-review` | 代码审查与安全和 @MX 标签合规检查 | `--staged`, `--branch`, `--security` |
| `coverage` | `test-coverage` | 测试覆盖率分析和缺口补充（16 种语言） | `--target N`, `--file PATH`, `--report` |
| `e2e` | — | E2E 测试（Claude-in-Chrome、Playwright CLI 或 Agent Browser） | `--record`, `--url URL`, `--journey NAME` |
| `clean` | `refactor-clean` | 死代码识别和安全移除 | `--dry`, `--safe-only`, `--file PATH` |

#### 文档与代码库

| 子命令 | 别名 | 目的 | 关键标志 |
|--------|------|------|----------|
| `project` | `init` | 生成项目文档（product.md、structure.md、tech.md、.moai/project/codemaps/） | — |
| `mx` | — | 扫描代码库并添加 @MX 代码级注解 | `--all`, `--dry`, `--priority P1-P4`, `--force`, `--team` |
| `codemaps` | `update-codemaps` | 在 `.moai/project/codemaps/` 中生成架构文档 | `--force`, `--area AREA` |
| `feedback` | `fb`, `bug`, `issue` | 收集用户反馈并创建 GitHub issues | — |

#### 默认工作流

| 子命令 | 目的 | 关键标志 |
|--------|------|----------|
| *(无)* | 完整自主 plan → run → sync 流水线。复杂度评分 >= 5 时自动生成 SPEC。 | `--loop`, `--max N`, `--branch`, `--pr`, `--resume SPEC-XXX`, `--team`, `--solo` |

### 执行模式标志

控制工作流执行期间智能体的调度方式：

| 标志 | 模式 | 说明 |
|------|------|------|
| `--team` | Agent Teams | 并行团队执行。多个智能体同时工作。 |
| `--solo` | Sub-Agent | 顺序单智能体委托（每阶段）。 |
| *(默认)* | 自动 | 系统根据复杂度自动选择（域 >= 3、文件 >= 10 或评分 >= 7）。 |

**`--team` 支持三种执行环境：**

| 环境 | 命令 | 领导者 | 工作者 | 最适合 |
|------|------|--------|--------|--------|
| 仅 Claude | `moai cc` | Claude | Claude | 最高质量 |
| 仅 GLM | `moai glm` | GLM | GLM | 最大成本节省 |
| CG（Claude+GLM） | `moai cg` | Claude | GLM | 质量+成本平衡 |

> **v2.7.1 新增**：CG 模式现在是**默认**团队模式。使用 `--team` 时，除非通过 `moai cc` 或 `moai glm` 明确更改，否则系统默认以 CG 模式运行。

> **注意**：`moai cg` 使用 tmux pane 级别环境隔离来分离 Claude 领导者和 GLM 工作者。如果从 `moai glm` 切换，`moai cg` 会自动先重置 GLM 设置 -- 无需在中间运行 `moai cc`。

### 自主开发循环（Ralph Engine）

一个结合 LSP 诊断和 AST-grep 的自主错误修复引擎：

```bash
/moai fix       # 单次执行：扫描 → 分类 → 修复 → 验证
/moai loop      # 循环修复：重复执行直到检测到完成标记（最多 100 次）
```

**Ralph Engine 工作流程：**
1. **并行扫描**：同时运行 LSP 诊断 + AST-grep + Linters
2. **自动分类**：将错误分为 Level 1（自动修复）至 Level 4（需用户介入）
3. **收敛检测**：相同错误重复出现时启用备选策略
4. **完成条件**：0 错误、0 类型错误、85%+ 覆盖率

### 推荐工作流链

**新功能开发：**
```
/moai plan → /moai run SPEC-XXX → /moai review → /moai coverage → /moai sync SPEC-XXX
```

**Bug 修复：**
```
/moai fix (或 /moai loop) → /moai review → /moai sync
```

**重构：**
```
/moai plan → /moai clean → /moai run SPEC-XXX → /moai review → /moai coverage → /moai codemaps
```

**文档更新：**
```
/moai codemaps → /moai sync
```

---

## TRUST 5 质量框架

所有代码变更均通过五项质量标准验证：

| 标准 | 说明 | 验证项 |
|------|------|--------|
| **T**ested | 已测试 | 85%+ 覆盖率、特征测试、单元测试通过 |
| **R**eadable | 可读性 | 清晰的命名规范、一致的代码风格、0 Lint 错误 |
| **U**nified | 统一性 | 一致的格式化、导入顺序、遵循项目结构 |
| **S**ecured | 安全性 | OWASP 合规、输入验证、0 安全警告 |
| **T**rackable | 可追溯 | 约定式提交、Issue 引用、结构化日志 |

---

## 任务指标日志

MoAI-ADK 在开发会话期间自动捕获任务工具指标：

- **位置**：`.moai/logs/task-metrics.jsonl`
- **捕获指标**：Token 使用、工具调用、持续时间、智能体类型
- **目的**：会话分析、性能优化、成本跟踪

当 Task 工具完成时，指标由 PostToolUse 钩子记录。使用此数据分析智能体效率和优化 Token 消耗。

---

## CLI 命令

| 命令 | 说明 |
|------|------|
| `moai init` | 交互式项目设置（自动检测语言/框架/方法论） |
| `moai doctor` | 系统状态诊断与环境验证 |
| `moai status` | 项目状态概览：Git 分支、质量指标等 |
| `moai update` | 更新到最新版本（支持自动回滚） |
| `moai update --check` | 仅检查更新，不安装 |
| `moai update --project` | 仅同步项目模板 |
| `moai worktree new <name>` | 创建新 Git worktree（并行分支开发） |
| `moai worktree list` | 列出活跃的 worktree |
| `moai worktree switch <name>` | 切换 worktree |
| `moai worktree sync` | 与上游同步 |
| `moai worktree remove <name>` | 移除 worktree |
| `moai worktree clean` | 清理过期 worktree |
| `moai worktree go <name>` | 在当前 Shell 中导航到 worktree 目录 |
| `moai hook <event>` | Claude Code 钩子分发器 |
| `moai glm` | 使用 GLM 5 API 启动 Claude Code（经济高效的替代方案） |
| `moai cc` | 不使用 GLM 设置启动 Claude Code（仅 Claude 模式） |
| `moai cg` | 启动 CG 模式 — Claude 领导者 + GLM 工作者（自动启动 Claude Code，需要 tmux） |
| `moai version` | 版本、提交哈希、构建日期信息 |

---

## CG 模式（Claude + GLM 混合）

CG 模式是一种混合模式，领导者使用 **Claude API**，工作者使用 **GLM API**。通过 tmux 会话级环境变量隔离实现。

### 工作原理

```
moai cg 执行
    │
    ├── 1. 向 tmux 会话环境变量注入 GLM 配置
    │      (ANTHROPIC_AUTH_TOKEN, BASE_URL, MODEL_* 变量)
    │
    ├── 2. 从 settings.local.json 移除 GLM 环境变量
    │      → 领导者窗格使用 Claude API
    │
    ├── 3. 设置 CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
    │      → 工作者在新窗格中继承 GLM 环境变量
    │
    └── 4. 启动 Claude Code（替换当前进程）

┌─────────────────────────────────────────────────────────────┐
│  LEADER（当前 tmux 窗格，Claude API）                        │
│  - 执行 /moai --team 时协调工作流                            │
│  - 处理 plan、quality、sync 阶段                             │
│  - 无 GLM 环境变量 → 使用 Claude API                         │
└──────────────────────┬──────────────────────────────────────┘
                       │ Agent Teams（新 tmux 窗格）
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  TEAMMATES（新 tmux 窗格，GLM API）                          │
│  - 继承 tmux 会话环境变量 → 使用 GLM API                     │
│  - 执行 run 阶段的实现任务                                   │
│  - 通过 SendMessage 与领导者通信                             │
└─────────────────────────────────────────────────────────────┘
```

### 使用方法

```bash
# 1. 保存 GLM API 密钥（首次）
moai glm sk-your-glm-api-key

# 2. 确认 tmux 环境（已在使用 tmux 则跳过）
# 如果需要新的 tmux 会话:
tmux new -s moai

# 提示：将 VS Code 终端默认设置为 tmux，
# 可自动在 tmux 环境中启动，跳过此步骤。

# 3. 启动 CG 模式（自动启动 Claude Code）
moai cg

# 4. 运行团队工作流
/moai --team "任务描述"
```

### 注意事项

| 项目 | 说明 |
|------|------|
| **tmux 环境** | 如果已在使用 tmux，无需创建新会话。将 VS Code 终端默认设置为 tmux 会更方便。 |
| **自动启动** | `moai cg` 会在当前窗格自动启动 Claude Code。无需单独运行 `claude`。 |
| **会话结束时** | session_end 钩子自动清除 tmux 会话环境变量 → 下个会话恢复使用 Claude |
| **Agent Teams 通信** | 使用 SendMessage 工具可实现领导者↔工作者间通信 |

### 模式对比

| 命令 | 领导者 | 工作者 | 需要 tmux | 成本节省 | 使用场景 |
|------|--------|--------|-----------|----------|----------|
| `moai cc` | Claude | Claude | 否 | - | 复杂工作、最高质量 |
| `moai glm` | GLM | GLM | 推荐 | ~70% | 成本优化 |
| `moai cg` | Claude | GLM | **必需** | **~60%** | 质量与成本平衡 |

### 显示模式

Agent Teams 支持两种显示模式:

| 模式 | 说明 | 通信 | 领导者/工作者分离 |
|------|------|------|------------------|
| `in-process` | 默认模式，所有终端 | ✅ SendMessage | ❌ 相同环境变量 |
| `tmux` | 分割窗格显示 | ✅ SendMessage | ✅ 会话环境变量隔离 |

**CG 模式仅在 `tmux` 显示模式下支持领导者/工作者 API 分离。**

---

## Claude x GLM 多模型

MoAI-ADK 支持 **z.ai GLM** 作为 Claude Code 的替代 AI 后端，实现多模型开发工作流。

| 项目 | 详情 |
|------|------|
| GLM Coding Plan | **$10/月**起（[z.ai](https://z.ai/subscribe?ic=1NDV03BGWU)） |
| 兼容性 | 无需修改代码，直接与 Claude Code 配合使用 |
| 模型 | GLM-5、GLM-4.7、GLM-4.5-Air 及免费模型 |

**默认模型映射：**

| Claude 层级 | GLM 模型 | 输入（每百万 Token） | 输出（每百万 Token） |
|------------|----------|-------------------|-------------------|
| Opus | GLM-4.7 | $0.60 | $2.20 |
| Sonnet | GLM-4.7 | $0.60 | $2.20 |
| Haiku | GLM-4.5-Air | $0.20 | $1.10 |

> 免费模型也可使用：GLM-4.7-Flash、GLM-4.5-Flash。完整价格详见 [z.ai Pricing](https://docs.z.ai/guides/overview/pricing)。

**[注册 GLM Coding Plan](https://z.ai/subscribe?ic=1NDV03BGWU)**

---

## @MX 标签系统

MoAI-ADK 使用 **@MX 代码级注解系统**在 AI 智能体之间传递上下文、不变量和危险区域。

### 什么是 @MX 标签？

@MX 标签是内联代码注解，帮助 AI 智能体更快、更准确地理解您的代码库。

```go
// @MX:ANCHOR: [AUTO] 钩子注册分发 - 5+ 个调用者
// @MX:REASON: [AUTO] 所有钩子事件的中央入口点，变更影响范围广
func DispatchHook(event string, data []byte) error {
    // ...
}

// @MX:WARN: [AUTO] Goroutine 执行时没有 context.Context
// @MX:REASON: [AUTO] 无法取消 goroutine，潜在资源泄漏
func processAsync() {
    go func() {
        // ...
    }()
}
```

### 标签类型

| 标签类型 | 用途 | 说明 |
|---------|------|------|
| `@MX:ANCHOR` | 重要合约 | fan_in >= 3 的函数，变更影响范围广 |
| `@MX:WARN` | 危险区域 | Goroutines、复杂度 >= 15、全局状态变更 |
| `@MX:NOTE` | 上下文 | 魔法常数、缺少 godoc、业务规则 |
| `@MX:TODO` | 未完成工作 | 缺少测试、未实现的功能 |

### 为什么不是每个代码都有 @MX 标签？

@MX 标签系统**不设计用于给所有代码添加标签**。核心原则是**"仅标记 AI 需要首先注意的最危险/最重要的代码。"**

| 优先级 | 条件 | 标签类型 |
|--------|------|----------|
| **P1（关键）** | fan_in >= 3 | `@MX:ANCHOR` |
| **P2（危险）** | goroutine、复杂度 >= 15 | `@MX:WARN` |
| **P3（上下文）** | 魔法常数、无 godoc | `@MX:NOTE` |
| **P4（缺失）** | 无测试文件 | `@MX:TODO` |

**大多数代码不符合任何条件，因此没有标签。** 这是**正常的**。

### 示例：标签决策

```go
// ❌ 无标签（fan_in = 1、低复杂度）
func calculateTotal(items []Item) int {
    total := 0
    for _, item := range items {
        total += item.Price
    }
    return total
}

// ✅ 添加了 @MX:ANCHOR（fan_in = 5）
// @MX:ANCHOR: [AUTO] 配置管理器加载 - 5+ 个调用者
// @MX:REASON: [AUTO] 所有 CLI 命令的入口点
func LoadConfig() (*Config, error) {
    // ...
}
```

### 配置（`.moai/config/sections/mx.yaml`）

```yaml
thresholds:
  fan_in_anchor: 3        # < 3 个调用者 = 无 ANCHOR
  complexity_warn: 15     # < 15 复杂度 = 无 WARN
  branch_warn: 8          # < 8 个分支 = 无 WARN

limits:
  anchor_per_file: 3      # 每个文件最多 3 个 ANCHOR 标签
  warn_per_file: 5        # 每个文件最多 5 个 WARN 标签

exclude:
  - "**/*_generated.go"   # 排除生成的文件
  - "**/vendor/**"        # 排除外部库
  - "**/mock_*.go"        # 排除 mock 文件
```

### 运行 MX 标签扫描

```bash
# 扫描整个代码库（Go 项目）
/moai mx --all

# 仅预览（不修改文件）
/moai mx --dry

# 按优先级扫描（仅 P1）
/moai mx --priority P1

# 仅扫描特定语言
/moai mx --all --lang go,python
```

### 为什么其他项目也有很少的 MX 标签

| 情况 | 原因 |
|------|------|
| **新项目** | 大多数函数 fan_in = 0 → 无标签（正常） |
| **小项目** | 函数少 = 简单的调用图 = 标签少 |
| **高质量代码** | 低复杂度、无 goroutines → 无 WARN 标签 |
| **高阈值** | `fan_in_anchor: 5` = 更少的标签 |

### 核心原则

@MX 标签系统优化**"信噪比"**：

- ✅ **仅标记真正重要的代码** → AI 快速识别核心区域
- ❌ **标记所有代码** → 增加噪音，重要标签更难找到

---

## Design System: Hybrid Web & App Production (v3.2, SPEC-AGENCY-ABSORB-001)

> 只需描述你想要什么。Design System 会自主完成访谈、设计、构建、测试和学习。

MoAI-ADK 包含集成的 **Design System** — 用于自主制作网站和 Web 应用的专用线束。就像 `/moai "描述"` 运行完整开发工作流一样，`/moai design "描述"` 运行从简报到部署代码的完整创意生产流水线。

### 快速开始：一条命令，完整流水线

```bash
/moai design "为我的 AI 开发者工具创业公司制作 SaaS 落地页"
```

这一条命令触发**完整自主工作流**：

1. **客户访谈** — Manager-spec 进行关于业务、品牌和技术偏好的 9 个结构化问题（已配置则跳过）
2. **BRIEF 生成** — Manager-spec 将请求扩展为全面的项目简报
3. **文案 + 设计** — moai-domain-copywriting 撰写品牌一致的营销文案；moai-domain-brand-design 创建基于令牌的设计系统（路径 B）。替代路径 A：moai-workflow-design-import 解析 Claude Design 交接包
4. **代码实现** — expert-frontend 用 TDD 实现生产代码（默认：Next.js + Tailwind）
5. **质量保证** — evaluator-active 运行 Playwright 测试、Lighthouse 审计和 4 维评分
6. **GAN Loop** — 质量不达标时 expert-frontend 和 evaluator-active 通过 moai-workflow-gan-loop 迭代（最多 5 轮）
7. **自我学习** — (可选) Learner 检测模式并提出技能改进建议

**典型耗时**：完整落地页 15-45 分钟，完全自主。

### 流水线架构

```mermaid
flowchart LR
    REQ["🎯 /moai design '请求'"] --> INT["📋 客户访谈"]
    INT --> P["📝 Manager-Spec (BRIEF)"]
    P --> C["✍️ Copywriting"]
    P --> D["🎨 Brand Design"]
    C --> B["🔨 Builder (TDD)"]
    D --> B
    B --> E["🔍 Evaluator"]
    E -->|"FAIL（最多5轮）"| B
    E -->|"PASS（分数 ≥ 0.75）"| L["🧠 Learner (可选)"]
```

### 各技能职责

| 技能 | 功能 |
|------|------|
| **manager-spec** | 进行客户访谈，生成结构化 BRIEF 文档 |
| **moai-domain-copywriting** | 以结构化 JSON 撰写营销文案 — 标题、正文、CTA — 遵循品牌语调规则 |
| **moai-domain-brand-design** | 创建完整设计系统 — 颜色令牌、字体比例、间距、组件规范 (路径 B) |
| **moai-workflow-design-import** | 解析 Claude Design 交接包(ZIP/HTML)获取设计令牌和组件 (路径 A) |
| **expert-frontend** | 用 TDD（RED-GREEN-REFACTOR）实现生产代码。默认栈：Next.js、TypeScript、Tailwind、shadcn/ui |
| **evaluator-active** | Playwright 视觉测试 + Lighthouse 审计。4 维评分：设计质量(30%)、原创性(25%)、完整度(25%)、功能性(20%) |
| **moai-workflow-gan-loop** | GAN Loop 迭代管理：Builder-Evaluator 间协商 Sprint Contract、实施、评分、停滞检测 |

### GAN Loop：对抗性质量保证

evaluator-active **默认持怀疑态度** — 调整为发现缺陷而非合理化接受。

**自动不合格触发条件**（与分数无关）：
- 文案文本与 moai-domain-copywriting 输出不一致
- 检测到 AI 设计陈词滥调（紫色渐变 + 白色卡片 + 通用图标）
- 移动端视口崩溃
- 链接 404 错误
- Lighthouse 可访问性 < 80

**迭代流程**：evaluator-active 提供带 file:line 引用的具体反馈 → Builder 修复 → 重新评估。3 次失败后升级给用户。

### 品牌上下文：创意宪法

首次运行时，Design System 进行**结构化客户访谈**（4 阶段 9 个问题）：

| 阶段 | 问题 | 存储位置 |
|------|------|----------|
| 业务上下文 | 目标、目标客户、成功 KPI | `.moai/project/brand/target-audience.md` |
| 品牌身份 | 语调形容词、参考网站、设计偏好 | `.moai/project/brand/brand-voice.md`、`visual-identity.md` |
| 技术范围 | 所需页面、技术要求 | `.moai/project/tech.md` |
| 质量期望 | 优先事项 | `.moai/config/sections/design.yaml` |

品牌上下文作为**不可变约束**传递给所有技能。evaluator-active 将品牌一致性作为必须通过的标准进行评分。完成 5+ 项目后，访谈缩减为 3 个关键问题。

### 自我进化与安全

所有技能采用**静态 + 动态区**结构：
- **静态区**：核心原则（永不自动修改）
- **动态区**：规则、启发式、反模式（通过 Learner 进化）

**知识毕业**：observation（1次）→ heuristic（3次）→ rule（5次，置信度 ≥ 0.80）→ graduated（用户批准后应用）

**5 层安全架构**：
1. **Frozen Guard** — 阻止修改身份、安全护栏和伦理边界
2. **Canary Check** — 对最近 3 个项目进行影子评估，分数下降 > 0.10 则拒绝
3. **Contradiction Detector** — 标记与现有规则冲突的规则
4. **Rate Limiter** — 每周最多 3 次进化，24 小时冷却，最多 50 个活跃学习
5. **Human Oversight** — 展示 before/after diff 及证据，需用户批准

### 命令

```bash
# 自主工作流（推荐）
/moai design "为我的 AI 创业公司制作 SaaS 落地页"  # 完整流水线：访谈 → 构建 → 测试 → 学习

# 分步工作流
/moai design brief "开发者工具落地页"              # 仅访谈 + BRIEF（构建前审查）
/moai design build BRIEF-001                      # 从现有 BRIEF 运行完整流水线
/moai design import /path/to/design.zip           # Claude Design 交接包 import (路径 A)
```

### 默认技术栈（可配置）

| 层级 | 默认值 | 配置文件 |
|------|--------|----------|
| 框架 | Next.js + App Router | `.moai/project/tech.md` |
| 语言 | TypeScript（strict） | `.moai/project/tech.md` |
| 样式 | Tailwind CSS v4 | `.moai/project/tech.md` |
| 组件 | shadcn/ui | `.moai/project/tech.md` |
| 测试 | Vitest + Playwright | `.moai/config/sections/design.yaml` |
| 托管 | Vercel | `.moai/project/tech.md` |

### 从 /agency 迁移

使用 `/agency` 的现有项目可以通过以下命令迁移到 `/moai design`：

```bash
moai migrate agency
```

此命令将 `.agency/` 数据安全地移动到 `.moai/project/brand/` 和 `.moai/config/sections/design.yaml`。 原始数据保留为 `.agency.archived/` 以供恢复。

> [Design System 文档](https://adk.mo.ai.kr)

---

## 数据库工作流: /moai db

MoAI 项目的数据库元数据管理系统。通过四个子命令（init、refresh、verify、list）管理架构文档、迁移、ERD 图表和种子数据。

### 快速开始

```bash
# 初始化数据库元数据（交互式问卷）
/moai db init

# 重新扫描迁移并更新架构文档
/moai db refresh

# 检查 schema.md 和迁移文件之间的偏差
/moai db verify

# 显示 schema.md 中的所有表
/moai db list
```

### 子命令

| 命令 | 目的 | 使用时机 |
|------|------|---------|
| **init** | 数据库引擎、ORM、多租户策略和迁移工具的交互式设置。在 `.moai/project/db/` 中脚手架 7 文件模板集 | 新项目初始化、任何数据库工作之前 |
| **refresh** | 扫描迁移文件并从当前迁移状态重新生成 `schema.md`、`erd.mmd`（Mermaid ERD）和 `migrations.md` | 添加/修改迁移之后、里程碑同步时 |
| **verify** | 只读偏差检测：比较 `schema.md` 表集与实际迁移文件，发现偏差时以非零状态退出 | PR 提交前、CI/CD 管道中 |
| **list** | 只读表列表：以对齐的 Markdown 表格格式显示 `schema.md` 中的所有表 | 快速项目概览、文档审查 |

### 目录结构

`/moai db init` 在 `.moai/project/db/` 中创建以下结构：

```
.moai/project/db/
├── README.md              # 数据库概览和设置说明
├── schema.md              # 表架构文档（自动生成）
├── erd.mmd                # Mermaid 格式的实体关系图
├── migrations.md          # 迁移历史和顺序
├── rls-policies.md        # 行级安全策略（PostgreSQL）
├── queries.md             # 重要查询和性能注释
└── seed-data.md           # 示例数据和播种说明
```

### 支持的数据库技术

自动检测并支持 6 种迁移文件模式：

| 迁移类型 | 文件模式 | 示例 |
|--------|---------|------|
| **Prisma** | `prisma/migrations/*/migration.sql` | `20260401120000_add_users_table/migration.sql` |
| **Alembic** | `alembic/versions/*.py` | `a1b2c3d4e5f6_add_users_table.py` |
| **Rails** | `db/migrate/*.rb` | `20260401120000_add_users_table.rb` |
| **Raw SQL** | `db/migrations/*.sql` | `001_add_users_table.sql` |
| **Supabase** | `supabase/migrations/*.sql` | `20260401120000_initial_schema.sql` |
| **通用** | `migrations/*.sql` 或 `db/*.sql` | 支持自定义模式 |

通过常见包路径支持 16 个编程语言生态系统（Go、Python、TypeScript、Java 等）。

### 集成

- **PostToolUse Hook**：编辑迁移文件时自动刷新 `schema.md`、`erd.mmd`、`migrations.md`
- **偏差检测**：防止架构文档与实际迁移不同步
- **Mermaid 图表**：为文档和设计审查自动生成 ERD 图表
- **Phase 4.1a 数据库检测**：`/moai project` 根据检测到的数据库技术自动显示 `/moai db` 建议

### 配置

数据库设置存储在 `.moai/config/sections/db.yaml` 中：

```yaml
db:
  enabled: true
  dir: ".moai/project/db"
  auto_sync: true
  migration_patterns:
    - "prisma/migrations/*/migration.sql"
    - "alembic/versions/*.py"
    - "db/migrate/*.rb"
  engine: ""  # 在 init 问卷期间填充
  orm: ""     # 在 init 问卷期间填充
  multi_tenant: false
  migration_tool: ""
```

### 工作流示例

1. **新项目**：运行 `/moai db init`，回答关于数据库设置的 4 个问题
2. **开发中**：照常创建迁移；`/moai db` 自动同步文档
3. **PR 前**：运行 `/moai db verify` 检查架构偏差
4. **审查**：在 PR 中参考 `.moai/project/db/erd.mmd` 进行架构视觉审查

### 使用时机

- **始终启用**：在 `moai init` 期间为任何具有数据库的项目启用
- **Init**：新项目、数据库架构更改
- **Refresh**：大型迁移工作后、主要提交前
- **Verify**：CI/CD 管道的一部分、PR 前检查
- **List**：快速参考、文档生成

---

## 常见问题

### Q: 为什么不是所有 Go 代码都有 @MX 标签？

**A: 这是正常的。** @MX 标签是"按需添加的"。大多数代码足够简单安全，不需要标签。

| 问题 | 答案 |
|------|------|
| 没有标签是问题吗？ | **不是。** 大多数代码不需要标签。 |
| 何时添加标签？ | **仅高 fan_in、复杂逻辑、危险模式** |
| 所有项目都类似吗？ | **是的。** 每个项目中的大多数代码都没有标签。 |

详见上方的 **"@MX 标签系统"** 部分。

---

### Q: 如何自定义显示的状态栏段？

Statusline v3 提供**多行布局**和实时 API 使用量监控：

**Full 模式**（5 行 — 40 块独立条）：
```
🤖 Opus 4.6 │ 🔅 v2.1.74 │ 🗿 v2.7.12 │ ⏳ 5h 32m │ 💬 MoAI
CW: 🔋 █████████████████████░░░░░░░░░░░░░░░░░░░ 52%
5H: 🔋 █░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 4%
7D: 🔋 ██████████████████████░░░░░░░░░░░░░░░░░░░ 56%
📁 moai-adk-go │ 🔀 main │ 📊 +0 M38 ?2
```

**默认（Default）模式**（3 行 — 10 块内联条）：
```
🤖 Opus 4.6 │ 🔅 v2.1.74 │ 🗿 v2.7.12 │ ⏳ 16m │ 💬 MoAI
CW: 🔋 ██░░░░░░░░ 25% │ 5H: 🔋 █░░░░░░░░░ 12% │ 7D: 🔋 ░░░░░░░░░░ 3%
📁 moai-adk-go │ 🔀 fix/my-feature │ 📊 +0 M38 ?2
```

支持 2 种显示模式：

- **Full**（5 行）：所有段 + 40 块使用量条独立行显示（model、context、usage bars、git、version、output style、directory）
- **Default**（3 行）：核心段 + 10 块内联使用量条（model、context、usage bars、git status、branch、version）

直接编辑 `.moai/config/sections/statusline.yaml`：

```yaml
statusline:
  preset: default  # 或 full
  segments:
    model: true
    context: true
    usage_5h: true    # 5 小时 API 使用量条
    usage_7d: true    # 7 天 API 使用量条
    output_style: true
    directory: true
    git_status: true
    claude_version: true
    moai_version: true
    git_branch: true
```

> **注意**：从 v2.7.8 开始，段预设选择 UI 已从 `moai init`/`moai update` 向导中移除。请直接在上述 YAML 文件中配置。

---

### Q: 状态栏中的版本指示符是什么意思？

MoAI 状态栏显示带有更新通知的版本信息：

```
🗿 v2.2.2 ⬆️ v2.2.5
```

- **`v2.2.2`**：当前安装的版本
- **`⬆️ v2.2.5`**：新版本可用于更新

当您在最新版本上时，只显示版本号：
```
🗿 v2.2.5
```

**更新方法**：运行 `moai update`，更新通知将消失。

**注意**：这与 Claude Code 的内置版本指示符（`🔅 v2.1.38`）不同。MoAI 指示符跟踪 MoAI-ADK 版本，Claude Code 单独显示其自身版本。

---

### Q: 出现"允许外部 CLAUDE.md 文件导入？"警告

打开项目时，Claude Code 可能显示关于外部文件导入的安全提示：

```
External imports:
  /Users/<user>/.moai/config/sections/quality.yaml
  /Users/<user>/.moai/config/sections/user.yaml
  /Users/<user>/.moai/config/sections/language.yaml
```

**推荐操作**：选择 **"No, disable external imports"** ✅

**为什么？**
- 您的项目的 `.moai/config/sections/` 已包含这些文件
- 项目特定设置优先于全局设置
- 基本配置已嵌入 CLAUDE.md 文本
- 禁用外部导入更安全且不影响功能

**这些文件是什么？**
- `quality.yaml`：TRUST 5 框架和开发方法论设置
- `language.yaml`：语言偏好（对话、注释、提交）
- `user.yaml`：用户名（可选，用于 Co-Authored-By 属性）

---

## 贡献

欢迎贡献！详细指南请参阅 [CONTRIBUTING.md](CONTRIBUTING.md)。

### 快速入门

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 编写测试（新代码用 TDD，现有代码用特征测试）
4. 确认所有测试通过：`make test`
5. 确认 Lint 通过：`make lint`
6. 格式化代码：`make fmt`
7. 使用约定式提交信息提交
8. 提交 Pull Request

**代码质量要求**：85%+ 覆盖率 · 0 Lint 错误 · 0 类型错误 · 约定式提交

### 社区

- [Issues](https://github.com/modu-ai/moai-adk/issues) -- Bug 报告、功能请求

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=modu-ai/moai-adk&type=date&legend=top-left)](https://www.star-history.com/#modu-ai/moai-adk&type=date&legend=top-left)

---

## 许可证

[Apache License 2.0](./LICENSE) -- 详情请参阅 LICENSE 文件。

## 相关链接

- [官方文档](https://adk.mo.ai.kr)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code)
