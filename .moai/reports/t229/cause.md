# t229 — audit_multi codex 필드-본문 불일치: 원인 확정

| 항목 | 값 |
|---|---|
| 카드 | t229 (Class B — 재현됨, 원인 미확정 상태로 착수) |
| 워크트리 | `.claude/worktrees/t229` · 브랜치 `WT-audit-verdict-converge` |
| 측정 트리 | `worktree-t229` 생성 시점 base = `origin/main` |
| 측정 일자 | 2026-08-24 |
| codex 버전 | `codex-cli 0.149.0` (`/Users/goos/.local/bin/codex`) |

---

## 결론 한 줄

**필드-본문 모순이 아니다. 백엔드는 verdict 필드를 애초에 반환하지 않는다** — 그 필드는 MoAI 어댑터가 본문에서 정규식 하나로 **합성**한 값이고, 정규식이 안 맞으면 무조건 `pass`로 떨어진다. 관대한 방향으로만 틀린 이유가 이 기본값이다.

카드가 제시한 세 후보 중:

| 후보 | 판정 |
|---|---|
| (a) 백엔드 응답 자체의 필드-본문 모순 | **반증** — 응답에 verdict 필드가 없다 (F1) |
| (b) 어댑터 추출 오류 | **확정 — 이것이 근본 원인** (F2·F3·F4·F5) |
| (c) 수렴 엔진의 읽기 범위 | **사실이나 하류** — 엔진은 잘못 합성된 필드를 충실히 전달했을 뿐 (F7) |

---

## F1 — codex는 구조화된 verdict를 반환하지 않는다

`ReviewOutput.Verdict`를 만드는 곳은 코드 전체에서 한 군데다.

```
internal/cli/mcp_codex.go:1113  func synthesizeReviewOutput(reviewText string) ReviewOutput
```

호출자는 `mcp_codex.go:705`의 `runTurn` 하나뿐이고(`codex_review_rpc_test.go:122`은 테스트), codex 응답에서 뽑아 쓰는 것은 **본문 텍스트뿐**이다(`bestCodexReviewText`, `mcp_codex.go:1065`). 함수 주석도 이를 명시한다 — *"codex does NOT return a structured verdict enum — it returns free-form prose"*.

즉 필드와 본문은 서로 다른 두 출처가 아니라, **본문 하나에서 파생된 값과 그 원본**이다. 둘이 어긋난다는 것은 파생이 틀렸다는 뜻이다.

## F2 — 합성 규칙은 정규식 1개 + 관대한 기본값

```go
// mcp_codex.go:1102
var codexFindingBullet = regexp.MustCompile(`(?m)^\s*[-*]\s+\[[A-Za-z]+\d+\]`)

// mcp_codex.go:1113
func synthesizeReviewOutput(reviewText string) ReviewOutput {
    verdict := "pass"
    if codexFindingBullet.MatchString(reviewText) {
        verdict = "fail"
    }
    ...
}
```

- 판정 근거는 `- [P1]` 꼴 심각도 태그 불릿의 **존재 여부** 단 하나.
- 안 맞으면 `pass`. **증거 부재가 통과 근거로 쓰인다** — 관측 없는 통과 주장의 코드판이다.
- 주석은 이 정규식이 **codex-cli 0.146.1의 review-mode 출력**으로 검증됐다고 적는다. 설치본은 **0.149.0**.

기존 테스트(`codex_review_rpc_test.go:114` `TestSynthesizeReviewOutput_FindingBulletsMapToFail`)가 다루는 입력은 `- [P1]` / `- [P2]` / 산문 1줄 / 빈 문자열 **4건뿐**이다. 다른 서식은 한 건도 없다.

## F3 — 그런데 audit_multi는 review-mode를 절대 쓰지 않는다

```go
// mcp_convergence.go:382  performCodexAudit
params := map[string]any{ "target": target, "model": "",
    "prompt": codexAdversarialReviewPrompt(focus) }
out, _ := runCodexReviewRPC(ctx, binaryPath, codexMethodTurnStart, params)
```

`audit_multi` 경로는 항상 **adversarial 모드**(`turn/start` + 자유형 프롬프트)다. 그리고 그 프롬프트(`mcp_codex.go:1129`)는 출력 **서식을 전혀 지정하지 않는다** — "Report concrete findings with severity, file/line, confidence, and a recommendation."

**review-mode 불릿 관례로 눈금을 맞춘 정규식을, 서식이 자유로운 adversarial 산문에 적용하고 있다.** 이것이 불일치가 이 경로에서만 반복되는 이유다.

## F4 — 라이브 재현 (2026-08-24, codex-cli 0.149.0)

`mcp__moai__codex_audit(mode=adversarial, target=uncommittedChanges)` 1회 실행. 원문: `live-probe-body.txt`.

| 관측 | 값 |
|---|---|
| 반환된 `verdict` 필드 | **`pass`** |
| 본문 1행 | **`Verdict: inconclusive — ... so a pass would be unsupported.`** |
| `codexFindingBullet` 매치 수 | **0** (`grep -cE '^[[:space:]]*[-*][[:space:]]+\[[A-Za-z]+[0-9]+\]'` → `0`, rc=1) |

codex가 **명시적으로 통과를 거부**한 응답이 `pass`로 합성됐다. t197의 3회와 방향이 같다.

## F5 — 본문에 verdict 단어가 있는데 어댑터가 버린다

위 본문은 첫 줄에 `Verdict: inconclusive`를 적었다. 어댑터에는 이 문자열을 읽는 코드가 **없다**. 백엔드가 스스로 낸 판정을 무시하고 불릿 유무로 다시 추측하는 구조다.

본문은 `Findings: []` / `merge_status:` 를 담은 yaml 블록도 함께 냈다 — codex는 구조를 낼 의사가 있었고, 어댑터가 그것을 읽지 않는다.

## F6 — Findings는 항상 비어 있다 (교차 축, #1632)

```go
// mcp_codex.go:1121
Findings:  []Finding{},
```

하드코딩된 빈 슬라이스다. 따라서 수렴 엔진은 "차단 N건"을 셀 수단조차 없다. **이 카드의 수정 대상이 아니라 #1632의 축**이며, 두 카드가 같은 seam을 만지므로 착수 순서 조율이 필요하다.

## F7 — 수렴 엔진은 필드만 읽는다 (하류, 그러나 검출 지점)

`converge()`(`mcp_convergence.go:135`)가 보는 것은 `PerBackendVerdict.Verdict` 뿐이다. `Summary`(= 본문 원문)는 결과에 실려 나가지만 **판정에 한 번도 쓰이지 않는다**.

엔진은 자기 입력에 대해 정확하게 동작했다 — `codex.Verdict == "pass"` 였으니 `overall_verdict: pass`, 갈림 없으니 `disagreement_flag: false`. 버그는 여기가 아니다. 다만 **필드-본문 불일치를 잡아낼 장치가 파이프라인 어디에도 없다**는 사실은 여기서 확정된다.

## F8 — 소급 스캔: 도달 범위가 좁다 (부재 주장 아님)

스윕 스크립트: `retro-sweep.sh`, 결과: `retro-sweep.md`. 범위는 primary 체크아웃의 `.moai/reports/` + 전 워크트리의 `.moai/reports/` + 전 트리의 `.moai/state/audit-multi/*.json`.

| 관측 | 값 |
|---|---|
| 영속된 `ConvergenceResult` 상태파일 | **0건** |
| 백엔드별 판정표를 담은 리포트 (내용해시 중복 제거) | **9건, 전부 t197** |
| 그중 `pass` 필드 + FAIL 본문 | **2건** (`verdict-iter3.md`, `verdict-init-2.md`) |
| 산문으로만 증언된 추가 1건 | iter2 (`verdict-iter3.md:33`) — 별도 판정서 없음 |

상태파일 0건의 뜻은 **이 저장소 이력에서 `audit_multi`가 `session_id`를 넘긴 적이 한 번도 없다**는 것이다(`mcp_convergence.go:518` — 빈 세션 id면 영속화가 no-op). 따라서 t197 레인이 손으로 표를 적어 두지 않았다면 이 불일치는 **아무 흔적도 남기지 않았다**. 소급 스캔이 "3회"에서 멈춘 것은 그 이상이 없어서가 아니라 **볼 수 있는 기록이 거기까지**이기 때문이다.

## F9 — 범위 밖 관측 (보고만): codex 백엔드가 워크트리를 무시한다

라이브 프로브는 `.claude/worktrees/t229`에서 실행했고 그 트리에 `probe_vuln.go`를 두었는데, codex는 **primary 체크아웃**의 미커밋 변경(`internal/statusline/renderer.go`)을 리뷰했고 프로브 파일을 보지 못했다. 워크트리 안에서 audit을 돌리면 **다른 트리를 감사한다**. 별도 카드감.

---

## 수정 방향과 대표 mutant

카드가 지정한 대표 mutant: **"불일치를 감지는 하되 관대한 쪽을 채택하는 구현"**. AC가 "불일치를 로그에 남긴다"만 요구하면 이 mutant가 통과한다. 따라서 AC는 **채택된 결과값**을 걸어야 한다.

방향 4가지(설계는 plan 단계 소관):

1. **본문이 명시한 verdict를 읽는다** — `Verdict: pass|fail|inconclusive`, `FAIL 0.75` 류 점수 표기 포함. 본문 명시가 불릿 추론보다 우선.
2. **본문 판정과 불릿 추론이 갈리면 보수적인 쪽을 채택한다** (`fail` > `inconclusive` > `pass`). 감지만 하고 관대한 쪽을 고르면 mutant.
3. **adversarial 경로에서 verdict 단어도 불릿도 없으면 `inconclusive`** — `pass` 기본값 제거. 단, **native review-mode의 무불릿 = 정상 통과** 경로는 건드리면 안 된다(codex review 게이트의 clean-pass가 여기 걸림). 모드별로 갈라야 한다.
4. **수렴 엔진이 합성 근거를 본다** — 필드-본문 불일치가 `disagreement_flag` + `residual_risk_note`로 드러나야 한다. 조용히 지나가면 안 된다.

부수 의무 1건: `session_id`를 실제로 넘겨 `ConvergenceResult`를 영속화해야 F8의 소급 불가 상태가 반복되지 않는다.
