---
title: "@MX TAG 系统"
weight: 61
draft: false
---

@MX TAG 是代码级注解，是 AI 代理在开发会话之间传递 **上下文 · 不变量 · 危险区** 的标准手段。提示词可能被忽略，但刻在代码中的注释会随代码留存，下一个代理首次阅读代码的瞬间即可立即把握意图与约束。

> @MX TAG 的运维（扫描 · 添加 · 查询）通过 `/moai mx` 命令执行。本页面讨论标签系统自身的协议与生命周期。

## 标签语法

```go
// @MX:TAG_TYPE: [描述]
// @MX:SUB_KEY: [子值]
```

标签是内联源码注释，不是独立的 JSON 账本。通过 `grep` 或 `moai mx query` 收集。

## 标签类型

| 标签 | 用途 | 必需子行 |
|------|------|----------------|
| `@MX:NOTE` | 传递上下文与意图 | — |
| `@MX:WARN` | 标示危险区 | `@MX:REASON` |
| `@MX:ANCHOR` | 不变契约 (高 fan_in) | `@MX:REASON` |
| `@MX:TODO` | 未完成工作 | — |
| `@MX:DEBT` | 有意简化 (可工作代码) | `@MX:CEILING` + `@MX:UPGRADE` |

## 子行

`@MX:SPEC` · `@MX:LEGACY` · `@MX:REASON` · `@MX:TEST` · `@MX:PRIORITY` · `@MX:CEILING` · `@MX:UPGRADE`

- `@MX:REASON` 对 WARN · ANCHOR 是 **必需** 的。
- `[AUTO]` 前缀对代理生成的标签是 **必需** 的。

## 添加时机

**@MX:NOTE** — 魔法常量、超过 100 行的 exported 函数缺 godoc、无说明的业务规则。

**@MX:WARN** — 无 `context.Context` 的 goroutine/channel、循环复杂度 15 及以上、全局状态变更、if 分支 8 个及以上。

**@MX:ANCHOR** — fan_in 3 及以上、公开 API 边界、外部系统集成点。

**@MX:TODO** — 无测试文件的公开函数、未实现的 SPEC 要求、未处理即返回的错误。

**@MX:DEBT** — 采用了有意简化，在明示界限 (`@MX:CEILING`) 内准确，且存在重访触发器 (`@MX:UPGRADE`) 时。

## DEBT — 可工作简化的明确界限

`@MX:DEBT` 不是未完成工作的标示。代码 **已经完成且准确运行**，但记录其为明示界限内的有意简化。附带两个子行。

```go
// @MX:DEBT: in-memory map cache, no eviction
// @MX:CEILING: < 10k entries
// @MX:UPGRADE: switch to LRU when entry count exceeds 10k
```

无 `@MX:UPGRADE` 的 DEBT 没有终止条件，会 **悄悄腐化** (rot)。`moai mx query --kind DEBT --json` 将其显示为 `"rotRisk": "no-trigger"`。腐化信号是 `@MX:UPGRADE` 的缺失；`@MX:CEILING` 的缺失仅是质量备忘，不是腐化的判据。

> `@MX:TODO` 标示在 GREEN 阶段解决的未完成工作（代码尚未完成），`@MX:DEBT` 标示已完成且准确运行但具明示界限的简化（代码已完成）。DEBT 可正常跨多个 GREEN 阶段维持，TODO 的"3 次未解决即升为 WARN"规则不适用。

## 更新/移除时机

- **ANCHOR** — fan_in 变化或 SPEC 更新时更新。禁止自动删除，通过报告降级为 NOTE。
- **NOTE** — 函数签名变更时复审。
- **WARN** — 危险结构改善时移除。
- **TODO** — 解决时（测试通过或实现完成）移除。3 次重复未解决升为 WARN。
- **DEBT** — 界限或触发器变化时更新。`@MX:UPGRADE` 触发器触发且简化被替换时移除，与其他工作完成无关。无自动升级。

## 生命周期总结

```text
TODO     RED/ANALYZE 创建 → GREEN/IMPROVE 解决（移除）→ 3 次未解决升为 WARN
ANCHOR   fan_in ≥ 3 创建 → 调用数·SPEC 变化时更新 → fan_in < 3 时降为 NOTE（报告）→ 无自动删除
WARN     危险检测创建 → 结构性则持续 → 解决时移除
NOTE     需要上下文创建 → 签名变更后更新 → 代码删除时废弃
DEBT     有意简化创建 → UPGRADE 触发器触发时解决（替换简化）→ 无自动升级
```

## 语言注释语法

| 语言 | 前缀 | 示例 |
|------|--------|------|
| Go · Java · TS · Rust · C/C++ · Swift · Kotlin · Dart · Zig · Scala | `//` | `// @MX:NOTE:` |
| Python · Ruby · Elixir | `#` | `# @MX:WARN:` |
| Haskell | `--` | `-- @MX:ANCHOR:` |

## 配置 (`.moai/config/sections/mx.yaml`)

- **thresholds** — `fan_in_anchor`, `complexity_warn`, `branch_warn`
- **limits** — `anchor_per_file` (默认 3), `warn_per_file` (默认 5)。超出时 ANCHOR 从最低 fan_in 降级，WARN 仅保留 P1–P5 优先。
- **exclude** — `**/*_generated.go`, `**/vendor/**`, `**/mock_*.go` 等打标签排除模式
- **require_reason_for** — REASON 必需的标签类型

## 标签语言

标签描述与 `@MX:REASON` 遵循 `.moai/config/sections/language.yaml` 的 `code_comments` 设置（默认 `en`）。若是韩语项目，设为 `code_comments: ko` 则标签以韩语编写。

## 后续步骤

- [Hooks 指南](/zh/advanced/hooks-guide) — 与 hooks 一同处理代码上下文的基础
- [SPEC 基础开发](/zh/core-concepts/spec-based-dev) — SPEC 生命周期与 @MX TAG 联动
- [TRUST 5 质量框架](/zh/core-concepts/trust-5) — Readable 原则与 @MX:NOTE
