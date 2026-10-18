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

### Final AC matrix (all re-observed on the final tree, HEAD 997b85e933, this run)

| AC | Status | Command (one plain -run each) | Verbatim tail (this run) |
|---|---|---|---|
| AC-FLAC-001 | PASS | `go test ./internal/homestate -run '^TestFindOrphanRunsRequiresDeadStampAndCards$' -count=1 -v` (and the two sibling names) | `--- PASS ... (0.50s)` / `ok ... 0.853s`; ExcludesNonQualifying `ok 1.142s`; CountsNonTerminal `ok 1.476s` |
| AC-FLAC-002 | PASS | `go test ./internal/cli -run '^TestLeaderStartOrphanRefusalPrecedesRunWrite$' -count=1 -v` | `--- PASS ... (1.54s)` / `ok ... 2.967s` |
| AC-FLAC-003 | PASS | `go test ./internal/cli -run '^TestExplicitFactoryRunAdoptTakesOverAtomically$' -count=1 -v` | `--- PASS ... (13.27s)` with both subtests PASS / `ok ... 14.884s` |
| AC-FLAC-004 | PASS | `go test ./internal/cli -run '^TestExplicitFactoryRunAdoptRefusals$' -count=1 -v` | `--- PASS ... (10.08s)` with all three subtests PASS / `ok ... 11.518s` |
| AC-FLAC-005 | PASS | `go test ./internal/cli -run '^TestLeaderStartOrphanGuidanceRendersFacts$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 6.985s` (final re-run post lint fix) |
| AC-FLAC-006 | PASS | `go test ./internal/cli -run '^TestLeaderStartOrphanGating$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 21.342s` (14/14 rows) |
| AC-FLAC-007 | PASS | `go test ./internal/cli -run '^TestLeaderStartOrphanTouchesNoOtherRun$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 11.105s` (final re-run post lint fix) |
| AC-FLAC-008 | PASS | `go test ./internal/cli -run '^TestLeaderStartOrphanGLMMirrorsCC$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 6.395s` |
| AC-FLAC-009 | PASS | `go test ./internal/cli -run '^TestLeaderStartOrphanTimesRenderInKST$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.216s` |
| AC-FLAC-010 | PASS | `go test ./internal/cli -run '^TestLeaderAdoptGuidanceDocsMatchAnchors$' -count=1` + `^TestFactorySkillAssertionsMatchBehavior$` | `ok ... 1.137s` and `ok ... 13.258s` |

E3 regression guards on the final tree (10/10 `ok`, one plain -run per name): TestAdoptResumedLeaderRunAdoptsSingleCandidate 3.219s, SkipsWhenNotResumeShaped 1.915s, WarnsOnMultipleCandidates 2.393s, MintsWhenNoCandidates 2.107s, WarnsCodexLedCandidate 2.431s, TestStrictResumeCarrierSeen 1.454s, TestLeaderRunIDWasMinted 1.143s, TestFactoryLeadRefusesCodexLedRun 3.221s, TestFactoryJoinAndLeadAcceptNonCodexRun 4.973s, TestFactorySkillAssertionsMatchBehavior 13.178s. E4 homestate guards likewise (`ok` each; TestAdoptRun* with the mechanically extended signature, TestRetireRunIfDead*, TestFindResumableRun*, TestReconcile*, TestLiveRecordsTerminalSetMatchesSharedDefinition).

### E8 — the four README sentences after M4 (banned substrings absent, grep 0 each)

- **README.md (en)**: "A run whose leader has died is retired by the next lane join only when it holds no in-progress cards and no registered workers, so a finished run no longer stalls the join on `AMBIGUOUS_FACTORY`, while a run still carrying cards or workers stays active. A leader start without a resume option over such a run is refused with guidance and creates no run; adopt it deliberately with `--factory-run <run-id>`."
- **README.ko.md**: "리더가 죽은 런은 진행 중인 카드와 등록된 워커가 없을 때만 다음 레인 합류에서 회수되므로 끝난 런이 합류를 `AMBIGUOUS_FACTORY`에서 막지 않고, 카드나 워커가 남은 런은 활성으로 남는다. 이런 런 위에서 이어받기 옵션 없이 리더를 띄우면 안내와 함께 거부되고 어떤 런도 만들어지지 않는다. `--factory-run <run-id>`로 의도적으로 이어받는다."
- **README.ja.md**: "リーダーが死んだ run は、進行中のカードと登録済みワーカーを抱えていない場合に限り、次のレーン合流で retire される。そのため終わった run が合流を `AMBIGUOUS_FACTORY` で止めることはなく、カードやワーカーが残る run は有効のまま残る。そのような run の上で resume オプションなしにリーダーを起動すると、ガイド付きで拒否され、run は何も作られない。意図的に引き継ぐには `--factory-run <run-id>` を使う。"
- **README.zh.md**: "主导已经死掉的 run，只有在不持有进行中的卡片和已注册的泳道时，才会在下一条泳道加入时退役；因此结束的 run 不会让加入卡在 `AMBIGUOUS_FACTORY` 上，仍持有卡片或泳道的 run 保持活动状态。在这样的 run 上不带恢复选项启动主导会被拒绝并给出指引，且不会创建任何 run；用 `--factory-run <run-id>` 来有意接管。"

All four carry both anchor spans; the banned substrings count 0 (`grep -c` → 0, 0, 0, 0).

### Gaps and Residual-risk (run phase)

- **Lint baseline**: the pre-card lint baseline could not be measured in pre-flight (two attempts hit `parallel golangci-lint is running` — another session held the lock). The comparison anchor used instead: `git show 6400ae9d4e:internal/cli/factory_run_adopt.go` proves the 3 remaining errcheck findings exist verbatim at the card base. Final lint: exit 1 with exactly those 3; **0 NEW**.
- **The Hugo four-locale verify recipe** (`hns-oss-docs-verify` full checklist): the warning-free Hugo build (exit 0, `--quiet`), sitemap existence, 4-locale file/heading parity, and body-emoji scan were executed here; the URL-blacklist grep and Mermaid-direction grep were NOT run (no URL or Mermaid content was touched by this SPEC — the diff touches prose paragraphs and one section only). Named as a Gap, not silently skipped.
- **`moai spec lint --strict`**: 0 error(s), 0 warning(s) (INFO: `OwnershipTransitionUnmeasured` — the M1 transition commit carries no `Authored-By-Agent:` trailer; Info severity by schema, not repairable without `--amend`, which is forbidden. Future commits in this card should carry the trailer).
- **Coverage**: the ≥85% package-coverage run was not executed (host load rules; the swept-count rule governs the ACs). Named Gap.
- **The full-suite verdict** belongs to the CI run on the project's integration branch and is PENDING at report time (the lane does not push; the leader owns integration).
- Boundary grep (`AskUserQuestion` in non-test lines of internal/cli + internal/homestate): the only hits are pre-existing doc-comment text in `internal/cli/harness.go` and `internal/cli/pr_watch_cmd.go` (files this card never touched); none of the card's files appear.

### M1 — Orphan finder (AC-FLAC-001) — commit `06efd0b6a4`

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

### M2 — Deliberate adopt through the adoption write (AC-FLAC-003, AC-FLAC-004) — commit `bb668c752a`

RED (pre-fix, verbatim; the seam and the explicit-adopt path do not exist yet — the E2 swept set covers these names at 0/9):

```text
$ go test ./internal/cli -run 'TestExplicitFactoryRunAdoptTakesOverAtomically' -count=1 -v
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/factory_leader_orphan_test.go:312:11: undefined: classifyNamedRun
internal/cli/factory_leader_orphan_test.go:313:3: undefined: classifyNamedRun
internal/cli/factory_leader_orphan_test.go:328:22: undefined: classifyNamedRun
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

GREEN (verbatim, post-fix):

```text
$ go test ./internal/cli -run 'TestExplicitFactoryRunAdoptTakesOverAtomically' -count=1 -v
--- PASS: TestExplicitFactoryRunAdoptTakesOverAtomically (5.81s)
    --- PASS: TestExplicitFactoryRunAdoptTakesOverAtomically/dead-stamp_cross-backend (2.30s)
    --- PASS: TestExplicitFactoryRunAdoptTakesOverAtomically/live-owner (3.51s)
ok  github.com/modu-ai/moai-adk/internal/cli	7.154s
$ go test ./internal/cli -run 'TestExplicitFactoryRunAdoptRefusals' -count=1 -v
--- PASS: TestExplicitFactoryRunAdoptRefusals (8.33s)
    --- PASS: TestExplicitFactoryRunAdoptRefusals/codex-led (2.56s)
    --- PASS: TestExplicitFactoryRunAdoptRefusals/live-legacy-peer (1.44s)
    --- PASS: TestExplicitFactoryRunAdoptRefusals/owner-change-race (4.32s)
ok  github.com/modu-ai/moai-adk/internal/cli	9.570s
```

E3 regression guards stay green after M2 (10/10, each `ok`): TestAdoptResumedLeaderRunAdoptsSingleCandidate, TestAdoptResumedLeaderRunSkipsWhenNotResumeShaped, TestAdoptResumedLeaderRunWarnsOnMultipleCandidates, TestAdoptResumedLeaderRunMintsWhenNoCandidates, TestAdoptResumedLeaderRunWarnsCodexLedCandidate, TestStrictResumeCarrierSeen, TestLeaderRunIDWasMinted, TestFactoryLeadRefusesCodexLedRun, TestFactoryJoinAndLeadAcceptNonCodexRun, TestFactorySkillAssertionsMatchBehavior. The AdoptRun signature's mechanical test-caller updates leave their assertions untouched (TestAdoptRunRestampsOwnerAndAppendsResumedEvent, TestAdoptRunRefusesChangedStamp, TestAdoptRunRefusesNonActiveRun, TestRetireRunIfDead* all `ok`). `GOOS=windows GOARCH=amd64 go build ./...` exit 0. `gofmt -l` on all eight touched files: clean.

### M3 — Refusal gate and guidance block (AC-FLAC-002, 005..009) — commit `e4e8b2a499`

RED (pre-fix, behavioral, captured with the gate temporarily unwired from both call sites — the exact pre-M3 behavior; verbatim tails):

```text
$ go test ./internal/cli -run 'TestLeaderStartOrphanGating' -count=1 -v
--- FAIL: TestLeaderStartOrphanGating (20.96s)
    --- PASS: TestLeaderStartOrphanGating/resume-space-form (2.05s)
    ... (all 12 exempt rows PASS — today's behavior preserved)
    --- FAIL: TestLeaderStartOrphanGating/continue-flag (1.41s)
    --- FAIL: TestLeaderStartOrphanGating/short-cluster-cr (1.16s)
FAIL
$ go test ./internal/cli -run 'TestLeaderStartOrphanGuidanceRendersFacts' -count=1
    factory_leader_orphan_test.go:625: the launch over two orphans succeeded; want the orphan refusal
FAIL
$ go test ./internal/cli -run 'TestLeaderStartOrphanRefusalPrecedesRunWrite' -count=1
    factory_leader_orphan_test.go:583: the launch over an orphan succeeded; want the orphan refusal
FAIL
```

`TestLeaderStartOrphanTimesRenderInKST` drives the pure KST renderer directly and cannot go red against unwired call sites — its RED-now cell is ledger E2 (test absent, 0/9 swept). Two test-helper defects were found and fixed during the RED capture itself (not code defects): `runEntryWithStderr` evaluated `buf.String()` inside the return statement before the call ran (always-empty stderr), and the gating row env was set before `clearFactoryTestEnv` ran (unset again); both were repaired before any GREEN was recorded.

GREEN (verbatim tails, post-wiring):

```text
$ go test ./internal/cli -run 'TestLeaderStartOrphanRefusalPrecedesRunWrite' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	2.729s
$ go test ./internal/cli -run 'TestLeaderStartOrphanGuidanceRendersFacts' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	5.138s
$ go test ./internal/cli -run 'TestLeaderStartOrphanGating' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	24.675s
$ go test ./internal/cli -run 'TestLeaderStartOrphanTouchesNoOtherRun' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	9.262s
$ go test ./internal/cli -run 'TestLeaderStartOrphanGLMMirrorsCC' -count=1 -v
--- PASS: TestLeaderStartOrphanGLMMirrorsCC (10.30s)
    --- PASS: TestLeaderStartOrphanGLMMirrorsCC/takeover-backend (5.44s)
    --- PASS: TestLeaderStartOrphanGLMMirrorsCC/codex-refusal-text-equal (2.54s)
    --- PASS: TestLeaderStartOrphanGLMMirrorsCC/guidance-text-equal (2.33s)
$ go test ./internal/cli -run 'TestLeaderStartOrphanTimesRenderInKST' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	1.242s
```

The mirror subtest ran `-count=3` to retire a fixture flake (unpinned created_at rendered record-time into the block; now pinned): 3/3 `ok`.

E3 regression guards stay green after M3 (10/10, each `ok`, same name list as M2). `GOOS=windows GOARCH=amd64 go build ./...` exit 0. `gofmt -l` on the four touched files: clean.

### M4 — Guidance text (AC-FLAC-010) — commit `84f9387535`

RED (pre-fix, verbatim): `TestLeaderAdoptGuidanceDocsMatchAnchors` ran against the untouched docs — 32 failure lines (20 anchor-presence errors: each of the ten files missing each of the two anchors; plus the banned-substring, heading-parity, and twin checks all failing their absent-anchor premises), `FAIL github.com/modu-ai/moai-adk/internal/cli` — matching ledger E5's 0-occurrence state.

GREEN (verbatim tails):

```text
$ go test ./internal/cli -run 'TestLeaderAdoptGuidanceDocsMatchAnchors' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	1.561s
$ go test ./internal/cli -run '^TestFactorySkillAssertionsMatchBehavior$' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	13.258s
```

E5 flipped (verbatim): `grep -c 'leader-start/orphan' <ten files>` → 1 for each of the ten; `grep -c 'lane-join/dead-owner' <ten files>` → 1 for each of the ten; the four banned README substrings (`is retired automatically when the next lane joins` / `자동으로 회수되므로` / `次にレーンが合流した時点で自動的に retire されるので` / `会在下一条泳道加入时自动退役`) each count 0 in its README. Docs-site level-2 headings 10/10/10/10 (E6 was 9 each, +1 new section, locale-locale equality holds); README level-2 headings 12/12/12/12 unchanged. Twin skill files byte-identical (`cmp` exit 0); `make build` exit 0 with the catalog regen (`internal/template/catalog.yaml` one hash line — the factory.md skill entry), and a second `gen-catalog-hashes --all` produced no further change. Hugo build (`hugo --quiet` in docs-site) exit 0 with sitemap.xml emitted. The four replaced README sentences are quoted in the §E.2 final matrix.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-10-18T09:20:00Z"
run_commit_sha: "997b85e933"
run_status: "implemented"
ac_pass_count: 10
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 4
l44_pre_commit_fetch: "not-run (lane protocol: card worktree, lanes do not push; integration is the leader's)"
l44_post_push_fetch: "not-applicable (no push by the lane)"
new_warnings_or_lints_introduced: 0
lint_baseline_note: "3 pre-existing errcheck findings verified verbatim at card base 6400ae9d4e (factory_run_adopt.go warn lines, git show evidence); final lint exit 1 with only those 3"
cross_platform_build.darwin: "exit 0"
cross_platform_build.windows_amd64: "exit 0"
total_run_phase_files: 19
m1_to_m4_commit_strategy: "one conventional commit per milestone (M1 06efd0b6a4, M2 bb668c752a, M3 e4e8b2a499, M4 84f9387535) + one lint-fix commit 997b85e933; spec.md frontmatter draft->in-progress on the M1 commit"
spec_lint_strict: "0 error(s), 0 warning(s) (one INFO OwnershipTransitionUnmeasured: M1 commit lacks the Authored-By-Agent trailer; Info severity, amend forbidden)"
hugo_build: "exit 0 (docs-site, --quiet), sitemap.xml emitted"
twin_skill_files: "byte-identical (cmp exit 0); make build leaves no further tracked change (catalog regen idempotent)"
e2e_red_discipline: "behavioral RED captured per milestone (M1 compile RED for the new finder, M2 seam-undefined RED, M3 gate-unwired behavioral RED with 12/14 exempt rows green, M4 32-failure RED on untouched docs)"
ci_run_on_integration_branch: "PENDING at report time — the repository-wide test verdict is owned by the CI run on the integration branch, not by this lane"
```

### Sync-audit repair round (F1, F2 blocking; F3, F4) — commit `a3f42e6c89`

Verdict read: `.moai/reports/t1863/sync-audit-cold.md` (FAIL 0.86, two codex-gate P2s). Both premises were re-verified against the code before repair.

**F1 (blocking) — window refusals re-applied.** RED (verbatim, pre-fix): `TestExplicitFactoryRunAdoptReappliesWindowRefusals` both subtests FAILED with "the adopt proceeded over a run whose backend flipped to codex inside the window; want the refusal" (backend-flip) and "...gained a live legacy peer inside the window; want the refusal" (legacy-peer) — the T2 write adopted both window mutations because the stamp CAS still matched. GREEN (verbatim, post-fix): both subtests `--- PASS` / `ok ... 10.713s`. Fix: `adoptExplicitLeaderRun` now re-runs `refuseCodexLeaderRun` and the legacy-peer boundary after the fresh `classifyNamedRun` classification and before the CAS write (immediately-checked preconditions per plan M2's structure). The legacy-peer refusal was extracted into `refuseLiveLegacyPeerRun` (factory.go) so the T1 pre-check and the T2 window re-check share one predicate and one refusal text; `enterSelectedFactoryRun` now calls the same helper (T1 tests `TestEnterSelectedFactoryRun*` and `TestFactoryLeadRefusesCodexLedRun` re-run green).

**F2 (blocking) — cards snapshot discrimination.** RED (verbatim, probe pre-fix): a version bump + evidence_path write left `adoptCardsSnapshot` byte-equal — "snapshot equal despite a card-writing mutation: true". GREEN (verbatim, probe post-fix, probe deleted after): "snapshot equal despite a card-writing mutation: false" with the fields visible (`c-1||run|3|.moai/reports/probe-cards/card-c-1.md|...` → `c-1||run|4|.moai/reports/t1863/verdict.md|...`). Fix: the snapshot string now includes version and evidence_path, and `seedAdoptEntryRun` seeds non-default values (version 3, per-card evidence path) so a card-writing mutant diverges on the discriminating fields.

**F3 — dead close-error**: `adoptExplicitLeaderRun` results are now named `(adopted string, err error)` and the write assigns the named `err`, so `closeFactoryInto` no longer writes into a dead local.

**F4 — README anchor placement**: the anchor spans moved from inline mid-paragraph to their own trailing line in all four READMEs. The repair probe surfaced a real M4 defect beyond cosmetics: in the ja/zh READMEs the M4 edit had left the anchors mid-paragraph with the rest of the original paragraph running on after them; both copies now sit exactly once, on their own line after the full paragraph (`grep -o` counts 1/1/1/1 across the four locales). The first placement script silently no-op'd on ja/zh (full-width sentence period before the span, not a space) and a second script asserted presence instead of the space-prefix and briefly doubled the spans — caught by `grep -c` (2) and repaired in the same round; the lesson is recorded here, not in code.

Scoped verification (verbatim tails): AC-003 `ok 13.650s` (both legs), AC-004 `ok 12.999s` (3 variants), AC-007 `ok 9.750s`, gating `ok 33.653s` (14/14 rows), anchors test `ok 1.637s` post-F4, T1 regressions `TestEnterSelectedFactoryRun*` + `TestFactoryLeadRefusesCodexLedRun` each `ok`. gofmt clean; `GOOS=windows` build exit 0; lint exit 1 with only the 3 pre-existing baseline findings (0 NEW).

### Sync-audit round 3 repair (F1' residual) — commit `727f9e33f`

Verdict read: `.moai/reports/t1863/sync-audit-cold.md` round 2 section (FAIL 0.90 converging; F2/F3/F4 closed, F1' residual blocking).

**Behavioral RED (verbatim, scratch probe on the pre-fix tree, probe deleted after)**: a backend flip to `codex` landing after the classified read, then `AdoptRun` (6-arg, stamp-only CAS) — `adopt err=<nil>`, `row after adopt: backend="claude" pid=4242 run.resumed events=1` — the flip was silently overwritten and the run adopted. Compile RED for the final-form tests (RunOwner.Backend undefined, AdoptRun arity) captured verbatim before the implementation.

**Fix (the auditor's prescribed shape):**
1. **Backend axis → ATOMIC in the CAS**: `RunOwner` gains `Backend` (the classified row's `lead_backend`; `classifyRuns`' SELECT extended), and `AdoptRun`'s UPDATE WHERE gains `AND lead_backend=?` against the classified value — a flip landing after the classified read misses the write and rolls back, exactly like a stamp change. Refusal message extended to "no longer carries the classified owner stamp and backend" (the round-2 owner-change assertion's substring still matches).
2. **Legacy axis → peer guard INSIDE the transaction**: `AdoptRun` gains a `peerGuard func(context.Context) error` parameter, invoked after the write and before the commit — a refusal rolls the write back. It is caller-supplied because the broker database cannot join the homestate transaction (the factorymsg import direction is fixed); this is the narrowest window a cross-database check allows.
3. **Post-guard regression tests**: `TestAdoptRunCASRefusesPostClassificationBackendFlip` (flip after the classified read → refused, competing writer's row survives, no event) and `TestAdoptRunPeerGuardRefusesInsideWindow` (guard refuses inside the tx → write rolled back, stamp unchanged, no event).

**Judgment calls**: the backend condition lives in the UPDATE's WHERE clause (not a re-classify-and-compare inside the tx) because the WHERE is evaluated against the committed row at UPDATE execution time — SQLite serializes writes, so any flip committed before the UPDATE makes it miss; a second classification inside the tx would re-open the read-then-write gap it is meant to close. The resume path passes `nil` for the guard: its window contract is the warn-and-mint outcome E3 pins, and an in-write abort would change that contract; its backend axis is covered by the shared extended CAS regardless.

**GREEN (verbatim)**: `--- PASS: TestAdoptRunCASRefusesPostClassificationBackendFlip (0.31s)` / `--- PASS: TestAdoptRunPeerGuardRefusesInsideWindow (0.24s)`. Scoped verification: AC-003 `ok 8.832s` (both legs), AC-004 `ok 9.257s` (3 variants), round-2 window refusals `ok 8.762s`, T1 regressions `TestFactoryLeadRefusesCodexLedRun ok 4.405s` / `TestEnterSelectedFactoryRunRefusesLiveLegacyPeer ok 2.611s`, resume baselines (AdoptsSingleCandidate 3.458s / WarnsOnMultipleCandidates 1.889s / WarnsCodexLedCandidate 2.456s), homestate adopt baselines (RefusesChangedStamp 0.744s / RestampsOwner 0.539s / RefusesNonActiveRun 0.650s / RetireRunIfDeadCardReadHoldsWriteLock 0.574s). gofmt clean, `GOOS=windows` exit 0, lint exit 1 with only the 3 pre-existing baseline findings (0 NEW).
