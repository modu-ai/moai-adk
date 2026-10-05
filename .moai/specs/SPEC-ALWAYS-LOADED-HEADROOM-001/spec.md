---
id: SPEC-ALWAYS-LOADED-HEADROOM-001
title: "always-loaded 지시문 표면의 허용 제거 풀 A_adm 실측과 150,000자 런타임 한도 달성 가능성 판정"
version: "0.4.0"
status: completed
created: 2026-09-27
updated: 2026-09-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "CLAUDE.md, AGENTS.md, .claude/rules/moai, internal/template/templates, internal/cli"
lifecycle: spec-anchored
tags: "always-loaded, instruction-budget, admissible-pool, measurement, verdict, binding-clause-freeze, escalation"
tier: M
related_specs: [SPEC-ALWAYS-LOADED-DIET-002, SPEC-ALWAYS-LOADED-DIET-001]
---

# SPEC: always-loaded 지시문 표면의 허용 제거 풀 A_adm 실측과 150,000자 한도 달성 가능성 판정

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-09-27 | 0.4.0 | run·sync 완료, 상태 `completed`. run 커밋 `6a03d5272`(판정 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE`, S_init 18·17 집합과 S_live 모두, AC 10/10 PASS), sync 커밋 `fcae46594`, sync-audit PASS-WITH-DEBT 90.6(`.moai/reports/t1226/sync-audit.md`). 본문 무변경. 동결 해제 여부는 운영자 게이트로 남는다. | manager-develop |
| 2026-09-27 | 0.4.0 | plan-audit 3회차 PASS-WITH-DEBT 0.86(`.moai/reports/t1226/plan-audit-iter3.md`)의 필수 채무 셋을 run 착수 전에 `acceptance.md` 기존 AC 안의 검사 줄로 흡수했다(AC 개수 10·ID 불변, REQ 불변). **DEBT-1**: `ADMIT` 행 `chars ≤ gross`, M1·M1p 는 `chars == gross`(AC-ALH-003 (3)), M2 는 증거의 `pre_chars`·`post_chars` 로 `chars == pre − post` 이고 `pre == gross`(AC-ALH-004). M1p 증거에 `dup_source` 줄. **DEBT-2**: 절 행과 그 문단 행이 동시에 `ADMIT` 이면 FAIL 하는 `OVERLAP` 검사(감사자 awk). **DEBT-3**: `ADMIT` M1 목적지가 표면 계수 집합 안이면 FAIL 하는 `DESTIN` 검사, 17집합 귀결 한 줄. 선택 채무 일부: 18경로 목록 내용 고정(`PIN-live-OK`), `missing_init` 줄로 누락 경로 줄 수 규칙과 §D.3 충돌 해소, 하네스의 `MOAI_DISTRIBUTE_ALL` 비움과 `BH ≠ HEAD` 재실행 절차, `research.md` 옛 REQ 번호 정정. | manager-spec |
| 2026-09-27 | 0.3.0 | plan-audit 2회차 FAIL 0.77(`.moai/reports/t1226/plan-audit-iter2.md`) 수리 — 최종 수리 회차. **D1**: AC-ALH-008 을 `git log --first-parent --no-merges` 로 교체. 1회차 감사가 처방한 `--no-merges` 단독 형식은 develop 흡수로 들어온 다른 카드의 비병합 커밋까지 세므로 틀렸고(감사자 양성 대조 53행), 0.2.0 은 그 처방을 그대로 따랐다 — 결함의 출처가 처방이었음을 여기 기록한다. REQ-ALH-013 도 「카드 계보의 비병합 커밋」으로 고쳤다. **D2**: `total_init`·`hash_init`·`hash_live` 를 하네스·파이프라인 재실행으로 재현하는 검사와 파일별 `Σ gross == wc -m` 분할 완결성 검사 추가. **D3**: `net-negative` 기각에 M1·조건 1~4 충족·`pointer_chars ≥ gross` 증거 요구. **D4**: 하네스가 init 프로젝트 트리 전체를 내보내고, 런타임 관측은 격리된 `CLAUDE_CONFIG_DIR`·`HOME` 에서(관측자 사용자 범위 지시문 혼입은 가설로 표기). **D5·D6**: `count_set_init` 줄과 `count-set-init.txt` 가 집합 기준을 한 곳에서 정하고 AC-ALH-006·007 이 그것을 소비, 17집합 기계 줄과 재계산 추가. **D7**: Tier M 요구사항 상한 16 에 맞춰 옛 REQ-ALH-014(폐기 수치 격리)를 REQ-ALH-001 에 병합하고 옛 015~017 을 014~016 으로 다시 매겼다(0.2.0 행과 `progress.md` 대응표의 번호는 당시 번호다). 선택 D8~D12 도 반영(`[REF]` 표시, 하네스 재실행 절차·`harness_head`, `moai … init` 정규식 확장과 자기 신고 한계, REQ 층에서 헬퍼 함수명 제거, golangci-lint). | manager-spec |
| 2026-09-27 | 0.2.0 | plan-audit 1회차 FAIL 0.62(`.moai/reports/t1226/plan-audit.md`) 수리. 차단 결함 D1~D16 과 선택 결함 D17~D21 을 모두 반영했다. 요지: 런타임이 보고하는 계수 집합과의 대조 요구 신설(REQ-ALH-016, D1). `S_init` 산출을 워크트리 가드와 양립하는 커밋된 격리 하네스로 고정하고, 비격리 `moai init` 실행을 금지(REQ-ALH-017, D2). 측정 뒤에만 참이 되는 기계 줄 요구(D3·D4). 백엔드 줄을 관측값 형식으로 변경(D5). 후보 표에 `gross`·`gov`·`dest` 열 추가, 조건 1 에 선행 SPEC 의 단서 (a)/(b) 승계, 목적지 누적 수용량, 기각 사유 열거, `UNTRIED` 행의 상한 정의(D6~D9·D15·D17). `P절` 재조정 선택지를 넷으로 확장하고 F 를 원 트리의 서술값으로 격리(D10·D11). `/tmp` 금지를 증거 인용으로 한정(D12). `sec.py` 를 커밋하고 sha256 고정(D13). 무수정 검사를 카드 자신의 비병합 커밋으로 한정(D14). 표면별 동결 해시를 실측값으로(D16). 탐욕 해제 집합 명명(D18), 서문·yaml 처리와 UTF-8 로케일 고정(D19). | manager-spec |
| 2026-09-27 | 0.1.0 | 최초 작성. 카드 t1226(Tier M · 클래스 C). `SPEC-ALWAYS-LOADED-DIET-002`(t1175)가 채무로 넘긴 두 미결 — 허용 풀 `A_adm` 미측정(`acceptance.md §AC-ALD2-001.3`)과 `P절` 재조정 미결(`design.md §4.0`) — 을 이 카드가 측정으로 닫고, 그 결과로 150,000자 런타임 한도의 달성 가능성을 판정한다. 기준선은 오케스트레이터가 이 트리(`7fe658815`)에서 잰 `199111 total` 이다. | manager-spec |

---

## §A. 배경

Claude Code 런타임은 always-loaded 지시문 파일의 합계가 150,000자를 넘으면 매 세션 경고를 낸다. `SPEC-ALWAYS-LOADED-DIET-002`(t1175)는 18경로 합계를 기준선 246,943자에서 줄였으나 한도에는 닿지 못했고, 그 잔여를 **성격 미확정 채무**로 이 카드에 넘겼다. 성격이 확정되지 않은 이유는 하나다 — 구속 조항 동결(REQ-ALD2-002·003, 동결 해시 `d97b33d9…c6c3`, 170줄) 아래에서 실제로 더 들어낼 수 있는 양, 곧 허용 풀 `A_adm` 이 측정되지 않았다. t1175 가 확립한 것은 `A_adm ≤ 73,126`(공표값 계열)이라는 상한과, 원 트리 `172ef22eb` 에서 계산된 구조적 하한 F 의 세 값 `{170,528 · 171,695 · 172,863}` 뿐이다.

### 이 트리의 기준선 — 오케스트레이터 실측, 이 실행

- 기준 커밋: `7fe658815eb0d4110b9acadad56e5a85bee3ed3f` (`git rev-parse HEAD`, 깨끗한 트리). t1175 병합을 포함한 로컬 `develop` 이다.
- 명령: `SPEC-ALWAYS-LOADED-DIET-002/acceptance.md §AC-ALD2-001` 의 18경로 `wc -m` 블록 그대로(이 SPEC `acceptance.md` AC-ALH-002 에 전문 재수록).
- 관측: `199111 total`. 150,000 에 대한 잔여 **49,111**.
- 동결 다중집합 sha256(같은 트리, manager-spec 재실행): `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3` — t1175 기준선과 동일.

카드 본문의 197,897(병합 전 트리 `3a48485af`)과 198,361(t1175 HEAD `8e50ef148`)은 **다른 트리의 값**이다. 이 카드는 그 둘을 폐기 수치로 기록만 하고 어떤 산술에도 쓰지 않는다.

### 계량기가 두 벌이다 — 도그푸드 표면과 배포 표면

같은 18경로라도 이 저장소의 라이브 사본과 `moai init` 이 사용자 프로젝트에 까는 사본은 다르다. 템플릿은 `AGENTS.md` 를 `AGENTS.md.tmpl` 로, `user.yaml`·`language.yaml` 을 `.tmpl` 로만 갖고 있어 렌더링 뒤에야 실물이 생긴다(미렌더링 원본 합계 203,611 — `research.md §1`). 카드 본문은 「상시 로드 **템플릿** rules 풀」을 말하고, t1184 는 격리된 `moai init` 트리에서 같은 경고(18파일 249.2k, t1175 감축 이전 트리)를 관측했다. 사용자가 보는 경고는 배포 표면에서 난다. 이 SPEC 은 두 표면을 모두 재고 판정을 표면별로 낸다(REQ-ALH-002).

### 18경로가 런타임의 계수 집합과 같다는 것은 가설이다

18경로 가운데 `workflow/skill-routing.md` 는 두 트리 모두 `paths:` frontmatter(쉼표로 이은 문자열 값)를 가진 path-scoped 룰이다(4,966자, plan-audit 재측정). 런타임이 이 형태의 `paths:` 를 존중해 그 파일을 세지 않는지는 **관측되지 않았다.** 존중한다면 계수 대상은 17파일이고 잔여가 달라진다. 이 SPEC 은 런타임이 보고하는 파일 수·합계를 관측해 18경로 합계와 맞대라고 요구한다(REQ-ALH-015).

관측에 관한 가설 둘을 더 적는다. 둘 다 이번 plan 단계에서 관측하지 않았다. (1) 18경로 밖에도 init 트리에 상시 로드 파일이 있을 수 있다 — 그래서 하네스는 18경로가 아니라 init 프로젝트 트리 전체를 내보낸다(REQ-ALH-016). (2) 관측자의 사용자 범위 지시문(예: `~/.claude/CLAUDE.md`)이 런타임 합계에 섞일 수 있다 — 그래서 관측은 격리된 `CLAUDE_CONFIG_DIR`·`HOME` 에서 한다(REQ-ALH-015).

---

## §B. 용어

| 기호 | 뜻 |
|---|---|
| `S_init` | 이 트리의 코드로 격리된 홈에서 `moai init` 을 실행해 얻은 프로젝트 트리의 계수 집합(`count_set_init`, 기본 18경로) 표면 — **사용자가 보는 표면, 한도 판정의 1차 대상**. 산출 방법은 REQ-ALH-016 으로 고정 |
| `S_live` | 이 저장소 라이브 사본의 18경로 표면(도그푸드) — 기준선 199,111 |
| 후보 | always-loaded 표면에서 제거·이동·압축을 검토하는 절, 첫 제목 앞의 서문, 또는 구속 절 안의 비구속 문단 하나 |
| M1 / M1′ / M2 | companion 재배치 / 비구속 산문 중복 제거 / 제자리 압축(`SPEC-ALWAYS-LOADED-DIET-002 §B` 정의 승계) |
| 조건 1 | 후보 안에 구속 조항 줄(`[HARD]` / `MUST` / `shall `)이 0줄이고, **그리고** 후보가 (a) 구속 조항이 의존하는 표·제목·열거가 아니며 (b) 구속 조항의 범위를 좁히는 비구속 문장이 아니다(`SPEC-ALWAYS-LOADED-DIET-002/plan.md §C` M1 본문의 단서 승계) |
| 조건 2~4 | 같은 절의 M1 조건 — 2 역방향 인용 생존, 3 목적지 `paths:` 도달, 4 목적지 40,000자 수용량 |
| `A_adm` | 조건 1~4·REQ-ALD2-011·REQ-ALD2-013·표면 동결 해시를 모두 지키며 제거 가능한 자수의 합 |
| `R` | 제거에 따라 stub 에 새로 들어가는 포인터 줄의 자수(재유입) |
| `U` | 시도하지 못한 후보(`UNTRIED`)의 `gross` 자수 합 — 그 후보들이 줄 수 있는 제거량의 상한 |
| `T_min` | `현재 합계 − A_adm + R` — 동결 아래 달성 가능한 최저 합계. `UNTRIED` 후보를 전부 제거 가능하다고 가정한 낙관값은 `T_min − U` |
| F | t1175 가 원 트리 `172ef22eb` 에서 수율 외삽을 포함한 선행 산식으로 계산한 구조적 하한 — **역사적 서술값이며 이 카드의 `T_min`·판정 토큰의 입력이 아니다** |

---

## §C. GEARS 요구사항

- **REQ-ALH-001** (Ubiquitous) — The 판정서 `.moai/reports/t1226/verdict.md` shall 첫 20줄에 레인 백엔드 줄(형식 `레인 백엔드: <관측값> (출처: <관측 방법>)`), 기준 커밋 `7fe658815eb0d4110b9acadad56e5a85bee3ed3f`, 기준선 `199111 total` 과 잔여 49,111 을 적고, 다른 트리·다른 시점에서 잰 수치를 이 트리의 측정값으로 쓰지 않는다 — 197,897 과 198,361 은 폐기 수치로만, F 는 원 트리의 서술값으로만 적는다. 레인 백엔드의 값은 run 레인이 자기 서빙 모델을 관측한 결과이며, plan 단계가 미리 채우지 않는다.
- **REQ-ALH-002** (Ubiquitous) — The 측정 shall `S_init` 과 `S_live` 두 표면을 각각 재고, 표면마다 계수 집합 파일(`count-set-<s>.txt`)·`total_<s>`·후보 표·`T_min`·판정을 따로 적는다. 150,000자 한도 판정의 1차 대상은 `S_init` 이다. `total_<s>` 는 그 표면 계수 집합의 `wc -m` 합계이며, 측정은 UTF-8 문자 집합 로케일에서 수행하고 `locale charmap` 출력을 판정서에 적는다.
- **REQ-ALH-003** (Ubiquitous) — The 후보 표 shall 커밋된 스크립트가 생성하며, 계수 집합의 모든 절·서문·구속 절 안의 비구속 문단을 한 행씩 담고, 행마다 `gross` 자수·허용 자수·기제·구속 줄 개수·조건 1 단서 판정(`gov`)·조건 1~4 판정·목적지(`dest`)·판정·사유·증거 경로를 채운다. 파일마다 서문과 절 행의 `gross` 합은 그 파일의 `wc -m` 과 같다. 두 yaml 파일은 설정 데이터로서 사유 `config-data` 의 기각 행으로 싣는다.
- **REQ-ALH-004** (Where) — **Where** 후보의 파일이 `AGENTS.md` 인 경우, the 후보 표 shall 그 후보를 M2 로만 허용하고 M1·M1′ 로 허용하지 않는다(REQ-ALD2-011 승계).
- **REQ-ALH-005** (Unwanted) — The 측정 shall not 구속 조항 줄의 삭제·재작성·재배치를 요구하는 후보나 조건 1 단서 (a)·(b) 에 걸리는 후보를 `A_adm` 에 계상한다 — 같은 의무가 path-scoped 룰에 있다는 이유의 중복 제거(REQ-ALD2-013)도 포함한다.
- **REQ-ALH-006** (When) — **When** 후보가 M2 로 분류될 때, the 측정 shall 그 후보의 허용 자수를 커밋 트리 밖 사본에서의 실제 압축 시도로 재고, 시도 뒤 그 사본의 동결 다중집합이 표면 기준선 해시와 같음을 확인한다. 시도하지 않은 후보는 판정 `UNTRIED` 와 열거된 시도 불가 사유로 적고, 수율 외삽(예: 10.77%)으로 자수를 채우지 않는다.
- **REQ-ALH-007** (When) — **When** 측정이 `P절` 을 확정할 때, the 측정 shall `SPEC-ALWAYS-LOADED-DIET-002/design.md §4.0` 의 절 분할을 원 트리 `172ef22eb` 에서 커밋된 `sec.py` 로 재실행하고, `design.md §4.3` 의 기각 표에서 `J` 의 구성(특히 `kanban-dispatch.md ## Scope — when this rule is live` 1,309 의 포함 여부)을 함께 밝혀, 재조정을 (a)·(b)·공표값·기타 중 하나로 고른다. 결과 F 는 `F(172ef22eb) = N` 형식의 서술값으로만 적는다.
- **REQ-ALH-008** (Ubiquitous) — The 절 분할 스크립트와 후보 표 생성 스크립트 shall `.moai/reports/t1226/` 아래에 커밋되어 재실행 가능하다. `sec.py` 는 원본 sha256 `d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547` 과 같은 사본으로 커밋된다. 판정서의 증거 인용은 커밋된 경로만 가리키며, 임시 디렉터리 경로는 `$SCRATCH` 로 표기한다.
- **REQ-ALH-009** (Ubiquitous) — The 판정서 shall 표면마다 `current_<s>`·`A_adm_<s>`·`R_<s>`·`U_<s>`·`T_min_<s>` 줄을 적고, 모두 그 표면 계수 집합에 속한 행만으로 재현된다 — `A_adm` 은 후보 표 `ADMIT` 행의 합, `U` 는 `UNTRIED` 행의 `gross` 합, `R` 은 목적지별로 묶은 포인터 줄 표의 합.
- **REQ-ALH-010** (Ubiquitous) — The 판정서 shall 표면마다 판정 토큰 하나를 적는다 — `T_min < 150000` 이면 `ACHIEVABLE`, `T_min − U ≥ 150000` 이면 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE`, 그 사이(`T_min − U < 150000 ≤ T_min`)면 `UNDETERMINED`.
- **REQ-ALH-011** (When) — **When** `S_init` 의 판정이 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 또는 `UNDETERMINED` 일 때, the 판정서 shall 동결 해제 상신 절차를 담는다 — (a) 해제 후보 구속 조항 줄의 목록과 각 줄이 동결시키는 자수, (b) 150,000 에 닿는 탐욕 해제 집합(정의는 `plan.md` M6), (c) 위험(REQ-AMC-002 강등, `AGENTS.md` 자기충족성, 템플릿 중립성), (d) 결정권자는 운영자이며 리드가 `AskUserQuestion` 으로 상신한다는 명시. `UNDETERMINED` 일 때는 `UNTRIED` 목록과 `U` 를 상신 자료에 함께 싣는다.
- **REQ-ALH-012** (Unwanted) — The 레인 shall not 동결 해제 여부를 결정하거나 권고를 결정으로 적는다. 상신 절의 결론 문장은 `RECOMMEND:` 로 시작한다.
- **REQ-ALH-013** (Unwanted) — The 카드 계보의 비병합 커밋 shall not 18개 계수 파일과 그 템플릿 미러를 수정한다. develop 흡수로 들어온 다른 카드의 커밋은 이 요구의 대상이 아니다. 실제 감축은 후속 카드의 일이다.
- **REQ-ALH-014** (Ubiquitous) — The 조건 4 판정 shall 목적지마다 `현재 크기 + Σ(그 목적지로 가는 ADMIT M1 행의 허용 자수) < 40000` 을 라이브·템플릿 두 트리 각각에서 검사하며, 이미 40,000자 이상인 파일을 목적지로 허용하지 않는다(REQ-ALD2-014·015 승계).
- **REQ-ALH-015** (Ubiquitous) — The 판정서 shall `S_init` 의 계수 집합을 `count_set_init` 줄(`18` · `17` · `observed`) 하나로 고정하고 그 경로 목록을 `count-set-init.txt` 에 적는다. 런타임이 스스로 보고하는 always-loaded 파일 수와 합계는 사용자 범위 지시문이 없는 격리 설정(`CLAUDE_CONFIG_DIR`·`HOME`)에서 `S_init` 전체 트리 사본을 대상으로 관측하고, 격리 여부를 출처와 함께 적는다. 관측되면 관측 집합(`observed`)을 쓰고, 관측되지 않으면 그 사실을 Gap 으로 적고 18경로(`18`)를 쓰되 `skill-routing.md` 를 뺀 17경로 집합의 `T_min`·`U`·판정도 함께 적어, 두 판정이 다르면 `verdict_init` 을 `UNDETERMINED` 로 둔다.
- **REQ-ALH-016** (Ubiquitous) — The `S_init` 산출 shall 운영자의 실제 홈에 쓰지 않는 경로로만 이루어진다 — 실제 홈의 네 쓰기 지점(user-scope settings, profile ledger, 비공개 MoAI 홈, 셸 rc)을 임시 디렉터리로 돌리고 실행 전후 실제 홈이 바뀌지 않았음을 검증하는, 이 트리에 커밋된 하네스가 고정 플래그로 init 을 실행하고 산출된 프로젝트 트리 전체를 내보낸다. 격리되지 않은 `moai init` 실행은 금지된다.

---

## §D. 범위 밖

### Out of Scope — 18개 계수 파일의 실제 감축

- 이 카드는 측정과 판정만 한다. `CLAUDE.md`·`AGENTS.md`·두 yaml·14개 룰 파일과 그 템플릿 미러를 고치지 않는다.
- 이 카드가 추가하는 Go 코드는 `S_init` 산출용 격리 하네스 테스트 한 파일뿐이며, 제품 코드는 바꾸지 않는다.
- 판정이 `ACHIEVABLE` 이면 그 감축은 후속 카드로 발행을 요청한다. 발행은 리드의 일이다.

### Out of Scope — 동결 해제의 결정과 집행

- 동결 해시 `d97b33d9…c6c3` 의 해제 여부는 운영자가 정한다. 이 카드는 상신 자료를 만들 뿐 결정하지 않고, 해시를 바꾸지 않는다.

### Out of Scope — 출력 스타일과 토큰 가드

- `.claude/output-styles/moai/moai.md` 와 `internal/config/token_budget_guard.go` 의 `AlwaysLoadedTokenBudget` 은 다른 계량기다. 이 카드의 어떤 AC 도 그것을 재지 않는다.

### Out of Scope — 파일당 40,000자 초과 4건의 수리

- 이미 한도를 넘은 path-scoped 룰 4개의 수리는 카드 t1180 소관이다. 이 카드는 그 파일들을 목적지 후보에서 제외하는 데만 쓴다.

### Out of Scope — 런타임의 `paths:` 해석 규칙 자체

- 런타임이 쉼표 문자열 `paths:` 를 어떻게 해석하는지 코드나 문서로 확정하는 일은 하지 않는다. 이 카드는 런타임의 보고값을 관측할 뿐이다(REQ-ALH-015).

---

## §E. 성공 기준

- 두 표면 모두에서 후보 표가 스크립트로 재생성되며, `A_adm`·`U`·`R` 이 표에서 재현된다.
- `P절` 재조정이 네 선택지 중 하나로 닫히고 `F(172ef22eb)` 가 서술값으로 적힌다.
- `S_init` 과 `S_live` 각각에 판정 토큰 하나가 규칙대로 붙고, 런타임 계수 집합과의 대조 결과 또는 그 Gap 이 적힌다.
- 카드가 작성한 커밋은 18개 계수 파일과 그 템플릿 미러를 건드리지 않는다.

---

## §F. 교차참조

- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/spec.md` — REQ-ALD2-001(개정 3), REQ-ALD2-002·003·011·013·014·015
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/acceptance.md` — §AC-ALD2-001, §AC-ALD2-001.1(F 세 값), §AC-ALD2-001.3(`A_adm ≤ 73,126`), AC-ALD2-002(동결 해시)
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/plan.md §C` — M1 조건 1~4 와 조건 1 단서
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/design.md` — §1, §4.0, §4.3, §4.4
- `.moai/reports/t1175/verdict.md`, `sync-audit.md`, `sync-reaudit.md`
- `.moai/reports/t1226/plan-audit.md` — 1회차 감사
- `internal/cli/init_home_guard_test.go` — `prepareSafeInitHome`
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — 귀속 요구
