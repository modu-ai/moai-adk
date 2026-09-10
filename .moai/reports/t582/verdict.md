# t582 — 홈-상대 선언의 `..` 이탈 판정 기록

card: t582 · class B · Tier S · lane-9
branch: WT-home-dotdot-escape · base: 로컬 develop `d3b7d438d`

## 1. develop 재현 (전제 재측정)

### Claim
`~/../../etc/x` 는 develop `d3b7d438d` 의 분류기 결정 순서에서 `home-relative` 로 분류되고, 확장 결과가 홈 밖(`/etc/x`)으로 떨어진다. 카드 전제는 낡지 않았다.

### Evidence
분류기 결정 순서(`internal/cli/doctor_codex.go:685-699` @ `d3b7d438d`: IsAbs → 백슬래시 → `~`/`~/` → 잔여 `~` → relative)와 확장(`doctor_codex.go:715` `filepath.Join(home, p[2:])`)을 표준 라이브러리만으로 옮긴 재현 프로그램 `.moai/reports/t582/repro/main.go`.

명령: `go run ./.moai/reports/t582/repro > .moai/reports/t582/repro-premise.log 2>&1; echo "exit=$?"` → `exit=0`

```
"~/ok/SKILL.md"        shape=home-relative  expanded=/Users/example/ok/SKILL.md inside_home=true
"~/../../etc/x"        shape=home-relative  expanded=/etc/x                   inside_home=false
"~/x\\SKILL.md"        shape=oddly-formed   expanded=-                        inside_home=-
"/etc/x"               shape=absolute       expanded=-                        inside_home=-
"~/a/../b/SKILL.md"    shape=home-relative  expanded=/Users/example/b/SKILL.md inside_home=true
"~/.."                 shape=home-relative  expanded=/Users                   inside_home=false
```

대조군 4종: 정상 홈-상대(홈 안) · `..` 이탈(홈 밖) · 백슬래시(t571 가드가 거절) · 절대 경로(절대). 추가 관측: `~/..` 도 이탈한다. `~/a/../b` 는 `..` 를 품지만 홈 안에 머문다.

### Baseline-attribution
워크트리 HEAD `d3b7d438d`(`git merge --ff-only d3b7d438d` 후 `git rev-parse --short HEAD` → `d3b7d438d`), 이 실행.

### Gaps
- 프로덕션 코드(`classifyCodexSkillPath`, `judgeCodexSkillEntry`) 자체는 아직 실행하지 않았다 — `internal/cli` 컴파일·테스트는 리드 슬롯 승인 대기. 위 결과는 결정 순서를 옮긴 사본의 결과이며, 프로덕션 RED 는 슬롯에서 실패하는 재현 테스트로 세운다.
- 소비자 경로 두 곳(`codex_skills_prune.go:84-91` 삭제 판정, `doctor_codex.go:859-867` 누락 집계)이 같은 분류기와 같은 확장 함수를 쓴다는 것은 코드 판독이며 실행 관측이 아니다.

### Residual-risk
재현 사본이 프로덕션과 어긋날 가능성 — 슬롯의 RED 테스트가 닫는다.

## 2. 쓰기 쪽 관측 (코드 판독)

`upsertCodexSkillDisable`(`codex_skills_disable.go:231`)의 유일한 프로덕션 호출부는 `:426`, 입력은 `resolveCodexSkillMirrorPath` 가 `filepath.Join(root, mirror, skill, "SKILL.md")` 로 만든 절대 경로이며 `skill` 은 `/`·`\`·`.`·`..` 를 거절한다(`:141`). 따라서 오늘의 쓰기 경로는 `~/`·`..` 선언을 만들지 않는다. 다만 함수 자체의 가드(`:261` `ContainsAny(skillPath, "\"\\\n\r")`)는 `..` 를 거절하지 않는다.

## 3. 설계 판정 (리드)

A안 — `..` 세그먼트를 전부 거절, 읽기·쓰기가 판별식 `hasCodexDotDotSegment` 하나를 공유. 이동하는 칸 둘은 **의도된 변화**로 테스트에 고정: `~/a/../b/SKILL.md` home-relative → oddly-formed, `rel/../x/SKILL.md` relative → oddly-formed(t571 백슬래시 검사와 같은 자리 — 모든 비절대 선언). 절대 경로 `/tmp/a/../b` 는 IsAbs 우선이라 absolute 로 남는다(대조군).

## 4. production RED — 컴파일 ① (수리 전 트리)

### Claim
새 테스트 3개가 수리 전 production 코드에서 **옳은 이유로** 실패한다. 빌드 오류가 아니다.

### Evidence
사전 상태: `git rev-parse --short HEAD` → `d3b7d438d`; `git status --short` → `?? .moai/reports/t582/`, `?? internal/cli/codex_skills_dotdot_test.go` (production 파일 수정 0).

동시 실행 조건(06:34:07Z 직전 확인): `ps -eo pid,etime,comm,args | awk '$3=="go" && /internal\/cli/'` →
```
 3767       03:08 go               go test ./internal/cli/ -count=1 -timeout 60m
```
lane-6 전체 스위트 1건만 병행, 다른 스코프 컴파일 0건.

명령(06:34:17Z–06:34:26Z): `go test ./internal/cli/ -count=1 -v -run '^(TestCodexSkillPathDotDotSegmentRefused|TestJudgeCodexSkillEntryRefusesDotDotEscapeBeforeStat|TestCodexDotDotRefusalIsSymmetricAcrossReadAndWrite)$' > .moai/reports/t582/red-before-fix.log 2>&1` → `exit=1`

로그 판독(37행): 빌드 오류 표지(`[build failed]`·`undefined:`·`# ` 헤더) 0건.
`--- FAIL` 이름별: 최상위 3개 전부 + 부분 5개(`home_escape_to_etc`, `home_escape_to_parent`, `home_inner_dotdot_intended_change`, `relative_dotdot_intended_change`, `refuses_before_stat`).
`--- PASS` 이름별(대조군): `clean_home_relative`, `dotdot_prefixed_name_is_not_a_segment`, `plain_relative`, `absolute`, `absolute_with_dotdot`, `positive_control_clean_path_does_stat_once`.
실패 메시지:
```
codex_skills_dotdot_test.go:64: "~/../../etc/x" classified codexPathHomeRelative, want codexPathOddlyFormed
codex_skills_dotdot_test.go:64: "~/.." classified codexPathHomeRelative, want codexPathOddlyFormed
codex_skills_dotdot_test.go:64: "~/a/../b/SKILL.md" classified codexPathHomeRelative, want codexPathOddlyFormed
codex_skills_dotdot_test.go:64: "rel/../x/SKILL.md" classified codexPathRelative, want codexPathOddlyFormed
codex_skills_dotdot_test.go:102: Eligible = true for a home-escaping declaration (stat saw [/etc/x]); a deletion verdict must not be reachable here
codex_skills_dotdot_test.go:143: write side: Action = 0, want codexSkillDisableSkipped
```
`stat saw [/etc/x]` 가 결함의 production 관측이다 — 홈 밖 경로가 stat 되고 삭제 판정(`Eligible=true`)에 도달했다. §1 의 stdlib 복제본 Gap 은 이것으로 닫힌다.

### Baseline-attribution
워크트리 `d3b7d438d` + 미추적 테스트 파일 1개, 이 실행.

### Gaps
- 컴파일 ② (수리 후 GREEN + t571 기존 테스트) 는 아직 실행 전.
- 정정: 리드에게 "t571 기존 4개" 라고 적었으나 `codex_skills_path_shape_test.go` 의 최상위 테스트는 5개다(`TestCodexSkillPathBackslashSymmetry`, `TestCodexSkillPathPreservedShapes`, `TestJudgeCodexSkillEntryRefusesHomeRelativeBackslash`, `TestCodexBackslashRefusalIsSymmetricAcrossReadAndWrite`, `TestJudgeCodexSkillEntryBackslashHomeStillExpands`). ② 에는 5개 모두 담는다.

### Residual-risk
없음(RED 판정 범위 한정).

## 5. GREEN — 컴파일 ② (수리 후 트리)

### Claim
수리 후 새 테스트 3개와 t571 기존 테스트 5개가 모두 통과한다. 빈 매치로 인한 공허한 통과가 아니다.

### Evidence
수리 diff: `git diff --stat` →
```
 internal/cli/codex_skills_disable.go |  6 ++++++
 internal/cli/doctor_codex.go         | 29 ++++++++++++++++++++++++++---
 2 files changed, 32 insertions(+), 3 deletions(-)
```
- 읽기: `classifyCodexSkillPath` 에서 백슬래시 검사 바로 뒤에 `hasCodexDotDotSegment(p)` → `codexPathOddlyFormed`.
- 쓰기: `upsertCodexSkillDisable` 에서 `ContainsAny` 가드 바로 뒤에 같은 판별식 → skip(Reason 에 `".."` 명시).
- 공유 판별식: `hasCodexDotDotSegment(p) = slices.Contains(strings.Split(p, "/"), "..")`.
`gofmt -l` (변경 3파일) → 출력 없음, `gofmt_exit=0`.

동시 실행 조건(06:36:07Z 직전 재확인): `ps -eo pid,etime,comm,args | awk '$3=="go" && /internal\/cli/'` →
```
 3767       05:08 go               go test ./internal/cli/ -count=1 -timeout 60m
```
lane-6 전체 스위트 1건만 병행, 다른 스코프 컴파일 0건.

명령(06:36:14Z–06:36:21Z): `go test ./internal/cli/ -count=1 -v -run '^(TestCodexSkillPathDotDotSegmentRefused|TestJudgeCodexSkillEntryRefusesDotDotEscapeBeforeStat|TestCodexDotDotRefusalIsSymmetricAcrossReadAndWrite|TestCodexSkillPathBackslashSymmetry|TestCodexSkillPathPreservedShapes|TestJudgeCodexSkillEntryRefusesHomeRelativeBackslash|TestCodexBackslashRefusalIsSymmetricAcrossReadAndWrite|TestJudgeCodexSkillEntryBackslashHomeStillExpands)$' > .moai/reports/t582/green-after-fix.log 2>&1` → `exit=0`

로그 판독(56행): 최상위 `--- PASS` 이름 8개 — 요청한 8개와 일치(새 3 + t571 5). 부분 테스트 `--- PASS` 19개. `--- FAIL`·`^FAIL`·`[build failed]`·`undefined:` 0건. 마지막 줄 `ok  	github.com/modu-ai/moai-adk/internal/cli	0.937s`.

### Baseline-attribution
워크트리 `d3b7d438d` + 위 diff(production 2파일) + 미추적 테스트 파일 1개, 이 실행.

### Gaps
- `internal/cli` 패키지 전체 판정(기지 적색 4건 부분집합)은 미실행 — 리드 순번(lane-6 · lane-5 · lane-4 뒤).
- `golangci-lint` / `go vet` 는 패키지 컴파일을 동반하므로 승인 범위 밖이라 미실행.
- `GOOS=windows` 교차 빌드 미실행. Windows 에서의 쓰기 쪽 동작(호스트 구분자 → `/` 변환 뒤 가드)은 코드 판독이며 관측이 아니다.
- doctor 누락 집계(`codexStaleSkillFinding`)의 칸 이동은 공유 분류기 테스트로만 덮였고, 그 함수를 직접 실행한 테스트는 없다.

### Residual-risk
- 절대 경로의 `..` (`/tmp/a/../b`)는 읽기에서 absolute 로 stat 되지만 쓰기에서는 거절된다. t571 백슬래시(`/tmp/a\b`)와 같은 선례이며, 운영 쓰기 경로는 `..` 를 품은 절대 경로를 만들지 않는다.
- `.` 세그먼트(`~/./x`)는 거절하지 않는다 — 홈 밖으로 나가지 못하므로 범위 밖.
- doctor 보고에서 `..` 를 품은 relative 선언이 relative 칸에서 oddly-formed 칸으로 옮겨 세어진다(의도된 변화, 테스트 고정).

## 6. `internal/cli` 패키지 전체 판정 (리드 슬롯)

### Claim
커밋 `397d04df9` 트리에서 `internal/cli` 전체 스위트의 실패 집합은 기지 적색 4건과 정확히 같다. 이 카드가 들여온 새 실패는 0건이다.

### Evidence
관측기 자기검증(시작 전, 07:09:52Z): `ps -eo pid,etime,comm,args | awk '$3=="go" && /internal\/cli/'` → 출력 없음. 같은 awk 에 ps 필드 배치를 흉내 낸 합성 행을 넣은 양성 대조군 `go test ./internal/cli/ ...` → 매치 1행, 음성 대조군(`awk internal/cli` 행) → 매치 0행. 따라서 빈 출력은 "관측기가 볼 수 없음" 이 아니라 "다른 `go` 프로세스 없음" 이다.
시작 후 자기 포착(07:10:10Z): 같은 명령 → `53928       00:10 go               go test ./internal/cli/ -count=1 -timeout 60m`. 표본 수집기(`timeout 3900` 로 바깥에서 한정, 60초 간격)의 첫 표본(07:10:12Z)도 pid 53928 을 잡았다 — 수집기의 다른 인용 경로에서도 관측기가 동작함을 따로 확인.

명령(07:10:00Z–07:25:03Z, 파이프 없음): `go test ./internal/cli/ -count=1 -timeout 60m > .moai/reports/t582/full-suite.log 2>&1; GOTEST_EXIT=$?` → `.moai/reports/t582/full-suite.exit` 에 `GOTEST_EXIT=1`.

로그(1120행) 마지막 줄: `FAIL	github.com/modu-ai/moai-adk/internal/cli	900.693s`. `panic:`·`test timed out`·`[build failed]`·`[setup failed]` 0건 — 패키지가 끝까지 돌았다.

`--- FAIL` 전체 6행:
```
--- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (3.08s)
--- FAIL: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (111.18s)
        --- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (3.05s)
        --- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (2.83s)
--- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (3.58s)
--- FAIL: TestAuditLagUsesBinlagSeam (0.73s)
```
최상위 실패 집합(4) = {`TestHomeStateChangedSurfaceCoverageConsumesFreshProfile`, `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite`, `TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition`, `TestAuditLagUsesBinlagSeam`}.
기지 적색 4건과의 차집합: 실패 − 기지 = ∅, 기지 − 실패 = ∅.
들여쓰기된 2행은 `RunsBoundedFocusedSuite` 가 띄운 하위 스위트의 출력이며, 이름이 모두 기지 적색에 속한다.

이 카드의 테스트 8개(새 3 + t571 5)의 이름: 로그 전체에서 매치 0행. 같은 awk 로 기지 적색 이름 `TestAuditLagUsesBinlagSeam` 을 찾으면 1행 — 판독기가 이름을 잡을 수 있음을 보인 양성 대조군.

동시 실행 조건(`.moai/reports/t582/full-suite-concurrency.log`, 표본 15개, 07:10:12Z–07:24:13Z): 매 표본에 pid 53928(이 실행). 그 밖의 행은 pid 81598 한 건뿐, 07:16:13Z·07:17:13Z 두 표본에만 보였다 — 인자가 `go test ./internal/cli -run ^(TestHomeState.*|...)$ -count=1 -coverpkg=... -coverprofile=.../moai-home-state-coverage-...` 인 집중 스위트. 다른 레인의 전체 스위트나 스코프 실행은 표본에 없었다.

### Baseline-attribution
워크트리 `WT-home-dotdot-escape` HEAD `397d04df9`(부모 `d3b7d438d`), 실행 전 `git status --short` 출력 없음, 이 실행.

### Gaps
- 비-verbose 실행이라 통과한 테스트는 로그에 이름이 찍히지 않는다. 8개의 "실패 아님" 은 관측됐으나 이름별 `--- PASS` 는 이 실행에서 관측되지 않았다. 이름별 PASS 는 §5 의 verbose 스코프 실행(같은 production 내용 — 커밋 직후 `git status --short` 가 비었고 커밋 파일이 §5 실행 때 작업 트리와 같다)에서 관측됐다.
- pid 81598 이 이 실행의 자식(`RunsBoundedFocusedSuite` 가 띄운 하위 스위트)이라는 판단은 인자 모양과 시각(그 테스트 소요 111.18s 구간)으로 한 추론이다. 부모 pid 는 프로세스가 끝나 확인하지 못했다.
- `golangci-lint`/`go vet`, `GOOS=windows` 교차 빌드는 여전히 미실행.
- 15분 동안 60초 간격 표본이라, 60초보다 짧게 끝난 프로세스는 표본에 안 잡혔을 수 있다.

### Residual-risk
- 기지 적색 4건이 계속 적색이라, 그 테스트들이 덮는 경로에서 이 카드가 만든 회귀가 있다면 가려진다. 그 4건은 이름상 home-state 커버리지·변경 파일 판독·binlag 계기로, `..` 분류 경로와 겹치지 않는다(이름 기준 판독이며 실행 관측은 아님).
- 경합은 거짓 통과를 만들지 않으므로(리드 판정) 초록 판정의 유효성은 유지된다.
