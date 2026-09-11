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
| `moai spec lint [SPEC-ID | path/to/spec.md | SPEC directory ...]` | lint EARS 合规性与结构有效性 |
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
moai spec lint [SPEC-ID | path/to/spec.md | SPEC directory ...]
```

| 标志 | 说明 |
|--------|------|
| `--json` | JSON 格式输出 |
| `--sarif` | SARIF 2.1.0 格式输出 |
| `--strict` | 将警告视为错误 |
| `--format <fmt>` | 输出格式(table) |

| 参数形式 | 示例 |
|--------|------|
| SPEC-ID | `SPEC-SPC-001` — 解析为项目根目录下的 `.moai/specs/SPEC-SPC-001/spec.md`(与 `moai spec view` 规则相同) |
| 文件路径 | `.moai/specs/SPEC-SPC-001/spec.md` |
| SPEC 目录 | `.moai/specs/SPEC-SPC-001` — 读取其中的 `spec.md` |

三种形式可在一次调用中混用;不带参数时仍按原方式扫描整个语料库。看起来像 SPEC-ID 却解析不到文件的参数,不是对文档的指摘,而是 **参数错误(退出码 3)**,并会一并给出尝试过的路径。

此处会产生两个 advisory 级别的警告。`ModalityUnjudged` 会报告 linter 无法判定的需求,而不是静默略过。`REQTableRowsRejected` 会报告未被读作 REQ 定义的表格行。两者均为 advisory:`--strict` 不会将其提升为错误,也不改变退出码。

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
