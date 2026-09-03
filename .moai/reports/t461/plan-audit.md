# t461 plan-audit — SPEC-PRECOMMIT-GATE-SCOPE-001

- 감사일: 2026-09-03 / 감사자: plan-auditor (iteration 1, Tier M — retry ceiling 2)
- 대상 트리: `WT-precommit-gate-scope` @ `a239cf050` (`.claude/worktrees/t461`)
- 감사 산출물: spec.md + plan.md + acceptance.md + progress.md (Tier M 3종 + progress, 전부 실독)
- **판정: FAIL** (MP-7 clarification gate — 점수와 무관한 must-pass)
- 종합 점수: **0.875** (Tier M 임계 0.80 충족 — 그러나 M5 방화벽이 점수를 무효화한다)

---

## 1. 판정 요약

SPEC의 사실 기반은 실측 전건 적중이고 요구사항·AC 품질은 높다(8 REQ 전부 GEARS 5패턴 적합, 8 AC 전부 기계판정형 Given-When-Then). 결함 진입 커밋(`883d53852`/`52b5e4bf5`), twin 3,245 bytes 바이트 동일, 설치 호출점 `:575`, gate.yaml 기본값(`enabled: true`/`skip_tests: false`), `.moai/config/` 소거 뿌리, `QualityGate.Run`의 `Enabled` 공유 스위치 — 전부 본 워크트리에서 재검증해 SPEC 기술과 일치했다.

유일한 판정 장애는 **plan.md:70의 미해소 `[NEEDS CLARIFICATION]` 마커**다. 이것은 사양이 불완전해서가 아니라, 리드 지시("확정")에 대해 plan이 최종 선택을 운영자 게이트로 남겨둔 상태다. 마커 규약(`.claude/skills/moai-workflow-spec/SKILL.md:171-191`)은 이 마커가 "Implementation Kickoff Approval 이전에 반드시 해소"되어야 하고 plan-auditor가 이를 clarification gate finding으로 검출하도록 정한다. 따라서 본 판정은 **사양 보완 요구가 아니라 결정 게이트 라우팅**이다: 운영자가 (a)/(b)를 확정하면 manager-spec은 마커를 확정 기록으로 대체하고, 결함 델타(D1~D3)만 재감사하면 된다. 점수(0.875 ≥ 0.80)는 MP-7 FAIL을 상쇄하지 못한다(M5 방화벽, score-independent).

## 2. Must-Pass 결과

| MP | 결과 | 근거 |
|----|------|------|
| MP-1 REQ 번호 일관성 | **PASS** | spec.md:69,73,81,85,89,93,97,101 — REQ-001~008 연속, 공백·중복 0 (`grep -n "^### REQ-"` 실측) |
| MP-2 GEARS 형식 | **PASS** | 판정 레이어 = spec.md의 REQ-XXX(요구 레이어). REQ-001/005/008 Ubiquitous("The … shall"), REQ-002/003/006 Event-driven("When … shall"), REQ-004 Where(capability gate), REQ-007 State-driven("While … shall") — 8/8 적합. acceptance.md의 Given-When-Then은 검증 레이어로 MP-2 대상 아님(M3 § Scope 2계층 표). 비고: REQ-005가 `preCommitHookContent`·`TestPreCommitTemplateMatchesConstant` 식별자를 인용하나 이는 twin 계약의 대상물 자체를 지칭하는 주제어지 HOW 처방이 아니므로 위반으로 채점하지 않음 |
| MP-3 YAML frontmatter | **PASS** | spec.md:2-13 — 12 필드 전부 존재·형식 적합: `id`(패턴 합격), `title`(인용 문자열), `version "1.0.0"`, `status: draft`(enum), `created/updated: 2026-09-03`(ISO), `author`, `priority: P1`(enum), `phase: "v3.2.0 target"`(릴리스 타깃 — 스테이지명 아님), `module`, `lifecycle: spec-anchored`(enum), `tags`(CSV 문자열). snake_case 별칭(`created_at` 등) 0. `related_specs`는 추가 필드로 허용. 형제 산출물(plan/acceptance)에 `status:` 없음 — Artifact Statelessness 합격 |
| MP-4 언어 중립성 | **N/A (auto-pass)** | 단일 기능(훅/게이트 계약) SPEC — 16개 프로그래밍 언어 열거 의무 대상 아님. "16개 프로그래밍 언어" 중립 서술만 존재(spec.md:28-29) |
| MP-5 D7 cross-SPEC | **PASS** | 참조 SPEC 3건(`SPEC-PRECOMMIT-001`, `SPEC-PRETOOL-GATE-MOVE-001`, `SPEC-PRECOMMIT-PRESERVE-001`) 전부 `.moai/specs/`에 존재하고 `status: completed` — retired/superseded/archived 없음 → D7-4 미트리거, BLOCKING 없음 |
| MP-6 D8 syscall | **PASS (auto)** | `grep -c syscall spec.md` = 0 → D8-4 자동 통과 |
| MP-7 clarification gate | **FAIL** | `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → `plan.md:70:### [NEEDS CLARIFICATION: 축 (a) vs (b) 최종 확정]` — 감사 시점 미해소. research.md 부재(Tier M 정상 구성). 해소 의무는 acceptance.md:105(DoD-3 "plan.md §C의 [NEEDS CLARIFICATION] 마커가 해소돼 있다")와 progress.md `open_decisions` 2건에 이미 추적돼 있음 |

## 3. 차원 점수 (rubric-anchored)

| 차원 | 점수 | 밴드 | 근거 |
|------|------|------|------|
| Clarity | 1.0 | 1.0 | REQ 전 건 단일 해석. 축 조건부는 모호성이 아니라 명시 관리 대상이다 — spec.md:77-79 인용블록이 축별 충족 형태를 기술하고, acceptance.md:8-10 헤더와 AC-002의 축 분기(:25-26)가 판정 조건을 확정 경로에 둔다 |
| Completeness | 0.75 | 0.75 | frontmatter 12 필드 완비, HISTORY(:19)/배경 WHY(:25)/요구사항(:67)/AC(acceptance.md)/Out of Scope 3절(spec.md:113,120,127 — H2+특정 `-` 불릿, `internal/spec/lint.go:1101`의 substring 규약 충족). 유일 결손 = 명시적 WHAT/Scope 개관 절 부재(비핵심 1절) → 0.75 밴드 |
| Testability | 1.0 | 1.0 | AC 전 건 기계판정형: AC-001 grep 4문자열(acceptance.md:16-17), AC-003 정확 명령행(:32), AC-004 fixture 파일 직접 비교(:41), AC-005 `git diff <base> -- …` 명령 명시 + fast-subset 불변 항목(:47-50 — 실제 템플릿 :13/:51/:46-48과 일치), AC-006 종료코드 1(:56), AC-007 종료코드 0(:63). weasel word 0 |
| Traceability | 0.75 | 0.75 | AC-001→REQ-003, AC-002→REQ-002, AC-003→REQ-005, AC-004→REQ-006, AC-005→REQ-008, AC-006→REQ-004, AC-007→REQ-007, AC-008→REQ-002 — 전 건 유효 참조. 단 **REQ-001 미커버**(acceptance.md 전체에서 REQ-001 참조 0) → "1 REQ 미커버" 밴드 |

**종합 = 0.875** (4차원 산술평균). Tier M 임계 0.80 초과하나 MP-7 FAIL이 판정을 결정.

## 4. 사실 주장 검증 (감사 명령 + 출력)

### (i) `QualityGate.Run` Enabled 쇼트서킷 + 단독 CLI 스위치 공유 — **확인. 계획의 중심 메커니즘 주장은 참이다**

- `grep -n "func (g \*QualityGate) Run" -A 5 internal/hook/quality/gate.go` →
  `gate.go:381-384`: `if !g.config.Enabled { return true, "" }`
- `grep -n "NewQualityGate\|gate.Run(" internal/cli/gate.go` →
  `gate.go:96-97`: `gate := quality.NewQualityGate(cfg)` / `passed, output := gate.Run(ctx)` (단독 `moai gate` = `runGate`)
- 결론: 단독 실행도 동일 스위치를 통과하므로 템플릿 `gate.enabled` 반전은 단독 `moai gate`까지 끈다. plan.md §B Known Issue 2와 제약 6의 전제가 실측으로 성립 — **(b) 권고의 제약 조건은 무너지지 않았다.** 권고의 선례 근거(BranchGuard 기본 false 등)도 `main-checkout-branch-guard.md` § Mechanical Enforcement("distributed default is false")로 확인.

### (ii) 템플릿 gate.yaml 존재·기본값 — **확인**

- `internal/template/templates/.moai/config/sections/gate.yaml:14-19` → `enabled: true`, `skip_tests: false`, `disabled_steps: {}`. 로더 배선: `loader_gate.go:20` `loadGateSection` ← `loader.go:86`.

### (iii) `CleanMoaiManagedPaths`의 `.moai/config/` 소거 뿌리 — **확인**

- `internal/cli/update/deploy/deploy.go:92` — 주석 "The .moai/config/ directory is deleted entirely (backup was done by the Backup step)", `:187` — "Clean .moai/config/ entirely". REQ-006의 존재 근거가 성립한다.

### (iv) t237 = issue #1641, 동일 twin 파일 — **확인**

- `.moai/specs/SPEC-PRECOMMIT-PRESERVE-001/spec.md:430` — "Card t237 (issue #1641) edits `preCommitHookContent` and its template twin"; `:159`/`:208` — open + verified patch (`t312-precommit-vet @ b6f478b1a`). 충돌 기록이 형제 SPEC에 실재하며 본 SPEC의 scope-out(plan.md §B.3, REQ-008/AC-005)은 그 기록과 정합.

### 부수 실측 — SPEC §A.2 좌표표 전 건 적중

| 주장 | 검증 명령 | 결과 |
|------|-----------|------|
| 설치 호출점 | `grep -n "installPreCommitHookOptional(projectRoot" internal/cli/update_template_sync.go` | **:575** — 카드의 :574를 SPEC이 :575로 바로잡은 것까지 정확 |
| twin byte identity | `sed -n '40,112p' … \| sed '1s/^const preCommitHookContent = `//'` → `cmp` | **TWIN-IDENTICAL**, 쌍방 **3,245 bytes** (SPEC의 3,245 주장 정확) |
| twin 강제 테스트 | 파일 확인 | `TestPreCommitTemplateMatchesConstant` 존재 (`internal/cli/hook_install_precommit_test.go`) |
| 게임 규칙 기본값 | `grep -n "NewDefaultGateConfig" -A 5 internal/config/defaults.go` | `defaults.go:585-588` `Enabled: true, SkipTests: false` |
| remedy 키 3종 | `grep -n "type GateConfig struct" -A 22 internal/config/types.go` | `types.go:813-834` — `enabled`/`skip_tests`/`disabled_steps` 전부 존재 |
| 결함 진입/ 선행 커밋 | `git show --no-patch --format='%h %ad %s' 883d53852 52b5e4bf5` | `883d53852` 2026-07-28 SPEC-PRETOOL-GATE-MOVE-001 (#1189) / `52b5e4bf5` 2026-07-05 SPEC-PRECOMMIT-001 — 날짜·내용 전부 일치 |
| 훅 heavy gate 무조건 실행 | 템플릿 :64-70 실독 | `if command -v moai … then if ! moai gate … exit 1` + 유일 안내 `SKIP_MOAI_PRECOMMIT=1` — REQ-003의 문제 정의와 일치 |
| plan이 인용한 부수 파일 | `ls` | `audit_struct_yaml_symmetry_test.go`, `template-internal-isolation-doctrine.md`, `template-neutrality-check.yaml` 전부 존재 |

## 5. 감사 질문 — `[NEEDS CLARIFICATION]` 마커의 성격 판정

**실질: 정당한 gate-pending / 기계: MP-7 위반 → FAIL.**

- 규약 적합성: 마커는 규약이 허용하는 유일 위치(plan.md)에 있고(SKILL.md:175 "ONLY in plan.md and research.md"), 해소 시점(Implementation Kickoff Approval 이전)과 절차(운영자 AskUserQuestion)가 규약 그대로다(SKILL.md:173,188-191). acceptance.md DoD-3과 progress.md `open_decisions`가 해소 의무를 자기 추적한다 — 미해소 상태가 조용히 남는 구조가 아니다.
- 내용 적합성: plan은 (b)를 권고하며 근거가 실측 가능하다(선례 3건의 기본값 — 실측 확인; Known Issue 2 — gate.go:381-384 + cli/gate.go:96-97 실측 확인; 축 (a)의 16-언어 staged-필터 비용 논거). 마일스톤 순서도 결정 역전가능성 순(M1 스키마/배포 계약 최전단)으로 정합한다.
- 그러나 MP-7은 점수 독립 must-pass이고 "감사 시점 미해소 마커 = FAIL"이 기계 규칙이다. 규약 자체가 이 FAIL을 **해소 라우팅 신호**로 설계했다(감사 → clarification gate finding → 운영자 AskUserQuestion → 마커 대체). 따라서 이 FAIL은 manager-spec의 불성실이 아니라 **결정 권한의 정당한 이관 기록**이며, 수리도 사양 재작성이 아니라 결정+기록 대체다.

## 6. 결함 목록

- **D1** — MP-7 clarification gate — `plan.md:70` — `### [NEEDS CLARIFICATION: 축 (a) vs (b) 최종 확정]` 감사 시점 미해소 — Severity: **BLOCKING** — Class: blocking — 수리: (1) 운영자가 Implementation Kickoff Approval에서 (a)/(b) 확정, (b) 확정 시 opt-in 메커니즘 1~3도 동시 확정 → (2) manager-spec이 plan.md:70 마커를 확정 기록으로 대체하고 axis-conditional AC(AC-002/AC-006)의 통과 조건 확정 → (3) plan-auditor iteration 2에서 D1~D3 델타 재감사.
- **D2** — remedy 안내가 공유 스위치를 가리킨다 — `spec.md:83`(REQ-003) / `acceptance.md:16-17`(AC-001) — REQ-003이 요구하는 4문자열 중 `gate.enabled`은 Known Issue 2의 단독-CLI 공유 스위치라서, 실패 메시지를 따라 이 키를 반전시킨 사용자는 단독 `moai gate`까지 꺼뜨린다(제약 6 위반 경로). 축 (b)의 실제 remedy인 pre-commit opt-in 키(plan.md §C 후보 1~3)는 REQ-003/AC-001이 요구하지 않는다 — Severity: **SHOULD-FIX** — Class: blocking — 수리: (b) 확정과 같은 수정에서 REQ-003에 opt-in 키 명시 요구를 추가(또는 `gate.enabled` 안내에 공유-스위치 경고를 요구), AC-001 grep 목록에 반영.
- **D3** — REQ-001 미커버 — `acceptance.md:12-70` — AC-001~008이 REQ-002~008만 참조; REQ-001(커밋 단위 품질 계약, `spec.md:69-71`)에 대응 AC 없음 — Severity: **SHOULD-FIX** — Class: blocking — 수리: AC-002의 참조를 "REQ-001, REQ-002"로 확장하거나 REQ-001 전용 AC 1건 추가.
- **D4** — 명시적 WHAT/Scope 개관 절 부재 — `spec.md` 전체 — Severity: MINOR — Class: optional.
- **D5** — REQ-003 부제 "Event-detected" — `spec.md:81` — GEARS 5패턴명 아님(Event-driven 의정으로 추정), 진술문 자체는 적합 — Severity: MINOR — Class: optional.

## 7. Gaps (관측하지 않은 것)

- `go test`를 실행하지 않았다(감사 읽기전용 + 로컬 전체 스위트 금지). twin 동일성은 `cmp` 소스 비교(TWIN-IDENTICAL)로 확인했을 뿐 `TestPreCommitTemplateMatchesConstant`의 실행 출력이 아니다.
- `moai update`를 실제 실행해 REQ-006의 gate.yaml 소멸을 재현하지 않았다 — `deploy.go` 소거 뿌리 코드와 백업 주석(`:92`) 확인까지만 수행. Backup 단계가 복원까지 하는지는 확인하지 못했다(주석은 "deleted entirely"만 말한다).
- REQ-006 수리 메커니즘 후보(i~iii)의 기술적 타당성은 run-phase 설계 몫으로 남긴다 — plan은 이를 Constraint 3으로 올바르게 위임했다.
- 크로스모델 2차 감사(`mcp__moai__audit_multi`)를 호출하지 않았다 — 프로젝트 `audit_model` 설정을 확인하지 않은 채 단일(Claude) 판정으로 완료. 백엔드 fail-open 특성상 판정 자체에는 영향이 없으나 다중 백엔드 수렴 의견은 부재하다.

## 8. 권고 (FAIL 수리 경로)

1. **(운영자)** Implementation Kickoff Approval 게이트에서 축 (a)/(b)를 확정한다. (b) 권고의 근거는 본 감사에서 실측으로 유효 확인됐다(§4-(i)). (b) 확정 시 opt-in 메커니즘 1~3 중 하나를 같은 라운드에서 확정한다.
2. **(manager-spec, 델타 수정)** plan.md:70 마커를 확정 기록으로 대체한다. (b) 확정 시 REQ-003/AC-001에 pre-commit opt-in 키를 반영한다(D2). REQ-001 커버 AC를 반영한다(D3). D4/D5는 임의.
3. **(plan-auditor)** iteration 2는 본 보고서의 D1~D3 델타 재감사로 한정한다(Regression Check 포함). 전면 재감사 불요.

---

# Iteration 2 — Delta Re-audit (2026-09-03, Tier M ceiling 2/2)

- 대상: iter-1 결함 D1~D5 델타 + 신설 web 축(REQ-009/AC-010/plan M2/제약 7-8) + D5 재판정 + 미해소 질문 2건
- 트리 귀속: HEAD `4060b1dbe` (lane의 plan-phase 산출물 단일 커밋) 위의 **미커밋 수정 4파일**(`git status --porcelain`: spec/plan/acceptance/progress 전부 M) — 본 감사는 작업 트리 상태를 심판하며, 커밋은 레인이 판정 후 수행
- **판정: PASS**
- 종합 점수: **1.0** (Clarity 1.0 / Completeness 1.0 / Testability 1.0 / Traceability 1.0 — monotonic ≥ 0.875 ✓, Tier M 임계 0.80 ✓)

## Iter-2 Must-Pass 결과

| MP | 결과 | 근거 |
|----|------|------|
| MP-1 | **PASS** | REQ-001~009 연속(spec.md:77,81,89,93,97,101,105,109,113), AC-001~010 연속(acceptance.md:11~84) — 공백·중복 0 |
| MP-2 | **PASS** | 9/9 REQ GEARS 5패턴 적합. REQ-009 "Where … shall"(spec.md:115, Capability gate). REQ-003 "Event-detected"는 D5 판정대로 정캐논(아래). REQ-002 라벨의 axis-conditional 잔존은 D6(경미)로 별도 기록 — 진술문 자체는 적합 |
| MP-3 | **PASS** | frontmatter 12 필동 불변·적합, HISTORY 수리 행 추가(spec.md:24) |
| MP-4 | N/A | 불변 |
| MP-5 | **PASS** | 관련 SPEC 3건 불변, 전부 completed |
| MP-6 | **PASS** | 성장한 spec.md에서 `grep -c syscall` = 0 → 자동 통과 |
| MP-7 | **PASS** | 정확 그렙 `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → plan.md 0매치, research.md 부재(Tier M 정상). SPEC 디렉터리 전체의 유일 관련 텍스트는 spec.md:24 HISTORY의 "plan.md NEEDS CLARIFICATION 해소" — 해소 사실의 역사 기록이지 미해소 마커가 아니며, MP-7 스캔 범위(plan.md/research.md) 밖 |

## Iter-2 결함 델타

- **D1 RESOLVED (was BLOCKING)** — plan.md:70 마커 리터럴 소멸(그렙 실측 0매치). plan.md:66-78 `### 확정된 설계 (Implementation Kickoff Approval — 운영자 결정, 2026-09-03)`에 3건 기록: ① 축 (b), ② 메커니즘 1(`MOAI_PRECOMMIT=1` 환경 마커 + 신규 키 `gate.pre_commit.enabled`는 마커 하에서만 존중, 단독 `moai gate` 불변, 새 서브커맨드 없음, 러너 분기점 1개, 기각 사유 포함), ③ REQ-009 `moai web` 편집 표면(저장 경로 정확히 `.moai/config/sections/gate.yaml`). plan.md:78 axis-conditional 해제 선언. progress.md `open_decisions` 제거 → `decisions_recorded` 3건 전환.
- **D2 RESOLVED (was SHOULD-FIX)** — spec.md:91 REQ-003이 경로 + 4키(`gate.pre_commit.enabled` 포함) 명시, `SKIP_MOAI_PRECOMMIT=1` 유지. acceptance.md:16-18 AC-001이 5문자열 grep 기계판정 + `SKIP_MOAI_PRECOMMIT=1` 유지 절. AC-001의 Given이 opt-in 상태로 정합화됨.
- **D3 RESOLVED (was SHOULD-FIX)** — acceptance.md:75-82 신설 **AC-009** — REQ-001 커버. 마커 유/무 대비(`MOAI_PRECOMMIT=1 moai gate` → exit 0·heavy 미실행 vs 마커 없는 단독 → 기존 계약대로 실패)가 커밋 단위 계약의 기계적 증명. 이진 판정형. spec.md:127 성공 기준에도 REQ-001 항목 추가.
- **D4 RESOLVED (bonus, was MINOR/optional)** — spec.md:28-33 `### 범위 요약 (WHAT)` 신설.
- **D5 CLOSED — auditor-error (기각 인정, 변경 없음)** — 반박 실측 확인: `.claude/skills/moai-workflow-spec/SKILL.md` 5패턴 표 제5행이 문자 그대로 `Event-detected (replaces IF/THEN)`이고, `.claude/skills/moai-foundation-core/SKILL.md`(~:113)도 "Event-detected (replaces the deprecated conditional modality)"로 열거. 정캐논 확산 6개 스킬 파일. REQ-003(실패 메시지 요구)은 불요조건 감지 시 대응 — Event-detected 조형 그 자체다. iter-1의 나의 지적은 저장소 정캐논이 아닌 일반 하네스 루브릭 목록에 근거한 **감사자 오류**였다.
- **D6 신설 — MINOR / optional** — spec.md:81 REQ-002 부제 "(Event-driven, axis-conditional)"과 :85-87 인용블록이 축 (a)/(b) 분기를 현재형으로 기술 — 축 확정(plan.md:78 해제 선언, acceptance.md는 갱신) 후 미스윕된 표기. cross-file 개정 스윕 미완료형(verification-completeness §3 패턴). AC 통과 조건에는 영향 0. 수리: REQ-002 부제/블록을 확정 사후 서술로 1회 정리(임의).
- **D7 신설 — MINOR / optional** — plan.md M2 단계 목록이 seam 라우팅 등록면 2곳을 명명하지 않는다: `RouteForSection`(`internal/settings/sectionroute.go:113`)이 "gate"에 `RouteSeam`을 반환해야 하고 `sectionRootKeys`(sectionwrite.go)에 "gate" 화이트리스트가 필요. 누락 시 `WriteSectionViaSeam`이 "not seam-writable" 오류로 **요란하게 실패**하므로 침묵 결함이 될 수 없고 AC-010(2)이 가시 적적으로 잡는다. run-phase 착지 시 자기 검출.

## Web 축 사실 주장 실측 (전건 확인 — SPEC 주장과 불일치 0)

| 주장 | 검증 명령·출력 | 판정 |
|------|----------------|------|
| `SectionGate` 부재 (오늘) | `internal/settings/schema.go:26-63` SectionID 상수 목록 — identity/language/launch/statusline/quality/git_convention/quality_extras/git_strategy/llm/workflow/harness/ralph/feedback/observability/security/handoff/cache/report/mcp/crosssession — **gate 없음** | 확인. run-phase에서 신설 필요하다는 plan M2 전제 참 |
| `schemaSectionMetas()` | `schemaform.go:199` (plan은 :201 — 2행 drift, 심볼 인용 관례 범위) | 확인 |
| `parseSchemaForm` + `__present` companion | `schemaform.go:288` 주석 "companion(name+\"__present\") 패턴으로 'unchecked → false'와 '미제출 → preserve'", `:298` 함수 | 확인 |
| `schemaEditableField` 자격 | `schemaform.go:280-281` `Persist.Kind == PersistSeam \|\| PersistTypedSection` | 확인 — 위 FieldDef는 편집 대상 자격 보유 |
| workflow_agents 은닉 = 레지스트리 부재 | `agent_settings_test.go` (d) 블록(~:97, 폼 컨트롤 미렌더 단언) + `TestWorkflowAgentsWebSubmissionIgnored`(~:235, 제출 무시·workflow.yaml 불변) | 확인 — 노출은 FieldDef 등록으로 충분, 별도 플래그 불필요 |
| `WriteSectionViaSeam` 저장 경로 | `sectionwrite.go:52` — `filepath.Join(projectRoot, ".moai", "config", "sections", section+".yaml")` + `yamlpatch.PatchFile`(주석 보존) → section="gate"면 정확히 gate.yaml | 확인 |
| 명명 중복(`git_strategy.<mode>.hooks.pre_commit`) | `validation.go:318+` `checkStringField` 다수 사용 | 확인 — 제약 8 + REQ-009 명명-중복 주의 블록의 근거 성립 |

## 미해소 질문 2건 — run-phase 위임 적절성 판정

1. **게이트 패널 배치(신규 탭 vs 기존 탭)** — **수용.** 표현층 결정이고 AC-010은 노출+저장 경로+i18n 라벨을 심판하며 배치는 통과 조건이 아니다. plan M2 단계 2가 지배 선례(이름 기반 배치, 영속화 경로 불변)를 명명해 위임의 형태가 규율돼 있다. 스키마·영속화라는 역전하기 어려운 결정은 M1/M2에 구속돼 있으므로 배치 위임이 역전가능성 순서를 훼손하지 않는다.
2. **REQ-006 메커니즘(병합 vs 소거 제외)** — **수용.** AC-004가 결과(값 생존 — 손편집·web 작성 모두)를 기계 판정하고(plan.md M1 "어느 쪽이든 AC-004가 기계 판정한다"), 두 후보 모두 같은 이진 AC로 귀속된다. 이것은 운영자가 결정할 미해소 질문이 아니라 승인된 결과 계약 안의 구현 전략 선택이다 — NEEDS CLARIFICATION 형태가 아니다.

## Iter-2 Gaps

- `go test` 미실행(감사 읽기전용) — AC-009/AC-010의 RED는 run-phase M4 몫. 본 감사는 계약의 판정 가능성만 심판.
- `schemaSectionMetas` :201 vs 실측 :199 등 줄번호 미세 drift 1건 — 심볼 인용 관례 범위 내.
- 크로스모델 2차 감사(`audit_multi`) 미호출 — iter-1과 동일.
- iter-1에서 지적했던 `git-strategy.yaml` 재적용 전례 대비(REQ-006이 기계 AC를 갖는 이유)는 코드 변경으로 검증할 단계가 아니므로 plan 차원에서만 확인.

## Iter-2 판정 근거 요약

MP-7 해소(유일한 must-pass 장애 제거) + iter-1 결함 D1~D4 전건 해소 + D5 감사자 오류로 정정 + web 축의 5건 코드 주장 전건 실측 확인 + 미해소 질문 2건 모두 AC 심판형 위임으로 적절. 잔여 D6/D7는 optional 등급의 경미 항목으로 PASS 판정에 영향 없음(M6 — optional 발견 목록이 FAIL을 만들지 않는다). run-phase 진입 가능.

