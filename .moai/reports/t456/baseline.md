# t456 — 구현 전 baseline (리드 지시 ⑴)

리드 지시: *"baseline 을 구현 전에 재십시오. 구현 후에 재면 회귀 근거가 구현 전후로 갈립니다."*
본 문서가 그 구현 전 측정이다.

## 측정 조건

- 트리: `.claude/worktrees/t456` · 브랜치 `WT-statusline-landed`
- HEAD: develop `d63dee78d` 흡수 후의 병합 커밋 (t305 `e971b7e25`, t360 `d63dee78d` 포함)
- 판정 바이너리: `go build -o bin/moai-t456 ./cmd/moai` — **이 트리에서 빌드한 것**
  (설치본이 아니다. 도구-출처 귀속: 트리와 판정 빌드가 같아야 한다)
- base arm 보존: `/tmp/t456_moai_base` — 구현 후 교대 비교에 쓴다
- 로컬 darwin. **CI 아님.**

### t305 흡수를 먼저 한 이유

리드가 예고한 대로 t305(`WT-statusline-git-spawn`)가 창에서 develop 에 들어왔다.
같은 statusline 렌더 경로를 건드린 카드이므로, 흡수 **전** 값으로 baseline 을 잡으면
회귀 근거가 남의 변경 위에서 갈린다. 흡수 후 `d63dee78d` 기준으로 다시 쟀다.

t305 접촉 파일은 `internal/statusline/git.go`, 본 카드 범위는 `renderer.go` + 신규
`landed.go` — **파일 겹침 없음.**

## 측정 1 — 렌더 경로 git spawn 수 (구속력 있는 값)

계측은 t305 가 남긴 spawn 카운팅 shim 을 **그대로 재사용**했다
(`.moai/reports/t305/shim/git` → `.moai/reports/t456/shim/git`, 로그 변수명만 분리).
PATH 선두에 두면 모든 `git` 호출이 1행씩 기록된 뒤 실제로 실행된다. 부하와 무관하게
정확하므로 타이밍 인상이 아니라 증거다.

```
$ PATH="$PWD/.moai/reports/t456/shim:$PATH" ./bin/moai-t456 statusline < payload.json
rev-parse --git-dir --show-toplevel
status --porcelain --branch
→ 2
```

**baseline = 렌더당 git spawn 2회.** 구현 후 이 값이 **2로 유지**되어야 한다
(캐시 읽기는 파일 1회, 서브프로세스 0회).

## 측정 2 — 렌더 출력 (결함 자체)

```
🤖 Opus 5 | 🔅 v2.1.246 | 🗿 v3.1.3 | 💬 MoAI
🔋 5H: ░░░░░░░░░░ 0% | 🔋 7D: ░░░░░░░░░░ 0%
📁 . | 🅱️ WT-statusline-landed +3 | 💾 +0 M0 ?3
🔄 TODO: 76/4
```

`🔄 TODO: 76/4` — 76 중 48이 이미 `origin/develop` 이력에 이름이 있는데도
그 판정을 묻지 않는다. 이것이 카드가 지목한 표시 축이다.

## 측정 3 — 렌더 지연 (참고값, 구속력 없음)

```
$ uptime
10:09  up 5 days, 22:36, 27 users, load averages: 16.48 18.61 16.57

404.2 / 632.5 / 405.5 / 401.6 / 296.6 / 369.4 / 351.5 ms  (n=7)
```

**이 수치는 구속력이 없다.** t305 자신이 남긴 방법론 노트가 이유를 적어뒀다 —
before/after 를 몇 분 간격으로 재면 그 사이 부하가 달라져(그쪽 사례: load 9.49 vs 5.66)
절대 밀리초가 비교 불가능해진다. 지금 부하는 16.48로 t305 측정 시점(6.68)의 2.5배다.

그래서 t305 는 `paired_bench.py` 로 **두 바이너리를 한 실행 안에서 교대**시켰다.
본 카드도 같은 규율을 따른다: base arm 을 `/tmp/t456_moai_base` 로 보존해 두고,
구현 후 교대 실행으로 비교한다. 위 n=7 은 "구현 전에 한 번 쟀다"는 기록일 뿐,
회귀 판정의 근거는 교대 실행 쪽이다.

참고 — t305 가 측정한 교대 실행 결과(그쪽 트리, 그쪽 시점):
`base median 323.23ms → after median 151.72ms`.

## 상속된 적색 — 본 카드 소관 아님

`make build` 가 선행 검사 `agents-emit-check` 에서 실패한다:

```
agent-emit drift: committed .codex/agents/moai/*.toml differ from the .md source layer
```

develop 팁 `d63dee78d` 에서도 동일하게 실패하며, 본 브랜치가 develop 대비 더한 것은
`.moai/reports/t456/` 뿐이다(`git diff --stat develop..HEAD`). **상속된 적색이고 고치지 않는다.**
측정 바이너리는 `go build ./cmd/moai` 로 우회 빌드했다.

## Gaps — 이 문서가 관측하지 않은 것

- 캐시 부재·손상 경로의 렌더 출력. 구현 전에는 그 경로 자체가 없으므로 잴 대상이 없다.
  구현 후 반드시 관측한다(리드 지시 ⑵).
- ref SHA 를 서브프로세스 없이 읽는 경로(`.git/refs/...` vs `packed-refs`) 미실측.
- 위 spawn 2회는 이 payload 한 벌에 대한 값이다. 다른 세그먼트 조합(예: GitHub 세그먼트가
  TTL 만료 상태)에서는 자식 spawn 이 더 붙을 수 있으며, 그것은 t305/github.go 소관이다.
