# t1051 Plan-Audit Verdict — SPEC-WEB-CONSOLE-017

- **Iteration**: 1/2 (Tier M — `harness.plan_audit_tier_ceilings` M=2)
- **Verdict**: **PASS-WITH-DEBT** (skip-eligible 아님 — D1 수리로 artifact hash 변경 예정 + 본 판정은 skip 계약의 `PASS` 아님)
- **Overall Score**: **0.9375** (4차원 등산평균 — Tier M 임계 0.80 이상)
- **감사 대상**: `.moai/specs/SPEC-WEB-CONSOLE-017/` 4파일, 커밋 `80c2832df` (트리 `WT-save-observability`)
- **감사자**: plan-auditor (독립 — 작성자 reasoning context 미주입, M1 Context Isolation)

---

## 1. Must-Pass 결과 (7/7 PASS)

| # | 판정 | 근거 |
|---|------|------|
| MP-1 REQ 번호 일관성 | **PASS** | `grep '^- REQ-WC-017'` → spec.md:59,60,64,65,66,67 — 001..006 연속, 공백·중복 없음, 3자리 패딩 일관 |
| MP-2 GEARS 형식 (요구층) | **PASS** | 6건 전부 패턴 적합 — 001 `**When** … SHALL`(사건, spec.md:59) · 002 복합 `SHALL … **While** … SHALL NOT`(spec.md:60, GEARS 복합절 PASS-동등) · 003 `**When** … SHALL`(:64) · 004 `Every … SHALL`(일반, :65) · 005 `SHALL NOT`(금지형, :66) · 006 `SHALL`(:67). 판정 대상은 spec.md §2의 REQ-XXX 요구층. AC의 Given-When-Then은 검증층 올바른 형식으로 MP-2 채점 제외 |
| MP-3 Frontmatter 유효성 | **PASS** | spec.md:1-18 — 정본 12필드 전부 존재·타형 일치(id/title/version "0.1.0"/status draft/created·updated 2026-09-22/author/priority P1/phase "v3.2.0 target"(릴리스 라벨로 유효)/module/lifecycle/tags). snake_case 별칭(created_at 등) 0건. `moai spec lint` 실측: `✓ No findings` |
| MP-4 언어 중립성 | **N/A auto-PASS** | `module: internal/web` — 단일 언어(Go) 프로젝트 스코프 SPEC. 다국어 도구 나열 축 없음 |
| MP-5 D7 cross-SPEC | **PASS** | related_specs 4건 전부 존재, status 실측: 016=draft, 011=completed, GLM-KEY-INPUT-001=completed, JEV-OPTIN-MEASURE-001=completed — retired/superseded/archived 0건 → BLOCKING 없음. 016 draft 상태는 D3(하단)로 기록 |
| MP-6 D8 syscall | **PASS (auto)** | `grep -c syscall spec.md` = 0 → D8-4 자동 통과 |
| MP-7 clarification gate | **PASS** | `grep 'NEEDS CLARIFICATION'` → plan.md:66 1건 뿐이며 **부정문**(「`[NEEDS CLARIFICATION]` 마커를 붙이지 않는다 — 이 질문은 Kickoff 게이트를 막지 않는 것이 의도다」). 미해결 마커 아님. research.md 부재(Tier M) |

## 2. 차원 점수

| 차원 | 점수 | 밴드 | 근거 |
|------|------|------|------|
| Clarity | 1.00 | 1.0 | REQ 전부 단일 해석; §5.1 표(spec.md:94-103)가 문구·토큰을 소진; §5.2 접두어 결정이 3근거+비용으로 정당화(spec.md:109-115) |
| Completeness | 1.00 | 1.0 | HISTORY/§1/§1.1/§2/§3/§4 HARD-1..7/§5/§6/§F 전부 존재; §F에 `### Out of Scope —` H3 4개 + 불릿(spec.md:140-160); Frontmatter 완비 |
| Testability | 1.00 | 1.0 | AC 5건 전부 이진 판정 — DOM 요소+본문 문구(AC-001, acceptance.md:12), 9-seam 표 기반 서브테스트(:17), stderr 행수+`^moai web: ` 접두어+성공 0행(:24), sentinel 부재(:29), 성공 불변식(:35). weasel word 0건. 판별식을 본문 문구로 고정(HARD-4, spec.md:82)하고 양방향 가드 명시(acceptance.md:42) |
| Traceability | 0.75 | 0.75 | §D.3 표(acceptance.md:47-54) 양방향 완비·고아 0건. **단**, REQ-WC-017-003의 "exactly one"(spec.md:64)이 AC 어디에도 고정되지 않음 — AC-WC17-003(a)의 "1행 이상"(acceptance.md:24)은 요구의 규범 강도를 소진하지 못함 → D1. 단일 REQ의 부분 소진 = 0.75 밴드 |

## 3. 앵커 재측정 (배차 지시 이행 — 인용 전 자체 측정)

SPEC §1.1/§5의 모든 file:line을 본 감사가 본 트리에서 직접 재측정. **전 항목 정확 일치.**

| 앵커 (SPEC 주장) | 실측 | 일치 |
|---|---|---|
| 9개 `a.renderErrorPage(` call site | handlers.go:497,518,526,535,545,555,565,578,591 | ✓ |
| renderErrorPage 정의 + 500 | handlers.go:644, `http.StatusInternalServerError` :648 | ✓ |
| 인라인 슬롯 `save__msg save__msg--error` role="alert" | shell.templ:274 | ✓ |
| 폴백 문구 `Some fields could not be saved` | shell.templ:279 | ✓ |
| error 분기 렌더 블록 | shell.templ:273-281 | ✓ |
| 슬롯 스타일 | console.css:155 | ✓ |
| `BannerKind=="error"` → `vm.SaveMessage` | settings_shell.go:40-42 | ✓ |
| settingsSaveState | settings_shell.go:63-72 | ✓ |
| stderr 접두어 `web: ` / `moai web: ` | server.go:252 / :290 | ✓ |
| `bannerClass` error→`banner banner--warn` 사상 + 단일 강조 주석 | root.templ:26-34 | ✓ |
| `hx-boost="true"` | root.templ:56 | ✓ |
| 제출 전 경고 배너 선존재 (공허-참 함정 근거) | fieldsets.templ:607 | ✓ |
| 성공 배너 `Settings saved.`+ok | handlers.go:597-601 | ✓ |
| `log/slog` 임포트 | internal/web 0건 | ✓ |
| `os.Stderr` 비-저장 2곳 | :252 파일감시 / :290 브라우저 — 저장 경로 0건 | ✓ |
| `recordingSeams` 하니스 | partial_apply_repro_test.go:66 (t1049 a52ce60f0 소출품, 카드 이전 커밋 — 본 카드 프로덕션 변경 아님) | ✓ |
| §5.1 표 9개 배너 문구 | 9건 전부 handlers.go:497-592 실문과 동일 | ✓ |

## 4. 기계 검사 + 양성 대조

- **spec lint** (실측 `moai spec lint .moai/specs/SPEC-WEB-CONSOLE-017`): `✓ No findings — all SPEC documents are valid`. 공허 여부를 변이 3건으로 닫음(DoD-6 요구의 표본 검증):
  - 변이 A(/tmp 사본, `status:` 삭제): `FrontmatterInvalid` ERROR 1건 발화 → /tmp 트리 검사 자체가 산다는 전제 단언.
  - 변이 B(마커 유지, REQ-001 `SHALL`→`should`): **`ModalityUnjudged` WARNING 1건 발화**(REQ-WC-017-001 지목) → 수집기가 리스트-마커 REQ 줄을 보고 modality 룰이 소비함. 실제 SPEC의 `✓ No findings`는 비공허.
  - 변이 C(마커 제거 + `SHALL`→`should`): `✓ No findings` 침묵 → 굵은-글씨 형태엔 수집기가 눈멂(t1020/t1057 교훈의 오늘 바이너리 재현). **실제 SPEC은 수집되는 쪽(`- ` 마커) 형태** — 6줄 전부.
- **범위** (`git show --stat 80c2832df`): 4 SPEC 파일만, 프로덕션 0건. 부모 `616f7451d`와 HEAD 사이 커밋은 이 하나.
- **CDP 수치**: 콘솔 에러 2/0건·본문 ~110KB는 spec.md:49(§1.1)·:129(§6-1)에 **전제로 선언** — 본 SPEC의 측정으로 제시하지 않음. ✓ (배차 구속조건 충족)
- **[NEEDS CLARIFICATION]**: 0건 미해결(plan.md:66은 부정문 메타 언급 — §1 MP-7).
- **REQ-A/REQ-B 분리**: §2.1/§2.2 표면 분리 + 병합 금지가 spec.md:32, plan.md §G, acceptance.md:3에 삼중 명시. ✓
- **가드 판별식**: 본문 문구 전용(HARD-4), 클래스 금지 근거가 fieldsets.templ:607 선존재·root.templ:29-34 사상으로 코드 고착됨, 오류값 금지(REQ-005)+sentinel AC-004. ✓

## 5. Tier M 판정 (배차 항목 5)

**Tier M 정당 — threshold-shopping 아님.** 기계 가이던스는 Tier S(<300 LOC, <5파일, REQ 6/AC 5 모두 S 천장 8 이내). 그러나 (i) 작성자가 편향 방향을 스스로 공개했다(plan.md §B — 대가로 임계 0.75→0.80 상향 + 산출물 1건 추가를 명시), (ii) 편향의 방향이 **자기-처벌적**이다 — 상위 Tier는 감사자에게 더 엄격한 기준이지 작성자 이익이 아니다, (iii) 인용한 3근거(병합 금지 쌍 표면, 공허-참 함정 가드의 반증력 규칙 추적, 자격증명 sentinel AC)는 acceptance.md를 독립 파일로 요구할 실질이 있다. 과대 분류는 보수적 오류다. 관측 기록으로 남김.

## 6. 결함 목록

- **D1 — SHOULD-FIX — Class: blocking** — `acceptance.md:24` — AC-WC17-003(a)의 「저장-실패 행이 **1행 이상**」이 REQ-WC-017-003(spec.md:64)의 「**exactly one** stderr line」을 소진하지 못한다. 변이: 헬퍼가 실패마다 2행(본문+요약)을 남기면 AC (a)+(b) 전부 통과하면서 REQ-003 위반 — 모든 AC를 통과하는 변이가 존재한다. **필요 수리**: (a)을 「강제 실패 1건당 stderr 저장-실패 행이 **정확히 1행**」으로 조이거나 행수 `== 1` 단정을 추가. 1줄 수리. 수리 시 hash 변경 → **delta 재감사(D1 단건 범위)**.
- **D2 — MINOR — Class: optional** — `acceptance.md:13,18,25` — 각 release-blocking AC의 「RED 근거」는 관측된 RED(명령·stdout·exit·SHA 4요소, verification-completeness.md §2.1)가 아니라 **코드 수준 추론**이다. plan-phase에서 테스트가 미존재해 관측 불가능하며 run-phase E8(verbatim pre-GREEN)에서 완성되는 것이 정합이나, §D.2가 이를 "RED-now 셀"이라 부르는 표현이 관측과 추론을 구분하지 않는다. 명시 한 줄 추가 권고(「RED 관측은 run-phase RED에서 귀속됨」).
- **D3 — MINOR — Class: optional** — `spec.md:16` — related_specs 중 SPEC-WEB-CONSOLE-016만 `draft`. recordingSeams 하니스 실물은 이미 착지(t1049 a52ce60f0)해 의존 실체는 존재하나, 016의 문구 개정(REQ-WC16-007 수리)이 미착수라 §5.1 표 7-9행 문구의 상류 변동 가능성이 열려 있다. §6-3 연동 규정으로 이미 관리 중 — 조치 불요, 관측 기록.

## 7. 권고

1. manager-spec이 D1을 수리한다(acceptance.md AC-WC17-003(a) exact-count 조정 — 1줄).
2. 수리 커밋 후 **delta 재감사**(D1 단건 — Regression Check: D1 RESOLVED 여부) → clean PASS 예상.
3. 이후 Implementation Kickoff Approval — M1(REQ-A 수송 기구)은 사용자 가시 동작을 바꾸므로 운영자 판정 대상(progress.md §E.1 기재와 일치). `banner--error` 운영자 대기 질문은 HARD-6으로 비종속 확립돼 게이트를 막지 않는다.

## 8. 증거-수반 보고 (VCI 5-섹션)

- **Claim**: SPEC-WEB-CONSOLE-017 plan 산출물은 Tier M 임계 0.80을 0.9375로 통과하며 must-pass 7건 전부 통과. 단 D1(AC 강도 결손) 1건을 blocking-class로 기록.
- **Evidence**: 본 문서 §1-§4 전부 — 각 행의 명령과 실측 출력( lint 3-변이 발광/침묵, git show --stat 4파일, 관련 SPEC status 4건, grep 결과)은 2026-09-22 본 감사 실행의 관측값.
- **Baseline-attribution**: 트리 `WT-save-observability` @ `80c2832df` (plan 커밋) 기준, 본 감사 실행(2026-09-22)에서 측정. 코드 앵커는 같은 트리 working file에서 직접 재측정.
- **Gaps**: CDP 콘솔 에러 수·본문 크기(카드 전제로만 선언 — 브라우저 재측정 안 함), htmx 버전별 비-2xx 처리 세부(SPEC §6-2가 run-phase 설계 결정으로 위임), `go test ./internal/web/...` 기준선(plan.md §C가 run-phase 재측정 의무로 배치 — 본 감사 미실행), 016 문구 개정 착수 여부.
- **Residual-risk**: D1 미수리 상태 run 진입 시 "exactly one" 불변식이 AC에 고정되지 않은 채 구현 착지 가능. §5.1 안정부 문구에 묶인 가드 리터럴이 016 개정과 동반 갱신되지 않으면(§6-3 위반) 가드가 오탈 방향으로 굳는다.

---
---

# Iteration 2 — Delta Re-audit (D1 종결 확인)

- **Iteration**: 2/2 (Tier M 천장 — delta 스코프, 전면 재심 아님)
- **수리 커밋**: `a09821c98` — `fix(t1051): plan-audit D1 — AC-WC17-003(a) upper bound`
- **Verdict**: **PASS (clean)**
- **Overall Score**: **1.00** (Clarity 1.00 / Completeness 1.00 / Testability 1.00 / **Traceability 0.75 → 1.00**) — Tier M 임계 0.80 통과

## 1. D1 종결 — RESOLVED (직접 측정)

`git diff 9a3dd7de3..a09821c98` 실측 — acceptance.md 단일 파일, 공개된 2줄 변경만:

- **(a)** 「1행 이상」 → 「**정확히 1행** 기록되고 — 0행뿐 아니라 **2행 이상도 실패다**(REQ-WC-017-003의 exactly one; 하한만 단정하면 두 줄 변이가 REQ 를 어기고 통과한다)」. iter-1 D1이 지목한 2행 변이가 이제 AC 문언상 **직접 실패**한다 — (a)의 상한 단정이 변이를 소진. REQ-WC-017-003의 "exactly one"이 AC에 고정됨 → Traceability 1.00으로 갱신.
- **(b)** 「그 **모든** 행」 → 「그 **유일한 행**」 — exactly-one 의미와의 붕괴 일관. 접두어 검사는 존재하는 그 한 행에 그대로 적용되어 의미 손실 없음.
- 0+0=0 동어반복 근거 문장(하한 존재 이유)은 원문 그대로 보존 — 하한 단정의 근거가 상한 추가로 지워지지 않았다.

**변이 재판정**: 2행 변이(헬퍼가 본문+요약 2행 기록) → (a) 위반(2행 이상 실패) → AC-WC17-003 FAIL. 전 AC 통과 변이 소멸. **D1 CLOSED.**

## 2. 부수 스캔 — COLLATERAL 0건

delta는 acceptance.md:24(AC-WC17-003 Given/Then 행)·:25(RED 근거 행)의 2줄만 건드린다(-2/+2). REQ 정의(spec.md)·다른 AC(HARD-1..7, AC-001/002/004/005)·추적성 표(§D.3)·DoD 의미 변동 없음.

## 3. Regression Check (prior-iteration defects)

- **D1** (acceptance.md:24 — AC 강도 결손): **RESOLVED** — §1 실측대로 상한 고정.
- **D2** (RED 근거의 관측/추론 미구분): **RESOLVED** — :25에 「이 근거는 코드 판독(정적 추론)이며, 관측된 RED 출력 제출은 run-phase E8 의 몫이다」 추가 확인(diff 실측). run-phase E8이 귀속 몫이라는 구분이 이제 문언에 있다.
- **D3** (SPEC-WEB-CONSOLE-016 draft — 문구 상류 변동 가능성): **accepted-as-optional, recorded** — 본 SPEC 문서 외부 상태로 수리 대상이 아니며 §6-3 연동 규정이 지배. delta가 이 축을 건드리지 않았다(회귀 없음).

## 4. 기계 재측정 (본 실행 귀속)

- `moai spec lint .moai/specs/SPEC-WEB-CONSOLE-017` → `✓ No findings` (2026-09-22 본 감사 실행, 트리 `WT-save-observability` @ `a09821c98`).
- iter-1 판정서 본문 무손상(본 파일 iter-1 절 보존 — 감사 추적).
- 워킹 트리 clean, HEAD `a09821c98` 확인 후 커밋.

## 5. iter-2 증거-수반 보고 (VCI 5-섹션)

- **Claim**: D1 종결 — AC-WC17-003(a)이 REQ-WC-017-003의 exactly one을 고정하고 2행 변이가 문언상 실패한다. 부수 변동 0건. D2도 소진됨. 최종 판정 PASS 1.00.
- **Evidence**: `git diff 9a3dd7de3..a09821c98` 전문(§1 인용) — acceptance.md 2줄; `moai spec lint` ✓ No findings(본 실행).
- **Baseline-attribution**: 트리 `WT-save-observability` @ `a09821c98`, 2026-09-22 iter-2 감사 실행에서 측정.
- **Gaps**: run-phase E8의 관측 RED 출력(테스트 미존재로 plan-phase 불가 — D2 문언이 위임을 명시); CDP 전제값(브라우저 재측정 안 함 — iter-1과 동일).
- **Residual-risk**: §6-3(016 문구 개정 시 가드 리터럴 동반 갱신)이 지켜지지 않으면 가드 리터럴이 상류와 갈린다 — SPEC 문언에 규정돼 있어 run/sync 단계의 규율 사항.

## 6. 최종 처분

- **Verdict: PASS** — skip 계약상 유효(verdict PASS + 1.00 ≥ 0.80 + artifact-hash 기준점 = `a09821c98`; 이후 산출물 무변경이면 run-gate가 skip 판정).
- D3는 기록된 optional 항목으로 본 판정에 영향 없음(M6 — optional 결함은 판정을 만들지 않는다).
- 다음 관문: Implementation Kickoff Approval (M1 운영자 판정 대상 — progress.md §E.1과 일치).
