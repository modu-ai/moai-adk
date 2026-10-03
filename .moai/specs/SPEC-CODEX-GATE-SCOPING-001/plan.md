# SPEC-CODEX-GATE-SCOPING-001 — plan

> status 축 무상태 아티팩트(스키마 § Artifact Statelessness). 사이클: **tdd** (quality.yaml `development_mode: tdd` + 카드 제약 4 재현 우선). Tier **M** (plan-auditor PASS 문턱 0.80).

## §A Context

- 작업 트리: 본 워크트리(브랜치 `WT-codex-gate-scope`, 기준 HEAD `2de0a2cb6` — 착수 전 재판독할 것).
- SPEC 아티팩트: `.moai/specs/SPEC-CODEX-GATE-SCOPING-001/{spec,plan,acceptance,progress,decision-index}.md`.
- 기존 인프라 — PRESERVE 대상(재사용축)과 EXTEND 대상:
  - seam: `reviewScopeResolver`(주입형 스코프 해상), `reviewGateTreeScopeReader`(tree_scope 판독기), `reviewGateChangeDetector`(자기게이트), `treeScopeSkipLogger`·`reviewGateScopeLogger`(관측 싱크) — **EXTEND: 같은 선례의 주입형 seam으로 추가**.
  - 정책 지점: `treeScopeSkipApplies`(`codex_review_tree_scope.go:85-94`) — 두 자동 경로가 공유하는 자리. primary 생략은 이 자리에 형제 정책으로 얹는다.
  - 회귀 고정: `codex_review_gate_test.go`, `codex_review_gate_wtnobase_test.go`, `codex_review_gate_wiring_test.go`, `internal/config/codex_review_gate_tree_scope_test.go`, `internal/template/review_gate_registration_test.go` — **기존 green 유지가 선행 조건**(spec §D).
- 증거 원천: primary `.moai/reports/t1395/gate-block-disposition{,-2..9}.md`(읽기 전용) + 본 트리 사본 `.moai/reports/t1404/gate-block-evidence.md`.

## §B Known Issues (도메인 필터)

- **공유 리스트 오염 함정 [최상위]**: `reviewGateRuntimePrefixes`(`codex_review_gate.go:36-43`)는 카드 스코프가 함께 쓴다(`cardChangedPaths` `codex_review_scope.go:209`, `cardScopeKeyParts` `:277`). Facet 2의 설정 표면(`.claude/settings.json`, `.moai/config/`)을 이 공유 리스트에 추가하면 카드 스코프가 변한다(REQ-CGSC-005 위반). **트리 전용 세트로 분리**한다.
- **미러 중성성**: 템플릿 미러에 기존 `t1392` 표기 6건 존재(grep 실측 2026-10-03, 양 쌍둔). pre-existing — 확장 금지. 신규 주석은 카드 id·SPEC id·날짜·SHA 무첨가(C1-C8).
- **크로스 플랫폼(B1)**: `git rev-parse --git-dir/--git-common-dir` 출력은 상대 경로일 수 있다 — 세션 디렉터리 기준으로 해상 후 비교(`filepath` 축). `GOOS=windows GOARCH=amd64 go build ./...` 필수. WCI_EXCLUDES pathspec 문법은 기존 항목과 동일 스타일이므로 이식성 등가.
- **린트 버전**: 레인 로컬 golangci-lint는 CI 판과 다를 수 있다 — 판정은 CI 추종(교훈: 레인 lint는 CI 버전).
- **RED 관측 규율**: `-v`로 `=== RUN` 관측(셀렉터 0매칭 초록 방지), 판정 출력 tail 금지, 측정은 3패키지 스코프(전체 스위트 금지). R 5건(AC-001·005·007·008·010)의 적색 테스트는 plan 단계에서 이미 저작돼 현 트리에서 적색 관측됐다 — `internal/cli/codex_review_gate_primary_scope_red_test.go`(001·005·007·008), `internal/template/hook_gate_reports_exclude_test.go`(010). M2/M3/M4의 과업은 이 테스트들을 GREEN으로 뒤집는 것이지 새로 저작하는 것이 아니며, pre-GREEN 원문은 acceptance.md §D.0 관측 장부에 있고 E8은 재확인 축이다.

## §C Pre-flight

```bash
git branch --show-current; git rev-parse --short HEAD   # WT-codex-gate-scope 기대
go test -count=1 ./internal/cli/... ./internal/config/... ./internal/template/...   # 회귀 baseline (기존 green)
grep -c -e ':(top,exclude).moai/reports' -e ':(top,exclude).moai/state' .claude/hooks/moai/sync-phase-quality-gate.sh   # 배제 항목 멤버십 baseline (1 기대 = state 항목만 적중, reports 부재의 적 — M4 후 2). 항목 기준 검사이며 grep -c 줄 수 기준 아님; 기존 항목 수 20은 배열 원문 직독(:257-266)
GOOS=windows GOARCH=amd64 go build ./...                 # 크로스 플랫폼 baseline
```

baseline 적색이 있으면 착수 전 분류(선존 결함 vs 환경)를 기록한다.

## §D Constraints

- **PRESERVE**: `handle-codex-review-gate.sh` 셸 래퍼(무변경 — 스코핑은 Go), `tree_scope` 키의 값·기본값·판독 규율, `enabled` 기본 OFF, 카드 스코프 전 경로(요청·필터·receipt 바인딩), `reviewGateRuntimePrefixes` 공유 리스트 내용, 명시적 생산자(`moai verify codex-review`) 현행 동작.
- **순서**: 템플릿 미러 → `make build` → 추적본(sync 게이트 변경 한정, 한 커밋).
- **금지**: `--no-verify`, 로컬 전체 스위트, primary 체크아웃 쓰기, `git stash`(공유 스택), 공유 prefix 리스트 변경, 런타임 관리 표면의 자동 복원.
- **커밋**: Conventional Commits + `🗿 MoAI` 트레일러 + 카드 id(t1404) 본문 기재. `draft → in-progress` 전이는 manager-develop 소관.

## §E Self-Verification (manager-develop 납품 요구)

- E1 AC 이진 매트릭스(AC-CGSC-001~012, 명령+원문 출력), E2 크로스 플랫폼 빌드 2종, E3 커버리지 3패키지 ≥85%, E4 서브에이전트 경계 grep(해당 시), E5 lint(NEW vs baseline 구분), E6 커밋 SHA 목록, E8 RED 원문 — R 5건의 pre-GREEN 실패 출력은 acceptance.md §D.0 장부에 plan 단계 관측분으로 존재하며 E8은 GREEN 전환 직전 재실행으로 그 일치를 확인한다(G 4건은 특성화 관측으로 갈음). 항목별 (a)명령 (b)관측 출력 (c)baseline 귀속 3중 귀속.

## §F Milestones (의사결정 역전 가능성 순 — 설정면이 가장 변하기 쉽고 기계 변경이 가장 낮다)

| M | 내용 | 대상 파일 | 테스트(RED-first) | 우선순위 |
|---|------|-----------|-------------------|----------|
| M1 | **설정면**: `workflow.codex.review_gate`에 primary-scope 키 추가(제안명 `primary_scope`, 값 `skip`(기본)\|`review`), `NormalizeCodexReviewGateTreeScope` 형제 노멀라이저, defaults 상수·기본값, schema 섹션 필요 시 | `internal/config/types.go`, `internal/config/defaults.go`, `internal/settings/schema_sections.go` | `codex_review_gate_tree_scope_test.go` 확장 — 노멀라이저·기본값·판독 실패(키 부재·미지 값·읽기 오류·YAML 오류) 시 **skip 적용**, 명시 `review`만 복원(spec §F.2 정책표와 동일 방향) | High |
| M2 | **Facet 1**: primary 판별기(`git rev-parse --git-dir` vs `--git-common-dir`, 주입형 seam), `resolveReviewScope` 출력 확장(REQ-CGSC-001), primary 스킵 정책(두 자동 경로 공유 지점, REQ-CGSC-002·003), 스킵 로그 basis(REQ-CGSC-011) | `internal/cli/codex_review_scope.go`, `internal/cli/codex_review_tree_scope.go`, `internal/cli/codex_review_gate.go` | `codex_review_gate_test.go`(primary 픽스처: git-dir==common-dir / 연결 워크트리 / 비git), `codex_review_gate_wtnobase_test.go`(primary+WT- 부재 조합), `codex_review_gate_wiring_test.go`(경로 공유 단언 — REQ-CRO-006 선례 패턴) | High |
| M3 | **Facet 2**: 트리 전용 런타임-설정 제외 세트(공유 리스트와 분리), 트리 자기게이트 필터 반영(REQ-CGSC-007), 리뷰 결과 재분류 경로 — 발견 전부가 런타임 표면이면 비블록+기록(REQ-CGSC-008·011) | `internal/cli/codex_review_gate.go`(신규 순수 함수 중심) | `codex_review_gate_test.go` — porcelain 픽스처(설정 표면 전용 → false), 재분류 함수 픽스처, 공유 리스트 불변 단언(AC-CGSC-009) | High |
| M4 | **Facet 3**: `WCI_EXCLUDES`에 `':(top,exclude).moai/reports'` 추가 **및 같은 pathspec을 ① 커밋 diff(:307)·② tracked diff(:308)에도 적용**(현행 ③:309·④:278만 전달 — D3 선택: 수단 확장) — 템플릿 미러 선행 → `make build` → 추적본(REQ-CGSC-009·010) | `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` → `.claude/hooks/moai/sync-phase-quality-gate.sh` | `internal/template/hook_gate_reports_exclude_test.go` GREEN 전환(RED 원문: acceptance §D.0 RED-CGSC-010) — 미추적 팔 + tracked 팔 픽스처 + 루트 대조군 포함 | Medium |
| M5 | **통합 검증**: 3패키지 테스트 + `go vet` + golangci-lint(해당 패키지) + `GOOS=windows` 빌드 + 쌍둔 패리티 + `moai spec lint` 본 SPEC | — | 전 AC 재측정 + 회귀 baseline 대비 0 신규 적색 | Medium |

> **D3 선택 기록**: REQ-CGSC-009의 수집 경로 열거를 유지하고 M4 수단을 델타 삼팔 전체로 확장했다(감사 선택지 가)·나 가운데 나) — ③ 한 팔만의 배제는 tracked/committed reports 변경을 남기고, pathspec 적용은 기존 항목과 같은 문법의 기계 확장이라 같은 한 커밋 안에서 클래스를 닫는다. AC-CGSC-010 픽스처도 미추적 팔과 tracked 팔을 함께 관측한다(acceptance §D.0 RED-CGSC-010).

## §G Anti-Patterns

- 공유 `reviewGateRuntimePrefixes`에 설정 표면을 추가하는 것(카드 스코프 오염). reviewer verdict 값만으로 스코프 AC 판정. 판정 테스트 출력 tail 절단. 템플릿 역순 수정(추적본 먼저). 카드 스코프 회귀 확인을 "나중에"로 미룸. 기계 변경(M4)을 선두에 두는 것(결정 역전 순서 위반).

## §H Cross-References

- spec.md §F.1(설계 결정) · §F.5(선행 SPEC 관계) · decision-index.md Q1-Q3
- SPEC-CODEX-GATE-SCOPE-001 · SPEC-CODEX-REVIEW-OWNERSHIP-001 · `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E
- 증거: `.moai/reports/t1404/gate-block-evidence.md` (primary 원본은 읽기 전용 인용)
