# Plan — SPEC-PROJECT-NAVIGATOR-003

> Implementation plan, milestones, technical approach, risks. Milestones are ordered by **decision-reversibility** (highest-reversibility first, mechanical/refactor last) per `.claude/agents/moai/manager-spec.md` Step 4 plan.md ordering guidance. Origin: plan-phase (Tier L).

## §A. Context

### §A.1 Work location + branch

- **Worktree root**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/SPEC-PROJECT-NAVIGATOR-003` (L1 worktree of moai-adk-go)
- **Branch**: `feat/SPEC-PROJECT-NAVIGATOR-003`
- **Base HEAD**: `8df71b18d` (= `origin/main`, current and clean at plan-phase start)
- **Route**: Route B (Tier L, PR-mandatory per CLAUDE.local.md §23 — `enforce_admins:true`)

### §A.2 SPEC artifact paths

- `spec.md` — this plan's authorizing SPEC (full Tier L, 20 REQ, 20 AC)
- `acceptance.md` — Given-When-Then scenarios (20 AC, MUST-PASS)
- `design.md` — system design (tree-sitter integration architecture)
- `research.md` — codebase + web grounding (grammar availability matrix, parsing model)
- `progress.md` — §E lifecycle skeleton (this plan-phase commits §E.1 only)

### §A.3 Plan-auditor verdict

Plan-auditor has NOT yet reviewed this SPEC at authoring time. The plan-phase commit carries `plan_status: audit-ready` (progress.md §E.1) and the orchestrator opens the plan-PR; plan-auditor review happens on the PR. Implementation Kickoff Approval (the plan→run HUMAN GATE) fires after plan-auditor PASS.

### §A.4 PRESERVE targets (do NOT modify)

- `.moai/specs/SPEC-PROJECT-NAVIGATOR-001/**` — 001 is `status: completed`; its contract is frozen (REQ-NT-003).
- `.moai/specs/SPEC-PROJECT-NAVIGATOR-002/**` — 002 is `status: completed`; non-overlap codified in REQ-NT-019.
- `.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` — 001's regeneration script (frozen).
- `.claude/skills/moai-workflow-project/scripts/navigator-audit.sh` — 002's audit script (frozen).
- `.moai/project/navigator/{navigator,capability-map,progress-map}.md` — 001's runtime outputs (REQ-NT-003).
- `.moai/project/navigator/audit-report.{md,json}` — 002's runtime outputs (REQ-NT-019).
- `internal/hook/mx/complexity/**` — the prior-art tree-sitter user. 003's `internal/navigator/astx/` is a SIBLING package; the two share no import edge.
- All LSEL surfaces (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*` skills) — REQ-NT-018.

### §A.5 EXTEND targets (these ARE modified)

- `.claude/skills/moai/workflows/codemaps.md` — extended Phase 3 with the AST-enrichment capability-gated step (REQ-NT-001 + REQ-NT-002).
- `internal/navigator/astx/` — new Go package (sibling to `internal/hook/mx/complexity/`).
- `internal/navigator/astx/queries/*.scm` — 14 working + 2 scaffolded tree-sitter query files.
- `.claude/skills/moai-workflow-project/scripts/navigator-enrich.sh` — new bash script (sibling to `navigator-regen.sh` and `navigator-audit.sh`).
- `.claude/skills/moai-workflow-project/references/navigator-astx.md` — new Level-3 reference (per the progressive-disclosure pattern).
- The template-mirrored copies of all the above under `internal/template/templates/`.

## §B. Known Issues (auto-injected per manager-develop-prompt-template §B)

- **B1 Cross-platform Build Tags**: 003's cgo/nocgo split MUST compile on both `CGO_ENABLED=1` and `CGO_ENABLED=0`. Verified by AC-NT-015. The macOS/Linux CI jobs run with cgo enabled; a Windows CI job (if any) MUST also build — the nocgo path is the fallback.
- **B2 Cross-SPEC Policy Conflict Pre-Scan**: 001 and 002 are `completed`; 003 MUST NOT modify their surfaces (REQ-NT-003 + REQ-NT-019). Grep for cross-SPEC policy conflicts before each commit: `grep -r "navigator-regen\|navigator-audit\|navigator.md\|capability-map\|progress-map\|audit-report" internal/navigator/astx/ scripts/navigator-enrich.sh` should return ZERO matches (the extractor + script names MUST NOT reference 001/002 outputs beyond the single header-driven read of `capability-map.md`).
- **B3 C-HRA-008 / Subagent Boundary Discipline**: 003 is a library + script + skill extension. NO `AskUserQuestion` calls anywhere. Verified: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/navigator/astx/ scripts/navigator-enrich.sh .claude/skills/moai/workflows/codemaps.md` returns ZERO matches.
- **B4 Frontmatter Canonical Schema**: this SPEC's frontmatter uses canonical `created:` / `updated:` / `tags:` (snake_case aliases rejected).
- **B5 CI 3-tier Awareness**: spec-lint, golangci-lint, Go test (per OS) can each fail independently. The polyglot fixture (AC-NT-004) adds a meaningful Go-test cost; benchmark it locally before pushing.
- **B6 spec-lint Heading Convention**: `## Out of Scope` (h2) alone triggers `MissingExclusions` ERROR. The `### Out of Scope — <topic>` h3 sub-sections in spec.md §E comply.
- **B7 observer.go / capture path resolution**: not directly applicable (no hooks added); 003's `navigator-enrich.sh` uses `${CLAUDE_PROJECT_DIR:-$PWD}` like 001's and 002's scripts.
- **B8 Working Tree Hygiene**: do NOT modify `.moai/state/*`, `.moai/harness/*`, `.moai/cache/*`, `.moai/logs/*` (except `.moai/logs/navigator-astx.log` which 003 explicitly appends to per REQ-NT-008). Do NOT `git add -A`; stage specific paths.
- **B9 Git Commit + Push Performed Directly**: N/A for plan-phase. For run-phase (Tier L), commits land via Route B PR; `manager-develop` pushes to `feat/SPEC-PROJECT-NAVIGATOR-003` per milestone; the orchestrator (or `manager-git`) opens the PR.
- **B10 Untouched Paths PRESERVE (Scope Discipline)**: only the 003 SPEC directory + the EXTEND targets in §A.5. Never touch 001/002/LSEL surfaces.
- **B11 AskUserQuestion Prohibited (Subagent Boundary)**: this agent (manager-spec) returns a blocker report if a genuine user decision is required.
- **B12 Sync-phase CHANGELOG emission discipline (manager-docs only)**: not applicable at plan-phase; recorded for run-phase.

## §C. Technical approach (decided — recorded for traceability)

### §C.1 Why the chosen design

The design (design.md §1–§9) is the result of weighing four alternatives:

| Alternative | Verdict | Why rejected |
|-------------|---------|--------------|
| A. Reuse `internal/hook/mx/complexity` (extend the existing package to also extract symbols) | REJECTED | Couples two different consumers (`@MX` scoring vs codemaps enrichment) into one package. Different query targets (decision nodes vs named definitions). Complicates both packages. |
| B. Shell out to a tree-sitter CLI per file | REJECTED | Subprocess overhead dominates at scale. Adds a runtime binary dependency. The Go binding (`smacker/go-tree-sitter`) is already in `go.mod` and avoids both. |
| C. Add a 4th file under `.moai/project/navigator/` (e.g. `capability-map-enriched.md`) | REJECTED | Violates 001's REQ-PN-001 "exactly three ... and no other top-level Navigator files". |
| D. (CHOSEN) New sibling Go package `internal/navigator/astx/` + new sibling script + new sibling output files under `.moai/project/codemaps/` | CHOSEN | Preserves 001's contract. Reuses existing dependency. Follows prior-art pattern. Output surface matches the Epic boundary statement ("integrated into `/moai codemaps`"). |

### §C.2 Why per-language `.scm` query files (vs a universal query)

A universal query across 14 languages would either miss symbols (tree-sitter node types differ: Python `function_definition` vs Go `function_declaration` vs Rust `function_item`) or produce false positives (capturing too much to cover all languages). The per-language query is the right abstraction; it ships as data (REQ-NT-006), so it is cheap to maintain and extend.

### §C.3 Why dual output (md + json)

Mirrors 002's REQ-NA-005 contract: markdown is for human reading; JSON is for downstream tooling (a future LSEL cross-reference, a CI dashboard, an IDE extension). Both stay byte-identical across idempotent re-runs (REQ-NT-012).

## §D. Constraints (DO NOT VIOLATE)

- DO NOT modify any path listed in §A.4 PRESERVE.
- DO NOT introduce a new top-level Go dependency (REQ-NT-004 — reuse `smacker/go-tree-sitter`).
- DO NOT add an `AskUserQuestion` surface anywhere (B3, B11).
- DO NOT wire 003 as a SessionStart / Stop / PostToolUse hook (Advisory-Check Discipline — extraction is not latency-sensitive; REQ-NT-020 preserves the Agentless pipeline classification).
- DO NOT drop the cgo build-tag split (REQ-NT-015 — nocgo path MUST compile).
- DO NOT emit wall-clock timestamps in the enriched output (REQ-NT-012 idempotence — only `git log`-sourced committer dates).
- DO NOT add a 4th file under `.moai/project/navigator/` (REQ-PN-001 carry-forward).
- DO NOT use `--no-verify` on any commit; use Conventional Commit subjects; add the `🗿 MoAI` trailer.
- DO NOT push during plan-phase (the orchestrator handles push + plan-PR after plan-auditor review).

## §E. Self-Verification deliverables (manager-develop run-phase)

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E. Reported at sync-phase as the §E.3 run-phase audit-ready signal:

- **E1** AC Binary PASS/FAIL matrix (20 AC × PASS).
- **E2** Cross-platform build: `CGO_ENABLED=1 go build ./...` AND `CGO_ENABLED=0 go build ./...` both exit 0.
- **E3** Coverage: `go test -cover ./internal/navigator/astx/...` ≥ 85%.
- **E4** Subagent boundary grep: `grep -rn 'AskUserQuestion' internal/navigator/astx/ scripts/navigator-enrich.sh` → 0 matches.
- **E5** Lint: `golangci-lint run --timeout=2m` reports no NEW issues vs baseline.
- **E6** Branch HEAD + push state: list of run-phase commit SHAs + `git push origin feat/SPEC-PROJECT-NAVIGATOR-003` result.
- **E7** Blocker report (if any) — return as structured `## Missing Inputs` table; never call AskUserQuestion.
- **E8** (TDD only) verbatim RED failing-test output before GREEN.

## §F. Milestones (ordered by decision-reversibility)

> The highest-reversibility decisions ship first (data model, type interfaces, schema), so if plan-auditor or implementation discovers the design needs to pivot, the rework cost is concentrated in M1–M2 and does not cascade into M4–M6. Mechanical and refactoring steps ship last.

### §F.1 M1 — Grammar registration + language detection (data model + API)

**Highest reversibility.** This milestone freezes the package's public API and the language-registration data model. Getting the API wrong cascades into every consumer.

- New Go package `internal/navigator/astx/` with public API:
  - `SupportedLanguages() []string`
  - `Extract(ctx context.Context, sourcePath string, language string) (SymbolSet, error)`
  - `SymbolSet{ Supported bool; Symbols map[string][]Symbol; SourceBytes int64; Error error }`
  - `Symbol{ Name string; Kind string; File string; Line int }`
- The `supportedLanguages` registration table: 14 working + 2 scaffolded (r, flutter) entries (research.md §2).
- Language detection by file extension via the registration table.
- cgo/nocgo build-tag split: `measure_cgo.go` (real path) and `measure_nocgo.go` (stub returning `Supported: false` for every language).
- TDD: RED tests for `SupportedLanguages()` (returns the 16-language list with r/flutter marked scaffolded) AND `Extract()` on a Go fixture (returns expected symbols). GREEN: implement.
- Verification: `CGO_ENABLED=1 go test ./internal/navigator/astx/...` AND `CGO_ENABLED=0 go build ./...` both pass.
- Commit: `feat(SPEC-PROJECT-NAVIGATOR-003): M1 astx package + grammar registration + cgo split`

### §F.2 M2 — Per-language `.scm` query files + symbol extraction

**High reversibility.** The query files define what counts as a "symbol" per language. Getting a query wrong cascades into every row that touches that language.

- 14 query files under `internal/navigator/astx/queries/*.scm`:
  - go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, swift
  - Each captures at minimum `@symbol.function`, `@symbol.type` (or language-appropriate equivalents).
- 2 scaffolded stubs: `queries/r.scm`, `queries/dart.scm` — empty placeholders documenting that the language is known-but-unsupported (smacker has no grammar).
- `//go:embed queries/*.scm` for compile-time embedding (mirrors `internal/hook/mx/complexity`).
- Extract algorithm (design.md §3.1): parse → query → group captures → return `SymbolSet`.
- TDD: one polyglot fixture per language (14 fixtures + 2 scaffolded); RED tests for each language's expected symbol set. GREEN: write the query files iteratively until each test passes.
- Commit: `feat(SPEC-PROJECT-NAVIGATOR-003): M2 per-language .scm queries + symbol extraction`

### §F.3 M3 — Capability-row enrichment glue + output schema

**High reversibility.** The enriched-row schema is a new contract; downstream tooling (and 002 forward-compat) depend on its stability.

- Walk algorithm (design.md §3.2): per-row, resolve implementation-path, walk recursively, file-count ceiling with `truncated: true` marker (REQ-NT-014).
- Aggregate symbols: top-N `primary_files` (default 5), top-N `primary_symbols` (default 10), `symbol_count` (total dedup), `on_disk_verified` (true iff path exists AND ≥1 file parsed), `extract_language` (dominant by file count).
- Output schema (design.md §5): the JSON schema is the stable contract. Fields frozen here.
- Header-driven join (REQ-NT-011): read 001's `capability-map.md` header row, resolve columns by name.
- Provenance stamping: `extract_commit_sha = git rev-parse HEAD`, `captured_at = git log -1 --format=%cI <sha>`.
- TDD: RED tests for (a) header-driven join with permuted columns, (b) file-count ceiling truncation, (c) on-disk-verified false case. GREEN: implement.
- Commit: `feat(SPEC-PROJECT-NAVIGATOR-003): M3 row enrichment + header-driven join + output schema`

### §F.4 M4 — `scripts/navigator-enrich.sh` + `/moai codemaps` integration

**Medium reversibility.** The script wraps M1–M3 in a self-contained bash entry point; the codemaps skill-body extension is one capability-gated step.

- New script `scripts/navigator-enrich.sh` (sibling to `navigator-regen.sh` + `navigator-audit.sh`):
  - Self-contained bash: `git` + `awk` + `sed` + `grep` + the `moai navigator-astx` Go entry point (or equivalent). NO `jq`.
  - Reads 001's `capability-map.md` header-driven.
  - Atomic-write: `capability-symbols.{md,json}.tmp` → `mv` (REQ-NT-013).
  - Idempotent (REQ-NT-012).
  - Fail-open on every error mode (design.md §6).
  - Env: `CLAUDE_PROJECT_DIR`, `NAVIGATOR_PRE_RENAME_BARRIER` (test hook).
- Extended `.claude/skills/moai/workflows/codemaps.md` Phase 3 with the capability gate (REQ-NT-001 + REQ-NT-002):
  - IF `capability-map.md` exists → invoke `scripts/navigator-enrich.sh` → emit enriched files.
  - ELSE → info log + continue.
- TDD: RED test for end-to-end `/moai codemaps` invocation on a fixture project (with + without 001's capability-map). GREEN: wire the script.
- Commit: `feat(SPEC-PROJECT-NAVIGATOR-003): M4 navigator-enrich.sh + codemaps Phase 3 integration`

### §F.5 M5 — Template-First mirror + 16-language neutrality

**Medium reversibility.** Mirroring to template source + verifying §25 neutrality is mostly mechanical but MUST happen before the feature ships to downstream users.

- Mirror all new/extended files under `internal/template/templates/`:
  - `.claude/skills/moai/workflows/codemaps.md`
  - `.claude/skills/moai-workflow-project/scripts/navigator-enrich.sh`
  - `.claude/skills/moai-workflow-project/references/navigator-astx.md` (new Level-3 reference)
  - `internal/navigator/astx/**` (Go package sources)
- Run `make build` to regenerate embedded files.
- Template-neutrality grep (AC-NT-016): zero matches for internal SPEC IDs / REQ tokens / audit citations / internal dates / commit SHAs (C2/C3/C7 forbidden classes per CLAUDE.local.md §25.1).
- 16-language neutrality grep: no Go-only example in any template surface.
- CI guard: `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml` pass.
- Commit: `feat(SPEC-PROJECT-NAVIGATOR-003): M5 template-First mirror + neutrality verification`

### §F.6 M6 — Tests, MX tags, sync prep

**Lowest reversibility (mechanical).** Tests, `@MX` annotations on the new exported functions, and the sync-phase prep.

- Full `go test ./...` suite green.
- `golangci-lint run` green.
- Add `@MX:NOTE` to the exported `Extract`, `SupportedLanguages` functions and `@MX:ANCHOR` where fan_in ≥ 3 is anticipated.
- Add `@MX:SPEC:SPEC-PROJECT-NAVIGATOR-003` sub-line on each anchor.
- `moai spec lint SPEC-PROJECT-NAVIGATOR-003` green.
- Coverage ≥ 85% on `internal/navigator/astx/`.
- Verify all 20 AC PASS.
- Commit: `test(SPEC-PROJECT-NAVIGATOR-003): M6 full suite + MX tags + AC verification`

## §G. Anti-Patterns (named, prohibited)

- **AP-NT-001 — Extending `internal/hook/mx/complexity`**: REJECTED (§C.1 alternative A). Coupling two consumers into one package.
- **AP-NT-002 — Subprocess-per-file**: REJECTED (§C.1 alternative B). Subprocess overhead dominates at scale; adds runtime binary dependency.
- **AP-NT-003 — 4th file under `.moai/project/navigator/`**: REJECTED (§C.1 alternative C). Violates 001's REQ-PN-001.
- **AP-NT-004 — Wall-clock timestamp in output**: violates REQ-NT-012 idempotence. Use `git log -1 --format=%cI <sha>`.
- **AP-NT-005 — Wiring 003 as a SessionStart hook**: violates Advisory-Check Discipline. Extraction is not latency-sensitive.
- **AP-NT-006 — Dropping the cgo build-tag split**: violates REQ-NT-015. nocgo MUST still build.
- **AP-NT-007 — Modifying 001 or 002 surfaces**: violates REQ-NT-003 + REQ-NT-019. 001 and 002 are `status: completed`.
- **AP-NT-008 — LLM-driven phase selection in codemaps**: violates REQ-NT-020 (Agentless pipeline classification preserved).
- **AP-NT-009 — Hardcoding language-specific Go logic per language**: violates REQ-NT-006 (data-file extension). Every language MUST be data-driven via the registration table + `.scm` query.

## §H. Cross-References

- `spec.md` — authorizing SPEC (full Tier L).
- `acceptance.md` — Given-When-Then scenarios (20 AC).
- `design.md` — system design.
- `research.md` — grammar availability matrix + parsing model.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — the `draft → in-progress` transition is owned by `manager-develop` (M1 first run-phase commit); the `in-progress → implemented → completed` transition rides the sync commit from `manager-docs`.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — the full Section A-E template is REQUIRED for this Tier L delegation.
- `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification — `/moai codemaps` Agentless pipeline classification preserved (REQ-NT-020).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` § When to set era: explicitly — item 3 covers this SPEC's `era: V3R6` choice.
- `internal/hook/mx/complexity/` — prior-art tree-sitter user (sibling-package design, NOT a refactor target).

## §I. Open questions (none blocking)

The following were considered as potential `[NEEDS CLARIFICATION]` markers and resolved at the SPEC's own defaults (per the iter-2 fold pattern from 001/002). They are recorded here for traceability; NO orchestrator AskUserQuestion round is required.

| Question | Default resolved | Override path |
|----------|------------------|---------------|
| Top-N for `primary_symbols` | 10 | `navigator.astx.primary_symbols_n` config key |
| Top-N for `primary_files` | 5 | `navigator.astx.primary_files_n` config key |
| Per-path file-count ceiling | 2000 | `navigator.astx.max_files_per_path` config key |
| Output location | `.moai/project/codemaps/capability-symbols.{md,json}` | (frozen by this SPEC — changing it requires a future amendment SPEC) |
| r/flutter coverage posture | fail-open, `Supported: false` | (frozen — upgrade requires upstream grammar availability) |
| Era frontmatter | `era: V3R6` explicit | (set; item 3 of lifecycle-sync-gate § When to set era: explicitly) |
