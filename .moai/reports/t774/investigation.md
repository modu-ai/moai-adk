# t774 plan-phase 조사 기록 — slot 표면 실측 (2026-09-14, lane-6)

카드 판정 3항목과 [HARD] 통제군 obligation에 대한 이번 실행의 측정 기록. 모든 명령은 워크트리 `.claude/worktrees/t774`(기점 d416f8162)와 1차 체크아웃 로그에서 실행했다.

## 1. 표면 존재와 연대

- 구현: `internal/cli/slot.go`(acquire/status/release). 임대 레코드는 1차 체크아웃 `.moai/state/slot-leases`, 모든 링크드 워크트리 공유. 살아 있는 보유자 안에서 제2 획득은 거부된다(slot.go:7-11, 91-96).
- 감사 로그(`.moai/logs/slot-lease-audit.jsonl`) 최초 행: `{"ts":"2026-09-12T12:28:42Z","event":"acquire","resource":"heavy-test","session_id":"da011b34-…"}`. plan 시점 총 19행 — resource 분포 `heavy-test`×13, `t675-cli-tests`×4, `gotest-gateway`×2. 카드 발행(09-10)보다 2일 뒤 — 즉 침입 당시엔 표면이 없었고 지금은 가동 중이다(plan-audit 시점 25행으로 증가 — 레인들이 계속 쓰고 있다).
- 절차 문면: `.claude/rules/local/gitflow-lane-protocol.md` §8(100행)에 `moai slot acquire --resource <이름> --max-duration <상한>` → 실행 → `release`, `status` 조회, integration 분리 근거, --max-duration 만료, 선택형 가드(`workflow.slot_lease.enabled`, 기본 꺼짐)까지 **이미 존재**(e78fd0ee6, 2026-09-12, card t607 / SPEC-RESOURCE-SLOT-LEASE-001). `.claude/rules/moai/workflow/resource-slot-lease.md`는 템플릿 미러 본과 동일 존재(배포 사용자 표면).

## 2. 강제성 생시 재현 (카드 범위 리소스명으로)

```
$ moai slot acquire --resource t774-repro-demo --name "t774 plan-phase enforcement demo" \
    --command "go test ./internal/cli/"
slot t774-repro-demo acquired by 409b4ba3-… until 2026-09-13T17:37:00Z

$ moai slot acquire --resource t774-repro-demo --session t774-demo-b --name "intruder lane"
(출력 없음)
진짜 exit=3            # 파이프 없이 재측정 — 거부, 침묵, 대체 없음

$ moai slot status --resource t774-repro-demo
  session: 409b4ba3-… (pid 54135)
  name:    t774 plan-phase enforcement demo
  command: go test ./internal/cli/
  since:   2026-09-13T17:07:00Z
  bound:   30m0s, ends 2026-09-13T17:37:00Z

$ moai slot release --resource t774-repro-demo
slot t774-repro-demo released (was 409b4ba3-…)
```

status 필드 = 카드 (c) 요구 전부(보유자 세션+pid·이름·명령=패키지·시작·예상 종료). 첫 시도에서 `exit=$?`를 파이프 뒤에 찍어 tail의 0을 잘못 읽었다 — 종료 코드는 파이프 없이 캡처한다(측정 교훈).

## 3. 판정

- (a) 별도 표면 필요 여부: 표면은 이미 존재·가동. 결핍은 09-10 시점엔 표면 자체, 현재는 (i) 거부 종료 코드 3의 문서화, (ii) "무거운 실행"의 예시 패키지 명시(internal/cli 등), (iii) 이 카드의 통제군 기록 — 세 항목뿐이다(plan-audit D4가 같은 결론에 도달).
- (b) integration 재사용: 불가 — 병합 창과 실행 순번은 기록·수명이 다르다. §8과 카드가 동일하게 명시.
- (c) 필드: 전부 충족(위 status 출력).
- 통제군: 09-10 침입 3건 관측(부하 8~21, 세 레인 독립)이 "표면 없음 → 침범" 기록이고, 위 exit-3 재현이 "표면 있음 → 거부" 증명. 동시 600초 스위트 재실행으로 재시연하는 것은 부하 규율 위반이라 금지한다.
- 배포 여부: `moai slot` 동사는 제품 CLI. 규율 문면의 본래 집(레인 규칙 §8, `.claude/rules/local/`)은 로컬 전용이고, 미러 문서(resource-slot-lease.md)에 내부 카드 사정을 새기지 않는다(§25).
