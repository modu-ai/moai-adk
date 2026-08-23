# t173 §A 측정 기록 — moai update 청소 경로의 심볼릭 링크 인식

> 카드 t173 plan-phase §A ground-truth dossier. 모든 주장은 (a) 코드 인용 `file:line`
> 또는 (b) 재현 관찰(명령어 + verbatim 출력) 중 하나로 뒷받침된다. 추론만으로
> 뒷받침된 문장은 명시적으로 표기했다.

## 측정 조건

| 항목 | 값 |
|---|---|
| 코드 앵커 | commit `4b2f203fe` (본 워크트리 HEAD; main 은 측정 중 `a1b1ca696`로 진행됨 — 모든 line 인용은 `4b2f203fe` 기준) |
| 워크트리 브랜치 | `worktree-agent-ac385069535544f99` (디스패치 표기 `WT-clean-links`와 이름이 다르나 HEAD 커밋 `4b2f203fe`는 일치 — 판독 시 커밋 앵커 기준) |
| 재현 바이너리 | `/tmp/t173-moai` (`go build -o /tmp/t173-moai ./cmd/moai`, ldflags 미주입 → `v3.1.2 none built unknown`; 설치 안 함) |
| 재현 픽스처 | `/tmp/t173-fixture-31Bw4H` (`moai init` 직후 499 파일; 실제 프로젝트 아님) |
| 환경 | darwin/arm64 |
| 실행 예산 | build 1회 + `init` 1회 + `update` 4회 (Run A~D) — 초과 없음 |
| 날짜 | 2026-08-22 |

모든 update 실행 형태: `(cd <fixture> && /tmp/t173-moai update --yes --force </dev/null 2>&1 | tail -N)`.
`--yes`는 비TTY 필수(`internal/cli/update_template_sync.go:676-679` — 비TTY에서 확인 UI 거부),
`--force`는 버전 일치 short-circuit 우회(`update_template_sync.go:600` — init에 쓴 바이너리와
같은 버전이면 no-op). `projectRoot := "."`(`update_template_sync.go:588`).

---

## §1 분기 전환 추적

### 1.0 판정 코드 (현행 main = `os.Stat`)

`internal/cli/update/deploy/deploy.go` `backupThenRemove` (:371-399):

```
371  func backupThenRemove(diskPath, relTarget, backupBase string, tmplFS fs.FS) (int, error) {
372      info, err := os.Stat(diskPath)            ← 판정 지점 (Stat = 링크 추적)
374      if os.IsNotExist(err) { return 0, nil }   ← 대상 부재 = 성공 no-op (링크 자체는 미제거)
380      if !info.IsDir() {                        ← 파일 분기
381          if templateCarries(tmplFS, relTarget) { return 0, os.RemoveAll(diskPath) }
384          copyRegularFile(diskPath, ...)        ← os.ReadFile (deploy.go:465, 링크 추적)
387          return 1, os.RemoveAll(diskPath)
390      managed := templateManagedPaths(...)      ← 디렉터리 분기
394      backupUnmanagedTree(diskPath, ...)        ← WalkDir (deploy.go:437, 루트 Lstat)
398      return backedUp, os.RemoveAll(diskPath)
```

호출 계층: `CleanMoaiManagedPaths` (deploy.go:101)의 비-글롭 루프는 청소 직전에도
`os.Stat`(:139) → `IsNotExist`면 `Skipped (not found)`(:140-146)로 `backupThenRemove` 자체를
건너뛴다. 글롭 루트는 `filepath.Glob`(:116) 매치 결과별로 `backupThenRemove(match)`(:128)를
부르고, glob 매치에 대한 별도 skip 메시지는 없다(매치가 없으면 그냥 지나감).

### 1.1 `os.Stat`(현행) — 3형태 추적

**(a) 살아 있는 디렉터리 링크가 관리 뿌리에 있는 경우** (예: `.claude/rules/moai` → 외부 디렉터리)

1. `os.Stat`(:372)이 링크를 **추적** → `info.IsDir() == true` → 디렉터리 분기(:390).
2. `templateManagedPaths`(:390)는 템플릿 FS만 걷으므로 무사 통과.
3. `backupUnmanagedTree`(:394)의 `filepath.WalkDir`(:437)은 **루트를 Lstat**한다 — 루트
   엔트리가 링크면 `d.IsDir()==false && !d.Type().IsRegular()` → :441에서 즉시 스킵.
   **백업 0건. 링크 대상 디렉터리 내부의 비관리 파일은 백업되지 않는다.**
4. `os.RemoveAll`(:398)은 링크를 추적하지 않는다 → **링크만 제거, 대상 디렉터리 무사**.
5. 이후 Deploy가 같은 경로에 **실제 디렉터리를 재생성**한다(§2.3).
- 종결: **중단 없음. 링크는 조용히 소멸하고 실제 템플릿 디렉터리로 대체.**
- 관찰 근거: Run B(`.claude/rules/moai`), Run D(`.claude/skills/moai-livelink`) — §3.

**(b) 살아 있는 파일 링크가 파일 루트에 있는 경우** (예: `.claude/settings.json` → 외부 파일)

1. `os.Stat`(:372) 추적 → `IsDir()==false` → 파일 분기(:380).
2. `templateCarries(tmplFS, ".claude/settings.json")`(:381) = **false** — 템플릿은
   `.claude/settings.json.tmpl`만 보유(실측: `ls internal/template/templates/.claude/ |
   grep settings` → `settings.json.tmpl` 23,819B 유일)이고 렌더링 결과 경로는
   `fs.Stat` 대상이 아니다(deployer.go:133,150 참조).
3. `copyRegularFile`(:384) → `os.ReadFile`(:465)이 **링크를 추적해 대상의 바이트를 백업**
   → `RemoveAll`(:387) 링크만 제거.
- 종결: **중단 없음. 백업은 대상의 내용물, 링크는 소멸, 대상 파일 무사.**
- 관찰 근거: Run B — 백업본과 최종 배포본 모두 `OUTSIDE-SETTINGS-v1`(§3.2).

**(c) dangling 링크가 관리 뿌리에 있는 경우** (예: `.claude/agents/moai` → 부재 경로)

1. 청소 루프의 `os.Stat`(:139)이 링크를 추적 → 대상 부재 → **ENOENT** →
   `Skipped .claude/agents/moai (not found)`(:140-146). **링크 자체는 존재하는데도
   제거되지 않고 남는다.** (`backupThenRemove`의 :372-375 no-op도 같은 의미.)
2. 글롭 루트 아래 dangling이면: Glob은 이름 매칭이라 매치되지만(§2.4)
   `backupThenRemove` :372의 Stat이 ENOENT → `(0, nil)` **무소식 no-op** — 링크 잔존.
3. Deploy 단계: 템플릿의 `.claude/agents/moai/*` 파일 기록 시
   `os.MkdirAll(".claude/agents/moai", 0o755)`(`internal/template/deployer.go:189`)이
   dangling 링크에서 **EEXIST** (`mkdir ...: file exists`) → Deploy 실패 → update 중단.
- 종결: **중단(EEXIST). 부분 파괴 상태로 종료. 재실행해도 clean이 같은 Skip을 반복하므로
  같은 EEXIST가 영구히 재현 — 사용자가 링크를 수동 제거하기 전까지 update 불능.**
- 관찰 근거: Run D — §3.4.
- 보완: 업데이트 흐름상 이 에러는 반환되고 `cmd/moai/main.go:20-23`에서 `os.Exit(1)`.
  (Run D의 셸 exit 코드는 파이프(`| tail`)가 삼켜 직접 포착 못 함 — §5 Gaps.)

### 1.2 `os.Lstat` 치환 — **가설(실존하지 않는 코드)**

> 라벨: HYPOTHETICAL. 다음 4증거로 "t81의 Lstat 치환은 어느 트리에도 없다"를 확정했다.
> 1. main `4b2f203fe` deploy.go:372 = `os.Stat` (본 워크트리 직독).
> 2. `git diff main..WT-skills-canonical -- internal/cli/update/deploy/deploy.go` →
>    **빈 출력** (WT-skills-canonical = t81 워크트리 HEAD `6d0097abf`, ref 실재 확인).
> 3. t81 작업 트리(미커밋 포함) 직독: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t81/internal/cli/update/deploy/deploy.go:372` = `os.Stat`.
> 4. `git log --all --oneline -S "os.Lstat" -- internal/cli/update/deploy/deploy.go` → **0건**
>    (어느 브랜치의 어느 커밋도 이 파일에 `os.Lstat`를 넣은 적 없음).
>
> 즉 Lstat 형태는 t81 iter4 감사(D1)가 지적하고 기각시킨 **제안**으로만 존재한다.
> 아래 추적은 감사 보고서의 darwin/arm64 실측 인용과 본 코드 라인을 결합한 재구성이다.

치환 시 (`:372`만 `os.Lstat`로 바꾸고 나머지 그대로):

| 형태 | Lstat 결과 | 착지 분기 | 종결 |
|---|---|---|---|
| (a) 살아 있는 디렉터리 링크 | err=nil, `IsDir()==false`(모드 = symlink) | **파일 분기(:380)** — `templateCarries` false(디렉터리 루트는 템플릿이 파일로 보유하지 않음) → `copyRegularFile`(:384) → `os.ReadFile`(:465) 디렉터리 읽기 → **EISDIR** → `"back up %s"` 랩(:385) → clean 중단(:157-161) → **update 실패** | 하드 실패 |
| (b) 살아 있는 파일 링크 | err=nil, `IsDir()==false` | 파일 분기 — 현행과 동일하게 `ReadFile`이 링크 추적 → 백업 성공 → `RemoveAll` | 성공(현행 1.1(b)와 동일) |
| (c) dangling 링크 | **err=nil**(Lstat은 대상을 보지 않음), `IsDir()==false` | 파일 분기 → `copyRegularFile` → `ReadFile` ENOENT → **중단**. (주: :374의 IsNotExist no-op은 Lstat이 nil을 반환하므로 더 이상 발동하지 않는다 — 현행의 "조용한 잔존"이 "하드 실패"로 바뀜) | 하드 실패 |

감사 실측 인용(`.moai/reports/t81/plan-audit-iter4.md:117-123`, t81 워크트리):

```
live  Stat err=<nil> IsDir=true | Lstat err=<nil> IsDir=false
ReadFile(dir-symlink) err=read .../.agents/skills/moai-live: is a directory
gone  Stat IsNotExist=true | Lstat err=<nil>
```

### 1.3 형태 × 판정 종합표

| 대상 경로의 형태 | 현행 `os.Stat` | 가설 `os.Lstat` | 실측 |
|---|---|---|---|
| 살아 있는 디렉터리 링크 (관리 뿌리) | 성공 — 디렉터리 분기, 백업 0건, 링크만 제거, 실제 디렉터리로 재배포 | **EISDIR 하드 실패(clean 중단)** | Run B / Run D(livelink) |
| 살아 있는 파일 링크 (파일 루트) | 성공 — 대상 바이트 백업, 링크 제거, 대상 무사 | 성공(동일) | Run B |
| dangling 링크 (비-글롭 뿌리) | clean은 "Skipped (not found)"로 링크 잔존 → **deploy에서 EEXIST 중단(영구 루프)** | **ENOENT 하드 실패(clean 중단)** | Run D |
| dangling 링크 (글롭 매치 이름) | clean 무소식 no-op → 링크 **영구 잔존**(중단은 아님) | **ENOENT 하드 실패(clean 중단)** | Run D |
| 실제 디렉터리 (대조) | 성공 — 전량 정상 청소/재배포 | 성공(변화 없음) | Run A |
| hns-* 사용자 소유 경로 (링크 포함) | 성공 — 청소 뿌리 밖, 일절 미접촉 | (동일 — 청소 대상 아님) | Run C |

---

## §2 청소 뿌리 인벤토리

### 2.1 `ManagedCleanTargets` 7뿌리 (deploy.go:50-82; 상수 `internal/defs/dirs.go:409-414`, `files.go:6`)

| # | 루트 (DisplayPath) | 인용 | 글롭 | 형태 |
|---|---|---|---|---|
| 1 | `.claude/settings.json` | deploy.go:52-55 (`defs.SettingsJSON` files.go:6) | 아니오 | 파일 |
| 2 | `.claude/commands/moai` | deploy.go:56-59 (`CommandsMoaiSubdir` dirs.go:411) | 아니오 | 디렉터리 |
| 3 | `.claude/agents/moai` | deploy.go:60-63 (`AgentsMoaiSubdir` dirs.go:409) | 아니오 | 디렉터리 |
| 4 | `.claude/skills/moai*` | deploy.go:64-68 (`SkillsSubdir` dirs.go:410) | **예** (IsGlob) | 글롭 |
| 5 | `.claude/rules/moai` | deploy.go:69-72 (`RulesMoaiSubdir` dirs.go:412) | 아니오 | 디렉터리 |
| 6 | `.claude/output-styles/moai` | deploy.go:73-76 (`OutputStylesSubdir` dirs.go:413) | 아니오 | 디렉터리 |
| 7 | `.claude/hooks/moai` | deploy.go:77-80 (`HooksMoaiSubdir` dirs.go:414) | 아니오 | 디렉터리 |

**별도 8번째 제거**: `.moai/config` (deploy.go:168-182, :176 `backupThenRemove(configDir,...)`).
`ManagedCleanTargets` 목록 밖(감사 §A.4와 동일 구조)이지만 **같은 `backupThenRemove`를 공유**하므로
링크 의미론도 동일하게 적용된다(§1.1의 (a)/(c) 추적이 그대로 성립). `.moai/config`가 dangling 링크면
clean은 no-op(:372-375)이고 deploy가 `.moai/config/*` 기록 시 `MkdirAll` EEXIST로 중단 — Run D의
`.claude/agents/moai` 사례와 같은 형태(코드 추적; 실측은 §5 Gaps).

### 2.2 각 뿌리가 "디렉터리로의 심볼릭 링크"일 때 현행 동작

- 비-글롭 루트: §1.1(a) — Stat 추적 → 디렉터리 분기 → WalkDir이 링크 루트를 스킵(:441) →
  **백업 0건** → `os.RemoveAll` **링크만 제거**. 대상 디렉터리(와 그 내부 비관리 파일)는
  무사하지만 **백업도 받지 않는다** — 링크가 사라진 뒤 사용자는 원래 링크였다는 사실 자체를
  알 수 없다(진행 메시지는 평소와 동일한 `Removed <path>`).
- 파일 루트(1번): §1.1(b) — 대상 바이트가 백업되고 링크 소멸.
- 글롭 루트(4번): §2.4.

### 2.3 재배포는 실제 디렉터리를 만드는가 — **예 (Run B 실측)**

Deploy 단계(`update_template_sync.go:323` → `deployer.Deploy`)는 forceUpdate 배포기로
동작한다(`update_template_sync.go:130` — `NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)`).
파일별 흐름(deployer.go:121-211): 디렉터리 엔트리 스킵(:121-124) → `.tmpl` 렌더링(:133,150) →
`os.MkdirAll(destDir)`(:189) → `atomicWriteFile`(:201, tmp+rename :19-32). 존재 검사(:169-185)는
forceUpdate에서 건너뛴다. 청소가 링크를 지운 뒤면 `MkdirAll`이 **실제 디렉터리를 새로 만들고**
템플릿 파일이 실제 파일로 기록된다. Run B 사후 상태(§3.2): `.claude/rules/moai` =
`drwxr-xr-x` 실제 디렉터리 + `NOTICE.md`, `core/` … — 링크가 실제 디렉터리로 완전 치환됨.

### 2.4 글롭(`.claude/skills/moai*`) × 링크 (Run D 실측)

`filepath.Glob`(deploy.go:116)은 메타 문자 패턴을 디렉터리 엔트리 이름에 매칭한다(타입 무관).
같은 글롭 팔에서 형태별 결과가 갈렸다:

| 글롭 매치 항목 | 청소 결과 | 사후 상태(Run D 실측) |
|---|---|---|
| 실제 디렉터리(템플릿 스킬 전부 + bare `moai`) | 제거 → (deploy 중단 전까지) 재배포 예정이었으나 Run D에서는 deploy 중단으로 **미복구** | 소멸(§3.4 blast radius) |
| `moai-livelink` → 외부 실제 디렉터리 (생존 링크) | 매치 → Stat 추적 IsDir → 디렉터리 분기 → 백업 0건 → **링크만 제거** | 링크 소멸, 외부 대상 무사(`UNMANAGED-EXTRA-v1` 확인) |
| `moai-dangling-custom` → 부재 경로 | 매치(이름 매칭) → `backupThenRemove` :372 Stat ENOENT → `(0,nil)` **무소식 no-op** | **링크 영구 잔존** — 제거도 재배포도 안 됨 |

livelink(제거됨)와 dangling-custom(잔존)은 Glob의 이름 매칭 조건이 동일하므로, 결과 차이는
Glob이 아니라 `backupThenRemove`의 Stat 추적에서 비롯한다 — 이것이 "Glob은 dangling도
매치한다"는 간접 실측이다(감사 §A.10의 직접 실측 `Glob 이 dangling 을 매치 — 전부 재현`과 일치,
plan-audit-iter4.md:226).

### 2.5 사용자 소유 네임스페이스 (청소 뿌리와의 관계)

`IsUserOwnedNamespace`(`internal/cli/update/plan/plan.go:152-208+`)의 패턴족:
`.claude/skills/hns-*`(:158), `.claude/skills/harness-*`(:164), `.claude/skills/my-harness-*`(:169),
`.claude/agents/harness{,/}`(:174), `.moai/harness{,/}`(:179), `.claude/commands/harness{,/}`(:185),
`.claude/workflows/hns-*`(:191), `.claude/workflows/harness-*`(:197), 그리고
`.claude/skills/`의 **비-moai 첫 세그먼트 전부**(:202-207 — `seg != "moai" && !HasPrefix(seg, "moai-")`).
이 네임스페이스는 `ManagedCleanTargets` 7뿌리 어디에도 걸리지 않는다(글롭 `moai*`만 보면
`hns-*`는 이름부터 불일치). Run C 실측: `.claude/skills/hns-mine/SKILL.md`(실측 내용
`HNS-USER-OWNED-v1`)와 그 내부의 dangling 링크 `badlink` 모두 update 후 무결 — 백업 디렉터리에서
`hns-mine` 경로 검색 0건(`find .moai-backups -path "*hns-mine*" | wc -l` → `0`).

---

## §3 재현 매트릭스 (실측)

사전 단계(재현 환경 조성, 전부 /tmp):

```
$ mktemp -d /tmp/t173-fixture-XXXXXX          → /tmp/t173-fixture-31Bw4H
$ /tmp/t173-moai init /tmp/t173-fixture-31Bw4H → "✓ MoAI project initialized" (499 files)
```

### 3.1 Run A — 대조군 (실제 디렉터리 = "copied real dir" 형태)

```
$ (cd /tmp/t173-fixture-31Bw4H && /tmp/t173-moai update --yes --force)
  ✓ Removed .claude/rules/moai
  ✓ Removed .claude/output-styles/moai
  ✓ Removed .claude/hooks/moai (backed up 31 unmanaged file(s))
  ✓ Removed .moai/config (backed up 7 unmanaged file(s))
  ○ Deploying templates...
  ✓ Templates deployed
  ✓ Updated 495 files
  32 merged/added + 463 managed re-deployed · removed 385 under managed paths (all re-deployed)
  Backup: .moai-backups/20260822_042633
RUN-A: 성공 (전 단계 ✓)
```

판정: 링크가 없으면 전 과정 정상 — 청소→재배포→병합 완결. 이하 Run들의 차이는 링크만으로 귀속된다.

### 3.2 Run B — 살아 있는 디렉터리 링크 + 살아 있는 파일 링크

셋업(실측 명령):

```
$ rm -r /tmp/t173-fixture-31Bw4H/.claude/rules/moai
$ ln -s /tmp/t173-outside-rules      /tmp/t173-fixture-31Bw4H/.claude/rules/moai      # 디렉터리 링크
$ rm /tmp/t173-fixture-31Bw4H/.claude/settings.json
$ ln -s /tmp/t173-outside-settings.json /tmp/t173-fixture-31Bw4H/.claude/settings.json # 파일 링크
# 외부 대상: /tmp/t173-outside-rules/unmanaged-extra.md = "UNMANAGED-EXTRA-v1"
#           /tmp/t173-outside-settings.json             = "OUTSIDE-SETTINGS-v1"
```

실행:

```
$ (cd /tmp/t173-fixture-31Bw4H && /tmp/t173-moai update --yes --force)
  ○ Removing .claude/rules/moai...
  ✓ Removed .claude/rules/moai                    ← 백업 개수 문구 없음 = 0건 (§1.1(a) 3단계 실측)
  ...
  ✓ Templates deployed
  ✓ Merged 1 file(s) with 3-way merge engine
  ✓ Updated 495 files
RUN-B: 성공 — 중단 없음
```

사후 상태(실측):

```
$ ls -ld .claude/rules/moai .claude/settings.json
  drwxr-xr-x ... .claude/rules/moai        ← 실제 디렉터리로 재생성 (NOTICE.md, core/ … 포함)
  -rw-r--r-- ... .claude/settings.json     ← 실제 파일 (20 bytes)
$ cat .moai-backups/20260822_042702/pre-clean/.claude/settings.json
  OUTSIDE-SETTINGS-v1                       ← 파일 링크: 대상 바이트가 백업됨
$ cat .claude/settings.json
  OUTSIDE-SETTINGS-v1                       ← 최종본도 대상 내용 (3-way merge가 "사용자 내용"으로 복원)
$ find .moai-backups/20260822_042702/pre-clean -maxdepth 3
  ... .claude/settings.json  .claude/hooks/moai  .claude/commands/moai  .moai/config ...
  (※ .claude/rules/moai 없음 = 디렉터리 링크 뿌리의 백업 0건)
$ cat /tmp/t173-outside-rules/unmanaged-extra.md → UNMANAGED-EXTRA-v1   ← 외부 대상 무사
$ cat /tmp/t173-outside-settings.json        → OUTSIDE-SETTINGS-v1      ← 외부 대상 무사
```

판정: 살아 있는 링크 2형태 모두 **중단 없음**. 디렉터리 링크 = 백업 0건 + 링크 소멸 + 실제
디렉터리로 치환. 파일 링크 = 대상 바이트 백업/복원 + 링크 소멸. 두 경우 모두 링크였다는
사실은 진행 출력에 나타나지 않는다(링크 무인식).

### 3.3 Run C — hns 사용자 소유 경로

셋업: `.claude/skills/hns-mine/SKILL.md`(내용 `HNS-USER-OWNED-v1`) + 내부 dangling 링크
`.claude/skills/hns-mine/badlink -> /tmp/t173-gone`.

```
$ (cd /tmp/t173-fixture-31Bw4H && /tmp/t173-moai update --yes --force)
  ✓ Updated 495 files · removed 410 under managed paths (all re-deployed)
RUN-C: 성공
$ cat .claude/skills/hns-mine/SKILL.md → HNS-USER-OWNED-v1  (무결)
$ ls -la .claude/skills/hns-mine/      → SKILL.md + badlink(잔존)
$ find .moai-backups -path "*hns-mine*" | wc -l → 0
```

판정: 사용자 소유 네임스페이스(내부의 dangling 링크 포함)는 update 전체에서 일절 미접촉.

### 3.4 Run D — dangling 링크 3종 (핵심 재현)

셋업(실측 명령):

```
$ rm -r /tmp/t173-fixture-31Bw4H/.claude/agents/moai
$ ln -s /tmp/t173-gone        /tmp/t173-fixture-31Bw4H/.claude/agents/moai              # 비-글롭 뿌리, dangling
$ ln -s /tmp/t173-also-gone   /tmp/t173-fixture-31Bw4H/.claude/skills/moai-dangling-custom  # 글롭 매치, 템플릿 비보유 이름
$ ln -s /tmp/t173-outside-rules /tmp/t173-fixture-31Bw4H/.claude/skills/moai-livelink       # 글롭 매치, 생존 디렉터리 링크
# test -e /tmp/t173-gone → 부재 확인 (exit 1)
```

실행 (verbatim, 장식 코드 제외):

```
$ (cd /tmp/t173-fixture-31Bw4H && /tmp/t173-moai update --yes --force)
  ○ Removing .claude/agents/moai...
  ✓ Skipped .claude/agents/moai (not found)       ← Stat이 dangling을 추적해 ENOENT; 링크 미제거
  ○ Removing .claude/skills/moai*...
  ✓ Removed .claude/skills/moai*                  ← livelink 제거 + dangling-custom은 무소식 no-op
  ...
  ○ Deploying templates...
  ✗ Deployment failed: template deploy mkdir ".claude/agents/moai": mkdir .claude/agents/moai: file exists

moai update failed after the first destructive step.

  failed step:  Deploy Templates
  error:        deploy templates: template deploy mkdir ".claude/agents/moai": mkdir .claude/agents/moai: file exists
  backup dir:   .moai-backups/20260822_042800
  restore with: moai update --restore ".moai-backups/20260822_042800"

No automatic rollback was attempted. The project tree is in a
partially updated state; inspect the backup directory above, then run the
restore command to reapply it.

✗ Error: Deploy Templates: deploy templates: template deploy mkdir ".claude/agents/moai": mkdir .claude/agents/moai: file exists
RUN-D: 실패 (에러 반환 — main.go:20-23 경로로 exit 1; 셸 exit 코드 직접 포착은 §5 Gaps)
```

사후 상태(실측):

```
$ ls -la .claude/agents/        → harness(사용자 소유, 무사) + moai -> /tmp/t173-gone (dangling 링크 그대로)
$ ls    .claude/skills/         → hns-mine + moai-dangling-custom(잔존) 뿐 — 템플릿 스킬 전멸
$ find .claude/skills -type f | wc -l → 1        (hns-mine/SKILL.md 만)
$ find .claude/commands -type f | wc -l → 0
$ find .claude/output-styles -type f | wc -l → 0
$ find .claude/rules -type f | wc -l → 0
$ cat /tmp/t173-outside-rules/unmanaged-extra.md → UNMANAGED-EXTRA-v1  (livelink 대상 무사)
```

판정 (3형태):

1. **dangling 링크 at 비-글롭 뿌리 (`.claude/agents/moai`)**: clean이 링크를 "없는 경로"로
   오인해 남기고(`:139-146`), deploy의 `MkdirAll`(deployer.go:189)이 EEXIST로 죽는다 —
   **현행 제품 코드의 실측 결함.** clean이 이미 다른 뿌리 전부를 지운 뒤라 트리는
   부분 파괴 상태(스킬 38디렉터리·commands·output-styles·rules 전량 소멸), 복구는
   `moai update --restore` + 후속 update. 단 **재실행해도 clean의 Skip이 동일하게
   재현되어 같은 EEXIST로 죽는다** — 링크를 수동 `rm`하기 전까지 update는 영구 불능.
   (재실행 루프 주장의 근거: clean Skip은 Run D에서 실측, 재실행 자체는 실행하지 않음 —
   예산 4회 소진. Skip→EEXIST가 상태 비의존적이라는 점은 코드 경로 :139-146 + deployer.go:189에서
   직접 follow.)
2. **dangling 링크 at 글롭 매치 이름 (`moai-dangling-custom`)**: 중단 없음, 대신 **영구 잔존**
   — 어느 update에도 제거되지 않고(템플릿 비보유 이름이라 재배포도 안 함) 아무 메시지도 없음.
3. **생존 디렉터리 링크 at 글롭 (`moai-livelink`)**: §1.1(a)와 동일 — 백업 0건, 링크만 제거,
   대상 무사.

### 3.5 매트릭스 종합

| 형태 | update 종결 | 링크 최종 상태 | 대상(링크 지목 경로) | 백업 | 중단 지점 |
|---|---|---|---|---|---|
| 실제 디렉터리 (Run A) | 성공 | — (재배포) | — | 비관리 파일 정상 백업 | 없음 |
| 생존 디렉터리 링크 (Run B/D) | 성공 | 소멸 | 무사 | **0건** | 없음 |
| 생존 파일 링크 (Run B) | 성공 | 소멸 | 무사 | 대상 바이트 | 없음 |
| dangling, 비-글롭 뿌리 (Run D) | **실패** | **잔존** | (부재) | — | deploy `MkdirAll` EEXIST |
| dangling, 글롭 매치 (Run D) | 성공 | **영구 잔존** | (부재) | — | 없음(무소식) |
| hns-* 사용자 소유 (Run C) | 성공 | 무변화(내부 링크 포함) | 무사 | 미수집(뿌리 밖) | 없음 |

---

## §4 t81 D2~D4 원문 추출

출처: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t81/.moai/reports/t81/plan-audit-iter4.md`
(t81 워크트리, 읽기 전용 접근). 감사 판정은 **FAIL — blocking 4건(D1~D4)**(:258). 인용은
원문 그대로.

### D2 — probe-branch 실측 (plan-audit-iter4.md:137-158)

> ### D2 — 새 탐침 `moai-linkprobe` 가 실제 미러와 다른 코드 분기를 탄다 (AP-14 형태의 재현)
>
> `acceptance.md:97,110-113` (AC-CSC-008 fixture 5번 항목)
> **Severity: major — Class: blocking**
>
> fixture 표는 `moai-linkprobe` 를 `.claude/skills/hns-linkprobe-src/**SKILL.md**` 를 가리키는 링크로 규정한다 — **파일 링크**다. 반면 실제 미러(REQ-CSC-003)는 `.agents/skills/<name>` → `../../.claude/skills/<name>` 의 **디렉터리 링크**다. 측정하면 두 형태는 `backupThenRemove` 에서 갈린다:
>
> ```
> linkprobe(파일 링크)  Stat.IsDir=false  Lstat.IsDir=false   → 항상 파일 분기
> moai-live(디렉터리 링크) Stat.IsDir=true  Lstat.IsDir=false  → 분기가 뒤바뀜
> ReadFile(file-symlink) data="keepme" err=<nil>              → 성공
> ReadFile(dir-symlink)  err=is a directory                   → 실패
> WalkDir(dir-symlink) entries=1 (type=L---------)            → 정규 파일 0개
> ```
>
> 즉 `moai-linkprobe` 는 D1 의 실패 모드에서 **살아남는 유일한 fixture 항목**이다. 5·6번 단언은 통과하는데 1·2번은 실패한다. 그리고 AC-CSC-008 자신이 `[HARD]` 로 세운 AP-14("fixture 를 실 디렉터리로만 심는다 — 실제 산출물은 링크이므로 결함이 통과한다", plan.md:137)는 **fixture 형태가 제품 형태와 달라선 안 된다**는 규율인데, 그 규율을 닫으려고 추가한 항목이 같은 규율을 어겼다.
>
> acceptance.md:113 의 [HARD] 절은 *이름*이 의도적으로 비현실적이라는 사유만 적고, 형태(파일 vs 디렉터리) 차이는 언급하지 않는다. 읽는 사람은 이 차이를 알 수 없다.
>
> 부수적으로: 파일 링크에 대해 `copyRegularFile` 은 `os.ReadFile` 로 **대상의 바이트를 백업 트리에 복사한다**(실측 `data="keepme"`). §A.7 이 "복사는 비정규 항목을 건너뛴다"고 적은 것은 `WalkDir` 경로(디렉터리 링크)에 한해 참이고, 파일 링크 경로에서는 백업이 실제로 링크를 따라간다.
>
> **Required fix**: 탐침을 **디렉터리 링크**로 바꿀 것 — `.agents/skills/moai-linkprobe` → `../../.claude/skills/hns-linkprobe-src`(디렉터리), 6번 단언은 `.claude/skills/hns-linkprobe-src/SKILL.md` 가 그대로 읽힘. 이러면 제품 형태와 같은 분기를 타면서 청소 글롭 밖 대상이라는 성질은 유지된다. 그리고 형태를 제품과 일치시켰다는 사실을 [HARD] 절에 명시할 것.

### D3 — fixture 수 표기 드리프트 (plan-audit-iter4.md:160-171)

> ### D3 — plan M4 닫힘 조건의 fixture 개수가 개정을 따라오지 못했다
>
> `plan.md:101` — `닫힘 조건: AC-CSC-007(양팔 — 글롭 + 순서), AC-CSC-008(4형태 — dangling 팔 포함), AC-CSC-009`
> **Severity: major — Class: blocking**
>
> acceptance 는 **5형태 / 6단언**(매트릭스 acceptance.md:18 "미러 5형태 단일 테스트", 본문 acceptance.md:91 "아래 다섯 항목을 **모두** 심어 두고", acceptance.md:105 "다음 여섯 단언"). plan 만 iter-4 판본의 `4형태` 를 유지한다.
>
> 이것은 표기 오류에 그치지 않는다 — M4 의 **닫힘 조건**이므로, plan 만 읽는 구현자는 fixture 를 4개만 심고 `moai-linkprobe` 팔을 통째로 빠뜨린 채 마일스톤을 닫았다고 판단한다. iter-5 HISTORY(spec.md:21)는 "D6 개수 표기 3건"을 반영했다고 적었는데, 반영되지 않은 한 곳이 하필 닫힘 조건이다.
>
> 같은 계열의 잔여 드리프트: `plan.md:19` R5 완화 항목이 `AC-CSC-008(양팔)` — iter-2 시절 어휘로, 지금은 6단언이다.
>
> **Required fix**: plan.md:101 을 `AC-CSC-008(5형태 6단언 — dangling 팔 + 링크 추종 탐침 포함)` 로 정정. plan.md:19 의 `(양팔)` 도 함께 갱신.

### D4 — AC-CSC-012 공허 단언 (plan-audit-iter4.md:173-184)

> ### D4 — AC-CSC-012 의 2번 단언이 자기 Given 이 만드는 fixture 에서 공허하게 참이다
>
> `acceptance.md:163,167-168`
> **Severity: major — Class: blocking**
>
> 2번 단언은 자기 존재 이유를 본문에 적었다 — *"링크 모드뿐 아니라 **복사 모드에서도** 0이어야 한다 — 이 단언의 존재 이유가 복사 모드다"*(acceptance.md:168). 그런데 Given 절(acceptance.md:163)이 구성하는 fixture 는 **AC-CSC-002 의 배포**, 즉 링크 모드 1회뿐이다. 복사 모드를 발동시키는 절이 없다.
>
> 링크 모드에서 이 단언이 어떻게 되는지 측정했다: 디렉터리 링크에 대해 `backupThenRemove` 는 `os.Stat` → `IsDir()==true` → `templateManagedPaths` → `backupUnmanagedTree` 로 가고, `filepath.WalkDir` 은 링크 루트를 `Lstat` 하므로 항목 1개(type `L---------`)만 걷고 정규 파일을 **0개** 만난다. 즉 백업 파일 수는 구현이 무엇을 하든 0 이고, 2번은 **항상 참**이다.
>
> AC-CSC-005 는 "심볼릭 링크 생성이 실패하도록 주입된 배포기"라는 seam 이 존재함을 이미 규정한다. AC-CSC-012 의 Given 은 그 seam 을 쓰지 않는다.
>
> **Required fix**: AC-CSC-012 의 Given 에 **두 배포**를 명시할 것 — (링크 모드 1회) + (AC-CSC-005 의 주입으로 복사 모드 1회), 2번 단언을 두 결과 모두에 대해 요구. 그렇지 않으면 R8("복사 모드에서 매번 전량 백업", plan.md:22, 등급 **높음**)이 판정에서 빠진다.

### t173 AC 설계로의 이관 (요약 — 인용 아님)

- **D2 → t173 AC**: fixture의 *형태*(파일 링크 vs 디렉터리 링크 vs dangling)가 실제로 타는
  코드 분기를 결정한다 — t173의 AC fixture는 판정하려는 형태와 동일 형태여야 하고, 그
  일치를 [HARD] 절에 명시해야 한다(본 dossier §1.3 표가 형태×분기 매핑의 근거 데이터).
- **D3 → t173 AC**: AC의 형태 개수/단언 개수는 spec·plan·acceptance 3면에서 동일해야 하고,
  마일스톤 닫힘 조건이 개정을 따라가야 한다.
- **D4 → t173 AC**: 단언이 자기 Given이 구성한 fixture에서 구현과 무관하게 항상 참이면
  공허하다 — 특히 "디렉터리 링크의 백업 파일 수 == 0"은 현행 구현에서 WalkDir-스킵으로
  항상 0이므로(§1.1(a) 3단계, Run B 실측), 이 숫자를 재료로 한 단언은 다른 축(링크 제거
  여부, 대상 무사, 메시지)과 결합해야 반증 가능해진다.

---

## §5 Gaps (미측정) 및 잔여 위험

### 미측정

1. **moai 프로세스의 셸 exit 코드 직접 포착 실패** — Run A~D 모두 `| tail` 파이프로 묶어
   `$?`가 tail의 종료 코드를 반환했다. Run D의 실패는 verbatim 에러 출력으로 증명되며,
   exit 1은 `cmd/moai/main.go:20-23`(반환된 error → `os.Exit(1)`; ExitCoder 아님)의 코드
   경로 인용으로 대체한다. 실측 숫자가 필요하면 파이프 없이 1회 재현해야 한다(예산 소진으로 미실시).
2. **dangling 링크의 재실행 루프** — "재실행해도 같은 EEXIST"는 코드 추적(clean Skip :139-146
   는 Run D 실측 + deploy `MkdirAll` :189의 상태 비의존성)이며, 두 번째 update 실행으로
   직접 확인하지는 않았다(4회 예산 소진).
3. **파일 루트(`.claude/settings.json`)에 dangling 링크가 있는 경우** — 코드 추적: clean
   `:139` Stat ENOENT → "Skipped (not found)"로 링크 잔존 → deploy는 `atomicWriteFile`의
   rename(deployer.go:28)이 목적지 링크를 대체하므로 성공 추정. **미실측**(Run D의
   settings.json은 Run B 이후 실제 파일이었음). 3-way merge 백업 단계에서의 상호작용도 미측.
4. **`.moai/config`가 dangling 링크인 경우** — §2.1의 코드 추적만 있음(같은 `backupThenRemove`
   경로). 미실측.
5. **linux/windows 미측정** — 전 재현이 darwin/arm64. EISDIR/ENOENT/EEXIST의 errno 값 자체는
   POSIX 범용이나 Windows의 심볼릭 링크 생성/권한 차이는 별도 변수.
6. **Go 레벨 직접 검증(테스트 함수) 부재** — 트리에 테스트 파일 추가 금지 지시로 인해 모든
   측정이 바이너리 E2E(`/tmp/t173-moai update`) 형태. `backupThenRemove` 단위 경계의 정밀
   측정(예: backupUnmanagedTree의 스킵 라인 :441 직접 관찰)은 감사 D2/D4의 실측 인용으로 보강했다.
7. **Lstat 형태는 실행 불가** — 존재하지 않는 코드(§1.2 4증거). 해당 열 전체가
   감사 실측 인용 + 코드 재구성이며, t173의 SPEC이 Lstat을 도입하는 경우 run-phase에서
   신규 실측이 필요하다.

### 잔여 위험

- 픽스처가 `moai init` 직후 상태(사용자 커스터마이즈 없음)라 merge/restore 경로가 단순했다.
  사용자가 설정을 수정한 프로젝트에서 파일 링크(settings.json)가 3-way merge와 만나는
  정확한 내용 흐름은 Run B보다 복잡할 수 있다.
- 본 측정의 `moai` 바이너리는 ldflags 미주입(`none/unknown`)이라 버전 비교 경로를 `--force`로
  우회했다. 정식 빌드에서 버전 일치 short-circuit이 먼저 발생하면 사용자는 dangling-링크
  결함을 다른 버전 업데이트 시점에 처음 마주하게 된다(결함 자체는 동일).
- `/tmp` 픽스처와 `/tmp/t173-moai` 바이너리는 재현 물증으로 잔존 중(OS가 정리).
  `~/go/bin/moai`에는 설치하지 않았다.

---

## 인용 색인 (핵심 file:line — 전부 앵커 `4b2f203fe`, 본 워크트리)

- `internal/cli/update/deploy/deploy.go` — ManagedCleanTargets :50-82; CleanMoaiManagedPaths
  :101; 글롭 팔 :115-137 (Glob :116, backupThenRemove 호출 :128); 루트 Stat :139, Skip
  :140-146; config 제거 :168-182 (:176); backupThenRemove :371-399 (os.Stat :372, no-op
  :374-375, 파일 분기 :380-388, 디렉터리 분기 :390-398); templateCarries :426-429;
  backupUnmanagedTree :435-459 (WalkDir :437, 비정규 스킵 :441); copyRegularFile :464-473
  (os.ReadFile :465)
- `internal/template/deployer.go` — atomicWriteFile :19-32 (rename :28); 디렉터리 스킵
  :121-124; .tmpl :133,150; forceUpdate 존재검사 우회 :169-185; MkdirAll :189-191; 기록 :201
- `internal/cli/update_template_sync.go` — forceUpdate 배포기 :130; Inventory :281;
  guard+Clean :289-298 (:297); Deploy :323; mergeable :357-368; projectRoot :588; --force
  :590; 버전 short-circuit :600; 비TTY --yes :676-679
- `internal/cli/update_disk_backup.go` — guardFirstDestructiveStep :126-131
- `internal/cli/update/plan/plan.go` — IsUserOwnedNamespace :152-208+ (hns :158, harness :164,
  my-harness :169, agents/harness :174, .moai/harness :179, commands/harness :185,
  workflows :191/:197, 비-moai 스킬 :202-207)
- `cmd/moai/main.go` — error → os.Exit(1) :20-23
- `internal/defs/dirs.go` :6 (MoAIDir), :9 (ClaudeDir), :12 (BackupsDir), :372 (ConfigSubdir),
  :409-414 (Claude 하위 디렉터리 상수); `internal/defs/files.go` :6 (SettingsJSON)
- t81 감사: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t81/.moai/reports/t81/plan-audit-iter4.md`
  — D1 :95-135 (실측 :117-123), D2 :137-158, D3 :160-171, D4 :173-184, §A.10 :226, 판정 :258-261
