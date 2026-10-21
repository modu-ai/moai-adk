# SPEC-REVIEW-GATE-REPAIR-002 — 인수 기준

## §A 관측 규율 [HARD]

1. **기준 트리.** 모든 앵커의 트리는 `8249bff8c3`(이 워크트리의 HEAD, 로컬 main 꼭짓점)다. 측정 시점마다 `git rev-parse --short HEAD` 로 다시 확인한다.
2. **릴리스 게이트 0건(plan 시점).** 이 plan 단계는 테스트를 실행하지 않았다(카드 제약). 따라서 이 문서의 어떤 셀도 이 세션이 실행해 관측한 테스트 출력을 담지 않으며, 릴리스 게이트로 승격할 근거가 없다. 행위 기준은 전부 RG-PENDING 이고, run 단계가 RED 를 원문으로 기록한 뒤에야 통과 판정의 대상이 된다.
3. **분류.** 모든 인수 기준은 `RG-PENDING` 으로 시작한다. 그중 AC-RGR-011·012 는 가드 후보다 — 가드 테스트의 존재는 이 세션이 코드에서 확인했으나, 현재 트리에서의 GREEN 은 이 세션이 측정하지 않았다(SPEC-001 셀 B-1 의 원문 기준선은 다른 트리 `2aab5f797` 의 것이다). M1 기준선 측정이 현재 GREEN 을 확정하면 그때 RG-GUARD 로 올린다.
4. **셀 네 요소.** 릴리스 게이트로 승격하는 셀은 (1) 명령(단일 호출, 파이프·리다이렉트·`&&` 없음), (2) 명령의 원문 stdout, (3) 종료 코드(별도 필드), (4) 트리 SHA 를 모두 가진다(verification-completeness.md §2.1). 이 문서는 그런 셀을 하나도 인용하지 않는다.
5. **증거 원천 Gap.** 카드가 인용한 t1620 감사 보고서 두 개는 추적되지 않는 파일이다(spec.md §A.1). 이 문서의 기대값은 카드 본문 요약과 SPEC-001 HISTORY·acceptance.md 셀 B-2 의 t1674 노트에서 왔다.

## §B AC 매트릭스

### AC-RGR-001 — 판정이 기록되면 표식이 풀린다 (REQ-RGR-001)

- **Given** 배경 검토가 실행 중이고 그 키의 진행 중 표식이 소유자 pid 를 싣고 있다.
- **When** 검토가 판정(pass, fail, inconclusive 중 무엇이든)을 기록한다.
- **Then** 판정 기록 단계에서 그 키의 표식이 제거된다. 세 판정 각각에서 같다.
- **분류** RG-PENDING. 가드: 신규 `internal/cli/codex_review_marker_lifetime_test.go`. 현재 코드에는 완료 시 해제가 없다(spec.md §A.2 — `kickInFlight` 가 배타적 생성만 한다).

### AC-RGR-002 — 고아 표식은 새 검토·새 판정 앞에서 회수된다 (REQ-RGR-002)

- **Given** (a) 소유자 pid 가 죽은 표식, (b) pid 를 판독할 수 없는 표식(빈 파일, 이전 형식의 시각 한 줄 포함)이 있다.
- **When** 새 검토가 그 키에서 시작되거나 새 판정이 기록되려 한다.
- **Then** 두 경우 모두 표식이 먼저 회수되어 검토·기록이 진행된다. 회수에 실패해도 검토는 막히지 않는다(fail-open).
- **분류** RG-PENDING. 가드: `codex_review_marker_lifetime_test.go`. 과도기 픽스처로 이전 형식 표식을 포함한다(plan.md §F 3).

### AC-RGR-003 — 표식은 동시 실행만 막는다 (REQ-RGR-003)

- **Given** (a) 같은 키의 검토 두 개가 동시에 시작되려 하고, (b) 같은 키의 판정이 기록된 뒤 새 검토가 시작되려 한다.
- **When** (a) 두 번째 시작과 (b) 새 시작이 각각 일어난다.
- **Then** (a) 두 번째 시작은 표식에 막힌다. (b) pass·fail 이면 결과 캐시가 재사용을 결정하고(표식은 관여하지 않는다), inconclusive 이면 새 검토가 시작된다 — 표식이 이미 풀렸기 때문이다.
- **분류** RG-PENDING. 가드: `codex_review_marker_lifetime_test.go`.

### AC-RGR-004 — fresh inconclusive 턴 종료는 허용하고 메시지를 내며 재검토를 시작한다 (REQ-RGR-004)

- **Given** 현재 키의 fresh inconclusive 영수증이 있고 그 키의 검토가 진행 중이지 않다.
- **When** 턴이 끝난다.
- **Then** 허용 결정이 유지되고, systemMessage 가 `inconclusive` 를 담으며, 그 키의 새 검토가 시작된다(간격 없음). 검토가 이미 진행 중이면 시작하지 않는다.
- **분류** RG-PENDING. 가드: 신규 `internal/cli/codex_review_inconclusive_test.go`. **기대값 변경**: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence`(현재 inconclusive 캐시 재사용이 킥을 억제함 — 셀 B-2 t1674 노트)의 기대값이 바뀐다. 근거 REQ-RGR-004, 운영자 판정 Q4 = (a). 변경 기록은 progress.md §E.2 에 남긴다.

### AC-RGR-005 — 스위치 기본값은 커밋된 값이고 판독 실패는 상태 메시지를 낸다 (REQ-RGR-005)

- **Given** (a) HEAD 의 커밋된 `workflow.yaml` 은 `enabled: true` 이지만 작업 트리는 `false` 로 고쳐져 있고, (b) 커밋된 파일이 없거나 키가 없거나 값이 불리언이 아니다.
- **When** 게이트가 스위치를 읽는다.
- **Then** (a) 커밋된 값 true 를 따른다(작업 트리의 미커밋 false 는 무시된다). (b) 검토를 시작하지 않고 허용하며 systemMessage 가 `커밋 값 읽기 실패` 를 담는다.
- **분류** RG-PENDING. 가드: 강화된 `internal/cli/review_gate_config_key_test.go`(HEAD 커밋 픽스처 — B1 (3)). 현행 판독기는 작업 트리 파일을 읽는다(spec.md §A.2).

### AC-RGR-006 — 재정의 파일은 규칙을 만족할 때만 적용된다 (REQ-RGR-006)

- **Given** 주 체크아웃의 `.moai/state/codex-review-gate/run-switch.json` 이 (a) 형식이 맞고 `run_id` 가 단일 활성 run 과 같으며 세션이 member lane 이다, (b) 스키마가 틀렸다, (c) `run_id` 가 다른 run 이다, (d) 세션이 member lane 이 아니다.
- **When** 게이트가 스위치를 해석한다.
- **Then** (a)만 파일의 `codex_review_gate` 값을 적용하고, (b)(c)(d)는 파일이 없는 것과 같다(커밋된 값을 따른다).
- **분류** RG-PENDING. 가드: 신규 `internal/cli/codex_review_run_switch_test.go`. 해석 규칙의 원문은 SPEC-001 design.md §5 다.

### AC-RGR-007 — 스위치 출처가 관측 행에 기록된다 (REQ-RGR-007)

- **Given** 게이트가 스위치를 재정의 파일에서 해석했다(또는 커밋된 값에서 해석했다).
- **When** 해석이 끝난다.
- **Then** 표준 오류의 관측 행이 출처(재정의 파일 / 커밋된 값)를 필드로 담는다.
- **분류** RG-PENDING. 가드: `codex_review_run_switch_test.go`.

### AC-RGR-008 — 게이트 래퍼는 모든 턴 종료에서 바이너리를 호출한다 (REQ-RGR-008)

- **Given** 게이트가 꺼져 있다(커밋된 값 false 또는 재정의 파일 off)。
- **When** 턴 종료 래퍼(`handle-codex-review-gate.sh`)가 실행된다.
- **Then** 래퍼는 셸에서 스위치를 해석하지 않고 처리기를 호출하며, 처리기가 낸 `스위치 꺼짐` systemMessage 가 래퍼의 stdout JSON 을 통해 전달된다. 래퍼의 조용한 종료 35·36·60행은 존재하지 않는다.
- **분류** RG-PENDING. 가드: 격리 래퍼 실행(종료 코드·stdout JSON 원문, progress.md §E.2). **기대값 변경**: `TestReviewGateWrappersSelfGateBeforeBinaryResolution`(`internal/template/review_gate_registration_test.go:214`)의 게이트 래퍼 행 — 전제가 "자체 게이트 선행"에서 "모든 턴 종료에서 처리기 호출"로 바뀐다. 진입 래퍼 행은 유지. 근거 REQ-RGR-008, v0.2.6 판정.

### AC-RGR-009 — 래퍼의 조용한 종료와 처리기 실패가 상태 이름을 낸다 (REQ-RGR-009)

- **Given** (a) moai 바이너리가 세 계층 모두에 없고, (b) 처리기 호출이 0 이 아닌 종료로 끝난다.
- **When** 턴 종료 래퍼가 각 경우에 실행된다.
- **Then** (a) systemMessage 가 `바이너리 없음` 을 담는다. (b) systemMessage 가 `처리기 실패` 를 담는다. 두 경우 모두 래퍼는 0 으로 끝나고 차단 결정은 없다.
- **분류** RG-PENDING. 가드: 격리 래퍼 실행 — (a)는 SPEC-001 v0.2.10 이 관측한 방식(세 계층 부재 조건), (b)는 SPEC-001 v0.2.11 이 관측한 방식(종료 3 스텁). 두 이름은 래퍼 범위로, 열두 처리기 어휘에 더해지지 않는다.

### AC-RGR-010 — `tree_scope 건너뜀` 메시지가 사유를 구별한다 (REQ-RGR-010)

- **Given** (a) `tree_scope: skip` 이 명시 설정된 트리 클래스 세션과, (b) 설정 없이 기본 판정으로 제외된 트리 클래스 세션이 있다.
- **When** 각 세션의 턴이 끝난다.
- **Then** 두 경우 모두 상태 이름은 `tree_scope 건너뜀` 으로 같고, 메시지 본문이 (a) 명시 설정 제외와 (b) 기본 판정 제외를 구별해 담는다. 허용 결정은 바뀌지 않는다.
- **분류** RG-PENDING. 가드: `internal/cli/codex_review_gate_test.go` 의 신규 함수. 사유 판별 기준은 M1 의 스킵 행렬 관측이 확정한다(plan.md §F 8).

### AC-RGR-011 — 가드 후보: 키당 차단 상한의 경합 가드가 유지된다 (R-1 회귀)

- **Given** 같은 리뷰 키에서 턴 종료 여럿이 배리어로 동시에 진행된다.
- **When** 전부 차단 예약을 거친다.
- **Then** 차단은 키당 총 C 회를 넘지 않는다.
- **분류** RG-PENDING(가드 후보). 가드: `TestReviewGate_ConcurrentBlockCap`(`internal/cli/codex_review_gate_concurrency_test.go:23`) — 존재는 이 세션이 확인했다. 현재 트리에서의 GREEN 은 미측정(M1 이 측정, M7 이 재실행). 상류 수리 근거: SPEC-001 v0.2.11 SF-1(커밋 `4bd5d433a` RED, `a13e02705` GREEN), 코드 앵커 `codexReviewReserveBlock`(`internal/cli/codex_review_gate.go:380`).

### AC-RGR-012 — 가드 후보: 진입 전달과 pass 재사용, 건너뜀 결정이 유지된다 (SPEC-001 REQ-RG-001·002·005·006 보호)

- **Given** M5·M6 의 변경이 캐시·표식·스위치·래퍼 표면에서 일어난 뒤다.
- **When** SPEC-001 셀 B-1 의 진입·재사용·건너뜀 가드들(`TestReviewEntry_FreshPassAllows`, `TestReviewGate_DisabledAllows`, `TestReviewEntry_DisabledAllows`, `TestReviewGate_LoopPreventionAllows` 등)을 현재 트리에서 실행한다.
- **Then** 전부 PASS 다 — 진입 훅은 여전히 차단하지 않고, pass 재사용과 건너뜀 결정은 그대로다.
- **분류** RG-PENDING(가드 후보). 셀 B-1 의 원문 기준선은 트리 `2aab5f797` 의 것이며 현재 트리 GREEN 은 미측정(M1 이 측정, M7 이 변경 전후 비교표를 progress.md §E.2 에 남긴다).

## §C 기준선 (이 세션의 관측 — plan 단계)

이 세션은 코드 앵커를 확인했고 테스트를 실행하지 않았다(§A.2). 관측의 원천 명령과 요지:

- **파일 존재**: `ls` — `internal/cli/codex_review_cache_test.go`, `codex_review_delay_test.go`, `codex_review_entry_delivery_test.go`(이미 존재 — SPEC-001 M2 분), `codex_review_gate_wiring_test.go`, `review_gate_config_key_test.go`, `internal/template/review_gate_registration_test.go`, `codex_review_gate_symlink_test.go` 존재. `codex_review_inconclusive_test.go` 부재(신규 대상 확정).
- **래퍼 두 면**: `cmp .claude/hooks/moai/handle-codex-review-gate.sh internal/template/templates/.claude/hooks/moai/handle-codex-review-gate.sh` → 무출력(바이트 동일). 래퍼 본문 35·36·60·79–80·83–86행의 조용한 종료는 `Read` 로 확인했다.
- **표식·재사용·잠금 앵커**: `grep`/`sed` — `codexReviewKickMarkerPath`(`codex_review_delay.go:82`), `kickInFlight`(`:96`, 내용 시각 한 줄), 캐시 inconclusive 재사용 분기(`codex_review_gate.go:227–259`), `codexReviewReserveBlock`(`:380`), `acquireGateLockNoFollowImpl` 두 플랫폼 존재, `AcquireGateLock`(`gate_lock.go:99`)의 링크 추종 열기, `codexReviewBreakStaleBlockLock`(`codex_review_gate.go:353`)의 `os.Stat`·`os.ReadFile`.
- **스위치 판독기**: `readCodexReviewGateEnabled`(`internal/cli/mcp_codex.go:2602`, 작업 트리 `os.ReadFile`). `factorymsg.ResolveActiveRun`(`internal/factorymsg/store.go`), `resolveFactoryCardRun`(`internal/cli/factory_card.go:2875`).
- **테스트 기준선의 부재(Gap)**: 위 테스트들의 현재 트리(`8249bff8c3`)에서의 실행 출력은 없다. M1 이 기대값 변경 대상(`TestReviewGate_InconclusiveReceiptIsNotPassEvidence`, `TestReviewGateWrappersSelfGateBeforeBinaryResolution`)과 가드 후보(AC-RGR-011·012)의 원문 기준선을 progress.md §E.2 에 새로 쌓는다.

## §D RED 확립 규율 (run 단계)

- 이 SPEC 은 릴리스 게이트를 승격하지 않는다(plan 시점 릴리스 게이트 0건).
- 신규 RED 테스트(`codex_review_marker_lifetime_test.go`, `codex_review_inconclusive_test.go`, `codex_review_run_switch_test.go`)와 기대값 변경은 각 마일스톤 커밋의 RED 선행을 커밋 그래프로 증명한다: 마일스톤 커밋의 첫 부모 트리에 덮어쓰기 단위를 얹어 기록된 명령이 EXPECTED_RED 로 실패해야 한다. **덮어쓰기 단위의 테스트는 첫 부모의 비테스트 코드만으로 컴파일되어야 한다. 컴파일 실패는 `tdd-result-contract.md` 의 TOOL_FAILURE 이며 RED 로 세지 않는다.** 명령·종료 코드·원문 출력은 progress.md §E.2 에 남긴다.
- 결함을 고정한 기존 테스트의 기대값 변경(AC-RGR-004·008, 조건부로 wiring 테스트)은 근거(REQ 번호 또는 판정)와 함께 progress.md 에 기록한다. 삭제는 하지 않는다.

## §E 완료 정의 (DoD)

- 열린 질문 1(처리기 오류 경로 명명)의 리더 판정이 기록되어 있고, 판정 전에 M6 가 그 분기를 바꾸지 않았다.
- AC-RGR-001–010 각 항목이 "가드 GREEN" 이거나 "마일스톤 커밋의 새 테스트는 첫 부모에서 EXPECTED_RED 로 실패하고 마일스톤 커밋과 최종 트리에서 GREEN 이다(검사 기록은 progress.md §E.2)" 상태다. AC-RGR-011·012 는 M1 의 현재 GREEN 측정과 M7 의 변경 전후 비교표가 progress.md §E.2 에 있다.
- 기대값 변경 세 곳(`TestReviewGate_InconclusiveReceiptIsNotPassEvidence`, `TestReviewGateWrappersSelfGateBeforeBinaryResolution`, 리더 판정 시 wiring 테스트)의 "이전 기대값 → 새 기대값 → 근거" 기록이 progress.md 에 있다.
- M6 커밋마다 래퍼의 두 면이 같은 커밋에 있다(`git show --stat` 확인). 래퍼의 격리 실행 관측(종료 코드·stdout JSON)이 §E.2 에 있다.
- F5·F2 잔여·N1·N3·F6·F7·F3 의 후속 등록이 이 SPEC 의 범위 밖 절과 일치한다. `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/` 과 `.moai/config/sections/workflow.yaml` 을 바꾼 흔적이 없다.
- 릴리스 게이트 승격 사례는 이 SPEC 에서 인용하지 않는다.

## §F 잔여 위험

- 이 문서의 앵커는 darwin/arm64 한 플랫폼에서 읽은 것이다. 윈도 경로(`codexReviewBreakStaleBlockLock`, `gate_lock_windows.go`)는 읽기만 했고 실행하지 않았다.
- 기준선 실행 출력이 없다(§A.2, §C Gap). M1 이 채운다. 채우기 전까지 어떤 AC 도 통과로 읽지 않는다.
- `tree_scope` 스킵 행렬은 카드 기록과 코드 주석으로만 파악했다. M1 관측이 REQ-RGR-010 의 사유 판별 기준을 확정한다. 행렬이 카드 기록과 다르면 REQ-RGR-010 을 재개정한다.
- 래퍼의 빈 stdout(종료 0, 출력 없음) 경우는 REQ-RGR-009 의 범위에 두지 않았다 — 처리기는 최소 `{}` 를 내므로 이 경우는 비정상이지만, 관측도 규격도 없다. 잔여로 남긴다.
- 표식의 버킷 성분 유지 여부는 run 단계 결정사항이다(plan.md §D M5). 결정이 REQ-RGR-001–003 의 의미를 벗어나면 이 문서를 고친다.
- 열린 질문 1 의 판정이 (a) 이면 처리기 어휘가 열세가 되므로 spec.md 의 「Out of Scope — 열두 상태 어휘의 변경」 절을 개정해야 한다.
- N1(SPEC-001 `spec.md:53`)은 카드 보도이며 미검증이다. 이 문서는 그 문장을 인용하지 않는다.
