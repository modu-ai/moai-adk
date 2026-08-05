# Research — SPEC-PROJECT-NAVIGATOR-003

> Tree-sitter-specific grounding: grammar availability across the 16 languages, the parsing/symbol model, symbol-extraction patterns, CI matrix implications, and the binary-dependency / parsing-compatibility surface. Builds on 001's `research.md` (Aider repo map, Codebase-Memory, AGENTS.md, MemGPT/Mem0, Claude compaction) — that material is incorporated by reference, not duplicated here.

## §1. Web grounding (external research)

### Aider — repo map via tree-sitter + PageRank (carried from 001, specialized to 003)

- **Source**: https://aider.chat/2023/10/22/repomap.html , https://aider.chat/docs/repomap.html
- **Claim**: Aider builds a repository map by parsing the codebase with tree-sitter, constructing a graph of symbol definitions and references, and running PageRank to rank the most important symbols. Only the highest-ranked symbols are delivered to the LLM as context for the current edit.
- **Implication for 003**: confirms that tree-sitter-based symbol extraction is the right MECHANICAL primitive for "what symbols implement this capability". The PageRank symbol-ranking step is explicitly OUT OF SCOPE for 003 (spec.md §E Out of Scope — PageRank), but the def/reference graph construction pattern informs the per-language query design (§3 below): a symbol-extraction query MUST capture definitions, not references, to avoid drowning the row in call-site noise.
- **Cost calibration**: Aider's documented cost is per-edit (small subset of symbols ranked, LLM-fed). 003's cost is per-capability-row (walk one `implementation-path`, extract ALL definitions in that path, emit one row). The two cost profiles are different; 003's is bounded by REQ-NT-014's file-count ceiling.

### Codebase-Memory (arXiv 2603.27277) — tree-sitter + SQLite persistence (carried from 001, specialized)

- **Source**: https://arxiv.org/html/2603.27277v1
- **Claim**: tree-sitter + MCP knowledge graph persisted in a single SQLite file; survives sessions; reported ~10× token reduction and ~2.1× fewer tool calls vs file exploration.
- **Implication for 003**: confirms tree-sitter as the right parsing primitive. 003 deliberately does NOT adopt the SQLite persistence layer — markdown rollups (REQ-NT-010 dual output) are human-readable, diff-friendly, and add no binary dependency. SQLite-per-project remains a possible future evolution if markdown rollups become a bottleneck at monorepo scale (carried forward from 001's decision).

### tree-sitter — official documentation

- **Source**: https://tree-sitter.github.io/tree-sitter/ , https://tree-sitter.github.io/tree-sitter/using-parsers
- **Claim**: tree-sitter is an incremental error-recovering parser generator. Grammar definitions are written in JavaScript; generated parsers are C source; bindings exist for many host languages. Parsing is fault-tolerant — a syntax error in one part of a file does not prevent parsing the rest.
- **Implication for 003**:
  - **Error recovery** (REQ-NT-008 malformed-source tolerance): tree-sitter's error-recovering property means a syntactically broken source file still produces a partial parse tree. 003's extractor MUST still tolerate a hard parse failure (encoding issue, binary file, file removed mid-walk), but in practice most "broken" files will yield a partial symbol set rather than a hard failure — the extractor SHOULD use the partial set, not discard it.
  - **Query language** (`.scm`): the tree-sitter query language is a small S-expression DSL that matches node types and captures named bindings. It is the right abstraction for REQ-NT-007's per-language symbol definition: each language gets a query file that captures `(function_definition name: ...)` or `(method_declaration name: ...)` etc.
  - **Incremental parsing** is NOT used by 003 — each extraction pass re-parses the file from scratch. Incremental parsing would matter if 003 maintained a long-lived parse cache; it does not (idempotence REQ-NT-012 means re-extraction produces byte-identical output, so caching is an optimization a future SPEC may add, not a current requirement).

### smacker/go-tree-sitter — Go binding

- **Source**: https://github.com/smacker/go-tree-sitter , `github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82` in this repo's `go.mod`
- **Claim**: cgo-based Go binding for tree-sitter. Ships grammar packages per language as sub-packages (`/golang`, `/python`, `/typescript/typescript`, etc.). Used in production by Aider (Python host) and by other Go tooling.
- **Implication for 003**:
  - **cgo required** (REQ-NT-015 cgo/nocgo split): the library uses cgo. The existing `internal/hook/mx/complexity/measure_{cgo,nocgo}.go` pattern is the canonical way this repo handles that, and 003's `internal/navigator/astx/` follows the same pattern.
  - **Grammar availability** (research §2 below): 14 of the 16 MoAI-supported languages have grammar sub-packages; `r` and `dart`/`flutter` do not, and 003 fails-open for those (REQ-NT-005).
  - **Version pinning**: the library is at a pseudo-version (`v0.0.0-20240827...`) — no semver release. 003 inherits the version already in `go.mod` (no new top-level dependency, REQ-NT-004).

### Codebase-knowledge-graphs for AI coding agents (carried from 001, specialized)

- **Source**: https://www.developersdigest.tech/blog/codebase-knowledge-graphs-ai-coding-agents
- **Claim**: a good map REDUCES context rather than inflating it; every claim needs provenance (commit hash + timestamp).
- **Implication for 003**: 003's enriched rows are rollup references, not content copies. The `primary-symbols` column carries a small top-N symbol list (configurable, default ≤10) per row, NOT every symbol in the path. This keeps the enriched `capability-symbols.md` smaller than the source tree it summarizes.

## §2. Grammar availability matrix — 16 supported languages vs smacker/go-tree-sitter

The MoAI-ADK 16-language support set (CLAUDE.local.md §15): `go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift`.

Verification method: `ls ~/go/pkg/mod/github.com/smacker/go-tree-sitter@<version>/` to enumerate grammar sub-packages present in the pinned version.

| # | Language | smacker sub-package | Status for 003 |
|---|----------|---------------------|----------------|
| 1 | go | `/golang` | WORKING — primary pilot language (already used by `internal/hook/mx/complexity`) |
| 2 | python | `/python` | WORKING — already used by complexity |
| 3 | typescript | `/typescript/typescript` | WORKING — already used by complexity |
| 4 | javascript | `/javascript` | WORKING — already used by complexity |
| 5 | rust | `/rust` | WORKING — already used by complexity |
| 6 | java | `/java` | WORKING — new in 003 |
| 7 | kotlin | `/kotlin` | WORKING — new in 003 |
| 8 | csharp | `/csharp` | WORKING — new in 003 |
| 9 | ruby | `/ruby` | WORKING — new in 003 |
| 10 | php | `/php` | WORKING — new in 003 |
| 11 | elixir | `/elixir` | WORKING — new in 003 |
| 12 | cpp | `/cpp` | WORKING — new in 003 |
| 13 | scala | `/scala` | WORKING — new in 003 |
| 14 | swift | `/swift` | WORKING — new in 003 |
| 15 | r | (none) | SCAFFOLDED — `Supported: false` (REQ-NT-005) |
| 16 | flutter (dart) | (none) | SCAFFOLDED — `Supported: false` (REQ-NT-005) |

**Implication**: 14/16 language coverage on day 1; 2/16 fail-open. This is a documented, data-driven coverage claim, not an aspirational one — REQ-NT-004 and REQ-NT-005 are calibrated to this matrix.

**Upgrade path** (research-driven): when `smacker/go-tree-sitter` (or a successor) adds `r` and/or `dart` grammar sub-packages, adoption is a data-file change (new `.scm` query file + a registration-table entry), NOT a Go source edit (REQ-NT-006).

## §3. Symbol-extraction patterns — per-language query design

### §3.1 The "symbol" definition (cross-language)

For 003's purpose (capability-row enrichment), a "symbol" is a **named definition** that contributes to the capability's implementation. The minimal cross-language set:

| Symbol kind | Examples across languages |
|-------------|---------------------------|
| function / method | Go `func`, Python `def`, Rust `fn`, Java method, TS `function` |
| type / class / interface | Go `type`, Python `class`, Rust `struct`/`enum`/`trait`, Java `class`/`interface`, TS `interface`/`type` |
| (optionally) exported constant / macro | Go `const`, Rust `const`/`static`, C/C++ `#define` |

003 captures the first two categories by default; the third is per-language opt-in via the `.scm` query.

### §3.2 Query file shape (cross-language)

Each `.scm` query file (e.g. `queries/go.scm`) captures the symbol kinds named above. Example for Go (carried from `internal/hook/mx/complexity/queries/go.scm`, extended for 003's symbol set):

```scheme
; 003 symbol captures for Go
(function_declaration name: (identifier) @symbol.function)
(method_declaration name: (field_identifier) @symbol.method)
(type_declaration (type_spec name: (type_identifier) @symbol.type))
```

Each language's query file is authored during M2 (plan.md §F.2). The 14 query files are SMALL (estimated 5–20 lines each); the 2 scaffolded stubs (`r.scm`, `dart.scm`) are empty placeholders that the registration table marks `Supported: false`.

### §3.3 Prior art in this repo

`internal/hook/mx/complexity/queries/{go,python,typescript,javascript,rust}.scm` — the existing 5 query files for the complexity package. They capture decision nodes (if/for/switch) for cyclomatic complexity, NOT symbol definitions, but the file structure + embed pattern is identical and 003's query files follow the same shape. 003 ships 14 query files under `internal/navigator/astx/queries/*.scm` (sibling directory, no overlap with the complexity package).

## §4. CI matrix implications

### §4.1 cgo enablement

`smacker/go-tree-sitter` requires cgo. The project's CI matrix MUST run with `CGO_ENABLED=1` on at least one job to exercise the real tree-sitter path. The existing CI already handles this (the complexity package's tests require cgo); 003's tests join the same job.

The nocgo path (`CGO_ENABLED=0 go build ./...`) MUST still compile (REQ-NT-015). This is verified by the existing `internal/hook/mx/complexity/measure_nocgo.go` pattern; 003 mirrors it.

### §4.2 Polyglot fixture cost

003's test suite requires a polyglot fixture: one source file per supported language, each declaring a known symbol set the assertions can verify against. The fixture is ~16 small files (one per language; the 2 scaffolded languages' fixtures assert the `Supported: false` path). This is smaller than the complexity package's fixture suite because 003 does not measure complexity, only presence/absence of definitions.

### §4.3 Grammar build-time cost

The 14 grammars compile into the binary at build time (each grammar is a Go sub-package). Compilation cost is paid once per `make build`; runtime cost is per-file parse. Neither is a CI bottleneck — `internal/hook/mx/complexity` already pays this cost for 5 grammars and the build time is dominated by other factors.

## §5. Binary-dependency / parsing-compatibility surface

### §5.1 What 003 does NOT add

- **No new top-level Go dependency** — `github.com/smacker/go-tree-sitter` is already in `go.mod` (REQ-NT-004).
- **No new binary dependency** for end users — `/moai codemaps` is a skill + bash script; the extractor is a Go entry point inside the existing `moai` binary (or a sibling binary compiled from the same module), not a separate install.
- **No network calls** at extraction time — grammars are compiled in, source files are local, `git log` is local.
- **No subprocess-per-file** — the extractor is a single process that walks the path and parses each file in-tree. (Rejected alternative: shelling out to a tree-sitter CLI per file — subprocess overhead dominates at scale; research §6 anti-pattern AP-NT-002 in plan.md.)

### §5.2 What 003 inherits

- **cgo toolchain requirement** — inherited from `internal/hook/mx/complexity`. Downstream users who build from source without cgo get the nocgo stub path (REQ-NT-015).
- **smacker/go-tree-sitter's parsing compatibility** — the library's own per-grammar compatibility with the language's syntax. 003 does NOT control upstream grammar quality; a grammar that mis-parses a language construct is an upstream issue. 003's fail-open contract (REQ-NT-008) handles this: a mis-parsed file is skipped with a warning, not aborted.

### §5.3 Forward-compat surface

- **Grammar library evolution** — if `smacker/go-tree-sitter` publishes a breaking API change, 003's `internal/navigator/astx/` package is the single integration point. The package's public API (`Extract(ctx, filePath, language)`, `SupportedLanguages()`) is the stable contract; the library's API is encapsulated behind it.
- **Language addition** — REQ-NT-006's data-file extension path: a new `.scm` + a registration-table entry. No Go source edit required.

## §6. Prior-work review (this repo, tree-sitter-specific)

### `internal/hook/mx/complexity/` — the canonical prior art

- **Path**: `internal/hook/mx/complexity/{complexity.go,complexity_test.go,measure_cgo.go,measure_nocgo.go,queries/*.scm}`
- **Status**: shipped, used by `@MX:WARN` cyclomatic-complexity detection.
- **Owns**: cyclomatic complexity measurement per function, via tree-sitter decision-node queries.
- **Boundary vs 003**:
  - Different consumer: complexity feeds `@MX` scoring; astx feeds `/moai codemaps` enrichment.
  - Different query target: complexity captures decision nodes (if/for/switch); astx captures named definitions (function/type/class).
  - Same library + same cgo pattern + same embed pattern: the two packages share infrastructure but not code (no import edge).
- **Implication**: 003's `internal/navigator/astx/` package is a SIBLING to `complexity`, not a refactor of it. Extending `complexity` to also extract symbols would couple two different consumers (MX scoring vs codemaps enrichment) into one package and complicate both; the sibling-package design preserves separation of concerns.

### `.claude/skills/moai/workflows/codemaps.md` — the integration target

- **Path**: `.claude/skills/moai/workflows/codemaps.md` (171 lines)
- **Status**: shipped, Agentless fixed-pipeline.
- **Owns**: codebase-wide architecture documentation under `.moai/project/codemaps/`.
- **Phase 3 (map generation)** is the integration point: 003 adds an AST-enrichment step inside Phase 3, emitting `capability-symbols.{md,json}` as sibling output files. The pipeline classification (Agentless) is preserved (REQ-NT-020).
- **codemaps-extract.js fan-out**: the bundled high-count fan-out script under `.claude/workflows/`. 003 does NOT depend on it (003 is deterministic Go, not an LLM fan-out), but the two are complementary: `codemaps-extract.js` layers architecture INSIGHT on top of the deterministic baseline, while 003 layers MECHANICAL symbol extraction. A future SPEC may teach `codemaps-extract.js` to consult 003's enriched rows, but that coupling is not in scope here.

### `.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` and `navigator-audit.sh`

- **Status**: shipped (001 + 002).
- **Owns**: 001's regeneration script and 002's audit script.
- **Boundary vs 003**: 003 adds a SIBLING script `scripts/navigator-enrich.sh`, sibling to `navigator-regen.sh` and `navigator-audit.sh`. The three scripts form a triptych (regen / audit / enrich), each self-contained bash with the same portability + atomic-write + idempotence contract. 003 does NOT modify 001's or 002's scripts.

## §7. Summary — what the research changes in this SPEC

| Research finding | Where it shaped the SPEC |
|------------------|--------------------------|
| tree-sitter is the right parsing primitive (Aider, Codebase-Memory) | REQ-NT-004 reuses `smacker/go-tree-sitter` already in `go.mod` |
| Error recovery means partial parses are usable, not discarded | REQ-NT-008 malformed-source tolerance |
| Query language (.scm) is the right per-language symbol abstraction | REQ-NT-006 + REQ-NT-007 (data-file extension; per-language symbol definition) |
| Map reduces context, not inflates it | `primary-symbols` column is a top-N (default ≤10), NOT every symbol |
| Every claim needs provenance | REQ-NT-009 `extract-commit-sha` + ISO-8601 `captured-at` per row |
| Grammar coverage: 14/16 on day 1, 2/16 fail-open | REQ-NT-004 + REQ-NT-005 calibrated to the §2 matrix |
| cgo required, nocgo must still build | REQ-NT-015 mirrors the `complexity` package's build-tag split |
| Prior art in `internal/hook/mx/complexity` | 003's `internal/navigator/astx/` is a SIBLING package, not a refactor |
