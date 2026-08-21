# Acceptance — SPEC-CLI-STATE-DIR-BOUND-001

> REQ 당 최소 1개의 반증 가능한 AC. 각 AC 는 **판정하는 명령**과 **기대 관측**을 함께 명시하며, 기대 관측은 사람의 판단 없이 한 번의 관측으로 결정 가능해야 한다. 구현자는 명령을 실제로 실행하고 리터럴 출력을 인용해야 한다.
>
> v0.3.1 개정: AC-009 경우 2·4 의 경고를 **stderr + 안정 부분문자열**로 고정, AC-015 에 REQ-10(해석된 루트 출력) 판정 명령 4 추가.
>
> iter3 개정: AC-003 기대 계수 정정(8 vs 6 오류), AC-001/AC-010 을 환경변수 분기를 실제로 태우도록 재작성, AC-008 부정 단언을 성공 경로로 한정, AC-009 에 이전-실패 케이스 추가, AC-015 신설(clean 의 환경변수 제외), AC-014 에 `chain.go:61` 주석 추가.

## §D AC Matrix

| AC ID | REQ | Severity | 요약 |
|---|---|---|---|
| AC-001 | REQ-1 | MUST-PASS | `CLAUDE_PROJECT_DIR` 설정 시 걷기 없이 그 루트를 쓰고, 값은 정규화된다 |
| AC-002 | REQ-2 | MUST-PASS | 조상 체인의 홈 대역 미끼 `.moai/state` 를 주장하지 않음 |
| AC-003 | REQ-6 | MUST-PASS | 걷기 구현 1개, 읽기 계열 5곳은 공유 진입점, clean 은 전용 진입점 |
| AC-004 | REQ-3 | MUST-PASS | `.moai` 있고 `.moai/state` 없는 트리 → 그 자리에서 실패 (부류 A) |
| AC-005 | REQ-5 | MUST-PASS | 서브테스트 1·2 보존, 3 은 결정적 에러로 강화, 스킵 0건 |
| AC-006 | REQ-4 | MUST-PASS | 신규 체크아웃 `<cwd>/.moai/state` 폴백 보존 |
| AC-007 | REQ-6 | MUST-PASS | `resolveChainStore` 의 인라인 env 확인이 공유 헬퍼로 흡수 |
| AC-008 | REQ-7 | MUST-PASS | chain 경로가 두 분기에서 **동일**하게 `.moai/state/chain` (성공 경로 한정) |
| AC-009 | REQ-7 | MUST-PASS | 레거시 `.moai/chain` 처분 4경우 — 이전 / 경고 / 무동작 / 이전 실패 |
| AC-010 | REQ-8 | MUST-PASS | 반환 경로가 정규화 형태 — **환경변수 분기와 걷기 분기 양쪽** |
| AC-011 | REQ-3 | MUST-PASS | 유효 루트 아래 맨 `.moai` 하위 디렉터리에서 실패 (부류 B, 의도적) |
| AC-012 | 횡단 (§E fail-open) | MUST-PASS | `loadRegistryForOverlay` fail-open 불변 |
| AC-013 | 횡단 (플랫폼) | MUST-PASS | `GOOS=windows go vet ./internal/cli/...` exit 0 |
| AC-014 | 횡단 (문서 정합) | MUST-PASS | 옛 주석 3곳 제거 (`state.go`, `tokens.go`, `chain.go:61`) |
| AC-015 | REQ-9 + REQ-10 | MUST-PASS | `clean` 은 env 를 따르지 않으며, 해석된 루트를 출력으로 드러낸다 |

전 항목 MUST-PASS. Tier M (AC 상한 16, REQ 상한 16 — 현재 AC 15 / REQ 10, 예산 내).

> v0.3.1: REQ-10(해석된 루트 출력)은 **AC 를 새로 만들지 않고** AC-015 에 판정 명령 4 로 접었다. 이유 둘: (1) `clean` 은 REQ-10 이 "최소한" 으로 지목한 소비자이자 REQ-9 의 대상이라 같은 시나리오에서 함께 관측된다. (2) AC 상한 16 에 여유를 남긴다 — 새 AC 를 만들면 16/16 이 되어 이후 정정이 예산을 넘긴다.

## §D.1 Severity / Traceability

| REQ | 검증하는 AC |
|---|---|
| REQ-1 (읽기·추가 env 우선 + 정규화) | AC-001 |
| REQ-2 (홈 미통과) | AC-002 |
| REQ-3 (정지 후 실패) | AC-004 (부류 A), AC-011 (부류 B) |
| REQ-4 (신규 체크아웃 폴백) | AC-006 |
| REQ-5 (기존 테스트 처리) | AC-005 |
| REQ-6 (걷기 1개 · 진입점 분리) | AC-003, AC-007 |
| REQ-7 (chain 경로 + 레거시 처분) | AC-008, AC-009 |
| REQ-8 (정규화 계약) | AC-010 |
| REQ-9 (파괴적 소비자 env 제외) | AC-015 (판정 명령 1-3) |
| REQ-10 (해석된 루트 출력) | AC-015 (판정 명령 4) |
| — 횡단 | AC-012 (§E fail-open 제약), AC-013 (플랫폼 게이트), AC-014 (문서 정합) |

- **iter1 매핑 오류 정정(유지)**: 구 AC-003 은 REQ-2 로 잘못 매핑되어 있었고 본문은 REQ-6 의 문언이었다 — REQ-6 으로 이동. 구 AC-008(fail-open)은 REQ-3 으로 잘못 매핑되어 있었다 — AC-012 로 번호를 바꾸고 횡단 제약으로 분류.
- `go test ./...` 는 로컬에서 실행하지 않는다. 아래 모든 명령은 패키지 단위이며, 전 패키지 판정은 PR CI 가 낸다.

## §D.2 사전 규약

### 경로 비교 규약 (AC-001/002/004/010/011/015 공통)

REQ-8 에 따라 해석 결과는 `EvalSymlinks` 정규화 형태다. darwin 에서 `t.TempDir()` 은 `/var/...`, 해석 결과는 `/private/var/...` 로 **문자열이 다르다**(spec.md REQ-8 실측). 따라서 **기대값 쪽을 정규화**한 뒤 비교한다. 테스트 헬퍼 하나로 통일할 것:

```go
func normPath(t *testing.T, p string) string {
    t.Helper()
    if r, err := filepath.EvalSymlinks(p); err == nil { return r }
    return p
}
```

[HARD] **양쪽을 정규화하는 비교는 REQ-8 을 검증하지 못한다.** `normPath(got) == normPath(want)` 는 구현이 정규화를 하든 안 하든 통과한다 — iter2 의 AC-001 이 그 형태였고, 그래서 REQ-1 과 REQ-8 의 모순을 가렸다. 아래 AC 들은 **실제값은 정규화하지 않고, 기대값만 정규화**하여 비교한다:

```go
if got != normPath(t, want) { t.Errorf(...) }   // 실제값 raw, 기대값 정규화 — REQ-8 을 결정한다
```

### 진입점 이름 규약 (AC-003/AC-007/AC-015 판정용)

이 절은 판정을 이진화하기 위해 식별자를 고정한다. 요구사항 계층(spec.md)은 동작만 규정하며 이름을 강제하지 않는다.

| 진입점 | 성격 | 소비자 |
|---|---|---|
| `findStateDir()` | 환경변수 우선 → 보호된 걷기 | 읽기·추가 5곳 |
| `findStateDirNoEnv()` | 보호된 걷기만 (환경변수 미참조) | `clean.go` 1곳 |

두 진입점은 **같은 걷기 구현**을 공유해야 한다(REQ-6). 다른 이름을 택할 경우 구현자는 progress.md 에 대응표를 남기고 아래 grep 을 그에 맞게 치환한다.

---

## §D.3 Given-When-Then 시나리오

### AC-001 — `CLAUDE_PROJECT_DIR` 우선 + 값 정규화 (REQ-1)

```
GIVEN 완전 소유 임시 트리 root 가 있고
  AND root/.moai/state 가 존재하며 (걷기가 찾을 미끼)
  AND root/explicit/.moai/state 가 존재하고
  AND CLAUDE_PROJECT_DIR = root/explicit  (t.TempDir() 이 준 미정규화 형태 그대로)
WHEN 읽기 계열 해석 헬퍼를 호출하면
THEN got == normPath(root/explicit/.moai/state)          ← 실제값은 정규화하지 않는다
  AND got != normPath(root/.moai/state)                  ← 걷기를 타지 않았다
```

**판정 명령**: `go test ./internal/cli/ -run 'TestStateDirHonoursProjectDirEnv' -v`
**기대 관측**: 출력에 `--- PASS: TestStateDirHonoursProjectDirEnv` 가 있고 `--- SKIP` / `no tests to run` 이 없다.
**작성 대상**: 신규.
**iter2 정정**: iter2 의 AC-001 은 양쪽을 `normPath` 로 감싸 어느 구현에서나 통과했다. 위 형태는 환경변수 값을 정규화하지 않는 구현에서 **실패**한다 — REQ-1 과 REQ-8 의 관계를 실제로 결정한다.

### AC-002 — 홈 대역 미끼를 주장하지 않음 (REQ-2, 관측된 A/B 비대칭의 회귀)

```
GIVEN 완전 소유 임시 트리 안에서
  AND 조상 체인 위쪽에 $HOME 역할을 하는 미끼 디렉터리가 있고 그 아래 .moai/state 가 있으며
  AND 시작 디렉터리는 그 미끼의 후손이고, 시작점과 미끼 사이 어느 조상도 .moai 를 갖지 않으며
  AND CLAUDE_PROJECT_DIR 은 설정되지 않았을 때
WHEN state 해석을 수행하면
THEN 에러를 반환한다
  AND (에러가 아닐 경우) got != normPath(미끼/.moai/state)
```

**판정 명령**: `go test ./internal/cli/ -run 'TestStateDirDoesNotCrossHomeBoundary' -v`
**기대 관측**: `--- PASS: TestStateDirDoesNotCrossHomeBoundary`.
**작성 대상**: 신규.
**reproduction-first 의무**: 이 테스트는 **M1 이전 코드에서 반드시 실패한다**. 구현자는 M1 전에 실행해 빨간 출력을 progress.md 에 인용한다.
**스케줄 의존성**: plan.md §E D3(홈 경계 주입 방식)이 M1 에서 결정되기 전까지 **실행 불가**. M1 이전으로 일정을 잡지 말 것.
**테스트 격리**: `t.Setenv("HOME", …)` 은 병렬 오염 위험이 있다(CLAUDE.local.md §13). 주입 seam 을 쓰거나 비병렬로 둔다. 선택을 progress.md 에 기록.

### AC-003 — 걷기 1개, 진입점 분리 (REQ-6)

```
GIVEN §A 표의 6개 호출 지점
WHEN 소스를 검사하면
THEN 읽기·추가 5곳은 findStateDir() 를 호출하고
  AND clean.go 1곳은 findStateDirNoEnv() 를 호출하며
  AND state.go 에 상향 걷기 루프가 남아 있지 않다
```

**iter2 정정 (E1)**: iter2 의 명령 `grep -n "findStateDir()" …` 은 정의(`state.go:212 func findStateDir() …`)와 주석(`chain.go:61 // … falls back to findStateDir() directory walk.`)까지 잡아 **오늘 8을 반환한다** — 기대값 6 은 틀렸다. 게다가 정의는 변경 후에도 살아남으므로(그것이 공유 헬퍼다) `state.go` 가 2가 되는 일은 없다. 아래는 **호출 전용 패턴**이다.

**판정 명령 1** (읽기·추가 계열 호출 수, 파일별):
```bash
grep -c "err := findStateDir()" internal/cli/clean.go internal/cli/tokens.go internal/cli/chain.go internal/cli/state.go
```
**기대 관측 1**: 정확히 아래 값.

| 파일 | 변경 전 (현재 실측) | 변경 후 기대 |
|---|---|---|
| `internal/cli/clean.go` | 1 | **0** (전용 진입점으로 이동) |
| `internal/cli/tokens.go` | 1 | 1 |
| `internal/cli/chain.go` | 2 | 2 |
| `internal/cli/state.go` | 2 | 2 |
| 합 | 6 | **5** |

**판정 명령 2** (clean 의 전용 진입점):
```bash
grep -c "err := findStateDirNoEnv()" internal/cli/clean.go
```
**기대 관측 2**: 정확히 `1`.

**판정 명령 3** (걷기 루프 소멸):
```bash
grep -c "for {" internal/cli/state.go
```
**기대 관측 3**: 정확히 `0`. (현재값 `1` — 그 하나가 문제의 걷기다. 위임안을 택했다면 0 이 올바른 사후 조건이다.)

### AC-004 — 프로젝트 루트에서 멈추고 실패 — 부류 A (REQ-3)

```
GIVEN 완전 소유 임시 트리에서
  AND <root>/.moai 는 존재하나 <root>/.moai/state 는 존재하지 않으며
  AND <root> 의 조상에는 .moai/state 를 가진 디렉터리가 있고
  AND CLAUDE_PROJECT_DIR 은 설정되지 않았을 때
WHEN <root>/sub 에서 state 해석을 수행하면
THEN 에러를 반환한다 (조상에서 성공하지 않는다)
  AND 에러 메시지가 normPath(<root>) 문자열을 포함한다 — 어디서 멈췄는지 지목
```

**판정 명령**: `go test ./internal/cli/ -run 'TestStateDirStopsAtProjectRoot' -v`
**기대 관측**: `--- PASS: TestStateDirStopsAtProjectRoot`.
**작성 대상**: 신규. spec.md §E 동작 변경 1행의 검증 지점.

### AC-005 — 기존 테스트: 1·2 보존, 3 강화, 스킵 0건 (REQ-5)

```
GIVEN internal/cli/tokens_state_dir_test.go 의 세 서브테스트
WHEN 이 SPEC 이 착지하면
THEN 서브테스트 1 ("an ancestor state dir wins over the starting directory") 의 단언 의미가 보존되고
  AND 서브테스트 2 는 보존되되 경로 비교가 §D.2 규약을 따르며
  AND 서브테스트 3 은 err != nil 을 직접 단언하도록 강화되고
  AND 파일 어디에도 t.Skip 이 없다
```

**판정 명령 1**: `go test ./internal/cli/ -run 'TestFindStateDirFromWalksUp' -v`
**기대 관측 1**: 세 서브테스트 모두 `--- PASS`. `--- SKIP` 없음.

**판정 명령 2**: `grep -c "t.Skip" internal/cli/tokens_state_dir_test.go` → 기대 `0`.

**판정 명령 3**: `grep -c "err == nil && strings.HasPrefix" internal/cli/tokens_state_dir_test.go` → 기대 `0` — 서브테스트 3 의 환경 의존 조건부 단언이 결정적 에러 단언으로 교체되었음을 뜻한다.

**iter1 정정 반영(유지)**: 서브테스트 1 을 **반전시키지 않는다.** 그 트리는 `root` 에 `.moai/state` 가 있고 시작점이 그 아래이므로 권고안에서도 같은 조상을 반환한다 — 무변경 통과한다. 반전은 올바른 구현이 만들지 않는 계약을 고정한다. 테스트 이름(`WalksUp`) 개명은 SHOULD 이며 이 AC 의 판정 조건이 아니다.

### AC-006 — 신규 체크아웃 폴백 보존 (REQ-4)

```
GIVEN 프로젝트 루트를 찾을 수 없는 임시 디렉터리에서
  AND CLAUDE_PROJECT_DIR 이 설정되지 않았을 때
WHEN resolveTokensStateDir 을 호출하면
THEN 에러가 아니며 got == normPath(<cwd>/.moai/state)
```

**판정 명령**: `go test ./internal/cli/ -run 'TestResolveTokensStateDirFallsBackToCwd' -v`
**기대 관측**: `--- PASS: TestResolveTokensStateDirFallsBackToCwd`.
**작성 대상**: 신규. 이 폴백이 사라지면 사전 스캐폴딩 없는 신규 체크아웃에서 `moai tokens record` 가 깨진다.

### AC-007 — `resolveChainStore` 의 인라인 env 확인이 헬퍼로 흡수 (REQ-6)

```
GIVEN chain.go 가 오늘 EnvClaudeProjectDir 을 두 곳에서 읽고 있을 때
      (line 64 resolveChainStore — 범위 안 / line 83 resolveCWD — 범위 밖)
WHEN 이 SPEC 이 착지하면
THEN resolveChainStore 안의 참조는 0건이 되고
  AND 파일 전체 참조는 정확히 1건이며 그 1건은 resolveCWD 안에 있다
```

**판정 명령 1**: `grep -c "EnvClaudeProjectDir" internal/cli/chain.go` → 기대 `1` (현재값 `2`).

**판정 명령 2**: `awk '/^func resolveChainStore/,/^}/' internal/cli/chain.go | grep -c "EnvClaudeProjectDir"` → 기대 `0` (현재값 `1`).

**판정 명령 3**: `awk '/^func resolveCWD/,/^}/' internal/cli/chain.go | grep -c "EnvClaudeProjectDir"` → 기대 `1` (현재값 `1`, 불변). `resolveCWD` 는 범위 밖이며 **건드리지 않는다.**

세 명령이 함께 "범위 안은 비었고 범위 밖은 그대로"를 이진 판정한다.

### AC-008 — chain 경로가 두 분기에서 동일 — 성공 경로 한정 (REQ-7)

```
GIVEN 완전 소유 임시 트리 root 에 .moai/state 가 존재하고
  AND 레거시 .moai/chain 도, 정본 .moai/state/chain 도 이벤트 파일을 갖지 않을 때
      (= REQ-7 경우 3 이후의 정상 상태, 이전(migration)이 개입하지 않는 성공 경로)
WHEN CLAUDE_PROJECT_DIR = root 로 chain 디렉터리를 해석하고
  AND 이어서 CLAUDE_PROJECT_DIR 없이 root/sub 에서 해석하면
THEN 두 결과가 서로 같고
  AND 둘 다 normPath(root/.moai/state/chain) 이며
  AND 어느 쪽도 root/.moai/chain 이 아니다
```

**판정 명령**: `go test ./internal/cli/ -run 'TestChainDirIsCanonicalUnderBothBranches' -v`
**기대 관측**: `--- PASS: TestChainDirIsCanonicalUnderBothBranches`.
**작성 대상**: 신규. **오늘의 코드에서 반드시 실패한다** — env 분기는 `.moai/state/chain`, walk 분기는 `.moai/chain` 을 만든다(spec.md REQ-7 산술). 구현 전 빨간 출력을 progress.md 에 인용할 것.
**iter2 정정 (E7)**: 이 AC 의 "어느 쪽도 `.moai/chain` 이 아니다" 단언은 **성공 경로에 한정**한다. REQ-7 경우 4(이전 실패)는 레거시 경로를 계속 쓰는 것이 정의된 동작이므로, 이 AC 의 Given 이 그 상황을 배제한다. 무조건 단언은 REQ-7 의 선언된 예외와 충돌한다.
**이진성**: 두 결과의 동등과 정본 일치를 함께 단언하므로, 정본을 `.moai/chain` 으로 잘못 고르면 실패한다.

### AC-009 — 레거시 chain 이벤트 처분 4경우 (REQ-7)

REQ-7 의 네 경우 전부를 단언한다.

```
경우 1 — 레거시만 존재:
GIVEN root/.moai/chain/events.jsonl 이 존재하고
  AND root/.moai/state/chain/events.jsonl 은 존재하지 않을 때
WHEN chain 디렉터리를 해석하면
THEN root/.moai/state/chain/events.jsonl 이 존재하고 내용이 원본과 바이트 동일하며
  AND root/.moai/chain/events.jsonl 은 더 이상 존재하지 않는다

경우 2 — 양쪽 존재 (기대 상태, R4):
GIVEN 두 경로에 모두 events.jsonl 이 있을 때
WHEN 해석하면
THEN 정본(.moai/state/chain) 이 선택되고
  AND 레거시 파일은 내용·존재 모두 변경되지 않으며
  AND stderr 에 안정 부분문자열 "chain: legacy events retained at" 를 포함하는 줄이 나온다
  AND 그 줄이 레거시 경로 문자열을 포함한다

경우 3 — 정본만 존재:
GIVEN 정본에만 events.jsonl 이 있을 때
WHEN 해석하면
THEN 아무 이전도 일어나지 않고 경고도 없다

경우 4 — 이전 실패 (신규, REQ-7 의 선언된 예외):
GIVEN 레거시에만 events.jsonl 이 있고
  AND 정본 경로로의 이전이 실패하도록 만들어진 상태일 때
      (예: .moai/state/chain 을 쓰기 불가 퍼미션으로 생성)
WHEN 해석하면
THEN 해석은 레거시 경로를 반환하고 (에러로 중단하지 않는다)
  AND stderr 에 안정 부분문자열 "chain: migration failed, using legacy path" 를 포함하는 줄이 나오며
  AND 그 줄이 레거시 경로 문자열을 포함하고
  AND 레거시 events.jsonl 의 내용이 바이트 단위로 손실되지 않는다
```

**경고 문자열 고정 (v0.3.1, iter3 감사 Testability 지적)**: v0.3.0 까지 경우 2·4 는 "관측 가능한 경고 (stderr 또는 printer)" 라고만 적어 **스트림도 문자열도 지목하지 않았다**. grep 할 수 없는 경고는 반증 가능한 관측이 아니며, 하필 이 둘이 REQ-7 에서 데이터가 걸린 가장 위험한 두 분기다. 따라서:

- **스트림**: stderr 로 고정한다. printer 추상화를 쓰더라도 최종 목적지가 stderr 여야 한다.
- **안정 부분문자열**: 위 두 리터럴을 계약으로 취급한다. 문구를 다듬을 수는 있으나 **이 부분문자열은 보존**한다 — 바꾸려면 이 AC 를 함께 고친다.
- **경로 동반**: 두 경우 모두 경고 줄이 레거시 **경로**를 담아야 한다. 경로 없는 경고는 사용자가 어디를 봐야 할지 모르므로 R4 의 완화("경고 문구가 상황을 정확히 설명한다")를 이행하지 못한다.

테스트는 stderr 를 캡처해 `strings.Contains` 로 두 조건(부분문자열 + 경로)을 각각 단언한다.

**판정 명령**: `go test ./internal/cli/ -run 'TestChainLegacyEventsRelocation' -v`
**기대 관측**: `--- PASS: TestChainLegacyEventsRelocation` 과 **네** 서브테스트 각각의 `--- PASS`.
**작성 대상**: 신규. REQ-7 의 "조용히 고아가 되지 않는다"를 지키는 유일한 기계적 확인이다.
**플랫폼 주의**: 경우 4 를 퍼미션으로 만들 때 Windows 에서는 퍼미션 모델이 달라 재현되지 않을 수 있다. 그 경우 `t.Skip` 대신 **주입 가능한 실패 seam**(이전 함수의 오류 반환을 테스트가 강제)으로 구성한다 — 스킵은 AC-005 의 스킵 금지 정신과 어긋난다.

### AC-010 — 정규화 계약, 양 분기 (REQ-8)

```
분기 A — 걷기 분기:
GIVEN CLAUDE_PROJECT_DIR 이 설정되지 않았고
  AND t.TempDir() 이 준 (darwin 에서 미정규화) 경로 아래에 유효한 프로젝트 트리가 있을 때
WHEN 해석하면
THEN got == filepath.EvalSymlinks(got)      ← 반환값이 이미 정규화 형태

분기 B — 환경변수 분기 (신규, iter2 가 빠뜨린 쪽):
GIVEN CLAUDE_PROJECT_DIR 이 의도적으로 미정규화된 경로로 설정되어 있을 때
      (darwin: t.TempDir() 이 준 /var/... 형태를 그대로 사용)
WHEN 읽기 계열 해석을 수행하면
THEN got == filepath.EvalSymlinks(got)
  AND got 는 CLAUDE_PROJECT_DIR 원문 접두사를 그대로 물고 있지 않다
```

**판정 명령 1**: `go test ./internal/cli/ -run 'TestStateDirReturnsNormalizedPath' -v`
**기대 관측 1**: `--- PASS` 와 두 서브테스트(분기 A / 분기 B) 각각의 `--- PASS`. 단언은 플랫폼 무관하게 성립한다(정규화된 경로는 재정규화해도 자기 자신). darwin 에서만 분기 B 가 실질적으로 구분력을 갖는다 — 그 사실을 테스트 주석에 남긴다.

**판정 명령 2**: `grep -c "EvalSymlinks" internal/cli/state.go` → 기대 `1` 이상 — 계약이 코드나 주석에 문자로 나타난다(REQ-8 은 주석 명시를 SHALL 로 요구).

**iter2 정정 (E4)**: iter2 의 AC-010 은 시나리오가 환경변수 설정 여부를 말하지 않아 구현자가 자연스럽게 걷기 분기만 태우고 통과했다. 분기 B 를 명시적으로 분리해 그 누락을 닫는다.

### AC-011 — 유효 루트 아래 맨 `.moai` 하위에서 실패 — 부류 B (REQ-3)

```
GIVEN 완전 소유 임시 트리에서
  AND root/.moai/state 가 존재하여 root 는 유효한 프로젝트 루트이고
  AND root/nested/.moai 는 존재하나 root/nested/.moai/state 는 없으며
  AND CLAUDE_PROJECT_DIR 은 설정되지 않았을 때
WHEN root/nested/sub 에서 state 해석을 수행하면
THEN 에러를 반환한다 (root 의 유효한 state 로 올라가지 않는다)
  AND 에러 메시지가 normPath(root/nested) 를 포함한다
```

**판정 명령**: `go test ./internal/cli/ -run 'TestStateDirStopsAtNestedBareMoai' -v`
**기대 관측**: `--- PASS: TestStateDirStopsAtNestedBareMoai`.
**작성 대상**: 신규.
**이 AC 는 의도적 회귀를 고정한다.** 오늘은 이 트리에서 성공하고, 변경 후에는 실패한다 — spec.md REQ-3 부류 B. *정답이 도달 가능한데 거부하는* 케이스이므로 우연이 아니라 결정임을 테스트로 못 박는다. 이 리포는 `internal/hook/.moai/` 형태의 이력이 있으므로 가상의 케이스가 아니다.

### AC-012 — `loadRegistryForOverlay` fail-open 불변 (횡단, §E 제약)

```
GIVEN state 해석이 실패하는 디렉터리에서 (.moai 없는 임시 트리, CLAUDE_PROJECT_DIR 미설정)
WHEN loadRegistryForOverlay 를 호출하면
THEN nil 을 반환하고 패닉하지 않으며 에러를 상위로 전파하지 않는다
```

**판정 명령**: `go test ./internal/cli/ -run 'TestLoadRegistryForOverlayFailsOpen' -v`
**기대 관측**: `--- PASS: TestLoadRegistryForOverlayFailsOpen`.
**작성 대상**: **신규 — 오늘 이 이름의 테스트는 존재하지 않는다.** iter1 은 `-run 'Overlay|RegistryForOverlay'` 패턴을 썼는데 매칭 테스트가 0개라 `ok … [no tests to run]` 으로 exit 0 이 되어 **아무것도 검증하지 않고 PASS 로 읽혔다.**
**왜 중요한가**: AC-004 와 AC-011 이 해석 실패 경우를 **의도적으로 늘리므로** 이 fail-open 경로의 통행량이 증가한다. 회귀 아님을 확인하는 유일한 지점이다.

### AC-013 — 크로스 플랫폼 컴파일 게이트 (횡단)

**판정 명령**: `GOOS=windows go vet ./internal/cli/...`
**기대 관측**: 출력 없음, exit 0. (이 결함의 원 증상이 windows 전용 CI flake 였으므로 형식이 아니다.)

### AC-014 — 옛 주석 제거 3곳 (횡단, 문서 정합)

```
GIVEN 세 주석이 변경 전 동작을 서술하고 있을 때
      state.go — "unbounded" / "inherits any ~/.moai/state on the machine"
      tokens.go — "existing walk-up convention (findStateDir, state.go)"
      chain.go:61 — "Prefers CLAUDE_PROJECT_DIR; falls back to findStateDir() directory walk."
WHEN 이 SPEC 이 착지하면
THEN 셋 모두 새 동작(보호된 해석 + 정규화 계약 + 정본 chain 경로)으로 교체되어 있다
```

**판정 명령 1**: `grep -c "unbounded\|inherits any" internal/cli/state.go` → 기대 `0`.
**판정 명령 2**: `grep -c "existing walk-up convention" internal/cli/tokens.go` → 기대 `0`.
**판정 명령 3**: `grep -c "falls back to findStateDir() directory walk" internal/cli/chain.go` → 기대 `0`.

**iter2 정정 (E1 부수)**: `chain.go:61` 주석은 REQ-1 의 끌어올리기 이후 사실과 어긋나지만 iter2 의 AC-014 는 `state.go` / `tokens.go` 만 훑었다. 세 번째 명령으로 닫는다.

### AC-015 — `clean` 은 env 를 따르지 않고, 해석된 루트를 드러낸다 (REQ-9 + REQ-10)

```
GIVEN 완전 소유 임시 트리 A 와 B 가 있고
  AND A/.moai/state/runs/<오래된 run> 이 존재하며 (삭제 대상 후보)
  AND B/.moai/state 도 존재하고
  AND cwd 는 A/sub 이며
  AND CLAUDE_PROJECT_DIR = B 로 설정되어 있을 때
WHEN runClean 의 state 해석을 수행하면
THEN 결과는 normPath(A/.moai/state) 이고            ← REQ-9
  AND normPath(B/.moai/state) 가 아니다
  AND B 아래의 어떤 파일도 삭제되지 않는다
  AND 삭제 후보를 열거하기 전에 해석된 루트를 명시하는 줄이 출력된다   ← REQ-10
  AND 그 줄이 normPath(A) 를 포함하고 normPath(B) 를 포함하지 않는다
```

**판정 명령 1**: `go test ./internal/cli/ -run 'TestCleanIgnoresProjectDirEnv' -v`
**기대 관측 1**: `--- PASS: TestCleanIgnoresProjectDirEnv`.
**작성 대상**: 신규.

**판정 명령 2** (정적 확인 — 환경변수가 clean 의 해석 경로에 아예 없음):
```bash
grep -c "EnvClaudeProjectDir\|CLAUDE_PROJECT_DIR" internal/cli/clean.go
```
**기대 관측 2**: 정확히 `0` (현재값도 `0` — 이 AC 는 그 성질의 **보존**을 요구한다).

**판정 명령 3** (clean 이 env 우선 진입점을 쓰지 않음):
```bash
grep -c "err := findStateDir()" internal/cli/clean.go
```
**기대 관측 3**: 정확히 `0` (AC-003 판정 명령 1 의 clean 행과 동일한 사실을 REQ-9 관점에서 재확인).

**판정 명령 4** (REQ-10 — 해석된 루트가 출력에 있다):
```bash
go test ./internal/cli/ -run 'TestCleanAnnouncesResolvedRoot' -v
```
**기대 관측 4**: `--- PASS: TestCleanAnnouncesResolvedRoot`. 테스트는 `runClean` 의 출력을 캡처해 (a) 해석된 루트를 담은 줄이 **삭제 후보 열거보다 먼저** 나오는지, (b) 그 줄이 `normPath(A)` 를 포함하고 `normPath(B)` 를 포함하지 않는지를 단언한다.
**작성 대상**: 신규.
**왜 필요한가 (v0.3.1, iter3 감사 F3)**: REQ-9 는 `clean` 을 환경변수에서 떼어냈고 그 대가로 **한 세션 안에서 해석이 갈린다** — `moai state dump` 는 B 를 읽고 `moai clean` 은 A 에서 지운다. 이 줄이 없으면 그 분기는 **아무것도 지워지지 않은 것을 운영자가 이상하게 여길 때에야** 발견된다. spec.md §E "선언된 부수 결과" 가 이 분기를 수용 결과로 선언했고, REQ-10 이 유일한 완화이며, 이 명령이 그 완화를 이진 판정한다. 부수 효과로 §D.6 의 `clean` 사후 확인이 자기증거화된다.

**왜 이 AC 가 존재하는가**: REQ-1 의 끌어올리기를 6곳에 일괄 적용하면 `clean.go:116` 의 `os.RemoveAll` 이 환경변수를 따라간다. 워크트리 안에서 실행한 `moai clean` 이 primary 체크아웃 아래를 지우는 형태이며, 이는 이 SPEC 의 발단이 된 위험이 다른 문으로 되돌아오는 것이다(spec.md REQ-9). 이 AC 가 그 문을 닫는다.

## §D.4 간접 검증

- **reproduction-first 증거 3건**: AC-002, AC-008, AC-011 의 테스트는 구현 전 트리에서 실패해야 한다. 세 빨간 출력이 progress.md 에 인용되어 있어야 한다.
- **삭제량 확인**: 이 변경은 규약 수렴이므로 `git diff --stat` 에서 `state.go` 의 순증가가 크면 위임이 아니라 복제를 했다는 신호다. 증가분이 20줄을 넘으면 plan.md §E D1 재검토. (REQ-7 의 이전 로직은 `chain.go` 쪽 증가이며 이 임계에 포함하지 않는다.)
- **race**: `go test -race ./internal/cli/...` 초록 — 홈 경계 주입에 환경변수를 썼다면 병렬 오염이 여기서 드러난다.
- **`m2SetupState` 무영향 확인**: `go test ./internal/cli/ -run 'M2' -v` 초록 (기존 6개 `TestStateM2_*` 매칭). spec.md §A 의 무영향 주장을 관측으로 뒷받침한다.
- **진입점 이름 대응표**: §D.2 의 이름(`findStateDir` / `findStateDirNoEnv`)과 다른 이름을 택했다면 progress.md 에 대응표가 있어야 하고, 위 grep 들이 그에 맞게 치환되어 실행되었어야 한다.

## §D.5 Closure Gate (Definition of Done)

- [ ] §D.3 의 MUST-PASS AC 15개를 모두 실행하고 리터럴 출력을 관측했다.
- [ ] `go test ./internal/cli/... ./internal/core/project/...` 초록.
- [ ] `go test -race ./internal/cli/...` 초록.
- [ ] `GOOS=windows go vet ./internal/cli/...` exit 0.
- [ ] `golangci-lint run --timeout=2m` 클린.
- [ ] AC-002 / AC-008 / AC-011 의 reproduction-first 실패 출력 3건이 progress.md 에 기록되어 있다.
- [ ] plan.md §E D3(홈 경계 주입 방식) 선택과 근거가 progress.md 에 기록되어 있다.
- [ ] REQ-7 의 chain 정본 경로와 4경우 처분이 코드와 사용자 안내(경고 문구)에 반영되어 있다.
- [ ] REQ-9 의 carve-out 이 코드에서 이름으로 드러난다(`clean` 이 전용 진입점을 쓴다).
- [ ] REQ-10 의 해석된-루트 줄이 여섯 소비자 전부에 있고, 최소한 `clean` 은 삭제 후보 열거 전에 출력한다.
- [ ] AC-009 의 두 경고 부분문자열이 코드에 리터럴로 존재하고 stderr 로 나간다.
- [ ] 로컬에서 `go test ./...` 를 실행하지 **않았다**.
- [ ] `internal/template/templates/` 를 건드리지 않았다 (미러 대상 없음).
- [ ] `internal/core/project/root.go` 를 수정하지 않았다 — 수정이 필요했다면 그 사실과 근거가 progress.md 에 있다.
- [ ] `internal/hook/chain_event.go` 를 수정하지 않았다 (범위 밖 — spec.md §B).

## §D.6 Forward-Looking Checks (Post-Merge)

- 머지 후 실제 머신에서 spec.md §A 의 A 프로브를 재실행: `~` 하위 임의 디렉터리에서 `CLAUDE_PROJECT_DIR` 없이 `moai state show-blocker` 실행 → **에러**가 나야 한다(오늘은 `No blockers found` 로 조용히 성공). 결함이 실제로 사라졌다는 최종 확인이며 유닛 테스트가 대신할 수 없다.
- REQ-7 이전 경로의 실사용 확인: `moai chain` 을 써 온 머신에서 `.moai/chain/events.jsonl` 이 `.moai/state/chain/` 으로 옮겨졌는지, 또는 경우 2 의 경고가 정확한 문구로 나오는지 1회 확인.
- REQ-9 의 실사용 확인: 워크트리 세션에서 `CLAUDE_PROJECT_DIR` 이 primary 를 가리키는 상태로 `moai clean --force` 를 dry-run(`--force` 없이) 실행해, 삭제 후보가 **워크트리 쪽**으로 열거되는지 확인.
- Windows CI 러너의 조상 `.moai/state` 생성 주체 확인은 **후속 카드**: `release-pr-multi-os.yml` 테스트 스텝 앞뒤에 진단 2줄 추가, 테스트 변경 0.
- 규약 파편화 정리(세 번째 걷기 `internal/config/token_budget_guard.go:51`)와 훅 writer 리터럴 공유도 후속 후보.

## §D.7 Quality Gate Criteria (TRUST 5)

- **Tested**: 신규 테스트 9종 + 강화 1종(AC-005 서브테스트 3)이 결함을 실제로 잡는다(reproduction-first 증거 3건 필수). `internal/cli` state 해석 경로 커버리지 ≥ 85%.
- **Readable**: 함수 주석이 새 동작과 정규화 계약을 서술하고(AC-010, AC-014), 진입점 이름이 "삭제하는가"라는 성격을 드러낸다(AC-015).
- **Unified**: 걷기 구현이 하나로 줄었다(AC-003). chain 경로가 하나로 줄었다(AC-008). 네 번째 규약을 만들지 않았다.
- **Secured**: 사용자 홈의 전역 상태를 프로젝트 상태로 오인하지 않으며, 환경변수가 삭제 대상을 결정하지 못한다(AC-015). 레거시 chain 이벤트를 조용히 잃지 않는다(AC-009).
- **Trackable**: SPEC ID 를 담은 Conventional Commit. 마일스톤별 분리 커밋 또는 단일 커밋 모두 허용(Tier M).
