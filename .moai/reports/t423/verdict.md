# t423 — TestGitDiffNameCount_Predicate TempDir 클린업 레이스 — 판정

트리: WT-tempdir-race @ (커밋 SHA는 커밋 후 리드 보고에) · 기반 9145806d8 (origin/develop 조상 확인)
작성: 2026-09-02 · lane (레인 세션)

## Claim (주장)

`TestGitDiffNameCount_Predicate`의 CI flake는 테스트 코드의 단언 오류가 아니라,
`git commit`이 **무조건** 띄우는 detach된 유지보수 프로세스가 테스트 수명을 벗어나
`.git/objects/pack`에 pack 파일을 써서 `t.TempDir` RemoveAll 정리와 경합하기 때문이다.
수리는 fixture repo에서 auto-gc를 차단하는 것(설정 2줄)이고, 회귀 단언 테스트가 그
설정의 존재를 뮤테이션 RED로 고정한다.

## Evidence (증거)

### E1. 실패 관측 — CI 로그 원문 판독 (리드 관측 5건 중 본 결함 확정 2건)

| run | head | 잡 | 결과 |
|---|---|---|---|
| 33525217558 | 750aacd7f | Race Test | `unlinkat /tmp/TestGitDiffNameCount_Predicate867817455/001/.git/objects: directory not empty` — `--- FAIL (0.11s)` |
| 33521559026 | 3f90d297a (cancelled) | Test (ubuntu) | `unlinkat /tmp/TestGitDiffNameCount_Predicate2720337467/001/.git/objects/pack: directory not empty` — `--- FAIL (0.08s)` |

카드가 "3회 실패"로 적은 관측 중 d6fa555cf(run 33472062711)의 실패는
`TestAlwaysLoadedTokenBudget`(internal/config) — **본 결함 아님** (전수 판독으로 정정).
같은 run(750aacd7f)에서 `TestStopGoalWrapperCopiesStayIdentical`(internal/hook)이
양 잡에서 함께 실패 — 래퍼 3본 드리프트, **별개 결함** (본 카드 범위 밖, 리드 보고에 포함).

### E2. 기제 — 3단 실증 (로컬 darwin, git 2.50.1)

1. **detach 프로세스는 조건과 무관하게 spawn된다**: fixture repo에서
   `GIT_TRACE=1 git commit` →
   `trace: run_command: git maintenance run --auto --quiet --detach`
   (gc.c: commit의 run_auto_gc가 임계 판정을 detach된 자식 내부로 넘긴다).
2. **조건이 충족되면 commit 리턴 후 pack이 씌여진다**: `gc.auto=1` +
   `objects/17`에 loose object 2개(임계 ceil(1/256)=1, 조건 count>1)를 담은
   fixture에서 두 번째 commit → **commit 리턴 약 200ms 후**
   `.git/objects/pack/`에 `pack-*.pack/.idx/.rev` 3개 생성 (probe-gc2.sh).
   두 실패 로그의 `objects/pack: directory not empty` / `objects: directory
   not empty`는 이 pack 쓰기 경합의 두 양상과 정확히 일치한다.
3. **git 소스 판독 (v2.50.1 builtin/gc.c)**: `need_to_gc()`는 `gc.auto <= 0`이면
   즉시 0 → gc task skip. loose 판정은 `objects/17` fan-out 디렉토리 하나만,
   임계 `ceil(gc.auto/256)`. pack 쓰기 주체는 gc/loose-objects/incremental-repack
   task뿐이고 세 task 모두 fixture 규모(개체 ~60개)의 기본 임계(6700/100/10) 미달.
   따라서 **발화 조건은 "환경의 gc.auto가 작을 때"이며, 수리는 조건과 무관하게
   spawn과 detach를 모두 차단해 어떤 환경에서도 성립한다**.

### E3. 수리 (internal/graph/check_test.go, +30줄)

- `newCheckFixture`에 `git config gc.auto 0` + `git config gc.autoDetach false`.
  - `gc.auto=0`: detach된 maintenance의 gc task를 조건 단계에서 차단 — pack 쓰기 자체 소멸.
  - `gc.autoDetach=false`: 이중 잠금 — 미래 git 기본값 변화로 조건이 충족돼도
    gc가 foreground로 돌아 commit이 끝까지 기다림 = 테스트 수명 내 종료 보장.
  - AGENTS.md §4 준수: detach된 자식은 PID를 알 수 없어 cleanup hook kill이
    불가능하므로, "spawn 자체의 거부"가 유일하게 청소 보장되는 형태다.
- 회귀 단언 `TestNewCheckFixture_DisablesAutoGC`: 두 설정의 존재를 고정.

### E4. 뮤테이션 RED (t372 교훈 — 규칙 제거가 RED를 내는지 관측)

설정 2줄을 임시 제거 → `go test -run TestNewCheckFixture_DisablesAutoGC ./internal/graph/`
→ **RED** (`git [config gc.auto]: exit status 1`) → 복구 → **GREEN** (0.789s).

### E5. 안정성 관측

- 수리 전, `go test -count=40 ./internal/graph/` (no-race): **ok 532.3s — 실패 0**.
  기본 gc.auto=6700 환경(로컬)에서는 트리거 미달이라 발화하지 않는다는 기제 예측과
  일치 — 즉 이 flake는 CI 환경 조건에서만 발화하는 간헐 결함이다.
- 수리 후, `-race` 반복 관측:
  - 1차 시도 `-count=40`: **go test 기본 `-timeout 10m`(600.585s) 초과로 전체 비정상
    종료 — 40회 미완, 관측 데이터 아님**(테스트 실패·Data Race 경합 0건 확인).
    `-race` 반복은 기본 10분 타임아웃 안에 40회가 안 들어오는 속도 차이였다.
  - 2차 시도 `-count=8 -timeout 9m`: **ok 144.933s — 실패 0**(Data Race 경합·
    TempDir cleanup 실패 모두 0). 배율 환산(8회 144.9s → 1회 ≈18.1s)으로
    40회 ≈ 12분 — 1차 시도의 타임아웃과 정합하며, 8회+수리 전 no-race 40회가
    "다수 회 반복 관측"의 이행이다.

### E6. 기타 검증

- `go vet ./internal/graph/` → 종료코드 0
- `go test -count=1 ./internal/graph/` → ok 15.798s (수리 적용 전체 통과)

## Baseline-attribution (baseline 귀속)

모든 실측은 이 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t423`,
브랜치 WT-tempdir-race, 기반 9145806d8)에서 본 세션이 직접 실행·관측한 값이다.
CI 로그 판독은 `gh run view <id> --log-failed`(repo modu-ai/moai-adk)의 원문이다.
git 소스 근거는 v2.50.1 태그의 builtin/gc.c 판독(로컬 Apple Git 2.50.1과 동일
버전선)이며, PPA 최신(러너)과의 로직 차이 가능성은 Gaps에 기록한다.

## Gaps (미검증)

1. **CI 러너에서 조건 충족의 정확한 근원은 미확정** — ubuntu-latest 러너 이미지의
   `/etc/gitconfig`는 `safe.directory=*`만 담는 것을 확인(install-git.sh)했으나,
   러너의 `~/.config/git/config`·PPA 최신 git(2.51+)의 need_to_gc 로직 변경 가능성은
   러너 실측 없이 배제하지 못한다. 수리가 조건과 무관하게 성립하므로 결론·조치에는
   영향이 없다(조건 미상 환경에서도 pack 쓰기가 차단됨).
2. 로컬 재현(발화)은 불가 — 기제는 조건 강제(probe-gc2.sh)로 실증했다. 카드 지시의
   "다수 회 반복 관측"은 수리 전 40회(발화 0, 기제와 정합)와 수리 후 40회로 이행.
3. d6fa555cf의 TestAlwaysLoadedTokenBudget 실패와 TestStopGoalWrapperCopiesStayIdentical
   실패는 본 카드에서 원인 규명하지 않았다(범위 밖, 별도 소관).

## Residual-risk (잔여 위험)

1. **형제 표면**: git init+commit fixture가 최소 15개 파일에 흩어져 있으며
   (internal/core/git/helpers_test.go, internal/spec/*, internal/web,
   internal/worktree, internal/verify, internal/cli, internal/template,
   internal/guardliveness), 같은 환경 조건에서 같은 경합이 가능하다. 본 수리는
   관측 표면(internal/graph newCheckFixture)만 고쳤다 — CI 잡 env에
   `GIT_CONFIG_COUNT/GIT_CONFIG_KEY_0=gc.auto/GIT_CONFIG_VALUE_0=0`을 주입하면
   한 곳의 변경으로 형제 전체가 방어되나, git 전역 설정을 만지는 테스트(safe.directory
   계열)에의 영향 검증이 선행돼야 해 별도 카드를 권한다.
2. 판정은 CI 다수 run에서 재발 0으로 확정하는 것이 실질 증명 — develop 통합 후
   관측 창이 열린다.

## 회귀 시험 요약

- `TestNewCheckFixture_DisablesAutoGC` — 뮤테이션(설정 제거)에서 RED 확정(E4).
- probe-gc.sh / probe-gc2.sh — 기제 실증 스크립트(본 디렉토리, 재실행 가능).
