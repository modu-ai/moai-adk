---
id: SPEC-GITHUB-FLOW-CI-RESIDUE-001
title: "GitHub Flow 전환 CI 잔여 — 착지 검사 기준 브랜치 전환·일반 PR 다중 OS 게이트·develop 트리거 정리·SPEC 상태 자동 동기화 PR 전환"
version: "0.5.1"
status: in-progress
created: 2026-10-07
updated: 2026-10-07
author: manager-spec (card t1535)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli/worktree, internal/cli, internal/factory, .github/workflows"
lifecycle: spec-anchored
tags: "github-flow, ci-residue, landing-check, multi-os, spec-lint, workflow-triggers, spec-status-sync"
tier: L
related_specs:
  - SPEC-GITHUB-FLOW-DEFAULT-001
  - SPEC-TODO-LANDING-ATTRIBUTION-001
  - SPEC-WORKTREE-BASEREF-001
  - SPEC-MAIN-COMMIT-BAN-001
---

# SPEC-GITHUB-FLOW-CI-RESIDUE-001 — GitHub Flow 전환 CI 잔여

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-07 | manager-spec (card t1535) | Initial plan-phase emission. Tier L. 17 GEARS REQs / 17 binary ACs. 5개 카드 스코프 전부를 워크트리 `.moai/worktrees/t1535` (branch `WT-github-flow-ci-residue`, HEAD `5a9d34fbb` = origin/main tip, PR #1751/t1453 흡수본 포함)에서 재측정했다. 스코프 1의 핵심은 #1751에서 이미 착지했고(§A 재측정 판정), 스코프 5의 todo 절반도 SPEC-TODO-LANDING-ATTRIBUTION-001로 이미 착지했다. 착지 검사 기준 브랜치 전환(sweep·done)이 M1, 워크플로 의미 수리가 M2, develop 트리거 제거가 M3. 모든 결함 진술은 본 회차에서 읽은 file:line을 인용한다. |
| 0.2.0 | 2026-10-07 | manager-spec (card t1535) | 레인 턴종료 게이트(codex) 발견 2건을 본 트리에서 재검증해 흡수한 델타. (1) patch-id 착지 판정(`landing_predicate.go:160-164`, `git patch-id --stable`)의 줄 내부 공백 정규화 — HEAD `5a9d34fbb` 기준 `/tmp` 실증으로 재현: 미반영 들여쓰기 변경·후행 공백 변경 카드 브랜치가 squash 병합 커밋과 동일 patch-id를 내 layer 2가 착지로 판정(B2·B3 모두 `layer2_landed=yes`); 공백 전용 신규 줄 추가는 미재현(줄 수 차이는 감지됨) → REQ-GFC-018. (2) `todo_issuance.go:212` 발급 probe의 리터럴 `develop` 변경 파일 기준 — S5와 동류(develop 부재 저장소: `files=[] ok=false` 미측정, 로컬 develop 존재: 오기준 측정) → REQ-GFC-019. 19 REQ / 19 AC. |
| 0.3.0 | 2026-10-07 | manager-spec (card t1535) | plan-audit iter1(FAIL 0.75, `.moai/reports/t1535/plan-audit-iter1.md`) 수리. **D1** — 스코프 4 (a) 전제 기각: GitHub Actions 기본 `bash -e`(pipefail 없음)는 파이프라인 마지막 명령 `tr`이 grep 실패를 삼켜 :41이 step을 죽이지 않는다(감사 3중 실증 + 본 회차 재관측 — 추출 스크립트 `No SPEC-IDs found in PR title`/exit 0). REQ-GFC-013을 "셸 경화(`set -euo pipefail`)와 명시적 무결과 흡수의 한 세트 착지" 강화로 재서술하고 AC-GFC-013을 regression-guard로 재분류(적색 목록 제외), REQ·design·AC의 제외 경계 계약을 수렴. **D2** — `gh pr create --auto` 무효 플래그 분리(`gh pr create` + `gh pr merge --auto`)·`pull-requests: write` 권한 명시(design D-3, decision-index Q3, AC-014). **D3** — AC-009를 YAML 원소 계수 + "매트릭스 실패→집계 실패" 단정으로 mutant-proof 재설계. **D4** — graph-freshness를 push 1블록(:17-18) 한정으로 정정, PR 이벤트(:9·:15) develop은 스코프 밖 명시, 블록 계수 11→9. **D5** — RED-now 원장을 4요소(명령·원문 stdout·종료 코드·SHA) 셀로 재구성(관측 적근거 7건, 테스트 기반 4건은 §2.1 undecidable disposition). **D6** — AC-015/017에 PyYAML `on:`→True 키 처리 + 9파일 양성 대조 명시. **D7** — AC-006의 예외 출구 제거(0건 단정 수렴). **D8** — ci.yml required 목록 주석 갱신을 M2에 흡수. **D9** — 정보성 기록, 처분 불요. 19 REQ / 19 AC 유지. |
| 0.4.0 | 2026-10-07 | manager-spec (card t1535) | plan-audit iter2(FAIL 0.86, `.moai/reports/t1535/plan-audit-iter2.md`) 수리 — iter3가 Tier L 천장. **D10** — design D-3에 스텝 수준 `env: GH_TOKEN: ${{ github.token }}` 배선 명시(워크플로에 토큰 env 전무 — iter2 grep 0적중 실측, `gh` 무인증 실패 codex exit 4 실측) + AC-GFC-014에 배선 단정 추가. **D11** — AC-GFC-009에 (d) 실효 OS 집합 단정(`matrix.exclude` 반영 후 ⊇ {macos-latest, windows-latest})과 (e) 매트릭스 step의 실제 `go test` 존재 단정을 복원·추가 — 변이 프로브 자가검증으로 두 변이의 탈락을 픽스처 재현(canonical (a)-(e) PASS / exclude 변이 (d) FAIL / no-op 변이 (e) FAIL). **D12** — 원장 014 stdout을 실측 4로 정정(주석 3건 포함). **D13** — 원장 015를 리터럴 단일 명령 + 10행 원문 fenced 보존으로 교체. **D14** — AC-017을 강화: `pull_request`가 존재하면 그 branches에 main 포함 또는 무필터까지 단정. 19 REQ / 19 AC 유지. |
| 0.5.0 | 2026-10-07 | manager-spec (card t1535) | **분할 재계획** — plan-audit iter3가 Tier L 천장(3/3) 도달(FAIL 0.89, blocking 1건 D16; `.moai/reports/t1535/plan-audit-iter3.md` §Final Escalation). 감사자 분할 제안 채택: 스코프 4(SPEC 상태 자동 동기화 — REQ-GFC-013·014 + design D-3 + AC-GFC-013·014 + decision-index Q3)를 본 SPEC에서 제거하고 별도 카드로 분리. 분할 근거(iter3 §Final Escalation 인용): "무인 자동 병합 계약 자체가 이 저장소의 보호 설정과 양립하지 않는 구조 문제" — iter1 D2(무효 플래그)→iter2 D10(인증 배선)→iter3 D16(CI 미발화·auto-merge 정체)으로 매 수리가 다음 계층을 드러낸 3계층 정체이며, 잔여 처분은 문안 수리가 아니라 운영 정책 선택((a) 라벨+수동 병합 (b) GitHub App 토큰 (c) auto-merge 유지+운영자 승인 명시)이라 천장 정지가 옳은 판정. 본체(스코프 1·2·3·5 + 흡수 REQ)는 3회차 내내 감사-안정이라 즉시 run 자격. 남은 REQ·AC는 번호 재배열만(015-019 → 013-017), 내용 무수정. 스코프 4는 decision-index Q3에 hold로 기록 — 운영자 3택 결정 대기. **17 REQ / 17 AC** (19 − 2 제거 = 17 — iter3 분할 제안서의 "15건"은 산술 오판, 기계 계수 grep 17·17·17 일치로 정정). |
| 0.5.1 | 2026-10-07 | manager-spec (card t1535) | 분할 검증(.moai/reports/t1535/split-verification.md) D20 토큰 수리 — 재배열 잔여 참조 4곳 치환: (i) spec.md §A.2 스코프 3 행 "REQ-GFC-015 스코프 밖"→REQ-GFC-013 (ii) design D-4 "REQ-GFC-017 사전 확인 표"→REQ-GFC-015 (iii) design D-4 "REQ-GFC-015는 push 트리거 한정"→REQ-GFC-013 (iv) research.md §1 스코프 3 "REQ-GFC-015 스코프 밖"→REQ-GFC-013. 선택 흡수 2건: D21 — plan §H AC 범위 001..015→001..017, D22 — research 스코프 4 원장·progress split_history의 구번호에 "분할 전 REQ-GFC-013(스코프 4)" 구분 접미어(재배열 뒤 REQ-GFC-013은 push 트리거 요구를 가리키므로). 17 REQ / 17 AC 불변. + 같은 라운드에서 운영자 D16 결정 하달: (a) 라벨 부착+수동 병합 확정 — decision-index Q3 Operator verdict에 기록, 스코프 4 재설계는 후속 카드 t1557로 이관. |

## §A Context

### A.1 카드와 전환의 관계

카드 t1535는 GitHub Flow 전환(SPEC-GITHUB-FLOW-DEFAULT-001, in-progress)의 CI 잔여를 다룬다. 본 SPEC의 베이스 HEAD `5a9d34fbb`는 PR #1751(card t1453, "GitHub Flow cutover SPEC — strategy absorption")을 이미 포함하므로, 카드 본문의 관측(2026-10-06) 중 일부는 그 병합으로 수리됐을 수 있다. 그래서 본 SPEC은 §A.2의 재측정 판정을 전제로 쓰였다 — 전제가 사라진 항목은 요구사항이 되지 않는다.

cutover SPEC plan §H(후속 인벤토리)는 본 카드의 스코프 3을 명시적으로 예고하고 있다: "워크플로 push 트리거의 `develop` 제거 — develop 삭제 뒤 정리 카드(design D-13 단계 4)". 또 §I D-22는 "필수 체크 `Release PR Multi-OS Gate` 의 처분 — 필수 체크에서 뺀다(운영자가 수행), 태그 직전 `workflow_dispatch`"를 결정했다. 본 SPEC은 D-22와 정면으로 닿지 않는다 — release-pr-multi-os.yml을 재편하지 않고 일반 PR 쪽 게이트를 보강한다(§C).

### A.2 재측정 판정 (본 회차 실측, HEAD `5a9d34fbb`)

| 스코프 | 카드 전제 | 재측정 판정 | 근거 (본 회차 읽은 위치) |
|---|---|---|---|
| 1 | spec-lint.yml:70/92의 origin/develop 필수 fetch + release HEAD=origin/develop 검사 | **이미 착지 (핵심) + 잔여(죽은 코드)** | push 트리거 main 전용(spec-lint.yml:17-19), fetch main 전용이며 우회 없음(:70-71), release/* 완화의 기준이 `origin/main`(:107). 잔여: 도달 불가능한 develop-push elif 분기(:121-139) — 주석 스스로 "unreachable text … awaiting the release discipline's own re-design"라고 기록 |
| 2 | 일반 feature→main PR에 필수 다중 OS 런타임 검증 부재 | **유지 (holds)** | release-pr-multi-os.yml:42 `if: startsWith(github.head_ref, 'release/') \|\| workflow_dispatch` — release/* PR과 수동 실행만 매트릭스를 돌린다. ci.yml:126 `os: [ubuntu-latest]` — 필수 test 잡은 ubuntu 단일. race는 test-race-1/2(ci.yml:289+)로 분리된 advisory·비필수 잡(:266-280 주석: required 목록은 "Test (ubuntu-latest), Lint, Build (linux/amd64), Analyze (Go), Release PR Multi-OS Gate"). cross-compile build 잡은 매 PR 돌지만 컴파일만 검증(:110-116 주석) |
| 3 | develop push 트리거 ~10개 워크플로 잔존 | **유지 (holds) — 9개 파일·9개 push 블록** | ci.yml:18, codeql.yml:5, graph-freshness.yml:17-18(push 1블록 — :9 pull_request·:15 pull_request_target은 PR 이벤트로 REQ-GFC-013 스코프 밖), docs-i18n-check.yml:33, test-install.yml:5, workflow-parse-guard.yaml:14, judgment-first-consistency.yaml:34, lsel-leak-guard.yaml:11, template-neutrality-check.yaml:31 (spec-lint.yml은 이미 main 전용 — 스코프 1 수리분) |
| 4 | spec-status-auto-sync.yml:123 직접 main push·:41 제목 grep 실패 종료 | **부분 유지 — (a) 기각(plan-audit iter1), (b) 유지 → v0.5.0 분할(hold)** | (a) :41 제목 grep — **기각**: GitHub Actions 기본 `bash -e`(pipefail 없음)는 파이프라인 마지막 명령 `tr`이 grep의 exit 1을 삼켜 :42의 `-z` 가드가 도달해 종료 0이다(감사 iter1 3중 실증 — bash 실험·추출 스크립트 실행·`gh run list` 10건 중 9 success에 SPEC 없는 제목 포함 — + 본 회차 재관측: 추출 스크립트 `No SPEC-IDs found in PR title`/exit 0). 무방비 grep은 `set -euo pipefail` 경화 도입 시에만 폭발하는 **잠재 결함**이다. (b) :121-124 `git push origin main` — 유지: 이제 보호된 기본 브랜치에의 PR 없는 직접 push. **스코프 4 전체는 iter3 D16(무인 병합 계약과 보호 설정의 구조적 불양립)으로 v0.5.0 분할 — decision-index Q3 hold, 별도 카드로 재계획** |
| 5 | sweep·todo done의 착지 검사가 origin/develop 하드와이어 | **형태 변화 — worktree 절반은 유지, todo 절반은 이미 착지** | todo 표면은 이미 3단 해석 사슬(internal/factory/prlink_landedref.go:63-84 — ① `git_strategy.worktree_base_branch` → ② `refs/remotes/origin/HEAD` → ③ `DefaultLandedRef = "origin/main"` (prlink_landed.go:48))을 쓰고, 본 저장소 설정은 `worktree_base_branch: main`(git-strategy.yaml:5)이라 origin/main으로 답한다. 반면 `moai worktree sweep`은 `sweepDefaultBase`(sweep.go:245-254)가 `config.LoadGitFlowIntegrationConfig`의 `IntegrationTarget`을 따르고, 그 해석은 `manual` 모드 + `workflow: git-flow` + `develop_branch: develop`(loader_integration_branch.go:121-131)라 **origin/develop**으로 답한다 — develop 원격이 삭제되면 `git fetch origin develop`가 실패해 전 트리가 PRESERVE된다(sweep.go:201-206의 3-way 계약). `moai worktree done`의 `landingBase`(done.go:287-298)도 같은 사슬이라 ORIGIN_LANDING_UNCONFIRMED로 일괄 거부된다(done.go:349-357). 잔여: todo_autodone.go:2-4·:130의 도움말·주석이 여전히 `git fetch origin develop + git rev-parse origin/develop`을 이름붙인다(문서 잔여) |

리더의 10-06 실측("sweep이 161트리 전부 PRESERVE, done landing=unknown")은 본 베이스의 코드 경로와 정합한다 — 원인은 문자 하드와이어가 아니라 **설정 해석이 아직 develop을 가리킨다**는 점으로 정밀화된다. 그래서 M1의 수리는 (a) sweep·done의 기준 해석을 todo가 이미 쓰는 사슬과 통일하는 코드 수리와 (b) 문서 잔여 수리로 구성된다.

### A.3 방법론

`quality.yaml:4`의 `constitution.development_mode: tdd` — M1의 Go 수리는 RED-GREEN-REFACTOR로 진행한다. CI-YAML 스코프는 저장소에 actionlint·yamllint가 없음을 확인했으므로(전체 grep 1건 — docs-i18n-check.yml:137의 주석 언급뿐), 저장소 관례를 따라 워크플로 린터를 새로 도입하지 않고 YAML 파싱 + 행위 단정(grep·스텁 실행)으로 검증한다.

## §B Requirements (GEARS)

### B.1 착지 검사 기준 브랜치 전환 (M1 — 카드 스코프 5, 파이프라인 차단 긴급)

- REQ-GFC-001: **Where** 프로젝트가 `git_strategy.worktree_base_branch`를 설정한 경우, the `moai worktree sweep`·`moai worktree done`의 착지 검사 shall 그 설정값에 대응하는 `origin/<branch>`를 착지 판정의 기준 브랜치로 삼는다.
- REQ-GFC-002: **While** 설정값이 비어 있거나 상위 단계가 답하지 않는 경우, the 착지 기준 해석 shall 문서화된 하위 단계 순서(`refs/remotes/origin/HEAD` symref 읽기 → 컴파일 기본값 `origin/main`)로 내려가며, 어느 단계가 답했는지의 출처를 기록·공개한다 — SPEC-TODO-LANDING-ATTRIBUTION-001 REQ-TLA-011의 공개 원칙을 준용한다.
- REQ-GFC-003: **When** 기준 원격의 `git fetch` 또는 merge-base 판정이 0·1 이외의 종료로 끝나는 경우, the sweep shall 해당 트리를 PRESERVE한다 — 현행 3-way 계약(sweep.go:201-206)을 약화하지 않는다.
- REQ-GFC-004: **When** origin에서 develop ref가 존재하지 않는 경우, the sweep·done의 착지 검사 shall 해석된 기준 브랜치(본 저장소 설정에서는 origin/main)로 착지를 판정한다 — fetch 대상이 삭제된 ref라서 착지한 트리가 전부 PRESERVE되거나 done이 일괄 거부되는 상태가 남지 않는다.
- REQ-GFC-005: The todo 표면과 worktree 표면의 착지 질의 shall 하나의 해석 사슬을 공유한다 — 도움말 텍스트, 거부 사유, 질의 자체가 서로 다른 ref를 이름붙이지 않는다 (todo.go:139 "single place" 원칙의 worktree 확장).
- REQ-GFC-006: **While** GitHub Flow 전환이 유효한 동안, the 착지 검사의 사용자 대상 텍스트(도움말·주석·거부 사유) shall 존재하지 않는 `origin/develop` 기준을 이름붙이지 않는다 (todo_autodone.go:2-4·:130의 잔여 포함).
- REQ-GFC-016: **When** patch-id 비교가 통합 ref 위에서 카드의 누적 패치와 일치하는 후보 커밋을 발견하는 경우, the patch-id landing layer shall 후보 커밋과 카드 tip의 트리 내용이 공백까지 완전히 일치할 때만 착지로 선언한다 — `git patch-id --stable`의 줄 내부 공백 정규화는 미반영 공백성 커밋(들여쓰기·후행 공백 변경)을 착지로 위장하며, 이 판정으로 워크트리가 삭제되어 그 커밋이 유일본째 유실되어서는 안 된다 (landing_predicate.go:160-164, 재현 근거 research.md §3).
- REQ-GFC-017: **Where** 발급 안내 probe가 카드 브랜치의 변경 파일을 측정하는 경우, the lane files probe shall 변경 파일 기준을 프로젝트의 통합 베이스 해석(설정 우선 사슬 — REQ-GFC-001과 동일한 해석)에서 가져오고, 리터럴 `develop`을 기준 연산의 피연산자로 삼지 않는다 (todo_issuance.go:212 — develop 부재 저장소에서 `files=[] ok=false` 미측정, 로컬 develop 존재 시 오기준 측정).

### B.2 spec-lint 잔여 정리 (M2 — 카드 스코프 1의 잔여)

- REQ-GFC-007: **When** main 전용 push 트리거 아래 도달 불가능한 develop-push 정책 분기가 남아 있는 경우, the spec-lint policy script shall 그 분기를 제거해 도달 가능한 정책 분기만 남긴다 (spec-lint.yml:121-139).
- REQ-GFC-008: The spec-lint gate shall not 통합 기준 fetch(`git fetch origin main`, spec-lint.yml:70-71)의 실패를 무시하거나 우회할 수 있는 구성을 가진다 — 이미 착지한 이 불변식을 수리 과정에서 되살려서는 안 된다.

### B.3 일반 PR 다중 OS 집계 게이트 (M2 — 카드 스코프 2)

- REQ-GFC-009: **When** feature→main PR이 플랫폼 종속 변경으로 분류되는 경우, the CI workflow shall macOS 러너와 Windows 러너에서 해당 변경의 영향을 받는 패키지의 테스트(`go test`)를 실행하고 그 결과를 단일 집계 체크로 합산한다.
- REQ-GFC-010: The 집계 게이트 shall 브랜치 보호의 필수 체크로 등록 가능한 안정적인 단일 체크 이름을 게시한다 (등록 행위 자체는 운영자 콘솔 활 — decision-index Q2).
- REQ-GFC-011: **When** 변경 분류기가 플랫폼 종속 변경이 없다고 판정하는 경우, the 집계 게이트 shall 매트릭스를 건너뛰고 성공으로 보고한다; **When** 분류기 자체가 판정에 실패하는 경우, the 집계 게이트 shall 매트릭스를 실행한다 — release-pr-multi-os.yml:44-60의 실패-안전 관례(필터 오류 시 매트릭스 실행)를 준용한다.
- REQ-GFC-012: **While** cross-compile 빌드가 통과한 동안, the CI shall 그 결과를 어느 OS에서의 런타임 검증으로도 기록하거나 보고하지 않는다 — 컴파일 성공은 테스트 통과를 증명하지 않는다 (카드 원문: "cross-compile은 런타임 검증을 대체하지 않음").

### B.4 develop push 트리거 제거 (M3 — 카드 스코프 3)

- REQ-GFC-013: **While** GitHub Flow 전환이 유효한 동안, `.github/workflows/` 아래의 그 어떤 workflow도 push 트리거의 branches에 develop을 이름붙이지 않는다 (§A.2의 9개 파일).
- REQ-GFC-014: **When** M1·M2의 의미 수리가 아직 착지하지 않은 경우, the develop 트리거 제거 shall 착지 순서상 선행 수리 뒤에 실행된다 — 삭제가 수리보다 앞서 관련 워크플로를 침묵시키지 않는다 (카드 원문: "의미 수리 후 제거").
- REQ-GFC-015: **When** develop push 트리거 제거로 어떤 검사가 더 이상 develop push에서 실행되지 않게 되는 경우, the 제거 shall 각 워크플로가 main 대상 경로(push to main 또는 pull_request)에서 동일 검사를 여전히 실행함을 검증한 뒤에 진행된다 — 삭제가 검사 공백을 만들지 않는다 (docs-i18n-check.yml:27-33의 develop 유지 사유는 "레인이 develop에 직접 병합"이었으므로 전환 뒤 소멸하지만, main 경로 커버리지는 개별 확인이 필요하다).

## §C Exclusions

### Out of Scope — origin 기본 브랜치 전환과 develop 원격 삭제의 실행

- cutover SPEC plan §J D2("origin 기본 브랜치 develop→main + main 브랜치 보호")와 develop 원격 삭제·아카이브(D4)는 운영자 콘솔 활이다. 본 SPEC의 M1(REQ-GFC-004)이 그 삭제의 **선행 차단 조건**을 충족시키는 것까지가 이 SPEC의 책임이고, 삭제 실행은 아니다.

### Out of Scope — git-strategy.yaml의 flow 라벨 전면 전환

- 본 저장소 `manual.workflow: git-flow` 라벨의 github-flow 전환은 cutover SPEC M4-M5(CUTOVER-TIME)·운영자 D2 소관이다. 본 SPEC은 착지 기준 해석이 `worktree_base_branch: main`(git-strategy.yaml:5)을 우선하도록 만들어 라벨과 무관하게 main으로 답하게 한다. `doctor_git_strategy_workflow.go:97`의 wtBase≠target 진단과의 정합은 design.md D-1에서 다룬다.

### Out of Scope — release-pr-multi-os.yml의 재편

- 릴리스 시점 3-OS 매트릭스의 처분은 cutover §I D-22(필수 체크 제외·태그 직전 `workflow_dispatch`)가 결정했다. 본 SPEC은 그 워크플로를 수정하지 않고 일반 PR 쪽 집계 게이트를 보강한다. "전체 3-OS 검증은 릴리스 승인 전 별도 보장"은 이 워크플로 + D-22의 dispatch로 유지된다.

### Out of Scope — Todo 착지 해석 사슬의 재설계

- 3단 해석 사슬과 공개 규약은 SPEC-TODO-LANDING-ATTRIBUTION-001(completed)이 소유한다. 본 SPEC은 그 사슬을 worktree 표면에 재사용할 뿐 사슬 자체를 바꾸지 않는다.

### Out of Scope — workflow 린터(actionlint·yamllint) 신규 도입

- 저장소에 기존 사용례가 없으므로 관례 미러링 원칙에 따라 도입하지 않는다. 검증은 YAML 파싱 + 행위 단정으로 충분하다(§A.3).

### Out of Scope — SPEC 상태 자동 동기화 전달 경로 (스코프 4 — hold)

- iter3 D16(무인 auto-merge 계약과 이 저장소의 브랜치 보호 설정 간 구조적 불양립 — plan-audit iter3 §Final Escalation)으로 v0.5.0 분할했다. 전달 정책 3택 — (a) 라벨+수동 병합 (b) GitHub App 토큰 (c) auto-merge 유지+운영자 승인 명시 — 은 운영자 결정 사항이며 decision-index Q3에 hold로 기록돼 있다. 결정 후 별도 카드로 재계획한다. 측정 원장(research.md §1 스코프 4)은 hold 참조용으로 보존한다.

### Out of Scope — cutover §H 후속 인벤토리의 나머지 항목

- `internal/spec/lint_movingref.go` 정규식의 develop 대안, `release-drafter-cleanup.yml`·`auto-merge.yml`의 `release/*` 키, `push_develop` 와이어 식별자 개명 등 — 후속 카드 소관(각각 §H에 이미 기록돼 있다).
