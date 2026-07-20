---
title: moai constitution 宪法
weight: 84
draft: false
---

`moai constitution` 查询并验证 zone 注册表(FROZEN/EVOLVABLE zone 的成文化)。它是管理规则中哪些部分被冻结(FROZEN)、不可随意更改,哪些部分可进化(EVOLVABLE)的命令树。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai constitution list` | 列出 zone 注册表项 |
| `moai constitution guard` | 检查 FROZEN zone 违规(用于 CI 集成) |
| `moai constitution amend` | 提出通过 5-layer 安全门禁的宪法修订 |
| `moai constitution validate` | 验证 zone 注册表相对源文件的漂移·不变量 |

## moai constitution list

```bash
moai constitution list
moai constitution list --zone frozen --format json
```

| 标志 | 说明 |
|--------|------|
| `--zone <frozen\|evolvable>` | zone 过滤 |
| `--file <path>` | 文件路径过滤(部分匹配) |
| `--format <table\|json>` | 输出格式 |

## moai constitution guard

```bash
moai constitution guard --violations CONST-V3R2-001,CONST-V3R2-002
```

接受一组已更改的规则 ID,检查 FROZEN zone 违规。用于 CI 集成。

| 标志 | 说明 |
|--------|------|
| `--violations <ids>` | 已更改的规则 ID 列表(逗号分隔或重复标志) |

## moai constitution amend

```bash
moai constitution amend --rule CONST-V3R2-001 --before "..." --after "..." --evidence "..."
```

必须通过 FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight 5-layer 安全门禁才能应用。

| 标志 | 说明 |
|--------|------|
| `--rule <id>` | 规则 ID(CONST-V3R2-NNN)[必填] |
| `--before <text>` | 当前条款文本 [必填] |
| `--after <text>` | 新条款文本 [必填] |
| `--evidence <text>` | 修订依据(Frozen zone 必填) |
| `--dry-run` | 不修改文件,仅模拟 |

## moai constitution validate

```bash
moai constitution validate
```

确认注册表各项条款是否存在于源文件中,验证 zone_class enum·canary_gate 不变量并报告漂移。

| 标志 | 说明 |
|--------|------|
| `--strict` | 严格模式(强制所有检查) |
| `--fail-on-warning` | 将警告视为错误(含 `--strict`) |
| `--format <text\|json>` | 输出格式 |

退出码:0=正常,1=漂移/错误,2=致命(源文件缺失)。

## 相关文档

- [CLI 概览](/zh/getting-started/cli)
