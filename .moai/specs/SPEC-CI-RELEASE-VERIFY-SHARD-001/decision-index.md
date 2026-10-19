# decision-index.md — SPEC-CI-RELEASE-VERIFY-SHARD-001

Stateless on the status axis — the SPEC lifecycle lives in `spec.md` alone.
Authored at plan phase (card t1892). `interview.decision_gate: on` (`.moai/config/sections/interview.yaml`).
No row carries an authority anchor that resolves in the committed tree for this question, so every row routes `FOUNDER` and is never relabeled. All rows are implementation-level: none changes a shipped command's default user-visible behavior, removes a user-facing feature, or changes a template default (the touched files are the repository's own CI workflow, its maintainer release script, and repo tests; none has a template mirror). Each row's `Default:` is the published Default rule applied mechanically (rule order: preserves current behavior; else single-revert undo; else smaller user-visible surface), recorded as DEFAULT-APPLIED at plan close.

### Q1: Which job structure carries the 3 OS x 6 shard legs — one matrix job with the shard dimension, or per-OS shard jobs with per-OS aggregator jobs?

Label: FOUNDER

Class: implementation-level

Why unresolved: both structures run the same legs, and each keeps a different consumer intact. One matrix job keeps the required gate's text, its `needs` list, the job id `full-matrix-test` (read by the Python verifier at `scripts/ci-mirror/test/verifier-fail-closed_test.py:21`), and the filter-parity test untouched, but renames the legs and so needs `scripts/release.sh` updated (Q2). Per-OS shard jobs with aggregators named exactly `Release Verify (<os>)` would keep `release.sh` untouched, but a matrix job's `needs` result cannot be split per OS, so it forces three separate job definitions and a rewrite of the gate's `needs`/aggregation text — the surface the operator forbade relaxing.

Authority anchor: none in the committed register (research.md §C.4 carries the analysis; it is a card artifact, not authority).

Default: one matrix job `full-matrix-test` with a `shard` dimension and per-shard `include` entries (rule: preserves current behavior — the required gate, its `needs`, and the job id stay byte-identical, and the gate is the surface the operator named untouchable)

Alternate: three per-OS shard jobs plus three aggregator jobs named `Release Verify (<os>)`, gate `needs` rewritten

Justification: the Default rule's first clause is ambiguous here (each option preserves a different consumer), so the tie is broken toward the consumer the operator protected by name (the gate); the cost is paid in Q2, a repo-local script and its test. If the operator prefers to leave `release.sh` untouched, choose the Alternate and accept a gate-text rewrite.

Operator verdict: DEFAULT-APPLIED 2026-10-19T02:12Z lane-40/manager-spec

### Q2: How should `scripts/release.sh --require-matrix-run` read the sharded leg names?

Label: FOUNDER

Class: implementation-level

Why unresolved: the script finds each leg by exact name (`Release Verify (ubuntu-latest)`, line 360) and fails closed when none matches (measured: `shard-suffixed names -> <no leg>`, research.md L-E7). With Q1's Default the names become `Release Verify (<os>, <shard>)`, so the script must change or the flag stops working.

Authority anchor: none in the committed register.

Default: for each of the three OS values require exactly the six declared shard ids, once each, all `success` (missing on one or all OS, unknown name, duplicate, or non-success withholds the gate); extend the Go gh-double scenario table with the sharded all-green case and the partial cases (rule: preserves current behavior — the contract "no release unless every leg of all three OS succeeded on the dispatched commit" is kept, and only the name matching widens)

Alternate: leave `release.sh` unchanged and let `--require-matrix-run` fail closed on sharded names until a follow-up card repairs it

Justification: the Alternate breaks the flag for every sharded run, so the current contract is not preserved under it. The Default stays fail-closed on any missing or non-success shard.

Operator verdict: DEFAULT-APPLIED 2026-10-19T02:12Z lane-40/manager-spec

### Q3: Where should the shard selectors live, and what keeps the two workflows from drifting?

Label: FOUNDER

Class: implementation-level

Why unresolved: the selectors exist today only in `ci.yml`. The release workflow can carry its own copy guarded by a failing test, or both workflows can consume one shared source (a script or a composite action), which requires editing `ci.yml`, a workflow that carries required-check jobs, and adding a shared surface that must run under macOS bash 3.2 and Git Bash.

Authority anchor: none in the committed register.

Default: selectors inline in the release workflow's matrix `include` entries; `ci.yml` untouched; a new repository test under `internal/template/` compares the two workflows' selectors and remainder exclusion list and checks exact cover over the real `internal/cli` top-level test names the go tool lists under the shards' build conditions (`go test -race -list`) (rule: preserves current behavior — `ci.yml` is not edited and no shared runtime surface is added)

Alternate: extract the selector table into one shared script or composite action used by both workflows, keeping the same parity test as a guard

Justification: the Default converts drift from a silent hazard into a failing test at the cost of one duplicated table; the Alternate is a pure refactor that stays available behind the same test.

Operator verdict: DEFAULT-APPLIED 2026-10-19T02:12Z lane-40/manager-spec

### Q4: Which timeout ceilings do the shard legs use?

Label: FOUNDER

Class: implementation-level

Why unresolved: `ci.yml` uses per-shard ceilings (`-timeout` 20m, 25m, 15m; job 30, 35, 30 minutes) sized at about 2x ubuntu proxies. The release legs also run on macOS and Windows, whose per-shard durations are unmeasured (research.md gap G2).

Authority anchor: none in the committed register. The operator's "no timeout increase" bounds the answer from above (nothing may exceed the current 35m / 40) but does not choose between keeping the current values and mirroring `ci.yml`'s lower ones.

Default: keep the current release ceilings on every shard leg — `-timeout 35m` per test binary and `timeout-minutes: 40` per job (rule: preserves current behavior; not an increase)

Alternate: mirror `ci.yml`'s per-shard ceilings, giving earlier failure of a hung shard but a higher chance of a false failure on a slower OS

Justification: lower ceilings on unmeasured slower OSes would turn a slow-but-passing shard into a blocking failure with no evidence to size them; the first live run produces that evidence.

Operator verdict: DEFAULT-APPLIED 2026-10-19T02:12Z lane-40/manager-spec

### Q5: Should the release shard legs wrap the test command in `scripts/ci/retry-flaky.sh`?

Label: FOUNDER

Class: implementation-level

Why unresolved: every `ci.yml` race shard wraps the test command in the retry helper; the release legs run it on none today. The helper is bash with no bash-4-only construct (research.md L-E12) but has only ever executed on ubuntu, its registry is empty, and once an entry is added it retries a failed test on a leg that feeds a required gate.

Authority anchor: none in the committed register.

Default: do not wrap; release shard legs keep a single attempt with no retry (rule: preserves current behavior — the release gate has no retry today, and the operator forbade gate relaxation)

Alternate: wrap each shard as `ci.yml` does, for command-shape parity

Justification: command-shape parity buys nothing while the registry is empty and adds an unexercised layer on macOS and Windows; a retry on a required gate is a relaxation the operator named.

Operator verdict: DEFAULT-APPLIED 2026-10-19T02:12Z lane-40/manager-spec

### Q6: On which leg does `go vet ./...` run?

Label: FOUNDER

Class: implementation-level

Why unresolved: the base runs vet once per OS. With six legs per OS, leaving the step unconditional runs the same whole-repo vet six times per OS (18 total), a duplicate run the cost constraint forbids; running it on one leg per OS needs a choice of which leg.

Authority anchor: none in the committed register.

Default: run vet once per OS, on the lightest ubuntu shard by measured proxy (`cli-2`, 408 s), selected by a matrix flag (rule: preserves current behavior — vet coverage stays once per OS, and no duplicate run is added)

Alternate: run vet on every shard leg

Justification: a vet failure still fails the leg and therefore the gate; the Default avoids 15 redundant vet runs. A vet failure on that leg skips that leg's tests while the other shards still run; the gate fails either way.

Operator verdict: DEFAULT-APPLIED 2026-10-19T02:12Z lane-40/manager-spec

### Q7: Do the windows-only ancestor-`.moai` diagnostic steps run on every windows shard leg?

Label: FOUNDER

Class: implementation-level

Why unresolved: the temporary t169 diagnostic steps show where an ancestor `.moai/state` directory comes from on the windows runner. The instruction is to preserve them; with six windows legs they either print six times or are restricted to one leg.

Authority anchor: none in the committed register.

Default: keep both steps unchanged, gated only on `runner.os == 'Windows'`, so they run on every windows shard leg (rule: preserves current behavior — the steps and their conditions are untouched)

Alternate: restrict them to a single windows shard to cut log volume

Justification: the failing test (internal/hook, internal/cli) can land on any shard, so the diagnostic must be present wherever it can fail; the output volume is a bounded 60-line listing per step.

Operator verdict: DEFAULT-APPLIED 2026-10-19T02:12Z lane-40/manager-spec

### Q8: If a macOS or Windows shard overruns its ceiling on the first live run, what is the response?

Label: FOUNDER

Class: implementation-level

Why unresolved: per-shard macOS and Windows durations are unmeasured (research.md gap G2). By the SPEC's own arithmetic the heaviest shard (rosterguard, 782 s on ubuntu) could land at roughly 26-39 minutes on a 2x-3x slower OS against the unchanged 35-minute test ceiling and 40-minute job limit, so the first live run can still fail for the reason the SPEC exists to remove. The operator fixed shard count and forbade a timeout increase, so the response to an overrun is the open choice.

Authority anchor: none in the committed register.

Default: record the overrun and open a rebalance card that splits the overrunning shard by selector; shard count and ceilings stay as shipped and no timeout is raised (rule: preserves current behavior — the operator-fixed shard count and ceilings are unchanged)

Alternate: take one measurement before landing with a `workflow_dispatch` of the release workflow on the candidate ref (one added run, not a duplicate matrix) and pre-split the heaviest shard if it shows an overrun

Justification: the Default follows the operator constraints as written and uses the first live run as the measurement (AC-RVS-016). The Alternate trades one extra run for earlier knowledge; whether to take it is flagged for the leader and is not a SPEC edit.

Operator verdict: DEFAULT-APPLIED 2026-10-19T02:35Z lane-40/manager-spec
