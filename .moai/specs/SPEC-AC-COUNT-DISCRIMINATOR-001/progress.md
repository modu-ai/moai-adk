---
id: SPEC-AC-COUNT-DISCRIMINATOR-001
title: "AC 개수 자가검사 판별자 — 진행 기록"
version: "0.4.1"
status: draft
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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
