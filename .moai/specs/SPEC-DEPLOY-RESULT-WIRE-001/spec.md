---
id: SPEC-DEPLOY-RESULT-WIRE-001
title: "배포 결과 seam 소비 — 스킬 미러 복사 폴백을 CLI 가 사용자에게 알린다"
version: "0.1.0"
status: draft
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

| 호출부 | 현재 writer | 근거 |
|---|---|---|
| update template sync | `out := cmd.OutOrStdout()` (`update_template_sync.go:51`) | 진행 TUI 전용 |
| clean reinstall | `opts.Out`, nil 이면 `os.Stderr` (`update_clean_install.go:55-56,138-140`) | 주석이 "progress / diagnostic writer" 로 명시 |
| init | 경고 수집기 → stderr (`init.go:704-706`, 주석이 REQ-CTX-012 인용) | 이미 stderr |

세 곳 중 두 곳은 이미 경고를 stderr 로 낸다. 남은 한 곳(update template sync)의 `out` 은 stdout 이므로 **경고를 그 writer 에 실으면 독트린 위반**이다.

### A.5 두 번째 배포부터는 폴백이 `copy` 가 아니라 `skipped` 다

sync-audit F1 실측: 폴백 플랫폼에서 1회차는 `mode=copy`, 2회차부터는 같은 스킬이 `mode=skipped` 로 관측된다(배포기가 자기 산출물과 사용자 디렉터리를 가르는 판별자를 갖지 않기 때문). 따라서 `CopyFallbackUsed()` 는 **1회차에만 참**이다. 이 SPEC 이 닫는 것은 "폴백이 일어난 실행에서 침묵하지 않는다" 이며, 2회차 이후의 고착 상태 고지는 판별자(승계 카드 `t173`)가 들어와야 가능하다 — §D 에 잔존으로 명시한다.

### A.6 `skipped` 경고 문구는 오귀속한다

sync-audit F4: `skill_mirror.go:196` 의 문구 `a non-symlink entry already exists at … — left untouched` 는 A.5 상황에서 **배포기 자신의 지난 복사본을 사용자 항목으로 오귀속**한다. 폴백 플랫폼 2회차 이후 매 실행마다 스킬 수(현재 34건)만큼 발생한다. 문구 수정은 판별자 소관(승계 카드)이므로 이 SPEC 은 **문구를 고치지 않는다**. 다만 그 문구를 사용자에게 올릴지 말지는 이 SPEC 의 결정 사항이다 — §B.D3.

## §B. 설계 결정

### D1 — init 경로를 범위에 넣는다

세 호출부 전부를 배선한다. 근거는 §A.3 — init 배선의 실제 비용이 update 호출부와 같은 급(호출부 한 곳에서 결과를 읽어 기존 경고 통로에 싣는 것)이다.

**넣지 않았을 때의 결과를 기록한다**: 폴백 플랫폼(권한 없는 Windows 등) 사용자가 MoAI-ADK 를 처음 만나는 명령은 `moai update` 가 아니라 `moai init` 이다. init 을 빼면 **가장 흔한 첫 접촉에서 여전히 침묵**하고, `CHANGELOG` 의 "does not currently reach you" 도 절반만 지울 수 있다. 비용이 같은 급인데 침묵의 절반이 남는 배분이므로 넣는다.

### D2 — 소비는 선택적 인터페이스 승격으로만 한다

`Deploy` 시그니처를 바꾸지 않고, `ResultDeployer` 를 필수로 만들지 않는다. 소비 형태는 `deployer.go:67` 이 문서화한 그대로 `if rd, ok := dep.(template.ResultDeployer); ok { … }` 다. `AC-CSC-006` 이 판정을 반환값에 고정한 것은 의도된 계약이므로 재론하지 않는다.

### D3 — `skipped` 모드 경고 문구는 사용자에게 올리지 않는다

§A.6 의 오귀속 문구를 그대로 올리면, 폴백 플랫폼 2회차 이후 매 실행마다 **34줄의 사실과 다른 경고**("당신이 만든 항목을 건드리지 않았습니다" — 실제로는 우리 복사본)가 나간다. 침묵보다 나쁘다: 사용자를 자기 파일을 찾게 만들고, 찾을 것이 없다.

그래서 이 SPEC 이 사용자에게 올리는 것은 (a) **복사 폴백 발생 사실**(정확함)과 (b) **`failed` 모드 경고**(문구가 정확함: "symlink and copy both failed") 두 가지다. `skipped` 는 판별자가 들어올 때까지 보류한다. 이 보류는 §D 에 잔존으로 명시한다 — 무언의 생략으로 남기지 않는다.

### D4 — 통지는 스킬 수에 비례하지 않는 요약이다

34개 스킬이 복사로 떨어져도 34줄이 나가서는 안 된다. 통지는 **개수 + 결과**를 담은 고정 길이 요약이다(취지: "N개 스킬이 링크 대신 복사됐고, 그 복사본은 이후 정본 갱신을 따라가지 않는다"). 스킬 이름 열거는 기본 출력에 넣지 않는다.

### D5 — 스트림은 stderr

§A.4 독트린에 따라 경고는 stderr 다. update template sync 의 `out`(stdout)에 싣지 않는다.

## §C. 요구사항 (GEARS)

- **REQ-DRW-001** — 배포 실행이 복사 폴백 미러 항목을 하나 이상 보고할 때(When), CLI 는 폴백이 일어났다는 사실과 그 개수를 담은 통지를 사용자에게 표시해야 한다(shall).
- **REQ-DRW-002** — 배포 실행이 복사 폴백 항목도 실패 항목도 보고하지 않을 때(When), CLI 는 미러 관련 출력을 어느 스트림에도 내보내서는 안 된다(shall not).
- **REQ-DRW-003** — 배포 실행이 `failed` 모드 미러 항목을 보고할 때(When), CLI 는 해당 항목의 경고를 사용자에게 표시해야 한다(shall).
- **REQ-DRW-004** — CLI 는 미러 통지를 프로젝트 출력 독트린이 경고에 배정한 스트림, 곧 stderr 로 내보내야 하며(shall), stdout 으로 내보내서는 안 된다(shall not) — 근거 §A.4.
- **REQ-DRW-005** — 배포기가 결과 보고 확장(`ResultDeployer`)을 구현하지 않는 경우(Where), 배포는 확장이 없을 때와 동일하게 완료되어야 하며(shall), CLI 는 미러 통지를 내보내지 않아야 한다(shall not).
- **REQ-DRW-006** — 이 변경은 `Deploy` 의 시그니처를 바꾸어서는 안 되며(shall not), 결과 보고 확장을 `Deployer` 구현의 필수 요건으로 만들어서도 안 된다(shall not).
- **REQ-DRW-007** — 통지의 길이는 미러 대상 스킬 수에 비례해 증가해서는 안 된다(shall not) — 근거 §B.D4.
- **REQ-DRW-008** — `moai init` 경로, `moai update` 템플릿 동기화 경로, clean-reinstall 경로 **세 곳 모두**가 결과 seam 을 소비해야 한다(shall) — 근거 §A.2·§A.3·§B.D1.
- **REQ-DRW-009** — 배포기가 자기 산출물과 사용자 항목을 가르는 판별자를 갖기 전까지(While), CLI 는 `skipped` 모드 항목의 경고 문구를 사용자에게 표시해서는 안 된다(shall not) — 근거 §A.6·§B.D3.

## §D. 범위 밖 (Exclusions)

### Out of Scope — 오귀속 문구 자체의 수정

- `skill_mirror.go:196` 의 `a non-symlink entry already exists at …` 문구는 이 SPEC 이 **고치지 않는다**. 정확한 문구를 쓰려면 배포기가 자기 복사본과 사용자 디렉터리를 구분해야 하고, 그 판별자는 승계 카드(`t173`, sync-audit F1/F4) 소관이다.
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
