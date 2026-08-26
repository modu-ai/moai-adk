# t229 M4 — 회귀 고정 마감 기록

> 카드 t229 · SPEC-CODEX-VERDICT-SYNTH-001 · 브랜치 `WT-audit-verdict-converge`
> 측정 트리: `1e1edf6b4` (클린) · 측정 일자 2026-08-25

## 1. 이 세션이 실제로 한 일

M4 의 테스트 3건은 직전 세션 마지막 커밋 `1e1edf6b4` 에 **이미 함께 착지해 있었다**(파일 `internal/cli/codex_verdict_regression_test.go`, 커밋 제목은 docs 였다). 따라서 이 세션의 M4 작업은 코드 추가가 아니라 **검증과 기록**이다 — 착지한 것이 실제로 도는지, DoD 가 요구한 검사 중 남은 것이 무엇인지.

남아 있던 것은 둘이었다:

1. **전 패키지 스위트** — `a84b25917` 이후 한 번도 완주하지 않았다(`m1-boundary.md` §남은 일 1번).
2. **DoD 211 의 뮤테이션 독립성 검사** — "mutant (e) 가 AC-CVS-006 에서만 실패하고 AC-CVS-001~005 는 전부 통과함을 확인". 감사 iter2 가 검증 불가 항목을 이것으로 교체했으므로, 확인하지 않으면 DoD 가 닫히지 않는다.

둘 다 이 세션에서 실행했다.

## 2. 전 패키지 스위트 — 초록

```
$ go test ./internal/cli/ -count=1 -timeout 1200s
ok  	github.com/modu-ai/moai-adk/internal/cli	510.756s
```

판정은 **출력 본문의 `ok` 행**으로 했다. 래퍼 종료코드는 근거로 쓰지 않았다(`m1-boundary.md` 가 지목한 실패 형태 — 백그라운드 래퍼의 exit 0 이 패키지 FAIL 을 가린 사고). 실측 510.756초는 `CLAUDE.local.md` §6 의 `internal/cli` 타임아웃 하한 600s 근거와 같은 크기대다.

증거 파일: `.moai/state/verify/t229-m4/pkg-cli.log`

## 3. 정적 검사 — 양 플랫폼 초록

| 명령 | 결과 |
|---|---|
| `go vet ./internal/cli/...` | rc=0, 출력 없음 |
| `GOOS=windows go vet ./internal/cli/...` | rc=0, 출력 없음 |

`m1-boundary.md` 가 미측정으로 남겼던 `GOOS=windows go vet` 이 여기서 닫힌다. 다만 [HARD] **vet 초록은 컴파일만 증명한다** — Windows 동작의 근거가 아니며, 그 판정은 CI 매트릭스 몫이다.

증거 파일: `.moai/state/verify/t229-m4/vet-darwin.log` · `vet-windows.log`

## 4. DoD 211 — mutant (e) 독립성, 실측으로 확인

### 심은 것

`adoptConservativeVerdict` 의 집합 최댓값 채택을 **대입 열 의미론**으로 바꿨다. 이것이 acceptance.md §B-2 가 mutant (e) 로 지목한 형태 — "M2 를 가장 자연스러운 방식으로 쓴 구현" — 이다.

```go
	adopted := ""
	for _, s := range signals {
		adopted = s.verdict          // ← 최댓값 비교를 제거: 나중 신호가 앞선 것을 덮는다
	}
	return adopted
```

### 관측

```
--- PASS: TestSynthesizeReviewOutput_AdversarialVerdictLine
--- PASS: TestSynthesizeReviewOutput_VerdictLineDirections
--- PASS: TestSynthesizeReviewOutput_FindingBulletsMapToFail
--- PASS: TestSynthesizeReviewOutput_RecordsSignalDivergence
--- PASS: TestSynthesizeReviewOutput_NoNoteWhenSignalsAgree
--- PASS: TestConverge_SurfacesSignalDivergence_WithoutBlocking
--- PASS: TestRunMultiAudit_ForwardsSynthesisNoteToPerBackendVerdict
--- PASS: TestSynthesizeReviewOutput_AdversarialNeverPassesUnknownFormat
--- PASS: TestSynthesizeReviewOutput_UnknownMethodIsConservative
--- PASS: TestSynthesizeReviewOutput_NativeCleanReviewStaysPass
--- PASS: TestSynthesizeReviewOutput_ModeSplitsTheSameBody
--- PASS: TestSynthesizeReviewOutput_LiveProbeBodyStaysInconclusive
--- PASS: TestCodexTask_OutputTextUnchangedByVerdictSynthesis
--- PASS: TestSynthesizeReviewOutput_ScoredVerdictIsRead
--- PASS: TestSynthesizeReviewOutput_ScoredVerdictDoesNotMatchProse
--- FAIL: TestSynthesizeReviewOutput_AdoptsMostConservativeSignal
--- PASS: TestSynthesizeReviewOutput_SignalOrderDoesNotMatter
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.179s
```

갈라진 행:

```
K1 stated fail then scored pass:      adopted "pass", want "fail"
K2 scored pass then stated fail:      adopted "pass", want "fail"
K4 stated inconclusive then scored pass: adopted "pass", want "inconclusive"
```

### 이 관측이 말하는 것

- **AC-CVS-006 만 죽는다.** AC-CVS-001~005 를 담당하는 테스트 15건이 전부 초록인 채 mutant 가 통과한다. 즉 AC-006 이 없었다면 이 구현이 **초록 신호를 달고 착지**했을 것이고, 이 SPEC 이 없애려던 §0 위반을 이 SPEC 이 새로 만들어 냈을 것이다. D1 이 경고한 그대로다.
- **`SignalOrderDoesNotMatter` 는 mutant (e) 를 잡지 못한다.** K1·K2 는 텍스트 순서만 다르고 신호 **수집** 순서는 둘 다 `stated → scored` 로 같으므로, 대입 열 의미론 아래에서도 두 행은 **같은 값**(`pass`)을 낸다. 이 테스트가 겨냥한 것은 mutant (f)(텍스트 등장 순서 의존)이며 (e) 가 아니다 — 두 증인이 서로 다른 변종을 담당한다는 사실이 여기서 관측됐다.
- 세 행이 갈린 이유는 모두 하나다: 관대한 신호(`scored: pass`)가 **수집 순서상 뒤에 오기 때문에** 앞선 보수적 신호를 덮는다.

### 원복

```
$ git checkout -- internal/cli/mcp_codex.go
$ shasum -a 256 -c .moai/state/verify/t229-m4/mcp_codex.sha256
internal/cli/mcp_codex.go: OK
```

심기 전 해시 `0750534e…` 와 원복 후 해시가 일치한다. 뮤테이션은 트리에 남지 않았다.

증거 파일: `.moai/state/verify/t229-m4/mutant-e.log`

## 5. AC 별 충족 상태

| AC | 증인 | 상태 |
|---|---|---|
| AC-CVS-001 미인식 서식은 통과 아님 | `AdversarialNeverPassesUnknownFormat` · `UnknownMethodIsConservative` (C1~C8 corpus 순회) | PASS |
| AC-CVS-002 점수 표기 인식 | `ScoredVerdictIsRead` · `ScoredVerdictDoesNotMatchProse` | PASS |
| AC-CVS-003 native 무불릿 보존 | `NativeCleanReviewStaysPass` · `ModeSplitsTheSameBody` | PASS |
| AC-CVS-004 불일치 채택 + 기록 + 전달 | `RecordsSignalDivergence` · `NoNoteWhenSignalsAgree` · `Converge_SurfacesSignalDivergence_WithoutBlocking` · `RunMultiAudit_ForwardsSynthesisNoteToPerBackendVerdict` | PASS |
| AC-CVS-005 회귀 방어 (M4) | `FindingBulletsMapToFail`(native 명시) · `LiveProbeBodyStaysInconclusive` · `CodexTask_OutputTextUnchangedByVerdictSynthesis` | PASS |
| AC-CVS-006 P-CONS 집합 최댓값 | `AdoptsMostConservativeSignal`(K1~K8 단일 단언) · `SignalOrderDoesNotMatter` | PASS |

## 6. 관측하지 않은 것 (Gaps)

- **커버리지 수치** — `-cover` 를 돌리지 않았다. 이 세션은 스위트 1회로 510초를 썼고, 커버리지 재실행은 그만큼을 다시 지출한다. 필요하면 sync-audit 이 요구할 때 측정한다.
- **`golangci-lint`** — 미실행. 판정은 CI 몫으로 남긴다.
- **라이브 codex 프로브** — 하지 않았다. SPEC §A.2 가 금지한다(바이너리 랙으로 오판한다).
- **mutant (g) 쌍 특수화** — 오늘 어떤 테스트로도 검출되지 않는다는 것이 acceptance.md §B-2 각주의 확정 사실이므로, 심어 보지 않았다. 심어도 전 행 통과가 예상되며 그 예상 자체가 문서화된 결론이다.
- **CI 매트릭스(darwin×2 / windows)** — 로컬 판정은 조기 신호일 뿐이며, 전 패키지 판정은 PR 의 CI 가 낸다.

## 7. 잔여 위험 (Residual risk)

- **넷째 신호**. AC-CVS-006 이 오늘 잡는 것은 순서 의존(e·f) 뿐이고, 쌍 특수화(g)는 신호가 셋이고 그중 하나가 `fail` 전용인 동안 일반 규칙과 구별되지 않는다. 넷째 신호를 들이는 사람이 그 등가성을 깨는 조합 행을 함께 넣지 않으면, 오늘의 corpus 는 그를 지켜주지 않는다.
- **t234(#1632)와의 seam 충돌**. 같은 함수 `synthesizeReviewOutput` 을 Findings 축에서 다시 고친다. 시그니처 `(reviewText, method string)` 를 되돌리면 모드 구분이 사라진다 — 코드 주석과 SPEC §D 양쪽에 남겨 두었다.
- **스위트 소요 510초**. 로컬 게이트로 상시 돌리기에는 비싸다. 카드 단위 검증은 대상 테스트로, 전 패키지 판정은 CI 로 가는 규율이 이 패키지에서 특히 구속력을 갖는다.
