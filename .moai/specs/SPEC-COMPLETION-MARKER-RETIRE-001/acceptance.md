# Acceptance Criteria — SPEC-COMPLETION-MARKER-RETIRE-001

Each AC has a verifiable command (run from the project root `/Users/goos/MoAI/moai-adk-go`). Commands assume the bash working directory is the project root. "Zero matches" means the command prints nothing and exits cleanly (use `; echo "exit=$?"` to confirm exit status where a grep returning 1 is the success condition).

---

## D. AC Matrix

### AC-CMR-001 — Stop handler marker code removed + non-marker behavior preserved (REQ-CMR-001, REQ-CMR-004, REQ-CMR-005)

Given the run-phase removal is complete, the Stop handler source contains no marker symbols, and the surviving non-marker behavior (REQ-CMR-004) is preserved.

```bash
grep -nE 'defaultCompletionMarkers|hasCompletionMarker|completionMarkers|NewStopHandlerWithMarkers' internal/hook/stop.go
# Expected: zero matches
```

```bash
# Given-When-Then: GIVEN a session-end event WHEN the Stop handler runs THEN it still compiles + returns allow
go build ./internal/hook/... && go test ./internal/hook/ -run TestStopHandler_Handle -count=1
# Expected: build OK + tests PASS (surviving non-marker Stop behavior preserved)
```

### AC-CMR-002 — Persistent-mode subsystem retired (REQ-CMR-002)

```bash
test ! -f internal/hook/lifecycle/persistent_mode.go && echo "RETIRED"
# Expected: prints RETIRED
```

```bash
grep -rnE 'ActivatePersistentMode|DeactivatePersistentMode|CheckPersistentMode|PersistentMode\b' internal/ pkg/ cmd/ --include='*.go' | grep -v '_test.go'
# Expected: zero matches (no production references remain)
```

### AC-CMR-003 — PreCompact persistent-mode read removed (REQ-CMR-003)

```bash
grep -nE 'readPersistentMode|persistent-mode|Execution Mode' internal/hook/compact.go
# Expected: zero matches
```

```bash
# Edge: the worktrees section in compact.go is preserved (not collateral-removed)
grep -n 'readWorktrees' internal/hook/compact.go
# Expected: at least one match (readWorktrees intact)
```

### AC-CMR-004 — Zero production references to markers or persistent-mode (REQ-CMR-001..003, primary gate)

```bash
grep -rnE 'moai>DONE|moai>COMPLETE|CompletionConfig|MarkersConfig|CompletionMarker|completionMarker|PersistentMode|persistent-mode|persistentMode|LogCompletionMarkers' internal/ pkg/ cmd/ --include='*.go' | grep -v '_test.go'
# Expected: zero matches (exhaustive production, non-test sweep across all Go source)
```

### AC-CMR-005 — Config struct + YAML + models field removed, config package green (REQ-CMR-006, REQ-CMR-007, REQ-CMR-008)

```bash
grep -nE 'CompletionConfig|MarkersConfig|Completion ' internal/config/types.go internal/config/defaults.go
# Expected: zero matches
grep -n 'LogCompletionMarkers' pkg/models/config.go
# Expected: zero matches
grep -nE '^[[:space:]]*completion:' .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml
# Expected: zero matches (completion: block removed from both local + template YAML)
```

```bash
# Struct↔YAML symmetry + loader-completeness guards stay green after the coupled removal
go test ./internal/config/... -count=1
# Expected: PASS (CONFIG_STRUCT_YAML_MISMATCH / YAML_SECTION_NO_LOADER not triggered)
```

### AC-CMR-006 — Test suite green after test retirement (REQ-CMR-009)

```bash
test ! -f internal/hook/stop_completion_test.go && echo "stop_completion_test RETIRED"
test ! -f internal/hook/lifecycle/persistent_mode_test.go && echo "persistent_mode_test RETIRED"
# Expected: both RETIRED lines printed
```

```bash
grep -rnE 'PersistentMode|CompletionMarker|moai>DONE|moai>COMPLETE|NewStopHandlerWithMarkers' internal/hook/ internal/config/ --include='*_test.go'
# Expected: zero matches (no test still references retired symbols)
```

```bash
# Full suite green gate (the authoritative run-phase pass condition)
go test ./... -count=1
# Expected: ok across all packages (no cascading failures)
```

### AC-CMR-007 — docs-site 4-locale + README parity, zero residual markers (REQ-CMR-013)

```bash
# Widened scope (auditor D4): docs-site/ entire tree (content + docs-site/README.md), excluding transient agent worktrees
grep -rnE 'moai>DONE|moai>COMPLETE|<moai>' docs-site/ | grep -v '/worktrees/'
# Expected: zero matches (all 4 content locales + docs-site/README.md updated; no lagging locale)
```

```bash
# Per-locale residual count must be identically zero — content parity check
for loc in en ko ja zh; do
  c=$(grep -rlE 'moai>DONE|moai>COMPLETE|<moai>' "docs-site/content/$loc/" 2>/dev/null | grep -v '/worktrees/' | wc -l | tr -d ' ')
  echo "$loc=$c"
done
# Expected: en=0 ko=0 ja=0 zh=0
```

```bash
# docs-site/README.md (auditor D4 addition — the 7 EN landing-doc hits at 517/563/745/778/782/783/810)
grep -nE 'moai>DONE|moai>COMPLETE|<moai>' docs-site/README.md
# Expected: zero matches
# Note: root README.md / README.ko.md have NO <moai> hits (verified by D4 grep) — not in scope.
```

### AC-CMR-008 — Complete consumer removal (both trees) + catalog regeneration (REQ-CMR-010, REQ-CMR-012)

```bash
# Complete consumer set zero-residual across BOTH the live tree AND the template mirror
# (auditor D1/D6 — widened to the full 8 consumers incl. .claude/agents/, all skills/workflows/rules).
# Exclude transient agent worktrees AND the plan-auditor memory artifact (not a consumer).
grep -rnE 'moai>DONE|moai>COMPLETE|<moai>' .claude/ internal/template/templates/.claude/ \
  | grep -v '/worktrees/' | grep -v '/agent-memory/'
# Expected: zero matches (live tree + template mirror both clean)
```

```bash
# Template-mirror subset explicitly verified (the 5 template-managed consumers).
# release.md + release-update-specialist.md are dev-only (no template mirror) — correctly absent here.
grep -rnE '<moai>' \
  internal/template/templates/.claude/output-styles/moai/moai.md \
  internal/template/templates/.claude/output-styles/moai/einstein.md \
  internal/template/templates/.claude/skills/moai/SKILL.md \
  internal/template/templates/.claude/skills/moai/workflows/loop.md \
  internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md 2>/dev/null
# Expected: zero matches
```

```bash
# NOTE: internal/template/embedded.go does NOT exist — the project uses //go:embed all:templates
# (internal/template/embed.go:28). The correct "templates regenerated" gate is the catalog-hash
# regeneration that `make build` runs (gen-catalog-hashes.go --all over internal/template/catalog.yaml).
make build && git diff --exit-code -- internal/template/catalog.yaml
# Expected: make succeeds (templ-generate + catalog-hash regen + go build all exit 0) AND
#           catalog.yaml has no post-build diff (hashes already regenerated in-commit).
```

```bash
# SSOT-mirrored rule byte-parity guard (spec-workflow.md is mirror-tracked).
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror' -count=1
# Expected: PASS (live .claude/rules/.../spec-workflow.md == template mirror, byte-identical)
```

### AC-CMR-009 — moai.md XML-exception clauses removed (REQ-CMR-011)

```bash
# D5 correction: the actual wording is an EXCEPTION carve-out, not "sole permitted XML".
# Both exception clauses (line 48 "— except completion markers ..." and line 729
# "(except <moai> markers)") must be gone. Match the REAL strings.
grep -niE 'except completion markers|except .moai. markers|except <moai>|<moai>DONE|<moai>COMPLETE' .claude/output-styles/moai/moai.md
# Expected: zero matches (no marker carve-out remains in the "No XML in user-facing output" rule)
```

```bash
# Positive assertion: the "No XML tags in user-facing output" rule still EXISTS (we removed the
# exception, not the rule itself).
grep -niE 'No XML tags in user-facing output|never raw XML' .claude/output-styles/moai/moai.md
# Expected: ≥1 match (the rule survives; only its marker exception was stripped)
```

### AC-CMR-010 — SPEC traceability recorded (REQ-CMR-014)

```bash
grep -lE 'SPEC-COMPLETION-MARKER-RETIRE-001|partially_superseded_by|superseded.*completion.marker' .moai/specs/SPEC-PERSIST-001/spec.md .moai/specs/SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001/spec.md
# Expected: both files reference this SPEC's partial supersession of the marker/persistent-mode surface
```

### AC-CMR-011 — `/moai loop` retains an explicit exit signal (REQ-CMR-015, auditor D8)

```bash
# The marker-based loop-exit check is gone from BOTH the live loop.md and its template mirror.
grep -nE 'moai>DONE|moai>COMPLETE|<moai>' .claude/skills/moai/workflows/loop.md internal/template/templates/.claude/skills/moai/workflows/loop.md
# Expected: zero matches (Step 1 Completion Check + :100 + :157 no longer reference the marker)
```

```bash
# A replacement success-exit signal EXISTS (Option a: natural-language completion sentence).
# loop.md Step 1 / Completion Conditions must still describe an explicit success-exit path.
grep -niE 'completion (condition|sentence|signal)|exit(ing)? (the )?loop|loop.*complete|all .*conditions satisfied' .claude/skills/moai/workflows/loop.md
# Expected: ≥1 match — the loop still has a documented explicit success-exit path (not only the
#           4 implicit conditions). If Option (b) was chosen at GATE-2, the replacement token
#           satisfies this instead.
```

```bash
# The 4 non-marker exit conditions are PRESERVED unchanged (zero-errors+tests+coverage /
# max-iterations / memory-pressure / user-interruption).
grep -niE 'max.?iteration|memory.?pressure|user.?interrupt|coverage.*threshold|zero errors' .claude/skills/moai/workflows/loop.md
# Expected: ≥3 matches (the backstop exit conditions remain intact)
```

---

## D.1 — Severity Classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-CMR-004 | MUST-FIX | Primary zero-production-reference gate — defines "retired" |
| AC-CMR-006 | MUST-FIX | `go test ./...` green is the authoritative pass condition |
| AC-CMR-005 | MUST-FIX | Config guards must not regress on coupled struct/YAML removal |
| AC-CMR-001/002/003 | MUST-FIX | Per-site removal completeness (Go runtime) |
| AC-CMR-007 | MUST-FIX | 4-locale + README parity is a HARD docs obligation |
| AC-CMR-008 | MUST-FIX | Complete consumer removal (both trees) + catalog-hash regen + mirror-drift guard |
| AC-CMR-011 | MUST-FIX | `/moai loop` must retain an explicit exit signal — losing it silently degrades loop control (auditor D8) |
| AC-CMR-009 | SHOULD-FIX | XML-exception removal; design-dependent on GATE-2 Option (a) |
| AC-CMR-010 | SHOULD-FIX | Traceability hygiene |

## D.2 — Edge Cases

- **Edge-CMR-1 (worktrees preservation):** `compact.go` `readWorktrees` + P2 "Active Worktrees" section AND `compact_coverage_test.go` `TestReadWorktrees_*` MUST remain after `readPersistentMode` removal (AC-CMR-003).
- **Edge-CMR-2 (assertion-count comment — auditor D3):** `defaults_test.go` AC-WSE-007 "36-assertion oracle" comment MUST be reduced by **3** (DetectInOutput + Markers.Done + Markers.Complete), not 2; suite stays green (AC-CMR-006).
- **Edge-CMR-3 (transient worktree + memory exclusion):** `.claude/worktrees/agent-*/` copies AND the `agent-memory/plan-auditor/` memory artifact are excluded from all consumer grep gates (`grep -v '/worktrees/'`, `grep -v '/agent-memory/'`) — runtime/memory artifacts, not source of truth.
- **Edge-CMR-4 (loop-exit replacement, not absence — auditor D8):** under Option (a), AC-CMR-011 asserts the marker is ABSENT but an explicit natural-language exit signal is PRESENT. The marker is a LIVE `/moai loop` exit consumer, so pure absence (with no replacement) would be a regression, not a valid end-state.
- **Edge-CMR-5 (dev-only no-mirror):** `release.md` + `release-update-specialist.md` are dev-only (no template mirror) — AC-CMR-008 template-subset grep correctly omits them; creating template copies would be a §25 isolation violation.
- **Edge-CMR-6 (no embedded.go):** `internal/template/embedded.go` does not exist (the project uses `//go:embed all:templates`). AC-CMR-008 asserts catalog.yaml hash regen + build success instead — NOT a vacuous diff on a nonexistent file.

## D.3 — Definition of Done

- All MUST-FIX ACs pass.
- `go test ./... -count=1` green.
- `make build` succeeds and produces no `catalog.yaml` diff (NOT `embedded.go` — that file does not exist; `//go:embed all:templates` + catalog-hash regen is the mechanism).
- All 4 docs-site locales + `docs-site/README.md` free of marker references (parity).
- `/moai loop` retains an explicit success-exit signal (AC-CMR-011) — the marker is replaced, not merely removed.
- `TestRuleTemplateMirrorDrift` passes (SSOT-mirrored `spec-workflow.md` byte-parity preserved).
- SPEC frontmatter `status` transitioned per the standard lifecycle by the owning agents (manager-spec sets initial `draft`; downstream transitions owned by manager-develop / manager-docs).
- GATE-2 design decision (Option a vs b) confirmed before run-phase entry.
