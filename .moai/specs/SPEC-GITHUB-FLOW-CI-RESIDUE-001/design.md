# SPEC-GITHUB-FLOW-CI-RESIDUE-001 — Design

## D-1 착지 기준 해석 사슬의 통일 (M1)

**결정**: `sweepDefaultBase`(sweep.go:245-254)와 `landingBase`(done.go:287-298)의 기준 해석을 `factory.LandedRefForWithLevel`(internal/factory/prlink_landedref.go:63-84) 사슬로 교체한다.

사슬 순서: ① `git_strategy.worktree_base_branch` 설정값 → ② `git symbolic-ref refs/remotes/origin/HEAD` 읽기(쓰기 형태 없음 — REQ-TLA-012 준용) → ③ `DefaultLandedRef = "origin/main"`(prlink_landed.go:48).

**대안 기각**:
- (a) 설정만 고치기(本 저장소 git-strategy.yaml의 workflow 라벨 전환) — dev-local 설정이라 제품 결함을 고치지 못하고, 라벨 전환 자체는 cutover SPEC D2 소관이다.
- (b) main 하드와이어 — 카드 지시("prefer honoring the configured git-strategy base branch … falling back to main") 위반이고 git-flow 사용자의 현행 동작을 깨뜨린다.

**호환 성질**: git-flow 프로젝트는 `worktree_base_branch: develop` 설정(1단) 또는 origin/HEAD=develop(2단)으로 여전히 origin/develop 기준을 얻는다 — 사슬 교체는 GitHub Flow 전환 저장소의 답만 바꾼다.

**파생 수리**:
- D-1.1 `sweepDefaultBase`의 "no target → error + --base 안내" 분기 소멸 — 사슬은 항상 답한다. `--base` 플래그 우선순위는 유지(sweep.go:261-266). Long 도움말(sweep.go:201-216)의 "origin/develop under git-flow" 서술을 사슬 서술로 갱신.
- D-1.2 `landingBase`의 provenance(done.go:287-298 반환값, done.go:379 거부 사유 표기)를 사슬 단계 어휘로 교체 — `LandedRefConfigured|LandedRefOriginHEAD|LandedRefDefault` + `todoRefLevelSource`(todo.go:155-163) 급의 공개 문구.
- D-1.3 `doctor_git_strategy_workflow.go:97`의 wtBase≠target 진단 — 사슬 통합 뒤 이 진단은 "설정과 flow 해석의 불일치 정보성 경고"로 남는 게 적절하다(착지 기준은 더 이상 target을 따르지 않음). 진단 문구에 그 사실을 반영하고, 행동 변화(게이트화 등)는 하지 않는다.

## D-2 다중 OS 집계 게이트의 형태 (M2)

**결정**: 신규 워크플로 파일(예: `.github/workflows/pr-multi-os-gate.yml`)로 분리한다. ci.yml에 매트릭스를 추가하는 대안보다 되돌리기 쉽고(단일 revert), release-pr-multi-os.yml과의 구분이 파일 경계로 명확해진다.

구조(기존 관례 재사용):
1. `detect`: dorny/paths-filter(고정 SHA — release-pr-multi-os.yml:57의 핀 관례)로 플랫폼 종속 변경 판정. `continue-on-error: true` + 집계 잡의 `!= 'false'` 비교로 실패-안전(REQ-GFC-011 — release-pr-multi-os.yml:44-60의 동일 관례).
2. 플랫폼 종속 판정 기준(런페이즈에서 확정): `internal/hook/**`(Windows 경로 처리)·`internal/cli/worktree/**`(플랫폼별 git 호출) 등 OS 민감 패키지의 파일 변경. 과도하게 넓히지 않는다 — 전체 `./...`가 아니라 영향 패키지 대상 실행이 카드 요지다.
3. `matrix` 잡: `os: [macos-latest, windows-latest]` × 대상 패키지 `go test` (REQ-GFC-009). ci.yml의 ast-grep 설치 같은 전제가 필요한지는 런페이즈 판별 — 필요하면 해당 패키지만 선별 설치.
4. `multi-os-runtime-gate` 집계 잡: 매트릭스 결과 집합, 안정적 체크 이름 게시(REQ-GFC-010). 스킵 시 성공 보고. **집계 의미론**: 집계 잡은 매트릭스 잡을 `needs`로 요구해 매트릭스 실패가 집계 실패로 전파돼야 한다 — `needs` 없는 독립 녹색 잡이나 `if: always()` 단독의 무조건 성공은 상시 exit-0 변이로서 AC-GFC-009의 변이 단정이 탈락시킨다.

**필수 체크 등록**: decision-index Q2 — 등록은 운영자 콘솔 활. 워크플로는 등록 가능한 이름만 게시한다.

## D-3 SPEC 상태 자동 동기화의 PR 전환 (M2) — v0.5.0 분할로 제외

스코프 4의 전달 설계였다. iter1 D2(무효 플래그)→iter2 D10(인증 배선)→iter3 D16(CI 미발화·auto-merge 정체)의 3계층 정체 끝에 plan-audit iter3(§Final Escalation)이 "무인 자동 병합 계약 자체가 이 저장소의 보호 설정과 양립하지 않는 구조 문제"로 판정했고, v0.5.0에서 본 SPEC에서 분제외(분할)됐다 — decision-index Q3 hold(운영자 3택 결정 대기), 결정 후 별도 카드로 재계획. 이 파일의 D-4·D-5는 번호를 유지한다(재배열 없음).

## D-4 develop 트리거 제거 목록 (M3)

9개 파일·9개 **push 블록** (v0.3.0 — iter1 D4 정정): ci.yml:18, codeql.yml:5, graph-freshness.yml:17-18, docs-i18n-check.yml:33(+27-28 주석 — "develop 병합 경로" 사유의 소멸 서술로 갱신), test-install.yml:5, workflow-parse-guard.yaml:14, judgment-first-consistency.yaml:34, lsel-leak-guard.yaml:11, template-neutrality-check.yaml:31. 각 파일의 REQ-GFC-015 사전 확인 표를 run-phase progress에 남긴다.

**스코프 밖 명시**: graph-freshness.yml:9(`pull_request`)·:15(`pull_request_target`)의 `branches`에 남는 develop은 **제거하지 않는다** — REQ-GFC-013은 push 트리거 한정이고, PR 이벤트의 develop 항목은 develop 원격 삭제 뒤 발화할 이벤트 자체가 사라지는 불활성 잔여다. 그 표기 정리는 cutover 정리 카드의 몫으로 남긴다(스코프 확대 방지).

## D-5 spec-lint 잔여 제거 (M2)

spec-lint.yml:121-139의 develop-push elif와 :83-89·:100-106 주석의 develop 서술을 제거·간소화한다. `:107 base_sha=$(git rev-parse origin/main)`과 fetch 단계(:70-71)는 그대로 — REQ-GFC-008의 불변식.
