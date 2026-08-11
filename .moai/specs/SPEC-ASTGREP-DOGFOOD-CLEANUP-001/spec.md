---
id: SPEC-ASTGREP-DOGFOOD-CLEANUP-001
title: "Local Dogfood ast-grep Ruleset Curated-Baseline Cleanup"
version: "0.2.0"
status: completed
created: 2026-07-08
updated: 2026-08-12
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
| 0.2.0 | 2026-08-12 | GOOS | plan refresh. REQ-ADC-002/003 은 `SPEC-CONFIG-AUDIT-REPAIR-001` M2 (commit 6e70a29fd, PR #1142, 2026-07-25) 가 해소 → 제거. 잔여 REQ-001/004/005/006 (순수 위생) 으로 축소. REQ-006 제거 대상을 현재 누출 `SPEC-GATE-ASTGREP-REPAIR-001` (sgconfig.yml:10, #1445 체인이 추가) 로 갱신. REQ-004 빈 stub 개수 9→10 (typescript 누락 정정). §A.3 gate 프레이밍을 #1142 이후 실제(warn-only ON)로 정정. |

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

### §A.2 실측된 로컬 dogfood 트리 현재 상태 (2026-08-12 재검증)

> **v0.2.0 refresh (2026-08-12)**: 최초 실측(2026-07-08) 이후 `SPEC-CONFIG-AUDIT-REPAIR-001` M2
> (#1142)가 sgconfig `ruleDirs` 정렬 + phantom `utils` 제거를 이미 수행했다. 본 테이블은 현재 상태를
> 반영하며, 이미 해소된 항목은 ✅로 표시한다.

| 항목 | 실측 상태 (2026-08-12) | 정리 성격 |
|------|------------------------|-----------|
| `go-hardcoding.yml` (root) | ORPHAN. Go 코드·config 어디서도 참조 안 함(grep 검증). `go/hardcoding.yml`이 동일 규칙을 큐레이션 세트로 대체. sgconfig.yml:7 주석만이 존재를 언급 | 제거 대상 (REQ-001) |
| `sgconfig.yml` ruleDirs | ✅ **이미 `[go, security]`로 정렬됨** (#1142). phantom `utils` 제거됨. → REQ-002/003 해소 | (해소됨) |
| `sgconfig.yml` 주석 | 주석에 `SPEC-GATE-ASTGREP-REPAIR-001` 내부 ID 잔존 (line 10, #1445 체인이 추가). 원본 `SPEC-ASTG-UPGRADE-001`은 #1142로 제거됐으나 새 ID가 추가됨 → AC-006 여전히 FAIL | SPEC-ID strip (REQ-006) |
| `go/` (5 파일) | 실제 vetted 룰 (idioms, resource-safety, error-handling, hardcoding, concurrency). **메시지 한국어** | 유지 (내용 무접촉, REQ-009) |
| `security/` (5 파일) | 실제 vetted 룰 (secrets, injection, web, credentials, crypto). **메시지 한국어** | 유지 (내용 무접촉, REQ-009) |
| 빈 stub 디렉터리 10개 | cpp, flutter, java, javascript, python, r, rust, scala, swift, **typescript** — `.gitkeep`만 존재. (v0.1.0의 9개 목록에서 `typescript`가 누락됐었음 — 정정) | 제거 대상 (REQ-004) |
| 데모 stub 디렉터리 5개 | csharp, elixir, kotlin, php, ruby — 동일 3룰명(null-deref/unused-var/todo-marker) 반복되는 미검증 boilerplate. **메시지 영어** | 제거 대상 (REQ-005) |

**§2.2 기술과의 실측 차이 (교정)**: §2.2는 "메시지 언어 혼재 ko/en"이라 기술하나, 실측 결과
언어 분포는 반대다 — 삭제 대상 데모 stub은 **영어**, 유지 대상 go/security 실제 룰은 **한국어**.
따라서 "메시지 영어 통일"(§2.2 항목)은 데모 stub 제거로 트리 수준의 언어 혼재를 해소하되,
유지되는 go/security의 한국어 메시지 영어화는 **본 SPEC 범위 밖**으로 지연한다 (§C Out of Scope 참조).

### §A.3 위험 프레이밍 — 런타임 영향 없음 (low-risk hygiene)

> **v0.2.0 정정**: v0.1.0은 "gate 기본 OFF (`AstGrepGate.Enabled` 컴파일 기본값 false; gate 로더 부재)"라
> 기술했으나, `SPEC-CONFIG-AUDIT-REPAIR-001` M2 (#1142)가 gate 로더(`loader_gate.go`)를 추가하면서
> 이 전제가 바뀌었다. 현재 기본값은 `AstGrepGate{Enabled: true, BlockOnError: false, WarnOnlyMode: true}`
> (`internal/config/defaults.go`) — 즉 **차단 없는 권고 모드로 켜져 있고**, 차단(blocking)만이
> `gate.yaml` opt-in이다. (CLAUDE.local.md §2.2 2026-08-01 정정 참조.)

결론은 동일하다: gate가 warn-only 모드이므로 본 정리는 **런타임 동작 변경이 전혀 없다** — 순수
위생(hygiene) 작업이다. 오히려 invalid rule을 가진 빈/데모 stub 디렉터리를 제거하면 config-mode
스캔 파싱이 더 안정적이 된다(grep 검증: ruleDirs는 이미 `[go, security]`로 축소돼 stub는 sg scan이
무시하지만, orphan root 파일은 여전히 잡음). 로컬 트리는 git-tracked이므로 모든 변경은 되돌릴 수 있다.

## §B GEARS 요구사항 (Requirements)

> GEARS 표기 (current). `<subject>`는 일반화된 명사 (룰셋 / sgconfig / dogfood 트리 / ast-grep 엔진).

### 정리 대상 (하이진 defect 해소)

> **v0.2.0**: REQ-ADC-002 (ruleDirs `[go, security]` 정렬) 와 REQ-ADC-003 (phantom `utils`
> 제거) 은 `SPEC-CONFIG-AUDIT-REPAIR-001` M2 (commit 6e70a29fd, PR #1142, 2026-07-25) 가
> 이미 해소했다. 본 SPEC에서 제거하며, 번호 공간은 추적성 보존을 위해 비워둔다 (재번호화 없음).
> 잔여 정리 대상은 아래 REQ-001/004/005/006 의 4개.

- **REQ-ADC-001** (Ubiquitous): 로컬 dogfood ast-grep 룰셋은 하위 디렉터리 룰 파일에 의해
  대체(supersede)된 orphan root 룰 파일(`go-hardcoding.yml`)을 포함하지 **않아야 한다(shall not)**.

- ~~**REQ-ADC-002**~~ (RESOLVED by #1142): `sgconfig.yml`의 `ruleDirs`는 이미 `[go, security]`로 정렬됨.

- ~~**REQ-ADC-003**~~ (RESOLVED by #1142): phantom `utils` ruleDir는 이미 제거됨.

- **REQ-ADC-004** (Ubiquitous): 로컬 dogfood 트리는 `.gitkeep`만 담은 빈 언어 stub 디렉터리
  (cpp, flutter, java, javascript, python, r, rust, scala, swift, **typescript**)를 포함하지
  **않아야 한다(shall not)**. (v0.2.0 정정: `typescript`가 v0.1.0 목록에서 누락됐었음.)

- **REQ-ADC-005** (Ubiquitous): 로컬 dogfood 트리는 미검증 데모 stub 언어 디렉터리
  (csharp, elixir, kotlin, php, ruby)를 포함하지 **않아야 한다(shall not)**.

- **REQ-ADC-006** (Unwanted): `sgconfig.yml`은 주석에 내부 SPEC-ID 참조를 포함하지
  **않아야 한다(shall not)**. (v0.2.0 갱신: 원본 누출 `SPEC-ASTG-UPGRADE-001`은 #1142로
  제거됐으나, `SPEC-GATE-ASTGREP-REPAIR-001` (line 10, #1445 체인이 2026-08-11 추가) 가 새로
  잔존한다. run-phase는 이 접두사를 strip하되 기술 노트 본문은 보존한다.)

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

### Out of Scope — ast-grep gate blocking 활성화 및 Go 코드
- gate 로더 자체는 `SPEC-CONFIG-AUDIT-REPAIR-001` M2 (#1142) 가 이미 추가했다(`loader_gate.go`).
  본 SPEC 범위 밖인 것은: gate **차단(blocking) 활성화**(`gate.yaml` opt-in), `moai ast-grep` CLI 동작
  변경 등 Go 코드 (`internal/`/`pkg/`/`cmd/`) 변경. 본 SPEC은 순수 파일/설정 위생 작업이다.

## §D 성공 기준 (Success Criteria)

- ✅ (이미 달성) 로컬 dogfood 트리의 `sgconfig.yml` `ruleDirs`가 `[go, security]`로 정렬됨 — #1142.
- orphan root 파일(`go-hardcoding.yml`) / 빈 stub 10개 / 데모 stub 5개 디렉터리가 모두 제거됨.
- `sgconfig.yml` 주석의 내부 SPEC-ID 참조가 제거됨 (기술 노트 본문은 보존).
- `sg scan --config .../sgconfig.yml`이 누락 ruleDir 오류 없이 파싱됨.
- go/security 실제 룰 내용이 보존됨 (무접촉).
- 템플릿 트리 무접촉 + template-neutrality CI 미트리거 확인.

## §E 교차 참조 (Cross-References)

- `CLAUDE.local.md §2.2` — 로컬 dogfood 트리 격리 정책 + 본 정리 작업의 지연 명시(원천).
- `CLAUDE.local.md §15/§25` — 언어 중립성 / template-internal-isolation (템플릿에만 바인딩; 로컬 예외).
- `SPEC-ASTGREP-MULTILANG-001` (completed) — 템플릿 `[go, security]` curated baseline (미러 패턴 원천).
- `SPEC-CONFIG-AUDIT-REPAIR-001` (M2, #1142, commit 6e70a29fd) — REQ-ADC-002/003 해소 원천
  (ruleDirs 정렬 + phantom utils 제거 + gate 로더 추가). 본 SPEC의 §A.3 gate 프레이밍 정정 근거.
- `SPEC-GATE-ASTGREP-REPAIR-001` (#1445) — go/error-handling.yml 정교화 + sgconfig.yml:10 기술 노트
  추가 (REQ-006의 새 SPEC-ID 누출 발생원).
- `internal/template/split_namespace_test.go` / `internal_content_leak_test.go` — 템플릿 격리 CI 가드
  (로컬 트리는 이 가드 범위 밖 — 위반 없음 확인용).
