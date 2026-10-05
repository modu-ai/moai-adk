# SPEC-CODEX-GATE-SCOPE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` · `decision-index.md` (Tier M 산출 집합 + progress + decision gate).
- Tier: **M**. 근거는 AC 예산 — AC 13건(Tier S 상한 8 초과)·REQ 10건. LOC·파일 수는 S 범위로도 읽히나 검증 표면(판별기 + receipt 양축, env 매트릭스)이 M 판정이다.
- plan_status: **audit-ready** — plan-auditor 판정 기록(2026-10-01): verdict **PASS**, overall **0.96**(Tier M 임계 0.80), iteration **1/3**, blocking **0**. 보고서: `.moai/reports/t1383/plan-audit-iter1.md`. 감사 트리: `f4aa9bf99f8fd343d83e1178629450d61510bc25`(본 레인이 현재 HEAD와 일치함을 재확인 — 감사 이후 아티팩트 해시 무변경 전제 유지). 감사자가 spec.md §A 코드 좌표 8건 전부 재측정 적중, REQ-CRT-006·REQ-MCP-012 인용 실재 확인.
- optional-minor 4건(D1 plan.md `-run` 패턴 비고정 · D2 §A/HISTORY 측정 출처의 HEAD SHA 핀 누락 · D3 acceptance RED 기록 계약의 트리-SHA 요소 누락 · D4 §F 병합 후 창 묵시적 서술) — **run-phase 반영 예정**(레인 판정: 이 판정이 측정된 아티팩트 해시를 무효화하지 않는다). 본 문서는 이 행 외에 수정하지 않는다.
- 측정 원천: 본 트리(`.claude/worktrees/t1383`, 브랜치 `WT-gate-scope-lane`, 2026-10-01) 코드 좌표 직접 판독 — spec.md §A. 라이브 큐 카드 t1373·t1378·t1383 본문 판독 — **미커밋 출처**로서 decision-index Q2·Q4·Q5 에 권위 한계 표기. worker-63 구조 관측은 카드 본문 기재분(리드 제공, 본 레인 미독립 재현 — spec.md §A.6 출처 표기).
- 카드 전제 정정: 없음. 리드 지시 R1-R4 는 측정과 정합했고, 발견 7a(래퍼 쌍둔 주석 차이)는 본 트리 diff 로 재확인(비주석 행 0).
- 충돌 사전 검사: SPEC-CODEX-REVIEW-TARGET-001(completed) REQ-CRT-006 회귀선과 정합 — 본 SPEC 은 그 경로를 대체하지 않고 비카드 스코프로 고정한다(spec.md §F). t1373·t1378 은 미착지 카드로, 정합 방식은 REQ-CGS-004(라벨 파싱 부재) 하나다.

## §E.2 Run-phase Evidence

실행 트리: 카드 워크트리 분리(런타임 agent worktree, 베이스 f4aa9bf99 = 감사 트리와 동일 커밋), 브랜치 `WT-gate-scope-impl`. 커밋: M1 `6ca762e1f`(SPEC 산출 + 회귀선) → M3-M5 `70a8efe59`(구현 + 테스트). push 없음(레인 통합 몫).

### RED 관측 (acceptance §C — 구현 전 트리 SHA `6ca762e1f`)

- RED-1(컴파일 실패 — 테스트가 구현 심볼을 먼저 참조): `.moai/reports/t1383/red/m2-red1-compile-failure.txt` — `undefined: reviewScopeResolver ... [build failed]`.
- RED-2(스텁 선언으로 컴파일 가능화 후 런타임 실패, `-v` `=== RUN` 21행 관측): `.moai/reports/t1383/red/m2-red2-runtime-failures.txt` — 신규 동작 13테스트 전부 `--- FAIL`, 회귀선 2건(`TreeScopeRequestShapeUnchanged`·`CardScopeFailOpenOnMissingReviewer`)만 `--- PASS`.
- M1 회귀선의 변경 전(f4aa9bf99) 초록 관측: `.moai/reports/t1383/red/m1-treescope-regression-prechange-GREEN.txt`.
- GREEN 전환: `.moai/reports/t1383/red/m4-green-after-implementation.txt` — 셀렉터 17테스트 0실패.
- 요지 파일: `.moai/reports/t1383/red/README.md`. (`.moai/reports/*`는 gitignore — 디스크 보존이며 판정 출력은 본 문서와 완료 보고서가 운반한다.)

### AC 매트릭스 (게이트 조립 요청 필드 + receipt 바인딩 관측 — acceptance §A)

| AC | 검증 테스트 | 상태 | 근거(명령: `go test -count=1 -v -run '^TestCodexReviewGate|^TestCodexReviewScope|^TestProduceCodexReviewReceipt' ./internal/cli/`) |
|---|---|---|---|
| AC-CGS-001 | TestCodexReviewGate_CardScopeRequestIsCardDiff · TestProduceCodexReviewReceipt_CardScopeRecordsCardState | PASS | `--- PASS` — review/start target `{type: baseBranch, branch: <재계산 merge-base>}`, thread/start cwd = 카드 워크트리, primary 경로 미참조 |
| AC-CGS-002 | TestCodexReviewGate_TreeScopeRequestShapeUnchanged | PASS | `--- PASS` — 변경 전 트리(f4aa9bf99)에서도 초록; target.type `uncommittedChanges` 고정 |
| AC-CGS-003 | TestCodexReviewGate_StaleEnvKeepsTreeScope | PASS | `--- PASS` — env 라벨 2종에서 스코프 불변 + 트리 요청 형태 유지 |
| AC-CGS-004 | TestCodexReviewGate_BranchAloneDecidesCardScope | PASS | `--- PASS` — env 부재 + `WT-` 브랜치 → card, Branch/MergeBase 단정 |
| AC-CGS-005 | TestCodexReviewScope_LabelValuesDoNotAffectDecision | PASS | `--- PASS` — 임의 라벨 2종에서 `reflect.DeepEqual` 동일 스코프(양성 관측) |
| AC-CGS-006 | TestCodexReviewGate_FrozenProjectDirBypassed | PASS | `--- PASS` — payload project_dir + freeze env 모두 primary 지정, 세션 cwd 트리로 스코프·요청 |
| AC-CGS-007 | TestCodexReviewGate_EmptyCardDiffSkipsReviewer | PASS | `--- PASS` — 빈 카드 diff에서 reviewer 미호출(codexLookPath 가드 통과) + ALLOW |
| AC-CGS-008 | TestCodexReviewScope_UnidentifiedFallsToTree | PASS | `--- PASS` — detached / 비 WT- / 비git / merge-base 불가 4케이스 전부 tree, 근거 문자열 포함 |
| AC-CGS-009 | TestCodexReviewScope_ReceiptBoundToScopeState | PASS | `--- PASS` — 동일 카드 상태 재차단 유지 + 트리 상태 비매치 + 리더 트리 무영향 |
| AC-CGS-010 | TestCodexReviewGate_AndProducerSeeSameScope | PASS | `--- PASS` — 게이트/생산자의 cwd·target 객체 `reflect.DeepEqual` 동일 |
| AC-CGS-011 | TestCodexReviewGate_CardScopeFailOpenOnMissingReviewer (+기존 fail-open 테스트) | PASS | `--- PASS` — 카드 스코프에서 reviewer 부재 ALLOW, 호출 0 |
| AC-CGS-012 | TestCodexReviewScope_AbsorbedDevelopRecomputesBase | PASS | `--- PASS` — 흡수 후 merge-base 재계산, 흡수 전 바인딩 불일치 |
| AC-CGS-013 | TestCodexReviewGate_ScopeLogObservability | PASS | `--- PASS` — 클래스·근거(card: branch match + 브랜치명 / tree: 상이 문자열)·env 맥락 기록 |

미매핑 REQ 없음(acceptance §D 추적표 준수). REQ-CGS-001은 전 테스트의 전제 구조(판별 → 요청)로, REQ-CGS-002는 AC-001·012로, REQ-CGS-008은 AC-011로 각각 검증.

### 소관 패키지 재측정 (단위 = 패키지 전체)

`unset <레인 env> && go test -count=1 -timeout 30m -coverprofile=… ./internal/cli/... ./internal/codexwiring/...`

- 1차 실행(27.7분): FAIL 4건 → 즉시 판별. (a) `TestSyncGateLanguageDetectionMatchesScript/kotlin_source`·(b) `TestCharacterize_AuditPinPrecedenceAndBackendDefault` — **기저 재측정에서 동일 적색**: `git archive f4aa9bf99 | tar -x` 로 뽑은 무변경 트리에서 동일 실패 확인(본 카드 변경과 무관한 기존 결함 — 전자는 spec §A.7이 범위 밖으로 명시한 pre-t1379 Go 미러 결함, 후자는 운영자 실감사 핀을 읽는 환경 의존 가드). (c) `TestAuditLagUsesBinlagSeam`·(d) `TestWSR006_ReviewGateRootMatrix` — 본 카드 변경분(아래 정합 참조).
- 2차 실행(24.6분): **실패 (a)(b)만 잔존 — 신규 실패 0**. `.moai/reports/t1383/e3-owning-packages-rerun.txt`.
- 정합 2건: `TestWSR006_ReviewGateRootMatrix` 행 16-18은 REQ-CGS-005가 바꾼 분리 규칙(리뷰 게이트=세션 체인, 멀티 게이트=기존 체인)으로 재고정 — AC-WSR-006의 멀티 게이트 규약은 불변. `TestAuditLagUsesBinlagSeam` 허용 목록에 카드 diff 기저 측정 좌표 추가(대상 선정용 읽기이지 binary-lag 판정이 아님 — binlag.Evaluate 단일성 유지).
- 커버리지(전체 패키지 프로파일, 수정 파일): `codex_review_scope.go` 88.9%(12함수) · `codex_review_receipt.go` 93.3% · `codex_review_gate.go` 96.1% · `codex_stop_chain.go` 87.3% — 파일당 85% 기준 충족. 최저 함수 `cardScopeKeyParts` 75.6%(walk 오류·symlink 읽기 실패 방어선 미개방 — 신규 유미공 개정 테스트로 31.1%→75.6%).
- 정적: `go vet`(darwin) 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · `golangci-lint run --timeout=2m`(v2.1.6 = CI 판) `0 issues.` · gofmt 0.
- 범위 침범: `git diff --stat`에 `mcp_convergence.go`·`codex_sync_gate.go` 없음. 라벨 파싱: 프로덕션 4파일에서 `worker-<숫자>`/`lane-<숫자>` 리터럴 0(env 상수는 `config.EnvMoaiFactoryWorker`로만 참조). AskUserQuestion: 접촉 파일 전체 호출 0(선존 금지 문장 1건은 주석).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-10-01"
run_commit_sha: "70a8efe59917c87b83b5de2cc849557cd26728d3"
run_status: "complete"
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "not-run (isolated agent worktree — no shared-checkout commit; lane owns integration)"
l44_post_push_fetch: "not-run (no push by design — lane pushes develop after local merge)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: "go build ./... exit 0; go vet 0; golangci-lint 0 issues"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
total_run_phase_files: 13
m1_to_mN_commit_strategy: "M1 (SPEC intake + regression line, 6ca762e1f) -> M3-M5 single feat commit (70a8efe59) -> docs commit (progress only)"
known_base_reds_not_in_scope:
  - "TestSyncGateLanguageDetectionMatchesScript/kotlin_source — pre-existing on f4aa9bf99 (spec §A.7 out-of-scope Go-mirror defect)"
  - "TestCharacterize_AuditPinPrecedenceAndBackendDefault — pre-existing on f4aa9bf99 (reads the operator's live audit pin)"
ci_verdict: "PENDING — the repository-wide test verdict belongs to the CI run on the integration branch after the lane merges; this report claims the owning-package runs only"
```


## §E.4 Sync-phase Audit-Ready Signal

- sync_commit_sha: pending-backfill-sync
- sync_date: 2026-10-01
- sync 요약: sync 산출물 3건 — CHANGELOG [Unreleased] ### Added 항목(카드 t1383 3-phase close 서술), 본 §E.4 시그널, spec.md frontmatter `in-progress → completed` 전이. run 커밋 `6ca762e1f`→`70a8efe59`, tip `09e6c9cef`. sync 커밋은 레인이 직접 수행 — manager-docs 위임이 두 번 구조적으로 실패(스폰 격리가 자체 트리 생성 / 비격리 스폰이 레인 트리 고정 — 서브에이전트 cwd는 스폰 세션 트리에 고정, manager-docs 차단 보고 `feedback_sync_dispatch_tree_anchor.md` 참조)하여 §16 위임 대상 부재 상당으로 레인이 manager-docs의 사전 점검(B12 중복 0·AC 13·경로 존재)과 초안을 그대로 실행. 소유 예외는 완료 보고에 기록.

## §F Phase 4 Mode Selection

- **Decision: `serial`** — 단일 구현 위임(manager-develop) 1스폰, 마일스톤 순차.
- 입력: tier M · 구현 표면 3파일(codex_review_gate.go · codex_review_receipt.go · 테스트) · 단일 도메인(Go 게이트 로직) · 코딩 헤비(Anthropic coding-task caveat) · fanout 이득 없음(파일 간 의존: 판별기→요청→receipt 순차).
- Kickoff: 운영자 직답 폼 충족(2026-10-01, 리드 경유 AskUserQuestion 중계) — Q2·Q4·Q5 전부 승인, decision-index 반영 완료. iter1 plan-audit PASS 0.96; 확정 편집으로 해시 조건 무효화 → run 진입 시 Phase 1 재실행 예정(정상 경로).
- 소관: 레인(오케스트레이터) 기록, 2026-10-01.
