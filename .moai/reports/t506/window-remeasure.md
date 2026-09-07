# t506 — integration window: absorb, radius, remeasure

Window held as `lane-2`. Tree `.claude/worktrees/t506`, branch
`WT-codex-ghost-skills`. Date 2026-09-07.

## Tip re-read at window time

```
$ git fetch origin develop
$ git rev-parse --short origin/develop
0b1e27877
```

Matches the tip the lead named. Card HEAD before absorb: `905f0fc48`.

## Absorb

```
$ git merge --no-edit origin/develop
Auto-merging CHANGELOG.md
CONFLICT (content): Merge conflict in CHANGELOG.md
```

One conflict, resolved **both-keep**: exactly three lines were deleted (the
`<<<<<<<`, `=======`, `>>>>>>>` markers at lines 12, 15, 21), leaving both
sides' entries intact — this card's entry followed by the seven entries develop
brought in. Nothing was chosen between; a CHANGELOG slot collision is an
ordering question, not a selection one.

```
$ git diff --name-only --diff-filter=U | wc -l
0
$ git commit --no-edit   →   781d91dac
```

## Radius intersection, counted by name

```
$ git diff --name-only ace1c5440 905f0fc48 | sort > mine     # 27 files
$ git diff --name-only ace1c5440 origin/develop | sort > theirs   # 304 files
$ comm -12 mine theirs
CHANGELOG.md
                                                             # count: 1
```

**One overlapping file, and it is not a code file.** No `.go` file this card
touched was touched by develop in the same interval — the cards that ran in the
same packages today (t496, t498, t499, t501) landed in different files.

**A near-miss worth recording, because a wrong 0 looks exactly like a right
one.** The first attempt at this comparison used `e27f00b5a` as the card
endpoint and returned an intersection of **0**. That endpoint is the run-phase
commit; `CHANGELOG.md` was written later, in the sync commit, so the operand was
missing the very file that overlapped. The conflict git had just raised is what
exposed it — it is the control that proves this comparison is non-vacuous. A 0
computed against a truncated operand is indistinguishable, on its face, from a
0 that means no overlap.

## Remeasurement in the merged tree

All four commands run at `781d91dac`, in this tree, in this run.

```
$ go test -count=1 -timeout 900s ./internal/cli/... ./internal/codexwiring/...
rc=0
ok  github.com/modu-ai/moai-adk/internal/cli/update/merge   5.819s
ok  github.com/modu-ai/moai-adk/internal/cli/update/plan    3.582s
ok  github.com/modu-ai/moai-adk/internal/cli/update/report  2.876s
ok  github.com/modu-ai/moai-adk/internal/cli/wizard         4.975s
ok  github.com/modu-ai/moai-adk/internal/cli/worktree       8.761s
ok  github.com/modu-ai/moai-adk/internal/codexwiring        4.792s
FAIL lines: 0

$ go vet ./internal/cli/... ./internal/codexwiring/...
rc=0, 0 lines of output

$ go build ./...                              rc=0
$ GOOS=windows GOARCH=amd64 go build ./...    rc=0

$ golangci-lint run --timeout=5m ./internal/cli/... ./internal/codexwiring/...
rc=0
0 issues.
```

The full-suite verdict belongs to CI on the develop push; this batch is scoped
to the two packages the card touched, per the standing local-load discipline.

## The `~/.codex/config.toml.bak` question — settled by measurement

The lead read a `config.toml.bak` with today's mtime and asked whether this
card's tool wrote it. It did not, and it is **also not** this lane's safety
copy — the lead's own stated hypothesis. Three measurements settle it:

```
$ stat -f '%Sm' -t '%Y-%m-%d %H:%M:%S' .moai/reports/t506/baseline-measurement.md
2026-09-07 12:54:54          # this card's first artifact

$ ls -la ~/.codex/ | grep config.toml.bak
-rw------- 17212  Sep  7 12:20  config.toml.bak                    # 34 min EARLIER
-rw------- 15133  Aug 22 02:22  config.toml.bak-20260822-022202
-rw------- 16338  Sep  1 13:33  config.toml.bak-20260901-133347

$ shasum -a 256 ~/.codex/config.toml ~/.codex/config.toml.bak /tmp/t506-codex-config.safety.toml
9f6e3a95…ca33a  config.toml
56f8003d…6e9b   config.toml.bak                      # differs from live
9f6e3a95…ca33a  /tmp/t506-codex-config.safety.toml   # this lane's safety copy
```

The safety copy lives under `/tmp` — this lane wrote nothing inside
`~/.codex/`. The `.bak`'s content differs from the live file (630 lines /
17212 bytes vs 629 / 17275, both declaring 49 entries), so it is a snapshot of
an earlier state, and its name matches neither this card's backup format
(`config.toml.bak-<UTC>T…Z`) nor the existing dated one. Its mtime precedes
this card's first artifact by 34 minutes.

**What created it is not observed by this card**, and is left unclaimed rather
than guessed.

**Correction to this card's earlier wording, kept beside the original.**
The run and lane reports said *"new `.bak` 0 (no write was even attempted)"*
without saying what "new" was measured against. The accurate statement is
**"`.bak` files created by this card's tool: 0"** — the count was 3 before the
lane dry-run (run-phase measurement) and 3 after (lane measurement), and the
live file's sha256 was identical at three separate points across the card.

## Quality-gate output naming files that do not exist

A quality gate reported ruff errors in `t506-docs-insert.py`,
`t506-readme-row.py`, `t506-changelog.py`, and `t506-close.py`. None of the
four exists:

```
$ find . -name 't506-*.py' -not -path './.git/*' | wc -l
0
$ ls /tmp/t506-*.py            # zsh: no matches found
```

They were sync-phase scratch scripts, written and deleted outside the tree, and
they appear in no commit. No action was taken. Recorded here so the next reader
of that gate output does not go looking for files that were never in the
repository.

## Gaps

- The `--force` write path was not exercised against the live config, by
  design — cards t504 and t502 need those 49 entries.
- Coverage was not re-measured in the merged tree; the last figures are the run
  phase's (`internal/codexwiring` 89.5%, `internal/cli` 80.7% against an 85%
  target, a pre-existing package baseline this card neither lowered nor
  repaired).
- The docs-site pages were built, not read: `hugo` exiting 0 says the build
  succeeded, not that the new section renders as intended.
