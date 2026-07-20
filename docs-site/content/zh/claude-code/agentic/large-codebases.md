---
title: 大型代码库
weight: 80
draft: false
description: "整理在数百万行的单一树或多包 monorepo 中高效使用 Claude Code 的上下文收窄策略。"
---

# 大型代码库

无论是数百万行的单一仓库，还是由多个包组成的 monorepo，Claude Code 在大型代码库中同样运转良好。只是默认配置假定的是小项目，因此**把上下文只收窄到每项工作实际触及的部分**的策略必不可少。

{{< callout type="info" >}}
**一句话总结**：大型代码库的真正问题不是"文件多"，而是**与当前工作无关的指令与文件填满了上下文**。无关的令牌既拉低质量又抬高成本 —— 上下文收窄就是代币经济学。
{{< /callout >}}

## 决定启动位置

在哪里运行 `claude`，决定了之后的一切。

| 启动位置 | 文件访问范围 | 加载的 CLAUDE.md | 适合的情形 |
|---------|-----------|---------------|---------|
| **仓库根** | 全部 | 仅根（下级按需） | 跨多个包/子系统的工作 |
| **子目录** | 仅该子树 | 该目录 + 所有上级目录 | 限于一个包/子系统的工作 |

若工作只聚焦于一个包（例如 `packages/api/`），就在该目录运行 `claude`。`packages/web/` 的指令根本不会被加载，无需费力删规则，上下文自然变轻。

## 按目录拆分 CLAUDE.md

把所有规则塞进一个根 CLAUDE.md 会产生三个问题。

- 过长导致可读性下降
- 为通用于所有包而写得太笼统，失去用处
- 与工作无关的指令也在每个会话被加载

解法是分层：根只放仓库全局规则，各子目录放该领域的规则。

```markdown
# ./CLAUDE.md（根，所有会话加载）
This is a monorepo with three packages:
- packages/api: Node.js REST API with Express, TypeScript, PostgreSQL
- packages/web: React frontend with Vite, TypeScript, TailwindCSS
- packages/shared: shared TypeScript utilities

Run commands from the package directory.
```

```markdown
# ./packages/api/CLAUDE.md（仅在此目录工作时加载）
This package is the REST API server.

- Run tests: `npm test` (uses Vitest)
- Run dev server: `npm run dev` (port 3001)
- Database migrations: `npm run migrate`

API routes are in src/routes/. Never write raw SQL in handlers.
```

当 Claude 从 `packages/api/` 启动时，根与 `packages/api/` 的 CLAUDE.md 都会加载，但 `packages/web/` 的指令**不会加载**。

## 排除无关的 CLAUDE.md

其他团队的包或遗留代码的指令，用 `claudeMdExcludes` 配置跳过。

```json
{
  "claudeMdExcludes": [
    "**/packages/admin-dashboard/**",
    "**/packages/legacy-*/**"
  ]
}
```

根 CLAUDE.md 仍然加载，只有被排除的包的指令从上下文中剔除。

## 拦截生成代码与第三方代码

已在 `.gitignore` 中的路径（node_modules、dist、build）会自动从搜索结果中排除。

对已提交的生成代码或第三方 SDK，用权限规则从读取层面直接拦截。生成文件又长又重复，对上下文的浪费尤其大。

```json
{
  "permissions": {
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)",
      "Read(./**/*.generated.*)",
      "Read(./vendor/**)"
    ]
  }
}
```

## 代码智能 (LSP) 插件

为找一个符号定义而逐行读文件，是令牌视角下最昂贵的探索。装上语言服务器插件后，跳转定义、查找引用、直接查询类型错误都成为可能，文件读取本身可以大幅减少。

```bash
/plugin install typescript-lsp@claude-plugins-official
```

- 支持 TypeScript、Python、Go、Rust 等主要语言
- 系统中须已安装对应语言的 LSP 二进制（参考[插件文档](/claude-code/extensibility/plugins)）

## 用工作树只检出需要的目录

以 `--worktree` 创建的工作树可通过 `worktree.sparsePaths` 配置只检出**列出的目录**而非全部。

```json
{
  "worktree": {
    "sparsePaths": [
      ".claude",
      "packages/api",
      "packages/shared"
    ]
  }
}
```

- 创建更快（只取需要的部分而非整体复制）
- 节省磁盘空间
- 还可用 `symlinkDirectories` 消除多个工作树的 node_modules 重复。

```json
{
  "worktree": {
    "sparsePaths": ["packages/api", "packages/shared"],
    "symlinkDirectories": ["node_modules"]
  }
}
```

列在 `symlinkDirectories` 中的目录会以符号链接共享主检出中的内容。

## 授予对其他包/仓库的访问权

从一个包启动后又需要修改兄弟包时，用 `additionalDirectories` 扩大访问范围。

```json
{
  "permissions": {
    "additionalDirectories": [
      "../shared",
      "../web"
    ]
  }
}
```

也可以不用配置而用运行时标志。

```bash
claude --add-dir ../shared --add-dir ../web
```

## 按包添加技能

每个包可以拥有该领域专属的技能。技能只在需要时加载，是无上下文负担地存放包专属知识的好容器。

```bash
mkdir -p packages/api/.claude/skills/api-testing
```

```markdown
# packages/api/.claude/skills/api-testing/SKILL.md
---
name: api-testing
description: API 包的测试模式
---

## Test structure
Tests are in `src/__tests__/` mirroring `src/`.

## Running tests
- All: `npm test`
- Single file: `npm test -- src/__tests__/routes/users.test.ts`

## Test utilities
- `src/__tests__/helpers/db.ts`: setupTestDb(), teardownTestDb()
- `src/__tests__/helpers/auth.ts`: createTestUser(), getAuthToken()
```

在 `packages/api` 工作时 api-testing 技能自动加载，在 `packages/web` 则不会。

## 跨包工作的协调

当同一变更触及多个包时（例如更新共享类型并修改所有调用处），两条原则有效。

- **在一个会话中处理完整变更**：一次性加载相关文件，保持决策一致性。
- **先把计划存成文件**：把计划留在 Markdown 文件里。会话拉长后上下文会被压缩，但存到磁盘的计划不会消失。"重要状态留在文件里"也是运营智能体循环的基本功。

## 具体配置示例：monorepo

以下是完整配置示例。根放仓库全局拦截规则，包放该包的工作树·访问配置（MoAI-ADK 项目的话，`.moai/config/sections/workflow.yaml` 这类工作流配置也位于根）。

**根**（`.claude/settings.json`）：

```json
{
  "permissions": {
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)"
    ]
  }
}
```

**packages/api**（`.claude/settings.json`）：

```json
{
  "worktree": {
    "sparsePaths": [
      ".claude",
      "packages/api",
      "packages/shared"
    ],
    "symlinkDirectories": ["node_modules"]
  },
  "permissions": {
    "additionalDirectories": ["../shared"],
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)"
    ]
  }
}
```

这套配置的效果如下。

- 工作树只检出 `.claude/`、`packages/api/`、`packages/shared/`
- 可访问 shared 包
- 拦截对生成/第三方文件的访问

## 技巧与窍门

### 限定范围的搜索

在做大变更之前先掌握影响范围。收窄搜索范围的习惯会减少需要读的文件数。

```bash
grep -r "FunctionName" packages/api/  # 只搜 api
grep -r "FunctionName" packages/      # 所有包
```

### 分层分析

若变更触及 DB·API·UI 等多个层，就分别理解每一层，并在一个会话中只专注一项变更。

### 文档化指令

为了大变更之后文档不至陈旧，在变更计划中加入"修改 docs"条目。

## 相关文档

- [上下文窗口](/claude-code/context-memory/context-window)
- [工作树](/claude-code/agentic/worktrees)
- [最佳实践](/claude-code/agentic/best-practices)

## 参考资料

- [Set up Claude Code in a monorepo or large codebase（官方文档）](https://code.claude.com/docs/en/large-codebases)
- [Best practices for Claude Code（官方文档）](https://code.claude.com/docs/en/best-practices)

{{< callout type="tip" >}}
在 monorepo 里最省力的第一步是"做某个包的工作就在该包目录运行 `claude`"。一个配置文件都不用碰就切断了无关指令的加载 —— 性价比最高的习惯。
{{< /callout >}}
