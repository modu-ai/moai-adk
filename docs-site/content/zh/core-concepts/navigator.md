---
title: 导航器绑定令牌
weight: 25
draft: false
---
# 导航器绑定令牌

让代码和文档互相指向,代理修改一方时就能立即拉起另一方的上下文。**导航器绑定令牌** (Navigator Binding Tokens)是三个编写用令牌,将设计决策·代码符号·SPEC 连接成一个可寻址的图。这些令牌聚在一起形成单个产出物 `.moai/project/navigator/nav-graph.json`。

## 三个令牌

导航器集成层将三组绑定令牌合并到一个图中。

| 组 | 令牌形式 | 编写位置 | 角色 |
|------|----------|----------|------|
| `NAV:DEC` | `@NAV:DEC-<id>` | 设计文档(`.moai/project/*.md`, `.moai/docs/**/*.md`) | 将设计决策连接到 SPEC 或符号 |
| `NAV:SYM` | `@NAV:SYM:<symbol>` | 代码注释 + 设计文档 | 将文档位置连接到代码的命名符号 |
| `MX:SPEC` | `@MX:SPEC:<SPEC-ID>` | 代码注释(`@MX:` 标签的子行) | 将代码位置连接到 SPEC |

`MX:SPEC` 已经由 [MX 标签系统](/zh/advanced/mx-tags/)处理。导航器集成层**只消费** MX 扫描器的 `SpecAssociator` 输出,不重新扫描。所以不要新写这个令牌,遵循现有 MX 标签规则。

## 何时编写令牌

### 编写 `@NAV:DEC-<id>` 时

- `.moai/project/tech.md`、`structure.md`、`product.md` 或 `.moai/docs/` 下的设计文档中,某个决策对应特定 SPEC 或代码符号时。
- 希望以后修改代码时该决策的上下文重新浮现时。

### 编写 `@NAV:SYM:<symbol>` 时

- 文档位置或代码注释需要绑定到命名代码符号,以便阅读图的人可以从文档到代码(或符号到符号)移动时。

`@MX:SPEC:` 在这里不编写。它已经是 mx-scanner 表面。重新编写是不必要的。

## 令牌语法

两个令牌都不应包含空值。扫描器遇到空值时会在 `.moai/logs/navigator-sync.log` 留下诊断警告并跳过该项,但不会停止整个图构建(fail-open)。

### `@NAV:DEC-<id>`

`<id>` 必须匹配 `[A-Z][A-Z0-9-]*`。只允许大写 ASCII 和数字,以及内部连字符。与 SPEC-ID 域令牌一致的规则。`@NAV:DEC-` 前缀是不二义的判别器,所以 id 本身不会在没有前缀的情况下出现。

### `@NAV:SYM:<symbol>`

`<symbol>` 必须匹配 `[A-Za-z_][A-Za-z0-9_.]*`。只要是标识符形状即可,语言中立。包限定形式(`pkg.ParseHeader`)是惯例,短形式(`ParseHeader`)也接受,通过对现有符号集的后缀匹配来解析。

## 扫描根

导航器集成层扫描以下表面。

- **设计文档** — `.moai/project/{product,structure,tech}.md` 和 `.moai/docs/**/*.md`。
- **代码** (仅 `@NAV:SYM`) — 排除 `*_test.go` 和 `vendor/` 的 Go `*.go` 文件。设计文档表面也一起。

以下**不扫描**。

- `.moai/specs/` — 已经由 mx-scanner 的基于正文的关联覆盖。
- `.moai/reports/`, `.moai/state/` — 临时的或运行时状态。
- 现有三个导航器链的源代码(仅消费)。

## 产出物 — `nav-graph.json`

产出为单个文件 `.moai/project/navigator/nav-graph.json`。形状如下。

```json
{
  "provenance": { "extract_commit_sha": "...", "captured_at": "..." },
  "nodes": [
    { "entity_type": "decision", "identifier": "...", "display_name": "..." }
  ],
  "edges": [
    { "edge_type": "dec-edge", "source_node": "...", "target_node": "...", "source_path": "...", "line_number": 0 }
  ]
}
```

`entity_type` 是 `decision | spec | symbol` 之一,`edge_type` 是 `dec-edge | spec-edge | sym-edge` 之一。

这个产出物是**字节稳定**的。在同一个 git HEAD 上运行两次会得到字节相同的结果。不打印墙上时间戳,所以不管谁何时运行结果都相同。审计和可重现性构建在这个性质之上。

{{< callout type="info" >}}
**fail-open** — 图构建始终返回退出码 0。即使有错误令牌也不会中断,只留下诊断警告,构建健全部分的图。
{{< /callout >}}

## 编写示例

设计文档中指向决策和符号,代码注释中引用相同的决策和符号的最简单形状。

设计文档(`tech.md`):

```markdown
# Tech

会话层采用 OAuth2 进行委托访问。

决策 @NAV:DEC-AUTH-STRATEGY: OAuth2 over client-credentials。

头部解析器(see @NAV:SYM:pkg.ParseHeader)提取 Bearer 令牌。
```

代码(`auth/auth.go`):

```go
package auth

// @NAV:DEC-AUTH-STRATEGY: 实现 OAuth2 client-credentials 流。
// @NAV:SYM:auth.ParseBearer 从 Authorization 头提取 Bearer 令牌。
func ParseBearer(h string) string { ... }
```

这两个文件的图会创建三个节点(决策 `AUTH-STRATEGY`、符号 `pkg.ParseHeader`、符号 `auth.ParseBearer`)和它们之间的边。阅读图的人可以从设计文档到代码,从代码到设计依据自由移动。

## 向前兼容性

令牌语法、绑定记录的 5-字段形状、图模式都是向前兼容的(仅添加)。后续里程碑可能添加字段,但不会改变现有字段的名称和形状。一次编写的令牌长期有效。

## 相关文档

- [MX 标签系统](/zh/advanced/mx-tags/) — `@MX:SPEC` 令牌的源规则。导航器集成层消费这个输出。
- [基于 SPEC 的开发](/zh/core-concepts/spec-based-dev/) — SPEC 生命周期和 `@MX:SPEC` 的上层上下文。
- [代理指南](/zh/advanced/agent-guide/) — 代理如何在代码注释和设计文档之间移动。
