---
id: SPEC-ASTGREP-DOGFOOD-CLEANUP-001
title: "Local Dogfood ast-grep Ruleset Curated-Baseline Cleanup"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: GOOS
priority: P3
phase: "maintainer-tooling"
module: ".moai/config/astgrep-rules"
lifecycle: spec-anchored
tags: "astgrep, dogfood, cleanup, tooling, local"
---

# SPEC-ASTGREP-DOGFOOD-CLEANUP-001 — 로컬 dogfood ast-grep 룰셋 curated-baseline 정리

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-08 | GOOS | 최초 초안 (draft). CLAUDE.local.md §2.2가 후속 SPEC 소관으로 명시적 지연한 로컬 dogfood 트리 정리 작업의 착수 SPEC. |

## §A 배경 (Context / Why)

`CLAUDE.local.md §2.2`는 로컬 maintainer 저장소의 `.moai/config/astgrep-rules/` 트리를
**dogfood-experimental / local-only / 템플릿 미러 제외**로 명시하고, 그 정리를 명시적으로
후속 SPEC에 지연했다:

> "16-언어 정식 룰셋 배포(메시지 영어 통일 + 데모 stub → 실제 패턴 + utils 정리 + SPEC-ID
> strip + `sg` config-mode 검증)는 별도 후속 SPEC 소관."

본 SPEC이 그 지연된 정리 작업이다. 사용자가 저작을 명시적으로 승인했다.

### §A.1 완료된 자매 SPEC과의 경계 (SSOT)

**배포 대상 템플릿 트리 `internal/template/templates/.moai/config/astgrep-rules/`는 이미
깨끗하다** — 완료된 `SPEC-ASTGREP-MULTILANG-001` (`status: completed`, template-first
curated production baseline)이 이를 확정했다. 템플릿 트리의 검증된 형태:

- `sgconfig.yml` → `ruleDirs: [go, security]` (실제 vetted 룰을 최소 1개 담은 디렉터리만 선언)
- `go/hardcoding.yml`
- `security/{injection.yml, credentials.yml, crypto.yml}`

[HARD] 본 SPEC은 **템플릿 트리를 절대 건드리지 않는다**. 오직 로컬 maintainer 저장소의
dogfood 트리 `.moai/config/astgrep-rules/`에만 작용한다. 템플릿의 `[go, security]` curated
접근을 로컬 트리에 **미러(정렬)**하는 것이 본 SPEC의 패턴 원천이다.

### §A.2 실측된 로컬 dogfood 트리 현재 상태 (2026-07-08 검증)

manager-spec가 저작 전 실측한 상태 (git ls-files 41개 tracked + 1개 untracked):

| 항목 | 실측 상태 | 정리 성격 |
|------|-----------|-----------|
| `go-hardcoding.yml` (root, 1008B) | ORPHAN. SPEC-SLQG-001 provenance. 3룰 중 `go-no-raw-getenv`은 no-op. `go/hardcoding.yml`(1425B)가 대체(supersede)함 | 제거 대상 |
| `sgconfig.yml` (503B) | `ruleDirs` 18개 항목 — 존재하지 않는 `utils` 포함 + 15개 언어 디렉터리(대부분 stub) + 주석에 `SPEC-ASTG-UPGRADE-001` 내부 ID | ruleDirs 정렬 + SPEC-ID strip |
| `go/` (5 파일) | 실제 vetted 룰 (idioms, resource-safety, error-handling, hardcoding, concurrency). **메시지 한국어**. **오늘(02:29) 수정됨** | 유지 (내용 무접촉) |
| `security/` (5 파일) | 실제 vetted 룰 (secrets, injection, web, credentials, crypto). **메시지 한국어**. **오늘 수정 + `credentials.yml` untracked(in-flight)** | 유지 (내용 무접촉) |
| 빈 stub 디렉터리 9개 | cpp, flutter, java, javascript, python, r, rust, scala, swift — `.gitkeep`만 존재 | 제거 대상 |
| 데모 stub 디렉터리 5개 | csharp, elixir, kotlin, php, ruby — 동일 3룰명(null-deref/unused-var/todo-marker) 반복되는 미검증 boilerplate. **메시지 영어** | 제거 대상 |

**§2.2 기술과의 실측 차이 (교정)**: §2.2는 "메시지 언어 혼재 ko/en"이라 기술하나, 실측 결과
언어 분포는 반대다 — 삭제 대상 데모 stub은 **영어**, 유지 대상 go/security 실제 룰은 **한국어**.
따라서 "메시지 영어 통일"(§2.2 항목)은 데모 stub 제거로 트리 수준의 언어 혼재를 해소하되,
유지되는 go/security의 한국어 메시지 영어화는 **본 SPEC 범위 밖**으로 지연한다 (§C Out of Scope 참조).

### §A.3 위험 프레이밍 — 런타임 영향 없음 (low-risk hygiene)

ast-grep pre-tool gate는 기본 OFF (`AstGrepGate.Enabled` 컴파일 기본값 false; gate 로더 부재).
따라서 본 정리는 **런타임 동작 변경이 전혀 없다** — 순수 위생(hygiene) 작업이다. 실제 영향은
명시적 `moai ast-grep` CLI 경로 + `sg` 설치 + config-mode 실행 시에만 발생한다.
로컬 트리는 git-tracked이므로 모든 변경은 되돌릴 수 있다.

## §B GEARS 요구사항 (Requirements)

> GEARS 표기 (current). `<subject>`는 일반화된 명사 (룰셋 / sgconfig / dogfood 트리 / ast-grep 엔진).

### 정리 대상 (하이진 defect 해소)

- **REQ-ADC-001** (Ubiquitous): 로컬 dogfood ast-grep 룰셋은 하위 디렉터리 룰 파일에 의해
  대체(supersede)된 orphan root 룰 파일(`go-hardcoding.yml`)을 포함하지 **않아야 한다(shall not)**.

- **REQ-ADC-002** (Ubiquitous): `sgconfig.yml`의 `ruleDirs` 목록은 실제 vetted 룰을 최소 1개
  담은 디렉터리(`go`, `security`)만 선언해야 하며(shall), 완료된 템플릿 baseline을 미러해야 한다.

- **REQ-ADC-003** (Unwanted): `sgconfig.yml`은 디렉터리로 존재하지 않는 ruleDir(`utils`)를
  선언하지 **않아야 한다(shall not)**.

- **REQ-ADC-004** (Ubiquitous): 로컬 dogfood 트리는 `.gitkeep`만 담은 빈 언어 stub 디렉터리
  (cpp, flutter, java, javascript, python, r, rust, scala, swift)를 포함하지 **않아야 한다(shall not)**.

- **REQ-ADC-005** (Ubiquitous): 로컬 dogfood 트리는 미검증 데모 stub 언어 디렉터리
  (csharp, elixir, kotlin, php, ruby)를 포함하지 **않아야 한다(shall not)**.

- **REQ-ADC-006** (Unwanted): `sgconfig.yml`은 주석에 내부 SPEC-ID 참조
  (`SPEC-ASTG-UPGRADE-001`)를 포함하지 **않아야 한다(shall not)**.

### 검증 (verification)

- **REQ-ADC-007** (Event-driven): **When** maintainer가 정리 후
  `sg scan --config .moai/config/astgrep-rules/sgconfig.yml`을 실행하면, ast-grep 엔진은
  누락된 ruleDir 오류 없이 설정을 파싱해야 한다(shall).

### 범위 가드 (scope guards — 반드시 지켜야 할 경계)

- **REQ-ADC-008** (Where): **Where** 정리가 수행되는 곳에서, 모든 변경은 로컬
  `.moai/config/astgrep-rules/` 트리로 한정되어야 하며(shall), `internal/template/templates/`를
  수정하지 **않아야 한다(shall not)**.

- **REQ-ADC-009** (Unwanted / race guard): 정리는 활발히 수정 중인 `go/`·`security/` 룰 내용,
  그리고 in-flight untracked 파일(`security/credentials.yml`)을 삭제하거나 덮어쓰지
  **않아야 한다(shall not)**.

- **REQ-ADC-010** (Where): **Where** 로컬 트리가 `internal/template/templates/` 밖에 위치하므로,
  본 SPEC의 변경은 template-neutrality CI (`template-neutrality-check.yaml`)를 트리거하지
  않아야 하며(shall), 로컬 저장소 커밋으로만 반영되어야 한다(shall).

## §C 제외 범위 (Out of Scope)

본 SPEC이 의도적으로 **만들지 않는(NOT build)** 것들. 각 항목은 무엇을, 왜 제외하는지 명시한다.

### Out of Scope — Template tree
- 배포 대상 `internal/template/templates/.moai/config/astgrep-rules/`는 완료된
  `SPEC-ASTGREP-MULTILANG-001`이 이미 확정했으므로 본 SPEC은 절대 수정하지 않는다.
- 템플릿의 `sgconfig.yml` / `go/` / `security/` 어느 것도 건드리지 않는다.

### Out of Scope — Full 16-language vetted ruleset authoring
- §2.2의 "16-언어 정식 룰셋 배포"는 각 언어별 positive/negative fixture 검증을 요구하는
  대규모 배포(distribution) 노력으로, 배포 대상은 **템플릿 트리**이며 별도 후속 SPEC(Tier L) 소관이다.
- 본 SPEC은 로컬 dogfood 트리를 이미 검증된 curated baseline(`[go, security]`)로 **정렬**할 뿐,
  새로운 16개 언어의 실제 룰을 저작하지 않는다.

### Out of Scope — go/security 한국어 메시지 영어화
- 유지되는 로컬 `go/`·`security/` 룰의 한국어 메시지 영어 통일은 본 SPEC 범위 밖이다.
- 근거: (1) 로컬 dogfood 트리는 배포되지 않으므로 §25 template-neutrality가 바인딩하지 않음,
  (2) 해당 디렉터리가 오늘 활발히 수정 중(untracked `credentials.yml` in-flight)이라 병렬 세션
  레이스 위험이 있음. 영어화는 이 룰들이 템플릿으로 승격될 때 배포 SPEC이 처리한다.

### Out of Scope — ast-grep gate 활성화 및 Go 코드
- `AstGrepGate.Enabled` 활성화, gate 로더 추가, `moai ast-grep` CLI 동작 변경 등 Go 코드
  (`internal/`/`pkg/`/`cmd/`) 변경은 본 SPEC 범위 밖이다. 본 SPEC은 순수 파일/설정 위생 작업이다.

## §D 성공 기준 (Success Criteria)

- 로컬 dogfood 트리가 `[go, security]` curated baseline으로 정렬됨 (템플릿 접근 미러).
- orphan / 빈 stub / 데모 stub / 존재하지 않는 ruleDir / 내부 SPEC-ID가 모두 제거됨.
- `sg scan --config .../sgconfig.yml`이 누락 ruleDir 오류 없이 파싱됨.
- go/security 실제 룰 내용과 in-flight untracked 파일이 보존됨.
- 템플릿 트리 무접촉 + template-neutrality CI 미트리거 확인.

## §E 교차 참조 (Cross-References)

- `CLAUDE.local.md §2.2` — 로컬 dogfood 트리 격리 정책 + 본 정리 작업의 지연 명시(원천).
- `CLAUDE.local.md §15/§25` — 언어 중립성 / template-internal-isolation (템플릿에만 바인딩; 로컬 예외).
- `SPEC-ASTGREP-MULTILANG-001` (completed) — 템플릿 `[go, security]` curated baseline (미러 패턴 원천).
- `internal/template/split_namespace_test.go` / `internal_content_leak_test.go` — 템플릿 격리 CI 가드
  (로컬 트리는 이 가드 범위 밖 — 위반 없음 확인용).
