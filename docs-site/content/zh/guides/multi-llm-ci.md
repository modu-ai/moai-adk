---
title: "Multi-LLM CI 指南"
description: "在 GitHub Actions 中使用多个 AI 模型进行代码审查自动化"
date: 2026-04-27
draft: false
weight: 10
---

# Multi-LLM CI 指南

了解如何使用 MoAI-ADK 的 Multi-LLM CI 功能在 GitHub Actions 中设置多个 LLM 进行代码审查。

## 概述

### 什么是 Multi-LLM CI？

MoAI-ADK 的 Multi-LLM CI 功能提供集成的 CI/CD 管道，在 GitHub Actions 中使用多个 AI 模型同时执行代码审查。

### 支持的 LLM

| LLM | 提供商 | 触发方式 | 特性 |
|-----|--------|----------|------|
| **Claude** | Anthropic | `/claude` 评论 | Issue/PR 审查，OAuth 认证 |
| **Codex** | OpenAI | PR 打开时自动 | ⚠️ 仅限私有仓库 |
| **Gemini** | Google | PR 打开时自动 | API Key 认证 |
| **GLM** | Zhipu AI | PR 打开时自动 | Token 认证 |

### 用户收益

- **同时进行多 LLM 审查**：在一个 PR 中同时获得多个 LLM 的反馈
- **统一管理**：通过 `moai github` CLI 进行一致的设置
- **安全认证**：每个 LLM 专用认证处理
- **语言检测**：自动检测项目语言并分配适当的 LLM

## 入门指南

### 前提条件

- macOS (arm64) - v1.0 基线
- Go 1.23+
- GitHub 仓库
- 各 LLM 账户和 API 令牌

### 初始设置

```bash
moai github init
```

此命令将：
- 创建 `.github/workflows/` 目录
- 部署 workflow 模板
- 部署 composite actions
- 引导 GitHub Secrets 设置

### LLM 认证设置

```bash
# Claude (OAuth)
moai github auth claude

# Codex (私有仓库)
moai github auth codex

# Gemini
moai github auth gemini

# GLM
moai github auth glm
```

### GitHub Secrets 设置

每个 LLM 所需的 Secrets：
- `CLAUDE_CODE_OAUTH_TOKEN` - Claude OAuth 令牌
- `CODEX_AUTH_JSON` - Codex 认证 JSON（base64 编码）
- `GEMINI_API_KEY` - Gemini API Key
- `GLM_API_KEY` - GLM API 令牌

### 测试第一个 PR

创建 PR 时，会自动添加 LLM Panel 评论：

```markdown
## LLM Code Review Status

| LLM | Status |
|-----|--------|
| Claude | Pending (添加 `/claude` 评论) |
| Codex | ✓ Ready |
| Gemini | ⚠️ Token missing |
| GLM | ✓ Ready |

触发单独审查：
- 添加 `/claude` 评论以触发 Claude
- 添加 `/codex` 评论以触发 Codex
- 添加 `/gemini` 评论以触发 Gemini
- 添加 `/glm` 评论以触发 GLM
```

## LLM 认证设置

### Claude 设置

#### OAuth 令牌颁发

1. 安装 [Claude Code](https://claude.ai/download)
2. 登录后颁发 OAuth 令牌
3. 自动保存到 `.claude/settings.local.json`

#### moai github auth claude

```bash
moai github auth claude
```

**交互式设置过程：**
```
未找到 Claude OAuth 令牌。
是否安装 Claude Code 并登录？ (y/n): y

[已确认] OAuth 令牌已保存到 settings.local.json。
将 GitHub Secret: CLAUDE_CODE_OAUTH_TOKEN 设置为：
<token-value>
```

### Codex 设置（仅限私有仓库）

#### 认证 JSON 创建

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
OpenAI auth.json 文件路径： ~/.codex/auth.json
读取文件以生成 GitHub Secret...
⚠️ Codex 仅限于私有仓库使用 (REQ-SEC-001)

生成的 Secret：
CODEX_AUTH_JSON=eyJ0...
```

### Gemini 设置

```bash
moai github auth gemini
```

输入 API Key 后，自动提供 GitHub Secret 设置指南。

### GLM 设置

```bash
moai github auth glm
```

从 GLM 令牌路径 (`~/.moai/.env.glm`) 自动读取。

## Workflow 模板详解

### llm-panel.yml

**触发器：** PR 打开时

**作用：** 自动创建面板评论，显示每个 LLM 的状态

**备注：** 通过 `/claude`、`/codex`、`/gemini`、`/glm` 评论触发单独审查

### claude.yml / claude-code-review.yml

- **claude.yml**：Issue 触发（初始审查）
- **claude-code-review.yml**：PR 触发（变更审查）

**特性：** 仅通过 `/claude` 评论触发

### codex-review.yml

**安全约束：**
- 仅在 `private` 仓库上运行 (REQ-SEC-001)
- visibility 检查阻止公开仓库

**workflow:**
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
- PR synchronize 时自动触发

### glm-review.yml

- GLM 专用环境设置（setup-glm-env action）
- 自动环境变量注入

### Composite Actions

#### detect-language

**输入：** 仓库根路径
**输出：** language 环境变量 (`detected_language`)

**支持的语言：** Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift（16种语言）

#### setup-glm-env

为 GLM 团队模式设置所需的环境变量：
- `ANTHROPIC_AUTH_TOKEN` (GLM endpoint)
- `ANTHROPIC_BASE_URL` (https://glm.modu-ai.kr)

## 高级配置

### github-actions.yaml 自定义

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

#### 自动更新检查

```bash
moai github status
```

**示例输出：**
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

Runner 版本检查已集成到系统诊断中 (T-27)。

## 故障排除

### PR 评论触发器不起作用

#### 检查清单

1. ✅ GitHub Actions workflow 是否已启用？
   - Repository → Actions → workflows

2. ✅ GitHub Secrets 是否已配置？
   - Settings → Secrets and variables → Actions

3. ✅ Workflow 权限是否正确？
   - 需要 `contents: read`、`pull-requests: write`

### LLM 特定错误处理

#### Claude

**错误：** `CLAUDE_CODE_OAUTH_TOKEN expired`
**解决：** 重新运行 `moai github auth claude`

#### Codex

**错误：** `repository visibility check failed`
**原因：** 尝试在公开仓库上使用 Codex
**解决：** 将仓库设为私有

#### Gemini

**错误：** `GEMINI_API_KEY quota exceeded`
**解决：** 在 Google Cloud Console 中增加配额

#### GLM

**错误：** `GLM_API_KEY authentication failed`
**解决：** 验证 `~/.moai/.env.glm` 中的令牌

## 下一步

- [CLI 参考](/docs/commands/)
- [Workflow 配置](/docs/configuration/)
- [安全策略](/docs/security/)
