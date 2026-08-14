# SPEC-AGENT-MODEL-ENFORCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-08
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M 표준 — 편차 없음)
issue: #1376
open_clarifications: 2 (plan.md §B D4 스텁 배치 위치, D5 감사 로그 형식)
blocking_gate: M1 페이로드 실측 — PreToolUse가 Agent/Task에 발화하는지 미검증

## §E.2 Run-phase Evidence

### M1 — 페이로드 실측 결과 (REQ-AME-002 3문항)

측정 방법: 격리된 임시 프로젝트(`scratchpad/m1probe`)에 PreToolUse `Agent|Task` matcher 블록 +
stdin 덤프 훅을 배선하고, `claude -p`(Claude Code 2.1.226) 헤드리스 세션에서 실제 Agent spawn을
발생시켜 페이로드를 포획. `Bash` matcher 블록을 **양성 대조군**으로 함께 배선해
"훅 미발화"와 "프로브 고장"을 구분함.

| # | 질문 | 관측된 답 | 근거 |
|---|------|-----------|------|
| (a) | **PreToolUse가 Agent/Task에 발화하는가** | **YES** | 포획 레코드 `hook_event_name: "PreToolUse"`, `tool_name: "Agent"`. 양성 대조군(Bash) 3건도 동시 포획되어 프로브 정상 동작 확인 |
| (b) | **`tool_input`이 `subagent_type`을 담는가** | **YES** | `tool_input.subagent_type == "Explore"`. 픽스처: `internal/hook/testdata/agent_pretool_payload.json` |
| (c) | **`tool_input`이 `model` 키를 담을 수 있는가** | **YES** | 2차 프로브에서 model 인자를 명시한 spawn 유도 → `tool_input.model == "haiku"` 실제 포획. 픽스처: `internal/hook/testdata/agent_pretool_payload_with_model.json` |

판정: **M1 게이트 통과**. REQ-AME-003의 재라우팅 분기(PreToolUse 미발화 시 M2 이후 중단)는
발동하지 않음. M2-M4는 관측된 능력 위에 구축됨.

관측된 `tool_input` 키 집합 (model 부재 spawn): `{description, prompt, run_in_background, subagent_type}`
관측된 `tool_input` 키 집합 (model 보유 spawn): `{description, model, prompt, run_in_background, subagent_type}`

부수 관측(범위 밖, 기록만): PreToolUse 페이로드는 최상위에 `effort` 객체(`{"level":"medium"}`)를
담으며, 서브에이전트 내부에서 발생한 도구 호출은 최상위 `agent_id` / `agent_type`을 추가로 담는다.
effort는 §A.4에 따라 본 SPEC의 집행 대상이 아니며, 이 관측은 후속 SPEC의 입력 자료로만 남긴다.

### F4 독립 교차 검증 (PostToolUse Agent 분기 도달 불가)

`.moai/logs/task-metrics.jsonl` 3409행 전수 스캔 결과 UUID 형태 session_id **0건**
(전량 `sess-metrics-*` 등 테스트 픽스처 id). 즉 `logTaskMetrics`는 실세션에서 한 번도
기록한 적이 없으며, spec.md §A.1 F4의 "배선상 도달 불가" 판정이 독립 증거로 재확인됨.

### 순환 임포트 사전 점검 (plan.md §C 6번 — 차단 조건)

```
go list -deps ./internal/template/... | grep 'moai-adk/internal/hook'   → 출력 없음
```
역방향 의존 **부재 확인**. `internal/hook → internal/template` 직접 임포트가 안전하므로
REQ-AME-012의 해석기 직접 호출을 그대로 채택(제3 패키지 승격/인터페이스 주입 불필요).

### 종단 실측 (§E DoD 6번 — 게이트 기본값에서 spawn 미차단)

새 바이너리를 PATH 선두에 두고 워크트리에서 실세션 Agent spawn 1회 발생:

```
audit rows before: 0
(Explore 서브에이전트가 정상 실행되어 결과 반환 — 차단 없음)
audit rows after: 1
{"timestamp":"2026-08-08T07:04:56Z","session_id":"84091ded-2e06-4f9e-8c9c-297999d253d8",
 "agent":"Explore","declared_model":"","resolved_model":"sonnet","verdict":"missing"}
```

관측: (1) 게이트 기본값 `false`에서 spawn이 **차단되지 않음**, (2) 감사 로그에 정확히 1행 기록,
(3) 판정이 `missing`으로 정확 — 실세션 spawn이 model 인자를 담지 않았고 해석값은 `sonnet`,
(4) prompt 본문 미유출, (5) 실제 UUID session_id. 배선 전 구간이 실동작으로 확인됨.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-08
run_commit_sha: pending-backfill-run
run_status: PASS-WITH-DEBT
ac_pass_count: 28
ac_fail_count: 1          # AC-AME-053 (커버리지) — 사전 존재 격차, 아래 참조
ac_debt_count: 2          # AC-AME-041 / 042 — D4(b) 결정에 따른 측정 대상 변경
preserve_list_post_run_count: 6   # plan.md §A.3 PRESERVE 6항목 전부 무변경
new_warnings_or_lints_introduced: 0
total_run_phase_files: 12
m1_to_mN_commit_strategy: 3 commits (M1 / M2-M4 / M5)
```

### 미해결 항목 (정직 보고)

**AC-AME-053 — FAIL (사전 존재 격차, 본 SPEC 범위 밖)**
`internal/hook` 패키지 커버리지 목표 90%, 실측 **84.1%**. 단 이는 본 SPEC이 만든 격차가
아니다 — HEAD(4fc632280) 분리 워크트리 실측 baseline이 **83.9%** 이고 본 변경이 **+0.2pp**
올렸다. 신규 코드 자체는 8개 함수 중 7개 100%, `appendAgentModelAudit` 76.2%(도달 불가한
marshal 실패 분기 제외 시 실질 전량). 90% 달성은 `internal/hook` 176개 파일 전반의 테스트
보강이 필요하며 본 SPEC의 scope 봉투 밖이다.

**AC-AME-041 / AC-AME-042 — PASS-WITH-DEBT (측정 대상 불일치)**
두 AC는 스텁이 **독립 파일**로 신설된다는 전제로 `wc -c <스텁 경로>` / `grep -cE 'opus|...'
<스텁 경로>`를 지정한다. 오케스트레이터의 D4 확정은 (b) — 이미 항상 로드되는
`agent-common-protocol.md`의 한 절로 삽입 — 이므로 명령을 호스트 파일에 그대로 적용하면
36,585 bytes(> 2048)이고 alias 단어도 1건(기존 super-advisor 절의 "Opus/Sonnet") 잡힌다.
**추가된 절 자체**로 측정하면 요구사항 의도(REQ-AME-041/042)를 충족한다:

| 측정 | 값 | 기준 |
|---|---|---|
| 추가 절 크기 | 1,595 bytes | ≤ 2048 |
| 절 내 alias 리터럴 | 0건 | 0 |
| `model-policy.md` 교차 참조 | 1건 | ≥ 1 |
| `model` 언급 | 13건 | ≥ 1 |
| `model-policy.md` `paths:` 스코프 | 유지 | REQ-AME-043 |

acceptance.md 본문 수정은 manager-develop 소관 밖(수정 금지)이므로 여기 기록만 남긴다.
AC 명령의 측정 대상을 "추가된 절"로 정정하는 것은 manager-spec 재위임 사안이다.

**AC-AME-054 — 전체 수트 3건 실패 (전부 사전 존재)**
`TestLateBranchTemplateMirror/spec-assembly.md`, `TestRuleTemplateMirrorDrift/spec-workflow.md`,
`TestSanitizedPairParity/main-checkout-branch-guard.md`. run-phase 착수 전 baseline과 동일하며
본 변경이 추가한 4번째 실패는 **없다**(agent-common-protocol.md 미러를 동일 커밋에 반영).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-14
sync_commit_sha: 144573336d07da19f4b8a50aa26c38db2704afb5
sync_status: audit-ready
run_status_carried_forward: PASS-WITH-DEBT   # §E.3 verdict carried verbatim — NOT laundered into a clean PASS
b12_self_test_a_changelog_duplicate_check: PASS (grep -c 'SPEC-AGENT-MODEL-ENFORCE-001' CHANGELOG.md → 0 before emission)
b12_self_test_b_ac_count_match: PASS (31 distinct AC in acceptance.md = 28 PASS + 1 FAIL + 2 PASS-WITH-DEBT per §E.3)
b12_self_test_c_file_path_verification: PASS (3/3 SPEC paths cited in the CHANGELOG entry verified via ls before commit)
changelog_entry_position: "[Unreleased] ### Changed (lifecycle-hygiene entry, covers this SPEC + 2 siblings)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (single sync commit; updated: → 2026-08-14)"
  plan_md: "n/a (markdown-header convention)"
  acceptance_md: "n/a (markdown-header convention)"
  progress_md: "§E.4 populated"
canary_compliance_check:
  implementation_provenance: "landed at b0d3b61f8 (PR #1410, issue #1376) — NOT in this sync commit; this sync commit is documentation-only (SPEC lifecycle close)"
  go_build: PASS (measured in this worktree at 7f61332ef — post-merge verification, not attributable to this sync commit)
  cross_platform_build: PASS (GOOS=windows/amd64 exit 0, GOOS=darwin/arm64 exit 0 — measured in this worktree)
  coverage: "84.0% internal/hook measured in this worktree at 7f61332ef, against the SPEC's 90% bar — still FAIL, and still the pre-existing gap §E.3 named (84.1% at run-phase HEAD 4fc632280; the 0.1pp delta is subsequent-commit drift, not a regression from this SPEC)"
  ac_pass: "28/31 PASS + 1 FAIL + 2 PASS-WITH-DEBT per §E.3 — carried forward, NOT re-executed in sync-phase"
  implementation_surface_verified:
    - "internal/hook/agent_model_guard.go + agent_model_guard_test.go present (ls)"
    - "internal/config/types.go:404 AgentModelGuard field + :574 AgentModelGuardConfig struct (grep)"
    - "internal/config/defaults.go:683-685 AgentModelGuard{Enabled: false} default block (grep)"
    - "internal/hook/pre_tool.go:552 h.checkAgentModel(input) call site (grep)"
open_debt_carried_into_close:
  - id: AC-AME-053
    disposition: FAIL
    summary: "internal/hook coverage 84.0% against a 90% bar. Pre-existing gap outside this SPEC's scope envelope — run-phase baseline 83.9%, this SPEC moved it +0.2pp. Closing this SPEC does NOT close this gap; it remains open against internal/hook at large."
  - id: AC-AME-041
    disposition: PASS-WITH-DEBT
    summary: "Measurement-target mismatch. The AC command targets a standalone stub file; orchestrator decision D4(b) instead inserted the content as a section of the always-loaded agent-common-protocol.md. Measured against the added section: 1,595 bytes ≤ 2048 (requirement intent met). Correcting the AC's measurement target is a manager-spec re-delegation, not a manager-docs edit."
  - id: AC-AME-042
    disposition: PASS-WITH-DEBT
    summary: "Same D4(b) measurement-target mismatch. Measured against the added section: 0 alias literals (bar: 0), 1 model-policy.md cross-reference (bar: ≥1), 13 'model' mentions (bar: ≥1). Requirement intent met; the AC text still points at the superseded stub-file target."
debt_closure_note: "This SPEC closes as completed with the three items above OPEN and named. The close records that run-phase finished with declared debt; it does not assert the debt was resolved. AC-AME-053 is an internal/hook-wide coverage gap; AC-AME-041/042 need a manager-spec amendment to re-point their measurement commands."
mx_tag_validation: "no @MX tag changes — this sync commit touches only .moai/specs/ and CHANGELOG.md"
```
