# internal/spec — Module Conventions

## Purpose

`internal/spec` parses SPEC documents under `.moai/specs/SPEC-*/`, validates their YAML frontmatter against the canonical schema, maintains the SPEC catalog (`catalog.yaml`), and runs lint rules (`FrontmatterSchemaRule`, `OwnershipTransitionRule`, ...). The package is consumed by `moai spec lint`, `moai spec list`, and `moai spec validate` CLI subcommands.

This package is the enforcement boundary for SPEC-First DDD methodology. Lint failures here block run-phase progression; catalog hash drift here invalidates downstream regeneration.

## Conventions

- **Frontmatter canonical schema (SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`)**: 12 required fields — id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags. Optional fields — issue_number, depends_on, lint.skip, bc_id, tier. Any change to required-field count MUST update both the rule body in `lint.go` (`FrontmatterSchemaRule`) and the SSOT schema doc + template mirror. Tests must be updated in `lint_test.go`.
- **Status enum (8 values)**: draft, planned, in-progress, implemented, completed, superseded, archived, rejected. Validators MUST reject any other value. Transitions are governed by the Status Transition Ownership Matrix (see SSOT) — manager-spec owns `draft → planned`, manager-develop owns `draft|planned → in-progress|implemented`, manager-docs owns `implemented → completed`.
- **Lint rule pattern**: New rules implement `type Rule interface { Check(spec *Document) []Finding }`. Findings carry `Severity` (`error`/`warning`/`info`), `Code` (stable string like `OwnershipTransitionInvalid`), `Message`, and `Line` (1-indexed source line). Rules are observation-only — NEVER mutate the spec or call `os.Exec`, `ioutil.WriteFile`, or any file-modify primitive.
- **Heading convention (spec-lint `MissingExclusions`)**: SPEC body §B (or equivalent "Out of Scope" section) MUST use H3 (`###`) or H4 (`####`) sub-headings with "Out of Scope —" infix (e.g., `### Out of Scope — Performance benchmarks`). List-item-only sections fail the rule. Past offenders: SPEC-V3R6-CI-BASELINE-DRIFT-001 H3 retrofit, LEGACY-CLEANUP-002 list-to-heading promotion.
- **Catalog hash discipline**: When a SPEC body's `§A.3` evidence section changes, the SHA256 hash in `catalog.yaml` is invalidated. Regenerate via `gen-catalog-hashes.go --all` as a same-SPEC cascade (per L46 attribution discipline). Never edit `catalog.yaml` hash fields by hand.
- **Sibling SPEC PRESERVE**: When implementing a lint extension, the new rule MUST NOT false-flag closed sibling SPECs whose `status: implemented` was set BEFORE the rule shipped. Tests must include a "no false positive on closed SPECs" subtest using ARR-001 or COORD-001 as fixture.

## Key Patterns

- **Table-driven test convention**: Lint rule tests use `[]struct{ name, content string; wantFindings int }` plus a `for _, tt := range` loop with `t.Run(tt.name, ...)`. Each subtest is independently runnable via `go test -run TestXxx/subname ./internal/spec/`. Existing pattern: `TestFrontmatterSchemaRule_RequiredFields`.
- **`git log --follow` for transition detection**: When verifying `status:` transitions against git history (e.g., `OwnershipTransitionRule` checking that `manager-develop` authored the `in-progress` transition), use `git log --follow --format='%an|%ae|%s' <file>` and parse author + commit-subject pattern. Cache results per-file within a single `Check()` call — `git log` is the slow path. UnreachableGit branch (sandbox without git, e.g., `t.TempDir()` test fixture) MUST gracefully no-op without raising error severity.
- **Frontmatter parsing**: Use `gopkg.in/yaml.v3` (already vendored). Frontmatter delimited by `^---$` at line 1 and the next `^---$`. Body parsing uses `goldmark` AST for heading extraction.
- **SPEC ID regex (canonical)**: `^SPEC-[A-Z][A-Z0-9]+(-[A-Z0-9]+)*-\d{3}$` — enforced by `manager-spec` body pre-write self-check (SPEC-V3R6-SPEC-ID-VALIDATION-001 origin). Update tests in `lint_test.go` if the regex tightens.

## Cross-References

- Root CLAUDE.md §5 (SPEC-Based Workflow), §6 (Quality Gates — LSP)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — SSOT for frontmatter schema + Status Transition Ownership Matrix. **Future**: `OwnershipTransitionRule` cross-reference subsection added by M3.
- `.claude/agents/core/manager-spec.md` — `draft → planned` ownership
- `.claude/agents/core/manager-develop.md` — `draft|planned → in-progress|implemented` ownership
- `.claude/agents/core/manager-docs.md` — `implemented → completed` ownership
- `internal/spec/catalog.go` + `gen-catalog-hashes.go` — catalog hash regeneration tool
