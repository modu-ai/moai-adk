---
id: SPEC-RELUP-DUALAXIS-001
title: "acceptance — release-update 하네스 CC+Codex 이중 축 정착"
created: 2026-10-09
author: manager-spec
tier: M
---

# acceptance: SPEC-RELUP-DUALAXIS-001

## §A 범위·판정 규율

- 모든 릴리스 블로킹 AC는 verification-completeness.md §2 two-cell 규율을 따른다: **RED-now 셀**(구현 전 트리에서 적색 관측 — 단일 읽기전용 명령 + 그 명령의 축자 stdout + exit code 독립 필드 + 트리 SHA 핀)과 **green-path 셀**(어느 마일스톤이 뒤집는지 + 녹색 출력 형태)이 쌍으로 존재한다.
- **트리 핀**: 전체 RED-now 관측은 `2aab5f797` (worktree `.moai/worktrees/t1579`, branch `WT-high-10-07`)에서 수행했다. 문서 수준 핀이 이 문서이며, 개별 AC 핀이 없는 한 이 핀이 구속한다.
- **귀속 분리 (B-01·B-04)**: RED 셀의 귀속 SHA는 `2aab5f797b75983e132af451da68f69e3426557b`다 (형식: `git grep -c -h "<패턴>" <sha> -- <경로>`). 착지 GREEN 셀은 그 셀을 만든 착지 커밋(M1 `4fe4ffe7b` · M4 `18ea52c0a`)과 고정 커밋 `cb54103ea`(감사 대상 HEAD)에 귀속한다 — `2ed9d2681` 귀속은 SUPERSEDED(이력 보존; 명령·출력은 §D.3 LED 행과 §D.3-d 블록 참조). 하네스 3표면은 2aab5f797 이후 변경됐으므로(§D.3-d `DIFF-01`) 작업 트리 값은 RED 근거가 되지 않는다. 0힛은 빈 stdout이다. **블록 스코프 GREEN 셀(D1·D2·D3·D8)은 예외다**: 워크트리 가드가 git 호출을 이름 붙인 python 명령을 거부하므로(§D.3-d Gaps), 이 셀들은 작업 트리 파일을 git 호출 없는 python 한 호출로 읽는다 — 귀속은 측정 시점 HEAD `0cccc0d31`이며, `.claude/` 트리가 `d36e97571`과 동일함은 `git diff --quiet d36e97571 0cccc0d31 -- .claude/`(exit 0)로 관측했다. 이 라운드에서 실행한 명령의 exit는 `; echo "exit=$?"`로 관측했다(§D.3-d).
- **RED의 올바른 이유**: 각 RED는 "이 SPEC이 바꿀 표면이 오늘 비어 있다"는 이유로 적색이다 — 구현이 그 표면을 채우면 뒤집힌다. 선존재 파일이 못 만지는 wrong-reason red는 없다.
- **명령 형태**: 전부 단일 호출(파이프·리다이렉트·`&&`·`;`·서브셸 없음). **관측 전용 래퍼 `; echo "exit=$?"`는 예외로 허용한다 — 이 라운드의 고정 줄(pinned line)에서 exit 코드를 기록하는 용도에 한하며, 명령 본체는 여전히 단일 호출이다.** grep 일치 0개는 exit 1이다 — "빈 출력 + exit 1"은 완전한 관측이다.
- **비재현 관측 처분**: 재실행 불가능한 관측(예: internal/ 0힛 — 이미 녹색인 부재 클레임)은 회귀 가드로 분류하고 릴리스 블로킹에서 제외한다(undecidable disposition).
- **판정 보류 강등 (plan-audit iter3 CX-7/CX-8 + fresh-run iter2 CX-12 + 리더 재개 CX-13)**: AC-RDX-003/004/005(계측이 구조 면만 전달 — 빈 상수·"Return ok." 프롬프트 mutant 통과)와 AC-RDX-006(전-file 계수 — 산개 언급 mutant 판별 불가)은 verification-completeness §2 채택 기준(계측이 너무 얕아 채택 불가 — mutant-probe adoption bar; §2.1 처분군 적용)에 따라 회귀 가드(판정 보류)로 강등됐다. AC-RDX-017도 동일 처분(CX-12) — 리터럴 쌍(LED-018/019)이 동의어 바꿔쓰기 클래스에 우회됨이 실증돼 의미론 판정은 plan §E7 검토면으로 이관됐다. **AC-RDX-016도 동일 처분(CX-13)** — `source-first` 리터럴 면은 반전 가능(리터럴 유지·규칙 역전)이며 REQ-RDX-013의 의미론 면은 plan §E7 검토면 + 형제 카드 판별기로 이관된다. RED 셀은 전부 측정된 사실로 보존되며 판정은 §7의 형제 카드 계측으로 이관된다. 가드 자체는 게이트가 아니다 — 단 plan §E7 검토면은 REQ-RDX-013·015의 구속 판정면으로 §D.5 항목 6의 종결 게이트다(D4). **분할 강등 (운영자 결정 `d-20261010T000832Z-efde`)**: AC-RDX-001의 LED-001C 보존 면(F8), AC-RDX-008·010의 BP 면(F7: BP-008G·BP-010G), AC-RDX-009의 육면 면(F1·F2: LED-022..027), AC-RDX-014의 블록 면(F3: BLK-014G), AC-RDX-015의 블록 면(F4: BLK-015G)은 판정 보류 관측으로 내려간다. 블록 면을 잃은 AC-RDX-008·010·014·015는 블로킹에서 판정 보류로 옮긴다. 블로킹은 AC-RDX-001(LED-001)·002·007·009(제거면 LED-021 + 계수 LED-028)다. 구조 검사 설계는 후속 카드 소관이다 (`.moai/reports/t1579/followup-card-draft.md`).

## §D AC Matrix

| AC | 분류 | 주장 (Then) | RED-now (LED) | Green path |
|----|------|-------------|---------------|------------|
| AC-RDX-001 | 블로킹 | manifest `domain` 필드가 `Codex CLI upstream change tracking`을 명명 (domain 키 행 스코프 — CX-3) **그리고** `Claude Code` 절을 명명 (D9 — 필드 스코프 보존 절; 이 면은 판정 보류 F8, 분할 efde) | LED-001 (0/1); LED-001C 보존 절 (핀 1 — 녹색; 누락 입력 CTRL-F = 0) | M4 → LED-001 ≥1/0 (블로킹 면); LED-001C ≥1은 판정 보류 관측 (F8) |
| AC-RDX-002 | 블로킹 | manifest `domain` 필드가 best-practices 축을 명명 (동일 스코프) | LED-002 (0/1) | M4 → ≥1/0 |
| AC-RDX-003 | 회귀 가드(판정 보류 — CX-7) | (구조 면) runner에 codex 렌즈 셀렉터가 정의·호출되고 병합·agent 호출까지 흐른다 — top-level과 `run()` 공개 진입점 모두 — 내용 면(프롬프트 실문) 미측정 → 형제 카드 | LED-003 (0/1) + LED-016 (exit 1) + LED-020 (exit 1, no-codex-in-run) | M2 착지 신호(비게이트): grep ≥2/0 AND P3 `dispatch-ok codex=1 total=2`/0 AND P4 `run-ok codex=1 total=1`/0 |
| AC-RDX-004 | 회귀 가드(판정 보류 — CX-7) | (구조 면) runner에 커밋 복원 폴백 앵커 `CODEX_COMMITS_FALLBACK` 존재 — 상수 본문 내용 미측정(빈 문자열 통과) → 형제 카드 | LED-004 (0/1) | M2 착지 신호(비게이트): ≥1/0 |
| AC-RDX-005 | 회귀 가드(판정 보류 — CX-7) | (구조 면) runner에 6테마 체크리스트 앵커 `CODEX_THEME_CHECKLIST` 존재 — 테마 행 실문 미측정 → 형제 카드 | LED-005 (0/1) | M2 착지 신호(비게이트): ≥1/0 |
| AC-RDX-006 | 회귀 가드(판정 보류 — CX-8) | (구조 면) specialist가 codex 상태 파일을 두 사이트에 걸쳐 문서화 — 전-file 계수라 산개 언급 mutant 판별 불가 → 형제 카드 | LED-006 (0/1) + LED-017 (0/1) | M1 착지 신호(비게이트): LED-006 ≥2/0 AND LED-017 ≥1/0 |
| AC-RDX-007 | 블로킹 | specialist가 시드 `rust-v0.161.0`을 기술 | LED-007 (0/1) | M1 → ≥1/0 |
| AC-RDX-008 | 판정 보류 (분할 efde — F7) | specialist에 BP 상시 절차 **섹션 헤딩**이 존재 — 헤딩 행 스코프 `^#+ .*best-practice`(대소문자 무관 — 주석 안 가짜 헤딩은 세므로 판정 보류, F7) | LED-008 (0/1 — 헤딩 행 스코프; 핀 RED `BP-008R`) | M3 → ≥1/0 |
| AC-RDX-009 | 블로킹 (제거면 LED-021 + 계수 LED-028, b80b (a)) · 육면 판정 보류 (F1·F2, 분할 efde) | specialist **Phase 3 블록**(`### Phase 3` ~ `### Phase 4`)이 6종 캐노니컬 전문(`code.claude.com/docs/en/{hooks, sub-agents, skills, plugins, mcp, settings}`)을 각각 보유하고, 파일 전체에서 구형 `docs.anthropic.com` URL은 제거된다 (CX-14 제거면 + CX-18 육면 열거면; D2 — 육면만 블록 스코프, 제거면은 파일 전체) | LED-021 (핀 2aab5f797: `6` / exit `0` — 게이트, §D.3-d `LED-021R`) + LED-028 (핀 2aab5f797: 출력 없음 · exit 1 — 블로킹, §D.3-d `LED-028R`) + LED-022..027 (핀 RED 출력 없음 · exit 1 → 착지 GREEN `1` · exit 0, §D.3-d) | M3 → LED-021 = 0 / exit 1 (착지 SHA 귀속 — 미착지 시 RED); LED-028 = 6 / exit 0 (착지 후 lane 측정 — 6종 각 1 이상, 파일 전체); 육면은 판정 보류 관측 (F1·F2) |
| AC-RDX-010 | 판정 보류 (분할 efde — F7) | BP 섹션 **블록 안**(첫 `best-practice` 헤딩 ~ 다음 동급 헤딩 직전)의 deliverable 행이 `HTML proposal report`를 명명 (D8) | LED-010 (0/1 — 블록 스코프; 핀 RED `BP-010R`) | M3 → ≥1/0 |
| AC-RDX-014 | 판정 보류 (분할 efde — F3) | specialist Phase 0 codex 블록이 codex 상태 파일 부재 시 기본값(`rust-v0.161.0` + 경고)과 키군 4종(REQ-RDX-001)을 문서화 | LED-006 공유 (핀 0 / exit 1) + 블록 앵커 핀 `git grep -c -h -F "**Codex axis state"` 출력 없음 (§D.3-c — 고정 window 아님, D3) | M1 → 판정 보류 관측(비게이트) 블록 스코프 토큰: `### Phase 0`~`### Phase 1` 사이 codex 단락(71행 앵커 ~ 95행)에 기본값 + 경고 + 키군 4종 (B-07·B-08·D3) |
| AC-RDX-015 | 판정 보류 (분할 efde — F4) | runner CODEX_THEME_CHECKLIST 블록이 alpha watch 규범(watch 관찰목록 전용, 안정 탑재 시에만 채택)을 담는다 | LED-005 공유 (핀 0 / exit 1) + 블록 앵커 핀 `git grep -c -h -F "// Standing 6-theme"` 출력 없음 (§D.3-c, D3) | M2 → 판정 보류 관측(비게이트) 블록 스코프 토큰: `// Standing 6-theme` 주석(79행) ~ 닫는 괄호 `];`(92행) 블록 안에 `WATCH-LIST`와 `stable release` 토큰 (watch 규범 82–83행; B-08·D3) |
| AC-RDX-016 | 회귀 가드(판정 보류 — CX-13) | (구조 면) specialist BP 절차가 `source-first` 리터럴을 명명 — 리터럴 유지·규칙 역전 mutant에 우회됨(반전 클래스) → REQ-RDX-013 의미론 면은 plan §E7 검토면 + 형제 카드 판별기로 이관 | LED-014 (0/1) | M3 착지 신호(비게이트): ≥1/0 |
| AC-RDX-017 | 회귀 가드(판정 보류 — CX-12) | (구조 면) Phase 2 조기 종료의 축별 재범위화 — 리터럴 쌍(LED-018/019)은 동의어 바꿔쓰기(paraphrase) 클래스에 우회됨이 실증됨(CX-12) → 의미론 판정은 plan §E7 검토면으로 이관, 형제 카드가 판별기 흡수 | LED-018 (0/1) + LED-019 (1/0 — 구조 참고, 측정 사실 보존) | M1 착지 신호(비게이트): LED-018 ≥1/0 AND LED-019 =0/1 |
| AC-RDX-011 | 회귀 가드 (구 AC-RDX-012 합병 — D7) | (a) `grep -rn "last-codex-version" internal/` — 출력 없음 · exit 1 (Go 라이터 부재 보존, LED-011); (b) `grep -c "last-cc-version.json" .claude/agents/harness/hns-release-update-specialist.md` — 3 이상 (CC 축 절차 보존, LED-012) | — (오늘 녹색: (a) 0힛, (b) 3힛) | 유지 조건: run-phase 전체 |
| AC-RDX-012 | merged into AC-RDX-011 (D7 — 번호 유지, 판정 대상 아님) | — | — | — |
| AC-RDX-013 | 회귀 가드 | manifest의 `hns-release-update-run.js` 참조 1힛 유지 + `sprint_contract` dimensions·thresholds 판독 기준선 일치 (LED-013 + LED-015 — CX-3 판독면) | — (오늘 녹색: 1힛 + LED-015 기준선 출력) | 유지 조건: run-phase 전체 |

**집계 (B-06, 원장 재계수; D7 합병 반영)**: AC 16 = 블로킹 4 (001·002·007·009[제거면 LED-021 + 계수 LED-028]) + 판정 보류 10 (003·004·005·006·008·010·014·015·016·017) + 회귀 가드 2 (011[구 012 합병]·013) — 분할 결정 d-20261010T000832Z-efde 반영. RED-now 앵커 = grep 16 (LED-001·002·003·004·005·006·007·008·009·010·014·017·018·019·021·028) + 검증 동사 2 (LED-016·020, plan §E3-P3·P4). AC-RDX-009 육면 셀(LED-022..027)은 블로킹 AC의 쌍 구성원이므로 위 RED-now 집계에 넣지 않고 §D.3-d에 따로 적는다.

## §D.1 시나리오 (Given-When-Then — 블로킹 4종 + 판정 보류 10종 + 회귀 가드 2종 = 16)

> AC-RDX-003/004/005/006의 시나리오는 구조 면 관측을 기술한다 — CX-7/CX-8 판정 보류로 게이트 밖이며 판정은 형제 카드 계측으로 이관된다(iter3). AC-RDX-017은 CX-12(paraphrase 우회), AC-RDX-016은 CX-13(리터럴 반전)으로 판정 보류 — 의미론 판정은 plan §E7 검토면 + 형제 카드로 이관된다. AC-RDX-008·010·014·015는 분할 결정(efde)으로 판정 보류 — 구조 검사 설계는 후속 카드 소관이다.

- **AC-RDX-001** — **Given** manifest.json이 CC 단일 domain 문자열을 담은 상태로, **When** LED-001 명령(`domain` 키 행 스코프)과 LED-001C 명령(`"domain".*Claude Code`, D9)을 실행하면, **Then** LED-001 일치 개수가 1 이상이고(domain 필드가 codex 축을 명명 — source_request의 동일 문구는 매치 제외, CX-3) LED-001C 일치 개수가 1 이상이다(domain 필드가 Claude Code 축을 명명 — 누락 mutant는 CTRL-F에서 0으로 드러난다). LED-001C 면은 판정 보류(F8, 분할 efde)이고 블로킹 면은 LED-001이다.
- **AC-RDX-002** — **Given** 동일 상태로, **When** LED-002 명령(동일 스코프)을 실행하면, **Then** 일치 개수가 1 이상이다 (domain 필드가 best-practices 축을 명명).
- **AC-RDX-003** — **Given** runner가 CC 렌즈만 fan-out하는 상태로, **When** LED-003 명령과 LED-016(plan §E3-P3), LED-020(plan §E3-P4 `run()` 공개 경로) 명령을 실행하면, **Then** `selectCodexSweepTargets(args)` 출현이 2 이상(정의+top-level 병합 지점)이고, 모의 런타임이 관측한 `codex-release-notes:` 라벨 agent 호출이 top-level과 run() 양쪽에서 1 이상이다 (병합 제외는 LED-016, run() 경로 누락은 LED-020에서 좌초 — CX-5/CX-11; 주석·미연결 정의는 LED-003에서 좌초 — CX-2).
- **AC-RDX-004** — **Given** runner에 커밋 복원 절차가 없는 상태로, **When** LED-004 명령을 실행하면, **Then** `CODEX_COMMITS_FALLBACK` 앵커가 1 이상 관측된다.
- **AC-RDX-005** — **Given** runner에 테마 관찰목록이 없는 상태로, **When** LED-005 명령을 실행하면, **Then** `CODEX_THEME_CHECKLIST` 앵커가 1 이상 관측된다.
- **AC-RDX-006** — **Given** specialist 본문에 codex 상태 절차가 없는 상태로, **When** LED-006 명령과 LED-017 명령을 실행하면, **Then** `last-codex-version.json`이 2 이상(Phase 0 판독·기본값 사이트 + Phase 7a 기록 사이트)이고 `7a-codex` 기록 단계 리터럴이 1 이상이다 (Phase 0 단독·Phase 7a 단독 mutant 모두 좌초 — CX-6).
- **AC-RDX-007** — **Given** AC-RDX-006이 충족된 상태에서도 시드가 빠질 수 있으므로(mutant M-3), **When** LED-007 명령을 실행하면, **Then** `rust-v0.161.0`이 1 이상 관측된다.
- **AC-RDX-008** — **Given** specialist에 BP 축이 없는 상태로, **When** LED-008 명령(헤딩 행 스코프 `^#+ .*best-practice`, 대소문자 무관)을 실행하면, **Then** best-practice 섹션 헤딩이 1 이상 관측된다 (주석·산문에만 있는 단어는 헤딩이 아니므로 세지 않는다 — D8). 단 HTML 주석 안의 가짜 헤딩은 세므로 이 면은 판정 보류다(F7, 분할 efde).
- **AC-RDX-009** — **Given** Phase 3 URL 세트가 docs.anthropic.com 구형 나열인 상태로, **When** LED-021(제거면)·LED-028(6종 캐노니컬 URL 블로킹 계수 — b80b (a))과 LED-022..027(6종 전문 각각, 좌측 경계 + 이스케이프 점 + 종결 경계 패턴 — B-03·D1, Phase 3 블록 스코프 — D2)을 실행하면, **Then** 파일 전체의 `docs.anthropic.com` 계수는 0 / exit 1이고(블로킹 면 LED-021), 6종 캐노니컬 URL 각각의 파일 전체 계수는 1 이상이다(블로킹 면 LED-028 — 6종 전부 삭제 또는 부분 교체 mutant를 잡는다). 육면 블록 스코프 검사(Phase 3 블록 안 6종 각 1 이상, LED-022..027)는 판정 보류 관측이다(F1·F2, 분할 efde). LED-021은 파일 전체의 `docs.anthropic.com` 잔존만 세므로 레거시 잔존만 봉쇄하고, LED-028은 6종의 존재 계수만 세므로(블록 경계는 보지 않고, 좌측 경계는 행두·공백·따옴표·괄호·꺾쇠·대괄호 뒤만 인정하며 백틱 뒤는 인정하지 않는다) 블록 스코프 판정은 판정 보류 면에 남는다. **잔여 위험 선언 (LED-028 전체 파일 계수의 한계):** LED-028 counts the six canonical URLs over the whole file. A mutant that removes them from the Phase 3 block and re-adds them in prose passes LED-028. The check that catches it is the verdict-pending Phase 3 block face (LED-022..027). A block-scoped blocking gate belongs to the structural-check follow-up card. 육면 열거면과 블록 스코프의 봉쇄는 판정 보류로 강등되었다(F1·F2, 분할 efde — 구조 검사는 후속 카드) (CX-14 + CX-18 + D2). 착지 전 HEAD 2ed9d2681에서는 LED-021 = 1이었으므로 그 시점에는 이 시나리오가 RED였다 — **SUPERSEDED**(이력 보존). 고정 커밋 `cb54103ea`에서는 `git grep -c -h "docs.anthropic.com" cb54103ea -- .claude/agents/harness/hns-release-update-specialist.md; echo "exit=$?"`의 출력이 `exit=1`(0힛, 이 라운드 관측)이므로 LED-021 면은 GREEN이다.
- **AC-RDX-010** — **Given** BP 산출물이 명명되지 않은 상태로, **When** LED-010 명령(BP 섹션 블록 스코프 — §D.3-c)을 실행하면, **Then** BP 섹션 블록 안의 deliverable 행에 `HTML proposal report`가 1 이상 관측된다 (블록 밖 산문 언급은 세지 않는다 — D8). 주석 안의 deliverable은 세므로 이 면은 판정 보류다(F7, 분할 efde).
- **AC-RDX-014** — **Given** codex 상태 파일이 존재하지 않는 다음 스윕 실행을 상정하는 상태로, **When** specialist의 Phase 0 codex 블록(§D.3-c, 71행 앵커 ~ 95행 — `### Phase 1` 직전, D3)을 읽으면, **Then** 같은 블록에 (a) 부재 시 기본값 `rust-v0.161.0` + 경고 절차와 (b) 키군 4종(`last_analyzed_version` · `last_analyzed_date` · `last_master_research` · `analysis_history[]`)이 기술돼 있다 (REQ-RDX-004 + REQ-RDX-001 키군 절 — B-07·B-08). 스키마 문서화 단독 통과 mutant는 AC-RDX-006/007과 쌍으로 잡는다. (블록 스코프 면은 판정 보류 — F3, 분할 efde.)
- **AC-RDX-015** — **Given** alpha 테마가 안정에 미탑재 상태로, **When** runner의 CODEX_THEME_CHECKLIST 블록(§D.3-c, 79행 주석 ~ 92행 `];`, D3)을 읽으면, **Then** 같은 블록(79–92행)에 watch 관찰목록 규범("alpha 테마는 채택 아님 — 안정 탑재 시에만 채택 판정")이 기술돼 있다 (REQ-RDX-009 — 1차 스윕 watch 판정의 절차화; B-08). 블록 스코프 면은 판정 보류다(F4, 분할 efde): 토큰 존재는 문장 전체를 증명하지 않는다.
- **AC-RDX-016** — **Given** BP 절차에 원문 선행 강제가 없는 상태로, **When** LED-014 명령을 실행하면, **Then** `source-first` 리터럴이 1 이상 관측된다 (REQ-RDX-013 — mutant M-4의 기계 판정면).
- **AC-RDX-017** — **Given** specialist Phase 2가 무조건 조기 종료 문장을 담은 상태(CC 빈 주간에 codex/BP가 실행 전 종료 — 현재 본문 상태)로, **When** LED-018 명령과 LED-019 명령(제거면)을 실행하면, **Then** `only the CC axis` 리터럴이 1 이상이고 구형 무조건 문장 전문(`If no entries: emit "No new versions since vX.Y.Z" and stop`)은 0이다 — 리터럴만 주석으로 넣고 문장을 생존시키는 mutant는 LED-019에서 좌초한다 (CX-9 + CX-10, REQ-RDX-015).
- **AC-RDX-011** (구 AC-RDX-012 합병 — D7) — **Given** `internal/` 트리에 Go 라이터가 없고 specialist가 CC 축 절차를 보존한 상태로, **When** (a) `grep -rn "last-codex-version" internal/`과 (b) `grep -c "last-cc-version.json" .claude/agents/harness/hns-release-update-specialist.md`를 실행하면, **Then** (a) 출력이 없고 exit 1이다 (0힛 — REQ-RDX-005, mutant M-5), (b) 정수가 3 이상이다 (REQ-RDX-003 보존 축 — 오늘 녹색 3힛). 두 관측 모두 회귀 가드이며 오늘 녹색이라 RED-now 셀이 없다.
- **AC-RDX-012** — merged into AC-RDX-011 (D7). 두 관측 명령은 AC-RDX-011의 (a)·(b)로 이관되었고, 이 번호는 폐기 표시로만 남는다.
- **AC-RDX-013** — **Given** manifest가 골격과 sprint_contract 기준선을 보존한 상태로, **When** `grep -c "hns-release-update-run.js"`와 LED-015 python 판독을 실행하면, **Then** 1힛이고 dimensions·thresholds 출력이 기준선 `['Functionality', 'Consistency'] {'Functionality': 0.85, 'Consistency': 0.8}`과 일치한다 (REQ-RDX-011 — 회귀 가드, 오늘 녹색).

## §D.2 추적성 (AC ↔ REQ)

| AC | REQ | mutant 봉쇄 |
|----|-----|-------------|
| AC-RDX-001/002 | REQ-RDX-010 | M-1 (source_request 기만) — 필드 스코프로 봉쇄 강화 (CX-3); CC 절 누락 mutant — AC-RDX-001 보존 절 LED-001C (D9; 판정 보류 F8) |
| AC-RDX-003 | REQ-RDX-006 | M-2 + M-6 — ≥2 앵커 + §E3-P3 디스패치 관측 면 (CX-2/CX-5); 내용 면은 CX-7 이관(형제 카드) |
| AC-RDX-004 | REQ-RDX-007 | — (구조 면; 폴백 절차 실문은 CX-7 이관 — 형제 카드) |
| AC-RDX-005 | REQ-RDX-008 | — (구조 면; 테마 행 실문은 CX-7 이관 — 형제 카드) |
| AC-RDX-006 | REQ-RDX-001/002/003 | M-7 — 이중 사이트 앵커 + `7a-codex` 기록 단계 (CX-6); 산개 언급 판별은 CX-8 이관(형제 카드) |
| AC-RDX-007 | REQ-RDX-002 | M-3 (seed 누락) |
| AC-RDX-008 | REQ-RDX-012 | 주석 안 가짜 헤딩은 봉쇄되지 않는다 — 판정 보류 (F7, 분할 efde). 헤딩만 있는 껍데기(M-4)는 이 AC로 잡지 못한다 (AC-RDX-016 판정 보류 + E7(d)) |
| AC-RDX-009 | REQ-RDX-014 | LED-021 제거면은 레거시 잔존만 봉쇄 (CX-14); LED-028 계수 면이 6종 캐노니컬 URL 전부 삭제·부분 교체를 블로킹으로 봉쇄한다 (b80b (a)); 육면 열거면·블록 스코프·좌측 경계 면은 판정 보류 (F1·F2, 분할 efde) |
| AC-RDX-010 | REQ-RDX-014 | 블록 밖 산문은 세지 않는다 (D8); 주석 안 deliverable은 판정 보류 (F7, 분할 efde) |
| AC-RDX-014 | REQ-RDX-001 (키군, B-07) · REQ-RDX-004 | M-3의 제3 쌍 — 블록 스코프 면은 판정 보류 (F3, 분할 efde; 구조 검사 후속 카드) |
| AC-RDX-015 | REQ-RDX-009 | alpha-채택 오표기 mutant — 블록 스코프 토큰 면은 판정 보류 (F4, 분할 efde) |
| AC-RDX-016 | REQ-RDX-013 | M-4 — 구조 면; 리터럴 반전 클래스는 CX-13 이관(형제 카드 판별기 + plan §E7 검토면) |
| AC-RDX-017 | REQ-RDX-015 | M-8 — 구조 참고; paraphrase 판별은 CX-12 이관(형제 카드 판별기 + plan §E7 검토면) |
| AC-RDX-011 | REQ-RDX-005 · REQ-RDX-003 (보존 축 — 구 AC-RDX-012 합병) | M-5 (Go 침입) |
| AC-RDX-012 | merged into AC-RDX-011 (D7) | — |
| AC-RDX-013 | REQ-RDX-011 (골격 보존) | threshold 편집은 LED-015 판독면에 걸린다 (CX-3) |

### 게이팅 처분 (B-07) — 블로킹 AC 연결과 강등

분할 결정(efde) 이후 블로킹 AC를 가진 REQ는 002·010·014 세 종뿐이다. REQ-RDX-001·004·009·012는 블로킹 면을 잃어 강등되었다(아래 표). 처분은 다음과 같다 — 강등(판정 보류·회귀 가드), REQ-RDX-013·015는 E7 종결 게이트 편입(D4, §D.5 항목 6). 소유자 `unassigned, leader to issue`는 카드·SPEC id를 발행하지 않은 상태이다. 형제 카드는 t1627로 발행되어 있다(큐 queued, 리더 결정 2026-10-10). 검증 보류 행의 소유자는 t1627이고, 회귀 가드 행은 소유자 미지정 상태를 유지한다.

| REQ | 게이트 상태 | 처분 | 소유자 |
|---|---|---|---|
| REQ-RDX-001 | 없음 — AC-RDX-014 판정 보류 (F3, 분할 efde) | 강등: 판정 보류 가드 | t1627 (queued) |
| REQ-RDX-003 | 없음 — AC-RDX-006 판정 보류 (CX-8); 보존 면은 AC-RDX-011(합병) 회귀 가드 | 강등: 판정 보류 가드 | t1627 (queued) |
| REQ-RDX-004 | 없음 — AC-RDX-014 판정 보류 (F3, 분할 efde) | 강등: 판정 보류 가드 | t1627 (queued) |
| REQ-RDX-005 | 없음 — AC-RDX-011 회귀 가드 (합병 — (a) 부재 클레임, 오늘 녹색, RED-now 셀 없음) | 강등: 회귀 가드 (undecidable disposition) | unassigned, leader to issue |
| REQ-RDX-006 | 없음 — AC-RDX-003 판정 보류 (CX-7); 동작 면 E3-P3·P4는 착지 신호 | 강등: 판정 보류 가드 | t1627 (queued) |
| REQ-RDX-007 | 없음 — AC-RDX-004 판정 보류 (CX-7) | 강등: 판정 보류 가드 | t1627 (queued) |
| REQ-RDX-008 | 없음 — AC-RDX-005 판정 보류 (CX-7) | 강등: 판정 보류 가드 | t1627 (queued) |
| REQ-RDX-009 | 없음 — AC-RDX-015 판정 보류 (F4, 분할 efde) | 강등: 판정 보류 가드 | t1627 (queued) |
| REQ-RDX-011 | 없음 — AC-RDX-013 회귀 가드 (기준선 녹색, RED-now 셀 없음) | 강등: 회귀 가드 | unassigned, leader to issue |
| REQ-RDX-012 | 없음 — AC-RDX-008 판정 보류 (F7, 분할 efde) | 강등: 판정 보류 가드 | t1627 (queued) |
| REQ-RDX-013 | 블로킹 AC 없음 — AC-RDX-016 판정 보류 (CX-13); 구속 판정은 plan §E7(d) → §D.5 항목 6 (증거 항목 (d)) | E7 종결 게이트 편입 (D4) | t1627 (queued) |
| REQ-RDX-015 | 블로킹 AC 없음 — AC-RDX-017 판정 보류 (CX-12); 구속 판정은 plan §E7(a)–(c) → §D.5 항목 6 (증거 항목 (d): E7(a)) | E7 종결 게이트 편입 (D4) | t1627 (queued) |

블로킹 AC를 가진 REQ는 3종이다: 002 → AC-007, 010 → AC-001·002, 014 → AC-009(제거면 LED-021 + 계수 LED-028). 004 → AC-014, 009 → AC-015, 012 → AC-008은 판정 보류라 블로킹 AC가 없다(분할 결정 efde).

## §D.3 증거 원장 (Evidence Ledger — RED 귀속 `2aab5f797` · GREEN 귀속 측정 SHA; 2026-10-10 수리 셀은 §D.3-c · §D.3-d)

각 행: 명령은 축자 그대로 단일 실행됐고, stdout은 같은 실행에서 관측했고, exit code는 출력 유무와 문서화된 상태로 귀속했다(§A '귀속 분리'). 경로는 워크트리 루트 기준. 분할 결정(efde) 이후 LED-001C·LED-022..027·BLK-014G·BLK-015G·BP-008G·BP-010G 면은 판정 보류 관측이다. 블로킹 면은 LED-001·002·007·LED-021R/T·LED-028다.

| LED | 명령 (단일 호출) | stdout (축자) | exit | 판정 |
|-----|------------------|---------------|------|------|
| LED-001 | `git grep -c -h -E '"domain".*Codex CLI upstream change tracking' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/commands/harness/release-update/manifest.json` | (출력 없음) | `1` | RED (AC-001) — 핀 2aab5f797 git grep 귀속 (MP-8 재고정); HEAD GREEN은 §D.3-d LED-001 G — domain 필드 스코프 (CX-3 재앵커) |
| LED-001C | `grep -c '"domain".*Claude Code' .claude/commands/harness/release-update/manifest.json` | `1` | `0` | 보존 절 (AC-001, D9; 판정 보류 F8) — 핀 2aab5f797 = 1 (녹색: CC 절은 M4 이전부터 존재 — 보존 면); 누락 입력은 §D.3-d CTRL-F = 0 (exit 1) |
| LED-002 | `git grep -c -h -E '"domain".*best-practices axis' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/commands/harness/release-update/manifest.json` | (출력 없음) | `1` | RED (AC-002) — 핀 2aab5f797 git grep 귀속 (MP-8 재고정); HEAD GREEN은 §D.3-d LED-002 G — 동일 재앵커 (CX-3) |
| LED-003 | `grep -c "selectCodexSweepTargets(args)" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-003) — 디스패치 호출 앵커, 착지 후 ≥2 (CX-2 재앵커) |
| LED-004 | `grep -c "CODEX_COMMITS_FALLBACK" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-004) · GREEN 목표 `≥1` (AC-RDX-004 문턱) · 관측 개수: cb54103ea `2`, 이 라운드 작업 트리 `4` — 고정 명령·출력은 §D.3-d LED-004-현행 블록 |
| LED-005 | `grep -c "CODEX_THEME_CHECKLIST" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-005) |
| LED-006 | `grep -c "last-codex-version.json" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-006) |
| LED-007 | `git grep -c -h -F rust-v0.161.0 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md` | (출력 없음) | `1` | RED (AC-007) — 핀 2aab5f797 git grep 귀속 (MP-8 재고정); HEAD GREEN은 §D.3-d LED-007 G |
| LED-008 | `grep -ci -E "^#+ .*best-practice" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-008) — 헤딩 행 스코프 (D8); 핀 2aab5f797 = 출력 없음 (§D.3-d `BP-008R`) |
| LED-009 | `grep -c "code.claude.com" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED — 게이트 아님 (전체 파일 계수, 산문 포함 — CX-14). 핀 2aab5f797 재실행 = 출력 없음 (§D.3-d RED-summary) |
| LED-010 | 블록 스코프 — BP 섹션 블록 안 `HTML proposal report` 계수 (§D.3-d `BP-010R` / `BP-010G`; 작업 트리 전체 계수 아님) | `0` | `1` | RED (AC-010) — 블록 밖 산문 mutant 봉쇄 (D8) |
| LED-014 | `grep -c "source-first" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-016) |
| LED-015 | `python3 -c "import json;d=json.load(open('.claude/commands/harness/release-update/manifest.json'));sc=d['sprint_contract'];print(sc['dimensions'],sc['thresholds'])"` | `['Functionality', 'Consistency'] {'Functionality': 0.85, 'Consistency': 0.8}` | `0` | 회귀 가드 기준선 — 출력 불변 유지가 PASS (AC-013, CX-3 판독면) |
| LED-016 | plan §E3-P3 verb 축자 (모의-런타임 실행 — `node -e '...'`, plan §E3-P3 블록 참조) | stderr `REJECTED: no-codex-dispatch:1` | `1` | RED (AC-003 병합 관측면) — 현재 러너 디스패치는 CC 호출 1건, codex 라벨 0건 (CX-5, M2에서 `dispatch-ok codex=1 total=2`/exit 0으로 뒤집음) |
| LED-017 | `grep -c "7a-codex" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-006 기록 사이트) — Phase 7a 기록 단계 부재 (CX-6) |
| LED-018 | `grep -c "only the CC axis" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-017) — Phase 2 무조건 조기 종료 생존, codex/BP가 CC 널 주간에 실행 전 종료 (CX-9) |
| LED-019 | `grep -c 'If no entries: emit "No new versions since vX.Y.Z" and stop' .claude/agents/harness/hns-release-update-specialist.md` | `1` | `0` | RED-제거면 (AC-017) — 구형 무조건 문장 생존; 착지 후 0/exit 1이 PASS — 주석 포함 생존 전부 적색 (CX-10) |
| LED-020 | plan §E3-P4 verb 축자 (`node -e '...'` — `run()` 공개 경로, codex 전용 입력) | stderr `REJECTED: no-codex-in-run:0` | `1` | RED (AC-003 run() 진입점) — codex 전용 입력에서 run()이 agent 호출 0건 (CX-11, M2에서 `run-ok codex=1 total=1`/exit 0으로 뒤집음) |
| LED-021 | 핀 RED — §D.3-d `LED-021R`: `git grep -c -h "docs.anthropic.com" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md` | `6` | `0` | RED-제거면 (AC-009 게이트, B-01) — 핀 2aab5f797 귀속. HEAD 관측(`2ed9d2681`) = `1` / exit `0` (산문 179행, 미착지) |
| LED-021T | 착지 목표 — §D.3-d `LED-021T`: `grep -c "docs.anthropic.com" .claude/agents/harness/hns-release-update-specialist.md` | 목표 `0` | 목표 `1` | 착지 목표 셀 (AC-009 게이트, plan §F M3 항목 5) — 라벨 SHA: `cb54103ea` (M3 항목 5 착지 — `git show cb54103ea`의 specialist.md 179행에서 `docs.anthropic.com` 잔존 제거 확인). 고정 커밋 측정 완료 — 출력 `exit=1`(0힛); 명령은 아래 LED-021T 블록 |
| LED-028 | 펜스 §D.3-d `LED-028R` → `LED-028G` (6종 캐노니컬 URL 각 1 이상 · 파일 전체 · 단일 호출 — 블로킹, b80b (a)) | 핀 RED `(출력 없음)` → 착지 GREEN: lane 착지 후 측정 | 핀 1 / 착지 0 | 블로킹 — 6종 캐노니컬 URL 존재 계수 (AC-RDX-009; 결정 b80b (a)). 핀 2aab5f797 = 출력 없음 · exit 1 (RED). 착지 GREEN 측정: 고정 커밋 `cb54103ea`(클린 트리) = 6 · exit 0 (이 라운드 — §D.3-d LED-028G-현행). 2ed9d2681 귀속은 SUPERSEDED — 이력 LED-028G 보존 |
| LED-022 | 펜스 §D.3-d `LED-022R` → `LED-022G` (좌측 경계 + 이스케이프 점 + 종결 경계; GREEN은 Phase 3 블록 스코프 python — D1·D2) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED exit 1 · 착지 GREEN exit 0 (B-03 패턴, B-04 분리) |
| LED-023 | 펜스 §D.3-d `LED-023R` → `LED-023G` (좌측 경계 + 이스케이프 점 + 종결 경계; GREEN은 Phase 3 블록 스코프 python — D1·D2) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED exit 1 · 착지 GREEN exit 0 (B-03 패턴, B-04 분리) |
| LED-024 | 펜스 §D.3-d `LED-024R` → `LED-024G` (좌측 경계 + 이스케이프 점 + 종결 경계; GREEN은 Phase 3 블록 스코프 python — D1·D2) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED exit 1 · 착지 GREEN exit 0 (B-03 패턴, B-04 분리) |
| LED-025 | 펜스 §D.3-d `LED-025R` → `LED-025G` (좌측 경계 + 이스케이프 점 + 종결 경계; GREEN은 Phase 3 블록 스코프 python — D1·D2) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED exit 1 · 착지 GREEN exit 0 (B-03 패턴, B-04 분리) |
| LED-026 | 펜스 §D.3-d `LED-026R` → `LED-026G` (좌측 경계 + 이스케이프 점 + 종결 경계; GREEN은 Phase 3 블록 스코프 python — D1·D2) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED exit 1 · 착지 GREEN exit 0 (B-03 패턴, B-04 분리) |
| LED-027 | 펜스 §D.3-d `LED-027R` → `LED-027G` (좌측 경계 + 이스케이프 점 + 종결 경계; GREEN은 Phase 3 블록 스코프 python — D1·D2) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED exit 1 · 착지 GREEN exit 0 (B-03 패턴, B-04 분리) |
| LED-011 | `grep -rn "last-codex-version" internal/` | (출력 없음) | `1` | 회귀 가드 기준선 — 0힛 유지가 PASS (AC-RDX-011 (a) — 합병 D7) |
| LED-012 | `grep -c "last-cc-version.json" .claude/agents/harness/hns-release-update-specialist.md` | `3` | `0` | 회귀 가드 기준선 — ≥3 유지가 PASS (AC-RDX-011 (b) — 구 AC-012 합병, D7) |
| LED-013 | `grep -c "hns-release-update-run.js" .claude/commands/harness/release-update/manifest.json` | `1` | `0` | 회귀 가드 기준선 — 1 유지가 PASS (AC-013) |

**LED-001/002/003 재앵커 근거**: plan-audit iter1(CX-2/CX-3)으로 위 세 행의 명령을 교체했다 — (B-04 정정) 이 자리에 있던 '2aab5f797 이후 바이트 동일' 주장은 거짓이었다. 하네스 3표면은 2aab5f797 이후 변경됐고(§D.3-d `DIFF-01`), RED 귀속은 2aab5f797 `git grep` 재실행만 인정한다(§D.3-d RED-summary). **LED-016/017 신설 근거**: plan-audit iter2(CX-5/CX-6) — 같은 하네스 표면에서 본 실행 측정. **LED-018 신설 근거**: plan-audit iter3(CX-9) — 동일 측정 조건. **LED-019/020 신설 근거**: fresh post-split run(CX-10/CX-11) — 동일 측정 조건. **LED-021 신설 근거**: 리더 재개 재심(rcpt-0e776d398b9dd8d642d5f3ac, CX-14) — 동일 측정 조건. **LED-022..027 신설 근거**: 클로저 run(rcpt-ce61f029a011f5fa5ff6bead, CX-18) — 핀 2aab5f797에서는 0(RED, 출력 없음)이고 착지 d36e97571에서 1(GREEN)이다. 분리 기록은 §D.3-d(B-04)이며 게이트는 LED-021(B-01)이다. 나머지 LED 행은 원본 그대로다. LED-015의 세미콜론은 인용된 python 프로그램 내부의 것 — 셸 구분자가 아니므로 단일 호출 규약을 유지한다. LED-016/020의 오류 메시지는 어댑터 자체의 핸들러가 내는 결정적 한 줄이다(전체 스택 대신).

**라운드 3 수리 근거 (D1·D2·D3·D4·D7·D8·D9)**: LED-022..027은 좌측 경계 패턴(D1)과 Phase 3 블록 스코프(D2)로 재작성했다. BLK-014·015와 §D.3-c의 고정 window `-A N`·`-B N -A N`은 구조 경계(앵커·헤딩·닫는 괄호)로 교체했다(D3). LED-008·010과 BP-008·010은 헤딩 행·BP 섹션 블록으로 스코프했다(D8). LED-001C는 D9 보존 절이다. AC-RDX-011·012 합병으로 17 → 16(D7). REQ-RDX-013·015는 §D.5 항목 6의 E7 종결 게이트에 편입했다(D4).

**보조 관측 (동일 트리)**: `grep -c "Codex" manifest.json` → `0`/exit 1 · `grep -ci "codex" runner` → `0`/exit 1 · `grep -c "HTML" specialist.md` → `0`/exit 1 · `grep -rn "last-cc-version" internal/` → 출력 없음/exit 1 (Go 라이터 부재 — 상태 파일이 하네스 계층 소유임의 근거, spec.md §1.1 M4). `git rev-parse --short HEAD` → `2aab5f797`.

### §D.3-c 블록 경계 (구조 경계 — AC-RDX-014 · AC-RDX-015 · AC-RDX-008 · AC-RDX-010; B-08 · D3 · D8)

- **AC-RDX-014 블록**: `.claude/agents/harness/hns-release-update-specialist.md`의 Phase 0 codex 단락. 구조 경계: 포함 영역은 `### Phase 0 — Load State` 헤딩(49행)부터 `### Phase 1` 헤딩(96행) 직전까지이고, 블록은 첫 행 앵커 `**Codex axis state`(71행)부터 `### Phase 1` 헤딩 직전까지(71–95행)다. 게이트는 고정 window(`-A N`)를 쓰지 않는다 — 판정 토큰 6종(`rust-v0.161.0`, `emit warning`, `last_analyzed_version`, `last_analyzed_date`, `last_master_research`, `analysis_history[]`)의 블록 내 개수가 모두 1 이상이면 GREEN이다(§D.3-d BLK-014G). Phase 7a(Step 7a-codex)의 출현은 블록 밖이므로 판정에서 제외한다.
- **AC-RDX-015 블록**: `.claude/workflows/hns-release-update-run.js`의 CODEX_THEME_CHECKLIST 블록. 구조 경계: 시작은 주석 `// Standing 6-theme …` 행(79행), 끝은 상수 선언(85행) 이후 처음 나오는 닫는 괄호 행 `];`(92행)이다 — 블록은 79–92행. 게이트는 고정 window(`-B N -A N`)를 쓰지 않는다 (§D.3-d BLK-015G 참조). 115행의 `join` 참조는 블록 밖이므로 판정에서 제외한다.
- **AC-RDX-008 블록 (BP 헤딩 행)**: `.claude/agents/harness/hns-release-update-specialist.md`의 헤딩 행 스코프 — `^#+ .*best-practice`(대소문자 무관)가 헤딩 행(현 트리 299행 `## Best-Practices Axis …`)에 1 이상이면 GREEN이다. 주석·산문의 같은 단어는 헤딩이 아니므로 세지 않는다.
- **AC-RDX-010 블록 (BP 섹션 본문)**: 블록 = 첫 `best-practice` 헤딩 행부터 같거나 상위 레벨의 다음 헤딩 직전까지(현 트리 299–319행). 블록 안의 `HTML proposal report` 개수가 1 이상이면 GREEN이다(현 트리 316행 deliverable 행).
- **판독 규칙**: 블록 안 행만 판정한다. AC-RDX-014는 기본값 `rust-v0.161.0`, 경고(`emit warning`), 키군 4종 토큰이 71–95행 블록 안에 있을 때 GREEN이다. AC-RDX-015는 `WATCH-LIST`와 `stable release` 토큰(82–83행)이 79–92행 블록 안에 있을 때 GREEN이다. 네 블록 모두 핀 2aab5f797에서는 앵커(또는 헤딩)가 없어 출력이 없다(RED). 블록 경계는 python 한 호출이 헤딩·앵커·닫는 괄호로 계산하며, 블록 밖 행은 세지 않는다(§D.3-d GREEN 셀). 분할 결정(efde) 이후 이 네 블록(AC-RDX-014·015·008·010)의 판정은 판정 보류 관측이다 — 블록 경계 파서는 후속 카드 소관이다.
- **잔여 위험**: 블록 안에 문장이 있는지의 판정은 여전히 판독이다 — 토큰의 존재는 규범 문장의 존재를 증명하지 않는다. 기계화되지 않은 이 판정은 plan §E1 검토와 §E7 의미 검토 면이 제2 판정으로 남으며, §E7 검토 기록은 §D.5 항목 6의 종결 증거 (d)다(D4). 헤딩 regex는 펜스 안의 `# ` 행을 구분하지 않는다 — `best-practice`를 포함한 `#` 행은 현 트리에서 299행 헤딩 하나뿐이다(grep 실측, §D.3-d BP-008G).

### §D.3-d 펜스 원장 — 수리 셀 (B-01 · B-02 · B-03 · B-04 · B-08 · 라운드 3: D1 · D2 · D3 · D8 · D9)

G 블록 귀속(0cccc0d31·d36e97571·2ed9d2681)은 `.claude/` 트리 동일성으로 묶인다(git diff --quiet d36e97571 2ed9d2681 -- .claude/ exit 0). 수리 커밋은 hns-release-update-specialist.md 179행만 바꾼다. 감사 대상 커밋(`cb54103ea`)의 재측정 값은 이 파일에 고정 커밋 명령과 출력으로 기록한다 (LED-021T · LED-028G-현행 · LED-001/002/007 G의 pinned 행, plan-audit iter6 M-1). 운영자 브리프는 저장소에 없으므로 참조하지 않는다.

명령은 축자다. `(출력 없음)`은 git grep의 0힛이며, exit 1은 문서화된 git grep 상태다(Gaps 참조). 모든 행은 귀속 SHA를 명시한다. 표 셀의 명령은 식별용 축약이며, 실행 가능한 축자 명령은 이 펜스에 있다(verification-completeness §2.1 — 표 셀은 셸 메타문자를 훼손할 수 있다).

```text
RED-summary  git grep -c -h "<anchor>" 2aab5f797b75983e132af451da68f69e3426557b -- <path>   (re-run 2026-10-10; exit는 `; echo "exit=$?"`로 관측)
  no output (0 hits, exit 1): LED-001 LED-002 LED-003 LED-004 LED-005 LED-006 LED-007 LED-008(-i) LED-009 LED-010 LED-014 LED-017 LED-018 BLK-014R BLK-015R BP-008R BP-010R
  LED-019 (legacy unconditional stop sentence): 1 (exit 0)
  LED-021 (legacy domain): 6 (exit 0)
  LED-028 (six canonical URLs, blocking count, whole-file pinned absence): no output (exit 1)
  LED-022..027 (six canonical URLs, boundary-anchored, whole-file pinned absence): no output x6 (exit 1)
  LED-001C (domain Claude Code clause, preservation): 1 (exit 0 — green pinned by construction; the failing input is CTRL-F)

LED-021R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-01)
  cmd:     git grep -c -h "docs.anthropic.com" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  6
  exit:    0

LED-021G   tree 2ed9d2681f6825acb46d298362aa8f9f6fb9d6d6 (HEAD, pre-repair)   observed at HEAD (not a pass)
  cmd:     git grep -c -h "docs.anthropic.com" 2ed9d2681f6825acb46d298362aa8f9f6fb9d6d6 -- .claude/agents/harness/hns-release-update-specialist.md   (HEAD = 2ed9d2681f6825acb46d298362aa8f9f6fb9d6d6)
  stdout:  1   (line 179, prose)
  exit:    0

LED-021T   label SHA: cb54103ea (M3 item-5 landing — `git show cb54103ea -- .claude/agents/harness/hns-release-update-specialist.md`: 179행 `docs.anthropic.com` 잔존 제거, `--stat` 2 +-)
  cmd:     grep -c "docs.anthropic.com" .claude/agents/harness/hns-release-update-specialist.md
  target:  stdout 0, exit 1
  pinned (cb54103ea, this round): `git grep -c -h "docs.anthropic.com" cb54103ea -- .claude/agents/harness/hns-release-update-specialist.md; echo "exit=$?"` → verbatim output `exit=1` (no count line = zero hits; the working-tree `grep -c` above reads `0` / exit 1 on the same clean tree)
  status:  MET at cb54103ea (pinned, this round — see the pinned line above). **SUPERSEDED**: "NOT MET at 2ed9d2681 (HEAD) (LED-021G = 1)" — 2ed9d2681 is the pre-repair commit, kept as history in LED-021G

LED-028R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, b80b (a) — six canonical URLs, whole-file blocking count)
  cmd:     git grep -c -h -F 'https://code.claude.com/docs/en/' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (출력 없음)
  exit:    1
  note:    고정 문자열 형태는 기준보다 느슨하다(정식 문서 URL 아무 것이나 계수하며 여섯 종에 한정되지 않는다); 핀 시점의 0은 여섯 종의 0을 함의하므로 이 RED는 유효하다.
LED-028G   tree 2ed9d2681f6825acb46d298362aa8f9f6fb9d6d6 (HEAD, pre-repair)   GREEN (블로킹 면, b80b (a))
  cmd:     python3 -c 'import re,sys;t=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read();u=["hooks","sub-agents","skills","plugins","mcp","settings"];m=[k for k in u if re.search(r"(?:^|[\s\x27\"(<\[])https://code\.claude\.com/docs/en/"+re.escape(k)+r"(?![A-Za-z0-9_./-])",t,re.M)];print(len(m));sys.exit(0 if len(m)==6 else 1)'
  stdout:  6
  exit:    0
  note:    인용된 python 프로그램 내부의 `;`는 셸 구분자가 아니다 (LED-015 규약 — 단일 호출 형태 유지). 6종 존재 계수이며 좌측 경계(행두·공백·따옴표·괄호·꺾쇠·대괄호 뒤만 인정, 백틱 뒤 제외)와 우측 경계를 본다. 블록 경계는 보지 않는다 (블록 스코프 면 LED-022..027은 판정 보류로 분리).

LED-028G-현행   tree cb54103ea (pinned — clean worktree at HEAD cb54103ea; `git status --short` 공집합; this round)   GREEN (블로킹 면, b80b (a)) — 2ed9d2681 귀속은 SUPERSEDED, 이력 LED-028G 보존
  cmd:     python3 -c 'import re,sys;t=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read();u=["hooks","sub-agents","skills","plugins","mcp","settings"];m=[k for k in u if re.search(r"(?:^|[\s\x27\"(<\[])https://code\.claude\.com/docs/en/"+re.escape(k)+r"(?![A-Za-z0-9_./-])",t,re.M)];print(len(m));sys.exit(0 if len(m)==6 else 1)'; echo "exit=$?"
  stdout:  6   verbatim: `6` then `exit=0`
  note:    이 라운드에 cb54103ea 클린 트리에서 재실행했다. 블록 경계는 보지 않는다 (블록 스코프 면 LED-022..027은 판정 보류로 분리).

LED-004-현행   tree cb54103ea (pinned, this round)   GREEN 목표 `≥1` (AC-RDX-004 — 구조 면, 비게이트)
  cmd (pinned):   git grep -c -h "CODEX_COMMITS_FALLBACK" cb54103ea -- .claude/workflows/hns-release-update-run.js; echo "exit=$?"
  stdout / exit:  `2` / `exit=0` (cb54103ea의 77행 상수 선언, 111행 렌즈 프롬프트 보간)
  cmd (working tree, this round):   grep -c "CODEX_COMMITS_FALLBACK" .claude/workflows/hns-release-update-run.js; echo "exit=$?"
  stdout / exit:  `4` / `exit=0` (72행 상수 선언, 73행 STEPS 선언, 113행 보간, 114행 STEPS 결합)
  note:    두 관측 모두 식별자가 나타나는 행 수다. 작업 트리의 4는 C-3 수리로 STEPS 상수가 추가된 결과이며, GREEN 문턱 ≥1을 충족한다.

LED-022R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E "(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-022G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (Phase 3 block, python — D2)
  cmd:     python3 -c 'import re;b=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read();b=b[b.index("### Phase 3"):b.index("### Phase 4")];n=len(re.findall(r"(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)",b));print(n);raise SystemExit(0 if n else 1)'
  stdout:  1   exit: 0

LED-023R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E "(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/sub-agents([^A-Za-z0-9_./-]|$)" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-023G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (Phase 3 block, python — D2)
  cmd:     python3 -c 'import re;b=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read();b=b[b.index("### Phase 3"):b.index("### Phase 4")];n=len(re.findall(r"(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/sub-agents([^A-Za-z0-9_./-]|$)",b));print(n);raise SystemExit(0 if n else 1)'
  stdout:  1   exit: 0

LED-024R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E "(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/skills([^A-Za-z0-9_./-]|$)" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-024G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (Phase 3 block, python — D2)
  cmd:     python3 -c 'import re;b=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read();b=b[b.index("### Phase 3"):b.index("### Phase 4")];n=len(re.findall(r"(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/skills([^A-Za-z0-9_./-]|$)",b));print(n);raise SystemExit(0 if n else 1)'
  stdout:  1   exit: 0

LED-025R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E "(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/plugins([^A-Za-z0-9_./-]|$)" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-025G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (Phase 3 block, python — D2)
  cmd:     python3 -c 'import re;b=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read();b=b[b.index("### Phase 3"):b.index("### Phase 4")];n=len(re.findall(r"(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/plugins([^A-Za-z0-9_./-]|$)",b));print(n);raise SystemExit(0 if n else 1)'
  stdout:  1   exit: 0

LED-026R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E "(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/mcp([^A-Za-z0-9_./-]|$)" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-026G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (Phase 3 block, python — D2)
  cmd:     python3 -c 'import re;b=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read();b=b[b.index("### Phase 3"):b.index("### Phase 4")];n=len(re.findall(r"(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/mcp([^A-Za-z0-9_./-]|$)",b));print(n);raise SystemExit(0 if n else 1)'
  stdout:  1   exit: 0

LED-027R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E "(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/settings([^A-Za-z0-9_./-]|$)" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-027G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (Phase 3 block, python — D2)
  cmd:     python3 -c 'import re;b=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read();b=b[b.index("### Phase 3"):b.index("### Phase 4")];n=len(re.findall(r"(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/settings([^A-Za-z0-9_./-]|$)",b));print(n);raise SystemExit(0 if n else 1)'
  stdout:  1   exit: 0

CTRL-A  spoof reproduced (unescaped dots) — the defect
  cmd:     printf 'codeXclaudeXcom/docs/en/hooks\n' | grep -c "code.claude.com/docs/en/hooks"
  stdout:  1
CTRL-A2  diagnostic pair of CTRL-B (O2): unescaped https-prefixed spoof must print 1 — the defect with the scheme present
  cmd:     printf 'https://codeXclaudeXcom/docs/en/hooks\n' | grep -c -E 'https://code.claude.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  1   exit: 0
CTRL-B  negative control (B-03·O2, required): https-prefixed escaped input must print 0 — the zero reflects dot escaping, not a missing scheme
  cmd:     printf 'https://codeXclaudeXcom/docs/en/hooks\n' | grep -c -E '(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  0   exit: 1
CTRL-C  boundary control: prefix variant must not match (shipped pattern)
  cmd:     printf 'https://code.claude.com/docs/en/hooks-invalid\n' | grep -c -E '(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  0   exit: 1
CTRL-D  positive control: real URL before a closing paren must match (shipped pattern)
  cmd:     printf 'https://code.claude.com/docs/en/hooks)\n' | grep -c -E '(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  1   exit: 0
CTRL-E0  defect reproduction (D1): the OLD right-boundary-only pattern counts a nested spoof
  cmd:     printf 'https://attacker.invalid/https://code.claude.com/docs/en/hooks\n' | grep -c -E 'https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  1   exit: 0
CTRL-E   negative control (D1, required): the nested spoof must print 0 under the shipped pattern
  cmd:     printf 'https://attacker.invalid/https://code.claude.com/docs/en/hooks\n' | grep -c -E '(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  0   exit: 1
CTRL-F   negative control (D9, required): a domain value without the Claude Code clause must not match
  cmd:     printf '"domain": "Codex CLI upstream change tracking and the best-practices axis"\n' | grep -c -E '"domain".*Claude Code'
  stdout:  0   exit: 1
CTRL-P3a  block-scope control (D2): a canonical URL OUTSIDE the Phase 3 block must not count (synthetic input)
  cmd:     python3 -c 'import re;s="### Phase 3 x\nprose\n### Phase 4 y\nhttps://code.claude.com/docs/en/hooks\n";b=s[s.index("### Phase 3"):s.index("### Phase 4")];n=len(re.findall(r"(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)",b));print(n);raise SystemExit(0 if n else 1)'
  stdout:  0   exit: 1
CTRL-P3b  block-scope control (D2): a canonical URL INSIDE the Phase 3 block must count (synthetic input)
  cmd:     python3 -c 'import re;s="### Phase 3 x\nhttps://code.claude.com/docs/en/hooks\n### Phase 4 y\n";b=s[s.index("### Phase 3"):s.index("### Phase 4")];n=len(re.findall(r"(^|[^A-Za-z0-9_./:])https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)",b));print(n);raise SystemExit(0 if n else 1)'
  stdout:  1   exit: 0

BLK-014R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-08 · D3 — codex 단락 앵커 부재)
  cmd:     git grep -c -h -F "**Codex axis state" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
BLK-014G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (block 71–95 = anchor 71 .. `### Phase 1` 96 exclusive; per-token counts)
  cmd:     python3 -c 'L=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read().split("\n");p0=[i for i,l in enumerate(L) if l.startswith("### Phase 0")][0];p1=[i for i,l in enumerate(L) if l.startswith("### Phase 1")][0];a=[i for i in range(p0,p1) if "**Codex axis state" in L[i]];b="\n".join(L[a[0]:p1]) if a else "";t=[b.count(k) for k in ["rust-v0.161.0","emit warning","last_analyzed_version","last_analyzed_date","last_master_research","analysis_history[]"]];print(t);raise SystemExit(0 if a and min(t)>=1 else 1)'
  stdout:  [4, 1, 3, 2, 2, 1]
  exit:    0

BLK-015R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-08 · D3 — 주석 앵커 부재)
  cmd:     git grep -c -h -F "// Standing 6-theme" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/workflows/hns-release-update-run.js
  stdout:  (no output)   exit: 1
BLK-015G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (block 79–92 = comment 79 .. closing `];` 92; per-token counts)
  cmd:     python3 -c 'L=open(".claude/workflows/hns-release-update-run.js",encoding="utf-8").read().split("\n");s=[i for i,l in enumerate(L) if l.startswith("// Standing 6-theme")];d=[i for i,l in enumerate(L) if l.startswith("const CODEX_THEME_CHECKLIST")];e=[j for j in range(d[0],len(L)) if L[j].strip()=="];"] if d else [];b="\n".join(L[s[0]:e[0]+1]) if s and e else "";t=[b.count(k) for k in ["WATCH-LIST","stable release"]];print(t);raise SystemExit(0 if s and e and min(t)>=1 else 1)'
  stdout:  [1, 1]
  exit:    0

BP-008R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, D8 — heading absent)
  cmd:     git grep -c -h -i -E "^#+ .*best-practice" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
BP-008G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (heading row scope)
  cmd:     grep -c -i -E "^#+ .*best-practice" .claude/agents/harness/hns-release-update-specialist.md
  stdout:  1   exit: 0
BP-010R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, D8 — deliverable absent)
  cmd:     git grep -c -h -F "HTML proposal report" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
BP-010G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN (BP section block, python — D8)
  cmd:     python3 -c 'import re;L=open(".claude/agents/harness/hns-release-update-specialist.md",encoding="utf-8").read().split("\n");h=[i for i,l in enumerate(L) if re.match(r"#+ .*best-practice",l,re.I)];v=(len(L[h[0]])-len(L[h[0]].lstrip("#"))) if h else 0;e=[j for j in range(h[0]+1,len(L)) if re.match(r"#{1,%d} " % v,L[j])] if h else [];b="\n".join(L[h[0]:(e[0] if e else len(L))]) if h else "";n=b.count("HTML proposal report");print(n);raise SystemExit(0 if n else 1)'
  stdout:  1   exit: 0
LED-001C R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED-preservation (pinned green by construction — the CC clause predates M4)
  cmd:     git grep -c -h -E '"domain".*Claude Code' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/commands/harness/release-update/manifest.json
  stdout:  1   exit: 0
LED-001C G   tree 0cccc0d31 (working tree; .claude/ = d36e97571a4d39b3186a71781be9bf01f9434d7f)   GREEN
  cmd:     grep -c '"domain".*Claude Code' .claude/commands/harness/release-update/manifest.json
  stdout:  1   exit: 0

LED-001 R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned — MP-8 재고정; acceptance §D.3 LED-001 행)
  cmd:     git grep -c -h -E '"domain".*Codex CLI upstream change tracking' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/commands/harness/release-update/manifest.json
  stdout:  (no output)   exit: 1
LED-001 G   tree 18ea52c0a (M4 착지 커밋 — `git cat-file -t 18ea52c0a` = commit; 착지 재관측 `git grep -c -h -E '"domain".*Codex CLI upstream change tracking' 18ea52c0a -- .claude/commands/harness/release-update/manifest.json; echo "exit=$?"` → `1`, `exit=0`) · 고정 커밋 cb54103ea 재측정 (아래 pinned 행) · 2ed9d2681 측정과 "(operator brief)" 재측정 주장은 SUPERSEDED (브리프는 저장소에 없음)   GREEN (M4 착지)
  pinned (cb54103ea, this round): git grep -c -h -E '"domain".*Codex CLI upstream change tracking' cb54103ea -- .claude/commands/harness/release-update/manifest.json; echo "exit=$?" → verbatim `1` then `exit=0`
  cmd:     grep -c '"domain".*Codex CLI upstream change tracking' .claude/commands/harness/release-update/manifest.json
  stdout:  1   exit: 0

LED-002 R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned — MP-8 재고정)
  cmd:     git grep -c -h -E '"domain".*best-practices axis' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/commands/harness/release-update/manifest.json
  stdout:  (no output)   exit: 1
LED-002 G   tree 18ea52c0a (M4 착지 커밋 — `git cat-file -t 18ea52c0a` = commit; 착지 재관측 `git grep -c -h -E '"domain".*best-practices axis' 18ea52c0a -- .claude/commands/harness/release-update/manifest.json; echo "exit=$?"` → `1`, `exit=0`) · 고정 커밋 cb54103ea 재측정 (아래 pinned 행) · 2ed9d2681 측정과 "(operator brief)" 재측정 주장은 SUPERSEDED (브리프는 저장소에 없음)   GREEN (M4 착지)
  pinned (cb54103ea, this round): git grep -c -h -E '"domain".*best-practices axis' cb54103ea -- .claude/commands/harness/release-update/manifest.json; echo "exit=$?" → verbatim `1` then `exit=0`
  cmd:     grep -c '"domain".*best-practices axis' .claude/commands/harness/release-update/manifest.json
  stdout:  1   exit: 0

LED-007 R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned — MP-8 재고정)
  cmd:     git grep -c -h -F rust-v0.161.0 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-007 G   tree 4fe4ffe7b (M1 착지 커밋 — `git cat-file -t 4fe4ffe7b` = commit; 착지 재관측 `git grep -c -h -F rust-v0.161.0 4fe4ffe7b -- .claude/agents/harness/hns-release-update-specialist.md; echo "exit=$?"` → `5`, `exit=0`) · 고정 커밋 cb54103ea 재측정 (아래 pinned 행) · 2ed9d2681 측정과 "(operator brief)" 재측정 주장은 SUPERSEDED (브리프는 저장소에 없음)   GREEN (M1 착지)
  pinned (cb54103ea, this round): git grep -c -h -F rust-v0.161.0 cb54103ea -- .claude/agents/harness/hns-release-update-specialist.md; echo "exit=$?" → verbatim `5` then `exit=0`
  cmd:     grep -c rust-v0.161.0 .claude/agents/harness/hns-release-update-specialist.md
  stdout:  5   exit: 0

DIFF-01   git diff --name-only 2aab5f797b75983e132af451da68f69e3426557b d36e97571a4d39b3186a71781be9bf01f9434d7f
  stdout (9 paths, verbatim):
    .claude/agents/harness/hns-release-update-specialist.md
    .claude/commands/harness/release-update/manifest.json
    .claude/workflows/hns-release-update-run.js
    .moai/research/upstream-update-20261007.md
    .moai/research/upstream-update-20261008.md
    .moai/specs/SPEC-RELUP-DUALAXIS-001/acceptance.md
    .moai/specs/SPEC-RELUP-DUALAXIS-001/plan.md
    .moai/specs/SPEC-RELUP-DUALAXIS-001/progress.md
    .moai/specs/SPEC-RELUP-DUALAXIS-001/spec.md
```

Gaps (이 라운드가 관측하지 않은 것):
- **exit 코드 (라운드 3에서 관측)**: 이 라운드에서 다시 실행한 RED·GREEN·CTRL 명령은 `; echo "exit=$?"`를 붙여 실행했고, 출력된 exit 값을 셀에 옮겼다. 이 라운드가 다시 측정하지 않은 행(LED-011·013 등)은 이전 라운드 관측을 인용한다.
- **워크트리 가드 거부 (verification-claim-integrity §3.1 — 대체 측정은 조용히 바꾸지 않고 여기 기록)**: 블록 스코프 셀을 `python3 -c` 안에서 `subprocess`로 `git show <SHA>:<path>`를 호출하는 형태로 실행하자 워크트리 가드가 `this command names git in a form too complex to verify that it stays inside the worktree`로 거부했다(2026-10-10 측정). 그래서 (a) 핀 RED는 평문 `git grep -c -h`(앵커 부재 — 출력 없음 · exit 1), (b) GREEN 블록 셀은 git 호출 없이 작업 트리 파일을 읽는 python 한 호출로 나눴다. 핀 SHA에서 블록 경계를 직접 계산한 GREEN은 이 워크트리에서 얻을 수 없다. 대신 `.claude/` 트리가 `d36e97571`과 동일함을 `git diff --quiet d36e97571 0cccc0d31 -- .claude/`(exit 0)로 관측했다.
- **검증 동사 E3-P3·P4(LED-016·020)**: 핀 2aab5f797에서 node 실행으로 재확인하지 않았다. 전제는 LED-003(`selectCodexSweepTargets(args)` 출현 = 출력 없음, 핀)으로만 간접 확인했다.
- **LED-021T**: 해소 — M3 항목 5 착지 커밋 `cb54103ea`에서 이 라운드 재측정(고정 커밋 명령·출력은 LED-021T 블록). 이전 기록 「M3 항목 5 착지 커밋 없음」은 SUPERSEDED.
- **E3-P5·E3-P6 (이 라운드 추가 — plan §E)**: 두 검증 동사는 러너 `run()` 공개 경로를 모의 spawn으로 한 번 실행해 codex 렌즈 프롬프트 본문만 측정한다. 라이브 GitHub 호출은 없었다. E3-P6의 중단 조건 needle(`stop when the baseline tag rust-v0.161.0 appears`, `or when the list ends`)은 릴리즈 목록이 최신순(newest first)이라는 가정에 의존하며, 이 가정은 라이브로 관측되지 않았다.
- **LED-028 잔여 위험 선언**: LED-028 counts the six canonical URLs over the whole file. A mutant that removes them from the Phase 3 block and re-adds them in prose passes LED-028. The check that catches it is the verdict-pending Phase 3 block face (LED-022..027). A block-scoped blocking gate belongs to the structural-check follow-up card.

## §D.4 간접 검증 항목

- **러너 파스**: plan §E3 CommonJS require() 스모크 — `node --check`의 무음 통과 한계 보강.
- **manifest JSON 타당성**: plan §E4 python json.load — domain 문자열 편집 후 문법 훼손 잡기.
- **spec-lint**: plan §E6 — `### Out of Scope —` h3 (MissingExclusions)·frontmatter 12 필드 (FrontmatterInvalid) 0건.
- **시드 신선도**: plan §C — run-phase 착지 직전 npm/gh 재측정으로 `rust-v0.161.0` 유효성 재판정 (0.162 승격 대응).

## §D.5 종결 게이트 (Definition of Done)

1. 블로킹 AC 4종 전부 GREEN (001·002·007·009 — 009는 제거면 LED-021) (RED-now가 대응 마일스톤에서 뒤집힘 — exit code 포함 관측). AC-RDX-009의 구형 URL 제거면은 M3 항목 5로 이 카드가 수행한다(게이트: 착지 SHA의 LED-021 = 0 / exit 1). 현재 d36e97571에서는 LED-021 = 1이므로 이 항목은 미충족이다(B-02). 판정 보류 가드 10종(AC-RDX-003/004/005/006 — CX-7/8, AC-RDX-008·010·014·015 — 분할 efde, AC-RDX-017 — CX-12, AC-RDX-016 — CX-13)은 구조 면 착지 신호로 기록되고 판정은 형제 카드 계측으로 이관된다 — 가드 자체는 게이트 아님(verification-completeness §2 채택 기준 — 계측이 너무 얕아 채택 불가; §2.1 처분군 적용). plan §E7 검토면은 REQ-RDX-013·015의 구속 판정면이며 항목 6의 종결 게이트로 편입된다(D4 option (a)).
   - **기록 주석 — 리더 판정 494e (spec §1.2 D8, 0.11.0; 근거: 커밋 `cb54103ea` 메시지 원문 「Leader ruling 494e (scope: DoD item 1 unchanged; LED-028 only in AC-RDX-009, §D and M3 item 5)」).** (a) 위 항목 1의 문구는 HEAD 그대로 둔다 — 수정 없음. LED-028은 이 항목 밖이다: 항목 1은 AC-RDX-009를 제거면 LED-021로만 명명하며, LED-028의 블로킹 면은 AC-RDX-009·§D·plan M3 항목 5에 한정된다. 잔여 위험(블록 밖에서 다시 쓴 URL은 LED-028을 통과하고, 판정 보류 면 LED-022..027이 잡는다)은 선언만 하고 바꾸지 않는다. (b) 위 항목 안의 상태 문장 「현재 d36e97571에서는 LED-021 = 1이므로 이 항목은 미충족이다(B-02)」는 **SUPERSEDED**다 — d36e97571 시점의 기록이며 이력으로 보존한다. 고정 커밋 `cb54103ea`에서 `git grep -c -h "docs.anthropic.com" cb54103ea -- .claude/agents/harness/hns-release-update-specialist.md; echo "exit=$?"`의 출력은 `exit=1`(0힛, 이 라운드 관측)이므로 LED-021 면은 그 커밋에서 충족이다. 항목 1 전체(블로킹 AC 4종 GREEN)의 확인은 run-exit E2 재측정 소관이다.
2. 회귀 가드 2종 기준선 유지 (AC-RDX-011 = LED-011·LED-012 합병, AC-RDX-013 = LED-013·LED-015; 변화 없음).
3. spec.md REQ-RDX-001..015 전부 구현 대응물 존재 — REQ↔AC 추적성 §D.2 공백 없음. 블로킹 AC가 없는 REQ의 게이팅 처분은 §D.2 '게이팅 처분 (B-07)' 표가 SSOT다.
4. 변경 스코프 단언 (CX-15 재정식 + 클로저 CX-16/17 보강) — 병합 베이스 고정: `CARD_BASE=$(git merge-base origin/main HEAD)`(plan 시점 관측 `2aab5f797b75983e132af451da68f69e3426557b`; 감사가 핀한 base가 우선). **네 채널을 개별 관측**한다: (a) 커밋분 `git diff --name-only "$CARD_BASE"..HEAD`, (b) 스테이지 분 `git diff --cached --name-only`, (c) 비추적 분 `git status --porcelain`의 `??` 행, (d) **비스테이지 추적 편집 `git diff --name-only`(인자 없음 — CX-16의 제4 채널: unstaged tracked 편집은 (a)~(c) 어디에도 안 보인다)**. 네 채널 전부가 허용 경로만 반환: 하네스 3표면(`.claude/agents/harness/hns-release-update-specialist.md`·`.claude/workflows/hns-release-update-run.js`·`.claude/commands/harness/release-update/manifest.json`) + `.moai/specs/SPEC-RELUP-DUALAXIS-001/` + **카드 입력 연구 2건(`.moai/research/upstream-update-20261007.md`·`upstream-update-20261008.md` — CX-17)**. bare `git diff --name-only`의 공집합은 측정이 아니라 부재다.
5. 미해결 판단 표지 0개 — 열린 판단은 전부 spec.md §1.2 결정 기록으로 봉쇄. MP-7 리터럴 검사가 이 항목의 게이트다.
6. **E7 검토 게이트 — 증거 항목 (d) (D4 option (a))**: REQ-RDX-013·REQ-RDX-015의 구속 판정면은 plan §E7 검토면이다. 종결 시 run-exit E1 인간 검토 기록에 **E7(a)** 판정(REQ-RDX-015 — Phase 2 조기 종료가 CC 축 한정인지)이 있어야 하며, 이 기록이 이 게이트의 증거 항목 (d)다. E7(b)–(d)는 같은 검토 기록에 함께 남긴다. 기록이 없으면 두 REQ는 종결 시점 검증 없이 남으므로 게이트 미충족이다. 각 판정은 긍정 판단으로 증거와 함께 기록해야 한다 — 부정 판단, 보류 판단, 기록 없음은 전부 종결 차단이다(F5, 분할 efde).

## §D.6 선향 체크 (착지 후 다음 스윕이 검증할 것)

- 다음 `/harness:release-update` 실행이 codex 렌즈를 실제로 fan-out하는지 — 첫 실행 관측까지 렌즈 프롬프트의 커밋 복원 실효성은 미검증으로 남는다(spec.md §7).
- `last-codex-version.json` 첫 생성이 Phase 7a-codex 절차를 따르는지 — 기계 로컬이라 본문 문서화가 규범면.
- 0.162 alpha → stable 승격 시 6테마 관찰목록 승계가 절차대로 동작하는지.
