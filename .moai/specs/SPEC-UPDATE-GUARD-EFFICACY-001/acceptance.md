# SPEC-UPDATE-GUARD-EFFICACY-001 — 인수 기준

## §A 판정 규율 (모든 AC에 구속)

이 SPEC은 "가드는 있는데 판정이 무르다"는 부채를 해소하는 것이 목적이므로, **자기 자신의 AC가 같은 무름을 재생산하면 안 된다.** 아래 다섯 조항이 모든 AC에 구속된다.

### A.1 `-run` 선택자의 공허 통과 금지

`go test -run <pattern>`은 패턴이 **0개**에 매치해도 exit 0으로 끝난다. 따라서 `-run`을 쓰는 모든 AC는 다음 두 가지를 함께 요구한다.

1. **존재 grep** — 대상 테스트 함수명이 소스에 실재함을 확인한다.
2. **`--- PASS:` 행 관측** — `-v` 출력에서 해당 테스트(또는 서브테스트) 경로의 `--- PASS:` 행이 실제로 찍혔음을 확인한다. 맨 `ok` 한 줄만으로는 통과로 인정하지 않는다.

`exit 0`만 인용하는 `-run` 판정은 이 SPEC에서 무효다.

### A.2 운영자 머신 상태 의존 금지

판정의 판별력이 실행 머신의 상태(홈 디렉터리 존재 여부, 설치된 도구, 기존 파일 유무)에 좌우되면 안 된다. **이것이 부채 D2 자체이므로 이 SPEC의 AC에서 재생산하는 것은 명시적 위반이다.** 판정에 픽스처가 필요하면 판정 명령이 직접 만들고, 만든 직후 그 픽스처가 비어 있지 않음을 확인한다.

### A.3 절대 앵커 금지

판정에 커밋 SHA를 핀으로 박거나 줄 번호를 앵커로 쓰지 않는다. 비교 기준(base)은 판정 시점에 `git merge-base origin/main HEAD`로 계산하고, 코드 위치는 내용 토큰(함수명, 식별자, 문자열 리터럴)으로 지목한다. 저작 시점 절대 앵커는 검증 시점에 낡는다.

### A.4 반증 의무

커버리지·가드 계열 AC는 **변경 전 baseline 수치**와 **가드가 실패하는 것을 관측한 기록**을 함께 요구한다. 통과만 관측된 가드는 실효성이 증명되지 않은 것으로 본다(REQ-UGE-013). 반증 방법은 두 가지 중 하나다.

- **overlay 반증** — `git show <base>:<path>`로 변경 전 파일을 꺼내 `-overlay`로 덮어 실행하고 `--- FAIL`을 관측한다.
- **변형(mutation) 반증** — 명시된 변형을 프로덕션 코드에 적용해 실행하고 `--- FAIL`을 관측한 뒤 되돌린다.

### A.5 보존 가드 예외 (preservation guard carve-out)

"변하지 않아야 한다"를 주장하는 AC는 기대값이 baseline과 같을 수 있다. 그 자체는 A.4 위반이 아니며, **무엇이 그 값을 움직이는지**를 명시하면 보존 가드로 인정한다. 해당 AC: AC-UGE-006(회귀 없음), AC-UGE-013(템플릿 중립성 `0`), AC-UGE-014(빌드 exit 0). 세 항목 모두 `-run` 선택자를 쓰지 않으므로 A.1의 0-매치 공허 모드가 없다.

---

## §B baseline 실측 (plan-phase 관측)

아래는 plan-phase에서 `.claude/worktrees/e2-data-survival`(PR #1275 브랜치, HEAD `457e1013f`)에 대해 **실제로 실행해 관측한** 값이다. run-phase는 리베이스 후 §C 절차로 **재측정**하고, 값이 달라지면 달라진 값을 baseline으로 삼는다(수치를 옮겨 적는 것은 baseline이 아니다).

| 항목 | 측정 명령 | 관측값 |
|---|---|---|
| 사용자 영역 가드가 구동하는 레지스트리 함수 | `go test -run 'TestMoaiUpdate_PreservesUserArea' -covermode=set -coverpkg=./internal/cli/...,./internal/cli/update/... -coverprofile=P ./internal/cli/` → `go tool cover -func=P` | `CleanMoaiManagedPaths 66.7%` / `MigrateLegacyMemoryDir 26.9%` / `runCleanReinstall 0.0%` / `BackupMoaiConfig 0.0%` |
| update 서브시스템의 비이음매 HOME 호출부 | `grep -c 'userHomeDir()' internal/cli/update_clean_install.go internal/cli/update_template_sync.go` | `update_clean_install.go:1` / `update_template_sync.go:2` |
| stat 이음매 존재 | `grep -rn 'osStatFn\|statFn' internal/cli/*.go` | 매치 0건 (이음매 부재) |
| `mergeBackPreserveInventory` 커버리지 | `go test -covermode=set -coverprofile=P ./internal/cli/` → `go tool cover -func=P \| grep mergeBackPreserveInventory` | `94.1%` (M6 이후. M6 이전은 `64.3%`) |
| `snapshotDir` 사용처 | `grep -rln 'snapshotDir(' internal/cli/*_test.go` | `update_safety_test.go` 1개 파일 |

**§1.3 정정 반영**: `runCleanReinstall`·`BackupMoaiConfig`의 `0.0%`는 위 `-run` 선택자 스코프 한정이며, 두 함수 모두 패키지 전체 테스트에서는 이미 구동된다. 이 SPEC의 AC는 "커버리지를 0 위로 올린다"가 아니라 **"사용자 영역 보존 어서션이 그 함수를 지난다"**를 판정한다 — 그래서 AC-UGE-009/010은 반드시 가드 스코프 선택자로 측정하고, 동시에 A.4 반증을 요구한다.

---

## §C 인수 기준

### M1 — stat 이음매 (REQ-UGE-001, 002, 003)

#### AC-UGE-001 — 주입 가능한 stat 이음매가 존재하고 프로덕션이 그것을 경유한다

```bash
BASE=$(git merge-base origin/main HEAD)

# (a) 이음매 선언 — 기본값이 os.Stat 인 패키지 변수
grep -nE '^var [a-zA-Z]*[Ss]tatFn = os\.Stat$' internal/cli/update_preserve_inventory.go

# (b) mergeBackPreserveInventory 본문이 os.Stat 을 직접 부르지 않고 이음매를 경유
sed -n '/^func mergeBackPreserveInventory/,/^}/p' internal/cli/update_preserve_inventory.go \
  | grep -c 'os\.Stat('
sed -n '/^func mergeBackPreserveInventory/,/^}/p' internal/cli/update_preserve_inventory.go \
  | grep -cE '[a-zA-Z]*[Ss]tatFn\('

# (c) baseline — 변경 전에는 이음매가 없고 os.Stat 직접 호출이 1건
git show "$BASE":internal/cli/update_preserve_inventory.go \
  | sed -n '/^func mergeBackPreserveInventory/,/^}/p' | grep -c 'os\.Stat('
```

기대: (a) 한 행이 출력된다. (b) 첫 grep `0`, 둘째 grep `1`. (c) `1`.

**실패 조건**: 이음매를 선언만 하고 프로덕션이 여전히 `os.Stat`을 직접 부르면 (b)의 첫 grep이 `1`이 되어 실패한다. 함수 본문 창(`sed`)으로 스코프를 좁혔으므로 파일 내 다른 함수의 `os.Stat`은 계수에 들어오지 않는다.

#### AC-UGE-002 — stat 실패 서브테스트가 어떤 플랫폼에서도 건너뛰지 않는다

```bash
# (a) 존재 grep (§A.1)
grep -c 'func TestMergeBackPreserveInventory_PartialRestore' internal/cli/update_preserve_partial_test.go

# (b) stat 서브테스트 본문에 플랫폼/권한 스킵 가드가 없다
sed -n '/t.Run("stat_failure/,/^	})/p' internal/cli/update_preserve_partial_test.go \
  | grep -cE 't\.Skip|runtime\.GOOS|os\.Geteuid|os\.Chmod'

# (c) PASS 행 관측 (§A.1) — SKIP 이면 이 행이 나오지 않는다
go test -run 'TestMergeBackPreserveInventory_PartialRestore' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -E '^\s+--- (PASS|SKIP|FAIL): TestMergeBackPreserveInventory_PartialRestore/stat_failure'

# (d) windows 타깃 컴파일 — 스킵 없이 컴파일되는지
GOOS=windows GOARCH=amd64 go vet ./internal/cli/; echo "win-vet-exit=$?"
```

기대: (a) `1`. (b) `0`. (c) `--- PASS: …/stat_failure…` 한 행 (`SKIP`/`FAIL`이면 실패). (d) `win-vet-exit=0`.

**실패 조건**: `chmod` 방식을 남겨 두면 (b)가 `≥1`이 되어 실패한다. 이음매 주입이 실제로 분기에 닿지 못하면 (c)가 `FAIL`이 된다.

**미검증 (§3.5로 이월)**: 이 판정은 macOS/Linux에서 실행되므로 **Windows에서의 실제 실행**은 CI 잡이 관측한다. (b)+(d)는 "스킵 조건이 소스에 없고 windows 타깃으로 컴파일된다"까지만 증명하며, Windows 런타임 통과는 CI 결과를 인용해야 한다.

#### AC-UGE-003 — 가드가 변경 전 코드에 대해 실패한다 (§A.4 overlay 반증)

```bash
BASE=$(git merge-base origin/main HEAD)
D=$(mktemp -d /tmp/uge-m1.XXXXXX)
git show "$BASE":internal/cli/update_preserve_inventory.go > "$D/reverted.go"
printf '{"Replace":{"%s/internal/cli/update_preserve_inventory.go":"%s/reverted.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" \
  -run 'TestMergeBackPreserveInventory_PartialRestore' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -cE '^(--- FAIL|FAIL)'
rm -rf "$D"
```

기대: `≥1` (변경 전 코드에는 이음매가 없으므로 주입이 컴파일되지 않거나 분기가 도달되지 않아 실패). 출력이 `0`이면 가드가 M1 변경에 의존하지 않는다는 뜻이므로 AC 실패다.

#### AC-UGE-004 — `mergeBackPreserveInventory` 커버리지가 baseline 이상이다

```bash
go test -covermode=set -coverprofile=/tmp/uge-m1-cov.out -count=1 ./internal/cli/ >/dev/null
go tool cover -func=/tmp/uge-m1-cov.out | grep 'mergeBackPreserveInventory'
```

기대: 백분율이 **94.1% 이상**. baseline은 §B 표(M6 이후 `94.1%`)이며 run-phase 리베이스 후 재측정한 값으로 갱신한다. 이음매 도입으로 stat 분기가 모든 실행에서 도달하므로 하락하면 안 된다.

---

### M2 — HOME 이음매 균일화 (REQ-UGE-004, 005)

#### AC-UGE-005 — 세 호출부가 이음매를 경유한다

```bash
BASE=$(git merge-base origin/main HEAD)

# (a) 변경 후 — update 서브시스템에 비이음매 호출이 남지 않는다
grep -c 'userHomeDir()' internal/cli/update_clean_install.go internal/cli/update_template_sync.go

# (b) 변경 후 — 이음매 호출이 각각 존재한다
grep -c 'userHomeDirFn()' internal/cli/update_clean_install.go internal/cli/update_template_sync.go

# (c) baseline (§A.3 — SHA 핀 대신 merge-base 계산)
git show "$BASE":internal/cli/update_clean_install.go | grep -c 'userHomeDir()'
git show "$BASE":internal/cli/update_template_sync.go | grep -c 'userHomeDir()'

# (d) 범위 밖 호출부는 건드리지 않았다 (spec.md §3)
grep -c 'userHomeDir()' internal/cli/glm.go
```

기대: (a) 두 파일 모두 `0`. (b) `update_clean_install.go:1`, `update_template_sync.go:2`. (c) 각각 `1`, `2`. (d) `1` — 변경 전과 동일(범위 밖 보존).

**주의**: `grep 'userHomeDir()'`는 `userHomeDirFn()`에 매치하지 않는다(`Fn` 때문에 `()`가 이어지지 않음). 두 패턴이 서로 오염되지 않음을 (a)와 (b)가 함께 확인한다.

#### AC-UGE-006 — 동작 변화가 없다 (§A.5 보존 가드)

```bash
go test -count=1 ./internal/cli/... 2>&1 | tail -20; echo "cli-test-exit=${PIPESTATUS[0]}"
go vet ./...; echo "vet-exit=$?"
```

기대: `cli-test-exit=0`, `vet-exit=0`.

**이 값을 움직이는 것**: 이음매 경유로 HOME 해석 결과가 달라지면(예: `userHomeDirFn`의 기본값이 `userHomeDir`이 아니게 되거나 nil이면) 기존 테스트가 깨진다. `-run` 선택자를 쓰지 않으므로 0-매치 공허 모드가 없다(§A.5).

---

### M3 — 기계 독립적 HOME 반경 판정 (REQ-UGE-006, 007, 008)

#### AC-UGE-007 — 판정이 자기가 만든 고정 HOME 픽스처를 대상으로 한다

판정 절차는 sentinel HOME 트리를 **직접 만들고**, 그 트리가 비어 있지 않음을 확인한 뒤, 테스트 바이너리만 그 HOME으로 실행한다. `go test`에 직접 HOME을 주면 Go 툴체인의 캐시/모듈 경로까지 옮겨가므로, **먼저 컴파일하고 컴파일된 바이너리만 sentinel HOME으로 실행한다.**

```bash
SENT=$(mktemp -d /tmp/uge-home.XXXXXX)
mkdir -p "$SENT/.claude/hooks/moai"
printf 'probe\n' > "$SENT/.claude/hooks/moai/probe.txt"

# (a) REQ-UGE-007 — before 스냅샷이 비어 있지 않아야 삭제 방향이 판별된다
find "$SENT/.claude/hooks" -mindepth 1 | sed "s|^$SENT||" | sort > /tmp/uge-home-before.txt
test -s /tmp/uge-home-before.txt; echo "before-nonempty=$?"

# (b) 테스트 바이너리 컴파일 (툴체인은 실제 HOME 사용)
go test -c -o /tmp/uge-cli.test ./internal/cli/; echo "compile-exit=$?"

# (c) 존재 grep + PASS 행 관측 (§A.1)
grep -c 'func TestEnsureGlobalSettingsEnv_HooksRemovalRadius' internal/cli/update_home_radius_test.go
HOME="$SENT" /tmp/uge-cli.test -test.run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -test.v 2>&1 \
  | grep -E '^--- (PASS|SKIP|FAIL): TestEnsureGlobalSettingsEnv_HooksRemovalRadius'

# (d) sentinel HOME 이 불변인가
find "$SENT/.claude/hooks" -mindepth 1 | sed "s|^$SENT||" | sort > /tmp/uge-home-after.txt
diff /tmp/uge-home-before.txt /tmp/uge-home-after.txt; echo "diff-exit=$?"

# (e) NFR-UGE-002 — 프로세스 환경 변수 변조가 아니라 이음매로 격리했는가
grep -c 't.Setenv("HOME"' internal/cli/update_home_radius_test.go
sed -n '/^func ensureGlobalSettingsEnv/,/^}/p' internal/cli/update.go | grep -c 'userHomeDirFn()'
rm -rf "$SENT"
```

기대: (a) `before-nonempty=0`. (b) `compile-exit=0`. (c) 첫 grep `1`, `--- PASS: …` 한 행. (d) `diff` 무출력 + `diff-exit=0`. (e) 첫 grep `0`(프로세스 HOME 변조 금지), 둘째 grep `1`(이음매 경유).

**AC-UDS-013과의 차이 (부채 D2 해소 지점)**: 선행 AC는 운영자의 **실제** `~/.claude/hooks`를 스냅샷했고, 그 디렉터리가 없는 머신에서는 before/after가 모두 0행이라 삭제 방향에 대해 공허하게 통과했다. 이 AC는 (a)에서 sentinel 트리를 **판정 자신이 만들고 비어 있지 않음을 확인**하므로, 어떤 머신에서도 삭제 방향이 판별된다. 판별력이 코드에만 의존한다(§A.2).

#### AC-UGE-008 — 이음매를 우회하면 판정이 실패한다 (§A.4 변형 반증)

```bash
SENT=$(mktemp -d /tmp/uge-home-neg.XXXXXX)
mkdir -p "$SENT/.claude/hooks/moai"
printf 'probe\n' > "$SENT/.claude/hooks/moai/probe.txt"

# 변형: 테스트의 userHomeDirFn 주입을 제거한 사본을 overlay 로 덮는다.
# (프로덕션이 실제 HOME 을 해석하게 되어 sentinel 의 probe 를 지운다)
D=$(mktemp -d /tmp/uge-m3.XXXXXX)
sed '/userHomeDirFn = func/,+2d' internal/cli/update_home_radius_test.go > "$D/no_seam_test.go"
printf '{"Replace":{"%s/internal/cli/update_home_radius_test.go":"%s/no_seam_test.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" -c -o /tmp/uge-cli-neg.test ./internal/cli/; echo "compile-exit=$?"
HOME="$SENT" /tmp/uge-cli-neg.test -test.run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -test.v >/dev/null 2>&1

# probe 가 지워졌는가 = 판정이 삭제 방향을 실제로 잡아내는가
test -f "$SENT/.claude/hooks/moai/probe.txt"; echo "probe-survived=$?"
rm -rf "$SENT" "$D"
```

기대: `compile-exit=0`, `probe-survived=1` (파일이 **사라졌다** = 변형이 sentinel HOME을 건드렸다 = AC-UGE-007의 `diff`가 이 변형을 잡아낸다).

`probe-survived=0`이면 변형이 sentinel을 건드리지 못했다는 뜻이고, 그러면 AC-UGE-007의 `diff`도 이 방향을 판별하지 못한다 → **AC-UGE-007을 PASS로 인정하지 않는다**(REQ-UGE-008).

> 위 `sed` 삭제 범위는 run-phase가 작성할 실제 주입 코드 형태에 맞춰 조정한다. 조정하더라도 **관측 대상은 동일**하다: 이음매를 뺀 변형이 sentinel의 probe를 삭제해야 한다.

---

### M4 — 사용자 영역 보존 가드 확장 (REQ-UGE-009 ~ 013)

#### AC-UGE-009 — 보존 가드가 `runCleanReinstall`을 구동한다

```bash
# (a) 존재 grep (§A.1) — 가드 테스트명은 run-phase 가 확정하되 아래 토큰을 포함해야 한다
grep -rn 'func TestCleanReinstall_PreservesUserArea' internal/cli/

# (b) PASS 행 관측 (§A.1)
go test -run 'TestCleanReinstall_PreservesUserArea' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -E '^--- (PASS|FAIL): TestCleanReinstall_PreservesUserArea'

# (c) 가드 스코프 커버리지 — 이 가드가 실제로 그 함수를 지나는가
go test -run 'TestCleanReinstall_PreservesUserArea' -covermode=set \
  -coverpkg=./internal/cli/...,./internal/cli/update/... \
  -coverprofile=/tmp/uge-m4a.out -count=1 ./internal/cli/ >/dev/null
go tool cover -func=/tmp/uge-m4a.out | grep 'runCleanReinstall'

# (d) 사용자 영역 어서션이 실제로 그 안에 있는가 (커버리지만으로는 부족)
sed -n '/func TestCleanReinstall_PreservesUserArea/,/^}/p' internal/cli/*_test.go \
  | grep -cE 'snapshotDir|harness'
```

기대: (a) 한 행. (b) `--- PASS: …`. (c) `runCleanReinstall`의 백분율이 **0.0%보다 크다** (baseline은 §B의 `0.0%` — 가드 스코프 한정 수치). (d) `≥2` — 스냅샷 비교와 사용자 소유 경로가 모두 등장.

**(d)가 필요한 이유**: 커버리지가 올랐다는 것만으로는 "그 함수를 지났다"까지만 증명하고 "사용자 영역 보존을 주장했다"는 증명하지 못한다. 도달성과 어서션은 다른 것이다(§1.3 정정과 같은 계열의 함정).

#### AC-UGE-010 — 보존 가드가 `backup.BackupMoaiConfig`를 구동한다

```bash
grep -rn 'func TestBackupMoaiConfig_PreservesUserArea' internal/cli/
go test -run 'TestBackupMoaiConfig_PreservesUserArea' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -E '^--- (PASS|FAIL): TestBackupMoaiConfig_PreservesUserArea'
go test -run 'TestBackupMoaiConfig_PreservesUserArea' -covermode=set \
  -coverpkg=./internal/cli/...,./internal/cli/update/... \
  -coverprofile=/tmp/uge-m4b.out -count=1 ./internal/cli/ >/dev/null
go tool cover -func=/tmp/uge-m4b.out | grep 'BackupMoaiConfig'
sed -n '/func TestBackupMoaiConfig_PreservesUserArea/,/^}/p' internal/cli/*_test.go \
  | grep -cE 'snapshotDir|harness'
```

기대: 존재 grep 한 행, `--- PASS: …`, `BackupMoaiConfig` 백분율 **> 0.0%** (baseline §B `0.0%`), 마지막 grep `≥2`.

#### AC-UGE-011 — `MigrateLegacyMemoryDir`의 파괴적 분기 두 갈래가 결정적으로 도달된다

```bash
# (a) 두 서브테스트가 실재한다 — 이름에 분기 정체성이 드러나야 한다
grep -cE 't\.Run\("(rename_when_state_absent|backup_then_remove_when_both_exist)"' internal/cli/*_test.go

# (b) 두 서브테스트 각각의 PASS 행 관측 (§A.1)
go test -run 'TestMigrateLegacyMemoryDir_PreservesUserArea' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -E '^\s+--- (PASS|FAIL): .*/(rename_when_state_absent|backup_then_remove_when_both_exist)'

# (c) 커버리지가 baseline 26.9% 를 넘는다
go test -run 'TestMigrateLegacyMemoryDir_PreservesUserArea' -covermode=set \
  -coverpkg=./internal/cli/...,./internal/cli/update/... \
  -coverprofile=/tmp/uge-m4c.out -count=1 ./internal/cli/ >/dev/null
go tool cover -func=/tmp/uge-m4c.out | grep 'MigrateLegacyMemoryDir'
```

기대: (a) `2`. (b) `--- PASS:` 두 행 (두 분기 모두). (c) 백분율이 **26.9%보다 크다**(§B baseline).

**실패 조건**: 한쪽 분기만 구동하면 (b)에 한 행만 나와 실패한다. 두 서브테스트가 있어도 실제로는 같은 분기를 두 번 지나면 (c)가 baseline을 넘지 못한다.

#### AC-UGE-012 — 글로브 확장 변형이 반증된다 (§A.4 변형 반증)

```bash
# 변형: CleanMoaiManagedPaths 의 `.claude/skills/moai*` 글로브를
#       `.claude/skills/*` 로 넓혀 사용자 소유 harness-* 스킬까지 삼키게 한다.
BASE=$(git merge-base origin/main HEAD)
D=$(mktemp -d /tmp/uge-m4d.XXXXXX)
sed 's|\.claude/skills/moai\*|.claude/skills/*|' internal/cli/update/deploy/deploy.go > "$D/widened.go"
diff <(cat internal/cli/update/deploy/deploy.go) "$D/widened.go" >/dev/null; echo "mutation-applied=$?"
printf '{"Replace":{"%s/internal/cli/update/deploy/deploy.go":"%s/widened.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -cE '^(--- FAIL|FAIL)'
rm -rf "$D"
```

기대: `mutation-applied=1` (변형이 실제로 파일을 바꿨다 — `0`이면 sed가 아무것도 못 바꾼 것이므로 이 반증 자체가 공허하다), 그리고 FAIL 계수 `≥1`.

**이 AC가 메우는 공백**: 선행 SPEC은 `.claude/skills/harness-ios-patterns/`가 위험하다는 주장을 **글로브 범위를 읽는 것만으로** 세웠고, 글로브를 실제로 넓혔을 때 가드가 실패하는지는 관측하지 않았다(REQ-UGE-012).

#### AC-UGE-013 — 템플릿 중립성 (§A.5 보존 가드, NFR-UGE-002)

```bash
BASE=$(git merge-base origin/main HEAD)
git diff --name-only "$BASE"..HEAD -- internal/template/templates/ | wc -l
git status --porcelain -- internal/template/templates/ | wc -l
```

기대: 둘 다 `0`.

**이 값을 움직이는 것**: 이 SPEC의 커밋이 템플릿 자산을 하나라도 건드리면 첫 명령이 `≥1`이 된다. `merge-base`로 기준을 계산하므로 `origin/main` 머지로 들어온 타인의 템플릿 변경은 계수에 들어오지 않는다(§A.3).

#### AC-UGE-014 — 회귀 없음 · 크로스 플랫폼 (§A.5 보존 가드, NFR-UGE-003/004)

```bash
go build ./...; echo "build-exit=$?"
GOOS=windows GOARCH=amd64 go build ./...; echo "win-build-exit=$?"
go test -count=1 ./... 2>&1 | grep -cE '^(FAIL|--- FAIL)'
golangci-lint run --timeout=3m 2>&1 | tail -5
```

기대: `build-exit=0`, `win-build-exit=0`, FAIL 계수 `0`, lint가 `0 issues.`(또는 변경 전 baseline 대비 신규 지적 0건).

#### AC-UGE-015 — 병렬 테스트 위험 부재 (NFR-UGE-001)

```bash
# 패키지 변수를 재할당하는 테스트 파일들에 t.Parallel() 이 없어야 한다
grep -c 't\.Parallel()' internal/cli/glm_tools_test.go
grep -rlE 'userHomeDirFn = |[a-zA-Z]*[Ss]tatFn = ' internal/cli/*_test.go \
  | xargs grep -c 't\.Parallel()'
```

기대: 첫 grep `0`(현행 baseline 유지), 둘째 — 나열된 모든 파일이 `0`.

**이 값을 움직이는 것**: 이 SPEC이 추가하는 테스트가 `t.Parallel()`을 부르면 즉시 `≥1`이 되어 실패한다. `userHomeDirFn`은 `glm_tools.go:123`의 패키지 변수이고 `setupToolsTestHome`이 47개 호출부에서 같은 변수를 재할당하므로, 어느 한쪽이 병렬화되는 순간 데이터 경쟁이 된다.

---

## §D 추적 매트릭스

| REQ | AC | REQ | AC |
|---|---|---|---|
| REQ-UGE-001 | AC-UGE-001 | REQ-UGE-008 | AC-UGE-008 |
| REQ-UGE-002 | AC-UGE-003 | REQ-UGE-009 | AC-UGE-009 |
| REQ-UGE-003 | AC-UGE-002 | REQ-UGE-010 | AC-UGE-010 |
| REQ-UGE-004 | AC-UGE-005 | REQ-UGE-011 | AC-UGE-011 |
| REQ-UGE-005 | AC-UGE-006 | REQ-UGE-012 | AC-UGE-012 |
| REQ-UGE-006 | AC-UGE-007 | REQ-UGE-013 | AC-UGE-003, 008, 012 |
| REQ-UGE-007 | AC-UGE-007 (a) | | |

NFR 매핑: NFR-UGE-001 → AC-UGE-015 · NFR-UGE-002 → AC-UGE-013 · NFR-UGE-003 → AC-UGE-002 (d), AC-UGE-014 · NFR-UGE-004 → AC-UGE-006, AC-UGE-014 · NFR-UGE-005 → 툴체인 강제(§spec.md NFR-UGE-005), AC 미매핑.

## §E Definition of Done

1. §C의 AC-UGE-001 ~ 015 전부 PASS이며, 각 판정 명령의 **실제 출력**이 progress.md에 인용되어 있다.
2. A.4가 요구하는 반증 3건(AC-UGE-003 overlay, AC-UGE-008 이음매 제거 변형, AC-UGE-012 글로브 확장 변형)이 각각 실패를 관측한 기록과 함께 남아 있다.
3. §B의 baseline이 리베이스 후 트리에서 **재측정**되어 있다(plan-phase 수치를 옮겨 적은 것은 baseline이 아니다).
4. AC-UGE-002의 Windows 런타임 통과가 CI 결과로 인용되어 있거나, 인용 불가 시 Gaps에 명시되어 있다.
