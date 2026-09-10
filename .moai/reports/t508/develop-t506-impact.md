# t508 — impact of card t506 landing on develop

Measured after the t526 integration window, reading the `develop` ref directly (the t508 working
tree was NOT modified — a `plan-auditor` run was live against `0b1e27877` at the time).

- t508 branch base: `0b1e27877`
- local `develop` at measurement: `5d08342e4` (carries t506, `SPEC-CODEX-GHOST-SKILLS-PRUNE-001`)

## Why this matters to t508

The card's priority rationale was that the hazard is **latent** until a surface that WRITES
`[[skills.config]]` lands, naming t502 and t506 as the coming writers. One of them has landed:
`internal/cli/codex_skills_prune.go` (`moai clean --codex-skills`) writes the user's
`~/.codex/config.toml`:

```
git show develop:internal/cli/codex_skills_prune.go | grep -n 'WriteFile'
205:	if err := os.WriteFile(backupPath, raw, 0o600); err != nil {
211:	if err := os.WriteFile(cfgPath, pruned, 0o600); err != nil {
```

## Question asked: can the prune verb CREATE the fatal shape?

**Measured answer: no.** Two independent reasons, both read from the landed source:

1. **Removal is whole-entry, computed once.** `pruneCodexSkillEntries` drops the closed range
   `[e.StartLine, e.EndLine)` for every eligible entry, over one line slice:

   ```go
   for i := e.StartLine; i < e.EndLine; i++ { drop[i] = true }
   ```

   It never removes a subset of an entry's lines, so it cannot strip an `enabled` line while
   leaving its `[[skills.config]]` header standing — which is what would manufacture the fatal
   shape out of a healthy entry.

2. **An unreadable line disqualifies the whole entry.** `judgeCodexSkillEntry` returns a skip when
   `e.FirstUnrecognizedLine >= 0`. A non-boolean value such as `enabled = 1` does not match
   `skillEnabledKeyRe`, so it is an unrecognised line and its entry is PRESERVED rather than
   pruned. The disqualification happens to point the safe way here.

So the two cards do not compose into a new defect. t508's premise is unchanged, and its scope does
not widen to t506's code.

## What DID change, and is worth stating in the SPEC

The hazard moved from **latent** to **live** in one specific sense: `moai clean --codex-skills`
will now happily rewrite a `config.toml` that codex cannot load at all. A user in the fatal set
runs the prune, sees moai report success, and still has a codex that exits 1 — with nothing
anywhere telling them why. The doctor finding t508 adds is the only surface that would.

This strengthens the priority argument without changing the scope: t508 diagnoses, t506 writes,
and neither needs the other's code.

## The quoted-value leniency survives t506 — the reversal is still live work

```
git show develop:internal/codexwiring/skills.go | grep -n 'skillEnabledKeyRe'
131:	skillEnabledKeyRe   = regexp.MustCompile(`^enabled\s*=\s*(?:(true|false)|"(true|false)"|'(true|false)')\s*(#.*)?$`)
```

Unchanged: the quoted alternations are still there, and the tri-state still folds them onto
true/false. §B.1's reversal has not been overtaken.

## Citation drift — the SPEC's `file:line` references must be restamped

The SPEC was authored against `0b1e27877`. Measured against `develop`:

| SPEC cites | develop | drift |
|---|---|---|
| `skills.go:32` — `SkillEnabledUnspecified` | 34 | +2 |
| `skills.go` ~60 — the quoted-leniency rationale ("DEMOTES a live registration") | 125 | +65 |
| `skills.go` — `skillEnabledKeyRe` | 131 | — |
| `doctor_codex.go:114` — `codexFinding` | 114 | 0 |
| `doctor_codex.go:654` — `codexStaleSkillFinding` | 654 | 0 |
| `doctor_codex.go` ~732 — the `missing == 0 && unresolvedShape == 0` early return | 732 | 0 |
| `doctor_codex.go` ~286 — `check.Status = uikit.CheckWarn` | 283 | −3 |
| `doctor.go:137` — `doctorExitStatus` | 142 | +5 |

Three citations moved. They must be restamped after the branch absorbs develop, and restamping is
not cosmetic: a `file:line` that points at the wrong line is a citation that cannot be checked, and
an uncheckable citation is indistinguishable from an unverified claim.

## New surface the SPEC did not account for

`internal/codexwiring/skills.go` grew ~144 lines on develop: `SkillEntry` now carries `StartLine`,
`EndLine`, and `FirstUnrecognizedLine`, and the package exports `SplitConfigLines` /
`JoinConfigLines`. Two consequences for t508's run-phase:

- The parser change t508 proposes (a state for a declared-but-non-boolean `enabled`) now lands in a
  file with a second consumer — the prune verb — that reads `Enabled` and `FirstUnrecognizedLine`.
  Any change to what those fields mean must be checked against `judgeCodexSkillEntry`, not only
  against the doctor.
- `internal/codexwiring/skills_extent_test.go` (new, 221 lines) pins the extent behaviour. The
  run-phase test suite must not collide with it.

## Gaps

- Whether card t502 (the other named writer) has landed was NOT checked.
- The prune verb's behaviour was read, not executed. The claim "cannot create the fatal shape" rests
  on reading `pruneCodexSkillEntries` and `judgeCodexSkillEntry`; no fixture was run against it.
  A run-phase check is cheap and should be added rather than inheriting this reading.
- The line-drift table was measured against `develop` at `5d08342e4`. develop moves; restamp
  against the tree the branch actually absorbs, not against this table.
