---
title: moai doctor 诊断
weight: 60
draft: false
---

`moai doctor` 运行全面的系统诊断。它检查 Claude Code 配置、依赖项、项目结构、语言相关的开发工具和环境，并可为检测到的问题提供修复建议。

## 概览

```bash
moai doctor [OPTIONS]
```

## 标志

| 标志 | 说明 |
|------|------|
| `-v, --verbose` | 显示详细诊断信息 (工具版本、语言检测) |
| `--fix` | 为检测到的问题提供修复建议 |
| `--export` | 将诊断结果导出到 JSON 文件 |
| `--check <tool>` | 仅运行特定检查 (例如 git、go、config) |

## 子命令

`moai doctor` 提供深入特定领域的子命令。

| 命令 | 说明 |
|------|------|
| `moai doctor config` | 配置诊断 — 检查带 provenance 的合并设置 |
| `moai doctor hook` | 显示 27 个事件的钩子覆盖表 |
| `moai doctor permission` | 诊断权限解析 |
| `moai doctor sandbox` | 沙箱后端可用性诊断 |

`moai doctor config` 又提供 `dump` (转储合并设置) 和 `diff <tier-a> <tier-b>` (比较两个设置层级)。

## 示例

```bash
moai doctor                            # 完整诊断
moai doctor --verbose                  # 详细诊断
moai doctor --export diagnostics.json  # 导出结果
moai doctor hook                       # 钩子覆盖表
```

---

相关: [项目状态](/cli-reference/status) · [CLI 概览](/getting-started/cli)
