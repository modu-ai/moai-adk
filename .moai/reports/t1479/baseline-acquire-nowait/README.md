# Baseline — `moai integration acquire` without `--wait`, live foreign holder (card t1479, D9)

Captured BEFORE any SPEC-MERGE-WINDOW-QUEUE-001 code change, and committed in its own commit
ahead of the run phase (verification-claim-integrity §2.3). AC-MWQ-010 (SPEC v0.5.0 numbering;
AC-MWQ-007 when captured, AC-MWQ-011 in v0.3.0-v0.4.0) compares against these files. Only this
reference was edited after capture; the captured bytes are unchanged since `3bc274dac`.

## Build under measurement

- Binary built from this branch's tree at HEAD `1e1d0cc84` (plan-only commits on top of develop
  `d7112d005`; no Go file differs from develop): `go build -o <scratch>/moai-plantree ./cmd/moai`,
  exit 0. Built without release ldflags, so `moai version` prints the default `v3.1.3`; the tree
  HEAD above is the provenance.

## Fixture

- A scratch git repository (`git init -b develop`) whose `.moai/state/integration-lock.json` is
  `record.json` here: holder `fixture-holder-A` / `lane-a`, `pid 1` (always live, so the holder
  reads as LIVE, never stale), `pid_source: session-owner`, fixed `acquired_at`.
- Environment: factory/kanban variables unset; `CLAUDE_PROJECT_DIR=<FIXTURE>`.

## Invocations and observations

| Form | Command (args after the binary) | exit | stdout | stderr |
|---|---|---|---|---|
| human | `integration acquire --session fixture-lane-B --name lane-b --card t0002 --branch develop` | `1` | empty (0 bytes) | `human.stderr` |
| json | same + `--json` | `1` | empty (0 bytes) — the refusal emits no JSON | `json.stderr` (byte-identical to `human.stderr` after normalization) |

- Record after both invocations: unchanged (`record.json` is the post-run file, content identical
  to the fixture as written).
- Side effect: `.moai/state/integration-mutation.lock` is created by the mutation section.

## Normalization (applied to stderr only)

- `^time=\S+ ` → `time=<T> ` (the slog WARN timestamp).
- the scratch fixture path → `<FIXTURE>`.

The pid and the `since` timestamp come from the fixture record and are therefore already fixed; no
other normalization was needed. The run phase's AC-MWQ-010 test applies the same two rules.
