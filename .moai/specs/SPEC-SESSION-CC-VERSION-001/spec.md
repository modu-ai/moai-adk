---
id: SPEC-SESSION-CC-VERSION-001
title: "Running-binary staleness visibility and the launcher-mediated --resume emergency path"
version: "0.1.0"
status: draft
created: 2026-10-04
updated: 2026-10-04
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/session
lifecycle: spec-anchored
tags: "session, doctor, cc-version, staleness, launcher, resume, factory-lane"
era: V3R6
tier: M
related_specs: [SPEC-FACTORY-SELF-DISPATCH-001, SPEC-V3R6-MULTI-SESSION-COORD-001]
---

# SPEC-SESSION-CC-VERSION-001 — Running-binary staleness visibility and the --resume emergency path

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-04 | Initial draft. Split from the t1348 investigation verdict (`.moai/reports/t1348/verdict.md` § "MoAI가 자동화할 수 있는 것"), items ② and ④ only; items ① and ③ are absorbed by card t1482 and are excluded here. |

## §A Background

Every claim below was re-verified in this tree (`WT-session-cc-version` at `30ce3a02d`, cut from
develop) on 2026-10-04. The t1348 verdict was measured on an older tree, so nothing is carried
over without re-reading the code it names.

### A.1 The defect this visibility closes

A running Claude Code process never picks up an updated binary — the update takes effect at the
next process start (t1348 §2.1: official docs plus machine measurement). That verdict also
measured this machine on 2026-10-03: of 31 live `claude` processes, only 4 ran 2.1.288 (the
version the `~/.local/bin/claude` symlink targeted) while one had been pinned to 2.1.281 for
eight days. A long-lived factory lane can therefore fall behind its own leader invisibly; the
symptoms (a missing flag, a changed hook-input field, a messaging behavior difference) surface
as cross-session misbehavior, never as a version report.

MoAI's surfaces that could carry that report carry none of it:

| Claim | Evidence (this tree) |
|---|---|
| The session registry record carries no version | `internal/session/registry.go:120-129` — `Entry` marshals exactly `session_id`, `spec_id`, `phase`, `started_at`, `last_heartbeat`, `pid`, `host`, `cwd`; `moai session list --json` marshals these entries directly (`internal/cli/session.go:148`) |
| No process-version reader exists anywhere | `grep -rn "lsof" internal/session/ --include="*.go"` → 0 hits; the repo's only lsof uses are cwd discovery (`internal/discovery/leader_readers_darwin.go:72`) and port lookup (`internal/cli/web_port.go:69`) |
| `moai doctor` reports only the installed version, one row per host | `internal/cli/doctor.go:449` `checkClaudeCode` reads `CLAUDE_CODE_VERSION` or execs `claude --version` — no per-session running version |

### A.2 The reads are cheap, and this machine already proved the shape works

t1348 §2.2 measured the probe shape live: `lsof -a -d txt -p <pid>` output carries the running
binary's path, and `grep -oE '(versions/|claude-code/)[0-9.]+'` over it recovers the version.
The Linux twin reads `os.Readlink("/proc/<pid>/exe")` over the same path shapes. The installed
side needs no process at all: the `claude` binary found on PATH (typically the
`~/.local/bin/claude` symlink) resolves to a path whose version segment **is** the installed
version. Each read degrades independently — a dead pid, an unreadable mapping, an npm-style
install with no versioned path, an unsupported platform — and every degradation means
`unknown`, never an error.

### A.3 The relaunch loop and the resume conflict, read in this tree

The relaunch clear policy (`--clear-policy relaunch`) turns the lane launcher into a supervising
loop: `runFactoryLaneRelaunch` (`internal/cli/factory_lane_relaunch.go:57`) resolves the `claude`
binary once, then per iteration re-enters the lane-join gate, leases the next card, ensures its
worktree, and starts one interactive session **in that card worktree** with the launcher-built
`claudeArgs` (`factory_lane_relaunch.go:115`). Three facts decide item ④'s shape:

1. The one-shot lane join (the default `clear-each` policy) already passes unknown tokens
   through to the child. `moai cc` runs with `DisableFlagParsing: true`
   (`internal/cli/cc.go:107`) and its launcher-side scanners (`parseProfileFlag`,
   `parseLauncherEntry`, `parseFactoryLaneLabel`, the worktree handlers) inspect only the flags
   they own, so a `--resume <id>` token survives into the child argv alongside the injected
   `--name` and `--settings` flags. The pass-through exists mechanically; nothing validates it,
   tests it as a unit, or documents it as the emergency form.
2. Under `relaunch`, those same `claudeArgs` are handed to **every** card session the loop
   starts. A `--resume <id>` token would resume the interrupted conversation inside a foreign
   card worktree — on every subsequent card, for as long as the lane runs.
   `grep -n "resume" internal/cli/factory_lane_relaunch.go` → 0 hits: no guard exists.
3. The exact spelling the t1348 proposal suggested, `moai cc -f lane-<n>`, is refused today
   before any launch: `-f/--factory takes no argument (bare -f starts the factory leader); a
   lane joins with -l or --lane` (`internal/cli/factory.go:92`, `factoryFlagUsageError`). The
   emergency form is the lane join: `moai cc -l --name lane-<n> -- --resume <session-id>`.

So item ④ is not "add a resume flag" — the token already reaches the child on the one-shot
path. It is: make the emergency form explicit and validated as a named, testable assembly, and
close the one path where the token leaks across cards (the relaunch loop).

## §B Requirements (GEARS)

### B.1 The version reads

- **REQ-SCV-001** — **When** a session view resolves a registry entry whose process is alive,
  the session view shall report the Claude Code version that process is running, read from the
  process's own binary mapping (macOS: the text mapping of the pid; Linux: `/proc/<pid>/exe`),
  and shall not take it from the installed binary, an environment variable, or the registry
  record.
- **REQ-SCV-002** — The session view shall report the latest installed version as the version
  segment of the resolved `claude` binary found on PATH (the symlink target), accepting both
  `versions/X.Y.Z` and `claude-code/X.Y.Z` path shapes, and shall not spawn the `claude`
  process to obtain it.
- **REQ-SCV-003** — **When** any version read fails — a dead pid, an unreadable mapping, a
  resolved path carrying no version segment, or an unsupported platform — the view shall render
  `unknown` for that value, shall not infer or substitute a value, and the command carrying the
  view shall still exit 0.
- **REQ-SCV-004** — The running-version read shall be an injectable dependency: the probe
  behind it shall be a package-level seam, unit tests shall supply fixture pid→path mappings
  and fixture PATH resolutions through it, and no test of this behavior shall spawn a process
  or read the real process table.

### B.2 The surfaces

- **REQ-SCV-005** — **When** `moai session list --cc-version` runs, each listed entry shall
  carry the session's running version and the installed version (additive JSON fields plus a
  human-readable rendering), and an entry whose reads degraded shall carry `unknown` rather
  than being omitted.
- **REQ-SCV-006** — **When** `moai session list` runs without `--cc-version`, the command shall
  perform no per-process probes and shall emit exactly the field set it emits today.
- **REQ-SCV-007** — `moai doctor` shall carry a session-staleness check that reports, per live
  registry session, the running version against the installed version, and **When** a live
  session runs a version older than the installed one, the check shall warn naming both
  versions; the check shall be read-only, at most one probe per live entry, and shall never
  change `moai doctor`'s exit status (an advisory check in the `checkFlagSlot` pattern).

### B.3 The resume emergency path

- **REQ-SCV-008** — The launcher shall assemble the lane-join child argv in one pure function
  of (pass-through args, session name, settings flag) with no process, filesystem, or
  environment side effects, and the assembled argv shall carry all three: the injected session
  name, the injected settings, and the pass-through tokens (a `--resume <session-id>`
  included).
- **REQ-SCV-009** — **When** a `--resume` token is present without a following value, the
  launcher shall refuse before any launch, with an error naming the required
  `--resume <session-id>` form.
- **REQ-SCV-010** — **When** the clear policy is `relaunch` and the child arguments carry a
  `--resume` token, the supervising loop shall refuse to start, and its error shall name the
  safe one-shot lane-join form; the loop shall neither carry the token into a second card
  session nor silently strip it.

## §C Constraints

### C.1 The default list path is on the orchestrator's critical path

`moai session list --json` is the third command of the orchestrator's pre-spawn sync-check
batch (`internal/cli/session.go:45-46`'s own doc). That cost discipline is why REQ-SCV-006
pins the default path: version probing is flag-gated behind `--cc-version`, and the doctor
check bounds itself to at most one probe per live entry.

### C.2 Platform split follows the house pattern

`internal/session` already carries build-tagged platform files (`proc_info_bsd.go`,
`proc_info_linux.go`, `proc_info_other.go`, `proc_info_windows.go`) behind the `procInfoFunc`
seam (`session_pid.go:50-57`). The version readers follow the same shape: darwin (lsof txt),
linux (`/proc/<pid>/exe`), every other platform → unsupported → `unknown`.
`GOOS=windows GOARCH=amd64 go build ./...` must pass with the windows path degrading to
`unknown`.

### C.3 Advisory, never enforcing

Nothing here restarts, kills, blocks, or nags. The staleness report is information a leader
acts on at a card boundary (t1348 §A's procedure); the relaunch guard's refusal is
misfire-prevention on a command shape that cannot mean what it says under that policy, not a
policy gate on resuming.

### C.4 Template-First

Go source under `internal/session` and `internal/cli` has no template mirror, so the
Template-First rule does not apply to this SPEC's code changes.

## §D Exclusions

Explicitly out of scope. Each may be taken up separately.

### Out of Scope — items absorbed by card t1482

- Item ① of the t1348 candidate list: the `--clear-policy` help/documentation and the
  recommended lane value.
- Item ③: the relaunch loop's version-change one-line log (`claude 2.1.287 → 2.1.288` between
  cards).

### Out of Scope — enforcement

- Any automatic restart, kill, notification, or blocking of stale sessions. The report is
  advisory; the restart decision stays with the leader/operator at a card boundary.

### Out of Scope — Claude Code behavior itself

- Hot-swapping the binary into a running process, auto-update control
  (`DISABLE_AUTOUPDATER`-style configuration), or changing which version a lane launches with.
  The update path is Claude Code's own; this SPEC only makes the staleness visible.

### Out of Scope — the web console

- Any `internal/web` display of the new fields. The CLI surfaces (`session list`, `doctor`)
  own this SPEC's deliverable; a console widget is a separate change.

### Out of Scope — version persistence and history

- Recording versions into the registry file, tracking a session's version history over time,
  or diffing versions across heartbeats. A running process's version never changes; the live
  read at query time is the whole truth.
