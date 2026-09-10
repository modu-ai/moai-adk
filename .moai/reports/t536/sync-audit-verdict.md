# sync-audit 판정 — SPEC-TODO-HOME-TEMP-GUARD-001 (카드 t536)

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536` · 브랜치 `WT-home-fallback` · HEAD `029ab039f`(감사 시작·종료 시점 모두 재판독, 이동 없음).
감사자가 생성한 모든 증거는 `.moai/reports/t536/sync-audit/` 아래에 반출했고, 아래 인용은 전부 그 파일을 지목한다. 반출하지 않은 자료는 근거 자리에 두지 않았다.

**최종 판정: PASS-WITH-DEBT** (가중 조화평균 **87.8**, must-pass 두 축 모두 독립 통과)

---

## 차원 점수

| 차원 | 점수 | 판정 | 증거 (이 실행, 이 트리) |
|---|---|---|---|
| Functionality (40%) · must-pass | 88/100 | PASS | `go test ./internal/kanban/ -run '<AC 8종>' -count=1 -v` → `--- PASS` 18건 / `--- FAIL` 0건, 스윕 집합 비어 있지 않음 확인 (`audit-ac-kanban-v.txt`). `go test ./internal/cli/ -run '<AC>' -timeout 900s` → PASS 6 / SKIP 1(선언된 수용 손실) / FAIL 0 (`audit-ac-cli-v.txt`). 독립 뮤턴트 4종 주입: M-AUD-1·3·4 전부 포착, M-AUD-2 **미포착**(F1) — `audit-mutants.txt`, `audit-mutants2.txt` |
| Security (25%) · must-pass | 94/100 | PASS | `go vet ./internal/kanban/ ./internal/cli/ ./internal/web/` → 출력 없음, `vet exit=0` (`audit-quality.txt`). `/usr/bin/grep -nE "exec\.Command\|os\.Remove\|RemoveAll\|WriteFile\|MkdirAll\|Rename\|password\|token\|secret\|apikey" internal/kanban/temp_origin.go internal/cli/todo.go` → `temp_origin.go` 0건, `todo.go` 6건 전부 파서 어휘("token")를 담은 **주석 산문**이며 자격증명 아님. 순수성은 `TestTodoQueueRoot_PureGuardIsSilent` PASS 로 재확인 (`audit-pureguard.txt`) |
| Craft (20%) | 85/100 | PASS | `go test ./internal/kanban/ -cover -count=1` → `coverage: 86.3% of statements`(목표 85% 상회). `gofmt -l` → 이 카드 파일 0건 (`audit-quality.txt`) |
| Consistency (15%) | 82/100 | PASS | `git diff 6b71fdaa5..HEAD -- .moai/specs/SPEC-{WEB-TODO-QUEUE,STATE-ANCHOR}-001/` → 양쪽 **0 바이트**, 대조군으로 같은 범위 `.moai/specs/` 는 455줄 변경 (`audit-siblings.txt`). 커밋 4본 전부 Conventional Commits + 카드 id |

가중 조화평균: `1 / (0.40/88 + 0.25/94 + 0.20/85 + 0.15/82)` = **87.8**. must-pass 방화벽: Functionality 88 · Security 94, 둘 다 기본 임계 독립 충족 → 방화벽 미발화.

---

## Claim

1. AC 8건은 실제로 PASS이며, 대표 부분집합을 감사자가 이 트리에서 재유도했다. 공허하게 통과하는 항목도, 빈 스윕 집합 위에서 통과하는 항목도 없다.
2. AC-THG-007 의 route (i) 대조군은 가드를 단일 차이로 실제로 격리한다. 감사자가 독립 재현했다.
3. 뮤턴트 E 가 `internal/cli` 게이트 우회 테스트에서 못 잡히는 것은 **진짜 층 경계**이며, 그 테스트가 공허해진 것이 아니다.
4. AC-THG-002 의 darwin 한정 판정은 정직하게 범위 지어져 있다.
5. D8 판독(「성질은 성립하고 baseline 이 낡았다」)은 양쪽을 재측정한 결과 옳다. 금지 패키지 위반은 숨어 있지 않다.
6. 형제 SPEC 산출물 무수정은 바이트 수준에서 참이고, AC-WTQ-008 의 존속과 AC-SA-011 의 가지 이행은 실질적으로도 정당하다.
7. §C.1 은 14행이며 각 처분이 개별 이행됐다. 「수용된 손실」은 포기가 아니라 구성 불가다.
8. 미해결 부채 3건이 남았다 — 그중 1건은 이 감사가 새로 찾았다.

---

## Evidence

### ① AC 재유도 — 공허성·빈 집합 검사 (지적 1)

```
$ go test ./internal/kanban/ -run '<AC 8종 셀렉터>' -count=1 -v
--- PASS 18건 / --- FAIL 0건
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.880s
```

스윕 집합이 비어 있지 않음을 **세어서** 확인했다(`-run` 셀렉터가 0건을 물어도 `ok` 를 찍는 부류를 배제). 서브테스트까지 열거된 전문은 `audit-ac-kanban-v.txt`.

공허성은 세는 것으로 끝내지 않고 **직접 주입한 뮤턴트 4종**으로 판정했다.

| 뮤턴트 | 무엇을 깨는가 | 결과 |
|---|---|---|
| M-AUD-1 | `TempOriginReason` → 상수 `("", false)` | **7개 테스트 함수 FAIL**, 2개 패키지 걸쳐. AC-THG-001(a)(b)(c)·002·005·006 전부 비공허 |
| M-AUD-2 | `defaultTempRoots` → `{os.TempDir()}` 만 | **전부 GREEN — 미포착** (F1) |
| M-AUD-3 | 가드를 `homeTodoQueueRoot` ok 검사 **뒤로** 이동 (D12 배치 주장) | 갈래 (c) + `HomeUnresolvableWritesNothing` FAIL — 배치 주장이 실제로 고정돼 있다 |
| M-AUD-4 | 대체 루트를 `base/.moai/state/todo` 로 (D10 계층 어긋남) | kanban 6건 + cli 2건 FAIL — 읽힘 단언이 실제로 계층 오류를 잡는다 |

세 번 모두 주입 전 트리 청결 확인 → 주입 → 실행 → 복원 → **sha256 재대조**로 닫았다. 복원 후 세 파일 해시가 baseline(`baseline-hashes.txt`)과 바이트 동일:
`0ed1a6d4…`(`temp_origin.go`) · `5c63194e…`(`todo_root.go`) · `d1aaa16f…`(`todo.go`).

### ② route (i) 대조군 독립 재현 (지적 2)

먼저 등가성을 판정했다. `git diff HEAD~3 HEAD~2 -- internal/kanban/todo_root.go` 가 보이는 M2 배선은 **삽입 두 곳뿐**이고, route (i) 패치는 정확히 그 두 곳의 역이다. `tempOriginSubstituteRoot` 자체는 남기지만 해상도 경로에서는 호출되지 않으므로 런타임 영향이 없다.

**한 가지 비등가가 남는다**: route (i) 는 `internal/cli/todo.go` 의 `warnTempOriginQueueRefusal`(PersistentPreRun 배선)을 그대로 두는데, pre-guard 트리에는 그 함수가 없었다. 우회 테스트는 `newTodoCmd().Execute()` 를 돌리므로 그 PreRun 이 **실제로 실행된다**. 다만 그 경로는 읽기 전용 `TempOriginRefusal` 을 호출해 stderr 한 줄을 쓸 뿐 파일 시스템을 건드리지 않으므로, **측정 축(canary HOME 아래 오염 계수)에서는 등가**다. 측정되지 않는 축에서만 다르다.

감사자 재현:

```
$ (route (i) 주입) go test ./internal/cli/ -run TestGuardBypassMutant_ObserveHomePollution -count=1
--- FAIL: TestGuardBypassMutant_ObserveHomePollution (0.08s)
    todo_axisa_guard_test.go:202: the mutant produced home pollution under
    …/TestGuardBypassMutant_ObserveHomePollution1548414414/001/.moai/todo (1 entr(ies))
$ (가드 복원) go test ./internal/cli/ -run TestGuardBypassMutant_ObserveHomePollution -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.809s
```

주입 1건 / 복원 0건 — 커밋된 `m3/routei-cli-bypass.txt` 와 같은 수치다. 복원 후 세 파일 해시 재대조 완료. 전문: `audit-routei.txt`.

**판정: 대조군은 정당하다.** 두 실행의 차이는 가드 하나이고, 그것이 「가드가 일했다」와 「뮤턴트가 안 돌았다」를 가른다.

### ③ 뮤턴트 E 미포착 — 경계인가 공허인가 (지적 3)

`TestGuardBypassMutant_ObserveHomePollution` 은 뮤턴트 E 아래에서 전제(`TempOriginReason(nonGit)` 가 임시여야 함)에서 멈춘다. 이것이 약화인지 판정하려면 **그 테스트가 무언가에는 실제로 FAIL 하는지**를 봐야 한다. 감사자의 M-AUD-4 실행이 그 답이다:

```
--- FAIL: TestGuardBypassMutant_ObserveHomePollution (0.13s)   ← M-AUD-4 아래
```

즉 이 테스트는 **실패할 수 있는 테스트**다. 부재 단언만 갖고 있지 않고, ⑴ 반환 루트 = 대체 루트 ⑵ 카드가 project-local 큐에 실제 착지, 두 **양성 단언**을 함께 갖기 때문이다. 뮤턴트 E 아래 전제에서 멈추는 것은 설계대로다 — 부재 단언은 base 가 임시로 분류될 때만 무언가를 말하고, 판별식을 죽이면 그 전제가 무너지므로 공허한 초록 대신 전제 위반을 보고하는 편이 옳다. 그리고 판별식 회귀는 kanban 층이 잡는다(M-AUD-1 에서 7건 FAIL).

**판정: 진짜 층 경계다.** `guard-boundary.md` 는 이 사실을 「함께 보고한다」 절에 명시적으로 적었고, 다음 사람이 뮤턴트 E 로 오염 재현을 기대하다 오진하는 경로까지 닫았다. 공허화가 아니다.

### ④ AC-THG-002 darwin 한정 (지적 4)

```
$ /usr/bin/grep -n "runtime.GOOS|go:build|Skip" internal/kanban/temp_origin_test.go \
    internal/kanban/todo_root_temp_guard_test.go internal/cli/todo_temp_guard_test.go
(0건)
```

GOOS 게이팅도 skip 도 없다 — 테스트는 모든 플랫폼에서 **돈다**. 리눅스처럼 임시 트리가 심링크 뒤에 있지 않은 환경에서는 `resolved == unresolved` 가 되어 단일 케이스로 축퇴하고, 테스트는 그 사실을 `t.Logf` 로 **기록한다**("spellings coincide on this platform"). 실제로 이 기계에서는 두 철자가 갈라졌음이 M-AUD-1 출력에 남아 있다:
`unresolved="/var/folders/kt/…" resolved="/private/var/folders/kt/…"`.

**판정: 정직하게 범위 지어져 있다.** 카드는 darwin PASS 를 리눅스 판정으로 재사용하지 않았고, 테스트 자신이 어느 칸을 쟀는지 로그로 남긴다. 크로스플랫폼인 척하지 않는다.

### ⑤ D8 — 양쪽 재측정 (지적 5)

```
$ git diff 412c8cb14..HEAD -- internal/statusline/ internal/config/ internal/hook/ internal/session/ internal/stateanchor/ | wc -c
18650
$ git diff 6b71fdaa5..HEAD -- (같은 경로) | wc -c
0
$ git diff --stat 6b71fdaa5..HEAD | tail -1        # 빈 결과의 대조군
43 files changed, 3226 insertions(+), 24 deletions(-)
$ git merge-base --is-ancestor 412c8cb14 6b71fdaa5 ; echo $?
0
```

리터럴 diff 안의 6개 파일을 **하나씩 귀속**했다:

| 파일 | 최초 커밋 |
|---|---|
| `internal/config/defaults.go` · `types.go` · `testdata/shipped_key_inventory.yaml` | `12c7782f8` merge: absorb origin/develop (**카드 t401**) |
| `internal/config/interview_recommendation_mode_test.go` | `099c7bbe4` feat(**t401**) |
| `internal/stateanchor/stateanchor.go` · `stateanchor_test.go` | `5df939476` feat(SPEC-STATE-ANCHOR-VALIDATE-001) (**t537**) |

t536 커밋에 귀속되는 파일은 0건이다. 감사자가 재생성한 리터럴 diff 는 커밋된 `m3/d8-literal.diff` 와 **바이트 동일**(`diff` 무출력). 전문: `audit-d8.txt`, `audit-d8-literal.diff`, `audit-d8-card.diff`.

**판정: 「성질은 성립하고 baseline 이 낡았다」가 옳은 판독이다.** 접촉 금지 위반이 리터럴 diff 안에 숨어 있지 않다. 카드가 이 사실을 수리하지 않고 **보고한** 것도 옳다 — plan.md 본문은 manager-spec 소관이다.

### ⑥ 형제 SPEC — 바이트 주장과 실질 주장 (지적 6)

바이트 주장: 두 SPEC 디렉터 모두 두 범위(`6b71fdaa5..HEAD`, `412c8cb14..HEAD`)에서 **0 바이트**. 대조군으로 같은 범위의 `.moai/specs/` 는 이 카드 자신의 2파일 455줄이 변경됐다 — 즉 필터가 도달하지 않은 침묵이 아니다. 두 디렉터의 실재도 `ls -d` 로 확인했다(`audit-siblings.txt`).

실질 주장 판정:

- **AC-WTQ-008 은 정직하게 존속한다.** 그 AC 의 Given 은 AC-WTQ-006 에서 상속되며 주어가 "git 이 primary checkout 으로 해석하지 못하는 디렉터"다 — 「임시」라는 말이 전제에 없다. 이 카드는 홈 폴백에 도달하는 **모집단을 줄일** 뿐, 비임시 비git base 에 대한 보증은 그대로다. 그리고 그 보증은 실제로 시험되고 있다: `TestResolveTodoQueueRootAdopting_AdoptsLocalQueue`(A3 행, AC-WTQ-008 산출)가 `declareNonTemporary(t)` 를 달고 **원 단언을 유지한 채** PASS 한다(`audit-ac-kanban-v.txt`, `audit-dispositions.txt`). 축소가 아니라 좁힘이다.
- **AC-SA-011 의 가지 이행은 소급 재해석이 아니다.** REQ-SA-011 의 두 번째 가지(「오염을 만들지 못한 뮤턴트는 그 사실과 가드 경계와 함께 보고하라」)는 이 카드보다 **먼저** 그 SPEC 에 쓰여 있었고, 그 SPEC 의 §5 가 이 결정을 후속 카드에 위임했다. 이행의 산출물(`guard-boundary.md`)이 실재하고 plan.md 가 요구한 세 항목을 담는다. 결정적으로 **보증이 철회되지 않았다**: 우회 뮤턴트는 여전히 회귀를 잡는다(M-AUD-4 아래 FAIL, route (i) 아래 FAIL). 잡는 층이 테스트 층에서 리졸버 층으로 내려갔을 뿐이다.

### ⑦ §C.1 14행 (지적 7)

정본 표는 `progress.md:75-94`이며 **14행**이다: A1-A6(6) · B1(1) · C1-C7(7). 처분 분해는 보존 이관 5 + 가지 이행 1 + 의도적 갱신 1 + 비임시 사본 6 + 수용된 손실 1 = **14**. (배차문의 「보존 6 + 가지 이행 1 + …」 = 15 는 배차문의 재진술 오류이고, 산출물 자체는 M2 커밋 메시지·progress.md·CHANGELOG 셋 다 14 로 일치한다.)

개별 이행 확인:

```
$ for t in <6개 _NonTemp 사본>; do /usr/bin/grep -rl "func $t(" internal/; done
internal/kanban/todo_root_nontemp_copy_test.go   (C1·C2·C3·C4)
internal/web/todo_section_nontemp_copy_test.go   (C5)
internal/cli/todo_queue_root_test.go             (C6)
```

6건 전부 실재하고 전부 PASS. A1·A2·A3 는 `declareNonTemporary(t)` 를 달고 원 단언 유지(`todo_root_test.go:107,191,258`).

**「수용된 손실」(C7, `cli.TestAxisACanaryHomeSweep_TodoFamily`)을 가장 세게 봤다.** 구성 불가 주장의 근거는 `exec.Command` 자식 프로세스가 패키지 수준 이음매(`TempRootsFn`)를 물려받지 않는다는 것이다. 소스를 읽어 확인했다 — 그 테스트는 `go test ./internal/cli/ -run TestTodo` 를 **자식 프로세스로** 띄우고 canary HOME 아래 오염을 센다. 자식은 별개 프로세스이므로 부모가 세운 Go 변수 스텁이 도달할 길이 원리상 없다. 환경변수로 우회하려면 프로덕션 코드에 테스트 전용 환경변수 판독 경로를 새로 뚫어야 하는데, 그것은 이 카드의 범위 밖 설계 변경이다.

포기가 아니라는 증거도 있다: 그 테스트는 **한 줄도 수정되지 않았고**(카드 범위 diff 에서 함수명 0회 등장 — 같은 파일은 83줄 바뀌었으므로 필터가 죽은 것이 아니다), 자기 안에 빈 스윕 집합 방어(`passes == 0` → `t.Fatalf`)를 이미 갖고 있으며, 기본 실행에서 SKIP 하는 것도 원래 설계다(`MOAI_AXIS_A_CANARY_SWEEP=1` 게이트). 그리고 그것이 지키던 축(우회 시 오염)은 A6 행이 다른 층에서 계속 지킨다.

**판정: 진짜 구성 불가이며, 포기가 아니다.** 다만 「수용된 손실」이 실제로 손실이라는 점은 아래 잔여 위험에 남긴다.

### ⑧ sync_commit_sha (지적 8)

```
$ /usr/bin/grep -rn "sync_commit_sha" .moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/
progress.md:419:sync_commit_sha: pending-backfill       # 이 커밋은 자기 해시를 인용할 수 없다 (D3 백필 창)
progress.md:469:- **`sync_commit_sha` 는 이 커밋에서 `pending-backfill` 이다.** …
```

(감사자 자신의 첫 판독은 `spec.md` 만 훑어 「키 부재」로 보였다. 범위를 SPEC 디렉터 전체로 넓혀 재측정한 결과가 위이며, 첫 판독은 절단이 만든 것이었다.)

**부채는 실재하고 HEAD 에서 미상환이다.** 상환 주체는 **이 레인**이다 — 자리표시자를 쓴 주체가 뒤따르는 커밋에서 실 SHA 로 채운다. 리드가 아니다: 리드는 이 커밋 해시를 알지만 SPEC 산출물의 소유자가 아니고, 병합 후에는 백필 창이 닫힌다. 병합 전에 레인이 갚아야 한다.

### ⑨ docs-site 4개 로케일 (지적 9)

```
$ /usr/bin/grep -rn "moai/todo" docs-site/content/*/utility-commands/moai-todo.md
en/…:215: … Projects without git metadata keep the queue at `~/.moai/todo/<project-key>/backlog.db`.
ja/…:215: … git メタデータのないプロジェクトは `~/.moai/todo/<project-key>/backlog.db` にキューを置きます。
ko/…:215: … git 메타데이터가 없는 프로젝트는 `~/.moai/todo/<project-key>/backlog.db`에 큐를 둡니다.
zh/…:215: … 没有 git 元数据的项目把队列放在 `~/.moai/todo/<project-key>/backlog.db`。
```

4개 로케일 모두 같은 줄에 같은 문장이 있고, 이 가드가 임시 루트 base 에 대해 그것을 **거짓으로 만든다**.

**판정: 편집하지 않은 것 자체는 옳고, 추적 주체를 남기지 않은 것이 결함이다.** 내부 가드를 겨눈 sync 커밋 안에서 4개 로케일 문서를 고치는 것은 범위 규율 위반이 맞다. 그러나 범위 규율은 「이 커밋에서 고치지 않는다」를 방어할 뿐 「아무도 소유하지 않게 둔다」를 방어하지 않는다. 실측: 이 항목은 `followup-candidates.md`(후보 2건)에 **없다**. CHANGELOG 산문 한 줄이 유일한 기록이고, 산문은 큐가 아니다.

---

## Baseline-attribution

이 실행에서, 이 트리(`029ab039f`)를 상대로 재측정한 것:

- AC 8건 재유도 — `go test`(kanban·cli·web), 스윕 계수 포함
- 뮤턴트 4종(M-AUD-1~4) 주입·실행·복원·해시 대조 — 커밋된 뮤턴트 A/C/D/E 를 **인용하지 않고** 감사자가 독립 설계
- route (i) 독립 재현 — 커밋된 `m3/routei-*.txt` 수치와 대조
- D8 양쪽 diff 재생성 + 커밋본과 바이트 대조 + 6파일 커밋 귀속
- 형제 SPEC diff 2범위 + 대조군
- §C.1 14행 개별 확인
- `go vet` 3패키지 · `go test -cover` kanban · `gofmt -l`

carry-over 로 쓴 값은 없다. `~/.moai/todo` 계수 344 는 **재측정하지 않았고**, 따라서 이 판정의 근거로 쓰지 않았다(아래 Gaps).

---

## Gaps — 관측하지 않은 것

1. **linux / windows 셀 전부 미측정.** darwin/arm64 한 대에서만 쟀다. AC-THG-002 의 심링크 등가는 리눅스에서 축퇴 케이스가 되고, 그 칸은 CI 몫이다. 이 레인은 push 하지 않으므로 인용할 CI 결과가 없다.
2. **windows 테스트 컴파일성 미검증.** 카드가 인용한 `GOOS=windows go build` 는 테스트 파일을 컴파일하지 않는다. 감사자도 재실행하지 않았다.
3. **`golangci-lint` 미실행.** 카드도, 감사자도 돌리지 않았다. CI 몫.
4. **`go test ./...` 미실행** — 저장소 규율상 금지. 3개 패키지(`internal/{kanban,cli,web}`)만 봤다. 이 가드를 임포트하는 제4의 소비자가 있다면 이 감사는 보지 못한다.
5. **`find ~/.moai/todo -maxdepth 1 -type d | wc -l` = 344 미재측정.** AC-THG-008 의 이 셀은 감사자가 확인하지 않았다. 카드의 수치를 재측정 없이 재진술하지 않는다.
6. **커버리지는 `internal/kanban` 만 측정**(86.3%). `internal/cli`·`internal/web` 커버리지는 재지 않았다.
7. **뮤턴트 공간은 전수가 아니다.** 4종을 골라 걸었다. 못 걸어본 변형이 더 잡히지 않을 수도 있다 — F1 이 그 부류의 첫 사례다.

---

## Findings (구조화 결함 목록)

- **F1** [Medium] [blocking] `internal/kanban/temp_origin.go:60-62` — **생산 임시 루트 집합의 `/tmp`·`/var/folders` 항목을 지키는 테스트가 0건이다.** `defaultTempRoots` 를 `{os.TempDir()}` 하나로 줄이는 뮤턴트(M-AUD-2)가 `internal/kanban`·`internal/cli` 전 테스트를 **초록으로 통과**한다. 등가 뮤턴트가 아님을 직접 쟀다 — 같은 프로브를 두 판본에서 돌린 결과 `/tmp/t203-probe` 가 `isTemp=true reason="/tmp"` → `isTemp=false reason=""` 로 뒤집힌다(`audit-rootset-probe.txt`). 하필 그 경로가 이 SPEC 이 확보한 **유일한 생산 오염 실측**(`~/.moai/todo/t203-probe-d7a16ea2`)의 기원이다. AC-THG-006 의 생산 루트 서브테스트는 `/tmpfoo` **부정** 방향만 물으므로 이 뮤턴트에도 통과하고, 긍정 방향(멤버십)에는 산출 테스트가 없다. 즉 REQ-THG-002 의 집합 정의 절이 산출 AC 없이 출하됐다 — 이 SPEC 이 네 번(D2·D10·D14·D19) 잡아낸 「공허한 초록」의 다섯 번째 사례이며, 앞의 넷과 달리 **이번엔 출하됐다**. 지금 코드의 동작은 옳으므로 기능 결함이 아니라 회귀 방어 부재다. **Required fix**: 주입 이음매를 쓰지 않고 `defaultTempRoots()` 의 반환 자체를 판정하는 테스트를 추가한다 — 예: `defaultTempRoots()` 가 `/tmp` 와 `/var/folders` 를 **원소로 담는지** 직접 단언하거나, `TempOriginReason("/tmp/<임의>")` 가 `reason == "/tmp"` 로 임시 판정됨을 단언(경로를 만들지 않아도 어휘적 정규화 경로로 성립). 어느 쪽이든 M-AUD-2 를 RED 로 만들 수 있어야 하고, 그 RED 를 실연해 기록한다.
- **F2** [Low] [blocking] `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/progress.md:419` — `sync_commit_sha: pending-backfill` 이 HEAD `029ab039f` 에서 미상환이다. 자리표시자 자체는 D3 백필 창의 승인된 관행이지만 부채는 실재한다. **상환 주체는 이 레인이며 병합 전이다** — 백필 창은 병합과 함께 닫히고, SPEC 산출물은 리드 소관이 아니다. **Required fix**: `029ab039f` 를 실 SHA 로 채우는 후속 커밋 1본을 병합 전에 얹는다.
- **F3** [Low] [optional] `docs-site/content/{en,ja,ko,zh}/utility-commands/moai-todo.md:215` — 4개 로케일이 "git 메타데이터 없는 프로젝트는 `~/.moai/todo/<project-key>/` 에 큐를 둔다"고 적고 있으며, 이 가드가 임시 루트 base 에 대해 그 문장을 거짓으로 만든다. sync 커밋 안에서 고치지 않은 것은 범위 규율상 옳다. 결함은 **추적 주체가 없다는 것**이다 — 실측상 `followup-candidates.md`(후보 2건)에 이 항목이 없고, 기록은 CHANGELOG 산문 한 줄뿐이다. **Required fix**: 카드 요청 1건으로 올리거나 `followup-candidates.md` 에 후보 3으로 추가한다. 문서 자체는 이 카드에서 건드리지 않는다.
- **F4** [Info] [optional] `internal/kanban/temp_origin_test.go:49` — 축퇴 케이스 로그가 `os.Getenv("GOOS")` 로 플랫폼 라벨을 만든다. `GOOS` 는 빌드 시 상수이지 런타임 환경변수가 아니므로 그 자리는 런타임에 빈 문자열이 된다(같은 파일 `:184` 에 `runtimeNote()` 가 이미 있다). 로그 한 줄의 라벨 결함이고 판정에는 영향이 없다. **Required fix**: `os.Getenv("GOOS")` → `runtime.GOOS`, 또는 `runtimeNote()` 단독 사용.
- **F5** [Info] [optional] `internal/web/write_safety_test.go` 가 `gofmt -l` 에 잡힌다. **이 카드 소관이 아니다** — `git log` 귀속상 `b6bd0d011` (카드 t544)이며 develop 흡수로 들어왔다. t536 에 대한 결함이 아니라 통합 트리의 관측으로만 남긴다. **Required fix**: 없음(t544 또는 통합 소관).

blocking 2건(F1·F2)은 둘 다 **회귀 방어와 기록**의 부채이지 동작 결함이 아니다. optional 3건은 재량이며, 그중 F5 는 이 카드에 대한 결함이 아니다.

---

## Residual-risk — 관측했음에도 여전히 틀릴 수 있는 것

1. **뮤턴트 4종은 표본이다.** F1 은 다섯 번째 시도에서 나왔다. 여섯 번째가 더 나올 개연성을 이 감사는 배제하지 못한다 — 특히 `normalizeForTempCompare` 의 Lstat 실패 분기와 `pathWithin` 의 `filepath.Rel` 오류 분기는 뮤턴트를 걸어보지 않았다.
2. **false positive 는 조용하다.** 가드가 잘못 발화하면 실제 프로젝트의 홈 큐가 말없이 회수된다. 그 방향은 `TestTempOrigin_LexicalResembler`(자체 양성 대조군 보유)가 지키지만, 그 테스트가 겨누는 것은 **주입된** 루트 집합이다 — 생산 집합에서의 false positive 는 F1 과 같은 사각에 있다.
3. **route (i) 의 비등가 한 자리.** stderr 안내 줄이 pre-guard 트리에는 없었다. 측정 축에서는 무해하다고 판정했으나, 그 판정 자체가 「PersistentPreRun 이 파일 시스템을 건드리지 않는다」는 소스 판독에 기대고 있다 — 실행으로 반증되지는 않았다.
4. **darwin 한 대.** 리눅스에서 `/tmp` 가 실디렉터일 때 `normalizeForTempCompare` 의 두 앵커(정규화형·어휘형)가 같은 값으로 붕괴하는데, 그 상태에서 포함 판정이 이 기계와 같이 동작하는지는 재지 않았다.
5. **수용된 손실은 실제로 손실이다.** C7 이 지키던 것은 「자식 프로세스 경계를 넘는 실제 sweep」이며, 그 축은 이제 어떤 테스트도 기본 실행에서 지키지 않는다. A6 이 대신 지키는 것은 **같은 프로세스 안의** 우회다. 자식 프로세스 경계에서만 나타나는 회귀는 `MOAI_AXIS_A_CANARY_SWEEP=1` 을 손으로 켜기 전까지 보이지 않는다.
6. **3개 패키지 밖은 보지 않았다.** `ResolveTodoQueueRoot` / `TempOriginRefusal` 의 제4 소비자가 있다면 이 감사의 사각이다.
7. **감사자 자신의 절단 오류가 한 번 있었다.** `sync_commit_sha` 를 `spec.md` 만 훑어 「부재」로 읽었고, 범위를 넓혀 정정했다. 같은 부류의 좁은 판독이 이 보고 안에 더 남아 있을 수 있다.

---

## 판정 근거 요약

must-pass 두 축(Functionality 88 · Security 94)이 독립적으로 임계를 넘었으므로 방화벽은 발화하지 않는다. blocking 2건은 동작 결함이 아니라 회귀 방어(F1)와 기록(F2)의 부채이고, 둘 다 **병합 전에 갚을 수 있는** 모양이다. 그래서 FAIL 이 아니라 **PASS-WITH-DEBT** 다.

이 카드의 검증 규율은 감사 대상 중 상위에 든다 — 테스트마다 전제 가드와 양성 대조군을 달았고, 못 잡은 뮤턴트를 숨기지 않고 보고했으며, D8 을 수리하는 대신 양쪽을 재측정해 보고했다. F1 이 그 규율의 예외가 아니라 **그 규율이 스스로 네 번 이름 붙인 부류의 다섯 번째 사례**라는 점이, 이 결함을 optional 이 아니라 blocking 으로 분류한 이유다.
