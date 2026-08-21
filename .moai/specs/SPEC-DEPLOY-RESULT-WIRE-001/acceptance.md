# acceptance — SPEC-DEPLOY-RESULT-WIRE-001

> [HARD] 모든 판정은 **프로덕션 배선**에서 이뤄진다. 기본값·미주입 구성에서만 성립하는 판정은 이 SPEC 에서 무효다 — iter-1 의 `AC-DRW-003` 이 정확히 그 형태로 실패했다(감사 D1). 그래서 각 AC 는 **위증 검사** 절을 갖는다: 어떤 잘못된 구현에서 이 판정이 붉어지는가.

## §D. 판정 기준 매트릭스

| AC | 대응 REQ | 검증 수단 | 강도 |
|---|---|---|---|
| AC-DRW-001 | REQ-DRW-001 | Go 테스트 (복사 폴백 보고 → 통지·개수 관측) | MUST |
| AC-DRW-002 | REQ-DRW-002 | Go 테스트 (폴백 없음 → 무출력, 위증 팔 포함) | MUST |
| AC-DRW-003 | REQ-DRW-004 | Go 테스트 (**stdout 위험 두 호출부 × 프로덕션 배선**, stderr 양성 · stdout 음성) | MUST |
| AC-DRW-004 | REQ-DRW-007, REQ-DRW-003 | Go 테스트 3팔 (복사 2 vs 34 · `failed` **4** vs 34 줄 수 동일 · 상한 2건 vs 3건) | MUST |
| AC-DRW-005 | REQ-DRW-005, REQ-DRW-006 | Go 테스트 (확장 미구현 배포기 이중체) | MUST |
| AC-DRW-006 | REQ-DRW-003 | Go 테스트 (`failed` 경고 도달 + fail-open 유지) | MUST |
| AC-DRW-007 | REQ-DRW-008 | Go 테스트 (세 호출부 3팔) | MUST |
| AC-DRW-008 | REQ-DRW-009 | Go 테스트 (`skipped` 전용 → 오귀속 문구 부재) | MUST |
| AC-DRW-009 | REQ-DRW-010, REQ-DRW-004 | Go 테스트 3단언 (seam 기본값 AST · 기본값의 `ResultDeployer` 만족 런타임 · 호출부 `ErrOut` 주입 AST) | MUST |

요구사항 10개 / 판정 기준 9개 — Tier M 상한 16/16 이내.

## §D.1 판정 기준

### AC-DRW-001 — 복사 폴백이 사용자에게 도달한다

**Given** 결과 seam 이 **복사 폴백 항목을 보고하는** 배포기가 CLI 경로에 주입돼 있고(`template` 패키지 안이면 심볼릭 링크 실패 주입 seam 으로, `internal/cli` 쪽이면 `ResultDeployer` 를 구현하고 `SkillMirrors` 에 `MirrorModeCopy` 항목을 담아 돌려주는 이중체로 — 두 구성 모두 판정에 동등하다), 그 경로를 캡처 가능한 writer 로 실행할 때,
**When** 배포를 1회 실행하고 캡처된 출력을 읽으면,
**Then** 출력에 (1) 복사 폴백이 일어났음을 나타내는 통지가 **1건 이상** 있고, (2) 그 통지가 폴백 스킬 **개수**를 담고 있어야 한다.

[HARD] 통지의 존재만 단언하지 않는다. 개수 없는 통지는 규모를 감추고, 개수 단언이 없으면 배선을 부분적으로 끊어도 통과할 여지가 남는다.

**위증 검사**: 소비부(`if rd, ok := dep.(template.ResultDeployer); ok`)를 제거한 구현 — 판정 붉어짐. 통지는 내지만 개수를 빼먹은 구현 — 2번 단언에서 붉어짐.

### AC-DRW-002 — 폴백이 없으면 아무것도 나가지 않는다

**Given** 결과 seam 이 `MirrorModeSymlink` 항목만 보고하는(폴백·실패 0건) 배포기가 주입돼 있고,
**When** 같은 CLI 경로를 실행해 stdout · stderr 를 **둘 다** 캡처하면,
**Then** 양쪽 어디에도 미러 관련 문자열이 **없어야** 한다.

"미러 관련 문자열" 의 판정 어휘는 run-phase 가 통지 문자열을 확정할 때 그 문자열의 안정 부분식(예: 미러 경로 토큰)으로 고정하고, 이 AC 와 AC-DRW-001 이 **같은 부분식**을 쓴다 — 서로 다른 어휘를 쓰면 두 팔이 같은 대상을 보지 않는다.

**위증 검사**: 결과의 내용을 보지 않고 통지를 무조건 내보내는 구현(예: `if rd, ok := …; ok { emit(summary) }` 를 개수 검사 없이 작성) — 이 판정이 붉어져야 한다. AC-DRW-001 은 그 구현에서 초록이므로, **이 AC 가 없으면 무조건 출력 구현이 통과한다.** 이것이 이 AC 의 존재 이유다.

### AC-DRW-003 — 통지는 stderr 로 나간다 (stdout 위험 두 호출부 × 프로덕션 배선)

**Given** stdout 위험 호출부 **둘**을 각각 프로덕션과 동일하게 배선한 상태에서 — 곧 (a) update template sync 는 `out = cmd.OutOrStdout()` 인 채로, (b) clean-reinstall 은 `CleanReinstallOptions.Out` 에 **stdout 을 주입한 채로**(프로덕션 호출부 `update.go:425`·`:627` 이 실제로 하는 것과 동일) — stdout 버퍼와 stderr 버퍼를 **서로 다른 버퍼**로 잡아 두고,
**When** 두 호출부를 각각 폴백 보고 조건으로 실행하면,
**Then** 두 호출부 **각각에 대해** 다음 두 단언이 참이어야 한다.

1. stderr 버퍼에 통지가 있다.
2. stdout 버퍼에 통지가 **없다**.

[HARD] `Out` 을 주입하지 않은 구성에서 판정하지 않는다. clean-reinstall 의 `nil ⇒ os.Stderr` 기본값은 **프로덕션에서 실행되지 않는 분기**이므로(spec §A.4), 그 구성에서의 초록은 제품에 대해 아무것도 증명하지 않는다. iter-1 이 이 형태로 실패했다.

[HARD] 1번만 쓰지 않는다. 두 스트림 모두에 쓰는 구현은 1번만으로 통과하며, 이는 `internal/cli/CLAUDE.md:14` 의 "Never mix" 를 정확히 위반하는 형태다.

**위증 검사**: 통지를 `out` 에 싣는 구현(iter-1 의 plan M3 이 지시했던 형태) — clean-reinstall 팔의 2번 단언에서 붉어져야 한다. `ErrOut` 을 추가했으나 호출부에서 주입하지 않아 nil 로 남는 구현 — 1번 단언에서 붉어져야 한다.

### AC-DRW-004 — 통지 길이가 스킬 수에 비례하지 않는다 (세 팔)

**Given** 같은 테스트 프로세스 안에서 항목 수만 다른 결과를 각각 주입할 수 있고,
**When** 아래 세 팔을 각각 실행하면,
**Then** 각 팔의 단언이 참이어야 한다.

1. **복사 팔 (비비례성)** — `MirrorModeCopy` **2개** 실행과 **34개** 실행의 통지 줄 수가 **같다**.
2. **`failed` 팔 (비비례성)** — `MirrorModeFailed` **4개** 실행과 **34개** 실행의 통지 줄 수가 **같다**.
3. **`failed` 팔 (상한 자체)** — `MirrorModeFailed` **2개** 실행은 경고 문구를 **2건**, **34개** 실행은 **3건** 보여 준다.

[HARD] 2번 팔의 낮은 표본은 반드시 **상한(3) 위**여야 한다. iter-2 는 복사 팔을 그대로 베껴 낮은 표본을 **2**로 두었는데, `REQ-DRW-003` 이 개수 요약 1줄 + 예시 최대 3건이므로 **올바른 구현이** `1+2=3` 대 `1+3=4` 로 **붉어진다** — 판정이 제품이 아니라 자기 산술을 검사하고 있었다(감사 N2). 복사 팔에서는 요약이 항상 1줄이라 이 문제가 생기지 않아 베끼기가 무사통과했다.

[HARD] 이것을 "경고를 한 줄로 이어 붙인다" 로 고치지 않는다. 그러면 아무도 명세하지 않은 출력 형식을 판정이 몰래 강제하게 되고, 문구를 run-phase 소관으로 둔 spec §D(구현 세부)와 충돌한다.

[HARD] 상한 판정을 2번 팔에 겹쳐 싣지 않는다. 줄 수 동일성(비비례)과 상한(3건)은 서로 다른 요구사항이며, 한 팔에 겹치면 어느 쪽이 깨져서 붉어졌는지 구분되지 않는다. 3번 팔이 상한을 본다.

[HARD] 2번 팔을 빼지 않는다. iter-1 은 `failed` 를 항목별로 내보내도록 설계해 놓고 비비례성 판정을 복사 팔에만 걸었다(감사 D2). 2번 팔이 그 모순을 검출하는 유일한 팔이다.

**위증 검사**: `failed` 항목을 전량 항목별로 출력하는 구현 — 2번 팔에서 `4 != 34` 로 붉어진다. 상한을 3이 아닌 값(예: 5)으로 둔 구현 — 3번 팔에서 붉어진다. 상한을 1로 둔 구현 — 3번 팔의 `N=2 → 2건` 에서 붉어진다.

### AC-DRW-005 — 확장을 구현하지 않는 배포기에서도 배포가 완료된다

**Given** `template.Deployer` 만 구현하고 `DeployWithResult` 를 **구현하지 않는** 테스트 이중체를 CLI 경로에 주입하고,
**When** 배포를 실행하면,
**Then** 세 단언이 모두 참이어야 한다.

1. 배포가 오류 없이 완료된다.
2. panic 이 발생하지 않는다.
3. 미러 통지가 어느 스트림에도 나가지 않는다.

[HARD] 이중체는 `ResultDeployer` 를 **컴파일 시점에 만족하지 않아야** 한다. `DeployWithResult` 를 정의해 두고 nil 을 돌려주는 형태로 쓰면 이 판정이 검사하려는 상태(확장 부재)가 재현되지 않는다.

**위증 검사**: 타입 단언 없이 `dep.(template.ResultDeployer)` 를 무조건 승격하는 구현 — 2번 단언에서 panic 으로 붉어져야 한다.

### AC-DRW-006 — `failed` 경고가 도달한다

**Given** 결과 seam 이 `MirrorModeFailed` 항목(경고 문구 포함)을 보고하는 배포기가 주입돼 있고,
**When** 배포를 실행해 stderr 를 읽으면,
**Then** 세 단언이 모두 참이어야 한다.

1. 실패 **개수**가 stderr 에 나타난다.
2. 실패 경고 문구가 **1건 이상 3건 이하** stderr 에 나타난다.
3. 배포 자체는 오류 없이 완료된다(선행 SPEC 의 fail-open 계약 유지).

**위증 검사**: `failed` 를 전량 무시하는 구현 — 1·2번에서 붉어짐. `failed` 를 배포 오류로 승격하는 구현 — 3번에서 붉어짐(선행 `REQ-CSC-011` 회귀).

### AC-DRW-007 — 세 호출부가 모두 배선돼 있다 (3팔)

**Given** `moai init` 경로, `moai update` 템플릿 동기화 경로, clean-reinstall 경로 각각에 대해 AC-DRW-001 과 동일한 폴백 보고 조건을 구성할 수 있고,
**When** 세 경로를 각각 실행하면,
**Then** 세 경로 **모두**에서 통지가 사용자 표시 통로에 도달해야 한다.

- init 팔의 도달 지점은 `InitResult.Warnings` 이며(그 통로가 `internal/cli/init.go:706` `p.Collect` → `:386` `defer p.emitSummary(cmd.ErrOrStderr())` 로 stderr 에 렌더된다), 판정은 그 슬라이스에 통지가 담기는 것으로 한다.
- update 두 팔의 도달 지점은 캡처된 **stderr** 다(AC-DRW-003 과 같은 배선).

[HARD] 한 팔만 단언하는 형태로 쓰지 않는다. 세 호출부는 서로 다른 파일에 있고 한 곳만 배선한 구현이 실재 가능하므로, 팔이 하나라도 빠지면 REQ-DRW-008 은 검사되지 않는다.

**위증 검사**: 세 호출부 중 하나의 소비부를 지운 구현 — 해당 팔에서 붉어져야 한다(팔 3개를 각각 독립적으로 확인한다).

### AC-DRW-008 — 오귀속 문구는 사용자에게 나가지 않는다

**Given** 결과 seam 이 `MirrorModeSkipped` 항목만 보고하고 그 항목들이 오귀속 경고 문구를 담고 있는 상태에서(`template` 패키지 안이면 미러 대상 경로에 링크가 아닌 실 디렉터리를 미리 만들어 재현한다),
**When** 배포를 실행해 stdout · stderr 를 모두 읽으면,
**Then** `a non-symlink entry already exists` 계열 문구가 어느 스트림에도 **없어야** 한다.

이 판정은 §B.D3 의 결정을 고정한다. 판별자(승계 카드 `t173`)가 들어와 문구가 정확해지면 이 AC 는 그 카드에서 뒤집히는 것이 정상이며, 그때까지는 이 상태가 의도된 상태다.

**위증 검사**: `res.Warnings()` 를 통째로 forward 하는 구현 — 이 판정이 붉어져야 한다(`Warnings()` 는 모드를 구분하지 않고 모든 비어 있지 않은 경고를 돌려주므로, `skipped` 문구가 그대로 실린다).

### AC-DRW-009 — 프로덕션 배선이 실제로 통지를 낼 수 있다 (세 단언)

**Given** 아무것도 주입하지 않은 상태 — 곧 프로덕션이 실제로 실행하는 구성에서,
**When** 아래 세 가지를 각각 관측하면,
**Then** 세 단언이 모두 참이어야 한다.

1. **seam 기본값의 구성** — update template sync 의 배포기 주입 지점(plan R1)이 아무것도 주입되지 않았을 때 `template.NewDeployerWithRendererAndForceUpdate` 를 임베드 FS · renderer · `forceUpdate=true` 로 호출한다. **검증 수단**: `go/parser` 로 `internal/cli/update_template_sync.go` 를 파싱해 seam 변수의 기본값 표현식이 그 호출이고 세 번째 인자가 `true` 임을 확인한다(이 저장소의 소스 스캔 가드 선례: `internal/template/agent_askuser_audit_test.go`).
2. **seam 기본값의 능력** — 1번이 만드는 배포기가 `template.ResultDeployer` 를 **만족한다**. **검증 수단**: 기본값 함수를 실제로 호출해 `_, ok := got.(template.ResultDeployer)` 로 확인한다(런타임 단언). 만족하지 않으면 프로덕션에서 통지가 영원히 나가지 않으면서 AC-DRW-001·003·007 은 이중체 위에서 초록으로 남는다.
3. **clean-reinstall 호출부가 `ErrOut` 을 주입한다** — `internal/cli/update.go` 안의 `CleanReinstallOptions` 복합 리터럴 **전부**가 `ErrOut` 키를 갖고 그 값이 `cmd.ErrOrStderr()` 다. **검증 수단**: `go/parser` 로 `update.go` 를 파싱해 `CleanReinstallOptions` 복합 리터럴을 모두 수집하고(착수 시점 실측 2건 — `:418`, `:624`), 각각의 키 집합에 `ErrOut` 이 있는지 확인한다. **리터럴을 개수로 고정하지 않는다** — 나중에 세 번째 호출부가 생겨도 자동으로 판정 대상이 되어야 한다.

[HARD] 3번 단언은 `AC-DRW-003` 의 두 번째 위증 검사("`ErrOut` 을 추가했으나 호출부에서 주입하지 않아 nil 로 남는 구현")를 **실재하게 만드는 단언**이다. `AC-DRW-003` 은 자기 Given 안에서 writer 를 직접 잡으므로 호출부 리터럴을 볼 수 없고, 따라서 그 위증 검사는 자기 AC 안에서 재현 불가능했다(감사 N4). 이 계보가 다섯 번 만들어 낸 결함은 **검사할 수 없는 것을 검사한다고 주장하는 판정**이므로, 주장을 낮추는 대신 **볼 수 있는 AC 로 옮겼다**.

[HARD] 이 AC 가 D1·D5·N4 의 공통 위험을 닫는다. 다른 모든 AC 는 무언가를 **주입한** 구성에서 판정하므로, 주입하지 않은 프로덕션 구성이 실제로 통지를 낼 수 있는지는 **오직 이 AC 만** 본다.

**위증 검사**: seam 기본값을 테스트용 no-op 배포기로 둔 구현 — 1번에서 붉어진다. `ResultDeployer` 를 만족하지 않는 배포기로 둔 구현 — 2번에서 붉어진다. `CleanReinstallOptions.ErrOut` 필드만 추가하고 두 호출부 리터럴에 주입하지 않은 구현 — 3번에서 붉어진다.

## §D.2 심각도

- MUST 9건 / SHOULD 0건. MUST 하나라도 FAIL 이면 판정 전체 FAIL.

## §D.3 추적성

각 AC 는 위 매트릭스에서 정확히 하나 이상의 REQ 에 매핑된다. 매핑되지 않은 REQ 는 없다 — REQ-DRW-001..010 전건이 최소 1개 AC 에 덮인다. 겹치는 두 곳을 명시한다: `REQ-DRW-003` 은 AC-DRW-004(2번 팔 = 비비례, 3번 팔 = 상한)와 AC-DRW-006(도달)이 함께 덮고, `REQ-DRW-004` 는 AC-DRW-003(행동: 실제 스트림)과 AC-DRW-009(구조: 호출부 주입)가 함께 덮는다.

## §D.4 간접 검증 (AC 아님)

다음은 유용하지만 Go 테스트로 반복 실행되지 않으므로 AC 로 세우지 않는다.

- 실제 Windows(권한 없는 계정)에서 `moai init` 실행 → 통지 육안 확인.
- `CHANGELOG.md` `[Unreleased]` 의 "The fallback warning does not currently reach you" 문장 갱신 — sync-phase 산출물.

## §D.5 완료 정의 (Definition of Done)

- MUST 9건 전부 PASS.
- 변경 패키지 대상 테스트 통과(`go test ./internal/cli/... ./internal/core/project/...`) — **전체 스위트는 로컬에서 돌리지 않는다**(CLAUDE.local.md §4/§6).
- `go vet ./internal/cli/... ./internal/core/project/...` exit 0, `golangci-lint run` 무결점.
- `GOOS=windows go vet ./internal/cli/... ./internal/core/project/...` exit 0 — 단, 이것은 **컴파일만 증명**하며 Windows 동작 근거가 아니다(§D.4 로 분리).
- 선행 SPEC 의 `AC-CSC-010`(seam 토글 불변식)이 여전히 PASS.

## §D.6 전방 점검

- 승계 카드(`t173`)가 판별자를 들여오면 AC-DRW-008 은 뒤집힌다. 그 카드는 이 AC 를 "정확해진 문구가 도달한다" 로 대체할 책임을 진다.
- `CleanReinstallOptions.ErrOut` 이 추가되면 그 구조체를 쓰는 다른 호출부(테스트 포함)가 nil 을 남기게 된다. nil 은 `os.Stderr` 로 떨어지도록 두되, **그 기본값을 판정 근거로 쓰지 않는다**(AC-DRW-003 [HARD]).
- **잔존(기록된 한계) — AST 가드는 리터럴을 증명하지, 런타임 값을 증명하지 않는다.** `AC-DRW-009` 의 1번·3번 단언은 소스 구조를 본다. 소스가 `ErrOut: cmd.ErrOrStderr()` 라고 적혀 있어도 `cmd` 가 그 시점에 stderr 가 아닌 writer 를 들고 있으면 실제 출력은 다를 수 있다. 그 구멍을 완전히 닫으려면 `runUpdate` 를 종단간으로 돌려 실제 스트림을 캡처해야 하고, 이는 이 카드 범위를 넘는 통합 테스트다. **완화**: 2번 단언은 런타임이고, `AC-DRW-003` 이 같은 writer 배선을 행동으로 판정한다 — 셋이 함께 서면 남는 구멍은 "cobra 커맨드의 `ErrOrStderr()` 가 stderr 가 아닌 경우" 뿐이며, 그것은 이 SPEC 이 아니라 cobra 배선의 문제다. 이 한계를 감수하고 진행한다.
