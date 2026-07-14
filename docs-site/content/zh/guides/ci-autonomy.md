---
title: 自主 CI/CD 指南
weight: 10
draft: false
---

MoAI-ADK 的自主 CI/CD 系统自动管理拉取请求质量。它把本地会话中
`/moai loop` 执行的"诊断 → 修复 → 验证"回路延伸到了 CI，即使开发者不
手动验证质量，CI 也能自行保障质量 —— 这是把代理式回路工程应用到仓库
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

`/moai sync` 创建 PR 之后，CI 监视脚本与 CI 回路技能一起运行"诊断 → 修复
→ 重新验证"回路。它把本地的诊断式自我修复回路延伸到了 PR 流水线之上。

**CI 监视脚本 (`scripts/ci-watch/run.sh`)**

```bash
sh scripts/ci-watch/run.sh <PR_NUMBER> [BRANCH]
```

- 以 30 秒间隔轮询 `gh pr checks`，区分必需 (required) 检查与
  辅助 (auxiliary) 检查
- 退出码：`0` 全部通过 · `2` 必需检查失败（向 stdout 输出结构化
  JSON 交接）· `3` 30 分钟硬超时 · `1` 错误
- 必需检查清单从 SSoT 文件读取，并支持用于测试的环境变量
  覆盖 (`MOAI_CIWATCH_GH`、`CIWATCH_TIMEOUT_SECONDS` 等)

**CI 回路技能 (`moai-workflow-ci-loop`)**

当监视脚本交接一个必需失败时，`moai-workflow-ci-loop` 技能会对失败
进行分类，并最多尝试 3 次安全的自动修补。语义层级的失败（自动修复有
风险的情形）会上报给用户。

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
- [/moai loop](/zh/utility-commands/moai-loop) —— 迭代修复回路
- [/moai fix](/zh/utility-commands/moai-fix) —— 自动错误修复
- [多 LLM CI](/zh/guides/multi-llm-ci) —— Multi-LLM CI 集成
