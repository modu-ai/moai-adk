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

### run-phase 중 계획 산출물 개정 (v0.6.0, 2026-09-08)

- **성격**: 이 회차는 감사 부채 상환도 운영자 결재의 착지도 아니다. **run-phase 실행 중에 발견된 계획 산출물(acceptance.md)의 결함을 리드 승인 아래 고친 개정**이며, 소관상 manager-spec이 수행했다. 범위는 `acceptance.md §C`의 AC-SLB-012 한 곳 + `spec.md`의 HISTORY 행 + 이 절뿐이다. 코드·다른 AC·다른 절은 손대지 않았다
- **[HARD] 소관 경계가 실제로 지켜졌다 — 이것이 기록의 요점이다.** 결함을 발견한 것은 M-A2를 수행한 run-phase 에이전트이고, 그 에이전트는 **계획 산출물이 자기 소관 밖이므로 스스로 고치지 않았다**(`§E.2`의 M-A2 절에 「acceptance.md는 수정하지 않았다(run-phase 소관 밖) — 이 발견은 후속 결재 사항으로 남긴다」로 기록돼 있다). 발견 → 상신 → 리드 승인 → manager-spec 개정의 경로가 그대로 돌았다
- **고친 내용**: AC-SLB-012가 축자 고정한 「선행 공백 없는 형」 픽스처(`REQ-FX-011`)는 본문에 `the system shall`이 이미 있어 옛 접촉 조건(`strings.Contains(upper, " SHALL")`)에서도 적합으로 판정된다. 따라서 수리를 되돌려도 그 픽스처는 0건에 머물고 **AC가 기대한 뮤턴트 RED가 서지 않는다.** 축자 원문은 **지우지 않고** 정정을 나란히 붙였으며(이 카드의 확립된 처리 방식 — `spec.md:316`의 L1/L2 개수 정정, 31·90 철회와 같은 형태), 실제로 가드를 만드는 교정 픽스처 `REQ-FXB-011`과 그것을 기계로 단언하는 `oldShallContact` 헬퍼(`TestModality_ShallContactIsWordBoundary`)를 가리키게 했다
- **네 번째다.** 「매치하지 않는 대조군은 대조군이 아니다」가 이 카드에서 네 번째로 나온 자리이며, 회차 수를 숫자로 남겼다(§A 규칙 8이 세운 논리 — 같은 모양이 반복되면 개별 정정으로는 다음 회차를 막지 못한다)
- **결합 관계를 함께 적었다**: 단어 경계 수리는 AC-SLB-007에 인접한 별개 항목이 아니라 그 수치의 전제다. 옛 조건에서는 `(SHALL)` 꼴이 보이지 않아 `…해야 한다(SHALL)` 형태의 한국어 요구사항이 전부 「판정 불가」로 오집계되고, 그러면 이 마일스톤의 산출물인 「정확한 판정 불가 건수」가 처음부터 틀린 수가 된다(M-A2 실측)
- **§A 규칙 8의 세 수치는 움직이지 않았다 — 재유도로 확인했다.** AC 식별자 **16** · REQ **14** · lint 경고 **14**. 이 개정은 `### AC-SLB-` 제목을 새로 만들지 않았고 REQ도 늘리지 않았으므로 이동이 없다. 규칙 8이 요구하는 대로 「움직이지 않았다」도 잰 결과로 적는다
- **do-not-disturb 무이동 확인**: `BLIND-AXES-DISCRIMINATOR-GATE` CLOSED와 C-d 선택 근거 미변경 · `SPEC-INIT-001` 선택된 대가와 「어휘를 넓혀 고치지 않는다」 [HARD] 미변경 · 20/61 재유도와 31·90 철회 기록 미변경 · 각 AC의 뮤턴트 기록 미변경(AC-SLB-012의 기록은 지우지 않고 정정을 덧붙였다)

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

### M-A2 — 축 2 갈래 B: 무판정 발화 + SHALL 접촉 조건 단어 경계화 (카드 t518, cycle_type=tdd)

착수 시점 트리: 워크트리 `.claude/worktrees/t518`, 브랜치 `WT-spec-lint-axes`, HEAD `6cfcfef00`(M-A1 착지분). **갈래 A(한국어 modality 판정)는 손대지 않았다** — 결재로 후속 카드에 넘어갔고, 이 마일스톤은 한국어 어휘를 한 글자도 세우지 않는다.

**변경한 파일 1개**: `internal/spec/lint.go` — ① `isModalityMalformed`를 3상태 `judgeModality`로 대체(conforming / malformed / **unjudged**) ② SHALL 접촉 조건을 `strings.Contains(upper, " SHALL")` → `\bSHALL\b` 단어 경계 매치로 교체(REQ-SLB-014) ③ `EARSModalityRule.Check`에 `ModalityUnjudged` 발화 지점 추가. 테스트 `internal/spec/lint_modality_unjudged_test.go`(신규). `lint_req_table.go`·`lint_req_widen.go`는 손대지 않았고(M-A1 소관), `internal/cli`도 손대지 않았다(형제 SPEC 소관).

**확정된 무판정 코드 이름: `ModalityUnjudged`** (AC-SLB-008 주석이 M3 산출물에 기록하라고 요구한 값). 등급은 warning + `Advisory: true`, 발화 지점에서 코드 단위로 정한다.

**판정 가능성의 두 경로 — 어느 쪽도 한국어 어휘가 아니다.** ① 영어 modality 접두사 다섯 개 중 하나로 시작 → 기존 malformed/conforming 판정을 그대로 탄다. ② 본문 어디든 **단어 경계 SHALL 토큰**이 있음 → conforming. 둘 다 아니면 `ModalityUnjudged`. 경로 ②가 한국어 요구사항을 판정 가능하게 만드는 유일한 이유이며, 그것이 판정하는 것은 한국어가 아니라 **기존 어휘가 이미 알고 있던 SHALL 토큰**이다 — 코퍼스가 `…해야 한다(SHALL)` 꼴로 괄호 안에 그 토큰을 이미 쓰고 있다. 옛 접촉 조건(선행 공백 요구)으로는 `(SHALL)`이 보이지 않아 그런 요구사항이 전부 「판정 불가」로 오집계되므로, **REQ-SLB-014의 교체가 없으면 이 마일스톤의 산출물인 「정확한 판정 불가 건수」가 처음부터 틀린 수가 된다.** 이것이 접촉 조건 교체가 갈래 B 안쪽인 실측된 이유다(SPEC §I가 가설로 적어 둔 것을 여기서 실측으로 바꾼다).

#### AC 판정표

| AC | REQ | 판정 | 실측 근거 |
|---|---|---|---|
| AC-SLB-006a | 003 · 006 | **PASS** | `TestModality_EnglishControlStillJudged` — 영어 SHALL 누락 요구사항에서 `ModalityMalformed` 1건 · `ModalityUnjudged` 0건(회귀 없음) |
| AC-SLB-006b | 006 | **PASS** | `TestModality_KoreanControlAnnounced` — 한국어 동일 결함에서 두 코드 중 **정확히 하나**(Unjudged 1 · Malformed 0). 「둘 다 0」은 `t.Fatalf`로 분리해 FAIL 사유를 구분했다. 영어/한국어 코드 집합이 실제로 갈리는 것을 **단언으로** 확인 |
| AC-SLB-007 | 006 · 007 | **PASS** | `TestModality_SilenceDiffersFromConformance` — acceptance.md §C 축자 픽스처 2종. 적합(`…해야 한다(SHALL).`) Unjudged 0 · Malformed 0, 판정 불가(`…구현 재량에 맡긴다.`) Unjudged 1, 두 finding 집합이 다름 |
| AC-SLB-008 | 007 | **PASS** | `TestModality_UnjudgedIsAdvisory` — severity warning + `Advisory: true`, error 등급 0건, `--strict`에서 `HasErrors()` false. **픽스처가 narrow 목록형이라 `Widened=false`임을 먼저 단언한다** — 자문 표시가 t385 넓힘 경로에서 온 것이 아님을 배제하기 위해서다 |
| AC-SLB-008b | 008 | **PASS** | `TestModality_UnjudgedIsNotReportedConforming` — `judgeModality`의 반환값을 직접 읽어 판정 불가 본문이 `modalityJudgedConforming`이 **아님**을 단언한다. 부재(=finding 0건)로 추론하지 않는다 — 부재가 곧 결함의 모양이기 때문이다 |
| AC-SLB-012 | 014 | **PASS(단, 축자 픽스처 1종 결함 발견 — 아래)** | `TestModality_ShallContactIsWordBoundary` — 세 픽스처 전부 Malformed 0 · Unjudged 0. 가드를 만드는 것은 **교정 픽스처 1종뿐**이며, 그 사실이 `oldShallContact` 헬퍼로 기계 단언돼 있다. 수치 공표 갈래는 아래 「코퍼스 이동」 |
| AC-SLB-010 | 002 · 010 | **PASS(완결)** | M-A1의 「부분 PASS」 남은 절반. `TestModality_AdvisoryDoesNotLeakIntoErrors` — 표 유래(Widened) + 무판정(코드 단위 자문) 두 자문 경로를 한 픽스처에서 동시에 발화시키고 error 등급 0 · strict `HasErrors()` false를 단언. 픽스처가 실제로 양쪽을 갖는지(table≥1 AND list≥1)를 먼저 단언한다 |

**이 마일스톤이 잡지 않은 AC**: 005(기각의 관측 가능성 — 판단은 아래 별도 절) · 009(재계수 = M4). 001a/001b/001c/002/003/004/011은 M-A1이 이미 잡았다.

#### [HARD] acceptance.md §C의 AC-SLB-012 축자 픽스처 한 종은 가드를 만들 수 없다 — 기계로 확인했다

acceptance.md가 「선행 공백 없는 형」으로 축자 고정한 줄은

```
- **REQ-FX-011** — When the tool runs, the system shall report.(SHALL)
```

인데, 이 본문은 `the system shall`에서 이미 **선행 공백이 붙은 SHALL을 갖는다**. 즉 옛 접촉 조건(`strings.Contains(upper, " SHALL")`)도 이 줄을 적합으로 판정하므로, 뮤턴트를 되돌려도 이 픽스처는 계속 0건이고 **RED를 만들지 못한다**. acceptance.md가 이 픽스처에 기대한 뮤턴트 가드(「선행 공백 없는 형에서 ModalityMalformed가 1건이 되어야 한다」)는 쓰인 그대로는 성립하지 않는다.

이것은 이 카드가 두 번 경고한 「매치하지 않는 대조군은 대조군이 아니다」와 같은 모양이므로, 산문으로 적지 않고 **테스트가 기계로 단언한다**: `oldShallContact(verbatim011)`가 참임을 `TestModality_ShallContactIsWordBoundary`가 확인한다. 축자 픽스처는 **지우지 않고 그대로 남겼고**(SPEC에 대한 증거이므로), 가드를 실제로 만드는 교정 픽스처를 하나 더 세웠다:

```
- **REQ-FXB-011** — When the tool runs, the tool reports.(SHALL)
```

이 본문은 SHALL 출현이 괄호 안 하나뿐이라 옛 조건은 놓치고 단어 경계 조건은 본다. 뮤턴트 M2가 이 픽스처에서만 RED를 만든다. **acceptance.md는 수정하지 않았다**(run-phase 소관 밖) — 이 발견은 후속 결재 사항으로 남긴다.

#### 뮤턴트 9건 + 재실행 1건 — 잡히지 않은 것도 적는다

증거: `.moai/reports/t518/mutants-MA2.txt`(1차) · `.moai/reports/t518/mutants-MA2-m8-rerun.txt`(M8 재실행), 재현: `python3 .moai/reports/t518/mutants-MA2.py`. **`-run` 필터 없이 `internal/spec` 패키지 전체를 돌린다** — 「이 마일스톤의 가드가 잡았다」와 「기존 테스트가 잡았다」는 다른 사실이고, 필터를 걸면 그 구분이 보이지 않는다.

| 뮤턴트 | 결과 | RED가 된 테스트 |
|---|---|---|
| M1 무판정 발화 제거 | CAUGHT | 006b · 007 · 008 · 008b · 010 |
| M2 접촉 조건을 선행 공백형으로 되돌림 | CAUGHT | 012 · 007 · 008b · census · **M-A1의 `TestTableCollection_CorpusListFindingsUnchanged`** |
| M3 무판정 판정을 conforming으로 흡수(3번째 상태 소거) | CAUGHT | 006b · 007 · 008 · 008b · 010 · census |
| M4 무판정 finding의 `Advisory` 제거 | CAUGHT | 008 · 010 · M-A1 코퍼스 단언 |
| M5 무판정 finding을 error 등급으로 승격 | CAUGHT | 008 · 010 |
| M6 [경계] 단어 경계를 맨 부분문자열로 넓힘(SHALLOW가 SHALL로 셈) | CAUGHT | 012 |
| M7 [경계] SHALL 토큰 경로 제거(접두사만 판정 가능) | CAUGHT | 007 · 008b · M-A1 코퍼스 단언 |
| M8 [경계] 영어 접두사 하나 삭제(`"THE "`) | **1차 NOT CAUGHT → 가드 신설 후 CAUGHT** | (1차: 없음) → `TestModality_EveryEnglishPrefixIsStillJudged` |
| M9 [경계] 접두사 매칭을 원문(대소문자 보존)에 걸기 | CAUGHT | 006a · 006b · 012 · 008b · M-A1 `TestTableCollection_AdvisoryDoesNotGate` |

**[HARD] M8은 1차에서 패키지 전체를 초록으로 통과했다 — 지우지 않고 적는다.** 이 마일스톤의 픽스처가 전부 `When`으로 시작했고 기존 EARS 테스트도 Ubiquitous 접두사 삭제를 잡지 못했다. 삭제된 채였다면 `The system does X` 꼴 요구사항 전부가 malformed에서 **조용히 unjudged로 재분류**됐을 것이다 — 이 카드가 고치는 결함과 정확히 같은 모양의, 코드가 아니라 **계측의** 사각이다. 다섯 접두사 각각에 SHALL 없는 본문을 걸고 malformed를 요구하는 `TestModality_EveryEnglishPrefixIsStillJudged`를 신설해 CAUGHT로 바꿨다(재실행 exit=1, RED 1건). **기록을 남기는 이유**: 지금 CAUGHT인 것은 가드를 나중에 붙였기 때문이지 처음부터 있었기 때문이 아니다.

M2·M4·M7·M9가 M-A1의 코퍼스 단언을 RED로 만든 것도 적어 둔다 — M-A1이 「변별력이 7건과 순서 의존 위험에 한정된다」고 유보를 달았던 그 단언이, 축 2 변경에 대해서는 실제로 물성을 가졌다.

#### 코퍼스 이동 — 옛 값과 새 값을 나란히 (AC-SLB-012 둘째 갈래 · REQ-SLB-014)

**[HARD] 두 값 모두 같은 트리·같은 모집단에서 재유도했다. 설치본 바이너리는 쓰지 않았다** — M-A1이 실측한 대로 설치본(`v3.1.2-1490-ge79c010b8`)은 이 트리와 다른 소스이고, 그것으로 잰 baseline은 `MovingRefUnpinned`에서 어긋난다. 「이전」은 `git archive HEAD`로 뽑은 소스에서 빌드했고, 「이후」는 작업 트리에서 빌드했다.

```
# 이전(HEAD 6cfcfef00 소스 빌드), 파이프 없음, rc=0
/tmp/t518-MA2-head/moai-head spec lint   →  0 error(s), 4394 warning(s)
# 이후(작업 트리 빌드), 파이프 없음, rc=0
/tmp/t518-MA2-after-moai spec lint       →  0 error(s), 4617 warning(s)
```

증거: `.moai/reports/t518/lint-before-MA2.txt` · `.moai/reports/t518/lint-after-MA2.txt`. 「이전」 4394는 M-A1이 기록한 착지 후 값과 일치한다(저장된 숫자를 재사용한 것이 아니라, 같은 방법으로 다시 재서 같은 값이 나왔다).

| 코드 | 이전(옛 값) | 이후(새 값) | 증감 |
|---|---|---|---|
| `ModalityUnjudged` | **0**(코드 자체가 없었음) | **460** | **+460** |
| `ModalityMalformed` | **412** | **175** | **−237** |
| CoverageIncomplete | 3637 | 3637 | 0 |
| MovingRefUnpinned | 115 | 115 | 0 |
| StatusTransitionInvalid | 102 | 102 | 0 |
| LegacyEARSKeyword | 48 | 48 | 0 |
| MissingExclusions | 26 | 26 | 0 |
| StatusGitConsistency | 18 | 18 | 0 |
| FrontmatterInvalid | 14 | 14 | 0 |
| StatusTokenUnrecognized | 7 | 7 | 0 |
| SyncSHASlotFormat | 6 | 6 | 0 |
| InvalidREQID | 6 | 6 | 0 |
| SpecsDirMissingSpecFile | 2 | 2 | 0 |
| OwnershipTransitionInvalid | 1 | 1 | 0 |
| **합계** | **4394** | **4617** | **+223** |
| **error 등급 합** | **0** | **0** | **0** |

−237 + 460 = +223. 잔여 미귀속 0.

**−237의 귀속 — 전부 접촉 조건 교체다.** 두 출력의 `ModalityMalformed` finding 집합을 (파일, 행) 키로 대조하면 **사라진 것 237건, 새로 생긴 것 0건**이다. 사라진 237건 전부가 「단어 경계 SHALL 있음 AND 선행 공백 SHALL 없음」을 만족한다(`\bSHALL\b` 매치 참 · `" SHALL"` 매치 거짓, 237/237). 접두사를 가진 항목은 절대 unjudged가 될 수 없으므로(판정 경로 ①), 이 −237이 무판정 코드 신설로 흘러간 것이 아님도 구조적으로 성립한다.

**SHALL 바로 앞 문자의 분포(237건)**: `*` 231건 · `(` 3건 · `` ` `` 3건. 즉 실제 코퍼스에서 옛 조건이 놓치던 지배적 모양은 SPEC이 예상한 `(SHALL)` 괄호형이 **아니라 마크다운 볼드 표기**였다. 괄호형은 3건뿐이다. 예상과 실측이 다르므로 실측을 적는다.

**반대 방향(교체가 새로 켜는 쪽)은 이 코퍼스에서 0이다.** 옛 조건은 `" SHALLOW"`를 SHALL로 인정했고 새 조건은 인정하지 않는다 — 그런 항목이 있었다면 `ModalityMalformed`가 늘었어야 하는데 새로 생긴 것은 0건이다. 즉 **이 코퍼스에는 SHALLOW형 오인이 없었다**. 이 방향은 코퍼스가 아니라 단위 테스트(M6 뮤턴트)가 지킨다.

**+460의 귀속**: 460건 전부 새로 드러난 것이다(접두사가 없으므로 옛 코드에서는 `false`를 돌려주고 아무 finding도 내지 않았다). 출처 필드(REQ-SLB-013) 기준 분해는 **표 수집 유래 5건 · 목록 유래 455건**이며, 표 유래 5건은 전부 `SPEC-INIT-001`이다. 78개 SPEC 디렉터리에 걸쳐 있고 상위는 `SPEC-HARNESS-CLI-COVERAGE-001` 22 · `SPEC-HARNESS-EVOLVE-003` 22 · `SPEC-CODEX-BODY-NEUTRALITY-001` 16이다.

**독립 유도 일치**: 두 바이너리 출력의 차분(−237 / +460)과, 출하 Go 코드가 코퍼스를 직접 순회해 낸 census(`TestModality_CorpusUnjudgedCensus`: `REQ entries=3985 | unjudged=460 (78 SPEC dirs; 5 table-sourced, 455 list-sourced) | malformed=175 | conforming=3350 | recovered by word boundary=237`)가 두 수치 모두에서 일치한다. 서로 다른 두 경로(출력 파싱 / 코드 직독)가 같은 값을 낸다.

**이 마일스톤의 산출물 — 판정 불가 요구사항의 정확한 건수는 460이다** (수집된 REQ 항목 3,985개 중, 78개 SPEC 디렉터리에 분포). 이 수가 갈래 A 후속 카드의 크기를 정한다. **저장하지 않는다** — census 테스트가 매 실행 재유도하며, 위 수치는 이 시점의 재유도값이지 코퍼스의 성질이 아니다.

**종료 코드 계약은 움직이지 않았다(REQ-SLB-010)**: `spec lint` 기본 모드는 이전·이후 모두 rc=0, `--strict`는 이전·이후 모두 rc=1이며 두 실행 모두 ERROR 등급 출력 행이 0이다. `ModalityUnjudged`가 자문이므로 strict 승격 대상이 아니고, strict의 rc=1은 이 변경 **이전부터** 비자문 warning들이 만들고 있던 상태다.

#### AC-SLB-005(기각의 관측 가능성)의 소관 — 이 마일스톤이 아니라 별도 마일스톤으로 미룬다

M-A1이 「전용 finding 코드와 발화 지점이 필요하며 수집기 범위 밖」이라며 남긴 항목이다. 이 마일스톤은 발화 지점을 만드는 일을 했으므로 후보였고, **미루기로 판단했다.** 근거 셋:

1. **소관이 축 1이다.** AC-SLB-005는 acceptance.md **§B(축 1)**에 있고 그 Given이 AC-SLB-004의 기각 픽스처다. 기각을 관측 가능하게 하려면 기각한 행을 기록하는 쪽, 즉 `lint_req_table.go`의 판별식(`isTableDefinitionRow`)이 바뀌어야 한다 — 축 2 변경이 필요로 하는 범위가 아니다.
2. **한 측정에 두 원인이 섞인다.** 코퍼스의 기각 행은 787건이다(M-A1 census). 같은 회차에 발화하면 +460과 +787이 한 재측정에 겹쳐 들어오고, REQ-SLB-009가 요구하는 **출처별 귀속**이 코드별 귀속으로만 남는다. 이 카드가 두 번 데인 자리가 정확히 「총량은 나오되 귀속이 나오지 않는」 상태다.
3. **결정이 하나 남아 있다.** 787건을 REQ 단위로 발화할지 표 단위로 접을지는 설계 결정이고, 그 선택이 코퍼스 수치를 세 자릿수로 움직인다. 축 2 착지와 묶으면 두 결정이 한 diff에서 리뷰된다.

**소관 지정: M-A2b(축 1 — 기각의 관측 가능성), M4(재계수) 이전에 실행.** M4보다 뒤로 가면 재계수가 끝난 뒤 코퍼스가 다시 움직여 재계수를 다시 해야 한다. 그때까지 `SPEC-INIT-001`의 `| REQ-N-001 | … 않아야 한다 |` 행은 **여전히 조용한 미수집**이며, M-A1이 적은 그 유보는 이 마일스톤에서도 해소되지 않았다.

#### 검증 명령 (전부 파이프 없이 rc 판독, 범위는 `internal/spec`)

```
go test ./internal/spec/...            → rc=0, ok  github.com/modu-ai/moai-adk/internal/spec  70.786s
go vet ./internal/spec/...             → rc=0, 무출력
gofmt -l internal/spec/                → rc=0, 무출력
golangci-lint run ./internal/spec/...  → rc=0, "0 issues."
```

증거: `.moai/reports/t518/gotest-MA2.txt`. **전체 스위트는 로컬에서 돌리지 않았다**(CLAUDE.local.md §4). 전 패키지 판정은 CI 몫이며 이 레인은 push하지 않는다.

뮤턴트 복원 후 작업 트리에서 바이너리를 다시 빌드해 코퍼스를 재측정했고, 코드별 분포가 기록된 「이후」와 **완전히 동일**함을 확인했다(4617, 13개 코드 전부 일치) — 뮤턴트 실행이 소스를 남기지 않았다는 증거다.

#### RED 증거

`.moai/reports/t518/red-MA2.txt`. 2단계로 잡았다: ① 신규 심볼 미정의로 인한 빌드 실패(약한 RED) ② 옛 의미를 새 모양에 담은 스텁(`judgeModality`가 conforming/malformed 2상태만)을 넣어 **단언 수준 RED**를 다시 채취. 기록된 것은 ②이며, 6개 테스트가 「무판정 0건」·「두 픽스처의 finding 집합이 같음」·「단어 경계가 괄호형 SHALL을 인정하지 않음」으로 각각 실패했다. `TestModality_EnglishControlStillJudged`는 이 단계에서 이미 GREEN이다 — 회귀 없음을 재는 대조군이므로 옳다.

#### 잔여 위험 (관측된 것만)

- **`ModalityUnjudged` 460건은 자문이지만 소음이다.** `spec lint` 기본 출력이 4394 → 4617로 늘었다(+5.1%). 등급상 게이트를 흔들지 않고, 그 가시성이 이 카드의 성과이기도 하지만, 갈래 A가 착지하기 전까지는 460건이 계속 출력에 남는다. 이 대가는 결재된 것이다(spec.md §I 「받아들인 대가」).
- **판정 경로 ②의 사각**: 접두사가 없고 SHALL이 있는 **영어** 문장(`System shall X`)은 conforming으로 처리된다 — 옛 코드와 동일한 처리이며 이 카드가 넓히지 않았다. 코퍼스에서 이 모양이 몇 건인지는 재지 않았다(미측정, 갈래 A 소관).
- **`(SHALL)` 인정의 부작용은 재지 않았다.** 단어 경계 조건은 `SHALL NOT`·`shall not`도 SHALL 존재로 인정한다(옛 조건도 그랬다 — 선행 공백이 있으므로). 즉 Unwanted 요구사항의 부정형을 적합으로 읽는 성질은 이 교체가 만든 것이 **아니라 원래 있던 것**이며, 이 카드는 그것을 바꾸지 않았다.
- **acceptance.md §C의 축자 픽스처 1종이 가드를 만들지 못한다**(위 절). 코드가 아니라 기준 문서의 결함이므로 run-phase에서 고치지 않았다.

### M-A2b — 축 1: 기각의 관측 가능성 (카드 t518, cycle_type=tdd)

**소관**: AC-SLB-005(REQ-SLB-005). M-A2가 「M4(재계수) 이전에 실행」으로 지정해 남긴 마일스톤이며, 그 지정 근거 셋(소관이 축 1 · 두 원인이 한 측정에 섞임 · 설계 결정 하나가 남아 있음)이 그대로 이 마일스톤의 전제다.

**결재된 설계 — 표 단위 접기 + 줄에 N.** 기각된 표 하나당 자문 finding 한 줄을 내고, 그 줄이 해당 표의 기각 행 수 N을 싣는다. 787줄을 행 단위로 내는 안은 **기각**됐다(M-A2가 이미 기본 출력을 +5.1% 늘렸고, 여기에 787을 더하면 경고를 아무도 읽지 않는 구간에 들어가 결재된 대가를 초과한다). 행 단위 `--verbose` 확장도 **검토 후 기각** — 발화 경로가 둘이 되면 반경이 넓어지고 확장 경로 수치를 따로 검증해야 한다. acceptance.md AC-SLB-005에 세 판정 기준으로 전사했다(v0.7.0, 리드 승인).

**[HARD] 접기가 바꾼 관측 단위를 새로 적는다.** 행 → 표. 지금 보이는 것은 **어느 표가 기각됐는지**이고, **그 안의 어느 행인지는 N이라는 수로만** 보인다. v0.6.0까지의 AC 문장이 시사하던 것과 같지 않으므로 같다고 읽히게 두지 않는다.

#### 산출물

`internal/spec/lint_req_table_rejection.go`(신규) — `collectRejectedTables`(인접 표 줄의 최대 연속 구간을 한 표로 묶고 그 안의 기각 행만 셈) · `REQTableRejectionRule`(코드 `REQTableRowsRejected`, warning + advisory) · `parseRejectedRowCount`(렌더된 메시지에서 N을 되읽음). `internal/spec/lint.go`에 규칙 등록. 테스트 `internal/spec/lint_req_table_rejection_test.go`(9건).

**N을 줄에 싣는 이유 — 계기가 하나뿐이면 조용해져도 모른다.** N이 없으면 787의 유일한 증거는 census 테스트 하나다. N이 실리면 **출력 자체가 두 번째 계기**가 되고, 두 계기의 불일치가 침묵이 아니라 RED로 나타난다. N은 struct 필드가 아니라 **렌더된 메시지**에서 되읽는다 — 운영자가 실제로 보는 표면을 재는 것이 대조의 요점이기 때문이다.

#### AC 판정

| AC | 판정 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-SLB-005 기준 1(표 하나당 한 줄) | PASS | `go test -run 'TestTableRejection_OneLinePerRejectedTable\|TestTableRejection_MultipleRowsFoldIntoOneFinding' ./internal/spec/` | 두 테스트 PASS. 3행 1표 픽스처 → finding 1건 |
| AC-SLB-005 기준 2(줄이 기각 행 수 N을 실음) | PASS | `go test -run TestTableRejection_MixedTableCountsOnlyRejectedRows ./internal/spec/` | PASS. 수집 1 · 기각 2인 혼합 표에서 N=2 |
| AC-SLB-005 기준 3(모든 N의 합 = census) | PASS | `go test -run TestTableRejection_CorpusSumEqualsRowCensus -v ./internal/spec/` | `corpus rejection: rows=787 emitted lines=81 sum(N)=787` |
| AC-SLB-005 대조쌍 | PASS | `go test -run TestTableRejection_ControlPairDiverges ./internal/spec/` | PASS. 수집 표 0건 / 기각 표 2건으로 갈림 |
| AC-SLB-005 뮤턴트 가드 | PASS | 아래 뮤턴트 표 M1 | 6개 테스트 RED, AC-SLB-004는 GREEN 유지(비대칭 성립) |
| AC-SLB-004(기각 자체) 회귀 없음 | PASS | `go test -run TestTableCollection ./internal/spec/` | 전건 PASS |

**AC-SLB-005는 이로써 닫혔다.** 남은 AC 중 이 마일스톤 소관은 없다. M4(재계수, AC-SLB-009)와 갈래 A 후속은 미착수다.

#### 뮤턴트 — 잡히지 않은 것도 남긴다

| # | 뮤턴트 | 1차 결과 | 비고 |
|---|---|---|---|
| M1 | 발화 제거(`Check`가 nil 반환) | **잡힘** — 6건 RED | `TestTableCollection_DiscriminatorRejectsDispositionTables`(AC-SLB-004)와 census는 **GREEN 유지**. AC-SLB-005가 예고한 비대칭이 실측으로 성립 |
| M2 | N이 기각 행이 아니라 표의 전체 REQ 행을 셈 | **잡힘** — 3건 RED | 코퍼스 합 대조가 함께 잡았다(부풀린 N을 검출) |
| M3 | 접기 제거(행 단위 발화) | **잡힘** — 3건 RED | `OneLinePerRejectedTable`은 **PASS로 남았다** — 그 픽스처는 표마다 기각 행이 1개라 행 단위와 표 단위가 구별되지 않는다. 그 가드의 경계이므로 적어 둔다 |
| M4 | 끝맺음 `flush()` 제거(본문이 표 줄로 끝나는 경우 마지막 표 유실) | **잡히지 않음(1차)** | 모든 픽스처와 코퍼스의 모든 spec.md가 개행으로 끝나 `strings.Split`이 빈 줄을 만들고, 그것이 부수효과로 flush를 일으켰다. 즉 이 줄은 **어떤 테스트도 도달하지 못하는 살아 있는 코드**였다 |
| M4 재실행 | 같은 뮤턴트, 가드 추가 후 | **잡힘** — 1건 RED | 가드 신설: `TestTableRejection_TableRunningToEndOfBody`(개행 없이 표 줄로 끝나는 본문). **1차 결과는 지우지 않는다** — 「나중에 가드를 붙여서 잡혔다」는 「가드가 있었다」와 다른 사실이다 |
| M5 | 메시지에서 N 제거(`%.0s%.0d`) | **잡힘** — 5건 RED | 렌더된 표면을 재는 설계가 실제로 그 표면을 보고 있음의 증거 |
| M6 | 규칙 미등록(`lint.go`에서 주석 처리) | **잡힘** — 1건 RED | `TestTableRejection_RuleIsRegistered`만 RED. 나머지는 `Check`를 직접 부르므로 GREEN — 「고립된 통과」와 「사용자에게 도달」의 구별 |

#### RED 증거 (구현 전, 축자)

```
$ go test ./internal/spec/ -run 'TestTableRejection' -count=1
# github.com/modu-ai/moai-adk/internal/spec [github.com/modu-ai/moai-adk/internal/spec.test]
internal/spec/lint_req_table_rejection_test.go:33:11: undefined: REQTableRejectionRule
internal/spec/lint_req_table_rejection_test.go:230:11: undefined: REQTableRejectionRule
internal/spec/lint_req_table_rejection_test.go:262:11: undefined: parseRejectedRowCount
FAIL	github.com/modu-ai/moai-adk/internal/spec [build failed]
FAIL
```

빌드 실패 수준의 RED다. 단언 수준 RED는 뮤턴트 M1이 사후에 같은 역할을 한다(6건이 「신호 0건」으로 실패) — M1의 출력이 곧 「발화가 없을 때 이 테스트들이 무엇을 말하는가」의 증거다.

#### 코퍼스 이동 — 세 바이너리를 같은 코퍼스 위에서 돌렸다

설치된 바이너리(`~/go/bin/moai`, `v3.1.2-1490-ge79c010b8`)는 **쓰지 않았다** — 이 트리보다 한참 뒤처져 있다. `git archive`로 두 시점을 꺼내 소스에서 빌드하고, 작업 트리도 빌드해 셋을 같은 코퍼스 위에서 파이프 없이 돌렸다.

| 바이너리 | 트리 | `moai spec lint` (인자 없음) | rc |
|---|---|---|---|
| A | `458fc7ebc` (M-A2 착지) | `0 error(s), 4617 warning(s)` | 0 |
| B | `2cfccd4eb` (HEAD, M-B1 포함) | `0 error(s), 4617 warning(s)` | 0 |
| C | 작업 트리 (M-A2b) | `0 error(s), 4698 warning(s)` | 0 |

**코드별 귀속(B → C)**: 코드 카운트 표를 diff 하면 **추가된 행이 정확히 하나** — `81 REQTableRowsRejected`. 다른 13개 코드는 한 건도 움직이지 않았다.

```
$ diff codes-B.txt codes-C.txt
5a6
>   81 REQTableRowsRejected
```

**출처별 귀속(REQ-SLB-009)**: 81건 **전부** 표 수집 축(축 1) 유래이며, t385 구분자 넓힘 유래는 **0건**이다. 이 코드는 `REQEntry`를 뒤에 두지 않으므로 `Source` 필드를 싣지 않는다 — 기각된 행은 애초에 항목을 만들지 않았고, 그 사실 자체가 보고 대상이기 때문이다. 귀속은 필드가 아니라 **발화 지점이 하나뿐이라는 구조**가 만든다(`REQTableRejectionRule`만이 이 코드를 낸다).

**출력에서 되읽은 합 — 두 번째 계기**:

```
$ grep -o 'rejected_rows=[0-9]*' out-C.txt | sed 's/rejected_rows=//' | awk '{s+=$1} END {print NR, s}'
81 787
```

Go census(직접 행 훑기)와 출하 바이너리 출력(메시지 파싱)이 **787로 일치**한다. 두 경로가 서로를 재고, 4698 − 4617 = 81도 발화 줄 수와 정확히 맞는다.

#### 귀속되지 않은 +4 — 코드 탓이 아님이 밝혀졌고, 잔여는 M-A3로 넘긴다

브리핑이 인용한 M-B1 무인자 대조군의 **4621**은 이 코퍼스에서 재현되지 않는다. A와 B가 **같은 코퍼스에서 정확히 같은 4617**을 낸다는 것이 측정이며, 이는 M-B1 코드 변경이 총량을 **0만큼** 움직였다는 뜻이다. 따라서 +4는 **코드 귀속이 아니다** — 남은 후보는 그 측정 시점의 코퍼스 상태이거나 다른 호출 형태이며, 둘 중 어느 쪽인지는 **재지 않았다(Gap)**. 추측으로 메우지 않고 M-A3(재계수) 몫으로 남긴다.

#### 검증 명령 (전부 파이프 없이 rc 판독, 범위는 `internal/spec`)

```
go test ./internal/spec/... -count=1   → rc=0, ok  github.com/modu-ai/moai-adk/internal/spec  128.261s, `--- FAIL` 0건
```

**전체 스위트는 로컬에서 돌리지 않았다**(CLAUDE.local.md §4). `internal/cli`는 건드리지 않았다(형제 SPEC 반경, M-B1로 이미 착지).

#### §A 규칙 8 세 수치 — 재유도, 움직이지 않음

```
grep -oE '^### AC-SLB-[0-9]+[a-z]?' acceptance.md | sort -u | wc -l  → 16
grep -cE '^\s*[-*]\s+\**\s*REQ-SLB-[0-9]+\s*\**\s*(\([^)]*\)\s*\**\s*)?(—|:)' spec.md → 14
moai spec lint .moai/specs/SPEC-SPEC-LINT-BLIND-AXES-001/spec.md → rc=0, `0 error(s), 14 warning(s)`
```

AC 식별자를 새로 만들지 않고 기존 AC-SLB-005를 구체화했으므로 16은 그대로다. 문서 개정 후 코퍼스를 다시 돌려도 4698 / 81 / 787로 불변임을 확인했다.

#### 잔여 위험 (관측된 것만)

- **표의 경계는 「연속한 표 줄」로 정의했다.** 빈 줄 하나가 사이에 들어간 논리적 한 표는 두 표로 세어진다. **합(787)은 이 선택에 영향받지 않고**(모든 기각 행은 정확히 한 구간 안에 들어간다) 줄 수 81만 영향받는다. 다른 정의를 골랐다면 81이 달라졌을 것이며, 이 수는 코퍼스의 성질이 아니라 이 정의 아래의 재유도값이다.
- **`SPEC-INIT-001`의 `| REQ-N-001 | … 않아야 한다 |`는 여전히 미수집이다.** 달라진 것은 그것이 **조용하지 않다**는 점뿐이다 — 이제 그 표가 기각됐다는 줄이 나온다. 어느 행인지는 N으로만 보이므로, 그 행을 특정하려면 여전히 문서를 열어야 한다. 이것이 결재된 대가의 현재 모양이다.
- **기본 출력이 4617 → 4698로 +1.8% 늘었다.** M-A2의 +5.1%에 얹힌 값이며 전부 자문이라 게이트를 흔들지 않는다. 접기가 없었다면 +17%였다.
- **코퍼스 합 대조는 같은 판별식을 양쪽에서 쓴다.** census와 규칙이 모두 `isTableDefinitionRow`를 부르므로, 판별식 **자체**가 틀렸다면 두 계기가 함께 틀린다. 이 대조가 재는 것은 「접기·발화·등록 경로가 행을 잃거나 만들지 않는가」이지 「판별식이 옳은가」가 아니다 — 후자는 AC-SLB-004와 축자 픽스처의 몫이다.


### M-A3 — 코퍼스 재계수 (카드 t518, AC-SLB-009)

이 마일스톤은 코드를 바꾸지 않는다. 재계수 자체가 산출물이며, 그 수가 이 카드가 무엇을 드러냈는지에 답한다.

#### 측정 창과 앵커

- **앵커**: `0705eb8d7` (origin/develop `fbb5cfd8d` 흡수 직후), 2026-09-07T18:40Z ~ 18:56Z.
- 창을 열기 전과 닫은 뒤 `git status --porcelain`을 각각 읽었고, 두 번 모두 미추적 1건(`.moai/reports/t518/ma3-merged-spec-tests.txt`)뿐이었다. 이 경로는 `.moai/reports/` 아래라 lint 코퍼스(`.moai/specs/`)에 들어가지 않는다. `git rev-parse HEAD`도 창 전후로 같은 값이었다.
- 측정 명령은 전부 전경에서 하나씩 돌렸다 — 배경 작업을 만들지 않았으므로 측정 중 트리에 쓰는 것이 없다.
- **원격은 이 앵커 이후 `bce6d7e08`까지 움직였다.** 재흡수는 하지 않았다 — 측정 도중 코퍼스를 갈아 끼우면 비교 대상 자체가 사라진다. 앵커가 낡은 것이 아니라 의도적으로 고정된 것임을 뒤에 읽는 사람이 알도록 적어 둔다.
- **이 기록 자체가 코퍼스 안에 있다 — 그래서 쓴 뒤 다시 쟀다.** 이 §E.2 M-A3 절을 `progress.md`에 쓴 직후 같은 바이너리로 재실행: `0 error(s), 4757 warning(s)`(rc=0). 총계는 움직이지 않았고, 이 SPEC 디렉터리의 기여도 `spec.md`의 `CoverageIncomplete` 14건 그대로다 — `progress.md`는 0건을 낸다. 본문이 `origin/develop`·`bce6d7e08`을 언급하지만 `MovingRefUnpinned`는 115에서 움직이지 않았다. 그 재실행 뒤에 이 절을 두 번 더 손봤으므로(이 줄과 증거 표) **커밋된 상태에서 세 번째로 다시 쟀다**: 역시 `0 error(s), 4757 warning(s)`, 15개 코드 카운트 전부 일치(`diff` 무출력). 한 번의 판독은 순간의 관측이라는 성질이 이 절 자신에게도 걸린다.

#### 계측기 — 설치된 바이너리는 쓰지 않았다

`~/go/bin/moai`(`v3.1.2-1490-ge79c010b8`)는 이 트리보다 한참 뒤처져 있고, M-A1에서 이미 그 탓에 4380/4378의 2건 간극이 났다. 세 바이너리를 모두 **소스에서 빌드**했다.

| 바이너리 | 출처 | 이 카드의 수리 포함 여부 |
|---|---|---|
| `moai-b7` | `b7ceffa69` 아카이브 (이 브랜치의 카드 前 조상) | 없음 |
| `moai-dev` | `fbb5cfd8d` 아카이브 (origin/develop tip, 흡수의 다른 부모) | 없음 |
| `moai-wt` | 작업 트리 = `0705eb8d7` | M-A1 + M-A2 + M-A2b |

**「수리 전」을 두 갈래로 잡은 이유와, 그것이 대조군인 이유**: 두 트리는 서로 약 190 커밋 떨어져 있고 코드도 다르다. 그런데 같은 코퍼스 위에서 두 출력이 **바이트 단위로 동일**하다.

```
$ shasum -a 256 out-b7.txt out-dev.txt
15645cb7c6b57263e3cdb4acf6d71e2b5ff47bbccece47d80fad32e3dd084e21  out-b7.txt
15645cb7c6b57263e3cdb4acf6d71e2b5ff47bbccece47d80fad32e3dd084e21  out-dev.txt
$ cmp out-b7.txt out-dev.txt   → rc=0
```

이것이 이 재계수의 대조군이다. **흡수가 들여온 약 190 커밋이 lint 거동에 기여한 양이 정확히 0**임을 보이므로, 뒤에 나오는 증감 전부를 이 카드에 귀속시킬 수 있다. 대조가 아무것도 매치하지 않는 종류의 공허한 대조가 아니라는 점도 함께 적어 둔다 — 두 바이너리는 실제로 4,416건을 출력했고, 그 4,416건이 서로 같다.

#### 재계수 — 같은 코퍼스, 파이프 없음, rc 직독

```
$ moai-b7  spec lint   → rc=0,  0 error(s), 4416 warning(s)
$ moai-dev spec lint   → rc=0,  0 error(s), 4416 warning(s)
$ moai-wt  spec lint   → rc=0,  0 error(s), 4757 warning(s)
```

**새 baseline 4,416 · 새 총계 4,757 · 순증 +341.**

#### 코드별 증감표 — 옛 값과 새 값을 나란히, 움직이지 않은 행도 싣는다

움직이지 않은 행을 빼지 않는 이유는 0도 잰 결과이기 때문이다. 15개 코드 전수다.

| 코드 | 수리 전 (4,416) | 수리 후 (4,757) | 증감 | 귀속 축 |
|---|---|---|---|---|
| `CoverageIncomplete` | 3,658 | 3,674 | **+16** | 축 1(표 수집) — **해석 유보, 아래 참조** |
| `ModalityUnjudged` | 0 | 481 | **+481** | 축 2(무판정 발화) |
| `ModalityMalformed` | 412 | 175 | **−237** | 축 2(SHALL 단어 경계화) |
| `REQTableRowsRejected` | 0 | 81 | **+81** | 축 1(기각의 관측 가능성) |
| `MovingRefUnpinned` | 115 | 115 | 0 | — |
| `StatusTransitionInvalid` | 102 | 102 | 0 | — |
| `LegacyEARSKeyword` | 48 | 48 | 0 | — |
| `MissingExclusions` | 27 | 27 | 0 | — |
| `StatusGitConsistency` | 18 | 18 | 0 | — |
| `FrontmatterInvalid` | 14 | 14 | 0 | — |
| `StatusTokenUnrecognized` | 7 | 7 | 0 | — |
| `InvalidREQID` | 6 | 6 | 0 | — |
| `SyncSHASlotFormat` | 6 | 6 | 0 | — |
| `SpecsDirMissingSpecFile` | 2 | 2 | 0 | — |
| `OwnershipTransitionInvalid` | 1 | 1 | 0 | — |
| **합** | **4,416** | **4,757** | **+341** | |

검산: +16 +481 −237 +81 = +341. 총합 차 4,757 − 4,416 = 341과 일치한다.

**줄 단위 대조 — 총량이 아니라 실제 줄을 맞췄다.** 코드·파일·줄번호를 키로 두 출력의 집합 차를 구했다.

```
$ comm -13 keys-b7.txt keys-wt.txt | wc -l   → 578   (추가)
$ comm -23 keys-b7.txt keys-wt.txt | wc -l   → 237   (제거)
추가 내역:  16 CoverageIncomplete · 481 ModalityUnjudged · 81 REQTableRowsRejected
제거 내역: 237 ModalityMalformed
```

578 − 237 = 341. **`CoverageIncomplete`의 제거는 0건**이므로 +16은 순수 신규다.

#### 옛 재유도값들 — 지우지 않고 나란히, 왜 다른지와 함께

| 트리 | 계측기 | 코퍼스 | 총계 |
|---|---|---|---|
| `b7ceffa69` | HEAD 소스 빌드 | 그 시점 코퍼스 | 4,378 |
| `b7ceffa69` | 설치 바이너리 (**틀린 계측기**, 철회됨) | 그 시점 코퍼스 | 4,380 |
| M-A1 작업 트리 | 소스 빌드 | 그 시점 코퍼스 | 4,394 |
| `458fc7ebc` (M-A2) | 소스 빌드 | 그 시점 코퍼스 | 4,617 |
| `2cfccd4eb` (M-B1) | 소스 빌드 | M-A2b 시점 코퍼스 | 4,617 |
| M-A2b 작업 트리 | 소스 빌드 | 그 시점 코퍼스 | 4,698 |
| **`b7ceffa69`** | **소스 빌드** | **병합 코퍼스 (이 측정)** | **4,416** |
| **`fbb5cfd8d`** | **소스 빌드** | **병합 코퍼스 (이 측정)** | **4,416** |
| **`0705eb8d7`** | **소스 빌드** | **병합 코퍼스 (이 측정)** | **4,757** |

옛 값과 새 값을 그대로 빼면 수리 효과와 코퍼스 증가가 섞인다(AC-SLB-009의 [HARD] 조항). 비교는 표 아래 세 줄, 즉 **같은 코퍼스 위 재유도값 사이**에서만 했다.

#### 흡수가 기여한 몫 — 카드의 효과와 분리해서 잰다

같은 `moai-b7` 바이너리를 **두 코퍼스** 위에서 돌렸다. `b7ceffa69` 아카이브 트리에는 `.git`이 없으므로 git 의존 코드 4종이 발화하지 않는다 — 그래서 비교는 git 비의존 코드에 한정한다.

```
$ (archived b7 tree)  moai-b7 spec lint → rc=0,  0 error(s), 4250 warning(s)
$ (병합 트리)          moai-b7 spec lint → rc=0,  0 error(s), 4416 warning(s)
```

git 의존 4종(`StatusTransitionInvalid` 102 + `StatusGitConsistency` 18 + `StatusTokenUnrecognized` 7 + `OwnershipTransitionInvalid` 1 = **128**)은 archived 트리에서 0이다. 4,250 + 128 = **4,378** — 기록된 `b7ceffa69` 시점 총계와 정확히 일치하며, 이는 옛 4,378을 이번 실행에서 다시 세운 셈이다.

따라서 **흡수가 기여한 몫은 +38**이다(4,416 − 4,378 = 4,288 − 4,250):

| 코드 | 흡수 전 | 흡수 후 | 증감 |
|---|---|---|---|
| `CoverageIncomplete` | 3,621 | 3,658 | +37 |
| `MissingExclusions` | 26 | 27 | +1 |
| 나머지 git 비의존 7종 | 변동 없음 | 변동 없음 | 0 |

코퍼스 자체의 이동: SPEC 디렉터리 **794 → 807** (+13, 삭제 0), 파일 3,352 → 3,412 (+60). 새 SPEC 13개는 전부 Codex 계열·플랫폼 계열이며, `REQTableRowsRejected`를 한 건도 만들지 않았다(81/787이 옛 코퍼스와 동일).

**이 +38은 카드의 +341과 겹치지 않는다.** +341은 같은 병합 코퍼스 위 두 계측기 사이의 차이라 코퍼스 증가가 양변에서 상쇄되기 때문이다.

#### 이 카드가 드러낸 것 — 네 수를 섞지 않는다

한 개의 「누락 요구사항 총수」로 합치면 그 수는 아무것도 뜻하지 않는다. 성질이 다른 네 가지다.

1. **새로 보이게 된 요구사항 (표 수집) — 16건.** 이전에는 `REQEntry`가 아예 만들어지지 않던 표 정의 행이다. `SPEC-CODEX-PARTIAL-WIRING-001` 11건 + `SPEC-INIT-001` 5건. M-A1이 옛 코퍼스에서 낸 귀속과 SPEC별로 완전히 같다 — 병합 코퍼스에서 독립적으로 재현됐다.
2. **기각이 관측 가능해진 행 — 81줄에 담긴 787행 (76개 SPEC).** 이들은 **수집되지 않는다.** 항목을 만들지 않으므로 요구사항으로 세어서는 안 되고, 달라진 것은 그 사실이 조용하지 않다는 점뿐이다. 출력에서 되읽은 합이 787이며, 대조로 수리 전 출력에 같은 탐침을 걸면 0줄/0행이다(탐침이 무효라서 0이 아님을 보이려고, 같은 출력에서 실재하는 토큰 `ModalityMalformed`를 세어 412를 얻는 catch-all 대조를 함께 걸었다).
3. **판정 불가로 새로 발화한 요구사항 — 481건 (80개 SPEC).** 이전에는 이 481건에 대해 lint가 **아무 의견도 내지 않았다**. 이 중 5건은 (1)에서 새로 수집된 표 REQ이고, 나머지 476건은 원래 수집되던 목록 REQ다. 즉 **부채의 대부분은 「안 보이던 요구사항」이 아니라 「보였지만 판정되지 않던 요구사항」**이다.
4. **오집계에서 회수된 요구사항 — 237건.** 단어 경계 수리로 `ModalityMalformed`가 412 → 175로 줄었다. **회수가 재라벨이 아님을 줄 단위로 확인했다**: 제거된 237개 위치 중 `ModalityUnjudged`로 다시 나타난 것은 **0건**이다.

```
$ comm -12 removed-malformed-loc.txt added-unjudged-loc.txt | wc -l  → 0
$ comm -12 removed-malformed-loc.txt added-coverage-loc.txt  | wc -l  → 0
$ comm -12 added-coverage-loc.txt   added-unjudged-loc.txt   | wc -l  → 5
```

**한 문장으로 재유도 가능한 형태**: 수리 전 이 코퍼스에서 lint는 — 표로 정의된 요구사항 **16건을 항목으로 만들지 못했고**, 그 밖의 표 행 **787개를 말없이 버렸으며**, 만들어 둔 항목 중 **481건에 아무 판정도 하지 않았고 237건에 틀린 판정을 했다**.

#### 출처별 귀속 (REQ-SLB-013 / AC-SLB-009 [HARD] 조항)

| 증감 | 표 수집 유래 (축 1) | 목록 수집 유래 | t385 구분자 넓힘 유래 |
|---|---|---|---|
| `CoverageIncomplete` +16 | 16 | 0 | 0 |
| `REQTableRowsRejected` +81 | 81 | 0 | 0 |
| `ModalityUnjudged` +481 | 5 | 476 | 0 |
| `ModalityMalformed` −237 | 0 | −237 | 0 |

- 축 1/목록 분리는 `REQEntry.Source` 필드가 성립시키며(AC-SLB-011이 따로 잰다), 여기서는 그 필드가 가르는 두 모집단을 **줄 위치 조인**으로 다시 확인했다: 새 `CoverageIncomplete` 16개 위치와 새 `ModalityUnjudged` 481개 위치의 교집합이 5, `ModalityMalformed` 제거분과의 교집합이 0.
- `REQTableRowsRejected`는 항목을 뒤에 두지 않아 `Source`를 싣지 못한다. 귀속은 필드가 아니라 **발화 지점이 `REQTableRejectionRule` 하나뿐이라는 구조**가 만든다.
- **t385 구분자 넓힘 유래가 전 행 0인 것은 구조적 사실이다** — 그 변경은 수리 전 두 바이너리에 이미 들어 있어 양변에서 상쇄된다. 이 열의 0은 「재지 않았다」가 아니라 「같은 것을 양쪽에서 빼서 0」이다.

#### `CoverageIncomplete` 행의 해석 유보 (AC-SLB-009 [HARD] 조항)

이 행의 +16은 **단독으로 부채를 뜻하지 않는다.** AC 수집기(`parser.go:67,218`)가 이 코퍼스의 AC 문법을 읽지 못하므로, 새로 수집된 REQ는 실제 AC 유무와 무관하게 「AC 미참조」로 보고된다. 이 SPEC 자신이 그 반례다 — acceptance.md에 AC가 16개 실재하는데도 REQ 14건 전부가 발화하며, 이번 측정에서도 이 SPEC의 spec.md는 수리 전후 모두 `CoverageIncomplete` 14건으로 동일하다. **「진짜 미참조」와 「AC를 읽지 못함」의 분리는 t528이 착지한 뒤에만 가능하다.**

#### 귀속되지 않은 +4 — 후보 셋을 지웠고, 남은 하나는 복원 불가다

M-B1 무인자 대조군의 4,621과 M-A2b의 4,617 사이 +4를 이 마일스톤에서 좁혔다. **추측으로 메우지 않는다.**

| 후보 | 판정 | 근거 |
|---|---|---|
| 코드 귀속 | **기각** | M-A2b에서 `458fc7ebc`·`2cfccd4eb` 두 바이너리가 같은 코퍼스에서 정확히 같은 4,617을 냈다. 이번 측정도 같은 성질을 재확인한다 — `b7ceffa69`·`fbb5cfd8d` 두 바이너리 출력이 바이트 동일. |
| 카드 자신의 커밋된 `spec.md` 내용 | **기각** | 두 시점(`458fc7ebc` / `2cfccd4eb`)의 spec.md를 꺼내 같은 바이너리로 각각 lint: 둘 다 `0 error(s), 14 warning(s)`. 커밋된 축에서는 움직이지 않았다. |
| 호출 형태 | **기각** | 두 측정 모두 무인자였음이 기록된 명령에 남아 있다. 덧붙여 `spec lint .moai/specs`(디렉터리 인자)는 코퍼스 스캔이 아니라 `ParseFailure` 1건·rc=1을 내므로 4,621을 낼 수 있는 형태가 아니다. |
| **그 시점 작업 트리의 미커밋 코퍼스 상태** | **남은 유일 후보 · 측정 불가(Gap)** | 그때의 미커밋 상태는 이후 덮여 복원되지 않는다. 어느 파일이었는지 특정할 근거가 없으므로 적지 않는다. |

`REQTableRowsRejected`가 아니라는 것까지는 말할 수 있다 — 그 코드는 M-A2b에서야 생겼고 M-B1 시점 바이너리에 없다. 그 이상은 Gap이다.

#### 검증 명령 (전부 파이프 없이 rc 판독, 범위는 `internal/spec`)

```
go test ./internal/spec/... -count=1
  → rc=0, ok github.com/modu-ai/moai-adk/internal/spec 73.827s
  → `--- FAIL` 0건, `test timed out` 0건, `(cached)` 아님
```

이 마일스톤은 코드를 바꾸지 않았으므로 새 테스트도 새 뮤턴트도 없다 — 재계수가 산출물이다. `internal/cli`는 건드리지 않았다(형제 SPEC 반경). 전체 스위트는 로컬에서 돌리지 않았다(CLAUDE.local.md §4).

#### 증거 파일

| 경로 | 내용 |
|---|---|
전체 출력 4종은 원본이 6.5MB라 `gzip -9`로 넣었다(`gunzip -c <경로>`로 읽는다). **SHA256 동일성 주장은 압축 전 원본에 대한 것이다** — gzip 헤더가 파일명을 싣기 때문에 `.gz`끼리는 1바이트 다르다.

| 경로 | 내용 |
|---|---|
| `.moai/reports/t518/ma3/out-b7.txt.gz` | `moai-b7` 전체 출력 (병합 코퍼스), 4,416 |
| `.moai/reports/t518/ma3/out-dev.txt.gz` | `moai-dev` 전체 출력 (병합 코퍼스), 4,416 — 압축 전 원본이 out-b7과 SHA256 동일 |
| `.moai/reports/t518/ma3/out-wt.txt.gz` | `moai-wt` 전체 출력 (병합 코퍼스), 4,757 |
| `.moai/reports/t518/ma3/out-b7-oldcorpus.txt.gz` | `moai-b7` × archived `b7ceffa69` 코퍼스, 4,250 (git 비의존분) |
| `.moai/reports/t518/ma3/codes-b7.txt` · `codes-dev.txt` · `codes-wt.txt` | 코드별 카운트 |
| `.moai/reports/t518/ma3/added.txt.gz` · `removed.txt` | 줄 단위 집합 차 (578 / 237) |
| `.moai/reports/t518/ma3/out-wt-postwrite.txt.gz` | 이 절을 쓴 뒤 재실행, 4,757 — 15개 코드 카운트가 `codes-wt.txt`와 전부 일치(diff는 내가 덧붙인 `TOTAL` 줄 한 개뿐) |
| `.moai/reports/t518/ma3/gotest-internal-spec.txt` | 테스트 출력 |

#### 잔여 위험 (관측된 것만)

- **코퍼스 전체의 REQ 항목 총수는 재지 않았다.** 출력에 그 수를 드러내는 표면이 없다. 따라서 「16건」은 절대 수이지 비율이 아니며, 「전체의 몇 퍼센트가 안 보였는가」는 이 산출물이 답하지 않는다(Gap).
- **흡수 기여분 +38의 분해는 git 비의존 코드에 한정된다.** archived 트리에 `.git`이 없어 git 의존 4종은 그 코퍼스에서 개별로 재지 못했다. 4,250 + 128 = 4,378이 맞아떨어지는 것은 정합성 대조이지, 그 4종의 옛 코퍼스 값을 각각 독립 측정한 것이 아니다.
- **81이라는 줄 수는 「연속한 표 줄」이라는 표 경계 정의 아래의 재유도값**이다. 합 787은 이 선택에 영향받지 않는다. M-A2b에 적힌 성질이 병합 코퍼스에서도 그대로다.
- **`ModalityUnjudged` 481은 「잘못됐다」가 아니라 「의견 없음」의 수**다. 이 481건 중 실제로 잘 쓰인 것과 실제로 깨진 것의 분리는 이 카드의 범위가 아니며(갈래 A, §E 범위 밖), 이 수는 그 분리를 하기 전의 모집단 크기다.
- **앵커 이후 원격이 `bce6d7e08`까지 움직였다.** 이 표의 모든 수는 `0705eb8d7`에서 잰 값이며, 다음 흡수 뒤에는 코퍼스 증가분만큼 다시 움직인다. 그때의 비교도 같은 코퍼스 위 두 재유도값 사이에서 해야 한다.


### M1 — 판독 기준 고정 + 여섯 줄 재유도 + 표 경로 반사실 (카드 t518, AC 없음 · plan.md §F M1)

**[HARD] 이 마일스톤은 계획 순서를 어겨 M2-M4 뒤에 실행됐다 — 그 대가를 먼저 적는다.**

plan.md §F는 M1을 첫 마일스톤으로 두었고, 그 이유는 「판독 기준이 고정된 뒤에 재는 값이 이 SPEC의 baseline이 된다」(spec.md §A)는 것이었다. 실제 실행 순서는 M-A1(=M2) → M-A2·M-A2b(=M3) → M-A3(=M4) → **M1**이다. 대가는 둘이고, 둘 다 관측된 것이다:

1. **M2-M4가 baseline 없이 돌았다.** M-A3의 재계수는 자기 안에서 대조군(`moai-b7` vs `moai-dev` 바이트 동일)을 세워 증감을 귀속시켰으므로 **그 마일스톤은 이 결손에 의해 무효화되지 않는다** — M-A3이 잰 것은 lint 출력의 코드별 증감이고, M1이 재는 것은 정의 행의 modality 분포다. 두 축이 다르다.
2. **아래 여섯 줄의 증감을 「기준 변경분」과 「코퍼스 증가분」으로 분해할 수 없다.** 옛 값의 판독 기준이 글로 남아 있지 않기 때문이다 — 그것이 그 값들이 「잠정」이었던 이유 그 자체다. M1을 먼저 돌렸다면 이 분해가 가능했다. 지금은 불가능하며, 아래에 Gap으로 적는다. **이 결손은 M1을 다시 실행해도 회복되지 않는다.**

#### 측정 앵커

- 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518` · 브랜치 `WT-spec-lint-axes` · HEAD **`1e273079b`**(측정 직전 `git rev-parse --show-toplevel` + `--short HEAD`로 확인).
- M-A3의 앵커(`0705eb8d7`)와 다르다. 그 사이에 이 카드의 커밋 세 개(`b65850dbe`, `a4fbaeb82`, `1e273079b`)가 얹혔다. **셋 중 어느 것도 `.moai/specs/SPEC-*/spec.md`에 정의 행을 더하지 않는다** — 두 개는 `progress.md`(코퍼스 글롭 밖)이고 하나는 `spec.md` frontmatter의 `status`/`updated` 두 줄이다.
- 계측기는 **기존 blast-radius 하네스를 확장한 것**이다(plan.md §F M1 「기존 계측기 확장」, §G 「새 측정 하네스를 만들기」 금지). 새 파일을 만들지 않았다 — `internal/spec/lint_req_widen_decompose_test.go`에 `measureM1ModalityCensus`를 더하고 기존 리포트에 두 블록을 덧붙였다.

#### [HARD] 판독 기준 — 재기 전에 글로 고정한다

옛 여섯 줄이 잠정이었던 이유는 값이 틀려서가 아니라 **어떻게 읽었는지가 적혀 있지 않아 재유도할 수 없어서**였다. 아래가 그 기준이며, 정본은 코드 주석(`internal/spec/lint_req_widen_decompose_test.go`, `measureM1ModalityCensus` 위)이다.

| 축 | 고정된 기준 |
|---|---|
| **본문 문자열** | `parseREQsWide`가 내는 `REQEntry.Text` — `reqLineWidePattern`의 캡처 그룹 2를 `strings.TrimSpace`한 것, **그 이상 아무것도 하지 않은 것**. 재트림·소문자화·볼드 제거 없음 |
| **왜 그 문자열인가** | `judgeModality`가 라이브 경로에서 받는 문자열과 **바이트 단위로 같다**. 판독기가 자기가 설명하는 검사와 어긋날 수 없게 만드는 유일한 선택이다 |
| **영어 접두사(판정됨)** | `strings.HasPrefix(strings.ToUpper(text), p)`, `p ∈ modalityPrefixes`. 대소문자 무시 — 라이브 술어가 그렇게 하기 때문 |
| **전부 대문자 형** | `strings.HasPrefix(text, p)`, `p`는 대문자 그대로 |
| **첫 글자만 대문자 형** | `strings.HasPrefix(text, "When "/"While "/"Where "/"If "/"The ")` |
| **그 밖의 대소문자** | 위 둘은 서로소이므로 `대소문자무시 − (전부대문자 + 첫글자대문자)`로 정의된다. **옛 표에는 이 행이 없었고, 그래서 두 대소문자 행을 합계와 대조할 수 없었다** |
| **볼드 마커** | 접어 넣지 않고 **자기 축으로 분리**한다. `**항상** …`로 쓰인 본문은 `*`로 시작하므로 라이브 관점에서 「한글로 시작」이 아니다. 볼드를 벗긴 관점이 그 판단을 몇 줄 움직이는지를 따로 싣는다 |
| **한글** | 룬이 U+AC00–U+D7A3(음절) · U+1100–U+11FF(자모) · U+3130–U+318F(호환 자모) 중 하나 |
| **모집단** | `parseREQsWide`가 잡은 정의 행 전부. spec.md §A가 「수집기가 실제로 잡은 정의 행 위에서 잰 값」이라 한 것과 같은 모집단 |

#### 재유도 — 여섯 줄 전부, 옛 값과 나란히

재현 명령(파이프 없음, rc 직독):

```
MOAI_T362_CORPUS_SCAN=1 go test -count=1 -timeout 15m ./internal/spec/ -run 'TestCorpusRejectedREQIDDecomposition'
→ rc=0,  ok  github.com/modu-ai/moai-adk/internal/spec  0.995s
```

**인용 대상은 `.moai/reports/t518/m1-census.txt`다**(추적됨, 이 카드의 증거 경로). 실행 로그는 `.moai/reports/t518/m1-harness-run.txt`.

**[HARD] 그렇게 한 이유 — 하네스가 남의 카드 증거를 덮어쓴다.** 하네스는 리포트 전체를 `decomposeReportRelPath`가 고정한 `.moai/reports/t362/m2-gate0-decomposition.txt`에 쓴다. 그 파일은 **추적되는 t362의 증거**이고, 재실행은 그것을 오늘 코퍼스로 다시 쓴다 — 실측: t362가 커밋해 둔 `wide_extractions_total=1085`가 `4018`로 바뀌고 기각 표본 열거가 통째로 교체됐다(`git diff --stat` → `387 insertions(+), 93 deletions(-)`). **되돌렸다**(`git checkout -- .moai/reports/t362/m2-gate0-decomposition.txt`, 이후 `git status --porcelain .moai/reports/t362/` 무출력, `wide_extractions_total=1085` 복귀). 그 뒤 `[M1]`·`[M1-cf]` 두 블록만 위 경로로 추출해 헤더(트리·HEAD·명령)를 붙여 커밋했다.

이 덮어쓰기 위험은 **이 카드의 확장이 만든 것이 아니다** — 하네스를 재실행하는 누구나 겪는다. 고치지 않고 보고한다(범위 밖, §E.3 F-A2).

| # | 측정 | 옛 값(잠정, v0.3.0) | **새 값(기준 고정, `1e273079b`)** | 증감 | 성질 |
|---|---|---|---|---|---|
| 1 | 수집된 정의 행 | 3,932 → 3,952 → 3,953 | **4,018** | +65 (3,953 기준) | 모집단 |
| 2 | 영어 modality 접두사(대소문자 무시) | 2,399 | **2,417** | **+18** | 실제로 판정되는 행 |
| 3 | 그중 전부 대문자 형 | 463 | **463** | **0** | **움직이지 않았다 — 그래서 싣는다** |
| 4 | 그중 첫 글자만 대문자 형 | 1,933 | **1,951** | **+18** | |
| 5 | 본문이 한글로 시작 | 149 | **189**(볼드 유지) / 198(볼드 제거) | **+40 / +49** | 구조적으로 한 번도 판정되지 않는 행 |
| 6 | 본문에 한글이 섞임 | 501 | **570** | **+69** | |

**[HARD] 옛 값을 지우지 않았다.** 이 카드의 규칙은 「움직인 수치를 조용히 새 값으로 갈아 끼우지 않는다」이며(plan.md §G), 증감 0인 3행도 그 규칙에 따라 실린다.

새로 드러난 두 사실 — 옛 표가 답하지 않던 것들:

- **행 2 = 행 3 + 행 4 + 3.** 잔여 3행은 두 명명된 대소문자 형 어느 쪽도 아니다(`WHEN`/`When` 이외의 표기). 옛 표는 이 잔여 행이 없어 `2,399 vs 463+1,933=2,396`의 3 간극을 드러내지 못했다 — **같은 3이 옛 값에도 있었다.** 이것은 기준 고정이 곧바로 되찾은 정보다.
- **볼드 마커가 실제로 9행을 움직인다**(189 → 198). 「볼드 마커 처리」가 잠정 사유로 이름만 올라 있던 축의 실측값이다. 라이브 관점(189)이 정본이다 — `judgeModality`가 보는 문자열이 그것이기 때문이며, 그 9행은 **한글로 시작하지만 볼드 때문에 접두사 검사에도 걸리지 않아** 여전히 무판정으로 떨어진다(결론은 바뀌지 않는다).
- **접두사별 분해**(옛 표에 없던 축): `THE ` 1,608 · `WHEN ` 412 · `WHILE ` 178 · `WHERE ` 165 · `IF ` 54. 합 2,417 = 행 2. 첫 매치에서 멈추므로 중복 계수가 없다.

#### 반사실 — 표 경로에 자문을 적용하지 않았다면 (plan.md §F M1의 미측정분)

**[HARD] 이 반사실의 창은 닫히지 않았다 — 배차문의 전제를 정정한다.** 「M-A1/M-A2가 착지했으므로 수리 전 창이 닫혔다」는 골든 채취(트리 상태)에는 맞지만 **이 반사실에는 맞지 않는다.** 「표 경로에 자문 미적용」은 과거 트리의 상태가 아니라 **severity 배정의 가정**이고, `reqFindingSeverity`의 입력이 `REQEntry.Widened` 하나뿐이며 표 유래 항목은 `REQEntry.Source == REQSourceTable`로 지금도 식별 가능하므로, 이 트리에서 직접 셀 수 있다. M-A3처럼 아카이브에서 바이너리를 빌드할 필요가 없었다.

`[M1-cf]` 블록 축자:

```
m1cf_entries_source_list=4018
m1cf_entries_source_table=20
m1cf_table_modality_malformed=0  # would become SeverityError
m1cf_table_modality_conforming=15
m1cf_table_modality_unjudged=5
m1cf_table_invalid_reqid=0  # would become SeverityError  pattern=^REQ-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)?-\d{3}(?:-\d{3})?$
```

**판정: 자문 미적용의 반사실 비용은 error-severity finding 0건이다.** 표 수집이 들여온 20개 항목 중 `ModalityMalformed`로 판정되는 것도, `InvalidREQID`에 걸리는 것도 없다. 15건은 판정되어 적합, 5건은 무판정이며 무판정은 자기 코드(`ModalityUnjudged`, 항상 자문)로 나가므로 `Widened` 분기와 무관하다.

**그러므로 「표 경로를 자문으로 둔 결정」은 이 코퍼스 위에서 아무 finding도 억누르지 않고 있다.** 이것은 그 결정을 정당화하지도 부정하지도 않는다 — 자문 처리는 *지금 무엇이 억눌리는가*가 아니라 *새 코퍼스 사실을 저자의 회귀로 오독시키지 않는다*를 위한 것이고, 억누르는 양이 0이라는 사실은 **그 결정을 되돌리는 비용도 0**임을 말한다. 그 되돌림은 이 카드의 범위 밖이며(`reqFindingSeverity`의 THE DEBT 주석이 이미 승격 조건을 적어 두었다), 이 수치는 그 조건을 재는 사람에게 남기는 입력이다.

#### Gap — 재실행으로 회복되지 않는 것

- **[HARD] 여섯 줄 증감의 원인 분해를 하지 못했다.** 옛 값의 판독 기준이 기록되지 않아, `+18`·`+40`·`+69` 각각이 (a) 코퍼스 증가 (b) 판독 기준 변경 (c) 기존 SPEC 본문 수정 중 무엇에서 왔는지 갈라지지 않는다. 유일하게 귀속 가능한 성분은 코퍼스 크기다 — 추적된 `spec.md`가 base `0b1e27877`에서 **790**, 지금 **804**(`git ls-tree -r --name-only 0b1e27877 -- .moai/specs` 의 `/spec.md` 행 수 vs `git ls-files -- '.moai/specs/SPEC-*/spec.md' | wc -l`), **+14 SPEC**. 나머지는 미분해다.
- **옛 기준으로 재측정해 분해하는 길은 없다.** 그 기준이 무엇이었는지가 관측 불가능하므로, 「옛 기준을 재현해 새 코퍼스에서 돌린다」는 절차 자체가 정의되지 않는다. **이것이 M1을 나중에 돌린 실제 비용이며, 다음 회차에 회복되는 종류의 것이 아니다.**
- **이 기준 아래의 다음 측정부터는 분해가 가능하다.** 기준이 코드 주석과 이 절에 고정됐고 계측기가 재실행 가능하므로, 다음 재유도의 증감은 코퍼스 증가와 본문 수정으로 갈린다(기준 성분이 0이 되므로).
- **`m1_body_starts_hangul_*`는 「판정되지 않는다」의 하한이지 전부가 아니다.** 영어 접두사도 없고 `SHALL`도 없는 행은 첫 글자가 한글이 아니어도 무판정으로 떨어진다. 그 전체 수는 `ModalityUnjudged` 481(M-A3)이며 이 census가 다시 재지 않았다.

#### manager-spec 후속 — `spec.md` §A의 잠정 표

**[HARD] `spec.md`는 이 에이전트의 소관이 아니므로 고치지 않았다.** `spec.md:89-98`의 여섯 줄은 지금도 「잠정」이라 적혀 있고, 그 문장(「M1이 기준을 고정한 뒤 다시 재는 값이 이 SPEC의 baseline이 된다」)의 조건은 **충족됐다** — 기준은 위에 고정됐고 값은 재유도됐다. 그 표를 새 값으로 갱신하고 잠정 표시를 떼는 것은 계획 아티팩트 개정이며 manager-spec 소관이다. §E.3에 **F-A1**로 올린다. 이는 M-A2 에이전트가 `acceptance.md`에 대해 한 처리(v0.6.0 기록)와 같은 경계다.

### M5 — 문서 정리: 판별식과 무판정의 의미를 코드에 남긴다 (카드 t518, AC 없음 · plan.md §F M5)

plan.md §F M5는 「판별식과 무판정 코드의 의미를 코드 주석과 SPEC에 남김」이다. **코드 주석 쪽이 이 마일스톤의 소관이고, SPEC 쪽은 아니다** — `spec.md`는 계획 아티팩트이므로 §E.3에 후속으로 올린다.

#### [HARD] 먼저 잰 것: 세 파일 중 둘은 이미 되어 있었다

M5를 「할 일」로 받아 그대로 세 파일에 글을 더하는 것은, 이미 있는 설명 옆에 두 번째 사본을 만드는 일이다. 그래서 **쓰기 전에 읽었다.**

| 파일 | M5가 요구하는 내용 | 실측 상태 |
|---|---|---|
| `internal/spec/lint_req_table.go` | C-d 판별식 + L1 어휘를 **왜** 좁게 골랐는가(두 독립 어휘 사이의 재현성이지 잔존 수의 크기가 아니다) | **이미 있다** — 파일 헤더의 `WHY C-d AND NOT A WIDER LEXICON` 문단이 C-d 20(두 어휘에서 동일) vs C-e 61↔63을 그대로 적고, 이어지는 `[HARD] THE MISS IS A CHOSEN COST` 문단이 `SPEC-INIT-001`의 `REQ-N-001` 실측 놓침과 「어휘를 넓혀 고치지 않는다」를 못박는다. M-A1이 남긴 것 |
| `internal/spec/lint_req_table_rejection.go` | finding이 **표 단위로 접히는** 이유와 N을 싣는 이유 | **이미 있다** — `THE FOLD IS PER TABLE, NOT PER ROW — AND THE COST OF THE FOLD IS NAMED`(787행 vs 460 findings, +5.1%, 관측 단위가 행에서 표로 옮겨간다는 대가 명시)와 `WHY N RIDES ON THE LINE`(N이 두 번째 계측기가 되어 코퍼스 총합 대조가 가능해진다, 대조는 `TestTableRejection_CorpusSumEqualsRowCensus`). M-A2b가 남긴 것 |
| `internal/spec/lint.go` | `judgeModality`의 **세 상태**와, SHALL 접촉 조건을 단어 경계로 바꾼 이유 | **부분적** — 세 상태는 `modalityVerdict` 타입과 상수 블록에, 단어 경계 교체 사유는 `shallWordPattern` 위에 각각 있었으나, **`judgeModality` 함수 자신은 한 줄짜리 주석뿐**이었다 |

**그래서 이 마일스톤이 실제로 더한 것은 두 곳뿐이다.** 나머지를 다시 쓰지 않은 것이 이 마일스톤의 산출물의 일부다 — 같은 설명의 두 번째 사본은 다음 저자에게 「어느 쪽이 정본인가」를 묻게 만들고, 그 질문에 답이 없으면 둘은 갈라진다.

#### 더한 것 ①: `judgeModality` 함수 주석 (`internal/spec/lint.go`)

한 줄이던 함수 주석을, 독자가 그 함수에 착지했을 때 세 상태의 경계를 읽을 수 있는 분량으로 늘렸다. 담은 것:

- 세 결과와 각각의 조건(CONFORMING / MALFORMED / **UNJUDGED**).
- **UNJUDGED는 판정이 아니라는 것** — 그것을 「괜찮다」로 읽는 호출자가 이 SPEC이 없애려는 결함을 재생산한다. `isModalityMalformed`의 `false`만이 뒤 두 상태를 뭉개며, 그 함수의 주석이 이미 「false는 더 이상 well formed를 뜻하지 않는다」고 적고 있다는 연결을 명시했다.
- 단어 경계 교체가 **한국어를 판정하는 것이 아니라는 것** — 코퍼스가 `…해야 한다(SHALL)`로 쓰므로 기존 어휘가 이미 아는 SHALL이 괄호 안에 있을 뿐이다. 옛 선행 공백 조건은 그 `(SHALL)`을 못 봤다.

**타입 위의 설명을 함수로 복사하지 않았다** — 함수 주석은 세 상태의 *경계와 오독 위험*을 말하고, 타입 주석은 각 상수의 *정의*를 말한다. 축이 다르다.

#### 더한 것 ②: `isTableDefinitionRow` 주석 (`internal/spec/lint_req_table.go`)

한 줄이던 것에 세 가지를 더했다: 이것이 **축 1의 유일한 결정점**이라는 것, 접촉이 셀이 아니라 **행 전체**에 대해 일어난다는 것, 그리고 **판별 근거는 파일 헤더에 있으니 거기를 읽으라는 것**(여기에 다시 적지 않는다 — 그것이 사본을 만드는 일이다). 함께: 이 함수가 기각한 행은 침묵하지 않으며 `REQTableRejectionRule`이 표 단위로 보고한다는 연결.

#### 검증 — 주석만 바꿨으므로 무회귀 확인이다

```
$ go test -count=1 -timeout 20m ./internal/spec/...
rc=0
ok  	github.com/modu-ai/moai-adk/internal/spec	77.665s
```

(트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518` · HEAD `fc02d2542` · 전문 `.moai/reports/t518/m5-spec-tests.txt`.) `gofmt -l`은 두 파일 모두 무출력, `go vet ./internal/spec/` rc=0.

**[HARD] 이 초록은 「주석이 옳다」를 뜻하지 않는다.** 주석은 컴파일되지 않으므로 어떤 테스트도 그 내용을 판정하지 않는다 — 이 실행이 세우는 것은 **편집이 코드를 깨지 않았다**는 것 하나뿐이다. 주석의 정확성은 다음 독자가 코드와 대조해 판정할 몫이며, 그것이 이 절이 각 문단에 「무엇을 담았는지」를 적어 두는 이유다.

#### M-A3의 미추적 증거를 이 커밋에서 추적으로 옮겼다

`.moai/reports/t518/ma3-merged-spec-tests.txt`(`ok  github.com/modu-ai/moai-adk/internal/spec  79.069s`)는 M-A3 창에서 만들어졌으나 미추적으로 남아 있었다. 내용을 확인했고 M-A3이 병합 트리에서 돌린 `internal/spec` 테스트 출력이 맞다 — 이 커밋에서 커밋한다. M-A3 기록이 「창 전후 두 번 다 미추적 1건이었다」고 적은 관측은 **그 시점에 참이었고 지금도 그 시점에 대해 참이다**; 지금 추적으로 옮기는 것이 그 문장을 거짓으로 만들지 않는다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: fe75ec5c8          # 마지막 run-phase 구현 커밋(M5). 이 §E.3 기록 자체는 뒤따르는 커밋에 실린다
run_record_commit_sha: 3d675d99d   # 이 §E.3 절이 실린 커밋. 커밋은 자기 SHA를 담을 수 없어 placeholder 로 착지한 뒤 backfill 했다
run_status: implemented
ac_pass_count: 16                  # 식별자 16개 전수
ac_fail_count: 0
ac_pass_with_limitation_count: 2   # AC-SLB-003(변별력 제한) · AC-SLB-012(축자 픽스처 1종 결함, 개정으로 닫힘)
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not_performed        # push 하지 않음 (배차 지시)
l44_post_push_fetch: not_applicable
new_warnings_or_lints_introduced: not_measured    # golangci-lint 미실행 — E5
cross_platform_build: not_measured                # GOOS 교차 빌드 미실행 — E2
coverage: not_measured                            # -cover 미실행 — E3
total_run_phase_files: 7           # 구현·테스트 파일. 아래 「파일」 절이 전수
m1_to_mN_commit_strategy: "마일스톤당 커밋 1개 이상, 계획 순서와 다른 실행 순서. M2→M3→M3'→M4→M1→M5. 아래 마일스톤 대조표가 계획↔실행↔커밋을 잇는다"
verification_tree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518
verification_branch: WT-spec-lint-axes
verification_head_at_record: fe75ec5c8
```

### E1 — AC 판정 행렬 (식별자 16개 전수)

「판정한 마일스톤」은 그 AC를 판정한 기록이 있는 절이며, 「근거」는 그 절이 인용한 테스트 이름 또는 증거 경로다. 상세는 각 마일스톤 절에 있고, 이 표는 전수 색인이다.

| AC | 판정 마일스톤 | 판정 | 근거 |
|---|---|---|---|
| AC-SLB-001a | M-A1 | **PASS** | `TestTableCollection_ListFormUnchanged` |
| AC-SLB-001b | M-A1 | **PASS** | `TestTableCollection_TableFormCollected` |
| AC-SLB-001c | M-A1 | **PASS** | `TestTableCollection_ControlPairDiverges` |
| AC-SLB-002 | M-A1 | **PASS** | `TestTableCollection_AdvisoryDoesNotGate` |
| AC-SLB-003 | M-A1 | **PASS(변별력 제한 명시)** | `TestTableCollection_CorpusListFindingsUnchanged` — narrow 경로가 코퍼스 전체에서 `LegacyEARSKeyword` 7건만 내므로 단언의 변별력이 그 7건과 순서 의존 `DuplicateREQID`에 한정된다. **제한이 M-A1 절에 적혀 있으므로 은폐가 아니다** |
| AC-SLB-004 | M-A1 | **PASS** | `TestTableCollection_DiscriminatorRejectsDispositionTables` — 부재 단언이므로 **M-A1 뮤턴트 표와 짝으로만 읽는다** |
| AC-SLB-005 | M-A2b | **PASS** | 기준 5개 전부 + 뮤턴트 M1. `corpus rejection: rows=787 emitted lines=81 sum(N)=787` |
| AC-SLB-006a | M-A2 | **PASS** | `TestModality_EnglishControlStillJudged` |
| AC-SLB-006b | M-A2 | **PASS** | `TestModality_KoreanControlAnnounced` — 두 코드 중 정확히 하나. 「둘 다 0」을 `t.Fatalf`로 분리 |
| AC-SLB-007 | M-A2 | **PASS** | `TestModality_SilenceDiffersFromConformance` |
| AC-SLB-008 | M-A2 | **PASS** | `TestModality_UnjudgedIsAdvisory` — 픽스처가 `Widened=false`임을 먼저 단언해 t385 경로를 배제 |
| AC-SLB-008b | M-A2 | **PASS** | `TestModality_UnjudgedIsNotReportedConforming` — `judgeModality` 반환값 직독, 부재로 추론하지 않음 |
| **AC-SLB-009** | M-A3 | **PASS — 이 실행에서 처음 판정한다** | 아래 「AC-SLB-009 판정」 |
| AC-SLB-010 | M-A1(절반) → M-A2(완결) | **PASS** | `TestModality_AdvisoryDoesNotLeakIntoErrors` — 두 자문 경로를 한 픽스처에서 동시 발화 |
| AC-SLB-011 | M-A1 | **PASS** | `TestTableCollection_SourceRecordedPerOrigin` + `..._SourceDoesNotDecideSeverity` |
| AC-SLB-012 | M-A2 | **PASS(축자 픽스처 1종 결함 발견 → v0.6.0 개정으로 닫힘)** | `TestModality_ShallContactIsWordBoundary` — 가드를 만드는 것은 교정 픽스처 1종뿐이며 그 사실이 `oldShallContact` 헬퍼로 기계 단언돼 있다 |

**PASS 16 · FAIL 0.** 그중 둘은 한정어를 달고 통과한다(003의 변별력 제한, 012의 픽스처 결함) — **한정어는 판정의 일부이지 각주가 아니다.**

#### AC-SLB-009 판정 (이 실행에서 `acceptance.md` 기준에 대조해 판정)

배차문의 관측대로 **어느 절에도 AC-SLB-009의 명시적 PASS 행이 없었다.** M-A3은 그 AC의 주제이면서 판정을 적지 않았다. 이 실행에서 `acceptance.md`의 기준 다섯 항을 M-A3 기록에 하나씩 대조했다.

| AC-SLB-009 요구 | M-A3 기록의 대응 | 충족 |
|---|---|---|
| ① 수리 전 경로와 ② 수리 후 경로를 **같은 시점의 같은 코퍼스**에 실행 | `moai-b7`·`moai-dev`(수리 전) vs `moai-wt`(수리 후), 앵커 `0705eb8d7` 한 시점, 같은 병합 코퍼스. rc 직독 4,416 / 4,416 / 4,757 | 예 |
| 코드별 증감표, 각 행이 **두 재유도값**을 나란히 | 「코드별 증감표 — 옛 값과 새 값을 나란히, 움직이지 않은 행도 싣는다」, 15개 코드 전수 | 예 |
| 증거 파일 경로가 보고서에 인용 | `.moai/reports/t518/ma3/` 11파일, `progress.md`의 증거 표에 경로별로 인용됨(추적 확인: `git ls-files -- .moai/reports/t518/ma3/` 11행) | 예 |
| **[HARD]** 각 행이 **출처별로** 갈라짐(표 수집 유래 vs t385 넓힘 유래) | 「출처별 귀속」 절 — 증감 16행 전부 표 수집 유래, t385 넓힘 유래 0행. 줄 위치 조인으로 재확인 | 예 |
| **[HARD]** 저장된 4,346과 빼지 않음 | 「옛 값과 새 값을 그대로 빼면 수리 효과와 코퍼스 증가가 섞인다」를 명시하고 같은 코퍼스 위 세 재유도값 사이에서만 비교 | 예 |
| **[HARD]** `CoverageIncomplete` 행 별도 표기 + t528 유보 본문 명기 | 「`CoverageIncomplete` 행의 해석 유보」 절 | 예 |

**판정: AC-SLB-009 PASS.** AC 자신의 판정문(「증감표가 없으면 완료가 아니다 … `CoverageIncomplete` 해석 유보는 FAIL이 아니다」)이 요구하는 것이 전부 있다.

**[HARD] 이 판정은 소급 기록이며, 내가 M-A3을 다시 돌린 것이 아니다.** 위 여섯 행은 전부 커밋된 M-A3 기록과 커밋된 증거 파일의 **재판독**이다. 이 실행에서 새로 관측한 것은 증거 파일의 추적 여부(`git ls-files` 11행) 하나뿐이다.

### 마일스톤 대조 — 계획 M1-M5 vs 실제 착지

| plan §F | 계획 내용 | 실제 실행 | 커밋 | 순서 |
|---|---|---|---|---|
| M1 | 기존 계측기 확장 + 판독 기준 고정 + 반사실 | **M1** | `fc02d2542` | **5번째** |
| M2 | 축 1 표 수집 (RED→GREEN) | M-A1 | `6cfcfef00` | 1번째 |
| M3 | 축 2 갈래 B 무판정 발화 + SHALL 단어 경계 | M-A2 | `458fc7ebc` | 2번째 |
| M3(잔여) | — (계획에 없던 분리) | **M-A2b** — AC-SLB-005 기각 관측 가능성. M-A2가 「소관이 축 1」이라며 별도 마일스톤으로 미룬 것 | `fc0540f41` | 3번째 |
| M4 | 코퍼스 재계수 | M-A3 | `b65850dbe` · `a4fbaeb82` | 4번째 |
| M5 | 문서 정리 | **M5** | `fe75ec5c8` | 6번째 |

**[HARD] 실행 순서는 계획 순서가 아니다: M2 → M3 → M3' → M4 → M1 → M5.** M1이 마지막에서 두 번째로 밀린 대가는 §E.2 M1 절의 첫 문단이 값으로 적었다(여섯 줄 증감의 원인 분해 불가). 이 표를 싣는 이유는 커밋 이력만으로는 계획 마일스톤과 실행 마일스톤의 대응이 복원되지 않기 때문이다 — 이름이 다르고(M2 vs M-A1), 개수도 다르다(5 vs 6).

### 파일 — run-phase가 만든 것 (전수)

| 파일 | 성격 | 착지 마일스톤 |
|---|---|---|
| `internal/spec/lint_req_table.go` | 신규 — 표 수집기 + C-d 판별식 | M-A1 (주석 M5) |
| `internal/spec/lint_req_table_test.go` | 신규 — 축 1 테스트 | M-A1 |
| `internal/spec/lint_req_table_rejection.go` | 신규 — 기각 관측 가능성 | M-A2b |
| `internal/spec/lint_req_table_rejection_test.go` | 신규 | M-A2b |
| `internal/spec/lint.go` | 변경 — 3상태 modality + `shallWordPattern` + `judgeModality` 주석 | M-A2 (주석 M5) |
| `internal/spec/lint_req_widen.go` | 변경 — `parseREQsWithProvenance`가 표 수집을 병합 | M-A1 |
| `internal/spec/lint_req_widen_decompose_test.go` | 변경 — `measureM1ModalityCensus` 확장 | M1 |

### E2 — 크로스 플랫폼 빌드

**미측정.** GOOS 교차 빌드를 이 실행에서 돌리지 않았다. 부재를 통과로 읽지 않는다.

### E3 — 커버리지

**미측정.** `-cover`를 실행하지 않았다.

### E4 — 서브에이전트 경계

이 실행은 서브에이전트를 스폰하지 않았다. 해당 없음.

### E5 — lint

**미측정.** `golangci-lint`를 실행하지 않았다. 실행한 정적 검사는 `gofmt -l`(두 파일 무출력)과 `go vet ./internal/spec/`(rc=0)뿐이며, 그 둘은 lint 게이트가 아니다.

### E6 — push 상태

**push 하지 않았다** — 배차 지시. 이 카드가 만든 커밋은 로컬에만 있다.

### E7 — 상태 전이

`spec.md` frontmatter는 M-A1 착지 시 이미 `draft → in-progress`로 전이됐다(v0.6.0 기록). 이 실행은 BLIND-AXES의 상태를 움직이지 않았다 — `in-progress → implemented → completed`는 manager-docs 소관이다.

### E8 — RED 증거

이 실행이 착지시킨 두 마일스톤(M1 · M5)은 **어느 쪽도 새 AC를 만들지 않는다.** M1은 계측 확장(측정이 산출물, 단언 없음), M5는 주석뿐이다. 따라서 RED-GREEN 쌍이 성립하지 않으며 **RED 증거가 없는 것이 결손이 아니다.** AC를 가진 마일스톤들(M-A1·M-A2·M-A2b)의 RED 증거는 각 절에 있다.

### manager-spec 후속 (계획 아티팩트 — 이 에이전트의 소관 밖)

| # | 아티팩트 | 내용 | 근거 |
|---|---|---|---|
| **F-A1** | `spec.md` §A (`:89-98`) | 여섯 줄 표가 아직 「잠정」이다. 그 표 자신의 조건(「M1이 기준을 고정한 뒤 다시 재는 값이 baseline이 된다」)이 **충족됐다** — 기준은 `measureM1ModalityCensus` 주석과 §E.2 M1에 고정됐고 값은 재유도됐다(4,018 / 2,417 / 463 / 1,951 / 189 / 570). 새 값으로 갱신하고 잠정 표시를 떼되, **옛 값을 지우지 말고 병기**해야 한다(§G 규율) | §E.2 M1 |
| **F-A2** | 계획 밖 — 도구 결함 | blast-radius 하네스가 `decomposeReportRelPath`로 **추적되는 t362 증거 파일**에 쓴다. 재실행이 그 카드의 커밋된 수치를 오늘 코퍼스로 덮어쓴다(실측 후 되돌림). 출력 경로를 호출자가 정할 수 있게 하거나 카드별로 갈라야 한다. **이 카드의 범위 밖** | §E.2 M1 |
| **F-A3** | `plan.md` §F | 실행이 6개 마일스톤으로 갈렸다(M-A2b가 분리). 계획 표는 5개이며 M-A2b에 대응하는 행이 없다 | 위 마일스톤 대조표 |

### Gaps — 이 실행에서 관측하지 않은 것

- **커버리지·lint·교차 빌드 미측정**(E2/E3/E5). 부재를 통과로 읽지 않는다.
- **M-A1·M-A2·M-A2b·M-A3의 테스트를 이 실행에서 개별 재실행하지 않았다.** `go test ./internal/spec/...` 전체가 rc=0으로 통과했으므로 **그 안에 포함되어 통과했다**는 것까지가 관측이며, 각 AC 테스트가 셀렉터에 실제로 잡혔음을 테스트 단위 출력으로 보이지는 않았다. 패키지 초록은 그 안의 특정 테스트가 실행됐다는 사실보다 약하다.
- **AC-SLB-003의 변별력 제한은 이 실행에서 다시 재지 않았다.** M-A1이 적은 「narrow가 코퍼스 전체에서 `LegacyEARSKeyword` 7건만 낸다」는 그 시점 코퍼스의 값이고, M1이 코퍼스가 그 뒤 움직였음을 보였다(정의 행 4,018). 그 7이 지금도 7인지는 미측정이다.
- **M1의 여섯 줄 증감 원인 분해 불가**(§E.2 M1 Gap). 재실행으로 회복되지 않는다.

### Residual-risk — 관측했음에도 여전히 틀릴 수 있는 것

- **주석은 컴파일되지 않는다.** M5가 더한 두 주석의 정확성은 어떤 테스트도 판정하지 않는다. `go test` rc=0이 세우는 것은 편집이 코드를 깨지 않았다는 것뿐이다.
- **M1 census는 자기가 설명하는 함수와 같은 코드를 부른다.** `parseREQsWide`와 `modalityPrefixes`를 그대로 쓰므로, **그 함수들 자체가 틀렸다면 census도 같은 방향으로 틀린다.** census가 재는 것은 「분포가 어떠한가」이지 「수집기가 옳은가」가 아니다 — 후자는 M-A1의 축자 픽스처 몫이다.
- **M1의 반사실 0건은 오늘 코퍼스의 성질이다.** 표 유래 항목 20개 중 malformed가 0인 것은 코퍼스가 그렇게 생겼기 때문이지 구조적 보장이 아니다. 코퍼스가 자라면 0이 아니게 될 수 있고, 그때 자문 처리의 억제량이 처음으로 0이 아니게 된다.
- **`ModalityUnjudged` 481은 「깨졌다」가 아니라 「의견 없음」의 모집단 크기다.** 그중 실제로 잘 쓰인 것과 깨진 것의 분리는 갈래 A(후속 카드) 소관이며, 이 카드는 그 분리를 하지 않았다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
