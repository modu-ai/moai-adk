# SPEC 감사 보고서 — SPEC-CTX-BLIND-DOUBLE-001

- **반복**: 2/2 (Tier M 상한 2회, `harness.plan_audit_tier_ceilings`)
- **판정**: **FAIL** (단일 원인 — 1회차 결함 D2 가 3개 인용처 중 1곳에서 미해소)
- **종합 점수**: **0.91** (1회차 0.84 → **+0.07**, 점수 회귀 없음 → STOP 에스컬레이션 미발동. Tier M 통과선 0.80)
- **감사 트리**: `.claude/worktrees/t539`, 브랜치 `WT-ctx-blind-double`, HEAD `52f863f36`
- **감사자 격리**: 저작자의 추론 맥락은 M1 Context Isolation 에 따라 무시했다.

### 이력

| 회차 | 판정 | 점수 | 요지 |
|---|---|---|---|
| 1 | FAIL | 0.84 | MP-7 미해소 `[NEEDS CLARIFICATION]` 1건 + 재현되지 않는 집계 2건(D1·D2) + 경계가 열린 후보 목록(D3) + RED 없는 인수 기준 3건(D4) 외 8건 |
| 2 | **FAIL** | **0.91** | must-pass **7/7 PASS**(MP-7 해소). D0·D1·D3·D4·D5·D6·D7·D8·D9·D10·D12 **해소 확인**. **D2 는 `sweep.md` 에서 미해소** — 3개 산출물이 정본으로 지목하는 문서에 반증된 근거가 그대로 남아 있고, 같은 문서가 자기 자신과 모순한다 |

> **판정을 한 줄로**: 이번 회차의 수리는 실질적으로 훌륭하다 — 84건 열거는 `sweep-test.tsv` 와 **위치·타입.메서드까지 100% 일치**하고, 42건 재계수는 **정확히 재현**되며, D4 는 `verification-completeness.md` §2.1 을 인용해 정확히 처분됐다. FAIL 은 새 결함이 아니라 **1회차 D2 의 잔여분** 때문이다: `sweep.md` §E2 표가 스윕에서 빠져 반증된 `ctx.Done 1` 이 살아남았고, 그 결과 `sweep.md` 는 48행과 110행이 서로 다른 말을 한다. 고칠 곳은 3줄이다.

---

## 감사자 자기 정정 — 1회차 보고서의 결함 3건

> 이 절을 먼저 놓는 이유: 아래 3건 중 2건은 manager-spec 이 옳았고 내가 틀렸으며, 1건은 **내 오류가 3개 산출물로 전파됐다**. 감사자의 주장도 관측이어야 한다는 규율(`verification-claim-integrity.md` §1)은 감사자 자신에게 먼저 적용된다.

**A1. 1회차 D6 의 "5행" 은 틀렸다 — 실제 6행. manager-spec 이 옳다.**
원인은 내 명령의 절단이다. 1회차에 `diff … | head -20` 을 돌렸고, 한 행의 차이가 `diff` 출력 4줄을 차지하므로 **20줄 = 정확히 5행**에서 잘렸다. 잘린 자리가 마침 경계와 일치해 출력이 완결돼 보였다.
```
$ diff .moai/reports/t539/sweep-raw.tsv .moai/reports/t539/sweep-test.tsv   ← 절단 없이 재실행
… (6개 훙크) …
412c412
< funclit	update/checker_test.go:724	-	-	*http.Request r	yes
---
> funclit	update/checker_test.go:724	-	-	*http.Request r	no
--- '^<' 행 수: 6
```
`update/checker_test.go:724` 가 내 목록에서 빠져 있었다. **이것은 내가 이 SPEC 에 지적했던 것과 같은 부류의 결함이다** — 경계 잘린 출력을 완결된 집합으로 읽은 것. SPEC 은 이미 6행으로 고쳐 적었다(`spec.md:234-236`, `research.md:25-38`, `sweep.md:7`), 그리고 내 오류를 명시적으로 기록해 두었다(`research.md:36-38`). 정확하다.

**A2. 1회차의 "`CoverageIncomplete` 15건은 단일 파일 호출 위양성" 은 틀렸다 — 그러나 manager-spec 의 "편집으로 설명되지 않는다" 도 틀렸다.**
두 가설 모두 틀렸고, 뮤턴트 프로브가 판정했다. 먼저 호출 축을 배제했다 — 두 호출이 지금 같은 값을 낸다:
```
$ moai spec lint .moai/specs/SPEC-CTX-BLIND-DOUBLE-001/spec.md   (1회차와 동일 호출)  → 0 error(s), 5 warning(s)
$ cd .moai/specs/SPEC-CTX-BLIND-DOUBLE-001 && moai spec lint spec.md              → 0 error(s), 5 warning(s)
```
그다음 `/tmp` 사본에 뮤턴트를 넣어 **원인을 격리**했다(SPEC 트리는 건드리지 않았다):
```
control (현재 사본)                                        → 0 error(s),  5 warning(s)
mutant A: acceptance.md 에서 리터럴 'maps ' 만 제거          → 0 error(s), 21 warning(s)
mutant B: acceptance.md 를 통째로 삭제                       → 0 error(s), 21 warning(s)
mutant C: 모든 매핑을 존재하지 않는 REQ-CBD-999 로 재지정      → 0 error(s), 18 warning(s)
```
**메커니즘이 확정됐다**: 커버리지 규칙은 `maps` 를 담은 **줄**을 매핑 줄로 읽고 그 줄의 REQ id 를 커버로 센다. 그래서 (i) `maps` 가 없는 acceptance.md 는 **acceptance.md 가 아예 없는 것과 구분되지 않고**(A ≡ B ≡ 21), (ii) 토큰이 있으면 id 는 실제로 해석된다(C 에서 999 로 바꾼 13건이 uncovered 로 돌아왔고, `maps` 뒤 두 번째 id 3건 — 002/005/007 — 만 살아남아 정확히 16−3=13 이다).
즉 15→0 의 델타는 **편집으로 완전히 설명된다**: manager-spec 이 표에 `maps ` 접두사를 넣었기 때문이다. 1회차에 내가 "위양성으로 확정" 이라고 쓴 것은 실행하지 않은 판정이었다.

**A3. `app.go:212` 는 내 오류이고, 3개 산출물로 전파됐다.**
1회차 보고서(`plan-audit.md:54`)에서 42건의 누락분을 "`web/app.go:212` 와 `web/glmkey.go:115`" 라고 적었다. **`web/app.go:212` 는 `funclit` 행이다.** 42건 집합은 `method` × `no` 이므로 그 자리에 들어가는 행은 `web/app.go:256` 이다:
```
$ awk -F'\t' '$2 ~ /^web\/app\.go/ {...}' .moai/reports/t539/sweep-prod.tsv
funclit  web/app.go:212  -.-                  [*http.Request r]  verdict=no
method   web/app.go:256  app.selectedProfile  [*http.Request r]  verdict=no
```
manager-spec 은 내 값을 그대로 옮겼고, 지금 `spec.md:167` · `research.md:197` · `sweep.md:110` 세 곳에 `app.go:212` 가 들어 있다. **건수(2)는 옳고 줄 번호만 틀렸다.** 아래 E2 로 접는다.

---

## Must-Pass 결과 — 7/7 PASS

| # | 기준 | 결과 | 근거 (이 트리, 이 회차 재실행) |
|---|---|---|---|
| MP-1 | REQ 번호 일관성 | **PASS** | `REQ-CBD-001` ~ `REQ-CBD-016` **16개**, 결번·중복 없음, 3자리 padding 일관. `REQ-CBD-016`(추적성)은 D10 수리로 신설 |
| MP-2 | GEARS 형식 준수 | **PASS** | **요구사항 층(`spec.md` §2)에 대해 판정했다.** `acceptance.md` 의 `AC-CBD-*` 는 Given-When-Then 이며 검증 층의 올바른 형식이므로 Group 4 에서 채점했고 여기서 감점하지 않았다. 16개 라벨 전부 정본 modality — 1회차 D7 의 `(event-detected)` 2건이 `(event-driven)` 으로 교체됐다(`spec.md:135,137`). 린터의 `ModalityMalformed` 5건은 영어 `SHALL` 정규식이 한국어 종결형을 읽지 못하는 로케일 산물로, 부류로 해석했다 |
| MP-3 | YAML frontmatter 유효성 | **PASS** | 정본 12필드 전부 존재, 거부 별칭 0건, `FrontmatterInvalid` 0건. `version: "0.2.0"` 으로 갱신되고 HISTORY 에 0.2.0 행이 추가됐다. `tier: M` · `era: V3R6` 은 문서화된 선택 필드(스키마 § Optional Fields / `lifecycle-sync-gate.md`), `phase: "v3.1.4 target"` 은 릴리스 라벨이지 lifecycle 토큰이 아니다 |
| MP-4 | §22 언어 중립성 | **N/A (자동 통과)** | `internal/` Go 코드로 한정된 단일 언어 SPEC |
| MP-5 | D7 교차 SPEC 조정 | **PASS** | `/usr/bin/grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` → 자기 자신 외 참조 0건 |
| MP-6 | D8 크로스 플랫폼 규율 | **PASS (자동)** | `/usr/bin/grep -c 'syscall'` → 5개 산출물 전부 **0** |
| MP-7 | 해소 게이트 | **PASS** (1회차 FAIL → 해소) | `/usr/bin/grep -c 'NEEDS CLARIFICATION'` → `spec.md:0 plan.md:0 acceptance.md:0 research.md:0 progress.md:0`. 운영자 결정이 `plan.md` §F "수리의 소속 — 해소됨"(`:256-270`)과 `spec.md` §3 네 번째 Out of Scope 절, `REQ-CBD-011` 본문에 반영됐다 |

**M5 방화벽**: must-pass 실패 없음. 이번 FAIL 은 방화벽이 아니라 **Retry Loop Contract**(1회차 열거 결함의 미해소)에서 나온다.

---

## 부문 점수 (0.00-1.00)

| 부문 | 1회차 | 2회차 | 밴드 | 근거 |
|---|---|---|---|---|
| 명확성 | 0.85 | **0.93** | 0.75~1.0 | 후보 집합이 닫혔고(`REQ-CBD-009` + `plan.md` §F 84건), modality 라벨이 정본화됐으며, 수리 소속이 `REQ-CBD-011` · `spec.md` §3 · `plan.md` §F 세 곳에 일관되게 기술됐다 |
| 완전성 | 0.90 | **0.90** | 0.75~1.0 | 42건 표가 재계수 값으로 교체되고 정정 사유까지 담겼다. 상승이 없는 이유는 E1 — 정본으로 지목된 `sweep.md` 의 (a)축 표가 스윕에서 빠졌다 |
| 검증가능성 | 0.72 | **0.88** | 0.75~1.0 | 세 축이 모두 개선됐다. (1) `AC-CBD-008` 이 닫힌 84건 집합 위에서 반증 가능해졌고 자기 반증가능성 근거까지 본문에 담겼다(`acceptance.md:134-138`). (2) D4 가 `verification-completeness.md` §2.1 undecidable disposition 을 인용해 3건을 regression-guard 로 처분하고 DoD 에서 "통과로 기록하지 않는다"를 명시했다. (3) `-list` 출력 파일 보존이 `plan.md` §D · §F M2-4 · `AC-CBD-007` 에 배선됐다 |
| 추적성 | 0.90 | **0.92** | 0.75~1.0 | 16↔16 리터럴 ID 양방향 차집합 **공집합**, `AC-CBD-015` 가 `REQ-CBD-016` 에 걸렸고(D10), `REQ-CBD-009` 의 절 참조가 `§6` 으로 교정됐다(D12). 만점이 아닌 이유는 E3 — 린터의 커버리지 초록이 토큰 의존이라 약한 신호다 |

**종합 = 조화평균(0.93, 0.90, 0.88, 0.92) = 0.91.** 1회차 0.84 대비 상승이므로 점수 회귀 STOP 조항은 발동하지 않는다.

---

## 회귀 점검 — 1회차 결함 13건의 처리 상태

| 1회차 | 상태 | 재측정 근거 |
|---|---|---|
| **D0** MP-7 미해소 마커 | **RESOLVED** | 5개 산출물 전부 0건. 운영자 결정이 3곳에 반영됨 |
| **D1** 42건 내역이 43으로 합산 | **RESOLVED** | `spec.md:162-174` 표를 `sweep-prod.tsv` 로 전수 재계수: hook `Handle` **18** ✓ (`^hook/` 21 − `hook/quality` 3, 18건 전부 메서드명이 `Handle`), `hook/quality` **3** ✓, web 핸들러 **14** ✓ (screens 7 + profile_crud 4 + handlers 3), web 기타 **2** ✓, `statusline/git.go` **1** ✓, `update/local.go` **3** ✓ (`:63 CheckLatest`/`:180 Download`/`:190 Replace` — 셋 다 정확), `update/updater.go` **1** ✓ (`:126 updaterImpl.Replace` — 정확). **합 42 = 측정된 총계** ✓. 줄 번호 1건만 잔여(E2) |
| **D2** LSP (a)축 근거가 테스트 파일에서 새어 나옴 | **부분 해소 — `sweep.md` 미수리** | `research.md:83` 은 정확히 교체됨(`request.go:70 ctx.Err()`, `:53 t.Call(ctx, …)`, "`ctx.Done` 은 이 파일에 0건"). `spec.md:220-224` 는 잔여 위험 3번에 사례로 승격. **그러나 `sweep.md:48` 은 `| LSP transport `request.go` | 예 | `ctx.Done` 1 |` 그대로** → E1 |
| **D3** 후보 목록 경계 열림 | **RESOLVED (전수 검증)** | 아래 § 84건 전수 대조 참조 — 84/84 위치·타입.메서드 일치, `consults=yes` 0건, 중복 0건 |
| **D4** M1 인수 기준 3건 RED 부재 | **RESOLVED (모범적)** | `acceptance.md:8-24` 표에 `분류` 열 신설, 3건 **regression-guard**, 나머지 12건 release-blocking. `:26-39` 분류 주석이 `verification-completeness.md` §2.1 을 인용하고 "통과로 기록되지 않는다"를 명시. DoD `:204` 가 "깨지지 않았음만 확인한다"로 소비 규칙까지 고정 |
| **D5** "GLM 테스트 파일 7개" 분모 | **RESOLVED** | `spec.md:88-91` 이 분모 22 + 근거 명령 + "분모를 22 로 넓혀도 결론은 바뀌지 않는다"를 담음 |
| **D6** `sweep-raw.tsv` 지위 미표기 | **RESOLVED (내 오류 정정 포함)** | `spec.md:233-237` · `research.md:23-38` · `sweep.md:7` 3곳에 폐기 표기 + 6행 목록. 내 5행 오류도 `research.md:36-38` 에 기록 |
| **D7** `(event-detected)` 라벨 | **RESOLVED** | 16개 라벨 전부 정본 modality |
| **D8** `-list` 기준선 증거 부재 | **RESOLVED (절차로)** | `plan.md:42-45` · `:116-117`, `AC-CBD-007:109-116` 에 셀렉터별 로그 보존이 배선되고, plan 단계 간극이 인용 블록으로 명시됨 |
| **D9** `deployer.go` (a) 계수 4 vs 2 | **RESOLVED** | `plan.md:53-55` 가 3토큰 계수 규칙을 고정, `research.md:87` 이 2로 정정 + 4의 유래 설명. **단 `sweep.md:51` 은 4 그대로** → E1 에 포함 |
| **D10** `AC-CBD-015` 고아 | **RESOLVED** | `REQ-CBD-016` 신설, `AC-CBD-015` 가 `maps REQ-CBD-016` |
| **D11** 보고서 경로 관례 | **미조치(의도적, optional)** | 위임 지시대로 SPEC 디렉터리 유지. `planArtifactNames` 밖이라 skip-eligibility 해시 무영향 — 유지 판단 타당 |
| **D12** `REQ-CBD-009` 절 참조 | **RESOLVED** | `spec.md:129` `research.md **§6**`, `plan.md:10` 도 `§6`. `research.md` 실제 §6 = "run-phase 후보" ✓ |

**12/13 해소, 1건(D2) 부분 해소.**

---

## 이번 회차 발견

**E1. `sweep.md` §E2 표가 스윕에서 빠졌다 — 반증된 근거가 정본에 살아 있고, 문서가 자기 자신과 모순한다 — `.moai/reports/t539/sweep.md:48,51,54` — Severity: major — Class: blocking — (1회차 D2 의 미해소 잔여분)**

세 산출물이 `sweep.md` 를 **정본**으로 지목한다: `spec.md:231`("조사 자료: … `sweep.md`(정본)"), `research.md:3`("정본: … `sweep.md`. 이 문서는 그 측정을 SPEC 산출물로 고정한 것이며, **새 사실을 주장하지 않는다**"), `plan.md:287`("조사 정본"). 그런데 정정은 종속 문서에만 들어갔다. 현재 `sweep.md` §E2 (a)축 표가 담고 있는 값:

| `sweep.md` 행 | 현재 값 | 실측 / 정정본 | 관련 |
|---|---|---|---|
| `:48` | `LSP transport request.go \| 예 \| ctx.Done 1` | **`request.go` 의 `ctx.Done` 은 0건.** 1건은 `request_test.go:154` 의 히트 | D2 |
| `:51` | `template deployer.go \| 예 \| 4` | 3토큰 규칙으로 **2** (`plan.md:53-55` 가 고정한 규칙) | D9 |
| `:54` | `hook 핸들러 21종 (hook/*.go Handle)` | `Handle` 은 **18**건 | D1 |

재측정:
```
$ /usr/bin/grep -c 'ctx\.Done' internal/lsp/transport/request.go internal/lsp/transport/request_test.go
internal/lsp/transport/request.go:0
internal/lsp/transport/request_test.go:1
$ /usr/bin/grep -n 'request.go\|request_test' .moai/reports/t539/sweep.md
48:| LSP transport `request.go` | 예 | `ctx.Done` 1 |          ← 정정 주석 없음
```
가장 무거운 것은 **같은 문서 안의 모순**이다. `sweep.md:54` 는 "hook 핸들러 21종"이라 적고, 56줄 뒤 `:110-112` 는 "hook `Handle` **18**"이라 적으면서 "21 은 `hook/quality` 3건을 두 번 센 값"이라고 그 21을 명시적으로 반박한다. 한 문서가 두 값을 동시에 주장한다.

이것은 `verification-completeness.md` §3(cross-layer revision sweep — "정정은 그것이 시작된 파일에서 끝나지 않는다")이 정확히 겨냥하는 형태이고, `research.md:3` 이 선언한 종속 관계를 뒤집는다 — 종속 문서가 정본보다 정확해졌다.

**필요한 수정** (3줄):
- `:48` → `` | LSP transport `request.go` | 예 | `request.go:70` `ctx.Err()`, `:53` `t.Call(ctx, …)` — `ctx.Done` 은 이 파일에 0건 | ``
- `:51` → `4` 를 `2`(`:167`, `:354`)로
- `:54` → `hook 핸들러 21종` 을 `hook `Handle` 18종`으로 (`:110` 과 일치시킨다)

**E2. `app.go:212` 는 `app.go:256` 이어야 한다 — `spec.md:167` · `research.md:197` · `sweep.md:110` — Severity: minor — Class: blocking — (감사자 오류의 전파, A3)**

42건 집합은 `method` × `no` 인데 `web/app.go:212` 는 `funclit` 행이다. 해당 method 행은 `web/app.go:256 app.selectedProfile` 이다(위 A3 의 `awk` 출력). **건수 2 는 옳고 줄 번호만 틀렸다** — 42 합계에는 영향이 없다.
**필요한 수정**: 세 곳의 `app.go:212` → `app.go:256`. (원본 오류가 내 1회차 보고서에 있었음을 함께 기록하면 다음 독자가 출처를 추적할 수 있다.)

**E3. `moai spec lint` 의 커버리지 초록은 토큰 의존이라 약한 신호다 — Severity: minor — Class: optional**

A2 의 뮤턴트 프로브가 보인 대로, 커버리지 규칙은 `maps` 를 담은 줄만 매핑으로 읽는다. 접두사가 없는 acceptance.md 는 **acceptance.md 부재와 구분되지 않는다**(mutant A ≡ mutant B ≡ 21 warnings). 지금 표가 `maps` 를 달고 있으므로 초록은 정당하지만, 그 초록이 보증하는 것은 "매핑 줄이 존재한다"이지 "매핑이 옳다"가 아니다.
이 SPEC 의 추적성 판정은 린터가 아니라 **감사자의 리터럴 ID 집합 차분**(아래 Evidence E5)에 근거한다. 조치 불필요 — 다만 향후 이 SPEC 계열이 린터 초록을 추적성 증거로 인용하지 않도록 남긴다.

---

## 84건 전수 대조 (D3 검증 — 이번 회차의 핵심 측정)

`plan.md` §F M3(`:140-245`)의 열거를 기계 추출해 `sweep-test.tsv` 와 대조했다.

```
enumerated count: 84          duplicates: 0
TSV method rows with consults != yes: 153

A) 열거에 있으나 TSV 의 method/non-yes 에 없는 항목  →  0
B) 열거 항목 중 TSV verdict 가 yes 인 항목            →  0
C) 열거의 `타입.메서드` 라벨이 TSV 와 불일치한 항목    →  0
```

세 조건 전부 0이다. 즉 84건은 (1) 실재하는 행이고, (2) 전부 "컨텍스트를 보지 않는" 대역이며, (3) 라벨까지 원본과 일치한다. 후보군별 합도 검산된다 — 4 + 4 + 46 + 20 + 10 = **84**.

이로써 `AC-CBD-008` 의 "미측정 0건" 은 **닫힌 집합 위의 반증 가능한 조건**이 됐다. 1회차 D3 의 핵심 지적이 해소됐다.

---

## 증거 5절

### Claim

SPEC-CTX-BLIND-DOUBLE-001 v0.2.0 은 must-pass 7/7 을 통과하고 루브릭 종합 0.91 을 얻는다. 1회차 결함 13건 중 12건이 해소됐고, D2 가 `sweep.md` 한 곳에서 미해소로 남아 Retry Loop Contract 에 따라 판정은 FAIL 이다. 이번 회차에 검증한 실질 수리 — 84건 전수 열거, 42건 재계수, D4 처분 — 은 전부 이 트리에서 재현됐다.

### Evidence

**E1 — 84건 전수 대조**: 위 § 참조. `comm -23` / `comm -12` 3방향 전부 0.

**E2 — 42건 재계수**:
```
$ awk -F'\t' '$1=="method" && $6=="no" {print $2"\t"$3"."$4}' sweep-prod.tsv | wc -l   → 42
hook/* 21 (그중 Handle 18, quality 3) · web/* 16 (핸들러 3파일 14 + app.go:256, glmkey.go:115)
statusline/git.go 1 · update/local.go 3 (:63,:180,:190) · update/updater.go 1 (:126)
18+3+14+2+1+3+1 = 42 ✓
```

**E3 — 정본 vs 종속 문서 모순**:
```
$ /usr/bin/grep -n 'request.go' .moai/reports/t539/sweep.md   → 48:| LSP transport `request.go` | 예 | `ctx.Done` 1 |
$ /usr/bin/grep -c 'ctx\.Done' internal/lsp/transport/request.go        → 0
$ sed -n '54p;110p' .moai/reports/t539/sweep.md   → "hook 핸들러 21종" / "hook `Handle` **18**"
```

**E4 — 린터 뮤턴트 프로브** (`/tmp` 사본, SPEC 트리 무변경):
control 5 / mutant A(‘maps ’ 제거) 21 / mutant B(acceptance.md 삭제) 21 / mutant C(REQ-CBD-999 재지정) 18.

**E5 — 추적성 집합 차분**:
```
$ comm -3 <(grep -o 'REQ-CBD-[0-9]\{3\}' spec.md|sort -u) <(grep -o 'REQ-CBD-[0-9]\{3\}' acceptance.md|sort -u)
(무출력 — 대칭차 공집합, 16↔16)
AC 표 15행 / H3 절 15개 일치
```

**E6 — must-pass 동사**:
```
$ /usr/bin/grep -c 'NEEDS CLARIFICATION' spec.md plan.md acceptance.md research.md progress.md → 전부 0
$ /usr/bin/grep -c 'syscall' (동일 5파일)                                                      → 전부 0
$ /usr/bin/grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' *.md | sort -u                             → 자기 자신뿐
$ moai spec lint spec.md                                                                        → 0 error(s), 5 warning(s)
```

**E7 — raw/test 6행 차분** (절단 없이): 위 A1 참조.

### Baseline-attribution

모든 측정은 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t539`, 브랜치 `WT-ctx-blind-double`, HEAD `52f863f36` 에서 이 감사 세션(2026-09-08, 2회차)에 새로 실행됐다. SPEC 문서에서 옮겨 적은 수치는 없다. 린터 뮤턴트는 `/tmp/t539-lintprobe*` 사본에서만 수행했고 SPEC 트리와 `internal/` 은 변경하지 않았다. 커밋하지 않았다.

### Gaps

1. **M2 뮤턴트를 이번에도 재현하지 않았다** — 공유 트리에 고의 훼손을 넣는 행위이며 감사 창은 배타적 쓰기 창이 아니다. M2 생존은 `m2-run-*.log` 를 읽어 확인했을 뿐이다.
2. **M3 84건의 (d)축을 판정하지 않았다.** 열거가 `sweep-test.tsv` 와 일치한다는 것만 확인했다 — 각 후보가 실제로 가려지는지는 run-phase 뮤턴트의 몫이고, 이 감사는 그에 대해 어떤 주장도 하지 않는다.
3. **(a)축 재측정을 이번 회차에 다시 돌리지 않았다.** 1회차에 12행 중 10행 일치를 확인했고, 이번에는 정정된 2행(LSP·deployer)의 서술만 대조했다. 표에 오르지 않은 실물의 오분류 가능성은 여전히 열려 있다.
4. **`funclit` 43건이 "전부" `httptest` 서버 핸들러라는 주장**은 개수만 재현했고 전수 확인하지 않았다(1회차와 동일).
5. **교차 백엔드 2차 의견 없음.** `/usr/bin/grep -rn 'audit_model' .moai/config/` → exit 1(무매치)이므로 미요청이 규정에 맞으나, 이 판정은 단일 감사자(claude) 판정이다.
6. **`golangci-lint` / 전체 스위트 미실행** — plan-phase 감사 대상이 아니며 레인 검증 부하 규율상 금지된다.

### Residual-risk

1. **E1 을 고쳐도 `sweep.md` 에 다른 스테일 값이 남아 있을 수 있다.** 나는 D1·D2·D9 가 건드린 3행만 대조했다. `sweep.md` §E2 표 전체를 정정본 기준으로 다시 훑는 것이 안전하다 — 이 문서는 §5 만 스윕됐고 §E2 는 통째로 빠졌으므로, 빠진 범위가 표 3행에 그친다는 보장이 없다.
2. **정본 지정이 뒤집힌 채로 남는 위험.** `research.md:3` 은 "새 사실을 주장하지 않는다"고 선언하지만 지금은 정본보다 정확하다. E1 을 고치면 해소되지만, 고치지 않고 `research.md` 를 정본으로 승격하는 방향으로 처리하면 세 파일의 상호 참조를 함께 바꿔야 한다.
3. **`AC-CBD-008` 의 84 는 `52f863f36` 에 고정된 수다.** run-phase 가 다른 커밋에서 시작하면 `sweep-test.tsv` 가 달라지고 열거가 어긋난다. `plan.md` §C 사전 점검이 집계 재현을 확인하지만, 84건 열거 자체의 재대조는 절차에 없다 — 사전 점검에 "열거 84건을 `sweep-test.tsv` 와 재대조" 를 넣으면 닫힌다.
4. **점수 0.91 은 조화평균이며 가중치에 민감하다.** 최저 부문(완전성 0.90)과 최고(명확성 0.93)의 폭이 좁아 이번에는 민감도가 낮다.
5. **이 감사도 1회차에 3건을 틀렸다.** 절단된 출력, 실행하지 않은 인과 판정, 행 종류를 섞은 줄 번호 — 셋 다 "관측했다고 적었으나 관측하지 않은" 형태다. 이번 회차의 판정도 같은 위험을 진다.

---

## 권고

**FAIL 은 좁다.** 고칠 것은 `sweep.md` 3줄(E1)과 세 파일의 줄 번호 1개(E2)이며, 둘 다 기계적이다. SPEC 산출물 5종 자체는 이번 회차에 검증한 범위에서 내부적으로 정확하다 — 84/84 전수 일치와 42건 정확 재현은 plan-phase 산출물로서 드문 수준이다.

3회차 재감사가 필요한 범위는 **E1·E2 두 항목의 해소 여부뿐**이다. 전면 재감사가 아니다. 다만 Tier M 상한은 2회이므로, 3회차 진입 대신 다음 중 하나를 권고한다:

1. **(권장) E1·E2 를 즉시 수리하고 오케스트레이터가 해소를 직접 확인한 뒤 Implementation Kickoff Approval 로 진행한다.** 두 항목 모두 판정에 판단이 필요 없는 리터럴 교체이고, 확인 명령은 이 보고서에 그대로 적혀 있다(`/usr/bin/grep -c 'ctx\.Done' internal/lsp/transport/request.go` → 0 이 근거, `sweep.md:48/51/54` 와 `app.go:256` 이 대상). 이 경로에서 감사자 재판정은 필요 없다 — 남은 것은 검증이 아니라 반영이다.
2. 운영자가 3회차 상한 연장을 명시적으로 택한다(드문 선택).

**한 가지만 당부한다.** E1 은 "정정을 한 파일에서 끝냈다"는 형태의 결함이고, 1회차 D2 의 지적에 `sweep.md:47` 이 이미 이름으로 적혀 있었다. 다음 정정에서는 **지적이 이름을 댄 모든 인용처를 목록으로 만들고 하나씩 지워 나가는 편**이 같은 잔여를 막는다 — 이번 회차의 D1 은 세 파일 모두 고쳐졌고 D2 는 두 파일 중 하나만 고쳐졌는데, 두 결함의 차이는 난이도가 아니라 그 목록의 유무였다.
