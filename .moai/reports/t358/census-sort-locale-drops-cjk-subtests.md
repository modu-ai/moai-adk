# The census dedup is locale-dependent and silently drops CJK subtest names

Card t358 · SPEC-CI-TEST-OBSERVABILITY-001 · found in **sync-phase**, 2026-08-30,
while running the AC-CTO-005a re-parse check against the downloaded artifact.

**Not fixed.** Sync-phase does not modify the run-phase deliverable, and the
ONE-run constraint means the fix could not be re-observed in CI on this card.
This file records the observation and stops there. Card issuance is the
operator's call; the lead carries it to that list.

---

## Claim

`scripts/ci-census/test-census.sh` pipes every total and every listed row through
`sort -u` with no `LC_ALL` pin, so deduplication follows whatever collation the
calling environment happens to have. Under a UTF-8 collation locale, byte-distinct
lines whose only difference is CJK text can collate **equal**, and `sort -u` drops
all but one of them.

The same stream therefore yields different census totals on different machines,
with no signal that anything was dropped.

## Evidence

One input: the `test-stream.json.gz` uploaded by job `99248050305`
(ubuntu-latest) of run
`https://github.com/modu-ai/moai-adk/actions/runs/33308057570`, downloaded and
decompressed to 23,256,287 B.

Raw event count, no dedup at all — the ceiling any correct dedup must reach:

```
$ jq -r 'select(.Action=="pass" and (.Test // null) != null) | "\(.Package)\t\(.Test)"' stream.json | wc -l
   19513
```

The census, run twice on that identical file, differing only in collation:

```
$ bash scripts/ci-census/test-census.sh stream.json | tail -1
=== totals: packages=136 passed=19507 skipped=100 nothing-ran=3 failed=0 build-failed=0 ===

$ LC_ALL=C bash scripts/ci-census/test-census.sh stream.json | tail -1
=== totals: packages=136 passed=19513 skipped=100 nothing-ran=3 failed=0 build-failed=0 ===
```

`LC_ALL=C` reproduces the runner's console figure exactly (`passed=19513`, read
from the job log at line 753). The ambient locale loses **six**.

The six, isolated by C-collated `comm` between the two dedup results:

```
github.com/modu-ai/moai-adk/internal/cli          TestDetectPromptLang/こんにちは
github.com/modu-ai/moai-adk/internal/cli          TestDetectPromptLang/カタカナ
github.com/modu-ai/moai-adk/internal/mx           TestExtractSpecIDs/여러_SPEC_ID
github.com/modu-ai/moai-adk/internal/mx           TestExtractSpecIDs/중복_SPEC_ID
github.com/modu-ai/moai-adk/internal/mx           TestExtractSpecIDs_Patterns/여러_SPEC_ID
github.com/modu-ai/moai-adk/internal/tui/internal TestFillRight/已知
```

All six are subtests whose names are CJK. The reverse direction is empty — the
UTF-8 run contains no line the C run lacks — so this is strictly loss, not a
different partition.

## Baseline-attribution

Measured in the sync phase, in worktree `.claude/worktrees/t358`, against tree
`451690025`, using the census as committed on `WT-ci-test-observability`.
Host: darwin/arm64, `LANG=en_US.UTF-8`, `LC_COLLATE=en_US.UTF-8`, BSD `sort`.

The runner figure it is compared against is not a re-derivation: it is the
verbatim `=== totals:` line printed by job `99248050305` in run `33308057570`.

## Blast radius

**Six** call sites in `test-census.sh` use bare `sort -u` — every dedup in the
script, without exception (measured: `grep -n 'sort -u' scripts/ci-census/test-census.sh`
→ lines 81, 95, 118, 119, 142, 145):

| Line | What it dedups | What loss looks like |
|---|---|---|
| 81 | `BUILD FAILED` import paths | a build failure not printed |
| 95 | `FAILED` package+test pairs | a failing test not named |
| 118, 119 | packages for the `FAILED PKG` split | a package-level failure misclassified |
| 142 | the `SKIPPED TEST` / `NOTHING RAN` row listing | a skipped test or a zero-test package not printed |
| 145 | the `count()` helper — `packages`, `passed`, `skipped`, `nothing-ran`, `failed`, `build-failed` | every number in the totals line |

Line 142 is the one that matters most. A skipped CJK-named subtest colliding with
a sibling would be dropped from the printed census — and identifying a skipped
test by name from CI artifacts is the entire purpose of this SPEC.

The failure rows (81, 95, 118, 119) are the more alarming class in principle,
since a dropped row there hides a failure rather than a skip. In practice a
dropped `FAILED` row does not turn a job green — the exit code is captured from
`go test`, not from the census — so the loss would be one of readability, not of
the red signal itself.

## Is CI affected today?

**No, on the evidence available.** All three runners in run `33308057570`
produced C-locale-consistent figures: ubuntu `passed=19513` matches the raw event
count exactly, and macOS `19512` / windows `18854` are consistent with their
respective failure counts rather than with silent drops. GitHub's ubuntu images
default to `LANG=C.UTF-8`, whose collation is byte-order.

So this is **latent**, not active. What makes it worth a card anyway is the
failure shape: a wrong number with no signal — the same class of defect this SPEC
was written to eliminate, reproduced inside the instrument built to detect it.

## Gaps — what was NOT observed

- **Only the ubuntu artifact was downloaded.** The macOS and windows streams were
  not decompressed, so the locale behaviour was not re-derived from their data.
- **The runners' locale was not read directly.** No `locale` command was run in
  CI; the conclusion "runners behave as C" is inferred from the totals matching
  the raw event count, not from an environment dump.
- **`skipped` / `nothing-ran` / `failed` were not separately probed** for
  collation loss on this stream — they happened to be identical across both runs
  here, which shows those particular values were unaffected, not that they are
  immune.
- **No fix was written or tested.** `LC_ALL=C` on the `sort` calls is the obvious
  candidate and is not verified.

## Residual risk

If a runner image ever ships a different default locale, or a developer runs the
census locally to reproduce a CI figure, the numbers diverge silently and the
divergence will read as a real change in the suite. The repo has many
Korean-named subtests, so the collision surface is not hypothetical — six were
found in a single stream.
