---
id: SPEC-AC-COUNT-DISCRIMINATOR-001
title: "AC 개수 자가검사 판별자 — 진행 기록"
version: "0.5.0"
status: in-progress
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".claude/agents/moai, .claude/rules/moai/development, internal/template/templates/.claude, internal/template/templates/.codex, internal/spec"
lifecycle: spec-anchored
tags: "b12, changelog, acceptance-criteria, self-test"
---

# 진행 기록 — SPEC-AC-COUNT-DISCRIMINATOR-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **L** — iter-2 에서 **M 에서 재분류**했다. 산정(`plan.md` §F 전수 열거, fixture 를 3파일로 셈): 행 1-5 → 5, 행 6 fixture 3종 → 8, 행 7 스냅샷 → 9, 행 8-10 정규화 3건 → 12, 행 11 debt-list → 13, 행 12 catalog.yaml → 14, 행 13 progress.md → 15, **행 14 `.codex` TOML → 16**. **16 > 15 이므로 Tier L**(`spec-workflow.md:141-142`). 여기에 Tier L 산출물 2종(행 15·16)이 더해져 최종 16행 / 18파일 — 어느 쪽으로 세도 L 이다
  - **넘긴 한 건**: `internal/template/templates/.codex/agents/moai/manager-docs.toml`. B12 절의 **세 번째 배포 반송자**이며 계수기 명령까지 통째로 담는다. 미러 `.md` 에서 `make agents-emit` 이 기계 생성하고, golden 게이트가 `make build` 의 선행 타깃이라 재생성을 빠뜨리면 빌드가 정지한다(감사 N1)
  - **경계는 이미 이름 붙어 있었다**: 0.3.0 판의 `plan.md` §F 와 이 절이 둘 다 "상한 15 에 닿아 있다 — 여기서 파일이 더 붙으면 L 이다" 라고 적었다. 감사가 찾은 누락 1건이 정확히 그 한 건이었다
  - **13-15 범위 표기도 버렸다**(감사 N8): §F 자신의 계수 규칙으로는 단일 값 15 였고 13 은 행 수였다. 범위로 적으면 없는 여유가 있는 것처럼 읽히고, 그 오독이 이번 누락을 결정적으로 만들었다
  - LOC 는 여전히 300 미만이지만 **파일 축이 상위 축**이므로 판정은 파일 수가 낸다 — 0.3.0 판이 M 을 지킬 때 쓴 것과 같은 규칙을 반대 방향으로 적용한 것이다
- Tier L 이 바꾸는 것: 산출물 **5종**(+`design.md` +`research.md`) · plan-auditor 임계 **0.85** · plan-audit 반복 상한 **3**(`harness.yaml:78`) · git 경로 **Route B**(단, 이 리포는 `.claude/rules/local/repo-local-pr-policy.md` 의 git-flow 규정이 우선 — 카드는 `develop` 로 통합하고 카드 PR 을 내지 않는다) · 요구/수락 상한 각 **25**
- 산출물: `spec.md` + `plan.md` + `acceptance.md` + **`design.md`** + **`research.md`** (+ `progress.md`) — Tier L 5종 충족
- 요구 **8건** / 수락 **8건** — Tier L 상한(각 25) 이내. iter-2 수리에서도 **추가 없음**: N1-N11 을 모두 기존 REQ/AC 의 문면 확장으로 처리했다(REQ-ACD-002·006·007 확장, AC-ACD-001·002·004·005·006·007·008 확장)
- SPEC ID 정규식 자가검사: `PASS` (`ID="SPEC-AC-COUNT-DISCRIMINATOR-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`)
- ID 충돌 검사: `ls -d .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001` → rc=1 (미존재)
- 판별자 결정: 문면 규약(예약 토큰 `[RETIRED]` · `[REF]`, **식별자 등장에 인접**) + 세 상태(live·excluded·ambiguous) 계수기. 파서 기각 근거는 `spec.md` §3
- 실측 대비 정정 1건: 검출기 플래그 29개 중 8개가 오탐(`spec.md` §8) — 진짜 과다 계상은 플래그 집합 한정 14 파일 / 21 식별자(iter-1 에서 오탐 1건 추가 확인 → 오탐 4파일)
- 실측 대비 추가 관측 2건: 과다 계상의 **두 번째 축**(남의 SPEC 식별자 인용) — 이 SPEC 자신의 `acceptance.md` 가 스윕 18 / 실제 8. 상한 스캔 156/602 (`.moai/reports/t338/multidomain-scan.sh`). **세 번째 축**(같은 파일 안 별칭)은 iter-1 감사가 세우고 이 판이 전수로 넓혔다 — 상한 85파일 (`.moai/reports/t338/alias-shape-scan.sh`)
- 미결 결정: **없음** (미결 표지 잔여 0건) — 인용 축 소급 범위는 운영자가 **선택 A**(규약만 확장, 소급 없음)로 종결했다(`plan.md` §E M0, `spec.md` §2.2)
- M0 결정의 근거 실측(운영자, 이 트리, 이 카드 변경 0 상태): `.moai/reports/t338/pre-terminal-scan.sh` → `.moai/reports/t338/pre-terminal-scan.txt` — multi-domain 156 / `completed` **122** / pre-terminal **34** / status 없음 0
- 받아들인 비용(운영자 명시): 인용 축 종단 이전 **34건**이 각자 다음 sync 에서 규약을 처음 만난다. 누락이 아니라 채택된 비용이다
- 상시 부채·후보: **축별로 기록하고 합계를 만들지 않는다**(`.moai/reports/t338/debt-list.md`) — 폐기 축 `completed` 12건은 **하한**(검출기가 놓친다), 인용 축 `completed` 122건과 별칭 축 85파일은 **상한**(형태만 본다). 방향이 반대인 수를 더한 134 는 어느 쪽도 뜻하지 않으므로 삭제했다(감사 D4). 목록 기록만, 편집 없음(REQ-ACD-008 · AC-ACD-008)
- 정직하게 남긴 잔여 2건: ① 세 번째 상태는 **부분 표시**를 잡으므로 토큰이 하나도 없는 순수 인용은 정지가 아니라 과다 계상으로 남는다(접두사 단위 판정과 네이티브 접두사 선언 기제는 범위 밖 — `spec.md` §6). ② baseline 스냅샷은 재생성으로 회귀를 흡수할 수 있다 — 상태별 집계와 재생성 사유 기재는 그것을 **없애는** 장치가 아니라 **보이게 하는** 장치다(감사 D7, REQ-ACD-006 주석)

### v0.5.0 조항 개정 — 운영자 판정(재계획 아님)

- **무엇을 좁혔나**: AC-ACD-006 / REQ-ACD-006 의 **실패 표면**. 스냅샷에 **없는** 파일은 **보고하되 실패시키지 않는다**. 그대로 실패하는 것 둘: ① 스냅샷에 **있는** 파일의 live 개수·상태별 집계·정지 상태 변화(양방향), ② 스냅샷에 있는데 글롭이 더는 잡지 않는 **사라짐**
- **왜(운영자 판정)**: run-phase 8건 전부 PASS 뒤 `origin/develop` 흡수에서 깨끗한 머지가 수트를 붉게 만들었다 — `ac_count_clause_test.go:430`, `SPEC-TODO-DESTRUCTIVE-GUARD-001/acceptance.md: counts 16 but is absent from the snapshot`. **신규 `acceptance.md` 추가 속도 실측: 7일 59건 / 3일 28건 / 최근 24시간 5건.** 설계대로면 무관한 카드의 평범한 SPEC 저작이 하루 대여섯 번 이 테스트를 붉히고 매번 손 재생성을 요구한다 — 큐에 이미 열려 있는 CI-상시-레드 카드 2건과 같은 부류다. 이 SPEC 이 겨눈 병은 **이미 잰 파일의 조용한 과다 계상 회귀**이고 그 탐지는 손대지 않았다
- **이음매 해소(운영자 결정을 문면화)**: 종전 AC-ACD-006 항목 5(d)("스냅샷에 없던 정지는 언제나 실패")와 새 2번("부재 파일은 실패시키지 않는다")은 **새 파일이 정지할 때** 만난다. **결정: 부재를 먼저 판정한다** — 스냅샷이 담지 않은 파일은 세든 정지하든 보고·비실패. (d) 는 실제로 겨눈 회귀(**`COUNT` 로 기록된 파일이 정지하기 시작함**)로 좁혔다. 독자가 이 화해를 찾을 자리 셋: `acceptance.md` AC-ACD-006 항목 5(d) 아래 인용 블록 · `spec.md` §3.5 규칙 4 · `spec.md` HISTORY 0.5.0 행
- **보고 의무를 요구 출력으로 못 박았다**: 부재 파일마다 **경로 + 이번 실행의 관측**(`COUNT <n>` 또는 `HALT <식별자…>`)을 출력해야 한다. 보고를 내지 않는 좁힘은 그 파일에 대해 검사를 **꺼 버린 것과 구별할 수 없고**, 그 구별이 이 완화의 전부다
- **함께 정합화한 인용 4곳**: `spec.md` §3.5 규칙 1(+ "판정 규칙은 넷") · `spec.md` §4 REQ-ACD-006 글롭 주석 · `research.md` §B(글롭을 얼린 이유) · `design.md` §C.2 판정표. 셋 다 종전 `"어떤 차이든 실패"` 를 근거로 인용하고 있었다 — 글롭을 얼린 논거 자체는 유지된다(글롭 변경은 **이미 잰** 파일을 사라짐으로 만들 수 있고, 그것은 여전히 실패다)
- **손대지 않은 것**: 다른 REQ·AC 전부(요구 8 / 수락 8 불변, Tier L 불변) · `internal/spec/**` 구현 · manager-docs B12 절 · baseline 스냅샷(재생성 없음) · `plan.md` 본문(§F·M5 는 "상태 변화를 판정한다"만 말하므로 개정 불필요, 버전 필드만 정렬)
- **plan-audit 재실행 없음(정직하게 적는다)**: 기록된 **PASS 0.91 (iter-3)** 은 이 개정 **이전** 판(v0.4.1)에 대한 판정이며, v0.5.0 문면은 **어떤 감사도 받지 않았다.** 재감사는 불필요로 판정됐다(운영자). 따라서 §E.3 의 audit-ready 신호를 이 개정이 갱신하지 않는다
- **자가검사 재측정(6산출물, 이 개정 후)** — 이 SPEC 자신의 문서도 자기가 정의한 규약 아래 있으므로, 개정이 정지를 만들지 않았음을 실측으로 보인다. 명령: `python3 .moai/reports/t338/iter2-scratch/counter.py <파일> adj`. 여섯 파일 모두 `rc=0 · ambiguous=0` 이고 개정 전 값에서 하나도 움직이지 않았다 — `spec.md 41` · `plan.md 11` · `acceptance.md 24` · `design.md 2` · `research.md 10` · `progress.md 47`. (개정 전 같은 명령의 값도 동일했다. 이 절이 값을 적는 것은 스냅샷 갱신이 아니라 **정지 부재**의 증거이며, 판정에 쓰이는 것은 값이 아니라 `ambiguous=0` 이다 — REQ-ACD-008 의 재유도 규율이 여기에도 적용된다.)

### iter-1 감사(FAIL 0.67) 수리 기록 — v0.3.0

- **D1(critical) 해소**: 예약 토큰을 **행이 아니라 식별자 등장**에 묶었다(인접 규칙). 근거는 이 트리 재현 — `.moai/reports/t338/collision-scan.sh` → 한 행에 접두사 2개 이상인 행 **123행 / 56파일**, 결정적 사례 `SPEC-AGENT-PARALLEL-OPT-001/acceptance.md:248`(스윕 54, 행 단위 정규화 시 52 rc=0, 의도는 53). REQ-ACD-002·§3.1·§3.2 재작성; 세 상태가 새 단위에서 **망라적·배타적**임을 세웠다. AC-ACD-001 이 바로 그 입력을 쓰고 **53 과 52 를 구별해** 판정한다(원본은 `completed` 이므로 트리 밖 사본에 적용)
- **D2(critical) 결정**: 세 번째 축 = 같은 파일 안 **별칭**(`spec.md` §2.3). 실측 `.moai/reports/t338/alias-shape-scan.sh` → 형태 상한 **85파일**; 참 별칭(`SPEC-HUMANIZE-001` `AC-001` ~ `AC-HUM-001`)과 오탐(`SPEC-STATUSLINE-001` `AC-SL-001` ~ `AC-SL-NF-001`)을 각 1건 손으로 확인 → **형태 판정은 정지 규칙으로 쓸 수 없다**. **채택: 토큰을 늘리지 않고 `[REF]` 의 뜻을 넓힌다** — "다른 곳에 선언된 기준을 가리킨다"(남의 SPEC 이든 같은 파일 정본 철자든). 계수 규칙 무변경, 요구/수락 추가 없음, **Tier 영향 없음**. 소급 범위는 인용 축과 동일(선택 A — 85 는 미검증 후보 목록)
- **D3 해소**: `SPEC-V3R2-ORC-001` 의 유일 플래그 `AC-ORC-001-05` 가 오탐임을 손으로 확인(온전한 Given/When/Then; "retired" 는 검사 대상). 나머지 둘은 참으로 확인(CDS `:13-15` 표 셀 `RETIRED`, UDD 3건). → 오탐 **4파일**, 진짜 **14파일 / 21식별자**. AC-ACD-007 은 목록 인용을 버리고 **대상 집합 재유도 + 플래그별 손 판정 기록**으로 바꿨다. ORC-001 은 폐기 축에서 빠지고 **별칭 축 대상으로 남는다**(34 → 17)
- **D4-D8 해소**: 합계 134 삭제(축별·방향 명시) / 글롭 고정(깊이 1, `_archive/` 제외) + 얼린 수 삭제(같은 트리에서 601·602·603 이 동시에 참) / 센티널 앵커 + "정확히 1건·비어 있지 않음" 단언 / 스냅샷 상태별 집계 + 재생성 규율 / `manager-docs.md` 쌍 `diff` 무출력 단언(산문 절반 커버)
- **D9-D14 해소**: M0 비용 근거를 L82 의 진실에 맞춰 정정(멈춤 성질은 34건 기본형에 적용되지 않는다 — 저작 맥락·값어치 시점·잔여의 국지성 셋으로 교체) / 부채 목록·스냅샷 경로 명명 / AC-ACD-002 §3→§2 인용 정정 / REQ-ACD-007 `Where`→`When` / §F 에 `progress.md` 추가 / §7 중립성을 개정 절 자체까지 확장
- **감사 판정 정정 1건**: 감사 D13 은 `internal/template/catalog.yaml` 이 "올바르게 제외됐다"고 했으나 **틀렸다**. 엔트리 존재는 근거가 못 된다 — 엔트리가 담은 것은 **내용 해시**다. 실측: `shasum -a 256 internal/template/templates/.claude/agents/moai/manager-docs.md` → `27d6252a…3b7954` 가 `catalog.yaml:125` `hash:` 와 일치하고, `make build`(`Makefile:24`)가 `gen-catalog-hashes.go --all` 로 재생성한다. 미러를 고치면 추적 파일이 함께 바뀌므로 §F 12행에 넣었다. (`manager-develop-prompt-template.md` 는 catalog 엔트리가 없어 해당 없음 — grep 0건)
### iter-2 감사(FAIL 0.72) 수리 기록 — v0.4.0

- **N1(critical) 해소 + Tier 재분류**: 배포 반송자가 셋임을 확정했다(실측 `git grep -l -F "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+'"` → 로컬 절 · 템플릿 미러 · `.codex` TOML; CHANGELOG 는 역사 기록이라 정의 지점이 아니다). 생성 원본은 **템플릿 미러**다(`agentemit/golden_test.go:30-34` — `templatesDir="../templates"`, `agentMDRoot=".claude/agents/moai"`), 로컬 `.claude/` 가 아니다. `.codex` TOML 은 `catalog.yaml` 에 엔트리가 없어(`grep -c '\.codex'` → `0`) 해시 경로와 **별개 파일**이다. → §F 14행 추가 · M4 에 `make agents-emit` 을 `make build` **앞에** 명시 · `spec.md` §7 을 "세 반송자" 로 정정 · REQ-ACD-007 을 "모든 배포 반송자(손 미러 + 기계 생성물)" 로 확장 · AC-ACD-005 항목 6(생성 표면 + golden 게이트 + 뮤턴트) 추가 · **Tier M → L**
- **N2(major) 해소**: 행 끝 뮤턴트 기대값 `52` → **`54`**. 실측으로 뒤집힘을 확인했다 — 인접 구현은 `COUNT 54`, 행 단위 구현만 `52`. 종전 문면은 항목 2(53)·항목 3(52)·뮤턴트 절(52)이 **동시 만족 불가능**했다. `acceptance.md` AC-ACD-001 과 `plan.md` M3 양쪽을 고치고, 세 값(54/53/52)이 각각 무엇을 뜻하는지 판독표로 분리했다
- **N3(major) 해소 — 양쪽 절반**: ① *corpus 가 정지 파일을 어떻게 다루는가* → `spec.md` §3.5 + REQ-ACD-006 + AC-ACD-006 항목 5. 정지를 스냅샷의 **일급 `HALT` 항목**으로 기록하고 **상태 변화**를 판정한다(양방향 실패 + 미기록 정지는 언제나 실패). **건너뛰기·0 계상 금지** — 그것이 정지를 통과로 읽는 것이고 이 SPEC 이 없애려는 병이다. 정당성: B12(그 SPEC 이 옳은 수를 내는가)와 corpus 검증자(계수기가 안정적으로 분류하는가)는 **다른 것을 판정**하므로, 후자에서 정지는 정당한 관측이다. ② *예시 문서가 규약에 걸리지 않는 법* → `spec.md` §3.4. **예외가 아니라 적용**: 식별자 단위 균일성(모든 등장 표시 또는 한 등장도 표시 안 함)을 지키고, 한 식별자가 두 모양을 동시에 보여야 하면 **서로 다른 합성 ID** 를 쓴다. 싼 우회로(코드 스팬 제외)는 마크업 앵커라 REQ-ACD-004 가 막는다. 이 SPEC 자신의 표에 적용해 정지를 없앴다 — 실측 `AMBIGUOUS AC-SYN-010 AC-SYN-012` → `COUNT 24 … ambiguous=0`
- **N4·N5(major/minor) 해소**: 인접의 두 경계를 M1 문면에 못 박았다 — **같은 행**(줄바꿈은 인접을 깬다, 실측 live 3 vs 2)이고 **공백·탭이 아닌 문자는 무엇이든 인접을 깬다**(닫는 백틱 포함, 실측 53 vs 54). `spec.md` §3.1 + REQ-ACD-002 + `plan.md` M1 + AC-ACD-001 Given(배치를 문면 그대로 고정) + AC-ACD-004 fixture 2경우 추가(줄바꿈 개입 / 백틱 개입)
- **N6(minor) 해소**: AC-ACD-007 항목 5 비공허 가드 — 대상 1건 이상이거나, 0건이면 후보 전체와 제외 사유를 §E.2 에 기록해 **"0건이 측정 결과이지 미실행이 아님"** 을 보인다. AC-ACD-005 항목 1 이 이미 같은 가드를 걸고 있었으므로 일관성의 결손이었다
- **N7(minor) 해소**: 얼린 `123/56` 을 재유도 + 방향 라벨로 교체. 같은 HEAD 에서 **125** 로 이동해 있었다 — corpus 에 이 SPEC 자신의 파일이 들어 있어 **자기가 인용하는 수를 자기가 움직인** 것이다. 방향: 상한도 하한도 아니다(정당한 두 계열 인용을 과다 계상 / 같은 접두사 충돌을 과소 계상). D1 의 근거는 이 수가 아니라 결정적 사례 1건의 손 검증이다
- **N8-N11 정리**: N8 범위 표기 `13-15` → 단일 값 + Tier 산술 표 · N9 `testdata/account/` → `testdata/ac_count/`(실측: `account` 디렉터리 부재, 검증자 파일명과 정합) · N10 AC-ACD-008 항목 4 에서 폐기 축 절을 기계 대조에서 **명시적 면제**하고 손 판정 기록을 출처로 지정(그 절엔 대응하는 기계 값이 없어 판정 불가능한 항목이었다) · N11 AC-ACD-002 의 배제 보장 범위를 실측대로 좁힘(`pre-terminal-scan.txt` 는 34건만 열거, `completed` 122건 미포함 — 다만 미표시 파일은 두 값이 같으므로 AC 가 오탐 실패하지는 않는다)
- **감사 지적 중 반박한 것: 없음.** N1-N11 전부를 독립적으로 재측정해 확인했고, 모두 실재하는 결함이었다
- kickoff 게이트: **진입 가능** — plan-phase 산출물 Tier L 5종 완결, 미결 결정 0건, iter-1 D1-D14 + iter-2 N1-N11 반영 완료. 게이트 자체는 오케스트레이터가 운영자에게 묻는다. **재감사 시 임계는 0.85**(Tier L), 반복 상한은 3 이므로 iter-3 이 상한 없이 가능하다

## §E.2 Run-phase Evidence

측정 트리: 워크트리 `.claude/worktrees/t338`, 브랜치 `WT-ac-count-sweep`, 시작 HEAD `0a663b542`.
아래 수치는 모두 이 실행에서 명령을 돌려 관측한 값이다. 얼린 값을 옮겨 적은 항목은 없다.

### E.2.1 마일스톤 진행

M0 은 운영자 결정으로 이미 종결(선택 A)이므로 M1 부터 실행했다.

| 마일스톤 | 결과 | 관측 |
|---|---|---|
| M1 규약 문면 | 완료 | `.claude/agents/moai/manager-docs.md` B12 절 개정(81-164행). 인접의 두 경계를 문면에 못 박음 |
| M2 계약 + 앵커 | 완료 | 센티널 `# MOAI-AC-COUNTER-BEGIN` / `# MOAI-AC-COUNTER-END`(119·147행). stdout=정수 1개 / `AMBIGUOUS`+해소 한 줄+rc≠0 / 상태별 집계는 stderr |
| M3 검증자 | 완료 | `internal/spec/ac_count_clause_test.go` + fixture 3종 + baseline 스냅샷. 뮤턴트 3종 사망 관측 |
| M4 반송자 3종 + 빌드 | 완료 | 미러 → `make agents-emit` → `make build`. 6파일이 함께 이동 |
| M5 정규화 + 부채 목록 | 완료 | 재유도 결과 대상 **4건**(계획 기대 3건 — 아래 E.2.4) |

### E.2.2 RED 증거 (TDD)

검증자를 먼저 쓰고 절을 고치기 전에 돌려 RED 를 세웠다. 관측 출력(축약 없이 첫 행):

```
--- FAIL: TestACCounterExtractedFromBothCarriers (0.00s)
    ac_count_clause_test.go:132: .../manager-docs.md: expected exactly one "# MOAI-AC-COUNTER-BEGIN" / "# MOAI-AC-COUNTER-END" sentinel pair, got begin=0 end=0
```

같은 이유로 `TestACCounterFixtureCorpus` · `TestACCounterHaltsOnPartialMarking` ·
`TestACCounterFullCorpusMatchesBaseline` · `TestACCounterCorpusMutantIsDetected` ·
`TestACCounterReachesCodexCarrier` 여섯 건이 RED. 절 개정 후 GREEN.

M4 순서 규약도 실측으로 확인했다 — 미러만 고치고 `make agents-emit` 없이 golden 게이트를 돌리면:

```
golden_test.go:109: .codex/agents/moai/manager-docs.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
```

### E.2.3 뮤턴트 3종 — 모두 사망, 모두 원복

| 뮤턴트 | 심은 것 | 관측된 사망 | 원복 |
|---|---|---|---|
| (a) 미러 한 글자 | 미러의 `exc = 0` → `exc = 1` | ① `... are not byte-identical` ② golden sha256 불일치 | 두 게이트 모두 `ok` |
| (b1) fixture 부분 표시 | `shapes.md` 의 `AC-SYN-002` — 표시된 두 등장 중 한 곳에서 예약 토큰을 지움 | `expected exit 0, got 3 (stdout="AMBIGUOUS AC-SYN-002\n resolve: …")` | `ok` |
| (b2) 인접 토큰을 행 끝으로 | `adjacency.md` 의 `AC-SYN-010` — 인접해 있던 예약 토큰을 행 끝으로 옮김 | `live count = 5, want 4` / 집계 `live=5 excluded=1` | `ok` |
| (c) 임의 corpus 파일에 토큰 심기 | `SPEC-ADVISOR-RUNG-001/acceptance.md` 첫 식별자 뒤 `[RETIRED]` | `.moai/specs/SPEC-ADVISOR-RUNG-001/acceptance.md: snapshot records COUNT 7, this run HALTs (AC-ADV-001)` — **경로를 지목한다** | `ok`, `git diff --stat` 무출력 |

### E.2.4 AC 판정 표

| AC | 판정 | 명령 / 관측 |
|---|---|---|
| AC-ACD-001 | PASS | 원본 스윕 `54`; 트리 밖 사본(코드 스팬 **안**에 `[REF]`) 계수기 `53` rc=0; 행 끝 뮤턴트 `54`(사망); 원복 `53`; 원본 `git diff --stat` 무출력. 항목 3 의 `53`/`52` 구별은 같은 사본에 행 단위 구현을 돌려 `COUNT 52` 를 직접 관측해 세웠다 |
| AC-ACD-002 | PASS | 대조군 **`.moai/specs/SPEC-ADVISOR-RUNG-001/acceptance.md`** (세 스캔 어디에도 없음, 후보 491건 중 첫 건). 스윕 `7` = 계수기 `7`, 둘 다 0 아님, rc=0 |
| AC-ACD-003 | PASS | `TestACCounterHaltsOnPartialMarking`. 뮤턴트 이전 rc=0 live=3; 이후 `AMBIGUOUS AC-SYN-002` + 해소 한 줄(`[RETIRED]`/`[REF]` 문자열 포함) rc=3, 맨 정수 없음; 원복 rc=0 live=3 |
| AC-ACD-004 | PASS | (가) 네 오탐 파일 모두 스윕=계수기(18=18, 12=12, 20=20, 34=34, 정규화 **이전** 값) 이고 여덟 식별자(AC-RA-02/07/11/14/17, AC-CMR-002, AC-LSPMCP-RETIRE-007, AC-ORC-001-05) 상태 전부 `live`. (나) `adjacency.md` 여섯 경우 live=4 excluded=2 — 손계산 일치 |
| AC-ACD-005 | PASS | 항목1 추출 1건·비어있지 않음·바이트 동일 / 항목2 뮤턴트 (a) / 항목3 `manager-docs.md` 쌍 `diff` 무출력 / 항목4 prompt-template 쌍 차이 **171행 1건**(줄 수를 바꾸지 않는 1행 치환으로 편집해 행 번호 보존) / 항목5 개정 절·미러 SPEC-ID `0` / 항목6 (a) TOML 이 개정 명령을 담음 (b) golden `ok` (c) 뮤턴트 (a) 로 RED |
| AC-ACD-006 | PASS | corpus 재유도 `606`파일(글롭 깊이 1, `_archive/` 제외), 스냅샷 `.moai/reports/t338/ac-count-baseline.txt` 608행(주석 2 + 606). 차이 0. 뮤턴트 (c) 가 경로를 지목. 항목5 (b)-(e) 네 전이는 `TestACBaselineComparisonTransitions` 가 직접 실행한다 — 현재 corpus 에 `HALT` 파일이 0건이라 그 분기가 실행되지 않는 공허를 막기 위함 |
| AC-ACD-007 | PASS (항목 5a) | 대상 **4건**(0건 아님). 아래 E.2.5 가 플래그별 손 판정을 식별자 단위로 담는다 |
| AC-ACD-008 | PASS | 종단 상태 + 인용 축 종단 이전 + 별칭 축 후보(정규화 1건 제외) **510경로**에 좁힌 `git diff --stat` 무출력. `debt-list.md` 축 3절, 방향 라벨(하한/상한/상한), 총계 줄 `grep -c '^합계\|^총계\|^Total'` → `0`. 항목 수 대조: 인용 124 = `pre-terminal-scan.sh` 재유도 `status=completed`, 별칭 85 = `grep -c '^=== '` 재유도. 폐기 축 절은 면제이며 출처를 E.2.5 로 명시 |

### E.2.5 플래그별 손 판정 (AC-ACD-007 Given 3-4)

재유도: `overcount-detector.sh` → 606파일 스캔 / 20파일 / 41식별자. `pre-terminal-scan.sh` → multi-domain 156, completed **124**, pre-terminal **32**. `alias-shape-scan.sh` → **85**파일.
출력은 `.moai/reports/t338/rederive/` 에 있다. (계획 시점 값 122/34 에서 이동했다 — 재유도 규율대로 이 실행 값을 쓴다.)

**종단 상태 파일의 플래그 24건** — 종단이라 정규화 대상은 아니지만, 부채 목록 절 1 의 출처가 되므로 하나씩 열어 읽었다.

| 식별자 | SPEC | 판정 | 근거(등장 행) |
|---|---|---|---|
| AC-AEL-008 | AGENT-EMIT-LINEAGE-001 | **과다 계상** | `:165` "폐기 — 종전 AC-AEL-008 … 함께 폐기했다" |
| AC-MTP-025 | MODEL-TIER-PLANTYPE-001 | **과다 계상** | `:25` "AC-MTP-025 retired (D5 descope)" |
| AC-PCP-015 | PRECOMMIT-PRESERVE-001 | **과다 계상** | `:389` "the defect the retired AC-PCP-015 had" |
| AC-DCP-010 | AGENT-PARALLEL-OPT-001 | 과다 계상(**인용 축**) | `:248` 남의 SPEC 기준을 인용. 선택 A 로 정규화 대상 아님 |
| AC-DVR-012 | V3R6-DOCS-V3-REBUILD-001 | 과다 계상(**네 번째 모양**) | `:64` 파일 안에 `AC-DVR-012a`/`012b` 로만 등장 — 스윕이 접미 앞에서 끊어 읽는다. 세 축 어디에도 안 들어감(부채 목록 관측 노트) |
| AC-APO-050 · AC-APO-052 · AC-APO-070 | AGENT-PARALLEL-OPT-001 | live | `:222 :224 :248` 온전한 MUST 표 행. 어휘는 판정 대상 서술 |
| AC-002 | ASTGREP-EDIT-001 | live | `:48` "The retired skill paths are gone" — 폐기가 기준의 *주제* |
| AC-CMR-002 | COMPLETION-MARKER-RETIRE-001 | live | `:24` 헤딩 주제 |
| AC-CAR-001 | CONFIG-AUDIT-REPAIR-001 | live | `:25` 기대 출력 문구 안의 "retired" |
| AC-E2E-014 | E2E-REVIVAL-001 | live | `:36` 온전한 검증 행 |
| AC-LSPMCP-RETIRE-007 | LSPMCP-RETIRE-001 | live | `:84` 표 행의 "superseded" 는 판정 대상 |
| AC-WF001-04 · AC-WF001-07 | V3R2-WF-001 | live | `:104 :186` 헤딩 주제 |
| AC-RA-02/07/11/14/17 | V3R3-RETIRED-AGENT-001 | live | `:46 :174 :272 :348 :423` 전부 주제 |
| AC-SDF002-B-10 · B-11 | V3R4-STATUS-DRIFT-FOLLOWUP-002 | live | `:114 :121` 다른 SPEC 이름 안의 "RETIRED" |
| AC-WADA-012 | V3R6-WORKFLOW-AGENT-DOC-ALIGN-001 | live | `:138` "retired terms replaced" — 주제 |
| AC-WC6-019 | WEB-CONSOLE-006 | live | `:38` "MAY be retired" 는 *테스트*의 처분 |

합: live 19 · 과다 계상 5(그중 폐기 축 3). 절 1 의 항목 수 3 이 이 표에서 나온다.

**종단 이전 파일** — 정규화 대상 재유도 결과.

| SPEC | status | 플래그 | 판정 | 처리 |
|---|---|---|---|---|
| AC-COUNT-DISCRIMINATOR-001 | draft | 10건 | 전부 인용 또는 합성 예시 | **대상 아님** — 인용 축은 선택 A, 그리고 이 SPEC 자신의 `acceptance.md` 는 run-phase 가 편집할 수 없다(소유 경계) |
| CONFIG-DEAD-SWEEP-001 | in-progress | AC-CDS-001/002/006 | 전부 과다 계상 | 정규화(`:13-15` 표 셀에 `RETIRED` 판정) |
| GRAPH-FRESHNESS-CADENCE-001 | in-progress | AC-GFC-011/012 | 전부 과다 계상 | 정규화(`:20 :141` `:21 :146` "withdrawn at v0.2.0") — **계획 기대에 없던 파일**(아래 참조) |
| UPDATE-DOC-DRIFT-001 | draft | AC-UDD-006 | 과다 계상 | 정규화. 손 판정으로 **AC-UDD-002 · AC-UDD-003 을 추가**했다(`:574` "RETIRED at v0.3.0", `:778`) — 검출기가 3건 중 1건만 잡는 하한임을 이 파일이 다시 보였다. `AC-UDD-001` 은 live 로 남겼다(회귀 가드) |
| V3R2-ORC-001 | implemented | AC-ORC-001-05 | **live** | 폐기 축에서 제외(오탐). 별칭 축에서 단축 철자 `AC-01`~`AC-17` 을 `[REF]` 로 정규화 |

**계획 기대와의 차이 2건, 근거를 남긴다.**
① `SPEC-GRAPH-FRESHNESS-CADENCE-001` 은 plan-phase 이후 트리에 들어온 파일이라 계획의 기대 목록에 없었다. 재유도가 대상을 정한다는 AC-ACD-007 의 규율대로 포함했다.
② 별칭 축의 종단 이전 후보는 재유도에서 **17파일**이 나왔지만, `spec.md` §6 이 "이 카드가 손대는 별칭 파일은 `SPEC-V3R2-ORC-001` 하나뿐 … 나머지는 전수 판정조차 하지 않는다" 로 이미 범위를 닫아 두었다. 나머지 16파일은 절 3 후보 목록에만 실린다.

**정규화 전/후 값**

| 파일 | 스윕 | 계수기(전) | 계수기(후) | rc | 표시한 등장 수 |
|---|---|---|---|---|---|
| SPEC-CONFIG-DEAD-SWEEP-001 | 18 | 18 | **15** | 0 | 3 |
| SPEC-GRAPH-FRESHNESS-CADENCE-001 | 14 | 14 | **12** | 0 | 4 |
| SPEC-UPDATE-DOC-DRIFT-001 | 27 | 27 | **24** | 0 | 8 |
| SPEC-V3R2-ORC-001 | 34 | 34 | **17** | 0 | 46 |

`SPEC-V3R2-ORC-001` 의 17 은 `spec.md` §2.3 이 손으로 확인한 정본 헤딩 수와 일치한다.
AC-ACD-007 항목 2 대로 `live` 판정 식별자는 계수기에서도 live 다 — `AC-ORC-001-05`(n=1) · `AC-UDD-001`(n=4) · `AC-GFC-001`(n=3) · `AC-CDS-003`(n=2) 확인.

### E.2.6 baseline 스냅샷이 움직인 이유 (AC-ACD-006 항목 4)

M5 정규화로 네 행이 움직였고, 전부 **live → excluded** 이동이다. 새로 생긴 정지는 없다.

```
SPEC-CONFIG-DEAD-SWEEP-001        COUNT 18 live=18 excluded=0  →  COUNT 15 live=15 excluded=3
SPEC-GRAPH-FRESHNESS-CADENCE-001  COUNT 14 live=14 excluded=0  →  COUNT 12 live=12 excluded=2
SPEC-UPDATE-DOC-DRIFT-001         COUNT 27 live=27 excluded=0  →  COUNT 24 live=24 excluded=3
SPEC-V3R2-ORC-001                 COUNT 34 live=34 excluded=0  →  COUNT 17 live=17 excluded=17
```

### E.2.7 이 SPEC 여섯 산출물 회귀 감시

계수기를 여섯 산출물 전부에 다시 돌렸다. `acceptance.md` 만 AC 가 보장하고 나머지 다섯은
어떤 기준도 지키지 않으므로, 이 실행에서 직접 재측정했다.

| 파일 | live | excluded | ambiguous | rc |
|---|---|---|---|---|
| acceptance.md | 24 | 3 | 0 | 0 |
| research.md | 10 | 0 | 0 | 0 |
| spec.md | 41 | 1 | 0 | 0 |
| plan.md | 11 | 1 | 0 | 0 |
| design.md | 2 | 0 | 0 | 0 |
| progress.md | 15 → **47** | 0 | 0 | 0 |

시작 HEAD `0a663b542` 의 값과 다섯 건은 동일하고, `progress.md` 만 `15 → 47` 로 움직였다 —
이 절(`§E.2`)이 판정 기록으로 식별자를 다수 이름으로 담기 때문이며, 그 이동은 이 실행이 만든 것이다.
**여섯 건 모두 `ambiguous=0`** 이므로 정지 회귀는 없다.

> `§E.2` 초고는 실제로 한 번 정지시켰다 — `AC-SYN-002` 와 `AC-SYN-010` 을 표시된 모양과
> 표시 없는 모양으로 함께 담았기 때문이다(`AMBIGUOUS AC-SYN-010 AC-SYN-002`, rc=3).
> `spec.md` §3.4 규칙 1 을 적용해(예외가 아니라 적용) 두 자리의 표시를 산문으로 바꿔 해소했고
> 재측정으로 확인했다. 규약을 예시하는 문서가 자기 규약에 걸린 세 번째 사례다.
corpus 전체: `files=606 halting=0 files-with-excluded=5`(시작 시점 1 → 정규화 4건이 더해졌다).

### E.2.8 잔여 위험

- **통합 충돌 가능성**: `SPEC-GRAPH-FRESHNESS-CADENCE-001` 은 다른 카드(t322)가 소유한 in-progress 파일이다. 이 브랜치의 편집은 4행뿐이지만, 그 카드가 같은 파일을 움직이면 병합에서 만난다. 리드가 통합 순서를 정할 때 읽을 수 있도록 여기 적는다
- **스냅샷 고무도장**: `spec.md` §3.5 · REQ-ACD-006 이 적은 그대로다. 상태별 집계와 이 절의 사유 기재는 가시화이지 방지가 아니다
- **`HALT` 경로의 실데이터 부재**: 현재 corpus 에 정지 파일이 0건이라 스냅샷 파서의 `HALT` 분기는 합성 입력으로만 실행된다

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-28
run_commit_sha: "23df21c9e"   # M1 commit; backfilled in a follow-up commit (a commit cannot cite its own SHA)
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
ac_ids_passed: [AC-ACD-001, AC-ACD-002, AC-ACD-003, AC-ACD-004, AC-ACD-005, AC-ACD-006, AC-ACD-007, AC-ACD-008]
preserve_list_post_run_count: 510   # terminal + citation-axis pre-terminal + alias-axis candidates (minus the one alias target); narrowed `git diff --stat` produced no output
l44_pre_commit_fetch: "git fetch origin develop; git rev-list --count --left-right origin/develop...HEAD -> 19  2 (origin ahead 19, branch ahead 2 — integration is the lead's window, not this run's)"
l44_post_push_fetch: "N/A — this run does not push; the dispatch withholds push and merge pending the lead's reply"
new_warnings_or_lints_introduced: 0   # gofmt -l (empty), go vet ./internal/spec/... ./internal/template/... clean on all three GOOS
cross_platform_build:
  darwin: "GOOS=darwin go vet ./internal/spec/... ./internal/template/... -> OK"
  linux: "GOOS=linux go vet ./internal/spec/... ./internal/template/... -> OK"
  windows: "GOOS=windows go vet ./internal/spec/... ./internal/template/... -> OK"
  note: "GOOS vet proves compilation of the test-excluded build only; runtime behaviour on linux/windows is not established by this run"
tests:
  affected_packages: "go test ./internal/spec/... -count=1 -> ok (29.985s); go test ./internal/template/... -count=1 -> ok (internal/template 24.875s, agentemit 0.536s)"
  full_suite: "NOT RUN locally by instruction (parallel-lane load hazard) — CI on the integration branch is the full-suite judge"
  new_test_cases_executed: 9   # TestAC* top-level tests, 3 of them with subtests; every one observed running
mutants_killed: 3   # mirror one-character; fixture partial-marking + end-of-line token move; corpus token plant. All restored, all re-verified green
total_run_phase_files: 24   # `git status --short` at commit time: 14 modified + 10 new
m1_to_mN_commit_strategy: "single M1 commit on WT-ac-count-sweep carrying M1-M5 (23df21c9e), plus one SHA-backfill follow-up commit; no push, no develop merge (integration awaits the lead)"
```

**파일 24건의 내역** — `git status --short` 를 커밋 직전에 다시 읽어 센 값이다.

| 묶음 | 수 | 경로 |
|---|---|---|
| B12 절 (로컬 + 미러) | 2 | `manager-docs.md` ×2 |
| B12 불릿 (로컬 + 미러) | 2 | `manager-develop-prompt-template.md` ×2 |
| 기계 생성 반송자 | 1 | `.codex/agents/moai/manager-docs.toml` |
| catalog 해시 | 1 | `internal/template/catalog.yaml` |
| 검증자 + fixture | 4 | `internal/spec/ac_count_clause_test.go`, `testdata/ac_count/*.md` ×3 |
| baseline 스냅샷 + 생성기 | 2 | `ac-count-baseline.txt`, `run-scratch/gen-baseline.sh` |
| 부채 목록 + 재유도 출력 | 4 | `debt-list.md`, `rederive/*.txt` ×3 |
| M5 정규화 | 4 | 타 SPEC `acceptance.md` ×4 |
| 이 SPEC 산출물 | 4 | `spec.md` · `plan.md` · `acceptance.md`(frontmatter `status`/`updated` 만) + `progress.md` |

계획 §F 는 16행 / 18파일을 셌다. 차이는 두 갈래다 — 정규화 대상이 3 → **4**(E.2.5 ①),
그리고 §F 가 세지 않은 증거 파일(재유도 출력 3 · 스냅샷 생성기 1)을 이 실행이 커밋에 넣었다.
어느 쪽으로 세도 Tier **L** 이다.

> **`design.md` · `research.md` 는 `draft` 로 남겼다.** 소유권 매트릭스가 M1 전이의 대상으로
> 명시하는 것은 4종(spec / plan / acceptance / progress)이고, Tier L 이 더한 두 산출물은
> 그 열거에 없다. 6종을 일괄로 뒤집는 것이 더 자연스러워 보이지만, 문서화되지 않은 전이를
> 임의로 넓히는 쪽이 드리프트다 — 미해결 항목으로 적어 둔다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
