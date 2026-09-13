# F06 verdict

## Claim

F06 is fixed: Critical and High security findings now share one blocking
contract across sync quality gates and the sync-auditor. Continuing requires an
explicit, user-approved exception record; it is no longer a phase-dependent
warning choice.

## Evidence

Added `.claude/rules/moai/core/security-decision-contract.md`, referenced it
from the auditor, and replaced the CRITICAL-only / HIGH-warning wording in the
sync quality gate. The static contract test verifies the severity table,
references, and absence of the stale policy.

Command:

```text
.claude/hooks/tests/test-security-decision-contract.sh
PASS: one severity and exception contract is referenced by sync audit
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f06`, based on local `develop` commit
`480c82508` before this card's commit.

## Gaps

The workflow remains text-routed: this card does not change a remote security
scanner's API or force a user prompt at runtime. The contract test is therefore
static and must accompany any future router implementation.

## Residual-risk

A future document can still invent a local severity rule unless it is checked by
the contract test. The canonical path is deliberately named in the affected
agent and workflow files to make such drift detectable.
