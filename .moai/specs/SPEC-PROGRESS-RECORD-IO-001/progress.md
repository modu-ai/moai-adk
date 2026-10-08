# SPEC-PROGRESS-RECORD-IO-001 — progress record

status: in-progress (M1 probe + route-(ii) run body complete 2026-10-08 — M2 gate verified, M3
sweep + pinned disposition note landed, M4 CI protocol recorded; open items: F16 probe test
unauthored, AC-PRI-008 source-level note, section-b-stale-cells debt; card t1598, base a2a184ad3)

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
    immediate Lstat+fd-match pre/post checks (`progress_metadata_darwin.go:39-45, :52-58`). The
    microsecond gap between the post-copy re-check and `cp`'s own open remains a
    name-swap→symlink race in which **victim-overwrite stays possible** — named here as the
    LIVE residual harm class (route-(ii) honesty: naming it is required, not stale). The window
    cannot close on darwin without cgo (copyfile(3) is userspace; no fd-anchored ACL copy
    exists).
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
  dispatches NOTHING and pushes NOTHING** — the decisive runs arrive via the landing flow
  (card branch → local develop integration → leader batch push → release PR), and AC-CI-007's
  run URL + per-family executed/skipped counts are recorded at that point.
- **Recording rule** (restated for the landing flow): read the `-json` stream per family —
  decisive-PASS requires an explicit per-test pass Action for EVERY test in the family; a
  `skip` Action is recorded as SKIP and is NOT a PASS. Platform composition note from this
  round's measurement: the family's linux-tagged member
  (`TestAppendProgressRecordOverwritesInheritedDefaultAcl`, `//go:build linux`) executes ONLY
  on the linux leg — per-OS executed counts differ by exactly that member, and each OS's count
  is judged against its own platform's family set.
- No local darwin run is cited as evidence for GOOS-tagged families (REQ-CI-008); the darwin
  runs above are early local signals for the darwin leg only.

### Open items + run-phase debts (route-(ii) run body, 2026-10-08)

1. **`section-b-stale-cells` (N5, plan-audit-4) — OPEN.** plan.md §B :24/:27 carry route-(i)
   vocabulary residue ("The core work item (M1 probe → M2 fix)" on the F14 row; "Re-documented
   post-fix at the reduced harm class" on the residual row) against the operative route (ii).
   Disposal owner: **plan-phase artifact edit, lane-routed post-run** (manager-spec
   re-delegation) — this round does NOT edit plan.md or any plan-phase artifact. Non-operative
   (summary-table cells only; §A / M2 / M3 step 3 / AC-PRI-008 are correct per plan-audit-4).
2. **F16 close-hygiene probe unauthored — OPEN (orchestrator disposition).**
   `TestAppendProgressRecordSeedCloseHygiene` (plan M3 step 4; AC-PRI-005 scenario member;
   round-1 audit repair D4) has ZERO Go-code occurrences (repo-wide sweep this round:
   SPEC-artifact mentions only — plan.md:135, acceptance.md:78, progress.md:29; zero
   close-hygiene coverage under any other name). The landed F16 behavior (deferred close
   `audit_ceiling.go:676-684`; close-before-rename `:716-720`) therefore has NO committed
   regression guard, and both family commands above swept 6/7 named and 26/27 declared members
   respectively. Authoring it is a test-file addition inside internal/runtime — outside this
   round's records-only envelope; the runs above are decisive for the guards that exist. Needs
   an orchestrator decision: focused test-authoring re-delegation, landing-flow follow-up, or
   explicit debt disposition.
3. **AC-PRI-008 source-level re-documentation — OPEN (same envelope).** The seeder doc comment
   (`progress_metadata_darwin.go:11-20`) still carries the stale exec-exception framing
   (grep -c → 1, verbatim block above); the route-(ii) content keys are carried by the pinned
   note in §E.2. A comment-only edit inside internal/runtime was outside this round's envelope;
   the close-phase reviewer owns the pinned-content review (the `ac-content-enforcement` debt,
   plan-audit-4).
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
| AC-PRI-005 | PASS-WITH-DEBT | Formal run exit 0, 58 RUN, zero FAIL/SKIP, `-skip` anchor per N3; open member: close-hygiene probe unauthored (open item 2) |
| AC-PRI-006 | PASS | Package regression exit 0, ok ×2 (subsumes the AC's named command, stricter flags) |
| AC-CI-007 | PENDING (CI-only) | M4 protocol recorded; decisive runs from landing flow; verdict PENDING at record time |
| AC-PRI-008 | PASS-WITH-DEBT (record-level) | Pinned note (all content keys ≥1; victim-overwrite named live; stale framing absent from note); as-is source greps verbatim; source-level open (open item 3) |
| AC-PRI-009 | NOT-EVIDENCE (dispositioned) | Q2 record |

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
