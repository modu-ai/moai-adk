# SPEC-V3R2-RT-002 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                        | Description                                                            |
|---------|------------|-------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial G/W/T conversion of 15 ACs (AC-V3R2-RT-002-01 through -15)     |

---

## Scope

본 문서는 `spec.md` §6 의 15 AC 를 Given/When/Then 형식으로 변환합니다. 각 AC 는 happy-path + edge-case + test-mapping 표기를 포함.

표기 약속:
- **Test mapping** 은 어떤 Go test function (or manual verification step) 이 AC 를 cover 하는지 명시.
- **Sentinel** 은 negative path 에서 test 가 기대하는 literal error 또는 systemMessage 문자열.

---

## AC-V3R2-RT-002-01 — SrcProject deny rule wins over lower tiers

Maps to: REQ-V3R2-RT-002-004, REQ-V3R2-RT-002-005.

### Happy path

- **Given** a `RulesByTier` populated with one rule at `SrcProject`: `{Pattern: "Bash(rm*:*)", Action: DecisionDeny, Source: SrcProject, Origin: ".claude/settings.json"}`
- **And** all other tiers (SrcPolicy/User/Local/Plugin/Skill/Session/Builtin) empty
- **When** `Resolve("Bash", []byte("rm -rf /"), ResolveContext{Mode: ModeDefault, IsInteractive: true})` is called
- **Then** the result has `Decision == DecisionDeny`
- **And** `ResolvedBy == config.SrcProject`
- **And** `Origin == ".claude/settings.json"`
- **And** `Trace.Tries[<idx>].Tier == SrcProject` and `Trace.Tries[<idx>].Matched == true`

### Edge case — multiple tiers all match

- **Given** identical pattern at SrcUser (allow) AND SrcProject (deny) AND SrcLocal (allow)
- **When** Resolve is called
- **Then** `ResolvedBy == SrcUser` (highest priority among matchers)
- **And** the result is `DecisionAllow` (SrcUser 의 action)

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_SrcProjectDenyWins` (existing baseline; verify Origin assertion)
- `internal/permission/resolver_test.go::TestResolve_HigherTierWinsOverLower` (existing/extend)

---

## AC-V3R2-RT-002-02 — Pre-allowlist resolves go test to allow

Maps to: REQ-V3R2-RT-002-006.

### Happy path

- **Given** an empty `RulesByTier` (no user-defined rules in any tier)
- **When** `Resolve("Bash", []byte("go test ./..."), ResolveContext{Mode: ModeDefault, IsInteractive: true})` is called
- **Then** the result has `Decision == DecisionAllow`
- **And** `ResolvedBy == config.SrcBuiltin`
- **And** `Origin == "pre-allowlist"`

### Edge case — pre-allowlist 8 patterns

- **Given** the same empty RulesByTier
- **When** Resolve is called for each of: `Read("foo")`, `Glob("*.go")`, `Grep("pattern")`, `Bash("go test ...")`, `Bash("golangci-lint run")`, `Bash("ruff check .")`, `Bash("npm test")`, `Bash("pytest tests/")`
- **Then** all 8 calls return `Decision == DecisionAllow` with `ResolvedBy == SrcBuiltin` and `Origin == "pre-allowlist"`

### Edge case — pre-allowlist 미커버 input

- **Given** the same empty RulesByTier
- **When** Resolve is called for `Bash("rm test.txt")` (write op, not in allowlist)
- **Then** the result is `DecisionAsk` with `Origin == "no matching rule (default)"`

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_PreAllowlistGoTest` (existing baseline)
- `internal/permission/resolver_test.go::TestResolve_PreAllowlistAllPatterns` (extend, M2)
- `internal/permission/resolver_test.go::TestResolve_NotInAllowlistDefaultsToAsk` (existing/extend)

---

## AC-V3R2-RT-002-03 — Bubble routes Write prompt to parent session

Maps to: REQ-V3R2-RT-002-012.

### Happy path

- **Given** a fork agent spawned with `permissionMode: bubble` under a parent terminal session
- **And** ResolveContext: `{Mode: ModeBubble, IsFork: true, ParentAvailable: true, ForkDepth: 1}`
- **And** `RulesByTier` contains a rule at SrcProject: `{Pattern: "Write(/tmp/*)", Action: DecisionAsk, Source: SrcProject, Origin: ".claude/settings.json"}`
- **When** `Resolve("Write", []byte("/tmp/test.txt"), ctx)` is called
- **Then** the result has `Decision == DecisionAsk`
- **And** `SystemMessage == "Bubble mode: routing to parent session"`
- **And** the bubble dispatch contract is invoked (in production: parent's AskUserQuestion channel; in test: assert ShouldBubble(...) returns true)

### Edge case — fork without parent

- **Given** the same fork agent
- **And** `ResolveContext{IsFork: true, ParentAvailable: false}`
- **When** Resolve is called
- **Then** the result is `DecisionDeny` with `SystemMessage == "Bubble target parent unavailable — decision deferred"`

### Edge case — non-fork in bubble mode

- **Given** `ResolveContext{Mode: ModeBubble, IsFork: false}` (top-level session in bubble mode — uncommon but valid)
- **When** Resolve is called for a rule with `DecisionAsk`
- **Then** routing is NOT triggered (ShouldBubble returns false because IsFork=false)
- **And** the prompt appears in the current session's own AskUserQuestion channel

### Test mapping

- `internal/permission/bubble_test.go::TestBubbleDispatcher_ShouldBubble_ForkWithParent` (existing)
- `internal/permission/resolver_test.go::TestResolve_BubbleParentClosed` (new, M1)
- `internal/permission/resolver_test.go::TestResolve_BubbleNonFork` (new, M1)

---

## AC-V3R2-RT-002-04 — Hook PermissionDecision overrides session/skill/builtin

Maps to: REQ-V3R2-RT-002-011.

### Happy path

- **Given** PreToolUse hook returns `HookResponse{PermissionDecision: "allow"}`
- **And** `RulesByTier[SrcSession]` contains `{Pattern: "Bash(*)", Action: DecisionDeny, ...}` (would normally deny)
- **When** Resolve is called with `ResolveContext{HookResponse: hookResp}`
- **Then** the result has `Decision == DecisionAllow`
- **And** `ResolvedBy == config.SrcBuiltin` (hooks recorded under SrcBuiltin tier)
- **And** `Origin == "PreToolUse hook"`
- **And** the SrcSession deny rule is NOT consulted (hook overlay short-circuits the walk)

### Edge case — hook returns deny

- **Given** the same setup
- **And** hook returns `PermissionDecision: "deny"`
- **And** `RulesByTier[SrcUser]` contains an allow rule
- **When** Resolve is called
- **Then** the result is `DecisionDeny` with `Origin == "PreToolUse hook"`
- **And** the SrcUser allow rule does NOT override (hook overlay wins)

### Edge case — hook returns ask

- **Given** hook returns `PermissionDecision: "ask"` and IsInteractive=true
- **When** Resolve is called
- **Then** the result is `DecisionAsk` with `Origin == "PreToolUse hook"`

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_HookOverridesAllTiers` (existing baseline)
- `internal/permission/resolver_test.go::TestResolve_HookDecisionDeny` (existing/extend)

---

## AC-V3R2-RT-002-05 — `moai doctor permission --trace` emits 8-tier JSON

Maps to: REQ-V3R2-RT-002-007, REQ-V3R2-RT-002-015.

### Happy path

- **Given** `RulesByTier` populated with rules at SrcUser (`Bash(go build:*)` allow) and SrcBuiltin (pre-allowlist)
- **When** the user runs `moai doctor permission --tool Bash --input "go build" --trace`
- **Then** stdout contains a JSON section labelled `--- Resolution Trace (JSON) ---`
- **And** the JSON is parseable (`json.Unmarshal` returns nil error)
- **And** the JSON's `tries[]` array enumerates all tiers visited until first match (in this case, SrcPolicy missed → SrcUser matched → walk halts)
- **And** each `tries[]` entry has `tier`, `matched: true|false`, `rule` (if matched), `reason`

### Edge case — `--all-tiers` flag

- **Given** the same setup
- **When** the user runs `moai doctor permission --all-tiers --tool Bash --input "go build"`
- **Then** stdout dumps all rules from all 8 tiers (regardless of match)
- **And** human-readable format is used (multi-line, indented)

### Edge case — non-matching input

- **Given** `RulesByTier` is empty (only pre-allowlist active)
- **When** `moai doctor permission --tool Bash --input "rm -rf /" --trace`
- **Then** the JSON's `tries[]` enumerates all 8 tiers with `matched: false`
- **And** the final `Decision == DecisionAsk` (or DecisionDeny in non-interactive)
- **And** `Origin == "no matching rule (default)"`

### Test mapping

- `internal/cli/doctor_permission_test.go::TestDoctorPermission_TraceJSONFormat` (new, M1)
- `internal/cli/doctor_permission_test.go::TestDoctorPermission_AllTiersFlag` (new, M1)
- `internal/cli/doctor_permission_test.go::TestDoctorPermission_NoMatchTrace` (new, M5)

---

## AC-V3R2-RT-002-06 — Plan mode denies Write regardless of allowlist

Maps to: REQ-V3R2-RT-002-020.

### Happy path

- **Given** `RulesByTier[SrcUser]` contains a rule explicitly allowing `Write(*)` 
- **And** `ResolveContext{Mode: ModePlan}`
- **When** `Resolve("Write", []byte("/tmp/x"), ctx)` is called
- **Then** the result has `Decision == DecisionDeny`
- **And** `Origin == "plan mode denies writes"`
- **And** `SystemMessage == "plan mode denies writes"`
- **And** the SrcUser allow rule is NOT consulted (plan mode short-circuits)

### Edge case — plan mode + Read tool

- **Given** `ResolveContext{Mode: ModePlan}`
- **When** `Resolve("Read", []byte("/tmp/x"), ctx)` is called
- **Then** the result is `DecisionAllow` (Read is not a write op; pre-allowlist matches)

### Edge case — plan mode + Bash with non-write command

- **Given** `ResolveContext{Mode: ModePlan}`
- **When** `Resolve("Bash", []byte("ls -la"), ctx)` is called
- **Then** the result is `DecisionAsk` (no write detected; falls through 8-tier walk; no rule matches → default ask)

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_PlanModeDeniesWrite` (existing baseline)
- `internal/permission/resolver_test.go::TestResolve_PlanModeAllowsRead` (existing/extend)

---

## AC-V3R2-RT-002-07 — `bypassPermissions` rejected in strict_mode

Maps to: REQ-V3R2-RT-002-022.

### Happy path

- **Given** `.moai/config/sections/security.yaml` has `permission.strict_mode: true`
- **And** an agent is spawned with `permissionMode: bypassPermissions`
- **When** the spawn entry point calls `RejectIfStrict(ModeBypassPermissions, true)` (or `resolver.ValidateMode(ModeBypassPermissions, false, true, 0)`)
- **Then** the call returns an error with sentinel `permission mode rejected: bypassPermissions not allowed in strict mode`
- **And** the agent spawn aborts before any tool invocation

### Edge case — strict_mode false (default)

- **Given** `permission.strict_mode: false`
- **And** the same spawn with bypassPermissions
- **When** RejectIfStrict is called
- **Then** the call returns nil (no error)
- **And** the agent spawns normally; subsequent Resolve calls short-circuit to allow

### Edge case — strict_mode true + ModeAcceptEdits

- **Given** `permission.strict_mode: true`
- **And** spawn with `permissionMode: acceptEdits` (not bypassPermissions)
- **When** RejectIfStrict is called
- **Then** the call returns nil (acceptEdits is allowed under strict_mode)

### Test mapping

- `internal/permission/spawn_test.go::TestRejectIfStrict_RejectsBypass` (new, M2)
- `internal/permission/spawn_test.go::TestRejectIfStrict_AllowsAcceptEdits` (new, M2)
- `internal/permission/resolver_test.go::TestResolve_BypassPermissionsRejectedInStrictMode` (new, M1)

---

## AC-V3R2-RT-002-08 — Bubble parent unavailable causes deny

Maps to: REQ-V3R2-RT-002-050.

### Happy path

- **Given** a fork agent in `ModeBubble` with `IsFork: true, ParentAvailable: false` (parent terminal session closed)
- **And** `RulesByTier[SrcProject]` matches with `Action: DecisionAsk`
- **When** `Resolve("Bash", []byte("git push"), ctx)` is called
- **Then** the result has `Decision == DecisionDeny`
- **And** `SystemMessage == "Bubble target parent unavailable — decision deferred"`
- **And** no AskUserQuestion is dispatched

### Edge case — pre-allowlist match still allows

- **Given** the same fork+ParentAvailable=false
- **When** `Resolve("Read", []byte("file.go"), ctx)` is called
- **Then** the result is `DecisionAllow` (pre-allowlist matches before reaching ask path)

### Edge case — fork transitions to live parent mid-session

- **Given** ParentAvailable transitions from false to true between two Resolve calls
- **When** the second call is made
- **Then** the second call routes normally to parent (current ResolveContext value is consulted, not cached)

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_BubbleParentClosed` (new, M1)
- `internal/permission/bubble_test.go::TestIsParentAvailable_EmptyID` (existing baseline)

---

## AC-V3R2-RT-002-09 — Frontmatter strict-validation rejects unknown permissionMode

Maps to: REQ-V3R2-RT-002-008.

### Happy path

- **Given** an agent file `.claude/agents/moai/<name>.md` with frontmatter `permissionMode: ultra-bypass`
- **When** CI runs `go test ./internal/template/ -run TestAgentFrontmatter_PermissionModeStrictEnum`
- **Then** the test fails with sentinel `PERMISSION_MODE_UNKNOWN_VALUE: <file> declares permissionMode: ultra-bypass; allowed: default|acceptEdits|bypassPermissions|plan|bubble.`

### Edge case — frontmatter omits permissionMode

- **Given** an agent file with no `permissionMode` key in frontmatter
- **When** the audit test runs
- **Then** the agent is implicitly `default` and the test passes (no error)

### Edge case — all 5 valid values

- **Given** five agent files each declaring one of: `default`, `acceptEdits`, `bypassPermissions`, `plan`, `bubble`
- **When** the audit test runs
- **Then** all 5 pass without error

### Edge case — markdown body example

- **Given** an agent file with no frontmatter `permissionMode` but a markdown code block in body containing `permissionMode: example` (illustration)
- **When** the audit test runs
- **Then** the body is NOT scanned (walker only inspects frontmatter delimiter `---` block)
- **And** the test passes

### Test mapping

- `internal/template/agent_frontmatter_audit_test.go::TestAgentFrontmatter_PermissionModeStrictEnum` (new, M2)

---

## AC-V3R2-RT-002-10 — Hook UpdatedInput re-matches against mutated path

Maps to: REQ-V3R2-RT-002-013.

### Happy path

- **Given** PreToolUse hook returns `HookResponse{UpdatedInput: []byte("/safe/path")}` (no PermissionDecision)
- **And** original input is `/dangerous/path`
- **And** `RulesByTier[SrcProject]` contains rule `{Pattern: "Write(/safe/*)", Action: DecisionAllow}`
- **When** Resolve is called with original input `[]byte("/dangerous/path")`
- **Then** the resolver re-runs Resolve internally with the mutated input `/safe/path`
- **And** the rule matches the mutated path → `Decision == DecisionAllow`
- **And** result.UpdatedInput equals `[]byte("/safe/path")` (preserved for downstream)
- **And** `ResolvedBy == SrcProject`

### Edge case — hook UpdatedInput does NOT match any rule

- **Given** hook mutates input to `/elsewhere/path` which matches no allowlist rule
- **When** Resolve re-runs
- **Then** the result is `DecisionAsk` with `Origin == "no matching rule (default)"`

### Edge case — nested mutation prevented

- **Given** hook A returns UpdatedInput; resolver re-runs; hypothetically a chained hook B would mutate further
- **When** Resolve is re-invoked with `newCtx.HookResponse = nil` (explicit clear)
- **Then** no second mutation occurs (hook re-entry blocked by nil-clear at line 149 of resolver.go)
- **And** the test verifies single re-execution only

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_HookUpdatedInputReMatch` (new, M1)
- `internal/permission/resolver_test.go::TestResolve_HookUpdatedInputNoLoop` (new, M3)

---

## AC-V3R2-RT-002-11 — Legacy `bypassPermissions` action migrated with deprecation warning

Maps to: REQ-V3R2-RT-002-040.

### Happy path

- **Given** a v2 settings file at `.claude/settings.json` (or `.moai/settings.json`) contains a rule:
  ```json
  {"pattern": "Bash(*)", "action": "bypassPermissions", "origin": ".claude/settings.json"}
  ```
- **When** the loader reads the rule and calls `MigrateLegacyBypassRules([]PermissionRule{rule})`
- **Then** the returned rule list has the rule with `Action == DecisionAllow` (mapped from action="bypassPermissions" → effective allow)
- **And** the returned warnings slice contains an entry naming `.claude/settings.json` with deprecation message
- **And** stderr emits the deprecation warning at first read
- **And** `.moai/logs/permission.log` contains the same warning entry

### Edge case — no legacy rules present

- **Given** all rules use modern actions (`allow`, `ask`, `deny`)
- **When** MigrateLegacyBypassRules is called
- **Then** the returned warnings slice is empty
- **And** no log entry is added

### Edge case — multiple legacy rules

- **Given** 3 rules with `action: bypassPermissions` from different files
- **When** MigrateLegacyBypassRules is called
- **Then** all 3 rules are migrated to `Action: DecisionAllow`
- **And** 3 distinct warnings are returned (one per origin file)

### Test mapping

- `internal/permission/migration_test.go::TestMigrateLegacyBypassRules_HappyPath` (new, M5b)
- `internal/permission/migration_test.go::TestMigrateLegacyBypassRules_NoLegacy` (new, M5b)
- `internal/permission/migration_test.go::TestMigrateLegacyBypassRules_MultipleOrigins` (new, M5b)

---

## AC-V3R2-RT-002-12 — Same-tier conflict resolves by specificity then fs-order

Maps to: REQ-V3R2-RT-002-042.

### Happy path

- **Given** two rules at the same `SrcLocal` tier match `Bash(git push origin main)`:
  - Rule A: `{Pattern: "Bash(git push*)", Action: DecisionAllow, Origin: "/path/a/settings.local.json"}`
  - Rule B: `{Pattern: "Bash(git push origin main)", Action: DecisionDeny, Origin: "/path/b/settings.local.json"}`
- **When** Resolve is called for `Bash("git push origin main")`
- **Then** Rule B wins (more specific — fewer wildcards)
- **And** the result has `Decision == DecisionDeny` with `Origin == "/path/b/settings.local.json"`
- **And** `.moai/logs/permission.log` contains a conflict log entry naming both rules

### Edge case — equal specificity, fs-order tiebreak

- **Given** two rules with identical specificity (both have one wildcard) but different file origins:
  - Rule A: `Origin == "/zzz/last.json"` (lexicographic last)
  - Rule B: `Origin == "/aaa/first.json"` (lexicographic first)
- **When** Resolve is called and both match
- **Then** Rule A wins (per spec.md REQ-042: "the rule whose Origin file path comes later in filesystem scan order SHALL win")
- **And** the conflict is logged

### Edge case — no conflict (single match)

- **Given** only one rule matches in a tier
- **When** Resolve is called
- **Then** the single rule is returned without conflict resolution
- **And** no log entry is added (conflict log only fires on ≥2 matches)

### Test mapping

- `internal/permission/conflict_test.go::TestResolveConflict_SpecificityWins` (new, M3)
- `internal/permission/conflict_test.go::TestResolveConflict_FsOrderTiebreak` (new, M3)
- `internal/permission/conflict_test.go::TestResolveConflict_SingleMatchNoLog` (new, M3)

---

## AC-V3R2-RT-002-13 — Policy tier deny overrides project allow

Maps to: REQ-V3R2-RT-002-014.

### Happy path

- **Given** `RulesByTier[SrcPolicy]` contains `{Pattern: "Bash(curl:*)", Action: DecisionDeny, Origin: "/etc/moai/policy.json"}`
- **And** `RulesByTier[SrcProject]` contains `{Pattern: "Bash(curl:*)", Action: DecisionAllow, Origin: ".claude/settings.json"}`
- **When** Resolve is called for `Bash("curl -X POST ...")`
- **Then** the result has `Decision == DecisionDeny`
- **And** `ResolvedBy == config.SrcPolicy`
- **And** `Origin == "/etc/moai/policy.json"`
- **And** the SrcProject rule is recorded in trace as `tier: SrcProject, matched: true (but lower priority)` — but actually, since SrcPolicy matches first, the walk halts

### Edge case — only SrcProject has rule

- **Given** SrcPolicy is empty and SrcProject has the same allow rule
- **When** Resolve is called
- **Then** the result is `DecisionAllow` from SrcProject

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_PolicyDenyOverridesProjectAllow` (new, M1)
- `internal/permission/resolver_test.go::TestResolve_HigherTierWinsOverLower` (existing/extend)

---

## AC-V3R2-RT-002-14 — Fork at depth 4 degrades non-plan modes to bubble

Maps to: REQ-V3R2-RT-002-023.

### Happy path

- **Given** `ResolveContext{Mode: ModeAcceptEdits, IsFork: true, ForkDepth: 4}` (depth exceeds limit)
- **When** Resolve is called for any tool
- **Then** the result has `Decision == DecisionAsk`
- **And** `ResolvedBy == config.SrcBuiltin`
- **And** `Origin == "fork depth limit"`
- **And** `SystemMessage == "Fork depth 4 exceeds limit - mode degraded to bubble"`
- **And** the 8-tier walk is short-circuited (does not consult any rule)

### Edge case — depth exactly 3 (boundary)

- **Given** `ForkDepth: 3` (at limit but not exceeding)
- **When** Resolve is called
- **Then** the depth gate does NOT trigger (boundary is `> 3`)
- **And** the 8-tier walk proceeds normally

### Edge case — depth 4 with ModePlan

- **Given** `ForkDepth: 4, Mode: ModePlan`
- **When** Resolve is called for `Write(/tmp/x)`
- **Then** the depth gate is bypassed (ModePlan exempt)
- **And** the result is `DecisionDeny` with `Origin == "plan mode denies writes"` (plan-mode short-circuit)

### Edge case — depth 4 with ModeBubble

- **Given** `ForkDepth: 4, Mode: ModeBubble`
- **When** Resolve is called for any tool
- **Then** the depth gate is bypassed (ModeBubble exempt)
- **And** the 8-tier walk + bubble routing proceeds normally

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_ForkDepth4DegradeToBubble` (new, M1)
- `internal/permission/resolver_test.go::TestResolve_ForkDepth3WithinLimit` (new, M4)
- `internal/permission/resolver_test.go::TestResolve_ForkDepth4PlanMode` (new, M4)

---

## AC-V3R2-RT-002-15 — Non-interactive mode fails closed (ask becomes deny)

Maps to: REQ-V3R2-RT-002-041.

### Happy path

- **Given** `ResolveContext{IsInteractive: false}` (e.g., CI environment)
- **And** `RulesByTier` is empty (no rule matches → would default to ask)
- **When** Resolve is called
- **Then** the result has `Decision == DecisionDeny` (fail closed)
- **And** `Origin == "no matching rule (default)"`
- **And** `.moai/logs/permission.log` contains an entry: `[<timestamp>] Unreachable prompt: tool=<tool> input=<truncated input>`

### Edge case — interactive mode (default)

- **Given** `IsInteractive: true` and the same empty RulesByTier
- **When** Resolve is called
- **Then** the result is `DecisionAsk` (interactive — would dispatch AskUserQuestion at orchestrator)
- **And** no log entry is added

### Edge case — non-interactive but rule matches

- **Given** `IsInteractive: false` and a rule matches with `Action: DecisionAllow`
- **When** Resolve is called
- **Then** the result is `DecisionAllow` (matching rule wins; fail-closed only fires on no-match)
- **And** no log entry is added

### Test mapping

- `internal/permission/resolver_test.go::TestResolve_NonInteractiveAskBecomesDeny` (new, M1)
- `internal/permission/resolver_test.go::TestResolve_NonInteractiveMatchedRuleAllows` (new, M4)
- `internal/permission/resolver_test.go::TestLogUnreachablePrompt_FilePath` (new, M4)

---

## Summary table — AC → REQ → Test

| AC | REQs covered | Test files |
|----|--------------|------------|
| AC-01 | REQ-004, REQ-005 | `resolver_test.go::TestResolve_SrcProjectDenyWins`, `TestResolve_HigherTierWinsOverLower` |
| AC-02 | REQ-006 | `resolver_test.go::TestResolve_PreAllowlist*` (3 cases) |
| AC-03 | REQ-012 | `bubble_test.go::TestBubbleDispatcher_*`, `resolver_test.go::TestResolve_BubbleParentClosed` |
| AC-04 | REQ-011 | `resolver_test.go::TestResolve_HookOverridesAllTiers`, `TestResolve_HookDecisionDeny` |
| AC-05 | REQ-007, REQ-015 | `cli/doctor_permission_test.go::TestDoctorPermission_*` (3 cases) |
| AC-06 | REQ-020 | `resolver_test.go::TestResolve_PlanModeDeniesWrite`, `TestResolve_PlanModeAllowsRead` |
| AC-07 | REQ-022 | `spawn_test.go::TestRejectIfStrict_*`, `resolver_test.go::TestResolve_BypassPermissionsRejectedInStrictMode` |
| AC-08 | REQ-050 | `resolver_test.go::TestResolve_BubbleParentClosed`, `bubble_test.go::TestIsParentAvailable_EmptyID` |
| AC-09 | REQ-008 | `template/agent_frontmatter_audit_test.go::TestAgentFrontmatter_PermissionModeStrictEnum` |
| AC-10 | REQ-013 | `resolver_test.go::TestResolve_HookUpdatedInputReMatch`, `TestResolve_HookUpdatedInputNoLoop` |
| AC-11 | REQ-040 | `migration_test.go::TestMigrateLegacyBypassRules_*` (3 cases) |
| AC-12 | REQ-042 | `conflict_test.go::TestResolveConflict_*` (3 cases) |
| AC-13 | REQ-014 | `resolver_test.go::TestResolve_PolicyDenyOverridesProjectAllow`, `TestResolve_HigherTierWinsOverLower` |
| AC-14 | REQ-023 | `resolver_test.go::TestResolve_ForkDepth4DegradeToBubble`, `TestResolve_ForkDepth3WithinLimit`, `TestResolve_ForkDepth4PlanMode` |
| AC-15 | REQ-041 | `resolver_test.go::TestResolve_NonInteractiveAskBecomesDeny`, `TestResolve_NonInteractiveMatchedRuleAllows`, `TestLogUnreachablePrompt_FilePath` |

Total new test functions: **~30 across 4 new test files (`spawn_test.go`, `conflict_test.go`, `migration_test.go`, `cli/doctor_permission_test.go`) + 3 existing files extended (`resolver_test.go`, `bubble_test.go`, `agent_frontmatter_audit_test.go`)**.

---

## Definition of Done

본 SPEC 은 다음 모두 true 일 때 done:

1. 위 15 AC 가 모두 `go test ./internal/permission/ ./internal/cli/ ./internal/template/` 에서 PASS.
2. 워크트리 루트의 `go test ./...` 전체 PASS, 0 cascading regressions.
3. `make build` 성공, `internal/template/embedded.go` clean 재생성 (security.yaml 미러 변경 반영).
4. `go vet ./...` 와 `golangci-lint run` 0 warnings.
5. `progress.md` 의 `run_complete_at: <timestamp>` 와 `run_status: implementation-complete` 갱신.
6. CHANGELOG entry 가 `## [Unreleased] / ### 추가` 아래 존재.
7. 7 MX tags 가 plan.md §6 대로 삽입 (3 ANCHOR, 2 NOTE, 2 WARN).
8. `.moai/config/sections/security.yaml` 와 template mirror 의 `permission.{strict_mode, pre_allowlist, session_rules}` 키 정합.
9. `manager-git` 가 연 PR 의 모든 required CI checks GREEN (Lint, Test ubuntu/macos/windows, Build all 5, CodeQL).

---

End of acceptance.md.
