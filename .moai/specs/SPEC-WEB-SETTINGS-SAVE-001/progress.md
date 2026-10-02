# progress — SPEC-WEB-SETTINGS-SAVE-001 (card t1393)

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출 완료(2026-10-01, manager-spec): `spec.md` · `plan.md` · `acceptance.md` · `progress.md` · `decision-index.md` — Tier M.
- SPEC ID 사전 검사: `SPEC-WEB-SETTINGS-SAVE-001` 정규식 PASS(원문 출력 plan 세션에 기록). ID 충돌 0(카탈로그 jq 조회).
- 스코프 ① 근본원인은 **확정되지 않았다** — plan.md §A.2 가설 체인 + M1 관측-RED 재현이 판정한다. (c) t1314 seam/no-op 게이트가 plan-phase 증거상 최유력 후보다.
- 미결정 4행(decision-index.md): Q1 가드 술어(EVIDENCE-NEEDED) · Q2 dirty 배지 정체(EVIDENCE-NEEDED) · Q3 핀 3중 정합 표면(DECIDED — SPEC-MODEL-MATRIX-UPDATE-001) · Q4 미확증 거부 의도(POLICY-COVERED — AGENTS.md §3).
- 상태: `status: draft` — plan-audit 대기. Run 진입은 plan-audit PASS + Kickoff(운영자 직답, 레인) 후.
- **plan-audit iter 1 FAIL 0.93 → 수리(iter 2/3, 2026-10-01)**: R1-R8 전부 적용 — R1 Blocker 7행 4요소 처분(EV-001·002 오늘 측정 고정 + AC-WSS-001~004·010 regression-guard 재분류, §2.1) · R2 재측정 단위 5패키지(settings·profile 추가) · R3 `unset MOAI_*` 글로브 폐기→명시 변수 나열(10개, D3 실측 기반) · R4 AC-WSS-001 운영자 실효 요청 형태 고정 + 미재현 분기 명문화(codex overlay 보고는 미검증 Gap으로 기록) · R5 Q1 DECIDED(fetch 없는 원격추적 도달성 — `git merge-base --is-ancestor` ∨ `git cherry` patch-id, REQ-WSS-302에 술어 명시) · R6 `tmux_preferred` 오기 수정 · R7 spec.md frontmatter `tier: M` 추가 · R8 GLM resolver 현재값 `{glm-5.3-flash, effort 빈 값}` 기재 + 기대 델타를 "resolver 폴백 모델 전환"으로 정정. 옵션 D9(AC-WSS-014 REQ 인용)·D11(마일스톤 ### 헤딩 전환)도 함께 적용. D10은 감사 판정대로 의도된 시늅 명명으로 보존.

## §E.2 Run-phase Evidence

### M1 (2026-10-02, this tree HEAD 271746474) — 스코프 ① 진단: 미재현 분기 확정

- **Q2 해소**: 운영자 설치본 = `~/go/bin/moai` v3.2.0-rc.23 (build `moai_cp/20260925_122548-1711-gd194083fb`, 2026-09-30 빌드). `git merge-base --is-ancestor 9be71a4f1 d194083fb` exit 0 — **설치본은 t1314 seam 재작성을 포함**. d194083fb→HEAD 저장경로 파일(handlers.go·schemaform.go·sectionapply.go·sectionwrite.go·yamlpatch) 무변화 — 유일 변화는 t1278 UI 개편(폼 계약 무변경) + t1381 생성물 재생성.
- **dirty 배지**: 양 빌드 모두 런타임 생산자 없음(rc.23 app.js `dirty` 0히트 실측; settings_shell.go:61은 error|saved|clean만 반환) — 운영자 "미저장 표시"는 실 dirty 상태가 아니다.
- **관측-RED 재현 시도**: 브라우저-충실 풀폼(렌더 페이지에서 추출한 152키 교차 탭 POST — 양성 대조로 폼 실재성 확인: worktree 4토글+companion·감사 라디오 현재값 전량 포함)에 운영자 편집(auto_create·auto_merge ON, claude.effort→high)을 얹어 POST → **본 트리에서 GREEN**(수정 전 트리, `go test -count=1 -run TestFullFormSave ./internal/web/` ok 0.711s). 연산자 실디스크 형태(빈 claude 핀 업서트 + codex 핀 보존 + 미모델링 키 보존) 변형도 GREEN.
- **가설 판정**: (a) 원자거절·(b) 시늅 예외/렌더 500·(c) t1314 no-op 오판 — 전부 미재현. REQ-WWS-006 중복 거절도 배제(rc.23 codex 패널은 폼 요소 0짜리 읽기전용 미러 실측).
- **판정서**: `.moai/reports/t1393/root-cause.md` — codex overlay Gap("최소 제출 정상 저장")은 풀폼 실측으로 대체·소멸.
- **처분**: AC-WSS-001 명문화된 미재현 분기 — 구현 강행 없음, 재현 테스트 2종을 영구 회귀 가드로 보존. 잔여 설명(스테일 moai web 프로세스 등)은 root-cause.md §6 Gaps에 기록.

### M2 (2026-10-02) — 스코프 ① 수리 처분: 수리 대상 공집합

- M1 미재현 분기에 따라 수리 대상이 없다 — 재현 테스트 2종이 수정 전 트리에서 태어나 GREEN인 것이 그 자체로 최강 증거다(결함이 있었다면 적색이었을 테스트가 무수정 트리에서 통과).
- AC-WSS-005 무손실 회귀 0: `go test -count=1 -run 'TestApplySchemaEdits|TestNestedSeamEdit|TestPatchFile|TestSyncToProjectConfig|TestSharedNestedSeam' ./internal/settings/ ./internal/profile/` — settings 29 PASS + profile 전항 PASS, 양 패키지 ok. t1314 무손실 테스트 무수정.
- 산출물 변경 0 — 이 행이 M2의 완료 기록이다.

### M3 (2026-10-02) — 스코프 ③ 원격병합 확인 가드

- **가드 착지**: `cleanupSessionWorktree`의 dirty→unpushed 가드 뒤에 착지 확인 단계 추가(session_worktree.go). 술어는 Q1 DECIDED 그대로 — (i) `git merge-base --is-ancestor HEAD refs/remotes/origin/develop` 또는 (ii) `git cherry refs/remotes/origin/develop HEAD` 출력에 `+` 행 부재(patch-id 등가 — 스쿼시 커버). fetch 없음(REQ-WSS-304), 오류·참조 부재·detached HEAD는 fail-open 보존(REQ-WSS-302), merge-first 순서 불변(REQ-WSS-303 — 가드가 함수 내부라 3종료 경로 web.go:115·init.go:466 직접 호출 + profile_setup.go:310 시늉 경유 모두 상속, REQ-WSS-305).
- **관측-RED(배선 전)**: PushedUnmergedPreserved·MissingIntegrationRefPreserved·LandingCheckErrorPreserved 3건 적색 — 푸시-미병합/참조부재/검사오류 트리가 전부 제거되는 결함의 실측. 배선 후 동일 3건 GREEN.
- **술어 실측 정정**: `git cherry`는 등가 커밋을 빈 출력이 아니라 `- <sha>`로 표기 — 스쿼시 픽스처에서 공허-출력 판정이 못 박히는 것을 실측으로 잡고 "no + line"으로 수리. SquashMergedPatchIdRemoved가 이 규약을 고정.
- **t673 컨트롤 승계**: 구 TestCleanupSessionWorktree_PushedBranchStillRemovable(푸시=제거 가능 전제)는 REQ-WSS-306이 뒤집은 계약 — 가드 착지 시 적색 관측 후 승계 처리, 시나리오는 착지 테스트 파일이 재커버(파일 헤더에 SUPERSEDED COVERAGE NOTE).
- **검증**: `go test -count=1 -run 'TestCleanupSessionWorktree|TestSessionExitAutoMerge|TestPRMergeCleanup' ./internal/cli/` ok 22.930s — t673 가족·자동머지 순서 가족(AC-WSS-013)·PR-병합 가족 무수정 GREEN.

### M4 (2026-10-02) — 스코프 ② 감사 핀 3중 정합

- **변경 3표면**: ① Go 기본값(defaults.go Audit 블록) — claude effort medium→high + GLM 핀 신설 `{glm-5.3(DefaultGLM53), max}` · ② 배포 템플릿 workflow.yaml — 동일 값 + 주석 갱신, `make build` 재생성 드리프트 0(git status 클린) · ③ resolver 폴백 — mcp_claude.go `claudeAuditDefaultEffort` medium→high, mcp_glm.go `glmAuditDefaultModel` config.DefaultGLM53 전환 + `glmAuditDefaultEffort = "max"` 신설, `resolveGLMAuditModelEffort`가 effort를 반환.
- **codex 불변 확인**: `{gpt-6.1-sol, high}` 3표면 유지 — TestAuditConfig_DefaultProfile 단언 GREEN(AC-WSS-008).
- **관측-RED(수정 전 값 단언 5건 적색 관측 후 갱신)**: mcp_audit_config_test.go:61(claude effort want medium), mcp_glm_fallback_test.go:28/50/82(폴백 track DefaultGLMHigh·want glm-5.3-flash), retained_model_surfaces_char_test.go:160-162. 원문 출력은 verdict.md.
- **부수 결함 건 발견·수리**: `resolveGLMTaskModel`(glm_task.go)이 `glmAuditDefaultModel`을 공유 — 핀 변경이 태스크 경로를 몰래 따라가는 구조(REQ-AMP-008/REQ-WSS-204 위반 1커밋 거리). 캐릭터리제이션 테스트가 즉시 적색으로 잡았고(glm task default = glm-5.3, want flash), `glmTaskDefaultModel = config.DefaultGLMHigh` 상수를 분리해 태스크 경로를 불변으로 고정. TestResolveGLMTaskModel_BackendDefault·TestGLMAuditPin_TaskResolutionUnaffected GREEN.
- **갱신 테스트**: mcp_glm_fallback_test.go(파생 가드를 분기-고정 가드로 재작성), mcp_glm_audit_pin_test.go(태스크 상수 지시 + 무핀 케이스 {default, max}), retained_model_surfaces_char_test.go(감사 폴백 분기 단언), mcp_audit_config_test.go(claude high + GLM 핀 신설 단언).

### M5 (2026-10-02, HEAD 360c41b5d 측정) — 재측정 + 증거 반출 + N1-N4

- **N1 백필**: EV-001/002의 `<PENDING-BID67HBR9>` stdout-꼬리·exit code 셀을 기준선 원장 발췌로 채움(acceptance.md §D.8). N3: 표식의 착지 지점을 같은 셀에 명기. N4: 재분류 5행(001-004·010) 심각도 셀에 `Blocker (§D.0 regression-guard)` 병기. N2: AC-WSS-014 행에 차등 판독 문구 반영.
- **전체 재측정(AC-WSS-014)**: 10변수 env-scrub 1회 호출, `-count=1 -timeout 30m` 소관 5패키지 통째 — web ok 55.595s · cli FAIL 1801.638s(30m 창·기준선 동일 형상) · config ok 9.348s · settings ok 2.383s · profile ok 2.262s, FINAL-EXIT=1. 원문 전문 `.moai/reports/t1393/final-measure-20261002.txt`(4,506행), 수정 후 셀 EV-003/004.
- **차등 판독**: 기존 적색 2건 동일 + TestStopChainMemberCostWithinBudget(기존 적색) 통과(유리 방향 변동). 유일 신규 적색은 TestAuditLagUsesBinlagSeam 스윕이 M3 착지 술어의 merge-base 좌표를 미선언으로 깃은 것 — `44a5d8c6a`에서 t1383 선례대로 선언, 선별 재실행 PASS 2.43s. 신규 유발 결함 0.
- **품질 게이트**: go vet(web·config·cli) exit 0 · golangci-lint 소관 5패키지 `0 issues.`(v2.1.6=CI 판) · gofmt 변경 파일 클린 · make build 드리프트 0.
- **decision-index Q2**: 리드 처분 기록 반영(바이너리 식별·배지 생산자 부재·풀폼 GREEN — 측정 완결, 스크린샷 재판독만 운영자 몫).
- **verdict**: `.moai/reports/t1393/verdict.md` — AC 자체 점검 행렬 16/16(regression-guard 001은 비재현 분기 이행), 커밋 대장 5건, Gaps 5건, Residual-risk 4건.
- **§E.3**: 아래 시그널 — run-phase 종결. §E.4는 sync 몫.

## §E.3 Run-phase Audit-Ready Signal

- run_phase_complete_at: 2026-10-02 (HEAD 44a5d8c6a — 측정은 360c41b5d에서, 스윕 선언 후속)
- AC 16행 자체 점검: 전항 근거 명령·관측이 verdict.md에 반출됨(verification-claim-integrity §3 5-섹션 형식). regression-guard 재분류 행(001-004·010)은 §D.0 처분대로 pass 미기록 — 001은 비재현 입증 셀로, 002/003/004/010은 관측-RED 선행의 GREEN 테스트로 각각 충족.
- 무손실 회귀 0: t1314 테스트 무수정(EV-002→EV-004 동일 판정).
- 신규 유발 적색 0(차등 판독 — verdict 전체 재측정 절 표).
- 남은 것: sync(E.4 · completed 전이 · sync_commit_sha) — 레인의 통합 창 요청은 완료 보고 후 리드 소관.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
