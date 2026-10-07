# progress.md — SPEC-GATE-BOTTLENECK-001 (card t1575)

## §E.1 Plan-phase Audit-Ready Signal
plan_status: audit-ready (self — lane-authored SPEC, card t1575의 4종 계약 그대로)
plan_complete_at: 2026-10-07

## §F Phase 4 Mode Selection
Input: tier M, scope 3 files (gate + cache + tests), 1 domain (cli), Go.
Decision: direct (orchestrator/lane-direct, no Agent spawn) — 단일 파일 중심 수리형 구현으로 serial 대용 direct가 정합.

## §E.2 Run-phase Evidence

- M1 (cache) 착지: `internal/cli/codex_review_cache.go`(신설) + `codex_review_gate.go` (4a) 조회·(4b) 기록 삽입.
  - 재판정: `go test ./internal/cli/ -run 'TestReviewGate' -count=1` → ok (7.1s)
  - 신규 2종: TestReviewGate_CacheHitReusesFailVerdict(캐시 히트 → BLOCK·RPC 0회·skip 카운트)·TestReviewGate_StaleKeyRunsLiveReview(키 이동 → 라이브 리뷰 전환 — 실패 픽스처로 판별력 강화)
  - fixture 교훈: receipt 기록이 porcelain을 더럽히면 키가 이동 — fixture가 실repo의 `.moai/state/` ignore를 반영해야 함(실측).
- M3(스코프 분리): 계약 문서화 완료(SPEC § Scope split — 카드 diff 축은 t1383 소관).
- M4(재현 조건부): 캐시 히트 경로 = 전체 재현 생략(REQ-GBN-003, 구현됨). 라이브 리뷰의 0건 조건부는 review request 파라미터 확장이 필요 — M2와 함께 잔여.

## §E.3 Run-phase Audit-Ready Signal
run_status: partial — M1 착지, M2(지연 블록)·라이브 리뷰 0건 조건부 잔여.

## 체크포인트 (다음 세션 재개 지점)

- M2 지연 블록: Stop에서 수신 영수증 부재 시 백그라운드 리뷰 기동(`moai verify codex-review` 재사용) + ALLOW, 판정은 다음 턴 진입 훅(UserPromptSubmit/PreToolUse 신설 배선 — settings.json.tmpl + 훅 매니페스트 M4)에서 receipt 조회해 집행. async:true 불채택(블록 능력 상실).
- 라이브 리뷰 0건 조건부: reviewRequestParams에 재현 지시자 추가(codex 측 지원 확인 선행).
- 검증 남은 것: 턴당 벽시간 전후 비교 실측, no-edit 자체허용 회귀 없음(기존 TestReviewGate_NoEditTurnAllows가 계속 green으로 보호 중), fail-open 불변(신규 테스트 2종이 담보).

## 턴종료 게이트 발견 처분 (r9 — 본 카드 diff 발견 0건)

턴종료 게이트가 merge-ref 전체 diff를 보며 보고한 14건(patch-id P1=t1561 발행·after 힌트 재계산·bundle 3건·picked 스킵·todo_issuance 6건·todo.go 2건·backlog_store/relation)은 **전부 t1542 card-review 원장과 리더 원장(t1533 라인·t1561·t1562·t1559)에 기재된 기존 행의 재관측**이다. 본 M1 diff(codex_review_cache.go·gate 조회/기록·테스트)에서의 발견은 0건. 본 카드는 이들 수리의 소관이 아니며 원장 행이 소유한다.

## 턴종료 게이트 발견 처분 (r10 — 본 카드 diff 2건 수리, 타 소관 11건 재관측)

- **본 소관 2건 수리(커밋 대상)**: ① fail receipt에 ExitCode=1 기록(기존 produceCodexReviewReceipt와 동일 매핑 — pass형 0이 다른 소비자에게 성공 증거로 읽히는 결함) ② 캐시 차단 메시지가 "무엇을 고칠지"를 잃는 문제 — fail 시 요약·발견을 `.moai/state/verify/codex-review/<head>-<digest>.md`(트리키 동일·런타임 관리 영역)에 보존하고 캐시 차단 이유에 첨부. 판별 테스트 TestReviewGate_LiveFailPreservesDetailForCachedBlock 추가. `go test ./internal/cli/ -run 'TestReviewGate' -count=1` → ok (24.0s).
- **타 소관 11건 재관측**: protected_zone_shell(1·t1500/1510 계열)·todo_issuance(4·t1559)·factory_bundle(3·t1561/62)·todo.go(1·t1554)·backlog_relation/store(2·t1542 card-review 원장). 전부 r9 원장 행의 재관측 — 본 카드 수리 소관 아님.

## 턴종료 게이트 발견 처분 (r11 — 본 카드 diff 발견 0건 연속)

r10 수리 판정 통과(리뷰어 명시: 캐시 회귀 테스트 3개 통과). 13건 전부 기존 원장 행의 재관측 — factory_card 3·factory_bundle 3·todo_issuance 4·todo.go 2·backlog_store/relation 2의 계열 분포는 r9/r10과 동일. 소관 카드(t1542 원장·t1561·t1562·t1559·t1454·t1554)가 소유하며 본 카드 수리 소관 아님.

## CI 적색 2건 수리 (리더 지시 2026-10-07 — M1 PR #1793)

- **① TestStopChainEffectParityGolden /codex_review_gate 2 leg** — REQ-GBN-001이 바꾼 공유 저장소 계약의 갱신: Claude 게이트가 라이브 판정을 기록하므로(의도된 M1 동작) "Claude 통과 뒤 영수증 부재 → unmeasured" 골든이 더 이상 성립하지 않는다. 두 leg(codex installed no receipt·stale receipt HEAD moved)를 gate_failed 기대로 갱신하고, 영수증 기록 자체를 전제 단언(receiptForCurrentState 헬퍼 — 기록이 말없이 빠지면 leg가 공허해지는 것을 막음)으로 고정. 판정 일치(Deny=Deny)는 유지 — parity 위반 아닌 관측 클래스 갱신. unmeasured 행위 커버리지는 cap 서브테스트가 유지.
- **② TestCheckProtectedZonePosixBackslashConvertedAbsoluteness** — origin/main(5e5ff8e31, t1570 #1797) 흡수 후 green. 코드 변경 0건.
- 흡수: origin/main → 병합 HEAD 64a13b837. 재판정: `go test ./internal/cli/ -run 'TestStopChainEffectParityGolden' -count=1` → ok (40.9s) · `-run 'TestReviewGate|TestCheckProtectedZone'` → cli ok (7.0s)·hook ok (5.0s).

## 최종 판정 (2026-10-07 — PR #1793 @ 92897ca84 전 체크 초록)

- **CI**: 29 pass·0 fail·0 pending — Test (ubuntu-latest) 12m37s·Race Test 1/2 (17m17s/18m40s)·Build×5·Lint·Constitution·Integration×3·spec-lint·spec-status-sync·graph-freshness-conflict-guard·CodeRabbit 포함. CodeRabbit 판독은 리더 몫(AH §4 병합 판정).
- **CI 수리 이력**: 골든 2 leg 계약 갱신(7986851b9) → LiveAxis 오탐 제거(1a3bf99a7) → inconclusive exit 2(92897ca84, 쌍둥이 포함).
- **게이트 라운드 원장 (r12~r17)**: 본 카드 diff 발견 0건 6라운드 연속. 전부 기존 원장 계열의 재관측/신규 인스턴스 — integration 병합 창 재검증 4건·factory_card/bundle 선행의존성 5건·todo_issuance 4건·backlog relation/store 2건·todo.go 2건·landing_predicate 공백 1건·audit_receipt_guard 재시작 표식 1건·todo_auto_lane 1건·integration remeasure 2건. 소관: t1538/t1542 원장·t1561·t1562·t1559·t1454·t1554·t1509.
- **잔여(M2/M3, 체크포인트)**: → 아래 M2 세션 기록으로 이어짐.

## M2 세션 (2026-10-07, lane-6 신세션 — 체크포인트 continuation)

- **전제 이행**: origin/main 흡수 병합 `c6cbfa60e`(M1 착지 PR #1793 = d5fe44c42 포함, progress.md add/add 충돌은 ours-superset으로 해소 — ours가 유일 신규 7줄 추가, theirs 소실 0건 실측). plan.md 병합 M2(재현 계약)/M3(지연 블록) 번호와 체크포인트 호칭(M2=지연 블록) 엇갈림은 남은 작업 집합이 동일하여 체크포인트 명칭으로 진행.
- **§F 모드 재확인**: 기록된 direct(lane-direct) 유지 — 체크포인트 설계가 완결된 continuation이므로 재위임 없이 직접 구현.
- **REQ-GBN-002 구현 (지연 블록)**: `codex_review_gate.go` — 캐시 미스 arm을 라이브 RPC에서 `codexReviewBackgroundKick`(= `os.Executable() verify codex-review --project-root <scope>`, detached, 로그 `.moai/logs/codex-review-bg.log`) 기동 + ALLOW로 교체. 기동 실패도 ALLOW(AC-GBN-004). 캐시 히트 블록 경로 불변. `recordCodexReviewReceipt`는 호출자 소실로 삭제 — 처분 논리는 프로듀서가 단일 소유(7-pre 재분류 포함, fail 상세 보존을 `produceCodexReviewReceipt`로 이동 — r10 수리의 "무엇을 고칠지" 유지).
- **진입 훅 신설**: `codex_review_delay.go` — `HandleCodexReviewEntry`(신선 FAIL → 차단+보존 상세, 그 외 전부 ALLOW/fail-open) + `runCodexReviewEntry`. 래퍼 `handle-codex-review-entry.sh`(repo+template 쌍생, 동일 awk 셀프게이트 — 꺼짐 시 0비용). settings.json.tmpl + 리포 settings UserPromptSubmit 등록(timeout 5). 등록 가드 테스트 확장(TestReviewEntryRegisteredIn{Template,Repo}Settings).
- **parity 골든 갱신**: "no receipt"/"stale receipt" 2 leg를 REQ-GBN-002의 선언된 §D3.4 편차 leg로 갱신(Claude=ALLOW+kick 1회, Codex=unmeasured continuation 불변) — M1이 세운 "Claude가 라이브 판정 기록" 전제는 지연 블록으로 대체. "FAIL receipt" leg는 produce-first 순서로 수정(구 순서는 라이브 경로 전제).
- **테스트 이전 원장**: 라이브 경로 소멸로 요청-형태·재분류·failed-turn 계약을 프로듀서 수준으로 이전 — gate 6종 삭제/재작성·cache 3종 재작성·ownership/gatePath 5곳·wiring(BlockVerdict receipt 기반, HandlerError 은퇴)·scope 6종·live(보안 단정 producer+진입 차단)·rpc_error(producer inconclusive/2)·selfreview(기준 요청=producer). gatePath에 kick 시임 보호 내장.
- **REQ-GBN-003 구현 (조건부 재현)**: `codexReviewReproductionNote` — `codexReviewSessionParams`가 review/start의 `developerInstructions`에 형식 핀 뒤에 부기. codex 측 지원 근거: developerInstructions는 오늘도 산 채로 소비되는 기존 프로토콜 필드(형식 핀이 같은 경로로 실림, mcp_codex.go codexReviewSessionParams) — 새 파라미터 키가 아닌 기존 문자열 확장이므로 지원이 구조적으로 확정. 캐시 히트 스킵(REQ-GBN-001)·0건 판정 수용 모두 요청 계약에 명시.
- **테스트 실측 (family3, 이 트리 커밋 전 기준)**: `go test ./internal/cli/ -run '<게이트 패밀리 14종 필터>' -count=1 -timeout 25m` → **ok (751.2s)** — 라이브 codex 리뷰(실제 fail receipt → 진입 훅 차단까지 end-to-end)·parity 골든·지연 블록·진입 훅·재분류 전 부속 포함 전부 초록. 라이브 픽스처에 `.moai/` gitignore 추가(미포함 시 receipt 쓰기가 키를 이동 — cacheTestRoot 교훈 클래스 재현).
- **턴종료 게이트 발견 처분 (M2 세션 r1 — 본 카드 diff 2건 수리, 타 소관 1건)**:
  - **본 소관 ① (P1) embed manifest 누락**: 새 진입 래퍼가 untracked라 allowlist("git-tracked file set"에서 생성, embedemit.go)에 미포함 → `git add` 후 `make embed-manifest`로 수리, grep 1행 관측. 수리 전 상태에서는 `moai init/update` 배포물에 래퍼 부재 → hook-missing 폴백이 조용히 무시(진입 집행 사망).
  - **본 소관 ② (P2) 중복 kick**: 동일 트리 상태에서 연속 Stop 시 백그라운드 리뷰 중복 기동 — 트리 키별 진행 마커(`kickInFlight`, `.moai/state/verify/codex-review/<head>-<digest>.kick`, 리뷰 예산 만료 시 사판정 재기동, 기동 실패 시 마커 제거로 재시도 허용) 구현 + 테스트 3종(InFlightKickNotRepeated·StaleKickMarkerRekicks·FailedKickIsRetryable). 마커 계열 테스트 `-run 'TestReviewGate_|TestReviewEntry_'` → ok (25.6s).
  - **타 소관 1건 (P2)**: `internal/hook/quality/gate_node_pm.go` bun `x`에 `--bun` 미부착(Node shebang 따라 Bun 전용 환경 실패, gate_typecheck.go 동일) — 본 카드 diff 아님(`git diff --name-only | grep -c gate_node_pm` = 0 실측). t1572(bun runtime exec, 6c1bd84cd) 계통의 신규 인스턴스 — 원장 행 소관, 본 카드 수리 소관 아님.
- **턴종료 게이트 발견 처분 (M2 세션 r2 — 본 카드 diff 2건 수리)**:
  - **(P2) hook 서브커맨드 개수 원장**: `codex-review-entry` 추가로 TestHookCmd_SubcommandCount·TestHookCmd_PrePushSubcommandCount 기대 45→46(원장 행 추가), TestHookValidEventTypes_AllHaveSubcommands의 도메인 훅 유틸 목록에 진입 훅 등록.
  - **(P2) 템플릿 내부 추적 ID 누출**: 배포 템플릿 래퍼 주석의 SPEC-GATE-BOTTLENECK-001·REQ-GBN-002·AC-GBN-004를 일반 동작 서술로 교체(TestTemplateNoInternalContentLeak — 템플릿 무결성 규칙, AGENTS.local.md §내부 ID 금지). 리포 로컬 쌍은 추적 ID 유지(템플릿 검사 대상 아님). 수리 후 TestTemplateNoInternalContentLeak·TestReviewGate*·TestReviewEntry* → ok (5.8s), 개수 원장 3종 → ok (2.6s).
- **벽시간 실측 (AC-GBN-003 + 체크포인트 전후 비교 — bin/moai 실측, 2026-10-07)**:
  - 측정 설계: /tmp 픽스처 리포(uncommitted 소스 변경 1건, gate enabled + primary_scope: review)에 `bin/moai hook codex-review-gate`·`hook codex-review-entry`를 실제 페이로드로 실행. 리뷰 엔진은 실제 codex(shim, codex-cli 0.160.1).
  - **arm1 kick 경로 Stop: 2,526ms** — "background review started" 관측. moai 콜드 스타트가 지배(기존 동기 경로의 리뷰 대기 시간이 Stop에서 소거됨).
  - **arm2 cache-hit Stop: 4,567ms** — "reusing the cached pass verdict for the unchanged tree (skip #1)" 관측.
  - **arm3 진입 훅: 4,740ms** — "background verdict pass for the unchanged tree — allowing" 관측.
  - 배경 리뷰 자체는 백그라운드에서 완주(receipt 스냅숏 기록 관측, verdict pass) — 어느 턴도 리뷰를 대기하지 않음. 비교 기준: 변경 전 동기 리뷰는 리더 실태 조사 기준 11/12 Stop 차단·누적 14m13s.
  - 부수 관측(2차 측정): P2 중복 방지가 실환 동작 — "a background review for this tree state is already in flight; not re-kicking". 2차의 arm2/3이 조기 실행된 것은 측정 스크립트 폴링이 `.kick` 마커를 receipt로 오독한 스크립트 결함(3차에서 snapshots/*.json으로 수정) — 게이트 코드 결함 아님.
- **턴종료 게이트 발견 처분 (M2 세션 r3 — 커밋된 카드 diff 리뷰, 본 소관 2건 수리)**:
  - **(P1) 상태 경로 git 루트 앵커**: 하위 디렉터리 세션의 scope.Dir로는 `.moai/state`가 하위 디렉터리 안에 생겨(템플릿 .gitignore 미커버) receipt 쓰기 자체가 트리 키를 이동 → 진입 훅이 판정을 못 읽음. 게이트·진입 훅·프로듀서 3곳 모두 tree 클래스 scope.Dir를 `reviewExclusionRoot`(git toplevel, 실패 시 원값)로 앵커 — 실행·저장·조회 동일 루트. 회귀 테스트 TestReviewEntry_SubdirSessionReadsRootReceipt 추가.
  - **(P2) 마커 배타적 획득**: stat-then-write 경합으로 겹치는 Stop 둘이 모두 기동권을 얻을 수 있음 — `O_CREATE|O_EXCL` 원자 획득으로 교체(신규=기동권, 기존+신선=in-flight, 기존+만료=제거 후 1회 재경합, 패배 시 in-flight). 회귀 테스트 TestKickInFlight_ExclusiveAcquisition 추가.
  - 부수 수리: 앵커로 git toplevel이 `/private/var`(`EvalSymlinks`) 해석되는 macOS 심링크로 문자열 비교 4곳(CacheMissKicks·WTBranch·NonSkipValues·WTSessions) 조정. 수리 후 영향계열 `-run 'TestReviewGate_|TestReviewEntry_|TestKickInFlight|TestProduceCodexReviewReceipt|TestCodexReviewGate_WTBranch|TestTreeScope|TestCodexReviewGateNonGitDir'` → **ok (95.6s)**.
- **턴종료 게이트 발견 처분 (M2 세션 r4 — r3 수리 커밋 리뷰, 본 소관 2건 수리)**:
  - **(P2) 프로듀서 저장 루트 미앵커**: 키는 앵커된 git 루트에서 계산되지만 `RecordReceipt(root, r)`가 원래 `--project-root`(하위 디렉터리)에 저장 → 진입 훅이 루트에서 조회해 FAIL을 놓침. 앵커 후 `root = scope.Dir`로 저장 루트 통일. 회귀 테스트 TestProduceCodexReviewReceipt_SubdirRootStoresAtGitRoot(하위 dir에서 produce → 루트 receipt 조회 → 하위 dir 세션 진입 차단 end-to-end).
  - **(P2) 만료 마커 교체 경합**: stat(stale)→remove→create 경합에서 두 호출이 모두 기동권 획득 가능(한쪽의 remove가 다른 쪽의 신선 마커를 지움). 교체를 **원자 rename 절차**로 교체: path를 고유 톰브스톤으로 이동(경합 시 ENOENT→재시도) → 이동 파일이 신선이면 동일 파일 복원+in-flight, 만료면 폐기 후 excl-create로 소유권 결정. 회귀 테스트 TestKickInFlight_LiveStealIsRestored(도난된 신선 마커의 byte-identical 복원).
  - 부수: 신규 테스트의 entrySeams 스텁 누락 수정. 수리 후 영향계열 → **ok (83.4s)**.
- **명시적 대기 (2026-10-08, 레인 종료 시점 — r4 푸시로 갱신)**: 레인 소관 전부 착지(PR #1802 open, head = r4 수리 푸시). 대기 사유: PR CI 최종·CodeRabbit 판독·병합·done = **리더 소관**. 재확인 지점: 리더 응답 또는 PR #1802 체크 완료. 팩토리 기록 거부(t1575 트리 미결속)는 리더 참고 사항으로 메모리 `project_card_t1575_landing.md`에 기록.





