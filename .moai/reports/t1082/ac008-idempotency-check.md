# t1082 — AC-FLH-008 과 Send 멱등 처리 대조 (2026-09-23)

card: t1082 · SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 · 요청: 리드(t1100 plan-audit iter-2 ND2 연계)

## 판정

BOUND 전후에 같은 멱등 키로 보낸 dispatch 는 Send 가 duplicate 로 처리하지 않고 **거부한다.** AC-FLH-008 의 기대와 모순된다. 계약 조정이 필요하며, 결정은 리드가 t1082·t1100 두 레인 사이에서 한다. 이 레인은 코드를 고치지 않았다.

## 근거 (코드 읽기 — 실행 재현 아님)

- 스키마: `messages` 테이블 `UNIQUE(sender_session,idem_key)` (`internal/factorymsg/store.go:292`).
- `Send` (`store.go:587`) 의 삽입이 UNIQUE 로 실패하면(`:631`) 같은 `sender_session`·`idem_key` 의 기존 행을 읽는다(`:634`).
- 기존 행과 요청의 `kind`·`recipient_session`·`recipient_generation`·`task_ref`·`correlation_id`·`payload` 중 하나라도 다르면 `idempotency key collision with different request` 오류를 돌려준다(`:635-636`). 모두 같으면 기존 봉투를 돌려준다(`:638-640`).
- handoff 가 BOUND 되면 lane endpoint 의 session UUID 와 generation 이 바뀐다(REQ-FLH-008). 따라서 BOUND 뒤 같은 키로 새 endpoint 에 보내면 `recipient_session`/`recipient_generation` 이 달라 거부된다.
- AC-FLH-008(acceptance.md "Duplicate dispatch and same-lane redispatch")은 "identical idempotency keys before/after BOUND … one body execution and one accepted receipt exist, current duplicate returns duplicate disposition, and old generation returns stale NACK" 을 요구한다.
- 층 불일치: AC 의 `duplicate` 는 수신 측 receipt 처분 `DispositionDuplicate`(`store.go:34`, ACK 조건 `:763`)이고, 거부는 송신 측 `Send` 에서 먼저 일어난다.

좌표는 이 트리(HEAD 기준 develop `f0fdd88e4` 흡수 후)에서 다시 뽑은 값이다. t1100 이 인용한 `store.go:625` 는 같은 충돌 분기다.

## 갈림길

- (a) t1082 가 AC-FLH-008 을 좁힌다 — BOUND 뒤 재전송은 새 키를 쓰고, "같은 키 duplicate" 는 BOUND 전·같은 세대 안에서만 요구한다. 스키마 변경 없음.
- (b) t1100 이 멱등 판정 기준을 lane slot 으로 바꿔 recipient 가 바뀌어도 duplicate 로 처리한다 — 스키마·멱등 영역, t1100 소관.
- (c) Send 가 BOUND 재지정에 한해 recipient 불일치를 허용한다 — 멱등 의미 변경이라 t1100 경계에 걸린다.

## 영향과 공백

- AC-FLH-008 은 plan 의 M3 범위다. M1 진행에는 영향이 없고, M3 착수 전까지 결정이 필요하다.
- 실행 재현(같은 키, BOUND 전후 recipient 변경)은 하지 않았다 — M1 작업 중인 트리에 테스트 파일을 넣지 않기 위해서다.
