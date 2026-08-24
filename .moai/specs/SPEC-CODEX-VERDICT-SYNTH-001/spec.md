---
id: SPEC-CODEX-VERDICT-SYNTH-001
title: "codex verdict 합성 — 버전 드리프트에도 판정 불가를 통과로 읽지 않는다"
version: "0.2.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, audit-multi, verdict, convergence, version-drift, fail-open"
---

# SPEC-CODEX-VERDICT-SYNTH-001 — codex verdict 합성의 관대 편향 제거

## HISTORY

- 2026-08-24 · v0.2.0 · manager-spec · 리드 정정 4건 반영. 세션 영속화 요구사항 제거(관측으로 강등) · t234/t246 카드 id 명시 · 버전 드리프트 내성을 지배 원칙으로 승격.
- 2026-08-24 · v0.1.0 · manager-spec · 최초 작성. 카드 t229 (Class B) 의 원인 보고서 `.moai/reports/t229/cause.md` 를 근거로 함. **단, 착수 트리 실측 결과 cause.md 의 F2·F5 는 현재 코드에 대해 스테일** — §A.2 참조.

---

## §0 지배 원칙 [HARD]

> **판정 불가를 통과로 읽지 않는다.**

이 SPEC 의 모든 요구사항은 이 한 줄에서 파생된다. 어댑터가 codex 의 응답 서식을 알아보지 못했을 때 내려야 할 값은 `pass` 가 아니라 `inconclusive` 다. 서식을 못 알아본 것은 **아무것도 관측하지 못한 상태**이지, 통과의 근거가 아니다.

---

## §A 배경

### A.1 관측된 사고

`audit_multi` 의 codex 백엔드가 **본문에서는 통과를 거부했는데 필드로는 `pass` 를 반환**한 사례가 t197 기록에서 3회, t229 라이브 프로브에서 1회 관측됐다. 라이브 프로브 원문은 `.moai/reports/t229/live-probe-body.txt` 이고, 그 본문 첫 줄은 다음과 같다.

```
Verdict: inconclusive — the requested new uncommitted Go file is absent, so a pass would be unsupported.
```

codex 백엔드는 구조화된 verdict 를 반환하지 않는다. `ReviewOutput.Verdict` 는 `internal/cli/mcp_codex.go` 의 `synthesizeReviewOutput` 이 리뷰 산문에서 **합성**한 값이다. 따라서 "필드와 본문의 모순" 이라는 표현은 정확하지 않고, 실제로는 **하나의 본문에서 파생된 값이 원본과 어긋난 것** — 즉 합성 규칙의 결함이다.

### A.2 [HARD] cause.md 의 스테일 구간 — 착수 트리 실측

착수 트리(`WT-audit-verdict-converge`, base `origin/main` @ `294b4b6ab`)에서 직접 측정한 결과, cause.md 의 F2·F5 는 **현재 코드에 대해 참이 아니다**.

| 측정 | 명령/방법 | 관측 |
|---|---|---|
| 본문 명시 verdict 파서 존재 | `grep -n codexStatedVerdict internal/cli/mcp_codex.go` | `mcp_codex.go:1125` 에 존재 (`f505955a9` t178 에서 착지) |
| 라이브 프로브 본문의 합성 결과 | `synthesizeReviewOutput(<live-probe-body.txt>)` 를 임시 테스트로 호출 | **`"inconclusive"`** (`pass` 아님) |
| 불릿 매치 | 동일 호출 | `false` |
| 명시 verdict 매치 | 동일 호출 | `["Verdict: inconclusive", "inconclusive"]` |

라이브 프로브가 `pass` 를 받은 이유는 코드 결함이 아니라 **프로세스 랙**이다. `ps` 로 확인한 `moai mcp-server` 프로세스 다수가 `Sun Aug 23 16:25` 등 t178 착지(`Aug 23 16:59`)·t186 착지(`Aug 23 18:03`) **이전**에 기동됐고, 설치 바이너리(`~/go/bin/moai`, mtime `Aug 24 20:19`)는 해당 정규식을 이미 담고 있다. 즉 프로브는 구 코드의 서버 프로세스를 측정했다.

**이 SPEC 의 범위는 그래서 R1·R2 전체가 아니라 잔여분이다.**

### A.3 [HARD] 구조적 원인 — 판정이 CLI 한 버전의 출력 서식에 묶여 있다

증상(관대 기본값)보다 한 겹 아래에 있는 것이 이것이다.

| 항목 | 값 |
|---|---|
| 설치된 codex | **0.149.0** (`/Users/goos/.local/bin/codex`) |
| `codexFindingBullet` 주석이 명시한 눈금 기준 | **0.146.1 의 review-mode 출력** (`mcp_codex.go:1111-1112`) |
| 그 정규식이 실제 적용되는 경로 | 서식을 전혀 지정하지 않는 adversarial 프롬프트 (`mcp_codex.go:1163`) |

즉 **한 버전의 출력 관례로 눈금을 맞춘 판별기를, 서식 계약이 없는 응답에 적용**하고 있다. codex 가 업그레이드되거나 프롬프트가 서식을 흔들면 판별기는 아무것도 못 알아보고, 그때 떨어지는 값이 `pass` 다. 이 구조에서는 **서식이 바뀔 때마다 조용히 관대해진다** — 서식이 바뀌었다는 신호조차 남지 않은 채로.

따라서 교정은 "정규식을 0.149.0 에 다시 맞춘다" 가 아니다. 그것은 다음 버전에서 같은 사고를 반복한다. 교정의 형태는 **서식 드리프트를 견디는 판정** 이어야 한다 — 아는 서식이 하나도 맞지 않으면 `inconclusive` 로 떨어진다(§0).

### A.4 잔여 결함 (현재 트리 기준, 실측)

| # | 결함 | 근거 |
|---|---|---|
| G1 | 명시 verdict 파서가 `Verdict:` 라벨 형태만 인식. t197 기록에 나타난 **점수 표기**(`FAIL 0.75` / `PASS 0.88`)는 매치되지 않음 | `codexStatedVerdict` 정규식이 `verdict` 라벨을 필수로 요구 (`mcp_codex.go:1125`) |
| G2 | 보수 채택이 **명시적 순서 테이블이 아니라 대입 순서**로 구현됨. `verdict := "pass"` → 명시값 대입 → 불릿이면 `fail` 덮어쓰기 | `synthesizeReviewOutput` 본문 (`mcp_codex.go:1144-1156`) |
| G3 | **관대 기본값 잔존**: 아는 서식이 하나도 안 맞으면 무조건 `pass`. adversarial 모드에서도 동일 | 동일 함수 첫 줄 `verdict := "pass"` |
| G4 | 합성 시 두 신호가 갈렸다는 사실이 **어디에도 기록되지 않음**. `converge()` 는 `PerBackendVerdict.Verdict` 만 읽고 `Summary` 는 판정에 쓰지 않음 | `mcp_convergence.go:135-201` |

G3 이 §A.3 의 구조적 원인이 표면으로 나오는 지점이다.

### A.5 모드 구분이 필요한 이유 [HARD]

관대 기본값을 무조건 없애면 **정상 통과 경로가 깨진다**. native review-mode(`review/start`)의 무불릿 응답은 codex 가 실제로 "차단 사유 없음" 을 말한 것이고, Stop 훅 게이트 `HandleCodexReviewGate`(`internal/cli/codex_review_gate.go:66`)의 clean-pass 가 여기에 걸려 있다. adversarial-mode(`turn/start`)의 무불릿·무판정문은 반대로 **아무것도 관측되지 않은 상태**다. 두 모드는 갈라야 한다.

착수 트리 실측 결과 **seam 은 이미 존재한다**. `synthesizeReviewOutput` 의 유일한 호출자는 `runTurn`(`mcp_codex.go:680`)이고, 그 시그니처가 이미 `method` 를 받고 있다.

| 호출 경로 | 진입점 | `runTurn` 에 전달되는 method |
|---|---|---|
| Stop 훅 리뷰 게이트 | `HandleCodexReviewGate` → `runCodexReviewRPC` | `codexMethodReviewStart` |
| `codex_audit` 단일 백엔드 | `handleCodexAudit`(`mcp_codex.go:1212`) | 기본 `codexMethodReviewStart`, adversarial 요청 시 `codexMethodTurnStart` |
| `audit_multi` codex 백엔드 | `performCodexAudit`(`mcp_convergence.go:412`) | 항상 `codexMethodTurnStart` |
| `codex_task` (리뷰 아님) | `runCodexTaskTurn`(`codex_task.go:97`) | `codexMethodTurnStart` |

따라서 모드 파라미터를 새로 만들 필요가 없다. 선택한 seam 은 plan.md §C 에 기록한다.

---

## §B 요구사항 (GEARS)

### REQ-CVS-001 — 점수 표기 판정 인식

The system shall recognize a verdict stated in codex's score form (`FAIL <score>` / `PASS <score>` / `INCONCLUSIVE <score>` at the head of a line) as a body-stated verdict, in addition to the existing `Verdict: <word>` label form.

### REQ-CVS-002 — 보수 채택 순서

When two or more verdict signals are available for one review body, the system shall adopt the most conservative of them, ordered `fail` > `inconclusive` > `pass`.

### REQ-CVS-003 — 서식 드리프트 내성 (§0 의 직접 구현)

While the review turn ran in adversarial mode (`turn/start`), the system shall synthesize `inconclusive` for a body whose format matches none of the known verdict signals — regardless of which codex CLI version produced it.

While the review turn ran in native review mode (`review/start`), the system shall synthesize `pass` for a body carrying neither a stated verdict nor finding bullets.

### REQ-CVS-004 — 합성 근거 표면화

Where two verdict signals disagree within one backend's review body, the system shall record the disagreement on the review output, and the convergence engine shall set `disagreement_flag` and name the disagreement in `residual_risk_note`.

### REQ-CVS-005 — 회귀 보존

The system shall not change the verdict synthesized for any input covered by `TestSynthesizeReviewOutput_FindingBulletsMapToFail` when the turn ran in native review mode, and shall not change `codex_task`'s returned output text.

---

## §C 후속 카드와의 순서 — t234 (= GitHub #1632)

**이 SPEC 이 먼저 착지한다.** 리드가 t234 를 이 카드가 끝날 때까지 보류하기로 결정했고, 이 SPEC 은 `Findings: []Finding{}` 하드코딩(`mcp_codex.go:1152`)을 **그대로 둔다**.

t234 는 이 SPEC 이 손대는 것과 **같은 함수** `synthesizeReviewOutput` 을 Findings 추출 축에서 다시 고친다. 이 사실을 여기 SPEC 본문에 산문으로 적어 두는 이유는, 코드 주석은 리팩터링에서 사라질 수 있고 그러면 다음 편집자가 두 축이 한 함수에서 만난다는 것을 알 길이 없어지기 때문이다.

t234 착수자에게 남기는 메모: 이 SPEC 이 `synthesizeReviewOutput` 의 시그니처를 `(reviewText, method string)` 으로 바꾼다. t234 는 바뀐 시그니처 위에서 반환 구조의 `Findings` 필드만 채우면 되며, 시그니처를 되돌리면 모드 구분이 사라져 §0 의 원칙이 깨진다.

---

## §D 범위 밖 (Out of Scope)

이 SPEC 은 아래를 **고치지 않는다**. 관측만 기록한다.

### Out of Scope — Findings 추출 (t234 / GitHub #1632)

- `synthesizeReviewOutput` 이 `Findings: []Finding{}` 를 하드코딩하는 문제는 이 SPEC 의 대상이 아니다.
- 후속 카드 **t234** 소관이며, 리드가 이 카드 착지까지 보류한다. 순서와 seam 충돌은 §C 참조.

### Out of Scope — codex 백엔드의 트리 오독 (t246)

- 워크트리 안에서 실행한 audit 이 primary 체크아웃을 리뷰한 관측(cause.md F9)은 이 SPEC 의 대상이 아니다.
- 카드 **t246** 소관이다.

### Out of Scope — 수렴 결과 영속화 (관측만)

- `.moai/state/audit-multi/` 가 저장소 전체에서 0건이라는 소급 스캔 결과(`.moai/reports/t229/retro-sweep.md`)는 이 SPEC 의 대상이 아니다. 후속 카드 개설 여부는 리드 판단이다.
- 이 0건이 뜻하는 것은 **볼 수 있는 기록이 거기까지** 라는 것이다 — 증거 도달 범위의 한계이지, 과거에 문제가 없었다는 뜻이 아니다. 소급 스캔이 t197 의 3회에서 멈춘 것도 그 이상이 없어서가 아니라 남은 기록이 거기까지이기 때문이다.
- 따라서 이 SPEC 은 `session_id` 배선을 요구사항으로 두지 않는다.

### Out of Scope — MCP 서버 프로세스 랙

- 장수 `moai mcp-server` 프로세스가 구 바이너리를 붙들고 있어 최신 수정이 반영되지 않는 문제(§A.2)는 이 SPEC 의 대상이 아니다. 운영 규율 소관이다.

### Out of Scope — GLM 백엔드

- GLM 경로(`performGLMAudit`)의 판정 합성은 손대지 않는다. 이 SPEC 은 codex 경로의 합성 seam 에 한정된다.

---

## §E 제약

- 언어: Go. 대상 패키지 `internal/cli`.
- 검증: `go test ./internal/cli/... -timeout 600s`. **`go test ./...` 금지** (저장소 규율).
- 방법론: TDD (RED → GREEN → REFACTOR).
- fail-open 불변식 유지: 어떤 경로도 hard error 를 반환하지 않는다. 백엔드 부재·오류는 `inconclusive` 로 떨어진다.
- 독립성 불변식 유지: `backendCallFn` 시그니처에 verdict 를 실어 보내지 않는다(`mcp_convergence.go:368`).

---

## §F 참조

- 원인 보고서: `.moai/reports/t229/cause.md`
- 라이브 프로브 원문: `.moai/reports/t229/live-probe-body.txt`
- 소급 스캔: `.moai/reports/t229/retro-sweep.md`
- 선행 SPEC: `SPEC-AUDIT-MULTI-MODEL-001` (수렴 엔진 원본)
- 선행 카드: t178 (`f505955a9`), t186 (`4505df411`)
- 후속 카드: t234 (= GitHub #1632, Findings 추출) · t246 (워크트리 오독)
