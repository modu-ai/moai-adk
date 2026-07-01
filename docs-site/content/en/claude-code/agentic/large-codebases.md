---
title: Large Codebases
weight: 80
draft: false
description: "Strategies for efficiently using Claude Code in million-line single trees or multi-package monorepos by narrowing context to only what the current task touches."
---

Claude Code works regardless of scale, but as a codebase grows, the default behavior tuned for small projects starts to cause problems. Instructions and file reads unrelated to your task fill the context window, waste tokens, and ultimately degrade response quality.

{{< callout type="info" >}}
**Core Principle**: The key to a large codebase is not "having everything read" but "loading only the part your current task touches into context."
{{< /callout >}}

## 1. Start Location Determines Context Scope

Where you run `claude` determines both the file access scope and the range of `CLAUDE.md` loaded at startup. It is the first thing to decide.

| Start location | File access | CLAUDE.md loaded at startup | When it fits |
| --- | --- | --- | --- |
| **Repository root** | All files | Root only (subdirectories on demand when read) | Work spans multiple packages or subsystems |
| **Subdirectory** | That subtree only | That directory plus all parent directories | Work is confined to one package or subsystem |

**Tip**: If you are focused on a single package (e.g., `packages/api/`), just run `claude` from that directory. Other packages' instructions automatically stay out of context.

## 2. Splitting CLAUDE.md by Directory

When you cram every rule into a single root `CLAUDE.md`, it either bloats by carrying every subsystem's rules or becomes useless by being too generic. When you split instructions per directory, Claude loads the repository-wide rules plus **only the rules for the code you're working on right now**.

**Root CLAUDE.md** (loaded by all sessions):

```markdown
# ./CLAUDE.md (root, loaded by all sessions)
This is a monorepo with three packages:
- packages/api: Node.js REST API with Express, TypeScript, PostgreSQL
- packages/web: React frontend with Vite, TypeScript, TailwindCSS
- packages/shared: shared TypeScript utilities

Run commands from the package directory.
```

**Package-specific CLAUDE.md** (loaded only when working in that directory):

```markdown
# ./packages/api/CLAUDE.md (loaded only when working in this directory)
This package is the REST API server.

- Run tests: `npm test` (uses Vitest)
- Run dev server: `npm run dev` (port 3001)
- Database migrations: `npm run migrate`

API routes are in src/routes/. Never write raw SQL in handlers.
```

When you start in `packages/api/`, the root + packages/api/ CLAUDE.md load together, while `packages/web/`'s instructions never enter context. Commit these files to the repository so teammates can share them.

## 3. Excluding Irrelevant CLAUDE.md

When starting from the root, a subdirectory's `CLAUDE.md` loads the moment you read a file there. For areas you never work on—such as another team's package or legacy code—you can block them entirely with `claudeMdExcludes`.

```json
{
  "claudeMdExcludes": [
    "**/packages/admin-dashboard/**",
    "**/packages/legacy-*/**"
  ]
}
```

Patterns are matched as globs against absolute paths, so to match anywhere in the tree, start with `**/`. For personal use, put them in `.claude/settings.local.json`.

## 4. Blocking Generated and Vendor Code

Claude's content search respects `.gitignore` by default, so `node_modules/`, `dist/`, and `build/` are excluded from search results without any extra configuration. Vendor SDKs or generated code that is committed to the repository can be blocked with `Read` deny rules in `permissions.deny`.

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

Deny rules cover both Claude's built-in file tools and recognizable Bash commands such as `cat`, `head`, `grep`, and `find`.

## 5. Code Intelligence (LSP) for Symbol Lookup

Finding a symbol's definition or call sites can balloon into many file reads and grep calls. Attaching a code intelligence plugin (LSP-based) lets Claude query the language server directly for go-to-definition, find-references, and type-error checks.

```bash
/plugin install typescript-lsp@claude-plugins-official
```

The official marketplace provides plugins for major languages such as TypeScript, Python, Go, and Rust. Each developer machine must have that language's language server binary installed.

This feature pairs well with `claudeMdExcludes` and `Read` deny rules: the first two push irrelevant content out of context, while code intelligence keeps Claude from reading files to find definitions.

## 6. Narrowing Worktree Scope with Sparse Checkout

The `--worktree` flag starts a session in a new worktree to isolate changes from the main checkout. You can apply a git sparse checkout with `worktree.sparsePaths` to check out only the needed directories to disk.

```json
{
  "worktree": {
    "sparsePaths": [
      ".claude",
      "packages/api",
      "packages/shared"
    ],
    "symlinkDirectories": ["node_modules"]
  }
}
```

Paths are relative to the repository root. List directories to check out; root-level files such as `package.json` are always included. Adding `symlinkDirectories` shares large directories like `node_modules` via symbolic links instead of duplicating them across worktrees.

Benefits:
- Faster worktree creation (partial checkout vs full)
- Reduced disk space usage
- Eliminate `node_modules` duplication across multiple worktrees

```json
{
  "worktree": {
    "sparsePaths": ["packages/api", "packages/shared"],
    "symlinkDirectories": ["node_modules"]  // share main node_modules
  }
}
```

## 7. Additional Directory Access for Cross-Package Work

When starting from one package directory but needing access to sibling packages:

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

Or grant access at invocation time:

```bash
claude --add-dir ../shared --add-dir ../web
```

This lets you maintain per-package isolation while enabling explicit cross-package collaboration.

## 8. Adding Package-Specific Skills

Each package can have automation commands (Skills) specific to that area.

```bash
mkdir -p packages/api/.claude/skills/api-testing
```

```markdown
# packages/api/.claude/skills/api-testing/SKILL.md
---
name: api-testing
description: API package testing patterns
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

When working from `packages/api/`, the api-testing skill loads automatically. When working from `packages/web/`, it does not.

## 9. Coordinating Cross-Package Changes

When a change spans multiple packages—such as fixing a shared type and all its call sites—two strategies help maintain consistency.

**One session for the entire change**: Handle the shared edit and its call sites together so the rationale for each edit stays consistent rather than re-deriving it per package.

**Save the plan beforehand**: Write and save the plan as a markdown file. Long sessions compact the context partway through, but a saved plan survives even when the conversation history is lost.

## 10. Concrete Monorepo Configuration Example

Here is a complete setup for a monorepo.

**Root** (`.claude/settings.json`):

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

**packages/api** (`.claude/settings.json`):

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

With this configuration:
- `.claude/`, `packages/api/`, and `packages/shared/` are checked out only (worktree sparse)
- Shared package is accessible
- Generated and vendor files are blocked

## 11. Tips and Tricks for Large Codebases

### Scope-Based Search

When making large changes, understand the impact scope first:

```bash
grep -r "FunctionName" packages/api/  # search api only
grep -r "FunctionName" packages/      # search all packages
```

### Layer-by-Layer Analysis

When a change touches multiple layers (database, API, UI), understand each layer separately, then focus on one change per session.

### Documentation Directives

After large changes, keep documentation synchronized. Add "update docs" to your change plan so documentation stays current with code changes.

## References

This guide is based on Anthropic's official [Set up Claude Code in a monorepo or large codebase](https://code.claude.com/docs/en/large-codebases) documentation. See also [Best practices for Claude Code](https://code.claude.com/docs/en/best-practices).
