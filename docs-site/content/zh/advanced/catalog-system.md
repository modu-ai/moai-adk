---
title: 目录系统
weight: 80
draft: false
---

通过3级目录清单和slim init优化项目初始化。

## 概述

MoAI-ADK v2.15+的目录系统通过**3级清单**管理所有代理、技能、插件、规则。
通过 `moai init --slim` 仅部署项目所需的最小模板集，从而缩短初始化时间。

## 3级清单

| 层级 | 说明 | 部署标准 |
|------|------|----------|
| **Tier 1 (Core)** | 核心基础设施 — 编排器、质量门禁、基础技能 | 始终部署 |
| **Tier 2 (Standard)** | 标准扩展 — 各语言规则、框架技能 | 检测到项目语言/框架时 |
| **Tier 3 (Optional)** | 可选 — 领域技能、平台专属设置 | 显式请求或项目设置时 |

## 目录文件

目录清单以YAML格式定义：

```yaml
# 目录条目示例
- id: moai-workflow-tdd
  tier: 1                    # 1=Core, 2=Standard, 3=Optional
  type: skill
  path: .claude/skills/moai/workflows/tdd.md
  languages: []              # 空数组 = 全部语言
  frameworks: []
  hash: abc123...             # 内容哈希（完整性校验）
```

## SlimFS过滤器

`moai init --slim` 通过SlimFS过滤器限制部署文件：

```bash
# 完整安装（全部层级）
moai init my-project

# Slim安装（仅Tier 1 + 检测到的Tier 2）
moai init --slim my-project
```

### 过滤逻辑

1. Tier 1始终包含
2. 检测项目语言（Go、Python、TypeScript等）
3. 仅包含与检测到的语言对应的Tier 2条目
4. Tier 3被排除

## Typed Loader

`LoadCatalog()` 函数以类型安全的方式加载清单：

- 3级分类校验
- 哈希完整性检查（Hash Sentinel）
- 缺失字段检测
- 100%测试覆盖率

## 目录使用

### 项目初始化

```bash
# 常规初始化 — 部署所有模板
moai init my-project

# Slim初始化 — 仅部署最小模板集
moai init --slim my-project
```

### 更新

```bash
# 基于目录的更新
moai update                  # 更新所有层级
moai update --slim           # 以slim模式更新
```

## 相关文档

- [安装](/zh/getting-started/installation) — 安装指南
- [初始设置](/zh/getting-started/init-wizard) — init向导
- [更新](/zh/getting-started/update) — 更新指南
- [技能指南](/zh/advanced/skill-guide) — 技能编写指南
