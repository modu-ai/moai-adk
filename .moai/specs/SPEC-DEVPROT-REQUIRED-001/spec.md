---
id: SPEC-DEVPROT-REQUIRED-001
title: "develop 브랜치 필수 상태검사 승격 설계 — 무PR 직접 push 통합 모델과의 공존"
version: "0.2.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "branch-protection, required-checks, develop, git-flow, ci, runbook"
tier: M
---

# SPEC-DEVPROT-REQUIRED-001 — develop 브랜치 필수 상태검사 승격 설계

## HISTORY

| 날짜 | 버전 | 내용 |
|---|---|---|
| 2026-09-02 | 0.1.0 | 최초 작성 (card t324, plan-phase, design-only). research.md 측정값 + 본 문서 작성 중 재검증(workflow YAML 재판독 + live check-runs 실측)으로 수립. **research.md §2.2의 codeql 전제가 반증되어 §1.3에 정정 반영.** |
| 2026-09-02 | 0.2.0 | plan-audit iteration 1 개정 (D1/D2): 해소 대기 마커 3건 운영자 해소로 DECIDED — phase-1 필수 세트에 `Analyze (Go) (go)` 포함(4컨텍스트), verify ref = 카드별 `verify/<card-id>`, B1 대기 = `gh run watch --exit-status`(모호 시 commit status 폴백). AC 3건 추가(총 14). |

---

## §1 배경과 측정된 현재 상태

### 1.1 문제

이 리포는 2026-08-27부터 git-flow 카드 모델로 운영된다: 카드 워크트리(`WT-<slug>`)는 `develop`에서 분기하고, 완료 카드는 통합 워크트리에서 로컬 `develop`에 `git merge --no-ff`로 합쳐진 뒤 `origin/develop`에 직접 push된다. 카드 단위 PR은 존재하지 않으며 `origin/develop`의 CI가 통합을 판정한다(`.claude/rules/local/gitflow-lane-protocol.md` §4).

현재 `develop`은 보호되지 않음(§1.2). 카드의 사전 검증 없이 깨진 병합이 develop에 착지할 수 있고, 그 결함은 push **후** CI에서야 발견된다. 본 SPEC은 develop에 required status checks를 도입하는 변경을 **설계**한다 — 단, 무PR 직접 push 모델이 보호 하에서도 살아남는 통과 경로를 함께 설계하는 것이 조건이다. **보호 설정 적용 자체(gh api PUT)는 운영자 게이트로 남는다**(Out of Scope).

### 1.2 현재 상태 (측정: 2026-09-02, worktree t324 @ tree `fa8ff89ba` + live `gh api` GET)

| 항목 | 값 | 비고 |
|---|---|---|
| develop 보호 | **없음** — `gh api repos/modu-ai/moai-adk/branches/develop/protection` → 404 "Branch not protected" | research.md §1.1 |
| main 보호 | `strict: false`, contexts 5개(`Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`, `Release PR Multi-OS Gate`), `enforce_admins: true`, PR 필수(0 approvals) | research.md §1.2 (live read) |
| ci.yml push 트리거 | `branches: [main, develop]` | 본 문서 작성 중 yq로 재확인 |
| codeql.yml push 트리거 | `branches: [main, develop]` | 상동 |
| develop HEAD `fa8ff89ba` check-runs | `Test (ubuntu-latest)`·`Lint`·`Build (linux/amd64)`·`Analyze (Go) (go)` 전부 success 보고 | live 실측 (§1.3) |

`.moai/docs/git-local-workflow-doctrine.md` §23.2(2026-07-20 스냅샷)의 main 보호 수치는 live 상태와 어긋난다(`strict`, conversation-resolution, context 수). **보호 상태의 SSOT는 live API 조회다.** 본 SPEC run-phase는 이 낡은 스냅샷에 고지를 붙인다(REQ-DPR-013).

### 1.3 research.md §2.2 전제 정정 — `Analyze (Go) (go)`는 docs-only push에서도 보고된다

research.md §2.2는 "codeql의 skip-marker가 PR 전용이라 docs-only 직접 push에서는 `Analyze (Go) (go)`가 아예 보고되지 않는다"고 결론했다. **이 전제는 반증됐다:**

- codeql.yml의 **실제** `analyze` 잡 조건은 `needs.detect.outputs.go_code == 'true' || github.event_name == 'push' || github.event_name == 'schedule'`이다(작성 중 재판독, codeql.yml analyze 잡 `if:` 블록). push 이벤트에서는 go_code와 무관하게 항상 실행된다. PR 전용인 것은 보조적인 `analyze-skip-marker` 잡뿐이며, 이는 PR 이벤트에서만 필요한 보완이다(codeql의 `pull_request` 트리거는 `branches: [main]`뿐).
- **판별 실측**: develop HEAD `fa8ff89ba`는 docs-only 머지(변경 파일 전부 `.md`, first-parent diff로 확인)이며, 그 SHA의 check-runs에 `Analyze (Go) (go)` = **success**가 보고됐다(`gh api .../commits/fa8ff89ba.../check-runs`, 이번 실행).

설계 귀결: `Analyze (Go) (go)`는 **동반 변경 없이도** develop required 후보로 자격이 있다. phase-2로 남길 이유가 있다면 그것은 "push마다 CodeQL 분석이 도는 지연 비용"이지 자격 결격이 아니다(REQ-DPR-004, REQ-DPR-007의 조건부 확장은 B1 verify ref 경로를 위한 것이다 — verify ref push는 현재 어느 워크플로도 트리거하지 않는다).

---

## §2 요구사항 (GEARS)

| ID | 패턴 | 요구사항 |
|---|---|---|
| REQ-DPR-001 | Ubiquitous | The develop required-check set은 develop push 이벤트에서 **항상 보고되는** 컨텍스트만 포함한다 — phase-1 필수 세트(DECIDED, plan.md §A.1 D-1)는 `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)` 4컨텍스트다. 앞 세 개는 ci.yml의 무조건 잡·짝 skip-marker로 항상 보고(research §2.1 — 작성 중 재판독으로 확인), `Analyze (Go) (go)`는 codeql analyze 잡의 push 분기로 항상 보고(§1.3) — main 필수 세트에서 구조적 부재 컨텍스트(`Release PR Multi-OS Gate`)만 제외한 패리티다. |
| REQ-DPR-002 | Unwanted | The develop required-check set은 develop push에서 구조적으로 보고될 수 없는 컨텍스트를 포함하지 않는다 (`shall not`) — `Release PR Multi-OS Gate`(PR 전용 워크플로, develop SHA에는 구조적 부재 → 영구 Pending → 모든 직접 push 봉쇄), test-install 계열(workflow-level paths 필터 → 미발화 시 Pending). |
| REQ-DPR-003 | Capability gate (`Where`) | **Where** 후보 컨텍스트의 push 보고가 동반 워크플로 변경 없이는 불가능하면, the design은 해당 후보를 phase-2(동반 변경 선행 조건부)로 분류한다 — 해당 예: `Race Test`(skip-marker 없는 조건부 잡, ci.yml test-race `if:`). |
| REQ-DPR-004 | Ubiquitous | The 설계는 `Analyze (Go) (go)`가 모든 develop push에서 보고됨을 전제로 한다 — 근거: codeql.yml `analyze` 잡의 `github.event_name == 'push'` 분기 + docs-only 머지 `fa8ff89ba`에서의 success 실측(§1.3). **DECIDED(D-1)**: `Analyze (Go) (go)`는 phase-1 필수 세트에 포함한다 — main 필수 세트 패리티(live 5컨텍스트에 포함)가 지연 비용에 우선한다. 이에 따라 B1 verify 확장은 codeql.yml에도 필수다(REQ-DPR-007). |
| REQ-DPR-005 | Event-driven (`When`) | **When** 통합 창이 필수 상태검사 이력이 없는 신규 merge-commit SHA를 만들면, the window procedure는 그 SHA가 develop에 도달하기 전에 필수 컨텍스트 초록을 확보하는 사전검증 경로(verify-branch seeding, B1)를 제공한다 — merge SHA를 **카드별** verify ref(`verify/<card-id>`, DECIDED D-2)에 push → 초록 대기 → **동일 SHA**를 develop에 push (SHA 범위 상태의 문서화된 경로, research §3.3). |
| REQ-DPR-006 | State-driven (`While`) | **While** B1 경로가 운영 중이 아니면, the operator runbook은 필수 검사가 실제 push 주체(admin 자격)를 게이트하지 않는다는 B2 과도기 의미를 명시적으로 기록한다 — 이 단계의 보호 효과는 비-admin 차단·force-push/삭제 차단·push 후 사후 CI 판정이며, "필수 검사가 창을 사전 게이트한다"고 기술하지 않는다. |
| REQ-DPR-007 | Capability gate (`Where`) | **Where** B1이 사용되면, ci.yml과 codeql.yml **양쪽**의 push 트리거가 카드별 verify ref 패턴(`verify/*`)을 포함하도록 확장된다 — 필수 세트가 `Analyze (Go) (go)`를 포함하므로(DECIDED D-1) codeql도 무조건 대상이다 (현재 양 워크플로 push 트리거는 `[main, develop]`뿐 — verify ref push는 아무것도 트리거하지 않음). |
| REQ-DPR-008 | State-driven (`While`) | **While** B1의 verify-run이 진행 중이면, the window procedure는 통합 락(`moai integration acquire`)을 홀드 상태로 유지하고 develop push 완료 후에 반납한다 — CI 대기 시간만큼 창 점유가 늘어나는 것이 명시된 비용이다. |
| REQ-DPR-009 | Event-detected (`When`) | **When** 필수 컨텍스트가 없는(또는 빨간) SHA가 보호된 develop에 push되면, the runbook은 예상 거부 형태(GH006 `Required status check "..." is expected/failing`)와 회복 경로(동일 SHA를 verify ref에서 초록으로 만든 뒤 재push)를 문서화한다. |
| REQ-DPR-010 | Ubiquitous | The operator runbook은 적용(`gh api -X PUT .../branches/develop/protection`)과 롤백 명령을 정확한 인자와 함께 제공하며, 실행은 운영자 게이트로 남는다 — 본 SPEC의 run-phase는 보호 설정을 변경하지 않는다. |
| REQ-DPR-011 | Ubiquitous | The rollout 순서는 [① 동반 워크플로 변경 → ② 창 절차 갱신 → ③ 보호 적용(마지막)]이며, the runbook은 이 순서 위반(특히 B1 이전의 `enforce_admins: true`)을 창 고착(window-bricking) 오류로 명시하고 롤백 명령으로 즉시 복귀하도록 안내한다. |
| REQ-DPR-012 | Event-detected (`When`) | **When** 보호 적용 직전 7일 내 후보 컨텍스트의 초록 실행 기록이 없으면(GitHub 설정 가능 전제조건, research §3.7), the runbook은 적용을 보류하고 해당 컨텍스트가 develop에서 초록으로 실행될 때까지 기다리도록 지시한다 — 사전 확인 절차(읽기 전용 check-runs 조회)를 포함한다. |
| REQ-DPR-013 | Ubiquitous | The doctrine 문서(`.moai/docs/git-local-workflow-doctrine.md` §23.2 계열)는 develop 보호 설계를 기록하고, 보호 상태 수치는 낡은 스냅샷이 아닌 live API 조회를 SSOT로 삼는다는 고지를 함께 남긴다. |

요구사항 13개 (Tier M 상한 16 이내).

---

## §3 제약조건 (docs-검증 GitHub 의미론, research §3)

1. **직접 push 게이트**: 보호된 브랜치에는 필수 검사 통과 전까지 로컬 변경 push 불가(GH006).
2. **상태의 SHA 범위 스코핑**: 필수 검사는 최종 커밋 SHA에 대해 성공해야 하며, 이전 SHA로 트리거된 검사는 사용되지 않는다. 성공으로 치는 상태는 `success`·`skipped`·`neutral`. — B1의 기계적 근거(다른 브랜치에서 초록이 된 SHA는 보호 브랜치 push에서도 인정).
3. **workflow-level skip → Pending**: paths/branches 필터로 워크플로 자체가 발화하지 않으면 관련 검사는 Pending에 머문다. 잡 단위 조건부 skip은 "Success"로 보고된다. — REQ-DPR-002의 배제 근거.
4. **admin bypass 기본값**: 보호 규칙은 기본적으로 admin에게 적용되지 않는다(`enforce_admins` = "Do not allow bypassing"). — B2의 근거이자 한계.
5. **설정 가능 전제**: 컨텍스트를 필수로 지정하려면 최근 7일 내 리포에서 해당 컨텍스트가 성공한 이력이 있어야 한다.
6. **`strict`(up-to-date)은 PR 전용 의미론** — develop은 `strict: false`를 유지한다(직접 push와 무관).
7. **잡 이름 유일성**: ci.yml의 `Test (ubuntu-latest)` 짝 skip-marker(동일 렌더 이름)는 문서화된 허용 패턴이나, **다른 워크플로**에 같은 이름의 잡이 있으면 안 된다.

---

## §4 설계 — 세 가지 축과 단계적 롤아웃

### (a) 필수 검사 선택

| 단계 | 컨텍스트 | 근거 |
|---|---|---|
| **phase-1 (필수 세트, DECIDED D-1)** | `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)` | 앞 세 개는 ci.yml이 모든 develop push에서 발화하며 무조건 보고(실측, §1.2). `Analyze (Go) (go)`는 codeql analyze 잡의 push 분기로 무조건 보고(§1.3 정정) — main 필수 세트 패리티(구조적 부재 컨텍스트 제외) 근거로 phase-1 포함 확정. |
| **phase-2 후보 (동반 변경 필요)** | `Race Test` | skip-marker 부재로 docs-only push에서 미보고 — marker 짝 추가라는 동반 변경 후에만 후보(REQ-DPR-003). |
| **영구 배제** | `Release PR Multi-OS Gate`, test-install 계열 | REQ-DPR-002. |

`Constitution Check`는 `continue-on-error: true` 잡으로 사실상 항상 초록 — 필수 검사로서의 판별력이 없어 후보에서 제외한다(실측: ci.yml constitution-check 잡).

### (b) 레인 push 통과 경로 — 운영자 결정 축

| 옵션 | 메커니즘 | 평가 |
|---|---|---|
| **B2 — admin bypass (enforce_admins=false)** | 필수 검사는 비-admin에만 구속; admin 자격 push는 우회, CI는 push 후 develop에서 실행 | 동반 변경 0. 단 "필수"가 실제 push 주체를 게이트하지 못함 — 정직한 프레이밍 필수(REQ-DPR-006). |
| **B1 — merge-commit 사전검증 (verify-branch seeding)** | merge SHA → **카드별** `verify/<card-id>` ref push(DECIDED D-2) → CI 초록 대기(`gh run watch --exit-status`, timeout 상한, 모호한 watch 실패 시 commit status API 폴백 — DECIDED D-3; 상태는 SHA에 부착) → 동일 SHA를 develop에 push → **착지 후 해당 verify ref 삭제 청소**(D-2) | 문서화된 SHA 범위 경로로 기계적으로 성립(research §3.3). 동반 변경: ci.yml·codeql.yml push 트리거 `verify/*` 확장(양쪽 — D-1로 Analyze 필수 포함). 비용: 창 점유 +CI 지연, 동일 SHA 2회 실행(ref-scoped concurrency group — 상호 취소 없음, 실측 ci.yml:35-37). 이점: 깨진 병합을 **착지 전**에 발견. 카드별 ref는 통합 락 `--force` 변위·비정상 인터리빙에서도 cancel-in-progress 간섭이 없고 실행 이력의 카드 귀속이 명확하다. |
| B3 — develop PR 필수 | main 모델 복제 | **기각** — 카드 전제(무PR push 모델 공존)와 모순. |
| B4 — PR local-merge 허용 | 카드별 PR 개설 후 로컬 머지 push | **기각** — 모델이 회피하는 PR 객체를 요구, PR 없이는 검증 불가. |

### (c) enforce_admins

- `false` = B2 과도기 의미(모델 무변경, 게이트는 사실상 비-admin용).
- `true` = 운영자 포함 모두에게 진짜 사전 게이트 — **B1이 살아 있는 뒤에만** 선택 가능.
- **권장 단계화**: B2 즉시 → B1 동반 변경 + 창 절차 전환 → `enforce_admins` 최종 결정(각 단계 운영자 게이트). 최종 적용은 런북의 운영자 절차(REQ-DPR-010–011).

### 롤아웃 순서 규율 (위반 시 창 고착)

```
① 워크플로 변경(ci.yml + codeql.yml `verify/*` 확장)
② 창 절차 갱신(B1 사전검증 절차 + 락 홀드 연장)
③ 보호 적용(운영자 gh api PUT — 마지막)
```

③을 ①②보다 먼저, 또는 B1 부재 상태에서 `enforce_admins: true`로 적용하면 통합 창의 모든 push가 GH006으로 거부된다 — 런북이 금지 상태로 명시하고 롤백 명령을 붙인다(REQ-DPR-011).

---

## §5 검증 관점 (run-phase)

- 워크플로 YAML 변경은 `actionlint` + `yq` 트리거 단언으로 기계 검증(변경 전 baseline actionlint exit 0 — 회귀 가드).
- 런북의 모든 읽기 명령(GET)은 실행 가능해야 한다; 쓰기 명령(PUT)은 실행하지 않고 형식·인자만 검증한다.
- verify 패턴은 두 정준형으로 통일한다: 워크플로 트리거 항목 `verify/*`(ci.yml·codeql.yml 동일), 절차·런북의 카드별 표기 `verify/<card-id>`(DECIDED D-2).
- 세부 AC와 RED-now 측정: `acceptance.md`.

---

## Out of Scope

### Out of Scope — 보호 설정 적용 자체
- `gh api -X PUT/PATCH .../branches/develop/protection` 실행. 본 SPEC은 설계와 런북을 delivered하며, 적용은 운영자 게이트다(REQ-DPR-010).

### Out of Scope — PR 기반 대안
- B3(develop PR 필수 전환)과 B4(카드별 PR + 로컬 머지 push)는 카드 전제와 모순되어 설계에서 기각됐고 구현하지 않는다.

### Out of Scope — Go 코드 변경
- `internal/`·`pkg/`·`cmd/`의 어떤 Go 코드도 이 SPEC의 범위가 아니다. 통합 락·세션 도구 변경 없음.

### Out of Scope — main / release/* 보호 변경
- main과 `release/*`의 현행 보호는 그대로 둔다. 본 SPEC은 develop만 다룬다.

### Out of Scope — phase-2 동반 변경의 실행
- `Race Test` skip-marker 짝 추가 등 phase-2 동반 변경은 본 SPEC run-phase에서 실행하지 않는다(설계만 기록).
