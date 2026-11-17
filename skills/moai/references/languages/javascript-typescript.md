# Language: JavaScript and TypeScript

When to read: `internal/quality` sent you here because the project's detected languages include `javascript` or `typescript`. The project's own declarations (CI steps, `package.json` `scripts`, `AGENTS.md`) win over every default below.

## Which package manager

The lockfile decides; never mix managers in one project:

| Lockfile | Install | Run a script |
| --- | --- | --- |
| `package-lock.json` | `npm ci` | `npm run <script>` |
| `pnpm-lock.yaml` | `pnpm install --frozen-lockfile` | `pnpm run <script>` |
| `yarn.lock` | `yarn install --frozen-lockfile` (Yarn 1) or `yarn install --immutable` (Yarn 2+) | `yarn <script>` |
| `bun.lock` or `bun.lockb` | `bun install --frozen-lockfile` | `bun run <script>` |

## Evidence and defaults

Prefer the scripts in `package.json` (`lint`, `test`, `typecheck`, `build`, `format:check`). Without a script:

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format | `npx --no-install prettier --check .` | a Prettier config and the `prettier` dependency exist |
| lint | `npx --no-install eslint .` | an ESLint config and the `eslint` dependency exist |
| type check | `npx --no-install tsc --noEmit` | `tsconfig.json` and the `typescript` dependency exist |
| test | the `test` script | always the script; test runners differ too much for a default |
| build | the `build` script | the script exists |

## Dependency existence

- `npm view <name> name version` prints the package; a missing name ends with `E404`. Scoped names (`@scope/name`) are checked whole.
- Type packages are separate: `@types/<name>` exists only for some packages; check it the same way.
- Add with the project's manager (`npm install <name>`, `pnpm add`, `yarn add`, `bun add`) and show the lockfile entry.

## Traps

- Plain `npx <name>` downloads and runs a package that is not installed; always pass `--no-install` in a check.
- Some test runners start in watch mode on a terminal (`vitest`, `jest --watch`): run `vitest run` or the CI script so the run ends.
- `.only(` and `.skip(` in a test file change what runs; a new one is a finding.
- ESM and CommonJS differ (`"type": "module"`, `.mjs`, `.cjs`); keep the module form the file already uses.
- `npm install` without the lockfile's manager rewrites the lockfile; a VERIFY run installs with the frozen command above.
