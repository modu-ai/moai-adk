# SPEC-GITHUB-FLOW-CI-RESIDUE-001 — Research (재측정 원장)

측정 조건: 워크트리 `.moai/worktrees/t1535`, branch `WT-github-flow-ci-residue`, HEAD `5a9d34fbb` (origin/main tip, PR #1751 포함). 측정일 2026-10-07. 아래 모든 행은 본 회차의 Read/Grep/Bash 관측이다 — 카드 본문(2026-10-06 관측)과 다른 판정은 그 차이를 명시했다.

## 1. 스코프별 재측정 원장

### 스코프 1 — spec-lint.yml (판정: 핵심 이미 착지, 잔여 죽은 코드)

| 관측 | 위치 | 내용 |
|---|---|---|
| push 트리거 | spec-lint.yml:17-19 | `push: branches: [main]` — develop 부재. :13-16 주석이 cutover 사유 기록 |
| fetch | spec-lint.yml:70-71 | `git fetch origin main:refs/remotes/origin/main` — main 전용, 우회 구성 없음. :62-69 주석: develop ref 지명 시 exit 128로 전체 fetch 실패를 명시 |
| release 완화 기준 | spec-lint.yml:107 | `base_sha=$(git rev-parse origin/main)` — origin/develop 아님 |
| **잔여** | spec-lint.yml:121-139 | `elif [[ "$EVENT_NAME" == "push" && "$REF_NAME" == "develop" ]]` — main 전용 트리거 아래 도달 불가. :86-89 주석: "unreachable text … awaiting the release discipline's own re-design (the cutover card's call)" |
| 카드와의 차이 | — | 카드의 ":70/92 origin/develop 필수 fetch" 전제는 #1751로 소멸. REQ-GFC-007은 잔여만 다룬다 |

### 스코프 2 — 다중 OS (판정: 유지)

| 관측 | 위치 | 내용 |
|---|---|---|
| release 필터 | release-pr-multi-os.yml:42 | `if: startsWith(github.head_ref, 'release/') \|\| github.event_name == 'workflow_dispatch'` — 일반 PR은 매트릭스 미실행 |
| docs-only 스킵 관례 | release-pr-multi-os.yml:44-60 | dorny/paths-filter + `continue-on-error: true` + `!= 'false'` — 필터 오류 시 매트릭스 실행(실패-안전). D-2가 이 관례를 재사용 |
| 필수 test 매트릭스 | ci.yml:126 | `os: [ubuntu-latest]` |
| race 성격 | ci.yml:266-280 | "separate advisory, non-required job" — required 목록(:271-274): Test (ubuntu-latest), Lint, Build (linux/amd64), Analyze (Go), Release PR Multi-OS Gate |
| cross-compile | ci.yml:110-116 | "Cross-platform runtime coverage … moved to release time" — build 잡은 컴파일·vet만 |
| 카드와의 차이 | — | 없음 — 카드 라인 번호(:42·:126) 그대로 유효 |

### 스코프 3 — develop push 트리거 (판정: 유지 9개)

`grep -rn "develop" .github/workflows/` 관측(주석 제외 판독):

- ci.yml:18 `branches: [main, develop]` (push)
- codeql.yml:5 `branches: [main, develop]` (push)
- graph-freshness.yml:17-18 (push 1블록 — :9 `pull_request`·:15 `pull_request_target`은 PR 이벤트 블록으로 REQ-GFC-013 스코프 밖; v0.3.0에서 "push 3블록"이라 쓴 초기 판독을 감사 iter1 D4가 정정했다 — 본 파일 :5-18의 이벤트 블록 3개 중 push는 마지막 하나뿐)
- docs-i18n-check.yml:33 `- develop` (:27-28 주석 — "lane work merges into develop directly" 사유)
- test-install.yml:5 (push)
- workflow-parse-guard.yaml:14 (push)
- judgment-first-consistency.yaml:34 (push)
- lsel-leak-guard.yaml:11 (push)
- template-neutrality-check.yaml:31 (push)

spec-lint.yml은 이미 main 전용 — 수리 대상은 9개 파일·9개 push 블록. 카드의 "~10개"와 합치.

### 스코프 4 — spec-status-auto-sync.yml (판정: 부분 유지 — (a) 기각(iter1), (b) 유지 → v0.5.0 분할 hold, decision-index Q3)

| 관측 | 위치 | 내용 |
|---|---|---|
| 제목 grep | :41 | `SPEC_IDS=$(echo "$PR_TITLE" | grep -oE 'SPEC-[A-Z0-9-]+-[0-9]+' | tr '\n' ' ')` — **v0.3.0 기각(plan-audit iter1 D1, 본 회차 재관측)**: GitHub Actions 기본 `bash -e`(pipefail 없음)에서 파이프라인 종료는 마지막 명령 `tr`의 것이라 grep의 exit 1이 승격되지 않고 :42의 `-z` 가드가 도달한다. 본 회차 관측 — 워크플로 run 블록 추출 후 `PR_TITLE="chore: no spec id here" bash <extracted>` → stdout `No SPEC-IDs found in PR title`, exit 0. 생산 관측도 동일(`gh run list` 최근 10건 중 9 success, SPEC 없는 제목 포함). 무방비 grep은 `set -euo pipefail` 경화 시에만 폭발하는 잠재 결함 — 경화와 명시적 흡수는 한 세트(분할 전 REQ-GFC-013 — 스코프 4, v0.5.0 제거) |
| 직접 push | :121-124 | `git push origin main` — **살아있는 결함 (b)** |
| 발화 조건 | :26 | `if: github.event.pull_request.merged == true` — SPEC 없는 제목의 "모든" 병합 PR이 이 경로를 지난다 |
| 재사용 가능 요소 | :77-82·:92·:100-118 | loop-prevention·staged-diff 가드·재시도 루프 — PR 전환 후에도 유효 |

### 스코프 5 — 착지 검사 (판정: 형태 변화)

**todo 절반 — 이미 착지:**

- 해석 사슬: internal/factory/prlink_landedref.go:63-84 — ① `config.LoadWorktreeBaseBranch` → ② `git symbolic-ref refs/remotes/origin/HEAD` (읽기 전용) → ③ `DefaultLandedRef`. 단계 열거형: :36-46.
- 기본값: internal/factory/prlink_landed.go:48 `const DefaultLandedRef = "origin/main"`.
- 소유 SPEC: SPEC-TODO-LANDING-ATTRIBUTION-001 — **status: completed** (frontmatter 본 회차 확인).
- 소비자: todo.go:139-160(todoLandedRef·Resolved), todo_landed.go:168, todo_pr.go:236, todo_triage.go:498.
- 본 저장소 설정: git-strategy.yaml:5 `worktree_base_branch: main` — **TRACKED, HEAD 커밋본 확인**(`git show HEAD:…` 5행). 즉 todo done은 지금 origin/main으로 답한다.
- **문서 잔여**: todo_autodone.go:2-4·:130 — "git fetch origin develop + git rev-parse origin/develop". todo_triage.go:352 주석도 origin/develop 언급.
- 리더 관측("done landing=unknown")과의 차이: 10-06 실측 바이너리는 이 사슬 착지 이전 빌드로 추정(§J.2 환경 특이 — 설치 빌드 rc.24 지적). 코드 기준으로는 이미 수리됨.

**worktree 절반 — 유지 (파이프라인 차단):**

- sweep: internal/cli/worktree/sweep.go:245-254 `sweepDefaultBase` → `config.LoadGitFlowIntegrationConfig(root).IntegrationTarget`; 비면 오류 + `--base` 안내(:251-252). 3-way 계약(fail 시 PRESERVE): :201-206 Long 서술.
- done: internal/cli/worktree/done.go:287-298 `landingBase` → 같은 설정 사슬, provenance 반환; 거부 경로 :349-381 (ORIGIN_LANDING_UNCONFIRMED — fetch 실패 fail-closed).
- 설정 해석: internal/config/loader_integration_branch.go:104-131 — `Manual && GitFlowWorkflow`일 때 `IntegrationTarget = DevelopBranch`; 본 저장소 git-strategy.yaml `mode: manual` + `manual.workflow: git-flow` + `develop_branch: develop` → **develop**. develop 원격 삭제 시 `git fetch origin develop` 실패 → sweep 전 트리 PRESERVE·done 일괄 거부. 리더의 161트리 관측과 정합.
- 인접 진단: internal/cli/doctor_git_strategy_workflow.go:97 — wtBase≠target 드리프트 진단 존재 (design D-1.3).

## 2. 방법론·관례 관측

- `quality.yaml:4` — `development_mode: tdd` → M1은 TDD.
- 워크플로 린터: actionlint·yamllint 사용례 없음(전체 grep — docs-i18n-check.yml:137의 주석 언급뿐). 검증 편성 = python3 YAML 파싱 + 행위 단정.
- 템플릿 배포물: internal/template/templates/.github/ = labels.yml, branch-protection.json.gtmpl, workflows/label-sync.yml, actions/detect-language/action.yml — 본 카드가 만지는 dev 워크플로는 템플릿 불포함. `make build` 불요.
- cutover SPEC plan §H — "워크플로 push 트리거의 develop 제거 — develop 삭제 뒤 정리 카드(design D-13 단계 4)" 예고. §I D-22 — Release PR Multi-OS Gate 필수 체크 제외(운영자)·태그 직전 workflow_dispatch.
- SPEC ID 충돌: `.moai/specs/` 1057개 디렉터리 중 `SPEC-GITHUB-FLOW-CI-RESIDUE-001` 부재 (GITHUB-FLOW-DEFAULT-001·GITHUB-WORKFLOW만 존재). ID 정규식 사전 검사 PASS (실행 출력: `PASS`).
- decision 게이트: `.moai/config/sections/interview.yaml:6` `decision_gate: on` → decision-index.md 산출.

## 3. v0.2.0 델타 — 레인 턴종료 게이트(codex) 발견 2건의 재검증 원장

HEAD `5a9d34fbb`에서 본 회차(2026-10-07) 재검증. 원 발견: `.moai/reports/t1535/turn-gate-findings-20261007.md` #1·#10 (레인 미검증 주장 → 본 절이 직접 검증).

### 3.1 주장 (1) — patch-id 착지 판정의 공백 정규화 → **재현 (형태 정밀화)**

- 코드 관측: internal/cli/worktree/landing_predicate.go:90-104(`landingPatchIDs` — `git patch-id --stable`), :137-165(`LandedByPatchID` layer 2 — 카드 누적 diff의 patch-id vs 통합 ref 커밋들의 patch-id, :160-164 동일성 판정). 파일 머리 주석(:3-4): SPEC-GITHUB-FLOW-DEFAULT-001 REQ-GFD-002 / design D-5 소유.
- 실증 절차: `/tmp/t1535-patchid/` 스크래치 저장소에서 squash 병합 + 미반영 공백 커밋 재현 후 layer 2의 정확한 git 수열(diff 플래그는 landingDiffFlags :109와 동일 — `--no-ext-diff --no-color --no-renames --binary`)로 patch-id 비교.
- 결과:
  - B1(공백 전용 **신규 줄** 추가): card `684bc2d3…` ≠ ref `7688e792…` → `layer2_landed=no` — **미재현** (줄 수·추가 줄 차이는 patch-id가 감지).
  - B2(**들여쓰기** 변경, `world`→`   world`): card = `7688e79215b0cd5f2f46c96516bae3a0987d703c` = ref squash → **`layer2_landed=yes` 재현**.
  - B3(**후행 공백** 변경, `world`→`world  `): 동일 id → **`layer2_landed=yes` 재현**.
- 판정: 결함은 "공백 전반"이 아니라 **줄 내부 공백 정규화**로 정밀화된다 — 미반영 들여쓰기·후행 공백 커밋(gofmt 성 재정렬 포함)을 얹은 카드 브랜치가 squash 병합 뒤 착지로 판정되어 워크트리 삭제가 허용되고, 그 커밋은 트리가 폐기되며 유실된다. layer 3(PR head SHA 정확 비교, :210)과 layer 1(ancestor)은 이 결함이 없다 — layer 2 한정.
- 흡수: REQ-GFC-016 / AC-GFC-016 / M1. 수리 방향은 행위 수준으로만 고정(후보 발견은 patch-id, 착지 선언 전 후보 커밋과 카드 tip의 트리 내용 공백 포함 일치 확인) — 기계 선정은 런페이즈 design 재량.

### 3.2 주장 (2) — 발급 probe의 리터럴 develop 기준 → **재현 (S5 동류)**

- 코드 관측: internal/cli/todo_issuance.go:212 `base := issuanceGitOneLine(root, "merge-base", "develop", branch)` (주석 :204도 "merge-base with develop"). 발화면: `todoLaneFilesProbe = productionLaneFilesProbe`(:27 — 런페이즈 스텁 지점 존재). :213-215 — base가 비면 `nil, false` 반환 → 프레젠테이션은 "unmeasured"로 보고.
- 저장소 관측: 본 워크트리 저장소에 로컬 `develop`과 `main`이 **공존**(`git branch --list develop main` — 둘 다 존재). 따라서 두 실패 형태가 모두 성립한다: (a) develop 부재 저장소(main 순수) — merge-base 실패 → `files=[] ok=false` 미측정(codex 주장 그대로); (b) develop 존재·스테일 저장소(본 저장소) — merge-base(develop, WT-브랜치)가 main 기반 카드의 진짜 분기점을 놓쳐 오기준 변경 파일 목록.
- 판정: S5의 설정 해석 문제와 **동류**이지만 별도 호출점이라 REQ-GFC-001/005의 사슬 통일(sweep·done 진입점)만으로는 덮이지 않는다 — 착지 질의가 아니라 변경 파일 측정 기준이므로 소비자 명시가 필요.
- 흡수: REQ-GFC-017 / AC-GFC-017 / M1. 수리는 REQ-GFC-001과 동일한 설정 우선 해석을 이 호출점에도 적용.
