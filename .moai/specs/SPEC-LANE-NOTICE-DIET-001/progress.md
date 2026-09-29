# Progress — SPEC-LANE-NOTICE-DIET-001

Card: t1335 · Branch: WT-bootstrap-notice-diet · Plan-phase tree: 7bef423c0

## §E.1 Plan-phase Audit-Ready Signal

plan_complete_at: 2026-09-29T20:43:00+09:00
plan_status: audit-ready

- Artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier S scope;
  acceptance.md included per the t1335 dispatch).
- SPEC ID regex check: PASS (verbatim `PASS SPEC-LANE-NOTICE-DIET-001`).
- RED-now measurements: EV-1 (2 hits, exit 0), EV-2 (1 hit, exit 0) at
  7bef423c0 — see acceptance.md §A ledger (SSOT for all EV entries).
- Baseline-green observations: EV-3 ledger verbatim
  `ok  	github.com/modu-ai/moai-adk/internal/hook	0.523s` (anchored
  selector); EV-4 counts 0:0 with exit 1 (no-match exit = green for an absence
  probe); EV-5 `git status --porcelain -- internal cmd pkg` empty, exit 0.
- D1 RESOLVED (2026-09-29, operator decision relayed by the lead): no run-id
  token in the join line — the recorded default held; marker removed from
  plan.md §I.
- plan-auditor iteration-1 FAIL repairs applied (D1-D10, including optional
  D7-D10); audit report: `.moai/reports/t1335/plan-audit-iter1.md`. Delta
  summary for re-verdict: D1 marker gone · D2 grant-verb markers in
  AC-LND-002 + M1 · D3 AC-LND-007 → regression-guard with untracked-aware EV-5
  probe · D4 EV-4 exit 1 corrected · D5 sentence budget demoted to guidance in
  REQ-LND-005 (grammar-shaped-check justification) · D6 byte figures corrected
  to 624-byte line / ≈597-byte string · D7 this §E.1 cites the ledger, not the
  first-run duration · D8 auditors tail kept in M2 candidate + REQ-LND-003 ·
  D9 M1/M2 swapped to red-first order · D10 AC-LND-005 restated to actual
  pinned coverage; M3 extends the locale test.

## §E.2 Run-phase Evidence

Run-phase tree pin: implementation landed at `f5583700c`; the final probe
alignment landed at `1958e6a18` (branch `WT-bootstrap-notice-diet`, worktree
t1335). Every probe below was executed in this run against this tree
(env-scrubbed per-invocation form). Commit chain: M1 `fd7ae7ed4` (assertion
sweep RED + SPEC status flip) → M2 `dff54c2b9` (const rewrite GREEN) → M3
`f5583700c` (locale test extension) → M4 `1958e6a18` (EV-2 probe alignment).

### AC verdict matrix (per verification-claim-integrity §2 attribution)

| AC | Verdict | Command (verbatim) | Observed output | Tree |
|----|---------|--------------------|-----------------|------|
| AC-LND-001 | PASS (RED flipped) | `grep -c "plan-phase artifacts to manager-spec" internal/hook/lane_spawn_authority.go` | RED-now: stdout `2`, exit 0 (at 7bef423c0) → post-diet: stdout `0`, exit 1 (at f5583700c) — both occurrences (comment + const) gone | 7bef423c0 → f5583700c |
| AC-LND-002 | PASS (RED flipped) | `grep -c '"manager-spec"' internal/hook/lane_spawn_authority_test.go` | RED-now: stdout `1`, exit 0 (at 7bef423c0) → post-sweep: stdout `0`, exit 1 (at 1958e6a18); the required-marker list asserts only the six compressed markers incl. both grant-verb markers | 7bef423c0 → 1958e6a18 |
| AC-LND-003 | PASS (regression-guard re-observed) | `go test ./internal/hook -run '^(TestFactoryWorkerNoticeCarriesSpawnAuthority\|TestKanbanCompanionNoticeCarriesSpawnAuthority\|TestLaneSpawnAuthorityFailOpenPreserved)$' -count=1 -timeout 120s` | `ok  	github.com/modu-ai/moai-adk/internal/hook	0.588s`, exit 0 (at 1958e6a18; post-M2 first observation `0.683s` at f5583700c) | f5583700c / 1958e6a18 |
| AC-LND-004 | PASS (regression-guard re-observed) | `grep -rc laneSpawnAuthority internal/hook/session_start_factory_i18n.go internal/hook/session_start_kanban_i18n.go` | both files `:0`, exit 1 (no-match exit = green for an absence probe, matching EV-4's recorded semantics) | f5583700c |
| AC-LND-005 | PASS (M3 deliverable) | `go test ./internal/hook -run '^TestFactoryWorkerNoticeLocaleWordOrders$' -count=1 -timeout 120s` | `ok  	github.com/modu-ai/moai-adk/internal/hook	0.581s`, exit 0 — the test now pins zh label-first order (against the en count-first contrast) and leading/trailing-newline hygiene on `laneJoin`/`laneJoinNoCount`/`companionJoin` across en/ko/ja/zh | f5583700c |
| AC-LND-006 | PASS (regression-guard re-observed) | `TestLaneSpawnAuthorityFailOpenPreserved` (inside the EV-3 selector) | `ok` (see AC-LND-003 row) — fail-open empty-notice behavior unchanged | 1958e6a18 |
| AC-LND-007 | PASS with disclosure | `git status --porcelain -- internal cmd pkg` | empty, exit 0 at f5583700c; the run-phase diff surface is exactly `lane_spawn_authority.go`, `lane_spawn_authority_test.go`, `session_start_factory_test.go` — see Disclosure 2 below (the third path is beyond acceptance.md's whitelist enumeration but inside plan.md §F M3's named deliverable) | f5583700c |

### Quality gates

- `go vet ./internal/hook/` — exit 0, no output (at f5583700c).
- `GOOS=windows GOARCH=amd64 go build ./internal/hook/` — `windows build ok`,
  exit 0 (at f5583700c).
- Blast-radius sweep `go test ./internal/hook -run 'Notice\|SpawnAuthority'
  -count=1 -timeout 300s` — `PASS` / `ok ... 24.750s`, exit 0 (at 1958e6a18):
  every notice-composition test in the package, factory and kanban, lead and
  lane, is green on the compressed authority.
- Full package suite `go test ./internal/hook/ -count=1 -timeout 900s` — FAIL
  after 452.970s on exactly 2 tests: `TestContractRoleScopedAllowWithoutLaneMarker`
  (contract_sign_guard_test.go) and `TestHMPSourceGuardGoLiterals`
  (hmp_source_guard_test.go, hits in `contract_sign_guard.go:167,283` +
  `shell_tool.go:19`). **Not introduced by this work**: both subject files are
  outside this run's diff surface (`git diff 7bef423c0..HEAD --name-only`
  lists none of them), and the HMP SourceGuard red is already tracked by card
  t1350 at this base tree (recorded in 7bef423c0's own commit message).
  Recorded as pre-existing baseline red, not an AC failure of this SPEC.

### E8 — verbatim pre-GREEN evidence (TDD)

M1 landed the sweep as a literal RED against the unchanged const (observed
before M2, at fd7ae7ed4's working-tree state):

```
--- FAIL: TestFactoryWorkerNoticeCarriesSpawnAuthority (0.00s)
    lane_spawn_authority_test.go:53: factory worker notice re-inlines the specialist mapping ("manager-spec"):
    lane_spawn_authority_test.go:53: factory worker notice re-inlines the specialist mapping ("manager-develop"):
    lane_spawn_authority_test.go:53: factory worker notice re-inlines the specialist mapping ("manager-docs"):
    lane_spawn_authority_test.go:53: factory worker notice re-inlines the specialist mapping ("plan-phase artifacts to"):
--- FAIL: TestKanbanCompanionNoticeCarriesSpawnAuthority (0.00s)
    lane_spawn_authority_test.go:77: kanban companion notice re-inlines the specialist mapping ("manager-spec"):
    ... (same four absence hits on the kanban twin)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.575s
```

(The rendered-notice tails are elided here for brevity; each failure printed
the full long-form authority as its got-value. Red for the right reason: the
const still carried the inline mapping when this run was captured.)

### Deviations and disclosures

1. **M1 literal-RED expectation.** plan.md §F M1 expected the compressed
   marker tests to be literal RED against the unchanged const. Observed: they
   PASSED — the long-form const is a superset containing every compressed
   marker as a substring. The sweep therefore ALSO added absence assertions
   (`manager-spec` / `manager-develop` / `manager-docs` /
   `plan-phase artifacts to` must not appear in the rendered notice), which
   produce the honest literal RED (captured above) and pin AC-LND-001 at the
   notice-output level in addition to the EV-1 source grep. The pointer-only
   stub mutant remains killed by the two grant-verb markers (audit D2).
2. **EV-5 whitelist enumeration gap (plan-phase artifact defect).**
   acceptance.md's EV-5 whitelist names `lane_spawn_authority.go`,
   `lane_spawn_authority_test.go`, and the two i18n files "(only if M3
   trims)" — but plan.md §F M3's own deliverable extends
   `TestFactoryWorkerNoticeLocaleWordOrders`, which lives in
   `session_start_factory_test.go`, and the dispatch change surface names the
   same file. That test file is therefore in the sanctioned scope despite the
   whitelist omission; REQ-LND-008 (out-of-scope surfaces untouched) holds —
   no leader-notice i18n field, `session_stale_run.go`, launcher, or
   `internal/cli/` file was modified. The whitelist enumeration should be
   reconciled at sync; manager-develop is barred from editing acceptance.md
   body and reports the gap instead.
3. **EV-2 probe alignment (1958e6a18).** The M1 absence sweep initially
   carried the forbidden names as interpreted Go string literals, so the EV-2
   probe still hit the test file (2 hits). The forbidden-substring literals
   were switched to raw strings — assertions byte-identical (absence), probe
   honestly reads 0 hits / exit 1. Comments in the test name the constraint.
4. **Join-line trim (M3 verification item).** No trim made: all 4 locales'
   join lines are already single-sentence core form; `%[n]` pins, the
   count/no-count split, and the 4-locale set are untouched. Outcome recorded
   per plan.md §F M3: "no change needed".

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29T21:40:00+09:00
run_commit_sha: "1958e6a18"
run_status: complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 4  # leader-notice i18n fields (both tables), laneNextCardRule/laneOwnedCardRule, session_stale_run.go, launcher+internal/cli — all byte-identical (EV-5 probe + diff surface)
l44_pre_commit_fetch: n/a  # card worktree lane; no push per dispatch, leader batch-pushes
l44_post_push_fetch: n/a   # same — no push performed by this run
new_warnings_or_lints_introduced: 0  # go vet clean; golangci-lint deferred to CI verdict (lane lint uses CI golangci version)
cross_platform_build:
  windows_amd64: ok
total_run_phase_files: 3  # lane_spawn_authority.go, lane_spawn_authority_test.go, session_start_factory_test.go
m1_to_mn_commit_strategy: per-milestone commits (M1 fd7ae7ed4 / M2 dff54c2b9 / M3 f5583700c / M4 1958e6a18)
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29T21:43:58+09:00
sync_commit_sha: "8ba841146"  # a commit cannot cite its own hash; backfilled in the following commit
sync_status: complete
b12_self_test_a: pass  # grep -c 'SPEC-LANE-NOTICE-DIET-001' CHANGELOG.md → 0 before emission (no duplicate entry)
b12_self_test_b: pass  # distinct AC ids in acceptance.md = 7 (AC-LND-001..007); CHANGELOG entry cites 7/7 PASS
b12_self_test_c: pass  # all file paths cited in the CHANGELOG entry verified present (ls)
changelog_entry_position: "[Unreleased] › Changed"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (merged close, single sync commit)"
  plan_md: n/a  # plan.md frontmatter carries no status field
  acceptance_md: n/a  # acceptance.md frontmatter carries no status field; R2 whitelist line landed in this same commit
  progress_md: §E.4 written in this commit
canary_compliance_check:
  doc_surface_impact: none  # grep across README*.md + docs: the only hit (README.md:67) describes the notice's lane-model recommendation default, not the dieted spawn-authority prose; locale READMEs have 0 hits
  readme_edit: skipped  # nothing references the removed prose
  acceptance_body_edit_scope: R2 whitelist line only (run-phase-reported enumeration gap, sanctioned by the close contract)
  mx_tag_validation: pass  # lane_spawn_authority.go header comment already names the t1335 rewording; no new exported symbols
```

Sync-phase notes:

1. **Documentation impact check (skipped with evidence).** The change surface is
   an English-only internal hook const + its tests. `grep -r` across
   `README*.md` and `docs` for `laneSpawnAuthority|lane_spawn_authority|bootstrap
   notice|Bootstrap notice` hits only `README.md:67`, whose sentence describes
   the notice's lane-model recommendation default (leader/plan/run/sync backend
   mix) — content the diet did not touch. No doc edit is called for.
2. **CHANGELOG convention observed**: per-card entries under `[Unreleased]`
   (e.g. SPEC-PRIMARY-LOCALMD-RETIRE-001, SPEC-AUTONOMY-CLOSURE-001) — a
   per-card entry was added accordingly.
3. **R2 whitelist reconciliation**: the acceptance.md line 75 whitelist line
   (`session_start_factory_test.go`) landed in this same sync commit per the
   close contract; it reconciles the run-phase-reported enumeration gap (§E.2
   item 2) and touches no other acceptance body content.
