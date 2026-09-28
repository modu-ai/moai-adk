# SPEC-ALWAYS-LOADED-HEADROOM-001 — 구현 계획

카드 t1226 · Tier M · 클래스 C · 워크트리 `.claude/worktrees/t1226` · 브랜치 `WT-always-loaded-headroom` · 기준 커밋 `7fe658815`

---

## §A. 맥락

이 카드는 측정과 판정을 한다. 산출물은 판정서 `.moai/reports/t1226/verdict.md`(뼈대는 plan 단계에서 작성), 후보 표 TSV 두 개와 보조 표, 스크립트, 증거 파일, 그리고 `S_init` 산출용 격리 하네스 테스트 한 파일이다. 18개 계수 파일과 그 템플릿 미러는 바뀌지 않는다(REQ-ALH-013).

기준선(오케스트레이터 실측, 이 트리): `199111 total`, 잔여 49,111, 동결 해시 `d97b33d9…c6c3`.

v0.2.0 에서 plan-audit 1회차(`.moai/reports/t1226/plan-audit.md`)의 결함 D1~D21 을, v0.3.0 에서 2회차(`plan-audit-iter2.md`)의 D1~D12 를 반영했다. 대응표는 `progress.md §E.1`.

---

## §B. 결정 지점 — 되돌리기 비싼 순서

### D1. 판정 대상 표면 — `S_init` 을 1차로, `S_live` 를 병기 (기본값 채택, 리드 재정 가능)

산출물에서 답이 나온다고 판단해 해소가 필요한 clarification 마커를 달지 않았다. 근거:

- 카드 본문이 「상시 로드 **템플릿** rules 풀」을 대상으로 적었다.
- 경고는 사용자 세션에서 나고, t1184 가 격리된 `moai init` 트리에서 같은 경고를 관측했다(`SPEC-ALWAYS-LOADED-DIET-002/spec.md §A`, 249.2k — t1175 감축 **이전** 트리의 값이므로 재측정한다).
- 템플릿은 `AGENTS.md` 와 두 yaml 을 `.tmpl` 로만 갖고 있어, 렌더링된 트리가 아니면 사용자가 받는 실물을 잴 수 없다(`research.md §1`, 미렌더링 원본 203,611).

한도 판정은 `S_init` 에 대해 내리고, t1175 채무의 직접 승계를 위해 `S_live` 도 같은 방법으로 잰다. 리드가 다르게 정하면 M1 전에 이 절만 고치면 된다.

### D2. `S_init` 산출 경로 — 커밋된 격리 하네스 테스트 (plan-audit D2)

워크트리 격리 가드는 인라인 `HOME=` 지정을 거부한다(plan-audit 재측정). 그런데 `runInit` 은 실제 홈의 네 곳 — user-scope settings(`userHomeDirFn`), profile ledger(`profile.BaseDirOverride`), `MOAI_HOME`, 셸 rc(`project.ConfigureShellEnvFn`) — 에 쓴다(`internal/cli/init_home_guard_test.go:6-13`, `internal/cli/init.go:923`, `internal/core/project/initializer.go:673`). `moai init` 에는 전역 쓰기를 건너뛰는 플래그가 없다(`moai init --help` 확인). 그래서 CLI 바이너리를 직접 돌리는 길은 격리가 불가능하거나 가드와 충돌한다.

채택한 경로: 기존 테스트 헬퍼 `prepareSafeInitHome`(세 홈 seam 을 `t.TempDir()` 아래로 돌리고, 셸 rc seam 을 카운팅 스파이로 바꾸고, 실제 홈 지문을 전후 비교한다)를 쓰는 격리 하네스 테스트 `TestHeadroomInitSurfaceExport` 를 `internal/cli/` 에 커밋한다. 테스트는 `-args -headroom-export=<dir>` 테스트 플래그가 있을 때만 돌고(없으면 `t.Skip`), 고정 플래그로 init 을 실행해 init 이 만든 프로젝트 트리 전체(`.git` 제외)를 그 디렉터리로 복사한다 — 18경로 밖의 상시 로드 파일이 있는지 런타임 관측이 드러낼 수 있도록 하기 위해서다(plan-audit 2회차 D4). 환경변수 인라인 지정이 필요 없으므로 가드와 양립한다. 고정 플래그와 계약은 `acceptance.md` AC-ALH-009.

대안(채택하지 않음): 리드가 워크트리 밖 세션에서 `HOME`·`CLAUDE_CONFIG_DIR` 를 돌려 `moai init` 을 실행하는 것. 리드 의존이 생기고 바이너리 출처 증명이 따로 필요하다. 하네스가 실패하면 이 경로를 블로커로 올린다.

### D3. 표면별 동결 해시 — `S_init` 은 실측 (plan-audit D16)

동결 해시 `d97b33d9…c6c3` 는 라이브 16개 마크다운에 대한 값이다. `S_init` 은 같은 파이프라인을 산출 트리에 돌린 `hash_init` 을 기준선으로 쓴다. 두 해시가 다른 것은 결함이 아니다.

### D4. 런타임 계수 집합 (plan-audit D1)

`skill-routing.md` 는 쉼표 문자열 `paths:` 를 가진 path-scoped 룰이다. 런타임이 그것을 세는지는 가설이다. 관측은 레인이 직접 할 수 없을 가능성이 크므로(기동 경고는 새 세션에서만 보인다) 리드에게 관측을 요청한다. 관측은 사용자 범위 지시문이 없는 격리 `CLAUDE_CONFIG_DIR`·`HOME` 에서, 하네스가 내보낸 트리 전체의 사본을 대상으로 한다 — 관측자의 `~/.claude/CLAUDE.md` 같은 지시문이 합계에 섞일 수 있다는 가설 때문이다(관측으로 확인되지 않음). 집합 기준은 판정서 `count_set_init` 한 줄(`18`·`17`·`observed`)과 `count-set-init.txt` 가 정하고, AC-ALH-006·007 이 그것을 소비한다. 답이 없으면 `count_set_init = 18` 로 두고 17집합 기계 줄을 함께 적어, 두 판정이 갈리면 `UNDETERMINED` 로 둔다(AC-ALH-010).

---

## §C. 마일스톤

### M1 — 격리 하네스와 표면 생성 (우선순위: High)

- `internal/cli/` 에 `TestHeadroomInitSurfaceExport` 를 작성한다(`prepareSafeInitHome` + init 명령 실행 + 프로젝트 트리 전체 복사 + 18경로의 (경로, 존재, sha256) `t.Log`). 테스트 플래그가 비면 Skip.
- `acceptance.md` AC-ALH-009 의 명령으로 실행하고 출력을 `.moai/reports/t1226/harness-run.txt` 로 커밋한다. 파일 첫 줄은 실행 직전 `git rev-parse HEAD` 의 `harness_head = <sha>` 다. 같은 명령을 `build_head` 에서 다시 돌리면 `$SCRATCH/init-surface` 가 재현된다(AC-ALH-002·003·004 가 그렇게 재검증한다). 실행한 모든 셸 명령은 `.moai/reports/t1226/commands.log` 에 한 줄씩 남긴다.
- `locale charmap` 이 UTF-8 인지 확인하고 두 표면의 계수 집합 파일(`count-set-<s>.txt`)을 쓰고 `wc -m`·`hash_<s>` 를 잰다. `build_head`·`harness_sha256` 줄을 적는다.
- 리드에게 `S_init` 트리 사본에 대한 격리 런타임 기동 경고 관측을 요청한다(D4).

### M2 — `P절` 재조정 확정 (우선순위: High)

- `sec.py` 는 plan 단계에서 `.moai/reports/t1226/sec.py` 로 이미 커밋했다(sha256 `d0e61541…78547`, 원본과 일치).
- `git archive 172ef22eb <18경로>` 를 `$SCRATCH/orig` 에 풀고 `sec.py` 를 돌려 출력 원문을 `.moai/reports/t1226/sec-172ef22eb.txt` 로 커밋한다.
- `design.md §4.3` 기각 표를 읽어 `J` 가 `kanban-dispatch.md ## Scope — when this rule is live`(1,309)를 포함하는지 적는다(`J_includes_kanban_scope`). 재실행의 kanban 구속 0 절 합과 함께 (a)·(b)·(공표값)·(기타) 중 하나를 고른다.
- F 는 `F(172ef22eb) = N` 로만 적고, M2 항의 수율 외삽을 포함한 서술값이며 판정 입력이 아님을 명시한다.

### M3 — 후보 전수와 조건 1~4 판정 (우선순위: High)

- `candidates.py` 를 작성·커밋한다. 입력은 표면 트리 루트, 출력은 15열 TSV 의 스크립트 산출 열(`file`·`section`·`gross`·`bind`, 나머지는 자리표시). 대상은 16개 마크다운의 서문·절·구속 절 안의 비구속 문단과 두 yaml(`config-data` 기각). `--keys` 모드는 네 열만 낸다.
- 조건 1 단서 `gov`: `design.md §4.3` 기각 목록을 출발점으로 두 표면에서 재확인하고, 새 후보마다 「이 절이 사라지면 어떤 구속 조항의 적용 대상이 달라지는가」를 판단해 근거를 증거 파일에 적는다.
- M1 후보마다 조건 2(역방향 인용 `grep -rIn -F "§ <절 제목>"`), 조건 3(목적지 `paths:`), 조건 4(`dest-sizes.tsv` 기준 누적 수용량, 두 트리)를 건다.
- `AGENTS.md` 후보는 M2 만(REQ-ALH-004).

### M4 — M2 압축 시도 (우선순위: High)

- M2 후보를 `$SCRATCH` 사본에서 실제로 압축하고 `post_hash` 를 잰다. 표면 해시와 같으면 줄어든 자수를 `chars` 로 `ADMIT`, 다르면 `REJECT`(`rewraps-binding-line`).
- 시도하지 못한 후보는 `UNTRIED` 와 `untried:<사유>` 로 적는다. 그 `gross` 합이 `U` 이며, `U` 가 커져도 상신 절차는 피해지지 않는다(REQ-ALH-011).

### M5 — `R`·`T_min`·판정 토큰 (우선순위: Medium)

- `ADMIT` M1 행의 (file, dest) 쌍마다 포인터 줄 원문을 쓰고 `wc -m` 해 `pointers-<s>.tsv` 를 만든다.
- 표면마다 기계 줄(`current_*`·`A_adm_*`·`R_*`·`U_*`·`T_min_*`·`verdict_*`)을 적는다.

### M6 — 동결 해제 상신 절차 (우선순위: Medium · `verdict_init` 이 INFEASIBLE 또는 UNDETERMINED 일 때)

- 구속 조항 줄마다 「그 줄 하나만 해제했을 때 새로 허용되는 자수」를 잰다(그 줄 때문에 `REJECT` 된 후보들의 `gross` 기준).
- **탐욕 해제 집합**: 해제 줄 1개당 새로 허용되는 자수가 큰 순으로 줄을 더해 가다가 `T_min − 누적 허용 < 150000` 이 되는 첫 지점까지의 집합. 여러 줄이 같은 후보를 묶으면 그 후보는 묶는 줄이 **모두** 해제될 때만 허용으로 센다(겹침 처리). 이 집합은 줄 수 기준 최소를 보장하지 않으므로 「최소」라 부르지 않는다.
- 해제 형태를 셋으로 나눠 적는다 — (U1) 구속 줄의 companion 재배치 허용, (U2) 구속 줄 문구는 두고 하드랩만 풀기(바이트 동결 → 의미 동결), (U3) `AGENTS.md` 에 M1 허용. 위험: U1 은 REQ-AMC-002 강등(로드되지 않는 턴에 의무가 사라짐), U3 은 `AGENTS.md` 자기충족성과 비-Claude 하네스 읽기 경로 훼손, 템플릿에 걸리는 해제는 중립성 검사 대상.
- 결론은 `RECOMMEND:` 문장으로만 쓴다. 결정은 운영자, 상신은 리드의 `AskUserQuestion`.

### M7 — 인수 조건 일괄 검증 (우선순위: Low · 기계적)

- AC-ALH-001~010 명령을 한 턴에 병렬로 실행하고 출력을 `.moai/reports/t1226/ac-verify.md` 에 저장한다.
- 이 트리에서 빌드한 바이너리로 `moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001` 재실행.
- 하네스 테스트 파일에 CI 판 `golangci-lint run ./internal/cli/...` 를 돌려 0 issues 를 확인한다.

---

## §D. 제약

- 18개 계수 파일과 템플릿 미러 수정 금지(REQ-ALH-013). 압축 시도는 `$SCRATCH` 사본에서만.
- 비격리 `moai init` 실행 금지(REQ-ALH-016). 설치된 `~/go/bin/moai` 로 `S_init` 을 만들지 않는다.
- 전체 테스트 스위트를 로컬에서 돌리지 않는다. Go 검증은 하네스 테스트 하나(`-run '^TestHeadroomInitSurfaceExport$'`)와, 하네스가 기존 테스트와 충돌하지 않는지 보는 `go vet ./internal/cli/` 와 CI 판 `golangci-lint run ./internal/cli/...`(v2.1.6)로 한정한다.
- 환경 격리가 필요한 실행은 `unset … && <명령>` 한 호출로 한다.
- 수치는 모두 이 트리·이 실행의 명령 출력으로만 적는다. 197,897·198,361 은 폐기 수치, F 는 원 트리 서술값이다.

---

## §E. 위험

| 위험 | 완화 |
|---|---|
| 하네스가 `prepareSafeInitHome` 의 지문 비교에 걸려 실패한다 | 실패 출력을 그대로 커밋하고 D2 대안 경로를 리드에게 블로커로 올린다. 비격리로 우회하지 않는다 |
| `S_init` 트리의 계수 집합이 18경로와 다르다 | `acceptance.md §D.3` 첫 항목과 AC-ALH-010 규칙대로 처리 |
| 런타임 계수 관측이 오지 않는다 | 형태 2 로 닫고, 18·17 두 판정이 갈리면 `UNDETERMINED` |
| 조건 2 역방향 인용 스윕이 커진다 | 절 제목 목록을 한 번에 `grep -F -f` 로 돌리고 증거 파일 하나로 묶는다 |
| `gov` 판정이 사람 판단에 의존한다 | `§4.3` 기각 목록을 출발점으로 하고, 판단마다 근거를 증거 파일에 적는다(AC-ALH-003 검토 항목) |
| 판정이 `UNDETERMINED` 로 끝난다 | 실패가 아니다. 상신 자료에 `UNTRIED` 목록과 `U` 를 싣는다 |

---

## §F. 안티패턴

- 수율 외삽으로 M2 자수를 채우는 것.
- `S_live` 만 재고 사용자 표면의 판정을 내리는 것.
- 격리되지 않은 `moai init` 으로 `S_init` 을 만드는 것.
- `UNTRIED` 를 늘려 상신을 피하는 것.
- 동결 해제를 레인이 결정하거나, 권고를 결정처럼 적는 것.
- 판정서 증거로 커밋되지 않은 임시 경로를 인용하는 것.

---

## §G. 교차참조

- `spec.md`, `acceptance.md`, `research.md`, `progress.md`
- `.moai/reports/t1226/plan-audit.md`
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/` — plan.md §C, design.md §1·§4.0·§4.3·§4.4, acceptance.md §AC-ALD2-001·AC-ALD2-002
- `.moai/reports/t1175/verdict.md`
- `internal/cli/init_home_guard_test.go`, `internal/cli/init_deploy_exit_test.go`(`runInitWithFlags`)
