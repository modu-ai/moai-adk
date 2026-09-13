# t654 card verdict — AS-5 launcher integration (run+sync window)

- card: t654 · spec: SPEC-MOAI-GATEWAY-001 v0.13.0 · worktree `.claude/worktrees/t654`
- branch: `WT-gateway-launchers` · merge base `6de8dd489` (= origin/develop d416f8162 + local develop 0f5dbe096 흡수)
- run HEAD `059f4e700` · sync commit `ce4a3c6ed` · verdict commit: 이 커밋

## Claim

AS-5의 이번 창 범위(자동화 가능 전 항목)를 착지·실측했고, 라이브 전제가 필요한 항목 6건은 창 대기 Gap으로 기록했다. SPEC은 다중 카드 시리즈 정책대로 `in-progress`를 유지하며 이 카드로 종결되지 않는다(라이브 창 카드 t844·t851 잔존). umbrella 종결 전이는 후속 sync 몫이다.

## Evidence

**커밋 (8건, 카드 id 전부 포함)**

| 커밋 | 내용 |
|---|---|
| `15a3a21f5` | plan-phase SPEC 0.13.0 개정 (6파일 +326/−7) |
| `99bbe5cad` | plan-audit 수선 D1-D3 + 감사 보고서 |
| `6de8dd489` | develop 흡수 (0f5dbe096) |
| `5672f5029` | A5-M1 launch 조립 auth 표시(authMethod/outputPolicy) + 대기 오류 게이트 대조군 |
| `aec09a5d6` | A5-M2 appliedEpoch 영구 복원(Store.Rebase 동일 트랜잭션·v2 manifest·RestoreRebaseLedger) + fork prefix 원장 대조(receipt.ChainDigest·Manager.ForkSession) |
| `3e28b60e3` | A5-M3/M4 GLM text-only 게이트 + 구독/API 이중 경로 자동화 보장(PKCE-only catalog·토큰 저장소 접근 0·자동전환 구조적 불가) |
| `2113c1e14` + `886959071` | A5-M5/M6 증거·progress §E.2/§E.3·@MX:NOTE |
| `059f4e700` | t708 비침범 가드 소유 시리즈 면제(아래 별절) |
| `ce4a3c6ed` | sync — §E.4 신호 + CHANGELOG 연기 |

**감사 이력**

- plan-audit: CONDITIONAL PASS 0.90 (`.moai/reports/t654/as5-plan-audit.md`) — D1(launcher.go:142 제2 게이트 위치)·D2((d) 강제 전제 집합 명명)·D3(진입 함수 비공유 정정) 반영 완료(`99bbe5cad`).
- sync-audit: **PASS 0.93** (`.moai/reports/t654/as5-sync-audit.md`) — Functionality 95·Security 94·Craft 90·Consistency 93, must-pass 방화벽 통과, 차단 finding 0.

**품질 실측 (2026-09-14, 본 워크트리)**

- AC 기계 판정: `go test ./internal/gateway/... -run 'TestRebaseLedgerRestoration|TestForkPrefixCrossCheck'` GREEN (RED 로그: `as5-m2-epoch-red.log`·`as5-m2-forkprefix-red.log` / GREEN: `as5-m2-abc-green.log`). `go test ./internal/cli/ -run 'TestGatewayLaunch'` 전부 PASS.
- 커버리지: receipt 88.1% / gateway 91.7% / 신규 ForkSession 88.9%·ForkAt 87.5%.
- lint 변경 패키지 `0 issues.` / gofmt 클린 / `GOOS=windows GOARCH=amd64 go build ./...`·vet exit 0 / Windows 명명 시험 16개 skip 0.
- CHANGELOG: 미발행 — B12 사전 발행 grep `grep -c 'SPEC-MOAI-GATEWAY-001' CHANGELOG.md` → 0 (전후 동일). AC-MG-026 (d) deploy-gate 창으로 연기.

**t708 비침범 가드 개정 (리드 재판정 승인, verdict 명시 조항)**

- 개정 diff: 커밋 `059f4e700` — `internal/cli/gateway_preserve_test.go` 1파일. `preservedDiffViolations`에 면제 조항 추가(동일 diff가 `.moai/specs/SPEC-MOAI-GATEWAY-001/`를 실으면 면제) + 주석에 판정 근거와 대리 한계 명시 + 양방향 판별 시험(`TestPreservedDiffDiscriminatorRejectsViolations` 기존 유지, `TestPreservedDiffExemptionOwningSeriesTouch` 신설).
- 근거: 가드 원안(5b10efb66, t708 AC-EVR-011)의 주석은 scope를 t708 카드 diff로 묶으나 영구 테스트로 착지해 소유 시리즈(receipt/**)의 진화를 얼림 — t653 e45f50a8d(00:29)가 가드(01:07)에 선행하는 평행 착지로 실증된 소유 축. 운영자 무응답 폴백안을 리드가 재판정 승인(2026-09-14); 병합 창에서 재판정 가능.
- 잠금 본연 기능 보존: SPEC 터치 없는 침범 diff·위반+타시리즈 diff·빈 diff("unmeasurable")는 모두 여전히 적색(sync-audit 직접 재현).

## Baseline-attribution

모든 명령·출력은 2026-09-14, 본 워크트리(branch `WT-gateway-launchers`)에서 직접 실행. run 측정 baseline `6de8dd489` → 커밋 진행 표대로. 사전 존재 환경 실패(codexbridge lifecycle_subprocess 4건 + gateway `TestAppServerSubprocessHTTPToolContinuation`)는 `git archive 6de8dd489` 추출 트리에서 동일 재현해 이 diff 이전 환경 의존으로 귀속(sync-audit 독립 재현 포함). 최종 전체 스위트는 레인 규율상 CI 소관(수선은 시험 파일 1개, 판정 시험 3건 green — `as5-guard-exemption-green.log`; 수선 전 전체 로그 `as5-final-suite.log` 보존).

## Gaps (창 대기 6건 — progress.md §E.3/§E.4와 동일 목록)

1. AC-MG-026 (a) GREEN 전환 — AS-014~022 전수 PASS 뒤 대기 오류 리터럴 2곳(gpt.go:54·launcher.go:142) 제거 + 대조군 시험 gate-open 전환.
2. AS-014·017·018·019 실제 PTY 실증 — t851(400 해소) 이후 라이브 창 (증거 경로 `as5-pty-*.log` 준비됨).
3. AS-021 실계정 이중 모드 실측 — 동일 라이브 창.
4. AS-022 GitHub CI 증거 — 리드 push 후 `gh workflow run release-pr-multi-os.yml` → `test-stream-release-verify-windows-latest` 아티팩트 판독 → `as5-windows-ci-verdict.md` (절차: `as5-windows-ci-prep.md`).
5. AC-MG-026 (d) rc 배포 게이트 — 강제 전제 집합 {(a),(b),(c),AS-017,AS-019,AS-021} PASS 후 실행(대기 명령 기록됨). 이번 창에서 실행하지 않음.
6. (M2b) appliedEpoch 생산 호출자 배선 — compaction 실세션 흐름(t844)에서 이어짐.

**sync-audit optional findings (차단 0, 다음 sync에서 묶어 처리 권고)**

- F1: §E.2 M3 GREEN 본문 수치 표류(`ok 7.440s` vs 로그 `ok 6.914s`) — progress.md 정정.
- F2: "사전존재 2하위시험 skip" 주장이 비verbose 로그에서 관측 불가 — verbose 로그 보충.
- F3: `as5-m4-authmodes-green.log`가 cli 측 1테스트 누락(재실행으로 양측 PASS 확인됨) — 로그 보충.
- F4: engine.go:39 주석이 미배선 ForkPrefix를 현재형 서술 — **t844에 codexbridge.Config 생성 시 ForkPrefix 배선 의무 이관 권고**.
- F5: `TestAppServerSubprocessHTTPToolContinuation` flaky(같은 트리 ok/FAIL 혼재) — lifecycle_subprocess 계열 정리 카드 대상.

## Residual-risk

- 가드 면제 대리가 "SPEC 디렉터리 터치"라는 기계 대리 — 위장 면제는 plan-audit 표면에서 걸리는 구조이며 코드 주석에 한계 명시됨.
- overlay authMethod/outputPolicy의 사용자 가시 렌더링은 클라이언트 실측(라이브 창)에서 확정.
- manifest v1 경험본의 v2 복원은 epoch 0 — dev 단계라 실 피해 경로 없으나 프로덕션 데이터 형태는 라이브 창에서 재판정.
- 이미지 탐지는 Messages 블록 형태 스캔 — upstream 형식 변화 시 음성 픽스처가 회귀를 잡는 구조.
- Windows 회귀는 release 시점 CI에만 노출(운영자 결정 4 잔여 위험 유지 — ci.yml 상시 레그 미채택).
- 교차모델 audit_multi는 서버 바이너리 래그(8050369e5)로 anchor 거부 2회 — fail-open으로 세션 내 Claude 판정 확정(plan-auditor 기록). 바이너리 갱신 뒤 재판정 여지.
