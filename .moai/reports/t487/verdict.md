# t487 판정 — dirty `.claude/settings.json` 작성자 귀속(SPEC-SETTINGS-ORIGIN-001)

> 카드: t487 · 브랜치: `WT-settings-origin` · 측정 트리: 본 워크트리(base `25a3212a9`, HEAD `c96f854e5`)
> 관측: 2026-09-05/06 · 방법: 전 워크트리 shape 스윕(읽기 전용, cross-tree git 0) + 저장소 코드경로 전수 목록 + 템플릿/추적 이력 키 순서 측정 + 외부 유입 후보 열거
> 보존 사본 기준: `b669972dc738d1bf925281dcc90f152e`(t452 dirty) — pre-flight에서 재확인(아래 Evidence P0)

## 판정 요약

**Q1 — 이 저장소의 어떤 코드 경로도 dirty 모양을 생산할 수 없다(전수 목록 + 교차확인).**
모든 저장소 내 작성자는 (a) 렌더링된 템플릿 순서(hooks-먼저), (b) Go map 재직렬화(알파벳 정렬), (c) permissions 영역 외 바이트 보존 외과적 치환, 셋 중 하나만 낸다. dirty 순서는 이 셋 어느 것도 아니다.

**스윕 결정적 발견 — dirty 계열 인스턴스는 t452 하나가 아니라 둘이다.**
`t334` 워크트리의 tracked `.claude/settings.json` 워킹 사본이 **현재도 dirty**(md5 `4f455d9425a396d38c202f2614bc918f`, HEAD blob `437834679...`과 불일치). dirty 계열의 11-키 순서·결측 3키·narrow matcher·sto=1을 전부 공유하되 ask=`[]`(t452판은 `[sudo]` 1개)로 t452판보다 더 얇다. → **H4(계통적 작성자) 확정: 우연 아닌 반복되는 클래스.** 나머지 78개 트리는 전부 자기 HEAD blob과 바이트 일치(blob 대조 전수 실측).

**해석 (a) 대 (b) 판정**: t334 파일 mtime(2026-08-28 03:00)이 트리 생성(2026-08-27 23:59)보다 **3시간 뒤** — 워크트리 생성 시점 복제(b)가 아니라 **이미 존재하는 트리에 대한 세션 중간 쓰기(a)**. (참고: 깨끗한 트리 중에도 mtime>생성시각이 있어 시간축 단독은 판별식이 아니다 — blob 대조가 판별식이고, 그것이 두 인스턴스 모두에서 dirty 판정을 준다.)

**H1–H4 갱신**: H1(CC 런타임 재작성) — 이 머신에서 CC가 재작성한 프로젝트 settings.json 4건(mo.ai.kr·moai-cowork·MINK·agent-latte)은 전부 **알파벳 정렬** 순서; dirty는 정렬 아님 → 표준 CC 재작성 경로는 기각(역사적 CC 버전의 순서는 로컬에서 검증 불가 — Gap). H2(낡은 moai 렌더/바이너리) — 커밋된 모든 상태(템플릿 93 커밋 이력, tracked 64 커밋, 릴리스 태그 v3.0.0/v3.1.0/v3.1.2, 2025 TS 시대, 디스크 상 7개 시대 표본 바이너리의 임베디드 템플릿) 어디에도 dirty 순서 없음 → 기각(커밋 안 된 로컬 빌드 잔여 가능성은 Gap). H3(수기/AI 재구성) — **잔존 최유력**: dirty-t452는 당대 develop과 키 **집합**이 동일하고 순서만 다르며(재직렬화가 아니라 재구성의 서명), ask 6→1로 내용이 줄었고(기계적 재작성은 보존), 두 인스턴스가 서로 다른 내용(두 번의 별도 쓰기). t485 C4에 따라 프로세스 특정은 하지 않는다. H4 — 위대로 확정.

## Claim (주장)

1. **C1 (Q1)**: 이 저장소의 `.claude/settings.json` 작성 코드 경로는 총 7개(아래 E3 표)이며, 어느 것도 §1 fingerprint(키 순서·ask 1건·narrow matcher·sto=1)를 생산할 수 없다. cross-check(2차 패턴, 저장소 루트 경계)는 목록 밖 경로를 0건 노출.
2. **C2 (스윕)**: 79개 트리(primry 체크아웃 포함) 전수 스윕에서 dirty-family shape 적중 1건(t334), dirty md5 `b669...` 정확 일치 0건, HEAD-blob 불일치 1건(t334 동일). t452 트리 자체는 병합 후 복원되어 현재 깨끗함(t480 기준 인용).
3. **C3 (H4)**: dirty 계열 작성자는 반복되며(2인스턴스/6일/서로 다른 트리), 워크트리 생성 시점이 아니라 세션 중간에 쓴다.
4. **C4 (Q2)**: 외부 유입 축에서 — CC 표준 재작성은 순서 축에서 기각(로컬 실측), 디스크 바이너리 7개 시대 표본 기각, 타 체크아웃 13건 전수 불일치, npm TS 선대 템플릿은 **본 spawn에 web 도구 없어 미검증 — 레인 플래그**. 잔존 최유력은 수기/AI 재구성(H3, 귀속 불가 — C4-dead-axis 존중).
5. **C5 (Q3)**: 재발 방지 권고 정확히 1건 — 카드 워크트리 병합 창에 `.claude/settings.json` drift 단정(pre-merge assertion) 추가. 프로세스 레벨(t485 C4 준수).

## Evidence (증거 — 명령 + 실측 출력, 2026-09-05/06 본 세션, 본 트리)

### P0 — pre-flight (보존 사본 무결성)

```
$ git branch --show-current && git rev-parse --short HEAD
WT-settings-origin / c96f854e5
$ md5 -q .../t480/preserved-copies/settings.json.worktree-dirty .../settings.json.develop
b669972dc738d1bf925281dcc90f152e / 568a3d3a32a360731d1d02f686f27d24   (SPEC §1 값과 일치)
$ git worktree list | wc -l → 79
```

### E1 — 스윕 분포 (79 트리, shape 기반, cross-tree git 0)

명령: `run-sweep.sh`(보고서 동봉; md5/jq/grep/stat만, git 0회) → `sweep.tsv` 1행/트리.

```
md5 빈도: 568a3d3a…×28 · c58ff4b1…×20 · 3d4e7057…×17 · 437834679…×6 · 9c970839…×3
        · d9c2ea63…×2 · 4f455d94…×1(t334) · 0f9a934d…×1 · 0f194c9d…×1
NOFILE 0건 · dirty md5 b669… 일치 0건
shape 클러스터(kom/결측키/sto/narrow/extended):
  50트리 N/absent/1/1/0 (구 tracked 상태 계열) · 28트리 N/absent/3/0/1 (최신 계열)
  · 1트리(t334) Y/absent/1/1/0  ← dirty 계열 키 순서 유일 일치
```

키 순서 3축 실측(전 FileType 동일 명령 계열 `jq -r 'keys_unsorted|join(",")'`):

```
dirty 보존사본 : $schema,respectGitignore,cleanupPeriodDays,skillListingBudgetFraction,env,attribution,permissions,hooks,statusLine,outputStyle,showThinkingSummaries
t334 워킹사본  : (위와 동일 — 전체 일치)
primary/develop: $schema,hooks,statusLine,skillListingBudgetFraction,showThinkingSummaries,cleanupPeriodDays,outputStyle,env,permissions,attribution,respectGitignore
```

### E2 — t334 dirty 확정 (blob 대조, 본 트리 git show — 읽기 전용)

```
$ git worktree list --porcelain | grep -A2 t334
  HEAD 4837b0521… (WT-cli-test-cwd, 2026-08-28 07:43 최종 커밋)
$ git show 4837b0521:.claude/settings.json | md5 -q
  437834679fcc2435761e264cec2ffc15   ← 워킹사본 4f455d94… 와 불일치 = dirty
  (blob 키 순서 = 현행 순서, ask 6엔트리)
t334 file: 23,556B · ask=[] · narrow matcher(l.287) · sto 언급 1회(l.306)
mtime 2026-08-28 03:00:07 / 트리 birthtime 2026-08-27 23:59:04 (3시간 뒤 쓰기)
diff(dirty-t452 보존사본 ↔ t334): t334가 navigator·AskUserQuestion pre-tool 훅 부재, ask 비어 있음 — dirty-t452가 t334판보다 내용이 많음(별도 2회 쓰기, 동일 순서 계열)
```

전수 blob 검증(`verify-blobs.sh` → `blob-check.tsv`): **78 CLEAN_MATCHES_HEAD / 1 DIRTY_VS_HEAD(t334)**. 즉 dirty-family 외의 숨은 변형(toast되지 않은 다른 dirty)은 0건.

### E3 — Q1 코드경로 전수 목록 (작성자/판정)

| # | 경로 (file:line) | 메커니즘 | 직렬화 순서 | dirty 생산 가능? |
|---|---|---|---|---|
| W1 | `internal/cli/update/deploy/deploy.go:364,565` (+`CleanMoaiManagedPaths`:29, 항목 등록 :59-60) | `moai update` 임베디드 템플릿 재배포 | 렌더 순서(hooks-먼저) | **NO** — E4/E5에서 어떤 커밋 템플릿 상태도 dirty 순서 아님 |
| W2 | `internal/cli/init.go:861` + `internal/core/project/initializer.go` | `moai init` 렌더 | 동일 | **NO** (같은 템플릿 원천) |
| W3 | `internal/cli/update/merge/merge.go:199,219,232,240` | 3-way merge 재작성 | merge 엔진 `nIndent(map)` → **알파벳** (`internal/merge/strategies.go` 실측) | **NO** — dirty는 정렬 아님 |
| W4 | `internal/migration/migrations/m002_settings_cleanup.go:125` | hooks 정리 map round-trip | `json.MarshalIndent(map)` → 알파벳 | **NO** |
| W5 | `internal/cli/update_deny_migration.go:100` | deny 항목 제거 map round-trip | 알파벳 | **NO** |
| W6 | `internal/config/toolpolicy/codegen.go:247`, `tier_render.go:149` (호출 `internal/cli/tool_policy.go:119,151`) | permissions 영역 외과적 치환 | 나머지 바이트 보존( `codegen.go:120-125` splice) | **NO** — 최상위 키 순서 불변 |
| W7 | `internal/cli/update.go:958` | **전역** `~/.claude/settings.json` env 보장(map round-trip) | 알파벳 | **NO** — 대상이 project tracked 파일 아님(축 미달) |

비-대상(참고): `internal/cli/settings.go:67` `mutateSettingsLocal` = `settings.local.json` 전용(알파벳). `internal/core/project/autonomy_bundle.go` = user-scope. 판독 전용(무기록 확인): `doctor_hook_wiring.go:48-49`, `hook_stop_goal.go`, `glm.go:170`(주석), `launcher.go`(주석+local 파일), `permission/stack.go`, `profile/preferences.go`, `web/app.go`(YAML sections), `config/resolver.go` 등.

스크립트/CI/훅: `scripts/`, `.github/workflows/`, `Makefile`, `.claude/workflows/`, `.claude/hooks/**`에서 settings.json 쓰기 리다이렉션/cp/mv/tee **0건**(rg 실측, E6).

**교차확인(AC-001, 2차 패턴, 저장소 루트 경계)**: 상수 역추적(`defs.SettingsJSON`/`SettingsLocalJSON`/`UserSettingsFileName` → 사용처 전수) + internal/ 밖 전역 `settings.json` 스윕(문서 docs-site 텍스트만) → 목록 밖 작성 경로 **0건**. 누락 0 = FAIL 해당 없음.

### E4 — 템플릿/추적 이력 키 순서 (H2 축, M3)

`.tmpl` 최상위 키(감사자 검증 형식 `grep -E '^  "[a-zA-Z$]+"'`, `git show <state>:…`):

```
e21d85d8d(2026-02-07 생성기 전환): hooks,statusLine,outputStyle,cleanupPeriodDays,env,permissions,teammateMode,sandbox,attribution,…
v3.0.0(c6f86d097) / v3.1.0(ed04e40e6) / v3.1.2(4b2f203f): $schema,hooks,statusLine,skillListingBudgetFraction,showThinkingSummaries,cleanupPeriodDays,model,outputStyle,env,permissions,attribution,respectGitignore,includeGitInstructions,plansDirectory
2026-07-02/07-10/07-17/07-25/08-01 표본 + HEAD: 동일 계열(hooks-먼저) — dirty 순서 없음
tracked .claude/settings.json(64 커밋 이력 표본 8개: 02-03·02-07·02-14·02-21·05-09·06-15·07-25·08-10): 전부 별형, dirty 순서 없음
TS 시대(2025 태그 backup-before-v0.0.1): env,hooks,permissions — 불일치
```

→ **dirty 순서를 가진 커밋된 상태는 존재한 적 없음.** dirty-t452의 키 **집합**은 현행과 동일(순서만 다름) = 재구성 서명.

### E5 — 디스크 바이너리 임베디드 템플릿 (M4)

```
$ strings <bin> | grep -A1 '"respectGitignore"'
primary bin/moai(2026-08-27 03:26): "respectGitignore": true, → "includeGitInstructions": true,   (현행 tail 순서)
~/go/bin/moai(2026-09-03) / backup-ed04e40e6(08-16) / 1787295234(08-21) / bak-20260810: 동일(tail 순서)
moai.backup.1770202209(2026-02-04): respectGitignore/skillListingBudgetFraction 문자열 0건(그 이전 세대)
```

→ 임베디드 템플릿이 dirty 순서를 갖는 바이너리: 표본 전수 불일치. `moai update` 재배포 경로로는 dirty가 나올 수 없음(커밋 상태 기준).

### E6 — 타 프로젝트/체크아웃 (M4)

```
CC-재작성 추정 4건(mo.ai.kr·moai-cowork·MINK·agent-latte): $schema,attribution,cleanupPeriodDays,env,hooks,… (알파벳 정렬 — moai 렌더와 상이한 단일 작성자 클래스)
moai-렌더 4건(moai-3.0-test·youtube·school·copythat): 템플릿 순서 — dirty 아님
형제 변형(moai-adk-codex·-opencode·moai-ask·moai-rank): hooks-먼저 변형 — 불일치
전역 ~/.claude/settings.json(md5 97bad773…): CC 관리 키(대략 정렬) — 불일치
```

→ dirty 순서를 낸 외부 체크아웃/작성자: 이 머신에서 0건.

### E7 — baseline 재사용 (AC-008)

t480의 6대 소거(tracked 전 이력 ask축·생성기 전환·릴리스 템플릿 5종 ask축·전역/로컬 ask-null·설치 바이너리 버전)는 **인용만** — 본 판정 Evidence에 재실행 0건. §1 fingerprint는 H1–H4 가설로만 사용(§판정 요약의 갱신은 본 카드 실측에 근거).

## Baseline-attribution (귀속)

- 모든 측정은 2026-09-05/06, 본 세션, t487 워크트리(`25a3212a9` 베이스 + 카드 커밋 `c96f854e5`)에서 실행한 위 명령들의 실측 출력이다. 스윕 원시 데이터: `.moai/reports/t487/sweep/`(tree-list.txt·tree-heads.txt·sweep.tsv·blob-check.tsv·run-sweep.sh·verify-blobs.sh).
- blob 대조(`verify-blobs.sh`)의 git 호출은 전부 본 트리에서의 `git show <sha>:<path>` 읽기 전용 객체 판독이다(cross-tree git 0회 — harness guard 준수).
- "CC가 프로젝트 settings.json을 알파벳으로 재작성한다"는 **추론**이다: 4개 프로젝트가 moai 렌더 순서(대조군 4개 실측)와 다른 동일한 정렬 순서를 공유하는 것으로부터의 추론이며, CC 소스를 직접 읽은 근거가 아니다.
- 바이너리 검사는 시대 표본(2026-02-04·08-03·08-10·08-16·08-20·08-21·08-27·09-03)이지 전수(~140개 backup)가 아니다.

## Gaps (미검증)

1. **npm TS moai-adk 선대의 배포 템플릿** — 본 spawn에 web 도구가 없어 미검증. 레인 플래그: `npm` 레지스트리의 moai-adk 과거 버전 settings 템플릿 원문 대조 필요. (다만 본 리포 TS 시대 tracked 상태는 불일치 실측됨.)
2. **작성 프로세스의 이름** — t485 C4 dead axis. H3(수기/AI 재구성)은 잔존 최유력이지만 귀속 아니다: 두 인스턴스를 쓴 세션/에이전트 특정은 이 카드의 방법으로 불가능.
3. **역사적 CC 버전의 직렬화 순서** — 현재 CC의 재작성이 정렬임은 로컬 실측이나, 과거 버전이 다른(비정렬) 순서를 썼을 가능성은 로컬에서 검증 불가.
4. **커밋되지 않은 로컬 빌드 바이너리** — 커밋 상태 기준 전수 기각이지만, 커밋 안 된 트리에서 빌드된 바이너리의 임베디드 템플릿은 원천적으로 재구성 불가.
5. 디스크 backup 바이너리 전수(~140개) 미검사 — 시대 표본 8개만 실측.
6. t334 dirty 사본은 **현존하며 미보존 상태** — 본 카드는 기록만 하고 수정하지 않았다(바이트 보존 준수). 별도 보존 복사는 레인/리드 결정 사항.

## Residual-risk (잔여 위험)

- 작성자가 여전히 활동 중일 수 있다(H4: 6일 간 2회, 79트리 기준). 세 번째 인스턴스는 임의 시점에 발생할 수 있다.
- t334 워크트리의 dirty 사본이 병합 창에 들어가면 t452와 같은 병합 거부가 재연된다 — 병합 전 drift 점검이 없으면.
- 스윕은 시점 측정이다(2026-09-05/06). 이후 발생한 쓰기는 본 판정에 없다.
- H3 가설이 옳더라도 "AI 세션이 썼다"는 일반적 진술에 그친다 — 재발 방지는 특정이 아니라 검출·차단으로 설계해야 한다(C5 권고의 근거).

## Q3 — 재발 방지 권고 (정확히 1건, t485 C4 준수)

**카드 워크트리 병합 창(git-flow 레인 프로토콜의 develop 병합 직전 단계)에 `.claude/settings.json` drift 단정을 상시 절차로 추가한다**: `git --no-optional-locks status --porcelain -- .claude/settings.json` 출력이 비면 통과, 비지 않으면 **사본 보존(path+md5) + 리드 blocker 보고**(자동 복구 금지).

- 근거(premise, 실측): ① 79트리에서 유일하게 발견된 dirty(t334)는 9일간 아무도 눈치채지 못했다 — 병합 창 검사만이 보증된 검출 지점이다(t452에서는 병합 거부가 우연히 잡았음). ② dirty tracked 파일이 병합에 실리면 낡은 내용이 develop에 흘러들 수 있다(t480은 다행히 역방향이었음). ③ `--no-optional-locks` 형식 필수(t485 C1: 일반 `git status`는 index.lock을 잡는다).
- C4 준수: config 파일 파기가 아니라 프로세스 레벨 검출 절차이며, 작성자 귀속을 요구하지 않는다. 구현은 별도 카드.

## AC 준거

| AC | 상태 | 근거 |
|---|---|---|
| AC-001 | PASS | E3 표 + 2차 패턴 교차확인(누락 0) |
| AC-002 | PASS | E3 표 'dirty 생산 가능?' 열 전 항목 marker-연계 판단 |
| AC-003 | PASS | E1/E2 — 79 열거 = 79 점검, dirty 1건 path+md5+3 marker, primary 포함, 수정/폐기/push 0 |
| AC-004 | PASS | M4 실행(E5·E6) + Gap 1(npm not-fetchable 플래그) — "repo 생산자 없음" 분기 근거 = C1 |
| AC-005 | PASS | Q3 섹션 1건, premise 명시, 프로세스 레벨 |
| AC-006 | PASS | 본 문서(5 섹션) |
| AC-007 | PASS | 정량·부정 주장 전항에 명령+출력 병기, 빈 출력≠0 명시(스윕 NOFILE 0건은 실측) |
| AC-008 | PASS | E7 — t480 재실행 0건, fingerprint는 가설로만 사용 |

## lane-8 후속 측정 (run 종결 후 레인-레벨 추가, 2026-09-06)

> 아래는 manager-develop의 run 판정(HEAD `c8086a384` 시점 판정서 본문)에 레인이 나란히 붙인 정정·추가 측정이다. 본문 원문은 지우지 않는다(`판정 옳아도 논거 갈림` 정격).

### L1 — Gap 1 해소: npm TS 선행판 실측 → 소거

run은 Gap 1을 "web 도구 없어 미검증 — 레인 플래그"로 남겼다. 레인(웹 도구 보유)이 webReader로 닫았다:

```
$ webReader https://www.npmjs.com/package/moai-adk
  moai-adk v0.2.29 · TypeScript · 최종 배포 약 1년 전 · repo: modu-ai/moai-adk
$ webReader https://api.github.com/repos/modu-ai/moai-adk/git/trees/main?recursive=1
  .claude/settings.json blob 존재 (665d646e…, 9,248B) — TS 시대 산출물이 repo 루트에 커밋돼 있음
$ webReader https://raw.githubusercontent.com/modu-ai/moai-adk/main/.claude/settings.json
```

실측 형상(TS 시대): 최상위 키 순서 `cleanupPeriodDays, env, permissions, hooks, statusLine, outputStyle, companyAnnouncements, enabledPlugins` — `$schema`·`skillListingBudgetFraction`·`respectGitignore`·`showThinkingSummaries` **모두 부재**; 훅은 Python(`uv run … .py`, `.sh` 아님); `permissions.ask` 19엔트리; `outputStyle: "R2-D2"`.

→ dirty 지문($schema 선행 11-키 Go-시대 집합·`.sh` 훅·narrow matcher·sto=1)과 **전축 불일치**. npm TS 선행판은 dirty 출처 후보에서 소거된다(Gap 1 닫힘). 단 HEAD 1상태만 실측했으므로 TS 히스토리컬 버전별 템플릿은 잔여 가능성으로 남는다 — 우선순위 낮음: dirty 사본이 전부 Go-시대 마커($schema URL·skillListingBudgetFraction 0.02 등)를 지니므로 TS-시대 출처와는 내용적으로 모순.

### L2 — t334 dirty 사본 보존 반출 (Gap 6 정정)

```
$ cp <t334>/.claude/settings.json .moai/reports/t487/preserved-copies/settings.json.t334-dirty
$ md5 -q .moai/reports/t487/preserved-copies/settings.json.t334-dirty
4f455d9425a396d38c202f2614bc918f   ← 원본 지문과 정확 일치 (원본 트리는 건드리지 않음)
```

- 시크릿 스캔: 히트 3건 전부 오탐 — `"Read(./secrets/**)"`(l.135)·`"Edit(./secrets/**)"`(l.139)는 권한 항목의 경로 토큰, 40자+ 문자열 1건은 훅 경로(`ta**sk-c**ompleted.sh`의 `sk-c` 부분매치). **실제 시크릿 0건 → 커밋 가능.**
- env 키 4개는 현재 템플릿(`settings.json.tmpl` l.409-419, `MOAI_CONFIG_SOURCE` l.412)과 동일 집합 — dirty 전용 추가 키 없음.
- t334 vs t452-dirty 정규화 비교(레인 실측): 두 변종은 같은 패밀리, 유일한 정규화 차이는 `ask` 리스트 1곳(`[]` vs `["Bash(sudo:*)"]`).
- 이 반출로 본문 Gap 6("현존하며 미보존")은 해소된다. 원본 t334 트리는 [HARD]대로 미수정·미삭제.

### L3 — birth==mtime 통째 교체 서명 + 03:00시 관찰

t334 파일의 birth time == mtime(2026-08-28 03:00:07)이며 트리 생성(2026-08-27 23:59:04)보다 3시간 뒤다. in-place 수정이었다면 birth가 트리 생성시각에 남아야 하므로, birth==mtime은 **temp+rename 통째 교체** 서명이다 — E2의 "세션 중간 쓰기(a)" 판정을 교체 방식 차원에서 강화한다. 또한 t334 쓰기(03:00:07)와 t452 dirty 쓰기 창(2026-09-04 03:52 병합 직전, t480 기록)이 심야 시간대에서 겹친다 — 2 포인트라 **과결론 금지**, 관찰만 기록한다.
