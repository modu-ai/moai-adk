---
title: moai pr PR 监视
weight: 96
draft: false
---

`moai pr` 是在 CI/CD 工作流中监视·管理 pull request 的命令。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai pr watch <PR_NUMBER>` | 监视 PR 的 CI 检查(或用 `--abort` 中止活动监视) |

## moai pr watch

```bash
moai pr watch 123
```

针对指定的 PR 编号监视 `gh pr checks`。

| 标志 | 说明 |
|--------|------|
| `--abort` | 中止活动的 CI 监视循环 |
| `--report` | 为该 PR 编号发出合并就绪报告 |
| `--branch <name>` | 用于报告上下文的分支名(默认:main) |

## 示例

```bash
# 监视 PR CI 检查
moai pr watch 42

# 中止活动监视
moai pr watch 42 --abort

# 合并就绪报告
moai pr watch 42 --report
```

## 相关文档

- [moai github](/zh/cli-reference/github)
- [CLI 概览](/zh/getting-started/cli)
