# t1492 verdict — develop CI "Race Test" red (internal/cli -race 1200s timeout)

Card t1492 · Class B (cause measured first, no SPEC) · branch `WT-race-cli-timeout` · base develop `948d444b9`.

## Claim

1. The red `Race Test` is not flaky and not one slow test. `internal/cli` under `-race` has no headroom under the `-timeout 20m` cap, so any slow runner or added test pushes it over.
2. Minimal fix: raise the `test-race` job to `-timeout 35m` and `timeout-minutes: 40` in `.github/workflows/ci.yml`. This is a ceiling raise, not a structural fix.
3. Whether the fix turns CI green is NOT observed (see Gaps).

## Evidence

Failing run `37161322193` (head `30ce3a02d`), from `gh run view 37161322193 --log-failed`:

```
FAILED PKG    github.com/modu-ai/moai-adk/internal/cli
  FAIL	github.com/modu-ai/moai-adk/internal/cli	1200.344s
```

Workflow line (before): `go test -json -race -count=1 -timeout 20m ./...` with `timeout-minutes: 25`.

Recurrence — Race Test wall time per develop run (scratch script `race-durations.sh`, output verbatim):

```
37161322193 failure 1391s
37158038465 failure 1390s
37155531103 failure 1387s
37149225709 failure 1394s
37147187709 failure 1394s
37139467432 failure 1387s
37135148076 success 1257s
37133567019 failure 1383s
37122965651 failure 1393s
37121390995 failure 1342s
```

Run `37139467432` (read directly) fails identically: `FAIL github.com/modu-ai/moai-adk/internal/cli 1200.388s`. The other failing runs have the same ~1390s wall time (1200s test cap plus setup); I did not read their logs one by one.

Per-test breakdown from the uploaded `test-stream-ci-race-ubuntu-latest` artifacts of the red run and the last green run `37135148076` (scratch scripts `analyze.py`, `groups.py`, `diff.py`):

```
RED  37161322193: pkg elapsed 1200.344s, top-level finished 3948, sum 1197.4s, t.Parallel paused 276 (0 finished)
GREEN 37135148076: pkg elapsed 1127.825s, top-level finished 5012, sum 1134.4s, t.Parallel paused 422
common tests: 3883   GREEN sum 716.1s   RED sum 1023.9s   (+43%)
tests only in RED (finished): 65, sum 173.5s
tests >=1s: 244 tests, 981.4s ; tests <1s: 3704 tests, 216.0s
slowest single test: TestFactoryEnsureCardWorktreeConcurrentRealMaterializer 34.1s
family totals: Factory 316.4s (156 tests), Managed 103.2s, QAS 81.6s, Codex 74.0s
```

Reading: no single dominant test. The last green run already used 1127.8s of the 1200s cap (94%). The red run lost headroom two ways: the same 3883 tests ran 43% slower on the hosted runner, and new tests added about 173s. The 276 `t.Parallel` tests never started, so the true red runtime is above 1200s.

Fix applied (`git diff`, `.github/workflows/ci.yml`): `timeout-minutes: 25 -> 40`; `-timeout 20m -> 35m`; the comment now cites the measured figures above. Sizing estimate (an estimate, not a measurement): (1128 + 173) x 1.43 = about 1860s = 31m, inside 35m with about 13% margin.

Static checks on the edited file:

```
python3 yaml.safe_load: parsed ok; test-race timeout-minutes = 40
actionlint .github/workflows/ci.yml: exit=0 (no output)
```

## Baseline-attribution

All timings come from the two CI artifacts/logs named above (this run's downloads, not carried over from memory). The workflow edit was checked on tree `WT-race-cli-timeout` after base develop `948d444b9`. The failing run's head is `30ce3a02d`, a different commit from the card base; the fix is in the CI file only, so the base difference does not change the diagnosis. Tool provenance: `actionlint` and `python3` are installed binaries; no moai build was used for any measurement.

## Gaps

- NOT observed: a green `Race Test` after the change. It can only be read from a CI run on a pushed head, and push belongs to the leader. No local `go test -race ./internal/cli` was run: it is a heavy package suite and this lane did not take a `moai slot` lease for it.
- NOT measured: how long the full `internal/cli` race run takes now (the red run never finished its parallel tests). The 31m figure is an estimate built from one +43% variance sample.
- NOT verified: causes of the +43% runner variance and of the +173s of new tests beyond the family totals above.
- NOT checked: whether the 3 other failing runs I did not open share this exact cause (only the wall time matches).

## Residual-risk

- Ceiling raise only. The suite keeps growing; the structural fix is to shard `internal/cli` across two jobs or cut Factory lease/stall tests that wait in real time. Recommend a follow-up card.
- `.github/workflows/release-pr-multi-os.yml:210` runs the same `go test -race ./...` with `-timeout 25m` on the required `Release PR Multi-OS Gate` ubuntu leg. By the same arithmetic it can hit the same wall. I did not change it (out of this card's scope; changing a required gate needs the leader's call) and did not measure it.
- A 40-minute advisory job is slower feedback; it is non-required, so merges are not blocked by it.
- Scratch files in `.moai/reports/t1492/` (`stream/`, `stream-green/`, `*.py`, `*.sh`, logs) stay uncommitted; their loss costs only re-download time (`gh run download`).
