---
title: 目录系统
weight: 80
draft: false
---

代币经济学并非只适用于代币的原则。分发到项目中的每一个模板文件，最终都是会话可能加载的上下文候选。目录系统以"只分发必要之物"为原则，从初始化阶段就开始削减这项成本。

## 概述

MoAI-ADK v2.15+ 的目录系统用 **3 层清单** 管理所有智能体、技能、插件与规则。使用 `moai init --slim` 时，只挑选项目所需的最小模板集进行分发，因此初始化更快，留在项目中的文件也更轻量。

## 3 层清单

所有分发对象都属于三个层级之一。

| 层级 | 说明 | 分发标准 |
|------|------|----------|
| **Tier 1 (Core)** | 核心基础设施 — 编排器、质量门禁、基础技能 | 始终分发 |
| **Tier 2 (Standard)** | 标准扩展 — 按语言的规则、框架技能 | 检测到项目语言/框架时 |
| **Tier 3 (Optional)** | 可选 — 领域技能、平台专属配置 | 显式请求或项目配置时 |

## 目录文件

目录清单以 YAML 格式定义。

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

`hash` 字段保存内容哈希，加载器可以据此验证已分发文件是否损坏或被任意修改。

## SlimFS 过滤器

`moai init --slim` 通过 SlimFS 过滤器限制分发文件。

```bash
# 完整安装（全部层级）
moai init my-project

# Slim 安装（仅 Tier 1 + 检测到的 Tier 2）
moai init --slim my-project
```

### 过滤逻辑

过滤器分四步运行。

1. Tier 1 始终包含
2. 检测项目语言（Go、Python、TypeScript 等）
3. 只包含与检测到的语言对应的 Tier 2 条目
4. 排除 Tier 3

## Typed Loader

`LoadCatalog()` 函数以类型安全的方式加载清单。它不依赖字符串解析，而是按结构体逐一校验，因此清单错误会在分发前被拦截。

- 3 层分类校验
- 哈希完整性检查 (Hash Sentinel)
- 缺失字段检测
- 100% 测试覆盖率

## 目录的应用

### 项目初始化

```bash
# 常规初始化 — 分发所有模板
moai init my-project

# Slim 初始化 — 仅分发最小模板集
moai init --slim my-project
```

### 更新

更新也基于同一份目录运行，因此以 slim 初始化的项目用 slim 方式更新即可。

```bash
# 基于目录的更新
moai update                  # 更新所有层级
moai update --slim           # 以 slim 模式更新
```

## 相关文档

- [安装](/zh/getting-started/installation) — 安装指南
- [初始设置](/zh/getting-started/init-wizard) — init 向导
- [更新](/zh/getting-started/update) — 更新指南
- [技能指南](/zh/advanced/skill-guide) — 技能编写指南
