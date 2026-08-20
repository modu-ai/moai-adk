# t161 후속 — 조상 `.moai/state` 는 누가 만드는가

리드 지시로 이어서 조사. 코드 변경 없음(보고만).

---

## 0. 내 귀속 논리 2건, 정정 수용 — 직접 재측정함

| 내 주장 | 실측 | 판정 |
|---|---|---|
| `dd060a191` 은 이번 배치 소속 | `git merge-base --is-ancestor dd060a191 4100d8767` → **참** | **틀렸다** — PR #1582 에 포함돼 이미 main 에 있었다 |
| "그땐 이 테스트가 없었다" | `gh pr checks 1582` → `Release Verify (windows-latest) pass 17m23s` | **틀렸다** — 같은 테스트가 Windows CI 를 이미 통과한 이력이 있다 |

**내 실수의 원인**: `git log --oneline -- internal/cli/tokens_test.go` 로 "그 파일을 마지막에
건드린 커밋"만 보고 **포함 관계를 검사하지 않은 채** "이번 배치 소속"으로 단정했다.
날짜(08-17)가 배치 기간과 겹친 것도 오판을 굳혔다. 파일 이력은 소속을 증명하지 않는다 —
소속은 `merge-base --is-ancestor` 로만 확인된다.

따라서 "신규 취약 테스트, 회귀 아님" 결론은 철회한다. 같은 테스트가 같은 OS 에서 통과하다
실패했으므로 **조건 (2)(조상 `.moai/state` 의 존재)가 새로 성립하게 된 것**이 맞다.

## 1. CI 로그 1차 증거 — 가설 2(8.3 단축경로)가 다시 배제됐다, 이번엔 실측으로

실패 런 `32313961117` 의 windows 잡(`96262413227`) 로그를 읽었다.

**(a) 실패한 tokens 테스트는 1개뿐이다.**

```
--- FAIL: TestRunTokensRecordAppendsLedger (0.02s)
    tokens_test.go:190: read ledger: open C:\Users\RUNNER~1\...\001\.moai\state\token-accounting.jsonl:
                        The system cannot find the path specified.
```

같은 패키지의 `internal/hook` 도 FAIL 이지만 원인은 `TestBranchGuard_Latency`
(median 23.8ms = 기준단위의 1.51x, 상한 1.50x) — 머신 코스트 기반 성능 단언이지
state 경로와 무관하다. **별개 건**이다.

**(b) 로그가 8.3 가설을 스스로 반증한다.** 같은 런에서 `TestRunTokensRecordSiblingSubagentFiles`
가 찍은 레코드 JSON:

```json
"cwd":"C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\TestRunTokensRecordSiblingSubagentFiles1195110583\\001"
```

즉 `t.Chdir(<단축형 tmp>)` 후 **`os.Getwd()` 가 같은 단축형을 그대로 돌려준다**. 따라서
cwd 폴백 분기를 탔다면 원장은 테스트가 읽는 바로 그 경로에 쓰였을 것이다. 그런데 경로가
없었다 ⇒ **폴백이 아니라 walk-up 분기가 탔고, 조상을 반환했다.**

t161 본 보고서는 이 결론을 로컬 재현으로 뒷받침했는데, 이제 **CI 1차 로그로도 확인**된다.
(로그에 `RUNNER~1` 과 `runneradmin` 두 형태가 섞여 나오지만 — 예: 25행, 413행 —
Getwd 가 chdir 한 형태를 보존하므로 읽기/쓰기 경로가 갈리지는 않는다.)

## 2. 결정적 새 사실 — 이 실패는 **간헐적**이다

`release/v3.1.1` 의 Release Verify 런 이력(실측):

| 런 | 시각(UTC) | 결과 |
|---|---|---|
| 32089524673 | 08-18 01:48 | success (PR #1582 — windows leg pass) |
| 32281097678 | 08-19 17:21 | success |
| 32313961117 | 08-19 23:36 | **windows FAIL** (본 건) |
| 32316474626 | 08-20 00:13 | success |
| 32318581872 | 08-20 00:46 | failure — **macos leg**, 별건 |
| 32318925824 | 08-20 00:52 | success |

**같은 브랜치에서 실패와 성공이 번갈아 난다.** 배치의 어떤 커밋이 조상 `.moai/state` 를
**무조건** 만든다면 매 런 실패해야 한다. 간헐성은 그 디렉터리가 **런마다 생기기도 하고
안 생기기도 한다**는 뜻이고, `go test ./...` 가 패키지를 **병렬 프로세스**로 돌리므로
"만드는 패키지"와 `internal/cli` 의 실행 순서가 런마다 달라지는 레이스와 맞아떨어진다.

리드의 교차 오염 가설을 지지하는 실측이다.

## 3. 만드는 주체 — 찾지 못했다. 범위는 좁혔다

### 3.1 로컬 실측: `internal/cli` 는 홈 레벨 state 를 만들지 않는다

`MOAI_HOME` 을 빈 임시 디렉터리로 돌려 `internal/cli` 전체를 돌렸다
(`paths.MoaiHome()` 은 절대경로 `MOAI_HOME` 을 루트 전체 대체로 존중한다).

```
$ MOAI_HOME=/tmp/<probe> go test -count=1 -timeout 20m ./internal/cli/
생성된 것:  <probe>/cache/update_check.json
            <probe>/.env.glm
생성 안 된 것: <probe>/state          ← 없음
```

**프로브 한계 2가지 (정직하게 기록)**:
- 이 프로브에서 `internal/cli` 는 **FAIL** 했다. 실패 20건은 전부 GLM 자격증명 경로 테스트
  (`TestGetGLMEnvPath_*`, `TestSaveGLMKey_*`, `TestInjectGLMEnv_*`, `TestLoadGLMKey_*`,
  `TestRunGLM_SavesKey`, `TestRunCG_NoAPIKey`)로, HOME 파생 경로를 단언하는데 `MOAI_HOME`
  이 루트를 갈아치워 깨진 것이다 — **프로브가 만든 인공물이지 제품 결함이 아니다**.
  이들 중 state 디렉터리를 만드는 것은 없으므로 위 음성 결과 자체는 유효하다.
- `os.UserHomeDir()` 를 **직접** 쓰는 writer 는 `MOAI_HOME` 을 무시하므로 이 프로브에
  **잡히지 않는다**. 그런 writer 가 있다면 내 실제 `~/.moai/state` 로 갔을 것이고,
  지금 이 머신은 다른 레인들이 동시에 돌고 있어 mtime 귀속이 불가능하다.

### 3.2 정적 조사: 홈 스코프 `.moai/state` 생성 후보 3개

| # | 위치 | 생성 시점 | 평가 |
|---|---|---|---|
| (a) | `internal/cli/deps.go:120-126` — `InitDependencies` 가 `paths.StateDir()` 를 구해 `<state>/loop` 을 `loop.NewFileStorage` 에 넘김 | **지연 생성** — `FileStorage.SaveState` (`internal/loop/storage.go:29`) 가 `MkdirAll` | **간헐성과 부합.** 루프가 실제로 상태를 저장하는 런에서만 `~/.moai/state/loop` 이 생긴다 |
| (b) | `internal/statusline/model_cache.go:52` — `WriteModelCache(homeDir, ...)` 가 `<homeDir>/.moai/state` 를 `MkdirAll` | 호출 시 | 현재 **자기 테스트에서만** 참조되고 인자는 tempDir. 프로덕션 호출자 0건 |
| (c) | `internal/config/cache.go` — `LoadWithCache` 가 `<configDir>/state/config-cache.json` 을 씀 → `<configDir>/state` 생성 | config 로드 시 | **이 클래스는 이미 알려져 있다**: `internal/cli/doctor_golden_test.go:60-68` 주석이 "a prior test's config load writes `<cwd>/.moai/state/config-cache.json` (the cache's MkdirAll side effect)" 라고 명시하고 `MOAI_CONFIG_CACHE_DISABLED` + HOME 고정으로 방어한다. 조상에 떨어지는지는 resolve 된 configDir 에 달렸다 |

### 3.3 배제한 후보

`deps.go:124` 의 폴백 `filepath.Join(os.TempDir(), ".moai", "state")` — 이게 만들어지면
`%TEMP%\.moai\state` 이므로 **모든** `t.TempDir()` 의 조상이 된다. 다만 리눅스에서도
`/tmp/.moai/state` 가 되어 `/tmp/TestX/001` 의 조상이므로 **리눅스도 같이 깨져야 한다**.
리눅스가 통과하므로 이 경로는 아니다. (이 폴백은 `paths.StateDir()` 가 에러일 때만 탄다.)

### 3.4 값싸게 확정하는 방법 (제안만, 실행 안 함)

`release-pr-multi-os.yml` 의 테스트 스텝 **앞뒤**로 진단 스텝 하나씩 추가:

```
dir "%USERPROFILE%\.moai"   &  dir "%TEMP%\.moai"
```

테스트 코드 변경 0, 스텝 2개. 다음 실패 런에서 조상이 어느 레벨에 생겼는지 즉시 드러난다.
워크플로 수정은 "보고만" 범위를 넘어 하지 않았다.

## 4. 같은 클래스의 범위 — 리드 요청 항목

**walk-up 소비자 (프로덕션, `findStateDir()` 호출)**: 4파일 6곳

```
internal/cli/clean.go:65
internal/cli/tokens.go:377      ← 이번에 주입으로 우회시킨 곳
internal/cli/chain.go:67
internal/cli/chain.go:353
internal/cli/state.go:78
internal/cli/state.go:154
```

**cwd 에 기대는 테스트**: `internal/cli` 에서 `t.Chdir` 사용 **25파일**. 그중 state
디렉터리를 함께 다루는 것이 **7파일**, tokens 계열 2개를 빼면 **5파일**:

```
internal/cli/inventory_test.go
internal/cli/state_m2_test.go
internal/cli/clean_home_carveout_test.go
internal/cli/doctor_golden_test.go
internal/cli/hook_routing_ledger_test.go
```

**다만 노출 여부는 파일 수와 다르다.** walk 은 cwd 를 **첫 번째로** 검사하므로,
`<cwd>/.moai/state` 를 **미리 만들어 두는** 테스트는 조상까지 올라가지 않아 면역이다.
취약한 것은 "폴백 분기가 타기를 기대하는" 테스트뿐이다 — 실패한 `AppendsLedger` 가 정확히
그 형태였다. 위 5개 중 `inventory_test.go:613` 과 `clean_home_carveout_test.go:250` 은
state 디렉터리를 `MkdirAll` 로 선생성하고, `doctor_golden_test.go` 는 cwd 를 비워 두는
대신 **HOME 고정 + config 캐시 비활성**으로 이 클래스를 이미 명시적으로 방어한다.

**판단**: 표면은 넓지만 대부분 이미 방어돼 있다. 카드 지시대로 **손대지 않았다.**

---

## 미검증 (Gaps)

- **만드는 주체 미확정.** §3.2 후보 3개로 좁혔을 뿐 어느 것이(혹은 다른 것이) 러너에서
  실제로 만들었는지는 확인하지 못했다. 러너 파일시스템을 볼 수 없고, 로컬 프로브는
  `os.UserHomeDir()` 직접 사용 writer 를 놓친다(§3.1).
- **`internal/cli` 밖은 프로브하지 않았다.** 교차 오염이 유력한데, 다른 패키지들
  (`internal/statusline`, `internal/hook`, `internal/session`, `internal/goal`)을 같은
  방식으로 돌려보지 않았다. 지금 이 머신에서 다른 레인들이 `go test` 를 돌리고 있어
  추가 스위트 실행이 부하·간섭을 만든다고 판단했다.
- **간헐성의 주기·조건 미측정.** 성공/실패 6개 런의 이력만 봤고, 실패율이나 어떤 런에서
  조상이 생겼는지는 로그에 단서가 없다.

## 잔여 위험

- 원인이 확정되지 않았으므로 **같은 클래스가 다시 터질 수 있다.** t161 수정은 tokens
  3곳의 노출만 없앴을 뿐, §4 의 walk-up 소비자 6곳과 아직 방어되지 않은 테스트는
  조상이 생기는 런에서 조용히 가로채인다. 조용하다는 게 핵심이다 — `clean.go` 나
  `chain.go` 는 실패하는 대신 **엉뚱한 디렉터리를 대상으로 정상 동작**할 수 있다.
- `TestBranchGuard_Latency`(머신 코스트 1.51x vs 상한 1.50x)는 별개 건으로 남아 있다.
  같은 런에서 함께 실패했으니 windows 레그가 붉은 이유가 두 가지였다는 점만 기록해 둔다.
