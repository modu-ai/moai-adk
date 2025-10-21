<!-- Copilot instructions for AI coding agents in the MoAI-ADK repo -->

# Copilot guidance — MoAI-ADK (concise)

Follow these focused instructions so an AI coding agent becomes productive immediately in this repository.

1. Big picture (what to read first)

- Read `README.md` and `CLAUDE.md` top-to-bottom to understand the 3-step Alfred workflow: `/alfred:1-plan` → `/alfred:2-run` → `/alfred:3-sync`.
- Familiarize with project scaffolding under `moai-adk-ts/` (CLI, core, utils). Key entry: `moai-adk-ts/src/index.ts` and `moai-adk-ts/src/cli/index.ts`.
- SPECs live in `.moai/specs/` and follow a strict TAG/HISTORY front-matter. Example: `.moai/specs/SPEC-AUTH-001/spec.md`.

2. Build / test / dev commands (exact)

- Preferred runtime: Bun (recommended) or Node.js >=18. Use `bun` when present.
- Build: from `moai-adk-ts/` run `bun run build` (falls back to `npm run build`).
- Dev CLI: `bun run dev` (runs `tsx src/cli/index.ts`).
- Tests: `bun run test` or `bun run test:watch` (Vitest). Coverage: `bun run test:coverage`.
- Lint & format: `bun run lint`, `bun run format` (Biome).
- Full CI check: `bun run ci` (clean, build, check, test).

3. Project conventions and patterns

- SPEC-First & TDD-First: Every feature should have `@SPEC`, `@TEST`, `@CODE` tags and a HISTORY section. Search with ripgrep patterns used in docs: `rg '@(SPEC|TEST|CODE|DOC):' -n`.
- JIT Retrieval: Agents should not preload whole files — prefer referencing identifiers and loading concrete files only when needed (see `CLAUDE.md` "JIT Retrieval").
- Hooks use `$CLAUDE_PROJECT_DIR` env var. Always call hooks using that variable (see `CLAUDE.md` hooks examples and `.claude/hooks/*`).
- Directory layout: templates and generators live in `moai-adk-ts/src/core/project/*` (TemplateManager, TemplateProcessor, TemplateValidator).
- System checks and language detection are centralized in `moai-adk-ts/src/core/system-checker` (use `SystemDetector`/`SystemChecker` for environment reasoning).

4. Safe edit rules (must-follow)

- Read files entirely before editing. AGENTS.md enforces this; follow it strictly.
- Keep changes small and add/update tests for behavioral changes. New features require SPEC + tests before implementation.
- Avoid committing secrets; `.moai`, `.claude`, and `CLAUDE.md` are important project files — prefer creating backups before overwriting (see `src/cli/commands/project/backup-merger.ts`).

5. Integration points & external deps

- CLI wiring: `moai` binary -> `moai-adk-ts/dist/cli/index.js`. Changes to CLI require build + smoke test via `bun run dev`.
- Language/toolchain choices are data-driven: look for `moai-adk-ts/src/core/project/template-processor.ts` and `package.json` scripts to infer test/lint commands for target language.
- Uses external tools: Bun, tsx, tsup, Vitest, Biome, Typedoc.

6. Quick examples (what to run locally)

- Build & test snapshot:
  - cd moai-adk-ts && bun run clean && bun run build && bun run test:ci
- Run CLI in dev mode:
  - cd moai-adk-ts && bun run dev -- <moai args>

7. Where to look for examples

- Template generation: `moai-adk-ts/src/core/project/template-manager.ts`
- CLI flows and prompts: `moai-adk-ts/src/cli/commands/init/*` and `src/cli/prompts/*`
- System detection & checks: `moai-adk-ts/src/core/system-checker/*`
- Agent policies and command recipes: `.claude/commands/alfred/` and `AGENTS.md`

If anything here is ambiguous or you want deeper examples (unit test style, typical SPEC layout, or common patch patterns), tell me which area to expand and I will iterate.
