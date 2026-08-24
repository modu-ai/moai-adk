# SPEC-SESSION-TELEMETRY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored (Tier L set complete): `spec.md`, `plan.md`, `acceptance.md`, `design.md`,
  `research.md`, plus this `progress.md`.
- Tier: L (justification in `plan.md` §B — 23 files, plus an always-loaded doctrine read
  procedure). Threshold 0.85.
- SPEC ID regex check executed as Bash, output `PASS`. ID uniqueness confirmed:
  `ls .moai/specs | grep -i "SESSION-TELEMETRY"` returns no match.
- Budget: **9 requirements, 11 acceptance criteria** (ceilings 25 / 25).
- Every absence-satisfied criterion carries a pre-change baseline measured in this tree at
  `dfbf828a6`.
- Two open verification items are recorded in `plan.md` §F as run-phase measurements, not as
  blockers.
- Status: `draft`. Awaiting plan-audit.

## §E.2 Run-phase Evidence

Branch `WT-web-live-todo`, base `5564e16f2`, run HEAD `3bb8f4393`. Every baseline below was
re-measured in this tree during this run, not carried from `acceptance.md`.

### Re-measured pre-change baselines (at `5564e16f2`)

| Command | Observed |
|---|---|
| `grep -rn '"raw_pct"' internal/` | 4 hits — `statusline/context_usage.go:63`, `statusline/context_usage_test.go:150`, `cli/tokens.go:86`, `cli/tokens_test.go:283` |
| `grep -rln "state/context-usage.json" .claude internal/template/templates` | the same 4 doctrine files `acceptance.md` names |
| `grep -rln "context-usage.json" docs-site/content \| wc -l` | `12` |
| `diff -q` on each mirror pair | both print nothing, exit 0 |
| `golangci-lint run ./internal/statusline/... ./internal/cli/...` | `0 issues.` |

All four agree with `acceptance.md`. No disagreement to report.

### Per-criterion matrix

| AC | REQ | Verdict | Command → observed |
|---|---|---|---|
| AC-ST-001 | REQ-ST-001 | **PASS** | `go test ./internal/statusline/ -run TestPerSessionRecordPath -v` → `--- PASS`. Asserts `state/context-usage/<S>.json` exists and `state/context-usage.json` does not |
| AC-ST-002 | REQ-ST-002 | **PASS (fixture half) / PARTIAL (grep half)** | `-run TestRecordKeyIsThePayloadSessionID -v` → `--- PASS` (payload `payload-S`, sidecar `sidecar-T`; the record is named for `S`, none for `T`). Grep half: `grep -rn 'CurrentSideChannelFile\|current-session-id' internal/statusline/` → **2 hits, both inside the AC-ST-002 fixture itself** (its doc comment and the `os.WriteFile` that creates the sidecar the criterion mandates). Non-test source: **0 hits**. The two halves are mutually unsatisfiable as written while the fixture lives in-package — see § Finding 1 |
| AC-ST-003 | REQ-ST-004 | **PASS** | `-run TestRecordedModelIsBackendResolved -v` → `--- PASS`. `display_name = "Opus 5 (1M context)[1m]"` + `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.3` → recorded model `glm-5.3`, no `[1m]`, not the Claude name |
| AC-ST-004 | REQ-ST-003 | **PASS** | `-run TestModelAndEffortRoundTripAndOmit -v` → `--- PASS`. Present: round-trip unchanged. Absent: neither `model` nor `effort` appears in the marshalled JSON, write still succeeds |
| AC-ST-005 | REQ-ST-005 | **PASS** | `grep -rEc '^func ReadSessionTelemetry\(' internal/statusline/*.go` summed → `1`. `grep -rEn '^func [A-Z][A-Za-z]*\(.*\) \(?\*?SessionTelemetryRecord' internal/statusline/*.go` → exactly `context_usage.go:239` and no other. Baseline both `0` |
| AC-ST-006 | REQ-ST-005 | **PASS (count deviates, see note)** | `grep -rn '"raw_pct"' internal/` → 3 hits, all inside `internal/statusline`. `grep -rln '"raw_pct"' internal/ \| grep -v '^internal/statusline' \| wc -l` → `0` (baseline `2`). The binding half — every hit inside `internal/statusline` — passes. `acceptance.md` forecast 4 hits in 2 files; observed 3 in 3, because the `internal/cli` fixture was not moved verbatim: it is now marshalled from `statusline.SessionTelemetryRecord`, so it spells no schema at all |
| AC-ST-007 | REQ-ST-006 | **PASS** | Removal: `grep -rn 'context-usage' internal/cli/tokens.go` → exactly one hit, the help string (now `:411`), neither constant nor struct. Preservation: `go test ./internal/cli/ -run TestRunTokensRecord` → `ok`; the context block still carries `raw_pct` when a record exists and is absent (exit 0) when none does |
| AC-ST-008 | REQ-ST-007 | **PASS** | `-run TestHostileKeyIsRefused -v` → `--- PASS` with all four subtests (`empty`, `parent_traversal`, `path_separator`, `absolute_path`). Tree snapshot diff: zero files created anywhere, including inside the per-session directory; each render still returned its line. Pre-fix measurement recorded in the M4 commit: `../escape` wrote `state/escape.json` |
| AC-ST-009 | REQ-ST-008 | **PASS** | (a) `grep -rln "state/context-usage.json" .claude internal/template/templates \| wc -l` → `0`. (b) `state/context-usage/` → the same four files the baseline named. (c) both `diff -q` print nothing, exit 0 |
| AC-ST-010 | REQ-ST-003 | **PASS** | `-run TestReadsPreviousSchemaRecord -v` → `--- PASS`. The fixture is marshalled from the pre-change field set, so key order and indentation are the marshaller's own |
| AC-ST-011 | REQ-ST-009 | **PASS** | (a) `grep -rln "context-usage.json" docs-site/content \| wc -l` → `0` (baseline `12`). (b) `context-usage/` → `12`. (c) per locale → `3 en / 3 ja / 3 ko / 3 zh`. `hugo --quiet` exits 0 with no warning output |

### plan.md §F — the two open verification items, now measured

**§F.1 — does the widened payload defeat the write throttle? NO.** Measured by
`TestThrottleUnaffectedByModelAndEffort`: five renders with unchanged context, model, and
effort leave the record's mtime untouched. The two fields are therefore kept **IN** the throttle
comparison (`sameSemanticPayload`) — no fallback was needed, and the staleness hazard the
fallback would have introduced does not arise. The same test asserts the other direction: a
changed effort and a changed model each reach disk on the very next render.

**§F.2 — does a GLM-backed session's payload carry `effort`? NOT OBSERVED.** `env | grep -c
ANTHROPIC_DEFAULT` → `0`; no GLM session was running on this machine during the run, and no
render payload from one was capturable. Left as a gap; nothing is inferred. The Claude-backend
path is observed (`effort` present and round-tripped).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-24
run_commit_sha: 3bb8f4393
run_status: complete
ac_pass_count: 11
ac_fail_count: 0
ac_partial_count: 1        # AC-ST-002 grep half — see §E.2 Finding 1
preserve_list_post_run_count: 3   # internal/web untouched; on-disk path name; best-effort render contract
cross_platform_build:
  host: "go build ./... → exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... → exit 0"
tests:
  command: "go test -cover ./internal/statusline/... ./internal/cli/... ./internal/spec/..."
  result: "all ok, exit 0"
  coverage:
    internal/statusline: 90.5%
    internal/cli: 78.8%
    internal/spec: 89.1%
lint:
  command: "golangci-lint run --timeout=8m ./internal/statusline/... ./internal/cli/... ./internal/spec/..."
  result: "0 issues."
  new_warnings_or_lints_introduced: 0
template_first:
  mirror_pairs_identical: true
  make_build_run: true
  neutrality_guard: "MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ → ok"
total_run_phase_files: 27
m1_to_mN_commit_strategy: "one commit per milestone (M1..M6) plus one AC-repair commit; nothing pushed"
pushed: false
```

### Finding 1 — AC-ST-002's two halves conflict as written

The criterion requires BOTH a fixture that creates `.moai/state/current-session-id.txt` inside
`internal/statusline` AND that
`grep -rn 'CurrentSideChannelFile\|current-session-id' internal/statusline/` return zero. A
fixture that creates that file necessarily spells its name, so the grep cannot return zero while
the fixture exists in-package. The run preserved the half `acceptance.md` itself calls
load-bearing (the two-id fixture) and removed every avoidable mention elsewhere: the writer's doc
comment no longer names the sidecar, so non-test source carries zero hits. The remaining two hits
are the fixture's own. This is reported rather than resolved — `acceptance.md` is not this
agent's to edit.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-24
sync_commit_sha: pending-backfill-SPEC-SESSION-TELEMETRY-001
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-SESSION-TELEMETRY-001' CHANGELOG.md → 0 (no duplicate; emission permitted)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 11; the CHANGELOG entry states 11"
b12_self_test_c: "every path named in the CHANGELOG entry verified present with ls before commit"
changelog_entry_position: "[Unreleased] → ### Changed, appended last"
frontmatter_status_transitions:
  spec.md: "in-progress → completed (updated: refreshed)"
  plan.md: "n/a — carries no frontmatter block"
  acceptance.md: "n/a — carries no frontmatter block"
  design.md: "n/a — carries no frontmatter block"
  research.md: "n/a — carries no frontmatter block"
  progress.md: "n/a — carries no frontmatter block"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy that its own sync tests"
```

### Coverage regression — NOT measured, and what was measured instead

**Claim withheld.** No coverage-regression figure exists for this SPEC. The pre-change package
rate for `internal/statusline` was never measured on this tree, so no delta can be stated, and
none is stated. This is a **gap**, not a pass.

**What the lead substituted, and why.** Measuring the pre-change rate requires a second checkout
of the base commit in its own worktree — the run-phase edits are spread across seven commits on
this branch and cannot be selectively reverted in place. The lead judged that cost above the
benefit for a change whose new surface is four small functions, and directed a per-function
coverage read instead.

**What was measured** (this sync, worktree `t207`, HEAD `e6b01f9b2`):

```
go test -coverprofile=/tmp/st.cov ./internal/statusline/
  → ok  github.com/modu-ai/moai-adk/internal/statusline  25.123s  coverage: 90.5% of statements

go tool cover -func=/tmp/st.cov | grep context_usage.go
  SessionTelemetryPath      100.0%
  usableSessionKey          100.0%
  buildContextUsageRecord   100.0%
  ReadSessionTelemetry      100.0%
  ...
  writeContextUsage          80.0%
  isTemplateSourceDir        88.9%
  resolveProjectDir          87.5%
```

Every function this SPEC **adds** — `SessionTelemetryPath`, `usableSessionKey`,
`buildContextUsageRecord`, `ReadSessionTelemetry` — is at 100.0%, above the 90.5% package rate.
Stated precisely, because the aggregate would hide it: `writeContextUsage` (80.0%),
`isTemplateSourceDir` (88.9%) and `resolveProjectDir` (87.5%) sit **below** the package rate.
All three pre-date this SPEC; `writeContextUsage` is modified by it, the other two are untouched.
Whether this SPEC moved `writeContextUsage`'s own figure is unknown for the same reason the
package delta is unknown.

### GLM-backend `effort` — still unobserved

`plan.md` §F.2 asked whether a GLM-backed session's payload carries `effort`. Run-phase recorded
`env | grep -c ANTHROPIC_DEFAULT → 0`: no GLM session ran on this machine during the run, and no
render payload from one was capturable. **Nothing changed at sync.** The answer remains
unobserved, and nothing about it is inferred from the Claude-backend path (which is observed:
`effort` present and round-tripped). Carried forward as an open measurement, not as a defect.

### Docs-site consistency — verified, unchanged

The twelve pages were written during run-phase M6; sync verified rather than rewrote them.

| Command | Observed |
|---|---|
| `grep -rln "context-usage.json" docs-site/content \| wc -l` | `0` (pipeline rc 1 — a genuine zero-match, not a swallowed failure) |
| `grep -rln "context-usage/" docs-site/content \| wc -l` | `12` |
| per locale | `3 en / 3 ja / 3 ko / 3 zh` |

Spot-read of the three `en` pages: each describes the per-session record, its key, and the
`model` / `effort` fields as implemented. No page describes superseded behaviour. **No docs-site
file was modified by this sync commit.**

## §G Ownership correction — DISCHARGED

`f276b9742` rescoped **AC-ST-002**'s grep half to non-test source, closing the run's Finding 1
(the criterion's two halves were mutually unsatisfiable: the fixture it mandates must create the
sidecar file, and creating it spells the name the unscoped grep forbids). The correction's
content was ratified; what was owed was the **ownership path** — `acceptance.md` body content
belongs to `manager-spec` (`spec-frontmatter-schema.md` § Forbidden ownership crossings), and
that edit had been applied orchestrator-direct under time pressure rather than through a
re-delegation.

**Discharged.** The edit was re-delegated and re-authored by `manager-spec` in commit
**`e6b01f9b2`** — `docs(SPEC-SESSION-TELEMETRY-001): re-author AC-ST-002 under its owner (t207)`,
touching `acceptance.md` alone (15 insertions, 9 deletions). Observed:

```
git show --stat e6b01f9b2
  e6b01f9b20a076adca75293d44d8204a755c5527
  docs(SPEC-SESSION-TELEMETRY-001): re-author AC-ST-002 under its owner (t207)
  .../specs/SPEC-SESSION-TELEMETRY-001/acceptance.md | 24 ++++++++++++++--------
  1 file changed, 15 insertions(+), 9 deletions(-)
```

The ratified substance survives unchanged, and the re-authoring added what the hasty version
could not: at the baseline commit BOTH the scoped and the unscoped grep return zero, because the
fixture that spells the sidecar's name does not exist there yet — which explains the defect
rather than only recording it. Nothing is owed at sync; this section is closed.
