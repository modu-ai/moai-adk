---
id: SPEC-JEV-LEDGER-001
title: "Plan — Jev decision ledger recorder and per-consumer config surface"
version: "0.2.1"
created: 2026-10-26
updated: 2026-10-26
---

# Plan — SPEC-JEV-LEDGER-001

Tree measured for 0.2.0: `0c2d2aa4c230d2a6671037cefea260ede18e295d` (card worktree `t1762`, branch `WT-jev-internal-jev`). Every 0.2.0 claim about existing code comes from a command run against that tree (§I). Between `cb30e05b3e` (the 0.1.0 measurement tree) and `0c2d2aa4c2` only this SPEC's five files changed, so the 0.1.x code citations still hold.

0.2.1 (plan-audit repair, D1-D9) was measured on `954adab25e7d6d83b63f8fb1c8f1aafa859f50c1`. Between `0c2d2aa4c2` and `954adab25e` only this SPEC's five files changed (§I), so every 0.2.0 code citation still holds; the new 0.2.1 citations are listed in §I.

The scope is the split-"a" ruling recorded in spec.md HISTORY 0.2.0: M1 and M2 only. The sections run from the decisions most likely to change to the mechanical steps.

---

## §A. Data model — the ledger row (highest change likelihood)

One JSON object per line in `.moai/state/jev/decisions.jsonl`. Field names are snake_case and stable; later fields are additive only. The recorder serializes what the caller supplies, except for the two fields it owns (`outcome_label`, `action` under shadow).

| Field | Type | Supplied by | Notes |
|---|---|---|---|
| `ts` | RFC 3339 UTC string | caller | |
| `consumer` | string | caller | |
| `event` | string | caller | the -002 adapter supplies `pre-tool` / `stop` / `subagent-stop` |
| `harness` | string | caller | |
| `call_id` | string | caller | groups the rows of one request |
| `question_id` | string | caller | one row per question |
| `rubric_version` | string | caller | a question-set version, not a model version |
| `model_id` | string | caller | the response's interpreted model id; empty when unanswered. Written from its own input only (REQ-JEVL-009). |
| `request_model` | string | caller | the request alias. Never copied into `model_id`, and `model_id` never copied into it. |
| `availability` | string | caller | |
| `probability` | number \| null | caller | raw, unthresholded; `null` when absent |
| `choice` / `score` | optional | caller | kind-dependent |
| `input_tokens` | int | caller | stored as supplied; its unit label is a -003 concern |
| `mode` | string | caller (effective mode) | REQ-JEVL-006 |
| `configured_mode` | string | caller (raw configured mode) | REQ-JEVL-006 |
| `threshold` | number \| null | caller | REQ-JEVL-009: always present; the supplied value, `null` when none is supplied; `0` stays `0`. Recorded only — acting on it is out of scope (-002 selects no action on it) |
| `config_warning` | optional string | caller | REQ-JEVL-008; omitted when empty |
| `action` | string | recorder | `none` whenever `mode` is `shadow` (REQ-JEVL-010) |
| `outcome_label` | string | recorder | `unlabeled` at write time (REQ-JEVL-010) |

The card's seven required fields (`model_id`, `rubric_version`, `question_id`, raw probability, `mode`, `action`, `outcome_label`) all appear. `deadline_exceeded` left the row at 0.2.0 with the deadline behaviour (-002); -002 adds it as an additive field.

## §B. Interfaces and placement

**Package placement (decision-index Q7).** The recorder is a new package `internal/jevledger`, not files inside `internal/jev`. Two reasons, both read from the tree:

1. `internal/jev/display_only_test.go:15-19` states that `internal/jev` "cannot mutate a card, a file, a branch, or a queue, because it has no dependency that could", and `TestPackageImports_AreStandardLibraryOnly` (`:20`) enforces a stdlib-only import set on that ground. A file-writing recorder inside `internal/jev` would keep that test green (`os` is stdlib) while falsifying the sentence the test exists to protect. A separate package keeps the claim true.
2. Card t1968 owns the `internal/jev` client cache and retry in another lane. A separate package gives that lane zero file overlap. **No file in `internal/jev` is edited by this SPEC.**

`internal/jevledger` production code imports the standard library only. It does not import `internal/jev`; tests may import it to compare against `jev.ModelID`.

**Recorder surface.** A recorder value is constructed from a project root and performs no I/O at construction (REQ-JEVL-001). One append method takes one row value and returns a result that names whether the row landed or was dropped and why. It never logs, never prints, and never panics on I/O failure (REQ-JEVL-012). The caller resolves the project root; -002 will pass `resolveHookProjectRoot()`'s value, which this SPEC does not touch.

**Config surface.** `WorkflowJevConfig` gains a `Consumers` field of a map type with a custom decode (§D). A pure resolver function in `internal/config/jev_consumers.go` turns one raw entry into an effective mode, the configured mode, an optional threshold, and a list of warnings as data.

**File placement of the yaml decode (audit D1).** `internal/config/jev_consumers.go` holds both custom `UnmarshalYAML(*yaml.Node) error` methods of §D and the resolver, and therefore imports `gopkg.in/yaml.v3`. This follows the repository convention: lenient per-entry decoders live in a feature-named file that imports the yaml library, not in `types.go` — `internal/config/slot_lease_config.go:14-22` (`SlotLeaseResourceConfig.UnmarshalYAML`) and `internal/config/loader_card_age_alarm.go:13,42` (`CardAgeAlarmConfig.UnmarshalYAML`); `types.go` imports no yaml library. `types.go` gains only the `Consumers` field and the struct declarations. AC-JEVL-014 permits the yaml library in `jev_consumers.go` and forbids `net`, `net/http`, `internal/jev`, and `internal/jevcred` there; the standard-library-only condition binds `internal/jevledger` alone.

**Budget constants.** Two named constants in `internal/config/defaults.go` (the repository rule: thresholds live there): budget 5 s, margin 1 s. A small derivation function takes a budget and a margin and returns the offset, so the offset is never a third stored value and so -002 can feed it test-scale values (250 ms / 50 ms). The constants are inert here: nothing in this SPEC reads them at run time.

## §C. The authority boundary (card item 5)

Evidence read in 0.1.x for this conclusion: `.moai/specs/SPEC-JEV-CORE-001/spec.md` lines 94-133 and 160-178. The 0.2.0 reduction removes every clause that touched a hook output or a report, so only the ledger write remains in question.

| CORE clause | How this SPEC meets it without amending |
|---|---|
| REQ-JEVC-011, display-only with the inert-sibling-record carve-out | The ledger row sits alongside the event it names and changes no object. Condition (i), no code path reads a row back to write a field of anything it names: the recorder never opens the ledger for reading (REQ-JEVL-011), and no reader ships here. Condition (ii): the row carries its own `request_model` / `model_id`. Condition (iii): nothing selects on a row (REQ-JEVL-014). The carve-out delegates authorisation "by name in a downstream SPEC"; this SPEC names itself (decision-index Q1). |
| REQ-JEVC-012 | No Jev answer reaches any surface: this SPEC has no caller of `Ask` at all. |
| REQ-JEVC-015/016/017 | The `consumers` block lives inside the existing `workflow.jev` key. `enabled: false` stays the template value (`internal/template/templates/.moai/config/sections/workflow.yaml:350-351`) and the compiled default (`internal/config/defaults.go:1659-1661`). |
| REQ-JEVC-018..021 | Untouched: no credential, no request. |

**User-visible surfaces.** None changes. No command, hook output, or report is added or altered.

**Stale comment.** The `WorkflowJevConfig` doc comment says the struct "carries a bare Enabled flag and nothing else" (`internal/config/types.go:1052-1058`). M1 rewrites that sentence and keeps its point: no knob changes the alias or the endpoint.

**Guards that stay unchanged.** `TestJevCallPath_HasExactlyTheDeclaredConsumers` scans `internal/cli` only (`internal/cli/doctor_jev_test.go:192-266`). This SPEC adds no `internal/cli` file, so its allow-list is not edited; the by-name entries for `hook_jev.go` and `jev_cost.go` move with those files to -002 and -003.

## §D. Config surface

```yaml
workflow:
    jev:
        enabled: false
        # consumers: per-consumer mode and threshold. No consumer is compiled
        # in this release, so an entry here has no effect until one ships.
        # mode: shadow | advise | steer | confirm  (all run as shadow today)
        consumers: {}
```

**Why a typed field cannot carry this (observed at `d1aa4ed656`, files unchanged since).** `loadYAMLFile` runs `yaml.Unmarshal(data, target)` and turns any decode error into `ErrInvalidYAML` (`internal/config/loader.go:494-495`). `loadWorkflowSection` (`:236-251`) on any error logs `slog.Warn("failed to load workflow config, using defaults", ...)` and returns (`:239-241`) before `cfg.Workflow = wrapper.Workflow` (`:247`). A plain `Threshold *float64` therefore resets the **whole** `workflow` section on `threshold: "high"` and emits a log record, breaking REQ-JEVL-007 and REQ-JEVL-008 at once.

**Decode path (D6).** Two custom `UnmarshalYAML(*yaml.Node) error` methods, both in `internal/config/jev_consumers.go` (§B, file placement), make the decode tolerant at consumer granularity; neither can fail the section:

- The map type behind `Consumers`: a mapping node decodes each value on its own; any other node shape gives an empty map plus one section-level warning. Returns nil either way.
- `JevConsumerConfig`: keeps the raw `mode` and `threshold` nodes as `yaml.Node` values and returns nil for any shape.

The resolver interprets the raw nodes: `mode` is the scalar's value for a string scalar, unrecognized otherwise; `threshold` is a `*float64` for an `!!int` / `!!float` scalar in [0, 1], absent plus a warning otherwise; absent stays absent with no warning, and `0` stays distinguishable from absent. Warnings are returned as data and never logged. `AutonomyKickoffConfig.JevMinConfidence *float64` (`internal/config/types.go:1085`) is not the pattern to copy and is not touched.

**Template-first.** The distributed change is `internal/template/templates/.moai/config/sections/workflow.yaml`, then `make build`. Editing an existing template file leaves the embed manifest's file set unchanged, so `embed-manifest-check` (`Makefile:72-74`, run by `build` at `:34`) is expected to stay green; the DoD checks that `git status --short` shows only the intended files afterwards. The local dogfood `.moai/config/sections/workflow.yaml` is not required to change.

## §E. Milestones (priority-ordered)

Sizes are estimates, not measurements.

| M | Priority | Content | Files (prod / test) | Expected changed lines |
|---|---|---|---|---|
| M1 | High | `Consumers` map with per-consumer tolerant decode (§D), resolver, budget and margin constants plus the derivation, template block, `make build`, loader-level tests | `internal/config/types.go`, `internal/config/defaults.go`, `internal/config/jev_consumers.go` (new), `internal/config/jev_consumers_test.go` (new), `internal/template/templates/.moai/config/sections/workflow.yaml` | ~110 prod + ~170 test + ~6 yaml |
| M2 | High | Recorder package: row type, lazy append (0700/0600, single write, never reads), write-side open (§F), fail-open drop, and its tests with the platform split of §F | `internal/jevledger/ledger.go` (new), `internal/jevledger/open_unix.go` (new), `internal/jevledger/open_other.go` (new), `internal/jevledger/ledger_test.go` (new), `internal/jevledger/fifo_unix_test.go` (new), `internal/jevledger/fifo_other_test.go` (new) | ~150 prod + ~330 test |

**Total estimate: 11 distinct files (M1 5, M2 6), ~260 production + ~500 test + ~6 yaml ≈ 770 changed lines.** Tier M bounds are 300-1000 lines and 5-15 files, so Tier M holds honestly; it is not at a ceiling. REQ count 12 of 16, AC count 11 of 16 (Tier M ceilings, `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier). This is above the illustrative ~10 files / ~400 lines card guidance by one file and about 370 lines, almost all of it tests; the leader's ruling fixed this scope, and it is recorded here rather than hidden.

M2 needs nothing from M1; M1 needs nothing from M2. Order M1 → M2, one commit per milestone, each naming the card id.

### AC closure map

No AC spans a milestone boundary any more: every AC closes at the one milestone that can satisfy its whole green path.

| AC | Final at | Release-blocking |
|---|---|---|
| AC-JEVL-006 | M1 | no |
| AC-JEVL-007 | M1 | **yes** |
| AC-JEVL-013 | M1 | no |
| AC-JEVL-017 | M1 | no |
| AC-JEVL-001 | M2 (its config half passes at M1 and is recorded as partial) | **yes** |
| AC-JEVL-009 | M2 | **yes** |
| AC-JEVL-010 | M2 | no |
| AC-JEVL-012 | M2 | **yes** |
| AC-JEVL-014 | M2 | no |
| AC-JEVL-018 | M2 | no |
| AC-JEVL-011 | M2 self-verification | regression-guard, not counted as a pass |

The release-blocking set is judged after M2. AC-JEVL-001's config-side test can pass at M1; that is recorded in `progress.md` §E.2 as partial evidence, and the AC is reported PASS only at M2.

## §F. Technical approach notes

- **Write-side regular-file open (REQ-JEVL-011/012).** A plain `O_APPEND|O_CREATE|O_WRONLY` open blocks forever on a FIFO with no reader and appends to a symlink's target. The read-side counterpart already exists: `internal/runtime/regular_open_nofollow.go:1` (`//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly`) opens with `O_NOFOLLOW` and `O_NONBLOCK` (`:19`) and then requires the opened descriptor to be regular; `internal/runtime/regular_open_stat.go:1` carries the complement tag `!(linux || darwin || freebsd || netbsd || openbsd || dragonfly)`. `openRegularFile` is unexported and read-only, so the recorder carries its own write-side pair:
  - `open_unix.go`, same tag as `regular_open_nofollow.go:1`: `O_WRONLY|O_APPEND|O_CREATE|O_NOFOLLOW|O_NONBLOCK` at 0600, then `Stat` on the opened handle; anything not regular is closed and the row dropped. A FIFO with no reader fails the open at once (`ENXIO`); a FIFO **with** a reader opens successfully and is refused only by the handle check; a final-element symlink fails the open (`ELOOP`).
  - `open_other.go`, same complement tag as `regular_open_stat.go:1` (audit O-b: not `_windows.go`, so no GOOS is left without an implementation): `Lstat` refuses a link, then the same handle check. No later step resolves the path again.
- **Platform split of the tests (audit R1).** `syscall.Mkfifo` and `syscall.O_NOFOLLOW` do not exist on windows (`GOOS=windows go doc syscall.Mkfifo` and `GOOS=windows go doc syscall.O_NOFOLLOW` both print `no symbol ... in package syscall`, exit 1, §I). A test file without a build constraint that names either symbol fails to **compile** for windows, and the windows-latest release gate runs `go vet ./...` and `go test` (audit evidence, `release-pr-multi-os.yml:212,316`). A run-time `t.Skip` cannot help, because compilation fails before any test runs. So:
  - `fifo_unix_test.go` carries the FIFO cases and every Unix-only flag, under the tag `//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly` — the constraint on `internal/runtime/regular_file_race_unix_test.go:1`, and the same set the production `open_unix.go` uses, so the fixture never runs where the production path is the `Lstat` variant.
  - `fifo_other_test.go` carries the complement tag and one test that skips with a named reason ("named pipes unsupported on this platform; FIFO cases run on the unix set only"), following the counterpart pattern of `internal/cli/audit_plan_cmd_fifo_other_test.go:1` (`//go:build !unix`, returns `errors.ErrUnsupported`). The visible SKIP line is the continued-firing signal on windows: the cases are reported absent, not silently missing.
  - `ledger_test.go` (no tag) holds only portable cases: mkdir failure, symlink (skipped by name where `os.Symlink` fails), concurrency, modes (skipped by name on windows; it compiles everywhere because it uses `os.FileMode` only), fields, corrupt line.
- **Concurrency.** Each row is marshalled to one byte slice ending in `\n` and written with one `Write` on the handle from the open above. Rows of a few hundred bytes stay below the practical atomic-append size on local filesystems; network filesystems are not claimed (§H).
- **Lazy creation.** The directory and file are created inside the append call, never at construction (REQ-JEVL-001).
- **Never reads.** The handle is write-only; no code path opens the ledger for reading, which is what keeps CORE carve-out condition (i) true by construction and also why REQ-JEVL-018 is about terminated malformed lines only.

## §G. Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| A later lane assumes the recorder lives in `internal/jev` | Medium | spec.md §D and §B here name `internal/jevledger`; -002 and -003 import it from there. |
| The decode tolerance regresses to a typed field | Medium | AC-JEVL-007 goes through the real loader and catches the reset-the-section mutant. |
| The FIFO tests break the windows compile | Was high (R1) | §F platform split plus the DoD `GOOS=windows go vet` command. |
| Ledger growth without rotation | Low | Out of scope; nothing writes a row in the shipped build. |

## §H. Residual risks recorded, not fixed

- CORE REQ-JEVC-011's "(today: ...)" parenthetical no longer lists every carve-out user once this SPEC lands. It is illustrative and not amended (decision-index Q1).
- Single-write append atomicity across processes is measured only on the local filesystem.
- An unterminated final fragment (a crashed partial write) is not repaired; the next row joins it on one line, and the -003 reader will count that line as one malformed line. The recorder does not read the ledger, by design.
- `O_NOFOLLOW` protects the final path element only. A link at `.moai/state/jev` would let `MkdirAll` and the open succeed and land rows outside the project (audit O-a, probe 4). Out of scope (spec.md §D).
- The tests run on darwin locally; the Linux behaviour of `ENXIO` / `ELOOP` is left to CI.
- Deleting a named test after landing is visible only in a review diff (audit D6). The continued-firing answers in acceptance.md rest on the `progress.md` §E.2 list of named `--- PASS` lines, which is read once at the run-phase self-verification; after that, CI runs whole packages and stays green when a test is removed. Review of the diff that removes it is the only surface that catches it.
- AC-JEVL-012 case (v) (a directory at the ledger path) does not separate a block-list handle check from an `IsRegular` check on the unix set, because the kernel refuses a write-only open of a directory (`EISDIR`) before the handle check runs; case (iv) is the case that pins the handle check there.

## §I. Commands that grounded the 0.2.0 plan (this run, tree `0c2d2aa4c2`)

- `git rev-parse HEAD` → `0c2d2aa4c230d2a6671037cefea260ede18e295d`.
- `git diff --name-only cb30e05b3e 0c2d2aa4c2` → the five files of this SPEC directory and nothing else (backs the "0.1.x citations still hold" statement above).
- `ls internal/jev/` → `display_only_test.go`, `jev_test.go`, `jev.go`; `grep -rln 'os.OpenFile\|os.WriteFile\|MkdirAll' internal/jev/` → `internal/jev/display_only_test.go` only (no production file in `internal/jev` writes a file).
- Read `internal/jev/display_only_test.go:15-51`: the stdlib-only import assertion and its "CANNOT mutate ... a file" rationale.
- `grep -n 'go:build' internal/runtime/regular_*.go internal/cli/audit_plan_cmd_fifo_*_test.go` → `regular_file_race_unix_test.go:1` and `regular_open_nofollow.go:1` `//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly`; `regular_open_stat.go:1` the complement; `audit_plan_cmd_fifo_unix_test.go:1` `//go:build unix`; `audit_plan_cmd_fifo_other_test.go:1` `//go:build !unix`.
- `GOOS=windows go doc syscall.Mkfifo; echo "exit=$?"` → `doc: no symbol Mkfifo in package syscall` / `exit=1`. `GOOS=windows go doc syscall.O_NOFOLLOW; echo "exit=$?"` → `doc: no symbol O_NOFOLLOW in package syscall` / `exit=1`.
- `GOOS=windows go vet ./internal/config/; echo "exit=$?"` → `exit=0` (windows test-compile baseline before M1).
- `grep -n 'WalkDir\|ReadDir\|...' internal/cli/doctor_jev_test.go` → the allow-list scan reads `internal/cli` only (`:192-266`).
- `grep -n '^build:\|embed' Makefile` → `build:` at `:34` depends on `embed-manifest-check` (`:72-74`, read-only).
- `grep -n -A3 '^    jev:' internal/template/templates/.moai/config/sections/workflow.yaml` → `:350` `jev:`, `:351` `enabled: false`.
- `grep -n 'state/' .gitignore internal/template/templates/.gitignore` → `.gitignore:401:.moai/state/`, `internal/template/templates/.gitignore:243:.moai/state/`.

0.2.1 grounding commands (this run, tree `954adab25e`):

- `git rev-parse HEAD` → `954adab25e7d6d83b63f8fb1c8f1aafa859f50c1`; `git status --short` → empty.
- `git diff --name-only 0c2d2aa4c2 954adab25e` → the five files of this SPEC directory and nothing else.
- `grep -n "UnmarshalYAML\|^import\|yaml.v3\|^type " internal/config/slot_lease_config.go internal/config/loader_card_age_alarm.go; echo "exit=$?"` → `loader_card_age_alarm.go:13: "gopkg.in/yaml.v3"`, `loader_card_age_alarm.go:42: func (c *CardAgeAlarmConfig) UnmarshalYAML(node *yaml.Node) error {`, `slot_lease_config.go:17: "gopkg.in/yaml.v3"`, `slot_lease_config.go:22: func (r *SlotLeaseResourceConfig) UnmarshalYAML(node *yaml.Node) error {` (among other lines), `exit=0` (backs §B file placement).
- `git show 0c2d2aa4c2:.moai/specs/SPEC-JEV-LEDGER-001/spec.md | sed -n '72p;74p;78p'` → the 0.1.2 REQ-JEVL-007/008/009 wording, whose threshold-recording, warning-passthrough, and per-attempt / request-alias duties are now carried by REQ-JEVL-009 here and the spec.md §D.1 rows.
- The acceptance.md §L ledger E1-E12 and B1 were re-run one by one on `954adab25e`; every output and exit code equals the 0.2.0 record.

0.1.x grounding commands (trees `cb30e05b3e`, `d1aa4ed656`, `82aff00a6d`) remain valid for the cited lines above, because no non-SPEC file changed between those trees and `0c2d2aa4c2`; their list is in git history at `0c2d2aa4c2:.moai/specs/SPEC-JEV-LEDGER-001/plan.md` §I.
