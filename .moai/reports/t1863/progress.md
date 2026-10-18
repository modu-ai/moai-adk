# t1863 진행 기록 (lane-18, Class C plan→run→sync)

카드: t1863 [운영자 지시 10-18·리더 배치·t1849 축소 후속·Class C·P2] 리더 시작 이어받기(t1849) 축소 재계획
run: tn1p4e · lane: lane-18 · 트리: `.moai/worktrees/t1863` · 브랜치: `WT-10-18-t1849` · base `2178966bb`(임대 시점 main)

## 1. 이월 입력 (t1849)

- t1849 = SPEC-FACTORY-LEADER-ADOPT-001(Tier M, 15 REQ) plan 감사 3라운드: iter1 FAIL 0.69 → iter2 FAIL 0.78 → iter3(델타) **FAIL 0.84 상한**(codex rcpt-130ae0e02eb75497a1cfd24d, blocking 4).
- 잔여 블로커(iter3): D14(AC-015, `--spawn` 사전검사가 AC-009 예외를 무시하는 변이 통과), D15(AC-013, 양성 사례 부재로 `return false` 변이 통과), D16(AC-014 금지 부분문자열 vs plan M3 수식어 삽입 지시 모순), D17(progress 계수 14→15 미갱신).
- D14-D17 문구 정정은 커밋 28485201a로 반영됐으나 **미감사**(t1849 progress 명시: "NOT audited-PASS").
- hold 룰링: board d-20261017T143532Z-7b17 + 상위 규칙 d-20261015T094733Z-9e76(규칙 B). 트리·판정 보존.

## 2. 결정 기록

- **D-t1863-1 (스코프 분할, 10-18)**: 핵심만 새 SPEC으로 — 채택: REQ-001(파인더: active·소유자 사망·비종말 카드 보유), 002(신규 시작 트리거), 004 재구술(기존 `--factory-run <id>` 의도적 인수 경로의 원자적 보장 — 대화형 선택지가 아니라 안내가 지목하는 명령), 005(거부 시 무기록 중단), 008(비대화형 안내 블록+비영 종료, DP-1), 009(예외 집합: resume 옵션·--factory-run·fresh 아님·고아 없음 → 현행 동작), 010(타 run 무기록), 011(glm≡cc 평행), 012(KST 렌더링), 014(안내 문서화 — D16 은 "수식어 삽입"이 아니라 **문장 치환**으로 해소). 분리 보류(탈출 변이 잔존, 후속 카드 제안): 003/006/007(대화형 adopt/new/cancel 선택), 013(스트림 대화성 판정), 015(--spawn 사전 거절). 근거: iter3 mutant 표 — 001/004/009/012의 변이는 이미 잡히고, 블로커는 전부 보류 항목+문구 계열.
- **SPEC id: SPEC-FACTORY-LEADER-ADOPT-CORE-001** — 리더에게 즉시 보고(카드 spec_id 보정 요청, kickoff 전제).
- **리더 룰링 접수(10-18)**: (1) spec_id 보정 공식 동사 부재·DB 직접 수정은 운영자 지시 범위 밖 → 운영자 결정 대기 항목으로 리더 큐에 등록됨. **정지 조건: plan 감사 PASS/PWD 확정 후 kickoff 기록 전에 멈추고 리더에게 통지 — 보정 없이 kickoff 시도 금지.** (2) 분리 보류 항목(003/006/007·013·015)은 규칙 B 후속 후보로 기록만(v3.2 입장 동결 9891 — 큐 미발행). 분리 근거: iter3 mutant 표에서 블로커 D14/D15/D16이 전부 이 항목들+문구 계열에 집중(001/004/009/012 형태는 견고) — 핵심 채택 경로를 먼저 착지시키고 탈출 변이 표면은 별도 감사 기회에 닫는 것이 상한 소진(0.84×3)의 재발 방지.
- manager-spec 스폰(SPEC-FACTORY-LEADER-ADOPT-CORE-001 저작, t1849 판정·REQ 원문·감사 교훈 5건 적재). 완료 통보 대기.
- 11:0xZ(02:0xZ UTC) — **저작 완료**: spec.md(REQ-FLAC-001..010, GEARS)·plan.md(M1 파인더→M2 인수 기록 경로→M3 거부 게이트+안내→M4 문서 앵커, 선형)·acceptance.md(AC-FLAC-001..010, 변이 명명·E1-E6 이 트리 HEAD 2178966bb 실측: E1 0/3 RED·E2 0/9 RED·E3 10/10·E4 9/9 GREEN·E5 앵컨 부재×10·README 4개 로케일 금지 부분문자열 :70 실재)·progress.md(§E.1 audit-ready)·decision-index.md(FOUNDER 3행: Q1 refuse-vs-mint·Q2 통합 인수 기록 경로·Q3 터미널 무관 거부). lint `✓ No findings`(설치 빌드 ee981a932). 판단: Q2(모든 --factory-run 인수를 AdoptRun 통합 — 운영자 반전 가능, M2+AC-003 국소)·Q3(대화성 판정 이월의 직접 귀결 — 대화형 시작도 거부, 별도 카드 전까지)·M2 선순위(존치 결정 가역성)·E5 재설계(RETIRE 형제가 docs 에 adopt 단어 추가로 선행 RED 사망 → 앵커-스팬 grep)·앵커 2개 축소(chooser 부재). 미커밋(감사 후 레인 커밋).
- 다음: plan-audit iter1 스폰.
- 02:1x-02:3xZ — **plan-audit iter1: FAIL 0.94**(점수는 Tier M 0.80 통과, codex 필수 게이트 콘텐츠 실패 rcpt-a5d3cfca67ec1502a6f5873d, must_pass 9/10, blocking 4): D1(치명·설계: AdoptRun 이 lead_backend 미기록 + RecordFactoryRunStart 가 runs 행 아닌 todo 스토어 기록 — M2 "변경 불필요" 주장 거짓, AC-003/008 backend 조항 미달성)·D2(단일 고아 픽스처 — 첫 고아만 반환 변이 탈출)·D3(boot-proof 기저 제외 픽스처 부재 — Basis≠BasisPeer 변이 탈출, D14류 잔존)·D4(AC-002 stdin 단언 공허 — 검증 가능 계기 없음)·D5/D6 선택. 선대 D15/D16/D17 재발 없음, D14 는 본축 폐쇄. FOUNDER Q1-Q3 는 plan 단계 비차단(레지스터 확인)·자율 kickoff 는 운영자 판정까지 차단 — §E.1 기술 정확. 판정 파일 `.moai/reports/t1863/plan-audit-iter1.md`(audited 2178966bb, plan_artifact_hash b60815eb). 문서 수리 manager-spec 재위임(D1-D4 필수·D5/D6 선택).
- 02:5xZ — 수리 완료(SPEC 0.2.0: M2 AdoptRun UPDATE 확장 SET lead_pid/process_start/backend, A2 2-고아·H boot-proof 픽스처, counting-reader 계기, row 12 -cr, 인용 정정 — 6/6 처방 형태). **iter2(델타): FAIL 0.94 — D1-D6 전부 폐쇄 확인**(인용 전건 정확·계수 유지·lint 클린), 신규 2건: D7(블로커: 단일 고아 픽스처에서 resume 인수가 고아를 소진 — strictResumeCarrierSeen 생략 변이 통과 + 변이 문장 허위)·D8(선택: 양측 glm 씨드로 backend 등식 공허). 잔여 1회(3/3 상한 — t1849 와 동일 벼랑). 수리 2차 재위임(D7 2-고아 resume 행·D8 씨드 교차).
- 03:0x-06:3xZ — 수리 2차 완료(SPEC 0.3.0: AC-006 rows 13-14 — 2-고아 픽스처로 마스킹 폐쇄 + 동일 계열 --factory-run 마스킹 선제 폐쇄, AC-003/008 씨드 교차 + per-launcher 생략 변이 명명, plan 미수정). 감사자 5시간 한도 사망→재설정 경과 후 재개(판독 완료 국면). 임대 만료·수거·재임대 1회.
- 06:38Z — **plan-audit iter3: PASS 1.00**(audited 2178966bb, plan_artifact_hash a9f3f549…, convergence pass, 필수 codex 게이트 pass rcpt-f983500100cd7824ab8932f4, blocking 0): D7/D8 코드 근거 폐쇄 확인(마스킹 기전·FindResumableRun 2후보·AdoptRun WHERE·leaderRunIDWasMinted 순수성), 신규 결함 0, plan·decision-index 바이트 불변 검증. **plan 단계 완료.**
- 06:40Z — SPEC 커밋 `47a3d8a7c0`(문서 5종). **kickoff 전 정지 — 리더 통지**(spec_id 보정은 운영자 결정 대기, 보정 없는 kickoff 시도 금지 — 룰링 10-18). FOUNDER Q1-Q3 미판정 상태 유지(자율 kickoff 차단 — decision-index 기재대로).
- 06:5xZ — 영수증 보완(감사자 자발): 중단 전 영수증(rcpt-f9835001…)의 트리 정합성 문제를 스스로 발견 → 신규 codex 재실행으로 **동일 판정 확인**, `rcpt-4248f383fd6e70c84af62ee1` 로 판정 파일 전면 재인용(선행 실행분은 동일 형태 기록으로 보존), 수렴 재검증 ok/unmet []. 판정 불변(PASS 1.00·blocking 0).
- 07:1xZ — **운영자 결정 접수(보드 d-20261018T070800Z-52cc, 리더 전달)**: FOUNDER Q1=거부·안내(run 미생성) / Q2=모든 명시 --factory-run 리더 시작에 인수 기록 경로 통일 / Q3=chooser 카드 착지 전까지 터미널 무관 거부 — 셋 다 작성본 그대로. decision-index 3개 판정 칸 기입 + 보드 인용(커밋 5e6c58daba). **해시 불변 검증**: spec+plan+acceptance 연접 SHA-256 = a9f3f549…(편집 전후 동일 — decision-index 는 해시 대상 외). **카드 spec_id 보정 완료**(리더, 운영자 승인 — SPEC-FACTORY-LEADER-ADOPT-CORE-001, 백업 db + card.fields 이벤트).
- 07:1xZ — **T5 증거 매듭 실측**(plan 단계 변주): T5 진입이 감사 시점 트리(2178966bb)를 요구하나 판정 파일이 그 커밋에 미추적 → "artifact absent at commit" 거부 실측. t1828 T11/T13 전례의 정당한 풀림 적용: 감사자가 동일 판정을 판정 파일 추적 커밋(5e6c58daba) 기준으로 재앵커(내용 불변·해시 결속 동일, 재감사 아님) → 판정 파일 커밋 → T5(5e6c58daba:iter3) → T7 → T8a. 재발행 요청 발송, 대기.
- 07:2x-07:3xZ — T5 착지(v8, 증거 5e6c58daba:plan.md — 감사 앵커 커밋에 추적된 plan 산출물). **T7 천장 거부 실측**: 라운드 3 ≥ Tier M 상한 2 → Rung 3 hold. 원인 추적(코드 완독): Rung 0 pass-through 암(REQ-ACE-013, admission-clean 이면 천장 무관 승인)이 실패 — 파싱 필드 전부 정상, 실패 요인은 **hashOK=false**: 시임이 `ComputeHash(specDir)`(런타임 조성 — 파일 순서 [acceptance, decision-index, design, plan, research, spec, tasks], 각 `이름:길이\0`+바이트+\0, decision-index 포함)와 판정 기재 해시를 비교하는데 감사자가 단순 연접 SHA(a9f3f549)를 기재. 기계 조성 산출: `a0f8cb8e…`(판정 기입 반영). 감사자 교정 요청(양쪽 해시 표기 동시 갱신 — 미갱신 시 duplicate-keys). 교정 도착 시 T7 재시도.
- 07:4x-07:5xZ — T7 매듭 완전 해소 3단: (1) 해시 조성 교정(감사자, 게이트 실함수로 독립 검증 동일값) → 재시도 여전히 hold → (2) 모듈 내 임시 프로브로 승인 술부 직접 실행 → 실패 사유 확정: **`duplicated decision key(s): plan_artifact_hash`** — 6행 프로즈 편집이 구성 설명을 키 줄에 부착해 값 불일치 → (3) 6행 값 정화 교정. **T7 kickoff 착지(07:37:24, 증거 PASS 판정·audited 5e6c58daba)** — kickoff 는 비임대 상태(lease-holding 목록 외, holder 공실 정상). 임대 만료 1회 자연 경과(수거 불요 — kickoff 무임대 설계).
- 07:5xZ — **T8a 착지: `t1863: run (v10)`** — `moai factory decide --gate kickoff --choice approve --decider audit`(자율 감사 심의). 첫 시도 거부 → 원인: §E.1 의 bare `audit_ready: true` 신호 행 누락(저작 시 plan_status 만 기재; card_audit_kickoff.go:47-94 판독부 확인) → 신호 행 추가(소유권: §E.1 기계 신호 한 줄 기계적 보완) → 승인. §F 모드 선택 로그 기록(serial, Tier M) → run 진입 문서 커밋 `6400ae9d4e`(판정 3종+§G+§F). manager-develop 스폰(M1→M4 직렬, RED 먼저, Section A-E 템플릿).

## 3. 산출물

- SPEC 디렉터리: `.moai/specs/SPEC-FACTORY-LEADER-ADOPT-CORE-001/` (manager-spec 스폰 예정)
- plan-audit: `.moai/reports/t1863/plan-audit-iter*.md` (예정)

## §E.2 Run-phase Evidence

Tree: `.moai/worktrees/t1863` (branch `WT-10-18-t1849`). Every command below was run in this tree with the environment scrubbed (`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && TMPDIR=/private/tmp/moai-lda-verify`) and `-count=1`, one plain `-run` pattern per invocation.

### M1 — Orphan finder (AC-FLAC-001) — commit `<M1-SHA>`

RED (pre-fix, verbatim; the finder does not exist yet, so the three named tests cannot compile — the E1 swept set is 0/3):

```text
$ go test ./internal/homestate -run 'TestFindOrphanRunsRequiresDeadStampAndCards' -count=1 -v
# github.com/modu-ai/moai-adk/internal/homestate [...]
internal/homestate/factory_run_orphan_test.go:183:21: db.FindOrphanRuns undefined (type *FactoryDB has no field or method FindOrphanRuns)
internal/homestate/factory_run_orphan_test.go:211:21: db.FindOrphanRuns undefined (type *FactoryDB has no field or method FindOrphanRuns)
internal/homestate/factory_run_orphan_test.go:248:21: db.FindOrphanRuns undefined (type *FactoryDB has no field or method FindOrphanRuns)
FAIL	github.com/modu-ai/moai-adk/internal/homestate [build failed]
FAIL
```

GREEN (verbatim tails, post-fix):

```text
$ go test ./internal/homestate -run 'TestFindOrphanRunsRequiresDeadStampAndCards' -count=1 -v
--- PASS: TestFindOrphanRunsRequiresDeadStampAndCards (0.47s)
ok  	github.com/modu-ai/moai-adk/internal/homestate	0.861s
$ go test ./internal/homestate -run 'TestFindOrphanRunsExcludesNonQualifyingRuns' -count=1 -v
--- PASS: TestFindOrphanRunsExcludesNonQualifyingRuns (0.45s)
ok  	github.com/modu-ai/moai-adk/internal/homestate	0.670s
$ go test ./internal/homestate -run 'TestFindOrphanRunsCountsNonTerminalCardsOnly' -count=1 -v
--- PASS: TestFindOrphanRunsCountsNonTerminalCardsOnly (0.55s)
ok  	github.com/modu-ai/moai-adk/internal/homestate	0.812s
```

E4 regression guards stay green after M1 (9/9, each `ok`): TestAdoptRunRefusesChangedStamp, TestLiveRecordsTerminalSetMatchesSharedDefinition, TestReconcilePreservesDeadOwnerRunWithLiveRecords, TestReconcilePreservesDeadOwnerWithOnlyWorkers, TestReconcileRetiresDeadOwnerWithTerminalRecordsOnly, TestFindResumableRunAdoptsSingleDeadOwnerWithLiveRecords, TestFindResumableRunSkipsLiveOwner, TestFindResumableRunSkipsRunWithoutLiveRecords, TestFindResumableRunReportsMultipleCandidates.

`gofmt -l` on the two touched files: clean. spec.md frontmatter `draft → in-progress` recorded on this milestone's commit (Status Transition Ownership Matrix, manager-develop).
