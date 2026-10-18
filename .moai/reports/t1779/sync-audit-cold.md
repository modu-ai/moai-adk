auditor-model: glm-5.3-flash

# t1779 cold sync-audit

verdict: PASS-WITH-DEBT
audited_sha: f8294f08cbdb311cd3f5f81a86aba0eb20e4f805
overall_score: 93/100 (가중: Functionality 40% · Security 25% · Craft 20% · Consistency 15%) — must-pass 전부 PASS

## Iteration history

- **1라운드 (HEAD `131dd6f6b6`)**: FAIL — F1 [P2] 재귀 팔이 gpt 백엔드 키 하나에 의존, 일반 Codex 세션 중첩 미커버. receipt `rcpt-e5e99caa2d58b5eb6b4ea451`.
- **2라운드 (HEAD `33ed6b72d9`, F1/F2 델타)**: FAIL — F1 직접 문 폐쇄·F2 폐쇄 확인, F4 [P2] 신규: --spawn 문이 `req.Env`를 버려 마커 미도달. F3 [P3] 라이터 핀이 앰비언트에 가려짐. receipt `rcpt-8a9e37b7b6b8a76030c73305`.
- **3라운드 (HEAD `f8294f08cb`, F4/F3 델타)**: **PASS-WITH-DEBT** — F4·F3 폐쇄 본 감사 실측 확인, 필수 codex 게이트 pass(결함 0건). receipt `rcpt-6ba8ed87f6f35a120e135200`.

## 심사 범위 (3라운드)

- 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1779` (브랜치 `WT-10-15-codex-3`)
- 델타: `33ed6b72d9..f8294f08cb` — 3파일 +37 (codex_launcher.go +5 · codex_launcher_test.go +3 · codex_lane_preflight_test.go +29)
- 감사 계획: `config_status: ok`, `cross_model_active: true`, codex gate **required**, `plan_source: config`

## F4/F3 수리 판정 (3라운드)

### F4 — 폐쇄(spawn 문 자가 식별)

- **수리**: `buildCodexSpawnCommandWithEnv`가 레인 키 blanks 바로 뒤에 `MOAI_CODEX_SESSION=1`을 명령행에 부착(codex_launcher.go:398-403 부근, diff 실측). 환경 모델 확인: tmux 창은 tmux **서버**의 env를 상속하고 이 프로세스의 것이 아니므로("A tmux window inherits the tmux SERVER's environment", codex_launcher.go:378-380 주석) 팬 env의 유일 채널은 이 명령행 배정 — 수리 위치가 정확. 생산 경로 확인: `defaultCodexSpawnLaunch` → `buildCodexSpawnCommandWithEnv`(:289) → `tmuxSpawnFn`(:293).
- **런타임 증거(본 감사 실측)**: `TestCodexSpawnAnchorsToPanePID` 스텁 tmux 출력 — `... MOAI_FACTORY_ROLE= MOAI_CODEX_SESSION=1 MOAI_HOME=... exec /test/codex` — 마커가 팬 명령행에 실제 착지.
- **양문형 시험**: `TestCodexSpawnCommandCarriesSessionMarker` — 직접 문(codexChildEnv)·spawn 문(buildCodexSpawnCommand) 셀이 모두 마커 단언. RED 이력(수리 전 spawn 셀 마커 누락)은 위임 기록과 정합.
- **ripple 수리**: `checkCodexSpawnRealAssembly` want-문자열에 `MOAI_CODEX_SESSION=1` 추가(:842·:853 두 시험이 헬퍼 공유 — "두 exact-pin 시험" 청구와 일치). 본 감사 실측 `TestCodexSpawn` 그룹 전체 PASS.

### F3 — 폐쇄(세정 후 라이터 증명)

- `TestCodexChildEnvStampsSessionMarkerWithoutAmbient` — `t.Setenv(key,"")` + `os.Unsetenv`로 앰비언트를 제거한 뒤 스탬프 "1" 단언: 2라운드 F3(앰비언트 가림)이 지적한 회귀 방어 공백 해소. 본 감사 실측 PASS(0.930s 그룹).

### 잔여 채무 (PASS-WITH-DEBT 근거 — 전부 optional 등급, 후속 카드 후보로 기록됨)

- D-t1779-a: managed-divert claude 자식은 미표식 — codex 세션 경로가 아닌 별도 코너.
- D-t1779-b: `factoryLaneRefusal`(factory_card.go:120-124, 큐 동사)의 동일 계통 맹점 — 본 카드 이전 선존재, 계통 범위.
- D-t1779-c: upstream `shell_environment_policy.inherit=core` 설정 시 마커가 세션 셸에 미도달 — 설정 의존 코너(기본 inherit=all).
- 세 항목 모두 카드 수용 문언(AC-1..4) 밖의 협소면이며 명명·추적됨 — 판정을 PASS에서 PASS-WITH-DEBT로 낮춘 근거.

## Dimension Scores (3라운드)

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | AC-1..4 전부 충족: 세 거부 팔 + 양문(직접/spawn) 자가 식별 + 등록 행 0 + 롤백 양성 대조 + 기존 패밀리 녹색 — 본 감사 실측 |
| Security (25%) | 90/100 | PASS | 마커는 식별자 상수; 거부면 확대만; ps 프로브 읽기 전용 |
| Craft (20%) | 93/100 | PASS | F3 폐쇄로 라이터 회귀 방어 복원; gofmt 클린 |
| Consistency (15%) | 90/100 | PASS | envkeys 상수 관례·기존 exact-pin/scrub 패턴 유지 |

## 5-Section Evidence-Bearing Report (3라운드)

### Claim
1. F4: spawn 팬 명령행이 마커를 운반해 양문이 동일하게 자가 식별. 2. F3: 앰비언트 없이 스탬프 착지 증명. 3. ripple 수리로 exact-pin 시험 무손상. 4. 전체 델타 회귀 없음.

### Evidence (명령 + 원문 출력)
- `git diff 33ed6b72d9..f8294f08cb` 전체 직독(3파일) + 생산 배선 확인(codex_launcher.go:289/:293/:381-385/:842/:853).
- 감사자 자체 스콥 시험(입장 load 30.62·mem free 83%, `-p 4 -count=1`, 단일 `-run`, TMPDIR 격리, 한 번에 한 호출):
  - `go test ./internal/cli/ -run 'TestCodexSpawnCommandCarriesSessionMarker|TestCodexChildEnvStampsSessionMarkerWithoutAmbient' -v` → 2/2 PASS, `ok ... 0.930s`.
  - `go test ./internal/cli/ -run 'TestCodexSpawn' -v` → 전체 PASS, `ok ... 1.538s` — 스텁 tmux 출력에 `MOAI_CODEX_SESSION=1` 실측(위 인용).
  - `go test ./internal/cli/ -run 'TestCodexLaneLaunch'` → `ok ... 4.388s`(프리플라이트 회귀 무).
  - `gofmt -l` 델타 3파일 → 출력 없음, `gofmt-exit:0`.
- 교차모델 3라운드: `mcp__moai__audit_multi` → `overall_verdict: pass`, `plan_source: config`, per-backend `codex / required / pass`(findings 0건), receipt `rcpt-6ba8ed87f6f35a120e135200`.

### Baseline-attribution
- 이 런, 이 트리(HEAD `f8294f08cb`)에서 직접 실행한 git/grep/go test 관측. 자원 입장: 라운드당 1분 load 실측(30.62 입장; 전 라운드 32.51·69.83 대기 기록은 각 라운드 절에 기재).

### Gaps
- 본 감사자는 tmux를 실제 띄워 팬 안에서 중첩 `-l`을 관측하지 않음 — 팬 마커 착지는 스텁 tmux 명령행 실측 + 배선 직독으로 확증(중첩 거부 동작 자체는 가드 절반 시험 실측).
- 설치 moai 바이너리 지체(ee981a932 — HEAD f8294f08c의 조상; 도구 자체 경고) — 판정은 트리 대상 검증에 근거, 설치 바이너리 출력에 의존 없음.
- codex 0.162.1 마커 무-export 전제 미검증(판정 불필요 — 자가 식별은 전제와 무관하게 유효).

### Residual-risk
- D-t1779-a/b/c 잔여 코너(위 잔여 채무) — 마커 미도달 경로는 여전히 존재하며, 이 게이트는 보안 경계가 아닌 최선 노력 사전 점검.

## Binding run conditions (후속 조치 권고)

1. D-t1779-b(큐 동사 factoryLaneRefusal 계통 맹점) → 별도 카드 승격 권장(높은 우선).
2. D-t1779-a(managed-divert claude 자식 표식) → 후속 카드 후보.
3. D-t1779-c(inherit=core 코너) → 문서화 또는 codex 설정 가이드 반영 후보.

## 수용 매트릭스 (카드 AC 기준, 최종)

| AC | 판정 | 증거 |
|----|------|------|
| AC-1 (실행 전 한 줄 거부·등록 행 0) | PASS | TERM=dumb·비TTY·샌드박스 3팔 + `assertCodexLaneRefusal`(줄 1·exit 비영·런치 0·행 0) — 실측 PASS |
| AC-2 (Codex 세션 재귀 거부) | PASS | gpt 키 + MOAI_CODEX_SESSION 마커, 직접·spawn 양문 표식 — 실측 PASS |
| AC-3 (롤백 경로 시험 확인) | PASS | `TestCodexSpawnStampFailureRollsBackPendingRegistration` — 1라운드 실측 PASS |
| AC-4 (사용 가능 환경 무발화) | PASS | TestMain 핀 + 재무장 대조, 기존 패밀리 그룹 실측 PASS |
