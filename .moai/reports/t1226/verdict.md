# t1226 — SPEC-ALWAYS-LOADED-HEADROOM-001 판정서

레인 백엔드: claude-opus-5-5[1m] (출처: 이 run 세션 시스템 프롬프트의 모델 식별 문장 "The exact model ID is claude-opus-5-5[1m]" — 런타임 자기 보고이며 API 응답 메타데이터로 교차 확인하지 않았다)
카드: t1226 · Tier M · 클래스 C · 워크트리 `.claude/worktrees/t1226` · 브랜치 `WT-always-loaded-headroom` · run 단계 레인 manager-develop
기준 커밋: `7fe658815eb0d4110b9acadad56e5a85bee3ed3f` (`git rev-parse HEAD`, 깨끗한 트리 — 오케스트레이터 실측)
기준선(`S_live`): `199111 total` — 150,000 에 대한 잔여 49,111. 이 run 에서 `build_head` 트리를 다시 재도 `199111` 이다(아래 S_live)
동결 다중집합 sha256(`S_live`, 170줄, 기준 커밋): `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`
폐기 수치(다른 트리의 값, 산술에 쓰지 않음): 197,897(`3a48485af`) · 198,361(`8e50ef148`)
plan-audit iter3 은 Tier M 상한(plan_audit_tier_ceilings.M = 2) 초과 — 리드 사후 승인
킥오프: `CLAUDE.local.md §31` 자율 승인(운영자 정책). 동결 해제 여부는 카드 본문이 명시한 운영자 게이트로 남는다
증거 인용 규칙: 커밋된 `.moai/reports/t1226/` 경로만 인용한다. 세션 임시 디렉터리는 `$SCRATCH` 로 적는다
명령 기록: `.moai/reports/t1226/commands.log` (자기 신고)

## 요약

| 표면 | total | A_adm | R | U | T_min | T_min − U | 판정 |
|---|---:|---:|---:|---:|---:|---:|---|
| `S_init` (18경로) | 203,413 | 14,201 | 477 | 79 | 189,689 | 189,610 | STRUCTURALLY-INFEASIBLE-UNDER-FREEZE |
| `S_init` (17경로, `skill-routing.md` 제외) | 198,447 | 13,717 | 454 | 79 | 185,184 | 185,105 | STRUCTURALLY-INFEASIBLE-UNDER-FREEZE |
| `S_live` (18경로) | 199,111 | 13,716 | 477 | 79 | 185,872 | 185,793 | STRUCTURALLY-INFEASIBLE-UNDER-FREEZE |

세 판정 모두 `T_min − U ≥ 150000` 이다. 동결 아래에서는 `UNTRIED` 행을 전부 지워도 150,000 에 닿지 않는다. 사용자 표면(`S_init`)에서 동결 해제 없이 한도에 닿으려면 시도한 M2 행 56,179자 중 48,868자(87.0%)를 의무·조건·수치를 지우지 않고 없애야 한다. 이번 실측 수율은 16.3%(9,179자)다.

## 측정 환경

charmap = UTF-8
build_head = 05d79c8b27b1aacfc56946672e5187b0fcabf96e
harness_sha256 = 8776d7b5809b47a6215a79497f6036414ceb2967d32067db612d07289b49e9c5

- `locale charmap` → `UTF-8`.
- `build_head` 는 하네스 테스트 커밋이다. 하네스 실행 직전 `git rev-parse HEAD` 가 이 값이었고 `harness-run.txt` 첫 줄 `harness_head` 와 같다. 이후 커밋은 `.moai/reports/t1226/` 만 바꿨다.
- `git diff --quiet 7fe658815 05d79c8b2 -- CLAUDE.md AGENTS.md <yaml 2개> .claude/rules/moai/core .claude/rules/moai/workflow internal/template/templates` → 종료 코드 0. 기준 커밋 이후 계수 파일·미러에 변화가 없으므로 기준선 199,111 이 그대로 유효하다.
- `develop` 은 run 시작 시점에 `c501bd1da` 로 앞서 있었다(`git rev-list --count --left-right develop...HEAD` → `3 7`). 흡수하지 않았다.

## S_init

count_set_init = 18
missing_init = -
total_init = 203413
hash_init = 93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477
runtime_observed = no
total_init_17 = 198447
current_init_17 = 198447
A_adm_init_17 = 13717
R_init_17 = 454
U_init_17 = 79
T_min_init_17 = 185184
verdict_init_17 = STRUCTURALLY-INFEASIBLE-UNDER-FREEZE

- 산출: `TestHeadroomInitSurfaceExport`(`internal/cli/init_headroom_export_test.go`)를 AC-ALH-009 명령 그대로 `build_head` 에서 실행했다. `--- PASS: TestHeadroomInitSurfaceExport (0.54s)`. 출력 원문은 `.moai/reports/t1226/harness-run.txt`. 하네스는 `prepareSafeInitHome` 으로 실제 홈의 네 쓰기 지점을 돌리고, `MOAI_DISTRIBUTE_ALL` 을 비운 뒤 `--non-interactive --root <t.TempDir()> --name headroom-probe --language go --mode tdd --llm claude` 로 init 을 실행해 `.git` 을 뺀 프로젝트 트리 전체를 `$SCRATCH/init-surface` 로 내보냈다.
- 18경로 존재: 하네스 `t.Log` 의 `headroom-path` 18줄이 모두 `present` → `missing_init = -`.
- `total_init`: `(cd $SCRATCH/init-surface && xargs wc -m < .moai/reports/t1226/count-set-init.txt | tail -1)` → `203413 total`.
- `hash_init`: AC-ALH-004 파이프라인을 `$SCRATCH/init-surface` 에서 실행 → `93de7321…4477`, 구속 줄 `170`. 라이브 해시와 다른 것은 결함이 아니다(plan D3) — 템플릿 판 규칙 문구가 라이브와 다르다.
- RED 증거: 복사 함수를 미구현 상태로 둔 첫 실행 출력 `init_headroom_export_test.go:94: export: copyHeadroomTree: not implemented` / `--- FAIL` 은 `.moai/reports/t1226/harness-red.txt`.
- 런타임 계수 집합(AC-ALH-010): 형태 2 로 닫는다. 레인은 워크트리 밖에서 격리 `CLAUDE_CONFIG_DIR`·`HOME` 으로 `claude` 를 띄울 수 없다. 18집합과 17집합의 판정이 같으므로 `verdict_init` 은 18집합 규칙 토큰 그대로이고 `verdict_init_reason` 줄은 두지 않는다. 리드에게 보내는 관측 요청은 아래 Gaps 에 원문으로 적었다.

## S_live

total_live = 199111
hash_live = d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3

- `git archive 05d79c8b2 CLAUDE.md AGENTS.md .moai/config/sections .claude/rules/moai | tar -x -C $SCRATCH/live` 뒤 `(cd $SCRATCH/live && xargs wc -m < .moai/reports/t1226/count-set-live.txt | tail -1)` → `199111 total`.
- 파이프라인 → `d97b33d9…c6c3`, 구속 줄 `170`. 선행 SPEC 기준선과 같다.

## P절 재조정

P절 재조정: (공표값)
J_includes_kanban_scope = yes
F(172ef22eb) = 171695

- 재실행: `git archive 172ef22eb CLAUDE.md AGENTS.md .claude/rules/moai | tar -x -C $SCRATCH/orig` → `(cd $SCRATCH/orig && python3 .moai/reports/t1226/sec.py 3)`. 출력 원문 `.moai/reports/t1226/sec-172ef22eb.txt`. `sec.py` sha256 `d0e61541…78547`(원본과 일치).
- 산술은 `.moai/reports/t1226/psection.py` 가 재현하고 출력은 `.moai/reports/t1226/evidence/p-section.txt` 에 있다. 핵심 줄:
  - 재실행 절 풀(L1 서문 제외, 구속 0 절) 합계 `93641`. 15개 파일 칸이 `design.md §4.0` 표와 정확히 같고, 다른 것은 kanban 하나다 — 재실행 `7724`, 표 칸 `6415`.
  - 재실행 kanban `## Scope — when this rule is live` = `1309`(b=0, L2).
  - `design.md §4.3` J = `12026`(자수가 적힌 13행), kanban Scope 행을 포함한다.
- 결론: 공표 합계 93,641 이 원 트리의 실제 절 풀이다. 표의 kanban 칸 6,415 는 Scope 1,309 를 미리 뺀 전사값이고, 그 1,309 는 J 에 한 번 들어 있다. J 는 Scope 를 한 번만 빼므로 (b) 의 「이중 차감」은 성립하지 않는다. (a) 의 열 합 92,332 는 Scope 를 두 번 빼는 값이다. 따라서 공표값 행이 맞다.
- `F(172ef22eb) = 246943 − 73126 − 10522 + 8400 = 171695`. M1 상한 `93641 − 12026 − 8489 = 73126`, M2 상한 `round(97721 × 3092/28717) = 10522`.
- F 는 원 트리의 서술값이며 M2 항에 수율 외삽(10.77%)을 포함한다. 이 카드의 `T_min`·판정 토큰의 입력이 아니다.

## A_adm 후보 표

- 표: `.moai/reports/t1226/candidates-init.tsv`(411행), `.moai/reports/t1226/candidates-live.tsv`(407행). 보조: `dest-sizes.tsv`, `pointers-init.tsv`, `pointers-live.tsv`.
- 생성: `candidates.py`(분할) → `build.py`(판단 층 `judgments.tsv` + M2 시도 `m2/replacements*.txt` + 포인터 `pointers.txt` 결합). 재실행하면 TSV 와 증거 파일이 같은 내용으로 다시 나온다.
- 분할 규칙: fence 밖 `#`~`###` 제목 기준 평면 분할(서문·yaml 은 한 행), 파일별 행 `gross` 합 = `wc -m`. 구속 줄이 있는 절은 그 안의 비구속 문단을 `¶n` 행으로 따로 싣는다.

| 판정·사유 | init 행 | init gross | live 행 | live gross |
|---|---:|---:|---:|---:|
| REJECT `bind>0` | 99 | 138,984 | 99 | 139,355 |
| REJECT `governs-scope`·`narrows-scope`·`citation-breaks` | 181 | 63,315 | 181 | 61,694 |
| REJECT `config-data` | 2 | 1,483 | 2 | 227 |
| ADMIT M1 (chars = gross) | 18 | 5,022 | 18 | 5,022 |
| ADMIT M2 (시도, chars = 실제 감소량) | 110 | 56,179 → 9,179 | 106 | 54,609 → 8,694 |
| UNTRIED (`untried:machine-marker`) | 1 | 79 | 1 | 79 |

(`gov` 기각 행 gross 에는 `bind>0` 절 안의 문단 행이 들어 있어 두 줄의 합은 total 과 같지 않다. 분할 행만의 합이 total 이다.)

- **M1 판단 기준**: REQ-AMC-002 가 재배치를 허용하는 부류(근거·절차·예시·사건 기록·교차참조)만 M1 후보로 삼았다. 키워드가 없어도 규범 문장인 행은 M2(제자리 압축)만 허용했다. 목적지는 기존 companion 중 두 트리 모두 40,000자 미만인 것만 썼다. `verification-claim-integrity-detail.md` 는 self-keyed(`**/verification-claim-integrity*.md`)라 조건 3 을 만족하지 못해 쓰지 않았다.
- **ADMIT M1 18행**: acp 3(→ `agent-common-protocol-reference.md`, 누적 588자, 두 트리 39,227 / 39,194), askuser 5(→ `askuser-protocol-reference.md`), mcp-tools 4(→ `moai-mcp-tools-catalogue.md`), cross-session 2(→ `cross-session-messaging-detail.md`), kanban 3(→ `kanban-dispatch-mechanics.md`), skill-routing 1(→ `skill-routing-detail.md`). 포인터는 목적지당 한 줄로 묶었고(설계 결정 A), 원문은 `evidence/pointers/<s>/*.txt`.
- **M2**: 시도 대상 행마다 압축본을 `$SCRATCH/post-<s>` 사본에 적용했다. 모든 시도를 적용한 사본의 동결 해시가 표면 해시와 같다 — `post_hash_init = 93de7321…4477`, `post_hash_live = d97b33d9…c6c3`. 사본 합계는 `189212`(init) / `185395`(live)이고, 각각 `total − A_adm` 과 같다. 행별 전후 원문과 `pre_chars`·`post_chars`·`post_hash` 는 `evidence/m2/<s>/<id>.md`.
- **조건 1 단서(`gov`)**: `design.md §4.3` 기각 목록의 절 가운데 이 트리에 남은 것은 모두 `§4.3` 분류 그대로 a/b 로 두었다. 새 판단의 근거는 행마다 `judgments.tsv` `note` 열과 `evidence/rejects.md` 에 있다. `§4.3` 이 M1 만 기각한 절(REQ-ALD2-013 의무 블록 5개, `## 9` import 줄을 담은 절)은 M2 로 시도했다.
- **목적지 누적 수용량**: `dest-sizes.tsv`(두 트리 `wc -m`)와 AC-ALH-003 (4) 검사로 확인했다.

## T_min Claim

current_init = 203413
A_adm_init = 14201
R_init = 477
U_init = 79
T_min_init = 189689
current_live = 199111
A_adm_live = 13716
R_live = 477
U_live = 79
T_min_live = 185872

- 재현: `build.py` 출력 `surface=init total=203413 A_adm=14201 R=477 U=79 T_min=189689`, `surface=init_17 total=198447 A_adm=13717 R=454 U=79 T_min=185184`, `surface=live total=199111 A_adm=13716 R=477 U=79 T_min=185872`. AC-ALH-006 이 TSV 에서 같은 값을 다시 계산한다(`.moai/reports/t1226/ac-verify.md`).
- `current_<s>` 는 그 표면 계수 집합의 현재 합계이고 `total_<s>` 와 같다.

## 판정

verdict_init = STRUCTURALLY-INFEASIBLE-UNDER-FREEZE
verdict_live = STRUCTURALLY-INFEASIBLE-UNDER-FREEZE

- `S_init`: `T_min − U = 189689 − 79 = 189610 ≥ 150000`. 17집합도 `185184 − 79 = 185105 ≥ 150000` 로 같은 토큰이다.
- `S_live`: `185872 − 79 = 185793 ≥ 150000`.
- t1175 가 넘긴 잔여 채무의 성격이 이것으로 확정된다. 이 트리에서 동결을 지키는 한 150,000 은 구조적으로 닿지 않는다. 노력 부족이 아니다.

## 동결 해제 상신 절차

`verdict_init` 이 STRUCTURALLY-INFEASIBLE-UNDER-FREEZE 이므로 아래 네 항목을 운영자 판단 자료로 올린다. 레인은 해제 여부를 판단하지 않는다.

### (a) 해제 후보 구속 조항 줄과 줄마다 묶인 자수

- 대상은 REJECT `bind>0` 인 99개 절, 구속 줄 170개 전부다. 줄 원문은 절별로 `.moai/reports/t1226/evidence/bind-init.md` 에 있다.
- 한 절의 모든 구속 줄을 해제했을 때 새로 허용되는 자수는 그 절의 `gross` 에서 이미 허용된 그 절 문단 행의 `chars` 를 뺀 값이다. 여러 줄이 한 절을 묶으면 그 줄이 모두 해제될 때만 센다(겹침 규칙). 계산은 `unfreeze.py`, 결과는 `.moai/reports/t1226/unfreeze-init.txt`.
- 전체: 99개 절, 170줄, 새로 허용되는 자수 합 131,257.

### (b) 탐욕 해제 집합

해제 줄 1개당 새로 허용되는 자수가 큰 순서로 절을 더해, `T_min − 누적 < 150000` 이 되는 첫 지점에서 멈춘 집합이다. 줄 수 기준 최소를 보장하지 않는다.

| 순위 | 줄 | 새로 허용 | 누적 | 파일 · 절 |
|---:|---:|---:|---:|---|
| 1 | 1 | 2,281 | 2,281 | `AGENTS.md` · 5. Core behaviors |
| 2 | 1 | 2,150 | 4,431 | `core/agent-common-protocol.md` · Verbatim batch, output contracts, and CLI idioms |
| 3 | 1 | 2,089 | 6,520 | `core/agent-common-protocol.md` · Hook Invocation Surface |
| 4 | 1 | 1,901 | 8,421 | `workflow/kanban-dispatch.md` · Card classes |
| 5 | 1 | 1,752 | 10,173 | `CLAUDE.md` · Selection Decision Tree |
| 6 | 1 | 1,711 | 11,884 | `core/verification-claim-integrity.md` · 3.1 Refused-tool degradation |
| 7 | 1 | 1,631 | 13,515 | `workflow/session-handoff.md` · Output Surface (User-Facing) |
| 8 | 1 | 1,617 | 15,132 | `workflow/cross-session-messaging.md` · A send result has three shapes |
| 9 | 1 | 1,612 | 16,744 | `workflow/context-window-management.md` · Context Window Targets |
| 10 | 1 | 1,523 | 18,267 | `core/moai-mcp-tools.md` · The `project_root` input |
| 11 | 1 | 1,500 | 19,767 | `core/agent-common-protocol.md` · Ledger Closure |
| 12 | 1 | 1,499 | 21,266 | `CLAUDE.md` · 14. Parallel Execution Safeguards |
| 13 | 2 | 2,828 | 24,094 | `workflow/session-handoff.md` · Field-by-Field Specification |
| 14 | 1 | 1,384 | 25,478 | `workflow/session-handoff.md` · Auto-Memory Integration (Mandatory) |
| 15 | 1 | 1,377 | 26,855 | `core/verification-claim-integrity.md` · 3. The 5-Section Evidence-Bearing Report Format |
| 16 | 2 | 2,721 | 29,576 | `core/agent-common-protocol.md` · Pre-Edit Sync Check |
| 17 | 1 | 1,359 | 30,935 | `workflow/session-handoff.md` · Canonical Format (Verbatim Spec) |
| 18 | 1 | 1,296 | 32,231 | `workflow/main-checkout-branch-guard.md` · Rules |
| 19 | 2 | 2,561 | 34,792 | `core/moai-constitution.md` · Lessons Protocol |
| 20 | 1 | 1,245 | 36,037 | `core/verification-claim-integrity.md` · 2.3 Ordering attribution |
| 21 | 1 | 1,225 | 37,262 | `workflow/kanban-dispatch.md` · Kanban Dispatch Protocol |
| 22 | 1 | 1,208 | 38,470 | `workflow/session-handoff.md` · When To Generate (5 Triggers) |
| 23 | 1 | 1,205 | 39,675 | `workflow/kanban-dispatch.md` · Completion is read, never trusted |
| 24 | 1 | 1,163 | 40,838 | `core/agent-common-protocol.md` · Subagent Prohibitions |

- 필요량 `189689 − 149999 = 39,690`. 24개 절, 구속 줄 27개, 새로 허용 40,838자에서 `T_min` 이 148,851 이 된다.
- 이 수치는 해제된 절 전체를 목적지로 옮길 수 있다는 가정(아래 U1)을 둔 상한이다. 옮긴 절마다 포인터가 새로 생기며 그 재유입은 이 표에 없다(Gaps).
- **27줄은 하한이다(sync-audit F3).** 포인터 재유입과 목적지 수용량을 반영하면 필요한 줄 수는 늘어난다. 근거는 아래 두 가지다.
- 순위 1 `AGENTS.md` 는 U1 만으로는 옮길 수 없고 U3 도 함께 허용해야 한다. `AGENTS.md` 행을 빼고 U1 만 허용하면 집합은 여전히 24절·27줄이지만, 얻는 자수가 39,692 여서 `T_min` 이 149,997 에 그친다(`unfreeze.py` 를 `AGENTS.md` 행을 뺀 `candidates-init.tsv` 에 돌린 출력 `T_min_after = 149997`. 감사자 재계산과 같다). 여유는 3자이고, 옮긴 절마다 생길 포인터(24개 이상)의 재유입만으로도 150,000 을 넘는다.
- 이 집합에서 `core/agent-common-protocol.md` 가 내놓는 자수는 9,623 이다(순위 2·3·11·16·24 의 합). 그런데 기존 companion `core/agent-common-protocol-reference.md` 는 이미 38,639자(`wc -m`, 라이브·`build_head` 동일. 템플릿 미러 38,606자)여서 40,000자 상한 안에 9,623자를 받을 수 없다. 새 companion 이 필요하다. sync-audit F3 은 이 파일을 39,227자로 적었으나 이 run 에서는 재현되지 않았다(`wc -m` 38,639, `wc -c` 38,887). 어느 값이든 결론은 같다.

### (c) 위험

- **U1 — 구속 줄의 companion 재배치 허용**: 위 탐욕 집합이 전제하는 형태다. REQ-AMC-002 강등에 해당한다 — companion 이 로드되지 않는 턴에는 그 의무가 사라진다. 순위 1 의 `AGENTS.md` 는 U3 없이는 옮길 수 없다.
- **U2 — 문구는 두고 하드랩만 풀기(바이트 동결 → 의미 동결)**: 줄 바뀜만 달라지므로 동결 해시가 바뀐다. 문단 압축 수율이 오를 수 있으나, 그 양은 이 카드가 재지 않았다(Gaps). 의미 보존 검사가 기계화되지 않은 영역이 커진다.
- **U3 — `AGENTS.md` 에 M1 허용**: 파일의 자기충족성 선언과 비-Claude 하네스의 읽기 경로를 훼손한다(REQ-ALD2-011 의 근거).
- **템플릿 중립성**: `S_init` 해제는 템플릿 미러를 바꾸므로 `template-neutrality-check` 와 미러 동등 검사(REQ-ALD2-007)를 다시 거쳐야 한다.
- **동결 해시 교체**: 어느 형태든 `d97b33d9…c6c3`(라이브)·`93de7321…4477`(`S_init`) 기준선이 바뀐다. 선행 SPEC AC-ALD2-002 의 기준이 함께 움직인다.
- **동결을 풀지 않는 두 번째 경로 — 조건 1 단서 풀의 제자리 압축(sync-audit F2)**: REQ-ALH-005 는 조건 1 단서 (a)·(b) 에 걸린 63,315자(`governs-scope`·`narrows-scope`·`citation-breaks` 181행)를 M1 뿐 아니라 M2 제자리 압축에서도 제외한다. 선행 SPEC 은 이 단서를 M1 에만 적용했다(`SPEC-ALWAYS-LOADED-DIET-002/plan.md:80`, "옮기지 않는다"). 이 풀에 M2 를 허용하면 기존 M2 풀 56,179자와 합친 119,494자에 대해 균일 수율 40.9% 에서 `T_min` 이 150,000 아래로 내려간다(`(189689 + 9179 − 150000) / (63315 + 56179) = 0.409`). 이 카드가 잰 수율은 16.3%(9,179 / 56,179)다. 제자리 압축은 구속 줄을 건드리지 않으므로 동결 해시가 바뀌지 않는다. 따라서 이 경로는 동결 해제가 아니라 범위 해석의 문제이고, 운영자에게 올릴 또 하나의 선택지다. 40.9% 에 닿는지는 재지 않았다.

### (d) 결정권자

- 해제 여부는 운영자가 정한다. 리드가 리드 창에서 `AskUserQuestion` 으로 상신한다. 레인은 결정하지 않고 권고만 적는다.

RECOMMEND: 운영자에게 네 선택지를 함께 올리는 것을 권고한다 — (1) 동결 유지와 150,000 한도 경고 수용, (2) U1 범위를 위 탐욕 집합(27줄은 하한)으로 한정한 해제 검토, (3) U2 수율을 먼저 재는 후속 측정 카드, (4) 동결 해시를 건드리지 않고 조건 1 단서 풀에 M2 제자리 압축을 허용하는 범위 해석과 그 수율(목표 40.9%, 실측 16.3%)을 재는 후속 측정 카드(sync-audit F2). 레인 권고 순서는 (3)·(4) 측정 → (1)/(2) 선택이며, 어느 것도 레인이 정하지 않는다.

## Gaps

- **런타임 계수 집합 미관측(AC-ALH-010 형태 2)**. 레인은 워크트리 밖에서 격리 `CLAUDE_CONFIG_DIR`·`HOME` 으로 `claude` 를 띄울 수 없다. 리드에게 보내는 관측 요청 원문: 「`.moai/reports/t1226/harness-run.txt` 의 하네스를 `build_head`(05d79c8b2)에서 다시 돌려 `$SCRATCH/init-surface` 를 만든 뒤, 상위 경로에 `CLAUDE.md`·`.claude/` 가 없는 빈 디렉터리로 사본을 옮기고, `CLAUDE_CONFIG_DIR`·`HOME` 을 빈 임시 디렉터리로 돌린 워크트리 밖 세션에서 그 사본을 작업 디렉터리로 `claude` 를 띄워, 기동 시 always-loaded 경고 줄(파일 수·합계)을 원문 그대로 옮겨 주십시오. 결과는 `runtime_files_init`·`runtime_total_init`·`runtime_isolated = yes`·`runtime_source` 줄로 판정서에 반영합니다.」 18·17 두 집합의 판정이 같으므로 이 관측은 판정 토큰을 바꾸지 않는다.
- **탐욕 집합의 포인터 재유입**: (b) 표는 해제된 절을 옮길 때 새로 생길 포인터 자수를 넣지 않았다.
- **U2 수율**: 하드랩 해제가 문단 압축을 얼마나 늘리는지는 재지 않았다.
- **M2 의미 보존의 기계 검사 없음**: 압축본이 의무·조건·예외·수치를 지우지 않았다는 것은 작성자 검토뿐이다(REQ-ALD2-012 준용 검토 항목). 전후 원문은 `evidence/m2/` 에 있다.
- **`MOAI:LEARNED-WORKFLOW` 표지 79자**: 학습 워크플로 작성기가 쓰는 표지라 줄여도 되는지 코드로 확인하지 않았다(`untried:machine-marker`).
- **`c3` 판정**: companion `paths:` 가 그 내용이 필요한 작업에 도달하는지는 선행 카드가 같은 목적지를 쓴 선례에 기댔다. 런타임 로드를 관측하지 않았다.
- **검증 스크립트 경로**: 워크트리 가드가 awk·루프가 섞인 복합 명령을 거부해, git 이 없는 검사(awk·python)는 `$SCRATCH` 의 bash 스크립트로 돌렸다. git 이 들어가는 검사(`git archive`, `git log`, `git ls-files`)는 스크립트에 넣지 않고 한 줄 명령으로 따로 돌렸다.

## Residual-risk

- **판정은 판단 층과 압축 수율에 기댄다.** `gov` a/b 와 `citation-breaks` 로 기각한 행은 init 181개, gross 63,315자다. 그 전부가 오판이고 전량 지울 수 있다면 `T_min` 은 약 126,374 로 내려간다. 같은 행에 이번 M2 수율(16.3%)을 적용한 추정은 약 179,000 으로 여전히 한도 위다(외삽, 판정 입력 아님). 판정이 뒤집히려면 기각 판단 대부분이 틀리거나, 규범 산문을 87% 가까이 줄일 수 있어야 한다.
- **압축은 한 번, 한 작성자의 시도다.** 다른 작성자는 더 줄일 수 있다. 다만 동결 아래 한도에 닿으려면 시도한 M2 텍스트의 87.0% 를 의미를 잃지 않고 없애야 한다.
- **`commands.log` 는 자기 신고다.** 정규식 검사는 기록된 명령만 본다. 기록에서 빠진 `moai init` 실행은 잡지 못한다. 이 run 은 `moai init` 을 실행하지 않았다.
- **하네스는 init 명령 경로를 테스트 안에서 돈다.** 설치 바이너리의 `moai init` 과 같은 코드이지만, 플래그 파싱 이전 단계(셸 진입점)는 거치지 않는다.
- **레인 백엔드 줄은 런타임 자기 보고다.** 서빙 모델을 API 응답으로 교차 확인하지 않았다.
