---
title: moai web 控制台
weight: 50
draft: false
---

`moai web` 启动基于浏览器的设置编辑器 **MoAI Web Console**。它复用与终端配置文件向导 (`moai profile`) 相同的校验与持久化逻辑，让你可以在 Web UI 中编辑配置文件偏好以及项目的 user / language / statusline 部分。

## 概览

```bash
moai web [OPTIONS]
```

控制台**仅绑定回环地址 (127.0.0.1)**。没有外部数据库、没有鉴权、也没有网络暴露。默认情况下，如果目标端口被过时的 moai 实例占用，控制台会终止该实例并重新绑定。非 moai (外部) 进程绝不会被终止 — 此时控制台会报告错误并建议使用 `--port`。

## 标志

| 标志 | 说明 |
|------|------|
| `--port <N>` | 在 127.0.0.1 上绑定的 TCP 端口 (默认: `3041`) |
| `--no-open` | 不自动打开浏览器 |
| `--no-reuse` | 不从过时的 moai 实例回收端口；端口冲突时直接失败 |

## 示例

```bash
moai web                 # 绑定 127.0.0.1:3041 并打开浏览器
moai web --port 9000     # 绑定其他端口
moai web --no-open       # 启动但不打开浏览器
moai web --no-reuse      # 端口占用时失败而非回收
```

## 编辑对象

Web 控制台可编辑：

- **配置文件偏好** — 模型、语言、显示设置等按配置文件保存的设置
- **项目设置** — `.moai/config/sections/` 下的 user / language / statusline 部分

保存时会经过与终端向导相同的校验，因此无论使用哪条路径，结果都一致。

## 控制台界面

控制台的界面语言可在页眉右侧的选择器中从 English · 한국어 · 日本語 · 中文 中选择。下面的界面是设为 English 时的状态，因此本节在括号中一并标注界面上显示的英文表述。

页眉中并排显示项目名称、当前配置文件以及主要设置摘要(`lang · model · effort · dev`)。其下是配置文件选择器和配置文件(Profiles) 卡片，接着是用户信息(Identity) · 语言(Language) · LLM · 第三方 LLM(3rd Party LLM) · 智能体(Agents) · 报告(Report) 六个标签页。修改后的值通过底部的保存设置(Save settings) 按钮写入。

![MoAI Web Console 初始界面。页眉的项目名称与配置文件、配置文件(Profiles) 卡片、六个标签页、用户信息(Identity) 标签页的显示名称(Display name) 输入框、保存设置(Save settings) 按钮](/images/profile/web-console-overview.png)

在配置文件(Profiles) 卡片中可以切换配置文件、删除(Delete)，也可以填写新配置文件名称(New profile name) 后用创建配置文件(Create profile) 新建。选择其他配置文件时，页眉中的配置文件显示也会随之变化。下面是切换到 `moai-cowork` 配置文件后打开语言(Language) 标签页的样子。

![切换到 moai-cowork 配置文件后控制台的语言(Language) 标签页。对话语言(Conversation language)、提交信息语言(Commit message language)、代码注释语言(Code comment language)、文档语言(Documentation language) 四个项目](/images/profile/web-console-switch.png)

在 LLM 标签页中修改权限模式(Permission mode) 与模型(Model)、推理强度(Effort level)。与终端向导 "Model Settings" 步骤所处理的值相同。

![moai-adk 配置文件的 LLM 标签页。权限模式(Permission mode)、模型(Model)、推理强度(Effort level) 三个项目](/images/profile/web-console-llm-tab.png)

## 配置文件记录的范围

在控制台中切换配置文件时，该选择会作为当前项目的记录留在 `~/.moai/claude-profiles/launch.yaml` 中。在同一项目中不带 `-p` 运行 `moai cc` 时会使用该值。

{{< callout type="note" >}}
按项目的记录将包含在下一个版本中。当前发布的版本不区分项目，只处理一条全局记录。
{{< /callout >}}

控制台读取的值与写入的值都以当前项目为准，因此界面上显示的配置文件与实际记录的配置文件始终一致。不过，在用 `moai cc -p X` 启动的会话中打开控制台时，`CLAUDE_CONFIG_DIR` 已经确定，因此会与记录无关地原样显示 `X`。

选择顺序与限制在[配置文件管理](/zh/cli-reference/profile#配置文件的自动选择)中有详细说明。

---

相关: [配置文件管理](/zh/cli-reference/profile) · [CLI 概览](/zh/getting-started/cli)
