# t506 — lane-side live dry-run (closing a residual risk the run phase named)

Taken by the lane session, not by the run agent, after the run phase closed.
Tree `.claude/worktrees/t506`, branch `WT-codex-ghost-skills`, HEAD `e27f00b5a`.
Date 2026-09-07.

## Why this exists

The run phase's own Residual-risk section stated that
`moai clean --codex-skills` **had not been exercised through a built binary** —
every observation came from Go tests against `t.TempDir()` fixtures. That is a
real gap: unit tests exercise the functions, not the wiring from the command
surface down to them.

Dry-run is read-only by construction, so the gap can be closed against the live
config without touching it. A safety copy was taken first anyway
(`/tmp/t506-codex-config.safety.toml`, sha256 identical), because "read-only by
construction" is the claim under test, not a premise to rely on.

## Claim

The verb, built from this tree, runs end to end against this machine's real
`~/.codex/config.toml`, judges all 49 entries eligible, and writes nothing.

## Evidence

```
$ go build -o /tmp/t506-moai-run ./cmd/moai
build rc=0

$ /tmp/t506-moai-run clean --codex-skills > /tmp/t506-dryrun.txt 2>&1; echo "rc=$?"
rc=0

$ tail -1 /tmp/t506-dryrun.txt
· 49 of 49 entries eligible for removal. Run with --force to actually remove.

$ wc -l < /tmp/t506-dryrun.txt
50            # 49 per-entry lines + 1 summary
```

Post-run state of the live file:

```
$ shasum -a 256 ~/.codex/config.toml
9f6e3a953880630afcfa6a40abe846b785e8fc0067e17b3a7ec1d513f71ca33a

$ grep -c '^\[\[skills.config\]\]' ~/.codex/config.toml
49

$ ls ~/.codex/config.toml.bak* | wc -l
3             # all three pre-date this card; no backup was created, because
              # no write was attempted
```

The sha256 is byte-identical to the value measured before the run phase began
and to the safety copy taken immediately before this invocation.

## Baseline-attribution

The binary was compiled from this tree in this run (`go build ./cmd/moai`,
rc 0) — not `~/go/bin/moai`, whose lag against this tree is unknown and which
would not carry the new verb at all. Both the tree coordinate and the judging
build's coordinate are therefore the same commit, `e27f00b5a`.

The "49 of 49" figure agrees independently with three earlier measurements: the
raw `grep -c` on the file, the shape/existence split in
`baseline-measurement.md`, and the doctor's own `49 stale skill entries`.

## Gaps

- **The `--force` write path was NOT exercised against the live config**, and
  deliberately so: this card delivers the verb, and pruning this machine's 49
  entries would destroy the observation target cards t504 and t502 depend on.
  The write path's evidence remains the fixture tests and their backup+sha256
  assertions.
- Dry-run output was checked for its count and its summary line; the 49 listed
  paths were not compared one-by-one against
  `observed-skill-paths.txt`.
- Nothing was observed about Codex's own behaviour toward these entries.

## Residual risk

A verb that is correct in dry-run can still be wrong under `--force`; the two
paths share the eligibility judgment but not the write. What this run
establishes is the wiring and the judgment on real input — not the write.
