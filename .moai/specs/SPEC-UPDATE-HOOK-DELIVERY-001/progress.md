# SPEC-UPDATE-HOOK-DELIVERY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-03
tier: M
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md
baseline_sha: d592b0551
open_decision: RESOLVED at run-phase M1 — Option B (detect + guide), operator decision 2026-09-03; verdict recorded in design.md §G
```

### M1 — Decision landing (2026-09-03, tree b77ae5d5e)

- **Decision: Option B (detect + guide only).** Operator decision fixed 2026-09-03 at Implementation Kickoff Approval. `moai update` merge behavior unchanged; `moai doctor` gains a read-only check that flags template hook entries missing from the project's `.claude/settings.json` within carried event keys, with per-entry guidance. Verdict + rejection rationale in design.md §G.
- **N/A markings (option-gated, per acceptance.md header):** REQ-UHD-007 (Option A) N/A; REQ-UHD-010 (Option C) N/A; AC-UHD-004, AC-UHD-005 (Option A gate) N/A; AC-UHD-007 (Option C gate) N/A. Binding option-gated set: REQ-UHD-008, REQ-UHD-009, AC-UHD-006, AC-UHD-013.
- **Detection identity rule:** entry identity = `.claude/hooks/moai/*.sh` handler path (canonical-JSON extraction); no-identity fallback = canonical JSON. Template side = embedded `settings.json.tmpl` rendered with the project's `hook.opt_in.enabled` (no false positives on opt-out projects). No template content change required.
- **[SUPERSEDED 2026-09-03, sync-audit F1 — verdict 1220c70af]** The shipped identity rule adds the matcher: entry identity = handler path **+ `|matcher`** (`doctor_hook_delivery.go` identity join; the matcher is read from the entry's `"matcher"` field). The path-only record above — and design.md §G's identical clause — described an identity that would collapse the template's four `handle-pre-tool.sh` PreToolUse blocks (4 matchers) into one, so a user carrying any ONE block would read as full parity with 3 blocks missing. The implementation kept path+matcher (guarded by `TestCheckHookDelivery_FlagsMissingEntryInCarriedEventKey`, whose missing block shares the script with a kept block). design.md is frozen post-close, so this note is the supersession record.

### M1 pre-flight baseline (tree b77ae5d5e, branch WT-update-hook-delivery, worktree clean)

| Check | Command | Observed |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Affected packages | `go test ./internal/cli/update/... ./internal/merge/...` | all `ok`, exit 0 |
| Lint | `golangci-lint run --timeout=2m ./internal/cli/... ./internal/merge/...` | `0 issues.` |

### Defect-chain anchor re-verification (absorbed tree b77ae5d5e vs plan-phase d592b0551)

| Anchor | Plan-phase | This tree | Moved? |
|---|---|---|---|
| `pruneToShared` map-only recursion / wholesale copy / shared-key exclusion | base.go:112-131 (:124 recurse, :128 copy, :116-121 exclusion) | identical content at base.go:112-131 | no |
| "Only user changed" keeps user array | strategies.go:427-429 | identical content at strategies.go:427-429 | no |
| `checkHooksConfig` single os.Stat, settings.json never opened | doctor.go:768-783 | identical content at doctor.go:767-783 | 1 line |

## §E.2 Run-phase Evidence

Run-phase completed 2026-09-03 on branch WT-update-hook-delivery (recovered from an API-interrupted session; all evidence below re-measured in the recovery session). This tree also absorbed origin/develop at b77ae5d5e before M1.

### AC matrix (Option B binding set; N/A per M1 decision)

| AC | Status | Verification (command → observed, this run, tree below) |
|---|---|---|
| AC-UHD-001 | PASS (guard) | `go test ./internal/cli/update/... -count=1` → `ok … internal/cli/update/plan` + `ok … internal/cli/update/merge` (template-new event-key delivery unchanged; no merge behavior changed under Option B) |
| AC-UHD-002 | PASS (guard) | `TestMergeKeepsUserDeletionInCarriedEventKey` (M2, f51cb973d) in the same run → `ok … internal/cli/update/merge` |
| AC-UHD-003 | PASS (Option B report branch) | Defect surface pinned green at M2 (`TestMergeDropsTemplateAdditionInsideCarriedEventKey` — asserts the template-side addition inside a carried key is dropped silently today, i.e. the gap EXISTS); resolved by the Option B detector: smoke fixture (below) shows `moai doctor` naming the missing entry with its event key + remediation |
| AC-UHD-006 | PASS | Smoke A (below): warn names `hooks.PreToolUse missing handle-pre-tool.sh (matcher AskUserQuestion)` + remediation incl. `git status --porcelain \| grep '^ D'`; rc=0; unit test `TestCheckHookDelivery_FlagsMissingEntryInCarriedEventKey` additionally asserts settings.json content+mtime unchanged (read-only) |
| AC-UHD-008 | PASS | `TestCheckHookDelivery_InvalidJSONWarnsGracefully` (warn naming settings.json; broken file byte-unchanged after the check) in targeted run → `ok … internal/cli` |
| AC-UHD-009 | PASS | `TestCheckHookDelivery_NonArrayHookValueWarnsAndSkips` (anomalous key named and skipped; other carried keys still resolve) in targeted run → `ok` |
| AC-UHD-010 | PASS | `TestCheckHookDelivery_SilentWhenAllEntriesPresent` (full parity → ok, no report) + the check is stateless/read-only, so repeated doctor runs are idempotent; update-side idempotence unchanged (Option B, no merge change) |
| AC-UHD-011 | PASS (guard) | Non-hook subtree untouched by construction — the check reads only; `TestCheckHookDelivery_DoesNotFlagTemplateNewEventKeys` + `TestCheckHookDelivery_UserAuthoredEntryNotFlagged` pin the comparison direction; merge tests unchanged → `ok` |
| AC-UHD-012 | PASS (boundary) | Card-scope diff `git diff 7664729ab^..940adf966 -- internal/` = 8 files: `internal/cli/{doctor.go,doctor_hook_delivery.go,doctor_hook_delivery_test.go,testdata/doctor-*.golden}` + `internal/cli/update/merge/base_test.go` (M2) + `internal/template/catalog.yaml` (hash refresh) — zero matches on `installPreCommitHookOptional` / `.git/hooks` write paths on both the card scope and the full `d592b0551..940adf966` range. (F2 repair, sync-audit 1220c70af: the original citation ran over the absorbed-develop range — 53 files/+2498 — because this branch absorbed origin/develop at b77ae5d5e before M1; the conclusion held on every scope the auditor measured.) |
| AC-UHD-013 | PASS | `TestCheckHookDelivery_MissingSettingsJSONInformational` (ok; no settings.json created) + `TestCheckHookDelivery_NoHooksKeyInformational` (ok) in targeted run → `ok` |
| AC-UHD-004, AC-UHD-005, AC-UHD-007 | N/A | Option A / Option C gated; not selected (M1) |

**Adoption-gate letter note (sync-audit F3, verdict 1220c70af):** AC-UHD-003's four mechanical RED elements (failing command + verbatim RED output + exit code + tree SHA at M2) were not captured as specified — M2 pinned the defect surface with a green-at-birth characterization (`TestMergeDropsTemplateAdditionInsideCarriedEventKey`) instead, substituting green-characterization for RED-observation, and plan M2(a)'s other two characterization tests (REQ-UHD-001/003) went undelivered. The substance stands independently: the green test proves the silent drop is live, the M3 mutation check burns the detector path in both directions, and the sync-auditor re-verified drop + resolution via its own real-CLI smoke. Disclosed letter deviation, not a silent one.

### Mutation check (non-vacuous green proof)

Corrupted the detection path in `doctor_hook_delivery.go` (`if _, found := userIDs[id]; !found` → `… && false`), then:

```
$ go test ./internal/cli/ -run 'TestCheckHookDelivery|TestHookEntryIdentities' -count=1
--- FAIL: TestCheckHookDelivery_FlagsMissingEntryInCarriedEventKey (0.01s)
--- FAIL: TestCheckHookDelivery_OptInAwareRendering (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.390s
FAIL
```

Both RED-direction tests burn the detection path (the GREEN-direction `SilentWhenAllEntriesPresent` correctly still passes — the mutation empties `missing`, not the ok-path). Reverted; `diff` against pre-mutation backup byte-identical; targeted run back to `ok`.

### Real-CLI smoke (built binary `bin/moai` @ f51cb973d+working tree, /tmp fixtures, opt_in=false rendering)

Smoke A — fixture whose `.claude/settings.json` carries 3 of the template's 4 PreToolUse matcher blocks:

```
$ cd /tmp/t466-smoke-miss && bin/moai doctor --check "Hook Delivery"
    STATUS  CHECK          MESSAGE
    warn    Hook Delivery  hooks.PreToolUse missing handle-pre-tool.sh (matcher AskUserQuestion); re-add each missing entry under the named event key (copy the block from the template settings.json of your moai version); after moai update verify no managed file was deleted: git status --porcelain | grep '^ D'
    0 ok, 1 warn, 0 fail
rc=0
```

Smoke B — fixture carrying all 4 blocks: `ok   Hook Delivery  hook entries match the shipped template` / `1 ok, 0 warn, 0 fail`, rc=0.

### Coverage (targeted run basis; package-wide verdict belongs to the full internal/cli run)

```
$ go test ./internal/cli/ -run 'TestCheckHookDelivery|TestHookEntryIdentities' -coverprofile=… -count=1
doctor_hook_delivery.go: checkHookDelivery 81.7% / renderedTemplateHooks 69.2% / hookEntryIdentities 96.3% / sortedTemplateEventKeys 100.0%
```

`renderedTemplateHooks`'s uncovered branches are internal-anomaly error paths (embedded-load / render failure) unreachable in a healthy binary.

### Quality gates (this run)

| Gate | Command | Observed |
|---|---|---|
| Lint | `golangci-lint run --timeout=2m ./internal/cli/...` | `0 issues.` (2 staticcheck S1011 findings surfaced during the run — fixed in-tree before commit: test-loop append spread + detail-loop append spread) |
| Vet | `go vet ./internal/cli/ ./cmd/moai/` | exit 0 |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Merge/update packages | `go test ./internal/cli/update/... ./internal/merge/... -count=1` | 4× `ok` |
| Doctor goldens | `UPDATE_GOLDEN=1 go test … -run TestDoctorGolden` → regenerated, then targeted run | `ok` |

### Golden snapshot refresh (two attribution lines)

`internal/cli/testdata/doctor-*.golden` regenerated. The diff carries (1) this SPEC's new `Hook Delivery` line + counts (4 ok, 8 warn / Pass 18) — mine; and (2) the `Agent Emit Embed` message losing its `(not a MoAI project root)` suffix — a PRE-EXISTING stale golden inherited by this tree: the message change landed in 8f6cc5d7f (t427, 2026-09-02) while the goldens were pinned at 96bfa0c99 (t392, 2026-09-01); the refresh resolves that inherited red alongside this SPEC's line.

### Scope / reverse-dependency

Touched packages: `internal/cli` (only). Reverse dependencies via `go list -json ./...`: `cmd/moai` (sole importer of `internal/cli`) — verified by `go vet` + `go build` (test files: none, main package). Doctor registration diff (doctor.go) is 4 added lines inside `runGroupedChecksObserved`'s `workspaceChecks`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: 7664729ab, f51cb973d, 2b4582a43 (M1 decision landing / M2 characterization / M3 detector+smoke; a05c17164 catalog-hash chore follow-up excluded)
run_status: complete
ac_pass_count: 10
ac_fail_count: 0
ac_na_count: 3
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-applicable — isolated card worktree, lane does not push; lead batch-pushes develop (gitflow-lane-protocol §4)
l44_post_push_fetch: not-applicable — same; remote landing is the lead's verification
new_warnings_or_lints_introduced: 0
cross_platform_build.windows_amd64: pass
total_run_phase_files: 9
m1_to_mN_commit_strategy: per-milestone commits (M1 decision landing 7664729ab, M2 characterization f51cb973d, M3 detector+smoke this commit)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-03
sync_commit_sha: "85743a34a" # D3 backfilled — the close commit 85743a34a (docs(SPEC-UPDATE-HOOK-DELIVERY-001): sync-phase 3-phase close (card t466))
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-UPDATE-HOOK-DELIVERY-001' CHANGELOG.md → 0 pre-emission (no duplicate; emission proceeded)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 13 (AC-UHD-001..013, non-zero, plausible); CHANGELOG entry cites 13 ACs (10 PASS / 3 N/A, matching §E.3 counts)"
b12_self_test_c: "all cited paths ls-verified — internal/cli/doctor_hook_delivery.go · internal/cli/doctor.go · internal/cli/doctor_hook_delivery_test.go · internal/cli/testdata/doctor-{dark,light,nocolor}.golden exist"
changelog_entry_position: "CHANGELOG.md [Unreleased] → Added, top position (SPEC-UPDATE-HOOK-DELIVERY-001)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (3-phase close merged into the single sync commit — no separate Mx commit)"
  plan_acceptance_progress: "plan.md / acceptance.md carry no status field (artifact statelessness) — nothing to transition; progress.md is body-only"
  updated_bumped: "2026-09-03 → 2026-09-03 (same-day close; spec.md only)"
canary_compliance_check:
  spec_body_edits: 0   # spec.md / plan.md / acceptance.md body untouched — spec.md frontmatter status/updated only
  codemaps_regeneration: "not executed — no exported API surface added (4 unexported leaf functions in internal/cli); the internal/cli codemap is unaffected by one doctor check registration (4 added lines in doctor.go)"
docs_site_decision: "UPDATED — docs-site/content/{ko,en,ja,zh}/cli-reference/doctor.md gained a 'Hook Delivery' section (4-locale same-change, ko canonical, badge v3.1.4), inserted after Home Disk Usage before exit codes: the page enumerates doctor checks, so the new check belongs there. hugo build verified warning-free; section-count parity holds across the 4 locales."
readme_decision: "no change — README.md:734 describes `moai doctor` generically with one illustrative example (Home Disk Usage), not an enumeration of checks, so the row is not stale; a 4-locale README resync for a non-enumeration surface is out of proportion (scope discipline)."
mx_tag_changes: "validated, 0 added / 0 removed / 0 updated — no MUST-level gate fires: all 4 functions in doctor_hook_delivery.go are unexported (checkHookDelivery / renderedTemplateHooks / hookEntryIdentities / sortedTemplateEventKeys), checkHookDelivery fan_in = 1 (doctor.go registration only), and all are covered by doctor_hook_delivery_test.go. @MX:NOTE is a 'consider' for exported functions; none apply."
e3_run_sha_backfill: "§E.3 run_commit_sha backfilled pending-backfill-run → 7664729ab, f51cb973d, 2b4582a43 — performed by manager-docs under the lead's explicit sync dispatch (the D3 exemption normally assigns §E.3 backfill to manager-develop; the orchestrator routed it here); values match the run session's §E.2/§E.3 record"
```


