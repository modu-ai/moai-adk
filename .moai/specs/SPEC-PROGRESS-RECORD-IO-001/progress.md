# SPEC-PROGRESS-RECORD-IO-001 — progress record

status: in-progress (route-(ii) run body complete 2026-10-08 — M2 gate verified; M3 sweep, F16
close-hygiene probe (mutant-observed RED), and AC-PRI-008 source-level re-documentation landed;
M4 CI protocol recorded; open: section-b-stale-cells debt (plan-phase edit), AC-CI-007 (landing
flow); card t1598, base a2a184ad3)

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`, `decision-index.md`
  (Tier M set + decision index; `interview.decision_gate: on`).
- Tier: M (seeder rewrite + probe + family promotion; affected files < 8; REQ 9/16, AC 8/16
  within ceilings).
- Frontmatter: 12 canonical fields present; SPEC ID regex check PASS (verbatim `PASS` output
  cited in the authoring session); status `draft`.
- RED material: `.moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go` — the
  acceptance target, red-by-construction under `t.Setenv("PATH", "")` against the current
  exec-based darwin seeder; the in-package RED observation is scheduled at M2 entry (no test
  files written at plan phase).
- Scope correction recorded: F15/F16 measured ALREADY LANDED on a2a184ad3 (verification-only);
  F14 is the open P1 work item.
- Open decisions: `decision-index.md` Q1-Q3 unresolved (evidence/operator), Q4 default-applied.
- **Plan-audit round 1: FAIL 0.81** — verdict `.moai/reports/t1598/plan-audit-1.md` (audited
  SHA c404a0af4, artifact hash d262d9ad…dce6; blocking D1-D3, P2 D4-D7, P3 D8-D10). Revision
  round applied the full defect list without restructuring (MP-1~MP-9 clean results preserved):
  D1 AC-PRI-009 real-seeder fd-anchoring guard via the `seedFileMetadataFn` seam (mid-seed name
  swap, victim untouched; mutant-killer for path-based re-open regressions); D2 route-(ii)
  branch-conditional disposition in acceptance §D + conditional DoD in §F; D3 decisive surface
  re-pinned to `release-pr-multi-os.yml` (release/*→main PR or workflow_dispatch) with
  `-json` SKIP-vs-PASS recording rule, card-PR premise dropped (measured: ci.yml ubuntu-only,
  gate excludes internal/runtime); D4 close-hygiene probe `TestAppendProgressRecordSeedCloseHygiene`
  wired into M3 + AC-PRI-005; D5/D8 structured probe headings pinned in M1 step 6 and gated on
  the §E.2 content carrier; D6 AC-PRI-008 stale `victim-overwrite` text absence; D7 no-SKIP
  decisive runs for AC-PRI-003/005; D9 build-tag clause on the probe bullet; D10 t1560 citation
  location qualifier. REQ coverage now complete: REQ-PRI-004 → AC-PRI-009.
- **Plan-audit round 2: PASS-WITH-DEBT 0.94** (D1-D10 resolved; debts N1/N2/N3) — verdict
  `.moai/reports/t1598/plan-audit-2.md`.
- **Route-(ii) re-scope amendment (this commit)** — decision-index Q2 ADOPTED (leader ruling
  (a), overridable by an operator 상위 전결; §E.2 carrier; M1 probe ad9ba32b2):
  AC-PRI-003/004/009 dispositioned NOT-EVIDENCE; AC-PRI-005 pinned to the `-skip` anchor on
  route (ii) (N3); N1/N2 stale git-model prose corrected to the measured GitHub-Flow regime
  (committed `AGENTS.local.md` §4.1, 2026-10-05 transition); the write-allowed/delete-denied
  rename residual (`audit_ceiling.go:721`) absorbed as documentation (decision-index Q5,
  default-applied). MP-1~MP-9 clean results preserved. AC-PRI-008 made branch-conditional (the
  `victim-overwrite` → 0 predicate is route-(i)-only; route (ii) requires the honest harm class
  named).
- **Plan-audit round 3: FAIL 0.94 — single blocking finding N4** (`.moai/reports/t1598/plan-audit-3.md`,
  HEAD e2eaaa918): the amendment's cross-layer sweep missed plan.md itself — §A Context and
  §F M2 still instructed the route-(i) work verbatim (Go-native no-exec seeder; AC-PRI-009
  wiring; the measured-unachievable family+grep gate). N4 fix applied: §A corrected to the
  measured state (route (ii) adopted via Q2; probe ad9ba32b2 measured no writable pure-Go
  fd-xattr route; honest residuals named), §F M2 re-scoped to the route-(ii) disposition record
  + F15/F16 guard maintenance under the `-skip` anchor (no implementation, no promotion, no
  exec-count grep; the implementation ACs dispositioned NOT-EVIDENCE, not restated as tasks).
  Everything else verified sound by round 3 — untouched.
- Ready for plan-audit: plan-audit-4 delta re-audit pending (N4 hunks only, before run
  re-entry).

## §E.2 Run-phase Evidence

### M1 — darwin fd-xattr route probe (2026-10-08, tree f60b42fa8)

- Probe artifact: `.moai/state/verify/t1598/probe-darwin-fd-xattr.md` (local, gitignored by
  design; the §E.2 carrier below is the committed evidence). Probe program sha256
  `b4da207ab35264db4408477aefca1dcd267ee34eb08ca24ab4b7ca64e2c37321`; platform Darwin 27.0.0
  (APFS), uid 501 non-root. Grep keys present: `## xattr name` / `## blob layout` /
  `## fd-set result` / `## fork decision` (4/4).
- **Fork decision: `route (ii)`** — the pure-Go fd-xattr route is measured NOT writable as
  non-root. Decisive raw output (verbatim, from the probe run against an ACL-carrying file —
  positive control `chmod +a "group:_guest deny read"` seeded and visible first):

```
== raw listxattr sweep (path form) ==
names(1): ["com.apple.provenance"]
== raw listxattr sweep (fd form: Flistxattr) ==
names(1): ["com.apple.provenance"]
```

  plus the candidate-name fd-set attempts (non-root, held fd):

```
  Fsetxattr kauth-filesec                              -> err=<nil>
  Fsetxattr com.apple.system.kauth_filesec             -> err=operation not permitted
```

  A nil-error write to the inert `kauth-filesec` name leaves the file's enforced ACL unchanged
  (`ls -le` still `0: group:_guest deny read`) — the kernel ignores these names on APFS; there
  is no ACL-carrying xattr to copy and no writable one that would carry semantics. The
  `com.apple.system.*` namespace is EPERM-gated for non-root on both get and set. cgo ACL APIs
  excluded by policy (`tech.md` no-CGo clause); `clonefileat`/`SYS_COPYFILE`/`setattrlist`
  measured-rejected in t1560 (plan §G anti-pattern — not re-opened, no new evidence).
- **decision-index Q2 (FOUNDER) FIRES.** Per plan M1 step 5 the milestone STOPS here: no M2
  implementation, no held-family promotion, no plan/acceptance artifact edit from run phase.
  The re-scope amendment (and the route-(ii) AC-PRI-005 anchor one-liner, N3/debt-3 disposal)
  belongs to manager-spec; the F14 closure posture (Q2 default: keep the exec seeder +
  re-document the residual at the measured harm class; alternate: private-dir staging with a
  justified held-family re-scope) is the plan-phase owner's decision.
- RED material preservation (codex attempt-5 P2-d, recorded before any re-scope): held family
  `.moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go` sha256
  `7b97c796d0243774a3781c52db9e0bd2463787f7ae30026c5d35a887a64ff381`; it was never observed RED
  in-package in this session (M2 entry did not happen — the fork stops the milestone first);
  the plan-audit round-1 RED re-execution (overlay, exit 1, `exec: "chmod": executable file not
  found in $PATH`) was taken on c404a0af4 and carries per `.moai/reports/t1598/plan-audit-2.md`
  evidence item 1 (docs-only delta between c404a0af4 and f60b42fa8).
- M2/M3/M4: **not executed** — blocked on the Q2 verdict. AC-PRI-002/003/004/009 evidence is
  route-conditional and intentionally absent here.

### Q2 verdict record — route (ii) ADOPTED (plan-phase amendment, 2026-10-08)

Recorded by manager-spec on explicit coordinator re-delegation; this sub-section is the §E.2
carrier for the dispositioned ACs. The M1 evidence above is manager-develop's, untouched.

- **Ruling (a), ADOPTED**: keep the current exec-based darwin seeder; re-document the residual
  at its measured harm class.
- **Status: LEADER ruling (card t1598, 2026-10-08) — OVERRIDABLE by an operator 상위 전결.** An
  operator ruling reverses to route (i) and re-opens the dispositioned ACs.
- **M1 probe citation: ad9ba32b2** ("M1 darwin probe — route (ii) measured, Q2 fired").
- Dispositioned under route (ii): AC-PRI-003 (family GREEN), AC-PRI-004 (no-exec grep), and
  AC-PRI-009 (post-fix fd-anchoring guard) → NOT-EVIDENCE; this record is their replacement
  carrier.
- AC-PRI-002's RED evidence: the plan-audit round-1 overlay re-execution (exit 1,
  `exec: "chmod": executable file not found in $PATH`, tree c404a0af4, plan-audit-1.md evidence
  item 1) — the in-package landing is dispositioned with the family (a permanently-red test is
  not committed).
- AC-PRI-005 on route (ii) runs with `-skip '^TestAppendProgressRecordPreservesAllMetadataAxes$'`
  (N3 disposal).

### M2 — route-(ii) gate verification (2026-10-08, measurements at tree 6abc63d30)

- **Q2 disposition record: verified present in §E.2** — the sub-section "Q2 verdict record —
  route (ii) ADOPTED (plan-phase amendment, 2026-10-08)" above carries ruling (a), the
  OVERRIDABLE status, the M1 probe citation ad9ba32b2, and the dispositioned set
  (AC-PRI-003/004/009 NOT-EVIDENCE). M2 cites it; it is not re-authored here.
- **Held-family content hash + tree SHA duty (round-2 debt carrier) — both halves now
  recorded**: `shasum -a 256 .moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go`
  → `7b97c796d0243774a3781c52db9e0bd2463787f7ae30026c5d35a887a64ff381` (exit 0). Byte-identical
  to M1's recorded value (M1 pinned tree f60b42fa8); re-verified this round at tree 6abc63d30 —
  the file is unchanged across the plan-phase delta, and this re-verification tree SHA
  (6abc63d30) completes the hash+tree pair.
- **F15/F16 guard family (dispatch command, decisive)**:
  `go test -race -count=1 -v -run 'TestAppendProgressRecordSwapKeepsForeignFile|TestAppendProgressRecordTempSwapFailsClosed|TestAppendProgressRecordPreservesACL|TestAppendProgressRecordPreservesXattr|TestAppendProgressRecordAclExactlyOriginal|TestAppendProgressRecordSeedCloseHygiene|TestAppendProgressRecordNewFileModeAppliesUmask' ./internal/runtime`
  → exit 0; `ok  	github.com/modu-ai/moai-adk/internal/runtime	2.127s`; **6 RUN / 6 PASS /
  ZERO SKIP lines**. Swept-set accounting (verification-completeness §1.1): the selector names
  7 tests, 6 matched — `TestAppendProgressRecordSeedCloseHygiene` matched **zero** tests (the
  name exists only in SPEC artifacts; see open items). The run is decisive for the 6 swept
  guards and is NOT evidence for the unauthored probe. Full log:
  `evidence/m2-guard-family-r1.txt`.
- **Formal AC-PRI-005 decisive run (plan M2 guard-maintenance / M3 step 1 command)**:
  `go test -race -count=2 -v -run '^TestAppendProgressRecord' -skip '^TestAppendProgressRecordPreservesAllMetadataAxes$' ./internal/runtime/`
  → exit 0; `ok  	github.com/modu-ai/moai-adk/internal/runtime	3.882s`; **58 RUN (26 darwin
  parent tests × count=2 + 3 subtests × 2), ZERO FAIL, ZERO SKIP**. The `-skip` anchor excluded
  only the dispositioned RED family (zero matches — not committed, per the Q2 record). Platform
  note: the family's 27th declared member `TestAppendProgressRecordOverwritesInheritedDefaultAcl`
  is `//go:build linux`-tagged (`progress_metadata_linux_test.go:1`) and structurally cannot
  execute on darwin — its decisive verdict is CI-owned (see M4). Full log:
  `evidence/m3-family-formal-r2.txt`.

### M3 — regression sweep + pinned disposition note (2026-10-08, measurements at tree 6abc63d30)

- **Package regression (affected packages only)**:
  `go test -race -count=1 -timeout 30m ./internal/runtime/...` → exit 0;
  `ok  	github.com/modu-ai/moai-adk/internal/runtime	20.320s` +
  `ok  	github.com/modu-ai/moai-adk/internal/runtime/gobin	1.535s`; ZERO SKIP. Subsumes
  AC-PRI-006's named command (`go test -timeout 30m ./internal/runtime/`) — same suite under
  stricter flags (`-race -count=1`) plus the gobin subpackage. Full log:
  `evidence/m3-package-regression-r3.txt`.
- **go vet matrix** (each exit 0, no output): `go vet ./internal/runtime/...` (darwin);
  `GOOS=windows GOARCH=amd64 go vet ./internal/runtime/...`;
  `GOOS=linux GOARCH=amd64 go vet ./internal/runtime/...`.
- **gofmt**: `gofmt -l internal/runtime/` → empty list, exit 0 — all formatted. NO files touched
  this round (records phase).
- **Pinned disposition note (AC-PRI-008 route-(ii) branch — §E.2 carrier)**. The exec-based
  darwin seeder is KEPT (Q2 ruling (a)); the following is the operative route-(ii) record:
  - **Residual window `fd-verify→rename`**: the seeder's path-based steps (`chmod -N`, `cp -p`)
    sit behind the held-fd pre-checks (`audit_ceiling.go:685/697/712`) and the seeder's own
    immediate Lstat+fd-match pre/post checks (`progress_metadata_darwin.go`). The unprotected
    swap window runs from the seeder's immediate pre-seed verification to `cp`'s own open — a
    name-swap→symlink race inside it means **victim-overwrite stays possible**; the post-copy
    re-check bounds the post-copy span before the rename and fails closed. Named here as the
    LIVE residual harm class (route-(ii) honesty: naming it is required, not stale). The
    window cannot close on darwin without cgo (copyfile(3) is userspace; no fd-anchored ACL
    copy exists).
  - **kauth_filesec follow-up disposition**: M1 (ad9ba32b2) measured the inert
    `kauth_filesec`-class name — a nil-error fd-set write to it leaves the enforced ACL
    unchanged (the kernel ignores it on APFS) — and the EPERM-gated `com.apple.system.*`
    namespace for non-root. A kauth_filesec reimplementation remains a follow-up-SPEC candidate
    (source :18); excluded from this SPEC by decision-index Q4.
  - **Rename-permission residual (`delete-denied`)**: a write-allowed/**delete-denied** ACL on
    progress.md makes the atomic rename (`audit_ceiling.go:721`) fail EPERM where the
    pre-repair in-place `os.WriteFile` (`:664`) succeeded, so no §G record lands. The
    unwritable-dir/hardlink in-place fallback (`:641-643`, `:664`) is the noted future fix
    shape (decision-index Q5, default-applied).
  - **Route**: `route (ii)` is the adopted branch (decision-index Q2, leader ruling (a),
    OVERRIDABLE by an operator 상위 전결; probe citation ad9ba32b2; §E.2 Q2 record above).
- **AC-PRI-008 as-is source measurement** (the AC's greps against the kept seeder, verbatim —
  the source comment is NOT edited this round; records-phase envelope. Command [1]'s pattern is
  elided here to keep this §E.2 free of the stale token per the route-(ii) →0 predicate; the
  byte-exact command text is acceptance.md §C AC-PRI-008 command [1], re-runnable as written):
  ```
  $ grep -c "<acceptance.md AC-PRI-008 command [1] pattern: the stale exec-exception token>" internal/runtime/progress_metadata_darwin.go
  1                (exit 0 — stale framing still in source at :13)
  $ grep -n "fd-verify" internal/runtime/progress_metadata_darwin.go
  (no output, exit 1 — key absent)
  $ grep -n "kauth_filesec" internal/runtime/progress_metadata_darwin.go
  18:// seeder on darwin; a kauth_filesec reimplementation is a follow-up-SPEC
  $ grep -n "delete-denied" internal/runtime/progress_metadata_darwin.go
  (no output, exit 1 — key absent)
  $ grep -n "route (ii)" internal/runtime/progress_metadata_darwin.go
  (no output, exit 1 — key absent)
  $ grep -c "victim-overwrite" internal/runtime/progress_metadata_darwin.go
  1                (exit 0 — the F13-flagged residual at :38)
  ```
  The source-level predicate set (keys ≥1 each, stale count → 0) is therefore NOT met in source
  as-is; the pinned note above carries the route-(ii) content at the record level — which is
  the DoD's named evidence surface for AC-PRI-008 ("carry observed evidence in progress.md
  §E.2"). Source-comment re-documentation stays OPEN (see open items).

### M4 — GOOS-tagged CI observation protocol (recorded 2026-10-08, tree 6abc63d30; no runs dispatched)

- **Decisive surface** (plan M4 / AC-CI-007): the `release-pr-multi-os.yml` 3-OS leg —
  `go test -json -race -timeout 35m ./...` (linux+macos+windows). Surface verified present this
  round: `.github/workflows/release-pr-multi-os.yml` exists; `workflow_dispatch:` trigger at
  :16; the json/race/35m leg at :217. The justifying exclusions stand as measured in plan
  phase: `ci.yml` is ubuntu-only (:94) and `pr-multi-os-gate.yml` tests only hook +
  cli/worktree packages (:112) — no card-level CI surface runs the GOOS-tagged runtime
  families.
- **Trigger**: a `release/*`→`main` PR, or `workflow_dispatch` on the card branch. **This round
  dispatches NOTHING and pushes NOTHING** — the decisive runs arrive via the main-based landing
  flow (card branch → card PR merge to `main`; the CI subject is the `release-pr-multi-os.yml`
  leg — a `release/*`→`main` PR or a `workflow_dispatch` on the branch), and AC-CI-007's run URL
  + per-family executed/skipped counts are recorded at that point. (Sync-phase wording fix, card
  t1598 §E.4: the retired develop-integration citation removed per plan-audit-5's P2; the
  plan-audit-4 hash binding is unaffected — progress.md is a non-hash-subject.)
- **Recording rule** (restated for the landing flow): read the `-json` stream per family —
  decisive-PASS requires an explicit per-test pass Action for EVERY test in the family; a
  `skip` Action is recorded as SKIP and is NOT a PASS. Platform composition note from this
  round's measurement: the family's linux-tagged member
  (`TestAppendProgressRecordOverwritesInheritedDefaultAcl`, `//go:build linux`) executes ONLY
  on the linux leg — per-OS executed counts differ by exactly that member, and each OS's count
  is judged against its own platform's family set.
- No local darwin run is cited as evidence for GOOS-tagged families (REQ-CI-008); the darwin
  runs above are early local signals for the darwin leg only.

### F16 close-hygiene probe + AC-PRI-008 source re-documentation (2026-10-08; coordinator-authorized after 3f1f942b4)

Both items were outside the records-only run body's envelope and were explicitly authorized by
the coordinator. Measurements ran on the working tree = 3f1f942b4 + the two Go changes this
commit carries (byte-identical to this commit's tree).

- **F16 probe authored**: `internal/runtime/audit_ceiling_close_hygiene_test.go`
  (`//go:build darwin || linux`, matching `audit_ceiling_replace_test.go`) —
  `TestAppendProgressRecordSeedCloseHygiene`, two subtests: a `seedFileMetadataFn` wrapper
  captures the held descriptor and delegates to the real seeder (success path — the replace
  completes) or injects `errSeedInjected` (abort path — the deferred close is the only closer);
  after `appendProgressRecord` returns, a second `Close` must satisfy
  `errors.Is(err, os.ErrClosed)`.
- **Observed-failure completion (verification-completeness §1.1)** — born-green is not
  completion: with both closes temporarily dropped (the deferred `:677` close AND the
  close-before-rename `:718` block), the probe ran RED — exit 1, BOTH subtests FAIL:
  `audit_ceiling_close_hygiene_test.go:45: the held descriptor survived the successful replace
  (second Close: <nil>)` and `:78: … the held descriptor survived the aborted replace (second
  Close: <nil>) …`. The mutations were reverted byte-identically (`git diff --stat
  internal/runtime/audit_ceiling.go` → empty) and the probe re-ran green on both subtests.
  Non-vacuous on both paths. Evidence: `evidence/f16-probe-mutant-red.txt`.
- **AC-PRI-005 formal family re-run (now with the probe)**:
  `go test -race -count=2 -v -run '^TestAppendProgressRecord' -skip '^TestAppendProgressRecordPreservesAllMetadataAxes$' ./internal/runtime/`
  → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/runtime	4.039s`; **64 RUN (27 darwin
  parents + 5 subtests) × 2, ZERO FAIL, ZERO SKIP**. The M2 curated 7-name selector re-run →
  exit 0, `ok  	github.com/modu-ai/moai-adk/internal/runtime	1.544s`, 9 RUN (7/7 parents incl.
  the new probe + 2 probe subtests), ZERO SKIP — the previous 6/7 swept-set gap is closed.
  Evidence: `evidence/r4-family-formal.txt`, `evidence/r4-m2-guard-family-7of7.txt`.
- **AC-PRI-008 source-level re-documentation landed** (comment-only; zero behavior change; the
  pre/post Lstat+fdMatchesName checks untouched): the seeder header + F13-paragraph trailing
  sentence re-documented under the ADOPTED Q2 ruling (a) — the exec-based seeder KEPT
  (OVERRIDABLE by an operator ruling), the `fd-verify→rename` window named LIVE
  (victim-overwrite stays possible), the `delete-denied` rename residual (decision-index Q5),
  kauth_filesec as the named future-fix direction. Re-measurement (verbatim):
  ```
  [1] stale exec-exception token count → 0 (exit 1, zero rows)   — was 1 (:13)
  [2] fd-verify       → 2 hits (:23, :57)                        — was 0
  [3] kauth_filesec   → 1 hit  (:34)                             — was 1 (:18), retained
  [4] delete-denied   → 2 hits (:28-29)                          — was 0
  [5] route (ii)      → 2 hits (:22, :55)                        — was 0
  [6] victim-overwrite → 2 hits (:25, :56), named LIVE           — honest residual kept
  ```
  All AC-PRI-008 route-(ii) source predicates now MET; M3 step 3(d)'s delete-denied note is
  landed in source and stands in the §E.2 pinned note above.
- **Package regression re-run**: `go test -race -count=1 -timeout 30m ./internal/runtime/...`
  → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/runtime	26.548s` +
  `ok  	github.com/modu-ai/moai-adk/internal/runtime/gobin	1.875s`, ZERO SKIP. Evidence:
  `evidence/r4-package-regression.txt`.
- **vet + gofmt**: darwin / GOOS=windows / GOOS=linux `go vet ./internal/runtime/...` all exit
  0; `gofmt -l internal/runtime/` → empty (both touched files formatted).

### Open items + run-phase debts (route-(ii) run body, 2026-10-08)

1. **`section-b-stale-cells` (N5, plan-audit-4) — OPEN.** plan.md §B :24/:27 carry route-(i)
   vocabulary residue ("The core work item (M1 probe → M2 fix)" on the F14 row; "Re-documented
   post-fix at the reduced harm class" on the residual row) against the operative route (ii).
   Disposal owner: **plan-phase artifact edit, lane-routed post-run** (manager-spec
   re-delegation) — this round does NOT edit plan.md or any plan-phase artifact. Non-operative
   (summary-table cells only; §A / M2 / M3 step 3 / AC-PRI-008 are correct per plan-audit-4).
2. **F16 close-hygiene probe — RESOLVED** (coordinator-authorized follow-up; the F16
   sub-section above). `TestAppendProgressRecordSeedCloseHygiene` authored at
   `internal/runtime/audit_ceiling_close_hygiene_test.go` (`darwin || linux`), both subtests
   green on the current tree, its RED observed on the double-close-drop mutant before the
   byte-identical revert (verification-completeness §1.1). The family now sweeps 7/7 curated
   names and 27/27 darwin parents; the linux leg stays CI-owned (M4).
3. **AC-PRI-008 source-level re-documentation — RESOLVED** (coordinator-authorized follow-up;
   the F16 sub-section above). The seeder comment is re-documented under the adopted Q2 ruling
   (a); the stale exec-exception token count is 0 (was 1 at :13) and all five route-(ii)
   content keys measure ≥1 in source (verbatim re-measurement above). The close-phase reviewer
   still owns the pinned-content review (the `ac-content-enforcement` debt, plan-audit-4).
4. **AC-CI-007 — PENDING the landing flow** (M4 protocol above). The repository-wide test
   verdict is owned by CI on the project's integration branch and is PENDING at this record
   time.

**Route-(ii) AC status matrix (this round's evidence)**:

| AC | Status under route (ii) | Evidence (this §E.2) |
|----|-------------------------|----------------------|
| AC-PRI-001 | PASS (standing) | M1 record: probe artifact + fork decision `route (ii)` + verbatim decisive output (landed ad9ba32b2) |
| AC-PRI-002 | PASS (standing — recorded measurement) | Q2 record: overlay RED exit 1 (`exec: "chmod": executable file not found in $PATH`), tree c404a0af4 |
| AC-PRI-003 | NOT-EVIDENCE (dispositioned) | Q2 record |
| AC-PRI-004 | NOT-EVIDENCE (dispositioned) | Q2 record |
| AC-PRI-005 | PASS | Formal re-run exit 0, 64 RUN (27/27 darwin parents incl. the probe + 5 subtests), zero FAIL/SKIP; probe RED observed on the drop-both-closes mutant; linux leg CI-owned (M4) |
| AC-PRI-006 | PASS | Package regression exit 0, ok ×2 (subsumes the AC's named command, stricter flags) |
| AC-CI-007 | PENDING (CI-only) | M4 protocol recorded; decisive runs from landing flow; verdict PENDING at record time |
| AC-PRI-008 | PASS | Source-level predicates met: stale token 0 (was 1); fd-verify / kauth_filesec / delete-denied / route (ii) / victim-overwrite (LIVE) all ≥1 in source (verbatim re-measurement); §E.2 pinned note stands |
| AC-PRI-009 | NOT-EVIDENCE (dispositioned) | Q2 record |

## §E.3 Run-phase Audit-Ready Signal

run_status: complete (route (ii) adopted; substance commits 3f1f942b4 + e6eb6da2f —
AC-PRI-005/006/008 PASS, AC-PRI-003/004/009 dispositioned via the §E.2 Q2 record,
AC-CI-007 landing-flow pending)
governing_audit: plan-audit-4 + plan-audit-5 (PASS-WITH-DEBT 0.94; the round-5 delta
re-affirmation carries a zero hash-subject diff through a2fcca6c3; receipt
rcpt-5bbbb0b8650328de4d6b115f)
operator_override: §G row-level ceiling override recorded at a2fcca6c3

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete (3-phase close — route (ii) operative, card t1598)
sync_complete_at: 2026-10-08T12:44:19Z
sync_commit_sha: 8d10007ca9804eeb58bcb0a34aca922989603198
governing_audit: plan-audit-4 PASS-WITH-DEBT 0.94 (audited_sha 6abc63d30e4bb133fead8fc24b5b1013c00581a9, must_pass_failed=0, blocking_count=0) — re-affirmed unchanged by plan-audit-5 (receipt rcpt-5bbbb0b8650328de4d6b115f; the seven hash-subject paths carry a zero diff 6abc63d30..a2fcca6c3)
frontmatter_status_transitions:
  in-progress→implemented→completed: manager-docs, the single sync commit (merged close per the Status Transition Ownership Matrix; trailer Authored-By-Agent: manager-docs)
changelog_entry_position: CHANGELOG.md `## [Unreleased]` → `### Added` → first bullet
b12_self_test_a: PASS — `grep -c "SPEC-PROGRESS-RECORD-IO-001" CHANGELOG.md` → 0 (exit 1) before emission (no duplicate from a parallel BATCH-SYNC session)
b12_self_test_b: PASS — live AC count = 9 (AC-PRI-001..006, 008, 009 + AC-CI-007; reserved-token marks 0, ambiguous=0); the CHANGELOG entry references the same 9-criterion set with the route-(ii) disposition
b12_self_test_c: PASS — every path claimed in the CHANGELOG entry ls-verified this round (internal/runtime/{audit_ceiling.go, progress_metadata_darwin.go, progress_fd_unix.go, progress_metadata_acl.go, audit_ceiling_close_hygiene_test.go, audit_ceiling_acl_test.go, audit_ceiling_umask_test.go, audit_ceiling_replace_test.go, progress_metadata_acl_test.go, progress_metadata_linux_test.go}, .moai/specs/SPEC-PROGRESS-RECORD-IO-001/progress.md)
canary_compliance_check: n/a — this SPEC defines no forward-looking policy exercised by its own sync tests
mx_tag_validation: 1 @MX:DEBT added on seedFileMetadata (internal/runtime/progress_metadata_darwin.go) with @MX:CEILING + @MX:UPGRADE + @MX:SPEC sub-lines — the route-(ii) kept exec seeder recorded as a deliberate working simplification; 0 tags removed; audit_ceiling.go's 3 existing tags validated unchanged (at the per-file ANCHOR cap); the new probe test file needs no tag (test-only surface)

AC final matrix (route-(ii) close posture — evidence in §E.2): AC-PRI-001 PASS (standing) ·
AC-PRI-002 PASS (standing — recorded measurement) · AC-PRI-005 PASS · AC-PRI-006 PASS ·
AC-PRI-008 PASS · AC-PRI-003/004/009 NOT-EVIDENCE (dispositioned by the §E.2 Q2 record) ·
AC-CI-007 PENDING (CI-only — the `release-pr-multi-os.yml` decisive runs ride the landing flow;
run URL + per-family counts recorded at that point).

Open-items ledger (honest carry-forward at close):

1. `section-b-stale-cells` (N5, plan-audit-4) — OPEN. plan.md §B :24/:27 two-cell
   route-(i)-vocabulary residue; the disposal owner is a plan-phase artifact edit (manager-spec
   re-delegation, lane-routed post-run) — non-operative summary cells; §A / M2 / M3 step 3 /
   AC-PRI-008 are correct per plan-audit-4.
2. AC-CI-007 — PENDING the landing flow (CI-only; M4 protocol above). The repository-wide
   verdict is owned by CI on the project's integration branch and is PENDING at this record.
3. The receipt-fence admission wedge (plan-audit-4 Gaps-4 / plan-audit-5 Gaps-3 — the
   Start-marker advance on refusal-spawn, `internal/auditreceipt/store.go:1254` vs
   `internal/hook/audit_receipt_guard.go:153`) and the ceiling §G-release reader gap —
   issuance-ledger items for the operator/leader queue, not sync-phase work.
4. progress.md §E.2 M4 Trigger wording (plan-audit-5's backend P2 against :228 — the retired
   develop-integration citation) — FIXED this close in the §E.4-adjacent pass (re-worded to the
   main-based landing + release-pr-multi-os/dispatch CI subject). The plan-audit-4 hash binding
   is unaffected: progress.md is a non-hash-subject run-phase record (plan-audit-5 evidence
   item 2).
5. `ac-content-enforcement` debt (plan-audit-4, reviewer-enforced) — carried: the close-phase
   reviewer owns the AC-PRI-008 pinned-content review; AC-PRI-001's §E.2 verbatim-quote duty is
   met by the M1 record (the decisive raw output quoted above).

§G ceiling-release record: intact and untouched by this close — the 4 ceiling-refusal holds, the
LEADER-RULING release (receipts rcpt-ce322e7663362a796bf14bde + rcpt-f8998a6b0b8f1acd2b23d016),
and the OPERATOR row-level override row stand as recorded.

## §G Override and Refusal Record

- 2026-10-08T12:11:08Z SPEC-PROGRESS-RECORD-IO-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-3.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-4.md
- 2026-10-08T12:11:19Z SPEC-PROGRESS-RECORD-IO-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-3.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-4.md
- 2026-10-08T12:13:30Z SPEC-PROGRESS-RECORD-IO-001 ceiling-release outcome=release decided_by="LEADER-RULING (operator-overridable 상위 전결, delivered 2026-10-08 ~20:05 KST)" reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2) but the governing verdict .moai/reports/t1598/plan-audit-4.md is PASS-WITH-DEBT 0.94 with must_pass_failed=0 and blocking_count=0, audited_sha 6abc63d30e4bb133fead8fc24b5b1013c00581a9; LEADER-RULING ① admits run-phase entry on the recorded evidence (leader personally read the verdict file; receipts rcpt-ce322e7663362a796bf14bde + rcpt-f8998a6b0b8f1acd2b23d016; the receipt-fence wedge is an admission-infra defect, not an audit failure — guard repair routed to the operator issuance queue)" evidence=.moai/reports/t1598/plan-audit-4.md
- 2026-10-08T12:15:00Z SPEC-PROGRESS-RECORD-IO-001 ceiling-release outcome=release decided_by="OPERATOR DECISION (AskUserQuestion response, relayed by leader 2026-10-08)" reasons="row-level override of the REQ-ACE-006 ceiling HOLD for plan-audit round 4 (verdict PASS-WITH-DEBT 0.94, audited_sha 6abc63d30); leader applied factory row override + T8 approve" evidence=.moai/reports/t1598/plan-audit-4.md
- 2026-10-08T12:12:48Z SPEC-PROGRESS-RECORD-IO-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-3.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-4.md
- 2026-10-08T12:13:03Z SPEC-PROGRESS-RECORD-IO-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-3.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1598/.moai/reports/t1598/plan-audit-4.md
