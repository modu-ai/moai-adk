# Reference: Navigator AST Enrichment (`navigator-astx`)

> Level-3 progressive-disclosure reference for the AST symbol enrichment that
> `/moai codemaps` runs in Phase 3 (the capability-gated extension step).
> Load this reference when extending the supported language set, debugging a
> missing-symbol extraction, or integrating the enriched output downstream.

## What it is

The `navigator-enrich` step reads 001's `capability-map.md`, walks each row's
implementation-path, and extracts named symbols (functions, methods, types)
via tree-sitter grammars. It writes two sibling files under
`.moai/project/codemaps/`:

- `capability-symbols.md` — human-readable table (spec-id, title, path, on-disk
  marker, file count, symbol count, primary symbols).
- `capability-symbols.json` — machine-readable envelope with the full
  `primary_files` + `primary_symbols` arrays per row.

## Supported languages (16)

14 working grammars (extract symbols) + 2 scaffolded (fail-open, Supported: false):

| Working (14)                              | Scaffolded (2) |
|-------------------------------------------|----------------|
| go, python, typescript, javascript, rust, | r              |
| java, kotlin, csharp, ruby, php, elixir,  | flutter (dart) |
| cpp, scala, swift                         |                |

The two scaffolded languages have no upstream grammar in the
`smacker/go-tree-sitter` dependency; `Extract` returns `Supported: false`
without attempting a parse.

## Extension model (adding a language)

Adding a 17th language (or upgrading r/flutter from scaffolded) is a
data-driven change — no per-language Go logic:

1. Add a `queries/<lang>.scm` file capturing `@symbol.function` and
   `@symbol.type` (or language-appropriate kinds).
2. Add one row to the `seededGrammars` map in `measure_cgo.go` binding the
   grammar to the embedded query bytes.
3. For a newly-supported language, add the extension to the registration
  `langMeta` table in `astx.go` and remove it from `scaffoldedLanguages`.

The `smacker/go-tree-sitter` sub-package for the grammar must exist; if it
does not, the language stays scaffolded until upstream availability.

## Output schema (JSON)

```json
{
  "extracted_at": "<ISO-8601 committer date of extract_commit>",
  "extract_commit": "<git rev-parse HEAD>",
  "source_capability_map": ".moai/project/navigator/capability-map.md",
  "rows": [
    {
      "spec_id": "...",
      "title": "...",
      "implementation_path": "...",
      "on_disk_verified": true,
      "extract_language": "go",
      "primary_files": ["..."],
      "primary_symbols": [{"name":"...","kind":"type","file":"...","line":1}],
      "symbol_count": 47,
      "truncated": false,
      "supported": true
    }
  ]
}
```

Schema is forward-compatible: additive field additions only; no field removals
or semantic redefinitions.

## Invariants

- **Fail-open**: every error mode (absent capability-map, missing path, parse
  failure, cgo-disabled build) degrades to a log line + continue — the step
  never aborts `/moai codemaps`.
- **Idempotent**: provenance is sourced from `git rev-parse HEAD` + the
  committer date, never wall-clock. Two runs on the same HEAD produce
  byte-identical output.
- **Atomic writes**: each output file is written `<file>.tmp` then renamed.
  The `NAVIGATOR_PRE_RENAME_BARRIER` env var is a test hook for the
  atomic-rename fixture.
- **Boundary integrity**: writes ONLY `.moai/project/codemaps/` and
  `.moai/logs/navigator-astx.log`. It never touches 001/002 outputs or any
  LSEL surface.

## Build constraints

The package mirrors `internal/hook/mx/complexity`'s cgo/nocgo split:

- `CGO_ENABLED=1` (default on macOS/Linux): full tree-sitter implementation
  (`measure_cgo.go`).
- `CGO_ENABLED=0` (e.g. cross-compile, minimal containers): stub returns
  `Supported: false` for every language (`measure_nocgo.go`); the build still
  compiles and `Extract` never panics.
