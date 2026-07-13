---
title: Large Codebases
weight: 80
draft: false
description: "Context-narrowing strategies for using Claude Code efficiently in multi-million-line single trees or multi-package monorepos."
---

# Large Codebases

Claude Code works well even on large codebases — whether a multi-million-line single repository or a monorepo of many packages. But the defaults assume a small project, so a **strategy of narrowing context to only what each task actually touches** is essential.

{{< callout type="info" >}}
**One-line summary**: The real problem in a large codebase is not "many files" but **irrelevant guidance and files filling the context**. Irrelevant tokens degrade quality while raising cost — context narrowing IS tokenomics.
{{< /callout >}}

## Choosing Where to Start

Where you run `claude` determines everything that follows.

| Starting location | File access scope | CLAUDE.md loaded | Best when |
|---------|-----------|---------------|---------|
| **Repository root** | Everything | Root only (subdirectories on demand) | Work spanning multiple packages/subsystems |
| **A subdirectory** | Only that subtree | That directory + every ancestor | Work confined to one package/subsystem |

If the work focuses on one package (say `packages/api/`), run `claude` in that directory. Guidance from `packages/web/` never loads in the first place, so the context lightens on its own without any effort spent pruning rules.

## Splitting CLAUDE.md by Directory

Cramming every rule into a single root CLAUDE.md creates three problems:

- It gets too long, hurting readability
- Trying to apply to every package makes it too generic to be useful
- Guidance irrelevant to the task loads every session anyway

The fix is layering: keep repo-wide rules at the root and put each area's rules in its own subdirectory.

```markdown
# ./CLAUDE.md (root, loaded in every session)
This is a monorepo with three packages:
- packages/api: Node.js REST API with Express, TypeScript, PostgreSQL
- packages/web: React frontend with Vite, TypeScript, TailwindCSS
- packages/shared: shared TypeScript utilities

Run commands from the package directory.
```

```markdown
# ./packages/api/CLAUDE.md (loaded only when working in this directory)
This package is the REST API server.

- Run tests: `npm test` (uses Vitest)
- Run dev server: `npm run dev` (port 3001)
- Database migrations: `npm run migrate`

API routes are in src/routes/. Never write raw SQL in handlers.
```

When Claude starts in `packages/api/`, the root and `packages/api/` CLAUDE.md files both load, but the guidance in `packages/web/` does **not**.

## Excluding Irrelevant CLAUDE.md Files

Skip guidance from other teams' packages or legacy code with the `claudeMdExcludes` setting.

```json
{
  "claudeMdExcludes": [
    "**/packages/admin-dashboard/**",
    "**/packages/legacy-*/**"
  ]
}
```

The root CLAUDE.md still loads; only the excluded packages' guidance leaves the context.

## Blocking Generated and Vendor Code

Paths already in `.gitignore` (node_modules, dist, build) are automatically excluded from search results.

For committed generated code or vendor SDKs, block reads outright with permission rules. Generated files are long and repetitive, making them an especially large context waste.

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

## Code Intelligence (LSP) Plugins

Reading a file line by line to locate a symbol definition is the most expensive exploration in token terms. Install a language-server plugin and go-to-definition, find-references, and direct type-error queries become available, dramatically cutting the file reads themselves.

```bash
/plugin install typescript-lsp@claude-plugins-official
```

- Major languages are supported: TypeScript, Python, Go, Rust, and more
- The language's LSP binary must be installed on the system (see the [plugins document](/claude-code/extensibility/plugins))

## Checking Out Only the Needed Directories with Worktrees

Worktrees created via `--worktree` can check out **only the listed directories** — not everything — with the `worktree.sparsePaths` setting.

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

- Creation gets faster (only the needed parts instead of a full copy)
- Disk space is saved
- `symlinkDirectories` can also eliminate node_modules duplication across worktrees.

```json
{
  "worktree": {
    "sparsePaths": ["packages/api", "packages/shared"],
    "symlinkDirectories": ["node_modules"]
  }
}
```

Directories listed under `symlinkDirectories` are shared as symbolic links to the main checkout's copy.

## Granting Access to Other Packages/Repositories

If you started in one package but need to modify a sibling package, widen access with `additionalDirectories`.

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

Runtime flags work instead of settings, too.

```bash
claude --add-dir ../shared --add-dir ../web
```

## Per-Package Skills

Each package can have skills for its own area. Skills load only when needed, making them a good vessel for package-specific knowledge with no context burden.

```bash
mkdir -p packages/api/.claude/skills/api-testing
```

```markdown
# packages/api/.claude/skills/api-testing/SKILL.md
---
name: api-testing
description: Test patterns for the API package
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

Work in `packages/api` and the api-testing skill loads automatically; in `packages/web`, it does not.

## Coordinating Cross-Package Work

When one change touches multiple packages (say, updating a shared type and fixing every call site), two principles apply:

- **Handle the whole change in one session**: load the related files together to keep decisions consistent.
- **Save the plan to a file first**: leave the plan in a markdown file. Long sessions get their context compacted, but a plan saved to disk never disappears. "Persist important state to files" is also a fundamental of operating agentic loops.

## A Concrete Configuration Example: Monorepo

Here is a complete configuration example. The root carries repo-wide deny rules, and each package carries its own worktree and access settings (in a MoAI-ADK project, workflow settings like `.moai/config/sections/workflow.yaml` also live at the root).

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

The effect of this configuration:

- Worktrees check out only `.claude/`, `packages/api/`, and `packages/shared/`
- The shared package is accessible
- Generated/vendor file access is blocked

## Tips and Tricks

### Scoped Searches

Before making a big change, map the blast radius first. The habit of narrowing search scope reduces how many files must be read.

```bash
grep -r "FunctionName" packages/api/  # search only api
grep -r "FunctionName" packages/      # all packages
```

### Layer-by-Layer Analysis

For changes touching multiple layers — DB, API, UI — understand each layer separately, and focus one session on one change.

### Documentation Directives

So docs do not go stale after a large change, include a "update docs" item in the change plan.

## Related Documents

- [Context Window](/claude-code/context-memory/context-window)
- [Worktrees](/claude-code/agentic/worktrees)
- [Best Practices](/claude-code/agentic/best-practices)

## References

- [Set up Claude Code in a monorepo or large codebase (official docs)](https://code.claude.com/docs/en/large-codebases)
- [Best practices for Claude Code (official docs)](https://code.claude.com/docs/en/best-practices)

{{< callout type="tip" >}}
The easiest first move in a monorepo: "for single-package work, run `claude` in that package's directory." It cuts irrelevant guidance loading without touching a single config file — the highest-return habit for the cost.
{{< /callout >}}
