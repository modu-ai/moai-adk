---
id: SPEC-AUTONOMY-BATCH-GATE-001
title: "Acceptance criteria — 배치 게이트 요약"
version: "0.4.1"
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
module: ".claude/rules/moai/workflow, .claude/skills/moai/workflows, internal/template/templates, internal/hook"
tier: L
---

# SPEC-AUTONOMY-BATCH-GATE-001 — Acceptance

## §D AC Matrix

검증 계층이다. 요구사항 계층(GEARS)은 `spec.md` §C가 소유하고, 여기서는 반복하지 않는다. 각 기준은 `Given … When … Then …` 형식이고 `AC-NNN`으로 라벨링했다. 분류는 두 가지다.

- **release-blocking**: RED-now 칸(명령 · 원문 stdout · 종료 코드 · 트리 SHA)을 §Evidence Ledger에 갖는다. 문서 수준 고정: **트리 `72e09d27b`**(전체 `72e09d27b0f7f8140bb11223c0d44a36d44984bb`, 브랜치 `WT-batch-approval-gate`). 0.3.0에서 RED 칸 전부를 이 트리에서 다시 돌렸고, 0.4.0에서 바꾼 칸(RED-3의 3d–3f, RED-6의 6d, RED-7의 7c·7d)은 칸에 적힌 트리 `08692e732`에서 다시 돌렸다(0.4.0의 SPEC 파일 수정은 이후 커밋 `4dec6281c`가 담았고, `git diff --stat 08692e732 HEAD -- . ':(exclude).moai/specs'`가 빈 출력, exit 0이라 소스 파일은 같다). 칸 자체의 고정이 문서 수준 고정보다 우선한다. 0.2.0 값은 `c50da9c2f`에서 쟀고 두 트리의 소스 파일은 같다(`git diff --stat c50da9c2f HEAD -- . ':(exclude).moai/specs'`가 빈 출력, exit 0). 이 고정은 칸 자체에 고정이 없는 모든 기준에 적용된다.
- **regression-guard**: 현재 GREEN인 가드가 계속 GREEN이어야 하거나, 감사가 읽어 판정하는 시나리오. RED-now 칸 없음. 시나리오 기준은 기계 점검이 불가능하며 그 사실을 §D.1에 공시했다.

검증 명령은 단순 명령·리터럴 경로·`-run` 앵커 패턴만 쓴다(워크트리 가드 안전). 테스트 계열 기준의 통과는 `--- PASS` 줄과 *비어 있지 않은 하위 테스트 집합*을 함께 읽어야 한다(빈 selector의 `ok`는 통과가 아니다 — `verification-completeness.md` §1.1).

### AC-001 — 정본 절이 라이브와 미러에 존재한다 (release-blocking, RED: RED-1)

- **Given** 변경 전 트리에서 `auto-semantics.md`는 §9.1에서 끝나는 §9이다.
- **When** run-phase가 정본 절을 작성한다.
- **Then** 라이브와 템플릿 미러 양쪽에 제목 `### 9.2 The batch gate summary`가 정확히 한 번 있다.
- Verify: `grep -c "^### 9.2 The batch gate summary" .claude/rules/moai/workflow/auto-semantics.md` → `1`
- Verify: `grep -c "^### 9.2 The batch gate summary" internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` → `1`
- 얕은 기준이다: 제목만 있는 빈 절도 통과한다. 그래서 AC-002~007이 절 *안의* 내용을 점검한다.

### AC-002 — 요약 서식, 질문 하나, 승인 범위, 빼내기와 읽기 규칙, 선호 배출 비약화 (release-blocking, RED: RED-2)

- **Given** plan→run Kickoff 행에서 운영자 결정을 기다리는 카드가 둘 이상이다.
- **When** 정본 절의 서식 규칙(REQ-BGS-004·005·006·020)을 점검한다.
- **Then** 정본 절은 (a) 보고서 머리가 Kickoff 행임을 적고 행 필드 전부 — 카드 id, SPEC id, 독립 plan-audit 판정 참조(최종 반복 판정·반복 식별자·점수·PASS 기준 대비 여유), 해시 불변 확인, 기록 참조, keep-set·리더 보유 권한 점검 결과와 분류 근거, `counter_refs=` — 와 반대 증거 행 우선 정렬 및 보고서가 질문보다 앞선다는 순서(앵커 A05–A08), (b) 질문이 정확히 하나이고 한 호출에 요약 질문이 하나뿐이며 승인이 나열된 승인 가능 행에만 미치고 나중에 온 행·예약 행·차단 행은 덮이지 않는다는 진술(A09–A11), (c) 행 수와 무관한 빼내기 수단 — 질문 채널의 자동 자유 입력에 빼낼 카드 id를 적는 방식 — 과 빼내도 나머지 행의 승인이 유지됨(A12–A13), 질문 문구가 읽기 규칙(적힌 카드 id는 빼내고 나열된 나머지는 승인)을 적는다는 것(A41), 읽을 수 없는 입력이나 표에 없는 카드 id는 아무것도 승인하지 않고 같은 질문을 다시 낸다는 것(A42), (d) 승인이 Kickoff의 다른 조건을 약화하지 않으므로 카드마다 티어·모드 선호·PR 전략·체인 범위가 run 진입 전에 디스크에 있어야 한다는 것(A45)을 명시한다.
- Verify: `go test -count=1 -v -run '^TestBatchGateSummaryDoctrine$' ./internal/template/` — `--- PASS: TestBatchGateSummaryDoctrine/real_section ` (이름 뒤 공백 포함)과 A05–A13·A41·A42·A45 하위 테스트가 모두 `--- PASS`이고 출력 계수는 AC-007의 비공허 점검을 따른다.
- 변이 탐침: 빼내기 수단을 "행마다 선택지"로 바꾼 본문(행이 4개를 넘으면 성립하지 않는다)은 A12 위반, 읽기 규칙을 "적힌 id만 승인"으로 뒤집은 본문은 A41 위반, 읽을 수 없는 입력을 "전부 승인"으로 읽는 본문은 A42 위반, 요약 승인이 모든 카드의 선호를 한 번에 배출한다고 쓴 본문은 A45 위반으로 거부된다.

### AC-003 — 반대 증거 필드, `none searched=`, 행별 결정 기록, 배치 식별자, 기록 시점 재확인 (release-blocking, RED: RED-2)

- **Given** 어떤 행에도 반대 증거가 없거나 있다.
- **When** 정본 절의 REQ-BGS-007·008·009·019 규칙을 점검한다.
- **Then** 정본 절은 `counter_refs=` 필드(A14)와 원천 목록 — 감사 경고·부채, 여유, 미해결 decision-index 행, 감사 교차 불일치, 열린 차단·대기 기록, 행 간 경로 겹침(A15), `counter_refs=none`에는 `searched=`와 공백 없는 단일 토큰이 따라야 하고 보고서와 기록 양쪽에 적용된다는 문장(A16)과 `searched=` 없는 `counter_refs=none`은 진술이 아니라는 문장(A17), 승인된 행마다 §10 세 필드를 순서대로 싣고 이어서 `counter_refs=`를 싣는 한 줄 자기 결정 기록(A18), `ladder_path`의 게이트 행 슬러그 뒤 `;batch=<id>`와 `<id>`의 형식 `YYYYMMDDTHHMMSSZ`·한 요약 안 동일·결정 보드에 같은 값이 있으면 다음 빈 초(A19), 기록 없는 행은 미승인(A20), 기록 직전 REQ-BGS-011 네 조건 전부의 재확인(A21)과 REQ-BGS-010 분류(운영자 hold 포함)의 재확인(A43), 표류한 행은 승인으로 기록하지 않고 거부하며 운영자에게 알림(A22)을 담는다.
- Verify: AC-002와 같은 명령(A14–A22·A43 하위 테스트).
- 변이 탐침: 판정과 해시만 다시 읽는 문단(1회차까지의 읽기)은 A21 위반, 분류 재확인을 뺀 문단은 A43 위반, `searched=` 없이 `counter_refs=none`만 쓴 문단이나 값에 공백을 허용하는 문단은 A16 위반, 같은 `batch=` 값이 이미 있어도 그대로 쓰는 문단은 A19 위반으로 거부된다.

### AC-004 — 예약 행과 contract 모드 제외 (release-blocking, RED: RED-2)

- **Given** 행이 keep-set 범주이거나 리더 세션이 쥐는 권한이거나, 모드가 `contract`이다.
- **When** 정본 절의 REQ-BGS-010·012 규칙을 점검한다.
- **Then** 정본 절은 REQ-BGS-010이 나열한 keep-set 세 범주를 각각 — 환경상 불가능(A23), 운영자 보유(A24), 외부 공유 시스템의 되돌릴 수 없는 조작(A25) — 단일 승인에서 제외한다고 명시하고, 같은 요구사항이 나열한 리더 보유 권한 여섯 항목을 모두 제외 목록에 담되 "operator gates" 항목은 REQ-BGS-001이 요약 모집단으로 정한 운영자 형태 plan→run Kickoff 행을 포함하지 않는다고 한정하고, keep-set 범주에 걸리지 않는 그 행은 요약 행이며 REQ-BGS-011로 분류된다고 쓰며(A26), contract 모드에서는 서명이 plan→run 게이트이고 요약 행이 생기지 않는다고 명시한다(A27). 정본 절에 "user-facing behavior changes"가 없다.
- Verify: AC-002와 같은 명령(A23–A27 하위 테스트).
- Verify: `grep -rcF "user-facing behavior" .claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` → 두 파일 모두 `:0`(출처가 출하된 문서로 한정됐는지의 부재 점검; 양성 대조 불가 — 부재가 정답이다).
- 변이 탐침: 여섯 항목 중 하나를 뺀 목록은 A26 위반으로 거부된다. "operator gates"를 한정 없이 나열해 운영자 형태 Kickoff 행까지 예약하는 목록(요약 모집단이 빈다)도 한정 문장이 없으므로 A26 위반으로 거부된다.

### AC-005 — 차단 행은 승인에서 빠진다: 양성 규칙과 차단 토큰 (release-blocking, RED: RED-2)

- **Given** 행의 최신(최종 반복) 독립 plan-audit 판정이 PASS가 아니거나(plan-auditor가 낸 판정이 아니라 계획 산출물을 쓴 세션의 자기 진술인 경우 포함), 계획 단계가 audit-ready를 기록하지 않았거나, 판정 뒤 계획 산출물 해시가 바뀌었거나, 열린 차단이 있다.
- **When** 정본 절의 REQ-BGS-011 규칙을 점검한다.
- **Then** 정본 절은 (a) 판정 참조가 현재 계획 산출물의 최종 반복 판정에 결속된다고 쓰고(A28) 그 판정은 plan-auditor가 낸 독립 판정이며 작성 세션의 자기 진술 PASS는 차단 상태라고 쓰고(A44), (b) 승인 가능 조건이 PASS·audit-ready 기록·해시 불변·열린 차단 없음 *넷 모두*라고 양성형으로 쓰고 그 밖의 행은 차단으로 보고되어 승인에서 빠진다고 쓰며(A29 — 보고 위치가 표 안인지 밖인지는 Q3), (c) PASS-WITH-DEBT(A30), BYPASSED(A31), FAIL(A32), INCONCLUSIVE(A33), 부재한 판정(A34)을 각각 차단 상태로 명명하고, (d) audit-ready 미기록(A35), 해시 변경(A36), 열린 차단(A37)을 각각 차단으로 명시한다.
- Verify: AC-002와 같은 명령(A28–A37·A44 하위 테스트 — 앵커 하나당 변이 하나).
- 변이 탐침(`verification-completeness.md` §2): 부정 열거형 문단(FAIL·INCONCLUSIVE·부재만 차단)은 A30·A31 위반으로 거부된다. 나머지 다섯을 차단하고 *PASS-WITH-DEBT만* 통과시키는 문단은 A30, *BYPASSED만* 통과시키는 문단은 A31, 각각 한 토큰만 풀어 둔 문단은 A32·A33·A34 위반으로 거부된다. 이전 반복의 PASS를 인용해도 된다고 허용하는 문단은 A28 위반이다. 계획 산출물을 쓴 세션의 자기 진술 PASS를 판정으로 인정하는 문단은 A44 위반이다. 차단 행을 보고 없이 조용히 빼는 문단은 A29 위반이다. 통과할 수 있는 변이: 앵커를 둔 채 뒤에 모순 문장(예외 조항)을 두는 본문(어휘적 한계, G-4) — AC-011이 감사 읽기로 막는다.

### AC-006 — 구성원 규칙: Kickoff 행 한정, 붙잡지 않기, 동일 판정 1회 (release-blocking, RED: RED-2)

- **Given** 카드가 Kickoff가 아닌 게이트 행(예: sync 차단 승인)에서 기다리거나, 일부만 준비됐거나, 한 판정이 여러 카드를 막고 그중 일부가 차단 행이거나 예약 행이다.
- **When** 정본 절의 REQ-BGS-001·002·003 규칙을 점검한다.
- **Then** 정본 절은 요약이 plan→run Kickoff 행에만 적용된다고 쓰고(A01), 나머지 게이트 행 — factory decide 행들, sync 차단 승인, 카드 선택 — 은 요약 행이 아니고 운영자 답이 필요하면 개별 질문이라고 쓰고(A39), 레인은 교차 카드 배치를 만들지 않으며(A02), 준비된 카드를 기다리게 하지 않고(A03), 동일 판정 주체는 영향 카드를 전부 적어 한 번 묻되(A04) 그 질문은 차단 행과 예약 행을 이름에 올리지 않는다(A40)고 명시한다.
- Verify: AC-002와 같은 명령(A01–A04·A39·A40 하위 테스트).
- 변이 탐침: "인벤토리의 어떤 게이트 행이든 묶는다"고 쓴 본문(1회차까지의 읽기)은 A01 위반, 요약에 sync 차단 승인 행을 받아들이는 본문(sync-auditor가 FAIL을 낸 행이 계획 감사 네 조건만으로 통과한다)은 A39 위반, 차단 행(감사 FAIL 카드)을 공유 판정 질문에 이름 올리는 본문은 A40 위반으로 거부된다.

### AC-007 — 가드 테스트는 반증 가능하고 비공허하다: 변이 본문은 거부된다 (release-blocking, RED: RED-2)

- **Given** `plan.md` M1의 앵커 표(45행: A01–A45)와 실제 절.
- **When** 앵커를 하나씩 뺀 변이 본문 각각을 같은 검사 함수에 넣는다.
- **Then** 실제 절은 위반 0이고, 변이마다 *해당* 앵커 하나가 위반으로 보고된다. A38은 정본 절에 결과 수치("176"류)가 없다는 부재 앵커다. 하위 테스트 `--- PASS` 줄 수는 앵커 표 행 수 + 1 = 46 이상이다. 하위 테스트는 비어 있지 않다: 변이마다 `mutant <ID> rejected: reported=<ID>`(두 ID가 같다) 줄을, 실제 절은 `real_section violations=0` 줄을 출력에 낸다.
- Verify: `go test -count=1 -v -run '^TestBatchGateSummaryDoctrine$' ./internal/template/` — 출력을 파일로 돌려 계수한다(`agent-common-protocol.md` file-redirect). 통과 판독은 다섯이 모두 성립할 때만이다. (1) `--- PASS` 하위 테스트 줄 46개 이상, `--- FAIL` 0줄. (2) `=== RUN` 줄 수가 `--- PASS` 줄 수와 같다(최상위 테스트 줄을 포함해 양쪽을 같은 방식으로 센다. 시작하고 끝나지 않은 하위 테스트가 없다). (3) 출력에 `[no tests to run]`이 없다. (4) `mutant <ID> rejected: reported=<ID>` 꼴 줄이 45개이고 각 줄의 두 ID가 같다. (5) `real_section violations=0` 줄이 하나다.
- 변이 탐침: 빈 `t.Run` 45개와 자명한 `real_section`만 가진 테스트 파일은 (1)–(3)을 통과하지만 (4)에서 0줄이라 거부된다. 앵커 문구를 *모순 문장과 함께* 남긴 본문은 이 점검을 통과할 수 있다(어휘적 한계, G-4) — AC-011~013·017이 감사 읽기로 막는다. 검사 없이 줄만 찍는 하위 테스트도 출력으로는 구별되지 않는다(G-4) — sync 감사가 테스트 소스를 읽는다.

### AC-008 — 기준선이 변경보다 앞선 별도 커밋으로, 추적되는 경로에 남는다 (release-blocking, RED: RED-3)

- **Given** run-phase가 시작되지 않았다.
- **When** M0가 기준선 산출물을 기록하고, 이후 M2가 정본 절 커밋을 만든다.
- **Then** 두 시점으로 나뉜다. 판정에 필요한 산출물이 다르므로 한 시점의 점검을 다른 시점에 요구하지 않는다(MP-9).
  - **M0 종료 시점**(정본 절 커밋 D가 없어도 판정된다): `.moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`가 존재하고, 무시되지 않고, 줄 머리 레이블 `Command:`, `Observed output:`, `Classification method:`, `Limits:`를 같은 줄에 비어 있지 않은 본문과 함께 싣고, 그 파일을 처음 추가한 커밋 B는 그 파일 하나만 담는다.
  - **끝 점검 시점**(D가 있어야 판정된다; 완료 보고와 DoD에서 판정하며 M0 종료 조건이 아니다): 그 파일을 건드린 커밋은 B 하나뿐이고, 정본 절 제목을 들이거나 바꾼 커밋은 D 하나뿐이며, B는 D의 조상이고 B와 D는 서로 다르다.
- M0 Exit: 기준선 파일이 존재하고, 무시되지 않으며, 네 요소 레이블 줄을 모두 싣고, 기준선 커밋은 그 파일 하나만 담는다.
- Verify — M0 종료 점검 (`plan.md` M0의 Exit와 같은 문장):
  - V1: `ls .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md` → 경로 출력, exit 0
  - V2: `git check-ignore -v .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md` → 출력 없음, exit 1 (추적 가능한 위치)
  - V3: 네 줄 — `grep -c "^Command: [^ ]" <기준선 경로>`, `grep -c "^Observed output: [^ ]" <기준선 경로>`, `grep -c "^Classification method: [^ ]" <기준선 경로>`, `grep -c "^Limits: [^ ]" <기준선 경로>` → 각각 1 이상 (레이블 뒤 같은 줄에 공백이 아닌 글자가 하나 이상 있어야 센다)
  - V4: `git log --format=%h --diff-filter=A -- <기준선 경로>` → 정확히 한 줄 = B
  - V5: `git diff-tree --no-commit-id --name-only -r <B>` → 정확히 한 줄 = 기준선 경로 (같은 커밋에 다른 파일이 섞였으면 줄이 더 나온다)
- Verify — 끝 점검 (D가 필요):
  - V6: `git log --format=%h -- <기준선 경로>` → 정확히 한 줄 = B. B 뒤에 이 파일을 다시 쓴 커밋이 없다 — 곧 정본 절 커밋 D의 시점에도 HEAD의 시점에도 파일이 B의 내용 그대로다.
  - V7: `git log --format=%h -S"### 9.2 The batch gate summary" -- internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` → 정확히 한 줄 = D (이 제목의 출현 수를 바꾼 커밋. 두 줄 이상이면 제목이 지워졌다 다시 쓰인 것이므로 실패로 보고하고 D를 명시한다)
  - V8: `git merge-base --is-ancestor <B> <D>` → exit 0. 이 조상 판정이 순서의 증인이다(VCI §2.3 — 커밋 메시지 주장이 아니라 그래프). 같은 커밋도 자기 자신의 조상이므로 B와 D가 같으면 이 판정만으로는 부족하다 — V5가 B를 기준선 파일 하나로 묶으므로 D가 B와 같을 수 없다(D는 정본 절 파일을 담는다).
- 변이 탐침(`verification-completeness.md` §2): (m1) 빈 파일이나 한 줄짜리 파일을 단독 커밋한 브랜치는 V1·V2·V4·V5를 통과하지만 V3에서 거부된다. (m2) 레이블만 있고 본문이 비어 있는 파일("Command: " 뒤가 빔)은 V3에서 거부된다. (m3) 기준선과 정본 절을 한 커밋에 담은 브랜치는 V5가 둘 이상의 경로를 내서 거부된다. (m4) 정본 절 커밋 뒤에 기준선 파일을 다시 쓴 브랜치는 V6이 두 줄을 내서 거부된다. (m5) 기준선 커밋이 정본 절 커밋보다 나중인 브랜치는 V8이 exit 1이라 거부된다. (m6) 기준선을 `.moai/reports/t1344/`에 둔 브랜치는 V1(경로 부재)과 V2(exit 0)에서 거부된다. (m7) 정본 절 제목을 지웠다 다시 쓴 브랜치는 V7이 두 줄을 내서 거부된다. 통과할 수 있는 변이: 레이블마다 의미 없는 한 글자 본문(어휘적 한계, `spec.md` G-10) — sync 감사가 파일을 읽어 판정한다.
- `ComputeHash` 영향: 이 파일은 `planArtifactNames`(`internal/runtime/audit_cache.go:89-96`) 밖이라 추가해도 캐시된 감사 판정이 무효가 되지 않는다(`research.md` R4.3).

### AC-009 — Go 변경이 허용 목록 안에 있다 (release-blocking, RED: RED-4)

- **Given** 브랜치가 `develop`에서 분기한 상태이다.
- **When** run-phase가 끝난다.
- **Then** `develop`과의 merge-base 이후 바뀐 `.go` 파일은 정확히 일곱 개이고 전부 허용 목록 안에 있다. 제품 Go 넷: `internal/hook/session_start_factory.go`, `internal/hook/session_start_factory_i18n.go`, `internal/hook/session_start_kanban.go`, `internal/hook/session_start_kanban_i18n.go`. 테스트 Go 셋: `internal/hook/session_start_kanban_i18n_test.go`, `internal/hook/session_start_leader_gate_notice_test.go`, `internal/template/batch_gate_summary_doctrine_test.go`. 이 목록 밖의 `.go` 파일이 하나라도 있으면 실패다.
- Verify: `git diff --name-only develop...HEAD -- '*.go'` → 정확히 위 일곱 줄(경로 바이트 순서: 훅 다섯 줄 `session_start_factory.go`, `session_start_factory_i18n.go`, `session_start_kanban.go`, `session_start_kanban_i18n.go`, `session_start_kanban_i18n_test.go`, 이어서 `session_start_leader_gate_notice_test.go`, 마지막에 `internal/template/batch_gate_summary_doctrine_test.go`)
- 읽기 단계: 제품 Go 네 파일의 변경이 리더 공지 문장 필드와 그것을 기존 블록에 합치는 한 줄에 한정되는지는 `git diff --numstat develop...HEAD -- internal/hook/session_start_factory.go internal/hook/session_start_factory_i18n.go internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go` 출력을 sync 감사가 읽어 판정한다(기계 점검 없음, `spec.md` G-8).
- 평가는 병합 전에만 유효하다. 카드가 `develop`에 병합되면 merge-base가 카드 tip이 되어 범위가 비고 공허하게 통과한다(`gitflow-lane-protocol.md` §8의 한계). 병합 뒤 근거는 병합 트리와 카드 브랜치 트리의 동일성이다.
- 같은 목적의 파일 수 계수(Tier 재계수, `spec.md` §A.5)는 `.go` 외 파일까지 센다. 이 기준은 `.go` 목록만 본다.

### AC-010 — 고정 가드와 예산이 GREEN을 유지한다 (regression-guard)

- **Given** 변경 전 트리에서 아래 가드가 전부 GREEN이다(§Evidence Ledger GREEN 기준선).
- **When** run-phase가 정본 절과 포인터를 라이브·미러에 반영한다.
- **Then** 가드가 전부 계속 `--- PASS`이고 `TestAlwaysLoadedTokenBudget`의 headroom이 0 이상이며 `run.md`는 199줄을 넘지 않는다. 리더 공지 쪽 기존 고정 테스트 63개도 그대로 `--- PASS`이고 `--- FAIL`은 없다. (`kanban-dispatch.md`·`run.md`의 바이트·줄 증가 한도는 AC-016이 가진다.)
- Verify: `go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/`
- Verify: `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/`
- Verify: `go test -count=1 -v -run '^(TestSubSkillLOCCeiling|TestEntryRouterLOCCeiling)$' ./internal/skills/`
- Verify: `go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestSpecAssembly_RewrittenToCLIPath|TestSpecAssembly_NoNewInternalTokens)$' ./internal/cli/`
- Verify: `go test -count=1 -v -run '^(TestImplementationKickoffApprovalPreservedBeforeGoal|TestTemplateNoInternalContentLeak|TestContractModeEmitterSites)$' ./internal/template/`
- Verify: `wc -l .claude/skills/moai/workflows/run.md` → `199 …` 이하
- Verify: 리더 공지 고정 테스트 — `research.md` R3.1 E17의 23개 이름 명령 한 줄 그대로. 읽는 법: `--- PASS` 63줄 이상, `--- FAIL` 0줄, `=== RUN`과 같은 수(비어 있지 않은 하위 테스트 집합). 새 테스트는 AC-015가 따로 다룬다.

### AC-011 — 시나리오: 혼합 배치 (regression-guard, 감사 읽기)

- **Given** 같은 Kickoff 게이트 행에서 카드 A(독립 감사 PASS, 반대 증거 없음), 카드 B(PASS이고 미해결 decision-index 행 하나), 카드 C(감사 FAIL), 카드 D(감사 PASS-WITH-DEBT)가 대기한다.
- **When** 리더가 배치 게이트 요약을 만든다.
- **Then** 보고서에서 B 행이 A 행보다 먼저 나오고 B의 `counter_refs=`가 그 decision-index 행을 가리키며, A는 `counter_refs=none searched=<공백-없는-토큰>`이고, A·B 행은 keep-set·리더 보유 권한 점검 결과와 근거를 싣고, C와 D는 차단으로 보고되어(요약 표 안이든 표 밖이든 — Q3의 열린 부분) 승인에서 빠지고, 질문은 하나이며 승인은 A·B에만 미친다.
- 변형: B의 티어·모드 선호·PR 전략·체인 범위가 카드·SPEC 계약에도 없고 디스크에도 없으면 B는 요약으로 승인된 뒤에도 run-phase에 들어가지 못하고 B의 선호 질문이 따로 나간다(REQ-BGS-020).
- Verify: sync 감사가 정본 절을 읽고 위 다섯 가지와 변형이 절의 규칙으로 도출됨을 확인한다(기계 점검 없음, §D.1).

### AC-012 — 시나리오: 빼내기와 읽기 규칙 (regression-guard, 감사 읽기)

- **Given** A·B·E 세 승인 가능 행과 차단 행 C가 있는 요약에 단일 승인 질문이 나갔고, 그 질문의 명시적 선택지는 질문 채널의 선택지 상한 안에 있으며 질문 문구가 읽기 규칙을 적고 있다.
- **When** 운영자가 자유 입력으로 B의 카드 id를 적는다.
- **Then** A·E는 같은 `batch=` 식별자와 각자의 `counter_refs=`를 싣는 자기 §10 결정 기록을 남기며 승인되고, B는 개별 질문으로 남고, C는 애초에 승인에 들어 있지 않으며, 기록 없는 행은 승인으로 치지 않는다.
- 변형: 운영자가 "괜찮아 보여요"처럼 카드 id로 읽을 수 없는 글을 적거나 보고서에 없는 카드 id를 적으면 어느 행도 승인되지 않고 같은 질문이 다시 나간다.
- Verify: sync 감사가 REQ-BGS-006·009에서 도출되는지 읽는다.

### AC-013 — 시나리오: 붙잡지 않기와 동일 판정 1회 (regression-guard, 감사 읽기)

- **Given** 카드 D는 준비됐고 카드 E는 아직 감사 중이다. 별도로 하나의 게이트 설정이 카드 F·G·H·I를 막고 있다. F·G·H는 차단도 예약도 아니고 I는 감사 FAIL(차단 행)이다.
- **When** 리더가 운영자 질문을 구성한다.
- **Then** D는 E를 기다리지 않고 제시되며 E는 다음 요약이나 개별 질문으로 간다. F·G·H는 같은 판정 질문 하나에 세 카드를 모두 적어 한 번 묻고, I는 그 질문에 이름이 오르지 않는다.
- Verify: sync 감사가 REQ-BGS-002·003에서 도출되는지 읽는다.

### AC-014 — 낡은 표현 정합 (release-blocking, 무조건 — Q4 범위 안, RED: RED-5)

- **Given** `spec-assembly.md:202-208`에 "stays MANDATORY and score-independent"와 "does NOT substitute for the gate"가, `moai.md:144,240`에 카드별 필수 운영자 Kickoff를 말하는 "Score-independent … never bypasses it"와 "score-independent"가 있고, 두 문서 어디도 `auto-semantics.md` §9.1·§9.2를 인용하지 않는다(RED-5).
- **When** run-phase M4가 그 줄들을 §9.1·§9.2에 맞춰 다시 쓴다.
- **Then** 라이브와 미러 두 사본 모두에서 위 낡은 문구가 없고, `§9.1`과 `§9.2`가 각각 한 번 이상 인용되며, 보존 문구(`[HARD] The Implementation Kickoff Approval`, `moai plan render-html`, `Fail-open`)와 줄 수(`spec-assembly.md` 597, `moai.md` 라이브 284 / 미러 282)와 라이브↔미러 사전 차이 헌크 넷이 그대로다. 이 기준은 두 파일만 덮는다(REQ-BGS-014).
- Verify (아래 grep 다섯 줄은 각각 라이브와 미러 경로에 같은 명령으로 두 번씩 돌린다):
  - `grep -c "stays MANDATORY" .claude/skills/moai/workflows/plan/spec-assembly.md` → `0`
  - `grep -c "does NOT substitute for the gate" .claude/skills/moai/workflows/plan/spec-assembly.md` → `0`
  - `grep -c "never bypasses it" .claude/skills/moai/workflows/moai.md` → `0`
  - `grep -ci "score-independent" .claude/skills/moai/workflows/moai.md .claude/skills/moai/workflows/plan/spec-assembly.md` → 두 파일 모두 `0`
  - `grep -c "§9.2" .claude/skills/moai/workflows/moai.md .claude/skills/moai/workflows/plan/spec-assembly.md` → 두 파일 모두 1 이상(같은 방식으로 `§9.1`도 1 이상)
- Verify: `wc -l .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md` → `597`, `284`, `282`
- Verify: `diff .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md` → exit 1이고 헌크 헤더가 정확히 `215c215`, `245c245`, `253d252`, `283,284c282` 넷
- Verify: `go test -count=1 -v -run '^(TestSpecAssembly_RewrittenToCLIPath|TestSpecAssembly_NoNewInternalTokens)$' ./internal/cli/` → 두 `--- PASS`
- 변이 탐침: 낡은 구절 하나만 지우고 같은 주장을 말만 바꿔 유지하는 변이(예: "always mandatory")는 위 grep 일부를 통과할 수 있다(어휘적 한계). `§9.1`·`§9.2` 인용 요구가 포인터 누락 변이를 막고, 사본 한쪽만 고친 변이는 `diff` 헌크 헤더와 두 경로 grep이 막는다. 말만 바꾼 변이는 sync 감사가 다시 쓴 문구를 §9.1에 대조해 읽어 막는다(기계 점검 없음).

### AC-015 — 리더 공지 문장이 두 리더 공지의 영어 에이전트용 사본과 네 로케일 운영자용 사본에 있다 (release-blocking, RED: RED-6)

- **Given** 팩토리 리더 공지와 칸반 리더 공지는 지금 배치 게이트 요약을 언급하지 않는다(`research.md` E8, RED-6).
- **When** run-phase M5가 두 리더 공지에 문장 하나를 en·ko·ja·zh로 더한다.
- **Then** (a) 두 리더 공지 × 네 로케일 여덟 조합 모두에서 렌더된 공지가 포인터 토큰 `.claude/rules/moai/workflow/auto-semantics.md` §9.2와 이름 토큰 `batch gate summary`를 그대로 담는다(REQ-BGS-016·017; 영어 렌더가 곧 에이전트용 사본이다. 두 토큰은 번역하지 않는 주소 취급이다). (b) 그 문장은 질문 도구 이름, `SPEC-`, 카드 id 꼴(`t` 뒤 숫자 3–5자리), ISO 날짜, `moai todo`, 단어 `lead`, `리드`, `epic`을 담지 않고 정본의 서식·제외·반대 증거 규칙을 되풀이하지 않는다(REQ-BGS-016). (c) 레인·동반 공지에는 포인터가 없다(REQ-BGS-017). (d) 알림 소스 네 파일의 주석 제외 `AskUserQuestion` 토큰 수가 0이다. (e) 기존 고정 테스트 63개가 그대로 `--- PASS`다(REQ-BGS-018, AC-010). (f) 변이 본문 — 한 로케일 누락, 질문 도구 이름 포함, 포인터 누락, 이름 토큰 누락, 카드 id 포함, 한쪽 리더 공지만 보유 — 은 각각 점검 함수에서 거부된다. (g) SessionStart 핸들러 수준에서 칸반 리더와 팩토리 리더 각각, 에이전트용 영어 사본(`additionalContext`)과 운영자용 현지어 사본(`systemMessage`, ko 한 경로 이상) 모두에 문장이 실린다.
- Verify: `go test -count=1 -v -run '^TestLeaderNoticeBatchGatePointer$' ./internal/hook/` — 출력을 파일로 돌려 계수한다. 통과 판독은 넷이 모두 성립할 때만이다. (1) `--- PASS` 줄 17개 이상(여덟 조합 + 변이 여섯 이상 + 레인·동반 부재·소스 스캔·핸들러 수준 각 1 이상)이고 `--- FAIL` 0줄. (2) `=== RUN` 줄 수가 `--- PASS` 줄 수와 같다(최상위 테스트 줄을 포함해 양쪽을 같은 방식으로 센다). (3) `[no tests to run]`이 없다 — 빈 selector의 `ok … [no tests to run]`은 통과가 아니다(RED-6에 그 출력이 있다). (4) 변이마다 `mutant <이름> rejected: reported=<이름>` 줄(두 이름이 같다)이 6개 이상이다.
- Verify: `grep -c AskUserQuestion internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory.go internal/hook/session_start_factory_i18n.go` → 네 파일 모두 `:0`(exit 1, 줄 순서는 인자 순서와 다를 수 있다)
- Verify: `` grep -cF 'auto-semantics.md` §9.2' internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory_i18n.go `` → 두 파일 모두 `:4`(로케일 값 하나씩. 주석에 같은 토큰을 되풀이하지 않는다)
- Verify: `` grep -cF 'batch gate summary' internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory_i18n.go `` → 두 파일 모두 `:4`(로케일 값마다 한 줄에 한 번. 주석에 되풀이하지 않는다)
- 변이 탐침(`verification-completeness.md` §2): 한 로케일의 문장을 지운 공지는 (a)와 두 grep `:4`에서 실패하고, 질문 도구 이름을 넣은 문장은 (b)·(d)에서 실패하고, 포인터를 뺀 문장은 (a)와 포인터 grep에서 실패하고, 이름 토큰만 뺀 문장(포인터 주소만 남은 문장)은 (a)와 이름 grep에서 실패하고, 한쪽 리더 공지에만 문장을 둔 구현은 (a)에서 실패한다. 빈 하위 테스트 열일곱 개만 둔 테스트 파일은 (1)–(3)을 통과하지만 (4)에서 0줄이라 거부된다. 통과할 수 있는 변이: 포인터를 둔 채 뜻이 정본과 반대인 문장(어휘적 한계, `spec.md` G-7) — sync 감사가 네 로케일 문장을 읽어 막는다.

### AC-016 — 정본은 한 곳에만 있고 나머지는 포인터만 지니며 크기 한도 안이다 (release-blocking, RED: RED-7)

- **Given** 정본 절 본문 어휘 `counter_refs=`·`searched=`는 지금 어느 규칙·스킬·훅 소스에도 없다(RED-7).
- **When** run-phase가 정본 절을 두 사본의 `auto-semantics.md`에 쓰고 `kanban-dispatch.md`·`run.md`의 라이브·미러 네 곳에 포인터를 더한다.
- **Then** (a) 정본 전용 어휘 두 토큰은 정본 두 사본(라이브·미러 `auto-semantics.md`)에만 나타난다 — 규칙·스킬·훅 소스의 다른 어떤 비테스트 파일에도 없다. (b) 포인터 네 곳 각각에 `§9.2`가 한 번 이상 나타난다. (c) `kanban-dispatch.md` 라이브·미러의 바이트 순증가가 각각 1,000 미만이다. (d) `git diff --numstat develop...HEAD`에서 `run.md` 두 사본은 추가 줄 수와 삭제 줄 수가 같고(줄이 늘지 않음), `kanban-dispatch.md` 두 사본은 추가 줄 수가 3 이하다. (c)(d)의 왼쪽 끝은 읽는 시점의 `git merge-base develop HEAD`다 — 리터럴 base SHA로 재지 않는다(`gitflow-lane-protocol.md` §8: 흡수하는 순간 리터럴 핀 범위에 다른 카드의 커밋이 들어온다).
- Verify (a): `grep -rlF "counter_refs=" .claude/rules .claude/skills internal/template/templates/.claude/rules internal/template/templates/.claude/skills internal/hook --exclude='*_test.go'` → 정확히 두 줄(`.claude/rules/moai/workflow/auto-semantics.md`, `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`)
- Verify (a): 같은 명령의 토큰만 `searched=`로 바꿔 → 같은 정확히 두 줄
- Verify (b): `grep -c "§9.2" .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md` → 네 파일 모두 1 이상
- Verify (c): 먼저 `git merge-base develop HEAD` → 40자 SHA 한 줄, exit 0. 읽는 시점마다 다시 구하고 문서에 핀하지 않는다(이 값을 M이라 부른다). 이어서 `git cat-file -s M:.claude/rules/moai/workflow/kanban-dispatch.md`(M은 앞 출력)와 `wc -c .claude/rules/moai/workflow/kanban-dispatch.md`의 차 < 1000; 미러는 `git cat-file -s M:internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`와 `wc -c internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`의 차 < 1000
- Verify (d): `git diff --numstat develop...HEAD -- .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` → 각 줄의 첫 두 수(추가·삭제)가 위 (d)를 만족. 세 점 형태는 읽는 시점의 merge-base를 왼쪽 끝으로 쓴다.
- 범위 비공허 대조: `git diff --name-only develop...HEAD` → 한 줄 이상. 0줄이면 (c)(d)의 판정은 PASS가 아니라 '측정 불가'다.
- 한계: 병합 전에만 유효하다 — 카드가 `develop`에 병합되면 merge-base가 카드 tip이 되어 범위가 비고 (c)(d)가 공허하게 통과한다. 위 대조가 이를 '측정 불가'로 만든다. 병합 뒤 근거는 병합 트리와 카드 브랜치 트리의 동일성이다(`gitflow-lane-protocol.md` §8). 계획 시점 기준값(26807/26485바이트)은 `research.md` R5에만 둔다.
- 변이 탐침(`verification-completeness.md` §2): 정본 §9.2 본문 전체를 1,000바이트 미만으로 줄여 `kanban-dispatch.md`에 붙여 넣은 변이는 (c)(d)를 통과하지만 (a)에서 세 번째 파일이 나와 거부된다. 포인터만 있는 1,500바이트짜리 군더더기 문단을 더한 변이는 (a)(b)를 통과하지만 (c)에서 거부된다. 줄을 새로 더한 `run.md` 변이는 (d)에서 거부된다. 다른 카드의 `kanban-dispatch.md` 편집을 흡수한 브랜치에서 리터럴 SHA를 왼쪽 끝으로 쓴 측정은 다른 카드의 바이트를 이 카드의 증가로 센다 — 읽는 시점의 merge-base는 세지 않는다. 병합된(범위가 빈) 상태에서 (c)(d)를 통과로 읽는 판정은 범위 비공허 대조에서 '측정 불가'로 거부된다. 통과할 수 있는 변이: 두 토큰을 피해 정본의 뜻을 풀어 쓴 문단이 (c)(d)의 한도 안에 드는 경우(어휘적 한계, `spec.md` G-9) — sync 감사가 편집 헌크를 읽어 막는다.

### AC-017 — 시나리오: 응답과 기록 사이에 표류한 행 (regression-guard, 감사 읽기)

- **Given** 단일 승인 질문이 A·B·E 세 승인 가능 행에 대해 나갔고, 운영자가 답하기 전에 A의 계획 산출물이 바뀌어 판정 이후 해시가 달라졌다.
- **When** 운영자의 승인 답이 도착해 리더가 결정 기록을 쓰기 시작한다.
- **Then** 리더는 행마다 기록 직전에 최신 판정과 해시를 다시 읽고, A는 더 이상 REQ-BGS-011을 만족하지 않으므로 승인으로 기록하지 않고 거부해 운영자에게 알리며, B·E는 같은 `batch=` 식별자의 자기 기록으로 승인된다.
- Verify: sync 감사가 REQ-BGS-019에서 도출되는지 읽고, A의 기록 부재와 거부 보고를 확인한다(기계 점검 없음, §D.1).

## §Evidence Ledger (RED-now 관측, 트리 `72e09d27b`, 브랜치 `WT-batch-approval-gate`)

각 항목은 명령(단일 호출) · 원문 stdout · 종료 코드 · 빨간 이유를 가진다. 종료 코드는 별도 필드로 적었다(단일 호출 형식에서는 `; echo $?`를 못 쓰므로). 0.3.0에서 이 에이전트가 전부 다시 실행했다(종료 코드는 관측용으로 `; echo "exit=$?"`를 붙여 얻은 값).

### RED-1 — AC-001: 정본 절이 없다

- 명령 A: `grep -c "^### 9.2 The batch gate summary" .claude/rules/moai/workflow/auto-semantics.md`
  - stdout: `0` · exit: `1`
- 명령 B: `grep -c "^### 9.2 The batch gate summary" internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`
  - stdout: `0` · exit: `1`
- 빨간 이유: `auto-semantics.md`는 §9.1(`:173`)에서 끝나는 §9 뒤에 §10(`:183`)이 곧바로 이어진다. §9.2는 없다.
- 초록 경로: M2가 두 파일에 제목을 쓰면 둘 다 `1`.

### RED-2 — AC-002~007: 가드 테스트가 없다

- 명령: `ls internal/template/batch_gate_summary_doctrine_test.go`
  - stdout: `ls: internal/template/batch_gate_summary_doctrine_test.go: No such file or directory` · exit: `1`
- 빨간 이유: 테스트 파일이 없다. 존재하지 않는 테스트를 `-run`으로 고르면 "테스트 0개 실행, ok"가 되어 *공허하게 통과*하므로(`verification-completeness.md` §2.1 근거 사례) RED 칸은 `go test`가 아니라 파일 존재 여부로 잡았다.
- 초록 경로: M1이 파일을 만들고(정본 절이 없으므로 처음엔 `FAIL`), M2가 절을 써서 `--- PASS`로 뒤집는다. 통과 출력의 하위 테스트 집합이 비어 있지 않음을 AC-007이 요구한다.

### RED-3 — AC-008: 기준선 산출물이 없다 (0.3.0에서 정정)

- 명령 3a: `ls .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`
  - stdout: `ls: .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md: No such file or directory` · exit: `1`
- 명령 3b: `git check-ignore -v .moai/reports/t1344/baseline-gate-rounds.md`
  - stdout: `.gitignore:235:.moai/reports/*	.moai/reports/t1344/baseline-gate-rounds.md` · exit: `0` (옛 위치는 무시된다)
- 명령 3c: `git check-ignore -v .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`
  - stdout: (비어 있음) · exit: `1` (새 위치는 무시되지 않는다)
- 양성 대조: `git check-ignore -v .moai/reports/t1344/plan-audit-iter1.md` → `.gitignore:235:.moai/reports/*	.moai/reports/t1344/plan-audit-iter1.md` · exit `0`(무시 규칙이 이 형태로 발화한다). `git check-ignore -v .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/plan.md` → 비어 있음 · exit `1`(추적되는 SPEC 파일은 무시되지 않는다). `git ls-files .moai/reports/t1386 .moai/reports/t1381` → 비어 있음(최근 카드 보고서는 추적되지 않는다).
- 빨간 이유: 파일이 아직 없다(기준선 미측정, G-1). 0.2.0 칸이 든 "`.moai/reports/<card-id>/` 아래 산출물은 추적 대상이다" 전제는 거짓이었다: `.moai/reports/*`는 무시되고(`.gitignore:235`, 운영자 지시 2026-09-14) 추적되는 보고서 파일은 지시 이전에 강제 추가된 옛 항목뿐이다(`research.md` R4.1). 그 경로에서는 `git log --diff-filter=A`가 커밋을 내지 못하므로 AC-008의 조상 판정에 도달할 길이 없었다.
- 명령 3d (0.4.0 신설, 트리 `08692e732`; 네 레이블 각각 같은 형태): `grep -c "^Command: [^ ]" .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`
  - stdout: (비어 있음; stderr `ugrep: warning: .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md: No such file or directory`) · exit: `2`. 레이블 `Observed output:`, `Classification method:`, `Limits:`로 바꾼 같은 형태 셋도 같은 출력·같은 종료 코드(이 실행이 넷 모두 돌렸다).
  - 양성 대조(같은 `[^ ]` 형태): `grep -c "^## §A [^ ]" .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/plan.md` → `1` · exit `0`. 이 명령 형태는 존재하는 파일에서 발화한다 — 3d의 빈 출력은 파일 부재의 결과다.
- 명령 3e (트리 `08692e732`): `git log --format=%h -- .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`
  - stdout: (비어 있음) · exit: `0`. V6의 기대는 정확히 한 줄이므로 지금은 0줄이다.
- 명령 3f (트리 `08692e732`): `git log --format=%h -S"### 9.2 The batch gate summary" -- internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`
  - stdout: (비어 있음) · exit: `0`. V7의 기대는 정확히 한 줄이다.
  - 양성 대조: 같은 형태에 제목만 `### 9.1 The default-autonomous Kickoff transition`으로 바꾼 명령 → `147c25d77` 한 줄 · exit `0`. `-S` 형태가 발화하고 기존 제목의 출현 커밋이 한 줄임을 보인다(V7의 기대 기수 근거).
  - V5 형태의 양성 대조: `git diff-tree --no-commit-id --name-only -r 08692e732` → SPEC 디렉터리 아래 여덟 경로 · exit `0`(경로만 찍는 형태이며 머리 줄이 없다).
- 초록 경로: M0가 추적되는 SPEC 디렉터리 안에 파일을 만들고 단독 커밋하면 V1–V5가 통과한다(M0 종료). M2가 정본 절 커밋을 만든 뒤 끝 점검 V6–V8이 통과한다.

### RED-4 — AC-009: 허용 목록의 .go 변경 일곱 줄이 아직 없다

- 명령: `git diff --name-only develop...HEAD -- '*.go'`
  - stdout: (비어 있음) · exit: `0` (트리 `72e09d27b`에서 다시 쟀다. 이 트리에서 `git merge-base develop HEAD`는 `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`다)
- 빨간 이유: 기대값은 *정확히 일곱 줄*이다. 빈 출력은 일곱 파일이 아직 없다는 뜻이다. 빈 출력은 "Go가 안 바뀐" 상태와도 같으므로, 이 기준은 특정 일곱 줄이 생겼음을 요구해서 변경이 없는 빈 실행을 통과시키지 않는다.
- 초록 경로: M1이 템플릿 가드 테스트 한 줄을, M5가 제품 Go 넷과 훅 테스트 둘을 더해 일곱 줄이 된다. 허용 목록 밖의 `.go` 파일은 줄을 늘려 실패한다. 허용 파일 *안*의 범위 초과 변경은 이 명령이 못 잡는다 — AC-009 읽기 단계(`spec.md` G-8).

### RED-5 — AC-014: 낡은 표현이 아직 있고 §9.1·§9.2 인용이 없다

모두 트리 `72e09d27b`에서 이 실행이 직접 쟀다(라이브 사본. 미러 사본은 `spec-assembly.md`가 동일, `moai.md`의 편집 대상 줄 `:144`·`:240`이 동일하다 — `research.md` E15).

- 명령 5a: `grep -c "stays MANDATORY" .claude/skills/moai/workflows/plan/spec-assembly.md`
  - stdout: `1` · exit: `0`
- 명령 5b: `grep -c "does NOT substitute for the gate" .claude/skills/moai/workflows/plan/spec-assembly.md`
  - stdout: `1` · exit: `0`
- 명령 5c: `grep -c "never bypasses it" .claude/skills/moai/workflows/moai.md`
  - stdout: `1` · exit: `0`
- 명령 5d: `grep -ci "score-independent" .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md .claude/skills/moai/workflows/plan/spec-assembly.md`
  - stdout(줄 순서는 인자 순서와 다르게 찍혔다): `.claude/skills/moai/workflows/moai.md:2` · `.claude/skills/moai/workflows/plan/spec-assembly.md:1` · `internal/template/templates/.claude/skills/moai/workflows/moai.md:2` · exit: `0`
- 명령 5e: `grep -c "§9.2" .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/moai.md`
  - stdout: `.claude/skills/moai/workflows/moai.md:0` · `.claude/skills/moai/workflows/plan/spec-assembly.md:0` · exit: `1`
- 명령 5f: `grep -c "§9.1" .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/moai.md`
  - stdout: `.claude/skills/moai/workflows/moai.md:0` · `.claude/skills/moai/workflows/plan/spec-assembly.md:0` · exit: `1`
- 빨간 이유: 5a–5d는 기대값 `0`인데 낡은 문구가 각각 있고, 5e·5f는 기대값 1 이상인데 두 문서가 정본 절을 한 번도 인용하지 않는다.
- 초록 경로: M4가 정합하면 5a–5d는 `0`(grep -c는 0일 때 exit 1이므로 통과 판독은 stdout), 5e·5f는 두 문서 모두 1 이상.

### RED-6 — AC-015: 리더 공지 문장도 그 테스트도 없다

- 명령 6a: `ls internal/hook/session_start_leader_gate_notice_test.go`
  - stdout: `ls: internal/hook/session_start_leader_gate_notice_test.go: No such file or directory` · exit: `1`
- 명령 6b: `` grep -cF 'auto-semantics.md` §9.2' internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory_i18n.go ``
  - stdout: `internal/hook/session_start_kanban_i18n.go:0` · `internal/hook/session_start_factory_i18n.go:0` · exit: `1`
- 명령 6c: `go test -count=1 -run '^TestLeaderNoticeBatchGatePointer$' ./internal/hook/`
  - stdout: `ok  	github.com/modu-ai/moai-adk/internal/hook	0.937s [no tests to run]` · exit: `0`
- 명령 6d (0.4.0 신설, 트리 `08692e732`): `` grep -cF 'batch gate summary' internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory_i18n.go ``
  - stdout: `internal/hook/session_start_kanban_i18n.go:0` · `internal/hook/session_start_factory_i18n.go:0` · exit: `1`
  - 양성 대조(같은 `grep -cF` 형태, 토큰만 `moai`): `internal/hook/session_start_kanban_i18n.go:14` · `internal/hook/session_start_factory_i18n.go:45` · exit `0`. 이 형태는 두 파일에서 발화한다 — 6d의 0은 이름 토큰의 부재다.
- 빨간 이유: 6a·6b·6d가 RED 신호다 — 테스트 파일이 없고 두 로케일 표 어디에도 포인터 토큰과 이름 토큰이 없다. 6c는 *왜 RED 칸을 `go test`로 잡지 않았는가*의 증거다: 존재하지 않는 테스트를 `-run`으로 고르면 "테스트 0개 실행, ok, exit 0"이 되어 공허하게 통과한다(`verification-completeness.md` §2.1). 그래서 AC-015의 통과 판독은 `--- PASS` 줄이 17 이상이고 `=== RUN`과 같은 수이며 `[no tests to run]`이 없고 변이 거부 줄이 있을 때만이다.
- 초록 경로: M5가 테스트 파일을 먼저 만들고(포인터 부재로 `FAIL`을 관측), 필드와 네 로케일 값을 더해 `--- PASS`로 뒤집는다. 6b와 6d는 두 파일 각 `:4`가 된다.

### RED-7 — AC-016: 정본 어휘가 어디에도 없고 포인터도 없다 (0.3.0 신설)

- 명령 7a: `grep -rlF "counter_refs=" .claude/rules .claude/skills internal/template/templates/.claude/rules internal/template/templates/.claude/skills internal/hook --exclude='*_test.go'`
  - stdout: (비어 있음) · exit: `1`
- 명령 7a-대조: 같은 형태에 토큰만 `ladder_path=`로 바꾼 명령
  - stdout: `.claude/rules/moai/workflow/auto-semantics.md` · `.claude/skills/moai-lane-watchdog/SKILL.md` · `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` · `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md` · exit: `0` (이 명령 형태가 적중을 낸다 — 7a의 빈 출력은 부재다)
- 명령 7a2: 7a의 토큰만 `searched=`로 바꾼 명령
  - stdout: (비어 있음) · exit: `1`
- 명령 7b: `grep -c "§9.2" .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md`
  - stdout: 네 파일 모두 `:0` · exit: `1`
- 명령 7c (0.4.0 형태 변경, 트리 `08692e732`): `wc -c .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`
  - stdout: `26807 .claude/rules/moai/workflow/kanban-dispatch.md` · `26485 internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` · `53292 total` · exit: `0`. 읽는 시점의 `git merge-base develop HEAD`는 이 트리에서 `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`(= 로컬 `develop`)이고 그 시점의 바이트 수는 `research.md` R5의 계획 시점 측정(26807/26485)과 같다. 그래서 순증가는 0이다.
- 명령 7d (0.4.0 형태 변경, 트리 `08692e732`): `git diff --numstat develop...HEAD -- .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`
  - stdout: (비어 있음) · exit: `0`
  - 범위 비공허 대조: `git diff --name-only develop...HEAD` → SPEC 디렉터리 아래 여덟 경로(한 줄 이상) · exit `0`. 이 트리에서 범위는 비어 있지 않다 — 7d의 빈 출력은 "포인터 네 곳이 아직 안 바뀜"이지 "측정 불가"가 아니다.
- 빨간 이유: 7a·7a2는 기대값이 정확히 두 줄인데 0줄이고, 7b는 기대값이 1 이상인데 0이다. 7c·7d는 상한 절이라 지금은 증가가 0이어서 공허하게 만족한다 — 이 두 절은 RED 신호가 아니라 포인터만 넘치는 변이를 막는 상한이다.
- 초록 경로: M2가 정본 절을 쓰면 7a·7a2가 정본 두 사본의 두 줄이 되고, M3가 포인터를 더하면 7b가 네 파일 모두 1 이상이 되며 7c·7d는 한도 안에 머문다.

### GREEN 기준선 (regression-guard가 공허하지 않음을 보이는 변경 전 관측, 트리 `c50da9c2f`; 0.3.0에서 재실행하지 않았다 — 두 트리의 소스 파일이 같다)

| 가드 | 명령 | 관측 |
|---|---|---|
| 상시 로딩 예산 | `go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/` | `always-loaded surface = 64227 tokens (budget 77600, headroom 13373, 16 entries)` · `--- PASS` · exit 0 |
| 미러 동일성 | `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/` | `--- PASS` · 하위 테스트 9개 · exit 0 |
| LOC 한도 | `go test -count=1 -v -run '^(TestSubSkillLOCCeiling\|TestEntryRouterLOCCeiling)$' ./internal/skills/` | 두 테스트 `--- PASS` · exit 0 |
| 보존 구간·스펙 조립 | `go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment\|TestAutoRankMirrorParity\|TestSpecAssembly_RewrittenToCLIPath)$' ./internal/cli/` | 세 테스트 `--- PASS` · exit 0 |
| 리더 공지 고정 테스트 | `research.md` R3.1 E17의 23개 이름 명령(`./internal/hook/`) | `--- PASS` 63줄 · `--- FAIL` 0줄 · `=== RUN` 63줄 · `ok  github.com/modu-ai/moai-adk/internal/hook  6.364s` · exit 0 |
| 알림 소스 질문 도구 토큰 | `grep -c AskUserQuestion internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory.go internal/hook/session_start_factory_i18n.go` | 네 파일 모두 `:0` · exit 1(적중 없음) — 줄 순서는 인자 순서와 다르게 찍혔다 |

AC-010의 나머지 테스트(`TestImplementationKickoffApprovalPreservedBeforeGoal`, `TestTemplateNoInternalContentLeak`, `TestContractModeEmitterSites`, `TestSpecAssembly_NoNewInternalTokens`)의 변경 전 기준선은 이번 플랜에서 돌리지 않았다 — Gap. run-phase Pre-flight(`plan.md` §C)에서 먼저 측정한다.

## §D.1 Edge cases

- **행이 하나뿐**: 둘 이상일 때만 요약이다. 하나면 개별 질문이다(정본 절 적용 대상 문장).
- **모두 차단 행**: 승인 질문이 없다. 차단 사유를 보고하고 끝낸다.
- **질문 채널이 없는 하니스**: 같은 요약이 blocker 보고서로 나간다(`spec.md` G-6).
- **요약 직후 새 행 도착**: 이미 나간 승인은 그 행을 덮지 않는다(REQ-BGS-005).
- **PASS 기준 직상 점수**: 여유가 작은 행은 반대 증거로 올라가 앞쪽에 놓인다(REQ-BGS-004·007).
- **시나리오 기준(AC-011~013, AC-017)은 기계 점검이 불가능하다**: 감사가 정본 절을 읽고 도출 가능성을 판정한다. 이 한계는 §G-4에서 공시했다.
- **리더 공지가 나가지 않는 원천**: resume·clear·compact·fork에서는 리더 공지가 방출되지 않는다(기존 `…ForSource` 게이트). 문장이 새 방출 경로를 만들지 않는다(AC-015는 이 동작을 바꾸지 않는다).
- **todo 비활성**: 칸반 리더 공지에 `moai todo`가 없어야 한다. 문장은 이를 건드리지 않는다(AC-015 (b)).
- **알 수 없는 로케일**: `kanbanMessagesFor`·`factoryMessagesFor`가 영어로 폴백하므로 문장도 영어로 나간다(포인터 토큰은 같다).
- **레인·동반 세션**: 문장이 없다(AC-015 (c)). 레인은 카드 1장만 쥔다(REQ-BGS-001).
- **PASS-WITH-DEBT·BYPASSED 판정의 카드**: 승인 가능 행이 아니라 차단 행이다(REQ-BGS-011). 운영자가 이런 카드를 받아들이려면 요약 밖의 개별 경로로 간다.
- **배치 식별자 충돌**: 결정 보드에 같은 `batch=` 값이 이미 있으면 식별자는 다음 빈 초로 넘어가고(REQ-BGS-009), 한 호출에는 요약 질문이 하나뿐이다(REQ-BGS-005). 측정하지 않은 가정에 기대지 않는다(`design.md` §D.3). 점검 뒤에 쓰인 기록과의 충돌은 sync 감사가 찾는다(`spec.md` G-5).
- **Kickoff 외 게이트 행**: sync 차단 승인, 카드 선택, factory decide 행들은 요약 행이 아니라 개별 질문이다(REQ-BGS-001, A39).
- **선호가 디스크에 없는 카드**: 요약 승인 뒤에도 그 카드의 선호 질문이 따로 필요하다(REQ-BGS-020).
- **읽을 수 없는 자유 입력 · 보고서에 없는 카드 id**: 아무것도 승인하지 않고 같은 질문을 다시 낸다(REQ-BGS-006, A42).

## §D.2 Quality gates

- 새·변경 Go 파일(테스트 셋 + 제품 넷): `golangci-lint`는 변경 범위만, CI 버전 기준으로(`gofmt -l` 출력 없음, `go vet` 포함). 전체 스위트는 CI에 맡긴다. `internal/hook` 패키지 전체는 분 단위 스위트이므로 레인에서는 이름 선택만 돌린다.
- `moai spec lint SPEC-AUTONOMY-BATCH-GATE-001` — 트리에서 빌드해 커밋 스탬프를 박은 바이너리로, 판정 빌드 커밋을 함께 기록(VCI §2.2, `research.md` R1).
- 미러 동일성, 중립성, 예산, LOC, 보존 구간 가드 전부 GREEN(AC-010).
- 정본 절에 카드 id·SPEC ID·날짜·"176" 결과 수치가 없다(A38).

## §D.3 Definition of Done

- AC-001~010·AC-014~016 PASS(명령 + 원문 출력 + 트리 SHA), AC-011~013·AC-017 감사 읽기 통과. AC-014와 AC-015는 Q4·Q5 범위 판정(2026-10-02)으로 무조건이다.
- 기준선: M0 종료에서 AC-008 V1–V5가, 끝 점검에서 V6–V8(기준선 파일을 건드린 커밋은 B 하나, 정본 절 제목을 들인 커밋 D 하나, B가 D의 조상)이 통과한다. 조상·`-S` 판정은 M0 종료 조건이 아니다.
- `progress.md` §E.2에 RED 원문, 변이 거부 출력, 예산·미러·LOC 출력, 사전 차이 줄(`kanban-dispatch.md:177`, `moai.md`)을 건드리지 않았다는 확인이 있다.
- 결과 수치(176→10~20)를 어디에도 단언하지 않았다.

## §D.4 Traceability (REQ → AC)

| REQ | AC |
|---|---|
| REQ-BGS-001, 002, 003 | AC-006, AC-013 |
| REQ-BGS-004, 005, 006, 020 | AC-002, AC-011, AC-012 |
| REQ-BGS-007, 008, 009 | AC-003, AC-011, AC-012 |
| REQ-BGS-010, 012 | AC-004 |
| REQ-BGS-011 | AC-005, AC-011 |
| REQ-BGS-013 | AC-016, AC-001, AC-010 |
| REQ-BGS-014 | AC-014 |
| REQ-BGS-015 | AC-008 |
| REQ-BGS-016, 017, 018 | AC-015, AC-009, AC-010 |
| REQ-BGS-019 | AC-003, AC-017 |
| (전체 반증 가능성) | AC-007 |
