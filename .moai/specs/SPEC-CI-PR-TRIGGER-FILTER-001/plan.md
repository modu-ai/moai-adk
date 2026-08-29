# plan.md — SPEC-CI-PR-TRIGGER-FILTER-001

## §A Context

- **브랜치/기준점**: 워크트리 `.claude/worktrees/t294`, 브랜치 `WT-freshness-trigger` @ `51daada00`
  (base = `origin/develop`). 카드 t294.
- **산출물**: `.moai/specs/SPEC-CI-PR-TRIGGER-FILTER-001/{spec,plan,progress}.md`
  (Tier S — AC는 spec.md §3 인라인, `acceptance.md` 없음. 근거: `spec-workflow.md` § SPEC Complexity Tier
  표 — Tier S = 2 files + progress.md).
- **결함 한 줄**: `.github/workflows/graph-freshness.yml` L4 `pull_request:` 가 무필터라 base가 무엇이든
  발화한다. 작성 시점(2026-08-26)에는 `branches: [main]`과 동치였고, git-flow 전환(2026-08-27)이 그
  문장의 의미만 조용히 넓혔다 — 기록된 판단 없음 = 누락.
- **조치 방향(권고)**: `pull_request:`에 `branches: [main, develop]` 명시 + 이유 주석 2~3줄. 1파일, 1블록.
- **읽기 근거**: `.moai/reports/t294/investigation.md` (조사 전문) + 본 세션 실측 5건
  (`yq` 트리거 판독 · `gh pr list` base 분포 · PR #1681 base · `actionlint` baseline · GitHub 문서 확인).

### §A.5 PRESERVE 목록 (범위 절제)

`.github/workflows/graph-freshness.yml` 의 **`on:` 블록 밖 전부** — `permissions:`, `concurrency:`,
`jobs:` 전체. 그리고 `.github/workflows/` 의 **다른 17개 파일 전부**, 특히 결정 축 A의 3종
(`lsel-leak-guard.yaml`, `spec-lint.yml`, `docs-i18n-check.yml`).

## §B Known Issues

- **B-이중발화-오귀속**: 카드 서술과 달리 `branches:` 필터는 base를 거른다(head 아님). develop→main PR의
  중복 런은 이 변경으로 닫히지 않는다 — spec.md §1.3. 이 사실을 kickoff 전에 운영자가 알아야 하며,
  모르고 승인하면 "카드가 원하던 것"이 배달되지 않는다.
- **B-실행관측-불가**: develop 동결 + 카드 브랜치 CI-inert. 실행 기반 회귀 관측 경로가 없다.
  파서 단언으로 대체하고, 관측은 유예 등록(REQ-PTF-005 / AC-PTF-006).
- **B-범위주장-반증**: 카드가 열거한 "다른 5개는 `branches: [main]` 명시" 중 `lsel-leak-guard.yaml`이
  거짓. 그 대신 `spec-lint.yml`·`docs-i18n-check.yml`이 미열거 상태로 같은 형태를 갖는다(조사 §4 표).
- **B-도구의존**: AC가 `yq`(v4)와 `actionlint`에 의존한다. 본 세션 실측으로 둘 다 PATH에 존재
  (`/opt/homebrew/bin/yq`, `/Users/goos/go/bin/actionlint`). CI에는 없을 수 있으므로 이 검증은
  **레인 로컬 전용**이며 CI 게이트로 승격하지 않는다.
- **B-충돌없음**: `e4eb15ea4`(t291 M2)가 같은 파일을 수정했으나 트리거 블록 비접촉(조사 §5).

## §C Pre-flight

```bash
git branch --show-current                      # WT-freshness-trigger
git rev-parse --short HEAD                     # 51daada00 (진행 중 재확인)
yq -o=json '.on' .github/workflows/graph-freshness.yml     # RED 전제: pull_request → null
actionlint .github/workflows/graph-freshness.yml           # baseline rc=0, 무출력 (본 세션 실측)
ls internal/template/templates/.github                     # 템플릿 미러 존재 여부 직접 확인
```

## §D Constraints

- PRESERVE 목록(§A.5) 밖 쓰기 금지.
- `--no-verify`, `--amend`, force-push 금지. Conventional Commits + 카드 id 병기(`(t294)`).
- 커밋 전 재판독: `git rev-parse --short HEAD` + `git branch --show-current` (AGENTS.md §2).
- 실행 기반 관측을 수행했다고 주장 금지 — 유예 등록만 (REQ-PTF-005).
- kickoff 결정 축 A·B가 닫히기 전 run-phase 진입 금지.

## §E Self-Verification (manager-develop 보고 양식)

E1 AC 매트릭스(AC-PTF-001~006, 명령 + 출력 전문) / E2 해당 없음(Go 빌드 무관 — 워크플로 YAML 1파일) /
E3 해당 없음(코드 미변경 → 커버리지 축 없음) / E4 서브에이전트 경계 grep(해당 없음) /
E5 `actionlint` (AC-PTF-005) / E6 커밋 SHA + push 상태 / E7 blocker / E8 RED 전문(§C의 `yq` 출력 —
구현 전 캡처본).

## §F Milestones (결정 가역성 순 — 바뀔 가능성이 큰 결정이 앞선다)

### M0 — kickoff 결정 축 (Priority High, 코드 변경 없음)

운영자가 두 축을 닫기 전에는 M1을 시작하지 않는다.

- **축 A — 범위**: graph-freshness 1개만(권고) vs paths-scoped 3종 동시 정규화.
  권고 근거: 3종은 `paths:`로 이미 좁혀져 노출 폭이 작고, 4파일 동시 변경은 diff 귀속을 흐린다.
- **축 B — 이중 발화**: 선언만 고치고 중복 런은 별도 카드로(권고) vs concurrency 키를 head SHA로 변경.
  권고 근거: head SHA 키는 push 런이 브랜치 보호가 요구하는 PR 런을 취소할 수 있고, 그 위험은
  미측정이다. 중복 런의 실비용은 릴리스 시점 develop→main PR 1회당 런 1개.
- **축 B 부산물**: 축 B를 "별도 카드"로 닫으면, 그 카드 발행(`moai todo add`)은 운영자 소관이다.

### M1 — 트리거 선언 명시 + 이유 주석 (Priority High, AC-PTF-001·002)

`.github/workflows/graph-freshness.yml` `on:` 블록:

```yaml
on:
  # main was the only pull-request base when this trigger was authored (6786c3fa4,
  # 2026-08-26), so the unfiltered form meant `branches: [main]`. The git-flow switch
  # (11216d13f) widened the same statement without a recorded decision; the filter is
  # explicit here so the firing set is declared rather than inherited.
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]
```

검증: `yq -o=json '.on.pull_request.branches'` → `["main","develop"]`;
`grep -n -B6 '^  pull_request:'` 에 주석 노출.

### M2 — 보존·비회귀 확인 (Priority High, AC-PTF-003·004·005)

- `yq -o=json '.on.push'` 불변 · `yq '.concurrency.group'` 불변 · `cancel-in-progress` 불변.
- `git diff --stat origin/develop -- .github/workflows/` → 1 파일; hunk가 `on:` 블록 안에만.
- `actionlint .github/workflows/graph-freshness.yml` rc=0 무출력.

### M3 — 유예 관측 등록 (Priority Medium, AC-PTF-006)

`progress.md` §E.2 에 `DEFERRED-OBS` 행 1개를 적는다 — develop 동결 해제 후 실행할 명령
(`gh run list --workflow=graph-freshness.yml --limit 30 --json event,headSha,headBranch`)과, 그 판독이
확인할 것(같은 headSha 에 event 가 `push`/`pull_request` 두 개로 남는지)을 이름으로 적는다.
**관측했다고 적지 않는다.**

## §G Anti-Patterns (이 카드에서 특히)

- 필터를 넣고 "이중 발화 해결"이라고 적는 것 — spec.md §1.3 이 반증한다.
- `branches: [main]`로 좁혀 PR #1677 형태(develop-base PR)의 병합 전 검사를 조용히 잃는 것.
- 결정 축 A를 스스로 넓혀 4개 워크플로를 함께 고치는 것(운영자 결정 선점).
- `yq`/`actionlint` 로컬 통과를 CI 판정으로 환산하는 것 — 판정 주체는 develop 통합 후 CI다.

## §H Cross-References

- 조사 전문: `.moai/reports/t294/investigation.md`
- 선례 주석: `.github/workflows/spec-lint.yml` L3-15, `.github/workflows/docs-i18n-check.yml`
- 관련 SPEC: SPEC-GRAPH-FRESHNESS-CADENCE-001(잡 내용·주기), SPEC-GITFLOW-DOCTRINE-ALIGN-001(git-flow 전환)
- Tier 규정: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
