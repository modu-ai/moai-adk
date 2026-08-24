---
id: SPEC-CODEX-VERDICT-SYNTH-001
title: "codex verdict 합성 — 모르는 서식을 통과로 읽지 않는다"
version: "0.3.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tier: S
tags: "codex, audit-multi, verdict, version-drift, fail-open"
---

# SPEC-CODEX-VERDICT-SYNTH-001 — codex verdict 합성의 관대 편향 제거

## HISTORY

- 2026-08-24 · v0.3.0 · manager-spec · 리드 판정 반영. **Tier M → S** 축소 · 근거를 `.moai/reports/t229/premise-revision.md` 의 7행 실측표로 교체 · §A.2 를 프로세스 랙에서 **바이너리 랙**으로 정정(프로세스 랙 가설 철회) · 보수 채택 순서(구 REQ-CVS-002)를 결함 목록에서 제거해 유지보수 메모로 강등 · AC 를 서식 열거형에서 **속성형** 으로 전환 · 후속 카드 t248 명시.
- 2026-08-24 · v0.2.0 · manager-spec · 리드 정정 4건 반영. 세션 영속화 요구사항 제거(관측으로 강등) · t234/t246 카드 id 명시 · 버전 드리프트 내성을 지배 원칙으로 승격.
- 2026-08-24 · v0.1.0 · manager-spec · 최초 작성. 카드 t229 (Class B).

---

## §0 지배 원칙 [HARD]

> **판정 불가를 통과로 읽지 않는다.**

어댑터가 codex 의 응답 서식을 알아보지 못했을 때 내려야 할 값은 `pass` 가 아니라 `inconclusive` 다. 서식을 못 알아본 것은 **아무것도 관측하지 못한 상태**이지, 통과의 근거가 아니다.

이 원칙의 실무적 귀결이 이 SPEC 의 형태를 정한다: 교정은 **"정규식을 하나 더 추가한다"가 아니라 "모르면 `inconclusive`"** 다. 서식 하나를 더 알게 만드는 수정은 눈앞의 사례만 닫고 구조는 그대로 두어, 다음 서식에서 같은 자리가 다시 뚫린다.

---

## §A 배경

### A.1 일차 근거 — `premise-revision.md`

이 SPEC 의 일차 증거는 `.moai/reports/t229/premise-revision.md` 다. `cause.md` 는 조사 시점에 두 번 스테일했으므로, 두 문서가 어긋나면 `premise-revision.md` 가 이긴다.

그 문서가 현재 트리(`294b4b6ab` = `origin/main`)에서 `synthesizeReviewOutput` 을 직접 호출해 얻은 7행 실측표가 아래 §A.3 의 결함 목록 전부의 근거다. 여기서 다시 유도하지 않는다.

### A.2 [HARD] 라이브 프로브가 `pass` 를 받은 이유 — 바이너리 랙

MCP 도구를 서빙하는 것은 설치 바이너리 `~/go/bin/moai` 다. 그 바이너리가 어느 커밋에서 빌드됐는지 직접 물었고, 관련 수정 2건의 포함 여부를 기계로 판정했다.

```
$ ~/go/bin/moai version
 v3.1.2   a1b1ca696   built 2026-08-24T11:19:07Z

$ git merge-base --is-ancestor f505955a9 a1b1ca696 ; echo $?   # t178 → 1
$ git merge-base --is-ancestor 4505df411 a1b1ca696 ; echo $?   # t186 → 1
$ git rev-list --count a1b1ca696..origin/main                  # 259
```

`f505955a9`(t178)·`4505df411`(t186) **둘 다 설치 바이너리의 조상이 아니다.** 그 바이너리는 `origin/main` 보다 **259 커밋** 뒤처져 있다. 즉 프로브는 두 수정 이전의 코드를 측정했다 — 이것이 **바이너리 랙**이며, 프로브 관측 자체는 유효하되 그 대상이 현재 main 이 아니었다는 뜻이다.

**프로세스 랙 가설은 채택하지 않는다.** `ps` 에 `moai mcp-server` 프로세스가 한 건도 없었고, 이름이 비슷해 보이는 `moai-mcp-imweb` / `moai-mcp-smartstore` / `moai-mcp-cafe24` / `moai-mcp-threads-poster` 는 **Claude 데스크톱 플러그인으로 이 저장소와 무관**하다. 이들을 moai-adk 의 프로세스로 귀속해서는 안 된다.

이에 따라 `cause.md` 의 F2 는 정정(현재 트리의 정규식은 2개), F5 는 철회(현재 트리는 `Verdict: <word>` 라벨을 읽는다), F4 는 대상 정정(구 바이너리를 측정).

### A.3 남은 결함 — 실측 기준 3건

`premise-revision.md` §3 의 표를 그대로 인용한다.

| # | 결함 | 실측 근거 (실측표 행) | 성격 |
|---|---|---|---|
| G1 | 점수 표기(`FAIL 0.75 / 1.00`)를 명시 verdict 로 인식하지 못해 `pass` 로 합성 | 2행 | live defect |
| G3 | 아는 서식이 하나도 안 맞으면 `pass` (adversarial 포함) | 4행·5행 | live defect — **구조적 원인** |
| G4 | 두 신호가 갈렸다는 사실이 결과 어디에도 기록되지 않음 | `converge()` 가 `Summary` 를 판정에 안 씀 (`mcp_convergence.go:135`) | live gap |

**G1 은 G3 의 한 사례다.** 점수 표기를 못 알아본 것이 곧 "아는 서식이 안 맞은" 상황이고, 그때 `pass` 로 떨어진 것이 G3 다. 따라서 G1 만 닫는 수정은 G3 을 남긴다 — 이 구분이 §C 의 AC 형태를 정한다.

핵심 실측 1행: **t197 기록에 나타난 바로 그 서식(`FAIL 0.75 · 차단 2건`)이 현재 main 에서도 `pass` 로 합성된다.** 카드의 핵심 결함은 좁아졌을 뿐 사라지지 않았다.

### A.4 구조적 원인 — 판정이 CLI 한 버전의 출력 서식에 묶여 있다

| 항목 | 값 |
|---|---|
| 설치된 codex | 0.149.0 |
| `codexFindingBullet` 주석이 명시한 눈금 기준 | 0.146.1 의 review-mode 출력 |
| 그 정규식이 실제 적용되는 경로 | 서식을 전혀 지정하지 않는 adversarial 프롬프트 |

한 버전의 출력 관례로 눈금을 맞춘 판별기를 서식 계약이 없는 응답에 적용하고 있다. t178·t186 은 서식을 **하나 더 알게** 했을 뿐, 모르는 서식이 통과로 떨어지는 구조(`verdict := "pass"`, `mcp_codex.go:1145`)는 그대로다.

### A.5 결함이 아닌 사항 — 보수 채택의 구현 형태 (유지보수 메모)

보수 채택이 명시적 순위 테이블이 아니라 대입 순서로 구현돼 있다(`mcp_codex.go:1144-1156`). 이것은 **결함이 아니다** — 실측표 7행 어디에도 현재 조합이 잘못된 값을 낸 반례가 없고, 현 구현은 이미 fail 편향으로 올바르게 동작한다. 신호가 셋 이상으로 늘어날 때 취약해질 수 있다는 **유지보수성 지적**으로만 남기며, 이 SPEC 의 요구사항으로 두지 않는다.

### A.6 모드 구분이 필요한 이유 [HARD]

관대 기본값을 무조건 없애면 **정상 통과 경로가 깨진다**. native review-mode(`review/start`)의 무불릿 응답은 codex 가 실제로 "차단 사유 없음" 을 말한 것이고, Stop 훅 게이트 `HandleCodexReviewGate`(`internal/cli/codex_review_gate.go:66`)의 clean-pass 가 여기에 걸려 있다(실측표 6행 = 보존 대상). adversarial-mode(`turn/start`)의 미인식 서식은 반대로 **아무것도 관측되지 않은 상태**다.

seam 은 이미 존재한다. `synthesizeReviewOutput` 의 유일한 프로덕션 호출자는 `runTurn`(`mcp_codex.go:680`)이고, 그 시그니처가 이미 `method` 를 받는다. `codexMethodReviewStart` = native, `codexMethodTurnStart` = adversarial.

---

## §B 요구사항 (GEARS)

결함 요구사항 3건(REQ-CVS-001~003) + 회귀 방어 1건(REQ-CVS-004). 회귀 방어는 새 결함이 아니라 기존 정상 동작의 보존 의무다.

### REQ-CVS-001 — 미인식 서식은 통과가 아니다 (G3, 구조적 원인)

While the review turn ran in adversarial mode (`turn/start`), the system shall synthesize `inconclusive` for any review body whose format matches none of the known verdict signals — regardless of which codex CLI version produced it, and regardless of how many signal formats are known at the time.

### REQ-CVS-002 — 점수 표기 판정 인식 (G1, REQ-CVS-001 의 한 사례)

The system shall recognize a verdict stated in codex's score form (`FAIL <score>` / `PASS <score>` / `INCONCLUSIVE <score>` at the head of a line) as a body-stated verdict, in addition to the existing `Verdict: <word>` label form.

### REQ-CVS-003 — 신호 불일치 기록 (G4)

Where two verdict signals diverge within one backend's review body, the system shall record the divergence on the review output and adopt the more conservative of the two, and the convergence engine shall set `disagreement_flag` and name the divergence in `residual_risk_note`.

### REQ-CVS-004 — 회귀 방어

The system shall keep synthesizing `pass` for a bullet-less clean review on the native review path (`review/start`), and shall not change `codex_task`'s returned output text.

---

## §C AC 형태에 대한 구속 [HARD]

**AC 는 "새 서식 하나를 더 읽는다" 가 아니라 "임의의 미인식 입력이 `pass` 로 떨어지지 않는다" 를 걸어야 한다.**

서식 목록으로 AC 를 쓰면 그 목록은 구현이 대상으로 삼은 서식의 목록과 같아지고, 구현은 자기가 읽도록 만든 것만 통과시켜 AC 를 만족시킨다. 그러면 다음 서식에서 같은 자리가 다시 뚫린다. **속성이 요구사항이고, 서식 corpus 는 그 속성의 증인일 뿐이다.**

구체적 형태는 acceptance.md §B 가 정의한다. corpus 에 케이스를 하나 더 넣는 일이 단언문 수정을 요구하면 그 AC 는 속성형이 아니다.

---

## §D 후속 카드와의 순서 — t234 (= GitHub #1632)

**이 SPEC 이 먼저 착지한다.** 리드가 t234 를 이 카드가 끝날 때까지 보류하기로 결정했고, 이 SPEC 은 `Findings: []Finding{}` 하드코딩(`mcp_codex.go:1152`)을 **그대로 둔다**.

t234 는 이 SPEC 이 손대는 것과 **같은 함수** `synthesizeReviewOutput` 을 Findings 추출 축에서 다시 고친다. 이 사실을 SPEC 본문에 산문으로 적어 두는 이유는, 코드 주석은 리팩터링에서 사라질 수 있고 그러면 다음 편집자가 두 축이 한 함수에서 만난다는 것을 알 길이 없어지기 때문이다.

t234 착수자에게 남기는 메모: 이 SPEC 이 `synthesizeReviewOutput` 의 시그니처를 `(reviewText, method string)` 으로 바꾼다. t234 는 바뀐 시그니처 위에서 반환 구조의 `Findings` 필드만 채우면 되며, 시그니처를 되돌리면 모드 구분이 사라져 §0 의 원칙이 깨진다.

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — Findings 추출 (t234 / GitHub #1632)

- `synthesizeReviewOutput` 의 `Findings: []Finding{}` 하드코딩은 이 SPEC 의 대상이 아니다. 후속 카드 **t234** 소관이며 리드가 이 카드 착지까지 보류한다. 순서와 seam 충돌은 §D 참조.

### Out of Scope — codex 백엔드의 트리 오독 (t246)

- 워크트리 안에서 실행한 audit 이 primary 체크아웃을 리뷰한 관측(cause.md F9)은 카드 **t246** 소관이다.

### Out of Scope — audit 출력에 산출 바이너리 커밋 기록 (t248)

- §A.2 가 드러낸 파급: 259 커밋 뒤처진 바이너리가 이 라운드의 모든 audit 호출을 서빙했고, **어느 판정이 현재 코드를 측정한 것인지 판정 출력만으로는 알 수 없다**. audit 결과에 산출 바이너리의 커밋을 함께 기록하는 일은 카드 **t248** 소관이며(리드 등록, t246 연계), 이 SPEC 은 손대지 않는다.

### Out of Scope — 수렴 결과 영속화 (관측만)

- `.moai/state/audit-multi/` 가 저장소 전체에서 0건이라는 소급 스캔 결과(`.moai/reports/t229/retro-sweep.md`)는 이 SPEC 의 대상이 아니다. 후속 카드 개설 여부는 리드 판단이다.
- 이 0건이 뜻하는 것은 **볼 수 있는 기록이 거기까지** 라는 것이다 — 증거 도달 범위의 한계이지, 과거에 문제가 없었다는 뜻이 아니다.
- 따라서 이 SPEC 은 `session_id` 배선을 요구사항으로 두지 않는다.

### Out of Scope — GLM 백엔드

- GLM 경로(`performGLMAudit`)의 판정 합성은 손대지 않는다. 이 SPEC 은 codex 경로의 합성 seam 에 한정된다.

---

## §F 제약

- 언어: Go. 대상 패키지 `internal/cli`. 방법론: TDD.
- 검증: `go test ./internal/cli/... -timeout 600s`. **`go test ./...` 금지** (저장소 규율).
- fail-open 불변식 유지: 어떤 경로도 hard error 를 반환하지 않는다.
- 독립성 불변식 유지: `backendCallFn` 시그니처에 verdict 를 실어 보내지 않는다(`mcp_convergence.go:368`).
- **검증은 `go test` 로 한다.** MCP 라이브 프로브를 근거로 쓰지 않는다 — §A.2 의 바이너리 랙이 그 경로를 신뢰할 수 없게 만든다.

---

## §G 참조

- **일차 근거**: `.moai/reports/t229/premise-revision.md` (7행 실측표)
- 원인 보고서(부분 스테일): `.moai/reports/t229/cause.md`
- 라이브 프로브 원문: `.moai/reports/t229/live-probe-body.txt`
- 소급 스캔: `.moai/reports/t229/retro-sweep.md`
- 선행 SPEC: `SPEC-AUDIT-MULTI-MODEL-001`
- 선행 카드: t178 (`f505955a9`), t186 (`4505df411`)
- 후속 카드: t234 (= GitHub #1632) · t246 · t248
