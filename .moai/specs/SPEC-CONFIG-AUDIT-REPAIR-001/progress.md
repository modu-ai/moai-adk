# Progress — SPEC-CONFIG-AUDIT-REPAIR-001

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored 2026-07-25; amended to 0.2.0 same day after Implementation Kickoff: all 3 design decisions RESOLVED (D1 ast-grep gate RESTORE default-OFF — user-directed reversal of the removal recommendation; D2 tool-policy dev-only; D3 mcp-matrix reword) and plan-auditor SHOULD-FIX items folded (touch-point inventory, AC-CAR-002/005/007(c) tightening, LOW counting convention pinned in HISTORY). No [NEEDS CLARIFICATION] markers remain. Run-ready.

## §E.2 Run-phase Evidence

### M-2 — ast-grep gate restore (REQ-CAR-011/019/020/021)

| Item | Evidence |
|------|----------|
| gate.yaml loader (REQ-CAR-011) | `internal/config/loader_gate.go` + `Loader.Load` registration; `go test -run TestLoadGateSection ./internal/config/` → PASS (enable / default-off / partial-override) |
| Explicit default OFF (REQ-CAR-019) | `defaults.go NewDefaultGateConfig` sets `AstGrepGate.Enabled: false` explicitly; `TestLoadGateSection_DefaultOff` + `TestPreToolHandler_LoadGateConfig/astgrep_gate_default_off_via_config` PASS |
| Guard reachability (AC-CAR-011) | `TestQualityGate_Run_AstGrepGuardReachable` (zig marker + pure-Go suppression violation → Run returns false via gate.go astgrep branch) + `TestPreToolHandler_LoadGateConfig/astgrep_gate_enabled_via_config` (pre_tool mapping) — both PASS |
| V1 deletion (REQ-CAR-021) | `RunAstGrepGate` + `runSGConfig`/`runSGRule`/`parseSGScanOutput`/`astGrepScanMatch` deleted from `astgrep_gate.go`; V1 tests + `TestRunAstGrepGate_V1_V2_Equivalence` removed; `go build ./...` green |
| sgconfig `utils` phantom (REQ-CAR-020a) | `grep -n utils sgconfig.yml` → 0; ruleDirs now `[go, security]` (loadable curated set; demo-stub language dirs excluded pending ASTGREP-DOGFOOD-CLEANUP-001) |
| Config-mode loadability (REQ-CAR-020b) | go/ + security/ rules converted to sg-native `rule:` nested format; 2 sg-unparseable multi-node demo rules dropped (go-http-response-body-not-closed, go-mutex-not-deferred); manual run `sg scan --config .moai/config/astgrep-rules/sgconfig.yml --json internal/config/loader_gate.go` → exit 0, output `[]` |
| Graceful degrade without `sg` (REQ-CAR-020c) | `TestRunAstGrepGateV2_NoSgCLI` (PATH="") PASS — enabled gate skips scan, no hard failure |
| gate.yaml template+local | template-first: `internal/template/templates/.moai/config/sections/gate.yaml` + local mirror; `make build` exit 0; `TestStructYAMLSymmetry_Gate` + `TestAuditLoaderCompleteness` PASS |

### M-3 — H4+M3 distribution codification (REQ-CAR-013/014)

| Item | Evidence |
|------|----------|
| tool-policy graceful CLI (REQ-CAR-013) | `internal/cli/tool_policy.go`: absent-file guard prints one-line dev-only notice, exit 0 for both `list` and `build`; verified in clean `/tmp/tpclean` — `moai tool-policy list` → notice + exit=0, no stack |
| tool-policy dev-only declaration | CLAUDE.local.md §2 Local-Only Files entry added (`.moai/config/sections/tool-policy.yaml`) |
| mcp-matrix reword (REQ-CAR-014) | Template refs removed: `grep -rn 'mcp-matrix' internal/template/templates/` → 0 matches (project.md table row + doc-generation.md Step 3.6.2 generalized to matrix-optional wording, both trees); CLAUDE.local.md §2 entry added |
| Build + guards | `make build` exit 0; `go test ./internal/cli/ ./internal/template/...` green |

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
