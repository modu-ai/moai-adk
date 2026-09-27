# sync-audit — SPEC-CI-VERDICT-PRODUCER-001 (card t1268, Tier M)

- Auditor: sync-auditor (독립 재판정, 실행자 자체검증과 별개 판정)
- 측정 지점: worktree `.claude/worktrees/t1268`, branch `WT-ci-verdict-producer`, HEAD `f42c900f7`, merge-base `bf3d5144f`
- 측정 일자: 2026-09-26 · 모든 증거는 이 실행, 이 트리 기준 (verification-claim-integrity.md §2 귀속)

## Overall Verdict: **PASS-WITH-DEBT** (harmonic 89.9 / 100 ≥ Tier M 0.80)

3-phase close는 유효하다. 실행자가 주장한 8개 AC 전부를 감사자가 자체 재실행으로 확인했고, L1(AC-AE-012(c) 재판정 가능성)과 L2(Q2 결합 분리·무의존성 규율) 렌즈도 통과다. PASS가 아닌 PASS-WITH-DEBT인 이유는 수용 기준 §E 게이트("Coverage ≥ 85% for touched packages")를 문자 그대로 읽으면 touched package인 `internal/verify`가 84.6%로 0.4pp 미달이기 때문이다(신규 코드 `HasLocalPass`는 85.7%로 충족, 미달분은 전부 기존 코드 — F5). 이 외 옵션 결함 3건(F1·F2·F3)을 문서화된 부채로 남긴다.

## Claim

1. AC-CV-001..008 전부 PASS다 — 감사자가 실행자의 매트릭스를 믿지 않고 동일 명령을 재실행해 확인했다.
2. AC-AE-012(c)는 작성된 그대로 재판정 가능해졌다 — 확장된 `TestContradictoryEvidenceTrips` limb-(c) 케이스가 진짜 same-head(로컬 pass + CI failure) 조립이고, 이 테스트의 통과가 그 기준의 기계적 재판정이다(L1).
3. 생산자 경로는 감지기·체크포인트·훅과 결합돼 있지 않고, go.mod/go.sum은 merge-base 대비 byte 무변경이며, 새 게이트 장치가 없다(L2).
4. 감점 없이 통과한 차원은 없고, 문서화된 부채 5건(F1–F6 중 optional 4건 + 정보 2건)이 판정에 반영됐다.

## Dimension Scores (flat weighted, harmonic mean)

| Dimension | Score | Verdict | Evidence (verbatim, this run) |
|-----------|-------|---------|-------------------------------|
| Functionality (40%) | 94/100 | PASS | `go test -count=1 -run 'TestContradictoryEvidence' -v ./internal/escalation/` → `--- PASS: TestContradictoryEvidenceCISameHeadTrips (0.01s)` · `--- PASS: TestContradictoryEvidenceCINoTrip (0.05s)` (5 subcases) · `--- PASS: TestContradictoryEvidenceCIFreshness (0.01s)` · `ok github.com/modu-ai/moai-adk/internal/escalation 0.643s` ; `go test -count=1 -run 'TestCIVerdict' -v ./internal/cli/` → 4 tests PASS (FromJSON/Fetch/Degradation 4-fixtures/NoDetectorReference) · `ok ... internal/cli 0.770s` ; `go test -count=1 ./internal/civerdict/ ./internal/verify/` → `ok` × 2 ; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 |
| Security (25%) | 88/100 | PASS | 기록 스키마는 5필드뿐(`head_sha, conclusion, run_id, observed_at, producer` — run_id는 gh databaseId 숫자) — secret 유출면 없음. degradation 4-fixture가 전부 "한 줄 메시지 + exit 0 + 기록 0건"을 단언(`verdictFiles` 0 검증, 소스의 전 degrade 경로가 `Save` 전에 return). `grep -rn "escalation\.\|Checkpoint\|Observe(" internal/civerdict/civerdict.go internal/cli/ci_verdict.go internal/verify/localpass.go` → exit 1(영ヒ트). 단, F1의 파일명 순회 원시동작으로 감점 |
| Craft (20%) | 90/100 | PASS | 원자 쓰기 정확: `os.CreateTemp(dir, ".civerdict-*.tmp")` → 동일 디렉터리 `os.Rename` → deferred cleanup(civerdict.go:104-127). 커버리지 재측정: `go test -count=1 -cover ./internal/escalation/ ./internal/civerdict/` → `coverage: 88.7%` / `coverage: 87.3%` (실행자 주장과 동일); `HasLocalPass 85.7%`. `golangci-lint run --timeout=5m ./internal/civerdict/... ./internal/escalation/... ./internal/cli/... ./internal/verify/...` → `0 issues.` ; `go vet` → clean. RED 근거 진위: RED 출력의 단언 행 번호(272/318/320/343)가 M2 커밋 테스트 파일의 행 번호와 정확히 일치하고, `48901d9b8` 시점 감지기 상태(`notObs := []string{notObservedCIVerdict}` 무조건 시드, operational.go:325)가 RED 실패 문구와 정확히 일치 — 동일 테스트의 pre-GREEN 상태로 인정(단, F4 귀속 표현 유보) |
| Consistency (15%) | 88/100 | PASS | 커밋 체인: 9커밋 전부 `card t1268` 포함 확인(`git log --format='%h %s' bf3d5144f..HEAD`); 2개 실제 전이에 `Authored-By-Agent` 트레일러 확인 — `5984038bf`(in-progress→completed, manager-docs ✓ 소유자 일치), `48901d9b8`(draft→in-progress, manager-develop ✓ 소유자 일치). root.go 등록은 단일 `rootCmd.AddCommand(newCIVerdictCmd(defaultGhRunner))` + 기존 `tools` 그룹 재사용. 하우스 `gh` 셸아웃 패턴·`convergenceFile` 파스 선례·verify store 원자 쓰기 규율 준수. 감점: [SYNC] 미정의 MX 마커(F3), CHANGELOG 스키마 명칭 드리프트(F2), 형제 store의 파일명 살균 규율과의 불일치(F1의 일관성 축) |

**Harmonic mean = 4 / (1/94 + 1/88 + 1/90 + 1/88) = 89.9 / 100** — Tier M threshold 0.80 통과. must-pass 차원(Functionality, Security)은 각각 독립 통과.

## Lead-Style Lenses

### L1 — AC-AE-012(c)는 작성된 그대로 재판정 가능한가 — **PASS**

- 원문 limb: SPEC-AUTONOMY-ESCALATION-001 acceptance.md:94 "(c) a local verification pass and a recorded CI failure on the same head". 재판정 형태: `TestContradictoryEvidenceCISameHeadTrips`.
- 소스 정독 확인: 픽스처 `escalationtest.HeadSHA = "0123456789..."` 가 픽스처 워크트리 `.git/refs/heads/WT-other` 에 기록되고(`escalationtest.go:80`), 감지기 `r.headSHA()` 는 `ReadHead(r.root)` 로 그 값을 읽는다(detector.go:409). `writeLocalPass(fixtureHead)` + `writeCIVerdict(fixtureHead, failure)` — **두 증거 모두 체크포인트 head와 동일**. 약화된 바꿔말하기가 아니다: 단언이 "정확히 1건의 contradictory-evidence 기록", 관측문이 "passed locally" + "CI failed" + head SHA를 명명, not-observed에 ci verdict 부재까지 전부 검사한다.
- limb (e) 보존: `e-no-ci-verdict` 서브테스트가 기록 0건 + `"ci verdict"` not-observed 나열을 단언 — 상수 문구는 "no recorded CI verdict producer" → "no record for the checkpoint head"로 바뀌었으나 이는 plan.md §D가 명시적으로 허용한 변경(같은 커밋에서 테스트 갱신, M2에서 함께 착지)이다.
- RED 진위: RED는 HEAD `48901d9b8` 시점 작업트리에서 캡처됐고(테스트 파일은 미커밋 — 정상적 TDD 순서), 실패 문구·행 번호가 M2 커밋 테스트와 정확히 일치하므로 "같은 테스트의 pre-GREEN 상태"로 인정한다(F4의 표현 유지보수 권고).
- 재판정 자체: `go test -run 'TestContradictoryEvidence'` 전체 PASS(감사자 직접 실행) — **AC-AE-012(c)의 기계적 재판정이 실제로 수행됐다**.

### L2 — Q2 결합 분리 + 무의존성 규율 — **PASS**

- 생산자 경로가 `escalation.Checkpoint`·훅을 부르지 않음: 소스 grep exit 1(영히트) + `TestCIVerdictNoDetectorReference`(ci_verdict.go 소스에서 `escalation.`, `Checkpoint`, `Observe(` 금지 심볼 스캔) PASS. `internal/civerdict`는 잎 패키지로 cli→civerdict, escalation→civerdict 한 방향 임포트 — 순환 없음, plan.md M1이 허용한 sibling-package 선택.
- go.mod/go.sum: `git diff --name-only bf3d5144f..HEAD -- go.mod` / `-- go.sum` → **쌍방 공출력**(실행자의 gomod-guard와 동일 결과 재현).
- gh degradation: 4 fixtures(gh-absent=`exec.ErrNotFound`, gh-fails, gh-unparseable, no-run-for-head) 전부 한 줄 메시지 + 에러 nil(exit 0) + `verdictFiles == 0` — 오류 경로 어디서도 파일을 쓰지 않는다(소스의 전 degrade 분기가 `saveCIVerdict` 도달 전 return).
- 새 게이트 장치 없음: `grep -rn "ci-verdict\|civerdict\|ci_verdict" internal/hook/ .claude/settings.json .moai/config/sections/` → exit 1(영히트). 동사는 어떤 훅에도 배선돼 있지 않다.

## Findings

- **F1 [Medium] [optional]** `internal/civerdict/civerdict.go:84` (`Path`) 및 `:92` (`Save`) / `:52` (`Validate`) — `head_sha`가 살균 없이 파일명으로 쓰인다(`filepath.Join(projectRoot, VerdictDir, headSHA+".json")`). `--from-json` 입력의 `head_sha`가 `"../../x"` 형태면 검증 디렉터리 밖으로 순회해 5필드 JSON으로 기존 파일을 덮어쓸 수 있다(로컬 호출자 권한 한정, 내용은 스키마 구속 — 임의 JSON 덮어쓰기 원시동작). 형제 규율과의 불일치: `internal/verify/store.go:21`은 스냅샷 파일명을 살균한다("sanitized key prefix"). 정상 경로(40-hex SHA)에서는 무해하다. — Required fix: `Validate()`에 head가 40-hex임을 검사하는 한 줄 추가(또는 store.go 선례의 살그 적용). 후속 카드 몫 권장.
- **F2 [Low] [optional]** `CHANGELOG.md:12` — 스키마 서술이 실제 필드와 어긋난다: 기록된 "run URL, source, recorded-at" vs 실제 `run_id`(URL 아닌 databaseId), `producer`, `observed_at`. — Required fix: 해당 괄호 서술을 실제 필드명으로 정정.
- **F3 [Low] [optional]** `internal/escalation/operational.go:369` — `@MX:NOTE: [SYNC]` 마커가 트리 전체에서 유일한 신조형이다(mx-tag-protocol은 에이전트 생성 태그에 `[AUTO]`를 필수로 규정; 트리 내 다른 [SYNC] 사용처 없음). — Required fix: `[AUTO]`로 통일하거나 [SYNC]를 프로토콜에 정의.
- **F4 [Info] [optional]** `progress.md` §E.2 E8 — RED 근거의 귀속이 "HEAD `48901d9b8`"만 적는데, 그 커밋 트리에는 테스트 파일이 없었다(작업트리 미커밋 상태로 측정 — TDD 정상 순서). 진위는 행 번호·감지기 상태 정합으로 확인했으나, 귀속 문구가 "커밋된 트리"로 오독될 여지가 있다. — Required fix: 향후 RED 셀에 "테스트 파일 미커밋 작업트리 상태"를 명시(verification-completeness §2.1 네 요소 중 tree 귀속 정밀도).
- **F5 [Info] [optional]** `internal/verify` 패키지 커버리지 84.6% < 수용 기준 §E 게이트 85% — 미달 함수가 전부 기존 파일에 집중(`store.go:61 Save 66.7%`, `claim_lock.go:131 isStaleClaim 0%` 등 — merge-base `bf3d5144f`에서 이미 존재하던 파일로 구조적 확인), 신규 `HasLocalPass` 85.7%는 충족. 실행자가 §E.2 E3에 정직하게 기재했다. 본 카드 결함이 아니므로 감점 최소화하고 부채로 남긴다.
- **F6 [Info] [optional]** `70a4cab06`(plan 커밋, (none)→draft) — `Authored-By-Agent` 트레일러 없음(소유자 manager-spec). lint가 Info 등급으로만 다뤄 침묵. 후속 전이 2건은 트레일러가 정확하다.

## Evidence-bearing format (VCI §3)

**Claim** — 위 Dimension Scores 표의 4개 행과 L1/L2 각 판정.

**Evidence** — 각 행의 Evidence 셀에 인용한 명령+출력 전체. 주요 재실행: detector 3테스트 + cli 4테스트 + civerdict/verify 패키지, 커버리지 3건, lint/vet/windows-build, spec lint/audit, 가드 3건(git diff 공출력), decoupling grep 2건(영히트), 커밋 체인 열람.

**Baseline-attribution** — 전부 이 실행, 이 트리: worktree `.claude/worktrees/t1268`, HEAD `f42c900f7`, merge-base `bf3d5144f`, 2026-09-26. 실행자 §E.2가 인용한 수치(88.7/87.3/85.7/84.6%)를 감사자가 독립 재측정해 동일값 확인 — carry-over가 아니다.

**Gaps** — 명시적으로 관측하지 않은 것: (1) `internal/cli` 전체 스위트 미실행(레인-로컬 검증 규율 — `TestCIVerdict*` 스코프만 관측, 전체 판정은 origin/develop CI 몫); (2) 실제 `gh` 라이브 fetch 미측정(전 테스트가 주입 runner); (3) RED의 역사적 재실행 불가(정합성으로만 입증); (4) merge-base 시점 verify 패키지 커버리지 수치의 직접 재측정 불가(구조 확인만); (5) `plan-audit-iter-2.md` 본문 미정독(파일 존재만 확인).

**Residual-risk** — (1) gh가 미래에 새 conclusion 어휘를 내놓으면 no-record로 강하해 판정이 조용히 누락될 수 있다(안전하나 누락 — 실행자 §E.3 잔여위험 인용); (2) F1의 순회 원시동작은 악의적 로컬 입력(예: 프롬프트 주입된 에이전트 명령)에서 현실화될 수 있다; (3) `mapGHConclusion`의 failure-family 접기(timed_out/startup_failure→failure)가 운영 관점의 "실패"보 넓어 극단적 오탐 여지는 낮으나 존재.

## Recommendations

- 후속 카드로 F1(head 40-hex 검증)을 권한다 — 한 줄 검증으로 형제 store와의 규율 불일치와 순회 원시동작을 함께 닫는다. F2·F3는 다음 sync 스윕에서 함께 정리 가능한 문서·마커 수준이다.
- `moai ci-verdict`의 자연 호출자(리드 일괄 push 직후) 사용 시 개발 흐름 문서에 한 줄 예시를 추가하면 채택이 빨라진다(본 감사 범위 밖 — 제안).
- docs-site "49 commands" 카운트 플래그: progress.md §E.4에 기록돼 있어 침묵 유실되지 않았음을 확인 — 다음 릴리스 스윕에서 별도 리포지토리 쪽 갱신이 필요하다.

## Verification (spec lint / audit 재실행)

```
$ moai spec lint SPEC-CI-VERDICT-PRODUCER-001
✓ No findings — all SPEC documents are valid            (exit 0)
$ moai spec audit --filter-spec SPEC-CI-VERDICT-PRODUCER-001
Modern-era clean: 1 / Drift findings: 1
  [INFO] SPEC-CI-VERDICT-PRODUCER-001 (V3R6) — EraAutoDetected   (exit 0)
```

INFO EraAutoDetected는 era 미명시에 대한 정보성 통지일 뿐 drift 결함이 아니다. 생산 SPEC에 `era: V3R6` 명시는 선택 사항(옵션 권고, 비차단).

— sync-auditor, 독립 판정 (moai 🗿)
