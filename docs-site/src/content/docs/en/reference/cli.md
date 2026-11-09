---
title: CLI essentials
description: Core Rust moai-cli discovery, profile, package, and managed-instruction commands.
---

This is a selective reference for the Rust **moai-cli** RC, not the Go **moai** executable. Do not substitute one binary for the other. The RC's internal JSON commands are not a promise of a published installer or a stable public command vocabulary. Start with [Getting started](/en/guides/getting-started/) and read [Safety and recovery](/en/reference/safety/) before any apply.

## Discover without changing the project

Run these with the Rust binary you built or otherwise verified:

~~~sh
moai-cli --version
moai-cli --help
moai-cli project-apply --help
moai-cli identity .
moai-cli install-inspect --json
moai-cli baseline-inventory .
~~~

Help and version are plain CLI output. The other commands return JSON envelopes. Identity reports clone/workspace identity and state roots without writing. Linked worktrees share a clone ID but have distinct workspace IDs. IDs currently derive from canonical paths: moving a checkout can change them. Do not proceed on an unexpected target or a git failure.

Install inspection describes observed installations; it does not install anything. Baseline inventory is read-only and optionally accepts `--go-baseline-commit SHA` and `--go-repo REPO` to inspect a recorded Go baseline. Neither command authorizes migration or Go removal.

## Command notation

The following are **syntax**, not copy-and-run approval examples. PROJECT is the existing target directory; PACKAGE is an explicit local skill-package directory. DIGEST comes from the reviewed plan's `plan_digest`. REFERENCE must identify authorization actually recorded by the invoking host for that operation and scope. Never invent it.

~~~text
moai-cli project-plan PROJECT
moai-cli project-apply PROJECT --plan-digest DIGEST --permission-reference REFERENCE
moai-cli package-inspect PACKAGE
moai-cli package-resolve PROJECT [--repo-package PACKAGE] [--user-package PACKAGE]
moai-cli agents-plan PROJECT
moai-cli agents-apply PROJECT --plan-digest DIGEST --permission-reference REFERENCE
moai-cli apply-recover PROJECT --receipt ID --permission-reference REFERENCE
~~~

## Project profile

`project-plan` reads a typed Profile JSON object from stdin and writes a plan without changing files. New planning inputs require `schema_version: 2`. Required fields are `schema_version`, `project_id`, `purpose`, `tools`, `development_methods`, and `hosts`; hosts use `claude` and/or `codex`. Optional fields include `package_pin`, `interview_revision`, and structured `decisions`. Unknown fields and invalid values are rejected; do not put credentials or interview transcripts into a profile.

Inspect the returned `action`, per-key `delta`, target binding, expected hash, and digest. Actions are `create`, `update`, `unchanged`, or `conflict`. Apply takes the **exact plan object from the successful envelope's result on stdin**, not the whole envelope or a newly edited profile. It writes the project profile under `.moai-cli/project.yaml` and the profile baseline/receipt records. A successful profile write does not prove a host loaded it.

## Packages

`package-inspect` computes the `moai-package-tree-v1` digest over SKILL.md and support files. This checks content identity, not publisher trust or executable safety.

`package-resolve` is read-only. A profile package pin requires the explicitly supplied repository package with that exact digest; absence or mismatch fails without fallback. Without a pin, it selects only the explicitly supplied user package. The flags are syntactically optional, but resolution still needs the package for the applicable branch. There is no implicit home scan or host activation.

## Managed instructions

`agents-plan` reads this request shape from stdin:

~~~json
{
  "protocol_version": 1,
  "language": "en",
  "workflow_reference": ".moai-cli/workflows",
  "local_file": "none"
}
~~~

This is a request example, not a directory-creation command. The referenced repository-relative workflow directory must already exist, and the project needs its profile. Use the generation workflow first when workflows have not yet been generated; do not create an empty directory merely to satisfy the check. Set `language` to `ko` for Korean output. `local_file` accepts `none`, `ignored`, or `tracked` and controls the personal-file opt-in.

Review every file decision and warning. Any `conflict` prevents apply. Feed the exact successful plan's `result` to `agents-apply` on stdin with its digest and genuine authorization reference. This family manages an AGENTS.md section, optional AGENTS.local.md creation, relevant .gitignore sections, and the generation manifest—not arbitrary user files. For host-specific setup see [Claude Code](/en/guides/claude-code/) or [Codex](/en/guides/codex/).

## Results and failures

Internal operations return one stdout JSON envelope: `{protocol_version, operation, ok, result}` on success, or `{protocol_version, operation, ok, error}` on failure. Check `ok` and the process exit status before reading `result`. Exit codes are 0 for success, 1 for an operation failure, and 2 for command-line misuse. Help, version, launchers, and unknown commands are exceptions to the envelope contract.

Use `error.kind` and `next_actions` when present. A stale plan needs review and replanning, not a forced digest replacement; a partial apply needs receipt inspection. This page intentionally omits forge writes, release/migration operations, and the rest of the internal operation surface.
