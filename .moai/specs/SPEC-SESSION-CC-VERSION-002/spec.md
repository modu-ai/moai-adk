---
id: SPEC-SESSION-CC-VERSION-002
title: "Structural resume-argv interpretation and install-root-anchored version extraction (t1515)"
version: "0.1.0"
status: completed
created: 2026-10-05
updated: 2026-10-06
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "launcher, resume, argv-interpretation, claude-options, cc-version, install-root, card-review"
era: V3R6
tier: M
related_specs: [SPEC-SESSION-CC-VERSION-001]
depends_on: [SPEC-SESSION-CC-VERSION-001]
---

# SPEC-SESSION-CC-VERSION-002 — Structural resume-argv interpretation and install-root-anchored version extraction

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-05 | Initial draft. Carries exactly the two residual items predecessor SPEC-SESSION-CC-VERSION-001 (merged as develop `985bd43da`) marked in code with `@MX:DEBT` + `@MX:CEILING` + `@MX:UPGRADE: t1515`, per the approved follow-up card t1515: W1 — structural replacement of the hand-scan resume guard/validator whose five card-review rounds (r1-r5, `.moai/reports/t1465/card-review.md`) each found one more argv shape the hand-synced maps misjudged; W2 — anchoring version-segment extraction on the claude install root (r5 P2②). Requirements trace to the r-round or DEBT marker that motivates each. |

## §A Background

Every claim below was re-verified in this tree (`WT-session-cc-version` at `099250516`, the
predecessor's final DEBT-marker commit) on 2026-10-05. The full r1-r5 findings/dispositions
history is `.moai/reports/t1465/card-review.md`; the code both items live in was read in full.

### A.1 W1 — the structural root cause is table drift

`internal/cli/lane_resume.go` carries two pure interpreters of the child argv —
`validateResumeArgs` (REQ-SCV-009's valueless-resume refusal) and `carriesResumeToken` (the
relaunch guard's carrier detection) — whose value-ness knowledge lives in three hand-synced
maps: `claudeValueTakingOptions` (measured from `claude --help` 2.1.289), `launcherValueTakingOptions`
(repo-owned), and `ambiguousValueOptions` (the optional-value class). Five card-review rounds
each found one more shape the hand-rolled scan misjudged:

| Round | Finding | Class |
|---|---|---|
| r1 P1 | the `-r` short alias unrecognized | token shape |
| r1 P2 | scan did not stop at Claude's separator | segment semantics |
| r2 P1 | attached `-r<uuid>` unrecognized | token shape |
| r2 P2 | value-taking options unknown to the scanner — bypass (guard) and false refusal (validator) | table drift |
| r3 P1 | short clusters (`-pr<uuid>`) and optional-value options (`-w`) | token shape + value class |
| r4 P2 | `-p` post-separator consumed a phantom value via the launcher table | segment semantics |
| r5 P2① | `--remote-control-session-name-prefix` (required-value in 2.1.289) missing → a legit call whose prefix value is literally `--resume` falsely refused | table drift |

Each repair added a rule or a table entry; the ceiling the r3-r5 leaders named is structural:
correctness depends on enumerating a moving option surface, and every enumeration gap
manifests silently in one of the two dangerous directions — a validator false refusal, or a
guard false pass that leaks one conversation across every card the loop starts. The DEBT
markers (`lane_resume.go:51-53`) name the upgrade: "replace the hand-rolled scan with a real
argv parse or spawn-time resume detection".

### A.2 The regression contract is the design's hard boundary

All r1-r5 reproductions exist as regression tests in `internal/cli/lane_resume_test.go` — 19
test functions, ~51 cases — and they MUST stay green with the new design; none may be weakened.
Reading the suite settles the design space more sharply than the DEBT text does:

- The suite forces per-option claude knowledge in BOTH modes. `TestPostSeparatorLauncherFlagsAreInert`
  requires the validator to REFUSE `["--", "--profile", "--resume"]` (the launcher-only `--profile`
  consumes nothing in the claude segment), while `TestAmbiguityNeverRefusesInValidator` requires
  the validator to PASS `["-w", "--resume"]` (claude's optional-value option shields silently).
  Both argvs have the shape `[unknown-flag, flag-shaped --resume]`; no shape-only rule
  distinguishes them — only the option's measured claude class does. A purely table-less scanner
  therefore cannot satisfy the suite.
- The suite pins exact function signatures: `validateResumeArgs(args []string) error` and
  `carriesResumeToken(args []string) bool` are called directly by 14 of the 19 tests.
- The suite pins both polarities and the r3 leader ruling verbatim: the guard is fail-closed
  (an unprovable resume carrier fires; a false fire costs a restatable launch, a false pass
  leaks a resume), the validator never refuses on ambiguity, a `--` is never consumed as an
  option's value, and everything after Claude's own second separator is prompt text.

### A.3 W2 — the version read is not anchored to the claude install root

`internal/session/ccversion.go` `versionSegmentFromPath` (line 176) takes the FIRST
`(?:^|/)(?:versions|claude-code)/<n>` match of `versionSegmentRe` (line 167), so an unrelated
`versions/` prefix upstream of the real install root satisfies the read. Measured in this tree
at `099250516` via a `go test -overlay` probe (repo untouched):

```
$ go test -overlay /tmp/t1515-w2-red/overlay.json ./internal/session/ -run '^TestT1515RedOverlayRepro$' -count=1 -v
    zz_t1515_red_test.go:10: versionSegmentFromPath = "9", want 2.1.281 (unanchored first-match reads the unrelated prefix)
--- FAIL: TestT1515RedOverlayRepro (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/session	0.308s
```

`/opt/versions/9/tools/claude/versions/2.1.281` reads `9` — a stale process could escape the
doctor staleness warning (`REQ-SCV-007` of the predecessor) on exactly the overlay layouts the
read exists to catch. The anchor vocabulary already exists one function over:
`mappingPathNamesClaudeBinary` (line 156) recognizes the three real claude product-directory
shapes (`/claude/versions/<v>`, `/claude-code/<v>`, trailing `/<v>/claude`). Degraded reads
render `unknown` and never an inferred value (`REQ-SCV-003` of the predecessor) — that
contract is unchanged.

### A.4 The design direction (binding card decision)

The card binds: EITHER split launcher argv and Claude argv at `--` and parse each with its own
option model — noting r4 already does the segment split, so the real remaining question is how
the CLAUDE side's value-ness is decided without a hand-synced table — OR detect actual resume
at spawn time. Spawn-time detection is rejected in plan.md §A.4 (it converts
refusal-before-start into a post-hoc kill; the leak has already happened, and the suite pins
zero-seams refusal). The chosen direction is the segment split with a runtime-derived claude
option model and polarity-safe unknown defaults — §A.2's reading shows the model cannot be
abolished, so the structural fix makes it (a) derived from the binary being launched rather
than hand-synced, and (b) non-load-bearing for safety: every residual gap falls to the safe
polarity per mode.

## §B Requirements (GEARS)

### B.1 W1 — structural interpretation of the resume argv

- **REQ-SCV-011** — **When** the launcher interprets child arguments, it shall read them as
  two segments split at MoAI's `--` pass-through marker — the launcher segment before it, the
  claude segment after it — and **When** Claude's own argument separator (a second `--`) is
  met, the interpretation shall stop, judging no token after it; a `--` token shall never be
  consumed as an option's value. The launcher segment's value-taking flags shall be decided by
  the launcher's own parsers' surface (repo-owned, not drift-prone) and shall consume values
  only before MoAI's separator — after it, `-p` is claude's boolean `--print` and none of the
  launcher's flags exist.
- **REQ-SCV-012** — The claude segment's option model shall be derived from the very `claude`
  binary the launcher resolves to execute — from that binary's own `--help` option synopsis —
  through a package-level seam in the `procInfoFunc` pattern, bounded in time, and shall not be
  carried by a hand-synced table alone; **When** the derivation fails or the binary is absent,
  the interpretation shall degrade to the compile-time measured snapshot (which carries
  source-version provenance in its comments), and no derivation failure shall change any exit
  status. No test of this behavior shall spawn a process.
- **REQ-SCV-013** — The guard and the validator shall classify tokens through the one model of
  REQ-SCV-012 with opposite safe defaults. An option the model marks required-value (`<value>`
  synopsis) shall consume the next token in both modes — unless that token is `--`. An option
  the model marks optional-value (`[value]` synopsis) shall consume the next token in the
  validator only — ambiguity never refuses — while the guard judges the token after it. An
  option the model marks boolean, or cannot classify at all, shall consume nothing in either
  mode, so the guard judges the token after it (fail-closed: an unprovable resume carrier
  fires) and the validator judges it too.
- **REQ-SCV-014** — The three hand-synced maps (`claudeValueTakingOptions`,
  `launcherValueTakingOptions`, `ambiguousValueOptions`) shall be replaced by the single
  derived model of REQ-SCV-012 plus its snapshot fallback, and the interpretation's safety
  shall not depend on the snapshot's completeness: a snapshot gap manifests only as one of
  three named outcomes — a false guard fire (a restatable launch), a validator pass (claude's
  own error surfaces at launch), or the validator false-refusal residual plan.md §A.4 names,
  which arises only under the compound condition that the derivation is down, the option is
  absent from the snapshot, and its value is resume-shaped — and it shall never be a silently
  leaked resume under the relaunch policy. On completion
  the `@MX:DEBT` + `@MX:CEILING` + `@MX:UPGRADE: t1515` markers in `lane_resume.go` and
  `ccversion.go` shall be removed — the upgrade trigger has fired — and any residual shall
  carry a fresh, honestly-named marker instead.
- **REQ-SCV-015** — All r1-r5 regression tests in `internal/cli/lane_resume_test.go` shall
  pass unchanged through the rework, and the r5 P2① instance shall be present in the measured
  model as required-value, so that a call carrying
  `--remote-control-session-name-prefix --resume` (a prefix value that literally reads
  `--resume`) is passed by the validator and never falsely refused.

### B.2 W2 — install-root-anchored version extraction

- **REQ-SCV-016** — **When** the version-segment extractor reads an install path, it shall
  anchor the extraction on the claude product-directory shapes `mappingPathNamesClaudeBinary`
  already recognizes — `/claude/versions/<v>`, `/claude-code/<v>`, or a trailing `/<v>/claude`
  binary-name shape — and an unrelated `versions/` or `claude-code/` prefix outside a claude
  product directory shall never satisfy the read: `/opt/versions/9/tools/claude/versions/2.1.281`
  reads `2.1.281`.
- **REQ-SCV-017** — **When** a path carries no anchored claude product-directory version
  segment, the extractor shall return empty and every caller shall render `unknown`, never an
  inferred value — the predecessor's degradation contract (`REQ-SCV-003`) and the
  ` (deleted)`-suffix stripping of the mapping read are unchanged.

## §C Constraints

### C.1 The regression suite is the contract, not a suggestion

`internal/cli/lane_resume_test.go`'s existing 19 test functions are the accumulated r1-r5
verdicts. They stay byte-identical; new tests go into new sibling files, and
`internal/cli/lane_resume_test.go` stays byte-identical. The
two walk-function signatures stay (`validateResumeArgs([]string) error`,
`carriesResumeToken([]string) bool`) because the suite calls them directly.

### C.2 Purity of the interpreters

`validateResumeArgs` and `carriesResumeToken` stay pure functions of their `[]string` input
reading package state only — the derivation of REQ-SCV-012 happens behind a package seam at
the launcher-entry call sites, never inside the walk. No test of this SPEC's behavior spawns a
process (the predecessor's `REQ-SCV-004` seam discipline, extended to the model seam).

### C.3 The refusal texts are user-facing contracts

`resumeRequiresValueError` and `relaunchResumeRefusal` name their required forms verbatim and
the suite asserts substrings of them; the rework changes neither text.

### C.4 Platform split and template neutrality

The derivation spawns the resolved claude binary with `--help` — plain `exec.Command`, no
build tags needed — and `GOOS=windows GOARCH=amd64 go build ./...` must stay green. Go source
under `internal/session` and `internal/cli` has no template mirror; Template-First does not
apply.

## §D Exclusions

### Out of Scope — spawn-time resume detection

- The card's alternative branch: detecting actual resume after the child starts. It converts
  the pinned refusal-before-start (`TestRelaunchRefusesResumeToken` asserts gate, lease, and
  launch seam counts of 0) into a post-hoc kill, and the conversation is already attached by
  the time detection could fire — the leak it exists to prevent has already happened.

### Out of Scope — the emergency form's user-facing shape

- Changing the emergency spelling (`moai cc -l -- --resume <session-id>`), the refusal texts,
  or the launcher's own cobra entry parsing. The r1-r5 repairs and this SPEC leave the entry
  surface (`factoryFlagUsageError`, `laneFlagNameError`, the desugared `--name`) untouched.

### Out of Scope — generalizing beyond the resume interpretation

- A general-purpose argv parser library, a CLI-framework swap, or any use of the option model
  beyond the two interpreters and their guard/validator consumers. The model exists to decide
  value-ness for the resume scan; it is not an argument-parsing framework.

### Out of Scope — version-read surface changes

- Any change to `moai session list`, `moai doctor`, the registry record, or the platform probe
  family (the predecessor's M2/M3 surfaces). W2 touches only the segment-extraction function
  the two reads already share.

### Out of Scope — hot-swap and update control

- Claude Code's own update path, binary hot-swapping into running processes, or changing which
  version a lane launches with — excluded by the predecessor and unchanged.

## §F Verification Summary (machine-verifiable)

- The full r1-r5 regression suite: 19 test functions, each `--- PASS`, zero FAIL, zero skips —
  with `lane_resume_test.go` unmodified (AC-SCV-015).
- The option model classifies injected fixture help text through the seam with no process
  spawn (`grep` over the new test files → 0 `exec.Command`) (AC-SCV-012).
- Structural greps: the three map identifiers → 0 references in `internal/cli/` (baseline: 9,
  measured at `099250516`); the two `@MX:UPGRADE: t1515` markers → 0 (baseline: 2, measured at
  `099250516`) (AC-SCV-014).
- The overlay repro shape, as a permanent table case:
  `versionSegmentFromPath("/opt/versions/9/tools/claude/versions/2.1.281")` = `2.1.281`
  (RED-now measured at `099250516`, `--- FAIL` with output `= "9"`) (AC-SCV-016).
- Env-scrubbed verification:
  `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -timeout 30m ./internal/cli/... ./internal/session/...`;
  `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `go vet` and
  `golangci-lint run` on the changed packages (§D of acceptance.md).
