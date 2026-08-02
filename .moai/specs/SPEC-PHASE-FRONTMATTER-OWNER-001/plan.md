# SPEC-PHASE-FRONTMATTER-OWNER-001 — 구현 계획

## §A 맥락

- **작업 위치**: 격리 워크트리 `.claude/worktrees/phase-frontmatter/`(브랜치 `spec/phase-frontmatter-owner`). run-phase 진입 시 base 상태는 §C 절차로 **재확인**한다 — 이 문장을 근거로 확인을 생략하지 않는다.
- **Tier**: S (2 산출물 — spec.md / plan.md, AC는 spec.md §3에 인라인, + progress.md)
- **REQ / AC**: REQ 7개 + NFR 1개 = **8** / AC **8**. 양쪽 모두 Tier S 상한 8에 정확히 도달했다 — 추가하려면 기존 항목에 접거나 tier-up을 검토해야 한다.
- **성격**: 산출물이 **전부 문서**다. Go 코드 변경은 0이며, 실행하는 유일한 Go 명령은 템플릿 중립성 가드 테스트다.

### §A.1 범위 축소 — 이 계획은 v0.4.1 계획의 후속이 아니라 대체다

병렬 SPEC `SPEC-PHASE-FIELD-VALIDATION-001`(커밋 `998744216`, PR #1285)이 run-phase 도중 `origin/main`에 착지해 M4(lint 규칙)와 M5(9개 실물 정정)를 먼저 구현했다. 두 마일스톤과 그것만을 위해 존재하던 REQ/AC는 삭제되었다(spec.md HISTORY v0.5.0, §1.12).

**남은 작업의 실제 크기**: 문서 2개 + 미러 2개 = 4파일. 그중 **M1~M3은 이미 이 브랜치에 착지해 있고**, 미착지 작업은 **M7 한 건**이다(§F). 이 비대칭이 계획의 형태를 정한다 — 새로 만드는 것이 아니라 이미 만든 것 하나를 고친다.

## §B 알려진 이슈

- **B1 — `origin/main`이 전진했다.** 이 브랜치의 base는 낡았다. `origin/main`은 `998744216`(형제 SPEC)을 포함하며 그 커밋은 `internal/spec/lint.go` + `internal/spec/lint_phase_test.go`를 건드린다. 이 브랜치의 (삭제 예정) M4 커밋도 **같은 경로**를 건드리므로 리베이스/머지 시 하드 충돌이 확정이다 — 그 해소는 오케스트레이터의 브랜치 수술 소관이며 이 계획의 범위 밖이다.
- **B2 — 미러 패리티는 현재 성립하나 조건부다.** 대상 2쌍은 지금 `diff` 0행이다(spec.md §1.9). 리포에는 §25 중립화로 **의도적으로 divergent한** 미러 쌍이 실재하므로(`workflows/plan.md`, 6행), "모든 미러를 동일하게"로 일반화하면 안 된다.
- **B3 — 이 셸의 `ls` alias.** `ls | grep '^name'`은 항상 0매칭이다. 파일 존재 판정은 `git ls-files` 또는 glob으로 한다.
- **B4 — 공유 체크아웃 레이스.** 이 리포는 병렬 세션이 흔하다. 커밋 직전 `git rev-parse --short HEAD` + `git branch --show-current`를 재확인하고, `git add -A` 대신 pathspec으로 스테이징한다.
- **B5 — `grandfather` 토큰은 스키마 문서에 이미 존재한다.** Status Enum 절의 "grandfathered SPECs that carry `status: planned`" 때문에 `grep -ci grandfather`는 무편집 상태에서 이미 `1`이다(실측). AC-PFO-016 (d)가 `eraDemotableCodes`만 겨냥하는 이유이며, 판정을 임의로 `grandfather`로 넓히면 죽은 검사가 된다.

## §C 사전 점검

run-phase 진입 직후, 편집 전에 실행한다.

```bash
# C1. base 재확인 (저작 시점 기록을 신뢰하지 않는다)
git branch --show-current; git rev-parse --short HEAD
git fetch origin main && git rev-list --count --left-right origin/main...HEAD

# C2. 미러 존재 + 현재 패리티 (B2)
git ls-files 'internal/template/templates/**/spec-frontmatter-schema.md' \
             'internal/template/templates/**/spec-assembly.md'
diff .claude/rules/moai/development/spec-frontmatter-schema.md \
     internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md | wc -l

# C3. 착지한 가드의 실제 동작 재확인 (문서가 기술할 대상)
git show origin/main:internal/spec/lint.go | grep -n 'phaseWorkflowStageTokens\[strings.ToLower'
git show origin/main:internal/spec/lint.go | grep -n -A4 'var eraDemotableCodes'

# C4. AC baseline 재계측 (spec.md §3의 "현 워크트리 관측"이 여전히 유효한지)
S=.claude/rules/moai/development/spec-frontmatter-schema.md
grep -c -i 'case-sensitiv' "$S"; grep -c -i 'case-insensitiv\|case-fold' "$S"
grep -c 'FrontmatterPhaseInvalid' "$S"; grep -c 'eraDemotableCodes' "$S"
```

**C3이 이 SPEC의 유일한 실질 의존성이다.** 문서가 기술할 대상이 코드이므로, 코드를 읽지 않고 문서를 쓰면 §1.13이 진단한 오기를 반복한다.

## §D 제약

- **PRESERVE**: `internal/spec/` 전체. 이 SPEC은 Go 코드를 **읽기만** 한다. 가드의 매처·심각도·`eraDemotableCodes` 구성은 형제 SPEC 소유이며 변경 금지(spec.md §4).
- **의무**: 문서 수정 시 미러를 **같은 커밋에서 함께** 수정하고 `make build`를 실행한다(NFR-PFO-001, CLAUDE.local.md §2).
- **금지 (이 SPEC이 편집하는 2개 파일에 한정)**: `spec-frontmatter-schema.md`와 `plan/spec-assembly.md`의 미러를 로컬과 다른 내용으로 두는 것. 이 두 쌍에 한해 로컬만 고치는 것도, 미러만 고치는 것도 AC-PFO-014 FAIL이다.

  > **바이트 동일은 리포 일반 규칙이 아니다.** 위 금지는 이 SPEC이 편집하는 2개 파일에만 적용된다. 일반 의무는 "같은 **의미의** 편집을 양쪽에 적용하되 각 사본의 §25 중립화를 보존한다"이며 무조건 동일화가 아니다(B2, spec.md §1.9).
- **금지**: 정정 문구에 SPEC ID·REQ 토큰·내부 날짜·commit SHA를 담는 것. 이 문구는 **미러에 그대로 들어가야** 하므로 §25 중립성이 로컬 사본까지 구속한다 — 로컬에만 SPEC ID를 넣으면 패리티가 깨진다.
- **금지**: `--no-verify`, `--amend`, main 직접 push.
- **범위 규율**: 스키마 문서에서 바꾸는 것은 **§ Prohibited phase values 절의 매칭 서술 문단 하나**다. 인접 문단(정의·금지 열거·era 결속)은 이미 옳으므로 손대지 않는다.

## §E 자기 검증

각 마일스톤 종료 시 spec.md §3의 해당 AC 명령을 실행하고, **명령 + 축약 없는 출력**을 `progress.md` §E.2에 기록한다. 요약된 증거("통과함")는 증거가 아니다.

## §F 마일스톤

### M7 — 값 계약이 착지한 가드를 정확히 기술하도록 정정 (유일한 미착지 작업)

REQ-PFO-016 · AC-PFO-016 / AC-PFO-014

이것이 남은 유일한 실질 작업이다. M1~M3은 이미 이 브랜치에 착지해 있고 그 AC(001~004·006·007)는 현재 PASS다. 번호를 M4로 되쓰지 않고 M7을 쓰는 이유는 M4·M5가 **삭제된 마일스톤**이지 이름이 재사용된 마일스톤이 아니기 때문이다 — progress.md와 커밋 이력에 남은 M4 언급이 계속 lint 규칙을 가리키도록 보존한다.

#### [산출물 요건 — 정확한 최종 상태]

`.claude/rules/moai/development/spec-frontmatter-schema.md` § Prohibited phase values 절에서 **다음 한 문단을 정확히 찾아 교체한다.**

**제거 대상 (현재 존재, 리터럴):**

```
The prohibition binds the whole value, case-sensitively. A legitimate label that happens to contain one of these words as a substring — `"v3.0.0 — Phase 2 — Runtime Hardening"`, or a milestone label naming a sync layer — is valid and MUST NOT be rejected.
```

**교체 후 (두 문단, 리터럴):**

```
The prohibition binds the whole trimmed value, compared case-insensitively — `PLAN`, `Sync`, and a value padded with surrounding whitespace are rejected exactly as `plan` is. It does NOT bind substrings: a legitimate label that merely contains one of these words inside a longer phrase — `"v3.0.0 — Phase 2 — Runtime Hardening"`, or a milestone label naming a sync layer — is valid and MUST NOT be rejected. Whole-value comparison is what makes case-insensitivity safe here; substring matching would false-flag those labels.

Enforcement lives in `internal/spec/lint.go` `FrontmatterSchemaRule`, which emits finding code `FrontmatterPhaseInvalid` at error severity. That code is deliberately absent from `eraDemotableCodes`, so it is NOT demoted to an advisory warning on grandfather-era SPECs: the guard exists to catch an authoring mistake on an in-flight SPEC, and the era heuristic classifies almost every in-flight SPEC as grandfathered.
```

교체 후 이 절의 나머지 문단(정의 문단, 금지 열거 문단, `**This field is load-bearing.**` era 결속 문단)은 **변경하지 않는다.**

#### [미러 요건]

`internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md`에 **바이트 동일한 교체**를 적용한다. 위 교체 문구는 SPEC ID·REQ 토큰·내부 날짜·commit SHA를 담지 않으므로 §25 중립성을 이미 만족하며, 따라서 두 사본은 동일해야 한다(AC-PFO-014).

`spec-assembly.md`와 그 미러는 **이 마일스톤에서 변경하지 않는다** — M3이 이미 착지했고 AC-PFO-004가 현재 PASS다.

#### [검증 순서]

```bash
# 1. AC-PFO-016 (4항 전부)
S=.claude/rules/moai/development/spec-frontmatter-schema.md
grep -c -i 'case-sensitiv' "$S"                  # 기대 0   (교체 전 1)
grep -c -i 'case-insensitiv\|case-fold' "$S"     # 기대 >=1 (교체 전 0)
grep -c 'FrontmatterPhaseInvalid' "$S"           # 기대 >=1 (교체 전 0)
grep -c 'eraDemotableCodes' "$S"                 # 기대 >=1 (교체 전 0)

# 2. make build (템플릿 임베드 재컴파일)
make build

# 3. AC-PFO-014 패리티 + 중립성
diff "$S" internal/template/templates/"$S" | wc -l | tr -d ' '     # 기대 0
diff .claude/skills/moai/workflows/plan/spec-assembly.md \
     internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md | wc -l | tr -d ' '   # 기대 0
go test -count=1 -run 'TestInternalContentLeak|TestTemplateNeutrality' ./internal/template/

# 4. 이미 착지한 6개 AC 재확인 (회귀 없음)
#    AC-PFO-001 / 002 / 003 / 004 / 006 / 007 — spec.md §3의 명령을 그대로 실행
```

**4번을 생략하지 않는다.** M7의 편집은 AC-PFO-002가 `sed` 범위를 앵커하는 `Prohibited phase values` 절 **안**에서 일어나므로, 절 제목이나 4토큰을 실수로 건드리면 AC-PFO-002가 조용히 깨진다.

### M8 — 회귀 검증

NFR-PFO-001 · AC-PFO-014

```bash
go build ./...
go test -count=1 ./internal/template/
git diff --stat -- .claude/ internal/template/    # 2파일이어야 한다 (로컬 + 미러)
```

Go 코드를 변경하지 않았으므로 전체 테스트 스위트는 필요 없다. 변경 표면이 템플릿 임베드이므로 `internal/template/`만 돌린다.

## §G 안티패턴

- **AP-1 — 가드 매처를 문서에 재구현 수준으로 옮기기.** 정정의 목적은 **가리키는 것**이지 복제하는 것이 아니다. `phaseWorkflowStageTokens` 맵 내용이나 Go 조건식을 문서에 옮기면 형제 SPEC과 중복 원천이 되어, 그쪽이 바뀔 때 이쪽이 조용히 낡는다(spec.md §4).
- **AP-2 — 오기 문장을 삭제만 하고 대체 서술을 쓰지 않기.** AC-PFO-016 (a)는 삭제만으로도 PASS한다. 정정은 교체이며 (b)(c)(d)가 그것을 강제한다.
- **AP-3 — 정정 문구에 SPEC ID를 넣기.** 로컬에만 넣으면 미러 패리티가 깨지고(AC-PFO-014 FAIL), 미러에도 넣으면 §25 중립성 CI 가드에 걸린다. SPEC ID 수준 cross-reference는 이 SPEC 본문(§1.12)에만 둔다.
- **AP-4 — 판정을 `grandfather`로 넓히기.** 그 토큰은 무편집 상태에서 이미 `1`이므로 죽은 검사가 된다(B5). `eraDemotableCodes`가 판별력을 가진 유일한 토큰이다.
- **AP-5 — M1~M3 산출물을 "다시 확인한다"며 재작성하기.** 이미 착지했고 AC가 PASS다. 재작성은 §D 범위 규율 위반이며 AC-PFO-002의 앵커를 깨뜨릴 위험만 만든다.
- **AP-6 — 삭제된 M4·M5를 부활시키기.** lint 가드와 9개 정정은 형제 SPEC 소유다. "우리 것도 있으면 좋겠다"는 방향은 가드 이중화이며 spec.md §4가 명시적으로 배제한다.
- **AP-7 — `make build`를 건너뛰기.** 템플릿은 `//go:embed`로 바이너리에 컴파일되므로, 미러 파일만 고치고 빌드하지 않으면 배포 산출물에 반영되지 않는다.

## §H 교차 참조

- `.claude/rules/moai/development/spec-frontmatter-schema.md` — M7 편집 대상 (로컬)
- `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` — M7 편집 대상 (미러)
- `.claude/skills/moai/workflows/plan/spec-assembly.md` + 미러 — M3에서 이미 착지, M7에서 변경 없음
- `internal/spec/lint.go` `FrontmatterSchemaRule` — 문서가 기술할 대상 (**읽기 전용, 변경 금지**)
- `SPEC-PHASE-FIELD-VALIDATION-001` — 가드 소유 SPEC (spec.md §1.12)
- `.moai/docs/template-internal-isolation-doctrine.md` §25 — 미러 중립성 규약

## §I 미해결 질문 요약

**미해결 항목 없음 (0건).**

| 항목 | 결정 | 근거 |
|---|---|---|
| 가드 매처의 매칭 축 | 대소문자 **무시** + 값 전체 일치 (착지한 구현) | spec.md §1.8 — 결정 축은 부분 문자열 대 값 전체 일치이며, 값 전체 일치 위에서 대소문자 구분은 판별력이 없다 |
| 가드 소유권 | `SPEC-PHASE-FIELD-VALIDATION-001` | spec.md §1.12 — 이 SPEC은 위층(계약·소유권·문서 정확성)만 다룬다 |
| 배포 문서의 가드 참조 형태 | **코드 위치**(`internal/spec/lint.go` + finding 코드), SPEC ID 아님 | spec.md §1.9 — SPEC ID는 §25 중립성 때문에 미러에 들어갈 수 없어 패리티를 깨뜨린다 |
| `acceptance.md` 처분 | **삭제** — Tier S는 2 산출물이며 AC는 spec.md §3에 인라인 | Tier S 산출물 계약. 파일을 남기면 소유자 없는 고아가 된다 |
