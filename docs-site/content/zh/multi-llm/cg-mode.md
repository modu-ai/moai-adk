---
title: CG 停用与配置迁移
weight: 20
draft: false
description: CG 停用与配置迁移
---

`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。

## 变更前预览

在项目根目录预览选项。预览不会修改配置，也不会创建备份。

```bash
moai migrate cg
moai migrate cg --target claude-only
```

## 迁移为 Claude 单一角色

只有接受取消 GLM 队友自动分配这一角色变化时，才使用以下命令应用迁移。

```bash
moai migrate cg --target claude-only --apply --accept-role-change
```

迁移会写入 `llm.team_mode: claude`、`llm.gateway.teammate_mode: in-process` 和 `llm.gateway.teammate_provider: inherit`。这会取消原有混合角色分配，并不会保留 Claude 领队与 GLM 队友窗格的分工。

## 混合配置的验证条件

`claude-glm` 表示 Claude 领队搭配 tmux 中的 GLM 队友。目前 TEAMMATE 集成验证尚未通过，因此不能应用或启动该方案，只能预览。安装 tmux 或设置 `verified: true` 都不能解除限制。

```bash
moai migrate cg --target claude-glm
```

## 配置保留与错误处理

迁移保留未知配置值、注释、GLM 模型配置和凭据引用。应用前会将原始字节完整备份到 `.moai/backups/cg-migration/<source-sha256>.yaml`。YAML 排版可能改变。再次选择同一目标会返回未更改，改选另一目标则会被拒绝。

gateway 值冲突、非空的 `llm.mode`、重复键和不支持的 YAML 别名需要明确处理。预览和前置检查失败不会改变原文件或备份目录。写入过程中失败可能留下备份；请检查错误和保留的原文件后再重试。

## 凭据处理 {#tmux-env-security}

旧 CG 环境变量注入步骤不再描述可用启动器的行为。迁移不会启动服务提供方，也不会转移凭据。备份包含原始配置，请限制其访问权限。

## 下一步

接受角色变化并迁移到 `claude-only` 后，请明确选择受支持的启动器。`moai cc` 和 `moai glm` 不会重建旧混合角色。GPT gateway 启动另有集成验证条件，此迁移不会将其启用。

- [CLI](/zh/cli-reference/launchers/)
