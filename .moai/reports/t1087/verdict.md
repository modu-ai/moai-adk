# t1087 판정서 — test-browser CI job 첫 러너 실측 FAIL 원인 규명

- 카드: t1087 (Class B — 원인 규명 선행, plan 생략)
- 판정: **defect-absent (guard 논리) / defect-present (러너 환경 전제)** — 환경 실패로 분류, 수리 방향은 job 워크플로
- 브랜치: `WT-browser-runner` (base: origin/develop `d323f68fd`)
- 작성: 2026-09-23, 레인 세션 (worktree `.claude/worktrees/t1087`)

## Claim (주장)

1. test-browser job 의 `TestAppJsHandlersFireRuntime`·`TestAppJsHandlersFireSelectorMiss` FAIL 은 **가드(probe/driver) 논리의 결함이 아니라 러너 환경 실패**다 — 카드가 가른 두 갈래 중 「환경(CfT·pip) 실패 → job 워크플로 수정」 축이다.
2. 살해 기제의 1순위 후보: **Ubuntu 24.04 러너의 AppArmor unprivileged-userns 제한** — 러너에 설치된 system Chrome 은 AppArmor 프로필이 있어 기동되지만, job 이 sha256 핀해서 내려받은 Chrome-for-Testing 바이너리는 프로필이 없어 sandbox 가 요구하는 user namespace 생성이 거부되고, Chrome 은 CDP 포트를 노출하기 전에 즉사한다.
3. 부수 결함 (별도 수리 대상): 드라이버 `launchFireGuardChrome` 이 Chrome 의 stdout/stderr 와 종료 코드를 폐기하기 때문에, 러너가 왜 죽었는지를 CI 로그가 영원히 이름 대지 못한다 — 이번 판정이 문헌 교차에 의존하게 된 직접 원인.

## Evidence (증거)

### E1. CI 로그 — d323f68fd (run 35761021136, job 106858836238 "Test (browser fire guard)")

명령: `gh run view 35761021136 --job 106858836238 --log` (전문: `ci-log-d323f68fd.txt`)

선행 단계 전부 녹색:

```
2026-09-22T17:29:54.87Z websockets 15.0.1
2026-09-22T17:29:56.13Z chrome-linux64.zip: OK          ← sha256 핀 검증 통과
2026-09-22T17:29:59.44Z LINT OK: 8 entries + 7 exclusions cover 13 inventory groups
```

녹색 단계(17:30 스탬프 구간)에서 두 테스트 모두 동일 지점 즉사:

```
appjs_fire_guard_test.go:301: headless Chrome exited before exposing a CDP port
--- FAIL: TestAppJsHandlersFireRuntime (0.38s)
appjs_fire_guard_test.go:352: headless Chrome exited before exposing a CDP port
--- FAIL: TestAppJsHandlersFireSelectorMiss (0.32s)
FAIL	github.com/modu-ai/moai-adk/internal/web	0.702s
```

해석: `cmd.Start()` 는 성공했다 (exec 실패였으면 `start headless Chrome: %v` 로 빠졌을 것). Chrome 프로세스는 **탄생 후 0.4초 안에 스스로 죽었고**, `DevToolsActivePort` 를 쓰지 못했다 (appjs_fire_guard_test.go:244 의 `case <-done:` 분기). 다운로드·무결성·pip·manifest 는 전부 정상 — 실패 위치는 Chrome 기동 한 곳뿐이다.

### E2. CI 로그 — b0d9e0bbc (run 35735866560, job 106772819557)

명령: `gh run view 35735866560 --job 106772819557 --log` (전문: `ci-log-b0d9e0bbc.txt`)

```
chrome-linux64.zip: OK
appjs_fire_guard_test.go:301: headless Chrome exited before exposing a CDP port
--- FAIL: TestAppJsHandlersFireRuntime (0.27s)
--- FAIL: TestAppJsHandlersFireSelectorMiss (0.21s)
```

서명이 E1 과 바이트 수준으로 동일 (시각만 13:48). 두 헤드 × 두 러너 인스턴스에서 재현 — 일시적 장애가 아니라 결정적 환경 상태다.

### E3. 로컬 게이트 재현 — 동일 트리, 정상 환경

명령 (worktree t1087 @ `d323f68fd`, darwin/arm64 + system Chrome, 전문: `local-gated-test.log`):

```
MOAI_BROWSER_GUARD=1 go test ./internal/web/ -run 'AppJsHandlersFire' -v -count=1 -timeout 10m
```

```
--- PASS: TestAppJsHandlersFireRuntime (18.51s)
--- PASS: TestAppJsHandlersFireSelectorMiss (15.97s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	35.012s
exit=0
```

동일 커밋, 동일 probe/driver/manifest 가 완전한 순환(서버 기동 → Chrome 기동 → probe CDP → selector-miss 조작 검증)을 통과. **녹색과 적색의 차이는 코드가 아니라 환경 한 변수**다.

### E4. 러너 이미지 구성 (actions/runner-images Ubuntu2404-Readme, 2026-09-23 열람)

ubuntu-24.04 러너 이미지에는 `Google Chrome 152.0.7977.82` 와 `Chromium 152.0.7977.0` 이 설치돼 있다. system Chrome 이 설치 돌아간다는 것은 브라우저 공유 라이브러리 닫힘집합이 러너에 존재한다는 뜻이며, CfT 153 의 라이브러리 요구는 Chrome 152 와 실질적으로 같다. 「러너에 라이브러리 부재」를 살해 기제로 보기 어렵다.

### E5. 문헌 교차 — CfT on Ubuntu 24.04 의 확립된 실패 기제

- Chromium 이슈 트래커 "Chrome for Testing cannot run on Ubuntu 24.04": Ubuntu 가 `/opt/google/chrome/chrome` (Chrome stable) 에 AppArmor 프로필을 적용하지만, 내려받는 CfT 바이너리는 어떤 프로필에도 속하지 않아 user namespace 생성이 AppArmor 에 의해 거부된다.
- Puppeteer 공식 troubleshooting (puppeteer#12818): `/etc/apparmor.d/chrome` 가 Puppeteer 가 내려받은 CfT 바이너리의 userns 사용을 막는다고 명시, 우회로 `--no-sandbox` 또는 CfT 경로용 AppArmor 프로필 설치(`apparmor_parser -r`)를 안내.
- Ubuntu 24.04 는 커널 AppArmor 정책(`apparmor_restrict_unprivileged_userns`)으로 프로필 없는 바이너리의 unprivileged userns 생성을 기본 차단 (Launchpad #2046844).
- 실제 CI 수리 선례: vitest.dev CI 가 job 안에서 CfT 경로 AppArmor 프로필을 쓰고 `apparmor_parser` 로 적재하는 커밋을 착지.

이 기제는 관측된 모든 사실과 맞물린다: system Chrome(프로필 있음)은 러너에서 잘 돌고, 핀된 CfT(프로필 없음)는 sandbox 초기화 단계에서 즉사하며, 사망 시각은 0.2~0.4초, `DevToolsActivePort` 미기록.

### E6. 코드 판독 — 진단 불능의 원인 (appjs_fire_guard_test.go)

`launchFireGuardChrome` (appjs_fire_guard_test.go:210-250) 은 Chrome `exec.Command` 에 Stdout/Stderr 를 연결하지 않고, `go func() { _ = cmd.Wait(); close(done) }()` 로 **종료 코드마저 폐기**한다. 그래서 :244 의 `t.Fatal("headless Chrome exited before exposing a CDP port")` 은 "죽었다"는 사실만 말하고 "왜"를 영구히 잃는다. 이번 판정서의 Gaps 절이 존재하는 직접 원인.

## Baseline-attribution (측정 baseline 귀속)

| 측정 | 대상 | 시점 |
|---|---|---|
| E1 | GH 호스티드 러너 ubuntu-latest, HEAD `d323f68fd`, run 35761021136 | 2026-09-22 17:29-17:30 UTC |
| E2 | 동일 러너 이미지, HEAD `b0d9e0bbc`, run 35735866560 | 2026-09-22 13:47-13:48 UTC |
| E3 | worktree t1087 @ `d323f68fd` (origin/develop 과 동일 트리), darwin/arm64, 로컬 system Chrome | 2026-09-23 (이번 런) |
| E4/E5 | actions/runner-images main 브랜치 Readme + 문헌 (2026-09-23 열람) | 문헌 근거 — 기계 측정 아님 |

## Gaps (미검증)

1. **러너에서의 Chrome stderr 원문 미관측.** E6 대로 드라이버가 폐기하기 때문에 어떤 러너 로그에도 죽음의 이유가 없다. AppArmor-userns 를 1순위 기제로 판정한 것은 관측(즉사 형태·system-CfT 격차)과 문헌(E5)의 합성이지, 러너 커널 메시지의 직접 관측이 아니다. 확정하려면 드라이버에 stderr 수집을 넣거나(코드 수리 → Kickoff 게이트) job 에 진단 스텝을 추가해 재측정해야 한다.
2. **docker 재현은 무효.** Apple Silicon 의 Docker Desktop 이 amd64 chrome 을 Rosetta 로 번역하며 `rosetta error: failed to open elf at /lib64/ld-linux-x86-64.so.2` 로 즉사 — 라이브러리·sandbox 어느 축도 측정하지 못했다 (zip 무결성·퍼미션만 유효). 컨테이너 재현 대신 문헌 교차(E5)로 대체했다.
3. **red phase (0→1→0 변이 순환) 는 CI 에서 끝내 미도달.** job 이 녹색 단계에서 죽어 §E.2 M4 의 세 번째 미검증 항목(변이 순환)은 이번에도 미검증 상태로 남는다. 로컬에서는 red 방향의 절반(selector-miss 조작 → exit 1 판정)이 E3 로 통과했지만, CI 의 실바이너리 변이 순환 전체는 별도다.

## Residual-risk (잔여 위험)

- GitHub 러너 이미지가 AppArmor 제한을 이미지 차원에서 해제하거나 구성을 바꾸면(이미지는 자주 갱신된다) 이 실패가 저절로 녹색으로 바뀔 수 있다 — 그때 이 판정서의 기제 절은 「과거 환경」이 된다. 수리가 지연돼도 무해하나, 재발 시 같은 서명이면 이 문서를 먼저 대조할 것.
- 실측으로 1순위 기제가 틀리고 라이브러리 결손이 확인되면(가능성 낮음 — E4) 수리 내용은 apt 의존 설치로 바뀐다. 수리의 **축(job 워크플로)은 어느 쪽이든 동일**하다.
- `--no-sandbox` 우회를 택할 경우 러너에서의 sandbox 격리가 사라진다 — repo-root 서빙 가드(REQ-AFG-012)는 manifest 가 reversible 효과만 다루도록 이미 제한하고 있으나, 프로필 설치(E5 선례)가 격리를 보존하는 정석 수리다.

## 수리 방향 제안 (리드 상신 — 이 레인은 착수하지 않음)

1. **job 워크플로 수리 (환경 실패 축 — 카드 판정에 따른 본수리).** Chrome 기동 전에 CfT 경로용 AppArmor 프로필을 설치하고 `apparmor_parser -r` 로 적재 (E5 의 확립된 CI 패턴). `--no-sandbox` 우회보다 sandbox 격리를 보존.
2. **드라이버 진단력 수리 (별도, Kickoff 게이트 필요).** `launchFireGuardChrome` 이 Chrome 의 stderr 와 종료 코드를 수집해 실패 시 `t.Fatal` 이 실제 사유를 이름 대게 수리 — 이번 적색이 2일간 「원인 불명」으로 남은 것의 직접 원인이며, 재발 시간을 몇 시간으로 줄인다.
3. red phase 도달 검증은 1·2 착지 후 다음 develop push 의 CI 에서 자동으로 이뤄진다 (첫 녹색 러너 실측).

## 원문 근거 파일 (이 디렉터리)

- `ci-log-d323f68fd.txt` — run 35761021136 job 106858836238 전문 (349행)
- `ci-log-b0d9e0bbc.txt` — run 35735866560 job 106772819557 전문
- `local-gated-test.log` — E3 로컬 재현 전문
