---
title: Claude Design交接
description: Claude Design官方功能和handoff包使用方法
weight: 30
draft: false
---

# Claude Design交接

## Claude Design概述

**Claude Design**是Anthropic于2026年4月17日发布的**AI驱动设计生成工具**。在Claude.ai中用自然语言生成UI/UX设计。

- **基础模型:** Claude 3.5 Opus
- **访问:** https://claude.ai/design
- **输出格式:** 设计令牌、组件规范、静态资源、handoff包

## 支持的订阅计划

| 计划 | Claude Design支持 | 备注 |
|---|---|---|
| Free | 不支持 | 无法访问Claude.ai/design |
| Pro | 支持 | $20/月 |
| Max | 支持 | $200/月 |
| Team | 支持(默认关闭) | 按团队计费 |
| Enterprise | 支持(默认关闭) | 合同基础 |

**注意:** Team和Enterprise计划默认由管理员禁用。请要求团队管理员启用。

## 支持的输入格式

| 格式 | 说明 |
|---|---|
| **文本** | 用自然语言描述设计要求 |
| **图像** | 上传参考设计模型 |
| **DOCX/PPTX** | 现有文档或演示文稿 |
| **XLSX** | 数据表和结构化信息 |
| **网页截图** | 来自URL的网站屏幕截图 |
| **Figma** | Figma文件和框架 |
| **GitHub** | 仓库代码和README |

## 导出Handoff包

### 步骤1: 打开Claude.ai/design

在浏览器中打开**https://claude.ai/design**

### 步骤2: 创建设计

在Claude Design界面:
- 输入自然语言设计描述
- 上传参考图像/文档
- 实时UI预览生成

### 步骤3: 导出包

从Claude Design的Export或Share菜单:
- **ZIP格式:** 所有设计文件、令牌、资源
- **PDF格式:** 静态文档版本(可选)
- **Canva/Figma格式:** 在外部工具中继续编辑(可选)

**推荐:** 导出为ZIP格式

### 步骤4: 本地保存

将导出的文件保存到本地文件系统:

```bash
# 示例: ~/Downloads/my-design.zip
```

## 将包导入到MoAI-ADK

### 步骤5: 重新运行/moai design

在Claude Code中:

```
/moai design
```

选择路径A(Claude Design)后:

```
输入包路径: ~/Downloads/my-design.zip
```

### 步骤6: 自动转换

`moai-workflow-design-import`技能:
- 解析包
- 将设计令牌转换为JSON
- 提取组件规范
- 复制静态资源

结果文件:
```
.moai/design/
├── tokens.json          # 设计令牌
├── components.json      # 组件规范
└── assets/              # 图像、图标
```

## 包版本支持

| 版本 | 发布 | 状态 | 备注 |
|---|---|---|---|
| v1.0(初始) | 2026-04-17 | 支持 | 标准ZIP格式 |
| v1.1 | 2026-05-xx | 计划 | 扩展元数据 |
| v2.0(预览) | 将来 | 不支持 | 需要手动兼容性更新 |

## 错误代码

包导入失败时:

| 错误代码 | 原因 | 解决方案 |
|---|---|---|
| `DESIGN_IMPORT_NOT_FOUND` | 找不到文件 | 检查路径,验证文件存在 |
| `UNSUPPORTED_FORMAT` | 非ZIP格式 | 重新导出为ZIP格式 |
| `UNSUPPORTED_VERSION` | 不支持的版本 | 从最新Claude Design重新导出 |
| `SECURITY_REJECT` | 安全检查失败 | 联系管理员 |
| `MISSING_MANIFEST` | 包结构损坏 | 生成新包并重试 |

## 备用路径

当Claude Design不可用或包导入失败时:

### 选项1: 切换到路径B

```
包导入失败。切换到路径B?
→ [是] [否]
```

选择将自动切换到代码设计工作流

### 选项2: 重新生成包

返回Claude.ai/design:
1. 修改现有设计或创建新设计
2. 从Export菜单重新导出ZIP
3. 用新文件重新运行`/moai design`

## 团队协作

Claude DesignTeam订阅功能:

- **实时协作:** 多个团队成员同时编辑设计
- **共享链接:** 在团队外共享设计(只读)
- **版本历史:** 恢复以前的版本
- **评论:** 记录设计反馈

**注意:** 默认禁用;需要团队管理员启用

## 后续步骤

- 包导入后,参见[GAN Loop](./gan-loop.md)指南
- 代码实现期间,检查[Sprint Contract协议](./gan-loop.md#sprint-contract协议)
- 审查评分标准[4维评分](./gan-loop.md#4维评分)
