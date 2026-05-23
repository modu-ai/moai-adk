---
title: 入门指南
description: 使用/moai design命令启动混合设计工作流
weight: 20
draft: false
---

# 入门指南

## 前置条件

- MoAI-ADK项目初始化完成
- `.moai/project/brand/`目录准备或待创建
- Claude Code桌面客户端v2.1.50或更高版本

## 品牌背景采访

运行`/moai design`时首先启动**品牌背景采访**。

### 创建三个品牌文件

采访在`.moai/project/brand/`中生成这些文件:

1. **brand-voice.md** — 品牌语调、术语、消息指导
2. **visual-identity.md** — 颜色、排版、视觉语言
3. **target-audience.md** — 客户档案和偏好

### 采访过程

1. 在Claude Code中运行`/moai design`
2. 看到"品牌背景不完整"消息
3. 选择品牌采访
4. `manager-spec`代理提问
5. 自由形式回答
6. 三个文件自动生成

示例问题:
- "您的品牌语调是专业的还是亲切的?"
- "选择您的三种主要品牌颜色"
- "您的目标客户的主要问题是什么?"

## 路径选择

品牌背景设置后,显示路径选择UI:

### 选项1(推荐) — 使用Claude Design

在**Claude.ai/design**中生成设计,然后导出为**handoff包**

**要求:**
- Claude.ai Pro、Max、Team或Enterprise订阅

**优势:**
- 直观的UI/UX
- 实时协作(Team订阅)
- 支持多种输入格式(文本、图像、Figma、GitHub)

### 选项2 — 代码设计

使用**moai-domain-copywriting**和**moai-domain-brand-design**技能

**要求:**
- 完整的`brand-voice.md`和`visual-identity.md`

**优势:**
- 无需额外订阅
- 自动设计令牌生成
- 版本控制友好

## 首次运行

```bash
# 在Claude Code中运行
/moai design
```

执行顺序:
1. 检查`.agency/`(显示迁移指南)
2. 验证品牌背景
3. 采访缺失文件
4. 显示路径选择UI
5. 执行选择路径工作流

## 验证设置

检查生成的品牌文件:

```bash
ls -la .moai/project/brand/
# brand-voice.md
# visual-identity.md
# target-audience.md
```

## 后续步骤

- **选择路径A:** 参见[Claude Design交接](./claude-design-handoff.md)指南
- **选择路径B:** 参见[代码路径](./code-based-path.md)指南

## 故障排除

### 更新品牌背景

直接编辑文件:

```bash
# 用您喜欢的编辑器编辑
vim .moai/project/brand/brand-voice.md
```

更改会在下一次`/moai design`运行时自动应用。

### 重新运行采访

```bash
# 备份当前文件,然后重新采访
mv .moai/project/brand .moai/project/brand.backup
/moai design
```
