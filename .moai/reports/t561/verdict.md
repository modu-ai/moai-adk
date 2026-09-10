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
