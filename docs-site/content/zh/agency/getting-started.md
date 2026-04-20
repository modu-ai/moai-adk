---
title: 快速开始
weight: 20
draft: false
---
# 快速开始

## 前提条件

在开始使用 AI Agency 之前，请确保您已经：

- 安装了最新版本的 MoAI（v2.9.0 或更高）
- 安装了 Node.js 18 或更高版本
- 有一个 GitHub 账户（用于版本控制和部署）
- 安装了 Claude Code（v2.1.50 或更高）

## 你的第一个 AI Agency 项目

### 步骤 1：创建项目结构

```bash
mkdir my-landing-page && cd my-landing-page
moai init
```

这会创建基本的 Agency 项目结构：

```
my-landing-page/
├── .agency/
│   ├── brand.yaml          # 品牌定义
│   ├── rules/              # 进化规则库
│   │   ├── copy-rules.md
│   │   ├── design-rules.md
│   │   └── seo-rules.md
│   └── feedback/           # 反馈历史
├── .claude/                # Claude Code 配置
├── content/
│   ├── pages/              # 网页内容
│   ├── components/         # 可复用组件
│   └── assets/             # 图片和资源
├── site/                   # 生成的网站输出
└── briefing.md             # 项目简报
```

### 步骤 2：定义品牌上下文

编辑 `.agency/brand.yaml`：

```yaml
brand:
  name: "您的品牌名称"
  tagline: "品牌口号"
  
  colors:
    primary: "#0066FF"
    secondary: "#FF6B35"
    text: "#1A1A1A"
    background: "#FFFFFF"
  
  tone:
    - professional
    - friendly
    - innovative
  
  values:
    - "创新"
    - "可靠性"
    - "用户中心"
  
  target_audience:
    description: "您的目标用户描述"
    demographics: "年龄、行业等信息"
```

### 步骤 3：创建项目简报

编辑 `briefing.md`：

```markdown
# 项目简报

## 项目目标
描述该项目的主要目标和成功指标。

## 核心功能
列出该项目应包含的核心功能。

## 内容需求
- 主页：产品概览和调用行动
- 关于我们：公司背景和团队介绍
- 功能页面：详细的产品功能列表
- 定价页面：清晰的定价表
- 常见问题：FAQs
- 联系我们：联系表单

## 设计参考
链接到参考网站或设计灵感。

## 特殊要求
任何特定的需求或约束条件。
```

{{< callout type="info" >}}
简报越详细，AI Agency 生成的内容质量就越高。花 15-20 分钟仔细编写简报是值得的。
{{< /callout >}}

### 步骤 4：生成初稿

在 Claude Code 中运行：

```
/moai agency generate
```

AI Agency 将：
1. 分析您的品牌和简报
2. 生成页面框架和内容结构
3. 创建初始网站文件
4. 生成营销文案

### 步骤 5：评估和反馈

生成内容后：

1. **预览网站**：在本地服务器上查看生成的网站
2. **收集反馈**：记下您想改进的地方
3. **提交反馈**：使用以下命令提交反馈

```
/moai agency feedback
```

反馈可以包括：
- "标题太长，请简化"
- "这个颜色组合不符合我们的品牌"
- "需要添加更多的社会证明"
- "这个按钮文案不清楚"

### 步骤 6：进化和改进

系统将：
1. 分析您的反馈
2. 生成改进版本
3. 更新内部规则
4. 展示改进内容

这个过程会重复，直到您满意为止。

### 步骤 7：发布

当内容完成后：

```bash
# 构建最终网站
npm run build

# 本地测试
npm run preview

# 部署到 Vercel
moai agency deploy
```

## 目录结构详解

### `.agency/brand.yaml`
定义您品牌的所有视觉和语言特征。这是 FROZEN 区域的核心，确保所有输出的一致性。

### `.agency/rules/`
存储系统学到的规则。每次提交反馈时，系统会更新这些规则：
- `copy-rules.md`：文案风格规则
- `design-rules.md`：设计决策规则
- `seo-rules.md`：SEO 优化规则

### `content/pages/`
您的网页内容，通常以 MDX 格式存储，支持 React 组件。

### `content/components/`
可复用的网站组件：导航栏、页脚、卡片、按钮等。

### `site/`
生成的最终网站（自动生成，通常在 .gitignore 中）。

## 常见工作流程

### 快速迭代循环

```
1. 查看当前版本
2. 提交一条反馈
3. 系统改进
4. 查看改进结果
5. 重复（2-4 步）
```

### 大版本更新

当您需要显著改变时：

```bash
# 编辑简报
vim briefing.md

# 生成新版本（保留当前反馈历史）
/moai agency generate --preserve-rules

# 审核更改
# 提交反馈...
```

## 最佳实践

{{< callout type="info" >}}
遵循这些最佳实践以获得最佳效果：
{{< /callout >}}

1. **精确的简报** - 越具体越好，系统能更准确地理解您的需求
2. **及时反馈** - 尽快反馈，让系统学习您的偏好
3. **保持品牌一致** - 品牌定义一旦设定，应该保持稳定（FROZEN 区域）
4. **允许进化** - 相信系统的学习能力，不要在早期阶段过度干预
5. **测试内容** - 在发布前在多个设备和浏览器上测试

## 排查常见问题

### 生成的内容与品牌不符
- 检查 `.agency/brand.yaml` 中的定义是否准确
- 在简报中提供更清晰的品牌描述
- 检查反馈历史中是否有冲突的规则

### 某些页面缺失或不完整
- 确保 `briefing.md` 中清晰列出了所有需要的页面
- 生成后使用 `/moai agency feedback` 请求添加缺失的内容
- 检查项目日志了解生成过程中的任何错误

### 样式或布局看起来不正确
- 清除本地缓存：`rm -rf site/ node_modules/.cache`
- 重新构建：`npm run build`
- 检查浏览器的开发者工具查看任何 CSS 错误

## 下一步

- 继续阅读[代理 & 技能](./agents-and-skills)了解系统如何工作
- 查看[自我进化系统](./self-evolution)了解反馈如何改进系统
- 参考[命令参考](./command-reference)获取所有可用命令的完整列表
