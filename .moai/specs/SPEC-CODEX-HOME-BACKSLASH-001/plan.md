# SPEC-CODEX-HOME-BACKSLASH-001 — 구현 계획

## A. 컨텍스트

Tier S. 함수 하나(`classifyCodexSkillPath`), 파일 하나(`internal/cli/doctor_codex.go`), 그리고 테스트. 산출물이 작다는 것이 등급을 정하지, 결함의 무게가 정하지 않는다 — 다만 보안 축이므로 AC는 별도 `acceptance.md`로 분리했다(Tier S의 통상 인라인 배치에서 벗어난 지점, 카드 지시).

기준선: 워크트리 `.claude/worktrees/t571`, 브랜치 `WT-codex-home-backslash`, HEAD `ee194493f`. 선행 t540은 로컬 develop 착지·원격 미착지(§spec.md HISTORY).

## B. 알려진 이슈

- **선행 미착지.** 이 브랜치는 `origin/develop`(`3ac58b5a1`)에서 잘려 로컬 develop 을 흡수했다. 병합 창에서 `origin/develop`을 다시 흡수하면 t540 과의 충돌 가능성이 있다 — 같은 함수를 만지기 때문이다. 흡수 후 **병합 트리에서 재측정**한다.
- **부재 가드의 함정.** REQ-CHB-002 를 「`~/x\SKILL.md`가 oddly-formed 인가」로만 검사하면, 나중에 누구든 분류기를 「전부 oddly-formed」로 망가뜨려도 통과한다. 그래서 구속력 있는 AC 는 **두 팔 대칭**이다(AC-CHB-001).

## C. 사전 점검

1. `git -C . rev-parse --short HEAD` — 편집 직전 재판독.
2. `.moai/reports/t571/repro-asymmetry.log` 존재 확인 — 기준선 인용의 출처.
3. `grep -n "classifyCodexSkillPath" internal/cli/*.go` — 소비자 2곳이 여전히 그 둘인지 확인.

## D. 제약

- 시접(seam)은 **기존 것만** 쓴다: `osStatFn`, `codexUserHomeDir`, `configPathSeparator`. 새 시접을 만들지 않는다.
- 시접을 덮는 테스트는 `t.Parallel()` 금지 + `t.Cleanup` 복원(REQ-CHB-007) — 이 패키지의 기존 규율(`codex_skills_prune_test.go:10`, `codex_config_path_test.go:14`).
- 분기 순서 재배치 외의 로직 변경 금지. 새 shape 상수 추가 금지.
- 검증 범위는 `./internal/cli/...`. 전체 스위트 로컬 실행 금지.
- **[HARD] 인접 소스-텍스트 가드 — `TestSinglePathShapeClassifier`.** `internal/cli/codex_skills_prune_test.go:606` 에 이미 있는 가드가 M2 의 편집 자유도를 좁힌다. 그 가드는 `internal/cli` 와 `internal/codexwiring` 아래의 **비-테스트** `.go` 파일을 전부 훑어, 한 파일이 리터럴 `filepath.IsAbs` 와 리터럴 `HasPrefix(p, "~` 를 **둘 다** 담고 있으면서 `func classifyCodexSkillPath` 는 담고 있지 **않으면** 「두 번째 shape 분류기」로 판정해 실패시킨다(면제 조건은 오직 같은 파일에 그 함수 정의가 있는 것뿐이다).
    - 따라서 다음 두 가지가 금지된다: (a) 분류기를 별도 헬퍼 파일로 **추출**하는 것 — 추출된 파일이 두 리터럴을 갖고 함수 정의는 원래 자리에 남으면 즉시 붉어진다. (b) `strings.HasPrefix(p, "~/")` 의 **철자를 바꾸는 것**(예: `p[0] == '~'`, `strings.HasPrefix(p, string(homeMarker))`) — 가드는 의미가 아니라 **문자열**을 보므로, 철자를 바꾸면 면제 판정이 아니라 매치 실패 쪽이 무너져 다른 파일이 걸리거나 이 파일이 통과하는 식으로 의도와 어긋난다.
    - §F M2 의 코드 블록은 두 리터럴을 원래 자리·원래 철자로 유지하므로 이 가드를 깨지 않는다(실측 확인). 이 제약은 **구현자가 그 사실을 모른 채 리팩터링하는 것**을 막기 위해 적는다 — 어기면 AC-CHB-006 의 패키지 회귀에서 늦게 잡히고, 실패 사유가 이 카드와 무관해 보인다.

## E. 자기 검증

`acceptance.md`의 AC-CHB-001 ~ 009 각각이 실행 가능한 명령과 엄격한 통과 기준을 갖는다. 뮤턴트 AC 는 둘이며 서로 다른 것을 세운다: AC-CHB-002(옛 순서 복원)가 RED 를 세우지 못하면 두 팔 테스트가 공허하고, AC-CHB-009(`IsAbs` 위로 백슬래시 검사 올리기)가 RED 를 세우지 못하면 AC-CHB-003 의 `/tmp/a\b` 칸이 공허하다. 어느 쪽이든 공허하면 되돌아간다.

## F. 마일스톤

변경 가능성이 높은 결정부터 둔다.

### M1 — 두 팔 판별식을 먼저 세운다 (RED)

새 테스트 파일 `internal/cli/codex_skills_path_shape_test.go`. 표 기반으로 두 팔(`~/x\SKILL.md`, `x\SKILL.md`)의 shape 이 **서로 같고** 그 공통 값이 `codexPathOddlyFormed` 임을 단언한다. 보존 칸(§spec.md 3.3의 네 항목)도 같은 표에 넣는다. 이 시점에서 두 팔 단언은 RED 여야 한다 — RED 를 관측한 출력을 증거로 남긴다.

### M2 — 순서 재배치 (GREEN)

`classifyCodexSkillPath`에서 `ContainsRune('\\')` 검사를 독립 분기로 올린다:

```go
if filepath.IsAbs(p) { return codexPathAbsolute }
if strings.ContainsRune(p, '\\') { return codexPathOddlyFormed }
if p == "~" || strings.HasPrefix(p, "~/") { return codexPathHomeRelative }
if strings.HasPrefix(p, "~") { return codexPathOddlyFormed }
return codexPathRelative
```

함수 위 doc 주석의 「ordering is load-bearing」 문단을 새 순서로 고쳐 쓴다 — 주석이 옛 순서를 계속 주장하면 그것 자체가 다음 독자를 오도한다.

### M3 — 소비자 판정 고정

`judgeCodexSkillEntry(~/x\SKILL.md)` → `Eligible=false`, skip 사유가 oddly-formed 를 말하고, **`osStatFn` 호출 0회**. 호출 계수는 시접 오버라이드로 관측한다.

### M4 — 뮤턴트 2종 (한 배타 창, 하나씩)

공유 트리에서의 고의 훼손이므로 **하나의 배타 창** 안에서, 각 주입 **직전** 트리 상태(`git status --short`, `git rev-parse --short HEAD`)를 재확인하고 수행한다. 두 뮤턴트를 동시에 주입하지 않는다 — 주입 → 관측 → 복구 → 다음.

1. **옛 순서 복원**(AC-CHB-002) — M1 의 두 팔 테스트가 RED 로 가는지 본다. `.moai/reports/t571/mutant-old-ordering.log`.
2. **`IsAbs` 위로 백슬래시 검사 올리기**(AC-CHB-009) — `TestCodexSkillPathPreservedShapes` 가 `/tmp/a\b` 칸에서 RED 로 가는지 본다. `.moai/reports/t571/mutant-isabs-hoist.log`.

두 번째가 없으면 §spec.md 3.3 불변식 1 과 §G 의 두 번째 안티패턴은 아무것도 관측하지 않은 주장으로 남는다. 각 복구 후 해당 테스트를 다시 돌려 `EXIT=0` 을 보이고, 이 창에서 만든 바이너리는 귀속 불가이므로 폐기한다.

### M5 — 회귀 범위

`go test ./internal/cli/... -count=1 -timeout 600s`. SKIP 계수를 `.moai/reports/t571/skip-baseline-pre-m2.log` 에 고정된 사전 측정값과 **정확히 같은지** 비교한다(AC-CHB-006).

## G. 안티패턴

- 확장된 홈에 백슬래시 검사를 거는 것(Windows 파괴 — §spec.md 3.2).
- `IsAbs` 위로 백슬래시 검사를 올리는 것(§spec.md 3.3 불변식 1). AC-CHB-003 의 POSIX-절대-백슬래시 칸이 이것을 잡는다 — **그리고 그 칸이 실제로 잡는다는 것은 AC-CHB-009 의 뮤턴트가 세운다**(그 뮤턴트 없이는 「잡는다」가 관측되지 않은 주장이다).
- 두 팔 중 한 팔만 단언하는 AC.
- 백슬래시 경로를 정규화해 「구제」하려는 시도(범위 밖).

## H. 교차 참조

- `SPEC-CODEX-SKILL-PATH-SLASH-001`(t540) — 선행. 같은 함수 인접.
- `internal/cli/codex_skills_disable.go:261` — 대칭의 기준선이 되는 쓰기 쪽 가드.
- `.moai/reports/t571/repro-asymmetry.log` — 기준선 증거.
