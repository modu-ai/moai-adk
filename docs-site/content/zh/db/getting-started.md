---
title: 入门
description: 使用 /moai db init 初始化项目的数据库元数据
weight: 10
draft: false
---

只需一次 `/moai db init`，项目的数据库元数据就位。之后每次添加迁移文件，
架构文档都会随之更新，代理和团队成员始终能在同一位置看到最新架构。

## 前置要求

在开始数据库工作流之前，需要以下条件：

1. 由 `/moai project` 命令生成的 `.moai/project/product.md` 和 `.moai/project/tech.md` 文件
2. 受支持的数据库引擎（PostgreSQL, MySQL, SQLite, MongoDB 等）
3. ORM 或查询构建器（GORM, sqlc, Prisma, SQLAlchemy, ActiveRecord 等）
4. 迁移工具（golang-migrate, Flyway, Liquibase, Alembic 等）

## 分步初始化指南

### 第 1 步：确认项目元数据

先确认 `/moai project` 生成的必需文件是否存在：

```bash
ls -la .moai/project/
# 以下文件必须存在:
# - product.md
# - tech.md
# - structure.md
```

如果这些文件不存在，请先运行 `/moai project`。

### 第 2 步：初始化数据库元数据

现在运行 `/moai db init` 命令：

```bash
/moai db init
```

### 第 3 步：回答访谈问题

MoAI 会就以下 4 个项目进行交互式提问：

1. **数据库引擎** —— 正在使用的数据库（PostgreSQL, MySQL, SQLite, MongoDB 等）
2. **ORM/查询构建器** —— 数据访问层工具
3. **多租户策略** —— 单一架构、每租户一个架构、每租户一个 DB，或无
4. **迁移工具** —— 架构变更管理工具

针对每个问题选择合适的选项。

### 第 4 步：检查生成的文件

初始化后，`.moai/project/db/` 目录会生成以下文件：

```
.moai/project/db/
├── README.md              # DB 部分概述
├── schema.md              # 自动生成的表注册表
├── erd.mmd                # 实体关系图
├── migrations.md          # 迁移文件索引
├── rls-policies.md        # Row-level security 规则 (Supabase/Postgres)
├── queries.md             # 常用查询库
└── seed-data.md           # 种子数据模式
```

各文件的作用：

- `schema.md` —— 自动记录所有表、列、数据类型、约束
- `erd.mmd` —— 用 Mermaid 语法可视化表之间的关系
- `migrations.md` —— 已应用迁移文件的时间线
- `queries.md` —— 供 AI 代理参考的常用查询示例集

### 第 5 步：编写第一个迁移并同步

向项目添加新的迁移文件。例如 Go/golang-migrate 的情况：

```bash
# 在 db/migrations/ 目录创建迁移文件
touch db/migrations/001_create_users_table.sql
```

编写完迁移文件后，用以下命令更新架构文档：

```bash
/moai db refresh
```

这条命令会：
- 扫描所有迁移文件
- 向 schema.md 添加新表信息
- 更新 erd.mmd 图表
- 更新 migrations.md 时间线

### 第 6 步：验证漂移（可选）

要确认是否存在漂移：

```bash
/moai db verify
```

结果：

- `架构文档已同步` —— 迁移与文档一致
- 输出漂移报告 —— 详细显示差异（exit code: 1）

## 问题排查

### "Missing prerequisite files" 错误

若 `.moai/project/product.md` 和 `.moai/project/tech.md` 不存在：

```bash
/moai project
```

请先运行上述命令生成项目元数据。

### 迁移文件未被识别

确认项目的语言和迁移工具是否被正确检测：

```bash
cat .moai/config/sections/language.yaml
```

检查 `language` 字段，必要时可在 `.moai/config/sections/db.yaml` 中手动指定 `migration_patterns`。

### 自动同步不工作

确认 PostToolUse 钩子是否正确注册：

```bash
grep -A5 "PostToolUse" .claude/settings.json
```

如果没有钩子，请重新运行 `/moai db init` 或在 `.claude/settings.json` 中手动注册。
