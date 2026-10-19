auditor-model: glm-5.3-flash

# t1779 cold sync-audit

verdict: PASS
audited_sha: cea40378863982d7d63c35e0d8996a82e8490734
overall_score: 94/100 (가중: Functionality 40% · Security 25% · Craft 20% · Consistency 15%) — must-pass 전부 PASS

## Iteration history

- **1라운드 (HEAD `131dd6f6b6`)**: FAIL — F1 [P2] 재귀 팔이 gpt 백엔드 키 하나에 의존, 일반 Codex 세션 중첩 미커버. receipt `rcpt-e5e99caa2d58b5eb6b4ea451`.
- **2라운드 (HEAD `33ed6b72d9`, F1/F2 델타)**: FAIL — F1 직접 문 폐쇄·F2 폐쇄 확인, F4 [P2] 신규: --spawn 문이 `req.Env`를 버려 마커 미도달. F3 [P3] 라이터 핀이 앰비언트에 가려짐. receipt `rcpt-8a9e37b7b6b8a76030c73305`.
- **3라운드 (HEAD `f8294f08cb`, F4/F3 델타)**: PASS-WITH-DEBT — F4·F3 폐쇄 확인, 잔여 채무 D-t1779-c(셸 정책 미러)·D-t1779-d(픽스처 세정)와 후속 후보 a·b 기록. receipt `rcpt-6ba8ed87f6f35a120e135200`.
- **판정 후 사건**: 타 레인 pkill로 moai MCP 서버 사망(보드 2fe7) → 종료체인 시작 마커가 감사 종료 후 기록되어 기존 영수증 전량 "시작 이전" 판독 → 재발행 시도 중 MCP 표면 철수·CLI codex-review fail 무기록 확인(판정 파일 사건 절 기재). 운영자 /mcp로 서버 부활.
- **4라운드 (HEAD `cea4037886`, P2-A/B 델타)**: **PASS** — D-t1779-c·d 폐쇄 본 감사 실측 확인, 필수 codex 게이트 pass(결함 0건), 신규 영수증 `rcpt-3ae1832530cd4c5aadefe9b9`(시작 마커 이후 발행). 잔여 a·b는 본 카드 델타 밖 후속 카드 후보로 라우팅 — 본 카드 판단에서 채무 아님.

## 심사 범위 (4라운드)

- 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1779` (브랜치 `WT-10-15-codex-3`)
- 델타: `f8294f08cb..cea4037886` — 프로덕션 1커밋(`04989cba68`) + 기록 전용 4커밋(판정 파일 커밋·reports untrack 2건·수리 기록). 기록 전용 커밋의 코드 영향 0 실측(`git diff` empty).
- 감사 계획: `config_status: ok`, `cross_model_active: true`, codex gate **required**, `plan_source: config`

## D-t1779-c / D-t1779-d 폐쇄 판정 (4라운드)

### D-t1779-c (P2-A) — 폐쇄: 셸 정책 미러

- `runCodexFactoryLane`이 `codexShellEnvSetArgs(append(slices.Clone(laneEnv), config.EnvMoaiCodexSession+"=1"))`로 레인 키와 **같은** `shell_environment_policy.set` 미러에 마커를 실어 보냄(codex_launcher.go:1085-1092 부근) — inherit=core/none에서도 세션 셸의 `moai codex -l`이 마커를 읽음. `slices.Clone`으로 laneEnv 비변경 — 기존 레인 키 스트림 무손상.
- **시험(본 감사 실측 PASS 2.29s)**: `TestCodexLaneLaunch_MirrorsSessionMarkerIntoShellEnvPolicy` — 실제 `-l` 조립 argv를 포착해 `codexShellEnvSetValues(argv)[MOAI_CODEX_SESSION] == "1"` 단언(전체 조립 경로 핀).

### D-t1779-d (P2-B) — 폐쇄: 양 세정기 커버

- `factoryAmbientEnvKeys`(→`clearFactoryAmbientEnv`, factory_test.go)와 `netScrubLaneEnv` 세정집합(lane_entry_golden_m1_test.go)에 `config.EnvMoaiCodexSession` 추가 — codex 호스팅 시험이 마커 없는 좌석에서 실게이트를 재무장.
- **시험(본 감사 실측 PASS)**: `TestLaneAmbientScrubCoversCodexSessionMarker` — 두 헬퍼 각각 서브테스트로 세정 단언.

### 회귀 실측 (본 감사)

- `TestCodexLaneLaunch|TestCodexSpawn` 결합 실행 → `ok ... 9.333s`(프리플라이트 14종 + 스폰 그룹 — 양문 마커·exact-pin·거부 3팔 전부 포함).
- `gofmt -l` 델타 4파일 → 출력 없음, `gofmt-exit:0`.
- 기록 전용 커밋의 코드 영향 0: `git diff f8294f08cb..cea4037886 -- internal/ cmd/` 중 프로덕션 변경은 `04989cba68` 하나로 한정(diff stat 실측).

## Dimension Scores (4라운드)

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 97/100 | PASS | AC-1..4 충족 + 양문 자가 식별(직접 env·spawn 명령행·셸 정책 미러 3경로) + 세정기 커버 — 전부 본 감사 실측 |
| Security (25%) | 90/100 | PASS | 마커는 식별자 상수; 거부면 확대만; 신규 미러는 기존 t1819 경로 재사용 |
| Craft (20%) | 95/100 | PASS | F3·P2-B 모두 세정-후 단언으로 회귀 방어 완비; gofmt 클린 |
| Consistency (15%) | 92/100 | PASS | 레인 키와 동일 미러 경로 재사용 — 패턴 일치 |

## 5-Section Evidence-Bearing Report (4라운드)

### Claim
1. P2-A: 셸 정책 미러에 마커 편승 완료. 2. P2-B: 양 세정기 마커 커버 완료. 3. 기록 전용 커밋 코드 영향 0. 4. 회귀 없음.

### Evidence (명령 + 원문 출력)
- `git diff f8294f08cb..cea4037886` 전체 직독(6파일, +65/−13) + 배선 확인(codex_launcher.go:1085-1092).
- `go test ./internal/cli/ -run 'TestCodexLaneLaunch_MirrorsSessionMarkerIntoShellEnvPolicy|TestLaneAmbientScrubCoversCodexSessionMarker' -v` → 2/2 PASS, `ok ... 3.826s`.
- `go test ./internal/cli/ -run 'TestCodexLaneLaunch|TestCodexSpawn'` → `ok ... 9.333s`.
- `gofmt -l` → 출력 없음(exit 0). 입장: load 23.05·mem 80% 실측 후 단일 Go 호출 직렬.
- 교차모델 4라운드: `mcp__moai__audit_multi` → `overall_verdict: pass`, `plan_source: config`, per-backend `codex / required / pass`(findings 0건), receipt `rcpt-3ae1832530cd4c5aadefe9b9` — `created_at: 2026-10-18T23:58:37Z`(시작 마커 20:24:18Z 이후 실측), 저장소 기록 확인.

### Baseline-attribution
- 이 런, 이 트리(HEAD `cea4037886`)에서 직접 실행한 git/grep/go test 관측. 영수증은 재접속된 서버(신규 인스턴스, build 7705b365e2)가 발행.

### Gaps
- 감사자 미실측: 라이브 codex 세션에서의 종단 중첩 관측(스텁/argv/미러 경로 실측 + 가드 절반 시험이 대리).
- 설치 moai 바이너리 지체(7705b365e — HEAD의 조상; 도구 자체 경고) — 판정은 트리 대상 검증에 근거.
- 작업 중 `progress.md`의 MM 상태(레인 기록 진행 중) 관측 — 코드 무관.

### Residual-risk (후속 카드 후보 — 본 카드 채무 아님)
- D-t1779-a: managed-divert claude 자식 미표식 코너.
- D-t1779-b: `factoryLaneRefusal`(큐 동사)의 동일 계통 맹점 — 기선존재, 별도 카드 권장(높은 우선).

## Binding run conditions

- 없음. 후속 라우팅 권고는 위 Residual-risk 두 건.

## 수용 매트릭스 (카드 AC 기준, 최종)

| AC | 판정 | 증거 |
|----|------|------|
| AC-1 (실행 전 한 줄 거부·등록 행 0) | PASS | TERM=dumb·비TTY·샌드박스 3팔 + 한 줄·비영 exit·런치 0·행 0 단언 — 실측 PASS |
| AC-2 (Codex 세션 재귀 거부) | PASS | gpt 키 + 마커, 직접 env·spawn 명령행·셸 정책 미러 3경로 표식 — 실측 PASS |
| AC-3 (롤백 경로 시험 확인) | PASS | `TestCodexSpawnStampFailureRollsBackPendingRegistration` — 1라운드 실측 PASS |
| AC-4 (사용 가능 환경 무발화) | PASS | TestMain 핀 + 재무장 대조, 패밀리 그룹 실측 PASS(4라운드 재확인) |
