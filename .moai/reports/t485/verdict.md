# t485 verdict — index.lock 경합 생성자 확정

- 카드: t485 · 브랜치: `WT-indexlock-origin` · SPEC: `SPEC-INDEXLOCK-CREATOR-001`
- 측정 트리: develop tip `25a3212a9` + 본 카드 커밋 (worktree `.claude/worktrees/t485`)
- 관측 창: 2026-09-04 18:18–18:42 KST (관측기 4회 실행, 누적 약 24분)
- 방법: git 호출 0개 수동 폴러 (0.12–0.25초 간격 stat glob) + 라이브 홀더 시
  `lsof` 즉시 귀속 + 동시(`eps`)·스로틀(`ps`) 스냅샷. 배경 부하 0,
  관측기 자가종료 + 외부 `timeout` 이중 경계.

## Claim

**C1 — 경합 계열(메시지에 경로 있음) `index.lock` 을 실제로 쥐는 프로세스는
`moai statusline` 이 spawn 하는 `git status --porcelain` 이다.** 확정. 구성 증거:

1. **홀더 직접 귀속 1건**: 18:32:28 KST, `t480` 트리 락 홀더 PID 9076 =
   `/Applications/Xcode.app/.../git status --porcelain`, 부모 9031 =
   `moai statusline` (거리 0초 스냅샷. `summary.jsonl` 해당 행).
2. **락 순간의 동시 생존 6건**: 락 존재 순간의 스냅샷에서 산 `git status
   --porcelain` — 전원의 부모가 `moai statusline`. 3단 체인 완성 사례:
   `claude --name lane-7 (93615)` → `moai statusline (67817)` → `git status
   --porcelain (67854)` (18:35:03, `eps-1788514503-t482.txt`).
3. **생산자 코드 경로 (본 트리에서 재검증)**:
   `internal/statusline/git.go:37` `c.repo.Status()` — 렌더마다 호출, 실패는
   조용히 삼킴(41–45행); `internal/core/git/manager.go:103`
   `execGit(ctx, m.root, "status", "--porcelain", "--branch")` —
   `--no-optional-locks` 면제 없음. `grep -rn "OPTIONAL_LOCKS|no-optional-locks"
   internal/ pkg/ cmd/` = **0히트**. 일반 `git status` 는 인덱스 갱신 시
   기회적 쓰기로 `index.lock` 을 획득한다.
4. **후보 배제**: 셸 git 래퍼 없음(`type git` = `/usr/bin/git`, 프로필 청결),
   moai 소스 `git -c` 사용 0, 레인 env `GIT_CONFIG_*` 0(내 세션 + lane-3 생존
   프로세스), 리드 스윕·타 레인 조작은 관측된 귀속과 모순 없는 배경일 뿐
   귀속 사례 0.

**C2 — 락 수명은 수십 ms 다.** 실홈 95건 중 94건은 `lsof` 가 PID를 돌려줬는데도
직후 `ps` 3회 버스트(50ms 간격)에서 이미 사라졌고, 532건은 stat→lsof 사이
(수십 ms) 소멸. → (a) 재시도 1회 우회가 11사례 전부 통과한 이유, (b) 사후 `ps`
캡처가 11번 모두 "git 프로세스 0건"을 본 이유(캡처 지연 ≫ 락 수명),
(c) lane-4 의 "직후 ls 시 락 이미 부재" 관측이 정확히 이 프로필을 가리켰던
이유,가 하나로 설명된다.

**C3 — 경합 강도는 세션 활동에 비례한다.** 관측 행 1,427건 중 t485(본 세션,
가장 활발) 868 · primary 120 · t480 65 · t483 50 … 렌더 빈도가 곧 락 빈도다.

**C4 (부수 확인) — 원인 A(fsmonitor `/dev/null`)의 주체는 config 파일 축에서
못 잡는다. 축 판정: config 파일 추적은 사축(dead axis).** 근거: 이 리포 전체
(Go 소스·템플릿·훅·status_line.sh·러처)에 `fsmonitor` 문자열 0히트 — moai 는
주입자가 아니다. 전역/로컬/config.worktree/`--show-origin` 라이브 모두 청결.
`.env.glm` 내 `GIT_*` 0. 셸 래퍼 없음. 레인 env `GIT_CONFIG_*` 0. 남는 유일한
흔적은 `.git/config` 의 무귀속 쓰기 1건(KST 2026-09-03 23:04:30, 리드도
"set/unset 주체 미확인"으로 기록). 판정: (a) 런타임 주입(GIT_CONFIG_*/`git -c`)
은 위 점검으로 **부정**됐고, (b) 남은 가설은 "일시적 config 쓰기→제거" 하나 —
그 주체는 상태에서 복원 불가다. 원인 A 는 t466 창에서 값 소실로 이미 해소
판정을 받았으므로, 재추적이 필요해지면 축은 **프로세스 레벨**(`.git/config`
fs-event 감시)이나 운영자 질의로 바꿔야 하고 config 파일 그리기는 그만둔다.

**C5 — 죽은 잠금(잔재)은 별개 메커니즘이다.** 8/28 lock-sweep 의 죽은 잠금 3건
(primary 09:06:56 · t215 09:07:55 · glm-settings-persist 09:07:57)은 mtime 이
1분 창에 인접 = 동시 죽음 파동(kill-wave) 서명. 본 카드의 경합(라이브,
sub-second)과 다른 클래스다. lane-13 t471 의 0바이트 잔재도 이 계열에 속한다.

## Evidence

| # | 주장 | 명령/산출물 | 관측 출력 |
|---|---|---|---|
| E1 | 홀더 직접 귀속 | `summary.jsonl` (summarize.py) | `{"epoch":1788513948,"tree":"t480","holder_pid":9076,"holder_cmd":"...git status --porcelain","holder_ppid":"9031","ppid_cmd":"moai statusline"}` |
| E2 | 3단 체인 | `eps-1788514503-t482.txt` | `67854 67817 … git status --porcelain` + `67817 93615 … moai statusline` + 93615=`claude … --name lane-7`(`ps-1788513588.txt`) |
| E3 | 코드 경로 | Read `internal/core/git/manager.go:94-106` | `func (m *gitManager) Status()` … `execGit(ctx, m.root, "status", "--porcelain", "--branch")` |
| E4 | statusline 호출점 | Read `internal/statusline/git.go:26-45` | `CollectGitStatus` 가 렌더마다 `repo.Status()` 호출, 오류 삼킴 |
| E5 | 면제 플래그 부재 | `grep -rn "OPTIONAL_LOCKS\|no-optional-locks" internal/ pkg/ cmd/` (본 트리) | 0행 |
| E6 | 락 수명 프로필 | `summary.md` | REAL 95 · GONE 532 · EMPTY 2 (GONE = "No such file or directory" lsof 기록) |
| E7 | 트리 분포 | `events.jsonl` 집계 | t485 868 · primary 120 · t480 65 · t483 50 · t482 48 · t481 44 · t476 43 · t473 43 · t472 41 · t484 39 · t477 34 · t479 17 · develop 12 · t478 3 |
| E8 | 후보 배제 | `type git` · 프로필 grep · moai 소스 grep · `ps eww` 키 점검 | 전부 부정(본문 C1.4) |
| E9 | fsmonitor 축 | 리포 전체 `fsmonitor` grep + config 5면 + `.env.glm` + 레인 env | 주입 경로 0히트(본문 C4) |

원본 전체: `.moai/reports/t485/raw/` (events.jsonl · summary.{md,jsonl} ·
lsof-*.txt 629 · eps-*.txt · ps-*.txt · holder-*.txt · watch.log ·
watch-indexlock.sh · summarize.py).

## Baseline-attribution

- 모든 코드 인용은 **본 카드 트리**(develop `25a3212a9` + 카드 커밋,
  `.claude/worktrees/t485`)에서 재측정했다. 최초 `manager.go:88` 인용은
  뒤처진 primary 작업트리에서 읽은 것으로 **기각**하며, 정정값은 `:103`이다.
- 캡처 수치 전부는 2026-09-04 18:18–18:42 KST, 이 머신, 이 관측기의 실측이다.
  리드 장부의 수치(113샘플/42% 등)는 인용일 뿐 본 측정이 아니다.
- 락 수명 "수십 ms"는 측정값이 아니라 관측 상한(upper bound) 추정이다 —
  stat(0.12초 격자)과 lsof 소요 사이에 존재했다는 것까지만 관측됐다.

## Gaps

- 실홈 95건 중 94건은 홀더 PID 의 부모를 못 찾았다(초단명). 직접 귀속 1건 +
  동시 생존 6건이 전부다. 나머지 94건에 **다른 생성자가 섞였을 가능성은
  배제하지 못한다** — 다만 관측된 귀속은 전원 statusline 이고, 반대 방향
  증거(비-statusline 홀더의 귀속 사례)는 0건이다.
- 역방향 경합(레인의 git 쓰기가 락을 쥐는 동안 statusline 이 조용히 실패)은
  fatal 로그를 남기지 않아 사례로 집계되지 않는다. git.go:41-45 가 그 침묵의
  코드 근거다.
- primary 120건은 개별 귀속하지 않았다. 리드 세션 자신의 statusline 이
  최유력 후보지만 이번 창에서 잡지 못했다.
- fsmonitor 주체: 상태에서 복원 불가 판정(C4). 주체 특정은 이 카드에서
  실패로 남는다.
- 홀더의 cwd 가 어느 트리를 겨냥했는지는 ps 명령줄로 알 수 없어, 동시 생존
  6건을 "같은 트리의 형제"라고까지는 주장하지 않는다(머신 전역 스냅샷이다).

## Residual-risk

- **원인 B(경로 없는 `Unable to write index.`)는 미해결 그대로다** — 본 카드
  범위 밖(배차 명시). lane-14 캡처(실패 순간 git 0건)와 본 카드의 sub-second
  락 프로필을 겹치면, 원인 B 도 "포착 지연 ≫ 사건 수명" 계열일 가능성이
  있으나 이건 가설이다 — 측정 없이 원인 B 에 적용하지 말 것.
- 경합 빈도는 활성 레인 수×활동 강도에 비례한다. 레인이 줄면 사건 자체가
  줄어들어 재시도 우회가 계속 통할 것이다 — 그것이 원인 소멸으로 읽히지
  않게 주의할 것.
- 수리(`GIT_OPTIONAL_LOCKS=0` 부착 또는 statusline 캐시)는 **별도 판단
  사안**이다 — 배차 원칙(확정 후 별도)대로 본 카드는 착수하지 않았다.
- 관측기 자체는 검증됐으나(실홈 95건) 0.12초 격자라 그보다 짧은 창의 사건은
  아예 관측되지 않았을 수 있다. 수치는 하한으로 읽을 것.
