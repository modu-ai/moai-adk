# SPEC-SYNC-SHA-SLOT-FORMAT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored in worktree `.claude/worktrees/t299`, branch `WT-sha-slot-format`,
frozen baseline `BASELINE_SHA = a6bbbf82b`. Tier S, era V3R6.

Artifacts: `spec.md`, `plan.md`, `acceptance.md`, this file. Note the Tier-S deviation: `acceptance.md`
is present although Tier S normally inlines its AC in `spec.md` §3 — written on the lead's explicit
instruction, because the regression triad's falsifiability contract needs the room.

Measurements recorded in `spec.md` §B, each with the command that produced it. Commands the plan phase
actually ran, for reproduction:

```
grep -h '^sync_commit_sha:' .moai/specs/*/progress.md | wc -l                                  → 346
… | sed 's/^sync_commit_sha:[[:space:]]*//' | grep -cE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)'  → 334
… | grep -vE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)'                                            → 12 (11 SPECs)
… | grep -c '#'                                                                                 → 105
grep -rh 'pending-backfill' .moai/specs/*/progress.md | grep -oE 'pending-backfill[a-zA-Z0-9-]*' | sort | uniq -c
                                                                                                → 28 spellings
diff .claude/rules/moai/development/spec-frontmatter-schema.md \
     internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md      → exit 0
```

### Plan audit iter-1 → v0.2.0 remediation

Verdict FAIL 0.80 (`.moai/reports/t299/plan-audit-iter1.md`), three blocking defects, all remediated
at spec.md v0.2.0. The central correction: the flagged set is **5**, not 12 — REQ-SSF-005's exemption
removes seven placeholders — and all five sit in `completed` SPECs, so the predicted non-advisory
count is **0**, not 2. Re-derived here:

```
python3 .moai/reports/t299/grammar_check.py          → total 346 | SHA 334 | PLACEHOLDER 7 | FLAGGED 5
… | grep -E '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)' | grep -c '#'   → 99  (of 334 conforming)
grep -h '^mx_commit_sha:' .moai/specs/*/progress.md | wc -l         → 85
… | grep -vcE '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)'               → 9   (3 deliberate declarations)
grep -A6 'var eraDemotableCodes' internal/spec/lint.go              → MissingExclusions, FrontmatterInvalid
```

The `pending-backfill` family command in `spec.md` §B.3 now excludes this SPEC's own directory: the
verbatim command quoted in this file self-matches twice, which moved the cited 29 to 31 for anyone
re-running it.

Open questions carried to the operator: the nine `completed`-SPEC slots (repair or leave — `spec.md`
§E) and the t354 coordination recommendation (`spec.md` §D.4).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
