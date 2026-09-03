# SPEC-FMT-GATE-001 — progress (card t465)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-03
tier: M
artifacts:
  - .moai/specs/SPEC-FMT-GATE-001/spec.md
  - .moai/specs/SPEC-FMT-GATE-001/plan.md
  - .moai/specs/SPEC-FMT-GATE-001/acceptance.md
baseline:
  tree: d592b0551
  command: "gofmt -l ."
  result: "154 files (.moai/reports/t465/gofmt-l.txt)"
preceding_card:
  id: t457
  branch: WT-gofmt-drift
  tip: e1fdf00d1
lead_decisions:
  d1_binding: "make fmt-check lands in the SAME commit as gate activation (single-commit delivery; no early landing before t457 — an ownerless red)"
  d2_records_only: ".golangci.yml gofmt-linter exclusion AGREED as correct scope discipline; follow-up card candidate registered by the lead — issuance is the operator's call, this lane does not issue it"
plan_audit:
  repaired:
    - "D1 — spec.md REQ-FG-006 whole-tree predicate replaced with the tracked-variant form (aligned with acceptance.md §D.3)"
  accepted_minor_debt:
    - "D2 — tip-SHA recording form: deferred to run-phase judgment by lead instruction (seen, not missed)"
    - "D3 — re-pin path: deferred to run-phase judgment by lead instruction (seen, not missed)"
```

## §E.2 Run-phase Evidence

> 측정 트리: activation 전 `bafa7a5a3`(develop fully absorbed) → activation 후 `a95939df5`.
> 모든 명령은 본 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t465`, branch `WT-format-gate-zero`)에서 2026-09-03 실행.

### Activation 커밋

- **SHA `a95939df5`** — `feat(ci): activate gofmt format gate — ci.yml lint step + make fmt-check (card t465, SPEC-FMT-GATE-001)`
- diff stat: `.github/workflows/ci.yml | 6 ++++++` · `Makefile | 13 ++++++++++ -1` — 정확히 2 코드 파일 (리드 결정 D1: 단일 activation 커밋, `make fmt-check` + CI 스텝 동반)
- CI 스텝 배치: lint 잡 내 templ codegen drift guard **뒤**·golangci-lint **앞** (`make fmt-check` 호출 — 리드 지시의 "한 줄 공통 논리" 형태. lint 잡 runner는 `ubuntu-latest` + `actions/checkout@v7`(fetch-depth 0)라 make·git·gofmt 모두 사용 가능)

### §C Pre-flight (activation 커밋 직전, 트리 `bafa7a5a3` + 미커밋 2파일)

| 항목 | 명령 | 실측 출력 | exit |
|---|---|---|---|
| 활성 전제 1 (REQ-FG-003) | `git merge-base --is-ancestor e1fdf00d1 HEAD` | (무출력) | 0 |
| 활성 전제 2 (REQ-FG-004) | `gofmt -l . \| wc -l` | `0` | 0 |

> 주: 가드가 `test -z "$(gofmt -l .)"` 복합형을 거부해 동치의 plain 형태(`gofmt -l . | wc -l` → `0`)로 측정했다.

### AC 이진 판정

| AC | 상태 | 명령 | 실측 출력 | exit |
|---|---|---|---|---|
| AC-FG-001 (더러운 트리 → 게이트 실패) | PASS | `printf ' ' >> pkg/version/version.go && make fmt-check` | `gofmt violations found (run gofmt -w or make fmt):` / `pkg/version/version.go` / `make: *** [fmt-check] Error 1` | 2 |
| AC-FG-002 (녹색 트리 → 통과) | PASS | `make fmt-check` (clean, 2회: activation 전·후) | (무출력 — silent-on-success) | 0 / 0 |
| AC-FG-003 (t457 선행) | PASS | `git merge-base --is-ancestor e1fdf00d1 a95939df5` | (무출력) | 0 |
| AC-FG-004 (활성 시점 녹색) | PASS | `gofmt -l . \| wc -l` @ 트리 `a95939df5` | `0` | 0 |
| AC-FG-005 (배포 표면 불변) | PASS (귀속 정정 필요 — 아래 주석) | `git show --name-only 9e1b6a379 e00102f88 ce546a373 a95939df5` (무 pathspec 대조군 포함) | 4개 SPEC 커밋 전체 파일 목록에서 `internal/template/templates/**` 0건 | — |
| AC-FG-006 (로컬 패리티 `make fmt-check`) | PASS | clean: `make fmt-check` / dirty: AC-FG-001과 동일 변형 | clean 무출력 exit 0 / dirty 파일 목록 출력 exit 2 | 0 / 2 |

**Mutant probe (plan §E "명령 이진성")**: tracked 파일 `pkg/version/version.go`에 trailing space 1자 주입 → `make fmt-check` exit 2 + 해당 파일명 출력 → `git checkout -- pkg/version/version.go` 복원 → 재실행 exit 0. 게이트가 공허 초록이 아님을 실측으로 확인 (verification-completeness §1.1 observed-failure 충족).

### AC-FG-005 귀속 정정 (리드 판독 필요)

acceptance.md §D의 **문자 그대로 판정 명령** `git diff --name-only d592b0551..HEAD -- internal/template/templates/ | wc -l` → **`10`**.
이 10건은 전부 activation 이전 `bafa7a5a3`(develop 흡수 머지)을 통해 들어온 **타 카드의 develop 착지** 소관이다:
`manager-docs.md`/`plan-auditor.md`/`verification-claim-integrity.md`/`verification-completeness.md`/`doc-generation.md`/`sync-audit-4dim.js`/`manager-docs.toml`/`plan-auditor.toml`/`.git_hooks/pre-commit`/`gate.yaml`
(착지 커밋: t461 `8ab11ed99`·`8347656b8`, t447/t367/t348/t345/t302/t300 develop 머지군 — `git log bafa7a5a3 --not d592b0551 -- internal/template/templates/`로 귀속).
본 SPEC이 인도한 4개 커밋(`9e1b6a379`·`e00102f88`·`ce546a373`·`a95939df5`)의 template 경로 수는 **0** (무-pathspec 대조군으로 커밋 존재·파일 목록 확인 후 판정).
문자 그대로 AC의 base `d592b0551`이 develop 흡수 이전 값이라 범위가 이 SPEC의 인도분을 초과한다 — plan 본문 수정 없이 이 귀속 기록으로 판정한다 (base 재고정 여부는 리드/감사 판단).

### 부가 검증

- `git show --stat HEAD` → 정확히 `.github/workflows/ci.yml`, `Makefile` 2파일
- `git status --porcelain | wc -l` → `0` (activation 커밋 후 작업 트리 clean)
- ci.yml YAML 구문: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` → `YAML-OK`
- 스테이징: `git status --short` 재판독 후 explicit pathspec 2건만 (`git add .github/workflows/ci.yml Makefile`)

### 미측정 (Gaps)

- **CI Lint 잡 green 판정** — 본 레인은 push하지 않는다(레인 프로토콜 §4). 리드 일괄 develop push 후 CI 판정이 최종 근거 (acceptance §D.5 3번째 게이트).
- CI runner에서의 `make fmt-check` 실제 실행 — ubuntu-latest의 make 선태조건은 러너 이미지 문서 근거이며 CI 실행 전 미측정. develop push 후 CI가 1차 실측이 된다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: a95939df5
run_status: run-complete-pending-ci
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0        # internal/template/templates/** — SPEC 4커밋 기준 0변경 (AC-FG-005 귀속 주석 참조)
l44_pre_commit_fetch: skipped          # 레인은 push하지 않음 — 창은 리드 소관
l44_post_push_fetch: not-applicable    # 본 레인 push 없음 (리드 일괄 push 후 원격 착지 검증은 리드 몫)
new_warnings_or_lints_introduced: none-observed   # 본 SPEC 커밋은 .go 파일 0건 수정 — lint 대상 표면 미변경
cross_platform_build:
  performed: false
  reason: "코드 변경이 CI 워크플로 YAML + Makefile에 한정 — Go 소스/빌드 그래프 무변경 (plan §D: 런 검증 = 게이트 이진 판정 + CI)"
total_run_phase_files: 2               # .github/workflows/ci.yml, Makefile
m1_to_mN_commit_strategy: "M1 단일 activation 커밋(리드 결정 D1) + 문서 커밋(본 evidence) — 코드 2파일은 M1에 전량"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
