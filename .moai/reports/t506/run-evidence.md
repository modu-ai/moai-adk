# t506 run-phase 증거 — SPEC-CODEX-GHOST-SKILLS-PRUNE-001

트리: `.claude/worktrees/t506` · 브랜치 `WT-codex-ghost-skills` · base `5ef63b1df`
측정 시각: 2026-09-07. 아래 모든 값은 **이 트리에서 이 실행으로** 관측한 것이다.

## 1. 라이브 설정 불변 (spec.md §D 첫 번째 Out of Scope)

run 단계의 **첫 명령 직전**과 **마지막 명령 직후**, 두 번 측정했다.

| 시점 | 명령 | 관측 |
|---|---|---|
| 첫 명령 직전 | `shasum -a 256 ~/.codex/config.toml` | `9f6e3a953880630afcfa6a40abe846b785e8fc0067e17b3a7ec1d513f71ca33a` |
| 마지막 명령 직후 | 같은 명령 | `9f6e3a953880630afcfa6a40abe846b785e8fc0067e17b3a7ec1d513f71ca33a` |

동일하다. 이 기계의 49건은 손대지 않았다.

`~/.codex/` 에 `config.toml.bak*` 3건이 있으나 **전부 이 실행 이전 것이다** — `config.toml.bak`(09-07 12:20, 이 세션의 run 시작 ~14:05 이전), `config.toml.bak-20260822-022202`, `config.toml.bak-20260901-133347`. 이 카드가 만드는 백업의 이름 형식은 `config.toml.bak-<YYYYMMDD>T<HHMMSS>Z`(UTC, `T`/`Z` 포함)이며 셋 중 어느 것도 그 형식이 아니다. 모든 테스트는 `CODEX_HOME` 을 `t.TempDir()` 로 돌려놓고 돌았다.

## 2. 뮤턴트 게이트 2건

### 뮤턴트 이전 GREEN (두 게이트 공통 기준선)

```
$ go test -timeout 900s ./internal/cli/...
ok  	github.com/modu-ai/moai-adk/internal/cli	361.086s
(17 packages ok, FAIL 0)   exit 0
```

### 뮤턴트 1 — indeterminate → missing 재분류 (AC-CGP-004, must-pass)

적용 지점: `internal/cli/codex_skills_prune.go` `judgeCodexSkillEntry` 의 stat switch `default:` 분기 — **프루너 자신의 적격 판정 분기**. `doctor_codex.go:442-450` 의 doctor 자체 indeterminate 처리는 건드리지 않았다.

```
$ go test -timeout 900s ./internal/cli/...
--- FAIL: TestPruneCodexSkillEntriesNeverPruneClasses (0.00s)
    codex_skills_prune_test.go:211: entry ".../indeterminate/SKILL.md" judged eligible; no never-prune row may be removed
    codex_skills_prune_test.go:214: entry ".../indeterminate/SKILL.md" skipped with no reason recorded
    codex_skills_prune_test.go:219: row ".../indeterminate/SKILL.md" was removed
    codex_skills_prune_test.go:223: output differs from input although nothing was eligible
FAIL	github.com/modu-ai/moai-adk/internal/cli	316.337s
exit 1
```

RED 의 판별식은 **테스트 이름**이다 — `TestPruneCodexSkillEntriesNeverPruneClasses` 가 지목됐고, 그 외 이름은 하나도 실패하지 않았다.

### 뮤턴트 2 — 파서 판정 무시 후 텍스트 재훑기 (AC-CGP-017, must-pass)

적용 지점: `pruneCodexSkillEntries` 의 실격 판정 자리 — 판정 직전에 `e.FirstUnrecognizedLine` 을 겉모양 기반 재훑기 결과로 덮어썼다.

```
$ go test -timeout 900s ./internal/cli/...
--- FAIL: TestPruneCodexSkillEntriesSwallowedRegistrationSurvives (0.00s)
    codex_skills_prune_test.go:261: the swallowing ghost was judged eligible:
      [{Entry:{Path:.../gone Enabled:0 StartLine:0 EndLine:6 FirstUnrecognizedLine:-1} Eligible:true SkipReason:}]
--- FAIL: TestPruneCodexSkillEntriesSwallowedRegistrationNarrowSurvives (0.00s)
    codex_skills_prune_test.go:300: the swallowing ghost was judged eligible:
      [{Entry:{Path:.../gone Enabled:0 StartLine:0 EndLine:5 FirstUnrecognizedLine:-1} Eligible:true SkipReason:}]
FAIL	github.com/modu-ai/moai-adk/internal/cli	377.607s
exit 1
```

출력의 `FirstUnrecognizedLine:-1` 이 이 뮤턴트가 무엇을 하는지 그대로 보여준다 — 텍스트 재훑기는 삼켜진 범위를 "깨끗하다"고 판정했고, 그 판정이 멀쩡한 등록을 삭제 대상에 넣었다. 7-d 가 파서-상태 판정과 텍스트 재훑기를 가르는 유일한 행이라는 SPEC 의 주장이 실물로 확인된 것이다.

### 뮤턴트 이후 GREEN (두 뮤턴트 모두 되돌린 뒤)

```
$ go test -count=1 -timeout 900s ./internal/cli/...
ok  	github.com/modu-ai/moai-adk/internal/cli	343.376s
(17 packages ok, FAIL 0)   exit 0
```

## 3. 7-d / 7-d' 선행 단언 — 픽스처가 실제로 발화했다

두 테스트 모두 적격 판정을 단언하기 **전에** 다음을 단언하고, 실패 시 그 자리에서 시끄럽게 죽는다("the fixture did not open a literal"):

```
ParseSkillEntries(fixture) → len == 1, entries[0].Path == <gone>
```

두 단언 모두 통과했다. 이 값은 SPEC 이 인용한 실측(`ENTRIES=1`, `path="/gone"`)과 일치하며, 이 실행에서 다시 관측된 것이다. 파서 축에서도 같은 사실을 독립적으로 못박았다 — `TestParseSkillEntriesSwallowedRegistration` 은 `StartLine=0, EndLine=6, FirstUnrecognizedLine=2` 를, 좁은 변형은 `FirstUnrecognizedLine=2` 를 각각 단언한다.

## 4. AC PASS/FAIL 매트릭스

| AC | 판정 | 검증 수단 | 관측 |
|---|---|---|---|
| AC-CGP-001 | PASS | `TestPruneCodexSkillEntriesRemovesMissingAbsolute` | ok |
| AC-CGP-002 | PASS | `TestPruneCodexSkillEntriesRemovesMissingHomeRelative` (`codexUserHomeDir` 씨앗 경유) | ok |
| AC-CGP-003 | PASS | `TestPruneCodexSkillEntriesNeverPruneClasses`(행 1~7c, 12항목 선행 단언) + `...SwallowedRegistrationSurvives`(7-d) + `...NarrowSurvives`(7-d') + `TestRunCleanCodexSkillsReportsSkippedEntries`(탈락 사유 열거) | ok |
| AC-CGP-004 | **PASS (must-pass)** | 뮤턴트 1 — §2 | GREEN→RED→GREEN, 지목 테스트 1건 |
| AC-CGP-005 | PASS | `TestPruneCodexSkillEntriesEnabledIsNotAGate` | ok |
| AC-CGP-006 | PASS | `TestRunCleanCodexSkillsDryRunWritesNothing` | ok (sha256 대신 `bytes.Equal` — 바이트 동일성은 sha256 동일성보다 강하다) |
| AC-CGP-007 | PASS | `TestPruneCodexSkillEntriesIgnoresHeaderInsideDocString` | ok |
| AC-CGP-008 | PASS | `TestRunCleanCodexSkillsBacksUpBeforeWriting` — 백업 내용 == 실행 전 설정, 보고에 경로 + sha256 | ok |
| AC-CGP-009 | PASS | `TestRunCleanCodexSkillsFailsOpen` 4 서브테스트(홈 미해석 / 설정 없음 / 읽히지 않음 / 항목 0건) | ok |
| AC-CGP-010 | PASS | `TestSinglePathShapeClassifier` | ok — **공허하지 않음을 확인**(§5) |
| AC-CGP-011 | PASS | `TestSkillsParserStaysReadOnly` | ok |
| AC-CGP-012 | PASS | `TestCleanCmdRejectsBothScopeFlags` | ok |
| AC-CGP-013 | PASS | `TestConfigLinesRoundTrip` (codexwiring, 재조립 함수 직접 호출) — 변형 a/b/c + CRLF-무종결자 + 빈 문자열 + 단독 개행 + 이중 개행 | ok |
| AC-CGP-014 | PASS | `TestPruneCodexSkillEntriesPreservesUntouchedBytes` 3 변형(개행 있음 / 없음 / CRLF) | ok |
| AC-CGP-015 | PASS | `TestCleanCmdHelpNamesCodexScope` | ok |
| AC-CGP-016 | PASS | `TestPruneCodexSkillEntriesEntryAtEOF` 2 변형 | ok |
| AC-CGP-017 | **PASS (must-pass)** | 뮤턴트 2 — §2 | GREEN→RED→GREEN, 지목 테스트 2건 |

## 5. AC-CGP-010 가드의 비공허성

이 가드는 텍스트 스캔이라 "아무것도 못 잡는데 초록"일 수 있다. 그래서 두 번째 분류기를 실물로 심어 발화를 확인하고 지웠다.

```
$ go test -run TestSinglePathShapeClassifier ./internal/cli/     # probe 심은 뒤
--- FAIL: TestSinglePathShapeClassifier (0.01s)
    codex_skills_prune_test.go:626: ../../internal/cli/zz_probe_second_classifier.go
      recombines IsAbs with a ~ prefix check: a second shape classifier
exit 1

$ rm internal/cli/zz_probe_second_classifier.go
$ go test -run TestSinglePathShapeClassifier ./internal/cli/     # 지운 뒤
ok  	github.com/modu-ai/moai-adk/internal/cli	0.928s          exit 0
```

## 6. 나머지 검증

| 축 | 명령 | 관측 |
|---|---|---|
| 두 패키지 전체 | `go test -count=1 -timeout 900s ./internal/cli/... ./internal/codexwiring/...` | exit 0 — 18 packages ok, FAIL 0 |
| vet | `go vet ./internal/cli/ ./internal/codexwiring/` | exit 0, 무출력 |
| lint | `golangci-lint run --timeout=5m ./internal/cli/... ./internal/codexwiring/...` | exit 0 — `0 issues.` |
| 커버리지 | `go test -cover ./internal/codexwiring/...` | `89.5% of statements` |
| 커버리지 | `go test -cover ./internal/cli/` | `80.7% of statements` (패키지 기존 baseline; 이 카드의 신규 파일은 아래) |
| 신규 파일 커버리지 | `go tool cover -func` 중 `codex_skills_prune.go` | `judgeCodexSkillEntry 100.0%` / `pruneCodexSkillEntries 100.0%` / `runCleanCodexSkills 86.5%` / `codexSkillDisplayPath` (측정 당시 0% → 전용 테스트 추가 후 커버됨) |
| 크로스 플랫폼 빌드 | `go build ./...` / `GOOS=windows GOARCH=amd64 go build ./...` | 둘 다 exit 0 |
| 서브에이전트 경계 | `grep -rn 'AskUserQuestion\|mcp__askuser' <touched files>` | 매치 0 (exit 1) |

## 7. 미검증 (Gaps)

- **`internal/cli` 패키지 커버리지 80.7% 는 85% 목표 미달이다.** 이 카드가 낮춘 것이 아니라 패키지의 기존 상태이며, 이 카드의 신규 파일 자체는 판정 함수 2개가 100%다. 패키지 전체를 끌어올리는 것은 이 카드의 범위 밖이다.
- **쓰기 실패 시 백업이 남는가** — plan §F 위험표의 완화책으로 적혀 있으나 AC 가 아니고, 픽스처로 확인하지 않았다. 코드상 백업이 먼저 쓰이고 그 뒤 쓰기 실패가 error 로 반환되므로 남지만, **관측하지 않았다.**
- **전 패키지 스위트(`go test ./...`)는 돌리지 않았다.** 로컬 금지 사항이며(CLAUDE.local.md §6), 전 패키지 판정은 CI 몫이다. 이 카드가 만진 두 패키지 밖의 영향은 미관측이다.
- **`moai clean --codex-skills` 를 실제 바이너리로 손으로 돌려보지 않았다.** cobra 배선은 `newCleanCmd()` 를 직접 실행하는 테스트로 검증했고 `go build` 는 통과하지만, 설치된 바이너리를 통한 end-to-end 실행은 관측하지 않았다.
- **뮤턴트는 2건뿐이다.** AC 가 요구한 두 경계(부재 판정 / 범위 내용물)는 덮었으나, 다른 실격 부류(relative, oddly-formed, 홈 미해석, path 없음, 경로 존재)에는 뮤턴트를 걸지 않았다. 그 부류들은 픽스처 통과만으로 서 있다.
- **GEARS 모달리티 축은 사람이 읽어 판정했다.** `moai spec lint` 의 초록은 이 축에 대해 아무 말도 하지 않는다(`isModalityMalformed` 가 영문 접두사에만 반응) — acceptance.md §D.3 마지막 항목이 요구한 명시다.

## 8. 잔여 위험

- **AC-CGP-010 가드는 모양 검사다.** `filepath.IsAbs` + `HasPrefix(p, "~` 조합을 찾으므로, 파라미터 이름이 `p` 가 아닌 두 번째 분류기(`HasPrefix(s, "~`)는 빠져나간다. §5 가 보인 것은 "이 모양의 분류기는 잡힌다"이지 "모든 분류기가 잡힌다"가 아니다.
- **백업 파일명 형식이 기존 `~/.codex/` 의 것과 다르다.** 이 카드는 `config.toml.bak-<UTC>T<...>Z`, 기존 3건은 `config.toml.bak-<YYYYMMDD>-<HHMMSS>`(로컬시). 생산자가 다르며 통일은 범위 밖이지만, 사용자가 두 형식을 보게 된다.
- **`osStatFn` 씨앗을 다른 SPEC(REQ-UGE-001)과 공유한다.** 씨앗을 재할당하는 테스트가 병렬로 돌면 데이터 레이스다. 이 카드의 테스트는 `t.Parallel()` 을 부르지 않고 `t.Cleanup` 으로 복원하지만, 규율이지 기계적 보장이 아니다. 주석에 소비자가 하나가 아님을 적어 두었다.
- **파서의 다섯 갈래 분류가 넓어지면 판정이 바뀐다.** 예컨대 `enabled` 매처가 새 형태를 받아들이면 지금 실격되는 항목이 적격이 된다. 그 변경은 이 카드의 뮤턴트가 잡지 못한다.
