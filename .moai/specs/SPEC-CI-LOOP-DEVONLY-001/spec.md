---
id: SPEC-CI-LOOP-DEVONLY-001
title: ci-loop 스크립트 의존성의 배포 트리 격리 및 dev-only 전환
version: 0.3.0
status: in-progress
created: 2026-08-01
updated: 2026-08-02
author: manager-spec
priority: HIGH
phase: plan
module: template
lifecycle: spec-anchored
tags: "template, distribution, ci-loop, dev-only, constitution, neutrality"
tier: M
---

## HISTORY

- 2026-08-01 — plan-phase 최초 저작 (v0.1.0).
- 2026-08-02 — 독립 감사 FAIL(0.56) 반영 재작성 (v0.2.0). `moai pr watch`가 watch 루프를
  수행한다는 v0.1.0의 핵심 근거가 실행 검증으로 반증되어 §1·REQ-CLD-005·§4를 재도출.
  GOOS 결정 A(두 프로토콜 파일 분리 처리)·B(required-checks.yml 이연) 반영.
- 2026-08-02 — run-phase 진입 시점 GOOS 결정 반영 개정 (v0.3.0). 이연했던 판정 2건을
  승격: `moai pr watch` 동작 불변(AC-CLD-018, M6)과 `archiveSkill` 보존 동작
  (AC-CLD-019, M5). 미커버 REQ 0건, `archiveVersion` 미결 결정 종결.
  요구사항·범위 제외는 불변이며 판정 계층만 확장했다.

## 1. 배경 (Context)

배포 템플릿 트리(`internal/template/templates/`)는 `scripts/ci-watch/` 및 `scripts/ci-autofix/`
경로의 셸 스크립트를 참조한다. 그러나 해당 스크립트는 개발 저장소 루트에만 존재하고
템플릿 트리에는 포함되지 않는다. 결과적으로 배포된 모든 사용자 프로젝트는
**실행 파일이 존재하지 않는 워크플로를 강제하는 문서와 Frozen HARD 규칙**을 함께 받는다.

측정된 표면 (실측 근거는 research.md §A):

- 스크립트 경로를 참조하는 템플릿 파일: 8개, 총 27개 참조
- 스킬 식별자 `moai-workflow-ci-loop`를 참조하는 템플릿 항목 + `catalog.yaml`: 9개
- 합집합: 템플릿 측 14개, 개발 저장소 미러 측 13개
- `zone-registry.md` 전체: Frozen 항목 **71개**, `canary_gate: true` **73개**.
  이 중 ci-loop 귀속분은 `CONST-V3R5-004..021`의 **18개 부분집합**이다
  (18은 전체가 아니라 부분집합이라는 점이 v0.1.0에서 부정확하게 서술되었다)
- 실제 스크립트: 9개 파일

### 1.1 능력의 실제 소재 (v0.1.0 근거 반증)

v0.1.0은 `moai pr watch --help`의 서술을 근거로 "watch 능력은 CLI로 배포되며 스크립트는
얇은 래퍼"라고 판단했다. 이 판단은 **명령을 실행한 결과 반증되었다** (research.md §B):

```
$ moai pr watch 999 --branch main
[ci-watch] Use scripts/ci-watch/run.sh to start the watch loop.
EXIT=0
```

`internal/cli/pr_watch_cmd.go`의 기본 모드는 **셸 스크립트를 실행하라는 안내문을 출력하고
종료**한다. 폴링 루프, 30초 주기, 30분 타임아웃, `exit 2`는 전부
`scripts/ci-watch/run.sh`(219행)에 있으며 CLI에는 `os.Exit(2)`가 존재하지 않는다.
CLI가 실제로 제공하는 것은 `--abort`(상태 파일 플래그 설정)와 `--report`(보고서 서식화)뿐이다.

`--help`의 `Long` 텍스트는 구현되지 않은 계약을 서술한다. **도움말 산문은 구현의 증거가 아니다.**

따라서 배포된 사용자는 watch 능력을 **가지고 있지 않다**. 이는 v0.1.0의 결론을 뒤집으며,
`ci-watch-protocol.md`와 `ci-autofix-protocol.md`의 처리 방향이 달라지는 근거가 된다(§2).

## 2. 목적 (Purpose)

배포 트리에서 개발 전용 셸 스크립트 의존성과 스킬 래퍼를 제거한다.
두 프로토콜 파일은 실행 주체의 존재 여부에 따라 다르게 처리한다:

- `ci-watch-protocol.md` — **배포 중단**. 배포되는 어떤 산출물도 watch 루프를 수행하지 않으므로,
  수행 주체 없는 동작을 규율하는 규칙을 배포할 이유가 없다.
- `ci-autofix-protocol.md` — **유지하되 스크립트 의존성만 제거**. `cycle_type=autofix`는
  `manager-develop`의 실재하는 사이클이며, 그 반복 상한·에스컬레이션·보호 파일 규칙은
  스크립트와 무관하게 구속력을 갖는다.

ci-loop 오케스트레이션 자산은 기존 유지관리자 전용 하네스와 동일 등급의 dev-only 자산으로 격리한다.

## 3. 요구사항 (GEARS)

### 3.1 배포 트리 격리

**REQ-CLD-001** — The template tree shall contain no reference to `scripts/ci-watch`
or `scripts/ci-autofix`.

**REQ-CLD-002** — The template tree and the template catalog shall contain no reference
to the skill identifier `moai-workflow-ci-loop`.

**REQ-CLD-003** — The embedded template filesystem compiled into the binary shall contain
no file under `.claude/skills/moai-workflow-ci-loop/`.

**REQ-CLD-004** — **Where** the template catalog declares a skill entry, the catalog shall
declare no entry whose `name` is `moai-workflow-ci-loop`.

**REQ-CLD-005** — The template tree shall contain no file at
`.claude/rules/moai/workflow/ci-watch-protocol.md`.

**REQ-CLD-006** — **When** the catalog hash generator is run against the post-change template
tree, it shall produce no modification to the committed `catalog.yaml`.

### 3.2 능력 서술의 정직성

**REQ-CLD-007** — The distributed rule tree shall not assert that the shipped binary performs
the CI watch loop.

**REQ-CLD-008** — The shipped binary shall not emit user-visible text instructing the user to
run a shell script that is absent from the distribution.

**REQ-CLD-009** — **While** `cycle_type=autofix` remains a declared `manager-develop` cycle,
the autofix safety rules — iteration limit, escalation contract, semantic-failure handling,
and protected-file list — shall remain expressed in the distributed rule tree.

**REQ-CLD-010** — The autofix protocol shall express its entry trigger without naming a shell
script path as a precondition.

### 3.3 헌법(Frozen 존) 정합성

**REQ-CLD-011** — **When** a Frozen clause's source text is amended or its source file is
removed, the corresponding `zone-registry.md` entries shall be amended or removed in the same
change.

**REQ-CLD-012** — The zone registry shall contain no entry whose `file` points to a source
file absent from the same tree.

**REQ-CLD-013** — **When** `moai constitution validate` is run after the change, the count of
findings attributed to `ci-autofix-protocol.md` or `ci-watch-protocol.md` shall be zero.

**REQ-CLD-014** — The repo-root registry and the template registry shall carry identical
clause text for every retained ci-attributed entry.

### 3.4 개발 전용 자산 보존

**REQ-CLD-015** — The development repository shall retain all nine ci-loop shell scripts and
the ci-loop skill outside the template tree.

**REQ-CLD-016** — **Where** a dev-only asset exists, the template tree shall carry no trace of it.

### 3.5 배포된 프로젝트에 대한 영향

**REQ-CLD-017** — **When** a user who previously received the skill runs the update command,
the system shall archive the retired skill directory rather than leaving it unmanaged.

**REQ-CLD-018** — The system shall not delete user-authored content while retiring the skill
directory.

### 3.6 중립성

**REQ-CLD-019** — The template tree shall remain neutral across all sixteen supported
languages, and shall carry no internal SPEC identifier, requirement token, internal date,
or commit hash.

## 4. 범위 제외 (Exclusions)

### Out of Scope — CLI 동작 변경

- `moai pr watch`의 플래그 집합, 종료 코드, `--abort` / `--report` 동작 변경
- watch 루프를 Go로 구현해 CLI에 이식하는 작업 (별도 SPEC 소관)
- 이 SPEC이 허용하는 유일한 `internal/cli/pr_watch_cmd.go` 변경은 **문자열 수정**이며
  (REQ-CLD-008), 동작·플래그·종료 코드는 불변이다

### Out of Scope — 스크립트 자체의 재작성

- `scripts/ci-watch/lib/classify.sh`의 Go 전용 도구 하드코딩을 16개 언어 대응으로 일반화
- 스크립트를 템플릿에 포함시키기 위한 언어 중립화 (명시적 기각)

### Out of Scope — required-checks.yml 부재 결함 (이연)

- `.github/required-checks.yml`은 템플릿 트리에 없고, 사용자 프로젝트에서 이를 생성하는
  경로를 확인하지 못했다. 이는 이 SPEC이 제거하려는 것과 **같은 종류의 결함**이다.
- **이연 근거**: REQ-CLD-005에 따라 `ci-watch-protocol.md`가 배포에서 제거되면 이 전제 조건을
  요구하던 파일 자체가 사라진다. 실측 결과 배포 트리에서 이 파일을 요구하는 나머지 참조는
  `zone-registry.md`(M2가 제거), ci-loop `SKILL.md`(M3가 제거),
  `sync/delivery.md:460`의 ci-loop 스킬 서술(M3가 제거) 뿐이며,
  **그 외 어떤 배포 파일도 이 전제를 요구하지 않는다** (research.md §G.1).
  즉 이 결함은 본 SPEC 완료 시점에 잔여 참조 0으로 소멸한다. 소멸하지 않는 잔여가
  발견되면 그때 별도 SPEC으로 승격한다.

### Out of Scope — 기존 헌법 드리프트 일괄 해소

- `moai constitution validate`가 보고하는 ci-loop 무관 드리프트 59건
- `constitution-check` CI 잡을 자문 모드에서 차단 모드로 전환하는 작업

### Out of Scope — 이미 배포된 프로젝트의 소급 정리

- 사용자가 수정한 스킬 사본의 내용 병합 또는 마이그레이션
- `.moai/archive/` 버전 태그 체계의 재설계

## 5. 참조

- research.md — 4개 질문의 실측 근거와 명령별 원문 출력, 반증 기록
- plan.md — 마일스톤, EXTEND 봉투, 위험
- acceptance.md — 판정 명령과 관측된 baseline
