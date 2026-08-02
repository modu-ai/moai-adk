# SPEC-PROFILE-MEMORY-001 — 인수 기준

## §A 판정 규칙

### A.1 공허한 GREEN 방지 (필독)

이 리포에서 실측된 공허-통과 함정을 회피하기 위해 아래를 강제한다.

1. **`go test -run <패턴>` 단독 금지.** 패턴이 0개 테스트에 매칭돼도 `exit 0` 이다. 모든 테스트 AC는 `-v` 와 함께 실행하고 `--- PASS: <정확한 테스트명>` 행을 **개수까지** 단언한다.
2. **토큰 grep 은 저작의 증거이지 도달의 증거가 아니다.** 동작을 주장하는 AC는 반드시 테스트로 구동한다. 특히 **함수 단위 판정은 호출자 배선을 판정하지 못한다** — 런치 경로 동작은 런치 레벨 AC로 판정한다(AC-PM-017 / AC-PM-018).
3. **절대 앵커 금지.** 줄 번호·커밋 SHA로 앵커하지 않는다. 함수명과 관측 가능한 동작으로만 앵커한다.
4. **가드 AC는 반증 가능해야 한다.** 가드를 제거하면 해당 테스트가 실패해야 하며, 그 왕복을 AC-PM-014(3)에서 실증한다.
5. **POSIX ERE 만 사용.** `\s` / `\d` 등 GNU·PCRE 확장 금지. 이 머신의 `grep` 은 ugrep 이라 `\s` 가 통과하지만 stock BSD `/usr/bin/grep` 에서는 리터럴 `s` 로 해석되어 0매치 → 거짓 실패가 된다. `[[:space:]]` / `[0-9]` 를 쓴다.

### A.2 공통 전제

모든 명령은 리포 루트 `/Users/goos/MoAI/moai-adk-go` 에서 실행한다.

---

## §B AC 행렬 (21건)

| AC | 대응 REQ | 성격 |
|----|----------|------|
| AC-PM-001 | REQ-PM-001, 002, 005(레거시 키 절) | 원장 왕복 — 레거시 키 보존 + 이중 기록 |
| AC-PM-002 | REQ-PM-003 (함수 층) | 프로젝트 스코프가 전역을 이긴다 |
| AC-PM-003 | REQ-PM-004 | 레거시 원장(projects 없음) 무변경 동작 |
| AC-PM-004 | REQ-PM-006 | 옵트아웃 스위치가 둘 다 끈다 |
| AC-PM-005 | REQ-PM-007 | project root 부재 → 전역 폴백, 무에러 |
| AC-PM-006 | REQ-PM-008 | 고아 항목 건너뛰고 전역으로 |
| AC-PM-007 | REQ-PM-009 | 경로 키 정규화 대칭 |
| AC-PM-008 | REQ-PM-011 | 미존재 디렉터리 기록 거부 |
| AC-PM-009 | REQ-PM-012, 013, 015 | 최초 `-p <new>` 런치에서 기록 생존 |
| AC-PM-010 | REQ-PM-014 | 기록 실패가 런치를 막지 않는다 (시임 주입) |
| AC-PM-011 | REQ-PM-016, 017, **018** | 탐지 술어 + 고지 함수 fresh/populated |
| AC-PM-012 | REQ-PM-019, 020, **018** | 자격증명 무복사 + 미배송 명령 무언급 + 플랫폼 중립 |
| AC-PM-013 | REQ-PM-010 | 웹 쓰기 시임 시그니처 무변경 |
| AC-PM-014 | REQ-PM-021, 022 | 테스트 샌드박스 가드 (반증 왕복 포함) |
| AC-PM-015 | REQ-PM-023 | 하드코딩 부재 |
| AC-PM-016 | 전체 | 회귀 없음 |
| AC-PM-017 | REQ-PM-003 (**런치 층**) | 런치가 실제로 project-scoped 해석을 쓴다 |
| AC-PM-018 | REQ-PM-016 (**런치 층**) | 런치 1회에 고지 정확히 1회 (이중 발화 탐지) |
| AC-PM-019 | REQ-PM-024 | 웹 콘솔 읽기 측 project-scoped |
| AC-PM-020 | REQ-PM-005 (원자성 절) | 원자 쓰기 보존 |
| AC-PM-021 | REQ-PM-021, 022 (오염 판정) | M5-2 후행 가드 |

REQ 커버리지: REQ-PM-001..024 전건이 최소 1개 AC에 라우팅된다. REQ-PM-018 은 AC-PM-011(술어 동작) + AC-PM-012(플랫폼 중립 부재-grep) 두 곳에 라우팅된다.

---

## §C 시나리오

### AC-PM-001 — 원장 왕복이 레거시 키를 보존하고 두 곳에 기록한다

**Given** `bypass: true`, `model: claude-opus-4-6`, `last_profile: old` 만 담긴 `launch.yaml` 이 샌드박스 base에 있고, 프로필 디렉터리 `new` 가 존재하며
**When** `RecordLastUsedProfileForProject(<projectRoot>, "new")` 를 호출하면
**Then** 재읽기한 원장에 `bypass`, `model` 이 원래 값 그대로 남아 있고, `last_profile == "new"` 이며, `projects[<정규화된 projectRoot>] == "new"` 이다.

```bash
go test ./internal/profile/ -run '^TestRecordForProject_PreservesLegacyKeys$' -v -count=1 2>&1 | tee /tmp/ac-pm-001.log
grep -c '^--- PASS: TestRecordForProject_PreservesLegacyKeys' /tmp/ac-pm-001.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-002 — 프로젝트 스코프 항목이 전역 `last_profile` 을 이긴다 (함수 층)

**Given** 원장이 `last_profile: global-one` 과 `projects: {<projA>: proj-one}` 을 함께 담고, 두 프로필 디렉터리가 모두 존재하며
**When** `ResolveLaunchProfileForProject(<projA>, "")` 를 호출하면
**Then** `"proj-one"` 을 반환한다. 같은 원장에 대해 `ResolveLaunchProfileForProject(<projB>, "")` 는 `"global-one"` 을 반환한다.

```bash
go test ./internal/profile/ -run '^TestResolveForProject_ProjectScopeWinsOverGlobal$' -v -count=1 2>&1 | tee /tmp/ac-pm-002.log
grep -c '^--- PASS: TestResolveForProject_ProjectScopeWinsOverGlobal' /tmp/ac-pm-002.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

**한계 명시**: 이 AC는 함수를 직접 호출하므로 **런처가 이 함수를 부르지 않아도 통과한다.** 런치 경로 배선은 AC-PM-017이 판정한다.

---

### AC-PM-003 — 레거시 원장은 업그레이드 전과 동일하게 동작한다

**Given** `projects:` 키가 전혀 없고 `last_profile: legacy` 만 있는 원장, 그리고 존재하는 `legacy` 디렉터리
**When** `ResolveLaunchProfileForProject(<anyRoot>, "")` 와 `ResolveLaunchProfile("")` 를 각각 호출하면
**Then** 둘 다 `"legacy"` 를 반환하고, 이후 원장 파일 바이트는 변하지 않았다(해석은 읽기 전용).

```bash
go test ./internal/profile/ -run '^TestResolveForProject_LegacyLedgerUnchanged$' -v -count=1 2>&1 | tee /tmp/ac-pm-003.log
grep -c '^--- PASS: TestResolveForProject_LegacyLedgerUnchanged' /tmp/ac-pm-003.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-004 — `MOAI_NO_PROFILE_FALLBACK=1` 이 두 조회를 모두 끈다

**Given** 원장에 유효한 프로젝트 스코프 항목과 유효한 `last_profile` 이 둘 다 있고, `MOAI_NO_PROFILE_FALLBACK=1` 이 설정된 상태에서
**When** `ResolveLaunchProfileForProject(<projA>, "")` 를 호출하면
**Then** `""` 를 반환한다 (프로젝트 항목이 있어도 무시).

```bash
go test ./internal/profile/ -run '^TestResolveForProject_OptOutDisablesBothLookups$' -v -count=1 2>&1 | tee /tmp/ac-pm-004.log
grep -c '^--- PASS: TestResolveForProject_OptOutDisablesBothLookups' /tmp/ac-pm-004.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-005 — project root 를 알 수 없으면 전역으로 폴백하고 에러를 내지 않는다

**Given** `projects` 항목이 있는 원장과 유효한 `last_profile`
**When** `ResolveLaunchProfileForProject("", "")` 와 `GetCurrentNameForProject("")` 를 호출하면
**Then** 각각 전역 `last_profile` 값과 그 이름을 반환하며, 두 함수 모두 단일 string 을 반환한다(에러 반환 타입 없음).
**And** `RecordLastUsedProfileForProject("", "some-existing")` 은 성공하고 원장의 `projects:` 맵을 **변경하지 않는다**.

```bash
go test ./internal/profile/ -run '^TestForProject_EmptyRootFallsBackToGlobal$' -v -count=1 2>&1 | tee /tmp/ac-pm-005.log
grep -c '^--- PASS: TestForProject_EmptyRootFallsBackToGlobal' /tmp/ac-pm-005.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

**도달성 주석**: 이 AC가 REQ-PM-007의 유일한 판정 경로다. 런치 경로는 root 획득 실패 시 조기 반환하므로(plan.md §C RESOLVED-3) 이 조건이 런치에서 발화하지 않는다.

---

### AC-PM-006 — 디렉터리가 사라진 프로젝트 항목은 건너뛰고 전역으로 간다

**Given** `projects: {<projA>: ghost}` 이지만 `ghost` 디렉터리는 없고, `last_profile: alive` 이며 `alive` 디렉터리는 존재할 때
**When** `ResolveLaunchProfileForProject(<projA>, "")` 를 호출하면
**Then** `"alive"` 를 반환한다.

```bash
go test ./internal/profile/ -run '^TestResolveForProject_StaleProjectEntrySkipped$' -v -count=1 2>&1 | tee /tmp/ac-pm-006.log
grep -c '^--- PASS: TestResolveForProject_StaleProjectEntrySkipped' /tmp/ac-pm-006.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-007 — 경로 키 정규화가 쓰기·읽기 양쪽에서 대칭이다

**Given** `t.TempDir()` 가 반환한 심링크 경유 경로(macOS `/var/folders/...`)를 project root 로 쓰고
**When** 그 경로로 기록한 뒤 **동일 경로로** 해석하면
**Then** 기록한 프로필 이름이 그대로 나온다.
**And** 심링크 해소된 형태(`/private/var/...`)로 해석해도 같은 이름이 나온다.

```bash
go test ./internal/profile/ -run '^TestProjectKey_NormalizationSymmetric$' -v -count=1 2>&1 | tee /tmp/ac-pm-007.log
grep -c '^--- PASS: TestProjectKey_NormalizationSymmetric' /tmp/ac-pm-007.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-008 — 존재하지 않는 디렉터리 이름의 기록을 거부한다

**Given** 샌드박스 base에 `nope` 디렉터리가 없는 상태에서
**When** `RecordLastUsedProfileForProject(<root>, "nope")` 를 호출하면
**Then** non-nil 에러를 반환하고, 원장 파일은 생성되지도 변경되지도 않는다(호출 전후 바이트 동일, 또는 호출 전 부재였다면 여전히 부재).

```bash
go test ./internal/profile/ -run '^TestRecordForProject_RejectsMissingDirectory$' -v -count=1 2>&1 | tee /tmp/ac-pm-008.log
grep -c '^--- PASS: TestRecordForProject_RejectsMissingDirectory' /tmp/ac-pm-008.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-009 — 최초 `-p <new>` 런치에서도 기록이 살아남는다 (순서 회귀 가드)

**Given** `brand-new` 프로필 디렉터리가 존재하지 않고, `launchClaudeFunc` 가 실제 exec 대신 호출을 기록하는 fake 로 대체되고, `findProjectRootFn` 이 샌드박스 프로젝트 루트로 고정된 상태에서
**When** `unifiedLaunch("brand-new", "claude", nil)` 을 실행하면
**Then** (a) `<base>/brand-new` 디렉터리가 생성돼 있고, (b) 원장의 `last_profile` 과 프로젝트 스코프 항목이 모두 `brand-new` 이며, (c) fake `launchClaude` 가 호출된 시점에는 이미 (a)와 (b)가 참이다.

이 AC가 **D4 순서 불변식의 판정자**다. 기록을 `EnsureDir` 앞으로 되돌리면 (b)가 실패한다.

```bash
go test ./internal/cli/ -run '^TestUnifiedLaunch_FirstTimeNewProfileIsRecorded$' -v -count=1 2>&1 | tee /tmp/ac-pm-009.log
grep -c '^--- PASS: TestUnifiedLaunch_FirstTimeNewProfileIsRecorded' /tmp/ac-pm-009.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-010 — 기록 실패가 런치를 막지 않는다 (시임 주입)

**Given** `recordLastProfileFn` 패키지 시임이 항상 `errors.New("injected ledger failure")` 를 반환하는 fake 로 교체되고, **`launcherStderr` 시임이 버퍼로 교체되고**(`plan.md` M4-2 / `design.md` §E.1 — 이 시임 없이는 (c)를 관측할 수 없다), `launchClaudeFunc` 가 호출을 기록하는 fake 이며, `findProjectRootFn` 이 샌드박스 루트로 고정되고, 대상 프로필 디렉터리는 정상 생성 가능한 상태에서
**When** `unifiedLaunch("work", "claude", nil)` 을 실행하면
**Then** (a) `unifiedLaunch` 는 nil 에러를 반환하고, (b) fake `launchClaude` 가 호출됐으며, (c) 주입된 stderr 버퍼에 경고 문자열이 담겨 있다.

**시임을 쓰는 이유 (plan.md §D7)**: 파일 권한으로 실패를 유도하면 root 실행에서 chmod 가 무시되고 Windows CI 에서 판정할 수 없다. "프로필 디렉터리 부재"는 4.5단계가 기록 직전에 생성하므로 성립하지 않는 전제이고, "base 읽기 전용" 단독은 `MkdirAll` 이 먼저 실패해 (b) 전제가 깨진다.

```bash
go test ./internal/cli/ -run '^TestUnifiedLaunch_RecordFailureDoesNotBlockLaunch$' -v -count=1 2>&1 | tee /tmp/ac-pm-010.log
grep -c '^--- PASS: TestUnifiedLaunch_RecordFailureDoesNotBlockLaunch' /tmp/ac-pm-010.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-011 — 탐지 술어와 고지 함수 (fresh / populated)

**Given** 프로필 A 디렉터리에는 `.claude.json` 이 없고, 프로필 B 디렉터리에는 있을 때
**When** 각각을 대상으로 고지 함수를 주입된 `io.Writer` 로 실행하면
**Then** A 에서는 출력 버퍼에 고지 문구가 나타나고, B 에서는 출력 버퍼가 비어 있다.
**And** `profile.HasClaudeConfig("A") == false`, `profile.HasClaudeConfig("B") == true`.
**And (REQ-PM-018 판정)** 프로필 A 디렉터리에 `.claude.json` **하나만** 생성하면 `HasClaudeConfig("A")` 가 `true` 로 바뀌고, 반대로 `.credentials.json` 만 생성한 경우에는 `false` 를 유지한다 — 즉 판정은 `.claude.json` 존재 **단독**으로 결정되며 다른 자격증명 파일에 반응하지 않는다.

```bash
go test ./internal/profile/ ./internal/cli/ -run '^(TestHasClaudeConfig_DecidesOnClaudeJSONAlone|TestFreshProfileNotice_WriterContent)$' -v -count=1 2>&1 | tee /tmp/ac-pm-011.log
grep -c -E '^--- PASS: (TestHasClaudeConfig_DecidesOnClaudeJSONAlone|TestFreshProfileNotice_WriterContent)' /tmp/ac-pm-011.log
```
**기대**: 종료 코드 0, grep 결과 `2` (두 테스트가 모두 존재하고 모두 PASS).

**한계 명시**: 고지 함수를 1회 호출해 1회 출력을 보는 것은 **이중 발화를 탐지하지 못한다.** 발화 횟수는 AC-PM-018이 런치 레벨에서 판정한다.

---

### AC-PM-012 — 자격증명 무복사, 미배송 명령 무언급, 플랫폼 중립

**Given** 구현이 완료된 트리에서
**When** 아래를 실행하면
**Then** 세 grep 모두 0 매치다.

```bash
# (1) 자격증명/계정 상태 파일을 복사·이동·시드하는 코드가 없다 (REQ-PM-019)
grep -rn -E '\.credentials\.json|claudeAiOauth' \
  internal/profile/ internal/cli/launcher.go | grep -v '_test.go' | tee /tmp/ac-pm-012a.log
test ! -s /tmp/ac-pm-012a.log && echo "PASS-a" || echo "FAIL-a"

# (2) 고지 문구가 미배송 명령을 광고하지 않는다 (REQ-PM-020)
grep -rn -E 'seed-auth|--seed|moai profile seed' \
  internal/profile/ internal/cli/launcher.go | tee /tmp/ac-pm-012b.log
test ! -s /tmp/ac-pm-012b.log && echo "PASS-b" || echo "FAIL-b"

# (3) 탐지 술어가 플랫폼별 자격증명 저장소를 조회하지 않는다 (REQ-PM-018)
grep -rn -E 'security[[:space:]]+find-generic-password|Keychain|keychain' \
  internal/profile/ | grep -v '_test.go' | tee /tmp/ac-pm-012c.log
test ! -s /tmp/ac-pm-012c.log && echo "PASS-c" || echo "FAIL-c"
```
**기대**: `PASS-a`, `PASS-b`, `PASS-c` 모두 출력.

**주의**: 이 AC는 부재-grep 이므로 그 자체로는 저작만 증명한다. REQ-PM-018의 동작 측 판정은 AC-PM-011의 `And (REQ-PM-018 판정)` 절이 담당한다. 정규식은 POSIX ERE 로만 작성했다(§A.1-5).

---

### AC-PM-013 — 웹 쓰기 시임 시그니처가 변하지 않았다

**Given** `internal/web/app.go`
**When** 시임 필드 선언을 확인하면
**Then** `recordLastProfile func(name string) error` 형태가 그대로 유지되고, `internal/web` 패키지 테스트가 시임 시드 변경 없이 통과한다.

```bash
grep -n 'recordLastProfile func(name string) error' internal/web/app.go
go test ./internal/web/ -count=1
```
**기대**: grep 이 정확히 1행 출력, `go test` 종료 코드 0.

---

### AC-PM-014 — 테스트 샌드박스 가드 (반증 왕복 포함)

**Given** `internal/profile/main_test.go` 와 `internal/web/main_test.go` 의 `TestMain` 샌드박스
**When** 아래 3단계를 실행하면
**Then** 가드가 반증 가능함이 실증된다.

```bash
# (1) 가드 테스트가 두 패키지 모두에 존재하고 통과한다
go test ./internal/profile/ ./internal/web/ -run '^TestProfileBaseDirIsSandboxed$' -v -count=1 2>&1 | tee /tmp/ac-pm-014.log
grep -c '^--- PASS: TestProfileBaseDirIsSandboxed' /tmp/ac-pm-014.log   # 기대: 2

# (2) 실제 홈 원장이 테스트 실행으로 변경되지 않는다 (internal/web 포함 — 감사 D9)
LEDGER="$HOME/.moai/claude-profiles/launch.yaml"
BEFORE=$( [ -f "$LEDGER" ] && shasum -a 256 "$LEDGER" | cut -d' ' -f1 || echo ABSENT )
go test ./internal/profile/ ./internal/cli/ ./internal/web/ -count=1 >/dev/null
AFTER=$( [ -f "$LEDGER" ] && shasum -a 256 "$LEDGER" | cut -d' ' -f1 || echo ABSENT )
[ "$BEFORE" = "$AFTER" ] && echo "PASS-untouched" || echo "FAIL-home-ledger-mutated"

# (3) 반증 왕복 — TestMain 의 sandboxProfileBaseDir() 호출을 임시 주석 처리하면
#     TestProfileBaseDirIsSandboxed 가 FAIL 해야 한다. 확인 후 즉시 원복한다.
```
**기대**: (1) `2`, (2) `PASS-untouched`, (3) 주석 처리 시 FAIL·원복 후 PASS 를 run-phase 증거로 기록.

---

### AC-PM-015 — 하드코딩이 없다

**Given** 구현 완료 트리
**When** 아래를 실행하면
**Then** 신규 YAML 키와 상태 파일명이 상수로 선언돼 있고, 리터럴 산재가 없다.

```bash
# 상수 선언 존재 (POSIX ERE — §A.1-5)
grep -n -E 'projectsKey[[:space:]]*=[[:space:]]*"projects"|claudeConfigStateFile[[:space:]]*=[[:space:]]*"\.claude\.json"' internal/profile/profile.go
# 상수 밖 리터럴 사용이 없다 (선언 행 제외 시 0 매치)
grep -rn -E '"projects"|"\.claude\.json"' internal/profile/ internal/cli/launcher.go \
  | grep -v '_test.go' \
  | grep -v -E 'projectsKey[[:space:]]*=|claudeConfigStateFile[[:space:]]*=' | tee /tmp/ac-pm-015.log
test ! -s /tmp/ac-pm-015.log && echo "PASS" || echo "FAIL"
```
**기대**: 첫 grep 이 2행 출력, 마지막이 `PASS`.

---

### AC-PM-016 — 회귀 없음

**Given** 전체 변경이 적용된 트리
**When** 전체 스위트와 린트를 실행하면
**Then** 둘 다 통과한다.

```bash
go build ./... && go vet ./...
go test ./... -count=1
golangci-lint run ./internal/profile/... ./internal/cli/... ./internal/web/...
```
**기대**: 세 명령 모두 종료 코드 0.

---

### AC-PM-017 — 런치가 실제로 project-scoped 해석을 사용한다 (감사 D1c)

**Given** 원장에 `last_profile: global-one` 과 `projects: {<projA>: proj-one}` 이 함께 있고 두 프로필 디렉터리가 모두 존재하며, `findProjectRootFn` 이 `<projA>` 로 고정되고, `launchClaudeFunc` 가 전달받은 `profileName` 을 기록하는 fake 인 상태에서
**When** `unifiedLaunch("", "claude", nil)` (bare 런치, `-p` 없음) 을 실행하면
**Then** fake `launchClaude` 가 받은 `profileName` 은 `"proj-one"` 이다 (`"global-one"` 이 아니다).

이 AC가 **iter-1 감사 D1 치명 결함의 판정자**다. 런처가 `ResolveLaunchProfile`(전역 전용)을 계속 부르면 `"global-one"` 이 전달되어 실패한다. AC-PM-002는 함수를 직접 호출하므로 이 결함을 탐지하지 못한다.

```bash
go test ./internal/cli/ -run '^TestUnifiedLaunch_UsesProjectScopedResolution$' -v -count=1 2>&1 | tee /tmp/ac-pm-017.log
grep -c '^--- PASS: TestUnifiedLaunch_UsesProjectScopedResolution' /tmp/ac-pm-017.log
```
**기대**: 종료 코드 0, grep 결과 `1`.

---

### AC-PM-018 — 런치 1회에 고지가 정확히 1회 발화한다 — 명시 `-p` 와 bare 런치 양쪽 (감사 D5b + iter-2 N1)

세 개의 케이스를 한 테스트 함수 안의 하위 케이스로 판정한다.

**케이스 A — 명시 `-p`, fresh (D5 이중 발화 판정자)**
**Given** `fresh` 프로필 디렉터리에 `.claude.json` 이 없고, 런처의 stderr 가 주입된 버퍼로 대체되고, `launchClaudeFunc` 가 fake 이며, `findProjectRootFn` 이 샌드박스 루트로 고정된 상태에서
**When** `unifiedLaunch("fresh", "claude", nil)` 을 **1회** 실행하면
**Then** 버퍼에서 고지 문구의 출현 횟수가 **정확히 1** 이다 (0도 2도 아니다).

`unifiedLaunchDefault` 4.5단계와 `launchClaudeDefault` 1단계 양쪽에서 고지하면 횟수가 2가 되어 실패한다.

**케이스 B — bare 런치가 `projects:` 를 통해 fresh 로 해석 (N1 판정자)**
**Given** 원장에 `projects: {<projA>: fresh}` 가 있고 `<base>/fresh` 디렉터리는 **존재하지만 `.claude.json` 이 없으며**, `findProjectRootFn` 이 `<projA>` 로 고정되고, stderr 가 주입 버퍼이며, `launchClaudeFunc` 가 fake 인 상태에서
**When** `unifiedLaunch("", "claude", nil)` (bare 런치, `-p` 없음) 을 **1회** 실행하면
**Then** 버퍼에서 고지 문구의 출현 횟수가 **정확히 1** 이다.

이 케이스가 **iter-2 N1 의 판정자**다. 4.5단계를 `originalProfile`(= `""`)로 게이트한 구현에서는 4.5단계가 통째로 건너뛰어지고 `launchClaudeDefault` 의 침묵 재호출만 남으므로 횟수가 **0** 이 되어 실패한다. 케이스 A 만으로는 이 결함을 탐지할 수 없다 — 명시 `-p` 경로에서는 두 변수가 같기 때문이다.

**케이스 C — populated 음성 대조군, 두 경로 모두 (감사 iter-3 NEW-4)**

C 는 **두 경로를 각각** 판정한다. 명시 `-p` 경로에서의 0은 게이트 변수가 무엇이든 성립하므로 수정의 정당성을 전혀 증명하지 못한다 — 더 강한 음성 대조군은 bare 경로 쪽이다.

- **C-1 (명시 `-p`)**: **Given** `.claude.json` 이 있는 `populated` 프로필 **When** `unifiedLaunch("populated", "claude", nil)` 을 실행하면 **Then** 출현 횟수가 **0** 이다.
- **C-2 (bare 해석, 강한 대조군)**: **Given** 원장에 `projects: {<projA>: populated}` 가 있고 `<base>/populated` 에 `.claude.json` 이 존재하며 `findProjectRootFn` 이 `<projA>` 로 고정된 상태에서 **When** `unifiedLaunch("", "claude", nil)` 을 실행하면 **Then** 출현 횟수가 **0** 이다.

C-2 가 B 와 짝을 이뤄 "bare 경로에서 고지가 fresh 에만 나온다"를 성립시킨다. C-1 만 있으면 넓힌 게이트가 populated 에까지 고지를 내는 과잉 발화를 잡지 못한다.

```bash
go test ./internal/cli/ -run '^TestUnifiedLaunch_FreshProfileNoticeEmittedExactlyOnce$' -v -count=1 2>&1 | tee /tmp/ac-pm-018.log
grep -c '^--- PASS: TestUnifiedLaunch_FreshProfileNoticeEmittedExactlyOnce' /tmp/ac-pm-018.log
```
**기대**: 종료 코드 0, grep 결과 `1` (A / B / C-1 / C-2 네 케이스를 모두 포함한 함수가 PASS).

**전제 시임**: 세 케이스 모두 `launcherStderr` (`io.Writer` 패키지 시임, `plan.md` M4-2 / `design.md` §E.1) 를 버퍼로 교체해 출력을 포착한다. 이 시임 없이는 런치 경로 출력을 관측할 수 없다.

---

### AC-PM-019 — 웹 콘솔 읽기 측이 project-scoped 다 (감사 D6)

**Given** 원장에 `last_profile: global-one` 과 `projects: {<projA>: proj-one}` 이 함께 있고 두 프로필 디렉터리가 모두 존재하며, `CLAUDE_CONFIG_DIR` 은 미설정일 때
**When** `profile.GetCurrentNameForProject(<projA>)` 를 호출하면
**Then** `"proj-one"` 을 반환한다.
**And** `internal/cli/web.go` `runWeb` 이 `web.Config{... ProfileName: ...}` 에 넘기는 값이 `profile.GetCurrentName()` (전역 전용) 이 아니라 project-scoped 변종임을 확인한다.

```bash
go test ./internal/profile/ -run '^TestGetCurrentNameForProject_ProjectScoped$' -v -count=1 2>&1 | tee /tmp/ac-pm-019.log
grep -c '^--- PASS: TestGetCurrentNameForProject_ProjectScoped' /tmp/ac-pm-019.log   # 기대: 1

# 배선 확인: 전역 전용 호출이 web.go 에 남아 있지 않다
grep -n 'GetCurrentNameForProject(projectRoot)' internal/cli/web.go        # 기대: 1행
grep -n 'ProfileName:[[:space:]]*profile\.GetCurrentName()' internal/cli/web.go | tee /tmp/ac-pm-019b.log
test ! -s /tmp/ac-pm-019b.log && echo "PASS-wired" || echo "FAIL-still-global"
```
**기대**: 테스트 grep `1`, 배선 grep 1행, `PASS-wired`.

---

### AC-PM-020 — 원자 쓰기가 보존됐다 (감사 D10)

**Given** 구현 완료 트리
**When** 아래를 실행하면
**Then** 기록기가 여전히 temp 파일 + `os.Rename` 경로를 쓰고, 실패 시 부분 상태를 남기지 않는다.

REQ-PM-005는 (a) 레거시 키 보존과 (b) 원자 쓰기 보존 두 절을 담는다. AC-PM-001은 (a)만 판정하므로, 구현자가 원자 쓰기를 평범한 `os.WriteFile` 로 바꿔도 AC-PM-001은 통과한다. 이 AC가 (b)를 담당한다.

```bash
# (1) 원자 쓰기 프리미티브가 기록 경로에 남아 있다
grep -n -E 'os\.CreateTemp|os\.Rename' internal/profile/profile.go | tee /tmp/ac-pm-020a.log
test -s /tmp/ac-pm-020a.log && echo "PASS-primitives" || echo "FAIL-atomicity-removed"

# (2) 기록 실패 시 부분 상태가 남지 않는다 (동작 판정 — 유도 방법은 아래 Given 에 고정)
go test ./internal/profile/ -run '^TestRecordForProject_NoPartialStateOnFailure$' -v -count=1 2>&1 | tee /tmp/ac-pm-020.log
grep -c '^--- PASS: TestRecordForProject_NoPartialStateOnFailure' /tmp/ac-pm-020.log
```
**기대**: (1) `PASS-primitives`, (2) grep 결과 `1`.

#### (2)의 실패 유도 방법 — 반드시 이 방법을 쓴다

**Given** 프로필 디렉터리 `ok` 가 정상 존재하고(가드 1~3을 모두 통과), 원장 경로 `<base>/launch.yaml` 이 파일이 아니라 **비어 있지 않은 디렉터리**인 상태에서
**When** `RecordLastUsedProfileForProject(<root>, "ok")` 를 호출하면
**Then** (a) non-nil 에러를 반환하고, (b) 실패는 `os.CreateTemp` **이후** 의 `os.Rename` 단계에서 발생하며, (c) `defer os.Remove(tmpName)` 이 temp 파일을 회수하여 `<base>` 에 `.launch-*.tmp` 잔여물이 **0개** 이고, (d) 원장 경로의 디렉터리 내용이 변하지 않았다.

**조건 (b)·(c)가 이 AC의 존재 이유다.** 이것이 없으면 구현자가 가드 3(디렉터리 부재)으로 실패를 유도할 수 있는데, 그 경로는 `os.CreateTemp` 에 **도달조차 하지 않으므로** 원자성에 대해 아무것도 판정하지 못하고 AC-PM-008 과 실질적으로 동일한 테스트가 된다. 조건 (c)는 temp 잔여물 개수로 "CreateTemp 이후였다"를 관측 가능하게 만든다.

**권한 조작을 쓰지 않는 이유**: 제약 C5(파일 권한 의존 테스트 금지)와 정면 충돌한다. 디렉터리-as-파일 기법은 권한을 건드리지 않으므로 root 실행에서도 성립한다.

**미검증 표기 (run-phase 확인 대상)**: 이 레시피는 POSIX `rename(2)` 의미론(비어 있지 않은 디렉터리를 대상으로 하는 rename 은 `ENOTEMPTY`/`EISDIR` 로 실패) 에 근거한 추론이며, **이 계획 단계에서 실행해 확인하지 않았다.** run-phase 에서 실제 실패 발생 여부를 먼저 확인하고, Windows 를 포함해 재현되지 않으면 `profile` 패키지에 rename 시임을 도입하는 대안으로 전환한다(제약 C6 때문에 시임은 차선으로 둔다).

---

### AC-PM-021 — 후행 샌드박스 가드가 오염을 실제로 판정한다 (감사 D12)

**Given** Go 는 테스트 파일을 사전순으로 실행하므로 `main_test.go` 의 가드는 `profile_test.go` 의 오염보다 **먼저** 실행되어 M5-2 누락을 탐지하지 못한다
**When** `profile_test.go` 보다 뒤에 정렬되는 `internal/profile/zz_sandbox_guard_test.go` 의 후행 가드를 실행하면
**Then** `profile.BaseDirOverride` 가 여전히 샌드박스 값(비어 있지 않고 실제 홈 경로가 아님)임을 단언하여 통과한다.
**And** M5-2(`TestGetBaseDir_Default` 의 `t.Cleanup` 복원)를 제거하면 이 가드가 FAIL 한다.

```bash
# (1) 후행 가드 파일이 profile_test.go 보다 뒤에 정렬된다
ls internal/profile/*_test.go | sort | tail -1   # 기대: .../zz_sandbox_guard_test.go

# (2) 패키지 전체 실행에서 후행 가드가 통과한다 (개별 -run 실행은 오염을 재현하지 못하므로 전체 실행으로 판정)
go test ./internal/profile/ -v -count=1 2>&1 | tee /tmp/ac-pm-021.log
grep -c '^--- PASS: TestSandboxSurvivesPackageRun' /tmp/ac-pm-021.log

# (3) 반증 왕복 — TestGetBaseDir_Default 의 t.Cleanup 복원을 임시 제거하면
#     TestSandboxSurvivesPackageRun 이 FAIL 해야 한다. 확인 후 즉시 원복한다.
```
**기대**: (1) `zz_sandbox_guard_test.go`, (2) grep 결과 `1`, (3) 제거 시 FAIL·원복 후 PASS 를 run-phase 증거로 기록.

---

## §D 완료 정의 (Definition of Done)

- [ ] AC-PM-001 ~ AC-PM-021 전항 실행, 각 명령의 실제 출력과 종료 코드를 증거로 기록
- [ ] AC-PM-014 (3) 및 AC-PM-021 (3) 반증 왕복 실증 완료 (FAIL 관측 → 원복 → PASS 관측)
- [ ] `spec.md` §C 범위 제외 위반 없음 — 특히 자격증명 시드 코드 부재(AC-PM-012)
- [ ] `internal/template/templates/` 무접촉 (`git diff --name-only` 로 확인)
- [x] plan.md §C RESOLVED-1 ~ RESOLVED-4 확인 (2026-08-02 사용자 결정 반영 완료)
- [ ] AC 명령의 정규식에 `\s` / `\d` 등 GNU·PCRE 확장이 없음 (§A.1-5)
