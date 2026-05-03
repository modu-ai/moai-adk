# SPEC-V3R3-HYBRID-001 Acceptance Criteria (Phase 1B)

> Given/When/Then scenarios for `moai hybrid` Multi-LLM Mode acceptance.
> Companion to `spec.md` v0.1.0 and `plan.md` v0.1.0.
> All 18 ACs cover the 18 REQs (1-to-1 minimum mapping confirmed in plan.md §1.4).

## HISTORY

| Version | Date       | Author              | Description                                                  |
|---------|------------|---------------------|--------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | 18 G/W/T scenarios + edge cases for SPEC-V3R3-HYBRID-001     |

---

## 1. Acceptance Criteria

### AC-HYBRID-01: `moai hybrid --help` lists 4 allow-list providers (REQ-HYBRID-001, 002)

**Given** a user has installed moai-adk-go ≥ v3R3 first minor release
**And** the user is at any shell prompt (tmux not required for help output)

**When** the user runs `moai hybrid --help`

**Then** stdout contains the synopsis line `moai hybrid <provider> [-p profile] [-- claude-args...]`
**And** stdout contains the description of all four allow-list providers: `glm`, `kimi`, `deepseek`, `qwen`
**And** exit code is 0

**Edge cases**:
- `moai hybrid -h` (short flag): same output.
- `moai hybrid` (no args): cobra usage hint + exit code 0 (cobra default for missing required arg with hint mode).
- Help output references SPEC-V3R3-HYBRID-001 §10 BC migration (one-line link).

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridHelpListsFourProviders`

---

### AC-HYBRID-02: Unknown provider rejected (REQ-HYBRID-002)

**Given** the 4-provider allow-list (`glm`, `kimi`, `deepseek`, `qwen`) is enforced

**When** the user runs `moai hybrid mistral` (or any non-allow-list value)

**Then** exit code is non-zero
**And** stderr contains the sentinel string `UNKNOWN_PROVIDER`
**And** stderr lists the four allowed values: `glm`, `kimi`, `deepseek`, `qwen`
**And** Claude Code is NOT launched

**Edge cases**:
- `moai hybrid GLM` (uppercase): rejected (case-sensitive); stderr suggests lowercase.
- `moai hybrid glm-` (with trailing dash): rejected.
- `moai hybrid ""` (empty string): cobra "argument required" error, distinct from UNKNOWN_PROVIDER.
- `moai hybrid setup mistral sk-xxx`: setup subcommand also rejects unknown provider.

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridUnknownProviderRejected`

---

### AC-HYBRID-03: `moai cg` returns BC error after upgrade (REQ-HYBRID-003, 010, 014)

**Given** a user has upgraded from moai-adk-go v2.x (where `moai cg` was active) to v3R3

**When** the user runs `moai cg` (with or without args, e.g., `moai cg -p work`)

**Then** exit code is non-zero
**And** stderr contains the sentinel string `MOAI_CG_REMOVED`
**And** stderr contains the actionable suggestion `use 'moai hybrid glm' instead`
**And** stderr references `SPEC-V3R3-HYBRID-001 §10 BC Migration` (one line)
**And** Claude Code is NOT launched (no `unifiedLaunch` invocation)
**And** `cgCmd` is still registered to rootCmd (REQ-HYBRID-014: BC stub, not silently absent)

**Edge cases**:
- `moai cg --help`: prints help text describing the removal + migration path (does NOT execute `runCGRemoved`).
- `moai cg -p profile_name`: BC error path; profile flag is ignored.
- `moai cg -- claude-args`: BC error path; claude-args are not forwarded.

**Test Anchor**:
- `internal/cli/cg_removal_test.go TestCGCommandReturnsBCError`
- `internal/cli/cg_removal_test.go TestCGCommandIsRegistered`
- `internal/cli/cg_removal_test.go TestCGCommandDoesNotLaunchClaudeCode`

---

### AC-HYBRID-04: `.moai/config/sections/llm.yaml` schema includes `provider` + 4-provider config (REQ-HYBRID-004)

**Given** the embedded template `.moai/config/sections/llm.yaml` rendered after `moai init` or `moai update`

**When** a user (or test) parses the YAML file

**Then** the file contains a top-level `provider:` key (default empty string)
**And** the file contains a top-level `providers:` map
**And** `providers.glm.{base_url, env_var, models.{high, medium, low}}` is present and populated
**And** `providers.kimi.{base_url, env_var, models.{high, medium, low}}` is present
**And** `providers.deepseek.{base_url, env_var, models.{high, medium, low}}` is present
**And** `providers.qwen.{base_url, env_var, models.{high, medium, low}}` is present

**Edge cases**:
- Existing `glm:` section (legacy, v2.x baseline) is preserved as alias for v3R3 lifetime.
- `team_mode: ""` (default) and `provider: ""` (default) — neither hybrid nor legacy cg active.
- Comments in YAML reference SPEC-V3R3-HYBRID-001 for migration context.

**Test Anchor**: `internal/llm/registry_test.go TestEmbeddedLLMYAMLSchema`

---

### AC-HYBRID-05: SPEC-pinned baseline models match research.md §3 (REQ-HYBRID-005)

**Given** the embedded template `.moai/config/sections/llm.yaml` is loaded

**When** a test reads `providers.<name>.models.high` for each of the 4 providers

**Then** the values match the SPEC-pinned baseline (per research.md §3 verification):
- `providers.glm.models.high == "glm-4.7"` (or substring match for case-insensitive variant `GLM-4.7`)
- `providers.kimi.models.high == "kimi-k2.6"` (substring match acceptable)
- `providers.deepseek.models.high == "deepseek-v4-pro"`
- `providers.qwen.models.high == "qwen3-coder-plus"` (or current verified value at run-phase entry)

**And** `providers.<name>.models.low` similarly matches the pinned low-tier values

**And** users can override via `~/.moai/config/sections/llm.yaml` `providers.<name>.models.high: "<custom>"` (project-layer override)

**Edge cases**:
- Provider releases new model (e.g., GLM-4.8): user override path remains stable; SPEC pinned baseline updated in next minor release.
- Provider deprecates model (e.g., DeepSeek deprecates `deepseek-v4-pro` in 2026-Q3): SPEC update + CHANGELOG note.

**Test Anchor**: `internal/llm/registry_test.go TestPinnedBaselineModels`

---

### AC-HYBRID-06: `moai glm` and `moai cc` preserve v2.x behavior (REQ-HYBRID-006)

**Given** a user has upgraded from v2.x to v3R3
**And** the user has not modified `.moai/config/sections/llm.yaml` since upgrade

**When** the user runs `moai glm` (single-LLM all-GLM mode)

**Then** the launch behavior matches v2.x baseline bit-for-bit (no functional regression):
- GLM env-vars injected to `settings.local.json`.
- `team_mode: "glm"` written to llm.yaml.
- Claude Code launched with GLM as the active LLM for the lead session.
- All existing `internal/cli/glm_test.go` tests pass.

**And When** the user runs `moai cc`

**Then** the launch behavior matches v2.x baseline:
- GLM env-vars cleared (with `MOAI_BACKUP_AUTH_TOKEN_*` restoration if applicable).
- `team_mode: ""` (cleared).
- Claude Code launched with Claude as the active LLM.

**Edge cases**:
- `moai glm setup sk-xxx`: still saves to `~/.moai/.env.glm` (legacy path preserved).
- `moai glm status`: still shows masked GLM key from `~/.moai/.env.glm`.

**Test Anchor**: existing `internal/cli/glm_test.go TestRunGLM`, `internal/cli/launcher_test.go TestRunCC` unchanged + new `internal/cli/hybrid_test.go TestNonHybridCommandsUnaffected` regression sentinel.

---

### AC-HYBRID-07: `moai hybrid <provider>` happy-path launch (REQ-HYBRID-007)

**Given** the user is inside a tmux session (`$TMUX` env-var set)
**And** `~/.moai/.env.glm` exists and contains a valid `GLM_API_KEY`
**And** `.moai/config/sections/llm.yaml` exists at project root

**When** the user runs `moai hybrid glm` (or `moai hybrid glm -p work`)

**Then** the CLI:
- (a) Loads the GLM API key from `~/.moai/.env.glm`.
- (b) Injects `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic`, `ANTHROPIC_DEFAULT_*_MODEL` into the active tmux session (verifiable via `tmux show-environment`).
- (c) Writes `provider: "glm"` and `team_mode: "hybrid"` to `.moai/config/sections/llm.yaml`.
- (d) Sets `teammateMode: "tmux"` in `.claude/settings.local.json`.
- (e) Launches Claude Code via `unifiedLaunch(profileName, "claude_hybrid", filteredArgs)`.
- (f) Prints success card with provider name, base URL, and "next steps".

**Edge cases**:
- `--proxy <url>` flag overrides `ANTHROPIC_BASE_URL` (see AC-HYBRID-13).
- `-p profile` flag forwarded to launcher.
- Existing `MOAI_BACKUP_AUTH_TOKEN_GLM` (provider-namespaced) preserved across re-runs.

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridGLMHappyPath`

---

### AC-HYBRID-08: `moai hybrid <provider>` outside tmux is rejected (REQ-HYBRID-008)

**Given** the user is NOT inside a tmux session (`$TMUX` env-var not set)
**And** `MOAI_TEST_MODE` is not set to `1` (test bypass disabled)

**When** the user runs `moai hybrid kimi` (or any provider)

**Then** exit code is non-zero
**And** stderr contains the sentinel string `HYBRID_REQUIRES_TMUX`
**And** stderr contains the recovery path: `tmux new -s moai && moai hybrid <provider>`
**And** stderr suggests `moai glm` as alternative for non-tmux single-LLM mode (if the user intended GLM)
**And** Claude Code is NOT launched

**Edge cases**:
- `MOAI_TEST_MODE=1` env-var bypasses the tmux check (for test environments only).
- Inside a non-tmux multiplexer (e.g., zellij, screen): same rejection (only tmux session-level env-var injection is supported).

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridRequiresTmux`

---

### AC-HYBRID-09: Missing API key rejected with setup hint (REQ-HYBRID-009)

**Given** the user is inside a tmux session
**And** `~/.moai/.env.deepseek` does NOT exist or is empty

**When** the user runs `moai hybrid deepseek`

**Then** exit code is non-zero
**And** stderr contains the sentinel string `HYBRID_MISSING_API_KEY`
**And** stderr contains the setup instruction `moai hybrid setup deepseek <api-key>` (or `moai hybrid setup deepseek` for interactive prompt)
**And** stderr references the env-var fallback `DEEPSEEK_API_KEY`
**And** Claude Code is NOT launched

**Edge cases**:
- `~/.moai/.env.deepseek` exists but key is empty string: same rejection.
- `DEEPSEEK_API_KEY` env-var is set in shell env: launch proceeds (env-var fallback per existing pattern in `getProviderAPIKey`).
- File permissions deny read: distinct error path (file-system error, not `HYBRID_MISSING_API_KEY`).

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridMissingAPIKey`

---

### AC-HYBRID-10: `moai update` migrates `team_mode: cg` (REQ-HYBRID-011)

**Given** a project with `.moai/config/sections/llm.yaml` containing `team_mode: "cg"` and `provider: ""` (or `provider` field absent — pre-v3R3 schema)

**When** the user runs `moai update` for the first time after upgrading to v3R3

**Then** post-update, the file contains:
- `team_mode: "hybrid"`
- `provider: "glm"`
**And** stderr emits a one-time migration notice referencing SPEC-V3R3-HYBRID-001 §10
**And** the migration is **idempotent** (running `moai update` a second time does not re-emit the notice nor modify the file)
**And** atomic-write semantics are preserved (no partial state visible to concurrent readers — per SPEC-V3R3-UPDATE-CLEANUP-001 atomic write infrastructure)

**Edge cases**:
- `team_mode: "claude"` or `"glm"` (non-cg): no migration; file unchanged.
- `team_mode: "cg"` AND `provider: "glm"` (already partial migration): only the `team_mode` field is updated; no notice (idempotent).
- `team_mode: "cg"` AND user has manually set `provider: "kimi"`: no rewrite (user choice respected); stderr emits a warning that `team_mode: "cg"` is incompatible with `provider != "glm"`.

**Test Anchor**:
- `internal/cli/migration_test.go TestMigrateCGTeamMode`
- `internal/cli/migration_test.go TestMigrateCGTeamModeIdempotent`

---

### AC-HYBRID-11: Active provider env-vars visible to teammates (REQ-HYBRID-012)

**Given** the user has successfully run `moai hybrid qwen` (assuming Qwen3 endpoint is configured or `--proxy` is supplied)
**And** Claude Code has launched with `team_mode: "hybrid"` + `provider: "qwen"`

**When** a Claude Code teammate spawns in a new tmux pane (Agent Teams mode)

**Then** the teammate's env shows:
- `ANTHROPIC_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/anthropic` (or the `--proxy` URL if used)
- `ANTHROPIC_AUTH_TOKEN=<masked-but-functional Qwen3 key>`
- `ANTHROPIC_DEFAULT_OPUS_MODEL=qwen3-coder-plus` (or pinned baseline value)

**And** the teammate sends LLM requests to the Qwen3 endpoint (verifiable via packet capture in test environment, or via observed teammate behavior matching Qwen3-Coder model traits)

**And** the leader pane (Claude) is unaffected — leader's env still points to Anthropic API.

**Edge cases**:
- Teammate spawned in same pane (not a new pane): inherits leader env (Claude), not Qwen — this is the correct behavior per `moai cg` legacy architectural choice.
- Multiple teammates in different panes: all inherit the same Qwen3 env (single active provider per session).
- Switching provider mid-session via `moai hybrid kimi` from a different pane: leader pane's existing teammates remain on Qwen3; only new panes inherit Kimi env.

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridProviderTeammateInheritance` (mocked tmux for unit test; integration test in `tests/integration/`).

---

### AC-HYBRID-12: Endpoint failure propagates without silent fallback (REQ-HYBRID-013)

**Given** the user has a valid GLM API key and is inside tmux
**And** the GLM endpoint `https://api.z.ai/api/anthropic` is unreachable (simulated via DNS block, network partition, or HTTP 503 response)

**When** the user runs `moai hybrid glm`

**Then** the underlying HTTP error (DNS NXDOMAIN, connection refused, HTTP 503, etc.) is propagated to stderr
**And** the CLI does NOT silently fall back to a different provider (`kimi`, `deepseek`, `qwen` are not auto-tried)
**And** the CLI does NOT silently fall back to Claude (Anthropic API)
**And** exit code is non-zero
**And** stderr suggests the user verify the provider's status page (e.g., status.z.ai for GLM)

**Edge cases**:
- HTTP 401/403 (auth failure): handled separately by AC-HYBRID-16 (HYBRID_AUTH_FAILED).
- HTTP 429 (rate limit): propagated as-is; no auto-retry (out of scope for this SPEC).
- Slow response (timeout): controlled by `API_TIMEOUT_MS=3000000` (3000 sec) — same as legacy CG pattern.

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridEndpointUnreachableNoFallback`

---

### AC-HYBRID-13: `--proxy` overrides ANTHROPIC_BASE_URL (REQ-HYBRID-015)

**Given** the user is inside tmux with valid Kimi API key
**And** the user wants to route Kimi requests through a self-hosted Anthropic-compat adapter

**When** the user runs `moai hybrid kimi --proxy https://my-proxy.example.com/anthropic`

**Then** the tmux session env shows `ANTHROPIC_BASE_URL=https://my-proxy.example.com/anthropic`
**And** the SPEC-pinned `providers.kimi.base_url` is NOT used (proxy takes precedence)
**And** the launch otherwise proceeds normally
**And** `ANTHROPIC_AUTH_TOKEN` still uses the Kimi API key (proxy is responsible for forwarding auth)

**Edge cases**:
- `--proxy` URL invalid (e.g., `htp://malformed`): rejected before launch with clear error.
- `--proxy` URL with trailing slash vs no trailing slash: normalized; both work.
- `--proxy` provided AND provider has 1st-class Anthropic-compat (e.g., `moai hybrid glm --proxy <url>`): proxy still wins (user override always wins).

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridProxyOverride`

---

### AC-HYBRID-14: `moai hybrid <provider> tools` surface stub (REQ-HYBRID-016)

**Given** the v3R3 first minor release has shipped
**And** SPEC-GLM-MCP-001 implementation has NOT yet shipped (PR #769 still plan-only or implementation-pending)

**When** the user runs `moai hybrid glm tools --help`

**Then** the help output lists `enable` and `disable` subcommand stubs
**And** the help text references SPEC-GLM-MCP-001 (PR #769) as the implementation source
**And** running `moai hybrid glm tools enable` returns a stub message: `tool integration is implemented in SPEC-GLM-MCP-001 follow-up; this is a forward-looking surface`

**And When** SPEC-GLM-MCP-001 implementation ships (post-merge of GLM-MCP-001 follow-up PR)

**Then** `moai hybrid glm tools enable` attaches Z.AI Vision/WebSearch/WebReader MCP servers (per GLM-MCP-001 REQ definition).

**Edge cases**:
- `moai hybrid kimi tools enable`: stub message explaining that tool integration is GLM-specific (Z.AI MCP server); Kimi/DeepSeek/Qwen tool integration is future SPEC.
- `moai hybrid glm tools list`: stub message; no list yet.

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridToolsSubcommandStub`

---

### AC-HYBRID-15: 5th provider PR rejected by allow-list audit (REQ-HYBRID-017)

**Given** a future PR adds a new file `internal/llm/providers/mistral.go` (or any non-allow-list provider)

**When** CI runs `go test ./internal/template/ -run TestProviderAllowlist`

**Then** the test FAILS with sentinel string `PROVIDER_ALLOWLIST_VIOLATION`
**And** the failure message references the four allowed values: `glm`, `kimi`, `deepseek`, `qwen`
**And** the failure message references SPEC-V3R3-HYBRID-001 REQ-HYBRID-017 + the atomic-reversal SPEC requirement

**Edge cases**:
- PR adds `internal/llm/providers/mistral_test.go` (test file only): test ignores `_test.go` suffix; only `.go` files counted.
- PR adds `internal/llm/providers/registry_test.go`: ignored (registry helper, not a provider).
- PR modifies existing `glm.go` (no new file): test passes (still 4 entries).
- PR adds `internal/llm/providers/glm_v2.go` (variant): test fails because regex/pattern targets the 4 specific filenames (`glm.go`, `kimi.go`, `deepseek.go`, `qwen.go`).

**Test Anchor**: `internal/template/provider_allowlist_audit_test.go TestProviderAllowlist`

---

### AC-HYBRID-16: HTTP 401 masks API key (REQ-HYBRID-018)

**Given** the user runs `moai hybrid glm` with an invalid GLM API key
**And** the Z.AI endpoint returns HTTP 401 (or 403)

**When** the launch path receives the auth failure

**Then** stderr contains the sentinel string `HYBRID_AUTH_FAILED`
**And** stderr contains the hint `verify the key via 'moai hybrid status glm'`
**And** if any error message in stderr includes the API key, it is **masked** using the `maskAPIKey` pattern (e.g., `sk-x****wxyz` showing only first 4 + last 4 characters; never the full key)
**And** the API key is NOT logged in any stderr line, log file, or telemetry payload (REQ-UPC-022 telemetry from SPEC-V3R3-UPDATE-CLEANUP-001 explicitly excludes API keys)

**Edge cases**:
- HTTP 403 (forbidden, key valid but insufficient permissions): same `HYBRID_AUTH_FAILED` path; hint mentions checking subscription tier.
- API key length < 8 chars: `maskAPIKey` returns `****` per existing implementation.
- Multi-line error message: every line containing the key substring is sanitized.

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridAuthFailureMasksKey`

---

### AC-HYBRID-17: `moai hybrid glm` matches `moai cg` v2.x behavior bit-for-bit (REQ-HYBRID-005, 014)

**Given** a user previously used `moai cg` in v2.x with a known-good workflow
**And** the user has captured the post-launch state: `tmux show-environment`, `.claude/settings.local.json`, `.moai/config/sections/llm.yaml`

**When** the same user upgrades to v3R3 and runs `moai hybrid glm` with the same preconditions (same API key, same profile, same tmux session)

**Then** the resulting state matches v2.x bit-for-bit on these surfaces:
- `tmux show-environment | grep ANTHROPIC_*` returns identical key set + values (modulo `team_mode: "hybrid"` vs `"cg"` in llm.yaml — semantic mapping).
- `.claude/settings.local.json teammateMode == "tmux"` (unchanged).
- `.claude/settings.local.json env` contains the same GLM env-var keys + values (unchanged).
- Claude Code launches with the same provider routing (lead = Claude, teammates = GLM).
- `MOAI_BACKUP_AUTH_TOKEN_GLM` slot used (provider-namespaced) in place of legacy `MOAI_BACKUP_AUTH_TOKEN`.

**And** existing v2.x `moai cg` workflow (e.g., `/moai run SPEC-XXX --team`) produces identical observable behavior.

**Edge cases**:
- Single backup slot (legacy `MOAI_BACKUP_AUTH_TOKEN`) → namespaced slot (`MOAI_BACKUP_AUTH_TOKEN_GLM`): migration path documented; legacy slot read for one transition cycle then cleaned.
- v2.x captured `team_mode: "cg"` config → v3R3 `team_mode: "hybrid"` + `provider: "glm"`: REQ-HYBRID-011 auto-migration handles.

**Test Anchor**: `internal/cli/hybrid_test.go TestHybridBackwardCompatWithCG`

---

### AC-HYBRID-18: Provider registry contains exactly 4 entries at runtime (REQ-HYBRID-002, 017)

**Given** the v3R3 binary is built and `internal/llm/providers/registry.go` is loaded

**When** `providers.List()` is called

**Then** the returned slice contains exactly 4 entries: `["glm", "kimi", "deepseek", "qwen"]`
**And** the order is deterministic (alphabetical or registration order, defined as part of the test contract)

**And When** `providers.Lookup("glm")` is called

**Then** a non-nil Provider implementation is returned

**And When** `providers.Lookup("mistral")` (any non-allow-list value) is called

**Then** an error is returned with sentinel `UNKNOWN_PROVIDER` and message listing the four allowed values

**Edge cases**:
- `providers.Lookup("")`: distinct error (empty string), distinct from UNKNOWN_PROVIDER.
- `providers.Lookup("GLM")` (uppercase): UNKNOWN_PROVIDER (case-sensitive).
- Goroutine concurrency: registry is read-only after init; no race condition.

**Test Anchor**: `internal/llm/registry_test.go TestProviderRegistryConsistency`

---

## 2. Edge Cases Summary

| Scenario | Expected | Test Anchor |
|---|---|---|
| `moai hybrid` (no provider arg) | cobra usage hint | `TestHybridHelpListsFourProviders` |
| `moai hybrid GLM` (uppercase) | UNKNOWN_PROVIDER (case-sensitive) | `TestHybridUnknownProviderRejected` |
| `moai cg --help` | BC notice in help text (NOT runCGRemoved exec) | `TestCGCommandHelp` |
| `moai cg -p work` | BC error; profile flag ignored | `TestCGCommandReturnsBCError` |
| `team_mode: "cg"` + `provider: "kimi"` (manual) | warning, no auto-rewrite | `TestMigrateCGTeamModeUserChoice` |
| Empty `~/.moai/.env.glm` | HYBRID_MISSING_API_KEY | `TestHybridMissingAPIKey` |
| `--proxy` invalid URL | clear pre-launch error | `TestHybridProxyInvalidURL` |
| HTTP 429 from provider | propagate as-is, no auto-retry | `TestHybridRateLimitPropagated` |
| New `internal/llm/providers/mistral.go` PR | PROVIDER_ALLOWLIST_VIOLATION | `TestProviderAllowlist` |
| `MOAI_TEST_MODE=1` bypass tmux check | succeed (test only) | `TestHybridTmuxBypassInTestMode` |
| Provider releases new model (GLM-4.8) | user override path stable | (manual; SPEC update process) |
| `moai cg` invocation cobra fallback (after binary upgrade where cgCmd not registered) | regression sentinel: cgCmd MUST be registered | `TestCGCommandIsRegistered` |
| `moai hybrid glm tools enable` (pre-GLM-MCP-001) | stub message, references SPEC-GLM-MCP-001 | `TestHybridToolsSubcommandStub` |
| `moai hybrid kimi tools enable` (no Z.AI MCP) | stub explaining GLM-specific scope | `TestHybridToolsKimiNoOp` |

---

## 3. Quality Gate Criteria

Per `.claude/rules/moai/core/moai-constitution.md` TRUST 5 framework + plan-auditor PASS criteria:

### 3.1 Tested

- [ ] All 18 ACs have at least one test anchor (verified in plan.md §1.3 Deliverables).
- [ ] Test files compile and pass on macOS, Ubuntu, Windows (CI 3-OS matrix).
- [ ] Code coverage ≥ 85% for `internal/llm/` package (TRUST 5 minimum).
- [ ] Code coverage ≥ 90% for `internal/cli/hybrid.go` (CLI surface = critical).
- [ ] All existing `internal/cli/glm_test.go` and `internal/cli/launcher_test.go` tests continue to pass (regression sentinel for REQ-HYBRID-006).

### 3.2 Readable

- [ ] All exported functions in `internal/llm/` and `internal/cli/hybrid.go` have godoc comments.
- [ ] Provider interface (`internal/llm/provider.go`) documents the contract: Anthropic-compat HTTP only, no native SDKs.
- [ ] Sentinel strings (`MOAI_CG_REMOVED`, `UNKNOWN_PROVIDER`, `HYBRID_REQUIRES_TMUX`, etc.) are extracted to a central `errors.go` or constants file (per CLAUDE.local.md §14 hardcoding prohibition).
- [ ] No magic numbers / strings beyond pinned baselines (which live in provider metadata files).

### 3.3 Unified

- [ ] `gofmt`, `goimports`, `golangci-lint` pass with zero warnings.
- [ ] Naming conventions match existing codebase (e.g., `snake_case.go` for Go files; `CamelCase` for exports).
- [ ] Error wrapping uses `fmt.Errorf("...: %w", err)` (per CLAUDE.local.md §3 Code Standards).

### 3.4 Secured

- [ ] API keys are NEVER logged in plain text (REQ-HYBRID-018).
- [ ] `maskAPIKey` is used at every error path that mentions a key.
- [ ] `~/.moai/.env.<provider>` files are mode `0600` (existing pattern).
- [ ] No new secrets are committed to the repository (verified by CI secret-scanner).
- [ ] BC stub `moai cg` does NOT exfiltrate any pre-existing config (only emits structured error).

### 3.5 Trackable

- [ ] CHANGELOG entry references `BC-V3R3-HYBRID-001` and SPEC-V3R3-HYBRID-001 §10.
- [ ] Conventional Commits used: `feat(hybrid):`, `fix(cg):`, `refactor(llm):`, `docs(spec):` etc.
- [ ] All MX tags (10 planned per plan.md §6) inserted at run-phase end.
- [ ] PR description references SPEC ID + AC list + plan-auditor PASS verdict.

---

## 4. Definition of Done (DoD)

A milestone (M1-M5) is **DONE** when:

1. All test anchors for the milestone's REQ coverage are GREEN.
2. `go test ./...` passes from repo root with 0 failures (no cascading from M1's RED tests after M4 GREEN).
3. `make build` succeeds with 0 errors (embedded.go regenerated).
4. Embedded-template parity verified (all `.claude/...` and `.moai/...` changes mirrored in `internal/template/templates/`).
5. plan-auditor PASS at C1-C18 (final M5 check before sync-phase).

The full SPEC is **DONE** when:

1. All 5 milestones (M1-M5) are DONE.
2. `progress.md` records `run_complete_at: <ISO-8601>` and `run_status: implementation-complete`.
3. `/moai sync SPEC-V3R3-HYBRID-001` succeeds (PR creation + docs-site 4-locale sync per CLAUDE.local.md §17).
4. `BC-V3R3-HYBRID-001` is announced in the v3R3 first minor release notes.

---

End of acceptance.md.
