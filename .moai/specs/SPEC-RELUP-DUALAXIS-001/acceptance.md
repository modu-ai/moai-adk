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
- **귀속 분리 (B-01·B-04)**: RED 셀의 귀속 SHA는 `2aab5f797b75983e132af451da68f69e3426557b`다 (형식: `git grep -c -h "<패턴>" <sha> -- <경로>`). 착지 GREEN 셀은 측정 시점 HEAD(`d36e97571a4d39b3186a71781be9bf01f9434d7f`)에 귀속한다. 하네스 3표면은 2aab5f797 이후 변경됐으므로(§D.3-d `DIFF-01`) 작업 트리 값은 RED 근거가 되지 않는다. 0힛은 빈 stdout이다(exit 1은 문서화된 상태이며, 본 도구 표면은 exit를 표시하지 않는다 — §D.3-d Gaps).
- **RED의 올바른 이유**: 각 RED는 "이 SPEC이 바꿀 표면이 오늘 비어 있다"는 이유로 적색이다 — 구현이 그 표면을 채우면 뒤집힌다. 선존재 파일이 못 만지는 wrong-reason red는 없다.
- **명령 형태**: 전부 단일 호출(파이프·리다이렉트·`&&`·`;`·서브셸 없음). grep 일치 0개는 exit 1이다 — "빈 출력 + exit 1"은 완전한 관측이다.
- **비재현 관측 처분**: 재실행 불가능한 관측(예: internal/ 0힛 — 이미 녹색인 부재 클레임)은 회귀 가드로 분류하고 릴리스 블로킹에서 제외한다(undecidable disposition).
- **판정 보류 강등 (plan-audit iter3 CX-7/CX-8 + fresh-run iter2 CX-12 + 리더 재개 CX-13)**: AC-RDX-003/004/005(계측이 구조 면만 전달 — 빈 상수·"Return ok." 프롬프트 mutant 통과)와 AC-RDX-006(전-file 계수 — 산개 언급 mutant 판별 불가)은 verification-completeness §2 채택 기준(계측이 너무 얕아 채택 불가 — mutant-probe adoption bar; §2.1 처분군 적용)에 따라 회귀 가드(판정 보류)로 강등됐다. AC-RDX-017도 동일 처분(CX-12) — 리터럴 쌍(LED-018/019)이 동의어 바꿔쓰기 클래스에 우회됨이 실증돼 의미론 판정은 plan §E7 검토면으로 이관됐다. **AC-RDX-016도 동일 처분(CX-13)** — `source-first` 리터럴 면은 반전 가능(리터럴 유지·규칙 역전)이며 REQ-RDX-013의 의미론 면은 plan §E7 검토면 + 형제 카드 판별기로 이관된다. RED 셀은 전부 측정된 사실로 보존되며 판정은 §7의 형제 카드 계측으로 이관된다. 게이트가 아니다.

## §D AC Matrix

| AC | 분류 | 주장 (Then) | RED-now (LED) | Green path |
|----|------|-------------|---------------|------------|
| AC-RDX-001 | 블로킹 | manifest `domain` 필드가 `Codex CLI upstream change tracking`을 명명 (domain 키 행 스코프 — CX-3) | LED-001 (0/1) | M4 → ≥1/0 |
| AC-RDX-002 | 블로킹 | manifest `domain` 필드가 best-practices 축을 명명 (동일 스코프) | LED-002 (0/1) | M4 → ≥1/0 |
| AC-RDX-003 | 회귀 가드(판정 보류 — CX-7) | (구조 면) runner에 codex 렌즈 셀렉터가 정의·호출되고 병합·agent 호출까지 흐른다 — top-level과 `run()` 공개 진입점 모두 — 내용 면(프롬프트 실문) 미측정 → 형제 카드 | LED-003 (0/1) + LED-016 (exit 1) + LED-020 (exit 1, no-codex-in-run) | M2 착지 신호(비게이트): grep ≥2/0 AND P3 `dispatch-ok codex=1 total=2`/0 AND P4 `run-ok codex=1 total=1`/0 |
| AC-RDX-004 | 회귀 가드(판정 보류 — CX-7) | (구조 면) runner에 커밋 복원 폴백 앵커 `CODEX_COMMITS_FALLBACK` 존재 — 상수 본문 내용 미측정(빈 문자열 통과) → 형제 카드 | LED-004 (0/1) | M2 착지 신호(비게이트): ≥1/0 |
| AC-RDX-005 | 회귀 가드(판정 보류 — CX-7) | (구조 면) runner에 6테마 체크리스트 앵커 `CODEX_THEME_CHECKLIST` 존재 — 테마 행 실문 미측정 → 형제 카드 | LED-005 (0/1) | M2 착지 신호(비게이트): ≥1/0 |
| AC-RDX-006 | 회귀 가드(판정 보류 — CX-8) | (구조 면) specialist가 codex 상태 파일을 두 사이트에 걸쳐 문서화 — 전-file 계수라 산개 언급 mutant 판별 불가 → 형제 카드 | LED-006 (0/1) + LED-017 (0/1) | M1 착지 신호(비게이트): LED-006 ≥2/0 AND LED-017 ≥1/0 |
| AC-RDX-007 | 블로킹 | specialist가 시드 `rust-v0.161.0`을 기술 | LED-007 (0/1) | M1 → ≥1/0 |
| AC-RDX-008 | 블로킹 | specialist에 BP 상시 절차 섹션 존재 | LED-008 (0/1) | M3 → ≥1/0 |
| AC-RDX-009 | 블로킹 | specialist Phase 3 URL 세트가 6종 캐노니컬 전문(`code.claude.com/docs/en/{hooks, sub-agents, skills, plugins, mcp, settings}`)을 각각 보유하고 구형 `docs.anthropic.com` URL은 제거된다 (CX-14 제거면 + CX-18 육면 열거면) | LED-021 (핀 2aab5f797: `6` / exit `0` — 게이트, §D.3-d `LED-021R`) + LED-022..027 (핀 RED `0` → 착지 GREEN `1`, §D.3-d) | M3 → LED-021 = 0 / exit 1 (착지 SHA 귀속 — 미착지 시 RED); 육면은 착지 후 각 1 유지 |
| AC-RDX-010 | 블로킹 | specialist가 `HTML proposal report` 산출물을 명명 | LED-010 (0/1) | M3 → ≥1/0 |
| AC-RDX-014 | 블로킹 | specialist Phase 0 codex 블록이 codex 상태 파일 부재 시 기본값(`rust-v0.161.0` + 경고)과 키군 4종(REQ-RDX-001)을 문서화 | LED-006 공유 (핀 0 / exit 1) + 블록 스코프 핀 `git grep -n -m 1 -A 23` 출력 없음 (§D.3-c) | M1 → 블록 스코프 GREEN: 71–94행 블록에 기본값 + 경고 + 키군 4종 (B-07·B-08) |
| AC-RDX-015 | 블로킹 | runner CODEX_THEME_CHECKLIST 블록이 alpha watch 규범(watch 관찰목록 전용, 안정 탑재 시에만 채택)을 담는다 | LED-005 공유 (핀 0 / exit 1) + 블록 스코프 핀 `git grep -n -m 1 -B 6 -A 7` 출력 없음 (§D.3-c) | M2 → 블록 스코프 GREEN: 79–92행 블록의 watch 규범 문장(81–84행) (B-08) |
| AC-RDX-016 | 회귀 가드(판정 보류 — CX-13) | (구조 면) specialist BP 절차가 `source-first` 리터럴을 명명 — 리터럴 유지·규칙 역전 mutant에 우회됨(반전 클래스) → REQ-RDX-013 의미론 면은 plan §E7 검토면 + 형제 카드 판별기로 이관 | LED-014 (0/1) | M3 착지 신호(비게이트): ≥1/0 |
| AC-RDX-017 | 회귀 가드(판정 보류 — CX-12) | (구조 면) Phase 2 조기 종료의 축별 재범위화 — 리터럴 쌍(LED-018/019)은 동의어 바꿔쓰기(paraphrase) 클래스에 우회됨이 실증됨(CX-12) → 의미론 판정은 plan §E7 검토면으로 이관, 형제 카드가 판별기 흡수 | LED-018 (0/1) + LED-019 (1/0 — 구조 참고, 측정 사실 보존) | M1 착지 신호(비게이트): LED-018 ≥1/0 AND LED-019 =0/1 |
| AC-RDX-011 | 회귀 가드 | `internal/`에 `last-codex-version` 참조 0힛 유지 (Go 라이터 부재 보존) | — (오늘 녹색 — 부재 클레임, 비재현) | 유지 조건: run-phase 전체 |
| AC-RDX-012 | 회귀 가드 | specialist의 `last-cc-version.json` 문서화 ≥3힛 유지 (CC 축 절차 보존) | — (오늘 녹색 3힛) | 유지 조건: run-phase 전체 |
| AC-RDX-013 | 회귀 가드 | manifest의 `hns-release-update-run.js` 참조 1힛 유지 + `sprint_contract` dimensions·thresholds 판독 기준선 일치 (LED-013 + LED-015 — CX-3 판독면) | — (오늘 녹색: 1힛 + LED-015 기준선 출력) | 유지 조건: run-phase 전체 |

**집계 (B-06, 원장 재계수)**: AC 17 = 블로킹 8 (001·002·007·008·009·010·014·015) + 가드 9 = 판정 보류 6 (003·004·005·006·016·017) + 회귀 가드 3 (011·012·013). RED-now 앵커 = grep 15 (LED-001·002·003·004·005·006·007·008·009·010·014·017·018·019·021) + 검증 동사 2 (LED-016·020, plan §E3-P3·P4). AC-RDX-009 육면 셀(LED-022..027)은 블로킹 AC의 쌍 구성원이므로 위 RED-now 집계에 넣지 않고 §D.3-d에 따로 적는다.

## §D.1 시나리오 (Given-When-Then — 블로킹 8종 + 판정 보류 6종 + 회귀 가드 3종 = 17)

> AC-RDX-003/004/005/006의 시나리오는 구조 면 관측을 기술한다 — CX-7/CX-8 판정 보류로 게이트 밖이며 판정은 형제 카드 계측으로 이관된다(iter3). AC-RDX-017은 CX-12(paraphrase 우회), AC-RDX-016은 CX-13(리터럴 반전)으로 판정 보류 — 의미론 판정은 plan §E7 검토면 + 형제 카드로 이관된다.

- **AC-RDX-001** — **Given** manifest.json이 CC 단일 domain 문자열을 담은 상태로, **When** LED-001 명령(`domain` 키 행 스코프)을 실행하면, **Then** 일치 개수가 1 이상이다 (domain 필드가 codex 축을 명명 — source_request의 동일 문구는 매치 제외, CX-3).
- **AC-RDX-002** — **Given** 동일 상태로, **When** LED-002 명령(동일 스코프)을 실행하면, **Then** 일치 개수가 1 이상이다 (domain 필드가 best-practices 축을 명명).
- **AC-RDX-003** — **Given** runner가 CC 렌즈만 fan-out하는 상태로, **When** LED-003 명령과 LED-016(plan §E3-P3), LED-020(plan §E3-P4 `run()` 공개 경로) 명령을 실행하면, **Then** `selectCodexSweepTargets(args)` 출현이 2 이상(정의+top-level 병합 지점)이고, 모의 런타임이 관측한 `codex-release-notes:` 라벨 agent 호출이 top-level과 run() 양쪽에서 1 이상이다 (병합 제외는 LED-016, run() 경로 누락은 LED-020에서 좌초 — CX-5/CX-11; 주석·미연결 정의는 LED-003에서 좌초 — CX-2).
- **AC-RDX-004** — **Given** runner에 커밋 복원 절차가 없는 상태로, **When** LED-004 명령을 실행하면, **Then** `CODEX_COMMITS_FALLBACK` 앵커가 1 이상 관측된다.
- **AC-RDX-005** — **Given** runner에 테마 관찰목록이 없는 상태로, **When** LED-005 명령을 실행하면, **Then** `CODEX_THEME_CHECKLIST` 앵커가 1 이상 관측된다.
- **AC-RDX-006** — **Given** specialist 본문에 codex 상태 절차가 없는 상태로, **When** LED-006 명령과 LED-017 명령을 실행하면, **Then** `last-codex-version.json`이 2 이상(Phase 0 판독·기본값 사이트 + Phase 7a 기록 사이트)이고 `7a-codex` 기록 단계 리터럴이 1 이상이다 (Phase 0 단독·Phase 7a 단독 mutant 모두 좌초 — CX-6).
- **AC-RDX-007** — **Given** AC-RDX-006이 충족된 상태에서도 시드가 빠질 수 있으므로(mutant M-3), **When** LED-007 명령을 실행하면, **Then** `rust-v0.161.0`이 1 이상 관측된다.
- **AC-RDX-008** — **Given** specialist에 BP 축이 없는 상태로, **When** LED-008 명령을 실행하면, **Then** best-practice 섹션이 1 이상 관측된다.
- **AC-RDX-009** — **Given** Phase 3 URL 세트가 docs.anthropic.com 구형 나열인 상태로, **When** LED-021(제거면)과 LED-022..027(6종 전문 각각, 이스케이프 점 + 종결 경계 패턴 — B-03)을 실행하면, **Then** 6종 캐노니컬 URL이 각 1 이상이고 `docs.anthropic.com` 계수는 0 / exit 1이다 — URL 전부 삭제 mutant(양면 통과)는 육면 열거면이, 부분 교체 mutant는 LED-021이 봉쇄한다 (CX-14 + CX-18). 착지 전 d36e97571에서는 LED-021 = 1이므로 이 시나리오는 아직 RED다.
- **AC-RDX-010** — **Given** BP 산출물이 명명되지 않은 상태로, **When** LED-010 명령을 실행하면, **Then** `HTML proposal report`가 1 이상 관측된다.
- **AC-RDX-014** — **Given** codex 상태 파일이 존재하지 않는 다음 스윕 실행을 상정하는 상태로, **When** specialist의 Phase 0 codex 블록(§D.3-c, 71–94행)을 읽으면, **Then** 같은 블록에 (a) 부재 시 기본값 `rust-v0.161.0` + 경고 절차와 (b) 키군 4종(`last_analyzed_version` · `last_analyzed_date` · `last_master_research` · `analysis_history[]`)이 기술돼 있다 (REQ-RDX-004 + REQ-RDX-001 키군 절 — B-07·B-08). 스키마 문서화 단독 통과 mutant는 AC-RDX-006/007과 쌍으로 잡는다.
- **AC-RDX-015** — **Given** alpha 테마가 안정에 미탑재 상태로, **When** runner의 CODEX_THEME_CHECKLIST 블록(§D.3-c, 79–92행)을 읽으면, **Then** 같은 블록(81–84행)에 watch 관찰목록 규범("alpha 테마는 채택 아님 — 안정 탑재 시에만 채택 판정")이 기술돼 있다 (REQ-RDX-009 — 1차 스윕 watch 판정의 절차화; B-08).
- **AC-RDX-016** — **Given** BP 절차에 원문 선행 강제가 없는 상태로, **When** LED-014 명령을 실행하면, **Then** `source-first` 리터럴이 1 이상 관측된다 (REQ-RDX-013 — mutant M-4의 기계 판정면).
- **AC-RDX-017** — **Given** specialist Phase 2가 무조건 조기 종료 문장을 담은 상태(CC 빈 주간에 codex/BP가 실행 전 종료 — 현재 본문 상태)로, **When** LED-018 명령과 LED-019 명령(제거면)을 실행하면, **Then** `only the CC axis` 리터럴이 1 이상이고 구형 무조건 문장 전문(`If no entries: emit "No new versions since vX.Y.Z" and stop`)은 0이다 — 리터럴만 주석으로 넣고 문장을 생존시키는 mutant는 LED-019에서 좌초한다 (CX-9 + CX-10, REQ-RDX-015).
- **AC-RDX-011** — **Given** `internal/` 트리에 Go 라이터가 없는 상태로, **When** `grep -rn "last-codex-version" internal/`을 실행하면, **Then** 출력이 없다 (0힛 — REQ-RDX-005, mutant M-5). 회귀 가드: 오늘 녹색이라 RED-now 셀이 없다.
- **AC-RDX-012** — **Given** specialist가 CC 축 절차를 보존한 상태로, **When** `grep -c "last-cc-version.json"`을 실행하면, **Then** 3 이상이다 (회귀 가드 — 오늘 녹색 3힛).
- **AC-RDX-013** — **Given** manifest가 골격과 sprint_contract 기준선을 보존한 상태로, **When** `grep -c "hns-release-update-run.js"`와 LED-015 python 판독을 실행하면, **Then** 1힛이고 dimensions·thresholds 출력이 기준선 `['Functionality', 'Consistency'] {'Functionality': 0.85, 'Consistency': 0.8}`과 일치한다 (REQ-RDX-011 — 회귀 가드, 오늘 녹색).

## §D.2 추적성 (AC ↔ REQ)

| AC | REQ | mutant 봉쇄 |
|----|-----|-------------|
| AC-RDX-001/002 | REQ-RDX-010 | M-1 (source_request 기만) — 필드 스코프로 봉쇄 강화 (CX-3) |
| AC-RDX-003 | REQ-RDX-006 | M-2 + M-6 — ≥2 앵커 + §E3-P3 디스패치 관측 면 (CX-2/CX-5); 내용 면은 CX-7 이관(형제 카드) |
| AC-RDX-004 | REQ-RDX-007 | — (구조 면; 폴백 절차 실문은 CX-7 이관 — 형제 카드) |
| AC-RDX-005 | REQ-RDX-008 | — (구조 면; 테마 행 실문은 CX-7 이관 — 형제 카드) |
| AC-RDX-006 | REQ-RDX-001/002/003 | M-7 — 이중 사이트 앵커 + `7a-codex` 기록 단계 (CX-6); 산개 언급 판별은 CX-8 이관(형제 카드) |
| AC-RDX-007 | REQ-RDX-002 | M-3 (seed 누락) |
| AC-RDX-008 | REQ-RDX-012 | — (섹션 존재면) |
| AC-RDX-009 | REQ-RDX-014 | 부분 교체는 LED-021, URL 전부 삭제는 육면 열거면이 봉쇄 (CX-14/CX-18); 육면 패턴은 이스케이프 점 + 종결 경계 (B-03) |
| AC-RDX-010 | REQ-RDX-014 | — |
| AC-RDX-014 | REQ-RDX-001 (키군, B-07) · REQ-RDX-004 | M-3의 제3 쌍 (스키마 문서화 + 시드 + 부재 기본값 + 키군 4종 — 블록 스코프 게이트) |
| AC-RDX-015 | REQ-RDX-009 | alpha-채택 오표기 mutant 봉쇄 |
| AC-RDX-016 | REQ-RDX-013 | M-4 — 구조 면; 리터럴 반전 클래스는 CX-13 이관(형제 카드 판별기 + plan §E7 검토면) |
| AC-RDX-017 | REQ-RDX-015 | M-8 — 구조 참고; paraphrase 판별은 CX-12 이관(형제 카드 판별기 + plan §E7 검토면) |
| AC-RDX-011 | REQ-RDX-005 | M-5 (Go 침입) |
| AC-RDX-012 | REQ-RDX-003 (보존 축) | — |
| AC-RDX-013 | REQ-RDX-011 (골격 보존) | threshold 편집은 LED-015 판독면에 걸린다 (CX-3) |

### 게이팅 처분 (B-07) — 블로킹 AC 연결과 강등

블로킹 AC가 연결되지 않은 REQ 9종 중 REQ-RDX-001은 AC-RDX-014에 편입해 게이팅한다(신규 AC를 세우지 않아 블로킹 집계 8을 유지). 나머지 8종은 아래처럼 강등한다. 소유자 `unassigned, leader to issue`는 카드·SPEC id를 발행하지 않은 상태이며, 형제 카드 발행은 리더 소관이다.

| REQ | 게이트 상태 | 처분 | 소유자 |
|---|---|---|---|
| REQ-RDX-001 | 블로킹 AC-RDX-014에 편입 (키군 절 — 블록 스코프 Then) | 게이팅 | — |
| REQ-RDX-003 | 없음 — AC-RDX-006 판정 보류 (CX-8) | 강등: 판정 보류 가드 | unassigned, leader to issue |
| REQ-RDX-005 | 없음 — AC-RDX-011 회귀 가드 (오늘 녹색 부재 클레임, RED-now 셀 없음) | 강등: 회귀 가드 (undecidable disposition) | unassigned, leader to issue |
| REQ-RDX-006 | 없음 — AC-RDX-003 판정 보류 (CX-7); 동작 면 E3-P3·P4는 착지 신호 | 강등: 판정 보류 가드 | unassigned, leader to issue |
| REQ-RDX-007 | 없음 — AC-RDX-004 판정 보류 (CX-7) | 강등: 판정 보류 가드 | unassigned, leader to issue |
| REQ-RDX-008 | 없음 — AC-RDX-005 판정 보류 (CX-7) | 강등: 판정 보류 가드 | unassigned, leader to issue |
| REQ-RDX-011 | 없음 — AC-RDX-013 회귀 가드 (기준선 녹색, RED-now 셀 없음) | 강등: 회귀 가드 | unassigned, leader to issue |
| REQ-RDX-013 | 없음 — AC-RDX-016 판정 보류 (CX-13); 의미론은 plan §E7(d) | 강등: 판정 보류 가드 | unassigned, leader to issue |
| REQ-RDX-015 | 없음 — AC-RDX-017 판정 보류 (CX-12); 의미론은 plan §E7(a)–(c) | 강등: 판정 보류 가드 | unassigned, leader to issue |

나머지 REQ 6종(002·004·009·010·012·014)은 블로킹 AC를 가진다: 002 → AC-007, 004 → AC-014, 009 → AC-015, 010 → AC-001·002, 012 → AC-008, 014 → AC-009·010.

## §D.3 증거 원장 (Evidence Ledger — RED 귀속 `2aab5f797` · GREEN 귀속 측정 SHA; 2026-10-10 수리 셀은 §D.3-c · §D.3-d)

각 행: 명령은 축자 그대로 단일 실행됐고, stdout은 같은 실행에서 관측했고, exit code는 출력 유무와 문서화된 상태로 귀속했다(§A '귀속 분리'). 경로는 워크트리 루트 기준.

| LED | 명령 (단일 호출) | stdout (축자) | exit | 판정 |
|-----|------------------|---------------|------|------|
| LED-001 | `grep -c '"domain".*Codex CLI upstream change tracking' .claude/commands/harness/release-update/manifest.json` | `0` | `1` | RED (AC-001) — domain 필드 스코프 (CX-3 재앵커) |
| LED-002 | `grep -c '"domain".*best-practices axis' .claude/commands/harness/release-update/manifest.json` | `0` | `1` | RED (AC-002) — 동일 재앵커 (CX-3) |
| LED-003 | `grep -c "selectCodexSweepTargets(args)" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-003) — 디스패치 호출 앵커, 착지 후 ≥2 (CX-2 재앵커) |
| LED-004 | `grep -c "CODEX_COMMITS_FALLBACK" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-004) |
| LED-005 | `grep -c "CODEX_THEME_CHECKLIST" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-005) |
| LED-006 | `grep -c "last-codex-version.json" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-006) |
| LED-007 | `grep -c "rust-v0.161.0" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-007) |
| LED-008 | `grep -ci "best-practice" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-008) |
| LED-009 | `grep -c "code.claude.com" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED — 게이트 아님 (전체 파일 계수, 산문 포함 — CX-14). 핀 2aab5f797 재실행 = 출력 없음 (§D.3-d RED-summary) |
| LED-010 | `grep -c "HTML proposal report" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-010) |
| LED-014 | `grep -c "source-first" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-016) |
| LED-015 | `python3 -c "import json;d=json.load(open('.claude/commands/harness/release-update/manifest.json'));sc=d['sprint_contract'];print(sc['dimensions'],sc['thresholds'])"` | `['Functionality', 'Consistency'] {'Functionality': 0.85, 'Consistency': 0.8}` | `0` | 회귀 가드 기준선 — 출력 불변 유지가 PASS (AC-013, CX-3 판독면) |
| LED-016 | plan §E3-P3 verb 축자 (모의-런타임 실행 — `node -e '...'`, plan §E3-P3 블록 참조) | stderr `REJECTED: no-codex-dispatch:1` | `1` | RED (AC-003 병합 관측면) — 현재 러너 디스패치는 CC 호출 1건, codex 라벨 0건 (CX-5, M2에서 `dispatch-ok codex=1 total=2`/exit 0으로 뒤집음) |
| LED-017 | `grep -c "7a-codex" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-006 기록 사이트) — Phase 7a 기록 단계 부재 (CX-6) |
| LED-018 | `grep -c "only the CC axis" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-017) — Phase 2 무조건 조기 종료 생존, codex/BP가 CC 널 주간에 실행 전 종료 (CX-9) |
| LED-019 | `grep -c 'If no entries: emit "No new versions since vX.Y.Z" and stop' .claude/agents/harness/hns-release-update-specialist.md` | `1` | `0` | RED-제거면 (AC-017) — 구형 무조건 문장 생존; 착지 후 0/exit 1이 PASS — 주석 포함 생존 전부 적색 (CX-10) |
| LED-020 | plan §E3-P4 verb 축자 (`node -e '...'` — `run()` 공개 경로, codex 전용 입력) | stderr `REJECTED: no-codex-in-run:0` | `1` | RED (AC-003 run() 진입점) — codex 전용 입력에서 run()이 agent 호출 0건 (CX-11, M2에서 `run-ok codex=1 total=1`/exit 0으로 뒤집음) |
| LED-021 | 핀 RED — §D.3-d `LED-021R`: `git grep -c -h "docs.anthropic.com" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md` | `6` | `0` | RED-제거면 (AC-009 게이트, B-01) — 핀 2aab5f797 귀속. HEAD 관측(`d36e97571`) = `1` / exit `0` (산문 179행, 미착지) |
| LED-021T | 착지 목표 — §D.3-d `LED-021T`: `grep -c "docs.anthropic.com" .claude/agents/harness/hns-release-update-specialist.md` | 목표 `0` | 목표 `1` | 착지 목표 셀 (AC-009 게이트, plan §F M3 항목 5) — 라벨 SHA: M3 항목 5 착지 커밋 (미정, 착지 시 기입). 미측정 |
| LED-022 | 펜스 §D.3-d `LED-022R` → `LED-022G` (이스케이프 점 + 종결 경계) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED 0 · 착지 GREEN 1 (B-03 패턴, B-04 분리) |
| LED-023 | 펜스 §D.3-d `LED-023R` → `LED-023G` (이스케이프 점 + 종결 경계) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED 0 · 착지 GREEN 1 (B-03 패턴, B-04 분리) |
| LED-024 | 펜스 §D.3-d `LED-024R` → `LED-024G` (이스케이프 점 + 종결 경계) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED 0 · 착지 GREEN 1 (B-03 패턴, B-04 분리) |
| LED-025 | 펜스 §D.3-d `LED-025R` → `LED-025G` (이스케이프 점 + 종결 경계) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED 0 · 착지 GREEN 1 (B-03 패턴, B-04 분리) |
| LED-026 | 펜스 §D.3-d `LED-026R` → `LED-026G` (이스케이프 점 + 종결 경계) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED 0 · 착지 GREEN 1 (B-03 패턴, B-04 분리) |
| LED-027 | 펜스 §D.3-d `LED-027R` → `LED-027G` (이스케이프 점 + 종결 경계) | 핀 RED `(출력 없음)` → 착지 GREEN `1` | 핀 1 / 착지 0 | 육면 열거면 (AC-009) — 핀 RED 0 · 착지 GREEN 1 (B-03 패턴, B-04 분리) |
| LED-011 | `grep -rn "last-codex-version" internal/` | (출력 없음) | `1` | 회귀 가드 기준선 — 0힛 유지가 PASS (AC-011) |
| LED-012 | `grep -c "last-cc-version.json" .claude/agents/harness/hns-release-update-specialist.md` | `3` | `0` | 회귀 가드 기준선 — ≥3 유지가 PASS (AC-012) |
| LED-013 | `grep -c "hns-release-update-run.js" .claude/commands/harness/release-update/manifest.json` | `1` | `0` | 회귀 가드 기준선 — 1 유지가 PASS (AC-013) |

**LED-001/002/003 재앵커 근거**: plan-audit iter1(CX-2/CX-3)으로 위 세 행의 명령을 교체했다 — (B-04 정정) 이 자리에 있던 '2aab5f797 이후 바이트 동일' 주장은 거짓이었다. 하네스 3표면은 2aab5f797 이후 변경됐고(§D.3-d `DIFF-01`), RED 귀속은 2aab5f797 `git grep` 재실행만 인정한다(§D.3-d RED-summary). **LED-016/017 신설 근거**: plan-audit iter2(CX-5/CX-6) — 같은 하네스 표면에서 본 실행 측정. **LED-018 신설 근거**: plan-audit iter3(CX-9) — 동일 측정 조건. **LED-019/020 신설 근거**: fresh post-split run(CX-10/CX-11) — 동일 측정 조건. **LED-021 신설 근거**: 리더 재개 재심(rcpt-0e776d398b9dd8d642d5f3ac, CX-14) — 동일 측정 조건. **LED-022..027 신설 근거**: 클로저 run(rcpt-ce61f029a011f5fa5ff6bead, CX-18) — 핀 2aab5f797에서는 0(RED, 출력 없음)이고 착지 d36e97571에서 1(GREEN)이다. 분리 기록은 §D.3-d(B-04)이며 게이트는 LED-021(B-01)이다. 나머지 LED 행은 원본 그대로다. LED-015의 세미콜론은 인용된 python 프로그램 내부의 것 — 셸 구분자가 아니므로 단일 호출 규약을 유지한다. LED-016/020의 오류 메시지는 어댑터 자체의 핸들러가 내는 결정적 한 줄이다(전체 스택 대신).

**보조 관측 (동일 트리)**: `grep -c "Codex" manifest.json` → `0`/exit 1 · `grep -ci "codex" runner` → `0`/exit 1 · `grep -c "HTML" specialist.md` → `0`/exit 1 · `grep -rn "last-cc-version" internal/` → 출력 없음/exit 1 (Go 라이터 부재 — 상태 파일이 하네스 계층 소유임의 근거, spec.md §1.1 M4). `git rev-parse --short HEAD` → `2aab5f797`.

### §D.3-c 블록 경계 (AC-RDX-014 · AC-RDX-015 블록 스코프 게이트 — B-08)

- **AC-RDX-014 블록**: `.claude/agents/harness/hns-release-update-specialist.md`의 Phase 0 codex 단락. 시작 71행(첫 출현 — `**Codex axis state — separate file (REQ-RDX-001)**`), 끝 94행(`Seed semantics` 단락의 마지막 줄). 다음 헤딩 `### Phase 1`은 96행이다. 게이트 명령: `git grep -n -m 1 -A 23 "last-codex-version.json" <SHA> -- <경로>` — `-m 1`은 첫 출현(71행)만, `-A 23`은 94행까지다. Phase 7a의 두 번째 출현은 판정에서 제외한다.
- **AC-RDX-015 블록**: `.claude/workflows/hns-release-update-run.js`의 CODEX_THEME_CHECKLIST 블록. 시작 79행(주석 `// Standing 6-theme …`), 끝 92행(`];`), 상수 선언 85행. 게이트 명령: `git grep -n -m 1 -B 6 -A 7 "CODEX_THEME_CHECKLIST" <SHA> -- <경로>` — `-B 6`은 79행, `-A 7`은 92행까지다. 113행의 `join` 참조는 블록 밖이므로 판정에서 제외한다.
- **판독 규칙**: 블록 안 행만 판정한다. AC-RDX-014는 기본값 `rust-v0.161.0`, 경고(`emit warning`), 키군 4종이 71–94행 안에 있을 때 GREEN이다. AC-RDX-015는 watch 규범 문장(81–84행)이 79–92행 안에 있을 때 GREEN이다. 두 블록 모두 핀 2aab5f797에서는 앵커가 없어 출력이 없다(RED).
- **잔여 위험**: 블록 안에 문장이 있는지의 판정은 여전히 판독이다. 기계화되지 않은 이 판정은 plan §E1 검토와 §E7 의미 검토 면이 제2 판정으로 남는다.

### §D.3-d 펜스 원장 — 수리 셀 (B-01 · B-02 · B-03 · B-04 · B-08)

명령은 축자다. `(출력 없음)`은 git grep의 0힛이며, exit 1은 문서화된 git grep 상태다(Gaps 참조). 모든 행은 귀속 SHA를 명시한다. 표 셀의 명령은 식별용 축약이며, 실행 가능한 축자 명령은 이 펜스에 있다(verification-completeness §2.1 — 표 셀은 셸 메타문자를 훼손할 수 있다).

```text
RED-summary  git grep -c -h "<anchor>" 2aab5f797b75983e132af451da68f69e3426557b -- <path>   (re-run 2026-10-10)
  no output (0 hits): LED-001 LED-002 LED-003 LED-004 LED-005 LED-006 LED-007 LED-008(-i) LED-009 LED-010 LED-014 LED-017 LED-018
  LED-019 (legacy unconditional stop sentence): 1
  LED-021 (legacy domain): 6
  LED-022..027 (six canonical URLs, escaped + boundary): no output x6

LED-021R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-01)
  cmd:     git grep -c -h "docs.anthropic.com" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  6
  exit:    0

LED-021G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   observed at HEAD (not a pass)
  cmd:     git grep -c -h "docs.anthropic.com" HEAD -- .claude/agents/harness/hns-release-update-specialist.md   (HEAD = d36e97571a4d39b3186a71781be9bf01f9434d7f)
  stdout:  1   (line 179, prose)
  exit:    0

LED-021T   label SHA: M3 item-5 landing commit (TBD — not measured)
  cmd:     grep -c "docs.anthropic.com" .claude/agents/harness/hns-release-update-specialist.md
  target:  stdout 0, exit 1
  status:  NOT MET at d36e97571 (LED-021G = 1)

LED-022R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-022G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   GREEN (observed at HEAD)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)' d36e97571a4d39b3186a71781be9bf01f9434d7f -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  1   exit: 0

LED-023R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/sub-agents([^A-Za-z0-9_./-]|$)' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-023G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   GREEN (observed at HEAD)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/sub-agents([^A-Za-z0-9_./-]|$)' d36e97571a4d39b3186a71781be9bf01f9434d7f -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  1   exit: 0

LED-024R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/skills([^A-Za-z0-9_./-]|$)' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-024G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   GREEN (observed at HEAD)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/skills([^A-Za-z0-9_./-]|$)' d36e97571a4d39b3186a71781be9bf01f9434d7f -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  1   exit: 0

LED-025R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/plugins([^A-Za-z0-9_./-]|$)' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-025G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   GREEN (observed at HEAD)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/plugins([^A-Za-z0-9_./-]|$)' d36e97571a4d39b3186a71781be9bf01f9434d7f -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  1   exit: 0

LED-026R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/mcp([^A-Za-z0-9_./-]|$)' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-026G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   GREEN (observed at HEAD)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/mcp([^A-Za-z0-9_./-]|$)' d36e97571a4d39b3186a71781be9bf01f9434d7f -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  1   exit: 0

LED-027R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-04)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/settings([^A-Za-z0-9_./-]|$)' 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
LED-027G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   GREEN (observed at HEAD)
  cmd:     git grep -c -h -E 'https://code\.claude\.com/docs/en/settings([^A-Za-z0-9_./-]|$)' d36e97571a4d39b3186a71781be9bf01f9434d7f -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  1   exit: 0

CTRL-A  spoof reproduced (unescaped dots) — the defect
  cmd:     printf 'codeXclaudeXcom/docs/en/hooks\n' | grep -c "code.claude.com/docs/en/hooks"
  stdout:  1
CTRL-B  negative control (B-03, required): escaped pattern must print 0
  cmd:     printf 'codeXclaudeXcom/docs/en/hooks\n' | grep -c -E 'code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  0
CTRL-C  boundary control: prefix variant must not match
  cmd:     printf 'https://code.claude.com/docs/en/hooks-invalid\n' | grep -c -E 'https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  0
CTRL-D  positive control: real URL before a closing paren must match
  cmd:     printf 'https://code.claude.com/docs/en/hooks)\n' | grep -c -E 'https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)'
  stdout:  1

BLK-014R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-08)
  cmd:     git grep -n -m 1 -A 23 "last-codex-version.json" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:  (no output)   exit: 1
BLK-014G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   GREEN (block 71–94; verdict lines verbatim, prefix stripped)
  cmd:     git grep -n -m 1 -A 23 "last-codex-version.json" d36e97571a4d39b3186a71781be9bf01f9434d7f -- .claude/agents/harness/hns-release-update-specialist.md
  stdout:
    71: **Codex axis state — separate file (REQ-RDX-001)**: read `.moai/state/last-codex-version.json`
    73: family (`last_analyzed_version` / `last_analyzed_date` / `last_master_research` /
    74: `analysis_history[]`) — it is a distinct file, never merged into the CC file.
    75: - If file missing: default `since_codex = "rust-v0.161.0"`, emit warning (REQ-RDX-004).
    82:   "last_analyzed_version": "rust-v0.161.0",
  exit:    0

BLK-015R   tree 2aab5f797b75983e132af451da68f69e3426557b   RED (pinned, B-08)
  cmd:     git grep -n -m 1 -B 6 -A 7 "CODEX_THEME_CHECKLIST" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/workflows/hns-release-update-run.js
  stdout:  (no output)   exit: 1
BLK-015G   tree d36e97571a4d39b3186a71781be9bf01f9434d7f   GREEN (block 79–92; verdict lines verbatim, prefix stripped)
  cmd:     git grep -n -m 1 -B 6 -A 7 "CODEX_THEME_CHECKLIST" d36e97571a4d39b3186a71781be9bf01f9434d7f -- .claude/workflows/hns-release-update-run.js
  stdout:
    81: // observed PR numbers and the MoAI exposure surface. Alpha-window themes remain
    82: // WATCH-LIST observations — adoption is judged only when the theme lands in a
    83: // stable release; an alpha-window theme is never reported as adopted drift
    84: // (REQ-RDX-009).
    85: const CODEX_THEME_CHECKLIST = [
    86-91:   "thread", "rollout", "subagent", "compaction", "MCP", "other",
  exit:    0

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
- **exit 코드**: git grep 0힛의 exit 1과 1힛의 exit 0은 본 도구 표면이 표시하지 않는다(`(Bash completed with no output)`로만 보인다). 본 원장은 출력 유무와 문서화된 git grep 상태로 귀속했고, exit를 관측했다고 주장하지 않는다.
- **검증 동사 E3-P3·P4(LED-016·020)**: 핀 2aab5f797에서 node 실행으로 재확인하지 않았다. 전제는 LED-003(`selectCodexSweepTargets(args)` 출현 = 출력 없음, 핀)으로만 간접 확인했다.
- **LED-021T**: 미측정 — M3 항목 5 착지 커밋이 아직 없다.

## §D.4 간접 검증 항목

- **러너 파스**: plan §E3 CommonJS require() 스모크 — `node --check`의 무음 통과 한계 보강.
- **manifest JSON 타당성**: plan §E4 python json.load — domain 문자열 편집 후 문법 훼손 잡기.
- **spec-lint**: plan §E6 — `### Out of Scope —` h3 (MissingExclusions)·frontmatter 12 필드 (FrontmatterInvalid) 0건.
- **시드 신선도**: plan §C — run-phase 착지 직전 npm/gh 재측정으로 `rust-v0.161.0` 유효성 재판정 (0.162 승격 대응).

## §D.5 종결 게이트 (Definition of Done)

1. 블로킹 AC 8종 전부 GREEN (RED-now가 대응 마일스톤에서 뒤집힘 — exit code 포함 관측). AC-RDX-009의 구형 URL 제거면은 M3 항목 5로 이 카드가 수행한다(게이트: 착지 SHA의 LED-021 = 0 / exit 1). 현재 d36e97571에서는 LED-021 = 1이므로 이 항목은 미충족이다(B-02). 판정 보류 가드 6종(AC-RDX-003/004/005/006 — CX-7/8, AC-RDX-017 — CX-12, AC-RDX-016 — CX-13)은 구조 면 착지 신호로 기록되고 판정은 형제 카드 계측·plan §E7 검토면으로 이관된다 — 게이트 아님(verification-completeness §2 채택 기준 — 계측이 너무 얕아 채택 불가; §2.1 처분군 적용).
2. 회귀 가드 3종 기준선 유지 (LED-011/012/013 변화 없음).
3. spec.md REQ-RDX-001..015 전부 구현 대응물 존재 — REQ↔AC 추적성 §D.2 공백 없음. 블로킹 AC가 없는 REQ의 게이팅 처분은 §D.2 '게이팅 처분 (B-07)' 표가 SSOT다.
4. 변경 스코프 단언 (CX-15 재정식 + 클로저 CX-16/17 보강) — 병합 베이스 고정: `CARD_BASE=$(git merge-base origin/main HEAD)`(plan 시점 관측 `2aab5f797b75983e132af451da68f69e3426557b`; 감사가 핀한 base가 우선). **네 채널을 개별 관측**한다: (a) 커밋분 `git diff --name-only "$CARD_BASE"..HEAD`, (b) 스테이지 분 `git diff --cached --name-only`, (c) 비추적 분 `git status --porcelain`의 `??` 행, (d) **비스테이지 추적 편집 `git diff --name-only`(인자 없음 — CX-16의 제4 채널: unstaged tracked 편집은 (a)~(c) 어디에도 안 보인다)**. 네 채널 전부가 허용 경로만 반환: 하네스 3표면(`.claude/agents/harness/hns-release-update-specialist.md`·`.claude/workflows/hns-release-update-run.js`·`.claude/commands/harness/release-update/manifest.json`) + `.moai/specs/SPEC-RELUP-DUALAXIS-001/` + **카드 입력 연구 2건(`.moai/research/upstream-update-20261007.md`·`upstream-update-20261008.md` — CX-17)**. bare `git diff --name-only`의 공집합은 측정이 아니라 부재다.
5. `[NEEDS CLARIFICATION]` 마커 0개 — 열린 판단은 전부 spec.md §1.2 결정 기록으로 봉쇄.

## §D.6 선향 체크 (착지 후 다음 스윕이 검증할 것)

- 다음 `/harness:release-update` 실행이 codex 렌즈를 실제로 fan-out하는지 — 첫 실행 관측까지 렌즈 프롬프트의 커밋 복원 실효성은 미검증으로 남는다(spec.md §7).
- `last-codex-version.json` 첫 생성이 Phase 7a-codex 절차를 따르는지 — 기계 로컬이라 본문 문서화가 규범면.
- 0.162 alpha → stable 승격 시 6테마 관찰목록 승계가 절차대로 동작하는지.
