# SPEC-TODO-CARD-ISSUANCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

_<pending: the orchestrator appends the audit-ready signal after the independent plan-audit PASS; this agent writes none>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Plan-phase Record (manager-spec, nothing here is run-phase evidence)

### 이터레이션 1 기록 (역사 — 이터레이션 2 가 아래에서 일부를 정정했다)

- 카드: t1454, Class C, Tier L. SPEC ID `SPEC-TODO-CARD-ISSUANCE-001`.
- 계획 시작 트리: `2de0a2cb6`(전체 SHA `2de0a2cb613b04765a1554f86685a3b48e0be806`, 로컬 develop 팁), 브랜치 `WT-card-issuance-overlap-graph`, 워크트리 `.claude/worktrees/t1454`. RED-now 측정 시점의 워킹 트리는 깨끗했다(`git status --short` 빈 출력).
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `spec-compact.md`, `decision-index.md`(`interview.decision_gate: on`), 이 `progress.md`.
- 개수(이터레이션 1 시점): 요구 24(모듈 5), 수용 기준 24 = 출시 차단 21 + 회귀 가드 3(AC-TCI-007·009·023), 결정 행 15(FOUNDER 13, EVIDENCE-NEEDED 1, DECIDED 1), 변이 77(기준이 잡는 73, 이유와 함께 수용 4).
- SPEC ID 사전 점검: Bash 로 `[[ "SPEC-TODO-CARD-ISSUANCE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` 를 실행해 `PASS` 를 출력했다. 유일성: `ls -d .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` 부재(`No such file or directory`), `git grep -c "SPEC-TODO-CARD-ISSUANCE-001" -- .moai/specs` 적중 0.
- 기준선 반입(이터레이션 1 계획): 카드의 git 쪽 헤드라인 수치는 이 트리에서 전부 재현됐고(`research.md` §3.1), 큐 쪽 수치는 스냅숏(2026-10-03 11:45, `last_seq` 1463) 위에서 다시 쟀다(§3.2). 이 계획의 새 측정은 §3.3 과 부록 A. 카드의 scratchpad 는 다른 세션의 `/tmp` 이므로 M0 가 스크립트를 복사한다(sha256 은 §3.4) — **복사 위치는 이터레이션 2 가 `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` 로 정정했다**(이터레이션 1 이 적은 `.moai/reports/t1454/baseline/` 은 `.gitignore:235` 에 걸려 추적될 수 없다). **오케스트레이터에게**: 그 scratchpad(`.../061c4c0e-c21e-4332-8ccc-0d1d51432c49/scratchpad/cards/`)가 정리되기 전에 보존하면 M0 의 재구성 부담이 준다. 이 에이전트는 SPEC 디렉터리 밖에 쓰지 않는다.
- 판단 도구의 출처(이터레이션 1): SPEC lint 는 이 트리의 HEAD(`2de0a2cb6`)로 빌드한 바이너리를 **경로로** 호출했다. 설치된 바이너리(`/Users/goos/go/bin/moai`, rc.26)는 커밋 `802a72235` 로 만들어졌고 그 커밋은 HEAD 의 조상이며 HEAD 와 424 커밋 차이가 난다 — 이 판정에 쓰지 않았다.
- 한계: SPEC 디렉터리 밖(큐, 규칙 파일, 코드)은 이 단계에서 건드리지 않았다. 병렬 쓰기가 아니라 한 에이전트가 SPEC 파일을 썼고, 첫 `spec.md` 는 나머지 일곱 파일보다 앞선 별도 호출로 썼다(일괄 병렬 쓰기에서 벗어난 점).

이터레이션 1 의 lint 기록(그 시점 HEAD `2de0a2cb6`, 미커밋 SPEC): `spec lint` 종료 0 과 `--strict` 종료 0, 둘 다 `✓ No findings — all SPEC documents are valid`; 양성 대조(사본의 `phase:` 를 `plan` 으로 바꿈) 종료 1 `ERROR FrontmatterPhaseInvalid`. AC 개수 가드 `TestACCounterFullCorpusMatchesBaseline` 종료 0, `absent-from-snapshot … COUNT 24`.

### 이터레이션 2 기록 (감사 FAIL 0.72 — Tier L 기준 0.85 — 대응, 이 에이전트가 한 일만)

**범위와 경계.** 대상은 감사 지적 D1~D14(이 SPEC 의 결정 D1~D15 와 별개 번호)이고, SPEC 디렉터리 밖은 건드리지 않았다. 커밋·푸시를 하지 않았고 `plan_complete_at`·`plan_status: audit-ready` 를 쓰지 않았으며(§E.1 은 비어 있다) `.moai/reports/t1454/plan-audit-iter1.md` 는 읽기만 했다. spec.md `version` 을 `"0.2.0"` 으로 올리고 HISTORY 에 0.2.0 행을 더했다(`updated` 는 이미 2026-10-03 이었다).

**측정 트리와 판정 도구.** 트리 `1894984c3254d62f2d57ced961fef8af63eb59c8`(브랜치 `WT-card-issuance-overlap-graph`; 계획 시작 develop 팁 `2de0a2cb6` 위의 SPEC 첫 커밋). 시작 시 `git rev-parse HEAD` 가 위 SHA, `git status --short` 빈 출력. 판정 도구: 행 대부분은 설치된 일반 `git`·`ls`·`wc`·`cmp`·`go test -list`(도구가 트리 빌드가 아니므로 judging-build 좌표 없음). lint 와 `spec audit` 는 **이 트리에서 빌드해 경로로 호출**했다 — `go build -o /tmp/moai-t1454 ./cmd/moai` 종료 0, 판정 빌드의 소스는 HEAD `1894984c3` 와 SPEC 디렉터리의 미커밋 편집(코드는 HEAD 와 같다). 설치된 `moai` 는 어떤 측정에도 쓰지 않았다. 이 셸의 `ls` 는 긴 목록(점 항목 포함)으로 출력하므로 `ls` 행은 종료 코드로 판정한다.

#### RED-now 원장 L1~L36 — 이 트리에서 다시 돌린 관측 (각 명령은 `acceptance.md` 의 것과 같다; 종료 코드는 `; echo "exit=$?"` 로 따로 읽었다)

| 행 | 관측한 stdout | 종료 |
|---|---|---|
| L1(새 형태) | (empty); stderr `error: pathspec '.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/baseline.md' did not match any file(s) known to git` / `Did you forget to 'git add'?` | 1 |
| L2~L20, L25~L27 | (empty) | 1 (행마다 각각 읽었다: L2 L3 L4 L5 L6 L7 L8 L9 L10 L11 L12 L13 L14 L15 L16 L17 L18 L19 L20 L25 L26 L27) |
| L21(새 형태) | `0` | 0 |
| L22(새 형태) | (empty); stderr `error: pathspec '.claude/rules/moai/workflow/card-issuance.md' did not match any file(s) known to git` / `Did you forget to 'git add'?` | 1 |
| L23(새 형태) | (empty); stderr `error: pathspec '.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/anchors.md' did not match any file(s) known to git` / `Did you forget to 'git add'?` | 1 |
| L24 | `   40037 .claude/skills/moai/workflows/gtd.md` | 0 |
| L28, L29, L30, L32, L33, L35, L36 | (empty) | 1 (각각 읽었다) |
| L31 | (empty); stderr `error: pathspec 'internal/homestate/hub_files.txt' did not match any file(s) known to git` / `Did you forget to 'git add'?` | 1 |
| L34 | (empty); stderr `error: pathspec '.claude/rules/local/card-issuance-thresholds.md' did not match any file(s) known to git` / `Did you forget to 'git add'?` | 1 |

#### 대조군 C1~C22

| 행 | 관측한 stdout | 종료 |
|---|---|---|
| C1 | `internal/cli/todo_analysis_add_test.go:2` | 0 |
| C2 | `internal/kanban/backlog_analysis_test.go:2` | 0 |
| C3 | `internal/cli/todo.go:2` | 0 |
| C4 | `internal/kanban/backlog_schema_freeze_test.go:2` | 0 |
| C5 | `internal/graph/graph_test.go:3` | 0 |
| C6 | `internal/web/todo_route_test.go:6` | 0 |
| C7 | `internal/cli/factory_m3_test.go:1` | 0 |
| C8(새 형태) | `1` | 0 |
| C9(새 형태) | `1` | 0 |
| C10 | `.moai/specs/SPEC-DRIFT-CLOSE-BODY-001/spec.md:1` | 0 |
| C11(새 형태) | `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/research.md` | 0 |
| C12 | `.claude/rules/moai/workflow/kanban-dispatch-detail.md` | 0 |
| C13 | `internal/homestate/factory.go` | 0 |
| C14 | `.claude/rules/local/gitflow-lane-protocol.md` | 0 |
| C15 | `internal/homestate/fr_schema_test.go:1` | 0 |
| C16 | `.claude/rules/moai/workflow/kanban-dispatch.md:15` | 0 |
| C17 | (empty) | 1 |
| C18 | (empty) | 0 |
| C19 | `0` | 0 |
| C20 | `1` | 0 |
| C21 | `333` | 0 |
| C22 | `214` | 0 |

**게이트 D2 의 두 부정/양성 통제.** 부정 통제(오늘 t1453 은 닫힘): L21 `0`. 눈먼 선택자 통제: 고정 두 부모 병합 `b05c3be9049822851b4cc2088cd3fc011fe39899` 에서 `46be0b8c87c76591cff7ec46b007571dafd74ca5`(`merge(t1407)`)는 첫 부모 `5e31abbcb3d09a03fc782bcfc842d8221b508dba` 의 조상이 아니고(C17 종료 1) 병합의 조상이며(C18 종료 0), `--first-parent` 선택자는 0(C19), 어느 부모든 선택자는 1(C20)이다. 병합의 부모 둘은 G11 이 `git rev-list --parents -n 1` 로 읽었다. 대조 전에 시도한 `4315f0d0e`(t1448 병합)는 첫 부모 `5e31abbcb…` 의 조상이기도 해서(`git merge-base --is-ancestor` 종료 0) 두 번째 부모로**만** 닿는 사례가 아니었다 — 그래서 `merge(t1407)` 를 골랐다.

#### 맥락·가드 행 G1~G18 (이 트리에서)

G1 (empty) 종료 0 · G2 `TestTodoListJSON_GoldenByteIdentity` + `ok …/internal/cli 1.893s` 종료 0 · G3 `TestClassifyCardText` + `ok …/internal/kanban 0.513s` 종료 0 · G4 `28301` · G5 `   28092 .claude/rules/moai/workflow/kanban-dispatch.md` · G6 `   43138 …/kanban-dispatch-detail.md`(G5·G6 은 두 파일을 한 호출로 읽었고 합계 71230) · G7 `… differ: char 25113, line 181` 종료 1 · G8 (empty) 종료 0 · G9 `2de0a2cb613b04765a1554f86685a3b48e0be806` · G10 `4315f0d0e9742b4e3741eb20efc1614f7f395238` · G11 `b05c3be9… 5e31abbcb… 46be0b8c8…` · G12 `ls .moai/reports/t1454` → 긴 목록과 `plan-audit-iter1.md` 종료 0 · G13 종료 1 · G14 `.gitignore:235:.moai/reports/*	.moai/reports/t1454/baseline/x.txt` 종료 0 · G15~G18 (empty) 종료 1.

#### 이터레이션 1 대비 출력이 달라진 행 (종료 코드 포함)

| 행 | 이터레이션 1 | 이터레이션 2 | 이유 |
|---|---|---|---|
| L1 옛 형태 `ls .moai/reports/t1454` | stderr `No such file or directory`, 종료 1 | 긴 목록(`plan-audit-iter1.md` 포함), **종료 0** | 감사 보고서가 그 디렉터리에 쓰였다(D3 의 실패 모습). 새 L1 은 `git ls-files --error-unmatch …/baseline/baseline.md` 종료 1 |
| L21, C8, C9 | `--first-parent` 형태 `0`/`1`/`1`, 종료 0 | 어느 부모든 형태 `0`/`1`/`1`, 종료 0 | 형태 교체(D2) — 값은 같고 선택자가 다르다 |
| L22, L23, C11 | `ls` 형태 | `git ls-files --error-unmatch` 형태 | `ls` 는 모양이 셸마다 다르고 추적을 말하지 않는다. L23 은 초안 경로도 옮겼다 |
| G2, G3 | 시간 `1.528s`, `0.409s` | 시간 `1.893s`, `0.513s` | 시간만 다르다(이름과 `ok` 는 같다) |
| 그 밖의 L2~L20·L24~L27, C1~C7·C10, G1·G4~G10 | — | 같은 출력, 같은 종료 | 재현됨 |

재현되지 않아 회귀 가드로 내릴 행: **없다**. 이터레이션 1 의 L1 은 명령을 바꿔 RED 로 되살렸고, AC-TCI-021 은 RED 행(L24)이 Mode B 에서 영영 붉으므로 출시 차단이 아니라 **조건부**로 내렸다.

#### lint 와 감사 도구

| 명령 | 종료 | 출력 |
|---|---|---|
| `/tmp/moai-t1454 spec lint .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` | 0 | `INFO OwnershipTransitionUnmeasured spec.md … commit 1894984c3254d62f2d57ced961fef8af63eb59c8 … has no Authored-By-Agent trailer` 다음 `0 error(s), 0 warning(s)` |
| `/tmp/moai-t1454 spec lint --strict .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` | 0 | 같은 INFO 한 줄과 `0 error(s), 0 warning(s)` |
| 중간 실행(편집 도중): 같은 두 명령 | 0 / **1** | 경고 9건(`VacuousTestAssertion` — plan.md 의 `…` 축약 run-pattern 3건, acceptance.md 의 `--- PASS: 이름` 뒤 공백 없음 6건) → 고친 뒤 위 두 줄이 종료 0/0 |
| 이중 점검: 두 번째 수정 직후 한 번 더(plan.md 의 `-run <자리표시자>` 3건이 같은 경고로 잡혀 `-run` 플래그를 문장으로 풀어 씀) | 0 / 1 → 0 / 0 | 최종 두 줄이 위와 같다 |
| 추가 하위 디렉터리 허용 점검: SPEC 디렉터리 사본(scratchpad)에 `baseline/baseline.md`·`baseline/hub-files.txt`·`m5-draft/anchors.md` 를 더해 `spec lint --strict` | 0 | `✓ No findings — all SPEC documents are valid` |
| `/tmp/moai-t1454 spec audit --filter-spec SPEC-TODO-CARD-ISSUANCE-001 --json` | 0 | `drift_findings` 에 `EraAutoDetected`(INFO, `H-5`) 하나뿐 — drift 없음 |
| `go test ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$' -count=1` | 0 | `ok  github.com/modu-ai/moai-adk/internal/spec 40.312s` |

`OwnershipTransitionUnmeasured` 는 이 SPEC 의 첫 커밋(`1894984c3`)에 `Authored-By-Agent` 트레일러가 없다는 INFO 이고 종료 코드에 영향이 없다(이터레이션 1 lint 는 첫 커밋 전이라 이 줄이 없었다). 린트가 아무것도 훑지 않은 것이 아님은 이터레이션 1 의 양성 대조(종료 1)와 중간 실행의 경고 9건(종료 1)이 보인다.

#### 변이 탐침 (바꾼 기준마다)

코드가 없으므로 대부분은 "변이가 기준을 만족하면서 요구를 위반할 수 있는가"를 읽어 판정하는 저작 시점의 탐침이다. **실행해 관측한 것은 둘**이다: (i) MU-58 의 변이(`--first-parent` 게이트) — C19·C20 이 같은 선택자가 같은 고정 커밋에서 0 과 1 로 갈리는 것을 보였다; (ii) MU-78 의 변이(무시되는 경로에만 기준선) — G14 가 옛 경로를 무시 규칙에 걸리게 하고 G13 이 디스크에 있는 같은 위치의 파일이 `git ls-files --error-unmatch` 로 종료 1 임을 보였다.

| 기준 | 바꾼 점 | 시도한 변이 | 결과 |
|---|---|---|---|
| AC-001 | 추적 경로·순서 증인 | 기준선을 `.moai/reports/` 에만(MU-78), 기준선이 코드와 한 커밋(MU-79), 카드 id 없는 변경 커밋 | 앞의 둘은 잡힘(a)·증인 2. **셋째는 잡히지 않는다** — 순서 증인 3 이 카드 id 메시지에만 기대므로 수용된 잔여로 기록했다 |
| AC-008 | finding 표 둘 확장 | `archived_findings` 만 누락(MU-80), `columnExpr` 없는 SELECT(MU-81), 큐 병합 누락(MU-82), 튜플 미고정(MU-83) | 각각 (d)(a)·(b)·(f)·(c) 가 잡는다(읽기 판정) |
| AC-013 | `relate`·`add` 로 좁힘 | 자기 간선 한 분기(MU-38), 둘째 `--parent` 덮어쓰기(MU-39), 부분 쓰기(MU-40), live 만 보는 부모 검사(MU-84), `supersedes` 순환 미검사(MU-85) | (a)(d)(바이트 동일)(e)(a) 가 잡는다. 이전 MU-38(`parent-of` 순환)은 어떤 동사로도 실행할 수 없어 폐기했고 이유를 기준 본문과 `design.md` §5.3 에 적었다 |
| AC-016 | 두 번째 부모·흡수 병합 | 첫 부모 경로만 순회(MU-86), 자체 정규식으로 흡수 방향 병합 귀속(MU-87) | (e)(f) 가 잡는다(읽기 판정) |
| AC-019 | 출하 목록 | 런타임에 SPEC·reports 경로를 읽음(MU-88), 임베드 사본 불일치(MU-89), 틀린 기록 개수(MU-90) | (d)(e)(f) 가 잡는다. **완전성**(목록에서 빠진 허브)은 못 잡아 수용했다 |
| AC-020 | 게이트 술어·D4 읽기 | `--first-parent` 게이트(MU-58 실행 관측), 배포 사본에 수치·경로·카드 id(MU-91) | 통제 행과 `TestCardIssuanceTemplateIsMechanismOnly` 가 잡는다 |
| AC-021 | 조건부·도달성 | 스텁 없음(MU-92), 경로 미고정(MU-93) | (h) 가 잡는다(읽기 판정) |
| AC-003~006, -010~012, -014, -015, -017, -018, -022, -024 | green-path 명령·출력 모양만 추가 | 기존 변이 유지 | 이름이 없는 시험을 선택자가 0개 고르고 `ok` 를 내는 변이는 `-list` 짝(이름 수)이 잡는다 — 형태는 기존 시험으로 관측했다(G3) |

변이 총수: 93개 명명 = 89개 기준이 잡음 + 4개 수용(MU-74~77).

#### 점검 목록 (이 에이전트가 실행해 읽은 것)

- 미해결 질문 표지(계획 문서 규약의 대문자 표지): SPEC 파일 여덟 개에 대한 `grep` 이 종료 1(적중 0)이었다. 새로 들어오지 않았다.
- `status:` 필드: `grep -n "^status:"` 를 `spec.md` 를 뺀 일곱 형제 파일에 걸어 종료 1(적중 0). `decision-index.md` 에 `status:` 가 없다.
- 요구·기준 추적성은 `git grep` 이 아니라 `grep` 으로 읽었다(작업트리 가드가 awk 프로그램을 거부한다고 지시받았고 awk 를 쓰지 않았다): 요구 24개 모두 `**Covers**` 줄에서 하나 이상 인용된다(AC-001→001, -002→002, -003→002·006, -004→003, -005→004, -006→005, -007→006, -008→007·008·009, -009→007, -010→010, -011→011, -012→012·017, -013→013, -014→014, -015→015, -016→016, -017→018, -018→019, -019→020, -020→021·022, -021→021, -022→023, -023→024, -024→참조). 새 REQ 는 더하지 않았다.
- 개수: `grep -c -E "^- \*\*REQ-TCI-[0-9]+\*\*" spec.md` 가 24, `grep -c -E "^## AC-TCI-[0-9]+" acceptance.md` 가 24. 요구 24 ≤ 25, 기준 24 ≤ 25, 모듈 5. 기준 분류 = 출시 차단 20 + 조건부 1(AC-TCI-021) + 회귀 가드 3(AC-TCI-007·009·023).
- SPEC ID 점검(이터레이션 1 의 `PASS` 출력)은 ID 가 바뀌지 않아 다시 쓰지 않았다.

#### 이 반복이 모든 명령 밖에서 한 판단 (공개)

- **D4 읽기**: 수치는 로컬 전용 규칙(`.claude/rules/local/card-issuance-thresholds.md`, 이 저장소에서 추적됨)에만, 배포 사본은 메커니즘만. 대안(사본에 카드 경로 없이 수치 / 기준선을 가리킴)은 `design.md` §8 에 반려 이유와 함께 있다. 이것은 설계 판단이고 `decision-index.md` Q7 에 선호 없이 열려 있다.
- **D1 출하 입력**: 임베드 데이터 파일 `internal/homestate/hub_files.txt` 와 시험 셋(`design.md` §12). 설정 키와 Go 상수는 반려 이유와 함께 기록했다.
- **새로 찾은 것**: 카드→파일 간선 층의 first-parent 맹점(REQ-TCI-016, AC-TCI-016, `design.md` §6). 단일 귀속 지점 `subjectAttribution` 이 흡수 방향 병합 제목을 어떻게 다루는지는 **읽지 못했다**(M3 첫 RED 시험이 관측한다).
- **AC-TCI-013 의 green path 는 M2(`add` 플래그)와 M3(`relate`)에 걸친다.** 기준은 M3 가 끝나야 모두 초록이다.

#### 공백 (관측하지 못한 것)

1. green-path 셀의 통과 출력은 **모양**일 뿐이다 — 새 시험 이름이 아직 없어 어느 것도 실행하지 않았다.
2. 카드→파일 간선의 로컬/CI 일치(합성 병합 참조)는 측정하지 못했다.
3. 허브 목록의 완전성은 시험이 다시 재지 못한다.
4. 게이트의 고정 SHA 대조 네 행은 미푸시 브랜치 `WT-github-flow-default` 의 객체가 있는 클론에서만 재현된다.
5. 변이 탐침 대부분은 읽기 판정이다.
6. `moai gtd pr t1453` 의 착지 SHA 형태, t1453 의 병합 제목 형태는 확인하지 못했다(게이트는 닫힌 쪽으로 안전).
7. 이 반복은 `go test` 를 `-list` 와 `TestACCounterFullCorpusMatchesBaseline` 에만 썼다. `go test ./...` 는 돌리지 않았다.

#### 트리 상태

마지막 확인은 아래 `git status --short` 한 번이다(오케스트레이터가 커밋한다; 커밋·푸시·SPEC 디렉터리 밖 파일 변경 없음).

### 이터레이션 3 기록 (감사 FAIL 0.82 — Tier L 기준 0.85, 마지막 허용 반복 — 대응, 이 에이전트가 한 일만)

**범위와 경계.** 대상은 이터레이션 2 감사의 지적 D15~D25(이 SPEC 의 결정 D1~D15 와 별개 번호)이고 SPEC 디렉터리 밖은 건드리지 않았다. 커밋·푸시를 하지 않았고 `plan_complete_at`·`plan_status: audit-ready` 를 쓰지 않았으며(§E.1 은 비어 있다, §E.2~§E.4 는 자리 표시만 남아 있다) `.moai/reports/t1454/` 는 읽기만 했다. spec.md `version` 을 `"0.3.0"` 으로 올리고 HISTORY 에 0.3.0 행을 더했다. 감사 보고서의 모든 근거를 편집 전에 이 트리에서 다시 읽었다.

**측정 트리와 판정 도구.** 트리 `ad02a56779473afc2ef72e9ecfe16d517a297a7b`(브랜치 `WT-card-issuance-overlap-graph`). 시작 시 `git rev-parse HEAD` 가 이 SHA, `git status --short` 는 빈 출력. 증거 행은 설치된 일반 `git`·`ls`·`wc`·`cmp`·`go test -list`(도구가 트리 빌드가 아니므로 judging-build 좌표 없음)로 읽었고, 전 행을 `acceptance.md` 의 원문 명령 그대로 한 스크립트가 셸 없이(`shlex`) 실행했다. lint 와 `spec audit` 는 이 트리 `ad02a5677` 에서 `go build -o /tmp/moai-t1454 ./cmd/moai`(종료 0)로 만든 빌드를 **경로로** 호출했다. 설치된 `moai` 는 쓰지 않았다. 이 세션의 셸은 레인 표지가 있고 worktree 가드가 `awk`·heredoc 복합 호출을 거절하므로 `grep`·`python3 <파일>` 로 대체했다(대체한 것: 순서·추적성 점검의 awk 동사 — 점검 목록 참조).

#### 지적별 처분과 재검증

| 지적 | 처분 | 파일 | 이 반복이 다시 읽은 근거 |
|---|---|---|---|
| D15 순서 증인이 B 앞의 제품 커밋을 못 봄 | **수정** | acceptance.md(AC-TCI-001 증인 3 문장 정정·증인 4 `card-head:` 정의·증인 5 추가, C23~C27, MU-94·MU-110, 완료 정의 #2), plan.md(§D, F.2, R14, F.12), design.md §12.3·12.4, spec-compact.md | C23 `0`, C24 `3`, C25 `0`, C26 `0`, C27 `2` (전부 종료 0, 고정 SHA 위). 감사의 `0`·`0`·`3` 이 재현됐다 |
| D16 REQ-TCI-002(c) 양성 겹침 기준 없음 | **수정** | spec.md(Module A "예상 파일" 용어), acceptance.md(AC-TCI-002 (e), AC-TCI-003 (f)(g)·시험 8개, L37~L39, MU-95~98), plan.md F.3, design.md §3.1, spec-compact.md | AC-TCI-003 의 필수 시험에 `unmeasured` 시험만 있음을 문면에서 읽었다. L37~L39 는 새 시험 이름의 부재(종료 1) |
| D17 REQ-TCI-013 마지막 절 기준 없음 | **수정** | acceptance.md(AC-TCI-013 (f)·시험 5개, L40, MU-99~102), plan.md F.5, design.md §5.2, spec-compact.md | `BacklogSemanticRelations` 여섯 이름과 `internal/cli/todo_relate.go:63-65` 의 거절 문장을 읽었다. 기존 시험 `TestRelateAndUnrelateRefusals` 를 레인 변수를 지우고 돌려 PASS(G25) |
| D18 속성 없는 카드의 저장 형태 | **수정** | acceptance.md(AC-TCI-008 (g)·시험 13개, L41·L42, MU-103·104), plan.md F.4, spec-compact.md | 독립 Go 프로그램이 값 형 `omitempty` 구조체 필드는 `{"id":"t1","issuance":{}}`, 포인터 형은 `{"id":"t1"}` 를 냄을 관측 |
| D19 개수 귀속 | **수정** | spec.md §A.1·§G, plan.md F.2, acceptance.md AC-001 (d), research.md §10 | 이 반복이 14:41 에 읽은 값: `WT-` 브랜치 367, 워크트리 89, `.moai/specs` 디렉터리 1,031, `card:` 를 가진 SPEC 33 |
| D20 게이트 선택자 | **수정** | acceptance.md(L21·C8·C9·C19~C22 앵커·SHA 핀, C28~C32, G23, AC-TCI-020 판독 1~5, MU-107·112), plan.md F.8·§G, design.md §13, spec.md REQ-TCI-021·§G, spec-compact.md, decision-index.md Q1 | C28 `5`, C29 `5`(착지 후보 0), C30 `0`(t1448 후보 1), C31 `1`, C32 `0`. 감사가 본 흡수 병합 다섯 개 재현(G23) |
| D21 AC-022(d) | **수정(한계 명시)** | acceptance.md(AC-TCI-022 (d)·MU-109·MU-113), plan.md F.7, spec-compact.md | 문면 판정(변이가 `//`·`url(` 을 통과) — 코드 없음 |
| D22 허브 목록 전제 | **좁힘 — 검증하지 못한 수용 잔여로 격하** | spec.md(§F Out of Scope·§G), design.md §7.2·7.4, plan.md R15, acceptance.md AC-TCI-019 | `git ls-files -- internal/config/defaults.go` 가 이 저장소의 경로를 출력하고 `git ls-files | grep -c 'defaults.go$'` 는 4. 사용자 프로젝트는 볼 수 없어 명령으로 확인하지 못했다. 입력을 명시 `files` 속성으로 좁혀 노출을 줄였다 |
| D23 정적 스캔·건너뜀 | **수정(한계 명시)** | acceptance.md(AC-TCI-019 (d)(f)·시험 7개, L43·L44, MU-105·106·111), design.md §12.3, plan.md F.6 | `.github/workflows/ci.yml:131` 의 test 작업이 `fetch-depth: 0` 임을 `grep -n fetch-depth` 로 읽었다 |
| D24 `tree:` 출처·`ls` 문장 | **수정** | acceptance.md(L11, AC-001 `card-head:`, G12), research.md §10 항목 17 | 이 셸의 `ls .moai/reports/t1454` 는 파일 이름 둘만 출력(종료 0) |
| D25 시험 패키지 귀속 | **수정 + 새 오류 하나 발견** | plan.md §D·F.8 단계 6, acceptance.md AC-TCI-021 (f)(g) | `todo_skill_doc_parity_test.go`·`todo_classify_doc_parity_test.go` 는 `internal/cli`, `TestGoldenCommittedArtifactsMatchEmission` 은 `internal/template/agentemit`(G21 은 `internal/template` 에서 이름 3개, G22 는 `agentemit` 에서 1개) — AC-TCI-021 (f) 의 이터레이션 2 문면은 이 시험을 `internal/template` 에서 찾았다 |

이 밖에 이 반복이 새로 찾은 것: **plan §D 의 SCRUB 접두사가 팩토리 레인에서 부족하다.** 레인 가드(`factoryLaneRefusal`)는 `MOAI_FACTORY_ROLE`·`MOAI_FACTORY_WORKER`·`MOAI_KANBAN_BACKEND` 를 읽는데 이터레이션 2 의 SCRUB 는 kanban 변수 다섯만 지웠다. 그 접두사로 돌린 기존 시험 `TestRelateAndUnrelateRefusals` 는 `lane boundary` 거절로 붉었고(G24, 종료 1) 팩토리 변수까지 지우자 통과했다(G25, 종료 0). plan §D 를 아홉 변수로 고쳤다.

#### 원장 재실행 (acceptance.md 의 L·C·G 전 행, 이 트리 `ad02a5677`)

실행: `python3 <scratchpad>/ledger.py` — 문서 표의 `| L`·`| C`·`| G` 행에서 명령을 그대로 뽑아 `shlex` 로 나누어 실행하고 문서의 stdout·종료 코드와 대조한다. 결과: **99행(L1~L44, C1~C32, G1~G20·G23, 펜스 G21·G22) 모두 문서와 일치**, 불일치 0. 출력(요약): L1~L20·L25~L27·L28~L30·L32·L33·L35~L44 는 stdout 비어 있음·종료 1(L1·L22·L23·L31·L34 는 stderr `did not match any file(s) known to git`); L21 `0`(종료 0); L24 `   40037 .claude/skills/moai/workflows/gtd.md`; C1~C16 은 문서의 값; C17 종료 1, C18 종료 0, C19 `0`, C20 `1`, C21 `333`, C22 `214`, C23 `0`, C24 `3`, C25 `0`, C26 `0`, C27 `2`, C28 `5`, C29 `5`, C30 `0`, C31 `1`, C32 `0`; G1 비어 있음; G2·G3 이름 + `ok`(시간 `1.468s`·`0.502s`); G4 `28301`; G5 `28092`; G6 `43138`; G7 `differ: char 25113, line 181` 종료 1; G8 비어 있음; G9 `2de0a2cb613b…`; G10 `4315f0d0e974…`; G11 두 부모; G12 파일 이름 둘; G13 종료 1; G14 `.gitignore:235:…` 종료 0; G15~G18 비어 있음·종료 1; G19 `2de0a2cb613b…`; G20 비어 있음; G21 이름 3개 + `ok`(`0.448s`); G22 이름 1개 + `ok`(`0.450s`); G23 다섯 줄. 재현되지 않아 회귀 가드로 내릴 행: **없다**.

**출력이나 명령이 이터레이션 2 와 달라진 행.** 값이 바뀐 행은 없다(G2·G3·G21·G22 의 `ok` 줄 시간만 다르다). 명령·귀속이 바뀐 행: L21, C8, C9(핀한 SHA·id 뒤 `[^0-9]`), C19~C22(앵커·핀), G9(핀한 SHA 쌍; 움직이는 형태는 G19 로 이유와 함께 보존), G10(핀), G12(셸 의존 문장). 새 행: L37~L44, C23~C32, G19~G25. 움직이는 ref 를 그대로 둔 곳: G19, AC-TCI-020 의 게이트 판독, AC-TCI-020 Mode B·AC-TCI-021 (b) 의 `git merge-base develop HEAD` — 질문이 "지금 작업 트리에서 닿는가"·"지금 흡수한 기준이 무엇인가"라 뒤집힘이 신호이므로(`verification-completeness.md` §4 의 판별 질문) 핀하지 않았고 이유를 각 자리에 적었다.

**D15 증인 5 의 실제 출력(실제 이력, 고정 SHA).** 질의: `git rev-list --count --grep=t1454 <B>^ -- internal cmd .claude` 의 형태를 이 카드의 첫 plan 커밋에 건 C23 이 `0`. "기준선이 제품 커밋 뒤에 착지" 변이의 형태는 카드 t1460 을 대역으로 같은 형태를 `1894984c3` 에 건 C24 가 `3`. 같은 이력을 증인 3 의 범위로 보면 C25·C26 이 `0`·`0`. 아직 `B` 가 없으므로 이 카드 자신의 B 에 대한 값은 M0 가 낸다.

#### lint 와 감사 도구

| 명령 | 종료 | 출력 |
|---|---|---|
| `/tmp/moai-t1454 spec lint .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` | 0 | `INFO OwnershipTransitionUnmeasured spec.md … commit 1894984c3254d62f2d57ced961fef8af63eb59c8 … has no Authored-By-Agent trailer` 다음 `0 error(s), 0 warning(s)` |
| `/tmp/moai-t1454 spec lint --strict .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` | 0 | 같은 INFO 한 줄과 `0 error(s), 0 warning(s)` |
| 양성 대조: 사본(scratchpad)의 `phase:` 를 `plan` 으로 바꿔 같은 lint | 1 | `ERROR FrontmatterPhaseInvalid … phase "plan"` 과 `1 error(s), 0 warning(s)` |
| `mcp__moai__spec_audit`(이 SPEC, `project_root` = 이 워크트리) | — | `drift_findings` 에 `EraAutoDetected`(INFO, `H-5`) 하나, `modern_era_clean: 1` (서빙 빌드는 `v3.2.0-rc.25` 로 트리보다 오래돼 보조 증거일 뿐이다) |
| `go test ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$' -count=1` | 0 | `ok  github.com/modu-ai/moai-adk/internal/spec 9.084s` |

lint 는 이 문서 편집을 모두 마친 뒤 다시 돌렸다(아래 최종 점검 줄).

#### 변이 탐침 (추가·변경한 기준마다)

**명령으로 실행해 관측한 것**: MU-94 의 형태(C23~C27, 대역 카드), MU-107 의 형태(C28~C32, 고정 SHA), MU-108 의 형태(G21·G22), MU-103 의 메커니즘(독립 Go 프로그램), MU-99~102 의 전제(G25 — 오늘의 `relate` 가 범위 밖 값을 거절). **읽기로 판정한 것**(코드가 없어 변이를 만들어 돌리지 못했다): MU-95~102 의 기준 단언, MU-104~106, MU-109. **돌리지 못한 것과 이유**: 가짜 저장소에서 MU-94 의 "제품 커밋 → 기준선 → 이후 커밋" 전체 순서를 만드는 일은 worktree 가드가 다른 저장소의 `git` 을 거절할 수 있어 하지 않았다 — 대역 카드의 실제 이력으로 질의 형태를 관측하는 것으로 대신했다.

변이 총수: 113개 명명 = 105개 기준이 잡음 + 8개 수용(MU-74~77, MU-110~113). 이터레이션 3 이 새로 이름 붙인 것은 20개(MU-94~113)이고 16개는 기준이 잡는다.

#### 점검 목록 (이 에이전트가 실행해 읽은 것)

- 개수: `python3 <scratchpad>/xcheck.py` — 요구 `spec.md` 24·`spec-compact.md` 24, 수용 기준 `acceptance.md` 24·`spec-compact.md` 24; 기준별 "필수 시험 N개" 의 이름 수가 N 과 같고 green-path 의 `-run` 가지 이름 집합과 일치(AC-002 2, -003 8, -004 3, -005 2, -006 3, -008 13, -010 3, -011 3, -012 2, -013 5, -014 3, -015 4, -016 5, -017 3, -018 4, -019 7, -022 4 — 불일치 0); 필수 시험 이름이 plan.md 에 있는지(없는 것은 기존 시험 `TestSD_AC014_MCPMatchesCLIWithProjectRoot` 하나이고 plan 은 `^…$` 형태로 인용한다). 분류: 출시 차단 20(-001~006, -008, -010~020, -022, -024) + 조건부 1(-021) + 회귀 가드 3(-007, -009, -023).
- 미해결 질문 표지: `grep -n -i -E 'NEEDS[ -]CLARIFICATION'` 을 SPEC 디렉터리의 `*.md` 에 걸어 편집 전 적중이 이 파일의 서술 한 줄뿐이었고 그 줄을 표지 없는 문장으로 고친 뒤 0 이다(최종 점검 줄이 다시 읽는다).
- `status:` 필드: `grep -n -E '^status:'` 를 `spec.md` 를 뺀 일곱 형제 파일에 걸어 적중 0(종료 1). `decision-index.md` 에 `status:` 가 없다.
- 템플릿 경로의 카드 id: 이 반복이 `internal/template/templates/` 아래 어떤 파일도 편집하지 않았다(SPEC 디렉터리만 변경 — 위 `git status --short`). 계획의 파일 목록에 새 `t####` 를 템플릿 사본에 더하는 항목이 없음을 읽었다(plan §D 템플릿 가드, REQ-TCI-021 의 "no card id … in any template copy").
- 추적성: 요구 24개 모두 `**Covers**` 줄에서 하나 이상 인용된다(awk 동사 대신 `grep`·`python3` 로 읽었다; AC-TCI-002 의 REQ-TCI-002 인용은 그대로, 새 기준은 새 REQ 를 더하지 않았다).
- 순서 점검(CN-4 의 awk 동사 대체): `grep -n -i -E 'before|after|first|prior to|앞서|먼저'` 대신 AC-TCI-001 증인 1~5, plan §D·F.2·F.12, 완료 정의 #2 를 직접 읽어 M0 → 제품 커밋 순서가 plan 과 acceptance 에서 같음을 확인했다(증인 5 가 앞쪽을 본다).
- 교차 모순 점검: 증인 개수(1~5)·`card-head:`·게이트 판독(1(a)(b), 2, 3, 4, 5)·시험 개수·변이 개수·핀(`ad02a5677`)을 spec·plan·design·acceptance·decision-index·spec-compact 에서 대조했다. 남은 "증인 1~4" 는 MU-94 문장 안의 의도된 서술 둘뿐이다.

#### 공백 (관측하지 못한 것)

1. green-path 셀의 통과 출력은 **모양**일 뿐이다 — 새 시험 이름이 아직 없어 어느 것도 실행하지 않았다.
2. D16~D18·D21·D23 의 변이는 기준 문면에 대한 읽기 판정이다. 이 카드 자신의 B 에 대한 증인 1~5 는 M0 이전이라 실행하지 못했다(대역 이력으로 질의 형태만 관측).
3. 출하 허브 목록과 사용자 프로젝트의 경로 겹침, 카드→파일 간선의 로컬/CI 일치, 허브 목록의 완전성, 어휘 이웃의 정밀도는 측정하지 못했다(`spec.md` §G).
4. 게이트의 고정 SHA 대조 행(C17~C20, C28·C29·C31)은 미푸시 브랜치의 객체가 있는 클론에서만 재현된다. 착지 되돌림(revert)·t1453 착지 제목 형태는 확인하지 못했다.
5. 큐 쪽 baseline 수치와 스냅숏 위 figure 35개는 이 반복에서 다시 재지 않았다(M0 의 몫). `go test ./...` 는 돌리지 않았고 `go test` 는 `-list`, `TestACCounterFullCorpusMatchesBaseline`, `TestRelateAndUnrelateRefusals`(G24·G25)에만 썼다.
6. 설치된 `moai` 와 서빙 중인 MCP 빌드는 트리 HEAD 보다 오래돼 판정에 쓰지 않았다(MCP `spec_audit` 는 보조 증거).

#### 이터레이션 3 트리 상태

**최종 점검(모든 편집 뒤).** 원장 재실행 99행 일치·불일치 0, `spec lint` 종료 0(`0 error(s), 0 warning(s)`), `--strict` 종료 0(같은 출력), 미해결 질문 표지 grep 종료 1(적중 0), `spec.md` 를 뺀 일곱 파일의 `^status:` grep 종료 1(적중 0), `spec.md` 는 `version: "0.3.0"`·`status: draft`, 요구 24·수용 기준 24.

작업 시작 시 `git status --short` 는 빈 출력이었고, 모든 편집을 마친 뒤 마지막으로 읽은 값은 SPEC 디렉터리의 수정 파일 여덟(`acceptance.md`, `decision-index.md`, `design.md`, `plan.md`, `progress.md`, `research.md`, `spec-compact.md`, `spec.md`)뿐이다. HEAD 는 여전히 `ad02a5677`·브랜치 `WT-card-issuance-overlap-graph` 이고 커밋·푸시·SPEC 디렉터리 밖 파일 변경은 없다.

### 이터레이션 4 기록 (감사 FAIL 0.83 — Tier L 기준 0.85, 리더가 허용한 마지막 델타 감사 — 대응, 이 에이전트가 한 일만)

**범위와 경계.** 대상은 이터레이션 3 감사의 지적 D26~D33(이 SPEC 의 결정 D1~D15 와 별개 번호)이고 SPEC 디렉터리 밖은 건드리지 않았다. 커밋·푸시를 하지 않았고 `plan_complete_at`·`plan_status: audit-ready` 를 쓰지 않았으며(§E.1 은 비어 있다, §E.2~§E.4 는 자리 표시만 남아 있다) `.moai/reports/t1454/` 는 읽기만 했다. 리더가 감사 상한에서 델타 감사 한 번을 더 허용한 근거는 범위가 바뀌지 않는다는 것이라 요구·기준·마일스톤을 더하지 않고 지적이 닿는 문면만 고쳤다(REQ 24, AC 24 그대로, 새 마일스톤 없음; 새 시험 이름은 `TestTodoGraphViewDoesNotWaitOnQueueLock` 하나). spec.md `version` 을 `"0.4.0"` 으로 올리고 HISTORY 에 0.4.0 행을 더했다. 감사 보고서의 근거를 편집 전에 이 트리에서 다시 읽었고(아래 "재검증") 틀린 지적은 없었다.

**측정 트리와 판정 도구.** 트리 `8fff427eb98a24e864cf41f1aeb2835aa0eb94ca`(브랜치 `WT-card-issuance-overlap-graph`; 시작 시 `git rev-parse --short HEAD` 가 `8fff427eb`, 브랜치 이름 일치). 문서 수준 핀 `ad02a5677` 과는 제품 경로가 같다(`acceptance.md` G26 이 빈 출력·종료 0). 증거 행은 설치된 일반 `git`·`ls`·`wc`·`cmp`·`go test -list` 로 읽었고 전 행을 한 스크립트(`iter4/ledger4.py`, scratchpad)가 셸 없이(`shlex`) 문서의 원문 명령 그대로 실행했다. lint·`spec audit` 는 이 트리에서 `go build -o /tmp/moai-t1454 ./cmd/moai`(종료 0)로 만든 빌드를 **경로로** 호출했다. 설치된 `moai` 는 쓰지 않았다. worktree 가드가 복합 호출·awk 를 거절하므로 `python3 <파일>`·`grep` 으로 대체했다(대체한 것: 이터레이션 3 점검의 awk 동사). D26 반례는 워크트리 밖 scratchpad 의 스크래치 저장소(`git` 2.54.0, Apple Git-157)에서 만들었다.

#### 지적별 처분과 재검증

| 지적 | 처분 | 파일 | 이 반복이 다시 읽은 근거 |
|---|---|---|---|
| D26 증인 3 의 개수 비교 | **수정** | acceptance.md(AC-TCI-001 증인 3·완료 정의 #2·스크래치 표·C25·C26·C37~C40·MU-114·115), plan.md(§D·F.2·F.12·§G), design.md §12.4, spec-compact.md | 스크래치 S2b: 옛 증인 3 `2`·`2` PASS(변이 생존), 새 형태(`--full-history` SHA 목록) `3` 대 `2` FAIL. S3b 도 같다. S2d(곁가지에서 넣었다 되돌림): 옛 형태 `1/1` 과 `--full-history` 없는 목록 `1/1` 이 모두 통과, `--full-history` 목록만 `3/1` 로 실패. 실제 이력(t1460 대역): 옛 `3` 대 `4`(정당한 이력에 거짓 실패), 새 형태 같은 네 SHA |
| D27 게이트 `A` 가 메시지 전체 | **수정** | acceptance.md(AC-TCI-020 판독 1~3·닫는 문단·MU-116, C29~C32 명령 교체, C33~C36), plan.md(F.8·§G), design.md §13.3·13.4, decision-index.md Q1, spec.md(REQ-TCI-021·§G), spec-compact.md | t1439 `d53e6ca59…`: `T`=1, 옛 `A`=1(후보 0), 새 `A`=0(후보 1). 핀에서 `T`=333, 제목 줄 `absorb` 76, 본문에만 28, 옛 `A`=104 가 감사와 일치(`python3`). 흡수 병합 `b05c3be90…` 에서 `T`=5·새 `A`=5, `HEAD` 에서 `T`=0·`A`=0 |
| D28 AC-TCI-022 (b) | **수정** | acceptance.md(AC-TCI-022 (b1)~(b3)·시험 5개·L45·G27~G29·MU-66 문구·MU-117·118·122), plan.md(F.7·§G), design.md §9, spec-compact.md | 스크래치(`python3 fcntl`): 잡았다 놓아도 락 파일의 바이트·mtime·크기가 그대로, 락을 쥔 동안 두 번째 설명자의 블로킹 flock 은 1.0초 안에 돌아오지 못함, `LOCK_NB` 는 즉시 실패, 락을 쥔 채 일반 읽기는 즉시 돌아옴. `board_lock_unix.go:65-93` 을 읽어 락 도우미 경유는 소유자 기록을 파일에 쓴다는 점을 확인했다(감사의 `backlog_store.go` grep 은 도우미 파일을 보지 못했다) |
| D29 여섯 마일스톤 | **수정** | spec.md §A.2 | `grep -n -E '여섯 마일스톤'` 적중이 그 한 줄뿐이었다 |
| D30 겹침 음성·섞임 | **수정** | acceptance.md(AC-TCI-003 Given L3·N4·(g)(h)·MU-119, 경계 사례), spec.md Module A 용어, design.md §3.1, plan.md F.3, spec-compact.md | 문면 판정 — 새 시험 이름 없음(`TestIssuanceInFlightOverlapNoneIsMeasured` 의 입력) |
| D31 템플릿 사본 수치 검사 범위 | **좁힘(수용 잔여 MU-121)** | acceptance.md(AC-TCI-020 Mode A 문단), spec.md §G, plan.md §D, design.md §8 | `card_id_leak_test.go` 를 읽었다. 수치 검사를 다섯 사본으로 넓히면 3·5 같은 흔한 수에서 거짓 적중이 나므로 설계 변경이 필요해 넓히지 않았다 |
| D32 건너뜀 정책 시험 | **수정** | acceptance.md(AC-TCI-019 (f)·MU-106), design.md §12.3, plan.md F.6 | 문면 판정 |
| D33 제품 경로 필터 | **수정(범위 명시)** | acceptance.md(AC-TCI-001 정의·증인 3~5·DoD #2, C23~C26·G20·G26), plan.md, design.md, spec-compact.md | `ls -d pkg scripts .codex .agents mods` 가 모두 존재. 필터를 `internal cmd pkg scripts .claude .codex` 로 넓혔고 C23 `0`·C24 `3`·G20 빈 출력은 그대로다. `.agents`·`mods`·`Makefile`·`go.mod` 는 이 SPEC 이 고치지 않는 경로라 수용된 범위 밖으로 명시했다 |

**M5 압축과 카드 id 기준선(이터레이션 3 보고서의 추정 위험).** `internal/template/card_id_leak_test.go` 를 읽었다: `TestCardIDBaselineHasNoStaleEntries`(`:69`)는 기준선 (파일, id) 짝의 id 가 그 파일에서 사라지면 `stale baseline entry` 로 붉다. M5 가 만지는 파일에 걸린 짝은 넷이다(`gtd.md t696` — 템플릿 사본 `:244`, `kanban-dispatch.md t1330` — `:94`, `kanban-dispatch-detail.md t133`·`t224`). 그러므로 **M5 의 압축은 기준선 짝을 지울 수 있다** — 특히 증거 기록 문장인 t1330·t696 이 압축 후보다. 지우면 같은 커밋에서 `cardIDBaseline` 항목을 지워야 하고, `t696` 짝이면 `TestCardIDBaselineIsPerFileAndPerLiteral`(`:98`)이 그 짝을 양성 대조 상수로 쓰므로 상수도 다른 짝으로 고쳐야 한다. 이를 plan.md §D·F.8 단계 6·F.12 와 spec.md §E(조건부 파일)·acceptance.md AC-TCI-021 (g)·MU-120 에 적었다. 현재 트리에서 `go test ./internal/template -run '^TestCardIDBaselineHasNoStaleEntries$' -count=1 -v` 는 `--- PASS`·`ok`(0.431s)다. 변이(문장을 지우되 항목은 남김)를 만들어 돌리지는 않았다 — 읽기 판정이다. 계획의 파일 목록이 템플릿 경로에 새 `t####` 를 더하는 항목은 없다(조건부 파일 `card_id_leak_test.go` 는 템플릿 경로가 아니고 기존 id 를 지우는 방향뿐이다).

#### 원장 재실행 (acceptance.md 의 L·C·G 전 행, 이 트리 `8fff427eb`)

실행: `python3 <scratchpad>/iter4/ledger4.py` — 표의 `| L`·`| C`·`| G` 행과 펜스 행(G21·G22·C37~C40)에서 명령을 그대로 뽑아 `shlex` 로 나누어 실행하고 문서의 stdout·종료 코드와 대조한다(go test 줄은 소요 시간만 떼고 비교). 결과: **`rows=112 ok=112 mismatch=0`**(L1~L45, C1~C40, G1~G23·G26~G29). `unset … && go test` 복합 행 G24·G25 는 따로 돌렸다 — G24(kanban 변수 다섯만 지움): `--- FAIL: TestRelateAndUnrelateRefusals` 와 `lane boundary` 거절 문장, 종료 1; G25(팩토리 변수까지 지움): `--- PASS: TestRelateAndUnrelateRefusals (7.13s)` 와 하위 시험 여덟 개 PASS, `ok … 8.203s`, 종료 0. 새·바뀐 행의 관측: L45 (empty)·종료 1; C25·C26 (empty)·종료 0; C23 `0`, C24 `3`; C29 `5`, C30 `0`, C31 `1`, C32 `0`; C33 `1`, C34 `0`, C35 `1`, C36 `1`; C37 `3`, C38 `4`, C39·C40 같은 네 SHA(`2de0a2cb6…`, `a165d164d…`, `6ac7b0aae…`, `b2c6af74c…`)·종료 0; G20·G26 (empty)·종료 0; G27 (empty)·종료 1; G28 `internal/kanban/backlog_store.go:15`; G29 `internal/web/app.go`. 재현되지 않아 회귀 가드로 내릴 행: **없다**.

**출력이나 명령이 이터레이션 3 과 달라진 행.** 값이 바뀐 행은 G12 하나다(파일 이름 둘 → 넷: 감사 보고서 iter3 과 `verdict.md` 가 디렉터리에 더 쓰였다 — 셸 의존 행이라 종료 코드 0 으로 판정한다). 시간만 다른 행: G2·G3·G21·G22. 명령이 바뀐 행(값은 같다): C23·C24·G20(필터), C25·C26(개수 → SHA 목록이고 값은 `0`·`0` → 빈 목록), C29~C32(둘째 질의 `A` 를 같은 줄 형태로). 새 행: L45, C33~C40, G26~G29.

#### lint 와 감사 도구

| 명령 | 종료 | 출력 |
|---|---|---|
| `/tmp/moai-t1454 spec lint .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` | 0 | `INFO OwnershipTransitionUnmeasured spec.md … commit 1894984c3254d62f2d57ced961fef8af63eb59c8 … has no Authored-By-Agent trailer` 다음 `0 error(s), 0 warning(s)` |
| `/tmp/moai-t1454 spec lint --strict .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` | 0 | 같은 INFO 한 줄과 `0 error(s), 0 warning(s)` |
| 양성 대조: 사본(scratchpad)의 `phase:` 를 `plan` 으로 바꿔 같은 lint | 1 | `1 error(s), 0 warning(s)` (`FrontmatterPhaseInvalid`) |
| `/tmp/moai-t1454 spec audit --filter-spec SPEC-TODO-CARD-ISSUANCE-001 --json` | 0 | `modern_era_clean: 1`, `drift_findings` 에 `EraAutoDetected`(INFO, H-5) 하나 |
| `go test ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$' -count=1` | 0 | `ok  github.com/modu-ai/moai-adk/internal/spec 16.452s` |

lint 는 모든 편집을 마친 뒤 돌렸다. 종료 코드는 `0 error(s)` 줄과 호출의 성공 여부로 읽었다(이 도구는 오류가 있을 때만 비제로이고 양성 대조가 1 을 낸다).

#### 변이 탐침 (추가·변경한 기준마다)

**명령으로 실행해 관측한 것**: MU-114·115(스크래치 저장소 열넷, 옛 형태 대 새 형태 — 표는 `acceptance.md`), MU-116 의 형태(t1439 에서 옛 `A` `1`, 새 `A` `0` — C34·C35·C36), MU-117·118 의 전제(`python3 fcntl` 로 flock 의 동작). **읽기로 판정한 것**: MU-119(경로 비교 방식)·MU-66 의 새 문구는 기준 문면에서, MU-120 은 `card_id_leak_test.go` 를 읽고 현재 통과를 실행해 확인했을 뿐 변이를 만들지 않았다. **돌리지 못한 것과 이유**: `TestTodoGraphViewDoesNotWaitOnQueueLock` 은 코드가 없고 이 반복이 `internal/web` 에 시험 파일을 둘 수 없어(SPEC 디렉터리 밖 편집 금지) 락을 쥔 채 기본 보기가 제때 200 을 낸다는 통제를 실행하지 못했다 — 큐 읽기가 락 없음이라는 `backlog_store.go` 머리 주석을 읽은 것이 근거다. 2초가 `-race` 아래 건강한 보기에 충분한지도 관측하지 못했다.

변이 총수: 122개 명명 = 112개 기준이 잡음 + 10개 수용(MU-74~77, MU-110~113, MU-121·122). 이터레이션 4 가 새로 이름 붙인 것은 9개(MU-114~122)이고 7개(MU-114~120)는 기준이 잡는다.

#### 점검 목록 (이 에이전트가 실행해 읽은 것)

- 개수: `python3 <scratchpad>/iter4/xcheck4.py` — 요구 `spec.md` 24·`spec-compact.md` 24, 수용 기준 `acceptance.md` 24·`spec-compact.md` 24; 기준별 "필수 시험 N개" 의 이름 수가 N 과 같고 green-path 의 `-run` 가지 이름 집합과 일치(AC-022 는 5; 17개 기준 불일치 0); 필수 시험 이름이 plan.md 에 모두 있다; 분류 출시 차단 20·조건부 1·회귀 가드 3; 변이 번호 MU-1~122 에 빠진 번호 없음.
- 미해결 질문 표지: SPEC 디렉터리 여덟 `*.md` 에 `NEEDS[ -]CLARIFICATION` 을 `python3` 로 세어 모두 0. `status:` 줄은 `spec.md` 에만 1(다른 일곱 파일 0); `decision-index.md` 에 `status:` 가 없다.
- 교차 일관: 증인 3 의 형태(SHA 목록)·제품 경로 필터·게이트 판독의 `A` 형태·시험 개수·변이 개수를 spec·plan·design·acceptance·decision-index·spec-compact 에서 대조했다. 남은 `internal cmd .claude` 는 변경 이력 표의 옛 값 둘뿐이고 남은 `0.3.0` 은 HISTORY 의 0.3.0 행과 이 문서의 이전 반복 기록뿐이다.
- 추적성: 요구 24개 모두 `**Covers**` 줄에서 하나 이상 인용된다(새 REQ·AC 를 더하지 않았다).
- 순서 점검: 증인 1~5, plan §D·F.2·F.12, 완료 정의 #2 를 직접 읽어 M0 → 제품 커밋 순서가 plan 과 acceptance 에서 같음을 확인했다(증인 3 이 SHA 목록을 읽는다는 점도 세 곳이 같다).

#### 공백 (관측하지 못한 것)

1. green-path 셀의 통과 출력은 **모양**일 뿐이다 — 새 시험 이름이 아직 없어 어느 것도 실행하지 않았다.
2. 이 카드 자신의 B 에 대한 증인 1~5 는 M0 이전이라 실행하지 못했다(스크래치 저장소와 대역 이력으로 질의 형태만 관측). `--full-history` 의 병합 선택은 이 머신의 `git` 2.54.0 에서만 관측했다.
3. `TestTodoGraphViewDoesNotWaitOnQueueLock` 의 통제(락 아래 기본 보기)와 2초 상한의 충분성은 관측하지 못했다(위).
4. 핀의 `T`·`A` 분해(333/76/28)는 `python3` 줄 단위 읽기이고 원장 행이 아니다. 제목 줄에 `absorb` 가 든 76개 가운데 착지로 읽히는 제목 셋은 제목 모양만 읽은 추정이다.
5. 출하 허브 목록과 사용자 프로젝트의 경로 겹침, 카드→파일 간선의 로컬/CI 일치, 허브 목록의 완전성, 어휘 이웃의 정밀도, 되돌려진 t1453 착지는 측정하지 못했다(`spec.md` §G). 큐 쪽 baseline 수치와 스냅숏 위 figure 35개는 다시 재지 않았다(M0 의 몫).
6. 설치된 `moai` 와 서빙 중인 MCP 빌드는 트리 HEAD 보다 오래돼 판정에 쓰지 않았다. `go test` 는 `-list`, `TestACCounterFullCorpusMatchesBaseline`, `TestRelateAndUnrelateRefusals`(G24·G25), `TestCardIDBaselineHasNoStaleEntries` 에만 썼고 `go test ./...` 는 돌리지 않았다.

#### 이터레이션 4 트리 상태

**최종 점검(모든 편집 뒤).** 원장 재실행 112행 일치·불일치 0(`ledger4.py`, G24·G25 는 따로), `spec lint` 종료 0(`0 error(s), 0 warning(s)`), `--strict` 종료 0(같은 출력), 요구 24·수용 기준 24(`xcheck4.py`), 미해결 질문 표지 0, `spec.md` 를 뺀 일곱 파일의 `^status:` 줄 0, `spec.md` 는 `version: "0.4.0"`·`status: draft`.

작업 시작 시 `git rev-parse --short HEAD` 는 `8fff427eb`, 브랜치 `WT-card-issuance-overlap-graph` 였고, 모든 편집을 마친 뒤 마지막으로 읽은 `git status --short` 는 SPEC 디렉터리의 수정 파일 일곱(`acceptance.md`, `decision-index.md`, `design.md`, `plan.md`, `progress.md`, `spec-compact.md`, `spec.md`)뿐이다(`research.md` 는 바꾸지 않았다). HEAD 는 여전히 `8fff427eb` 이고 커밋·푸시·SPEC 디렉터리 밖 파일 변경은 없다(스크래치 파일은 scratchpad 와 `/tmp/moai-t1454` 빌드뿐이다).
