---
title: "GitHub 集成指南"
description: "用 moai github 子命令解析 issue 并与 SPEC 关联"
draft: false
weight: 10
---

MoAI-ADK 的 GitHub 集成功能提供解析 GitHub issue 并与 SPEC 文档关联的
轻量级 CLI 工具。所有命令都通过本地安装的 `gh` CLI 获取当前仓库的
issue 数据。

> **范围说明**:本页只讲实际分发的 `moai github` 子命令及随附的
> GitHub Actions 资产。把多个 LLM 作为面板挂到 PR 上的"多 LLM 审查面板"
> 目前不包含在分发版本中。

## 前置要求

- 已安装 MoAI-ADK(macOS · Linux · Windows)
- 已安装并认证 GitHub CLI(`gh`)(`gh auth login`)
- GitHub 仓库

## moai github 子命令

`moai github` 提供两个活跃子命令。两者共同支持 `--dry-run` 标志,
可在不做实际变更的情况下预览将执行的操作。

### 解析 issue: `moai github parse-issue`

```bash
moai github parse-issue 123
```

用 `gh` CLI 获取指定编号的 issue,并以卡片形式输出编号·标题·作者·
标签·正文摘要·评论数。

### 关联 SPEC: `moai github link-spec`

```bash
moai github link-spec 123 SPEC-ISSUE-123
```

在 GitHub issue 与 SPEC 文档之间建立双向链接,并把该映射保存到
`.moai/github-spec-registry.json`。SPEC ID 在保存前会经过格式验证。

```bash
# 不做实际变更,仅确认计划
moai github link-spec 123 SPEC-ISSUE-123 --dry-run
```

## 随附分发的 GitHub Actions 资产

`moai init` 会在 `.github/` 下部署以下两个资产。

### Label Sync 工作流(`.github/workflows/label-sync.yml`)

以 `.github/labels.yml` 为单一真实来源,同步仓库标签。

- **触发**:`workflow_dispatch`(手动,支持 `dry_run` 输入)或
  `.github/labels.yml` / 工作流文件 push 到 `main` 时自动运行
- **权限**:`issues: write`、`pull-requests: write`、`contents: read`
- **行为**:用 EndBug/label-sync 动作把 `labels.yml` → 仓库标签反映

### detect-language 组合动作(`.github/actions/detect-language/action.yml`)

以仓库第一个源文件的扩展名为准检测主语言,并作为
`language` 输出值导出。

- **支持语言(16 种)**:Go, Python, TypeScript, JavaScript, Rust, Java,
  Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift
- **实现备注**:用 `find ... -print -quit` 在首个匹配后立即退出,以在
  `set -o pipefail` 环境下避免 broken-pipe 失败

## 故障排查

### 找不到 `gh` 命令时

`moai github` 子命令依赖本地 `gh` CLI。用 `gh --version` 确认安装,
用 `gh auth login` 完成认证。

### 无法获取 issue 时

确认当前目录是否在目标仓库的工作树内,以及 `gh` 是否对该仓库有访问权限。

### SPEC ID 验证失败

`link-spec` 只接受遵循 `SPEC-` 前缀的有效 SPEC ID。确认 ID 格式后重新运行。

## 下一步

- [参阅 CLI 参考](/zh/workflow-commands/)
- [参阅 Workflow 设置](/zh/advanced/settings-json/)
- [确认安全策略](/zh/advanced/security-notes/)
