auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-PRECOMMIT-AFFECTED-SCOPE-001 (iteration 2, delta)

verdict: FAIL
audited_sha: deb50abacc90089aef80e0f856e744e76da24d53
overall_score: 0.875
must_pass_failed: 0
blocking_count: 4
plan_artifact_hash: 03d018748f51e5b17309a109278746cdd82a503c8d7ae38021acde7c955ef0c2
scope: delta
fix_scope: .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#LEDGER-PAS-009-A .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#AC-PAS-016 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#LEDGER-PAS-016-A .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md#Definition-of-Done .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe009.sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe016.sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/plan.md#B.2 .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/spec.md#REQ-PAS-009
defect_class: N1=ac-wording N2=ac-wording N3=ac-wording N4=ac-wording
convergence_overall: fail
required_backend: codex fail
Iteration: 2/3
Auditor Version: plan-auditor/v1 (sonnet-5-5)

Reasoning context ignored per M1 Context Isolation.

## 판정 요약

iter1의 blocking 결함 D1~D7, D9는 모두 문서 수준에서는 해소되었습니다. 필수 통과 항목은 전부 통과이고 종합 점수는 0.875로 Tier M 임계값 0.80을 넘습니다. 그런데도 FAIL입니다. 이번 수정이 새로 도입한 결함 4건을 제가 직접 확인했고, 필수 백엔드 codex도 FAIL을 냈기 때문입니다. 새 RED-now 증거 셀 두 곳은 지금 트리에서 실제로 빨강이지만, 그 빨강을 구현이 초록으로 되돌릴 수 있는 경로가 문서에 없습니다(N1, N2). 같은 문서 안에서 argv 형식(N3)과 `--amend` 의미(N4)도 서로 어긋납니다.

judging build: 이 감사에서 HEAD deb50abac의 트리로 `go build ./cmd/moai`한 `/private/tmp/t1673-audit/bin/moai2`(ldflags 없음, 소스는 HEAD와 동일). `git diff --stat 3642a3a9b deb50abac -- internal cmd go.mod go.sum` 출력이 비어 있어 Go 코드는 iter1 트리(2febc5c2f)와도 같습니다. 설치본 `moai`와 MCP 서버는 빌드 `ee981a932`(HEAD의 조상)이며 `moai verify audit-plan`·`audit_multi`가 이 빌드로 동작했습니다.

## Must-Pass Results

- [PASS] MP-1. `### REQ-PAS-001`~`015` 연속·중복 없음, 012는 철회 자리. 추적 스크립트: `COLLECTED: 15 REQ definitions (acceptance input: read)`, UNCOVERED/ORPHAN 줄 없음.
- [PASS] MP-2 (요구 계층만). REQ-PAS-011이 기본값(Ubiquitous)과 값 검증(Where)으로 분리되어 iter1 D12 해소(spec.md diff). REQ-PAS-003~005, 007~009, 014, 015도 문장 구조는 GEARS/레거시 EARS 범위입니다. 다만 REQ-PAS-009의 `--amend`는 N4 참조.
- [PASS] MP-3. frontmatter 12필드 그대로(diff 범위 밖), `moai spec lint SPEC-PRECOMMIT-AFFECTED-SCOPE-001 --json` 출력은 info 한 건(`OwnershipTransitionUnmeasured`)뿐이며 린터의 침묵은 교차 근거로 쓰지 않습니다.
- [N/A] MP-4. 단일 언어(Go) 한정.
- [PASS] MP-5, MP-6, MP-7. 참조 SPEC 상태·`syscall`·미해결 마커는 iter1과 같은 결과(이번 diff는 해당 참조를 추가하지 않음). 마커 grep은 iter1 이후 재확인하지 않고 diff를 읽어 신규 마커가 없음을 확인했습니다.
- [PASS] MP-8. 릴리스 차단으로 분류된 AC는 이제 001, 005, 009, 013, 016 다섯 개이고(acceptance.md:L11,15,19,23,26), 각각 4요소 RED-now 셀(명령·원문 stdout·종료코드·트리 SHA)에 연결됩니다. 세 레저 명령을 이 감사에서 재실행해 RED가 재현됐습니다(Evidence E1~E3). 이 판정은 "RED가 재현되는가"만 말하며, 아래 N1·N2가 보여 주듯 그 빨강이 구현으로 초록이 되는지는 MP-8이 확인하지 않는 영역입니다.
- [PASS] MP-9. CN-4 출력 `COLLECTED: 6 milestones in plan order (M1 M2 M3 M4 M5 M6), 0 exit bindings, 8 ordering candidates`, CONFLICT 줄 없음. 후보 8건을 읽었고 마일스톤 순서와 충돌하는 순서 의무는 없습니다(P1의 "구현 전 RED 기록"은 M1~M6 어느 순서와도 양립). `Exit:` 묶음이 없어 기계적 CONFLICT 검출은 불가능함을 Gaps에 둡니다.

## Policy P1에 대한 판정 (조정자 질문 a)

질문: AC-002/003/004/006/011을 "adoption-pending"으로, AC-007/008을 regression-guard로 돌린 P1이 MP-8을 충족하는가.

판정: MP-8의 문언은 "릴리스 차단으로 분류된 AC"에만 RED-now 셀을 요구하므로, 이 다섯 AC를 릴리스 차단에서 뺀 시점에 MP-8은 이들을 대상으로 삼지 않습니다. 따라서 P1은 MP-8을 문언상 충족합니다. 그러나 두 가지를 분명히 적어 둡니다.

1. `adoption-pending`은 verification-completeness.md §2.1이 정의한 분류(릴리스 차단, 회귀 가드)에 없는 제3의 분류입니다. §2는 RED-now 셀이 없는 기준을 "미채택"이라 부르므로, 이 SPEC은 사실상 "미채택 기준 5개를 DoD가 run phase 증거로 채우게 한다"고 선언한 셈입니다. 이는 합법적이지만 REQ-PAS-003, -004, -006, -011에는 채택된(RED-now 보유) AC가 하나도 남지 않습니다. 이 요구들의 판별력은 run phase에 가서야 확정됩니다.
2. P1이 내세우는 이유("지금 가능한 빨강은 빈 sweep의 초록뿐", LEDGER-PAS-VACUOUS)는 사실과 다릅니다. 같은 현재 트리에서 동작 수준의 진짜 빨강이 최소 세 가지 있습니다(E4): docs만 바꾼 커밋과 테스트 파일만 바꾼 커밋 모두 게이트가 `go test ./...`를 실행합니다(AC-006 후반부 "명시적 skip", AC-004 "`{a}`만"에 대한 빨강). `test_scope.mode: bogus`를 담은 `gate.yaml`로 `moai gate`가 종료코드 0으로 통과합니다(AC-011 "bogus 거부"에 대한 빨강).

P1을 더 단단하게 만들 조건(필수는 아님, 아래 S1): AC-004·AC-006(후반)·AC-011에 위 E4 프로브를 RED-now 셀로 붙여 `adoption-pending`을 3개 줄이고, AC-002·AC-003만 순수 단위 수준으로 남깁니다. AC-007·AC-008을 회귀 가드로 돌린 것은 수용합니다. 지금 전체 범위 `./...`이 두 요구의 안전한 결과와 같으므로 "오늘 초록"이 구조적으로 맞고, 그 가드를 깨는 변이(모듈 필터가 trigger보다 앞섬, 영(0) 종료코드 신뢰)가 각 AC에 명시돼 있습니다.

## Defect Resolution (조정자 질문 b, iter1 D1~D7·D9)

| iter1 결함 | 판정 | 근거 |
|------------|------|------|
| D1 RED-now 부재 | 해소(MP-8), 단 신규 N1·N2 | 릴리스 차단 5개 모두 셀 보유, 재실행으로 RED 재현. P1 정책은 위에서 판정 |
| D2 git 어댑터 rename/delete | 해소 | REQ-PAS-005가 rename 감지 끔을 명시(spec.md diff), AC-PAS-005에 실제 git 어댑터 사례와 변이(기본 이름 목록이 `a/x.go`를 누락) 추가, AC-PAS-009 (iv) |
| D3 testdata `.go` 소유 | 해소 | REQ-PAS-003이 testdata 규칙을 확장자 분기 앞에 둠, plan B.2 step 2(a), AC-PAS-003의 `p/testdata/fix.go`, AC-PAS-004의 `a/testdata/f.go` |
| D4 불완전 그래프 | 해소 | REQ-PAS-008에 "errored or incomplete" 추가, plan step 4a, AC-PAS-008 세 번째 사례가 실제 `go list` 어댑터 사용. 변이 "영(0) 종료코드 신뢰" 명시 |
| D5 widening이 모듈 필터 뒤 | 해소 | REQ-PAS-007 "모든 파일 버리기 전", plan step 0, AC-PAS-007 nested 레이아웃 사례 |
| D6 AC-PAS-014 timeout 변이 | 해소(문서) | AC가 실행기 인자 값(30/60/120/300 s) 단언 + `2*g.config.TestTimeout` 변이를 명시적으로 겨냥. `gate.go:619` 호출 지점은 iter1에서 확인한 그대로이며 이 diff가 코드를 바꾸지 않았음 |
| D7 AC-PAS-003 모순 | 해소 | Given이 AC-001 그래프에 `p`, `pp`, `q`, `r`을 명시하고 기대 `{p, q, r}`가 REQ-PAS-003·004와 일치. 사소한 보완은 M1 |
| D9 AC-PAS-015 가변 상태 | 해소 | 첫 절(현재 `core.hooksPath` 값) 삭제, "이 AC는 공유 설정의 실시간 값을 단언하지 않는다"(acceptance.md:L126) |
| D8, D10~D14 | D8=AC-009에 worktree·amend 반영(단 N4), D10=spec.md A.2/Out of Scope·R2·runbook 3에 "추론, 미측정" 반영, D11=F6 수정(설명 "자체 외부 테스트 패키지가 import" 부분은 `internal/config` 자기 자신이 48번째였다는 뜻으로 읽히며 내 iter1 집계 47과 모순 없음), D12=REQ-PAS-011 분리, D13=`EmbedFiles` 사용으로 plan 수정, D14=대칭 가드 한 줄 추가 | 모두 diff에서 확인 |

## Category Scores

| 차원 | 점수 | 밴드 | 근거 |
|------|------|------|------|
| Clarity | 0.75 | 0.75 | plan B.2 step 6 "import paths"와 AC-016·레저의 `./a ./b` 충돌(N3), REQ-PAS-009 `--amend` 문구와 AC (v) 불일치(N4) |
| Completeness | 1.00 | 1.0 | 필수 절·`### Out of Scope — <topic>` 3개 유지, 불완전 그래프 처리 추가 |
| Testability | 0.75 | 0.75 | D6·D7·D9 해소. 반면 AC-016과 두 프로브의 초록 경로가 도달 불가능(N1·N2). 순수 단위 AC는 P1에 의존 |
| Traceability | 1.00 | 1.0 | 추적 스크립트 UNCOVERED 0, ORPHAN 0, 철회 자리 012/012/M4는 서로 짝지어 유지. 클래스 수 5/5/5(progress.md)가 매트릭스 표와 일치 |

종합 = (0.75 + 1.00 + 0.75 + 1.00) / 4 = 0.875 (Tier M 임계값 0.80 이상). 요구 14 live + 1 철회, AC 15 live + 1 철회로 상한 16/16 안입니다.

## Defects Found (이번 수정이 도입했거나 남긴 것)

N1. LEDGER-PAS-009-A의 초록 경로가 도달 불가능하고 DoD 2가 그 불가능을 요구 — acceptance.md:L158-178(LEDGER-PAS-009-A), L239(DoD 2), evidence/probe009.sh:L14-15 — 프로브는 기존 도우미의 질의(스크럽된 `git diff --cached --name-only`)를 셸로 고정 실행하고 `--no-renames` 질의와 비교한다. 구현이 어댑터를 어떻게 바꿔도 이 비교는 바뀌지 않으므로 레저가 말하는 "M3가 뒤집으면 `[same]`·exit 0"과 DoD 2의 "`probe009.sh` exits 0"은 성립할 수 없다. 빨강은 맞지만(E2) §2가 말하는 "영원히 빨강인 기준"에 해당한다(codex P1과 일치, 코드 읽기로 확인). — Severity: major — Class: blocking — Required fix: 레저의 green path와 DoD 2에서 probe009를 초록 조건으로 쓰는 문장을 삭제하고, 초록 경로를 `TestAffectedScope_ChangedSetMatchesCommit`(실제 어댑터 호출)으로만 정의한다. 프로브는 "구현 전 현상 기록"으로 명시한다.

N2. AC-PAS-016과 probe016.sh의 초록 경로가 가짜 `go` 때문에 도달 불가능 — acceptance.md:L129-134, L138-156, evidence/probe016.sh:L9-13 — 가짜 `go`는 모든 호출을 로그에 한 줄 쓰고 빈 stdout·종료코드 0으로 끝낸다. plan B.2는 같은 게이트가 그래프 질의로 `go list -e -json ./...`를 한 번 실행한다고 하므로, 구현 후에는 로그의 첫 줄이 `list ...`가 되고 그래프는 비어 있어 불완전 그래프 fallback(`./...`)이 발동한다. 프로브는 `sed -n 1p`로 첫 줄만 판정한다. AC-016 (i)의 "recording `go` on PATH ... argv is `test ./a ./b`"도 같은 문제를 갖고, (ii)의 "b의 테스트가 실패"는 기록용 `go`로는 만들 수 없다. 또한 프로브는 커밋 종료코드를 출력만 하고 판정하지 않는다(codex P2: 훅이 거부해도 GREEN). — Severity: major — Class: blocking — Required fix: 가짜 `go`는 `list` 하위 명령을 실제 `go`에 위임하고 `test` 호출만 기록하도록 바꾼다. 프로브·AC의 GREEN 조건에 "커밋 종료코드 0과 HEAD 변경"을 넣고, (ii)는 실제 `go test`로 실패 테스트를 실행하는 형태로 정의한다.

N3. 게이트가 `go test`에 넘기는 인자 형식이 plan과 AC에서 충돌 — plan.md B.2 step 6("Hand the remaining import paths") vs acceptance.md:L132(AC-016 (i) `test ./a ./b`), L154-156(레저 RESULT 비교 문자열) — plan대로 구현하면 `test example.com/m/a example.com/m/b`를 만들고 릴리스 차단 AC-016 (i)과 프로브가 이를 빨강으로 판정한다(codex P2와 일치). — Severity: major — Class: blocking — Required fix: 한 형식을 고른다. 모듈 루트 상대 `./a ./b` 형이 중첩 모듈에서도 안정적이면 plan step 6과 REQ에 그 형식을 명시하고, 아니면 AC와 레저를 import path 형식으로 바꾼다.

N4. REQ-PAS-009의 `--amend`와 AC (v)/plan B.3의 의미가 다름 — spec.md REQ-PAS-009("the set the commit will record, including ... `--amend`"), plan.md B.3("amended commit newly records ... not re-tested, accepted behaviour"), acceptance.md:L87-88(AC (v) "exactly the newly staged file") — "the commit will record"를 문자 그대로 읽으면 amend는 이전 커밋의 파일도 기록하는 커밋이므로 AC (v)와 어긋난다. 수용된 증분 검사 예외가 REQ에 없다(codex P2와 일치). — Severity: SHOULD-FIX — Class: blocking(AC의 기대값이 REQ에서 도출되지 않음) — Required fix: REQ-PAS-009에 "`--amend`의 경우 amend되는 HEAD 대비 인덱스에 새로 스테이징된 변경"을 한 절로 명시한다.

선택 항목(단독 FAIL 사유 아님, M6 optional):
S1. P1의 근거 문장이 사실과 다름(AC-004, AC-006 후반, AC-011은 E4처럼 지금 동작 수준 RED가 가능). — SHOULD-FIX — Class: optional — Required fix: LEDGER-PAS-VACUOUS의 "the only available red"를 정정하고 가능하면 위 세 AC에 E4 형태 셀을 붙인다.
S2. AC-PAS-003 Given에 새 패키지 `p`, `pp`, `q`, `r`이 테스트 파일을 가진다는 문장이 없어 REQ-PAS-006의 drop 규칙과 겹쳐 읽힐 수 있다. — MINOR — optional — Required fix: "all with test files"를 추가.
S3. LEDGER-PAS-016-A·013-A의 Command에 `<moai built from the tree under measurement>` 자리 표시자가 있어 그대로 재실행할 수 없다(`<`, `>`는 MP-8이 금지하는 메타문자와 같은 글자). — MINOR — optional — Required fix: 구체 경로(예: `bin/moai`)를 쓰고 빌드 명령을 앞 줄에 둔다. 이번 감사는 자리 표시자를 직접 치환해 실행했다.
S4. AC-PAS-001의 RED-now가 LEDGER-PAS-016-A(REQ-001 미구현 사유)에 묶여 있는데 AC-001이 가르는 직접/전이 의존자 구분은 그 프로브가 판별하지 못한다(그래프에 전이 사례 없음). 전이 폐포 구현도 프로브를 초록으로 만든다. — MINOR — optional — Required fix: 레저에 "AC-001의 판별 부분은 단위 테스트의 RED가 담당"임을 한 줄 적는다.

## Regression Check

- D1: 해소(MP-8 문언). 신규 N1·N2가 새 셀의 초록 경로를 막는다.
- D2: 해소. D3: 해소. D4: 해소. D5: 해소. D6: 해소(문서). D7: 해소. D9: 해소.
- iter1 선택 항목 D8, D10~D14: 반영됨.
- 정체(stagnation) 신호 없음: 이번 blocking 4건은 모두 새로 추가된 문서·프로브의 결함입니다.

## Recommendation

1. N1·N2·N3: 프로브·레저·AC-016의 초록 경로를 구현이 실제로 도달할 수 있게 고친다(위 Required fix). 이 셋은 한 번에 묶어 수정하는 것이 안전합니다(같은 파일, 같은 AC-016/레저 블록).
2. N4: REQ-PAS-009 한 절.
3. S1~S4는 같은 패스에서 처리하면 좋지만 재감사 통과 조건은 아닙니다.
4. 재감사는 위 enumerated delta와 CN-4 전체 실행으로 범위를 좁힐 수 있습니다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

C1. LEDGER-PAS-013-A, -009-A, -016-A의 RED가 현재 트리에서 재현된다.
C2. probe009는 기존 도우미의 질의를 충실히 복제한다(소스 대조).
C3. probe009와 probe016은 구현 후에도 초록이 될 수 없다.
C4. 현재 트리에서 AC-004, AC-006 후반, AC-011에 대한 동작 수준 RED가 존재한다(P1 근거 반박).
C5. 추적성·린트·CN-4는 이번 수정 후에도 깨지지 않았다.
C6. plan B.2 step 6과 AC-016의 argv 형식이 충돌한다.

### Evidence

E1 (C1, 013-A). iter1에서 같은 레시피를 실행한 결과(`exit=1`, `ast-grep suppression at .../go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line`)가 있고, 이번 diff는 Go 코드를 바꾸지 않았다. 명령 `git diff --stat 3642a3a9b deb50abac -- internal cmd go.mod go.sum` 출력이 비어 있다. 이 감사에서 013-A를 새 바이너리로 다시 돌리지는 않았다(Gaps G-b).

E2 (C1·C2, 009-A). 명령 `env PROBE_DIR=/private/tmp/t1673-audit/pas-probe sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe009.sh`, 종료코드 1, 원문 출력:
```
(ii)  git commit -a            (a/x.go staged, d/z.go modified unstaged)
  commit records : a/x.go d/z.go 
  helper form    : a/x.go    [DIFFERENT]
(iii) git commit d/z.go        (path-limited)
  commit records : d/z.go 
  helper form    : a/x.go    [DIFFERENT]
(iv)  git mv old/o.go new/o.go (rename)
  commit records : new/o.go old/o.go 
  helper form    : new/o.go    [DIFFERENT]
exit=1
```
레저의 기록과 일치한다. 충실성: 실제 도우미 `stagedFiles`(gate.go:L1390-L1404)는 `git diff --cached --name-only`를 `cmd.Env = stepEnv()`로 실행하고, `stepEnv()`는 `gitenv.Env()`로 `RepoScopingVars` 11개(`GIT_DIR`, `GIT_COMMON_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE`, `GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`, `GIT_NAMESPACE`, `GIT_PREFIX`, `GIT_SUPER_PREFIX`, `GIT_CEILING_DIRECTORIES`, `GIT_QUARANTINE_PATH`)를 제거한다(gitenv.go 출력 확인). 프로브의 `env -u` 목록이 이 11개와 동일하고 질의 인자도 같으므로 복제는 충실하다. 차이는 정렬(`sort`)뿐이며 판정에 영향이 없다. 다만 프로브는 Go 도우미를 호출하지 않는 셸 복제이므로 N1의 불가능성이 생긴다.

E3 (C1, 016-A). 명령 `env PROBE_DIR=/private/tmp/t1673-audit/pas-probe sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe016.sh /private/tmp/t1673-audit/bin/moai2`, 종료코드 1, 원문 출력:
```
graph-freshness: step disabled in gate config (explicit skip notice)

quality gate steps (3 configured):
  - go vet: disabled — turned off by the gate.disabled_steps setting
  - golangci-lint: disabled — turned off by the gate.disabled_steps setting
  - go test: executed in 203ms at 2026-10-18T10:19:38.818+09:00 — go test ./...
commit exit=0
--- fake go argv:
test ./...
RESULT: RED (wanted argv 'test ./a ./b', got 'test ./...')
```
사유가 맞다(REQ-001 미구현으로 게이트가 `./...`를 전달). N2의 근거: probe016.sh L9-L13에서 가짜 `go`는 `echo "$@" >> "$FAKE_GO_LOG"`, `exit 0`뿐이고, L52 `GOT="$(sed -n 1p "$W/go.log")"`로 첫 줄만 판정하며, 커밋 종료코드는 L49 `echo "commit exit=$?"`로 출력만 한다.

E4 (C4). 명령 `sh /private/tmp/t1673-audit/redcheck.sh`(probe016이 만든 스크래치 저장소에서 docs 파일만, 테스트 파일만 커밋하고, bogus 모드 `gate.yaml`로 `moai2 gate` 실행). 원문 요점:
```
--- docs-only commit ...:  go test ./...   (commit exit=0, 로그: test ./...)
--- test-only commit ...:  go test ./...   (로그: test ./...)
--- AC-011 shape:          go test: skipped — turned off by the gate.skip_tests setting
gate exit=0
```
docs-only와 test-only 커밋 모두 `./...`이고, `test_scope.mode: bogus`가 거부되지 않고 종료코드 0이다. 세 경우 모두 대응 AC의 기대와 다르므로 지금 빨강이다.

E5 (C5). 추적 스크립트 출력 `COLLECTED: 15 REQ definitions (acceptance input: read)`(UNCOVERED/ORPHAN 없음), CN-4 `COLLECTED: 6 milestones in plan order (M1 M2 M3 M4 M5 M6), 0 exit bindings, 8 ordering candidates`와 CONFLICT 없음, `moai spec lint ... --json` 출력은 info 한 건. `git rev-parse 888ce5c035a2527d81507441d71e69c929042ad0` 출력이 같은 SHA로 존재해 레저가 핀한 트리가 실재한다. `git status --short` 출력이 비어 있어 이 감사 중 추적 파일 변경이 없다.

E6 (필수 백엔드). `moai verify audit-plan`은 iter1과 같은 계획(codex required, `enforced_required: ["codex"]`)을 출력한다. `mcp__moai__audit_multi`(project_root, card_id=t1673, 같은 target 형태) 결과 요점: codex `gate: required`, `verdict: fail`(P1 1건, P2 4건), `overall_verdict: fail`, `residual_risk_note: "required-backend FAIL: codex"`, `audit_receipt: "rcpt-55e720358b7dd6d425f127c2"`, `build_commit: ee981a932`. codex 5건과 이 감사의 대응: probe009 초록 불가 = N1(직접 읽어 확인), 가짜 `go`가 그래프 질의를 무력화 = N2(probe016.sh 읽어 확인, 구현이 `go list`를 PATH의 `go`로 호출한다는 전제는 plan B.2에서 읽은 것이며 구현이 없어 실행으로 확인하지는 못함), argv 형식 충돌 = N3(plan·AC 문장 대조로 독립 확인), 커밋 거부에도 GREEN = N2에 포함(스크립트 읽어 확인), `--amend` 의미 불일치 = N4(독립 확인). 다이제스트를 `.moai/state/audit-plan-result.json`에 쓰고 `moai verify audit-plan --project-root ... --result-file ...`로 읽은 `convergence_check`는 `{"ok": true, "unmet": []}`로, 응답한 필수 백엔드의 리뷰 FAIL이다.

### Baseline-attribution

- 트리: HEAD `deb50abacc90089aef80e0f856e744e76da24d53`, 감사 시작·종료 시 `git status --short` 출력 없음. 모든 관측은 이 실행에서 이 트리에 대해 이뤄졌다.
- 도구 빌드(verification-claim-integrity §2.2): E2·E3·E4·spec lint는 이 HEAD에서 빌드한 `moai2`를 경로로 호출했다. `moai verify audit-plan`과 MCP 호출은 빌드 `ee981a932`(HEAD의 조상)이 judge했으며, 서버가 `build_lag`로 이를 스스로 보고했다.
- 레저가 핀한 트리 `888ce5c035a2527d81507441d71e69c929042ad0`와 HEAD 사이에서 Go 코드는 같다(E1의 빈 diff는 3642a3a9b 대비 HEAD 기준).

### Gaps (미검증)

- G-a. 구현이 아직 없어 "초록 경로가 도달 불가능"(N1·N2)은 프로브 스크립트와 plan 문장을 읽어 얻은 판단이다. 가짜 `go`가 `go list`를 삼키는 효과를 구현으로 재현하지는 못했다.
- G-b. LEDGER-PAS-013-A는 이번 감사에서 새 바이너리로 다시 실행하지 않았다. iter1 재현 결과와 "Go 코드 동일"(E1)에 의존한다.
- G-c. CN-4 스크립트는 계획에 `Exit:` 줄이 없어 기계적 충돌 검출이 불가능하다. 후보 8건을 읽어서 판정했다.
- G-d. `plan_artifact_hash`는 `ComputeHash` 알고리즘을 읽고 셸로 재구현한 값이다(iter1과 같은 방법). 입력 파일은 acceptance·decision-index·plan·spec이며 evidence/ 하위 파일은 해시 대상이 아니다.
- G-e. 린터가 수집한 요구 개수는 확인하지 않았으므로 린터의 침묵을 교차 근거로 쓰지 않았다.
- G-f. 이 감사가 워크트리에 `.moai/state/audit-plan-result.json`을 덮어썼다(런타임 상태 디렉터리, 커밋하지 않음).

### Residual-risk (관측했음에도 남는 위험)

- N1~N4가 닫혀도, 흔한 커밋의 affected set이 `internal/cli`·`internal/hook`을 포함해 예산을 넘는 위험은 Q1=B 판정의 결과로 그대로 남는다(SPEC이 수용, 추론으로 표기).
- 순수 단위 AC(002, 003)의 판별력은 run phase에서 RED가 실제로 기록되기 전까지 확정되지 않는다.
- R3(경로로 저장소 파일을 읽는 테스트는 보이지 않음)은 수용된 위험이다.

## Operational Notes (unverified)

- [measured] 스크래치는 모두 `/private/tmp/t1673-audit/` 아래이며 추적 파일은 수정하지 않았다. 확인 명령: `git status --short` (출력 없음).
- [inferred] N1~N3은 acceptance.md의 레저 블록과 두 프로브를 한 번에 고치는 것이 가장 싸다. 근거: 세 결함이 같은 AC-016/레저/DoD 문장에 걸려 있다.
- [assumption] 가짜 `go`의 `list` 위임 방식(실제 `go` 경로를 환경변수로 넘기는 래퍼)은 이 감사에서 시험해 보지 않았다.

AUDIT-VERDICT: FAIL spec=SPEC-PRECOMMIT-AFFECTED-SCOPE-001 receipts=rcpt-55e720358b7dd6d425f127c2
