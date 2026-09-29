# progress.md — SPEC-WORKFLOW-TASKS-001

Card: t1334 — 카드 수행 tasks 도구 진행 표시 의무화

## §F Phase 4 Mode Selection

Decision: serial — Tier S(텍스트·부트스트랩 텍스트·미러 <5파일, 근확실)로 fan-out/agent-team 이득이 없고, always-loaded 교리 파일 편집은 단일 세션 배치가 캐시 경제상 유리.

DP1 (pre-creation approval) note: 운영자 직접 배차 시점에 본 카드 범위에 대한 SPEC 생성이 사전 승인되었으므로 DP1은 사후 감사 후 Kickoff 게이트로 통합 처리됨.

Tier: S (orchestrator auto-resolved; spec.md + plan.md + progress.md — acceptance criteria는 spec.md §3 inline)

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID: SPEC-WORKFLOW-TASKS-001 (regex Bash check PASS, uniqueness 0건 확인)
- Artifacts: spec.md (REQ-TASKS-001..006 + AC-TASKS-001..006, GEARS 5패턴, Out of Scope 3 항목), plan.md (M1-M3, 훅 강제 defer 권고), progress.md (본 파일)
- spec lint: `moai spec lint .moai/specs/SPEC-WORKFLOW-TASKS-001/spec.md` → `✓ No findings — all SPEC documents are valid` (0 error / 0 warning, this run, this tree @ WT-task-progress-ritual)
- Frontmatter: 12 canonical fields 검증 완료 (phase: "v3.2.x target", status: draft)
- Dedup: .moai/specs/ 내 task-tracking/TaskCreate 규율 선점 SPEC 없음 확인 (SPEC-KANBAN-TODO-CLI-001 등은 todo 큐 CLI 영역으로 본 주제와 불중복)
- Evidence path: .moai/specs/SPEC-WORKFLOW-TASKS-001/progress.md
- plan-audit: PASS 0.91 (Tier S threshold 0.75, must-pass 7/7) — .moai/reports/plan-audit/SPEC-WORKFLOW-TASKS-001-review-1.md; 차단 결함 D1(유령 테스트 파일 참조 + §E4 셀렉터 누락)은 감사 후 manager-spec 수리 + 오케스트레이터 재측정 확인(유령 참조 grep 0적중, 실측 골든 포인터 plan.md:21/34/47/61). plan.md 편집으로 아티팩트 해시는 review-1 대비 변경 — run 게이트가 델타 재감사함.
- audit_verdict: PASS (run-gate 재감사 — 해시 미스로 재실행, 점수 0.95로 상승, review-1 결함 7건 전부 해소 독립 확인)
- audit_report: .moai/reports/plan-audit/SPEC-WORKFLOW-TASKS-001-review-2.md
- audit_at: 2026-09-29T11:53:33Z
- auditor_version: plan-auditor (subagent, opus)
- plan_artifact_hash: f55bafb358fa1c3a3a1fcfbae0bc0f2471bbe87d526637d59abd537274719415
- audit_cache: MISS → re-executed (수리 편집이 sticky 판정 무효화; Tier S 단일 패스)
- plan_complete_at: 2026-09-29T20:16:00+09:00
- plan_status: audit-ready

## §E.2 Run-phase Evidence

모든 측정은 this run / this tree — WT-task-progress-ritual @ `2c9ec1c20` (커밋 시점 기준, 달리 명시한 경우 해당 HEAD 표기).

### E8 RED — 골든 테스트 선확장, 주입 텍스트 변경 전 실패 출력 (test-first falsifiability)

Command (env-scrub 단일 복합 호출, worktree 루트에서):

```
go test ./internal/hook/ -run '^(TestSD_AC019_NextCardRuleInjection|TestFactoryGuideTeachesLaneFormsInEveryLocale|TestKanbanCompanionSurfaceCarriesNoLaneRuleSentence)$'
```

Verbatim RED 출력 (발췌, 전체 68행 — 3개 테스트 전부 FAIL, 4 로케일 × 2 토큰):

```
--- FAIL: TestSD_AC019_NextCardRuleInjection (0.00s)
    --- FAIL: TestSD_AC019_NextCardRuleInjection/claude_lane:_every_policy,_source_startup,_four_locales (0.00s)
        session_start_factory_rule_test.go:95: locale en policy clear-each: rule does not carry the tasks-discipline token TaskCreate (SPEC-WORKFLOW-TASKS-001)
        session_start_factory_rule_test.go:95: locale en policy clear-each: rule does not carry the tasks-discipline token TaskUpdate (SPEC-WORKFLOW-TASKS-001)
        session_start_factory_rule_test.go:95: locale en policy clear-when-full: rule does not carry the tasks-discipline token TaskCreate (SPEC-WORKFLOW-TASKS-001)
        session_start_factory_rule_test.go:95: locale en policy clear-when-full: rule does not carry the tasks-discipline token TaskUpdate (SPEC-WORKFLOW-TASKS-001)
        [... ko/ja/zh × clear-when-full/relaunch 동일 패턴, glm lane 2행, gpt owned-card 2행 계속]
--- FAIL: TestFactoryGuideTeachesLaneFormsInEveryLocale (0.00s)
    session_start_factory_worker_test.go:46: en laneNextCardRule lacks the tasks-discipline token TaskCreate: [...]
    [... 4 로케일 × 2 필드 × 2 토큰 계속]
```

RED는 정확히 기대한 이유로 적색이다 — 새 단정이 주입 텍스트의 부존재를 보고한다(implementation 부재 RED). 측정 HEAD: `7bef423c0` (M2 커밋 전 working tree).

### E8 GREEN — §E4 셀렉터 실행 (plan.md:47 셀렉터 그대로)

Command:

```
go test ./internal/hook/ -run '^(TestSD_AC019_NextCardRuleInjection|TestFactoryGuideTeachesLaneFormsInEveryLocale|(TestFactory.*Lane|TestKanban.*Lane))$' -v
```

Verbatim 출력 (발췌):

```
--- PASS: TestSD_AC019_NextCardRuleInjection (0.00s)
    --- PASS: TestSD_AC019_NextCardRuleInjection/claude_lane:_every_policy,_source_startup,_four_locales (0.00s)
    --- PASS: TestSD_AC019_NextCardRuleInjection/glm_lane:_source_startup_carries_the_rule (0.00s)
    --- PASS: TestSD_AC019_NextCardRuleInjection/gpt_lane:_owned-card_rule_names_the_card_and_the_two_CLI_verbs_only (0.00s)
    [... 7개 나머지 하위테스트 전부 PASS]
--- PASS: TestFactoryGuideTeachesLaneFormsInEveryLocale (0.00s)
--- PASS: TestKanbanCompanionSurfaceCarriesNoLaneRuleSentence (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.572s
```

**swept-set 기록**: 최상위 PASS 2개(TestSD_AC019_NextCardRuleInjection·TestFactoryGuideTeachesLaneFormsInEveryLocale, AC-TASKS-005 대상 — 4 로케일 커버는 하위테스트에서) + 신규 characterization 핀 1개. **발견**: plan.md §E4 셀렉터의 와일드카드 팔(`TestFactory.*Lane$` / `TestKanban.*Lane$`, 끝 앵커)은 본 트리의 최상위 테스트 0개와 일치한다 — Lane 이 들어간 테스트명은 전부 뒤 접미가 이어진다(목록: TestFactoryLaneHandoff*, TestKanbanCompanionSurface*, TestKanbanNameChoicesUseLaneNotation 등). 유효 측정 집합은 명시 2개이며 이는 골든 단정의 전부라 충분하나, 셀렉터 와일드카드 팔은 현재 공회전이다(plan.md 본문이라 run-phase 수정 불가 — sync-auditor 참고 바람).

### AC-TASKS-004 — 교리 조항 readback

Command: `grep -n "TaskCreate\|TaskUpdate\|t1330" .claude/rules/moai/workflow/kanban-dispatch.md`

Verbatim 출력:

```
92:[HARD] A lane session tracks its card's execution on the session's task tools. At card intake it registers the card's execution stages via `TaskCreate` before beginning the first stage; at every stage transition it keeps the list current via `TaskUpdate` so the list always shows the stage in flight; and it reports completion only while the list reflects the end state — or carries an explicit annotation naming why it does not. A task list that contradicts its completion report is the same gap as a missing evidence file (§ Completion is read, never trusted).
94:The measured precedent this rule codifies: lane-1 card t1330 held a 7-task list through the card's whole run, one `TaskCreate` per stage at intake and one `TaskUpdate` per transition. The clause above is the rule; that card is its evidence, not an instance list to extend.
```

규칙형 조항 + 실측 사례(t1330) 인용 + 양 토큰 verbatim — REQ-TASKS-004 구조 요건 충족.

### AC-TASKS-006 — make build + 미러 parity

- `make build` → exit 0 (darwin/arm64, `go build ... -o bin/moai ./cmd/moai` 최종행 관측; 이전 시도에서 CLT 링커 PATH 선행 시 `library 'resolv' not found` 재현 — 기본 PATH 사용이 정답)
- parity: `cmp .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` → `IDENTICAL` (변경 전에도 `IDENTICAL`, 편집+cp 후 재확인 `PARITY_IDENTICAL` — 양면 동일 커밋 이동, drift 0)
- `go test ./internal/template/ -run TestTemplateNeutralityAudit` → `ok github.com/modu-ai/moai-adk/internal/template 0.474s`

### E1 AC Binary PASS/FAIL Matrix

| AC | Status | Verification | Actual Output |
|----|--------|--------------|---------------|
| AC-TASKS-001 | PASS-WITH-DEBT | 처방 수준만 본 run에서 관측 가능 — 레인 세션 런타임 컴플라이언스는 다음 레인 카드에서 관측 | 부트스트랩 문장 + 교리 조항 착지(위 GREP/테스트 증거). 기계 강제는 plan.md §G defer 권고 |
| AC-TASKS-002 | PASS-WITH-DEBT | AC-001과 동일 사유 | 동일 |
| AC-TASKS-003 | PASS-WITH-DEBT | AC-001과 동일 사유 — 교리가 § Completion is read, never trusted 와 접속됨을 조항문에서 관측 | 조항 92행 차입 문장 관측 |
| AC-TASKS-004 | PASS | 위 grep readback | TaskCreate·TaskUpdate·t1330 3요소 모두 조항 내 관측 |
| AC-TASKS-005 | PASS | §E4 셀렉터 -v 실행 | `ok ... internal/hook 0.572s`, 4 로케일 하위테스트 전부 PASS |
| AC-TASKS-006 | PASS | cmp parity + make build | `PARITY_IDENTICAL`, build exit 0 |

### 전체 패키지 스위트 + 사전 존재 적색 귀속

Command (레인 env 완전 스크럽 후): `go test -timeout 30m ./internal/hook/` → exit 1, 458.935s, 실패 1건:

```
--- FAIL: TestHMPSourceGuardGoLiterals (0.03s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	458.935s
```

**사전 존재 적색 귀속 (baseline-integrity attribution)**: 이 테스트는 파일 내용 스캔형(격리 실행도 동일 FAIL, 0.412s, env 무관·결정론)이며, 스캔 대상 파일 4개(`contract_sign_guard.go`, `contract_sign_guard_test.go`, `shell_tool.go`, `hmp_source_guard_test.go`)가 base `7bef423c0` 와 byte 동일함을 `git show 7bef423c0:<path>` + `cmp` 로 측정(SAME ×4). 본 SPEC의 diff 는 이 파일들을 건드리지 않는다(§diff stat: 7 파일 전부 본 SPEC 범위). base 트리에서도 동일하게 적색이다 — 카드 t1350 (HMP SourceGuard CI 적색) 소관.

**env 누수 교훈 (재현 기록)**: 첫 전체 스위트 실행은 스크럽 부족(레인 세션 env `MOAI_FACTORY_WORKER=worker-72` 등 잔존)으로 `TestContractRoleScopedAllowWithoutLaneMarker` 도 실패했다. 원인: `contractLaneGate()`(contract_sign_guard.go:144-151)가 `MOAI_FACTORY_WORKER != ""` 조항을 읽고, 테스트는 `EnvFactoryRole` 만 세팅해 나머지 2개 변수는 프로세스 env 를 그대로 읽는다. 완전 스크럽(MOAI_KANBAN* + MOAI_FACTORY* 11개) 재실행에서 이 실패는 소멸 — 코드 결함이 아니라 레인 세션에서 스위트를 돌릴 때의 env-isolated 검증 형식 문제다.

### M2 조건부 발견 — 칸반 표면 (plan.md:60)

`internal/hook/session_start_kanban_i18n.go` 에는 레인 역할 규칙 문장이 존재하지 않는다 — 컴패니언 조인 통지는 역할 없는 한 줄(`companionJoin`) + 영어 전용 `laneSpawnAuthority`(card t224)뿐이다. 레인 세션(kanban 런처의 `lane-N` 포함)의 레인 규칙은 `factoryLaneRuleForSource`(session_start_factory.go:92)가 factory i18n 표에서 단일 공급하며, 세션 시작 542행에서 채널 무관 주입된다. 따라서 kanban i18n 소스는 변경하지 않았고, 이 발견을 기계적 핀으로 고정했다: `TestKanbanCompanionSurfaceCarriesNoLaneRuleSentence`(session_start_kanban_i18n_test.go) — 칸반 표면이 규율을 이중 기재하는 것을 차단(단일 공급처 유지). characteration 핀이므로 RED 대상이 아니다.

### E6 커밋 목록 (push 없음 — 레인 규율)

```
2c9ec1c20 feat(SPEC-WORKFLOW-TASKS-001): M1 lane task-discipline clause in kanban dispatch doctrine + template mirror
6d658ef64 feat(SPEC-WORKFLOW-TASKS-001): M2 lane tasks-discipline sentence in factory lane rules (4 locales)
```

M2 커밋(첫 run-phase 커밋)에서 spec.md frontmatter `status: draft` → `status: in-progress` 전이 수행(본 소관, `updated:` 유지 — 동일 일자).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29T21:40:00+09:00
run_commit_sha: pending-backfill-run
run_status: "complete — 6/6 AC (3 PASS, 3 PASS-WITH-DEBT 처방 수준), 1 사전 존재 적색(t1350 소관) 귀속 완료"
ac_pass_count: 3
ac_fail_count: 0
ac_pass_with_debt_count: 3
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a — 레인 세션, push 소관 없음(2026-09-02 운영자 지시, 리더 일괄)
l44_post_push_fetch: n/a — 동일
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin_arm64: "make build exit 0"
cross_platform_build.windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
cross_platform_build.linux_amd64: not-run (CI가 origin/develop 에서 수행 예정)
total_run_phase_files: 7
m1_to_mN_commit_strategy: "M2(Go+골든테스트+spec 전이, 6d658ef64) → M1(교리+미러 동일 커밋, 2c9ec1c20) → M3(증거, 본 커밋) — always-loaded 교리 편집은 후행 배치(cache-aware directive 3)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29
sync_commit_sha: 203075eca
sync_status: "complete — 3-phase close; CHANGELOG entry added; spec.md status in-progress → completed"
changelog_entry_position: "[Unreleased] / ### Added (1st entry)"
b12_self_test_a: "pre-emission grep -c SPEC-WORKFLOW-TASKS-001 CHANGELOG.md → 0 (exit 1, no match) — proceed"
b12_self_test_b: "AC live-identifier count via spec.md §3 (SSOT) → 6 (AC-TASKS-001..006); CHANGELOG entry cites 6 AC (3 PASS, 3 PASS-WITH-DEBT) — match"
b12_self_test_c: "all cited paths ls-verified — kanban-dispatch.md, mirror, internal/hook/session_start_factory_i18n.go, progress.md, CHANGELOG.md all exist"
frontmatter_status_transitions.in-progress_to_implemented: merged into sync commit (sync-commit transition)
frontmatter_status_transitions.implemented_to_completed: merged into sync commit
canary_compliance_check.public_docs: "no change — see divergence findings"
canary_compliance_check.codemaps: "skipped — see divergence findings"
canary_compliance_check.mx_tags: "no new tags — see divergence findings"
```

### Close summary

- **What synced**: run-phase commits `2c9ec1c20` (M1 doctrine clause + template mirror) + `6d658ef64` (M2 factory i18n 4-locale) + `341292a5d` (M3 evidence) closed with ONE sync commit carrying the CHANGELOG entry, the spec.md `status: in-progress → completed` transition, and this §E.4 signal.
- **CHANGELOG entry added** under `[Unreleased] / ### Added`, after B12 self-tests a/b/c (results in the YAML block above).
- **Divergence findings (verify-don't-assume, this run / this tree @ `341292a5d` + staged sync edits)**:
  - **(a) README / docs-site**: NO public docs surface change required, as expected. Measured: `grep -l TaskCreate README.md README.ko.md README.ja.md README.zh.md` → 0 matched files (the doctrine + hook bootstrap text is internal lane discipline, not a user-facing feature); no docs-site directory in this tree. Approved scope (CHANGELOG + close only) held — no blocker.
  - **(b) codemaps**: SKIPPED, expected. The change surface is doctrine text + i18n string constants + test files — no API signatures, no package structure, no exported symbol changes → no codemap regeneration warranted.
  - **(c) MX tags**: NO new tags, expected. `grep -c "@MX"` across the 4 changed Go files → 0. Text-only i18n string and test-assertion changes add no exported functions, no high fan-in symbols, no dangerous patterns.
- **Note for sync-auditor** (carried from §E.2): plan.md §E4 selector wildcard arms (`TestFactory.*Lane$` / `TestKanban.*Lane$` end-anchored) match 0 top-level tests in this tree — selector wildcards idle; effective measured set is the two explicit golden tests, which cover the full golden assertion. plan.md body is run-phase-immutable, recorded here for the audit trail.
