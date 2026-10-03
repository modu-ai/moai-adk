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

- 카드: t1454, Class C, Tier L. SPEC ID `SPEC-TODO-CARD-ISSUANCE-001`.
- 계획 시작 트리: `2de0a2cb6`(전체 SHA `2de0a2cb613b04765a1554f86685a3b48e0be806`, 로컬 develop 팁), 브랜치 `WT-card-issuance-overlap-graph`, 워크트리 `.claude/worktrees/t1454`. RED-now 측정 시점의 워킹 트리는 깨끗했다(`git status --short` 빈 출력).
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `spec-compact.md`, `decision-index.md`(`interview.decision_gate: on`), 이 `progress.md`.
- 개수: 요구 24(모듈 5), 수용 기준 24 = 출시 차단 21 + 회귀 가드 3(AC-TCI-007·009·023), 결정 행 15(FOUNDER 13, EVIDENCE-NEEDED 1, DECIDED 1), 변이 77(기준이 잡는 73, 이유와 함께 수용 4).
- SPEC ID 사전 점검: Bash 로 `[[ "SPEC-TODO-CARD-ISSUANCE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` 를 실행해 `PASS` 를 출력했다. 유일성: `ls -d .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` 부재(`No such file or directory`), `git grep -c "SPEC-TODO-CARD-ISSUANCE-001" -- .moai/specs` 적중 0.
- 기준선 반입: 카드의 git 쪽 헤드라인 수치는 이 트리에서 전부 재현됐고(`research.md` §3.1), 큐 쪽 수치는 스냅숏(2026-10-03 11:45, `last_seq` 1463) 위에서 다시 쟀다(§3.2). 이 계획의 새 측정은 §3.3 과 부록 A. 카드의 scratchpad 는 다른 세션의 `/tmp` 이므로 M0 가 스크립트를 `.moai/reports/t1454/baseline/` 로 복사한다(sha256 은 §3.4). **오케스트레이터에게**: 그 scratchpad(`.../061c4c0e-c21e-4332-8ccc-0d1d51432c49/scratchpad/cards/`)가 정리되기 전에 보존하면 M0 의 재구성 부담이 준다. 이 에이전트는 SPEC 디렉터리 밖에 쓰지 않는다.
- 판단 도구의 출처: SPEC lint 는 이 트리의 HEAD(`2de0a2cb6`)로 빌드한 바이너리(`go build -o <scratchpad>/moai-t1454 ./cmd/moai`)를 **경로로** 호출했다. 설치된 바이너리(`/Users/goos/go/bin/moai`, rc.26)는 커밋 `802a72235` 로 만들어졌고 그 커밋은 HEAD 의 조상이며 HEAD 와 424 커밋 차이가 난다 — 이 판정에 쓰지 않았다.
- 한계: SPEC 디렉터리 밖(큐, 규칙 파일, 코드)은 이 단계에서 건드리지 않았다. 병렬 쓰기가 아니라 한 에이전트가 SPEC 파일을 썼고, 첫 `spec.md` 는 나머지 일곱 파일보다 앞선 별도 호출로 썼다(일괄 병렬 쓰기에서 벗어난 점).

### lint 기록

| 명령 | 종료 | 출력 |
|---|---|---|
| `moai-t1454 spec lint .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` | 0 | `✓ No findings — all SPEC documents are valid` |
| `moai-t1454 spec lint --strict .moai/specs/SPEC-TODO-CARD-ISSUANCE-001` | 0 | `✓ No findings — all SPEC documents are valid` |
| 양성 대조(린트가 붉어질 수 있음): `spec.md` 사본의 `phase:` 를 `plan` 으로 바꿔 `moai-t1454 spec lint <사본 경로>` | 1 | `ERROR FrontmatterPhaseInvalid … phase "plan" is a workflow-stage token, not a release target` 외 `CoverageIncomplete` 경고(사본 디렉터리에 acceptance.md 가 없어서) |

잔여 lint 지적: 없음. 위 두 줄은 SPEC 파일의 **마지막 편집 이후** 다시 돌려 읽은 값이다(처음 한 번은 편집 전에 돌렸고 같은 결과였다). 종료 코드 0 은 린트가 아무것도 훑지 않았다는 뜻이 아님을 양성 대조(세 번째 줄, 종료 1)가 보인다.

### AC 개수 가드

`go test ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$' -count=1 -v` → 종료 0, `--- PASS: TestACCounterFullCorpusMatchesBaseline (25.47s)`, 출력에 `absent-from-snapshot .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/acceptance.md: COUNT 24`(보고만 하고 실패하지 않음 — 기준선 재생성 불필요).

### 트리 상태

마지막 확인에서 `git status --short` 는 `?? .moai/specs/SPEC-TODO-CARD-ISSUANCE-001/` 한 줄뿐이었다(커밋·푸시·SPEC 디렉터리 밖 파일 변경 없음; 오케스트레이터가 커밋한다).
