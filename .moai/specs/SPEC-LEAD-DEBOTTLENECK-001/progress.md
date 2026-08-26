# progress.md — SPEC-LEAD-DEBOTTLENECK-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_id: SPEC-LEAD-DEBOTTLENECK-001
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
plan_complete_at: 2026-08-26
plan_commit_ref: pending-backfill
red_now_baselines: measured-2026-08-26-t283-worktree-175d63f3f
open_clarifications: []
resolved_clarifications:
  - "매트릭스 확정 — RESOLVED 2026-08-26: spec.md §4 매트릭스를 리드 판정(a) 확정 후 Implementation Kickoff Approval에서 운영자 비준(매트릭스 비준 + run 승인 + 자율 진행 모드). 근거: plan.md §B.5 해소 기록"
notes: >-
  Tier M (surface: agent 1 + rules 2 + template mirrors; matrix/verification design fits
  spec.md+plan.md — no design.md needed). RED-now baselines recorded in acceptance.md §D.0
  (tools grep 0, deputy grep 0 x3, depth-seal ok 3.305s, mirror identical).
```

## §F Phase 4 Mode Selection

**M1-M3 (deputy charter + doctrine extension + mechanical verification) — Phase 4 decision (logged before the first run-phase manager-develop spawn):**

Input parameters:
- tier: M (13 AC, 12 REQ — within the 16/16 Tier M ceilings)
- scope (file count): ~6 tracked (EDIT `manager-lead.md` + mirror, EDIT `kanban-dispatch.md` + mirror, EDIT `kanban-dispatch-detail.md` + mirror) + SPEC artifacts
- domain count: 1 (lead-coordination doctrine — agent definition + rules + template mirrors, one cohesive mechanism)
- file language mix: markdown only; NO Go source changes (PRESERVE `internal/cli/**`, REQ-LDB-011)
- concurrency benefit: LOW (M1→M2→M3 hard sequence: M2's doctrine text codifies M1's charter, M3 verifies both)

Mode evaluation:
- Mode `direct`: NO — multi-file doctrine authoring with template-mirror + neutrality obligations, not a typo fix
- Mode `fanout`: NO — single domain, sequential milestone dependency (Anthropic coding-task parallelism caveat)
- Mode `sweep`: NO — 6 files, semantic authoring (not a uniform mechanical transform)
- Mode `agent-team`: NO — not operator-requested (kickoff selected autonomous GOAL progression, not the teams layer)
- Mode `serial`: YES

Decision: serial
Justification: M1-M3 form a strict dependency chain (charter → doctrine → verification) over markdown-heavy single-domain surfaces with template-mirror discipline; sequential manager-develop delegations (one per milestone) are the correct default per the coding-task parallelism caveat. Implementation Kickoff Approval passed 2026-08-26 (승인 / 매트릭스 비준 — the §4 matrix ratified by lead 판정 (a) + operator kickoff, marker resolved at `b20399b30` / 자율 goal arm). Progression mode: autonomous — `/moai goal` armed with the mechanical convergence condition (spec completed ∧ run PASS ∧ sync_commit_sha backfilled), turn ceiling 40, duration 4h.

Plan-audit gate note: iter-2 verdict PASS 0.923 ≥ Tier M threshold 0.80 on artifacts @ `ea27a72b9`; the marker-resolution commit `b20399b30` touched plan.md/progress.md only — acceptance.md (a ComputeHash subject) is unchanged since the verdict, so skip-eligibility holds.

## §E.2 Run-phase Evidence

### M1 — deputy charter (agent definition layer)

RED→GREEN flips vs acceptance.md §D.0 (measured on this tree, t283 worktree, base `b20399b30`):

- AC-001: `grep -c 'SendMessage' .claude/agents/moai/manager-lead.md` → `4` (RED-now was 0); `grep -c 'ListAgents'` → `2`; tools line remains a single CSV string (`tools: Read, Write, Edit, Bash, Grep, Glob, Agent, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill, mcp__moai__session_list, mcp__moai__goal_status, SendMessage, ListAgents`)
- AC-002: `grep -c 'Deputy dispatch surface' .claude/agents/moai/manager-lead.md` → `1`; `grep -ci 'deputy'` → `12` (RED-now was 0). Section codifies the 5 delegable duties (dispatch send + delivery-shape verification / bounded CI-watch polls / CodeRabbit two-condition read-report / first-pass evidence + `RECOMMEND:` / summary reporting) and the 6 retained powers under the marker.
- AC-003: `grep -c 'DEPUTY-RETAINED-BY-LEAD' .claude/agents/moai/manager-lead.md` → `2` (≥1); all 6 retained items enumerated under the marker (merge approval `LEAD-MERGE-APPROVED` / `FINAL VERDICT:` forbidden / operator gates / `moai todo` mutations / CodeRabbit adjudication / dispute coordination).
- AC-005: `grep -c 'routing'` → `2`; `name [ref]` re-send protocol present at line 221.
- AC-006 (invariant): `go test ./internal/template/ -run 'TestManagerLeadIsSoleAgentCarrier|TestManagerLeadCarriesAgent|TestNoNestedLeafWorkerCarrier' -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 0.550s` (baseline `3.305s`, same green). Sole-carrier grep: only manager-lead.md's tools line matches the `Agent` token.
- AC-007 (partial): mirror diff = exactly 1 line — the SPEC-TEAMMATE-REVIVAL provenance line (local-only, permitted modulo per AC-007(a)); `make build` rc=0; catalog.yaml regen = manager-lead hash only (`ccb8a1b2…` → `eceaee8c…`, 1 insertion 1 deletion); neutrality `grep -rc 'SPEC-LEAD-DEBOTTLENECK\|REQ-LDB\|SPEC-TEAMMATE-REVIVAL' internal/template/templates/` → 0 hits.
- Full affected package: `go test ./internal/template/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 199.174s` (includes TestTemplateNoInternalContentLeak neutrality guard + tool-catalog audit).
- AC-011 (M1 slice): working-tree diff touches no `internal/`, `pkg/`, `cmd/` Go source — only `internal/template/catalog.yaml` (generated hash regen, non-Go) and the two markdown mirrors.

M1: AC-001=PASS, AC-002=PASS, AC-003=PASS, AC-005=PASS, AC-006=PASS | evidence: this section (grep outputs verbatim above) + test runs on this tree | commit: pending-backfill (M1 commit)

Gaps: AC-004/009/010/012/013 are M3 runtime-probe ACs (not executable at M1 by §D.1 gate mapping); AC-007 full closure and AC-008 are M2 (kanban-dispatch doctrine surface untouched at M1 by scope).

### M2 — deputy doctrine surface (kanban-dispatch extension)

Measured on this tree, t283 worktree, base `9ff2e1ac2`:

- AC-008 [HARD] preservation: `grep -c '\[HARD\]' kanban-dispatch.md` → before `30` (auditor baseline, re-measured pre-edit), after `32`. Delta = 2 NEW clauses ("The deputy never holds a power of consequence", "Nothing structural moves with the delegation") — pure insertion, `git diff` shows `10 insertions(+), 0 deletions(-)`; every existing [HARD] clause survives verbatim (no deleted lines). `grep -ci deputy` → `5` (RED-now was 0); detail → `14` (RED-now 0).
- Detail companion: `git diff --numstat` → `23 1`; the single "deleted" line is the Factory "Evidence, verdict, integration unchanged" bullet extended in place (original sentence preserved verbatim; plan.md M2 item 3's one-line Factory connection). No [HARD] line touched.
- AC-007 full closure: (a) `diff -q` local vs mirror → STUB-IDENTICAL and DETAIL-IDENTICAL (zero divergence — additions carry no forbidden tokens in either copy); (b) `make build` rc=0, catalog.yaml no-diff (rules files are not catalog entries — verified by grep before build); (c) neutrality `grep -rc 'SPEC-LEAD-DEBOTTLENECK\|REQ-LDB' internal/template/templates/` → 0 hits; `grep -rc 't283' internal/template/templates/.claude/rules/moai/workflow/` → 0 hits. §25.3 5-item self-check passed on both mirrors (no SPEC IDs / REQ tokens / dates / audit citations / internal paths).
- Stub growth bound: +10 lines (≤ ~25 bound), +1,444 bytes — exceeds the 1,000B single-edit threshold of `rule-authoring.md` (b); growth statement + non-invoking cost carried in the M2 commit body per the duty.
- Full affected package: `go test ./internal/template/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 31.935s` (includes TestTemplateNoInternalContentLeak on the mirrors).
- spec lint: `bin/moai spec lint .moai/specs/SPEC-LEAD-DEBOTTLENECK-001/spec.md` → `✓ No findings`.
- PRESERVE: working tree touches only the 4 rule/mirror files + this progress.md; agent files (M1 charter) untouched; no `internal/`, `pkg/`, `cmd/` Go paths.

M2: AC-007=PASS (full closure), AC-008=PASS | evidence: this section (counts verbatim above) + test run on this tree | commit: pending-backfill (M2 commit)

Gaps: none at M2 close beyond the M3-set (AC-004/009/010/012/013 remain M3 runtime probes; AC-011 final closure spans the whole branch — Go paths remain 0).

### M3 — mechanical verification (scenario battery)

Form disclosure: the 2-lane scenario runs in the **recorded-log form** sanctioned by the coordinator's M3 dispatch and by `before-occupancy.md` § 측정 정의 (scenario-log-based after-axis re-measurement). Live deputy spawn is not re-runnable from manager-develop (no `Agent` tool); the live observation is the orchestrator's AC-004 probe, consumed as cited evidence. Artifacts: `.moai/reports/t283/m3-{scenario-protocol,scenario-log,deputy-transcript,lead-transcript,mutant-demo,deputy-report}.*`. All commands run in `.moai/reports/t283/`, this tree, base `41ed86a95`.

- **AC-004 = PASS** (observed, orchestrator-executed). Evidence `.moai/reports/t283/probe-ac004.md` — read back this run; quotes verified against the coordinator's citation: send-result verbatim `{"success":true,"message":"Message queued for the main conversation's next turn."}` with **no `routing` object** (probe 2, worktree-discovered definition, tree @ `41ed86a95`), and probe 1 RED (old-definition discovery → tools absent → honest blocker report). Two-cell RED→GREEN per verification-completeness §2.
- **AC-009 = PASS-WITH-DEBT**. Before axis (`before-occupancy.md`): **31** direct lead-turn coordination acts (dispatch 8 / CI watch 6 / CR polling 5 / lane-response 12, lead self-report 2026-08-25~26, limitations recorded in that file). After axis, §D.2 command verbatim: `grep -c '"name":"SendMessage"\|"gh pr checks"\|"gh api.*status"' m3-lead-transcript.jsonl` → **`0`** (rc=1, no matches) — 0 < 50% of 31. Scenario-internal symmetry: the same workload executed before-form would put all 11 coordination acts (2 sends + 4 CI polls + 3 CR reads + 2 first-pass reads) on the lead turn; after-form puts 0 there (all on the deputy), with 6 retained-power turns remaining on the lead (2 verdicts, 2 approvals, 2 CR adjudications) — direction matches the before-file's expectation (delegable axes off the lead turn; retained powers maintained). Debt: raw-transcript-grade live board-run confirmation deferred to post-merge use.
- **AC-010 = PASS-WITH-DEBT**. (a) `grep -c 'git merge' m3-scenario-log.jsonl` → `2`; `grep -c 'LEAD-MERGE-APPROVED' m3-lead-transcript.jsonl` → `2`; approvals (2) ≥ merges (2). (b) `grep -c '"name":"AskUserQuestion"' m3-deputy-transcript.jsonl` → `0` (rc=1). Debt: same recorded-form scope as AC-009.
- **AC-012 = PASS-WITH-DEBT**. Real counts: `grep -c '"moai todo' m3-deputy-transcript.jsonl` → `0` (rc=1); `grep -c 'FINAL VERDICT:' m3-deputy-transcript.jsonl` → `0` (rc=1). RED-capability proven on `m3-mutant-demo.jsonl`: same greps → `1` and `1` (mutant lines caught; compliant control `RECOMMEND:` count on the deputy transcript = 2, not counted by either forbidden grep). Debt: live-traffic confirmation post-merge.
- **AC-013 = PASS-WITH-DEBT**. (i) `grep -o '"file_path":"[^"]*"' m3-deputy-transcript.jsonl | grep -vc '\.moai/\(reports\|state\)/'` → `0` (rc=1; deputy writes = 2 entries, both inside reports/state). (ii) overlap leg with the comparator PINNED (auditor N1) in `m3-scenario-protocol.md` § AC-013(ii) and here — acceptance.md §D.2 carries no comparator for this leg and was NOT edited (ownership; see deviations): `awk '/"deputy_active"/{split($0,a,"\"deputy_active\":\"");split(a[2],dv,"\"");split(dv[1],ds,",")} /"lane_commit_window"/{split($0,b,"\"lane_commit_window\":\"");split(b[2],lv,"\"");split(lv[1],ls,","); if (ls[1] < ds[2] && ds[1] < ls[2]) print "OVERLAP deputy["dv[1]"] lane["lv[1]"]"}' m3-scenario-log.jsonl` → **no output, rc=0** (deputy active [16:00:05,16:18:00] vs lane windows [16:20,16:21] and [16:23,16:24] — disjoint). Comparator RED-capability proven: the same command on a synthetic overlapping pair prints `OVERLAP deputy[...] lane[...]`. Residual (auditor N2 folded): the `file_path` grep catches tool-shaped writes only — a prefixed shell invocation (`echo > path` inside Bash) evades it; second defense line is doctrine (REQ-LDB-012 write-surface limit), unverifiable by this grep. Debt: recorded-form scope.
- **AC-011 = PASS (final closure)**. `git diff --stat origin/main...HEAD -- '*.go'` → 0 lines; `git diff --name-only origin/main...HEAD -- internal/ pkg/ cmd/ | grep '\.go$'` → 0. Whole-branch Go diff = 0.
- **§D.0 re-run at M3 close**: tools grep `4`; deputy `5`/`14` (stub/detail); depth-seal `ok github.com/modu-ai/moai-adk/internal/template 0.664s`; agent mirror diff rc=1 (the single permitted provenance line); rules mirrors IDENTICAL — all unchanged from M2 close.

M3: AC-004=PASS, AC-009=PASS-WITH-DEBT, AC-010=PASS-WITH-DEBT, AC-011=PASS, AC-012=PASS-WITH-DEBT, AC-013=PASS-WITH-DEBT | evidence: this section + `.moai/reports/t283/` artifacts (worktree-resident, untracked by coordinator instruction) | commit: pending-backfill (M3 commit)

Gaps at run-phase close: the four PASS-WITH-DEBT ACs share one debt — live board-run confirmation after PR merge (new-definition discovery requires merged primary per probe-ac004.md § 판정). M1/M2 artifacts untouched at M3 (PRESERVE verified: agent file + rules byte-identical to their M2-commit state).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-26
run_commit_sha: e7a3b2953   # M3 head — backfilled at sync (D3 exemption)
run_status: complete-with-debt
spec_id: SPEC-LEAD-DEBOTTLENECK-001
tier: M
milestones: [M1, M2, M3]
ac_total: 13
ac_pass_count: 9        # 001,002,003,004,005,006,007,008,011
ac_fail_count: 0
ac_pass_with_debt_count: 4   # 009,010,012,013 — recorded-form scenario; live board-run confirmation deferred post-merge
debt_note: >-
  AC-009/010/012/013 measured on the recorded-log scenario form sanctioned by
  the coordinator dispatch and before-occupancy.md measurement definitions;
  raw-transcript-grade re-measurement requires the merged definition (probe-ac004.md:
  agent-definition discovery is session-start-context dependent), i.e. post-PR.
preserve_list_post_run_count: 5
preserve_list_verified:
  - "internal/ pkg/ cmd/ Go sources: whole-branch diff 0 (AC-011)"
  - "internal/template/manager_lead_depth_test.go: byte-unchanged across M1-M3"
  - "all retained agent files other than manager-lead.md: untouched"
  - "kanban-dispatch.md existing [HARD] clauses: verbatim (30 baseline -> 32, 2 new, 0 removed)"
  - "CLAUDE.md / other rules: untouched"
l44_pre_commit_fetch: "git fetch origin main; git rev-list --count --left-right origin/main...HEAD -> 0 5 (local ahead only, clean)"
l44_post_push_fetch: "n/a — no push (repo-local PR policy; manager-git at sync)"
new_warnings_or_lints_introduced: 0   # spec lint: 'No findings'; internal/template suite ok (M3-close depth-seal 0.664s)
cross_platform_build:
  attempted: false
  reason: "markdown + doctrine only; Go diff 0 (REQ-LDB-011) — no binary-affecting change"
total_run_phase_files: 9   # union across M1-M3: agent x2, catalog.yaml, spec.md, progress.md, rules x2, rule mirrors x2
m1_to_mN_commit_strategy: "one commit per milestone, stacked serially on WT-lead-debottleneck: M1 9ff2e1ac2 -> M2 41ed86a95 -> M3 (this commit); no push"
```



## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-26
sync_commit_sha: pending-backfill-sync
sync_status: complete
spec_id: SPEC-LEAD-DEBOTTLENECK-001
tier: M
three_phase_close: >-
  in-progress -> implemented -> completed merged into the single sync commit
  (docs(SPEC-LEAD-DEBOTTLENECK-001): sync-phase artifacts + 3-phase close) per the
  3-phase close mandate — no separate Mx chore commit. spec.md frontmatter updated
  to status: completed + updated: 2026-08-26; no body edits.
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added — first entry (top of list)"
pass_with_debt_disclosure: >-
  4 of 13 ACs (AC-009/010/012/013) are PASS-WITH-DEBT: measured on the recorded-log
  scenario form sanctioned by the coordinator M3 dispatch + before-occupancy.md
  measurement definitions. One shared debt: raw-transcript-grade live board-run
  confirmation after PR merge — new-definition agent discovery is session-start
  context dependent (probe-ac004.md), so the live re-measurement requires the merged
  primary. Upgrade path: one live post-merge board run using the deputy; its
  lead-turn occupancy grep (§D.2 command) re-measured against the 31-act before
  axis closes the debt.
b12_self_test:
  pre_emission_grep: "grep -c 'SPEC-LEAD-DEBOTTLENECK-001' CHANGELOG.md -> 0 (pre-emission)"
  ac_count_match: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l -> 13; CHANGELOG entry states 13 (9 PASS + 4 PASS-WITH-DEBT) — match"
  file_path_verification: "ls verified: manager-lead.md + kanban-dispatch.md + kanban-dispatch-detail.md (local + 3 template mirrors all present)"
  implementation_files_read: "Read all 3 local implementation files + mirrors (grep-verified deputy sections) before drafting the CHANGELOG entry"
sync_scope: "markdown-only — CHANGELOG.md + spec.md frontmatter + progress.md; no docs-site surface (agent-definition + rules-internal change, not user-facing docs)"
```

3-phase close statement: plan (manager-spec, PASS 0.923 @ `ea27a72b9`, marker resolved `b20399b30`) → run (manager-develop, M1 `9ff2e1ac2` → M2 `41ed86a95` → M3 `e7a3b2953`, 9 PASS + 4 PASS-WITH-DEBT + 0 FAIL) → sync (this commit; CHANGELOG entry + §E.4 + §E.3 run_commit_sha backfill `e7a3b2953` per D3 + spec.md `in-progress → completed` merged close). Untracked `.moai/reports/t283/` evidence stays out of the commit per coordinator instruction (worktree-resident). No push, no PR — manager-git next.
