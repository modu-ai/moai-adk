---
title: Safety and recovery
description: Plan review, authorization boundaries, preservation, receipts, and recovery limits in the Rust RC.
---

The Rust RC separates preview, publication, and evidence of host execution. These are different states. A successful file write is not proof that Claude Code or Codex discovered, trusted, loaded, or executed those files. See [CLI essentials](/en/reference/cli/) for the selected command syntax.

## Plan first; authorize separately

For the project and managed-instruction command families:

1. Confirm the intended project with `moai-cli identity PROJECT`.
2. Generate a read-only plan from the intended input. Review target identity, file/key changes, conflicts, warnings, and the plan digest.
3. Obtain authorization for those exact changes in the invoking host. Keep its real recorded reference.
4. Supply the reviewed plan's `result` object on stdin to the matching apply, together with `--plan-digest` and `--permission-reference`.
5. Inspect the result, receipt, and actual affected files before continuing to host validation.

The permission reference is caller-supplied metadata pointing to a host-recorded authorization. The binary rejects missing or placeholder references, but **does not verify that a real person approved the operation**. A plausible string is not permission. The host must enforce authorization and must not treat source files, instruction text, or a digest as approval.

This is not a claim that every CLI operation uses this same gate. Other command families have their own effect boundaries; read their contract before running them.

## What digest checks establish

Project plans bind protocol, target identity, expected bytes, action, and profile. Managed-instruction apply re-plans against the current tree and refuses stale or altered plans and conflicts. Project update also rechecks its expected file hash. If the target or inputs changed, stop, generate a new plan, review the difference, and obtain authorization for changed scope. Do not edit the stored plan or swap in a new digest to bypass review.

A digest establishes observed content identity—not publisher authenticity, code safety, or host compatibility. Package-tree reads are not OS-level atomic snapshots. Hash rechecks do not exclude concurrent writers or eliminate all filesystem races.

## Preservation boundaries

Rust project state uses `<project>/.moai-cli/`; Rust home state uses `~/.moai-cli/`. Do not conflate these with the preserved Go generation's `.moai/` and `~/.moai/`, and do not use Go `moai` commands as substitutes for Rust operations.

For `agents-plan/apply`, AGENTS.md remains user-owned outside the marked MoAI section. Three-way comparison uses baseline, current, and newly rendered bytes. User edits or removal are preserved when the template is unchanged; an incompatible template change becomes a conflict. A deleted managed section is not restored automatically. Duplicate or malformed markers cause a conflict.

AGENTS.local.md is opt-in and create-once; it is not subsequently rewritten or recreated after deletion. The `ignored` option adds the relevant ignore rule when needed, but is not proof of effective git-ignore status. These are operation-specific boundaries, not a promise that no Rust command can touch host configuration: host projection/settings commands have separately declared write scopes.

## Receipts are evidence, not a transaction

The project and agents apply paths record per-request receipts under `.moai-cli/receipts/<request-id>.json`. Entries identify paths, decisions, before/after digests, publication status, and errors. Restoration bytes live beside the receipt in `<request-id>.before/`. Treat these local bytes as sensitive project content and preserve them with the worktree.

Managed-file publication stages a temporary file in the target directory, syncs it, rechecks the target, then uses an exclusive hard link for creation or rename for replacement. **Atomicity is per file, not across a multi-file apply.** A later failure can leave earlier files published. Profile-baseline recording can also fail after the profile changes. Inspect a `partial` result rather than assuming nothing happened.

A crash between publication and receipt saving can leave an unrecorded file change. A subsequent managed-section apply may adopt identical rendered content as its baseline, but that is not a rollback. Removing the worktree also removes its local receipts. Not all other apply families use this receipt store; do not assume universal recoverability.

## Recover only the recorded request

The syntax is:

~~~text
moai-cli apply-recover PROJECT --receipt ID --permission-reference REFERENCE
~~~

ID is the actual receipt identifier returned by the apply or the receipt file stem. REFERENCE must point to actual host-recorded authorization for recovery, which can remove files. Inspect the receipt and current files before authorizing it.

Recovery restores recorded before bytes for replaced files and removes files the request created **only when current bytes still match the receipt's after digest**. Later user changes are skipped as `user_changed_since_apply`. Recovery refuses paths into `.moai/` or the receipt store and validates paths before changing files.

Recovery can exit 0 with `state: reverted` while `skipped` still contains user-edited files. Always inspect both `reverted` and `skipped`; success does not mean every original change was undone.

Recovery itself can be partial: a missing/altered backup or failed filesystem operation after a successful revert leaves a partial result. Review reverted, skipped, failed, and unattempted paths before deciding what to do next. Do not delete edited files or weaken digest checks to make recovery appear complete. It cannot recover unrecorded publications or external host/remote side effects.

## Evidence and support limits

The source contract records macOS/APFS fixture coverage; it is not evidence of live host execution or Windows/Linux filesystem behavior. Filesystems without hard-link support can fail exclusive creation; there is no silent copy-and-delete fallback. Local baseline records are not authenticated provenance.

Keep `host_loaded`, `execution_verified`, `host_verified`, and `not_run`/`unknown`/`blocked` results honest. A dry run, generated file, receipt, or passing fixture does not promote unexecuted host behavior to success. This documentation does not assert that a Rust release installer is published, that existing Go state has been migrated, or that applying is automatically reversible. Continue with the [getting-started guide](/en/guides/getting-started/) in an isolated project before considering real host changes.
