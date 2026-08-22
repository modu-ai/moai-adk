---
id: SPEC-DEPLOY-RESULT-WIRE-001
title: "배포 결과 seam 소비 — 스킬 미러 복사 폴백을 CLI 가 사용자에게 알린다"
version: "0.3.0"
status: in-progress
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: P2
phase: "v3.1.3 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "deployer, skill-mirror, cli, warning, output-stream, codex"
---

# SPEC-DEPLOY-RESULT-WIRE-001 — 배포 결과 seam 소비

## HISTORY

- 2026-08-22 (plan-phase, iter-3, v0.3.0) — plan-audit iteration 2(FAIL 0.75, 반복 상한 2/2 도달) 후 리드가 **정정 후 PASS-with-debt** 로 수용한 5건을 닫는다. **N2 가 구현자를 함정에 빠뜨릴 유일한 건이었다**: `AC-DRW-004` 2번 팔이 **올바른 구현에서 붉어졌다** — `REQ-DRW-003` 이 개수 요약 1줄 + 예시 최대 3건이므로 실패 2건은 `1+2=3` 줄, 34건은 `1+3=4` 줄이고 셋은 넷이 아니다. 낮은 표본을 상한 **위**(4)로 올리고, 상한 자체를 판정하는 **3번 팔**을 따로 세웠다(경고를 한 줄로 이어 붙이는 우회는 명세되지 않은 출력 형식을 판정이 몰래 강제하게 되므로 채택하지 않았다 — spec §D 가 문구를 run-phase 소관으로 둔다). 원인은 복사 팔을 그대로 베낀 것이며, 복사 팔은 요약이 항상 1줄이라 같은 산술이 성립하지 않았다.
  **N4 는 주장을 낮추지 않고 팔을 더해 닫았다.** `AC-DRW-003` 의 두 번째 위증 검사(`ErrOut` 을 추가만 하고 주입하지 않는 구현)는 자기 Given 안에서 재현 불가능했다 — 어떤 AC 도 `update.go` 의 옵션 리터럴을 보지 않았다. `AC-DRW-009` 에 **3번 단언**(`CleanReinstallOptions` 복합 리터럴 전부가 `ErrOut: cmd.ErrOrStderr()` 를 갖는다, `go/parser` 로 확인)을 추가해 그 위증 검사를 실재하게 만들었다. 이 계보가 다섯 번 만들어 낸 결함이 **검사할 수 없는 것을 검사한다고 주장하는 판정**이므로, 주장을 낮추면 그 형태가 그대로 남는다.
  **N5 는 부채로 미루지 않고 갚았다** — `AC-DRW-009` 1번 단언에 검증 수단(AST 가드)을 명시하고, 2번을 런타임 단언으로 분리했다. 다만 **AST 가드는 리터럴을 증명하지 런타임 값을 증명하지 않는다**는 한계는 남으므로 `acceptance.md` §D.6 에 기록된 잔존으로 명시했다. **N1** — `§D` 의 `:196` → `:205`(iter-2 가 §A.6 만 고치고 §D 를 놓쳐 문서가 자기모순이었다). **N6** — plan `§E`·`§H` 의 `AC-DRW-001..008` → `..009`. **N3** — `§A.4` 의 호출 행 `:625` → `:624`(재측정). 요구사항 10 / 판정 9 불변.
  **이번 개정 주기에서 배운 것을 규율로 남긴다.** v0.2.0 은 D1 에서 규칙을 뽑았다 — 「판정은 프로덕션 배선에서, 각 판정은 어떤 잘못된 구현에서 붉어지는지를 적는다」. 그리고 그 규칙을 **형식으로는** 9/9 전부에 적용했으나(위증 검사 절), **효력으로는** 규칙이 유래한 결함 바로 한 층 위에 적용하지 못했다 — 위증 검사를 적었지만 그것이 자기 AC 안에서 재현 가능한지는 보지 않았다. 그것이 N4 다. **규칙을 세우고 그 규칙의 출처에 적용하지 않는 것**이 이 계보의 반복 형태이며, 개별 수정보다 이 형태를 기록하는 편이 쓸모 있다.
- 2026-08-22 (plan-phase, iter-2, v0.2.0) — plan-audit iteration 1(FAIL 0.75, Tier M 임계 0.80)의 결함 7건을 닫는다. **무게중심은 D1 하나다: §A.4 의 스트림 실측이 틀렸다.** clean-reinstall 의 `opts.Out` 이 nil 이면 `os.Stderr` 인 것은 사실이나, **프로덕션 호출부 두 곳이 모두 stdout 을 주입**한다(`update.go:425`·`:627` 의 `Out: out`, 그 `out` 은 `:154` `cmd.OutOrStdout()`) — nil 분기는 프로덕션에서 **실행되지 않는다**. 그래서 "세 곳 중 두 곳이 이미 stderr" 는 거짓이고, stderr 로 나가는 곳은 **init 하나**, stdout 위험은 **둘**이다. 파생 결함 두 개를 함께 닫았다: plan M3 이 clean-reinstall 통지를 `out`(=stdout)에 실으라고 지시해 **명세대로 구현하면 REQ-DRW-004 를 매 실행 위반**했고, `AC-DRW-003` 은 plan R4 가 판정을 "기본 경로(주입 없음 = stderr)" 에 고정해 **프로덕션에 존재하지 않는 경로에서 초록**이었다. 후자는 선행 카드가 세 차례 감사에 걸쳐 제거한 「자기가 막으려는 실패에서 통과하는 판정」과 같은 계열이며, 이 계보에서 **네 번째** 재발이다.
  **스스로 물려받은 교훈 하나를 기록한다.** iter-1 은 `nil ⇒ os.Stderr` 라는 **기본값 선언**을 읽고 스트림을 판정했다. 기본값은 호출부가 무엇을 주입하는지 말해 주지 않는다 — **기본값의 존재는 그 기본값이 쓰인다는 근거가 아니다.** D5(주입 seam 의 기본값)와 D7(음성 팔의 위증 검사 부재)도 같은 형태이며, 세 건을 같은 규율로 닫았다: **판정은 프로덕션 배선에서 이뤄지고, 각 AC 는 어떤 잘못된 구현에서 붉어지는지를 본문에 적는다.**
  나머지: **D2** — `REQ-DRW-007` 의 비비례성이 `failed` 항목의 항목별 출력과 모순 → 통지 전체(요약 + `failed`)에 같은 상한 규칙을 적용하고 `AC-DRW-004` 에 `failed` 팔 추가. **D3** — `AC-DRW-003` 을 stdout 위험 **두 호출부**에 명시적으로 고정. **D4** — 오귀속 문구 행 번호 `196` → **`205`**(재측정). **D5** — plan R1 주입 seam 의 기본값을 `REQ-DRW-010` / `AC-DRW-009` 로 가드. **D6** — `failed` 문구의 정확성을 1-of-3 표본에서 일반화한 것을 **3-of-3 실측**으로 교체(§A.7 신설). **D7** — 모든 AC 에 위증 검사(어떤 잘못된 구현에서 붉어지는가)를 명시. 요구사항 9 → **10**, 판정 8 → **9**(Tier M 상한 16/16 이내).
  **범위 결정 하나는 유지한다.** 감사가 §A.3 의 행 인용 4건을 전부 재측정해 정확함을 확인하고, 논증이 필요로 했으나 적지 않았던 연결 두 개(`InitResult` 포인터가 `PhaseExecutor` 를 그대로 통과 · `p.Collect` 가 실제로 stderr 에 도달)까지 확인했다. init 범위 포함은 그대로 둔다.
- 2026-08-22 (plan-phase, iter-1, v0.1.0) — Tier M 최초 작성. 선행 SPEC(`SPEC-CODEX-SKILLS-CANONICAL-001`, `release/v3.1.3` 머지 `23ffb43a1`)이 만든 반환 seam 의 **프로덕션 소비자가 0건**이라는 sync-audit F2 실측을 닫는다. 카드 전제 두 가지를 실측으로 정정했다 — (1) 호출부는 1곳이 아니라 **3곳**이고, (2) 그중 `moai init` 경로를 넣는 비용이 카드 문구("materially larger change")보다 **현저히 작다**. 근거는 §A.2·§A.3.

## §A. 검증된 기준선 (실측)

### A.1 소비자 0건 — 닫으려는 갭

`internal/template/deployer.go:58-74` 의 `ResultDeployer` 는 `DeployWithResult` 로 `*DeployResult` 를 돌려주고, `DeployResult` 는 `CopyFallbackUsed() bool` · `Warnings() []string` · `MirrorMode(skill) (MirrorMode, bool)` 세 접근자를 노출한다. 프로덕션 코드에서 이 인터페이스로 승격하는 곳은 없다(sync-audit F2, grep 실측). 그래서 심볼릭 링크를 만들 수 없는 환경에서 미러가 **복사본으로 떨어진 사실이 사용자에게 도달하지 않는다**.

이 침묵은 이미 공표된 제약이다 — `CHANGELOG.md` `[Unreleased]` 가 "The fallback warning does not currently reach you" 라고 적고 있다. 이 SPEC 이 그 문장을 지울 수 있게 만든다.

### A.2 호출부는 3곳이다 (카드 전제 정정 1)

```
internal/cli/update_template_sync.go:323   deployer.Deploy(...)
internal/cli/update_clean_install.go:439   deployer.Deploy(...)
internal/core/project/initializer.go:356   i.deployer.Deploy(...)
```

카드는 범위를 "internal/cli (update 경로)" 로 적었고 이는 앞의 두 곳만 덮는다. 세 번째는 `moai init` 경로이며 `internal/core/project.Initializer` 뒤에 있다 — `internal/cli/init.go` 는 `Deploy` 를 직접 부르지 않고, `Deployer` 를 만들어 `project.NewInitializer` 에 넘긴 뒤 결과만 받는다.

### A.3 init 경로 배선 비용 — 카드 전제 정정 2

카드는 init 배선을 "initializer 와 phase executor 를 관통해 결과를 되돌려야 하므로 두 update 호출부보다 실질적으로 큰 변경" 으로 적었다. **실측 결과 그렇지 않다.** 되돌릴 통로가 이미 존재한다:

| 실측 항목 | 값 |
|---|---|
| `InitResult.Warnings []string` | `internal/core/project/initializer.go:97` — 이미 존재하는 필드 |
| `deployTemplates` 시그니처 | `(ctx, opts, result *InitResult) error` (`:323`) — `result` 를 **이미 인자로 받는다** |
| 같은 필드의 기존 적재부 | `:202`, `:208`, `:228`, `:247`, `:264`, `:275` — 6곳이 이미 append |
| 사용자 표시부 | `internal/cli/init.go:706` — `for _, w := range result.Warnings { p.Collect(w) }`, 요약 패널로 stderr 에 렌더 |

즉 init 배선은 `deployTemplates` 안에서 `result.Warnings` 에 append 하는 형태로 닫히며, `Initializer` 인터페이스·`InitResult` 구조체·phase executor 중 **어느 것도 바꾸지 않는다**. 이 실측이 §B.D1 의 근거다.

### A.4 출력 스트림 — 프로젝트 독트린

`internal/cli/CLAUDE.md:14` — "stdout = structured machine-readable output … **stderr = human progress messages, warnings, errors**. Never mix." 세 호출부의 현재 writer 는 서로 다르다:

| 호출부 | 프로덕션에서 실제로 쓰이는 writer | 실측 근거 |
|---|---|---|
| update template sync | **stdout** | `out := cmd.OutOrStdout()` (`update_template_sync.go:51`) |
| clean reinstall | **stdout** | `opts.Out` 은 nil 이면 `os.Stderr` 로 떨어지지만(`update_clean_install.go:55-56,138-140`), **프로덕션 호출부 두 곳이 모두 stdout 을 주입한다** — `update.go:425` `Out: out`(호출 `:418`, `runUpdate` 스코프)과 `update.go:627` `Out: out`(호출 `:624`, `emitDryRunReinstallPlan` 의 `out` 인자, `:362` 에서 같은 `out` 을 전달). 두 경로의 `out` 은 모두 `update.go:154` `out := cmd.OutOrStdout()` 이다. **nil 분기는 프로덕션에서 실행되지 않는다.** |
| init | **stderr** | 경고가 `InitResult.Warnings` → `p.Collect`(`init.go:706`) → `p.emitSummary(cmd.ErrOrStderr())`(`init.go:386` defer) |

**stderr 로 나가는 호출부는 하나(init)뿐이고, stdout 위험은 둘이다.** 두 update 계열 모두 경고를 현재 writer 에 실으면 독트린 위반이다.

[HARD] 이 표를 "기본값" 으로 다시 쓰지 않는다. iter-1 이 정확히 그렇게 해서 틀렸다 — `nil ⇒ os.Stderr` 라는 **기본값 선언**은 호출부가 무엇을 주입하는지 말해 주지 않는다. 스트림 판정은 **주입값 실측**으로 한다.

### A.5 두 번째 배포부터는 폴백이 `copy` 가 아니라 `skipped` 다

sync-audit F1 실측: 폴백 플랫폼에서 1회차는 `mode=copy`, 2회차부터는 같은 스킬이 `mode=skipped` 로 관측된다(배포기가 자기 산출물과 사용자 디렉터리를 가르는 판별자를 갖지 않기 때문). 따라서 `CopyFallbackUsed()` 는 **1회차에만 참**이다. 이 SPEC 이 닫는 것은 "폴백이 일어난 실행에서 침묵하지 않는다" 이며, 2회차 이후의 고착 상태 고지는 판별자(승계 카드 `t173`)가 들어와야 가능하다 — §D 에 잔존으로 명시한다.

### A.6 `skipped` 경고 문구는 오귀속한다

sync-audit F4: `skill_mirror.go:205` 의 문구(sync-audit 은 `:196` 으로 인용했고 iter-1 이 그 인용을 재측정 없이 옮겼다 — `grep -n "non-symlink entry already exists"` 실측값은 **205** 다) `a non-symlink entry already exists at … — left untouched` 는 A.5 상황에서 **배포기 자신의 지난 복사본을 사용자 항목으로 오귀속**한다. 폴백 플랫폼 2회차 이후 매 실행마다 스킬 수(현재 34건)만큼 발생한다. 문구 수정은 판별자 소관(승계 카드)이므로 이 SPEC 은 **문구를 고치지 않는다**. 다만 그 문구를 사용자에게 올릴지 말지는 이 SPEC 의 결정 사항이다 — §B.D3.

### A.7 `failed` 모드 문구는 세 곳에서 생산된다 — 3-of-3 정확

iter-1 은 §B.D3 에서 `failed` 문구가 "정확하다" 고 적으면서 표본 하나(`symlink and copy both failed`)만 보았다. 재측정 결과 `MirrorModeFailed` 를 내는 자리는 **셋**이다(`grep -n` 실측):

| 행 | 문구 | 귀속 정확성 |
|---|---|---|
| `skill_mirror.go:217` | `cannot replace stale link: %v` | 정확 — 대상은 우리 링크이고 실패는 `os.Remove` 오류 그대로 |
| `skill_mirror.go:226` | `cannot create %s: %v` | 정확 — 미러 디렉터리 생성 실패, 사용자 항목을 언급하지 않음 |
| `skill_mirror.go:243` | `symlink and copy both failed: %v — skill is reachable via .claude/skills only` | 정확 — 두 시도 모두 실패했다는 사실과 남은 접근 경로를 그대로 서술 |

세 문구 모두 §A.6 의 오귀속(우리 산출물을 사용자 항목이라 부르는 형태)을 갖지 않는다. `failed` 를 사용자에게 올린다는 §B.D3 의 결정은 **3-of-3 실측** 위에 선다. (같은 `Lstat` 분기 안에 있는 `:205` 는 `MirrorModeSkipped` 이지 `failed` 가 아니다 — 두 모드를 한 문단에서 뭉뚱그리지 않는다.)

## §B. 설계 결정

### D1 — init 경로를 범위에 넣는다

세 호출부 전부를 배선한다. 근거는 §A.3 — init 배선의 실제 비용이 update 호출부와 같은 급(호출부 한 곳에서 결과를 읽어 기존 경고 통로에 싣는 것)이다.

**넣지 않았을 때의 결과를 기록한다**: 폴백 플랫폼(권한 없는 Windows 등) 사용자가 MoAI-ADK 를 처음 만나는 명령은 `moai update` 가 아니라 `moai init` 이다. init 을 빼면 **가장 흔한 첫 접촉에서 여전히 침묵**하고, `CHANGELOG` 의 "does not currently reach you" 도 절반만 지울 수 있다. 비용이 같은 급인데 침묵의 절반이 남는 배분이므로 넣는다.

### D2 — 소비는 선택적 인터페이스 승격으로만 한다

`Deploy` 시그니처를 바꾸지 않고, `ResultDeployer` 를 필수로 만들지 않는다. 소비 형태는 `deployer.go:67` 이 문서화한 그대로 `if rd, ok := dep.(template.ResultDeployer); ok { … }` 다. `AC-CSC-006` 이 판정을 반환값에 고정한 것은 의도된 계약이므로 재론하지 않는다.

### D3 — `skipped` 모드 경고 문구는 사용자에게 올리지 않는다

§A.6 의 오귀속 문구를 그대로 올리면, 폴백 플랫폼 2회차 이후 매 실행마다 **34줄의 사실과 다른 경고**("당신이 만든 항목을 건드리지 않았습니다" — 실제로는 우리 복사본)가 나간다. 침묵보다 나쁘다: 사용자를 자기 파일을 찾게 만들고, 찾을 것이 없다.

그래서 이 SPEC 이 사용자에게 올리는 것은 (a) **복사 폴백 발생 사실**과 (b) **`failed` 모드 경고** 두 가지다. (b) 의 문구 정확성은 생산 지점 **셋 전부**를 확인한 결과이며 근거는 §A.7 이다 — 표본 하나에서 일반화하지 않았다. `skipped` 는 판별자가 들어올 때까지 보류한다. 이 보류는 §D 에 잔존으로 명시한다 — 무언의 생략으로 남기지 않는다.

### D4 — 통지 **전체**가 스킬 수에 비례하지 않는다

34개 스킬이 복사로 떨어져도 34줄이 나가서는 안 된다. 복사 폴백 통지는 **개수 + 결과**를 담은 고정 길이 요약이다(취지: "N개 스킬이 링크 대신 복사됐고, 그 복사본은 이후 정본 갱신을 따라가지 않는다").

**같은 상한이 `failed` 출력에도 적용된다.** iter-1 은 비비례성을 복사 요약에만 걸어 두고 `failed` 는 항목별로 내보내게 설계했는데, `failed` 도 34건이 될 수 있으므로 그 설계는 자기 요구사항과 모순이었다. 정정: `failed` 출력도 **개수 요약 1줄 + 예시 경고 최대 3건**으로 상한을 둔다. 스킬 이름이 필요한 진단 정보는 예시 3건이 나르고, 그보다 많은 경우 개수가 규모를 나른다. 스킬 이름 전량 열거는 어느 모드에서도 기본 출력에 넣지 않는다.

### D5 — 스트림은 stderr, 판정은 프로덕션 배선에서

§A.4 독트린에 따라 경고는 stderr 다. **두 update 계열 모두** 현재 writer 가 stdout 이므로(§A.4), 두 곳 다 별도의 stderr writer 를 잡아야 한다 — update template sync 는 `cmd.ErrOrStderr()`, clean-reinstall 은 `CleanReinstallOptions` 에 새 `ErrOut io.Writer` 를 두고 호출부에서 `cmd.ErrOrStderr()` 를 주입한다(기존 `Out` 의 의미는 바꾸지 않는다).

[HARD] 이 결정의 판정은 **주입이 일어난 프로덕션 배선**에서 이뤄진다. "주입 없는 기본 경로" 에서 판정하면 프로덕션에 없는 구성에서 초록이 나온다 — iter-1 의 plan R4 가 정확히 그 형태였다.

## §C. 요구사항 (GEARS)

- **REQ-DRW-001** — 배포 실행이 복사 폴백 미러 항목을 하나 이상 보고할 때(When), CLI 는 폴백이 일어났다는 사실과 그 개수를 담은 통지를 사용자에게 표시해야 한다(shall).
- **REQ-DRW-002** — 배포 실행이 복사 폴백 항목도 실패 항목도 보고하지 않을 때(When), CLI 는 미러 관련 출력을 어느 스트림에도 내보내서는 안 된다(shall not).
- **REQ-DRW-003** — 배포 실행이 `failed` 모드 미러 항목을 보고할 때(When), CLI 는 실패 항목의 **개수**와 그중 **최대 3건의 경고 문구**를 사용자에게 표시해야 한다(shall) — 상한 근거는 §B.D4.
- **REQ-DRW-004** — CLI 는 미러 통지를 프로젝트 출력 독트린이 경고에 배정한 스트림, 곧 stderr 로 내보내야 하며(shall), stdout 으로 내보내서는 안 된다(shall not) — 근거 §A.4.
- **REQ-DRW-005** — 배포기가 결과 보고 확장(`ResultDeployer`)을 구현하지 않는 경우(Where), 배포는 확장이 없을 때와 동일하게 완료되어야 하며(shall), CLI 는 미러 통지를 내보내지 않아야 한다(shall not).
- **REQ-DRW-006** — 이 변경은 `Deploy` 의 시그니처를 바꾸어서는 안 되며(shall not), 결과 보고 확장을 `Deployer` 구현의 필수 요건으로 만들어서도 안 된다(shall not).
- **REQ-DRW-007** — 통지 **전체**(복사 폴백 요약과 `failed` 출력을 모두 포함)의 길이는 미러 대상 스킬 수에 비례해 증가해서는 안 된다(shall not) — 근거 §B.D4. 이 요구는 REQ-DRW-003 의 3건 상한과 함께 읽으며, 둘 사이에 예외는 없다.
- **REQ-DRW-008** — `moai init` 경로, `moai update` 템플릿 동기화 경로, clean-reinstall 경로 **세 곳 모두**가 결과 seam 을 소비해야 한다(shall) — 근거 §A.2·§A.3·§B.D1.
- **REQ-DRW-009** — 배포기가 자기 산출물과 사용자 항목을 가르는 판별자를 갖기 전까지(While), CLI 는 `skipped` 모드 항목의 경고 문구를 사용자에게 표시해서는 안 된다(shall not) — 근거 §A.6·§B.D3.
- **REQ-DRW-010** — 판정을 위해 도입되는 배포기 주입 지점(seam)은 아무것도 주입되지 않았을 때 프로덕션 배포기를 사용해야 한다(shall). 주입 지점의 도입이 프로덕션 실행에서 사용되는 배포기를 바꾸어서는 안 된다(shall not) — 근거: 주입 지점은 판정 수단이지 동작 변경이 아니다.

## §D. 범위 밖 (Exclusions)

### Out of Scope — 오귀속 문구 자체의 수정

- `skill_mirror.go:205` 의 `a non-symlink entry already exists at …` 문구는 이 SPEC 이 **고치지 않는다**. 정확한 문구를 쓰려면 배포기가 자기 복사본과 사용자 디렉터리를 구분해야 하고, 그 판별자는 승계 카드(`t173`, sync-audit F1/F4) 소관이다.
- 이 SPEC 은 그 문구를 사용자에게 **올리지 않기로** 결정할 뿐이다(REQ-DRW-009).

### Out of Scope — 폴백 미러의 고착 해소

- §A.5 의 결과로, 폴백 플랫폼에서 **2회차 이후 실행은 통지를 내지 않는다** — 그때의 모드는 `copy` 가 아니라 `skipped` 이기 때문이다. 미러가 낡은 복사본으로 고착돼 있다는 사실은 이 SPEC 이 닫는 창 밖에 있다.
- 이것을 각주가 아니라 본문에 적는 이유는, "폴백이 통지된다"를 "폴백 상태가 항상 통지된다"로 읽을 여지를 없애기 위해서다. 닫히는 것은 **폴백이 일어난 그 실행**이다.

### Out of Scope — 배포기 내부 출력

- 배포기가 직접 출력하도록 바꾸지 않는다. 선행 SPEC 의 `REQ-CSC-005` 가 "배포기 내부에서 직접 출력해서는 안 된다(shall not)" 로 못박았고, 그 근거(`internal/template` 에 출력 표면 부재)는 여전히 유효하다.

### Out of Scope — `MirrorMode(skill)` 접근자의 사용

- 스킬 단위 모드 조회는 이 SPEC 의 통지 형태(요약)에 필요하지 않다. 접근자는 테스트·후속 카드용으로 남는다.

### Out of Scope — 구현 세부

- 통지 문자열의 정확한 어휘, 헬퍼 함수 이름과 배치, 기존 printer/collector 헬퍼 재사용 여부는 run-phase 판단이다.

## §E. 비기능 제약

- **출력 스트림**: `internal/cli/CLAUDE.md:14` (stdout/stderr 분리).
- **OS 중립성**: 통지 경로는 darwin / linux / windows 에서 동일하게 동작해야 한다. 폴백 자체는 `runtime.GOOS` 분기가 아니라 `os.Symlink` 실패라는 관측된 능력 부재로 결정된다(sync-audit §6 실측).
- **Template-First**: 이 SPEC 은 `.claude/` · `.moai/` 아래 신규 파일을 만들지 않으므로 템플릿 미러 대상이 없다. `CHANGELOG.md` 는 템플릿이 아니다.
- **회귀 없음**: `.claude/skills/` 배포 산출물은 이 변경으로 달라지지 않는다 — `AC-CSC-010` 의 동일 프로세스 seam 토글 불변식은 그대로 통과해야 한다.

## §F. 교차 참조

- `SPEC-CODEX-SKILLS-CANONICAL-001` — seam 생산자. `REQ-CSC-005`(반환 결과 형태), `AC-CSC-006`(반환값 기준 판정), §A.9b(출력 표면 부재 실측).
- `.moai/reports/t81/sync-audit.md` — F1(고착) · F2(소비자 0건, 이 SPEC 의 출발점) · F4(오귀속 문구).
- `internal/template/skill_mirror.go` — `DeployResult`, `SkillMirrorEntry`, `MirrorMode` 상수 4종.
- `internal/template/deployer.go:58-74` — `ResultDeployer` 와 문서화된 승격 형태.
- `internal/cli/CLAUDE.md` — 출력 스트림 독트린.
- `CHANGELOG.md` `[Unreleased]` — 이 SPEC 이 착지하면 "The fallback warning does not currently reach you" 항목을 갱신한다.
