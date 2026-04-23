---
id: SPEC-V3R2-RT-003
title: "Sandbox Execution Layer (Bubblewrap / Seatbelt / Docker)"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 2 — Runtime Hardening"
module: "internal/sandbox/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-RT-002
  - SPEC-V3R2-RT-005
bc_id: [BC-V3R2-003]
related_principle: [P7 Sandbox Default, P6 Permission Bubble, P11 File-First]
related_pattern: [S-3, S-1]
related_problem: [P-C03, P-C01]
related_theme: "Layer 3: Runtime"
breaking: true
lifecycle: spec-anchored
tags: "sandbox, safety, security, v3r2, breaking, runtime, owasp"
---

# SPEC-V3R2-RT-003: Sandbox Execution Layer (Bubblewrap / Seatbelt / Docker)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | GOOS | Initial v3 Round-2 draft. New SPEC — no v3-legacy predecessor. Addresses P-C03 (no sandbox default, CRITICAL) per OWASP Top 10 for Agentic Apps 2025. |

---

## 1. Goal (목적)

Introduce an ephemeral sandbox layer that wraps implementer-agent tool executions in an OS-appropriate isolation primitive — Bubblewrap on Linux, Seatbelt (`sandbox-exec`) on macOS, Docker in CI — so that prompt-injection, supply-chain, and code-action vulnerabilities cannot damage the host machine even if permission checks are bypassed. The sandbox is the third safety layer behind SPEC-V3R2-RT-001 (audit trail via hook JSON) and SPEC-V3R2-RT-002 (permission envelope); together they form the defense-in-depth posture master §1.2 commits to.

Default-on for role profiles `implementer`, `tester`, `designer` per master §5 Principle P7 — the 2026 consensus (OWASP Top 10 for Agentic Apps December 2025, Cline npm-token exfiltration incident, Claude Code `rm -rf ~/` incident) is that approval prompts alone are empirically exploitable. moai v3 inverts the ecosystem default: every tool surveyed in r2-opensource-tools.md except snarktank/ralph has "none" for sandboxing; moai leads by correcting this gap while staying language-neutral (the sandbox wraps the shell layer, not the language toolchain).

Master §8 BC-V3R2-003 commits to AUTO migration: the v2→v3 migrator populates agent frontmatter with `sandbox: seatbelt` on macOS or `sandbox: bubblewrap` on Linux; opt-out requires `sandbox: none` with SPEC-documented justification. Master §10 R3 tracks the per-OS divergence risk.

## 2. Scope (범위)

In-scope:

- `Sandbox` Go string enum in `internal/sandbox/context.go` with values `"none" | "bubblewrap" | "seatbelt" | "docker"` matching master §4.3 Layer 3 type block.
- `SandboxLauncher.Exec(cmd []string, sandbox Sandbox, opts SandboxOptions) ([]byte, error)` wrapping any tool invocation in the OS-appropriate isolation.
- Per-OS backends:
  - Linux: `bwrap --unshare-all --die-with-parent --bind <scope> <scope> --ro-bind <read-only-dirs> -- cmd`
  - macOS: `sandbox-exec -p <profile>` with profile generated from `SandboxOptions`
  - Docker: `docker run --rm --network=<policy> -v <scope>:<scope> ...` for CI
- Network egress control: default denylist with allowlist for `github.com`, `registry.npmjs.org`, `pypi.org`, `proxy.golang.org`, `crates.io`, `repo.maven.apache.org`, `rubygems.org`, `pub.dev`, extensible via `.moai/config/sections/security.yaml` key `sandbox.network_allowlist`.
- File-write scope: writes restricted to the agent's worktree (from SPEC-V3R2-ORC-004) plus `.moai/state/`; all other paths mount read-only.
- Per-role default mapping surfaced in `workflow.yaml` `role_profiles.*.sandbox`:
  - `implementer` → `seatbelt` / `bubblewrap` (default-on per P7)
  - `tester` → `seatbelt` / `bubblewrap` (default-on per P7)
  - `designer` → `seatbelt` / `bubblewrap` (default-on per P7)
  - `researcher`, `analyst`, `reviewer`, `architect` → `none` (read-only, no sandbox needed)
- Feature detection: `moai doctor sandbox` verifies bubblewrap/sandbox-exec/docker availability and reports which backend will be used per agent on the current host.
- Opt-out: `sandbox: none` is permitted in agent frontmatter but the migrator never writes it; any agent with `sandbox: none` fails the v3 consolidator CI lint (SPEC-V3R2-ORC-002) unless accompanied by a SPEC comment naming a justification.
- LSP carve-out: `moai_lsp_*` ACI commands (from SPEC-V3R2-SPC-004) may require filesystem + local-socket access; the sandbox profile permits `~/.cache/` read-write and localhost-UNIX-socket use; master §12 Q7 surfaces this as validation at alpha.2.
- Environment variable scrubbing: sandbox strips `AWS_*`, `GITHUB_TOKEN`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `NPM_TOKEN`, `GH_TOKEN` from child process env unless the agent explicitly opts in via `sandbox.env_passthrough` in frontmatter.

Out-of-scope (addressed by other SPECs):

- Permission modes and resolver stack — SPEC-V3R2-RT-002.
- Hook JSON for passing sandbox verdicts back to Claude Code — SPEC-V3R2-RT-001.
- Worktree setup — SPEC-V3R2-ORC-004.
- Windows native sandbox (AppContainer / Win32 Job Objects) — deferred to v3.1+ per master §10 R3 (Docker is the CI fallback).
- Container registry management (pulling the Docker image) — deferred to SPEC-V3R2-EXT-004 migration framework.
- LSP server relocation into the sandbox process — deferred pending validation (master §12 Q7).

## 3. Environment (환경)

Current moai-adk state:

- No sandbox layer exists per problem-catalog.md P-C03 (CRITICAL). moai today has `permissions.allow` lists but no isolation per r2-opensource-tools.md §B Anti-pattern 1.
- 6 implementer agents already run with broad Write authority (expert-backend, expert-frontend, manager-ddd, manager-tdd, expert-refactoring, researcher per problem-catalog.md P-A11).
- `.moai/config/sections/security.yaml` exists and is loaded per r6 §5.1; this SPEC extends its schema.
- `workflow.yaml` role_profiles exist per CLAUDE.md §4 Dynamic Team Generation; this SPEC adds a `sandbox` field to each profile.

OS ecosystem reference:

- Bubblewrap (Linux): available in most distros; `bwrap` binary ships with Flatpak and many Desktop Linux distros. Rootless; uses user namespaces.
- sandbox-exec (macOS): ships in `/usr/bin/sandbox-exec` on every Mac since 10.5; `man sandbox-exec` documents the SBPL profile format.
- Docker (CI fallback): available in GitHub Actions, GitLab CI, CircleCI; moai emits a suggested `Dockerfile` snippet but does not manage image pulls.

Security incidents informing this SPEC:

- OWASP Top 10 for Agentic Apps (December 2025): codifies ephemeral sandboxing as mandatory for implementer agents.
- Cline 2026 npm-token exfiltration (r2-opensource-tools.md §5): approval-fatigue safety is empirically exploitable.
- Claude Code `rm -rf ~/` incident (master §1.2 rationale): the original trigger for moai's sandbox commitment.
- Alibaba LLM spontaneous cryptomining (r2-opensource-tools.md Executive summary): supply-chain and side-channel threats require isolation, not just permission.

Affected modules:

- `internal/sandbox/context.go` — new file, Sandbox enum + launcher.
- `internal/sandbox/bubblewrap.go` — Linux backend.
- `internal/sandbox/seatbelt.go` — macOS backend.
- `internal/sandbox/docker.go` — CI fallback backend.
- `internal/sandbox/profile.go` — SBPL and bwrap argument generators.
- `.moai/config/sections/security.yaml` — `sandbox.*` keys.
- `.moai/config/sections/workflow.yaml` — `role_profiles.*.sandbox` field.
- `.claude/agents/moai/*.md` — frontmatter `sandbox:` field migrated in by SPEC-V3R2-MIG-001.
- `internal/cli/doctor.go` — `sandbox` sub-subcommand.

## 4. Assumptions (가정)

- Linux users have `bwrap` available (Flatpak-era desktop distros, most server distros); when unavailable, `moai doctor sandbox` reports the gap and falls back to `sandbox: none` with user confirmation.
- macOS `sandbox-exec` is present on every supported Mac (10.5+); the binary is deprecated by Apple but still functional as of macOS 15.
- Docker backend is only invoked in CI mode (auto-detected via `CI=1` env var or explicit `--sandbox docker` flag); local dev defaults to bwrap/seatbelt.
- Network allowlist covers common package registries; projects with private registries extend the list via `.moai/config/sections/security.yaml`.
- LSP servers (gopls, pyright, tsserver, rust-analyzer) can run inside the sandbox profile with carve-outs for `~/.cache/` and localhost UNIX sockets; validation is tracked in master §12 Q7.
- Sandbox startup overhead is under 50 ms p95 (bwrap is ~10 ms, sandbox-exec ~20 ms, Docker ~1-5 seconds — CI-only so latency is acceptable).
- Windows users running in WSL2 fall through to bubblewrap; native Windows sandbox support is deferred to v3.1+.
- Agents with `sandbox: none` represent escape-hatch scenarios (dogfooding, legacy workflows) and are expected to be rare; CI lint flags them for review.
- The migrator (SPEC-V3R2-MIG-001) populates `sandbox:` per-OS at migration time; teams on mixed-OS should verify the migrated frontmatter matches their CI platform.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-V3R2-RT-003-001: The `Sandbox` type SHALL be a typed string enum with exactly 4 values: `"none"`, `"bubblewrap"`, `"seatbelt"`, `"docker"`.
- REQ-V3R2-RT-003-002: The system SHALL provide per-OS backends implementing the `SandboxBackend` interface: `Linux → bubblewrap`, `macOS → seatbelt`, `CI → docker`.
- REQ-V3R2-RT-003-003: Every agent role profile with `implementer`, `tester`, or `designer` designation SHALL default to the OS-appropriate sandbox backend, not `"none"`.
- REQ-V3R2-RT-003-004: Sandbox profile generation SHALL produce a deterministic artifact (checksum-stable across identical inputs) for auditability.
- REQ-V3R2-RT-003-005: `moai doctor sandbox` SHALL report the detected backend availability and per-agent resolved backend for the current host.
- REQ-V3R2-RT-003-006: Sandbox invocation SHALL scrub the environment variables `AWS_*`, `GITHUB_TOKEN`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `NPM_TOKEN`, `GH_TOKEN` from the child process unless the agent opts in via `sandbox.env_passthrough` frontmatter.
- REQ-V3R2-RT-003-007: File-write scope in the sandbox SHALL be restricted to the agent's worktree root plus `.moai/state/` by default; all other paths SHALL mount read-only.
- REQ-V3R2-RT-003-008: Network egress SHALL be denied by default; the pre-seeded allowlist covers `github.com`, `registry.npmjs.org`, `pypi.org`, `proxy.golang.org`, `crates.io`, `repo.maven.apache.org`, `rubygems.org`, `pub.dev`.

### 5.2 Event-Driven Requirements

- REQ-V3R2-RT-003-010: WHEN an agent is spawned with `sandbox: seatbelt` on macOS, the launcher SHALL generate an SBPL profile, invoke `sandbox-exec -p <profile>` wrapping the tool command, and forward stdout/stderr 1:1 to the caller.
- REQ-V3R2-RT-003-011: WHEN an agent is spawned with `sandbox: bubblewrap` on Linux, the launcher SHALL invoke `bwrap --unshare-all --die-with-parent --bind <writable-scope> <writable-scope> --ro-bind <read-only-dirs> <read-only-dirs> -- <cmd>`.
- REQ-V3R2-RT-003-012: WHEN the sandbox backend is unavailable on the current host, the launcher SHALL exit with error `SandboxBackendUnavailable`, emit a SystemMessage, and refuse to execute rather than silently degrading to unsandboxed execution.
- REQ-V3R2-RT-003-013: WHEN a sandboxed process attempts to write outside the allowed scope, the backend SHALL return `EPERM` / `EACCES` and the launcher SHALL surface a clear error naming the attempted path.
- REQ-V3R2-RT-003-014: WHEN a sandboxed process attempts network egress to a non-allowlisted host, the backend SHALL block the connection and emit a `SystemMessage` naming the host.
- REQ-V3R2-RT-003-015: WHEN `CI=1` is detected in the environment, the launcher SHALL prefer the `docker` backend for implementer agents regardless of OS, unless overridden by explicit `sandbox: seatbelt|bubblewrap` frontmatter.

### 5.3 State-Driven Requirements

- REQ-V3R2-RT-003-020: WHILE `.moai/config/sections/security.yaml` sets `sandbox.required: true`, any agent with `sandbox: none` in its frontmatter SHALL fail to spawn with error `SandboxRequired`.
- REQ-V3R2-RT-003-021: WHILE a sandbox profile is in use, the LSP carve-out SHALL permit read-write access to `~/.cache/` and bind-mount `/tmp` as a tmpfs under the sandbox scope.
- REQ-V3R2-RT-003-022: WHILE the agent has `permissionMode: plan`, the sandbox backend SHALL mount every path read-only (write scope is empty because plan mode forbids writes entirely per SPEC-V3R2-RT-002 REQ-020).

### 5.4 Optional Features

- REQ-V3R2-RT-003-030: WHERE a project extends `.moai/config/sections/security.yaml` with `sandbox.network_allowlist: [host1, host2, ...]`, the listed hosts SHALL be appended to the default allowlist for every agent in the project.
- REQ-V3R2-RT-003-031: WHERE an agent declares `sandbox.env_passthrough: [VAR_NAME, ...]` in its frontmatter, the listed variables SHALL be preserved in the child process environment despite the default scrubbing rules.
- REQ-V3R2-RT-003-032: WHERE `moai doctor sandbox --profile <agent-name>` is invoked, the system SHALL print the resolved profile (SBPL for macOS, bwrap args for Linux, Dockerfile snippet for CI) without executing the agent.
- REQ-V3R2-RT-003-033: WHERE an agent declares `sandbox: none` with a `sandbox.justification: "..."` frontmatter field, CI lint SHALL allow the opt-out but surface the justification in `moai doctor sandbox` output.

### 5.5 Unwanted Behavior

- REQ-V3R2-RT-003-040: IF a sandboxed process invokes `sudo`, `su`, or any setuid binary, THEN the sandbox backend SHALL deny the operation with exit code and emit SystemMessage naming the attempted escalation.
- REQ-V3R2-RT-003-041: IF sandbox profile generation produces an invalid SBPL or bwrap argument string, THEN the launcher SHALL refuse to execute and emit error `SandboxProfileInvalid` with the offending directive named.
- REQ-V3R2-RT-003-042: IF a sandboxed process produces output exceeding 16 MiB, THEN the launcher SHALL truncate to 16 MiB and emit a SystemMessage naming the truncation.
- REQ-V3R2-RT-003-043: IF an implementer agent spawns with `sandbox: none` AND `sandbox.required: true` AND no `sandbox.justification` field, THEN spawn SHALL fail with error naming the agent file.

### 5.6 Complex Requirements

- REQ-V3R2-RT-003-050: WHILE `moai doctor sandbox` detects backend unavailability (bwrap missing on Linux, for example), WHEN the user invokes an implementer agent, THEN the system SHALL offer a one-time AskUserQuestion prompt to install the backend or opt into `sandbox: none` with SystemMessage-logged consent.
- REQ-V3R2-RT-003-051: WHILE the agent's permission stack (SPEC-V3R2-RT-002) has resolved a tool to `PermissionDecision: "allow"`, WHEN the tool is dispatched inside the sandbox AND the sandbox also blocks the action, THEN the sandbox verdict wins and a SystemMessage names the divergence so users can reconcile settings.

## 6. Acceptance Criteria (수용 기준)

- AC-V3R2-RT-003-01: Given an agent with `sandbox: seatbelt` running on macOS invokes `Bash(touch /etc/passwd)`, When executed, Then the command fails with EPERM and a SystemMessage names the sandbox-denied path. (maps REQ-V3R2-RT-003-010, -013)
- AC-V3R2-RT-003-02: Given an agent with `sandbox: bubblewrap` on Linux invokes `Bash(curl https://evil.example.com)`, When the call is made, Then network is blocked with SystemMessage naming `evil.example.com`. (maps REQ-V3R2-RT-003-011, -014)
- AC-V3R2-RT-003-03: Given `workflow.yaml` role_profiles default for `implementer` is `seatbelt`, When an implementer agent spawns on macOS, Then the resolved sandbox is `seatbelt`, not `none`. (maps REQ-V3R2-RT-003-003)
- AC-V3R2-RT-003-04: Given `moai doctor sandbox` runs on a system without `bwrap` installed, When output is produced, Then stdout reports `bubblewrap: unavailable` and suggests installation or `sandbox: none` opt-out. (maps REQ-V3R2-RT-003-005, -050)
- AC-V3R2-RT-003-05: Given an implementer agent spawns on a host lacking the required backend, When it tries to run, Then spawn fails with `SandboxBackendUnavailable` (no silent unsandboxed execution). (maps REQ-V3R2-RT-003-012)
- AC-V3R2-RT-003-06: Given env contains `ANTHROPIC_API_KEY=sk-...` and sandbox is active, When the child process reads `ANTHROPIC_API_KEY`, Then the value is empty. (maps REQ-V3R2-RT-003-006)
- AC-V3R2-RT-003-07: Given agent frontmatter declares `sandbox.env_passthrough: [GH_TOKEN]`, When sandbox launches, Then `GH_TOKEN` is preserved in the child env despite default scrubbing. (maps REQ-V3R2-RT-003-031)
- AC-V3R2-RT-003-08: Given `security.yaml` has `sandbox.required: true`, When an agent with `sandbox: none` (no justification) is spawned, Then spawn fails with `SandboxRequired`. (maps REQ-V3R2-RT-003-020, -043)
- AC-V3R2-RT-003-09: Given agent in `plan` mode, When sandbox profile is generated, Then every filesystem path is mounted read-only (writable scope is empty). (maps REQ-V3R2-RT-003-022)
- AC-V3R2-RT-003-10: Given `security.yaml` adds `sandbox.network_allowlist: [internal.company.com]`, When the sandboxed process contacts `internal.company.com`, Then the connection succeeds. (maps REQ-V3R2-RT-003-030)
- AC-V3R2-RT-003-11: Given `CI=1` is set and the agent is `implementer` on Linux, When the sandbox resolves, Then the backend is `docker` (CI preference overrides OS default). (maps REQ-V3R2-RT-003-015)
- AC-V3R2-RT-003-12: Given a sandboxed process tries `sudo systemctl restart foo`, When executed, Then the call is denied with SystemMessage naming the setuid escalation. (maps REQ-V3R2-RT-003-040)
- AC-V3R2-RT-003-13: Given sandbox profile generation receives an input with a null byte, When generated, Then profile compiler emits `SandboxProfileInvalid` error naming the offending field. (maps REQ-V3R2-RT-003-041)
- AC-V3R2-RT-003-14: Given a sandboxed process writes 32 MiB to stdout, When the launcher returns, Then output is truncated to 16 MiB and SystemMessage names the truncation. (maps REQ-V3R2-RT-003-042)
- AC-V3R2-RT-003-15: Given an agent with `sandbox: none` and `sandbox.justification: "dogfooding legacy workflow X"`, When CI lint runs, Then the build succeeds and `moai doctor sandbox` prints the justification. (maps REQ-V3R2-RT-003-033)
- AC-V3R2-RT-003-16: Given permission resolver returned `allow` for `Bash(rm -rf /tmp/test)` but sandbox policy denies writes to `/tmp`, When dispatched, Then the sandbox verdict wins and a SystemMessage names the divergence between permission layer and sandbox. (maps REQ-V3R2-RT-003-051)

## 7. Constraints (제약)

- Technical: Go 1.22+; relies on OS binaries `bwrap` (Linux), `sandbox-exec` (macOS), `docker` (CI). Feature detection via `exec.LookPath`.
- Backward compat: v2.x agents without `sandbox:` frontmatter are migrated AUTO per master §8 BC-V3R2-003; opt-out requires explicit `sandbox: none` with justification.
- Platform: macOS 10.5+, Linux with user-namespaces enabled, GitHub Actions / GitLab CI / CircleCI runners. Windows native sandbox deferred to v3.1+ (Docker in WSL2 is the workaround).
- Performance: Sandbox startup overhead MUST be ≤ 50 ms p95 for bwrap and seatbelt; Docker may exceed this but is CI-scoped. Steady-state syscall overhead MUST NOT exceed 10% vs unsandboxed baseline for typical test runs.
- Binary size: Sandbox backends MUST NOT add more than 500 KiB to `bin/moai`.
- LSP validation: master §12 Q7 requires validation at alpha.2 that `moai_lsp_*` commands (SPEC-V3R2-SPC-004) work under the sandbox profile on all three backends.
- Env scrubbing list is enumerated (not regex) per REQ-V3R2-RT-003-006; projects extending the list use `sandbox.env_scrub` key in security.yaml (additive to defaults, not replacing).

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Per-OS divergence (bwrap vs seatbelt vs docker) creates bug surface in alpha.2 | H | M | Per-OS backends are separate files; CI matrix across macOS/Linux/Docker; `sandbox: none` fallback gated on doctor detection (master §10 R3). |
| LSP carve-out insufficient for pyright/tsserver/rust-analyzer — breaks ACI | M | H | Master §12 Q7 schedules validation at alpha.2; per-language LSP sandbox profiles documented as EVOLVABLE zone. |
| Sandbox startup latency regresses implementer-heavy workflows | L | M | 50 ms p95 budget per REQ constraints; Docker only in CI; bwrap/seatbelt are sub-20ms. |
| Env scrubbing misses a secret (custom var name) | M | H | `sandbox.env_passthrough` is opt-in but opt-out via `sandbox.env_scrub` is opt-out; doctor surfaces scrubbed vs passed env per agent. |
| Users disable sandbox globally via `sandbox: none` to silence failures | M | M | CI lint in SPEC-V3R2-ORC-002 requires `sandbox.justification` when using `sandbox: none`; doctor surfaces the justification and origin agent. |
| macOS sandbox-exec deprecation by Apple | L | H | Seatbelt still functional in macOS 15; v3.1+ revisits via Apple's newer `App Sandbox` entitlement model. |
| Windows-native absence strands Windows-only teams | M | M | Docker in WSL2 covers the gap for v3.0; master §10 R3 tracks the risk. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-RT-002 (provides `PermissionMode` integration: `plan` mode affects sandbox mount scope per REQ-V3R2-RT-003-022).
- SPEC-V3R2-RT-005 (provides `Source` provenance for `sandbox:` frontmatter, so `moai doctor sandbox` can report origin).
- SPEC-V3R2-ORC-004 (worktree MUST for implementers provides the writable scope the sandbox mounts read-write).

### 9.2 Blocks

- SPEC-V3R2-HRN-003 (harness thorough mode assumes sandboxed execution when `evaluator-active` scores artifacts).
- SPEC-V3R2-WF-003 (multi-mode router's `--mode loop` Ralph fresh-context iteration requires sandbox for implementer roles).
- SPEC-V3R2-ORC-001 (agent roster reduction — manager-cycle and builder-platform gain `sandbox:` defaults during consolidation).

### 9.3 Related

- SPEC-V3R2-MIG-001 (v2→v3 migrator writes `sandbox: seatbelt|bubblewrap` to agent frontmatter per OS).
- SPEC-V3R2-SPC-004 (ACI commands including `moai_lsp_*` must run under the sandbox; this SPEC defines the carve-out).
- SPEC-V3R2-CON-001 (P7 sandbox default lives in the FROZEN zone declared here).
- SPEC-V3R2-MIG-003 (security.yaml loader picks up new `sandbox.*` keys).

## 10. Traceability (추적성)

- Theme: master §4.3 Layer 3 Runtime; §5 Principle P7; §8 BC-V3R2-003; §10 R3 per-OS risk.
- Principle: P7 (Sandboxed Execution as Default Safety Layer); secondary P6 (Permission Bubble), P11 (File-First — sandbox is thin shell wrapper, not framework).
- Pattern: S-3 (Ephemeral Sandboxed Execution, priority 8); S-1 (Permission stack composes with sandbox).
- Problem: P-C03 (no sandbox default, CRITICAL); P-C01 (no permission bubble, CRITICAL — composes with this SPEC).
- Master Appendix A: Principle P7 → primary SPEC-V3R2-RT-003.
- Master Appendix C: Pattern S-3 → primary SPEC-V3R2-RT-003 (priority 8).
- Wave 1 sources: r2-opensource-tools.md §A Pattern 5 (OWASP mandate), §B Anti-pattern 1 (approval-fatigue); r1-ai-harness-papers.md §12 CodeAct anti-pattern flag (sandboxing is non-trivial).
- Wave 2 sources: design-principles.md P7 (Sandbox Default); pattern-library.md S-3; problem-catalog.md Cluster 5 P-C03.
- BC-ID: BC-V3R2-003 (sandbox-by-default for implementer agents, AUTO migration on per-OS basis).
- Priority: P0 Critical — master §1.2 commits v3 to correcting the ecosystem sandbox gap; blocks harness thorough and loop mode.
