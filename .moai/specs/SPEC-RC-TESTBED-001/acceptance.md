# acceptance.md — SPEC-RC-TESTBED-001

Tier M. All ACs are two-cell pairs per `.claude/rules/moai/development/verification-completeness.md`
§2 (RED-now observed on the pre-work tree + green path naming the flipping milestone).

**Document-level pin:** the RED-now cells were first measured on tree **`fa8ff89ba`** (branch
`WT-rc-testbed`, card t281 worktree). Following the D1 resolution absorb (origin/develop into
this branch at `a04afea53`, which replaced CLAUDE.local.md §4.1 with the landed 08-29 chain),
the AC-RC-007 anchors were **re-measured on `a04afea53`** — both still 0, exit 1 — and that
re-measurement supersedes the pre-absorb basis for CLAUDE.local.md. The other doc targets
(version-management.md, gitflow-lane-protocol.md) were untouched in `fa8ff89ba..6b03e1757`, so
their `fa8ff89ba` pins hold unchanged.

**Empty-sweep convention (§1.1).** For every AC here the RED-now observation is a **0-hit grep**:
stdout `0`, exit code `1`. That 0 is the *expected starting red* — the rule is missing today —
NOT a vacuous pass. The green form asserts stdout ≥1 AND exit 0. The swept set is the single
named file in each command (no selector ambiguity, no cross-file bleed); a 0 therefore always
means "token absent from the one file swept", and a nonzero count can only come from that file.

**Mutant probes performed before adoption (§2).** Per-criterion probes are recorded in the AC
blocks; the load-bearing result: the card's representative mutant — *declare develop standing
without documenting the rc build + clean reinstall procedure* — satisfies AC-001..006 only by
performing the actual documentation work, and leaves **AC-007 red** (the CLAUDE.local.md §4.1
pointer is exactly what the mutant skips). AC-007 is adopted as the mutant-killer. Residual risk
(token-stuffing mutants that embed all anchor literals while violating the policy's substance)
is accepted and routed to sync-audit human reading — noted per AC cluster, not hidden.

## §D AC Matrix

| AC | REQ | Criterion | Flipped by |
|---|---|---|---|
| AC-RC-001 | REQ-RC-001 | Local RC Numbering section exists (version-management.md) | M1 |
| AC-RC-002 | REQ-RC-002 | counter-precedent record exists (version-management.md) | M1 |
| AC-RC-003 | REQ-RC-003 | BUILD_ID ordering warning exists (version-management.md) | M1 |
| AC-RC-004 | REQ-RC-004 | develop 갱신 section exists (gitflow-lane-protocol.md) | M2 |
| AC-RC-005 | REQ-RC-004 | squash-blindness citation exists (gitflow-lane-protocol.md) | M2 |
| AC-RC-006 | REQ-RC-005 | BranchGuard route note exists (gitflow-lane-protocol.md) | M2 |
| AC-RC-007 | REQ-RC-006/007 | §4.1 pointers to numbering policy + refresh section exist (CLAUDE.local.md) | M3 |
| AC-RC-008 | REQ-RC-008 | rc_version_format re-application note exists (version-management.md) | M1 |

## §D.1 Release-blocking ACs (evidence-ledger form)

**AC-RC-001 — Local RC numbering policy section exists.**
Given the pre-work tree `fa8ff89ba`, When a reader searches version-management.md for the
numbering-policy section, Then the section is absent (red now); after M1 the same search finds
it (≥1, exit 0).
- RED-now command: `grep -c "Local RC Numbering" .moai/docs/version-management.md`
- verbatim stdout: `0` · exit code: `1` · tree: `fa8ff89ba`
- Green path (M1): same command → stdout ≥1, exit 0.
- Mutant probe: a heading-only section (heading present, no policy body) satisfies this AC
  alone → defeated by pairing with AC-002/003 content tokens (plan §D prescribes the four
  mandatory content elements); adopted as a cluster with them.

**AC-RC-002 — Tag-exclusion rule + counter-precedent recorded.**
Given `fa8ff89ba`, When the no-tags rule for local rc builds is searched, Then no counter-
precedent record exists; after M1 the v3.1.0-rc.0/.1 local-tag episode is recorded as a
counter-precedent predating the ldflags-only rule.
- RED-now command: `grep -c "counter-precedent" .moai/docs/version-management.md`
- verbatim stdout: `0` · exit code: `1` · tree: `fa8ff89ba`
- Green path (M1): ≥1, exit 0.
- Mutant probe: a rule stating "no tags" while silently rewriting the tag history (dropping the
  counter-precedent) stays red on this AC → adopted.

**AC-RC-003 — BUILD_ID ordering warning exists.**
Given `fa8ff89ba`, When build-ordering guidance is searched in version-management.md, Then no
BUILD_ID identity warning exists (the token appears only in Makefile comments today); after M1
the warning cites the SPEC-BINARY-LAG-VISIBILITY-001 incident (v3.1.2 newer than v3.1.3-rc.5).
- RED-now command: `grep -c "BUILD_ID" .moai/docs/version-management.md`
- verbatim stdout: `0` · exit code: `1` · tree: `fa8ff89ba`
- Green path (M1): ≥1, exit 0.
- Mutant probe: a policy that lets builds be ordered by version string while mentioning BUILD_ID
  in passing is a substance violation routed to sync-audit reading (cluster residual risk).

**AC-RC-004 — develop refresh section exists.**
Given `fa8ff89ba`, When the lane protocol is searched for the develop-refresh procedure, Then no
such section exists (research §3: greps for refresh/recreate/재생성 returned nothing
develop-related on every lens); after M2 the develop 갱신 section exists.
- RED-now command: `grep -c "develop 갱신" .claude/rules/local/gitflow-lane-protocol.md`
- verbatim stdout: `0` · exit code: `1` · tree: `fa8ff89ba`
- Green path (M2): ≥1, exit 0.
- Mutant probe: a section that documents only the *what* (refresh local develop) without the
  criterion or route is defeated by AC-005/006; adopted as a cluster.

**AC-RC-005 — Merge-shape-agnostic criterion justified.**
Given `fa8ff89ba`, When the refresh criterion's justification is searched, Then no
squash-blindness citation exists in the lane protocol; after M2 the criterion keys to
`origin/develop` and cites SPEC-WORKTREE-SQUASH-MERGE-001 for why `git branch --merged` is not
a sufficient criterion (without asserting "--merged is empty" as a develop-lane fact — C4).
- RED-now command: `grep -c "SPEC-WORKTREE-SQUASH-MERGE-001" .claude/rules/local/gitflow-lane-protocol.md`
- verbatim stdout: `0` · exit code: `1` · tree: `fa8ff89ba`
- Green path (M2): ≥1, exit 0.
- Mutant probe: a criterion phrased as `git branch --merged origin/develop` alone (the exact
  blind spot) would flip AC-004 while leaving the requirement violated — this AC's grep forces
  only the citation's presence; the merge-shape-agnostic phrasing itself is not grep-forceable
  and is routed to sync-audit reading (same disclosure as AC-003/AC-006); adopted.

**AC-RC-006 — BranchGuard-safe route named.**
Given `fa8ff89ba`, When the route for develop regeneration is searched, Then the lane protocol
names no BranchGuard constraint; after M2 the section names BranchGuard and routes regeneration
through launcher worktree entry (`moai cc -w develop` / `EnterWorktree`), never primary-checkout
`git branch`, never cross-tree `git -C`.
- RED-now command: `grep -c "BranchGuard" .claude/rules/local/gitflow-lane-protocol.md`
- verbatim stdout: `0` · exit code: `1` · tree: `fa8ff89ba`
- Green path (M2): ≥1, exit 0.
- Mutant probe: recommending the operator-terminal sentinel route to subagents (C5 — both
  exemptions unreachable from tool-spawned subagents) is a substance violation; the token check
  is shallow here by nature, substance routed to sync-audit reading.

**AC-RC-007 — §4.1 pointer wiring exists (MUTANT-KILLER for the card's representative mutant).**
Given `fa8ff89ba`, When CLAUDE.local.md §4.1 (the declaration site of the standing develop) is
searched for the pointer pair, Then no pointers exist; after M3 the §4.1 additions point to the
Local RC Numbering section of version-management.md AND to the develop 갱신 section of the lane
protocol, as pointer lines only.
- RED-now command (i): `grep -c "Local RC Numbering" CLAUDE.local.md`
  · verbatim stdout: `0` · exit code: `1` · tree: `fa8ff89ba`
- RED-now command (ii): `grep -c "develop 갱신" CLAUDE.local.md`
  · verbatim stdout: `0` · exit code: `1`. First measured at `856ed147b` (doc targets
  byte-identical to `fa8ff89ba` at that point); **re-measured post-absorb at `a04afea53`**
  after the D1 resolution replaced §4.1 with the landed 08-29 chain — still `0`, exit 1.
  The re-measurement on `a04afea53` is the live basis for both anchors of this AC.
- Green path (M3): both commands → stdout ≥1, exit 0.
- Mutant probe: **the representative mutant** — amend §4.1 to declare develop standing while
  skipping the rc build + clean-reinstall documentation — leaves this AC red (the declaration
  site carries no pointer pair). Defeats it; adopted.

**AC-RC-008 — moai-update re-application note exists.**
Given `fa8ff89ba`, When the git-strategy.yaml reset hazard is searched, Then version-
management.md does not mention `rc_version_format`; after M1 the note exists (keys reset by
`moai update`, re-apply after every update — CLAUDE.local.md §2.3).
- RED-now command: `grep -c "rc_version_format" .moai/docs/version-management.md`
- verbatim stdout: `0` · exit code: `1` · tree: `fa8ff89ba`
- Green path (M1): ≥1, exit 0.
- Mutant probe: noting the key without the re-application instruction is a substance violation
  routed to sync-audit reading.

## §D.2 Regression-guard ACs

None adopted. The C3 doctrine-consistency check (git-workflow-doctrine.md residual prohibition
lines) is **green at RED time** on this tree (`fa8ff89ba`: probe returned no output) — a
criterion green at arrival is vacuous per verification-completeness §2 and is therefore NOT
recorded as an AC. It is executed as M4 verification evidence in progress.md §E.2 instead.

## §D.3 Edge cases

- CLAUDE.local.md byte growth: M3 delta must be reported even though a pointer-only edit is
  expected under the 1,000-byte single-edit threshold (rule-authoring duty (b)).
- Post-merge primary/worktree copy divergence (C1): RESOLVED — D1 option (b); M3 lands on the
  landed 08-29 chain (absorbed into this branch at `a04afea53`). The AC grep is anchor-token
  based and unaffected by the §4.1 text version.
- If a RED probe at run-phase start returns ≥1 (another actor landed overlapping content),
  re-baseline + blocker report to lead — never silent absorption (plan §C).

## §D.4 Severity & gates

All 8 ACs are release-blocking (binary PASS/FAIL). Quality gate: `moai spec lint` on the SPEC
directory with zero error-severity findings; byte-delta report for M3; 5-section evidence format
for the completion report.

## §D.5 Definition of Done

All 8 ACs green on the final tree with verbatim evidence; M4 verification sweep recorded in
progress.md §E.2; D1 resolution recorded (chain landed via `9a161687a` + `6b03e1757`, absorbed
at `a04afea53` — no open clarification markers remain); commit on `WT-rc-testbed` with
`card: t281` in the body; WT branch NOT pushed (integration is the lead's window).
