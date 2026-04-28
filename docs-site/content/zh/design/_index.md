---
title: 设计系统
description: 结合Claude Design和代码路径的混合设计工作流
weight: 60
draft: false
---

# 设计系统

MoAI-ADK的设计系统支持**混合方法**。选择Claude Design或代码设计来构建品牌一致的网络体验。

## 两条路径

```mermaid
flowchart TD
    A["用户请求<br>/moai design"] --> B["建立<br>品牌背景"]
    B --> C{选择路径}
    C -->|路径A| D["使用Claude Design<br>在claude.ai/design<br>中创建设计"]
    C -->|路径B| E["代码设计<br>brand-voice.md +<br>visual-identity.md"]
    D --> F["导出<br>handoff包"]
    E --> G["生成<br>文案+<br>设计令牌"]
    F --> H["解析和<br>转换包"]
    G --> H
    H --> I["expert-frontend<br>代码实现"]
    I --> J["GAN Loop<br>评估和迭代"]
    J --> K["Sprint Contract<br>基础完成"]
    K --> L["最终工件"]
```

## 主要特点

- **品牌一致性** — 品牌背景在每个阶段应用
- **Sprint Contract协议** — 每次迭代的明确接受标准
- **4维评分** — 设计质量、独创性、完整性、功能性
- **反AI-Slop规则** — 防止肤浅AI生成内容
- **无障碍合规** — WCAG AA标准自动验证

## 后续步骤

- **[入门指南](./getting-started.md)** — 使用/moai design启动第一个项目
- **[Claude Design交接](./claude-design-handoff.md)** — 了解Claude Design功能和包导出
- **[代码路径](./code-based-path.md)** — 使用brand-voice.md进行设计
- **[GAN Loop](./gan-loop.md)** — Builder-Evaluator迭代过程
- **[迁移指南](./migration-guide.md)** — 转换现有.agency/项目

## 要求

- 最新MoAI-ADK版本
- Claude Code桌面客户端v2.1.50或更高版本
- 路径A: Claude.ai Pro或更高订阅
- 路径B: 完整的品牌背景文件
