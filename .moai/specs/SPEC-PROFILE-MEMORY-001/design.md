# SPEC-PROFILE-MEMORY-001 — 설계

> Tier L 산출물. `plan.md` §D 의 결정을 자료구조·API 표면·제어흐름 관점에서 확장한다. 결정 자체의 근거는 `plan.md` §C/§D, 코드 실측은 `research.md` 를 참조한다.

---

## §A 설계 목표와 비목표

| 목표 | 대응 REQ |
|------|----------|
| 프로필 기억을 프로젝트 단위로 분리하되 전역 폴백을 잃지 않는다 | REQ-PM-001..004 |
| 유령 프로필 이름이 원장에 들어가지 못하게 한다 | REQ-PM-011..013 |
| 새 프로필 전환을 사전에 고지한다(치유는 하지 않는다) | REQ-PM-016..020 |
| 기존 호출자 5곳 중 4곳을 무변경으로 유지한다 | REQ-PM-007, 010 |

| 비목표 | 근거 |
|--------|------|
| 인증 상태 시드 | `spec.md` §C — 플랫폼별 운반체 상이(Keychain vs 파일) |
| 원장 GC | `plan.md` §C RESOLVED-2 — 정리 메커니즘 선례 부재 |
| 원장 잠금 | `spec.md` §C — 크로스 플랫폼 제약 C5와 얽힘, L-003으로 수용 |

---

## §B 자료구조 설계

### B.1 원장 스키마 진화

현행(v1) → 목표(v2):

```yaml
# v1 (현행)
bypass: true
model: claude-opus-4-6
last_profile: moai-adk

# v2 (목표) — v1의 상위집합
bypass: true                 # 미지·레거시 키: read-modify-write 로 보존
model: claude-opus-4-6
last_profile: moai-adk       # 전역 폴백 (유지)
projects:                    # 신규 — 프로젝트 절대경로 → 프로필명
  /Users/goos/MoAI/moai-adk-go: moai-adk
  /Users/goos/MoAI/mo.ai.kr: mo.ai.kr
```

**하위호환 방향 2개를 모두 만족한다.**

- **전진(v1 원장 → v2 코드)**: `projects` 키 부재 → 맵 조회가 미스 → 전역 `last_profile` 폴백. 업그레이드 전과 동일 동작(REQ-PM-004).
- **후진(v2 원장 → v1 코드)**: v1 해석기는 `last_profile` 만 읽고 미지의 `projects` 키는 무시한다. v2 기록기가 항상 두 곳에 쓰기 때문에(REQ-PM-002) `last_profile` 이 항상 최신이므로, 다운그레이드해도 "마지막 런치가 이긴다"는 v1 의미가 그대로 살아난다. 이것이 `last_profile` 을 남기는 설계적 이유다.

### B.2 키 정규화 함수

```
normalizeProjectKey(projectRoot) =
  ""                          if projectRoot == ""
  EvalSymlinks(Clean(root))   if EvalSymlinks succeeds
  Clean(root)                 otherwise
```

정규화는 **쓰기와 읽기 양쪽에 동일하게** 적용한다. 한쪽만 적용하면 macOS `/var` ↔ `/private/var` 비대칭으로 항목이 영원히 매치되지 않는다.

**알려진 불완전성**: 성공 경로와 폴백 경로가 서로 다른 문자열 네임스페이스를 만든다. 한 프로젝트가 최악의 경우 두 키를 차지할 수 있으며(`plan.md` §D1), 대소문자 무시 파일시스템도 정규화하지 않는다. 동작상 무해하므로 한계 L-002로 수용한다.

---

## §C API 표면 설계

### C.1 3쌍 대칭 구조

| 기존 (유지, 얇은 래퍼) | 신규 (project-aware) | 반환 |
|------------------------|----------------------|------|
| `ResolveLaunchProfile(profileName)` | `ResolveLaunchProfileForProject(projectRoot, profileName)` | `string` |
| `RecordLastUsedProfile(name)` | `RecordLastUsedProfileForProject(projectRoot, name)` | `error` |
| `GetCurrentName()` | `GetCurrentNameForProject(projectRoot)` | `string` |

신규 (술어, 쌍 없음): `HasClaudeConfig(name) bool`.

**`projectRoot == ""` 의 계약** — 모든 신규 함수에서 동일한 의미를 갖는다:

- 해석: `projects` 맵을 건너뛰고 전역 `last_profile` 만 본다.
- 기록: `projects` 맵을 건드리지 않고 `last_profile` 만 쓴다.
- 어느 경우에도 에러가 아니다(REQ-PM-007).

이 계약 덕분에 기존 3함수를 `...ForProject("", x)` 한 줄 위임으로 축소해도 관측 동작이 바이트 단위로 동일하다.

### C.2 해석 우선순위 (REQ-PM-003)

```
ResolveLaunchProfileForProject(root, name):
  1. name != ""                       → return name          # 명시 -p 최우선
  2. MOAI_NO_PROFILE_FALLBACK == "1"  → return ""            # 단일 옵트아웃 (REQ-PM-006)
  3. 원장 로드 실패/파싱 실패          → return ""
  4. root != "" 이고 projects[key] 가
     유효한 이름 + 존재하는 디렉터리   → return projects[key] # 프로젝트 스코프
  5. last_profile 이
     유효한 이름 + 존재하는 디렉터리   → return last_profile  # 전역 폴백
  6. otherwise                        → return ""
```

4번의 "유효 + 존재" 검증 실패 시 5번으로 **폴스루**하는 것이 REQ-PM-008이다. 4번에서 조기 반환하면 고아 항목이 전역 폴백을 가려버린다.

### C.3 기록 동작 (REQ-PM-002, 011)

```
RecordLastUsedProfileForProject(root, name):
  1. name == "" || name == "default"   → error (기존 계약 유지)
  2. !isValidProfileName(name)         → error (기존 계약 유지)
  3. !isDir(baseDir/name)              → error (신규 — REQ-PM-011)
  4. 원장 read-modify-write:
       existing[last_profile] = name
       if root != "": existing[projects][normalizeProjectKey(root)] = name
  5. 원자 쓰기 (temp + Rename)         (기존 계약 유지 — REQ-PM-005)
```

3번이 신규 가드다. 이 가드만 단독으로 넣으면 최초 `-p <new>` 런치가 깨지므로 §D의 순서 재배열과 반드시 함께 간다.

---

## §D 제어흐름 설계 — `unifiedLaunchDefault` 재배열

### D.1 재배열 전후

```
BEFORE                                  AFTER
─────────────────────────────────       ─────────────────────────────────
1 resolveMode                           1 resolveMode
2 ResolveLaunchProfile(name)  ← 전역만   2 findProjectRootFn()      ← 앞으로
3 findProjectRootFn()                   3 ResolveLaunchProfileForProject(root, name)
4 applyXxxMode(root, profileName)       4 applyXxxMode(root, profileName)
                                        4.5 EnsureDir(profileName) + 고지  ← 신규
5 RecordLastUsedProfile(orig)           5 recordLastProfileFn(root, orig)
6 launchClaude(profileName)             6 launchClaude(profileName)
```

**4.5단계와 5단계의 게이트 변수가 다르다.** 4.5는 `profileName`(해석 후), 5는 `originalProfile`(사용자 입력). 4.5를 `originalProfile` 로 게이트하면 bare 런치가 `projects:` 를 통해 fresh 프로필로 해석되는 경로에서 고지 횟수가 0이 된다 — 결함 3을 재현하는 바로 그 경로다(`plan.md` §D4 게이트 변수 표, 감사 iter-2 N1).

두 개의 변경이 있다: **(a) 2·3단계 자리 교환** — project-scoped 해석이 `root` 를 필요로 하기 때문. **(b) 4.5단계 신설** — 기록이 디렉터리 존재를 전제하기 때문.

### D.2 재배열의 정확성 조건

| 조건 | 확인 |
|------|------|
| 4단계가 소비하는 `profileName` 이 "해석 완료된 값" 이라는 의미를 유지한다 | 3단계가 4단계보다 앞이므로 유지 |
| `root` 획득 실패 시 동작이 바뀌지 않는다 | 조기 반환 유지 (RESOLVED-3). 단 해석보다 앞서므로 실패 시 해석 자체가 실행되지 않는다 — 종전에는 해석 후 실패했으나, 어느 쪽이든 런치는 동일하게 중단되므로 관측 동작 동일 |
| `CLAUDE_CONFIG_DIR` 설정 시점이 모드 설정과 충돌하지 않는다 | 4.5단계는 4단계 이후이므로 GLM/CG env 주입 완료 후 |
| 기록 대상 이름과 생성 디렉터리가 항상 같다 | `originalProfile != ""` 일 때 `resolved == profileName == originalProfile` (§D.3) |

### D.3 기록 시점 불변식 (REQ-PM-015)

```
originalProfile != ""  ⟹  ResolveLaunchProfileForProject(root, originalProfile) == originalProfile
                       ⟹  profileName == originalProfile
```

해석 함수의 1번 규칙(명시 `-p` 최우선)이 이 불변식의 근거다. 따라서 **4.5단계가 보장하는 디렉터리(`profileName`)와 5단계가 기록하는 이름(`originalProfile`)은, 기록이 일어나는 모든 경우 동일하다.** 두 값이 갈라지는 유일한 경우는 `originalProfile == ""`(bare 런치)이며 그때는 5단계 기록 자체가 일어나지 않으므로 불변식이 깨지지 않는다. `@MX:ANCHOR` 로 고정한다.

게이트 변수가 단계별로 다르다는 점은 §D.1 말미와 `plan.md` §D4 게이트 변수 표를 따른다 — 4.5는 `profileName`, 5는 `originalProfile` 이다.

### D.4 고지 발화 지점이 하나여야 하는 이유

재배열 후 `EnsureDir` 호출 지점은 둘이다(4.5단계 신설분 + `launchClaudeDefault` 기존분). 고지를 "`EnsureDir` 직후"라는 규칙으로 표현하면 두 곳 모두에 적용되어 이중 발화한다. 따라서 규칙을 **위치**로 못 박는다: 4.5단계 단 하나.

고지 로직을 `EnsureDir` **안에** 넣는 대안은 더 나쁘다 — `EnsureDir` 는 런치 전용이 아니므로 비런치 경로로 고지가 새어 나간다.

---

## §E 테스트 가능성 설계

### E.1 시임 4종

| 시임 | 기존/신규 | 용도 |
|------|-----------|------|
| `findProjectRootFn` | 기존 | 프로젝트 루트 고정 (AC-PM-009/010/017/018) |
| `launchClaudeFunc` | 기존 | exec 대체 + 전달 인자 관측 (AC-PM-017의 핵심) |
| `recordLastProfileFn` | **신규** | 기록 실패 주입 (AC-PM-010) |
| `launcherStderr` (`io.Writer`) | **신규** | 런치 경로 stderr 포착 (AC-PM-010(c), AC-PM-018 세 케이스) |

`recordLastProfileFn` 은 파일 권한 의존을 제거하기 위한 것이다. 권한 기반 실패 유도는 root 실행에서 무력화되고 Windows CI 에서 판정 불가다(제약 C5).

`launcherStderr` 는 **없으면 AC-PM-010(c)와 AC-PM-018이 판정 불가**이기 때문에 필요하다. `warnNoModelResolved(w io.Writer, ...)` 라는 기존 선례는 **함수 시그니처만** 공급한다 — 그 호출 지점이 `os.Stderr` 를 하드코딩하므로 런치 경로 출력은 포착되지 않는다. 5단계 기록 실패 경고도 같은 방식으로 `os.Stderr` 에 직접 쓴다. `internal/cli` 패키지에는 현재 `io.Writer` 시임이 존재하지 않으므로 신규 선언이 필요하며, 선언 형태는 기존 5개 시임(`unifiedLaunchFunc` / `launchClaudeFunc` / `findProjectRootFn` / `newDetectorFn` / `injectTmuxSessionEnvFn`)과 동일한 패키지 레벨 `var` 패턴을 따른다. 선언 위치는 `plan.md` M4-2.

### E.2 판정 계층 분리

| 계층 | 판정 대상 | AC |
|------|-----------|-----|
| 함수 | 원장 스키마·해석 우선순위·기록 가드 | 001-008, 011, 019, 020 |
| 런치 | 호출자 배선·순서·발화 횟수 | 009, 010, 017, 018 |
| 트리 | 부재-grep·상수화·회귀 | 012, 013, 015, 016 |
| 테스트 인프라 | 샌드박스 격리 | 014, 021 |

iter-1 감사의 치명 결함(D1)은 **함수 계층 AC만 있고 런치 계층 AC가 없었기 때문에** 탐지되지 않았다. 계층 분리를 명시적으로 유지하는 것이 이 설계의 회귀 방어다.

### E.3 샌드박스 실행 순서 문제

Go 는 테스트 파일을 사전순 실행한다. 따라서:

```
main_test.go              ← TestMain + 가드 (오염 이전)
preferences_test.go
profile_test.go           ← TestGetBaseDir_Default 가 BaseDirOverride 오염
zz_sandbox_guard_test.go  ← 후행 가드 (오염 이후)   ← 신규
```

`main_test.go` 의 가드만으로는 오염을 탐지할 수 없다(항상 먼저 실행되어 통과). 후행 가드가 실제 판정자다. 파일명 `zz_` 접두사는 사전순 최후 정렬을 위한 관례적 수단이다.

---

## §F 위험과 완화

| 위험 | 완화 |
|------|------|
| 기록만 바꾸고 해석을 안 바꿈 → `projects:` 가 dead write | AC-PM-017 (런치 계층), plan.md §G 첫 안티패턴 |
| 고지 이중 발화 | AC-PM-018 (횟수 판정), 발화 지점을 위치로 고정 |
| 원자 쓰기가 조용히 사라짐 | AC-PM-020 (프리미티브 grep + 부분상태 테스트) |
| 샌드박스가 공허하게 통과 | AC-PM-021 (후행 가드 + 반증 왕복) |
| 동시 런치로 프로젝트 항목 유실 | 완화하지 않음 — 한계 L-003으로 수용 |
| 정규화 폴백 이중 키 | 완화하지 않음 — 한계 L-002로 수용 |
