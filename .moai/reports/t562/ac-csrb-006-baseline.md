# AC-CSRB-006 — doctor regression-guard baseline (t562 M1 window)

- Command: `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard' -count=1 -timeout 1800s -v`
- Exit code: `0` (PASS)
- Tree SHA: `69cfdce6f` (branch `WT-codex-read-inverse`, 2026-09-08) — PRE-change: the doctor
  conversion this card will add in M2 does not exist yet, so these counter values are the
  PRE-CHANGE behaviour M3's re-run must reproduce exactly.
- Raw output: `ac-csrb-006-baseline.log` (same directory) — carries the finding's summary and
  detail verbatim, rendered by the test's `t.Logf`.

## Exercised-count (the proof this baseline exercised the finding path)

**4 fixture entries flowed through `codexStaleSkillFinding`** — decomposition:

| Disposition | Count | Where it is visible in the captured output |
|---|---|---|
| resolved (the ok-style pass — counted neither missing nor unresolved) | **1** | NOT rendered anywhere: `missing=1` over 4 declared proves the existing file was read as resolving |
| counted missing (enabled bucket) | **1** | `1 with a path that no longer exists (1 enabled, 0 disabled, 0 unspecified, 0 non-boolean)` |
| relative (reported, never statted) | **1** | `1 relative entry (not checked: the resolution base is not observed)` |
| oddly-formed (reported, never statted) | **1** | `1 oddly-formed entry (not checked: backslash or ~other-user shape)` |
| **total traversed** | **4** | `declares 4 [[skills.config]] entries` — the phrase renders `len(entries)` as parsed from the fixture, and 1+1+1+1 accounts for all four |

The parsed-count phrase is the traversal witness the lead's requirement asks for: a diff against
an unexercised capture (e.g. `finding not produced`, or a detail naming 0 declared entries) would
now be distinguishable from a legitimate unchanged result, because the guard FAILS both of those
shapes before the counters are ever compared (the `!ok` fatal and the declared-count assertion).

## Fixture (all under t.TempDir(), home pinned via stubCodexHome — the real ~/.codex is never read)

1. absolute path to an EXISTING file (`liveSkillFile`) — `enabled = true` → resolves
2. absolute path to a MISSING file (`absentSkillPath`) — `enabled = true` → missing, enabled bucket
3. `relative-skills/SKILL.md` — `enabled = true` → relative, never statted
4. `C:\Users\u\SKILL.md` (backslash, raw bytes; the parser decodes no TOML escapes) — `enabled = true`
   → oddly-formed on this '/'-separator host, never statted

## Corroboration (not the binding baseline)

`.moai/reports/t540/ac-006-base.log` exists in this tree (absorbed), but per t540's own
`run-m1-m2.md` §2.7 it is t540's AC-CSPS-006 executed-test-control log — a `--- PASS:` count of the
full `internal/cli` suite (6886), not a doctor-output capture. It corroborates nothing about the
doctor finding's counters; this file + `ac-csrb-006-baseline.log` are the binding baseline for
AC-CSRB-006.

## M3 obligation (plan.md §E M3, as amended by the plan-audit D1 fix)

Re-run the same guard test post-M2 and diff against `ac-csrb-006-baseline.log`. Any diff — the
existing-file entry becoming counted-missing, a counter moving, the summary changing — is a FAIL,
never a new expectation.
