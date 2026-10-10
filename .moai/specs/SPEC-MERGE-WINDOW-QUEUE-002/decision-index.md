# SPEC-MERGE-WINDOW-QUEUE-002 — Decision Index

> `interview.decision_gate: on` — 인터뷰 없이 카드 지시로 진행한 plan 단계에서 확정되지 않은 결정의 장부. 라벨 어휘: `DECIDED` / `POLICY-COVERED` / `EVIDENCE-NEEDED` / `FOUNDER` 4종만.

### Q1: `moai factory merge ready`의 REFUSED 판정 중 어디까지 비제오 exit으로 바꾸는가?

- Label: `FOUNDER`
- Class: `product-level` — 배포 명령의 사용자 보이는 exit 코드 동작을 바꾼다.
- Authority anchor: (착지 검증 불가 — 이 질문을 기결한 선행 완료 SPEC의 행도, 설정 키도 없다. 가장 가까운 계약은 SPEC-FACTORY-LANE-AUTONOMY-001의 verdict 설계(`factory_merge.go` 상단 계약 주석: "every refusal here is a verdict, not a failure: the command exits 0")이며, 그것은 반대 방향의 정당화라 앵커로 쓸 수 없다.)
- Why unresolved: 카드 t1582는 "INVALID 판정 + nil 반환의 남은 인스턴스 수리"를 지시하지만, REFUSED 판정 전체를 비제오로 뒤집을지 아니면 측정 실패 클래스(재측정 레코드 부재/무효)만 뒤집을지는 미확정 — 전자는 FACTORY-LANE-AUTONOMY-001의 문서화된 verdict 계약과 충돌한다.
- Default: 측정 실패 클래스만 비제오, 창 경합 verdict(waiting·holder)는 exit 0 유지 (rule: preserves current behavior — 창 경합 verdict의 현재 동작을 보존하는 유일한 옵션이며 undo는 본 SPEC 커밋의 단일 revert)
- Alternate: REFUSED 판정 전체 비제오 (FACTORY-LANE-AUTONOMY-001 verdict 계약의 명시적 개정을 수반 — 별도 계약 갱신 필요)
- Operator verdict: Operator verdict for SPEC-MERGE-WINDOW-QUEUE-002 Q1 (card t1582): only the measurement-failure class of REFUSED verdicts of moai factory merge ready exits non-zero; window-contention verdicts (waiting, holder) keep exit 0 (operator's selected answer: "측정 실패만 비제로"). Recorded in decision board record d-20261009T235134Z-0afd (standing). This is the row's recorded Default; the alternate (all REFUSED non-zero) was offered and not chosen.

### Q2: R7-1 원자화의 잠금 기계를 무엇으로 쓰는가?

- Label: `FOUNDER`
- Class: `implementation-level`
- Authority anchor: (해당 없음 — 구현 기계 선택, 선행 기결 행이나 설정 키 없음)
- Why unresolved: hold 기록 + release의 원자화를 기존 `withIntegrationLockMutation` 위에 얹을지(정책 파일 쓰기가 이 mutation의 범위에 들어가는지의 설계 확인 필요) 아니면 새 잠금 경로를 만들지는 run 단계에서 `ReleaseIntegrationLock`의 승격 로직과의 정합을 읽어야 확정된다.
- Default: 기존 `withIntegrationLockMutation` 기계 재사용 — 새 잠금 경로를 만들지 않는다 (rule: smaller user-visible surface — 직렬화 표면의 증가 없음, undo는 본 SPEC 커밋의 단일 revert)
- Alternate: policy 쓰기와 lock release를 묶는 새 원시(primitive) 도입
- Operator verdict: DEFAULT-APPLIED (구현 수준 — run 단계 진입 시 M4 절차에 따라 적용; plan close 시 채움)

### Q3: R7-2 수리의 "빈 스윕" 판정 기준은 무엇인가?

- Label: `FOUNDER`
- Class: `implementation-level`
- Authority anchor: `.moai/specs/SPEC-MERGE-WINDOW-QUEUE-001/spec.md` §C.3 REQ-MWQ-015 — "a count of zero, a runner-reported empty sweep ... shall make the record invalid"의 계약을 이 SPEC이 정밀화한다 (갱신 자체는 카드 t1582의 fix 방향 지시로 확정 — "총 테스트 수 기준 판정+패키지 skip을 종료로")
- Why unresolved: REQ-MWQ-015의 "runner-reported empty sweep" 정의를 마커 기반에서 총 per-test pass 0 기반으로 바꾸는 정밀화는 본 SPEC의 권한 안이지만, 그 계약 문구의 확정은 plan-audit의 계약 검토 대상이다.
- Default: 총 per-test pass 0만 빈 스윕으로 거부 — 마커의 존재는 판정 신호가 아니다 (rule: smaller user-visible surface — 거짓 거부를 제거하는 방향, 카드 fix 방향과 일치). D5 분할(v0.2.0)과 정합: 스윕 판정은 REQ-MWQ2-003, 빈 스윕 무효는 REQ-MWQ2-009.
- Alternate: 마커가 패키지 skip output 이벤트 안에 있을 때만 무시하는 세분화 방식 (구현 복잡도 증가, 방어 이득 미측정)
- Operator verdict: DEFAULT-APPLIED (구현 수준 — plan close 시 채움)

---

> Q1은 운영자 처분 대상(FOUNDER product-level) — 자율 Kickoff 전 리더가 AskUserQuestion으로 운영자에게 전달해야 한다. Q2·Q3은 구현 수준으로 plan close 시 Default가 채워진다.
