# 카드 t560 진행 기록

- **카드**: t560 · 이슈 modu-ai/moai-adk#1691 · t516 후속
- **레인**: kanban lane
- **워크트리**: `.claude/worktrees/t560` (`origin/develop` 팁 `3ac58b5a1` 기점)
- **브랜치**: `WT-hook-env-scrub` (생성 직후 개명 완료)
- **측정 환경**: darwin 25.6.0 · `git version 2.50.1 (Apple Git-155)`

> **SPEC 없음.** t516 과 같은 형태의 카드다(원인이 이미 규명돼 있고 산출이 코드 수리와 증거뿐).
> `.moai/specs/` 아래에 t560 SPEC 은 존재하지 않으며 만들지 않았다. 3-phase close 의
> `status` 전이 대상도 없다. 리드의 sync 지시는 `progress.md §E.4` 를 지목했는데, 그 문서는
> SPEC 을 전제한다 — 이 카드에서는 `.moai/reports/t560/progress.md`(이 파일, t290·t293 선례)와
> `verdict.md` 가 그 자리를 대신한다. 지시와 실제 산출물 형태가 어긋난 지점이므로 기록해 둔다.

---

## 재범위 — 원문은 이미 처리된 결함이었다

배차 원문은 「GH #1691 수리」였다. 착수 전 전제 재확인에서 **주 결함이 이미 카드 t516 으로
수리돼 `origin/develop` 에 착지해 있음**을 확인했다:

- `internal/hook/quality/step_git_env.go` 존재, `gate.go:1169` 에서 `cmd.Env = stepEnv()` 로
  **배선됨**(공허 헬퍼 아님), 회귀 테스트 `gate_step_git_env_test.go` 동반
- GH #1691 코멘트에 운영자가 직접 "Queued as card **t516**" 이라고 답한 기록

워크트리를 만들기 **전에** 멈추고 리드에 보고 → 운영자 판정으로 **「t516 후속 1: 훅 경유
`exec.Command` 환경누출 스윕」으로 재범위**. 이 파일 이후의 모든 내용은 재범위된 카드의 것이다.

원문 그대로 진행했다면 이미 있는 수리를 다시 만들었을 것이다. 배차 전 교차확인이 값을 한 지점.

---

## 분류 — 20곳 닫힌 집합

`internal/hook` 의 실제 `exec.Command` 호출식 20개(원본 목록 `exec-sites-raw.txt`),
1+6+5+1+7 = 20:

| 분류 | 수 | 처분 |
|---|---|---|
| t516 이 이미 수리 | 1 | `gate.go:1139`, 불변 |
| 이 카드가 스크럽 | 6 | `gate.go:1065`, `agentmemory.go:140`, `navigator_detect.go:500`, `session_end.go:255`, `security/guardian.go:246`, `worktree_create.go:145` |
| 의도적 비스크럽 | 5 | `worktree_base_branch.go:145,146,162,171,189` — 주변 컨텍스트가 곧 답 |
| 동형·도달불가 | 1 | `quality/tool_registry.go:143` |
| 범위 밖(git 자식 아님) | 7 | tmux ×5, `sg` ×2 |

`prefilter.go` 는 초기 계수에서 1건으로 잡혔으나 정규식 **주석**이었다(호출식 아님).

---

## §E.4 — Gaps (미검증으로 남긴 것)

1. **스크럽 6곳 중 도달성이 확립된 것은 `gate.go:1065` 하나뿐이다.** pre-commit 훅이
   `moai gate` 를 부르고 `stagedFiles` 가 그 프로세스 안에서 돈다. 나머지 5곳은 부모가
   Claude Code 이고 Claude Code 는 `GIT_DIR` 을 export 하지 않는다 — **누출 표면을 닫은 것이지
   관측된 사고의 수리가 아니다.** 커밋 메시지·CHANGELOG 에도 이 구분을 유지했다.
2. **`quality/tool_registry.go:143`(`RunTool`) 미수리.** 동형이나(임의 자식 + `cmd.Dir` +
   `cmd.Env` nil) 호출자는 `Formatter`/`Linter` 뿐이고 그 둘은 패키지 밖 비테스트 생성자 0.
   대조군 비어있지 않음 확인: `runStep` 8히트 vs `quality.NewFormatter|NewLinter` 패키지 밖 0히트.
   죽은 코드를 고치면 그 수리는 어떤 테스트로도 검증할 수 없다. 삼각형 사장 여부는 별도 카드.
3. **제보자 환경 미재현.** Linux/WSL2, moai-adk 3.1.2. 이 카드의 모든 실측은 darwin/git 2.50.1.
4. **`GIT_CONFIG_COUNT`/`KEY_n`/`VALUE_n` 주입 경로 미착수.** t516 이 의도적으로 남긴 것이며
   (동작 변수는 설계상 통과) 원리상 `core.bare` 도 세울 수 있다. 여전히 열림 — t516 후속 2.
5. **`internal/hook` 밖 미스윕.** 전 트리 447곳/86 비테스트 파일. 이 카드는 훅 계층으로
   범위를 한정했다(git 훅이 부모일 수 있는 층).

---

## §E.4 — 위임 실패 (다음 판독자를 위한 기록)

**`manager-develop` 서브에이전트를 스폰했고, 주간 사용 한도로 작업 도중 사망했다.**

- 마지막 메시지: `"All five sites RED. Applying the fixes."` — **완료 보고가 아니었다.**
  그런데 **idle 통지는 도착했다.** 즉 통지가 종료를 뜻하지 않는 것을 넘어, 이 경우
  **통지 자체가 실패였다.**
- 실제로 남긴 트리 상태는 **컴파일 불가**였다: `agentmemory.go`·`navigator_detect.go` 두 파일에
  `cmd.Env = gitenv.Env()` 를 넣고 **import 를 넣지 않았다**.
- 미착수: Part A(`internal/gitenv` 테스트), Part C(③ 불변 가드), 5곳 중 3곳의 프로덕션 적용.

**처리**: 남긴 테스트 코드는 읽어서 품질을 확인한 뒤 채택했고, **「다섯 곳 전부 RED」라는
주장은 인용하지 않았다** — 6개 RED 를 전부 다시 관측한 뒤에야 증거로 썼다(미수리 3곳은
직접, 이미 적용된 2곳은 뮤테이션, 헤드라인 1곳은 적용 전 관측). 나머지 작업은 레인이 마쳤다.

이 사실이 기록에 없으면 다음 판독자가 위임 산출물을 **검증된 것**으로 읽는다.

---

## 이 카드가 새로 확립한 것

**GH #1691 결함 1(`core.bare` 뒤집힘)은 재현된다.** t516 은 「미검증」으로 남겼는데, 그 비재현은
**재현 모델이 발화하는 형태가 아니었기 때문**이다 — 워크트리가 **없는** 호스트로 모델을 만들어
`GIT_DIR` 이 평범한 `.git` 을 가리켰다. 판별식(격리 임시 저장소, git 2.50.1/darwin):

| 상속된 변수 | 호스트 `core.bare` |
|---|---|
| `GIT_DIR=<host>/.git/worktrees/<wt>` | **true** |
| `GIT_DIR=<host>/.git` | false |
| `GIT_INDEX_FILE=<host>/.git/worktrees/<wt>/index` | false |

프로덕션 코드로도 재현했다 — `gate.go:1139` 의 스크럽을 뮤테이션으로 걷어내니
`core.bare: "false" before, "true" after`.

t516 이 자기 파일에 적은 **「재현 모델이 실제보다 강하면 결함을 숨긴다」의 반대 방향**이다.
모델이 실제보다 **약해서** 놓쳤다. **「재현되지 않았다」와 「발화하는 형태를 아직 안 쟀다」는
같은 출력(변화 없음)을 낸다.**

---

## 현재 단계

✅ run 완료(`30e9f37ef`) · sync 진행 중 · 병합·push 없음(리드 소관).
판정 근거는 `verdict.md`(5구역).
