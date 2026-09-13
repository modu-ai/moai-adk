# t524 — Group 4 verb on the original SPEC (read-only sanity run)

Command:

```
bash .moai/reports/t524/run/run-fixtures.sh .moai/reports/t524/run/verb-template.sh \
  .moai/reports/t524/run/real .moai/reports/t524/run/work/wd-real \
  .moai/specs/SPEC-CODEX-PARTIAL-WIRING-001
```

Output (`real/SPEC-CODEX-PARTIAL-WIRING-001.txt`, `exit=0`):

```
COLLECTED: 11 REQ definitions (acceptance input: read)
ORPHAN: REQ-CW-009
ORPHAN: REQ-CW-010
```

- `COLLECTED: 11` with no `UNCOVERED:` line: the verb reads this SPEC's definitions and resolves its shorthand mappings. The mapping this card started from, `REQ-CPW-001, 002`, was the axis the auditor miscounted in the original case.
- The two `ORPHAN:` lines are **not** orphans of this SPEC. `grep -n 'REQ-CW-0\(09\|10\)'` shows them only at `spec.md:20` and `spec.md:170`, where they are requirements of the predecessor SPEC, cited next to that SPEC's `AC-CW-012`. The verb correctly narrows here, and the auditor dismisses the lines on reading.
- Follow-up in the same card: the `ORPHAN:` prose in both agent copies was amended to name this cause ("another SPEC's requirement cited next to that SPEC's AC"). The verb bytes did not change.

Not observed: whether the plan-auditor agent itself, given this verb, reaches the right AC-4 / AC-5 verdict on this SPEC. That is an LLM judgement, and this sanity run does not exercise it.
