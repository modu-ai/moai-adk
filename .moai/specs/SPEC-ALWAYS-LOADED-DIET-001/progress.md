# Progress — SPEC-ALWAYS-LOADED-DIET-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- tier: M (3 artifacts: spec.md + plan.md + acceptance.md)
- REQ count: 16 / 16 (Tier M ceiling)
- AC count: 16 / 16 (Tier M ceiling)
- open items: none — the 3 former plan.md §D4 open questions are resolved by user decision (budget ratchet deferred to a separate backlog card; M3 control is documentation-only, no Go change; growth-statement threshold = 1,000 bytes per single edit). See plan.md §D4-1..D4-3.
- baseline observed (2026-08-16): `files=14 bytes=295044 tokens=73761 headroom=1239`; `go test ./internal/config/ -run TestAlwaysLoaded` → ok
- plan_complete_at: 2026-08-17 (산출물 커밋 be1958a4d, 2026-08-17 00:32 +0900 — 실제 실행일)
- plan_audit: iter1 FAIL 0.73 (8 blocking) → iter2 **PASS 0.845** vs Tier M threshold 0.80 (Clarity 0.85 / Completeness 0.88 / Testability 0.75 / Traceability 0.92); re-audit ceiling (Tier M = 2) reached, no iter3
- plan_audit reports: `.moai/reports/plan-audit/SPEC-ALWAYS-LOADED-DIET-001-review-{1,2}.md`
- iter2 blocking 3 applied orchestrator-direct (auditor's own recommendation — local shell edits, ~15 lines):
  - D1 `AC-ALD-009` passed with no companion (`wc -c` failure → empty → bash arithmetic 0 → `sum` equals the original 21,003, which is exactly the PASS lower bound). Fixed with `test -f` + `companion >= 1`. Re-run on the untouched tree → `MISSING …detail.md`, exit 1.
  - D2 six AC touched files with no existence guard, breaching this document's own §A trap rule 6. `AC-ALD-006` actually PASSed (`missing_lines=0`). Guards added to AC-ALD-004/005/006/008/009/013; `AC-ALD-006` re-run → exit 1.
  - D3 `REQ-ALD-013` claimed all four guard slots while the glob covered three. `,**/MEMORY.md` added (13 chars, zero always-loaded cost) in spec.md §3.3 + plan.md D1/M3. Rationale: the guard counts that slot conditionally, so a future repo-root `MEMORY.md` would admit up to `memoryHeadByteCap` 25,600 B (~6,400 tokens) unstated — larger than the worst-corner headroom of 2,597 tokens.
- post-fix verification (this tree): `moai spec lint …/spec.md` → `✓ No findings`, exit 0; REQ 16 / AC 16 unchanged
- repository state (2026-08-17 갱신): 산출물 4개는 커밋됨 — 브랜치 `spec-always-loaded-diet-plan` 의 `be1958a4d`, PR #1576 에 포함. 리포 로컬 all-tier PR 정책은 브랜치+PR 로 충족했고 머지 대기.
- revision (2026-08-17): PR #1576 CodeRabbit 인라인 리뷰의 🟠 Major 5건 반영 — acceptance.md(§A 규율 4 + `headroom` frontmatter 한정, AC-ALD-006 `BASE_REF` 고정 + `sort -u` 제거, AC-ALD-013 문단 스코프 인용 판정, AC-ALD-014 슬롯 4개 단언)과 progress.md 본 기록. 🟡 Minor 4건·🔵 Trivial 2건은 사용자 결정으로 이번 개정 범위 밖.
- Implementation Kickoff Approval (plan→run) has NOT been requested or granted.

## §E.2 Run-phase Evidence

### M1 — `kanban-dispatch.md` 분리 (2026-08-16)

- Phase 1 re-audit: **skipped as eligible** — plan-audit iter2 PASS 0.845 (Tier M threshold 0.80) + artifact content unchanged since the verdict (squash-merge of PR #1576 did not alter artifact bytes).
- Implementation Kickoff Approval: granted orchestrator-side (run-phase M1 dispatch received 2026-08-16). The §E.1 "NOT been requested" line reflects plan-phase authoring time.
- Tree: worktree `feat/spec-always-loaded-diet`, pre-M1 HEAD `062a995d9` (base origin/main).

**Construction (verbatim move, no rewrite — REQ-ALD-005/AP-2):**

- `kanban-dispatch-detail.md` = frontmatter + ownership declaration + `sed -n '13,59p;62,92p;124,152p'` of the original. Line 61 dropped as the collapsed duplicate blank at the Class-B lift-out point (plan M1 step 3's 1-byte cleanup).
- Stub = stay lines `1,6 / 7,12 / 93,98 / 60 / 99,123 / 153,230` + pointer line (after the Loading-scope quote) + one separator blank + footer version line. The Class B paragraph (original line 60, 495 B) relocated **whole-line** into `## Completion is read, never trusted`, between "Before moving a card…" and "This applies equally…".

**Baseline (pre-edit, this run, this tree, HEAD 062a995d9):**

```
files=14 bytes=295044 tokens=73761 headroom=1239   # acceptance §A fm() form
bytes=295044 tokens=73761 headroom=1239             # plan §C grep -rL form (identical)
```

**Post-split (this run, this tree, working tree):**

```
files=14 bytes=287068 tokens=71767 headroom=3233    # fm() form
bytes=287068 tokens=71767 headroom=3233              # grep -rL form
```

Delta: −7,976 bytes → headroom 1,239 → 3,233 (+1,994). AC-ALD-001 checkpoint (final gate is post-M4) already holds: 3,233 > 1,239.

**M1-decidable AC matrix (all PASS, commands run in this tree):**

| AC | Command (form) | Observed | Verdict |
|---|---|---|---|
| AC-ALD-003 | 9 BINDING headings `grep -c` in stub; 6 LEAD-ONLY `grep -c` in stub; anchor `names that path in its completion report` both files | 9×1; 6×0; relocated_in_stub=1 relocated_in_companion=0 | PASS |
| AC-ALD-004 | `grep -c '\[HARD\]'` stub / companion | stub=6 companion=3 (sum 9 = original census 9) | PASS |
| AC-ALD-005 | 7 headings `grep -c` in companion | 7×1 | PASS |
| AC-ALD-006 | `comm -23` orig(BASE_REF=be1958a4d) vs stub+companion, non-empty sorted | missing_lines=0; positive contrast vs /dev/null=140 | PASS |
| AC-ALD-007 | `fm()` paths keys | manager-kanban_key=1 todo_key=1 (domain-keyed, plan D2) | PASS |
| AC-ALD-008 | (a) names within ±2 lines of the `kanban-dispatch-detail.md` mention (pipe form), (b) ownership_decl, (c) footer | (a) 6×1, (b) ownership_decl=3, (c) version line present | PASS |
| AC-ALD-009 | `wc -c` both files | stub=13027 companion=8866 sum=21893 (range 21003..21903; overhead 890 B, within the 600–900 estimate) | PASS |
| AC-ALD-012 | `headroom` files= + `fm()` paths_key | files=14, exists=1 paths_key=1 (cache-aware-execution-reference.md is M2 — AC completes at M2) | PASS (M1 scope) |
| AC-ALD-002 | guard + budget const + go diff | `ok github.com/modu-ai/moai-adk/internal/config 0.607s` exit=0; budget_const=1; go_changed=0 | PASS |

Note (AC-ALD-008a form): acceptance.md's `<(grep -A2 -B2 …)` process-substitution form yields empty `grep -c` output under this environment's zsh; the pipe form `grep -A2 -B2 … | grep -c` is semantically identical and was used. Bash (sync-audit) runs the acceptance form directly.

**Other verification:** `go build ./...` → exit 0. `go test ./internal/config/` (full package) → `ok github.com/modu-ai/moai-adk/internal/config 2.551s`. Subagent boundary: `grep -rn 'AskUserQuestion'` on the two rule files → 2 matches, both pre-existing doctrine-text mentions (original census 2: 1 stayed in stub `Boundaries`, 1 moved verbatim to companion `Entry into the board`); no new occurrences. golangci-lint: n/a — 0 Go files changed; markdown has no lint gate. RED-state evidence: n/a — documentation milestone, no test-first artifact (recorded as a gap, not fabricated).

**Residual (recorded, deliberately not fixed):** the companion line "This is the same shape as the CodeRabbit section below." now reads stale inside the companion (the CodeRabbit section stayed in the stub). Rewording it would break AC-ALD-006's line-set preservation and violate AP-2 — the line must stay verbatim; the referenced section exists in the stub.

### M2 — G3~G7 규율 편입 (2026-08-16)

- Tree: worktree `feat/spec-always-loaded-diet`, pre-M2 HEAD `a203a7c3a` (= the M1 commit; working tree clean at dispatch).
- Construction: directives 6-10 appended to `cache-aware-execution.md` `## Directives` — same shape as 1-5 (numbered + bold title + zone prefix + one paragraph), prefix `[ZONE:Evolvable] [HARD]` per REQ-ALD-006..010. Per-directive bytes (wc-accurate): d6=339 / d7=360 / d8=313 / d9=337 / d10=437. Plus one Cross-references bullet for the companion (111 B) and footer version 1.0.0 → 1.1.0. Nothing else in the file touched — `git diff --stat`: 12 insertions, 1 deletion (the deletion is the old version line). New companion `cache-aware-execution-reference.md` (3,991 B): frontmatter `description:` + self-keyed `paths: "**/cache-aware-execution.md"` (plan M2 step 2 — rationale-store, no domain keying), ownership-declaration blockquote (goal-directive-detail.md shape), cited-numbers section, per-directive rationale for 6-10.
- G7 / D3 resolution: directive 10 carries the one-sentence axis distinction (main-session cache vs `agent-common-protocol.md` § Per-Spawn Model Injection subagent model-resolution) and names both SSOTs inline; neither SSOT body was revised (per plan D3 — the cross-reference is textual from d10 only, so `agent-common-protocol.md` is untouched).

**Baseline (pre-edit, attribution: M1 evidence measured at HEAD a203a7c3a, unchanged until this edit):**

```
headroom (fm() form, at a203a7c3a per §E.2 M1): files=14 bytes=287068 tokens=71767 headroom=3233
cache-aware-execution.md = 4,497 B
```

**AC-ALD-010 positive contrast (pre-edit, this run, this tree, HEAD a203a7c3a):** all six patterns 0 — `G3_at_mention=0 G3_context_audit=0 G4_output_len=0 G5_quiet=0 G6_session_length=0 G7_thinking=0`.

**Post-M2 (this run, this tree, working tree):**

```
files=14 bytes=288975 tokens=72243 headroom=2757    # acceptance §A fm() form
n=$(wc -c < .claude/rules/moai/workflow/cache-aware-execution.md)
before=4497 after=6404 delta=1907
```

Delta: +1,907 bytes (= 5 directive paragraphs 1,786 B + 5×2 separator newlines + companion cross-reference bullet 111 B) → headroom 3,233 → 2,757 (−476 tokens). Byte identity: 287,068 + 1,907 = 288,975 exactly. AC-ALD-001 checkpoint holds: 2,757 > 1,239 (final gate is post-M4).

**M2-decidable AC matrix (all PASS, commands run in this tree):**

| AC | Command (form) | Observed | Verdict |
|---|---|---|---|
| AC-ALD-010 | 6-pattern `grep` on post-edit file (positive contrast above: pre-edit 6×0) | G3_at_mention=1 G3_context_audit=1 G4_output_len=1 G5_quiet=2 G6_session_length=1 G7_thinking=1 — all >= 1 | PASS |
| AC-ALD-011 | `wc -c` before/after | before=4497 after=6404 delta=1907; 1000 <= 1907 <= 2000 | PASS |
| AC-ALD-012 | `headroom` files= + `fm()` paths_key on companion | files=14 (unchanged — companion excluded by `paths:`), exists=1 paths_key=1 | PASS |
| AC-ALD-013 | unlabeled-numeric-paragraph awk + negative selftest | numeric_paragraphs=2 (table block + reconciliation paragraph, both labeled), citation_exit=0, selftest_unlabeled=1 | PASS |
| AC-ALD-002 (checkpoint) | guard + budget const + go diff | `ok  github.com/modu-ai/moai-adk/internal/config 0.494s` exit=0; budget_const=1; go_changed=0 | PASS |

**Post-M2 guard re-run (verbatim):**

```
$ go test ./internal/config/ -run TestAlwaysLoaded -count=1; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/config	0.494s
exit=0
```

**Other verification:** `go build ./...` → exit 0 (no output). Subagent boundary: `grep -rn 'AskUserQuestion'` on the two rule files → 1 match, pre-existing intro text (line 3, present in the pre-edit file; original census unchanged), 0 new occurrences; companion 0. golangci-lint: n/a — 0 Go files changed; markdown has no lint gate. RED-state evidence: n/a — documentation milestone, no test-first artifact (recorded as a gap, not fabricated; E8 docs-milestone gap).

### M3 — 재발 통제 (2026-08-16)

- Tree: worktree `feat/spec-always-loaded-diet`, pre-M3 HEAD `6d701d26c` (= the M2 commit; working tree clean at dispatch).
- Construction: new `.claude/rules/moai/development/rule-authoring.md` (4,062 B). Frontmatter: `description:` + the plan D1 expanded glob, verbatim — `paths: "**/.claude/rules/**,**/CLAUDE.md,**/.claude/output-styles/**,**/MEMORY.md"` (keys all four guard slots; AP-6 satisfied — the control itself is `paths:`-scoped and excluded from the surface it polices). Body: four-slot surface definition (no-`paths:` rules / `CLAUDE.md` / output styles / `MEMORY.md` head) with the all-slots binding rationale; recurrence grounding (~4x growth in three months, ~half in-place expansion → both modes covered); the statement duty as one `[ZONE:Evolvable] [HARD]` clause with four lettered parts — (a) new always-loaded file → bytes + cost justification in the change description, (b) single edit growing an existing file by more than 1,000 bytes → same statement sized to the growth, (c) the statement must address the cost paid by sessions that never need the file (grounded per plan M3 step 3: `session-handoff.md` closest cost-naming precedent; `native-idiom-and-register.md` "zero burden for English sessions" true for behavior, false for context bytes), (d) `paths:`-scope-first question before the duty fires; threshold calibration table (typo ~100 B / one-line ~200 B / paragraph ~800 B pass; new HARD clause ~1,200 B / new `##` section ~2,500 B fire); explicit no-cumulative-delta-secondary-trigger clause (accepted residual risk per plan D4-3); self-compliance note. Size calibration vs development-category siblings: branch-origin-protocol.md 4,180 B, karpathy-quickref.md 2,068 B.
- Zero Go changes (plan M3 step 5 / D4-2): one file added; `internal/` untouched.

**Post-M3 instrumentation (acceptance §A `fm()`/`headroom` forms, this run, this tree, HEAD `6d701d26c` + working-tree rule-authoring.md):**

```
paths_key=1
files=14 bytes=288975 tokens=72243 headroom=2757
```

Surface delta: **0 bytes** — rule-authoring.md carries top-level `paths:` so the guard excludes it; files stays 14, bytes/headroom byte-identical to post-M2 (288,975 / 2,757). AC-ALD-001 checkpoint holds: 2,757 > 1,239 (final gate is post-M4).

**M3-decidable AC matrix (all PASS, commands run in this tree):**

| AC | Command (form) | Observed | Verdict |
|---|---|---|---|
| AC-ALD-014 | acceptance §B verbatim form (`test -f` guard, `fm()` paths line, 4 slot fragments via `grep -F`, 4 body patterns) | paths line exact; slot_rules=1 slot_CLAUDE.md=1 slot_output-styles=1 slot_MEMORY.md=1; new_rule_case=1 growth_case=5 threshold=3 non_invoking_cost=1 | PASS |
| AC-ALD-001 (checkpoint) | `headroom` | files=14 bytes=288975 tokens=72243 headroom=2757 (> 1,239) | PASS |
| AC-ALD-002 (checkpoint) | guard + budget const + go diff | `ok  github.com/modu-ai/moai-adk/internal/config 0.457s` exit=0; budget_const=1; go_changed=0 | PASS |

**Post-M3 guard re-run (verbatim):**

```
$ go test ./internal/config/ -run TestAlwaysLoaded -count=1; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/config	0.457s
exit=0
```

**Other verification:** `go build ./...` → exit 0 (no output). `go test ./internal/config/ -count=1` (full package) → `ok  	github.com/modu-ai/moai-adk/internal/config	5.690s` exit=0. Subagent boundary: `grep -cn 'AskUserQuestion\|mcp__askuser'` on rule-authoring.md → 0. Neutrality pre-check (M4 mirrorability, local file): spec_id=0 req=0 sha=0 date=0 — no sanitization needed at mirror time. golangci-lint: n/a — 0 Go files changed; markdown has no lint gate. RED-state evidence: n/a — documentation milestone, no test-first artifact (recorded as a gap, not fabricated; E8 docs-milestone gap).

### M4 — 미러 + 빌드 + 최종 계측 (2026-08-16)

- Tree: worktree `feat/spec-always-loaded-diet`, pre-M4 HEAD `fcc07d1bc` (= the M3 commit; working tree clean at dispatch — verified `git status --short` empty).
- Mirror scope: the 5 created/modified rule files, copied to `internal/template/templates/` at the corresponding paths. Pre-mirror baseline established: the two MODIFIED files' pre-edit dev versions (`a203a7c3a~1` kanban-dispatch.md, `6d701d26c~1` cache-aware-execution.md) are **byte-identical to the template copies** (`diff` exit 0, 0 lines) — the mirror applies to a clean baseline, no foreign template divergence to preserve.
- Neutrality of the sources (pre-copy, this run, this tree): all 5 dev files scanned 0 across `SPEC-ALWAYS` / `REQ-ALD|AC-ALD` / generic `SPEC-…` / generic `REQ/AC-…-NNN` / `[0-9a-f]{40}` / `[0-9a-f]{7,8}+punct` / `2026-` / `/Users/goos` / `CLAUDE.local` / `Audit N Finding` — the rule bodies were authored neutral from M1 on (provenance lives in this file, not the rules). **Sanitization applied: none** — the byte-identical copy IS the sanitized form; plan §F's "no wholesale cp" guards against dirty sources, and the scan is the evidence this source is clean.
- `make build`: exit 0. `catalog.yaml` regeneration ran (12407 bytes) with **zero hash change** (rule files are not catalog entries — `grep -c 'rules/moai' catalog.yaml` = 0), so no build artifacts enter the commit; `//go:embed all:templates` compiles the FS into the binary with no generated files. `git status` shows exactly the 5 mirror paths.

**Post-M4 final measurement (acceptance §A `headroom` form + plan §C `grep -rL` form, this run, this tree, working tree with mirrors):**

```
files=14 bytes=288975 tokens=72243 headroom=2757    # acceptance §A fm() form
bytes=288975 tokens=72243 headroom=2757              # plan §C grep -rL form (identical)
```

**Final gate: headroom 2,757 > 1,239 (strictly greater).** Byte-identical to post-M3 (288,975 / 2,757) — M4 touches only `internal/template/templates/`, which the guard does not measure; the measurement chain M1(3,233) → M2(2,757) → M3(2,757) → M4(2,757) is monotone above baseline at every checkpoint and lands above the plan's worst-corner projection (2,597). Observed headroom recorded for the budget-ratchet backlog card per plan D4-1: **2,757 tokens** (input to the separate card; `AlwaysLoadedTokenBudget` unchanged at 75,000 — D4-1).

**M4-decidable AC matrix (all PASS, commands run in this tree):**

| AC | Command (form) | Observed | Verdict |
|---|---|---|---|
| AC-ALD-001 (FINAL) | `headroom` | headroom=2757 > 1239 | PASS |
| AC-ALD-015 | acceptance §B verbatim loop + `make build` | 5×MIRRORED, MISSING 0, build_exit=0 | PASS |
| AC-ALD-016 | acceptance §B verbatim `test -f` + 4-pattern `grep -cE` + strict leak test | 5 rows `spec_id=0 req=0 sha=0 date=0`; `MOAI_TEMPLATE_LEAK_STRICT=1 go test … -run TestTemplateNoInternalContentLeak` → ok, exit=0; positive contrast live (same SPEC pattern on local spec.md = 3 matches — not a false-pass machine) | PASS |
| AC-ALD-002 (final) | guard + budget const + go diff (`origin/main...HEAD`) | `ok … 0.281s` exit=0; budget_const=1; go_changed=0 | PASS |
| AC-ALD-007 (mirror re-check) | `fm()` paths keys on mirrored companion | mirror carries the same domain-keyed `paths:` line (byte-identical copy) | PASS |

**Other verification:** `go build ./...` → exit 0. `GOOS=windows go build ./...` → exit 0; `GOOS=linux go build ./...` → exit 0 (embed compiles cross-platform; markdown-only change). `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1` (FULL package, includes `TestSanitizedPairParity` + `TestRuleTemplateMirrorDrift` + `TestTemplateNeutralityAudit`) → `ok … 37.769s` exit=0. The 3 CI `template-neutrality-check.yaml` targets run in isolation exactly as CI runs them: `TestTemplateNeutralityAudit` → PASS, `TestTemplateNoInternalContentLeak` (narrow) → PASS, strict tier → PASS (all exit 0). `go test ./internal/config/ -count=1` (full package) → `ok … 1.924s` exit=0. Subagent boundary: `grep -c 'AskUserQuestion\|mcp__askuser'` on the 5 mirrors → kanban stub 1 + kanban companion 1 + cache parent 1 + cache reference 0 + rule-authoring 0; sum 3 = original census sum 3 (kanban original 2 — 1 stayed in stub `Boundaries`, 1 moved verbatim to companion; cache original 1 — pre-existing intro line); 0 new occurrences, all doctrine-text mentions, none a subagent invocation. golangci-lint: n/a — 0 Go files changed (`git diff --name-only origin/main...HEAD -- '*.go'` = 0); markdown has no lint gate. RED-state evidence: n/a — documentation+mirror milestone, no test-first artifact (recorded as a gap, not fabricated; E8 docs-milestone gap).

**Residual (recorded, deliberately not fixed):** the 5 mirrored files are enrolled in NO parity registry (`workflowOptMirroredPaths`, `sanitizedPairPaths`) — enrollment would be a Go test-file edit outside this SPEC's §F scope envelope (plan §A: "편집 대상 파일 4개 + 미러 4개 + M3 신설 1개(+미러) = 최대 10개 파일"). Consequence: a future single-tree edit to these 5 rules is not caught by `TestRuleTemplateMirrorDrift`; the byte-parity enrollment is a natural candidate for a follow-up card alongside the D4-1 budget ratchet.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-08-16
spec_id: SPEC-ALWAYS-LOADED-DIET-001
branch: feat/spec-always-loaded-diet
worktree: /tmp/wt-run-ald
base: 062a995d9   # origin/main at M1 dispatch; plan artifacts squash-merged via PR #1576 as be1958a4d
run_phase_commits:
  - a203a7c3a   # M1 kanban-dispatch.md split (stub + domain-keyed detail companion)
  - 6d701d26c   # M2 cache-aware-execution.md directives 6-10 + self-keyed reference companion
  - fcc07d1bc   # M3 rule-authoring.md recurrence control (paths-scoped, 4-slot)
  - 9269a5e56   # M4 template mirror (5 files) + rebuild + final measurement
ac_pass_count: 16
ac_fail_count: 0
final_headroom_tokens: 2757   # observed post-M4, both measurement forms agree; baseline 1239; D4-1 budget-ratchet card input
always_loaded_files: 14   # unchanged pre/post diet (files split, not removed)
bytes_pre: 295044
bytes_post: 288975   # net -6069 B (-1518 tokens); M1 -7976, M2 +1907, M3 0, M4 0
preserve_list_post_run_count: 0   # no Go/config/hook surface touched; go_changed=0 across the run phase
l44_pre_commit_fetch: true        # worktree branch only; primary checkout untouched; explicit-pathspec staging, no -A/-a
l44_post_push_fetch: n/a          # not pushed — branch awaits orchestrator
new_warnings_or_lints_introduced: 0   # 0 Go files changed; golangci-lint n/a; template package green incl. strict tier
cross_platform_build:
  darwin_arm64: pass   # go build ./... exit 0
  windows_amd64: pass  # GOOS=windows go build ./... exit 0 (embed compiles)
  linux_amd64: pass    # GOOS=linux go build ./... exit 0
total_run_phase_files: 10   # 5 dev-surface files (M1-M3) + 5 template mirrors (M4); + progress.md evidence
m1_to_mN_commit_strategy: one-commit-per-milestone (M1/M2/M3) + M4 commit + sha backfill commit
mirror_status:
  mirrored: 5
  sanitization_applied: none   # sources authored neutral; byte-identical copy is the sanitized form (scan evidence in §E.2 M4)
  parity_registries_enrolled: 0   # residual, recorded for follow-up card
ci_neutrality_targets_local:
  test_template_neutrality_audit: pass
  test_template_no_internal_content_leak_narrow: pass
  test_template_no_internal_content_leak_strict: pass
```

AC matrix summary for the run phase (detail per milestone in §E.2 above):

| AC | Decided at | Verdict |
|---|---|---|
| AC-ALD-001 | M4 (final gate) | PASS — headroom 2757 > 1239 |
| AC-ALD-002 | M1/M2/M3 checkpoints + M4 final | PASS — guard ok, budget_const=1, go_changed=0 |
| AC-ALD-003 ~ AC-ALD-009 | M1 | PASS — see §E.2 M1 matrix |
| AC-ALD-010 ~ AC-ALD-013 | M2/M3 | PASS — see §E.2 M2/M3 matrices |
| AC-ALD-014 | M3 | PASS — see §E.2 M3 matrix |
| AC-ALD-015, AC-ALD-016 | M4 | PASS — 5×MIRRORED, build exit 0; 4-token scan 5×0 + strict leak exit 0 |

Gaps (E8): documentation/mirror milestone — no RED-state test-first artifact exists for any of M1-M4 (stated, not fabricated). Coverage: n/a — 0 Go files changed in the run phase; `internal/config` and `internal/template` packages both green (runs cited above).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-17
run_commit_sha: 67466e997   # run-phase squash on main (PR #1577); per-milestone a203a7c3a/6d701d26c/fcc07d1bc/9269a5e56 inside
sync_commit_sha: pending-backfill-k4v9x   # D3 self-referential-hazard placeholder; backfilled in the follow-up commit on this branch
sync_status: complete
changelog_entry_position: CHANGELOG.md [Unreleased] / Added
ac_count_in_changelog: 16
frontmatter_status_transitions:
  spec_md: in-progress -> completed   # single sync-commit terminal close (3-phase close contract; no separate Mx chore commit)
  plan_md: n/a (markdown-header convention, no frontmatter)
  acceptance_md: n/a (markdown-header convention, no frontmatter)
  progress_md: n/a (this file)
canary_compliance_check:
  spec_lint: pending (sync is markdown-only; spec body untouched)
  changelog_single_entry: grep -c 'SPEC-ALWAYS-LOADED-DIET-001' CHANGELOG.md == 1   # observed 0 pre-emission, 1 post-write
  ac_count_match: 16   # distinct-AC grep on acceptance.md == 16 == count referenced in CHANGELOG entry
  mx_tag_validation: n/a   # zero source files touched by sync; run-phase carried no @MX surface (markdown-only SPEC)
```

### Sync-phase attribution

- **Frontmatter transition carried by this sync commit**: single `in-progress -> completed` merged close on `spec.md` (sole YAML-frontmatter artifact — plan.md / acceptance.md / progress.md use the markdown-header convention, verified by grep before the transition set was claimed). `updated:` already reads 2026-08-17 (run-phase refresh), still current.
- **CHANGELOG emission discipline (B12)**: pre-emission `grep -c 'SPEC-ALWAYS-LOADED-DIET-001' CHANGELOG.md` == 0 (no duplicate from a parallel session); distinct-AC grep on acceptance.md == 16, matching the AC count cited in the CHANGELOG entry; every file path cited in the entry (spec.md + 5 dev-surface rule files + 5 template mirrors) verified via `ls` before commit.
- **README / docs-site: unchanged.** Internal rule architecture — no CLI, no config key, no documented user-facing behavior. The 5 template mirrors distribute rule files to user projects, but those are template-internal artifacts already covered by the M4 neutrality gates, not README/docs-site content. Scope discipline: nothing touched.
- **Residual carried forward (from §E.2 M4)**: the 5 mirrored rule files are not enrolled in a byte-parity registry (`workflowOptMirroredPaths` / `sanitizedPairPaths`) — a Go test-file edit outside this SPEC's scope envelope; recorded as a follow-up card candidate alongside the D4-1 budget ratchet (headroom 2,757 tokens is the ratchet's new baseline input).
- **`sync_commit_sha` backfill**: placeholder `pending-backfill-k4v9x` in this commit; commit 1's real SHA read after landing and substituted in the immediately-following `chore(SPEC-ALWAYS-LOADED-DIET-001): backfill sync_commit_sha` commit (amend is FORBIDDEN per the D3 exemption pattern).

## §F Phase 4 Mode Selection

- decision: `sub-agent` (Mode 5 — sequential per milestone)
- rationale: coding-heavy documentation surgery on one always-loaded rule file per milestone; single-writer discipline protects the AC-ALD-009 byte-sum attribution and the shared §C measurement chain (M1→M4 ordering is load-bearing: M2 adds back bytes M1 removed, M3 adds the recurrence control, M4 mirrors and takes the final measurement)
- autonomous progression: goal armed orchestrator-side (ac_converge family); milestones delegated one at a time
