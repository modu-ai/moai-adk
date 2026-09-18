# 게이트웨이 추론 엔벨로프 복구 — 운영자 안내

> 카드 t708 · SPEC-GATEWAY-ENVELOPE-REPAIR-001 · REQ-EVR-010 (AC-EVR-012)

## 1. 이 문서가 다루는 형태 — CauseReasoning 400

GPT 계열 게이트웨이 대화에서 모델을 전환하면(예: gpt-5.6-sol → gpt-5.6-luna) 외부 클라이언트가
재인코드 과정에서 게이트웨이가 발행한 `redacted_thinking` 추론 엔벨로프 블록을 떨어뜨리는 경우가
있습니다. 이때 엔벨로프에 묶인 `toolu_moai_v1_*` 도구 마커는 살아 남기 때문에, 이후 모든 재생
요청이 게이트웨이 검증기에서 거부되며 응답 본문은 다음 고정 안내로 시작합니다.

```
conversation history changed, lacks reasoning, or belongs to another model family/account;
start a new conversation (reason: replayed history omits gateway-issued reasoning recorded
at this position; replay the history unmodified or start a new conversation)
```

`reason:` 절이 `replayed history omits gateway-issued reasoning` 인 400이 이 복구 경로의 대상입니다
(분류상 `CauseReasoning`). 이 형태는 카드 t703에서 실측으로 확정됐습니다.

## 2. 복구 절차 — 명시적 호출만

런처 세션 시작 시 `--repair-envelope` 플래그를 `--resume <UUID>` 와 함께 전달합니다.

```
moai cc --resume <UUID> --repair-envelope
```

- 복구는 **사용자가 명시적으로 호출할 때만** 실행됩니다. 400 분류만으로는 대화 기록, 페이로드,
  트랜스크립트 어디도 수정되지 않습니다(자동 수술 없음).
- `--repair-envelope` 는 `--resume <UUID>` 없이는 거부됩니다. `--continue` 와는 함께 쓸 수 없습니다.
- 복구가 성공하면 세션이 정상 재개됩니다. 거부되면 런처가 위 고정 안내 문구를 그대로 출력하고
  시작을 중단합니다.

## 3. 복구가 수행하는 일과 그 한계

복구기는 패밀리 네이티브 트랜스크립트(런처가 소유한 비공개 네임스페이스의 JSONL)에서 사라진
엔벨로프의 **원본 바이트를 그대로** 찾아, 마커가 살아 남은 어시스턴트 경계의 맨 앞에 다시 삽입합니다.

- **바이트 정확성**: 삽입되는 것은 게이트웨이가 실제 발행한 바이트뿐입니다. 각 경계마다
  `sha256(캐리어 원본 바이트) == 마커 내장 opaque_sha256` 이 자가 증명(self-attestation)으로
  확인되며, 어떤 엔벨로프도 새로 만들거나 재인코딩하지 않습니다.
- **전부 아니면 전무**: 하나의 경계라도 증명 불가능하면 전체 복구가 중단되고 아무것도 수정되지
  않습니다. 부분 복구는 없습니다.
- **비파괴 — 원본 보존**: 수정에 앞서 트랜스크립트 원본이 aside 파일로 그대로 보존되며 절대
  삭제되지 않습니다. 공개 콘텐츠(사용자 발화, 도구 결과, 텍스트, 도구 사용 블록)는 한 바이트도
  바뀌지 않습니다. 삽입은 thinking 블록뿐입니다.
- **단일 시도**: 복구 시도는 런처 대화 기록에 내구 레코드로 남고, 한 번 시도된 대화에서 다시
  호출되면 더 이상 주입하지 않습니다. 이 한도는 프로세스 재시작 후에도 유지됩니다. 거부는
  시도를 소모하지 않습니다 — 실제 주입 시도만 기록됩니다.

## 4. 거부 형태와 의미

| 거부 | 의미 | 조치 |
|---|---|---|
| `source-gone` (no verbatim source carrier) | 트랜스크립트에서 원본 엔벨로프 바이트를 더 찾을 수 없습니다 — 클라이언트 압축, 슬라이싱, 기록 손실 | 복구 불가. 새 대화를 시작하거나 아래 포크 경로로 |
| `digest` 불일치 | 트랜스크립트에 남은 캐리어의 다이제스트가 마커가 자가 증명하는 값과 다릅니다 — 기록이 변형됐거나 다른 엔벨로프입니다 | 복구 불가. 원본을 신뢰할 수 없는 상태입니다 |
| `already-attempted` | 이 대화에서 이미 한 번 복구를 시도했습니다 (내구 레코드 존재) | 단일 시도 한도. 재시도 불가 — 고정 안내대로 새 대화 권장 |
| `aside` 충돌 잔재 (aside 쓰기와 기록 쓰기 사이 크래시) | aside 파일은 남았는데 시도 기록이 없어 다음 복구가 영구 거부됩니다 (부분 시도 안전 규칙) | 수동 복구: 트랜스크립트가 원본 그대로임을 확인한 뒤 `<트랜스크립트>.moai-repair-aside` 파일을 삭제하면 재시도할 수 있습니다 |
| 정상 이력 (재원 없음, no stripped boundary) | 떨어진 엔벨로프가 없습니다 — 복구할 대상이 없습니다 | 아무 수정도 없습니다. 문제가 다른 축에 있는 것입니다 |

거부 시에도 aside나 내구 레코드는 만들어지지 않고, 대화 상태는 호출 전과 바이트 단위로 동일합니다.

## 5. 포렌식 위치

- **원본 보존 (aside)**: 복구된 트랜스크립트와 같은 디렉터리의 `<트랜스크립트 경로>.moai-repair-aside`
  — 복구 직전 원본 그 자체입니다.
- **시도 기록 (durable record)**: `<패밀리 루트>/families/<FamilyID>/repair/<UUID>.json` —
  `attempted` 플래그, 경계별 `{digest, position}` provenance, aside 파일명과 aside 원본의
  `aside_sha256` 을 담습니다. 이 기록이 단일 시도 한도의 근거입니다.

## 6. 복구 불가능한 형태와 정상 경로

- **라인지 불일치**(다른 루트에서 재생, `no recorded history exists for this session`):
  복구 대상이 아닙니다. 정상 경로는 런처가 제공하는 포크입니다 — `--resume <부모UUID> --fork-session`
  으로 자식 루트를 시드한 뒤 이어갑니다.
- **꼬리 미발행 경계 웨지**(클라이언트가 기록만 있고 게이트웨이가 발행하지 않은 어시스턴트 턴을
  들고 오는 형태): 이 카드의 대상이 아니며, 형제 카드의 재뿌리기 정책이 소관입니다 —
  `SPEC-GATEWAY-WEDGE-REROOT-001` (t700).
- **직렬 턴 chain-class 400**(모델 전환 없는 새 세션에서의 `CauseChain`): 역시 복구 대상이
  아니며, 카드 t707이 원인 규명을 소관으로 합니다.

## 7. 참조

- SPEC: `.moai/specs/SPEC-GATEWAY-ENVELOPE-REPAIR-001/`
- 보안 근거(원본 재주입이 검증기를 우회하지 않는 이유): 위 SPEC spec.md §3
- 구현: `internal/gateway/conversation/repair.go` (복구기), `internal/cli/gateway_repair.go` (런처 플래그)
- 검증: 복구된 재생은 요청 시점에 변경되지 않은 게이트웨이 Check가 심판합니다 — 복구 경로 자체는
  수용 집합을 넓히지 않습니다.
