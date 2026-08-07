# Acceptance — SPEC-RALPH-CONFIG-REDESIGN-001

## §D. AC Matrix

| AC ID | REQ | Severity | Summary |
|-------|-----|----------|---------|
| AC-RCR-001 | REQ-RCR-003 | MUST | 23 inert leaf keys removed from local + template ralph.yaml |
| AC-RCR-002 | REQ-RCR-002 | MUST | RalphConfig struct UNCHANGED — all 5 fields preserved (prior field-removal defect corrected) |
| AC-RCR-003 | REQ-RCR-004 | MUST | 3 missing live keys added (max_iterations, auto_converge, human_review) with defaults matching NewDefaultRalphConfig |
| AC-RCR-004 | REQ-RCR-005 | MUST | cfg.Ralph wired into NewRalphEngine (Option A) OR ralph.yaml documents advisory-only status (Option B-doc fallback) |
| AC-RCR-005 | REQ-RCR-006 | MUST | LintAsInstruction + WarnAsInstruction read paths at post_tool.go unchanged |
| AC-RCR-006 | REQ-RCR-007 | MUST | Session.StaleSeconds dead pipeline removed (4 sites: SessionConfig field, ralphFileWrapper field+tag, loader injection block, defaults entries); go build green |
| AC-RCR-007 | REQ-RCR-008 | MUST | go build ./... + scoped go test exit 0 |
| AC-RCR-008 | REQ-RCR-009 | MUST | /moai loop runs; lint-as-instruction injection still governed by user edit |
| AC-RCR-009 | REQ-RCR-010 | MUST | template + local ralph.yaml match on the 5-key surface |

## §D.1 AC Definitions (Given-When-Then)

### AC-RCR-001 — 23 inert leaf keys removed
**Given** the post-redesign `ralph.yaml` (local + template),
**When** `grep -En '^(  )?(enabled|lsp|ast_grep|loop|hooks):' .moai/config/sections/ralph.yaml internal/template/templates/.moai/config/sections/ralph.yaml` runs,
**Then** zero matches for `lsp:`, `ast_grep:`, `loop:`, `hooks:` blocks AND zero matches for a top-level `enabled:` under `ralph:`;
**And when** `grep -cE 'auto_start|timeout_seconds|poll_interval_ms|graceful_degradation|config_path|security_scan|quality_scan|require_confirmation|cooldown_seconds|zero_errors|zero_warnings|tests_pass|coverage_threshold|post_tool_lsp|stop_loop_controller|trigger_on|severity_threshold|check_completion' .moai/config/sections/ralph.yaml internal/template/templates/.moai/config/sections/ralph.yaml` runs,
**Then** the count is 0 (all 23 inert leaf keys gone).

### AC-RCR-002 — RalphConfig struct UNCHANGED (prior defect corrected)
**Given** the post-redesign source tree,
**When** `sed -n '330,343p' internal/config/types.go` (or reading the `RalphConfig` struct) runs,
**Then** the struct exposes exactly 5 fields — `MaxIterations`, `AutoConverge`, `HumanReview`, `LintAsInstruction`, `WarnAsInstruction` — unchanged from the pre-SPEC state;
**And when** `grep -rn '\.AutoConverge\|\.HumanReview\|\.MaxIterations\|\.LintAsInstruction\|\.WarnAsInstruction' internal/ralph/engine.go internal/cli/deps.go internal/hook/post_tool.go` runs,
**Then** each field has at least one live read site (engine.go:62 / engine.go:74 / deps.go:135 / post_tool.go:426 / post_tool.go:439 — line numbers may drift; the reads survive);
**And when** `go build ./internal/config/...` runs,
**Then** exit 0 (no field removal broke compilation).

### AC-RCR-003 — 3 missing live keys added with correct defaults
**Given** the post-redesign `ralph.yaml` (local + template),
**When** the file is read,
**Then** it contains `max_iterations: 5`, `auto_converge: true`, `human_review: true` under the `ralph:` namespace (top-level, NOT nested under `loop:`);
**And when** `diff <(grep -E '^(  )?(max_iterations|auto_converge|human_review|lint_as_instruction|warn_as_instruction):' .moai/config/sections/ralph.yaml) <(grep -E '^(  )?(max_iterations|auto_converge|human_review|lint_as_instruction|warn_as_instruction):' internal/template/templates/.moai/config/sections/ralph.yaml)` runs,
**Then** exit 0 (local and template agree on the 5-key surface).

### AC-RCR-004 — engine wired (Option A) OR documented advisory (Option B-doc fallback)
**Given** the post-redesign source tree,
**When** `grep -n 'NewDefaultRalphConfig\|cfg\.Ralph\|NewRalphEngine' internal/cli/deps.go` runs,
**Then** EITHER (Option A) `NewRalphEngine` is called with the loaded `cfg.Ralph` (not `NewDefaultRalphConfig()`) — the `NewDefaultRalphConfig()` call at the engine-construction site is gone or reduced to a fallback,
**Or** (Option B-doc fallback) `deps.go` is unchanged AND the `ralph.yaml` header comment documents that `max_iterations`/`auto_converge`/`human_review` are advisory-only;
**And when** `go build ./...` runs,
**Then** exit 0.

### AC-RCR-005 — LintAsInstruction + WarnAsInstruction preserved
**Given** the post-redesign source tree,
**When** `grep -n 'LintAsInstruction\|WarnAsInstruction' internal/config/types.go internal/hook/post_tool.go` runs,
**Then** the fields exist in `RalphConfig` and are read at `post_tool.go` (the two `return cfg.Ralph.<Field>` lines survive — line numbers may drift);
**And when** `.moai/config/sections/ralph.yaml` is read,
**Then** both `lint_as_instruction` and `warn_as_instruction` are present under `ralph:`.

### AC-RCR-006 — Session.StaleSeconds dead pipeline REMOVED (owned by THIS SPEC, not deferred)
**Given** the post-redesign source tree (after M3),
**When** `grep -rn '\.StaleSeconds' --include='*.go' internal/ cmd/ pkg/` runs,
**Then** ZERO matches total — the producer-side pipeline is gone (SessionConfig field, ralphFileWrapper field+tag, loader injection block, defaults entries all deleted by M3) AND there were never any runtime consumers to leave behind;
**And when** `grep -rn 'StaleSeconds\|stale_seconds' internal/config/types.go internal/config/loader.go internal/config/defaults.go` runs,
**Then** zero matches (the field declaration, the wrapper field + yaml tag, the injection block, and both default assignments are all absent);
**And when** `go build ./internal/config/...` runs,
**Then** exit 0 (the dead pipeline removal compiles — no dangling reference survives).

### AC-RCR-007 — build + scoped test green
**Given** the post-redesign source tree,
**When** `go build ./...` runs,
**Then** exit 0;
**And when** `go test ./internal/ralph/... ./internal/config/... ./internal/hook/... ./internal/cli/... ./internal/settings/...` runs,
**Then** exit 0.

### AC-RCR-008 — /moai loop regression guard
**Given** a `/tmp/test-project` initialized from the rebuilt template with a deliberate Go lint error on disk (e.g. an unused import or an ineffectual assignment),
**When** `/moai loop` runs,
**Then** the loop starts, detects the diagnostic, attempts a fix, and converges or aborts per the engine's normal decision flow;
**And when** `ralph.lint_as_instruction: true` is set,
**Then** the lint finding is injected as an instruction (verified by a log line or trace showing the injection path at `internal/hook/post_tool.go`).

### AC-RCR-009 — template + local match on the 5-key surface
**Given** the post-sync state,
**When** `diff .moai/config/sections/ralph.yaml internal/template/templates/.moai/config/sections/ralph.yaml` runs,
**Then** exit 0 (the two files are byte-identical — same 5 keys, same defaults, modulo comments).

## §D.2 Severity Definitions

- **MUST** — blocks run-phase completion.
- **SHOULD** — fixed in the same PR if cheap; otherwise `@MX:DEBT` + follow-up SPEC.

## §D.3 Indirect Verification

- `golangci-lint run` exit code is a sanity check, not a gate (the repo carries baseline warnings).
- The `/moai loop` smoke test (AC-RCR-008) requires a real diagnostic on disk; run-phase must set up a minimal `/tmp` fixture.
- AC-RCR-004's Option A path requires verifying the `InitDependencies` reorder did not introduce a config-loader → engine cycle. If a cycle is found, retreat to Option B-doc and record the blocker in §E.2.

## §D.4 Closure Gate

All MUST ACs MUST be PASS. §E.2 evidence MUST cite verbatim command + output for each AC.

## §D.5 Forward-Looking Checks

- After merge, `moai spec audit` classifies this SPEC as `era_final: true` (V3R6, 3-phase close).
- The prior SPEC's vacuous AC pattern (grepping for struct fields that never existed) MUST NOT recur in any follow-up SPEC. The pattern to use instead: grep for REAL behavior (inert-key absence, live-key presence, build green, loop smoke).
- A future SPEC proposing to extend `RalphEngine.Decide()` with NEW config knobs (beyond the 5 struct fields) MUST cite this SPEC's §C decision and demonstrate a live consumer for each new knob.
