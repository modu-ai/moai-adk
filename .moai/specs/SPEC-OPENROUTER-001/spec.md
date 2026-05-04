---
id: SPEC-OPENROUTER-001
version: 0.1.0
status: Draft
created: 2026-05-04
updated: 2026-05-04
author: claude (issue triage)
priority: P2
issue_number: 780
harness_level: standard
language_policy: 16-language-neutral
template_first: true
related_specs:
  - SPEC-GLM-001
  - SPEC-CI-MULTI-LLM-001
---

# SPEC-OPENROUTER-001: OpenRouter Provider Backend Integration

## HISTORY

| Version | Date       | Author                | Description                                                                          |
|---------|------------|-----------------------|--------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | claude (issue triage) | Initial draft from issue #780 — feature request for OpenRouter API support           |

---

## Metadata

| Field          | Value                                                                                  |
|----------------|----------------------------------------------------------------------------------------|
| SPEC ID        | SPEC-OPENROUTER-001                                                                    |
| Title          | OpenRouter Provider Backend Integration                                                |
| Status         | Draft (pending design review)                                                          |
| Priority       | P2 (user-requested feature, no blocker)                                                |
| Harness Level  | standard                                                                               |
| Source Issue   | #780 — "[Feature] Do you support the OpenRouter Api?"                                  |
| Owner (impl)   | TBD (likely expert-backend)                                                            |
| Owner (design) | manager-spec (post-triage)                                                             |
| Related Files  | `internal/cli/glm.go`, `internal/config/defaults.go`, `internal/config/envkeys.go`, `internal/template/templates/.moai/config/sections/llm.yaml` |

---

## 1. Background and Motivation

### 1.1 User Request (Issue #780)

> "I already have an OpenRouter account that can use the GLM model. I don't want to need to create another z.AI account. OpenRouter has access to many models."

The user wants a single account/billing surface (OpenRouter) instead of provisioning separate provider accounts (Z.AI for GLM, Anthropic Console for Claude, etc.). OpenRouter aggregates **300+ models** behind a single API and a unified billing relationship.

### 1.2 Why This Fits MoAI-ADK Today

MoAI-ADK already supports a non-Anthropic backend via the **GLM mode** (`moai glm` / `moai cg`). The architectural pattern is:

1. User authenticates Claude Code against an **Anthropic-compatible endpoint** (i.e., `/v1/messages` API surface) by overriding two env vars:
   - `ANTHROPIC_BASE_URL` — the proxy/provider base URL
   - `ANTHROPIC_AUTH_TOKEN` — the provider's auth token
2. Compatibility env vars are added to neutralize Anthropic-specific beta headers that the third-party endpoint does not implement:
   - `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`
   - `DISABLE_PROMPT_CACHING=1`
3. Model name mapping (`claude-opus-4-7` → `glm-5.1`, etc.) is handled by Claude Code via env vars or by the provider's own model alias resolution.

OpenRouter exposes an **Anthropic Messages API-compatible endpoint** at `https://openrouter.ai/api/v1` (verify before implementation). This means the existing GLM integration scaffolding can be reused with new defaults and a separate setup command.

### 1.3 OpenRouter Key Characteristics

| Attribute           | Value                                                              |
|---------------------|--------------------------------------------------------------------|
| Auth scheme         | Bearer token (`Authorization: Bearer sk-or-v1-...`)                |
| Base URL            | `https://openrouter.ai/api/v1` (OpenAI-format) and Anthropic-compat surface |
| Model namespacing   | `<vendor>/<model>` (e.g., `anthropic/claude-opus-4`, `z-ai/glm-4.6`, `openai/gpt-5`, `google/gemini-2.5-pro`) |
| Special headers     | `HTTP-Referer`, `X-Title` (recommended for OpenRouter analytics; optional) |
| Pricing             | Per-token, billed against OpenRouter credit balance                |
| Rate limits         | Per-account, varies by model (some Anthropic models throttle hard) |
| Privacy policy      | Configurable per-request (`X-OR-Privacy: pass-through` etc.)       |

---

## 2. Goals and Non-Goals

### 2.1 Goals

- **G1**: Allow users to point Claude Code at OpenRouter as the backend provider via a new `moai openrouter` command (or `moai or`).
- **G2**: Reuse existing GLM-mode env-injection plumbing (`ANTHROPIC_BASE_URL` / `ANTHROPIC_AUTH_TOKEN` / `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` / `DISABLE_PROMPT_CACHING`) to minimize new code surface area.
- **G3**: Provide a model mapping mechanism so that `high`/`medium`/`low` performance tiers map to user-selected OpenRouter model IDs (default suggestion: `anthropic/claude-opus-4`, `anthropic/claude-sonnet-4-5`, `anthropic/claude-haiku-4-5` — but user-overridable in `llm.yaml`).
- **G4**: Document the trade-offs vs. direct Anthropic / direct Z.AI accounts (latency, cost markup, privacy, BYOK availability).
- **G5**: Add OpenRouter to the multi-LLM CI matrix in `SPEC-CI-MULTI-LLM-001` as a fifth provider option (separate follow-up SPEC).

### 2.2 Non-Goals (Explicit Exclusions)

- **NG1**: This SPEC does NOT add a generic "any OpenAI-compatible endpoint" provider. OpenRouter integration is the immediate scope; future generic providers will need a separate SPEC.
- **NG2**: This SPEC does NOT replace direct GLM mode (`moai glm` remains supported). OpenRouter is an additional option, not a replacement.
- **NG3**: This SPEC does NOT integrate OpenRouter in Agent Teams parallel mode (`moai cg` analog) in v1. CG-style hybrid mode for OpenRouter is a v2 follow-up if user demand exists.
- **NG4**: This SPEC does NOT bundle OpenRouter credit purchase or wallet management — users manage credits via openrouter.ai.

---

## 3. Requirements (EARS Format)

### REQ-OR-001 (Ubiquitous): `moai openrouter` command group

[Ubiquitous] The system SHALL expose an `moai openrouter` cobra subcommand group with the following children:

- `moai openrouter setup [api-key]` — store API key in `~/.moai/.env.openrouter`
- `moai openrouter status` — show stored credential state (masked) and active model mapping
- `moai openrouter` — launch Claude Code with OpenRouter backend env vars injected

Acceptance:
- AC-001.1: `moai openrouter --help` enumerates all three subcommands and global flags (`-p`, `--permission-mode`, `-b`).
- AC-001.2: Subcommand structure mirrors `moai glm` for user familiarity.

### REQ-OR-002 (Event-Driven): `moai openrouter setup <key>`

**WHEN** the user runs `moai openrouter setup sk-or-v1-...`, **THEN** the system SHALL:

1. Validate the key prefix matches `sk-or-v1-` (OpenRouter convention).
2. Persist the key to `~/.moai/.env.openrouter` with file mode `0600`.
3. Echo a confirmation message including a masked preview (e.g., `sk-o***x9fa`).

Acceptance:
- AC-002.1: Invalid prefix returns a non-zero exit code with the message `invalid OpenRouter key format (expected sk-or-v1-...)`.
- AC-002.2: File `~/.moai/.env.openrouter` exists with mode `0600` after success.
- AC-002.3: Stored key is never echoed in full to stdout, stderr, or any log file.

### REQ-OR-003 (Event-Driven): `moai openrouter` launch

**WHEN** the user runs `moai openrouter [-- claude-args...]`, **THEN** the system SHALL:

1. Load API key from `~/.moai/.env.openrouter`. If missing, exit with `no OpenRouter credentials configured; run 'moai openrouter setup <key>' first`.
2. Set the following env vars in the spawned Claude Code process:
   - `ANTHROPIC_BASE_URL=https://openrouter.ai/api/v1` (or the Anthropic-compatible subpath if required by OpenRouter — to be confirmed during implementation against OpenRouter docs)
   - `ANTHROPIC_AUTH_TOKEN=<key>`
   - `ANTHROPIC_MODEL=<resolved-from-llm.yaml>` (per performance_tier)
   - `ANTHROPIC_SMALL_FAST_MODEL=<resolved-from-llm.yaml>` (low tier)
   - `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`
   - `DISABLE_PROMPT_CACHING=1`
3. Optionally inject OpenRouter-specific headers via `ANTHROPIC_CUSTOM_HEADERS` (if Claude Code supports this env var) for analytics: `HTTP-Referer: https://moai.ai.kr`, `X-Title: MoAI-ADK`.
4. Exec `claude` (replace current process) with the same `--` passthrough arg semantics as `moai glm`.

Acceptance:
- AC-003.1: Spawned `claude` process has the four required env vars set (verifiable via `claude --print 'echo $ANTHROPIC_BASE_URL'` if such introspection is possible, otherwise via integration test using a fake exec).
- AC-003.2: Auto mode is rejected with the same message as GLM mode (`Auto mode is not available with OpenRouter (third-party provider)`).

### REQ-OR-004 (State-Driven): `moai cc` clears OpenRouter env

**WHILE** the user runs `moai cc` (Claude direct mode), **THEN** the system SHALL remove OpenRouter env vars from `settings.local.json` and tmux session env (analogous to `removeGLMEnv`).

Acceptance:
- AC-004.1: After `moai cc`, `settings.local.json` `env` block does NOT contain `ANTHROPIC_BASE_URL` set to OpenRouter URL nor `ANTHROPIC_AUTH_TOKEN` set to the OpenRouter key.
- AC-004.2: tmux session env is cleared of the same vars.

### REQ-OR-005 (Ubiquitous): Model mapping in `llm.yaml`

[Ubiquitous] The system SHALL extend `internal/template/templates/.moai/config/sections/llm.yaml` with an `openrouter:` block:

```yaml
llm:
  openrouter:
    base_url: "https://openrouter.ai/api/v1"
    # Default model mapping (user-overridable)
    models:
      high: "anthropic/claude-opus-4"
      medium: "anthropic/claude-sonnet-4-5"
      low: "anthropic/claude-haiku-4-5"
    # Optional: alternative cost-optimized presets
    # presets:
    #   glm-only:
    #     high: "z-ai/glm-4.6"
    #     medium: "z-ai/glm-4.5"
    #     low: "z-ai/glm-4.5-air"
    headers:
      http_referer: "https://moai.ai.kr"
      x_title: "MoAI-ADK"
```

Acceptance:
- AC-005.1: `internal/config/defaults.go` defines `DefaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"`.
- AC-005.2: `internal/config/envkeys.go` defines no new env constant (reuses existing `EnvAnthropicBaseURL`, `EnvAnthropicAuthToken`).
- AC-005.3: Template-First rule respected: change made in `internal/template/templates/.moai/config/sections/llm.yaml` first, then `make build`.

### REQ-OR-006 (Optional): Custom model presets

[Optional] **WHERE** the user defines a `presets:` block in `llm.yaml` `openrouter:`, **THEN** `moai openrouter setup --preset <name>` SHALL switch the active mapping to the named preset (saved in `~/.moai/.env.openrouter` as `OPENROUTER_PRESET=<name>`).

This allows a user to maintain multiple model mappings (e.g., one Anthropic-only, one GLM-only via OpenRouter, one mixed) and switch between them.

Acceptance:
- AC-006.1: `moai openrouter status` shows the active preset name.
- AC-006.2: `moai openrouter --preset glm-only` overrides the default mapping for that single launch.

### REQ-OR-007 (Unwanted Behavior): No silent model fallback

[Unwanted] **IF** the resolved OpenRouter model ID does not exist in OpenRouter's catalog (HTTP 400 / 404 from OpenRouter), **THEN** the system MUST surface the upstream error to the user verbatim and SHALL NOT silently fall back to a different model.

Rationale: silent fallback to a cheaper model risks both unexpected output quality changes and unexpected cost shifts.

Acceptance:
- AC-007.1: A failed OpenRouter call produces a Claude Code error message that includes the OpenRouter error body (model not found, insufficient credits, rate limited, etc.).
- AC-007.2: No retry-with-different-model logic exists in the OpenRouter wrapper.

### REQ-OR-008 (State-Driven): Statusline context-window override

**WHILE** OpenRouter mode is active and the active model is one of the well-known GLM variants (`z-ai/glm-5.1`, `z-ai/glm-4.6`, `z-ai/glm-4.5-air`), **THEN** the statusline SHALL apply the same `MOAI_STATUSLINE_CONTEXT_SIZE` override pattern documented for GLM mode in `llm.yaml` (issue #653 lineage).

Acceptance:
- AC-008.1: Statusline gauges report correct headroom for the GLM models when accessed via OpenRouter.
- AC-008.2: For non-GLM models routed via OpenRouter, the statusline trusts Claude Code's reported context window without override.

### REQ-OR-009 (Ubiquitous): Documentation and discoverability

[Ubiquitous] The system SHALL document OpenRouter mode in:

1. `moai openrouter --help` (cobra-generated)
2. `README.md` and `README.ko.md` — add to "Provider Modes" section alongside GLM
3. `docs-site/content/{ko,en,ja,zh}/...` — 4-locale doc page (per `CLAUDE.local.md §17`)

Acceptance:
- AC-009.1: All four locales have the new docs page (per `scripts/docs-i18n-check.sh`).
- AC-009.2: README "Provider Modes" table lists three modes: Claude / GLM / OpenRouter.

### REQ-OR-010 (Ubiquitous): Test coverage

[Ubiquitous] All OpenRouter command code SHALL have table-driven Go tests achieving:

- 90%+ line coverage on `internal/cli/openrouter.go`
- Integration test that spawns `moai openrouter setup → status → launch (mocked exec)` end-to-end
- Settings.local.json env injection / removal tests parallel to existing `glm_test.go` patterns

Acceptance:
- AC-010.1: `go test ./internal/cli/... -run OpenRouter -cover` reports >= 90%.
- AC-010.2: A regression test verifies that `moai cc` after `moai openrouter` restores the user's prior `ANTHROPIC_AUTH_TOKEN` (per existing GLM API key preservation pattern in `glm.go:516`).

---

## 4. Open Questions

These MUST be resolved during the `/moai plan` phase before implementation. They are intentionally open because they require investigation against OpenRouter's live API surface and user preference.

### OQ1 — Anthropic-compatible base URL path

OpenRouter's documentation page describes both an OpenAI-compatible endpoint (`/api/v1/chat/completions`) and an Anthropic Messages-compatible endpoint. **Action**: Verify the exact base URL path Claude Code expects so that `ANTHROPIC_BASE_URL` resolves the `/v1/messages` route correctly. If the path differs from `https://openrouter.ai/api/v1`, adjust `DefaultOpenRouterBaseURL` accordingly.

Source to consult: <https://openrouter.ai/docs> (verify before implementation).

### OQ2 — Default model mapping policy

Should the default `models.high/medium/low` mapping ship as **Anthropic models** (matching Claude Code's native behavior) or as **mixed/cheaper models** (matching the user's apparent intent in #780, where they mention GLM via OpenRouter)? Recommendation: ship Anthropic defaults (least-surprise) and provide a `glm-only` preset in the comment example for users who want cheaper routing.

### OQ3 — Provider routing preferences

OpenRouter supports per-request `provider` routing (e.g., prefer Cerebras for speed, prefer DeepInfra for cost). Should `llm.yaml` expose this as a `routing:` field (e.g., `routing.preferred_providers: [cerebras, anthropic]`)? Recommendation: defer to v2; v1 uses OpenRouter's default routing.

### OQ4 — BYOK (Bring-Your-Own-Key) integration

OpenRouter supports BYOK where users plug their own provider keys (e.g., own Anthropic key) into OpenRouter and pay reduced markup. Is BYOK in scope? Recommendation: out of scope for v1 (purely a user-side OpenRouter dashboard configuration); document only.

### OQ5 — Cost visibility in statusline

OpenRouter exposes cost-per-request via response headers (`x-ratelimit-credit`, `x-cost`, etc.). Should the statusline show real-time spend? Recommendation: defer to v2 (would require statusline plumbing extensions); v1 directs users to <https://openrouter.ai/activity>.

---

## 5. Risks

| ID  | Risk                                                                                           | Severity | Probability | Mitigation                                                          |
|-----|------------------------------------------------------------------------------------------------|----------|-------------|---------------------------------------------------------------------|
| R1  | OpenRouter Anthropic-compatible endpoint does not implement all `/v1/messages` features Claude Code uses (e.g., extended thinking, file uploads) | P1       | Medium      | OQ1 verification + integration test matrix; document feature-parity gaps |
| R2  | Cost markup vs. direct Anthropic surprises users                                               | P2       | High        | Document OpenRouter's ~5% markup explicitly in setup output         |
| R3  | OpenRouter outage takes down user's MoAI workflow                                              | P2       | Low         | `moai cc` fallback documented; users can switch back instantly      |
| R4  | Rate limits from upstream provider (Anthropic, OpenAI) propagate as opaque OpenRouter errors   | P2       | Medium      | REQ-OR-007 surfaces verbatim error; document common error codes     |
| R5  | OpenRouter privacy default routes prompts through provider that logs them                      | P1       | Low         | Document `X-OR-Privacy` header in advanced docs; do not default-enable |
| R6  | API key leak via process env (visible in `ps -ef`)                                             | P1       | Low         | Same risk as GLM mode; documented pattern is acceptable             |

---

## 6. Implementation Approach (Sketch)

This is a sketch only — full task decomposition belongs in `plan.md` after `/moai plan SPEC-OPENROUTER-001`.

### Phase 1 — Config & Defaults

- Add `DefaultOpenRouterBaseURL` constant in `internal/config/defaults.go`.
- Extend `internal/template/templates/.moai/config/sections/llm.yaml` with `openrouter:` block.
- Run `make build` to regenerate embedded templates.

### Phase 2 — Setup & Storage

- Create `internal/cli/openrouter.go` modeled on `internal/cli/glm.go`.
- Implement `moai openrouter setup` (stores key in `~/.moai/.env.openrouter`, mode 0600).
- Implement `loadOpenRouterKey()` mirror of `loadGLMKey()`.
- Implement `moai openrouter status` (masked display + active model mapping).

### Phase 3 — Launch & Env Injection

- Implement `runOpenRouter` mirror of `runGLM`.
- Reuse compatibility-env injection pattern (`CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`, `DISABLE_PROMPT_CACHING=1`).
- Resolve performance_tier → model ID via `llm.yaml`.

### Phase 4 — Mode-Switching Plumbing

- Extend `removeGLMEnv` (or factor into a generic `removeProviderEnv`) to also clear OpenRouter env vars on `moai cc`.
- Update `injectTmuxSessionEnv` and related helpers if a future `moai cor` (Claude + OpenRouter teammates) is added — out of scope for v1.

### Phase 5 — Documentation

- Update `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`.
- Add `docs-site/content/{ko,en,ja,zh}/guides/openrouter.md`.
- Update CHANGELOG.md.

### Phase 6 — Tests

- Table-driven tests in `internal/cli/openrouter_test.go`.
- Integration test that mocks `claude` exec and verifies env vars passed.
- Regression test for `moai cc` token preservation (mirror of `oauth_token_preservation_test.go`).

---

## 7. Acceptance Summary

| Criterion                                                                                               | Verification                                  |
|---------------------------------------------------------------------------------------------------------|-----------------------------------------------|
| OpenRouter setup stores key with `0600` perms                                                           | `internal/cli/openrouter_test.go` filesystem check |
| `moai openrouter` launches Claude Code with correct env vars                                            | Integration test with mocked `exec`           |
| `moai cc` removes OpenRouter env from `settings.local.json` and tmux                                    | Settings round-trip test                      |
| Default models map to `anthropic/*` IDs                                                                 | YAML schema validation                        |
| `moai openrouter --help` documents all subcommands                                                      | CLI smoke test                                |
| Documentation present in 4 locales                                                                      | `scripts/docs-i18n-check.sh`                  |
| 90%+ test coverage on new code                                                                          | `go test -cover`                              |

---

## 8. Dependencies and References

### 8.1 Codebase References

- `internal/cli/glm.go` — pattern source for provider command structure
- `internal/config/defaults.go:41` — `DefaultGLMBaseURL` pattern to mirror
- `internal/config/envkeys.go:87-90` — `EnvAnthropicBaseURL` / `EnvAnthropicAuthToken` constants (reused, not duplicated)
- `internal/template/templates/.moai/config/sections/llm.yaml` — provider config schema
- `.moai/specs/SPEC-GLM-001/spec.md` — sibling provider integration SPEC
- `.moai/specs/SPEC-CI-MULTI-LLM-001/spec.md` — multi-LLM CI integration (potential follow-up to add OpenRouter as 5th LLM)

### 8.2 External References

- OpenRouter docs: <https://openrouter.ai/docs> (verify Anthropic-compatible endpoint path)
- OpenRouter quickstart: <https://openrouter.ai/docs/quickstart>
- OpenRouter model catalog: <https://openrouter.ai/models>
- Issue #780 (this spec's source): <https://github.com/modu-ai/moai-adk/issues/780>

### 8.3 CLAUDE.local.md Compliance

- §2 Template-First Rule — all changes go through `internal/template/templates/` first
- §14 No hardcoded URLs in Go code — `DefaultOpenRouterBaseURL` constant required
- §15 16-language neutrality preserved — provider integration is language-agnostic
- §17 4-locale docs sync — README + docs-site changes must hit ko/en/ja/zh

---

## 9. Triage Notes

This SPEC was drafted automatically by Claude in response to issue #780 to give the maintainer a concrete starting point for design review. It is **Status: Draft** and intentionally leaves open questions (§4) and risk decisions (§5) for human review before any code is written.

Recommended next steps for the maintainer:

1. Review §4 Open Questions and decide on positions (especially OQ1, OQ2).
2. Run `/moai plan SPEC-OPENROUTER-001` to convert this draft into a full plan.md + research.md + tasks.md set.
3. Optionally re-prioritize: P2 reflects "user-requested feature, no blocker" — bump to P1 if OpenRouter unlocks a customer commitment.
