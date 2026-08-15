---
title: moai web 控制台
weight: 50
draft: false
description: "启动本地运维控制台的 moai web 命令 — 标志、路由、端口回收行为。"
---

`moai web` 启动本地运维界面 **MoAI Web Console**。它让你在浏览器中查看项目的 SPEC 目录与看板链，以及会话、目标、验证的状态，并在同一界面里修改配置文件偏好与项目设置。

界面构成以及各区域读取什么，在 [MoAI Web Console](/zh/advanced/moai-web-console/) 中说明。本页整理命令本身 — 标志、路由与端口处理。

## 概述

```bash
moai web [OPTIONS]
```

控制台**只绑定回环地址（127.0.0.1）**。完全没有外部数据库、认证与网络暴露。

## 标志

| 标志 | 默认值 | 说明 |
|------|--------|------|
| `--port <N>` | `3041` | 要在 127.0.0.1 上绑定的 TCP 端口 |
| `--no-open` | `false` | 不自动打开浏览器 |
| `--no-reuse` | `false` | 不从陈旧的 moai 实例回收端口，端口冲突时直接失败 |

## 示例

```bash
moai web                 # 绑定 127.0.0.1:3041 并打开浏览器
moai web --port 9000     # 绑定到其他端口
moai web --no-open       # 不打开浏览器直接启动
moai web --no-reuse      # 端口被占用时不回收而是失败
```

## 端口回收行为

若目标端口已被陈旧的 moai 实例占用，默认行为会终止该实例并重新绑定。**绝不会终止** moai 以外的外部进程；这种情况下会报错，并建议用 `--port` 换一个端口。加上 `--no-reuse` 后，即使是陈旧的 moai 实例也不回收，直接失败。

即使打开浏览器失败，服务端仍会保持运行。手动打开终端里打印的地址即可。

## 路由

控制台提供以下路径。四个界面是只读的，会以 405 拒绝 GET 以外的方法。

| 路径 | 方法 | 作用 |
|------|------|------|
| `/` | GET | 概览 — 统计磁贴、看板链、进行中的 SPEC、注意列表、会话 |
| `/kanban` | GET | 链会话看板 + SPEC 流水线 |
| `/specs` | GET | SPEC 目录。`?q=` 搜索、`?status=` 筛选、`?id=` 打开详情 |
| `/monitor` | GET | 会话・目标・验证・史诗 |
| `/settings` | GET | 设置九个标签页。`?tab=` 指定标签页，`?profile=` 指定编辑对象配置文件 |
| `/events` | GET | SSE 流 — 只推送更新信号 |
| `/save` | POST | 保存设置 |
| `/profile/create` · `/profile/rename` · `/profile/delete` | POST | 配置文件生命周期 |
| `/glm-key/reveal` | POST | 显示已保存的 GLM API 密钥 |
| `/__shutdown__` | POST | 关闭服务端 |

`/events` 是一条保持连接的流。用 `curl` 直接打开时响应不会结束，因此被超时切断才是正确行为。

## 停止

在终端按 `Ctrl+C`，或使用导轨底部的退出按钮。两种方式都会先处理完进行中的请求再安全退出。

## 配置文件记录的范围

在控制台中切换配置文件时，该选择会作为当前项目的记录留在 `~/.moai/claude-profiles/launch.yaml` 中。在同一项目里不带 `-p` 运行 `moai cc` 时会使用这个值。

控制台读取与写入的值都以当前项目为准，因此界面上显示的配置文件与实际记录的配置文件始终一致。不过，在以 `moai cc -p X` 启动的会话中打开控制台时，`CLAUDE_CONFIG_DIR` 已经确定，因此无论记录如何都会直接显示 `X`。

选择顺序与约束在[配置文件管理](/zh/cli-reference/profile/)中详细说明。

---

相关: [MoAI Web Console](/zh/advanced/moai-web-console/) · [配置文件管理](/zh/cli-reference/profile/) · [CLI 概述](/zh/getting-started/cli/)
