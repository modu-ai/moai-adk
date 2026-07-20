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

---

相关: [配置文件管理](/cli-reference/profile) · [CLI 概览](/getting-started/cli)
