---
id: SPEC-ROLE-LOAD-PREDICATE-001
title: "인수 기준 — Codex 역할 로드 판별식의 계약 중립화 (승계 노선, 축소판)"
version: "0.5.0"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, role-load, acceptance-criteria, predicate-design"
---

## §A 읽는 법과 공통 규약

> **축소판이다.** plan-audit iter-3 `FAIL 0.75`에 리드의 확장 종료 조건이 걸려, 감사자의 범위 축소안을 적용했다. 이 파일에는 **AC 일곱**만 있고 번호 **005·007·010·011은 없다** — 후속 카드 후보로 이관됐고 목록과 각 절단의 대가는 `spec.md` §F가 적는다. 이관된 번호는 재사용하지 않는다.

- **결과 어휘.** `NOT_RUN` = 실행되지 않았거나, 실행되었어도 판정 대상 성질이 측정되지 않음. `INVALID` = 리드가 측정 결함(하네스·픽스처 결함)으로 기록한 실행. `FAIL` = 측정되었고 판정식이 `true`가 아님.
- **[HARD] 결과 규약 — 실패는 비영 종료로 말한다.** 이 문서의 **모든** 판정 명령은 참일 때 exit 0, 거짓일 때 **비영 종료**로 끝난다. `jq -se` 기반 명령은 그 성질을 이미 갖고, 그렇지 않은 명령은 `{ echo false; exit 1; }`로 닫는다(iter-1 D10).
- **[HARD] 진단은 판정보다 먼저 인쇄한다.** 적중 수를 세는 명령은 각 카운트를 `|| true`로 감싸 먼저 확정하고 `printf`로 출력한 **뒤에** 판정한다. `&&` 사슬 안에서 `grep -c`가 무적중 exit 1을 내면 약속한 진단이 인쇄되지 않고 「0 적중」과 「파일 부재」가 같은 출력으로 보인다(iter-2 E13).
- **판정 산출물 경로.** 모든 중간 파일은 `.moai/reports/t1171/` 아래에 쓴다. `/tmp` 고정 경로를 쓰지 않는다(iter-1 D24).
- **테스트 이름.** run 단계에서 만들 이름이다. 이름을 바꾸면 이 파일의 명령도 함께 고친다. 이름이 없으면 `pass` 수가 0이 되어 FAIL이다.
- **[HARD] `skip`·`fail` 검사는 지정 테스트로 한정한다.** 패키지 전체에 `skip == 0`을 걸지 않는다 — `./internal/spec`에는 구조적 skip이 넷 있고(`TestACCounterBaselineRegenerate` 등), 그 중 첫째의 기본 꺼짐은 `TestACRegenerationGateDefaultsOff`가 불변으로 지킨다. 패키지 전역 `skip == 0`은 그 사실과 상호 배제되어 어떤 구현으로도 만족될 수 없다(iter-2 E1).
- **빈 입력 금지.** 모든 판정식은 이벤트 0개·증거 파일 부재·빈 결과집합에서 `true`를 내지 않는다. 각 AC는 그 성질을 확인하는 **양성 대조**를 함께 지정한다. run 단계는 판정식 전수를 빈 입력으로 한 번씩 돌려 출력과 종료 코드를 `.moai/reports/t1171/plan-checks/empty-input.md`에 남긴다.
- **환경 정리.** `internal/cli`·`internal/spec` 명령은 kanban·factory 환경 변수를 **같은 호출 안에서** 지운다(`unset … && go test …`).
- **계산값 금지.** 명령에는 실행 중에 계산한 값을 `git`·`go` 명령의 인자로 넘기는 형태를 쓰지 않는다(워크트리 세션 가드가 거부한다 — 실측). 카드 분기점은 리터럴 `0356e8117`이다.
- **[HARD] 한 호출에 `git`과 명령 치환이 공존하면 거부된다(iter-3 F9 → iter-4 G3).** iter-3 감사자가 `AC-RLP-009`의 원문 한 호출을 두 번 발행해 두 번 거부됐고, iter-4 감사자가 **그 2단계 분할도** 세 형태로 발행해 세 번 거부한 뒤 방아쇠를 격리했다: `git log` 둘만 있는 호출은 통과하고(`cd`도 방아쇠가 아니다), `git`과 `$( … )`가 한 호출에 함께 있으면 거부된다. 따라서 이 항은 「복잡도」라는 막연한 사유가 아니라 **공존 조건**이며, `AC-RLP-009`의 판정은 git 단계와 치환 단계를 물리적으로 가른 **네 단계로 발행**한다(§B). 단일 종료 코드 규약은 네 단계 모두 exit 0일 때만 `true`로 읽는 것으로 대체하고, 그 접속이 기계가 아니라 절차임을 AC 본문이 함께 적는다.

### §A.0 픽스처 선별 술어 (SSOT)

원본 디렉터리 `.moai/reports/t1143/m4-live-evidence/ac012-sessions/`는 기록 **26개**를 담는다. 이 SPEC이 쓰는 하위 에이전트 기록은 그중 `session_meta.payload.source.subagent.thread_spawn.agent_role` **표찰을 가진 12개**이며, 선별은 **그 표찰만** 본다. 나머지 14개는 표찰 없는 부모 세션이고 어느 역할 본문과도 일치하지 않으므로 실물 음성 대조(AC-RLP-003 N5)로 쓴다.

[HARD] 본문이 기대표와 일치하는지를 선별 조건으로 쓰지 않는다 — 그러면 양성 갈래가 구성상 항진이 된다(REQ-RLP-015, iter-1 D13).

[HARD] **전면 교체만 막는 것으로는 부족하다.** 「표찰 보유 **그리고** 본문 일치」로 구현해도 이 26개에서는 표찰 보유 12개가 전부 자기 역할과 일치하므로 선별 수가 그대로 12가 되어 혼합 구현이 통과한다(iter-2 E12). 그래서 AC-RLP-003의 S 갈래는 **표찰을 가지면서 본문이 어긋나는** 합성 기록 둘(N3·N4)을 선별 대상 집합에 넣고, 그 둘이 **선별에 포함됨**을 요구한다 — 표찰만 보는 구현은 포함하고 본문과 결합한 구현은 배제하므로, 그 한 항이 두 구현을 기계로 가른다.

### §A.1 AC ↔ 요구사항 매핑

| AC | 요구사항 | 종류 |
|---|---|---|
| AC-RLP-001 | REQ-RLP-003 | 결정적 |
| AC-RLP-002 | REQ-RLP-005, REQ-RLP-007 | 결정적 |
| AC-RLP-003 | REQ-RLP-001, REQ-RLP-004, REQ-RLP-007, REQ-RLP-015 | 결정적(픽스처) |
| AC-RLP-004 | REQ-RLP-002 | 결정적(픽스처) |
| AC-RLP-006 | REQ-RLP-010, REQ-RLP-011 | 결정적(타입 반사) |
| AC-RLP-008 | REQ-RLP-012 | 결정적 |
| AC-RLP-009 | REQ-RLP-014 | 결정적 |

**이 카드에 AC가 없는 요구사항 넷 — 숨기지 않고 적는다.**

| 요구사항 | 이 카드의 처분 |
|---|---|
| `REQ-RLP-006`(선고정) | 판정식이 `AC-RLP-007 [RETIRED]`뿐이었고 그것이 이관됐다. LIVE가 범위 밖이므로 막을 대상도 없다 → 후속 카드 후보 |
| `REQ-RLP-008`(LIVE 상한) | 판정식이 `AC-RLP-005 [RETIRED]`뿐이었고 그것이 이관됐다 → 후속 카드 후보 |
| `REQ-RLP-009`(양면 고지) | **산문 의무로 유지**, 검증은 plan-audit·sync-audit의 **읽기**에 위임. 문자열 판정으로는 지킬 수 없음이 세 회차로 측정됐다(`spec.md` §F.2) |
| `REQ-RLP-013`(영구 미충족 선언) | 같음 — 산문 의무 유지, 검증은 감사 읽기에 위임 |

[HARD] 위 넷에 대해 이 카드는 **기계가 지킨다고 주장하지 않는다.** 그 주장을 뺀 것이 축소의 요점이며, 주장을 남긴 채 판정식을 못 갖는 상태가 공허 초록이다.

---

## §B 인수 기준

### AC-RLP-001 — 지문은 TOML 명세 파싱값에서 계산된다 (REQ-RLP-003)

**Given** 여러 줄 문자열(`'''…'''`)로 `developer_instructions`를 담은 역할 TOML 픽스처 셋 — (a) 여는 구분자 직후에 줄바꿈이 있는 것, (b) 없는 것, (c) 본문 안에 `''`를 포함한 것, (d) 키가 없는 것,
**When** 구현의 추출 함수로 각 픽스처의 `developer_instructions`를 얻으면,
**Then** 값이 모두 TOML 명세 파싱 결과와 바이트 동일하고, (a)에서 정규식 추출값과는 **다르며** 그 차이가 정확히 선행 바이트 하나(`\n`)임이 기록되고, (d)는 추출 실패로 보고된다.

**양성 대조.** (a) 픽스처에서 정규식 추출 경로를 쓰는 변이가 이 테스트에서 실패해야 한다.

**음성.** (d)에서 추출이 빈 문자열을 조용히 반환하면 FAIL.

```bash
mkdir -p .moai/reports/t1171 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexRoleBodyFingerprintParse$' -count=1 -timeout=120s > .moai/reports/t1171/ac-rlp-001.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexRoleBodyFingerprintParse")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexRoleBodyFingerprintParse/(leading_newline|no_leading_newline|inner_quotes|missing_key|regex_mutant_rejected)$")))]|length)==5 and ([.[]|select((.Output//"")|test("^PARSE_LEADING_DELTA 1 0x0a$"))]|length)==1 and ([.[]|select((.Action=="fail" or .Action=="skip") and ((.Test//"")|test("^TestCodexRoleBodyFingerprintParse")))]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1171/ac-rlp-001.jsonl
```

---

### AC-RLP-002 — 기대표의 적격성 (REQ-RLP-005, REQ-RLP-007)

**Given** 생성된 `internal/template/templates/.codex/agents/moai/*.toml` 전부와 합성 픽스처 셋 — (a) 두 역할의 본문이 같은 것, (b) 한 역할의 본문이 빈 것, (c) 디렉터리가 빈 것,
**When** 기대표(해시 → 역할 이름)를 만들면,
**Then** 실제 트리에서는 역할 파일 수와 서로 다른 본문 해시 수가 같고 0이 아니며, (a)(b)(c)는 **서로 다른 고유 사유 코드**로 적격성 실패를 보고한다.

**양성 대조.** 세 사유 코드가 서로 다름을 테스트가 `TABLE_DISTINCT_REASONS 3`으로 출력하고 판정식이 그 수를 본다.

**음성.** 역할 파일이 0개인 디렉터리에서 표가 "적격"으로 보고되면 FAIL.

```bash
mkdir -p .moai/reports/t1171 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexRoleBodyExpectationTable$' -count=1 -timeout=120s > .moai/reports/t1171/ac-rlp-002.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexRoleBodyExpectationTable")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexRoleBodyExpectationTable/(real_tree_wellformed|colliding_bodies|empty_body|empty_dir)$")))]|length)==4 and ([.[]|select((.Output//"")|test("^TABLE_DISTINCT_REASONS 3$"))]|length)==1 and ([.[]|select((.Action=="fail" or .Action=="skip") and ((.Test//"")|test("^TestCodexRoleBodyExpectationTable")))]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1171/ac-rlp-002.jsonl
```

---

### AC-RLP-003 — 지문 유도, 선별 술어, 그리고 판별력 (REQ-RLP-001, 004, 007, 015)

**Given** 워크트리 안에 고정된 픽스처 `internal/cli/testdata/codex-rollouts-t1171/` — 원본 26개 기록 **전부**의 사본, 그 실행 시점 역할 TOML 12개 사본(기대표 원본), 그리고 **표찰을 가지면서 본문이 어긋나는 합성 기록 둘**(N3 = 표찰 보유·역할 본문 항목 제거, N4 = 표찰 A·본문 B). 선별 대상 집합은 그래서 **28개**다,
**When** 선별 술어(§A.0 — 표찰 보유)로 고르고, 유도 함수가 각 기록에서 `type == "response_item"`, `payload.role == "developer"`인 항목의 본문 sha256을 모아 기대표와 대조하면,
**Then** 다음 일곱 갈래가 모두 성립한다.

| 갈래 | 입력 | 기대 |
|---|---|---|
| S (선별 — 혼합 구현 판별) | 28개(실물 26 + 합성 2) | 표찰 보유 **정확히 14개**가 선별된다(실물 12 + 합성 2). **N3·N4가 선별에 포함되어야 한다** — 본문 일치와 결합한 구현은 그 둘을 빼고 12를 내므로 이 항에서 실패한다 |
| P (양성) | 선별된 실물 12개 + 같은 버전 기대표 | 각 기록의 일치 역할 집합이 정확히 `{그 기록의 agent_role}`, 12/12 |
| N1 (버전 어긋남) | 실물 12개 + **다른 버전** 트리의 기대표 | 12건 모두 `false` |
| N2 (역할 파일 부재) | 실물 12개 + 빈 기대표 | 12건 모두 `false` |
| N3 (표찰만 있음) | 합성 기록 — 표찰 보유, 본문 항목 제거 | 선별에는 포함되고 판정은 `false` |
| N4 (역할 뒤섞임) | 합성 기록 — 표찰 A, 본문 B | 선별에는 포함되고 판정은 `false` |
| N5 (실물 부모 세션) | 선별에서 빠진 실제 14개 | 선별되지 않고, 강제로 판정에 넣으면 14건 모두 `false` |

N3은 이 AC의 핵심 판별력이다 — `agent_role` **표찰**이 있다는 것은 "codex가 역할 X를 띄웠다고 여긴다"는 뜻일 뿐이고 본문 바이트가 들어갔다는 뜻이 아니다. 그리고 N3·N4가 **선별에는 포함되고 판정에서는 거짓**이라는 짝이 선별 층과 판정 층을 분리해 고정한다.

**양성 대조.** 선별 수(`SELECTED_BY_LABEL 14 OF 28`), 양성 일치 수(`FINGERPRINT_POSITIVE_MATCHES 12`), 부모 세션 거짓 수(`PARENT_SESSIONS_FALSE 14`)를 각각 별도 토큰으로 출력하고 판정식이 셋을 본다. 유도 함수가 항상 `false`를 내는 변이는 P에서, 항상 `true`를 내는 변이는 N1~N5에서, **선별을 본문 일치와 결합한 변이는 S에서** 실패해야 한다 — 세 방향 변이를 테스트 안 변이 표에 둔다.

```bash
mkdir -p .moai/reports/t1171 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexRoleBodyFingerprintDerive$' -count=1 -timeout=180s > .moai/reports/t1171/ac-rlp-003.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexRoleBodyFingerprintDerive")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexRoleBodyFingerprintDerive/(S_selection_by_label|P_all_twelve|N1_version_mismatch|N2_absent_roles|N3_label_without_body|N4_crossed_roles|N5_real_parent_sessions|hybrid_selection_mutant_rejected)$")))]|length)==8 and ([.[]|select((.Output//"")|test("^SELECTED_BY_LABEL 14 OF 28$"))]|length)==1 and ([.[]|select((.Output//"")|test("^FINGERPRINT_POSITIVE_MATCHES 12$"))]|length)==1 and ([.[]|select((.Output//"")|test("^PARENT_SESSIONS_FALSE 14$"))]|length)==1 and ([.[]|select((.Action=="fail" or .Action=="skip") and ((.Test//"")|test("^TestCodexRoleBodyFingerprintDerive")))]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1171/ac-rlp-003.jsonl
```

---

### AC-RLP-004 — 계약 중립성: 거절 의무를 가진 두 역할 (REQ-RLP-002)

**Given** AC-RLP-003과 같은 픽스처의 실물 12개 — 그중 `manager-lead`와 `mission-governor`는 그 실행에서 nonce를 되돌리지 않았다(각각 `LEAD BLOCKED: The delegation named no work.`와 `blocker` 결정 JSON),
**When** 로드 판별식만으로 12개 기록 전부를 같은 함수로 판정하면,
**Then** 12건 모두 `true`이고(두 역할 포함), 같은 12건에서 nonce 회수 결과는 10건만 참으로 남는다(로드 참·행동 거짓의 동시 성립). 판정 함수 입력에 응답 텍스트 필드가 없다는 성질은 AC-RLP-006의 타입 검사가 담보한다.

**양성 대조.** 12건 전부를 같은 함수로 판정했음을 `CONTRACT_NEUTRAL_LOAD_TRUE 12 NONCE_TRUE 10` 한 토큰으로 출력하고 판정식이 본다. 두 역할만 통과하고 나머지가 실패하는 구현은 역할별 예외를 심은 것이므로 FAIL이다.

**음성.** 역할 이름으로 분기하는 변이는 실패해야 한다.

```bash
mkdir -p .moai/reports/t1171 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexRoleLoadPredicateContractNeutral$' -count=1 -timeout=180s > .moai/reports/t1171/ac-rlp-004.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexRoleLoadPredicateContractNeutral")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexRoleLoadPredicateContractNeutral/(manager-lead|mission-governor|all_twelve_same_function|role_name_branch_mutant_rejected)$")))]|length)==4 and ([.[]|select((.Output//"")|test("^CONTRACT_NEUTRAL_LOAD_TRUE 12 NONCE_TRUE 10$"))]|length)==1 and ([.[]|select((.Action=="fail" or .Action=="skip") and ((.Test//"")|test("^TestCodexRoleLoadPredicateContractNeutral")))]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1171/ac-rlp-004.jsonl
```

---

### AC-RLP-006 — 경로 A/B 분리를 구조로 집행한다 (REQ-RLP-010, REQ-RLP-011)

**이 AC는 결정적이며 LIVE 증거를 읽지 않는다.** 행동 필드 목록의 출처는 증거 JSON이 아니라 **Go 타입에 대한 반사**다(iter-2 E5).

**Given** 로드 판별 함수의 입력 타입, 행동 필드를 담는 타입(`behaviour` 구조체), 그리고 그 타입의 필드 목록을 반사로 열거해 만든 합성 증거 사본들 — 각 필드를 단독으로 뒤집은 N개와 "열 역할 중 절반만 행동 참"인 사본 하나,
**When** 타입 반사 검사와 전수 변이를 실행하면,
**Then** 세 가지가 성립한다.

1. **타입 검사** — 로드 판별 함수의 입력 타입에 행동 필드가 없고, 그 타입의 어느 필드 이름에도 `behaviour`·`nonce`가 (대소문자 무시) 들어 있지 않다. 반사로 확인하며, 함수 본문에서 행동값을 읽으면 애초에 컴파일되지 않는다.
2. **전수 변이** — 반사로 센 행동 필드 수와 실제로 변이시킨 수가 **같다**. 테스트는 두 수를 한 토큰에 담아 `BEHAVIOUR_FIELDS_MUTATED <n> OF <n>`으로 출력하고, 판정식은 **두 수의 일치**를 본다(iter-2 E6).
3. **결합 변이 기각** — 로드 판정이 행동 필드 중 **임의의 하나**를 읽도록 만든 변이는 위 사본들 중 최소 하나에서 결과가 갈려 실패하고, 테스트가 갈린 개수를 출력한다.

**양성 대조.** 2의 두 수 일치와 3의 갈린 개수 ≥ 1을 판정식이 각각 본다. 변이가 어느 사본에서도 갈리지 않으면 이 AC는 아무것도 재지 않으므로 FAIL이다.

**음성.** 입력 타입에 행동 필드가 하나라도 있으면 FAIL. 반사 필드 수가 0이면 토큰이 `0 OF 0`이 되어 정규식에 걸리지 않으므로 FAIL.

```bash
mkdir -p .moai/reports/t1171 && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/cli -run '^TestCodexRoleLoadPredicateStructurallyIndependent$' -count=1 -timeout=120s > .moai/reports/t1171/ac-rlp-006.jsonl; jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexRoleLoadPredicateStructurallyIndependent")]|length)==1 and ([.[]|select(.Action=="pass" and ((.Test//"")|test("^TestCodexRoleLoadPredicateStructurallyIndependent/(input_type_has_no_behaviour_field|exhaustive_field_mutations|half_true_arm|coupled_mutant_rejected)$")))]|length)==4 and ([.[]|select((.Output//"")|test("^BEHAVIOUR_FIELDS_MUTATED ([1-9][0-9]*) OF \\1$"))]|length)==1 and ([.[]|select((.Output//"")|test("^COUPLED_MUTANT_DIVERGED [1-9][0-9]*$"))]|length)==1 and ([.[]|select((.Action=="fail" or .Action=="skip") and ((.Test//"")|test("^TestCodexRoleLoadPredicateStructurallyIndependent")))]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0' .moai/reports/t1171/ac-rlp-006.jsonl
```

---

### AC-RLP-008 — 형제 SPEC의 파일을 한 바이트도 고치지 않았다 (REQ-RLP-012)

**Given** 카드 분기점 리터럴 `0356e8117`과 형제 SPEC 두 디렉터리,
**When** 분기점부터 HEAD까지 두 디렉터리의 변경을 열거하면,
**Then** 변경된 파일이 **0개**다.

**받아들이는 오염의 방향 (iter-2 E8 — 선택과 그 대가).** 두 형태가 각각 다른 쪽에서 깨진다. 세 점 `develop...HEAD`(merge-base)는 병합 전에는 정확하나 **병합 뒤 공집합**이 되어 양성 대조가 죽는다. 리터럴 핀은 병합 뒤에도 양성 대조를 유지하나, develop 흡수로 **다른 카드가 형제 두 디렉터리를 건드린 경우 이 카드가 하지 않은 일로 거짓**이 될 수 있다. 이 AC는 **리터럴 핀**을 택하고 후자의 오염을 받아들인다 — 완료 판정은 병합 뒤 리드가 읽으므로 공집합 쪽이 더 비싸다. 오염이 실제로 발생하면 `changed_siblings`에 잡힌 파일의 귀속을 `git log HEAD..<상대 ref> -- <경로>`로 가려 `INVALID`로 기록하고 재판정한다(귀속 질문은 트리 비교로 답하지 않는다 — `gitflow-lane-protocol.md` §8).

**양성 대조.** 명령은 이 카드가 실제로 무언가를 바꾸었음을 함께 출력한다(`changed_total >= 1`). 그 수가 0이면 "형제 무변경"은 부재가 아니라 **미측정**이므로 거짓을 낸다.

**음성.** 두 디렉터리 아래 어떤 파일이든 한 줄 고치면 `false`.

```bash
mkdir -p .moai/reports/t1171 && git diff --name-only 0356e8117..HEAD > .moai/reports/t1171/ac-rlp-008-all.txt && git diff --name-only 0356e8117..HEAD -- .moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001 .moai/specs/SPEC-CODEX-AUDIT-READONLY-001 > .moai/reports/t1171/ac-rlp-008-siblings.txt && awk 'FILENAME ~ /-all\.txt$/{a++;next} {s++} END{printf "changed_total=%d changed_siblings=%d ",a+0,s+0; ok=(a+0>=1 && s+0==0); print (ok?"true":"false"); exit (ok?0:1)}' .moai/reports/t1171/ac-rlp-008-all.txt .moai/reports/t1171/ac-rlp-008-siblings.txt
```

---

### AC-RLP-009 — 이 SPEC의 acceptance.md가 AC 코퍼스 스냅숏에 반영되었다 (REQ-RLP-014)

**Given** 기본 실행에서 도는 코퍼스 대조 테스트 `TestACCounterFullCorpusMatchesBaseline`, 스냅숏 `.moai/reports/t338/ac-count-baseline.txt`(추적됨), 대상 경로 `P = .moai/specs/SPEC-ROLE-LOAD-PREDICATE-001/acceptance.md`,
**When** 그 테스트를 돌려 출력과 스냅숏을 함께 읽으면,
**Then** 아래 네 항이 모두 성립하고, 1단계 산출물의 신선도 두 항(`fresh`·`ordered` — 아래 iter-4 G4 절)도 함께 성립한다.

1. `TestACCounterFullCorpusMatchesBaseline`의 `pass`가 1이고, **그 테스트의** `fail`·`skip`이 0이다.
2. 그 테스트 출력에 `absent-from-snapshot <P>` 줄이 **없다**. 이 테스트는 스냅숏에 없는 파일을 실패가 아니라 **정보 줄로 보고**하므로(실측), 그 줄의 부재가 곧 "이 파일이 스냅숏에 들어갔다"는 신호다.
3. 스냅숏에 `<P>  COUNT ` 로 시작하는 행이 정확히 하나 있다.
4. 카드 범위에서 `P`를 건드린 커밋이 있다면, 그중 **가장 이른 커밋이 스냅숏 파일도 함께 건드린다**(같은 커밋 재생성 — REQ-RLP-014). `P`를 건드린 커밋이 카드 범위에 없으면 이 항은 비어 있음으로 통과한다 — **이 카드의 `P`는 카드 범위 안에서 커밋되므로 이 항이 실제로 구속한다.**

[HARD] **AC 수를 `grep -c '^### AC-RLP-'`와 비교하지 않는다.** 코퍼스 카운터는 파일 안의 **모든** `AC-*` 식별자를 세므로 자기 AC 수와 같지 않다 — 자기 AC 일곱에 인용한 형제 AC 식별자가 더해진다. 판정은 카운터 자신의 출력과 스냅숏 행의 존재로 하고, 수를 손으로 재현하지 않는다. iter-3이 이 자리에 `COUNT 15`를 실측으로 적었으나 그 값은 그 판 기준이었고, 축소로 자기 AC가 열하나에서 일곱으로 줄어 다시 달라진다 — **그래서 이 절은 어떤 수도 단언하지 않는다**(iter-3 F7).

**[HARD] 판정은 네 단계로 발행한다 (iter-3 F9의 2단계 분할을 iter-4 G3이 재분할).** iter-3의 2단계 형태도 이 워크트리 세션에서 **거부됐다** — iter-4 감사자가 세 형태(`cd` 접두 포함 / `cd` 없이 절대경로 / 경로 형태 무관)로 발행해 셋 다 거부하고, 방아쇠를 격리했다: `git log` 둘만 있는 호출은 통과하고, **`git`과 명령 치환(`$( … )`)이 한 호출에 공존하면 거부된다.** 그래서 git을 쓰는 단계와 치환으로 판정하는 단계를 물리적으로 가른다 — **1a(git만) · 1b(go만) · 2a(git만) · 2b(git 없는 판정)**. **네 단계가 모두 exit 0일 때만 이 AC는 `true`다**; 앞 단계가 비영으로 끝나면 뒤 단계를 실행하지 않는다. 이 접속 자체는 사람이 지키는 절차이고(종료 코드 넷을 AND하는 장치는 없다), 그중 기계가 볼 수 있는 부분만 아래 신선도 항이 본다.

**[HARD] 1단계 산출물의 신선도를 판정 항으로 넣는다 (iter-4 G4). 택한 길은 기계화이며, 닫는 것과 닫지 않는 것을 함께 적는다.** iter-4 감사자가 손으로 쓴 한 줄 jsonl을 먹여 2단계가 `true`/exit 0을 내는 것을 실측했다 — `go test`가 한 번도 돌지 않은 상태다. 그래서 1a가 측정 시점 HEAD를 `ac-rlp-009-head.txt`에, 2a가 판정 시점 HEAD를 `ac-rlp-009-judge-head.txt`에 적고, 2b가 `fresh`(두 파일이 비어 있지 않고 내용이 같다)와 `ordered`(jsonl이 `head` 표지보다 새롭다)를 **판정 항으로** 본다.

- **닫는 것:** 다른 커밋에서 만든 산출물의 재사용(`fresh=0`), 그리고 `.moai/reports/t1171/`에 남아 있는 iter-1~3 잔여 파일의 우연한 적중 — 그 파일들에는 head 표지 쌍이 없어 `fresh=0`이다(그 잔여 재고는 `progress.md` 측정 표가 이미 경고한다).
- **닫지 않는 것:** 고의로 위조한 한 벌. 세 파일을 손으로 쓰고 `touch`로 순서를 맞추면 일곱 항이 모두 성립한다. **이 잔여는 절차이고 기계가 아니며**, 위 단계 접속과 함께 `progress.md` Residual-risk에 적는다.

**양성 대조 (같은 2a·2b, `P`만 바꿔 두 방향 관측).** `P`가 스냅숏에 이미 있는 경로(예: `.moai/specs/SPEC-ADVISOR-RUNG-001/acceptance.md`)이면 2·3이 성립해 `true`가 나오고, 아직 없는 이 SPEC의 경로이면 `false`가 나온다. 즉 이 판정식은 "스냅숏 반영 여부"를 실제로 가른다. 4항은 양성 대조 경로에서 비어 있음으로 통과하고 이 SPEC 경로에서 구속한다. **2a는 `P`를 리터럴로 담으므로**(계산값 금지 — §A) 양성 대조를 돌릴 때는 2a의 그 리터럴 경로도 함께 바꿔 발행한다. **예시 경로에 `AC-`를 부분 문자열로 담은 SPEC을 쓰지 않는다** — 카운터 패턴에 걸려 이 파일의 AC 수에 유령 하나가 더해진다(iter-3 F7에서 실측된 형태).

**RED-now와 green path (verification-completeness §2).** 지금은 `absent=1 row=0`으로 `false`이며, 그 사유는 스냅숏이 아직 재생성되지 않았기 때문이다(impossible도 wrong-reason도 아니다). green path는 M3 — 같은 커밋에서 `MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1`을 돌려 스냅숏을 갱신하면 2·3·4가 함께 성립한다.

**음성.** 스냅숏을 되돌리면 `false`. 스냅숏만 갱신하고 다른 커밋에 넣으면 4항이 `false`.

1a단계 (git만 — 측정 시점 HEAD 표지):

```bash
mkdir -p .moai/reports/t1171 && git rev-parse HEAD > .moai/reports/t1171/ac-rlp-009-head.txt
```

1b단계 (go만 — 측정):

```bash
rm -f .moai/reports/t1171/ac-rlp-009.jsonl && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -json ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$' -count=1 -timeout=300s > .moai/reports/t1171/ac-rlp-009.jsonl
```

2a단계 (git만 — 판정 시점 HEAD 표지 + 커밋 범위. 리터럴 인자만 쓰고 명령 치환을 쓰지 않는다):

```bash
git rev-parse HEAD > .moai/reports/t1171/ac-rlp-009-judge-head.txt; git log --reverse --format=%H 0356e8117..HEAD -- .moai/specs/SPEC-ROLE-LOAD-PREDICATE-001/acceptance.md > .moai/reports/t1171/ac-rlp-009-p.txt; git log --format=%H 0356e8117..HEAD -- .moai/reports/t338/ac-count-baseline.txt > .moai/reports/t1171/ac-rlp-009-b.txt
```

2b단계 (git 없는 판정 — `P`(그리고 2a의 리터럴 경로)를 바꾸면 양성 대조가 된다):

```bash
P=.moai/specs/SPEC-ROLE-LOAD-PREDICATE-001/acceptance.md; B=.moai/reports/t338/ac-count-baseline.txt; J=.moai/reports/t1171/ac-rlp-009.jsonl; H=.moai/reports/t1171/ac-rlp-009-head.txt; JH=.moai/reports/t1171/ac-rlp-009-judge-head.txt; pass=$(jq -se '[.[]|select(.Action=="pass" and .Test=="TestACCounterFullCorpusMatchesBaseline")]|length' "$J"); bad=$(jq -se '[.[]|select((.Action=="fail" or .Action=="skip") and .Test=="TestACCounterFullCorpusMatchesBaseline")]|length' "$J"); absent=$(jq -r 'select(.Action=="output")|.Output' "$J" | grep -c -F "absent-from-snapshot $P" || true); row=$(grep -c -F "$P  COUNT " "$B" || true); same=$(awk 'FILENAME ~ /-b\.txt$/{b[$1]=1;next} {if(np==0){first=$1} np++} END{print (np==0 || (first in b)) ? 1 : 0}' .moai/reports/t1171/ac-rlp-009-b.txt .moai/reports/t1171/ac-rlp-009-p.txt); fresh=0; [ -s "$H" ] && [ -s "$JH" ] && cmp -s "$H" "$JH" && fresh=1; ordered=0; [ "$J" -nt "$H" ] && ordered=1; printf 'pass=%s bad=%s absent=%s row=%s same_commit=%s fresh=%s ordered=%s\n' "$pass" "$bad" "$absent" "$row" "$same" "$fresh" "$ordered"; [ "$pass" -eq 1 ] && [ "$bad" -eq 0 ] && [ "$absent" -eq 0 ] && [ "$row" -eq 1 ] && [ "$same" -eq 1 ] && [ "$fresh" -eq 1 ] && [ "$ordered" -eq 1 ] && echo true || { echo false; exit 1; }
```

---

## §C 완료 정의

- §B의 **일곱** AC가 모두 `true`다(모두 exit 0; `AC-RLP-009`는 네 단계 모두 exit 0).
- **형제 SPEC 세 AC(`AC-DHR-012`·`AC-DHR-023`·`AC-CAR-012b`)의 충족은 이 SPEC의 완료 조건이 아니다.** 셋의 상태는 **이 트리에서 실행 불가 (`.moai/reports/t1100/` 부재)** 이며, `AC-DHR-012`·`AC-DHR-023`은 **영구 미충족**으로 남는다(`spec.md` §A.4·§C.6). 판정서는 이 상태를 그대로 적고 실행 결과를 인용하지 않는다.
- **이 카드는 LIVE 측정을 포함하지 않는다.** 새 판별식이 새 실행에서도 참인지의 **재확인**은 후속 카드 후보이며, 그 손실을 `spec.md` §F.1이 적는다. 이 카드의 근거는 nonce 판별식이 거짓을 낸 **그 실행의 실물 기록** 위에서의 12/12다.
- `REQ-RLP-009`·`REQ-RLP-013`의 충족은 **plan-audit·sync-audit의 읽기**가 판정한다(§A.1). 이 카드는 그 둘에 판정식을 두지 않으며, 그 사실을 완료 보고에 적는다.
- 픽스처 사본 26개의 sha256, 선별 결과(12/26), 합성 기록 넷(N1·N2·N3·N4)의 sha256이 `progress.md`에 있다.
- 판정식 **전수**를 빈 입력으로 돌린 기록(명령·출력·종료 코드)이 `.moai/reports/t1171/plan-checks/empty-input.md`에 있다.

## §D 품질 게이트

- `go vet ./internal/cli ./internal/spec` 무출력.
- `golangci-lint run ./internal/cli/... ./internal/spec/...` 신규 지적 0.
- `go test ./internal/cli/... ./internal/spec/...` 통과(패키지의 구조적 skip 넷은 정상 — §A). 전 패키지 판정은 CI 몫이다.
- 재측정 범위에 `./internal/spec`이 포함된다.
