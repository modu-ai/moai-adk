# SPEC-ZONE-SHELL-CD-TRACKING-001 — decision index

Authored at plan phase (interview.decision_gate: on, `.moai/config/sections/interview.yaml:6`). Stateless on the status axis; the SPEC's lifecycle lives in spec.md alone. Rows route by the four labels DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER; an unverifiable citation never relabels a row.

### Q1: Should a post-`--` hyphen-leading cd directory operand (`cd -- -zone`) be tracked like any other operand — denying a protected deletion — instead of keeping the documented untracked reset (allow)?

Label: FOUNDER
Class: product-level
Authority anchor: none admissible. The only candidate, decision-board record `d-20261009T171254Z-f6ac`, is a card-scoped ruling (`scope=card:t1584`, `kind=ruling`). The register admits only standing board records and signed mission contracts, and a card ruling is evidence for its card, not authority (SPEC-FACTORY-DECISION-AUTO-001 design §2).
Why unresolved: Both the plan-audit D7 required fix (deny — the shape deletes `-zone/a.log`, a protected path; the drafted allow control was ruled a self-contradiction with SPEC-ZONE-SHELL-PARSING-001 REQ-ZSP-001) and the card t1584 dispatch (operator-approved expansion axis) direct deny, but both records live under `.moai/reports/` (untracked); no completed SPEC's HISTORY/Amendments row or `.moai/config/sections/*.yaml` setting decides the identical question, and the authority register cannot cite untracked material. This SPEC's REQ-ZSCD-001/REQ-ZSCD-005 encode the deny direction pending the operator's own decision.
Provenance note: via former leader record d-20261009T171254Z-f6ac (tmm974 handover, 2026-10-09T13:08Z); the current session operator has not restated these pins.
Awaiting: a standing decision-board record, or the operator's first-hand statement.
Operator verdict:
Relayed text (not a verdict): "`--` 뒤 하이픈 선행 cd 디렉터리 피연산자는 추적·deny (`cd -- -zone && rm a.log` deny·REQ-ZSP-001 정렬) — 운영자 핀 2026-10-09, 리더 경로 전달"

### Q2: Does an outside-root cd destination keep the documented untracked reset (allow), recorded outside the regression matrix?

Label: FOUNDER
Class: product-level
Authority anchor: none admissible (same record as Q1; a card-scoped ruling is evidence for its card, not authority, per SPEC-FACTORY-DECISION-AUTO-001 design §2).
Why unresolved: The card directs the retention of a landed allow behavior that is an HONEST accepted under-match (v0.2.0 rationale, per audit round 1 D2): a `..`-reaching relative spelling from an outside-root cwd reaches back into the zone while the guard answers allow — the tracked set was reset to the root, so the `..`-leading candidate is classified root-escaping while real bash resolves it inside the project (codex isolated repro at the audited tree: `cd <parent>/002/deep && rm ../../001/zone_dir/a.log` → allow, vs direct `rm zone_dir/a.log` → deny; code path `zoneNextCwd` :451 reset + `zoneRelativeToSet` :366 tracked-cwd-only join). The RETENTION is the open decision: a card may keep a landed under-match only on an honest record, and it may not mislabel it as sound; no committed register anchor decides the retention, so the operator must formalize keeping a known-reaching under-match outside the matrix (the alternative — hardening the class — grows scope beyond this card's two files). REQ-ZSCD-003 and its Out of Scope record carry this rationale and keep the reset pending the operator's decision.
Provenance note: via former leader record d-20261009T171254Z-f6ac (tmm974 handover, 2026-10-09T13:08Z); the current session operator has not restated these pins.
Awaiting: a standing decision-board record, or the operator's first-hand statement.
Operator verdict:
Relayed text (not a verdict): "루트 밖 cd 목적지는 문서화된 미추적 리셋 유지 — 역방달 도달 가능 명시·'의미상 안전' 라벨 제거·경화는 향후 결정 — 운영자 핀 2026-10-09, 리더 경로 전달"

### Q3: Where do the cd-class regression cells live?

Label: FOUNDER
Class: implementation-level
Default: extend the landed `internal/hook/protected_zone_shell_matrix_test.go` cell catalogue with a cd group (rule: the option that preserves current behavior — no new test file)
Alternate: a dedicated cd matrix test file
Why unresolved: A file-placement judgment surfaced during assembly; no policy or prior SPEC row decides it, and it needs no operator data — the published Default rule ranks it.
Operator verdict: DEFAULT-APPLIED 2026-10-09T03:15:40Z manager-spec (t1584-spec, plan-phase)
