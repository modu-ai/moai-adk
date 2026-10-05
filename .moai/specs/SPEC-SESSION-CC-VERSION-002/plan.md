# SPEC-SESSION-CC-VERSION-002 — Implementation Plan

## §A Context

Producer/consumer split unchanged from the predecessor: `internal/cli` owns the resume
interpreters (`internal/cli/lane_resume.go` — the two walk functions, the guard's call site in
`internal/cli/factory_lane_relaunch.go`, the validator's call site in the `cc.go` entry path);
`internal/session` owns the version reads (`internal/session/ccversion.go`).

Evidence base (all read this plan-phase, tree `099250516`): `.moai/reports/t1465/card-review.md`
(the r1-r5 findings and dispositions), `internal/cli/lane_resume.go` + `lane_resume_test.go`
(19 test functions read in full — the regression contract of §A.2 of spec.md),
`internal/session/ccversion.go` + `ccversion_test.go`, and the W2 RED measurement (overlay
probe, `--- FAIL`, output `= "9"`).

### §A.1 Tier justification — M

| Signal | Measurement |
|---|---|
| Files | ~5 — `internal/cli/lane_resume.go` (rework), new `internal/cli/lane_resume_model.go` + its test file, `internal/session/ccversion.go`, `internal/session/ccversion_test.go` |
| Packages touched | 2 (`internal/cli`, `internal/session`) |
| LOC estimate | ~300-500, inside the M band (300-1000) — the walk shrinks as the maps dissolve |
| Cross-cutting schema change | none — no registry, no JSON, no surface change |
| Doctrine / docs sweep | none — the DEBT-marker retirement is an in-file edit under REQ-SCV-014 |
| Milestones | 5 |

Budget: 7 requirements / 7 acceptance criteria (ceilings 16 / 16).

### §A.2 The regression contract

`internal/cli/lane_resume_test.go` is the r1-r5 verdict set: 19 test functions, ~51 cases,
called against the two walk functions directly. Existing tests stay byte-identical and green;
the two signatures stay. This is the plan's hardest constraint and it is checked per
milestone, not at the end.

### §A.3 The pinned distinctions any design must reproduce

- `["--", "--profile", "--resume"]` — validator REFUSES (`TestPostSeparatorLauncherFlagsAreInert`).
- `["-w", "--resume"]` — validator PASSES (`TestAmbiguityNeverRefusesInValidator`).
- `["--append-system-prompt", "--resume"]` — guard does NOT fire, validator passes
  (`TestGuardSkipsOptionValues`, `TestValidatorSkipsOptionValues`) — required-value claude
  options consume even a flag-shaped value (measured, r2).
- `["-w", "--resume", "<id>"]` — guard FIRES (`TestGuardFiresOnAmbiguousValueOption`).
- `["--", "-p", "--resume", "<id>"]` — guard FIRES (`TestPostSeparatorLauncherFlagsInertGuard`).
- A `--` is never an option's value; the second `--` ends judgment
  (`TestSeparatorInterplaySkipsValues`, `TestSeparatorWinsOverAmbiguousValue`,
  `TestResumeScanStopsAtClaudeSeparator`).

No shape-only rule separates the first two — the option's measured claude class does. The
model stays; the hand-syncing goes.

### §A.4 The design decision (stated explicitly, with rejected alternatives)

**Chosen: two-segment interpretation over a runtime-derived claude option model with
polarity-safe unknown defaults.** One knowledge surface — a claude option model with three
measured classes (required-value / optional-value / boolean), derived at launcher entry from
the resolved claude binary's own `--help` synopsis through a package seam, with the current
measured table downgraded to a provenance-carrying compile-time fallback snapshot. The two
walk functions interpret the two argv segments (REQ-SCV-011) through the model; every token
the model cannot classify takes the safe default per mode — the guard judges the token after
it (fail-closed), the validator treats only a model-known optional-value option as shielding
(never-refuse). Rationale in one sentence: the r1-r5 history shows the failure class is
enumeration of a moving surface, so the fix must both track the surface (derivation) and stop
trusting the enumeration for safety (polarity defaults) — §A.3 shows the enumeration cannot be
abolished outright, because the pinned suite itself distinguishes `-w` from `--profile` only
by measured class.

Rejected alternatives, and why:

1. **Spawn-time resume detection** (the card's other branch) — converts refusal-before-start
   into a post-hoc kill: by the time a child could report that it resumed, the conversation is
   already attached in the foreign worktree and the leak has happened. The suite pins the
   opposite contract (`TestRelaunchRefusesResumeToken`: gate/lease/launch counts all 0 before
   any refusal). Rejected.
2. **Table-less polarity-only scanner** (no claude-side knowledge at all) — §A.3's first two
   pins are indistinguishable by shape; a table-less validator either breaks
   `TestPostSeparatorLauncherFlagsAreInert` (consume-always default) or
   `TestAmbiguityNeverRefusesInValidator` (judge-next default). The suite itself requires the
   per-option knowledge. Rejected as infeasible against the pinned contract.
3. **Runtime derivation without polarity defaults** (trust the derived model fully) — a
   derivation failure then degrades silently into exactly the r1-r5 class one level up
   (misread help text replaces a stale table). Rejected: the defaults are what make a
   residual gap non-silent-safe instead of dangerous.
4. **Hand-synced table + a drift tripwire test** (snapshot `claude --help`, fail the test on
   drift) — the staleness still manifests silently in production between re-measurements; CI
   has no claude binary to re-measure with, so the tripwire is manual. Not structural. Rejected.
5. **Compile-time codegen from a checked-in help fixture** — generated staleness is still
   staleness, plus machinery. Rejected.

Residual, named honestly (this is the SPEC's own ceiling, not a hidden one): when the
derivation fails AND a genuinely required-value claude option is absent from the fallback
snapshot AND its value is literally resume-shaped, the validator can still false-refuse (the
r5 shape). The compound condition requires the primary source to be down; the guard side is
unaffected (it judges unknown tokens anyway). If run-phase confirms this residual survives
M4, it carries a fresh `@MX:DEBT` + `@MX:CEILING` + `@MX:UPGRADE` marker per REQ-SCV-014.

## §B Known Issues (relevant subset)

- **B1 Cross-platform** — the derivation is a plain `exec.Command` of the resolved binary
  path; no syscall, no build tags. Gate: `GOOS=windows GOARCH=amd64 go build ./...` green
  before any commit.
- **B3/B11 Subagent boundary** — no `AskUserQuestion` anywhere in the touched packages; the
  existing grep discipline applies to the new file too.
- **B5 CI 3-tier** — pre-existing lint baseline vs NEW findings must be distinguished in §E;
  measure the baseline before the first edit.
- **B8 Hygiene** — no runtime-managed files; stage by explicit pathspec.
- **B10 PRESERVE** — `internal/cli/lane_resume_test.go`'s existing content, both refusal-text
  constants, `internal/cli/cc.go` and `internal/cli/factory_lane_relaunch.go`'s call sites
  (unless the model wiring requires a seam injection at entry — behavior-preserving only),
  `internal/session/ccversion_test.go`'s existing tests, everything outside the two packages.

## §C Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-session-cc-version, per staleness rule
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m ./internal/cli/... ./internal/session/... 2>&1 | tail -5   # NEW-vs-baseline split
unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -timeout 30m ./internal/cli/... ./internal/session/...
grep -n "claudeValueTakingOptions\|launcherValueTakingOptions\|ambiguousValueOptions" internal/cli/lane_resume.go | wc -l   # expect 9 (baseline)
grep -rn "@MX:UPGRADE: t1515" internal/cli/lane_resume.go internal/session/ccversion.go    # expect 2 (baseline)
```

## §D Constraints (DO NOT VIOLATE)

- `internal/cli/lane_resume_test.go` existing 19 test functions byte-identical and green;
  `internal/session/ccversion_test.go` existing tests green. New tests are added, never
  substituted.
- Signatures preserved: `validateResumeArgs(args []string) error`,
  `carriesResumeToken(args []string) bool`.
- Refusal texts `resumeRequiresValueError` and `relaunchResumeRefusal` byte-identical.
- Both polarities and the r3 ruling preserved (guard fail-closed, validator never refuses on
  ambiguity, `--` never a value, post-second-`--` prompt text).
- The walk functions stay pure; derivation behind the seam; no test spawns a process.
- No `--no-verify`, no `--amend`, no force-push; Conventional Commits with the `🗿 MoAI`
  trailer; card id `t1515` in every commit message body.
- Stage by explicit pathspec; do not touch `.moai/state/`, `.moai/harness/`, or files outside
  `internal/cli/` and `internal/session/` (plus this SPEC's directory).

## §E Self-Verification (manager-develop deliverables)

Per the 5-section evidence-bearing format; every item names command + verbatim output + tree:

- **E1** — AC PASS/FAIL matrix over `acceptance.md`, one row per AC-SCV-011..017.
- **E2** — cross-platform build: both build commands, exit 0 each.
- **E3** — coverage: `go test -cover ./internal/cli/ ./internal/session/` — new/changed files ≥ 85%.
- **E4** — boundary grep: `grep -rn 'AskUserQuestion' internal/cli/ internal/session/ | grep -v _test | grep -v "// "` → 0 hits.
- **E5** — lint: `golangci-lint run ./internal/cli/... ./internal/session/...` — 0 new issues vs the §C baseline.
- **E6** — regression-preservation proof: the 19-function suite green with
  `git diff --name-only 099250516 -- internal/cli/lane_resume_test.go` empty (base-pinned —
  the bare form compares working tree to index, so a staged mutation would report empty).
- **E7** — §F run-phase measurements with gaps recorded, not inferred.
- **E8** — RED evidence (TDD): verbatim pre-GREEN failures for AC-SCV-012 (model seam), AC-SCV-013
  (polarity cases with an injected model), and AC-SCV-016 (the anchored-extraction table case —
  its RED-now was measured at plan time on `099250516`, `--- FAIL`, `= "9"`; re-capture on the
  pre-implementation tree without pipes before GREEN).

## §F Milestones

Each milestone leaves the tree green and the 19-function suite green. Ordered by decision
reversibility: the model's shape (M1) and the interpretation semantics (M2) are the expensive
to unwind; M3-M5 are bounded and mechanical.

**Review attention.** M1 and M2. M1 fixes the option-model type, the derivation seam's
signature, and the snapshot's provenance format — every later behavior consumes it. M2 moves
the two walk functions onto the model and sets the polarity defaults — the five-round battle
zone; its GREEN run is the regression contract.

### M1 — The claude option model and its derivation seam (highest reversibility cost)

New `internal/cli/lane_resume_model.go` (name at implementer's discretion; one file, one
concern):

- The model type: option → class ∈ {required-value, optional-value, boolean}.
- The snapshot: today's measured knowledge converted into the model's snapshot form — every
  entry the pinned suite exercises stays (`--append-system-prompt`, `--settings`, `--model`,
  `-w`, `-d`, `--teleport`, the launcher set's claude-side context, `-p` as claude-boolean) —
  with provenance comments naming source version and date (initially 2.1.289, 2026-10-04).
- The derivation seam (package var, `procInfoFunc` style): resolve → `--help` (bounded) →
  parse the option synopsis into classes; failure → snapshot, silently.
- RED fixtures first: a fixture help text through the seam yields the expected classes; a
  failing derivation yields the snapshot; the no-spawn grep over the new test files.
- The walk functions are NOT yet moved — the suite stays green with the old walk consuming
  nothing new.

Ships alone: the model exists, tested, unused by the walk.

### M2 — The two-segment interpretation (the semantics decision)

`internal/cli/lane_resume.go`: move `validateResumeArgs` and `carriesResumeToken` onto the
model per REQ-SCV-011/013 — explicit two-segment parse (launcher segment: repo-owned value
flags pre-separator only; claude segment: model classes with polarity defaults), exact resume
tokens and r-cluster rules first, `--` never a value, second `--` ends judgment. The three
maps dissolve (REQ-SCV-014's structural half).

- RED fixtures first: for each model class × {guard, validator} × {flag-shaped next token,
  plain next token, `--` next, end-of-argv}, the pinned outcome — including the §A.3 pair that
  forces per-option knowledge, exercised through an injected model so the cases do not depend
  on the snapshot's completeness.
- GREEN: the full 19-function suite green, byte-identical test file (`git diff` empty).
- Do NOT special-case r5's option here — it lands with the snapshot re-measure in M4.

Ships alone: the interpretation is structural; behavior on all pinned inputs unchanged.

### M3 — W2: install-root-anchored version extraction

`internal/session/ccversion.go` + `internal/session/ccversion_test.go`:

- RED first: the overlay repro as a permanent table case
  (`/opt/versions/9/tools/claude/versions/2.1.281` → `2.1.281`) — its RED was measured at
  plan time (`099250516`, `--- FAIL`, `= "9"`); re-capture verbatim on the
  pre-implementation tree.
- GREEN: anchor `versionSegmentFromPath` on the `mappingPathNamesClaudeBinary` product-
  directory shapes (`/claude/versions/<v>`, `/claude-code/<v>`, trailing `/<v>/claude`);
  an unrelated prefix yields "".
- Degradation cases: no anchored segment → "" → `unknown` (REQ-SCV-017); the existing
  `TestVersionDegradationRendersUnknown`, `TestRunningVersionFromDeletedBinary`, and mapping
  tests stay green.

Ships alone: the doctor staleness read can no longer be escaped by an overlay prefix.

### M4 — Snapshot re-measure and marker retirement (the r5 instance)

- Re-measure the installed claude's `--help` (observe the installed version this run — do not
  carry 2.1.289 forward unobserved); update the snapshot's classes and provenance; confirm
  `--remote-control-session-name-prefix` is required-value (r5 P2① — REQ-SCV-015's second
  half) with a validator test for the prefix-value-reads-`--resume` shape.
- Remove the two `@MX:DEBT`+`@MX:CEILING`+`@MX:UPGRADE: t1515` marker sets
  (`lane_resume.go:51-53`, `ccversion.go:173-175`).
- If the §A.4 residual survives, author its fresh marker with an honest ceiling.

Ships alone: the upgrade trigger is discharged; the code no longer advertises the debt it no
longer carries.

### M5 — Verification batch (mechanical tail)

The env-scrubbed scopes of acceptance.md §C; builds ×2 (native + `GOOS=windows`); `go vet`;
`golangci-lint run` on both changed packages; coverage ≥ 85% on new/changed files; evidence
into `progress.md` §E.2/§E.3.

## §G Anti-Patterns

- **Special-casing r5's option in the walk** instead of the measured snapshot — the defect
  class is enumeration in the walk; the instance is fixed where enumeration belongs.
- **Weakening a pinned test to make the rework pass** — any `lane_resume_test.go` edit to an
  existing test is a blocker report, not a fix.
- **Spawning `claude --help` inside a unit test** — the seam exists so tests inject fixture
  text; `exec.Command` in a test of this SPEC's behavior violates the seam discipline.
- **Moving the derivation into the walk functions** — breaks purity, makes every call
  potentially spawn, and makes the pinned direct-call tests environment-dependent.
- **Letting the guard's fail-closed leak into the validator** (or the reverse) — the two
  modes resolve unknowns in opposite directions by design; one shared "conservative" default
  breaks one side's pinned contract.
- **Re-tuning refusal texts** — they are asserted substrings of pinned errors.
- **Time predictions** — priority labels and phase ordering only.

## §H Cross-References

- `.moai/reports/t1465/card-review.md` — the r1-r5 findings/dispositions history (the
  requirements source).
- `.moai/specs/SPEC-SESSION-CC-VERSION-001/` — the predecessor (merged `985bd43da`): REQ-SCV-001..010,
  the emergency form, the version reads.
- `internal/cli/lane_resume.go:51-53` and `internal/session/ccversion.go:173-175` — the two
  DEBT marker sets this SPEC discharges.
- `internal/session/ccversion.go:143-161` — `mappingPathNamesClaudeBinary`, the anchor
  vocabulary W2 reuses.
- `internal/session/session_pid.go` — the `procInfoFunc` seam pattern M1's derivation follows.
- `internal/cli/factory_lane_relaunch.go:75` — the `claudeLookPath` resolution site, where
  the launcher resolves the claude binary (the derivation's input; site pinned by the
  plan-audit).
