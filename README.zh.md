<p align="center">
  <img src="./assets/images/moai-adk-og.png" alt="MoAI-ADK" width="100%">
</p>

<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>为 Tokenomics 而构建的 Agentic 开发套件</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <a href="./README.ja.md">日本語</a> ·
  中文
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
  <a href="https://adk.mo.ai.kr"><strong>官方文档</strong></a> ·
  <a href="https://adk.mo.ai.kr/book">图书：Claude Code 实战 Agentic 编程</a> ·
  <a href="https://discord.gg/Z7E7Mdc5aN">Discord</a>
</p>

---

> **"Vibe Coding 的目的不是快速的生产力，而是代码质量。"**

MoAI-ADK (Agentic Development Kit) 是一套以 **Tokenomics** (Token 经济学) 为北极星的 Agentic 开发套件：用更少的 Token 达到同样的代码质量，用同样的 Token 达到更高的质量。模型选择、推理深度和上下文使用由系统管理——而非听天由命。

以 Go 编写的单一二进制文件。在 macOS、Linux、Windows 上零依赖即刻运行。

---

## 为什么是 Tokenomics

Token 价格持续下降，但 Agentic 开发消耗 Token 的速度比降价更快。并行运行的代理越来越多，上下文越来越长，推理越来越深——因此你的真实成本 **不由模型的标价决定，而由你如何运营 Token 决定**。

MoAI-ADK 的答案分为三部分：

1. **为每个任务分配合适的模型与推理深度**——深度规划、廉价实现、独立验证。
2. **给上下文瘦身**——最小化常驻加载的指令，并度量提示缓存命中率。
3. **让系统守护预算**——按代理跟踪 Token 用量，在触顶前优雅停止，绝不中途崩溃。

---

## 三大支柱

### 支柱 1 — Tokenomics (Token 经济学)

最大化每美元质量的智能资源分配。No-Haiku 三层模型策略 (max / medium / low)、感知计费方案的层级配置 (API 计量计费 vs. 订阅方案)、Claude × GLM 混合模式 (CG 模式，在实现密集型工作上降低 60-70% 成本)，以及在预算超支前优雅中止的 Token 断路器。

### 支柱 2 — 递归自我学习

循环积累观察；Harness 学习；指令进化。Routing Observation Ledger 记录路由决策，Curator 将其转化为改进提案，四层学习阶梯 (观察 → 启发式 → 规则 → 自动更新) 升级 Harness——始终置于用户批准门之后。

### 支柱 3 — Agentic Harness

与其直接编写代码，不如设计一个让代理良好工作的环境：11 个代理的目录、基于 SPEC 的三阶段工作流 (plan → run → sync)、TRUST 5 质量门，以及能从自然语言请求生成项目专属 Harness 的 Harness v4 Builder。

---

## v3 数字一览

从 v2.14.0 (2026-04-24) 到 v3.0.0-rc12 (2026-07-13) —— **80 天**：

- 两个标签之间 **2,373 次提交** —— feat 727 · docs 517 · fix 240
- **9 个候选版本** (rc1 → rc12)
- 代理目录整合 **22 → 11** (10 个 MoAI 自定义代理 + 内置 Explore，更少的代理、更廉价的委派)
- **480+ 份 SPEC 文档** 在 `.moai/specs/` 下驱动 SPEC 优先开发
- **27** 个模板管理的 `moai-*` 技能 · **36** 个顶级 CLI 命令 · 支持 **16** 种编程语言

---

## 快速开始

### 1. 安装

#### macOS / Linux / WSL

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

#### Windows (PowerShell 7.x+)

> **推荐**：为获得最佳体验，请使用 WSL 并执行上面的 Linux 安装命令。

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

> 需要先安装 [Git for Windows](https://gitforwindows.org/)。

#### 从源码构建 (Go 1.26+)

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk && make build
```

> 预构建的二进制文件可在 [Releases](https://github.com/modu-ai/moai-adk/releases) 页面获取。

### 2. 初始化项目

```bash
moai init my-project
```

交互式向导会自动检测你的语言、框架和方法论，选择模型策略，并生成 Claude Code 集成文件。

### 3. 用 Claude Code 开始开发

```bash
claude        # 在项目内启动 Claude Code
```

```text
/moai plan "Add JWT login"      # 编写 SPEC
/moai run SPEC-AUTH-001         # TDD/DDD 实现
/moai sync SPEC-AUTH-001        # 同步文档 + 创建 PR
```

你也可以直接用自然语言提问——`/moai "fix the login bug"` 会经过意图分析 (Analyze-First 路由) 并落到正确的工作流，任何对话语言均可。

```mermaid
flowchart TD
    A["/moai project"] --> B["/moai plan"]
    B -->|"SPEC document"| C["/moai run"]
    C -->|"implementation complete"| D["/moai sync"]
    D -->|"PR created"| E["Done"]
```

### 4. Windows 注意事项：非 ASCII 用户名路径

如果你的 Windows 用户名包含非 ASCII 字符 (韩文、中文等)，可能会遇到由 Windows 8.3 短文件名转换引起的 `EINVAL` 错误。解决办法：

```powershell
# Option 1: point MoAI at an ASCII-only temp directory
$env:MOAI_TEMP_DIR="C:\temp"
New-Item -ItemType Directory -Path "C:\temp" -Force

# Option 2: disable 8.3 filename generation (requires admin)
fsutil 8dot3name set 1
```

第三个选项是创建一个仅含 ASCII 用户名的 Windows 账户。

---

## 系统要求

| 平台 | 支持的环境 | 备注 |
|----------|----------------------|-------|
| macOS | Terminal, iTerm2 | 完全支持 |
| Linux | Bash, Zsh | 完全支持 |
| Windows | **WSL (推荐)**, PowerShell 7.x+ | 不支持原生 cmd.exe |

**前置条件：**

- 所有平台必须安装 **Git**
- **Claude Code** —— MoAI-ADK 是面向 Claude Code 的 Harness
- **Windows 用户**：**必须** 安装 [Git for Windows](https://gitforwindows.org/) (含 Git Bash)；**不支持** 旧版 Windows PowerShell 5.x 和 cmd.exe
- **推荐**：`gh` CLI (PR 自动化) · `tmux` (CG 模式) · 你所用语言的 lint/测试工具链 (例如 `golangci-lint`)

---

## 视觉识别 — 吉祥物主题

文档站点([adk.mo.ai.kr](https://adk.mo.ai.kr))与 `moai web` 控制台共享 **吉祥物灰** 主题 —— `#8c8c8c`（neutral gray），源自 모두의AI 角色吉祥物(MascotCoding / MascotTalking / MascotBubble)。吉祥物在主视觉区、404 页面、章节分隔处作为情感锚点出现。

---

## 设计谱系 — Harness Engineering

MoAI-ADK 有意继承了 Lilian Weng 在 [**Harness Engineering for Self-Improvement**](https://lilianweng.github.io/posts/2026-07-04-harness/) (2026-07-04) 中提出的 Harness 工程框架，将其设计模式与自我改进循环转化为一个可运行的实现。

> **什么是 Harness？** —— "Harness 是围绕基础模型的系统，它编排执行并决定模型如何思考与规划、如何调用工具与行动、如何感知与管理上下文、如何存储产物、如何评估结果。" —— Lilian Weng (2026-07-04)

Weng 预测，通往递归自我改进 (RSI) 的近期路径不是"模型编辑自己的权重"，而是 **改进训练管线与部署系统——即 Harness**。MoAI-ADK 走的正是这条路：它递归改进的是 Harness (技能与代理指令)，而非模型权重。

### 继承映射 — 从 Weng 的框架到 MoAI-ADK

| Lilian Weng 的 Harness 概念 | MoAI-ADK 实现 |
|---|---|
| **Harness** —— 围绕基础模型的执行/运营层 | MoAI-ADK = Claude Code Harness (单一 Go 二进制 + CLAUDE.md 编排器) |
| **模式 1：工作流自动化** —— plan → execute → observe → improve 目标循环 | `/moai goal` 引擎、`/moai loop` Ralph Engine、Analyze-First 路由 |
| **模式 2：文件系统持久记忆** —— "把持久状态放进文件" | `.moai/specs/`、`progress.md`、`usage-log.jsonl`、`.moai/state/`、会话交接 |
| **模式 3：子代理与后台作业** —— 让并行性显式且可检视 | 11 个保留代理、`Agent()` 生成、动态工作流 |
| **Self-Harness** —— 提案-评估-接受；受限编辑 + 回归门 | `internal/harness/` 四层阶梯 + 五层安全管线 (applier = 受限编辑，回归门 = 验证) |
| **Meta-Harness** —— "优化 Harness 的 Harness" | `builder-harness` —— 用 Harness 构建 Harness；`/moai project` 自动生成一个 |
| **"改进改进者"** —— RSI 的近期路径是部署系统的改进 | 递归 Harness 进化 —— 循环积累观察；Harness 升级自己的技能/代理指令 |
| **"评估器与权限位于循环之外"** —— 防御奖励劫持 | 第 5 层用户批准门 + 实现启动批准 —— 人类监督位于进化循环之外 |
| **"人类沿栈上移，而非退出循环"** | 编排器是唯一的人类接触点；由 AskUserQuestion 把关的决策与 SPEC 批准门 |

> Weng 的警告被忠实遵守：评估器与权限控制必须留在 Harness 进化循环 **之外**。MoAI-ADK 将 Tier-4 自动更新绑定到用户批准门，使自动进化永远无法在没有人类监督的情况下作为闭环运行。

---

## Tokenomics 深入解读

### No-Haiku 三层模型策略

模型与推理深度 (effort) 按工作阶段和 SPEC 规模 (Tier S/M/L) 声明式分配。策略层级构成一个封闭集合——`max`、`medium`、`low`——由 `internal/config/model_routing.go` 中的 HARD lint 规则校验 (封闭集合：effort `low/medium/high/xhigh/max`、tier `S/M/L`、phase `plan/run/sync`)。

| 策略 | 目标方案 | 特点 |
|--------|-------------|-----------|
| **max** | Max $200/月 | 最高质量 —— 规划与审计使用 Opus 级模型 |
| **medium** | Max $100/月 | 质量与成本均衡 |
| **low** | Plus $20/月 | 无 Opus 访问 —— 以 Sonnet 为中心的路由 |

"No-Haiku" 这个名字标志着 v3 的转变：不再把质量关键阶段路由到最便宜的模型——廉价模型只用在安全的地方，绝不用在需要独立判断的地方。

### 感知计费方案的层级配置 (plan_type)

同一工作流在 **API 计量计费 vs. 订阅方案** 下的最优分配不同。感知计费方案的配置为每种计费方案应用独立的 Tier × Phase 模型/effort 矩阵，并为 GLM 后端叠加 effort 覆盖层。

### Claude × GLM 混合模式 (CG 模式)

`moai cg` 以 Claude 为 leader、GLM 为 worker 运行：战略、规划与审计留在 Claude API 上，大批量实现交给 GLM。在实现密集型工作上可削减 **60-70%** 的成本。

MoAI-ADK 支持 **z.ai GLM** 作为 Claude Code 的替代后端——无需任何代码修改。

| 项目 | 详情 |
|------|---------|
| GLM Coding Plan | **$10/月** 起 ([z.ai](https://z.ai/subscribe?ic=1NDV03BGWU)) |
| 兼容性 | 与 Claude Code 直接兼容 |
| 模型 | glm-5.2[1m]、glm-4.7、glm-4.5-air 及免费模型 |

**默认模型映射：**

| Claude 层级 | GLM 模型 | 输入 (每 1M Token) | 输出 (每 1M Token) |
|-------------|-----------|----------------------|------------------------|
| Opus / Sonnet / Haiku / Fable | glm-5.2[1m] | $2.00 | $8.00 |

> 四个 Claude 层级全部统一到单一 1M 上下文模型 `glm-5.2[1m]`。若在层级槽位之间混用 1M 上下文模型与 200K 上下文模型，会破坏代理生成的会话共享——1M 上下文会话与 200K 上下文会话无法共享。

> `[1m]` 后缀激活 Claude Code 的 1M Token 上下文模式。Claude Code 会在调用上游 z.ai API 之前解析并剥离该后缀。此映射通过四个 `ANTHROPIC_DEFAULT_*_MODEL` 环境变量实现 (`OPUS`/`SONNET`/`HAIKU`/`FABLE`，最后一个自 Claude Code v2.1.202 起获得官方支持)，全部设为 `glm-5.2`。

**模式对比：**

| 命令 | Leader | Workers | tmux | 成本节省 | 适用场景 |
|---------|--------|---------|------|--------------|----------|
| `moai cc` | Claude | Claude | 否 | — | 复杂工作，最高质量 |
| `moai glm` | GLM | GLM | 推荐 | ~70% | 最大化成本节省 |
| `moai cg` | Claude | GLM | **必需** | **~60%** | 质量与成本平衡 |

**CG 模式实操：**

```bash
# 1. Save your GLM API key (once)
moai glm sk-your-glm-api-key

# 2. Make sure you are inside tmux (skip if already there)
tmux new -s moai

# 3. Launch CG mode (starts Claude Code automatically)
moai cg
```

CG 模式通过 tmux 会话级环境变量将 leader 与 worker 隔离：GLM 配置注入 tmux 会话环境 (worker 在新 pane 中继承)，并从 `settings.local.json` 中移除 (leader pane 保持在 Claude API 上)。会话结束钩子会自动清理 tmux 环境。

### Token 断路器

`internal/runtime/budget.go` 以"警告优先"策略按代理跟踪 Token 用量：用量攀升时发出警告，并在硬阈值处执行 **优雅中止** (保存进度 + 发出交接消息)。它绝不会自动清除你的会话。

### 上下文瘦身 + 提示缓存

- 常驻加载上下文预算守卫——精简的 CLAUDE.md 加上按路径限定的规则文件，压低每回合固定成本
- **缓存命中率** 状态栏区段让瘦身效果实时可测
- 验证输出遵循文件重定向契约——长日志写入磁盘；上下文只携带退出码和有界尾部

→ 阅读更多：[Tokenomics 总览](https://adk.mo.ai.kr/zh/advanced/tokenomics-overview) · [提示缓存](https://adk.mo.ai.kr/zh/cost-optimization/prompt-caching)

---

## 递归自我学习

MoAI-ADK 的核心创新是一个代理从自身运行中学习的递归系统。它由两个动作组成：积累观察的循环，与从中进化的 Harness。

```mermaid
flowchart TD
    A["User request"] --> B["Goal set via /moai goal"]
    B --> C["Loop executes"]
    C --> D["Observe results"]
    D --> E{"Goal met?"}
    E -->|"No"| C
    E -->|"Yes"| F["Observations recorded"]
    F --> G["Pattern learning (Curator)"]
    G --> H["Instruction evolution (approval gate)"]
    H --> C
```

### 自我进化的 Harness

```
loop runs → observations accumulate (Routing Ledger) → patterns learned (Curator) → instructions evolve (approval gate)
```

- **Routing Observation Ledger** (`internal/harness/routing/`) —— 以保护隐私的摘要形式记录路由决策与门禁证据
- **四层学习阶梯** (`internal/harness/learner.go`) —— 观察 (≥1) → 启发式 (≥3) → 规则 (≥5) → 自动更新 (≥10，需用户批准)；置信度下限 0.70
- **五层安全管线** —— observer (`internal/harness/observer.go`) → learner → applier (`internal/harness/applier.go`，快照优先的受限编辑) → 配置/标记更新器 → 用户批准门；每次应用都可通过 `moai harness rollback` 撤销
- 产物位于 `.moai/harness/` 下 (`usage-log.jsonl`、已学习的规则)

```bash
moai harness status      # learning state: observations, patterns, proposals
moai harness apply       # apply a proposal (passes the user approval gate)
moai harness rollback    # revert the last application
moai harness disable     # turn learning off
```

### /moai goal — 声明式 Agentic 循环

声明一个完成条件，会话就会持续工作，直到条件成立或达到回合上限 (默认 30)。在 `internal/goal/` 中实现为按会话的目标状态 (`.moai/state/goal/<session-id>.json`)，配备混合双层 Stop-hook 评估器——Tier 1 机械检查 (退出码、grep 计数、文件存在性、回合上限) 与 Tier 2 通过检查点进行的编排器自我评估。

```text
/moai goal "go test ./... exits 0 and every AC is recorded as PASS"
/moai goal status
/moai goal clear
```

### /moai loop vs /moai fix — 诊断式自我修复

`/moai loop` 是构建在 Ralph Engine (`internal/ralph/engine.go`) 之上的目标引擎预设：它并行运行 LSP 诊断 + AST-grep + linter 扫描，将发现从 Level 1 (可自动修复) 到 Level 4 (需要人工) 分类，并迭代直到队列清空——配备在同一错误重复出现时切换策略的收敛检测，以及作为安全停止的硬性迭代上限。

| 命令 | 目标 | 执行方式 | 使用时机 |
|---------|------|-----------|-------------|
| `/moai fix` | 单遍修复 | 一次 扫描-分类-修复-验证 | 明确的错误、快速修复 |
| `/moai loop` | 重复直到完成 | 诊断 → 分类 → 修复 → 验证 循环 | 复合错误、根因修复 |

### Analyze-First 路由

与语言无关的意图分析是 `/moai` 的默认路由。请求按语义分类——绝不依赖英文关键词匹配——因此任何对话语言都适用：

1. 意图分析 (与语言无关的分类)
2. 上下文充分性检查 (上下文不足时触发苏格拉底式访谈)
3. 执行计划组合 (技能 / 代理 / 动态工作流链)
4. 编排模式选择 (solo-sequential / parallel-subagents / dynamic-workflow)

### 会话交接自动恢复

在上下文窗口阈值处 (1M 上下文模型为 50%，200K 模型为 90%)，MoAI 发出一条可直接粘贴的恢复消息——包含进度状态、已应用的教训和可验证的前置条件——`/clear` 之后只需粘贴一次，下一个会话即可继续。

→ 阅读更多：[自我进化的 Harness](https://adk.mo.ai.kr/zh/advanced/self-evolving) · [决策记忆](https://adk.mo.ai.kr/zh/advanced/decision-memory)

---

## Agentic Harness

与其直接编写代码，不如构建代理工作的环境。

### 11 代理目录

11 个保留代理：10 个 MoAI 自定义代理加上 Anthropic 内置的 `Explore`。

| 类别 | 代理 | 角色 |
|----------|-------|------|
| **Manager** | manager-spec | plan 阶段 SPEC 编写 |
| | manager-develop | run 阶段 TDD/DDD/autofix 实现 |
| | manager-docs | sync 阶段文档 |
| | manager-git | PR 创建与路由 |
| | manager-design | design 阶段协作 (Claude Design) |
| **Evaluator** | plan-auditor | 独立计划审计 (防止偏见) |
| | sync-auditor | 四维质量评分 (Functionality 40 · Security 25 · Craft 20 · Consistency 15) |
| **Builder** | builder-harness | 搭建项目专属的代理、技能、命令与钩子 |
| **Advisor** | super-advisor | 按需高推理咨询 (E1-E4 升级) |
| **Specialist** | e2e-tester | Web/移动端/桌面端 E2E 测试执行 (CLI 优先) |
| **Built-in** | Explore | 只读代码库探索 |

规划与审计在设计上是分离的——作者绝不为自己的工作打分。

```mermaid
flowchart TD
    U["User request"] --> M["MoAI Orchestrator"]
    M --> MG1["Managers: spec / develop / docs / git / design"]
    M --> EV["Evaluators: plan-auditor / sync-auditor"]
    M --> BD["Builder: builder-harness"]
    M --> AD["Advisor: super-advisor"]
    M --> EX["Explore (built-in)"]
```

### SPEC 三阶段生命周期

```
/moai plan → [plan-auditor audit] → Implementation Kickoff Approval (human gate) → /moai run → /moai sync → [sync-auditor scoring]
```

- 生命周期恰好三个阶段——**plan → run → sync**
- Tier S/M/L 规模分级决定验证深度与 PR 路由
- GEARS 格式需求加验收标准 (AC) —— 完成与否由证据判定，而非"看起来完成了"

```mermaid
flowchart TB
    subgraph Plan ["Plan Phase"]
        P1["Explore codebase"] --> P2["Analyze requirements"]
        P2 --> P3["Author SPEC (GEARS format)"]
    end

    subgraph Run ["Run Phase"]
        R1["Analyze SPEC, plan execution"] --> R2["TDD/DDD implementation"]
        R2 --> R3["TRUST 5 quality validation"]
    end

    subgraph Sync ["Sync Phase"]
        S1["Generate documentation"] --> S2["Update README/CHANGELOG"]
        S2 --> S3["Create pull request"]
    end

    Plan --> Run
    Run --> Sync
```

### 开发方法论 — TDD 与 DDD

MoAI-ADK 在 `moai init` 时根据项目状态选择方法论 (`--mode <ddd|tdd>`，默认：tdd)；之后可通过 `.moai/config/sections/quality.yaml` 中的 `development_mode` 更改。

```mermaid
flowchart TD
    A["Project analysis"] --> B{"New project or<br/>10%+ test coverage?"}
    B -->|"Yes"| C["TDD (default)"]
    B -->|"No"| D["DDD"]
    C --> F["RED → GREEN → REFACTOR"]
    D --> G["ANALYZE → PRESERVE → IMPROVE"]
```

| 方法论 | 循环 | 适用于 |
|-------------|-------|-----|
| **TDD** (默认) | RED (失败测试) → GREEN (最小化通过) → REFACTOR (在绿灯测试下提升质量) | 新项目与功能开发 |
| **DDD** | ANALYZE (依赖、领域边界) → PRESERVE (特征测试) → IMPROVE (在测试保护下增量变更) | 覆盖率 < 10% 的既有代码 |

### TRUST 5 质量门

每次代码变更都按五项标准校验：

| 标准 | 含义 | 校验 |
|-----------|---------|------------|
| **T**ested | 已测试 | 85%+ 覆盖率、特征测试、单元测试通过 |
| **R**eadable | 可读 | 命名清晰、风格一致、0 lint 错误 |
| **U**nified | 统一 | 格式一致、import 排序、遵循项目结构 |
| **S**ecured | 安全 | OWASP 合规、输入校验、0 安全警告 |
| **T**rackable | 可追踪 | 约定式提交、issue 引用、结构化日志 |

### Harness v4 Builder

```text
/moai harness "build me a harness for CLI template development"
```

自然语言请求经过领域/目标/约束提取和批准门，然后生成项目专属的代理、技能与命令。`/moai project` 生成项目文档 (product.md、structure.md、tech.md、codemaps/) 并同时自动配置一个 Harness。

### 编排原语

静态 Agent Teams 层已在 v3 退役。保留三种编排原语，按"计划由谁持有"选择：

| 原语 | 形态 | 适用场景 |
|-----------|-------|----------|
| 顺序子代理 | 编排器逐回合委派 | 编码密集型工作 |
| 并行扇出 | 单回合内多个只读 `Agent()` 调用 | 研究、评审、审计 |
| 动态工作流 | 脚本编排数十个代理；结果保存在脚本变量中 | 代码库扫描、大型迁移 |

原生 Claude Code teammate 运行时 (`moai cg` tmux pane) 不受此次退役影响。

### Ultracode —— xhigh 强度 + 自动编排

```text
/effort ultracode
```

`/effort ultracode` 将 `xhigh` 推理强度与自动动态工作流编排相结合 (Claude Code v2.1.154+)：对会话中的每个实质性任务，自动选择最优的编排原语，大规模扇出以脚本形式运行，其中间结果保存在脚本变量中，而非会话上下文里。适合用于大型并行扫描、审计与迁移 —— 整个代码库扫描或数百个独立任务 —— 这类场景中扇出本身就是主要成本。若只针对单个请求，则在请求前加上 `ultracode` 关键词，而无需切换整个会话。

→ 阅读更多：[动态工作流与 Ultracode](https://adk.mo.ai.kr/zh/advanced/ultracode-workflows)

### 决策记忆

MoAI-ADK 捕获你的 AskUserQuestion 决策并个性化未来的推荐：

- **三层记忆** —— Core (热门偏好) / Recall (近期会话) / Archival (28 天 TTL，软删除)
- **自适应放置** —— 问题在不确定性最高处触发 (p ≈ 0.5)；推荐遵循你被观察到的统计多数，而非系统默认值
- **衰减策略** —— 幂律权重 `(age+1)^(-0.5)`；使用某偏好会刷新它
- **控制** —— `moai preference list | decay-scan | toggle`；敏感的安全领域给出附带披露的中立推荐

→ 阅读更多：[Harness v4 Builder](https://adk.mo.ai.kr/zh/advanced/harness-v4-builder) · [Catalog 系统](https://adk.mo.ai.kr/zh/advanced/catalog-system)

---

## 为什么选择 Go

基于 Python 的 MoAI-ADK (约 73,000 行) 已用 Go 完全重写。

| 方面 | Python 版 | Go 版 |
|--------|---------------|------------|
| 分发 | pip + venv + 依赖 | **单一二进制**，零依赖 |
| 启动时间 | ~800ms 解释器启动 | **~5ms** 原生执行 |
| 并发 | asyncio / threading | **原生 goroutine** |
| 类型安全 | 运行时 (mypy 可选) | **编译期强制** |
| 跨平台 | 需要 Python 运行时 | **预构建二进制** (macOS、Linux、Windows) |
| 钩子执行 | Shell 包装器 + Python | **编译后二进制**，JSON 协议 |

---

## 工具参考

### `/moai` 斜杠子命令

> **重要区分**：`moai` (终端 CLI) ≠ `/moai` (Claude Code 斜杠命令)。前者是在 shell 中运行的 Go 二进制 (`moai init`、`moai doctor`)；后者是在 Claude Code 聊天中运行的 AI 工作流路由器 (`/moai plan`、`/moai run`)。它们是不同的工具。

→ 详情：[工作流命令](https://adk.mo.ai.kr/zh/workflow-commands) · [实用命令](https://adk.mo.ai.kr/zh/utility-commands)

16 个入口——15 个具名子命令加上自然语言默认入口：

| 子命令 | 角色 |
|------------|------|
| `plan` / `run` / `sync` | SPEC 三阶段流水线 |
| `goal` / `loop` / `fix` | 声明式目标循环 · 迭代修复 · 单遍修复 |
| `project` / `harness` | 项目文档 + Harness 生成 · Harness 生命周期 |
| `review` / `gate` / `clean` | 代码评审 · 提交前质量门 · 死代码清除 |
| `mx` / `codemaps` / `feedback` | @MX 注解 · 架构文档 · GitHub issue 报告 |
| `e2e` | 多平台 E2E 测试 (Web/移动端/桌面端，CLI 优先) |
| *(自然语言)* | Analyze-First 路由进入自主的 plan → run → sync 流水线 |

### CLI 命令 (36 个顶级命令)

`moai` 二进制注册了 36 个顶级命令。其中 `goal`、`handoff`、`harness`、`init`、`launchers`、`loop`、`pr`、`session`、`spec`、`tool-policy`、`worktree` 等 **11 个命令在文档站点拥有独立的 CLI 参考页面**。

→ 详情：[CLI 参考](https://adk.mo.ai.kr/zh/cli-reference)

日常常用集合：

| 命令 | 描述 |
|---------|-------------|
| `moai init` | 交互式项目设置 (语言/框架/方法论自动检测) |
| `moai doctor` | 系统健康诊断与环境验证 |
| `moai status` | 项目状态摘要 (Git 分支、质量指标) |
| `moai update` | 更新到最新版本 (支持自动回滚) |
| `moai update -c` | 重新运行 init 向导以编辑配置 (不同步模板) |
| `moai cc` / `moai glm` / `moai cg` | 纯 Claude / 纯 GLM / Claude-leader + GLM-workers 混合会话 |
| `moai worktree <new\|list\|switch\|sync\|remove\|clean\|go>` | 面向并行 SPEC 开发的 Git worktree 管理 |
| `moai session <list\|register\|current>` | 多会话协调 |
| `moai spec <audit\|archive\|lint\|list\|new>` | SPEC 生命周期工具 |
| `moai goal <arm\|status\|clear>` | 目标引擎 CLI |
| `moai harness <status\|apply\|rollback\|disable>` | Harness 学习生命周期 |
| `moai handoff <save\|list>` | 会话交接记录 |
| `moai preference <list\|decay-scan\|toggle>` | 决策记忆管理 |
| `moai hook <event>` | Claude Code 钩子分发器 |
| `moai web` | Web 控制台 —— 设置 CRUD、SPEC 看板、代理配置 (en/ko/ja/zh) |
| `moai inventory` | 会话、worktree 与 Harness 的只读清单 (支持 `--json`) |
| `moai version` | 版本、提交哈希与构建日期 |

另有注册：`mx`、`clean`、`codemaps`、`feedback`、`loop`、`lsp`、`ast-grep`、`agent`、`workflow`、`statusline`、`telemetry`、`constitution`、`state`、`tool-policy`、`migrate`、`profile`、`pr`、`github`、`research`。

### 钩子

→ 详情：[钩子指南](https://adk.mo.ai.kr/zh/advanced/hooks-guide) · [钩子参考](https://adk.mo.ai.kr/zh/advanced/hooks-reference)

所有钩子事件都遵循 Claude Code hooks 协议，通过 JSON stdin/stdout 通信：

- **27 种事件类型** —— SessionStart、PreToolUse、PostToolUse、SessionEnd、Stop、SubagentStop、PreCompact、PostCompact、TeammateIdle、TaskCompleted 等
- **4 种钩子类型** —— command (shell 脚本)、prompt (LLM 评估)、agent (子代理验证)、http (webhook 端点)
- 任务指标捕获到 `.moai/logs/task-metrics.jsonl`，用于会话分析与成本跟踪

### 状态栏

→ 详情：[状态栏](https://adk.mo.ai.kr/zh/advanced/statusline)

MoAI 在 Claude Code 终端底部渲染丰富的状态栏：模型层级/effort、MoAI 版本 (含更新标记)、Git 分支与变更状态、上下文窗口用量 (CW%)、缓存命中率，以及会话成本/Token。

CW% 带有两阶段 `/clear` 标记——在模型特定阈值处的软警告 (Opus 4.8、GLM-5.2[1m] 等 1M 上下文模型为 50%；200K 模型为 90%)，以及在绝对上限处的硬标记。Claude Code 会把 GLM-5.2 误报为 200K 模型 (上游 Issue #653)；MoAI 在 `internal/statusline/memory.go` 中将其修正为 1M，因此请信任 MoAI 状态栏的 CW%。

### 输出风格

→ 详情：[进阶主题](https://adk.mo.ai.kr/zh/advanced)

| 风格 | 特点 | 受众 |
|-------|-----------|----------|
| **MoAI** (expert) | 密集、简洁 | 有经验的开发者 |
| **MoAI-Easy** (basic) | 友好、解释性 —— 产品默认值 | 新用户 |
| **MoAI-Learn** (learn) | 苏格拉底式导师 | 学习者 |

通过 `/config` 切换 (存储在优先级最高的 `settings.local.json` 中)。输出风格在会话开始时只读取一次——更改在 `/clear` 或新会话后生效。

### @MX 标签系统

→ 详情：[@MX 标签](https://adk.mo.ai.kr/zh/advanced/mx-tags)

@MX 标签是内联代码注解，在 AI 代理之间传递上下文、不变量契约与危险区域。

```go
// @MX:ANCHOR: [AUTO] Hook registry dispatch - 5+ callers
// @MX:REASON: [AUTO] Central entry point for all hook events, changes have wide impact
func DispatchHook(event string, data []byte) error {
    // ...
}
```

| 标签 | 用途 | 触发条件 |
|-----|---------|---------|
| `@MX:ANCHOR` | 不变量契约 | fan_in >= 3 —— 变更影响面广 |
| `@MX:WARN` | 危险区域 | goroutine、复杂度 >= 15、全局状态变更 |
| `@MX:NOTE` | 上下文 | 魔法常量、缺失文档、业务规则 |
| `@MX:TODO` | 未完成的工作 | 缺失测试、未实现的功能 |

该系统优化信噪比：**只有 AI 必须最先注意到的代码才有标签。** 大多数代码不符合任何标准且不带标签——这是正常且有意为之的。阈值和每文件上限在 `.moai/config/sections/mx.yaml` 中配置；用 `/moai mx --all` 扫描 (或 `--dry`、`--priority P1`)。

### Worktree 隔离

→ 详情：[Git Worktree](https://adk.mo.ai.kr/zh/worktree)

`/moai plan --worktree` 为每个 SPEC 提供隔离的 git worktree 以进行并行开发；`moai worktree` 管理其生命周期 (`new --tmux` 会在 worktree 内自动创建 tmux 会话)。

### 支持的 16 种语言

→ 详情：[CLI 参考](https://adk.mo.ai.kr/zh/cli-reference)

go · python · typescript · javascript · rust · java · kotlin · csharp · ruby · php · elixir · cpp · scala · r · flutter · swift —— 通过项目标记检测，每种语言运行各自的标准 lint/格式化/测试工具链。未安装的工具会被优雅跳过。

---

## 文档导航

完整文档位于 [adk.mo.ai.kr](https://adk.mo.ai.kr)，按以下 12 个章节组织 (中文版路径)：

| 章节 | 简介 |
|------|------|
| [入门指南](https://adk.mo.ai.kr/zh/getting-started) | 介绍、安装、Windows 指南、初始化向导、快速开始、CLI、FAQ |
| [核心概念](https://adk.mo.ai.kr/zh/core-concepts) | MoAI-ADK 是什么、宪章、Harness 工程、SPEC 开发、DDD、TRUST 5 (涵盖三大支柱) |
| [工作流命令](https://adk.mo.ai.kr/zh/workflow-commands) | plan、run、sync、project、harness、design |
| [实用命令](https://adk.mo.ai.kr/zh/utility-commands) | fix、loop、gate、mx、review、clean、codemaps、e2e、feedback、goal、moai |
| [CLI 参考](https://adk.mo.ai.kr/zh/cli-reference) | status、profile、doctor、inventory、update、web 等 36 个顶级命令的逐条参考 |
| [Claude Code 指南](https://adk.mo.ai.kr/zh/claude-code) | 基础、上下文与记忆、Agentic 模式、可扩展性 |
| [多 LLM](https://adk.mo.ai.kr/zh/multi-llm) | CG 模式、模型策略 |
| [成本优化](https://adk.mo.ai.kr/zh/cost-optimization) | 提示缓存 (Prompt Caching) |
| [指南](https://adk.mo.ai.kr/zh/guides) | CI 自治、多 LLM CI |
| [Git Worktree](https://adk.mo.ai.kr/zh/worktree) | 指南、示例、FAQ |
| [进阶](https://adk.mo.ai.kr/zh/advanced) | Tokenomics、Token 预算、状态栏、settings.json、钩子、技能、Harness v4、自我进化、决策记忆、安全笔记、目录系统等 |
| [贡献](https://adk.mo.ai.kr/zh/contributing) | 参与贡献本项目 |

---

## FAQ

### Q: 为什么不是每个函数都有 @MX 标签？

**这是正常的。** 标签只标记高扇入、复杂或危险的代码。每个项目中的大多数代码都不符合任何标签标准——未打标签的文件不是缺陷。

### Q: 状态栏里的版本指示器是什么意思？

```
🗿 v3.0.0-rc11 ⬆️ v3.0.0-rc12
```

第一个值是已安装的 MoAI-ADK 版本；箭头表示有可用更新 (运行 `moai update` 即可清除)。它与 Claude Code 自身的版本指示器是分开的。

### Q: Claude Code 询问 "Allow external CLAUDE.md file imports?"

请选择 **"No, disable external imports."** 你项目的 `.moai/config/sections/` 已经包含这些文件，项目级设置优先生效，而且禁用外部导入是更安全的选择，不会损失任何功能。

---

## 贡献

欢迎贡献！详细指南见 [CONTRIBUTING.md](CONTRIBUTING.md)。

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 编写测试 (新代码用 TDD，既有代码用特征测试)
4. 确保测试、lint 与格式化通过：`make test` · `make lint` · `make fmt`
5. 使用约定式提交消息并打开 pull request

**代码质量要求**：85%+ 覆盖率 · 0 lint 错误 · 0 类型错误 · 约定式提交

### 社区

- [Discord](https://discord.gg/Z7E7Mdc5aN) —— 实时讨论与技巧
- [Issues](https://github.com/modu-ai/moai-adk/issues) —— 缺陷报告、功能请求 (也可以在 Claude Code 内使用 `/moai feedback`)

---

## Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=modu-ai/moai-adk&type=date&legend=top-left)](https://www.star-history.com/#modu-ai/moai-adk&type=date&legend=top-left)

---

## 许可证

[Apache License 2.0](./LICENSE) —— 详见 LICENSE 文件。

## 链接

- [官方文档](https://adk.mo.ai.kr)
- [图书：Claude Code 实战 Agentic 编程](https://adk.mo.ai.kr/book)
- [CHANGELOG](./CHANGELOG.md)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code)
- [Discord 社区](https://discord.gg/Z7E7Mdc5aN)
