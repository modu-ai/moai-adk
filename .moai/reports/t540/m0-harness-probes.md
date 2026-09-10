# t540 — M0 harness-shape probes (report only; M0 NOT built)

Lane-1, 2026-09-08. Scope granted by the lead: determine whether the M0 harness is FIXABLE.
No M0 implementation, no acceptance.md edit. `codex --version` = `codex-cli 0.153.4`, re-stamped
at each probe. All probes ran against an isolated `CODEX_HOME` under the session scratchpad; the
real `~/.codex` was never named.

## Claim

1. A `[[skills.config]]` entry does **not** introduce a skill. Placement under a skill root does.
2. The disable gate **is** observable: `enabled = false` suppresses a skill that otherwise renders,
   and restoring `enabled = true` brings it back.
3. AC-CSPS-001's Then clause, as written, is **vacuous** — it asserts the null state. Detail below.
4. The auth axis does not block M0 (measured separately; see §E.4).

## Evidence

Fixture, identical across all three cells:

```
$CODEX_HOME/config.toml
$CODEX_HOME/skills/t540probe/SKILL.md     (frontmatter: name: t540probe)
```

Command, identical across all three cells:

```
CODEX_HOME=<scratch>/home timeout 90 codex debug prompt-input "hi" < /dev/null
```

| # | `config.toml` skills.config | rc | stderr | `grep -c 't540probe'` |
|---|---|---|---|---|
| 1 | (no entry) | 0 | 0 bytes | **1** |
| 2 | `path = <abs>`, `enabled = false` | 0 | 0 bytes | **0** |
| 2b | `path = <abs>`, `enabled = true` | 0 | 0 bytes | **1** |

Cell 1 rendered roots and the probe line, verbatim:

```
`r0` = `<scratch>/m0probe/home/skills`
`r1` = `/Users/goos/.agents/skills`
`r2` = `<scratch>/m0probe/home/skills/.system`
- t540probe: t540 harness-shape probe skill, isolated CODEX_HOME only (file: r0/t540probe/SKILL.md)
```

Cell 2b is the attribution control: the same fixture, the same command, only the flag flipped
back. Without it, cell 2's zero would not be attributable to `enabled = false` rather than to
run-to-run variance.

Prior cell, measured earlier the same day and consistent with the above: a `[[skills.config]]`
entry with `enabled = true` pointing at a SKILL.md at an **arbitrary path outside any skill root**
produced `grep -c` **0**, and that path appeared in neither the roots list nor the skill list.

### Exported evidence (the citation target)

The deciding lines of every cell are exported to tracked paths and this report cites those files,
not the scratch tree they were extracted from:

- `.moai/reports/t540/probe-harness-shape.log` — the three harness cells: per-cell rc, stderr and
  stdout byte counts, `grep -c 't540probe'`, the rendered probe line, the roots table, and the
  surviving fixture `config.toml`.
- `.moai/reports/t540/probe-auth-axis.log` — the auth-axis cell: rc, stderr/stdout byte counts,
  the roots table, and the arbitrary-declared-path hit count.

Both were extracted from the scratch outputs mechanically, not retyped. Per
`agent-common-protocol-reference.md` § Evidence export obligation, the raw renders themselves are
**deliberately not exported** and are therefore **not cited** anywhere in this report — see
Residual-risk.

## Baseline-attribution

All cells measured this run, this host (darwin), `codex-cli 0.153.4` re-stamped per probe, against
an isolated `CODEX_HOME` in the session scratchpad. No figure is carried over from §E.2 or from
the t502 measurements.

## The polarity finding

`acceptance.md` AC-CSPS-001 states:

> **Then** the probe skill appears in the rendered prompt-input skill roots — establishing that
> Codex resolves the forward-slash form.

and its fallback clause treats non-appearance in the slash arm as refutation.

Cell 1 shows appearance is the **default**: with no `skills.config` entry at all, a skill under a
skill root appears. So "appears" cannot discriminate resolution from non-resolution — if Codex
failed to resolve a slash-form path, the entry would simply be inert and the skill would appear
anyway, from the root. The criterion is satisfied by the null state, which is the vacuous-green
shape `verification-completeness.md` §1.1 names.

The discriminating signal runs the other way. Cells 2 and 2b show the observable event is
**disappearance under a disable entry**. A slash-arm probe therefore reads: place the SKILL.md
under a skill root, declare a `[[skills.config]]` entry carrying the **forward-slash** form of its
path with `enabled = false`, and observe whether the skill **disappears**. Disappearance means the
path matched — Codex resolved the slash form. Persistence means it did not.

This is a description of the shape the evidence implies, not an implementation and not an
amendment. Whether AC-CSPS-001 is amended in place or the correction becomes a follow-up card is
the lead's decision.

## Gaps

- **The Windows axis is untouched.** Every cell ran on darwin, where `filepath.Separator` is
  already `/`, so the slash form and the native form are the same string. Nothing here measures
  what Codex does with `C:/…` against an on-disk `C:\…`. AC-CSPS-001 remains OPEN and unmeasured.
- **The control arm of the real M0 (native-backslash form) was not run** — it is unrunnable here
  for the same reason.
- **Skill-root isolation is incomplete.** `r1` resolved to the real `/Users/goos/.agents/skills`
  even under an isolated `CODEX_HOME`. Probes are therefore not hermetic on a developer machine;
  a distinctive probe name (`t540probe`) was used so a collision would be visible. On a clean CI
  runner this would not arise.
- **Only the `enabled` flag was varied.** Path *shape* (relative, directory-shaped, symlinked) was
  not varied; the t502 report covers that axis and was not re-measured here.
- **No claim is made that the harness, so shaped, will pass.** Only that the gate is observable
  and the current Then clause is not.

## Residual-risk

- Reading "the gate is observable" as "M0 is ready" would be the same two-axes-is-not-half-closed
  error the §E.4 `do_not_claim` list guards against. The Windows axis is the gate; it is untouched.
- The polarity finding rests on cell 1 being a true control. If some other property of the
  arbitrary-path cell (not its position outside a skill root) explains its zero, the inference
  that placement is what introduces a skill would be wrong. The two explanations were not
  separated by a further probe.
- **The raw prompt-input renders are not exported and are a known loss.** Four files totalling
  ~139 KB (`p1.out`, `p2.out`, `p2b.out`, and the auth cell's `out.txt`) stay in the session
  scratchpad and will not survive `/tmp` clearance. They were left behind under the selection
  criterion — the lines that decided each verdict are in the two exported logs, and a 139 KB
  attachment of rendered system-prompt text would not make the verdict more checkable. The
  consequence is real and is accepted: an auditor cannot re-read the full render, only the
  extracted lines and the counts taken from it. Nothing in this report rests on a line that was
  not exported.
- **The per-cell `config.toml` fixtures were overwritten in place**, so only cell p2b's survives
  and is what the export carries. Cells p1 and p2 are described in prose and by their measured
  outcome; their exact fixture bytes are gone. Re-running the probes would reconstruct them, but
  reconstruction is not the original.

## Cross-layer sweep after the AC-CSPS-001 amendment (2026-09-08)

**Claim.** The amendment that inverted the AC-CSPS-001 observable to disappearance left one
plan-layer inconsistency, now repaired, and no residual echo of the old appearance phrasing.

**Evidence.**

1. Inconsistency found and repaired — `plan.md` §E M0 bullet 2 still prescribed two arms while the
   amended AC requires three:

   ```
   -- Run BOTH arms on a Windows host: slash-form `path` and native-backslash-form `path` (control).
   ++ Run all THREE arms on a Windows host: a baseline arm with no `skills.config` entry, then the
      slash-form and native-backslash-form arms (both `enabled = false`).
   ```

   Following the un-amended procedure would have run two arms without a baseline and could not have
   satisfied the amended AC's exit condition.

2. Residual-echo sweep — clean, both files:

   ```
   $ /usr/bin/grep -nEi 'appears|appear |appearance|나타남|나타난' <spec.md|plan.md>
   spec.md: (0 hits)
   plan.md: (0 hits)
   ```

**Baseline-attribution.** Measured in this run, in this worktree
(`.claude/worktrees/t540`, branch `WT-codex-path-escape`), against `769ae6f3c`.

**Gaps.** The sweep covered `spec.md` and `plan.md` only — the two artifacts the amendment's
observable is cited in. `acceptance.md` carries the amended text itself and was not swept for echo.

**Residual-risk.** A paraphrase of the appearance framing that shares none of the swept tokens would
not have been caught; the repair above was found by reading the procedure, not by the token sweep.
