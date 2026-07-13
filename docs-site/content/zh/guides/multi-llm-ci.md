---
title: "Multi-LLM CI 指南"
description: "在 GitHub Actions 中用多个 AI 模型自动化代码审查"
draft: false
weight: 10
---

本文介绍如何用 MoAI-ADK 的 Multi-LLM CI 功能在 GitHub Actions 中设置多个
LLM 代码审查。审查者也没有理由被绑在一个模型上 —— 各模型的强项和单价
不同，因此多 LLM 分配的托克诺米克斯视角同样适用于 PR 审查。

## 概述

### 什么是 Multi-LLM CI？

MoAI-ADK 的 Multi-LLM CI 功能提供在 GitHub Actions 中由多个 AI 模型同时执行代码审查的集成 CI/CD 流水线。

### 支持的 LLM

| LLM | 提供方 | 触发方式 | 特点 |
|-----|--------|-------------|------|
| **Claude** | Anthropic | `/claude` 评论 | Issue/PR 审查, OAuth 认证 |
| **Codex** | OpenAI | PR open 自动 | 仅限私有仓库 |
| **Gemini** | Google | PR open 自动 | API Key 认证 |
| **GLM** | Zhipu AI | PR open 自动 | Token 认证 |

## 快速开始

### 前置要求

- 已安装 MoAI-ADK（macOS · Linux · Windows）
- GitHub repository
- 各 LLM 账户及 API Token

### 初始设置

```bash
moai github init
```

这条命令执行的工作：
- 创建 `.github/workflows/` 目录
- 部署 workflow 模板
- 部署 composite actions
- GitHub Secrets 设置指南

### LLM 认证设置

```bash
# Claude (OAuth)
moai github auth claude

# Codex（私有仓库）
moai github auth codex

# Gemini
moai github auth gemini

# GLM
moai github auth glm
```

### 设置 GitHub Secrets

各 LLM 所需的 Secrets：
- `CLAUDE_CODE_OAUTH_TOKEN` - Claude OAuth Token
- `CODEX_AUTH_JSON` - Codex 认证 JSON（base64 编码）
- `GEMINI_API_KEY` - Gemini API Key
- `GLM_API_KEY` - GLM API Token

### 第一个 PR 测试

创建 PR 后会自动添加 LLM Panel 评论：

```markdown
## LLM Code Review Status

| LLM | Status |
|-----|--------|
| Claude | Pending (add `/claude` comment) |
| Codex | ✓ Ready |
| Gemini | ⚠️ Token missing |
| GLM | ✓ Ready |

Trigger individual reviews:
- Add `/claude` comment to trigger Claude
- Add `/codex` comment to trigger Codex
- Add `/gemini` comment to trigger Gemini
- Add `/glm` comment to trigger GLM
```

## LLM 认证设置

### Claude 设置

#### 获取 OAuth Token

1. 安装 [Claude Code](https://claude.ai/download)
2. 登录后获取 OAuth Token
3. 自动保存到 `.claude/settings.local.json`

#### moai github auth claude

```bash
moai github auth claude
```

**交互式设置流程：**
```
找不到 Claude OAuth Token。
是否安装 Claude Code 并登录? (y/n): y

[已确认] OAuth Token 已保存到 settings.local.json。
请将以下值设置到 GitHub Secret: CLAUDE_CODE_OAUTH_TOKEN:
<token-value>
```

### Codex 设置（仅限私有仓库）

#### 生成认证 JSON

```json
{
  "token": "sk-...",
  "base_url": "https://api.openai.com/v1"
}
```

#### moai github auth codex

```bash
moai github auth codex
```

**交互式设置：**
```
OpenAI auth.json 文件路径: ~/.codex/auth.json
读取文件并生成 GitHub Secret...
注意: Codex 仅可在私有仓库中使用 (REQ-SEC-001)

生成的 Secret:
CODEX_AUTH_JSON=eyJ0...
```

### Gemini 设置

```bash
moai github auth gemini
```

输入 API Key 后自动提供 GitHub Secret 设置指南。

### GLM 设置

```bash
moai github auth glm
```

从 GLM Token 路径（`~/.moai/.env.glm`）自动读取。

## 理解 Workflow 模板

### llm-panel.yml

**触发：** PR opened

**作用：** 自动生成以可视化方式显示各 LLM 状态的面板评论

**备注：** 用 `/claude`、`/codex`、`/gemini`、`/glm` 评论触发单独审查

### claude.yml / claude-code-review.yml

- **claude.yml**: Issue 触发（草稿审查）
- **claude-code-review.yml**: PR 触发（变更审查）

**特点：** 仅通过 `/claude` 评论触发

### codex-review.yml

**安全约束：**
- 仅在 `private` 仓库中运行 (REQ-SEC-001)
- 通过 `visibility` 检查阻止公开仓库

**workflow：**
```yaml
private-guard:
  runs-on: ubuntu-latest
  steps:
    - name: Check Repository Visibility
      run: |
        if [[ "${{ github.repository_visibility }}" == "public" ]]; then
          echo "::error::Codex review is restricted to private repositories"
          exit 1
        fi
```

### gemini-review.yml

- 自动语言检测（detect-language action）
- PR synchronized 时自动触发

### glm-review.yml

- GLM 专用环境设置（setup-glm-env action）
- 自动注入环境变量

### Composite Actions

#### detect-language

**输入：** repository 根路径
**输出：** language 环境变量（`detected_language`）

**支持语言：** Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift（16 种）

#### setup-glm-env

设置 GLM 团队模式所需的环境变量：
- `ANTHROPIC_AUTH_TOKEN` (GLM endpoint)
- `ANTHROPIC_BASE_URL` (https://glm.modu-ai.kr)

## 高级设置

### 自定义 github-actions.yaml

#### 基本结构

```yaml
# .moai/config/sections/github-actions.yaml
llm_review:
  enabled: true
  runners:
    claude: true
    codex: true
    gemini: true
    glm: true
  triggers:
    on_pr_open: true
    on_comment:
      claude: "/claude"
      codex: "/codex"
      gemini: "/gemini"
      glm: "/glm"
```

#### 按语言分配 LLM

```yaml
language_rules:
  go:
    - gemini
    - claude
  python:
    - claude
    - glm
  typescript:
    - codex
    - claude
```

### Runner 版本管理

#### 检查自动更新

```bash
moai github status
```

**输出示例：**
```
✓ GitHub Actions Runner
  Version: 2.700.1 (10 days old)
  Status: OK

⚠️ Update available: 2.701.0
Run: moai doctor --fix
```

#### Doctor 集成

```bash
moai doctor
```

runner 版本检查被整合到系统诊断中。

## 问题排查

### PR 评论触发不工作时

#### Checklist

1. GitHub Actions workflow 是否已启用？
   - Repository → Actions → 检查 workflows

2. GitHub Secrets 是否已设置？
   - Settings → Secrets and variables → Actions

3. Workflow permissions 是否正确？
   - 需要 `contents: read`、`pull-requests: write`

### 各 LLM 错误应对

#### Claude

**Error:** `CLAUDE_CODE_OAUTH_TOKEN expired`
**解决：** 重新运行 `moai github auth claude`

#### Codex

**Error:** `repository visibility check failed`
**原因：** 试图在公开仓库中使用 Codex
**解决：** 将仓库转为私有

#### Gemini

**Error:** `GEMINI_API_KEY quota exceeded`
**解决：** 在 Google Cloud Console 中增加 quota

#### GLM

**Error:** `GLM_API_KEY authentication failed`
**解决：** 检查 `~/.moai/.env.glm` Token

## 下一步

- [CLI 参考](/zh/workflow-commands/)
- [Workflow 配置参考](/zh/advanced/settings-json/)
- [确认安全策略](/zh/advanced/security-notes/)
