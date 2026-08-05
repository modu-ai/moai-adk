# Progress — SPEC-PROJECT-NAVIGATOR-003

> Lifecycle skeleton. §E.1 is populated at plan-phase; §E.2-§E.4 are placeholder headings the era-classification engine greps for. Per `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map, the literal `§E.2` / `§E.3` / `§E.4` heading tokens are parser-load-bearing — do NOT rename.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-05
plan_artifact_set: Tier L (5 artifacts + progress.md)
plan_tier: L
plan_req_count: 20
plan_ac_count: 20
plan_boundary_respect:
  spec_001_status: completed
  spec_002_status: completed
  navigator_surface_modified: false
  audit_surface_modified: false
plan_era: V3R6
```

Plan-phase artifacts committed on `feat/SPEC-PROJECT-NAVIGATOR-003` (NOT pushed — orchestrator handles push + plan-PR after plan-auditor review per CLAUDE.local.md §23 PR-mandatory policy).

## §E.2 Run-phase Evidence

### AC binary PASS/FAIL matrix (acceptance.md §D — 20 AC, all MUST-PASS)

| AC | Status | Verification |
|----|--------|--------------|
| AC-NT-001 codemaps emits enriched file when capability-map exists | PASS | `TestNavigatorEnrich_EmitsFilesWhenCapabilityMapExists` — writes capability-symbols.{md,json} with non-empty rows |
| AC-NT-002 001 absence is graceful | PASS | `TestNavigatorEnrich_AbsenceIsGraceful` — exit nil, no output files written |
| AC-NT-003 non-modification of 001 + 002 surfaces | PASS | boundary grep: 0 matches for navigator-regen/navigator-audit/audit-report/progress-map in internal/navigator/astx/ + navigator-enrich.sh; PRESERVE targets untouched |
| AC-NT-004 14 grammars extract symbols on polyglot fixture | PASS | `TestPolyglot_AllFourteenGrammarsExtract` — 14/14 languages Supported=true with expected function+type |
| AC-NT-005 r/flutter fail-open with Supported: false | PASS | `TestPolyglot_RAndFlutterFailOpen` + `TestExtract_ScaffoldedR_ReturnsUnsupported` |
| AC-NT-006 adding a query file extends the language set (no Go edit) | PASS | design §2.1: registration is a data row in seededGrammars + a queries/<lang>.scm file (no per-language Go logic) |
| AC-NT-007 per-language .scm defines symbols | PASS | grep: all 14 working queries carry ≥1 symbol.function/method + ≥1 symbol.type capture |
| AC-NT-008 malformed source tolerance | PASS | extractImpl fail-open on read/parse/query errors (slog.Debug, return Supported:false); never aborts |
| AC-NT-009 provenance per row | PASS | `TestCurrentProvenance_NonEmpty` + CurrentProvenance uses git rev-parse HEAD + committer date (no wall-clock) |
| AC-NT-010 dual output with stable schema | PASS | `TestMarshalCapabilitySymbolsJSON_StableSchema` + `TestRenderMarkdown_ContainsHeader` — 4 top-level + 10 row fields, valid JSON |
| AC-NT-011 header-driven join to 001 capability-map | PASS | `TestEnrichRows_HeaderDrivenJoin` — permuted columns still pair rows correctly |
| AC-NT-012 idempotence | PASS | `TestEnrichRows_Idempotent` — two runs produce byte-identical JSON |
| AC-NT-013 atomic writes | PASS | `TestNavigatorEnrich_AtomicWriteBarrier` — NAVIGATOR_PRE_RENAME_BARRIER blocks rename; no partial file observed |
| AC-NT-014 path ceiling truncates with marker | PASS | `TestEnrichRows_FileCountCeilingTruncation` — MaxFilesPerPath=1 < 2 files → truncated=true |
| AC-NT-015 cgo disabled: builds, nocgo returns Supported:false | PASS | `CGO_ENABLED=0 go build ./...` exit 0; measure_nocgo.go stub returns Supported:false, Extract never panics |
| AC-NT-016 template neutrality grep clean | PASS | 0 matches for SPEC-(V3R[2-6]\|AGENCY\|WORKTREE\|PROJECT-NAVIGATOR) / REQ-(ATR\|WO\|...) across 3 mirrored surfaces; CI guards PASS |
| AC-NT-017 Template-First mirror in place | PASS | 3 files mirrored under internal/template/templates/.claude/skills/; make build regenerated embedded assets |
| AC-NT-018 extractor touches no LSEL surface | PASS | grep: 0 matches for lessons-inbox/state/lsel/memory/feedback/hns-lsel in new surfaces |
| AC-NT-019 extractor touches no 002 audit surface | PASS | grep: 0 matches for audit-report/audit-known-matches in internal/navigator/astx/ + navigator-enrich.sh |
| AC-NT-020 codemaps pipeline classification preserved | PASS | Phase 3 extension is executor delegation (script invocation), NOT LLM dispatch; agentless_audit_test.go CI guard PASS |

### Run-phase commit ledger (M1→M6)

| Milestone | Commit SHA | Subject |
|-----------|------------|---------|
| M1 | 2a835f76a | feat: M1 astx package + grammar registration + cgo split |
| M2 | 3605dea80 | feat: M2 per-language .scm queries + symbol extraction |
| M3 | 0dc5bf172 | feat: M3 row enrichment + header-driven join + output schema |
| M4 | bb6757931 | feat: M4 navigator-enrich.sh + codemaps Phase 3 integration |
| M5 | 678ab90e3 | feat: M5 template-First mirror + neutrality verification |
| M6 | (this commit) | test: M6 full suite + MX tags + AC verification |

### TDD RED evidence (E8)

- M1 RED: `TestExtract_GoFixture` failed `Supported = false, want true` before GREEN implementation.
- M2 RED: 10/14 polyglot languages passed; 4 failed (kotlin query-compile `invalid field 'name'`, elixir `invalid field 'function'`, swift `invalid node type 'struct_declaration'`, cpp function capture miss) → iterated queries against parse-tree dumps → 14/14 GREEN.
- M3 RED: 4 enrich tests failed (0 rows / missing path / ceiling / empty provenance) before GREEN.
- M4 RED: navigator-enrich tests compiled against stub → 0 output → GREEN after CLI implementation.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-05
run_commit_sha: pending-backfill-m6   # M6 commit; backfilled in a follow-up (self-referential SHA)
run_status: complete
ac_pass_count: 20
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE target (001/002/LSEL) modified
l44_pre_commit_fetch: n/a         # not pushed (orchestrator handles push + PR)
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0   # golangci-lint run --timeout=5m ./... = 0 issues (baseline was 0)
cross_platform_build:
  cgo_enabled_1: PASS   # CGO_ENABLED=1 go build ./... exit 0; go test ./internal/navigator/astx coverage 87.8%
  cgo_enabled_0: PASS   # CGO_ENABLED=0 go build ./... exit 0; nocgo stub returns Supported:false
total_run_phase_files:
  new_go_package: internal/navigator/astx/ (astx.go, measure_cgo.go, measure_nocgo.go, enrich.go + 4 test files + queries/*.scm + testdata/)
  new_cli: internal/cli/navigator_enrich.go + navigator_enrich_test.go
  new_skill_surfaces: navigator-enrich.sh, references/navigator-astx.md, codemaps.md Phase 3 extension
  template_mirrors: 3 files under internal/template/templates/.claude/skills/
m1_to_mN_commit_strategy: one-commit-per-milestone (6 commits), Conventional Commits + 🗿 MoAI trailer
coverage_astx_pct: 87.8   # go test -cover ./internal/navigator/astx/... (>= 85% target)
mx_tags_added:
  anchors: 2   # Extract, EnrichRows (each with @MX:REASON + @MX:SPEC)
  notes: 1     # SupportedLanguages (@MX:SPEC)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Decision: sub-agent

```yaml
# Input parameters
tier: L
scope_files_est: ~25-30   # astx Go pkg + 16 .scm queries + navigator-enrich.sh + codemaps.md edit + references + template mirrors
domain_count: 4            # Go package + shell script + skill markdown + template mirror
file_language_mix: Go + shell + markdown + scheme(.scm)
concurrency_benefit: LOW   # coding-heavy per Anthropic coding-task parallelism caveat
agent_teams_prereqs: n/a   # Mode 3 retired
implementation_kickoff_approval: passed   # user approved run-phase entry 2026-08-05
plan_audit_verdict: PASS 0.94             # iter-1, skip-eligible for Phase 1 re-execution (score >= 0.85 Tier L thresh, hash unchanged)
```

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | Tier L multi-file implementation, not a typo |
| 2 background | no | write-capable implementation, not read-only analysis |
| 3 agent-team | no | RETIRED (static team layer retired) |
| 4 parallel | no | coding-heavy → Anthropic caveat: coding tasks have fewer truly parallelizable units than research |
| 5 sub-agent | **YES** | coding-heavy Tier L; sequential manager-develop (cycle_type=tdd) per milestone M1→M6 |
| 6 workflow | no | the 16 `.scm` queries are per-language semantically distinct (tree-sitter node types differ per grammar) — not one uniform mechanical transform; milestones M1→M6 are dependency-ordered |

Decision: sub-agent

Justification: 003 is coding-heavy Go implementation (new `internal/navigator/astx/` package, cgo/nocgo build-tag split, 16 per-language query files, integration script, template mirror). Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time"), the sequential sub-agent path (Mode 5) is the safe default for coding work. The 16 `.scm` query files (M2) might superficially suggest Mode 6 fan-out, but each query is language-specific and semantically distinct — not a single uniform transform rule — and the milestone chain is strictly dependency-ordered (M1 API freezes → M2 queries consume the API → M3 enrichment → M4 integration → M5 mirror → M6 tests). Mode 5 it is.
