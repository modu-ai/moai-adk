---
title: 自主 CI/CD 指南
weight: 10
draft: false
---

通过MoAI-ADK的自主CI/CD系统自动管理Pull Request质量。

## 概述

由SPEC-V3R3-CI-AUTONOMY-001引入的自主CI/CD系统，是由8个层级构成的
质量自动化基础设施。从pre-push hook到auto-fix循环，无需开发者手动
验证质量，CI会自动保障质量。

## 8层架构

| 层级 | 名称 | 优先级 | 说明 |
|------|------|----------|------|
| T1 | Pre-push Hook | P0 | push前自动质量验证 |
| T2 | Branch Protection | P0 | main分支保护规则 |
| T3 | Auto-fix Loop | P1 | CI失败时自动修复 |
| T4 | Auxiliary Workflows | P2 | 辅助工作流整理 |
| T5 | Worktree State Guard | P1 | 保障工作树状态完整性 |
| T6 | i18n Validator | P2 | 验证4国语言文档一致性 |
| T7 | BODP | P0 | 分支起源决策协议 |
| T8 | Release Workflow | P1 | 发布自动化 |

## Pre-push Hook (T1)

在push前于本地自动执行质量验证。

```bash
# 自动安装（moai init / moai update时）
.git/hooks/pre-push → moai hook pre-push
```

执行的验证：

- `go vet` / `golangci-lint`（根据项目语言自动检测）
- `go test ./...`（测试套件）
- MX标签完整性检查

## Auto-fix Loop (T3)

CI失败时自动调用 `/moai loop` 修复错误。

```yaml
# .github/workflows/ci.yml (自动生成)
- name: Auto-fix on failure
  if: failure()
  run: |
    claude -p "/moai loop --max-iterations 3"
```

## BODP — 分支起源决策协议 (T7)

创建新分支/工作树时自动决定base分支。

### 3信号评估

| 信号 | 来源 | 含义 |
|--------|------|------|
| Signal A | SPEC `depends_on` + diff路径重叠 | 代码依赖关系 |
| Signal B | `git status` 中匹配 `.moai/specs/<NewSpecID>/` | 工作树同位置 |
| Signal C | `gh pr list --head <branch> --state open` ≥ 1 | 当前分支PR |

### 决策矩阵

| 信号 | 决策 |
|--------|------|
| 仅有A | `stacked` — 基于当前分支 |
| 有B | `continue` — 在当前上下文中继续 |
| 仅有C | `stacked` — 基于当前分支 |
| 无任何信号 | `main` — 基于origin/main |

### 审计追踪

所有BODP决策都记录在 `.moai/branches/decisions/<branch-name>.md` 中。

## i18n Validator (T6)

自动验证4国语言文档的一致性。

```bash
scripts/docs-i18n-check.sh
```

验证项目：

- 4个locale间文件数量/路径一致
- front matter `title` 存在
- H1 heading存在
- 遵守MoAI术语表

## Worktree State Guard (T5)

保障工作树的状态完整性：

- 检测未提交的更改
- 确认工作树与主分支的同步状态
- 在 `moai status` 中显示状态

## 相关文档

- [工作树指南](/zh/worktree/guide) — Git Worktree完整指南
- [/moai loop](/zh/utility-commands/moai-loop) — 反复修复循环
- [/moai fix](/zh/utility-commands/moai-fix) — 自动错误修复
- [多LLM CI](/zh/guides/multi-llm-ci) — Multi-LLM CI集成
