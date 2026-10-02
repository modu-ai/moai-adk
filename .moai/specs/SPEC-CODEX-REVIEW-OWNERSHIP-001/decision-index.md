# decision-index.md — SPEC-CODEX-REVIEW-OWNERSHIP-001

`interview.decision_gate: on` — 카드 t1422 조립 중 표면화된 결정 중 운영자가 인터뷰에서 확정하지 않은 것. `Operator verdict` 는 작성 시점에는 비어 있었고, 2026-10-02 에 운영자 위임(Jev `jev-1.13.0`)·리더 승인으로 전 행이 채워졌다. **Q6·Q7 은 신뢰도 0.5 미만이라 PROVISIONAL** 이며 리더에 통지됐다. 라벨 어휘: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER. 번호는 plan.md §G 해결된 결정과 일치한다. 각 행은 무엇이 미결이었는지와 왜인지만 적으며 권고를 싣지 않는다.

**권위 등록부 점검.** 권위 인용은 커밋된 산출물 — product.md, 완료 SPEC 의 HISTORY·`## Amendments` 행, `.moai/config/sections/*.yaml` 운영자 설정, 프로젝트 헌장 — 만 쓴다. 미커밋 자료(큐 카드 본문, `.moai/reports/**` 증거)는 권위가 아니다. Jev 판정(`.moai/reports/t1422/jev-decisions.md`)은 표시 전용 신호라 권위 등록부 밖이다. 아래 모든 행의 후보 권위를 커밋 트리에서 확인하지 못했으므로 DECIDED·POLICY-COVERED 로 라우팅한 행은 없다. verdict 칸의 Jev 응답은 운영자가 위임한 판단의 기록이지 권위 인용이 아니다.

### Q1: `tree_scope` 를 템플릿에 주석 예시로만 싣는가, 라이브 키(인벤토리·콘솔 필드·i18n 포함)로 싣는가?

- Label: EVIDENCE-NEEDED
- Authority anchor: 없음
- Why unresolved: 라이브 키로 싣는 비용은 `shipped_key_reader_test.go` 가 새 shipped 키를 통과시키는지에 달려 있는데 그 측정이 없다. 이 레인은 그 가드를 실행해 보지 않았고, 인벤토리(804항목)·`schema_sections.go`·`fieldsets.templ`·`i18n.js` 4로케일 확장 범위도 추정이다.
- Operator verdict: Jev (operator-delegated), comment_example — 템플릿에 주석 예시로만 배포, 구조체는 파싱, 인벤토리·스키마·콘솔·i18n 불변, confidence 0.74, 2026-10-02

### Q2: 키 이름과 값 — `workflow.codex.review_gate.tree_scope` = `review` | `skip` 으로 정하는가?

- Label: FOUNDER
- Authority anchor: 없음 — 기존 규약(snake_case 키, 소문자 열거 값, 부재=배포 기본)은 형식을 정할 뿐 이름을 정하지 않는다. 결정 D1 은 "새 `review_gate` 키"까지만 정하고 이름·값 집합은 이 SPEC 작성자에게 위임했다.
- Why unresolved: 이름(`tree_scope`/`non_card` 등)과 값 어휘는 외부 설정 표면이라 한 번 배포되면 되돌리기 어렵다. 부재·알 수 없는 값을 `review` 로 읽는 방향도 이 SPEC 이 제안한 것이며 커밋된 권위가 정한 것이 아니다.
- Operator verdict: Jev (operator-delegated), `workflow.codex.review_gate.tree_scope` = review|skip (최종), confidence 1.00, 2026-10-02

### Q3: 온디맨드 자기 리뷰 표면 — 전용 도구 `codex_review`/`glm_review` 인가, 감사 도구에 `cardDiff` 대상을 추가하고 보유를 넓히는가?

- Label: FOUNDER
- Authority anchor: 없음 — 결정 D2 의 Jev 신뢰도 0.44 는 임계 0.5 미만이고 Jev 는 권위 등록부 밖이다. 코디네이터 지시(조건부 구속)는 큐/대화 산출물이다.
- Why unresolved: 비교 표(plan.md §B.1)가 측정 가능한 근거를 두 열로 갈랐고 불리한 행(도구 수·`project_root` 문서·i18n 정합, +≈70 LOC, +≈11 파일)도 실재한다. 어느 열의 비용을 택할지는 측정이 아니라 선호 판단이다.
- Operator verdict: Jev (operator-delegated), new_review_tools(전용 도구), confidence 0.44 (D2 intake 시점 값 — 임계 미만이었으나 리더가 세 조건(비교 표·결함 (b) 수리·결과 advisory)을 걸어 승인했고 plan.md §B.1 과 REQ-CRO-008/010 이 조건을 충족), 2026-10-02

### Q4: t1404 를 이 SPEC 착지 뒤 잔여(진부함 감지·런타임 경로 확장·장부 분류·검토자 빌드 노후)로 편집하는가, 닫는가, 유지하는가?

- Label: FOUNDER
- Authority anchor: 없음 — 큐 카드 본문은 미커밋 출처다.
- Why unresolved: 항목별 처분(plan.md §D)은 흡수 2·분리 6 이지만 분리 항목 중 T4·T5 는 관측하지 못한 전제(원 처분 보고서 부재)에 기댄다. 카드 소유자가 잔여를 가치 있다고 보는지는 코드로 정해지지 않는다.
- Operator verdict: Jev (operator-delegated), edit_to_residual — 착지 뒤 리더가 t1404 를 잔여로 편집, 착지 전에는 닫지 않음(sync 단계 인계 항목, 코드 변경 아님), confidence 0.92, 2026-10-02

### Q5: 이 저장소의 추적된 `.moai/config/sections/workflow.yaml` 에 `review_gate.enabled: true` 를 반영하는가, 그리고 primary 로컬 `tree_scope: skip` 은 누가 언제 적는가?

- Label: FOUNDER
- Authority anchor: 없음 — 추적 `workflow.yaml` 에는 `review_gate` 키가 없고(`grep -c review_gate` → 0, 본 트리), primary 사본의 `enabled: true` 는 로컬 사본이다. 설정이 이 질문을 문면 그대로 덮지 않는다.
- Why unresolved: 반영하면 레인 카드 스코프 Stop 게이트가 켜져 카드 리뷰 단계와 이중 리뷰가 된다(plan.md §E). 반영하지 않으면 레인에서는 카드 리뷰 단계가 유일한 수단이다. 운영자의 의도 설정일 수 있어 코드가 대신 정할 수 없다.
- Operator verdict: Jev (operator-delegated), do_not_commit — 추적 workflow.yaml 에 `enabled: true` 를 넣지 않음, 리더가 primary 의 비추적 로컬 설정에 `tree_scope: skip` 을 적음(운영자 소유 파일, 이 SPEC 은 편집하지 않음), confidence 0.97, 2026-10-02

### Q6: 자기 리뷰의 모델 해상 — 감사 핀(`workflow.audit.*`)을 적용하는가, 게이트와 같이 핀 없이 가는가?

- Label: FOUNDER
- Authority anchor: 없음 — `.moai/config/sections/workflow.yaml` 의 감사 핀 주석은 핀이 감사 진입점에 적용되고 작업 위임에는 적용되지 않는다고 적지만, 자기 리뷰가 어느 쪽인지는 문면이 덮지 않는다.
- Why unresolved: 자기 리뷰는 감사도 작업 위임도 아닌 제3의 경로다. 핀을 적용하면 운영자의 감사 독립성 설정이 개발자 자기 점검에 번지고, 적용하지 않으면 모델이 기본값에 맡겨진다.
- Operator verdict: Jev (operator-delegated), no_pins — 감사 핀 미적용, 모델 = 선택 입력 `model` 또는 백엔드 기본, confidence 0.22 (< 0.5, **PROVISIONAL** — 리더에 통지됨; 되돌림 비용은 해상기 호출 한 번, plan.md §B.1), 2026-10-02

### Q7: CLI 거울(`moai self-review`)을 이 SPEC 범위에 넣는가?

- Label: FOUNDER
- Authority anchor: 없음 — 결정 D2 는 "CLI 거울 포함"을 잠정 형태로 적었을 뿐이다.
- Why unresolved: 규약("MCP 와 CLI 는 같은 구현")과 MCP 서버 노후 우회라는 이점이 있으나 약 70 LOC·REQ 1·AC 1 의 비용이 든다. 단순성 사다리 1단("이것을 지어야 하는가")에 걸리는 질문이다.
- Operator verdict: Jev (operator-delegated), include — 가장 낮은 우선순위의 폐기 가능 마일스톤(M4)으로 포함, confidence 0.38 (< 0.5, **PROVISIONAL** — 리더에 통지됨; 폐기 규칙은 plan.md §G), 2026-10-02

### Q8: 감사 도구의 `baseBranch` 해상 체인 정렬(감사 도구 쪽 `cardDiff` 노출)을 후속 카드로 분리하는가?

- Label: FOUNDER
- Authority anchor: 없음
- Why unresolved: 결함 (b)는 새 도구의 카드 스코프로 닫지만 감사관이 카드 diff 를 감사할 때 쓰는 `codex_audit`/`glm_audit` 의 대상 해상은 그대로다. 그 비일치를 이 SPEC 이 안고 갈지 별도 카드가 닫을지가 미결이다.
- Operator verdict: Jev (operator-delegated), split_followup — 감사 도구는 이 SPEC 에서 불변, 후속 카드 필요(발행은 리더 소관, 레인은 발행하지 않음), confidence 1.00, 2026-10-02
