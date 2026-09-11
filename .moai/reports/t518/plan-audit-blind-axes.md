# SPEC Review Report: SPEC-SPEC-LINT-BLIND-AXES-001

카드: t518 · 워크트리: `.claude/worktrees/t518` · 브랜치 `WT-spec-lint-axes` · HEAD `0b1e27877`
감사 대상: `.moai/specs/SPEC-SPEC-LINT-BLIND-AXES-001/` (spec.md · plan.md · acceptance.md · progress.md, Tier M)

Iteration: 1/2 (Tier M ceiling)
**Verdict: FAIL**
Overall Score: **0.67** (조화평균) / 0.69 (산술평균) — Tier M PASS 임계 **0.80** 미달
(임계 출처: `.claude/rules/moai/workflow/spec-workflow.md:141`)

> Reasoning context ignored per M1 Context Isolation. 판정은 트리의 아티팩트와 소스만 근거로 한다.
> 이 감사는 형제 SPEC `SPEC-SPEC-LINT-ID-ARG-001`을 판정하지 않는다. 비중복 주장 확인 목적으로만 읽었고 편집하지 않았다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-SLB-001`~`REQ-SLB-012`, spec.md:83-94. 결번·중복 없음, 3자리 zero-pad 일관. 독립 측정으로 정의 행 12개 확인(수집기 정규식 재현).
- **[PASS] MP-2 GEARS 형식** — 12개 REQ 전부 `SHALL`/`SHALL NOT`을 명시한 요구사항 층 표기. **판정 층위: 요구사항 층(spec.md §C)에서만 판정했다.** acceptance.md의 Given-When-Then 항목은 검증 층이므로 여기서 감점하지 않았다(Group 4에서 별도 판정). 단 REQ-SLB-005의 패턴 라벨이 틀렸다 — D9 참조(본문은 다섯 패턴 중 하나에 해당하므로 MP-2 자체는 PASS).
- **[PASS] MP-3 YAML frontmatter** — spec.md:2-14. 12개 정식 필드(`id`/`title`/`version`/`status`/`created`/`updated`/`author`/`priority`/`phase`/`module`/`lifecycle`/`tags`) 전부 존재 + 선택 필드 `tier: M`. snake_case 별칭 사용 없음. 기계 확인: `moai spec lint <spec.md>` 결과에 `FrontmatterInvalid` 0건.
- **[N/A] MP-4 §22 프로그래밍-언어 중립성** — 이 SPEC은 `internal/spec`(Go) 단일 언어 범위이고 16개 지원 프로그래밍 언어 도구 체계를 다루지 않는다. 자동 통과. (SPEC이 다루는 "언어"는 대화 로케일 축(영어/한국어)이지 프로그래밍 언어 축이 아니다 — CLAUDE.local.md §15의 두 축 구분 적용.)
- **[PASS] MP-5 D7 교차 SPEC 정합** — 본문이 인용한 SPEC-ID 전부 실측: `SPEC-CODEX-PARTIAL-WIRING-001` completed · `SPEC-BINLAG-INVOCATION-001` completed · `SPEC-INIT-001` completed · `SPEC-CC-DOCS-ALIGNMENT-001` completed · `SPEC-MX-001` implemented · `SPEC-CLI-001` completed · `SPEC-LOOP-001` implemented · `SPEC-QUALITY-001` implemented. retired/superseded/archived 없음 → BLOCKING 없음. `SPEC-V3R4-CC2X-ADOPT-002`는 디렉터리는 존재하고 `spec.md`가 없다(`research.md`만) — spec.md:43이 그 사실을 정확히 설명하므로 D7-5 SHOULD 대상 아님.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — 4개 아티팩트 전부 `syscall` 문자열 0건(`grep -c` 실측). D8-4에 따라 자동 PASS.
- **[PASS(기계) / 실질 미해결] MP-7 해명 게이트** — `grep -rn '[NEEDS CLARIFICATION'` 매치 0건, `research.md` 부재(Tier M 아티팩트 집합상 정상). **다만 progress.md:10이 "미해결: D1·D2·D3는 run-phase 진입 전 확정 필요"라고 스스로 적고 있다.** 마커 관용구를 쓰지 않았을 뿐 실질은 미해결 3건이며, 그중 D1은 아래 D-1에서 **반증**됐다. MP-7은 기계적으로 통과하되, 이 상태로 Implementation Kickoff Approval에 올려서는 안 된다.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | REQ 12개 대부분 단일 해석. 다만 REQ-SLB-005 패턴 오라벨(spec.md:87), `자문 표시`의 기전이 plan.md D3에서 미결(plan.md:55-57), 무판정 코드명이 acceptance.md:66에서는 확정 `ModalityUnjudged` / plan.md:51에서는 "(가칭)" |
| Completeness | 0.75 | 0.75 | HISTORY/§A/§C/§D/§E/§F/§G + Tier M 3-아티팩트 전부 존재. `### Out of Scope — <topic>` H3 5개 각각 `-` 불릿 보유(spec.md:111-132) — `MissingExclusions` 0건으로 기계 확인. 감점: 자문 처리의 반사실(counterfactual)이 수치 없이 비어 있고(D-5), 재계수를 오염시키는 인접 결함이 §E에 없다(D-3) |
| Testability | 0.75 | 0.75 | AC 대부분이 매치 수·이름 붙인 코드로 단언하고 종료 코드 단독 단언이 없다(모범적). 감점: AC-SLB-001a/003이 존재하지 않는 수리 전 골든에 의존(D-7), AC-SLB-007의 픽스처가 구현 상대적(D-8) |
| Traceability | **0.50** | 0.50 | **acceptance.md의 12개 AC 중 REQ-ID를 인용하는 것이 0개.** 번호 정렬도 008에서 어긋난다(AC-SLB-008은 REQ-SLB-007의 등급 절을 검증). REQ-SLB-008에 대응 AC 없음, REQ-SLB-011/012는 AC가 아니라 §A 규율·§E 체크박스로만 실현 |

조화평균 = 4 / (1/0.75 × 3 + 1/0.50) = **0.667**. Tier M 임계 0.80 미달 → FAIL.

---

## Defects Found (구조화 결함 목록)

### D-1 — BINLAG 반례가 사실과 반대이고, 그 위에 D1 판별식 권고가 서 있다
- **위치**: `spec.md:41` · `plan.md:38` · `plan.md:31`
- **설명**: spec.md:41은 "`SPEC-BINLAG-INVOCATION-001`은 (정의는 100행의 목록 형식에 따로 있어서) **53개 목록에 들지 않는다**"고 적는다. **이 트리에서 반대로 측정된다.**
  - 100행 실제 내용: `**REQ-BLI-001** — The project shall not …` — **목록 불릿이 없다**(`repr` = `'**REQ-BLI-001** — The project shall not '`). 수집기 정규식 `^\s*[-*]\s+`가 요구하는 선두 불릿이 없으므로 매치되지 않는다.
  - 수집기 재현 측정: BINLAG의 narrow 매치 **0**, ID-선두 표 행 **16**. 즉 BINLAG는 **53개의 구성원이며 600행 중 16행을 기여한다**(멤버십 직접 확인).
  - 따라서 spec.md:41이 "잔여(residual)"로 남긴 "53개 안에도 처분·추적 표가 섞여 있을 수 있다"는 **가능성이 아니라 확인된 사실**이고, 600은 상한일 뿐 아니라 **최소 10행 과다 계상**이 실측됐다(아래).
  - 연쇄: plan.md:38은 후보 판별식 4("같은 문서에 목록 정의가 이미 있으면 표 기각")를 "BINLAG를 정확히 설명한다(그 SPEC은 애초에 53개에 들지 않는다)"는 근거로 필수 채택 권고한다. 전제가 거짓이므로 **근거가 무너지고, 후보 4는 BINLAG에 대해서도 발화하지 못한다**(목록 정의가 0개이므로).
  - 더 나아가 후보 4는 **53개 전체에 대해 구조적으로 무효(inert)** 다. 53의 정의 자체가 "수집기 0매치"(spec.md:36) = 목록 정의 0개이므로, "목록 정의가 있으면 기각"은 53개 어디에서도 조건이 성립하지 않는다.
  - 권고 집합 2+4의 실측 성능(BINLAG 16행 대상): 후보 2(모든 셀이 ID 토큰)가 **6행** 기각, 나머지 **10행**은 2·4 둘 다 통과해 요구사항으로 수집된다. 통과하는 행의 예 — `spec.md:124` `| REQ-BLI-001 | **이 카드가 실제로 세우는 것** | 인용 규율이 곧 이 요구다. 거처: … |`. 이는 plan.md:31이 스스로 "정의 아님 — 기각 대상"으로 분류한 바로 그 모양이다.
- **Severity: critical** — **Class: blocking**
- **필요한 수리**: (1) spec.md:41의 BINLAG 서술을 실측대로 정정하고 BINLAG를 53의 구성원으로 기록. (2) plan.md의 후보 4를 삭제하거나, "53에 대해 무효이며 53 밖의 혼재 문서에만 작동한다"는 적용 범위를 명시. (3) 처분 표(비-ID 셀 + 노트 칸)를 실제로 기각하는 판별식을 **plan-phase에서 확정**하고, 그 판별식을 BINLAG 16행에 시범 적용한 수치(기각 N / 수집 M)를 plan.md에 기록. 확정 없이 run-phase로 넘기면 REQ-SLB-004는 구현이 자기 채점표를 쓰는 요구가 된다.

### D-2 — 코퍼스 유래 수치가 상수로 고정됐고, 저작 중에 이미 움직였다
- **위치**: `spec.md:25` · `spec.md:35` · `spec.md:53` · `acceptance.md:84-85`
- **설명**: spec.md는 모집단을 "786개 디렉터리 / `spec.md` 보유 784개", 수집된 정의 행을 "3,932"로 적는다. **같은 워크트리에서 지금 재측정하면 788 / 786 / 3,952다.** 증분은 정확히 귀속된다: `spec.md` +2 = 이 SPEC + 형제 SPEC, 정의 행 +20 = 12(blind-axes) + 8(id-arg). 즉 **이 SPEC 자신의 저작이 이 SPEC이 상수로 못박은 수치를 움직였다.**
  - 축 2 수치(2,399 / 149 / 501)는 3,932 모집단 위에서 계산됐으므로 함께 낡았다.
  - 파급: REQ-SLB-009 / AC-SLB-009는 §B baseline(`4,346 @ 0b1e27877`, 784-파일 모집단)에 재측정치를 "나란히 싣는" 비교를 요구한다. 수리 후 실행은 ≥786 파일 모집단에서 돌므로, **증감표의 각 행은 수리 효과와 코퍼스 증가가 섞인 값이 된다.** §D:103이 "다른 트리·다른 시점의 수치를 baseline으로 재사용하지 않는다"고 적었지만, 같은 트리 안의 모집단 이동은 다루지 않는다.
  - 53과 600은 우연히 불변이다(새 SPEC 2개가 목록 형식이라). 불변이라는 사실이 방법의 안전성을 뜻하지는 않는다.
- **Severity: critical** — **Class: blocking**
- **필요한 수리**: (1) §A의 모집단·정의 행 수치를 "측정 시점의 재유도값"으로 표기하고 측정 명령을 함께 싣기(현재는 확정값처럼 읽힌다). (2) AC-SLB-009를 "동일 모집단 위 재유도"로 다시 쓰기 — 고정 4,346과의 단순 대조가 아니라, 수리 전 바이너리를 **현재 코퍼스**에 다시 돌려 baseline을 재수립하거나 양쪽 실행을 고정 파일 목록으로 제한할 것.

### D-3 — §B의 코드별 내역 합이 스스로 인용한 총계와 맞지 않는다
- **위치**: `spec.md:73`
- **설명**: §B가 나열한 12개 코드의 합은 3588+412+117+102+48+25+18+14+7+6+6+2 = **4,345**이고, 같은 절 :71이 인용한 관측 꼬리는 **4,346**이다. 증거 파일 `.moai/reports/t518/baseline-lint.txt`를 코드별로 세면 누락 항목이 나온다 — `OwnershipTransitionInvalid 1`. "필수 귀속"이라고 표시한 절 자체가 1건을 흘렸다.
- **Severity: major** — **Class: blocking**
- **필요한 수리**: `spec.md:73`에 `OwnershipTransitionInvalid 1`을 추가하고, 합이 :71의 꼬리와 일치함을 문서에 명시.

### D-4 — 재계수가 인접 결함으로 오염되는데 SPEC이 그 사실을 말하지 않는다
- **위치**: `spec.md:107-132`(§E) · `acceptance.md:82-86`(AC-SLB-009)
- **설명**: `CoverageIncomplete`는 baseline 4,346건 중 **3,588건**으로 압도적 다수이고, `doc.REQs`를 먹는 네 규칙 중 하나다. 그런데 AC 수집기는 `internal/spec/parser.go:67`에서 `##` 제목에 영어 토큰 `acceptance`를 요구하고 `parser.go:218`에서 `AC-[A-Z0-9]+-[0-9]+-[0-9]+:` 문법을 요구한다 — 이 코퍼스가 쓰지 않는 문법이다. 실측(이 SPEC 자신에게 도구를 돌림):

  ```
  moai spec lint .moai/specs/SPEC-SPEC-LINT-BLIND-AXES-001/spec.md   # 파이프 없음
  rc=0
  0 error(s), 12 warning(s)   ← 12건 전부 CoverageIncomplete
  WARNING CoverageIncomplete … spec.md 80  REQ REQ-SLB-012 is not referenced by any AC
  ```

  acceptance.md에 AC 12개가 실재하는데도 REQ 12개가 전부 "AC 미참조"로 보고된다. 즉 축 1이 표에서 최대 600개 REQ를 새로 수집하면 **최대 600건의 `CoverageIncomplete`가 추가로 발화되고, 그 대부분은 이 SPEC이 범위 밖으로 둔 AC 수집기 결함의 산물**이다. REQ-SLB-009는 "새로 드러난 부채를 세는 일은 수리의 일부"라고 하지만, 그 숫자를 실부채와 인접 결함의 인공물로 가를 방법이 없다.
  - 형제 SPEC은 이 결함을 자기 `spec.md:77-79`에 기록하고 "어느 카드도 소유하지 않은 채로 남는다"고 명시했다. **blind-axes에는 언급이 전혀 없다.**
  - 판정: 이 SPEC이 "영어 앵커 계열을 일반적으로 고친다"고 **과장 주장하지는 않는다**(제목·REQ-SLB-006/007 모두 modality로 한정) — 그 점은 옳다. 결함은 과장이 아니라 **상호작용의 미기록**이다.
- **Severity: major** — **Class: blocking**
- **필요한 수리**: (1) §E에 `### Out of Scope — AC 수집기(parser.go)`를 추가하고 소유자 부재를 기록. (2) AC-SLB-009의 증감표에서 `CoverageIncomplete` 행을 별도 표기하고 "이 행은 AC 수집기 사각지대와 교란돼 있어 단독으로 부채를 뜻하지 않는다"는 해석 규칙을 AC 본문에 못박기.

### D-5 — 자문 결정의 반사실이 수치 없이 비어 있다. 그 수치를 재는 도구가 이미 트리에 있다
- **위치**: `spec.md:84`(REQ-SLB-002) · `spec.md:143`(§F t385 선례) · `plan.md:55-57`(D3)
- **설명**: SPEC은 표 수집 항목을 자문으로 처리한다고만 말하고, **자문 처리를 하지 않았을 때 무엇이 벌어지는지를 어디에서도 서술하지 않는다.** 결정이 뒤집기 어려운 상태로 남는다(비용이 보이지 않으므로).
  - 트리에는 이미 답이 있다. `internal/spec/lint_req_widen.go:85-91`: "doc.REQs feeds four **error-severity** findings — ModalityMalformed, InvalidREQID, DuplicateREQID, CoverageIncomplete — and none of those codes is in eraDemotableCodes … Measured live before the wiring: 25 ModalityMalformed and 6 InvalidREQID errors appear".
  - 계측기도 이미 있다. `internal/spec/lint_req_widen_decompose_test.go:282-311`은 네 규칙 전부에 대해 narrow vs wide 폭발 반경을 출력하고(`blast_ModalityMalformed_error narrow=… wide=…` 등), `internal/spec/lint_req_widen_corpus_test.go`는 `MOAI_T362_CORPUS_SCAN`으로 게이트된 코퍼스 측정 하네스다.
  - 그런데 spec.md §G와 plan.md M1은 측정 하네스를 **새로 승격시킨다**고만 적고 이 두 파일을 한 번도 인용하지 않는다. 이미 있는 것을 다시 만드는 방향이고(Enforce Simplicity 사다리 2단), 동시에 반사실 수치를 놓친다.
- **Severity: major** — **Class: blocking**
- **필요한 수리**: (1) `lint_req_widen_decompose_test.go`의 blast-radius 출력을 근거로 "자문 없이 표를 수집하면 error 등급이 몇 건 생기는가"를 실측해 spec.md §A 또는 §B에 싣기. (2) §G/M1을 "기존 하네스 재사용 + 표 경로 확장"으로 다시 쓰고 두 파일을 §H 상호참조에 추가.

### D-6 — AC와 REQ 사이에 명시적 인용이 하나도 없다
- **위치**: `acceptance.md:20-91` 전체
- **설명**: AC-SLB-001a·001b·002·003·004·005·006a·006b·007·008·009·010 — **어느 것도 REQ-ID를 인용하지 않는다.** 번호가 대응을 암시하지만 008에서 어긋난다: AC-SLB-008("무판정 코드의 등급")이 검증하는 것은 REQ-SLB-007의 등급 절이고, **REQ-SLB-008("판정하지 않은 요구사항을 적합하다고 보고해서는 안 된다")에는 대응 AC가 없다**(AC-SLB-007이 부분적으로만 걸친다). REQ-SLB-011·012는 AC가 아니라 acceptance.md §A 규율과 §E 체크박스로만 실현된다.
  - 이 SPEC은 §E:126-128에서 t524(plan-auditor가 축약 REQ 표기를 통과시키는 결함)를 인용한다. **같은 약점을 자기 문서에서 재생산하고 있다.**
- **Severity: major** — **Class: blocking**
- **필요한 수리**: 각 AC 제목 또는 Then 절에 `(REQ-SLB-00X)`를 붙이고, REQ-SLB-008 전용 AC를 추가(예: 판정 불가 본문에 대해 "적합" 신호가 0건임을 매치 수로 단언).

### D-7 — 두 AC가 존재하지 않는 수리 전 산출물에 의존하고, 그 채취 창은 M2에서 닫힌다
- **위치**: `acceptance.md:23`(AC-SLB-001a) · `acceptance.md:39-41`(AC-SLB-003) · `plan.md:75-81`(§F 마일스톤)
- **설명**: AC-SLB-001a는 "각 항목의 ID·본문·행번호는 **수리 전 구현의 결과와 바이트 단위로 같으며**"를, AC-SLB-003은 "**수리 전후의 두 트리**에서"를 요구한다. 그런데 M1~M5 어디에도 수리 전 골든을 채취하라는 산출물이 없다(M1 산출물은 "재현 가능한 측정 명령, 확정된 baseline 표"로, REQ 항목 단위 골든을 포함하는지 불명). M2가 착지하면 "수리 전 트리"는 재현 비용이 급등한다.
  - 값싼 대안이 이미 존재한다: `parseREQsWithProvenance`는 narrow `parseREQs`를 **그대로 호출해 보존**하므로(`lint_req_widen.go:94-105`), 무변형 단언은 두 트리 비교가 아니라 **한 트리 안에서 narrow 경로 대조**로 쓸 수 있다.
- **Severity: major** — **Class: blocking**
- **필요한 수리**: M1에 골든 채취를 증거 경로와 함께 산출물로 추가하거나, AC-SLB-001a/003을 보존되는 `parseREQs` 경로 대조로 다시 기술.

### D-8 — 출현을 단언하는 AC 4건에 뮤턴트 가드가 없다 (SPEC 자신의 규율 위반)
- **위치**: `acceptance.md:33-36`(002) · `:51-54`(005) · `:69-73`(007) · `:75-78`(008), 위반 대상 규율 `acceptance.md:14`(§A 규칙 3) · `spec.md:94`(REQ-SLB-012)
- **설명**: 가드 보유는 001b(:29) · 003(:42) · 004(:48) · 006b(:67) 4건뿐이다. 출현을 단언하면서 가드가 없는 것: AC-SLB-002(자문 finding 존재), AC-SLB-005("정확히 표 개수만큼 존재"), AC-SLB-007("후자는 1건 이상"), AC-SLB-008(자문 등급). 특히 **AC-SLB-005는 REQ-SLB-005를 검증하는 유일한 AC**인데 가드가 없어 비공허 증거가 없다.
- **Severity: major** — **Class: blocking**
- **필요한 수리**: 4건 각각에 뮤턴트를 명시하거나, 기존 가드의 사정거리 안에 있음을 AC 본문에 근거와 함께 기록.

### D-9 — 축 2의 하한을 지키는 AC가 구현이 스스로 채점표를 쓰는 구조다
- **위치**: `acceptance.md:69-73`(AC-SLB-007) · `plan.md:50-53`(D2)
- **설명**: 축 2의 하한("판정할 수 없으면 **발화**하라", REQ-SLB-006)은 요구로 잘 서 있고(SHALL, `spec.md:88`), 갈래 A만 구현하면 **AC-SLB-007이 실제로 FAIL한다** — 그 점은 옳다. 그러나 007의 픽스처 "판정 불가한 한국어 요구사항"의 내용이 고정돼 있지 않고, 무엇이 "판정 불가"인지는 갈래 A의 표지 집합에 따라 정해진다. 그 표지 집합을 고르는 주체가 같은 run-phase다. 따라서 **A의 사정거리 안쪽 문장을 픽스처로 고르면 007은 무해하게 통과한다.** 축 2 하한이 걸린 유일한 AC가 그런 자유도를 갖는다.
- **Severity: minor** — **Class: blocking**
- **필요한 수리**: A의 표지 집합이 정해지기 **전인 지금** acceptance.md에 픽스처 문장을 축자로 못박기.

### D-10 — REQ-SLB-005의 GEARS 패턴 라벨이 틀렸다
- **위치**: `spec.md:87`
- **설명**: `(Where)`로 라벨했으나 본문은 "판별식이 어떤 표를 정의 표가 아니라고 **기각하는 경우**"로, 사건 조건이다. GEARS에서 `Where`는 capability gate / feature flag / static config를 뜻하고 조건부 사건은 `Event-driven`이다. 본문 자체는 다섯 패턴 중 하나에 해당하므로 MP-2는 통과하지만, 라벨은 이 SPEC이 §C에서 스스로 다는 modality 주장이다.
- **Severity: minor** — **Class: blocking**
- **필요한 수리**: `(Event-driven)`으로 정정.

### D-11 — 형제 SPEC이 ID로 지명되지 않아 범위 경계가 검증 불가다
- **위치**: `spec.md:113` · `spec.md:142`
- **설명**: §E와 §F 모두 "별도 형제 SPEC"이라고만 적고 ID를 쓰지 않는다. 반대 방향은 명시적이다 — 형제가 `related_specs: [SPEC-SPEC-LINT-BLIND-AXES-001]`(그 spec.md:15)와 본문 :20에서 이 SPEC을 ID로 지명한다. 비대칭이며, 이 문서만 읽는 독자는 축 3의 소유자를 확인할 수 없고 비중복 주장도 검증할 수 없다. D7 정규식 스캔에도 잡히지 않는다.
- **Severity: minor** — **Class: blocking**
- **필요한 수리**: §E·§F에 `SPEC-SPEC-LINT-ID-ARG-001`을 명기하고 frontmatter에 `related_specs` 추가.

### D-12 — 축 2 수치가 표에서는 확정값 모양으로, 두 줄 뒤에서는 잠정값으로 제시된다
- **위치**: `spec.md:51-58`(표) vs `spec.md:60`(단서)
- **설명**: 600에 대해서는 규율이 지켜졌다 — :37 "(미방문 요구사항의 **상한**)", :41 "600은 상한이며", plan.md:62까지 세 자리 전부 상한임을 밝힌다. **축 2에는 같은 규율이 없다**: :51-58 표는 "이 트리에서 … 잰 값"이라는 제목 아래 **149를 볼드**로 제시하고, 잠정성은 :60에서야 나온다. 표만 인용되면 확정값으로 전파된다.
  - 철회 기록 상태: (a) 4,344 우연 일치는 :75에 "**[HARD] 철회된 가설을 기록한다**"로 명시적으로 있다 ✓. (b) 이전 축 2 측정(1,787/177/516)은 :60에서 "어느 쪽도 정본으로 채택하지 않는다"로 처리돼 조용한 대체는 아니다 ✓ — 그러나 차이의 원인을 "…때문으로 **보이며**"라는 추정으로만 적었고, 그 측정이 **알려진 결함(마지막 구분자까지 먹는 greedy `sed`가 본문을 잘라냄)** 의 산물이라는 사실은 기록되지 않았다. 방법 결함이 기록되지 않으면 같은 방법이 다시 유도된다. 재측정치 2,399/149/501도 정본으로 채택되지 않았다 ✓(옳은 처리).
- **Severity: minor** — **Class: blocking**
- **필요한 수리**: :51-58 표의 각 행에 잠정 표기를 인라인으로 달고, 볼드를 해제. :60에 이전 측정의 방법 결함을 사실로 기록.

### D-13 — AC-SLB-010은 no-op으로 만족되고 가드도 짝도 없다
- **위치**: `acceptance.md:88-92`
- **설명**: "종료 코드가 뒤집히는 픽스처가 없다"는 부재 단언이고, 아무것도 구현하지 않아도 참이다. §A 규칙 3이 "부재 단언은 그 자체로는 아무것도 증명하지 않는다"고 스스로 적었는데 010에는 뮤턴트도 대조 짝도 없다. (AC-SLB-001a와 006a도 no-op으로 만족되지만 이 둘은 **설계상 대조군**이고 각각 001b·006b와 짝지어져 있으므로 정상이다.)
- **Severity: minor** — **Class: optional**
- **필요한 수리**: 001b/006b가 발화시키는 자문 finding이 error 등급 집계에 들어가지 않음을 매치 수로 단언하는 형태로 010을 다시 쓰기(그러면 001b의 뮤턴트가 010까지 덮는다).

### D-14 — 축 1 대조쌍의 비교가 AC가 아니라 산문과 체크박스에 있다
- **위치**: `acceptance.md:31`(산문 주석) · `acceptance.md:96`(§E 체크박스)
- **설명**: 두 대조쌍은 모두 존재한다 ✓ — 축 1은 001a(목록)/001b(표), 축 2는 006a(영어)/006b(한국어). 축 2는 **AC-SLB-007이 두 픽스처의 finding 집합을 실제로 비교**한다 ✓. 축 1에는 그런 AC가 없다: 001a와 001b는 각자 자기 픽스처에만 단언하고, "둘이 서로 다른 결과를 낸다"는 비교는 :31 산문과 :96 완료 정의 체크박스에만 있다. 다만 001b의 뮤턴트(수집 0 → RED)와 자문 표시 유무의 차이가 "둘 다 무시됨"을 실질적으로 배제하므로 치명적이지는 않다.
- **Severity: minor** — **Class: optional**
- **필요한 수리**: 001a/001b의 결과를 한 테스트에서 대조하는 AC를 추가하거나, :31의 대조 요건을 AC의 Then 절로 승격.

### D-15 — 무판정 코드명이 두 문서에서 확정도와 표기가 다르다
- **위치**: `acceptance.md:66`(`ModalityUnjudged` 확정 사용) vs `plan.md:51`(`ModalityUnjudged`(가칭))
- **Severity: minor** — **Class: optional**
- **필요한 수리**: 이름을 확정하고 양쪽을 맞추거나, AC를 "이름 붙인 무판정 코드(구현이 확정)"로 중립화하되 M3 산출물에 그 이름을 기록하도록 명시.

---

## 확인했고 결함이 아닌 것 (기록 — 다시 유도하지 않기 위해)

- **종료 코드 규율 — 통과.** acceptance.md:12(§A 규칙 1) · spec.md:93(REQ-SLB-011) · acceptance.md:101(§E 체크박스)이 종료 코드 단독 단언을 금지한다. §B:70이 baseline 명령에 "**파이프 없음**"을 명시한다. AC-SLB-010조차 판별식이 종료 코드가 아니라 `error 등급 findings의 존재 여부`다. 12개 AC 중 종료 코드를 판별식으로 쓰는 것은 **0건**.
- **10 vs 11 불일치 처리 — 통과.** spec.md:43이 카드의 10과 이 트리의 11을 나란히 적고 어느 쪽도 조용히 채택하지 않으며 M1에서 행 단위 확인을 예정한다. 독립 재측정 결과 이 트리에서 `SPEC-CODEX-PARTIAL-WIRING-001`은 **11**이다(SPEC 쪽이 맞다).
- **53 / 600 / 최대 보유 6개 — 재현됨.** 수집기 정규식과 표 행 정규식을 그대로 옮겨 재현: blind 53, rows 600, top-6 = INIT-001 52 · CC-DOCS-ALIGNMENT-001 33 · MX-001 29 · CLI-001 26 · LOOP-001 23 · QUALITY-001 23 — spec.md:36-38과 완전 일치.
- **Out of Scope 절 — 통과.** `### Out of Scope — <topic>` H3 5개(spec.md:111·116·121·126·130)가 각각 구체적 `-` 불릿을 보유. 기계 확인으로 `MissingExclusions` 0건.
- **REQ/AC 예산 — 통과.** REQ 12 ≤ 16, AC 12 ≤ 16 (Tier M 상한, `spec-workflow.md:147-150`).
- **자문 기전의 실재 — 확인.** `REQEntry.Widened`와 `reqFindingSeverity`(`lint.go:712-717`)가 이미 존재하고, `parseREQsWithProvenance`는 narrow에 없는 항목을 자동으로 `Widened=true`로 표시한다. 표 행은 narrow에 결코 잡히지 않으므로 REQ-SLB-002는 기존 통로로 거의 무비용으로 성립한다. plan.md D3의 `FromTable` 별도 필드 제안은 재계수 분리 목적상 합리적이다.

---

## 브리핑에서 받은 사실 중 이 트리에서 재현되지 않은 것

- **"`SPEC-BINLAG-INVOCATION-001`은 53에 들지 않는다(100행에 목록 형식 정의를 갖는다)"** — 재현되지 않는다. 100행은 `**REQ-BLI-001** — …`로 **불릿이 없고**, narrow 매치 0 / 표 행 16으로 **BINLAG는 53의 구성원**이다. 브리핑과 spec.md:41이 같은 오류를 공유한다. 이것이 D-1의 근거이며, 브리핑이 "반례가 아니므로 잔여는 미검증으로 남는다"고 본 지점은 실제로는 **잔여가 확인된** 지점이다.

---

## Recommendation

FAIL. 아래 순서로 수리한 뒤 iteration 2를 요청할 것. 이 SPEC은 Tier M ceiling 2이므로 다음이 마지막 반복이다.

1. **D-1을 먼저 닫는다.** spec.md:41 정정 → plan.md:38 후보 4의 적용 범위 정정 또는 삭제 → **처분 표를 실제로 기각하는 판별식을 plan-phase에서 확정**하고 BINLAG 16행(현재 2+4로는 6 기각 / 10 수집)에 시범 적용한 수치를 plan.md에 기록. D1이 열린 채로는 REQ-SLB-004가 검증 불가다.
2. **D-2**: §A 수치를 재유도값으로 표기하고, AC-SLB-009를 동일 모집단 재유도로 다시 쓴다.
3. **D-3**: spec.md:73에 `OwnershipTransitionInvalid 1`을 추가해 합을 4,346에 맞춘다.
4. **D-4**: §E에 AC 수집기 out-of-scope 절을 추가하고, AC-SLB-009 증감표의 `CoverageIncomplete` 행에 교란 해석 규칙을 못박는다.
5. **D-5**: `lint_req_widen_decompose_test.go`의 blast-radius로 자문 미적용 반사실을 실측해 싣고, §G/M1을 기존 하네스 재사용으로 다시 쓴다.
6. **D-6**: 12개 AC에 `(REQ-SLB-00X)`를 달고 REQ-SLB-008 전용 AC를 추가한다.
7. **D-7 / D-8 / D-9**: 골든 채취를 M1 산출물로 올리거나 001a·003을 narrow 경로 대조로 재기술 · 002/005/007/008에 뮤턴트 명시 · AC-SLB-007 픽스처 문장 축자 고정.
8. **D-10 ~ D-12**: 라벨·형제 ID·축 2 표 잠정 표기 정정.
9. D-13 ~ D-15는 optional. 오케스트레이터 재량이며 이것들만으로 FAIL을 만들지 않았다.

**게이트 판단**: progress.md:10이 D1·D2·D3를 미해결로 자인하고 있고 그중 D1은 위에서 반증됐다. Implementation Kickoff Approval은 최소한 D-1이 닫힌 뒤에 열어야 한다.

---

## 검사하지 않은 것 (Gaps — 명시)

- **`go test ./internal/spec/...`를 돌리지 않았다.** 감사 지시가 read-only이므로, plan.md:61의 사전 점검("기존 테스트가 현재 GREEN인지 확인")은 **미검증**이다. 이 패키지의 현재 테스트 상태에 대해 나는 아무것도 주장하지 않는다.
- **전체 코퍼스 `moai spec lint` 재실행을 하지 않았다.** §B의 총계·코드별 내역은 기존 증거 파일 `.moai/reports/t518/baseline-lint.txt`를 판독해 검증했고, 도구는 **단일 파일 1회**만 직접 실행했다(위 D-4의 12건). 4,346이 `0b1e27877`의 값이라는 커밋 귀속은 증거 파일의 기록을 신뢰한 것이며 내가 그 커밋에서 재측정하지는 않았다.
- **53개 blind SPEC 개별 분류를 하지 않았다.** BINLAG 1건만 행 단위로 분류했다(6 추적 / 10 처분). 나머지 52개에서 정의 표 대 처분 표의 비율은 **미측정**이며, 따라서 600의 실제 과다 계상 폭은 "최소 10행"까지만 말할 수 있다.
- **형제 SPEC `SPEC-SPEC-LINT-ID-ARG-001`을 판정하지 않았다.** 비중복 확인 목적의 발췌 판독만 했다.
- **축 2 수치(2,399 / 149 / 501)를 독립 재측정하지 않았다.** 판독 기준이 SPEC에서 아직 고정되지 않았으므로(M1 소관) 재측정해도 귀속 가능한 대조가 되지 않는다. 정의 행 총수만 재측정했고 **3,952**로 SPEC의 3,932와 다르다(D-2).
- **크로스모델 2차 의견을 구하지 않았다.** 프로젝트 `audit_model` 설정을 확인하지 않았고 `audit_multi` / `codex_audit` / `glm_audit`을 호출하지 않았다. 이 판정은 Claude 단독 앵커다.

## Residual-risk

- D-1의 수리가 "판별식을 하나 더 추가"로 끝나면, 추가된 판별식이 처분 표를 잡는지는 다시 53개 코퍼스에서 재야 한다. plan.md §C의 시범 적용이 그 자리이지만, 지금은 "몇 개가 남는지 잰다"까지만 적혀 있고 **정탐/오탐을 가르는 판정 기준이 없다**.
- 재계수(REQ-SLB-009)는 D-2와 D-4를 둘 다 고쳐야 해석 가능해진다. 한쪽만 고치면 증감표는 여전히 원인 미상의 숫자를 낸다.
- `isModalityMalformed`는 `" SHALL"`(선행 공백)을 요구한다(`lint.go:792` 이하). 코퍼스의 한국어 요구사항은 `해야 한다(SHALL)` 꼴로 괄호가 앞에 붙으므로, 갈래 A가 `(SHALL)`을 SHALL 대응물로 인정하려면 이 접촉 조건까지 함께 다뤄야 한다 — plan.md D2에 언급이 없다. (추론이 아니라 소스 판독이지만, 갈래 A 설계에 미치는 영향은 아직 가설이다.)
