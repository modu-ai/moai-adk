# SPEC-TUI-CHECKLIST-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-26T21:01:12Z

Plan 단계 산출물 작성 완료 (manager-spec, 카드 t1969, Tier L): spec.md v0.1.0 (REQ 20개), plan.md, acceptance.md (AC 23개, 증거 원장 L1~L14), design.md, research.md. 워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`, plan 측정 HEAD `cb30e05b3e`.

델타 1회차 (manager-spec, 측정 HEAD `d241917d17a3754541ef5ad52536d68023faf142`): 1회차 plan 감사(FAIL 0.71, `.moai/reports/t1969/plan-audit-verdict.md`)의 결함 D1~D20 을 고쳤다. spec.md v0.1.1 (REQ 23개), acceptance.md (AC 24개, 증거 원장 L1~L19 — L15~L19 는 이 HEAD 에서 측정). 레인 결정 Q1~Q3 은 plan.md §G 에 결정과 "리더가 다른 안을 택하면" 단락으로 적었고, Kickoff 기록에서 리더에게 알린다.

델타 2회차 (manager-spec, 측정 HEAD `b422a13d5377639d58fc672c986b81c678fe3318`): 2회차 plan 감사(FAIL 0.75, `.moai/reports/t1969/plan-audit-verdict-iter2.md`)의 결함 N1~N13 과 리더 codex_audit(영수증 `rcpt-bd48fb9e7212e0659ea0f6a3`)의 P2(= N8)를 고쳤다. spec.md v0.1.2 (REQ 23개 그대로, REQ-TCL-022·REQ-TCL-023 문장 수정, 결정 D-10 추가), acceptance.md (AC 24개 그대로, 증거 원장 L1~L25 — L20~L25 는 이 HEAD 에서 측정), design.md §3.2·§3.3.1·§3.4·§3.6·§4·§8, plan.md §A.0·§B·§C·§E·§F·§H·§I, research.md §7·§8 정정과 §9·§10 추가. 레인 결정 Q1·Q2·Q3, D-9, `main` 흡수 시점 규칙, 글리프 `◆`/`○` 는 다시 열지 않았다. 마지막 감사 회차(Tier L 상한 3) 대기.

델타 3회차 (manager-spec, 측정 HEAD `69358ecba8a9b8f996b3e9b0e4b39ae36728a80f`, 천장 정책의 자동 델타 1회): 3회차 plan 감사(FAIL 0.80, `scope: delta`, `.moai/reports/t1969/plan-audit-verdict-iter3.md`)의 `fix_scope` 앵커 안에서 BLOCKING M1~M3 과 SHOULD-FIX M4 를 고쳤다 — spec.md REQ-TCL-015·REQ-TCL-016(v0.1.3), acceptance.md AC-TCL-016·AC-TCL-017, design.md §3.3.1·§3.4·§4. REQ 23·AC 24 와 id 집합은 그대로다. 레인 결정 Q1~Q3, D-9, D-10, 글리프는 다시 열지 않았다. 새 측정 칸(트리 `69358ecba8`)은 AC-TCL-016·AC-TCL-017 본문 안에 적었다(원장 행은 더하지 않음).

범위 축소 v0.2.0 (manager-spec, 측정 HEAD `6f05252d6ddf62362a60cadbfc3350e0d068dbc4`, 천장 뒤 편집 — 리더 결정 기록 §15.98 SPLIT): 델타 4회차 감사(FAIL 0.80, `.moai/reports/t1969/plan-audit-verdict-iter4.md`, 차단 N1)를 받아 이 SPEC 을 렌더러 + `moai init` 연결로 좁혔다. update 연결은 카드 t1978 로 넘겼다(plan.md §K, spec.md §3.4). REQ 23·AC 24 와 id 집합은 그대로다 — 지우거나 다시 매긴 id 는 없고, 각 id 의 update 조각만 걷어내 `[DEFERRED → t1978]` 로 표시했다. Tier 는 측정 근거로 L 을 유지했다(plan.md §A.1). 새 원장 L26~L33 을 이 HEAD 에서 쟀다. 아래 미해결 부채의 처분:
- M5 → 수리(궤도에 `printer/printer.go`·`severity_line.go`, X1·X7·X9 정정, research.md §9 한계 문단).
- M6 → 수리(AC-TCL-013 O4b (i) 범위 `-- internal pkg go.mod go.sum`).
- M7 → 수리(AC-TCL-012 `init_writer_replaced`).
- M8 → 수리(검색 모양 `PillWarn`·`PillErr`, 원장 L26, 일치 9줄을 X7·X9·X11 로 분류).
- M9 → 수리(`inventory_complete` 궤도 항목 열여섯 개를 하위 테스트로 열거).
- R-M2a(4회차 N5) → research.md:127 문단에 정정·이월 표시.
- R-M2b(4회차 N3) → P22 가 update 행으로 이월되어 AC-TCL-024 (2) 의 괄호를 지움.
- R-M1(4회차 N6) → spec.md §6 에서 M5 를 지움(AC-TCL-016 은 init 만 — M4).
- R-M4, R-M4b → t1978 로 이월(plan.md §K K-12).
- 4회차 N1·N2·N4·N7 → t1978 로 이월(plan.md §K K-1~K-4).

델타 4 v0.2.1 (manager-spec, 측정 HEAD `274441db21a4be3cec01b78c0f9d496893d12479`, 리더 결정 기록 §15.102 — 최종 확인 감사 전 마지막 수리): 확인 감사 5회차(FAIL 0.73, `.moai/reports/t1969/plan-audit-verdict-iter5.md`, 차단 B1·B2)의 결함을 한 커밋으로 고쳤다. REQ 23·AC 24 와 id 집합은 그대로다. 처분:
- B1 → spec.md REQ-TCL-010·§3.3, design.md §1·§4, acceptance.md AC-TCL-011: 체크리스트 줄 ≤ 24 + 직접 출력 줄 다섯(`init.go:708` 2줄, `:750`, `:784`, `:1034` P5)의 기준선 대비 증가 0. 깨끗한 init 정의에서 "보존 줄 0" 을 지우고 P5 를 넣었다. 셈(코드 읽기): 체크리스트 23 + 직접 출력 5 = 28. 화면 전체 상한은 운영자 질문으로 열림(spec.md §5, t1978 공유).
- M2 → spec.md §3.3 에 줄 단위를 한 번 정의(접기·ANSI 제거 뒤 `\n` 줄, 터미널 행 아님).
- B2 + 4회차 N4 → acceptance.md AC-TCL-017 (a1) `w-err`, (c) `w-resume`·`c-resume`, (e) `w-nil`·`c-nil` 단언과 반례 `init.go:427`; AC-TCL-017 `maps` 와 spec.md §6 에 REQ-TCL-015; plan.md §K K-3 철회; spec.md §3.5 정합; plan.md 의 매달린 "리더 판단 필요" 제거.
- S1 → acceptance.md AC-TCL-013 에 여섯 고정의 주입 방법(기존 시험 경로만, update `warn` = 사용자가 고친 pre-push 훅, update `fail` = `confirmViaPreviewFn`), design.md §7.
- S2 → spec.md REQ-TCL-012 update 절을 세 고정 실행으로 좁힘, §3.4 F-10, plan.md §K K-16.
- S3 → design.md §3.7 주입 지점 표(새 시험 전용 시접 `checklistSessionStartFn` — M3, 새 파일, 영역 규칙 불변; 기존 `newInitPhaseExecutorFn` 감싸기), acceptance.md AC-TCL-016.
- M1 → design.md §3.3.1(P22 열거 행), §3.4(`runUpdate` 문장), §4.1 의 Q3 문단을 §4 로 이동; acceptance.md AC-TCL-017 의 "(d)·(e)" → "(c)·(e)", AC-TCL-024 (4) 의 X11 에서 P22 제거.
- M3 → acceptance.md AC-TCL-003·AC-TCL-004 를 init 계획표 밖 키의 probe 계획으로(골든 `checklist-probe-{fail,warn}.golden`), design.md §7, plan.md §A.1·M2.
새 원장 L34~L41 을 이 HEAD 에서 쟀다. 레인 결정 Q1~Q3, D-9, 글리프, `main` 흡수 시점 규칙은 다시 열지 않았다.

미해결 부채 (v0.1.3 시점 기록 — 처분은 위 v0.2.0 문단):
- M5 (SHOULD-FIX) — 궤도 추적이 메서드 수신자 작성기를 보지 못해 `internal/cli/printer/printer.go` 가 X9 로 분류됨(design.md §3.3.1 궤도 정의·X9 행, research.md §9 궤도 추적). 3회차 감사의 근거 위치는 design.md:131·:174 로, 앵커 §3.3.1 안이지만 감사가 fix_scope 결함으로 지정하지 않았으므로 손대지 않았다.
- M6 (SHOULD-FIX) — AC-TCL-013 O4b (i) 범위가 `-- internal` 뿐이라 `pkg/`·`go.mod`·`go.sum` 변경을 보지 못함(acceptance.md AC-TCL-013).
- M7 (MINOR) — AC-TCL-012 명령 수준에 조건 2(작성기 교체) 하위 테스트 없음.
- M8 (MINOR) — 검색 모양에 `tui.Pill(… PillWarn|PillErr …)` 없음(`update_template_sync.go:336` → `update_tux.go:120`).
- M9 (MINOR) — `inventory_complete` (다) 가 테스트 안의 궤도 목록만 기준으로 판정(acceptance.md AC-TCL-024).
- 잔여 R-M2a — research.md:127 이 "이미 최신" 경로를 여전히 본문 `update_template_sync.go:249-255` 로 적는다. 실제 명령 경로는 `:1260-1266`(design.md §3.4, AC-TCL-017 (d)). research.md 는 fix_scope 밖.
- 잔여 R-M2b — acceptance.md AC-TCL-024 (2) 의 "P22 는 REQ-TCL-023 의 `✓`" 괄호가 P22 를 항목 안 행으로 적는다. P22 는 이제 항목 밖 행이다(design.md §3.3.1). AC-TCL-017 (d) 가 우선한다고 적어 두었지만, `p22_update_up_to_date` 하위 테스트를 쓰는 run 단계가 두 문구를 함께 읽게 된다.
- 잔여 R-M1 — spec.md §6 추적표 사본은 AC-TCL-016 의 마일스톤을 M4 로만 적는다. `update` 하위 테스트는 M5(update 연결) 뒤에 녹색이 된다.
- 잔여 R-M4 — 깨끗한 update 의 정의(spec.md §3.3)가 `--yes` 를 적지 않는다. `--yes` 없는 실행의 확인 화면 줄은 줄 예산에 넣지 않았다(design.md §4). AC-TCL-011 의 고정 실행이 `--yes` 로 돈다는 전제가 문서 어디에도 고정되어 있지 않다.
- 잔여 R-M4b (문구만) — design.md §3.5 "update 의 확인은 `analyze` 항목을 닫은 뒤 열린다" 는 본문이 직접 확인을 띄우는 경우(`skipConfirm` 거짓, `update_template_sync.go:355`)를 적은 것이다. 실제 명령에서는 확인이 바깥 함수(`:1287`)에서 항목이 열리기 전에 돈다(design.md §4). 두 경우 모두 "질문 중 열린 줄 없음"은 성립하므로 결함이 아니라 문구 보강 대상이다. §3.5 는 fix_scope 밖.

### Plan-phase delta 5 (card t1969, decision record 15.107)

입력: 교차 모델 감사(리더 audit_multi, 영수증 `rcpt-b881bc20c399d54e51107c54`, codex 판정 fail)의 P2 결함 4건과 R-2, 감사 HEAD `eb2a9dfc21`. 리더 결정 기록 §15.107 의 앵커 제한 수리. REQ·AC id 집합은 그대로(REQ-TCL-001~023, AC-TCL-001~024). spec.md 는 version `0.2.1` → `0.2.2` 만 바꿨다(`updated` 는 이미 2026-10-27).

- (1) 확인 — `internal/cli/init.go:1073` `retrackSectionFiles(opts.ProjectRoot, cmd.ErrOrStderr())` → `internal/cli/update_manifest_retrack.go:80-99` → `retrackManifestFiles` → `:57` `"  manifest retrack skip %s: %v"` 후 `continue`. 나머지 파일 `Save` 가 성공하면 `nil` 이라 `init.go:1074` `p.Collect` 는 불리지 않는다. 수리: design.md §3.3.1 범위 문단(149), P17 행(169), X11 행(191) — P17 init 지점을 범위 안 보존 행으로, update 지점은 이월 유지. acceptance.md AC-TCL-024 — maps REQ-TCL-022(보존 줄은 성공 항목에서도 한 번)라 이 기준을 골랐다: Given 행 목록, (3a) `p17_init_manifest_retrack`, 녹색 경로 25 → 26줄, 이월 괄호, 뮤턴트 두 개.
- (2) 확인 — `hook_install.go:181-183`·`hook_install_precommit.go:297-299` `if skip { return nil }`, 설치 줄 `:262`·`:381` 은 `if !skip` 안; `init.go:1045` `mcpDeclined := !opts.MCPProvision` → `:194-196` `return nil`; `opts.MCPProvision = true` 는 `:623`(대화형) 뿐. 암묵 `Done` 규칙은 design.md §2(77행). 수리: design.md §4 표 두 행과 조건부 생략 문단(260-270, 기존 키 `checklist.status.skipped`, 새 키 없음). acceptance.md AC-TCL-001 — maps REQ-TCL-003(`○` 건너뜀)이라 이 기준을 골랐다: 명령 수준 절 `TestChecklistInitConditionalSkip`(`no_hooks`·`mcp_declined`·`control_not_skipped`), RED 칸(트리 `eb2a9dfc21`, `/usr/bin/grep -rn TestChecklistInitConditionalSkip internal/cli` → 빈 출력, 종료 1), 뮤턴트 세 개.
- (3) 확인 — `go run <스크래치>/fmtprobe/main.go`(go1.26.8 darwin/arm64) → `Changed 7` / `変更 %!s(int=7)件`. 수리: acceptance.md AC-TCL-008 Given(인자별 (번호, 동사) 목록 일치 + 대표 인자 `fmt.Sprintf` 결과에 `%!` 없음)과 뮤턴트·사실 칸 한 개.
- (4) 확인 — 다만 리더 지시의 "AC-TCL-013" 은 AC-TCL-022 다(`TestChecklistInitHunkConfinement` 는 acceptance.md AC-TCL-022, 옛 317행 "계속 발화"). 그 문장이 판정을 M6 삭제 전 기록에만 맡겨, 기록 뒤 커밋이 `init.go:728` 을 고쳐도 녹색이었다. 수리: AC-TCL-022 뮤턴트 두 개와 계속 발화(sync 감사가 최종 HEAD 에서 두 명령을 다시 돌려 판정 파일에 남김, 측정 SHA = 최종 HEAD). AC-TCL-013 정규화 목록의 기존 규칙은 그대로다.
- (5) R-2 — AC-TCL-013 정규화 목록에 하이픈 ISO 규칙 `\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z\b` → `<ISO_TS>`(사실 칸 `python3 -I <스크래치>/iso_probe.py`, 양성 2·음성 3, 종료 0)와 "목록 밖 값은 막힘으로 실패" 한 줄.

손대지 않은 것(앵커 밖, 다음 감사가 볼 수 있음): design.md §6 의 "지시자 개수가 같아야"(322행)는 AC-TCL-008 의 새 조건보다 약하다. design.md §3.2 항목 경계 문장(109행)은 그대로 두고 §4 에서 건너뜀 항목이 열린 채 바이트를 받지 않음을 적었다. spec.md HISTORY 줄은 이번 위임 범위(version·updated 줄)가 아니라 더하지 않았다. `.moai/config/sections/*.yaml` 을 `:1073` 앞 단계가 읽는지는 따라가지 않았다 — AC-TCL-024 (3a) 가 그 경우를 막힘 보고로 돌린다.

### Plan-phase delta 6 (card t1969, decision record 15.108)

입력: 8회차 plan 감사(`.moai/reports/t1969/plan-audit-verdict-iter8.md`, 감사 HEAD `aee74a15132a4ca72c25df2441473d37418ef1d8`, 차단 N8-1, 선택 N8-2~N8-7). 리더 결정 기록 §15.108 — §15.107 의 예외(차단 결함이 델타 5 가 남긴 잔여 문장이지 새 설계 결함이 아님). REQ·AC id 집합은 그대로(REQ-TCL-001~023, AC-TCL-001~024). spec.md version `0.2.2` → `0.2.3`(`updated` 는 이미 2026-10-27).

hunk:
- N8-1 (차단) → acceptance.md §F 완료 정의의 AC-TCL-022 항목 한 hunk: M6 삭제 전 기록과 sync 감사의 최종 HEAD 재실행(측정 SHA = 최종 HEAD, 40자)이 **모두** §A 녹색이어야 성립, 기록만으로는 판정하지 않음, 지운 뒤의 `[no tests to run]` 은 판정에 쓰지 않음. AC-TCL-022 본문(계속 발화)과 같은 규칙이다.
- N8-3 → spec.md frontmatter version 한 줄과 HISTORY v0.2.2(델타 5, §15.107, 커밋 `aee74a1513`)·v0.2.3(이 델타, §15.108) 두 행.
- N8-4 → design.md §6 "지시자 개수가 같아야" → "(인자 번호, 동사) 목록이 같아야" (한 줄).
- N8-5 → design.md §3.2 항목 경계 문장 끝에 "호출자가 여는 즉시 `Skip` 으로 닫는 조건부 생략 항목(§4)은 이 규칙 밖" 한 구절 (한 줄).
- 이 단락 자체(progress.md §E.1).

미룬 것(deferred):
- N8-2 — 소스 확인: `internal/cli/update_namespace_protect.go:64`·`internal/cli/update_cleanup.go:208` 의 형식은 `2006-01-02T15-04-05Z`(시각 부분 하이픈)라 acceptance.md:208 의 콜론 규칙이 R-2 대상을 잡지 못한다는 지적은 맞다. 미룬 이유: 감사가 두 수리안(정규식 확장 대 이름 변경)을 내놓아 선택이 필요하고, 정규식을 바꾸면 같은 줄의 사실 칸(트리 `eb2a9dfc21` 의 탐침 출력)을 다시 재야 해 한 줄 편집으로 닫히지 않는다. 막힘 규칙(acceptance.md:212)이 조용한 통과를 막는다.
- N8-6 — spec.md §6 추적표와 plan.md M4 목록 두 파일에 걸친 편집이라 한 줄 기준을 넘는다(조건부 `Skip` 여는 줄을 plan.md M4 에 적는 일도 함께 따라온다).
- N8-7 — AC-TCL-008 의 번호 규칙을 Go `fmt` 규칙(직전 인자 번호 + 1)으로 바꾸는 것은 테스트의 판정 알고리즘을 바꾸는 AC 본문 결정이라 문구 정렬이 아니다. `%!` 검사가 상당 부분을 받친다.

### Plan-phase delta 7 (card t1969, decision record 15.111)

입력: codex P2 두 건(영수증 `rcpt-50fb41f52e6c83eab02ffa04`, 레인 실행 `rcpt-ac0ee402340ec744d7eb4a7a`, 둘 다 `eff053068d`)과 이월 선택 결함 N8-6·N9-1·N9-2·N8-2. 리더 결정 기록 §15.111(1회 승인, 1커밋). 기준 HEAD `eff053068d9421bc76401bf9220dfff484249b32`. REQ·AC id 집합은 그대로(REQ-TCL-001~023, AC-TCL-001~024). spec.md version `0.2.3` → `0.2.4`.

hunk(기준 `eff053068d` 의 줄 번호):
- codex P2-a (AC-TCL-008 서명) → acceptance.md:138(Given: 서명 = Go `fmt` 가 실제로 소비하는 (인자 번호, 동사) 쌍을 번호 순으로 늘어놓은 목록, 소비 규칙, `%[3]s %[1]s %d` 유효, `%s %s %[3]s` 와 `%[2]s %s %[3]s` 는 다른 서명, 실제 `fmt.Sprintf` 대표 값 횟수 비교), acceptance.md:141(뮤턴트 하나, 사실 칸 둘 — `fmtprobe7`·`extraprobe7`, 트리 `eff053068d`, go1.26.8), design.md:329(§6 문구 정렬). 이 회차가 델타 6 의 이월 N8-7 을 닫는다.
- codex P2-b (AC-TCL-016 비활성 대조) → design.md:243-244(§3.7 표 두 행: 시접을 수집기 수준으로 — 활성·비활성 모두 수집기 생성 직후 한 번, 부르는 곳은 init.go 밖, init.go 변경은 R1 안의 세션 시작 줄·`:388` 뿐), design.md:249(비활성에서도 부름, 생산 값 `nil` 이면 부르지 않음 → AC-TCL-013 바이트 보존), acceptance.md:244-245(대조 실행의 구성과 주입), acceptance.md:255(뮤턴트 둘). 따라온 hunk: plan.md:123(M3 시접 설명이 "세션 시작 때" 라 같은 모순을 담아 함께 고침).
- N8-6 → spec.md:188(현재 190) `M2, M4`, plan.md:133(M4 목록에 AC-TCL-001 명령 수준). plan.md 의 다른 마일스톤 표는 AC 별 마일스톤을 적지 않아 고칠 곳이 없다.
- N9-1 → acceptance.md:365(§F: 기록은 `go test` 라 §A 판정, 최종 HEAD 재실행은 git 명령이라 출력·종료 코드·측정 SHA 와 영역 조건으로 판정, 완전성 검사는 재실행에 들지 않고 M6 에서 판정).
- N9-2 → spec.md HISTORY 에 `eb2a9dfc21`(§15.104) 행과 v0.2.4 행, frontmatter version.
- N8-2 → acceptance.md:208 한 줄: 구분자 `[:-]` 로 넓혀 생산 형식 `2006-01-02T15-04-05Z`(`update_namespace_protect.go:64`, `update_cleanup.go:208`·`:445`)을 잡고, 사실 칸을 Go RE2 탐침(`isoprobe7`, 트리 `eff053068d`)으로 다시 쟀다. 콜론 형식도 그대로 잡으므로 기존 판정을 줄이지 않는다.
- 이 단락 자체(progress.md §E.1).

REQ-TCL-021 이 성립하는 이유(P2-b): 시접 `checklistSessionStartFn` 을 부르는 곳은 `internal/cli/checklist_session.go` 나 `internal/cli/init_warnings.go`(수집기 생성자) — 어느 쪽이든 init.go 밖이다. init.go 쪽 변경은 R1 `[383,389]` 안의 세션 시작 줄과 `:388` 수집기 생성 줄(plan.md §B 가 이미 "388 수정 가능" 으로 둠)뿐이라 영역 밖 hunk 가 생기지 않는다.

미룬 것(deferred): 없음. 남은 관찰: AC-TCL-022 의 최종 HEAD 재실행에는 카드 id 완전성 검사가 들지 않는다(본문대로). (1) 은 트리 diff 라 영역 판정 자체는 영향받지 않는다.

### Plan-phase delta 8 (card t1969, decision record 15.115)

입력: codex 가 찾은 AC-TCL-008 손 소비 규칙의 셋째 `fmt` 문법 구석 — 번호 뒤 문자 그대로의 퍼센트(`%[1]s %[2]% %s` 와 `%[1]s %[2]% %[2]s` 가 실제 `fmt` 에서 둘 다 `A % B`)와 선택 결함 셋. 리더 결정 기록 §15.115(1회 승인, 1커밋): 산문 규칙을 다시 고치지 않고 지운다. 기준 HEAD `d239bc3e75686e383708e26408ddd454bf2b5653`. REQ·AC id 집합은 그대로(REQ-TCL-001~023, AC-TCL-001~024), AC-TCL-008 의 뜻(네 언어 서식 안전)도 그대로. spec.md version `0.2.4` → `0.2.5`.

hunk(기준 `d239bc3e75` 의 줄 번호):
- 소비 규칙 삭제 → acceptance.md:138: 서명 정의·소비 규칙·대표 값 표를 지우고, 실제 `fmt` 실행 절차 (i)~(iii)(en 값에서 `%!` 없는 첫 기준 인자 조합을 `fmt` 로 찾음, 네 언어를 같은 인자로 실행, `%!` 부재와 기준 인자별 출현 횟수 비교, 위치는 비교하지 않음)와 고정 회귀 묶음(같음: 문자 그대로 퍼센트 쌍, 다름: `%s %s %[3]s` 대 `%[2]s %s %[3]s`, 유효: `%[3]s %[1]s %d`, 다름(`%!`): `Changed %d` 대 `変更 %s件`, 같음(대조): `%s of %s` 대 `%[2]sの%[1]s`)으로 바꿈. "해석기가 없다" 를 명시해 해석기 대 `fmt` 단언(선택 결함)을 닫음.
- acceptance.md:141: 뮤턴트를 "해석기를 넣는 구현", "`%!` 만 보는 구현" 으로 바꾸고, 사실 칸을 실제로 돌린 탐침으로 — `fmtexec8`(트리 `d239bc3e75`, go1.26.8, stdout 8줄 그대로), `extraprobe7`(트리 `eff053068d`) 유지, `fmtprobe7` 사실 칸은 `fmtexec8` 이 같은 값을 덮어 지움.
- design.md:329: §6 문구를 "실제 `fmt` 실행으로만 판정, 해석기 없음" 으로.
- 선택 결함 design.md:243: 시접 부르는 곳의 둘째 예(세션 시작 함수 — R1 에서 수집기 생성 줄 앞에 들어가 수집기가 아직 없음)를 지우고 "수집기가 세션을 받는 지점(`:388` 이 부르는 `newWarnCollector`)" 으로.
- 선택 결함 acceptance.md:261: AC-TCL-017 (c) "세션 시작 때" → "수집기가 만들어진 직후".
- spec.md frontmatter version 과 HISTORY v0.2.5 행, 이 단락.

손대지 않은 것: spec.md HISTORY v0.2.4 행과 progress.md 델타 7 단락의 "서명" 문구는 그 회차의 기록이라 그대로 둔다. acceptance.md:290·:307 의 "세션 시작 때" 는 머리줄·스피너를 여는 시점에 대한 뮤턴트라 시접과 무관하다.

미룬 것(deferred): 없음.

### Plan-phase delta 9 (card t1969, decision record 15.119)

입력: 리더 audit_multi(영수증 `rcpt-5abf31cc8a860a40811fc6b9`, codex 필수, FAIL)의 P2 두 건 — (1) 인자를 빠뜨린 번역 en `Project %.2s: %s` 대 ja `プロジェクト %[2]s` 가 v0.2.5 절차에서 같다고 판정됨(plan 감사 D1 과 같음), (2) 깨진 en `broken %` 에서 절차의 탐색에 끝나는 조건이 없음(D2 와 같음). 리더 결정 기록 §15.119(1회 승인, 1커밋): 구석을 하나 더 막지 말고 판정 방법을 바꾼다. 기준 HEAD `c4349ba30f1eb9ca96a9162a32e723e945cee794`. REQ·AC id 집합은 그대로(REQ-TCL-001~023, AC-TCL-001~024). spec.md version `0.2.5` → `0.2.6`.

hunk(기준 `c4349ba30f` 의 줄 번호):
- acceptance.md:138: 범용 판정 절차·회귀 묶음 문장을 지우고 "형식 키는 호출 지점 표의 모든 칸에서 `fmt.Sprintf(카탈로그 값, 인자...)` 가 기대 문자열과 바이트 단위로 같다" 로. 서식 밖 조건(키 집합 동일, 비어 있지 않음, 값 ≠ 키, 이모지 범위, ja·zh ≠ en 과 허용 목록, en 키 수)은 그대로.
- acceptance.md:138 뒤 새 항목: 호출 지점 표 열 행(`checklist.summary.` 의 `created_files`·`created_dirs`·`changed`·`preserved`·`warnings`·`elapsed`·`deploy_mode`·`next_steps`·`backup`·`recover`), 인자 형과 값(AC-TCL-005 고정값), en 값, en 기대 출력, ko·ja·zh 칸은 "M1 이 채움". 키 수 가드(값에 `%` 바이트가 있는 키 ⊆ 표, 표 키 ⊆ 네 카탈로그). 형식 지시 없는 키 목록, 막힘 보고 경로.
- acceptance.md:141-154: 뮤턴트를 표 기준으로 다시 씀(빠뜨린 인자, 깨진 en, 가드 두 방향). 사실 칸은 실제로 돌린 탐침 — `table9`(트리 `c4349ba30f`, go1.26.8: v0.2.5 절차가 두 반례에서 실패하는 것과 표 판정이 잡는 것), `encol9`(표의 en 기대 출력 열 개). v0.2.5 의 `fmtexec8`·`extraprobe7` 사실 칸은 지운 절차의 근거라 지웠다. `eb2a9dfc21` 의 `fmtprobe` 사실 칸은 표 판정에도 쓰여 남겼다.
- design.md:329: §6 문구를 호출 지점 표와 키 수 가드로.
- spec.md frontmatter version 과 HISTORY v0.2.6 행, 이 단락.

정직한 한계: en 값과 en 기대 출력은 이 회차에 처음 정한 이 카드의 en 문구다(design.md §1 의 ko 화면을 옮겨 만든 초안 — 이전 SPEC 에는 en 문구가 없었다). 형식 키를 이 열 개로 정한 것도 이 회차의 판단이다. `elapsed` 인자를 문자열(렌더러가 고정 시계로 잰 경과를 만든 값)로 둔 것도 그렇다. ko·ja·zh 기대 출력은 적지 않았다 — 문구가 run 단계 산출물이기 때문이며, M1 이 같은 커밋에서 채우고 가드와 바이트 비교가 판정한다. plan.md M1 항목은 이번 위임 범위(손대지 않을 파일) 밖이라 "표의 세 칸을 채운다" 를 적지 않았다 — AC-TCL-008 본문이 그 의무를 적는다.

미룬 것(deferred): 없음.

### Plan-phase scope split (card t1969, decision record 15.123)

입력: plan 감사 12회차(FAIL 0.90, 차단 D1, `.moai/reports/t1969/plan-audit-verdict-iter12.md`) — 호출 지점 표의 ko·ja·zh 칸에 정확성 규칙이 없다. 리더 결정 기록 §15.123(합의된 안전 고정, 수리 아님, 1회 실행·1커밋): 후속 카드 t1989(이 카드에 의존)가 AC-TCL-008, 그 번역 서식·일관성 검사, 호출 지점 표, D1~D5 를 맡는다. 기준 HEAD `f6a603fa627917767f7a47fbc00333b609bc645c` — 넘기기 전 문구는 그 커밋에 있다. spec.md version `0.2.6` → `0.2.7`.

지운 것: acceptance.md 의 AC-TCL-008 제목과 본문 전부(Given, 호출 지점 표와 키 수 가드, 형식 지시 없는 키 목록, RED·녹색 경로, 뮤턴트, 사실 칸 `fmtprobe`·`table9`·`encol9`, 계속 발화). design.md §6 의 서식 지시 키 문장과 ja·zh ≠ en·허용 목록 문장, §7 probe 단락의 "AC-TCL-008 의 키 수 조건" 괄호. plan.md M1 의 `TestChecklistCatalogFourLocales` 줄과 "en 과 다른 문구여야 한다(허용 목록 예외만)" 절. spec.md §6 추적표의 AC-TCL-008 행, §3.5 의 카탈로그 규칙 줄. REQ-TCL-009 에서 AC-TCL-008 에만 대응하던 절(네 카탈로그에 있는 키, ja·zh ≠ en 과 허용 목록). spec.md §3.1 끝에 이전 줄 하나를 더했다.

다시 가리킨 것: plan.md §H R-5(키 동치·en 과 다름 → 카드 t1989), plan.md §A 산출물 줄(AC 23개), plan.md §F E1(AC-TCL-008 을 뺀 23개), acceptance.md §F 완료 정의 범위(AC-TCL-008 건너뜀), spec.md §3.5 렌더러 계약 범위, acceptance.md 범위 문단(8번 기준을 넘김).

새 수: REQ 23(REQ-TCL-001~023, id 그대로), AC 23(AC-TCL-001~007, AC-TCL-009~024, 다시 매기지 않음). REQ-TCL-009 는 덮개가 남는다 — AC-TCL-009(미지원 언어 대체), AC-TCL-011, AC-TCL-023(원시 키 부재)의 `maps`. 덮개 없는 REQ 는 없다.

Tier·예산 다시 셈: Tier L 유지. 예산 상한은 Tier L 이 REQ 25·AC 25(spec-workflow.md § SPEC Complexity Tier)이고 새 수는 23·23 으로 둘 다 안이다. Tier M 상한(16·16)에는 여전히 넘는다. plan.md §A.1 의 파일·줄 근거는 이 회차에 바뀌지 않는다 — 지운 것은 문서 검사 하나이고, 만들거나 고칠 파일 목록(카탈로그 yaml 네 개 포함)은 그대로다.

덮개가 없어진 조건(t1989 로 넘김): 네 카탈로그 `checklist.*` 키 집합 동치, 값이 비어 있지 않음, 값 ≠ 키, 이모지 범위, ja·zh 값 ≠ en 과 허용 목록, en 키 수 = design.md §6 이 카드 범위 키 수, 서식 지시 키의 언어별 출력 일치(호출 지점 표).

미룬 것(deferred): 위 조건 전부 — 카드 t1989.

### Plan-phase delta 10 (card t1969, leader ruling after audit iter16)

입력: 리더가 전한 codex P2 — AC-TCL-021 의 `git log --no-merges --grep=t1969 …` 은 병합 커밋이 들여온 템플릿 변경(충돌 해결·악성 병합)을 보지 못한다(감사자가 임시 저장소에서 재현). 리더 판정: 검사 범위를 넓히고 재현을 음성 대조로 더한다, 다른 것은 바꾸지 않는다. 기준 HEAD `d6f566c442c2eb16e00931aa08752335d504f366`. spec.md version `0.2.7` → `0.2.8`.

측정(일부러 만든 저장소, 레포 밖 scratch, git 2.54.0 (Apple Git-157), 고정 날짜·이름이라 SHA 재현 가능; 기준 `fe5d48ef1851fa06afe2f275f835ffcdd04118e1`, 로컬 `main` `ce31c645638c8c9d0a6f5d8965e68f92e80b531e`, 카드 커밋 `41b952a305325c4d1ffb55915e0c53dce3a9c0bb`):
- 깨끗한 흡수 HEAD `fa21a881a4548f0ce4d994e198533be9d4be5cee`: 옛 명령 빈 출력 종료 0, 새 명령(`--merges --cc`) 빈 출력 종료 0.
- 병합 해결이 `evil.md` 를 더한 HEAD `3f4e5b53a1f9ddd4c4ac426b52e4c37006e32781`: 옛 명령 빈 출력 종료 0(RED), 새 명령 `internal/template/templates/evil.md` 종료 0(빨강).
- 이 트리(`d6f566c442`)에서 두 명령 모두 빈 출력 종료 0(원장 L42·L43).

hunk: acceptance.md 문서 핀 한 문장, 원장 L42·L43 과 그 설명 한 문장, AC-TCL-021 의 Given-When-Then((2) 추가)·종류·뮤턴트·사실 칸. plan.md §B 범위 검사 단락 한 문장. spec.md frontmatter version 과 HISTORY v0.2.8 행, 이 단락. REQ·AC id 와 수는 그대로(REQ 23, AC 23).

같은 맹점을 훑은 결과(고치지 않음, 리더 판정 범위 밖): AC-TCL-022 (2) 는 `--merges` 로 병합을 직접 보고 (1) 은 트리 diff 라 이 맹점이 없다. AC-TCL-019 뒤 절의 위저드 검사(`--no-merges … -- internal/cli/wizard`)는 같은 모양이라 같은 맹점이 있다. §A 카드 id 완전성 검사와 AC-TCL-013 O3 도 `--no-merges` 지만 목적이 커밋 선택이라 같은 결함 종류로 보지 않는다. 처분은 리더 판단에 남긴다.

미룬 것(deferred): 없음.

### Plan-phase delta 11 (card t1969, leader ruling — AC-TCL-019 extension)

입력: 델타 10 단락이 남긴 AC-TCL-019 뒤 절의 같은 맹점. 리더 판정: "AC-TCL-019 확장 승인, AC-TCL-021 과 같은 모양", 다른 것은 바꾸지 않는다. 기준 HEAD `27f404a9d39eb206832a741f50fb3e4d6a288f90`. spec.md version `0.2.8` → `0.2.9`.

측정(일부러 만든 저장소, 레포 밖 scratch, git 2.54.0 (Apple Git-157); 작성 날짜·이름은 고정했으나 커밋 시각은 실제 시각이라 SHA 는 다시 만들면 달라진다; 기준 `4d7c09825db3fa2d0b911c3e3db8abdbde6e9e3d`, 로컬 `main` `1dac172db32e5e2a94eb3f83f83c73e42c2129ab`, 카드 커밋 `421c3428afe4c723a879349629f51cba2766dfa2`):
- 깨끗한 흡수 HEAD `37cdc46c503cb6c933d4c170166be97afa9c22eb`: 옛 명령 빈 출력 종료 0, 새 명령(`--merges --cc`) 빈 출력 종료 0.
- 병합 해결이 `evil.go` 를 더한 HEAD `9e7fdf1f01b7cab7f68ef58b49a214203dd68e50`: 옛 명령 빈 출력 종료 0(RED), 새 명령 `internal/cli/wizard/evil.go` 종료 0(빨강).
- 이 트리(`27f404a9d3`)에서 두 명령 모두 빈 출력 종료 0(원장 L44·L45). 범위의 병합 커밋은 0개다.

hunk: acceptance.md 문서 핀 한 문장, 원장 L44·L45 와 그 설명 한 문장, AC-TCL-019 녹색 경로의 보조 검사((2) 추가)·뮤턴트·사실 칸. plan.md §B 범위 검사 단락 한 문장. spec.md frontmatter version 과 HISTORY v0.2.9 행, 이 단락. REQ·AC id 와 수는 그대로(REQ 23, AC 23).

미룬 것(deferred): 없음.

## §E.2 Run-phase Evidence

런 단계 기록(manager-develop, cycle_type=tdd). 모든 줄은 이 트리 `.moai/worktrees/t1969`(브랜치 `WT-moai-init-update`)에서 실제로 돌린 명령과 그 출력이다. 증거 앞에 측정 HEAD 를 적는다.

### 사전 점검 (HEAD `f029bf4ad32ffd8c24d8e8ed75708f26fb759909`, M0 앞)

- `git branch --show-current` → `WT-moai-init-update`
- `git rev-parse --short HEAD` → `f029bf4ad3`
- 섹션 B 검사 `git log --oneline cb30e05b3e..main -- internal/cli/init.go internal/cli/wizard internal/template/templates internal/tui` → 빈 출력(종료 0). 로컬 `main` = `cb30e05b3e`, `git rev-list --count --left-right HEAD...main` → `15	0`(이 브랜치가 앞서고 `main` 에 새 커밋 없음). t1966 의 wizard·init 변경은 `main` 에 들어오지 않았다 — 막힘 아님. 위임 지시(섹션 B)에 따라 origin/main 은 흡수하지 않는다(plan.md §B·킥오프 기록의 N-C "M0 앞 흡수"와 다르며 위임 지시를 따랐다).
- `git log --oneline -3 origin/main -- internal/cli/init.go` → `8c36a7904b t1909: remove general moai codex ...`, `cd19e5494d merge: absorb main ec193e4777 ...`, `75a6a9ff60 fix(SPEC-USERASSET-DEPLOY-GUARD-001): ...`
- `git diff --stat cb30e05b3e -- internal/cli/init.go` → 빈 출력(종료 0)
- t1966 브랜치 `git rev-parse WT-moai-init-moai` → `6d38362acb5688c5d7fabddc0cbc2b83febe162b`; `git diff -U0 cb30e05b3e...WT-moai-init-moai -- internal/cli/init.go` 의 hunk 머리: `@@ -103,5 +103,4`, `-237,5 +236,4`, `-259,0 +258,7`, `-332,0 +338,19`, `-561,0 +586,20`, `-596,9 +640,7`, `-730,20 +771,0`, `-775,6 +797`, `-866 +883`. 이 카드의 R3 [919,1075]·R2 [1077,1087]·R1 [383,389] 와 겹치는 hunk 는 없다(가장 가까운 것이 `-866`).
- `go build ./...` → 종료 0. `GOOS=windows GOARCH=amd64 go build ./...` → 종료 0.
- t1966 레인 통지: 피어 메시지 채널이 이 서브에이전트에 없음(도구 없음) — 전달 결과 없음. 틈(Gap)으로 기록한다.

### M0 — 기준선 특성화 (AC-TCL-013)

- RED(골든 부재). 명령 `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v`, 출력(발췌, 골든 디렉터리가 비어 있는 트리):

```text
--- FAIL: TestChecklistFallbackMatchesBaseline (7.16s)
    --- FAIL: TestChecklistFallbackMatchesBaseline/init_clean (1.68s)
    --- FAIL: TestChecklistFallbackMatchesBaseline/init_warn (1.26s)
    --- FAIL: TestChecklistFallbackMatchesBaseline/init_fail (0.63s)
    --- FAIL: TestChecklistFallbackMatchesBaseline/update_clean (1.31s)
    --- FAIL: TestChecklistFallbackMatchesBaseline/update_warn (1.33s)
    --- FAIL: TestChecklistFallbackMatchesBaseline/update_fail (0.96s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	8.314s
```

  각 하위 테스트의 실패 문구는 `baseline golden .../<name>-stdout.golden is missing` 이다(여섯 모두 고정 실행을 끝낸 뒤 골든 비교에서 멈춤).
- 골든 생성: `UPDATE_GOLDEN=1 go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` (변경 전 코드, 골든 12개). 줄 수: init-clean stdout 34 / stderr 24, init-warn 34 / 30, init-fail 33 / 3, update-clean 80 / 0, update-warn 80 / 1, update-fail 37 / 1.
- 뮤턴트: 정규화 규칙 `\b\d{8}T\d{6}Z\b` 를 지운 채 `update_warn` 을 돌림 → `FAIL`, `expected: "! user-modified pre-push hook was replaced; previous hook backed up at .git/hooks/pre-push.bak.<TS>"` / `actual: "... pre-push.bak.20261027T070543Z"`. 규칙을 되돌린 뒤 녹색.
- 양성 대조(테스트 안 단언): init_clean — 직접 출력 조각(슬림 안내 두 줄, `user-asset ensure: `, 워크트리 안내 `shared across concurrent sessions`, `Initializing MoAI project...`) 각각 정확히 1번; init_warn — `Failed to apply autonomy tier bundle`; init_fail — `template deployment`; update_clean — `Pre-push hook installed` 정확히 1번; update_warn — `user-modified pre-push hook was replaced`; update_fail — `confirm merge for`. 모두 통과.
- 같은 테스트 `-count=2` 연속 실행 모두 통과(결정성). `<ISO_TS>` 규칙은 이번 여섯 실행의 어느 골든도 바꾸지 않았다(골든에 `ISO_TS` 0건, 규칙은 그대로 둠).

#### O4a — M0 커밋 직후, 다음 커밋 전

- `git rev-parse HEAD` → `5f14cff995dd0258833fc19e37123bf96d44ca7a`
- `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` (측정 HEAD 위와 같음):

```text
--- PASS: TestChecklistFallbackMatchesBaseline (7.87s)
    --- PASS: TestChecklistFallbackMatchesBaseline/init_clean (0.83s)
    --- PASS: TestChecklistFallbackMatchesBaseline/init_warn (0.88s)
    --- PASS: TestChecklistFallbackMatchesBaseline/init_fail (0.87s)
    --- PASS: TestChecklistFallbackMatchesBaseline/update_clean (2.25s)
    --- PASS: TestChecklistFallbackMatchesBaseline/update_warn (1.82s)
    --- PASS: TestChecklistFallbackMatchesBaseline/update_fail (1.22s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	8.923s
```

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Decision: serial

### 입력 매개변수

- tier: L (spec.md `tier` 기준 5개 산출물)
- scope: 신규 파일 약 20개 이상(렌더러·세션·계획·카탈로그 4개·골든·테스트) + `internal/cli/init.go` 3개 영역 삽입
- domain count: 2 (`internal/tui`, `internal/cli`) + 카탈로그 4개 로케일
- file language mix: Go 소스·테스트, YAML 카탈로그, 골든 텍스트
- concurrency benefit: LOW (코딩 중심, 마일스톤 간 의존: M1 카탈로그 → M2 렌더러 → M3 세션 → M4 연결)
- delegate: manager-develop, effort: high, cycle_type: tdd (위임 지시 섹션 A)
- Agent Teams 전제: 해당 없음(명시 요청 없음)

### 모드 평가표

| 모드 | 선택 | 근거 |
|------|------|------|
| direct | 아니오 | 의미 변경이 있는 신규 구현 |
| serial | 예 | 코딩 중심, 마일스톤 순서 의존, 한 트리에 쓰는 작성자 하나 |
| fanout | 아니오 | 병렬 독립 연구 아님 |
| sweep | 아니오 | 균일한 기계적 변환 아님 |

### 판단 근거

코딩 중심 작업이고 마일스톤이 앞 단계 산출물(카탈로그 키, 렌더러 API)에 의존하므로 병렬화 이득이 낮다. 한 작업 트리에는 작성자 하나만 쓴다. Kickoff 는 리더가 자율 형태로 승인했고(감사 PASS 0.92), 이 기록은 첫 구현 커밋 앞에 남긴다.
