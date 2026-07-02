# SPEC-HARNESS-EVO-PIPE-REPAIR-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-02
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
author: manager-spec

### plan-auditor 감사 결과 (iteration 1/3, 2026-07-02)

- **Verdict: PASS-WITH-DEBT** · Overall Score: **0.85** (harmonic mean; Clarity 0.85 / Completeness 0.88 / Testability 0.78 / Traceability 0.90)
- 임계: Tier M 선언 0.80 충족. 단, frontmatter `tier:` 필드 부재 → 엄격 규칙상 Tier L 0.85로 감사(경계선 충족) — D4 참조.
- Must-Pass: MP-1 PASS (REQ-HEP-001~013 연속·중복 0, grep 실측) · MP-2 PASS (13 REQ 전건 GEARS When/While/Where/shall/shall-not) · MP-3 PASS (canonical 12필드 전건, rejected alias 0) · MP-4 N/A (다언어 툴체인 열거 없음)
- D7 cross-SPEC: 참조 4 SPEC 전건 `completed`(V4-001/PROPOSAL-GEN-001/CLASSIFIER-WIRING-001/LOOP-CLOSURE-001), retired/superseded 충돌 0; PROPOSAL-GEN-001 부분 supersede는 plan §H에 reconcile 명시 ✓. Epic 미작성 3 SPEC(RUN-REPORT/WRITE-SURFACE/REQ-ARTIFACT) NOT-FOUND는 의도된 forward-ref (SHOULD, 결함 아님).
- D8 syscall: 4 artifact 전건 0 match → auto-PASS.
- 앵커 검증: spec.md §B 인용 앵커 **전량 live 재검증 일치** (types.go:217-245, mapper.go:36/:41-44, learner.go:98-100, hook.go:148-163, frozen_guard.go:21-37, harness.go:486, v4lifecycle.go:32-36, doctor.go:194, Runner :1/:31/:56, harness.md:91/:135/:173/§2.2, learner SKILL:135-144, harness.yaml:119-121, settings grep). 실데이터 검증: usage-log 608 ✓ / promotions 16 ✓ / applied·proposals 부재 ✓ / pattern_key 3형 + confidence 1 + to_tier {observation,auto_update}만 존재 ✓. verification-claim-integrity §F 미검증 항목 명시 우수.
- **Defects (요지)**:
  - **D1 [major/Testability]** AC-HEP-013a의 8-표면 일괄 `diff live↔template` 검증법은 2/8 표면에서 false-fail: harness.yaml은 SPEC 이전부터 대규모 divergence(실측 diff 수십 hunk), settings.json↔settings.json.tmpl은 render-pair(템플릿 변수+bash 이중 항목, tmpl:116/118 vs local:82). "8개 전부 MIRRORED" 주장은 6/8만 byte-identical. → 표면 클래스별 검증(6 byte-diff / settings 토큰-grep / harness.yaml rate_limit 키 수준)으로 AC 재작성 필요.
  - **D2 [major/Clarity+Testability]** EC-4 전제 사실 오류: github/release는 "manifest 있음+Runner 부재"가 아니라 **manifest·Runner 모두 부재**(find 실측: manifest는 release-update/ 1개뿐; v4lifecycle.go:64-66 `ManifestMissing` partial state). REQ-HEP-005 축2(manifest 필수)대로 구현 시 doctor가 github/release 2건 finding → plan §E-7 "수리 후 0 findings" 전역 스모크 false-fail. → EC-4 전제 정정 + command-only thin 하네스에 대한 doctor 정책(skip/INFO vs finding) 정의 필요.
  - D3 [minor/Traceability] spec.md:142 "AC-HEP-001 ~ 015" — AC-HEP-015 비실존(최대 014); acceptance.md DoD-1 "001a ~ 013b"는 역으로 014 누락. 양단 범위 인용 드리프트.
  - D4 [minor/Completeness] frontmatter `tier: M` 필드 부재(부재=Tier L 0.85 감사 규칙); plan/progress 선언과 정합하도록 추가 권고.
  - D5 [minor/Clarity] REQ/AC-HEP-011 "frozen_guard.go와 문서 일치"가 agents/{harness,moai}만 명명 — learner SKILL L1의 `.moai/project/brand/**`(SKILL:142)는 frozenPrefixes에도 allowedPrefixes에도 없어 처분 미정의(naive 전체 목록 재작성 시 오정렬 위험).
  - D6 [minor/Testability] AC-HEP-004 Given "make build 후 렌더된 settings" — 로컬 .claude/settings.json은 make build 산출물 아님(직접 편집/moai init 렌더); AC-HEP-008 presence-grep은 파일 전역 산문과 충돌해 "(Phase 0 절)" 스코핑이 수동.
- 정보성: (i) 기존 hook_harness_observe_stop_test.go가 현행 Stop 핸들러 동작 고정 — M2 classify 연쇄는 fail-open으로 그린 유지 or 의식적 갱신 필요; (ii) template harness.md:91/:253에 REQ-HRN-FND-018 토큰 기존재(중립성 CI 허용 상태) — REQ-HEP-013 grep 가드가 REQ-HEP로 스코프된 것은 정확; (iii) `related_specs`는 비표준 optional 필드(표준은 depends_on, decoder 허용).
- 재감사 조건: D1·D2는 acceptance.md/plan.md 소폭 수정으로 해소 가능 — 수정 반영 시 iteration 2 불요(PASS-with-debt 승인, run-phase 진입 전 D1/D2 정정 의무).

## §E.2 Run-phase Evidence

run-phase 실행자: manager-develop (cycle_type=tdd) · Route A(Hybrid Trunk main-direct) · worktree 격리(agent-a74fcf2669c9458e4) → `git push origin HEAD:main`.

### 20-AC 이진 매트릭스 (전건 PASS)

| AC | Status | 검증 명령 | 관측 출력(요지) |
|----|--------|-----------|------------------|
| AC-HEP-001a | PASS | `go test -run TestMapper_RealDataAutoUpdate\|TestMapper_CurrentDataProducesCandidates ./internal/harness/proposalgen/` | `ok ... proposalgen` — user_prompt::(auto_update)+moai_subcommand(rule) 2 candidates; baseline fixture 3 candidates |
| AC-HEP-001b | PASS | `go test -run TestMapper_PreActionableExcluded ./internal/harness/proposalgen/` | observation/heuristic/recommendation/approval_required → 0; rule/auto_update → 1 |
| AC-HEP-002a | PASS | `go test -run TestMapper_RealDataSchemaPass ./internal/harness/proposalgen/` | `user_prompt::` / `agent_invocation:Bash:` / `session_stop::` 빈-segment 채택; apply_outcome/code_change 거부 |
| AC-HEP-002b | PASS | `grep -c "code_change\|error_pattern\|tool_failure\|repeated_edit" internal/harness/proposalgen/mapper.go` | `0` (EventType enum 파생, 수기 목록 제거) |
| AC-HEP-003a | PASS | `go test -run TestRunHarnessObserveStop_AutoClassifyChain ./internal/cli/` | Stop 경로 → tier-promotions.jsonl 생성(classify 연쇄) |
| AC-HEP-003b | PASS | `go test -run TestRunHarnessObserveStop_ClassifyFailOpen ./internal/cli/` | usage-log 차단 시 훅 exit 0(fail-open), tier-promotions 미생성 |
| AC-HEP-004 | PASS | `grep -c "handle-harness-observe.sh" internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` | tmpl=2(windows/비-windows), local=1(단일) — PostToolUse 등록 |
| AC-HEP-005a | PASS | `go test -run TestDoctor_ValidHarness_Passes ./internal/cli/harness/` | valid 하네스 → 0 ERROR |
| AC-HEP-005b | PASS | `go test -run TestDoctor_ZeroHarness_Graceful ./internal/cli/harness/` / `moai harness doctor` (0-harness) | 0 harness → exit 0, "No v4 harnesses found" |
| AC-HEP-006 | PASS | `go test -run TestDoctor_DefectClass_Detected ./internal/cli/harness/` | B5 결함 fixture → runner+agent 축 ERROR ≥2 |
| AC-HEP-007a | PASS | `moai harness doctor --project-root <repo>` (수리 전) | release-update 2 ERROR(MANIFEST_PATH `.claude/commands/harness/manifest.json` + agent `harness-devkit-release-update-specialist`), github/release INFO, exit 1 |
| AC-HEP-007b | PASS | `moai harness doctor --project-root <repo>` (수리 후) + `git status internal/template/` | 0 ERROR / 2 INFO / exit 0; template 청정 |
| AC-HEP-008 | PASS | `awk '/^## Phase 0/,/^## Phase 1/' harness-build-entry.md \| grep -o "status\|apply\|rollback\|disable\|list\|edit\|remove\|doctor" \| sort -u` | 8 verb 전건: apply disable doctor edit list remove rollback status |
| AC-HEP-009 | PASS | commands/moai/harness.md argument-hint + SKILL.md :70·§harness grep | list/edit/remove/doctor 열거 |
| AC-HEP-010 | PASS | `grep -c "\.claude/harness/" harness-builder.md` | `0` (OR-분기 제거, 단일 정본) |
| AC-HEP-011 | PASS | harness.md Layer1 + learner L1 grep | `.claude/agents/harness/` FROZEN 제거(allowed-write), `.claude/agents/moai/` FROZEN 유지, brand/** 유지 |
| AC-HEP-012 | PASS | `grep -c "floor of 1\|1 per 7-day\|per 7-day window" harness.md` | `0`; harness.yaml rate_limit(3/week·24h) 기준 + REQ-HRN-FND-018 supersede provenance |
| AC-HEP-013a | PASS | class-A 6 `diff live↔template` / class-B settings 토큰 grep / class-C harness.yaml rate_limit 키 | 6 byte-identical; settings 토큰 존재; rate_limit max_per_week:3·cooldown_hours:24 양측 일치; make build 성공 |
| AC-HEP-013b | PASS | `grep -rn "HARNESS-EVO-PIPE-REPAIR\|REQ-HEP" internal/template/templates/` + `TestSplitHarnessNamespaceNoLeak` | 0 matches + PASS (+ TestTemplateNoInternalContentLeak PASS) |
| AC-HEP-014 | PASS | `grep -c -i "doctor\|smoke" harness-builder.md` | ACTIVATE 절에 스모크 게이트 실행 + 결함 시 활성 선언 금지 계약 존재(2 matches) |

### DoD 증거

- **DoD-2 (real-data 스모크)**: `moai harness propose --input <repo>/.moai/harness/learning-history/tier-promotions.jsonl --dry-run` → `proposals: 3 | reason: ok | evaluated: 4` (session_stop:: / subagent_stop:unknown: / user_prompt:: 모두 auto_update). 수리 전 동일 데이터는 `reason: no-actionable-patterns` / 0 candidates(구조적 0). **해소 실증**.
- **DoD-3 (exemplar 게이트 회귀)**: 수리 전 `moai harness doctor` → release-update `[ERROR](runner) MANIFEST_PATH ".claude/commands/harness/manifest.json" does not resolve` + `[ERROR](agent) references "harness-devkit-release-update-specialist" ... does not exist` (2 ERROR); 수리 후 → `Scanned 3 harness(es): 0 ERROR, 2 INFO` exit 0.
- **DoD-4 (template parity+중립성)**: class-A 6표면 byte-identical, class-C harness.yaml rate_limit 키 일치, neutrality grep 0, TestSplitHarnessNamespaceNoLeak + TestTemplateNoInternalContentLeak PASS.

### classify 5s 예산 실측 (EC-5)

`moai hook harness-observe-stop` on 608-line usage-log(격리 tempdir, HOI+learning enabled) → `real 0.53` wall (프로세스 기동 지배; classify 자체 sub-ms) → 5s 예산 내. `auto-classify 2 patterns → 2 promotions` 관측.

### 커버리지 / 빌드

- `internal/harness/proposalgen`: coverage **90.6%** (≥85% ✓)
- `internal/cli/harness/doctor.go`: Doctor 91.7% / checkHarness 86.7% / NewHarnessDoctorCmd 84.0% (파일 집계 ≥85%)
- cross-platform: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- `go test ./...` exit 0 (90 pkg ok) · `go test -race` touched pkg PASS · `golangci-lint run` 0 issues

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-02
run_commit_sha: "<M6 progress 커밋 — 본 §E.2/§E.3 기록 커밋; M1..M5 = 2ef2c6a8b/3c2d1ef11/3a91c80ce/143b816a9/48e226b03>"
run_status: complete
ac_pass_count: 20
ac_fail_count: 0
preserve_list_post_run_count: 0   # frozen_guard.go / Tier.String 방출부 / reader tolerance / observe wrapper 3종 / runtime-managed 무변경
l44_pre_commit_fetch: "0 0 (매 커밋 전 git fetch origin main + rev-list; 병렬 세션 disjoint 스코프)"
l44_post_push_fetch: "origin/main == HEAD 매 push 후 확인 (2ef2c6a8b/3c2d1ef11/3a91c80ce/143b816a9/48e226b03 순차 fast-forward)"
new_warnings_or_lints_introduced: 0   # golangci-lint run 0 issues on touched pkgs
cross_platform_build:
  linux_darwin: exit 0
  windows_amd64: exit 0
total_run_phase_files: 22   # Go src 5(mapper/types/hook/doctor/harness_route) + Go test 4(mapper/classify-chain/doctor/propose×2) + settings 2 + doc live 6 + doc template 6 + catalog 1 + exemplar Runner 1 + SPEC frontmatter 3(spec/plan/acceptance status) + progress 1
m1_to_mN_commit_strategy: "M1(fix Go 어휘/스키마) → M2(feat classify+observer) → M3(feat doctor) → M4(fix exemplar local) → M5(docs dispatcher/mirror) → M6(progress §E.2/§E.3); pathspec 제한 커밋, Route A main-direct push"
subagent_boundary_grep: "0 matches in modified files (mapper/types/hook/doctor/harness_route); 잔여 match는 pre-existing 벤치(scaffolder 문자열/agent_lint 탐지코드/cobra doc-string/testdata fixture)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-03
sync_commit_sha: "58ec3a2ea2c641fff8bba8b6a5c22f9785e2c08f"
sync_status: ready
changelog_entry_position: "[Unreleased] HARNESS-EVO-PIPE-REPAIR-001 추가 (20 AC PASS, 0 FAIL)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed"
  plan_md: "in-progress → completed"
  acceptance_md: "in-progress → completed"
  updated_field: "2026-07-02 → 2026-07-03 (current date)"
canary_compliance_check:
  template_mirror_parity: "8 surfaces — spec.md/plan.md/acceptance.md live+template; harness.md / harness-build-entry.md / harness-builder.md / moai SKILL.md / commands/moai/harness.md / learner SKILL.md (8/8 mirrors exist, byte-match or render-pair per AC-HEP-013a class taxonomy)"
  template_neutrality: "TestSplitHarnessNamespaceNoLeak + TestTemplateNoInternalContentLeak PASS (0 SPEC-ID / 0 REQ-HEP token / 0 audit citation in template-managed surfaces)"
  changelog_ac_count: "20 AC rows in acceptance.md (grep -c 'AC-HEP-' = 20 ✓)"
b12_self_test_a_pre_emission_grep: "grep -c 'HARNESS-EVO-PIPE-REPAIR' CHANGELOG.md = 0 (pre-emission ✓ no duplicate)"
b12_self_test_b_file_path_validation: "9 key files exist (mapper.go / mapper_test.go / types.go / hook.go / doctor.go / runner.js / settings.json / settings.json.tmpl / harness.yaml)"
b12_self_test_c_ac_count_match: "20 AC rows in acceptance.md; CHANGELOG entry references 20-AC PASS (match ✓)"
sync_auditor_placeholder: "[orchestrator will invoke sync-auditor after sync commit lands; pending external audit]"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 0.95 Mode Selection

Mode Selection (orchestrator-direct, 2026-07-02, run-phase 진입 직전 기록):

- Input parameters: tier=M · scope≈10-15 files(internal/harness + internal/cli/hook·harness Go + settings.json(.tmpl) + 8 doc surface) · domain count≈4(Go source / hook wiring / settings / doc-template mirror) · file mix=Go+shell/json+markdown 혼합 · concurrency benefit=LOW(coding-heavy TDD, inter-milestone 의존 M2←M1·M4←M3·M6←전체) · Agent Teams prereqs=미충족(solo 세션, thorough·team.enabled·env 3조건 불성립)
- Mode evaluation:
  - Mode 1 trivial → not selected(다중 마일스톤 semantic 변경)
  - Mode 2 background → not selected(Write 작업, run_in_background:false 필수)
  - Mode 3 agent-team → not selected(capability-gate 3조건 미충족)
  - Mode 4 parallel → not selected(coding-heavy, Anthropic coding-task parallelism caveat)
  - Mode 6 workflow → not selected(≥30-file mechanical uniform transform 아님; ultracode 활성이나 coding-task caveat이 Mode 6를 배제)
  - **Mode 5 sub-agent → SELECTED**
- Decision: sub-agent
- Justification: coding-heavy TDD 신규/수리 코드 + 마일스톤 간 강한 의존(M2 RED은 M1 GREEN 어휘에 의존, M4는 M3 doctor 사용, M6는 전체 검증). Anthropic coding-task parallelism caveat에 따라 순차 sub-agent(manager-develop cycle_type=tdd)가 정답. ultracode 활성 상태이나 Mode 6(Workflow fan-out)은 §E anti-pattern("Workflow for coding-heavy/new-code")에 해당하여 배제.

## §G IGGDA Kickoff Predicate

Implementation Kickoff Approval 판정 (orchestrator-direct, 2026-07-02):

- (a) Intent clarity 100%: PASS — plan-phase AFK 권장옵션 채택(Epic 4-SPEC 구조 + cleanup 4건) + 본 세션 사용자 명시 승인 "착수"
- (b) plan-auditor PASS: PASS — verdict PASS-WITH-DEBT 0.85(Tier M 임계 0.80 충족); D1~D6 v0.1.1 반영(auditor 재감사 불요 계약 충족)
- (c) Tier S/M: PASS — Tier M(NOT L)
- (d) No dangerous keyword AND no destructive scope: **FAIL** — 위험 도메인 키워드 `session` 매칭(Stop-hook 맥락, §H.3 over-inclusive); `--pr` 부재·비파괴 스코프이나 (d)는 키워드 단독 FAIL
- Verdict: **explicit-gate**(조건 d FAIL) → mandatory blocking AskUserQuestion 발행 → 사용자 응답 "착수" 수신 → run-phase 인가
- Pre-spawn 관측: git divergence 0 0(동기화) · 세션 레지스트리 [] · working tree에 병렬 세션 live 신호(disjoint 스코프 — docs/config/release) → pathspec 제한 커밋 규율 하 흡수 진행
