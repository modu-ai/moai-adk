# t508 — codex `[[skills.config]].enabled` acceptance lab

Every row below was measured in THIS run, in an isolated `CODEX_HOME` under `/tmp/t508lab`.
The real `~/.codex` was never written (it was read once, for the census at the bottom).

- Binary: `/Users/goos/.local/bin/codex`, `codex-cli 0.153.4`
- Probe command: `codex mcp list` (a config-loading, non-interactive verb; `codex --version`
  does NOT load the config and returns rc=0 on every fixture — it is not a usable probe)
- Fixture body: one `[[skills.config]]` entry, `path` pointing at an existing `SKILL.md`,
  the `enabled` line varied per row.

## Measured acceptance table

| # | `enabled` line | rc | codex message |
|---|---|---|---|
| 1 | `enabled = true` | 0 | (normal output — control) |
| 2 | *(line absent)* | 1 | ``missing field `enabled` in `skills.config` `` |
| 3 | `enabled = 1` | 1 | ``invalid type: integer `1`, expected a boolean in `skills.config.enabled` `` |
| 4 | `enabled = "true"` | 1 | ``invalid type: string "true", expected a boolean in `skills.config.enabled` `` |
| 5 | `enabled = 'false'` | 1 | ``invalid type: string "false", expected a boolean in `skills.config.enabled` `` |
| 6 | `enabled = true`, **`path` line absent** | 0 | (normal output) |

Row 1 is the control that makes rows 2-5 attributable: the lab config itself is loadable.
Row 6 is the scope control: `path` is OPTIONAL to codex, `enabled` is not — so this defect
class is scoped to `enabled` and does not widen to `path`.

## What this means for the moai parser

`internal/codexwiring/skills.go` reads the same lines and reaches a DIFFERENT verdict on
three of the five fatal rows:

| Fixture | codex | `ParseSkillEntries` reading | Agreement |
|---|---|---|---|
| `enabled = true` | accepts | `SkillEnabledTrue` | ✅ |
| *(absent)* | **rc=1** | `SkillEnabledUnspecified` — documented as "asserts nothing" | ❌ silent |
| `enabled = 1` | **rc=1** | `SkillEnabledUnspecified` (regex does not match the line) | ❌ silent |
| `enabled = "true"` | **rc=1** | **`SkillEnabledTrue`** | ❌ **actively wrong** |
| `enabled = 'false'` | **rc=1** | **`SkillEnabledFalse`** | ❌ **actively wrong** |

The last two rows are the harder finding, and the card did not anticipate them. The absent-key
case degrades to a NEUTRAL reading (`Unspecified` asserts nothing). The quoted-value case
degrades to a CONFIDENT HEALTHY reading — the parser reports a live, enabled registration for a
config on which codex will not start at all.

The quoted-value leniency is deliberate and tested:

- `internal/codexwiring/skills.go` ~line 60 — "A quoted `true` is a TOML string rather than a
  boolean, so it is arguably malformed; reading it as false, however, silently DEMOTES a live
  registration to stale bookkeeping, which is the more damaging misreading."
- `internal/codexwiring/skills_test.go:66` `TestParseSkillEntriesEnabledQuoted` asserts
  `'true'` → `SkillEnabledTrue`.
- `internal/cli/doctor_codex_test.go:297` — `EnabledKey: "\"true\""` commented
  "quoted string, still true".

**Its stated rationale is falsified by row 4.** There is no live registration to demote: codex
exits 1 before loading anything. Reversing it is therefore a correction, not a preference — but
it reverses a documented, tested decision, so it is an operator call, not a lane call.

## Blast radius of an ERROR grade

`internal/cli/doctor.go:137` `doctorExitStatus` — any `uikit.CheckFail` makes `moai doctor`
exit 1 (`Warn` stays exit 0, deliberately). Reporting this shape as ERROR therefore turns
`moai doctor` non-zero for an affected user. That is the correct semantics (their codex is
non-functional) but it IS a behavioural change to a command hooks and CI wrappers read.

`internal/cli/doctor_codex.go:114` `codexFinding{summary, detail}` carries no severity field and
every problem lands on `uikit.CheckWarn` (~line 286). There is currently **no ERROR grade on
this check at all** — the severity axis has to be built, not merely selected.

## Real-machine census (read-only)

```
$ grep -c '^\[\[skills\.config\]\]' ~/.codex/config.toml
49
$ grep -A3 '^\[\[skills\.config\]\]' ~/.codex/config.toml | grep '^enabled' | sort | uniq -c
  49 enabled = false
```

49/49 declare a bare boolean. Nothing on this machine is in the fatal set — consistent with the
card's premise that the hazard is latent until a WRITING surface (t502 / t506) lands. The file
was read, never written.

## Gaps

- One codex version only (0.153.4). Whether `enabled` became required in a particular release,
  and whether older versions tolerate its absence, was not measured.
- Only single-entry configs were probed. Whether codex reports the FIRST offending entry or all
  of them is unmeasured — it affects finding wording, not the verdict.
- `enabled` inside a multi-line string or with a trailing comment was not probed; the parser has
  its own handling for both and those paths are unchanged by this card.
- `moai doctor`'s end-to-end exit code was read from `doctorExitStatus`, not executed against a
  fatal fixture. The run-phase guard must execute it.
