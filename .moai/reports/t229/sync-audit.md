# sync-audit — SPEC-CODEX-VERDICT-SYNTH-001 (card t229)

- 감사자: sync-auditor (독립 재검증 세션) · 2026-08-26
- 측정 트리: `.claude/worktrees/t229` @ `eb2d56a9d` (branch `WT-audit-verdict-converge`, 클린)
- SPEC 상태: `status: completed` · sync commit `62ea002dc` · backfill `eb2d56a9d`
- 방침: §E 자가보고는 의심 대상으로 취급하여 이 트리에서 재실행·로그 판독으로 재확인 (VCI §2 — 모든 귀속은 이번 실행·이 트리 기준)

## 1. Claim — 이 감사가 확인한 주장

| # | 주장 | 판정 |
|---|---|---|
| C-1 | AC-CVS-001..006 6건 전부 이 트리에서 초록 | 확인 |
| C-2 | P-CONS(`adoptConservativeVerdict`)가 집합 최댓값이며 순서 무관 | 확인 (코드 정독 + 정합 실행) |
| C-3 | 미인식 fall-through 가 모드별로 갈림 (review/start → `pass`, turn/start → `inconclusive`) | 확인 |
| C-4 | 점수 표기 정규식이 t197 서식(`FAIL 0.75 / 1.00`)을 읽고 산문을 걸러냄 | 확인 |
| C-5 | K3·K7 기대값이 P-CONS 유도값 그대로 (관측 동작으로 낮춤 없음) | 확인 |
| C-6 | mutant (e) 독립성 — AC-CVS-006 만 죽고 나머지 16건 통과 | 증거 판독로 확인 (`mutant-e.log` 원문) |
| C-7 | 3-phase close 무결성 — sync commit 이 spec.md 머리말 2줄 외에 본문 미수정 | 확인 |
| C-8 | backfill `eb2d56a9d` 가 `sync_commit_sha` 를 실제 SHA로 정확히 채움 | 확인 |
| C-9 | 카드 커밋 범위 초과 파일 없음 | 확인 |
| C-10 | CHANGELOG 항목이 코드와 정확히 대응 (필드명·AC 6건) | 확인 (계수 귀속 1건 부정확 — F1) |

## 2. Evidence — 이번 실행에서 직접 관측한 출력

### 2.1 대상 테스트 재실행 (Functionality)

명령: `go test ./internal/cli/ -run 'TestSynthesizeReviewOutput|TestConverge_|TestRunMultiAudit_|TestCodexTask_OutputText' -count=1 -v -timeout 600s`

```
$ grep -c '=== RUN' → 34            # 셀렉터 0매칭 아님 (34케이스 — successions.md 관측과 동일)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.443s
(--- FAIL 행 0건)
```

AC별 증인 매핑 (RUN 이름에서 직접 확인): AC-001 `AdversarialNeverPassesUnknownFormat`·`UnknownMethodIsConservative` / AC-002 `ScoredVerdictIsRead`·`ScoredVerdictDoesNotMatchProse` / AC-003 `NativeCleanReviewStaysPass`·`ModeSplitsTheSameBody` / AC-004 `RecordsSignalDivergence`·`Converge_SurfacesSignalDivergence_WithoutBlocking`·`RunMultiAudit_ForwardsSynthesisNote` / AC-005 `FindingBulletsMapToFail`·`LiveProbeBodyStaysInconclusive`·`CodexTask_OutputTextUnchanged` / AC-006 `AdoptsMostConservativeSignal`·`SignalOrderDoesNotMatter` — 6 AC 전부 증인 존재.

### 2.2 해시 핀 + 정적 검사 + 린트 (Craft)

```
$ shasum -a 256 -c .moai/state/verify/t229-m4/mcp_codex.sha256
internal/cli/mcp_codex.go: OK        # rc=0 — M4 검증 시점과 바이트 동일

$ go vet ./internal/cli/...
(출력 없음) rc=0

$ golangci-lint run ./internal/cli/... --timeout=5m
0 issues.                            # rc=0 — m4-close.md §6 가 미실행으로 남긴 갭을 이 감사에서 폐쇄

$ grep -c 'SPEC-CODEX-VERDICT-SYNTH-001' CHANGELOG.md
1                                    # 중복 없음 (B12)
```

### 2.3 변경 함수 커버리지 (Craft — 이 감사 신규 측정)

명령: `go test ./internal/cli/ -run 'TestSynthesizeReviewOutput|TestConverge_|TestRunMultiAudit_|TestCodexTask_OutputText' -count=1 -coverprofile=…` + `go tool cover -func`

```
codexVerdictSignalsOf      100.0%      codexVerdictRank          100.0%
adoptConservativeVerdict   100.0%      codexUnrecognizedVerdict  100.0%
synthesizeReviewOutput     100.0%      describeSignalDivergence  100.0%
collectSynthesisNotes      100.0%      converge                   93.8%
```

패키지 전체 수치(8.4%)는 대상 부분집합 실행이므로 무의미 — 이 카드가 변경한 함수만 봐도 전부 포화. m4-close.md §6 이 커버리지를 "sync-audit 이 요구하면 측정"으로 남겨둔 갭을 폐쇄함.

### 2.4 P-CONS 코드 정독 (Functionality)

- `adoptConservativeVerdict` (mcp_codex.go:1297): `codexVerdictRank` 상대 **엄격 최댓값 루프**. 최댓값은 교환법칙이 성립하므로 신호 수집 순서·텍스트 등장 순서 어느 쪽에도 의존하지 않음. 순위 불명 문자열은 rank 0으로 채택 불가 — 넷째 신호가 와도 규칙 문장이 아니라 데이터만 늘어남.
- `codexUnrecognizedVerdict` (mcp_codex.go:1320): `codexMethodReviewStart` → `pass`, 그 외 전부(turn/start 포함) → `inconclusive`. plan.md §C.2 표와 일치.
- `codexScoredVerdict` (mcp_codex.go:1239): `(?m)^[\s>#*_]*(PASS|FAIL|INCONCLUSIVE)\b[ \t]+[01]\.\d+\b` — 줄 머리 + 대문자 판정어 + 0~1 소수. `FAIL 0.75 / 1.00` 매치 확인, 산문 `the suite reported PASS 12 times` 는 줄 머리 아님·`12` 가 `[01]\.\d+` 불일치로 이중 거부.
- 신호 분리: `codexVerdictSignalsOf` 는 집합 수집만, 채택은 별도 함수 — 넷째 신호 추가 시 채택 규칙 불변이라는 구조적 주장이 코드 형태와 일치.
- `Findings: []Finding{}` 하드코딩 유지 (mcp_codex.go:1355) — t234 축 미침범 확인.
- `synthesizeReviewOutput` 프로덕션 호출자는 `runTurn`(mcp_codex.go:775) 유일 — 전 저장소 grep 실측.

### 2.5 K3·K7 기대값 (spec 위반 없음)

`codex_verdict_scored_test.go:55` `{"K3 stated pass then scored fail", …, "fail"}` · `:59` `{"K7 scored inconclusive then stated pass", …, VerdictInconclusive}` — acceptance.md §B-2 표와 동일한 P-CONS 유도값이며, 파일 헤더에 [HARD] "기대값은 P-CONS 에서 도출, 관측 동작에 맞춰 낮추지 말 것" 주석이 살아 있음. AC-CVS-006 은 corpus 순회 **단일 단언**(루프 1개, `want` 비교 1개) — 행 추가가 단언문 수정을 요구하지 않는 속성형 확인.

### 2.6 증거 로그 판독 (§E 주장 재확인)

- `pkg-cli.log`: `ok github.com/modu-ai/moai-adk/internal/cli 510.756s` — §E.2 인용과 동문.
- `mutant-e.log` (원문 판독): `--- FAIL: TestSynthesizeReviewOutput_AdoptsMostConservativeSignal` 단 1건, 갈라진 행 K1·K2·K4 (`adopted "pass", want "fail"/"inconclusive"`), 나머지 16건 `--- PASS`. DoD 211 독립성 주장과 정확히 일치 — "실패가 관측된 검사"임.
- `baseline-rebased.md`: C1~C8 **8/8 RED** 표 (rebase 후 재측정) — AC-CVS-001 RED 기대 "전부" 충족 기록.
- `m1-boundary.md`: M1 경계 K3·K7 RED → M2 후 초록 전환 관측 + 메커니즘 분석 (두 본문 모두 `stated` 매치로 fall-through 미타격, 붉은 원인은 점수 신호 부재). 문서화된 설계대로 — 결함 아님.
- `succession.md`: 병합 `4561f432c` (충돌 0) + 병합 트리 대상 테스트 34케이스 초록 + 참여자 수 축 별도 카드 연기(운영자 승인 2026-08-26) 기록.

### 2.7 3-phase close · backfill · 범위 (Consistency)

```
62ea002dc — spec.md (status: in-progress→completed, updated: 2줄) + progress.md + CHANGELOG.md(+1)  # 본문 수정 0
eb2d56a9d — progress.md 1줄: sync_commit_sha: "pending-backfill" → "62ea002dc"                     # 자기참조 물리학 해소, 정확
55b2ca3e1 / d68b6ea7c / a84b25917 — internal/cli 5·2·3파일 (코드+테스트만, 범위 외 0)
```

`git diff origin/main..HEAD --stat`: internal 9파일(전부 이 카드 소관 — 신규 5 + 기존 3 최소 적응 + convergence) + SPEC 아티팩트 + 보고서 + CHANGELOG 1줄. 기존 테스트 적응 2건은 시그니처 마이그레이션(`codexMethodReviewStart` 명시)이며, `mcp_convergence_test.go` 픽스처 교체("codex:ok, no findings" → "Verdict: pass — no findings.")는 **구 픽스처가 우연히 결함 본체(미인식 본문→pass)에 기대고 있었던 것**을 주석으로 문서화한 정당한 재귀거임 — 회귀 은폐와 구별됨 (기대값 자체는 `pass` 로 불변).

### 2.8 CHANGELOG 정확성 (B12)

필드명 `SynthesisNote`·`describeSignalDivergence`·`codexScoredVerdict`·`adoptConservativeVerdict` 전부 실재 확인. AC 6건 = acceptance.md 실측과 일치. 인용 커밋(`55b2ca3e1`~`1e1edf6b4`, run close `1a6c0fac0`, merge `4561f432c`) 전부 실재. "34 cases re-verified" — 이 감사의 재실행(34 RUN)과 일치. 예외 1건: F1 참조.

## 3. Baseline-attribution

전항의 명령·출력은 2026-08-26 이 감사 세션이 트리 `eb2d56a9d`(워크트리 `.claude/worktrees/t229`)에서 직접 실행한 것이다. §E.2/§E.3 의 원측정 귀속(M4 `1e1edf6b4`, 병합 트리 `4561f432c`)은 로그 원문 판독으로 재확인했고, 해시 핀은 이 트리에서 재검증(`OK`)했다 — `mcp_codex.go` 는 M4 검증 시점과 바이트 동일.

## 4. Findings

- **F1** [low] [optional] `CHANGELOG.md`·`progress.md` §E.2 M4 행 — "3 fixed regressions"(`회귀 고정 3건`)를 `codex_verdict_regression_test.go` 에 귀속시키나 해당 파일은 3건 중 **2건**(LiveProbe·CodexTask)만 담고, 세 번째(`FindingBulletsMapToFail` native 명시 확장)는 `codex_review_rpc_test.go` 에 있다. 커버리지 자체는 실재(3건 전부 이 감사에서 판독·실행)하며 수 허위는 아니지만 파일별 계수 귀속이 하나 어긋남. 수정안: "3 fixed regressions across codex_verdict_regression_test.go (2) + the extended codex_review_rpc_test.go (1)" 로 표현.
- **F2** [low] [optional] `codex_verdict_regression_test.go:12` — 회귀 픽스처를 `../../.moai/reports/t229/live-probe-body.txt` (패키지 밖 보고서 트리)에서 읽는다. 이 브랜치에서 파일이 추적되므로(`git ls-files` 확인) CI 재현성에는 문제 없으나, `.moai/reports/` 를 휘발성으로 취급하는 저장소 관례와의 결합이 후속 정리에서 테스트를 부러뜨릴 수 있다. 수정안(후속): `internal/cli/testdata/live-probe-body.txt` 로 복제 이동. "사고 원문을 옮겨 적지 않는다"는 의도는 유지 가능.
- **F3** [info] [optional] 잔여 관측 — 신호 인식기가 유형별 첫 매치만 수집하므로(`FindStringSubmatch`), 하나의 본문에 같은 유형의 상반된 신호가 둘 있으면(예: `Verdict: fail` 과 이후 `Verdict: pass`) 앞선 것만 집합에 들어온다. 신호 모델이 "인식기당 1독해"인 현 계약에서 P-CONS 는 그대로 성립하며, 미인식 형태는 §0 에 따라 `inconclusive` 로 안전하게 떨어지므로 결함이 아니다 — 넷째 신호 논의와 같은 맥락의 조건부 잔여 위험으로만 기록.

## 5. Gaps — 이 감사도 관측하지 않은 것

- 전 패키지 스위트 판정·CI 매트릭스(darwin×2/windows) — PR CI 몫 (§E.4 가 명시한 정직한 갭; 로컬 재실행은 부하 규율 위반).
- 라이브 codex 프로브 — SPEC §A.2 가 금지(바이너리 랙).
- mutant (g) 쌍 특수화 실측 — acceptance.md §B-2 각주가 "오늘 검출 불가"로 확정한 문서화된 등가성이므로 심지 않음 (설계 문서화 완료).
- 패키지 전체 커버리지 수치 — 대상 테스트 기준 변경 함수 포화(§2.3)로 이 카드 범위는 폐쇄했으나 전체 수치는 미측정.

## 6. Residual-risk

- **넷째 신호**: AC-CVS-006 이 오늘 잡는 것은 순서 의존 변종까지이며, 쌍 특수화는 신호가 셋이고 하나가 `fail` 전용인 동안 구별 불가 — 넷째 신호 도입자는 쌍 특수화를 가르는 조합 행을 함께 넣어야 한다(spec.md §A.5 [HARD]·m4-close.md §7 에 문서화됨).
- **t234 seam**: 같은 함수를 Findings 축에서 다시 고친다. `(reviewText, method)` 시그니처 되돌리기 금지가 코드 주석(mcp_codex.go:1343-1345)과 spec.md §D 양쪽에 새겨져 있으나, 최종 방어는 리뷰다.
- F3 의 동형 신호 다중 발생 사례.

## 7. Dimension Scores

| Dimension | Score | 근거 요약 |
|---|---|---|
| Functionality | 0.96 | AC 6/6 이 트리 재실행 초록(34 RUN·ok·FAIL 0) + P-CONS/모드/정규식 정독 일치 + 증거 로그 원문 판독 일치. F1 계수 귀속 실수만 감점. |
| Security | 0.95 | 변경 방향이 판정 보수화(fail-safe) — 게이트 관대 편향 제거. 신규 공격면 없음(정적 정규식, 외부 입력 검증 추가 없음, 비밀 미관련), fail-open 불변식 유지. |
| Craft | 0.92 | 변경 함수 커버리지 100%(converge 93.8%), golangci-lint 0 issues(감사 폐쇄), vet 클린, mutant 독립성 관측 기록, RED 8/8·K3/K7 부분 RED 기록 완전. F2 픽스처 위치 감점. |
| Consistency | 0.95 | 3-phase close·소유권 매트릭스·close-subject 규약 정확히 준수, 커밋 메시지(Conventional + SPEC + 카드 id + 🗿 MoAI) 준수, 범위 초과 0, 기존 코드 스타일 정합. |

**Harmonic mean** = 4 / (1/0.96 + 1/0.95 + 1/0.92 + 1/0.95) ≈ **0.945**

## 8. Binding Verdict

**PASS** — 6개 AC 전부 이 트리 재실행과 코드 정독으로 확인됐고 3-phase close·범위·CHANGELOG 가 무결하며, 잔여 항목은 전부 optional(F1~F3)이라 차단 결함이 없다.

(산출 기준: harmonic mean 0.945, must-pass 축인 Functionality·Security 각각 0.96/0.95 통과.)
