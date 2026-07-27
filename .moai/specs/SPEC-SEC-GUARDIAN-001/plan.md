# SPEC-SEC-GUARDIAN-001 — Implementation Plan

> Milestones are ordered by decision-reversibility: the highest-change-likelihood decisions (pattern-config schema / vuln-class taxonomy, the Layer-3 surface decision, the L1/L2/L3 hook-event contract) lead; mechanical wiring and neutrality passes trail. Human review should focus on M1-M3 and the §I settled design-fork decisions.

## §A Context

### A.1 Work location & distribution model

- **Go source (compiled, tested, NOT template)**: `internal/hook/security/` (new package: pattern table + scanner + per-layer handlers) and `internal/cli/hook.go` (subcommand wiring). This is compiled into the binary via the normal Go build — NOT template content (REQ-SG-052).
- **Template source (author FIRST)**: `internal/template/templates/.claude/hooks/moai/*.sh.tmpl` (3 new shell wrappers) + `internal/template/templates/.claude/settings.json.tmpl` (hook wiring). Template-First per CLAUDE.local.md §2.
- **Local sync (author SECOND, lockstep)**: `.claude/hooks/moai/*.sh` siblings + `.claude/settings.json` rendered copy.
- **Build**: `make build` recompiles the binary AND re-embeds the template tree via `//go:embed all:templates`.
- This SPEC is **Go + template hybrid** (unlike SPEC-1 which was markdown-only): the pattern engine is Go with tests; the wiring is Template-First shell + settings.json.

### A.2 Extension points (verified 2026-07-24)

1. **Hook wrapper pattern** — `.claude/hooks/moai/handle-*.sh` (31 wrappers) all carry the identical 3-tier `moai` binary resolution chain + silent `exit 0` fail-open (`hook-independence.md` §5). New guardian wrappers reuse this exact shape.
2. **`moai hook <event>` dispatch** — `internal/cli/hook.go` registers subcommands (`pre-tool`, `post-tool`, `stop`, plus specialized `db-schema-sync`, `spec-status`, `harness-classify`). New guardian subcommands register the same way (precedent verified: lines 47-160).
3. **PostToolUse registration** — `settings.json.tmpl` lines 67-106: matcher `Write|Edit|MultiEdit`, carries `handle-post-tool.sh` (async, 10s) + `status-transition-ownership.sh` (5s). Layer 1 adds a sibling entry here (async, 5s).
4. **Stop registration** — `settings.json.tmpl` lines 107-149: carries `handle-stop.sh` (5s) + `sync-phase-quality-gate.sh` (60s) + `handle-stop-goal.sh` (120s). Layer 2 adds a sibling Stop entry.
5. **PreToolUse registration** — `settings.json.tmpl` lines 51-66: matcher `Write|Edit|Bash`, carries `handle-pre-tool.sh` (which ALREADY has an inline regex heuristic — the Bash Risk-Amplifier warn signal — a precedent for a fast in-wrapper security scan). NOTE: the REJECTED L3-B alternative (a `type: agent` PreToolUse `Bash(git commit *)` hook) would have registered HERE; the settled L3-A surface uses a Stop-hook sibling (item 4) instead, so NO PreToolUse entry is added by this SPEC.
6. **Opt-in blocking precedent** — `sync-phase-quality-gate.sh` uses `MOAI_SYNC_GATE_BLOCKING` (advisory opt-out); the guardian mirrors this with `MOAI_SECURITY_*` env flags.
7. **Reference-skill vocabulary** — `moai-ref-owasp-checklist`, `moai-ref-llm-security`, `moai-ref-secops`, `moai-ref-supply-chain` (source of the vuln-class taxonomy; consulted at authoring time, NOT shipped).

### A.3 PRESERVE list (do NOT modify)

- All existing hook wrappers (`handle-*.sh`) and their behavior — new wrappers are ADDED, none rewritten.
- `sync-phase-quality-gate.sh` and `status-transition-ownership.sh` bodies — Layer 3 (L3-A, settled) EXTENDS the sync-gate MODEL via a SEPARATE sibling commit-time Stop hook (`security-commit-review`); it does NOT rewrite the `sync-phase-quality-gate.sh` script. The extension is model-reuse (commit-detection + language-neutral gate shape), not an in-place body edit.
- `handle-pre-tool.sh` Bash Risk-Amplifier block — preserved verbatim (Layer 3, if PreToolUse-realized, is a SEPARATE hook entry, not an edit to this block).
- `catalog.yaml` agent/skill counts (`expectedAgentCount=10`, `expectedSkillCount=28`, `expectedTotal=38`) — no new agent, no new skill dir (REQ-SG-002 non-goal).
- All existing Go handlers under `internal/hook/` — the new `internal/hook/security/` package is additive.
- `.claude/workflows/*.js` (user-owned) — no shipped script.
- Runtime-managed files (`.moai/state/*`, `.moai/reports/*`, `.moai/cache/*`, `.moai/logs/*` contents).

## §B Known Issues (auto-injected, filtered for a Go + template SPEC)

- **B1 — Cross-platform build tags**: the pattern engine is pure-Go regex (no syscall) so no build tags needed; verify `GOOS=windows GOARCH=amd64 go build ./...` passes anyway.
- **B3 / B11 — Subagent boundary (C-HRA-008)**: `internal/hook/security/` MUST NOT contain `AskUserQuestion`/`mcp__askuser`. Verification: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/security/ | grep -v _test.go` → 0. Same grep over `.claude/hooks/moai/` (excluding comments) → 0 (REQ-SG-040).
- **B4 — Frontmatter canonical schema**: spec.md uses `created:`/`updated:`/`tags:` (done). Snake_case aliases prohibited.
- **B6 — spec-lint heading convention**: spec.md §C uses `### Out of Scope — <topic>` h3 sub-headings with `-` bullets (done — 7 sub-headings + literal "out of scope").
- **B7 — hook cwd/path resolution**: guardian handlers reading the working tree MUST resolve project root via `$CLAUDE_PROJECT_DIR` (fallback `$PWD`), never `os.Getwd()` leaking into `internal/hook/.moai/` (the observer.go anomaly). Git diff commands run from the resolved project root.
- **B8 / B10 — working-tree hygiene & scope discipline**: run-phase commits touch ONLY the new `internal/hook/security/` package, `internal/cli/hook.go`, the 3 new wrappers (template + local), and settings.json (template + local). `git add` specific paths only; never `-A`. Do NOT touch runtime-managed files.
- **B9 — Tier L → Route B (PR)**: Tier L routes through a PR (`spec-workflow.md` Route B). `manager-git` opens the PR; run-phase manager-develop commits on the feature branch.
- **Template neutrality (§25 / §15)**: shell wrappers + settings.json wiring MUST NOT embed `SPEC-SEC-GUARDIAN-001`, internal dates, or SHAs, and MUST NOT privilege Go. Pattern-config comments live in Go (`internal/`) which is exempt from template neutrality (it is not template content). CI guard: `internal_content_leak_test.go` + `template-neutrality-check.yaml`.
- **settings.json.tmpl requiredKeys test**: `settings_test.go:requiredKeys` enforces the template env/hook keys — adding hook entries must not remove existing required keys; verify `go test ./internal/template/... ./internal/cli/...` after wiring.

## §C Pre-flight (run before any edit)

```bash
# 1. Branch + baseline
git branch --show-current && git rev-parse HEAD
# 2. Confirm the hook wrapper 3-tier resolution shape (copy target)
grep -c 'command -v moai' .claude/hooks/moai/handle-post-tool.sh
# 3. Baseline: current PostToolUse/Stop hook registrations
grep -n 'PostToolUse\|Stop\|handle-post-tool\|sync-phase-quality-gate' internal/template/templates/.claude/settings.json.tmpl | head
# 4. Baseline lint + build
golangci-lint run --timeout=2m 2>&1 | tail -5
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
# 5. Boundary baseline (must stay 0 after edits)
grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ | grep -v "^[^:]*:[0-9]*:[ \t]*#" || echo "boundary clean"
# 6. Catalog counts baseline (must be unchanged)
grep -n 'expectedAgentCount\|expectedSkillCount\|expectedTotal' internal/template/catalog_tier_audit_test.go internal/template/catalog_loader_test.go
```

## §D Constraints (DO NOT VIOLATE)

- REQ-SG-002: no overlap with SPEC-1 deep scan; no new agent; no new subcommand `/moai security`; no `.js`.
- REQ-SG-040: no `AskUserQuestion`/`mcp__askuser` in any hook script or in `internal/hook/security/`.
- REQ-SG-042: advisory-by-default; blocking ONLY behind an explicit opt-in env flag. Layer 3 dormant (no-op) unless enabled.
- REQ-SG-050 / REQ-SG-051: Template-First byte-lockstep for wrappers + settings.json; 16-language neutrality on all template content; no internal SPEC-ID/date/SHA leak.
- REQ-SG-052: Go code with tests (`t.TempDir`, no OTEL env in parallel tests); NOT template content.
- REQ-SG-053: single-source pattern table (Go), not scattered / not duplicated per language.
- REQ-SG-060: fail-open (silent no-op) on missing `moai`/`jq`/`git`.
- Never `--no-verify`, never force-push, `git add` specific paths only. Both-tree lockstep per edit.

## §E Self-Verification (run-phase completion evidence targets)

The run-phase completion report (manager-develop §E) MUST carry, per the 5-section evidence format:

- **E1 — AC matrix**: every AC in acceptance.md PASS/FAIL with the verbatim `go test` / `grep` / `diff` command + output.
- **E2 — cross-platform build**: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- **E3 — coverage**: `go test -cover ./internal/hook/security/...` ≥ 85% per package.
- **E4 — boundary grep**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/security/ .claude/hooks/moai/ | grep -v _test.go | grep -v '^[^:]*:[0-9]*:[ \t]*#'` → 0 matches.
- **E-parity**: `diff` each new wrapper's local vs template `.sh`/`.sh.tmpl` render; settings.json local vs template render — lockstep.
- **E-neutrality**: neutrality grep on the edited template files → 0 internal-token matches; `go test ./internal/template/... -run 'InternalContentLeak|Neutrality'` PASS.
- **E-catalog**: `go test ./internal/template/... -run 'Catalog'` PASS (agent/skill counts unchanged).
- **E5 — lint**: `golangci-lint run` → 0 NEW issues (pre-existing baseline reported separately).

## §F Milestones (reversibility-ordered — highest-change-likelihood first)

### M1 — Pattern-config schema + vulnerability-class taxonomy (Go, single-source)
**Change-likelihood: HIGH (the data model every layer consumes).**
- Define the single-source pattern table in `internal/hook/security/patterns.go`: a `[]VulnClass` where each class carries `{Name, Severity, Description, Patterns}` and patterns are language-neutral regexes (or a per-language-variant map keyed by the 16 languages). Organize by vuln CLASS, not by language (REQ-SG-011).
- Populate ≈25 patterns across the enumerated classes (REQ-SG-012), drawing vocabulary from the 4 security reference skills.
- TDD: `patterns_test.go` asserts class coverage (each enumerated class present), 16-language applicability (no class is single-language-only), and no false-positive on benign fixtures.
- Covers: REQ-SG-011, -012, -053.
- Review focus: is the class taxonomy the right cut? Are the regexes tuned to avoid noisy false positives? Is the per-language variant model or the language-agnostic model the right schema? (design.md §P records the schema decision.)

### M2 — Layer 1 scanner + PostToolUse handler + wrapper (regex-only, advisory)
**Change-likelihood: HIGH (the always-on core; false-positive rate shapes UX).**
- `internal/hook/security/scan.go`: read the PostToolUse payload (`tool_input.content` for Write, `tool_input.new_string` for Edit), run the pattern table, produce findings.
- `moai hook security-scan` subcommand in `internal/cli/hook.go` → dispatches the Layer-1 handler; emits `additionalContext` (async) advisory findings; never blocks (REQ-SG-014). Regex-only, in-process (REQ-SG-013).
- Template wrapper `handle-security-scan.sh.tmpl` (3-tier resolution + fail-open) + settings.json.tmpl PostToolUse entry (matcher `Write|Edit|MultiEdit`, async, 5s) + local siblings (byte-lockstep).
- TDD: `scan_test.go` (t.TempDir, no OTEL env parallel) asserts findings on a dangerous fixture, silence on a clean fixture, and no-block behavior.
- Covers: REQ-SG-010, -013, -014, -015, -061 (PostToolUse schema), -060 (fail-open), -052.
- Review focus: is the advisory-only path unambiguous? Is async the right posture (never stalls the edit)? Payload-shape handling for Write vs Edit vs MultiEdit.

### M3 — Layer-3 surface decision + commit-time cross-file review (opt-in, dormant default)
**Change-likelihood: HIGH (surface settled to L3-A by orchestrator — see §I).**
- Realize Layer 3 per the settled surface decision L3-A (design.md §L3): extend the sync-gate model with a sibling `security-commit-review` Stop hook that fires on ANY commit landing (deterministic Go), emitting a structured escalation signal the orchestrator translates into an `Agent()` cross-file review. Dormant (no-op) unless `MOAI_SECURITY_COMMIT_REVIEW` is enabled (REQ-SG-032).
- The cross-file review reads related files (read-only) to trace data flow (IDOR / auth-bypass / cross-file SSRF); advisory-by-default; opt-in blocking (REQ-SG-030, -031, -033).
- Template wrapper + settings.json entry (self-gating dormant) + local siblings.
- Covers: REQ-SG-030, -031, -032, -033.
- Review focus: is the dormant-by-default self-gate correct (no per-commit cost unless enabled)? Is the escalation signal shape complete for the orchestrator to spawn the `Agent()` review?

### M4 — Layer 2 turn-end diff review (Stop, advisory, opt-in escalation)
**Change-likelihood: MEDIUM (reuses M1 engine; escalation is bounded).**
- `moai hook security-turn` subcommand: at Stop, compute the turn's working-tree diff, run the Layer-1 pattern engine over it, surface high-severity findings via `systemMessage` (advisory). LLM/blocking escalation only under `MOAI_SECURITY_TURN_REVIEW` opt-in (REQ-SG-020, -021, -022, -023).
- Template wrapper `handle-security-turn.sh.tmpl` + settings.json Stop entry + local siblings.
- Once-per-turn discipline (avoid re-firing on every Stop) modeled on the sync-gate once-per-commit sentinel where applicable.
- Covers: REQ-SG-020, -021, -022, -023.
- Review focus: how is "the turn's diff" scoped (working-tree diff vs a recorded baseline)? Escalation opt-in wiring parity with `MOAI_SYNC_GATE_BLOCKING`.

### M5 — Hook boundary + orchestrator-translation + fail-open hardening
**Change-likelihood: MEDIUM (contract compliance across all 3 layers).**
- Verify + test the exit-code + structured-JSON contract per layer (REQ-SG-061): PostToolUse advisory additionalContext; Stop systemMessage / decision block; Layer-3 event schema.
- Confirm no hook invokes AskUserQuestion (REQ-SG-040); document the orchestrator-translation path for a block (REQ-SG-041); `--skip-hook` + `.moai/logs/hook-skip.log` audit for any gate (REQ-SG-043).
- Fail-open tests: missing `moai`/`jq`/`git` → silent no-op exit 0 (REQ-SG-060).
- Covers: REQ-SG-040, -041, -042, -043, -060, -061.

### M6 — Template-First parity + neutrality + catalog-count regression (mechanical)
**Change-likelihood: LOW (mechanical lockstep + guard compliance).**
- Byte-lockstep: each new wrapper local↔template; settings.json local↔template render.
- Neutrality grep on edited template files → 0 internal tokens; `internal_content_leak_test.go` PASS.
- Catalog counts unchanged (`catalog_tier_audit_test.go`); `go build ./...` (host + windows) exit 0; full `go test ./...`.
- Covers: REQ-SG-050, -051, -052.

## §G Anti-Patterns (avoid)

- **Layer-1 shelling out per pattern** (regex-in-a-loop via subprocess) — violates REQ-SG-013 (in-process only) and the 5s budget.
- **Blocking an Edit/Write by default** — violates REQ-SG-014 / REQ-SG-042 (advisory-first).
- **Scattering patterns across shell wrappers** — violates REQ-SG-053 (single-source Go table).
- **Elevating one language in the pattern table or wrappers** — violates REQ-SG-011 / REQ-SG-051 (16-language neutrality). Patterns are keyed by vuln class, applied across languages.
- **Running an LLM on every commit by default** — violates REQ-SG-032 (Layer 3 dormant/opt-in). Per-commit LLM cost is only paid when explicitly enabled.
- **A hook calling AskUserQuestion** — violates REQ-SG-040. Hooks emit JSON; the orchestrator translates a block.
- **Embedding SPEC-ID/date/SHA in a shell wrapper or settings.json** — violates REQ-SG-051 (§25 neutrality). Go source comments are exempt (not template content).
- **A new PostToolUse hook that homogenizes the wrapper-layer shared failure mode** — per `hook-independence.md` §7, the new wrapper must carry the 3-tier resolution + fail-open (mode A, acceptable) and add NO new un-logged bypass.

## §H Cross-References

- Layer architecture + exit-code/JSON contract + pattern-config schema + vuln-class taxonomy: `design.md`.
- Codebase research (hook system, settings.json wiring, existing gate hooks, precedents): `research.md`.
- Reused precedents: `sync-phase-quality-gate.sh`, `status-transition-ownership.sh`, `handle-pre-tool.sh` (Bash Risk-Amplifier), `hook-independence.md`.
- Hook boundary + orchestrator translation: `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface; Recovery-Signal Carve-Out (`runtime-recovery-doctrine.md` §4).
- Layering complement: SPEC-SEC-DEEPSCAN-001 (deep on-demand scan).
- Template-First + neutrality: CLAUDE.local.md §2 / §14 / §15 / §25; `internal/template/CLAUDE.md`.

## §I Resolved Design Forks (settled by orchestrator)

- **Layer-3 surface = L3-A (settled)** — extend the `sync-phase-quality-gate.sh` model with a sibling commit-time `security-commit-review` Stop hook that detects a commit landing (deterministic Go), reads the commit's changed + related files to surface cross-file candidates, and emits a structured escalation signal the orchestrator translates into an `Agent()` cross-file review. The native `type: agent` PreToolUse variant is NOT chosen — rationale: boundary-cleanliness (hooks emit structured JSON, orchestrator translates), speed, and consistency with the existing gate-hook precedent. Layer 3 stays OFF/dormant by default, opt-in via `MOAI_SECURITY_COMMIT_REVIEW` (consistent with `MOAI_SYNC_GATE_BLOCKING`).
- **L2/L3 escalation delivery = orchestrator-mediated `Agent()` (settled)** — hooks emit a structured block/finding on the exit-0 stdout JSON channel; the orchestrator reads it and spawns any needed `Agent()` review. Guardian hooks NEVER call a `type: agent`/`type: prompt` sub-model directly and NEVER call `AskUserQuestion` (REQ-SG-040 stays firm, grep-verifiable).

## Resolved Decisions (settled)

- **Tier** — RESOLVED: Tier L. Go pattern engine + 3 hook handlers + CLI wiring + tests + 3 shell wrappers (template+local) + settings.json (template+local) + single-source config = >15 files, security-critical, >1000 LOC.
- **SPEC ID** — RESOLVED: `SPEC-SEC-GUARDIAN-001` (regex self-check PASS; decomposition SPEC | SEC | GUARDIAN | 001; no collision).
- **Go vs shell for Layer 1** — RESOLVED: Go (single-source testable pattern table per REQ-SG-053 + the mission-anticipated `moai hook security-*` subcommand), NOT an inline shell regex. The shell wrapper only forwards to the Go handler (mode-A resolution + fail-open).
- **Default posture** — RESOLVED: L1 + L2 default-ON (regex-only, advisory); L3 default-OFF (dormant, opt-in). Blocking everywhere is opt-in only.
