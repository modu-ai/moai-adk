# t560 격리 재실행 기록 (리드 승인 1회) — 다른 실패가 섞여 멈춤

트리: `.claude/worktrees/t560`, 브랜치 `WT-hook-env-scrub`, 흡수 병합 `2c07f89ff` 위 증거 커밋 `35d4c8c89` 직전 트리(코드는 `2c07f89ff` 와 같음).
목적: 창에서 실패한 `TestSessionStart_DeferredScanDoesNotBlockReturn` 이 홈 격리 상태에서도 단독으로 실패하는지 1회 확인(리드 결정 (c)).
판정 규칙(리드): 그 테스트 1개만 실패하고 부하가 높으면 다음 창에서 병합(a), 다른 실패가 섞이면 멈추고 보고.

## 실행

명령(리드 지정 형태, 가드 통과):

```
unset CLAUDE_CONFIG_DIR MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_PROJECT_DIR CLAUDE_PROJECT_DIR && MOAI_HOME=<세션 스크래치 빈 폴더> go test ./internal/hook/ -count=1
```

- 출력 `hook-isolated.txt`, 종료 `hook-isolated.exit` → `exit=1`, `FAIL github.com/modu-ai/moai-adk/internal/hook 81.644s`.
- 전후 기록 `lease-db-and-load.txt`:
  - 실행 직전 경쟁 `go (test|build|vet).*internal/(cli|hook)` → 없음(exit 1)
  - `~/.moai/run/profile-leases.db` 전: mtime 1789067923(04:18:43), 352256 B / 후: 1789067923, 352256 B — 변화 없음
  - load average 전 11.01 10.96 15.63 / 후 19.18 13.17 16.02

## 결과

- FAIL `TestEnsureGLMCredentials` — 서브테스트 `GLM_models_without_AUTH_TOKEN_auto-injects_from_env.glm`: `session_start_test.go:322` 주입 메시지 비어 있음, `:342` `AUTH_TOKEN = ""`, `:345` `BASE_URL = ""`, `:348` `DISABLE_EXPERIMENTAL_BETAS` 미설정.
- FAIL `TestEnsureGLMCredentialsFilePerm` — `settings_perm_test.go:83: expected credential injection message, got empty (env file: <TempDir>/.moai/.env.glm, exists: true)`.
- `TestSessionStart_DeferredScanDoesNotBlockReturn` — 출력에 이름 0회. 이번 실행에서는 실패하지 않았다. 비 `-v` 실행이라 통과 줄은 따로 관측하지 않았다.

리드 규칙상 다른 실패가 섞였으므로 여기서 멈췄다. 추가 실행은 하지 않았다.

## 새 실패의 귀속 — 격리 형태의 부산물

- 두 테스트는 격리 없이 돌린 창 실행(`../window/hook-packages.txt`)에서 실패하지 않았다(이름 0회).
- 테스트는 `t.Setenv("HOME", dir)` 로 홈을 바꾼 뒤 `<dir>/.moai/.env.glm` 에 자격 증명을 심는다(`settings_perm_test.go`, "Pre-seed ~/.moai/.env.glm via HOME override").
- 코드는 `loadGLMKeyFromEnvFile()` 에서 `paths.GlmEnvFile()` 로 경로를 정한다(`internal/hook/session_start.go:1344-1345`). `paths.MoaiHome()` 는 `MOAI_HOME` 이 절대경로면 그 값을 먼저 쓴다(`internal/paths/paths.go:68-71`).
- 이 실행은 `MOAI_HOME` 을 빈 스크래치 폴더로 줬으므로, 로더는 테스트가 심은 `$HOME/.moai/.env.glm` 이 아니라 빈 폴더를 봤다 → 주입할 키가 없었다.
- 부수 관측: 이 두 테스트는 `MOAI_HOME` 이 절대경로로 설정된 환경이면 실패한다. `HOME` 만 바꾸고 `MOAI_HOME` 을 지우지 않기 때문이다.

## 리드 판정 (2026-09-11, 이 기록 뒤 수신)

- 판정 (ii): 추가 재실행 없이 다음 창에서 병합한다.
- 근거로 든 것: 첫 실행(창)에서 실패는 `TestSessionStart_DeferredScanDoesNotBlockReturn` 1개(부하 18~21), 이번 실행에서는 그 테스트가 실패 목록에 없음, 관련 핸들러·테스트 blob 이 develop tip 과 같고 t617 창의 develop 쪽 기록과 일치.
- 새 실패 2개는 격리 형태(리드가 지정한 `MOAI_HOME` 덮어쓰기)의 부산물로 보고 판정에서 뺐다. 리드가 격리 형태가 틀렸다고 직접 밝혔다.
- 부수 관측(HOME 만 바꾸고 MOAI_HOME 을 지우지 않는 GLM 테스트 2개)은 internal/hook lease DB 격리 카드 후보에 합친다.
- 이 디렉터리는 병합 창 증거 커밋에 넣는다. 창 순번: t553(lane-9) → t560.

## 미검증

- `MOAI_HOME` 을 건드리지 않고 `CLAUDE_CONFIG_DIR` 만 지운 형태에서의 결과.
- `homestate.AcquireAdmissionLock`·`CheckRuntimeAdmission` 이 TempDir 프로젝트 기준으로 실제 홈 아래에 쓰는지.
- 창 실행에서 실패했던 테스트가 이번에 통과한 이유(부하 차이인지, 격리로 lease DB 경합이 사라져서인지).
