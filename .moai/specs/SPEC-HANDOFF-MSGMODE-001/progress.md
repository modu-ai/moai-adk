# progress — SPEC-HANDOFF-MSGMODE-001 (message-v2)

Epic "Handoff-v2" M2/4. Tier S · doctrine-only. 선행 M1 = SPEC-HANDOFF-CTXGUIDE-001 (completed, origin `60db8e721`).

| Phase | Status | Signal |
|-------|--------|--------|
| plan  | completed | spec.md + plan.md + progress.md authoring (plan-auditor PASS 0.86) |
| run   | completed | doctrine 4-surface 편집 완료 (SSOT→render→mirror×2); AC-001..016 전부 PASS; frontmatter in-progress→completed |
| sync  | completed | spec/plan/progress frontmatter transitions + CHANGELOG [Unreleased] entry + 3-phase close |

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-04
tier: S
plan_auditor_threshold: 0.75
artifacts:
  - spec.md          # §1 배경·목표·범위(§1.3 Exclusions h3) + §2 REQ-MSGMODE-001..013 + §3 AC-MSGMODE-001..016 인라인
  - plan.md          # §A.1 Tier S + §A.2 4-surface 타깃 + §A.3 순서 + §A.4 PRESERVE + §A.5 R1..R5 + §A.6 자가검증 + §A.7 Out of Scope(h3)
  - progress.md      # 본 파일
req_count: 13        # REQ-MSGMODE-001..013
ac_count: 16         # AC-MSGMODE-001..016 (014 v1 byte-identity, 015 spec-lint, 016 parity sentinel)
scope: doctrine-only # Go/config/state 변경 0
target_surfaces:
  - .claude/rules/moai/workflow/session-handoff.md            # live SSOT
  - .claude/output-styles/moai/moai.md                        # live render (§8)
  - internal/template/templates/.claude/rules/moai/workflow/session-handoff.md  # mirror
  - internal/template/templates/.claude/output-styles/moai/moai.md              # mirror
open_decision:
  - "B3 9-vs-10 self-check 조율: template-completeness self-check(10)+mode(1)=11로 통일; paste-ready budget(9)는 별개 concern 유지 — plan.md §A.5 R1"
```

## §E.2 Run-phase Evidence

**4-surface 편집 요약** (SSOT→render→mirror×2, 각 표면 3편집 = 12편집 + AC-005/AC-010 정정 2×2):

- **SH** (`.claude/rules/moai/workflow/session-handoff.md`, SSOT): Canonical Format 스켈레톤 `# /effort ultracode` → `mode:` 라인; Field-by-Field Block 1에 line-order invariant + `mode:` orchestration-seed 소절(4-enum↔Mode 매핑표 + Excluded modes + Threshold reuse + SEED-not-permission + Directive binding + emit-discouraged/parse-accept + protocol-token + JSON-twin note) + ultracode bare/slash 재정합 bullet + `/goal` placement 갱신; anti-pattern L184 bare-ultracode 재정합.
- **M8** (`.claude/output-styles/moai/moai.md §8`, render): 스켈레톤 `# /effort ultracode` → `mode:`; compact mode-mapping 소절 신설; template-completeness self-check 10→11 items(+`mode:` 항목, `10 items` 잔여 0).
- **SH-mir / M8-mir** (템플릿 미러): SH/M8과 동일 편집, §25 neutrality(SPEC-ID·날짜·SHA·REQ 토큰 0).

**AC-MSGMODE-001..016 PASS/FAIL 행렬** (실제 grep 출력은 run-phase self-verification §E1 참조):

| AC | Verdict | Evidence (command → key output) |
|----|---------|---------------------------------|
| 001 | PASS | 스켈레톤 `mode:` 라인 + "OMIT for solo-sequential (default) → v1 byte-identical" 존재 |
| 002 | PASS | 4 enum + `Mode 3/4/5/6` SH(전체표) & M8(compact) 양쪽 존재 |
| 003 | PASS | `Mode 1 (trivial)`/`Mode 2 (background)` "NOT handoff-relevant" 제외 서술 |
| 004 | PASS | `domains ≥ 3 / files ≥ 10 / score ≥ 7` 재사용 + "introduces NO new threshold" |
| 005 | PASS | Directive binding 4규칙: ultrathink(항상)/bare ultracode(iff dynamic-workflow)/`/goal`(iff run+verifiable)/`--team`(iff agent-team) |
| 006 | PASS | bare `ultracode` opener vs `/effort ultracode` session-persistence 구분; 스켈레톤 opener-default `# /effort ultracode` 잔여 0(SH,M8) |
| 007 | PASS | "SEED"·"not a permission grant"·"Implementation Kickoff Approval"·"does NOT authorize autonomous run-phase entry" 존재 |
| 008 | PASS | `schema_version: 2` + "no JSON twin currently" forward-compat note |
| 009 | PASS | locale 표 컬럼 4(en/ko/ja/zh) 유지(신규 행 0); "locale-verbatim protocol token" SH & M8 |
| 010 | PASS | "immediately after `ultrathink.` (or after the `mode:` line when present)" 존재; `/goal` ordering `opener → mode: → # /goal` 참조; 잔여 미확장 clause 0 |
| 011 | PASS | "emit-discouraged" + "parse-accept" SH & M8; `grep -c 'never emitted'` = 0/0 (forward-authoring guard 유지) |
| 012 | PASS | `grep -c '11 items' M8`=1 + `mode:` 항목 + M8 `10 items`=0; `grep -c '9 items' SH`=1(불변); SH `11 items`=0 |
| 013 | PASS | mirror parity `diff` IDENTICAL(SH==SH-mir, M8==M8-mir); neutrality grep=0 |
| 014 | PASS | `grep -cE 'byte-identical|zero-diff|v1' SH` = 5 |
| 015 | PASS | `moai spec lint <spec.md>` & `<plan.md>` 각 "No findings" exit 0 (파일 경로 형식) |
| 016 | PASS | concern-name qualifier 3종 + "en / ko / ja / zh — 4 columns" SH sentinel & M8 sentinel 상호 일치 |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-05
run_commit_sha: "<this run commit — backfill in sync if needed>"
tier: S
scope: doctrine-only        # Go/config/state 변경 0 — go build 미적용
ac_pass_count: 16           # AC-MSGMODE-001..016 전부 PASS
ac_fail_count: 0
target_surfaces_edited: 4   # SH + M8 + SH-mir + M8-mir
mirror_parity: identical    # diff SH==SH-mir, M8==M8-mir → IDENTICAL
mirror_neutrality_grep: 0   # SPEC-HANDOFF-MSGMODE|SPEC-MSGMODE|2026-07-0[45] = 0 (SH-mir, M8-mir)
never_emitted_grep: 0       # B2 forward-authoring guard 유지 (SH,M8)
self_check_counts:
  sh_paste_ready_budget: "9 items (불변)"
  m8_template_completeness: "11 items (10→11, +mode:)"
spec_lint: "No findings (exit 0) — spec.md AND plan.md, 파일 경로 형식"
new_spec_lint_errors: 0
v1_byte_identity: preserved  # solo-sequential 라인 생략 = 공통 케이스 zero-diff
new_warnings_or_lints_introduced: 0
m1_to_mN_commit_strategy: "single consolidated run commit (plan-phase commit deferred; artifacts + implementation 동시 landing)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-07-05
sync_commit_sha: ""
tier: S
scope: doctrine-only        # Go/config/state 변경 0
frontmatter_status_transitions:
  spec.md: "in-progress → completed"
  plan.md: "in-progress → completed"
changelog_entry_position: "[Unreleased] → Changed section"
artifacts_updated:
  - spec.md                 # status: in-progress → completed; updated: 2026-07-05
  - plan.md                 # status: in-progress → completed; updated: 2026-07-05
  - progress.md             # §E.4 this signal populated
  - CHANGELOG.md            # [Unreleased]/Changed 항목 추가
b12_self_test:
  pre_append_grep_count: 0  # dedup gate PASS (no prior entry)
  ac_count_match: 16        # acceptance.md / spec.md §3 AC-MSGMODE-001..016 인라인
  file_path_verification:
    - "ls .moai/specs/SPEC-HANDOFF-MSGMODE-001/spec.md ✓"
    - "ls .moai/specs/SPEC-HANDOFF-MSGMODE-001/plan.md ✓"
    - "ls .moai/specs/SPEC-HANDOFF-MSGMODE-001/progress.md ✓"
    - "ls CHANGELOG.md ✓"
sync_auditor_gate_4d:
  functionality: PASS      # Tier S doctrine-only, 4-surface docstring sync completes without content drift
  security: PASS           # No secrets, no auth changes (doctrine-only)
  craft: PASS              # GEARS lint, neutrality grep 0, parity SH==SH-mir, M8==M8-mir
  consistency: PASS        # frontmatter epoch aligned 2026-07-05, AC-count matched, CHANGELOG format compliant
```
