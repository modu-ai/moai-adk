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
  - 측정 실행 시점에 HEAD가 `0538585ac`로 진행해 있었다(레인 오케스트레이터가 measurement-baseline.md에 트레일러 파싱 위험 기록을 추가한 커밋 — `.md` 단일 파일, 피시험 Go 코드 `lint_ownership.go`·`lint.go`는 두 커밋 간 바이트 동일). RED 관측의 코드 baseline은 `b642479ec`로 유효.

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

## AC-OWN-006 — 규칙 문서 정렬 (M3)

- **재작성 대상**: `## OwnershipTransitionRule Cross-Reference` 절 — 트레일러 WHO SSOT 서술, 발견 코드 3종(Invalid Warning / Unreachable Info / 신규 Unmeasured Info), 설계적 침묵 3지점과 사유, strict 승급의 정확한 서술(Warning-only, Info 무영향). 템플릿 중립 산문 — 신규 추가 식별자 없음 (아래 스캔).
- **순서**: 템플릿 사본 먼저 편집 → 동일 내용 로컬 사본 → 두 사본 같은 커밋(acceptance §D.1 (1)의 단일 커밋 대체 판정 — diff에 템플릿 경로 포함 + 내용 동일).
- **(1) 쌍둥이 일치 (편집 후)**:

```
$ cmp -s internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md .claude/rules/moai/development/spec-frontmatter-schema.md; echo "cmp_rc=$?"
cmp_rc=0
(양 사본 263행)
```

- **(2) 구 트리거 부재 — `grep -c "subject prefix"` (출력된 계수 판정, 종료 코드 무시 — F4 처분)**:

```
$ /usr/bin/grep -c "subject prefix" internal/template/templates/.../spec-frontmatter-schema.md
0
$ /usr/bin/grep -c "subject prefix" .claude/rules/moai/development/spec-frontmatter-schema.md
0
```

  (과정 기록: 첫 재작성본이 "Commit subject prefixes are NOT consulted" 문장에서 substring 매치 1건을 만들어 가드에 잡혔고, "The commit subject is never consulted as the WHO signal"으로 재서술해 0으로 확정 — 부재-가드가 실동작한 사례.)
- **(3) 중립성 토큰 스캔**: `/usr/bin/grep -c "t572\|OWNERSHIP-SILENCE"` (템플릿 사본) → 0. 카드 id·내부 SHA·내부 날짜 신규 추가 없음.
- **(4) 템플릿 중립성·유출 가드** (문구 재서술 후 재실행):

```
$ go test ./internal/template/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	34.687s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.470s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	1.005s
```

- **(5) drift 정정 기록**: spec.md HISTORY v0.1.0 행 + progress.md §E.1 (배차 전제 정정·세 갈래 판정) — 규칙 문서 자체에는 내부 식별자를 쓰지 않음 (위 (3)).
- **(6) 불접촉 확인**:

```
$ git diff --name-only b642479ec HEAD  (+ M3 워킹 트리)
[6 커밋 파일 + 규칙 문서 쌍둥이 2]
manager-develop.md: 없음 / .codex/agents/moai/*: 없음
```

- **always-loaded 비용 규율** (rule-authoring.md §statement duty): 절 크기 증가 ~600 바이트 < 1,000 바이트 단일 편집 임계 — 비용 진술 의무 미발화.

## E3 — 커버리지

```
$ go test -cover ./internal/spec/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	86.419s	coverage: 90.6% of statements
```

- 트리 SHA: a6274068e + M3 워킹 트리 / 목표 85% 이상 유지. (동일 실행의 변경 전 -cover 수치는 미측정 — Gap 기록. 신규 경로 전체가 RED→GREEN 테스트+뮤턴트로 도달 입증됐으므로 감소 위험은 새 코드 축에 없다.)

## E5 — lint 상태 (상속 vs 신규 구분)

```
$ golangci-lint run ./internal/spec/... --timeout=5m
internal/spec/zz_t528_overacceptance_test.go:100:15: Error return value of `f.Close` is not checked (errcheck)
	defer f.Close()
1 issues:
* errcheck: 1
$ echo "real_gcl_rc=$?"   # (파이프 없는 재실행)
real_gcl_rc=1
```

- **귀속**: 상속 — `git diff b642479ec HEAD -- internal/spec/zz_t528_overacceptance_test.go` = 0행 (파일 무변경, 마지막 터치 = t528 카드 커밋 `13fda0f6e`). 본 카드(diff 외부 파일)와 무관, t577 errcheck 축 소관. **신규 이슈 0건.**

## AC-OWN-004 보강 — 변경 후 코퍼스 lint (plan.md §E 보조 관측 + spec.md §7 Gap 종결)

- **커맨드**: `go run ./cmd/moai spec lint --strict` (백그라운드 실행; `go run`이 본 트리에서 빌드하므로 판정 빌드=측정 트리 — VCI §2.2 두 번째 좌표 성립)
- **verbatim 출력 (꼬리)**:

```
1 error(s), 4718 warning(s)
exit status 1
lint_rc=1
```

- **전문**: `.moai/reports/t572/postlint-strict.txt` (4,924행) / 트리 SHA: a6274068e + M3 워킹 트리
- **baseline 대비 (baseline: `0 error(s), 4698 warning(s)`, rc=1 @ 3ac58b5a1 — measurement-baseline.md)**:

| 항목 | baseline (3ac58b5a1) | 변경 후 (본 트리) | 델타 귀속 |
|---|---|---|---|
| 종료 코드 | 1 | **1** | **변경 없음** — Info 발견이 exit을 뒤집지 않음의 실측 보강 (1차 판정자는 AC-OWN-004 유닛 테스트) |
| error | 0 | 1 | 본 SPEC spec.md의 `MissingExclusions` 1건 — **plan-phase 산물** (§7 gap 참조) |
| warning | 4,698 | 4,718 | +20 = 본 SPEC 디렉터리의 plan-phase 콘텐츠(CoverageIncomplete 10 + ModalityUnjudged 10) — 정확히 상계, run-phase 코드 기인 0 |
| info (신규) | 0 | 199 | **본 카드의 산출** — 아래 볼륨 실측 |

- **unmeasured 볼륨 (spec.md §7 Gap 종결)**: **199건 — 199개 SPEC 각 1건씩** (룰이 창 안 최신 전환 1건만 판정). 전부 INFO·advisory — error/warning 계수 어디에도 불산입. per-SPEC 목록: `.moai/reports/t572/unmeasured-per-spec.txt`. 유한하다는 §3.3 예측의 실측 확정.
- **도그푸드 자기검증 (REQ-OWN-010 관측점)**: 본 카드 SPEC(`.moai/specs/SPEC-OWNERSHIP-SILENCE-001/`)의 OwnershipTransitionUnmeasured = **0건** — 본 카드의 plan/M1/M2 커밋들이 `Authored-By-Agent:` 트레일러를 달아 **수리 후 실제로 측정된 첫 전환**이 됐고, 매트릭스 판정 결과 통과(manager-develop가 draft → in-progress 수행 = 정당 소유자). 본 카드의 커밋 트레일러 검증: `git log -1 --format='%(trailers:key=Authored-By-Agent,valueonly)'` → M1/M2 각 `manager-develop` (위 각 커밋 직후 관측).

### 발견된 plan-phase 결함 기록 (run-phase 소관 아님 — B4 귀속 보고)

- `MissingExclusions` ERROR 1건: `.moai/specs/SPEC-OWNERSHIP-SILENCE-001/spec.md` — "'Out of Scope' section has no items — minimum one item required". spec.md §6 본문 구조의 조건으로, **manager-develop는 SPEC 본문을 수정할 수 없어**(B4, spec-frontmatter-schema.md § Forbidden ownership crossings) run-phase에서 수리하지 않는다. manager-spec 소관 — 리드 경유 귀속 보고. 본 카드의 변경 전 baseline 측정(3ac58b5a1)에는 본 SPEC 디렉터리가 없었으므로 이 error는 plan 커밋(b642479ec) 이후 존재 — run-phase 도입 결함이 아니다.

## AC-OWN-008 — 증거 경로 색인

| 파일 | 담당 AC |
|---|---|
| `.moai/reports/t572/run-evidence.md` (본 파일) | AC-OWN-001(RED 4요소) / 002 / 003 / 004(green+RED-now) / 005(뮤턴트 3건) / 006 / 007 / E2·E3·E5 |
| `.moai/reports/t572/run-preflight.md` | Section C 재측정 (C.1 좌표 / C.3 쌍둥이 / C.4 baseline GREEN / 인벤토리) |
| `.moai/reports/t572/postlint-strict.txt` | AC-OWN-004 보강 (변경 후 코퍼스 lint 전문 4,924행) |
| `.moai/reports/t572/unmeasured-per-spec.txt` | spec.md §7 Gap 종결 (unmeasured 199 SPEC×1건 목록) |
| `.moai/reports/t572/golangci-lint.txt` | E5 (상속 errcheck 1건 원문) |
| `.moai/reports/t572/measurement-baseline.md` | plan-phase 1차 근거 (본 카드가 인용) |
| 최종 AC 매핑 표 | progress.md §E.2 |

→ 전 증거가 `.moai/reports/t572/` 트래킹 경로에 반출. `/tmp` 반출 0건.
