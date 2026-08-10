---
title: 配置文件管理
weight: 40
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

该值以全局记录为准,因此当各项目记住的配置文件不同时,请一并参考[配置文件的自动选择](#配置文件的自动选择)中的限制。

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
用 `-p` 指定的配置文件始终优先。未指定时使用哪个配置文件,请参考下面的[配置文件的自动选择](#配置文件的自动选择)。首次使用某个配置文件时,设置向导会自动启动。
{{< /callout >}}

## 配置文件的自动选择

不带 `-p` 运行 `moai cc` 时,会参考 `~/.moai/claude-profiles/launch.yaml` 的记录来挑选配置文件。每次用 `-p` 运行具名配置文件,该记录都会更新。

{{< callout type="note" >}}
下面说明的按项目记忆将包含在下一个版本中。当前发布的版本只保留一条全局记录(`last_profile`),因此在项目 B 中用 `-p` 指定配置文件时,项目 A 记住的值会被覆盖。
{{< /callout >}}

`launch.yaml` 在全局记录之外,还维护一份以项目绝对路径为键的 `projects:` 列表。不带 `-p` 的运行按以下顺序确定配置文件。

1. 当前项目记住的配置文件(`projects:` 条目)
2. 全局记录(`last_profile`)
3. 默认配置文件

即使是已记录的配置文件,若其目录已被删除,也会跳过并进入下一顺序。用 `-p` 指定的名称优先于整个顺序,也可以用 `-p default` 明确指定默认配置文件。

要同时关闭这两项查找,请设置环境变量。

```bash
MOAI_NO_PROFILE_FALLBACK=1 moai cc    # 忽略记录,以默认配置文件运行
```

按项目的记录在用 `-p` 运行时留下,在[Web 控制台](/zh/cli-reference/web)中切换配置文件时也会一并更新。默认配置文件(`default`)不在记录范围内。

**需要了解的限制**

- 移动项目目录或更改其名称后,原有条目将与任何路径都不匹配。该条目会被静默跳过,因此不会影响运行。
- `projects:` 列表会随项目增多而一起增长,目前还没有可清理它的命令。
- `moai profile current` 会原样显示全局记录。因此在记住的配置文件与全局记录不同的项目中,`moai profile current` 给出的名称,可能与不带 `-p` 的 `moai cc` 实际启动的配置文件不一致。

## 新配置文件的首次运行

新建的配置文件目录中还没有存放 Claude Code 账号状态的 `.claude.json`。账号状态在任何平台上都按设置目录分别管理,因此即便正在使用的会话完好无损,用新配置文件首次运行时仍会出现登录·引导界面。

{{< callout type="note" >}}
下面的提示消息将包含在下一个版本中。当前发布的版本会在毫无预告的情况下切到登录界面。
{{< /callout >}}

启动器在拉起 Claude Code 之前,会向标准错误输出以下内容。

```
Notice: profile "work" has no Claude Code configuration yet.
  Claude Code will show the login / onboarding screen on this launch.
  Account state is not inherited between profiles; sign in once and it
  persists for this profile.
```

不会把凭据复制或移动到新配置文件中。账号状态的存放位置因平台而异,针对其中一方的复制在另一方就会出错。登录一次后该状态便留在该配置文件中,因此之后的运行不会再出现此界面。

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

- [moai web 控制台](/zh/cli-reference/web) - 在浏览器中切换与编辑配置文件
- [CLI 参考](/zh/getting-started/cli) - 全部 CLI 命令
- [快速开始](/zh/getting-started/quickstart) - 从头开始
- [初始设置](/zh/getting-started/init-wizard) - 项目初始化
