# t401 인계 판독 (lane-5, 2026-09-06)

card: t401 · SPEC-JUDGMENT-FIRST-MODE-001 v0.2.3 · status draft · Tier L
worktree `.claude/worktrees/t401` · branch `WT-analysis-pull` · HEAD `53f54e919`
전 소관 lane-6 소멸. **아래는 전부 이 세션이 이 트리에서 직접 잰 값이며, lane-6 기록의 인용이 아니다.**

## Claim

1. 작업 유실 없음. ahead 5는 **기록 전용**이며 코드 0.
2. M0 관측자 코드는 **이미 develop 조상**이다. 게이트 순서 위반은 기록돼 있고, 비준/되돌림은 운영자 결정으로 남아 있다.
3. 멈춘 지점은 **plan-phase 감사 iter-5 FAIL 0.95**. 그 판정이 근거로 든 기록 결함 2건은 `53f54e919`에서 수리됐고, iter-6은 리드 승인 대기다.
4. develop이 725커밋 나아갔음에도 **이 SPEC의 전제는 유효하다** — 대상 교리 4파일 전부 base 대비 무변경.
5. **다만 RED-now 셀 2개가 이동했다.** AC-JFM-023 half 1은 GREEN이 됐고, half 2의 원천 로그는 이제 실재하며 3/5까지 차 있다.
6. 리드가 지시한 채택 순서와 authority 공백 명시는 **plan 산출물에 이미 반영돼 있다.** 미이행분은 리드 상신 하나뿐이다.
7. 신규 결함 1건 — `calls_issued`는 **감사 대상 자신의 자기 신고**이고 독립 원천이 없다. 네 갈래 대조가 그 지점에서 공허하게 통과할 수 있다.

## Evidence

```
$ git rev-list --count --left-right refs/heads/develop...HEAD
725	5
$ git diff --stat 6352897a5..HEAD | tail -1
 17 files changed, 2636 insertions(+)          # 전부 .moai/specs + .moai/reports
$ git merge-base --is-ancestor d5caf2d8e origin/develop      ; echo rc=$?   → rc=0
$ git merge-base --is-ancestor d5caf2d8e refs/heads/develop  ; echo rc=$?   → rc=0
$ git ls-tree -r --name-only refs/heads/develop -- internal/hook/askuser_observer.go
internal/hook/askuser_observer.go
$ git ls-tree -r --name-only refs/heads/develop -- .moai/specs/SPEC-JUDGMENT-FIRST-MODE-001 .moai/reports/t401
(무출력 — 기록은 이 브랜치에만 있다)
```

교리 4파일 드리프트 — `git diff --quiet ad272be20 refs/heads/develop -- <path>`:

| 경로 | 판정 |
|---|---|
| `.claude/rules/moai/core/askuser-protocol.md` | UNCHANGED |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | UNCHANGED |
| `.claude/skills/moai/workflows/run.md` | UNCHANGED |
| `.claude/rules/moai/development/branch-origin-protocol.md` | UNCHANGED |

핀 좌표 축자 확인 (`git show refs/heads/develop:<path> | sed -n '<n>p'`) — 4/5가 핀 라인 그대로:
`askuser-protocol.md:64` → `3. **First option label**: MUST carry the \`(권장)\` …` ·
`spec-assembly.md:353` → `- First option: the recommended Choice with \`(권장)\` suffix …` ·
`spec-assembly.md:212` → `… or the \`(권장)\` first-option label. A` ·
`branch-origin-protocol.md:25` → `- [ZONE:Frozen] [HARD] Skill body BODP gate MUST follow … \`(권장)\` first,`

**AC-JFM-023 half 1 — RED-now 무효, 이제 GREEN:**

```
$ go test ./internal/hook/ -run 'AskUserQuestionObserver.*LabelDetect' -count=1 -v
=== RUN   TestAskUserQuestionObserverLabelDetect
--- PASS: TestAskUserQuestionObserverLabelDetect (0.01s)
    --- PASS: …/KoreanLabelPresent      --- PASS: …/EnglishLabelPresent
    --- PASS: …/LabelOnNonFirstOption   --- PASS: …/NoLabel
    --- PASS: …/EmptyOptions
ok  	github.com/modu-ai/moai-adk/internal/hook	0.453s
```
양방향(양성 2 · 음성 2)이 명명된 서브테스트로 존재한다. 셀이 적은 `[no tests to run]`(빈 스윕)은 `ad272be20` 시점 사실이고 지금은 거짓이다.
**주의**: `./internal/hook/...`(재귀)로 돌리면 하위 7패키지가 전부 `[no tests to run]`로 나와 실제 결과가 tail에서 밀려난다. 판별식은 `./internal/hook/` 단일 패키지 줄이다.

**AC-JFM-023 half 2 원천 — 관측자는 살아 있고 기록 중이다:**

```
$ ls -l /Users/goos/MoAI/moai-adk-go/.moai/logs/askuser-observations.jsonl
-rw------- 546 Sep 5 20:40
$ jq -s '{n:length, modes:…, positives:…, sessions:…}' …
{"n":3, "modes":{"push":3}, "positives":3, "sessions":2,
 "first":"2026-09-03T19:19:29Z", "last":"2026-09-05T11:40:42Z"}
```
`positives 3 >= 1` 충족 · `n 3 < 5` 미달(2행 부족). 세 행 전부 `mode: push`·`label_present: true`.
plan이 적은 "no observer log exists"는 그 시점 사실이고 **지금은 거짓**이다.

**리드 지시 2건의 현황** (`spec.md` 직독):
`409:### Out of Scope — proposal item #1 (decision gate)` — 제안 1번은 이미 범위 밖.
`420:### Out of Scope — authority-source definition (carry-forward)` — 공백이 명시돼 있고,
사유까지 적혀 있다(`424-425`: pull 모드는 authority 원천을 **아예 참조하지 않는다**; 라벨을 보류하는 출력 규약일 뿐).

## Baseline-attribution

전 측정은 2026-09-06, 워크트리 `.claude/worktrees/t401`, HEAD `53f54e919`, 비교 대상은
로컬 `develop` `3084f1071` 과 SPEC 기준 트리 `ad272be20`. 관측자 로그만 primary 체크아웃
(`/Users/goos/MoAI/moai-adk-go`)에서 읽었다 — 그 로그는 워크트리가 아니라 **묻는 세션의
`CLAUDE_PROJECT_DIR`** 아래 쌓이며, 이는 AC-JFM-018/023이 이미 적어 둔 성질이다.

## Gaps

- **iter-6 감사를 돌리지 않았다.** 리드 승인 사항이라 레인이 자가 착수하지 않는다.
- **`push` baseline 창을 수출하지 않았다.** `n>=5` 미달이고, 수출은 M0 산출물이라 Kickoff 게이트 뒤다.
- **AC-JFM-018 쪽(`pull` 창)은 손대지 않았다.** M1 이후 사안이다.
- **acceptance.md 616행 중 §D.6(363-511)만 정독**했다. 나머지 AC의 공허성은 이 회차에서 재지 않았다.
- 코드 검증은 `internal/hook` 단일 패키지 한정. 전체 스위트는 안 돌렸다(CI 몫).

## Residual-risk

**신규 결함 — `calls_issued` 대조의 한쪽 피연산자가 자기 신고다.**
AC-JFM-018 half 3 / AC-JFM-023 half 4의 네 갈래 대조는 `rows_recorded`(관측자가 남긴 것)와
`calls_issued`(**묻는 세션이 스스로 센 것**)를 맞춰 본다. 앞쪽은 기계 산출이지만 뒤쪽은
감사 대상 자신의 자기 보고이고, 셀은 그것을 "관측자 없이도 아는 값"이라는 **장점**으로만 적어 두었다.
독립 원천이 없으므로, 묻는 세션이 `calls_issued`를 `rows_recorded`에 맞춰 적으면
`rows_recorded == calls_issued`가 성립해 **부분 유실과 미배선이 그대로 통과**한다.
대조가 잡아내겠다고 선언한 바로 그 두 상태다. 이 성질은 어느 셀에도 gap으로 기재돼 있지 않다.

완화 후보(레인이 정하지 않음): ① `calls_issued`를 독립 원천에 결속 —
동일 구간의 트랜스크립트 `tool_use` 계수 등, ② 결속이 불가능하면 그 한계를 셀 안에
**명시적 gap**으로 적어 "대조가 자기 신고에 의존한다"가 판독자에게 보이게 할 것.

## 판정

**「이미 끝났다」도 「전제가 무너졌다」도 아니다.** plan 산출물은 서 있고 전제도 유효하다.
이 카드가 막혀 있는 지점은 레인 작업이 아니라 **게이트 2개**다 —
(1) iter-6 감사 승인, (2) Implementation Kickoff Approval(운영자 안건 3건 + M0 사전 착지 비준/되돌림).
레인이 지금 자가 착수할 수 있는 작업은 없다.
