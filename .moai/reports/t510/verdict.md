# t510 판정서 — moai 상태 파일이 의도하지 않은 자리에 쓰인다

카드: t510 (Class B · Tier M · 외부 재현 GH #1694) · 배차: 리드 2026-09-07
브랜치: `WT-state-write-locus` @ `0b1e27877` (= origin/develop, 세션 시작 시 fetch로 재확인)
판정 단계: plan-phase (가설 검증 + 열거 + 격리 재현). 수리는 이 판정 위에서 SPEC으로 진행.

---

## Claim (판정)

1. **두 축은 다른 기전이다 — 수리는 둘이다.** 축 A(홈 오염 341)와 축 B(cwd 오염, GH #1694)는
   서로 다른 리졸버의 서로 다른 결함이며, 관측이 만나는 지점은 없다.
2. **축 A는 이미 수리돼 있다.** 범인은 `internal/cli` todo 테스트 3형제이고, 커밋
   `e7a078970`(2026-09-02 03:13, card t422 fail-loud guard)이 3건을 수리 + 가드를 세웠다.
   오염은 그 커밋 이후 0건이다. 「고쳐짐」이지 「트리거 소멸」이 아니다 (근거는 Evidence §A-6).
3. **축 B는 HEAD에서 살아있고, 격리 프로브로 재현했다.** statusline이 상태를 stdin
   `current_dir`에 기록한다(`workspace.project_dir`을 후보에 두지 않음, git 루트 워크업 없음).
   GH #1694의 226개 stray `.moai`는 이 기전의 직접 귀결이다.
4. 생산 경로의 홈 폴백(`~/.moai/todo/<key>/`, todo_root.go fail-open 설계)은 **의도된 동작**으로
   아직 살아있다 — 비git 임시 디렉터에서 `moai todo`를 호출하면 지금도 홈에 큐가 생긴다.
   이 설계를 임시 디렉터에서 가드할지는 SPEC의 설계 결정 사항이다.

---

## Evidence

### A. 축 A (홈 오염 341) — 범인 특정과 정지 원인

**A-1. 전제 재측정 (리드 수치 대조)**

| 항목 | 리드 값 | 내 측정 | 판정 |
|---|---|---|---|
| `~/.moai/todo` 총 항목 | 345 | **343** (서브디렉터 343 = `001-*` 341 + 비-001 2, 파일 0) | 리드의 345는 `ls -la` alias가 `.`·`..`·total 행을 센 것. 세 번째 값 343이 옳다 |
| `001-*` 디렉터 | 341 | **341** (`/bin/ls` + `find -name '001-*'` 이중 확인) | 일치 |
| origin/develop | 0b1e27877 | 0b1e27877 | 일치 |
| 내부 구조 | `<dir>/.moai/state/todo/backlog.db` | 샘플은 `.moai/state/kanban/backlog.lock+backlog.json` | **둘 다 참** — 기간 내 리네임(아래 A-3)으로 두 구조가 공존. 단일 구조로 일반화하면 틀림 |

- 측정 오염 교정 2건: (a) 내 첫 `^001-` 카운트 0은 zsh 프로필 `alias ls='ls -la'` 때문 — 파이프
  측정은 `/bin/ls`로 재측정해 확정. (b) 리드의 345도 같은 alias 계열 아티팩트. **파이프로 이어지는
  ls 측정은 프로필 alias를 의심해야 한다.**

**A-2. 쓰기 지문 — 범인은 todo add 테스트 패밀리**

- `backlog.json` 209개 — 전부 (`grep '"text"'` 유일 히트) `"text": "first card"`.
- `backlog.db` 132개 — sqlite3 샘플 `SELECT id, text FROM items` → `t1|first card`.
- 209 + 132 = 341 (전 디렉터가 json-era 또는 sqlite-era 둘 중 하나, 단일 패밀리).
- `backlog.lock` 337개.

**A-3. 타임라인 × 구조 지문 — 같은 메커니즘, 두 스토리지 시대**

```
Aug 17~27   kanban 구조 (.moai/state/kanban/)   217개
Aug 27~9/2  todo   구조 (.moai/state/todo/)     124개  ← 8910c337c (08-27 20:06) 리네임 커밋과 정확히 경계가 일치
```
08-27 저녁 바이너리 재설치를 경계로 구조만 바뀌고 오염은 계속 — **리네임을 넘어 살아있던
하나의 메커니즘**이라는 뜻이고, 두 구조 = 두 별개 결함 가설은 기각된다.

**A-4. 「001」 basename의 출처**

`t.TempDir()`의 각 호출은 테스트 부모 아래에 **번호 디렉터(001, 002, …)**를 만든다. t510 격리
프로브 실행에서 직접 관측: `.../TestT510Probe_...2812549133/002/...` — 첫 TempDir 호출이 `001`.
todoFixture의 root = 첫 TempDir = `001` → `TodoQueueProjectKey`가 `basename+sha256(abs)[:4]`(8hex)
로 `001-<hash>`를 만든다. 비-001 2개(`proj-325ca0b6` 08-23, `t203-probe-d7a16ea2` 08-24)는
**비git 임시 디렉터에서 생산 바이너리로 moai todo를 호출한 표본** — 같은 폴백, 다른 basename.

**A-5. 정지 원인 — e7a078970 (2026-09-02 03:13, card t422)**

커밋 메시지가 직접 기술한다: 「three existing tests that set CLAUDE_PROJECT_DIR without a
committed git repo (**falling through to the real home fallback**) are repaired to use
todoFixture. Guard adoption immediately caught all three … including **TestTodoBareInvocationLists,
which wrote a seed add into the home fallback on every run**.」

- 기전 사슬: `CLAUDE_PROJECT_DIR`=비git TempDir → `gitcore.ResolveGitDirs` 실패 →
  `homeTodoQueueRoot` (`~/.moai/todo/001-<hash>`) → store 오픈이 lock+큐 파일 생성.
- 가드 이후 당일 잔여 3건(03:24, 03:28, 03:53) — 병합 전파 지연 가설(가드가 develop에 도달하기
  전까지 레인 브랜치는 구 테스트로 실행). 개별 귀속은 Gaps로 남긴다.

**A-6. 고쳐짐 vs 트리거 소멸의 판별**

Sep 3~7 동안 full `internal/cli` 스위트가 여러 차례 실행됐다(t498·t500·t503 종결 검증 포함,
t500은 오늘 41개 테스트파일 전수 검증). 그 기간 신규 `001-*` 디렉터 **0건** — 트리거가 활발히
돌던 상태에서의 부재이므로 「고쳐짐」 판정의 요건(활발한 트리거 + 부재)을 충족한다. 가드 코드도
HEAD에 존재(`internal/cli/todo_test.go:42` runTodo 게이트 확인).

**A-7. 홈 큐 전수 read-sweep — 리드 행위 귀속 (서드 메커니즘 후보 소멸)**

`~/.moai/todo` 아래 `backlog.db-shm/-wal` 132쌍의 mtime이 2026-09-07 12:03~12:04에 일괄 갱신된
것은 **리드 본인의 read-only sqlite3 루프**(리드 자인)다. birthtime 측정: shm 132개 **전부 오늘
생성** — 원래 polluter는 close 시 wal/shm을 정리했으므로 sweep 이전에는 파일이 없었다. 따라서
(i) 파일 존재·mtime 계수는 sweep 오염, (ii) 디렉터 수(343)·생성 타임라인·db 본체 계수는 무오염.
「홈 큐 전체를 여는 단일 프로세스」는 moai 코드에 존재하지 않는다(grep: homeTodoQueueRoot 참조는
todo_root.go의 단일 키 해석뿐).

### B. 축 B (cwd 오염, GH #1694) — 생산 기전 열거와 HEAD 재현

**B-1. 생산 cwd/env-앵커 상태-쓰기 패밀리 열거**

측정 단계 정리: `.moai/state` 리터럴 조인은 internal/에서 399건(테스트 포함) — 리드의 57건은
측정 단위가 달랐던 것. 판정에 필요한 것은 리터럴 수가 아니라 **쓰기 앵커의 소스**이므로, 생산
경로를 패밀리로 분류했다:

| # | 쓰기 표면 | 좌표 | 앵커 소스 | 쓰는 것 | 비고 |
|---|---|---|---|---|---|
| B1 | 세션 원격청구 | `internal/statusline/context_usage.go:176` `writeContextUsage` / `:278` `resolveProjectDir` (호출 `builder.go:178`) | stdin `workspace.current_dir` → `input.CWD` → `os.Getwd()`. **`workspace.project_dir`은 후보에 없음** (`types.go:184`에 필드 존재) | `<앵커>/.moai/state/context-usage/<sid>.json` 매 렌더 | **#1694 주범.** 세션이 cd한 디렉터마다 1개 생성 |
| B2 | landed 카운트 | `internal/statusline/landed.go:78` `landedCachePath` / `:118` `maybeRefreshLandedCounts` (호출 `builder.go:274` → `backlog.go:24` `resolveBoardRoot` → `resolveProjectDir`) | `worktree.original_cwd` 있으면 그것, 없으면 B1과 동일 사슬 | `<앵커>/.moai/state/landed/counts.json` | B1 형제. 워크트리 세션은 original_cwd로 올바른 앵커 |
| B3 | goal 상태 읽기 | `internal/statusline/builder.go:286` | B1과 동일 리졸버 | 읽기: `<앵커>/.moai/state/goal/...` | **읽기 측 결함** — cd한 세션이 goal 상태를 프로젝트 루트가 아닌 곳에서 읽음 |
| B4 | config 캐시 | `internal/config/cache.go:58` `cacheFilePath(<configDir>/state/config-cache.json)` ← `manager.go:74` `LoadWithCache` | CLI 사슬의 configDir(코드 전반에서 env → Getwd 패턴; 유도 지점은 run-phase에서 RED 테스트로 고정) | `<configDir>/state/config-cache.json` | lane-1 t507 관측(`fixtures/.moai/state/config-cache.json`)과 형태 일치 |
| B5 | 훅 계열 상태 | `internal/hook/path_resolve.go:66` 외, `file_changed.go:243` `mx.NewManager(stateDir)` | `CLAUDE_PROJECT_DIR` → `os.Getwd()` 폴백, 폴백 시 `cwd_fallback:true` 로그 | 훅 상태 | 훅 맥락에선 env가 있어 저위험. 수동 moai 호출 시 cwd 앵커 |
| B6 | 세션 레지스트리 | `internal/session/registry.go:174` | cwd는 **엔트리 필드**로만 사용(기록용), 쓰기 앵커는 훅 ProjectDir | `.moai/state/active-sessions.json` 등 | 쓰기-앵커 아님 — 열거에서 제외 가능 |

**올바른 구현 대조군** (같은 리포 안에 이미 있는 정답 형태):
- `internal/statusline/memory.go:73` `readLLMYAMLContextWindows` — Getwd에서 **조상 방향 워크업**,
  read-only, 실패 시 폴백. B1이 따라야 할 형태.
- `internal/kanban/todo_root.go:96` `primaryCheckoutRoot` — git common dir 기반 primary 해석.
  모든 워크트리/체크아웃에서 하나의 루트를 답하는 기존 시접(`gitcore.ResolveGitDirs`).

**B-2. HEAD에서의 격리 재현 (t.TempDir 랩, HOME 미변경, 커밋 0, 파일 삭제 완료)**

throwaway 프로브 `internal/statusline/t510_probe_test.go` (관측 후 삭제):

```
=== RUN   TestT510Probe_TelemetryAnchorsToCurrentDirNotProjectDir
    t510_probe_test.go:42: probe: record=/var/folders/.../TestT510Probe_.../002/.moai/state/
                          context-usage/t510probe0001.json (visited dir);
                          project dir carries no .moai — cwd-anchored write confirmed at HEAD
--- PASS: TestT510Probe_TelemetryAnchorsToCurrentDirNotProjectDir (0.00s)
ok  github.com/modu-ai/moai-adk/internal/statusline  0.437s
```

- stdin에 `current_dir`(임의 디렉터)과 `project_dir`(다른 임의 디렉터)을 **둘 다** 줬다.
- 결과: `resolveProjectDir`가 `current_dir`를 택하고, 기록이 **visited 디렉터 아래** 착지,
  프로젝트 디렉터에는 `.moai`가 하나도 생기지 않았다. #1694 형태의 최소 재현.
- 덤으로 A-4의 「001」 번호-디렉터 관측(같은 실행에서 001/002)도 직접 확인.

**B-3. #1694 정합성**

제보(moai-adk 3.1.2, macOS, 프로젝트 루트에 정상 `.moai/` 보유): 「그때 만지고 있던 디렉터에
쓴다. 프로젝트 안에 stray `.moai` 226개(내용 있는 것 221)」. B1은 statusline이 **렌더마다**
앵커 디렉터에 레코드를 쓰므로, 세션이 들른 디렉터 수만큼 디렉터가 생긴다. cd를 많이 하는
에이전트 세션 환경과 정확히 일치. lane-1의 t507 관측(context-usage + config-cache 2건)도 B1+B4로
각각 설명된다.

---

## Baseline-attribution

- 코드 판독·프로브·git 판정: 워크트리 `.claude/worktrees/t510`, 브랜치 `WT-state-write-locus` @
  `0b1e27877` (EnterWorktree로 생성, `git merge-base HEAD origin/develop` = HEAD 확인).
- 홈 측정: `/bin/ls`·`find`·`stat`(%Sm/%SB) 기반, 2026-09-07 본 세션에서 직접 실행.
- 프로브: `go test ./internal/statusline/ -run TestT510Probe -v -count=1` — 위 B-2에 출력 전문.
- sqlite 지문: `sqlite3 ~/.moai/todo/001-037389eb/.moai/state/todo/backlog.db 'SELECT id, text
  FROM items LIMIT 5'` → `t1|first card`.

## Gaps (관측하지 않은 것)

- 가드 이후 당일 잔여 3건(09-02 03:24/03:28/03:53)의 개별 귀속 — 병합 전파 지연 가설은 검증하지
  않았다(어느 레인이 어느 브랜치에서 돌았는지의 역추적 미수행).
- B4의 configDir 정확한 유도 지점 — 함수 단위로 고정하지 않았다(run-phase RED 테스트가 할 일).
- B5(훅 계열)는 코드 판독만 하고 실행 재현은 하지 않았다. 저위험 분류는 env-우선 설계(cwd_fallback
  마커)에 기반한다.
- 399 리터럴 중 테스트 파일분의 분류는 생략했다 — 판정은 생산 경로 위주.
- Claude Code가 statusline 프로세스에 `CLAUDE_PROJECT_DIR`을 주는지 여부 — 미검증. stdin 사슬이
  우선이라 어느 쪽이든 B1 재현에는 영향 없다.
- 리드 sweep의 wal/shm 외 부작용 — birthtime으로 디렉터 무오염까지만 확인했다.

## Residual-risk

- t422 가드의 표면은 `runTodo`/`runTodoWithClosedStdin` 2 헬퍼 — `newTodoCmd()` 직접 Execute
  우회 시 무가드(t422 verdict가 이미 지적). 새 테스트 헬퍼 추가 시 재발 가능.
- 생산 홈 폴백이 설계로 살아있어, 비git 임시 디렉터에서의 `moai todo` 호출은 지금도 홈 큐를
  만든다(A-4의 표본 2건). 이 설계를 고치는 것은 본 판정이 아니라 SPEC의 결정 사항.
- 축 B 수리 시 「쓰기 보류」 정책을 택하면 기존에 cwd에 의존해 상태를 읽던 소비자(goal 읽기 등)의
  가시성이 바뀐다 — 수리는 읽기·쓰기·표시 이름 세 관심을 분리해서 설계해야 한다.

---

## 수리 방향 (SPEC 입력)

1. **수리 1 — 축 B (사용자 영향, #1694)**: 상태 앵커의 단일 시접(single seam) 도입.
   우선순위: stdin `workspace.project_dir` → `worktree.original_cwd` → git 루트 워크업
   (`gitcore.ResolveGitDirs` 재사용, B1이 cd 내성을 갖는 형태) → 끝까지 못 찾으면 상태 쓰기
   보류(무프로젝트=무상태). B1·B2·B3·B4를 이 시접으로 통일한다. 표시용 basename은 종전대로
   current_dir에서 따로따로 (읽기·쓰기·표시의 관심 분리).
2. **수리 2 — 축 A**: 코드 수리는 HEAD에 이미 착지(e7a078970). SPEC은 (a) 가드 활성 + canary
   HOME 스윕으로 무오염을 재확인하는 검증 AC, (b) 선택 사항으로 생산 폴백의 임시-디렉터 가드
   (TempDir/`/tmp` 기원 경로에는 홈 큐를 만들지 않음)를 설계 결정으로 다룬다.
3. **GH #1694 회신**: 수리 착지 후 제보자(binsworld)에게 안내 — sync 단계에서 초안 확정.

## 관련 자료

- 오염 디렉터 343개(`~/.moai/todo/001-*` 341 + `proj-*`/`t203-probe-*` 각 1)는 **증거로 보존** —
  이 카드에서 삭제하지 않는다(카드 [HARD] #3). 삭제는 운영자 결정 사항.
- t422 판정서: `e7a078970` 커밋의 `.moai/reports/t422/verdict.md`

---

## 부칙 — GH #1694 제보 본문이 열거를 보강 (B-표 v2)

제보 본문의 파일 계수 표는 본 판정서 B-1 표가 놓친 멤버 2개를 드러냈다. 제보 4종 전부가
아래와 같이 열거 패밀리로 귀속된다 — 외부 증거가 내부 판정과 독립적으로 일치한다.

| 제보 계수 | 멤버 | 귀속 | 근거 좌표 |
|---|---|---|---|
| `state/context-usage.json` 213 | B1 세션 원격청구 | 동일 | `context_usage.go:176`/`:278` |
| `state/config-cache.json` 150 | B4 config 캐시 | 동일 | `config/cache.go:58` |
| `state/github/counts.json` 133 | **B2b — B2의 형제(신규 식별)** | `builder.go:265/267` → `statusline/github.go` `maybeRefreshGitHubCounts(boardRoot)` — B2와 같은 `resolveBoardRoot` 앵커 | `github.go:19` TTL 확인 |
| `state/session-memo.md` 7 | **B7 — 훅 계열 멤버(신규 식별)** | `internal/hook/memo/writer.go:12` `{projectDir}/.moai/state/session-memo.md`, compact 시점 기록(`hook/compact.go:52`) — 훅 사슬(B5 계열)이라 건수가 적다 | compact 훅 |

보강 관찰: 제보자의 「`state/` 아티팩트만 있고 `config/`·`specs/`·`project/`는 전혀 없다」는
관찰이 앵커 진단의 독립 확인이다 — 캐시 쓰기만 cwd 앵커를 타고, 초기화 산출물은 프로젝트 루트에만
생기기 때문이다. 제보 표기(`context-usage.json` 단일 파일, `schema_version: 1`)는 v3.1.2 시대
형태로 보이며 HEAD(`context-usage/<sid>.json`, schema 2, int writer_pid)와 세부가 다르지만
같은 멤버다 — 구버전 스키마 복원은 검증하지 않았다(Gaps).

R1 단일 시접의 적용 대상은 B1·B2(+B2b)·B3·B4이고, B7은 훅 사슬 정리와 함께
run-phase에서 범위 재판정을 받는다.

---

## 정정 — 수리 방향 1의 표시 경로 산문 (plan-audit iter-1 차단 D1+D2)

본 판정서의 「수리 방향 1」은 「표시용 basename은 종전대로 current_dir에서 따로따로」라고
썼으나 **트리는 반대다**: 표시 이름 유도 `extractProjectDirectory`
(`internal/statusline/builder.go:415-438`)는 이미 `workspace.project_dir`를 1순위로
사용하고, `types.go:184`의 주석도 "(used for display)"이다. 판정 작성자가 자신이 읽은
`builder.go:407` 우선순위 주석(project_dir 최우선)을 상태 앵커 문맥에만 적용하고 표시
문맥에는 전사하지 못한 산문 오류이다. plan-audit iter-1(0.85, FAIL — 차단 D1+D2)이 이
오류의 SPEC 전사를 잡았다. 표시 경로의 옳은 서술은 **「기존 extractProjectDirectory
동작 불변(표시 앵커는 project_dir 우선 유지)」**이다. 원문은 삭제하지 않고 이 정정과
나란히 보존한다.

---

## 부칙 3 — 현장 표본 3건 판별 (리드 수집 2026-09-07, sync-audit PASS 이후)

리드가 현장에서 수집한 stray 상태 표본 3건을 본 판정의 기전 분류에 대입했다.
**셋 다 본 SPEC이 수리한 cwd/env-앵커 상태-쓰기 가족이며, 제3 기전은 없다.**

| 표본 | 실물 지문 | 귀속 |
|---|---|---|
| 1. develop 워크트리 **템플릿 소스 트리** stray — primary 보존본을 본 카드 증거경로로 입안(`stray-state-20260907-develop/`) | `config-cache.json`: schema 2, `fingerprint:{}`(섹션 0발견 = 콜드 로드), User.Name 빈값·전부 en, written_at 2026-09-07T03:06:13Z(수리 전 바이너리). **`state/github/counts.json` 병존** | B2b github 카운트가 먼저 `.moai/state/github/`를 MkdirAll → 뒤이은 B4 캐시가 「`.moai`가 이미 존재할 때만 쓴다」(#1568 존재-가드)를 통과해 착지 — "B4 follows B1/B2b"(§E.2)의 실물 확인. 표본 1의 특수성은 **피해자**이다(템플릿 트리 안이라 `TestPublishedIdentitySet`을 적색으로 만들어 lane-10 창을 막음 — **가드는 작동했다**), 기전이 아니다 |
| 2. SPEC 디렉터 stray(lane-4 보존: `.moai/specs/SPEC-UPDATE-MIRROR-HEAL-001/.moai/state/`) | config-cache.json + context-usage/&lt;session&gt;.json 쌍(리드 계측: `.gitignore:228` 무시, 대조군 rc=1로 비공허 확인) | 동일 가족 — B1 생성 후 B4 추종 쌍. **실물 부재**: 창 시점에 어느 레인 워크트리에서도 발견 못 함(폐기 창 경과 추정) — 귀속은 리드 계측 기록과 표본 1의 동형 지문에 근거한다 |
| 3. 카드 원문 관측(홈 341 + 픽스처 2파일) | 본문 §A·§B | 축 A(테스트 격리 — e7a078970 선수리) + 축 B(본 수리) |

수리 후 전환의 실측은 AC-SA-002다: RED(보드 루트가 서브디렉터에 착지) → GREEN(git
common-dir 부모 앵커 — 템플릿 트리 시나리오도 이 경로로 primary 루트에 앵커된다).
sync-audit PASS 95.6(독립 재관측)에서 동일 결론이 유지됐다.
