---
id: SPEC-20261120-mflow-measure
title: "A load probe for controlled before/after runs of the M-Flow pipeline"
status: draft
revision: 1
created: 2026-11-20
---

## Intent

The user instruction (relayed through the card): repair every M-Flow audit/merge
bottleneck by cause and prove it with a controlled before/after measurement. This card's
slice is the measuring harness itself and the "before" runs: one probe that starts a target
command, samples the machine while it runs, and reports the numbers the instruction names —
so the before runs (in a low-load window) and the after runs (after the repairs land) are
the same measurement on the same Mac, comparable line by line.

Survey (read-only, 2026-11-20): the machine has 16 logical cores; the field observation
(load 121–154 against it) is a reference only. Heavy-slot queues have run 10–15 deep and an
independent audit spends 30–60 minutes mostly cold-building. macOS reports every needed
number without new dependencies: `sysctl -n vm.loadavg` (the 1-minute load),
`sysctl vm.swapusage` (swap in/out), `sysctl kern.memorystatus_vm_pressure_level` (the
kernel memory-pressure level), `vm_stat` (page counts), and `ps aux` (per-process CPU and
the running compile/test process names). The probe itself runs as `python3 -I` with the
standard library only — the same shape the session-check measurement used (card C-182's
probe).

## Non-goals

- Killing, pausing or reprioritising any other lane's work on the shared Mac, and bypassing
  any setting, slot limit or protection: the probe observes only (the instruction forbids it
  explicitly).
- The repairs themselves (C-207's warm start, heavy-slot caps, C-192's time-bound tests,
  the merge queue cards): this card measures before and after, it changes nothing.
- CI or Linux support: the sources are macOS (`sysctl`, `vm_stat`); on another OS the probe
  still runs the command and reports the exit, with the platform-specific fields null and
  one line saying the sampler had no macOS sources.
- A repository-wide benchmark suite: the harness takes one target command per invocation.

## Requirements

- REQ-1: While the probe runs a target command, it shall sample the machine until the command ends and report, in one JSON answer, the 1-minute load median and max, the sampled CPU user and system percentages with the utilisation, the memory-pressure level and the swap used, the peak count of concurrently running compile and test processes, the elapsed wall time, and how the command ended — exit code, or the signal that killed it, kept separate from a product failure. · failure: none — pinned by AC-1.
- REQ-2: While the probe answers, it shall record the run label, the UTC and KST times, the machine and core count, the git, rust and installed CLI versions, and the exact base and head the caller names, so a before row and an after row are comparable without other notes. · failure: none — pinned by AC-2.
- REQ-3: While the probe starts, it shall compare the current 1-minute load against the caller's low-load threshold and mark the answer with a warning field when the window is above it, so a run made outside a controlled window says so in the row itself. · failure: none — pinned by AC-3.
- REQ-4: While the probe writes its detail file (`--out`), it shall keep the full sample series and the answer in one JSON document the reports table is built from, without reading or writing anything under the real HOME except the caller-named report folder. · failure: none — pinned by AC-4.

## Acceptance

- AC-1 (REQ-1): a stub target (`python3 -c "print(7); raise SystemExit(3)"`) — the answer carries the exit code 3, an elapsed time above zero, load samples with a median and a max, and a peak process count — all asserted from the JSON, not from time.
  verify: `python3 -m unittest tests.mflow_probe_test`
- AC-2 (REQ-2): the caller passes a label, a base and a head — the answer names them beside the UTC and KST stamps and the machine facts (`sysctl hw.ncpu`, `git --version`, `rustc --version`, the installed CLI's `--version` line read from PATH).
  verify: `python3 -m unittest tests.mflow_probe_test`
- AC-3 (REQ-3): with the threshold set under the current load the answer carries `window: "above"` with the observed load; with a threshold above every sample it carries `window: "ok"` — both stub runs in one test module.
  verify: `python3 -m unittest tests.mflow_probe_test`
- AC-4 (REQ-1, REQ-4): a stub target that kills itself with SIGTERM reports `ended: "SIGTERM"` — separate from the exit-code field — and the `--out` document holds the sample series the summary was derived from (sample count equals the series length).
  verify: `python3 -m unittest tests.mflow_probe_test`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4): neighbours — the whole module passes under `python3 -m unittest` and the built binary's `spec check` accepts this SPEC; the probe runs no other repo check, so no scan entry is added (the unittest discover already runs the test module).
  verify: `python3 -m unittest tests.mflow_probe_test`; `moai spec check SPEC-20261120-mflow-measure`

## Scope

- write: `tools/mflow_probe.py`, `tests/mflow_probe_test.py`, this SPEC folder
- read-only: the reports folder under the user's MoAI home (`moai home` names it — `reports/moai-adk/` below it; the before/after tables are written there, outside the repository), `tools/local_check.sh` (the command set the before runs wrap)
- reserved by other cards: none of the audited files — the probe observes the pipeline, it does not edit it
- forbidden: killing or reprioritising other work, bypassing slot limits or protections, writing under the real HOME outside the caller-named report path, Linux or CI support

## Plan

1. Commit this SPEC.
2. RED: the stub-command tests (exit, elapsed, samples, label and stamps, window field, SIGTERM split, the `--out` document) against a probe that does not exist yet.
3. GREEN: `tools/mflow_probe.py` — the sampling loop (1-second default, `--interval`), the macOS sources with a graceful non-macOS answer, the JSON answer on stdout and the full document at `--out`.
4. The before runs: three probe runs of the representative audit/merge task set in a low-load window, after the C-197 audit lands — each run's row into `~/.moai/reports/moai-adk/mflow-before-after.md`; the run count and any reason for fewer runs recorded in the same table.
5. The after runs stay with the repairs' cards (C-207 and the others) — the probe and the table format are what this card ships.

## Risks

- The sampler adds its own load: one `ps aux` and a few `sysctl` calls a second is small next to a cargo build, and `--interval` can widen it; the report records the interval so a reader can judge the observer effect.
- A low-load window may not open while lanes are merging (the field observation): the before table then carries fewer runs with the recorded reason — the card allows "record count and reason if fewer runs are safe".
- The macOS sources change shape across versions: each parser answers null when its source is unreadable rather than guessing, so a row never shows a fabricated number.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | The probe is one Python file on the standard library, run as `python3 -I tools/mflow_probe.py` | The C-182 measurement's shape; no new dependency, no build, runnable from any worktree of the repo | A Rust subcommand behind the CLI gate |
| D2 | The before runs wrap the representative task set the audit/merge pipeline runs today, unmodified | The measurement must not bypass settings or protections; the before rows must describe today's pipeline | A tailored faster task set for the before rows |
| D3 | Signal endings are their own field, never folded into an exit code | The instruction separates SIGTERM/timeouts from product failures | A single exit-code field |

## Result

구현 에이전트가 채워요 — 이 커밋은 SPEC 과 계획까지예요.
