# progress.md — SPEC-GLM-FLASH-DEFAULT-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC authored 2026-08-27 by manager-spec (card t289, Tier M, status: draft).
- Artifact set: spec.md (10 REQ, GEARS) + plan.md (6 milestones) + acceptance.md (13 AC after audit D1 fix, Given-When-Then) + this skeleton.
- Ground-truth anchors verified against tree 410da655f (spec.md §5 Findings); SPEC ID regex check PASS; no ID collision in `.moai/specs/` (674 entries scanned, GLM-family SPECs enumerated, no FLASH-DEFAULT).
- revised: 2026-08-27, version 0.2.0 — plan-audit iter-1 defects D1-D7 applied (AC count 12→13 with REQ-005→AC-013 traceability row; boot-smoke recipe pinned to the buildTmuxInjectVars/setGLMEnv env-injection map; substring-matching wording corrected to registration-time guidance in spec.md REQ-006 + plan M3 + acceptance §D.3; REQ-002 twin scope disambiguated; off-schema `related_specs:` frontmatter key removed; REQ-001 ordering wording softened; overlay call-site inventory added to plan §C).
- plan-audit iter-2 (final, Tier M ceiling): **PASS 1.00** — D1-D7 verified resolved; RED-now pinned tree-wide; residual MINOR R1-R4 (R1 routed into the M6 delegation note: setGLMEnv leg needs t.Setenv, not t.TempDir).

## §E.0 Operator gate record (2026-08-27, orchestrator 전달)

1. **Implementation Kickoff APPROVED** (2026-08-27).
2. **Operator mid-dispatch additions are binding scope**: (a) glm-5.3-flash uses reasoning_effort max ONLY (no low/high states — collapse overlay must branch per-model); (b) the moai web console settings surface gains glm-5.3-flash (i18n labels ×4 locales).
3. **Progression: autonomous** — M1→M6 직진, 중간 승인 정지 없음. blocker만 보고.

## §F Phase 4 Mode Selection

- Input: tier M · scope ~12 files (Go 4-5, template yaml 1, i18n 1, tests 4-5) · domains 3 (config/template/statusline) + web i18n · language mix Go+YAML+JS · concurrency benefit LOW (coding-heavy, cross-file coupled constants).
- direct: not selected — multi-file semantic change, not a typo.
- serial: **selected** — coding-heavy coupled work; one manager-develop delegation carrying per-milestone commits (Anthropic coding-task parallelism caveat).
- fanout: not selected — write-coupled single-spec scope; fan-out would race the constants/twin surfaces.
- sweep: not selected — not mechanical-uniform; <30 files; semantic.
- Decision: serial
- Justification: the tier-slot constants, closed set, overlay, and template twin are one coupled change surface; a single serial delegation with per-milestone commits (M1..M6, plan order) keeps the twins coherent and the RED→GREEN chain attributable. Implementation Kickoff Approval passed (gate record above); preferences drained (autonomous progression).

## §E.2 Run-phase Evidence

Run-phase executed 2026-08-27 by manager-develop (cycle_type=tdd), branch `WT-glm-flash-default`, base 3a8f9cc9e → HEAD 9e1bb9e3d (M6). All attribution: this run, this tree, commands issued from the worktree root.

### §E.2.1 Milestone commits (M1→M6)

| M | Commit | Subject |
|---|--------|---------|
| M1 | a5454a505 | feat: register glm-5.3-flash + switch tier-slot defaults (also carries the M4 i18n labels — see note) |
| M2 | 5c67d1869 | feat: flash max-only effort overlay branch (wire + display threading) |
| M3 | 43d4e0031 | feat: explicit glm-5.3-flash context-window table entry |
| M4 | 8e91b7226 | test: flash web-label count pin (8 = 2 key families x 4 locales) |
| M5 | 337577028 | docs: glm-5.3-flash default in README + docs-site (4-locale) |
| M6 | 9e1bb9e3d | test: boot smoke TestGLMFlashDefaultEnvInjection |

Note: the flash i18n labels and the 5-model web set assertion landed inside M1 — the schema-parity guards (`TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`) fail the moment the derived set offers flash, so a green M1 commit required them. M4 then adds the AC-010 count pin. `make build` exit 0; `internal/template/catalog.yaml` byte-unchanged (`git diff --stat internal/template/catalog.yaml` empty after build); template-neutrality grep on the edited template llm.yaml: 0 hits for SPEC-IDs/dates.

### §E.2.2 AC matrix (E1)

| AC | Status | Command | Observed output (verbatim) |
|----|--------|---------|---------------------------|
| AC-001 closed set | PASS | `go test ./internal/config/ -run 'TestDefaultGLMConstants' -v -count=1` | `--- PASS: TestDefaultGLMConstants` + `defaults_test.go:432: ValidGLMModels() = [glm-5.3-flash glm-5.3 glm-5.1 glm-4.7 glm-4.5-air]` |
| AC-002 tier defaults | PASS | `go test ./internal/config/ -run 'TestNewDefaultLLMConfig_GLMTierMapping|TestDefaultGLMConstants' -count=1` | `ok github.com/modu-ai/moai-adk/internal/config` — all 7 constants + all 7 config fields asserted = "glm-5.3-flash"; `DefaultGLM53 = "glm-5.3"` preserved assertion included |
| AC-003 template twin + build | PASS | `grep -n 'glm-5.3-flash' internal/template/templates/.moai/config/sections/llm.yaml`; `make build` | lines 183-186: all four `llm.glm.models.*` = `"glm-5.3-flash"`; `make build` → `go build ... -o bin/moai` exit 0 |
| AC-004 glm-5.3 preserved | PASS | `go test ./internal/template/ -run 'TestCollapseClaudeEffortToGLMForModel|TestSessionGLMReasoningStateForModel' -count=1`; `go test ./internal/settings/ -count=1` | `ok github.com/modu-ai/moai-adk/internal/template` — glm-5.3 rows: low→low, medium→max, unrecognized→max; settings suite `ok` (glm-5.3 pin values still load) |
| AC-005 flash low→max | PASS | `go test ./internal/template/ -run TestCollapseClaudeEffortToGLMForModel -count=1` | `ok` — flash × {low, medium, high, xhigh, max, unrecognized, ""} all resolve `{thinking enabled, reasoning_effort: max}` |
| AC-006 non-flash unchanged | PASS | same command (mirror-image rows) | `ok` — glm-5.3 × low → low; glm-5.3/glm-5.1/"" × low → low (mirror-image regression rows in-test) |
| AC-007 totality | PASS | same command | `ok` — unrecognized effort under flash AND non-flash → max, no panic |
| AC-008 context window | PASS | `go test ./internal/statusline/ -run 'TestResolveGLMContextWindow|TestGLMContextWindowsFlashDirectEntry' -count=1` | `ok` — flash → 1000000 via direct entry; glm-5.3 → 1000000 retained; unregistered `glm-5.3-flash-lite` inherits 1M via substring |
| AC-009 web widget | PASS | `go test ./internal/web/ -run 'TestGLMModelSelectOptions' -v -count=1` | `--- PASS: TestGLMModelSelectOptions` — options = 5-set incl. glm-5.3-flash, Type select, Validate rejects out-of-set |
| AC-010 labels ×4 | PASS | `grep -c 'glm-5.3-flash' internal/web/assets/i18n.js`; `go test ./internal/web/ -run TestGLMFlashOptionLabelsAllLocales -v` | `8`; `--- PASS: TestGLMFlashOptionLabelsAllLocales` (2 key families × en/ko/ja/zh, non-empty) |
| AC-011 docs | PASS | `grep -rn 'glm-5.3-flash' README*.md docs-site/content/ \| wc -l` | `68` mentions across README ×4 + 20 docs-site pages; every page carries ≥1 (4-locale parity grep per file: en/ko/ja/zh each 4-5/4/4/4/2/1/1) |
| AC-012 boot smoke | PASS | `go test ./internal/cli/ -run TestGLMFlashDefaultEnvInjection -v -count=1` | `--- PASS: TestGLMFlashDefaultEnvInjection` — 4× `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` = glm-5.3-flash; `MOAI_STATUSLINE_CONTEXT_SIZE` = 1000000 (map leg); `CLAUDE_CODE_AUTO_COMPACT_WINDOW` = 1000000 (both legs); no live z.ai dependency |
| AC-013 overlay doc twin | PASS | `grep -n 'glm-5.3-flash' internal/template/templates/.moai/config/sections/llm.yaml` | line 209: `# FLASH EXCEPTION (glm-5.3-flash): ... EVERY Claude effort level — including low — resolves to thinking=enabled,` / `# reasoning_effort=max` inside the effort-mapping comment block, alongside the existing collapse table |

### §E.2.3 Full affected-package runs (E1/E2/E5)

- `go test ./internal/config/... ./internal/settings/... ./internal/statusline/... ./internal/web/... -count=1` → `ok` × 8 packages (config, config/atomicfile, config/toolpolicy, settings, settings/agentfm, settings/yamlpatch, statusline, web).
- `go test ./internal/template/ -count=1` → `ok`.
- `go test ./internal/cli/ -run 'TestGLMFlashDefaultEnvInjection|TestGLMReasoningEnvVarsForModel|TestLoadGLMConfig|TestSaveLLMSection|TestBuildTmuxInjectVars|TestGLMMaxContext' -count=1` → `ok`, 28 `--- PASS` rows (full internal/cli suite delegated to CI per lane-local verification discipline).
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `golangci-lint run --timeout=2m` → `0 issues.` (baseline pre-flight was also `0 issues.` — zero NEW findings).

### §E.2.4 Coverage (E3)

- `go test ./internal/statusline/ -cover` → `coverage: 90.5% of statements`.
- `go test ./internal/config/ -cover` → `coverage: 80.6% of statements` (package-level; dominated by pre-existing loader surface untouched by this SPEC — touched functions below verified directly).
- `go tool cover -func` on internal/template: `IsGLMFlashModel / CollapseClaudeEffortToGLMForModel / ResolveGLMReasoningForModel / SessionGLMReasoningStateForModel / CollapseClaudeEffortToGLM / ResolveGLMReasoning / SessionGLMReasoningState(ForEffort)` all `100.0%`.
- `ValidGLMModels` shows 0% in the config-only profile because its callers live in internal/web and internal/settings (cross-package); exercised green there (TestGLMModelSelectOptions, audit_pin_fields_test derive from it).

### §E.2.5 Boundary + hygiene (E4/E6/E7)

- `grep -rn 'AskUserQuestion|mcp__askuser' internal/{config,template,statusline,web} --include='*.go' | grep -v _test.go | grep -v '// '` → `0`.
- No push performed (orchestrator owns push+PR at sync). All staging by explicit pathspec; the runtime-touched `.moai/config/sections/{crosssession,feedback}.yaml` were excluded from every commit (B8/B10).
- E7 blockers: none. One coordinator mid-flight notice received (primary-checkout hotfix alignment): naming aligned (DefaultGLM53Flash, same i18n key prefixes, same 5-model set order); the beyond-scope primary changes (section rename "GLM Settings", app.js wireGLMFlashEffortLock) NOT replicated per instruction — merge-time reconciliation is orchestrator-owned.

### §E.2.6 RED evidence (E8)

Pinned RED-now baseline (tree 3a8f9cc9e, pre-implementation): `grep -rn 'glm-5.3-flash' internal/` → 0 hits tree-wide.

M1 RED (captured before GREEN, same tree):
```
$ go test ./internal/config/ -run 'TestNewDefaultLLMConfig_GLMTierMapping|TestDefaultGLMConstants' -count=1
internal/config/defaults_test.go:406:5: undefined: DefaultGLM53
internal/config/defaults_test.go:409:5: undefined: DefaultGLM53Flash
FAIL github.com/modu-ai/moai-adk/internal/config [build failed]

$ go test ./internal/web/ -run TestGLMModelSelectOptions -count=1
--- FAIL: TestGLMModelSelectOptions (0.00s)
    glm_tier_test.go:60: config.ValidGLMModels() = [glm-5.3 glm-5.1 glm-4.7 glm-4.5-air], want [glm-5.3-flash glm-5.3 glm-5.1 glm-4.7 glm-4.5-air]
    (x4 more rows: high/medium/low/fable options mismatch)
FAIL github.com/modu-ai/moai-adk/internal/web
```

M2 RED (overlay functions not yet implemented):
```
$ go test ./internal/template/ -run 'TestIsGLMFlashModel|ForModel' -count=1
internal/template/glm_effort_overlay_test.go:269:12: undefined: ResolveGLMReasoningForModel
internal/template/glm_effort_overlay_test.go:285:12: undefined: SessionGLMReasoningStateForModel
(too many errors)
FAIL github.com/modu-ai/moai-adk/internal/template [build failed]
```

M3 RED (direct table entry absent — red for the right reason: resolution alone would pass via substring, the divergence-guard assertion is what fails):
```
$ go test ./internal/statusline/ -run 'TestGLMContextWindowsFlashDirectEntry' -count=1
--- FAIL: TestGLMContextWindowsFlashDirectEntry (0.00s)
    memory_test.go:553: glmContextWindows has no explicit "glm-5.3-flash" entry — the flash window rides the glm-5.3 substring match instead of its own entry
FAIL github.com/modu-ai/moai-adk/internal/statusline
```

M6 note: the boot-smoke test is born-green — the behavior it asserts (flash defaults) is the M1 output, whose RED is pinned above; the smoke adds the env-level observation leg, not new production behavior.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-08-27"
run_commit_sha: "9e1bb9e3d"
run_status: "complete"
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: true
l44_post_push_fetch: null   # no push — orchestrator owns push+PR at sync
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: "exit 0"
  windows_amd64: "exit 0"
total_run_phase_files: 33   # 9 Go source/test files + i18n.js + template llm.yaml + 24 doc files
m1_to_mN_commit_strategy: "per-milestone commits M1..M6 + spec frontmatter flip on M1"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-08-27"
sync_commit_sha: "f1208eba4"
sync_status: "audit-ready"
verification_basis:
  - "§E.2: 13/13 AC PASS with verbatim evidence (AC-001..AC-013)"
  - "lint: 0 NEW warnings/lints introduced (§E.3 new_warnings_or_lints_introduced: 0)"
  - "dual-platform build: darwin exit 0, windows_amd64 exit 0"
  - "post-merge merged-tree smoke re-verified by orchestrator (origin/main merge landed clean after run_commit_sha 9e1bb9e3d)"
gaps:
  - "full internal/cli suite deferred to PR CI (lane-local verification scope per §4 discipline)"
  - "internal/config package coverage 80.6% — pre-existing, not introduced by this SPEC"
changelog_entry_position: "CHANGELOG.md [Unreleased] ### Changed"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (merged 3-phase close on the sync commit)"
b12_self_test:
  a_pre_emission_grep: "grep -c 'SPEC-GLM-FLASH-DEFAULT-001' CHANGELOG.md -> 0 (pre-emission)"
  b_ac_count_match: "acceptance.md distinct AC tokens = 13; CHANGELOG entry states 13 PASS, 0 FAIL"
  c_file_path_verification: "README.md, internal/web/assets/i18n.js, internal/config/defaults.go, internal/config/closed_sets.go, internal/template/glm_effort_overlay.go, internal/statusline/memory.go — all read on this tree before emission"
```
