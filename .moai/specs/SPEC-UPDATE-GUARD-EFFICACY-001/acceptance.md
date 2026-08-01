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
| `golangci-lint` | `golangci-lint run --timeout=5m` | `0 issues.` (exit 0) |
| `merge-base` | `git merge-base origin/main HEAD` | `83610e03e` (PR #1275 머지 후 `origin/main` 반영 완료) |
| base의 M6 보고 문구 | `git show "$BASE":internal/cli/update_preserve_inventory.go \| grep -c 'restored %d/%d'` | `3` |
| `deploy.go`의 실제 글로브 표현 | `grep -n 'SkillsSubdir' internal/cli/update/deploy/deploy.go` | `filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*")` — 리터럴 `.claude/skills/moai*`는 **존재하지 않음**(`grep -c` → `0`) |
| 세 HOME 호출부가 먹이는 대상 | `grep -A4 'homeDir, _ := userHomeDir()' <파일>` | `detectGoBinPathForUpdate(homeDir)` + `template.WithHomeDir(homeDir)` — **삭제 없음** |
| 백업 회전의 보존 슬라이스 | `grep -n 'backups\[' internal/cli/update/backup/backup.go` | `backups[:len(backups)-keepCount]` (오래된 초과분 삭제) |

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

# (0) base 위생 검사 — 이 검사 없이는 반증이 교란된다 (아래 설명)
git show "$BASE":internal/cli/update_preserve_inventory.go | grep -c 'restored %d/%d'

D=$(mktemp -d /tmp/uge-m1.XXXXXX)
git show "$BASE":internal/cli/update_preserve_inventory.go > "$D/reverted.go"
printf '{"Replace":{"%s/internal/cli/update_preserve_inventory.go":"%s/reverted.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" \
  -run 'TestMergeBackPreserveInventory_PartialRestore' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -cE '^(--- FAIL|FAIL)'
rm -rf "$D"
```

기대: (0) **반드시 `3`**. 그리고 FAIL 계수 `≥1` (변경 전 코드에는 이음매가 없으므로 주입이 컴파일되지 않거나 분기가 도달되지 않아 실패). FAIL 계수가 `0`이면 가드가 M1 변경에 의존하지 않는다는 뜻이므로 AC 실패다.

**(0)이 없으면 반증이 교란된다 — 실측으로 확인한 함정.** 이 SPEC의 브랜치가 `origin/main`을 머지하기 전에는 `merge-base`가 `a64548a2a`였고, 그 시점에 다음이 관측되었다.

```
$ git merge-base origin/main HEAD           # → a64548a2a
$ git show a64548a2a:internal/cli/update_preserve_inventory.go | grep -c 'restored %d/%d'
0
```

그 base에는 M6의 보고 문구 자체가 없으므로 overlay는 **이음매가 없어서가 아니라 보고 문구가 없어서** FAIL한다. 두 원인이 같은 FAIL로 보이므로 반증이 무엇을 증명했는지 알 수 없다. (0)이 `3`을 요구하면 base가 M6를 포함함이 보장되고, 남은 FAIL 원인은 이음매 부재뿐이다.

**(0)이 `3`이 아니면 AC를 실패로 판정하고 크게 알린다** — 조용히 넘어가면 교란된 반증을 통과로 오독한다. 이 경우 base가 아직 M6를 담고 있지 않다는 뜻이므로 plan.md §C1의 머지/리베이스를 먼저 수행한다.

> 머지 후 이 트리에서 재측정: `merge-base` = `83610e03e`, `(0)` = `3`. 교란은 해소되었으나 base는 다시 움직일 수 있으므로 가드는 남긴다.

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

#### AC-UGE-006 — 주입된 홈이 렌더링 입력까지 도달한다 (REQ-UGE-005)

이 AC는 **의도된 변화를 직접 관측**한다. "전체 스위트가 여전히 green"은 판정이 아니다 — 이 변경이 아무것도 바꾸지 않았을 때도, 정확히 의도대로 바꿨을 때도 똑같이 green이기 때문에 두 경우를 구별하지 못한다.

관측 방법은 **호출 계수 스파이**다. 렌더 산출물을 어서트하지 않는 이유는 §A.2다 — `gobin.Detect`는 `GOBIN`/`GOPATH`가 설정된 머신에서 `homeDir`을 쓰지 않고 조기 반환하므로, 렌더된 경로를 어서트하면 판별력이 운영자 환경 변수에 좌우된다. 스파이 계수는 그런 의존이 없다.

```bash
# (a) 존재 grep (§A.1)
grep -rn 'func TestUpdateSubsystem_HomeSeamReach' internal/cli/

# (b) PASS 행 관측 (§A.1)
go test -run 'TestUpdateSubsystem_HomeSeamReach' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -E '^--- (PASS|FAIL): TestUpdateSubsystem_HomeSeamReach'

# (c) 스파이가 실제 계수를 어서트하는가 (단순 non-nil 확인이 아님)
sed -n '/func TestUpdateSubsystem_HomeSeamReach/,/^}/p' internal/cli/*_test.go \
  | grep -cE 'if .*calls|want.*calls|calls !='
```

기대: (a) 한 행. (b) `--- PASS: …`. (c) `≥1`.

가드는 `userHomeDirFn`을 **계수하는 스파이**로 주입하고 세 호출부에 도달하는 경로를 구동한 뒤, 스파이 계수가 기대치 이상임을 어서트한다. 계수를 어서트하는 것이 핵심이다 — "주입해도 안 터지더라"는 관측은 도달을 증명하지 않는다.

#### AC-UGE-006F — 세 호출부를 되돌리면 도달 계수가 떨어진다 (§A.4 반증, REQ-UGE-013)

```bash
D=$(mktemp -d /tmp/uge-m2f.XXXXXX)
for f in update_clean_install.go update_template_sync.go; do
  sed 's|homeDir, _ := userHomeDirFn()|homeDir, _ := userHomeDir()|g' \
    "internal/cli/$f" > "$D/$f"
  # 변형이 실제로 적용되었는지 확인 (§G AP-6 — 공허한 변형 금지)
  diff "internal/cli/$f" "$D/$f" >/dev/null; echo "$f mutation-applied=$?"
done
printf '{"Replace":{"%s/internal/cli/update_clean_install.go":"%s/update_clean_install.go","%s/internal/cli/update_template_sync.go":"%s/update_template_sync.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" -run 'TestUpdateSubsystem_HomeSeamReach' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -cE '^(--- FAIL|FAIL)'
rm -rf "$D"
```

기대: 두 파일 모두 `mutation-applied=1`, FAIL 계수 `≥1`.

**이 반증이 증명하는 것**: 가드가 M2 변경에 의존한다. 되돌린 코드에서는 세 호출부가 프로세스 `$HOME`을 읽으므로 스파이가 그만큼 덜 불리고 계수 어서션이 깨진다. `mutation-applied=1` 확인이 없으면 sed가 아무것도 못 바꿔도 "FAIL 없음"을 통과로 오독할 수 있다.

#### AC-UGE-006R — 주입이 없을 때는 프로덕션 동작이 동일하다 (§A.5 보존 가드)

```bash
go test -count=1 ./internal/cli/... 2>&1 | tail -5; echo "cli-test-exit=${PIPESTATUS[0]}"
go vet ./...; echo "vet-exit=$?"
```

기대: `cli-test-exit=0`, `vet-exit=0`.

**AC-UGE-006과의 역할 분담**: 006은 주입이 **있을 때** 동작이 바뀌는 것을(의도된 변화), 006R은 주입이 **없을 때** 동작이 안 바뀌는 것을 확인한다. 006R 하나만으로는 REQ-UGE-005를 판정할 수 없다 — 그것이 최초 판의 결함이었다. `-run` 선택자를 쓰지 않으므로 0-매치 공허 모드는 없다(§A.5).

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

# (e) NFR-UGE-001 계열 (프로세스 전역 변조 금지) — 이음매로 격리했는가
#     ※ 최초 판은 이 항목을 NFR-UGE-002(템플릿 중립성)로 잘못 라벨했다. §D가 옳다.
grep -c 't.Setenv("HOME"' internal/cli/update_home_radius_test.go
sed -n '/^func ensureGlobalSettingsEnv/,/^}/p' internal/cli/update.go | grep -c 'userHomeDirFn()'
rm -rf "$SENT"
```

기대: (a) `before-nonempty=0`. (b) `compile-exit=0`. (c) 첫 grep `1`, `--- PASS: …` 한 행. (d) `diff` 무출력 + `diff-exit=0`. (e) 첫 grep `0`(프로세스 HOME 변조 금지), 둘째 grep `1`(이음매 경유).

**AC-UDS-013과의 차이 (부채 D2 해소 지점)**: 선행 AC는 운영자의 **실제** `~/.claude/hooks`를 스냅샷했고, 그 디렉터리가 없는 머신에서는 before/after가 모두 0행이라 삭제 방향에 대해 공허하게 통과했다. 이 AC는 (a)에서 sentinel 트리를 **판정 자신이 만들고 비어 있지 않음을 확인**하므로 판별력이 코드에만 의존한다(§A.2).

**sentinel이 비어 있지 않아야 하는 진짜 이유 (REQ-UGE-007의 근거 정정)**: 최초 판은 이 조건의 근거를 선행 AC-UDS-013의 논리("before가 비면 삭제 방향이 판별되지 않는다")에서 그대로 옮겨 적었다. 새 설계에서는 그 논리가 **성립하지 않는다** — 정상 경로에서 프로덕션은 주입된 픽스처만 건드리고 sentinel은 애초에 손대지 않으므로, sentinel이 비어 있든 아니든 정상 경로의 `diff`는 통과한다.

sentinel이 비어 있지 않아야 하는 이유는 **오직 카나리아 역할** 때문이다. AC-UGE-008이 이음매를 뺀 변형을 돌릴 때, 프로덕션은 프로세스 `$HOME`(= sentinel)로 해석해 `.claude/hooks/moai`를 지운다. 그때 **지워질 무언가가 거기 있어야** 그 사건이 관측된다. probe 파일이 없으면 변형이 아무것도 지우지 못해 카나리아가 침묵하고, AC-UGE-008이 공허하게 통과한다. 즉 (a)는 AC-UGE-007 자신을 위한 조건이 아니라 **AC-UGE-008을 비공허하게 만드는 조건**이다.

> 이 트리에서 실행해 확인함: `before-nonempty=0`, `compile-exit=0`, `--- PASS: TestEnsureGlobalSettingsEnv_HooksRemovalRadius (0.02s)`, `diff-exit=0`.

#### AC-UGE-008 — 이음매를 우회하면 판정이 실패한다 (§A.4 변형 반증)

```bash
SENT=$(mktemp -d /tmp/uge-home-neg.XXXXXX)
mkdir -p "$SENT/.claude/hooks/moai"
printf 'probe\n' > "$SENT/.claude/hooks/moai/probe.txt"

# 변형: 테스트의 userHomeDirFn 주입을 제거한 사본을 overlay 로 덮는다.
# (프로덕션이 실제 HOME 을 해석하게 되어 sentinel 의 probe 를 지운다)
D=$(mktemp -d /tmp/uge-m3.XXXXXX)
# 이음매 주입을 실제 HOME 해석으로 치환한다. 주입 행을 '삭제'하면 origFn 이
# 미사용이 되어 컴파일이 깨지므로, 치환이어야 한다.
sed 's|userHomeDirFn = func() (string, error) { return tmp, nil }|userHomeDirFn = userHomeDir|' \
  internal/cli/update_home_radius_test.go > "$D/no_seam_test.go"
diff internal/cli/update_home_radius_test.go "$D/no_seam_test.go" >/dev/null; echo "mutation-applied=$?"
printf '{"Replace":{"%s/internal/cli/update_home_radius_test.go":"%s/no_seam_test.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" -c -o /tmp/uge-cli-neg.test ./internal/cli/; echo "compile-exit=$?"
HOME="$SENT" /tmp/uge-cli-neg.test -test.run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -test.v >/dev/null 2>&1

# probe 가 지워졌는가 = 판정이 삭제 방향을 실제로 잡아내는가
test -f "$SENT/.claude/hooks/moai/probe.txt"; echo "probe-survived=$?"
rm -rf "$SENT" "$D"
```

기대: `mutation-applied=1`, `compile-exit=0`, `probe-survived=1` (파일이 **사라졌다** = 변형이 sentinel HOME을 건드렸다 = AC-UGE-007의 `diff`가 이 변형을 잡아낸다).

`probe-survived=0`이면 변형이 sentinel을 건드리지 못했다는 뜻이고, 그러면 AC-UGE-007의 `diff`도 이 방향을 판별하지 못한다 → **AC-UGE-007을 PASS로 인정하지 않는다**(REQ-UGE-008).

> **이 트리에서 실행해 확인함**: `mutation-applied=1`, `compile-exit=0`, `probe-survived=1`. 카나리아가 실제로 울린다.
>
> `sed` 치환 문자열은 현재 `update_home_radius_test.go`의 주입 형태(`userHomeDirFn = func() (string, error) { return tmp, nil }`)에 정확히 맞춰져 있다. run-phase가 주입 코드를 바꾸면 치환 문자열도 맞춰 조정하되, `mutation-applied=1` 확인은 반드시 남긴다 — 그것이 조정 실패를 잡는 장치다. **관측 대상은 불변**이다: 이음매를 뺀 변형이 sentinel의 probe를 삭제해야 한다.

---

### M4 — 사용자 영역 보존 가드 확장 (REQ-UGE-009 ~ 013)

#### AC-UGE-009 — 보존 가드가 `runCleanReinstall`을 구동한다

```bash
# (a) 존재 grep (§A.1) — 가드 테스트명은 run-phase 가 확정하되 아래 토큰을 포함해야 한다
grep -rn 'func TestCleanReinstall_PreservesUserArea' internal/cli/

# (b) PASS 행 관측 (§A.1)
go test -run 'TestCleanReinstall_PreservesUserArea' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -E '^--- (PASS|FAIL): TestCleanReinstall_PreservesUserArea'

# (c) 구조적 도달 — before 스냅샷 → 프로덕션 호출 → after 스냅샷 → 동등성 어서션
#     이 네 단계가 이 순서로 나타나는지 확인한다 (토큰 개수 세기가 아님)
sed -n '/func TestCleanReinstall_PreservesUserArea/,/^}/p' internal/cli/*_test.go \
  | grep -nE 'snapshotDir\(|runCleanReinstall\(|mapsEqual\(|t\.Errorf' \
  | awk -F: '{print $1" "$2}'
```

기대: (a) 한 행. (b) `--- PASS: …`. (c) 출력이 다음 순서를 만족한다 — `snapshotDir(` 등장 → `runCleanReinstall(` 등장 → `snapshotDir(` 재등장 → `mapsEqual(`(또는 동등 비교) → `t.Errorf`. 즉 **스냅샷이 프로덕션 호출을 사이에 두고 두 번 찍히고 비교된다.**

**(c)가 토큰 계수가 아니라 순서 검사인 이유**: 최초 판은 `grep -cE 'snapshotDir|harness' >= 2`였는데, 그것은 토큰이 **등장**했다는 것만 세고 스냅샷이 실제로 **비교**되었는지는 보지 않는다. 두 스냅샷을 찍고 버려도 통과한다. §1.3이 지목한 진짜 공백은 "바이트 동일성 어서션의 도달 범위"이므로, 판정도 어서션의 **구조**를 봐야 한다. 도달성과 어서션은 다른 것이다.

**커버리지 비교를 뺀 이유**: 최초 판의 `(c)`는 `-run 'TestCleanReinstall_PreservesUserArea'` 스코프의 `runCleanReinstall` 커버리지를 §B의 `0.0%` baseline과 비교했다. 그러나 §B의 `0.0%`는 **다른 선택자**(`-run 'TestMoaiUpdate_PreservesUserArea'`) 아래에서 측정된 값이다. 선택자가 다르면 두 수치는 비교 대상이 아니고, 새 선택자 아래 값이 0보다 크다는 것은 (b)의 PASS가 이미 함의한다 — 즉 아무것도 새로 증명하지 못한다. 진짜 증명은 아래 반증 AC가 한다.

#### AC-UGE-009F — `runCleanReinstall` 가드가 사용자 경로 삭제를 잡아낸다 (§A.4 반증, REQ-UGE-013)

`runCleanReinstall`의 파괴적 표면은 `scanDeprecatedPaths`가 만든 목록에 대한 `os.RemoveAll(abs)` 루프다(`update_clean_install.go`). 그 목록에 사용자 소유 경로를 넣으면 가드가 실패해야 한다.

```bash
D=$(mktemp -d /tmp/uge-m4af.XXXXXX)
# scanDeprecatedPaths 의 반환 직전에 사용자 소유 경로를 주입한다.
# (루프 밖 append 이므로 Lstat 존재 검사를 거치지 않는다)
perl -0pe 's/(\treturn found, nil\n\})/\tfound = append(found, ".moai\/harness")\n$1/' \
  internal/cli/update_cleanup.go > "$D/update_cleanup.go"
diff internal/cli/update_cleanup.go "$D/update_cleanup.go" >/dev/null; echo "mutation-applied=$?"
printf '{"Replace":{"%s/internal/cli/update_cleanup.go":"%s/update_cleanup.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" -run 'TestCleanReinstall_PreservesUserArea' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -cE '^(--- FAIL|FAIL)'
rm -rf "$D"
```

기대: `mutation-applied=1`, FAIL 계수 `≥1`.

> **이 트리에서 변형 적용까지 실행해 확인함**: `mutation-applied=1`, diff는 정확히 한 줄 삽입이며 위치는 `scanDeprecatedPaths` 본문 끝(`update_cleanup.go:145`, `return found, nil` 직전)이다.
>
> ```
> 144a145
> > 	found = append(found, ".moai/harness")
> ```
>
> **주의 (§G AP-6)**: `perl` 치환은 `\treturn found, nil\n}` 패턴에 걸리므로 이론상 동일 패턴의 다른 함수에도 걸릴 수 있다. 현재 트리에서는 한 곳뿐임을 확인했으나, run-phase는 `mutation-applied=1`에 더해 **diff 본문을 기록에 인용**해 의도한 함수에 들어갔는지 보인다. `mutation-applied=1`은 "무언가 바뀌었다"만 보증한다.
>
> 가드 FAIL 관측은 run-phase 몫이다 — `TestCleanReinstall_PreservesUserArea`가 아직 없으므로 plan-phase에서는 변형 적용까지만 확인했다.

#### AC-UGE-010 — 백업 서브시스템의 두 파괴적 표면에 보존 가드가 있다 (REQ-UGE-010)

> **재정박 근거 (spec.md REQ-UGE-010 주석 참조).** 최초 판은 `BackupMoaiConfig` 위에 "사용자 영역 불변" 가드를 세우려 했으나, 그 함수의 주 경로는 **순수 복사**라 그런 가드는 구조적으로 거의 공허하다 — 아무것도 지우지 않는 함수가 아무것도 지우지 않았음을 확인하는 꼴이고, 반증 변형을 정직하게 만들 수도 없다. 백업 서브시스템에서 실제로 지우는 곳은 아래 둘이다.

```bash
# (a) 존재 grep (§A.1)
grep -rn 'func TestBackupSubsystem_DestructiveSurfaces' internal/cli/

# (b) 두 서브테스트의 PASS 행 관측 (§A.1)
go test -run 'TestBackupSubsystem_DestructiveSurfaces' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -E '^\s+--- (PASS|FAIL): .*/(rollback_removes_only_own_backup_dir|rotation_keeps_newest)'

# (c) 회전 가드가 '어느 것이 살아남았는지'를 어서트하는가 (개수만 세지 않는가)
sed -n '/t.Run("rotation_keeps_newest"/,/^	})/p' internal/cli/*_test.go \
  | grep -cE 'newest|survivor|want.*\[\]string|Contains'
```

기대: (a) 한 행. (b) `--- PASS:` **두 행**. (c) `≥1`.

**두 표면**:
- `rollback_removes_only_own_backup_dir` — `BackupMoaiConfig`가 실패해 `os.RemoveAll(backupDir)` 롤백을 탈 때, 그 실행이 만든 백업 디렉터리만 사라지고 형제 백업과 사용자 영역은 남는다.
- `rotation_keeps_newest` — `CleanupOldBackups(projectRoot, keepCount)`가 **가장 최신 `keepCount`개를 남기고** 가장 오래된 초과분만 지운다.

**(c)가 개수 세기를 금지하는 이유**: "N개 남았다"는 어서션은 **어느 N개**인지 구별하지 못하므로, 아래 반증(최신/최오래 반전)을 통과시킨다. 살아남은 것의 정체성을 어서트해야 한다.

#### AC-UGE-010F — 회전 슬라이스를 역사적 버그 형태로 반전하면 가드가 실패한다 (§A.4 반증, REQ-UGE-013)

`backup.go`는 자기 주석에 과거 실제 버그를 기록해 두었다: *"A prior revision deleted `backups[keepCount:]` — the newest — which destroyed the most recent restore points on every rotation."* 그 반전이 그대로 카나리아다.

```bash
D=$(mktemp -d /tmp/uge-m4bf.XXXXXX)
sed 's|backups\[:len(backups)-keepCount\]|backups[keepCount:]|' \
  internal/cli/update/backup/backup.go > "$D/backup.go"
diff internal/cli/update/backup/backup.go "$D/backup.go" >/dev/null; echo "mutation-applied=$?"
printf '{"Replace":{"%s/internal/cli/update/backup/backup.go":"%s/backup.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" -run 'TestBackupSubsystem_DestructiveSurfaces' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -cE '^(--- FAIL|FAIL)'
rm -rf "$D"
```

기대: `mutation-applied=1`, FAIL 계수 `≥1`.

**이 반증의 가치**: 가상의 변형이 아니라 **이 저장소에서 실제로 발생했던 회귀**를 재현한다. 가드가 이것을 잡지 못하면, 같은 버그가 다시 들어와도 잡지 못한다는 뜻이다.

> **이 트리에서 변형 적용·컴파일까지 실행해 확인함**: `mutation-applied=1`, 변경 행은 `backup.go:257`, 변형 트리 `go build ./...` exit 0(컴파일되는 변형이므로 가드가 실제로 실행된다).
>
> ```
> 257c257
> < 	for _, backupName := range backups[:len(backups)-keepCount] {
> ---
> > 	for _, backupName := range backups[keepCount:] {
> ```
>
> 가드 FAIL 관측은 run-phase 몫이다 — `TestBackupSubsystem_DestructiveSurfaces`가 아직 없다.

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

#### AC-UGE-011F — 백업 단계를 제거하면 가드가 실패한다 (§A.4 반증, REQ-UGE-013)

`MigrateLegacyMemoryDir`의 both-exist 분기는 **백업한 뒤에** 레거시 디렉터리를 지운다(REQ-UDS-008). 백업 단계를 건너뛰게 만들면, 사용자 데이터가 복구 불가능하게 사라지므로 가드가 실패해야 한다.

```bash
D=$(mktemp -d /tmp/uge-m4cf.XXXXXX)
# backupLegacyMemoryDir 호출을 무력화한다 — 백업 없이 곧장 삭제로 간다.
sed 's|backupDir, err := backupLegacyMemoryDir(projectRoot, legacyDir)|backupDir, err := "", error(nil)|' \
  internal/cli/update/deploy/deploy.go > "$D/deploy.go"
diff internal/cli/update/deploy/deploy.go "$D/deploy.go" >/dev/null; echo "mutation-applied=$?"
printf '{"Replace":{"%s/internal/cli/update/deploy/deploy.go":"%s/deploy.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" -run 'TestMigrateLegacyMemoryDir_PreservesUserArea' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -cE '^(--- FAIL|FAIL)'
rm -rf "$D"
```

기대: `mutation-applied=1`, FAIL 계수 `≥1`.

**가드가 무엇을 어서트해야 하는가**: both-exist 분기 실행 후 **백업본이 실재하고 원본과 내용이 같아야** 한다. "레거시 디렉터리가 사라졌다"만 어서트하면 위 변형을 통과시킨다 — 변형도 레거시를 지우기 때문이다.

> 이 변형은 `err` 타입 추론 때문에 컴파일이 깨질 수 있다. 깨지면 run-phase는 컴파일되는 등가 변형(예: `backupLegacyMemoryDir`가 즉시 `"", nil`을 반환하도록 본문 치환)으로 바꾸되, **관측 대상은 불변**이다: 백업 없이 삭제되면 가드가 실패해야 한다. `mutation-applied=1`과 `compile-exit=0`을 함께 기록한다.

#### AC-UGE-012 — 글로브 확장 변형이 반증된다 (§A.4 변형 반증)

```bash
# 변형: CleanMoaiManagedPaths 의 skills 글로브를 넓혀 사용자 소유 harness-* 스킬까지
#       삼키게 한다. 코드의 실제 표현은 리터럴 경로가 아니라 defs 상수 조합이다.
D=$(mktemp -d /tmp/uge-m4d.XXXXXX)
sed 's|defs\.SkillsSubdir, "moai\*"|defs.SkillsSubdir, "*"|g' \
  internal/cli/update/deploy/deploy.go > "$D/widened.go"
diff internal/cli/update/deploy/deploy.go "$D/widened.go" >/dev/null; echo "mutation-applied=$?"
printf '{"Replace":{"%s/internal/cli/update/deploy/deploy.go":"%s/widened.go"}}\n' \
  "$(git rev-parse --show-toplevel)" "$D" > "$D/overlay.json"

go test -overlay="$D/overlay.json" -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -v ./internal/cli/ 2>&1 \
  | grep -cE '^(--- FAIL|FAIL)'
rm -rf "$D"
```

기대: `mutation-applied=1` (변형이 실제로 파일을 바꿨다 — `0`이면 sed가 아무것도 못 바꾼 것이므로 이 반증 자체가 공허하다), 그리고 FAIL 계수 `≥1`.

**이 AC가 메우는 공백**: 선행 SPEC은 `.claude/skills/harness-ios-patterns/`가 위험하다는 주장을 **글로브 범위를 읽는 것만으로** 세웠고, 글로브를 실제로 넓혔을 때 가드가 실패하는지는 관측하지 않았다(REQ-UGE-012).

**정정 이력 — 최초 판의 변형은 무동작(no-op)이었다.** 최초 판은 리터럴 `.claude/skills/moai*`를 겨냥했으나 코드에 그런 리터럴은 **존재하지 않는다**. 실제 표현은 `filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*")`(2행)이다. 실측:

```
$ grep -c '\.claude/skills/moai\*' internal/cli/update/deploy/deploy.go
0
$ sed 's|\.claude/skills/moai\*|.claude/skills/*|' deploy.go > widened.go
$ diff deploy.go widened.go >/dev/null; echo "mutation-applied=$?"
0        # 기대는 1 — sed 가 아무것도 바꾸지 못했다
```

즉 최초 판의 반증은 아무것도 반증하지 못한 채 통과할 수 있었다. `mutation-applied` 확인이 있었기 때문에 그 공허함이 드러났다 — 이것이 §G AP-6이 존재하는 이유다.

> **정정 후 이 트리에서 실행해 확인함**: `mutation-applied=1`, 변경된 행은 `deploy.go:53-54`, 가드 FAIL 계수 `4`. 반증이 실제로 작동한다.

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

기대: `build-exit=0`, `win-build-exit=0`, FAIL 계수 `0`, lint 출력이 **정확히 `0 issues.`**.

**lint 기대를 절대값으로 고정한 이유**: 최초 판은 "`0 issues.` 또는 변경 전 baseline 대비 신규 지적 0건"이라고 썼는데, §B에 lint baseline이 없어서 "신규 0건"을 판정할 기준이 존재하지 않았다 — 측정되지 않은 기준에 대한 상대 비교는 판정이 아니다. baseline을 실측했고 값이 `0 issues.`이므로(§B), 상대 비교를 버리고 절대값으로 고정한다.

> **이 트리에서 실측함**: `golangci-lint run --timeout=5m` → exit 0, `0 issues.`

#### AC-UGE-015 — 병렬 테스트 위험 부재 (NFR-UGE-001)

```bash
# 패키지 변수를 재할당하는 테스트 파일들에 t.Parallel() '호출'이 없어야 한다.
# 주석 줄(// ...)은 제외한다 — 아래 오탐 주석 참조.
grep -vE '^\s*//' internal/cli/glm_tools_test.go | grep -c 't\.Parallel()'

# xargs 를 쓰지 않는다 (아래 이식성 주석 참조)
grep -rlE 'userHomeDirFn = |[a-zA-Z]*[Ss]tatFn = ' internal/cli/*_test.go \
| while IFS= read -r f; do
    n=$(grep -vE '^\s*//' "$f" | grep -c 't\.Parallel()')
    printf '%s %s\n' "$f" "$n"
  done
```

기대: 첫 계수 `0`(현행 baseline 유지), 둘째 — 나열된 **모든** 파일의 계수가 `0`.

**주석 제외가 필요한 이유 — 이 판정을 실행해서 잡은 오탐.** 최초 판은 주석을 세지 않았고, 실행 결과 다음이 나왔다.

```
internal/cli/glm_tools_test.go 0
internal/cli/update_home_radius_test.go 1     # ← 위반처럼 보임
```

`1`의 정체는 `update_home_radius_test.go:17`의 **주석**이었다:

```go
// package-level variable, this test MUST NOT call t.Parallel().
```

즉 규율을 지키라고 적어 둔 주석이 규율 위반으로 계수되었다. 주석을 제외하면 두 파일 모두 `0`이다(재실행으로 확인). 이 오탐을 방치했다면 run-phase가 존재하지 않는 위반을 쫓았을 것이다.

**`xargs`를 쓰지 않는 이유 (이식성)**: 최초 판은 `... | xargs grep -c 't\.Parallel()'`였다. 입력이 비었을 때 BSD/macOS `xargs`는 명령을 실행하지 않지만 **GNU `xargs`(ubuntu CI)는 인자 없이 실행**하므로 `grep`이 stdin을 읽으려 멈춘다 — CI에서 판정이 행(hang)한다. `xargs -r`로도 고칠 수 있으나 `-r`은 GNU 확장이라 BSD에서 다시 문제가 되므로, `while read` 루프로 `xargs` 자체를 없앴다. 파일명당 계수를 함께 출력하므로 어느 파일이 위반인지도 드러난다.

> **미확인**: 이 판정을 GNU `xargs` 환경에서 직접 재현하지는 못했다(이 머신은 BSD `xargs`). 위 재작성은 `xargs`를 제거해 그 차이를 무의미하게 만드는 방식이므로, 재현 없이도 안전하다.

**이 값을 움직이는 것**: 이 SPEC이 추가하는 테스트가 `t.Parallel()`을 부르면 즉시 `≥1`이 되어 실패한다. `userHomeDirFn`은 `glm_tools.go:123`의 패키지 변수이고 `setupToolsTestHome`이 47개 호출부에서 같은 변수를 재할당하므로, 어느 한쪽이 병렬화되는 순간 데이터 경쟁이 된다.

---

## §D 추적 매트릭스

| REQ | AC | REQ | AC |
|---|---|---|---|
| REQ-UGE-001 | AC-UGE-001 | REQ-UGE-008 | AC-UGE-008 |
| REQ-UGE-002 | AC-UGE-003 | REQ-UGE-009 | AC-UGE-009, 009F |
| REQ-UGE-003 | AC-UGE-002 | REQ-UGE-010 | AC-UGE-010, 010F |
| REQ-UGE-004 | AC-UGE-005 | REQ-UGE-011 | AC-UGE-011, 011F |
| REQ-UGE-005 | AC-UGE-006, 006F, 006R | REQ-UGE-012 | AC-UGE-012 |
| REQ-UGE-006 | AC-UGE-007 | REQ-UGE-013 | AC-UGE-003, 006F, 008, 009F, 010F, 011F, 012 |
| REQ-UGE-007 | AC-UGE-007 (a) + AC-UGE-008 | | |

NFR 매핑: NFR-UGE-001 → AC-UGE-015, AC-UGE-007 (e) · NFR-UGE-002 → AC-UGE-013 · NFR-UGE-003 → AC-UGE-002 (d), AC-UGE-014 · NFR-UGE-004 → AC-UGE-006R, AC-UGE-014 · NFR-UGE-005 → 툴체인 강제(§spec.md NFR-UGE-005), AC 미매핑.

**반증 완전성 점검 (REQ-UGE-013)**: 이 SPEC이 새로 만들거나 고치는 가드는 여섯 개이며, 각각 자기 반증 AC를 가진다 — M1→003, M2→006F, M3→008, M4 reinstall→009F, M4 backup→010F, M4 migrate→011F. AC-UGE-012는 **기존** 가드(`TestMoaiUpdate_PreservesUserArea`)의 반증이므로 위 여섯과 별개다. 반증 없는 신규 가드는 **0개**다.

> 최초 판에서는 REQ-UGE-013이 003/008/012에만 매핑되어 M4의 신규 가드 세 개에 반증이 하나도 없었다. 006F/009F/010F/011F가 그 공백을 메운다.

## §E Definition of Done

1. §C의 AC 전부(001, 002, 003, 005, 006, 006F, 006R, 007, 008, 009, 009F, 010, 010F, 011, 011F, 012, 013, 014, 015 — 총 19개) PASS이며, 각 판정 명령의 **실제 출력**이 progress.md에 인용되어 있다.
2. A.4가 요구하는 반증 **7건**(003 / 006F / 008 / 009F / 010F / 011F / 012)이 각각 실패를 관측한 기록과 함께 남아 있다. 변형 반증은 `mutation-applied=1`도 함께 인용한다 — 그것 없이는 변형이 무동작이었는지 알 수 없다(§G AP-6).
3. §B의 baseline이 리베이스 후 트리에서 **재측정**되어 있다(plan-phase 수치를 옮겨 적은 것은 baseline이 아니다).
4. AC-UGE-002의 Windows 런타임 통과가 CI 결과로 인용되어 있거나, 인용 불가 시 Gaps에 명시되어 있다.
