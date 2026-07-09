# SPEC-TEMPLATE-RULES-CLEANUP-001 — 수용 기준 (acceptance.md)

모든 AC는 기계 검증 명령을 갖는다. `TPL=internal/template/templates/.claude/rules` 약칭 사용 (실행 시 전개).

## §D AC 매트릭스

### 공통 (Group P)

| AC | 대응 REQ | 검증 명령 | 기대 결과 |
|----|----------|-----------|-----------|
| AC-TRC-001 | REQ-TRC-003 | `go test ./internal/template/ -count=1` (pre-flight baseline + M7) | 기존 스위트 PASS (baseline green 유지) |
| AC-TRC-002 | REQ-TRC-001 | 편집된 각 template rule 파일 f에 대해 `grep -c '\[HARD\]' $f` 를 pre-flight baseline과 대비 | 전 파일 동수. 예외 2건만 허용: `design/constitution.md`(Sprint 블록 retired 주석에 따른 변동 — 델타 사유를 progress.md에 기록), `zone-registry.md`(파일 제거) |
| AC-TRC-003 | REQ-TRC-002 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift\|TestSanitizedPairParity' -count=1` | PASS (분류 매트릭스 준수 증명) |

### Group A — 깨진 참조

| AC | 대응 REQ | 검증 명령 | 기대 결과 |
|----|----------|-----------|-----------|
| AC-TRC-A1 | REQ-TRC-010 | `grep -rn 'dev-only-commands-isolation\.md\|git-local-workflow-doctrine\.md\|git-workflow-doctrine\.md\|pattern-library\.md' $TPL/` | 0 hits (exit 1) |
| AC-TRC-A2 | REQ-TRC-012 | `grep -nE '9[78]-\*|9[78]-[a-z]' $TPL/moai/development/agent-authoring.md` | 0 hits |
| AC-TRC-A3 | REQ-TRC-012 | `grep -c 'agents/local' $TPL/moai/development/agent-authoring.md` | ≥ 1 (`.claude/agents/local/` 보호 계약 프로즈 보존) |
| AC-TRC-A4 | REQ-TRC-011 | 편집 4개 파일의 해당 섹션에 인라인 프로즈 또는 shipped-equivalent 포인터 존재 (수동 diff 리뷰 + AC-TRC-002의 [HARD] 동수) | 조항 삭제 없음 |

### Group B — 중립성

| AC | 대응 REQ | 검증 명령 | 기대 결과 |
|----|----------|-----------|-----------|
| AC-TRC-B1 | REQ-TRC-020 | `grep -rEn '\b(REQ\|AC)-[A-Z][A-Z0-9]*-[0-9]+\b' $TPL/ \| grep -vE '(REQ\|AC)-XXX-'` | 0 hits (placeholder 제외 후) |
| AC-TRC-B2 | REQ-TRC-021 | `grep -n 'Epic 7\|TMC-001\|L51\|§24 namespace' $TPL/moai/core/askuser-protocol.md` | 0 hits |
| AC-TRC-B2b | REQ-TRC-021 | `grep -c 'Worked Example' $TPL/moai/core/askuser-protocol.md` | ≥ 1 (교육 예시 구조 보존) |
| AC-TRC-B3 | REQ-TRC-022 | `grep -rEn '\blessons? #[0-9]+' $TPL/` | 0 hits |
| AC-TRC-B3b | REQ-TRC-022 | `grep -rn 'W3 meta-analysis\|W0 fix\|W1/W2\|W3 케이스' $TPL/` | 0 hits (감사 실명 인스턴스 소거) |
| AC-TRC-B4 | REQ-TRC-023 | `grep -rn '2026-05-17\|2026-05-20\|2026-05-04\|2026-05-09\|2026-04-26\|2026-04-20' $TPL/` | 0 hits |
| AC-TRC-B5 | REQ-TRC-024 | `test ! -f $TPL/moai/core/zone-registry.md && test -f .claude/rules/moai/core/zone-registry.md && echo OK` | `OK` (template 제거 + local 보존) |
| AC-TRC-B6 | REQ-TRC-025 | `grep -rn 'zone-registry' internal/template/templates/` | 0 hits |
| AC-TRC-B7 | REQ-TRC-026 | `grep -rn 'CONST-V3R' $TPL/ ; grep -n 'MIG-003' $TPL/moai/core/settings-management.md` | 양쪽 0 hits |
| AC-TRC-B8 | REQ-TRC-027 | scratch: `d=$(mktemp -d) && (cd $d && moai init t && cd t && moai constitution list; echo "cl-exit=$?"; moai doctor >/dev/null 2>&1; echo "dr-exit=$?")` | 두 명령 모두 crash 없음 — exit 0 또는 문서화된 informative degradation (verbatim 출력을 §E.2에 기록) |

### Group C — 백포트

| AC | 대응 REQ | 검증 명령 | 기대 결과 |
|----|----------|-----------|-----------|
| AC-TRC-C1 | REQ-TRC-030 | `grep -c '§I Token Accounting' $TPL/moai/development/spec-frontmatter-schema.md` | ≥ 1 |
| AC-TRC-C2 | REQ-TRC-031 | `grep -nE 'MoAI (agent\|artifact\|policy projection\|projection\|is a harness\|CONSUMES)\|of MoAI.s reach\|MoAI-level policy' $TPL/moai/workflow/runtime-recovery-doctrine.md` | 0 hits (~8곳 moai-adk 전환 완료) |
| AC-TRC-C3 | REQ-TRC-032 | `grep -nE '\.moai/research/\|^Version:\|^Origin:\|CONST-V3R' $TPL/moai/workflow/runtime-recovery-doctrine.md` | 0 hits (sanitized 요소 미복사) |
| AC-TRC-C4 | REQ-TRC-032 | `go test ./internal/template/ -run TestSanitizedPairParity -count=1` | PASS |

### Group D — retired 어휘

| AC | 대응 REQ | 검증 명령 | 기대 결과 |
|----|----------|-----------|-----------|
| AC-TRC-D1 | REQ-TRC-040 | `grep -rn '\.moai/sprints' $TPL/` | 0 hits (config `design.yaml`은 rules 트리 밖 — 무접촉) |
| AC-TRC-D1b | REQ-TRC-040 | `awk '/Sprint Contract Protocol/,/^## /' $TPL/moai/design/constitution.md \| head -5 \| grep -ci 'retired\|historical'` | ≥ 1 (블록 선두에 retired-historical 주석) |
| AC-TRC-D2 | REQ-TRC-041 | `grep -n 'cohort' $TPL/moai/workflow/orchestration-mode-selection.md` | 0 hits |

### Group E — design 역드리프트

| AC | 대응 REQ | 검증 명령 | 기대 결과 |
|----|----------|-----------|-----------|
| AC-TRC-E1 | REQ-TRC-050 | `diff .claude/rules/moai/design/constitution.md internal/template/templates/.claude/rules/moai/design/constitution.md; echo exit=$?` | `exit=0` (byte-identical) |
| AC-TRC-E2 | REQ-TRC-051 | `grep -n 'design_docs' .moai/config/sections/design.yaml` | 0 hits |
| AC-TRC-E2b | REQ-TRC-051 | local/template design.yaml의 `design:` 하위 top-level key 집합 비교 (예: `grep -E '^    [a-z_]+:' local` vs `grep -E '^  [a-z_]+:' template` 정렬 비교 — 들여쓰기 차 감안, run-phase에서 정확 명령 확정) + `go test ./internal/config/ -run Symmetry -count=1` | key-set 일치 + Symmetry PASS |

### Group F — CI 가드

| AC | 대응 REQ | 검증 명령 | 기대 결과 |
|----|----------|-----------|-----------|
| AC-TRC-F1 (RED) | REQ-TRC-065 | M1 시점 (정리 전): 신규 가드 4종 각각 `go test ./internal/template/ -run <가드명> -count=1` | 4종 모두 FAIL — 각 가드가 기지 위반 ≥1건을 file:line:token으로 실명 검출; verbatim 출력 progress.md §E.2 기록 |
| AC-TRC-F2 (GREEN) | REQ-TRC-065 | M7 시점 (정리 후): 동일 명령 | 4종 모두 PASS |
| AC-TRC-F3 | REQ-TRC-060..063 | 가드 스코프 확인: (a) REQ/AC rules-스코프, (b) lessons/W#, (c) date 별도 테스트 함수, (d) CONST/SPEC-V3R rules 확장 — 코드 리뷰 + `go test ./internal/template/ -run TestLeakClassNoDateShaInDefaultTier -count=1` | date 패턴이 default-tier `leakClasses` 밖에 있음 (해당 테스트 PASS로 기계 증명) |
| AC-TRC-F4 | REQ-TRC-064 | 신규 가드 FAIL 출력에 file path + line + matched token + sentinel 문자열 포함 (RED 출력에서 확인) | 4종 전부 충족 |
| AC-TRC-F5 | REQ-TRC-066 | recurrence backstop 셀프테스트: synthetic re-leak probe → 가드 fire, clean replacement → pass | PASS (가드별 backstop 함수 존재) |
| AC-TRC-F6 | REQ-TRC-003 | `make build && go test ./internal/template/ ./internal/config/ -count=1` | exit 0, 전부 PASS |

## Given-When-Then 시나리오

### 시나리오 1 — 신규 사용자 배포 (Finding A/B 소비자 관점)

- **Given** M7 완료본 바이너리 (`make build` 이후)
- **When** 사용자가 빈 디렉터리에서 `moai init myproject` 실행
- **Then** 배포된 `.claude/rules/` 트리에 (i) 4개 unshipped 경로 인용 0건, (ii) `zone-registry.md` 부재, (iii) `moai constitution list`/`moai doctor`가 crash 없이 동작, (iv) REQ/AC/CONST/lessons/날짜 provenance grep 전부 0 hit

### 시나리오 2 — 재발 방지 (Group F 가드 관점)

- **Given** M7 이후 green 상태의 main
- **When** 기여자가 template rule 파일에 `lessons #42 참조` 또는 `REQ-NEW-001` 토큰을 재도입하고 `go test ./internal/template/`를 실행
- **Then** 해당 가드가 file:line:token + sentinel로 FAIL하고, pedagogical allowlist 등재 없이는 통과 불가

### 시나리오 3 — sanitized-pair 방향성 (Group C 경계)

- **Given** `runtime-recovery-doctrine.md` 백포트 완료 상태
- **When** `TestSanitizedPairParity`와 leak test 실행
- **Then** template 측에 `.moai/research/` 참조·Origin footer·CONST 토큰이 없고(AC-TRC-C3), local 측 내부 콘텐츠는 무접촉이며, 두 테스트 모두 PASS

## Edge Cases

1. **W3C 오탐**: `\bW[0-9]\b`류 광패턴은 `SPEC-W3-*` 등 오탐 여지 — 협패턴 + allowlist로 흡수 (design.md §3). 가드 도입 시 template 트리 전체에서 오탐 0건임을 RED/GREEN 대비로 확인.
2. **Pedagogical placeholder**: `manager-develop-prompt-template.md:175`의 `AC-XXX-001` 예시표는 regex-제외(`-XXX-`) 또는 allowlist 등재 — GREEN 상태에서 보존 확인 (`grep -c 'AC-XXX-001' $TPL/moai/development/manager-develop-prompt-template.md` ≥ 1).
3. **Footer 날짜**: `Last Updated:`/`Version:` 라인의 날짜는 date 가드의 라인-컨텍스트 제외 대상 — 단, Finding B 확인분(`design/constitution.md:423-424`)은 내부 작업 날짜로 판정되어 제거 대상. 제거 후 가드 제외 규칙과 충돌 없음을 GREEN으로 확인.
4. **`.moai/sprints` config 키**: `design.yaml gan_loop.sprint_contract.artifact_dir`는 라이브 — rules 프로즈만 정리하고 config grep은 AC 범위 밖 (AC-TRC-D1은 `$TPL/` 한정).
5. **zone-registry 부재 CLI**: `moai constitution list`가 빈 registry에서 error exit하는 경우 — REQ-TRC-027 조건부 Go 수정 발동, AC-TRC-B8 재실행.

## Quality Gates

- `go vet ./...` + `golangci-lint run` clean (M1/M3 Go 변경분)
- `go test ./internal/template/ ./internal/config/ -count=1` 전체 PASS
- `moai spec lint .moai/specs/SPEC-TEMPLATE-RULES-CLEANUP-001` clean
- coverage: 신규/수정 가드 테스트는 테스트 코드이므로 별도 커버리지 목표 없음; M3 조건부 Go 수정 발생 시 해당 패키지 기존 커버리지 비하락

## Definition of Done

1. AC-TRC-001..F6 전부 PASS (E1 매트릭스에 verbatim 출력)
2. 신규 가드 RED→GREEN 증거가 progress.md §E.2에 기록되고 커밋 히스토리에서 M1 가드 커밋이 정리 커밋에 선행
3. `make build` 성공 + scratch 배포 시나리오 1 통과
4. GREEN 상태에서 단일 push 완료, main CI green
5. progress.md §E.3 run-phase audit-ready 신호 기록
