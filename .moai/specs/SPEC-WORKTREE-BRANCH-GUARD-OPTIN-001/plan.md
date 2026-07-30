# Plan — SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001

> Implementation plan. Worktree-relative paths. Order: most-reversible / highest-change-likelihood decisions first (data-model + config wiring), mechanical edits later.

## §A Context

Plan-phase artifacts for the opt-in conversion of the Main-Checkout Branch-State Guard. The guard is template-wired (ships to all users) and currently has no config gate. Two behavior defects are addressed: (1) default-on interference for single-developer users, (2) over-broad regex patterns that match read-only git subcommands. The decision (confirmed with maintainer GOOS): default-OFF + pattern refinement. Exemption logic and fail-open norm preserved.

Run-phase module: `internal/hook` (primary), `internal/config` (schema), rule docs (mirror).

## §B Known Issues

1. **`git stash list` matches the stash pattern** because the optional group `(\s+(push|pop|apply|drop)\b)?` matches zero times. The existing comment at `branch_guard.go:58-60` claims it is excluded, but the regex does not enforce it. The fix is a negative lookahead in the optional subcommand position.
2. **`git merge-base` matches `git merge`** because `\bgit\s+merge\b` has a word boundary between `merge` and `-`. The fix anchors with whitespace-or-end (`\s|$`) after `merge`.
3. **No config gate** exists in `internal/config` or `internal/hook` today — a new `BranchGuardConfig` struct + default + provider read path is required.

## §C Pre-Flight (Run-Phase Entry Checks)

Run before M1:
- `go test ./internal/hook/... ./internal/config/...` — green baseline.
- `make build` — template embed compiles.
- `rg -n "MOAI_BRANCH_GUARD_EXEMPT|AgentType == \"manager-git\"" internal/hook/branch_guard.go` — confirm exemption anchors before editing (must return 2+ matches; do NOT weaken them).

## §D Constraints (Hard)

- Template default stays `enabled: false`. No `enabled: true` anywhere in `internal/template/templates/`.
- `isExemptAgent` body unchanged (exemption additive ABOVE the config gate in the deny decision order? — see design decision §E below: NO, gate sits BEFORE exemption in the call path but exemption semantics are unchanged; when disabled, neither path runs and that is correct — disabled means inert).
- fail-open path (`branch_guard.go:215-221`) unchanged for the enabled path.
- Mirror parity test (`rule_template_mirror_test.go`) and §25 sanitized-pair tests must stay green.

## §E Self-Verification (Design Decisions)

### Decision D1 — Gate at call site, NOT inside `checkBranchState`

**Choice**: gate the call to `checkBranchState` at the call site in `preToolHandler.Handle` (`pre_tool.go:454-463`), reading the flag from `h.cfg.Get()` via a new accessor.

**Rationale**:
- The `preToolHandler` already holds a `ConfigProvider` (`pre_tool.go:315`) and already has a `loadGateConfig()` pattern (`pre_tool.go:609`) that calls `h.cfg.Get()`. Reusing that pattern is the lowest-friction injection.
- `checkBranchState` is a free function that takes `(input, projectDir)`; adding a config arg would change its signature and cascade into test callers (M2's deny-origin test, M6's pattern-set tests). Keeping the signature stable preserves test call sites and isolates the behavior change.
- The gate is a one-line early-return at the call site: `if !h.cfg.Get().Workflow.BranchGuard.Enabled { /* skip */ }`. Minimal diff.

**Rejected alternative**: gate INSIDE `checkBranchState` by reading a package-level config or threading a new arg. Rejected because it either introduces global mutable state (test-isolation hazard per `internal/hook/CLAUDE.md` OTEL rule family) or breaks the existing signature.

### Decision D2 — Config struct placement

**Choice**: add `BranchGuardConfig{ Enabled bool }` as a nested field of `WorkflowConfig` (NOT a top-level section), keyed `branch_guard:` under `workflow:` in `workflow.yaml`. Neighbor: `WorkflowConfig.Worktree` (`types.go:365`).

**Rationale**:
- The branch guard is a workflow-policy knob (it gates a workflow hook behavior), so it belongs under `workflow.*` alongside `worktree.*`.
- A nested struct avoids creating a new section file (`.moai/config/sections/branch_guard.yaml`) — smaller blast radius, no loader wiring.
- Default literal goes into `NewDefaultWorkflowConfig` (`defaults.go:520`) next to the `Worktree: WorkflowWorktreeConfig{...}` block.

### Decision D3 — Pattern refinement via regex tightening (NOT denylist)

**Choice**: replace the two offending patterns with tightened regexes that anchor on whitespace/end after the dangerous token, rather than adding a separate denylist of read-only exceptions.

**Rationale**:
- A denylist (`git merge-base`, `git stash list`, `git stash show`, ...) grows over time and is reactive (every new read-only subcommand must be added). Anchoring is proactive.
- Tightened forms:
  - `git stash`: keep the optional mutating-subcommand group, but require that if a token follows and is NOT one of the mutating forms, it must not be present at all — implemented by matching bare `\bgit\s+stash\b` OR `\bgit\s+stash\s+(push|pop|apply|drop)\b`. This is the SAME surface regex, but the optional group must be made non-optional-when-token-follows. Concretely: `\bgit\s+stash(\s+(push|pop|apply|drop))?\b` already requires word-boundary AFTER the optional group; since `list` is `\blist\b`, the pattern `\bgit\s+stash(\s+(push|pop|apply|drop)\b)?\b` would NOT match `git stash list` because after `stash` the next token `list` is not in the set and the trailing `\b` would require the string to end at `stash`. **Verified reasoning**: the tightening is achieved by adding a `(?!.*\S)` style anchor or by switching to `\bgit\s+stash(\s+(push|pop|apply|drop)\b)?\s*$` — but `$` is wrong for compound commands. The clean form: `\bgit\s+stash(\s+(push|pop|apply|drop)\b)?(\s+--\S+)*\s*$` is too brittle. **The implementer SHOULD pick the cleanest regex** (likely: `\bgit\s+stash\b(\s+(push|pop|apply|drop)\b)?` AND a negative-following-token check, OR a tighter pattern). (REQ-2 mandates behavior, not syntax; the implementer picks any regex that passes AC-REQ-2a..2e.)
  - `git merge`: change to `\bgit\s+merge\s` (requires whitespace after `merge`, so `merge-base` does not match) OR `\bgit\s+merge(\s|$)`. The `\s` form is the minimal fix.

**Rejected alternative**: denylist of read-only forms. Rejected for the growth reason above.

### Decision D4 — Maintainer opt-in location

**Choice**: maintainer's `enabled: true` lives in `.moai/config/sections/workflow.local.yaml` (gitignored family) OR a `CLAUDE.local.md`-documented local config path. The template default in `internal/template/templates/.moai/config/sections/workflow.yaml` stays `enabled: false` (or, if the key is absent, zero-value `false` from the Go default).

**Rationale**: mirrors the `CLAUDE.local.md §22` family pattern (dev-settings intent). The local config family is the documented channel for maintainer-only intent that does not distribute.

## §F Milestones (Ordered by Decision-Reversibility)

### M1 — Add `BranchGuardConfig` schema + default (data-model decision)

**Files** (worktree-relative):
- `internal/config/types.go` — add `BranchGuardConfig` struct near `WorkflowWorktreeConfig` (line ~476) and a `BranchGuard BranchGuardConfig \`yaml:"branch_guard"\`` field on `WorkflowConfig` (near `Worktree` at line ~365).
- `internal/config/defaults.go` — set `BranchGuard: BranchGuardConfig{Enabled: false}` in `NewDefaultWorkflowConfig` (near line ~520).

**Exit**: `go build ./internal/config/...` green; `go test ./internal/config/...` green; new struct compiles and defaults to false.

**Why first**: the data model is the most-change-likely decision. If the maintainer wants a different struct shape (e.g., top-level `branch_guard:` section instead of nested), catching it here is cheapest.

### M2 — Refine `branchStatePatterns` regexes (behavior decision)

**Files**:
- `internal/hook/branch_guard.go:81-88` — replace the `stash` and `merge` patterns with tightened forms per Decision D3. Add a comment explaining why each pattern excludes the read-only forms.
- `internal/hook/branch_guard_test.go` (or the existing pattern-set test file) — add cases: `git stash list`, `git stash show`, `git merge-base`, `git merge-base --is-ancestor` MUST NOT match; `git stash`, `git stash push`, `git stash pop`, `git stash apply`, `git stash drop`, `git merge feature` MUST still match.

**Exit**: `go test ./internal/hook/... -run BranchState` green; falsifiable per AC-REQ-2.

**Why second**: regex behavior is the next-most-change-likely decision (regex syntax choice, anchor strategy).

### M3 — Wire the config gate at the call site

**Files**:
- `internal/hook/pre_tool.go:454-463` — add an early-skip reading `h.cfg.Get().Workflow.BranchGuard.Enabled`. Pattern:

  ```go
  if input.ToolName == "Bash" && len(input.ToolInput) > 0 {
      if !h.branchGuardEnabled() {
          // disabled (default) — guard inert, fail-open trivially preserved
      } else if decision, reason := checkBranchState(input, h.projectDir); decision == DecisionDeny {
          ...
      }
  }
  ```

  with a helper `func (h *preToolHandler) branchGuardEnabled() bool` that reads the provider defensively (nil → false → inert).

- `internal/hook/pre_tool_test.go` — add cases: (a) `Enabled: false` + `git stash list` → allow (not denied); (b) `Enabled: false` + `git reset --hard` → allow (not denied); (c) `Enabled: true` + `git reset --hard` in primary checkout + non-exempt → deny; (d) `Enabled: true` + `MOAI_BRANCH_GUARD_EXEMPT=1` → allow (exemption still works per REQ-6).

**Exit**: `go test ./internal/hook/...` green; falsifiable per AC-REQ-1, AC-REQ-3, AC-REQ-6.

**Why third**: depends on M1 (struct) + M2 (patterns) — but sits above the mechanical doc/test work.

### M4 — Full test sweep + characterization

**Files**:
- `internal/hook/branch_guard_test.go` — extend the existing M6 deny-origin / AC-WBG tests to also cover the disabled-default path and the read-only exclusions.
- `internal/config/workflow_test.go` (or `types_test.go`) — assert `NewDefaultWorkflowConfig().BranchGuard.Enabled == false`.

**Exit**: `go test ./internal/hook/... ./internal/config/...` green; `go test -race ./internal/hook/...` green.

### M5 — Rule mirror documentation (both sides)

**Files**:
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — source-side update. Document: (a) the guard is now DEFAULT-OFF (opt-in), (b) maintainer enables via local `workflow.branch_guard.enabled: true`, (c) pattern refinement excludes `git stash list/show` and `git merge-base`, (d) exemption logic + fail-open unchanged. Retain SPEC/REQ tokens for traceability (§25-sanitized pair source side).
- `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md` — template-side update. Mirror the doctrine text but sanitized (NO SPEC-ID / REQ tokens per CLAUDE.local.md §25).
- `internal/template/rule_template_mirror_test.go` — NO edit needed (the file is already excluded from the byte-parity allowlist and covered by the §25 sanitized-pair registry). Verify by running the mirror test.

**Exit**: both rule files updated; `go test ./internal/template/...` green (mirror parity + sanitized-pair parity + no-internal-content-leak).

### M6 — Maintainer local config + CLAUDE.local.md doc

**Files** (local-only, NOT template):
- `.moai/config/sections/workflow.local.yaml` (or the existing local config family) — add `workflow.branch_guard.enabled: true` for the maintainer's own machine. (This file is gitignored per CLAUDE.local.md §2 / §22 family — verify path with the maintainer's existing local config inventory.)
- `CLAUDE.local.md §22` family — add a new subsection documenting the maintainer intent: `enabled: true` is the maintainer's opt-in; the template default stays `false`; the key lives in local-only config and is NOT mirrored to the template.

**Exit**: maintainer's primary checkout has the guard active (verify with a targeted `git stash list` non-deny in a SECONDARY worktree path AND a deny on `git reset --hard` in the primary checkout — though testing the deny path requires the maintainer's real primary checkout, which is out of scope for automated CI).

### M7 — Verification + commit

**Commands**:
- `go test ./...` — full suite green.
- `go test -race ./internal/hook/... ./internal/config/...` — race-clean.
- `golangci-lint run --timeout=2m` — clean.
- `make build` — template embed compiles.
- `rg -n "enabled: true" internal/template/templates/` — MUST return 0 matches in the branch-guard context (template-neutrality check, CLAUDE.local.md §25).

**Commit**: single commit on the SPEC's feature branch (or split per milestone if the team prefers — the SPEC is Tier M, a single commit is acceptable).

## §G Anti-Patterns (Avoid)

- **AP-1**: Adding `enabled: true` to ANY template file. Template neutrality (§25) requires the distributed default to be `false`.
- **AP-2**: Weakening `isExemptAgent` or removing the `MOAI_BRANCH_GUARD_EXEMPT` check. The exemption is ADDITIVE and MUST remain.
- **AP-3**: Denylist approach for read-only forms — grows unbounded; prefer regex tightening.
- **AP-4**: Threading the config arg through `checkBranchState`'s signature — cascades into test call sites and breaks the M2 deny-origin test. Gate at the call site instead.
- **AP-5**: Editing only one side of the rule mirror — both source and template must be updated in the same commit (or the §25 sanitized-pair parity test fails).
- **AP-6**: Removing the `branch_guard.go` comment about list-only exclusion — UPDATE it to reflect the new regex so future readers are not misled (the old comment claimed exclusion that the regex did not enforce).

## §H Cross-References

- SPEC: `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001/spec.md`
- Acceptance: `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001/acceptance.md`
- Related SPEC: SPEC-WORKTREE-BRANCH-GUARD-001 (the completed guard SPEC).
- Doctrine: `.claude/rules/moai/workflow/main-checkout-branch-guard.md` (source) + template mirror.
- CLAUDE.local.md §22 (dev settings intent), §25 (template neutrality).
