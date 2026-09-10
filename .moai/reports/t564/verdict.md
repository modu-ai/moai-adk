# t564 — 인라인 AC 파서의 중복 ID 처리 판정 기록

card: t564 · class B · lane-9
branch: WT-dup-ac-id · base: `origin/develop` `d060e0d13` 에서 생성 → 로컬 develop `c3b931784` 흡수(`d02b892c2`, 트리 `36376d5fb` = develop 트리)
출처: 큐 카드 t564(t528 파생, 문안 저자 lane-4) — `internal/spec/parser.go` `buildTree` 가 같은 AC ID 를 두 번째로 만나면 오류를 기록하고 그 줄을 버린다(먼저 온 줄이 이긴다).

## 1. develop 재현 · 두 소비자의 비대칭

### Claim
카드의 결함은 develop 트리에서 그대로 재현된다.
- 같은 ID 를 인용하는 불릿이 진짜 선언보다 위에 있으면 진짜 선언이 버려진다. lint 는 중복을 전혀 보고하지 않고, 버려진 선언의 REQ 매핑 때문에 `CoverageIncomplete` 오탐을 낸다.
- 같은 사건에서 `moai spec view` 는 하드 에러로 종료한다.
- 선언이 위에 있으면 lint 결과는 우연히 맞지만 `spec view` 는 똑같이 죽는다. 비대칭은 줄 순서와 무관하다.
- 현재 코퍼스 노출은 0건이다(잠복).

수리 방향에는 설계 판단이 들어가므로 §2 에서 멈춘다.

### Evidence
기반 정렬:
- 새 워크트리 → `git rev-parse HEAD` `d060e0d13…`(= `origin/develop`, `refs/remotes/origin/HEAD` → `origin/develop`) → `git status --short` 출력 없음 → `git branch -m WT-dup-ac-id`.
- `git merge --no-ff -m "Merge local develop c3b931784 as t564 base (card t564)" c3b931784` → `merge_exit=0`, `d02b892c2 d060e0d13 c3b931784`, `HEAD^2` = `c3b931784`, `HEAD^{tree}` = `c3b931784^{tree}` = `36376d5fb…`, MERGE_HEAD 없음, index.lock 없음(`base-merge.log`).

코드 위치(흡수 트리 `d02b892c2`):
- `internal/spec/parser.go:137-143` — `if seenIDs[acLine.id] { result.Errors = append(…&DuplicateAcceptanceID{…}); continue }`.
- `internal/spec/parser.go:245` `acIDPattern` — 구분자(`:`·`—`·`–`)만 있으면 어떤 불릿이든 선언으로 받는다. 인용 불릿과 선언 불릿은 문법으로 구분되지 않는다.
- 소비자 1: `internal/spec/lint.go:683` `criteria, _ := ParseAcceptanceCriteria(body, false)` — 오류를 버린다. 결과 `doc.Criteria` 를 쓰는 곳은 `lint.go:1094` CoverageRule(`collectAllREQIDs`) 하나다.
- 소비자 2: `internal/cli/spec_view.go:72-86` — `DanglingRequirementReference`·`MissingRequirementMapping` 만 경고로 넘기고 그 밖의 오류는 `return fmt.Errorf("parse error: %w", err)`.
- 기존 테스트 `internal/spec/parser_test.go:125` `TestParser_AC_04_DuplicateID` 는 오류 타입만 단언하고, 어느 줄이 남는지는 단언하지 않는다.

t561 과의 겹침: `ParseAcceptanceCriteria(` / `buildTree(` 호출처는 `lint.go:683`, `spec_view.go:72`, `parser.go` 자체, 테스트뿐이다. t561 의 형제 추출기(`lint_coverage_sibling*.go`)는 이 파서를 부르지 않는다. 겹치지 않는다.

재현 바이너리: `/tmp/t623-lane9/moai-t561-fixed`(t561 수리 후 빌드, base `8203040b8`).
- `git diff --stat 8203040b8 HEAD -- internal/spec/parser.go internal/spec/errors.go internal/spec/lint.go internal/cli/spec_view.go internal/spec/ears.go` → 출력 없음(exit 0).
- 대조로 `-- internal/spec/lint_coverage_sibling.go` → `5 insertions`.
- 따라서 파서와 두 소비자의 동작은 흡수 트리와 같다. `internal/cli` 컴파일은 하지 않았다.

픽스처 3종(`repro/SPEC-DUPREPRO-00{1,2,3}/spec.md`):
- 실행 전 예측은 `repro.predicted` 에 기록했다.
- 실행 전후로 `.moai/specs/` 에 잠시 복사했다. 사전 부재 확인 `precheck_exit=1`(세 경로 모두 없음), 실행 뒤 삭제 `rm_exit=0`, `git status --short -- .moai/specs` 출력 없음.

| 픽스처 | AC 절 | `spec lint <path> --json` | `spec view <ID>` |
|---|---|---|---|
| A `SPEC-DUPREPRO-001` | 인용 불릿(매핑 없음) → 진짜 선언(`maps REQ-DUPREPRO-001-001`) | exit 0, 발견 1건 `CoverageIncomplete  REQ REQ-DUPREPRO-001-001 is not referenced by any AC`, 중복 ID 발견 없음 | exit 1, stdout 비어 있음, stderr `Parse error: duplicate acceptance criteria ID: AC-DUPREPRO-001-01 at depth 0.` |
| B `SPEC-DUPREPRO-002` | 진짜 선언 → 인용 불릿 | exit 0, 발견 0건 | exit 1, stderr `Parse error: duplicate acceptance criteria ID: AC-DUPREPRO-002-01 at depth 0.` |
| C `SPEC-DUPREPRO-003` | 선언 하나(대조군) | exit 0, 발견 0건 | exit 0, 트리 출력 `└── AC-DUPREPRO-003-01` / `└── AC-DUPREPRO-003-01.a: Given a SPEC: (maps REQ-DUPREPRO-003-001)` |

세 결과 모두 `repro.predicted` 와 같다. 원본 출력은 `repro-{A,B,C}-lint.json`·`.stderr`, `repro-{A,B,C}-view.out`·`.stderr` 에 있다.

코퍼스 노출(`go test ./internal/spec -count=1 -run '^TestT528CoverageAndParseErrorCensus$' -v` → exit 0, `census-develop-before.log`):
- 양성 대조 `positive control: uncovered REQ produces 1 CoverageIncomplete finding (rule fires)`.
- `spec.md walked = 834`, `files yielding >=1 acceptance criterion = 119`, `CoverageRule (CoverageIncomplete) findings = 2010`.
- `parse error kind section-not-found 467`, `DuplicateAcceptanceID / depth / other-fatal occurrences = 0`.

### Baseline-attribution
워크트리 `WT-dup-ac-id` HEAD `d02b892c2`(트리 `36376d5fb`), 이 실행. 재현 바이너리는 위 diff 로 이 트리와 파서·소비자 코드가 같음을 확인했다.

### Gaps
- 인라인 파서의 문법(t528/t565 표면)이 넓어지면 노출이 생길 수 있다. 그 변화는 재지 않았다. 노출 0 은 이 트리의 834개 파일에 대한 값이다.
- lint 쪽에서 다른 규칙이 `doc.Criteria` 를 쓰게 되면 영향이 커진다. 현재 소비는 CoverageRule 하나로 확인했다.

### Residual-risk
- 관찰(이 카드 범위 밖): 대조군 C 의 트리에 `Given a SPEC:` 만 남고 When/Then 이 보이지 않는다. `parseSingleACLine` 의 Given 정규식이 `, When` 을 소비한 뒤 When 정규식이 줄 머리에서 맞지 않는 것으로 보이나, 원인은 확인하지 않았다.

## 2. 멈춤 — 설계 판단이 필요한 지점 (리드 판정 대기)

1. **어느 줄을 남길 것인가.** 인용 불릿과 선언 불릿은 문법으로 구분되지 않는다(`acIDPattern` 은 구분자만 요구한다).
   - (a) 지금처럼 먼저 온 줄을 남기되 그 규칙을 명시한다.
   - (b) 내용이 더 많은 줄(REQ 매핑·Given/When/Then 이 있는 줄)을 남긴다. 휴리스틱이라 새 오판 표면이 생긴다.
   - (c) 둘 다 남기지 않고 오류로만 보고한다. 이 경우 두 소비자 모두 그 AC 를 잃는다.
2. **두 소비자의 비대칭을 어느 쪽으로 맞출 것인가.**
   - (가) lint 가 `DuplicateAcceptanceID` 를 새 발견 코드로 보고하고, `spec view` 는 경고로 내리고 트리를 계속 출력한다. 둘 다 보되 둘 다 죽지 않는다.
   - (나) lint 도 하드 에러로 올린다. 두 쪽이 모두 막는다.
   - (다) `spec view` 만 경고로 내린다. lint 는 계속 보지 못한다.
3. **범위.** (가)는 새 lint 발견 코드와 `internal/cli/spec_view.go` 수정을 포함한다. `internal/cli` 컴파일(리드 슬롯 승인)이 필요하다. 문법 쪽 대응(인용 불릿을 선언에서 빼기)은 t565/t528 표면과 겹친다.
