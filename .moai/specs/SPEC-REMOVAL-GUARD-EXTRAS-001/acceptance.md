# SPEC-REMOVAL-GUARD-EXTRAS-001 — Acceptance Criteria

채택 규율: `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell 채택 + 뮤턴트 프로브 + empty-sweep). RED-now 셀의 귀속 원장: `research.md` §R (트리 SHA `0b1e27877` 고정 — 이동 브랜치명 아님). 분류가 release-blocking인 AC의 RED 셀은 네 요소(명령·verbatim 출력·exit 경로·트리 SHA)를 갖춘다. 라이브 RED가 현재 트리에서 재실행 불가능해지는 경우 그 AC는 regression-guard로 재분류되고 pass로 기록되지 않는다.

쌍 규율: 모든 방향 완화 AC(001-004)는 방향 강화 AC(005-007)와 쌍이다 — "한쪽만 넣으면 반대 방향이 되살아난다"(카드 HARD).

## 증거 원장 (evidence ledger — AC 표가 셀 id로 인용)

**[LEDGER-R1]** (RED-now, tree `0b1e27877`, plan 페이즈 본 세션)
- 명령: heredoc 쓰기 — `cat > /tmp/t511-spec-repro.md <<'EOF'` / 본문 1행 `pattern literal: (?i)rm\s+-rf\s+/[^.]` / 본문 2행 `example form: rm -rf /tmp/moai-spec-repro-123` / `EOF` / `echo WROTE_OK`
- verbatim 출력: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]`
- exit 경로: PreToolUse deny — 명령 미실행 (`echo` 미도달)
- 발화 주체: extras 정규식 경로 (메시지가 패턴 리터럴 자체; 구조 체크 메시지 형식과 상이) — research.md §R.2-R1

**[LEDGER-R2]** (RED-now, 동일 트리) 명령: `echo example: rm -rf /tmp/moai-spec-repro-456` · verbatim: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]` · exit 경로: PreToolUse deny — 미실행 · §R.2-R2

**[LEDGER-R3]** (RED-now, 동일 트리) 명령: `rm -rf /tmp/t511-scratch-nonexistent-dir` · verbatim: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]` · exit 경로: PreToolUse deny — 미실행 · §R.2-R3

**[LEDGER-BASE]** (baseline) 명령: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run 'TestDangerousRemoval' -count=1` · verbatim: `ok  	github.com/modu-ai/moai-adk/internal/hook	0.669s` · exit 0 · §R.1

**[LEDGER-MUT-A]** (런 페이즈 채움 — 뮤턴트 A: C1 라인 재추가 → `make build` → 허용 방향 3테스트 RED 출력 verbatim + 원복 확인)
**[LEDGER-MUT-B]** (런 페이즈 채움 — 뮤턴트 B: `dangerousRemovalTarget` 호출 블록 임시 제거 → 차단 방향 2테스트 RED 출력 verbatim + 원복 확인)

사후 감사 개정에서 추가된 셀 (post-audit, 동일 트리 `0b1e27877` 재판독 후 측정):

**[LEDGER-R4]** (RED-now — AC-RGE-009용, plan-audit 이후 신규 실행) 명령: `echo example: rm -rf /tmp/moai-spec-repro-456` · verbatim 출력: `Dangerous command blocked: (?i)rm\s+-rf\s+/[^.]` · exit 경로: PreToolUse deny — 명령 미실행 (거부 자체가 증거) · 트리 SHA: `0b1e27877` (직전 staleness 재판독). 현재 배포 경로가 여전히 결함 라인을 로드해 unquoted 데이터 언급을 거부한다는 사전 구현 상태의 생 관측 — M1+M2가 뒤집는 상태.

**[LEDGER-GREP3]** (RED-now — AC-RGE-008용, 사후 감사 개정 실측) 명령: `grep -rn 'rm\\s+-rf' internal/ .moai/config/` · verbatim 출력 (3매치):
```
internal/settings/testdata/sections/security.yaml:7:        - rm\s+-rf\s+/[^.]
internal/template/templates/.moai/config/sections/security.yaml:12:    - 'rm\s+-rf\s+/[^.]'            # rm -rf targeting root paths
.moai/config/sections/security.yaml:7:        - rm\s+-rf\s+/[^.]
```
· exit 코드 0 (매치 있음) · 트리 SHA: `0b1e27877`. 스코프가 소스 트리(`internal/`) + 로컬 config(`.moai/config/`)로 한정되므로 본 SPEC 아티팩트(`.moai/specs/`, `.moai/reports/`)와 docs-site는 애초 범위 밖 — 아티팩트 텍스트의 리터럴 인용이 매치로 계수될 여지가 없다.

## AC 매트릭스

| AC | 요구 | 방향 | RED 셀 | GREEN 경로 | 분류 |
|----|------|------|--------|------------|------|
| AC-RGE-001 | REQ-RGE-001, 003 | 완화 | [LEDGER-R1] (트리 0b1e27877) | M1 라인 제거 + M2 `TestDangerousRemovalDeployed_AllowsDeepPathHeredocDataMention` GREEN — 배포 policy에서 `""`(허용), 스윕 수 명시. fixture는 R1의 **실공간 예시 형태 행만** 재현(패턴 리터럴 행 의도적 부재, plan.md M2.2 fixture 규율) | release-blocking |
| AC-RGE-002 | REQ-RGE-001 | 완화 | [LEDGER-R2] | M1+M2 `TestDangerousRemovalDeployed_AllowsUnquotedDataMention` GREEN — fixture에 패턴 리터럴 없음(실공간 형태만) | release-blocking |
| AC-RGE-003 | REQ-RGE-002 | 완화 | [LEDGER-R3] | M1+M2 `TestDangerousRemovalDeployed_AllowsScratchCleanupExecution` GREEN | release-blocking |
| AC-RGE-004 | REQ-RGE-001, 006 | 완화(회귀 가드) | 없음 — quoted는 접힘으로 현재 GREEN (green-now 셀: [LEDGER-BASE] + M2 배포 policy 재실행) | M2 `TestDangerousRemovalDeployed_AllowsQuotedDataMention` GREEN — builtin-only GREEN이 배포 policy에서도 유지됨을 고정 | regression-guard |
| AC-RGE-005 | REQ-RGE-004 | 강화 | [LEDGER-MUT-B] — 뮤턴트 B에서 RED 복귀 (헤드에선 green-now) | M2 `TestDangerousRemovalDeployed_DeniesProtectedTargetsAllOrders` GREEN (구조 체크 경유, 메시지 `removal of protected path %q` 형식) | release-blocking |
| AC-RGE-006 | REQ-RGE-005 | 강화 | [LEDGER-MUT-B] — 동일 뮤턴트 | M2 `TestDangerousRemovalDeployed_DeniesHeredocBodyProtectedTarget` GREEN — M3 문서화 한계의 현재 동작 특성화 | release-blocking |
| AC-RGE-007 | REQ-RGE-006, 007 | 판별 | [LEDGER-MUT-A] — 뮤턴트 A에서 허용 3테스트 RED 복귀. 재추가에도 GREEN이면 테스트가 extras를 로드하지 않는 것 — REQ-RGE-006 위반으로 M2 재작업 | M2 배포 policy 헬퍼(템플릿 임베디드 FS 원칙 경로)가 AC-001/002/003의 유일한 판정기 | release-blocking |
| AC-RGE-008 | REQ-RGE-003 | 제거 | [LEDGER-GREP3] — 스코프 grep(`internal/` + `.moai/config/`, SPEC 아티팩트·docs-site 범위 밖)이 사전 수정 트리에서 **정확히 3매치**(templates:12, dogfood:7, testdata:7)를 실측 (트리 0b1e27877) | M1 후 동일 스코프 grep → **0매치** — 단, M2.2 fixture 규율(패턴 리터럴 트리 착지 금지)이 충족돼야 도달 가능한 green이다 | release-blocking |
| AC-RGE-009 | REQ-RGE-008 | 배포 | [LEDGER-R4] — 사후 감사 시점 라이브 재현: 배포 경로가 여전히 결함 라인을 로드해 unquoted 데이터 언급을 거부 (트리 0b1e27877, verbatim + exit 경로 기록) | M1 `make build` exit 0 + AC-RGE-007의 템플릿 원본 판독 테스트가 편집된 yaml을 읽는 것 자체가 임베디드 반영 증명 | release-blocking |
| AC-RGE-010 | REQ-RGE-009, 010 | 중립성+범위 | 없음 (문제 없음 — 회귀 방지). **F1 해결**: REQ-RGE-010(타 extras 라인 무변경)은 절차적 PRESERVE 항목으로 두지 않고 본 AC에 **실측 기준으로 흡수**했다 — 하나의 측정이 두 요구를 함께 판정한다 | M1 3사본 각각 `git diff --numstat` → **deleted=1 added=0** (템플릿 파일은 추가 행 0 = 중립성 위반 주석 없음, 그리고 전체 변경이 정확히 1행 삭제 = 나머지 5개 extras 라인·헤더·나머지 전부 무변경 = REQ-RGE-010 실측) + CI `template-neutrality-check.yaml` green. 판별 뮤턴트(런 페이즈, 선택): 타 extras 라인 1행 변경 시 numstat가 (0,1)에서 벗어나 RED — 측정이 실제로 두 요구를 가르는지 확인 | regression-guard (불변 보존형 — 현재 트리가 불변을 만족하므로 RED-now가 존재하지 않는 방향) |
| AC-RGE-011 | (판정 규율) | 규율 | 없음 | 모든 GREEN 판정이 스윕 테스트 수를 명시; `-run` 셀렉터 0매치는 gap으로 보고 — 런 페이즈 각 검증 배치와 sync 감사에 적용 | release-blocking (판정 규율) |

## 채택 판정 절차 (런 페이즈)

1. M2 완료 시점에 AC-001~007의 GREEN을 배치로 관측 (단일 턴 다중 Bash, env-scrub 복합 1회 호출): `unset … && go test ./internal/hook/ -run 'TestDangerousRemoval' -count=1 -v` — 스윕 수를 출력에서 집계해 AC-RGE-011 증거로 기록.
2. 뮤턴트 A/B를 §R.6 순서로 실행 — RED 출력 verbatim을 LEDGER-MUT-A/B에 기록 후 원복, `git status --porcelain` 클린 확인.
3. AC-008 grep 재실행 (0매치 확인 — 0이 "스캔 실패"가 아니라는 대조: grep 자체 exit 0이면 매치 0이 참; 패턴 오탈자 방지로 템플릿 파일명 경로를 grep 대상에 포함해 매치 3건이었음을 먼저 재현).
4. 미충족 AC는 pass로 기록 금지 — 회귀 가드로 재분류하거나 FAIL로 보고.

## 알려진 한계 기록 (코드 변경 없음 — spec.md §F 참조)

- 소비자 인지 heredoc folding 미유예: 데이터 소비 heredoc 본문의 보호 대상 언급은 구조 체크로 거부됨 (AC-RGE-006이 현재 동작을 고정). 천장·재검토 트리거는 spec.md §F 두 번째 절.
- 셸 변수 간접(`rm -rf $TARGET`)은 텍스트 수준 가드로 해결 불가 — 설계상 한계로 문서화, 수리 대상 아님.
