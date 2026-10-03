# decision-index.md — SPEC-TODO-CARD-ISSUANCE-001

`interview.decision_gate: on` — 카드 t1454 의 plan 조립 중 표면화된 결정 중 운영자가 인터뷰에서 확정하지 않은 열다섯 가지. 이 파일은 상태 필드가 없다(생애 주기는 `spec.md` 만 가진다). `Operator verdict` 는 작성 시점에 비어 있고 킥오프 게이트가 읽는다. 라벨 어휘: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER. 모든 행은 무엇이 미해결이고 왜 그런지(Detect), 어떤 선택지가 있고 근거가 어디 있는지(Explain), 무엇을 묻는지(Ask)를 적으며 선호 답을 싣지 않는다. 선택지별 측정 근거와 계획이 진행하는 작업 기본값은 `design.md` §11 에 있다(`recommendation_mode: pull` 이므로 이 파일은 권고하지 않는다). 권위 앵커는 이 트리에 커밋된 파일과 절이다 — 앵커를 확인하지 못한 행은 FOUNDER 로 올렸다.

## Q1: M5 를 t1453 착지 뒤로 미루는가, 지금 같은 규칙 파일을 함께 고치는가, 게이트 미충족이면 무엇을 인도하는가? (D1)

- Label: FOUNDER
- Authority anchor: 없음 — 근거가 라이브 큐 카드 t1454 본문(미커밋)에만 있다("M5 는 두 카드 착지 뒤"). 보조 사실: t1453 의 plan 은 그 브랜치(`WT-github-flow-default`)에만 있고 이 트리에 커밋돼 있지 않다.
- Why unresolved: 선택지 (i) M5 는 게이트 명령이 열릴 때까지 마지막에 둔다, (ii) 지금 편집하고 t1453 이 절체 시점에 흡수하게 한다, (iii) M5 를 후속 SPEC 으로 떼어 낸다. 게이트 미충족 시 인도물도 (a) 초안만 내고 SPEC 은 요구 문면대로 닫는다, (b) SPEC 을 in-progress 로 붙들어 둔다로 갈린다. 핀 기준으로 t1453 병합 수를 어느 부모 경로로든 닿는 술어로 읽은 값은 0, t1448 대조군은 1 이다(`--first-parent` 형태는 develop 을 흡수한 카드 브랜치에서 착지 후에도 0 을 읽는 눈먼 선택자라 쓰지 않는다, `design.md` §13). 어느 쪽이든 규칙 파일 여섯 개와 그 템플릿 사본이 걸려 있어 결정은 운영자의 몫이다.
- Operator verdict:

## Q2: 두 관계 저장소(큐 `findings`, GTD `gtd_relations`)를 어떻게 통합하는가? (D2)

- Label: FOUNDER
- Authority anchor: 없음 — "어떻게"를 정한 커밋 문서가 없다. 맥락 앵커(예약일 뿐 결정이 아니다): `.moai/specs/SPEC-TODO-AUTO-PICK-001/spec.md` § B.7 과 REQ-TAU-013(관계 레코드는 저장소 하나에 묶이지 않는다, 통합은 t1454 몫), `.moai/specs/SPEC-RELATION-PICKUP-FILTER-001/spec.md` § A.3(두 체계를 섞지 말 것).
- Why unresolved: 선택지 (i) 단일 저장소로 이주, (ii) 저장은 분리하고 읽기 해석기와 어휘 확장으로 통합, (iii) 읽기 시점 합집합만. 두 저장소는 식별 공간(`tN` 대 `gtd-<16hex>`)·저장 형태(배열 대 PK 표)·보관 의미·쓰기 경로·어휘(여덟 대 아홉 종류)·중복 제거 키가 다르다. 스냅숏에서 `gtd_relations` 는 0행, `findings`+`archived_findings` 는 125개 고유 쌍이다. 이주는 보관·복원·큐 병합·이주 코드를 건드린다.
- Operator verdict:

## Q3: 카드→파일 간선은 어디에 싣는가? (D3)

- Label: FOUNDER
- Authority anchor: 없음 — 질문 자체를 정한 선행 권위가 없다. 맥락: `.gitignore` 의 `.moai/project/graph/` 항목(산출물은 추적 제외), `internal/graph/graph.go` 머리 주석(동일 트리 동일 출력 계약), `internal/graph/gtd_private.go` 의 비공개 사이드카 선례(GTD 투영). 선례는 GTD 항목 간선에 대한 것이고 카드→파일 간선과는 대상이 다르다.
- Why unresolved: 선택지 (i) 큐에서 만든 비공개 `card-edges.jsonl` 사이드카, (ii) 커밋된 증거(카드 귀속 병합 커밋과 변경 파일)에서만 `edges.jsonl` 에 간선 생성, (iii) 둘 다. 카드가 인용한 "커밋된 edges" 전제는 낡았고(산출물이 추적 제외), 남는 제약은 동일 트리 동일 출력·CI 와 로컬의 불일치 방지·비공개 상태 비유출이다. (ii)는 열린 카드의 예상 파일을 그래프에 못 싣고, (i)은 소비자가 오늘 없다. `moai graph build` 의 git 증거 추출 시간은 측정하지 못했다.
- Operator verdict:

## Q4: M2 스키마는 JSON 컬럼 하나인가, 속성별 컬럼인가? (D4)

- Label: FOUNDER
- Authority anchor: 없음 — 두 선례가 반대 방향이다. `.moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/spec.md` REQ-TCD-002(분류는 JSON 컬럼 하나), `.moai/specs/SPEC-TODO-CLAIM-LEASE-001/spec.md` REQ-TCL-001(임대는 컬럼 둘), SPEC-TODO-TRANSITION-STAMPS-001(스탬프는 컬럼별).
- Why unresolved: 컬럼 하나는 닿는 곳(`ensureColumn` 두 표, INSERT·SELECT 목록, parity, 동결 튜플 문자열 둘, JSON 투영)을 한 번씩만 바꾸고, 속성별 컬럼은 속성 수만큼 곱한다. 반대로 JSON 컬럼은 SQL 에서 속성 조회를 `json_extract` 로 해야 하고 속성별 NOT NULL·CHECK 를 걸 수 없다. 큐는 `LoadPure` 로 통째로 메모리에 올려 Go 에서 읽고, 보관 1,047행에 SQL 조회가 없다. finding 처분은 별도 표라 어느 쪽이든 컬럼 하나가 는다.
- Operator verdict:

## Q5: M1 제시를 MCP 호출자에게 어떻게 전달하고, `gtd engage` 가 분석에 합류하는가? (D5)

- Label: FOUNDER
- Authority anchor: 없음 — 맥락 앵커: `.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md` REQ-TA-010(기계 분석기는 `add` 와 `analyze` 에서만 돈다), `internal/cli/factory_m3_test.go` 의 `TestSD_AC014_MCPMatchesCLIWithProjectRoot`(MCP 결과 텍스트가 CLI stdout 과 같아야 한다).
- Why unresolved: 전달 선택지 (i) 결과 텍스트의 id 줄 뒤에 덧붙임(제시가 없을 때는 기존 패리티 테스트가 그대로 통과), (ii) `StructuredContent` 에 실음(Claude Code 가 이를 어떻게 보이는지 측정하지 못했다), (iii) 별도 읽기 도구(`todo_add` 는 그대로, 호출자가 먼저 미리보기를 불러야 알림이 보임). engage 선택지 (a) 제시만 출력(분석기 비합류, REQ-TA-010 문면 변경 없음), (b) 분석기에도 합류(REQ-TA-010 문면 변경 필요, 정확 중복 거절·소견 기록이 engage 에 생김), (c) 변경 없음. engage 의 현재 경로는 분석·분류 결정기를 모두 거치지 않는다.
- Operator verdict:

## Q6: 카드 병합 경로를 허용하는 독트린 개정은 어디까지인가? (D6)

- Label: FOUNDER
- Authority anchor: 없음 — 맥락 앵커: `.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md` § B.1 표("한 카드를 다른 카드에 접어 넣기"는 눈에 보이지 않는 변경이라 금지)와 § E `Out of Scope — 카드 흡수·병합`. 그 판정은 분석이 일으키는 변형에 대한 것이고 운영자가 직접 호출하는 명시적 동사에 대한 것이 아니다.
- Why unresolved: 선택지 (i) 운영자 호출 동사 하나만 신설하고 분석·`analyze`·`relate`·레인은 계속 금지, (ii) 리더·레인에게도 허용, (iii) 병합 동사 없이 `drop`+`edit` 조합을 유지하고 관계 기록만 개선. 개정되는 문면은 ANALYSIS §B.1 표와 §E, REQ-TA-014 의 "never folds" 문장, `gtd.md` 의 두 문장이고, 코드가 "병합 동사의 부재"로 강제하던 불변식이 "분석 경로에서 병합이 호출 불가"로 바뀐다.
- Operator verdict:

## Q7: 크기 하한·상한, 동시 진행 한도, 파생 깊이 상한, 허브 임계값, 유사도 표시 하한, 구성요소 깊이를 SPEC 에 고정하는가, M0 재측정에서 정하는가? (D7)

- Label: EVIDENCE-NEEDED
- Authority anchor: 해당 없음(EVIDENCE-NEEDED). 카드 본문의 "수치 기준은 본 저장소 실측으로"는 라이브 큐(미커밋)에만 있다.
- Why unresolved: 값을 정하는 데 필요한 M0 재측정 산출물이 아직 추적되지 않았다. 이 계획이 잰 분포(분위수, 하한별 알림 비율, 기록된 관계의 재현율, 병합·동시 진행 대리 지표)는 `research.md` §3 에 있지만, (a) 어느 분위수를 한도로 삼는지(P90·P95 등)는 판단이고 (b) 정답 집합이 없어 정밀도는 못 쟀고 (c) 큐 쪽 값은 스냅숏이라 실행 단계에서 달라진다. 선택지 (i) SPEC 이 값을 고정, (ii) 값은 M0 기준선 기록에서 읽고 SPEC 은 기록과 측정 명령만 가리킨다(기준선은 `.gitignore:235` 에 걸리는 `.moai/reports/` 가 아니라 추적되는 `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` 에 놓는다), (iii) 배포 규칙은 메커니즘만 쓰고 값은 로컬 전용 규칙(`.claude/rules/local/`, 이 저장소에서 추적됨)에 둔다. (ii)와 (iii)는 서로 배타적이지 않다 — 값의 출처와 값이 적히는 자리를 따로 정하는 질문이다. 배포 사본에 값을 쓰는 읽기는 사용자 프로젝트에 이 저장소 큐의 분포가 배포되고, 기준선을 가리키는 읽기는 템플릿 사본에 카드 id 와 사용자 프로젝트에 없는 경로가 들어간다는 비용이 있다(`design.md` §8).
- Operator verdict:

## Q8: 어휘 이웃의 척도와 표시 방식은? 소음을 줄이는 대신 재현율을 포기하는가? (D8)

- Label: FOUNDER
- Authority anchor: 없음. 맥락: `internal/kanban/backlog_analysis.go` 임계값 주석(근접 중복 임계가 낮으면 소견이 카드 수를 넘어 운영자가 목록을 읽지 않게 된다 — 오류가 비대칭이라 값이 높다).
- Why unresolved: 측정(`research.md` §3.3): 발행 시점 최근접 이웃 점수 중앙값 0.10, 하한 0.3 에서 알림이 뜨는 추가 비율 8.4%(최근 4.9%), 그러나 Jev·에이전트가 기록한 관계 상대를 상위 3개에서 찾는 비율은 하한 없음 52%, 하한 0.3 에서 6%. 선택지 (i) 기존 `TokenSetJaccard` 재사용에 표시 하한, (ii) idf 가중 척도 신설(하한 없음 재현율 65%), (iii) 하한 없이 항상 상위 3개 표시(소음이 91%대), (iv) 확실한 것(정규화 일치)은 항상, 어휘 이웃은 `--dry-run` 에서만. 점수와 척도 이름은 어느 선택에서나 함께 표기한다.
- Operator verdict:

## Q9: closed-at 은 컬럼으로 저장하는가, 기존 스탬프에서 읽는 접근자인가? (D9)

- Label: FOUNDER
- Authority anchor: 없음. 맥락: `.moai/specs/SPEC-TODO-TRANSITION-STAMPS-001/spec.md`(picked_at·dropped_at·archived_at 스탬프를 정했다).
- Why unresolved: 스냅숏에서 `archived_at` 은 보관 1,047행 중 76행에만, `dropped_at` 은 dropped 90장 중 1장에만 있다. 저장 컬럼을 더해도 과거 카드는 NULL 이고 새 카드에서는 기존 두 스탬프와 겹친다. 선택지 (i) 저장하지 않고 접근자 하나가 두 스탬프에서 읽는다, (ii) 컬럼 하나를 더해 done·drop·undrop·undone 이 같은 코드 자리에서 쓴다, (iii) 과거 카드는 git 착지 커밋 시각으로 소급한다.
- Operator verdict:

## Q10: 파생 부모의 단일 원천은 카드 속성인가, 관계인가? (D10)

- Label: FOUNDER
- Authority anchor: 없음.
- Why unresolved: M2 의 "스폰한 카드/부모"와 M3 의 `parent-of`·`follow-up-of` 는 같은 사실의 두 표현이다. 선택지 (i) 카드 속성이 원천이고 관계는 읽기 투영(자식당 부모 하나는 스칼라라서 자동), (ii) 관계가 원천이고 속성은 투영, (iii) 둘 다 쓰고 일관성을 검사. 측정: 파생 머리말이 있는 카드 473장 중 부모 id 가 큐에 없는 참조가 54건이고 본문 추론은 겹치는 분류를 만든다.
- Operator verdict:

## Q11: finding 처분(수용·병합·기각)은 무엇에 영향을 주는가? (D11)

- Label: FOUNDER
- Authority anchor: 없음. 맥락: `.moai/specs/SPEC-RELATION-PICKUP-FILTER-001/spec.md` REQ-RPF-001(`blocks`/`depends` 소견의 존재가 곧 미해결이라는 술어) — 순서쌍 종류에는 처분을 적용하면 이 술어와 충돌한다.
- Why unresolved: 선택지 (i) 기록만 한다(카드·순서·픽업 필터·표시 모두 불변), (ii) 처분이 있는 소견은 `list` 표시와 `machine-only` 표지에서 제외한다, (iii) 처분이 소견을 보관으로 옮긴다. 처분을 기록하는 동사가 오늘 없고, 기록된 125개 쌍 중 양쪽이 이미 done 이고 서로 언급하지 않는 쌍이 20개라 "소견이 처리됐는지"를 알 방법이 없다.
- Operator verdict:

## Q12: 구성요소별 부채 대장은 무엇으로 운반하는가? (D12)

- Label: FOUNDER
- Authority anchor: 없음. `sync-auditor.md` 의 `PASS-WITH-DEBT` 토큰은 있으나 대장은 어디에도 없다.
- Why unresolved: 선택지 (i) 큐의 카드(출처 `debt`, 구성요소 노드로 묶음) — 새 저장소가 없고 큐는 비공개·머신 로컬이다, (ii) 추적되는 파일(`.moai/` 아래 구성요소별 문서) — 검토 가능하지만 새 규약과 쓰기 주체가 필요하다, (iii) 감사 판정서 안의 표 — 흩어진다. "카드가 만든 결함은 병합 전 카드 안에서 고치고, 기존 결함은 대장으로"라는 규칙의 두 번째 가지가 오늘 갈 곳이 없다.
- Operator verdict:

## Q13: `add` 의 새 플래그(부모·출처·크기·예상 파일)는 명시 입력만 받는가, 본문에서 추론해 채우는가? (D13)

- Label: FOUNDER
- Authority anchor: 없음. 맥락: `.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md` § B.1(눈에 보이지 않는 자동 변경을 허용하지 않는다).
- Why unresolved: 선택지 (i) 명시 플래그만, 추론은 제시에서 읽기 전용 제안으로만, (ii) 본문에서 추론해 저장, (iii) 둘 다. 측정: 열린 카드 37장 중 본문에 경로가 보이는 것은 12장(32%), 파생 머리말 추론은 54건이 큐에 없는 부모를 가리킨다. `add` 의 폴스루 경로는 플래그를 받지 않는다.
- Operator verdict:

## Q14: 묶음은 어떤 메커니즘으로 레인에 직렬로 싣고, 임대 경로의 어디까지 바꾸는가? (D14)

- Label: FOUNDER
- Authority anchor: 없음. 맥락 앵커: `.moai/specs/SPEC-TODO-AUTO-PICK-001/spec.md` REQ-TAU-013(새 입력이 keep-set 과 임대 경로를 고치지 않는다), `internal/homestate/card_transition.go` 의 `guardAssign`(이전 카드 병합을 **할당** 간선에서 요구).
- Why unresolved: 오늘 `assign --after` 는 이전 카드가 병합되기 전에는 다음 카드를 레인에 할당하지 못하게 해서 "다음 카드를 같은 레인에 미리 걸어 둔다"가 불가능하다. 선택지 (i) 적재 동사만 신설하고 다음 카드 할당은 이전 병합 뒤 사람이 진행(임대 경로 불변, 자동 연속 없음), (ii) 팩토리 `cards` 행에 묶음 컬럼을 두고 선택 호가 묶음 예약·선행 미병합 건너뛰기를 읽는다(자동, 임대 경로 개정), (iii) 큐 쪽 발행 속성에 묶음을 싣고 선택 호가 읽는다. 선행 미병합 `after` 후보를 만난 선택 호가 오류로 끝나는지는 실행으로 확인하지 못했다.
- Operator verdict:

## Q15: 이 SPEC 이 더하는 경로 한정 규칙 파일과 스킬 본문도 40,000자 한도를 받는가? (D15)

- Label: DECIDED
- Authority anchor: `.moai/specs/SPEC-INSTRUCTION-BUDGET-SCOPE-001/spec.md` § 1 Context(훅 `charBudget = 40000` 이 `LoadReason` 을 게이트로 쓰지 않으므로 `paths:` 한정 규칙도 글롭 적재마다 잰다) 와 그 HISTORY 0.1.0 행(범위 결정 `(c) + (a)` — 문서 교리를 훅에 맞춘다, 훅 쪽을 좁히는 `(b)` 는 기각).
- Why unresolved: 해당 없음 — 앵커가 같은 조건의 같은 질문을 이미 정했다. 이 행은 새 `card-issuance.md` 와 편집되는 `gtd.md` 가 이 결정의 적용을 받는다는 사실만 기록한다. 동반 파일의 `paths:` 는 부모 패턴의 진부분집합이어야 한다는 같은 SPEC 의 규칙은 새 파일이 분할 산물이 아니라 신규 내용이므로 직접 적용되지 않지만, `kanban-dispatch*` 글롭과 자기 매칭하는 파일명은 피한다.
- Operator verdict:
