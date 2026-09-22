# t612 판정 — @MX 스캐너의 대기 WARN 결함 두 건 (F03 · F04)

- 카드: t612 (Class B, SPEC 없음, cycle_type=tdd)
- 브랜치: `WT-mx-pending-warn`
- RED 커밋: `504743a25` — 테스트와 수정 전 로그만 담았고 `scanner.go` 는 건드리지 않았다
- 수정 커밋: 이 판정 파일이 들어 있는 커밋(`504743a25` 의 바로 다음 커밋)

## 주장 (Claim)

1. F03(REASON 없는 WARN 뒤에 WARN 이 아닌 태그가 오면 진단 없이 사라진다)은 develop `d3b7d438d` 에서 재현됐다. 같은 지점에서 C2(REASON 없는 WARN 두 개가 이어지면 첫 WARN 이 진단 없이 사라진다)도 함께 재현됐다.
2. F04(대기 중인 WARN 과 REASON 사이의 `@MX:SPEC` 이 WARN 앞의 태그에 붙는다)도 develop `d3b7d438d` 에서 재현됐다.
3. 수정 뒤 C1–C5 는 모두 요구 상태를 만족하고, `internal/mx` 패키지 전체는 316건이 통과하고 실패는 0건이다.
4. 두 수정은 각각 자기 결함 테스트만 RED 로 만들고, 그 수정과 무관한 대조 셀은 GREEN 으로 남는다(수정별 뮤턴트).
5. `TestAC006_DanglingSpecRefWarning` 은 수정 뒤에도 DanglingSpecRef 분기에 도달한다(분기를 끄면 RED, 되살리면 GREEN).
6. 의도한 동작 변경이 하나 있다. 앞선 태그 없이 대기 중인 WARN 뒤에 오는 `@MX:SPEC` 은 이제 그 WARN 에 붙고, DanglingSpecRef 를 내지 않는다.

## 증거 (Evidence)

모든 명령은 `unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR && go test ./internal/mx/ ... -count=1 -v > <파일> 2>&1` 한 번의 호출로 실행했다. 출력은 파이프로 거르지 않고 파일로 받았다. 전체 로그는 이 디렉터리에 있다.

### 재현 — 수정 전 (`red-before-fix.txt`, exit=1)

F03 (C1) — 경고가 0건이고 NOTE 만 남는다:

```
=== RUN   TestPendingWarn_C1_ReasonlessWarnThenNote_EmitsMissingReason
    scanner_pending_warn_test.go:59: tag line=3 kind=NOTE spec="" reason=""
    scanner_pending_warn_test.go:59: warnings=0 []
    scanner_pending_warn_test.go:62: MissingReasonForWarn for WARN at line 2: want 1, got 0 ([])
--- FAIL: TestPendingWarn_C1_ReasonlessWarnThenNote_EmitsMissingReason (0.00s)
```

C2 — 둘째 WARN(3행)만 진단되고, 첫 WARN(2행)은 진단되지 않는다:

```
    scanner_pending_warn_test.go:72: warnings=1 [MissingReasonForWarn: .../fixture.go:3 - WARN tag without REASON]
    scanner_pending_warn_test.go:75: MissingReasonForWarn for first WARN at line 2: want 1, got 0 (...)
--- FAIL: TestPendingWarn_C2_ReasonlessWarnThenWarn_FirstWarnGetsMissingReason (0.00s)
```

F04 (C4) — 보고서가 main 에서 관측한 것과 같은 모양이다:

```
    scanner_pending_warn_test.go:101: tag line=2 kind=NOTE spec="SPEC-WARN-001" reason=""
    scanner_pending_warn_test.go:101: tag line=3 kind=WARN spec="" reason="shared state"
--- FAIL: TestPendingWarn_C4_SpecBetweenWarnAndReason_AttachesToWarn (0.00s)
```

앞선 태그 없는 대기 WARN 뒤의 SPEC — 수정 전에는 DanglingSpecRef 가 난다:

```
    scanner_pending_warn_test.go:177: warnings=1 [DanglingSpecRef: .../fixture.go:3 - @MX:SPEC "SPEC-WARN-001" without a preceding tag]
--- FAIL: TestPendingWarn_SpecOnPendingWarnWithoutPriorTag_NotDangling (0.00s)
```

같은 실행에서 C3, C5, `TestScannerSpecRef_Capture`, `TestAC006_DanglingSpecRefWarning` 은 PASS 였다.

### 짝 없는 WARN 의 기존 세 출구 측정 — 수정 전 (`exits-measure-before.txt`, exit=0)

F03 의 태그 보존 규칙은 이 측정에 근거한다. 세 출구 모두 경고를 한 건 내고, 태그 로그 줄은 한 줄도 찍히지 않았다(= WARN 이 `tags` 에 추가되지 않았다). 세 출구의 동작은 서로 일치한다.

| 출구 | 픽스처 (행 번호는 스캐너 기준) | 관측 출력 |
|---|---|---|
| E1 창 만료(@MX 가 아닌 줄) | `package fixture` / `// @MX:WARN: danger` / `func a() {}` … `func d() {}` (6행) | `warnings=1 [MissingReasonForWarn: .../fixture.go:2 - WARN tag without REASON within 3 lines]` — 태그 0 |
| E2 늦게 온 REASON | `package fixture` / `// @MX:WARN: danger` / `// @MX:PRIORITY: P1` ×3 / `// @MX:REASON: too late` (6행) | `warnings=1 [MissingReasonForWarn: .../fixture.go:2 - WARN tag without REASON within 3 lines]` — 태그 0 |
| E3 파일 끝 | `package fixture` / `// @MX:WARN: danger` | `warnings=1 [MissingReasonForWarn: .../fixture.go:2 - WARN tag without REASON]` — 태그 0 |

E2 의 중간 줄은 모두 @MX 하위 줄이라 E1 분기(@MX 가 아닌 줄)는 실행되지 않는다. 태그도 없으므로 태그 분기의 만료 검사도 돌지 않는다. 따라서 "within 3 lines" 문구를 낼 수 있는 곳은 늦은 REASON 분기뿐이다.

### 수정 전후 대조 셀

| 셀 | 입력 | 수정 전 | 수정 후 (`green-after-fix.txt`) |
|---|---|---|---|
| C1 | WARN(REASON 없음) → NOTE | FAIL — 경고 0 | PASS — `fixture.go:2 - WARN tag without REASON before the next tag`, 태그는 NOTE 하나 |
| C2 | WARN(REASON 없음) → WARN | FAIL — 3행만 경고 | PASS — 2행 경고(`before the next tag`) + 3행 경고(EOF), 태그 0 |
| C3 | WARN → REASON | PASS | PASS — WARN 하나, reason="shared state", 경고 0 |
| C4 | NOTE → WARN → SPEC → REASON | FAIL — NOTE 가 SPEC 을 가져감 | PASS — NOTE spec="", WARN spec="SPEC-WARN-001" |
| C5 | NOTE → WARN → REASON → SPEC | PASS | PASS — C4 와 같은 최종 상태 |
| (추가) | WARN(대기, 앞선 태그 없음) → SPEC → REASON | FAIL — DanglingSpecRef | PASS — WARN spec="SPEC-WARN-001", DanglingSpecRef 0 |

`green-after-fix.txt`: exit=0, `--- PASS` 11건, `--- FAIL` 0건. 실행된 테스트 11개의 이름이 로그에 모두 찍혀 있다(셀렉터가 0건에 맞아 공허하게 통과한 것이 아니다).

### 수정별 뮤턴트

뮤턴트는 해당 수정의 조건을 `if false && pendingWarnTag != nil` 로 바꿔 그 수정만 끄는 방식으로 만들었다. 매번 원래 코드로 되돌린 뒤 GREEN 을 다시 확인했다.

| 뮤턴트 | 로그 | RED 가 된 테스트 | GREEN 으로 남은 테스트 | 복원 후 |
|---|---|---|---|---|
| F03 끔 | `mutant-f03.txt` (exit=1) | C1, C2 | C3, C4, C5, E1, E2, E3, 대기 WARN SPEC, `TestScannerSpecRef_Capture`, `TestAC006` (9건) | `restore-f03.txt` exit=0, PASS 11 / FAIL 0 |
| F04 끔 | `mutant-f04.txt` (exit=1) | C4, 대기 WARN SPEC | C1, C2, C3, C5, E1, E2, E3, `TestScannerSpecRef_Capture`, `TestAC006` (9건) | `restore-f04.txt` exit=0, PASS 11 / FAIL 0 |

어느 뮤턴트도 전부를 RED 로 만들거나 아무것도 RED 로 만들지 않은 경우가 없다. 각 뮤턴트는 자기 분기에 걸린 테스트만 떨어뜨렸다.

### 가드 도달성 — `TestAC006_DanglingSpecRefWarning`

DanglingSpecRef 경고 추가 코드를 `if false { ... }` 로 감싸 끈 뒤 실행했다(`mutant-ac006.txt`, exit=1):

```
=== RUN   TestAC006_DanglingSpecRefWarning
    scanner_specref_test.go:73: expected DanglingSpecRef warning naming SPEC-X-001, got []
--- FAIL: TestAC006_DanglingSpecRefWarning (0.00s)
```

복원 후 전체 패키지 실행(`full-package-after-fix.txt`)에서 `--- PASS: TestAC006_DanglingSpecRefWarning (0.00s)` 를 확인했다. F04 수정 뒤에도 픽스처 `// @MX:SPEC: SPEC-X-001` 단독 입력은 DanglingSpecRef 분기에 도달한다.

### 전체 패키지 · 정적 검사

| 명령 | 결과 |
|---|---|
| `go test ./internal/mx/ -count=1 -timeout=600s -v` 수정 전 (`full-package-before-fix.txt`) | exit=1, `--- PASS` 312, `--- FAIL` 4 (새 RED 테스트 4건뿐) |
| 같은 명령 수정 후 (`full-package-after-fix.txt`) | exit=0, `--- PASS` 316, FAIL 0, `ok github.com/modu-ai/moai-adk/internal/mx 9.289s` |
| `--- PASS: TestAC002_CoverageLiftOnVsOff (1.01s)` | 600s 제한에서는 시간 초과 없이 통과 |
| `go vet ./internal/mx/` (`vet-final.txt`) | exit=0, 출력 0바이트 |
| `gofmt -l internal/mx/` | exit=0, 출력 없음 |

### 수정 내용 (`internal/mx/scanner.go`)

- F03: 새 독립 태그를 파싱한 직후 `pendingWarnTag` 가 남아 있으면 `MissingReasonForWarn: <file>:<line> - WARN tag without REASON before the next tag` 를 내고 대기를 비운다. WARN 은 `tags` 에 추가하지 않는다. 이 한 곳이 C1(뒤에 오는 태그가 WARN 이 아닌 경우)과 C2(뒤에 오는 태그가 WARN 인 경우)를 함께 닫는다.
- F04: `@MX:SPEC` 하위 줄 분기에서 대기 WARN 을 먼저 주인으로 삼는다. 대기 WARN 이 없으면 마지막으로 추가된 태그를, 둘 다 없으면 DanglingSpecRef 를 낸다. 해당 분기의 `@MX:NOTE` 도 이 소유 규칙에 맞게 고쳤다.
- `MissingReasonForWarn` 문자열을 읽는 Go 코드(`grep -rn "MissingReasonForWarn" internal pkg cmd`)는 모두 접두어만 비교하므로, 새 문구 꼬리는 기존 소비자를 깨지 않는다.

## 기준선 귀속 (Baseline-attribution)

- 보고서 트리: `main` `2213871af` — 보고서가 관측한 결과이며, 이번 작업에서 이 트리를 다시 측정하지는 않았다.
- 작업 트리: 로컬 `develop` `d3b7d438d` 로 fast-forward 한 워크트리. 수정 전 측정(`red-before-fix.txt`, `exits-measure-before.txt`, `full-package-before-fix.txt`)은 모두 이 트리에 테스트 파일만 더한 상태에서 실행했으며, 그 상태가 커밋 `504743a25` 의 트리다.
- 수정 후 측정(`green-after-fix.txt`, 뮤턴트·복원 로그, `full-package-after-fix.txt`, `vet-final.txt`, `gofmt`)은 `504743a25` 위에 `scanner.go` 수정을 더한 트리, 즉 이 판정 파일이 들어 있는 커밋의 트리에서 실행했다.
- 순서 증거: RED 테스트와 수정 전 로그는 수정보다 앞선 별도 커밋(`504743a25`)에 있으므로, 커밋 그래프만으로도 "테스트가 수정보다 먼저 존재했다"를 확인할 수 있다.

## 미검증 (Gaps)

- `main` `2213871af` 에서의 재현은 다시 돌리지 않았다. 보고서의 관측을 인용했을 뿐이다.
- `internal/mx` 밖의 패키지(스캐너 경고를 소비하는 CLI·사이드카·codemaps 쪽)는 실행하지 않았다. 범위 지시에 따른 것이며, 전체 판정은 CI 몫이다.
- darwin 밖의 플랫폼(linux·windows)에서는 측정하지 않았다.
- `MissingReasonForWarn` 문구 소비자 검색은 `internal`, `pkg`, `cmd` 의 Go 코드로 한정했다. 문서·템플릿·외부 도구가 전체 문구를 비교하는지는 확인하지 않았다.
- 실제 저장소를 스캔했을 때 새로 드러나는 경고 수(이전에 조용히 사라지던 WARN 의 개수)는 측정하지 않았다.

## 잔여 위험 (Residual-risk)

- 대기 WARN 이 `@MX:SPEC` 을 받은 뒤 끝내 짝을 찾지 못하면(창 만료·다음 태그·파일 끝) SpecRef 도 WARN 과 함께 버려진다. 수정 전에는 이 SPEC 이 엉뚱한 앞 태그에 붙거나 DanglingSpecRef 를 냈다. 수정 후에는 WARN 에 대한 `MissingReasonForWarn` 만 남고 SPEC 연결 자체에 대한 별도 진단은 없다.
- 새 태그가 대기 WARN 의 하위 줄 창을 닫으므로, `WARN → NOTE → REASON` 처럼 REASON 이 다른 태그 뒤에 오는 입력은 이제 경고를 낸다. 수정 전에도 이 REASON 은 WARN 에 붙지 않았으므로 짝 지어지는 결과는 달라지지 않았지만, 경고 수는 늘어난다.
- 기존 저장소 스캔에서 이전에는 보이지 않던 `MissingReasonForWarn` 이 새로 나타날 수 있다. 결함을 드러내는 의도된 결과지만, 경고 수를 기준으로 삼는 소비자가 있다면 영향을 받는다.
- 의도한 동작 변경: 앞선 태그 없이 대기 중인 WARN 뒤의 `@MX:SPEC` 은 더 이상 DanglingSpecRef 를 내지 않고 그 WARN 에 붙는다(`TestPendingWarn_SpecOnPendingWarnWithoutPriorTag_NotDangling` 로 고정).

## 미결 결정 (Open decision)

카드 문구 "WARN 을 잃지 않고" 는 두 가지로 읽힌다.

1. **진단만 한다** — 짝 없는 WARN 을 조용히 버리지 말고 `MissingReasonForWarn` 으로 알린다. WARN 자체는 `tags` 에 넣지 않는다.
2. **태그를 보존한다** — REASON 이 없더라도 WARN 을 `tags` 에 추가한다.

구현한 것은 1번이다. 근거는 다음과 같다.

- 같은 루프에 이미 있는 짝 없는 WARN 출구 세 곳(E1 창 만료, E2 늦은 REASON, E3 파일 끝)을 측정했더니 셋 모두 진단만 하고 태그를 추가하지 않았으며, 서로 일치했다(위 측정 표). F03 의 새 출구를 이 불변식에 맞췄다.
- 2번은 결함 수리가 아니라 네 출구 전체의 계약 변경이 된다. 리드도 이 규칙을 승인하면서 태그 보존은 원하지 않는다고 확인했다.
- 2번은 구현하지 않았다. 계약을 바꾸려면 네 출구를 함께 바꾸는 별도 카드가 필요하다.
