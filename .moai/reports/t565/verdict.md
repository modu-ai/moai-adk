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

## 3. 리드 판정 적용 — 구현·검증

리드 판정: 앵커 (b) 변형, 어휘 (가), 절 끝 (ㄱ). 원칙은 "진짜 AC 절을 더 잘 찾되 새 오탐을 만들지 않는다".

### Claim
- `internal/spec/parser.go` 에 판정 세 가지를 반영했다.
  - **앵커**: 레벨 2 이상 ATX 헤딩 가운데 첫 번째로 다음을 모두 만족하는 헤딩. 소문자로 바꾸고 `acceptance.md` 를 지운 텍스트가 ① 부정 표지(`out of scope`·`out-of-scope`·`non-goal`)를 포함하지 않고 ② 어휘 목록(`acceptance`·`success criteria`·`ac matrix`·`수락 기준`·`인수 기준`·`검수 기준`) 중 하나를 포함한다.
  - **어휘**: 리드 판정문에 이름이 나온 것만 넣었다. `수용 기준`·`성공 기준`·`AC summary` 는 넣지 않았고, 넣었을 때의 결과는 반사실로만 쟀다(아래).
  - **절 끝**: 앵커와 같거나 높은 레벨의 헤딩. `#` 헤딩도 절 끝이 된다(이전에는 `##` 로 시작하는 줄만).
  - 헤딩 판정은 `#` 1~6개 뒤에 공백이나 줄 끝이 오는 줄이다. 코퍼스에서 `##` 뒤에 공백이 없는 헤딩은 0건이라 이 조건으로 달라지는 파일은 없다.
- 테스트 16개 중 12개가 수정 전 코드에서 예측대로 실패했고, 수정 후 모두 통과한다. 뮤턴트 6개는 모두 예측한 테스트에서 잡혔다.
- 코퍼스 헤딩 센서스 결과, 판정 범위 안에서 **새 과소 계수 2건**이 생긴다. §5 에서 멈춘다.

### Evidence
RED → GREEN:
- 기준선 커밋(수정 전 측정과 RED 테스트만): `test(t565): RED heading-anchor tests and before-fix corpus baseline`. 수정 코드는 이 커밋에 없다.
- 실행 전 예측 `red.predicted`. `go test ./internal/spec -count=1 -run '^(TestT565|TestT528SectionScopingInvariant$)' -v` → exit 1, 실패 12·통과 4 모두 예측과 같다(`red-before-fix.log`).
- 수정 후 `go test ./internal/spec -count=1 -run '^(TestT565|TestT528)' -v` → exit 0, `--- PASS` 64줄, FAIL 0(`green-after-fix.log`). `gofmt -l` 세 파일 → 출력 없음, exit 0.
- 패키지 전체 `go test ./internal/spec -count=1 -timeout 1500s` → `ok  github.com/modu-ai/moai-adk/internal/spec 158.576s`, `exit=0`, `--- FAIL`·`FAIL` 줄 0(`green-spec-package.log`).
- 테스트 파일 `internal/spec/parser_ac_heading_anchor_test.go`: 파일명만 언급한 헤딩, Out of Scope 헤딩(영어·한국어), 어휘 8개(대조 2개 포함), 어휘 밖 헤딩(`## 7. Review Criteria`) 비앵커, 자기 하위 제목 판독, `###` 앵커의 같은 레벨 종료, `#` 헤딩의 `##` 절 종료.

뮤턴트(실행 전 예측 `mutants.predicted`). 작업 트리 `parser.go` 는 건드리지 않고 스크래치 사본을 `go test -overlay` 로 끼웠다. 각 사본의 차이는 `mutant-M{1..6}.diff`(M3 는 5줄 삭제, 나머지는 1줄 교체), 실행 기록은 `mutant-M{1..6}.log`.

| 뮤턴트 | 바꾼 것 | 실패한 테스트 | 예측과 |
|---|---|---|---|
| M1 | 부정 표지 검사 무력화 | OutOfScope/english, /korean | 같음 |
| M2 | `acceptance.md` 제거 생략 | FileMentionOnly | 같음 |
| M3 | 어휘를 `acceptance` 하나로 | Vocabulary 7개(한국어 3, AC Matrix, Success Criteria 2, `수락 기준 — acceptance.md`) | 같음 |
| M4 | 절 끝을 옛 규칙(`##` 접두)으로 | ReadsOwnSubheadings, h3-anchor, h1-ends-h2 | 같음 |
| M5 | 절 끝을 "더 높은 레벨"만으로 | ReadsOwnSubheadings, h3-anchor, T528SectionScopingInvariant | 같음 |
| M6 | 파일명 언급 헤딩을 전부 제외 | Vocabulary/`## 3. 수락 기준 — acceptance.md (Tier M)` | 같음 |

코퍼스 헤딩 센서스 — 탐침 `probe/zz_t565_anchor_after_census_test.go`(도우미는 `probe/zz_t565_heading_census_test.go`). 두 파일을 `-overlay` 로 끼워 `go test ./internal/spec -count=1 -run '^TestT565AnchorAfterCensus$' -v` → exit 0(`probe/anchor-after-census.log`).
- 대조: 옛 규칙을 다시 구현한 열이 §1 의 앵커 종류별 파일 수(319·29·10·4·5·467)를 그대로 재현한다. 절 안 줄 수는 1147 로 §1 의 1100 과 47 차이인데, §1 은 불릿 선언 모양만 셌고 여기서는 파서가 받는 줄을 모두 센다. 불릿 모양이 아닌 파서 수용 줄을 따로 세면 정확히 47 이다.
- 앵커된 파일 367 → 425, 절 안 파서 수용 줄 1147 → 1223(+76). 앵커나 판독 줄 수가 바뀐 파일 90.
- 늘어난 파일 12개(+88줄): `SPEC-DESIGN-ATTACH-001` +17, `SPEC-V3R5-STATUSLINE-STDINFIELDS-001` +11, `SPEC-V3R5-WORKFLOW-LEAN-001` +11, `SPEC-STEERING-ALIGN-RULE-SCOPING-001` +8, `SPEC-ERA-H3-NARROWING-001` +8, `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001` +7, `SPEC-DRIFT-CLOSE-BODY-001` +7, `SPEC-CI-PR-TRIGGER-FILTER-001` +6, `SPEC-MCP-WORKTREE-ROOT-001` +5, `SPEC-CI-DOCTOR-BIN-001` +4, `SPEC-HANDOFF-CTXGUIDE-001` +2, `SPEC-HANDOFF-MSGMODE-001` +2.
- **줄어든 파일 2개(−12줄)** — 원문 대조:
  - `SPEC-LEARN-CHANNEL-SCOPE-001` −7. 115행 `## §F. Success Criteria`(선언 0)가 154행 `## §I. Acceptance Criteria (Tier S inline — Given-When-Then)`(171~177행 선언 7)보다 앞에 있어 앵커를 가져간다.
  - `SPEC-SYNC-AUDIT-FALSIFICATION-001` −5. 136행 `## §H AC summary (full GWT in acceptance.md)` 아래 138~142행에 선언 5(`AC-SAF-001`~`005`)가 있다. 예전에는 파일명 덕분에 앵커됐는데, 이제 파일명을 지우면 어휘가 남지 않는다.
- 파일명만 언급한 `##` 앵커 10개 중 (가) 어휘로 잡히는 것 7개, **안 잡히는 것 3개**:
  - `SPEC-SPEC-LINT-ID-ARG-001` `## §D 수용 기준 — acceptance.md가 소유한다`(절 안 0줄)
  - `SPEC-SYNC-AUDIT-FALSIFICATION-001` `## §H AC summary (full GWT in acceptance.md)`(5줄 — 위 회귀)
  - `SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` `## §H. 성공 기준 (요약; 상세는 acceptance.md)`(0줄)
- 남은 `###` 이하 앵커 **6개**, 모두 절 안 0줄:
  - 옛 `###` + 단어 5개가 그대로 남았다: `SPEC-HARNESS-OUTCOME-CAPTURE-001`·`SPEC-HARNESS-REGRESSION-GATE-001` `### Preservation tests required GREEN (run in acceptance)`, `SPEC-V3R4-STATUS-LIFECYCLE-001` `### Phase-Gate Acceptance`, `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002` `### §B.5 Reproduction-First Acceptance …`, `SPEC-ZONE-REGISTRY-HARDEN-001` `### 1.3 F3 — plan.md 문서 의미론 vs 구현·acceptance 의미론 불일치`.
  - 새로 1개: `SPEC-COVERAGE-RULE-SCOPE-001` `### 2026-08-31 — 설계 결정이 인수 기준 하나를 만료시켰다 (AC-CRS-001-006b)`(앵커 없던 파일).
  - 옛 `###` + 파일명만 4개는 모두 사라졌다. `SPEC-STEERING-ALIGN-RULE-SCOPING-001` 은 `## E. Acceptance Criteria Reference` 로 옮겨 8줄을 읽는다.
- 부정 표지로 빠진 헤딩 중 어휘를 가진 것 2개: `SPEC-AC-COUNT-DISCRIMINATOR-001`·`SPEC-SELECTOR-CENSUS-001` 의 `### Out of Scope — … 수락 기준 …`.
- 앵커 절 뒤에 AC 를 부르는 헤딩이 또 있고 그 아래 파서 수용 줄이 있는 경우: 3건 14줄. `SPEC-LEARN-CHANNEL-SCOPE-001` 7줄(위 회귀), `SPEC-V3R2-SPC-001` `### 11.1 Before (v2.x acceptance format)` 3줄·`### 11.2 After (v3.0+ acceptance format)` 4줄(수정 전에도 읽히지 않던 형식 예시 절).
- 반사실 — 어휘에 `수용 기준`·`성공 기준`·`ac summary` 를 더하면: 판정 대비 21파일이 바뀌고 +18줄. 줄이 생기는 파일은 `SPEC-PREMERGE-SETTINGS-DRIFT-001` +13(84행 `## §3 수용 기준`, 88~100행 `maps REQ-PSD-…` 선언 13)과 `SPEC-SYNC-AUDIT-FALSIFICATION-001` +5 둘뿐이고, 나머지 19파일은 0줄이다.
- 반사실 — 앵커 선택 방식(어휘는 판정 그대로):
  - 절 안 파서 수용 줄이 0 인 앵커를 건너뛰고 다음 AC 헤딩으로 가면: 1230줄(+7), 바뀌는 파일은 `SPEC-LEARN-CHANNEL-SCOPE-001` 하나(0 → 7).
  - AC 헤딩 절을 모두 합치면(줄 중복 제거): 1237줄(+14), 바뀌는 파일 2개. `SPEC-LEARN-CHANNEL-SCOPE-001` 0 → 7, `SPEC-V3R2-SPC-001` 17 → 24(형식 예시 절 `### 11.1 Before`·`### 11.2 After` 의 7줄이 새로 읽힌다).
  - 빈 앵커 건너뛰기 + 어휘 확장(`수용 기준`·`성공 기준`·`ac summary`)을 함께 쓰면: 1248줄(판정 대비 +25, 수정 전 대비 +101), 바뀌는 파일 3개(`SPEC-LEARN-CHANNEL-SCOPE-001` 0 → 7, `SPEC-PREMERGE-SETTINGS-DRIFT-001` 0 → 13, `SPEC-SYNC-AUDIT-FALSIFICATION-001` 0 → 5).
  - 수정 전 파서보다 **적게** 읽는 파일: 판정 그대로 2개(위 두 회귀), 빈 앵커 건너뛰기 + 어휘 확장 0개.
- 코드 블록 안 `#` 줄: AC 어휘 헤딩 아래에서 0줄(awk). 같은 awk 로 헤딩 조건 없이 세면 258줄이라 계기가 동작한다. 그래서 절 끝 판정에 펜스 처리는 넣지 않았다.

### Baseline-attribution
워크트리 `WT-ac-heading-anchor`. 수정 전 측정은 HEAD `ce37a31b9`(수정 코드 없음), 수정 후 측정은 기준선 커밋 위에 수정 코드가 미커밋으로 얹힌 작업 트리, 이 실행.

### Gaps
- `moai spec view` 픽스처 L/N/C 재실행은 바이너리가 필요해 하지 않았다(`internal/cli` 컴파일 슬롯 미승인). 파서 동작은 위 테스트와 센서스로만 확인했다.
- 늘어난 12파일의 +88줄은 파일별 수만 봤고 줄마다 진짜 AC 선언인지는 읽지 않았다. 코퍼스 lint 전/후 비교(§4)에서 새로 생긴 발견은 전수 대조한다.

### Residual-risk
- 어휘가 헤딩 전체에서 부분 문자열로 맞기 때문에 `### 2026-08-31 — … 인수 기준 하나를 만료시켰다` 같은 이력 소제목도 앵커가 된다. 지금은 절 안 0줄이라 판독이 바뀌지 않지만, 이런 소제목 아래에 `AC-…:` 줄이 오면 읽힌다.

## 4. 코퍼스 lint 전/후 (판정 그대로의 구현)

### Claim
- 코퍼스 `CoverageIncomplete` 는 수정 전후 모두 **2010건**이고, 새로 생긴 발견 0건·사라진 발견 0건이다. `DuplicateAcceptanceID` 는 전후 모두 0건(증가 0). 발견 코드별 개수(16종)도 전후가 같다.
- 전수 원문 대조 대상(새 발견)과 무작위 표본 대상(사라진 발견)이 모두 빈 집합이다.
- 이 "변화 없음"은 계기가 수정을 못 본 결과가 아니다. 같은 Linter 로 재현 픽스처를 돌리면 수정 전 파서에서 L·N 이 `CoverageIncomplete` 1건씩, 수정 후 0건이다.

### Evidence
- 수정 전: `probe/lint-census-before.log`(exit 0, `findings = 3325  CoverageIncomplete = 2010  DuplicateAcceptanceID = 0`). 이 TSV 는 t564 바이너리 측정 `corpus-after-coverage.tsv`(2010행)와 `cmp` exit 0, 코드별 개수는 `moai-t564-fixed spec lint --json` 결과와 `diff` exit 0.
- 수정 후: 탐침 `probe/zz_t565_lint_census_test.go` 를 `-overlay` 로 끼워 `go test ./internal/spec -count=1 -run '^TestT565LintCensus$' -v -timeout 1500s` → `ok … 365.735s`, `exit=0`, `findings = 3325  CoverageIncomplete = 2010  DuplicateAcceptanceID = 0`(`probe/lint-census-after.log`).
- 비교: `comm -13` → `corpus-appeared.tsv` 0행, `comm -23` → `corpus-disappeared.tsv` 0행, `corpus-after-duplicate.tsv` 0행, `diff corpus-before-codes.tsv corpus-after-codes.tsv` exit 0.
- 양성 대조(탐침 `probe/zz_t565_repro_lint_test.go`, `BaseDir` = `.moai/reports/t565/repro`):
  - 수정 후 파서: `SPEC-HDGREPRO-001/002/003 CoverageIncomplete = 0/0/0`, exit 0(`probe/repro-lint-fixed.log`).
  - HEAD 의 수정 전 `parser.go` 를 `-overlay` 로 바꿔 끼움: `1/1/0`, exit 0(`probe/repro-lint-unfixed.log`). §1 의 바이너리 픽스처 결과(L·N 에서 `CoverageIncomplete`, C 는 0)와 같다.

### Baseline-attribution
수정 전 측정은 HEAD `ce37a31b9` 트리, 수정 후 측정과 양성 대조는 기준선 커밋 위 미커밋 수정 트리. 판정 빌드는 모두 `go test` 가 그 트리에서 새로 컴파일한 테스트 바이너리다.

- 읽는 줄이 바뀐 14파일에서 `CoverageIncomplete` 가 움직이지 않은 이유: 센서스(`probe/anchor-after-census.log`)에서 읽힌 줄이 매핑하는 REQ id 를 파일별 집합으로 비교하면 새로 매핑된 것 0, 매핑이 사라진 것 0 이다. 대조로 파일별 합계는 수정 전 580, 수정 후 580 이라 빈 집합 비교가 아니다. 늘어난 88줄과 줄어든 12줄은 모두 이미 다른 줄이 매핑한 REQ 를 다시 부르거나 REQ 매핑이 없는 줄이다.

### Gaps
- 반사실 조합(빈 앵커 건너뛰기 + 어휘 확장)의 코퍼스 lint 결과는 재지 않았다. `SPEC-PREMERGE-SETTINGS-DRIFT-001` 의 새 13줄은 `maps REQ-PSD-…` 를 달고 있어 `CoverageIncomplete` 가 줄어들 수 있다.

### Residual-risk
- 코퍼스 결과는 이 트리의 SPEC 834개에 대한 것이다. 새로 읽히는 줄에 REQ 매핑이 붙은 SPEC 이 앞으로 추가되면 `CoverageIncomplete` 가 달라질 수 있다.

## 5. 멈춤 — 판정 범위 안에서 생긴 새 과소 계수 (리드 판정 대기)

판정 원칙은 "새 오탐을 만들지 않는다"인데, 판정 그대로 구현하면 수정 전보다 적게 읽는 파일이 2개 생긴다(§3). 둘 다 판정문이 정하지 않은 지점에서 생긴다.

1. **AC 헤딩이 여럿일 때 어느 것을 앵커로 삼는가.** 판정 (b) 는 "첫 헤딩"을 유지한다. 그런데 `Success Criteria`·`AC Matrix` 가 어휘에 들어가면서, 앞에 있는 빈 요약 절이 뒤의 진짜 절에서 앵커를 가져간다(`SPEC-LEARN-CHANNEL-SCOPE-001` −7).
   - (i) 절 안에서 파서가 받는 줄이 0 인 앵커는 건너뛰고 다음 AC 헤딩으로 간다. 반사실 +7, 이 파일 하나만 바뀐다.
   - (ii) AC 헤딩 절을 모두 합친다. 반사실 +14 인데, `SPEC-V3R2-SPC-001` 의 형식 예시 절 7줄까지 새로 읽는다.
   - (iii) 첫 헤딩을 유지하고 −7 을 받아들인다.
2. **판정 어휘 밖의 헤딩 3개.** 파일명만 언급한 `##` 10개 중 `수용 기준`·`성공 기준`·`AC summary` 3개가 어휘에 없다. 그중 `SPEC-SYNC-AUDIT-FALSIFICATION-001` 은 절 안에 진짜 선언 5개가 있어 −5 가 된다. 코퍼스 헤딩 수로는 `## 수용 기준` 42개, `## 성공 기준` 33개로 `수락 기준`(12)·`인수 기준`(9)보다 많다.
   - (가′) 세 표현을 어휘에 더한다. 반사실 +18, 줄이 생기는 파일은 −5 회복과 `SPEC-PREMERGE-SETTINGS-DRIFT-001` +13(수정 전에도 읽히지 않던 `## §3 수용 기준`)뿐이고, 나머지 19파일은 앵커만 생기고 0줄이다.
   - (가) 판정 어휘를 유지하고 −5 를 받아들인다.
3. **1-(i)과 2-(가′)를 함께 쓰면** 1248줄이고 수정 전보다 적게 읽는 파일이 0개다. 레인 권고는 이 조합이다. 바뀌는 코드는 `findACSectionStart` 의 선택 루프와 어휘 목록 두 곳이다.
4. 코퍼스 lint 전/후(`CoverageIncomplete` 신규·소멸, `DuplicateAcceptanceID` 증가)는 §4 에 판정 그대로의 구현 기준으로 적는다.

## 6. 리드 2차 판정 적용 — 빈 앵커 건너뛰기 + 어휘 추가

리드 2차 판정: 레인 권고 (i)+(가′) 채택, (ii) 합집합 기각. 판정 기준은 "수정 전보다 덜 읽는 파일 0".

### Claim
- `findACSectionStart` 는 AC 절을 부르는 헤딩 가운데 **절 안에서 파서가 받는 줄이 1개 이상인 첫 헤딩**을 앵커로 삼는다. 그런 헤딩이 하나도 없으면 첫 AC 헤딩을 앵커로 삼는다(빈 절도 "절 없음" 오류를 내지 않는다).
- 어휘에 `수용 기준`·`성공 기준`·`ac summary` 를 더했다.
- 새 테스트 4개가 수정 전(1차 구현 `70b084194`)에 예측대로 실패하고 수정 후 통과한다. 뮤턴트 9개(신규 3 + 1차 6개 재적용)가 모두 예측한 테스트에서 잡혔다.
- 코퍼스에서 수정 전 파서보다 **적게 읽는 파일은 0개**다(아래 표).

### Evidence
RED → GREEN:
- 2차 기준선 커밋 `test(t565): RED round 2 for skip-empty anchor and the extra vocabulary`(수정 코드 없음, 부모 `70b084194`).
- 실행 전 예측 `red2.predicted`. `go test ./internal/spec -count=1 -run '^(TestT565|TestT528)' -v` → exit 1, 실패는 `TestT565AnchorSkipsEmptySection` 과 `TestT565AnchorVocabulary` 의 `§D 수용 기준`·`§H. 성공 기준 (요약)`·`§H AC summary (full GWT in acceptance.md)` 뿐이고 대조 `TestT565AnchorAllEmptySectionsStillAnchor` 는 통과, `--- PASS` 64줄(`red2-before-fix.log`).
- 수정 후 같은 명령 → exit 0, `--- PASS` 69줄, FAIL 0(`green2-after-fix.log`). `gofmt -l` 두 파일 → 출력 없음, exit 0.
- 패키지 전체 `go test ./internal/spec -count=1 -timeout 1500s` → `ok  github.com/modu-ai/moai-adk/internal/spec 101.203s`, `exit=0`, `--- FAIL`·`FAIL` 줄 0(`green2-spec-package.log`).

뮤턴트(실행 전 예측 `mutants2.predicted`, 차이 `mutant2-M{1..9}.diff`, 기록 `mutant2-M{1..9}.log`, 모두 `-overlay`):

| 뮤턴트 | 바꾼 것 | 실패한 테스트 | 예측과 |
|---|---|---|---|
| M7 | 빈 앵커 건너뛰기 제거(검사를 항상 참으로) | SkipsEmptySection | 같음 |
| M8 | 추가 어휘 3개 삭제 | Vocabulary 3개(수용 기준, 성공 기준, AC summary) | 같음 |
| M9 | 모두 빈 경우의 첫 헤딩 폴백 제거(`return -1`) | AllEmptySectionsStillAnchor | 같음 |
| M1 | 부정 표지 검사 무력화 | OutOfScope/english, /korean | 같음 |
| M2 | `acceptance.md` 제거 생략 | FileMentionOnly | 같음 |
| M3 | 1차 어휘 5줄 삭제 | Vocabulary 7개(1차 어휘 6 + `수락 기준 — acceptance.md`) | 같음 |
| M4 | 절 끝을 옛 규칙으로 | ReadsOwnSubheadings, h3-anchor, h1-ends-h2 | 같음 |
| M5 | 절 끝을 "더 높은 레벨"만으로 | ReadsOwnSubheadings, h3-anchor, T528SectionScopingInvariant | 같음 |
| M6 | 파일명 언급 헤딩 전부 제외 | Vocabulary 2개(`수락 기준 — acceptance.md`, `AC summary … acceptance.md`) | 같음 |

코퍼스 헤딩 센서스(`probe/anchor-after-census2.log`, exit 0). 이 실행에서 "ruled" 열은 2차 구현이다.
- 대조: 옛 규칙 앵커 종류별 파일 수와 불릿 모양 아닌 수용 줄 47 이 1차와 같다.
- 앵커된 파일 367 → 446, 절 안 파서 수용 줄 1147 → 1248(+101).
- 파일별 비교표 `per-file-lines-round2.tsv`(열: 파일, 수정 전 줄, 수정 후 줄, 수정 전 앵커 종류, 수정 전 앵커, 수정 후 앵커). 앵커나 줄 수가 바뀐 파일 104개: **늘어난 파일 13 · 같은 파일 91 · 줄어든 파일 0**, 줄 합계 +101. 센서스 자체 집계도 `files reading FEWER lines than the unfixed parser: ruled=0`.
- 1차에서 줄었던 두 파일(`SPEC-LEARN-CHANNEL-SCOPE-001`, `SPEC-SYNC-AUDIT-FALSIFICATION-001`)은 비교표에 없다(`grep` 으로 두 이름 0행). 표는 앵커 위치나 줄 수가 달라진 파일만 담으므로, 두 파일은 수정 전과 같은 앵커에서 같은 줄 수(7, 5)를 읽는다. 표 전체에서 `$3<$2` 인 행도 0개다.
- 읽힌 줄의 REQ 매핑: 파일별 합계 580 → 596. 새로 매핑된 16개는 모두 `SPEC-PREMERGE-SETTINGS-DRIFT-001`(84행 `## §3 수용 기준`), 매핑이 사라진 것 0.

### Baseline-attribution
워크트리 `WT-ac-heading-anchor`. RED 는 1차 구현 커밋 `70b084194` 트리 + 새 테스트, GREEN·뮤턴트·센서스는 2차 기준선 커밋 위 미커밋 2차 수정 트리, 이 실행.

코퍼스 lint 전/후(2차 구현):
- 탐침 `probe/zz_t565_lint_census_test.go` 를 `-overlay` 로 끼워 `go test ./internal/spec -count=1 -run '^TestT565LintCensus$' -v -timeout 1500s` → `ok … 301.856s`, `exit=0`, `findings = 3325  CoverageIncomplete = 2010  DuplicateAcceptanceID = 0`(`probe/lint-census-after2.log`).
- 수정 전 기준(`corpus-before-coverage.tsv`, 2010행)과 비교: `comm -13` → `corpus-appeared2.tsv` 0행, `comm -23` → `corpus-disappeared2.tsv` 0행, `corpus-after2-duplicate.tsv` 0행, `diff corpus-before-codes.tsv corpus-after2-codes.tsv` exit 0.
- 새로 생긴 `CoverageIncomplete`·`DuplicateAcceptanceID` 발견이 0건이므로 원문 대조 대상은 빈 집합이다.
- 양성 대조(2차 파서, `probe/repro-lint-round2.log`): `SPEC-HDGREPRO-001/002/003 CoverageIncomplete = 0/0/0`, exit 0. 수정 전 파서는 `1/1/0`(`probe/repro-lint-unfixed.log`)이라 같은 계기가 파서 변화를 본다.
- `SPEC-PREMERGE-SETTINGS-DRIFT-001` 의 새 매핑 16개가 lint 를 바꾸지 않은 이유: 수정 전에도 이 SPEC 의 `CoverageIncomplete` 는 0건이고(`grep` exit 1) 형제 `acceptance.md` 가 있다. 그 REQ 들은 이미 형제 파일로 커버돼 있었다.

### Gaps
- `moai spec view` 픽스처 재실행은 여전히 하지 않았다(`internal/cli` 컴파일 불필요 판정, spec view Gap 유지).

### Residual-risk
- 빈 앵커 건너뛰기는 앞선 AC 헤딩 절에 선언이 없다는 이유만으로 뒤의 절을 고른다. 앞 절이 산문으로만 AC 를 적은 진짜 절이고 뒤 절이 예시 절이면 예시를 읽는다. 코퍼스에서는 이 경우로 줄어든 파일이 0개다.
