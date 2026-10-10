# SPEC-REVIEW-GATE-REPAIR-001 — 설계: 상태 모델과 판정 표

이 문서는 요구사항을 실현하는 상태 모델을 적는다. 함수 이름과 자료 구조의 세부는 run 단계가 정한다. 여기서는 어떤 상태에서 무엇이 일어나야 하는지, 그리고 운영자 판정이 정한 규칙을 고정한다.

## 1. 리뷰 키와 영수증 상태

리뷰 키는 영수증의 결합 필드 다섯 개의 조합이다: head, tree_digest, config_digest, command, tool_version(`internal/verify/receipt.go:70-76`). 키가 하나라도 다르면 영수증은 다른 키의 것이다.

영수증 상태는 여섯 가지다.

| 상태 | 정의 |
|---|---|
| absent | 같은 키의 영수증이 없다 |
| stale | 같은 키의 영수증이 있으나 유효기간(24시간)이 지났다 |
| fresh-pass | 같은 키, 유효기간 안, 판정 pass |
| fresh-fail | 같은 키, 유효기간 안, 판정 fail |
| fresh-inconclusive | 같은 키, 유효기간 안, 판정 inconclusive |
| key-mismatch | 결합 필드 중 하나 이상이 다르다 |

## 2. 상태별 행동: 현재와 목표

목표 열의 "검토 시작"은 살아 있는 진행 중 표식을 존중한다. 표식은 검토가 어떤 판정(pass, fail, inconclusive)으로 끝나든 완료 시점에 풀리며, 소유자 pid 가 살아 있지 않은 고아 표식은 새 판정 전에 회수된다. 같은 커밋의 재검토는 표식이 아니라 결과 캐시가 정한다(§3). 간격은 없다(Q4 = (a)). 이 규칙은 목표이며, 현행 코드는 아직 구현하지 않는다(§3, 현행 코드와의 차이).

| 상태 | 진입(프롬프트) 현재 | 진입 목표 | 턴 종료 현재 | 턴 종료 목표 |
|---|---|---|---|---|
| absent | 허용 | 허용 | 검토 시작 후 허용 | 검토 시작 후 허용, 메시지 (REQ-RG-011, 012) |
| stale | 허용 | 허용 | 검토 시작 후 허용 | 검토 시작 후 허용, 메시지 (REQ-RG-011, 012) |
| key-mismatch | 허용 | 허용 | 검토 시작 후 허용 | 검토 시작 후 허용, 메시지 (REQ-RG-011, 012) |
| fresh-fail, 전달 기록 없음 | 차단 | 허용, 상세를 추가 맥락으로 한 번 전달, 전달 기록 (REQ-RG-002) | 차단 (stop_hook_active 가 거짓일 때) | 차단 횟수가 C 미만이면 차단(횟수 +1), 아니면 허용 + 메시지 (REQ-RG-003, 004, 012) |
| fresh-fail, 전달 기록 있음 | 차단 | 허용, 반복 없음 (REQ-RG-002) | 차단 (stop_hook_active 가 거짓일 때) | 차단 횟수가 C 미만이면 차단, 아니면 허용 + 메시지 (REQ-RG-003, 004, 012) |
| fresh-pass | 허용 | 허용 | 허용 | 허용, 메시지 없음 (REQ-RG-009) |
| fresh-inconclusive | 허용 | 허용 | 허용(재사용) | 재검토 시작(살아 있는 표식이 없을 때), 허용 + 메시지 (REQ-RG-005, 006, 012) |
| 스위치 꺼짐 | 검토 없음 | 검토 없음 | 검토 없음 | 허용 + 메시지 (REQ-RG-007, 010, 012) |
| 검토할 변경 없음 | 검토 없음 | 검토 없음 | 검토 없음 | 허용 + 메시지 (REQ-RG-010, 012) |
| stop-hook 해제 | — | — | 허용 | 허용 + 메시지 (REQ-RG-010, 012) |
| reviewer 없음 | 허용 | 허용 | 허용 | 허용 + 메시지 (REQ-RG-010, 012) |

## 3. 데이터 모델: 전달 기록과 차단 횟수

- **전달 기록**은 리뷰 키 단위로 하나 있다. 내용은 키와 전달 시각이다. 다른 키에는 적용되지 않는다. 키가 바뀌면 새 키의 기록이 새로 생긴다(REQ-RG-002).
- **차단 횟수**도 리뷰 키 단위다. 턴 종료가 그 키를 차단할 때마다 1 을 더한다. 허용은 값을 바꾸지 않는다(보드 판정 D3). 그 키에 fresh pass 영수증이 기록되면 기록을 제거한다. 키가 바뀌면 새 키는 0 에서 시작한다. C = 1(운영자 판정 Q2)(REQ-RG-003, 004).
- 두 기록은 영수증이 저장되는 상태 루트(`.moai/state/verify/` 아래)에 둔다. 같은 구조가 레인 저장소의 `codex-review/` 와 `snapshots/` 에서 관측되었다(research.md §4, 기계 로컬). 위치의 세부는 run 단계가 정한다. F1 이 영수증 결합 기준을 바꾸면 두 기록의 키도 함께 바뀔 수 있다. 이 SPEC 은 그 변경을 하지 않는다.
- **진행 중 표식의 수명 (규칙).** 표식은 검토가 시작될 때 만들어지며 소유자 pid 를 싣는다. 소유자는 `moai verify codex-review` 를 분리 실행한 검토 프로세스다(`internal/cli/codex_review_delay.go:52`). 턴 종료 훅은 검토를 시작한 뒤 곧 허용하고 끝나므로(`internal/cli/codex_review_gate.go:202–210`), 훅 프로세스의 pid 는 소유자가 될 수 없다. (가) 해제: 표식은 검토가 pass, fail, inconclusive 중 **어떤 판정으로 끝나든** 완료 시점에 풀린다. 완료 시점은 검토 결과 영수증을 기록하는 단계다(`internal/cli/codex_review_receipt.go:198`, 호출 `:245`). (나) 시작 실패: 검토 시작이 실패하면 풀린다(현행 동작, `internal/cli/codex_review_gate.go:203–205`). (다) 회수: 소유자 pid 가 살아 있지 않은 고아 표식은 새 판정을 내리기 전에 회수된다. 소유자 pid 를 읽을 수 없는 표식도 고아로 본다. 여기에는 표식을 만든 뒤 pid 를 쓰기 전의 짧은 창(`internal/cli/codex_review_delay.go:104–111`, 프로세스 시작은 `:62`)과 이전 형식의 시각 한 줄 표식이 포함된다. 판정이 불확실하면 회수하는 쪽을 택한다(fail-open). 표식이 검토를 영구히 막는 것을 막기 위해서다. 이 선택의 비용은 드문 중복 검토 한 번이다. (라) 표식은 **동시 실행만** 막는다. 같은 커밋의 재검토를 막는 것은 아래 결과 캐시다.
- **결과 캐시 (규칙).** 결과 캐시는 검증 영수증 저장소(`internal/verify/receipt.go`)다. 기록은 `RecordReceipt`(`:100`), 읽기는 `LoadReceipt`(`:117`), 재사용 판정은 `CheckReceipt`(`:63`)다. 같은 키의 완료된 pass 또는 fail 은 재사용되어 그 커밋의 재검토를 막는다(REQ-RG-005). inconclusive 는 pass 가 아니므로 재사용되지 않으며 재검토를 막지 않는다(REQ-RG-006).
- **현행 코드와의 차이.** 현행 표식 경로는 검토 예산 단위의 시간 버킷을 이름에 싣고(`internal/cli/codex_review_delay.go:76–80`), 내용은 시각 한 줄이다(`:111`). 완료 시 표식을 푸는 코드는 없다. 제거는 이전 버킷의 표식(`:98–100`)과 검토 시작 실패(`internal/cli/codex_review_gate.go:204`) 두 곳뿐이다. 현행 재사용 지점(`internal/cli/codex_review_gate.go:159–178`)은 수용된 영수증이 pass 든 inconclusive 든 허용하므로 REQ-RG-006 과 다르다. 위 두 규칙은 목표이며 현행 코드는 구현하지 않는다. 구현은 M5 가 맡는다.

## 4. 기준 해석 함수 — t1662 로 이관되었다(REQ-ACE-005, 리더 판정 d-20261010T070728Z-6cdf).

턴 종료 게이트와 감사 도구가 리뷰 기준 ref 를 고르는 방식을 하나의 해석 함수로 모으는 일은 후속 카드 t1662 가 맡는다(REQ-ACE-005). 이 SPEC 은 그 함수를 구현하지 않으며, 두 해석기의 기준을 맞추는 변경도 하지 않는다. 이 절은 그 이관을 기록하기 위해 남긴 자리이며, 함수의 설계와 후보 집합은 이 문서에 없다.

## 5. 실행 단위 스위치 (M6)

스위치는 run 단위에서 값이 하나다. 운영자 판정 Q3 = (b) 주 체크아웃의 미추적 재정의 파일, 보드 판정 D11 = 커밋된 설정값을 따름.

- **재정의 파일**: 주 체크아웃의 `.moai/state/codex-review-gate/run-switch.json`. `.moai/state/` 는 `.gitignore` 401행으로 무시되는 트리다(미추적). 파일 형식과 필드는 아래 '재정의 파일 스키마' 항목이 정한다. 파일은 실행 식별자와 활성화 값(on/off)을 싣는다.
- **재정의 파일 스키마**: JSON 객체 하나다. 필드는 다음 둘이며, 그 밖의 필드는 무시한다.
  - `run_id` — JSON 문자열. 단일 활성 run 의 `run_id` 와 바이트 단위로 같아야 한다(대소문자 구분, 앞뒤 공백 없음).
  - `codex_review_gate` — JSON 문자열. `"on"` 또는 `"off"` 만 허용한다.
  - 예: `{"run_id": "<활성 run 의 run_id>", "codex_review_gate": "off"}`
- **run 의 정의**: factory run 하나. `--run` 플래그가 가리키는 factory run id 로 지칭되며, 활성 run 은 하나다(`internal/cli/factory_card.go` 의 `--run` 도움말).
- **주 체크아웃 해석**: `IdentifyPrimaryCheckout`(`internal/auditreceipt/storeroot.go:142`, 주 체크아웃을 돌려준다)을 쓴다. 목표: 연결된 워크트리에서 읽는 모든 레인이 같은 경로에 도달한다(AC-RG-012, AC-RG-013 이 확인한다). 주 체크아웃을 식별할 수 없으면(연결되지 않은 복제본) 재정의 파일을 찾지 않는다. `StoreRoot` 는 트리에 workflow.yaml 이 있으면 그 트리를 돌려주므로 이 용도로 쓰지 않는다.
- **적용 조건**: 재정의 파일이 싣는 실행 식별자가 단일 활성 run 과 같을 때만 적용한다. 다른 식별자이거나 단일 활성 run 을 해석할 수 없으면 파일을 무시하고 커밋된 값을 따른다. 세션이 그 run 의 member lane 이 아니면(아래 'member lane 판정') 파일을 무시하고 커밋된 값을 따른다.
- **member lane 판정 (run 소속).** 카드 워크트리와 run 의 결합은 `cards` 행이 이미 기록한다. `cards` 는 `run_id` 와 `card_id` 로 행을 키 짓고(`internal/homestate/factory.go:77`, `PRIMARY KEY(run_id, card_id)`) `worktree_path` 열을 가진다(`internal/homestate/factory.go:66`). 카드 임대의 기록 단계 `factoryEnsureCardWorktree` 가 `RecordCardWorktree(ctx, runID, cardID, path, actor, now)` 로 그 열에 트리 경로를 쓴다(`internal/cli/factory_card.go:598`; 구현 `internal/homestate/card_worktree.go:20–62`). run 을 바꿀 때 이전 run 의 트리 경로는 `factoryCarryPreviousWorktree`(`internal/cli/factory_card.go:489–505`)가 찾아 새 run 의 행에 다시 기록한다(`internal/cli/factory_card.go:984–991`). 판정 규칙: 세션 작업 트리의 루트(세션 작업 디렉터리에서 `git rev-parse --show-toplevel` 이 돌려주는 경로)를 심볼릭 링크를 풀어 정규화한 절대 경로가, 해석한 단일 활성 run 의 `cards` 행 가운데 하나의 `worktree_path` 와 같으면 그 세션은 그 run 의 member lane 이다. 주 체크아웃, 다른 run 의 트리, 어느 행에도 없는 트리는 member 가 아니다. 경로를 정규화하지 못하는 세션 트리나 행도 같은 것으로 보지 않는다(member 아님, 커밋된 값을 따른다). 현재 코드에는 경로에서 행으로 가는 조회가 없다. `resolveFactoryCardRun`(`internal/cli/factory_card.go:2313–2322`)은 run 만 해석하고 membership 을 묻지 않는다. 가장 가까운 `PreviousCardWorktree`(`internal/homestate/card_worktree.go:73–90`)는 카드 id 로 찾는다. 경로에서 run 으로 가는 조회(`run_id` 와 `worktree_path` 로 읽기 한 번)는 run 단계(M6)가 추가한다. 결합 자체는 새로 만들지 않는다.
- **단일 활성 run 규칙 (코드가 적용하는 규칙)**: 기본값을 적용하는 함수는 `resolveFactoryCardRun`(`internal/cli/factory_card.go:2313–2322`)이며, 그 본체는 `factorymsg.ResolveActiveRun(ctx, root, "")`(`internal/factorymsg/store.go:343–395`)다. 명시 run 인자 없이 부르므로 규칙은 다음과 같다. (1) `runs` 테이블에서 `status='active'` 인 `run_id` 를 `run_id` 순으로 모은다(`store.go:359`). (2) 0개이면 해석하지 못한다(`NO_ACTIVE_FACTORY`, `store.go:373–374`). (3) 1개이면 그 run 이 단일 활성 run 이다(`store.go:375–376`). (4) 2개 이상이면 `ReconcileActiveRuns` 가 소유자가 죽은 것으로 입증된 run 만 활성 집합에서 뺀 뒤 다시 센다(`store.go:383`, `internal/homestate/factory_run_retire.go:332–356`). 정확히 1개가 남으면 그 run 이 단일 활성 run 이고, 0개나 2개 이상이면 해석하지 못한다(`store.go:387–394`, `AMBIGUOUS_FACTORY`). (5) 재조정은 생존자 가운데서 고르지 않는다(`store.go:378–381`). 따라서 해석하지 못하는 결과는 항상 닫힌 쪽이다. 연결된 워크트리의 루트는 주 체크아웃으로 정규화되므로(`internal/homestate/paths.go:128–143`) 어느 트리에서 불러도 같은 저장소의 run 집합을 읽는다.
- **재정의 파일의 판정 순서**: 순서는 결과를 바꾸지 않고 부작용만 줄인다. (가) 파일을 읽는다. 파일이 없거나, 읽기에 실패하거나, 스키마에 맞지 않으면(JSON 객체가 아님, 필드 누락, 타입 불일치, `codex_review_gate` 가 `on`·`off` 가 아님) 파일이 없는 것과 같다. (나) 파일이 맞을 때만 단일 활성 run 을 해석한다. 해석하지 못하면 파일이 없는 것과 같다. (다) `run_id` 가 해석한 run 과 같을 때만 `codex_review_gate` 값이 적용된다. 같지 않으면 파일이 없는 것과 같다. (라) 세션이 (나)에서 해석한 run 의 member lane 이 아니면(아래 'member lane 판정') 파일이 없는 것과 같다. 네 경우 모두 커밋된 값을 따른다(REQ-RG-007).
- **관측된 부작용**: 단일 활성 run 해석은 읽기 전용이 아니다. `homestate.OpenFactory` 는 디렉터리를 만들고 권한을 0700 으로 맞추며, DDL 을 실행하고 WAL 을 켠다(`internal/homestate/factory.go:167–176`, `:268–293`). 2개 이상일 때의 재조정은 소유자가 죽은 run 을 `retired` 로 바꾼다(`internal/homestate/factory_run_retire.go:342–345`). 판정 순서 (나) 는 이 부작용을 유효한 파일이 있을 때로 줄인다. 이 해석은 현행 동작이며 이 SPEC 은 바꾸지 않는다. 읽기 전용으로 세기만 하면 죽은 run 이 남아 활성 run 집계가 달라지므로, 읽기 전용 해석으로의 전환은 별도의 후속 판정으로 남긴다.
- **커밋된 값**: 판독 트리의 HEAD 커밋에 있는 `.moai/config/sections/workflow.yaml` 의 `codex.review_gate.enabled`. 작업 트리의 미커밋 값은 쓰지 않는다. 판독은 `git show HEAD:` 같은 읽기 전용 조회다. 판독에 실패하면(파일이 없거나 읽을 수 없거나, 키가 없거나 값이 불리언이 아니면) 검토를 시작하지 않고 허용하며, 메시지 상태 `커밋 값 읽기 실패`를 낸다(REQ-RG-012, fail-open).
- **우선순위**: 적용되는 재정의 값 > 커밋된 값.
- **추적 키의 역할**: 기본값의 원천으로만 남는다. 실행을 전역으로 끄거나 켜는 권한은 없다.
- **두 층**: 셸 래퍼와 Go 처리기가 같은 해석 절차를 쓴다(REQ-RG-008).

## 6. 판정 표

| 조건 | 결과 | 기록 |
|---|---|---|
| 키 일치, 기준 해석 성공, fresh pass | 허용, 재검토 없음 | 기준 ref 이름, merge-base SHA |
| fresh fail, 차단 횟수 < C | 차단, 차단 횟수 +1 | 차단 횟수 |
| fresh fail, 차단 횟수 = C | 허용 + 메시지 (REQ-RG-004, 012) | 차단 횟수 값 |
| inconclusive | 재검토 시작(살아 있는 표식이 없을 때, 간격 없음), 허용 + 메시지 | 시작 시각 |
| absent, stale, key-mismatch | 재검토 시작(살아 있는 표식이 없을 때), 허용 + 메시지 (REQ-RG-011) | 상태 이름 |
| 스위치 꺼짐 | 검토 없음, 허용 + 메시지 (REQ-RG-007, 012) | 스위치 출처 (재정의 파일 또는 커밋 값) |
| 검토할 변경 없음 | 허용 + 메시지 (REQ-RG-010, 012) | 사유 |
| stop-hook 해제 | 허용 + 메시지 (REQ-RG-010, 012) | 사유 |
| 커밋 값 읽기 실패 | 검토 없음, 허용 + 메시지 (REQ-RG-007, 012) | 상태 이름 |

## 7. 관측 가능성

- 메시지 상태 어휘: `absent`, `stale`, `key-mismatch`, `inconclusive`, `차단 횟수 소진`, `스위치 꺼짐`(출처 포함), `커밋 값 읽기 실패`(committed value read failed), `검토할 변경 없음`, `stop-hook 해제`, `reviewer 없음`.
- 메시지는 Stop 훅의 `systemMessage` 로 낸다(`.claude/rules/moai/core/hooks-system.md:103`). 차단 사유 전반의 가시성 확장은 F2 가 맡는다.
- 스위치 출처(재정의 파일인지 커밋된 값인지)를 검토가 꺼진 이유로 함께 기록한다.
