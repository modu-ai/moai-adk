# SPEC-MRGA-001 — Progress

## §E.2 Run-phase Evidence

AC-MRG-014 — the negative control on conjunct 3 of REQ-MRG-001.

The line below names a moving ref inside a git-command context, carries no SHA
and no frozen-baseline variable, and is nonetheless NOT a finding: it is a plain
instruction to measure, asserting nothing about what the result was. A rule that
drops the invariant-claim conjunct is the SIMPLER rule and passes every other
criterion in this SPEC while over-firing ~12x on the live corpus — this fixture is
the only thing that catches it.

- run before editing: `git fetch origin main && git rev-list --count --left-right origin/main...HEAD`
