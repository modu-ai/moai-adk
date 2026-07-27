---
title: moai ast-grep / ast-edit 结构化搜索与替换
weight: 77
draft: false
---

`moai ast-grep` 以语法树为单位扫描代码，`moai ast-edit` 则实际替换匹配到的代码。与基于文本的 `grep` 不同，两者都按语法结构匹配，因此不受空格、换行和变量名差异的影响。

两个命令都调用 [ast-grep](https://ast-grep.github.io/) CLI（`sg`）。若未安装 `sg`，两个命令都不会报错，只输出提示信息后退出。

> **读取与写入是分离的命令。** `ast-grep` 绝不修改文件，`ast-edit` 会修改。由于是独立命令，授予 `Bash(moai ast-grep:*)` 不会同时打开写入权限。

## moai ast-grep — 扫描（只读）

| 参数 | 说明 |
|--------|------|
| `--format` | 输出格式：`text`（默认）、`json`、`sarif` |
| `--lang` | 仅扫描指定语言（如 `go`、`python`、`typescript`） |
| `--severity` | 显示的最低严重级别（`error`、`warning`、`info`） |
| `--rules-dir` | 规则目录路径（默认 `.moai/config/astgrep-rules`） |
| `--dry` | 仅输出将要应用的规则列表，跳过实际扫描 |

```bash
# 扫描整个项目
moai ast-grep ./

# 仅 Go，输出 SARIF（用于上传至 GitHub code scanning）
moai ast-grep --format=sarif --lang=go ./internal/

# 仅显示 error 级别
moai ast-grep --severity=error ./
```

## moai ast-edit — 替换（修改文件）

不带 `--dry` 运行会 **直接修改文件**。建议先用 `--dry` 预览变更内容。

| 参数 | 说明 |
|--------|------|
| `--dry` | 不修改文件，仅输出将要变更的内容 |
| `--pattern` | 要匹配的 ast-grep 模式（与 `--rewrite` 配合使用） |
| `--rewrite` | 替换后的模式（与 `--pattern` 配合使用） |
| `--rule` | 仅应用指定 ID 的规则（规则模式） |
| `--lang` | 目标代码的语言 |
| `--rules-dir` | 规则目录路径（默认 `.moai/config/astgrep-rules`） |
| `--format` | 输出格式：`text`（默认）、`json` |

### 模式模式

同时指定 `--pattern` 和 `--rewrite` 会替换所有匹配的代码。仅指定其中一个会被拒绝并报错。

```bash
# 先预览
moai ast-edit --dry --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/

# 确认后实际应用
moai ast-edit --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/
```

### 规则模式

不带 `--pattern` 运行时，命令会读取规则目录，仅应用声明了 `fix:` 字段的规则。没有 `fix:` 的规则属于仅检测类型，会被跳过并报告数量。

```bash
# 应用所有 fix: 规则（预览）
moai ast-edit --dry ./internal/

# 仅应用特定规则
moai ast-edit --rule my-rule-id ./internal/
```

随附的规则集（`go/hardcoding`、`security/credentials`、`security/crypto`、`security/injection`）全部为 **仅检测**。自动替换可能改变语义或破坏编译，因此有意未添加 `fix:`。如需自动替换，请在项目自有规则中声明 `fix:`。

## 规则文件位置

两个命令默认都读取 `.moai/config/astgrep-rules/`。`sgconfig.yml` 定义生效的规则目录。

## 相关文档

- [moai loop](/zh/cli-reference/loop) — 基于诊断的迭代修复循环
- [CLI 概览](/zh/getting-started/cli)
