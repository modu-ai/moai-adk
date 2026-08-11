---
id: SPEC-GATE-ASTGREP-REPAIR-001
title: "Narrow the Over-Broad ast-grep Error-Wrapping Rule and Scope the moai gate ast-grep Scan — Plan"
version: "0.1.0"
status: in-progress
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

# 구현 계획 — SPEC-GATE-ASTGREP-REPAIR-001

## §A 맥락 (Context)

Tier **L (Large)**. 세 결함(D1 과잉 매칭 룰 / D2 whole-repo 스캔 + worktree 중복 + dedup 부재 / D3 moai gate gate.yaml 무시)을 단일 SPEC에서 해소한다. 이 세 결함은 `moai gate`가 22,293건의 가양성을 내는 결합된 원인이며, 어느 하나만 고쳐도 나머지 둘이 남아 시그널/노이즈 비가 회복되지 않으므로 단일 작업으로 통합했다.

### Tier 판정 근거 (S/M/L)

- **Tier L 확정**: 세 결합 결함이 ast-grep 룰 YAML + Go 소스(`internal/cli/gate.go`, `internal/hook/quality/astgrep_gate.go`, `internal/astgrep/scanner.go`) + 설정 파일(`.moai/config/sections/gate.yaml`, sgconfig.yml) + Template-First 미러링(`internal/template/templates/`)에 걸쳐 있다. 단일 도메인 단일 파일 변경이 아니며, REQ 10개 / AC 10개 이상을 예상한다. plan-auditor Tier L PASS threshold 0.85 대상.
- 본 plan-phase 산출물은 사용자 명시 요청에 따라 spec.md + plan.md + acceptance.md + progress.md 4개만 일차 저작한다. Tier L의 design.md / research.md는 plan-auditor 권고 또는 run-phase 진입 전 보강 필요 시 추가 저작한다(`tier: L` frontmatter는 그대로 유지).

### §A.1 산출물 범위

| 결함 | 주 변경 파일 | 본경로 |
|------|--------------|--------|
| D1 | `.moai/config/astgrep-rules/go/error-handling.yml` (룰 `go-error-not-wrapped` 정교화) + `internal/template/templates/.moai/config/astgrep-rules/go/error-handling.yml` (Template-First 미러) | 룰 YAML |
| D2 | `.moai/config/astgrep-rules/sgconfig.yml` (`globs` exclusion 추가) + (옵션) `internal/astgrep/scanner.go:419` `parseSGFindings` dedup | 설정 + (옵션) Go |
| D3 | `internal/cli/gate.go:57` (`DefaultGateConfig()` → `loadGateSection` 호출) | Go |

### §A.2 depends_on

- `depends_on: [SPEC-CONFIG-AUDIT-REPAIR-001]` — `internal/config/loader_gate.go` `loadGateSection` 로더의 원천 SPEC. 본 SPEC D3는 이 로더가 존재해야 재사용 가능하다 (verified 2026-08-11: `loader_gate.go` 존재, `loadGateSection` at line 14).

## §B 알려진 이슈 / 제약 (Known Issues & Constraints)

### B1. Template-First 미러 대상 결정 (dogfood vs distributed baseline)
- CLAUUDE.local.md §2.2는 로컬 `.moai/config/astgrep-rules/`를 "dogfood-experimental / local-only / 템플릿 미러 제외"로 명시. distributed baseline은 template-managed `go-hardcoding.yml`(루트) 단 1개.
- 그러나 본 SPEC이 정교화하는 `go-error-not-wrapped` 룰은 **dogfood 트리의 `go/error-handling.yml`**에 존재하며, distributed baseline `go-hardcoding.yml`에는 해당 룰이 없을 수 있다. [NEEDS CLARIFICATION: D1 정교화를 dogfood에만 적용하고 distributed baseline에는 새 룰을 추가하지 않을지, 아니면 distributed baseline에도 에러-래핑 룰을 신규 추가할지]. run-phase M1 착수 전 orchestrator가 AskUserQuestion으로 결정해야 한다 (D1 정교화가 distributed baseline에 영향을 주는지 여부).
- CLAUDE.local.md §25 template-neutrality (SPEC ID / 내부 작업 날짜 / commit SHA / ko 메시지 부제거)는 distributed baseline에 미러링하는 경우에만 바인딩된다.

### B2. `$ERR` 타입-제약의 ast-grep 표현 가능성
- ast-grep 메타변수는 타입을 직접 추론하지 않는다. 타입-제약은 다음 중 하나의 조합으로 표현해야 한다:
  - (a) `kind: identifier` + `regex`/`constraints`로 `err`/`errs` 이름 패턴 제한 — 간단하지만 우회적.
  - (b) `inside: { kind: if_statement, has: { pattern: if $A != nil } }` 관계 제약 — `if err != nil` 블록 안의 `return err`만 매칭. 더 정확하지만 구조가 복잡.
  - (c) (a)+(b) 조합.
- run-phase M1에서 세 후보를 positive/negative fixture(`return err`/`return nil`/`return 0`/`return true`/`return f()`)로 검증 후 최종 채택.

### B3. sgconfig.yml `globs` 지원 버전
- ast-grep의 `globs`는 비교적 최근 기능. `sg` 버전 0.40.5(확인됨)에서 지원 여부를 run-phase M2에서 먼저 검증해야 한다. 미지원 시 fallback은 `sg scan` 호출부에서 `--globs` CLI 플래그로 전달하거나 `parseSGFindings`에서 경로 기반 post-filter하는 것이다.
- `parseSGFindings`(`scanner.go:419`) post-filter가 가장 안전한 fallback — `sg` 버전에 의존하지 않으므로.

### B4. dedup vs globs 선택
- REQ-GAR-005는 (dedup 추가) 또는 (globs exclusion으로 동일 효과) 중 하나를 택하도록 허용. globs가 worktree를 근본적으로 제외하므로 dedup이 불필요해지는 것이 이상적이지만, 두 메커니즘이 중복 적용되면 위임-카운트가 부정확해질 수 있다. M2에서 어느 하나를 주 메커니즘으로 선택하고 다른 하나는 제거/보조로 둔다.

### B5. moai gate가 게이트 설정을 로드하지 못하는 경우 fallback
- gate.yaml이 없거나 파싱 불가인 경우, `loadGateSection`이 기본값을 반환하는지 아니면 에러를 반환하는지 확인 필요. 기본값 반환 시 advisory(defualt)로 폴백, 에러 반환 시 명시적 에러 처리. run-phase M3에서 로더 계약을 실측한다 (`loader_gate.go:14` 구독 + `loader_gate_test.go` TestLoadGateSection_* 테스트로 계약 확인).

### B6. 병렬 세션 레이스 (Pre-Spawn Sync Check)
- `.claude/worktrees/`가 4개 존재한다는 것 자체가 병렬 세션이 활동 중일 수 있음을 시사. run-phase spawn 전 Pre-Spawn Sync Check(`git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD`) + `moai session list --json --filter-spec=SPEC-GATE-ASTGREP-REPAIR-001` 필수. origin이 앞서거나 동일 SPEC 활동 세션이 있으면 STOP.

### B7. plan-phase에서 sg 실행 금지
- 본 plan-phase에서는 `sg scan`을 실행하지 않는다. 22,293건 카운트와 `return 0` 가양성은 run-phase M0에서 기계적으로 재관측한다 (verification-claim-integrity: 도구로 관측하기 전엔 결함 단언 금지).

## §C 사전 점검 (Pre-flight — run-phase 착수 전)

```bash
# 1. 현재 브랜치 + baseline
git branch --show-current
git rev-parse HEAD

# 2. 병렬 세션 레이스 점검 (Pre-Spawn Sync Check + multi-session)
git fetch origin main
git rev-list --count --left-right origin/main...HEAD
moai session list --json --filter-spec=SPEC-GATE-ASTGREP-REPAIR-001

# 3. worktree 존재 확인 (D2 영향 범위 파악)
ls .claude/worktrees/ 2>/dev/null | head -5
find .claude/worktrees -name '*.go' 2>/dev/null | wc -l

# 4. sg 설치/버전 확인 (B3)
sg --version

# 5. loader_gate.go 계약 확인 (B5)
grep -n "func.*loadGateSection" internal/config/loader_gate.go
go test -run TestLoadGateSection ./internal/config/... 2>&1 | tail -10

# 6. gate.yaml 현재 상태
cat .moai/config/sections/gate.yaml
```

## §D 기술 접근 (Technical Approach)

### D.1 D1 — ast-grep 룰 정교화 (M1)

1. **룰 패턴 후보 평가** (B2의 세 후보):
   - 후보 A: `kind: identifier` + `constraints: { regex: "^err(e?|s)$" }` — 이름 기반.
   - 후보 B: `inside: { kind: if_statement, has: { pattern: "if $ERR != nil" }, stopBy: end }` 관계 제약 — `if err != nil` 블록 내 `return $ERR`만 매칭.
   - 후보 C: A + B 조합.
2. **positive/negative fixture 작성**:
   - positive: `func f() error { err := g(); if err != nil { return err }; return nil }` → `return err`만 매칭.
   - negative: `return nil`, `return 0`, `return true`, `return f()`, `return 42` → 매칭 안 됨.
3. **`fix:` 가드**: `fix` 필드는 매칭 가드와 동일한 제약이 발화 조건이므로, 룰이 negative fixture를 매칭하지 않으면 fix도 자동으로 발화하지 않는다. 추가 가드가 필요 없는지 run-phase에서 확인 (REQ-GAR-003).
4. **Template-First 미러**: §B1의 결정에 따라 distributed baseline에 미러링. 미러 시 §25 template-neutrality 준수 (SPEC ID / 작업 날지 / commit SHA / ko 메시지 부제거).

### D.2 D2 — sgconfig globs + (옵션) dedup (M2)

1. **sgconfig.yml `globs` 추가**:
   ```yaml
   ruleDirs: [go, security]
   globs:
     - "!**/.claude/worktrees/**"
     - "!**/vendor/**"
     - "!**/*_test.go"
   ```
   - `!` 접두사로 제외 패턴 선언 (ast-grep glob 문서에 맞춤; run-phase에서 문법 검증).
2. **globs 지원 버전 점검**: sg 0.40.5에서 `globs`가 동작하는지 먼저 테스트. 미지원 시 fallback은 D.3의 post-filter.
3. **(옵션) dedup**: REQ-GAR-005의 대체 메커니즘. globs가 충분하면 dedup 불필요. globs 미지원 시 `parseSGFindings`(`scanner.go:419`)에서 `(file-basename, line, rule-id)` 키 기반 dedup 추가. 어느 하나만 주 메커니즘으로 채택 (B4).
4. **카운트 측정**: M0 baseline(22,293건) 대비 M2 적용 후 카운트 하락 관측.

### D.3 D3 — `moai gate` loadGateSection 호출 (M3)

1. **`internal/cli/gate.go:57` 수정**:
   ```go
   // Before:
   cfg := quality.DefaultGateConfig()
   // After (개념):
   cfg, err := loadGateConfigFromSection(projectDir)  // loadGateSection 재사용
   if err != nil {
       cfg = quality.DefaultGateConfig()  // 폴백 (B5)
   }
   ```
2. **로더 재사용**: PreToolUse 경로(`internal/hook/pre_tool.go:679-716`, `loadGateConfig`)와 동일한 `loadGateSection` 로더를 재사용하여 두 경로가 동일한 SSOT에서 설정을 읽도록 만든다.
3. **advisory 모드 검증**: gate.yaml `warn_only_mode: true`일 때 `moai gate`가 ast-grep 위반을 리포트하되 exit non-zero로 하드-블록하지 않는지 확인 (REQ-GAR-007).
4. **enabled=false 검증**: gate.yaml `enabled: false`일 때 ast-grep 서브-스캔이 건너뛰는지 확인 (REQ-GAR-008).
5. **hardcoding 준수** (CLAUDE.local.md §14): 게이트 임계값이나 config 키는 `internal/config/defaults.go`와 `internal/config/envkeys.go`에 이미 존재하는 상수를 재사용. 새 하드코드 불가.

## §E 자기 검증 (Self-Verification — plan-phase audit-ready signal)

plan-phase 산출물 완결성 체크:
- [ ] 12-필드 canonical frontmatter (4개 산출물 전부 status: draft)
- [ ] GEARS 요구사항 10개 (REQ-GAR-001..010), 각 기계 검증 가능
- [ ] Out of Scope 섹션 5개 `### Out of Scope —` H3 + `-` bullet (OutOfScopeRule 충족)
- [ ] acceptance.md에 mechanically-verifiable AC + 최소 2 Given-When-Then
- [ ] progress.md §E.1~§E.4 skeleton (§E.2~§E.4 placeholder)
- [ ] depends_on 명시 (SPEC-CONFIG-AUDIT-REPAIR-001)
- [ ] phase가 release target(`v3.1.0 target`) — lifecycle token(`plan`/`run`/`sync`/`mx`) 아님
- [ ] [NEEDS CLARIFICATION] B1 (dogfood vs distributed baseline 미러 대상) — orchestrator AskUserQuestion로 run-phase M1 착수 전 해소 필요

## §F 마일스톤 (Milestones — 우선순위 기반, 시간 추정 없음)

> 결정-가역성 역순: 가장 변경 가능성이 높은 결정(data-model/룰 정의/설정 경로)을 먼저.

- **M0 (Priority High) — characterization / reproduction (Reproduction-First, CLAUDE.md §7 Rule 4)**:
  - 현재 `moai gate` 카운트(~22,293건)를 baseline으로 캡처.
  - `return 0` / `return nil` / `return true` / `return f()`가 `go-error-not-wrapped`로 매칭되는지를 실패 테스트로 작성 (RED). fix 적용 시 `return 0`가 `fmt.Errorf(...)`로 감싸지는지도 실패 테스트로.
  - gate.yaml `warn_only_mode: true`임에도 `moai gate`가 exit non-zero로 하드-블록하는지를 실패 테스트로.
  - 이 테스트들은 M1~M3 적용 후 GREEN이 되어야 한다.

- **M1 (Priority High) — D1 룰 정교화** (가장 변경 가능성 높음 — 룰 정의):
  - D.1의 세 후보 평가 → 최종 패턴 채택.
  - `go-error-not-wrapped` 룰 정교화 + autofix 가드.
  - Template-First 미러 (B1 결정 후 distributed baseline 적용).
  - §25 template-neutrality 준수.

- **M2 (Priority High) — D2 스캔 범위**:
  - sgconfig.yml `globs` 추가 (D.2). sg 버전 호환성 검증.
  - (옵션) parseSGFindings dedup.
  - M0 baseline 대비 카운트 하락 관측.

- **M3 (Priority Medium) — D3 config wiring**:
  - `internal/cli/gate.go:57`에 `loadGateSection` 호출 추가 (D.3).
  - advisory 모드 / enabled=false 동작 검증 (REQ-GAR-007/008).

- **M4 (Priority Medium) — 통합 검증**:
  - `go test -count=1 ./...` full suite green.
  - `golangci-lint run --timeout=5m` green (또는 NEW vs baseline 구분).
  - `moai gate` smoke — 가양성 카운트가 예상 범위로 하락, `return <int>`/`return nil`/`return true` 가양성 0건, 진짜 `return err` 위만 매칭.
  - pre-commit이 `warn_only_mode: true`일 때 하드-블록하지 않는지 확인.

## §G 안티패턴 (Anti-Patterns — 하지 말 것)

- **AP-GAR-001**: ast-grep 스캐너(`internal/astgrep/scanner.go`)의 `Scan` 본경로 재작성. 본 SPEC은 최소 변경(globs / dedup / loadGateSection 호출)에 머문다 (REQ-GAR-010).
- **AP-GAR-002**: 룰 `fix:` 필드를 매칭 가드와 독립적으로 제거 — fix 자체를 제거하면 진짜 `return err` 위반에 대한 안전한 autofix도 사라진다. fix는 매칭 가드와 동일 발화 조건을 유지 (REQ-GAR-003).
- **AP-GAR-003**: distributed baseline 템플릿에 SPEC ID / 작업 날짜 / ko 메시지를 남기는 것 (§25 위반).
- **AP-GAR-004**: `moai gate`가 gate.yaml을 읽지 못할 때 조용히 hard-block으로 폴백하는 것 — 반드시 advisory 기본값으로 폴백 (B5).
- **AP-GAR-005**: plan-phase에서 `sg scan`을 실행하거나 22,293건 카운트를 관측 없이 사실로 단언하는 것 (B7, verification-claim-integrity).
- **AP-GAR-006**: `git add -A` / `git add .` — 병렬 세션이 worktree에 남긴 untracked 파일을 삼킬 수 있음. 경로-한정 add만.
- **AP-GAR-007**: dogfood 트리와 distributed baseline 템플릿을 혼동하여 한쪽 변경을 다른 쪽에 자동 전파 — §2.2 격리 정책 위반. B1 결정된 대상에만 적용.

## §H 교차 참조 (Cross-References)

- spec.md §A.2 (D1/D2/D3 세 결함), §B (REQ-GAR-001..010), §C (Out of Scope).
- `CLAUDE.local.md §2` (Template-First), `§2.2` (astgrep-rules local exception), `§6` (Testing), `§14` (hardcoding), `§25` (template-internal-isolation).
- `SPEC-CONFIG-AUDIT-REPAIR-001` (completed) — `loadGateSection` 로더 원천.
- `SPEC-ASTGREP-MULTILANG-001` (completed) — distributed baseline curated 접근.
- `SPEC-ASTGREP-DOGFOOD-CLEANUP-001` (draft) — 로컬 dogfood 트리 위생 (본 SPEC과 상호보완).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section A-E (Tier L 위임 템플릿).
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check (병렬 레이스 완화, B6).
