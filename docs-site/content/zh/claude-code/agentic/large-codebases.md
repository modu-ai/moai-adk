---
title: 大型代码库
weight: 80
draft: false
description: "整理在数百万行的单一代码树或多包 monorepo 中，让 Claude Code 聚焦于工作范围、收窄上下文的策略。"
---

大型代码库(数百万行单一仓库，或多包 monorepo)中 Claude Code 可以正常工作。但默认配置针对小项目优化，所以**将上下文收窄到当前工作涉及的部分的策略**是必需的。

{{< callout type="info" >}}
**核心**：大型代码库的问题不是"读取全部文件"。而是**与当前工作无关的指令和文件填满上下文**。
{{< /callout >}}

## 1. 确定起点

在哪里运行 `claude` 决定了一切。

| 启动位置 | 文件访问范围 | 加载的 CLAUDE.md | 适用场景 |
|---------|-----------|---------------|---------|
| **仓库根目录** | 全部 | 仅根目录(下层按需加载) | 跨多个包/子系统的工作 |
| **子目录** | 仅该子树 | 该目录 + 所有上级目录 | 限于单个包/子系统的工作 |

**提示**：如果只关注一个包(如 `packages/api/`)，就在那个目录运行 `claude`。这样 `packages/web/` 的指令就不会被加载。

## 2. 按目录拆分 CLAUDE.md

如果把所有规则都放在根目录：
- 太长，可读性差
- 太通用，没有用处
- 加载无关的指令

**解决**：在根目录放仓库全局规则，在每个子目录放该区域的规则。

```markdown
# ./CLAUDE.md (根，所有会话都加载)
This is a monorepo with three packages:
- packages/api: Node.js REST API with Express, TypeScript, PostgreSQL
- packages/web: React frontend with Vite, TypeScript, TailwindCSS
- packages/shared: shared TypeScript utilities

Run commands from the package directory.
```

```markdown
# ./packages/api/CLAUDE.md (该目录工作时才加载)
This package is the REST API server.

- Run tests: `npm test` (uses Vitest)
- Run dev server: `npm run dev` (port 3001)
- Database migrations: `npm run migrate`

API routes are in src/routes/. Never write raw SQL in handlers.
```

Claude 从 `packages/api/` 启动时：
- 根目录和 packages/api/ CLAUDE.md 都会加载
- packages/web/ 的指令**不会加载**

## 3. 排除无关的 CLAUDE.md

其他团队的包或遗留代码用 `claudeMdExcludes` 跳过：

```json
{
  "claudeMdExcludes": [
    "**/packages/admin-dashboard/**",
    "**/packages/legacy-*/**"
  ]
}
```

根目录 CLAUDE.md 仍会加载，被排除的包无法访问。

## 4. 屏蔽生成代码和第三方代码

`.gitignore` 中已有的路径(node_modules、dist、build)会自动从搜索结果中排除。

提交到仓库的生成代码或第三方 SDK 用权限规则屏蔽：

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

## 5. 代码智能 (LSP) 插件

逐行读取文件找符号定义效率很低。安装语言服务器插件后：

```bash
/plugin install typescript-lsp@claude-plugins-official
```

Claude 能进行 `go to definition`、`find references`、直接查询类型错误。

- 支持 TypeScript、Python、Go、Rust 等主要语言
- 需要 LSP 二进制文件(参考指南)

这样能大幅减少文件读取。

## 6. 用 Worktree 只检出需要的目录

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

用 `--worktree` 创建的工作树不检出全部，而是**仅检出列表中的目录**。

- 快速创建(全部克隆 vs 仅需部分)
- 节省磁盘空间
- 多个工作树的 node_modules 去重：

```json
{
  "worktree": {
    "sparsePaths": ["packages/api", "packages/shared"],
    "symlinkDirectories": ["node_modules"]
  }
}
```

## 7. 授予访问其他包/仓库的权限

从一个包启动但需要修改兄弟包时：

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

或在运行时：

```bash
claude --add-dir ../shared --add-dir ../web
```

## 8. 为每个包添加 Skills

每个包可以有仅限该区域的自动化命令(Skills)。

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

在 packages/api 中工作时自动加载 api-testing 技能，在 packages/web 中不会加载。

## 9. 协调跨包工作

同一变更影响多个包时(如：更新共享类型 + 修改所有调用处)：

**在一个会话中处理全部变更**：一次性加载所有文件，保持决策一致。

**事先编写计划**：把计划保存到 markdown 文件。会话变长后上下文会被压缩，但保存的计划不会消失。

## 10. 具体配置示例：Monorepo

下面是完整的配置示例。

**根目录**(其他设置如 `.moai/config/sections/workflow.yaml` 也放根目录)：

```json
// .claude/settings.json
{
  "permissions": {
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)"
    ]
  }
}
```

**packages/api** (`.claude/settings.json`)：

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

这样配置后：
- 仅检出 `.claude/`、`packages/api/`、`packages/shared/` (worktree)
- 可以访问 shared 包
- 屏蔽生成/第三方文件

## 11. 大型代码库技巧

### 按范围搜索

进行大规模变更时，先掌握影响范围：

```bash
grep -r "FunctionName" packages/api/  # 仅在 api 中搜索
grep -r "FunctionName" packages/      # 在所有包中搜索
```

### 按层分析

涉及多层(数据库、API、UI)的变更时，分别理解各层，但一个会话中只集中一个变更。

### 文档化指示

大规模变更后为保持文档更新，在变更计划中加入"修改文档"项。

## 参考

此指南基于 Anthropic 官方文档 [Set up Claude Code in a monorepo or large codebase](https://code.claude.com/docs/en/large-codebases) 编写。

更多策略请参考 Anthropic 的 [Best practices for Claude Code](https://code.claude.com/docs/en/best-practices) 文档。
