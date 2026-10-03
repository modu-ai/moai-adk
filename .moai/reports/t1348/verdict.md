# t1348 판정서 — 장기 레인 세션에 Claude Code CLI 업데이트를 반영하는 방법

- 카드: t1348 (Low)
- 측정 트리: `worktree-agent-a4cdd22ee542798ae` @ `02f939b14` (develop 팁에서 분기)
- 측정일: 2026-10-03
- 성격: 조사 + 절차 제안. 코드 변경 없음.

## 1. Claim (주장)

1. **실행 중인 Claude Code 프로세스는 새 바이너리를 집어 들지 못한다(핫스왑 없음).** 업데이트는 다음 프로세스 시작 때 적용된다. 공식 문서가 그렇게 말하고, 이 머신의 실측도 같다.
2. **따라서 "재시작 없이 반영"은 불가능하고, 길은 "카드 경계에서 새 프로세스로 갈아타기"뿐이다.** 다만 레인의 카드 경계는 원래 `/clear`로 문맥을 버리는 지점이므로, 그 자리에서 프로세스를 바꾸는 비용은 `/clear`와 거의 같다. "재시작 = 레인의 죽음"은 레인을 운영자가 직접 띄운 단일 세션으로 돌릴 때만 성립한다.
3. **MoAI에는 이미 그 장치가 있다.** `moai cc|glm -f lane --clear-policy relaunch`는 카드마다 자식 `claude` 프로세스를 새로 띄우는 상위 루프다(`internal/cli/factory_lane_relaunch.go`). 자식은 매번 `~/.local/bin/claude` 심볼릭 링크를 새로 따라가므로, 이 정책의 레인은 다음 카드부터 자동으로 최신 바이너리를 쓴다. 새로 만들 것은 없고, 문서화·기본값·가시화가 부족할 뿐이다.
4. 카드 중간에 꼭 갈아타야 할 때는 `claude --resume <id>`가 대화를 복원하지만, MoAI 런처가 주입하는 환경(레인 스탬프, `--settings`)과 bypass 권한 모드는 복원되지 않는다. 그래서 카드 중간 재개는 권장 경로가 아니라 비상 경로다.

## 2. Evidence (증거)

### 2.1 공식 문서 (이번 실행에서 실제로 가져온 URL)

- https://code.claude.com/docs/en/setup § Auto-updates — "Claude Code checks for updates on startup and periodically while running. Updates download and install in the background, then **take effect the next time you start Claude Code**."
  - 같은 문서 § Install on network storage — "A running session reads parts of the Claude Code executable from disk as it works" / 네이티브 설치기는 "any version a session on the same machine is running"을 지우지 않고 남긴다 → 실행 중 세션은 자기 버전 파일에 묶여 있다.
  - § Disable auto-updates — `DISABLE_AUTOUPDATER=1`은 백그라운드 확인만 끈다(`claude update`는 여전히 동작). 모든 경로를 막으려면 `DISABLE_UPDATES`.
  - § Configure release channel — `autoUpdatesChannel: "stable"`(약 1주 지연, 회귀 릴리스 건너뜀), `minimumVersion` 바닥값.
- https://code.claude.com/docs/en/cli-reference — `claude update`, `--resume/-r`(ID·이름·transcript 경로), `--continue/-c`, `--fork-session`, `--name/-n`, `--session-id`. 그리고 `claude respawn <id>` 행: "Restart a background session … Use `--all` to restart every running session, e.g. **to pick up an updated Claude Code binary**" — 공식적으로도 새 바이너리 반영 수단은 "재시작"이다.
- https://code.claude.com/docs/en/agent-view § The supervisor process — 백그라운드(`--bg`) 세션만 해당: "After an auto-update: the supervisor restarts itself onto the new version and moves idle sessions over in the background. Sessions that are working, waiting on you, or attached aren't interrupted." MoAI 레인은 터미널 전경 세션이라 이 경로를 타지 않는다.
- https://code.claude.com/docs/en/sessions § What a resumed session restores — 복원: 대화 이력, 모델, 에이전트, 권한 모드(단 `bypassPermissions`는 복원 안 됨), 활성 goal, 만료 안 된 예약 작업. 복원 안 됨: 백그라운드 Bash·monitor, 그리고 "`--mcp-config`, `--settings`, `--plugin-dir`, `--fallback-model`, `--add-dir` … pass them again when you resume". 1시간 이상 비활성 + 10만 토큰 초과 세션은 재개 시 요약/전체 선택 대화상자가 뜨고 캐시가 식어 있다.
- https://code.claude.com/docs/en/changelog — 2.1.282~2.1.288 범위에서 실행 중 바이너리 교체를 다룬 항목은 없었다(가져온 요약 기준).

### 2.2 이 머신의 실측

```
$ claude --version
2.1.288 (Claude Code)

$ ls -l ~/.local/bin/claude
~/.local/bin/claude -> ~/.local/share/claude/versions/2.1.288

$ lsof -a -d txt -p <pgrep -x claude 의 31개 PID> | grep -oE '(versions/|claude-code/)[0-9.]+' | sort | uniq -c
   1 claude-code/2.1.284      (Desktop 앱 내장본)
   1 versions/2.1.281
   5 versions/2.1.283
   1 versions/2.1.285
   4 versions/2.1.286
  15 versions/2.1.287
   4 versions/2.1.288

$ lsof -p 38102 | grep txt ; ps -o etime=,lstart= -p 38102
… /Users/goos/.local/share/claude/versions/2.1.281
08-19:21:57 Thu Sep 24 16:57:00 2026
```

→ 링크는 2.1.288을 가리키는데 실행 중 31개 프로세스 중 4개만 2.1.288이다. 9월 24일부터 8일째 도는 프로세스는 여전히 2.1.281에 매핑돼 있다. 핫스왑이 없다는 직접 증거다.

### 2.3 MoAI 런처 코드

- `internal/cli/factory_lane_relaunch.go` — `runFactoryLaneRelaunch`: 리스 → 카드 워크트리 확보 → `exec.Command(binaryPath, claudeArgs...)`로 대화형 세션 1개 실행·대기(`c.Run()`) → 반복. `binaryPath`는 `exec.LookPath("claude")` 결과(`internal/cli/mcp_claude_runner.go:20`)인 `~/.local/bin/claude`이고, 링크는 매 `exec` 때 새로 해석된다.
- `internal/config/envkeys.go:381-413` — `MOAI_FACTORY_CLEAR_POLICY`: `clear-each`(기본) / `clear-when-full` / `relaunch`.
- `grep -c 'clear-policy' internal/cli/cc.go internal/cli/glm.go` → `0` / `0`: `moai cc`·`moai glm` 도움말에 `--clear-policy`가 없다(파서는 `internal/cli/factory.go:191`에서 받는다).
- `moai session list --json` 필드: `cwd, host, last_heartbeat, phase, pid, session_id, spec_id, started_at` — 실행 바이너리 버전 필드는 없다.

## 3. Baseline-attribution (기준 귀속)

- 코드 인용은 모두 이 워크트리 HEAD `02f939b14`에서 읽은 것이다.
- 프로세스·버전 실측은 2026-10-03 이 머신(darwin)에서 위 명령으로 이번 실행 중에 관측했다.
- 문서 인용은 이번 실행에서 WebFetch로 가져온 페이지 본문이다(위 URL 5개 외 다른 URL은 인용하지 않았다).

## 4. Gaps (미검증)

- `relaunch` 정책의 자식이 실제로 새 버전으로 뜨는 장면은 직접 재현하지 않았다. 근거는 코드 읽기(매 반복 `exec`)와 심볼릭 링크 해석 규칙이다. 확인하려면 relaunch 레인에서 카드 두 장 사이에 업데이트가 끼었을 때 `lsof -p <자식 pid>`로 버전을 보면 된다.
- `moai cc -f lane-<n> -- --resume <session-id>` 형태로 런처를 거쳐 재개할 수 있는지(런처가 주입하는 `--name`·`--settings`와 충돌하는지) 실행해 보지 않았다.
- 31개 프로세스 중 어느 것이 레인이고 어느 것이 리더·서브 세션인지는 구분하지 않았다(버전 분포만 쟀다).
- 최초 관측(배너 ×12, 16h10m 유휴)은 카드 본문 인용이며 이번에 재측정하지 않았다.
- changelog는 최근 구간(2.1.282~288) 요약만 확인했다.

## 5. Residual-risk (잔여 위험)

- 설치기는 "같은 머신에서 실행 중인 버전"과 최신 2개를 남기고 지운다. 판별이 틀리면 오래 도는 레인이 바이너리를 잃고 죽을 수 있다(문서는 네트워크 저장소 시나리오에서 이 위험을 명시). 로컬 디스크에서는 이번 실측상 2.1.281이 남아 있어 현재는 보호되고 있다.
- 업데이트가 쌓인 채 레인이 며칠씩 구버전으로 돌면, 리더·레인 간 동작 차이(새 플래그, 훅 입력 필드, 메시징 채널)가 생길 수 있다. 실측상 지금 7개 버전이 공존한다.
- `DISABLE_AUTOUPDATER`로 배치 중 버전을 고정하는 안은 배너 소음을 없애지만, 보안 수정 반영을 사람 손에 맡기게 된다.

---

## 제안 절차 — 배치/카드 경계 재시작

### A. 권장: relaunch 정책으로 레인을 띄운다 (코드 변경 없음)

```
moai cc -f lane --clear-policy relaunch
```

- 레인 1개 = 상위 루프 1개. 카드마다 새 `claude` 자식이 카드 워크트리에서 뜬다.
- 카드가 끝나면(완료 보고 송신, 통합 창 release, 슬롯 반납 뒤) 운영자는 `/clear` 대신 `/exit`한다. 루프가 다음 카드를 리스하고 **새 바이너리로** 다음 세션을 띄운다.
- 업데이트 배너가 떠도 아무것도 하지 않는다. 다음 카드 경계가 곧 재시작이다.
- 카드 워크트리 규율("새 카드는 새 워크트리", "실행 중 세션의 워크트리 이동 금지")과도 원래 맞물린다 — 세션을 옮기는 게 아니라 끝내고 새로 띄우기 때문이다.

### B. clear-each / clear-when-full 레인을 쓰는 경우

배너를 본 레인은 다음 카드 경계에서 `/clear` 대신 이렇게 한다.

1. 완료 보고 송신 → 통합 창 release → 슬롯 반납까지 끝낸다(카드 하나의 수명이 닫힌 상태).
2. `/exit`로 세션을 끝낸다. 버리는 것은 `/clear`가 어차피 버릴 문맥뿐이다.
3. 같은 터미널에서 `moai cc -f lane-<n>`(또는 `moai glm …`)으로 같은 레인 이름으로 다시 합류한다. 새 카드는 새 워크트리에서 시작하므로 `--resume`이 필요 없다.
4. 리더는 이 재합류를 레인 소멸로 보지 않는다(레인 라벨이 같다). 큐가 위임 채널이므로 메시지 유실 비용도 없다.

### C. 비상: 카드 중간에 꼭 갈아타야 할 때

1. 진행 상황을 `progress.md`와 handoff 블록(`moai handoff save`)에 남긴다.
2. `/exit` 후 카드 워크트리에서 재개한다. 우선 `moai cc -w <카드 워크트리 절대경로>`로 들어가 handoff 블록을 붙여 넣는 방식을 쓴다 — 런처 주입 환경이 온전하다.
3. 대화 이력이 꼭 필요할 때만 `claude --resume <session-id>`를 쓴다. 이때 `--settings`/`--mcp-config`/`--plugin-dir` 재지정, bypass 권한 모드 재지정이 필요하고, 백그라운드 Bash·monitor는 돌아오지 않으며, 레인 스탬프(`MOAI_KANBAN_*`)는 런처를 거치지 않으면 비어 있다.

### D. 배치 경계에서 한꺼번에

- 배치 시작 전 `claude update` 한 번 → 모든 레인을 새로 띄운다.
- 배치 중 버전을 고정하고 싶으면 레인 env에 `DISABLE_AUTOUPDATER=1`(배너·백그라운드 다운로드 억제)을 두고, 배치 경계에서만 `claude update`. 선택 사항이며 트레이드오프는 §5 마지막 항목.

## MoAI가 자동화할 수 있는 것 (후속 카드 후보)

1. **`--clear-policy` 도움말·문서화 + 레인 권장값 지정** (Low, 문서/도움말 1~2파일): `moai cc`/`moai glm` 도움말에 현재 0건. 업데이트 반영 경로가 이미 있는데 보이지 않는 게 이번 카드의 근본 원인에 가깝다.
2. **실행 바이너리 노후 가시화** (Medium): `moai session list` 또는 `moai doctor`에 "이 세션의 실행 버전 vs `~/.local/bin/claude` 대상 버전"을 추가(macOS `lsof -a -d txt -p`, Linux `/proc/<pid>/exe`). 리더가 어느 레인이 몇 버전 뒤인지 보고 카드 경계 재시작을 지시할 수 있다.
3. **relaunch 루프의 버전 변경 한 줄** (Low): 카드 사이에 링크 대상이 바뀌었으면 `claude 2.1.287 → 2.1.288` 한 줄을 찍는다. 반영 여부를 증거로 남긴다.
4. **런처 경유 재개 검증** (Low): `moai cc -f lane-<n> -- --resume <id>`가 동작하는지 재현하고, 안 되면 `--resume` 통과 경로를 설계한다(§C의 비상 경로를 안전하게 만드는 일).
