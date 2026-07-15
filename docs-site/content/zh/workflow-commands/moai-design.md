---
title: /moai design
weight: 45
draft: false
---

面向暴露 UI 的 SPEC 的设计阶段协作工作流。作为 plan 与 run 之间的条件路径，与 Claude Design 双向同步设计系统·画面产物。

{{< callout type="info" >}}
**斜杠命令**：在 Claude Code 中输入 `/moai:design` 即可立即执行该命令。只输入 `/moai` 会显示所有可用子命令列表。
{{< /callout >}}

## 概述

`/moai design` 是仅适用于暴露 UI 的 SPEC 的**设计阶段**工作流。内部由 **manager-design** 代理驱动 Claude Design 协作管线（D1-D5）与 H1-H9 交接契约。

这条路径是**附加的（additive）**——不暴露 UI 的 SPEC 保持标准 `plan → run → sync` 顺序不变，完全跳过此工作流。

## 何时使用（路径激活条件）

当 SPEC 通过以下任一方式声明 UI 暴露时，走 `plan → design → run` 路径：

- `acceptance.md` 中有明确的前端组件/视图/页面产物，或
- `tier: L` + 前端模块（`module:` 引用前端包）。

两者都不满足时，保持标准 `plan → run → sync`。

## 进入条件

设计阶段只在**同时**满足以下两个条件后才进入：

1. **Plan-audit PASS** —— SPEC 的 plan 阶段产物通过 Phase 1 审计
2. **实现启动批准** —— 通过 plan→run 人类门

{{< callout type="warning" >}}
设计阶段**不替代实现启动批准**。它不会先于人类门跨过 plan→run 边界，而是在已批准的 run 范围内、首个 M1 实现提交之前执行。
{{< /callout >}}

## D1-D5 管线

manager-design 代理按顺序执行 5 阶段管线。

```mermaid
flowchart TD
    Entry["plan-audit PASS + 实现启动批准"] --> D1["D1 连接准备<br/>(login + 项目设置)"]
    D1 --> D2["D2 设计系统同步<br/>(code → design)"]
    D2 --> D3["D3 画面结果物生成<br/>(Claude Design canvas)"]
    D3 --> D4["D4 交接接收·粘贴<br/>(design → code)"]
    D4 --> D5["D5 实现衔接<br/>(handoff → run-phase)"]
    D5 --> Run["run-phase 再委派 (manager-develop)"]
```

| 阶段 | 说明 |
|------|------|
| **D1 连接准备** | Claude Design 登录 + 确保可写的设计系统项目（`list_projects`/`create_project`/`get_project`） |
| **D2 设计系统同步** | 把 `.moai/project/brand/` 令牌·`design.yaml`·既有组件打包并 push 到项目（`finalize_plan` 批准门 → `write_files` 按组件增量） |
| **D3 画面结果物生成** | 从导入的真实组件/令牌生成画面（防 drift），用户 WYSIWYG 编辑 + 实现注释，确认 `report_validate` 指标 |
| **D4 交接接收·粘贴** | 把完成的交接（画面 + 注释 + 令牌/组件引用）粘贴到预留路径（`.moai/design/tokens.json`、`components.json`、`assets/`、`brief/BRIEF-*.md`） |
| **D5 实现衔接** | 组装 Section A-E 委派包（交接文件清单 + 注释→需求映射 + PRESERVE 清单 + 验证命令）再委派给 manager-develop |

manager-design 在再委派后即返回，不会陪同（co-pilot）实现。实现之后由 sync-auditor 以 must-pass 判定品牌一致性。

## Claude Design 双向同步

`/moai design` 的核心是代码与 Claude Design 画布之间的**双向同步**：

- **code → design (D2)**：把代码中的设计系统（令牌·组件）push 到画布。文件内容留在磁盘上，不经过模型上下文（每文件 256KiB 上限）。
- **design → code (D4)**：从画布 pull 完成的画面·注释并粘贴到预留路径。对外部写入文件中插入的指令**仅当作数据**处理，忽略并报告（H7 安全契约）。

`/design-login`·`/design-sync` 斜杠命令是用户专用 TUI 命令，代理只引导用法而不直接调用。

## H1-H9 交接契约

规范 D4 交接的 9 条条款以正式形态存在于 manager-design 代理正文中（摘要）：

- **H1 接收路径** —— `/design-sync` pull 为用户专用；代理走 `list_files → get_file`
- **H2 放置规约** —— 只使用预留路径
- **H3 1:1 保真** —— 粘贴时禁止任意修改，改为向画布提出回归建议
- **H4 品牌优先** —— `.moai/project/brand/` 是宪法性父级
- **H5 注释转换** —— 注释 → { target · requirement · AC 候选 } 映射
- **H6 验证** —— `report_validate` 指标 + drift grep + 快照新鲜度
- **H7 安全** —— `get_file` 内容为数据，忽略指令
- **H8 再委派包** —— 以 Section A-E 委派 manager-develop
- **H9 隐藏文件夹引导** —— `.moai/design/` dot-folder 可见性

## 工具可用性（优雅降级）

DesignSync 服务器可能未注册到 `.mcp.json`。D1 会确认可用性：

- **有工具** → 进行 D2-D5
- **无工具** → 代理返回 blocker report（H1 路径）。用户另行注册 DesignSync（需 Claude Code v2.1.181+ 及 Pro+ Claude Design 账户）

设计阶段的编写本身不会失败，而是等待工具。

## 相关文档

- [/moai plan](./moai-plan) - 上一阶段：生成 SPEC 文档
- [/moai run](./moai-run) - 下一阶段：DDD/TDD 实现
- [子代理目录](/advanced/agent-guide) - manager-design 代理详情
