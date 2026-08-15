<p align="center">
  <img src="./assets/images/moai-adk-og.png" alt="MoAI-ADK" width="100%">
</p>

<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>Claude Code 的智能体开发框架 —— 把成本、自我改进、质量控制三件事一起管住</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> ·
  <a href="./README.ko.md">한국어</a> ·
  <a href="./README.ja.md">日本語</a> ·
  中文
</p>

<p align="center">
  <a href="https://book.mo.ai.kr" target="_blank"><strong>官方图书《Claude Code 实战 Agentic 编程》</strong></a><br>
  MoAI-ADK 作者亲写的 Harness 工程实战指南 — <a href="https://book.mo.ai.kr" target="_blank">book.mo.ai.kr</a>
</p>

<p align="center">
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
  <a href="https://codecov.io/gh/modu-ai/moai-adk"><img src="https://codecov.io/gh/modu-ai/moai-adk/branch/main/graph/badge.svg" alt="Codecov"></a>
  <br>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.1.0-blue.svg" alt="Release"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
</p>

<p align="center">
  <a href="https://adk.mo.ai.kr"><strong>官方文档</strong></a> ·
  <a href="https://adk.mo.ai.kr/book">图书：Claude Code 实战 Agentic 编程</a>
</p>

---

> **"模型是一个逐 Token 移动的随机工作者。它每一轮都无法记住自己该花多少成本、工作质量如何、或者上个会话中断在哪里。框架从外部强制执行这三件事。"**

---

## MoAI-ADK: 具备三大核心的智能体框架

MoAI-ADK（Agentic Development Kit）让 Claude Code 生成代码，再让这些代码以可预测的成本变得可靠，并走上持续改进的轨道。框架是从外部包裹模型的系统。模型是逐 Token 移动的随机工作者 —— 每一轮都不记得预算、质量标准、以及上个会话中断在哪里。成本上限、通过的测试套件、不断积累的学习循环、跨越 `/clear` 的连续性 —— 这些属性无法靠每轮重新提示来植入，必须由系统从外部强制执行。

需要的正是这三件事。MoAI-ADK 用这三者一起包裹 Claude Code —— 而不只是其中一个：

- **🪙 成本** —— 代币经济学：用更少的 Token 达到同样的质量，用同样的 Token 换取更高的质量。
- **🧠 自我改进** —— 智能体循环工程：框架越跑越好，把观察变成规则。
- **🛡️ 质量控制** —— 智能体框架：SPEC 生命周期、TRUST 5 门控、以及防止返工（最大的 Token 浪费）的隔离。

它不取代 Claude Code。它只是用结构包裹 Claude Code 留给用户自行处理的部分 —— 模型路由、质量门、成本控制、学习循环、会话连续性。用 Go 编写的单一二进制文件，在 macOS、Linux、Windows 上零依赖即可运行。

<p align="center">
  <img src="./assets/images/why-harness-infographic-zh.png" alt="面向 Claude Code 的智能体开发框架 —— 从外部包裹模型的结构" width="85%">
</p>

---

## v3.1 新功能 —— 看板模式

一个会话只有一个上下文窗口。长 SPEC 会把它填满，之后的每一步都背着前面的全部内容前进：早已用不上的计划在评审时仍占着窗口，而那次评审在写文档时还留在那里。常见的出路是 `/clear`，但它把包袱和上下文一起丢掉了。

看板模式把一项工作拆到**5 个终端而不是 1 个**。主导会话推动整条链，4 个伴随会话各自负责 `plan`、`run`、`review`、`sync` 中的一列，**只背自己那一列的上下文**。这并不是取消上限 —— 每个会话的上限依旧存在。真正改变的是：没有任何一个会话再扛四个阶段的历史，因此同样的预算能走得远得多，做完的阶段也可以在不丢卡片的前提下清空。

<p align="center">
  <img src="./assets/images/kanban-five-sessions.png" alt="看板模式的一次 run —— 主导会话与四个伴随会话各自在自己的终端中，使用各自的模型与推理强度运行" width="100%">
</p>

每一列都可以用不同的后端和推理强度。上图中 Plan 跑在 Opus 5 high，Run 跑在 GLM 5.2 xhigh，Sync 跑在 GLM 5.2 —— 因为各列需要的推理深度本就不同。

### 快速上手

```bash
moai cc -k                          # 主导 —— announce run-id 并铺好链
moai cc -k --name plan-<run-id>     # 伴随会话，各开一个终端
moai cc -k --name run-<run-id>
moai cc -k --name review-<run-id>
moai cc -k --name sync-<run-id>
```

伴随会话要**逐个新开终端、手动启动**。会话不会替你启动另一个会话。任意一列把 `moai cc` 换成 `moai glm`，那一列就跑在 GLM 后端。

看板有 `backlog → plan → run → review → sync → done` 六列。`backlog` 刻意不设归属会话，所以只有人主动放入时，工作才会进入看板。

```text
/moai todo "rename 提示已经过时"   # 添加卡片
/moai todo                          # 查看队列
```

有两条规则让看板保持诚实。主导只依据自己从卡片 `progress.md` 中**读到的证据**推进卡片 —— 不依据伴随会话的回复，因为回复是主张而非观测，而且跨会话送达本身也没有保证。阶段结束后，主导会请你对相应会话执行 `/clear`：`/clear` 由人亲手输入，无法作为指令发送。

### 用眼睛看板

`moai web` 会启动一个本地控制台。看板页面把 5 会话链与 SPEC 流水线并排展示，另有 Overview、Specs、Monitor、Settings 页面。

<p align="center">
  <img src="./assets/images/moai-web-overview.png" alt="moai web 控制台 Overview 页面 —— SPEC 统计、进行中 SPEC 列表、会话注册表" width="90%">
</p>

详细说明：[看板模式](https://adk.mo.ai.kr/zh/advanced/kanban-mode) · [`/moai todo`](https://adk.mo.ai.kr/zh/utility-commands/moai-todo)

---

## 为什么这三者必须同时具备

只优化成本是一个陷阱。只顾压低成本，质量会悄然下滑，紧接着是返工和调试循环 —— 而返工是所有 Token 支出中最贵的。只立质量门而没有学习循环，每个会话都会重犯同样的错误。跑没有成本上限的自主循环，一个失控任务就会耗尽额度。三者相互支撑：**成本因质量防止返工而保持经济，质量因循环捕获有效模式而保持可强制，循环因成本门在超额前停止而保持可负担。**

MoAI-ADK 的每一个设计决策都指向这三者中的一个。用哪个模型、推理多深、如何消耗上下文 —— 这些都不再逐轮听天由命，而是由系统决定，并记录决策让下一次运行更聪明。

<p align="center">
  <img src="./assets/images/three-axes-infographic-zh.png" alt="MoAI-ADK 的三大核心 — 代币经济学 · 智能体循环 · 智能体框架" width="90%">
</p>

---

## 🪙 成本 —— 代币经济学

Token 单价三年内跌了 **98%**（Linux Foundation），但同期企业 AI 支出反而涨了 **320%**。使用量暴增盖过了单价下降。智能体为解决单个任务要跑几十到几百步，按比例烧掉 token — 按量计费下这直接变成账单，订阅制下则是吃掉所有模型共享的周配额。

Uber 给 5,000 名工程师部署了 Claude Code，**四个月烧完一年的编码预算**，随后被迫施加月度 token 限额。Meta、Amazon、Microsoft 相继收回无限制 AI 政策。**代币经济学** — 把模型匹配到任务以提升 token 效率 — 成为科技行业的新基线。

传统的成本控制是为单价上涨而设计的，面对"单价在跌但总支出在涨"这个悖论便束手无策。瓶颈不在单价，而在使用量 — 更准确地说，是智能体收尾前要跑多少步。

**成本由分配决定，而非单价。** DeepSWE 排行榜（113 tasks，按 effort 分级视图）的实测数据展示了这一点。在同一个 Claude 系列内，单任务成本取决于模型*完成*任务的效率 — 而非一个 token 的价格。

| 模型 [effort] | Pass@1 | 单任务成本 | 输出 token | 步数 |
|---|---|---|---|---|
| claude-opus-5 [low] | 58% | **$1.66** | 20k | 36 |
| claude-opus-5 [medium] | 69% | $3.29 | 37k | 52 |
| claude-opus-5 [high] | 73% | $6.08 | 64k | 73 |
| claude-opus-5 [max] | 74% | $11.84 | 118k | 99 |
| claude-sonnet-5 [max] | 54% | **$26.40** | 214k | 268 |

Opus 5 在**最低** effort 下的分数高于 Sonnet 5 在**最高** effort 下的分数（58% vs 54%），而单任务成本只有其十六分之一（$1.66 vs $26.40）—— 尽管 Sonnet 的单 Token 价格更低。原因是 268 步对 36 步：写出账单的是重试循环，而不是 Token 费率。"用弱模型跑得更狠就能省钱"的直觉不成立。成本由**为任务分配合适的模型和推理深度**决定，而非单价。

#### 四个阶段：测量 → 路由 → 减脂 → 防御

<p align="center">
  <img src="./assets/images/why-tokenomics-infographic-zh.png" alt="代币经济学悖论 — 单价 98%↓，成本 320%↑" width="80%">
</p>

### 路由 —— 为每个任务分配合适的模型和推理深度

<p align="center">
  <img src="./assets/images/model-routing-infographic-zh.png" alt="智能体模型路由 — 11 个智能体分配到合适的模型与 effort" width="85%">
</p>

**Tier×Phase 矩阵**。根据工作阶段（plan / run / sync）和 SPEC 大小（Tier S / M / L）声明式分配模型和推理深度（effort）。需要深度推理的计划阶段分配高推理模型，机械重复较多的实现阶段分配轻量模型，最大化成本质量比。

**No-Haiku 3 层策略**。将 Haiku 从路由模型集中排除，工作分散到贴合任务性质的 3 层结构。Sonnet 以 low effort 承担单次完成、以输入为主的工作（Git 机械操作、只读检索）以最小化步数；Opus 承担所有多轮代理式行，`max` effort 只保留给两个调用频率最低的行。

**配置矩阵**。单一的 per-agent 配置矩阵将 12 个保留 agent 各自映射到一个 `{model, effort}` 对 —— 共 36 格。配置只有一个 —— `high` / `medium`（默认）/ `low`，通过 `llm.profile`（`moai init --profile`、`moai update --profile`）选择 —— 选取活动列；`moai model profile` 解析每个 agent 的格。包括 `Explore` 在内的每个保留 agent 都从矩阵获取 model+effort（任何位置都没有 Haiku）；只有用户自定义 agent 继承会话模型。

<p align="center">
  <img src="./assets/images/cg-mode-infographic-zh.png" alt="CG 模式 — Claude 领队 + GLM 执行者混合" width="85%">
</p>

### DeepSWE 基准测试 —— 性价比拐点在哪里

| 模型 [effort] | 分数 | 单任务成本 | 备注 |
|---|---|---|---|
| opus-5 [low] | 58%±2 | $1.66 | |
| opus-5 [medium] | **69%±1** | **$3.29** | **性价比拐点** |
| opus-5 [high] | 73%±2 | $6.08 | 分数 +4pt，成本 1.8 倍 |
| opus-5 [xhigh] | 73%±3 | $9.07 | **纯损失** — 与 high 同分，成本却 +49% |
| opus-5 [max] | 74%±4 | $11.84 | |
| glm-5.2 [max] | 44%±2 | $3.92 | API 按量计费下劣势 · z.ai 定额订阅下有价值 |
| sonnet-5 [max] | 54%±4 | $26.40 | 被 opus-5 [low] 帕累托支配 |

![DeepSWE 基准 — 模型×effort 的分数与单任务成本](./assets/images/deepswe-benchmark-2.png)

> 数据来源：[DeepSWE v1.1 排行榜](https://deepswe.datacurve.ai)（datacurve.ai，113 tasks，2026-07-25）

### 验证成本与预算管理 —— 让上下文更轻，超支前先停

**降低验证成本 —— 上下文更轻，证据落到磁盘。** 把冗长的验证输出重定向到磁盘文件，上下文中只保留退出码和 bounded tail（最多 50 行）。Prompt 缓存复用（缓存读取成本 0.1×）加上上下文减脂的 `/clear` 策略（1M 模型 50% / 200K 模型 90% 阈值时自动推荐）让窗口保持轻盈。

**预算管理 —— 超支前先停，下个会话接着来。** Token Circuit Breaker 在硬上限（默认 90%）时中止，把进度保存到 `progress.md`，并发布可粘贴的 resume 消息。Statusline 始终把上下文使用率、缓存命中率、rate limit 耗尽率显示在眼前。

---

## 🧠 自我改进 —— 智能体循环工程

不重复上个会话错误的会话最便宜。自我改进把每次运行变成下一次运行的素材：路由决策和门控证据被记录下来，重复出现的模式升格为规则，而声明的 goal 让会话一直工作到条件满足为止。

**`/moai goal` · `/moai loop`**. 声明一个完成条件，会话就会自动运行直到满足或达到回合限制（默认 30）。`--max-turns 0` 可启用自动压缩驱动的无限 goal，由 `--max-duration` 和停滞防护兜底。`/moai loop` 并行扫描 LSP 诊断 · AST-grep · linter，按问题级别分桶，队列空尽为止。

**Routing Ledger**. 将路由决策和门控证据记录为隐私保护摘要。观察升格为规则。

**4 层学习阶梯**. 观察（≥1）→ 启发式（≥3）→ 规则（≥5）→ 自动更新（≥10，需用户批准）；信任度下限 0.70。所有应用可通过 `moai harness rollback` 回滚。框架编辑（规则/智能体/钩子修改）遵循预测–验证纪律：每次编辑记录一条可证伪的预测，须通过 held-in/held-out 双重检查方可采纳，被否决的编辑也会留档。

**决策记忆**. 问题在不确定性最高处（p ≈ 0.5）出现；推荐跟随观察到的统计多数，而非系统默认。

---

## 🛡️ 质量控制 —— 智能体框架

返工是最大的 Token 浪费 —— 一个出货后返回的 bug，比所有路由优化加起来都贵。质量控制让"完成"意味着*经验证的完成*，并隔离工作，让并行智能体不会相互践踏。

### SPEC 3 阶段生命周期

plan → run → sync。Tier S/M/L 大小分类决定验证深度和 PR 路由。GEARS 格式要求 + 验收标准按证据判定完成。

<p align="center">
  <img src="./assets/images/spec-3phase-infographic-zh.png" alt="SPEC 三阶段工作流 — 规划 → 执行 → 同步" width="80%">
</p>

**TRUST 5 质量门**. Tested（85%+ 覆盖率）· Readable · Unified · Secured · Trackable，应用于所有变更。门控判定验证，而非智能体自判。

**12-Agent 目录**. MoAI 自定义 11 个 + 内置 Explore。规划和审计从一开始分离，编写方不能给自己的工作打分。

### 扩展点 —— 复用已验证模式做项目定制

**Harness v4 Builder**. 自然语言请求 → 领域 · 目标 · 约束提取 → 批准门控 → 项目专用 agent · skill · command · hook 脚手架。

**@MX 标签**. AI Agent 之间交换上下文、不变契约、危险区的内联代码注释。

**worktree 隔离**. 为每个 SPEC 准备独立的工作树。用 `moai cc -w <名称>` 进入，加上 `--spawn` 则在保留当前会话的同时于新窗口中打开。

---

## 基础设施支撑这三者

零依赖运行于 macOS、Linux、Windows 的 Go 单一二进制文件，是这三者（而不仅是代币经济学）的共同基座。钩子系统机械地强制门控，状态栏实时显示成本与上下文，SPEC 生命周期让工作跨越 `/clear` 仍可续接。三者都跑在同一个二进制上 —— 没有哪一个是事后补丁。

---

## 快速开始

### 安装

#### macOS / Linux / WSL

```bash
curl -fsSL https://adk.mo.ai.kr/install.sh | bash
```

#### Windows (PowerShell 7.x+)

```powershell
irm https://adk.mo.ai.kr/install.ps1 | iex
```

#### 从源码构建（Go 1.26+）

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk && make build
```

### 项目初始化

```bash
moai init my-project
```

交互式向导自动检测语言、框架、方法论，选择模型策略，并生成 Claude Code 集成文件。

### 第一个工作流

```bash
claude        # launch Claude Code inside the project
```

```text
/moai plan "Add JWT login"      # 编写 SPEC
/moai run SPEC-AUTH-001         # TDD/DDD 实现
/moai sync SPEC-AUTH-001        # 同步文档 + 创建 PR
```

自然语言也可以。`/moai "fix the login bug"` 会触发意图分析（Analyze-First 路由），读取请求并路由到合适的工作流。

### 系统要求

| 平台 | 支持环境 | 备注 |
|----------|----------------------|-------|
| macOS | Terminal, iTerm2 | 完全支持 |
| Linux | Bash, Zsh | 完全支持 |
| Windows | **WSL（推荐）**, PowerShell 7.x+ | 原生 cmd.exe 不支持 |

**前置要求**

- 所有平台必须安装 **Git**
- **Claude Code** — MoAI-ADK 是 Claude Code 的框架
- **推荐**：`gh` CLI（PR 自动化）· `tmux`（CG 模式）· 语言的 lint/test 工具链（例如 `golangci-lint`）

---

## 参考

### /moai 斜杠命令（16 个）

| 子命令 | 职责 |
|------------|------|
| `plan` / `run` / `sync` | SPEC 3 阶段流水线 |
| `project` / `harness` | 项目文档生成 · Harness 生命周期 |
| `goal` / `loop` / `fix` | 声明式 goal 循环 · 迭代修复 · 单次修复 |
| `review` / `gate` / `clean` | 代码审查（`--deep` 多智能体对抗式漏洞扫描） · pre-commit 质量门控 · 死代码移除 |
| `mx` / `codemaps` / `feedback` | @MX 注解 · 架构文档 · GitHub issue 报告 |
| `e2e` / `todo` | 多平台 E2E 测试（Web/移动/桌面，CLI 优先） · 看板 backlog 队列 |
| *（自然语言）* | Analyze-First 路由：自主 plan → run → sync 流水线 |

> **已退役（Retired）的 4 个子命令**：`design` · `brain` · `coverage` · `security`（SPEC-SUBCOMMAND-RETIRE-001，status: completed）。`security` 由 `moai-ref-owasp-checklist` + `moai-ref-llm-security` 技能替代，`e2e` 经 E2E-REVIVAL 复活为现役。

> → 详情：[Workflow Commands](https://adk.mo.ai.kr/zh/workflow-commands) · [Utility Commands](https://adk.mo.ai.kr/zh/utility-commands)

### CLI 命令（常用 13 个）

| 命令 | 说明 |
|---------|-------------|
| `moai init` | 交互式项目设置（自动检测语言/框架/方法论） |
| `moai doctor` | 系统状态诊断和环境验证 |
| `moai status` | 项目状态摘要（Git 分支、质量指标） |
| `moai update` | 更新到最新版本（支持自动回滚） |
| `moai cc` / `moai glm` / `moai cg` | Claude 专用 / GLM 专用 / 混合 Claude 领导 + GLM worker 会话 |
| `moai worktree <sync\|done\|remove\|clean\|recover\|snapshot\|verify\|restore>` | Git worktree 维护（进入 worktree 由启动器负责） |
| `moai session <list\|register\|current>` | 多会话协调 |
| `moai spec <audit\|archive\|lint\|list\|new>` | SPEC 生命周期工具 |
| `moai goal <arm\|status\|clear>` | Goal 引擎 CLI |
| `moai harness <status\|apply\|rollback\|disable>` | Harness 学习生命周期 |
| `moai handoff <save\|list>` | 会话交接记录 |
| `moai preference <list\|decay-scan\|toggle>` | 决策记忆管理 |
| `moai web` | Web 控制台 —— 5 个页面（Overview · Kanban · Specs · Monitor · Settings）、10 个设置标签 |

> 完整命令列表：[CLI Reference](https://adk.mo.ai.kr/zh/cli-reference)

### MCP 服务器

`moai init` 默认只配置**一个**活跃 MCP 条目——自托管的 `moai mcp-server`（一个本地 stdio 服务器）。它跨五个组暴露 17 个 MoAI 专用工具。四个 documented-but-disabled 条目（`context7`、`chrome-devtools`、`playwright`、`ast-grep`）通过 `moai mcp add <名称>` 激活。通用的 `moai mcp add|remove|list` CLI 通过 atomic-RWM seam 管理条目——用户绝不手编 `.mcp.json`。

| 组 | 工具 | 用途 |
|-------|-------|---------|
| SPEC lifecycle | `spec_progress`, `spec_audit`, `spec_drift` | 时代分类 + 漂移检测 |
| Verification | `verify_snapshot`, `verify_trend` | 按键证据快照 |
| Goal + session | `goal_arm`, `goal_status`, `session_list` | 自治循环 + 多会话协调 |
| Cross-model audit | `audit_multi`, `codex_audit`, `glm_audit`, `audit_cache` | 多审计者收敛 |
| Codex delegation | `codex_task`, `codex_setup`, `codex_job_*` | 后台跨模型任务 |

所有后端都 fail-open：GLM（`~/.moai/.env.glm`）和 codex（`~/.codex/auth.json`）是可选的——不可用的后端返回 `inconclusive`，绝不会引发 hard error。

> 详情：[MCP 服务器指南](https://adk.mo.ai.kr/zh/guides/mcp-server) · [Claude Code MCP](https://adk.mo.ai.kr/zh/claude-code/extensibility/mcp)

### 12-Agent 目录

| 分类 | Agent | 成本 | 职责 |
|----------|-------|------|------|
| **Manager** | manager-spec | 🔴 | Plan-phase SPEC 编写 |
| | manager-develop | 🔴 | Run-phase TDD/DDD/autofix 实现 |
| | manager-docs | 🔵 | Sync-phase 文档化 |
| | manager-git | 🩵 | PR 创建和路由 |
| | manager-design | 🟠 | Design-phase 协作（Claude Design） |
| | manager-kanban | 🔴 | 层级团队 Tier L 协调（唯一 Agent-carrier，depth-2 封闭） |
| **Evaluator** | plan-auditor | 🔴 | 独立计划审计（偏见防止） |
| | sync-auditor | 🔴 | 4 维质量评分（Functionality 40 · Security 25 · Craft 20 · Consistency 15） |
| **Builder** | builder-harness | 🟠 | 项目专用 agent·skill·command·hook 脚手架 |
| **Advisor** | super-advisor | 🔵 | 按需高推理咨询（E1-E4 升级） |
| **Specialist** | e2e-tester | 🟠 | Web/移动/桌面 E2E 测试执行（CLI 优先） |
| **Built-in** | Explore | ⚪ | 只读代码库探索 |

成本颜色以默认 `medium` profile 的 model×effort 单元为准（用 `moai model profile` 查看）：🔴 opus+high · 🟠 opus+medium · 🔵 opus+low · 🩵 sonnet+low · ⚪ 继承会话模型（用户添加的 agent）。切换 profile（`high`/`low`）时分配会变化。长期委托的进度记录在 Task 通道，由编排器以图标 Progress Board 转达。

### TRUST 5 质量门

| 准则 | 含义 | 验证 |
|-----------|---------|------------|
| **T**ested | 已测试 | 85%+ 覆盖率、特性化测试、单元测试通过 |
| **R**eadable | 可读 | 清晰命名、一致风格、lint 错误 0 |
| **U**nified | 统一 | 一致格式、import 顺序、项目结构合规 |
| **S**ecured | 安全 | OWASP 合规、输入验证、安全警告 0 |
| **T**rackable | 可追溯 | Conventional commits、issue 引用、结构化日志 |

### 方法论选择（TDD vs DDD）

```mermaid
flowchart TD
    A["Project analysis"] --> B{"New project or<br/>10%+ test coverage?"}
    B -->|"Yes"| C["TDD (default)"]
    B -->|"No"| D["DDD"]
    C --> F["RED → GREEN → REFACTOR"]
    D --> G["ANALYZE → PRESERVE → IMPROVE"]
```

| 方法论 | 周期 | 对象 |
|-------------|-------|-----|
| **TDD**（默认） | RED → GREEN → REFACTOR | 新项目和功能工作 |
| **DDD** | ANALYZE → PRESERVE → IMPROVE | 覆盖率 <10% 的现有代码 |

### 看板模式

`--kanban`（简写 `-k`）是一个会话启动器开关，武装一个 `kanban_chain` 目标预设——以多会话看板协调驱动单个 SPEC 走完 `plan → run → verify → sync`。看板的骨架是 **Origin-Trail Chain**：一个 append-only JSONL 谱系树，追踪 worktree 祖先，解决深度遗忘（`/clear` 后的根到叶链恢复），并通过心跳过期检测 dead leader 会话。

| 概念 | 作用 |
|---------|-------------|
| Origin-Trail Chain | `.moai/state/chain/events.jsonl` 处的 append-only JSONL 事件流 |
| WorktreeNode（13 字段） | 每会话状态：ID、父节点、深度、origin 链、里程碑、恢复目标 |
| CWD 冲突解决 | `(worktree_path, session_id)` 对消除复用路径的歧义 |
| 深度上限 | 限制嵌套复杂度 |

> **现已可用**：`moai cc -k`（或 `moai glm -k`）启动主导会话，`-k --name <role>-<run-id>` 逐个加入伴随会话 —— 每个终端一个，由人手动启动。`moai chain <status|lineage|back|list|prune>` 读取谱系，`moai todo <add|list|next|done>` 操作 `backlog` 列。启动顺序见上文“v3.1 新功能 —— 看板模式”一节。

> 详情：[看板模式指南](https://adk.mo.ai.kr/zh/advanced/kanban-mode)

---

## 读取 Statusline

```
🤖 Opus │ 🧠 xhigh·t │ ♻️ 87% │ 🔅 v2.1.212 │ 🗿 v3.0.1 │ ⏳ 2h 34m │ 💬 MoAI
🪫 CW: ████████░░ 88% (⚠️/clear) │ 🔋 5H: ████░░░░░░ 45% (4h 30m) │ 🪫 7D: ████████░░ 82% (Jan 21)
📁 moai-adk-go │ 🔀 modu-ai/moai-adk | 🅱️ feat/statusline ↑2 +3 │ 💾 +1 M2 ?0 │ 📋 [run SPEC-AUTH-001-run] │ 💌 PR #1042 (⌥approved)
```

| 元素 | 含义 |
|------|------|
| 🤖 模型 | 当前激活模型 |
| 🧠 effort | 推理努力等级 — 扩展思考启用时 `·t` 后缀 |
| ♻️ 缓存命中率 | Prompt 缓存命中率 |
| CW: 上下文 | 上下文窗口使用率 + 2 阶段 `/clear` 标记（⚠️ 软，🛑 硬） |
| 5H / 7D | 计费方案使用率 + 重置时间 |
| 📁 目录 | 项目目录名 |
| 🔀 仓库 | GitHub 仓库 identity `owner/name` |
| 🅱️ 分支 | 当前分支 + `↑`ahead `↓`behind + `+`dirty 计数 |
| 💾 git 状态 | staged / modified / untracked 计数 |
| 📋 任务 | 活跃 SPEC 工作流 `[命令 SPEC-ID-阶段]` |
| 💌 PR | 活跃 GitHub PR 编号 + 审查状态（`⌥state`） |

> 详情：[Statusline 指南](https://adk.mo.ai.kr/zh/advanced/statusline)

---

## Claude × GLM 多模型协同

MoAI-ADK 支持把 **z.ai GLM** 作为 Claude Code 的替代后端。切换只需改环境变量，不必改动任何代码；框架、SPEC 工作流和质量门在每一种后端上的行为完全一致。

| 项目 | 说明 |
|---|---|
| GLM Coding Plan | 每月 **$10** 起（[订阅](https://z.ai/subscribe?ic=1NDV03BGWU)） |
| 兼容性 | 与 Claude Code 直接对接 —— 无需改代码 |
| 模型 | glm-5.3、glm-4.7、glm-4.5-air，另有免费模型 |

### 三种执行模式

| 命令 | 领队 | 执行者 | tmux | 成本节省 | 适用场景 |
|---|---|---|---|---|---|
| `moai cc` | Claude | Claude | 不需要 | — | 追求最高质量的复杂工作 |
| `moai glm` | GLM | GLM | 建议 | 约 70% | 成本优化 |
| `moai cg` | Claude | GLM | **必需** | 约 60% | 质量与成本兼顾 |

**CG 模式**是两者的混合：Claude 领队负责策略、规划和审计，GLM 执行者承担大量实现工作，二者通过 tmux 会话级的环境隔离衔接。

```bash
moai glm sk-your-glm-api-key   # 保存密钥，一次即可
moai cg                        # 进入 CG 模式（Claude 领队 + GLM 执行者）
```

### 默认模型映射

每个 Claude 层级通过 `ANTHROPIC_DEFAULT_*_MODEL` 环境变量映射到对应的 GLM 模型：

| Claude 层级 | GLM 模型 | 上下文 |
|---|---|---|
| Opus | glm-5.3 | 1M |
| Sonnet | glm-5.3 | 1M |
| Haiku | glm-5.3 | 1M |
| Fable | glm-5.3 | 1M |

> 另有免费模型可用（GLM-4.7-Flash、GLM-4.5-Flash）。完整列表见 [z.ai 定价](https://docs.z.ai/guides/overview/pricing)。
>
> → 详情：[Multi-LLM 指南](https://adk.mo.ai.kr/zh/multi-llm)

---

## FAQ

### Q: 为什么不是每个函数都有 @MX 标签？

正常。标签只标记高扇入、复杂或危险的代码。任何项目中大部分代码都不符合任何标签阈值，没有标签的文件不是缺陷。

### Q: Statusline 版本显示是什么意思？

```
🗿 v3.0.1 ⬆️ v3.0.2
```

第一个值是当前安装的 MoAI-ADK 版本；箭头表示有可用更新。运行 `moai update` 后消失。

### Q: 不用 GLM 只用 Claude 可以吗？

可以。`moai cc` 启动 Claude 专用会话。CG 模式（`moai cg`，Claude 领导 + GLM Worker）和 GLM 专用（`moai glm`）是成本节省选项；框架 · SPEC 工作流 · 质量门控在所有三种模式中完全相同。

### Q: 适用于现有项目吗？

适用。`moai init` 检测项目状态并选择方法论 — 对覆盖率 <10% 的现有代码使用 DDD（特性化测试固定行为后渐进改进），对新/充分测试的代码使用 TDD。

---

## 社区与文档

### 贡献

随时欢迎贡献。详细流程见 [CONTRIBUTING.md](CONTRIBUTING.md)。

1. Fork 仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 编写测试（新代码用 TDD，现有代码用特性化测试）
4. 验证测试、lint、format 通过：`make test` · `make lint` · `make fmt`
5. 使用 Conventional commit 提交并打开 PR

**代码质量要求**：85%+ 覆盖率 · lint 错误 0 · 类型错误 0 · Conventional commits

### 社区

- [Issues](https://github.com/modu-ai/moai-adk/issues) — 错误报告、功能请求（Claude Code 内使用 `/moai feedback`）

### 许可证

[Apache License 2.0](./LICENSE) — 详情见 LICENSE 文件。

### 文档指南

[adk.mo.ai.kr](https://adk.mo.ai.kr) 在线文档分为 12 个章节。

| 章节 | 说明 |
|---------|-------------|
| [Getting Started](https://adk.mo.ai.kr/zh/getting-started) | 介绍、安装、Windows 指南、init 向导、快速开始、CLI 概览、FAQ |
| [Core Concepts](https://adk.mo.ai.kr/zh/core-concepts) | MoAI-ADK 同一性、宪法、harness 工程、SPEC 基于开发、DDD、TRUST 5 |
| [Workflow Commands](https://adk.mo.ai.kr/zh/workflow-commands) | `plan` · `run` · `sync` — SPEC 流水线主干 |
| [Utility Commands](https://adk.mo.ai.kr/zh/utility-commands) | `fix` · `loop` · `gate` · `review` · `clean` · `codemaps` · `e2e` · `feedback` · `goal` |
| [CLI Reference](https://adk.mo.ai.kr/zh/cli-reference) | `moai` 二进制所有命令 — `status`, `profile`, `doctor`, `update`, `web`, `goal`, `handoff`, `harness`, `init`, `worktree` 等 |
| [Claude Code Guide](https://adk.mo.ai.kr/zh/claude-code) | Claude Code 集成 — 基础、上下文·记忆、agentic、扩展性（skill·hook·plugin） |
| [Multi-LLM](https://adk.mo.ai.kr/zh/multi-llm) | CG 模式和模型策略 |
| [Cost Optimization](https://adk.mo.ai.kr/zh/cost-optimization) | Prompt 缓存策略和 token 成本降低 |
| [Guides](https://adk.mo.ai.kr/zh/guides) | CI 自动化、multi-LLM CI 等实战运营配方 |
| [Git Worktree](https://adk.mo.ai.kr/zh/worktree) | 并行 SPEC 开发用 worktree 指南、示例、FAQ |
| [Advanced](https://adk.mo.ai.kr/zh/advanced) | 代币经济学概述、Token 预算、statusline、settings.json、hook、@MX 标签、skill 指南、Harness v4 Builder、自进化、决策记忆、目录系统、安全笔记、CLAUDE.md/agent 指南 |
| [Contributing](https://adk.mo.ai.kr/zh/contributing) | 开源贡献指南 |

### 链接

- [官方文档](https://adk.mo.ai.kr)
- [图书：Claude Code 实战 Agentic 编程](https://adk.mo.ai.kr/book)
- [CHANGELOG](./CHANGELOG.md)
- [Claude Code](https://code.claude.com/docs/en)

---

## Star 历史

<a href="https://www.star-history.com/?type=date&repos=modu-ai%2Fmoai-adk">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=modu-ai/moai-adk&type=date&theme=dark&legend=top-left&sealed_token=9wFuBO5GMKxHZsaknxlIW3oypXLJlyW1qqq8T--aTRyfp6j9EK9KTR2vJvyAG8AKSs3Lindw7LUt-m-I6ysz9BoV6kdtrKlJYTViQAYR56A_3ie4ZVOqIw" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=modu-ai/moai-adk&type=date&legend=top-left&sealed_token=9wFuBO5GMKxHZsaknxlIW3oypXLJlyW1qqq8T--aTRyfp6j9EK9KTR2vJvyAG8AKSs3Lindw7LUt-m-I6ysz9BoV6kdtrKlJYTViQAYR56A_3ie4ZVOqIw" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=modu-ai/moai-adk&type=date&legend=top-left&sealed_token=9wFuBO5GMKxHZsaknxlIW3oypXLJlyW1qqq8T--aTRyfp6j9EK9KTR2vJvyAG8AKSs3Lindw7LUt-m-I6ysz9BoV6kdtrKlJYTViQAYR56A_3ie4ZVOqIw" />
 </picture>
</a>

<p align="center">
  <sub>由 MoAI-ADK 团队打造 · <a href="https://adk.mo.ai.kr">adk.mo.ai.kr</a></sub>
</p>
