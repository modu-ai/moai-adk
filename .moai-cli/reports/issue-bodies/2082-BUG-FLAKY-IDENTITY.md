## 관측
rc-ci run 42481918815(PR #2075 head 2fcdde18848e37f1b2908e475a0b786b83761441, ubuntu)에서 `product_cli` 52 passed, 1 failed:

```
thread 'identity_primary_checkout' panicked at tests/product_cli.rs:1296:46:
called `Result::unwrap()` on an `Err` value: Os { code: 2, kind: NotFound, message: "No such file or directory" }
```

같은 PR의 직전 head 6bcf809ad(변경 줄 동일, 이 시험·src/identity.rs 변경 없음)는 rc-ci 42481665300 success였다. PR #2075의 diff는 src/apply.rs recover·시험·docs뿐이다.

## 가설(미검증)
`tests/product_cli.rs` `tree_state`의 `walk`가 `read_dir`로 본 항목을 `symlink_metadata`(1296행)로 읽기 전에 그 파일이 사라졌다. 후보: 시험 안 git 호출이 남긴 임시 파일(lock 등) 또는 백그라운드 git 작업. 원인 확정 전에는 가설이다.

## 범위
- RED 먼저: 같은 시험을 반복 실행해(예: 반복 루프, Linux 컨테이너 가능하면) 재현을 시도하고 명령·횟수·결과를 기록한다. 재현이 안 되면 결함 아님 후보로 측정과 함께 보고하고 정지한다.
- 재현되면 원인을 증거로 확정한 뒤 시험 헬퍼(또는 실제 원인 쪽)를 최소 수정한다. 시험이 확인하려는 "쓰기 없음" 불변식은 약화하지 않는다.

## 수락 기준
- AC-1: 재현 명령과 결과(재현/비재현)를 기록.
- AC-2 (재분류, 아래 기록): regression-guard — RED 재현 불가. 수정 커밋 단독 + check set 전부 exit 0 + 수정 뒤 rc-ci(ubuntu) success 누적으로 판정한다. (원문: "수정 시 RED 단독 커밋이 수정 커밋보다 먼저, check set 전부 exit 0.")
- AC-3: 독립 AUDIT PASS.

## AC 재분류 기록 (RED 재현 불가)
- 원 AC-2: "수정 시 RED 단독 커밋이 수정 커밋보다 먼저, check set 전부 exit 0." — 본문 sha256 기준판 2026-11-08T13:28:10Z.
- 재분류: regression-guard. 로컬에서 실패 시험을 만들 수 없다 — macOS·Linux 로컬 반복 ~188회 0재현(레인 agent-5), 감사 레인 macOS 커밋 40회 .git 변화 0회.
- 대체 근거(결함 실재): rc-ci run 42481918815(product_cli.rs:1296 NotFound), 42482672162(claim_cli \"files vanished: .git/objects/maintenance.lock\"), 42482267478 attempt 1(product_cli.rs:2508), 42484616623 attempt 1(product_cli.rs:2508).
- 효과 판정: 이 수정 병합 뒤 rc-ci에서 같은 부류 실패가 다시 나오면 이 Issue를 다시 연다.
- 재승인: 리더(#2082 리더 판단 댓글, 사용자 결정 \"HOLD만 해제\" #1975 issuecomment-6721834190 아래), 감사 근거 agent-2 #2088 판정.

리더 조치: 같은 run의 실패 job은 `gh run rerun 42481918815 --failed`로 재실행(결과는 PR #2075 게이트에서 판정).
