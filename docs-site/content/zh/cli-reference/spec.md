---
title: moai spec 文档管理
weight: 35
draft: false
---

`moai spec` 管理 `.moai/specs/` 目录中的 SPEC 文档。它提供状态更新、漂移检测、验收标准查看、EARS/GEARS lint、原子化关闭、era 审计、归档的子命令。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai spec status` | 更新或列出 SPEC 状态 |
| `moai spec drift` | 检测 frontmatter status 与 git log 之间的漂移 |
| `moai spec view <SPEC-ID>` | 以树状结构查看验收标准 |
| `moai spec lint [spec.md...]` | lint EARS 合规性与结构有效性 |
| `moai spec close <SPEC-ID>` | 原子化 4-phase 关闭(status: completed + progress.md backfill) |
| `moai spec audit` | SPEC era 分类与 modern-era 状态漂移审计 |
| `moai spec archive` | 将已关闭的 SPEC 归档至 `.moai/specs/` 之外 |

## moai spec status

```bash
moai spec status <SPEC-ID> <new-status>   # 更新状态
moai spec status --list                   # 列出全部 SPEC
moai spec status --sync-git               # 从 git log 同步状态
```

| 标志 | 说明 |
|--------|------|
| `--dry-run` | 不写入,预览变更 |
| `--list` | 列出全部 SPEC 及状态 |
| `--sync-git` | 从 main 的 git log 同步 SPEC 状态 |
| `--yes` | `--sync-git` 非交互式自动确认(CI/管道必需) |

## moai spec drift

```bash
moai spec drift
```

| 标志 | 说明 |
|--------|------|
| `--json` | JSON 格式输出 |
| `--exit-code-on-drift` | 检测到漂移时返回退出码 1 |
| `--count` | 仅输出漂移数量 |
| `--no-cache` | 绕过 HEAD-SHA 结果缓存后重新计算 |

## moai spec lint

```bash
moai spec lint [spec.md...]
```

| 标志 | 说明 |
|--------|------|
| `--json` | JSON 格式输出 |
| `--sarif` | SARIF 2.1.0 格式输出 |
| `--strict` | 将警告视为错误 |
| `--format <fmt>` | 输出格式(table) |

## moai spec close

```bash
moai spec close SPEC-ID
```

以单次提交将 SPEC 原子化转换为 `status: completed`。

| 标志 | 说明 |
|--------|------|
| `--backfill-only` | 仅执行 progress.md backfill |
| `--dry-run` | 不提交,预览 |
| `--force` | 不确认强制关闭 |
| `--json` | JSON 格式输出 |

## moai spec audit

```bash
moai spec audit
```

扫描 `.moai/specs/SPEC-*/`,用 era 启发式对每个 SPEC 分类并检测 modern-era 状态漂移。

| 标志 | 说明 |
|--------|------|
| `--json` | JSON 格式输出 |
| `--filter-era <era>` | 按 era 过滤 |
| `--filter-spec <id>` | 按 SPEC ID 过滤 |
| `--include-grandfathered` | 包含 grandfather era SPEC |
| `--strict` | 严格模式 |

## moai spec archive

```bash
moai spec archive --dry-run   # 确认目标(不移动)
moai spec archive --yes       # 应用计划
```

归档超过 grace 窗口(默认 90 天)的 terminal SPEC。

| 标志 | 说明 |
|--------|------|
| `--dry-run` | 不移动,报告目标集合 |
| `--yes` | 确认移动(应用时必需) |
| `--grace-days <n>` | grace 窗口天数(0 = 默认值 90) |
| `--json` | 以 JSON 输出计划 |

## 相关文档

- [基于 SPEC 的开发](/zh/core-concepts/spec-based-dev)
- [CLI 概览](/zh/getting-started/cli)
