# SPEC-ROLE-INJECTION-BUDGET-001 — Relocation Ledger (relocation-ledger.md)

> M2 산출물 (REQ-RIB-003, AC-RIB-002). 재작성 전 36개 role-core 영역의 전수 재배치 감사.
> 규칙: 모든 잘려나간 조각은 목적지를 이름 대고, 구속 절은 의미 보존 압축 재작성으로 core 또는 상시 스텁에 남는다.
> 삭제 없음. `buildRoleCore` 표지 검증 무편집. 트리: 카드 워크트리 t1617 (WT-3-2-1), HEAD `2b1fe8e8b` (M1 커밋) 위에서 편집.
> 결과: core 36영역 17,793/18,114 → **31영역 3,784 UTF-16** (예산 3,946 · 설계 목표 3,800 이내). 5개 복합절은 상시 스텁으로 귀속(대량 재분류 아님 — 5/36).

## §A. 판정 요지 — 5개 복합절의 스텁 귀속

REQ-RIB-003/§D 는 구속 절의 처지를 「core 압축 재작성 **또는** 상시 스텁」으로 허용하고 대량(大規模) 재분류만 금지한다. ALB-0272·0276·0284·0294·0303 은 각 2~4 개의 의무를 묶은 복합절로, 3,946 core 예산 안의 압축으로는 의무 뉘앙스(전송 결과 판독, 크론 상한, 병합 창 절차 접점)를 잃는다 — REQ-RIB-004 가 금지하는 의무 포기에 해당하기 전에, 이 5개만 스텁에 전문 보전했다(5/36 = 비대량). 비용은 REQ-RIB-011 의 `stub-delta:` 기록과 rule-authoring 비산 문으로 정산한다(스텁 8,583 → 9,915 UTF-16, 바이트 델타 > 1,000).

## §B. 36영역 감사 표

| # | 행 id | before 요지 | after 위치 | after 요지 / 목적지 확인 |
|---|---|---|---|---|
| 1 | ALB-0254 | 리더가 큐 유일 생산자 | core-압축 | "Only the operator's requests enter the queue…" — 의미 보존 |
| 2 | ALB-0255 | 상시 생산원 | core-압축 | 조건 집합 본문은 `gtd.md` § Standing sources 가 소유(선행 확인) |
| 3 | ALB-0256 | 승격은 운영자 행 | core-압축 | `--auto` 정합 문단은 영역 밖 비표지 본문에 원형 생존 |
| 4 | ALB-0258 | lane 임대 예외 | core-압축 | 기타 큐 변경 금지 본문은 `gtd.md` § --auto 소유 |
| 5 | ALB-0260 | findings 부착 불행동 | core-압축 | — |
| 6 | ALB-0261 | pre-dispatch 교차검증 | core-압축 | 수단 상세(cards § The pre-dispatch cross-check)는 companion 소유 — "by hand" 절차 삭제 |
| 7 | ALB-0262 | 완결 SPEC 커버 확인 | core-압축 | — |
| 8 | ALB-0263 | 보고 불거부 | core-압축 | 왜 wording 뿐인지는 cards § cross-check 소유 |
| 9 | ALB-0264 | Card Cross-Check 섹션 | core-압축 | 기계 검증 절차는 cards § Report milestones 소유 |
| 10 | ALB-0266 | Class A 증거 | core-압축 | 근거는 cards § Card classes 소유 |
| 11 | ALB-0269 | 큐 위임 채널 | core-압축 | — |
| 12 | ALB-0270 | nudge≠위임 | core-압축 | 통지 경계 본문은 cross-session-messaging § idle notice 소유 |
| 13 | ALB-0271 | 발신 언어 | core-압축 | 분류 근거는 detail § Dispatch language 소유 |
| 14 | ALB-0272 | 발신 형식·전송 판독 | **스텁** § The dispatch cycle | 복합절(형식+판독) — 예시 블록은 detail § Dispatch format 소유, 전문은 스텁 보전 |
| 15 | ALB-0273 | 과업 목록 | core-압축 | 측정 선례는 mechanics § task list 소유 |
| 16 | ALB-0275 | card-review 단계 | core-압축 | 단계 순서·증거·천장은 detail § The card-review stage 소유 |
| 17 | ALB-0276 | 명시 대기·스톨 감시 | **스텁** § The dispatch cycle | 복합절(대기+워치독+사다리+cron) — watchdog 본문은 moai-lane-watchdog 스킬 소유 |
| 18 | ALB-0277 | 부수 상주 | core-압축 | 위임 행렬은 manager-lead.md § Deputy 소유 |
| 19 | ALB-0278 | RECOMMEND 요약 | core-압축 | — |
| 20 | ALB-0279 | 부수 무결정권 | core-압축 | retained 목록 상세는 detail § Deputy mode 소유 |
| 21 | ALB-0280 | 완료는 판독 | core-압축 | verdict-file/card-review 본문은 영역 밖 비표지 본문에 원형 생존 |
| 22 | ALB-0284 | CodeRabbit 2조건 | **스텁** § CodeRabbit | endpoint 측정은 gates § CodeRabbit 소유 — 술어 전문은 스텁 보전 |
| 23 | ALB-0286 | /clear 인수 | core-압축 | 메시지 구조는 cards § /clear handoff 소유 |
| 24 | ALB-0287 | /clear 1회·이동 후 | core-압축 | — |
| 25 | ALB-0290 | 런처 진입 | core-압축 | 런처 표는 mechanics § Isolation 소유 |
| 26 | ALB-0291 | Codex 진입 동사 금지 | core-압축 | — |
| 27 | ALB-0292 | done은 L2 전용 | core-압축 | L1/L2 경계는 worktree-integration § Glossary 소유 |
| 28 | ALB-0293 | 미푸시 유일 사본 | core-압축 | — |
| 29 | ALB-0294 | 세션 트리 체류·이동 금지 | **스텁** § Isolation and integration | 복합절(체류+금지+구성 규칙+clear 1회) — 이유 본문은 cards § Isolation rationale 소유 |
| 30 | ALB-0295 | 새 카드 새 트리 | core-압축 | — |
| 31 | ALB-0296 | WT- 브랜치 | core-압축 | slug 표·운반체 상세는 mechanics § Isolation · cards § PR-title 소유 |
| 32 | ALB-0297 | 가드 경로 규칙 | core-압축 | 거부 형상 실측은 mechanics § Isolation 소유 |
| 33 | ALB-0298 | 레인 로컬 검증 | core-압축 | 사고 기록은 gates § incident record 소유 |
| 34 | ALB-0299 | 백그라운드 부하 금지 | core-압축 | 사고 제2원인 본문은 gates § incident record 소유 |
| 35 | ALB-0300 | env 격리 형태 | core-압축 | 변이형·실측은 mechanics § Verification load 소유 |
| 36 | ALB-0303 | 자가 통합 | **스텁** § Isolation and integration | 복합절(자가 병합+변형 스코핑+창 절차) — 창 전문은 mechanics § Integration 소유 |

## §C. 특수 감사 행

- **배포본 +321 로컬 꼬리**(구 ALB-0303 영역 내 "moai worktree sweep…" 문장): 로컬 전용 내용. 목적지 확인 — `worktree-integration.md` § Hoist a tree's evidence before disposing it 와 `factory-dispatch-mechanics.md` § The lane's standard landing 6단계(Sweep) 가 동일 본문을 소유. core에서 정리, 템플릿 미러는 꼬리 무상태 유지 → 재작성 후 양 트리 바이트 동일(갈림 해소).
- **「### Dispatch format」·「### Lane waits are explicit…」빈 제거**: 유일 내용이던 표지 영역이 스텁으로 귀속되어 전체 본문에서 제목 정리.
- **gitflow 포인터 정합(REQ-RIB-010)**: 배포본 `## Isolation is provisioned by MoAI…` 절 생존(런처·L2·유일 사본 [HARD] 잔존), 세션 체류·이동 금지 [HARD] 는 스텁 § Isolation and integration 에 생존 — 상시 표면이므로 로컬 규칙 지목(`gitflow-lane-protocol.md` §1 → factory-dispatch.md 의 Isolation 절)은 살아 있는 절로 해석되고, 금지 본문은 한 홉(스텁)에서 읽힌다. 로컬 규칙 파일은 로컬 전용이라 무편집.
- **원장 행**(internal/template/testdata/binding_ledger.json): 36행 모두 after-text 재작성 processing='rewrite' — 31행 role-core 유지, 5행 location 을 `always:.claude/rules/moai/workflow/factory-dispatch-core.md` 로 전환. 행 수 변화 없음. anchor `a2a184ad3` 의 before_text/start_line 불변(REQ-ALB-015 원장 테스트 GREEN 확인).
