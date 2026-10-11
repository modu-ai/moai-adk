# SPEC-ZONE-SHELL-CD-TRACKING-001 — decision index

Authored at plan phase (interview.decision_gate: on, `.moai/config/sections/interview.yaml:6`). Stateless on the status axis; the SPEC's lifecycle lives in spec.md alone. Rows route by the four labels DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER; an unverifiable citation never relabels a row.

### Q1: Should a post-`--` hyphen-leading cd directory operand (`cd -- -zone`) be tracked like any other operand — denying a protected deletion — instead of keeping the documented untracked reset (allow)?

Label: DECIDED
Class: product-level
Authority anchor: board:d-20261009T175650Z-4ae6#a01b8d3d39050a3558c403e881d97a60bb47a0a4a8f6604c6484d0829623b8e5
Why unresolved: resolved by the operator's first-hand standing record; the earlier relayed pin is superseded as authority.
Provenance note: via former leader record d-20261009T171254Z-f6ac (tmm974 handover, 2026-10-09T13:08Z); the operator restated both answers first-hand on 2026-10-09 (standing record d-20261009T175650Z-4ae6); the relayed pin is superseded as authority.
Operator verdict: Q1 - a post-'--' hyphen-leading cd directory operand ('cd -- -zone') is TRACKED like any other operand and a protected deletion under it is DENIED (operator's selected answer: "추적하고 삭제 거부").

Pinned record (verbatim from `moai decision read --scope standing`; digest = SHA-256 of the two lines joined by one LF, with no trailing LF):

```text
decision record: decided_by=operator (first-hand answers via AskUserQuestion to leader session 069be28e, 2026-10-09T17:5xZ; recorded verbatim by leader) evidence_refs=AskUserQuestion round in leader transcript 069be28e; SPEC-ZONE-SHELL-CD-TRACKING-001 decision-index.md Q1/Q2 at c92d8d81b; lane-3 blocker 17df1928 ladder_path=operator first-hand statement (decision-index 'Awaiting' clause) id=d-20261009T175650Z-4ae6 scope=standing kind=standing-rule
  body: Operator verdicts for SPEC-ZONE-SHELL-CD-TRACKING-001 (card t1584), given first-hand by the current-session operator. Q1 - a post-'--' hyphen-leading cd directory operand ('cd -- -zone') is TRACKED like any other operand and a protected deletion under it is DENIED (operator's selected answer: "추적하고 삭제 거부"). Q2 - an outside-root cd destination KEEPS the documented untracked reset (allow) and the reaching-back under-match is recorded as a known limitation outside the regression matrix; hardening is not in this card (operator's selected answer: "유지하고 한계로 기록"). AC classification - AC-ZSCD-005 and AC-ZSCD-006 are approved as regression-guard criteria, not release-blocking (operator's selected answer: "회귀 가드로 승인"). These supersede the relayed, non-authoritative card ruling d-20261009T171254Z-f6ac as the authority anchor for Q1 and Q2.
```

### Q2: Does an outside-root cd destination keep the documented untracked reset (allow), recorded outside the regression matrix?

Label: DECIDED
Class: product-level
Authority anchor: board:d-20261009T175650Z-4ae6#a01b8d3d39050a3558c403e881d97a60bb47a0a4a8f6604c6484d0829623b8e5
Why unresolved: resolved by the operator's first-hand standing record; the earlier relayed pin is superseded as authority.
Provenance note: via former leader record d-20261009T171254Z-f6ac (tmm974 handover, 2026-10-09T13:08Z); the operator restated both answers first-hand on 2026-10-09 (standing record d-20261009T175650Z-4ae6); the relayed pin is superseded as authority.
Operator verdict: Q2 - an outside-root cd destination KEEPS the documented untracked reset (allow) and the reaching-back under-match is recorded as a known limitation outside the regression matrix; hardening is not in this card (operator's selected answer: "유지하고 한계로 기록").

Pinned record (verbatim from `moai decision read --scope standing`; digest = SHA-256 of the two lines joined by one LF, with no trailing LF):

```text
decision record: decided_by=operator (first-hand answers via AskUserQuestion to leader session 069be28e, 2026-10-09T17:5xZ; recorded verbatim by leader) evidence_refs=AskUserQuestion round in leader transcript 069be28e; SPEC-ZONE-SHELL-CD-TRACKING-001 decision-index.md Q1/Q2 at c92d8d81b; lane-3 blocker 17df1928 ladder_path=operator first-hand statement (decision-index 'Awaiting' clause) id=d-20261009T175650Z-4ae6 scope=standing kind=standing-rule
  body: Operator verdicts for SPEC-ZONE-SHELL-CD-TRACKING-001 (card t1584), given first-hand by the current-session operator. Q1 - a post-'--' hyphen-leading cd directory operand ('cd -- -zone') is TRACKED like any other operand and a protected deletion under it is DENIED (operator's selected answer: "추적하고 삭제 거부"). Q2 - an outside-root cd destination KEEPS the documented untracked reset (allow) and the reaching-back under-match is recorded as a known limitation outside the regression matrix; hardening is not in this card (operator's selected answer: "유지하고 한계로 기록"). AC classification - AC-ZSCD-005 and AC-ZSCD-006 are approved as regression-guard criteria, not release-blocking (operator's selected answer: "회귀 가드로 승인"). These supersede the relayed, non-authoritative card ruling d-20261009T171254Z-f6ac as the authority anchor for Q1 and Q2.
```

### Q3: Where do the cd-class regression cells live?

Label: FOUNDER
Class: implementation-level
Default: extend the landed `internal/hook/protected_zone_shell_matrix_test.go` cell catalogue with a cd group (rule: the option that preserves current behavior — no new test file)
Alternate: a dedicated cd matrix test file
Why unresolved: A file-placement judgment surfaced during assembly; no policy or prior SPEC row decides it, and it needs no operator data — the published Default rule ranks it.
Label note: stays FOUNDER. POLICY-COVERED needs a register anchor (a committed operator setting or constitution clause that covers the question as written). The published Default rule is stated in the manager-spec agent definition (`.claude/agents/moai/manager-spec.md` § FOUNDER row classes), which the register does not admit. The row is implementation-level with a Default, so its DEFAULT-APPLIED verdict below does not hold the Kickoff.
Operator verdict: DEFAULT-APPLIED 2026-10-09T03:15:40Z manager-spec (t1584-spec, plan-phase)
