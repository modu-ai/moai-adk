# SPEC-AUDIT-GATE-INTEGRITY-001 — Implementation Plan

> Tier: **M** (standard) | doc-only | 대상 파일: live 4 + template mirror 4 (총 8 편집) + `make build`

## §A Context

3-agent 병렬 감사(2026-07-09)의 P0 4결함을 문서-층에서 수정한다. 공통 원리: **게이트는 실제로 게이트해야 하고, 검증 주장은 실측 출력을 증거로 가져야 한다** (`verification-claim-integrity.md` §1.1 surfaces 1-2). 전체 결함 증거 앵커는 spec.md §A.3 표 참조.

### A.1 편집 대상 파일 매트릭스

| 파일 (live) | 관련 REQ | Template mirror | mirror 경로 |
|-------------|----------|-----------------|-------------|
| `.claude/agents/moai/plan-auditor.md` | AGI-001..003, 008 | YES (실측 확인) | `internal/template/templates/.claude/agents/moai/plan-auditor.md` |
| `.claude/agents/moai/sync-auditor.md` | AGI-004..007 | YES (실측 확인) | `internal/template/templates/.claude/agents/moai/sync-auditor.md` |
| `.claude/rules/moai/workflow/spec-workflow.md` | AGI-008, 009 | YES (실측 확인) | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` |
| `.claude/agents/moai/manager-spec.md` | AGI-010, 011 | YES (실측 확인) | `internal/template/templates/.claude/agents/moai/manager-spec.md` |

Mirror 존재는 2026-07-09 `ls` 실측으로 확인됨 (Template 트리는 live의 SUBSET이므로 run-phase 진입 시 재확인 의무 — §C Pre-flight).

## §B Known Issues / Risks

| # | 리스크 | 경감책 |
|---|--------|--------|
| B1 | 라인 번호 drift — spec.md §A.3의 실측 앵커(L127-137 등)는 병렬 세션 커밋으로 이동할 수 있다 | run-phase 진입 시 content-token 앵커(`### M5: Must-Pass Firewall`, `## HRN-003`, `mentally` 등)로 재탐색. 라인 번호는 참고용 |
| B2 | Template Neutrality 위반 — mirror에 SPEC ID/REQ 토큰 유입 시 CI guard(`template-neutrality-check.yaml`) FAIL | M5에서 mirror별 neutrality grep을 검증 배치에 포함 (AC-AGI-012) |
| B3 | `mentally` 토큰 잔존 — iter-1 plan-audit 실측으로 live 1건(self-check 섹션 단독) + template mirror 1건 확인 | M4에서 live 제거 + M5에서 mirror 제거 (AC-AGI-010b/012c). 편집 직전 `grep -n 'mentally'` 재실측으로 신규 출현 여부 확인 |
| B4 | 공유 checkout 병렬 세션 race — working tree에 무관한 미커밋 변경 다수 존재 | pathspec 한정 커밋 (`git add <경로> && git commit -- <경로>`)만 사용. `git add -A` 절대 금지 |
| B5 | plan-auditor.md 자체가 본 SPEC의 감사자다 — M1 편집 후 plan-auditor 재실행 시 자기 정의가 바뀌어 있음 | plan-audit는 SPEC 아티팩트를 심사할 뿐 자기 정의 파일과 무관. 순서 영향 없음. 단 run-phase에서 M1을 먼저 실행하면 이후 감사가 신규 MP-5/6 기준으로 동작함을 인지 |

## §C Pre-flight (run-phase 진입 검증)

run-phase 첫 턴에 단일 병렬 배치로 실행:

```bash
# 1. 결함 baseline 재확인 (4건 전부 여전히 존재하는가)
grep -c '(MP-5)' .claude/agents/moai/plan-auditor.md                      # 기대: 0
grep -c '### Hierarchical-Mode Output Example' .claude/agents/moai/sync-auditor.md  # 기대: 0
grep -c 'run-gate stream' .claude/rules/moai/workflow/spec-workflow.md    # 기대: 0
grep -c 'mentally' .claude/agents/moai/manager-spec.md                    # 기대: 1 (iter-1 실측 baseline)
# 1b. 신규 토큰 baseline (전부 0이어야 RED 상태 유지 — 2026-07-09 실측 완료: 전부 0)
grep -c 'severity=critical' .claude/agents/moai/plan-auditor.md           # 기대: 0
grep -c 'sub-criteria refinement' .claude/agents/moai/sync-auditor.md     # 기대: 0
grep -c 'project-language auto-detection' .claude/agents/moai/sync-auditor.md  # 기대: 0
grep -c 'plan-artifact hash' .claude/rules/moai/workflow/spec-workflow.md # 기대: 0
grep -c 'mentally' internal/template/templates/.claude/agents/moai/manager-spec.md  # 기대: 1 (mirror baseline)
# 2. mirror 존재 재확인
ls internal/template/templates/.claude/agents/moai/{plan-auditor,sync-auditor,manager-spec}.md
ls internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
# 3. 스킬 존재 재확인
ls -d .claude/skills/moai-ref-owasp-checklist .claude/skills/moai-ref-testing-pyramid
# 4. git 상태 — origin 발산 확인 (Pre-Spawn Sync Check)
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
```

어느 baseline이든 이미 수정되어 있으면 (병렬 세션 선점) 해당 REQ를 검증-후-SKIP 처리하고 blocker report로 보고.

## §D Constraints

- doc-only — Go 코드 무변경 (spec.md §C 참조)
- 편집 도구: `Edit` 우선 (Read-before-Edit 준수). `sed`/`awk` 금지
- 커밋: milestone별 pathspec 한정 커밋. 커밋 subject는 `fix(SPEC-AUDIT-GATE-INTEGRITY-001): M{N} <요약>` 형식
- 시간 예측 금지 — 우선순위/순서만 명시

## §E Self-Verification (run-phase 완료 기준)

- E1: acceptance.md §D AC matrix 전 행 PASS/FAIL 판정 + 명령 verbatim 출력 인용
- E2: `make build` exit 0 (template 편집 후)
- E3: template neutrality grep 0건 (AC-AGI-012)
- E4: `moai spec lint` 본 SPEC 무결 (정확한 CLI 플래그 형식은 run-phase에서 `moai spec lint --help`로 확인)
- E5: 편집 diff가 REQ 범위를 벗어나지 않음 (scope discipline — 무관 섹션 무접촉)

## §F Milestones

### M1 — plan-auditor D7/D8 Must-Pass 배선 (REQ-AGI-001..003)

우선순위: High. 대상: `.claude/agents/moai/plan-auditor.md`

1. § M5 Must-Pass Firewall: 도입부 "Four criteria" → "Six criteria" 갱신, `(MP-5)` / `(MP-6)` 정의 추가 — D7/D8 BLOCKING finding이 unresolved면 FAIL, must-pass-equivalent 의미론 명기 (REQ-AGI-002의 severity=critical fold 규칙 포함)
2. § Output Format `## Must-Pass Results` 표: MP-5 / MP-6 행 추가 (`[PASS/FAIL] MP-5 D7 cross-SPEC reconciliation: {evidence or "no BLOCKING finding"}` 형식)
3. Group 7/8 말미에 "BLOCKING → MP-5/MP-6 강제 FAIL" 역참조 1줄 추가

### M2 — sync-auditor 채점모델 정합 + 기계 검증 경로 (REQ-AGI-004..007)

우선순위: High. 대상: `.claude/agents/moai/sync-auditor.md`

1. frontmatter `skills:`에 `moai-ref-owasp-checklist`, `moai-ref-testing-pyramid` 추가 (YAML array 형식 유지)
2. § Evaluation Dimensions 직후에 모델 선택 규칙 명문화: flat = default, `Where evaluator_mode: hierarchical` → HRN-003. 두 모델의 관계는 `sub-criteria refinement` 리터럴 토큰으로 서술 (동일 4차원의 하위 세분화)
3. 차원별 기계 검증 명령 표 추가 — **project-language auto-detection 형식** (리터럴 토큰 포함, 자체 완결 표): 프로젝트 언어 자동감지 + 미설치 도구 graceful skip + 4개 언어 동등 예시(Go/Python/Node.js/Rust — 어떤 언어도 PRIMARY 승격 금지). 각 차원 최소 1 명령 + "verbatim 출력을 Evidence 셀에 인용" 의무 (vci §1.1 surface 2 cross-ref). 이 형식으로 live/mirror 편집 내용이 동일 (D3/MP-4 해소)
4. § HRN-003 하위에 `### Hierarchical-Mode Output Example` worked example 추가 (sub-criterion 행, anchor 점수, min aggregation, must-pass verdict 포함 — 리터럴 heading 유지)

### M3 — 보고서 파일명 split-brain 문서화 (REQ-AGI-008, 009)

우선순위: High. 대상: `.claude/rules/moai/workflow/spec-workflow.md` + `.claude/agents/moai/plan-auditor.md`

1. spec-workflow.md § Report Persistence: 두 스트림 명문화 — "plan-phase review stream" (`{SPEC-ID}-review-{N}.md`, plan-phase 적대적 리뷰, spec-assembly 소비) vs "run-gate stream" (`<SPEC-ID>-<YYYY-MM-DD>.md`, Phase 0.5 게이트, `audit_report.go` 구현). skip-eligibility 지정 (D4 교정 — Go 구현 정합): review 스트림의 final-iteration verdict가 run-gate 입력; artifact-hash 검사는 plan-artifact hash(`audit_cache.go` `ComputeHash` — plan artifacts whitespace-normalized SHA-256, cache key = specID+planArtifactHash) 재계산·대조; run-gate date-file은 verdict 기록 표면일 뿐 hash 대상 아님
2. plan-auditor.md § Output Format: 자기 보고서가 "plan-phase review stream"임을 명시 + run-gate stream 상호 참조 1-2줄
3. 두 파일 모두 상대 문서로의 mutual cross-reference 포함
4. **인접 문장 조정 (iter-2 N2)**: spec-workflow.md § skip policy 조건 3의 artifact-hash 파일 열거가 5-파일(spec/plan/acceptance/**research/design**, tasks 제외)로 되어 있어 Go 구현 `planArtifactNames` 4-파일(acceptance/plan/spec/**tasks** — `audit_cache.go:63-68` 실측, iter-2 감사 확인)과 상충 — M3.1 편집 시 이 인접 문장을 함께 조정한다: 4-파일 목록으로 정정하거나, "기계 hash 대상은 ComputeHash 4-파일; research.md/design.md 변경은 수동 skip-판단의 보수적 고려 대상"으로 구분하는 각주를 단다 (기존 doctrine-vs-Go drift의 노출이며 본 SPEC이 유발한 것 아님 — 편집은 run-phase M3 소관)

### M4 — manager-spec Bash 검증 전환 (REQ-AGI-010, 011)

우선순위: High. 대상: `.claude/agents/moai/manager-spec.md`

1. § SPEC ID Pre-Write Self-Check Protocol의 Step 2 "mentally" 제거 → 실행형 Bash one-liner로 대체: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` (verbatim 출력 = 증거; `\d` 미지원 각주 포함)
2. 4-step decomposition 지침 + regex 리터럴 + worked example 1건 유지
3. 역사 서사 압축: L32 chain context 문단 + 5-incident 열거 제거 (~30줄 → 핵심만). 제거분은 run-phase에서 memory/lessons로 이관 노트 남김
4. 동일 섹션의 stale 라인 인용 교정 (D6): `lint.go:573` → `lint.go` `specIDPattern` (실측 L649; content-token 앵커 우선, B1 방침과 일관)

### M5 — Template mirror 4건 + make build + 전체 검증 배치 (REQ-AGI-012)

우선순위: High (M1-M4 완료 후). 순서 의존: M1-M4 → M5

1. 4개 mirror에 동일 편집 적용 + Neutrality strip (SPEC-AUDIT-GATE-INTEGRITY / REQ-AGI 토큰, 내부 날짜·감사 인용 제거)
2. `make build` 실행, exit 0 확인
3. acceptance.md §D 전체 AC 검증 배치를 단일 턴 병렬 실행 (file-redirect contract 준수)
4. milestone별 pathspec 커밋 정리 확인

## §G Anti-Patterns

- **AP-1**: BLOCKING → FAIL 배선을 prose 1줄로만 추가하고 Must-Pass 표를 미갱신 (동일 drift 재발 — 기각된 대안 설계)
- **AP-2**: mirror 편집 시 live 파일 내용을 그대로 복사 (Neutrality 위반 — SPEC ID/내부 날짜 유입)
- **AP-3**: `git add -A` 또는 디렉터리 광역 add (병렬 세션 미커밋 변경 휩쓸림)
- **AP-4**: AC 검증을 "수정했으니 통과할 것" 추정으로 보고 (vci §1.1 위반 — 본 SPEC이 고치려는 바로 그 병리)
- **AP-5**: manager-spec 역사 서사를 지우면서 4-step 프로토콜 자체를 훼손 (압축 대상은 서사, 프로토콜 아님)
- **AP-6**: spec-assembly.md를 필수 편집 대상으로 확대 (Out of Scope — 이미 review 스트림 일관 사용)

## §H Cross-References

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surfaces 1-2, §3.2 (Evidence = verbatim output)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (12-field SSOT)
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase 0.5 Plan Audit Gate / § Report Persistence
- `internal/runtime/audit_report.go` + `audit_gate.go` (L221-232) + `audit_cache.go` (L30-58 `ComputeHash`) — run-gate stream · plan-artifact hash 구현, 무변경 참조 대상
- `internal/spec/lint.go` `specIDPattern` (실측 L649 — canonical SPEC ID regex, 무변경 참조 대상)
- `CLAUDE.local.md` §2.1 Template Content Neutrality + `.moai/docs/template-internal-isolation-doctrine.md` §25.1
- 감사 provenance: 3-agent 병렬 감사 2026-07-09 (agent-definitions / workflow-doctrine / SDD 웹 리서치)
