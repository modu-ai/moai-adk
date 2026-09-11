# F13 verdict

## Claim

F13 is fixed: D8 evaluates a syscall mention against a build constraint or
explicit exemption in the same Markdown section. A build tag in another
section cannot satisfy the syscall section.

## Evidence

The fixture runs the documented section-scoped detector against one untagged
syscall section and one locally tagged section; only the untagged case emits
BLOCKING.

Command:

```text
.claude/hooks/tests/test-plan-auditor-d8-contract.sh
PASS: D8 binds syscall coverage to its own section
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f13`, based on local `develop` commit
`ffbf09ce2` before this card's commit.

## Gaps

The fixture checks SPEC prose, not the compiler's target-specific build. Code
cross-build remains a separate implementation verification.

## Residual-risk

A section can carry a syntactically present but semantically wrong build tag.
The auditor must still read the target and platform relationship after the
mechanical candidate is surfaced.
