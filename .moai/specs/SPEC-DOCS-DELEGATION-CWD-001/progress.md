# progress.md — SPEC-DOCS-DELEGATION-CWD-001

## §A 현재 상태

- 카드: t1387 · SPEC 상태: `draft` (plan-phase 산출물 방출)
- 트리: `.moai/worktrees/t1387` · 브랜치: `WT-docs-delegation-cwd` · 작성 시점 HEAD: `e927266be`
- Tier M 산출물 4종 + decision-index 작성 완료 (2026-10-01, manager-spec)

## §B 결정 요약

- 수리 부류 확정: (a) 양상태 문서+화해 절차 + (c) 콘텐츠 가드. (b) 레인 직접 실행은 흐름에서 배제 — 소유 예외는 트리거+기록 의무의 최후 수단으로만 성문화(REQ-DSC-008). 근거: spec.md §A.4 트리거 통제 불가 부정 판정.
- 문서 착지지: 절차 본문은 paths 스코프 `kanban-dispatch-mechanics.md` § Isolation(부담 0), 의무 문장만 항상 적재 `agent-common-protocol.md` § User Interaction Boundary lane-session 단락(≤600B).
- 가드: `internal/template/docs_delegation_lane_flow_test.go` — 섹션+마커 콘텐츠 스캔, M1→M2 커밋 순서로 RED-first 를 그래프에 새김.
- **미러 안전 계측 정정(v0.1.1 — 프로브 2건 실측)**: 원 인용 `TestRuleTemplateMirrorDrift` 는 착지 쌍에 도달하지 않는다 — 프로브 1(C1 mechanics description 변이 → 테스트 `ok 0.360s` GREEN 유지)로 기계적 증명; mechanics 는 바이트 패리티 미등록, `agent-common-protocol.md` 는 §25 정화 쌍 제외. 교체 기제: `TestTemplateNoInternalContentLeak` + `TestSanitizedPairParity` — 프로브 2(C2 SPEC-ID 주입 → `FAIL`, 위반 라인 `class=C1-spec-id-prefix | match=SPEC-DOCS-DELEGATION-CWD-001`)로 RED 관측, 복원 sha256 전후 일치. AC-DSC-004 처분: (a) 유도 변이 RED(교체 기제 기준) + 기제 정정 기록.

## §C 열린 항목

- [해소 2026-10-01] decision-index Q1(배제 비준 APPROVED)·Q2(2회+거부 증거 CONFIRMED) — 운영자 판정이 kick-off AskUserQuestion 라운드(2026-10-01, card t1387 lane-1 pane direct answer)로 착지해 decision-index.md 의 Operator verdict 슬롯에 기입됨. D2 클로저 — kick-off 진입 가능 상태.
- `feedback_sync_dispatch_tree_anchor.md` 메모리 파일 미발견(양 저장소 검색) — spec.md §G 미검증 항목. run/sync 에서 발견되면 §A.3 에 보강.

## §E 진행 신호

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: draft-authored (plan-auditor 감사 대기)
- artifacts: spec.md · plan.md · acceptance.md · progress.md · decision-index.md (5파일, Tier M + decision gate on)
- frontmatter: 12 정규 필드 검증 완료 — SPEC ID 정규식 Bash 실행 PASS (`SPEC-DOCS-DELEGATION-CWD-001`), 기존 카탈로그 중복 없음(`ls .moai/specs | grep DELEGATION` 0건)
- lint: `moai spec lint SPEC-DOCS-DELEGATION-CWD-001` → "✓ No findings — all SPEC documents are valid" (1차: 0 error 14 warning → REQ↔AC 추적성 열 + `-run` 패턴 앵커 수리 후 0 findings; v0.1.1 델타 편집 뒤 재실행 0 findings — 2026-10-01 본 트리 실측)
- plan-audit r1: FAIL 0.875(MP-8 가림) — `.moai/reports/t1387/plan-audit-r1.md`. D1 델타 수리 완료(AC-004 처분 (a)+기제 교체, AC-005/006/007/008 RED-now 본 인 재측정치로 추가), D3·D4·D6 접수, D2는 kick-off 전 운영자 직답 유지. 재감사는 델타 스코프로 재위임 대기.

## §E.2 Run-phase Evidence

**트리 특이사항(격리 스폰)** — run-phase는 런타임 격리 워크트리 `.claude/worktrees/agent-a6ac0c2224378b493`(브랜치 `worktree-agent-a6ac0c2224378b493`, 기점 `f22e2d7ac`, card t1387 lane의 스폰)에서 수행됐다. 본 트리에는 plan-phase 산출물이 없어 SPEC 5파일을 카드 트리의 r2 감사본에서 바이트 동일 복사해 왔다(shasum 실측: spec.md `0938ee59…70a4`, plan.md `38ff1fb7…94c8`, acceptance.md `d24507cd…dbdb` — r2 감사 핀과 일치). 레인의 화해는 본 SPEC이 성문화하는 절차 본문 그대로 수행된다 — 카드의 첫 실측 사례. 커밋은 본 트리 브랜치에 쌓이며 push·병합은 하지 않는다(레인 화해 대기).

### M1 — 화해 절차 본문 + 항상 적재 의무 문장 (커밋 그래프 1번째)

- 편집 4파일: `.claude/rules/moai/workflow/kanban-dispatch-mechanics.md` + C2 미러, `.claude/rules/moai/core/agent-common-protocol.md` + C2 미러 — C1↔C2 편집 후 diff 바이트 동일 실측.
- AC-DSC-001 PASS: `grep -c "Reconciling an isolated specialist spawn" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `1`, exit=0 (RED-now 핀: `0` + exit=1).
- AC-DSC-002 PASS: `grep -c "runtime decision" .claude/rules/moai/core/agent-common-protocol.md` → `1`, exit=0 (lane-session 단락 27행 안 적중).
- AC-DSC-005 PASS: `grep -c "ownership exception"` → `1`; `grep -c "structurally"` → `1`; 양성 대조 `grep -c "twice"` → `2`.
- AC-DSC-006 PASS: `grep -c "git merge --ff-only" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `1`, exit=0. **범위 해석 기록**: 금지형 grep(`git -C` / `--git-dir`)은 소절 스코프로 측정 — 소절 내부 0적중(`sed` 추출 후 grep exit=1 ×2). 파일 전역은 기존 측정 인용(guard 거부 형상)이 6건(`git -C`, 행 50/79/81/91/101/104)과 1건(`--git-dir`, 행 81) 존재해 전역 0적중은 불가능 방향(impossible)이고, 신설 소절이 이들을 추가하지 않는 것이 criterion 의 실제 주장이다. 본 인 재측정: 편집 전 파일에 이미 6/1 건 존재(행 번호 실측).
- AC-DSC-007 PASS (N2 지시대로 UNESCAPED 형태): `grep -cE "implemented .* completed|3-phase close" .claude/rules/moai/workflow/kanban-dispatch-mechanics.md` → `1`, exit=0; 양성 대조 `## Integration into the release branch` → `1`.
- 항상 적재 예산: `agent-common-protocol.md` 증분 **487 바이트**(17,284 → 17,771) — ≤600B 목표 이내.

### M2 — 콘텐츠 가드 테스트 (커밋 그래프 2번째 — M1 뒤)

- 착지: `internal/template/docs_delegation_lane_flow_test.go` · `TestReconciliationProcedureDocumented` — 소절 제목 + 9개 필수 마커(5단계 + `WT-<slug>` + `3-phase close` + `ownership exception` + `structurally`) + 의무 문장 마커 2개(`runtime decision`, 소절 포인터) + 금지형(`git -C`/`--git-dir`) 소절 스코프 부정 검사(양성 대조: 소절 존재·마커 검사가 부정 검사보다 선행 — 빈 입력 공멸 차단). 로컬+템플릿 4사본 전부 스캔.
- AC-DSC-003 GREEN: `unset MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_ROLE MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL && go test -count=1 -run '^TestReconciliationProcedureDocumented$' ./internal/template/` → `ok github.com/modu-ai/moai-adk/internal/template 0.430s`, exit=0.
- **변이 프로브 관측 3건** (verification-completeness §1.1 관측된 실패 완료):
  1. 프로브 A (AC-DSC-003a — "증거 수확" 마커 단독 삭제 → `**Collect the reports.**` 로 치환): `--- FAIL: TestReconciliationProcedureDocumented` + `local: reconciliation procedure section lost a mandatory step marker "Harvest the evidence before disposal" (kanban-dispatch-mechanics.md)` — RED 관측.
  2. 복원: 마커 원문 복구 후 `diff` C1↔C2 바이트 동일 재실측 (`C1==C2 restored: identical`).
  3. 최종 GREEN: 안전 배치 `go test -count=1 -run '^TestReconciliationProcedureDocumented$|^TestRuleTemplateMirrorDrift$|^TestTemplateNoInternalContentLeak$|^TestSanitizedPairParity$' ./internal/template/` → `ok 1.811s`, exit=0.
  4. 추가 프로브 B (AC-DSC-003 소절 제목 변이 → `### Reconciling an isolated spawn`): `--- FAIL: ...` + 소절 부재 에러 관측 뒤 복원 — 소절 존재 검사가 물린다는 별도 관측.
- AC-DSC-004 GREEN (교체 기제 — AC 처분 (a) 기준): 같은 배치에서 `TestTemplateNoInternalContentLeak` + `TestSanitizedPairParity` 포함 `ok 1.811s`, exit=0 — 편집된 C2 미러 2파일이 누출·정화 쌍 계측을 통과.
- AC-DSC-003의 RED-now 핀(e927266be: 테스트 파일 부재)은 plan-audit r2 가 본 트리와 동일 조건에서 재측정 완료 — 본 트리 기점 f22e2d7ac 도 테스트 파일 부재 동일 상태에서 시작.
- 품질 게이트: `go vet ./internal/template/...` exit=0 · `moai spec lint SPEC-DOCS-DELEGATION-CWD-001` → `✓ No findings — all SPEC documents are valid` (설치 빌드 v3.2.0-rc.23, `moai_cp/20260925_122548-1711-gd194083fb` — r2 감사와 동일 보조 계측 caveat) · `make build` exit=0 (템플릿 편집 뒤 catalog.yaml 재생성 — 이 룰 2파일은 catalog 해시 대상이 아니어서 tracked 변경 0건 실측).

## §E.3 Run-phase Audit-Ready Signal

- run_status: complete
- run_complete_at: 2026-10-01
- run_commit_sha: pending-backfill-run (커밋 그래프로 대체 — M1 `39e51b022` → M2 커밋 순서가 AC-DSC-008 의 목격자; 본 필드는 sync 커밋이 아닌 run 마지막 커밋을 가리키며 D3 백필 창에 둔다)
- ac_pass_count: 9
- ac_fail_count: 0
- preserve_list_post_run_count: 0 (PRESERVE 위반 없음 — 편집 4룰파일+테스트 1+SPEC 산출물 5만 변경)
- l44_pre_commit_fetch: n/a (격리 워크트리 — 공유 체크아웃 아님, 스폰 시점 분기 상태 기점 f22e2d7ac)
- l44_post_push_fetch: n/a (push 없음 — 레인 화해 대기)
- new_warnings_or_lints_introduced: 0 (`go vet` exit=0, spec lint 0 findings)
- cross_platform_build.go_darwin: pass (`make build` exit=0, darwin/arm64 실측)
- cross_platform_build.windows: not-run (문서+테스트-only 변경 — CI 가 전 판정; GOOS=windows 빌드는 Go 소스 비변경으로 생략, Gaps 기재)
- total_run_phase_files: 10 (4 룰 + 1 테스트 + 5 SPEC 산출물)
- m1_to_mN_commit_strategy: M1(문서+SPEC 산출물+진행) → M2(가드 테스트+진행) 2커밋 — 커밋 그래프가 순서를 목격(VCI §2.3)
- isolated_spawn_note: 본 run-phase 자체가 REQ-DSC-002/003 의 첫 실측 사례 — 격리 트리 브랜치는 런타임 기명(`worktree-agent-a6ac0c2224378b493`) 유지, 레인이 화해 절차(WT- 개명 → ff-only 병합 → 증거 수확)를 수행한다. 증거 파일 `.moai/reports/t1387/run-ac-green.md` 은 gitignored 라 병합에 동행하지 않음 — REQ-DSC-007 수확 대상.

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete
- sync_complete_at: 2026-10-01
- sync_commit_sha: 383954c94e75c9f536f67342c65648af4cdd07cc  # D3-exempt backfill: the placeholder in the sync commit replaced with the real SHA (phase-owned field, manager-docs §E.4)
- changelog_entry_position: [Unreleased] / Fixed / SPEC-DOCS-DELEGATION-CWD-001 (1건, 선두)
- b12_self_test_a: 사전 방출 grep `grep -c 'SPEC-DOCS-DELEGATION-CWD-001' CHANGELOG.md` → `0` + exit=1 (중복 없음)
- b12_self_test_b: AC 카운터(tier M → acceptance.md) → stdout `9`, stderr `live=9 excluded=0 ambiguous=0` — CHANGELOG 엔트리의 9 AC 인용과 일치
- b12_self_test_c: CHANGELOG 인용 경로 6건 전부 존재 실측 — spec.md · progress.md · kanban-dispatch-mechanics.md · agent-common-protocol.md · docs_delegation_lane_flow_test.go · run-ac-green.md
- frontmatter_status_transitions: in-progress → implemented → completed (단일 싱크 커밋에서 병합 전이, spec.md `status: completed` + `updated: 2026-10-01`)
- mx_check: `internal/template/docs_delegation_lane_flow_test.go` — exported 심볼 0건(테스트 패키지 비공개 3함수), goroutine·복잡도 트리거 없음 → 신규 @MX 태그 불요
- codemap_freshness: `go run ./cmd/moai graph check`(본 트리 빌드, 본 트리 판독) — codemaps `value=31 threshold=40 verdict=fresh` (재생성 불요); mx-index·edges는 미추적 런타임 산출물 absent(신규 워크트리 상태, verdict 주체 아님)
- docs_site_readme_judgment: CHANGELOG-only — 내부 하네스 플로우 문서(레인 화해 절차·콘텐츠 가드)로서 README/docs-site 사용자 대면 표면과 무관
- canary_compliance_check: n/a — 본 SPEC이 자기 싱크 테스트를 가지는 forward-looking policy를 정의하지 않음(레인 플로우 절차 성문화이며, 절차의 실측 사례는 §E.2 격리 스폰 기록)
