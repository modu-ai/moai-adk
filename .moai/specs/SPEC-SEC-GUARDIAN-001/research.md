# SPEC-SEC-GUARDIAN-001 — Research (codebase analysis)

Read-only reconnaissance of the surfaces this SPEC extends. All findings verified 2026-07-24 against the current tree (branch `feat/security-absorb-deepscan-001`).

## §A The MoAI shell-hook system (the substrate)

- **31 wrapper scripts** `.claude/hooks/moai/handle-*.sh`, each a thin forwarder to `moai hook <event>`. Every wrapper carries the identical **3-tier `moai` binary resolution chain** + silent `exit 0` fail-open (`command -v moai` → `$HOME/go/bin/moai` → `$HOME/.local/bin/moai` → `exit 0`). Verified: `hook-independence.md` §5 + direct read of `handle-post-tool.sh` / `handle-pre-tool.sh`.
- **Template mirror**: every wrapper has a `.sh.tmpl` twin under `internal/template/templates/.claude/hooks/moai/`. Template-First byte-lockstep is the distribution model.
- **`moai hook <event>` dispatch** in `internal/cli/hook.go`: registers `pre-tool` / `post-tool` / `stop` (event constants) via a loop (lines 47-86), plus bespoke subcommands `list`, `agent [action]`, `harness-observe*`, `db-schema-sync`, `spec-status`, `harness-classify` (lines 90-160). **Adding a new `moai hook security-*` subcommand follows this exact precedent** — no new plumbing invented.
- **Go handler package layout**: `internal/hook/` has subpackages (`dbsync`, `memo`, `mx`). A new `internal/hook/security/` package fits the established structure.

## §B settings.json hook wiring (the registration surface)

From `internal/template/templates/.claude/settings.json.tmpl` (verified lines 45-185):

- **PreToolUse** (lines 51-66): matcher `Write|Edit|Bash`, one entry `handle-pre-tool.sh` (5s). NOTE: `handle-pre-tool.sh` already contains an **inline shell regex heuristic** — the Bash Risk-Amplifier warn signal (WARN-only, fail-open, reads the payload, counts shell metacharacters). This is a direct precedent that a fast regex security check CAN live in-session and be advisory/fail-open. This SPEC moves the heavier pattern logic into Go for testability + single-source, but the precedent proves the shape is sanctioned.
- **PostToolUse** (lines 67-106): matcher `Write|Edit|MultiEdit`, two entries — `handle-post-tool.sh` (async, 10s, LSP/AST/MX) + `status-transition-ownership.sh` (5s, advisory SPEC-body governance). A third `HookOptIn.Enabled`-gated harness-observe entry. **Layer 1 adds a sibling entry here** (async, 5s).
- **Stop** (lines 107-149): three entries — `handle-stop.sh` (5s) + `sync-phase-quality-gate.sh` (60s) + `handle-stop-goal.sh` (120s) + an opt-in harness-observe entry. **Layer 2 adds a sibling Stop entry**; **Layer 3 (L3-A) adds a sibling Stop entry** (dormant self-gating).
- Windows/non-windows command rendering is templated (`{{- if eq .Platform "windows"}}`).
- **`settings_test.go:requiredKeys`** enforces required env/hook keys; adding entries must not remove them.

## §C The two integration-precedent gate hooks

### `sync-phase-quality-gate.sh` (Stop gate — the Layer-3 model source)
- Fires on the Stop event, self-gates to a commit-subject match (`docs(...): sync-phase` / `chore(...): sync`), then runs a **language-neutral fast structural check** (a `detect_language` marker walk across all 16 languages, `code_delta_pattern` per language). Directly reusable model for "detect a commit landed, run a per-language check."
- **once-per-commit sentinel** (`.moai/state/sync-quality-gate.last`) prevents re-firing every turn-end while HEAD is unchanged — a pattern Layer 2/3 reuse.
- **`MOAI_SYNC_GATE_BLOCKING`** opt-out (blocking is the DEFAULT for vet/build there, but the ENV-gating machine is the model): advisory `{"systemMessage"}` vs blocking `{"hookSpecificOutput":{...,"decision":"block"}}` on exit-0 stdout. The guardian inverts the default (advisory-first) but reuses the emit machinery verbatim.
- **exit 0 always** — the blocking decision rides stdout JSON on exit 0 (exit 2 discards stdout). This is the canonical hook-block contract the guardian must follow (REQ-SG-061).

### `status-transition-ownership.sh` (PostToolUse gate)
- Fires on PostToolUse Write/Edit/MultiEdit of SPEC-artifact files, `jq`-degrades-gracefully (no-op if `jq` absent), `--skip-hook` first-arg override → `.moai/logs/hook-skip.log`, `set -e`, `${CLAUDE_PROJECT_DIR:-$PWD}` root. Advisory (exit 0 always). Model for the guardian's PostToolUse advisory posture + `--skip-hook` audit convention (REQ-SG-043).

## §D Hook boundary + orchestrator translation (the HARD constraint)

- `agent-common-protocol.md` § Hook Invocation Surface: hooks emit exit-code + structured JSON; they **MUST NOT** call AskUserQuestion; the orchestrator parses the JSON and runs the AskUserQuestion round (accept / `--skip-hook` override / abort). Canonical grep: `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ | grep -v "^[^:]*:[0-9]*:[ \t]*#"` → expected no matches. This binds every guardian hook (REQ-SG-040, -041).
- `hooks-system.md`: PostToolUse cannot block (async delivers only `additionalContext`); Stop can block via `hookSpecificOutput.decision` on exit 0; PreToolUse can block (exit 2 / permissionDecision). `type: "agent"` hooks have Read/Grep/Glob and return `{ok, reason}` — the native mechanism behind the Layer-3 L3-B alternative (documented but NOT chosen — L3-A is the settled surface).
- Recovery-Signal Carve-Out (`runtime-recovery-doctrine.md` §4): documentation-only at this layer (current hooks don't parse `stopReason`); the guardian inherits the same posture.

## §E Pattern vocabulary sources (reference skills — NOT shipped)

- `moai-ref-owasp-checklist` — injection / XSS / CSRF / auth / input-validation baseline.
- `moai-ref-llm-security` — ML-specific dangerous idioms (`torch.load(weights_only=False)`, unsafe deserialization of model artifacts, prompt-injection-adjacent output-validation gaps).
- `moai-ref-secops` — operational / API defenses (some overlap; the guardian focuses on edit-time source idioms, not runtime ops).
- `moai-ref-supply-chain` — dependency-confusion / typosquatting (mostly out of the per-file edit-time scope; noted for completeness).

These provide the vulnerability-class vocabulary the pattern table (design.md §P) is drawn from. They are consulted at authoring time; nothing from them is copied into shipped template content.

## §F Catalog / neutrality constraints (guards to respect)

- `catalog_tier_audit_test.go`: `expectedAgentCount=10`, `expectedSkillCount=28`; `catalog_loader_test.go`: `expectedTotal=38`. This SPEC adds NO agent and NO skill → these counts stay unchanged (a hook + Go handler is not catalog-counted).
- `internal_content_leak_test.go` + `template-neutrality-check.yaml`: shipped template content (the 3 wrappers + settings.json wiring) MUST NOT embed `SPEC-SEC-GUARDIAN-001`, internal dates, or SHAs, and MUST NOT elevate a single language. Go source under `internal/` is NOT template content and is exempt (patterns.go MAY carry rich comments).
- `internal/template/CLAUDE.md`: `templates/.claude/agents/harness/` must not exist; the guardian touches only `templates/.claude/hooks/moai/` + `settings.json.tmpl`.

## §G The SPEC-1 boundary (why there is no overlap)

SPEC-SEC-DEEPSCAN-001 (read: its spec.md + design.md) is a **markdown-only** playbook change to `/moai review --deep`: an on-demand, explicitly-invoked, six-phase multi-agent scan with a 3-voter adversarial panel, reviewer-vouched patches, and a timestamped `.moai/reports/security-deepscan-*/` results directory. Its §C Out of Scope EXPLICITLY defers "the lighter always-on in-session guardian" to "Epic SECURITY-ABSORB SPEC-2" — i.e. THIS SPEC. There is zero surface overlap: SPEC-1 touches `review.md` (workflow skill) + `review.md.tmpl` (command); SPEC-2 touches `internal/hook/security/` + hook wrappers + settings.json. Different files, different triggers (explicit `/moai review --deep` vs always-on PostToolUse/Stop/commit), different depth (multi-agent adversarial vs single-pass regex). The layering is complementary by construction.

## §H Settled design forks (resolved by orchestrator)

1. **Layer-3 surface** — SETTLED: L3-A (extend the sync-gate model via a sibling commit-time Stop hook + orchestrator escalation). The native `type: agent` PreToolUse git-commit variant is documented but NOT chosen (boundary-cleanliness + speed + gate-hook consistency). Commit-aware, dormant/opt-in by default.
2. **Escalation delivery** — SETTLED: orchestrator-mediated `Agent()`. Hooks emit a structured signal; the orchestrator spawns the review. Hooks never call a `type: agent`/`prompt` sub-model directly (keeps the deterministic hook fast + preserves the hook→orchestrator-translation boundary).
