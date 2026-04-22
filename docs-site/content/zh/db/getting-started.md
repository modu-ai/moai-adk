---
title: 入门
description: 使用 /moai db init 初始化项目的数据库元数据
weight: 10
draft: false
---

## 前置条件

开始数据库工作流前，你需要:

1. 由 `/moai project` 生成的 `.moai/project/product.md` 和 `.moai/project/tech.md` 文件
2. 支持的数据库引擎 (PostgreSQL、MySQL、SQLite、MongoDB 等)
3. ORM 或查询构建器 (GORM、sqlc、Prisma、SQLAlchemy、ActiveRecord 等)
4. 迁移工具 (golang-migrate、Flyway、Liquibase、Alembic 等)

## 分步初始化指南

### 步骤 1: 验证项目元数据

首先确认所需文件存在:

```bash
ls -la .moai/project/
# 这些文件应该存在:
# - product.md
# - tech.md
# - structure.md
```

如果这些文件不存在，请先运行 `/moai project`。

### 步骤 2: 初始化数据库元数据

现在运行 `/moai db init` 命令:

```bash
/moai db init
```

### 步骤 3: 回答访谈问题

MoAI 会提出 4 个交互式问题:

1. **数据库引擎** — 你的数据库 (PostgreSQL、MySQL、SQLite、MongoDB 等)
2. **ORM/查询构建器** — 数据访问层工具
3. **多租户策略** — 单一架构、租户级架构、租户级数据库或无
4. **迁移工具** — 架构变更管理工具

为每个问题选择适当的选项。

### 步骤 4: 检查生成的文件

初始化后，这些文件在 `.moai/project/db/` 中创建:

```
.moai/project/db/
├── README.md              # 数据库部分概述
├── schema.md              # 自动生成的表注册表
├── erd.mmd                # 实体关系图
├── migrations.md          # 迁移文件索引
├── rls-policies.md        # 行级安全规则 (Supabase/Postgres)
├── queries.md             # 常见查询库
└── seed-data.md           # 种子数据模式
```

文件说明:

- `schema.md` — 自动文档化所有表、列、数据类型和约束
- `erd.mmd` — 使用 Mermaid 语法可视化表关系
- `migrations.md` — 应用的迁移时间线
- `queries.md` — AI 代理可参考的常见查询示例

### 步骤 5: 编写第一个迁移并同步

向项目添加新的迁移文件。例如，使用 Go/golang-migrate:

```bash
# 在 db/migrations/ 中创建迁移文件
touch db/migrations/001_create_users_table.sql
```

编写迁移后，刷新架构文档:

```bash
/moai db refresh
```

此命令:
- 扫描所有迁移文件
- 向 schema.md 添加新表信息
- 更新 erd.mmd 图表
- 刷新 migrations.md 时间线

### 步骤 6: 验证漂移 (可选)

检查是否有任何漂移:

```bash
/moai db verify
```

结果:

- `架构文档已同步` — 迁移和文档匹配
- 漂移报告输出 — 显示详细差异 (退出代码: 1)

## 故障排除

### "Missing prerequisite files" 错误

如果 `.moai/project/product.md` 和 `.moai/project/tech.md` 不存在:

```bash
/moai project
```

首先运行此命令生成项目元数据。

### 未识别迁移文件

检查项目的语言和迁移工具是否正确检测:

```bash
cat .moai/config/sections/language.yaml
```

检查 `language` 字段。如果需要，可以在 `.moai/config/sections/db.yaml` 中手动指定 `migration_patterns`。

### 自动同步不工作

验证 PostToolUse 钩子是否正确注册:

```bash
grep -A5 "PostToolUse" .claude/settings.json
```

如果找不到钩子，请再次运行 `/moai db init` 或在 `.claude/settings.json` 中手动注册。
