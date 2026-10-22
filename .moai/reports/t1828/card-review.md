# t1828 card-review (codex_review scope=card)

- round 1: 2026-10-17 12:5xZ · backend codex · base `93f4fc75c` · tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1828` · verdict **fail**(advisory) — P2 1건
- round 1 수리: 커밋 `34212a4be` — RED 양방향(재등록 세대 오거부·빈 하트비트 오거부) 재현 후 세대 판별(registered_at ≤ HeartbeatAt + pid 생존) GREEN. 레인 독립 재실행: TestFactoryNextLeasesBehindExpiredLeaseOfReRegisteredHolder PASS 1.55s(ok 2.535s).
  - RED verbatim(수리 전, `go test ./internal/cli/ -run TestFactoryNextLeasesBehindExpiredLeaseOfReRegisteredHolder -count=1`): `--- FAIL: TestFactoryNextLeasesBehindExpiredLeaseOfReRegisteredHolder (1.64s)` / `factory_serial_slot_liveness_test.go:165: factory next refused behind an expired lease whose holder is dead: factory next: refused serial-slot: another serial card is in flight and holds the serial slot: t1 (holder lane-2)` / `FAIL github.com/modu-ai/moai-adk/internal/cli 2.587s`
  - RED verbatim(빈 하트비트 하위시험, `-run TestFactoryNextLeasesBehindExpiredLeaseWithDeadHolder`): `--- FAIL: TestFactoryNextLeasesBehindExpiredLeaseWithDeadHolder/empty_heartbeat (1.36s)` / 동일 거부 문구 / `FAIL github.com/modu-ai/moai-adk/internal/cli 6.179s`
  - GREEN: TestAutoLaneCycle `ok 56.365s` · TestFactorySerialSlotHeld `ok 1.085s` · 신규 4종 `ok 2.5-10.9s` (sync-audit F2 보강, 10-18)
- round 2: 2026-10-17 13:0xZ · backend codex · 동일 base/scope/card diff(424605284+34212a4be) · verdict **pass** — 결함 0건("변경 코드 정적 검토에서 수정이 필요한 결함을 발견하지 못했어")

## Round 1 — P2: 재사용된 레인 이름을 이전 임대 보유자로 오인 (factory_card.go:326)

`factoryWorkerPidAlive`가 `SELECT pid FROM workers WHERE label=?`로 OwnerLabel 의 pid 를 읽는다. `factory_slots.go`는 죽은 worker 행을 삭제하고 같은 label 을 새 프로세스에 재할당하므로, **레인 재합류(재부팅 후 같은 lane-N 라벨)** 시 새 프로세스의 생존이 옛 카드 행의 생존으로 읽혀, 죽은 보유자의 만료 임대가 다른 serial 카드의 임대를 계속 차단한다. pid 재사용 없이도 발생. 보고서 Residual-risk 가 pid 재사용만 적은 것은 불충분했다 — label 재할당은 상례다.

처방: 같은 프로세스 세대 판별을 기존 컬럼으로 — workers.registered_at ≤ card.HeartbeatAt(하트비트를 쓴 세대가 등록된 이후인 경우만) + pid 생존. 신규 등록이 하트비트보다 뒤면(새 세대) 해제. 스키마 변경 없음.

정적 검토 지적(테스트 미실행)이며 lane 이 메커니즘 검증 후 수리·재심사한다.
