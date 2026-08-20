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

## Home Disk Usage 检查 {{< new-badge v3.1.1 >}}

完整的 `moai doctor` 运行会带上 **Home Disk Usage** 项。它报告 `~/.moai` 主目录塞了多满,属于**建议 (advisory)** 性质的检查 —— 超标也不会拦住其他命令。

| 报告项目 | 内容 |
|----------|------|
| 总体大小 | `~/.moai` 的总容量与最大的 3 个条目 |
| 按配置档案分解 | 每个 `claude-profiles/<配置档案>` 的大小与分类拆分 |
| 发行版数量 | `releases/` 中剩余的二进制数量与当前版本 |
| 可清理量 | `moai clean --home` 实际能删除的估算字节数 |
| `~/.claude` | 只报告大小 —— 在任何路径下都不是清理对象 |

当可清理量超过阈值(编译默认值 500 MB)时,状态转为 WARN 并推荐 `moai clean --home`(默认 dry-run)。低于阈值则保持 OK。若 `~/.moai` 根本不存在,该检查报告"无可报告"并通过。

这个估算调用的是与 `moai clean --home` **同一个扫描器**,所以 doctor 报出的数字和 clean 实际删除的清单不会脱节。详见 [主目录卫生](/zh/advanced/home-hygiene)。

## 示例

```bash
moai doctor                            # 完整诊断
moai doctor --verbose                  # 详细诊断
moai doctor --export diagnostics.json  # 导出结果
moai doctor hook                       # 钩子覆盖表
```

---

相关: [项目状态](/zh/cli-reference/status) · [CLI 概览](/zh/getting-started/cli)
