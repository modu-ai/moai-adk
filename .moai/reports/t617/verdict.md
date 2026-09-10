# t617 판정서 — 세션 시작 마이그레이션 테스트의 무조건 SKIP 해소

- 카드: t617 (Tier S, Class B)
- 트리: `.claude/worktrees/agent-a5130b13e9e45a187`, 브랜치 `WT-migration-skip`
- 측정 기준 커밋: 테스트 파일은 `98e8bd8a2`(이전 워커 보존 커밋)에서 바뀌지 않았고, 모든 측정은 그 뒤 증거 커밋만 쌓인 트리에서 했다. 판정서 커밋 직전 HEAD 는 M4 커밋 `a4f996c51` 이다.
- `internal/hook/session_start.go` sha256: `939547cc38b5b1f8707356a1828b8f654c50249be0ef99a8c99b22f017b247c8` (로컬 `develop` 과 동일)

## 1. Claim

1. 네 테스트는 이제 실제로 실행되며, 현재 구현에서 모두 통과한다.
2. 네 계약마다 `session_start.go` 를 최소한으로 망가뜨리면, 그 계약을 겨냥한 테스트가 실패한다. 뮤턴트 넷 모두 빨간색을 냈다.
3. SKIP 사유 셋 가운데 첫째는 태어날 때부터 거짓이었다. 둘째는 기다리던 의존성이 이미 들어왔지만 마이그레이션 경로가 그것을 쓰지 않는다. 셋째는 오늘도 참이다. 테스트는 참인 사유를 통과한 것처럼 꾸미지 않는다.
4. `TestSessionStart_DeferredScanDoesNotBlockReturn` 은 이 카드가 원인이 아니다. 다만 결함 여부는 판정하지 않았다. 부하가 기록된 실행은 모두 LOADED 조건이었고(FAIL 41/63, 막힌 시간 500.4–676.4ms), 조용한 조건의 측정은 없다(§2.6.1, §4).
5. 동작 결함 두 건을 재확인했다(§6). 이 카드에서는 고치지 않는다.

## 2. Evidence

### 2.1 SKIP 경위 (1단계) — `skip-origin.txt`

- 네 개의 플레이스홀더와 세션 시작 마이그레이션 배선이 **같은 커밋** `aa10ce386` 에서 들어왔다. 부모 커밋의 `session_start.go` 에는 `runMigration|migration.NewRunner|migration_error` 가 0건이고, `aa10ce386` 에는 `migration.NewRunner` 호출, 실패 시 `data["migration_error"]`, `cfg.System.Migrations.Disabled` 게이트가 모두 있다. "통합을 기다린다"는 사유는 쓰인 날 이미 거짓이었다.
- `a43f519e3` 는 인라인 블록을 `runMigration` 함수로 옮긴 추출 리팩터다. 같은 게이트와 같은 Data 키를 유지한다.
- SystemMessage: `HookOutput` 은 `SystemMessage` 필드를 갖고 있고 `session_start.go` 도 운영자 공지에 이미 쓴다(466-520행). `runMigration` 은 쓰지 않는다.
- `migrations.disabled`: 오늘 `system.yaml` 을 읽는 유일한 로더는 `systemFileWrapper` 로 디코딩하는데, 그 구조체의 필드는 `Hook` 하나뿐이다. 원래 커밋 시점의 로더는 `system.yaml` 을 아예 읽지 않았다.

### 2.2 원래 파일에서 SKIP 재현 (2단계)

`git show d3b7d438d:internal/hook/session_start_migration_test.go` 로 옛 파일을 넣고 실행했다 — `step2-pristine-skip.txt`:

```
--- SKIP: TestSessionStart_InvokesMigrationRunner (0.00s)
--- SKIP: TestSessionStart_MigrationFailure_SurfacesViaSystemMessage (0.00s)
--- SKIP: TestSessionStart_DisabledViaSystemYaml (0.00s)
--- SKIP: TestSessionStart_EnabledByDefault (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.720s
EXIT=0
```

교체한 파일의 해시는 `fda22142…`(`step2-swap-sha.txt`)였다. 원복 뒤 해시는 기준값 `b16332642f4d3698bc0f811d68e35c36c4271232ec63567f30d621c841168641` 과 같았고, `git status --short` 에는 새 보고서 파일만 보였다(`step2-restore-proof.txt`, 기준값은 `baseline-sha.txt`).

### 2.3 초록 확인과 0건 매칭 대조 (3단계)

명령: `go test ./internal/hook/ -v -count=1 -run '^(TestSessionStart_InvokesMigrationRunner|TestSessionStart_MigrationFailure_DoesNotBlockSession|TestSessionStart_MigrationsDisabled_SkipsRunner|TestSessionStart_EnabledByDefault)$'` — `step3-green.txt`

최상위 `--- PASS:` 는 4줄이다.

```
--- PASS: TestSessionStart_InvokesMigrationRunner (0.45s)
--- PASS: TestSessionStart_MigrationFailure_DoesNotBlockSession (0.19s)
--- PASS: TestSessionStart_MigrationsDisabled_SkipsRunner (0.40s)
--- PASS: TestSessionStart_EnabledByDefault (0.56s)
ok  	github.com/modu-ai/moai-adk/internal/hook	2.093s
EXIT=0
```

하위 테스트 7개도 모두 PASS 다. 없어진 옛 이름 둘(`MigrationFailure_SurfacesViaSystemMessage`, `DisabledViaSystemYaml`)만 골라 돌리면 `testing: warning: no tests to run` 과 `[no tests to run]` 이 찍힌다(`step3-oldnames-zeromatch.txt`). 나머지 두 이름은 새 파일에서도 그대로 쓰이므로 대조 대상이 아니다.

### 2.4 뮤턴트 (4단계)

관측기가 살아 있다는 근거는 변형 없이 돌린 `step3-green.txt` 다. 같은 명령과 같은 grep 패턴으로 PASS 줄 4개를 잡았다. 뮤턴트마다 변형 diff, 실행 출력, 원복 해시를 따로 남겼다. 네 번 모두 원복 뒤 해시가 `939547cc…b247c8` 과 같았다.

| # | 계약 | 변형 (`mutant-<n>-diff.txt`) | 빨간색이 된 테스트 (`mutant-<n>-run.txt`) | 초록 유지 |
|---|---|---|---|---|
| M1 | REQ-020 러너 호출 | Task 4 의 `runMigration(...)` 호출을 빈 맵 대입으로 교체 | `InvokesMigrationRunner/malformed…`, `MigrationFailure_DoesNotBlockSession`, `MigrationsDisabled_SkipsRunner/control…`, `EnabledByDefault` 하위 3개 전부 | `InvokesMigrationRunner/control…`, `MigrationsDisabled_SkipsRunner/disabled…` |
| M2 | REQ-021 세션을 막지 않음 | `out := &HookOutput{Data: jsonData}` 직후, `migration_error` 가 있으면 `StopReason` 설정과 `SetContinue(false)` | `MigrationFailure_DoesNotBlockSession` 하나 (`Continue=false on migration failure`) | 나머지 3개 전부 |
| M3 | REQ-032 비활성이면 건너뜀 | 게이트를 `if true \|\| cfg == nil \|\| ...` 로 단락 | `MigrationsDisabled_SkipsRunner/disabled…` 하나 | 같은 픽스처의 enabled 대조군 포함 나머지 전부 |
| M4 | REQ-032 기본값은 활성 | 게이트의 `!` 제거: `cfg == nil \|\| cfg.System.Migrations.Disabled` | `EnabledByDefault/default_config_(key_missing)`, `MigrationsDisabled_SkipsRunner` 하위 2개 | `EnabledByDefault/nil_ConfigProvider`, `…/provider_returning_nil_config`, `InvokesMigrationRunner`, `MigrationFailure_DoesNotBlockSession` |

M2 와 M3 는 겨냥한 테스트 하나만 빨간색으로 만들었다. M1 과 M4 는 같은 분기를 공유하는 테스트까지 함께 빨갛게 만든다. 러너가 호출됐는지를 관측하는 창구가 `migration_error` 하나뿐이어서 생기는 겹침이다.

### 2.5 `migrations.disabled` 로더 탐침 (4b)

임시 테스트 `internal/config/zz_t617_probe_test.go` 로 쟀다. 소스는 `step4b-probe-source.go.txt` 에 남겼고, 파일은 측정 뒤 지웠다. `git status` 에서 사라진 것도 확인했다.

- 1차 실행(`step4b-loader-probe.txt`)은 **무효**다. 같은 파일에 넣어 둔 양성 대조군 `hook.opt_in.enabled` 가 `false` 로 나왔고, 로더가 `.moai/config/config/sections` 를 찾았다는 WARN 이 찍혔다. `Load` 인자를 잘못 줘서 파일을 읽지 못한 것이다.
- 2차 실행(`step4b-loader-probe-v2.txt`):

```
PROBE loader: cfg.System.Migrations.Disabled=false
PROBE loader positive control (same file): cfg.System.Hook.OptIn.Enabled=true
PROBE direct unmarshal into SystemConfig: Migrations.Disabled=true
EXIT=0
```

같은 `system.yaml` 에서 `hook` 키는 읽혔는데 `migrations.disabled: true` 는 `false` 로 남는다. 직접 언마셜하면 `true` 가 나오므로 YAML 자체의 문제가 아니다.

### 2.6 불안정 테스트 (5단계)

`TestSessionStart_DeferredScanDoesNotBlockReturn` 은 `Handle` 경과 시간이 500ms 를 넘으면 실패한다(`session_start_parallel_test.go:96-97`).

| 조건 | 파일 | 결과 |
|---|---|---|
| 옛 플레이스홀더 테스트 파일(대조군) | `step5-deferred-control-pristine.txt` | PASS 5/5 (0.45s, 0.46s, 0.45s, 0.46s, 0.45s), EXIT=0 |
| 이 카드의 테스트 파일 | `step5-deferred-cardtree.txt` | PASS 5/5 (0.45s, 0.45s, 0.46s, 0.44s, 0.47s), EXIT=0 |

실행 방식: `go test ./internal/hook/ -v -count=5 -run '^TestSessionStart_DeferredScanDoesNotBlockReturn$'`. 반복 실행용 셸 루프는 워크트리 가드가 거부해서, 한 프로세스 안에서 순서대로 5회 도는 `-count=5` 로 바꿨다.

관측이 말해 주는 것은 여기까지다. 단독으로 돌리면 두 조건 모두 통과했고, 이 카드의 파일이 결과를 바꾸지 않았다. 이전 워커가 hook 패키지 전체를 돌렸을 때 본 585ms 실패는 이번에 재현되지 않았다. 표시된 테스트 소요 시간(0.44-0.47s)이 500ms 기준에 가깝다는 점은 부하에 민감할 수 있다는 가설과 맞지만, 그 가설은 이번 측정으로 검증하지 않았다(§4 Gaps).

### 2.6.1 부하 조건을 기록한 재측정

리드 지시: 이 테스트는 **경과 시간**을 잰다. 다른 레인의 컴파일·테스트 부하가 겹치면 코드 결함이 없어도 500ms 를 넘을 수 있다. 그래서 실행마다 부하 조건을 증거로 남긴다. 위 5단계 실행 10회는 부하 기록이 없어 분류하지 않는다.

**분류 기준(측정 전에 고정, `flaky-criteria.txt`).** QUIET = 실행 직전·직후 1분 load 가 모두 4.0 이하이고 다른 `go test` 가 없음. LOADED = 1분 load 가 한 번이라도 8.0 이상이거나 다른 `go test` 가 있음. 그 밖은 INTERMEDIATE(기록만 하고 판정에 쓰지 않음). 기계는 16코어다(`flaky-env.txt`).

**관측기 확인.** 이 호스트의 `ps` COMM 열은 16자에서 잘리므로(`flaky-ps-layout.excerpt.txt`) 실행 파일 정체는 ARGS 로 읽었고, 도구 셸 자신의 `zsh` 줄은 뺐다. 부재를 보고하기 전에 반드시 있는 대상을 잡는지 먼저 확인했다. run-1 실행 도중 찍은 스냅샷이 그 실행의 `go test`(pid 4388)와 `hook.test`(pid 4514)를 잡았다(`flaky-load-1-during.excerpt.txt`). run-0 의 도중 스냅샷은 실행이 끝난 뒤에 찍혀 아무것도 잡지 못했으므로 관측기 확인으로 세지 않는다.

| 실행 | 측정자 | 증거 | 1분 load (직전→직후) | 다른 `go test` | 분류 | 결과 | 막힌 시간 |
|---|---|---|---|---|---|---|---|
| run-0, `-count=20` | 워커 | `flaky-load-0-liveness.excerpt.txt` | 10.40 → 11.11 | 없음 | LOADED | FAIL 13 / PASS 7 | 500.4–676.4ms |
| run-1, `-count=40` | 워커 | `flaky-load-1-liveness.excerpt.txt` | 11.05 → 10.65 | 없음 | LOADED | FAIL 25 / PASS 15 | 500.6–643.2ms |
| step5b, `-count=3` | lane-6 | `step5b-deferred-run.txt`, `step5b-load-before.excerpt.txt`, `step5b-load-after.excerpt.txt` | 11.23 → 14.96 | 직전 `go test ./internal/kanban/`, 직후 `go test ./internal/web/` | LOADED | FAIL 3 / PASS 0 | 555.4–577.8ms |
| 조용함 확인 1 | 워커 | `flaky-quiet-check-1.excerpt.txt` | 10.45 | 없음 | QUIET 아님 | (실행 안 함) | — |
| 조용함 확인 2 | 워커 | `flaky-quiet-check-2.excerpt.txt` | 9.11 | 없음 | QUIET 아님 | (실행 안 함) | — |

step5b 의 시각은 직전 2026-09-10T08:28:34Z, 직후 08:28:49Z 이고, 실패 메시지 세 줄은 `step5b-deferred-run.txt` 에 그대로 있다(572.885709ms, 577.812875ms, 555.436916ms).

**판정: 없음.** 관측된 실패는 모두 LOADED 조건, 곧 경합 아래의 빨강이라 결함 판정의 근거가 되지 않는다. QUIET 조건은 한 번도 오지 않았다(§4 Gaps). 이 카드의 테스트 파일은 원인이 아니다. `-run` 이 이 테스트 하나만 실행하고, 카드 파일에는 `init` 이나 `TestMain` 이 없다(`grep -c -E '^func (init|TestMain)\('` → 0, 같은 파일에서 `^func Test` 는 4 로 잡혀 패턴이 살아 있음). 패키지의 `TestMain` 은 기존 `main_test.go` 에만 있다.

**가설(검증하지 않음).** 테스트는 지연 스캔 함수에 2초 블록을 주입한다(`time.After(2 * time.Second)`). 스캔이 동기로 돌았다면 경과 시간이 2초 이상이어야 하는데 관측값은 500–676ms 다. 따라서 실패는 스캔이 동기로 돌아서가 아니라 부하 아래에서 동기 경로 자체가 500ms 예산을 넘은 것일 수 있다. QUIET 측정 없이는 이 해석도, 조용할 때도 막히는 결함일 가능성도 배제하지 못한다.

**증거 위생.** 기계 전체 `ps -eo args` 원본에는 다른 세션의 명령줄이 통째로 들어 있어 커밋하지 않았다. 커밋한 것은 발췌본(uptime, ps 헤더, 줄 수, `go test`/`.test` 줄, 테스트 출력)뿐이다. 원본은 저장소 밖 세션 스크래치로 옮겼다.

### 2.7 정적 검사 (6단계)

| 명령 | 파일 | 결과 |
|---|---|---|
| `go vet ./internal/hook/` | `step6-vet.txt` | 출력 없음, EXIT=0 |
| `gofmt -l internal/hook/session_start_migration_test.go` | `step6-gofmt.txt` | 출력 없음, EXIT=0 |
| `gofmt -l <일부러 형식을 깨뜨린 파일>` (양성 대조) | `step6-gofmt-control.txt` | 해당 경로 출력, EXIT=0 — 형식이 틀리면 경로를 찍는다는 것을 확인 |
| `golangci-lint run ./internal/hook/...` | `step6-golangci-lint.txt` | `0 issues.`, EXIT=0 |

## 3. Baseline-attribution

- 모든 측정은 워크트리 `.claude/worktrees/agent-a5130b13e9e45a187`(브랜치 `WT-migration-skip`)에서 했다. 이 트리의 부모는 `d3b7d438d` 다.
- 측정하는 동안 트리에 쌓인 커밋은 `.moai/reports/t617/` 아래 증거 파일만 더했다. 커밋 순서: `4ab940b5a`(1단계) → `11a819dd2`(2단계) → `5d76a9d82`(3단계) → `0e5912eb6`(M1) → `c114c88b5`(M2, 탐침) → `8e5e1dc3d`(M3) → `a4f996c51`(M4).
- 측정 대상 소스의 해시: `session_start.go` `939547cc…b247c8`, 테스트 파일 `b16332642f…c841168641`(`baseline-sha.txt`, HEAD `98e8bd8a2` 에서 기록).
- 명령마다 같은 호출 안에서 `MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR MOAI_KANBAN*` 환경 변수를 지웠다.
- `prior-run/` 의 이전 워커 출력은 판정 근거로 쓰지 않았다. 다시 잴 위치를 찾는 데만 참고했다.

## 4. Gaps

1. **REQ-021 의 로그 절반이 검증되지 않았다.** 테스트 픽스처는 형식이 틀린 version 파일로 실패를 만든다. 이 실패는 `runner.Apply` 의 버전 읽기 단계(`internal/migration/runner.go:74-77`)에서 곧바로 돌아오고, `migrations.log` 에 쓰는 `Append` 는 개별 마이그레이션의 `m.Apply` 실패 경로(98-107행)에만 있다. 게다가 이 테스트 바이너리에는 마이그레이션 레지스트리가 링크되지 않아서 그 경로에 도달할 수도 없다(`go list -deps -test ./internal/hook/` 결과 `step7-hook-test-deps.txt`: 패키지 395개 중 `internal/migration` 은 279행에 있고 `internal/migration/migrations` 는 없다). 이 카드의 테스트는 "러너 실패 시 세션을 막지 않고, 오류를 기록하며, 버전 파일을 그대로 둔다"까지만 검증한다.
2. **REQ-021 의 SystemMessage 절반은 테스트가 없다.** 구현이 없기 때문이다(발견 A). 이를 단언하는 테스트를 넣으면 곧바로 실패하므로, 이번에는 넣지 않았다.
3. **REQ-032 의 `system.yaml` 경로는 테스트가 없다.** 로더가 바인딩하지 않기 때문이다(발견 B). 활성/비활성 테스트는 `ConfigProvider` 로 플래그를 주입하므로, `system.yaml` 경로가 동작한다는 근거로 읽어서는 안 된다.
4. **불안정 테스트 — 조용한 조건 미측정(Gap).** 부하가 기록된 실행 세 건(run-0, run-1, step5b)은 모두 LOADED 였고, 조용함 확인 두 번(load 10.45, 9.11)도 QUIET 기준(4.0 이하)에 닿지 않았다. 배치가 도는 동안에는 조용한 조건이 오지 않으므로 따로 기다리지 않았다. 부하 탓인지, 조용할 때도 막히는 결함인지는 QUIET 측정이 있어야 판정한다. 동기 경로가 예산을 넘는다는 해석은 가설이다(§2.6.1). 5단계의 무기록 실행 10회는 분류하지 않았고, 셸 반복문이 가드에 막혀 `-count=N` 한 프로세스 반복으로 대신했다.
5. **M1 과 M4 는 겨냥한 테스트 하나만 골라내지 못한다.** 러너 호출을 관측하는 창구가 `migration_error` 하나라서 여러 테스트가 같이 빨개진다. 계약마다 빨간색이 나왔다는 사실은 성립한다.
6. **CI 전체 스위트, darwin/windows 매트릭스, 병합 트리 재측정은 이 카드의 범위 밖이다.** 레인과 CI 가 맡는다.

## 5. Residual-risk

- 관측 창구가 `Data["migration_error"]` 키 하나다. 나중에 다른 곳에서 이 키를 쓰게 되면 "러너가 호출됐다"는 추론이 무너진다. 테스트가 아닌 코드에서 이 키를 쓰는 곳은 `internal/hook/session_start.go:681` 한 곳뿐임을 확인했다(`step7-migration-error-writers.txt`).
- 테스트는 마이그레이션 레지스트리가 비어 있다는 사실에 기댄다. 누군가 hook 테스트 바이너리에 `internal/migration/migrations` 를 링크하면, 형식이 틀린 version 파일이라는 픽스처는 그대로 유지되지만 정상 경로의 동작이 달라질 수 있다.
- 네 테스트는 실제 `Handle` 을 끝까지 호출한다. `DeferredScanDoesNotBlockReturn` 과 같은 패키지에서 도는 만큼 전체 실행 시 부하에 조금 보탠다. 영향의 크기는 재지 않았다.

## 6. 발견 사항과 후속 카드 제안

### 발견 A — 마이그레이션 실패가 사용자에게 보이지 않는다 (REQ-V3R2-RT-007-021)

`runMigration` 은 실패를 `slog.Warn` 과 `HookOutput.Data["migration_error"]` 에만 남긴다. `Data` 는 `json:"-"` 라서 Claude Code 로 나가지 않는다. SPEC 은 "마이그레이션 번호와 오류를 담은 `HookResponse.SystemMessage` 를 반드시 내보낸다"고 요구하고, 필드와 기존 SystemMessage 누적 방식(`session_start.go` 466-520행)은 이미 있다.

제안 카드 문구:
> `session_start` 마이그레이션 실패를 `SystemMessage` 로 알린다 (REQ-V3R2-RT-007-021). 지금은 `Data["migration_error"]`(json:"-")와 slog 에만 남아 사용자에게 보이지 않는다. 기존 운영자 공지와 같은 방식으로 SystemMessage 에 덧붙이고, 비활성 상태에서는 SystemMessage 가 없어야 한다는 REQ-032 조건을 함께 테스트한다. 버전 읽기 실패처럼 `migrations.log` 에 기록되지 않는 실패 경로도 함께 다룬다.

### 발견 B — `system.yaml` 의 `migrations.disabled` 가 설정에 반영되지 않는다 (REQ-V3R2-RT-007-032)

`loadSystemSection` 은 `systemFileWrapper{Hook}` 만 디코딩하므로 `cfg.System.Migrations` 에 값을 쓰는 곳이 없다. 탐침에서 같은 파일의 `hook.opt_in.enabled: true` 는 읽혔고, `migrations.disabled: true` 는 `false` 로 남았다. 사용자가 이 키로 마이그레이션을 끌 수 없다.

제안 카드 문구:
> `system.yaml` 의 `migrations.disabled` 를 로더에 바인딩한다 (REQ-V3R2-RT-007-032). `internal/config/loader_system.go` 의 `systemFileWrapper` 가 `hook` 블록만 디코딩해서, `migrations.disabled: true` 가 `cfg.System.Migrations.Disabled` 에 닿지 않는다. 기본값을 유지하는 부분 덮어쓰기 계약을 지키면서 `migrations` 블록을 바인딩하고, 설정 표면 정직성 인벤토리(SPEC-CONFIG-KEY-HONESTY-001)의 분류도 고친다. 수락 기준: `system.yaml` 픽스처로 `Loader.Load` 를 거쳐 세션 시작 러너가 건너뛰어지는 것을 확인하고, 키를 지운 뮤턴트에서 실패해야 한다.

## 7. 병합 트리 재측정 — 통합 창 안 (lane-6)

리드 지명 후 창을 잡고(`moai integration acquire --name lane-6`) 로컬 develop 을 흡수한 트리에서 다시 쟀다. 흡수 전 측정은 병합 뒤 근거로 재사용하지 않는다.

**흡수 대상의 최신성.** `git fetch origin develop` 후 `git rev-list --count --left-right origin/develop...develop` → `0 84`. 원격에만 있는 커밋은 없으므로 흡수 대상은 로컬 develop `f66cdc918` 이다.

**흡수.** `git merge --no-edit develop` → HEAD `fedfe3f43`, 트리 `a8622d6a9`. 흡수 전에 `git merge-tree --write-tree --name-only develop HEAD` 가 예측한 트리와 같다(충돌 파일 0).

**델타 판정** — `merge-tree-delta.txt`. 카드 기준 `d3b7d438d` 이후 develop 이 바꾼 파일 145개를, 병합 트리에서 잰 `go list -deps -test ./internal/hook/` 의 모듈 내부 패키지 68개와 대조했다.

- `go.mod` / `go.sum` 변경: 없음.
- hook 테스트가 의존하는 패키지 안에서 바뀐 `.go` 파일 10개. 운영 코드는 `internal/kanban/todo_root.go`, `internal/merge/differ.go`, `internal/mx/scanner.go`, `internal/session/store.go` 이고 나머지는 해당 패키지의 테스트다.
- 의존 패키지 디렉터리 아래에서 바뀐 비-`.go` 파일(embed 후보) 4개. 모두 `internal/template` 아래다.
- 대조: `internal/hook` 자신이 의존 목록에 잡힌다(참).

델타가 `internal/hook` 밖의 의존에 닿으므로, 리드 지시대로 범위를 넓혀 `internal/hook` 패키지 전체를 한 번 돌렸다.

| 명령 | 결과 | 증거 |
|---|---|---|
| `go test ./internal/hook/ -v -count=1 -run '^(이 카드의 네 테스트)$'` | EXIT=0, 최상위 `--- PASS:` 4건 — `InvokesMigrationRunner` · `MigrationFailure_DoesNotBlockSession` · `MigrationsDisabled_SkipsRunner` · `EnabledByDefault` | `merge-tree-green.txt` |
| `go test ./internal/hook/ -count=1 -timeout 600s` | EXIT=1. 실패는 `TestSessionStart_DeferredScanDoesNotBlockReturn` 한 건뿐(`Handle blocked 606.187625ms`), 나머지는 통과. `FAIL … 110.478s` | `merge-tree-hook-package.txt` |

패키지 실행의 부하 조건은 `merge-tree-hook-load-before.txt` / `merge-tree-hook-load-after.txt` 에 발췌만 기록했다. 직전(2026-09-10T09:04:02Z) 1분 load 는 **27.71**, 직후(09:06:10Z)는 **61.06** 이었다. 두 스냅샷 모두 다른 `go test` 줄은 0 이었고, 관측기가 `ps` 자신을 잡는 것은 확인했다. load 기준으로 이 실행은 LOADED 다. 따라서 이 실패도 §2.6.1 과 마찬가지로 경합 아래의 빨강이라 판정 근거가 아니다. 조용한 조건 미측정 Gap(§4 의 4번)은 그대로 남는다.

**이 절의 Gaps.** 델타가 닿은 의존 패키지(`internal/kanban`, `internal/merge`, `internal/mx`, `internal/session`, `internal/template`) 자체의 테스트는 여기서 돌리지 않았다. 그 변경을 가져온 카드들이 각자의 창에서 잰 몫이다. `internal/hook` 하위 패키지(`internal/hook/...`)도 돌리지 않았다. 전체 판정은 리드의 일괄 push 뒤 CI 몫이다.
