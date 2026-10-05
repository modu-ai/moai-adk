---
title: moai contract Autonomy Contracts
weight: 100
draft: false
---

`moai contract` verifies, shows, and signs a SPEC's autonomy contract (`.moai/specs/<SPEC-ID>/contract.yaml`). A contract records what an agent may do without asking (actions), the paths it may write and the paths it must never touch (ownership), the conditions that stop it and hand control back to a person (escalate_on), and its budget. A signature binds that content and `acceptance.md` by hash, so `verify` reports a mismatch as soon as either changes after signing.

{{< callout type="info" >}}
In this release a signature **does not replace Implementation Kickoff Approval.** The contract is a record you can verify; the human approval gate stays in place.
{{< /callout >}}

## Subcommands

| Command | Description | Exit codes |
|---------|-------------|------------|
| `moai contract verify <SPEC-ID>` | Checks the contract. Read-only, safe to call from hooks | 0 valid · 1 invalid · 2 usage or I/O error |
| `moai contract show <SPEC-ID>` | Prints the contract sections, signature state, and derived sets (such as the effective never-write paths) | 0 printed · 2 usage or I/O error |
| `moai contract sign <SPEC-ID>...` | Signs contracts | 0 signed · 1 refused · 2 usage or I/O error |
| `moai contract kickoff-check <SPEC-ID> --card <card>` | Checks the contract-mode plan→run gate. Read-only | 0 pass · 1 fail · 2 usage or I/O error |
| `moai contract decide <card> --spec <SPEC-ID> --judgement <file\|->` | Evaluates a kickoff decision and records it in the contract store | 0 recorded · 1 store integrity error · 2 usage or I/O error |
| `moai contract revoke <card> --spec <SPEC-ID>` | Withdraws the signature of a signed contract | 0 revoked (or already revoked) · 1 not signed · 2 usage or I/O error |

`verify` and `show` accept `--json` to print a machine-readable object. `verify` reports invalidity only through a closed set of reason codes (`unsigned`, `acceptance_hash_mismatch`, `contract_digest_mismatch`, and so on).

## moai contract sign

```bash
moai contract sign SPEC-AUTH-001
moai contract sign SPEC-AUTH-001 --resign
moai contract sign SPEC-AUTH-001 --signer llm \
  --receipt .moai/specs/SPEC-AUTH-001/kickoff-receipt.json
```

| Flag | Description |
|------|-------------|
| `--signer <human\|llm\|llm+jev>` | Who signs. Defaults to the human path |
| `--receipt <path>` | Kickoff receipt, relative to the project root. Required for `llm` and `llm+jev` |
| `--resign` | Re-signs a signed contract after its `acceptance.md` changed. The previous signature is recorded as `supersedes` |

**Human path.** Runs only in an interactive terminal. It shows a signing summary and signs only after you type the confirmation token (the SPEC ID, or `sign N contracts` when signing several at once). If an agent-marker environment variable is set, or stdin is not a terminal, it refuses before any prompt. Signing several SPECs at once requires `workflow.autonomy.contract.batch_sign`.

**Receipt path.** `--signer llm` or `--signer llm+jev` together with `--receipt` signs from a kickoff receipt, with no terminal confirmation. This path is open only when `workflow.autonomy.mode` is `contract`, and it signs one SPEC at a time.

**Refusals.** A refused signing changes no file and prints one line, `refused <code> (<SPEC-ID>): <cause>`. The code comes from a closed set (`not_tty`, `confirmation_mismatch`, `already_signed`, `plan_audit_not_passing`, `verify_failed`, and so on). Each file to be signed is verified again before it is written and is written only when it verifies as validly signed; the author's comments and blank lines are kept.

## moai contract kickoff-check · decide · revoke

```bash
moai contract kickoff-check SPEC-AUTH-001 --card t42
moai contract decide t42 --spec SPEC-AUTH-001 --judgement judgement.json
moai contract revoke t42 --spec SPEC-AUTH-001
```

**kickoff-check.** Confirms that the card equals the contract's `card` field, that `verify` reports a valid signature, that the contract store records that signature, and that no revocation covers it. It writes nothing; on failure it lists every reason.

**decide.** Takes the deciding LLM's judgement (a JSON file, or `-` for standard input), evaluates the kickoff preconditions and outcome rules, and records the outcome (`approve`, `reject`, or `human`) in the store. A decision writes the receipt `.moai/specs/<SPEC-ID>/kickoff-receipt.json`. `decide` calls no LLM itself and never signs.

**revoke.** Appends a revoke event to the store and writes one `kind: revoke` escalation record. A run in progress stops at the next stage boundary. It never deletes a worktree or branch, pushes, changes the queue, or edits the contract or SPEC documents.

{{< callout type="warning" >}}
Autonomous kickoff **ships disabled.** In this version `kickoff-check` refuses `llm` and `llm+jev` signatures as `autonomous-kickoff-inactive`, so entering the run phase still requires a human signature.
{{< /callout >}}

## Related

- [Config section reference — workflow.yaml autonomy](/en/advanced/config-sections/#workflowyaml--autonomy)
- [moai spec](/en/cli-reference/spec)
- [Autonomy tier (MOAI_AUTONOMY_TIER)](/en/advanced/autonomy-tier) — similar name, unrelated setting
