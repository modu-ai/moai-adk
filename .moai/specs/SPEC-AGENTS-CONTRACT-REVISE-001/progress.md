# SPEC-AGENTS-CONTRACT-REVISE-001 — Progress

Created at plan phase (factory card t1900, run tn1p4e, lane-44) — 2026-10-19.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-19
artifacts: spec.md + plan.md + acceptance.md + progress.md + decision-index.md (Tier M set + progress + decision gate artifact)
tier: M
notes: Independent plan-audit verdict lands via the plan-auditor gate; this signal states artifact completeness for audit, not a verdict.

## §E.2 Run-phase Evidence

Owner: manager-docs-led run delegation (factory card t1900, run tn564n, lane-44). Every item: (a) command, (b) verbatim output, (c) tree SHA. Content measurements captured at HEAD `072e858752` (this tree); final sizes re-stamped in the §E.3 commit note where the measurement post-dates it.

**Status-transition exception (documented).** The first run-phase commit `b4067ef526` carries `spec.md` frontmatter `status: draft → in-progress` with trailer `Authored-By-Agent: manager-docs`. The Status Transition Ownership Matrix names `manager-develop` for `draft → in-progress`; this is the documented docs-card exception — plan.md names this card manager-docs-led with no manager-develop leg anticipated (plan.md header). The `OwnershipTransitionInvalid` lint warning is accepted and explained here.

**PREFLIGHT-HEAD handling.** plan.md:29 expected HEAD `8249bff8c3`; actual run-entry HEAD was `d01b8185f0` (plan artifacts committed after the plan pre-flight). Branch name verified `WT-agents-agents-md`; `wc -m AGENTS.local.md` at entry returned exactly `50579` — the pre-card pin matched, so no re-pin was needed and the PIN was never substituted into the projection (per plan.md:29's instruction).

- **E1 AC matrix** — PASS/FAIL per AC-ACR-001…016, each row with its deciding command and observed output (captured this run):

| AC | Status | Deciding command (abbreviated evidence; full battery in the run log below) | Observed output |
|---|---|---|---|
| AC-ACR-001 | PASS | `grep -c "receiver" AGENTS.local.md`; section listing | `1`; `### §4.0 Current card policy…`(L156), `§4.0-R`(L182), `§4.0-H…not the current card procedure`(L201), `§4.1 … historical`(L219) |
| AC-ACR-002 | PASS | `grep -c "handle simple, bounded operations directly\|Do the bounded work inline" AGENTS.md AGENTS.local.md internal/template/templates/AGENTS.md.tmpl` | `0` ×3; `Permitted direct execution` → 0; `never directly implement\|shall not implement, repair, build` → AGENTS.md 2, AGENTS.local.md 2; emergency/trivial sweep → 3 hits, each inside a "no exception" sentence |
| AC-ACR-003 | PASS | `grep -c "CodeRabbit" AGENTS.md internal/template/templates/AGENTS.md.tmpl` | `0` ×2; producer-neutral current-head clause + cross-reference present in both |
| AC-ACR-004 | PASS | `grep -c "codex-review-state" AGENTS.local.md .claude/rules/local/repo-local-pr-policy.md .claude/rules/local/gitflow-lane-protocol.md` | `1` ×3; `only after the draft` → 0 hits; `1841` → 2 hits with `inputs/REPORT.md`+`inputs/native-codex-rest-comment.json` cited; `t1868` → 1 hit with in-progress marker; `gate.*implemented` → 1 hit which is the negation sentence itself ("no sentence here presents the gate as already-implemented tooling") — no hit presents the gate as shipped |
| AC-ACR-005 | PASS | full conflict grep (7 patterns × 9 files incl. mirrors) | swept hit count 5 (repo-local-pr-policy.md 2, gitflow-lane-protocol.md 3), each inside a transition-scoped section whose immediately-preceding marker states the scope; AGENTS.local.md and all mirrors 0 |
| AC-ACR-006 | PASS | `grep -rn "release/\*" … \| grep -v hotfix` | 5 hits, none a complete-rule merge-flag claim; the one merge-method row (git-workflow-doctrine.md L114) now carries the full rule `release/* OR hotfix/* → --merge, else --squash (auto-merge.yml:341-345)`; `hotfix/*` present in AGENTS.local.md §4.0-R |
| AC-ACR-007 | PASS | `grep -rn "gh pr create --draft" AGENTS.local.md .claude/rules/local/repo-local-pr-policy.md` | manual draft path named operative in both, each with the `card t1882, not landed` marker; `shipped default\|already implements` → 0 hits |
| AC-ACR-008 | PASS | AC-008 clause grep on AGENTS.local.md §4.0 item 4 | workflow_run trigger / no draft filter / no config gate / `--match-head-commit` all named; `BOTH … test verification confirmed AND explicit approval — both together`; `connected` 1, `hardcod` 5, `delete-branch` 2 (recorded limitation), `local operation choice` 2 |
| AC-ACR-009 | PASS | decision-index close-time evidence row + reconciliation table + glob check | PIN-based projection: 252,650 − 8,844 − t1887 ≤ 140,000 (required t1887 ≥ 103,806; strict bound ≥ 105,755 incl. AGENTS.md +1,949); 4 chunks moved, 6-row reconciliation table = 6 pointer greps; [HARD] 47 → 46+19=65; companion glob check 13 PASS / 0 FAIL (`internal/contract.MatchGlob` semantics copy) |
| AC-ACR-010 | PASS | pointer greps on AGENTS.local.md | `templates/CLAUDE.md` 0; `AGENTS.md.tmpl` 1; `CLAUDE.md Section 4` 0; `delegation.yaml` 1; `embed_manifest_gen.go` 1 + explicit emit targets in §1 and §2 |
| AC-ACR-011 | PASS | `grep -c "Historical\|히스토리컬\|역사 기록"` + caution grep | git-workflow-doctrine.md 7 markers / git-local-workflow-doctrine.md 4; protection caution ("현재 원격 설정의 증거가 아니다") 1 each; region markers neutralize the §23.3-23.6 and §18.7 recipes |
| AC-ACR-012 | PASS | `grep -c "보존\|preserv"` + incident-framing grep | update-local-file-survival.md 5 / version-management.md 1; incident framing (`사고 기록`) 2; git-strategy note rewritten to preserve-and-verify |
| AC-ACR-013 | PASS | `grep -n "git restore\|git reset --hard\|git stash" AGENTS.local.md .claude/rules/local/*.md` | 2 hits (AGENTS.local.md L42, L176) — both prohibitions ("Do not run…", item 8 preservation ordering); zero recovery instructions |
| AC-ACR-014 | PASS | `.moai/reports/t1900/run-diff-list.md` + `make build` + C9 | diff list delivered (15 rows, root+mirror pairs); `make build` exit 0 ×2, no tracked-file drift; `go test ./internal/template/ -run 'TestCardID\|TestTemplateNoInternalContentLeak' -count=1` → `ok` (2.391s); mirror parity: core+mechanics byte-identical, dispatch.md differs only by the pre-existing one-sentence local addition |
| AC-ACR-015 | PASS | `grep -rin "kanban" …` | exactly 1 hit — the prohibition sentence in §4.0 item 9 ("does not authorize implementing or switching to Kanban mode…never treated as implementation approval"); 0 hits authorize Kanban work |
| AC-ACR-016 | PASS | shell-form lane grep + occurrence-level variant | `grep -rn "todo --auto-lane" … \| grep -v "/moai todo --auto-lane"` → 0 hits; occurrence-level `(^|[^/])moai todo --auto-lane` → 1 hit = the AGENTS.local.md:180 invariant statement itself (backtick-quoted subject of the prohibition, not a shell usage); shell `moai todo --auto` and slash `--auto-lane`/`--auto-leader` distinguished in §4.0 item 10 with the t1695 delegation note |

- **E2 Template cycle** — `make build` after the M2 mirror edit → exit 0 (tail: `catalog.yaml updated successfully (24330 bytes)` + the go build line, commit `b4067ef526` embedded); after the M3 mirror edits → exit 0 again; `git status --short` after each build showed only the intended edits (no tracked-file drift). C9 guard `-count=1` → `ok github.com/modu-ai/moai-adk/internal/template 2.391s` (fresh run, M3 tree).
- **E3 Size model** — decision-index "Close-time evidence row": reduction 8,844 (PIN 50,579 − post 41,735); projection from the PIN, t1887 share ≥ 103,806 (strict ≥ 105,755); fresh values recorded in the evidence row only, never overwriting the baseline row. Reconciliation table 6 rows = 6 pointer greps.
- **E4 Conflict grep** — AC-ACR-005 output: 5 surviving hits, all inside marked transition-scoped sections (repo-local-pr-policy.md historical block ×2; gitflow-lane-protocol.md §2 marker region ×2, §4 historical heading region ×1); AGENTS.local.md 0; mirrors 0. Swept hit count reported (5), not a zero-sweep claim.
- **E5 Diff list** — `.moai/reports/t1900/run-diff-list.md` (15 rows, one per touched doc + mirror; commits b4067ef526 / 1f129bb4d9 / e2350ad54c / 072e858752).
- **E6 Branch HEAD + push state** — branch `WT-agents-agents-md`; new commits `b4067ef526`, `1f129bb4d9`, `e2350ad54c`, `072e858752` (+ the M5 evidence commit); NOT pushed (lanes do not push; leader handles remote work).
- **E7 Blockers** — none. One discipline correction applied mid-run (coordinator notice): a /tmp scratch dir name carried the card token; renamed to `/tmp/globcheck-scratch/` and the decision-index citation updated — no rework of consumed artifacts.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
audit_ready: true
run_complete_at: 2026-10-20
milestones: M1 b4067ef526 · M2 1f129bb4d9 · M3 e2350ad54c · M4 072e858752 · M5 evidence commit (this commit)
ac_matrix: 16/16 PASS (§E.2 E1; evidence in this tree, HEAD at capture 072e858752)
debts_disposed:
- D5 (verdict): AC-ACR-004 verification used positive gate-marker greps (codex-review-state / Failed / 1841+evidence / t1868+marker) — disposed.
- D6 (verdict): AC-ACR-007 verification used claim-shaped greps (gh pr create --draft operative + t1882 marker; shipped-default/already-implements 0) — disposed.
- AC16-MIX: occurrence-level (^|[^/]) variant adopted as additional evidence; its single hit is the invariant statement itself — disposed.
- PREFLIGHT-HEAD: run-entry HEAD d01b8185f0 recorded; PIN 50,579 matched; never substituted into the projection — disposed.
- dispatch-brief D5 content (review-fix return path): §4.0 item 6 covers re-entry, fix-push to the same branch, window/slot release, and record-based lease reading — disposed.
- dispatch-brief D6 content (same-SHA rule + stranded auto_merge arm): §4.0 items 3-4 cover both — disposed.
- D7-RESIDUAL (dispose_in=sync): MF-6 evidence-cell wording — remains for the sync phase (sync-phase owned).
golangci_lint: `golangci-lint run --timeout=2m` → 2 issues (staticcheck, internal/cli/paths.go:172 ST1008 ×2) — pre-existing baseline; this card touched no .go files (diff list rows 1-15 contain none).
notes: spec.md status transition carried the docs-card exception trailer (recorded in §E.2).

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete
sync_complete_at: 2026-10-20
sync_commit_sha: pending-backfill-sync
b12_self_test_a: pre-emission `grep -c 'SPEC-AGENTS-CONTRACT-REVISE-001' CHANGELOG.md` → 0 (exit 1; no duplicate entry, emission cleared)
b12_self_test_b: AC live count 16 (awk counter on acceptance.md: `live=16 excluded=0 ambiguous=0`; no [RETIRED]/[REF] tokens in the file) == CHANGELOG entry AC count 16
b12_self_test_c: all 18 claimed file paths verified via `ls` (15 diff-list documents + 3 template mirrors) — 18 OK, 0 MISSING
changelog_entry_position: [Unreleased] > Changed (top entry)
frontmatter_status_transitions: spec.md `status: in-progress → completed` on this single sync commit (the implemented intermediate merged per the 3-phase close, manager-docs); `updated: 2026-10-20` already current, unchanged; plan.md / acceptance.md carry no frontmatter (§ Artifact Statelessness) — no `updated:` refresh surface exists
canary_compliance_check: n/a — docs card; this SPEC defines no forward-looking policy with its own sync-test canaries
mx_scan: `grep -n "@MX:"` over the 16 touched surfaces (AGENTS.md, AGENTS.local.md, companion, 2 local rules, 3 factory-dispatch + 3 mirrors, AGENTS.md.tmpl, 6 .moai/docs) → 0 hits (docs card, no exported-code surfaces; expected minimal, recorded)
d7_residual: disposed in sync — unit attribution of record added to decision-index.md (four-chunk figures are awk byte-lengths, C-locale `length()`, not `wc -m` counts); the spec.md:69 MF-6 evidence-cell label "awk char-length" is left as the documented cosmetic residue (spec body content is manager-docs-forbidden; verdict rules it imprecise-not-false; the same row's finding column already carries the precise unit — a literal cell relabel would require manager-spec re-delegation)
notes: the `OwnershipTransitionInvalid` lint warning, if emitted for the run-phase `draft → in-progress` transition (commit b4067ef526, trailer manager-docs), is the documented docs-card exception (§E.2) — known, not new.

## §F Phase 4 Mode Selection

- Input parameters: tier M · docs card (markdown + template mirrors + 2 Go-adjacent test surfaces) · domain count 2 (docs + template build cycle) · language mix ~100% markdown + 1 yaml.tmpl · concurrency benefit LOW (single-writer doc revision, inter-file consistency obligations) · agent-team prereqs: not requested.
- Mode evaluation: direct — not selected (multi-file doc revision, specialist-owned artifacts); fanout — not selected (coding/consistency-heavy, parallel write risk on shared doctrine files); sweep — not selected (semantic doc work, no uniform mechanical transform); **serial** — selected.
- Decision: serial
- Justification: doc-revision work with strict cross-file consistency (mirror parity, moved-clause reconciliation, conflict greps) is coding-heavy in the Anthropic caveat sense — sequential single-specialist (manager-docs-led, per plan.md M1-M5) keeps one writer per tree and preserves the ownership matrix boundaries.
- Plan→Run Kickoff gate (default autonomous form, auto-semantics §9.1) — MET, decision record: (1) independent audit cross = plan-auditor iter-3 PASS-WITH-DEBT 0.88 (Tier M threshold 0.80) @ HEAD 7d1fb4059a, verdict `.moai/reports/t1900/verdict.md` (iter-1..3 history preserved), codex required backend **pass** 0 findings, receipt rcpt-07198228b0f421603bcdd2ad, convergence_check ok:true unmet:[]; (2) evidence criteria = plan_status audit-ready (§E.1), plan-artifact set unchanged since the verdict (amend 7d1fb4059a landed BEFORE the iter-3 audit; `git status` clean), no open blocker; (3) decision = enter run phase, progression mode autonomous (no goal arming — serial lane flow), debts D5·D6·D7-RESIDUAL·AC16-MIX·PREFLIGHT-HEAD routed to run per the verdict's routing. Judgment channel: operator-directed flow (Jev/Codex not consulted — mechanical gate with all evidence on disk).
