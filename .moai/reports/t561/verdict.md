# t561 — GH #1696 spec lint CoverageIncomplete 오탐 판정 기록

card: t561 · class B · lane-9
branch: WT-lint-sibling-ac · base: 로컬 develop `8203040b8` (워크트리 병합 커밋 `467c33a50`, 트리 동일)
출처: GitHub issue #1696 (OPEN) — `moai spec lint` 가 Tier M SPEC 의 `acceptance.md` AC 를 못 읽어 `CoverageIncomplete` 오탐. 사례 `SPEC-STATE-ANCHOR-001`(11 REQ / 12 AC).

## 1. develop 재현 · 원인 확정 · t528 대조

### Claim
#1696 은 develop `8203040b8` 에서 **그대로 재현된다**(`CoverageIncomplete` 11건, 이슈와 같은 수). t528 로 덮이지 않는다. 원인은 "acceptance.md 를 안 읽는다" 가 아니라 — 형제 파일은 t362(`SPEC-COVERAGE-RULE-SCOPE-001`)가 이미 읽는다 — **형제 파일에서 REQ 매핑을 뽑는 문법이 `maps REQ-…` 한 형태만 받는다**는 것이다. 이 SPEC 의 `acceptance.md` 는 표 열(`| AC-SA-001 | REQ-SA-001, 002, 007 |`)로 매핑을 적어 한 건도 수집되지 않는다. 수리 방향은 설계 판단이라 여기서 멈춘다.

### Evidence
기반 정렬: 새 워크트리(`origin/develop` `d060e0d13` 기준 생성) → `git status --short` 출력 없음 → `git branch -m WT-lint-sibling-ac` → `git merge --no-ff -m "Merge develop 8203040b8 as t561 base (card t561)" 8203040b8` → `merge_exit=0`, `git log -1 --format='%h %p'` → `467c33a50 d060e0d13 8203040b8`, `git rev-parse --short HEAD^2` → `8203040b8`. `git rev-parse HEAD^{tree}` = `git rev-parse 8203040b8^{tree}` = `b8e81ced9ed025ed08f749876f40c3522c4555af`.

재현 바이너리: 09:19:25Z 실행 파일 이름 비교기(양성 대조 합성 행 2개 매치·`awk` 행 제외, 실측 출력 없음) → `go build -o /tmp/t623-lane9/moai-t561 ./cmd/moai` → `build_exit=0`.

재현(워크트리 루트): `moai-t561 spec lint SPEC-STATE-ANCHOR-001 --json > lint-develop-state-anchor.json` → `lint_exit=0`(경고는 advisory). 규칙별 집계 `CoverageIncomplete=11`. 대상 REQ: `REQ-SA-001` … `REQ-SA-011` 11개 전부. 메시지 예: `"REQ REQ-SA-001 is not referenced by any AC"`, `"severity":"warning"`, `"advisory":true`.

대조군 3종(같은 바이너리, develop 블롭으로 만든 임시 프로젝트; `controls/` 에 출력 사본):

| 사본 | 형제 `acceptance.md` | `CoverageIncomplete` | 대상 REQ |
|---|---|---|---|
| real(워크트리) | 원본(표 형식) | 11 | SA-001~011 |
| with | 원본(표 형식) | 11 | SA-001~011 |
| without | 없음 | 11 | SA-001~011 |
| maps(양성 대조) | 같은 매핑을 `(maps REQ-SA-…)` 형식으로 다시 씀 | **0** | — |

`with` 와 `without` 가 같은 11건이다 → 원본 `acceptance.md` 는 존재해도 커버리지에 0 을 기여한다. `maps` 사본이 0 이다 → 형제 파일 판독 경로 자체는 살아 있고, 막히는 곳은 매핑 문법이다. (임시 프로젝트 두 곳엔 설정 디렉터리가 없어 `OwnershipTransitionUnreachable=1` 이 추가로 떴다 — 커버리지와 무관.)

원인 코드(`8203040b8`):
- `internal/spec/lint_coverage_sibling.go` 머리주석 — t362 `SPEC-COVERAGE-RULE-SCOPE-001` M4(`9610e013e`, 2026-08-31): "CoverageRule reads the sibling itself … the covered set is taken with ExtractRequirementMappings over the file's full text".
- `internal/spec/ears.go:123` `ExtractRequirementMappings` — 1차 패턴 `(?i)maps\s+(REQ-[A-Z0-9-]+(?:\s*,\s*REQ-[A-Z0-9-]+)*)`. `maps` 키워드가 없는 표 행은 섹션으로 잡히지 않는다.
- 원본 `acceptance.md` 의 매핑 모양(13~24행): `| AC-SA-001 | REQ-SA-001, 002, 007 | … |` — 키워드 없음, 그리고 **숫자 꼬리 축약**(`002`, `007`).

t528 대조(카드의 첫 과제):
- t528 = `SPEC-AC-COLLECTOR-ANCHOR-001`. 고친 것은 `internal/spec/parser.go` 의 인라인(`spec.md`) AC 앵커(`acIDPattern`)다.
- 그 spec.md 364행: "형제 `acceptance.md` 처리는 이미 `lint_coverage_sibling.go`가 `ExtractRequirementMappings`로 해결했다. 다시 다루지 않는다." 443~445행 `### Out of Scope — 형제 acceptance.md 인라인 파싱`.
- 따라서 t528 은 이 표면을 명시적으로 범위 밖에 두었고, 그 근거("이미 해결")는 표 형식에 대해 성립하지 않는다. **#1696 은 t528 로 덮이지 않는다.**
- t362 SPEC 본문에서 표 형식 매핑 언급을 찾는 awk(`table|표|테이블|| ac-|column|열`) → 매핑 문법 관련 적중 없음.

코퍼스 규모(`8203040b8`, `git grep -l`, 파일 수):
- `acceptance.md` 전체 724
- 두 번째 열에 REQ 가 오는 표 행 `^\| *AC-[A-Z0-9-]+ *\| *REQ-` → 243 (대조: `SPEC-STATE-ANCHOR-001` 이 이 집합에 포함 = 1)
- `maps +REQ-` 형식 → 35
- 숫자 꼬리 축약 `REQ-[A-Z0-9-]*[0-9]{3} *, *[0-9]{3}` → 31

### Baseline-attribution
워크트리 `WT-lint-sibling-ac` HEAD `467c33a50`(트리 = develop `8203040b8`), 이 트리에서 빌드한 바이너리, 이 실행.

### Gaps
- 243 은 "두 번째 열" 모양만 센 하한이다. REQ 가 다른 열에 있는 표, 목록형 등 다른 모양은 세지 않았다.
- 243 개 파일이 실제로 lint 에서 몇 건의 `CoverageIncomplete` 를 만드는지(코퍼스 전체 lint)는 재지 않았다.
- 이슈 보고 바이너리(v3.2.0-rc.0)에 `9610e013e` 가 들어 있었는지는 확인하지 않았다 — 판정에는 불필요하다(develop 에서 재현됨).

### Residual-risk
- 표 형식을 받아들이면 형제 파일에서 산문 속 REQ 언급(예: 113행 `REQ-THRESHOLD-009/012`)도 커버리지로 셀 수 있다 — 넓힘의 방향과 한계가 설계 판단에 걸려 있다.

### 사건 기록 — 증거 커밋의 index.lock
- `git add .moai/reports/t561`(7개 경로) 성공 → `git commit` `commit_exit=128`: `fatal: Unable to create '/Users/goos/MoAI/moai-adk-go/.git/worktrees/t561/index.lock': File exists.` (이하 git 표준 안내 5줄).
- 직후(09:23:32Z): 락 경로 없음(`lock_ls_exit=1`), `HEAD` `467c33a50` 불변, `MERGE_HEAD` 없음(`merge_head_exit=1`). 락은 손으로 지우지 않았다. 3회 규칙 재시도 1회차.

## 2. 멈춤 — 설계 판단이 필요한 지점 (리드 판정 대기)

1. **표 형식을 어떻게 받을 것인가.** (a) `| AC-… |` 로 시작하는 행에서 모든 `REQ-…` 토큰을 매핑으로 본다 — 단순하지만 설명 열의 REQ 언급도 센다. (b) 표 머리에서 REQ 열을 찾아 그 열만 읽는다 — 정확하지만 머리 문구가 코퍼스마다 다르다. (c) 행의 두 번째 열만 읽는다 — 243 형태에 맞추지만 다른 배치는 못 본다.
2. **숫자 꼬리 축약(`REQ-SA-001, 002, 007`)을 펼칠 것인가.** 펼치면 `002`→`REQ-SA-002` 로 접두사를 추론한다. 안 펼치면 이 SPEC 에선 SA-002·SA-007 이 다른 행에서 완전형으로 등장해 결과는 같지만, 축약에만 의존하는 REQ 는 여전히 오탐이 된다(31개 파일).
3. **다른 카드와의 관계.** 이 수리는 `ears.go`/`lint_coverage_sibling.go` 쪽이고 `parser.go`(t564·t565 표면)와 겹치지 않는다. `ExtractRequirementMappings` 는 인라인 경로(`parseSingleACLine`)도 쓰므로, 함수 자체를 넓히면 spec.md 인라인 판정도 바뀐다 — 형제 경로 전용 추출기로 둘지 결정이 필요하다.

## 3. 리드 판정 · 수리 전 코퍼스 기준선 (t647 끼어들기로 중단)

### 리드 판정(원문 요지)
1. **(b) 헤더로 REQ 열 식별.** 헤더 칸이 REQ/요구/requirement(대소문자 무시)에 걸리는 열에서만 수집. 식별 못 한 표는 수집 0 — 오늘과 같은 동작. (a)·(c) 기각.
2. **축약 전개는 같은 칸 안에서만.** 앞선 완전 토큰 바로 뒤에 쉼표로 이어진 숫자 꼬리만 그 접두로 전개. 칸을 넘거나 완전 토큰이 앞에 없으면 전개 안 함. 전개/비전개 양쪽 대조군.
3. **형제 전용 추출기.** `ExtractRequirementMappings` 는 건드리지 않는다(인라인 판정 불변).
- 분류 B 유지. 한 문장: "acceptance.md 표에서 헤더로 찾은 REQ 열의 매핑(같은 칸 숫자 꼬리 축약 전개 포함)을 수집하는 형제 전용 추출기 추가".
- 추가 증거: develop 전 코퍼스 `CoverageIncomplete` 수 수리 전/후, 사라진 발견 무작위 5건 표 원문 대조. RED 먼저, 뮤턴트 2개(헤더 식별 제거·칸 경계 제거).

### 수리 전 기준선 — Evidence
명령(09:25:06Z–09:33:44Z, 워크트리 루트, develop 트리 빌드 바이너리 `/tmp/t623-lane9/moai-t561`): `spec lint --json > corpus-before.json 2> corpus-before.stderr` → `lint_exit=0`, JSON 1617919 바이트, 배열 길이 4938, 그중 `CoverageIncomplete` **3623**.
- 추적용 요약: `corpus-before-coverage.tsv`(파일 상대경로 ⇥ REQ ID, 정렬), `corpus-before-codes.tsv`(규칙별 수). 원본 JSON(1.6MB)은 저장소에 싣지 않고 `/tmp/t623-lane9/t561-corpus-before.json` 으로 옮겼다 — 수리 후 비교는 TSV 로 한다.

### Gaps
- 수리 코드는 아직 없다. t647(카탈로그 해시 누락) 끼어들기가 먼저이며, 그 병합 뒤 재개한다.
- 원본 JSON 은 추적되지 않는 `/tmp` 에만 있다 — 이 판정의 근거로 인용하는 것은 TSV 두 개뿐이다.

## 4. 수리 · RED/GREEN · 뮤턴트 4종

### Claim
수리 커밋 `54f0509af` 는 리드 판정 세 가지(헤더 식별 열만, 같은 칸 안 축약 전개만, 형제 전용 추출기)에 행 조건(같은 행에 `AC-…` id 가 있을 때만, 판정 (A))을 더해 구현한다. 새 테스트 4개는 수리 전 트리에서 빨갛고 수리 뒤 초록이다. 규칙 하나를 지운 뮤턴트 3종(B 헤더 식별 제거·C 칸 경계 제거·D 행 조건 제거)과 수리 자체를 끈 뮤턴트 A 가 모두 **실행 전에 적어 둔 테스트에서, 적어 둔 값으로** 빨개졌다. `ExtractRequirementMappings` 는 손대지 않았다.

### Evidence
수리 커밋 `git show --stat 54f0509af` → 13 files, 412 insertions, 삭제 0:
- `internal/spec/lint_coverage_sibling_table.go`(신규 130행) — `siblingTableREQIDs`·`splitTableRow`·`cellREQIDs`
- `internal/spec/lint_coverage_sibling.go`(+5) — 기존 `maps` 수집 뒤에 표 수집을 합집합으로 더함
- `internal/spec/lint_coverage_sibling_table_test.go`(신규 95행) + 픽스처 4종(`sibling-table-expand`·`-scoped`·`-no-header`·`-req-first`)

RED(`red-before-fix.log`, 수리 전 트리, 당시 테스트 3개) — `exit` FAIL:
- `HeaderColumnWithSameCellExpansion` FAIL: `uncovered REQs = [REQ-CST-001-001 REQ-CST-001-002 REQ-CST-001-003], want none`
- `OnlyHeaderColumnsAndOnlySameCell` FAIL: `uncovered REQs = [REQ-CST-002-001 REQ-CST-002-002 REQ-CST-002-003], want [REQ-CST-002-002 REQ-CST-002-003]`
- `UnidentifiedHeaderCollectsNothing` PASS — 오늘의 동작을 고정하는 대조 테스트라 수리 전에도 초록이 맞다.

GREEN(`green-after-fix.log`) — 형제 테스트 7개 전부 PASS, `ok github.com/modu-ai/moai-adk/internal/spec 2.572s`.

뮤턴트(명령은 모두 `go test ./internal/spec -count=1 -run '^TestCoverageSibling' -v > <log>`, 한 번에 하나, 실행 뒤 원복):

| 뮤턴트 | 바꾼 곳 | 예측(실행 전 기록) | 관측 | exit |
|---|---|---|---|---|
| A 수리 끔 | 합집합 루프 본문을 `_ = id` 로 | (파일 기록 없음 — Gaps) | FAIL 3: expand `[001 002 003]`, scoped `[002-001 002-002 002-003]`, req-first `[REQ-CST-004-001 REQ-CST-004-002]` / PASS 4 | 1 |
| B 헤더 식별 제거 | 헤더 조건에 `\|\| true` | FAIL scoped `[REQ-CST-002-002]`, FAIL no-header `[]` | 예측과 같음(`mutant-B-header.log`) / 나머지 PASS 5 | 1 |
| C 칸 경계 제거 | REQ 열 칸들을 `", "` 로 이어 한 칸처럼 전개 | FAIL scoped `[REQ-CST-002-003]` | 예측과 같음(`mutant-C-cell.log`) / 나머지 PASS 6 | 1 |
| D 행 조건 제거 | 행 조건에 `false &&` | FAIL req-first `[]` | 예측과 같음(`mutant-D-row.log`) / 나머지 PASS 6 | 1 |

- A 는 `RowWithoutACIsNotCoverage` 의 RED 도 겸한다: 이 테스트는 행 조건 판정 뒤에 추가돼 수리 전 트리 RED 로그에 없다. 수리를 끄면 `[REQ-CST-004-001 REQ-CST-004-002]` 로 빨개진다.
- D 적용 중 `git diff 54f0509af -- internal/` 의 변경 줄은 정확히 2줄(행 조건 한 줄의 전/후, `mutant-D-row.diff`) — 앞선 C 원복이 완전했다는 증거.
- 마지막 원복 뒤 `git diff --stat 54f0509af -- internal/` → 0 바이트(`post-mutation-diff.txt`), 같은 테스트 재실행 7개 PASS, `test_exit=0`(`green-post-mutation.log`).

### Baseline-attribution
워크트리 `WT-lint-sibling-ac` HEAD `54f0509af`, 이 실행. RED 는 수리 전 트리(`7d1f95bb1`)에서, GREEN·뮤턴트·원복 확인은 `54f0509af` 트리(뮤턴트는 그 위 단일 편집)에서 쟀다.

### Gaps
- 뮤턴트 A 의 예측은 실행 전에 파일로 남기지 않았다. B·C·D 는 `.predicted` 파일이 로그보다 먼저 쓰였다(파일 시각으로 확인 가능).
- `internal/spec` 전체 패키지 테스트는 이 절에서 돌리지 않았다 — 창에서 흡수한 트리 위에서 재측정한다.
- 인라인(`spec.md`) 판정 불변은 코드 범위(`ExtractRequirementMappings` 무변경)로만 주장하며, 코퍼스 수준에서는 §5 의 다른 규칙 15종 불변이 간접 증거다.

### Residual-risk
- 칸 분리는 `|` 기준이라 칸 안 인라인 코드의 파이프가 칸을 일찍 자른다. 이 경우 매핑을 잃을 수는 있어도 새로 만들지는 않는다(파일 머리주석에 적음).

## 5. 코퍼스 수리 전/후 · 표본 대조

### Claim
전 코퍼스 `CoverageIncomplete` 는 **3623 → 2010(−1613)** 이다. 새로 생긴 발견은 **0건**이고, 다른 규칙 15종의 건수는 한 건도 바뀌지 않았다. 사라진 발견 무작위 5건은 모두 요구사항 헤더 열·AC id 가 있는 행·완전형 REQ id 로 원문에서 확인된다. AC 없는 행은 여전히 커버리지를 만들지 않는다 — 다만 리드가 요구한 "무작위 5건이 여전히 발견으로 남는다"는 **3건만 성립**하고 2건은 사라졌으며, 그 2건은 같은 파일의 AC 우선 표가 정당하게 매핑한 경우다(아래). 이 차이는 수리의 결함이 아니라 "AC 없는 행" 표본 정의가 같은 REQ 의 다른 매핑을 배제하지 않았기 때문이다(리드 판정, 정정과 재대조는 §6).

### Evidence
측정 바이너리 확인 — 수리 후 코퍼스 lint 는 `/tmp/t623-lane9/moai-t561-fixed`(22:03:53 빌드, 수리 커밋 20초 전, 뮤턴트 A 적용 전)로 돌렸다:
- `go version -m` 의 `vcs.revision=2213871af…`, `vcs.modified=true` 는 **워크트리가 아니라 primary 체크아웃의 HEAD** 다(워크트리 HEAD 는 당시 `7d1f95bb1`). 그래서 출처 근거로 쓰지 않았다(`binary-fixed-buildinfo.txt`).
- 대신 바이너리 내용을 직접 봤다(`/usr/bin/grep -a -c`): 새 파일에만 있는 정규식 문자열 3개(헤더 `uirements?)?`, 숫자 꼬리 `^\s*,\s*([0-9]+)\b`, 행 조건 `\bAC-[A-Z0-9]+(?:-[A-Z0-9]+)*`) — 수리 후 바이너리 각 `1`(exit 0), 수리 전 바이너리 각 `0`(exit 1). 양성 대조 `CoverageIncomplete` 는 두 바이너리 모두 `1`. 세 문자열이 소스 트리에서 `lint_coverage_sibling_table.go` 한 곳에만 있음을 `/usr/bin/grep -rn --include='*.go'` 로 확인했다.

수리 후 lint: `moai-t561-fixed spec lint --json > t561-corpus-after.json` → 13:04:06Z–13:09:17Z, `LINT_EXIT=0`, stderr 0 바이트, JSON 1227235 바이트, 배열 길이 3325.

추출 계기 검증: 수리 전 TSV 를 같은 `jq` 식(`select(.code=="CoverageIncomplete")` → 워크트리 접두어 제거한 파일 경로 ⇥ 메시지의 REQ id, `LC_ALL=C sort`)으로 다시 만들어 기록본과 `cmp` → `cmp_exit=0`(바이트 동일). 같은 식으로 `corpus-after-coverage.tsv` 2010행.

규칙별 수(`corpus-before-codes.tsv` / `corpus-after-codes.tsv`): `CoverageIncomplete` 3623 → 2010, 나머지 15종(`FrontmatterInvalid` 14 … `SyncSHASlotFormat` 6) 전부 동일. 비커버리지 합계 4938−3623 = 1315 = 3325−2010.

집합 차: `LC_ALL=C comm -23 before after` → `corpus-disappeared.tsv` 1613행, `comm -13` → `corpus-appeared.tsv` 0행.

사라진 발견 무작위 5건 — `awk 'BEGIN{srand(561)} …'` 로 키를 붙여 정렬한 앞 5행(`disappeared-sample5.tsv`). 원문 행과 그 표의 헤더(구분선 바로 앞 줄):

| 발견 | 원문(acceptance.md) | 헤더 | 판독 |
|---|---|---|---|
| `SPEC-PRETOOL-GATE-MOVE-001` REQ-PGM-005 | 15행 `\| AC-PGM-006 \| Fast PreToolUse preserved — ast-grep \| REQ-PGM-005 \| MUST-PASS \| M4 \|` (16·24행도) | 8행 `\| AC ID \| Description \| Maps to REQ \| Severity \| M-plan \|` | REQ 헤더 열, AC 행, 완전형 |
| `SPEC-TEMPLATE-RULES-CLEANUP-001` REQ-TRC-060 | 80행 `\| AC-TRC-F3 \| REQ-TRC-060..063 \| … \|` | 76행 `\| AC \| 대응 REQ \| 검증 명령 \| 기대 결과 \|` | 완전형 `REQ-TRC-060` 만 수집. `..063` 범위 표기는 전개하지 않아 061·062·063 은 수리 후에도 발견으로 남는다(3건 확인) |
| `SPEC-PROFILE-MEMORY-001` REQ-PM-023 | 39행 `\| AC-PM-015 \| REQ-PM-023 \| 하드코딩 부재 \|` | 23행 `\| AC \| 대응 REQ \| 성격 \|` | REQ 헤더 열, AC 행, 완전형 |
| `SPEC-PROGRESS-MARKER-CANON-001` REQ-PMC-002 | 10행 `\| AC-PMC-002 \| REQ-PMC-002 \| MUST \| … \|` (13행도) | 7행 `\| AC ID \| REQ \| Severity \| Summary \|` | REQ 헤더 열, AC 행, 완전형 |
| `SPEC-INTEGRATION-LOCK-LIVENESS-001` REQ-INL-010 | 19행 `\| AC-INL-005 \| REQ-INL-010 \| Preserved invariant \| M3 \|` (20·26행도) | 13행 `\| AC \| Requirement \| Kind \| Flipping milestone \|` | REQ 헤더 열, AC 행, 완전형 |

AC 없는 REQ 우선 행 — 127행(`acless-rows-127.tsv`) 중 수리 전 발견이던 70건(`acless-flagged-before-70.tsv`)을 **전수** 대조했다: `comm -23 acless70 after` → **15행이 수리 후 발견에서 빠졌다**(`acless70-not-flagged-after.tsv`), 55건은 남았다. 빠진 15행은 전부 `SPEC-ZONE-REGISTRY-RESYNC-001` 의 REQ-ZRR-001~015 다(수리 후 이 SPEC 의 발견 0, 수리 전 15).
- 그 파일 197~213행 `§D.3 추적성` 표 `| REQ | AC |` 는 AC 칸에 `002, 006` 처럼 `AC-` 접두어 없는 숫자만 적는다 → 행 조건에 걸려 **수집되지 않는다**. 이 표가 "AC 없는 행" 표본에 들어간 이유다.
- 같은 파일 11~26행 `§D AC 매트릭스` 표(헤더 `| AC | 요구사항 | RED (현재 트리) | GREEN (목표) |`)가 15개 REQ 를 모두 AC 행에서 매핑한다: 015(13·25행), 001·003·005(14행 `REQ-ZRR-001, 003, 005` 같은 칸 전개), 001·003·013(15행), 002·003·009(16행), 004·006(17행), 005(18행), 007·008·011(19행), 011(20행), 008(21행), 010(22행), 013(23행), 014(24행), 012(26행). 15개가 빠짐없이 덮인다.

무작위 5건(`acless-sample5.tsv`): `comm -12 sample5 after` → 3건 남음(`REQ-WFD-008`, `REQ-WC9-007`, `REQ-WC9-014`, `acless-sample5-after.tsv`), `REQ-ZRR-005`·`REQ-ZRR-015` 는 위 이유로 빠졌다.

### Baseline-attribution
수리 전: develop 트리 빌드 바이너리 `moai-t561`, 09:25:06Z–09:33:44Z 실행(§3). 수리 후: `54f0509af` 내용이 들어간 `moai-t561-fixed`, 13:04:06Z–13:09:17Z 실행. 두 실행 모두 워크트리 루트, 같은 `.moai/specs` 트리: `git diff --stat 467c33a50 HEAD -- .moai/specs` → 0 바이트(`specs-tree-diff.txt`), 대조로 같은 명령의 `-- internal` → 751 바이트, `git status --short -- .moai/specs` 출력 없음.

### Gaps
- 수리 후 바이너리의 출처는 커밋 SHA 가 아니라 **바이너리 안의 새 문자열 3개**로 세웠다. 빌드와 커밋 사이 20초 동안 다른 편집이 없었다는 것은 그 문자열로는 증명되지 않는다 — 창에서 흡수한 트리로 다시 빌드해 재측정하면 닫힌다(리드 슬롯 승인 필요).
- AC 없는 행 무작위 5건을 뽑은 명령은 기록에서 복구하지 못했다. 표본 파일 자체는 커밋에 싣고, 대신 70건 전수 대조로 보완했다.
- 사라진 1613건 전부를 원문 대조하지는 않았다 — 무작위 5건만.
- 남은 55건이 "REQ 헤더 열에 있는 AC 없는 행"이라서 남았는지(행 조건이 막음) 아니면 헤더가 REQ 가 아니어서 남았는지는 표마다 확인하지 않았다. 행 조건의 차단은 코퍼스가 아니라 뮤턴트 D 로 세웠다.
- 원본 JSON 두 개는 `/tmp` 에만 있다. 판정 근거는 커밋한 TSV 들이다.

### Residual-risk
- `REQ-TRC-060..063` 같은 범위 표기와, `§D.3` 처럼 AC 칸에 접두어 없는 숫자만 적은 REQ 우선 표는 여전히 수집하지 않는다. 판정 규칙대로의 보수적 결과지만, 그런 표에만 매핑이 있는 SPEC 은 계속 오탐으로 남는다 — 넓힐지는 별도 판단이다.
- 헤더 정규식 `\breq(s|uirements?)?\b|요구` 는 `REQ` 가 들어간 어떤 헤더(예: `Related REQ`, `대응 REQ`)도 요구사항 열로 본다. 그 열에 매핑이 아닌 언급을 적은 표가 있으면 커버리지로 센다.

### 사건 기록 — 계측 도구
- 이 셸의 `grep`(ugrep 래퍼)은 바이너리 파일을 조용히 건너뛰어 대조군까지 출력 없이 `exit=1` 을 냈다. `/usr/bin/grep -a` 로 다시 쟀다.
- Go 의 VCS 스탬프는 워크트리(`.git` 이 파일)에서 primary 저장소의 HEAD 를 적었다. 워크트리 빌드의 출처 근거로 `vcs.revision` 을 쓰면 틀린다.

## 6. 정정 — AC 없는 행 표본의 정의 (리드 판독 반영)

### Claim
§5 의 "AC 없는 행이 수리 후에도 발견으로 남는가" 대조는 표본을 잘못 정의했다. 표본은 **"AC 없는 행에만 나오고, 같은 `acceptance.md` 의 다른 표 AC 행에서도 매핑되지 않은 REQ"** 로 정의했어야 했다. 리드는 §5 에서 빠진 15건(SPEC-ZONE-REGISTRY-RESYNC-001)을 수리 결함이 아니라 표본 정의 문제로 판정했다. 이 정의로 70건을 다시 거르면 매핑되지 않은 REQ 는 55건이고, **55건 전부가 수리 후에도 발견으로 남는다**(빠진 행 0).

### Evidence
예측(측정 전 기록, `acless70-redefined.predicted`): 다른 표에서 매핑됨 = ZRR 001~015 의 15건, 매핑 안 됨 = 55건, 그중 수리 후 발견에서 빠진 행 = 0.

판정은 린터 출력이 아니라 원문으로 했다. 70건이 걸친 `acceptance.md` 는 8개다. 각 파일에서 `|` 로 시작하고 `AC-<id>` 토큰이 있는 행만 뽑았다(`/usr/bin/grep -n -E '^[[:space:]]*[|].*AC-[A-Z0-9]'`, exit 0/0/0/0/1/1/1/0). 뽑힌 행 수는 ILA 19, SLB 13, WO 20, WFD 9, WEB-CONSOLE-007·008·009 각 0, ZRR 14다.
- 완전형 id 대조: 70개 REQ id 를 패턴 파일로 두고 `grep -H -o -w -F -f` → 14건, 전부 ZRR(001·002·004·005·007·008·010·011·012·013·014·015, 001·015 는 두 번) (`acless70-othertable-fullid.txt`).
- 쉼표 꼬리 대조: `grep -H -o -E 'REQ-[A-Z0-9]+(-[A-Z0-9]+)*-[0-9]+( *, *[0-9]+)+'` → 8건 (`acless70-othertable-tails.txt`).
  - ZRR 5건에서 003·006·009 가 더해져 15개 REQ 가 전부 덮인다.
  - ILA 3건(`REQ-ILA-001, 002` / `001, 002, 003` / `005, 009`)은 표본의 `REQ-ILA-011` 을 포함하지 않는다.
- WEB-CONSOLE 3파일 대조군: `AC-` 가 든 줄은 27·37·46줄 있지만, 행 시작 앵커를 뺀 `[|].*AC-[A-Z0-9]` 도 0·0·0 이다(exit 1). 이 파일들은 AC 를 표가 아니라 목록으로 적는다(예: 009 14행 `- **AC-WC9-001** (REQ-WC9-002, GCM-8) — …`). 따라서 앵커 때문에 빠진 것이 아니다.
- 분류: 매핑됨 15건(`acless70-mapped-elsewhere.tsv`), 매핑 안 됨 55건(`acless70-unmapped-elsewhere.tsv`). 예측과 같다.
- 수리 후 대조: `LC_ALL=C comm -23 unmapped55 corpus-after-coverage.tsv` → 0행(`acless70-unmapped-not-flagged-after.tsv`), `comm -12` → 55행. `comm_exit=0`.

### Baseline-attribution
워크트리 `WT-lint-sibling-ac` HEAD `db69e35fd`(SPEC 트리는 §5 와 같음), 수리 후 코퍼스 TSV 는 §5 의 `corpus-after-coverage.tsv`, 이 실행. 재빌드·재측정은 하지 않았다(리드 지시: 이미 가진 TSV 로).

### Gaps
- 원문 판정식은 표 헤더를 보지 않는다. AC 행이지만 요구사항이 아닌 열에 REQ 가 적힌 경우도 "매핑됨"으로 분류하므로, 매핑 안 됨 집합이 실제보다 작아질 수 있다(검사가 약해지는 방향). 이번에는 매핑됨 15건을 §5 에서 헤더 `요구사항` 열로 행 단위 확인했으므로 해당 사례는 없다.
- `..` 범위 표기와 목록형 AC 매핑(WEB-CONSOLE 3파일)은 판정식도 수리 코드도 매핑으로 보지 않는다. 두 쪽이 같은 정의를 쓰므로 이 대조의 결론은 바뀌지 않는다.

### Residual-risk
- WEB-CONSOLE-007·008·009 는 AC 를 목록 항목 `- **AC-…** (REQ-…)` 으로 매핑한다. 이 형태는 수리 범위(표 형식) 밖이라 49개 REQ 가 여전히 발견으로 남는다. 넓힐지는 별도 판단이다.
