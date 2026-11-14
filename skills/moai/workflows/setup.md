# SETUP

## Input

Repository facts, purpose, users, stack, constraints, and the selected hosts.

## Output

Interview decisions, a generation plan, portable files, and an apply receipt (planned; not implemented in 2026.11.0-rc.1). The current release reports the apply result as JSON on stdout instead.

## Procedure

1. Read the manifests, existing docs, skills, rules, hooks, agents, and the dirty state, and record a baseline of the user's files.
2. On first setup, run one full, in-depth interview covering goals, technology, collaboration, required features, domain, quality, security, cost, and the scope of autonomy. Ask the writing language item by item with `moai language-setup <project>`: with a host question tool, ask the `--questions` set and pass the answers back with `--answers`; without one, run it without either flag to answer on the terminal. It writes nothing; pass `result.profile` to `project-plan`, then `project-apply`. In a project whose Go-generation folder was kept as `.moai-backup`, its `language.yaml` values are already the suggested answers and `result.go_backup` lists them (with `project_id`, a name candidate, and the Go keys not carried over); say they came from the earlier generation and that nothing is copied (Issue #2817).
3. Build the project-specific agents and skills from what step 1 and the interview found. The tool installs no third-party skills (the external skill install operations were removed, Issue #2286, user decision 6774899518 row 35); a skill the user wants from elsewhere is the user's own install, outside this workflow. Running scripts and granting additional permissions are handled separately.
4. Design canonical workflows and resources plus the selected native roles, instead of a fixed bundle of agents. `generation-plan` previews the per-file three-way judgment from a typed request, and `generation-apply` publishes approved files (managed AGENTS.md section, portable workflows, requested native roles, skill projections with attribution) into the project state directory and the host official folders, recording baselines and `workflow_id`/`profile_sha256` in the generation manifest.
5. Preview, per file, whether it is created, merged as managed content, or in conflict, along with permission diffs and remote portability.
6. Within the approved scope, re-check the expected hashes, then apply through temp → validate → replace → receipt; the receipt step is planned and not implemented in 2026.11.0-rc.1.
7. **Allowed initial local commit — dry-run first, commit only.** Render the proposed initial commit (generated files, managed-section changes, receipt) as a dry-run diff and show it before anything is recorded. Only after the user approves that exact diff, create the local commit; the approval covers the commit alone — it never implies push or Issue publication (a local commit is not a publication permission), dirty user changes are not swept in, and an existing commit is never amended. Record the effect with `effect-intend` before and `effect-confirm` after (operation marker `setup_initial_commit`) in the setup receipt.
8. **Reviewed Issue publication — a separate approval.** Render the Issue draft from the committed state and show it for review; approving the initial commit never approves publication. Only after a separate publication approval, publish exactly the reviewed draft (after checking for duplicate Issues/SPECs/PRs), one Issue per approved draft, and then read it back (`gh issue view`) and confirm the published text matches the reviewed draft before reporting success. Record the effect with `effect-intend` before and `effect-confirm` after (operation marker `setup_issue_publish`) in the setup receipt.
9. Verify host discovery, explicit invocation, references, internal execution, and a zero diff on a repeated request.

## Exceptions and Permissions

If the baseline changes, replan. Native trust, credentials, and hook registration each require their own permission. Never force a remote installation.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
