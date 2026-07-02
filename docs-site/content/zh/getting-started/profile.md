---
title: 配置文件管理
weight: 80
draft: false
---
# 配置文件管理


通过MoAI-ADK的配置文件系统隔离管理多个Claude Code配置。

## 什么是配置文件？

配置文件是**隔离的Claude Code配置目录**（`CLAUDE_CONFIG_DIR`）。可以为每个配置文件维护独立的设置、模型选择和语言环境。

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

显示所有可用的配置文件。

```bash
moai profile list
```

### moai profile setup [name]

运行交互式设置向导。

```bash
moai profile setup          # 设置默认配置文件
moai profile setup work     # 设置"work"配置文件
```

**向导设置项目：**
- **Identity**: 用户名、角色
- **Languages**: 对话语言、代码注释语言
- **Model Settings**: 默认模型、1M上下文模型选择
- **Display**: 输出样式、状态栏设置

### moai profile current

显示当前活动的配置文件名称。

```bash
moai profile current
```

### moai profile delete [name]

删除配置文件。

```bash
moai profile delete old-profile
```

## 使用配置文件运行Claude Code

通过 `-p`（或 `--profile`）标志指定配置文件。

```bash
moai cc -p work          # 以work配置文件运行Claude
moai glm -p cost-save    # 以cost-save配置文件运行GLM
moai cg -p team          # 以team配置文件运行CG模式
```

{{< callout type="info" >}}
未指定配置文件时使用默认配置文件。首次运行时会自动启动设置向导。
{{< /callout >}}

## 选择1M上下文模型

设置配置文件时，可以选择支持1M上下文窗口的模型。

**支持的模型：**
- `claude-opus-4-8[1m]` - Opus 4.8 (1M context)
- `claude-sonnet-4-6[1m]` - Sonnet 4.6 (1M context)

可在设置向导的"Model Settings"步骤中选择，或直接编辑配置文件。

## 切换配置文件时的行为

| 切换 | 行为 |
|------|------|
| `moai cc` → `moai glm` | 自动注入GLM环境变量 |
| `moai glm` → `moai cc` | 自动移除GLM环境变量 |
| `moai cc` → `moai cg` | 仅将GLM env注入tmux会话，Leader仍保持Claude |

## 相关文档

- [CLI参考](/getting-started/cli) - 完整CLI命令参考
- [快速开始](/getting-started/quickstart) - 首次上手指南
- [初始设置](/getting-started/init-wizard) - 项目初始化
