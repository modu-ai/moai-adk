# Acceptance — SPEC-AC-LOCALE-TOKEN-001

> 하니스: **standard** · 측정 근거: `.moai/reports/t573/plan-research.md` (트리 `3ac58b5a1`, 2026-09-08)
> 모든 검증 grep 은 `/usr/bin/grep`. RED-now AC 는 baseline(현재) vs after 로 기술된다.

## HISTORY

- 2026-09-08: AC-004 검증식 재스코핑 — 문자열 계수 프록시에서 **zh 대상 라이브 사용 계수**로. 기준(`0`)은 그대로다.
  - before: `/usr/bin/grep -c 'grep -ci "desktop-native"' .moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md` → run 종료 실측 `5` (기대 `0` — FAIL)
  - after: `awk '/^## §D AC Matrix/{f=1} f && /grep -ci "desktop-native"[^`]*content\/zh/{n++} END{print n+0}' .moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md` → `0` (exit 0)
  - 사유: 프록시는 `grep -ci "desktop-native"` 문자열의 **행 수**를 세는데, run 이 남긴 5행은 zh 대상 라이브 사용 **0** + en 축 유지 1(REQ-004 가 유지를 지시 — en 은 Latin 로케일) + HISTORY·논거 인용 4(AC-005 가 before 명령 인용을 의무화)로 분해된다. 즉 프록시는 자기 SPEC 의 REQ-004·AC-005 가 의무화하는 인용을 위반으로 재는 셈이고, 이는 본 SPEC 이 없애는 결함형 — "문자열의 개수로 라이브 사용을 단정한다" — 의 자기-적용 사례다 (`verification-claim-integrity.md` §1: 참조의 존재는 참조 대상의 생존이 아니다).
  - **이 재스코핑은 통과 기준을 완화하지 않는다.** 본질 요구(zh 대상 왜곡 기준의 라이브 잔존 = 0)는 옛 읽기와 새 읽기 어느 쪽에서도 충족돼 있다 — 바뀐 것은 계측기가 재는 대상이지 요구가 아니다. 도달성은 뮤턴트로 실증했다: en 명령의 경로를 zh 로 바꾼 /tmp 사본에서 새 식은 `1` 을 낸다 (공허한 0 이 아님 — §D.4 대조군 기록).
  - **승인 귀속: 리드 승인 — run-phase 수리 dispatch (카드 t573, manager-spec 재위임; D-NEW-1 inline-fix 경로).**

## §D AC Matrix

| AC | 소관 | REQ | 검증 | 기대 (baseline → after) |
|---|---|---|---|---|
| AC-001 | M1 | REQ-002 | census 장부 존재 + REQ-002 필드 스키마 (RED-now) | 부재 → 존재 |
| AC-002 | M1 | REQ-002·007 | swept-set 계수 명시 (코퍼스·A1/A2/A3·A·A∩B 각 단계, 재측정값) (RED-now) | 없음 → 명시 |
| AC-003 | M2 | REQ-001·003 | 왜곡형 확정 건 전량 재작성 + REQ-001 4요소 충족 | — |
| AC-004 | M2 | REQ-004 | t538 AC-004 zh-명향 **라이브** 사용 0 — 섹션-절단 + 명령-범위 zh 경로 (RED-now → 2026-09-08 재스코핑, §D.4) | 프록시 5 → 라이브 zh 0 |
| AC-005 | M2 | REQ-003 | 재작성마다 t538 규율 HISTORY 항목 | — |
| AC-006 | M2 | REQ-001 | 뮤턴트 probe — 보강-제거 사본에서 재작성 기준 통과 | — |
| AC-007 | M3 | REQ-005 | 옵션 A/B 기준값 연산 제시 + docs-site 무변경 | — |
| AC-008 | 경계 | REQ-006 | 2차 관찰(근접 미스) 기록만, 수정 없음 | — |
| AC-009 | M1 | REQ-007 | A∖B 전수 사람 판독 표 (파일별 판정) | 부재 → 전수 |
| AC-010 | M1 | REQ-007 | 이름 붙은 잔여 철자 항목 4건 판정 기록 | 부재 → 기록 |

## §D.1 AC-001 — census 장부 (RED-now)

- **Given** 트리 `3ac58b5a1` 에서 `ls .moai/reports/t573/` → `No such file or directory` (exit 1 — 실측 2026-09-08).
- **When** `test -f .moai/reports/t573/census.md && /usr/bin/grep -c "AC id" .moai/reports/t573/census.md`
- **Then** baseline 부재 → after 장부 존재. REQ-002 7개 필드(파일, AC id, 토큰, 대상 문서+로케일, 계수 의미, 현재 히트 vs 기준, 판정) 스키마 헤더를 포함한다.

## §D.2 AC-002 — swept-set 계수 명시 (RED-now)

- **Given** 본 SPEC 이전에는 코퍼스에 대한 필터별 통과 계수가 명령과 함께 기록돼 있지 않았다(과거 기록값 17 은 B 명령 미기록으로 폐기 — plan-audit D1).
- **When** census.md 의 코퍼스·A1/A2/A3·A·A∩B·A∖B 각 단계마다 **정확한 명령 + 재측정값**이 있는지 확인한다. plan-시 스냅샷 참고값(트리 `3ac58b5a1` 작업 트리, 자기 SPEC 제외): 코퍼스 `2276` · A1 `55` · A2 `98` · A3 `62` · A `161` · A∩B `66` · A∖B `95` — 이 값들은 **스냅샷이며 기준이 아니다**; census.md 는 M1 시작 시 재측정값을 기록한다 (자기 아티팩트 착지로 계수가 움직인다 — t538 AC-011 동형).
- **Then** after 상태에서 각 단계의 명령+재측정 계수가 기록돼 있고, 어떤 기록 수치든 그 명령으로 재현된다 (REQ-007). 계수 0 인 단계는 "0 — 빈 집합"으로 명시한다.

## §D.3 AC-003 — 왜곡형 전량 재작성

- **Given** census 가 REQ-002 판정 필드로 왜곡형(distorts)을 확정한다. 본 plan 시점 확정 지식: SPEC-DOCS-LOCALE-PARITY-REPAIR-001 AC-004 (1건 — 최소).
- **When** 각 확정 건에 대해 재작성 후 기준 검증식을 실행한다.
- **Then** 각 재작성 기준이 (a) 정확한 계수 명령 (b) 로케일 명시 (c) 실측 기준선 (d) 왜곡 불가 근거를 진술한다 (REQ-001). census 확정 건수와 재작성 건수가 일치한다.

## §D.4 AC-004 — t538 AC-004 zh-명향 라이브 사용 0 (RED-now → 2026-09-08 재스코핑, HISTORY 참조)

- **Given** 최초 프록시는 문자열 계수였다: 트리 `3ac58b5a1` 에서 `/usr/bin/grep -c 'grep -ci "desktop-native"' .moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md` → `1` (RED-now 실측). run 종료 트리 `30915eddb` 에서 같은 프록시는 `5` 를 내 FAIL — 그러나 분해하면 zh 대상 라이브 사용 **0**(본질 요구 충족 — t538 §D.4 zh 축이 `原生桌面` ≥6행 + 매트릭스 3행 구조로 대체됨) + en 축 유지 1(REQ-004 가 지시) + HISTORY·논거 인용 4(AC-005 가 의무화). 프록시는 **인용을 라이브 사용으로 재는 문자열 계수**다 — 본 SPEC 이 없애려는 결함형 그 자체.
- **When** (재스코핑된 검증식 — 단일 호출, 파이프 없음):

  ```
  awk '/^## §D AC Matrix/{f=1} f && /grep -ci "desktop-native"[^`]*content\/zh/{n++} END{print n+0}' .moai/specs/SPEC-DOCS-LOCALE-PARITY-REPAIR-001/acceptance.md
  ```

- **Then** 트리 `30915eddb` 실측 출력 `0`, exit `0`. 제외의 기계적 근거: (a) `^## §D AC Matrix` 시작 조건 — HISTORY 블록(`## HISTORY` … `## §D AC Matrix` 사이)의 before-명령 인용은 절단면 밖이라 계수되지 않는다 (절 이름이 정확한 배제 근거다). (b) `[^`]*content\/zh` — **같은 백틱 명령 안에서** 옛 패턴 바로 뒤에 오는 경로가 zh 일 때만 계수하므로, en 축 유지 명령(`… content/en …`, t538 §D.4 :79)과 경로 없는 인용(t538 §D.4 :81 논거)은 탈락한다.
- **인용-증명 (대조군 + 뮤턴트, 트리 `30915eddb` 실측)**: (A) zh 제약을 빼면 절단면 안의 2행(en 라이브 + 논거 인용)이 잡힌다 — 매치기는 살아 있다. (B) 절단면 필터를 빼면 HISTORY 9행의 zh-명령 인용이 `1` 로 잡힌다 — 절단이 바로 인용 배제의 기제다. (C) 뮤턴트 — en 명령의 경로를 zh 로 바꾼 사본에서 `1` 로 뒤집힌다: 진짜 라이브 zh 사용이 생기면 기준이 발화한다. 즉 출력 `0` 은 공허한 0 이 아니다 (verification-completeness §1.1 빈-집합 가시성).
- **네 요소 (REQ-001)**: (a) 단일 awk 호출 — 읽기 전용, 파이프·연쇄 없음 (b) 대상 명명 — t538 acceptance.md 의 `## §D AC Matrix` 이후 영역에서 zh 경로를 갖는 라이브 옛-기준 행 (c) 기준 `0` — 분해 실측(zh 라이브 0)에서 도출, 완화 아님 (d) 트리 `30915eddb` 핀. 인용(HISTORY, 논거)이 이 기준을 실패시킬 수 없다 — 절단면이 구조적으로 인용 영역을 제외하기 때문이다.

## §D.5 AC-005 — HISTORY 규율

- **Given** t538 acceptance.md 의 HISTORY 선례(8-33행): 날짜 + before/after 명령 + 실측값 + 판정 영향 명시 + 승인 귀속.
- **When** 각 재작성의 HISTORY 항목을 검사한다.
- **Then** 5요소가 모두 있고, 측정값은 재작성 시점 트리 SHA 와 함께 기록돼 있다. 승인 귀속은 "리드 승인 — 카드 t573 발행 dispatch" 이다.

## §D.6 AC-006 — 뮤턴트 probe (왜곡 불가성)

- **Given** zh 문서의 ASCII 보강을 제거한 사본 생성: `sed 's/, desktop-native)/)/g; s/ (desktop-native)//g' docs-site/content/zh/utility-commands/moai-e2e.md > /tmp/t573-mutant.md` (본 트리 실측 — 이 사본의 `desktop-native` 출현 2, `原生桌面` 6행).
- **When** 재작성된 zh 기준의 검증식을 원본과 사본 각각에 실행한다.
- **Then** 사본에서도 기준이 통과한다. 사본에서 실패하는 기준은 ASCII 보강에 여전히 의존하므로 반려하고 재작성한다.

## §D.7 AC-007 — M3 옵션 제시 + docs-site 무변경 (RED-now)

- **Given** 트리 `3ac58b5a1` 기준 docs-site 는 본 SPEC 의 변경 대상이 아니다 (M3 실행 금지 — REQ-005).
- **When** run 종료 시 `git status --porcelain -- docs-site` 와 M3 옵션 문서(§F M3)를 검사한다.
- **Then** docs-site 변경 0. 옵션 A/B 각각의 재작성-기준 기준값이 실측 명령과 함께 적혀 있다 (A: `原生桌面` 6행; B: 되돌림 사본에서 6행 — 둘 다 통과, `/usr/bin/grep -c "原生桌面" /tmp/t573-zh-reverted.md` → `6` 실측).

## §D.8 AC-008 — 2차 관찰 경계

- **Given** census 가 계열 근접 미스(단위 불일치, 무한 도메인 vs 고정 기준선)를 발견할 수 있다.
- **When** census.md 의 2차 관찰 절을 검사한다.
- **Then** 기록은 있으되 해당 기준에 대한 수정 커밋은 없다 (REQ-006). 경계 위반은 M2 범위 초과로 분류한다.

## §D.9 AC-009 — A∖B 전수 사람 판독 (RED-now)

- **Given** plan-audit D4 — grep 계수식이 없는 A 소속 파일(A∖B, plan-시 스냅샷 95건)은 awk·rg·wc·python 등 비-grep 계수의 은신처이며, 아무도 읽지 않으면 census 의 "전수" 주장이 스스로의 경계를 못 지킨다.
- **When** census.md 의 A∖B 판독 표에서 파일 수를 센다: `grep -c '^| \.moai' census.md` 등 표 행 계수 — 파일별 판정(benign / 비-grep 계수 관찰 / 왜곡형)이 한 줄씩 있다.
- **Then** 판독 표의 파일 수 ≥ 재측정된 A∖B 수. 누락 파일이 있으면 census 는 미완이다.

## §D.10 AC-010 — 이름 붙은 잔여 철자 판정 (RED-now)

- **Given** plan-audit 이 A 필터 밖의 우회 철자 4류를 실재로 관측했다: 접두사 없는 `content/zh/`(SPEC-I18N-001-ARCHIVED), `$loc` 보간(SPEC-CC2178-TEAM-API-ALIGN-001:228), 접두사 없는 `ko/advanced/…`(SPEC-DOCS-V313-CATCHUP-001 spec.md:64), `.moai/docs/*.md` 대상 계수(SPEC-VERSION-STAMP-GUARD-001 acceptance.md:74).
- **When** census.md 에 이 4건 각각의 판정 항목이 있는지 확인한다.
- **Then** 4건 전부 판정 행을 갖는다 (왜곡형이면 M2 대상 편입, benign 이면 근거 명시).

## §D.11 품질 게이트

- 문서 SPEC 이므로 Go 커버리지/빌드 게이트는 부적용. 대신: (1) 모든 수치가 명령+출력+트리 SHA 귀속을 갖는다 (REQ-007, VCI §2 — AC-002·AC-009·AC-010 이 이 요구의 검증이다), (2) 빈 swept-set 통과 0, (3) `moai spec lint` — **plan-phase 에서 직접 실행해 흡수했다**: `moai spec lint .moai/specs/SPEC-AC-LOCALE-TOKEN-001/spec.md` → `✓ No findings — all SPEC documents are valid`, exit 0 (이 워크트리, 2026-09-08). 주석: 이 린터의 REQ 수집 패턴은 1-세그먼트 `REQ-001` 형 ID 를 수집하지 못하므로 `No findings` 는 REQ 레이어 판정이 아니다 — GEARS 형식 판정은 plan-audit 이 수동 수행한다 (plan-audit D8).
- 완료 정의(DoD): census 확정 건 = 재작성 건, AC-004 RED-now 해소, A∖B 전수 판독, M3 옵션 문서 존재, docs-site 무변경.
