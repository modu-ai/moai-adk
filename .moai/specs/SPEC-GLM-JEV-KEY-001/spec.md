---
id: SPEC-GLM-JEV-KEY-001
title: "moai glm --key flag and new moai jev --key command (shared credential storage)"
version: "0.1.0"
status: draft
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, glm, jev, credential, key-storage, typesafe"
tier: M
---

# SPEC-GLM-JEV-KEY-001

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-10-09 | 0.1.0 | Initial draft — factory card t1613 (Class C), plan phase (manager-spec). |

## Position in the chain

- Upstream: SPEC-GLM-KEY-INPUT-001 (completed) owns the GLM credential writer SSOT (`internal/glmcred`); SPEC-JEV-CORE-001 (completed) owns the TypeSafe credential writer SSOT (`internal/jevcred`, REQ-JEVC-018/020) and the `workflow.jev.enabled` gate.
- Downstream: factory card t1612 (terminal-output redesign) adds JEV guidance once this card lands. t1612 is NOT part of this SPEC's scope.

## §A. Problem Statement

The GLM API key has one sanctioned save surface, the `moai glm setup <key>` subcommand. Operators asked for a flag-shaped equivalent (`moai glm --key <value>`) that persists the same value to the same storage, without retiring the subcommand. Separately, the Jev (TypeSafe) credential has storage and readers (`internal/jevcred`, `~/.moai/.env.typesafe`, consumed by `moai doctor`, the web console, and the `jev_ask` gate) but NO top-level CLI command: `moai jev` is an unknown command on the installed build, so there is no CLI route to store that credential.

## §B. Goal

1. `moai glm --key <value>` stores the GLM API key through the existing `glmcred` writer — same file, same format, same value — while `moai glm setup <key>` keeps working unchanged and `moai glm [-p profile] [-f | -l]` launch behavior is untouched.
2. A new top-level `moai jev` command accepts `--key <value>` and stores the Jev credential through the existing `jevcred` writer, mirroring the GLM storage approach (mode 0600 dotenv file outside the repository).
3. Neither surface ever echoes the full key back; disclosure stays within the masked forms each domain already asserts.

## §C. Requirements (GEARS)

**REQ-GJK-001** (Ubiquitous) The `moai glm` command shall accept a `--key <api-key>` flag whose save path is the same credential writer `moai glm setup <key>` uses (`glmcred.Save`: file `~/.moai/.env.glm`, dotenv key `GLM_API_KEY`, mode 0600).

**REQ-GJK-002** (Event-driven) **When** `moai glm --key <value>` is run with no subcommand, the command shall store the value through that writer, print the existing masked confirmation `GLM API key stored (<masked>)`, and exit without launching a GLM session.

**REQ-GJK-003** (Event-driven) **When** the `--key` flag is detected together with any other argument token that reaches the scan region (a launch-entry flag such as `-p`, `-f`, `-l`, `-w`, `-b`, `--permission-mode`, `--spawn`, `--branch`), the command shall refuse with a usage error naming the conflict and store nothing. The scan region begins AFTER the pre-existing `--help`/debug scans and the manual subcommand routing — the refusal never overrides them (REQ-GJK-006 precedence): `moai glm --key K --help` prints help and stores nothing, and `moai glm setup --key K` routes to setup unchanged (the preserved legacy trap, plan.md §B.1). Tokens after a bare `--` are child passthrough, outside the scan region (AC-GJK-016).

**REQ-GJK-004** (Event-driven) **When** the `--key` flag carries no value (`moai glm --key` at end of arguments), the command shall fail with a usage error and store nothing.

**REQ-GJK-005** (Event-driven) **When** the trimmed `--key` value is empty, the command shall fail with the setup path's empty-key error and store nothing.

**REQ-GJK-006** (Event-driven) **When** `moai glm setup <key>` (or `status`, `tools`) is routed, the manual subcommand routing in `runGLM` shall take precedence over the `--key` scan, so no routed subcommand invocation changes behavior.

**REQ-GJK-007** (Ubiquitous) A top-level `moai jev` command shall be registered on the root command, accepting `--key <api-key>` and storing the value through `jevcred.Save` (file `~/.moai/.env.typesafe`, dotenv key `TYPESAFE_API_KEY`, mode 0600) — the same storage `moai doctor`'s Jev check, the web console, and the `jev_ask` opt-in guidance read.

**REQ-GJK-008** (Event-driven) **When** `moai jev --key <value>` is run, the command shall store the credential and print a confirmation that discloses at most what `jevcred.View` permits: the configured fact plus, for a credential longer than four characters, its final four characters — and no part of a credential of four characters or fewer.

**REQ-GJK-009** (Event-driven) **When** `moai jev` is run without `--key`, the command shall print its help text and exit 0 without touching storage.

**REQ-GJK-010** (Unwanted) Neither command shall echo the full key value to stdout, stderr, or any log output.

**REQ-GJK-011** (Where) **Where** a credential file already exists at a file mode wider than 0600, a save through either command shall tighten the file to mode 0600 (behavior inherited from the owning packages' `Save`).

**REQ-GJK-012** (Event-driven) **When** a trimmed `--key` value on either new save path contains a carriage-return or line-feed character, the command shall refuse with a validation error BEFORE invoking the credential writer, leaving the existing credential file byte-for-byte unchanged (reject-before-write, mirroring the `internal/web/jevkey.go` validator; the legacy `setup` surface is excluded by §D's unchanged-setup exclusion).

## §D. Exclusions

### Out of Scope — Terminal-output redesign (card t1612)

- Bare-`moai jev` rich guidance, JEV terminal rendering, and any 4-locale (en/ko/ja/zh) parity work for renderer surfaces belong to card t1612. This card only makes its own user-facing strings enumerable (see plan.md §D.7).

### Out of Scope — Wizard, gate, and MCP surfaces

- No changes to `internal/cli/init_jev_wizard.go`, the `workflow.jev.enabled` gate, or the `jev_ask` MCP tool.
- No changes to the web console's Jev credential section beyond what it already reads from `jevcred`.

### Out of Scope — Storage-model changes

- No new credential file (no `.env.jev`), no new environment-variable name, no second writer implementation. Storage stays delegated to `internal/glmcred` and `internal/jevcred`.
- No credential-reveal route for Jev (SPEC-JEV-CORE-001 REQ-JEVC-020 ships none; this SPEC adds none).
- No key rotation, multiple profiles, or expiration semantics.

### Out of Scope — GLM setup surface

- `moai glm setup <key>` keeps working unchanged; no deprecation warning, no aliasing, no removal.

## §E. Open Questions

None. All design decisions are pinned in plan.md §D and routed in decision-index.md (storage location and disclosure policy are decided by completed SPEC-JEV-CORE-001 REQ-JEVC-018/020; the remaining rows are implementation-level with defaults applied).
