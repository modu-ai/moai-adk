---
id: SPEC-GATE-ASTGREP-REPAIR-001
title: "Narrow the Over-Broad ast-grep Error-Wrapping Rule and Scope the moai gate ast-grep Scan"
version: "0.1.0"
status: completed
created: 2026-08-11
updated: 2026-08-11
author: GOOS
priority: P1
phase: "v3.1.0 target"
module: "internal/cli, internal/hook/quality, .moai/config/astgrep-rules"
lifecycle: spec-anchored
tags: "astgrep, moai-gate, error-wrapping, false-positives, gate-config, config-behavior-consistency"
tier: L
---

# SPEC-GATE-ASTGREP-REPAIR-001 — ast-grep 과잉 매칭 룰 정교화 + moai gate 스캔 범위/설정 정합

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-11 | GOOS | 최초 초안 (draft). 세 가지 결함(D1 과잉 매칭 룰, D2 whole-repo 스캔 + worktree 중복 + dedup 부재, D3 moai gate gate.yaml 무시)을 단일 Tier L SPEC으로 통합. |

## §A 배경 (Context / Why)

`moai gate`는 매 호출마다 저장소 전체에서 약 **22,293건**의 "error not handled / not wrapped" 가양성(false-positive)을 발생시킨다. 이 수치는 카운터 버그가 아니라 결합된 세 가지 결함(D1/D2/D3)이 만들어낸 예상 범위의 결과다. 본 SPEC은 세 결함을 모두 해소하는 단일 Tier L 작업이다.

### §A.1 사용자 스토리 (User Story)

maintainer가 `moai gate`를 pre-commit 품질 게이트로 사용할 때, 게이트가 2만 건이 넘는 가양성을 뿜어내면 (a) 시그널이 노이즈에 묻혀 진짜 에러-래핑 결함을 찾을 수 없고, (b) WarnOnlyMode임에도 불구하고 블로킹 모드로 동작해 커밋이 차단되는 현상이 발생한다. 본 SPEC은 게이트가 진짜 `error` 타입 반환에 대해서만 경고하고, worktree 중복/테스트 파일을 스캔에서 제외하며, gate.yaml의 advisory 설정을 그대로 반영하도록 만든다.

### §A.2 세 결함의 원인 (verified 2026-08-11)

#### D1 — ast-grep 룰 패턴이 과잉 매칭

- 위치: `.moai/config/astgrep-rules/go/error-handling.yml`, 룰 `go-error-not-wrapped`
- 현재 룰 본문:
  ```yaml
  id: go-error-not-wrapped
  rule:
    pattern: return $ERR
  fix: 'return fmt.Errorf("TODO: operation: %w", $ERR)'
  ```
- 근본 원인: `$ERR`은 제약 없는 단일 표현식 메타변수. ast-grep에서 제약 없는 메타변수는 단일 AST 노드라면 무엇이든 매칭하므로 `return 0`, `return nil`, `return true`, `return f()` 모두 매칭된다. 메타변수 이름 `ERR`은 장식일 뿐, ast-grep은 메타변수 이름에서 타입을 추론하지 않는다.
- 위험: autofix가 비-에러 반환(`return nil`, `return true`, `return 0`)을 `fmt.Errorf(...)`로 감싸면 적용 시 파국적이다.

#### D2 — whole-repo 스캔 + worktree 중복 + dedup 부재

- 위치: `internal/hook/quality/astgrep_gate.go:73`이 `scanner.Scan(ctx, projectDir)` 호출; `internal/astgrep/scanner.go:306`이 `sg scan --config <sgconfig.yml> --json <projectDir>` 실행.
- 근본 원인: 스캔 대상이 staged-file 필터나 `git diff --cached` 없이 저장소 루트 전체. `.claude/worktrees/`가 거의 완전한 Go 트리 4개를 중복 보관 → Go 파일 2,011(실제) → 9,772(worktree 포함, 4.9×); `return`-포함 라인 17,224 → 83,953. dedup이 없어 동일 소스 라인이 메인 트리 1건 + 각 worktree 사본 N건 = 중복 매칭.
- `sgconfig.yml`(`.moai/config/astgrep-rules/sgconfig.yml`)은 `ruleDirs: [go, security]`만 선언 — `globs`/exclusions 없음. `walkSourceFiles`가 `_test.go`/vendor를 제외하긴 하나 이는 suppression 사전점검에만 쓰이고 `sg scan` 자체에는 영향을 주지 않는다.

#### D3 — `moai gate`가 gate.yaml을 무시 (config/behavior 불일치)

- 위치: `internal/cli/gate.go:57`이 `quality.DefaultGateConfig()` 하드코딩; `DefaultAstGrepGateConfig()`(`internal/hook/quality/astgrep_gate.go:121-128`)은 `BlockOnError: true, WarnOnlyMode: false` 반환. gate.go에는 `config.Load` / `loadGateSection` 호출이 전혀 없다.
- `.moai/config/sections/gate.yaml`은 `ast_grep_gate: { enabled: true, block_on_error: false, warn_only_mode: true }`(advisory 의도)로 설정. PreToolUse 경로(`internal/hook/pre_tool.go:679-716`, `loadGateConfig`)는 이를 읽지만 독립 실행형 `moai gate` CLI(pre-commit 경로)는 읽지 않는다 — advisory config와 모순되게 블로킹 모드로 동작.
- 해소 방향: `moai gate`가 gate.yaml의 `ast_grep_gate` 섹션을 로드하도록 만든다. `loadGateSection`(`internal/config/loader_gate.go:14`, SPEC-CONFIG-AUDIT-REPAIR-001 M2에서 배선)를 재사용하여 WarnOnlyMode/BlockOnError/Enabled를 config에서 가져오고 PreToolUse 경로와 일치시킨다.

## §B GEARS 요구사항 (Requirements)

> GEARS 표기 (current canonical). `<subject>`는 일반화된 명사 (ast-grep 룰 / 스캐너 / moai gate CLI / 게이트 설정). `[Where ...][While ...][When ...]` 수식어는 임의 부분집합 체이닝.

### D1 — ast-grep 룰 정교화

- **REQ-GAR-001** (Unwanted): `go-error-not-wrapped` 룰은 **When** `return nil` / `return 0` / `return true` / `return f()`와 같이 `error` 타입이 아닌 단일 리터럴/호출 반환문이 매칭되면, ast-grep 엔진은 이를 `go-error-not-wrapped` 위반으로 보고하지 **않아야 한다(shall not)**.

- **REQ-GAR-002** (Ubiquitous): `go-error-not-wrapped` 룰은 `error` 타입(또는 `error`를 포함한 멀티값)으로 추론 가능한 `return` 문에 한해서만 매칭해야 하며(shall), 룰 정의는 `kind`/`inside`/`has`/`constraints` 등 ast-grep 정규 관계 제약을 통해 이 타입-제약을 기계적으로 표현해야 한다(shall).

- **REQ-GAR-003** (Unwanted): `go-error-not-wrapped` 룰의 `fix:` 필드는 **When** 매칭된 노드가 `error` 타입이 아님이 확인되면, ast-grep 엔진은 `fmt.Errorf("TODO: operation: %w", $ERR)`로 감싸는 autofix를 적용하지 **않아야 한다(shall not)**. fix는 반드시 REQ-GAR-002의 타입-제약과 동일한 가드 아래에서만 발화해야 한다(shall).

### D2 — 스캔 범위 / dedup

- **REQ-GAR-004** (Capability gate): **Where** `sg scan`이 프로젝트 루트를 재귀 순회하는 경우, sgconfig.yml의 `globs`(또는 동등한 exclusion 메커니즘)은 `.claude/worktrees/**`, `vendor/**`, 테스트 파일(`*_test.go` 또는 테스트 디렉터리)을 ast-grep 스캔에서 제외해야 하며(shall), 이 제외는 `sg scan` 본경로에 적용되어야 한다(shall) (`walkSourceFiles` suppression 사전점검에만 적용되는 기존 동작과 구별).

- **REQ-GAR-005** (Event-driven): **When** 동일한 (file, line, rule) 트리플이 메인 트리와 worktree 사본 양쪽에서 매칭되면, ast-grep 게이트는 이를 단일 위반으로 보고해야 한다(shall) (parseSGFindings 단계에서 (file-basename, line, rule-id) 키 기반 dedup, 또는 REQ-GAR-004의 globs exclusion으로 동일 효과를 내는 단일 메커니즘 중 하나를 택한다; plan.md §D에서 결정).

### D3 — `moai gate` 설정 정합

- **REQ-GAR-006** (State-driven): **While** `moai gate` CLI가 독립 실행(pre-commit 경로)으로 동작하는 경우, moai gate CLI는 `quality.DefaultGateConfig()` 하드코딩 대신 `internal/config/loader_gate.go`의 `loadGateSection`을 통해 `.moai/config/sections/gate.yaml`의 `ast_grep_gate` 섹션을 로드해야 하며(shall), 로드 결과가 Enabled/BlockOnError/WarnOnlyMode의 SSOT가 되어야 한다(shall).

- **REQ-GAR-007** (Event-detected): **When** gate.yaml의 `ast_grep_gate.warn_only_mode`가 `true`로 설정된 경우, moai gate CLI는 ast-grep 위반을 발견해도 exit non-zero로 커밋을 차단하지 **않아야 한다(shall not)** (advisory 모드 — 위험은 리포트하되 하드-블록은 지양). 이 동작은 PreToolUse 경로(`internal/hook/pre_tool.go`의 `loadGateConfig`)와 정합해야 한다(shall).

- **REQ-GAR-008** (Event-detected): **When** gate.yaml의 `ast_grep_gate.enabled`가 `false`로 설정된 경우, moai gate CLI는 ast-grep 서브-스캔을 건너뛰어야 한다(shall) (명시적 opt-out 존중).

### 범위 가드 (scope guards)

- **REQ-GAR-009** (Where): **Where** 본 SPEC이 템플릿 소스(`internal/template/templates/`)에 미러링을 수행하는 경우(CLAUDE.local.md §2 Template-First), 미러 대상은 distributed baseline인 `internal/template/templates/.moai/config/astgrep-rules/`에 한하며(shall), CLAUDE.local.md §25 template-neutrality (SPEC ID / 내부 작업 날짜 / commit SHA / ko 메시지 부제거)를 준수해야 한다(shall). 로컬 dogfood 트리 `.moai/config/astgrep-rules/`는 §2.2 local-only 예외 아래 있으며, distributed baseline과 dogfood 트리 중 어느 쪽에 ast-grep 룰 정교화를 먼저 적용할지는 plan.md §D에서 결정한다.

- **REQ-GAR-010** (Unwanted): 본 SPEC은 `internal/astgrep/scanner.go`의 `sg scan` 호출부 재작성, `parseSGFindings`의 출력 스키마 변경, 또는 ast-grep 룰 엔진 자체 교체를 수행하지 **않아야 한다(shall not)** — 최소 변경(globs 추가, dedup 키 추가, 룰 패턴 정교화, loadGateSection 호출 추가)에 머물러야 한다(shall).

## §C 제외 범위 (Out of Scope)

본 SPEC이 의도적으로 **만들지 않는(NOT build)** 것들. 각 항목은 무엇을, 왜 제외하는지 명시한다.

### Out of Scope — 전체 ast-grep 스캐너 재작성
- `internal/astgrep/scanner.go`의 `Scan` 함수 본경로 재작성, `sg scan` 서브프로세스 호출을 다른 라이브러리로 교체, 또는 `parseSGFindings`의 JSON 출력 스키마를 변경하는 작업은 본 SPEC 범위 밖이다. 최소 변경(globs exclusion 추가, dedup 키 추가)에 머문다 (REQ-GAR-010).
- 16-언어 정식 룰셋 배포(SPEC-ASTGREP-DOGFOOD-CLEANUP-001 §C에서 이미 후속 SPEC로 지연된 작업)도 본 SPEC 범위 밖이다.

### Out of Scope — 게이트의 다른 단계(vet/lint/test)
- `moai gate`의 `go vet`, `golangci-lint`, `go test` 단계는 본 SPEC이 건드리지 않는다. 본 SPEC은 ast-grep 서브-게이트와 그 설정 로딩 경로만 다룬다.
- gate.yaml의 다른 섹션(non-ast_grep_gate)도 본 SPEC 범위 밖이다.

### Out of Scope — worktree 수명주기 변경
- `.claude/worktrees/`의 존재 자체를 없애거나 worktree 생성/철거 로직을 변경하는 작업은 본 SPEC 범위 밖이다. 본 SPEC은 ast-grep 스캔이 worktree를 *무시*하도록 만들 뿐, worktree 자체를 건드리지 않는다.

### Out of Scope — PreToolUse 경로 재작성
- `internal/hook/pre_tool.go`의 `loadGateConfig`(이미 gate.yaml을 읽고 있음) 재작성은 범위 밖이다. 본 SPEC은 독립 실행형 `moai gate` CLI 경로(`internal/cli/gate.go`)가 동일한 `loadGateSection` 로더를 재사용하도록 만드는 것만 다룬다 (REQ-GAR-006).

### Out of Scope — autofix 안전성 프레임워크 신설
- ast-grep `fix:` 필드의 타입-가드를 런타임에 검증하는 범용 프레임워크 신설은 본 SPEC 범위 밖이다. 본 SPEC은 `go-error-not-wrapped` 단일 룰의 `fix`가 매칭 가드와 동일한 제약 아래 발화하도록(REQ-GAR-003) 만들 뿐이다.

## §D 성공 기준 (Success Criteria)

- `return nil` / `return 0` / `return true` / `return f()`가 `go-error-not-wrapped` 위반으로 보고되지 않는다 (REQ-GAR-001).
- 진짜 `return err`(wrapping 없이)는 여전히 위반으로 보고된다.
- autofix가 비-에러 반환에 적용되지 않는다 (REQ-GAR-003).
- ast-grep 스캔이 `.claude/worktrees/**`와 테스트 파일을 제외한다 (REQ-GAR-004). 전체 가양성 카운트가 22,293건에서 관측 가능하게 하락한다.
- 동일 (file, line, rule) 위반의 중복 보고가 제거된다 (REQ-GAR-005).
- `moai gate`가 gate.yaml의 `ast_grep_gate` 섹션을 로드한다 (REQ-GAR-006).
- `warn_only_mode: true`일 때 커밋이 하드-블록되지 않는다 (REQ-GAR-007).
- 전체 `go test ./...` + `golangci-lint run` + `moai gate` smoke이 green이다.

## §E 교차 참조 (Cross-References)

- `CLAUDE.local.md §2` (Template-First Rule), `§2.2` (astgrep-rules local exception / distributed baseline = `go-hardcoding.yml`), `§6` (Testing — t.TempDir, 병렬 테스트에서 OTEL env 금지), `§14` (hardcoding — env 상수는 envkeys.go, 임계값은 defaults.go), `§25` (template-internal-isolation).
- `SPEC-CONFIG-AUDIT-REPAIR-001` (completed) — `internal/config/loader_gate.go` `loadGateSection`의 원천 SPEC. 본 SPEC D3는 이 로더를 재사용한다 (`depends_on`).
- `SPEC-ASTGREP-MULTILANG-001` (completed) — distributed baseline 템플릿 `[go, security]` curated 접근.
- `SPEC-ASTGREP-DOGFOOD-CLEANUP-001` (draft) — 로컬 dogfood 트리 정리 (본 SPEC과 상호보완; 본 SPEC은 룰 정교화 + gate 로딩, DOGFOOD-CLEANUP은 트리 위생).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix.
- `internal/spec/lint.go` `FrontmatterSchemaRule` / `OutOfScopeRule`.
