# SPEC-INIT-QUIET-WIZARD-001 — 인수 기준

이 문서는 검증 층이다. 요구사항 원문(GEARS)은 spec.md §3 에 있고, 여기서는 각 요구가 충족됐는지 가르는 Given-When-Then 시나리오와 실행 명령만 둔다.

명령 작성 규칙:

- 영향 패키지로 한정한다. `internal/cli` 는 `-timeout 600s` 를 붙인다.
- grep 판정은 셸 래퍼를 피하려고 `git grep` 이나 `command grep` 을 쓰고, 0건 판정마다 같은 패턴이 기준 트리에서는 잡힌다는 대조군을 함께 둔다.
- 기준 비교는 두 종류이고, 종류마다 기준점이 다르다.
  - **기준 내용 추출은 핀한다.** `git show 120436f58:<파일>`, `git grep … 120436f58 -- …` 처럼 run 단계 변경 전의 내용을 읽는 명령(원문 추출·0건 판정의 대조군)은 plan 시점 워크트리 HEAD `120436f58` 에 고정한다. 흡수 뒤에도 같은 내용을 읽어야 하므로 핀이 옳다.
  - **카드 범위 diff 판정은 흡수한 `develop` 과의 merge-base 부터 잰다.** "이 카드가 이 파일을 바꾸지 않았다" 같은 판정을 리터럴 핀으로 재면, 통합 창에서 로컬 `develop` 을 흡수한 뒤의 트리에 다른 카드의 변경이 섞여 이 카드와 무관한 이유로 빨개진다(`.claude/rules/moai/development/verification-completeness.md` §2 wrong-reason red). 그래서 이런 판정은 읽는 시점에 merge-base 를 다시 구한다(`.claude/rules/local/gitflow-lane-protocol.md` §8). 워크트리 가드가 변수를 거부하므로 `CARD_BASE=$(git merge-base develop HEAD)` 대신 같은 뜻의 세 점 형식 `git diff develop...HEAD`(merge-base 에서 HEAD 까지)를 쓰고, 기준점은 `git merge-base develop HEAD` 한 줄로 따로 기록한다. 커밋 전 변경은 `git diff HEAD` 로 함께 본다. 대조군 `git diff --name-only develop...HEAD | wc -l` 이 0 이면 "변경 없음" 이 아니라 "측정 불가" 다. 이 판정은 카드가 develop 에 병합되기 전에만 쓴다 — 병합 뒤에는 merge-base 가 카드 tip 자신이 되어 범위가 비므로, 병합 뒤 근거는 병합 트리와 카드 브랜치 트리의 동일성으로 대신한다.
  - `origin/develop` 같은 원격 이동 참조는 어느 판정의 기준으로도 쓰지 않는다. 로컬 develop 이 원격보다 앞서 있으면 그 merge-base 는 흡수 전 분기점에 머물러 같은 오탐을 낸다.
- 워크트리 세션 가드는 변수·반복문·프로세스 치환이 섞인 git 명령을 거부한다. 그래서 모든 git 명령을 한 줄에 하나씩, 워크트리 루트에서 실행하는 형태로 적는다. 테스트 출력·비교용 추출물·지문 파일은 `.moai/state/verify/t583/` 에 남긴다(먼저 `mkdir -p .moai/state/verify/t583`).
- plan 시점 대조군 값은 2026-09-11 에 `120436f58` 에서 실제로 잰 값이다(각 AC 의 "plan 시점 대조군").
- **스윕 확인.** 선택자가 아무 테스트도 고르지 못하면 `go test` 는 `ok` 와 종료 코드 0 을 내고, 그 출력은 전부 통과한 출력과 구별되지 않는다(`.claude/rules/moai/development/verification-completeness.md` §1.1). 그래서 `-run` 이 한 테스트라도 지목하는 명령은 `-v` 로 돌려 출력을 파일에 받고, 지목한 최상위 테스트마다 `--- PASS: <이름> ` 줄이 있는지(`command grep -cE -- '^--- PASS: (…) '` 의 수가 지목 수와 같은지), 그리고 `[no tests to run]` 이 없는지(`command grep -c 'no tests to run'` 이 0)를 판정에 넣는다. go 의 PASS 줄은 이름 뒤에 ` (0.00s)` 가 붙으므로 `$` 로 끝을 고정하지 않는다.
- `./internal/cli` 또는 `./internal/core/project/...` 에 대한 `go test` 명령은 모두 AC-IQW-015 의 슬롯이다. 아래 각 AC 의 `go test` 줄 앞뒤에 AC-IQW-015 의 선언과 지문 채집을 둔다.

## §1 추적표

| AC | 요구 | 마일스톤 |
|---|---|---|
| AC-IQW-001 | REQ-IQW-001 | M3 |
| AC-IQW-002 | REQ-IQW-002, REQ-IQW-013 | M3 |
| AC-IQW-003 | REQ-IQW-004 | M3 |
| AC-IQW-004 | REQ-IQW-011 | M1 |
| AC-IQW-005 | REQ-IQW-012 | M1, M2 |
| AC-IQW-006 | REQ-IQW-003, REQ-IQW-009 | M2, M4 |
| AC-IQW-007a | REQ-IQW-010 | M2 |
| AC-IQW-007b | REQ-IQW-010 | M6 |
| AC-IQW-008 | REQ-IQW-015 | M2, M4 |
| AC-IQW-009 | REQ-IQW-005 | M2, M4 |
| AC-IQW-010 | REQ-IQW-006 | M4 |
| AC-IQW-011 | REQ-IQW-007 | M2, M4 |
| AC-IQW-012 | REQ-IQW-008 | M4 |
| AC-IQW-013 | REQ-IQW-013 | M4, M5 |
| AC-IQW-014 | 전 요구 (품질 게이트) | M6 |
| AC-IQW-015 | REQ-IQW-014, REQ-IQW-016 | M1~M6 (`./internal/cli`·`./internal/core/project/...` 대상 `go test` 슬롯 전부) |
| AC-IQW-016 | REQ-IQW-011, REQ-IQW-012 (spec.md §4.2 3·6항) | M1, M6 (M2 이후 새 대입 지점이 생기면 그 마일스톤 포함 — 대상은 검증 시점 스윕으로 정함) |

## §2 시나리오

### AC-IQW-001 — init 질문 집합은 4문항, 이 순서

- **Given** 임의의 프로젝트 루트
- **When** `InitQuestions(root)` 가 질문 집합을 돌려주면
- **Then** ID 목록이 정확히 `[conversation_language, user_name, agent_wiring, autonomy_tier]` 이고, 그룹 라벨이 차례로 `Basic, Basic, Quality & Workflow, Autonomy` 다. 선택자가 테스트를 실제로 골랐다.

```bash
go test ./internal/cli/wizard/... -run '^TestInitQuestions_QuietSet$' -count=1 -v > .moai/state/verify/t583/ac001.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestInitQuestions_QuietSet ' .moai/state/verify/t583/ac001.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac001.txt
```

기대: `test-exit=0`, PASS 줄 1, `no tests to run` 0.

### AC-IQW-002 — 제거 14문항이 init 에 없고, 페이지 3 전용 11문항은 흔적도 없다

- **Given** 제거 목록 14개(spec.md §2.2)
- **When** init 질문 집합, 번역 표, 답 저장 분기를 검사하면
- **Then** 14개 모두 `InitQuestions` 에 없다. 페이지 3 전용 11개는 ko/ja/zh 번역 항목이 없고, `saveAnswer`/`saveBoolAnswer` 에 넣어도 `WizardResult` 에 아무것도 저장되지 않는다. 공유 3개(`project_name`, `model_policy`, `report_format`)는 `ReconfigureQuestions` 와 번역 표에 남아 있다.

```bash
go test ./internal/cli/wizard/... -run '^(TestRemovedQuestionsAbsentFromInitSet|TestRemovedQuestionsHaveNoOrphanTranslations|TestRemovedQuestionsHaveNoCaptureBranch|TestSharedQuestionsRetainedForReconfigure)$' -count=1 -v > .moai/state/verify/t583/ac002.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: (TestRemovedQuestionsAbsentFromInitSet|TestRemovedQuestionsHaveNoOrphanTranslations|TestRemovedQuestionsHaveNoCaptureBranch|TestSharedQuestionsRetainedForReconfigure) ' .moai/state/verify/t583/ac002.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac002.txt
git grep -nE '"(project_mode|worktree_auto_create|todo_enabled|feedback_auto_submit|project_continuation|audit_model|audit_gate_claude|audit_gate_codex|audit_gate_glm|codex_audit_enabled|mcp_provision)"' -- internal/cli/wizard ':!*_test.go'; echo "exit=$?"
git grep -cE '"(project_mode|worktree_auto_create|todo_enabled|feedback_auto_submit|project_continuation|audit_model|audit_gate_claude|audit_gate_codex|audit_gate_glm|codex_audit_enabled|mcp_provision)"' 120436f58 -- internal/cli/wizard ':!*_test.go'
```

기대: `test-exit=0`, PASS 줄 4, `no tests to run` 0, 넷째 명령 `exit=1`. 다섯째 명령(대조군)은 1건 이상이어야 하며, 0 이면 넷째 판정은 무의미하다. plan 시점 대조군: `questions.go:11`, `translations.go:33`, `wizard.go:11`.

### AC-IQW-003 — reconfigure 질문 집합은 그대로다

- **Given** 이 SPEC 의 변경이 적용된 트리
- **When** reconfigure 질문 집합 고정 테스트를 돌리고, 그 테스트 본문을 기준 트리에서 뽑은 본문과 비교하고, reconfigure 진입 파일을 카드 범위 diff 로 재면
- **Then** 테스트가 통과하고, `TestReconfigureQuestionsOrder` 본문이 기준 트리와 같으며, 이 카드는 `internal/cli/update_wizard.go` 를 바꾸지 않았다 — 흡수한 로컬 `develop` 과의 merge-base 부터 잰 커밋 변경에도, 아직 커밋하지 않은 변경에도 이 파일이 없다.

```bash
go test ./internal/cli/wizard/... -run '^(TestReconfigureQuestionsOrder|TestQuestionOrder)$' -count=1 -v > .moai/state/verify/t583/ac003.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: (TestReconfigureQuestionsOrder|TestQuestionOrder) ' .moai/state/verify/t583/ac003.txt
git show 120436f58:internal/cli/wizard/questions_test.go | sed -n '/^func TestReconfigureQuestionsOrder/,/^}/p' > .moai/state/verify/t583/reconf-base.txt
sed -n '/^func TestReconfigureQuestionsOrder/,/^}/p' internal/cli/wizard/questions_test.go > .moai/state/verify/t583/reconf-head.txt
wc -l .moai/state/verify/t583/reconf-base.txt
diff .moai/state/verify/t583/reconf-base.txt .moai/state/verify/t583/reconf-head.txt; echo "diff-exit=$?"
git merge-base develop HEAD
git diff --name-only develop...HEAD | wc -l
git diff --quiet develop...HEAD -- internal/cli/update_wizard.go; echo "update_wizard-card-diff-exit=$?"
git diff --quiet HEAD -- internal/cli/update_wizard.go; echo "update_wizard-uncommitted-diff-exit=$?"
```

기대: `test-exit=0`, PASS 줄 2, `wc -l` 이 0 보다 큼(plan 시점 대조군: 52줄), `diff-exit=0`. `git merge-base` 가 찍은 SHA 를 progress.md §E.2 에 기록한다. 카드 범위 대조군(`--name-only … | wc -l`)은 1 이상이어야 하며, 0 이면 뒤의 두 종료 코드는 판정이 아니라 "측정 불가" 다. 대조군이 1 이상일 때 `update_wizard-card-diff-exit=0`, `update_wizard-uncommitted-diff-exit=0`. 이 세 줄은 카드가 develop 에 병합되기 전에만 유효하다(서두 명령 작성 규칙).

plan 시점 측정(2026-09-11, 워크트리 `120436f58`, progress.md §E.1): merge-base `93182d137159c4facbf87c66dea3fd69160a6b8b`, 카드 범위 대조군 0 — 카드 커밋이 아직 없으므로 지금은 "측정 불가" 가 맞는 상태이고, run 단계 첫 커밋이 1 이상으로 뒤집는다. 옛 리터럴 핀 형식 `git diff --quiet 120436f58 develop -- internal/cli/update_wizard.go` 은 이미 종료 코드 1 이다. 로컬 develop 에 t587 `c4990eea7`(`update_wizard.go:312-340`) 이 들어 있어, 흡수 뒤 리터럴 핀으로 재면 이 카드와 무관한 이유로 빨개진다는 실측이다. 세 점 형식이 실제로 빨개지는 방향은 알려진 입력으로 관측했다. plan-audit 2회차 D16 이 먼저 쟀고, v0.1.4 에서 같은 명령으로 다시 쟀다. `git diff --quiet 93182d137...develop -- internal/cli/update_wizard.go; echo "three-dot-known-red-exit=$?"` → `three-dot-known-red-exit=1` 이다. 로컬 develop(`81c1d58f9`)에 t587 `c4990eea7` 가 들어 있기 때문이며, `git merge-base --is-ancestor c4990eea7 develop` 종료 코드는 0 이다. 같은 범위의 `internal/cli/wizard/questions.go` → `three-dot-known-green-exit=0` 이다. 같은 형식이 바뀐 파일에서는 1, 바뀌지 않은 파일에서는 0 을 낸다는 RED/GREEN 대조이며, run 단계에서는 범위의 왼쪽 끝이 흡수한 develop 과의 merge-base 가 된다.

### AC-IQW-004 — 셸 설정 단계는 시접을 거쳐 실행되고, 그 도달이 실행으로 관측된다

- **Given** 셸 설정 단계 시접(REQ-IQW-011, 형태는 design.md §4)과 호출을 세기만 하고 실제 셸 설정 파일에 쓰지 않는 스파이
- **When** 아래 네 관측을 수행하면
  - **주 관측 (실행)**: `internal/cli` 테스트가 spec.md §4.2 홈 안전 점검표를 갖추고 시접에 스파이를 끼운 채 실제 `runInit` 을 실행한다.
  - **게이트 관측 (실행)**: `internal/core/project` 테스트가 시접에 스파이를 끼운 채 실제 `Init` 을 `SkipShellConfig=false` 와 `true` 로 각각 실행한다.
  - **보조 관측 (값 직접 읽기)**: 시접의 기본값이 운영 기본 함수와 같은 함수인지 `reflect` 함수 포인터 비교로 읽는다.
  - **본문 보존 관측 (소스 추출 비교)**: 옮기기 전 `configureShellEnv` 본문(`120436f58` 의 `internal/core/project/initializer.go:677-685`)에서 `shell.ConfigOptions{` 와 닫는 `})` 사이의 옵션 줄을 뽑고, 변경 뒤 `defaultConfigureShellEnv` 에서 같은 범위를 뽑아 비교한다(AC-IQW-010 방식).
- **Then**
  - 주 관측에서 스파이 호출이 정확히 1회다. `runInit` 이 기본 게이트(`SkipShellConfig=false`)를 셸 설정 단계까지 실제로 전달했다는 뜻이다.
  - 게이트 관측에서 `false` 는 1회, `true` 는 0회다.
  - 보조 관측에서 기본값은 운영 기본 함수다. 이 비교는 변수가 그 함수를 가리키는지만 보고, 그 함수 본문은 보지 않는다.
  - 본문 보존 관측에서 두 추출물의 `diff` 종료 코드가 0 이고, 기준 추출물이 비어 있지 않다. 이 관측이 REQ-IQW-011 의 "시접의 운영 기본값은 현재의 셸 설정 기록 동작" 절을 잰다. 재는 대상은 도달이 아니라 옮긴 옵션 값의 보존이므로, spec.md §4.1 이 금지한 "도달 증거로서의 소스 문자열 검사" 에 해당하지 않는다. 옵션 구조체 밖의 변경(로거 인자, `Configure` 호출 방식)은 이 비교가 보지 못하며 코드 리뷰가 받친다.
  - 실행 테스트 세 개 모두 선택자가 실제로 골랐다(스윕 확인).
  - 배선 제거 뮤턴트 2종이 주 관측 테스트를 RED 로 만든 출력이 progress.md §E.2 에 있다: 뮤턴트 A — `initializer.go` Step 6 게이트 반전(`if opts.SkipShellConfig`), 뮤턴트 B — Step 6 의 시접 호출 삭제. 두 경우 모두 `--- FAIL: <주 관측 테스트 이름>` 과 호출 0회를 가리키는 메시지가 나오고, 원복 뒤 다시 `--- PASS:` 가 나온다. 뮤턴트는 커밋하지 않는다.
  - 필드 누락 뮤턴트 E 가 본문 보존 관측을 RED 로 만든 출력이 progress.md §E.2 에 있다: `defaultConfigureShellEnv` 의 `ConfigOptions` 에서 `PreferLoginShell: true` 줄을 지운다. 기대는 `body-diff-exit=1` 이고 diff 출력에 `PreferLoginShell` 줄이 나오는 것이다. 이 뮤턴트에서는 스파이 관측·포인터 비교·배선 뮤턴트 A·B·AC-IQW-016 이 모두 통과하므로, 이 필드 누락을 잡는 관측은 본문 보존 관측뿐이다. 뮤턴트 E 는 소스를 추출해 비교할 뿐 `go test` 를 돌리지 않으므로 실제 홈에 닿지 않는다. 뮤턴트는 커밋하지 않는다.

```bash
go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac004-primary.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestRunInit_ShellConfigStepReachedViaSeam ' .moai/state/verify/t583/ac004-primary.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac004-primary.txt
go test ./internal/core/project/... -run '^(TestInitializer_ShellConfigSeamGate|TestConfigureShellEnvFn_DefaultIsProductionFunc)$' -count=1 -v > .moai/state/verify/t583/ac004-gate.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: (TestInitializer_ShellConfigSeamGate|TestConfigureShellEnvFn_DefaultIsProductionFunc) ' .moai/state/verify/t583/ac004-gate.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac004-gate.txt
# 본문 보존 관측
git show 120436f58:internal/core/project/initializer.go | sed -n '/^func (i \*projectInitializer) configureShellEnv()/,/^}/p' | sed -n '/shell\.ConfigOptions{/,/})/p' | sed '1d;$d' > .moai/state/verify/t583/ac004-body-base.txt
sed -n '/^func defaultConfigureShellEnv/,/^}/p' internal/core/project/initializer.go | sed -n '/shell\.ConfigOptions{/,/})/p' | sed '1d;$d' > .moai/state/verify/t583/ac004-body-head.txt
wc -l < .moai/state/verify/t583/ac004-body-base.txt
diff .moai/state/verify/t583/ac004-body-base.txt .moai/state/verify/t583/ac004-body-head.txt; echo "body-diff-exit=$?"
# 뮤턴트 A 적용 상태: FAIL 기대
go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac004-mutant-a.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- FAIL: TestRunInit_ShellConfigStepReachedViaSeam ' .moai/state/verify/t583/ac004-mutant-a.txt
# 뮤턴트 B 적용 상태: FAIL 기대
go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac004-mutant-b.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- FAIL: TestRunInit_ShellConfigStepReachedViaSeam ' .moai/state/verify/t583/ac004-mutant-b.txt
# 원복 후: PASS 기대
go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac004-revert.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestRunInit_ShellConfigStepReachedViaSeam ' .moai/state/verify/t583/ac004-revert.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac004-revert.txt
git diff --quiet -- internal/core/project/initializer.go; echo "reverted-exit=$?"
# 뮤턴트 E 적용 상태: 본문 비교 RED 기대 (go test 없음)
sed -n '/^func defaultConfigureShellEnv/,/^}/p' internal/core/project/initializer.go | sed -n '/shell\.ConfigOptions{/,/})/p' | sed '1d;$d' > .moai/state/verify/t583/ac004-body-mutant-e.txt
diff .moai/state/verify/t583/ac004-body-base.txt .moai/state/verify/t583/ac004-body-mutant-e.txt; echo "body-diff-exit=$?"
git diff --quiet -- internal/core/project/initializer.go; echo "reverted-exit=$?"
```

기대: 주 관측 `test-exit=0`·PASS 1·`no tests to run` 0. 게이트 `test-exit=0`·PASS 2·`no tests to run` 0. 본문 보존 관측은 기준 추출물 `wc -l` 이 1 이상(0 이면 추출 범위가 비어 판정 무의미)이고 `body-diff-exit=0`. 뮤턴트 A·B 파일은 각각 `test-exit` 가 0 이 아니고 FAIL 줄 1. 원복 파일은 `test-exit=0`·PASS 1·`no tests to run` 0. 뮤턴트 E 적용 상태의 비교는 `body-diff-exit=1`. `reverted-exit` 는 뮤턴트 적용 전 `initializer.go` 가 커밋된 상태였다는 전제에서 0. 테스트 이름과 `defaultConfigureShellEnv` 라는 함수 이름은 design.md 의 제안이며, run 단계에서 이름이 바뀌면 이 명령의 이름을 함께 고친다 — 그러지 않으면 스윕 확인이나 본문 추출이 먼저 실패한다. 이 AC 의 모든 `go test` 줄은 AC-IQW-015 슬롯이다.

plan 시점 대조군 (2026-09-11, 워크트리 `120436f58`):

- 추출 범위는 파일 안에 하나뿐이다: `git grep -c 'shell\.ConfigOptions{' 120436f58 -- internal/core/project/initializer.go` → `120436f58:internal/core/project/initializer.go:1`.
- 기준 추출물 `wc -l` → `4`. 내용은 `AddClaudeWarningDisable: true,`, `AddLocalBinPath: true,`, `AddGoBinPath: true,`, `PreferLoginShell: true,` 네 줄이다(정렬 공백 생략).
- RED-now(옳은 이유): 같은 명령으로 뽑은 현재 트리 추출물은 `0` 줄이고 `body-diff-exit=1` 이다. `defaultConfigureShellEnv` 가 아직 없으며, M1 이 뒤집는다.
- 뮤턴트 E 의 실패 관측: 기준 추출물에서 `PreferLoginShell` 줄만 뺀 입력과 비교하면 diff 출력은 `4d3` / `< 		PreferLoginShell:        true,` 이고 `mutant-e-diff-exit=1` 이다. 같은 파일끼리 비교한 대조는 `identity-diff-exit=0` 이다. 이 plan 시점 관측은 추출물 수준의 모의이며, 소스를 실제로 고친 뮤턴트 E 의 기록은 run 단계 §E.2 몫이다.

금지 뮤턴트: 시접을 우회해 `configureShellEnv` 가 `shell.NewEnvConfigurator` 를 직접 부르게 하는 변형은 주 관측 테스트를 RED 로 만들지만, 그 실행 중에 실제 셸 설정 파일에 쓴다. 만들지도 돌리지도 않는다(plan.md §D).

근거와 한계:

- **소스 문자열 검사를 증거로 쓰지 않는 이유**: `init.go` 나 `initializer.go` 에 특정 문자열이 있다는 것은 텍스트 패턴 추론이다. 배선을 우회하는 변형도 문자열은 남길 수 있어, 실행 전달을 증명하지 못한다(리드 결정, 2026-09-11).
- **"실제 셸 설정 파일이 바뀌지 않았다"는 간접 관측을 쓰지 않는 이유**: `internal/shell/config.go:93`, `:167` 은 설정 줄이 이미 있으면 기록을 건너뛴다. 이미 설정된 머신에서는 시접이 있든 없든 파일이 불변으로 나오므로, 그 관측은 시접과 무관하다. 게다가 기본 게이트의 전달을 실제 기록 함수로 관측하면 실제 `~/.zshenv` 에 쓰게 되어 spec.md §4.2 와 충돌한다.
- **커밋 조상 판정을 증거에서 뺀 이유**: 시접이 먼저 착지해야 한다는 순서는 plan.md §F 의 마일스톤 순서로 지키며, 이 AC 가 증명하는 대상은 순서가 아니라 도달이다.
- **잔여 위험**: 시접을 우회하는 구현 회귀는 호출 0회로 잡히지만, 그 실행이 실제 홈에 쓸 수 있다. 이것은 AC-IQW-015 의 슬롯 지문과 REQ-IQW-016 멈춤 규칙이 받친다(plan.md §J R10).

### AC-IQW-005 — 홈 안전 점검표 (새로 쓰거나 본문을 다시 쓴 init 실행 테스트)

- **Given** 이 SPEC 이 새로 쓰거나 본문을 다시 쓴 `internal/cli` init 실행 테스트 파일과 홈 안전 헬퍼. 대상 목록은 고정 목록이 아니라 검증 시점에 다시 뽑는다(아래 스윕). 제거된 필드 참조만 지운 기존 테스트(`init_agent_wizard_test.go:64-160`, `doctor_codex_e2e_test.go`, `init_workflow_wiring_test.go:100-134` 등)는 이 AC 의 대상이 아니며 AC-IQW-015 가 덮는다
- **When** 대상 목록을 스윕하고, 소스를 검사하고, 가드 자체 테스트를 돌리면
- **Then** 대상 목록은 이 카드가 추가한 `internal/cli/*_test.go`(커밋분과 미커밋분) 가운데 init 을 실행하는 호출(`runInit`, `runInit` 으로 시작하는 헬퍼, `prepareSafeInitHome`)을 담은 파일에 계획한 세 파일(`init_home_guard_test.go`, `init_quiet_wizard_test.go`, `init_shell_seam_test.go`)을 더한 합집합이다. 목록은 3개 이상이고, 목록의 모든 파일이 존재하며 검색기에 읽힌다. 목록의 어느 파일에도 `t.Setenv("HOME"`, `t.Parallel()`, 기존 HOME 헬퍼 호출 `runInitForAutonomyAtHomeCapturingOut(` 이 없다. 네 번째 파일 우회 뮤턴트 F 가 이 판정을 RED 로 만든 출력이 progress.md §E.2 에 있다(아래). 가드 판정 함수는 돌린 홈이 실제 홈과 같거나 그 아래일 때 오류를 돌려주고(음성 사례), 임시 디렉터리는 받아들인다. 실행 테스트는 전후 대조 목록 8항목(settings.json sha256, hooks/moai 존재, 셸 설정 6개 파일 `~/.zshenv`·`~/.zshrc`·`~/.zprofile`·`~/.profile`·`~/.bashrc`·`~/.bash_profile` 의 mtime·sha256) 중 하나라도 바뀌면 실패한다. 선택자가 가드 테스트를 실제로 골랐다.

```bash
go test ./internal/cli -run '^(TestHomeGuard_RejectsPathInsideRealHome|TestHomeGuard_AcceptsTempDir)$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac005.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: (TestHomeGuard_RejectsPathInsideRealHome|TestHomeGuard_AcceptsTempDir) ' .moai/state/verify/t583/ac005.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac005.txt
git diff --name-only --diff-filter=A develop...HEAD -- 'internal/cli/*_test.go' > .moai/state/verify/t583/ac005-added-committed.txt; echo "added-committed-exit=$?"
git ls-files --others --exclude-standard -- 'internal/cli/*_test.go' > .moai/state/verify/t583/ac005-added-uncommitted.txt; echo "added-uncommitted-exit=$?"
sort -u .moai/state/verify/t583/ac005-added-committed.txt .moai/state/verify/t583/ac005-added-uncommitted.txt > .moai/state/verify/t583/ac005-added.txt
git grep --untracked -lE 'runInit[A-Za-z]*\(|prepareSafeInitHome\(' -- 'internal/cli/*_test.go' | sort > .moai/state/verify/t583/ac005-callers.txt
comm -12 .moai/state/verify/t583/ac005-added.txt .moai/state/verify/t583/ac005-callers.txt > .moai/state/verify/t583/ac005-added-callers.txt
printf '%s\n' internal/cli/init_home_guard_test.go internal/cli/init_quiet_wizard_test.go internal/cli/init_shell_seam_test.go > .moai/state/verify/t583/ac005-planned.txt
sort -u .moai/state/verify/t583/ac005-planned.txt .moai/state/verify/t583/ac005-added-callers.txt > .moai/state/verify/t583/ac005-targets.txt
wc -l < .moai/state/verify/t583/ac005-targets.txt
git grep --untracked -l '^package cli' -- 'internal/cli/*_test.go' | sort > .moai/state/verify/t583/ac005-reach.txt
comm -12 .moai/state/verify/t583/ac005-targets.txt .moai/state/verify/t583/ac005-reach.txt | wc -l
git grep --untracked -lE 't\.Setenv\("HOME"|t\.Parallel\(\)|runInitForAutonomyAtHomeCapturingOut\(' -- 'internal/cli/*_test.go' | sort > .moai/state/verify/t583/ac005-forbidden.txt
comm -12 .moai/state/verify/t583/ac005-targets.txt .moai/state/verify/t583/ac005-forbidden.txt
comm -12 .moai/state/verify/t583/ac005-targets.txt .moai/state/verify/t583/ac005-forbidden.txt | wc -l
git grep -c 't.Setenv("HOME"' 120436f58 -- internal/cli/init_agent_wizard_test.go
git grep --untracked -c 'bash_profile' -- internal/cli/init_home_guard_test.go
```

기대: `test-exit=0`, PASS 줄 2, `no tests to run` 0. 두 추가분 명령은 종료 코드 0 이다(목록이 비어도 0). 대상 목록 수(`wc -l < …targets.txt`)가 3 이상이다. 도달 대조군 — 대상 목록과, 금지 토큰 검사와 같은 `git grep --untracked` 로 찾은 `package cli` 선언 파일의 교집합 수 — 이 대상 목록 수와 같다. 작으면 목록의 파일이 없거나 검색기가 읽지 못한 것이므로, 금지 토큰 교집합의 0 은 판정이 아니고 이 AC 는 실패다. 두 수가 같을 때 금지 토큰 교집합이 출력 없음·0 이면 "목록의 파일은 모두 있고 금지 토큰은 없다" 를 뜻한다. 패턴 대조군은 1 이상(plan 시점 대조군: 7), `bash_profile` 은 1 이상(대조 목록에 여섯째 셸 설정 파일이 들어 있음).

- 모든 `git grep` 에 `--untracked` 를 붙이는 이유는, run 단계가 새로 만든 파일을 커밋하기 전에 재면 추적 파일만 검색해 아무것도 읽지 않기 때문이다.
- 커밋분 추가 목록은 카드 범위 판정이므로 develop 병합 전에만 유효하다(서두 명령 작성 규칙).
- 계획한 세 파일 이름과 헬퍼 이름 `prepareSafeInitHome` 은 plan.md·design.md 의 제안이며, run 단계에서 바뀌면 이 명령의 이름을 함께 고친다. `init_home_guard_test.go` 는 `runInit` 을 직접 부르지 않을 수 있어 계획 목록에 따로 넣는다.
- 판별식은 init 을 실행하는 호출의 이름이다. 위 이름 중 어느 것도 부르지 않고 새 이름의 래퍼만 부르는 실행 테스트 파일은 스윕에서 빠질 수 있으므로, 래퍼를 새로 두면 판별식에 그 이름을 더한다.

뮤턴트 F (네 번째 파일 우회) — run 단계 progress.md §E.2 기록 의무:

- 적용: 계획한 세 파일 밖에 init 실행 테스트 파일(예: `internal/cli/init_quiet_wizard_mcp_test.go`)을 새로 두고, 그 안에서 `runInit(` 을 부르며 `t.Setenv("HOME", …)` 또는 `t.Parallel()` 을 쓴다. 기존 HOME 헬퍼 `runInitForAutonomyAtHomeCapturingOut(` 을 부르는 변형도 같다.
- 기대: 커밋하지 않은 상태로 위 스윕을 다시 돌리면, 그 파일이 미커밋 추가분과 호출 파일의 교집합으로 대상 목록에 들어가 대상 목록 수가 1 늘고, 금지 토큰 교집합이 그 파일을 찍어 1 이 된다 — RED. 옛 고정 세 파일 판정은 이 뮤턴트에서 존재·도달 대조군이 3, 금지 토큰 검사가 `exit=1` 로 통과했다(plan-audit 2회차 D9).
- 기록: 대상 목록 파일 내용, 대상 목록 수, 금지 토큰 교집합 출력과 수를 §E.2 에 남긴다. 뮤턴트 파일은 `go test` 를 돌리지 않고 스윕만 한 뒤 지운다 — 실행하지 않으므로 실제 홈에 닿지 않는다. 커밋하지 않는다.

plan 시점 대조군 (2026-09-11, 워크트리 `120436f58` — 계획한 세 파일 모두 아직 없고, 카드 추가분도 없음):

- 추가분: `added-committed-exit=0`, `added-uncommitted-exit=0`, 합친 목록 `0` 줄. 호출 파일은 기존 파일 `15` 개이지만 추가분과의 교집합은 `0` 이다 — 기존 실행 테스트는 대상에서 빠진다.
- 대상 목록 수 `3`(계획한 세 파일뿐). 하한 3 은 충족한다.
- 도달 대조군 `0`. 대상 목록 수 3 에 못 미치므로 RED(옳은 이유: M1·M2 가 파일을 만들기 전).
- 금지 토큰 교집합: 출력 없음, 수 `0`. 파일이 없어도 기대값과 같은 0 이 나온다 — 도달 대조군 없이 이 수만 읽으면 공허한 통과가 된다는 실측이다.
- `bash_profile` 명령: 출력 없음, 종료 코드 1. RED.
- 판정 형식 확인(기지 입력): 대상 목록 자리에 이미 있는 테스트 파일 세 개(`init_agent_wizard_test.go`, `init_autonomy_wiring_test.go`, `init_workflow_wiring_test.go`)를 넣으면 도달 대조군은 `3` 이고, 금지 토큰 교집합은 세 파일을 모두 찍는다. 도달 대조군이 채워질 수 있고 금지 토큰 교집합이 실제로 빨개질 수 있음을 관측했다.

### AC-IQW-006 — 미설정 = 기본값 (실행 테스트)

- **Given** 홈 안전 점검표를 갖춘 임시 프로젝트(셸 설정 단계 스파이 포함), 위저드 시접(`runWizardFn`)이 돌려주는 결과는 유지 4문항 답(`ConversationLang: "en"`, `UserName`, `AgentWiring: "claude"`, `AutonomyTier: "semi-auto"`)과 `RunWithDefaults` 의 시드 5개뿐
- **When** 실제 `runInit` 을 대화형 경로로 실행하면
- **Then** 아래 값이 모두 성립한다. 디스크 파일은 직접 읽고, workflow·feedback 값은 `config.NewLoader().Load(<project>/.moai)` 가 해석한 값으로 확인한다. 프로젝트 디렉터리는 파일시스템 루트가 아니다(REQ-IQW-003 예외 밖). 선택자가 테스트를 실제로 골랐다.

| 키 | 기대값 |
|---|---|
| `project.yaml` `project.name` | 프로젝트 디렉터리 이름 |
| `project.yaml` `project.mode` | `personal` |
| `report.yaml` `report.format` | `html+md` |
| `llm.yaml` `llm.profile` | `medium` |
| `llm.yaml` `llm.performance_tier` | 템플릿 값 `"medium"` 과 바이트 동일한 줄 |
| `workflow.worktree.auto_create` | `false` |
| `workflow.todo.enabled` | 디스크에 키 없음, 로더 해석 `true` |
| `feedback.auto_submit` | `false` |
| `workflow.project.continuation` | `card` |
| `workflow.audit.model` | 로더 해석 `claude` |
| `workflow.audit.gates` claude/codex/glm | 로더 해석 `required`/`required`/`advisory` |
| `workflow.codex.review_gate.enabled` | `false` |
| `.mcp.json` `mcpServers` | `moai` 키 포함 |

```bash
go test ./internal/cli -run '^TestRunInit_QuietWizardUnsetResolvesToDefaults$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac006.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestRunInit_QuietWizardUnsetResolvesToDefaults ' .moai/state/verify/t583/ac006.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac006.txt
```

기대: `test-exit=0`, PASS 줄 1, `no tests to run` 0.

### AC-IQW-007a — 관측기는 기본값이 아닌 값을 잡는다 (테스트 안 음성 대조군)

- **Given** AC-IQW-006 과 같은 실행으로 만든 프로젝트
- **When** 관측기를 돌리기 전에 복사본에서 제거 키 두 개를 기본값이 아닌 값으로 바꾸고(`project.yaml` `mode: team`, `workflow.yaml` `auto_create: true`) 같은 관측기를 돌리면
- **Then** 관측기가 두 키 모두를 불일치로 보고한다. 복사본을 바꾸지 않은 원본에 대해서는 불일치 0건이다. 선택자가 테스트를 실제로 골랐다.

```bash
go test ./internal/cli -run '^TestRunInit_QuietWizardObserverDetectsNonDefault$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac007a.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestRunInit_QuietWizardObserverDetectsNonDefault ' .moai/state/verify/t583/ac007a.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac007a.txt
```

기대: `test-exit=0`, PASS 줄 1, `no tests to run` 0.

### AC-IQW-007b — 코드 뮤턴트가 실행 테스트를 빨갛게 만든다 (기록 증거)

- **Given** GREEN 상태의 AC-IQW-006 테스트
- **When** 뮤턴트 A(대화형 블록에 `opts.ProjectMode = "team"` 삽입)와 뮤턴트 B(대화형 블록에 `opts.WorktreeAutoCreate, opts.WorktreeAutoCreateSet = true, true` 삽입)를 하나씩 적용해 같은 테스트를 돌리면
- **Then** 두 경우 모두 `--- FAIL: TestRunInit_QuietWizardUnsetResolvesToDefaults` 가 나오고, 실패 메시지가 해당 키를 가리킨다. 원복 후 다시 `--- PASS:` 가 나온다. 세 실행의 명령과 출력 파일 내용(FAIL·PASS 줄)을 progress.md §E.2 에 그대로 남긴다. 뮤턴트는 커밋하지 않는다. 세 실행 모두 AC-IQW-015 슬롯이다.

```bash
# 뮤턴트 A 적용 상태: FAIL 기대
go test ./internal/cli -run '^TestRunInit_QuietWizardUnsetResolvesToDefaults$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac007b-a.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- FAIL: TestRunInit_QuietWizardUnsetResolvesToDefaults ' .moai/state/verify/t583/ac007b-a.txt
# 뮤턴트 B 적용 상태: FAIL 기대
go test ./internal/cli -run '^TestRunInit_QuietWizardUnsetResolvesToDefaults$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac007b-b.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- FAIL: TestRunInit_QuietWizardUnsetResolvesToDefaults ' .moai/state/verify/t583/ac007b-b.txt
# 원복 후: PASS 기대
go test ./internal/cli -run '^TestRunInit_QuietWizardUnsetResolvesToDefaults$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac007b-revert.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestRunInit_QuietWizardUnsetResolvesToDefaults ' .moai/state/verify/t583/ac007b-revert.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac007b-revert.txt
git diff --quiet -- internal/cli/init.go; echo "reverted-exit=$?"
```

기대: 뮤턴트 두 파일은 `test-exit` 가 0 이 아니고 FAIL 줄 1(FAIL 줄이 0 이면 선택자가 테스트를 못 고른 것일 수 있어 판정 무효). 원복 파일은 `test-exit=0`, PASS 줄 1, `no tests to run` 0. `reverted-exit` 는 뮤턴트 적용 전에 `init.go` 가 커밋된 상태였다는 전제에서 0 이어야 한다. 전제가 성립하지 않으면(커밋 전 작업 중), 원복 확인은 뮤턴트 줄 문자열에 대한 `git grep -n 'ProjectMode = "team"' -- internal/cli/init.go` 종료 코드 1 로 대신한다.

### AC-IQW-008 — 대화형 4문항 실행의 섹션 파일은 비대화형과 같은 모양이다

- **Given** 같은 디렉터리 이름으로 만든 두 임시 프로젝트, 둘 다 홈 안전 점검표를 갖춤
- **When** 하나는 AC-IQW-006 의 대화형 실행, 다른 하나는 플래그 없는 `--non-interactive` 실행으로 init 하면
- **Then** 두 프로젝트의 `.moai/config/sections/` 아래 `workflow.yaml`, `project.yaml`, `report.yaml`, `feedback.yaml`, `llm.yaml` 이 바이트 동일하다. 특히 대화형 쪽 `workflow.yaml` 의 `workflow.audit` 아래에 `model`·`gates` 키가, `workflow` 아래에 `todo` 키가 삽입되지 않는다. 템플릿이 싣는 `audit.codex`·`audit.glm` 핀(`internal/template/templates/.moai/config/sections/workflow.yaml:85-91`)은 두 쪽 모두 그대로 남는다 — `audit:` 키 자체의 부재를 단언하지 않는다. 선택자가 테스트를 실제로 골랐다.

```bash
go test ./internal/cli -run '^TestRunInit_QuietWizardSectionFilesMatchNonInteractive$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac008.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestRunInit_QuietWizardSectionFilesMatchNonInteractive ' .moai/state/verify/t583/ac008.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac008.txt
```

기대: `test-exit=0`, PASS 줄 1, `no tests to run` 0.

### AC-IQW-009 — 대화형 경로는 MCP 항목 보장 호출을 기본으로 실행한다

- **Given** 홈 안전 점검표를 갖춘 대화형 실행, 주입 결과에 MCP 관련 필드 없음
- **When** 하네스 답을 `claude`, `codex`, `both` 로 각각 바꿔 `runInit` 을 실행하면(하위 테스트 3개)
- **Then** `claude` 와 `both` 에서는 표준 출력에 `Provisioned the moai MCP server entry in .mcp.json (default-on).` 가 있고 `.mcp.json` 에 `moai` 항목이 있다. `codex` 에서는 그 안내가 없다. 선택자가 최상위 테스트와 하위 테스트 3개를 실제로 돌렸다.

```bash
go test ./internal/cli -run '^TestRunInit_QuietWizardProvisionsMCPByDefault$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac009.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestRunInit_QuietWizardProvisionsMCPByDefault ' .moai/state/verify/t583/ac009.txt
command grep -cE -- '--- PASS: TestRunInit_QuietWizardProvisionsMCPByDefault/(claude|codex|both) ' .moai/state/verify/t583/ac009.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac009.txt
```

기대: `test-exit=0`, 최상위 PASS 줄 1, 하위 PASS 줄 3, `no tests to run` 0.

### AC-IQW-010 — 비대화형 고정 동작은 그대로다

- **Given** 이 SPEC 의 변경이 적용된 트리
- **When** 비대화형 고정 테스트 두 개를 돌리고, 그 본문을 기준 트리와 비교하면
- **Then** 둘 다 통과하고 본문 차이가 없다.

```bash
go test ./internal/cli -run '^(TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence|TestRunInit_WorkflowToggleFlagsAbsentByteIdentical)$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac010.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: (TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence|TestRunInit_WorkflowToggleFlagsAbsentByteIdentical) ' .moai/state/verify/t583/ac010.txt
git show 120436f58:internal/cli/init_agent_wizard_test.go | sed -n '/^func TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence/,/^}/p' > .moai/state/verify/t583/codex-absence-base.txt
sed -n '/^func TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence/,/^}/p' internal/cli/init_agent_wizard_test.go > .moai/state/verify/t583/codex-absence-head.txt
diff .moai/state/verify/t583/codex-absence-base.txt .moai/state/verify/t583/codex-absence-head.txt; echo "codex-absence-diff-exit=$?"
git show 120436f58:internal/cli/init_workflow_wiring_test.go | sed -n '/^func TestRunInit_WorkflowToggleFlagsAbsentByteIdentical/,/^}/p' > .moai/state/verify/t583/byte-identical-base.txt
sed -n '/^func TestRunInit_WorkflowToggleFlagsAbsentByteIdentical/,/^}/p' internal/cli/init_workflow_wiring_test.go > .moai/state/verify/t583/byte-identical-head.txt
diff .moai/state/verify/t583/byte-identical-base.txt .moai/state/verify/t583/byte-identical-head.txt; echo "byte-identical-diff-exit=$?"
wc -l .moai/state/verify/t583/codex-absence-base.txt .moai/state/verify/t583/byte-identical-base.txt
```

기대: `test-exit=0`, PASS 줄 2, 두 `diff-exit=0`, 기준 추출물이 비어 있지 않음(plan 시점 대조군: 20줄, 12줄).

### AC-IQW-011 — 플래그 경로는 계속 기록한다

- **Given** 홈 안전 점검표를 갖춘 대화형 실행(4문항 답만 주입)
- **When** `--project-mode team` 과 `--worktree-auto-create=true` 를 함께 주고 `runInit` 을 실행하면
- **Then** `project.yaml` 에 `mode: team`, `workflow.yaml` 에 `auto_create: true` 가 기록된다. 기존 플래그 기록 테스트도 통과한다. 선택자가 두 테스트를 실제로 골랐다.

```bash
go test ./internal/cli -run '^(TestRunInit_QuietWizardFlagsStillPersist|TestRunInit_WorkflowToggleFlagsPersist)$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac011.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: (TestRunInit_QuietWizardFlagsStillPersist|TestRunInit_WorkflowToggleFlagsPersist) ' .moai/state/verify/t583/ac011.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac011.txt
go test ./internal/core/project/... -run 'TestWriteProjectModeYAML|TestWriteWorkflowTogglesYAML' -count=1 -v > .moai/state/verify/t583/ac011-core.txt 2>&1; echo "test-exit=$?"
command grep -c 'no tests to run' .moai/state/verify/t583/ac011-core.txt
```

기대: 첫 실행 `test-exit=0`, PASS 줄 2, `no tests to run` 0. 둘째 실행 `test-exit=0`, `no tests to run` 0(기존 테스트 이름 접두사 선택이며 plan 시점에 정의가 있음).

### AC-IQW-012 — 워크트리 기록 규칙 서술이 하나로 정리된다

- **Given** 이 SPEC 의 변경이 적용된 트리
- **When** 위저드와 init 소스에서 워크트리 위저드 흔적을 찾고, `InitOptions` 워크트리 필드 주석에서 "wizard" 를 세면
- **Then** 위저드 패키지와 `internal/cli/init.go` 의 비테스트 소스에 `worktree_auto_create`·`WorktreeAutoCreate` 가 없고(`init_workflow_flags.go` 의 플래그 경로는 대상 아님), 필드 주석 블록에 "wizard" 가 0회다.

```bash
git grep -nE 'worktree_auto_create|WorktreeAutoCreate' -- internal/cli/wizard internal/cli/init.go ':!*_test.go'; echo "exit=$?"
git grep -cE 'worktree_auto_create|WorktreeAutoCreate' 120436f58 -- internal/cli/wizard internal/cli/init.go ':!*_test.go'
sed -n '/Worktree advisory/,/WorktreeAutoCreate bool/p' internal/core/project/initializer.go | awk 'tolower($0) ~ /wizard/ {c++} END {print "wizard-mentions=" c+0}'
git show 120436f58:internal/core/project/initializer.go | sed -n '/Worktree advisory/,/WorktreeAutoCreate bool/p' | awk 'tolower($0) ~ /wizard/ {c++} END {print "control-wizard-mentions=" c+0}'
```

기대: 첫 명령 `exit=1`, 둘째 명령(대조군) 1건 이상(plan 시점: `init.go:2`, `questions.go:2`, `translations.go:3`, `types.go:1`, `wizard.go:2`), `wizard-mentions=0`, `control-wizard-mentions` ≥ 1(plan 시점: 2). 주석 표제가 바뀌어 셋째 명령의 `sed` 범위가 비면 이 판정은 무효이며, 새 표제로 범위를 다시 잡아 잰다.

### AC-IQW-013 — 죽은 기록기와 필드가 정리되고, 유지 항목은 남는다

- **Given** plan.md §G 판정표
- **When** 삭제 항목과 유지 항목을 비테스트 소스에서 찾으면
- **Then** 삭제 항목은 0건, 유지 항목은 모두 존재한다.

```bash
git grep -nE 'writeWorkflowAuditYAML|writeWorkflowTodoYAML|writeFeedbackAutoSubmitYAML|writeWorkflowProjectContinuationYAML|AuditConfigSet' -- internal ':!*_test.go'; echo "deleted-exit=$?"
git grep -cE 'writeWorkflowAuditYAML|AuditConfigSet' 120436f58 -- internal/core/project/initializer.go
git grep -nE 'func writeProjectModeYAML|func WriteWorkflowTogglesYAML|func provisionMCPEntryUnlessDeclined|opts\.MCPProvision' -- internal ':!*_test.go'
go vet ./internal/cli/... ./internal/core/project/...
```

기대: `deleted-exit=1`. 둘째 명령(대조군) 1 이상(plan 시점: 5). 셋째 명령은 세 함수 선언과 `opts.MCPProvision` 을 읽는 줄·쓰는 줄(대화형 기본값)이 모두 나와야 한다(plan 시점 기준 트리에서는 5줄: 함수 3 + `init.go:337` 쓰기 + `init.go:1004` 읽기). `go vet` 종료 코드 0.

### AC-IQW-014 — 영향 패키지 전체 통과와 정적 검사

- **Given** M1~M5 가 끝난 트리
- **When** 영향 패키지 테스트와 정적 검사를 돌리면
- **Then** 모두 종료 코드 0 이고, 어느 패키지 출력에도 `[no test files]`·`[no tests to run]` 이 없다. plan.md §H 에서 갱신·삭제한 테스트의 처분이 커밋 메시지에 테스트별로 적혀 있다.

```bash
go test ./internal/cli/wizard/... -count=1 > .moai/state/verify/t583/ac014-wizard.txt 2>&1; echo "wizard-exit=$?"
go test ./internal/core/project/... -count=1 > .moai/state/verify/t583/ac014-project.txt 2>&1; echo "project-exit=$?"
go test ./internal/cli -count=1 -timeout 600s > .moai/state/verify/t583/ac014-cli.txt 2>&1; echo "cli-exit=$?"
command grep -cE 'no test files|no tests to run' .moai/state/verify/t583/ac014-wizard.txt .moai/state/verify/t583/ac014-project.txt .moai/state/verify/t583/ac014-cli.txt
golangci-lint run ./internal/cli/... ./internal/core/project/... > .moai/state/verify/t583/ac014-lint.txt 2>&1; echo "lint-exit=$?"
```

기대: 네 종료 코드 모두 0, 넷째 명령의 세 파일 수가 모두 0. 둘째·셋째 명령은 AC-IQW-015 슬롯이다. 전체 스위트는 로컬에서 돌리지 않는다. 최종 판정은 develop push 뒤 CI 다.

### AC-IQW-015 — 모든 init 실행 슬롯은 선언되고, 실제 홈 지문을 같은 명령으로 전후 채집한다

- **Given** run 단계에서 `./internal/cli` 또는 `./internal/core/project/...` 를 대상으로 하는 `go test` 호출 하나(이하 슬롯). 기존 init 실행 테스트만 도는 슬롯도, 새 테스트가 도는 슬롯도 같다
- **When** 작업자가 슬롯을 progress.md §E.2 슬롯 표에 `| SLOT-<번호> |` 로 시작하는 행으로 먼저 선언하고, 아래 지문 명령을 실행 직전에 before 이름으로, `go test` 를 돌린 직후에 after 이름으로 실행하고, 두 출력을 diff 하면
- **Then** 슬롯마다 다음이 모두 성립한다.
  - 두 번의 지문 명령 본문이 글자 그대로 같다(출력 파일 이름의 `before`/`after` 만 다름).
  - 두 `.err` 파일이 모두 0바이트이고, `before.out` 이 8줄이다.
  - `home-diff-exit=0` 이다.
  - 그 슬롯 행에 실행한 지문 명령, 두 `.out` 파일의 내용, 두 `fingerprint-exit` 값, `go test` 종료 코드, `home-diff-exit` 값이 기록돼 있다. `home-diff-exit` 기록은 슬롯 행마다 정확히 한 번이다.
  - 전체로는 선언 슬롯 수 = `home-SLOT-*-after.out` 파일 수 = `home-diff-exit=0` 기록 수이며, 셋 모두 1 이상이다. 0 이 아닌 `home-diff-exit` 기록은 없다.
  - 차이가 나거나 `.err` 가 비어 있지 않으면 작업자는 이후 작업을 멈추고 리드에게 보고하며 실제 홈 파일을 되돌리지 않는다(REQ-IQW-016) — 이때 이 AC 는 통과가 아니라 보류다.

지문 명령(기준 형태, `<n>` 은 슬롯 번호로 바꾸고 before/after 두 번 모두 같은 번호를 쓴다):

```bash
mkdir -p .moai/state/verify/t583
{ if [ -e "$HOME/.claude/settings.json" ]; then printf 'settings.json sha256=%s\n' "$(shasum -a 256 "$HOME/.claude/settings.json" | cut -d' ' -f1)"; else echo 'settings.json absent'; fi; if [ -d "$HOME/.claude/hooks/moai" ]; then echo 'hooks/moai present'; else echo 'hooks/moai absent'; fi; for f in .zshenv .zshrc .zprofile .profile .bashrc .bash_profile; do if [ -e "$HOME/$f" ]; then printf '%s mtime=%s sha256=%s\n' "$f" "$(stat -f %m "$HOME/$f")" "$(shasum -a 256 "$HOME/$f" | cut -d' ' -f1)"; else printf '%s absent\n' "$f"; fi; done; } > .moai/state/verify/t583/home-SLOT-<n>-before.out 2> .moai/state/verify/t583/home-SLOT-<n>-before.err; echo "fingerprint-exit=$?"
```

실행 후에는 위 명령에서 `before` 두 곳만 `after` 로 바꿔 그대로 실행한다. 이어서 슬롯별로:

```bash
diff .moai/state/verify/t583/home-SLOT-<n>-before.out .moai/state/verify/t583/home-SLOT-<n>-after.out; echo "home-diff-exit=$?"
wc -c .moai/state/verify/t583/home-SLOT-<n>-before.err .moai/state/verify/t583/home-SLOT-<n>-after.err
wc -l .moai/state/verify/t583/home-SLOT-<n>-before.out
```

run 단계 마감 시 전체로:

```bash
command grep -c '^| SLOT-' .moai/specs/SPEC-INIT-QUIET-WIZARD-001/progress.md
find .moai/state/verify/t583 -type f -name 'home-SLOT-*-after.out' | wc -l
command grep -cE '^\| SLOT-.*home-diff-exit=0' .moai/specs/SPEC-INIT-QUIET-WIZARD-001/progress.md
command grep -cE '^\| SLOT-.*home-diff-exit=[1-9]' .moai/specs/SPEC-INIT-QUIET-WIZARD-001/progress.md
```

기대: 슬롯별로 `home-diff-exit=0`, 두 `.err` 가 0바이트, `before.out` 이 8줄(0줄이면 지문이 비어 판정 무의미). 마감 시 앞 세 수가 서로 같고 1 이상, 넷째 수가 0. 셋 중 하나라도 다르면 선언되지 않았거나 지문·기록이 빠진 슬롯이 있다는 뜻이므로 실패다.

근거와 한계:

- **같은 명령을 강제하는 이유**: 앞뒤를 서로 다른 도구로 재면 내용이 같아도 출력 형식이 달라 거짓 차이가 난다. 카드 t661 에서 실제로 일어났다.
- **표준 오류를 분리하는 이유**: 권한 오류나 도구 부재가 표준 출력에 섞이면 지문 내용이 오염되고, 섞이지 않더라도 어떤 항목을 못 잰 채 같은 출력이 나올 수 있다. 그래서 `.err` 가 비어 있지 않으면 차이와 같이 다룬다.
- **`fingerprint-exit` 는 약한 신호다**: 묶음 명령의 종료 코드는 마지막 명령의 것이라 중간 실패를 가리지 못한다. 판정은 `.err` 0바이트, 8줄 출력, `home-diff-exit` 로 한다.
- **삼자 일치의 한계**: 이 판정은 선언·지문·기록이 서로 어긋난 슬롯을 잡지만, 선언도 지문도 없이 실행된 `go test` 호출은 잡지 못한다. 셸 이력을 기계적으로 열거할 수단이 없기 때문이다. 이 공백은 §C-4 의 "실행 전 선언" 규율에 기댄다.
- **`stat -f %m` 은 macOS 형식이다**: 레인이 Linux 라면 `stat -c %Y` 로 바꾸되, 한 슬롯의 before/after 는 반드시 같은 형식을 쓴다. 쓴 형식을 슬롯 행에 적는다.
- **판독의 실행 확인**: 기존 HOME 헬퍼를 쓰면 셸 감지도 임시 홈을 따라가 셸 설정 누출이 없다는 것은 `internal/shell/detect.go:128` 의 코드 판독일 뿐이다. 기존 헬퍼를 쓰는 테스트가 도는 첫 슬롯의 `home-diff-exit=0` 이 이 판독의 첫 실행 확인이며, 그 슬롯 행에 표시한다.
- **이미 설정된 머신**: `internal/shell/config.go:93`, `:167` 은 줄이 이미 있으면 기록을 건너뛰므로, 그런 머신에서의 지문 불변은 누출이 없었다는 강한 증거가 못 된다. 셸 설정 단계 시접의 효과는 AC-IQW-004 의 스파이 호출 횟수로 증명한다.
- **동시 세션**: 다른 세션이 같은 시각에 `~/.claude/settings.json` 을 고치면 이 SPEC 과 무관한 차이가 날 수 있다. 그래도 절차는 같다 — 멈추고 보고하며, 원인 판정은 리드가 한다.

### AC-IQW-016 — 시접을 바꾸는 테스트는 병렬로 돌지 않고, `t.Cleanup` 으로 원래 값을 되돌린다

리드가 내보낸 시접 `project.ConfigureShellEnvFn` 하나(design.md §4.2)를 받아들이며 붙인 조건이다(2026-09-11). spec.md §4.2 3·6항의 의무를, 고정 파일 목록이 아니라 시접에 값을 대입하는 모든 테스트에 대해 판정한다. AC-IQW-005 의 `t.Parallel()` 검사는 `internal/cli` 파일 세 개만 보므로 `internal/core/project` 게이트 테스트를 덮지 못한다 — 이 AC 가 그 공백을 맡는다.

- **Given** `internal/cli`·`internal/core/project`(하위 디렉터리 포함)의 테스트 파일 가운데 `ConfigureShellEnvFn` 에 값을 대입하는 테스트 함수나 헬퍼가 든 파일 전부(이하 대입 파일). 목록은 검증 시점에 대입 패턴으로 다시 뽑고, 아직 커밋하지 않은 파일도 잡도록 `git grep --untracked` 를 쓴다. plan.md M1 이 계획한 대입 파일은 두 개다 — `internal/cli` 홈 안전 헬퍼 파일(`init_home_guard_test.go`)과 `internal/core/project` 게이트 테스트 파일(`initializer_shell_seam_test.go`). run 단계에서 파일 이름이 바뀌어도 판정은 패턴으로 뽑은 목록을 쓴다.
- **When** 대입 파일을 스윕하고, 병렬 금지와 원복 두 절반을 각각 아래 관측으로 재면
- **Then**
  - **스윕이 비지 않았다.** 대입 파일이 모두 2개 이상이고, `internal/cli` 쪽 1개 이상, `internal/core/project` 쪽 1개 이상이다. 어느 하나라도 모자라면 이 AC 는 실패다 — 계획한 대입 지점이 없거나 패턴이 어긋났다는 뜻이며, 나머지 판정은 빈 목록 위의 통과라 무의미하다.
  - **병렬 금지 절반**
    - 텍스트 관측: 대입 파일 가운데 `t.Parallel(` 을 담은 파일이 0개다. 대입 파일은 모두 `t.Setenv(` 를 담는다.
    - 실행 관측: 시접을 바꾸는 헬퍼와 게이트 테스트는 시접을 바꾸기 전에 `t.Setenv` 를 한 번 이상 부른다(design.md §4.2). Go testing 은 `t.Setenv` 를 부른 테스트가 `t.Parallel` 을 부르거나, 자신이나 조상이 병렬인 테스트가 `t.Setenv` 를 부르면 `testing: test using t.Setenv, t.Chdir, or cryptotest.SetGlobalRandom can not use t.Parallel` 로 panic 한다(go1.26.8 `src/testing/testing.go:1752`, `:1765-1766`, `:1830-1838` 판독). 그래서 대입 파일 밖에서 헬퍼를 부르는 테스트에 `t.Parallel` 이 더해져도 실행이 실패한다. 이 경로는 텍스트 관측이 보지 못하는 곳이며, panic 이 실제로 나는지는 아래 뮤턴트 D 로 한 번 관측해 기록한다.
  - **원복 절반**
    - 텍스트 관측: 대입 파일마다 대입 줄이 2줄 이상(바꿔 끼우기·되돌리기)이고, 대입 파일은 모두 `t.Cleanup(` 을 담는다. 텍스트는 되돌리는 대입이 실제로 `t.Cleanup` 안에 있는지, 실행 경로에서 도달하는지는 보지 못한다. 그 부분은 아래 실행 관측이 맡는다.
    - 실행 관측 (`internal/core/project`): 게이트 테스트와 기본값 직접 읽기 테스트(AC-IQW-004 보조 관측)를 `-count=2` 로 한 프로세스에서 돌린다. `go test` 는 고른 테스트 목록 전체를 회차마다 차례로 다시 돌리므로, 둘째 회차의 기본값 테스트는 소스 순서와 무관하게 첫 회차 게이트 테스트 뒤에 실행된다. 게이트 테스트가 원복하지 않으면 그 기본값 테스트가 스파이의 함수 포인터를 읽어 실패한다.
    - 실행 관측 (`internal/cli`): 헬퍼는 시접을 바꾸기 전에, 현재 값이 테스트 패키지 초기화 때 잡아 둔 원래 값과 같은 함수인지 `reflect` 함수 포인터로 단언하고 다르면 `t.Fatal` 한다(design.md §5 4번). 주 관측 테스트를 `-count=2` 로 돌리면 둘째 회차의 이 단언이 첫 회차의 원복을 실행으로 확인한다.
  - 선택자가 테스트를 실제로 골랐다(스윕 확인).
  - 뮤턴트 3종이 판정을 RED 로 만든 출력이 progress.md §E.2 에 있다. 뮤턴트는 커밋하지 않는다. 셋 모두 남는 값이 스파이라 실제 셸 설정 파일에 쓰지 않으므로 plan.md §D 의 금지 뮤턴트에 해당하지 않는다.
    - 뮤턴트 C1 — `internal/core/project` 게이트 테스트에서 되돌리는 대입(`t.Cleanup` 등록)을 지운다. 기대: 그 파일의 대입 줄 수가 뮤턴트 전보다 1 줄어들고, `-count=2` 실행에서 `--- FAIL: TestConfigureShellEnvFn_DefaultIsProductionFunc` 가 1줄 이상. 하위 테스트마다 바꿔 끼우는 구현이면 줄어든 뒤에도 2 이상이 남아 텍스트 관측은 RED 가 되지 않을 수 있으므로, 이 뮤턴트의 판정은 실행 RED 로 한다.
    - 뮤턴트 C2 — `internal/cli` 헬퍼에서 되돌리는 대입을 지운다. 기대: 텍스트 관측 RED, `-count=2` 실행에서 `--- FAIL: TestRunInit_ShellConfigStepReachedViaSeam` 1줄(둘째 회차 진입 단언).
    - 뮤턴트 D — 대입 파일이 아닌 주 관측 테스트 파일에서, 그 테스트 첫 줄에 `t.Parallel()` 을 넣는다. 기대: 텍스트 관측은 통과하고(대입 파일이 아니므로), 실행은 `can not use t.Parallel` panic 으로 실패한다. 대입 파일 안에 넣은 변형은 텍스트 관측의 `parallel-exit` 가 잡는다.

```bash
git grep --untracked -lE 'ConfigureShellEnvFn[[:space:]]*=[^=]' -- 'internal/cli/*_test.go' 'internal/core/project/*_test.go' > .moai/state/verify/t583/ac016-swappers.txt; echo "sweep-exit=$?"
wc -l < .moai/state/verify/t583/ac016-swappers.txt
git grep --untracked -lE 'ConfigureShellEnvFn[[:space:]]*=[^=]' -- 'internal/cli/*_test.go' | wc -l
git grep --untracked -lE 'ConfigureShellEnvFn[[:space:]]*=[^=]' -- 'internal/core/project/*_test.go' | wc -l
git grep --untracked -lE --all-match -e 'ConfigureShellEnvFn[[:space:]]*=[^=]' -e 't\.Parallel\(' -- 'internal/cli/*_test.go' 'internal/core/project/*_test.go'; echo "parallel-exit=$?"
git grep --untracked -lE --all-match -e 'ConfigureShellEnvFn[[:space:]]*=[^=]' -e 't\.Setenv\(' -- 'internal/cli/*_test.go' 'internal/core/project/*_test.go' | wc -l
git grep --untracked -lE --all-match -e 'ConfigureShellEnvFn[[:space:]]*=[^=]' -e 't\.Cleanup\(' -- 'internal/cli/*_test.go' 'internal/core/project/*_test.go' | wc -l
git grep --untracked -cE 'ConfigureShellEnvFn[[:space:]]*=[^=]' -- 'internal/cli/*_test.go' 'internal/core/project/*_test.go'
go test ./internal/core/project/... -run '^(TestInitializer_ShellConfigSeamGate|TestConfigureShellEnvFn_DefaultIsProductionFunc)$' -count=2 -v > .moai/state/verify/t583/ac016-project.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: (TestInitializer_ShellConfigSeamGate|TestConfigureShellEnvFn_DefaultIsProductionFunc) ' .moai/state/verify/t583/ac016-project.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac016-project.txt
go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=2 -timeout 600s -v > .moai/state/verify/t583/ac016-cli.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- PASS: TestRunInit_ShellConfigStepReachedViaSeam ' .moai/state/verify/t583/ac016-cli.txt
command grep -c 'no tests to run' .moai/state/verify/t583/ac016-cli.txt
# 뮤턴트 C1 적용 상태: FAIL 기대 (적용 뒤 위 스윕의 -c 줄도 다시 돌려 그 파일 수가 적용 전보다 1 줄었는지 본다)
go test ./internal/core/project/... -run '^(TestInitializer_ShellConfigSeamGate|TestConfigureShellEnvFn_DefaultIsProductionFunc)$' -count=2 -v > .moai/state/verify/t583/ac016-mutant-c1.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- FAIL: TestConfigureShellEnvFn_DefaultIsProductionFunc ' .moai/state/verify/t583/ac016-mutant-c1.txt
# 뮤턴트 C2 적용 상태: FAIL 기대
go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=2 -timeout 600s -v > .moai/state/verify/t583/ac016-mutant-c2.txt 2>&1; echo "test-exit=$?"
command grep -cE -- '^--- FAIL: TestRunInit_ShellConfigStepReachedViaSeam ' .moai/state/verify/t583/ac016-mutant-c2.txt
# 뮤턴트 D 적용 상태: panic 기대
go test ./internal/cli -run '^TestRunInit_ShellConfigStepReachedViaSeam$' -count=1 -timeout 600s -v > .moai/state/verify/t583/ac016-mutant-d.txt 2>&1; echo "test-exit=$?"
command grep -c 'can not use t.Parallel' .moai/state/verify/t583/ac016-mutant-d.txt
```

기대: `sweep-exit=0`, 대입 파일 수 ≥ 2, `internal/cli` 쪽 ≥ 1, `internal/core/project` 쪽 ≥ 1. `parallel-exit=1`. `t.Setenv` 동시 포함 수와 `t.Cleanup` 동시 포함 수가 모두 대입 파일 수와 같다. `-c` 출력에서 파일마다 수가 2 이상이다. `internal/core/project` 실행은 `test-exit=0`·PASS 줄 4·`no tests to run` 0, `internal/cli` 실행은 `test-exit=0`·PASS 줄 2·`no tests to run` 0. 뮤턴트 C1·C2 파일은 `test-exit` 가 0 이 아니고 FAIL 줄 1 이상, 뮤턴트 D 파일은 `test-exit` 가 0 이 아니고 panic 줄 1 이상. 원복 뒤 두 정상 실행을 다시 돌려 위 PASS 줄 수를 다시 얻고, 원복 확인은 AC-IQW-004 와 같은 `git diff --quiet -- <파일>` 형식으로 한다. 이 AC 의 모든 `go test` 줄은 AC-IQW-015 슬롯이다. 테스트 이름은 design.md 의 제안이며, 이름이 바뀌면 명령의 이름을 함께 고친다.

plan 시점 대조군 (2026-09-11, 워크트리 `120436f58`):

- **RED-now (옳은 이유)**: 같은 스윕(`--untracked`)의 대입 파일 수는 0 이다. 시접도 대입 테스트도 아직 없다 — `git grep -c 'ConfigureShellEnvFn' 120436f58 -- internal` 이 종료 코드 1 이다. M1 이 이 수를 뒤집는다.
- **대입 패턴이 대입 줄을 잡는다**: 같은 패턴을 기존 시접 `userHomeDirFn` 에 쓰면(`git grep -lE 'userHomeDirFn[[:space:]]*=[^=]' 120436f58 -- 'internal/cli/*_test.go'`) 파일 8개가 잡히고, 그중 `update_home_seam_test.go` 의 대입 줄 수는 2(바꿔 끼우기·되돌리기)다.
- **`--all-match` 는 교집합이다**: 같은 대입 패턴에 존재하지 않는 토큰을 둘째 패턴으로 주면 `--all-match` 로는 0개, `--all-match` 없이는 8개다. 둘째 패턴을 `t\.Cleanup\(` 로 주면 8개다.
- **`--untracked` 가 필요하다**: 아직 커밋하지 않은 이 SPEC 디렉터리에서 `AC-IQW-015` 를 찾으면 `--untracked` 없이는 종료 코드 1, 있으면 파일 7개다. `--untracked` 를 빼면 run 단계에서 커밋 전에 잰 스윕이 0 으로 나와 빈 스윕 실패와 구별되지 않는다.

근거와 한계:

- **텍스트 관측이 못 보는 두 곳**: 되돌리는 대입이 실제로 `t.Cleanup` 안에 있는지, 그리고 헬퍼를 부르는 다른 파일의 테스트가 병렬인지는 텍스트로 알 수 없다. 앞의 것은 `-count=2` 재진입 관측이, 뒤의 것은 `t.Setenv` 병렬 충돌 panic 이 실행으로 메운다.
- **`-count=2` 가 보는 것은 원복 여부다**: 원복이 `t.Cleanup` 으로 등록됐는지는 보지 않는다. 테스트 함수 끝에서 직접 되돌리는 구현도 통과하지만, 그 구현은 테스트가 `t.Fatal` 로 중간에 끝나면 원복을 건너뛴다. "`t.Cleanup` 으로" 라는 형식은 텍스트 관측(`t.Cleanup(` 포함)에만 기대며, 그 관측은 같은 파일의 다른 `t.Cleanup`(예: 지문 재채집)으로도 충족된다 — 이 잔여 위험은 코드 리뷰가 받친다.
- **`go test -race` 를 증거로 쓰지 않는 이유**: 경합 검출은 두 고루틴이 실제로 겹칠 때만 보고한다. 같은 패키지에서 전역 변수를 건드리는 병렬 테스트가 하나뿐이면 `t.Parallel` 이 더해져도 조용하다.
- **패턴의 가정**: 대입 연산자 뒤의 값이 같은 줄에 있다고 가정한다(gofmt 형식). 값을 다음 줄로 넘긴 대입은 스윕에서 빠질 수 있고, 그 경우 대입 파일 수가 계획보다 적게 나와 스윕 실패로 드러난다.

## §3 경계 사례

- **프로필 없는 첫 실행**: `conversation_language`·`user_name` 미리 채움 값이 비어도 4문항 구성은 같다(AC-IQW-001 은 입력과 무관).
- **하네스 `codex`**: 대화형 기본값이 켜져 있어도 `codex` 는 보장 호출을 건너뛴다(AC-IQW-009).
- **자율 등급 `fully-autonomous` + 샌드박스 증명 없음**: 기존 강등 동작은 이 SPEC 의 대상이 아니며, 관련 기존 테스트가 계속 통과해야 한다(AC-IQW-014).
- **파일시스템 루트에서 init**: `project_name` 기본값이 디렉터리 이름 규칙을 따라 오늘 위저드의 `my-project` 치환과 달라질 수 있다. REQ-IQW-003 이 이 경우를 동일성 요구에서 명시적으로 제외했으며, 이 경우를 재는 AC 는 두지 않는다(측정하지 않음).
- **`--force` 재초기화와 기존 `.mcp.json`**: 미측정, 요구 아님.
- **셸 설정 파일이 없는 홈**: 지문은 `absent` 로 기록되고 before/after 가 같으면 통과다. 실행 도중 파일이 생기면 `absent` → mtime 줄로 바뀌어 차이로 잡힌다(AC-IQW-015).
- **셸 설정 줄이 이미 있는 홈**: 지문 불변은 시접 효과의 증거가 아니다. 시접 효과는 AC-IQW-004 가 스파이로 증명한다.

## §4 품질 게이트

- TRUST 5: 새 테스트는 실행 경로를 재고(Tested), 제거 선례 패턴을 따르며(Unified), 홈 쓰기 누출을 막고 실행마다 실제 홈을 잰다(Secured). 커밋마다 카드 id `t583` 을 적는다(Trackable).
- 커버리지: `internal/cli/wizard` 와 `internal/core/project` 는 85% 목표를 유지한다. 삭제한 코드의 테스트도 함께 지우므로 비율이 떨어지면 원인을 기록한다.
- LSP: 영향 패키지 오류·타입 오류 0.

## §5 완료 정의

- [ ] AC-IQW-001 ~ AC-IQW-016 전부 통과. AC-IQW-004 배선 제거 뮤턴트, AC-IQW-007b 코드 뮤턴트, AC-IQW-016 뮤턴트 C1·C2·D 증거가 progress.md §E.2 에 있음
- [ ] 새 테스트를 지목한 모든 `-run` 실행 출력에 지목 수만큼 `--- PASS:` 줄이 있고 `[no tests to run]` 이 없음
- [ ] plan.md §G 판정표의 각 항목이 재확인 명령과 함께 처리됨
- [ ] plan.md §H 고정 테스트 처분이 커밋 메시지에 기록됨
- [ ] 위저드 변경이 plan.md §F.1 구간 안에 머묾. 카드 범위 판정이므로 develop 병합 전에 흡수한 `develop` 과의 merge-base 부터 잰다: `git diff --stat develop...HEAD -- internal/cli/wizard`(커밋 변경)와 `git diff --stat HEAD -- internal/cli/wizard`(커밋 전 변경)로 파일 목록을 확인하고, `git diff --name-only develop...HEAD | wc -l` 이 0 이면 "측정 불가" 로 보고함
- [ ] 이 SPEC 이 새로 쓰거나 본문을 다시 쓴 init 실행 테스트는 코드 안 전후 대조(8항목)가 불변으로 기록됨. 필드 참조만 지운 기존 테스트는 기존 HOME 헬퍼를 유지함
- [ ] `./internal/cli`·`./internal/core/project/...` 대상 `go test` 슬롯마다 선언 행과 실제 홈 지문 증거(명령·두 지문·종료 코드·`home-diff-exit=0`)가 progress.md §E.2 에 있고, 선언 수·지문 파일 수·기록 수가 일치함. 기존 헬퍼 테스트가 돈 첫 슬롯이 표시됨
