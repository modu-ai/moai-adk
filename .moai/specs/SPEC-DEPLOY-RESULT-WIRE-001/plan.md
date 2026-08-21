# plan — SPEC-DEPLOY-RESULT-WIRE-001

> 마일스톤은 **되돌리기 어려운 결정부터** 배치한다. M1(통지 형태·스트림·수집 정책)이 확정되기 전에는 M2·M3 의 배선 코드가 무엇을 실어 나를지 정해지지 않는다.

## §A. 맥락

선행 SPEC 이 `internal/template` 에 반환 seam(`ResultDeployer` / `DeployResult`)을 만들었으나 **프로덕션 소비자가 0건**이다(sync-audit F2). 이 카드는 소비자를 만든다. 배포기 쪽은 손대지 않는다.

Tier **M** 근거: 예상 변경 파일 6-8개(통지 헬퍼 1 + 호출부 3 + 테스트 2-3 + `CHANGELOG.md`), 예상 LOC 150-350. 파일 수가 Tier S 상한(5)과 M 대역 하단의 경계에 걸치며, 세 호출부가 서로 다른 패키지(`internal/cli`, `internal/core/project`)에 흩어져 있어 판정 팔이 3개로 갈라진다 — 판정을 spec 본문에 인라인하기보다 `acceptance.md` 로 분리하는 편이 읽힌다. 그래서 M.

## §B. 알려진 이슈 / 위험

- **R1 — update template sync 호출부에 주입 seam 이 없다.** `update_template_sync.go:130` 은 `template.NewDeployerWithRendererAndForceUpdate` 를 인라인으로 만든다(반면 clean-reinstall 은 `opts.Deployer` 로 주입 가능, init 은 `project.NewInitializer(deployer, …)` 로 주입 가능). AC-DRW-007 의 update-sync 팔을 세우려면 이 파일에 이 저장소의 기존 seam 관용(`userHomeDirFn` 형태의 패키지 수준 함수 변수)을 하나 도입해야 한다. **테스트를 위해 프로덕션 구조를 바꾸는 것이 아니라, 이 저장소가 이미 쓰는 주입 형태를 같은 파일에 적용하는 것**이다.
  **이 seam 자체가 위험이다**: 기본값이 프로덕션 배포기가 아니거나 `ResultDeployer` 를 만족하지 않으면, 나머지 AC 는 전부 이중체 위에서 초록인 채로 **프로덕션은 영원히 침묵**한다. `REQ-DRW-010` / `AC-DRW-009` 가 이 구멍을 막는다(감사 D5).
- **R2 — `withSymlinkFunc` / `withMirrorCopyFunc` 는 비공개다.** `internal/cli` 테스트에서 쓸 수 없다. 그래서 CLI 쪽 판정은 `ResultDeployer` 를 구현하고 `DeployResult{SkillMirrors: …}` 를 돌려주는 **이중체**로 구성한다 — `DeployResult` · `SkillMirrorEntry` · `MirrorMode*` 상수가 모두 공개이므로 가능하다. 비공개 seam 을 공개로 승격하지 않는다(선행 SPEC 의 표면을 넓히는 변경이 되고, 이 카드 범위 밖이다).
- **R3 — stdout 위험 호출부는 둘이다** (iter-1 은 하나로 적었다 — 감사 D1). update template sync 의 `out` 은 `cmd.OutOrStdout()` 이고, clean-reinstall 의 `opts.Out` 도 **프로덕션 호출부 두 곳이 stdout 을 주입한다**(`update.go:425`·`:627` → `:154`). 어느 쪽이든 통지를 현재 writer 에 실으면 `internal/cli/CLAUDE.md:14` 위반이자 AC-DRW-003 의 2번 팔이 붉어진다. **두 곳 모두** 별도의 stderr writer 를 잡아야 한다.
- **R4 (은퇴) — "clean-reinstall 은 기본값이 stderr 이므로 기본 경로에서 판정한다".** iter-1 의 이 근거는 **틀렸다**: 그 기본 분기는 프로덕션에서 실행되지 않는다(R3). 기본값 선언은 호출부가 무엇을 주입하는지 말해 주지 않으며, 존재하지 않는 구성에서의 초록은 제품에 대해 아무것도 증명하지 않는다. **판정은 `Out` 이 stdout 으로 주입된 프로덕션 배선에서 한다**(AC-DRW-003 [HARD]).
- **R5 — 2회차 이후 침묵.** §A.5/§D 의 잔존. 이 카드가 닫지 않는다는 사실이 착지 후 "폴백이 통지된다" 로 과대 인용되지 않도록, `CHANGELOG` 문구를 M5 에서 **일어난 실행에 한정해** 쓴다.

## §C. 사전 점검 (착수 전)

1. `git log --oneline -1` — 워크트리가 `release/v3.1.3` 병합 상태인지 확인(선행 seam 존재 전제).
2. `grep -n "ResultDeployer" internal/ -r` — 프로덕션 소비자가 여전히 0건인지 재측정(착수 시점 기준선).
3. `grep -rn "deployer.Deploy(\|i.deployer.Deploy(" internal/ --include=*.go | grep -v _test` — 호출부가 여전히 3곳인지 재측정(§A.2 의 값이 스테일이 아닌지).
4. `sed -n '95,100p' internal/core/project/initializer.go` — `InitResult.Warnings` 필드 존재 재확인(§A.3 전제).
5. `grep -n "Out:" internal/cli/update.go` — clean-reinstall 옵션에 주입되는 writer 재측정. **기본값 선언이 아니라 주입값을 본다**(§A.4 [HARD]). 착수 시점에 stdout 이 아닌 값이 주입돼 있으면 §A.4·M3 을 다시 쓴다.

## §D. 제약

- `Deploy` 시그니처 무변경, `ResultDeployer` 필수화 금지(REQ-DRW-006).
- `internal/template` 무변경 — 이 카드는 소비자만 만든다.
- 경고는 stderr(REQ-DRW-004).
- 로컬 전체 스위트 금지. 변경 패키지 대상으로만 실행하고 전 패키지 판정은 CI 에 맡긴다(CLAUDE.local.md §4/§6).

## §E. 자가 검증

- AC-DRW-001..**009** PASS/FAIL 매트릭스(각 행에 실행 명령과 관측 출력). **AC-DRW-009 를 빼지 않는다** — 프로덕션 배선을 보는 유일한 판정이므로, 자가 검증에서 누락되면 나머지 8건이 전부 이중체 위에서만 초록일 수 있다.
- `go test ./internal/cli/... ./internal/core/project/...`, `go vet` 동일 범위, `golangci-lint run`.
- `GOOS=windows go vet` 동일 범위 — **컴파일만 증명**한다고 명시해 기록.
- 선행 `AC-CSC-010` 재실행 결과.

## §F. 마일스톤

### M1 — 통지 형태·스트림·수집 정책 확정 (Priority High)

가장 되돌리기 어려운 결정 셋을 먼저 못박고 **순수 함수 하나**로 구현한다.

- 입력 `*template.DeployResult` → 출력 통지 문자열(들). 폴백 0건이면 빈 결과.
- `MirrorModeCopy` 개수를 담은 요약 1줄 + `MirrorModeFailed` **개수 요약 1줄 + 예시 경고 최대 3건**. `MirrorModeSkipped` 는 **버린다**(REQ-DRW-009).
- [HARD] `failed` 를 항목별 전량 출력하지 않는다. iter-1 은 그렇게 설계해 `REQ-DRW-007`(비비례성)과 모순이었다(감사 D2). 상한은 복사 팔과 `failed` 팔에 **같이** 걸린다.
- 함수는 writer 를 받지 않고 문자열을 돌려준다 — init 경로가 writer 가 아니라 `[]string`(=`InitResult.Warnings`)을 필요로 하기 때문이다. 세 호출부가 같은 함수를 쓰려면 반환형이 문자열이어야 한다.
- 닫힘 조건: AC-DRW-001(개수 단언 포함) · AC-DRW-002(위증 팔 포함) · AC-DRW-004(**세 팔** — 비비례 2건 + 상한 1건, `failed` 비비례 팔의 낮은 표본은 **4**) · AC-DRW-006 · AC-DRW-008 이 이 함수 단위 테스트로 PASS.

### M2 — `moai init` 경로 배선 (Priority High)

- `internal/core/project/initializer.go` `deployTemplates` 에서 `i.deployer` 를 `template.ResultDeployer` 로 승격 시도. 성공하면 `DeployWithResult` 호출, 실패하면 기존 `Deploy` 그대로.
- 얻은 결과를 M1 함수에 넣어 나온 문자열을 `result.Warnings` 에 append. 표시부(`internal/cli/init.go:706`)는 **무변경**.
- 닫힘 조건: AC-DRW-005 · AC-DRW-007 init 팔.

### M3 — 두 update 경로 배선 (Priority High)

**[HARD] 두 경로 모두 통지를 `out` 에 싣지 않는다.** iter-1 의 M3 은 clean-reinstall 통지를 `out` 에 실으라고 지시했고, 그 `out` 은 프로덕션에서 stdout 이므로 **명세대로 구현하면 매 실행 `REQ-DRW-004` 를 위반**했다(감사 D1).

- `update_clean_install.go:439` — `opts.Deployer` 승격 후, M1 문자열을 **새 `CleanReinstallOptions.ErrOut io.Writer`** 에 출력한다. 기존 `Out`(진행/진단)의 의미는 바꾸지 않는다. 프로덕션 호출부 두 곳(`update.go:418` / `:624` 의 옵션 리터럴)에서 `ErrOut: cmd.ErrOrStderr()` 를 주입한다 — **주입을 빼면 AC-DRW-003 1번 팔이 붉어진다.** `ErrOut` 이 nil 인 경우는 `os.Stderr` 로 떨어뜨리되 그 기본값을 판정 근거로 쓰지 않는다.
- `update_template_sync.go:323` — 승격 후 **`cmd.ErrOrStderr()`** 에 출력. 판정을 위해 R1 의 배포기 주입 seam 도 이 마일스톤에서 도입하며, 그 기본값이 프로덕션 배포기이자 `ResultDeployer` 를 만족함을 함께 고정한다.
- 닫힘 조건: AC-DRW-003(**두 호출부, 프로덕션 배선**) · AC-DRW-007 나머지 두 팔 · **AC-DRW-009 세 단언 전부**(3번 단언이 두 옵션 리터럴의 `ErrOut` 주입을 본다 — 필드만 추가하고 주입을 빼면 여기서 붉어진다).

### M4 — 회귀 확인 (Priority Medium)

- 선행 `AC-CSC-010`(seam 토글 불변식) 재실행, `internal/template` diff 0 확인.
- 닫힘 조건: §E 자가 검증 전항 초록.

### M5 — `CHANGELOG.md` 갱신 (Priority Low)

- `[Unreleased]` 의 "The fallback warning does not currently reach you" 를 갱신한다. 새 문구는 **폴백이 일어난 실행**에 한정해야 하며(R5), 2회차 이후 침묵은 승계 카드 소관임을 남긴다.
- sync-phase 산출물이므로 run-phase 판정에는 넣지 않는다.

## §G. 안티패턴

- **AP-1** — `Warnings()` 를 통째로 forward 하는 것. `skipped` 오귀속 문구가 그대로 나가 REQ-DRW-009 를 깬다.
- **AP-2** — 스킬 이름을 열거하는 통지. 34줄이 되어 REQ-DRW-007 을 깬다.
- **AP-3** — update template sync 의 `out`(stdout)에 통지를 싣는 것. AC-DRW-003 2번 팔이 붉어진다.
- **AP-4** — `ResultDeployer` 를 `Deployer` 에 흡수하거나 `Deploy` 시그니처를 바꾸는 것. 선행 SPEC 이 선택적으로 설계한 이유를 지운다.
- **AP-5** — 비공개 test seam(`withSymlinkFunc` 등)을 공개로 승격해 CLI 테스트에서 쓰는 것. 이중체로 충분하다(R2).
- **AP-6** — 통지 존재만 단언하고 개수를 보지 않는 테스트. 배선을 부분적으로 끊어도 통과한다.
- **AP-7** — "34줄 미만" 같은 상한 단언으로 비비례성을 증명했다고 주장하는 것(AC-DRW-004 [HARD] 참조).
- **AP-8** — `internal/template` 을 함께 고치는 것. F1/F4 는 승계 카드 소관이며, 여기서 손대면 두 카드가 같은 파일에서 충돌한다.
- **AP-9** — **기본값이나 미주입 구성에서 판정하는 것.** clean-reinstall 의 `Out` nil 분기, `ErrOut` nil 분기, R1 seam 의 미주입 상태 — 어느 것도 "프로덕션이 이렇게 동작한다" 의 근거가 아니다. 판정은 프로덕션이 실제로 주입하는 값 위에서 한다(감사 D1·D5·D7 의 공통 형태).
- **AP-10** — 코드에 선언된 **기본값을 읽고 스트림을 판정하는 것**. 호출부가 무엇을 주입하는지 확인해야 한다. iter-1 이 이 형태로 §A.4 를 틀렸다.
- **AP-11** — **위증 검사를 적어 놓고 그것이 자기 AC 안에서 재현 가능한지 보지 않는 것.** iter-2 는 9/9 전부에 위증 검사를 달았으나 `AC-DRW-003` 의 한 건은 그 AC 가 볼 수 없는 대상(호출부 리터럴)을 가리켰다(감사 N4). 위증 검사를 쓸 때는 "이 Given 안에서 그 잘못된 구현을 실제로 만들 수 있는가" 를 함께 확인한다.
- **AP-12** — `failed` 경고 여러 건을 한 줄로 이어 붙여 줄 수 단언을 통과시키는 것. 명세되지 않은 출력 형식을 판정이 강제하게 되고 spec §D(구현 세부는 run-phase 소관)와 충돌한다.

## §H. 교차 참조

- `spec.md` §A.2·§A.3(호출부 3곳 + init 비용 실측), §B.D1..D5.
- `acceptance.md` §D.1 AC-DRW-001..009 + §D.6(기록된 잔존: AST 가드의 한계).
- `.moai/reports/t81/sync-audit.md` F1 · F2 · F4.
