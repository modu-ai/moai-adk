# SPEC-REMOVAL-GUARD-EXTRAS-001 — Research (검증 기록 원장)

저작: manager-spec, plan 페이즈, 2026-09-07. 모든 측정은 **이 러인 세션이 이 트리에서 직접 실행**했다.
Baseline 귀속: tree `0b1e27877` (`git rev-parse --short HEAD` 실측), 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t511`, 브랜치 `WT-danger-guard-regex`, 작성 시 `git status --porcelain` 클린.

## §R.1 Baseline — 기존 테스트는 통과한다 (builtin-only)

**Command**:
```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run 'TestDangerousRemoval' -count=1
```
**Verbatim output**:
```
ok  	github.com/modu-ai/moai-adk/internal/hook	0.669s
```
**해석**: t286의 4개 테스트(플래그 순서·quoted 데이터·일상 정리·quote-folding 후 차단)는 GREEN. 그러나 `dangerous_removal_test.go:12`의 `decideBash` 헬퍼가 `preToolHandler{policy: DefaultSecurityPolicy()}`로 **builtin-only** 구성 — extras 경로는 이 표면 어디에도 없다. 이것이 잔여 결함이 t286을 살아남게 한 메커니즘이다.

## §R.2 라이브 RED — 배포 policy 경로 3클래스 (plan 페이즈 본 세션 관측)

이 세션의 PreToolUse 가드는 `internal/cli/deps.go:240-243`이 조합한 배포 policy(기본 + extras)를 실행한다. 세 호출 모두 **실행 전 거부** — 명령 본문은 수행되지 않았고, 거부 자체가 증거다.

### R1 — heredoc 데이터 언급 (디스패치 지시 재현)

**Command (도구 호출 원문)**:
```
cat > /tmp/t511-spec-repro.md <<'EOF'
pattern literal: (?i)rm\s+-rf\s+/[^.]
example form: rm -rf /tmp/moai-spec-repro-123
EOF
echo WROTE_OK
```
**Verbatim hook output (도구 오류로 반환)**:
```
Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]
```
**Exit 경로**: PreToolUse deny — 명령 미실행. `echo WROTE_OK`는 도달하지 않음.
**발화 주체 특정**: 거부 메시지가 패턴 리터럴 자체 → 정규식 스캔 경로(`pre_tool.go:961` `fmt.Sprintf("Dangerous command blocked: %s", pattern.String())`). 구조 체크 경로라면 `removal of protected path %q`(`pre_tool.go:948`) — 세 사례 모두 대상이 비보호 깊은 경로라 구조 체크는 통과. 운영자 라이브 재현(2026-09-07)과 바이트 동일 — **5번째 독립 관측**.

### R2 — unquoted 데이터 언급 (plan 페이즈 신규 발견 방향)

**Command**: `echo example: rm -rf /tmp/moai-spec-repro-456`
**Verbatim hook output**: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]`
**Exit 경로**: PreToolUse deny — 미실행.
**의미**: 인용부호가 없으면 `substituteQuotedArguments` 접을 것이 없어 extras 정규식이 텍스트 그대로 매치한다. 구조 체크는 head 토큰이 `echo`라 비-removal로 올바르게 분류 — extras 제거가 유일한 수리.

### R3 — 실행 클래스 스크래치 정리 (#1658 원보고 형태)

**Command**: `rm -rf /tmp/t511-scratch-nonexistent-dir`
**Verbatim hook output**: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]`
**Exit 경로**: PreToolUse deny — 미실행.
**의미**: 데이터 언급이 아닌 **실제 제거 명령**이 extras 정규식 하나로 거부된다. #1658의 "routine scratch cleanup was refused" 그 자체.

## §R.3 패턴 수준 기계 확인 (Python `re`, 구성된 subject — 명령 텍스트에 발화 형태 미포함 목적)

| subject | `(?i)rm\s+-rf\s+/[^.]` 매치 |
|---------|------------------------------|
| `rm -rf /tmp/moai-spec-repro-123` | **True — 오탐** |
| `rm -rf /` (bare root, end-of-string) | **False — 자기 주석 커버리지 뒤집힘** |
| `rm -rf /usr` | True |
| `rm -rf /.` | False |
| `echo example: rm -rf /tmp/x` | **True — unquoted 데이터 오탐** |
| `rm -rf $HOME` | False — 간접 미커버 (문서화 한계와 일치) |

이 측정은 Go `regexp`와 동일 RE2 계열 의미론의 독립 확인일 뿐이며, 판정의 주 증거는 §R.2의 라이브 가드 관측이다.

## §R.4 소스 판독 사실 (직접 Read, tree 0b1e27877)

| 위치 | 확인 내용 |
|------|-----------|
| `internal/hook/pre_tool.go:934-973` | `checkBashCommand` — (a) 947-949 구조 체크 선행, (b) 956 `substituteQuotedArguments`, (c) 959-963 전체 `DangerousBashPatterns` 정규식 스캔 (기본+extras 병합 목록). 거부 메시지 형식 948·961 확인 |
| `internal/hook/pre_tool.go:298-314` | `MergeExtraPatterns` — extras 302-304에서 append-only (never replace) |
| `internal/cli/deps.go:240-243` | 배포 policy 조합: `DefaultSecurityPolicy()` + `MergeExtraPatterns(LoadExtraSecurityConfig(cwd))` — 세션 라이브 가드가 실행하는 것과 동일 조합 |
| `internal/hook/security/config.go:24-39` | `LoadExtraSecurityConfig` — `<project>/.moai/config/sections/security.yaml` 판독, 부재/파싱 실패 시 nil (graceful) |
| `internal/hook/branch_guard.go:197-199` | `substituteQuotedArguments` — quoted span → placeholder. heredoc 본문은 quoted span이 아니므로 접힘 생존 (R1이 그 증명) |
| `internal/hook/dangerous_removal.go` | 구조 체크 전체; `shellSeparators`(27행)에 `"\n"` 포함 → heredoc 본문 행이 명령 세그먼트로 심사됨 |
| 템플릿 `security.yaml:2,12` | 헤더 "Extra patterns extend (never replace)" + 12행 결함 라인 + 주석 "rm -rf targeting root paths" |
| dogfood `security.yaml:7` | 결함 라인 (포맷 다름, 동일 정규식) |
| `internal/settings/testdata/sections/security.yaml:7` | **디스패치에 없던 세 번째 사본** — `sectionwrite_test.go` `seedSectionFixture`와 `schema_sections_test.go` 소비. policy 소비자 아님(스키마/쓰기 테스트 fixture)이나 동일 결함 라인 |
| `internal/hook/dangerous_removal_test.go` | 4개 테스트 전부 builtin-only (12행) — extras 경로 무테스트 |
| import 방향 | `internal/template` → `internal/hook` import 0건 — 훅 테스트가 템플릿 임베디드 FS를 읽어도 사이클 없음 |

## §R.5 발견 — 디스패치 사실 대조

| 항목 | 디스패치 | 실측 | 처치 |
|------|----------|------|------|
| 사본 수 | 2곳 (템플릿:12, dogfood:7) | **3곳** (+testdata:7) | C3를 M1 범위에 포함 (동일 라인·동일 결함; settings 테스트 무손상 검증을 M1에 결합) |
| quoted 데이터 언급 | "already covered builtin-only, now also under deployed policy" | quoted는 참(접힘). **unquoted 데이터 언급은 오늘도 배포 policy에서 거부됨**(R2) — 완화 방향의 추가 RED | AC-RGE-002로 편입 |
| `[^.]` 커버리지 반전 | fact 6 | §R.3 기계 확인으로 전량 참 | spec.md §A.3 인용 |
| progress.md §F.1 표기 | "progress.md §F.1 skeleton" | 정본 Section Map에서 plan-phase 신호는 **§E.1** (§F는 Phase 4 Mode Selection, orchestrator 소관); §F.1은 plan.md의 File-Touch Inventory 관행 | progress.md는 §E.1 스켈레톤 + 카드 헤더(t503 선례), plan.md에 §A.6 EXTEND 목록으로 대응 |
| `substituteQuotedArguments` 위치 | branch_guard.go:197 | 197-199 확인 | 일치 |
| pre_tool.go extras 병합 | 302-304 | 302-304 (함수는 298 개시) | 일치 |

## §R.6 뮤턴트 명세 (런 페이즈 실행 대상)

| 뮤턴트 | 조작 | 예상 RED | 판별 대상 |
|--------|------|----------|-----------|
| A | C1 라인 템플릿 재추가 → `make build` | `TestDangerousRemovalDeployed_Allows*` 3종 | 테스트가 extras 병합을 실제로 로드하는가 (공허한 통과 차단, REQ-RGE-007) |
| B | `checkBashCommand`의 `dangerousRemovalTarget` 호출 블록 임시 제거 | `TestDangerousRemovalDeployed_Denies*` 2종 | t286 구조 체크의 회귀 방어가 살아 있는가 (REQ-RGE-007) |

두 뮤턴트 모두 실행 후 즉시 원복, 커밋 금지, `git status --porcelain` 클린 확인.

## §R.7 사후 감사 개정 측정 (post-audit, plan-audit PASS 0.87 이후)

트리 재판독: `git rev-parse --short HEAD` → `0b1e27877`, 브랜치 `WT-danger-guard-regex` 불변 (staleness rule).

1. **F2c 스코프 grep 실측** — 명령: `grep -rn 'rm\\s+-rf' internal/ .moai/config/` · 출력: 정확히 3매치 (`internal/settings/testdata/sections/security.yaml:7`, `internal/template/templates/.moai/config/sections/security.yaml:12`, `.moai/config/sections/security.yaml:7`) · exit 0. 스코프가 소스 트리+로컬 config로 한정돼 SPEC 아티팩트·reports·docs-site는 계수 대상이 아니다. 전문은 acceptance.md [LEDGER-GREP3].
2. **F3 라이브 재현 (AC-009 RED 셀용)** — 명령: `echo example: rm -rf /tmp/moai-spec-repro-456` · verbatim: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]` · PreToolUse deny, 미실행. 사후 감사 시점에도 배포 경로가 결함 라인을 로드한다는 생 관측. 전문은 acceptance.md [LEDGER-R4].
3. **F4 docs-site 부채 실측** — `docs-site/content/{ko,en,ja,zh}/advanced/config-sections.md` **131행**(4로케일 전부 동일 행)이 `- 'rm\s+-rf\s+/[^.]'`을 배포 설정 예시로 문서화. sync-phase 갱신 대상 (plan.md M3 §2).
