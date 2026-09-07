# SPEC-REMOVAL-GUARD-EXTRAS-001 — Implementation Plan

Card t511 · Branch `WT-danger-guard-regex` · Base `0b1e27877` (`origin/develop`) · Tier M · Harness standard

## §A Context

### §A.1 Mission summary

배포되는 `security.yaml` extras 목록에 남아 있는 초월된 정규식 `rm\s+-rf\s+/[^.]`을 3개 사본에서 제거하고, 배포 policy(기본 + extras 병합)를 exerc하는 양방향 회귀 표면을 신설한다. 본체 구조 체크(t286, `034dad48e`)는 PRESERVE — 코드 로직 변경은 0건이다. RED-now 증거는 plan 페이즈에서 3클래스 라이브 재현으로 확보됐다 (research.md §R).

### §A.2 Worktree + branch + base

- Worktree: `.claude/worktrees/t511` (L1)
- Branch: `WT-danger-guard-regex` (develop에서 분기)
- Base: `0b1e27877` — plan-phase 개시 시 `git rev-parse --short HEAD` 실측값. 런 페이즈 진입 전 재판독 (staleness rule)

### §A.3 SPEC artifacts

- `.moai/specs/SPEC-REMOVAL-GUARD-EXTRAS-001/spec.md` (REQ 10건)
- `.moai/specs/SPEC-REMOVAL-GUARD-EXTRAS-001/plan.md` (본 문서)
- `.moai/specs/SPEC-REMOVAL-GUARD-EXTRAS-001/acceptance.md` (AC 11건, two-cell)
- `.moai/specs/SPEC-REMOVAL-GUARD-EXTRAS-001/research.md` (검증 기록)
- `.moai/specs/SPEC-REMOVAL-GUARD-EXTRAS-001/progress.md` (§E.1 스켈레톤)

### §A.4 Plan-phase artifact set

Tier M = spec.md + plan.md + acceptance.md 3종 (+ research.md, progress.md 카드 관행 동반).

### §A.5 PRESERVE list (do NOT touch — scope discipline)

| 대상 | 이유 |
|------|------|
| `internal/hook/dangerous_removal.go` 전체 | t286 구조 체크 본체 — 로직 변경 0건 |
| `internal/hook/branch_guard.go` 전체 (`substituteQuotedArguments`, `quotedArgumentPattern` 포함) | t512/#1659 소관 |
| `internal/hook/pre_tool.go` `checkBashCommand` 로직 + `MergeExtraPatterns` | 현재 판정 순서(구조→quote folding→정규식)는 정확하다; 변경 불요 |
| `internal/cli/deps.go:240-243` 배포 policy 조합 | 정확한 조합 — 테스트가 따라갈 기준 |
| 기존 `TestDangerousRemoval_*` 4개 테스트 + `decideBash` 헬퍼 | builtin-only 구성 자체가 기본 policy의 회귀 표면 |
| `security.yaml`의 다른 extras 라인 5개 (`curl`, `wget`, `chmod`, `mkfs`, `dd`) | 범위 규율 — 관측만 |
| `internal/hook/security/config.go` 로더 | 정상 동작 |
| `.moai/reports/t511/` 이외의 다른 카드 증거 경로 | 카드 간 격리 |

### §A.6 EXTEND list (will touch)

| 대상 | 변경 |
|------|------|
| `internal/template/templates/.moai/config/sections/security.yaml` | 12행 1라인 삭제 (C1) — 주석 추가 금지 |
| `.moai/config/sections/security.yaml` | 7행 1라인 삭제 (C2) |
| `internal/settings/testdata/sections/security.yaml` | 7행 1라인 삭제 (C3) |
| `internal/hook/dangerous_removal_test.go` | 배포 policy 헬퍼 + 신규 테스트 함수군 추가 (기존 함수 보존) |
| `bin/moai` (빌드 산출물, 커밋 대상 아님) | `make build` 재생성 |

## §B Known Issues Auto-Injection (filtered to relevant categories)

- **§B.1 Cross-platform build tags**: 해당 없음 — 이 SPEC은 새로운 syscall/build-tag 코드를 만들지 않는다.
- **§B.2 Cross-SPEC policy conflict pre-scan**: `grep -rn "Retired\|TestHarnessRetirement\|superseded" internal/hook internal/settings` 선행. t286(034dad48e)과의 충돌 없음 확인 — 본 SPEC은 그 수리 위에 서는 보완이다. SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001 계열이 section 파일 스냅샷을 다루는지 §C 사전 비행에서 재확인 (본 plan 페이즈에서 internal/*.go grep 결과 BaseSnapshot 구현 부재 확인 — 재확인 요건 아님, 기록 목적).
- **§B.3 C-HRA-008 / Subagent boundary**: 새 테스트 코드에 AskUserQuestion 호출 없음. E-verification grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook | grep -v _test.go` → 0.
- **§B.4 Frontmatter canonical schema**: `created:`/`updated:`/`tags:` 정식 명칭 사용 — spec.md 작성 완료.
- **§B.5 CI 3-tier awareness**: spec-lint / golangci-lint / go test 각각 독립 판정. 기존 baseline 대비 NEW 결함만 분류.
- **§B.6 spec-lint heading convention**: spec.md의 Out of Scope는 `## F. Out of Scope` h2 + `### Out of Scope — …` h3 하위섹션으로 작성 완료 (MissingExclusions 회피).
- **§B.7 observer.go path resolution**: 해당 없음 — 훅 핸들러 로직 변경 없음.
- **§B.8 Working tree hygiene**: `moai update` 런타임 관리 파일 편집 금지; 커밋은 명시적 pathspec만.
- **§B.9 Git commit + push**: 이 리포는 git-flow 카드 워크트리 — 커밋은 본 워크트리 `WT-danger-guard-regex`에, push는 하지 않는다(리드 일괄). 커밋마다 카드 id t511 포함.
- **§B.10 Untouched paths PRESERVE**: §A.5 목록 이외 무변경.
- **§B.11 AskUserQuestion prohibited**: 서브에이전트 경계 준수 — blocker는 구조화 보고로.
- **§B.12 Sync-phase CHANGELOG**: manager-docs 소관 — `#1658`/`#1686` 이슈 참조와 "배포 템플릿 오탐 제거" 프레이밍 권장.

## §C Pre-flight Check List (런 페이즈 M1 개시 전 실행)

```bash
# 1. staleness rule — HEAD·브랜치 재판독
git rev-parse --short HEAD && git branch --show-current

# 2. 영향 패키지 baseline (전체 스위트 금지 — lane-local)
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ ./internal/settings/ -count=1

# 3. lint baseline
golangci-lint run ./internal/hook/... --timeout=2m 2>&1 | tail -3

# 4. RED-now 재확인 (선택 — plan 페이즈에서 3건 라이브 확보됨; 트리가 바뀌지 않았으면 재실행 불요)

# 5. template neutrality CI 가드 존재 확인
ls .github/workflows/template-neutrality-check.yaml
```

## §D Constraints (DO NOT VIOLATE)

- spec.md §D의 HARD 목록 전체 (branch_guard 무변경, 타 패턴 무변경, 기존 테스트 보존, 영어 주석/커밋).
- 템플릿 편집 순서는 Template-First: C1(템플릿) → `make build` → C2·C3. 역순 커밋 금지.
- 템플릿 파일에 이슈 번호·SPEC ID·날짜 주석 추가 금지 (REQ-RGE-009).
- 뮤턴트 프로브는 별도 커밋 없이 실행 후 되돌림 — 뮤턴트 상태로 커밋 금지, `git status --porcelain` 클린 확인 후 진행.
- 전체 스위트(`go test ./...`) 로컬 실행 금지 — lane-local 범위(`./internal/hook/ ./internal/settings/ ./internal/template/`)만.

## §E Self-Verification (skeleton — populated at run-phase)

E1 AC PASS/FAIL 매트릭스 / E2 크로스 플랫폼 빌드 / E3 커버리지(영향 패키지) / E4 경계 grep / E5 lint / E6 HEAD+push 상태 / E7 blocker / E8 RED 출력 — 각 항목 VCI 5-섹션 근거 형식. RED-now 셀은 research.md §R을 인용하고, 런 페이즈가 실행한 뮤턴트 RED를 추가한다. 모든 GREEN 판정은 스윕 테스트 수를 명시한다 (AC-RGE-011).

## §F Milestones

### M1 — 잔여 라인 제거 + 임베디드 재생성

1. C1: `internal/template/templates/.moai/config/sections/security.yaml` 12행(`- 'rm\s+-rf\s+/[^.]'` + 인라인 주석) 삭제.
2. `make build` (agents-emit-check → templ-generate → gen-catalog-hashes --all → go build). exit 0 확인.
3. C2: `.moai/config/sections/security.yaml` 7행 삭제. C3: `internal/settings/testdata/sections/security.yaml` 7행 삭제.
4. 검증: `go test ./internal/settings/ -count=1` (C3 fixture 소비 테스트 무손상 — schema_sections_test.go의 KEPT-key 경로 안정성 단언이 REMOVED-key 1개 삭제에 영향받지 않음을 관측), `go test ./internal/hook/ -run 'TestDangerousRemoval' -count=1` 여전히 `ok`.
5. 커밋: `feat(SPEC-REMOVAL-GUARD-EXTRAS-001): M1 remove superseded rm removal regex from security extras (t511)` — pathspec 명시 스테이징.

### M2 — 배포 policy 회귀 표면 (양방향) + 뮤턴트 프로브

1. `dangerous_removal_test.go`에 배포 policy 헬퍼 추가:
   - `decideBashDeployed(t *testing.T, command string) string` — `DefaultSecurityPolicy()`에 `MergeExtraPatterns(LoadExtraSecurityConfig(<템플릿 security.yaml 원본>))` 을 병합한 핸들러로 판정. 원칙 경로(A): `internal/template` 임베디드 FS에서 템플릿 원본을 읽어 `t.TempDir()`에 `.moai/config/sections/security.yaml`로 시딩 후 `LoadExtraSecurityConfig` 경유 (import 사이클 없음 — plan 페이즈에서 확인). 폴백(B): disk 경로 `../template/templates/.moai/config/sections/security.yaml`. 최후(C): 동등 extras 슬라이스 inline 구성 — 이 경우 REQ-RGE-006의 "배포 policy" 정의 충족 여부를 E-verification에 명시.
2. 허용 방향 신규 테스트 (전부 배포 policy). **fixture 규율 (F2, plan-audit 반영)**: 모든 테스트 fixture는 heredoc 본문의 **실공간 예시 형태 행만** 재현한다 — R1 관측 명령의 패턴 리터럴 행(`pattern literal: (?i)rm\s+-rf\s+/[^.]`)은 모든 fixture에서 **의도적으로 부재**다. AC-RGE-008의 스코프 grep이 M2 완료 시점에 0매치여야 하므로, 트리에 패턴 리터럴이 착지하는 fixture는 그 green path를 자기 모순으로 만든다. 패턴 텍스트 자체가 필요한 테스트는 Go에서 프로그램적으로 구성한다(문자열 연결 또는 문자 클래스 — 예: `"rm " + flags + " /tmp/x"`), 리터럴이 소스에 착지하지 않는다:
   - `TestDangerousRemovalDeployed_AllowsDeepPathHeredocDataMention` — R1의 예시 형태 행(실공간 `rm -rf /tmp/...` 행)만 담은 heredoc 명령 1건 이상, 패턴 리터럴 행 없음 (AC-RGE-001)
   - `TestDangerousRemovalDeployed_AllowsUnquotedDataMention` — R2 형태 (AC-RGE-002)
   - `TestDangerousRemovalDeployed_AllowsScratchCleanupExecution` — R3 형태 + 기존 OrdinaryCleanup 사례 재사용 (AC-RGE-003)
   - `TestDangerousRemovalDeployed_AllowsQuotedDataMention` — 기존 QuotedData 사례의 배포 policy 재실행 (AC-RGE-004)
3. 차단 방향 신규 테스트 (전부 배포 policy):
   - `TestDangerousRemovalDeployed_DeniesProtectedTargetsAllOrders` — `rm -fr /`, `rm -r -f /`, `--recursive --force`, quoted target, `~`, `$HOME`, `.git`, `node_modules`, `/usr`, `echo x && rm -fr /` (AC-RGE-005)
   - `TestDangerousRemovalDeployed_DeniesHeredocBodyProtectedTarget` — heredoc 본문 `rm -rf /` 형태 (AC-RGE-006, M3 특성화와 표면 공유)
4. 뮤턴트 프로브 (런 페이즈 실행, 커밋 없음):
   - **뮤턴트 A**: C1 라인을 템플릿에 재추가 → `make build` → 허용 방향 3테스트 RED 확인 → 원복. RED 출력 verbatim 보존 (AC-RGE-007의 RED 셀).
   - **뮤턴트 B**: `checkBashCommand`의 `dangerousRemovalTarget` 호출 블록 임시 주석 → 차단 방향 2테스트 RED 확인 → 원복 (AC-RGE-005/006의 RED 셀).
5. RED-first 규율: M2 테스트는 M1 라인 제거 **후**에는 GREEN이어야 한다 — RED는 (a) M1 이전 트리에서의 라이브 재현(R1-R3)과 (b) 뮤턴트 A/B 실행으로 충족한다. 테스트 함수 추가 커밋 시 `go test` 스윕 수 명시.
6. 커밋: `test(SPEC-REMOVAL-GUARD-EXTRAS-001): M2 deployed-policy bidirectional regression surface (t511)`.

### M3 — 문서화된 한계 기록 + 종결 준비

1. spec.md §F M3 절에 기록된 소비자 인지 heredoc 한계·셸 변수 간접이 코드와 불일치 없음을 확인 (기록 그대로 — 본 SPEC에서 코드 변경 없음).
2. **sync-phase 문서 부채 (F4, plan-audit 반영)**: docs-site 4개 로케일 `docs-site/content/{ko,en,ja,zh}/advanced/config-sections.md` **131행**이 제거 대상 정규식 `- 'rm\s+-rf\s+/[^.]'`을 배포 설정 예시로 문서화하고 있다 (본 개정 시점 4로케일 전부 실측 확인). **sync는 4개 로케일을 같은 변경으로 갱신해야 한다** (4-locale same-PR i18n 규칙, `docs-site-i18n-rules.md`) — 예시 블록에서 해당 행을 삭제하거나 구조 체크 시대의 예시로 교체. 4로케일 미갱신 상태의 병합은 낡은 문서를 배포한다: sync 완료 판정 전 `grep -rn "rm\\\\s+-rf" docs-site/content/` → 0매치 확인.
3. `acceptance.md` 증거 원장 완성: 각 AC에 RED-now 셀(트리 SHA 0b1e27877 고정, verbatim 출력, exit 경로) + green-path 셀(어느 마일스톤이 뒤집는지 + 예상 통과 출력) + 뮤턴트 RED 셀.
4. `go vet ./internal/hook/ ./internal/settings/ ./internal/template/` + 영향 패키지 `go test` 최종 스윕 (수 명시).
5. progress.md §E.2/§E.3 채움 (manager-develop). sync는 manager-docs.

## §G Anti-Patterns

- **builtin-only policy로 배포 동작 주장** — 이 SPEC의 근본 결함을 재생산한다 (REQ-RGE-006 위반).
- **셀렉터 0매치 GREEN** — `-run` 패턴이 0 테스트를 스윕하면 ok가 아니라 gap이다 (AC-RGE-011).
- **뮤턴트 없는 차단 AC 채택** — 현재 GREEN인 차단 AC는 뮤턴트 B 없이는 공허할 수 있다 (verification-completeness §2).
- **템플릿에 "수리" 주석 추가** — 중립성 위반 (REQ-RGE-009); 이력은 git과 SPEC 문서가 carry한다.
- **quote folding 의존 오탐 해석** — unquoted 데이터 언급은 folding을 통과하지 못하므로 extras 제거가 유일한 수리다.

## §H Cross-References

- spec.md §A (근거), §C (REQ), §F (Out of Scope + M3 한계)
- research.md §R (RED-now 증거 원장)
- `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell), `.claude/rules/moai/core/verification-claim-integrity.md` §2 (baseline 귀속)
- CLAUDE.local.md §2 (Template-First), `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (중립성 카탈로그)
