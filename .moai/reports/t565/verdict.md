# t565 — AC 절 헤딩 앵커가 좁으면서 헐겁다 · 판정 기록

card: t565 · class B · lane-9
branch: WT-ac-heading-anchor · base: `origin/develop` `d060e0d13` 에서 생성 → 로컬 develop `c77ef6247` 흡수(`bdbee5757`, 트리 `4e3700cd0` = develop 트리)
출처: 큐 카드 t565(t528 파생, 문안 저자 lane-4). `internal/spec/parser.go` `findACSectionStart` 는 `##` 로 시작하고(`###` 이하 포함) 소문자화 시 `acceptance` 를 포함하는 첫 헤딩을 AC 절 시작으로 잡는다.

## 1. develop 재현 · 코퍼스 실측

### Claim
- 카드의 두 방향이 develop 트리에서 모두 재현된다.
  - **헐거움**: 파일명 `acceptance.md` 를 언급한 `###` 소제목이 진짜 AC 절보다 앞에 있으면 그 소제목에 앵커한다. 진짜 AC 절은 읽히지 않는다.
  - **좁음**: `## 6. 수락 기준` 처럼 영어 `acceptance` 가 없는 헤딩은 AC 절로 인식하지 못한다.
- 코퍼스 실측에서 **세 번째 축**이 드러났다. `extractACLines` 는 `##` 로 시작하는 줄에서 멈추는데 `###` 도 여기에 걸린다. 그래서 진짜 AC 절 안의 `###` 하위 제목 아래 선언이 잘린다.
- 카드가 [HARD] 로 요구한 오차: 잘못 앵커된 절(`###` 이하 9개)에서 세어진 선언은 **0건**이다. 과대 계수는 없고 오차는 **과소 계수 한 방향**이다.
- t564 의 중복 합집합 수리와는 겹치지 않는다(`findACSectionStart` 는 t564 에서 바뀌지 않았다).
- 수리 방향에는 설계 판단이 들어가므로 §2 에서 멈춘다.

### Evidence
기반 정렬:
- 새 워크트리 → `git rev-parse HEAD origin/develop` 둘 다 `d060e0d13…` → `git status --short` 출력 없음 → `git branch -m WT-ac-heading-anchor`.
- `git merge --no-ff -m "Merge local develop c77ef6247 as t565 base (card t565)" c77ef6247`:
  - 1회차 `merge_exit=128`, 출력 전체 `fatal: Unable to write index.`(`base-merge.log`). 직후 14:14:55Z: HEAD `d060e0d13` 불변, MERGE_HEAD 없음, `index.lock` 경로 없음, 추적 파일 변경 없음.
  - 2회차 같은 실패(`base-merge-retry2.log`).
  - 원인 조사(읽기 전용): `.git/worktrees/t565/index` 권한 `-rw-r--r-- goos staff` 1539803 바이트, 디렉터리·파일 `test -w` exit 0, 디스크 여유 1.2Ti, ACL 없음, 확장 속성 `com.apple.provenance` 뿐, index 관련 git 설정은 `core.precomposeunicode true` 뿐.
  - `GIT_TRACE=1 git update-index --refresh` → exit 0(단, index mtime 불변 — 실제 쓰기 여부는 이 결과로 확정할 수 없다)(`index-write-probe.log`).
  - 3회차 `GIT_TRACE=1 GIT_TRACE_SETUP=1 git merge …` → `merge_exit=0`, `bdbee5757 d060e0d13 c77ef6247`(`base-merge-retry3-trace.log`). 락은 한 번도 손으로 지우지 않았다. 앞선 두 번의 원인은 확인하지 못했다.
- `git rev-parse --short HEAD^2` → `c77ef6247`. `HEAD^{tree}` = `c77ef6247^{tree}` = `4e3700cd0ad2c9f4dcfc6813c988bc0852ef3bd3`.

t564 와의 겹침: `git diff c3b931784 c77ef6247 -- internal/spec/parser.go` 에서 `findACSectionStart` 는 설명 주석 문맥 줄에 1번 나올 뿐, 바뀐 줄(`^[+-]`)에는 없다(exit 1). 대조로 바뀐 줄은 29개다.

코드 위치(흡수 트리 `bdbee5757`):
- `internal/spec/parser.go:66-74` `findACSectionStart` — `strings.HasPrefix(trimmed, "##") && strings.Contains(strings.ToLower(trimmed), "acceptance")` 인 첫 줄의 다음 줄을 돌려준다.
- `internal/spec/parser.go:77-90` `extractACLines` — `strings.HasPrefix(trimmed, "##")` 인 줄에서 `break`.
- 앵커를 직접 단언하는 파서 테스트는 없다(`grep -rn 'findACSectionStart\|section not found' internal/spec internal/cli --include='*_test.go'` → t528 탐침과 주석뿐).
- 산문 가드 근거: `internal/spec/lint_coverage_sibling.go:28-40` — spec.md 는 산문이 섞인 문서라 헤딩 스코핑과 `AC-…:` 문법이 산문을 AC 로 읽지 않게 막는다.

픽스처 재현(실행 전 예측 `repro.predicted`, 바이너리 `/tmp/t623-lane9/moai-t564-fixed` — 이 트리의 `internal/spec` 은 빌드 트리 `c77ef6247` 과 같다):
- `.moai/specs/` 에 잠시 복사했다. 사전 부재 `precheck_exit=1`, 실행 뒤 `rm_exit=0`, `git status --short -- .moai/specs` 출력 없음.

| 픽스처 | 구성 | `spec lint <path> --json` | `spec view <ID>` |
|---|---|---|---|
| L `SPEC-HDGREPRO-001`(헐거움) | `### Out of Scope — sibling acceptance.md inline parsing` 이 `## 6. Acceptance Criteria` 앞 | exit 0, `CoverageIncomplete  REQ REQ-HDGREPRO-001-001 is not referenced by any AC` | exit 0, `No acceptance criteria found in SPEC-HDGREPRO-001` |
| N `SPEC-HDGREPRO-002`(좁음) | AC 절 헤딩이 `## 6. 수락 기준` | exit 0, `CoverageIncomplete  REQ REQ-HDGREPRO-002-001 …` | exit 1, `Parse error: acceptance criteria section not found.` |
| C `SPEC-HDGREPRO-003`(대조군) | `## 6. Acceptance Criteria` 만 | exit 0, 발견 0건 | exit 0, `└── AC-HDGREPRO-003-01` 트리 |

세 결과 모두 예측과 같다. 원본은 `repro-{L,N,C}-lint.json`·`.stderr`, `repro-{L,N,C}-view.out`·`.stderr`.

코퍼스 실측 — 탐침 `probe/zz_t565_heading_census_test.go`:
- 실행할 때만 `internal/spec/` 에 복사하고 지웠다. 두 번 모두 사전 부재 exit 1, `rm_exit=0`, `git status --short -- internal/spec` 출력 없음.
- 명령 `go test ./internal/spec -count=1 -run '^TestT565HeadingAnchorCensus$' -v` → exit 0.
- 1회차 `probe/census-develop-before.log`, 절 밖 표본 출력을 더한 2회차 `probe/census-develop-before-samples.log`. 두 회차의 수치는 같다.
- 선언 모양 판별은 t528 판별식 B(`^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9.-]*[A-Za-z0-9])\*{0,2}\s*(.*)$`), 파서 수용은 `parseSingleACLine` 이다.
- 실행 전 기대는 `probe/census.predicted` 에 적었다(t528 수치 기준의 규모 기대).

| 앵커 종류 | 파일 | 절 안 선언 모양 | 절 안 파서 수용 |
|---|---|---|---|
| `##` + 단어 | 319 | 1071 | 994 |
| `##` + 파일명과 단어 | 29 | 109 | 101 |
| `##` + 파일명만 | 10 | 5 | 5 |
| `###` 이하 + 파일명만 | 4 | 0 | 0 |
| `###` 이하 + 단어 | 5 | 0 | 0 |
| 앵커 없음 | 467 | — | — |

- walked 834, anchored 367, unanchored 467.
- 앵커된 파일의 절 **밖**: 선언 모양 107, 파서 수용 55. 파서 수용분이 속한 헤딩 종류: 영어 acceptance 단어 22, AC 토큰 17, 기타 16.
- 앵커 없는 파일: 선언 모양 78, 파서 수용 17. 한국어 수락·인수·검수 헤딩 2, 기타 15.
- 잘못 앵커된 파일 중 뒤에 진짜 acceptance 헤딩이 있는 파일 2개, 그 아래 파서 수용 선언 8건(읽히지 않음).

표본(`probe/census-develop-before*.log`):
- `###` 이하 앵커(진짜 오탐):
  - `SPEC-AC-COLLECTOR-ANCHOR-001` → `### Out of Scope — 형제 acceptance.md 인라인 파싱`
  - `SPEC-HARNESS-OUTCOME-CAPTURE-001`·`SPEC-HARNESS-REGRESSION-GATE-001` → `### Preservation tests required GREEN (run in acceptance)`
  - `SPEC-V3R4-STATUS-LIFECYCLE-001` → `### Phase-Gate Acceptance`
  - `SPEC-STEERING-ALIGN-RULE-SCOPING-001` → `### A.1 Observed ground-truth (… acceptance.md)`
  - `SPEC-V3R6-CHANGELOG-CLEANUP-001` → `### A.3 — Fact 3: … (acceptance.md as SSOT)`
- `##` + 파일명만 앵커는 대부분 **진짜 AC 절**이다. 예: `## 3. 수락 기준 — acceptance.md (Tier M)`, `## §E — 인수 기준 요약 (상세는 acceptance.md)`, `## §F. AC Matrix (summary — full detail in acceptance.md)`, `## §H. 성공 기준 (요약; 상세는 acceptance.md)`, `## §D 수용 기준 — acceptance.md가 소유한다`. 영어 단어가 없는데 파일명 덕분에 우연히 제대로 잡혔다. 즉 이 10건은 좁은 규칙이 파일명 언급으로 가려진 경우다.
- 뒤에 진짜 절이 있는 경우: `SPEC-STEERING-ALIGN-RULE-SCOPING-001` 은 `## E. Acceptance Criteria Reference` 아래 파서 수용 선언 8건(171~178행)을 읽지 못한다. `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002` 는 뒤의 `## §D Acceptance Index` 아래 파서 수용 0건.
- 세 번째 축(절 안 `###` 에서 끊김):
  - `SPEC-DESIGN-ATTACH-001` 은 `## Acceptance Criteria` 에 앵커하지만 선언이 `### 구조 및 스키마 AC`(126~129행)·`### 동작 AC`(133~136행) 아래에 있어 읽히지 않는다.
  - `SPEC-V3R5-WORKFLOW-LEAN-001` 은 `## 3. EARS Requirements + Acceptance Criteria (inline)` 에 앵커하지만 `### 3.1 Tier Classification`(60·61행)·`### 3.2 …`(70행)·`### 3.3 …`(80행) 아래 선언이 잘린다.
  - `SPEC-AC-COLLECTOR-ANCHOR-001` 의 `### 1.2 거절 원인 …`(275~277행) 아래 선언도 절 밖이다.

### Baseline-attribution
워크트리 `WT-ac-heading-anchor` HEAD `bdbee5757`(트리 `4e3700cd0`), 이 실행. 픽스처 재현 바이너리는 `internal/spec` 코드가 이 트리와 같다.

### Gaps
- "영어 acceptance 단어" 22건 중 표본으로 파일을 확인한 것은 `SPEC-STEERING-ALIGN-RULE-SCOPING-001` 의 8건뿐이다(버킷당 표본 8개 상한). 나머지 14건의 파일 분포는 보지 않았다.
- "AC 토큰" 17건과 "기타" 16건도 표본 8개씩만 봤다.
- 탐침의 헤딩 분류는 문자열 휴리스틱이다. 진짜 AC 절인지는 표본을 사람이 읽고 판단했다.

### Residual-risk
- 앵커나 절 끝 규칙을 넓히면, 지금은 절 밖이라 무시되는 선언 모양 줄(앵커된 파일 107 − 55 = 52줄이 파서가 받지 않는 모양, 앵커 없는 파일 78 − 17 = 61줄)과 산문 속 `AC-…:` 줄이 새로 읽힐 수 있다. `lint_coverage_sibling.go:28-40` 의 산문 가드 근거가 이 방향을 경고한다.

## 2. 멈춤 — 설계 판단이 필요한 지점 (리드 판정 대기)

세 축이 서로 얽혀 있어 한 축만 고치면 다른 축의 결과가 바뀐다.

1. **앵커 헤딩 레벨(헐거움).** `###` 이하 앵커 9개는 절 안 선언 0건이고 진짜 AC 절이 아니다.
   - (a) 앵커를 `##` 정확히(`## ` 로 시작)로 제한한다. 9개 오탐이 사라지고 `SPEC-STEERING-ALIGN-RULE-SCOPING-001` 의 8건이 읽힌다.
   - (b) 레벨은 두고 파일명 `acceptance.md` 언급만 단어에서 뺀다. `###` + 단어 5개(`Preservation tests … (run in acceptance)`, `Phase-Gate Acceptance`)는 여전히 오탐이다.
   - (c) 첫 매치 대신 가장 그럴듯한 헤딩을 고른다. 판별 기준이 새로 필요하다.
2. **앵커 어휘(좁음).** 파일명만 언급한 `##` 10개는 대부분 한국어·`AC Matrix`·`Success Criteria` 로 된 진짜 AC 절이다. (a)나 (b)로 파일명 언급을 빼면 이들이 **새로 앵커를 잃는다**.
   - (가) 어휘에 한국어(수락·인수·수용·검수·성공 기준)와 `AC`·`Acceptance Criteria` 류를 더한다. 어휘 목록 관리가 생긴다.
   - (나) 파일명 언급을 계속 앵커로 인정한다(오늘의 우연을 규칙으로 올린다).
   - (다) 어휘는 두고 좁음은 이 카드에서 다루지 않는다.
3. **절 끝(세 번째 축).** `extractACLines` 가 `###` 에서 멈춰 절 안 하위 제목 아래 선언을 잃는다.
   - (ㄱ) 절 끝을 "앵커와 같거나 더 높은 레벨의 헤딩"으로 바꾼다. `SPEC-DESIGN-ATTACH-001`·`SPEC-V3R5-WORKFLOW-LEAN-001` 이 읽히지만, 앵커 절 안의 산문 하위 절(`### 3.1 Tier Classification` 같은 요구사항 절)의 `AC-…:` 줄까지 읽는다.
   - (ㄴ) 이 카드 범위 밖으로 둔다(카드 본문에 없던 축이다).
4. **범위와 영향.** 어느 조합이든 파서 수용 선언 수가 바뀌어 코퍼스 `CoverageIncomplete`(현재 2010)와 t564 의 중복 판정 입력이 함께 움직인다. 코퍼스 전/후는 t561 방식으로 잰다. 수리는 `internal/spec` 에만 닿고 `internal/cli` 컴파일은 필요 없다.
