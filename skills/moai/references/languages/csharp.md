# Language: C#

When to read: `internal/quality` sent you here because the project's detected languages include `csharp`. The project's own declarations (CI steps, `AGENTS.md`) win over every default below.

## Evidence and defaults

Run each command on the solution (`*.sln`, `*.slnx`) when there is one, otherwise on the project file (`*.csproj`).

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format | `dotnet format --verify-no-changes` | a solution or project file exists (it reads `.editorconfig`) |
| lint | the analyzers run inside the build | `<EnableNETAnalyzers>` or analyzer packages are set; `-warnaserror` only when CI uses it |
| type check | covered by `dotnet build` | |
| test | `dotnet test` | a test project exists (a reference to `Microsoft.NET.Test.Sdk`) |
| build | `dotnet build` | a solution or project file exists |

The SDK version comes from `global.json` when present; another SDK is `blocked`, not a pass.

## Dependency existence

- `https://api.nuget.org/v3-flatcontainer/<id in lower case>/index.json` answers 404 for a package that does not exist; recent SDKs also have `dotnet package search <id> --exact-match`.
- Add with `dotnet add <project> package <id>`; show the new `PackageReference` and, when the project locks, the `packages.lock.json` entry.

## Traps

- `dotnet format` without `--verify-no-changes` rewrites files; a VERIFY run never rewrites.
- `dotnet test --filter` that matches nothing reports no tests: that is `not_run`.
- `[Fact(Skip = …)]`, `[Ignore]` and `[Explicit]` hide tests; a new one is a finding.
- `--no-restore` and `--no-build` reuse earlier output; use them only right after a restore or build of the same tree.
