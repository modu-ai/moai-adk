# F14 verdict

## Claim

F14 is fixed: project documentation audit calls now pass `input_type=project`,
the auditor has a project-specific required document set and rubric, and the
project report path no longer depends on a SPEC `spec.md`.

## Evidence

The workflow and auditor input contracts were aligned and the static test
checks the typed dispatch, three required project files, and project report
path.

Command:

```text
.claude/hooks/tests/test-project-auditor-input-contract.sh
PASS: project audit uses a typed project input contract
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f14`, based on local `develop` commit
`7186f8800` before this card's commit.

## Gaps

The project rubric itself remains prose; no live plan-auditor subprocess was
available in this card.

## Residual-risk

Callers that omit `input_type` can still be rejected only after dispatch. The
explicit typed contract makes that omission a visible blocker rather than a
misleading SPEC audit.
