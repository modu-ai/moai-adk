---
title: moai tool-policy 工具策略
weight: 88
draft: false
---

`moai tool-policy` 管理工具/权限策略 SSOT。`.moai/config/sections/tool-policy.yaml` 是单一真实来源,由它生成(codegen)`settings.json` 的 permissions 块并查询策略项。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai tool-policy build` | 从 tool-policy.yaml 重新生成 settings.json permissions 块 |
| `moai tool-policy list` | 列出 tool-policy 项(thin query) |

## moai tool-policy build

```bash
moai tool-policy build
moai tool-policy build --local-only
```

重新生成本地 `.claude/settings.json` 与模板 `settings.json.tmpl` 的 permissions 块。

| 标志 | 说明 |
|--------|------|
| `--repo-root <path>` | 仓库根目录(默认:cwd) |
| `--policy <path>` | tool-policy.yaml 路径(默认:`<repo-root>/.moai/config/sections/tool-policy.yaml`) |
| `--local-only` | 仅重新生成本地 `.claude/settings.json`(跳过模板 .tmpl) |
| `--template-only` | 仅重新生成模板 settings.json.tmpl(跳过本地) |
| `--default-mode <mode>` | 覆盖 `permissions.defaultMode`(默认:保留既有值) |
| `--json` | 以 JSON 输出结果 |

## moai tool-policy list

```bash
moai tool-policy list
moai tool-policy list --risk-tier irreversible --decision deny
```

| 标志 | 说明 |
|--------|------|
| `--risk-tier <read\|write\|irreversible>` | 按风险层级过滤 |
| `--decision <allow\|deny\|ask>` | 按决策过滤 |
| `--tool <name>` | 按工具名过滤(精确匹配) |
| `--format <text\|json>` | 输出格式 |
| `--repo-root <path>` | 仓库根目录(默认:cwd) |
| `--policy <path>` | tool-policy.yaml 路径 |

## 相关文档

- [settings.json 指南](/zh/advanced/settings-json) —— permissions 块详解
- [config 段落参考](/zh/advanced/config-sections)
- [CLI 概览](/zh/getting-started/cli)
