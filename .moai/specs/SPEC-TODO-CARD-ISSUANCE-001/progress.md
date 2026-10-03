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

- NEEDS-CLARIFICATION 표지: `grep -n "NEEDS CLARIFICATION" <SPEC 파일 여덟 개>` 종료 1(적중 0). 새로 들어오지 않았다.
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
