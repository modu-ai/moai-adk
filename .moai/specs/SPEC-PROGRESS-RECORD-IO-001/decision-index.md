# SPEC-PROGRESS-RECORD-IO-001 — Decision Index

Authored at plan phase (card t1598, `interview.decision_gate: on`). Stateless on the status
axis — the SPEC lifecycle lives in `spec.md`. Every row states what is unresolved and why; no
row carries an embedded recommendation.

---

### Q1: Is the pure-Go fd-xattr route (the darwin analogue of the linux `system.posix_acl_access` blob) writable as non-root — and is it the route this SPEC adopts?

Label: EVIDENCE-NEEDED

Authority anchor: none — pending measurement (M1 probe, `plan.md` §F M1).

Why unresolved: the exact xattr name the kernel stores ACLs under on APFS, the blob layout, and
non-root fd-set writability via `golang.org/x/sys/unix` fd xattr APIs are unmeasured. The M1
probe exists to produce exactly this data; APFS may expose no ACL xattr at all, in which case
the route is dead and Q2 fires. The cgo ACL-API alternative is not a competing candidate — it is
excluded by policy (see Q1 note in `spec.md` REQ-PRI-005 and `tech.md` no-CGo clause).

Operator verdict:

---

### Q2: If the M1 probe measures no writable pure-Go route, how is F14 closed?

Label: FOUNDER

Class: product-level

Authority anchor: none — contingency branch; no prior SPEC or operator setting decides it.

Why unresolved: the branch changes the shipped CLI's security posture. Options: keep the current
exec-based darwin seeder and re-document the residual at its measured harm class, versus
private-directory staging with the held family's `PATH=""` constraint re-scoped and justified
(staging still shells out, so the family would weaken). The probe cannot answer which posture
the product should carry.

Default: keep the current seeder and re-document the residual (rule: the option that preserves
current behavior)

Alternate: private-directory staging with a justified re-scope of the held family

Operator verdict: LEADER-RULING (adopted) 2026-10-08 — option (a): keep the exec-based darwin
seeder and re-document the residual at its measured harm class. OVERRIDABLE by an operator
상위 전결; carrier: `progress.md` §E.2; M1 probe ad9ba32b2.

---

### Q3: Where do the linux/windows GOOS-tagged seeder families take their decisive verdict?

Label: EVIDENCE-NEEDED

Authority anchor: none — the decisive tagged runs have not been observed yet.

Why unresolved: the F9/F10/6b/6c/6d tagged families' decisive runs accumulate on the
release-pr-multi-os.yml 3-OS leg (per AC-CI-007); a local darwin run cannot execute GOOS-tagged
families by construction, so the verdict data does not exist until that workflow runs. M4
records the run URLs and per-family verdicts in `progress.md` §E.2 (AC-CI-007).

Operator verdict:

---

### Q4: Where does the post-fix kauth_filesec / ruling-(i) disposition live?

Label: FOUNDER

Class: implementation-level

Authority anchor: none — the disposition note today is a source comment on the base tree
(`internal/runtime/progress_metadata_darwin.go`); no register artifact decides its post-fix
home.

Why unresolved: the note can be re-documented in place within this SPEC (comment + SPEC docs) or
promoted into a follow-up card/SPEC that generalizes the darwin metadata route beyond
progress-record I/O. The in-place option preserves current behavior (the note already lives as a
comment), is single-revert, and has the smaller surface; a follow-up card is a queue admission —
the leader's/operator's act.

Default: in-place re-documentation inside this SPEC (rule: the option that preserves current behavior)

Alternate: issue a follow-up card for a general kauth_filesec route

Operator verdict: DEFAULT-APPLIED 2026-10-08T02:24:02Z manager-spec (lane-5, card t1598)

---

### Q5: The rename-permission residual (a write-allowed/delete-denied ACL on progress.md makes the atomic rename fail EPERM where the pre-repair in-place write succeeded — no §G record lands; `audit_ceiling.go:721`): documented residual, or a rename-failure→in-place-write fallback?

Label: FOUNDER

Class: implementation-level

Authority anchor: none — new in-card finding; the adopted Q2 ruling's behavior-preserving
posture is the nearest guidance but is itself a leader ruling, not a register artifact.

Why unresolved: implementing the fallback changes replace-path behavior (a new in-place write
mode on a rename EPERM); documenting keeps current behavior. The plan's scope discipline under
the adopted route (ii) points at documentation, but the fallback is a real future work item
shaped by the existing unwritable-dir/hardlink in-place path (`audit_ceiling.go:641-643`,
`:664`).

Default: documented residual, with the unwritable-dir/hardlink in-place fallback noted as the
future fix shape (rule: the option that preserves current behavior — consistent with the adopted
Q2 posture)

Alternate: add a REQ+AC for a rename-failure→in-place-write fallback (mirroring `:641-643`)

Operator verdict: DEFAULT-APPLIED 2026-10-08T06:21:09Z manager-spec (lane-5, card t1598)
