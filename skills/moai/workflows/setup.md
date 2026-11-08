# SETUP

## Input

Repository facts, purpose, users, stack, constraints, and the selected hosts.

## Output

Interview decisions, a generation plan, portable files, and an apply receipt (planned; not implemented in 2026.11.0-rc.1). The current release reports the apply result as JSON on stdout instead.

## Procedure

1. Read the manifests, existing docs, skills, rules, hooks, agents, and the dirty state, and record a baseline of the user's files.
2. On first setup, run one full, in-depth interview covering goals, technology, collaboration, required features, domain, quality, security, cost, and the scope of autonomy.
3. For each project-relevant external skill, confirm its source, license, version, trust, and target scope, install it into the host's official folder, and then read it in full. The typed operations carry this: `skill-source-plan` judges the candidate (license evidence, revision, telemetry policy) from a typed request, `skill-source-apply` copies approved trees into the host's official folder and records the provenance in `.moai-cli/manifests/sources.json`, and `skill-read-receipt` records that every installed file was read. Build the project-specific agents and skills from that reading. Running scripts and granting additional permissions are handled separately.
4. Design canonical workflows and resources plus the selected native roles, instead of a fixed bundle of agents. `generation-plan` previews the per-file three-way judgment from a typed request, and `generation-apply` publishes approved files (managed AGENTS.md section, portable workflows, requested native roles, skill projections with attribution) into the project state directory and the host official folders, recording baselines and `workflow_id`/`profile_sha256` in the generation manifest.
5. Preview, per file, whether it is created, merged as managed content, or in conflict, along with permission diffs and remote portability.
6. Within the approved scope, re-check the expected hashes, then apply through temp → validate → replace → receipt; the receipt step is planned and not implemented in 2026.11.0-rc.1.
7. Verify host discovery, explicit invocation, references, internal execution, and a zero diff on a repeated request.

## Exceptions and Permissions

If the baseline changes, replan. Native trust, credentials, and hook registration each require their own permission. Never force a remote installation.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
