# Acceptance — SPEC-STATE-ANCHOR-VALIDATE-001

측정 기준 트리: `.claude/worktrees/t537` @ `52f863f36` (`WT-resolve-validate`). 모든 RED-now 셀은 **본 레인이 이 트리에서 1차 판독한 코드 근거**를 가진 예정형이며, 실제 RED 관측(채택)은 run-phase M1이 committed RED 출력으로 수행한다(plan §D9).

등급 용어: **blocking** = RED-now 셀 + green path 셀을 갖춘 릴리스 게이트. **regression-guard** = 기준 트리에서 이미 GREEN이라 RED-now 셀을 가질 수 없는 회귀 방어(verification-completeness §2 undecidable disposition).

---

## §D AC 매트릭스

| AC | 요구사항 | RED-now (예정형 — 현재 트리 기준) | GREEN (목표) | 등급 |
|---|---|---|---|---|
| AC-SAV-001 | REQ-SAV-001, 002 | 상대경로 `ProjectDir`가 그대로 반환됨 (`stateanchor.go:65-66`) | 폴스루 — git 워크업 또는 "" | blocking |
| AC-SAV-002 | REQ-SAV-001, 002 | 부재 절대경로 `ProjectDir`가 그대로 반환됨 | 폴스루 — 후보 값 미반환 + "" 또는 구원 | blocking |
| AC-SAV-003 | REQ-SAV-001 | (happy path — 현재도 반환하나 무검증) | 유효 절대 디렉터 후보 그대로 반환 (워크트리 하위 케이스 포함) | blocking |
| AC-SAV-004 | REQ-SAV-001, 002, 003 | `OriginalCwd` 무조건 반환 (`stateanchor.go:68-70`) | 유효 → 반환 / 무효 → 폴스루 (빈 ProjectDir + 유효 OriginalCwd → 반환) | blocking |
| AC-SAV-005 | REQ-SAV-003, 007 | (이미 GREEN — regression-guard) | 빈 전체 → "" 유지, 테스트 무수정 | regression-guard |
| AC-SAV-006 | REQ-SAV-005, 006 | (이미 GREEN — regression-guard) | FromDirectory/기존 git 테스트 무수정 GREEN + 반환값 무재작성 | regression-guard |
| AC-SAV-007 | REQ-SAV-004 | (이미 GREEN — regression-guard) | 후보 경로 git-free 코드 검사 + 생산 diff 2파일 한정 | regression-guard |
| AC-SAV-008 | REQ-SAV-005 | (baseline GREEN — regression-guard) | 패키지 테스트 전수 GREEN + vet 클린 + coverage ≥85% | regression-guard |

---

## §D.1 AC 상세

### AC-SAV-001 — 상대경로 후보는 폴스루한다

maps REQ-SAV-001, REQ-SAV-002

- **Given** `ProjectDir`가 상대경로(예: `relative/dir`)인 `Session`과, 폴스루 판정에 쓸 `CurrentDir`가 주어지고,
- **When** `Resolve`가 실행되면,
- **Then** 반환값은 상대경로 후보가 **아니며**, 후보가 기각됐음을 관측할 수 있다(`CurrentDir` 없음 → `""`; git fixture `CurrentDir` → 워크업 결과).

판정 명령: `go test ./internal/stateanchor/ -run TestResolve_RelativeProjectDirFallsThrough -count=1`

**RED-now (예정형)**: 현재 `stateanchor.go:65-66`이 `!= ""`만 확인하고 반환 — 상대경로가 그대로 나온다. **RED 이유**: 무검증 반환(§B1) — 구현 후 같은 명령이 GREEN이 된다. **GREEN (M2)**: PASS.

### AC-SAV-002 — 부재 절대경로 후보는 폴스루한다

maps REQ-SAV-001, REQ-SAV-002

- **Given** 존재했다가 **삭제된** `t.TempDir()` 경로를 `ProjectDir`로 갖는 `Session`이 있고,
- **When** `Resolve`가 실행되면,
- **Then** 반환값은 그 경로가 아니며(부활 차단 — §1.1 기전의 Resolve 반환값 측 판정), 폴스루가 일어난다.
- **And** 폴스루는 "" 즉시 반환이 아니다 — `OriginalCwd`/git 워크업에 구원 기회가 남는다(§3 뮤턴트 4 차단).

판정 명령: `go test ./internal/stateanchor/ -run TestResolve_StaleProjectDirFallsThrough -count=1`

**RED-now (예정형)**: 현재 코드가 삭제 경로를 그대로 반환. **GREEN (M2)**: PASS. 뮤턴트 대응: IsAbs-only 구현은 이 AC를 실패한다.

### AC-SAV-003 — 유효 절대 디렉터 후보는 그대로 반환한다 (happy path + 워크트리 케이스)

maps REQ-SAV-001

- **Given** 실재하는 절대 디렉터(`t.TempDir()`)를 `ProjectDir`로 갖는 `Session`이 있고,
- **When** `Resolve`가 실행되면,
- **Then** 그 디렉터가 **그대로**(재작성 없이 — REQ-SAV-006) 반환된다.
- **And** 후보가 git common-dir 부모가 **아닌** 워크트리 디렉터여도(링크드 워크트리 fixture) 반환된다 — git-membership 뮤턴트(§3 항목 3)의 판별 케이스.

판정 명령: `go test ./internal/stateanchor/ -run TestResolve_ProjectDirValidDirWins -count=1` (기존 `TestResolve_ProjectDirWins`의 계약 갱신 후계자 — plan §D8)

**RED-now**: 해당 없음(현재도 반환) — 그러나 **무검증 반환과 구분되지 않는 상태이므로**, 이 AC의 값은 AC-SAV-001/002와의 뮤턴트 쌍(과잉 검증 뮤턴트를 잡는 방향)으로 성립한다. blocking 등급 유지: 계약 갱신 테스트가 이 AC를 운반한다. **GREEN (M2)**: PASS.

### AC-SAV-004 — OriginalCwd는 같은 술어를 통과한다

maps REQ-SAV-001, REQ-SAV-002, REQ-SAV-003

- **Given** (a) `OriginalCwd`가 상대경로/부재 절대경로인 Session, (b) `ProjectDir` 없이 `OriginalCwd`가 실재 절대 디렉터인 Session, (c) (a)에 git fixture `CurrentDir`를 얹은 Session이 있고,
- **When** `Resolve`가 실행되면,
- **Then** (a)는 후보 기각 + 폴스루, (b)는 그대로 반환, (c)는 git 워크업 결과로 구원된다.

판정 명령: `go test ./internal/stateanchor/ -run TestResolve_OriginalCwd -count=1` (무효/유효 하위 테스트 포함; 기존 `TestResolve_OriginalCwdSecond`의 계약 갱신 후계자 — plan §D8)

**RED-now (예정형)**: 현재 `stateanchor.go:68-70` 무조건 반환. **GREEN (M2)**: PASS.

### AC-SAV-005 — 빈 전체는 ""다 (regression-guard)

maps REQ-SAV-003, REQ-SAV-007

- **Given** 모든 필드가 빈 `Session`이 있고,
- **When** `Resolve`가 실행되면,
- **Then** `""`다 — 기존 `TestResolve_EmptySessionIsEmpty`가 **수정 없이** GREEN으로 유지된다.

판정 명령: `go test ./internal/stateanchor/ -run TestResolve_EmptySessionIsEmpty -count=1`

**RED-now**: 없다 — 이미 GREEN. **GREEN (전 마일스톤 유지)**: 무수정 GREEN. 폴스루 구현이 빈-부재 의미론을 실수로 바꾸면 이 AC가 잡는다.

### AC-SAV-006 — FromDirectory와 git 워크업은 불변이다 (regression-guard)

maps REQ-SAV-005, REQ-SAV-006

- **Given** 기존 테스트 `TestResolve_GitWalkUpFromSubdirectory`, `TestResolve_NonGitDirectoryIsEmpty`, `TestFromDirectory_WorktreeResolvesPrimaryCheckout`, `TestFromDirectory_EmptyIsEmpty`가 있고,
- **When** 경화가 착지하면,
- **Then** 네 테스트가 **수정 없이** 통과하고, `FromDirectory` 함수 본체 diff는 0이다.
- **And** 검증 통과 후보는 `EvalSymlinks` 재작성 없이 그대로 반환된다(AC-SAV-003의 그대로-반환 단언이 운반).

판정 명령: `go test ./internal/stateanchor/ -count=1` + `git diff --stat -- internal/stateanchor/stateanchor.go` 검사(FromDirectory 본체 무변경 확인 — run-phase M3가 관측 방법을 §E.2에 기록).

**RED-now**: 없다 — 이미 GREEN. **GREEN (전 마일스톤 유지)**: 무수정 GREEN 유지.

### AC-SAV-007 — 후보 경로는 git-free다 (regression-guard)

maps REQ-SAV-004

- **Given** 경화가 착지한 트리가 있고,
- **When** `stateanchor.go`의 후보 검증 경로(1·2단)를 코드로 검사하면,
- **Then** 후보 판정에 git 호출/spawn이 없다(`git` 패키지 사용이 `FromDirectory`에 한정됨 — grep으로 관측) — git-membership 뮤턴트(§3 항목 3)의 재도입을 잡는다.
- **And** 생산 변경이 `internal/stateanchor/` 2파일로 한정된다(plan §D7 — `git diff --stat <base>..HEAD` 관측).

판정 명령: `grep -n "git\." internal/stateanchor/stateanchor.go` (FromDirectory 내 1소재 한정) + diff 관측. **RED-now**: 없다(이미 그렇다). **GREEN (M3 유지)**: 관측 기록 §E.2.

### AC-SAV-008 — 패키지 게이트 (regression-guard)

maps REQ-SAV-005

- **Given** 경화가 착지한 트리가 있고,
- **When** 패키지 검증을 실행하면,
- **Then** `go test ./internal/stateanchor/ -count=1` GREEN, `go vet ./internal/stateanchor/` 클린, `go test -cover ./internal/stateanchor/` coverage ≥85%.

판정 명령: 위 3 명령. **RED-now**: 없다(baseline GREEN — §C). **GREEN (M3)**: 전부 클린.

---

## §D.2 기원·계약 traceability

| AC | 근거 |
|---|---|
| AC-SAV-001/002 | t510 sync-audit F1 (`.moai/reports/t510/sync-audit.md:120` 전문 재확인) + §1.1 부활 기전 (MkdirAll 3좌표 직접 판독) |
| AC-SAV-003 | §5 술어 격자 (a)+(b) 채택 + §3 뮤턴트 3 (git-membership 판별 — 워크트리 하위 케이스) |
| AC-SAV-004 | F1 (`ProjectDir`/`OriginalCwd` 동일 취급) + §3 뮤턴트 4 (폴스루 판별) |
| AC-SAV-005 | 부모 REQ-SA-003 계승 + 기존 테스트 자산 보존 |
| AC-SAV-006 | REQ-SAV-005/006 (FromDirectory 불변 + 값 무재작성 — 부모 D13 계승) |
| AC-SAV-007 | REQ-SAV-004 (§5 옵션 (c) 기각의 요구사항화) + plan §D7 |
| AC-SAV-008 | TRUST 5 Tested (85% 기준) + plan §D10 |

범위 밖 항목 근거: kanban 결함 = §7 코드 판독 (t536 소관), model_cache 정정 = §1.1 부록 (본 레인 직접 판독).
