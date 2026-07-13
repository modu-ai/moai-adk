---
title: 配置文件管理
weight: 80
draft: false
---
# 配置文件管理


通过 MoAI-ADK 的配置文件 (profile) 系统，可以隔离管理多套 Claude Code 设置。将工作用与个人用、高质量会话与节省成本会话各自分成一个配置文件，就无需每次都切换模型·语言·显示设置。

## 什么是配置文件？

配置文件是**隔离的 Claude Code 设置目录**（`CLAUDE_CONFIG_DIR`）。每个配置文件可以维持独立的设置、模型选择和语言环境。

```
~/.moai/claude-profiles/
├── default/           # 기본 프로필
│   ├── settings.json
│   └── settings.local.json
├── work/              # 업무용 프로필
│   ├── settings.json
│   └── settings.local.json
└── personal/          # 개인용 프로필
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
moai profile setup          # 기본 프로필 설정
moai profile setup work     # "work" 프로필 설정
```

**向导设置项目：**
- **Identity**：用户名、角色
- **Languages**：对话语言、代码注释语言
- **Model Settings**：默认模型、1M 上下文模型选择
- **Display**：输出风格、状态栏设置

### moai profile current

显示当前激活的配置文件名。

```bash
moai profile current
```

### moai profile delete [name]

删除配置文件。

```bash
moai profile delete old-profile
```

## 用配置文件运行 Claude Code

使用 `-p`（或 `--profile`）标志指定配置文件。

```bash
moai cc -p work          # work 프로필로 Claude 실행
moai glm -p cost-save    # cost-save 프로필로 GLM 실행
moai cg -p team          # team 프로필로 CG 모드 실행
```

{{< callout type="info" >}}
未指定配置文件时使用默认配置文件。首次运行时会自动启动设置向导。
{{< /callout >}}

## 选择 1M 上下文模型

设置配置文件时可以选择支持 1M 上下文窗口的模型。`[1m]` 后缀不是独立模型，而是 Claude Code 的原生上下文窗口修饰符。

**可选的模型别名：**
- `opus` / `opus[1m]`
- `sonnet` / `sonnet[1m]`
- `fable` / `fable[1m]`
- `haiku`, `opusplan`

可在设置向导的 "Model Settings" 步骤中选择，或直接修改配置文件的设置文件。1M 上下文模型适合大型代码库分析或长文档处理。

## 切换配置文件时的行为

| 切换 | 行为 |
|------|------|
| `moai cc` → `moai glm` | 自动注入 GLM 环境变量 |
| `moai glm` → `moai cc` | 自动移除 GLM 环境变量 |
| `moai cc` → `moai cg` | GLM env 仅注入 tmux 会话，Leader 保持 Claude |

## 相关文档

- [CLI 参考](/zh/getting-started/cli) - 全部 CLI 命令
- [快速开始](/zh/getting-started/quickstart) - 从头开始
- [初始设置](/zh/getting-started/init-wizard) - 项目初始化
