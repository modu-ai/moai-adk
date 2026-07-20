---
title: moai github GitHub 集成
weight: 92
draft: false
---

`moai github` 提供 GitHub 议题解析、SPEC 链接、工作流自动化命令。

公共标志接受 `--dry-run`(不做更改,仅显示将执行的内容)。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai github parse-issue <number>` | 解析 GitHub 议题并显示内容 |
| `moai github link-spec <issue-number> <spec-id>` | 将 GitHub 议题与 SPEC 文档双向链接 |

## moai github parse-issue

```bash
moai github parse-issue 123
```

解析 GitHub 议题并显示内容。

## moai github link-spec

```bash
moai github link-spec 123 SPEC-AUTH-001
```

在 GitHub 议题与 SPEC 文档之间创建双向链接。

## 示例

```bash
# 解析议题
moai github parse-issue 42

# 将议题链接到 SPEC(预览)
moai github link-spec 42 SPEC-AUTH-001 --dry-run

# 实际链接
moai github link-spec 42 SPEC-AUTH-001
```

## 相关文档

- [moai pr](/zh/cli-reference/pr) —— PR CI 监视
- [CLI 概览](/zh/getting-started/cli)
