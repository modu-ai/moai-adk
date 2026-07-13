---
title: 数据库架构管理
description: 自动跟踪并管理架构、迁移和种子数据的工作流
weight: 70
draft: false
---

MoAI-ADK 的数据库工作流对项目的架构元数据进行集中管理。使用 `/moai db`
命令扫描迁移文件、自动生成架构文档，并检测文档与实际迁移之间的漂移。

这份架构文档不只是给人看的文档。从**代理式挂具** (Agentic Harness) 的角度
看，`.moai/project/db/` 是代理的持久上下文（基于文件的记忆）。代理无需在
每个会话中重新读取迁移文件来重建架构，而是参考一份整理好的架构文档，因此
获取同样信息所耗费的 Token 大幅减少 —— 这是托克诺米克斯延伸到文档结构的
一个实例。

## 主要功能

- **交互式初始化** —— 通过 `/moai db init` 选择数据库引擎、ORM、迁移工具，并自动生成元数据模板
- **自动同步** —— PostToolUse 钩子检测迁移文件变更并自动刷新
- **漂移检测** —— 通过 `/moai db verify` 检查架构文档与迁移文件之间的不一致
- **支持 16 种语言** —— Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift

## 4 个子命令

```bash
/moai db init      # 通过交互式访谈初始化 DB 元数据
/moai db refresh   # 重新扫描迁移文件并重新生成架构文档
/moai db verify    # 检查漂移（只读）
/moai db list      # 以 Markdown 表格显示所有表
```

## 何时使用

- 新项目启动时设置数据库元数据
- 添加/编辑迁移文件后自动更新文档
- 向团队成员分享当前架构状态
- 验证架构文档与实际迁移状态之间的一致性

## 下一步

- **[入门](./getting-started.md)** —— 运行 `/moai db init` 与第一个迁移
- **[架构同步](./schema-sync.md)** —— PostToolUse 钩子与自动刷新机制
- **[迁移模式](./migration-patterns.md)** —— 16 种语言的默认迁移路径
- **[项目 DB 目录](./project-db-directory.md)** —— 7 个文件模板集介绍

## 相关文档

更多详情请参考 [/moai db 命令指南](/zh/db/getting-started)。
