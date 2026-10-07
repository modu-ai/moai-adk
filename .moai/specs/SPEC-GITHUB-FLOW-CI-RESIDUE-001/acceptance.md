# SPEC-GITHUB-FLOW-CI-RESIDUE-001 — Acceptance Criteria

## §A 측정 규약

- 모든 AC는 이분법으로 판정한다 — 명령 종료 코드·grep 계수·테스트 통과 여부.
- 검증 명령은 워크트리 루트(`.moai/worktrees/t1535`)에서 실행하는 평문 명령이다.
- Go 계열 검증은 변경 패키지 한정이다(`-timeout 30m`) — 전체 스위트 판정은 CI 몫이다.
- YAML 검증 편성: `python3` + `yaml.safe_load` (저장소에 actionlint·yamllint 부재 확인 — spec.md §A.3).
- RED-now 원장(verification-completeness.md §2.1 — 명령·원문 stdout·종료 코드·트리 SHA `5a9d34fbb3a2ba4a3cb0596d9db633f9ec55c95c` 4요소, 전부 본 회차 v0.3.0 시점 관측). AC 자체 검증 명령이 참조하는 단위 테스트는 런페이즈 TDD가 창출하므로, plan 시점의 적근거는 아래의 전제 probe 관측이다 — 이 시점의 `go test -run` 선택 0(`no tests to run`)은 판정 근거가 아닌 빈 스위프다.

| AC | 명령 (단일 실행) | 원문 stdout | 종료 | 판정 |
|---|---|---|---|---|
| AC-GFC-016 | `sh /tmp/t1535-patchid/run2.sh` | `B2_indent card=7688e79215b0cd5f2f46c96516bae3a0987d703c layer2_landed=yes` · `B3_trailing card=7688e79215b0cd5f2f46c96516bae3a0987d703c layer2_landed=yes` | 0 | **적색** — 공백성 미반영 커밋이 착지 판정됨 (B1 신규 줄은 `layer2_landed=no` 미재현) |
| AC-GFC-017 | `grep -n 'merge-base", "develop' internal/cli/todo_issuance.go` | `212:	base := issuanceGitOneLine(root, "merge-base", "develop", branch)` | 0 | **적색** — 리터럴 develop 기준 |
| AC-GFC-005 | `grep -c "LandedRefFor" internal/cli/worktree/sweep.go internal/cli/worktree/done.go` | `internal/cli/worktree/done.go:0` · `internal/cli/worktree/sweep.go:0` | 1 (무적중) | **적색** — 사슬 미채택 |
| AC-GFC-006 | `grep -rn "origin/develop" internal/cli/todo_autodone.go` | `internal/cli/todo_autodone.go:4:// \`git rev-parse origin/develop\`).` · `:130:confirmation (git fetch origin develop + git rev-parse origin/develop).` | 0 | **적색** — 잔여 서술 2건 |
| AC-GFC-007 | `grep -c 'REF_NAME" == "develop"' .github/workflows/spec-lint.yml` | `1` | 0 | **적색** — 도달 불가 분기 존재 |
| AC-GFC-013 | 리터럴 명령: 표 아래 [013] fenced 원장 | [013] 원문 10행 | 0 | **적색** — 9파일 검출(이 관측이 D6 양성 대조를 겸한다) |
| AC-GFC-001·002·003·004 | — 현재 나무에서 단일 명령 관측 불가(판정 대상 테스트는 런페이즈가 창출) | — | — | **§2.1 undecidable disposition** — plan 시점 release-blocking 자격 보류(regression-guard 분류), 런페이즈 TDD RED(E8)가 실패 출력을 포획해 두 셀을 완성한다. 녹색 경로 셀: M1이 뒤집는다 |

- AC-GFC-008·010·011·012·014·015은 유지·단정·과정 성격으로 RED 셀을 강제하지 않는다(015의 현행 적근거는 AC-GFC-013의 9파일 관측이 공유한다). 스코프 4 분할(v0.5.0)로 옛 AC-GFC-013(regression-guard)·014(직접 push)·그 적근거 행은 본 SPEC 원장에서 제외됐다 — 분할 카드의 원장으로 이관된다.

[013] 명령 (단일 실행 — 파이프·체인 없음):

```
python3 -c "
import yaml, glob
seen = 0
for f in sorted(glob.glob('.github/workflows/*.y*ml')):
    d = yaml.safe_load(open(f))
    trig = d.get('on') if isinstance(d.get('on'), dict) else d.get(True) if isinstance(d.get(True), dict) else None
    if not isinstance(trig, dict):
        continue
    push = trig.get('push')
    branches = push.get('branches') if isinstance(push, dict) else (push if isinstance(push, list) else None)
    if branches and 'develop' in branches:
        print('DEVELOP_PUSH_TRIGGER:', f)
        seen += 1
print('files_with_develop_push =', seen)
"
```

[013] 원문 stdout:

```
DEVELOP_PUSH_TRIGGER: .github/workflows/ci.yml
DEVELOP_PUSH_TRIGGER: .github/workflows/codeql.yml
DEVELOP_PUSH_TRIGGER: .github/workflows/docs-i18n-check.yml
DEVELOP_PUSH_TRIGGER: .github/workflows/graph-freshness.yml
DEVELOP_PUSH_TRIGGER: .github/workflows/judgment-first-consistency.yaml
DEVELOP_PUSH_TRIGGER: .github/workflows/lsel-leak-guard.yaml
DEVELOP_PUSH_TRIGGER: .github/workflows/template-neutrality-check.yaml
DEVELOP_PUSH_TRIGGER: .github/workflows/test-install.yml
DEVELOP_PUSH_TRIGGER: .github/workflows/workflow-parse-guard.yaml
files_with_develop_push = 9
```

## §B 공통 품질 게이트

- `go vet ./internal/cli/worktree/... ./internal/cli/... ./internal/factory/...` — 종료 0.
- `golangci-lint run ./internal/cli/worktree/... ./internal/factory/...` — 종료 0.
- `go test -timeout 30m ./internal/cli/worktree/... ./internal/factory/...` — 전부 통과.
- `python3 -c "import yaml, glob, sys; [yaml.safe_load(open(f)) for f in glob.glob('.github/workflows/*.y*ml')]" && echo YAML_OK` — `YAML_OK` 출력.

## §C 인수 기준

### AC-GFC-001 — 설정된 기준 브랜치가 착지 판정의 대상이다 (maps REQ-GFC-001)

- Given 본 저장소의 `git_strategy.worktree_base_branch: main` 설정 (git-strategy.yaml:5)
- When M1의 신규 단위 테스트가 `sweepDefaultBase`와 `landingBase`의 파생값을 요구하면
- Then 양쪽 모두 `origin/main`을 반환한다 (1단 — 설정이 답한다).
- 검증: `go test ./internal/cli/worktree/ -run 'SweepDefaultBase|LandingBase' -v` — 신규 케이스 포함 전부 통과.

### AC-GFC-002 — 하위 단계 폴백과 출처 공개 (maps REQ-GFC-002)

- Given 설정값이 비어 있고 origin/HEAD symref가 `refs/remotes/origin/main`을 가리키는 픽스처
- When 기준 해석을 요구하면
- Then `origin/main`과 2단(origin/HEAD) 출처를 반환하고, symref 읽기 실패 시에는 `origin/main`과 3단 출처를 반환한다.
- 검증: `go test ./internal/cli/worktree/ -run 'BaseProvenance|LandedRef' -v` — 단계별 케이스 통과.

### AC-GFC-003 — fetch 실패의 fail-closed 유지 (maps REQ-GFC-003)

- Given 기준 원격 fetch가 exit 128로 실패하는 러너 모사 (develop 삭제 상태의 이력 재현)
- When sweep 판정을 실행하면
- Then 해당 트리는 PRESERVE 판정을 받는다 — 조용한 폴백·삭제가 일어나지 않는다.
- 검증: `go test ./internal/cli/worktree/ -run Sweep -v` — 기존 3-way 계약 테스트 + 신규 케이스 통과.

### AC-GFC-004 — develop 원격 부재에서도 착지가 답한다 (maps REQ-GFC-004)

- Given `git fetch origin develop`는 실패하고 `git fetch origin main`은 성공하는 러너 모사
- When 착지한 카드 브랜치의 sweep·done 판정을 실행하면
- Then 기준 origin/main으로 landed 판정이 나온다 — 전 트리 PRESERVE·일괄 거부가 재현되지 않는다.
- 검증: `go test ./internal/cli/worktree/ -run 'Sweep|OriginLanding' -v` — 모사 케이스 통과.

### AC-GFC-005 — 두 표면이 하나의 사슬을 쓴다 (maps REQ-GFC-005)

- Given M1 착지 후의 코드
- When worktree 표면의 기준 해석 진입점을 찾으면
- Then todo 표면과 같은 `factory.LandedRefFor*` 사슬(또는 동일 위임)을 호출한다.
- 검증: `grep -n "LandedRefFor" internal/cli/worktree/*.go` — sweep.go·done.go에서 ≥2 적중.

### AC-GFC-006 — 존재하지 않는 origin/develop 기준 서술 0 (maps REQ-GFC-006)

- Given M1 착지 후의 코드
- When 착지 검사 사용자 대상 텍스트에서 origin/develop을 찾으면
- Then 적중이 0건이다.
- 검증: `grep -rn "origin/develop" internal/cli/todo_autodone.go internal/cli/todo_triage.go` — 0건, 예외 출구 없음(v0.3.0 D7 — 주석도 REQ-GFC-006의 대상이므로 판단형 예외를 둬서는 안 된다; 현재 2건 적색, §A 원장). worktree 표면의 도움말·거부 사유 텍스트의 정확성은 design D-1.1·D-1.2가 갱신하고 AC-GFC-001·002가 사슬 동작으로 검증한다.

### AC-GFC-007 — spec-lint의 도달 불가 develop 분기 0 (maps REQ-GFC-007)

- Given M2 착지 후의 spec-lint.yml
- When develop push 정책 분기를 찾으면
- Then 존재하지 않는다.
- 검증: `grep -c 'REF_NAME" == "develop"' .github/workflows/spec-lint.yml` — 0.

### AC-GFC-008 — spec-lint fetch 우회 구성 부재 (maps REQ-GFC-008)

- Given M2 착지 후의 spec-lint.yml
- When 통합 기준 fetch 단계를 읽으면
- Then `|| true` 등 실패 무시 구성이 없다 (현행 유지 — spec-lint.yml:70-71).
- 검증: `sed -n '/Fetch integration refs/,+1p' .github/workflows/spec-lint.yml` — `git fetch origin main:refs/remotes/origin/main` 단독, 무시 구성 부재.

### AC-GFC-009 — 플랫폼 종속 PR이 다중 OS 런타임 검증을 받는다 (maps REQ-GFC-009)

- Given 플랫폼 종속 변경(픽스처 diff)을 담은 feature→main PR
- When 집계 게이트 워크플로가 실행되면
- Then macOS 러너와 Windows 러너에서 `go test`가 실행되고 결과가 단일 집계 체크로 합산된다.
- 검증 (YAML-aware, mutant-proof — v0.3.0 D3 + v0.4.0 D11): python3 `yaml.safe_load`로 (a) 매트릭스 잡의 `strategy.matrix.os` **목록 원소**를 계수해 `{macos-latest, windows-latest} ⊆ os 목록`을 단정한다 — 행 계수가 아니라 원소 계수(단일 행 `os: [macos-latest, windows-latest]`도 2로 계수돼야 한다 — 행 계수 1 탈락 변이 차단); (b) 매트릭스 **잡의** `runs-on`이 `matrix.os`를 참조함을 단정; (c) 집계 잡이 매트릭스 잡을 `needs`로 요구함을 단정하고, 매트릭스 실패를 삼키는 무조건 성공 구성(집계 잡의 `if: always()` 단독 — `needs` 결과를 보지 않는 형태)이 없음을 단정한다 — 매트릭스가 전부 실패해도 집계만 녹색인 변이를 탈락시킨다; **(d)** `matrix.exclude`(및 include)를 반영한 **실효 OS 집합** ⊇ {macos-latest, windows-latest}를 단정한다 — 선언 목록을 그대로 두고 exclude로 Windows 실행만 빼는 변이를 탈락시킨다; **(e)** 매트릭스 잡의 step 본문에 실제 `go test` 호출이 존재함을 단정한다 — 테스트를 no-op(echo)로 바꾸는 변이를 탈락시킨다. 변이 검증은 픽스처로 재현돼 있다(v0.4.0 자가검증: canonical (a)-(e) PASS, exclude 변이 (d) FAIL, no-op 변이 (e) FAIL — (a)-(c)만으로는 두 변이 모두 통과). 런타임 행위는 착지 뒤 실제 PR에서 간접 검증(§E).

### AC-GFC-010 — 집계 게이트의 안정적 단일 체크 이름 (maps REQ-GFC-010)

- Given 집계 게이트 워크플로
- When 체크 이름을 찾으면
- Then 정확히 하나의 안정적 이름(예: `Multi-OS Runtime Gate`)이 정의돼 브랜치 보호 등록 대상이 될 수 있다.
- 검증: `grep -n "name: Multi-OS Runtime Gate" .github/workflows/<gate>.yml` — 집계 잡에 정확히 1건.

### AC-GFC-011 — 스킵·실패-안전 방향 (maps REQ-GFC-011)

- Given 문서 전용(비-플랫폼) PR과 필터 오류 상황
- When 집계 게이트가 평가되면
- Then 비-플랫폼 PR은 매트릭스를 건너뛰고 성공을 보고하며, 필터 오류 시에는 매트릭스가 실행된다.
- 검증: 워크플로 본문에서 paths-filter의 `continue-on-error: true` + 집계 잡의 `!= 'false'` 방향 단정 — release-pr-multi-os.yml:44-60과 동일 관례임을 `grep -n "continue-on-error" <gate>.yml`로 확인.

### AC-GFC-012 — cross-compile이 런타임 검증으로 기록되지 않는다 (maps REQ-GFC-012)

- Given CI 구성 전체
- When cross-compile build 잡의 역할 서술을 찾으면
- Then 그것이 다중 OS 런타임 검증을 대체한다는 서술이 없고, 집계 게이트는 `go test` 실행을 요구한다.
- 검증: `grep -rn "cross-compile" .github/workflows/ci.yml | head` — 기존 서술(:110-116) 유지 + 신규 게이트가 `go test` 기반(AC-GFC-009의 단정과 합유일).

### AC-GFC-013 — develop push 트리거 0 (maps REQ-GFC-013)

- Given M3 착지 후의 `.github/workflows/` 전체
- When 각 워크플로의 `on.push.branches`를 YAML로 파싱하면
- Then develop이 어디에도 없다.
- 검증 (YAML-aware, v0.3.0 D6): python3 `yaml.safe_load` 후 트리거 사전을 `d.get('on')`과 `d.get(True)` **양쪽**에서 조회한다 — PyYAML 1.1은 bare `on:`을 불리언 True 키로 파싱하므로(실측 `('True','bool')`) 한쪽만 보면 0블록 순회가 허위 녹색이 된다. **양성 대조**: 절차는 탐색한 워크플로 중 `on.push` 사전을 실제로 본 파일 수를 함께 보고하며, 그 수가 0이면 "0 검출"이 아니라 탐색 불능 오류로 판정한다 — 제거 전 이 절차는 9개를 검출한다(§A 원장 관측 `files_with_develop_push = 9`, 본 관측이 로더 능력의 증명). 제거 후 판정: `develop` 검출 0 + `on.push` 관측 파일 수 ≥ 1, 종료 0.

### AC-GFC-014 — 삭제 순서의 준수 (maps REQ-GFC-014)

- Given M1·M2가 아직 착지하지 않은 시점
- When M3 작업이 시작되려 하면
- Then 진입이 보류된다 — progress 기록에 순서 근거가 남는다.
- 검증: 과정 AC — run-phase progress.md에 M3 진입 시점의 M1·M2 착지 증거(커밋 SHA·AC 참조)가 기록돼 있는지 sync-phase가 확인.

### AC-GFC-015 — 삭제가 검사 공백을 만들지 않는다 (maps REQ-GFC-015)

- Given develop 트리거가 제거된 9개 워크플로 각각
- When main 대상 경로를 확인하면
- Then push to main 트리거가 남아 있거나, pull_request 트리거가 남아 있으면 그 branches가 main을 포함하거나 무필터다 — 동일 검사 경로가 main 대상으로 유지된다.
- 검증 (YAML-aware, v0.3.0 D6 + v0.4.0 D14 강화): AC-GFC-013과 동일한 True-key 양쪽 조회 로더를 쓴다 — 제거 대상 9개 파일 각각에 대해 ① `on.push.branches`에 `main` 포함, **또는** ② `on.pull_request`가 존재하면 그 `branches`에 `main` 포함 또는 `branches` 키 자체가 없음(무필터)을 단정한다 — `push.branches: [other]` + `pull_request.branches: [develop]` 조합의 검사 경로 소실 변이를 탈락시킨다. 양성 대조: 제거 직전 같은 절차가 9/9 파일을 판정 대상으로 검출함을 확인한 뒤 실행한다(0블록 순회는 오류). 종료 0.

### AC-GFC-016 — 공백성 미반영 커밋이 착지로 위장하지 않는다 (maps REQ-GFC-016)

- Given squash 병합된 변경 위에 미반영 줄 내부 공백 변경(들여쓰기 또는 후행 공백) 커밋을 얹은 카드 브랜치 픽스처
- When `LandedByPatchID`가 layer 2를 평가하면
- Then `(false, nil)`을 반환한다 — 착지 아님(PRESERVE) — 공백이 같아도 트리 내용이 다르면 착지로 선언하지 않는다.
- 검증: `go test ./internal/cli/worktree/ -run 'LandedByPatchID|PatchID' -v` — B2(들여쓰기)·B3(후행 공백) 픽스처 케이스 통과. RED-now 근거: 본 회차 `/tmp` 실증 — 미반영 들여쓰기 카드의 누적 patch-id `7688e79215b0cd5f2f46c96516bae3a0987d703c`가 squash 커밋의 patch-id와 동일 → `layer2_landed=yes` (research.md §3, HEAD `5a9d34fbb`).

### AC-GFC-017 — 발급 probe의 변경 파일 기준이 설정 해석을 따른다 (maps REQ-GFC-017)

- Given M1 착지 후의 todo_issuance.go
- When `productionLaneFilesProbe`의 merge-base 피연산자를 검사하면
- Then 리터럴 `develop`이 아니라 프로젝트 통합 베이스 해석(설정 우선 사슬)의 값이 쓰인다.
- 검증: `go test ./internal/cli -run 'LaneFilesProbe|Issuance' -v` — main 기반 픽스처에서 files 목록이 산출되는 케이스 통과 + `grep -c 'merge-base", "develop' internal/cli/todo_issuance.go` — 0.

## §D 추적표

| REQ | AC | 마일스톤 | 검증 성격 |
|---|---|---|---|
| REQ-GFC-001 | AC-GFC-001 | M1 | 단위 테스트 |
| REQ-GFC-002 | AC-GFC-002 | M1 | 단위 테스트 |
| REQ-GFC-003 | AC-GFC-003 | M1 | 단위 테스트(러너 모사) |
| REQ-GFC-004 | AC-GFC-004 | M1 | 단위 테스트(러너 모사) |
| REQ-GFC-005 | AC-GFC-005 | M1 | grep |
| REQ-GFC-006 | AC-GFC-006 | M1 | grep (0건, 예외 없음) |
| REQ-GFC-007 | AC-GFC-007 | M2 | grep |
| REQ-GFC-008 | AC-GFC-008 | M2 | 본문 판독 |
| REQ-GFC-009 | AC-GFC-009 | M2 | YAML 원소 계수 + 집계 전파 단정 + 간접(CI) |
| REQ-GFC-010 | AC-GFC-010 | M2 | grep |
| REQ-GFC-011 | AC-GFC-011 | M2 | 본문 단정 |
| REQ-GFC-012 | AC-GFC-012 | M2 | 본문 단정 |
| REQ-GFC-013 | AC-GFC-013 | M3 | YAML 단정 (True-key + 양성 대조) |
| REQ-GFC-014 | AC-GFC-014 | M3 | 과정 기록 |
| REQ-GFC-015 | AC-GFC-015 | M3 | YAML 단정 (True-key + 양성 대조) |
| REQ-GFC-016 | AC-GFC-016 | M1 | 단위 테스트(B2·B3 픽스처) + /tmp 실증 RED |
| REQ-GFC-017 | AC-GFC-017 | M1 | 단위 테스트 + grep |

## §E 간접 검증과 이 SPEC이 통과로 기록하지 않는 것

- **간접**: AC-GFC-009의 실제 러너 실행은 착지 뒤 첫 실제 PR의 CI 관측으로 간접 검증한다 — 워크트리에서 macOS·Windows 러너를 직접 돌릴 수 없다.
- **미측정**: 브랜치 보호 필수 체크 등록 여부(decision-index Q2 — 운영자 콘솔 활)는 본 SPEC의 통과 판정에 넣지 않는다. 워크플로가 등록 가능한 이름을 게시하는 것까지(AC-GFC-010)가 이 SPEC의 경계다.
- **미측정**: develop 원격 삭제 자체 — M1이 선행 차단 조건을 충족시키는 것까지만 증명한다.
- 본 SPEC은 로컬 전체 스위트 녹색을 요구하지 않는다 — 변경 패키지 녹색 + push 뒤 CI 판정이 계약이다.
