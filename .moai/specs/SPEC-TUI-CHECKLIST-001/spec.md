---
id: SPEC-TUI-CHECKLIST-001
title: "moai init 진행 출력을 그룹 체크리스트로 재설계 (렌더러 + init 연결)"
version: "0.2.10"
status: in-progress
created: 2026-10-27
updated: 2026-10-27
author: Goos Kim
priority: P1
phase: "v3.3.0 target"
module: "internal/tui, internal/cli"
lifecycle: spec-anchored
tags: "tui, init, progress, checklist, i18n, t1969"
tier: L
related_specs: [SPEC-V3R6-UPDATE-PROGRESS-001, SPEC-CLI-TUX-INIT-UPDATE-001, SPEC-V3R6-UPDATE-NOISE-001, SPEC-INIT-QUIET-WIZARD-001]
---

## HISTORY

- 2026-10-27 — v0.1.0 초안 (manager-spec, 카드 t1969 plan 단계, Tier L). 근거는 같은 디렉터리의 research.md(파일:행 사실, HEAD `cb30e05b3e`)와 design.md(렌더러 API·이벤트 모델·대체 규칙·i18n 키·골든 전략)다. 형제 카드 t1966 의 init.go 변경 범위는 공유 브랜치 ref `WT-moai-init-moai`(`d7f4d5c288`)에서 읽었다(research.md §7).
- 2026-10-27 — v0.1.1 plan 델타 1회차 (manager-spec). 1회차 plan 감사(FAIL 0.71, 감사 HEAD `d241917d17`)의 결함 D1~D20 을 고쳤다. REQ 20 → 23 (REQ-TCL-021 init.go 변경 영역 — 제약 C-2 를 REQ 로 올림, REQ-TCL-022 열린 항목 안 경고·오류·안내 줄 보존, REQ-TCL-023 성공 조기 반환은 정상 종료), AC 23 → 24 (AC-TCL-024 보존 줄 주입 시험). 경고 수 정의를 "수집기 수" 하나로 정했다(D6). 레인 결정 Q1~Q3 을 반영해 "리더 확인 대기" 문구를 지웠다(D16). 대체 골든을 `internal/cli/testdata/checklist/fallback/` 하위 디렉터리로 옮겨 기준선 선행 증거를 디렉터리 단위로 잰다(D2). 증거 원장에 L15~L19 를 더했다(측정 HEAD `d241917d17`).
- 2026-10-27 — v0.1.2 plan 델타 2회차 (manager-spec). 2회차 plan 감사(FAIL 0.75, 감사 HEAD `b422a13d53`)의 결함 N1~N13 과 리더 codex_audit 의 P2(= N8)를 고쳤다. REQ·AC 수는 그대로(23·24). 보존 줄 목록을 `internal/cli` 전체(하위 패키지 포함)에 대한 고정 검색 명령의 출력으로 정하고 23행으로 늘렸으며(N1), AC-TCL-024 는 행마다 하위 테스트 하나와 목록 완전성 하위 테스트를 갖는다(N2). 재개 경로를 실제 실행기의 이벤트 순서(검증 단계 `StepError` 가 먼저)에 맞춰 다시 썼다(N3). update 요약이 그려지는 지점을 템플릿 동기화 본문 끝으로 정했다(결정 D-10). 순서 증거에 O4(M0 트리 재실행)와 카드 id 완전성 검사를 더했다(N6·N7). 증거 원장에 L20~L25 를 더했다(측정 HEAD `b422a13d53`).
- 2026-10-27 — v0.1.3 plan 델타 3회차 (manager-spec, 천장 정책의 자동 델타 1회). 3회차 plan 감사(FAIL 0.80, 감사 HEAD `69358ecba8`)의 `fix_scope` 안 결함 M1~M4 만 고쳤다. REQ·AC 번호와 수는 그대로(23·24). REQ-TCL-015 가 update 의 심각도 줄(두 명령 작성기 모두)을 명시하고, 요약 없이 끝나는 실행에서 항목 밖 수집기 경고를 한 번 쓰게 했다(M1·M3). REQ-TCL-016 이 그 조항을 오류·취소 경로에 잇는다(M3). AC-TCL-016 에 `update` 하위 테스트, AC-TCL-017 에 실패 뒤 항목 밖 `Collect` 변형과 실제 명령 경로의 "이미 최신" (d) 를 넣었다(M1·M2·M3). design.md §3.3.1·§3.4·§4 를 실제 코드 순서와 실제 명령 경로에 맞추고 update 줄 예산을 다시 셌다(M2·M4). 범위 밖 결함 M5~M9 와 잔여 문구는 progress.md §E.1 의 미해결 부채로 남겼다.
- 2026-10-27 — v0.2.0 범위 축소 (manager-spec, 천장 뒤 분할 — 리더 결정 기록 §15.98, `.moai/reports/leader-plan-20261023/status-recovery-20261025-2300.md` 852-859행). 감사 경로: 1회차 FAIL 0.71(`d241917d17`) → 2회차 FAIL 0.75(`b422a13d53`) → 3회차 FAIL 0.80(`69358ecba8`) → 델타 4회차 FAIL 0.80(`6f05252d6d`, 차단 N1 — 깨끗한 update 의 24줄 예산이 설계로 성립하지 않음, 델타 자격 불충족). 리더 판정 SPLIT: 이 SPEC 은 (A) 렌더러(`internal/tui/checklist.go`, 네 언어 카탈로그 키, 골든 기준선)와 (B) `moai init` 연결로 좁히고, update 연결은 새 카드 t1978 로 넘긴다. REQ·AC id 는 하나도 지우거나 다시 매기지 않았다(23·24 그대로) — 각 id 의 update 쪽 조각만 걷어내 §3.4 와 plan.md §K 에 `[DEFERRED → t1978]` 로 적었다. 함께 고친 것: 3회차 감사 M5(궤도에 `printer/printer.go`·`severity_line.go`), M6(O4b (i) 범위에 `pkg go.mod go.sum`), M7(AC-TCL-012 `init_writer_replaced`), M8(검색 모양에 `PillWarn`·`PillErr`, 원장 L26), M9(`inventory_complete` 궤도 항목 열거), 4회차 N3(P22 가 update 행이라 함께 이월되어 AC-TCL-024 (2) 괄호 삭제), N5(research.md:127 이월 표시), N6(§6 추적표에서 M5 제거). N1·N2·N4·N7 과 update 쪽 잔여는 plan.md §K 로 이월. Tier 는 L 을 유지한다 — 측정 근거는 plan.md §A.1. 새 원장 L26~L33 은 HEAD `6f05252d6d` 에서 쟀다.
- 2026-10-27 — v0.2.1 plan 델타 4회차 (manager-spec, 리더 결정 기록 §15.102, `.moai/reports/leader-plan-20261023/status-recovery-20261025-2300.md` 886-891행). 확인 감사 5회차(FAIL 0.73, 감사 HEAD `274441db21`, 차단 B1·B2)를 받아 고쳤다. REQ·AC id 와 수는 그대로(23·24). B1: REQ-TCL-010 의 줄 예산을 "체크리스트가 그리는 줄(머리·환경·그룹 제목·항목·요약) ≤ 24, 그룹당 항목 ≤ 8" 로 다시 정의하고, 기존 직접 출력 줄(`init.go:708` slim 안내, `:750` 시작 줄, `:784` 사용자 자산 요약, `:1034` P5 안내)은 바꾸지 않고 따로 세며 M0 기준선 대비 증가 0 으로 정했다(§3.3). 깨끗한 init 정의에서 "보존 줄 0" 을 지우고 P5 를 넣었으며, 줄 단위를 한 번 정의했다(감사 M2). 화면 전체 줄 상한(24 고정 대 현실값)은 정하지 않았다 — t1978 과 같은 운영자 질문으로 열어 둔다(§5). B2: REQ-TCL-015 의 요약 없는 실행 조항 두 개(`nil` 반환 쓰기, 이미 나간 줄은 다시 쓰지 않음)는 init 전용이라 AC-TCL-017 (a1)·(c)·(e) 에 단언을 더하고 AC-TCL-017 의 `maps` 에 REQ-TCL-015 를 넣었으며, plan.md §K K-3 의 t1978 이월을 철회했다. 함께 고친 것: S1(update `warn`·`fail` 기준선 고정을 기존 시험 경로로 이름 붙임), S2(REQ-TCL-012 의 update 절을 세 고정 실행으로 좁힘), S3(AC-TCL-016 의 주입 지점, design.md §3.7), M1(살아 있는 것처럼 읽히던 update 문장), M3(AC-TCL-003·AC-TCL-004 를 init 계획표 밖 키로 그림). 새 원장 L34~L41 은 HEAD `274441db21` 에서 쟀다.
- 2026-10-27 — v0.2.1 보충, 버전 번호 그대로 (manager-spec, 리더 결정 기록 §15.104, 커밋 `eb2a9dfc21`). 6회차 감사 N-A 를 받아 acceptance.md AC-TCL-013 정규화 목록에 훅 백업 파일 시각 규칙 `\b\d{8}T\d{6}Z\b` → `<TS>`(`internal/cli/hook_install_shared.go:176`)와 그 사실 칸·뮤턴트를 더했다. REQ·AC id 와 수는 그대로(23·24).
- 2026-10-27 — v0.2.2 plan 델타 5회차 (manager-spec, 리더 결정 기록 §15.107, 커밋 `aee74a15132a4ca72c25df2441473d37418ef1d8`). 교차 모델 감사(감사 HEAD `eb2a9dfc21`)의 P2 결함 4건과 R-2 를 앵커 안에서 고쳤다. REQ·AC id 와 수는 그대로(23·24). P17 init 지점 retrack 경고를 보존 행으로 넣고 AC-TCL-024 에 `p17_init_manifest_retrack` 하위 테스트를 더했다(녹색 경로 25 → 26줄). `--no-hooks`·`!opts.MCPProvision` 조건부 생략 항목을 design.md §4 와 AC-TCL-001 명령 수준 절로 정했다. AC-TCL-008 카탈로그 비교를 (인자 번호, 동사) 목록 일치와 `fmt.Sprintf` 결과의 `%!` 부재로 바꿨다. AC-TCL-022 를 sync 감사의 최종 HEAD 재실행으로 판정하게 했다. AC-TCL-013 정규화 목록에 ISO 시각 규칙과 목록 밖 값 막힘 규칙을 더했다.
- 2026-10-27 — v0.2.3 plan 델타 6회차 (manager-spec, 리더 결정 기록 §15.108 — §15.107 의 예외: 8회차 차단 결함이 델타 5 가 남긴 잔여 문장이라 새 설계 결함이 아님). 8회차 plan 감사(`.moai/reports/t1969/plan-audit-verdict-iter8.md`, 감사 HEAD `aee74a1513`)의 차단 N8-1 을 고쳐 acceptance.md §F 완료 정의를 AC-TCL-022 본문(기록 + 최종 HEAD 재실행 모두 녹색)에 맞췄고, N8-3 으로 이 HISTORY 에 v0.2.2·v0.2.3 행을 더했다. 한 줄로 닫히는 선택 결함 N8-4·N8-5 도 고쳤다. REQ·AC id 와 수는 그대로(23·24). 나머지 선택 결함의 처분은 progress.md §E.1 델타 6 단락.
- 2026-10-27 — v0.2.4 plan 델타 7회차 (manager-spec, 리더 결정 기록 §15.111, 1회 승인·1커밋). codex P2 두 건(영수증 `rcpt-50fb41f52e6c83eab02ffa04`, 레인 실행 `rcpt-ac0ee402340ec744d7eb4a7a`, 둘 다 `eff053068d`)과 남은 선택 결함을 고쳤다. AC-TCL-008 의 서식 서명을 Go `fmt` 의 실제 인자 소비 규칙으로 정의하고 실제 `fmt.Sprintf` 대표 값 횟수 비교로 재게 했다. AC-TCL-016 비활성 대조 실행이 만들어지도록 시험 시접 `checklistSessionStartFn` 을 수집기 수준(활성·비활성 모두, 생산 값 `nil`)으로 옮겼다(design.md §3.7). 함께: N8-6(§6 추적표 AC-TCL-001 에 M4, plan.md M4 목록), N9-1(acceptance.md §F 의 AC-TCL-022 판정 문구), N9-2(위 v0.2.1 보충 행), N8-2(AC-TCL-013 ISO 시각 규칙의 구분자 `[:-]`). REQ·AC id 와 수는 그대로(23·24). 처분 상세는 progress.md §E.1 델타 7 단락.
- 2026-10-27 — v0.2.5 plan 델타 8회차 (manager-spec, 리더 결정 기록 §15.115, 1회 승인·1커밋). codex 가 AC-TCL-008 의 손으로 쓴 소비 규칙에서 셋째 `fmt` 문법 구석(번호 뒤 문자 그대로의 퍼센트, `%[1]s %[2]% %s` 와 `%[1]s %[2]% %[2]s` 가 둘 다 `A % B`)을 찾았다. 리더 판정대로 규칙을 다시 고치지 않고 지웠다 — AC-TCL-008 은 실제 Go `fmt` 실행만으로 판정한다(en 값에서 `%!` 없는 첫 기준 인자 조합을 `fmt` 로 찾고, 네 언어 출력의 `%!` 부재와 기준 인자별 출현 횟수를 비교), 고정 회귀 묶음을 단언한다. design.md §6 문구를 맞췄다. 함께 닫은 선택 결함: design.md §3.7 시접 부르는 곳의 둘째 예(수집기가 생기기 전에 도는 세션 시작 함수)를 수집기가 세션을 받는 지점으로 고침, AC-TCL-017 (c) 의 "세션 시작 때" → "수집기가 만들어진 직후", 해석기 대 `fmt` 단언(해석기가 없어져 성립하지 않음). REQ·AC id 와 수는 그대로(23·24), AC-TCL-008 의 뜻(네 언어 서식 안전)도 그대로. 처분 상세는 progress.md §E.1 델타 8 단락.
- 2026-10-27 — v0.2.6 plan 델타 9회차 (manager-spec, 리더 결정 기록 §15.119, 1회 승인·1커밋). 리더 audit_multi(영수증 `rcpt-5abf31cc8a860a40811fc6b9`, codex 필수, FAIL)의 P2 두 건 — 인자를 빠뜨린 번역이 v0.2.5 의 범용 판정 절차를 통과함(en `Project %.2s: %s` 대 ja `プロジェクト %[2]s`), 깨진 en 값(`broken %`)에서 절차의 탐색이 끝나지 않음 — 을 받아, 범용 판정 절차를 계약에서 빼고 AC-TCL-008 의 서식 조건을 고정 호출 지점 표(형식 키 × 네 언어 × 호출 지점의 실제 인자, 칸마다 기대 출력, `fmt.Sprintf` 바이트 비교)와 키 수 가드로 바꿨다. en 값과 en 기대 출력은 이 SPEC 이 정하고, ko·ja·zh 기대 출력은 M1 이 카탈로그 값과 같은 커밋에서 채운다. design.md §6 문구를 맞췄다. REQ·AC id 와 수는 그대로(23·24), AC-TCL-008 의 뜻과 서식 밖 조건도 그대로. 처분 상세는 progress.md §E.1 델타 9 단락.
- 2026-10-27 — v0.2.7 plan 범위 분할 (manager-spec, 리더 결정 기록 §15.123, 1회 실행·1커밋, 수리가 아닌 안전 고정). plan 감사 12회차(FAIL 0.90, 차단 D1 — 호출 지점 표의 ko·ja·zh 칸에 정확성 규칙이 없음)를 받아, 리더가 후속 카드 t1989(이 카드에 의존)를 내고 AC-TCL-008 과 그 번역 서식·일관성 검사, 호출 지점 표, D1~D5 를 그 카드로 넘겼다. 이 SPEC 에서 AC-TCL-008 의 제목·본문·뮤턴트·사실 칸·표·키 수 가드, design.md §6 의 서식·ja·zh ≠ en 문장, plan.md M1 의 AC-TCL-008 테스트 줄, §6 추적표 행, §3.5 카탈로그 규칙 줄을 지웠고, REQ-TCL-009 에서 AC-TCL-008 에만 대응하던 네 카탈로그 존재·ja·zh ≠ en 절을 걷어냈다(§3.1 끝의 이전 줄). REQ 23개 그대로, AC 24 → 23. 나머지 id 는 다시 매기지 않았다. 처분 상세는 progress.md §E.1 범위 분할 단락.
- 2026-10-27 — v0.2.8 plan 델타 10회차 (manager-spec, 리더 판정 — 감사 16회차 뒤 남은 결함은 codex P2 하나, 1회 실행·1커밋). AC-TCL-021 의 범위 검사가 `--no-merges` 라 병합 커밋이 들여온 템플릿 변경(충돌 해결·악성 병합)을 보지 못했다. 결합 diff 검사 `git log --merges --cc --name-only` 를 (2) 로 더해, 깨끗한 `main` 흡수는 빈 출력으로 두고 병합이 들여온 변경만 빨갛게 했다. 일부러 만든 저장소의 RED(옛 명령 빈 출력)·빨강(새 명령이 경로를 냄)·깨끗한 흡수 대조(빈 출력)를 사실 칸과 뮤턴트로, 이 트리의 기준선을 원장 L42·L43 으로 적었다. REQ·AC id 와 수는 그대로(23·23). 처분 상세는 progress.md §E.1 델타 10 단락.
- 2026-10-27 — v0.2.9 plan 델타 11회차 (manager-spec, 리더 판정 — AC-TCL-019 확장 승인, AC-TCL-021 과 같은 모양, 1회 실행·1커밋). AC-TCL-019 뒤 절의 위저드 검사가 `--no-merges` 라 병합 커밋이 들여온 위저드 변경(충돌 해결·악성 병합)을 보지 못했다. 결합 diff 검사 `git log --merges --cc --name-only … -- internal/cli/wizard` 를 (2) 로 더했다. 일부러 만든 저장소의 RED(옛 명령 빈 출력)·빨강(새 명령이 경로를 냄)·깨끗한 흡수 대조(빈 출력)를 사실 칸과 뮤턴트로, 이 트리의 기준선을 원장 L44·L45 로 적었다. REQ·AC id 와 수는 그대로(23·23). 처분 상세는 progress.md §E.1 델타 11 단락.
- 2026-10-27 — v0.2.10 리더 판정 반영 (manager-spec, 출처: run 종료 감사 1회차 `.moai/reports/t1969/sync-audit-verdict-iter1.md` F1 에 대한 리더 판정, 카드 t1969). 판정 원문: "AC-TCL-011's C counts only the lines the checklist draws; output that passes through with items unopened (user_asset_phase etc.) is counted in a separate bucket; the budget 24 is unchanged." §3.3 에 "따로 세는 넘김 줄" 정의를 더하고 체크리스트 셈에서 그 줄을 뺐다. acceptance.md AC-TCL-011 의 Given 을 고정 사용자 자산 픽스처(예산 셈)와 M0 고정(직접 출력 줄·골든 비교)으로 나누고, 넘김 줄 개수 단언과 "넘김 줄을 C 에 넣는 셈은 빨갛다" 는 뮤턴트를 적었다. 예산 24, 다섯 직접 출력 조각, REQ·AC id 와 수(23·23)는 그대로이며 구현(`internal/cli/checklist_wiring_test.go`)은 바꾸지 않았다.

## §1 배경과 문제

`moai init` 과 `moai update` 는 단계마다 진행 줄과 결과 줄을 따로 찍는다. 비TTY update 한 번이 stdout 에 60줄을 남기고(`internal/cli/testdata/tuxiu/postm4/update.notty.stdout.golden:1-60`), init 은 보고기 단계 줄 8개와 11줄짜리 완료 상자를 stderr 에 쌓는다(`internal/cli/testdata/tuxiu/postm4/init.notty.stderr.golden:1-24`). 무엇이 끝났고 무엇이 문제인지 한 화면에서 읽히지 않는다.

카드가 요구하는 모양은 macOS 정리 도구의 체크리스트다. 머리줄 1개, 환경 요약 1줄, 그룹 제목, 그룹 아래 항목마다 상태 한 줄. 실행 중인 항목만 제자리에서 바뀌고, 상세는 접혀 있다가 실패와 경고만 펼쳐진다. 끝에는 바뀐 수·보존한 수·경고 수·걸린 시간을 담은 요약이 붙는다.

재사용할 부품은 이미 있다. 줄 잔재 없이 제자리에서 다시 그리는 `ProgressLine`(SPEC-V3R6-UPDATE-PROGRESS-001), `StatusIcon`·글리프 상수, `Spinner`, `Theme`·모션 감소 판정, i18n 카탈로그(research.md §1). 새로 만드는 렌더러는 하나뿐이다.

**이 SPEC 의 범위(v0.2.0).** 렌더러와 `moai init` 연결까지다. `moai update` 연결(세션 생성, 템플릿 동기화 항목, update 요약, update 쪽 보존 줄, update 줄 예산)은 카드 t1978 이 맡는다. 이 카드의 줄 예산은 체크리스트가 그리는 줄에만 걸린다(REQ-TCL-010, §3.3). 화면 전체 — 기존 직접 출력 줄까지 합친 — 를 24줄로 고정할지 현실값으로 둘지는 init·update 공통의 운영자 질문으로 열려 있다(리더 결정 기록 §15.102, t1978 과 공유). 이 카드는 update 출력을 바꾸지 않는다(REQ-TCL-012).

## §2 확정한 결정

D-1·D-6·범위 밖 Q3 은 레인 판단(사다리 마지막 단계)으로 정했다. 되돌릴 수 있는 결정이며, Kickoff 기록에서 리더에게 알린다. 근거 기록은 `.moai/reports/t1969/progress.md` 의 결정 기록이다.

| ID | 결정 | 근거 |
|---|---|---|
| D-1 | 비TTY·CI·`--verbose`·`TERM=dumb` 에서는 렌더러를 켜지 않고 기준선 트리의 출력 경로를 그대로 탄다. 평문 체크리스트를 따로 두지 않는다. | 카드 문구 "TTY 아님/CI/--verbose 면 평문 줄 출력으로 폴백"을 "오늘의 평문 줄 출력으로 폴백"으로 읽었다. 오늘 비TTY 출력은 이스케이프 없는 평문 줄이라 문구를 문자 그대로 충족하고, 기존 버퍼 기반 테스트·CI 로그가 바이트 단위로 남는다. **다른 독해**: 카드가 `--verbose` 만 따로 "현재 상세 그대로"라고 적었으므로 비TTY·CI 를 평문 체크리스트(새 모양)로 바꾸라는 뜻일 수도 있다. 리더가 그 독해를 택하면 REQ-TCL-012·AC-TCL-013 과 M0 기준선을 다시 쓰는 델타 회차를 연다(plan.md §G). |
| D-2 | 그룹 표식은 `◆`, 건너뜀은 `○`. 카드 문구의 `➤`·`–` 는 쓰지 않는다. | `➤`(U+27A4)는 이모지 검사 범위, 둘 다 허용 목록 밖(design.md §1.1). Kickoff 기록에서 리더에게 알린다. |
| D-3 | 고루틴·티커 없이 이벤트가 올 때만 다시 그린다. | `Spinner` 의 무상태 계약. 동시성 위험을 들이지 않는다. |
| D-4 | 출력은 `cmd.SetOut`·`SetErr` 캡처로 모은다. 항목이 열린 동안의 바이트만 접고, 나머지는 즉시 넘긴다. 접히면 안 되는 줄(보존 줄 목록, design.md §3.3.1)은 따로 펼친다. | init.go 를 거의 건드리지 않고(plan.md §B) 정보 손실을 막는다. |
| D-5 | init 에 `--verbose` 를 새로 둔다. (update `--verbose` 가 대체 경로를 강제하는 효과는 t1978 이 더한다.) | init 에는 지금 플래그가 없다(research.md §2). |
| D-6 | 체크리스트가 켜지면 init 경고 패널과 완료 카드를 요약 블록이 대신한다. 대신하는 화면이 담던 데이터는 모두 요약에 옮긴다. 항목 밖에서 기록된 수집기 경고는 패널 대신 요약 블록이 다시 나열한다. (update 결과 알약을 요약이 대신하는 일은 t1978.) | 같은 경고를 항목 아래와 패널에 두 번 보이면 한 화면에 들어가지 않는다. 항목 밖 경고(예: 위저드 전 `init.go:427`)는 위저드 화면 위로 밀려나므로 요약이 다시 보여 준다. |
| D-7 | init 요약에는 "보존 수" 를 넣지 않는다. 렌더러 요약은 보존 수를 선택 필드로 지원한다(update 요약이 쓸 자리, t1978). | `InitResult` 에 보존 수 필드가 없다(research.md §2). |
| D-8 | 출력 채널은 바꾸지 않는다. init 은 stderr. (update 는 stdout — t1978.) | 상태 채널 규칙(design.md §5). |
| D-9 | 요약의 경고 수는 init 경고 수집기의 수(`warnCollector.Count()`, `internal/cli/init_warnings.go:59-63`)다. 심각도 줄(`emitSeverityLine`)과 보존 줄은 펼쳐 보이되 수에 더하지 않는다. | 기존 완료 카드도 같은 수를 보고한다(`init.go:1087` `p.Count()`). 수를 바꾸면 "보고하는 수를 바꾸지 않는다"(REQ-TCL-019)를 어긴다. |
| D-10 | 열린 항목은 다음 항목이 열릴 때, 요약이 그려질 때, 세션이 끝날 때, 그리고 단계 보고기가 실패를 알릴 때만 닫힌다. (update 요약을 그리는 지점 — 템플릿 동기화 본문 끝 — 은 결정 기록으로 design.md §4.1 에 남기고 t1978 이 다시 판단한다.) | 항목 경계를 네 경우로 고정해야 init 꼬리 단계의 출력이 항목 상세로 예측 가능하게 접힌다(design.md §3.2). |

## §3 요구사항 (GEARS)

### §3.1 렌더러

- **REQ-TCL-001** — The `internal/tui` package shall provide exactly one new group-checklist renderer that draws a header line, an environment-summary line, group titles, and exactly one status line per item, built from the existing `StatusIcon`/`Glyph*`, `Spinner`, `Theme`, and colour-downsampling primitives.
- **REQ-TCL-002** — **When** an item is running, the renderer shall redraw only that item's line in place, using the existing erase-line prefix, and shall never emit a redraw sequence for a line that already finished.
- **REQ-TCL-003** — **When** an item finishes, the renderer shall end its line with exactly one whitelisted status glyph: `✓` done, `!` done with warnings, `✗` failed, `○` skipped.
- **REQ-TCL-004** — **When** an item fails, the renderer shall expand beneath it, indented, every detail line recorded while that item was open, followed by the failure message.
- **REQ-TCL-005** — **When** a warning is recorded while an item is open, the renderer shall expand that warning beneath the item; the detail lines of an item that finished without warning or failure shall stay collapsed.
- **REQ-TCL-006** — The renderer shall end a completed run with a summary block carrying the changed count, the preserved count where the caller supplies one, the warning count defined by decision D-9, the elapsed time, and every data field of the completion surface it replaces.
- **REQ-TCL-007** — **While** `NO_COLOR` is set to a non-empty value, the renderer shall emit no SGR colour sequence.
- **REQ-TCL-008** — **While** `MOAI_REDUCED_MOTION` is set to a value other than empty or `0`, the renderer shall draw running items with the static run glyph and shall emit no Braille spinner frame.
- **REQ-TCL-009** — Every user-facing string the renderer draws shall come from a `checklist.*` catalogue key, and an unsupported language shall fall back to en.
- **REQ-TCL-010** — The init group/item plan shall hold at most 8 items per group; in a clean `moai init` run as defined in §3.3, the lines the checklist draws — header, environment line, group titles, item lines, and summary block — shall number at most 24, counted in the line unit §3.3 defines, and every direct-print line §3.3 lists as separately counted shall appear with unchanged text exactly as many times as in the M0 baseline run of the same fixture.

번역 문자열의 정확성 요구(네 카탈로그의 `checklist.*` 키 집합 동치, 값이 비어 있지 않고 키와 다름, 이모지 범위, ja·zh 값이 en 과 다름과 그 허용 목록, en 키 수, 서식 지시 키의 언어별 출력 일치)는 지운 것이 아니라 카드 t1989 로 넘겼다(리더 결정 기록 §15.123) — 넘기기 전 문구는 커밋 `f6a603fa62` 의 REQ-TCL-009 와 acceptance.md 8번 기준에 있다.

### §3.2 init 연결과 대체

- **REQ-TCL-011** — **While** the process stdout and stderr are both terminals, the command writers are those process streams, `CI` is empty, `TERM` is not `dumb`, and `--verbose` is off, `moai init` shall render the group checklist; the five inputs shall be gathered by one function that every command wiring the checklist calls.
- **REQ-TCL-012** — **When** any activation condition of REQ-TCL-011 does not hold, `moai init` shall emit the same bytes the baseline tree emits — the tree the fallback goldens were captured from (`cb30e05b3e`, or the re-baseline tree recorded under plan.md §B) — and, in each of the three fixed update runs of AC-TCL-013 (`clean`, `warn`, `fail`), `moai update`, which this SPEC does not wire, shall emit those baseline bytes.
- **REQ-TCL-013** — `moai init` shall accept `--verbose`, and `--verbose` shall force the REQ-TCL-012 output path.
- **REQ-TCL-014** — **While** the checklist is active, bytes written to the command writers while an item is open shall be recorded as that item's detail, and bytes written while no item is open shall reach the real writer immediately and unmodified.
- **REQ-TCL-015** — **While** the checklist is active, every warning recorded through the init warning collector, and every line that `moai init` emits at warn or error severity through the severity-line helper on either command writer (stdout or stderr), while an item is open shall appear exactly once, expanded under that item; every collector warning recorded while no item is open shall keep its legacy immediate line, if it had one, and shall be listed once in the summary block, or, when the run ends without drawing the summary, shall be written once to the real writer before the command writers are restored unless its text has already reached the real writer; and the consolidated init warning panel shall not be emitted.
- **REQ-TCL-016** — **When** a checklist run ends before the summary with a non-nil error or a wizard cancel, the session shall mark the open item failed, write every captured byte not yet shown verbatim to the real writer, write the collector warnings REQ-TCL-015 assigns to a run without a summary, and restore the command writers.
- **REQ-TCL-017** — The checklist shall render on stderr for `moai init`.
- **REQ-TCL-018** — **While** an interactive prompt runs, the checklist session shall hold no unfinished line, and the init wizard flow, the survey questions, and their content shall stay byte-identical to the baseline tree.
- **REQ-TCL-019** — The deployed file set, every reported count, and every command outcome of `moai init` shall equal those of the legacy surface for the same run, and the summary's warning count shall equal the collector count the legacy completion card reports.
- **REQ-TCL-020** — Every file under `internal/template/templates/` shall stay byte-identical to the baseline tree.
- **REQ-TCL-021** — This card's commits shall change `internal/cli/init.go` only inside the regions R1, R2, and R3 of plan.md §B, measured against the region base recorded there, and R3 changes shall be pure insertions.
- **REQ-TCL-022** — The preserved-line inventory of design.md §3.3.1 shall classify every hit of its fixed source sweep over `internal/cli` and its subpackages, and **while** the checklist is active in `moai init`, every line of an inventoried preserved row — direct-write warnings, error lines, and user-action notices that bypass the warning signal — shall appear exactly once on the real writer even when its enclosing item finishes successfully; a preserved warning or error line written while an item is open shall end that item with `!`, and a preserved notice shall change neither the item glyph nor the warning count.
- **REQ-TCL-023** — **When** a checklist run returns early with a nil error, the session shall close the open item, if one is open, as done, write every captured preserved line not yet shown, leave every finished line — a failed one included — as drawn, and restore the command writers.

### §3.3 깨끗한 실행의 정의 (REQ-TCL-010)

- **깨끗한 init**: 빈 임시 디렉터리에 처음 하는 init, 이전 `.moai` 없음, git 있음, 사용자 홈이 풀림(시험은 임시 홈), 기본 배포(`--all` 없음 — slim), 세션 워크트리 꺼짐, 지연 갱신 확인이 "새 버전 없음", 모든 꼬리 단계 성공, 수집기 경고 0. 보존 행 가운데 나오는 것은 P5(워크트리 안내, `init.go:1034`) 하나뿐이다 — P5 는 init 마다 조건 없이 나오므로 깨끗한 실행에서도 빠지지 않는다(v0.2.0 의 "보존 줄 0" 은 만족할 수 없는 조건이라 지웠다, 확인 감사 5회차 B1). M0 의 `fallback/init-clean-*` 골든과 M4 의 `rich/init-clean.golden` 은 이 같은 고정으로 찍는다.
- **줄 단위**(한 번만 정의한다, 확인 감사 5회차 M2): 두 작성기 출력에서 `\r\x1b[2K` 로 덮인 줄을 최종 한 줄로 접고 ANSI 시퀀스를 지운 뒤 `\n` 으로 나눈 줄 하나. 터미널 폭 때문에 화면에서 여러 행으로 접히는 줄도 한 줄로 센다. REQ-TCL-010 과 AC-TCL-011 은 이 단위만 쓴다.
- **체크리스트가 그리는 줄**(상한 24): 머리줄, 환경 줄, 그룹 제목, 항목 줄, 요약 블록(그 안의 빈 줄 포함). 아래 직접 출력 줄에도, 바로 다음의 넘김 줄에도 들지 않는 다른 모든 줄은 이 셈에 넣는다 — 새어 나온 상세 줄이 예산을 넘겨 드러나도록.
- **따로 세는 넘김 줄**(v0.2.10, 리더 판정): 체크리스트가 그리지 않고 항목을 열기 전에 명령 작성기로 그대로 지나가는 출력 — 사용자 자산 보장(`user_asset_phase.go` `writeInstallSummary`)이 Codex 쪽 매핑이 없는 자산마다 쓰는 줄(고정 조각 `unconverted reference (no Codex-side mapping`). 이 줄은 체크리스트 셈에 넣지 않고 셋째 묶음으로 따로 세며, 그 개수는 실행이 만든 넘김 줄 수와 같아야 한다. 줄 수가 임베드 카탈로그에 따라 달라지므로 예산 셈은 고정 사용자 자산 픽스처로 잰다(acceptance.md AC-TCL-011). 예산 24 는 그대로다.
- **따로 세는 직접 출력 줄**(텍스트를 바꾸지 않음, M0 기준선과 같은 횟수 — 증가 0): 아래 다섯 줄. 조각은 줄을 가려내는 고정 부분 문자열이며, 출처 행은 HEAD `274441db21` 에서 읽었다(acceptance.md 원장 L34·L35).

| 출처 | 쓰는 줄 (조각) | 작성기 | 깨끗한 init 의 줄 수 |
|---|---|---|---|
| `init.go:708` `emitSlimModeNotice` (`init.go:42-48`) | `Deploying core templates only (slim mode)`, `Common skills and agents install into your user folders` | stdout | 2 |
| `init.go:750` `p.Info` | `Initializing MoAI project...` | stderr | 1 |
| `init.go:784` `ensureUserAssetsLockedFn` → `user_asset_phase.go:54`·`:87` `writeInstallSummary` | `user-asset ensure: ` (충돌·분기 행은 깨끗한 실행에서 0) | stdout | 1 |
| `init.go:1034` `emitWorktreeAdvisory` → `worktree_advisory.go:46`·`:57` (보존 행 P5) | `this checkout is shared across concurrent sessions;` | stdout | 1 |

- **셈(코드 읽기, HEAD `274441db21`)**: 체크리스트가 그리는 줄 23(design.md §1 화면), 따로 세는 직접 출력 줄 5 — 합계 28줄(넘김 줄은 이 합계 밖에서 따로 센다). 체크리스트 셈은 23 ≤ 24 이고, 직접 출력 셈은 기준선과 같아야 한다(증가 0, 감소도 정보 손실이라 허용하지 않는다). `init.go:750` 을 항목 안으로 옮겨 줄을 줄이는 길은 R1·R2·R3 밖 수정이라 AC-TCL-022 가 막는다. 실측은 M0(기준선의 직접 출력 줄)과 M4(리치 골든)가 처음 한다.
- **화면 전체 줄 상한은 정하지 않는다.** 직접 출력 줄까지 합친 화면 전체를 24줄로 고정할지, 현실값(지금 셈으로 28줄, 80칸 터미널에서는 긴 줄이 여러 행으로 접혀 더 많다)으로 둘지는 운영자 제품 질문이며 t1978 과 같은 질문이다(리더 결정 기록 §15.102). 이 SPEC 은 "체크리스트 줄 ≤ 24, 직접 출력 줄 증가 0" 으로 닫는다.
- 깨끗한 update 의 정의와 그 줄 예산은 이 SPEC 에서 정하지 않는다 — [DEFERRED → t1978]. 델타 4회차 감사 N1 이 지금 정의(같은 바이너리로 init 한 프로젝트, `--force` 없음)로는 템플릿 동기화가 `internal/cli/update_template_sync.go:1260` 에서 건너뛰어 체크리스트가 그려지지 않고, 동기화가 도는 실행은 종료 행동 블록과 사용자 자산 요약 때문에 24줄을 넘는다고 보였다(design.md §4.1).

### §3.4 t1978 로 넘긴 요구사항 조각 [DEFERRED → t1978]

아래는 v0.1.3 의 REQ 문장에서 걷어낸 update 쪽 조각이다. 이 SPEC 은 이 조각들을 판정하지 않는다. t1978 의 SPEC 이 자기 REQ 로 다시 쓴다. 이월 항목 전체 표는 plan.md §K.

| # | 걷어낸 조각 | 걷어낸 곳 |
|---|---|---|
| F-1 | 깨끗한 update 실행의 24줄 상한과 update 계획표의 그룹 상한 | REQ-TCL-010 |
| F-2 | `moai update` 의 템플릿 동기화 경로에서 체크리스트를 그리는 일 | REQ-TCL-011 |
| F-3 | update 의 활성 조건이 하나라도 거짓일 때의 대체 경로(지금은 update 가 연결되지 않아 늘 기준선 바이트) | REQ-TCL-012 |
| F-4 | `moai update --verbose` 가 대체 경로를 강제하는 일 | REQ-TCL-013 |
| F-5 | update 심각도 줄(훅 설치, 마이그레이션, 보존 정리, 스냅숏 훅)의 항목 아래 펼침 | REQ-TCL-015 |
| F-6 | update 체크리스트의 stdout 채널 | REQ-TCL-017 |
| F-7 | update 병합 확인 중 열린 줄 없음 | REQ-TCL-018 |
| F-8 | update 의 갱신 파일 수·보존 수·경고 수 동치 | REQ-TCL-019 |
| F-9 | update 쪽 보존 행(design.md §3.3.1 의 P6~P19, P22, P20·P21 의 update 호출 지점) | REQ-TCL-022 |
| F-10 | (v0.2.1, 확인 감사 5회차 S2) update 출력 바이트 보존을 AC-TCL-013 의 세 고정 실행 밖으로 넓히는 일 — 세 실행이 닿지 않는 공유 파일의 실패 경로 줄(예: `hook_install.go:283`·`:291`·`:293`, `update/backup/settings_snapshot.go:115` `warnLine`)은 이 SPEC 이 지키지 않는다. update 를 연결하면 TTY update 에서 REQ-TCL-012 의 update 절이 더는 성립하지 않으므로 t1978 이 그 절을 고친다 | REQ-TCL-012 |

### §3.5 t1978 이 기댈 수 있는 인터페이스

이 SPEC 이 착지하면 t1978 은 아래를 그대로 쓴다. 이 계약은 이 SPEC 의 REQ 와 AC 가 지키므로, t1978 은 다시 증명하지 않고 연결만 더한다.

- **렌더러 계약**(REQ-TCL-001 부터 REQ-TCL-009 까지, AC-TCL-001 부터 AC-TCL-007 까지, AC-TCL-009, AC-TCL-023): `NewChecklist`·`Header`·`Item`·`Update`·`Detail`·`Warn`·`Keep`·`Done`·`Fail`·`Skip`·`Summary`(design.md §2). 계획표와 그룹·항목 키는 호출자가 넘긴다 — 렌더러에 init 전용 지식이 없다. 요약은 보존 수·백업 경로·복구 명령을 선택 필드로 받는다(AC-TCL-005 가 update 모양 요약으로 검사).
- **이벤트 모델**: 한 번에 열린 항목 하나, 새 `Item` 이 열린 항목을 암묵 `Done` 으로 닫음, 종료 전환 두 번은 panic, 이벤트가 올 때만 다시 그림(결정 D-3).
- **활성 판정**(REQ-TCL-011): 다섯 입력을 모으는 한 함수 `checklistEnv(cmd, verbose)` 와 시접 세 개(TTY 탐지, 프로세스 스트림, 강제 켜기 — design.md §3.1). update 는 같은 함수를 부르기만 하면 된다.
- **세션**(REQ-TCL-014, REQ-TCL-015, REQ-TCL-016, REQ-TCL-023): 캡처 작성기, 두 명령 작성기 모두의 경고 수신 인터페이스, 보존 줄 알림, 종료 처리(오류·취소·`nil` 조기 반환). REQ-TCL-015 의 요약 없는 실행 조항 두 개 — 요약 없이 끝나면 항목 밖 수집기 경고를 한 번 쓴다, 이미 실제 작성기에 나간 경고는 다시 쓰지 않는다 — 는 init 전용 조항이며 AC-TCL-017 (a1)·(a2)·(c)·(e) 가 판정한다(v0.2.0 의 t1978 이월은 리더 결정 기록 §15.102 로 철회, plan.md §K K-3). 세션은 init 전용 코드가 아니다 — 명령이 세션을 만들지 않으면 아무 일도 하지 않는다.
- **골든 전략**(design.md §7·§8): 기준선 선행 커밋(M0) 규칙과 렌더러 골든의 `MANIFEST.txt` 규칙. update 대체 경로 기준선 골든 6개(`fallback/update-{clean,warn,fail}-{stdout,stderr}.golden`)는 이 카드의 M0 이 이미 찍는다 — t1978 은 그 골든을 update 연결 전의 기준선으로 쓴다.
- **보존 줄 목록**(design.md §3.3.1): 이 카드에서 X11 로 분류해 둔 update 행을 t1978 이 P 행으로 되돌린다. 분류 조각과 행 id 는 그대로 둔다.

## §4 제약

- C-1 렌더러 파일은 `internal/tui/checklist.go` 하나다. 다른 tui 파일은 재사용만 하고 고치지 않는다. 예외: 카탈로그 yaml 네 개.
- C-2 init.go 변경 영역은 REQ-TCL-021 로 올렸다. t1966 이 바꾸는 줄과 맞닿은 줄(728, 750, 780-790)은 건드리지 않는다.
- C-3 `.github/workflows/` 를 고치지 않는다.
- C-4 새 코드의 사용자 대상 문자열은 카탈로그 키를 거친다(`internal/tui/i18n.go:81-82` 규칙).
- C-5 테스트 시접을 바꾸는 테스트는 `t.Parallel` 을 쓰지 않고 `t.Cleanup` 으로 원래 값을 되돌린다.
- C-6 대체 골든과 그 테스트는 코드 변경보다 먼저, 그 둘만 담은 커밋 하나(M0)로 들어간다(design.md §8, AC-TCL-013).
- C-7 이 카드는 update 연결 코드(update 세션 생성, 템플릿 동기화 항목 삽입, update 요약)를 넣지 않는다. init 과 update 가 함께 쓰는 파일(예: `hook_install.go`, `internal/cli/update/backup/file_snapshot.go`, `severity_line.go`)을 고칠 때는 세션이 없을 때 출력이 그대로인 형태로만 고친다 — AC-TCL-013 의 `update_*` 하위 테스트가 그 세 고정 실행이 닿는 줄에 한해 지킨다(닿지 않는 실패 경로 줄은 §3.4 F-10).

## §5 범위 밖

### Out of Scope — update 연결 (카드 t1978)

- `moai update` 의 체크리스트 연결 전부: 세션 생성, 템플릿 동기화 항목, update 요약과 그 위치, update 쪽 보존 줄, update `--verbose` 의 대체 강제, update 출력 채널, 병합 확인 경계
- 깨끗한 update 의 정의와 줄 예산(24줄 고정인지 "한 화면"인지 — 운영자 질문, t1978 에 열림)
- 델타 4회차 감사 N1·N2·N4·N7 과 update 쪽 잔여(plan.md §K)

### Out of Scope — 다른 명령과 경로

- `moai doctor`, `moai status`, `moai bundle add` 등 init·update 밖의 명령 출력
- update 의 clean-install 경로, 바이너리 갱신, `--check`, `--shell-env`, `--config`(reconfigure) 경로의 항목화
- init 의 사용자 자산 보장 단계(`init.go:780-790`) 항목화 — 레인 결정 Q3: t1966 이 `:775-780` 을 바꾸므로 그 카드 착지 뒤 후속 카드로 넘긴다. 이 단계는 `executor.Execute`(`init.go:829`)보다 먼저 돌아 열린 항목이 없으므로 출력이 넘김 규칙으로 지금 그대로 보인다

### Out of Scope — 화면 전체 줄 상한 (운영자 질문, t1978 과 공유)

- 직접 출력 줄까지 합친 init 화면 전체의 줄 상한(24줄 고정인지 현실값인지) — 리더 결정 기록 §15.102 가 정하지 않고 운영자 제품 질문으로 남겼다. 이 SPEC 은 체크리스트가 그리는 줄 ≤ 24 와 직접 출력 줄 증가 0 으로 닫는다(§3.3)
- 기존 직접 출력 줄(`init.go:708`, `:750`, `:784`, `:1034`)의 문구 축약·이동·항목화

### Out of Scope — 위저드와 질문

- init 위저드, update 병합 확인, 프로필 위저드의 흐름·질문·문구·스타일
- 대화형 입력 화면을 체크리스트 안에 넣는 일

### Out of Scope — 출력 형식 확장

- 비TTY 용 평문 체크리스트(결정 D-1, 레인 결정 Q1. 다른 독해를 택하면 델타 회차로 다시 연다)
- JSON 출력 모드, 상세 로그 파일 저장
- 애니메이션 티커, 진행 막대, 남은 시간 추정
- 경고 수 정의 변경(심각도 줄·보존 줄을 경고 수에 더하는 일, 결정 D-9)
- ja·zh 문구의 원어민 검수(초안은 이 카드가 넣는다)
- en·ko·ja·zh 밖 11개 언어 카탈로그

### Out of Scope — 배포·워크플로

- `internal/template/templates/` 아래 배포 템플릿
- `.github/workflows/` 와 CI 설정

## §6 추적표

AC 와 REQ 의 대응 선언은 acceptance.md 각 기준 제목의 `maps` 목록이 원본이다. 이 표는 마일스톤을 함께 보이기 위한 사본이다. M5(update 연결)는 t1978 로 넘어가 어느 행에도 없다(델타 4회차 감사 N6 해소).

| AC | 대응 REQ | 마일스톤 |
|---|---|---|
| AC-TCL-001 | REQ-TCL-001, REQ-TCL-003 | M2, M4 |
| AC-TCL-002 | REQ-TCL-002 | M2 |
| AC-TCL-003 | REQ-TCL-004 | M2 |
| AC-TCL-004 | REQ-TCL-005 | M2 |
| AC-TCL-005 | REQ-TCL-006 | M2 |
| AC-TCL-006 | REQ-TCL-007 | M2 |
| AC-TCL-007 | REQ-TCL-008 | M2 |
| AC-TCL-009 | REQ-TCL-009 | M2 |
| AC-TCL-010 | REQ-TCL-010 | M3 |
| AC-TCL-011 | REQ-TCL-010, REQ-TCL-009 | M4 |
| AC-TCL-012 | REQ-TCL-011 | M3, M4 |
| AC-TCL-013 | REQ-TCL-012 | M0, M4 |
| AC-TCL-014 | REQ-TCL-013 | M3, M4 |
| AC-TCL-015 | REQ-TCL-014 | M3 |
| AC-TCL-016 | REQ-TCL-015 | M4 |
| AC-TCL-017 | REQ-TCL-015, REQ-TCL-016, REQ-TCL-023 | M3, M4 |
| AC-TCL-018 | REQ-TCL-017 | M4 |
| AC-TCL-019 | REQ-TCL-018 | M4 |
| AC-TCL-020 | REQ-TCL-019 | M4 |
| AC-TCL-021 | REQ-TCL-020 | 전 구간 |
| AC-TCL-022 | REQ-TCL-021 | M4 |
| AC-TCL-023 | REQ-TCL-001, REQ-TCL-003, REQ-TCL-009 | M2 |
| AC-TCL-024 | REQ-TCL-022 | M4 |
