# SPEC-PROGRESS-RECORD-IO-001 — progress record

status: in-progress (M1 probe complete 2026-10-08, route (ii) measured — Q2 fired, milestone
stopped before M2; card t1598, base a2a184ad3)

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
- Ready for plan-audit: re-audit pending (scoped to the defect delta per the Retry Loop
  Contract).

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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
