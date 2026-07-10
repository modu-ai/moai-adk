# SPEC-CLIFIX-CONCURRENCY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md) by manager-spec from CLI audit 2026-07-10 §3 clusters 2/5 + §5 P2. Status: draft. Pending plan-audit. Sequenced after SPEC-CLIFIX-CONTRACT-001.
- 2026-07-11: iter-1 plan-audit FAIL (aggregate 0.77 < 0.80 Tier M threshold). Re-grounded against post-CRITICAL-001/CONTRACT-001 HEAD — all anchors re-verified by grep. 7 fixes applied (D1-D7). REQ/AC reduced 7/7→6/6: dropped REQ-001-005 (Upsert tier-transition already fixed at filestore.go:96 + read-side cascade/dedupe covers it). Pending iter-2 audit.
- 2026-07-11: D1 micro-fix (v0.2.1) — iter-2 plan-audit PASS 0.89 with single D1 debt (atomic-writer inventory under-count). Corrected inventory 4→6 sites (added atomicWrite preference/filestore.go:473 + glm.go:689-704 inline). Option (b) for atomicWrite (package-internal exception). AC-004 grep widened to include `func atomicWrite`. No other REQ/AC touched.

## §E.2 Run-phase Evidence

### M1 — Re-route writers + removeGLMEnv key-set (REQ-CONC-001-001 + REQ-CONC-001-002)

**AC PASS/FAIL matrix (M1 scope = AC-001 + AC-002):**

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CONC-001-001 | PASS | `go test -race ./internal/cli/ -run 'SettingsLocalConcurrent' -count=1 -v` | `PASS — TestSettingsLocalConcurrentWrites (0.12s)`; grep: 0 `writeSettingsMap` callers outside settings.go (all 5 re-routed) |
| AC-CONC-001-002 | PASS | `go test ./internal/cli/ -run 'RemoveGLMEnvComplete' -count=1 -v` && `grep -n 'EnvClaudeCodeAutoCompactWindow' internal/cli/launcher.go` | `PASS — TestRemoveGLMEnvComplete`; `launcher.go:292: delete(env, config.EnvClaudeCodeAutoCompactWindow)` |

**RED evidence (SPEC §D.5 — tests must fail on pre-fix commit):**
- `TestSettingsLocalConcurrentWrites` FAILED pre-M1: `parse settings.local.json: invalid character 'P' after top-level value` — concurrent `writeSettingsMap` (plain os.WriteFile) truncated the file mid-read by another goroutine.
- `TestRemoveGLMEnvComplete` FAILED pre-M1: `CLAUDE_CODE_AUTO_COMPACT_WINDOW was NOT removed by removeGLMEnv` — key missing from the delete set at launcher.go:268-281.

**Changes applied (GREEN):**
1. `internal/cli/glm.go`: 3 callers re-routed through `mutateSettingsLocal` — `ensureSettingsLocalJSON`, `injectGLMEnvForTeam`, `injectGLMEnv`.
2. `internal/cli/launcher.go`: 2 callers re-routed through `mutateSettingsLocal` — `removeGLMEnv`, `syncPermissionModeToSettingsLocal`; `removeGLMEnv` delete-set gains `config.EnvClaudeCodeAutoCompactWindow`.
3. `internal/cli/settings.go`: deleted dead `writeSettingsMap` (golangci-lint `unused` after all 5 callers re-routed).
4. `internal/cli/clifix_concurrency_repro_test.go`: new RED→GREEN reproduction tests.

**Cross-platform build:** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
**Lint:** `golangci-lint run --timeout=2m` → 0 issues (pre-flight baseline was also 0; the transient `writeSettingsMap unused` finding was resolved by deleting the dead function).
**Coverage:** `go test -cover ./internal/cli/` → 72.7% of statements (package-level).
**Race suite:** `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` → all 10 packages PASS.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 0.95 Mode Selection

**Input parameters:**
- tier: M
- scope (file count): ~12-15 (settings.go, glm.go, launcher.go, glm_tools.go, update_noise.go, harness_mute.go, preference/{filestore,decay,toggle}.go + new race/repro tests)
- domain count: 1-2 (internal/cli concurrency + atomic-write; single Go module)
- file language mix: 100% Go (implementation + tests)
- concurrency benefit: LOW (coding-heavy, sequential dependency between milestones M1→M4)
- Agent Teams prereqs: NOT MET — workflow.team.enabled=false (Sonnet 5/Opus 4.8 default-off); CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS unset; harness level not `thorough`

**Mode evaluation:**
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | non-trivial concurrency hardening across 6+ files |
| 2 background | no | write tasks (implementation); background writes raise main-session prompts |
| 3 agent-team | no | team.enabled=false — capability gate fails |
| 4 parallel | no | coding-heavy Go work; sequential milestone dependency (M1 seam reuse → M2 → M3 → M4); Anthropic coding-task parallelism caveat |
| 5 sub-agent | **yes** | coding-heavy, sequential per-milestone delegation; default fallback |
| 6 workflow | no | <30 files, semantic concurrency edits (not mechanical uniform transform) |

**Decision:** sub-agent (Mode 5, sequential manager-develop per milestone)

**Justification:** This is coding-heavy Go concurrency work with tight sequential dependency (M1 re-routes writers through the existing CRITICAL-001 seam; M2/M3/M4 build on that). Per Anthropic's coding-task parallelism caveat, sequential sub-agent delegation is the safe default. Agent Teams (Mode 3) is unavailable (team.enabled=false), and the scope is neither ≥30-file mechanical (Mode 6) nor research-parallel (Mode 4). Mode 5 with cycle_type=tdd (quality.yaml `development_mode: tdd`, AG-01 backward-compat) drives RED reproduction tests (race-detector failures) → GREEN fixes per milestone. Implementation Kickoff Approval obtained (plan→run HUMAN GATE, score-independent per REQ-ATR-015).
