# 공통 protocol v1 초안 (Issue #1987, 00-D)

이 문서는 MoAI의 stdin/stdout 계약과 상태 어휘를 한 곳에 모은 **초안**이다. P06이 사용자 미결이라 HOLD-00D-1~5는 모두 결정되지 않았고, 이 문서와 산출물은 선택지마다 무엇이 달라지는지를 보여 줄 뿐 어느 쪽도 고르지 않는다. HOLD-00D-3(오류 형태)과 HOLD-00D-4(exit code 표)의 결정 주체는 01-C #1990이며, 00-D는 선택지를 참조용으로만 정리한다.

코드(`src/`)는 바꾸지 않았다. 구현 계획 6단계(Rust 오류 타입·exit 매핑·version 검사)는 선택지가 정해진 뒤에 한다.

> **갱신(01-C #1990, 2026-11-08 사용자 결정)**: HOLD-00D-1·3·4가 결정되어 01-C가 구현했다. 모든 내부 operation은 stdout에 envelope 하나를 낸다. 성공은 `{protocol_version, operation, ok: true, result}`, 실패는 `{protocol_version, operation, ok: false, error{kind, message, retryable, next_actions?}}`이다. exit는 0/1/2를 유지하고, 실패 종류는 `kind` 14종으로 구분한다. 이 문서의 11종에 다음 세 가지를 더한 것이다.
> - `git_failed`(Issue #2033): `identity`에서 git이 "저장소 밖"이 아닌 다른 이유로 실패한 경우. 재시도 불가.
> - `blocked`(07-A #2008, `forge.read`): 인증 실패(401), 권한 거부(403), 또는 재시도 예산을 넘긴 rate limit. 재시도 불가.
> - `transport`(07-A #2008, `forge.read`): gh가 없거나, HTTP 응답이 없거나, 예상하지 못한 상태 코드. 재시도 가능.
>
> 추가는 #1990 HOLD-C2 / REQ-01C-02(r27)의 "필요 시 추가는 docs에 기록" 조항을 따랐고, 기존 11종의 의미는 바꾸지 않았다.
>
> 고른 envelope에는 상태 필드가 없으므로 아래 envelope 후보 A·B(`status`/`domain_status`) 어느 쪽과도 같지 않다. 실패 쪽 모양은 오류 후보 B와 같다. 런처(`-l`, `-a`, `cc|glm|codex`, `session`)는 대상이 아니다. 아래 `현행 계약` 절의 fixture는 이 변경에 맞춰 다시 기록했다. 세부 규칙은 `docs/rust-product-path.md`의 `Output envelope` 절에 있다.

## HOLD 현황

| HOLD | 주제 | 상태 | 이 Issue의 산출물 |
|---|---|---|---|
| HOLD-00D-1 | 상태 어휘 통합 수준과 envelope | envelope 결정됨(B: 모든 출력에 envelope, 상태 필드 없음). 상태 어휘 통합은 미결 | `status-crosswalk.json`의 variant A·B, envelope 후보 A·B |
| HOLD-00D-2 | schema 표현과 검증기 | HOLD 유지 | 현행 serde 타입을 그대로 옮긴 2020-12 schema. 검증기는 고르지 않았다 |
| HOLD-00D-3 | 오류 contract 형태 (01-C 소유) | 결정됨: B(stdout 실패 envelope) + `retryable`·`next_actions?` | 오류 객체 후보 A(stderr), B(stdout envelope) |
| HOLD-00D-4 | exit code 표 (01-C 소유) | 결정됨: A(0/1/2 유지, `kind`로 구분) | A·B·C 표와 관측표 |
| HOLD-00D-5 | version 규칙과 JSON 정규화 | HOLD 유지 | version 규칙 두 안, serde tuple과 JCS 예시 fixture |

## 산출물

| 경로 | 내용 |
|---|---|
| `tests/fixtures/protocol/v1/current/*.json` | 현행 8개 명령의 성공·실패 21건(exit, stdout, stderr). 초안보다 먼저 별도 커밋으로 남긴 기준선이다 |
| `schemas/protocol/v1/status-crosswalk.json` | 7개 상태 어휘 69개 토큰의 변환표, variant A·B |
| `schemas/protocol/v1/{digest,profile,plan,agents-request,agents-plan,package-result}.schema.json` | 현행 타입을 설명하는 JSON Schema 2020-12 |
| `schemas/protocol/v1/variants/` | HOLD별 후보: envelope A·B, 오류 A·B, exit code 표, version·정규화 |
| `tests/fixtures/protocol/v1/variants/hold-00d-5/` | 정규화 두 방식의 예시 바이트 |

`$id`는 `urn:moai:protocol:v1:<이름>` 형식을 제안한다. 소유하지 않은 도메인의 URL을 쓰지 않으려는 것이며, 이 체계도 공유 interface라 확정은 아니다.

## 현행 계약 (측정값, cd6a94418)

`protocol_fixtures_record_current_contract` 테스트가 임시 프로젝트에서 모든 하위 명령을 실행해 결과를 golden 파일과 비교한다. `MOAI_UPDATE_FIXTURES=1`로 실행하면 비교 대신 파일을 다시 쓴다. 두 번 다시 써도 바이트가 같았다(00D-AC-7).

임시 디렉터리 경로와, 그 경로에 묶이는 `target_binding`·`plan_digest`, 그리고 시각인 `collected_at`만 가렸다. `target_binding`은 대상 파일의 절대 경로를 해시한 값이라 같은 입력이어도 디렉터리가 다르면 `plan_digest`까지 달라진다. plan을 대상 위치에 묶으려는 의도된 설계지만, REQ-00D-09가 말하는 "실행 간 동일한 digest"와는 맞지 않으므로 차이로 기록한다.

### 요구와 현행의 차이

| 요구 | 현행 | fixture |
|---|---|---|
| REQ-00D-01 성공 envelope | 01-C 이후 내부 operation은 envelope 안 `result`로 출력. 런처(18)는 envelope 없음 | 06, 09, 11, 13, 16, 17, 18, 21 |
| REQ-00D-02 typed 오류와 kind별 exit | 01-C 이후 stdout 실패 envelope의 `error.kind`. exit는 kind별이 아니라 0/1/2(결정 A). 첫 낱말이 operation이 아닌 사용 오류(01)와 런처 오류(19)는 기존 평문 | 01~05, 07, 08, 12, 14, 15, 19, 20 |
| REQ-00D-03 미지 필드 거부 | 거부는 하지만 malformed JSON과 같은 메시지(`invalid typed JSON input`) | 03, 04 |
| REQ-00D-04 지원 밖 protocol_version | project-apply는 `stale or altered plan`, agents-plan은 `invalid agents request`로 나온다. 05-B #2002의 `plan-admit`만 전용 메시지 `unsupported protocol_version; supported: 1`을 내고, `contract::classify` 규칙으로 `kind: unsupported_protocol_version`이 된다(규칙과 단위 시험은 `src/contract.rs`) | 07, 15 |
| REQ-00D-09 digest 재현성 | plan digest가 절대 경로에 묶인다 | 06, 16 |
| REQ-00D-11 값 echo 금지 | MoAI가 직접 내는 메시지에는 입력 값이 없다(`effort must be one of …`는 허용값 목록). 다만 clap 사용 오류는 입력한 하위 명령 이름을 그대로 되돌려 준다(`unrecognized subcommand 'no-such-command'`) | 01, 19 |

### 추가된 operation (02-C #1993)

`codex.projection_plan`(`codex-projection-plan`)과 `codex.doctor`(`codex-doctor`)는 01-C envelope을 그대로 쓴다. 둘 다 아무것도 쓰지 않는다. 새 `kind`는 없다.

`codex.projection_apply`(`codex-projection-apply`, I4 #2241)도 같은 envelope을 쓴다. 쓰는 곳은 `--target` 아래의 계획된 파일뿐이다. 새 `kind`는 없다. 오류는 기존 kind로 분류된다(`codex_projection_apply_messages_have_a_deliberate_kind`가 고정한다).

| 상황 | kind |
| --- | --- |
| permission reference 없음 | `permission_reference_missing` |
| `--plan-digest`가 새로 계산한 plan과 다름 | `stale_plan` |
| source manifest 불일치·누락(REQ-I4-01) | `conflict` |
| 대상 파일 바이트가 다름 | `conflict` |
| 첫 파일을 만든 뒤 실패 | `partial` |

- 사용하는 kind는 `usage`, `invalid_input`, `not_found`, `io`이다.
- 메시지와 kind의 대응은 `tests/codex_adapter.rs`의 `codex_messages_have_a_deliberate_kind`가 고정한다.
- 위 golden fixture 21건은 00-D가 기록한 기존 명령의 기준선이라 이 두 operation을 담지 않는다.
- 두 operation의 출력은 `tests/codex_adapter.rs`가 검사하고, `result` 키는 `docs/rust-product-path.md`의 operation 표에 있다.

## 상태 변환표

7개 어휘를 모았다. Issue 본문은 6벌이라고 적었지만 실제로는 설치 lifecycle(02:148), VERIFY verdict(07:29), attempt(11:44-64), SYNC(06:67), Issue 상태(승인 계획 사본 267행), issue-report(22:590), migration(14:100)의 7개다. 키는 (도메인, 토큰)이다. `blocked`·`unknown`·`authorized`처럼 여러 도메인에 나오는 말은 도메인마다 따로 행을 둔다.

- variant A: 14개 값의 단일 enum. 69행 중 3행(`unsupported`, `partial`, `inherited_uninspected`)은 대응시키지 않았다. 대응은 했지만 의미가 줄어드는 행에는 `loss`를 적었다.
- variant B: 도메인 토큰은 그대로 두고, 상위 공통 상태(`ok|blocked|unknown|failed|not_run`)만 붙인다. 상위 집합에 "진행 중"이 없어서 69행 중 39행은 상위 값이 없다. 이 선택지를 고르면 envelope은 도메인 토큰을 항상 함께 실어야 한다.

두 variant 모두 `not_run`·`unknown`·`blocked`·`stale`·`uncertain`·`publish_unknown`·`inherited_uninspected`를 성공 값으로 올리지 않는다. 테스트 `status_crosswalk_is_complete_and_never_promotes`가 이를 검사하며, 일부러 금지 매핑을 넣었을 때 실패하는 것도 확인했다.

## 실행 소유권 4필드 (#2169, REQ-B2-01)

claim record(`src/claim.rs` `ClaimRecord`)과 handoff가 운반하는 동일 레코드는 소유권을 네 필드로 분리 저장한다. 어느 필드도 lock으로 해석하지 않는다 — 저널의 배타성은 `task_ref`와 `owner`만 읽는다(`acquire` 경로, REQ-B2-01 시험 `claim_record_four_ownership_fields_are_separate_and_not_locks`).

| 필드 | 뜻 | 설정자 | lock 여부 |
|---|---|---|---|
| `owner` (execution_owner의 별칭 — JSON 키는 `owner` 유지) | 실행 owner(`{kind, id, host_id?, pid?, process_start?}`) | `claim-acquire` 요청 | 저널 배타성의 유일한 판단 기준(단, lock은 아니며 advisory journal) |
| `assignee` | GitHub assignee — 표시 projection 전용 | `claim-acquire` 요청(기록만; claim은 GitHub에 절대 쓰지 않음 — REQ-06A-08) | 아니오 — 저널·게이트가 읽지 않음 |
| `provider_session` | 실행 세션의 host 식별자 | `claim-acquire` 요청 | 아니오 |
| `integration_owner` | 리더 통합 게이트 식별자(누가 이 claim의 통합 게이트를 소유하는지) | 통합 흐름(리더) — `claim-acquire`는 절대 설정하지 않음 | 아니오 |

세 가산 필드는 모두 serde-default(`Option`, 기록 없으면 직렬화에서 생략)이라 #2169 이전 레코드도 역직렬화가 불변이고, 스키마 시험 `claim_record_four_ownership_fields_are_separate_and_not_locks`가 분리 저장·비lock 해석·구형 레코드 호환을 함께 검사한다.

## CLI·MCP parity 정의 (01-C 실측용 체크리스트)

어느 envelope 선택지를 고르든 parity는 다음이 성립해야 한다. 실측은 01-C가 한다.

1. 같은 operation에서 MCP `structuredContent`와 CLI 결과(envelope을 쓰면 `result`)가 같다. "같다"의 기준(serde 직렬화 동일 또는 JCS 정규화 후 동일)은 HOLD-00D-5를 따른다.
2. MCP `outputSchema`는 이 디렉터리의 같은 schema 파일을 가리킨다(`$id`로 참조).
3. 실행 오류(CLI exit ≠ 0)는 MCP에서 `isError: true`로, 호출 형식 오류(CLI의 clap 사용 오류·malformed 입력)는 JSON-RPC protocol error로 나간다.
4. 오류 메시지는 두 경로 모두 입력 값을 echo하지 않는다.

## 검증 명령

```sh
export CARGO_HOME=$(test -d /tmp/moai-product-cargo/registry/cache && echo /tmp/moai-product-cargo || echo /tmp/moai-rc-product-cargo) CARGO_TARGET_DIR=/tmp/moai-target-<세션라벨>
cargo test --offline --locked protocol_fixtures      # 00D-AC-1
cargo test --offline --locked exit_code_table        # 00D-AC-2
for v in A B; do jq -e --arg v $v '.variants[$v] | ([.rows[]|select(.canonical==null and .no_map_reason==null)]|length)==0 and ([.rows[]|select((.source_token|IN("not_run","unknown","blocked","stale")) and (.canonical|IN("passed","applied","host_verified")))]|length)==0' schemas/protocol/v1/status-crosswalk.json; done   # 00D-AC-3
MOAI_UPDATE_FIXTURES=1 cargo test --offline --locked protocol_fixtures && git status --short tests/fixtures   # 00D-AC-7: 출력 없음 기대
```

schema 파일 자체의 형식(`$schema`, `$id`)은 `protocol_schemas_declare_2020_12_and_cover_current_fixtures`가 확인한다. 같은 테스트가 현행 성공 출력의 키가 schema와 맞는지도 보지만, 이는 키 대조일 뿐 schema 검증이 아니다. 검증기를 무엇으로 할지는 HOLD-00D-2라서 fixture 전체의 schema 검증은 실행하지 않았다(not_run).
