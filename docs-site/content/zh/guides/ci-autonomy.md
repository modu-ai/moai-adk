---
title: 自主 CI/CD 指南
weight: 10
draft: false
---

MoAI-ADK 的自主 CI/CD 系统自动管理拉取请求质量。它把本地会话中
`/moai loop` 执行的"诊断 → 修复 → 验证"循环延伸到了 CI，即使开发者不
手动验证质量，CI 也能自行保障质量 —— 这是把代理式循环工程应用到仓库
层级的案例。

## 概述

在 SPEC-V3R3-CI-AUTONOMY-001 中引入的自主 CI/CD 系统，是由 8 个层级组成
的质量自动化基础设施。从 push 前的本地验证 (pre-push hook) 到 CI 失败时
的自动修复 (auto-fix loop)，连成一道防线。

## 8-Tier 架构

| Tier | 名称 | 优先级 | 说明 |
|------|------|----------|------|
| T1 | Pre-push Hook | P0 | push 前自动质量验证 |
| T2 | Branch Protection | P0 | main 分支保护规则 |
| T3 | Auto-fix Loop | P1 | CI 失败时自动修复 |
| T4 | Auxiliary Workflows | P2 | 辅助工作流整理 |
| T5 | Worktree State Guard | P1 | 保障工作树状态完整性 |
| T6 | i18n Validator | P2 | 4 语言文档一致性验证 |
| T7 | BODP | P0 | 分支起点决策协议 |
| T8 | Release Workflow | P1 | 发布自动化 |

## Pre-push Hook (T1)

在 push 之前于本地自动执行质量验证。它是第一道防线，在本地提前切断
"到了 CI 才失败再折返"的往返成本。

```bash
# 自动安装 (moai init / moai update 时)
.git/hooks/pre-push → moai hook pre-push
```

执行的验证：

- `go vet` / `golangci-lint`（根据项目语言自动检测）
- `go test ./...`（测试套件）
- MX 标签完整性检查

## Auto-fix Loop (T3)

`/moai sync` 创建 PR 之后，编排器交接一个失败的必需 (required) 检查，
`manager-develop` 便以 `cycle_type=autofix` 周期运行"诊断 → 修复 →
重新验证"循环。它把本地的诊断式自我修复循环延伸到了 PR 流水线之上。

- **进入条件** — 只有当至少一个必需检查失败，且编排器指明待修复的 PR 与
  分支进行交接时，循环才会启动。编排器是唯一的进入点
- **迭代上限** — 每次 PR push 最多 3 次迭代。进入第 4 次时不再尝试修补，
  而是通过 blocking AskUserQuestion 上报给用户
- **语义层级失败** — data race、deadlock、panic 与测试断言失败一律不自动
  修补，交由人工判断
- **受保护文件** — 循环绝不修改机密、凭据文件与 CI 工作流定义，因为修补
  报告失败的那一层，会把真实失败变成虚假的 green

迭代上限、上报契约、语义层级失败处理与受保护文件清单的 SSoT 是
`.claude/rules/moai/workflow/ci-autofix-protocol.md`。

## BODP — Branch Origin Decision Protocol (T7)

创建新分支/工作树时自动决定 base branch。

### 3-Signal 评估

| 信号 | 来源 | 含义 |
|--------|------|------|
| Signal A | SPEC `depends_on` + diff path overlap | 代码依赖 |
| Signal B | `git status` 中匹配 `.moai/specs/<NewSpecID>/` | 工作树同位置 |
| Signal C | `gh pr list --head <branch> --state open` ≥ 1 | 当前分支 PR |

### 决策矩阵

| 信号 | 决策 |
|--------|------|
| 只有 A | `stacked` —— 基于当前分支 |
| 有 B | `continue` —— 在当前上下文继续 |
| 只有 C | `stacked` —— 基于当前分支 |
| 全部没有 | `main` —— 基于 origin/main |

### 审计追踪

所有 BODP 决策都记录在 `.moai/branches/decisions/<branch-name>.md` 中。
把决策留成记录而非猜测 —— MoAI 的"基于证据判定完成"原则同样适用于分支
决策。

## i18n Validator (T6)

自动验证 4 语言文档的一致性。

```bash
scripts/docs-i18n-check.sh
```

验证项目：

- 4 个 locale 之间的文件数量/路径一致
- front matter `title` 存在
- H1 heading 存在
- 遵守 MoAI 术语表

## Worktree State Guard (T5)

保障工作树的状态完整性：

- 检测未提交的变更
- 确认工作树与主分支的同步状态
- 在 `moai status` 中显示状态

## 相关文档

- [工作树指南](/zh/worktree/guide) —— Git Worktree 完整指南
- [/moai loop](/zh/utility-commands/moai-loop) —— 迭代修复循环
- [/moai fix](/zh/utility-commands/moai-fix) —— 自动错误修复
- [多 LLM CI](/zh/guides/multi-llm-ci) —— Multi-LLM CI 集成
