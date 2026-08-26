# plan.md — SPEC-LEAD-DEBOTTLENECK-001

Tier M 구현 계획. 마일스킨 순서는 결정 가역성 — 변경 가능성이 가장 높은 결정(권한 매트릭스, agent 본문 구조)을 앞에, 기계적/검증 단계를 뒤에 둔다.

## §A Context

- 리드 병목의 기계적 원인은 단일 지점이다: `.claude/agents/moai/manager-lead.md:10`의 `tools:` 허용목록에 `SendMessage`/`ListAgents` 부재. 본문은 이미 목표 상태를 선언("handles cross-session messaging")하지만 도구가 없다.
- 런타임은 이미 지원한다 (`orchestration-mode-selection.md:124` 실측 기록). 누락은 허용목록 선택이다.
- Go 변경 불필요 (조사 실측): 런칭 코드(`internal/cli/{kanban,factory,launcher}.go`, `internal/hook/session_start_{kanban,factory}.go`)는 에이전트 도구 허용목록을 제어하지 않는다. deputy는 에이전트 정의 + 교리 계층 구성물.
- 분배 표면 3종: `manager-lead.md` 1 + `kanban-dispatch{,-detail}.md` 2. 전부 `internal/template/templates/` 미러 대상 (현재 local==mirror 확인됨).

## §B Known Issues

1. **GLM unnamed 위험**: deputy는 반드시 UNNAMED spawn — named spawn은 in-process teammate으로 전환되어 결과를 반환하지 않는다 (manager-lead 본문 기존 위험, 상속).
2. **정지 티메이트 부활**: deputy가 TaskStop된 세션에 이름으로 메시지를 보내면 transcript에서 부활한다 (SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001 교리 — agent 본문 deputy 절에 인용 추가).
3. **routing-object 유실**: in-process mailbox가 이름을 가로챌 수 있음 (단일 대조 실험 기반 — 재발송 비용은 낮고 미발견 비용은 정지된 보드이므로 재발송 프로토콜 채택).
4. **rapid-burst refusal**: deputy의 다중 발송 팬아웃이 inbox 용량 초과로 거부될 수 있음 — 발송 결과를 읽고 거부를 보고해야 한다.
5. **매트릭스 확정 — 해소(RESOLVED 2026-08-26)**: spec.md §4 매트릭스는 리드 판정(a) 확정(위임 5축 = 실측 병목축 일치, 유지 6항 = 카드 경계 무오차) 후 Implementation Kickoff Approval에서 운영자 비준(매트릭스 비준 + run 승인 + 자율 진행 모드, 2026-08-26). §4 본문은 비준본 그대로 유지.

## §C Pre-flight

- [x] 매트릭스 초안(spec.md §4) 리드/운영자 확정 — 해소됨 (리드 판정 + 운영자 kickoff 비준, 2026-08-26 — §B.5)
- [ ] baseline 측정 재실행 (AC-009 "before" 계수 — 구현 전 리드 세션 시나리오 녹화)
- [ ] 중립성 카탈로그 §25.1 숙지 (미러에 금지되는 내용 클래스)

## §D Constraints

- PRESERVE: `internal/cli/**` 전체 Go, 타 에이전트 파일, `CLAUDE.md`, 12-에이전트 카탈로그, `kanban-dispatch.md` 기존 [HARD] 전부(확장 전용), UNNAMED 규율, queue-on-disk 채널, 동시 write-capable 금지, `manager_lead_depth_test.go` 가드.
- 템플릿 미러 + `make build` + 중립성 (로컬 사본은 SPEC 출처 가능, 미러는 중립화).
- always-loaded 비용 규율: `kanban-dispatch.md` stub은 always-loaded — deputy 절 추가분이 1,000B 초과 시 증가분 근거를 커밋 본문에 명시 (`rule-authoring.md` (b)).

## §E Self-Verification

§D.0 RED-now 표(acceptance.md)의 명령을 M1/M2/M3 종료 시 동일하게 재실행하고 결과를 progress.md §E.2에 기록한다. 특히: depth-seal 가드 green 유지, mirror 동일성, 중립성 0히트, Go diff 0.

## §F Milestones

### M1 (High) — deputy 헌장: 매트릭스 + agent 정의 확장
결정 가역성 최상위 — 본 마일스킨의 산출물이 매트릭스이며 나머지 전부가 여기에 의존한다.
1. `manager-lead.md` `tools:`에 `SendMessage, ListAgents` 추가 (CSV 유지) — AC-001
2. 본문에 "Deputy dispatch surface" 절 추가: 매트릭스 코드화 + `DEPUTY-RETAINED-BY-LEAD` 리터럴 마커 + routing 재발송 프로토콜 + 정지 티메이트 부활 금지 인용 — AC-002/003/005
3. 템플릿 미러 동기화(중립화) + `make build` — AC-007(부분)
4. depth-seal 가드 재실행 — AC-006

### M2 (High) — 교리 확장: kanban-dispatch deputy 표면
1. `kanban-dispatch-detail.md` § The lead works through manager-lead에 deputy 모드 추가 (작업 모드·매트릭스·위임 형태)
2. `kanban-dispatch.md` stub에 deputy [HARD] 경계 절 추가 (기존 [HARD] 보존 — 개수·핵심 문구 검증) — AC-008
3. 팩토리 문서 표면은 detail companion § Factory in-lane 3-stage에 한 줄 연결로 반영 (별도 파일 신설 없음)
4. 템플릿 미러 동기화 + `make build` + 중립성 스캔 — AC-007

### M3 (Medium) — 기계적 검증
1. 런타임 프로브: UNNAMED deputy → named 세션 SendMessage 발송 형상 관측 (AC-004 — 테스트 컴패니언은 리드 세션이 시나리오 시작 시 기동; 프로토콜 고정 토큰 `LEAD-MERGE-APPROVED`/`FINAL VERDICT:`/`RECOMMEND:` 사용)
2. 2+ 레인 시나리오 before/after 리드 턴 점유 계수 (AC-009 — before는 구현 전 Pre-flight에서 녹화)
3. 머지 승인 0 누락 + 게이트 우회 0 확인 (AC-010)
4. 전체 diff 검사: Go 경로 0회 (AC-011)

## §G Anti-Patterns

- **판정 승격**: deputy의 1차 권고를 최종 판정으로 치환 — "판정의 소재" 위반. 매트릭스의 보고/판정 분리가 방어선.
- **교리 재작성**: kanban-dispatch [HARD] 절을 deputy 서사에 맞게 다시 쓰기 — 확장 전용 위반. 문구 보존 검증(AC-008)으로 잡는다.
- **미러 누락**: `.claude/` 직전 편집 후 템플릿 미러 없이 커밋 — Template-First 위반.
- **이름있는 deputy spawn**: GLM 위험 전환. UNNAMED 규율 유지.
- **채널 승격**: SendMessage를 dispatch의 원천으로 취급 — queue-on-disk 불변 위반.

## §H Cross-References

- spec.md §4 매트릭스 (본 plan의 M1 입력)
- acceptance.md §D.0 RED-now 표 (모든 마일스킨 종료 시 재측정 기준)
- SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001 (deputy 메시징 위험 상속 원천)
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1/§25.3 (중립성 카탈로그·기여자 체크리스트)
- CLAUDE.local.md §2 [HARD] Template-First Rule (미러 + make build 규율)
