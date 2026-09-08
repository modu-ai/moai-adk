# t478 — 레인 독립 검증 (완료 보고를 읽지 않고 다시 잰 기록)

검증 트리: `.claude/worktrees/t478`, 브랜치 `WT-graph-gate-restamp`, HEAD `46f6a3236`
검증자: lane-4 (구현은 `manager-develop`, SPEC 은 `manager-spec`)
일자: 2026-09-04 · 전부 **로컬 측정** (CI 아님)

이 문서는 하위 에이전트의 완료 보고를 인용한 것이 아니라, 레인이 같은 항목을
**직접 다시 실행해** 얻은 출력이다.

## Claim

1. 이 카드가 겨눈 거짓 초록 — "코드맵 본문을 그대로 둔 채 스탬프만 다시 찍으면
   게이트가 통과한다" — 이 실물 트리에서 **닫혔다.**
2. `internal/graph` 패키지는 전부 통과한다.
3. `internal/cli` 의 잔여 실패 1건은 **이 카드 소관이 아니다.**

## Evidence

### E1 — 실물 트리 맨 재스탬프: 수리 전 대 수리 후

동일한 3단계 시퀀스(check → `graph stamp codemaps`(본문 미편집) → check)를
같은 워크트리에서, 각각 그 시점 소스로 빌드한 바이너리로 실행했다.

수리 전 (HEAD `456665e8d`, 빌드 `./bin/moai`):

```
1) codemaps  metric=described-source-diff value=146 threshold=40 verdict=stale
2) OK: stamped .../provenance.json   commit=456665e8d59e
3) codemaps  metric=described-source-diff value=0   threshold=40 verdict=fresh
```

수리 후 (HEAD `46f6a3236`, 빌드 `./bin/moai`):

```
1) codemaps value=146 verdict=stale
     content_anchor=ad272be20abff9e4f3b1b363fce3e48dac4c5132
     content_anchor_source=working-tree-differs-from-stamp
2) OK: stamped .../provenance.json   commit=46f6a3236558
3) codemaps value=134 verdict=stale
     content_anchor=732995c0a1fbb18a60f9b1a0fff9fc66ac6cabf3
     content_anchor_source=last-body-change
```

재스탬프 뒤에도 앵커가 재스탬프한 HEAD 로 따라오지 않고 **본문이 마지막으로 실제로
바뀐 커밋** `732995c0a` 으로 되돌아간다. 이것이 이 카드의 핵심 주장이다.

### E1a — [HARD] 판별식은 프로세스 종료코드가 아니라 codemaps 행이다

두 경우 모두 `graph check` 의 rc 는 **1** 이다. 이 워크트리는 `mx-index` 와 `edges`
가 미추적 런타임 산출물이라 부재이고, 그 두 행만으로도 rc 가 1 이 되기 때문이다.
따라서 **rc 만으로는 수리 전후를 구별할 수 없다** — 구별하는 것은 codemaps 행의
`verdict`(fresh → stale)와 `value`(0 → 134)다. CI 는 이 두 층을 부트스트랩하므로
거기서는 codemaps 가 fresh 이면 잡 전체가 초록이 되었을 것이다.

rc 는 파이프 없이 읽었다: `./bin/moai graph check --json > file; echo $?`.

### E2 — 패키지 테스트

```
$ go test ./internal/graph/... -count=1        → rc=0
ok  github.com/modu-ai/moai-adk/internal/graph         24.346s
ok  github.com/modu-ai/moai-adk/internal/graph/symbol   0.431s

$ go test ./internal/cli/... -count=1          → rc=1
--- FAIL: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.05s)
FAIL  github.com/modu-ai/moai-adk/internal/cli   398.938s
(그 외 하위 패키지 16개 전부 ok)
```

### E3 — 잔여 실패 1건의 귀속 (추론이 아니라 구조로 판정)

`TestBinaryLag_DoctorCheckNameSetIsUnchanged` 는 `internal/cli/binary_lag_test.go:188`
에서 baseline blob 의 `internal/cli/doctor.go` 와 작업 트리의 `doctor.go` 를 비교한다.

이 카드의 run-phase 변경 파일 전체(`git diff --name-only cd28923f2 46f6a3236`)는
아래와 같고, **`internal/cli/doctor.go` 는 들어 있지 않다**:

```
internal/cli/graph_check.go
internal/graph/check.go
internal/graph/check_regression_lock_test.go
internal/graph/check_restamp_anchor_test.go
+ .moai/specs/SPEC-GRAPH-GATE-RESTAMP-001/*  (4)
+ .moai/reports/t478/*                       (15)
```

테스트가 읽는 파일을 이 카드가 건드리지 않았으므로, 이 실패는 이 카드가 만들 수
없다. (하위 에이전트는 되돌려 재실행하는 방식으로 같은 결론에 도달했다 — 두 경로가
일치한다.)

### E4 — 작업 트리 위생

```
$ git status --porcelain .moai/project/codemaps/     (무출력)
$ git status --short                                  (무출력)
```

E1 의 재스탬프 실험은 `provenance.json` 을 만지므로 사전에 사본을 떠 두고 사후에
되돌렸다. 되돌린 뒤 추적 파일 수정 0 을 확인했다.

## Baseline-attribution

- E1 의 두 블록은 각각 HEAD `456665e8d` / `46f6a3236` 에서 `go build -o ./bin/moai
  ./cmd/moai` 로 그 트리 소스에서 빌드한 바이너리의 출력이다. 설치본
  (`~/go/bin/moai`)은 어느 단계에서도 쓰지 않았다.
- E2 는 HEAD `46f6a3236` 에서의 실행이다.
- E3 의 파일 목록은 `git diff --name-only cd28923f2 46f6a3236` 의 전체 출력이다.

## Gaps — 재지 않은 것

- **전체 스위트(`go test ./...`)를 돌리지 않았다.** 레인 규율상 금지이며, 전 패키지
  판정은 CI 몫이다. 따라서 `internal/graph` · `internal/cli` 밖 패키지에 대한 영향은
  이 문서가 아무것도 주장하지 않는다.
- `golangci-lint`, Windows/Linux 크로스 빌드 미실행.
- **CI 상의 Graph Freshness 잡 실측 없음.** 위 수치는 전부 로컬이다. 특히 E1a 가
  말하는 "CI 였다면 잡 전체가 초록" 은 코드 구조에서 따라오는 추론이지 관측이 아니다.
- 하위 에이전트가 돌린 뮤턴트 3종(`ls-files --others` 제거 / dirty 경로 앵커 누출 /
  C1 을 시스템 오류로)은 **레인이 재실행하지 않았다.** 그 증거는
  `mutant-1-drop-lsfiles-others.txt`, `mutant-2-anchor-leaks-into-dirty-path.txt`,
  `mutant-3-c1-as-system-error.txt` 이며 하위 에이전트 측정이다.

## Residual-risk

- **미커밋 사소 편집 우회는 열려 있다.** 코드맵 본문에 공백 한 글자를 미커밋으로
  넣고 스탬프하면 규칙 A 가 발화해 앵커가 S 가 되고 통과한다. 의도적 행위라
  "가장 값싼 경로" 는 아니지만 닫히지 않았다 — SPEC §I 에 범위 밖으로 명시돼 있다.
- **C2 분기는 픽스처가 없다.** 얕은 경계·루트 커밋이 모든 파일을 ADDED 로 보고하므로
  git 으로 도달할 수 없다는 측정에 근거해 의도적으로 비워 두었다. 도달 못 하는
  픽스처의 초록은 공허한 초록이라는 판단이며, 선언하고 남긴 것이지 누락이 아니다.
- **`internal/cli/graph_check.go` 는 `plan.md` §D 의 선언된 파일 집합 밖이다.**
  `--help` 문구가 이 변경으로 거짓이 되기 때문에 함께 고쳤고, `progress.md` §E.2.1
  에 숨기지 않고 기록했다. 범위 규율상 지적받을 수 있는 지점이라 여기에도 남긴다.
