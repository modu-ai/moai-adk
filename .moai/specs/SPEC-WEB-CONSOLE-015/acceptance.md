# SPEC-WEB-CONSOLE-015 — Acceptance Criteria

Fourteen criteria over twelve requirements. Every criterion names a command and an expected
result. Where a criterion is satisfied by an absence, its **measured pre-change baseline** is
stated, so a zero-hit result counts as evidence rather than as a vacuous pass — a criterion that
already passes on the untouched tree observes nothing and is a defect, which is why version 0.1.0's
AC-WC15-012 was deleted rather than reworded.

Baselines below were measured in this worktree at `dfbf828a6` and are quoted as observed.

## §A Framing

**AC-WC15-001** (REQ-WC15-001) — Given the branch at merge, When
`git diff <base>..HEAD -- internal/web/events.go internal/web/assets/app.js` is run, Then the diff
touches no transport behaviour: `EventSource` remains the primary channel, `POLL_MS` remains
30000, and `startPolling()` remains reachable only from the failure branch and the
missing-`EventSource` branch. Baseline (`app.js:638`, `:700`, `:721`, `:743`; `events.go:22`,
`:81`, `:95`) is the pre-change tree, so this criterion asserts preservation.

**AC-WC15-002** (REQ-WC15-002) — Two parts, because the property has a mechanical half and a
reading half, and version 0.1.0 conflated them into a grep that could not express its own
qualifier ("against any `.moai/state` path" is a semantic filter no grep applies).

*Mechanical half.* Given the merged tree, When

```
grep -rnE 'os\.(WriteFile|MkdirAll|Rename|Create)\(|WriteBestEffort\(|acquireLock\(|SaveFactoryRegistry\(|\.Mutate\(' \
  internal/web --include='*.go' | grep -v '_test\.go'
```

is run, Then its output is **exactly the two pre-existing lines** and nothing else. Baseline,
measured verbatim:

```
internal/web/profile_crud.go:38:	return os.MkdirAll(dir, 0o755)
internal/web/profile_crud.go:64:	return os.Rename(src, dst)
```

The assertion is an exhaustive inventory rather than a zero-hit grep: any write call this SPEC
introduces anywhere in non-test console code appears as a third line and fails the check, whatever
path it targets. That is strictly stronger than the qualifier version 0.1.0 tried to express.

*Reading half, stated as a separate step.* Given those two allowlisted lines, When
`internal/web/profile_crud.go:33-64` is read, Then both paths derive from `profile.GetProfileDir`
— the Claude profile directory — and neither derives from the project state directory. This step
is a human reading, declared as one, and is re-performed only if the inventory changes.

## §B Session telemetry cells

**AC-WC15-021a** (REQ-WC15-021) — Given the merged tree, When `internal/web` non-test Go source is
grepped for a declaration of the telemetry record's fields and for a reader of its file, Then the
count is **zero**; and When it is grepped for the reader symbol `SPEC-SESSION-TELEMETRY-001`
exports, Then the count is **≥ 1**. Baseline, measured: `grep -rn "internal/statusline"
internal/web --include='*.go'` returns rc=1 with no output, and
`grep -rn "context-usage\|ContextUsage" internal/web` returns exactly one line —
`viewmodel_ops.go:255`, the placeholder comment this SPEC removes. So the "≥ 1" half is new
information; the "zero" half asserts that the exclusivity survives the change and fails if a
second copy of the schema is introduced.

**AC-WC15-021b** (REQ-WC15-021) — Given a telemetry record for session `S` carrying a model, an
effort level, and a context percentage, and a chain role bound to `S`, When the kanban view model
is built, Then that role's model, effort, and context-percentage fields equal the record's values.
Baseline: on the pre-change tree the same build yields `Model: ""`, `Effort: ""`,
`ContextPct: -1` for **every** role — the literal placeholders at `viewmodel_ops.go:253-255` — so
any non-placeholder value is new information.

**AC-WC15-023** (REQ-WC15-023) — Given two sessions `A` and `B` where only `A` has a readable
telemetry record, When the console renders both role rows, Then `A` shows its own values and `B`
shows the "not recorded" marker — specifically, `B` does not show `A`'s values. This is the direct
regression test for the single-slot last-writer-wins race recorded in spec.md §A.5.

## §C Per-lane factory progress

**AC-WC15-043a** (REQ-WC15-043) — Given a factory registry mapping `lane-2` to PID `P`, an
active-sessions entry with PID `P` and session id `S`, and a record for `S`, When the factory view
model is built, Then the `lane-2` row carries `S`'s record values; **and** When the project's
state directory is listed recursively immediately before and immediately after the join call,
Then the two listings are identical — no file created, none removed, none modified. The
before/after listing replaces version 0.1.0's "when the merged tree is grepped, then no new file
is created", which named no observation a grep can make.

**AC-WC15-043b** (REQ-WC15-043) — Given a factory registry mapping `lane-4` to a PID present in no
active-sessions entry, and separately a lane whose resolved session has no record, When the
factory section is rendered, Then in both cases the `lane-4` row is **present**, carries its lane
number, and carries the unresolved marker — it is neither dropped nor rendered blank. Baseline:
the pre-change console renders no lane rows at all — `grep -rn "Factory\|factory" internal/web
--include='*.go' --include='*.templ'` (non-test) and the same grep for a lane role both return
zero, and `ChainRoles` (`viewmodel_ops.go:46`) fixes the four chain roles the view iterates — so
every half of this criterion is unobservable before the change.

**AC-WC15-044** (REQ-WC15-044) — Given two registered lanes with complete join data, When the
factory section is rendered, Then each row contains the lane number, the card identifier, the SPEC
identifier where present, the session state, and the stage. Baseline: as AC-WC15-043b — zero lane
rows exist pre-change.

**AC-WC15-045** (REQ-WC15-045) — Given a lane whose stage came from `estimateStage`
(`viewmodel_ops.go:266-275`) with `estimated == true`, When its row is rendered, Then the row
carries the estimated marker; and Given `estimated == false`, Then it does not. The second half is
what stops the marker being rendered unconditionally, which would pass the first half while
observing nothing.

**AC-WC15-046** (REQ-WC15-046) — Given a project root with no factory registry file, and
separately one whose registry file is malformed JSON, When the kanban page is requested, Then in
both cases the response status is 200 and the factory section renders zero lanes.

**AC-WC15-047** (REQ-WC15-047) — Given a registry in which `lane-1` and `lane-5` both carry PID
`P`, an active-sessions entry with PID `P`, and a record for that session, When the factory
section is rendered, Then **neither** row carries that record's card identifier or SPEC
identifier, and **both** rows carry the unresolved marker. Baseline: as AC-WC15-043b — no lane
rows exist pre-change, so this behaviour cannot pass on the untouched tree. Without this
criterion the natural map-lookup implementation attributes one session to both lanes silently,
which is the reachable hazard REQ-WC15-047 names.

## §D Cross-cutting

**AC-WC15-050** (REQ-WC15-050) — Given the merged tree, When the existing i18n governance test in
`internal/web/i18n_governance_test.go` runs, Then it passes with **no allowlist entry added**, and
every key this SPEC introduces is present in the `en`, `ko`, `ja`, and `zh` maps of
`internal/web/assets/i18n.js`.

**AC-WC15-051** (REQ-WC15-051) — Given a records directory holding one record written by a build
predating this SPEC's dependencies (generated in the test by marshalling the pre-change struct,
not hand-authored) and one written after, and a telemetry directory holding a snapshot for only
the second, When the console renders the chain, Then both rows render, the pre-change row showing
the "not recorded" marker for each field its record and snapshot do not carry, and no request
returns an error status.

**AC-WC15-052** (REQ-WC15-052) — Given the merged tree, When `internal/web/screens.templ` is
grepped for `are not recorded yet` and for `kanban.Record is extended`, Then both return **zero**;
and When the note-banner call in the kanban section is read, Then its third argument is a non-empty
translation key; and When that key is looked up, Then it is present in all four locale maps.
Baseline, measured: `screens.templ:192` currently reads

```
@noteBanner("info", "Stage is estimated from heartbeat. Model, effort and context usage are not
recorded yet, so they are left blank — they fill in once kanban.Record is extended.", "")
```

— one hit for each grep, and an empty third argument, which `noteBanner`
(`internal/web/widgets.templ:40-52`) renders as a bare `<span>` with no `data-i18n` attribute. So
each half of this criterion is new information.

## §E Definition of Done

- [ ] Every requirement in spec.md §B maps to at least one criterion above, and every criterion
      maps to a requirement.
- [ ] `SPEC-SESSION-TELEMETRY-001` and `SPEC-KANBAN-RECORD-SESSION-KEY-001` have landed. Until
      both have, the criteria in §B and §C cannot be satisfied — spec.md §A.5.
- [ ] Affected-package tests pass (`internal/web`); the full-suite verdict is read from CI, not
      from a local run.
- [ ] `templ generate` re-run and the generated files committed alongside their sources.
- [ ] `go vet ./...` and `golangci-lint run` clean.
- [ ] Coverage on `internal/web` does not regress from the pre-change baseline.
