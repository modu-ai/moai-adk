---
title: 数据库架构管理
description: 自动跟踪和管理架构、迁移和种子数据
weight: 15
draft: false
---

MoAI-ADK 的数据库工作流提供对项目数据库元数据的集中管理。使用 `/moai db` 命令，你可以扫描迁移文件、自动生成架构文档并检测漂移。

## 主要功能

- **交互式初始化** — 运行 `/moai db init` 选择数据库引擎、ORM 和迁移工具，然后自动生成元数据模板
- **自动同步** — PostToolUse 钩子自动检测迁移文件变化并刷新文档
- **漂移检测** — 使用 `/moai db verify` 检测架构文档和迁移文件之间的不一致
- **16 语言支持** — Go、Python、TypeScript、Rust、Java、Kotlin、C#、Ruby、PHP、Elixir、C++、Scala、R、Flutter、Swift

## 四个子命令

```bash
/moai db init      # 通过交互式访谈初始化数据库元数据
/moai db refresh   # 重新扫描迁移文件并重新生成架构文档
/moai db verify    # 检查漂移 (只读)
/moai db list      # 将所有表显示为 Markdown 表格
```

## 使用场景

- 为新项目设置数据库元数据
- 添加/编辑迁移文件后自动更新文档
- 与团队成员共享当前架构状态
- 验证架构文档和实际迁移状态的一致性

## 后续步骤

- **[入门](./getting-started.md)** — 运行 `/moai db init` 创建第一个迁移
- **[架构同步](./schema-sync.md)** — PostToolUse 钩子和自动刷新机制
- **[迁移模式](./migration-patterns.md)** — 16 种语言的默认迁移路径
- **[项目数据库目录](./project-db-directory.md)** — 7 文件模板集介绍

## 相关文档

有关详细信息，请参见 [/moai db 命令参考](../../reference/moai-db.md)。
