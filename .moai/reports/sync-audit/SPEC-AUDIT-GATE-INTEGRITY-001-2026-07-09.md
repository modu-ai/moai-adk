# Sync-Audit Evaluation Report — SPEC-AUDIT-GATE-INTEGRITY-001

- 감사 일시: 2026-07-09
- 감사자: sync-auditor (independent post-implementation assessment)
- 채점 모델: flat weighted-percentage (default) — harness.yaml `evaluator_mode`는 `final-pass`/`per-sprint`이며 `hierarchical` 미설정 (실측: `grep -n 'evaluator_mode' .moai/config/sections/harness.yaml` → L53 `final-pass`, L63 `per-sprint`)
- 평가 프로파일: `.moai/config/evaluator-profiles/default.md` (Functionality 40 / Security 25 / Craft 20 / Consistency 15)
- 대상 커밋: run 34523061b..e353d94a6 (M1-M5) / sync 61b7bcc0a / backfill 8cc3ed06e

## Evaluation Report
SPEC: SPEC-AUDIT-GATE-INTEGRITY-001
Overall Verdict: **PASS** (0.99 — weighted 98.7/100, weighted-harmonic 98.7/100)

### Dimension Scores

| Dimension | Score | Verdict | Evidence (실측 명령 — verbatim 로그는 아래 §Evidence Detail) |
|-----------|-------|---------|------|
| Functionality (40%) | 100/100 | PASS | 25/25 AC 전 행 독립 재실행 → progress.md §E.2 기록과 전 행 일치. `make build` exit=0 (fresh), `./bin/moai spec lint`(fresh binary) SPEC-범위 ERROR 0 |
| Security (25%) | 100/100 | PASS | doc-only 확인 (`git diff --name-only 34523061b^..e353d94a6` → .md 10개 + catalog.yaml, Go 소스 0), `git diff --stat -- go.mod go.sum` empty, secrets grep 4건 전수 검사 → 전부 OWASP-probe 문서 프로즈 자체, neutrality sum=0 |
| Craft (20%) | 95/100 | PASS | `go test -count=1 ./internal/template/` exit=0 (uncached, 1.128s), grep-stable 토큰 설계 준수, stale `:573` 인용을 content-token 앵커로 drift-proof 교정. Gap: golangci-lint 미재실행 (Go 무변경으로 정당화, §Gaps) |
| Consistency (15%) | 98/100 | PASS | live 4 ↔ mirror 4 토큰 동등 실측, 외부 앵커(evaluator-profiles → "Group 7 + Group 8") resolve 확인 (plan-auditor.md L301/L343 존재), out-of-scope 파일(spec-assembly.md) diff 0, frontmatter 3파일 무손상, close-subject full-ID 준수 |

### Must-Pass Results
- [PASS] Functionality — 모든 AC 충족 (partial credit 없음): 25/25 재실행 일치
- [PASS] Security — Critical/High finding 0건
- Hard threshold 발동 없음 → Overall FAIL override 없음

## AC 재실행 매트릭스 (독립 실측, 2026-07-09)

리포 루트 실행. 기록치 = progress.md §E.2.

| AC | 재실행 결과 | 기록치 | 판정 |
|----|------------|--------|------|
| AC-AGI-001 | 2 | 2 | PASS (≥2) |
| AC-AGI-002a | 2 | 2 | PASS (≥1) |
| AC-AGI-002b | 2 | 2 | PASS (≥1) |
| AC-AGI-003 | 3 / 3 | 3 / 3 | PASS (각 ≥2) |
| AC-AGI-004a | 4 | 4 | PASS (≥2) |
| AC-AGI-004b | 1 | 1 | PASS (≥1) |
| AC-AGI-005 | 1 | 1 | PASS (=1) |
| AC-AGI-006a | 1 | 1 | PASS (≥1) |
| AC-AGI-006b | 2/2/1/2 | 2/2/1/2 | PASS (각 ≥1) |
| AC-AGI-006c | 5 | 5 | PASS (≥1) |
| AC-AGI-007 | 1 / 1 | 1 / 1 | PASS (각 ≥1) |
| AC-AGI-008a | 2 / 2 | 2 / 2 | PASS (각 ≥1) |
| AC-AGI-008b | 1 / 1 | 1 / 1 | PASS (각 ≥1) |
| AC-AGI-009a | 1 | 1 | PASS (≥1) |
| AC-AGI-009b | 1 | 1 | PASS (≥1) |
| AC-AGI-010a | 1 | 1 | PASS (≥1) |
| AC-AGI-010b | 0 | 0 | PASS (=0) |
| AC-AGI-011 | 0 | 0 | PASS (=0) |
| AC-AGI-012a | 2/2/1/1 | 2/2/1/1 | PASS |
| AC-AGI-012b | 1/1/5 | 1/1/5 | PASS |
| AC-AGI-012c | 1 / 0 | 1 / 0 | PASS |
| AC-AGI-012d | 2/2/1/1 | 2/2/1/1 | PASS |
| AC-AGI-012e | sum=0 | 0 | PASS |
| AC-AGI-013 | make_build_exit=0 | exit=0 | PASS (fresh 재실행) |
| AC-AGI-014 | tool=0, SPEC-범위 ERROR 0 (self_lines=0) | tool=0, ERROR 0 + WARNING 1(E6 예외) | PASS — close 후 재실측에서는 StatusGitConsistency WARNING도 소멸(정상: run-push 전 기록 vs close 후 실측) |

25/25 PASS — 기록 주장과 실측 간 불일치 0건.

## R2 자기-정의 편집 평가 (sync-auditor.md — 본 감사자 자신의 파일)

1. **두-모델 선택 규칙 명확성**: 명확함. flat = default, `harness.yaml evaluator_mode: hierarchical`일 때만 계층 모델 — 조건이 단일 config 키에 결정적으로 바인딩되고 두 절이 상호 배타적. `sub-criteria refinement` 관계 서술로 두 모델이 차원 정체성을 공유함을 명시.
2. **worked example 정합성**: 산술 검증 완료 — Functionality min(0.75, 0.25)=0.25 < threshold 0.75 → must-pass FAIL → Overall FAIL. anchors 0.25/0.50/0.75/1.00만 사용, min-aggregation 표와 sub-criterion 표 일치, flat 출력 형식과 충돌 없음 (계층 형식은 hierarchical 모드에서 flat 표를 "대체"함을 명시).
3. **차원별 검증 명령 실행 가능성/언어중립성**: 전 명령 실재 toolchain (go test/pytest/npm test/cargo test 등), 4개 언어 동등 열거 + "no language is primary" 명시, 미설치 도구 graceful skip을 Gap으로 보고 의무화. **본 감사가 이 프로토콜을 그대로 적용해 수행됨** (Functionality: AC grep 배터리 + spec lint / Craft: go test / Security: grep probe + manifest diff) — 실행 가능성의 자기 실증.

## Findings (전수 — 신뢰도/심각도 병기)

- [LOW / pre-existing / confidence: medium] `.claude/agents/moai/sync-auditor.md:54-55` — 선택 규칙 트리거 `evaluator_mode: hierarchical`은 배포된 harness.yaml의 관측 값 집합(`final-pass`, `per-sprint`)에 없어 현행 config에서는 발화 불가능한 조건. HRN-003 시절부터의 pre-existing 상태(baseline grep 1)이며 본 SPEC은 profile-loader/Go 측을 명시적 Out of Scope 처리 — 회귀 아님. 후속 SPEC에서 harness.yaml enum에 `hierarchical` 추가 또는 "reserved value" 주석 권장.
- [INFO / pre-existing / confidence: high] `sync-auditor.md:114-117` Intervention Modes의 `per-iteration` vs harness.yaml thorough의 `per-sprint` 용어 드리프트. 본 SPEC 무접촉 영역.
- [INFO / confidence: high] 감사 시점 working tree의 `internal/template/catalog.yaml` dirty 상태는 병렬 세션의 template 편집(manager-docs.md/manager-git.md/llm.yaml) 유래 — 본 SPEC 커밋(61b7bcc0a)에 기인하지 않음.
- [INFO / resolved-in-observation] progress.md가 baseline으로 귀속한 `TestOutputStylesTemplateLiveParity` FAIL은 감사 시점 `-count=1` 재실행에서 소멸(전체 green) — 병렬 세션의 live moai-easy.md 착지로 해소된 것으로 관측. 조치 불요.

## Recommendations

- F-1 해소용 후속 SPEC: harness.yaml `evaluator_mode` enum에 `hierarchical` 정식 등재 또는 sync-auditor.md 트리거 절에 "config-reserved value" 주석.
- (재량) Intervention Modes `per-iteration` ↔ harness.yaml `per-sprint` 용어 통일.

## Evidence Detail (Claim / Evidence / Baseline / Gaps / Residual-risk)

**Claim**: SPEC-AUDIT-GATE-INTEGRITY-001의 25 AC 전 행 PASS, 12 REQ 충족, 회귀 없음.

**Evidence** (전부 본 감사 세션에서 직접 실행·관측):
- AC grep 배터리 2배치 (live 18행 + mirror 5행) — 출력 위 매트릭스 verbatim
- `make build` → `make_build_exit=0` (catalog hash regen 포함, bin/moai 재생성)
- `command -v moai` → `tool=0`; `./bin/moai spec lint`(fresh binary) → SPEC-범위 grep 라인 0, ERROR 0
- `go test -count=1 ./internal/template/` → `exit=0`, `ok ... 1.128s`
- `git diff --name-only 34523061b^..e353d94a6` → 11파일 전부 .md/catalog.yaml (Go 소스 0)
- `git diff --stat ... -- go.mod go.sum` → empty
- 사실관계 검증: `internal/spec/lint.go:649` `specIDPattern = ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (Bash `[0-9]{3}` 의미 동일 확인), `internal/runtime/audit_cache.go` `planArtifactNames` = {acceptance.md, plan.md, spec.md, tasks.md} 4-파일 세트 (REQ-AGI-009 서술과 일치), `.claude/skills/moai-ref-{owasp-checklist,testing-pyramid}` 실재, template 트리에 `verification-claim-integrity.md` + 두 skill 동봉 (mirror dangling-ref 없음)
- 로그 보존: /tmp/agi-audit-make-build.log, /tmp/agi-audit-lint.log, /tmp/agi-audit-lint-self.txt, /tmp/agi-audit-neutrality.txt, /tmp/agi-audit-template-tests2.log

**Baseline-attribution**: 모든 수치는 본 감사 세션(2026-07-09, HEAD=8cc3ed06e + 병렬 세션 working-tree 편집 잔존 상태)에서 신규 측정. progress.md §E.2 기록치는 비교 대상으로만 사용, 판정 입력으로 차용하지 않음.

**Gaps** (미검증 — 명시):
- `golangci-lint run` 미재실행 (Go 소스 무변경이 diff로 확증되어 생략; run-phase는 0 issues 주장)
- `GOOS=windows` cross-build 미재실행 (doc-only, B1 filter 정당 — run-phase와 동일 판단)
- 시나리오 1-4의 행동 검증(실제 plan-auditor가 D7 BLOCKING 상황에서 FAIL을 내는지)은 문서-층 배선 확인으로 갈음 — 에이전트 행동의 runtime 검증은 향후 실제 감사 발생 시에만 관측 가능
- `moai spec lint` exit=1 (repo-global)은 타 SPEC pre-existing ERROR 유래 — AC 정의상 판정 입력 아님(개별 원인 미조사)

**Residual-risk**:
- MP-5/MP-6는 문서-층 배선 — 기계 강제(lint rule/hook)는 Out of Scope로 미구현이므로 향후 plan-auditor 호출이 문서를 무시하면 재발 가능 (SPEC 자체가 이를 인지하고 후속 소관으로 명시)
- F-1: hierarchical 모드는 config에서 발화 불가 상태로 남음 — worked example의 실전 검증은 profile-loader 배선 후에만 가능

---
Verdict: **PASS** (0.99) — must-pass 2/2 충족, 발견 결함 중 본 SPEC 유래 신규 결함 0건.
