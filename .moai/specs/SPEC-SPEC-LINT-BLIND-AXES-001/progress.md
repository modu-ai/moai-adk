# SPEC-SPEC-LINT-BLIND-AXES-001 — 진행 기록

카드: t518 · 브랜치: WT-spec-lint-axes · base: 0b1e27877

## §E.1 Plan-phase Audit-Ready Signal

- 아티팩트: spec.md · plan.md · acceptance.md · progress.md (Tier M)
- SPEC ID 정규식 검사: `PASS` (Bash 실행, 출력 인용은 완료 보고 참조)
- baseline 귀속: `moai spec lint` @ `0b1e27877` → `0 error(s), 4346 warning(s)`, 증거 `.moai/reports/t518/baseline-lint.txt`. **저장값이 아니라 시점 재유도값이며**, 문서 저작 이후 재측정하면 4,366(증분 +20 = 이 SPEC 12 + 형제 8)

### 수리 1회차 (v0.2.0, 2026-09-07)

- 판정문: `.moai/reports/t518/plan-audit-blind-axes.md` — FAIL 0.67 (Tier M 임계 0.80), 결함 15건 중 blocking 12건
- 닫은 blocking 결함: D-1(BINLAG 반례 정정 + 후보 4 폐기 + 실측 비용표) · D-2(코퍼스 수치 재유도 표기 + AC-SLB-009 동일 모집단 재유도) · D-3(`OwnershipTransitionInvalid 1` 추가, 합 4,346 일치) · D-4(AC 수집기 out-of-scope 절 + `CoverageIncomplete` 교란 해석 규칙 + t528 의존) · D-5(기존 계측기 2개 인용, 신규 하네스 폐기) · D-6(14개 AC 전부 REQ 인용 + REQ-SLB-008 전용 AC 신설) · D-7(수리 전 골든 제거, narrow 경로 대조로 대체) · D-8(002·005·007·008 뮤턴트) · D-9(AC-SLB-007 픽스처 축자 고정) · D-10(REQ-SLB-005 라벨 `Event-driven`) · D-11(형제 SPEC ID 명기 + `related_specs`) · D-12(축 2 표 잠정 표기 + 볼드 해제 + 방법 결함 기록)
- 함께 처리한 optional: D-13(AC-SLB-010 재작성) · D-14(AC-SLB-001c 신설) · D-15(무판정 코드명 중립화)
- 요구사항 12건 유지 · 수용 기준 식별자 14건(기준 단위 10건) — 둘 다 Tier M 상한 16 이하
- 수집기 정규식 재유도: 이 spec.md의 REQ 정의 행 **12건 전부 수집**(목록 형식 유지 확인)
- **운영자 결재 대기 3건**(spec.md §I): D1 판별식 확정 · 축 2 갈래 A 포함 여부 · 자문 전파 `FromTable` 필드. **Implementation Kickoff Approval 전에 닫혀야 한다.**

### 수리 2회차 (v0.3.0, 2026-09-07)

- 판정문: `.moai/reports/t518/plan-audit-iter2.md` — FAIL 0.75 (Tier M 임계 0.80), blocking 5건 + optional 4건
- 닫은 blocking: **B-1**(§A:62 처분 행 수 10→8 정정, §I와 일치. 전수 확인: `grep -nE '^\s*\|\s*\*{0,2}\s*REQ-' .moai/specs/SPEC-BINLAG-INVOCATION-001/spec.md` → 처분 `:124-131` 8행 · 추적 `:140-147` 8행) · **B-2**(C-d·C-e 어휘 L1·L2를 §I에 축자로 싣고 재유도) · **B-3**(§A 빈 heredoc → 실행 가능한 판독기 `.moai/reports/t518/blind-axes-reader.py` + 관측 출력 축자 인용) · **B-4**(acceptance.md §E 체크박스를 §F와 일치 — 「12개 AC」는 실재하지 않았고 REQ-SLB-011은 대응 AC가 없다) · **B-5**(§F t525 라벨 `선행`→`후행`)
- **[HARD] C-d·C-e 수치 철회**: v0.2.0의 31·90은 어휘 목록이 문서에 없어 재유도 불가였다. L1(3개)·L2(16개)를 §I에 축자로 싣고 다시 재면 **C-d 20 · C-e 61**이다. 옛 값은 철회하며 더 이상 인용하지 않는다. 판독기가 §A의 재유도 가능한 수치 전부(788/786/3952/53/600/top-6/C-a 523/C-b 510/BINLAG 6·7·16·16)를 그대로 재현하므로 판독기 자체는 검증됐다. **어휘 민감도 기록**: 재감사의 독립 어휘는 C-d 20(일치) · C-e 63(2행 차) — 좁은 어휘는 재현되고 넓힌 어휘는 목록에 따라 흔들린다
- **신설 게이트**: `BLIND-AXES-DISCRIMINATOR-GATE`(spec.md §I.0) — **상태 OPEN.** 판별식이 미확정인 동안 run-phase 진입 금지(코드·테스트·M1·`draft → in-progress` 전부). 근거: 후보 선택이 AC-SLB-004의 통과 가능성 자체를 바꾼다(C-a면 착수 즉시 FAIL)
- **운영자 결재 완료 1건**: 미해결 3(자문 전파 경로) → **심각도 결정점 하나(기존 `reqFindingSeverity`/`Widened` 재사용) + 별도 출처 필드**. §B.2 신설, REQ-SLB-013 + AC-SLB-011 신설, REQ-SLB-002·009 개정, plan.md D3·M2·M4 반영. 기각한 대안(`Widened` 단독)과 그 대가(총량은 나오되 귀속이 나오지 않음)를 함께 기록
- 요구사항 **13건**(재유도: 수집기 정규식 grep → `13`, Tier M 상한 16 이하) · 수용 기준 식별자 **15건**(기준 단위 11건) — 둘 다 상한 이하. 묶음 셈법은 v0.2.0에서 바꾸지 않았다
- **남은 운영자 결재 2건**(spec.md §I): 미해결 1(판별식 — 게이트가 걸려 있다) · 미해결 2(축 2 갈래 A 포함 여부)

### 수리 3회차 + 미해결 1 결재 (v0.4.0, 2026-09-07)

- 판정문: `.moai/reports/t518/plan-audit-iter3.md` — **PASS 0.800**(Tier M 임계 0.80), PASS-with-debt. blocking 3건 중 이 SPEC 소관 2건(C-1 · C-2)을 상환했다. 재감사 없이 닫을 수 있다는 판정문의 처분을 따랐다
- **닫은 blocking — C-1**(`spec.md §I` L2 개수 라벨): 「L1의 3개 + 아래 **13개** = **16개**」 → 「L1의 3개 + 아래 **16개** = **19개**」. 덧수도 합도 틀렸었다. 실측: 문서 열거 백틱 토큰 **16개** · 판독기 `len(L1)`=**3** · `len(L2)`=**19** · `L2[3:]`와 문서 열거 **바이트 동일**(`True`). **원문을 지우지 않고 정정을 병기**했고, 라벨-열거 일치를 재는 **재유도 명령을 블록 안에 신설**했다(줄 번호가 아니라 불릿 라벨에 앵커돼 있어 위쪽에 줄이 끼어들어도 돈다). 관측: `L1 3 | L2 추가 16 | L2 합 19`
- **[HARD] C-1 정정 후 재유도 — 결재가 이 블록 위에서 이뤄졌으므로 라벨만 고치고 끝내지 않았다.** 판독기를 그대로 재실행해 출력 전문을 `spec.md §I`에 인용했다: `survivors: {'C-a': 523, 'C-b': 510, 'C-d': 20, 'C-e': 61}` · `BINLAG rejected out of 16: {'C-a': 6, 'C-b': 7, 'C-d': 16, 'C-e': 16}`(rc=0, 파이프 없음). **결재된 C-d = 20 불변.** 예상된 결과이지 요식이 아니다 — C-1은 **L2** 라벨의 오류였고 C-d는 **L1**만 쓰므로, 확인이 필요했던 것은 「C-d가 L2와 무관하다」는 사실 자체다. C-e 61도 불변(정본은 라벨이 아니라 열거였다)
- **닫은 blocking — C-2**(`acceptance.md`의 낡은 12-셈법): `:119`를 「AC 12개 … 12건」 → 「**AC 15개** … REQ **13건**」으로 `spec.md:185` 및 실측과 일치시켰다. **감사가 지목하지 않은 두 번째 인스턴스를 전수 훑기에서 찾았다** — `:21`의 인용-형식 주석도 「`CoverageIncomplete` 12건」을 안고 있었고 함께 정정했다. **조용히 고치지 않고 기록한다**
- **C-2 재발 방지 — `acceptance.md §A` 규칙 8 신설.** B-4(v0.3.0)와 C-2(v0.4.0)가 **한 회차 간격으로 같은 파일에서 같은 모양**으로 났으므로, 개별 정정으로는 세 번째를 막지 못한다고 보고 **수치 정본을 한 곳에 못박고 재유도 명령 3개를 그 옆에 뒀다**(AC 15 · REQ 13 · lint 경고 13). 인용하는 자리는 숫자를 홀로 두지 않고 규칙 8을 가리킨다. 이것은 재발을 막는 장치가 아니라 **재발을 관측 가능하게** 만드는 장치다
- 실측 재유도(파이프 없음): `grep -oE '^### AC-SLB-[0-9]+[a-z]?' acceptance.md | sort -u | wc -l` → **15** · 수집기 정규식 grep → **13** · `moai spec lint …/spec.md` → rc=**0**, `0 error(s), 13 warning(s)`
- **[운영자 결재 완료] 미해결 1 = 후보 `C-d`(좁은 어휘 L1). `BLIND-AXES-DISCRIMINATOR-GATE` → CLOSED.** 세 자리 전부에 반영: `spec.md §I.0`(상태 줄 + 무엇이 풀리는가) · `spec.md §I` 「미해결 1 — 결재 완료」(근거 전문) · `plan.md §B` + `§D1`(게이트 문단 + 결재 배너). 게이트가 막던 것(코드·테스트 작성, M1 착수, `draft → in-progress`)은 이 닫힘만으로 풀린다
- **[HARD] 선택 기준은 잔존 집합의 크기가 아니라 독립 유도 사이의 재현성이다.** C-d가 20으로 가장 작아서 뽑힌 것이 **아니다** — 표의 잔존 열만 읽으면 그렇게 읽히므로 문서에 명시했다. **C-d는 두 독립 어휘(이 SPEC의 L1 · 재감사가 따로 구성한 어휘)에서 같은 20을 냈고, C-e는 같은 두 어휘 사이에서 61 ↔ 63으로 움직인다.** 이 카드가 재유도 불가 수치에 두 번 데었으므로(v0.2.0의 31·90 철회 = B-2, 그 철회를 위해 만든 블록 안의 라벨 오류 = C-1), 운영자는 **목록의 저자와 함께 움직이지 않는 팔**을 골랐다
- **[HARD] 오탐(진짜 정의를 놓침)은 잔여 위험이 아니라 선택된 대가다.** 측정된 실물: `SPEC-INIT-001`의 `| REQ-N-001 | 시스템은 … 덮어쓰지 않아야 한다 |` — 진짜 정의인데 종결형이 L1에 닿지 않아 기각된다. 부채·TODO·후속 카드로 적지 **않는다.** 받아들이는 이유: 놓친 행은 REQ-SLB-005의 기각 기록 덕분에 **관측 가능한 미수집**으로 남아, 이 카드가 고치는 결함(공허한 초록)과 성질이 다르다. **나중에 L1을 넓혀 이것을 「고치지」 않는다** — 넓힌 목록은 재현성을 잃은 목록이며, 그것이 C-e를 기각한 바로 그 이유다. 놓침 축소는 어휘 확장이 아니라 어휘에 의존하지 않는 판별식으로 가야 하고, 그때는 별도 결재가 필요하다
- **이 닫힘이 함의하지 않는 것**: Implementation Kickoff Approval은 **별개의 미통과 게이트**이며 이 CLOSED를 착수 승인으로 읽지 않는다. **미해결 2(축 2 갈래 A 포함 여부)는 여전히 OPEN**이고, 형제 SPEC의 §H 미해결 1도 OPEN이다. 미해결 3은 v0.3.0 결재 완료 상태 그대로다
- **감사가 판정을 요구한 두 기각의 처분**:
  - **O-1(AC-SLB-008b 준-공허성) — 기각 유지, 감사도 타당하다고 판정.** 이 AC는 두 갈래이고 갈래 ②(무판정 신호 1건)가 출현 단언이라 비공허하므로, `§A` 규칙 4가 이 AC **안에서** 충족된다. 변경 없음
  - **O-2(`parseREQs` 미동결) — 결론(수리하지 않음) 유지, 그러나 논거를 정정한다.** v0.3.0이 든 논거는 「한 줄 동결은 수용 기준 안에 구현 제약을 넣는 일이므로 **거처가 없다**」였다. **그 논거는 성립하지 않는다** — 이 문서 자신이 반증한다: `REQ-SLB-002`는 `reqFindingSeverity`/`REQEntry.Widened`를 이름으로 지목한 구현 제약을 REQ 안에 담고 있고, `REQ-SLB-013`은 「쓰여서는 안 되며」라는 같은 종류이며, 형제 SPEC의 `REQ-SLI-008`은 순수한 don't-touch 제약을 REQ로 세운다. 게다가 `spec.md §D 제약`이 이런 항목의 자연스러운 거처이고 이미 파일 경로를 이름 대고 있다. **옳은 답은 「넣을 수 없다」가 아니라 「§D에 넣으면 된다」이며, 옛 논거는 합리화에 가까웠다.** 그럼에도 **수리하지 않는 결론은 유지한다 — 근거는 거처가 아니라 잔여가 얇다는 것**이다: 잔여는 「run-phase가 `parseREQs`를 건드리면 AC-SLB-001a/003의 대조가 자기 자신과의 비교가 되어 조용히 공허해진다」 하나이며, 이는 optional에 해당하고 FAIL을 만들지 않는다. 수리한다면 AC 본문이 아니라 `spec.md §D 제약`에 한 줄이다
- **감사가 남긴 optional 4건은 이번 회차에서 처리하지 않았다**(C-3 AC-SLB-011 갈래 ②의 뒤집기 주입 지점 · C-4 축 2 잠정 표 5수치의 명령 부재 · C-5 `progress.md:24`의 3952 · O-2). 전부 판정문에 이름과 함께 남아 있으며, 부재가 아니라 **미처리로 기록**한다
- 요구사항 **13건** · 수용 기준 식별자 **15건**(기준 단위 11건) — v0.3.0에서 변동 없음. 이번 회차는 수치를 만들지 않고 **낡은 수치를 정본에 맞췄다**
- **판정문의 do-not-disturb 항목 무이동 확인**: 판독기 바이트 동일 재현(위 출력) · BINLAG 행 열거(처분 `:124-131` 8행 · 추적 `:140-147` 8행) 미변경 · B-2 철회 기록(31·90 → 20·61, 어휘 민감도 63) 미변경 · `BLIND-AXES-DISCRIMINATOR-GATE`는 **운영자 결재로 닫혔지 임의로 지워지지 않았다**

### 미해결 2 결재 (v0.5.0, 2026-09-07)

- **판정문 없음 — 이 회차는 감사 부채 상환이 아니라 운영자 결재의 착지다.** iteration-3(`PASS 0.800`)에서 이 SPEC 소관 blocking은 v0.4.0에 전부 닫혔고, 남아 있던 것은 **운영자 결재 1건(미해결 2)**뿐이었다
- **[운영자 결재 완료] 미해결 2 = 갈래 B(무판정 발화)만. 갈래 A(한국어 modality 판정)는 후속 카드.** 반영한 자리: `spec.md §I` 「미해결 2 — 결재 완료」(근거 전문, 원문 병기) · `spec.md §E` 신설 절 「Out of Scope — 축 2 갈래 A」 · `plan.md §B` D2 결재 배너 + `§F` M3 · `acceptance.md` AC-SLB-012
- **[HARD] 근거는 순서다 — 어휘가 곧 판정 규칙이 되는 축을 같은 회차에 두 번 열지 않는다.** 한국어 modality 문법을 정의하는 것은 **그 어휘 목록 자체를 판정 규칙으로 세우는 일**이고, 이 카드는 바로 직전 회차(v0.4.0)에 판별식 축에서 같은 성질의 결정을 내렸다 — `C-d`(좁은 어휘 L1)를 고른 **이유가 「어휘 하나가 흔들리면 값이 흔들린다」**였다(C-e 61 ↔ 63 실측). 그 실측을 근거로 좁은 어휘를 고른 회차에 두 번째 어휘-판정 축을 새로 여는 것은 순서가 맞지 않는다. 재계수 해석이 이미 t528을 기다리며 유보돼 있는 사정과도 맞물린다 — 같은 회차에 축 2 판정까지 세우면 증감이 「표 수집 / t385 넓힘 / 새 한국어 판정」 셋의 합이 되어 귀속이 한 갈래 더 갈라진다
- **[HARD] 이 결재가 남기는 산출물: 판정 불가 요구사항이 정확히 몇 건인지 세는 것.** 「대략」이 아니라 정확한 수이며, 그 수가 갈래 A를 받을 후속 카드의 크기를 정한다. 축 2 §A 표의 여섯 줄이 아직 전부 잠정값인 것(판독 기준 미고정)도 이 때문이고, M1이 기준을 고정하고 M3이 발화를 세우면 확정된다
- **[HARD] 받아들인 대가는 부채가 아니라 결정된 대가다.** 이 카드가 착지해도 **한국어 SPEC은 여전히 판정되지 않는다.** 달라지는 것은 하나뿐이고 그 하나가 이 카드가 사는 값이다 — 판정되지 않은 요구사항이 **조용히 통과하는 대신 「판정 불가」로 눈에 보이게** 된다. TODO·부채·후속 카드 표기로 적지 **않는** 이유는 미해결 1의 놓침을 그렇게 적지 않은 이유와 같다: 모른 채 남은 것이 아니라 알고 고른 것이다
- **결재에 함께 들어온 세부 — `" SHALL"` 접촉 조건 → 단어 경계.** `internal/spec/lint.go`의 `isModalityMalformed`는 다섯 접두사(`WHEN `/`WHILE `/`WHERE `/`IF `/`THE `) 각각에서 `strings.Contains(upper, " SHALL")`로 SHALL을 본다(실측: `lint.go:793,796,799,802,806` 다섯 자리, 전부 같은 함수 안). **선행 공백을 요구하므로 문자열 첫머리의 `SHALL`과 구두점 뒤의 `SHALL`을 놓친다** — 코퍼스가 실제로 쓰는 `해야 한다(SHALL)` 꼴이 그 사각이다. 단어 경계 매치로 교체한다. **REQ-SLB-014** 신설 + **AC-SLB-012** 신설 + `plan.md` D2-a 신설
- **[HARD] 「움직인 수치를 공표하고 다른 이유를 적는다」를 주석이 아니라 요구·기준으로 세웠다.** 이 카드는 무기록 수치 교체에 **두 번** 데었다 — v0.2.0의 31·90 철회(어휘 없이 실려 재유도 불가, B-2)와 그 철회를 위해 만든 블록 안의 개수 라벨 오류(C-1). 세 번째를 문서 규율이 아니라 **게이트**로 막는다: AC-SLB-012의 둘째 갈래는 두 재유도값 · 코드별 증감 · 옛 값과 새 값 병기 **셋 중 하나라도 없으면 FAIL**이며, **증감이 0이어도 공표 의무는 남는다**(「움직이지 않았다」도 잰 결과다)
- **[HARD] 이 세부는 갈래 A가 아니라 갈래 B 안쪽이다 — 범위 확대로 읽지 않는다.** 접촉 조건은 한국어를 **판정**하는 장치가 아니라 **무엇을 「판정 불가」로 셀지 가르는 판별식**이다. 조건이 틀린 채면 이 결재의 산출물인 「정확한 판정 불가 건수」가 처음부터 틀린 수가 된다. 이 문장을 `spec.md` REQ-SLB-014 · §E 배제 절 · `acceptance.md` AC-SLB-012 경계 항 · `plan.md` D2-a 넷 모두에 적어, 다음 독자가 갈래 A 유예와의 모순으로 읽지 않게 했다
- **수치 이동 — 세 값이 함께 움직였고, 옛 값을 지우지 않고 이동으로 적었다.** REQ **13 → 14** · AC 식별자 **15 → 16**(기준 단위 11 → 12) · 이 SPEC의 lint 경고 **13 → 14**. 셋이 같은 방향으로 움직인 것은 우연이 아니다 — lint 경고는 전부 `CoverageIncomplete`이고 REQ 하나당 한 건씩 발화하므로 REQ 수를 따라간다. **`acceptance.md §A` 규칙 8(수치 정본)의 세 값과 재유도 명령 셋을 함께 갱신**했고, 규칙이 자기 값을 바꿀 때도 자기가 요구하는 규율(옛 값 병기)을 지켰다
- **AC 식별자 16은 Tier M 상한과 같은 값이다.** 규칙 6에 명시했다 — 다음 회차에 AC가 하나라도 더 늘면 **tier 상향 또는 범위 분할의 신호**이지 셈법을 다시 묶을 사유가 아니다. 이 카드는 v0.2.0의 묶음을 「천장을 맞추기 위한 사후 회계」로 철회했으므로 같은 일을 반대 방향으로도 하지 않는다
- **실측 재유도(파이프 없음, 이 트리)**: `grep -cE … REQ-SLB … spec.md` → **14** · `grep -oE '^### AC-SLB-[0-9]+[a-z]?' acceptance.md | sort -u | wc -l` → **16** · `moai spec lint …/spec.md` → rc=**0**, `0 error(s), 14 warning(s)`(전부 `CoverageIncomplete`)
- **함께 정정한 오기 — 감사가 지목하지 않았고 조용히 고치지 않았다.** `spec.md §I.0`이 「미해결 **2·4**도 여전히 열려 있고」라 적었으나 **「미해결 4」는 이 문서에 존재한 적이 없다** — §I가 연 결정은 1·2·3 셋뿐이고 3은 v0.3.0에 §B.2로 이관됐다. 원문을 지우지 않고 정정을 병기했다(있지도 않은 열린 결정을 찾는 다음 독자를 막기 위해서다)
- **남은 운영자 결재 0건.** §I가 연 결정 셋(1·2·3)이 전부 닫혔다. **Implementation Kickoff Approval은 여전히 별개의 미통과 게이트**이며, 이 결재를 착수 승인으로 읽지 않는다
- **판정문의 do-not-disturb 항목 무이동 확인**: `BLIND-AXES-DISCRIMINATOR-GATE` CLOSED와 C-d 선택 근거(재현성, 잔존 집합 크기 아님) 미변경 · `SPEC-INIT-001` 오탐을 선택된 대가로 두고 「어휘를 넓혀 고치지 않는다」 [HARD] 미변경 · L1/L2 개수(3 / 16 / 19)와 원문 병기 미변경 · 판독기 출력(C-a 523 · C-b 510 · C-d 20 · C-e 61) 미인용·미변경 · O-2의 「잔여가 얇다」 논거 미변경 · BINLAG 행 열거 미변경 · I-1의 카드 소관 아님(형제 SPEC)
- **이번 회차에 변경한 아티팩트**: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` **전부 내용 변경이 있다.** 버전만 오르고 내용이 그대로인 파일은 이 회차에 **없다** — 형제 SPEC이 v0.4.0에서 겪은 간극(내용 없는 버전 상승)을 반복하지 않았음을 확인해 적는다

## §E.2 Run-phase Evidence

### M-A1 — 축 1: 표 형식 REQ 수집 (카드 t518, cycle_type=tdd)

착수 시점 트리: 워크트리 `.claude/worktrees/t518`, 브랜치 `WT-spec-lint-axes`, HEAD `b7ceffa69`. 판별식은 결재된 **C-d**(좁은 어휘 L1)만 세웠다 — 남은 후보를 다시 저울질하지 않았다.

**변경한 파일 3개**: `internal/spec/lint_req_table.go`(신규 — 표 행 정규식 · L1 · C-d · 수집기 · 병합) · `internal/spec/lint.go`(`REQEntry.Source` 필드 + `REQSource` 타입만) · `internal/spec/lint_req_widen.go`(`parseREQsWithProvenance`가 표 항목을 행 순서로 병합). 테스트 `internal/spec/lint_req_table_test.go`(신규). `isModalityMalformed`는 손대지 않았다(M-A2 소관), `internal/cli`도 손대지 않았다(형제 SPEC 소관).

#### AC 판정표

| AC | REQ | 판정 | 실측 근거 |
|---|---|---|---|
| AC-SLB-001a | 001 · 003 | **PASS** | `TestTableCollection_ListFormUnchanged` — 목록 픽스처에서 live 결과가 `parseREQs`와 ID·본문·행번호 동일, 자문 0건, 출처 list, 항목 수 동일 |
| AC-SLB-001b | 001 · 002 | **PASS** | `TestTableCollection_TableFormCollected` — 정의 표 3행 → 3항목, 전 항목 자문, 전 항목 출처 table, 본문 셀 축자 일치 |
| AC-SLB-001c | 012 | **PASS** | `TestTableCollection_ControlPairDiverges` — 목록 자문 0 vs 표 자문 3, 단언으로 갈림(주석·체크박스 아님) |
| AC-SLB-002 | 002 | **PASS** | `TestTableCollection_AdvisoryDoesNotGate` — 표 픽스처 `ModalityMalformed` 1건이 **발화하고** 자문, error 등급 0건. 대조군(목록)은 error 등급 비자문 1건 |
| AC-SLB-003 | 003 | **PASS(제한 있음)** | `TestTableCollection_CorpusListFindingsUnchanged` — 코퍼스 791개 문서, 코드별 `narrow == live-비자문`. **제한**: narrow 경로가 코퍼스 전체에서 `LegacyEARSKeyword` 7건만 내므로(narrow 정규식이 거의 매치하지 않는다 — 이 SPEC §B.1이 인용한 사실) 이 단언의 변별력은 그 7건과 순서 의존 `DuplicateREQID` 위험에 한정된다. 공허하지는 않다(M5가 RED로 만든다) |
| AC-SLB-004 | 004 | **PASS** | `TestTableCollection_DiscriminatorRejectsDispositionTables` — BINLAG `:124`(처분) · `:140`(추적) 축자 픽스처에서 수집 0. 부재 단언이므로 **M2 뮤턴트와 짝으로만 읽는다** |
| AC-SLB-010 | 002 · 010 | **부분 PASS** | 표 절반은 AC-SLB-002가 잼(error 등급 무이동). **006b(무판정 발화) 절반은 M-A2 미착지로 미측정** |
| AC-SLB-011 | 013 · 002 | **PASS** | ① `TestTableCollection_SourceRecordedPerOrigin` ② `TestTableCollection_SourceDoesNotDecideSeverity` — 출처를 전부 뒤집어도 severity 분포 동일 |

**이 마일스톤이 잡지 않은 AC**: 005(기각의 관측 가능성 — 전용 finding 코드와 발화 지점이 필요하며 이 마일스톤의 수집기 범위 밖) · 006a · 006b · 007 · 008 · 008b · 012(전부 축 2 = M-A2) · 009(재계수 = M4).

#### 뮤턴트 8건 — 잡히지 않은 것도 적는다

증거: `.moai/reports/t518/mutants-MA1.txt`, 재현: `python3 .moai/reports/t518/mutants-MA1.py`.

| 뮤턴트 | 결과 | RED가 된 테스트 |
|---|---|---|
| M1 표 수집 되돌림(`parseREQsTable` → nil) | CAUGHT | 001b · 001c · 002 · 경계 |
| M2 판별식 C-d 제거(모든 ID-선두 행 수집) | CAUGHT | 004 · 경계 · 코퍼스 census |
| M3 자문 표시 제거(`Widened: false`) | CAUGHT | 001b · 001c · 002 |
| M4 출처를 `reqFindingSeverity`에 배선 | CAUGHT | 011② · 002 |
| M5 목록 분기 한 글자 훼손(`[-*]`→`[*]`) | CAUGHT | 001a · 001c · 002 · **003(코퍼스)** |
| M6 [경계] L1 매칭을 대소문자 무시로 넓힘 | **1차 NOT CAUGHT → 가드 신설 후 CAUGHT** | 경계 |
| M7 [경계] 행 정규식에서 볼드 마커 허용 제거 | **1차 NOT CAUGHT → 가드 신설 후 CAUGHT** | 경계 |
| M8 [경계] 본문 셀을 ID 뒤 **첫** 셀에서 취함 | CAUGHT | 001b · 002 |

**[HARD] M6·M7은 처음에 통과했다 — 지우지 않고 적는다.** 두 뮤턴트는 픽스처 어디에도 그 성질을 거는 행이 없어 전 스위트를 초록으로 통과했다. M7이 특히 무거웠다: 코퍼스의 실제 정의 표(`SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001`의 4행)가 `| **REQ-HCW-001** |` 꼴 볼드 ID를 쓰므로, 잡히지 않은 채였다면 볼드 허용이 장식인지 필수인지 아무도 알 수 없었다. `TestTableCollection_DiscriminatorBoundaries`(소문자 `shall` 행은 기각 · 볼드 ID 행은 수집)를 신설해 둘 다 CAUGHT로 바꿨다. **기록을 남기는 이유**: 지금 CAUGHT인 것은 가드를 나중에 붙였기 때문이지 처음부터 가드가 있었기 때문이 아니다.

#### 코퍼스 이동 — +16, 전부 `CoverageIncomplete`, 행 단위로 귀속됨

**[HARD] 두 값 모두 같은 트리·같은 모집단에서 재유도한 값이며, 저장된 숫자와 빼지 않았다.** 설치본 바이너리는 baseline 재사용에 쓰지 않았다 — 대신 HEAD 소스에서 `git archive HEAD`로 빌드한 「이전」 바이너리와 작업 트리에서 빌드한 「이후」 바이너리를 같은 코퍼스에 각각 돌렸다.

```
# 이전(HEAD b7ceffa69 소스 빌드), 파이프 없음, rc=0
/tmp/t518-head/moai-head spec lint   →  0 error(s), 4378 warning(s)
# 이후(작업 트리 빌드), 파이프 없음, rc=0
/tmp/t518-after-moai spec lint       →  0 error(s), 4394 warning(s)
```

증거: `.moai/reports/t518/lint-before-MA1.txt` · `.moai/reports/t518/lint-after-MA1.txt`.

| 코드 | 이전(HEAD 빌드) | 이후 | 증감 |
|---|---|---|---|
| CoverageIncomplete | 3621 | 3637 | **+16** |
| ModalityMalformed | 412 | 412 | 0 |
| InvalidREQID | 6 | 6 | 0 |
| DuplicateREQID | 0 | 0 | 0 |
| (그 외 9개 코드 전부) | — | — | 0 |
| **error 등급 합** | **0** | **0** | **0** |

**증감 16행의 출처 귀속(REQ-SLB-013 필드 기준): 16행 전부 표 수집 유래, t385 구분자 넓힘 유래 0행.** 목록 경로의 findings는 한 건도 움직이지 않았다(AC-SLB-003이 코퍼스 전체에서 이를 단언한다).

**16행이 어느 문서에서 왔는지 행 단위로**: `SPEC-CODEX-PARTIAL-WIRING-001` 11행(body 81-91) · `SPEC-INIT-001` 5행(body 107-111). 제거된 행은 0.

**수집 20 vs 발화 16의 차 4를 남겨 두지 않는다.** 코퍼스 전체에서 C-d가 수집한 표 항목은 **20**이고(`TestTableCollection_CorpusTableEntryCensus`: 표 행 807개 관측 · 20 수집 · 787 기각 · 수집이 생긴 SPEC 3개 · 20 전부가 수집기-blind SPEC 안), 그중 4항목은 `SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001`의 것으로 **AC가 이미 참조하고 있어 `CoverageIncomplete`가 발화하지 않고, 본문도 GEARS 적합이라 `ModalityMalformed`도 발화하지 않는다.** 즉 20 − 4 = 16이며 미귀속 잔여는 0이다. 이 4행은 「수집이 소음만 만들지 않는다」의 실물이기도 하다 — 새로 보이게 된 요구사항 중 일부는 이미 건강하다.

**독립 유도 일치**: 판독기(python, `.moai/reports/t518/blind-axes-reader.py`)가 blind 53개 안에서 낸 C-d 잔존 **20**과, 출하 Go 코드가 코퍼스 전체에서 낸 수집 **20**이 일치한다. 두 구현은 서로 독립이며, 이 일치가 C-d를 고른 기준(독립 유도 사이의 재현성)이 코드에서도 유지됨을 보인다.

**[HARD] `CoverageIncomplete` 행의 해석은 t528에 유보된다.** +16 전부가 이 코드이고, AC 수집기(`parser.go:67,218`)가 이 코퍼스의 AC 문법을 읽지 못하므로 「진짜 미참조」와 「AC를 읽지 못함」의 분리는 t528 착지 후에만 가능하다. 이 표의 +16을 부채 크기로 읽지 않는다.

**설치본 바이너리와의 차 2건도 적는다.** 운영자가 준 `baseline-merged.txt`(설치본 `moai-adk v3.2.0-rc.0`)는 4380, 같은 트리 HEAD 소스 빌드는 4378이다. 차이는 `MovingRefUnpinned` 117 vs 115 한 코드에 몰려 있고, **설치본이 HEAD와 같은 소스가 아니라는 뜻**이다. 그래서 이 마일스톤은 설치본 수치를 baseline으로 쓰지 않았다 — 썼다면 +16이 +14로 잘못 보고됐을 것이다.

#### 검증 명령 (전부 파이프 없이 rc 판독, 범위는 `internal/spec`)

```
go test ./internal/spec/...            → rc=0, ok  github.com/modu-ai/moai-adk/internal/spec  69.918s
go vet ./internal/spec/...             → rc=0, 무출력
gofmt -l internal/spec/                → rc=0, 무출력
golangci-lint run ./internal/spec/...  → rc=0, "0 issues."
```

증거: `.moai/reports/t518/gotest-MA1.txt`. **전체 스위트는 로컬에서 돌리지 않았다**(CLAUDE.local.md §4 — 병렬 레인 부하 사고). 전 패키지 판정은 CI 몫이며 이 레인은 push하지 않는다.

#### 잔여 위험 (관측된 것만)

- **순서 의존 `DuplicateREQID`**: 한 문서가 같은 REQ ID를 표 행과 목록 정의로 함께 갖고 **표 행이 앞서면**, `REQIDUniquenessRule`이 뒤의 목록 항목을 중복으로 보고해 **비자문 finding이 새로 생긴다**(REQ-SLB-003 위반). 오늘 코퍼스에 그런 문서는 없다(양쪽 `DuplicateREQID` 0 실측). 방어 기전을 넣지 않은 것은 의도이며, AC-SLB-003의 코퍼스 단언이 그것이 생기는 날 잡는다.
- **본문 셀 선택의 대가**: 정의 표가 뒤에 노트 열을 덧붙이면 본문 대신 노트가 `Text`가 된다. 이 트리에서 그런 모양은 관측되지 않았고, 항목은 어차피 자문이다.
- **L1의 선택된 놓침은 그대로다**: `SPEC-INIT-001`의 `| REQ-N-001 | … 않아야 한다 |`는 여전히 기각된다. 결재된 대가이며, **어휘를 넓혀 고치지 않는다.** 다만 이 놓침이 REQ-SLB-005(기각의 관측 가능성)로 이름을 갖게 되는 것은 M-A1 범위 밖이므로, **현재는 여전히 조용한 미수집**이다 — 이 카드가 고치려는 결함이 그 행에 대해서는 아직 남아 있다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
