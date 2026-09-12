# t525 흡수 사전 판독 — 병합 전 멈춤

트리: `.claude/worktrees/t525`, 브랜치 `WT-speclint-red`, HEAD `7151990276d171feada1aa9e18e582279fa19c4b`, 추적 수정 0.
흡수 대상: 로컬 develop `1412c57383cefe2f41b4a62004485670f51437ad`. merge-base `19cf21408d6992a5cd2acc28291f11cd5d9dc3fc`.
**병합은 시작하지 않았다.** 아래 두 가지가 리드 판단 사항이다.

## 1. 충돌 예측

`git merge-tree --write-tree --name-only develop HEAD` → exit 1:

```
CONFLICT (content): Merge conflict in .github/workflows/spec-lint.yml
Auto-merging internal/cli/spec_lint.go
CONFLICT (add/add): Merge conflict in internal/cli/spec_lint_test.go
```

### internal/cli/spec_lint_test.go (add/add)

- develop 쪽: 헬퍼 `newFixtureProject`·`runSpecLint`·`parseFindings`·`summaryLine`·`countCode`·`codeSeverityCensus`, 타입 `finding`, 상수 `prerepairPathFormCensus`·`fixtureSpecBody`, 변수 `findingLineRe`, 테스트 `TestSpecLint_IDArg_*` 등 12개.
- 브랜치 쪽: 헬퍼 `sl*` 접두사 10개, 타입 `slBaselineFile`, 상수 `slOutOfScopeHeading`·`slCleanSpecMD`·`slProgressMD`, 테스트 `TestSpecLintBaseline_*`·`TestSpecLintUpdateBaseline_*`·`TestSpecLint_DefaultAndStrictBehaviourUnchanged` 12개.
- 최상위 이름 충돌: 두 파일 사이 0. 브랜치 이름을 develop `internal/cli/*.go` 에서 찾기 → 0건. develop 이름을 브랜치 `internal/cli/*.go` 에서 찾기 → 브랜치 자신의 `TestSpecLint_DefaultAndStrictBehaviourUnchanged` 1건뿐(develop 쪽 이름이 아니라 `TestSpecLint_` 접두사 패턴에 걸린 것).
- import 합집합: `bytes crypto/sha256 encoding/hex encoding/json errors os os/exec path/filepath regexp sort strings testing`.
- 판독: 두 집합을 그대로 합치면 되고, 중복 함수는 없다. 컴파일은 합친 뒤에만 확인할 수 있다.

### .github/workflows/spec-lint.yml (content)

- develop(`1cfc6f544` "fix(ci): resolve review and release gate regressions"): 통합 ref 두 개를 fetch 하고 `Select SPEC lint policy` 단계가 이벤트로 strict 여부를 정한다 — release 스냅숏 PR(head 가 origin/develop 과 같아야 함)과 SPEC 입력을 건드리지 않은 develop push 는 `strict=false`, 나머지는 `strict=true`. 단계 둘: `Run strict SPEC lint`(`--strict`), `Run error-only SPEC lint`(플래그 없음).
- 브랜치(t525): 트리거 경로 두 목록에 `.moai/spec-lint-baseline.json` 추가, 실행 줄을 `--strict` 에서 `--baseline .moai/spec-lint-baseline.json` 으로 교체.
- 두 변경은 **같은 문제**(상시 경고 때문에 `--strict` 가 늘 빨갛다)를 다른 방식으로 푼다. CLI 는 `--baseline` 과 `--strict` 를 함께 받지 않는다(`internal/cli/spec_lint.go:163-164`, exit 3).
- AC-SLGS-011 실패 모양: "배선이 여전히 벌거벗은 `--strict` 이거나, 녹색 run 의 로그에 상태 줄이 없으면 미달".

## 2. 기준점 이후 새 규칙 코드

측정: `git grep -hE 'Code:[[:space:]]*"[A-Za-z]+"'` 와 `git grep -hE -A2 'Code\(\) string'` 을 `19cf21408`·`develop` 의 `internal/spec`(테스트 파일 제외)에 돌려 이름을 모았다. 기준점 30개, develop 34개, 사라진 코드 0.

| 코드 | 들여온 커밋 | 카드 | 심각도 | Advisory |
|---|---|---|---|---|
| `ModalityUnjudged` | `458fc7ebc` | t518 | warning | `true` (`lint.go:855`) |
| `REQTableRowsRejected` | `fc0540f41` | t518 | warning | `true` (`lint_req_table_rejection.go:150`) |
| `OwnershipTransitionUnmeasured` | `a6274068e` | t572 | info | 해당 없음 — strict 는 warning 만 올린다(`lint_ownership.go:420-427`) |
| **`DuplicateAcceptanceID`** | `d47dbf6da` | **t564** | **warning** | **없음** — 파일 주석 "Warning, not advisory, so --strict escalates it"(`lint_duplicate_acid.go:18,37`) |

판독: **t518 밖에서 들어온 비-advisory 규칙 코드가 1개 있다(`DuplicateAcceptanceID`, t564).** 리드 규칙에 따라 M3.4 에 들어가지 않는다.

## 미검증

- 코드 수집은 두 형태(`Code: "X"` 리터럴, `Code() string` 반환)만 본다. 상수 경유 등 다른 형태로 선언된 코드는 빠질 수 있다.
- 흡수 트리에서 실제 경고 인구, `DuplicateAcceptanceID` 발견 수, `--strict` 종료 코드는 재지 않았다(병합 전이라 흡수 트리가 없다).
- grandfathered SPEC 의 경고는 era 강등으로 advisory 가 되므로(`lint.go:368-371`), 이 코드가 실제로 게이트에 걸리는 수는 V3R6 SPEC 에서의 발견 수로만 정해진다.
