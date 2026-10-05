---
id: SPEC-CODEMAPS-FOLD-GUARD-001
title: "plan — codemaps fold 판정 단위 산문 보존 가드"
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
tier: S
---

# plan.md — SPEC-CODEMAPS-FOLD-GUARD-001

> 진행 기록은 `progress.md`가 운반한다. 이 문서는 상태 축이 없다(스키마 § Artifact Statelessness).

## §A. Context

- **발주**: 카드 t748. SPEC-CODEMAPS-REFRESH-002 run 중 첫 재생성(commit `e397ec00d`)이 5개 fold 단위 전부에 산문을 기록했고(§A.3(a1)을 편입 허가로 오용), M3(`cd03be0d3`)이 되돌렸다. 이 SPEC은 재발을 기계적으로 불가능하게 한다.
- **위치**: 카드 워크트리에서 run. 브랜치 `WT-codemaps-fold-ac`, 기준 HEAD `146faed9d` (plan-phase 저작 시점 — run 시작 시 재확인).
- **기준선(146faed9d 실측)**: 생성기 5문서 fold 토큰 히트 0(exit 1) / `fold-judgments.txt` fold 행 5(exit 0) / 가드 테스트 부재(exit 1). spec.md §A.2, §D 참조.
- **기존 인프라**: PRESERVE — `.moai/project/codemaps/**`(동결), `.moai/reports/t475/**`(동결), `internal/graph/check*.go`(t688이 안정시킨 체커 계약). EXTEND — `internal/graph`에 테스트 1개.

## §B. Known Issues (이 카드와 관련 있는 것만)

- **B-해당 (D1 계승)**: 손 열거한 보호 집합은 자기 결함을 재생산한다(REFRESH-002 plan-audit iter-1 D1). REQ-CFG-001이 데이터 주도 + floor 핀의 이중 잠금을 요구한 이유.
- **B-해당 (t747 F5)**: 탐침이 관측을 기록만 하고 단언하지 않는 결함 계열. 가드는 테스트 단언(`t.Errorf`/`t.Fatal`)으로 닫는다 — 로그만 남기는 헬퍼 함수 형태 금지.
- **B-해당 (verification-completeness §1.1)**: 보고서가 마지막 성공의 종료 코드를 돌려주는 report-not-verdict. 가드 실패는 반드시 테스트 실패로 귀결돼야 한다 — 보조 함수가 실패를 "반환만" 하는 구조에서 호출자가 단언을 잊으면 같은 결함이 된다.
- **B1 (해당 없음)**: 테스트 전용 변경 — syscall·빌드 태그 없음. 단 `GOOS=windows go build`는 E2에서 유지 확인한다.
- **B-해당 (동결 표면)**: `.moai/project/codemaps/**`는 기준선이자 다른 세션의 읽기 대상이다. RED 입증을 포함해 그 어떤 단계에서도 변조하지 않는다(REQ-CFG-005).

## §C. Pre-flight

run 시작 시 실행하고 출력을 progress.md에 남긴다:

```bash
# 1. 브랜치 + HEAD (§A의 값과 다르면 흡수 발생 — 재확인 후 진행)
git branch --show-current
git rev-parse --short HEAD

# 2. 기준선 3관 재측정 (spec.md §A.2와 동일 명령)
ls internal/graph/codemaps_fold_guard_test.go; echo "EXIT=$?"
grep -rn -F -e 'doctor_hook_delivery' -e 'step_git_env' -e 'prlink_landedref' \
  -e 'fieldsets_codex_templ' -e 'internal/core/git' \
  .moai/project/codemaps/overview.md .moai/project/codemaps/modules.md \
  .moai/project/codemaps/dependencies.md .moai/project/codemaps/entry-points.md \
  .moai/project/codemaps/data-flow.md; echo "EXIT=$?"   # 1이어야 한다
grep -c '^fold ' .moai/project/codemaps/fold-judgments.txt; echo "EXIT=$?"  # 5, exit 0

# 3. depends_on 상태 (둘 다 completed여야 한다)
grep -m1 '^status:' .moai/specs/SPEC-CODEMAPS-REFRESH-002/spec.md
grep -m1 '^status:' .moai/specs/SPEC-GRAPH-STAMP-ANCESTRY-001/spec.md

# 4. 테스트 패키지 현재 상태 (신규 실패 vs 기존 baseline 구분)
go test ./internal/graph/ 2>&1 | tail -3
```

## §D. Constraints

- **PRESERVE (변경 금지)**: `.moai/project/codemaps/**`(생성기 5문서 + fold-judgments.txt + provenance.json 전부), `.moai/reports/t475/**`, `internal/graph/`의 기존 프로덕션 파일, `.github/workflows/**`, `gate.yaml`.
- **변경 허용 경로**: `internal/graph/codemaps_fold_guard_test.go`(신설 테스트)와 `.moai/specs/SPEC-CODEMAPS-FOLD-GUARD-001/**`뿐이다. 프로덕션 Go 변경 0줄.
- 금지 명령: `--no-verify`, force-push, `git add -A`/`git add .`(명시적 pathspec만).
- 판정 규약 고정: exact-substring 히트(`grep -c -F` 계승). REQ 줄-형태 패턴(reqLineWidePattern 계열)을 판정에 수입하지 않는다 — 세 축 층 구분(spec.md §A.3).
- 테스트가 저장소 루트를 찾지 못하면 **skip이 아니라 실패**다 — 조용한 스킵은 비침묵 축 위반(verification-completeness §1.3의 부재 실행을 녹색으로 위장하는 형태)이다.
- 커밋: Conventional Commits + `🗿 MoAI` 트레일러, 카드 id(t748)를 메시지에 운반.

## §E. Self-Verification (run 완료 보고에 전부 수출)

각 항목은 5단계 증거 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)과 귀속 삼인조(명령 + verbatim 출력 + 트리 SHA)로 보고한다.

- **E1**: AC-CFG-001~005 PASS/FAIL 행렬 — 검증 명령과 실출력 포함.
- **E2**: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` → 둘 다 exit 0.
- **E3**: `go test ./internal/graph/` (변경 영향 패키지 — 전체 스위트는 CI 몫) → `ok`.
- **E5**: `golangci-lint run --timeout=2m` — 신규 이슈 0 (baseline과 구분 보고).
- **E7**: blocker 보고(있으면).
- **E8**: RED 실패 출력 verbatim — AC-CFG-002/003의 변조 사본 실패(`--- FAIL` 블록 전문)와 복원 후 PASS 출력. **RED 없이 GREEN만 있으면 이 항목은 공란이 아니라 미이행이다.**

## §F. Milestones

결정 역전성 순 — 바뀔 개연성이 큰 결정(판정 규약·집합 원천)을 앞에 두고 기계적 단계(파일 명명·커밋)를 뒤에 둔다.

- **M1 (Priority High) — 가드 테스트 신설.** `internal/graph/codemaps_fold_guard_test.go`에 REQ-CFG-001~004를 구현한다. 설계 결정(구현이 아니라 계약 수준):
  1. **판정 규약**: 문서 한 행이 보호 토큰(`fold-judgments.txt`의 경로·파일명 그대로)을 exact-substring으로 담으면 위반. `core/git` 단축형은 위반 아님 — 실측 근거 `overview.md:103`/`modules.md:171`(접힌 결과물).
  2. **집합 원천**: `fold-judgments.txt` 파싱(`^fold <unit>$`) + floor 5단위 핀. 기록 부재·파싱 실패·floor 누락은 전부 실패.
  3. **스탬프 무관성**: `provenance.json` 미독, `moai graph check` 미호출 — 판정 입력은 문서 사본과 기록 파일뿐.
  4. **입력 주입**: RED 입증을 위해 문서 디렉터리 입력을 주입 가능하게 한다(내부 헬퍼 시그니처 — HOW는 실행자 판단). live 표면은 어떤 경로로도 변조하지 않는다. 기본 스위트가 도는 실트리 스캔과 `/tmp` 주입 스캔은 **같은 판독 함수**를 흐른다(기본값 = 실제 문서 디렉터리로 매개변수화) — AC-CFG-002의 RED 증거가 실트리 스캔으로 전이되는 것은 이 단일 판독 경로가 보장한다.
  TDD: RED는 M2의 fixture로 관측한다 — 이 테스트 자체가 RED-first 대상인 "가드 부재" 상태에서 출발하므로, AC-CFG-001의 RED-now(ls 부재)가 그 자리를 대신한다.
- **M2 (Priority High) — RED/GREEN 입증 (fixtures 위에서).** `/tmp`에 codemaps 사본 3벌을 만든다: (a) fold 토큰 1행 주입본, (b) floor 행 1개를 잃은 기록 사본, (c) 생성기 문서 1개를 잃은 사본(REQ-CFG-003의 문서 무결성 트리거). (a)(b)(c)에 대해 가드 FAIL + 메시지에 단위·문서 명명을 관측하고(E8 verbatim), 온전한 사본에서 PASS를 관측한다. live `.moai/project/codemaps/**`의 git status가 종료 시점에 변조 0임을 함께 확인한다(REQ-CFG-005).
- **M3 (Priority Low) — CI 도달성 + 종결.** `go test ./internal/graph/`가 기본 스위트 경로에서 녹색임을 확인하고(별도 워크플로 편집 없음 — AC-CFG-005), `git diff --stat`이 허용 경로(D 제약)에 한정됨을 판정하고, §E 수출과 함께 종결 보고한다.

## §G. Anti-Patterns

- **가드를 `moai graph check` 계층으로 편입** — t688이 방금 안정시킨 체커 계약 침범 + 이 카드의 프로덕션 변경 0줄 원칙 위반. standalone 테스트가 정답(REQ-CFG-004).
- **floor 핀 없이 기록만 읽기** — 기록이 조용히 파이면 가드가 공허하게 녹색이 된다(빈 스윕 녹색, verification-completeness §1.1).
- **손 열거된 단위 목록을 하드코딩해 그것만 검사** — D1의 재생산. floor는 하한이지 집합 전체가 아니다.
- **RED 입증을 live 표면 변조+복원으로 수행** — REQ-CFG-005 위반. 사고의 재연이다.
- **`core/git` 단축형을 위반으로 판정** — 접힌 결과물(부모 산문)까지 지워버리는 과잉 금지. t475 §⑥의 fold 근거와 정면 충돌한다.
- **테스트 실패를 로그로만 남기는 헬퍼** — t747 F5 / report-not-verdict 재발.

## §H. Cross-References

- spec.md §A.3(세 축 해석), §A.4(t688 합성), §D(AC 2셀) — 이 plan의 상위 근거.
- `.moai/reports/t475/codemaps-accuracy-verification.md` §⑥·§④-b — floor 집합과 사건 전수 목록의 판정 본체(동결).
- `verification-completeness.md` §1.1/§1.2/§2 — 비침묵·공허 스윕·2셀 규율.
- `internal/spec/lint_req_widen.go` — 별개 lint 축(참조만, 수리 소관 아님).
