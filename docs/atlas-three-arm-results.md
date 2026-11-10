# C10b three-arm measurement (#2186)

Measured 2026-11-10 UTC on macOS arm64, product base
`05e652848dc9be17f5f298d32167b5eb732b783a` (#2185). The runner and
integration test are this Issue's additions. This is a controlled cumulative
augmentation experiment, with no model or task-dependent query planning.

## Protocol

`tools/atlas_three_arm.py --binary <moai-cli> --output-dir <directory>`
materializes C10a's frozen source in a temporary git repository and builds one
index shared by all arms. Child processes receive a synthetic home and an
allowlist of process environment variables. No installed home or credentials
are used. Fixture git config is isolated; installed repository hooks are unchanged.

- `rg`: `search` with only the task's literal pattern, matching the pinned rg file sets.
- `+Map`: lexical search followed by `atlas-map` selecting those hit paths.
- `+Graph`: the same two stages, followed by `atlas-graph` on those file node ids,
  with three hops and a 5000-edge cap.

Only query patterns and preceding operation results enter queries; ground truth
is used exclusively by the judge. All hits must remain inside the snapshot.
Every operation must have the same generation identity, fresh state and no
omissions. The runner refuses unavailable, partial, stale or truncated answers.
T-033 has no lexical seed, so its Map/Graph calls are skipped: empty Map paths
would select the entire corpus. It remains one of the 42 scored tasks.

## Observed results

All 42 scored tasks (33 RETRIEVAL, 9 GRAPH) run in every arm. C10a's unchanged
judge uses per-suite macro means and inclusive floors. Machine-readable per-task
reports, identity digests and summaries are in [evidence/atlas-c10b](evidence/atlas-c10b/summary.json).

| Arm | Precision mean | Recall mean | GRAPH file coverage | Judge | File citations | Response bytes | Op calls |
| --- | ---: | ---: | ---: | --- | ---: | ---: | ---: |
| rg | 0.9212 | 0.9697 | 1.0000 | passed | 140 | 119144 | 42 |
| +Map | 0.9212 | 0.9697 | 1.0000 | passed | 140 | 235496 | 83 |
| +Graph | 0.9212 | 0.9697 | 1.0000 | passed | 140 | 364752 | 124 |

Op counts exclude the one shared index build and the judge. Search runs 42 times
per arm; Map runs 41 times in each augmented arm; Graph runs 41 times in the last
arm. Map/Graph return records (85/170 cumulative view edges across tasks), but
add no file hits to this lexical-seeded experiment. Equal scores therefore show
no observed retrieval improvement. The test also asserts the invocation counts:
score equality alone cannot establish that an augmented stage ran.

Response bytes count the actual complete JSON stdout, including its envelope;
they are **not model tokens**. File citations are unique task/file pairs, not
span counts. Tokens are `not_run`. Each arm observed 376 fresh lexical spans,
with snapshot stale ratio 0. This is an unmodified-snapshot observation, not a
FRESHNESS mutation result: the eight mutation cases remain #2407 `not_run` here.

GRAPH's 1.0 is **relation-file coverage only**, as defined by C10a; the lexical
hit contains the expected referring file. It does not establish that an edge
from the referring file to its target exists. #2185's reported code-reference
producer gap remains HOLD. This experiment does not resolve it, redefine the
fixtures, or claim semantic graph correctness. The `rg` arm is the product's
lexical search verified equal to rg, not a claim that the product invokes rg.

## Source verification

Adopted Python's documented `subprocess.run` argv, `check=True`, captured binary
stdout and explicit `env` mapping. The runner executed under Python 3.9.6 and
3.14.2. Official source successfully retrieved/read 2026-11-10T23:27:44Z:
<https://docs.python.org/3.9/library/subprocess.html>, SHA-256
`d1f230b0f6349a6b0bfe84297afe7adc144ae211b43f3bea4a2fa30f53e2b65f`.
Shell command interpolation and token estimates were not adopted. C10a's source,
corpus and thresholds were read directly and remain unchanged. The original
Core Brain was not independently re-read; the recorded C10a protocol is used.

## Local verification

- `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR="$PWD/target" cargo test --offline --locked --no-fail-fast`: exit 0 (full suite).
- `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR="$PWD/target" cargo clippy --offline --locked --all-targets -- -D warnings`: exit 0.
- `cargo fmt --check`, `git diff --check`, `git diff --cached --check`: exit 0.
- `PATH=/opt/homebrew/bin:$PATH PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'`: exit 0, 176 tests, Python 3.14.2.
- Final restored `atlas_three_arm` test executable: exit 0, 1 passed (Python 3.14.2).
- Mutations: replace lexical pattern, omit Map stage, omit Graph stage: all killed
  by the integration test (exit 101); original bytes restored, final test passed.

The default Python 3.9 run failed five existing `flow_metrics` tests because
`datetime.fromisoformat` rejects `Z`. Python 3.13's run lacked PyYAML (two import
errors). No unrelated code or installed packages were changed; Python 3.14's
complete suite passed. Exact-head rc-ci, Linux/Windows runs and independent
review remain `not_run` until the leader requests a PR/audit.
