# SPEC-GITHUB-FLOW-CI-RESIDUE-001 — Implementation Plan

## §A 전제 (재측정 근거)

모든 결함 진술의 근거는 워크트리 `.moai/worktrees/t1535` (branch `WT-github-flow-ci-residue`, HEAD `5a9d34fbb`)에서 본 회차(2026-10-07) 직접 읽은 본문이다. 판정 요약:

- 스코프 1( spec-lint origin/develop fetch): **핵심 이미 착지** — spec-lint.yml:17-19(main 전용 push)·:70-71(main 전용 fetch, 우회 없음)·:107(기준 origin/main). 잔여는 :121-139의 도달 불가 분기.
- 스코프 2(다중 OS): **유지** — release-pr-multi-os.yml:42, ci.yml:126, race advisory ci.yml:266-280.
- 스코프 3(develop 트리거): **유지 9개** — §A.2 표 참조.
- 스코프 4(spec-status-auto-sync): **분할(hold)** — (a) :41 제목 grep은 iter1 D1에서 기각, (b) :121-124 직접 push는 유지였으나 전달 계약이 iter3 D16(보호 설정과의 구조적 불양립)으로 천장 — v0.5.0에서 본 SPEC에서 제외, decision-index Q3 hold(운영자 3택 대기).
- 스코프 5(착지 검사): **형태 변화** — todo 절반 이미 착지(prlink_landedref.go:63-84, prlink_landed.go:48, git-strategy.yaml:5), worktree 절반 유지(sweep.go:245-254, done.go:287-298, loader_integration_branch.go:121-131).

전제가 사라진 스코프 1의 핵심부는 요구사항에서 뺐고, 잔여만 REQ-GFC-007로 남겼다.

## §B Known Issues (이 카드가 만질 표면의 관측)

1. `sweepDefaultBase`(sweep.go:245-254)는 설정 대상이 비면 오류로 멈추고 `--base`를 요구한다 — 사슬 통합 뒤에도 이 설명 텍스트의 `origin/<branch>` 안내는 갱신이 필요하다.
2. `landingBase`(done.go:287-298)의 거부 사유 문자열(done.go:379)은 `baseProvenance`를 이름붙인다 — 사슬 전환 뒤 provenance 어휘("git-flow/github-flow 해석표")가 사슬 단계 어휘로 바뀐다.
3. `doctor_git_strategy_workflow.go:97`은 `worktree_base_branch ≠ IntegrationTarget` 드리프트를 진단한다 — 사슬 통합 뒤 이 진단의 비교 대상이 무엇인지 재판단이 필요하다(design D-1.3).
4. todo_autodone.go:2-4·:130 외에 todo_triage.go:352의 주석도 origin/develop을 이름붙인다 — 문서 잔여 일괄 점검 대상.
5. (분할 참고) spec-status-auto-sync.yml:92의 staged-diff 가드와 :100-118의 재시도 루프는 PR 전환 후에도 재사용 가능한 형태다 — 스코프 4 분할 카드가 이 관측을 이어받는다.

## §C Pre-flight

- [ ] 워크트리 확인: `git rev-parse --show-toplevel` = `.moai/worktrees/t1535`, branch `WT-github-flow-ci-residue`
- [ ] M1 착지 전 리더·운영자 공유: develop 원격 삭제 작업이 M1 착지 전에 실행되지 않도록 운영 일정 확인 (카드 지시: scope 1·5 재정의가 삭제의 선행 차단 조건)
- [ ] `go test ./internal/cli/worktree/...` 현행 녹색 확인 (수정 전 기준선)
- [ ] CI-YAML 검증 도구 편성: `python3 -c "import yaml, glob, sys; [yaml.safe_load(open(f)) for f in glob.glob('.github/workflows/*.y*ml')]"` (저장소에 actionlint·yamllint 없음 — §A.3)

## §D Constraints

1. **fail-closed 유지**: REQ-GFC-003 — sweep·done의 3-way 계약과 ORIGIN_LANDING_UNCONFIRMED 거부를 조용한 폴백으로 약화하지 않는다. 바뀌는 것은 "기준을 무엇으로 해석하는가"뿐이다.
2. **공개 원칙 유지**: REQ-GFC-002 — todo 사슬의 provenance 공개(REQ-TLA-011)를 worktree 표면 확장에도 적용한다. 단계 3 폴백은 보이게 남는다.
3. **git-flow 프로젝트 호환**: 사슬 통합은 git-flow 사용자의 현행 동작을 깨지 않는다 — `worktree_base_branch: develop` 설정 프로젝트는 1단에서, 미설정 + origin/HEAD=develop 프로젝트는 2단에서 여전히 origin/develop으로 답한다.
4. **D-22 비간섭**: release-pr-multi-os.yml은 수정하지 않는다.
5. **템플릿 무관**: 본 카드가 고치는 워크플로는 dev 저장소 소유이고 템플릿 배포물이 아니다(templates/.github에는 label-sync.yml·detect-language·branch-protection.json.gtmpl만 존재 — 본 회차 확인). `make build` 불요.
6. **커밋 규율**: AGENTS.md §2 — 카드 워크트리 내 작업, 명시적 pathspec 스테이징, 커밋 직전 `git rev-parse --short HEAD` + `git branch --show-current` 재독.

## §E Self-Verification

각 마일스톤 종료 시 acceptance.md의 해당 AC를 실행해 증거를 progress.md §E.2에 남긴다:

- E1: AC 판정표 — 각 AC의 명령·종료 코드 원문
- E2: `go vet ./internal/cli/worktree/... ./internal/cli/... ./internal/factory/...`
- E3: `golangci-lint run ./internal/cli/worktree/... ./internal/factory/...`
- E4: `go test -timeout 30m ./internal/cli/worktree/... ./internal/factory/...` (변경 패키지 한정 — 전체 스위트는 CI 몫)
- E5: YAML 파싱 + 행위 단정 전문 (§C의 편성 도구)
- E6: 수리로 만진 식별자를 이름에 담은 기존 테스트 계열 전체 재실행 (메모리 교훈 t1480 — `go test ./internal/cli/... -run 'Landed|Sweep|Done'` 성격의 계열)

## §F Milestones

### M1 (P0 — 파이프라인 차단 긴급): 착지 검사 기준 브랜치 전환

바뀔 가능성이 가장 큰 결정(착지 판정의 대상 변경)이므로 첫째 마일스톤. 카드 지시("scope 5 = M1")와도 일치.

1. **RED**: internal/cli/worktree에 기준 해석 사슬 테스트 작성 — (a) 설정 `worktree_base_branch: main` → `origin/main` (1단), (b) 미설정 + origin/HEAD → symref 값 (2단), (c) 둘 다 실패 → `origin/main` (3단, 출처 공개), (d) fetch 실패(exit 128 모사) → PRESERVE 유지, (e) git-flow 모양 설정(미설정 + origin/HEAD=develop) → `origin/develop` (호환)
2. **GREEN**: `sweepDefaultBase`(sweep.go:245-254)·`landingBase`(done.go:287-298)를 `factory.LandedRefForWithLevel`(prlink_landedref.go:63) 사슬로 전환. provenance 문자열과 `--base` 플래그 우선순위 유지. `Long` 도움말(sweep.go:201-216)의 기준 서술 갱신.
3. **문서 잔여**: todo_autodone.go:2-4·:130, todo_triage.go:352 주석의 origin/develop 표기 수리 (REQ-GFC-006).
4. doctor 진단 정합(design D-1.3) — 진단 대상 재판단 결과를 테스트나 주석으로 고정.
5. **patch-id 착지 판정 보강 (REQ-GFC-016)**: landing_predicate.go layer 2 — patch-id 후보 발견 뒤 착지 선언 전에 후보 커밋과 카드 tip의 트리 내용 공백 포함 일치 확인(B2·B3 재현 픽스처를 RED로 — research.md §3.1, AC-GFC-016). layer 1·3은 손대지 않는다.
6. **발급 probe 기준 해석 전환 (REQ-GFC-017)**: todo_issuance.go:212의 리터럴 `develop` merge-base 피연산자를 REQ-GFC-001과 동일한 설정 우선 해석으로 전환(main 기반 픽스처 테스트 — AC-GFC-017).
7. 검증: §E 배치 + acceptance AC-GFC-001..006, AC-GFC-016·017.

### M2 (P1): 워크플로 의미 수리 — 다중 OS 집계 게이트·spec-lint 잔여

1. **다중 OS 집계 게이트** (REQ-GFC-009..012): 신규 워크플로(권장 — ci.yml 분할보다 되돌리기 쉽다, design D-2). 플랫폼 종속 변경 감지(dorny/paths-filter, 기존 관례 재사용), macOS·Windows 대상 패키지 `go test`, 단일 집계 잡 `multi-os-runtime-gate`. 필터 오류 시 매트릭스 실행(실패-안전, REQ-GFC-011). 집계 잡은 매트릭스를 `needs`로 요구해 실패가 전파돼야 한다(design D-2.4 — 상시 exit-0 변이 금지).
2. **spec-lint.yml** (REQ-GFC-007): :121-139 도달 불가 분기 제거, :83-89 주석 갱신. REQ-GFC-008 불변식 확인(fetch 단계에 우회 구성 없음).
3. **ci.yml required 목록 주석 갱신 (v0.3.0 D8 흡수)**: ci.yml:266-280의 필수 체크 서술 주석에 신규 집계 게이트 체크 이름을 반영한다 — 주석 전용 변경, 동작 없음.
4. 검증: YAML 파싱 + AC-GFC-007..012.

### M3 (P2): develop push 트리거 제거 — M1·M2 착지 후

1. REQ-GFC-015의 사전 검증: 9개 워크플로 각각이 main 경로(push to main 또는 pull_request)를 유지하는지 표로 확인 — docs-i18n-check.yml은 특히(:27-33의 develop 유지 사유 소멸 확인).
2. 9개 파일·9개 **push 블록**에서 develop 제거 (REQ-GFC-013 — push 한정): ci.yml:18, codeql.yml:5, graph-freshness.yml:**17-18**(push 1블록 — :9 `pull_request`·:15 `pull_request_target`의 develop은 REQ-GFC-013 스코프 밖으로 제거하지 않는다. develop 원격 삭제 뒤 불활성 잔여이며 그 표기 정리는 cutover 정리 카드 귀속 — design D-4), docs-i18n-check.yml:33(+27-28 주석 갱신), test-install.yml:5, workflow-parse-guard.yaml:14, judgment-first-consistency.yaml:34, lsel-leak-guard.yaml:11, template-neutrality-check.yaml:31.
3. 검증: YAML-aware 단정(PyYAML True 키 양쪽 조회 + `on.push` 관측 파일 수 ≥ 1 양성 대조 — AC-GFC-013/015 절차) + AC-GFC-013..015.

### 마일스톤 간 의존

- M2 → M3: REQ-GFC-014 (카드 원문 "의미 수리 후 제거").
- M1은 독립 — 파이프라인 차단 해소가 목적이므로 M2와 병렬 진입 가능하나, 한 레인 한 카드 원칙상 직렬로 M1 → M2 → M3.
- 운영자의 develop 원격 삭제는 M1 착지 + 병합 뒤에만 실행된다(카드 지시의 선행 차단 조건).

## §G Anti-Patterns

- `git add -A`·`git commit -a` 금지 — 명시적 pathspec만.
- primary 체크아웃에서의 작업 금지 — 본 워크트리 안에서만.
- 로컬 전체 스위트(`go test ./...`) 금지 — 변경 패키지 한정, 전체 판정은 CI.
- sweep·done의 PRESERVE·거부를 조용한 폴백으로 약화 금지 — fetch 실패 우회는 REQ-GFC-003·REQ-GFC-008 위반이다.
- "이미 착지" 판정의 무근거 반복 금지 — §A.2의 file:line 인용을 계속 인용하고, 착지 후에는 재측정 값으로 갱신한다.
- 워크플로 편집 시 불필요한 재포맷 금지 — 트리거·단계만 만진다(스코프 절제).

## §H Cross-References

- SPEC-GITHUB-FLOW-DEFAULT-001 (in-progress) — 상위 전환 SPEC; §H 후속 인벤토리가 본 카드의 스코프 3을 예고, §I D-22가 릴리스 게이트 처분을 결정.
- SPEC-TODO-LANDING-ATTRIBUTION-001 (completed) — 착지 ref 3단 해석 사슬의 소유자. M1이 재사용하는 사슬.
- SPEC-WORKTREE-BASEREF-001 (completed) — `worktree_base_branch` 키의 원 소유자(카드 워크트리 생성 베이스). todo 사슬이 이 키를 착지 질의에 재사용한 선례가 M1의 확장 근거.
- acceptance.md — AC-GFC-001..017 대응표와 검증 명령.
- decision-index.md — Q1(POLICY-COVERED)·Q2(FOUNDER, implementation-level, DEFAULT-APPLIED)·Q3(스코프 4 hold — 운영자 3택 결정 대기).
