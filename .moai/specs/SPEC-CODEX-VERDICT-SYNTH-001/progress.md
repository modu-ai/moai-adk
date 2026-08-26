# SPEC-CODEX-VERDICT-SYNTH-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- 작성자: manager-spec
- Tier: **S** · 방법론: TDD
- 산출물: `spec.md` (REQ 4건 = 결함 3 + 회귀 방어 1, 지배 원칙 §0, 명명 속성 P-CONS) · `plan.md` (M1~M4) · `acceptance.md` (**AC 6건** · 속성형 2건 — AC-CVS-001 은 서식 corpus 8건(§B), AC-CVS-006 은 조합 corpus 8행(§B-2, M2 가드))
- 착수 트리: `.claude/worktrees/t229`, base `origin/main` @ `294b4b6ab`
- **일차 근거**: `.moai/reports/t229/premise-revision.md` §2 의 7행 실측표. `cause.md` 는 부분 스테일이며 어긋나면 premise-revision 이 이긴다
- v0.3.0 개정 (리드 판정): Tier M→S · §A.2 를 프로세스 랙 → **바이너리 랙**(설치본 `a1b1ca696`, `origin/main` 대비 259 커밋 뒤짐)으로 교체 · 보수 채택 순서를 결함에서 유지보수 메모로 강등 · AC 를 서식 열거형 → **속성형** 전환 · 후속 카드 t248 명시
- v0.4.0 개정 (plan-audit iter1 = PASS-WITH-DEBT 0.7625 수리):
  - **D1 (critical/blocking)** — 보수 채택 강등의 전제가 **2신호 구현에만** 성립함을 §A.5 에 명시. M2 가 세 번째 신호를 들이면 `Verdict: fail` + `PASS 0.95` → `pass` 세탁이 가능하므로 **AC-CVS-006** 을 신설해 채택값으로 고정. 요구사항은 강등 상태 유지(구현 형태 자유), plan.md §D M2 에 착수 게이트 추가
  - **D1 보강 (리드 추가 지시)** — 보수 규칙을 **순서 무관 집합 연산**으로 재서술. spec.md §A.5 에 명명된 속성 **P-CONS**(채택값 = 신호 **집합**의 최댓값, `fail` > `inconclusive` > `pass`)를 SPEC 본문 규칙으로 신설하고, "나중 신호가 앞선 것을 덮지 않는다" 류 순서 서술을 명시적으로 금지(신호 셋에만 유효 → 넷째에서 재발). AC-CVS-006 을 쌍 열거형 → **조합 corpus(§B-2, K1~K8) 순회 단일 단언**으로 재작성. 리드가 준 세 쌍은 K1·K3·K4 로 편입되어 증인이 됨. mutant (f) 순서 의존 · (g) **쌍 특수화** 추가, K1/K2(같은 집합·다른 순서) · K5(`scored × bullet`) · K6(3-신호) · K8(갈리지 않는 행)이 각각 증인
  - **D2 (major/blocking)** — §A.6 의 "게이트 차단" 논거가 측정상 거짓임을 확인(`isBlockVerdict` 는 `fail` 접두사만 차단, `codex_review_gate.go:116-117` / 종단 `:109`). **보고 정확성** 논거로 교체. AC-CVS-003 은 유지, mutant (b) 서술만 정정
  - **D3 (blocking)** — AC-CVS-001 의 RED 기대를 C1·C5·C7 → **C1~C8 전부**로 정정
  - **D9 (blocking)** — 스테일 인용 정정: `mcp_codex.go:1152`→`1155`, `mcp_convergence.go:368`→`367`. 추가로 감사 보고서에서 인용한 2건도 자체 실측으로 재정정: `codex_review_rpc_test.go:120`→`119`, `mcp_convergence.go:125-128`→`126-129`
  - 부수: **D4** 개방 항목 종결(`review-output.schema.json` 은 저장소에 파일로 존재하지 않음 — 실측), **D7** 모드 배선 증인을 도달 불가한 C7 → 도달 가능한 C5 로 교체
  - 미착수(리드 지시 범위 밖): D5(`Where`→`When` 패턴 정밀도) · D6(복합 REQ 분할) · D8(PR 경로 명시)
- 착수 순서 확정: 이 SPEC → t234 (= GitHub #1632)
- plan-audit iter2 대기

## §E.2 Run-phase Evidence

- 브랜치 `WT-audit-verdict-converge` · 마감 트리 `1e1edf6b4` (클린) · base `origin/main` @ `294b4b6ab`
- 방법론 TDD · 마일스톤 M1~M4 전부 착지

| 마일스톤 | 커밋 | 내용 |
|---|---|---|
| M1 | `55b2ca3e1` | 모드 seam(`synthesizeReviewOutput(reviewText, method)`) + 신호 수집·집합 최댓값 채택 구조 |
| M2 | `d68b6ea7c` | `codexScoredVerdict` 점수 표기 인식기 |
| M3 | `a84b25917` | `SynthesisNote` 불일치 기록 + `converge()` 반영 |
| M4 | `1e1edf6b4` | 회귀 고정 3건 (`codex_verdict_regression_test.go`) |

### 품질 게이트 (plan.md §E 전량)

```
$ go test ./internal/cli/ -count=1 -timeout 1200s
ok  	github.com/modu-ai/moai-adk/internal/cli	510.756s

$ go vet ./internal/cli/...
(출력 없음, rc=0)

$ GOOS=windows go vet ./internal/cli/...
(출력 없음, rc=0)
```

판정은 출력 본문의 `ok` 행으로 했다 — 백그라운드 래퍼 종료코드는 근거로 쓰지 않았다. 증거: `.moai/state/verify/t229-m4/{pkg-cli,vet-darwin,vet-windows}.log`

### DoD 211 — mutant (e) 독립성 실측

`adoptConservativeVerdict` 를 대입 열 의미론으로 바꿔 심은 뒤 t229 관련 테스트 17건 실행:

```
--- FAIL: TestSynthesizeReviewOutput_AdoptsMostConservativeSignal   (AC-CVS-006)
  K1 adopted "pass", want "fail" / K2 adopted "pass", want "fail"
  K4 adopted "pass", want "inconclusive"
나머지 16건 전부 PASS (AC-CVS-001~005 담당 테스트 포함)
```

AC-CVS-006 만 죽는다 — 이 AC 가 없었다면 mutant (e) 가 초록으로 착지했을 것이라는 D1 의 경고가 관측으로 섰다. 원복 후 해시 `0750534e…` 일치 확인. 증거: `.moai/state/verify/t229-m4/mutant-e.log` · 상세 `.moai/reports/t229/m4-close.md`

### 부수 관측

`SignalOrderDoesNotMatter` 는 mutant (e) 를 잡지 못한다 — K1·K2 는 신호 **수집** 순서가 같아 대입 열 아래에서도 같은 값을 낸다. 이 증인이 겨냥한 것은 mutant (f)(텍스트 등장 순서 의존)이며, 두 증인이 서로 다른 변종을 담당한다.

### 미관측 (Gaps)

커버리지 수치 · `golangci-lint` · 라이브 codex 프로브(SPEC §A.2 가 금지) · mutant (g) 쌍 특수화(오늘 검출 불가가 확정 사실) · CI 매트릭스 판정. 전 패키지 판정은 PR 의 CI 가 낸다.

## §E.3 Run-phase Audit-Ready Signal

- AC-CVS-001 ~ AC-CVS-006 **6건 전부 PASS** (증인 매핑: `.moai/reports/t229/m4-close.md` §5)
- RED 기록: C1~C8 8/8 (`baseline-rebased.md`) · K3·K7 부분 RED 와 M2 가 뒤집은 관측 (`m1-boundary.md`)
- K3·K7 기대값을 관측 동작에 맞춰 낮추지 않았다 — corpus 의 기대값은 P-CONS 에서 도출된 그대로다
- `TestSynthesizeReviewOutput_FindingBulletsMapToFail` 삭제되지 않았고 native 모드 명시 호출로 확장됐다 (`codex_review_rpc_test.go:122`)
- AC-CVS-001·006 단언문 모두 corpus 순회 단일 단언 — 행 추가가 단언문 수정을 요구하지 않는다
- 잔여 위험 3건(넷째 신호 · t234 seam 충돌 · 스위트 510초 비용): `m4-close.md` §7
- sync-audit 대기

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-26
sync_commit_sha: "pending-backfill"  # real SHA backfilled in the immediately following commit (D3 exemption)
b12_self_test_a: pass   # grep -c 'SPEC-CODEX-VERDICT-SYNTH-001' CHANGELOG.md -> 0 (pre-emission)
b12_self_test_b: pass   # AC count: acceptance.md AC-CVS-001..006 = 6 ACs; CHANGELOG entry states 6 PASS 0 FAIL
b12_self_test_c: pass   # cited paths verified: internal/cli/mcp_codex.go, codex_verdict_regression_test.go, codex_review_rpc_test.go, .moai/reports/t229/succession.md
changelog_entry_position: "[Unreleased] -> ### Fixed"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (3-phase close, merged into the single sync commit)"
  updated_field: "2026-08-25 -> 2026-08-26"
canary_compliance_check: not-applicable  # no forward-looking policy self-tested by this SPEC
```

**Verification basis.** §E.2 suite/mutant evidence (suite `ok internal/cli 510.756s` read from the log body line; mutant (e) independence observed; `mcp_codex.go` sha256 pin OK) + successor-session post-merge re-verification per `.moai/reports/t229/succession.md`: origin/main merged clean as `4561f432c`, targeted tests re-run on the merged tree (`ok internal/cli 1.034s`, 34 cases). **Full-package verdict is deferred to PR CI per the verification-load discipline — an honest gap, not a claim.**

**Scope deferral record.** The card↔SPEC scope gap (participant-count axis deferred to a new card by operator decision, 2026-08-26) is recorded in `.moai/reports/t229/succession.md` § "발견 — 카드↔SPEC 범위 갭"; no CHANGELOG content is emitted for the deferred axis.

**Code-read verification before CHANGELOG emission (B12).** Read on this tree: `synthesizeReviewOutput(reviewText, method)`, `codexVerdictSignalsOf` (3 signals: stated verdict label / scored verdict line / severity-tagged finding bullet), `adoptConservativeVerdict` (P-CONS set-max rule, order-independent), `codexUnrecognizedVerdict` (native review/start keeps pass; adversarial turn/start → inconclusive), `describeSignalDivergence` (SynthesisNote recorded only on genuine disagreement), plus both test files.
