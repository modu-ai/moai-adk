# internal/template — Module Conventions

## Purpose

`internal/template` owns the `go:embed` template system that powers `moai init` and `moai update`. Templates live under `internal/template/templates/` and are compiled into the binary via the `//go:embed all:templates` directive in `embed.go` (there is NO generated `.go` file). The package renders these templates with project-specific values (`GoBinPath`, `HomeDir`, project name) and deploys them to the user's project root.

This package is the contract between maintainer (who edits source templates) and user (who runs `moai init`). The 16-language neutrality policy and the `moai update` namespace protection contract both bind here.

## Conventions

- **Template-First Rule (CLAUDE.local.md §2 [HARD])**: All template changes MUST be made in `internal/template/templates/` FIRST, then `make build` (`gen-catalog-hashes --all` + `go build`) recompiles the binary; templates are embedded via `//go:embed all:templates`. Never add files directly to the local project's `.claude/`, `.moai/`, or `.agency/` without corresponding source in `templates/`. Verification: before committing, every new file under those dirs must have a sibling under `templates/`.
- **The embedded FS is NOT a generated file** — there is no `embedded.go` to edit. The embedded filesystem comes from the `//go:embed all:templates` + `//go:embed catalog.yaml` directives in `embed.go`. The source of truth is `templates/` itself: edit files there, then `make build` to recompile. Editing anywhere else is overwritten on the next `moai update` sync and will fail CI mirror-parity checks.
- **Language neutrality (CLAUDE.local.md §15 [HARD])**: `templates/` content treats all 16 supported languages equally (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift). Never elevate one language as "PRIMARY" or "enabled" while others are "planned". Dart canonical name is **"flutter"** (not "dart").
- **`.HomeDir`/`.GoBinPath` in `.sh.tmpl` (CLAUDE.local.md §14 [HARD])**: These template variables resolve at `moai init` time and become baked absolute paths. NEVER use them in fallback paths inside generated shell scripts — fallbacks MUST use `$HOME` for portability (e.g., `$HOME/go/bin/moai`, not `{{.HomeDir}}/go/bin/moai`).
- **`settings.local.json` separation (CLAUDE.local.md §2 [HARD])**: Runtime-managed file. NEVER include it in `templates/`. `settings.json.tmpl` renders `settings.json`; the `.local` variant is created at runtime by `moai cg`/`moai glm`/SessionStart hook with per-machine secrets.
- **`moai update` namespace contract (CLAUDE.local.md §24.4 [HARD])**: `update.go` enforces strict separation: `moai-*` skills + `.claude/agents/{core,expert,meta}/` are template-managed (overwrite-safe), while `my-harness-*` skills + `.claude/agents/harness/` are user-owned (NEVER delete/modify, ALWAYS backup). The `templates/.claude/agents/harness/` directory MUST NOT exist — its appearance is a §24.2 contract violation requiring cleanup chore.

## Key Patterns

- **`TemplateContext` injection** (`renderer.go`): `template.NewTemplateContext(WithGoBinPath(...), WithHomeDir(...))` constructs the context. Pass through `deployer.Deploy(ctx, projectRoot, mgr, ctx)`. Adding a new template variable means: (1) extend `TemplateContext` struct, (2) add `With*` option, (3) update `renderer_test.go` golden, (4) `make build`.
- **`posixPath` helper**: Use `{{posixPath .GoBinPath}}` in `.sh.tmpl` to force forward-slash separator even when rendered on Windows. Critical for shell scripts that must run on git-bash / wsl.
- **Protected directories during sync**: `.claude/` (local config), `.moai/project/` (user docs product.md/structure.md/tech.md), `.moai/specs/` (active SPECs) — `update.go` MUST preserve these. The protection is encoded in `update_archive.go` exclude lists; any new template path that intersects with these triggers a backup-first flow.
- **Mirror parity checks**: For SSOT docs that exist in both `.claude/rules/.../*.md` and `templates/.claude/rules/.../*.md`, byte-identity is verified by `internal/template/rule_template_mirror_test.go`. Both sides MUST be edited in the same commit.

## Cross-References

- Root CLAUDE.md §2 (Request Processing Pipeline), §9 (Configuration Reference)
- CLAUDE.local.md §2 (File Synchronization + Template-First Rule), §14 (Hardcoding prevention), §15 (Template language neutrality), §22 (Dev settings intent), §24 (Harness namespace separation), §24.4 (`moai update` contract)
- `internal/template/templates/` — single source of truth for deployable assets
- `internal/template/embed.go` — the `//go:embed all:templates` + `//go:embed catalog.yaml` directives (there is no generated `embedded.go`)
- `internal/template/rule_template_mirror_test.go` — byte-parity invariant for SSOT mirrors
- `internal/cli/update.go` + `update_archive.go` — namespace protection implementation
