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

## 3. 리드 판정 · 수리 · RED/GREEN · 뮤턴트 4종 · 슬롯

### 리드 판정(요지)
1. **(d) 같은 id 의 모든 줄에서 매핑을 합집합으로 모으고, 중복 자체를 발견으로 보고한다.** (a)·(b)는 버린 줄의 매핑을 조용히 잃고, (c)는 진짜 선언까지 버려 오탐을 키운다.
2. **(가)** lint 는 발견으로 보고한다. `spec view` 는 경고만 내고 트리를 출력한다(exit 0).
3. 슬롯: `internal/cli` 테스트와 `cmd/moai` 빌드를 1회 사전 확인 뒤 직렬로 묶는 것을 승인. 인용 불릿을 문법에서 빼는 방향은 t565 표면이라 이 카드에서 하지 않는다.

### lint 가 오늘 0건이던 이유
- `internal/spec/lint.go:683`(수리 전) `criteria, _ := ParseAcceptanceCriteria(body, false)` 가 오류 슬라이스를 통째로 버렸다.
- 저장소 전체에서 `DuplicateAcceptanceID` 를 참조하는 곳은 `errors.go`·`parser.go`·테스트뿐이었다(§1 `grep -n 'DuplicateAcceptanceID'` 출력). 발견을 내는 규칙이 없었다.
- 따라서 중복이 있어도 없어도 lint 결과는 같았다. §1 픽스처 A·B 가 그 관측이다.

### Claim
- 수리는 두 커밋이다: `d47dbf6da`(internal/spec), `3f20c89df`(internal/cli + 증거).
- 새 테스트 5개(internal/spec)와 view 테스트 1개가 수리 전에 빨갛거나 바이너리 재현으로 실패가 관측됐고, 수리 뒤 초록이다.
- 뮤턴트 4종이 모두 실행 전에 적어 둔 테스트에서, 적어 둔 이유로 빨개졌다.

### Evidence
수리 내용:
- `internal/spec/errors.go` — `DuplicateAcceptanceID.Line`(반복 줄의 1-기반 줄 번호) 추가. 메시지는 그대로다.
- `internal/spec/parser.go` — `acParsedLine.line` 기록. 중복 줄은 `duplicateReqIDs` 에 매핑을 모으고, 트리 완성 뒤 **자동 감싸기 전에** `mergeDuplicateReqIDs` 로 첫 노드에 합친다(이미 있는 id 는 건너뜀). 첫 줄의 Given/When/Then 문구는 그대로 유지한다.
- `internal/spec/lint.go` — `SPECDoc.DuplicateACIDs` 추가. `parseSPECDoc` 는 중복 오류만 보관하고 나머지 오류는 여전히 버린다. 규칙 목록의 `CoverageRule` 뒤에 `DuplicateAcceptanceIDRule` 등록.
- `internal/spec/lint_duplicate_acid.go`(신규) — 반복 줄마다 `DuplicateAcceptanceID` 경고 1건, 본문 기준 줄 번호. advisory 아님(`--strict` 가 올린다), `eraDemotableCodes` 에 넣지 않음.
- `internal/cli/spec_view.go` — `case *spec.DuplicateAcceptanceID:` 를 경고 출력으로 추가.

RED(`red-before-fix.log`, 수리 전 트리에 테스트만 추가, 예측 `red.predicted`) — exit 1, 예측과 같음:
- FAIL `TestParser_DuplicateID_CitingAboveKeepsDeclarationMapping`: `REQ-DUP-001-001 not collected from the duplicate declaration; covered = map[]`
- FAIL `TestParser_DuplicateID_UnionsMappingsOfEveryLine`: `REQ-DUP-002-001 not collected; covered = map[REQ-DUP-002-002:true]`
- PASS `TestParser_DuplicateID_DeclarationAboveStillReported`(오늘도 맞는 순서의 가드)
- FAIL `TestLintDuplicateACID_ReportsDuplicateAndKeepsCoverage`: `DuplicateAcceptanceID findings = [], want exactly 1`
- PASS `TestLintDuplicateACID_NoDuplicateNoFinding`(대조군)

view 쪽 RED 는 테스트가 아니라 §1 바이너리 재현(`repro-A-view.stderr`, exit 1)으로 관측했다. 슬롯이 1회라 수리 전 트리에서 `internal/cli` 를 따로 컴파일하지 않았다.

GREEN:
- `green-after-fix.log` → 5개 PASS, `ok … 1.338s`.
- `go test ./internal/spec -count=1` → `ok github.com/modu-ai/moai-adk/internal/spec 137.423s`(`green-spec-package.log`).
- `gofmt -l`(변경 Go 파일 8개) → 출력 없음.
- 끝 노드 단언: 첫 GREEN 뒤 테스트 1에 "자동 감싸기된 `.a` 자식이 매핑을 가진다"는 단언을 추가했다. 커버리지 집계는 감싸기 노드까지 돌기 때문에, 합치기를 감싸기 뒤로 옮기는 뮤턴트(M3)가 그 단언 없이는 살아남는다.

뮤턴트(`mutants.predicted` 에 실행 전 기록, 각각 단독 적용 → `go test ./internal/spec -count=1 -run '^(TestParser_DuplicateID_|TestLintDuplicateACID_)' -v` → 원복):

| 뮤턴트 | 바꾼 곳 | 관측 | exit |
|---|---|---|---|
| M1 합집합 끔 | 중복 줄 매핑 수집을 `_ = duplicateReqIDs` 로 | FAIL 3(인용 위 — 커버리지와 끝 노드 둘 다, 합집합, lint `CoverageIncomplete` 1) / PASS 2 | 1 |
| M2 lint 수집 끔 | `parseSPECDoc` 조건에 `&& false` | FAIL 1(lint `findings = [], want exactly 1`) / PASS 4 | 1 |
| M3 합치기를 감싸기 뒤로 | 호출 위치 이동 | FAIL 1(끝 노드 `RequirementIDs:[]`) / PASS 4 | 1 |
| M4 줄 번호 미기록 | `Line: 0` | FAIL 1(`line = 0, want 18 (second occurrence)`) / PASS 4 | 1 |

- 네 결과 모두 예측과 같다.
- M2·M3·M4 적용 중 `git diff d47dbf6da -- internal/spec` 는 각 뮤턴트 줄만 보였다(`mutant-M{2,3,4}-*.diff`). 앞 뮤턴트 원복이 완전했다는 뜻이다.
- 마지막 원복 뒤 `git diff --stat d47dbf6da -- internal/spec` → 0 바이트(`post-mutation-diff.txt`), 5개 재실행 PASS(`green-post-mutation.log`).

슬롯(리드 승인 1회):
- 사전 확인(13:56:29Z): 실행 파일 이름 비교기(`compile`·`link`·`cli.test`, 또는 `internal/cli`·`./cmd/moai` 를 인자로 가진 `go`).
  - 양성 대조: 합성 3줄 중 2줄 매치, `zsh` 줄 제외(`slot-precheck-control.out`).
  - 실제: `ps` 1535줄 중 매치 0(`slot-precheck-real.out`).
- `go test ./internal/cli -count=1 -run '^(TestSpecView_|TestSpecViewPlain_)' -v` → `cli_test_exit=0`. `TestSpecView_TraversalRejected`·`TestSpecView_DuplicateACIDWarnsAndRendersTree`·`TestSpecViewPlain_TreePassthrough`·`TestSpecViewPlain_CommandUsesGlamourGateway` PASS, `ok … 1.177s`(`slot-cli-specview.log`, 13:56:42Z–13:57:01Z).
- `go build -o /tmp/t623-lane9/moai-t564-fixed ./cmd/moai` → `build_exit=0`, 출력 0 바이트(`slot-build.log`, 13:57:13Z). 설치하지 않았다.
- 바이너리 내용: 새 규칙 메시지 `is declared on more than one line` 을 `/usr/bin/grep -a -c` 로 셌다. 새 바이너리 1(exit 0), 이전 `moai-t561-fixed` 0(exit 1). 대조 `CoverageIncomplete` 는 둘 다 1.
- 빌드 시점 트리: `internal/spec` 은 `d47dbf6da` 와 같았다(post-mutation diff 0). `internal/cli` 두 파일은 빌드 뒤 그대로 `3f20c89df` 로 커밋했다.

### Baseline-attribution
워크트리 `WT-dup-ac-id`. RED 는 `3f9ffcdc2` + 테스트 파일, GREEN·뮤턴트는 `d47dbf6da`, 슬롯은 `d47dbf6da` + `internal/cli` 두 파일(= `3f20c89df` 코드), 이 실행.

### Gaps
- view 테스트는 수리 전 트리에서 실행하지 않았다(슬롯 1회). view 쪽 실패는 바이너리 재현으로만 관측했고, view 수정을 되돌리는 뮤턴트는 돌리지 않았다.
- `internal/cli` 패키지 전체는 돌리지 않았다(창 규칙상 금지). spec view 테스트 4개만 실행했다.

### Residual-risk
- **합집합의 트레이드오프(리드 판정 반영).** 합집합은 커버리지를 넓게 본다. 인용 불릿에 `maps REQ-…` 꼬리가 있으면 작성자가 매핑으로 의도하지 않은 REQ 도 커버된 것으로 센다. 그 대가로 중복은 `DuplicateAcceptanceID` 경고로 **반드시 같이** 뜨므로 작성자가 모르고 지나칠 수 없다. 다만 grandfather 시대·종결 상태 SPEC 에서는 경고가 advisory 로 내려가 `--strict` 에서도 막지 않는다.
- 트리에 표시되는 문구(Given/When/Then)는 첫 줄의 것이다. 인용 불릿이 위에 있으면 `spec view` 는 인용 문구를 보여 준다. 매핑만 합집합이다.
- 코퍼스는 이 트리에서 중복 0건이라 규칙이 코퍼스에서 발화하는지는 픽스처로만 확인됐다.

### 범위 밖 카드 후보 — EARS 문장 분해의 When/Then 소실 (발행은 리드)
- 재현 명령: 픽스처 `repro/SPEC-DUPREPRO-003/spec.md` 를 `.moai/specs/` 에 두고 `/tmp/t623-lane9/moai-t561-fixed spec view SPEC-DUPREPRO-003`.
- 입력 줄: `- AC-DUPREPRO-003-01: Given a SPEC, When it is parsed, Then the declaration is kept (maps REQ-DUPREPRO-003-001)`
- 관측 출력(exit 0, `repro-C-view.out`): `└── AC-DUPREPRO-003-01.a: Given a SPEC: (maps REQ-DUPREPRO-003-001)` — When/Then 이 없다.
- 추정(미확인): `parseSingleACLine` 의 Given 정규식 `(?i)^Given\s+(.+?)(?:,\s*(?:When|then)|$)` 이 `, When` 까지 소비한 뒤, When 정규식 `^When\s+` 이 남은 문자열 머리(`it is parsed, …`)에서 맞지 않는 것으로 보인다.

## 4. 코퍼스 수리 전/후 · 픽스처 재실행

### Claim
- 코퍼스 lint 결과는 수리 전후가 **완전히 같다.** 발견 3325건, 규칙별 수 동일, `CoverageIncomplete` TSV 바이트 동일. 사라진 발견·새 발견 모두 0건이다. 코퍼스에 중복 AC id 가 0건(§1 census)이므로 예상한 결과다.
- §1 픽스처 3종을 수리된 바이너리로 다시 돌린 결과가 실행 전 예측(`repro-after.predicted`)과 같다.
  - A·B: lint 는 중복을 경고 1건으로 보고하고 오탐은 없다.
  - A·B: `spec view` 는 경고와 함께 트리를 출력하고 exit 0 이다.
  - C: 수리 전과 바이트 단위로 같다.

### Evidence
측정 바이너리:
- 수리 전 `/tmp/t623-lane9/moai-t561-fixed`(§1 에서 lint·파서 코드가 이 트리와 같음을 확인).
- 수리 후 `/tmp/t623-lane9/moai-t564-fixed`(§3 슬롯 빌드, 새 규칙 메시지 포함 확인).

코퍼스(워크트리 루트, `spec lint --json`):
- 수리 전: 13:48:14Z–13:54:01Z, `LINT_EXIT=0`, stderr 0 바이트, 배열 3325.
- 수리 후: 13:57:44Z–14:03:15Z, `LINT_EXIT=0`, stderr 0 바이트, 배열 3325.
- 추출은 t561 과 같은 jq 식(`select(.code=="CoverageIncomplete")` → 워크트리 접두어 제거 파일 경로 ⇥ 메시지의 REQ id, `LC_ALL=C sort`)이다. 수리 전 2010행, 수리 후 2010행.
- 계기 검증:
  - 수리 전 JSON 에서 TSV 를 다시 뽑아 기록본과 `cmp` → exit 0.
  - 수리 전 TSV 를 t561 수리 후 TSV(`.claude/worktrees/t561/.moai/reports/t561/corpus-after-coverage.tsv`)와 `cmp` → exit 0.
- 규칙별 수: `diff corpus-before-codes.tsv corpus-after-codes.tsv` → exit 0(16개 코드 전부 같음, `DuplicateAcceptanceID` 없음).
- 집합 차: `comm -23` → 0행(`corpus-disappeared.tsv`), `comm -13` → 0행(`corpus-appeared.tsv`). `cmp corpus-before-coverage.tsv corpus-after-coverage.tsv` → exit 0.

픽스처 재실행:
- `.moai/specs/` 에 잠시 복사했다. 사전 부재 `precheck_exit=1`, 실행 뒤 `rm_exit=0`, `git status --short -- .moai/specs` 출력 없음.
- 코퍼스 수리 후 측정이 끝난 뒤에 실행해 코퍼스 수에 섞이지 않았다.

| 픽스처 | `spec lint <path> --json` | `spec view <ID>` |
|---|---|---|
| A 인용 위 | exit 0, `DuplicateAcceptanceID  18  AC AC-DUPREPRO-001-01 is declared on more than one line; …` 1건, CoverageIncomplete 없음 | exit 0, stderr `Warning: duplicate acceptance criteria ID: AC-DUPREPRO-001-01 at depth 0`, stdout `└── AC-DUPREPRO-001-01.a: cited here only as a cross-reference note; the declaration follows below.: (maps REQ-DUPREPRO-001-001)` |
| B 선언 위 | exit 0, `DuplicateAcceptanceID  18  AC AC-DUPREPRO-002-01 …` 1건 | exit 0, 같은 모양의 경고, stdout `└── AC-DUPREPRO-002-01.a: Given a SPEC: (maps REQ-DUPREPRO-002-001)` |
| C 대조군 | exit 0, 발견 0건 | exit 0, stderr 0 바이트, `cmp after-C-view.out repro-C-view.out` → exit 0 |

- 원본 출력은 `after-{A,B,C}-lint.json`·`.stderr`, `after-{A,B,C}-view.out`·`.stderr` 에 있다.
- 원본 JSON 두 개(`/tmp/t623-lane9/t564-corpus-{before,after}.json`)는 저장소에 싣지 않는다.

### Baseline-attribution
워크트리 `WT-dup-ac-id`, 코드 `3f20c89df`(바이너리 빌드 트리와 같음 — §3), `.moai/specs` 는 기반 `d02b892c2` 와 같음(이 가지 커밋은 `.moai/specs` 를 건드리지 않음), 이 실행.

### Gaps
- 코퍼스가 중복 0건이라 코퍼스 전후 비교는 "수리가 아무것도 바꾸지 않는다"만 보인다. 새 동작의 관측은 픽스처와 테스트뿐이다.
- 통합 창에서 로컬 develop 을 흡수한 뒤의 재측정은 아직이다.

### Residual-risk
- A 의 `spec view` 는 인용 문구를 AC 본문으로 보여 준다(§3 잔여 위험의 실측). 작성자는 경고 줄로만 이를 알 수 있다.

## 5. 통합 창 — 흡수 · 병합 트리 재측정

### Claim
- 리드가 지명한 창에서 로컬 develop `93182d137` 을 카드 브랜치로 흡수했다(`2dd2634d6`).
- 흡수 델타는 `internal/spec`·그 의존 패키지·`internal/cli/spec_view*.go` 를 건드리지 않는다. 그래서 리드 승인 조건에 따라 코퍼스 재측정은 하지 않았다.
- 흡수한 트리 `5d0855ba0d032b0d330b17e639e136ec1b9a3c9e` 에서 다음이 통과한다:
  - `internal/spec` 패키지 전체와 새 테스트 5개
  - 창 슬롯 1회로 다시 돌린 spec view 테스트 4개(델타의 `internal/core/git` 때문에 `internal/cli` 가 다시 컴파일되는 것을 리드가 슬롯으로 승인)

### Evidence
창 획득과 흡수:
- `moai integration acquire --name lane-9` → `acquire_exit=0`, `release-integration window acquired by e5c0032b-ffbd-44b0-8997-c16d09d3541b on WT-dup-ac-id`(`window-acquire.log`).
- 흡수 대상: `git rev-parse develop` → `93182d137159c4facbf87c66dea3fd69160a6b8b`(리드 지명과 같음). `git merge-base HEAD develop` → `c3b931784…`(이 가지의 base).
- 흡수 델타: `git diff --name-only c3b931784 93182d137` → 194파일(`window-absorb-delta-files.txt`).
  - `^internal/spec/` 0건(exit 1), `^internal/cli/spec_view` 0건(exit 1).
  - 코드 경로는 `internal/cli/plan_audit_order_conflict_test.go`, `internal/core/git/manager.go`, `internal/core/git/status_optional_locks_test.go`, `internal/template/` 아래 8파일(catalog.yaml·에이전트·스킬 문서).
- 의존: `go list -deps` 로 흡수한 트리에서 `internal/spec` 의 저장소 패키지를 셌다. `config/atomicfile`·`defs`·`paths`·`pkg/models`·`config`·`constitution`·`execerr`·`spec` 8개이고, 모두 `embed=[]` 다(`window-merged-spec-deps.txt`). `internal/core/git`·`internal/template` 은 의존에 없다.
- 흡수 병합: `git merge --no-ff -m "Merge local develop 93182d137 into WT-dup-ac-id (card t564 window)" 93182d137` → `merge_exit=0`, `2dd2634d6 5cf566807 93182d137`, MERGE_HEAD 없음(exit 1)(`window-absorb-merge.log`).

사건 기록 — `index.lock`:
- 병합 직후와 14:08:35Z 에 `/Users/goos/MoAI/moai-adk-go/.git/worktrees/t564/index.lock`(0 바이트, 23:08 로컬 생성)이 있었다.
- 당시 HEAD `2dd2634d6`, MERGE_HEAD 없음, 추적 파일 변경 없음. 병합 커밋은 이미 만들어진 뒤라 실패한 명령은 없었다.
- 손으로 지우지 않았다. 14:10:51Z 에 다시 보니 경로가 없었다(exit 1). 같은 시각 `ps` 에 `git` 프로세스는 없었다(`window-git-procs.out` 빈 파일). 어느 프로세스의 락이었는지는 확인하지 못했다.

병합 트리 재측정:
- `go test ./internal/spec -count=1` → `pkg_exit=0`, `ok github.com/modu-ai/moai-adk/internal/spec 103.447s`(`window-merged-spec-package.log`).
- `go test ./internal/spec -count=1 -run '^(TestParser_DuplicateID_|TestLintDuplicateACID_)' -v` → `new_exit=0`, 5개 PASS(`window-merged-new-tests.log`).

슬롯(창 안 1회):
- 첫 사전 확인(14:10:51Z): 양성 대조 2줄 매치. 실제 `ps` 1475줄 중 1줄 매치 `32552 go go build -o /tmp/moai-t574-audit ./cmd/moai` — 다른 레인의 빌드다. 리드 지시대로 기다렸다(`window-slot-precheck-real.out`).
- 14:11:06Z `ps -p 32552` → 없음(exit 1).
- 두 번째 사전 확인(14:11:20Z): 양성 대조 2줄 매치, 실제 1476줄 중 매치 0(`window-slot-precheck2-real.out`). 매치가 없을 때만 테스트를 돌리는 조건으로 한 명령에 묶었다.
- `go test ./internal/cli -count=1 -run '^(TestSpecView_|TestSpecViewPlain_)' -v` → `cli_test_exit=0`, 4개 PASS(새 `TestSpecView_DuplicateACIDWarnsAndRendersTree` 포함), `ok … 0.984s`(`window-slot-cli-specview.log`, 14:11:20Z–14:11:30Z).

측정 트리: `git rev-parse HEAD HEAD^{tree}` → `2dd2634d679524218b155167580652ad4acb3873` / `5d0855ba0d032b0d330b17e639e136ec1b9a3c9e`. 추적 파일 변경 없음, `index.lock` 없음.

### Baseline-attribution
카드 브랜치 흡수 커밋 `2dd2634d6`(트리 `5d0855ba0…`), 이 창 안의 실행. 이 절을 담는 증거 커밋은 `.moai/reports/t564/` 만 바꾼다. develop 병합 커밋의 트리가 그 증거 커밋의 트리와 같은지는 병합 직후 확인하고, 병합 SHA 와 함께 완료 보고로 전달한다.

### Gaps
- darwin 로컬 실행이다. windows·linux 는 develop push 뒤 CI 판정에 맡긴다.
- `internal/cli` 는 spec view 테스트 4개만 돌렸다(창 규칙상 전체 금지). 델타의 `internal/core/git` 변경이 다른 `internal/cli` 테스트에 주는 영향은 이 카드에서 재지 않았다.
- `index.lock` 을 만든 프로세스는 확인하지 못했다.

### Residual-risk
- `go list -deps` 는 기본 빌드 태그 기준이다. 다른 태그에서만 import 하는 의존은 목록에 없다.
