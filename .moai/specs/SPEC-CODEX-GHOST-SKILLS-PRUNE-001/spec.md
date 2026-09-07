---
id: SPEC-CODEX-GHOST-SKILLS-PRUNE-001
title: "유령이 된 [[skills.config]] 등록을 지우는 동사 — 지울 수 있는 것만 지우고, 판정이 안 서면 손대지 않는다"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, skills, prune, destructive-verb, dry-run, safety-boundary"
related_specs: [SPEC-CODEX-SKILL-PATH-001, SPEC-CODEX-WIRING-001, SPEC-V3R6-MOAI-CLEAN-HOME-001]
---

# SPEC-CODEX-GHOST-SKILLS-PRUNE-001 — 유령 스킬 등록 제거 동사

## HISTORY

- 2026-09-07 (plan-phase, v0.1.0) — 카드 t506 축 1. 측정된 baseline(`.moai/reports/t506/baseline-measurement.md`)을 근거로 작성. 축 2(`codex_measured_version` 재측정)는 선행 두 축(t496 훅 축, t507 플러그인 매니페스트 축)이 착지하기 전이라 분리했다.

## §A 배경

Codex 는 등록된 스킬을 사용자 계층 `~/.codex/config.toml` 의 `[[skills.config]]` 배열-테이블에 기록한다. 그 `path` 가 가리키던 파일이 사라져도 Codex 는 항목을 지우지도, 불평하지도 않는다. 그래서 등록은 조용히 **유령**으로 남는다.

지금 이 저장소가 가진 것은 **보고 표면뿐이다.** `moai doctor` 의 Codex Wiring 검사가 유령 개수를 세어 알려주지만(`internal/cli/doctor_codex.go` 의 `codexStaleSkillFinding`), 그 파일의 머리주석이 스스로 못박은 대로 "The check READS only" — 지우는 손은 어디에도 없다. 사용자는 개수를 통보받고 손으로 편집해야 한다.

이 SPEC 은 그 손을 만든다. 다만 **지울 수 있다고 판정된 것만** 지운다.

### §A.1 측정된 baseline

정본: `.moai/reports/t506/baseline-measurement.md` (트리 `ace1c5440`, 2026-09-07). 요약:

| 분류 | 이 기계의 개수 |
|---|---|
| absolute → `fs.ErrNotExist` (`enabled = false`) | 49 |
| absolute → 존재함 | 0 |
| home-relative | 0 |
| relative | 0 |
| oddly-formed | 0 |
| indeterminate (`ErrNotExist` 아닌 stat 오류) | 0 |
| `path` 키 없는 항목 | 0 |

**안전 경계에 해당하는 네 분류가 이 기계에서 전부 0이다.** 살아 있는 설정은 경계를 한 번도 밟지 않는다 — 경계는 관측으로 확인되는 성질이 아니라 **테스트가 지는 의무**다(§C.4).

## §B 목표

1. 유령 항목을 제거하는 동사를 하나 추가한다.
2. 제거 대상을 **기계적으로 판정 가능한 부재**로만 한정한다.
3. 판정이 서지 않는 모든 경우에 아무것도 하지 않는다.
4. 되돌릴 수 있게 한다 — 쓰기 전 백업, 그리고 그 백업의 sha256 보고.

## §C 요구사항 (GEARS)

> REQ 번호는 **작성 순서**로 할당되고 절은 **관심사**로 묶인다. 그래서 001-021 이 문서 순서대로 나오지 않는다(018 은 §C.3, 019 는 §C.5, 017 은 §C.7 에 있다). 번호 집합에 빠짐도 중복도 없으며, 어긋난 순서는 결번의 징후가 아니다. plan-audit iter-1 이 인용한 번호를 그대로 유지하려고 재번호를 매기지 않았다.
>
> **모달리티에 대한 린터 주의.** `moai spec lint` 는 이 SPEC 에 `No findings` 를 낸다. 그 초록을 "GEARS 모달리티가 검사됐다"는 뜻으로 읽어서는 안 된다 — `internal/spec/lint.go:790` 의 `isModalityMalformed` 는 `WHEN `/`WHILE `/`WHERE `/`IF `/`THE ` 라는 **영문 접두사에만** 반응하므로, 한국어 REQ 본문에 대해서는 무조건 false 를 돌려준다. 즉 이 SPEC 에서 모달리티 축의 초록은 **공허하다.** 이 축은 사람이 읽어서 판정해야 하며, 그 사실이 plan-audit iter-1 의 MP-2 실패가 린트 초록과 공존한 이유다.

### §C.1 분류

- **REQ-CGP-001** (Ubiquitous) 시스템은 선언된 `path` 의 모양을 판정할 때 `classifyCodexSkillPath` 를 사용해야 한다. 두 번째 분류기를 작성해서는 안 된다. (근거: t468 이 이 4-모양 분류를 착지시킨 목적이 오탐 부류를 없애는 것이었고, 병렬 구현은 그것을 되살린다.)
- **REQ-CGP-002** (Ubiquitous) 시스템은 `~` 로 시작하는 선언을 확장할 때 `expandCodexHomeRelativePath` 를 사용해야 한다.
- **REQ-CGP-003** (Ubiquitous) 시스템은 설정 파일의 위치를 `resolveCodexHomeDir` 로 해석해야 한다. `~` 확장 씨앗(`codexUserHomeDir`)과 설정 위치 씨앗(`resolveCodexHomeDir`)은 **서로 다른 두 씨앗**이며, 각각 제 역할로만 쓰인다 — `CODEX_HOME` 은 설정을 찾는 데 쓰이고, `~` 를 펴는 데 쓰이지 않는다.

### §C.2 제거 적격

- **REQ-CGP-004** (Ubiquitous) 시스템은 다음 세 조건을 **모두** 만족하는 항목만 제거해야 한다:
  1. 비어 있지 않은 `path` 를 선언한다,
  2. 그 모양이 `codexPathAbsolute` 이거나, `codexPathHomeRelative` 이면서 홈 확장이 성공했다,
  3. 해석된 경로에 대한 `os.Stat` 이 `errors.Is(err, fs.ErrNotExist)` 를 만족하는 오류를 돌려준다.
- **REQ-CGP-005** (Unwanted) 시스템은 `enabled` 값을 적격 판정의 게이트로 사용해서는 안 된다.

  *근거(요구사항이 아님):* `enabled = true` 이면서 경로가 부재한 항목도 유령이다. 그 항목이 제거된다면 그것은 REQ-CGP-004 의 세 조건을 만족했기 때문이지 이 조항이 제거를 지시했기 때문이 아니다 — **이 조항은 제거를 지시하지 않는다.** 이 기계에는 그런 사례가 0건이므로 이 축은 테스트 픽스처에서만 발화한다(AC-CGP-005).

  v0.2.0 의 MP-2 수리는 이 문장을 "시스템은 그 항목을 제거해야 한다"라는 **무조건 제거 의무**로 바꿔 놓았고, 그 결과 `enabled = true` + 경로 부재 + 범위 안 미인식 줄을 동시에 가진 항목에 대해 REQ-CGP-005 는 "제거해야"를, REQ-CGP-018 은 "제거해서는 안 된다"를 지시하는 모순이 생겼다(plan-audit iter-2 E1). 원문을 지우지 않고 이 문단으로 강등한다 — 적격 의무는 원래부터 REQ-CGP-004 에 있었으므로 잃는 것이 없다.

### §C.2.1 조항 간 우선순위

- **REQ-CGP-022** (Ubiquitous) §C.3 의 실격 조항(REQ-CGP-006..011, REQ-CGP-018) 중 하나라도 걸리는 항목에 대해, 시스템은 §C.2 의 적격 조항이 무엇을 말하든 그 항목을 **보존해야 한다**. 실격이 적격을 이긴다.

  이 우선순위는 추론에 맡기지 않고 명시한다. 안전 경계가 "일곱 조항이 경계다"라는 산문으로만 서 있으면, 두 조항이 반대를 지시할 때 구현자가 어느 쪽을 따를지 예측할 수 없다.

### §C.3 절대 제거하지 않는 것

- **REQ-CGP-006** (Unwanted) 시스템은 `codexPathRelative` 항목을 제거해서는 안 된다.
- **REQ-CGP-007** (Unwanted) 시스템은 `codexPathOddlyFormed` 항목을 제거해서는 안 된다.
- **REQ-CGP-008** (Unwanted) 시스템은 홈이 해석되지 않은 home-relative 항목(`expandCodexHomeRelativePath` 가 `ok=false`)을 제거해서는 안 된다.
- **REQ-CGP-009** (Unwanted) 시스템은 `ErrNotExist` 가 아닌 stat 오류(권한 거부, 심링크 루프, I/O 오류 — **indeterminate 부류**)를 만난 항목을 제거해서는 안 된다.
- **REQ-CGP-010** (Unwanted) 시스템은 `path` 키를 선언하지 않은 항목을 제거해서는 안 된다.
- **REQ-CGP-011** (Unwanted) 시스템은 해석에 성공한 경로를 가진 항목을 제거해서는 안 된다. **디렉터리도 해석에 성공한 것**이며 부재가 아니다.
- **REQ-CGP-018** (Unwanted) 시스템은 **줄 범위 안에 인식되지 않는 줄을 담은 항목**을 제거해서는 안 된다.

  **인식 여부는 파서의 상태로 판정하며, 줄의 겉모양으로 판정하지 않는다.** 어떤 줄이 인식된 줄이려면 파서가 그 줄을 다음 다섯 갈래 중 하나로 **실제로 소비**했어야 한다 — `[[skills.config]]` 헤더 매치, `path` 대입 매치, `enabled` 대입 매치, 빈 줄, `#` 로 시작하는 온전한 주석 줄. 그 밖의 모든 줄은 미인식이며, 하나라도 범위 안에 있으면 그 항목은 적격에서 **탈락**한다.

  [HARD] 특히, **여러 줄 리터럴 분기(`internal/codexwiring/skills.go:113-121`)가 삼킨 줄은 그 겉모양이 위 다섯 갈래 중 무엇처럼 보이든 미인식으로 본다.** 파서는 그 줄들에 대해 switch 를 아예 실행하지 않으므로, 그 줄이 헤더처럼 보인다는 사실은 파서가 그것을 헤더로 **읽었다**는 뜻이 아니다.

  [HARD] **판정의 계산과 보고는 파서의 몫이다.** 파서는 항목마다 다섯 갈래 전부와 삼킴 분기를 포괄하는 판정을 계산해 보고해야 하며, 프루너는 그 값을 읽을 뿐 범위의 텍스트를 다시 해석해서는 안 된다. 보고되는 값은 **범위 안에서 처음으로 인식되지 않은 줄의 인덱스**이고, 전부 인식되었으면 그 사실을 나타내는 값이다 — 아래 보고 의무가 "어느 줄 때문인지"를 말할 수 있어야 하기 때문에 boolean 으로는 부족하다.

  이 계산이 파서 안에 있어야 하는 이유는 파서가 이미 답을 쥐고 있어서가 아니다. 지금의 파서는 `skills.go:133-138` 의 `case inEntry:` 무동작 경로에서 **빈 줄·주석 줄·미지의 키를 구분하지 못한다** — 셋 다 같은 자리로 떨어진다. 즉 이것은 새로 계산해야 하는 판정이며, 프루너로 옮기면 다섯 갈래 분류의 두 번째 구현이 생겨 REQ-CGP-001 이 분류기에 대해 금지한 것과 같은 분기가 파싱 축에 열린다.

  근거는 §C.3 의 나머지 조항과 같다 — 판정이 서지 않으면 손대지 않는다. 이 조항이 없으면 세 가지 실제 손실이 열린다: (a) 항목 안의 여러 줄 리터럴은 파서의 `openDelim` 분기(`internal/codexwiring/skills.go:113-125`)가 통째로 건너뛰므로 범위를 잘못 끊어 닫는 `"""` 만 고아로 남긴다, (b) `path`/`enabled` 아닌 키는 파서가 아예 보지 않으므로(`skills.go:135-139`) 헤더만 지우면 그 대입이 **앞 테이블에 재귀속**되어 유효한 TOML 의 의미가 조용히 바뀐다, (c) `anyTableRe`(`internal/codexwiring/configtoml.go:76`, `^\[\[?[^\]]*\]\]?\s*(#.*)?$`)는 `[` 로 **시작하기만 하면** 맞으므로 `["x", "y"]` 같은 배열 이어쓰기 줄이 테이블 헤더로 읽혀 범위가 일찍 닫힌다.

  (d) **주석 하나가 등록 하나를 삼킨다.** `multilineOpener`(`skills.go:83-90`)는 주석을 벗겨내지 않고 줄 전체에서 `"""` 를 세므로, `"""` 를 홀수 번 담은 **주석 줄**이 리터럴을 연다. 그 리터럴이 뒤따르는 `[[skills.config]]` 헤더와 그 `path` 줄을 삼키면, 파서는 항목 **하나**만(앞의 유령) 보고하고 그 범위는 삼켜진 멀쩡한 등록을 품는다. 이때 범위를 **텍스트로** 다시 훑으면 헤더/`path`/주석 — 전부 "인식되는 모양"이라 탈락하지 않고, 범위를 지우면 **멀쩡한 등록이 함께 파괴된다.** 이것이 이 카드가 막으려는 바로 그 손실이다.

  (d)는 손으로 걸은 추론이 아니라 **실행으로 확인했다**(2026-09-07, 이 트리 `ace1c5440`). `ParseSkillEntries` 를 아래 7줄에 돌린 결과는 `ENTRIES=1`, `path="/gone"` 이었고, `/exists` 는 파서에 아예 보이지 않았다. 헤더가 없는 좁은 변형(삼켜진 구간이 맨 `path` 줄만 담은 것)도 같은 결과였다 — 즉 "범위당 헤더 하나"로 제한하는 미봉책은 이 구멍을 막지 못한다.

  ```toml
  [[skills.config]]        # L0
  path = "/gone"           # L1
  # uses """ in prose      # L2  ← 여기서 리터럴이 열린다
  [[skills.config]]        # L3  ← 삼켜짐
  path = "/exists"         # L4  ← 삼켜짐
  # and """ again          # L5  ← 리터럴이 닫힌다
  [other]                  # L6
  ```

  (a)와 (d)를 닫는 것은 다섯 갈래의 목록이 아니라 **"파서가 실제로 소비했는가"** 라는 판정식이다. 겉모양으로 판정하는 한 구멍은 사라지지 않고 자리만 옮긴다.

  탈락한 항목은 오류가 아니다. 보고에 "인식되지 않는 줄이 있어 건너뜀"으로 열거하고, 사용자가 손으로 판단하게 둔다.

> §C.3 의 일곱 조항이 이 SPEC 의 **안전 경계**다. 부재의 증거가 없는 것은 부재가 아니다 — 관측되지 않은 부재를 근거로 등록을 지우는 것은 멀쩡한 등록을 파괴하는 일이다.

### §C.4 경계는 테스트 의무다

- **REQ-CGP-012** (Ubiquitous) indeterminate 부류를 missing 으로 재분류하는 뮤턴트는 테스트 스위트를 **반드시 RED 로 만들어야 한다**. 경계를 지키는 조항이 문서에 있다는 사실은 경계가 지켜진다는 증거가 아니다. §A.1 이 보인 대로 살아 있는 설정은 이 부류를 한 번도 발화시키지 않으므로, 기계적 증거는 픽스처에서만 나온다.

### §C.5 쓰기 규율

- **REQ-CGP-013** (Ubiquitous) 기본 동작은 dry-run 이다. 실제 쓰기는 명시적 opt-in 플래그가 있을 때에만 일어난다.
- **REQ-CGP-014** (event-driven) 실제 쓰기가 일어날 때, 시스템은 쓰기 **전에** 설정 파일을 백업해야 하며, 백업 경로와 그 sha256 을 보고해야 한다. 그 한 쌍은 판정 기록에 그대로 옮겨 적을 수 있는 형태여야 한다.
- **REQ-CGP-015** (Ubiquitous) 제거 대상이 아닌 모든 내용 — 다른 테이블, 주석, 여러 줄 리터럴, 공백 — 은 바이트 단위로 보존되어야 한다. 파일 전체를 재직렬화해서는 안 된다: 이 저장소는 TOML 의존성을 갖고 있지 않으며, 재직렬화는 사용자가 손으로 쓴 파일 전체를 다시 쓰는 일이다. (이 기계의 설정은 629줄이고 항목 블록은 324줄에서 시작한다.)
- **REQ-CGP-016** (Ubiquitous) `internal/codexwiring/skills.go` 의 파서는 read-only 로 남아야 한다. 쓰기 능력은 별도 표면에 둔다 — 읽기와 쓰기의 경계가 눈에 보여야 한다.

- **REQ-CGP-019** (Ubiquitous) 시스템은 원본의 줄 끝 상태 — 마지막 줄의 개행 유무, 그리고 CRLF 여부 — 를 보존해야 한다. `splitLines`(`internal/codexwiring/configtoml.go:204-209`)는 `strings.TrimSuffix(body, "\n")` 를 하므로 `"a\n"` 과 `"a"` 가 같은 슬라이스를 낸다. 줄 끝 상태는 그 함수의 출력에 남지 않으므로 **따로 운반되어야 한다.**

### §C.6 명령 표면

- **REQ-CGP-020** (Unwanted) 시스템은 `--home` 과 `--codex-skills` 가 동시에 지정된 호출을 실행해서는 안 된다. 스코프는 하나이며, 동시 지정은 사용법 오류로 거절한다.
- **REQ-CGP-021** (Ubiquitous) 시스템은 명령의 도움말이 그 명령의 실제 영향 범위와 일치하도록 유지해야 한다. `internal/cli/clean.go:34` 의 "Only ~/.moai is touched — ~/.claude is never modified" 는 `--home` 이 유일한 홈 스코프이던 시절에 쓰인 문장이며, `~/.codex/config.toml` 을 바꾸는 스코프가 붙는 순간 명령 전체에 대해 거짓이 된다. 이 문장의 수정은 이 카드의 범위 **안**이다.

### §C.7 입력 부재

- **REQ-CGP-017** (state-driven) 홈이 해석되지 않는 동안, 설정 파일이 없거나 읽히지 않는 동안, 또는 선언된 항목이 0건인 동안, 시스템은 아무것도 보고하지 않고 아무것도 바꾸지 않아야 한다. `codexStaleSkillFinding` 의 fail-open 자세를 그대로 따른다.

## §D Exclusions

이 절은 이 카드가 **만들지 않는 것**을 못박는다.

### Out of Scope — 이 기계의 49건 실행

- 이 카드는 이 기계의 `~/.codex/config.toml` 에서 49건을 **제거하지 않는다**. 그 49건은 t504(경로 값 모양)와 t502(`[[skills.config]]` 생산자)의 살아 있는 관측 대상이다.
- 인도물은 **동사**이지 그 동사의 이 기계 대상 실행이 아니다.

### Out of Scope — 매니페스트 버전 스탬프

- `codex_measured_version` 0.147.0 → 0.153.4 재측정(`agents-codex.yaml:14`)은 이 카드의 범위 밖이며 별도 후속 카드로 분리되었다.

### Out of Scope — 생산자 쪽 수정

- 유령 항목이 애초에 왜 생겼는지, 누가 썼는지는 이 카드가 다루지 않는다(t502 소관). 이 카드는 이미 생긴 것을 지우는 쪽만 본다.

### Out of Scope — doctor 보고 표면 변경

- `codexStaleSkillFinding` 의 렌더링(요약만 나오고 per-class detail 은 나오지 않는 현재 동작)은 관측되었으나 이 카드에서 고치지 않는다.

### Out of Scope — Codex 자신의 정리 동작

- Codex 가 스스로 이 항목들을 정리하는지 여부는 측정되지 않았다. doctor 의 주석이 "정리하지 않는다"고 단정하지만 그 단정은 상속된 것이며 이 카드에서 재검증하지 않는다.

## §E 성공 판정

`acceptance.md` 의 AC 전부 PASS. 특히 AC-CGP-004(indeterminate 뮤턴트 RED)는 must-pass 이며, 다른 AC 가 전부 PASS 여도 이것이 FAIL 이면 SPEC 은 FAIL 이다.

## §F 교차 참조

- `.moai/reports/t506/baseline-measurement.md` — 측정 정본
- `.moai/reports/t506/observed-skill-paths.txt` — 관측된 49개 경로
- `internal/cli/doctor_codex.go` — 재사용할 분류기와 fail-open 자세
- `internal/cli/clean_home.go` — 파괴적 동사의 선례(허용목록 스캐너, 같은 순회 안의 가드, dry-run 기본)
- `internal/codexwiring/skills.go` — read-only 파서
