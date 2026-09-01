# t382 AC-EH3-007 proposition 2 — per-row judgment of newly-flagged DRIFT rows

Tree: `c0cdb2fd3` (working tree, era.go/era_test.go/audit_test.go modified — uncommitted at
measurement time). Binaries: `./bin/moai-pre-t382` (pre-fix) and `./bin/moai` (post-fix), both
built from this tree by `make build`.

Measurement commands (both `--no-cache`; see the cache note at the bottom):

```
./bin/moai-pre-t382 spec drift --no-cache > .moai/reports/t382/m3-drift-before-nocache.txt
./bin/moai           spec drift --no-cache > .moai/reports/t382/m3-drift-after-nocache.txt
```

`era-exempt` row count: **250 → 228** (−22, the predicted 22 rows).

Of the 24-SPEC swept population, after the change:

| Disposition | Count | Note |
|---|---|---|
| still `era-exempt` | 1 | `SPEC-V3R5-INIT-WIZARD-EXPANSION-001` — no modern signal, earned exemption |
| `terminal-exempt` | 1 | `SPEC-HOOK-PREEDIT-INVESTIGATE-001` — superseded; row unchanged as predicted |
| now git-compared, `aligned` | 7 | frontmatter matches the git-implied status |
| now git-compared, `DRIFT` | 4 | judged individually below |
| **row dropped entirely** | 11 | `getGitImpliedStatus` returns `no git history found for <id>` (no commit on `main` names the SPEC-ID), and `detectDrift` drops the record. NOT a drift finding — an absence of one. Unforeseen by the SPEC; recorded here rather than silently folded into "aligned". |

## The 4 DRIFT rows, judged individually

| # | SPEC | frontmatter | git-implied | verdict |
|---|---|---|---|---|
| 1 | `SPEC-HARNESS-LEARNING-EVO-002` | completed | implemented | **genuine mismatch** |
| 2 | `SPEC-KANBAN-TODO-CLI-001` | in-progress | implemented | **genuine mismatch** |
| 3 | `SPEC-UPDATE-VERSION-FLAG-001` | completed | implemented | **genuine mismatch** |
| 4 | `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` | completed | in-progress | **FALSE POSITIVE** |

### 1 — SPEC-HARNESS-LEARNING-EVO-002 — genuine

`git log main --grep=SPEC-HARNESS-LEARNING-EVO-002` returns exactly one commit,
`96bedbfe3 feat(SPEC-HARNESS-LEARNING-EVO-002): delegation-map analyzer (L2) (#1424)`.
The same grep on `origin/develop` returns the same single commit. The frontmatter claims
`completed`; no close commit exists on either branch. The detector is right.

### 2 — SPEC-KANBAN-TODO-CLI-001 — genuine

`a84e48961 feat(SPEC-KANBAN-TODO-CLI-001): moai todo CLI with lock-guarded backlog store (#1529)`
is on `main`, so git implies `implemented`, while the frontmatter still reads `in-progress`
(`updated: 2026-08-14`). The frontmatter lags its own landed work. The detector is right.

### 3 — SPEC-UPDATE-VERSION-FLAG-001 — genuine

`852b2206b feat(SPEC-UPDATE-VERSION-FLAG-001): add --version <tag> flag to moai update (#1301)`
is the only commit on `main` (and on `origin/develop`). Frontmatter says `completed`; no close
commit exists. The detector is right.

### 4 — SPEC-V3R6-SESSION-HANDOFF-AUTO-001 — FALSE POSITIVE

The SPEC genuinely closed: `e979a4d13 chore(SPEC group C): Mx-phase close (status
implemented→completed, 2026-06-02)` names it in the commit BODY
(`- SPEC-V3R6-SESSION-HANDOFF-AUTO-001: Mx verdict EVALUATE-PASS, ...`), and the frontmatter's
`updated: 2026-06-02` matches that date exactly.

The detector mis-infers for two pre-existing reasons that compound:

1. `getGitImpliedStatus` greps commit **bodies**, not just subjects, so the newest match is
   `7cffb9717 feat(SPEC-HANDOFF-AUTORESUME-001): plan-phase 산출물` — a DIFFERENT SPEC's
   plan-phase commit that merely mentions this id in its body. Walking newest-first, the
   detector classifies from that subject and infers `in-progress`.
2. The real close commit uses a **combined/abbreviated scope** (`chore(SPEC group C)`), the exact
   form `.claude/rules/moai/development/spec-frontmatter-schema.md` § Close-subject full-ID
   mandate prohibits precisely because "the drift detector's exact-token SPEC-ID extraction
   cannot map an abbreviated prefix to its sibling SPECs, so combined-scope close subjects
   regenerate lifecycle drift false-positives."

Neither cause is introduced by SPEC-ERA-H3-NARROWING-001. The grandfather exemption was
**hiding** this defect, not preventing it. But under AC-EH3-007 proposition 3 as written
("오탐이 1건이라도 확인되면 이 AC는 실패이고 설계를 다시 연다"), one confirmed false positive
is a FAIL, so AC-EH3-007 is reported FAIL rather than passed by reinterpreting its own
criterion.

## Cache note — the first "after" measurement was vacuous

`moai spec drift` caches its report keyed on the HEAD SHA. Because this change is uncommitted,
HEAD did not move, and the first post-fix run was served from the pre-fix cache: the drift
output was **byte-identical** to the pre-fix run (`diff` empty, `era-exempt` 250 in both) while
`moai spec audit` already reported the new classification. The R4 probe still exited 0 on that
stale output, because its sweep population is derived from the V3R5 audit set — which had
already shrunk to 1. A green with a one-row sweep over a stale file.

Every figure in this document comes from `--no-cache` runs. Neither `spec.md` §3 nor `plan.md`
§C mentions `--no-cache`; the plan's §C pre-flight recipe (`./bin/moai spec drift > ...`) is
cache-exposed and would have produced this same vacuous green at run-phase entry.
