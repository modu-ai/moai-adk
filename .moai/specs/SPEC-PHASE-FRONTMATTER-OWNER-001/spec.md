---
id: SPEC-PHASE-FRONTMATTER-OWNER-001
title: "phase 프론트매터 값 계약 확립 — 저작 시점 방지와 사후 정정 소유자"
version: "0.5.0"
status: in-progress
created: 2026-08-02
updated: 2026-08-02
author: manager-spec
priority: P2
phase: "v3.0.2"
module: ".claude/rules/moai/development"
lifecycle: spec-anchored
tags: "spec-lint, frontmatter, phase, era-classification, ownership, authoring-guard"
tier: S
related_specs: [SPEC-PHASE-FIELD-VALIDATION-001]
---

# SPEC-PHASE-FRONTMATTER-OWNER-001 — phase 프론트매터 값 계약 확립

## HISTORY

| 날짜 | 버전 | 변경 |
|---|---|---|
| 2026-08-02 | 0.5.0 | **범위 축소 개정 (사용자 명시 승인).** 병렬 SPEC `SPEC-PHASE-FIELD-VALIDATION-001`(커밋 `998744216`, PR #1285)이 run-phase 도중 `origin/main`에 착지해 이 SPEC의 M4·M5를 먼저 구현했다. 세 가지를 한다. **(1) M4·M5 제거** — 기계 가드(REQ-PFO-008~011·013)와 9개 실물 정정(REQ-PFO-012)은 중복이므로 삭제하고 가드의 소유권을 형제 SPEC에 넘긴다(§1.12). **(2) §1.8 정정** — §1.8은 "대소문자 무시 매칭이 실물 8건에 즉시 오탐을 낸다"고 단언했으나 **그것은 부분 문자열 매칭과 결합했을 때만 참이다.** 착지한 매처는 대소문자 무시 + **값 전체 일치**여서 오탐이 0이고, 이 SPEC이 명세한 대소문자 구분 매처보다 오히려 **엄격하다**. 결정 축은 대소문자가 아니라 **부분 문자열 대 값 전체 일치**였다(실측 재확인, §1.8). **(3) 값 계약이 실제 배포물을 기술하도록 정정** — 워크트리에 이미 적용된 스키마 문서 편집이 "case-sensitively"라고 적어 **착지한 가드를 오기(誤記)하고 있다.** 문서가 하중 필드를 잘못 기술하는 것을 고치는 SPEC이 스스로 그러면 안 된다(REQ-PFO-016, §1.13). 결과: **Tier M → Tier S**, REQ 16 → **8**, AC 16 → **8**, `acceptance.md`는 삭제하고 AC를 §3에 인라인한다. |
| 2026-08-02 | 0.4.1 | plan-auditor 4차 PASS 0.86 이후 SHOULD-FIX 2건 적용. (S1) AC-PFO-009 (4)의 `== 3`이 올바른 구현을 FAIL시킬 수 있어 `>= 3`으로 완화. (S2) 모집단 `563` 리터럴이 판정값으로 쓰이는 4곳에서 리터럴 제거 — 이 SPEC 산출물이 커밋되면 모집단이 증가해 자기 판정을 깨뜨린다. *(v0.5.0 주: 두 항목 모두 이번 개정에서 삭제된 M4·M5 소관이었다.)* |
| 2026-08-02 | 0.4.0 | plan-auditor 3차 FAIL 0.78 대응, D1~D5. (D1) AC-PFO-011의 제목·GWT가 같은 AC의 판정 명령을 정면으로 반박. (D2) AC-PFO-009가 픽스처 **실재**만 판정하고 기대 방향을 검증하지 않는 거짓 보증문 동반. (D3) AC-PFO-008/010/011의 판정 대상이 M5가 정정하는 9개 중 하나여서 DoD 시점에 통과 불가능. (D4) 명명 앵커 3종이 산출물 요건으로 부재. (D5) AC-PFO-002 첫 명령의 baseline `0`이 거짓(실제 `4`)이고 무편집 상태에서 이미 PASS하는 죽은 검사. |
| 2026-08-02 | 0.3.0 | plan-auditor 2차 FAIL 0.76 대응. 회귀의 단일 원인은 v0.2.0이 새로 도입한 N1 — `--strict`는 이 코드베이스에서 `Finding.Severity`를 재작성하지 않으므로 v0.2.0의 `grep -c 'ERROR'` 조건은 올바른 구현으로도 영구 FAIL인 통과 불가능 판정이었다. 인용했던 거짓 선례 2건을 대체 없이 삭제. |
| 2026-08-02 | 0.2.0 | plan-auditor 1차 FAIL 0.79 대응. 정당 라벨 수를 `563 − 9 = 554`에서 **546**으로 정정(키 부재 8개를 정당 라벨로 계상하고 있었다). 템플릿 미러 2개가 실재하며 바이트 동일함을 확인해 NFR-PFO-001을 "템플릿 무변경"에서 "미러 동시 수정"으로 완화. 매칭 방식 클라리피케이션의 전제가 거짓이었음을 §1.8에 기록. |
| 2026-08-02 | 0.1.0 | 최초 작성. **최초 진단을 재구성한다.** 이 문제는 "`phase:` 전이를 수행할 소유자가 없다"로 처음 지목되었으나, `phase:`는 애초에 라이프사이클 필드가 아니라 **릴리스 타깃 라벨**이며(§1.1) 따라서 수행할 "전이" 자체가 없다. 실제 결함은 **저작 시점의 값 계약 부재**이고(§1.3), 소유자 부재는 그 2차 효과다(§1.4). |

## §1 배경

### §1.1 `phase:`는 릴리스 타깃 라벨이지 라이프사이클 필드가 아니다

`.claude/rules/moai/development/spec-frontmatter-schema.md`가 이 필드를 정의한다. 개정 전 원문은 이랬다.

```
| `phase` | string | non-empty, typically release target | e.g. `"v3.0.0"` |
```

`origin/main` 전수 스캔(`git ls-tree -r --name-only origin/main -- .moai/specs`, 판정 시점 관측):

```
전수 spec.md                                   564
phase: 키 보유                                 556
라이프사이클 토큰(plan|run|sync|mx, 값 전체)     0
```

키 부재 8개(564 − 556)는 §4에서 스코프아웃한다. 나머지 **556개 전부가 정당한 릴리스·마일스톤 라벨**이며 라이프사이클 토큰은 0건이다 — 형제 SPEC이 9건을 정정한 결과다(§1.12).

> **정정 이력 (모집단 산술)**: 최초 판은 정당한 라벨을 `563 − 9 = 554`로 계산했다. 거짓이다 — 키 부재 8개를 정당한 라벨로 계상했다. 모집단은 전수가 아니라 **키 보유 수**다. 이번 개정에서 판정 시점 관측값으로 재계측했으며, **어떤 리터럴도 판정값으로 고정하지 않는다** — 이 SPEC 자신의 산출물이 커밋되면 모집단이 증가하기 때문이다.

### §1.2 문제였던 9개 (역사 기록 — 이미 정정됨)

| status | phase (정정 전) | SPEC |
|---|---|---|
| completed | plan | SPEC-CI-LOOP-DEVONLY-001 |
| completed | plan | SPEC-ENVKEY-ANTHROPIC-SSOT-001 |
| completed | plan | SPEC-PIPELINE-FANOUT-ACTIVATION-001 |
| completed | plan | SPEC-UPDATE-GUARD-EFFICACY-001 |
| completed | plan | SPEC-UPDATE-REINSTALL-LOOP-002 |
| completed | plan | SPEC-WORKTREE-BRANCH-GUARD-001 |
| completed | plan | SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 |
| completed | sync | SPEC-REF-SEO-ABSORB-001 |
| draft | plan | SPEC-UPDATE-YAML-PRESERVE-001 |

9개 전부 `phase: "v3.0.2"`로 정정되어 `origin/main`에 착지했다(§1.12). 이 표는 결함의 **형태**를 보존하기 위한 기록이며, 이 SPEC은 더 이상 이 정정을 산출물로 요구하지 않는다.

### §1.3 근본 원인은 저작 시점이며, "누락된 전이"가 아니다

9개 전부 `author: manager-spec`이다. manager-spec이 값을 쓰기 직전 참조하는 체크리스트(`.claude/skills/moai/workflows/plan/spec-assembly.md` § Pre-Write Frontmatter Checklist)의 해당 항목은 `origin/main`에 지금도 원문 그대로 있다.

```
- [ ] `phase: "vX.Y.Z target"` — release phase string
```

형식 예시는 있으나 **금지값이 없다.** "release phase string"이라는 표현은 `plan`·`sync` 같은 단계 이름을 배제할 만큼 강하지 않으며, 실제로 manager-spec은 자기가 서 있던 단계의 이름을 그 자리에 적었다. 스키마 SSOT의 "typically release target"도 같은 약함을 공유한다 — `typically`는 규범이 아니라 경향 서술이다.

**따라서 수행되지 않은 "전이"란 존재하지 않는다.** `phase:`는 라이프사이클을 따라 값이 바뀌는 필드가 아니므로 `plan → run → sync`로 갱신할 주체를 지정하는 것은 문제를 잘못 푸는 것이다.

**이것이 이 SPEC에 남은 중심이다.** 기계 가드는 착지 후 잘못 쓰인 값을 잡지만, 애초에 쓰이지 않게 만들지는 못한다 — 저작자가 읽는 표면은 체크리스트이지 lint 출력이 아니다.

### §1.4 소유자 부재는 2차 효과다 — 한번 쓰이면 아무도 못 고친다

소유권 경계를 직접 확인했다. `.claude/agents/moai/manager-docs.md`:

> "manager-docs limited to frontmatter `status` + `updated` field transitions only" … "**NEVER** other frontmatter fields"

manager-develop도 같은 제약을 받는다(`spec-frontmatter-schema.md` § Forbidden ownership crossings). 그리고 `phase:` 정정은 `status:`를 바꾸지 않으므로 **Status Transition Ownership Matrix가 애초에 다루지 않는 행위다.** 결과적으로 세 관리 에이전트 중 누구도 이미 쓰인 `phase:` 값을 고칠 권한이 없고, 값은 사실상 불변이 된다.

형제 SPEC이 9건을 정정했다는 사실은 이 공백을 메우지 않는다 — 그것은 한 번의 예외적 일괄 정정이었지, 다음 오기(誤記)를 누가 고치는지에 대한 규정이 아니다.

### §1.5 이 필드는 하중을 받는다

`internal/spec/era.go`가 `phase:`를 H-5 era 판정의 타이브레이커로 읽는다.

```go
// H-5: tie-breaker via phase or created date
if matchesModernPhase(signals.FrontmatterPhase) ||
    isAfterModernThreshold(signals.FrontmatterCreated) {
    return EraV3R6, "H-5 (modern phase or created date)"
}
```

`matchesModernPhase`는 문자열이 `v3r6`을 포함하거나 `v3.0` / `"v3.0`으로 시작할 때만 참이다. 따라서 `matchesModernPhase("plan")` → **거짓**이며, H-5의 두 신호 중 하나가 조용히 무력화된다.

**다만 오분류를 일으키지는 않았다.** 9개 모두 `created`가 2026-07-28~08-02이고 `modernEraThreshold`는 `2026-04-01`이므로 두 번째 이접지가 발화해 era는 여전히 V3R6으로 확정됐다. 이 SPEC이 주장하는 바는 "era 분류가 깨졌다"가 아니라 **"두 신호짜리 술어가 한 신호로 퇴화했고, 그 퇴화가 created-date 폴백에 가려져 있다"**이다. 문서를 읽는 사람이 이 결속을 알 수 없다는 것이 남은 결함이다(REQ-PFO-003).

### §1.6 기계 가드는 이제 존재하며, 이 SPEC의 소유가 아니다

`moai spec lint`에 `phase:` 값 가드가 착지했다. 소유자는 **`SPEC-PHASE-FIELD-VALIDATION-001`**이며, 이 SPEC은 그 매처를 재기술하지 않고 소유자를 가리킨다(§1.12).

이 SPEC에 남는 의무는 가드의 **구현**이 아니라, 스키마 SSOT가 그 가드의 동작을 **정확히 기술하는 것**이다(REQ-PFO-016). 하중 필드의 계약을 적는 문서가 집행 규칙의 의미를 잘못 적으면, 그것은 계약이 없는 것보다 나쁘다 — 독자가 틀린 확신을 갖는다.

### §1.8 매칭 축은 대소문자가 아니라 부분 문자열 대 값 전체 일치였다

> **이 절은 v0.4.1까지 틀렸다.** 이전 판은 "대소문자 무시 매칭은 즉시 실물 8건에 오탐을 낸다"고 단언하고 그것을 **결정 축**으로 삼아 대소문자 구분 매칭을 SPEC의 결론으로 확정했다. **그 단언은 부분 문자열 매칭과 결합했을 때만 참이다.**

`origin/main`의 `phase:` 값 556개를 네 방식으로 계수했다(판정 시점 실측).

```
대소문자 무시 + 값 전체 일치 (착지한 매처)   → 0
대소문자 무시 + 부분 문자열                  → 8
대소문자 구분 + 부분 문자열                  → 0
'Runtime' 포함 값                            → 8
```

**두 축은 독립이며, 판별력을 가진 것은 두 번째다.**

- **부분 문자열 매칭**은 대소문자를 무시하는 순간 `Runtime`의 `Run`을 잡아 실물 8건에 오탐을 낸다. 이전 판이 관측한 것이 이것이다.
- **값 전체 일치**로 판정하면 대소문자를 무시해도 오탐이 없다. `"v3.0.0 — Phase 2 — Runtime Hardening"`을 케이스 폴딩해도 `"run"`과 **같지 않기** 때문이다. 실측 계수 `0`이 이를 확인한다.

따라서 이전 판이 배제한 축(대소문자 무시)은 **배제할 이유가 없었다.** 착지한 매처는 대소문자 무시 + 값 전체 일치이며, 이 SPEC이 명세했던 대소문자 구분 + 값 전체 일치보다 **엄격하다** — `PLAN`, `Sync`, `"  run  "` 같은 대소문자·공백 변형 오타를 추가로 잡으면서 오탐은 늘지 않는다.

**정정된 결론: 결정 축은 부분 문자열 대 값 전체 일치다. 대소문자 구분 여부는 값 전체 일치 위에서는 판별력이 없고, 무시하는 쪽이 더 넓게 잡는다.** 이전 판은 두 축이 얽힌 관측(`grep -icE` 대 `grep -cE`, **둘 다 부분 문자열**)에서 잘못된 축을 결정 축으로 지목했다 — 한 축만 바꾼 대조군을 만들지 않은 것이 오진의 기법적 원인이다.

> **왜 조용히 고치지 않고 기록하는가.** 이 SPEC의 HISTORY 관행이 "무엇이 틀렸는지 이름을 붙인다"이며, 이 오진은 사용자 왕복을 대체하겠다며 제시한 실측 근거 자체가 틀렸던 사례다. 근거가 틀린 채로 결론만 맞는 것은 다음 판단을 보증하지 못한다.

### §1.9 대상 문서 2개 모두 템플릿 미러를 가지며, 현재 바이트 동일하다

이 SPEC이 고치는 문서 두 개는 템플릿 미러가 실재하고 현재 패리티 상태다(양쪽 `diff` 0행, 판정 시점 실측).

```
internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
```

로컬만 고치면 지금 성립하는 패리티가 깨지고, 결함 있는 문구가 배포 사용자에게 그대로 나간다. 더 나쁜 것은 **탐지되지 않는다**는 점이다 — 로컬 경로만 grep하는 판정은 미러가 드리프트해도 전부 PASS한다. 따라서 CLAUDE.local.md §2 Template-First Rule에 따라 로컬과 미러를 같은 커밋에서 함께 수정하고 `make build`를 실행한다(NFR-PFO-001).

**중립성이 패리티의 형태를 규정한다.** 추가 문구는 §25 템플릿 중립성을 지켜야 하므로 SPEC ID·REQ 토큰·내부 날짜·commit SHA를 담지 않는다. 이 제약이 REQ-PFO-016의 형태를 결정한다 — 가드 소유자를 **SPEC ID로** 가리키면 미러에는 넣을 수 없어 패리티가 깨진다. 따라서 배포 문서에서는 소유자를 **코드 위치**(`internal/spec/lint.go` `FrontmatterSchemaRule`, finding 코드 `FrontmatterPhaseInvalid`)로 가리키고, SPEC ID 수준의 cross-reference는 이 SPEC 자신의 본문(§1.12)에만 둔다.

> **바이트 동일은 이 두 파일의 조건부 사실이지, 리포 일반 규칙이 아니다.** 리포에는 §25 중립화 때문에 **의도적으로 divergent한 미러 쌍이 실재한다** — `.claude/skills/moai/workflows/plan.md`와 그 미러는 `diff` **6행**이며, 그 차이는 로컬 사본의 내부 수치·내부 날짜·말미 `Updated:` 줄이 템플릿에서 제거된 결과다. 올바른 일반 의무는 "두 파일을 동일하게 만든다"가 아니라 **"같은 의미의 편집을 양쪽에 적용하되 각 사본의 중립화를 보존한다"**이다. 이 SPEC은 `workflows/plan.md`를 편집하지 않으므로 그 쌍은 대상이 아니다.

### §1.10 정정 소유자 결정 — manager-spec, 매트릭스 바깥 별도 절

두 결정 모두 보유 증거로 확정한다(사용자 왕복 불요).

**결정 1 — 소유자는 manager-spec(오케스트레이터 재위임 경유).** 근거는 경계 훼손이 가장 작다는 것이다. manager-spec은 이미 `spec.md`를 정본(canonical SSOT) 본문으로 소유하고 있고, 오케스트레이터 재위임을 통한 mid-run 편집 권한(D-NEW-1 inline-fix 패턴)도 이미 규정되어 있다. 새 권한을 만드는 것이 아니라 **기존 소유 범위가 이미 덮는 영역**이다.

| 기각한 대안 | 사유 |
|---|---|
| manager-docs | 허용 필드를 넓히려면 "**NEVER** other frontmatter fields"라는 밝은 선(bright line)을 흐려야 한다. 그 경계의 가치는 정확히 예외가 없다는 데 있다. 한 필드를 위해 예외를 뚫으면 다음 필드의 근거가 된다. |
| 오케스트레이터 직접 | 매트릭스의 유일한 오케스트레이터 행(`* → rejected`)조차 실제 기록은 manager-docs가 한다. 오케스트레이터가 `spec.md`를 직접 쓰는 선례가 없고, 이 SPEC이 그 선례를 만들 이유가 없다. |

**결정 2 — 매트릭스 바깥의 별도 절.** 매트릭스의 이름과 계약은 *Status Transition* Ownership Matrix다. `phase:` 정정은 status를 바꾸지 않으므로 그 안에 행을 추가하면 매트릭스가 무엇을 관장하는지가 흐려진다. 다만 매트릭스만 읽는 독자가 "소유자 없음"으로 오독할 위험은 실재하므로, 별도 절 배치와 **매트릭스 내 한 줄 포인터 주석**을 함께 요구한다(REQ-PFO-007).

### §1.12 형제 SPEC과의 분업 — 무엇이 넘어갔고 무엇이 남았나

run-phase 도중 `SPEC-PHASE-FIELD-VALIDATION-001`(커밋 `998744216`, PR #1285)이 `origin/main`에 착지해 이 SPEC의 기계 가드 축과 실물 정정 축을 먼저 구현했다.

| 축 | 소유 | 상태 (`origin/main` 실측) |
|---|---|---|
| lint 가드 (매처 · finding 코드 · 심각도 · grandfather 상호작용) | **SPEC-PHASE-FIELD-VALIDATION-001** | 착지 완료 — 이 SPEC은 재기술하지 않고 가리킨다 |
| 9개 실물 정정 | **SPEC-PHASE-FIELD-VALIDATION-001** | 착지 완료 — 전수 스캔 라이프사이클 토큰 0건(§1.1) |
| 스키마 SSOT 값 계약 규범화 + 금지값 열거 + era H-5 결속 | **이 SPEC** | 미착지 — `typically release target` 1건 잔존, 금지값 절 0건, `H-5` 0건 |
| 저작 체크리스트 + pre-write gate HALT | **이 SPEC** | 미착지 — 체크리스트 원문 그대로 |
| 정정 소유자 규정 (비전이 프론트매터 정정 절) | **이 SPEC** | 미착지 — 0건 |
| 배포 문서가 착지한 가드를 정확히 기술 | **이 SPEC** | 미착지 — 현재 워크트리 문구가 오기 상태(§1.13) |

**분업의 원칙: 가드는 형제 SPEC이 소유하고, 이 SPEC은 그 위층(저작 시점 계약·소유권·문서 정확성)만 다룬다.** 두 SPEC이 같은 매처를 기술하면 중복 원천이 되어, 한쪽이 바뀔 때 다른 쪽이 조용히 낡는다.

### §1.13 배포 문서가 착지한 가드를 오기(誤記)하고 있다

이 SPEC의 브랜치가 스키마 SSOT에 이미 추가한 문장이 있다.

```
The prohibition binds the whole value, case-sensitively.
```

**착지한 가드는 대소문자를 구분하지 않는다**(§1.8 실측). 즉 이 문장은 배포되는 규칙 문서가 집행 규칙의 의미를 잘못 말하는 상태다.

이것은 사소한 오타가 아니라 **이 SPEC의 주제에 대한 자기모순**이다. 이 SPEC의 논지는 "하중을 받는 필드의 계약을 문서가 약하게·부정확하게 적어서 결함이 났다"는 것이다. 그 문서를 고치는 SPEC이 같은 문서에 새로운 부정확성을 심으면, 세우려는 계약 자체가 신뢰를 잃는다. 게다가 이 오기는 **틀린 방향으로 안전하다** — 독자는 `PLAN`·`Sync` 같은 대소문자 변형이 허용된다고 읽지만 실제로는 거부된다. 문서를 믿고 쓴 값이 가드에 걸린다.

REQ-PFO-016이 이 정정을 요구한다.

## §2 요구사항 (GEARS)

### 값 계약 — `phase:`가 무엇인지 규범적으로 말한다

- **REQ-PFO-001** — 스키마 SSOT(`spec-frontmatter-schema.md`)는 `phase:`의 유효값을 **릴리스 또는 마일스톤 타깃 라벨**로 규범적으로 정의해야 한다(shall). "typically release target"이라는 경향 서술로는 부족하며, 규범 표현으로 대체해야 한다.
- **REQ-PFO-002** — 스키마 SSOT는 라이프사이클 단계 이름(`plan` / `run` / `sync` / `mx`)을 `phase:` 값으로 쓰는 것을 **명시적으로 금지해야 한다**(shall not). 금지 대상을 열거해야 하며, "릴리스 타깃을 쓰라"는 긍정 지시만으로 갈음해서는 안 된다 — 긍정 지시만 있던 것이 이 결함의 직접 원인이다(§1.3).
- **REQ-PFO-003** — 스키마 SSOT는 `phase:`가 `internal/spec/era.go`의 H-5 era 판정에 입력된다는 사실과, 라이프사이클 토큰이 그 술어의 한 신호를 무력화한다는 결과를 기술해야 한다(shall). 현재 문서를 읽는 사람은 이 필드가 하중을 받는다는 것을 알 수 없다.
- **REQ-PFO-016** — 스키마 SSOT의 금지 서술은 **착지한 집행 규칙의 동작을 정확히 기술해야 한다**(shall): 판정이 트림한 **값 전체**에 대해 **대소문자를 무시하고** 이루어진다는 점, finding 코드가 `FrontmatterPhaseInvalid`라는 점, 심각도가 error라는 점, 그리고 그 코드가 grandfather 강등 목록에서 **의도적으로 제외**되어 있다는 점. 이 서술은 매처를 재구현 수준으로 복제해서는 안 되며(shall not), 집행 지점을 **코드 위치**로 가리켜야 한다(shall) — SPEC ID 참조는 §25 템플릿 중립성 때문에 배포 사본에 넣을 수 없다(§1.9).

  > **정정 이력**: 이 SPEC의 브랜치는 같은 자리에 "The prohibition binds the whole value, case-sensitively."를 이미 써 넣었다. **거짓이며**(§1.8 실측), 배포 문서가 집행 규칙을 오기하는 상태다(§1.13). 이 REQ는 그 문장의 정정을 요구한다.

### 저작 시점 방지와 사후 정정 소유자

- **REQ-PFO-004** — `.claude/skills/moai/workflows/plan/spec-assembly.md`는 저작 시점에 두 표면에서 금지값을 차단해야 한다(shall): (i) Pre-Write Frontmatter Checklist의 `phase:` 항목이 REQ-PFO-002의 금지 열거를 담아야 하며, (ii) **When** manager-spec이 `phase:` 값으로 라이프사이클 토큰을 담은 프론트매터를 생성하면, pre-write gate가 Write를 중단하고 스키마 위반을 보고해야 한다(shall halt) — 기존 HALT 트리거(누락 필드·거부 별칭)에 금지값 조건을 추가한다.

  > 이 REQ는 v0.4.1의 REQ-PFO-004(체크리스트)와 REQ-PFO-005(gate HALT)를 병합한 것이다. 두 절은 같은 파일의 같은 관심사(저작 시점 차단)이고 한 번의 편집으로 함께 착지하며, Tier S 요구사항 상한 8에 맞추기 위한 통합이다. 판정은 AC-PFO-004가 두 명령으로 각각 수행하므로 검증력은 보존된다.

- **REQ-PFO-006** — 이미 쓰인 `phase:` 값을 정정할 소유자는 **manager-spec**(오케스트레이터 재위임 경유)으로 소유권 문서에 지정되어야 한다(shall). 현재는 세 관리 에이전트 모두 금지되어 있어 소유자가 0명이다(§1.4). 선택 근거와 기각한 대안은 §1.10에 기록한다.
- **REQ-PFO-007** — `phase:` 값 정정은 `completed → in-progress (amendment)` 전이를 **요구해서는 안 된다**(shall not require). 한 필드의 오타성 정정에 amendment 절차(status 왕복 + `amendment_of` 필드 + HISTORY `## Amendments` 절 + plan-audit 캐시 무효화)를 강제하는 것은 비용이 결함에 비해 과도하다. 정정 경로는 `status:`를 건드리지 않으므로 Status Transition Ownership Matrix의 어느 행에도 해당하지 않는다. 따라서 그 규정은 매트릭스 **바깥의 별도 절**로 배치하고(shall), 매트릭스에는 그 절을 가리키는 한 줄 주석을 남겨야 한다(shall) — 근거는 §1.10.

### 비기능 요구사항

- **NFR-PFO-001** — 문서를 수정하는 모든 마일스톤은 로컬 사본과 템플릿 미러를 **같은 커밋에서 함께** 수정해야 하며(shall), 이어서 `make build`를 실행해야 한다(CLAUDE.local.md §2 Template-First Rule). 대상 미러 2개는 §1.9가 열거한다. 미러에 추가하는 문구는 §25 템플릿 중립성을 지켜야 하며 SPEC ID·REQ 토큰·내부 날짜·commit SHA를 담아서는 안 된다(shall not).

## §3 수용 기준 (Tier S — 인라인)

### 판정 규약

- **판정 기준(base)은 실행 시점에 재계산한다.** 이 문서는 SHA를 판정 앵커로 쓰지 않는다 — 저작 시점의 SHA는 판정 시점에 낡는다. 위치 지정은 라인 번호가 아니라 내용 토큰으로 한다.
- **이 셸은 `ls`를 `ls -la`로 alias한다.** 파일 존재 판정에는 `git ls-files` 또는 glob을 쓴다.
- **baseline 두 값을 구분해 기록한다.** `origin/main`(편집 이전 상태)과 **현 워크트리**(M1~M3 편집이 이미 착지한 상태). 두 값이 다른 항목은 그 편집이 이미 존재한다는 뜻이지 남은 작업이 없다는 뜻이 아니다 — REQ-PFO-016이 요구하는 정정은 아직 미착지다.

공통 변수:

```bash
S=.claude/rules/moai/development/spec-frontmatter-schema.md
A=.claude/skills/moai/workflows/plan/spec-assembly.md
TS=internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
TA=internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md
```

### AC-PFO-001 — 스키마 SSOT가 유효값을 규범적으로 정의한다 (REQ-PFO-001)

**Given** `phase` 행이 "typically release target"이라는 경향 서술만 담고 있고,
**When** 그 행이 릴리스·마일스톤 타깃 라벨을 규범 표현(MUST)으로 정의하도록 개정되면,
**Then** `typically release target` 문자열이 사라진다.

```bash
grep -c 'typically release target' "$S"
```

- `origin/main` 관측 `1` · 현 워크트리 관측 `0`
- PASS 조건: `0`

### AC-PFO-002 — 스키마 SSOT가 라이프사이클 토큰을 명시적으로 금지한다 (REQ-PFO-002)

**Given** 금지값 열거가 없고,
**When** `plan` / `run` / `sync` / `mx` 네 토큰을 금지하는 절이 추가되면,
**Then** 금지 문맥과 네 토큰이 모두 검출된다.

```bash
# (a) `phase:` 값 문맥의 금지 표현
grep -c -i '`phase:`.*\(prohibit\|MUST NOT\|금지\)\|\(prohibit\|MUST NOT\|금지\).*`phase:`' "$S"
# (b) 금지 절 본문의 4토큰
sed -n '/Prohibited phase values/,/^## /p' "$S" | grep -oE '\b(plan|run|sync|mx)\b' | sort -u | wc -l | tr -d ' '
```

- `origin/main` 관측 (a) `0` / (b) `0`(절 부재) · 현 워크트리 관측 (a) `1` / (b) `4`
- PASS 조건: (a) `>= 1`, (b) `4`
- **[산출물 요건] 신설 절의 제목은 리터럴 `Prohibited phase values`를 포함해야 한다.** (b)의 `sed` 범위가 이 제목에 앵커되어 있다 — 다른 제목을 쓰면 출력이 비어 계수가 `0`이 되고 fail-closed로 떨어진다.

### AC-PFO-003 — 스키마 SSOT가 era.go 결속을 기술한다 (REQ-PFO-003)

**Given** 독자가 `phase:`의 era 결속을 알 수 없고,
**When** H-5 결속과 그 결과가 기술되면,
**Then** `matchesModernPhase` 또는 `H-5` 토큰이 검출된다.

```bash
grep -c 'matchesModernPhase\|H-5' "$S"
```

- `origin/main` 관측 `0` · 현 워크트리 관측 `1`
- PASS 조건: `>= 1`

### AC-PFO-004 — 저작 시점 차단이 두 표면에 존재한다 (REQ-PFO-004)

**Given** 체크리스트 항목이 형식 예시만 담고, pre-write gate가 "누락 필드 OR 거부 별칭"에만 HALT하고,
**When** 두 표면에 금지값 조건이 추가되면,
**Then** 두 판정이 모두 만족된다.

```bash
# (a) 체크리스트 항목 줄의 금지 표현
grep -n '`phase:' "$A" | grep -ci 'NEVER\|금지\|NOT a lifecycle'
# (b) pre-write gate 절의 금지값 HALT 트리거
sed -n '/Pre-write gate behavior/,/^- \.moai/p' "$A" | grep -ci 'prohibited value\|금지값\|lifecycle token'
```

- `origin/main` 관측 (a) `0` / (b) `0` · 현 워크트리 관측 (a) `1` / (b) `1`
- PASS 조건: (a) `>= 1` **AND** (b) `>= 1` — 한쪽만으로는 PASS가 아니다. 체크리스트는 사람이 읽는 표면이고 gate는 기계가 멈추는 표면이며, 서로를 대체하지 않는다.

### AC-PFO-006 — 정정 소유자가 지정된다 (REQ-PFO-006)

**Given** 세 관리 에이전트 모두 금지되어 소유자가 0명이고,
**When** 단일 소유 에이전트가 소유권 문서에 지정되면,
**Then** 지정이 검출되고, 명명된 에이전트가 유지 카탈로그 소속이다.

```bash
grep -n -i 'phase.*correction\|phase 값 정정' "$S" | head
```

- `origin/main` 관측 0행 · 현 워크트리 관측 `1`행
- PASS 조건: `>= 1`행, 그리고 그 행이 명명한 에이전트가 `manager-spec` / `manager-docs` / `manager-develop` 중 하나 (archived 에이전트 이름이면 FAIL)

### AC-PFO-007 — amendment 비요구 + 배치 2건 (REQ-PFO-007)

REQ-PFO-007은 **세 가지**를 요구한다: (a) amendment 비요구 명시, (b) 규정을 매트릭스 **바깥** 별도 절에 배치, (c) 매트릭스 **안에** 그 절을 가리키는 포인터 주석.

```bash
grep -c -i 'amendment.*not required\|amendment를 요구하지' "$S"                        # (a)
grep -c '^#\+ .*Non-transition frontmatter corrections' "$S"                            # (b0)
awk '/^## Status Transition Ownership Matrix/{f=1;next} /^## /{f=0} f' "$S" \
  | grep -c '^#\+ .*Non-transition frontmatter corrections'                             # (b)
awk '/^## Status Transition Ownership Matrix/{f=1;next} /^## /{f=0} f' "$S" \
  | grep -v '^#' | grep -c -i 'Non-transition frontmatter corrections\|비전이 프론트매터 정정'  # (c)
```

- `origin/main` 관측 전부 `0` · 현 워크트리 관측 (a) `1` / (b0) `1` / (b) `0` / (c) `1`
- PASS 조건: (a) `>= 1`, **(b0) `>= 1`**, (b) `0`, (c) `>= 1`
- **(b)는 (b0) 없이는 공허하다.** 절이 아예 없어도 (b)는 `0`을 반환한다 — `origin/main` baseline이 정확히 그 상태다. (b0)가 `0`이면 (b)의 `0`은 PASS가 아니라 미구현이다.
- **(c)의 매칭은 헤딩이 아닌 산문 줄이어야 한다.** 포인터를 인용 줄(`>`)로 쓰면 (b)의 헤딩 계수를 올리지 않는다.
- **[산출물 요건] 신설 절의 제목은 리터럴 `Non-transition frontmatter corrections`를 포함해야 한다.**

### AC-PFO-016 — 배포 문서가 착지한 가드를 정확히 기술한다 (REQ-PFO-016)

**Given** 브랜치가 스키마 SSOT에 "The prohibition binds the whole value, case-sensitively."를 써 넣었고 그것이 착지한 가드와 반대이고(§1.13),
**When** 그 서술이 대소문자 무시 + 값 전체 일치로 정정되고 finding 코드·심각도·grandfather 제외가 함께 기술되면,
**Then** 오기 문자열이 사라지고 네 사실이 모두 검출된다.

```bash
grep -c -i 'case-sensitiv' "$S"                       # (a) 오기 제거
grep -c -i 'case-insensitiv\|case-fold' "$S"          # (b) 정정된 매칭 서술
grep -c 'FrontmatterPhaseInvalid' "$S"                # (c) finding 코드
grep -c 'eraDemotableCodes' "$S"                      # (d) grandfather 강등 제외
```

- `origin/main` 관측 (a) `0` / (b) `0` / (c) `0` / (d) `0` — 서술 자체가 없다
- **현 워크트리 관측 (a) `1` / (b) `0` / (c) `0` / (d) `0` — 오기가 존재하고 정정이 미착지다. 이 SPEC의 남은 작업 중 유일하게 현재 FAIL인 항목이다.**
- PASS 조건: (a) `0` **AND** (b) `>= 1` **AND** (c) `>= 1` **AND** (d) `>= 1`
- **(a)와 (b)를 함께 판정하는 이유**: (a) 단독은 문장을 통째로 삭제해도 PASS한다. 정정은 삭제가 아니라 **올바른 서술로의 교체**이므로 (b)가 대체 서술의 실재를 요구한다.

  > **정정 이력 ((d) 선택자)**: 이 AC의 첫 판은 (d)를 `grep -ci 'eraDemotableCodes\|grandfather'`로 쓰고 baseline을 `0`으로 기록했다. **거짓이다** — `grandfather`는 Status Enum 절의 "grandfathered SPECs that carry `status: planned`" 문장 때문에 `origin/main`과 현 워크트리 양쪽에서 이미 `1`이며, 그 이접지는 무편집 상태에서 이미 충족되는 **비판별 검사**였다. 실측으로 확인해 정확한 토큰 `eraDemotableCodes`(양쪽 `0`)만 남겼다. 이 SPEC이 v0.4.0에서 AC-PFO-002에 대해 고쳤던 것과 **같은 결함 유형**(죽은 검사)이 다른 AC에서 재발한 것이다.
- 판정은 **로컬 사본**에서 하고, 미러 준수는 AC-PFO-014(패리티)와의 결합으로 함의된다.

### AC-PFO-014 — 로컬↔미러 패리티가 유지된다 (NFR-PFO-001)

**Given** 대상 문서 2개가 로컬과 템플릿 미러 사이에 바이트 동일 패리티를 이루고 있고(§1.9),
**When** 편집이 로컬에 적용되면,
**Then** 같은 편집이 미러에도 적용되어 패리티가 **여전히** 0행이다.

```bash
diff "$S" "$TS" | wc -l | tr -d ' '
diff "$A" "$TA" | wc -l | tr -d ' '
go test -count=1 -run 'TestInternalContentLeak|TestTemplateNeutrality' ./internal/template/
```

- `origin/main` 관측 양쪽 `0` · 현 워크트리 관측 양쪽 `0`
- PASS 조건: 두 `diff` 모두 `0` **AND** 중립성 가드 `ok` **AND** AC-PFO-001~004·006·007·016 PASS
- **단독 판정 금지.** 두 파일을 **전혀 편집하지 않아도** `diff`는 `0`이다. 이 AC는 로컬 grep AC들과의 결합으로만 미러 준수를 함의한다 — 로컬 grep이 PASS이고 미러가 로컬과 바이트 동일하면 미러도 같은 grep을 만족한다.

### Definition of Done

- [ ] AC-PFO-001 / 002 / 003 / 004 / 006 / 007 / 014 / 016 = **8개** 전부 PASS (Tier S 상한 8)
- [ ] **AC-PFO-016이 실제로 관측 PASS** — 이 SPEC의 남은 작업 중 현재 FAIL인 유일 항목이며, 이것 없이 나머지 7개가 PASS인 것은 착수 이전 상태와 구별되지 않는다
- [ ] AC-PFO-014가 AC-PFO-001~004·006·007·016 PASS와 **결합해서** 판정되었음 — 무편집 상태에서도 `diff` `0`이므로 단독 PASS로 읽지 않았음
- [ ] AC-PFO-007의 (b) 배치 판정이 (b0) 절-실재 판정과 **함께** 읽혔음 — (b0)가 `0`인 채로 (b)의 `0`을 PASS로 읽지 않았음
- [ ] AC-PFO-004의 (a)·(b) 두 표면이 **모두** 만족되었음 — 한쪽만으로 갈음하지 않았음
- [ ] `make build` 실행 후 템플릿 중립성 가드 통과
- [ ] `progress.md` §E.2에 위 관측이 명령 + 축약 없는 출력으로 기록됨

## §4 범위 밖 (Out of Scope)

### Out of Scope — lint 가드의 구현·매처·심각도

`phase:` 값 가드는 `SPEC-PHASE-FIELD-VALIDATION-001`이 소유하며 이미 착지했다(§1.12).

- 이 SPEC은 그 매처를 **재구현하지 않고, 재기술하지도 않는다.** 두 SPEC이 같은 매처를 기술하면 중복 원천이 되어 한쪽이 바뀔 때 다른 쪽이 조용히 낡는다.
- 이 SPEC이 요구하는 것은 배포 문서가 그 가드의 동작을 **정확히 가리키는 것**뿐이다(REQ-PFO-016).
- 가드의 심각도(error)·grandfather 제외·finding 코드 변경 요구는 이 SPEC의 소관이 아니다. 이견이 있으면 소유 SPEC에 제기한다.

### Out of Scope — 9개 실물 SPEC의 `phase:` 값 정정

형제 SPEC이 이미 전부 `"v3.0.2"`로 정정했고, 전수 스캔이 라이프사이클 토큰 0건을 확인한다(§1.1). 재적용은 무변경 편집이 되거나 불필요한 충돌을 만든다.

### Out of Scope — `phase:` 필드가 아예 없는 8개 SPEC

값이 틀린 것과 필드가 없는 것은 원인도 처방도 다르다. 후자는 필드 도입 이전의 유물이며, `FrontmatterSchemaRule`의 필수 필드 루프가 이미 `FrontmatterInvalid`로 검출하되 grandfather 정책에 의해 의도적으로 warning으로 강등한 기존 부채다. 이 SPEC이 손대면 grandfather 보호를 우회한다. 8개 중 6개가 `_archive/` 하위라는 점도 별도 판단을 요구한다 — 아카이브 문서의 소급 정규화는 별개 정책 문제다.

### Out of Scope — era.go의 H-5 술어 자체 수정

`matchesModernPhase`가 `v3.0`·`v3r6` 접두만 인정하는 것은 이 SPEC의 결함이 아니다. 이 SPEC은 그 함수에 올바른 입력이 들어가도록 값 계약을 세울 뿐, 함수의 매칭 범위를 넓히거나 좁히지 않는다. §1.5가 기록한 대로 현재 오분류는 0이므로 술어 변경은 필요를 증명하지 못한 변경이 된다.

### Out of Scope — 다른 프론트매터 필드의 값 계약

`module:`·`lifecycle:`·`tags:`도 비어있음만 검사받는다. 같은 계열의 공백일 수 있으나 그 필드들에서 실제 일탈이 관측되지 않았다. 관측 없이 가드를 세우는 것은 이 SPEC이 비판하는 것과 같은 종류의 미검증 처방이다. 일탈이 관측되면 별도 SPEC으로 다룬다.

## §5 성공 기준

§3의 AC 8개가 전부 PASS이며, 그중 **AC-PFO-016이 관측으로 PASS**해야 한다. 나머지 7개는 이미 브랜치에 착지해 있으므로, 016 없이 얻은 "8개 중 7개 PASS"는 이번 개정이 아무것도 바꾸지 않았다는 것과 구별되지 않는다.
