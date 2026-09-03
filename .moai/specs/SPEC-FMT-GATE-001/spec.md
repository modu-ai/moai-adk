---
id: SPEC-FMT-GATE-001
title: "활성 Go 포맷 게이트 도입 — CI Lint 잡 gofmt 검증 (card t465)"
version: "0.1.0"
status: in-progress
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "format-gate, ci, gofmt, dogfood, quality-gate"
tier: M
---

# SPEC-FMT-GATE-001 — 활성 Go 포맷 게이트 도입

## §A 배경 — 현재 활성 포맷 게이트는 0개

이 저장소에는 **지금 작동하는 Go 포맷 게이트가 존재하지 않는다.** 2026-09-03 본 트리(d592b0551)에서의
실측 근거:

| 표면 | 조사 명령 | 결과 |
|---|---|---|
| CI lint 잡 | `.github/workflows/ci.yml` lint 잡 + `.golangci.yml` | lint 잡은 무조건 실행되지만 golangci-lint enable-set이 `{errcheck, govet, ineffassign, staticcheck, unused}` — 포맷터 linter(gofmt/gofumpt/goimports) 부재 |
| `moai gate` 커밋 게이트 | `internal/cli/gate.go` + `.moai/config/sections/gate.yaml` | 스텝이 vet / typecheck / lint / test / ast-grep — format 스텝 부재 |
| git pre-commit 훅 | `git config --show-origin core.hooksPath` | `file:/Users/goos/MoAI/moai-adk-go/.git/config` → `/dev/null` — **이 저장소에서 git 훅 자체가 비활성** |
| 워크플로 전체 | `grep -rn gofmt .github/workflows/` | 0히트 |

동시에 포맷 부채는 이미 측정돼 있다: `gofmt -l .` @ 트리 `d592b0551` → **154 files**
(tracked-files 변형 `git ls-files -z '*.go' | xargs -0 gofmt -l` 동일 154). 전체 목록:
`.moai/reports/t465/gofmt-l.txt`. 이 154개 파일의 정리는 카드 t457(브랜치 `WT-gofmt-drift`,
tip `e1fdf00d1`) 소관이며 본 SPEC은 정리가 아니라 **게이트 활성화**만을 다룬다.

핵심 함의: 게이트를 지금 활성화하면 CI가 154 위반으로 전면 적색이 된다. 따라서 활성 커밋의
착지 순서가 요구사항의 일부가 된다(REQ-FG-003/REQ-FG-004).

## §B 게이트 표면 구분 — 무엇이 이 저장소 것이고 무엇이 배포물인가

이 저장소는 자기 게이트를 소비(dogfood)하는 동시에 템플릿을 사용자 프로젝트에 배포한다.
두 표면을 분리하여 각 REQ가 어느 표면을 묶는지 명시한다.

| 표면 | 경로 | 사용자 배포 여부 | 묶는 REQ |
|---|---|---|---|
| **이 저장소 자체 CI** (개별 저장소의 dogfood 표면) | 루트 `.github/workflows/ci.yml` | 안 됨 — 배포 템플릿의 `.github/workflows/`에는 `label-sync.yml`만 있음 | REQ-FG-001, REQ-FG-002, REQ-FG-004 |
| **로컬 개발 표면** (개별 저장소, repo-local) | 루트 `Makefile`, `git config core.hooksPath` | 안 됨 — `core.hooksPath`는 repository-local 설정이며 사용자 프로젝트에 전파되지 않음; 현재 `/dev/null`로 훅 비활성 상태 | REQ-FG-006 |
| **배포 표면** (사용자 프로젝트에 실림) | `internal/template/templates/**` | 됨 — `moai init`/`moai update`로 배포 | REQ-FG-005 (금지로 묶음) |

설계 결정: 게이트 활성 표면으로 **루트 CI lint 잡**을 선택했다. 사유:
1. CI는 이 저장소의 통합 판정 기관이며(모든 push/PR에서 무조건 실행) 로컬 훅 상태와 무관하게 기계적으로 강제된다.
2. pre-commit 표면은 이 저장소에서 `core.hooksPath=/dev/null`로 **비활성** — 게이트를 두어도 "문서로만 존재"하는 게이트가 된다(카드 미션 위반).
3. `moai gate`에 format 스텝을 추가하는 것은 배포 제품 코드(`internal/hook/quality`) 변경이며 모든 사용자 프로젝트의 커밋 게이트 거동을 바꾼다 — 별도 카드 소관의 명시적 결정이다.

## §C 요구사항 (GEARS)

### REQ-FG-001 — 포맷 위반 시 CI 실패 (event-driven)

**When** `gofmt -l .` 를 저장소 루트에서 실행했을 때 하나 이상의 파일 경로가 출력되는 트리에서 CI Lint 잡이
실행되면, the format-gate step shall 해당 CI 런을 실패시킨다(non-zero exit).

### REQ-FG-002 — 정상 트리 통과 (event-driven)

**When** `gofmt -l .` 를 저장소 루트에서 실행했을 때 출력이 0행이면, the format-gate step shall 해당 CI 런을
통과시킨다(exit 0).

### REQ-FG-003 — 착지 순서 고정: t457 선행 (capability gate)

**Where** activation commit(포맷 게이트를 활성화하는 커밋)이 저장소 역사에 존재하면, the activation commit
shall 커밋 `e1fdf00d1`(카드 t457 정리 브랜치 `WT-gofmt-drift` tip)을 조상으로 포함한다.
기계 판정: `git merge-base --is-ancestor e1fdf00d1 <activation-commit>` → exit 0.

### REQ-FG-004 — 활성 시점 녹색 (state-driven)

**When** activation commit이 생성되면, the activation commit's tree shall `gofmt -l .` 출력 0행을 만족한다.
기계 판정: activation commit을 체크아웃한 트리에서 `gofmt -l . | wc -l` → `0`.

### REQ-FG-005 — 배포 표면 불변 (unwanted)

이 SPEC이 인도하는 모든 커밋은 `internal/template/templates/**` 하위 어떤 파일도 수정하지 않는다
(the SPEC's commits shall not modify the deployment surface). 템플릿 중립성(SPEC-ID·내부 날짜·커밋 SHA·
macOS 편향 경로 금지, 16 프로그래밍 언어 동등 취급)은 그대로 유지된다.

### REQ-FG-006 — 로컬 패리티 검사 (capability gate)

**Where** 개발자가 저장소 루트에서 `make fmt-check` 를 실행하면, the Makefile shall
`git ls-files -z '*.go' | xargs -0 gofmt -l` 출력이 1행 이상일 때에만 non-zero로 종료한다
(tracked `.go` 한정 — untracked 노이즈 제외. clean checkout에서 전수 `gofmt -l .`로 판정하는
CI 게이트와 동치).

## §D 제약

- **게이트 기준선은 gofmt다.** `make fmt`(gofumpt)는 상위집합 포맷터로 남으며 gofumpt 출력은
  gofmt-clean하므로 기존 수정 경로가 그대로 유효하다. gofumpt를 게이트 기준선으로 올리는 것은
  154개 밖의 추가 위반을 만드는 별도 결정이다(Out of Scope).
- **게이트 범위는 저장소 루트 전체 `.go` 파일**이다 — `internal/navigator/astx/testdata/` 계열
  fixture `.go` 2개 포함(현재 154 목록에 포함, t457이 정리). 생성물 `internal/web/*_templ.go`도
  gofmt-clean을 유지해야 한다(templ codegen이 표준 출력은 gofmt-clean하므로 실측상 위험 낮음).
- 로컬 측정 시 worktree의 untracked `.go` 노이즈를 피하려면 tracked-files 변형
  `git ls-files -z '*.go' | xargs -0 gofmt -l` 을 쓴다(2026-09-03 실측에서 전수와 동일 결과).
- 전체 테스트 스위트 로컬 실행 금지(레인 검증 규율) — 본 SPEC의 런 검증은 게이트 명령 자체와
  CI 판정으로 충분하다.

## §E Out of Scope

### Out of Scope — 배포 표면(사용자 프로젝트 포맷 게이트)

- `internal/template/templates/**` 어떤 변경도 하지 않는다. 사용자 프로젝트에 포맷 게이트를
  배포하는 것(템플릿 CI 추가, `moai gate` format 스텝 추가 포함)은 별도 카드의 명시적 결정이다.
- 배포 템플릿의 `.github/workflows/`에 포맷 게이트 워크플로를 추가하지 않는다.

### Out of Scope — pre-commit / core.hooksPath 표면

- 이 저장소의 `core.hooksPath`(현재 `/dev/null`)를 변경하지 않는다. repository-local 설정이며
  게이트 활성 표면으로 평가·기각했다(§B).

### Out of Scope — 154개 파일 포맷 정리 및 gofumpt 게이트화

- 154개 파일의 gofmt 정리는 카드 t457 소관이다. 본 SPEC의 커밋은 어떤 `.go` 파일도
  `gofmt -w`로 수정하지 않는다(drive-by 정리 금지).
- 게이트 기준선을 gofumpt로 상향하는 것, `.golangci.yml` enable-set에 포맷터 linter를 추가하는
  것(pinning 주석이 명시하는 "deliberate future decision")은 본 SPEC 범위 밖이다.

## HISTORY

- 2026-09-03 — 최초 작성(plan-phase, card t465, Tier M). 기준선 실측: `gofmt -l .` @ `d592b0551`
  → 154 files(`.moai/reports/t465/gofmt-l.txt`). 선행 카드 t457 tip `e1fdf00d1` 착지 순서 REQ-FG-003으로 고정.
- 2026-09-03 (2차) — 리드 결정 반영(plan.md §I): D1 `make fmt-check`를 활성 커밋과 동일 커밋
  인도로 확정(BINDING), D2 `.golangci.yml` gofmt linter 제외 승인(기록 전용). plan-audit D1 결함
  수리: REQ-FG-006 판정 기준을 tracked-files 변형으로 정정(acceptance.md §D.3 표면과 일치).
