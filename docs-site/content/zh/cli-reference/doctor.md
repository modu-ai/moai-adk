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

## Hook Delivery 检查 {{< new-badge v3.1.4 >}}

`moai update` 可能在模板新增了钩子条目时,没有把它们写进现有项目的 `.claude/settings.json` 就悄悄跳过。**Hook Delivery** 检查就是找出这类缺失:在项目已经拥有的钩子事件键范围内,对比发行模板与项目设置,报告模板有而项目缺少的条目。

| 报告项目 | 内容 |
|----------|------|
| 缺失条目 | 以 `hooks.PreToolUse missing handle-pre-tool.sh (matcher AskUserQuestion)` 的形式指出,连事件键和匹配器一起说明 |
| 修复方法 | 在被指名的事件键下,从所用 moai 版本的模板设置中复制对应块补回 |
| 更新后核验 | 给出检查 `moai update` 是否删除了受管文件的命令(`git status --porcelain \| grep '^ D'`)—— 命中时先用 `git restore -- <路径>` 还原,再补回条目 |

这项检查是只读的 —— 绝不写入 `.claude/settings.json`。模板首次引入的事件键和用户自己编写的条目不计为缺失,opt-out 的条目也不会被要求(模板会按项目自身的 `hook.opt_in.enabled` 设置渲染)。全部一致时报 `ok`。

## Codex Wiring 诊断 {{< new-badge v3.1.4 >}}

完整的 `moai doctor` 运行会带上 **Codex Wiring** 项。它检视项目的 Codex 配线 (生成的钩子、MCP 注册、技能镜像) 是否完好,属于建议 (advisory) 性质的检查 —— fail-open 且只读,只报告发现的问题,绝不自行修复。

| 检查项目 | 内容 |
|----------|------|
| `.codex/hooks.json` 存在 · 键白名单 | 确认配线文件是否存在、键是否符合白名单。只要有一个多余的键,codex 就会静默忽略整个文件 —— 这项检查就是在替它观测这份沉默 |
| sidecar 哈希 | 将当前文件与部署时记录下来的钩子内容哈希对比,抓住手工改动钩子之后留下的偏差 |
| `moai` 二进制 PATH | 生成的钩子命令都是 `moai hook ...` 形式,PATH 上找不到 `moai` 时一条都不会触发 |
| `.codex/config.toml` 中的 `[mcp_servers.moai]` | 确认 MCP 注册表是否存在、是否与标准注册形态一致。这张表归用户所有,doctor 只报告,不修复 |
| `.agents/skills` 技能镜像 | Codex CLI 不会扫描 `.claude/skills`,所以没有镜像、或镜像链接断开时,codex 就看不到这个项目的 MoAI 技能 |
| 用户层 `[[skills.config]]` 注册 | 确认 `~/.codex/config.toml`(设置了 `$CODEX_HOME` 时则是该文件)里技能注册的路径与 `enabled` 键的形态。这一项只在 codex 参与时才检查 —— 也就是项目已接线,或 codex 已安装的时候 |

当检查发现问题时,代码里内置的修复指令会一并显示。

| 发现 | 指令 |
|------|------|
| 完全没有配线的项目 | `moai init --agent codex` |
| 钩子改动后的 sidecar 偏差 | 用 `codex /hooks` 重新信任改动过的钩子 |
| 技能镜像缺失 · 链接断开 | `moai update --templates-only --force --yes` |
| 指向的技能文件已消失的注册 | 移除该条目,或还原技能文件 |

这项检查能给出的发现里,被归为 fatal 的只有一个。用户层设置里的 `[[skills.config]]` 条目缺少 `enabled` 键,或者其中填的不是 bare TOML 布尔值时,codex 在每次调用中都以 exit 1 结束。命中这个唯一的 fatal 发现,doctor 的结果就变成 Fail,退出码随之变成 1。不过这一行为是在 codex-cli 0.153.4 上观测到的 —— 只确认了该版本如此,并未推广到其他版本。正确的做法是在每个条目里都明确写出 `enabled = true` 或 `enabled = false`。除这一项之外,所有发现都属于建议性质,不会改变 doctor 的退出码。

在没有安装 codex 的机器上,不带任何配线的 claude-only 项目会让这项检查静默跳过 —— 信息性跳过,不会产生警告行。顺带说明:把指向已不存在技能文件的幽灵 (ghost) 注册一次性收走的功能,在这项检查的指令范围之外,由单独的动词 `moai clean --codex-skills` 承担。

包含钩子信任模型与技能镜像在内的 Codex 配线全貌,详见 [Codex 双 harness](/zh/advanced/codex-dual-harness) 文档。

## 退出码

脚本和 CI 包装器调用 `moai doctor` 时，读的是退出码，而不是摘要那一行。

| 退出码 | 含义 |
|--------|------|
| `0` | 没有 Fail 项。Warn 属于劝告，不改变退出码 |
| `1` | 有一项以上 Fail —— 摘要里的 `Fail N` 原样体现 |

Constitution Registry 这一项不只确认注册表能否解析，它跑的是与 `moai constitution validate` **相同的漂移校验**。所以在同一个检出里，doctor 说 ok 而 validate 失败的情况不会出现。用 `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 绕过时，doctor 回到它自己的结构检查判定。

## 示例

```bash
moai doctor                            # 完整诊断
moai doctor --verbose                  # 详细诊断
moai doctor --export diagnostics.json  # 导出结果
moai doctor hook                       # 钩子覆盖表
```

---

相关: [项目状态](/zh/cli-reference/status) · [CLI 概览](/zh/getting-started/cli)
