---
description: "Contract-sign guard — the tool-call boundary deny of `moai contract sign` / `moai contract decide` from agent sessions. Loaded when working on the guard, its tests, or this rule itself."
paths: "internal/hook/contract_sign_guard.go,internal/hook/contract_sign_guard_test.go,internal/hook/contract_sign_guard_units_test.go,internal/template/templates/.claude/rules/moai/workflow/contract-sign-guard.md"
---

# Contract-Sign Guard — Mode-Independent Deny

The PreToolUse guard in `internal/hook/contract_sign_guard.go` denies agent
invocations of `moai contract sign` and `moai contract decide` at the
tool-call boundary. The deny is **mode-independent**: it is active in every
workflow autonomy mode, including `guided`.

The companion promise that "nothing changes under `guided`" is scoped in
words to the escalation detector; the contract-sign guard is not covered by
that promise. A contract signed by an agent is worthless in any mode, so the
guard does not weaken itself to preserve the sentence — the resolution is in
the wording, not the behavior.

## What the guard enforces

- **Human signing path** (`--signer human`, or `--signer` absent): denied in
  every session, with no exemption. Signing happens at an operator terminal;
  run the command there.
- **Non-interactive path** (`--signer llm` / `llm+jev`): denied when the
  calling session claims the factory worker role via the
  `MOAI_FACTORY_ROLE=worker` marker; allowed otherwise.
- **`moai contract decide`**: same role-marker gate as the non-interactive
  sign path.
- **Unclassifiable invocations**: a command carrying a contract sign/decide
  call whose structure cannot be classified (command substitution, an unknown
  signer value, or deeper shell indirection) is denied fail-closed.

Every deny carries the sentinel prefix `CONTRACT_SIGN_AGENT_VIOLATION:`,
which the orchestrator matches. `--receipt` is never inspected: the guard
recognizes no receipt argument.

An allowed call leaves the hook output byte-identical to the no-guard
baseline and writes no record: the guard fails closed on danger and stays
silent otherwise.

## Scope of this rule

`paths:`-scoped on purpose — a guard's documentation does not earn a place
in every session's always-loaded budget. The rule loads when a session reads
or edits the guard's source, its tests, or this file.
