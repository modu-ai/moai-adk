---
title: moai init 初始化
weight: 5
draft: false
---

`moai init` 在当前目录或新文件夹中初始化 MoAI 项目。它会部署 Claude Code 集成所需的 `.claude/`、`.moai/` 结构与配置,并在需要时通过交互式向导设置项目模式、语言与质量门禁。

## 用法

```bash
moai init [project-name]
```

| 模式 | 行为 |
|------|------|
| `moai init <name>` | 创建 `./<name>/` 文件夹并在其中初始化 |
| `moai init .` | 在当前目录初始化 |
| `moai init` | 在当前目录初始化(等同于 `moai init .`) |

最多接受 1 个参数。

## 主要标志

### 部署范围

| 标志 | 说明 |
|--------|------|
| `--all` | 部署完整目录(core + 可选包 + 生成的 harness)。默认是 core-only slim 模式 |
| `--force` | 重新初始化既有项目(会备份当前 `.moai/`) |
| `--no-hooks` | 跳过 git 钩子安装 |

### 项目默认值

| 标志 | 说明 |
|--------|------|
| `--root <dir>` | 项目根目录(默认:当前目录) |
| `--name <name>` | 项目名称(默认:目录名) |
| `--language <lang>` | 主编程语言 |
| `--framework <name>` | 框架(默认:自动检测或 `none`) |
| `--mode <ddd\|tdd>` | 开发方法论(默认:tdd) |
| `--non-interactive` | 跳过交互式向导 —— 仅使用标志与默认值 |

### 向导阶段

| 标志 | 说明 |
|--------|------|
| `--standard` | 提出 Phase 1 问题(项目模式、harness 配置、LSP、质量门禁、设计) |
| `--advanced` | 提出 Phase 1 + Phase 2 问题(含 `--standard`) |
| `--project-mode <personal\|team>` | 项目模式(默认:personal) |
| `--harness-profile <name>` | harness 评估配置:default、strict、lenient、frontend |
| `--enable-lsp` | 启用 LSP 集成(默认:false) |
| `--enforce-quality` | 强制质量门禁(默认:true) |
| `--enable-design` | 启用设计工作流(默认:true) |

### Git / 模型策略

| 标志 | 说明 |
|--------|------|
| `--git-mode <manual\|personal\|team>` | Git 工作流模式(默认:manual) |
| `--git-provider <github\|gitlab>` | Git 提供商 |
| `--github-username <name>` | GitHub 用户名(personal/team 模式必填) |
| `--model-policy <max\|medium\|low>` | 性能层级 —— 保存到 `llm.yaml` 的 `performance_tier` |
| `--plan-type <api\|subscription>` | 计费方案类型 —— 保存到 `llm.yaml` 的 `plan_type` |

## 示例

```bash
# 在新文件夹中初始化
moai init my-app

# 在当前目录初始化
moai init .

# 指定方法论
moai init --mode tdd

# 部署完整目录(绕过 slim 模式)
moai init --all

# 非交互式(用于 CI 等)
moai init . --non-interactive --language go
```

## 相关命令

| 命令 | 说明 |
|--------|------|
| `moai update` | 同步已初始化项目的模板 |
| `moai status` | 检查初始化状态 |
| `moai doctor` | 初始化后验证环境 |

## 参考

- [项目状态](/zh/cli-reference/status)
- [更新](/zh/cli-reference/update)
- [CLI 概览](/zh/getting-started/cli)
