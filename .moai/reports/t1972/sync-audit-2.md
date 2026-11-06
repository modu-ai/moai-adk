auditor-model: claude-opus-5-5

# t1972 sync 감사 2회차 — moai init 설문·프로필 동기화 잔여 결함 5건 (Class B)

verdict: PASS-WITH-DEBT
audited_sha: f5dbf380ec368428eedd7c77ad5fd8e12e980944

- 감사 범위: 카드 전체 9b632c1eb3..f5dbf380ec (8e0ca7660f F4+F5, 9a5dc4b672 F6, 388967b934 N6+N7, f5dbf380ec F6 되돌림+N7 머리 주석+1회차 판정 파일)
- 트리: /Users/goos/.moai/worktrees/projects/moai-adk-go/t1972 (브랜치 WT-moai-init-f4). 감사 시작 시 HEAD = f5dbf380ec368428eedd7c77ad5fd8e12e980944, `git status --short` 비어 있음. 추적 파일 수정·변이 git 명령 없음.
- 종합 점수: 0.85 (Functionality 80 / Security 95 / Craft 80 / Consistency 85, 가중 합 84.5)
- 판정 근거: 1회차의 유일한 blocking 결함(F6-R1/R2)은 F6 변경을 통째로 되돌려 사라졌다. 되돌림은 깨끗하다(wizard.go·골든 base와 동일, select-height 테스트 없음). F4·F5·N6·N7은 수리·테스트가 그대로 유지된다. 다만 카드 본문은 다섯 항목 모두 "재현 → 최소 수리"를 요구하는데 F6은 재현만 되고 수리되지 않았으며, F6을 넘겨받을 후속 카드가 큐에 아직 없다. 이 분리는 1회차 감사가 제시한 선택지 (c)이고 t1966 선례와 같으므로 차단 사유로 보지 않되, 후속 카드 발행을 이 카드를 닫기 전의 구속 조건(부채 D1)으로 둔다. 리더가 분리를 받아들이지 않으면 이 판정은 FAIL로 바뀐다(Functionality "부분 점수 없음" 규칙).

## 차원 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 80 | PASS (부채 조건부) | F4·F5·N6·N7 테스트 통과(E3), F6은 재현 기록만 있고 수리 없음 — 분리 결정 기록됨, 후속 카드 미발행(E6) |
| Security (25%) | 95 | PASS | 되돌림으로 보안 표면 변화 없음. N6은 깨진 파일을 건드리지 않고 거부. 비밀값·입력 경계 변화 없음 |
| Craft (20%) | 80 | PASS (커버리지 일부 미측정) | profile 패키지 85.0%, `SyncToProjectConfig` 82.9%(E5). internal/cli 커버리지 미측정. F6 RED 테스트가 삭제돼 F6 재현 증거는 문서에만 남음(D2) |
| Consistency (15%) | 85 | PASS | gofmt 출력 없음, go vet exit 0(E4). 무시 규칙 대상 판정 파일을 강제 추가(D3, 저장소 관행과는 일치) |

## 결함 목록 (structured defect-list)

- **D1** [Medium] [blocking-for-close → 부채 구속 조건] [in-diff, 재현 있음] `internal/cli/wizard/wizard.go` `selectHeight`(base 코드) — F6은 재현됐지만 이 카드에서 수리되지 않았고, 넘겨받을 후속 카드가 큐에 없다(`moai todo list` 210행에서 selectHeight·REQ-ITI-015·wrapped option 검색 결과는 원 카드 t1966 한 줄뿐). 진행 기록은 "carry it as its own card"라고만 적고 카드 번호가 없다. 요구 수정: 리더가 F6 후속 카드를 발행하고(실행 중 로케일과 실제 폭으로 높이 계산, 로케일별 빈 카드 줄 0 단정, 아래 D2의 base 골든 갱신 포함), 이 카드의 종결 기록에 "F6 분리 → t<번호>"를 남긴 뒤 닫는다. 레인은 큐에 카드를 넣을 수 없으므로(큐 진입은 운영자 행위) 리더 몫이다.
- **D2** [Low] [optional] [in-diff] `internal/cli/wizard/testdata/axis/init-regroup-second-group.golden`(base로 복원) — 복원된 골든은 80열 영어에서 "Auto mode" 줄바꿈 부분과 "Bypass permissions" 옵션이 가려진 상태(빈 줄 3행)를 고정한다. 즉 현재 테스트 스위트가 F6 결함을 정답으로 잠그고 있다. 또 F6 RED 테스트가 삭제돼 재현 증거는 1회차 판정문과 진행 기록에만 남는다. 요구 수정: 후속 카드에서 RED를 다시 세우고 골든을 함께 고친다. 이 카드에서는 조치 불필요.
- **D3** [Low] [optional] [in-diff] `.moai/reports/t1972/sync-audit.md` — `.gitignore:235 .moai/reports/*`가 무시하는 경로를 `git add -f`로 커밋했다. 감사 판정문은 로컬 보관이 원칙이라는 감사자 지침과 어긋나지만, 저장소에는 이미 같은 방식으로 추적되는 보고 파일이 1992개 있어 관행과는 맞는다. 이 파일은 `verdict: FAIL`·`audited_sha: 388967b9…` 기계 줄을 담은 채 브랜치와 함께 착지한다. 요구 수정(선택): 남길지 리더가 정한다.
- **D4** [Info] [optional] [in-diff] 브랜치 이력에 F6 수리(9a5dc4b672)와 그 되돌림(f5dbf380ec)이 둘 다 남는다. 트리 기준 순변화는 0이다(E1). squash 착지면 사라지고 --no-ff면 이력에 남는다. 조치 불필요.
- **이월(1회차 optional, 이번에도 미처리 — 판정 영향 없음)**: F4-R1 `init.go:96-97` "the wizard presents the same three pages" 주석 부정확, F5-R2(outside) 재구성 위저드가 저장된 셀프호스팅 URL을 미리 채우지 않아 Enter만으로 gitlab.com으로 덮어씀(데이터 손실성 — 운영자에게 보일 것), N6-R1 거부가 language·statusline 동기화까지 막음. F5-R1은 진행 기록 Iteration 2에 "대화형 경로는 저장기 분기에 도달하지 않음, 저장기 단위 시험"으로 정정되어 처리된 것으로 본다. N7-R1은 머리 주석 "items 1, 2, 5, 7"로 처리됨(E2).

## Claim / Evidence / Baseline-attribution

### C1. F6 흔적이 남지 않았다 — 확인
- E1a: `git diff --stat 9b632c1eb3 f5dbf380ec368428eedd7c77ad5fd8e12e980944 -- internal/cli/wizard/wizard.go internal/cli/wizard/testdata internal/cli/wizard/select_height_t1972_test.go; echo "exit=$?"` →
  ```
  exit=0
  ```
  (출력 없음 = 세 경로 모두 base와 동일)
- E1b: `ls internal/cli/wizard/ | grep -i select; echo "ls-exit=$?"` → `ls-exit=1` (select 이름 파일 없음)
- E1c: 카드 전체 diff stat(`git diff 9b632c1eb3 f5dbf380ec… --stat`):
  ```
   .moai/reports/t1972/sync-audit.md                 | 145 ++++++++++++++++++++++
   internal/cli/init.go                              |   4 +-
   internal/cli/init_help_default_t1972_test.go      |  30 +++++
   internal/cli/init_t1966_test.go                   |  10 +-
   internal/cli/update_wizard.go                     |  12 +-
   internal/cli/update_wizard_gitlab_t1972_test.go   |  92 ++++++++++++++
   internal/cli/wizard/questions.go                  |   7 +-
   internal/profile/sync.go                          |  12 +-
   internal/profile/sync_duplicate_key_t1972_test.go |  47 +++++++
   9 files changed, 348 insertions(+), 11 deletions(-)
  ```
- 의존 여부: 남은 wizard 패키지 변경은 `questions.go`의 `DefaultGitLabInstanceURL` 상수뿐이며 F5(`update_wizard.go`와 그 테스트)가 쓰는 것이다. F6 커밋 9a5dc4b672는 wizard.go·골든·select-height 테스트 셋만 건드렸고(1회차 확인), 되돌림 커밋 stat도 그 셋에 init_t1966_test.go·판정 파일만 더한다:
  ```
   .moai/reports/t1972/sync-audit.md                  | 145 +++++++++++++++++++++
   internal/cli/init_t1966_test.go                    |   2 +-
   internal/cli/wizard/select_height_t1972_test.go    |  96 --------------
   .../testdata/axis/init-regroup-second-group.golden |   6 +-
   internal/cli/wizard/wizard.go                      |  30 +----
   5 files changed, 150 insertions(+), 129 deletions(-)
  ```
  wizard 패키지 시험이 base 골든으로 통과(E3a)하므로 남은 코드가 F6 변경에 기대지 않는다.

### C2. N7 머리 주석이 사실과 맞다 — 확인
- E2: `grep -n '^// Item\|^// Card' internal/cli/init_t1966_test.go` →
  ```
  3:// Card t1966 — `moai init` wizard/flag regression tests (items 1, 2, 5, 7).
  14:// Item 1: the already-initialized probe must run BEFORE the wizard, so an
  53:// Item 2: --bundles catalog validation lives in validateInitFlags (PreRunE).
  78:// Item 5: development_mode is not a wizard question on the init path (TDD is
  117:// Item 7: the --enable-lsp flag is registered with default true. The usage string
  ```
  머리 주석의 항목 목록이 파일 안 Item 표제와 정확히 일치한다.

### C3. F4·F5·N6 수리와 테스트가 유지된다 — 확인
- 코드 변경 자체는 1회차 이후 그대로다: `git diff --stat 388967b934 f5dbf380ec…`가 init.go·update_wizard.go·questions.go·sync.go와 세 테스트 파일을 포함하지 않는다(C1 되돌림 stat). 1회차의 RED 실측(수리 전 코드에서 F4·F5·N6 새 테스트 실패)은 같은 코드에 대한 것이므로 유효하다.
- E3a: `timeout 300 go test -count=1 ./internal/cli/wizard/` → `ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	9.008s`
- E3b: `timeout 300 go test -count=1 ./internal/profile/` → `ok  	github.com/modu-ai/moai-adk/internal/profile	0.507s`
- E3c: `timeout 560 go test -count=1 ./internal/cli/ -run 'TestInitCmd_|TestInitBoolFlag|TestReconfigureBlank' -v`(결과 줄만 추림) →
  ```
  --- PASS: TestInitCmd_HasAutonomyTierFlag (0.00s)
  --- PASS: TestInitBoolFlagUsageDoesNotRepeatTheDefaultTheRendererPrints (0.00s)
  --- PASS: TestInitCmd_EnableLSPRegisteredDefaultMatchesHelp (0.00s)
  --- PASS: TestInitCmd_EnableLSPHelpStatesDefaultOnce (0.00s)
  --- PASS: TestInitCmd_Exists (0.00s)
  --- PASS: TestInitCmd_Use (0.00s)
  --- PASS: TestInitCmd_Short (0.00s)
  --- PASS: TestInitCmd_Long (0.00s)
  --- PASS: TestInitCmd_IsSubcommandOfRoot (0.00s)
  --- PASS: TestInitCmd_HasFlags (0.00s)
  --- PASS: TestInitCmd_NonInteractiveExecution (1.96s)
  --- PASS: TestInitCmd_PositionalArgCreatesDirectory (0.48s)
  --- PASS: TestInitCmd_DotArgUsesCurrentDirectory (0.70s)
  --- PASS: TestInitCmd_HasPage3OverrideFlags (0.00s)
  --- PASS: TestInitCmd_PlanTypeFlagRetired (0.00s)
  --- PASS: TestInitCmd_ProfilePersistence (0.69s)
  --- PASS: TestInitCmd_ProjectModeFlagRetired (0.00s)
  --- PASS: TestInitCmd_WorkflowBranchGuardFlagsRegistered (0.00s)
  --- PASS: TestReconfigureBlankGitLabURLRestoresTheDefault (0.01s)
  --- PASS: TestInitCmd_AllFlagRegistered (0.00s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/cli	5.071s
  ```
  F4(`TestInitBoolFlagUsageDoesNotRepeat…`), F5(`TestReconfigureBlankGitLabURLRestoresTheDefault`), N7 관련(`TestInitCmd_EnableLSP…`) 통과. N6(`TestDuplicateKeyUserYAMLIsNeverPartlySpliced`)은 E3b의 profile 패키지 실행에 포함되어 통과.

### C4. 형식·정적 검사 — 깨끗
- E4a: `gofmt -l internal/cli internal/profile; echo "gofmt-exit=$?"` → `gofmt-exit=0` (목록 출력 없음)
- E4b: `timeout 300 go vet ./internal/cli/ ./internal/cli/wizard/ ./internal/profile/; echo "vet-exit=$?"` → `vet-exit=0`

### C5. 커버리지(일부)
- E5: `timeout 300 go test -count=1 -coverprofile=<scratch>/prof.cov ./internal/profile/` → `ok  	github.com/modu-ai/moai-adk/internal/profile	1.403s	coverage: 85.0% of statements`; `go tool cover -func=… | grep …` →
  ```
  github.com/modu-ai/moai-adk/internal/profile/sync.go:39:		SyncToProjectConfig		82.9%
  total:									(statements)			85.0%
  ```

### C6. F6 후속 카드가 큐에 없다 — 확인(D1)
- E6: `moai todo list > <scratch>/todo-list.txt; echo "exit=$?"` → `exit=0`, 210행. `grep -n -i 'selectHeight\|select height\|t1972\|wrapped option\|REQ-ITI-015\|blank card'` → 49행 t1966(hold) 한 줄만 일치("(10) selectHeight 가 로케일 설명 길이로 계산"). F6을 넘겨받는 새 카드는 없다. t1972 자신도 이 목록 출력에는 보이지 않아(목록이 보여 주는 상태 범위 때문으로 추정, 미확인), 카드 원문은 리더 지시문의 요약으로만 대조했다.

### C7. 판정 파일 추적 상태(D3)
- E7: `git check-ignore --no-index -v .moai/reports/t1972/sync-audit.md .moai/reports/t1972/progress.md` →
  ```
  .gitignore:235:.moai/reports/*	.moai/reports/t1972/sync-audit.md
  .gitignore:235:.moai/reports/*	.moai/reports/t1972/progress.md
  ```
  `git ls-files .moai/reports/t1972/` → `.moai/reports/t1972/sync-audit.md`(progress.md는 미추적), `git ls-files .moai/reports | wc -l` → `1992`.

### F6 분리 결정에 대한 판단 (지시 항목 2)
- 카드 본문의 수용 기준은 "재현(RED) 먼저, 그다음 최소 수리; 재현이 안 되면 고정 테스트와 근거"다. F6은 재현된 경우라서 이 문장대로라면 수리가 필요하고, "재현했지만 미수리·분리"는 본문에 없는 세 번째 결말이다. 따라서 이 카드는 범위를 좁힌 채로 닫히는 것이다.
- 그래도 차단으로 보지 않는 이유: (1) 1회차 감사가 요구 수정 선택지로 (c) 분리·되돌림을 명시했고, (2) t1966이 같은 설계(로케일 최댓값 + 고정 폭 줄 계산)를 같은 이유로 감사 뒤 되돌린 선례가 있으며(커밋 5c3f0ebf27), (3) 수리 정지 규칙상 같은 계보의 재수리 요청보다 분리가 맞는 처방이고, (4) 결정이 진행 기록에 남아 있다.
- 카드가 이렇게 닫히려면 필요한 것: F6 후속 카드 발행(D1), 이 카드 종결 기록에 분리 사실과 후속 카드 번호, 1회차 재현 증거(로케일별 부족 행수, base 골든의 가려진 3행)를 후속 카드로 넘기기. 이 셋이 없으면 F6이 큐 어디에도 없게 되어 결함이 조용히 사라진다.

### Baseline-attribution
- 모든 명령은 이번 실행에서 트리 f5dbf380ec368428eedd7c77ad5fd8e12e980944(HEAD 확인: `git rev-parse HEAD` → `f5dbf380ec368428eedd7c77ad5fd8e12e980944`)를 대상으로 수행했다.
- 시험은 트리 소스를 `go test`가 직접 컴파일한 것이라 설치본 moai 바이너리와 무관하다. `moai todo list`만 설치본(MCP 서버 빌드 646955327e와 같은 rc.42로 추정, 미확인)을 썼고, 이것은 큐 읽기이지 트리 측정이 아니다.
- F4 도움말 렌더 실측과 RED 대조(overlay)는 1회차(388967b934)의 것을 이어 쓴다. 해당 소스 파일이 그 뒤 바뀌지 않았음을 C1 되돌림 stat으로 확인했다.

## Gaps
- 교차 모델 단계는 의도적으로 반복하지 않았다. 운영자의 카드당 1회 codex_audit 허용은 1회차에서 썼고(receipt `rcpt-88eb22e66235cd24de99b525`), 지시에 따라 codex_audit·audit_multi·glm·claude_audit와 모델을 부르는 `moai verify`를 호출하지 않았다. receipt 가드가 이 판정을 거부하더라도 우회하지 않는다.
- 실행하지 않음: 전체 `internal/cli` 스위트, `-race`, golangci-lint, 다른 OS, internal/cli 커버리지, 후보 바이너리 재빌드(도움말 렌더는 1회차 실측 인용).
- 지정된 `-run` 정규식은 1회차와 달리 `TestValidateInitFlags|TestRunInit_`를 뺀 형태로 실행했다(지시문 그대로). 그 시험들이 다루는 init.go 경로는 1회차 이후 바뀌지 않았다.
- t1966 원 판정문은 디스크에 없어 N7이 가리킨 주석은 여전히 확정할 수 없다. 카드 원문은 큐 목록에서 찾지 못해 리더 지시문 요약에 기댔다.
- 거부된 도구 호출 없음.

## Residual-risk
- 후속 카드가 발행되지 않으면 F6(80열 영어 2쪽에서 옵션이 가려짐, ko/ja 실제 질문 높이 1~2행 부족)이 큐 밖으로 사라지고, base 골든이 그 결함을 정답으로 계속 잠근다.
- F5 새 분기는 대화형 경로에서 도달하지 않는다(1회차 C2). 향후 다른 호출자가 빈 URL을 "유지"의 뜻으로 넘기면 저장된 셀프호스팅 URL이 gitlab.com으로 덮인다.
- N6 검사는 `map[string]any` 파싱 기준이라 yaml.v3는 받아들이지만 yamlpatch가 잘못 다루는 형태(앵커·별칭)는 검사되지 않았다.
- 스크래치의 `todo-list.txt`·`prof.cov`는 기계 로컬이다. 판정을 가른 줄은 위 Evidence에 옮겼다.

## 반복 이력
- 1회차: FAIL 0.74 @388967b934 — blocking F6-R1/R2(selectHeight 로케일 최댓값·고정 74열 → 빈 카드 줄, REQ-ITI-015 위반). 파일 `.moai/reports/t1972/sync-audit.md`.
- 2회차(이 문서): PASS-WITH-DEBT 0.85 @f5dbf380ec — F6 되돌림 확인, F6-R1/R2 소멸(분리), N7-R1 처리. 되돌림 diff에 수리에서 생긴 새 blocking 결함 없음(수리 정지 규칙 해당 없음).

## Operational Notes (unverified)
- 착지 전에 리더가 F6 후속 카드를 발행하는 것이 D1의 해소 조건이다. 카드 본문에 1회차 E7·E8 수치와 "base 골든의 빈 3행은 결함"이라는 점을 함께 적으면 그 카드의 RED가 바로 선다.
- 판정 파일이 둘(`sync-audit.md` FAIL, `sync-audit-2.md` PASS-WITH-DEBT)이다. 증거 재결속 도구가 파일 이름의 `-iter` 번호로 최신 판정을 고른다는 기억 기록이 있다(`-iter` 없는 이름은 1로 취급). 이 파일 이름은 `-iter` 형식이 아니므로, 재결속 시 1회차 FAIL 파일이 뽑힐 수 있다. 필요하면 리더가 `sync-audit-iter2.md` 같은 이름으로 디스크에서 바꾼다.
- F5-R2(저장 URL 미리 채우기)는 데이터 손실성이라 별도 카드 후보로 운영자에게 보일 것.
