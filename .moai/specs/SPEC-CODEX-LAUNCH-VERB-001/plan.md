# SPEC-CODEX-LAUNCH-VERB-001 — 구현 계획

> 순서는 **되돌리기 어려운 결정 먼저**다. §B 의 열린 결정과 §C 의 설계 판단이 앞에 오고, 기계적 단계는 §F 뒤쪽으로 밀었다. 검토자가 읽어야 할 곳은 §B 다.

## §A. 맥락

카드 t391 은 SPEC-CODEX-LAUNCHER-001 `plan.md` §B 의 (b) 판정 — "맨몸은 리드아웃, 기동은 명시 동사" — 를 뒤집는다. 뒤집는 근거는 판독본 C2 다: (b)를 택한 이유였던 "실수로 현재 세션이 codex 에 넘어갈 위험"이 현행 구현에서 무게가 다르다. cc/glm 은 `syscall.Exec` 로 프로세스를 교체해 돌아올 곳이 없지만, codex 는 0.8.0 개정 이후 `os/exec` 자식이라 자식이 끝나면 셸로 돌아온다.

승계 관계·측정 전제·범위 밖은 `spec.md` §A/§B/§C 가 가진다. 여기서는 반복하지 않는다.

## §B. 전제와 열린 결정

### B.1 검증하지 못한 전제 (그대로 넘긴다)

| # | 전제 | 상태 |
|---|---|---|
| P1 | 설치본 `~/go/bin/moai` 의 `--help` 문자열. 카드가 인용한 v3.1.2 문구는 **리드의 측정이며 이 세션의 것이 아니다**. 이 트리에서 재측정하지 않았다 | 미측정 (판독본 G1) |
| P2 | 이 트리에서 `go build` · `go test` **0건**. 인용한 모든 코드 성질은 소스 판독이며 실행 관측이 아니다 | 미측정 (판독본 G2) |
| P3 | `codex app` 경로의 실제 기동 거동. codex 자식을 실제로 띄워 보지 않았다 | 미측정 (판독본 G5) |
| P4 | codex 의 서브커맨드 표면은 **최상위 `codex --help` 만** 훑었다. 서브커맨드별 플래그와 숨은(hidden) 플래그는 훑지 않았다 — B7·B8 의 부재 주장은 그만큼 좁다 | 부분 측정 |
| P5 | tty 왕복은 CI 에서 관측 불가하므로 단언하지 않는다 (0.8.0 이 남긴 Gap 을 그대로 승계) | 구조적 미관측 |

**P2 의 무게를 낮춰 읽지 말 것.** 이 계획의 모든 파일 좌표(`file:line`)는 트리 `e79272713` 기준이며, run 진입 시점에 트리가 움직였다면 다시 재야 한다.

### B.2 REQ-CL-013 교차 SPEC 훑기 — 결과

판독본 G3 이 요구한 훑기를 수행했다.

```
grep -rln "REQ-CL-013" .moai/specs/
```

→ 4건, **전부 SPEC-CODEX-LAUNCHER-001 자기 산출물 안**이다: `spec.md`(정의), `plan.md`, `acceptance.md`, `progress.md`. Codex 계열 다른 8개 SPEC(`SPEC-CODEX-{DUAL-AGENTS,HOOK-ADAPTER,INIT,PHASE2,SESSION-MSG,SKILLS-CANONICAL,VERDICT-SYNTH,WIRING}-001`) 어느 것도 REQ-CL-013 을 수용 칸에서 참조하지 않는다.

**따라서**: REQ-CL-013 은 다른 SPEC 의 AC 를 통해 간접적으로 걸려 있지 않다. 이 카드가 그것을 유지하기로 한 판단(§C 범위 밖)은 SPEC-CODEX-LAUNCHER-001 하나만 상대하면 된다. 이 주장은 **SPEC ID 토큰 `REQ-CL-013` 의 문자열 부재**에 대한 것이며, 같은 규율을 다른 표현(예: "런처는 쓰지 않는다")으로 재기술한 AC 가 있다면 이 훑기는 그것을 잡지 못한다.

### B.3 열린 결정 — `-w` 의 의미론

**[NEEDS CLARIFICATION: `-w` 를 codex 자식에게 전달할 것인가, moai 가 소화할 것인가]**

`moai cc` 의 `-w` 는 **claude 에게 넘긴다** — `normalizeWorktreeFlag` 가 `--worktree <name>` 정규형으로 고쳐 argv 에 실어 보내고, 워크트리 생성·진입은 claude 가 한다(`cc.go:220-228`). codex 에는 대응물이 없다: 최상위 `codex --help` 를 `worktree` 로 grep 하면 0행이고 `-w,` 단축 플래그도 없다(B8; P4 의 범위 제한이 걸린다).

두 안:

- **(가) moai 가 소화한다 (권장)** — `-w` 를 argv 에서 **떼어내고**, 해석된 워크트리 루트를 `codexLaunchRequest.Dir` 로 쓴다. codex 는 그 디렉터리에서 시작한다. `cc` 와 사용자 표면은 같고(`-w <name>` 을 쓰면 그 트리에서 뜬다) 내부 기제만 다르다. REQ-CLV-007 의 "working directory" 표현이 이 안을 이미 가정한다.
- **(나) 자식에게 전달한다** — codex 가 모르는 플래그를 받게 되고, B7 에서 본 것과 같은 형태의 결함(모르는 토큰을 프롬프트로 읽음)을 재생산한다. **기각 권고.**

권고는 (가)다. 다만 이것은 사용자 표면의 의미를 정하는 결정이라 **Implementation Kickoff Approval 전에 운영자 판정이 필요하다.** (가)를 택하면 파생 결과 하나: `-w` 는 워크트리를 **생성하지 않는다**(cc 는 claude 를 통해 생성한다). 존재하지 않는 트리를 가리키면 진단하고 기동하지 않는 편이 조용히 프로젝트 루트로 떨어지는 것보다 낫다 — AC 로 단언한다.

### B.4 열린 결정 — 기본 동사의 argv 모델

**[NEEDS CLARIFICATION: 라우팅 표에서 맨몸 토큰 `""` 를 어느 클래스에 넣을 것인가]**

`codexVerbRouting` 은 문자열→클래스 표이고, **표에 없음이 곧 거절**이다(REQ-CLV-003 이 지키려는 성질). 맨몸 기동은 `"" → codexVerbLaunchCli` 로 옮기면 성립한다. 문제는 그 다음이다: 현행 `runCodexLaunch` 는 `Args: append([]string{verb}, tail...)` 로 **사용자가 친 원문 토큰**을 자식에게 싣는다. 맨몸이면 그 토큰이 빈 문자열 `""` 이라 codex 가 빈 인자를 받는다.

B7 이 이 문제를 한 단계 키운다: `cli` 토큰 자체가 codex 에 없는 서브커맨드라, 지금도 `moai codex cli` 는 codex 에게 **프롬프트 `cli`** 를 넘기고 있다. 따라서 두 가지를 함께 고쳐야 한다 — 맨몸의 빈 토큰과 `cli` 의 가짜 서브커맨드 토큰. 권고는 **동사 클래스가 자식 argv 를 결정하게** 하는 것이다(REQ-CLV-004): 기동 클래스는 tail 만 넘기고, `app` 클래스만 `app` 토큰을 앞에 붙인다.

**이 변경은 기존 시험 `codex_launcher_test.go:465`(`want := append([]string{"cli"}, tail...)`)의 기대값을 뒤집는다.** 물려받은 시험이 지금 고정하고 있는 것이 결함이라는 판단이므로, run 단계에서 그 셀을 고칠 때 **왜 기대값이 바뀌는지**를 커밋에 남긴다. 기대값을 조용히 맞추는 것은 이 카드가 가장 경계할 실패 형태다.

### B.5 승계 포인터의 소관

SPEC-CODEX-LAUNCHER-001 은 `completed` 다. 그 파일에 HISTORY 항목이나 `superseded_by` / `partially_superseded_by` 를 다는 일은 **sync-phase 소관이며 run-phase 편집 대상이 아니다.** run 단계에서 그 파일을 건드리면 완결 SPEC 의 본문을 고치는 것이 되고, 상태 전이 소관 규율에도 어긋난다. 이 계획은 그 포인터를 **부착 대상 목록으로만** 남긴다:

- `SPEC-CODEX-LAUNCHER-001/spec.md` frontmatter — `partially_superseded_by: [SPEC-CODEX-LAUNCH-VERB-001]` (REQ-CL-002 하나만 대체되므로 partial)
- 같은 파일 HISTORY — 대체 사실 1행
- REQ-CL-002 본문 옆 승계 주석 (편집 여부 자체가 sync 판단)

## §C. 설계 판단

### C.1 seam 은 이미 있다 — 새로 만들지 않는다

| 작업 항목 | seam | 성질 |
|---|---|---|
| 기본 동사 역전 | `codexVerbRouting` (`codex_launcher.go:62-67`) | 표에 없음 = 거절. **이 성질을 보존한다** |
| 자식 argv | `runCodexLaunch` 의 `codexLaunchRequest.Args` 조립 (`:246`) | 동사 클래스가 argv 를 결정하도록 뒤집는다 |
| CODEX_HOME | `codexDirectLaunch` 의 `exec.Cmd` (`:265-`) | 현재 `Env` 미지정(B4). `resolveCodexHomeDir` 가 이미 해석값을 안다 |
| `-w` | `resolveWorktreeL2Path` + `normalizeWorktreeFlag` (`launcher.go:866,934`) | **재사용**한다. codex 용 두 번째 해석기를 만들지 않는다 |
| 게이트 | `codexInitOfferGate` (`codex_init.go:152`) | 단일 통과 지점(B3). **손대지 않는다** — 상속은 확인 대상이지 설계 대상이 아니다 |

### C.2 게이트 상속은 확인해야 성립한다

B3 은 **호출 그래프**로 확인한 것이지 맨몸을 `cli` 로 옮긴 뒤의 거동을 관측한 것이 아니다(판독본 §5 가 이것을 잔여 위험으로 남겼다). 따라서 상속은 산문이 아니라 AC 로 단언한다 — 맨몸 + 미배선 + 비대화형 = 프롬프트 0, 기동 0 (AC-CLV-009a/b).

### C.3 `Env` 전달 형태

`c.Env = append(os.Environ(), "CODEX_HOME="+resolved)` 형태를 쓰면 부모 환경을 유지하면서 해석값을 덮어쓴다(뒤 항목이 이긴다). 이것이 REQ-CLV-005 의 "나머지 환경은 그대로"를 만족하는 가장 싼 형태다. **파일을 쓰지 않으므로 REQ-CL-013 과 충돌하지 않는다** — 이 구분이 REQ-CLV-006 이 명시적으로 존재하는 이유다.

`--spawn` 경로는 tmux 새 창을 여는 다른 경로다(`codexSpawnLaunchFn`). 환경 전달이 두 경로에서 같은 값을 갖는지는 AC 로 판정한다.

## §D. 제약

- **크로스 플랫폼**: `os/exec` 자식 + 종료코드 전파 유지, OS 빌드 태그 0건, codex 기동 경로에 `syscall` import 0건 (REQ-CLV-010). 이 성질은 0.8.0 이 감사 지적 D3 로 얻은 것이라 회귀시키면 같은 모순이 되돌아온다.
- **템플릿 중립성**: help 문안에 SPEC ID·내부 식별자·내부 날짜를 넣지 않는다.
- **비변경 규율**: 어느 동사도 파일을 쓰지 않는다 (REQ-CLV-006).
- **범위**: 4항 밖으로 나가지 않는다. settings 대칭성과 `-k`/`-f` 는 out of scope (`spec.md` §C).

## §E. 자기 검증

run 단계가 통과를 주장하려면 아래를 **실행하고 그 출력을 인용**해야 한다. 이 계획은 그 명령들을 지정할 뿐 결과를 예단하지 않는다 (P2).

| # | 검증 | 명령 |
|---|---|---|
| E1 | AC 전수 판정 표 | `acceptance.md` 의 AC-CLV-001..014 각 행 |
| E2 | 대상 패키지 시험 | `go test ./internal/cli/...` |
| E3 | 크로스 플랫폼 컴파일 | `GOOS=windows go vet ./internal/cli/...` |
| E4 | 빌드 태그·syscall 부재 | codex 기동 경로 파일에 대한 grep, base 히트 0 확인 포함 |
| E5 | 린트 | `golangci-lint run ./internal/cli/...` |
| E6 | 템플릿 중립성 | 템플릿 변경이 있을 때만 — 중립성 가드 |

**E4 는 공허해질 수 있는 형태다.** grep 이 0행을 내는 것이 "없음"의 근거이려면, 같은 패턴이 **잡아야 할 것을 실제로 잡는지**를 먼저 보여야 한다. base 히트 0인 패턴은 무엇을 심어도 0을 낸다.

## §F. 마일스톤

되돌리기 어려운 것 먼저.

### M1 — 동사·argv 모델 (사용자 표면이 여기서 정해진다)

- `codexVerbRouting` 에서 맨몸 토큰의 클래스를 기동으로 옮기고, `status` 를 리드아웃 별칭으로 보존
- 자식 argv 를 동사 클래스가 결정하도록 뒤집는다 — 기동 클래스는 tail 만, `app` 만 `app` 토큰 부착 (B4 결정 반영)
- 표에 없음 = 거절 성질 보존 확인
- 대응 AC: AC-CLV-001..005

### M2 — `-w` 경로 (B.3 판정 후에만 착수)

- `-w` 값 해석을 `resolveWorktreeL2Path` + 단축명 정규화에 재사용
- 해석 결과를 `codexLaunchRequest.Dir` 로
- 접두사 밖 절대경로 = 진단 + 기동 0
- 대응 AC: AC-CLV-010..012

### M3 — CODEX_HOME 명시 전달

- `codexDirectLaunch` 의 `exec.Cmd` 에 `Env` 지정, 나머지 환경 보존
- `--spawn` 경로의 환경 동등성 판정
- 대응 AC: AC-CLV-006..007

### M4 — 게이트·플랫폼 상속 확인 (기계적, 새 코드 없음)

- 맨몸 + 미배선 + 비대화형 = 프롬프트 0 · 기동 0 을 시험으로 단언
- 빌드 태그 0 · `syscall` 0 · 종료코드 전파 회귀 없음
- 대응 AC: AC-CLV-008..009, AC-CLV-013

### M5 — 세 론처 대조 시험과 help 문안 (기계적)

- 기존 교차 시험 위치에 "맨몸 호출이 기동으로 이어지는가" 셀 추가. **새 교차 시험 파일을 만들지 않는다** — B5 가 기존 파일의 존재를 확인했고, 없는 것은 파일이 아니라 축이다
- `codexCmd` 의 `Use`/`Long`/`Example` 을 역전된 기본값에 맞춰 갱신
- 대응 AC: AC-CLV-014, AC-CLV-004(help 표면 부분)

## §G. 안티패턴

- **기대값 조용히 맞추기.** `codex_launcher_test.go:465` 의 기대값은 결함을 고정하고 있다(B7). 바꿀 때 이유를 커밋에 남기지 않으면 다음 독자는 회귀로 읽는다.
- **"교차 시험이 없다"고 다시 말하기.** 있다(B5). 없는 것은 그 파일이 다루지 않는 축이다. 판독본 C6 자체가 이전 세션의 오보를 정정한 항목이다.
- **settings 대칭성을 되살리기.** REQ-CL-013 과 충돌하며 운영자가 요구에서 제외했다.
- **완결 SPEC 을 run 단계에서 편집하기.** 승계 포인터는 sync 소관(B.5).
- **`-w` 를 codex 에 그대로 넘기기.** B7 이 보여준 결함의 재생산(B.3 (나)안).
- **부재를 grep 0행으로 주장하기.** 패턴이 무엇도 잡지 못하면 0행은 아무 말도 하지 않는다(§E E4).

## §H. 상호 참조

- `.moai/reports/t391/verdict.md` — 이 카드의 측정 기반 (트리 `e79272713`)
- `.moai/specs/SPEC-CODEX-LAUNCHER-001/{spec,plan,acceptance}.md` — 승계 대상과 유지되는 13 REQ
- `.moai/specs/SPEC-CODEX-INIT-001/` — 상속하는 게이트의 소유 SPEC
- `internal/cli/{codex_launcher.go,codex_init.go,cc.go,launcher.go,launch_exec_posix.go}`
- `internal/cli/codex_launcher_test.go` — 확장 대상 교차 시험
