# Ledger reservations (`src/ledger/reservation.rs`)

Issue #3068 P3 · 라이브러리 API 뿐이며 명령은 없다(`moai card` 는 #3070). 스키마는 P1 그대로이고 새 표·열이 없다. 시각은 모두 DB 시계(`unixepoch()`)이고 호출자가 시각을 넘기지 않는다.

## API

모든 쓰기는 `WriteLedger::with_write_tx` 한 번(`BEGIN IMMEDIATE`) 안에서 일어난다.

| 함수 | 하는 일 |
|---|---|
| `reserve(w, card_id, holder, write_scope: &[String], lease_secs, worktree: Option<&str>, branch: Option<&str>) -> Result<Reservation, ReserveError>` | `ready` 카드를 잡는다: 다음 세대(`card.generation + 1`), `held` 행과 쓰기 범위 행, 카드 `ready→active`, `reserve` 사건 |
| `renew(w, card_id, holder, generation, lease_secs) -> Result<Reservation, FenceError>` | 현재 세대의 `held` 예약만. 새 `lease_until` = max(지금 lease 끝, DB 시계 + lease_secs). 살아 있는 lease 필요 |
| `release(w, card_id, holder, generation) -> Result<Released, FenceError>` | 현재 세대만 `released`. 카드가 `active` 면 `ready` 로, 이미 다른 상태(`review` 등)면 그대로. lease 가 끝났어도 가능 |
| `handoff(w, card_id, holder, generation, to_holder) -> Result<Reservation, FenceError>` | 옛 행 `released` + `to_holder` 의 새 `held` 행(세대 + 1, 같은 범위·worktree·branch·lease 끝). 카드는 `active` 그대로. 살아 있는 lease 필요 |
| `fence(w, card_id, holder, generation) -> Result<(), FenceError>` | 푸시·쓰기 전 확인. 성공하면 아무것도 쓰지 않는다 |
| `fence_check(r: &ReadLedger, ...) -> Result<(), FenceError>` | 같은 판정을 읽기 전용으로. 기록하지 않고 파일을 만들지 않는다 |
| `current(conn, card_id)`, `history(conn, card_id)` | 읽기 전용 조회(`ReadLedger::conn()` 도 가능) |

범위 겹침은 같은 트랜잭션에서 다른 카드의 활성 예약(`held`·`expired_candidate`·`held_for_review`)과 쌍마다 판정한다. 판정은 둘이다: 겹침 술어 `crate::scope::overlap`(대소문자를 접고 `.`·`..` 를 풀어 "둘 다 쓸 수 있는 경로가 있나"를 본다)과, 그 위에서 카드 쪽에만 쓰는 생성 경로 규칙 `crate::scope::card_overlap`(한쪽이 `GENERATED_SHARED_PATHS` 의 경로와 대소문자까지 정확히 같으면 그 쌍은 겹침이 아니다). 예약·넓히기(`card scope`)·`card next` 는 `find_overlap` 을 거쳐 `card_overlap` 을 부르고, 클레임 쪽 범위 검사(`scope::overlaps`)는 `overlap` 만 쓴다. `card check` 는 겹침이 아니라 경로가 예약 범위 안인지만 본다. 비어 있는 범위는 겹침을 보일 수 없어 거절한다.

## 한도 (N 받고 N+1 거절)

| 값 | 한도 |
|---|---|
| `lease_secs` | 60 – 86400 |
| 범위 패턴 개수 | 1 – 64 (중복을 합치기 전 목록 길이) |
| 패턴 길이 | 1 – 512 글자, 제어 문자 없음 |
| `holder`, `to_holder` | 1 – 200 글자, 제어 문자·앞뒤 공백 없음, 둘이 같으면 거절 |
| `worktree`, `branch` | 1 – 512 글자 |

## 오류 코드

`ReserveError`: `invalid_input` · `empty_scope` · `card_not_found` · `card_not_ready` · `already_held` · `scope_overlap` · (원장 오류는 `LedgerError::code`).
`FenceError`: `invalid_input` · `card_not_found` · `no_reservation` · `stale_generation` · `unknown_generation` · `not_held` · `wrong_holder` · `lease_expired` · (원장 오류).

## 거절된 호출이 남기는 것

| 거절 | 남는 것 |
|---|---|
| 입력 한도, 빈 범위, `card_not_found` | 아무것도 없다(행·사건 모두 그대로) |
| `reserve` 의 `card_not_ready`, `already_held`, `scope_overlap` | 아무것도 없다 |
| `renew`·`release`·`handoff`·`fence` 의 `no_reservation`, `stale_generation`, `unknown_generation`, `not_held`, `wrong_holder`, `lease_expired` | `event` 한 줄(`kind='fence_reject'`, `actor` = 요청한 holder, `payload` = `op`, `card_id`, `holder`, `offered_generation`, `current_generation`, `reason`). 예약·카드 행은 그대로이고, 그 사건은 거절을 찾은 같은 트랜잭션에서 커밋된다 |
| 쓰는 도중 프로세스가 죽음 | 롤백된다. 세대는 다시 같은 번호로 발급된다 |
