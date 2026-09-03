# t459 — LSEL 수집기 캡 무장 전이: 재현 결과와 가드

카드: t459 (Class B — plan 생략) · 브랜치 `WT-inbox-cap-marker` · 트리 `.claude/worktrees/t459`
기준 트리: develop `5107bbfff` 흡수 후 (t280 `1c6ee0503` 포함 — `git merge-base --is-ancestor` 로 확인)

---

## Claim

1. **위험은 재현된다.** curator 머신에서 `.moai/state/lsel/` 가 사라지면 수집기 캡이 무장하고, 다음 append 하나로 live inbox 전체가 `.1` 로 회전한다.
2. **해악의 실체는 회전이 아니라 드레인의 아카이브 실명(失明)이다.** `drain.sh` 는 `--inbox` 로 받은 live 파일만 읽고 `.N` 세대를 단 한 번도 참조하지 않는다. 회전된 stub 은 드레인이 영구히 도달하지 못한다.
3. **침묵 지점은 `backlog_check.sh` 였다.** 회전이 live 줄 수를 무너뜨리므로, stub 이 도달 범위를 벗어난 바로 그 순간 미독 경고가 오히려 조용해진다.
4. **이 머신은 이미 조건 하나만 남겨두고 있다.** primary inbox 는 캡의 1.78배다.
5. 가드를 세웠다 — 예방이 아니라 **관측**으로. 회전된 세대를 backlog 로 세어 경고가 침묵하지 않게 한다.

## Evidence

### R1 — 전이 재현 (`r1-transition.log`)

`t.TempDir()` 격리, 이 프로젝트의 `.moai/state/lsel/` 미접촉. 한 머신에서 마커만 제거한다.

```
STATE A (curator, marker present): inbox=1049682 bytes, cap=1048576
STATE A result: no .1 archive, inbox=1049788 bytes -> stand-down HELD
STATE B result: cap ARMED — 1049788 bytes moved to .1, live inbox now 106 bytes
HARM: drain-reachable stubs=1, archived-unreachable stubs=4183
```

t280 은 "마커 있음 → 대기"와 "마커 없음 → 회전"을 **독립된 두 상태**로 덮었다. 한 머신에서 전자에서 후자로 넘어가는 **전이**는 덮지 않았다 — `acceptance §B` 의 "가드 안 된 전이" 기록은 착지 후에도 참이다. 회귀 테스트로 고정: `internal/hook/t459_repro_test.go`.

### R2 — 아카이브 도달성 (`r2-drain-reach.log`)

실제 `.claude/skills/hns-lsel-curator/drain.sh` 를 `/tmp` 픽스처에 돌렸다. live 2건 + 아카이브 4건.

```
drain: read 2 stubs (offset 0→2) — ... 1 candidate cluster(s) emitted
UNREACHABLE: no archived stub in clusters.json
--- does drain.sh reference any archive generation? ---
0
```

`.jsonl.N` 참조 0회. 회전은 파일을 지우지 않지만(보존 2세대), 드레인 입장에서는 삭제와 구분되지 않는다.

### R3 — 근접도 (`r3-proximity.log`)

```
primary_inbox_bytes=1866583      # 캡 1048576 의 1.78배
lines=6424
offset=6409
```

캡 초과 조건은 **이미 충족돼 있다.** 이 머신과 회전 사이에 남은 것은 `.moai/state/lsel/` 디렉터리의 존재 하나뿐이다.

### 마커 부재의 도달 경로 — 실측

| 경로 | 실측 |
|---|---|
| 워크트리 | `ls .moai/state/lsel` → 이 트리에 **없음**. `.gitignore:224 **/.moai/state/` 라 어떤 워크트리에도 전파되지 않는다 |
| `moai update` | `.moai/config` 만 wipe (`CLAUDE.local.md §2.3`) — `.moai/state` 는 대상 아님 |
| 신규 클론 | gitignore 로 부재. 단 inbox 도 비어 있어 무해 |
| 수동 `rm -rf .moai/state` | 즉시 무장 |

### 가드 — RED → GREEN → 뮤턴트

| 단계 | 파일 | 결과 |
|---|---|---|
| RED | `m1-red.log` | `FAIL: t459: advisory silent after rotation — 40 archived stubs unreported; output:` (출력 완전 공백) |
| GREEN | `m1-green-final.log` | 4/4 PASS |
| M1 집계 무력화 (`ARCHIVED + 0`) | exit 1 | **사살** |
| M2 무음 조건 원복 (`-le THRESHOLD` 만) | exit 1 | **사살** |

회귀: `session_drain_test.sh` ALL PASSED · `drain_test.sh` ALL PASSED · `go test ./internal/hook/...` 11패키지 전부 ok.

## Baseline-attribution

- 트리: `.claude/worktrees/t459`, develop `5107bbfff` 흡수 직후 측정
- 명령: 위 각 로그 파일의 헤더에 그대로 있음
- 근접도 수치는 **primary 체크아웃**(`/Users/goos/MoAI/moai-adk-go`)을 읽은 값이며, 워크트리 값이 아님 — 두 트리의 inbox 는 별개 파일이다

## 리드 배차문 대비 정정 1건

배차문은 회전 손실을 "미독 stub 약 4.2k 줄"로 적었다. **4.2k 는 한 세대의 용량이지 오늘의 미독량이 아니다.** 실측 offset 6409 / 6424 줄 기준, 지금 회전이 일어나면 진짜 미독 손실은 **약 15건**이고 나머지 6409건은 이미 드레인된 것이다.

다만 이 정정이 위험을 축소하지는 않는다. 4.2k 규모의 손실은 **드레인이 멈춰 있는 동안** 마커가 사라지면 그대로 성립하며, 그 정지는 가정이 아니라 전례다(2026-08-04~25, 3주 정지 — `CLAUDE.local.md §28`). t280 설계가 그 조합에서 옳게 동작하는 부분도 있다: 정지했더라도 마커가 남아 있으면 stand-down 이 유지된다. 무너지는 것은 **정지 + 마커 부재**가 겹칠 때다.

## 가드가 이 모양인 이유

예방(회전 자체를 막는 것)은 채택하지 않았다. 마커를 대신할 durable 신호는 curator 스킬 설치 여부뿐인데, t280 이 "소유권은 스킬 설치가 아니라 드레인 활동으로 증명된다"고 명시적으로 기각한 축이다. 새 근거 없이 착지한 설계 결정을 뒤집는 것은 범위 밖이다.

대신 **측정된 침묵 지점**을 막았다. `backlog_check.sh` 는 이미 세션 시작마다 돌고, 회전 세대 수 파생은 이미 존재한다(`LessonsInboxArchiveGens`, `moai inbox status` 가 `archive_generations` 를 보고한다 — 그 표면은 멀쩡했고, 빠진 것은 **자동** 경고였다). 추가한 것은 글롭 한 줄 루프와 무음 조건의 `&& ARCHIVED -eq 0` 한 조각이다. 보존 세대 수는 Go 상수(`config.DefaultInboxArchiveGenerations`)에 있으므로 셸에 리터럴로 다시 적지 않고 글롭으로 발견한다(`CLAUDE.local.md §14`).

아카이브 줄을 전부 미독으로 세는 것은 과대보고가 아니다 — 회전은 마커 부재일 때만 일어나고, offset 파일은 그 마커 디렉터리 안에 살므로, 회전 시점의 드레인 회계상 offset 은 0이다.

## Gaps — 관측하지 않은 것

- **실제 SessionStart 배선에서의 발화**를 확인하지 않았다. 픽스처로 스크립트를 직접 호출해 검증했을 뿐, `.claude/settings.local.json` 훅 경유 실행은 재현하지 않았다.
- **회전을 이 머신에서 실제로 일으켜 보지 않았다.** 의도적이다 — primary 의 `.moai/state/lsel/` 를 지우는 것이 곧 사고다.
- 워크트리 inbox 2건(t267 318 B, v31-m4-ko-content 10 KB)이 왜 생겼는지, 즉 `CLAUDE_PROJECT_DIR` 이 워크트리를 가리키는 조건은 규명하지 않았다. 두 파일 다 캡의 1% 미만이라 이 카드의 판정에 영향이 없다.
- 아카이브 stub 의 **복구 절차**를 설계하거나 검증하지 않았다. 경고문은 별도 `--state-dir` 이 필요하다는 사실만 말하고 명령을 처방하지 않는다 — 공유 offset 을 재사용하면 live stub 을 건너뛰기 때문이며, 이 부작용은 코드를 읽어 판단한 것이지 실행해 본 것이 아니다.

## Residual-risk

- 가드는 **관측일 뿐 예방이 아니다.** 마커가 사라진 채 append 가 오면 회전은 그대로 일어난다. 달라지는 것은 그 사실이 다음 세션 시작에 보인다는 점뿐이다.
- 경고는 `backlog_check.sh` 가 배선돼 있을 때만 뜬다. 배선이 깨진 머신은 마커 부재의 원인이기도 하므로, 정확히 최악의 경우에 이 가드도 함께 침묵한다. 이 결합은 남아 있다.
- 3회째 회전부터는 가장 오래된 세대가 축출된다(보존 2). 경고를 무시하고 회전이 반복되면 손실은 그때 비가역이 된다.

---

## 재측정 — develop `515fa4acd` 흡수 후 (lane-5)

위 검증은 develop `5107bbfff` 기준이었다. 그 뒤 develop 이 215커밋 나아갔으므로 흡수하고 다시 쟀다.
흡수: `e7a3210e3` (`git merge develop`, 충돌 0).

**범위 산정** — 파일 델타 ∪ 역의존. Go 델타는 `internal/hook` 하나(`t459_repro_test.go`),
셸 델타는 `backlog_check.sh` + `backlog_check_test.sh`. 역의존은
`go list -f '{{.ImportPath}} {{join .Deps " "}}' ./...` 로 판정해 `cmd/moai`, `internal/cli`,
`codexadapter`, `codexwiring`, `feedback`, `migration/migrations`, `permission` 7개가 나왔다.

### Go 축 — `postmerge-go.log`

`internal/cli` 를 제외한 7패키지. `grep -c '^FAIL'` = **0**.

```
ok  internal/hook              149.341s      ok  internal/hook/quality      33.383s
ok  internal/hook/handoff        5.590s      ok  internal/hook/security     12.571s
ok  internal/hook/memo           2.536s      ok  internal/hook/testutil      1.441s
ok  internal/hook/memo/taxonomy  5.303s      ok  internal/hook/trace         5.792s
ok  internal/hook/mx            27.465s      ok  internal/permission         4.578s
ok  internal/hook/mx/complexity  5.974s      ok  internal/feedback           4.194s
ok  internal/hook/perf         143.119s      ok  internal/codexadapter       1.080s
                                             ok  internal/codexwiring        3.264s
```

`internal/cli` 는 **의도적으로 이 시점에 돌리지 않았다.** 600초를 넘는 스위트이고 같은 날 여러 레인이
동시에 돌려 다른 레인 단계가 죽은 전례가 있다. 통합 창 안에서 최종 병합 트리 기준으로 돌린다 —
그때가 판정에 쓰일 트리이기도 하다.

### 셸 축 — 3스위트 전부 PASS

```
backlog_check_test.sh   PASS: t459: rotation-archived stubs reported (advisory does not go silent)
                        backlog_check_test: PASS
drain_test.sh           ALL DRAIN TESTS PASSED
session_drain_test.sh   ALL SESSION DRAIN TESTS PASSED
                        (mutant probe: REJECTS the offset-only-advance mutant)
```

셋 다 `mktemp -d` + `trap 'rm -rf' EXIT` 격리다. 이 저장소의 `.moai/state/lsel/` 는 이번에도 건드리지 않았다.

### 이 재측정이 덧붙이지 못한 것

- `internal/cli` (위 사유 — 창 안으로 이월)
- 재현 자체를 다시 돌리지는 않았다. R1/R2/R3 은 `5107bbfff` 기준 관측이고, 이번에 다시 확인한 것은
  **가드가 흡수 후에도 살아 있다**는 것(`backlog_check_test.sh` 의 t459 케이스 PASS)까지다.
  전이 재현의 재실행은 회귀 테스트 `internal/hook/t459_repro_test.go` 가 `internal/hook` 통과로 대신한다.

### 미추적 스크래치 처분 — `commit-msg.txt` 삭제 (2026-09-03)

`.moai/reports/t459/commit-msg.txt` 는 미추적 상태로 남아 있었다. 판정 근거:

```
git log -1 --format=%B 0a5a01378 > /tmp/t459-msg-landed.txt
diff /tmp/t459-msg-landed.txt .moai/reports/t459/commit-msg.txt
→ 34d33 < (빈 줄 1개)   # git 이 붙이는 말미 개행 외에는 차이 없음
```

즉 이 파일은 **이미 착지한 커밋 `0a5a01378` 의 메시지 원문 그대로**이며, 증거가 아니라 그 커밋을
만들 때 쓴 작성용 스크래치다. 같은 내용이 커밋 메시지로 영구 보존돼 있으므로 사본을 남길 이유가 없다 —
삭제한다. 관측 사실은 이 절이 보존한다(조용히 지우지 않는다).
