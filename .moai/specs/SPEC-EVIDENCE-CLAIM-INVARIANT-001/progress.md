# SPEC-EVIDENCE-CLAIM-INVARIANT-001 — Progress

> Tier S (minimal). 4-phase lifecycle: plan → run → sync → Mx.

## §F.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-06-15
plan_status: audit-ready
plan_artifacts:
  - spec.md          # 12-field frontmatter + §A intent/provenance + §3 inline AC + §X.1 h3 Out of Scope
  - plan.md          # M1-M3 milestones, run-phase constraints (template neutrality), PRESERVE notes
  - progress.md      # this file — §F.1 signal
tier: S
artifact_count: 2    # spec.md + plan.md (acceptance.md inline in spec.md §3 per Tier S)
ac_count: 7          # AC-ECI-001..007, all MUST, all grep/file-existence falsifiable
spec_id_self_check: "decomposition: SPEC | EVIDENCE | CLAIM | INVARIANT | 001 → PASS"
frontmatter_schema: "12-field canonical PASS (created/updated/tags canonical names; +optional tier:S)"
out_of_scope_h3: present   # §X.1 '### X.1 Out of Scope —' + list items → MissingExclusions PASS
provenance: "IMP-06 (fable-ish 13-agent roadmap final adopted item)"
predecessors:
  - SPEC-HOOK-DISCIPLINE-WIRING-001   # IMP-01 (completed)
  - SPEC-STOP-EVIDENCE-GATE-001       # IMP-02/03 (completed); deferred IMP-06 OUT OF SCOPE at spec.md §B.2 line 185
defect_class_motivation: "L_manager_docs_false_backfill_report (claiming an unobserved verification)"
run_phase_deliverable: "1 doctrine rule file + template mirror (.claude/rules/moai/core/verification-claim-integrity.md)"
biggest_run_risk: "template mirror neutrality (internal_content_leak_test.go CI guard)"
```

## §F.1.1 Plan Audit Gate (plan-auditor iter-1)

```yaml
verdict: PASS-WITH-DEBT
score: 0.84              # Tier S threshold 0.75 (+0.09)
dimensions:
  clarity: 0.92
  completeness: 0.95
  testability: 0.62      # the debt — falsifiability
  traceability: 0.96
must_pass: "MP-1 AC consistency PASS / MP-2 N/A(doctrine matrix) / MP-3 frontmatter 12+tier:S PASS / MP-4 neutrality N/A->auto"
cross_ref_integrity: "all 5 cited files + line numbers EXACT (Skeptical@113, Verify-Don't-Assume@262, SectionE@167, verification-batch-pattern present, Verification Matrix@368 / Completion Report@574)"
defects:
  D1: "SHOULD-FIX -> REMEDIATED (orchestrator-direct). AC-ECI-006 vacuous grep: ERE backslash-pipe = literal pipe -> always 0-match (would pass on a real leak). Fix: verification commands moved table-cell -> fenced code block (§3.2) with real | alternation + escaped dot. Empirically proven: old form 0-match on 'SPEC-EVIDENCE' leak; new form matches; neutral string 0-match."
  D2: "MINOR -> REMEDIATED. AC-ECI-003 '또는 동치' synonym hatch -> §3.2 C3 concrete tokens (output-styles/moai/moai.md + manager-develop §E / E1-E7)."
  D3: "MINOR -> REMEDIATED. AC-ECI-007 human-tally grep -c -> §3.2 C7 single 'grep -cE ... -ge 4'."
  D4: "NO-ACTION. StatusGitConsistency WARNING (status draft vs git-implied implemented) = expected Hybrid-Trunk plan-on-main heuristic over-inference, NOT a SPEC defect; frontmatter draft is correct. lint.skip PROHIBITED (would suppress a useful signal on other SPECs). Independently confirmed by orchestrator Trust-but-verify V4 AND plan-auditor re-run."
remediation: "orchestrator-direct §3.2 code-block fix, committed with this plan-audit remediation commit (chore scope, status stays draft)"
report: ".moai/reports/plan-audit/SPEC-EVIDENCE-CLAIM-INVARIANT-001-2026-06-15.md"
falsifiability_post_fix: "testability debt addressed — all 7 AC now executable verbatim (C1-C7 in §3.2); run-phase §E runs the corrected idiom (running the vacuous form + reporting false 0-match PASS would itself violate this SPEC's invariant)"
```

## §F.2 Run-phase Evidence

> Structured per the 5-section evidence-bearing format that this SPEC codifies (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) — the run-phase self-report obeys the invariant it implements.

### Claim (주장)

run-phase delivers M1 (local doctrine rule) + M2 (template mirror); all 7 inline AC (AC-ECI-001..007) PASS via the §3.2 C1-C7 executable verification commands; mirror neutrality = 0 internal leaks; cross-ref count = 6 (≥4); CI guard `TestTemplateNeutralityAudit` PASS. status draft→in-progress (M1 commit).

### Evidence (증거 — actual command + verbatim output)

AC binary matrix (each row's verification command run verbatim from spec.md §3.2; outputs observed in this run):

| AC | Status | Verification Command (§3.2) | Actual Output (verbatim) |
|----|--------|-----------------------------|--------------------------|
| AC-ECI-001 | PASS | `test -f .claude/rules/moai/core/verification-claim-integrity.md && echo C1-PASS` | `C1-PASS` |
| AC-ECI-002 | PASS | `test -f internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md && echo C2-PASS` | `C2-PASS` |
| AC-ECI-003 | PASS | C3 — `grep -q 'unobserved' && grep -q 'Evidence absent' && grep -q 'output-styles/moai/moai.md' && grep -qE 'manager-develop.*§E\|E1[^0-9].*E7'` | `C3-PASS` |
| AC-ECI-004 | PASS | C4 — `for s in '주장' '증거' 'baseline 귀속' '미검증' '잔여 위험'; do grep -q "$s" \|\| echo FAIL; done` | (no FAIL line — all 5 sections present) `C4-DONE` |
| AC-ECI-005 | PASS | C5 — `grep -qE 'baseline' && grep -qE '귀속\|attribution' && grep -qE '측정\|measured'` | `C5-PASS` |
| AC-ECI-006 | PASS | C6 — `grep -nE 'SPEC-[A-Z]\|feedback_\|/Users/\|CLAUDE\.local\|[0-9a-f]{7,40} commit' "$MIRROR" \|\| echo C6-PASS-0-leaks` | `C6-PASS-0-leaks` (real-alternation grep on mirror, 0 matches) |
| AC-ECI-007 | PASS | C7 — `test "$(grep -cE 'agent-common-protocol\.md\|moai-constitution\.md\|manager-develop-prompt-template\.md\|verification-batch-pattern\.md\|output-styles/moai/moai\.md' "$RULE")" -ge 4` | `C7 count: 6` → `C7-PASS` |

Extended mirror neutrality sweep (full plan §D.1 forbidden catalogue, all `grep -c` on the mirror, all expect 0):
```
internal SPEC IDs (SPEC-[A-Z]):        0
REQ/AC tokens (REQ-|AC-[A-Z]):         0
IMP-NN roadmap refs (IMP-[0-9]):       0
internal dates (2026-NN-NN):           0
commit-SHA-like hex (\b[0-9a-f]{7,40}\b): 0
feedback_/L_ memory refs:              0
CLAUDE.local refs:                     0
/Users/ paths:                         0
fable-ish provenance:                  0
```

CI guard:
```
$ go test ./internal/template/ -run 'TestTemplateNeutralityAudit|InternalContentLeak|TestInternalContent' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	0.831s
```

### Baseline-attribution (baseline 귀속)

- All AC verification commands measured against the worktree tree at HEAD `d9cb89bd4` (worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a87f6a659bb36dc7e`; `git rev-parse HEAD` observed before edits). The 2 new files are the only run-phase additions.
- Pre-write baseline: both target files confirmed absent (`LOCAL ABSENT` / `MIRROR ABSENT`) and both target dirs present (`core/ EXISTS` / `template core/ EXISTS`) before Write.
- Cross-ref target existence verified before authoring §4 (all 5 cited files `EXISTS`); cross-ref anchors verified present (Skeptical Evaluation Stance @113, Verify-Don't-Assume @262, Section E @167, Verification Matrix @368 / Completion Report @574).
- Pre-spawn sync: `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` → `0 0` (synced).
- CI guard figure `0.831s` is this-run output of the `go test` invocation above (not carried over).

### Gaps (미검증 — what was NOT observed)

- Full `go test ./...` suite NOT run — run-phase added 2 markdown files only, no Go code; only the targeted template-neutrality CI guard was exercised (the relevant guard for the mirror file).
- `golangci-lint` / `go vet` NOT run — no Go source changed (plan §C: Go build unnecessary for markdown-only deliverable).
- `make build` NOT run — deliberately skipped per orchestrator constraint (`//go:embed all:templates` is compile-time; new markdown mirror needs no `embedded.go` regeneration for the AC, and skipping avoids entanglement with the parallel template workstream). embedded.go regeneration was therefore NOT verified — out of scope for AC-ECI-001..007.
- Cross-platform build (`GOOS=windows`) NOT run — no Go code.
- Coverage NOT measured — no Go code touched.
- M3 cross-ref pointer into existing files (output-styles/moai.md / manager-develop-prompt-template.md) NOT added — AC-ECI-007 is satisfied solely by the new rule→existing direction (per plan §F M3, the reverse pointer is SHOULD, not MUST; and the parallel-workstream constraint forbids touching shared files).

### Residual-risk (잔여 위험)

- AC-ECI-006 (mirror neutrality) is enforced both by the §3.2 C6 grep AND the `TestTemplateNeutralityAudit` CI guard; if the guard's pattern set diverges from C6, a future leak class could pass one but fail the other. Current run: both PASS, 0 leaks.
- The doctrine is policy-layer only (no runtime detector) — by design (spec.md §X.1). Enforcement of the invariant on live actor reports depends on actors reading the rule; no mechanical gate added in this SPEC.
- Commit + push performed against shared APFS tree with a known parallel workstream (uncommitted M files in shared dirs); staged ONLY the 4 explicit-path files. Push divergence re-checked at push time (see commit/push block below).

### Commit + Push (Hybrid Trunk, draft→in-progress transition)

- run_commit_sha: `6039e2fcd91ea4dc82854cd814453edc434c054c` (M1-M3, 4 files: 2 new doctrine + spec.md status + this progress.md)
- pre-push divergence (before push): `0 1` (origin/main NOT ahead; my branch 1 clean linear commit ahead of `d9cb89bd4` == origin/main; no race)
- push result: `d9cb89bd4..6039e2fcd  HEAD -> main` (exit 0; pre-push hook warn-only, NOT --no-verify; remote bypass log = Hybrid Trunk main-direct policy, CI 4 checks triggered)
- post-push divergence: `0 0` (origin/main HEAD == `6039e2fcd`)
- Authored-By-Agent: manager-develop (commit trailer, draft→in-progress canonical owner per Status Transition Ownership Matrix)

## §F.3 Sync-phase Audit-Ready Signal

(sync-phase에서 manager-docs가 채움)

## §F.4 Mx-phase Audit-Ready Signal

(Mx-phase에서 채움 — mx_commit_sha 등)
