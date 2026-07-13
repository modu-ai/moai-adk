---
id: SPEC-HARNESS-CADENCE-BUILD-001
version: "0.1.0"
status: draft
updated: 2026-07-13
---

# Acceptance Criteria — SPEC-HARNESS-CADENCE-BUILD-001

Verification philosophy: Go-observable behavior is verified through `go test` fixtures (reachability through the real call path — Validate/ListHarnesses/Remove wrapper/DoctorReport), never by grep alone. Doctrine text is verified by grep in BOTH trees (live + template mirror) because for prose deliverables the text is the deliverable; cross-file seams (recipe → queue contract, gate → manifest field, retrofit → CLI verb) get their own reachability ACs.

## §D AC Matrix

| AC | REQ | Severity | Verification (command → expected) |
|----|-----|----------|-----------------------------------|
| AC-HCB-010 | HCB-010 | MUST | `go test ./internal/harness/v4manifest/ -run 'Schedule' -count=1` — fixture manifests with `{"interval":"30m","mechanism":"loop","mode":"discovery-only"}` and `{"interval":"0 3 * * *","mechanism":"cron","mode":"discovery-only"}` both pass `Validate` |
| AC-HCB-011 | HCB-011 | MUST | Same test run — every pre-existing valid fixture WITHOUT `schedule` still passes `Validate` (regression table case); no existing test modified to accommodate the change |
| AC-HCB-012 | HCB-012 | MUST | Fixture with `"mode":"run"` (and `"mode":""`) → `Validate` returns error whose message contains `discovery-only` |
| AC-HCB-013 | HCB-013 | MUST | Fixture with `"mechanism":"daemon"` → error; fixture with `"interval":""` → error |
| AC-HCB-014 | HCB-014 | MUST | Decoder test: JSON omitting `mode` inside a declared `schedule` fails `Validate` (no silent default) |
| AC-HCB-001 | HCB-001/005 | MUST | `grep -n "recurrence" .claude/skills/moai/workflows/harness-builder.md` ≥ 1 match inside the PLAN gate section, AND the same file states the discovery-only description obligation for the question options (`grep -c "discovery-only"` ≥ 1); identical result on the template mirror |
| AC-HCB-003 | HCB-003/004 | MUST | harness-builder.md Artifact 5 content contract lists optional `schedule` with the three sub-fields (`grep -A3 'schedule' harness-builder.md` shows `interval`/`mechanism`/`mode`), and states omission-on-decline; both trees |
| AC-HCB-020 | HCB-020 | MUST | `grep -c "Scheduled Harness Discovery" .claude/rules/moai/workflow/cadence-bridge.md` ≥ 1 AND a `Recipe 4` heading exists; identical on template mirror |
| AC-HCB-021 | HCB-021 | MUST | Recipe 4 payload is prompt-form: recipe block contains a discovery-prompt template naming the harness + inline constraints; `grep -c "/harness:<name>"` inside the Recipe 4 payload block = 0 (payload does not invoke the entry command) |
| AC-HCB-022 | HCB-022 | MUST | `grep -c "scheduled runs never commit, never push, never enter run-phase" cadence-bridge.md` = 1 (live) and = 1 (mirror) — invariant present, unmodified, not restated |
| AC-HCB-023 | HCB-023 | MUST | Recipe 4 references the existing Discovery-to-Queue Contract section (grep for the section name near Recipe 4); no new queue path introduced (`grep -c ".moai/reports/cadence/<date>.md"` unchanged from baseline count) |
| AC-HCB-024 | HCB-024 | MUST | Eligibility table gains exactly one new row for the harness discovery prompt (table row grep, both trees) |
| AC-HCB-025 | HCB-025 | SHOULD | Recipe 4 (or its fallback note) covers Cron-unavailable degradation to `/loop` (grep, both trees) |
| AC-HCB-030 | HCB-030 | MUST | harness-builder.md ACTIVATE section contains the registration step: `grep -c "CronCreate" harness-builder.md` ≥ 1 AND paste-ready `/loop` emission with session-scoped caveat (`grep -c "session-scoped"` ≥ 1); registration ordered AFTER the smoke gate in the section flow; both trees |
| AC-HCB-031a | HCB-031 | MUST | `go test ./internal/cli/harness/ -run 'List' -count=1` — fixture harness with schedule: `HarnessEntry.Schedule` populated; `--json` output includes `"schedule"` |
| AC-HCB-031b | HCB-035 | MUST | Fixture without schedule: `--json` output omits the `schedule` key entirely; text output line byte-identical to baseline expectation |
| AC-HCB-032 | HCB-032 | MUST | `go test ./internal/cli/harness/ -run 'Remove' -count=1` — remove of schedule-bearing fixture prints an unregister notice containing the declared mechanism token (`cron`→`CronDelete`, `loop`→loop cancellation); schedule-less fixture prints no notice; removal atomicity tests still green |
| AC-HCB-033 | HCB-033 | MUST | `go test ./internal/cli/harness/ -run 'Doctor' -count=1` — fixture manifest with `"mode":"write"` yields an ERROR-severity DoctorFinding and non-zero-exit report state, reached through the doctor scan path (axis-2 Validate reuse), not asserted on Validate directly |
| AC-HCB-034 | HCB-034 | MUST | Boundary grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/harness/ \| grep -v _test.go \| grep -v '^[^:]*:[0-9]*:[ \t]*//'` → 0 matches; no Cron-tool invocation tokens in Go code |
| AC-HCB-035 | HCB-035 | MUST | `moai harness list` on this repo (release-update has no schedule) → output contains no schedule text; `moai harness doctor` exit code unchanged (0) |
| AC-HCB-040 | HCB-040/041 | MUST | harness-builder.md ANALYZE section contains the research sub-step: `grep -c "resolve-library-id"` ≥ 1, `grep -c "WebFetch"` ≥ 1, and the sub-step's findings are named as PLAN-aggregate input; both trees |
| AC-HCB-042 | HCB-042 | MUST | Research sub-step cross-references the MCP Fallback Strategy (grep "MCP Fallback"); degradation is continue-not-block (grep "never block" or equivalent authored phrase); both trees |
| AC-HCB-043 | HCB-043 | MUST | Research sub-step cross-references GLM routing (`grep -c "glm-web-tooling"` ≥ 1); both trees |
| AC-HCB-044 | HCB-044 | SHOULD | Skip clause with recorded rationale present (grep near the research sub-step); both trees |
| AC-HCB-050 | HCB-050 | MUST | harness-build-entry.md contains the Schedule Retrofit branch: `grep -c "Schedule Retrofit" harness-build-entry.md` ≥ 1 with detection rule (existing command file + scheduling intent); both trees |
| AC-HCB-051 | HCB-051 | MUST | Retrofit branch names `moai harness edit` as the path-discovery surface and orders: recurrence round → manifest edit → registration (section-flow grep); both trees |
| AC-HCB-052 | HCB-052 | MUST | Retrofit branch documents the command-only path: distinctive phrase "without fabricating a manifest" (or equivalent authored token) present; user-informed limitation stated; both trees |
| AC-HCB-053 | HCB-053 | MUST | `ls internal/template/templates/.claude/commands/harness/ internal/template/templates/.claude/agents/harness/ 2>&1` → both absent/empty (dev-only isolation unchanged); split-namespace leak test green (`go test ./internal/template/ -run 'SplitHarnessNamespace' -count=1`) |
| AC-HCB-060 | HCB-060 | MUST | `for f in <4 files>; do diff "$f" "internal/template/templates/$f"; done` → exit 0 ×4 (byte parity restored post-edit) |
| AC-HCB-061 | HCB-061 | MUST | `grep -rn "SPEC-HARNESS-CADENCE-BUILD\|REQ-HCB-" .claude/rules/moai/workflow/cadence-bridge.md .claude/skills/moai/workflows/harness-build*.md .claude/skills/moai/SKILL.md internal/template/templates/` → 0 matches |
| AC-HCB-062 | HCB-062 | MUST | `make build` green; `go test ./internal/template/... -count=1` green (neutrality + leak guards) |
| AC-HCB-070 | all Go | MUST | `go test ./... -count=1` green (full suite, not only touched packages) |
| AC-HCB-071 | all Go | MUST | `golangci-lint run` → 0 issues on touched packages |
| AC-HCB-072 | HCB-035 | MUST | `moai harness doctor` on the real repo → exit 0; github/release remain INFO (Runner asymmetry preserved) |
| AC-HCB-073 | doc | MUST | `moai spec lint` → 0 findings attributed to SPEC-HARNESS-CADENCE-BUILD-001 |

## Given-When-Then Scenarios

### S1 — Build with recurrence (cron)

- **Given** a user runs `/moai harness "build a harness for docs drift review"` and completes Discovery + name derivation,
- **When** the PLAN→GENERATE gate round asks the recurrence question and the user selects yes / nightly / cron,
- **Then** the generated `manifest.json` carries `"schedule": {"interval":"nightly","mechanism":"cron","mode":"discovery-only"}`, GENERATE completes, and ACTIVATE — only after `moai harness doctor` reports zero ERROR findings — issues CronCreate with the Recipe 4 discovery prompt; `moai harness list` then shows the schedule.

### S2 — Build declining recurrence (baseline preservation)

- **Given** the same build flow,
- **When** the user declines recurrence at the gate,
- **Then** the manifest contains no `schedule` key, and `list`/`edit`/`remove`/`doctor` outputs are indistinguishable from the pre-SPEC baseline for that harness.

### S3 — Scheduled discovery run finds work (invariant + queue)

- **Given** a registered nightly cron carrying the Recipe 4 discovery prompt for harness X,
- **When** the scheduled turn finds drift in X's domain,
- **Then** findings persist to the active TaskList (session ledger live) or `.moai/reports/cadence/<date>.md` (otherwise), are surfaced at the next interactive session, and the scheduled turn performs zero writes beyond the queue record, zero commits, zero pushes, and no run-phase entry.

### S4 — Retrofit a manifest-bearing harness

- **Given** `/moai harness "run release-update weekly on a schedule"` and `.claude/commands/harness/release-update.md` exists with a manifest,
- **When** the build-entry workflow detects existing-harness + scheduling intent,
- **Then** it routes to the Schedule Retrofit branch (NOT the Builder), runs the recurrence round, adds the `schedule` object to `release-update`'s manifest via orchestrator-mediated edit, registers per the chosen mechanism, and `moai harness list` surfaces it.

### S5 — Retrofit a command-only harness

- **Given** the same request targeting `github` (command-only, no manifest by design),
- **When** the Retrofit branch runs,
- **Then** it registers the schedule via the Recipe 4 prompt WITHOUT creating a manifest, and informs the user that `list`/`doctor` schedule surfacing is unavailable for manifest-less harnesses.

### S6 — Invalid schedule caught by the gate

- **Given** a hand-edited manifest with `"mode": "write"`,
- **When** `moai harness doctor` runs,
- **Then** it reports an ERROR finding naming the schedule violation and exits non-zero; the harness is not declared active by any ACTIVATE flow.

## Edge Cases

- **E1 — Cron tools unavailable**: registration degrades to the `/loop` paste-ready form; declared `mechanism: "cron"` in the manifest stays (declaration ≠ live registration); the ACTIVATE emission notes the degradation.
- **E2 — Session death with armed /loop**: nothing persists the loop; a later session re-arms from the declared schedule surfaced by `moai harness list`. No state registry is added (Exclusions).
- **E3 — Unknown extra keys inside schedule**: decoder tolerates them (Go JSON unknown-field tolerance); only the 3 known sub-fields are validated.
- **E4 — Concurrent session during retrofit manifest edit**: pre-spawn sync check discipline applies; a detected race degrades to report-only per the cadence-bridge fallback precedent.
- **E5 — Retrofit request whose name matches no harness**: falls through to the normal Builder NL path (creation), never errors.
- **E6 — Scheduled prompt fires while another session holds the checkout**: discovery run is read-only by construction except the queue record; queue write follows the same race-safe degradation as existing Recipes 1-2 backlog writes.

## Quality Gates / Definition of Done

- All MUST ACs pass with evidence recorded in progress.md §E.2 (verbatim command outputs; file-redirect contract for long logs).
- Coverage: touched packages (`internal/harness/v4manifest`, `internal/cli/harness`) ≥ 85% package-level after change.
- TRUST 5: Tested (table-driven fixtures), Readable (godoc on new exported types), Unified (gofmt/golangci-lint clean), Secured (no secrets; CLI boundary grep clean), Trackable (conventional commits per milestone, SPEC ID in commit scope).
- Three [NEEDS CLARIFICATION] markers in plan.md resolved before Implementation Kickoff Approval.
- No modification to: catalog invariant sentence, Implementation Kickoff Approval doctrine, AskUserQuestion monopoly surfaces, `internal/manifest`, learning-subsystem `internal/harness` manifest.jsonl lineage.
