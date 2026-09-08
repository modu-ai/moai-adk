# t572 run-evidence — SPEC-OWNERSHIP-SILENCE-001 (§E 자가 검증 원장)

- 트리: `.claude/worktrees/t572`, branch `WT-ownership-lint-silent`
- 시작 커밋(plan-phase): `b642479ec`
- 이 파일은 AC-OWN-001..008의 4요소 기록 원장이다. 각 항목: 커맨드 + verbatim 출력 + 종료 코드(별도 필드) + 트리 SHA.

---

## AC-OWN-001 — RED-now (4요소, right-reason)

- **RED 사유**: 무음 nil 분기(`lint_ownership.go:414-416`)가 발견을 반환하지 않기 때문이다. 컴파일 오류도, 무관 테스트 충돌도 아니다 — 두 서브테스트 모두 `expected exactly 1 OwnershipTransitionUnmeasured finding, got 0: []` 로 실패 = 분기가 조용한 nil로 통과하는 것이 관측된 원인이다.
- **(a) 커맨드** (단일 호출 형태):

```
go test ./internal/spec/ -run TestOwnershipTransitionUnmeasured -count=1
```

- **(b) verbatim stdout**:

```
--- FAIL: TestOwnershipTransitionUnmeasured (0.00s)
    --- FAIL: TestOwnershipTransitionUnmeasured/trailer_absent_emits_unmeasured (0.00s)
        lint_ownership_test.go:647: expected exactly 1 OwnershipTransitionUnmeasured finding, got 0: []
    --- FAIL: TestOwnershipTransitionUnmeasured/none_to_draft_emits_unmeasured_with_none_literal (0.00s)
        lint_ownership_test.go:701: expected exactly 1 OwnershipTransitionUnmeasured finding, got 0: []
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.419s
FAIL
```

- **(c) 종료 코드**: `1`
- **(d) 트리 SHA**: `b642479ec` (plan-phase 커밋 — 테스트 파일 추가분은 미커밋 워킹 트리, 피시험 코드는 정확히 이 커밋의 pre-fix 상태)

---

## AC-OWN-002 — 발급 + 보존 (GREEN)

- **주장**: 피의자 분기만 발급으로 바뀌었고 세 무음 지점(`:400-402`, `:405-407`, `:419-422`)과 형제 발견은 동작 불변.
- **(a) 커맨드**: `go test ./internal/spec/ -run 'TestOwnershipTransition|TestSkipOptOut|TestCommitOwnerKind|TestParseAuthoredByAgent' -count=1 -v`
- **(b) verbatim 출력 (판정 행)**:

```
--- PASS: TestOwnershipTransitionRule_Pass (0.00s)        # 보존: 매트릭스 통과 경로 (7 서브테스트)
--- PASS: TestOwnershipTransitionRule_Fail (0.00s)        # 보존: Invalid 발급 경로 (5 서브테스트)
--- PASS: TestOwnershipTransitionRule_UnreachableGit (0.00s)  # 보존
--- PASS: TestOwnershipTransitionRule_NoTransition (0.00s)    # 보존: rec==nil 무음 지점
--- PASS: TestOwnershipTransitionRule_EmptyFrontmatter (0.00s) # 보존
--- PASS: TestOwnershipTransitionDetected (0.00s)         # 갱신된 서브테스트 포함 (아래 AC-OWN-003)
--- PASS: TestOwnershipTransitionUnmeasured (0.00s)       # 신규: 2 픽스처 — prev-bearing + (none)→draft
--- PASS: TestOwnershipTransitionUnmeasuredStrictSafe (0.00s) # 신규: AC-OWN-004
--- PASS: TestSkipOptOut (0.00s)                          # 보존: lint.skip 옵트아웃
PASS
ok  	github.com/modu-ai/moai-adk/internal/spec	1.363s
```

- **(c) 종료 코드**: `0`
- **(d) 트리 SHA**: `c7b940e45` + M2 워킹 트리 (M2 커밋 직전 상태)
- **메시지 5요소**: `trailer_absent_emits_unmeasured` 서브테스트가 SPEC id / `"in-progress" → "implemented"` / SHA / subject / `no Authored-By-Agent trailer` 5요소를, `none_to_draft...` 서브테스트가 `"(none)" → "draft"` emptyOrValue 경로를 각각 단언 — 전부 GREEN.
- **diff 스코프 판독**: `git diff --stat` → `lint_ownership.go` 1파일(발급 분기+주석 2곳) + `lint_ownership_test.go` 1파일. 보존 분기 3곳의 본문과 형제 발견 블록은 diff에 없음(아래 AC-OWN-007의 diff 스코프 항목과 동일 판독).

## AC-OWN-003 — 무음 단언의 의도적 갱신

- **갱신된 테스트 (전수 — 인벤토리 1건)**:

| 테스트 | 갱신 전 | 갱신 후 | 사유 주석 |
|---|---|---|---|
| `TestOwnershipTransitionDetected/trailer_absent_silent_skip` | 무음 스킵 단언 (Invalid 0건만 검사) | `trailer_absent_emits_unmeasured` — Unmeasured Info 1건 + Info 등급 + Invalid 부재 단언 | `// intentional behavior change (SPEC-OWNERSHIP-SILENCE-001 REQ-OWN-006): silent skip → unmeasured Info.` (서브테스트 선두 3행 블록) |

- **형제 스윕**: `grep -n 'AuthoredByAgent'` 전수 — 빈 리터럴은 `:556` 1건뿐(테이블 행 12건 전부 비어있지 않은 값, plan-audit과 일치). 사유 없이 바뀐 단언 0건. 부모 테스트 doc comment의 "Negative fixture" 문장도 같은 사유 표기로 갱신.
- **단언 강화 확인**: 갱신은 단언을 약화시키지 않았다 — 옛 단언(Invalid 0)이 검사하던 것을 포함하고(Invalid 부재 유지) 발견 수 1건+등급까지 추가 단언.

## AC-OWN-004 — strict 안전 (F1 처분: RED 관측은 m2 주입 시점)

- **green-path (선언+관측)**: `TestOwnershipTransitionUnmeasuredStrictSafe` — (1) 신규 Info 단독 Report + Strict=true → `HasErrors()==false`, (2) 실제 Check() 반환 Report + Strict=true → `HasErrors()==false`, (3) Invalid(Warning)와 별개 코드·별개 등급 단언. verbatim PASS: `--- PASS: TestOwnershipTransitionUnmeasuredStrictSafe (0.00s)` (위 AC-OWN-002 출력 블록 내).
- **RED-now 셀 (관측 — m2 뮤턴트 주입 시점)**: 아래 AC-OWN-005 m2 블록의 4요소가 이 AC의 RED 관측이다. pre-fix 트리에는 주제 코드가 존재하지 않아 구현 전 RED는 구성 불가(verification-completeness §2.1 undecidable 성격 — plan-audit F1 수리 제안 A 채택, progress.md §F 기록).

## AC-OWN-005 — 변이 증거 (3뮤턴트 전부 검출)

### m1 — 발급 분기 삭제 (원래 nil 복원)

- **주입**: `if rec.AuthoredByAgent == "" { return nil }` (발급 블록 앞, `if false` 가드로 발급 블록 봉인)
- **커맨드**: `go test ./internal/spec/ -run TestOwnershipTransitionUnmeasured -count=1`
- **verbatim FAIL**:

```
--- FAIL: TestOwnershipTransitionUnmeasured (0.00s)
    --- FAIL: TestOwnershipTransitionUnmeasured/trailer_absent_emits_unmeasured (0.00s)
        lint_ownership_test.go:664: expected exactly 1 OwnershipTransitionUnmeasured finding, got 0: []
    --- FAIL: TestOwnershipTransitionUnmeasured/none_to_draft_emits_unmeasured_with_none_literal (0.00s)
        lint_ownership_test.go:718: expected exactly 1 OwnershipTransitionUnmeasured finding, got 0: []
--- FAIL: TestOwnershipTransitionUnmeasuredStrictSafe (0.00s)
    lint_ownership_test.go:784: expected the OwnershipTransitionUnmeasured finding from Check(), got none
```

- **종료 코드**: `1` / **판정**: 검출 — 도달성 증명(테스트가 이 분기를 실제로 지난다)
- **원복**: Edit 복구 → `git diff` 내 `MUTANT` 마커 0건 확인 (아래 원복 확인 참조)

### m2 — 등급 Warning 승급 (AC-OWN-004의 RED 관측을 운반)

- **주입**: 발급 블록 `Severity: SeverityInfo` → `Severity: SeverityWarning`
- **커맨드**: `go test ./internal/spec/ -run TestOwnershipTransitionUnmeasuredStrictSafe -count=1`
- **verbatim FAIL**:

```
--- FAIL: TestOwnershipTransitionUnmeasuredStrictSafe (0.00s)
    lint_ownership_test.go:788: expected HasErrors()==false with Strict=true for the Check() report, got true: [{File:.moai/specs/SPEC-FOO-001/spec.md Line:1 Severity:warning Code:OwnershipTransitionUnmeasured Message:SPEC SPEC-FOO-001 transition "in-progress" → "implemented" expected owner "manager-docs" but commit feedface1234 (feat(SPEC-FOO-001): M5 close-out implementation) has no Authored-By-Agent trailer — ownership transition unmeasured Advisory:false}]
    lint_ownership_test.go:797: unmeasured finding must be Info, not Warning (Warning is strict-escalatable), got: warning
```

- **종료 코드**: `1` / **판정**: 검출 — strict 승급 조건(lint.go:62) 회귀를 잡는 유일 판정자가 실동작 확인. **이 4요소(커맨드/verbatim FAIL/종료 코드 1/트리 SHA = c7b940e45 + m2 워킹 트리)가 AC-OWN-004의 RED-now 기록이다.**
- **원복**: Edit 복구 (동일 확인)

### m3 — emptyOrValue 제거 (원시 값 사용)

- **주입**: 발급 메시지의 `emptyOrValue(rec.PreviousStatus)` → `rec.PreviousStatus`
- **커맨드**: `go test ./internal/spec/ -run 'TestOwnershipTransitionUnmeasured/none_to_draft' -count=1 -v`
- **verbatim FAIL**:

```
    lint_ownership_test.go:732: expected finding message to contain "\"(none)\" → \"draft\"", got: SPEC SPEC-FOO-001 transition "" → "draft" expected owner "manager-spec" but commit a1b2c3d4e5f6a7b8 (feat(SPEC-FOO-001): plan-phase artifacts (M, 3 artifacts)) has no Authored-By-Agent trailer — ownership transition unmeasured
--- FAIL: TestOwnershipTransitionUnmeasured (0.00s)
    --- FAIL: TestOwnershipTransitionUnmeasured/none_to_draft_emits_unmeasured_with_none_literal (0.00s)
```

- **종료 코드**: `1` / **판정**: 검출 — plan §G 예측대로 prev-bearing 픽스처는 PASS, `(none) → draft` 픽스처만 FAIL ("픽스처 선택이 판정을 만든다"의 실증; 2-픽스처 설계가 m3를 관측 가능하게 만듦)
- **원복**: Edit 복구 (동일 확인)

### 원복 확인 (3뮤턴트 공통)

```
$ git diff internal/spec/lint_ownership.go | /usr/bin/grep -c 'MUTANT'
0
$ git diff --stat
 internal/spec/lint_ownership.go      | 46 ++++++++++++++-----
 internal/spec/lint_ownership_test.go | 89 ++++++++++++++++++++++++++++++++++--
 2 files changed, 120 insertions(+), 15 deletions(-)
```

→ 뮤턴트 흔적 0건, diff는 의도한 M2 변경만 담음. 생존 뮤턴트: 없음 (m1/m2/m3 전부 검출).

## AC-OWN-007 — 스코프된 검증 (M2 시점)

- **커맨드 + verbatim**:

```
$ go vet ./internal/spec/ && echo "VET_CLEAN_rc=0"
VET_CLEAN_rc=0
$ GOOS=windows GOARCH=amd64 go build ./internal/spec/ && echo "WIN_BUILD_rc=0"
WIN_BUILD_rc=0
$ go test ./internal/spec/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	109.590s
```

- **종료 코드**: 전부 `0` / **트리 SHA**: `c7b940e45` + M2 워킹 트리
- **전체 스위트 미실행**: 의도다(레인 부하 규율 — CI 몫). 로컬 `go test ./...` 실행 0회.
- **상속 적색 분리**: baseline(`b642479ec` 기준 `ok ... 78.980s`)이 GREEN이었으므로 internal/spec 패키지에 본 카드 이전 붉은 테스트 없음 — 분리 기록할 상속 적색 없음. CI spec-lint 잡·t577 errcheck 축의 develop 적색은 본 판정 축 밖(dispatch B5).

## AC-OWN-006 — (M3에서 아래에 추가)

## AC-OWN-008 — 증거 경로 색인

- 본 파일: `.moai/reports/t572/run-evidence.md` (AC-OWN-001/004-RED/005/007 원장)
- `.moai/reports/t572/run-preflight.md` (Section C 재측정 — C.1 좌표/C.3 쌍둥이/C.4 baseline GREEN)
- `.moai/reports/t572/measurement-baseline.md` (plan-phase 실측 — 본 카드가 인용하는 1차 근거)
- 최종 AC 매핑 표: progress.md §E.2
