---
id: SPEC-RELUP-DUALAXIS-001
title: "release-update 하네스 CC+Codex 이중 축 정착 — codex 체인지로그 축·상태 파일 codex 키·BP 상시 절차"
version: "0.11.0"
status: in-progress
created: 2026-10-09
updated: 2026-10-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/agents/harness/hns-release-update-specialist.md, .claude/workflows/hns-release-update-run.js, .claude/commands/harness/release-update/manifest.json, .moai/state (harness-layer writer, machine-local)"
lifecycle: spec-anchored
tags: "release-update, codex, dual-axis, harness, best-practices, state-schema, dev-only"
tier: M
era: V3R6
related_specs: [SPEC-UPDATE-ADD-CODEX-001, SPEC-CC2219-UPSTREAM-ALIGN-001]
---

# SPEC: release-update 하네스 CC+Codex 이중 축 정착 — codex 체인지로그 축·상태 파일 codex 키·BP 상시 절차

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-09 | manager-spec | 최초 작성 — 카드 t1579 (High·운영자 확장 지시 2026-10-07·builder-harness/SPEC 소관). 워크트리 t1579 @ `2aab5f797` 실측 13종(앵커 grep + exit code 전수 관측)을 근거로 REQ 14건·AC 13건(릴리스 블로킹 10 + 회귀 가드 3) 확정. 근거 연구: `.moai/research/upstream-update-20261007.md` (확장 스윕 1차) + `upstream-update-20261008.md` (2차). SPEC ID 사전 검증: `SPEC-RELUP-DUALAXIS-001` 정규식 `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` 매치 **PASS** (Bash 실행 관측) |
| 0.1.1 | 2026-10-09 | manager-spec | 자체 spec-lint 피드백 수리 — `CoverageIncomplete` 3건(REQ-RDX-004/009/013 미커버)에 AC-RDX-014/015/016 신설(블로킹 10→13, 총 AC 13→16 — Tier M 천장 16에 정확히 도달). AC-RDX-016의 `source-first` 앵커로 mutant M-4의 기계 판정면 확보(LED-014, 0/exit 1 실측). lint 재실행: 0 error / 0 warning |
| 0.2.0 | 2026-10-09 | manager-spec | plan-audit iter1 수리(FAIL 0.8125 — codex required 게이트, `.moai/reports/t1579/plan-audit.md`, rcpt-3e881c87c91a47a2117220b0). **CX-4(블로커)**: plan §E3을 런타임-형태 어댑터로 교체 — require·직접 실행 모두 `SyntaxError: Illegal return statement`(run.js:153, Node v22.14.0 본 재관측 exit 1), 러너는 ESM `export const meta`+top-level return/await 하이브리드라 export 적출+AsyncFunction 래핑 어댑터가 유일한 충실 평가형 — E3-P1 본 트리 **exit 0 관측**(`adapter-ok run=fn cc=1 shape-ok`), E3-P2(M2 종료형)는 codex 단언 포함. **CX-1**: plan §C 시드 재판정을 last-analyzed 의미론으로 재작성(D1·§7 전파 — 미분석 승격은 시드 미인상). **CX-3**: LED-001/002를 `domain` 필드 스코프로 재앵커 + AC-RDX-013에 sprint_contract 판독면 신설(LED-015 기준선 출력 기록, exit 0 — AC 수 16 불변). **CX-2**: AC-RDX-003을 `selectCodexSweepTargets(args)` ≥2(정의+top-level 디스패치 호출)로 재앵커 + §E3-P2 실질 생성 면(M2에 export 목록 확장 요건 추가). **D5**: §3 한계를 AC-RDX-014/015로 확장(iter2 블록 재판독이 제2 판정면). **D6**: frontmatter version 0.2.0 — 최신 HISTORY 행 정합 유지(0.1.1 지적의 재발 방지). **D7 처분**: progress §F 선기입은 라인 지시에 의한 것 — 오케스트레이터가 Phase 4에서 확정/수정(내용 수정 불요). **D8 처분**: 운영자 ①구절의 리포 참조 버전 비교 반쪽은 CARD-4 소관 기록 유지 — run-phase 위임 프롬프트가 M1/M2에 참조-버전 문맥을 운반할 것. 미 touched AC의 RED 셀은 바이트 불변 유지 |
| 0.3.0 | 2026-10-09 | manager-spec | plan-audit iter2 수리(FAIL 0.875 — Testability 0.75, iter1 4건 수리는 전수 재검증 통과). **CX-5**: AC-RDX-003에 병합 관측면 신설 — plan §E3-P3 모의-런타임 실행 동사(mock agent/parallel로 러너 top-level 블록 실행, 시드 codexDeltas 주입)가 병합·agent 호출까지 실측; LED-016 RED 본 트리 관측(stderr `REJECTED: no-codex-dispatch:1`, exit 1 — 현재 러너는 CC 호출 1건에 codex 라벨 0건), M2 GREEN 기대 `dispatch-ok codex=1 total=2`/exit 0. §D1에 병합 변수 `allTargets`+`parallel(allTargets` 호출식+codex 라벨 접두사 `codex-release-notes:` 핀, §3 M-6 신설. **CX-6**: AC-RDX-006을 이중 사이트 기준으로 확장 — LED-006 GREEN 문턱 ≥1→≥2(Phase 0 판독·기본값 + Phase 7a 기록) + LED-017 기록 단계 리터럴 `7a-codex` 신설(RED 0/exit 1 본 트리 관측), §3 M-7 신설 — REQ-RDX-003의 기계 면 확보, AC 수 16 불변. **CN-4 처분**: 본 SPEC 산출물은 awk 계열 정렬-검증 동사를 인용하지 않는다 — 전부 단일 grep·node -e·python3 -c(가드 통과형). 미 touched AC의 RED 셀 바이트 불변 |
| 0.4.0 | 2026-10-09 | manager-spec | plan-audit iter3 수리 + ceiling STOP 처분(FAIL — 점수 0.875→0.8125 회귀, max_iterations 3 도달. `harness.plan_audit_ceiling_policy` STOP 경로: 점수 회귀 시 무조건 반복 금지 → 범위 축소 + 분할 제안. 접수 대장 rcpt-3e881c87c91a47a2117220b0(iter1) · rcpt-88fb1bd8fa1246aab81b1a0d(iter2) · rcpt-28b141979bf08b6a367fe557(iter3)). **CX-9(설계 — 당면 수리)**: 스페셜리스트 Phase 2의 CC-null 조기 종료("If no entries … stop.")가 이중 축 설계에 생존 — CC 빈 주간에 codex/BP 축이 실행 전 종료. REQ-RDX-015 신설(축별 종료 — CC 널 델타는 CC 축만 중단, codex·BP 축은 같은 run에서 실행·기록, 미실행 축이 남으면 완료 요약 금지) + plan M1/M2에 축별 종료 인코딩 핀 + AC-RDX-017 신설(LED-018 `only the CC axis` → 0/exit 1 본 트리 관측 — 단일 호출 관측 가능해 릴리스 블로킹 유지), §3 M-8 신설. **CX-7/CX-8(계측 경화 — 이관)**: AC-RDX-003/004/005(구조 면만 측정 — 빈 상수·"Return ok." 프롬프트 통과)와 AC-RDX-006(산개 언급 mutant 통과)을 verification-completeness §2 채택 기준(너무 얕아 채택 불가 — mutant-probe adoption bar; §2.1 처분군 적용 — fresh run에서 인용 정밀도 정정)에 따라 회귀 가드(판정 보류)로 강등 — 더 깊은 계측기 제작 추격 금지(감사 지시 treadmill 중단), 내용·사이트 계측은 §7의 형제 카드 제안으로 이관. 미 touched AC의 RED 셀 바이트 불변. frontmatter version 0.4.0 동기화 |
| 0.5.0 | 2026-10-09 | manager-spec | fresh post-split run 수리(FAIL 0.875 — 잔여 2 차단+1 자문 완전 열거, 영수증 rcpt-3cee4da84bc0f2817fc60f76. 재분류 실질 정당·CX-9 설계 正確(LED-018 독립 재실행 정상 RED)·천장 +1 정책 순응 수용 — 단 **분할 실제 발행 추적**: 라인이 §7의 형제 카드 제안을 리더에 전달하는 것으로 충족). **CX-10(차단)**: AC-RDX-017에 제거면 신설 — LED-019가 구형 무조건 종료 문장 전문(`If no entries: emit "No new versions since vX.Y.Z" and stop`)을 계수, 현재 1/exit 0(생존)→M1 후 0/exit 1이 PASS — 주석 포함 생존 전부 적색(LED-018 단독의 주석 mutant 허점 봉쇄), §3 M-8에 병기. **CX-11(차단)**: plan M2-6에 `run()` 공개 export 동기화 핀(동일 allTargets 병합 경로) + §E3-P4 verb 신설 — run() 공개 경로를 codex 전용 입력으로 직접 호출, LED-020 RED 본 트리 관측(stderr `REJECTED: no-codex-in-run:0`, exit 1), M2 GREEN 기대 `run-ok codex=1 total=1`/exit 0. **자문(인용 정밀도)**: 재분류 근거 인용을 §2 채택 기준 1차 + §2.1 처분군으로 4곳 정정(acceptance §A·§D.5, spec §7·HISTORY 0.4.0). 미 touched AC의 RED 셀 바이트 불변. frontmatter version 0.5.0 동기화 |
| 0.6.0 | 2026-10-09 | manager-spec | fresh-run iter2 말단 처분(FAIL — 잔여 1 차단, 직전 수리 3건(LED-019 제거면·E3-P4 run()·인용 정밀도)는 감사 자체 재실행으로 전수 검증 통과, 영수증 rcpt-8ccb4c8ea2798a39261f927d). **CX-12**: LED-018/019 쌍은 paraphrase 우회 가능 — 다른 무조건 종료 문구가 쌍을 통과함이 codex 실증. 유한 리터럴 집합은 의미론을 보증하지 못하므로 AC-RDX-017의 기계 쌍을 회귀 가드(판정 보류)로 강등(verification-completeness §2 채택 기준 — paraphrase mutant가 요구를 위반하면서 쌍을 통과 → 채택 불가; §2.1 처분군 적용층). RED 셀은 측정 사실로 보존, LED-021 미신설(감사 처문 준수). **REQ-RDX-015 규범 유지** — 판정면을 plan §E7 검토면(run-exit E1 인간 검토 + run-phase 위임 프롬프트 운반 의무)으로 이관. 형제 카드 제안(§7)이 축별 종료 판별기를 흡수. **D17**: DoD REQ 범위 001..014→001..015. 미 touched AC의 RED 셀 바이트 불변. frontmatter version 0.6.0 동기화 |
| 0.7.0 | 2026-10-09 | manager-spec | 리더 재개 재심 3건 수리(말단 PASS-WITH-DEBT는 영수증 대조로 기각 — 재개 재심 rcpt-0e776d398b9dd8d642d5f3ac의 CX-13/14/15, 감사 처방 그대로 채택). **CX-13**: AC-RDX-016의 `source-first` 기계 면이 반전 가능(리터럴 유지·규칙 역전 — CX-12 동류) → 회귀 가드(판정 보류)로 강등, REQ-RDX-013 의미론 면은 plan §E7 검토면에 항목 (d)로 합류 + §7 형제 카드 판별기 확장. **CX-14(블로킹 유지 — 견고한 계측)**: AC-RDX-009에 제거면 신설 — LED-021 `docs.anthropic.com` 계수, 본 트리 1/exit 0(구형 URL 생존 실측)→M3 후 0/exit 1이 PASS(주변 서술 언급만으로 통과하는 mutant 봉쇄). **CX-15**: DoD 변경 스코프 단언 재정식 — 병합 베이스 고정(plan 시점 관측 `2aab5f797b75983e132af451da68f69e3426557b`)+허용 경로 열거(하네스 3표면+SPEC 디렉터리)+커밋/스테이지/비추적 3채널 개별 관측(bare git diff 공집합은 부재다). 블로킹 9→8, 가드(판정 보류) 5→6. 미 touched AC의 RED 셀 바이트 불변. frontmatter version 0.7.0 동기화 |
| 0.8.0 | 2026-10-10 | manager-spec | 클로저 감사 수리 — plan-audit FAIL(overall 0.625, MP-8·MP-9 실패, 차단 8; audited_sha `d36e97571`, 영수증 `rcpt-3325997180ccbf593d94461c`). 기준선: 감사 대상 d36e97571(커밋 2026-10-10T01:20:28+09:00), RED 귀속 트리 2aab5f797. **CX-16·CX-17**: d36e97571에서 착지 확인(감사 §1 CLOSED — DoD 항목 4의 비스테이지 추적 편집 채널, 연구 입력 2건의 허용 경로). **CX-18(B-01·B-03·B-04)**: LED-021 RED를 2aab5f797에 핀(`git grep -c -h` = 6, exit 0)하고 착지 목표(0 / exit 1)를 별도 셀로 분리; 육면 검사를 이스케이프 점 + 종결 경계 패턴으로 교체(음성 대조 `codeXclaudeXcom/docs/en/hooks` → 0, 양성 대조 → 1); LED-022..027을 핀 RED(출력 없음)와 착지 GREEN(`1`, d36e97571)으로 분리하고, '바이트 동일' 주장을 `git diff --name-only 2aab5f797… d36e97571` 관측(9 경로)으로 정정. **B-02(MP-9, 결정 기록 option (a))**: 구형 `docs.anthropic.com` 제거를 M3 항목 5로 편입하고 식별자 없는 후속 델타 문구를 삭제 — AC-RDX-009 게이트와 DoD 항목 1은 유지. **B-05**: 본 행 신설, frontmatter `updated` 2026-10-10. **B-06**: 집계 정정 — RED-now 앵커 grep 15 + 검증 동사 2, 가드 판정 보류 6 + 회귀 3, 블로킹 8 (plan §C·§E1·§E2 정정). **B-07**: REQ-RDX-001 키군 절을 블로킹 AC-RDX-014의 블록 스코프 Then에 편입(신규 AC 없음, 블로킹 집계 유지); 블로킹 AC가 없는 나머지 8종(003·005·006·007·008·011·013·015)은 강등 처분, 소유자 unassigned, leader to issue (acceptance §D.2 게이팅 처분 표). **B-08**: AC-RDX-014/015 게이트를 블록 스코프로 교체(경계: specialist 71–94행, runner 79–92행; acceptance §D.3-c). 상태는 in-progress 유지 |
| 0.9.0 | 2026-10-10 | manager-spec | 분할 수리 — 운영자 결정 `d-20261010T000832Z-efde` (레인 제안 (나) 분할, 리더 전달; 줄인 범위 감사 1회, 재실패 시 정지). **기계 4건**: F5 — E7 실패 조건 추가 (E7(a)–(d) 각각 긍정 판단 기록 요구; 부정·보류·기록 없음은 종결 차단; acceptance §D.5 항목 6, plan §E7); F6 — M-4·M-8 탐지 주장 정정 (검출 면은 E7(d)·E7(a) 검토이며 판정 보류 가드는 게이트가 아님; M-4의 shall-not 조항 번호 REQ-RDX-012 → REQ-RDX-013); F9 — 본 행 신설 (3차 편집 기록: M-2/M-6/M-7, §5 item 5 — round-3 커밋 `7fd68bbe2`); F10 — 미해결 판단 표지 리터럴 제거 (plan §B·§D3, acceptance §D.5 항목 5; MP-7 리터럴 검사 0줄). **판정 보류 강등 6면**: F1·F2 (AC-RDX-009 육면 면 LED-022..027 — 제거면 LED-021은 블로킹 유지), F3 (AC-RDX-014 블록 면 BLK-014G), F4 (AC-RDX-015 블록 면 BLK-015G), F7 (AC-RDX-008·010 BP 면 BP-008G·BP-010G), F8 (AC-RDX-001 보존 면 LED-001C). 블록 면을 잃은 AC-RDX-008·010·014·015는 블로킹에서 판정 보류로 옮긴다: 블로킹 8 → 4 (001·002·007·009). 구조 검사 설계(블록 경계 파서, 주석·코드펜스 제거, manifest.json 파싱, 감시 규범 문장 전체 일치)는 후속 카드 소관 (`.moai/reports/t1579/followup-card-draft.md`). frontmatter version 0.9.0 동기화 |
| 0.10.0 | 2026-10-10 | manager-spec | 착지 전 MP-8 재고정 + 육면 차단 개수 검사 — 운영자 결정 `d-20261010T053033Z-b80b`(ceiling-exception, 선택지 (a) 「차단 개수 검사 추가」)에 따라 수리 커밋 1건(MP-8 재고정 + LED-028 차단 계수 + 구형 호스트 잔존 제거)을 적용하고 계획 감사 1회를 잇는다. 선행 결정 `d-20261010T042422Z-06a3`(문장 수정 + 감사 1회)은 그 문장 수정이 트리에서 HEAD로 되돌려졌으므로 본 행에서 새 문장으로 재작성했다(acceptance AC-RDX-009 시나리오·§D.2 행, plan §D1 AC-RDX-009 행). **MP-8 재고정**: AC-RDX-001·002·007의 RED-now 셀을 핀 `2aab5f797` `git grep -c -h` 귀속으로 바꿨다(출력 없음 · exit 1). HEAD `2ed9d2681` GREEN은 §D.3-d의 G 블록에 기록했다. **LED-028 신설**: 6종 캐노니컬 URL 각 1 이상(파일 전체) 블로킹 계수를 추가했다(핀 RED 출력 없음 · exit 1, 착지 GREEN은 lane 측정). AC-RDX-009 블로킹 면은 LED-021 + LED-028이며, 육면 블록 스코프 LED-022..027은 판정 보류로 유지된다(F1·F2). **M3 항목 5 게이트 문구**: 착지 게이트에 LED-028(6종 각 1 이상, 착지 후 검증)을 추가했다. 작업 의미는 그대로다. **구형 호스트 잔존 제거**: 착지 커밋이 `hns-release-update-specialist.md:179`의 `docs.anthropic.com` 잔존을 제거한다(manager-develop 수행, 본 SPEC 편집 범위 밖). 집계(블로킹 4 · 판정 보류 10 · 회귀 가드 2)와 REQ 번호, DoD 블로킹 AC 집합은 불변이다(DoD 항목 1 문구는 HEAD와 동일하게 유지 — AC-RDX-009 게이트 문구 갱신 여부는 리더 판정 494e로 확정 — spec §1.2 D8 참조). **LED-028 좌측 경계 강화(중첩 URL 변이 차단)**: 명령을 좌측 경계(행두·공백·따옴표·괄호·꺾쇠·대괄호 뒤) 포함 형태로 교체했다 — 우측 경계만 있던 형태는 `https://attacker.invalid/https://code.claude.com/docs/en/…` 중첩 변이를 6종으로 계수했다. **잔여 위험 선언 (LED-028 전체 파일 계수의 한계)**: 블록 밖에서 다시 쓴 URL은 LED-028을 통과하며, 이를 잡는 것은 판정 보류 블록 면 LED-022..027이다(블록 스코프 블로킹 게이트는 구조 검사 후속 카드 소관). frontmatter version 0.10.0 동기화 |
| 0.11.0 | 2026-10-10 | manager-spec | plan-audit iter6 수리 (M-1·M-2; 감사 대상 커밋 `cb54103ea` 고정 재측정 — `.moai/reports/t1579/plan-audit-iter6.md` M-1·M-2). **M-1 (현행화)**: plan §A 귀속 문장·§C 헤더·§C L67을 고정 커밋 `cb54103ea` 기준으로 재서술했다 — `git grep -c -h "docs.anthropic.com" cb54103ea -- …; echo "exit=$?"` 출력 `exit=1`(0힛). 2ed9d2681 실측 1 / exit 0은 SUPERSEDED로 이력 보존. acceptance LED-021T 라벨 SHA를 `cb54103ea`(M3 항목 5 착지, `git show` 179행 실측)로 채우고 상태를 MET(고정 커밋)로 갱신했으며, 2ed9d2681 NOT MET과 d36e97571 상태는 SUPERSEDED 표지로 남겼다. 녹색 셀 LED-001 G · LED-002 G · LED-007 G의 귀속을 착지 커밋으로 바꿨다 — M4 `18ea52c0a`(domain·best-practices), M1 `4fe4ffe7b`(rust-v0.161.0 seed), `git cat-file -t` = commit, 각 착지에서 셀 값을 재관측 — 고정 커밋 측정 행을 함께 기록했다. AC-RDX-009 Then의 2ed9d2681 상태는 SUPERSEDED 처리. **M-2 (리더 판정 494e)**: spec §1.2에 D8 행을 신설했다 (DoD 항목 1 HEAD 문구 유지; LED-028 블로킹 면은 AC-RDX-009·§D·M3 항목 5 한정; 잔여 위험은 선언만). 0.10.0 행의 대기 문구는 D8 참조 확정 문구로 바꿨다(대체 전 문구는 D8 행에 기록). acceptance DoD 항목 1 문구는 불변이며, 바로 아래 기록 주석에 LED-028 항목 밖과 항목 안 상태 문장의 SUPERSEDED를 명시했다. §D.5 항목 5의 「열린 판단 전부 §1.2」 주장은 재검증 결과 유지(미수정). **코드 수리 기록 (후속 에이전트가 specialist.md·run.js에 적용, 디스크 확인 완료 — 본 행은 기록만 한다)**: C-1 — specialist Phase 1 codex 수집 경로를 단일 인용(`gh api 'repos/openai/codex/releases?per_page=30&page=1'`)으로 고쳤다; C-2 — Phase 1 명령과 codex 렌즈 프롬프트에 페이징과 기준선 중단 조건(baseline tag 출현 또는 목록 끝)을 추가했다 (REQ-RDX-006); C-3 — 폴백 절차 (1)–(3)을 러너 렌즈 프롬프트에 보간하고(`CODEX_COMMITS_FALLBACK_STEPS`) specialist Phase 1 문장에도 명기했다, 라벨 `commits-api-reconstruction`은 유지 (REQ-RDX-007). **검증 동사 E3-P5**(codex-prompt-pr-body-step)**·E3-P6**(codex-prompt-paging-stop)을 plan §E에 추가했다 — 원문 복사, GREEN은 각각 `…-ok` · exit 0, RED는 미수리 스냅샷 cb54103ea에서 exit 1로 재관측. **줄 참조 재고정**: specialist 118행 뒤 +2행 이동(Phase 2–4 헤딩 130/167/195, BP 헤딩 299·섹션 299–319·deliverable 316), acceptance §D.3-c의 run.js 렌즈 프롬프트 join 참조 113→115행 — 현재 상태 참조만 고쳤고 과거 관측과 G블록은 그대로다. **LED-004 녹색 셀**: `≥1` (AC-RDX-004 문턱; cb54103ea 2, 이 라운드 4). 블로킹 4 · 판정 보류 10 · 회귀 가드 2, REQ·AC 번호, 게이트·명령, DoD 항목 1 텍스트는 불변이다. frontmatter version 0.11.0 동기화 |

## 1. 문제 — 측정된 형태

release-update 하네스는 Claude Code 단일 축으로 태어났다. 운영자 확장 지시(2026-10-07 — "claude code 뿐만 아니라 codex 최신 체인지 로그를 분석해서… 동시에 범용적으로 사용할 수 있도록… 항상 관련 베스트프랙티스나 논문/전문/공식 클로드 베스트프랙티스 자료를 찾아서 html로 보고하고 카드 발행")는 1차·2차 확장 스윕을 **배차 메시지의 축어 지시문**으로 구동했을 뿐, 그 확장은 하네스 파일 어디에도 영구화돼 있지 않다. 1차 연구의 Phase 7.5 finding (confidence 0.9)이 정확히 이 결함을 명명한다: *"re-dispatch without the verbatim directive loses the expansion"*. 2차 스윕조차 재배차 메시지로 구동됐고, 그 Gaps 절은 "t1579(CARD-1) 착지 전까지 확장 스코프는 배차 메시지 의존"을 명시 기록했다.

**측정 트리**: `.moai/worktrees/t1579` (branch `WT-high-10-07`) @ `2aab5f797`. 이 SPEC의 모든 수치와 grep 관측은 이 트리의 실측이다.

```console
$ grep -c "Codex" .claude/commands/harness/release-update/manifest.json
0        # exit=1 — 매니페스트에 codex 축 표면 없음
$ grep -ci "codex" .claude/workflows/hns-release-update-run.js
0        # exit=1 — Runner fan-out에 codex 렌즈 없음
$ grep -c "last-codex-version" .claude/agents/harness/hns-release-update-specialist.md
0        # exit=1 — 본문 전체에 codex 상태 스키마 없음
$ grep -rn "last-cc-version" internal/
(출력 없음)   # exit=1 — 상태 파일의 Go 라이터 부재 — 하네스 계층 전용임을 확인
```

### 1.1 실측 요약 (트리 `2aab5f797`)

| # | 측정 대상 | 관측값 | 근거 |
|---|---|---|---|
| M1 | 매니페스트 domain | `"moai-adk-go dev-only maintainer tooling — Claude Code upstream change tracking"` — CC 단일 | manifest.json:3 |
| M2 | Runner codex 표면 | `grep -ci codex` = **0** (exit 1). fan-out은 CC `versionDeltas` 전용 (`selectResearchSweepTargets` 4힛 — CC 렌즈 셀렉터만 존재) | hns-release-update-run.js |
| M3 | 상태 파일 codex 키 | 스키마 키 없음. 운영 증거: 2차 스윕의 codex 결과가 CC 파일 `genuine_delta` **자유 서술**에 편입됨 — *"Codex 축(확장 스코프 2차): rust-v0.161.0 안정 승격 2026-10-07T15:58:45Z — 큐레이팅 14불릿 T1=4/T2=4/T3=6"* — 축은 존재하는데 스키마 홈이 없다 | `last-cc-version.json` (primary 체크아웃) 2026-10-08 항목 |
| M4 | Go 라이터 | `grep -rn "last-cc-version" internal/` = 0힛 — 상태 파일은 하네스 계층(specialist/runner)이 쓴다. codex 키 신설도 Go 변경을 수반하지 않는다 | 본 트리 실측 |
| M5 | BP 축 | 스페셜리스트 본문 `grep -ci best-practice` = **0**. 1·2차 스윕의 BP 축 산출(원문 패치 4건: multiagent-harmony·context-engineering·managed-agents·Opus 5.5 프롬프팅 가이드)은 절차 홈 없음 | specialist.md 본문 + 연구 2건 |
| M6 | codex 기준선 (2차 스윕 실측) | npm `@openai/codex` = 0.161.0 = 안정 (`rust-v0.161.0` 승격 2026-10-07T15:58:45Z), 설치 바이너리 0.160.1, 리포 테스트 핀 0.160.0. 0.160.1→0.161.0 델타는 **이미 큐레이팅 완료** (14불릿 T1=4/T2=4/T3=6) | upstream-update-20261008.md 축2 |
| M7 | alpha 관찰목록 | 0.161/0.162 alpha 릴리즈 본문은 1줄 제목(내용 없음) — 커밋 주제 복원으로 실측 (0.160.1 이후 98건 + 100건). 어댑터 노출 6테마(thread/rollout/서브에이전트/compaction/MCP/기타) 관측 완료, 안정 미탑재 watch 항목 | upstream-update-20261007.md 축2 표 |
| M8 | Phase 3 URL 세트 드리프트 | 스페셜리스트 Phase 3 URL 6종이 `docs.anthropic.com/en/docs/claude-code/*` — 실제 페치는 전부 `code.claude.com/docs/en/*`로 캐노니컬라이즈 (2회 연속 관측). 본문 `code.claude.com` = 0힛 | upstream-update-20261008.md finding (confidence 0.75) |
| M9 | 템픔릿 비대상 | 3개 표면 전부 사용자 소유 네임스페이스 (`.claude/agents/harness/`, `.claude/workflows/hns-*`, `.claude/commands/harness/`) — `moai update` 비대상, `internal/template/templates/` 미러 불요, `make build` 불요 | agent-authoring.md § Agent Directory Convention + dynamic-workflows.md hns-* 조항 |

### 1.2 설계 결정 기록 (재논의 금지 — plan.md §D와 쌍을 이룬다)

| # | 결정 | 근거 |
|---|---|---|
| D1 | codex 축 상태 파일은 **별도** `.moai/state/last-codex-version.json`으로 신설하고 CC 파일과 동일 키 계열(`last_analyzed_version` / `last_analyzed_date` / `last_master_research` / `analysis_history[]`)을 미러링한다. 시드 `last_analyzed_version` = **`rust-v0.161.0`** — 원본 릴리즈 태그 형태 그대로 기록한다(CC 파일의 plain semver "2.1.294"와 표기 형식이 달라도 태그 형태가 스윕 비교 기준이므로). 카드 후보안의 0.160.1 시드는 **폐기** | 카드 지시 "별도" 그대로. 2차 스윕이 0.160.1→0.161.0 안정 델타를 이미 분석·큐레이팅했으므로(§1.1 M6) 기준선은 0.161.0이다 — 0.160.1 시드는 1차 스윕 시점(2026-10-07) 기준이며 2차 실측(2026-10-08)이 우선한다. 실제 파일 생성은 다음 스윕 실행 시점의 하네스 절차가 수행한다(기계 로컬 — 본 SPEC은 본문의 스키마 문서화만 소유). **시드 의미론은 last-analyzed다** — 분석되지 않은 신규 안정 승격(예: 0.162 선행 승격)이 관측돼도 시드는 고정되고 그 델타는 다음 스윕의 분석 대상으로 기록된다(plan §C 재판정 규칙 — plan-audit iter1 CX-1) |
| D2 | Runner는 CC 렌즈와 병렬로 **codex 렌즈**(fan-out)를 얻는다. codex 렌즈는 릴리즈 본문이 비어 있을 때(밀도 높은 alpha 기간) 커밋 API 복원 폴백을 **요구 절차**로 문서화하고, 복원 항목은 전부 커밋-주제-유래로 라벨링한다. 관찰 목록 형식: 테마 행 = 테마 키 + 관측 PR 번호 목록 + MoAI 노출면 | 1차 연구 Phase 7.5 finding (b, confidence 0.8): *"Phase 1 needs a documented commits-API reconstruction fallback or the codex axis yields no content"*. 2차 #49713이 보인 정합 절차 — 커밋 제목만으로 판정하지 않고 PR 본문 확인으로 격상 — 를 렌즈 절차에 흡수 |
| D3 | 6테마 어댑터-노출 관찰목록은 **Runner의 codex 렌즈 프롬프트**에 상주한다(영어 키: `thread` / `rollout` / `subagent` / `compaction` / `MCP` / `other`). 스페셜리스트는 러너 산출을 받아 큐레이션·티어 분류·안정 승격 판정을 수행한다 — alpha 테마는 watch 관찰목록으로만 기록되고, 안정 릴리즈 탑재 시에만 채택 판정한다 | 기존 CC 축의 러너(비대화형 스윕)/스페셜리스트(인간 게이트) 분업 계승(§1.1 M2, specialist "Runner integration" 절). 영어 키는 coding-standards.md 에이전트 정의 영어 규정. 1차 스윕의 watch 판정("채택 아님, 준비 카드만 제안")이 절차 규범이 된다 |
| D4 | 매니페스트 `sprint_contract.dimensions`는 **Functionality/Consistency 2개 유지, thresholds(0.85/0.80) 불변**. `domain` 문자열만 이중 축 + BP 축을 명명한다 | 차원 추가는 이 하네스 향후 모든 run의 sync 채점 의미론을 바꾸는 정책 변경이다 — 운영자 지시(도메인 확장)는 그것을 요구하지 않는다. Enforce Simplicity 사다리 1단(YAGNI) |
| D5 | BP 축은 스페셜리스트 본문에 **상시 절차 섹션**으로 영구화한다: 스윕마다 공식(Anthropic/OpenAI) 게시 면을 스캔하고, BP 항목이 제안(카드 발행·문서 싱크 권고)의 근거가 되려면 **원문 패치(verbatim fetch) 선행**을 요구한다 — 검색 요약·2차 자료는 보고 전용 리드. HTML 제안 보고를 명명 산출물로 기록한다. 실행은 본 SPEC 소관이 아니다(절차 영구화만) | 운영자 "항상" 지시. 1차 Residual-risk "검색 경유 BP 리드는 제목/날짜 오정보 가능"이 2차 BP-1에서 실현됐다(게시일 2026-10 → 원문 패치로 2026-04-08 정정) — 원문 선행 강제가 그 재발 방지다 |
| D6 | 스페셜리스트 Phase 3 문서 URL 세트를 `code.claude.com/docs/en/*` 캐노니컬 형태로 갱신한다 | 2차 스윕 finding (confidence 0.75, *"t1579 owns the body — recorded as input to that card"*): 2회 연속 페치 전부 캐노니컬라이즈 관측. 본문을 고치는 SPEC이 본 SPEC이므로 함께 정착한다 |
| D7 | 본 SPEC은 절차의 **영구화**만 소유한다 — 스윕 실행 자체는 소관 밖(§6) | 카드 본문 "하네스 절차화". 1·2차 스윕은 이미 배차 메시지로 실행 완료 |
| D8 | 리더 판정 494e (커밋 `cb54103ea` 메시지 원문: 「Leader ruling 494e (scope: DoD item 1 unchanged; LED-028 only in AC-RDX-009, §D and M3 item 5)」). **DoD 항목 1은 HEAD 문구를 유지한다 — 수정하지 않는다.** LED-028(6종 캐노니컬 URL 블로킹 계수)의 블로킹 면은 AC-RDX-009·§D·M3 항목 5에만 있고 DoD 항목 1 밖이다. **잔여 위험은 선언만 하고 바꾸지 않는다**: 블록 밖에서 다시 쓴 URL은 LED-028을 통과하며, 이를 잡는 것은 판정 보류 면 LED-022..027이다 (AC-RDX-009 잔여 위험 선언, acceptance §D.3-d Gaps) | 운영자 소유 DoD 항목은 본 SPEC이 바꾸지 않고 기록만 맞춘다. 0.10.0 행의 「리더 답변 대기」를 이 결정으로 대체한다 (0.11.0, plan-audit iter6 M-2) |

## 2. 원인 — 확장이 파일이 아니라 배차문에 살았다

하네스는 SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001에서 CC 전용 능력으로 포팅됐고(매니페스트·러너·스페셜리스트·상태 파일 4면 모두 CC 전용으로 설계), 이후 운영자가 스코프를 확장할 때마다 새 카드가 아니라 **배차 메시지에 확장 지시문을 얹는** 방식으로 운영됐다. 배차문은 세션 생존 자료다 — 세션이 끝나면 소실되고, 재배차 시 지시문을 다시 붙이지 않으면 확장은 잃어버린다. 1차 스윕(2026-10-07)이 발견한 것은 기능 결함이 아니라 **영속성 결함**이다: 축은 운영 중인데(C3 판정 "확정"), 그 축을 담는 스키마·렌즈·절차가 하네스 파일에 없다(C4 판정 "구조적 괴리"). 2차 스윕이 같은 결함을 재확인했고, 상태 파일에서는 codex 결과가 CC 모양의 자유 서술 필드에 짓눌려 저장되고 있다(§1.1 M3) — 스키마 홈 부재가 이미 관측 비용을 내고 있다.

## 3. 이 카드의 대표 mutant

아래 mutant들이 이 SPEC의 AC를 통과하려면 AC가 너무 얕은 것이다(verification-completeness §2 mutant probe).

- **M-1 "출처 필드 기만"**: `source_request`(역사 서술 필드)에만 codex를 언급하고 `domain` 문자열은 CC-only로 남기는 mutant. AC-RDX-001이 `domain` 키 행 스코프 패턴(`'"domain".*Codex CLI upstream change tracking'`)을 grep하므로 잡힌다 — source_request에 동일 문구를 넣어도 매치되지 않는다(plan-audit iter1 CX-3 재앵커).
- **M-2 "주석 코덱스"**: JS 주석에만 `// TODO codex lens`를 추가하는 mutant. 검출 면은 run-phase 검증 동사 plan §E3-P2(export 경로의 실측 target 생성 — 미연결 정의·주석만 있으면 export 부재 또는 codex-count 단언에서 좌초)와 착지 신호 LED-003(`selectCodexSweepTargets(args)` 출현 ≥2)이다. AC-RDX-003은 판정 보류(CX-7)라 이 mutant는 릴리스 게이트가 보호하지 않는다 — run-phase 동사 관측으로만 잡힌다(plan-audit iter1 CX-2 재앵커; acceptance §A·§D.2).
- **M-3 "seed 누락"**: `last-codex-version.json` 스키마는 문서화하되 시드값을 빼는 mutant. AC-RDX-006은 통과하고 AC-RDX-007(`rust-v0.161.0`)에서 잡힌다 — 둘이 쌍인 이유다.
- **M-4 "BP 껍데기 섹션"**: `Best-Practices` 헤딩만 넣고 원문-패치 강제를 빼는 mutant. AC-RDX-008(섹션 존재)은 통과할 수 있다 — E7(d) 의미 검토뿐이다 (`source-first` 리터럴 AC-RDX-016은 판정 보류라 게이트가 아니다; shall-not 조항은 REQ-RDX-013이다 — F6). §3에 한계를 명시한다: 이 mutant는 grep 단일 판정면 밖이며 plan-auditor 서술 검증이 보완 판정면이다. AC-RDX-014/015의 블록 면은 분할 결정(efde)으로 판정 보류가 되었고, 블록 안 문장의 판독은 E7 검토와 후속 카드의 구조 검사가 맡는다(plan-audit iter1 D5). **CX-13**: `source-first` 리터럴 면도 반전 가능(리터럴 유지·규칙 역전) — AC-RDX-016의 기계 면을 회귀 가드(판정 보류)로 강등하고 REQ-RDX-013 의미론 면은 plan §E7 검토면 + 형제 카드 판별기로 이관한다(리더 재개 재심).
- **M-5 "Go 침입"**: 상태 파일 쓰기를 Go 런타임(`internal/`)으로 옮기는 mutant. AC-RDX-011(회귀 가드 — `internal/` grep 0힛 유지)에서 잡힌다. 상태 파일은 하네스 계층 소유가 측정으로 확인된 구조적 사실이다(§1.1 M4).
- **M-6 "병합 제외"**: `selectCodexSweepTargets`를 정의·export·직접 호출하되 그 target을 `parallel(...)` 병합에서 제외하는 mutant. LED-003(≥2)은 통과할 수 있다. 검출 면은 run-phase 검증 동사 plan §E3-P3(모의-런타임 관측 — codex 라벨 agent 호출 0건이면 LED-016에서 좌초)과 §E3-P4(`run()` 공개 경로)다. AC-RDX-003은 판정 보류(CX-5·CX-7)라 게이트 보호는 없고, run-phase 동사 관측으로만 잡힌다(plan-audit iter2 CX-5).
- **M-7 "단일 사이트 기록"**: codex 상태 문서를 Phase 0에만 두고 Phase 7a 기록 단계를 CC-only로 남기는 mutant. LED-006 단독(≥1)은 통과한다. 검출 면은 착지 신호 LED-006(≥2 — Phase 0 판독 사이트 + Phase 7a 기록 사이트)과 LED-017(`7a-codex` 기록 단계 리터럴)이며, AC-RDX-006은 판정 보류(CX-8)라 게이트 보호는 없다 — 비게이트 착지 신호 관측으로만 잡힌다(plan-audit iter2 CX-6; acceptance §D.2).
- **M-8 "CC 널 조기 종료"**: codex/BP 절차를 추가하면서 Phase 2의 무조건 조기 종료("If no entries … stop.")를 그대로 남기는 mutant — CC 빈 주간(2026-10-07→08 패턴)에 codex/BP가 실행 전 종료된다. 현재 본문 상태가 이 클래스다. 검출 면은 E7(a) 축별 종료 검토다 — LED-018(`only the CC axis` 0힛)은 판정 보류 가드라 게이트가 아니다(plan-audit iter3 CX-9; F6). 리터럴만 주석으로 넣고 문장을 생존시키는 형태는 LED-019(구형 문장 전문 제거면)가 잡는다(fresh-run CX-10). **한계 (CX-12)**: 동의어 바꿔쓰기(paraphrase) 클래스는 리터럴 쌍을 우회한다 — 다른 무조건 종료 문구가 쌍 통과임이 codex 실증됐고, 유한 리터럴 집합은 의미론을 보증하지 못한다. AC-RDX-017은 회귀 가드(판정 보류)로 강등되고 의미론 판정은 plan §E7 검토면(run-exit E1 인간 검토) + 형제 카드 판별기로 이관된다.

## 4. 요구사항 (GEARS)

> 게이팅 상태(블로킹 AC 연결과 강등 처분)의 SSOT는 acceptance.md §D.2 '게이팅 처분 (B-07)' 표다. 분할 결정(efde) 이후 REQ-RDX-001·004·009·012는 블로킹 AC를 잃어 강등되었다 (acceptance.md §D.2 게이팅 처분 표).

### codex 축 — 상태 스키마 (M1)

- **REQ-RDX-001** (Ubiquitous) — The release-update harness shall maintain a dedicated codex-axis state file at `.moai/state/last-codex-version.json`, separate from `.moai/state/last-cc-version.json`, mirroring the CC file's key family (`last_analyzed_version` / `last_analyzed_date` / `last_master_research` / `analysis_history[]`) (결정 D1). **게이팅**: 없음 — AC-RDX-014 판정 보류 (분할 결정 efde, F3); 키군 판독은 E7 검토와 후속 카드 (acceptance.md §D.2).
- **REQ-RDX-002** (Ubiquitous) — The specialist body shall document the `.moai/state/last-codex-version.json` schema in Phase 0 (read) and Phase 7a (write) with `last_analyzed_version` seeded at `rust-v0.161.0` — the stable promotion already analyzed by the 2026-10-08 sweep (결정 D1).
- **REQ-RDX-003** (Event-driven) — **When** Phase 7a persists state after a sweep, the specialist shall write BOTH `last-cc-version.json` (CC axis) and `last-codex-version.json` (codex axis) — a CC-only write leaving the codex baseline stale is prohibited.
- **REQ-RDX-004** (Event-driven) — **When** the codex state file is missing at Phase 0, the specialist shall default the codex since-baseline to `rust-v0.161.0` and emit a warning, mirroring the CC file's missing-file behavior.
- **REQ-RDX-005** (Unwanted) — The codex-axis state write shall not introduce a Go-side writer: `internal/` shall keep zero references to `last-codex-version` (the state file is harness-layer-owned, §1.1 M4 — mutant M-5 봉쇄).

### codex 축 — Runner 렌즈 (M2)

- **REQ-RDX-006** (Ubiquitous) — The Runner shall fan out a codex lens alongside the CC lens, implemented as a parallel selector function `selectCodexSweepTargets` emitting one read-only analysis target per codex release-window delta (stable channel, GitHub releases API snapshot) (결정 D2, plan §D 앵커).
- **REQ-RDX-007** (Event-driven) — **When** a codex release body is empty (alpha-dense windows carry 1-line titles), the codex lens shall reconstruct content from the commits API per the documented fallback procedure (`CODEX_COMMITS_FALLBACK`), and shall label every reconstructed item as commit-topic-derived — never as release-note text.
- **REQ-RDX-008** (Ubiquitous) — The codex lens shall classify observations against the standing 6-theme adapter-exposure checklist (`CODEX_THEME_CHECKLIST`: `thread` / `rollout` / `subagent` / `compaction` / `MCP` / `other`), each theme row carrying observed PR numbers and the MoAI exposure surface (결정 D3).
- **REQ-RDX-009** (Unwanted) — An alpha-window theme entry shall not be reported as adopted drift: alpha themes remain watch-list observations, and adoption judgment happens only when the theme lands in a stable release (1차 스윕 watch 판정의 절차화).

### 매니페스트 (M4)

- **REQ-RDX-010** (Ubiquitous) — The manifest `domain` string shall name both axes (Claude Code + Codex CLI upstream change tracking) and the best-practices axis (plan §D 고정 리터럴).
- **REQ-RDX-011** (Ubiquitous) — The manifest `sprint_contract.dimensions` shall remain `Functionality` + `Consistency` with unchanged thresholds — the dual-axis expansion is scored within the existing dimensions (결정 D4).

### BP 축 + 본문 정착 (M3)

- **REQ-RDX-012** (Ubiquitous) — The specialist body shall carry a standing best-practices procedure section: per-sweep scan of official Anthropic/OpenAI publishing surfaces, with the BP item inventory recorded per sweep (결정 D5).
- **REQ-RDX-013** (Unwanted) — A BP item shall not back a proposal (card issuance, docs-sync recommendation) unless its source article has been fetched verbatim (`source-first` 원문 패치 선행) — search-result summaries and secondary sources are report-only leads (결정 D5, mutant M-4 봉쇄).
- **REQ-RDX-014** (Ubiquitous) — The BP axis shall name the HTML proposal report as a deliverable, and the specialist Phase 3 doc-fetch URL set shall list the `code.claude.com/docs/en/*` canonical URLs (결정 D5 + D6).
- **REQ-RDX-015** (Event-driven) — **When** the CC axis observes a null delta (no new CC versions since its baseline), the specialist shall terminate only the CC axis and shall still execute and record the codex axis (release-window delta + commits fallback) and the best-practices axis (official-source scan + inventory) in the same run — a run shall not emit its completion summary while any axis remains unexecuted (plan-audit iter3 CX-9; M1은 Phase 2 조기 종료 문장을 `only the CC axis`로 재범위화하고 Phase 8 완료 게이트가 3축 실행 상태를 집계한다).

## 5. 알려진 구속 조건

1. **3개 표면 전부 dev-only 사용자 소유 네임스페이스다.** `.claude/agents/harness/`·`.claude/workflows/hns-*`·`.claude/commands/harness/`는 `moai update` 비대상이고 `internal/template/templates/` 미러가 없다(§1.1 M9) — 템픔릿 미러 작업과 `make build`는 이 카드에 존재하지 않는다. run-phase가 실수로 템플릿을 건드리면 그것은 scope 위반이다.
2. **상태 파일은 기계 로컬 gitignored다.** `last-codex-version.json`의 실제 생성은 CI/테스트가 판정할 수 없으므로, 본 SPEC의 AC는 스페셜리스트 본문의 쓰기 지점(스키마 문서화)을 측정면으로 삼는다. 2차 스윕이 이미 보여줬듯 스키마 홈 부재는 관측 비용을 내고 있다 — 본문 문서화가 스키마의 규범면이다.
3. **Runner 불변식은 유지된다.** AskUserQuestion·gh pr 호출 금지(HARD, AC-DHC-007a), `Date.now()`/`Math.random()` 금지(결정성), top-level 실행 + CommonJS export 가드 패턴 유지. codex 렌즈는 이 불변식 위에 병렬 구조로 얹힌다.
4. **에이전트·러너 본문은 영어다**(coding-standards.md Language Policy). 6테마 키와 절차 서술은 영어로 기록하고, 한국어 테마명(서브에이전트/기타)은 대응표로만 남긴다.
5. **Tier M 예산** — REQ 15건 / AC 16건. Tier M 천장은 요구사항·수용기준 각각 16이며, AC 16은 천장에 정확히 도달한다(AC-RDX-012는 AC-RDX-011로 합병됨 — D7). 분할 결정(efde) 이후 블로킹 4 · 판정 보류 10 · 회귀 가드 2이다.

## 6. 범위 밖 (Non-goals)

### Out of Scope — Go 측 codex fixture 재핀

- `internal/cli/managed_codex_tui_test.go`의 `testdata/codex-0.160.0` 핀 갱신, `internal/config/defaults.go:135` 계측 주석, auth.json/키링 실측, `resume --help` 대조 — 어댑터 conformance 재측정은 별도 카드(CARD-4, 2026-10-08 리더 발행)의 소관이다. 본 SPEC은 본문 절차만 소유하고 Go 코드를 1줄도 만지지 않는다.

### Out of Scope — 스윕 실행 자체

- 1차(2026-10-07)·2차(2026-10-08) 스윕은 이미 배차 메시지로 실행 완료됐다. 본 SPEC 착지 후의 실제 스윕 실행(codex 상태 파일 최초 생성 포함)은 다음 `/harness:release-update` 배차의 일이다(결정 D7).

### Out of Scope — CARD-2·CARD-3 (연구 제안 표의 형제 카드)

- CARD-2: context reduction ladder에 tool result clearing 량 추가(`context-window-management.md`+detail+미러+docs-site 4-locale). CARD-3: 멀티에이전트 조화 실패 패턴의 MoAI 독트린 매핑 참조 문서(+BP-1·BP-2 매핑 합류). 둘 다 1차 연구 제안 표의 독립 카드다.

### Out of Scope — docs-site 4-locale 동기화

- release-update 하네스는 dev-only 메인테이너 인프라로 사용자 프로젝트에 배포되지 않는다. 스페셜리스트가 수행하는 문서 싱크 절차(Phase 6)는 변경하지 않는다.

### Out of Scope — 템픔릿 미러링

- `internal/template/templates/` 측 미러 생성·동기화. 3개 표면이 전부 사용자 소유 네임스페이스라 미러가 존재하지 않는다(§1.1 M9, §5.1).

### Out of Scope — BP 매핑 참조 문서 작성

- BP 항목의 MoAI 독트린 매핑(부패성 휴리스틱 클래스 명명, elapsed-time budget 실험 등)은 CARD-3의 확장 범위다. 본 SPEC은 "원문 패치 → 제안 근거"의 절차 강제만 만든다.

## 7. 미검증 항목 (Gaps)

- **시드값의 유효기간** — `rust-v0.161.0`은 2026-10-08 기준 안정 최신이며 시드 의미론은 **last-analyzed**다. run-phase 착지 전에 0.162 승격이 관측돼도 시드는 올리지 않는다 — 미분석 델타는 다음 스윕의 분석 대상으로 기록된다(plan §C 재판정 규칙 — plan-audit iter1 CX-1).
- **6테마 관찰목록의 후속 변동** — 0.162 승격 시 테마 추가/삭제가 예상된다. 본 SPEC은 6테마를 시드로 고정하고(plan §D), 확장/축소는 스윕 재량으로 기록된다.
- **BP 공식 자료 URL 목록의 완결성** — BP-3(context engineering 문서)의 정확 경로는 2회 스윕에서도 미확정(검색 색인만 확인). BP 절차는 "원문 패치 시점에 경로 확정"을 요구하고 본 SPEC은 URL 인벤토리를 완결하지 않는다.
- **러너 codex 렌즈의 실제 fan-out 동작** — 본 SPEC 착지는 절차 편집이지 실행이 아니다. 첫 codex 렌즈 실행은 다음 스윕에서 관측되며, 그때까지 렌즈 프롬프트의 커밋 복원 실효성은 미검증 상태로 남는다(§6 스윕 실행 제외와 동일 뿌리).
- **CX-7/CX-8 계측 경화 부채 (형제 카드 제안 — plan-audit iter3 ceiling STOP 분할 결정)** — codex 렌즈 AC(AC-RDX-003/004/005)의 계측은 구조 면(식별자·상수 존재, 형태 키, 라벨 접두사)만 전달하고 내용 면(폴백 절차 실문·6테마 행 실문)을 측정하지 못한다 — 빈 상수·"Return ok." 프롬프트 mutant가 모든 기계 면을 통과한다(CX-7). AC-RDX-006의 이중 사이트 기준도 전-file 계수라 산개 언급 mutant(Phase 0에 파일명 2회·주석 속 `7a-codex`·Step 7a CC-only)를 걸러내지 못한다(CX-8). 두 AC군은 verification-completeness §2 채택 기준(계측이 너무 얕아 채택 불가 — mutant-probe adoption bar; §2.1 처분군 적용)에 따라 회귀 가드(판정 보류)로 강등됐고, 해결 계측(내용 계측기 — 상수 블록 내 절차 마커·테마 리터럴 계수 / 사이트 판별기 — 구획 스코프 추출의 가드 통과형 / **축별 종료 판별기** — 동의어 바꿔쓰기에 강건한 의미론 계측, CX-12 / **source-first 의미론 판별기** — 리터럴 반전 클래스, CX-13)은 **형제 카드로 리더/큐에 제안된다**. 본 SPEC의 M1/M2는 plan §D1이 핀한 내용을 서술 규율로 작성하며 E1 인간 검토가 제2 판정면이다.
- **구조 검사 설계 부채 (분할 결정 efde)** — 블록 경계 파서(AC-RDX-014·015·008·010 면), 주석·코드펜스 제거(AC-RDX-008·010), manifest.json 파싱(AC-RDX-001 LED-001C 면), 육면 블록 경계(AC-RDX-009 LED-022..027)는 후속 카드 소관이다 (`.moai/reports/t1579/followup-card-draft.md`). 그 전까지 이 면들은 판정 보류 관측이다.
