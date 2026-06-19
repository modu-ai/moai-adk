# Progress — SPEC-V3R6-RULES-SSOT-DEDUP-001

Lifecycle evidence ledger. §E.1 populated at plan-phase by manager-spec; §E.2/§E.3 by
manager-develop (run-phase); §E.4/§E.5 by manager-docs (sync/Mx-phase).

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored: spec.md + plan.md + acceptance.md + design.md + research.md + progress.md (6 files).
- Tier L; status: draft; era: V3R6.
- 12 GEARS requirements (REQ-SSD-001..012); 9 de-dup target groups mapped to 8 milestones + 1 zone-registry milestone (M_Z).
- Decisive research finding (research.md §3.3): zone-registry `clause:` is verbatim-source-substring-coupled to `moai constitution validate` → label-only reduction is UNSAFE; REQ-SSD-009 scoped SHOULD/partial (verbatim-shorter-excerpt + `paths:` narrowing) with a blocker note for follow-up SPEC `SPEC-V3R6-CONST-VALIDATE-LABEL-001`.
- File-deletion target (agent-teams-pattern.md) has a 3-clause AC (AC-SSD-008) + the CATALOG-SCRUB-001 dependency recorded (spec.md §F).
- Per-file mirror class table fixed (design.md §4): 4 classes (byte-parity allowlist=hooks-system ONLY / byte-identical-by-discipline=13 files / §25-sanitized=3 files / §25-divergent=zone-registry / no-mirror=lifecycle-sync-gate) — ACs use the correct verification per class.
- SPEC ID pre-write self-check: `decomposition: SPEC ✓ | V3R6 ✓ | RULES ✓ | SSOT ✓ | DEDUP ✓ | 001 ✓ → PASS`.

### plan-auditor iter-1 → iter-2 (PASS-WITH-DEBT 0.84 → fixes applied; hardest calls verified correct)
- iter-1 verdict: PASS-WITH-DEBT 0.84 (Tier L threshold 0.85, short 0.01, NOT skip-eligible). zone-registry CLI gate, file-deletion model, §25 trio usage, scope boundary all independently VERIFIED CORRECT.
- All 8 defects (D1-D8) independently re-confirmed against LIVE repo state before fixing (verification-claim-integrity §1.1 — no defect accepted on report alone):
  - D1/D2 (MAJOR): mirror-verification model corrected. ONLY hooks-system.md is in `workflowOptMirroredPaths` byte-parity allowlist; "TestRuleTemplateMirrorDrift PASS" is vacuous for the 13 non-enrolled byte-identical-by-discipline files → replaced with explicit per-file `diff -q`. zone-registry is §25-divergent (13-line CONST-V3R6-001 delta, verified `diff` = 13) → leak-test + scoped `diff`, NOT `diff -q`. design.md §4 table corrected.
  - D3 (MAJOR): AC-SSD-004 vacuous 7-keyword line-count (live=17) → per-keyword loop asserting each of 7 distinct commands ≥1.
  - D4 (MAJOR): verification-batch-pattern.md has NO 7-cmd block (already defers); the lone `coverprofile=cover.out` is the retained L29 race-note prose. T4 re-targeted to verified-no-op + re-sync sentinel; the false-failing `# 0` clause REMOVED.
  - D5/D6 (minor): every "(absent)" clause now uses a distinctive-line before(N)/after(<N) delta (pointer-presence was already-true pre-edit — agent-common-protocol=2, moai-constitution=1). alwaysLoad: grep the specific introduction line, not a bare version sweep.
  - D7 (minor): timeout reconciliation now edits BOTH hooks-system surfaces — the L323 timeout TABLE (PostToolUse currently grouped under 5s) AND the L244 JSON example (already 10s+async).
  - D8 (minor): agent-teams-pattern.md dangling-grep WIDENED to `internal/` broadly — names the 3 missed comment referrers (workflow_role_profiles_test.go:21,45; defaults_test.go:527; + rule_template_mirror_test.go:66; workflow.yaml:28).
- Artifacts edited iter-2: acceptance.md (mirror-class + distinctive-line conventions + per-AC fixes), design.md §4 + §6.1, plan.md (M2/M4/M5/M6/M8/M_Z mirror classes + AP-8..11), this §E.1.
- spec.md REQ wording unchanged (defects were AC-testability, not requirement-level); re-ran `moai spec lint` clean.

## § Mode Selection (Phase 0.95)

**Input parameters**:
- tier: L
- scope (file count): ~20+ files (9 de-dup targets × deployed + template mirror, plus M6 file deletion + 4 inbound referrers)
- domain count: 2 (`.claude/rules/moai/` deployed + `internal/template/templates/.claude/rules/moai/` mirror)
- file language mix: 100% markdown (spec.md §G: No Go source edit)
- concurrency benefit: LOW (inter-file dependency — mirror edits land same-commit per REQ-SSD-010; shared files across milestones: settings-management touched by M1+M2+M3, zone-registry by M3+M_Z; referrer rewrite cascade in M6)
- Agent Teams prereqs: not needed (Mode 5 selection — harness/team.enabled/env not required)

**Mode evaluation**:
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | Tier L, 9 milestones, file deletion — not trivial |
| 2 background | no | Write-heavy (mirror edits + deletion) — not read-only |
| 3 agent-team | no | team prereqs not met; markdown-heavy with inter-file deps |
| 4 parallel | no | inter-file dependency (mirror same-commit, shared files across milestones) — not independent |
| 5 sub-agent | **YES** | sequential per-milestone, Tier L Section A-E template, worktree isolation per memory race-mitigation pattern |
| 6 workflow | no | NOT a single uniform mechanical transform — each milestone applies a distinct de-dup strategy (pointer reduction / contradiction reconciliation / file deletion + referrer rewrite / CLI-gated zone-registry partial); multi-rule semantic edit |

**Decision**: sub-agent (Mode 5) with `isolation: "worktree"`

**Justification**: SSOT-DEDUP is markdown-heavy (no Go edit per §G) but multi-rule semantic — each of 9 milestones applies a distinct de-dup strategy. Inter-file dependency (mirror edits same-commit per REQ-SSD-010; settings-management + zone-registry shared across M2+M3+M_Z) rules out Mode 4 parallel. Worktree isolation per `feedback_shared_main_orphan_race` — Tier L cross-file work on shared main, multi-session race mitigated via manager-develop worktree commit → orchestrator cherry-pick to shared main (Sprint 16 CATALOG-SCRUB·VERSION-FORMAT proven pattern). Anthropic coding-task parallelism caveat: sequential sub-agent is the safe default for non-trivially-parallel work.

## §E.2 Run-phase Evidence

Run-phase executed via manager-develop in an isolation worktree (Mode 5 sub-agent, per § Mode Selection); orchestrator cherry-picked the 7 run commits to shared main (M3 + M8 verified no-ops — design premises disproven by live grep, verification-claim-integrity §1.1 surface 3):

- `d581c7d0b` M1 — AskUserQuestion 4-way de-dup (agent-common-protocol + moai-constitution pointers; CLAUDE.md §8 out-of-scope forward-item)
- `e93e99203` M2 — hooks config/timeout de-dup + contradiction reconcile (D7 BOTH surfaces L323 table + L244 JSON; D6 single alwaysLoad v2.1.119)
- `a08e9aefc` M4 — verification-batch re-sync sentinel (D4 verified no-op: no 7-cmd block exists to remove)
- `c10411102` M5 — Status Transition Ownership Matrix de-dup (lifecycle-sync-gate pointer; NO mirror RK-6)
- `de15de6d9` M6 — team files 4→2 + DELETE `agent-teams-pattern.md` (both trees) + 4 inbound referrer rewrites (D8: internal/ BROADLY) + max_teammates 10 vs 3-5 reconcile
- `6b073a4d8` M7 — skill 3-file SSOT boundary (writing-craft + ab-testing → skill-authoring pointers)
- `b69761a8a` M_Z — zone-registry `paths:` narrowing (REQ-SSD-009 SHOULD partial; B4 CLI gate respected — full clause reduction deferred to SPEC-V3R6-CONST-VALIDATE-LABEL-001)
- M3 (context-window 5-way) — verified no-op: copies already reduced to pointers in prior SPECs
- M8 (agent 3-file) — verified no-op: boundary already clean

29 files changed (deployed + template mirror pairs); 172 insertions / 564 deletions (net de-dup). PRESERVE clean — sibling SPEC artifacts (DOCS-RC2, HARNESS-MOAI-NAMESPACE) untouched.

## §E.3 Run-phase Audit-Ready Signal

- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- Go-test trio green: `TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestTemplateNeutralityAudit` ok (`internal/template`)
- `moai spec lint`: 26 errors / 87 warnings — the +1 NEW was StatusGitConsistency on this SPEC (status draft vs git-implied, resolved by this sync-phase `completed` transition); de-dup NEW defects = 0; remainder pre-existing baseline
- `moai constitution validate`: drift_count 77 → 77 (M_Z `paths:` narrowing introduced 0 NEW drift; AP-3 respected — no `clause:` blanked to a label)
- 15 ACs (acceptance.md SSOT): AC-SSD-001..008, 010..012 PASS; AC-SSD-009 PASS-WITH-BLOCKER-NOTE (zone-registry full reduction deferred per B4)
- 4 mirror classes verified per-file (design §4): hooks-system byte-parity (allowlist) / 13 byte-identical-by-discipline (explicit `diff -q`) / §25-sanitized (leak-test) / zone-registry §25-divergent (scoped `diff`) / lifecycle-sync-gate no-mirror
- manager-develop self-claim "comment-only _test.go edits" independently verified (diff showed comment/list-comment only, no Go logic change; new-diagnostics `forvar`/`any` at lines 69/83/116/176 are pre-existing, not M6-touched lines 21/45/66)

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: <backfill pending — close commit SHA>

Sync-phase orchestrator-direct (GLM manager-docs spawn context-limit fallback per `feedback_glm_orchestrator_direct_sync_mx`). frontmatter status in-progress → completed rides this sync commit (3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-008 — `completed` transition merged into the sync commit, no separate Mx chore). §E.5 Mx-phase retired (folded into §E.4 per LIFECYCLE-REDESIGN). CHANGELOG entry added (Sprint 16 RULES cohort).

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_
