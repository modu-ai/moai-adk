# t436 — `moai goal` 산문 조건 조용한 실패 수리

- 카드: t436
- 브랜치: `WT-goal-prose-arm` (워크트리 `.claude/worktrees/t436`)
- 측정 트리: 로컬 `develop` = `b7462203a` + 이 카드의 미커밋 변경
- 빌드 스탬프: `v3.1.2-1396-gb7462203a-dirty`

---

## Claim

산문 조건이 mechanical 로 오분류돼 `sh -c` 로 실행되고, 종료코드 0 이 원리상 불가능해 Stop 훅이 30턴 상한까지 차단하는 결함을 수리했다. 카드가 [HARD] 로 요구한 분리 실험을 먼저 수행해 원인을 특정했고, 그 결과에 따라 "분류만 고치는 안"을 기각하고 세 갈래를 함께 넣었다.

1. **명시 선언 접두사** `model:` / `cmd:` — `parseCondition` 안에서 처리되므로 CLI arm 경로와 MCP `goal_arm` 래퍼가 함께 덮인다.
2. **무장 시점 거부** — mechanical 로 분류된 조건의 첫 낱말이 아무 명령으로도 해석되지 않으면 non-zero 로 거부하고 상태 파일을 쓰지 않는다. 명시 `cmd:` 는 면제된다.
3. **평가 시점 backstop** — mechanical 조건이 exit 127 을 내면(그 조건의 `ExpectExit` 이 127 이 아닌 한) 선언대로 만족 불가로 보고 차단을 멈춘다.

---

## Evidence

### E1 — 분리 실험: t392 정규 조건문이 model 로 분류된 원인은 토큰 단독

`go test ./internal/cli/ -run TestT436Isolation -v -count=1` (실험 당시 scratch, 이후 `TestParseCondition_ReferentTokenIsTheSoleDiscriminator` 로 상설화):

```
A canonical(multiline+frame+token)     -> type=model      cmd=""
B canonical minus token                -> type=mechanical cmd="Every blocking acceptance criterion in .moai/specs/SPEC-XXX/ ..."
C single-line, no frame, token         -> type=model      cmd=""
D multiline+frame, no token            -> type=mechanical cmd="Every AC in acceptance.md has PASS evidence shown in the chat log; ..."
E korean prose                         -> type=mechanical cmd="모든 차단 AC가 통과 증거와 함께 대화에 표시 ..."
F korean prose multiline+frame         -> type=mechanical cmd=".moai/specs/SPEC-XXX/acceptance.md 의 모든 차단 AC 가 통과 증거를 갖는다; ..."
```

A↔B 는 `conversation` → `chat` 한 낱말만 다른데 분류가 갈린다. C 는 다중행도 틀도 없이 model 이고, D·F 는 둘 다 갖추고도 mechanical 이다. **다중행 모양과 ac_converge 틀의 기여도는 0**이며 판별자는 영어 부분문자열 하나뿐이다.

### E2 — 프로덕션 러너의 실제 종료코드

`go test ./internal/cli/ -run TestT436ExecProse -v -count=1` (scratch, `realCmdRunner{}.Run` 직접 호출):

```
korean prose               -> exit=127  err=<nil> out="sh: 모든: command not found\n"
english prose no metachar  -> exit=127  err=<nil> out="sh: Every: command not found\n"
real command               -> exit=0    err=<nil> out=""
```

exit 127 은 언어에 무관한 정확한 신호이며, Piece 3 의 판별 근거다.

### E3 — 수리 전 조용한 실패 (트리 빌드 바이너리)

```
armed goal for session t436-probe-0002 (mechanical condition, ceiling 30 turns): 모든 차단 AC가 통과 증거와 함께 대화에 표시된다
arm exit=0
```

경고 없음, exit 0.

### E4 — 수리 후 거부 (`./bin/moai`, 재빌드본)

```
ERROR
  Goal arm: "모든 차단 AC가 통과 증거와 함께 대화에 표시된다" was classified as a mechanical condition and would be
  run as a shell command, but its first word "모든" resolves to no command — the condition can never exit 0, so the
  goal would block every turn-end until the ceiling. ...
prose arm exit=1
```

### E5 — `cmd:` 면제 (`./bin/moai`)

```
armed goal for session t436-e2e-cmd (mechanical condition, ceiling 30 turns): cmd: t436-nonexistent-tool --check
cmd-prefix arm exit=0
```

### E6 — 테스트

| 명령 | 결과 |
|---|---|
| `go test ./internal/cli/ -run 'ParseCondition\|GoalArm\|DeclaredMechanical' -count=1` | `ok ... 0.857s` (하위 17 PASS, 0 FAIL) |
| `go test ./internal/cli/... -count=1` | 전 패키지 `ok` — `grep -Ev '^ok\|no test files'` 출력 공집합 |
| `go test ./internal/template/... ./internal/goal/... -count=1` | `internal/template 32.459s ok` · `agentemit ok` · `internal/goal ok` |
| `go vet ./internal/cli/... ./internal/goal/...` | rc=0 |
| `gofmt -l` (변경 8파일) | 공집합 |
| `make build` | 성공, `catalog.yaml` 바이트 불변 |
| 템플릿 중립성 grep (추가 줄 대상: SPEC/REQ/AC 토큰·날짜·SHA·카드 id·`CLAUDE.local`·절대경로) | 히트 0 |

---

## Baseline-attribution

모든 측정은 이 실행에서, 워크트리 `.claude/worktrees/t436`, 브랜치 `WT-goal-prose-arm`, 트리 = `b7462203a` + 이 카드의 변경분에 대해 수행했다. E1·E2·E3 은 수리 **전** 트리(= `b7462203a` 무변경)에서, E4·E5·E6 은 수리 **후** 같은 트리에서 측정했다. `internal/goal/evaluate.go` 의 gofmt 불일치는 `git show HEAD:internal/goal/evaluate.go` 를 별도 파일로 뽑아 `gofmt -l` 에 통과시켜 **HEAD 시점 선재**임을 확인했다 — 이 카드의 회귀가 아니다.

---

## Gaps

명시적으로 관측하지 **않은** 것:

- `golangci-lint` 미실행. `go vet` + `gofmt` 만 돌렸다.
- 커버리지 미측정. `-cover` 를 돌리지 않았으므로 커버리지 증감에 대해 아무 주장도 하지 않는다.
- `go test ./...` 미실행(프로젝트 규율). 전 패키지 판정은 CI 몫이다.
- MCP `handleGoalArm` 경로에 테스트가 없다. 게이트는 CLI 경로와 같은 두 함수(`parseCondition`·`unrunnableCommandToken`·`declaredMechanical`)를 호출하지만, 그것은 논증이지 측정이 아니다.
- Windows 미검증. `commandTokenResolves` 는 `sh` 를 띄운다. 설계상 spawn 실패는 fail-open(=허용)이므로 `sh` 가 없는 호스트에서 거부가 나지는 않으나, `GOOS=windows` 빌드도 실행도 하지 않았다.
- `run.md` 의 정규 `ac_converge` 조건문은 손대지 않았다. 이미 `conversation` 을 포함해 model 로 분류되므로 깨진 것은 없으나, **시스템에서 가장 많이 무장되는 조건문이 여전히 접두사가 아니라 그 낱말에 의존한다.**

---

## Residual-risk

- **거부는 첫 낱말 휴리스틱이라 놓친다.** `go home and rest` 는 `go` 가 해석되므로 무장을 통과하고 평가 시점에야 걸린다. Piece 3 가 정확히 그 경우를 위해 있지만, 사용자는 문 앞이 아니라 첫 턴 끝에서 알게 된다.
- **exit 127 은 의미가 겹친다.** 환경에 진짜로 없는 정상 명령도 "선언대로 만족 불가"로 중단된다. `cmd:` 면제가 무장 시점의 같은 문제는 풀지만 평가 시점은 풀지 않는다.
- **Piece 3 는 기존 종료 경로에 네 번째 모양을 추가한다.** `stop-goal` stdout JSON 을 `ceiling_exit` / `stagnation` / `yielded` 로 전수 분기하던 소비자가 있으면 새 `unsatisfiable` 를 만난다. 훅 본체와 대시보드 렌더러 외의 소비자는 찾지 못했으나 셸 래퍼를 전수 감사하지는 않았다.
- **백틱 위험 — 수리가 어디까지 막는가 (측정, 추정 아님)**

  정규 `ac_converge` 조건문은 `` `go test ./...` `` 를 백틱으로 담는다. mechanical 로 오분류되면 `sh -c` 명령치환으로 **전체 테스트 스위트가 실제 실행된다.** 오분류의 결과는 "영원히 실패" 가 아니라 **의도치 않은 명령 실행**이며, 결함의 종류가 다르다.

  리드 요청에 따라 수리의 도달 범위를 실측했다 — `go test ./internal/cli/ -run Backtick -count=1 -v`, 3케이스 전부 PASS (`internal/cli/goal_backtick_hazard_test.go`). 무장 경로만 실행하며, 모든 백틱 본문은 `true` 로 두어 예기치 않은 실행도 무해하게 했다.

  | 케이스 | 관측 | 실행되는가 |
  |---|---|---|
  | 정규 조건문 그대로 | `conversation` 때문에 model 로 분류, `cmd` 필드 공백 | **아니다** — 셸에 도달하지 않는다 |
  | 정규 조건문에서 referent 제거 (위험 형태) | mechanical → 첫 낱말 `Every` 미해석 → **무장 거부** | **아니다** — 명령치환 전에 막힌다 |
  | 첫 낱말이 해석됨 (`` true `true` ``) | 게이트 통과, 무장됨 | **그렇다** (평가 시점) |
  | 백틱이 첫 글자 (`` `true` && true ``) | 게이트가 판정 불가 형태로 보고 건너뜀 → 무장됨 | **그렇다** (평가 시점) |

  즉 **카드가 보고한 위험 형태는 막힌다** — 백틱을 담은 산문은 첫 낱말이 산문 단어이므로 무장 시점에 거부된다. 막히지 않는 것은 첫 낱말이 실제 명령인 경우이고, 이는 게이트가 첫 낱말만 판정하기 때문이다(양성 증거만으로 거부한다는 설계의 대가 — 문자열 나머지를 추론하려는 게이트는 정상 명령을 거부하게 되고 그쪽이 더 나쁜 오류다). exit 127 backstop 도 여기서는 돕지 않는다: 그 명령들은 해석되고 실행된다. 남는 노출을 테스트로 못박아 두었으므로(`TestBacktickHazard_KnownGaps`) 이 갭이 닫히면 테스트가 알려준다.

---

## 후속 제안 (이 카드 범위 밖)

`run.md` 의 정규 `ac_converge` 조건문에 `model:` 접두사를 붙여 낱말 의존을 끊는 것. 이는 발견성 편집이 아니라 **무장되는 조건 자체의 변경**이고, `TestRunmdAcConvergeProseMatchesFixture` 와 `TestAC009_*` 가 그 텍스트를 고정하고 있어 별도 카드가 맞다.

---

## 병합 준비 부기 — catalog 해시 재생성이 흡수한 남의 누락분

이 카드의 `internal/template/catalog.yaml` 한 줄(`moai` 엔트리 해시)은 **두 원인을 함께 덮는다.** 착지 후 이 줄이 t436 단독 산물로 읽히면 t191 이 catalog 를 빠뜨렸다는 사실이 사라지므로 여기 남긴다.

| # | 원인 | 근거 |
|---|---|---|
| 1 | 이 카드가 `.claude/skills/moai/workflows/goal.md` 를 수정 | `git diff --stat develop HEAD -- .claude/skills/moai/` → 1 file, +34/-1 |
| 2 | **develop 자체의 미재생성분** — t191 `aed638b8b` 가 `workflows/project.md` 와 `project/doc-generation.md` 를 얹으며 `catalog.yaml` 미접촉 | `git log --oneline <base>..develop -- internal/template/catalog.yaml` → 병합 커밋 `d77adf6b2` 1건뿐(재생성 아님). develop 저장값 `3fff7dba…` 가 갱신되지 않은 채 남아 있음 |

두 원인은 같은 엔트리를 가리키므로 한 번의 재생성이 둘을 함께 해소한다:

```
go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai
go test ./internal/template/ -run 'TestCatalogHashCoversSkillSubfiles' -count=1   # ok
```

`--entry moai` 로 충분함을 표적 테스트로 확인했다(위 명령 rc=0). 실패 메시지가 안내하는 `--all` 은 쓰지 않았다 — `--all` 은 `sync-auditor` 엔트리까지 재생성해 **t443 소관의 드리프트를 이 카드가 조용히 덮어버리기** 때문이다. 소관을 흐리지 않는 쪽을 골랐다.
