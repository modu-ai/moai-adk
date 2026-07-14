---
title: 配置文件管理
weight: 80
draft: false
---

用 MoAI-ADK 的配置文件系统隔离管理多套 Claude Code 设置。把工作用与个人用、高质量会话与省成本会话各分成一个配置文件后,就无需每次更改模型·语言·显示设置。

## 什么是配置文件?

配置文件是 **隔离的 Claude Code 设置目录**(`CLAUDE_CONFIG_DIR`)。可为每个配置文件维护独立的设置、模型选择、语言环境。

```
~/.moai/claude-profiles/
├── default/           # 默认配置文件
│   ├── settings.json
│   └── settings.local.json
├── work/              # 工作用配置文件
│   ├── settings.json
│   └── settings.local.json
└── personal/          # 个人用配置文件
    └── ...
```

## 命令参考

### moai profile list

显示所有可用配置文件。

```bash
moai profile list
```

### moai profile setup [name]

运行交互式设置向导。

```bash
moai profile setup          # 设置默认配置文件
moai profile setup work     # 设置 "work" 配置文件
```

**向导设置项:**
- **Identity**:用户名、角色
- **Languages**:对话语言、代码注释语言
- **Model Settings**:默认模型、1M 上下文模型选择
- **Display**:输出风格、状态栏设置

### moai profile current

显示当前活跃的配置文件名。

```bash
moai profile current
```

### moai profile delete [name]

删除配置文件。

```bash
moai profile delete old-profile
```

## 用配置文件运行 Claude Code

用 `-p`(或 `--profile`)标志指定配置文件。

```bash
moai cc -p work          # 用 work 配置文件运行 Claude
moai glm -p cost-save    # 用 cost-save 配置文件运行 GLM
moai cg -p team          # 用 team 配置文件运行 CG 模式
```

{{< callout type="info" >}}
未指定配置文件时使用默认配置文件。首次运行时会自动启动设置向导。
{{< /callout >}}

## 1M 上下文模型选择

设置配置文件时可选择支持 1M 上下文窗口的模型。`[1m]` 后缀不是单独的模型,而是 Claude Code 原生的上下文窗口修饰符。

**可选的模型别名:**
- `opus` / `opus[1m]`
- `sonnet` / `sonnet[1m]`
- `fable` / `fable[1m]`

在设置向导的 "Model Settings" 步骤选择,或直接修改配置文件设置文件。1M 上下文模型适合大规模代码库分析或长文档作业。

## 切换配置文件时的行为

| 切换 | 行为 |
|------|------|
| `moai cc` → `moai glm` | 自动注入 GLM 环境变量 |
| `moai glm` → `moai cc` | 自动移除 GLM 环境变量 |
| `moai cc` → `moai cg` | 仅向 tmux 会话注入 GLM env,Leader 保持 Claude |

## 相关文档

- [CLI 参考](/getting-started/cli) - 全部 CLI 命令
- [快速开始](/getting-started/quickstart) - 从头开始
- [初始设置](/getting-started/init-wizard) - 项目初始化
