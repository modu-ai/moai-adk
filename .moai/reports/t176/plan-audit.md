# SPEC Review Report: SPEC-DEPLOY-RESULT-WIRE-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.75** (Tier M PASS 임계값 0.80 — `spec-workflow.md:329-330`)

작성자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. 판정 대상은 디스크의 artifact 뿐이다.

## 0. 고정 상태 (audit anchor)

```
$ git log --oneline -1
7a1aa100d docs(spec): add SPEC-DEPLOY-RESULT-WIRE-001 plan-phase artifacts (card t176)
$ git status --short
(빈 출력)
$ git branch --show-current
WT-deploy-result-wire
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t176
```

본 감사의 모든 실측은 이 커밋 기준이며, 감사 중 트리 변경은 없었다.

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `spec.md:102-110` 에 REQ-DRW-001..009. 순차, 중복 0, zero-padding 일관.
- **[PASS] MP-2 GEARS 형식 준수** — 요구사항 계층(`REQ-XXX`, `spec.md:102-110`) 9건 전부가 GEARS 패턴에 대응한다: 001/002/003 event-driven(`…할 때(When) … shall`), 004/008 ubiquitous, 005 where(`구현하지 않는 경우(Where)`), 006/007 unwanted(`shall not`), 009 state-driven+unwanted(`갖기 전까지(While) … shall not`). 판정은 **요구사항 계층에 대해서만** 내렸다 — `acceptance.md` 의 AC-DRW-001..008 은 검증 계층이며 Given-When-Then 이 정상 형식이므로 Group 4 에서 별도 채점했다.
- **[PASS] MP-3 YAML frontmatter 유효성** — `spec.md:2-14`. 정본 12필드 전건 존재·타입 적합: `id`/`title`/`version:"0.1.0"`/`status:draft`/`created:2026-08-22`/`updated:2026-08-22`/`author`/`priority:P2`/`phase`/`module`/`lifecycle:spec-anchored`/`tags`. 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `tier: M` 은 허용된 선택 필드.
- **[N/A] MP-4 §22 언어 중립성** — 단일 언어(Go) 저장소 내부 코드 대상 SPEC. 템플릿 신규 파일 없음(`spec.md:140` 이 스스로 명시, 실측으로 확인). 자동 PASS.
- **[PASS] MP-5 D7 교차 SPEC 조정** — 본문 참조 SPEC 은 `SPEC-CODEX-SKILLS-CANONICAL-001` 1건. `.moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/spec.md:5` → `status: implemented`. retired/superseded/archived 아님 → 조정 의무 미발생. BLOCKING 0건.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — `grep -c syscall spec.md` = **0**. 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-DEPLOY-RESULT-WIRE-001/` = 매치 0건 (`research.md` 부재, Tier M 정상).

must-pass 실패는 없다. FAIL 판정은 아래 rubric 점수와 blocking 결함 3건에서 나온다.

## Category Scores

| 차원 | 점수 | Rubric Band | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | REQ-DRW-007 이 무한정("미러 대상 스킬 수에 비례")으로 쓰여 plan §F M1 의 설계와 충돌한다(D2). 나머지 8건은 단일 해석. |
| Completeness | 1.00 | 1.0 | HISTORY(`spec.md:19`) · 맥락 §A · 결정 §B · 요구사항 §C · 범위 밖 §D(`### Out of Scope — …` H3 5개, 각 `-` 항목 보유) · 비기능 §E · 교차참조 §F. frontmatter 완전. |
| Testability | 0.50 | 0.50 | AC-DRW-003 의 판정 경로가 **프로덕션에서 도달 불가**(D1), AC-DRW-004 가 `failed` 경로를 보지 않음(D2), AC-DRW-003 이 호출부에 결박되지 않음(D3). |
| Traceability | 1.00 | 1.0 | `acceptance.md:5-14` 매트릭스 실측 — REQ-DRW-001..009 전건이 최소 1개 AC 에 덮이고(005·006 은 AC-DRW-005 가 공동 덮음), AC 8건 모두 실재 REQ 를 지목한다. 고아 0건. |

**Overall = 조화평균(0.75, 1.00, 0.50, 1.00) = 0.75.** Tier M 임계값 0.80 미달.

## 1. init 범위 결정에 대한 판정 — **작성자가 옳다 (검증 완료)**

리드의 전제("`moai init` 배선이 두 update 호출부보다 실질적으로 큰 변경")는 **틀렸고**, 작성자의 §A.3 반박은 인용한 4개 라인 참조가 **전부 정확**하며 결론도 성립한다. 전 항목 재측정:

| §A.3 주장 | 실측 | 판정 |
|---|---|---|
| `InitResult.Warnings []string` @ `initializer.go:97` | `97: Warnings []string // Non-fatal warnings during initialization.` | 일치 |
| `deployTemplates` 가 `result *InitResult` 를 인자로 받음 @ `:323` | `323: func (i *projectInitializer) deployTemplates(ctx context.Context, opts InitOptions, result *InitResult) error {` | 일치 |
| 기존 append 6곳 @ `:202,208,228,247,264,275` | 동일 6곳, 파일 전체 grep 결과와 정확히 일치(다른 append 없음) | 일치 |
| 렌더러 @ `init.go:706` — `for _, w := range result.Warnings { p.Collect(w) }` | `706-708` 에 해당 루프 존재 | 일치 |

**작성자가 명시하지 않았으나 결론이 서기 위해 필요한 두 링크를 감사가 추가로 검증했다** — 두 링크 모두 성립한다:

1. **`InitResult` 포인터가 executor 를 관통해 그대로 돌아온다.** `phase.go:123` `result, err := pe.initializer.Init(ctx, opts)` → `phase.go:137` `return result, nil`. PhaseExecutor 는 `Warnings` 를 읽지도 복사하지도 재구성하지도 않는다(`grep Warnings phase.go` = 0건). `deployTemplates` 가 채운 슬라이스가 `init.go:691` 의 `result` 에 그대로 도달한다.
2. **`p.Collect` 가 실제로 stderr 로 렌더된다.** `init_warnings.go:52-56` 의 `Collect` 는 즉시 출력하지 않고 누적만 하며, `init.go:386` `defer p.emitSummary(cmd.ErrOrStderr())` 가 종료 시 `init_warnings.go:69-84` 의 요약 패널을 **stderr** 에 1회 방출한다. 경고 0건이면 아무것도 내보내지 않는다(`:72`). 즉 §A.3 의 "요약 패널로 stderr 에 렌더" 는 정확하다.

따라서 **`Initializer` 인터페이스·`InitResult` 구조체·`PhaseExecutor` 중 어느 것도 바꿀 필요가 없다**는 결론은 성립하며, D1(init 범위 포함) 의 근거는 건전하다. 이 범위 결정은 유지되어야 한다.

부수 확인: `initializer.go:199` `if i.deployer != nil` 은 `init.go:666` `project.NewInitializer(deployer, mgr, nil)` 로 항상 참이고, `deployTemplates` 실패는 비치명(`:200-204`)이라 `err == nil` 경로가 유지되어 `:706` 루프에 도달한다.

## Defects Found

### D1 — §A.4 의 clean-reinstall 스트림 실측이 틀렸고, 그 오류가 REQ-DRW-004 와 AC-DRW-003 을 함께 무력화한다

- **위치**: `spec.md:61`(§A.4 표), `spec.md:64`(§A.4 결론), `plan.md:16`(R4), `plan.md:59`(M3)
- **Severity: critical — Class: blocking**

SPEC §A.4 는 clean-reinstall 의 writer 를 이렇게 적었다:

> \| clean reinstall \| `opts.Out`, nil 이면 `os.Stderr` (`update_clean_install.go:55-56,138-140`) \| 주석이 "progress / diagnostic writer" 로 명시 \|

그리고 §A.4 는 이렇게 결론짓는다: **"세 곳 중 두 곳은 이미 경고를 stderr 로 낸다. 남은 한 곳(update template sync)의 `out` 은 stdout 이므로…"**

`nil ⇒ os.Stderr` 기본값이 존재한다는 것은 사실이다(`update_clean_install.go:138-140`). **그러나 프로덕션 호출부 2곳 모두 `Out` 을 주입하며, 주입값은 stdout 이다.** 실측:

```
$ grep -rn "CleanReinstallOptions{" internal/ | grep -v _test
internal/cli/update.go:418:  result, runErr := runCleanReinstall(ctx, cwd, CleanReinstallOptions{
internal/cli/update.go:624:  if _, runErr := runCleanReinstall(planCtx, cwd, CleanReinstallOptions{

update.go:425:    Out:   out,
update.go:627:    Out:   out,

$ awk '{if ($0 ~ /out[ ]*:=/) print NR": "$0}' internal/cli/update.go
154: 	out := cmd.OutOrStdout()
696: 	out := cmd.OutOrStdout()
783: 	out := cmd.OutOrStdout()
```

`update.go:418`·`:624` 는 모두 `runUpdate`(`:138`) 및 그 하위 `emitDryRunReinstallPlan`(`:592`, 호출부 `:362` 가 같은 `out` 을 전달) 안에 있고, 그 범위에서 유효한 `out` 정의는 `:154` `out := cmd.OutOrStdout()` 하나뿐이다(파일 전체에서 다음 `out :=` 는 `:696`). 즉 **프로덕션에서 `CleanReinstallOptions.Out` 은 언제나 stdout 이며, `nil ⇒ os.Stderr` 분기는 프로덕션에서 한 번도 실행되지 않는다.**

세 가지 귀결이 동시에 발생한다:

1. **§A.4 의 사실 주장이 거짓이다.** stderr 로 경고를 내는 호출부는 두 곳이 아니라 **init 한 곳**이다. stdout 위험 호출부는 한 곳이 아니라 **두 곳**(update template sync + clean-reinstall)이다.
2. **plan §F M3 의 설계가 REQ-DRW-004 를 정면으로 깬다.** M3 는 ``update_clean_install.go:439 — opts.Deployer 승격 후 M1 문자열을 `out` 에 출력`` 이라고 적었다. 그 `out` 은 stdout 이다. REQ-DRW-004 는 "stdout 으로 내보내서는 안 된다(shall not)" 이다. 착지하면 모든 실제 `moai update` clean-reinstall 실행이 이 요구사항을 위반한다.
3. **AC-DRW-003 이 그 위반을 잡지 못한다.** plan R4 는 "판정은 기본 경로(주입 없음 = stderr)에서 이뤄진다" 로 판정을 기본 경로에 고정했는데, **그 기본 경로는 프로덕션에 존재하지 않는다.** AC-DRW-003 은 초록이고 프로덕션은 위반인 상태가 성립한다 — 이 감사가 지목받은 "존재 이유인 실패에서 통과하는 AC" 부류의 정확한 사례이며, 선행 카드에서 3회 반복된 형태다.

**Required fix** (세 가지 모두 필요):
- §A.4 표의 clean-reinstall 행을 실측값(`프로덕션 주입값 = cmd.OutOrStdout() = stdout`, `nil 기본값은 프로덕션 미도달`)으로 교체하고, §A.4 결론 문장("세 곳 중 두 곳은 이미 stderr")을 "init 한 곳만 stderr, update 두 곳은 stdout"으로 정정한다.
- plan §F M3 의 clean-reinstall 배선을 `opts.Out` 이 아닌 별도 stderr writer 로 바꾼다(`CleanReinstallOptions` 에 `ErrOut io.Writer` 를 더하고 `update.go:425`·`:627` 에서 `cmd.ErrOrStderr()` 를 주입하거나, 통지만 `os.Stderr` 로 보낸다). R4 의 "기본 경로에서 판정" 논거는 폐기한다.
- AC-DRW-003 의 Given 을 **`Out` 이 stdout 으로 주입된 프로덕션 배선 형태**로 못박는다. "주입 없음" 구성에서만 판정하면 이 결함이 그대로 재현된다.

### D2 — REQ-DRW-007(비비례성) 이 plan M1·AC-DRW-006 의 `failed` 설계와 모순하고, 어떤 AC 도 그 경로를 보지 않는다

- **위치**: `spec.md:108`(REQ-DRW-007), `spec.md:94`(§B.D4), `plan.md:47`(M1), `acceptance.md:49-55`(AC-DRW-004), `acceptance.md:69-73`(AC-DRW-006)
- **Severity: major — Class: blocking**

REQ-DRW-007 은 무한정이다: "통지의 길이는 미러 대상 스킬 수에 비례해 증가해서는 안 된다(shall not)."

그런데 plan §F M1 은 이렇게 설계한다: "`MirrorModeCopy` 개수를 담은 요약 1줄 + **`MirrorModeFailed` 항목의 경고 각 줄**." 그리고 AC-DRW-006 은 "해당 스킬의 실패 경고가 stderr 에 있어야 한다" 로 **스킬 단위 경고 방출을 적극적으로 요구한다**. 즉 `failed` 경로의 출력 길이는 스킬 수에 정비례한다.

이 경로는 예외적이지 않다. 실측 — `internal/template/skill_mirror.go:222-228`:

```go
if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
    return SkillMirrorEntry{Skill: skill, Mode: MirrorModeFailed, Warning: ...}
}
```

`.agents/skills` 디렉터리를 만들 수 없는 환경에서는 이 분기가 **스킬마다** 발동한다. 미러 대상은 임베드 스킬 전수이고(`skill_mirror.go:178-187` 는 필터 없이 `skills` 전건을 돈다), 실측 개수는 34다(`ls -d internal/template/templates/.claude/skills/*/ | wc -l` = 34). 결과는 **34줄** — §B.D4 가 "나가서는 안 된다"고 적은 바로 그 출력이다. `:239-244`(symlink·copy 동시 실패)도 같은 형태로 대량 발동 가능하다.

AC-DRW-004 는 이 경로를 보지 않는다. Given 이 "**폴백** 스킬 수가 2개인 배포와 34개인 배포"이고, 이 SPEC 에서 "폴백"은 일관되게 복사 폴백을 뜻한다(§A.1, §B.D4). 따라서 REQ-DRW-007 위반이 판정망 밖에 있다.

**Required fix** (택1):
- REQ-DRW-007 의 적용 범위를 **복사 폴백 통지**로 한정하고(§B.D4 의 실제 논거가 그렇다), `failed` 경고의 길이 정책을 별도 요구사항으로 명시한다. 그리고 그 정책에 대응하는 AC 팔을 세운다.
- 또는 `failed` 경고도 요약 형태로 상한을 두고, AC-DRW-004 에 "`failed` 34건 vs 2건의 줄 수 동일" 팔을 추가한다. 이 경우 AC-DRW-006 을 "실패 사실과 개수가 도달한다"로 함께 조정해야 모순이 남지 않는다.

### D3 — AC-DRW-003(스트림 규율) 이 호출부에 결박되지 않아, 위험 호출부를 비껴갈 수 있다

- **위치**: `acceptance.md:38-47`
- **Severity: major — Class: blocking**

AC-DRW-003 의 Given 은 "AC-DRW-001 과 동일한 폴백 주입 조건" 이고, AC-DRW-001 의 Given(`acceptance.md:22`)은 `template` 패키지 seam 또는 `internal/cli` 이중체 중 **아무 구성이나 허용**한다("두 구성 모두 판정에 동등하다"). 어느 호출부에서 판정할지는 지정되지 않았다.

그런데 "Never mix" 위반이 가능한 지점은 특정돼 있다 — update template sync(plan R3) 와 clean-reinstall(D1). AC-DRW-003 을 세 호출부 중 안전한 한 곳에서만 세우면, 위험 호출부의 stdout 오염은 검사되지 않는다. AC-DRW-007 이 update 두 팔의 도달 지점을 stderr 로 못박아(`acceptance.md:82`) "stdout **에만** 쓰는" 구현은 잡지만, **양쪽 스트림에 모두 쓰는 구현**은 AC-DRW-007 을 통과한다 — 그리고 그것이 AC-DRW-003 의 [HARD] 가 스스로 지목한 실패 형태다(`acceptance.md:47`).

**Required fix**: AC-DRW-003 의 Given 에 판정 호출부를 명시한다 — 최소한 update template sync 와 clean-reinstall **두 곳 각각**에서 stdout 음성 팔을 세운다(init 은 §A.3 실측대로 이미 stderr 전용 통로라 제외 가능).

### D4 — 오귀속 문구의 라인 인용이 틀렸다

- **위치**: `spec.md:72`(§A.6), `spec.md:116`(§D)
- **Severity: minor — Class: optional**

두 곳 모두 `skill_mirror.go:196` 을 지목한다. 실측하면 `:196` 은 주석("Stat follows the link…")이고, 문구 `a non-symlink entry already exists at … — left untouched` 는 **`:205`** 에 있다. `.moai/reports/t81/sync-audit.md:168`(F4)이 같은 오류를 갖고 있어 그대로 상속된 것으로 보인다. 실측 정확성을 자기 권위의 근거로 삼는 SPEC 이므로 정정 가치가 있다.

**Required fix**: 두 곳의 `:196` 을 `:205` 로 정정한다(선택적으로 sync-audit F4 의 오류를 상속했다는 사실을 부기).

### D5 — plan R1 이 도입하는 주입 seam 의 프로덕션 기본값을 지키는 AC 가 없다

- **위치**: `plan.md:13`(R1), `plan.md:60`(M3)
- **Severity: minor — Class: optional**

M3 는 `update_template_sync.go` 에 패키지 수준 배포기 주입 seam 을 새로 도입한다(저장소 관용 형태이며 정당화도 적혀 있다 — `userHomeDirFn`·`webRunFn`·`findProjectRootFn` 실재 확인). 다만 이 카드의 8개 AC 중 어느 것도 **프로덕션 기본 경로가 여전히 실제 배포기를 만든다**는 것을 단언하지 않는다. seam 기본값 오배선이나 테스트 간 상태 누수는 "배포가 조용히 아무것도 하지 않는" 형태로 나타나며, 통지 관련 AC 는 전부 이중체를 주입한 상태에서 판정하므로 이를 볼 수 없다.

**Required fix**: M3 닫힘 조건에 "seam 미주입 상태에서 `update_template_sync` 가 `NewDeployerWithRendererAndForceUpdate` 결과를 사용한다"는 단언을 1줄 추가하거나, 기존 template-sync 회귀 테스트가 이를 이미 덮는지 run-phase 에서 확인해 기록한다.

### D6 — §B.D3 이 `failed` 문구의 정확성을 1/3 표본으로 일반화한다

- **위치**: `spec.md:90`
- **Severity: minor — Class: optional**

§B.D3 은 `failed` 경고를 올리는 근거로 "문구가 정확함: `symlink and copy both failed`" 를 든다. 그 문구는 `skill_mirror.go:243` 한 분기의 것이고, `MirrorModeFailed` 는 세 분기에서 나온다 — `:217`(`cannot replace stale link`), `:226`(`cannot create .agents/skills`), `:243`. 나머지 둘도 사실 관계는 정확하므로 결론은 바뀌지 않으나, 근거 진술이 표본 하나에 기대고 있다.

**Required fix**: §B.D3 에 `failed` 분기가 3종임을 적고 셋 다 오귀속이 없음을 명시한다.

### D7 — AC-DRW-002 에 자체 위증 검사가 없다

- **위치**: `acceptance.md:30-36`
- **Severity: minor — Class: optional**

AC-DRW-002 는 부재 단언이므로 **기능이 통째로 없을 때도 통과한다**. SPEC 은 이를 인지하고 [HARD] 로 "AC-DRW-001 의 음성 팔이며 함께 읽는다" 라고 못박았고, AC-DRW-001 은 위증 검사(`acceptance.md:28` — 소비부 제거 시 붉어져야 함)를 갖는다. MUST 하나라도 실패하면 전체 FAIL 이므로 **판정 집합 수준에서는 방어가 성립한다** — 결함으로 승격하지 않는 이유다. 다만 AC-DRW-002 자체에는 위증 검사가 없어, 두 팔이 분리된 테스트로 구현되고 한쪽이 누락되면 규율이 조용히 약해진다.

또한 AC-DRW-002 의 판정 어휘("미러 관련 문자열")가 run-phase 로 위임돼 있어(`acceptance.md:36`) 현 상태로는 이진 판정 불가다. 위임 자체는 §D "구현 세부" 범위 밖 결정과 일관되며, "안정 부분식(예: 미러 경로 토큰)" 이라는 가드도 걸려 있다.

**Required fix**(선택): AC-DRW-002 에 위증 검사 1줄을 추가한다 — 예: "두 팔은 같은 테스트 함수의 서브테스트로 구성하며, 소비부 제거 시 AC-DRW-001 팔이 붉어지는 것으로 이 팔의 유효성을 담보한다."

## 2. 리드가 지목한 결함 부류별 판정

| 부류 | 판정 | 근거 |
|---|---|---|
| 실패할 수 없는 AC | **결함 발견 (D1)** — AC-DRW-003 의 판정 경로가 프로덕션 미도달. "전부 심볼릭 링크일 때 무출력"(AC-DRW-002) 은 **적정** — 위증 검사를 가진 AC-DRW-001 과 MUST 쌍으로 묶여 있다(D7 은 강화 제안). | §1·D1·D7 |
| 선택적 인터페이스 계약 | **적정** — REQ-DRW-005/006 + AC-DRW-005 가 3단언(오류 없음·panic 없음·통지 없음) + [HARD] "이중체는 `ResultDeployer` 를 **컴파일 시점에 만족하지 않아야**". 하드 타입 단언 panic 은 2번, 배포 침묵은 1번이 잡는다. `DeployWithResult` 를 정의하고 nil 을 돌려주는 회피 형태도 [HARD] 가 봉쇄. | `acceptance.md:57-67` |
| 출력 분량·스트림 | **부분 결함 (D1·D2·D3)** — 형태(요약)·스트림(stderr) 결정과 정당화는 있으나, 스트림 실측이 틀렸고(D1) `failed` 분량이 미결(D2)이며 스트림 AC 가 호출부에 결박되지 않았다(D3). | §B.D4/D5, REQ-004/007 |
| 오귀속 문구 | **적정 (강점)** — §A.6 실측 → §B.D3 결정 → REQ-DRW-009 → AC-DRW-008 → §D 범위 밖 → §D.6 전방 점검까지 6단으로 명시했고, 승계 카드에서 뒤집히는 것이 정상임까지 적었다. 무언의 통과가 아니다. | `spec.md:70-72,86-90,110,114-117`, `acceptance.md:86-92,119` |
| 범위 규율 | **적정** — REQ-DRW-006 이 `Deploy` 시그니처 변경·`ResultDeployer` 필수화를 금지하고, AP-4/AP-5/AP-8 이 미러 로직·비공개 seam 승격·`internal/template` 동시 수정을 봉쇄한다. plan §D 에 "`internal/template` 무변경" 명시. 유일한 프로덕션 구조 변경은 R1 의 주입 seam 이며 정당화돼 있다(D5 는 그 seam 의 기본값 가드 부재). 청소 경로 언급 0건. | `spec.md:107`, `plan.md:28-29,75-82` |

## 3. 다른 문서의 주장과 충돌하는 것 (별도 요청 항목)

1. **본 SPEC §A.4 ↔ 실제 코드 — 충돌 (D1).** "세 곳 중 두 곳은 이미 경고를 stderr 로 낸다"는 거짓. 실측상 stderr 는 init 한 곳뿐이다.
2. **본 SPEC REQ-DRW-007 ↔ 본 SPEC AC-DRW-006 / plan M1 — 충돌 (D2).** 전자는 스킬 수 비례 금지, 후자 둘은 `failed` 스킬 단위 방출 요구.
3. **본 SPEC §A.6·§D ↔ 실제 코드 — 라인 인용 불일치 (D4).** `:196` vs 실제 `:205`. **t81 sync-audit F4(`sync-audit.md:168`)도 같은 오류를 갖고 있다** — 원본 감사의 결함이며 이 SPEC 이 상속했다.
4. **선행 SPEC 과의 충돌: 없음.** `REQ-CSC-005`(`SPEC-CODEX-SKILLS-CANONICAL-001/spec.md:201` — "배포기 내부에서 직접 출력해서는 안 된다") 를 본 SPEC §D 가 정확히 인용·존중한다. `AC-CSC-006`(반환값 기준 판정, `acceptance.md:70-76`) 을 §B.D2 가 "의도된 계약이므로 재론하지 않는다" 로 보존한다. `AC-CSC-010` 회귀 요구도 §E·plan M4 에 유지.
5. **t81 sync-audit 과의 충돌: 없음** (위 3번의 라인 오류 상속을 제외하면). F1(`sync-audit.md:154-160`, 실측 PROBE 출력 보유) → §A.5, F2(`:162`) → §A.1, F4(`:168`) → §A.6 로 각각 정확히 대응한다. F2 의 Required fix("`internal/cli` 의 init/update 경로에서 승격") 를 이 카드가 그대로 수행한다.
6. **선행 merge SHA `23ffb43a1` — 검증됨.** `git log --oneline -1 23ffb43a1` → `merge: skill mirror at .agents/skills for Codex CLI (card t81)`.
7. **`34건` 숫자 — 검증됨.** `internal/template/templates/.claude/skills/` 하위 디렉터리 34개. (로컬 `.claude/skills/` 는 44개지만 이는 배포 대상이 아닌 로컬 전용 `hns-*` 등을 포함한 값이며, 미러 대상은 배포 walk 에서 파생되므로 34가 맞다.)
8. **`CHANGELOG.md` 인용 — 검증됨.** `CHANGELOG.md:56` 에 "The fallback warning does not currently reach you." 문장 실재.

## Recommendation

blocking 3건(D1·D2·D3)을 닫은 뒤 iter-2 를 요청한다. optional 4건(D4~D7)은 오케스트레이터 재량이며, 이들만으로는 FAIL 을 정당화하지 않는다.

1. **D1 (최우선)** — §A.4 표의 clean-reinstall 행과 §A.4 결론 문장을 실측값으로 정정한다: 프로덕션 `Out` 은 `update.go:154` 의 `cmd.OutOrStdout()` 이며 `nil ⇒ os.Stderr` 분기는 프로덕션 미도달. 이어 plan §F M3 의 clean-reinstall 출력 대상을 `opts.Out` 에서 분리된 stderr writer 로 바꾸고, plan R4 의 "기본 경로에서 판정" 논거를 폐기한다.
2. **D1 (연동)** — AC-DRW-003 의 Given 을 "`Out` 이 stdout 으로 주입된 상태" 로 못박는다. 이 한 줄이 없으면 정정된 설계가 다시 판정망 밖으로 나간다.
3. **D3** — AC-DRW-003 에 update template sync · clean-reinstall **두 호출부 각각**의 stdout 음성 팔을 세운다.
4. **D2** — REQ-DRW-007 을 복사 폴백 통지로 한정하고 `failed` 경고의 길이 정책을 별도로 명시하거나, `failed` 도 요약화하고 AC-DRW-004 에 `failed` 팔을 추가한다(후자 선택 시 AC-DRW-006 문구 동반 조정 필요).
5. **범위 결정은 유지한다** — init 포함(D1 결정)은 §A.3 실측이 전건 정확하고, 감사가 추가 검증한 두 링크(executor 포인터 관통, `Collect`→`emitSummary` stderr 렌더)도 성립한다. 이 부분은 재작업 대상이 아니다.

## 감사 방법 (Gaps / Residual-risk)

- **실행한 것**: 읽기 전용 grep/awk/sed 실측 + `git log`/`git rev-parse`. SPEC artifact 편집 0건, 임시·probe 파일 생성 0건, 트리 상태 감사 전후 동일(clean).
- **Gaps (관측하지 않은 것)**: Go 테스트를 실행하지 않았다(현 SPEC 은 plan-phase 이며 구현이 없다). `golangci-lint`·`go vet` 미실행. 실제 Windows/권한 없는 환경에서의 폴백 재현 미수행 — D2 의 "34줄" 은 코드 경로 분석(`skill_mirror.go:222-228` 이 스킬마다 발동) + 스킬 수 실측(34)으로 도출한 것이지 실행 관측이 아니다.
- **Residual-risk**: D1 의 `out` 추적은 `update.go` 단일 파일 내 어휘 범위 분석(`out :=` 3곳 중 유효 정의 특정)에 근거한다. 클로저 캡처를 통한 재정의는 grep 으로 확인되지 않았으나 `:154`~`:696` 사이에 `out` 재선언이 없음은 확인했다. 반증하려면 `update_clean_install.go` 안에서 `out` 에 실제로 쓰이는 값을 런타임으로 관측해야 한다.
