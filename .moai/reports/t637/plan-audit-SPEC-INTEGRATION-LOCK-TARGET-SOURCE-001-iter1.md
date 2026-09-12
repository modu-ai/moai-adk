# SPEC Review Report: SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001
Iteration: 1/2 (Tier M 상한 = 2, `harness.yaml` `plan_audit_tier_ceilings.M`)
Verdict: FAIL
Overall Score: 0.74 (네 차원의 조화평균 — Tier M PASS 기준 0.80 미달)

- 감사 대상 트리: 워크트리 `.claude/worktrees/t637`, 브랜치 `WT-acquire-branch-record`, HEAD `a3b913b85`
  (`git diff --stat 1ad0fdc09 HEAD` = SPEC 4개 파일만 추가. 따라서 spec.md §B의 `1ad0fdc09` 행 번호는 HEAD에서도 유효)
- 입력: spec.md, plan.md, acceptance.md, progress.md (Tier M 3종 + progress), 증거 `.moai/reports/t637/verdict.md`
- Reasoning context ignored per M1 Context Isolation. 판정 근거는 SPEC 파일, 증거 보고서, 그리고 이 감사에서 직접 실행한 명령뿐이다.
- 교차 모델 감사: 실행하지 않았다. `.moai/config/sections/`에 `audit_model` 키가 없고(`grep -rn audit_model .moai/config/sections/` → 무출력), codex/GLM 백엔드는 diff를 검토하는데 이 SPEC 커밋은 미커밋 변경이 없으며 `baseBranch`는 원격 기본 헤드(main) 기준이라 SPEC이 아닌 develop↔main 전체 차이를 보게 된다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: spec.md:87-130에 REQ-ILT-001…013이 빈 번호·중복 없이 연속, 3자리 0 채움 일관.
- [PASS] MP-2 GEARS 형식 (요구 계층 `REQ-XXX`만 판정): 001/002/004 `While …, … shall`(L87-99), 003/005 `When …`(L93, L100), 006 `Where … when …`(복합, L101), 007 `Where …`(L102), 008/009/011/012/013 Ubiquitous·`shall not`(L103-130), 010 Ubiquitous(L116). IF/THEN 없음. 007의 "shall emit no warning"은 `shall not emit`과 의미가 같아 허용. AC 계층의 Given-When-Then(acceptance.md §D.2)은 검증 계층의 올바른 형식이라 여기서 판정하지 않았다.
- [PASS] MP-3 YAML frontmatter: spec.md:2-13에 `id`, `title`, `version: "0.1.0"`(따옴표 semver), `status: draft`, `created/updated: 2026-09-11`, `author`, `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags`(쉼표 문자열) 12필드 전부 + `tier: M`. 거부되는 별칭(`created_at` 등) 없음.
- [N/A] MP-4 언어 중립성: Go 코드베이스 내부 CLI 한정 SPEC. 템플릿 편집(kanban-dispatch 한 줄)에는 프로그래밍 언어가 등장하지 않는다.
- [PASS] MP-5 D7: 세 문서에서 참조한 SPEC ID 추출 → `SPEC-INTEGRATION-LOCK-ATOMIC-001 status: completed`, `SPEC-INTEGRATION-LOCK-LIVENESS-001 status: completed`, 자기 자신 `draft`. retired/superseded/archived 없음, 미존재 참조 없음. BLOCKING 없음.
- [PASS] MP-6 D8: `grep -c syscall` → 네 파일 모두 `0`. 교차 플랫폼 우려 없음.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' <SPEC dir>` → 무출력(rc=1). research.md는 없다(Tier M 비필수).

보조 증거: `moai spec lint spec.md` → `0 error(s), 0 warning(s)`, INFO 1건(`OwnershipTransitionUnmeasured` — D13). `spec_audit`(project_root=워크트리) → INFO `EraAutoDetected` 1건, drift 없음.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 요구 문장 자체는 대부분 단일 해석(spec.md:87-130). 그러나 plan.md:169 예시 문구가 AC-ILT-010과 충돌(D5), REQ-ILT-010의 "Every documented"가 AC 범위(5개 파일)보다 넓음(D6), 창 획득이 거절될 때 경고를 내는지 미정(D10), plan §D "stay unedited"와 §G의 테스트 파일 편집 목록이 어긋남(D8) |
| Completeness | 0.90 | 1.0 | HISTORY(L19), 배경 §A(L25), 전제 §B, 요구 §C, 제약 §D, `### Out of Scope — …` H3 3개와 `-` 항목(L146-169), 추적표 §F. 누락은 검증 범위에 한정됨(internal/template·internal/config 테스트, D7/D8) |
| Testability | 0.55 | 0.50 | 판정 명령 여러 개가 문자 그대로는 실행되지 않거나 공허하게 통과한다. `$EV`·`$BIN` 미정의(D1), 표 안의 `\|`로 0건 선택·0건 매치(D2, 실측), mutation row b의 대상 테스트가 구현에 따라 판별하지 못함(D3), 대소문자 불일치(D5) |
| Traceability | 0.90 | 1.0 → 0.75 사이 | 모든 REQ에 AC가 있고 모든 AC가 존재하는 REQ를 가리킨다(spec.md:173-185, acceptance.md:56-69 대조). 다만 REQ-ILT-006/007의 "manual mode" 절반과 REQ-ILT-003의 "non-blank" 조건에는 AC도 mutation row도 없다(D4, D15) |

조화평균: 4 / (1/0.75 + 1/0.90 + 1/0.55 + 1/0.90) = 0.744.

## 사용자 지정 점검 항목 판정

| 항목 | 판정 | 근거 |
|---|---|---|
| REQ↔AC 양방향 대응 | PASS (조건 절 일부 제외, D4) | 13 REQ 전부 ≥1 AC, 14 AC 전부 유효 REQ 참조 |
| AC가 구체 명령으로 기계 판정 가능 | FAIL | D1, D2, D5 |
| 픽스처 AC가 `CLAUDE_PROJECT_DIR=/tmp/t637-fx`를 같은 호출에서 지정, 실제 창 무접촉 | 대체로 PASS | acquire·status·release 모두 접두 지정(acceptance.md:40, 86-88, 98, 107). `integrationLockRoot()`는 이 변수를 가장 먼저 읽는다(integration.go:47). `--help`(AC-010b)는 RunE를 실행하지 않는다. 실제 lock 파일은 현재 없음(`ls …/.moai/state/integration-lock.json` → No such file), 가드는 해시 비교(acceptance.md:41-43). C3′의 status가 접두 없이 적힌 점만 보완 필요(D1에 병합) |
| stdout/stderr 분리 주장 | PASS (주장 사실) | `runIntegration`은 `cmd.SetOut(&out)`과 `cmd.SetErr(&out)`에 같은 버퍼를 준다(integration_lock_cli_test.go:32-34). 병합 헬퍼로는 row e를 판별할 수 없다는 plan.md:62-65 주장은 옳다. 분리 헬퍼로 쓰면 row e(경고를 stdout으로)는 AC-008의 "stdout은 정확히 acquired 한 줄"에서 실패하므로 판별한다. 경고의 writer가 `cmd.ErrOrStderr()`이어야 버퍼로 잡힌다는 점은 명시 권장(D9에 병합) |
| mutation guard 판별력 | 부분 FAIL | row a/c/d/e/f는 이름 붙은 테스트가 판별한다. row b는 판별이 구현 방식에 달려 있다(D3). 복합 조건의 manual-mode 절반에는 mutant가 없다(D4) |
| 구 레코드 호환 | PASS | AC-ILT-009가 키 없는 JSON을 직접 써서 텍스트·JSON·파일 바이트 불변을 단정. 역방향(새 레코드를 구 바이너리가 읽음)도 `ReadIntegrationLock`이 `json.Unmarshal`(DisallowUnknownFields 없음, integration_lock.go:186 부근)이라 안전 — SPEC이 명시하지 않았을 뿐 |
| 템플릿 중립성 | PASS (검사 명령은 D2 보완 필요) | 편집은 `<card-id>` 자리표시자뿐. 로컬↔템플릿 `diff -q` → rc=0(현재 동일). 템플릿 트리 안의 acquire 호출은 이 한 줄뿐(`grep -rn 'integration acquire' internal/template/templates`) |
| Tier M·파일 수 11 | PASS | 코드 3·테스트 3·문서 5 확인. 다른 사본 없음(`grep -rln 'Serialize by the recorded hold'` → 두 경로뿐). hns-release-specialist.md는 `.claude/agents/harness/`라 codex 방출(C2→C3) 대상 아님. 11개 파일, 약 270-450 LOC → spec-workflow.md:141 Tier M 대역(5-15 files) |
| catalog.yaml에 rules 항목 없음 | PASS | `grep -n rules internal/template/catalog.yaml` → 무출력. skills 항목만 있다 |
| 네 항목 대비 범위 초과·누락 | PASS | ① REQ-001/002 ② REQ-003~008 ③ REQ-009 ④ REQ-010/011. 불변식 REQ-012/013. 범위 밖 항목(§E)이 운영자 결정과 일치 |
| CLAUDE.local.md 충돌 메모 | PASS (사실 확인) | primary 작업 사본 734줄, acquire 338/357/371. 워크트리 772줄, 370/389/403. 워크트리 사본 = 로컬 `develop` blob과 동일(diff rc=0). 보충: primary는 `main` 체크아웃이고 `main` blob은 603줄이라, primary 사본의 차이는 main 위의 미커밋 수정이다. 충돌면은 release PR 또는 main 쪽 커밋 시점이다(선택 보강, D16) |

## Defects Found

| ID | 심각도 | 분류 | 위치 | 결함 | 필요한 수정 |
|---|---|---|---|---|---|
| D1 | major | blocking | acceptance.md:38-40, 86, 98, 107-108 | 픽스처 셀의 `2>"$EV/c1.stderr"`에서 쓰는 `$EV`가 문서 어디에도 정의되지 않는다(`grep -n 'EV=' <SPEC dir>/*.md` → 무출력). Bash 호출마다 새 프로세스라 앞 호출에서 정한 값은 이어지지 않는다. 비어 있으면 `cd /tmp/t637-fx-wt/cardA` 뒤 리디렉션 대상이 `/c1.stderr`(쓰기 불가 → acquire 자체가 실행되지 않음)가 되고, 상대경로면 픽스처 트리 안에 증거가 떨어진다. `$BIN`도 "이 트리에서 빌드한 바이너리"라고만 하고 경로가 없다. C3′의 `status`/`status --json`은 접두 없이 적혀 있다. MUST-PASS AC-004(b)/005(b)/006(b)의 판정 명령이 문자 그대로는 재현되지 않는다 | §A에 절대경로를 정의하고 각 셀 명령에 같은 호출로 싣는다. 예: `EV=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637/.moai/reports/t637/fixture; BIN=<worktree>/bin/moai; mkdir -p "$EV" && cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx "$BIN" …`. C3′의 status 두 줄도 접두를 명시한다 |
| D2 | major | blocking | acceptance.md:57, 62, 65(b), 67, 68 | 판정 명령이 마크다운 표 안에 있어 `\|`로 이스케이프돼 있다. 원문을 복사하면 (1) `go test -run 'A\|B'`는 RE2에서 리터럴 `|`라 0건 선택 → `ok … [no tests to run]`로 공허한 초록(AC-002, MUST-PASS AC-007·AC-013), (2) AC-012의 `grep -cE 't[0-9]{3}\|SPEC-\|…'`는 금지 토큰이 있어도 0을 낸다(실측: `printf 'SPEC-X\nt637\nfoo\n' \| grep -cE 't[0-9]{3}\|SPEC-'` → `0`, `\|` 없는 형태 → `2`), (3) AC-010(b)는 `\|`가 셸에서 인자로 넘어가 명령이 깨진다. AC-012의 "추가된 줄 수 = 1" 가드는 입력이 비지 않았음만 보이고 패턴이 매치할 수 있음은 보이지 않는다 | 여러 테스트 이름이나 파이프가 들어가는 명령은 표 밖 fenced code block으로 옮긴다(표에는 블록 번호만). go test 판정에는 `-v`를 필수로 두고 "지명한 테스트 수 = `--- PASS: <name>` 줄 수"를 관측값으로 추가한다. AC-012에는 양성 대조를 붙인다(같은 패턴을 `t637`이 든 입력에 돌려 ≥1) |
| D3 | major | blocking | acceptance.md:62, 134; internal/cli/integration_target_test.go:252 | mutation row b(경고 조건에서 git-flow 술어 제거)가 반드시 실패시켜야 하는 테스트는 AC-007의 github-flow 셀인데, 이 셀은 `develop_branch`가 **채워진** 채로 쓰인다("develop branch present", 기존 테스트도 `"fixture-integration"`). 경고 조건을 원시 `develop_branch` 값의 공백 여부로 구현하면, git-flow 술어를 떼어도 이 셀에서는 경고가 나지 않아 mutant가 살아남는다. 판별 여부가 run-phase 재량(plan.md §D "shape is run-phase latitude")에 달려 있다. absent-config 셀은 판별하지만 row b는 그 테스트를 지명하지 않는다 | AC-007 음성 셀에 `writeGitStrategyFixture(t, repo, "github-flow", "")`(develop 비움) 케이스를 추가해, git-flow 술어 외에는 경고와 무경고를 가르는 요소가 없게 만든다. row b의 지명 테스트를 그 케이스로 바꾸고, absent-config 테스트도 row b의 실패 대상에 함께 적는다 |
| D4 | minor | blocking | spec.md:101-102; acceptance.md:149-150, 131-138 | REQ-ILT-006/007은 "manual mode"를 조건으로 명시하고 §D.4는 "personal/team 프로필 + `workflow: git-flow` → 경고 없음"을 주장하지만, 이를 판정하는 AC도 mutation row도 없다. `mode == manual` 조건을 떼는 mutant를 막는 테스트가 없다(복합 조건의 절반이 무방비) | AC-ILT-007에 `mode: personal` + git-flow + 빈 develop 케이스(경고 없음, `branch_source == "caller"`)를 추가하고, §D.3에 row g "경고 조건에서 `mode == manual` 제거 → 해당 테스트 실패"를 추가한다. 이 케이스에는 `writeGitStrategyFixture`가 mode를 `manual`로 고정하므로 별도 fixture 문자열이 필요하다 |
| D5 | minor | blocking | plan.md:169-171 vs acceptance.md:65 | plan M5의 예시 문구 "Integration target branch — …"는 대문자 I라서, AC-010(a)의 대소문자 구분 `integration target` 포함 검사와 (b)의 grep 모두에서 실패한다. 계획과 수용 기준이 서로 모순된다 | 예시 문구를 소문자 `integration target branch (…)`로 맞추거나(verdict.md §5 C안 3번 문구와 일치), AC-010의 검사를 대소문자 무시(`strings.Contains(strings.ToLower(usage), …)`, `grep -i`)로 명시한다. 둘 중 하나만 고른다 |
| D6 | minor | optional | spec.md:116-118; internal/hook/integration_lock_guard.go:96, 100 | REQ-ILT-010은 "Every documented lane-level acquire invocation"으로 전칭이지만 AC-011은 5개 파일만 센다. 가드가 레인에게 출력하는 실행 지시(`reclaim with \`moai integration acquire\``, `\`moai integration acquire --force\``)도 `--card`가 없는 레인 대상 호출인데, REQ-013·AC-013은 가드 diff를 비워 두라고 요구한다. 대부분은 "문서화된"을 문서로 읽겠지만 문장상 모순 여지가 있다 | REQ-010을 "spec.md §B 전제 7이 열거한 문서 호출"로 한정하거나, §E에 "가드·CLI 런타임 메시지 안의 acquire 안내문은 범위 밖" 항목을 추가한다 |
| D7 | minor | optional | spec.md:134-137; internal/template/rule_template_mirror_test.go:137 | §D는 검증 범위를 cli/kanban/config/hook으로 적었지만 이 SPEC은 `internal/template`(임베드되는 템플릿)도 건드린다. 그 패키지에는 이 편집을 직접 지키는 `TestRuleTemplateMirrorDrift`(바이트 동일 강제)와 중립성·누출 감사 테스트가 있다 | §D의 패키지 목록에 `internal/template`를 추가하고, AC-012에 `go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift' -count=1 -v`(및 중립성 감사 테스트)를 넣는다 |
| D8 | minor | optional | plan.md:103-105, 159, 185; acceptance.md:68 | plan §D는 `loader_integration_branch_test.go`가 "unedited"로 남는다고 하면서 §F M1과 §G는 같은 파일에 seam 테스트를 추가한다고 적는다. 어느 AC도 `go test ./internal/config/...`를 실행하지 않으므로, 새 seam 테스트와 기존 `TestLoadGitFlowDevelopBranch`(loader_integration_branch_test.go:18)가 초록인지 판정되지 않는다 | §D 문구를 "기존 테스트 함수는 수정하지 않고, 새 테스트만 덧붙인다"로 고친다. AC-013에 `go test ./internal/config/... -run 'TestLoadGitFlowDevelopBranch\|<seam 테스트명>' -count=1 -v`를 추가한다(D2 규칙대로 표 밖에 둔다) |
| D9 | minor | optional | plan.md:93-95; acceptance.md:68; internal/cli/integration_target_test.go:134, 147, 160, 173 | plan §C는 "resolution function reports which tier won"이라고 한다. 반환값이 늘면 기존 호출 4곳이 컴파일되지 않는데, AC-013은 기존 테스트 편집을 "additive"로만 허용한다. `branch, wt, _ :=` 같은 기계적 적응은 추가가 아니다. 경고 writer(`cmd.ErrOrStderr()` 대 `os.Stderr`)도 미지정이다 | 기존 호출부의 기계적 시그니처 적응을 허용하고 §E.2에 기록하도록 명시하거나, 기존 함수는 두고 형제 함수를 추가하도록 plan §C를 고친다. 경고는 `cmd.ErrOrStderr()`로 쓴다고 plan §B1에 적는다 |
| D10 | minor | optional | spec.md:101, 103-105 | 창이 이미 다른 레인에게 잡혀 있어 `acquire`가 거절될 때 경고를 내는지 정해져 있지 않다. REQ-006의 "names the recorded caller branch"는 기록 이후를 암시하지만, 해석은 브랜치 결정 시점에 일어난다(integration.go:246 → 255) | 경고는 창 기록에 성공한 뒤에만 낸다(또는 해석 직후 항상 낸다) 중 하나를 REQ-006에 명시한다 |
| D11 | minor | optional | acceptance.md:66 | AC-011의 "정확히 8줄"은 run-phase가 로컬 develop을 흡수해 CLAUDE.local.md가 바뀌면 깨진다(primary 사본은 이미 다른 줄 번호를 가진다) | "8줄"을 "§B 전제 7의 8개 사이트 각각이 …을 포함"으로 바꾸거나, 흡수 후 재측정한 기준선을 §E.2에 기록하도록 한다 |
| D12 | minor | optional | acceptance.md:33-37 | 픽스처는 `/tmp`에만 있고 재구성 절차가 없다. 현재는 존재한다(`git -C /tmp/t637-fx worktree list` → main/WT-a/WT-b/develop 4개, develop_branch: develop). OS의 /tmp 정리로 사라지면 MUST-PASS 픽스처 셀 3개를 실행할 수 없다 | §A에 재구성 레시피(init, 브랜치 3개, worktree add 3개, git-strategy.yaml 작성)를 적는다 |
| D13 | minor | optional | 커밋 a3b913b85 | `moai spec lint` INFO `OwnershipTransitionUnmeasured`: 커밋에 `Authored-By-Agent` 트레일러가 없어 `(none) → draft` 소유 전이가 측정되지 않는다 | 다음 SPEC 커밋에 트레일러를 싣는다(선택) |
| D15 | minor | optional | spec.md:93-96; acceptance.md:145-146 | REQ-ILT-003의 "non-blank `--branch`" 조건과 §D.4의 `--branch "   "` 경계는 AC가 없다. 출처를 트림하지 않은 `branchFlag != ""`로 판정하는 mutant가 살아남는다 | AC-003에 공백 `--branch` 하위 테스트(`branch_source ≠ flag`)를 추가하거나, plan §C의 "출처는 브랜치를 정하는 같은 단계에서 결정" 제약을 AC 문장으로 올린다 |
| D16 | minor | optional | plan.md:129-138 | E3 메모는 사실이지만, primary가 `main` 체크아웃이고 그 차이가 main blob(603줄) 위의 미커밋 수정이라는 점이 빠져 있다. 충돌 시점(release PR / main 쪽 커밋)을 판단하는 데 필요한 정보다 | E3에 한 줄을 보탠다(선택) |

(D14는 D1에 병합했다. blocking 5건: D1, D2, D3, D4, D5.)

## Regression Check

해당 없음 (iteration 1).

## Recommendation

manager-spec에게 blocking 5건만 수정하도록 권한다. optional은 오케스트레이터 재량이다.

1. **D1** — acceptance.md §A에 `EV`와 `BIN`의 절대경로 정의를 추가하고, 세 픽스처 셀(L86, L98, L107)의 명령 앞에 같은 호출 안에서 `EV=… BIN=…; mkdir -p "$EV" && …`를 싣는다. C3′의 status 두 줄(L108)에도 `CLAUDE_PROJECT_DIR=/tmp/t637-fx` 접두를 적는다.
2. **D2** — AC-002/007/010(b)/012/013의 판정 명령을 표 밖 fenced block으로 옮겨 `|`가 이스케이프 없이 나타나게 한다. go test 판정마다 `-v`와 "`--- PASS:` 줄 수 = 지명한 테스트 수"를 관측값으로 추가하고, AC-012에는 패턴 양성 대조를 붙인다.
3. **D3** — AC-007에 github-flow + 빈 `develop_branch` 케이스를 추가하고, §D.3 row b의 지명 테스트를 그 케이스(와 absent-config 케이스)로 바꾼다.
4. **D4** — AC-007에 `mode: personal` + git-flow + 빈 develop 케이스를 추가하고, §D.3에 row g(`mode == manual` 제거)를 추가한다.
5. **D5** — plan.md M5 예시 문구와 AC-010의 대소문자 규칙을 하나로 맞춘다.

그대로 두어도 되는 부분: 필수 통과 7개 항목, 운영자 결정 네 항목의 범위 적합성, 실제 창 보호 설계(같은 호출 내 `CLAUDE_PROJECT_DIR`, 해시 가드), 구 레코드 호환 AC-009, stdout/stderr 분리 판단, 파일 수·Tier·catalog 판정. 재감사(iteration 2)는 위 blocking 5건의 변경분만 대상으로 한다.

---
Evidence (이 감사에서 실행, 트리 HEAD `a3b913b85`):
- `git diff --stat 1ad0fdc09 HEAD` → SPEC 4개 파일만 추가(170/221/47/185줄)
- `git grep -n "moai integration acquire" -- <5 files>` → 8줄, `--card` 0건
- `diff -q <local kanban-dispatch> <template kanban-dispatch>` → rc=0
- `grep -n rules internal/template/catalog.yaml` → 무출력
- `printf 'SPEC-X\nt637\nfoo\n' | grep -cE 't[0-9]{3}\|SPEC-'` → `0` / 이스케이프 없는 형태 → `2`
- `moai spec lint …/spec.md` → 0 error, 0 warning, INFO 1
- D7 verb → ATOMIC-001 completed, LIVENESS-001 completed; D8 → syscall 0; MP-7 → 무출력
- `wc -l` → worktree CLAUDE.local.md 772, primary 734, `git show main:CLAUDE.local.md` 603, `git show develop:CLAUDE.local.md`와 워크트리 사본 동일(rc=0)

Gaps: internal/cli 테스트와 픽스처 셀은 실행하지 않았다(컴파일 슬롯은 리드 승인 사항, 지시에 따름). 따라서 픽스처 셀의 stderr가 경고 외에 다른 줄을 포함하지 않는지(AC-005(b) `grep -c '' = 1`)는 관측하지 않았다. RE2의 `\|` 해석은 문서화된 RE2 문법에 근거했고 go test로 실측하지는 않았다(grep 경로만 실측).
Residual-risk: 바이너리가 다른 stderr 안내(버전 지연 경고 등)를 내면 AC-005(b)/006(b)의 "정확히 1줄"이 경고와 무관하게 실패할 수 있다.
