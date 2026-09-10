---
id: SPEC-CODEMAPS-REFRESH-001
title: "plan.md — codemaps 재생성 + 정확성 검증 + provenance 재스탬프"
version: "0.1.1"
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
phase: "v3.2.0 target"
module: ".moai/project/codemaps"
tier: M
---

# plan.md — SPEC-CODEMAPS-REFRESH-001

## §A. Context

- 카드 t432 (queue intent): "codemaps 최신성 게이트 상시 적색 — 재생성 + 정확성 검증".
- 기준선(2026-09-02, 본 워크트리 @ ad272be20에서 오케스트레이터 + manager-spec 이중 실측, 일치):
  - `moai graph check`: codemaps value=60 threshold=40 verdict=stale. contribution 13 described-worthy files vs first parent f3e11e113 (t413 머지). mx-index/edges는 본 워크트리에서 verdict=absent(untracked runtime artifact — 신규 worktree 예상 상태).
  - provenance: commit_sha `7fc0af324cf4...`, generated_at 2026-08-31T06:57:41Z, described_roots [internal cmd pkg], generated_by codemaps-gen.
  - `go list ./...` = 137 패키지. 팬텀 6 디렉터리(internal/design·evaluator·factory·migrate·research·state) 전부 부재 실측.
  - 스탬프 CLI: `moai graph stamp codemaps --commit <rev>` 존재(`internal/cli/graph_stamp.go:68` 헬프가 canonical form 문서화).
- development_mode는 tdd(quality.yaml)이나 이 SPEC은 Go 프로덕션 코드 변경이 없으므로 RED-GREEN-REFACTOR 대신 **regenerate → verify(증거 수출) → stamp → gate** 의 검증 주도 절차를 따른다. "테스트"에 해당하는 것이 M2의 기계적 검증 명령과 그 증거 파일이다.
- 합격 판정의 축이 게이트가 아니라 M2 증거라는 점이 이 plan의 설계 원칙이다(REQ-CMR-008).

## §B. Known Issues

1. **팬텀 6패키지(t304 소관)** — 현재 문서가 존재하지 않는 6개 패키지를 언급. 일부는 경고 노트(`modules.md`의 design/migrate/state/research 노트, evaluator 제거 노트), 일부는 서술 잔존(`### internal/factory` 섹션). 재생성이 이를 재유도할 수 있으며, 그 경우 기록만 하고 수정하지 않는다.
2. **docs-truth.md 드리프트 사례(실측)** — §1 "정확히 11개 retained agents" vs 현재 카탈로그 12개(CLAUDE.md §4: manager-design, e2e-tester, manager-lead 추가 세대). M2 스팟체크가 잡아야 할 유형의 실제 예.
3. **스탬프 고아화 함정** — worktree 브랜치 HEAD(WT-*)를 그대로 스탬프하면 squash 머지 시 조상 단절로 CI reachability 가드 적색(SPEC-STAMP-REACHABILITY-001이 닫은 계열). `--commit` 필수.
4. **스킬 Phase 4 검증의 증거 빈틈** — `/moai codemaps`의 기존 검증 단계는 증거 수출 의무가 없어 "재생성했다"와 "재생성 결과가 맞다"가 분리되지 않는다. M2가 이 빈틈을 메운다.
5. **병렬 카드 레이스** — t413 계열 다른 카드가 described roots를 건드리면 value가 다시 오른다. run 시작 시 기준선 재측정(§C)과 종결 직전 재판독(§E)으로 흡수한다. 게이트가 advisory(blocking: false)라 다른 카드의 머지로 재적색이 되어도 본 SPEC의 종결 판정은 종결 시점 측정에 귀속된다.

## §C. Pre-flight (run 시작 시 실행)

기준선은 카드 배차 시 실측됐으나 게이트는 시간에 민감하므로, run 첫 턴에 다음을 한 배치로 재측정한다:

```bash
moai graph check                          # codemaps stale 확인 — 이미 fresh면 리드에 보고 후 M1 재판단
git rev-parse --short HEAD && git branch --show-current   # 트리·브랜치 재판독
go list ./... | wc -l                     # 137 기대 (드리프트 시 리포트)
for d in internal/design internal/evaluator internal/factory internal/migrate internal/research internal/state; do test -d "$d" && echo "EXISTS $d"; done; echo done   # 전부 무출력 기대
git merge-base HEAD origin/develop        # 스탬프 리비전 후보 확보
```

- 재측정에서 codemaps가 이미 fresh(다른 행위자가 재스탬프)면 M1~M3을 실행하지 말고 리드에 보고한다 — 불필요한 재생성은 t413 이후 내용을 다시 섞을 위험만 산다.
- `go list ./...`가 137이 아니면 값과 함께 리포트에 기록한다(차단 아님 — 문서가 따라가야 할 값이 바뀌었다는 뜻).

## §D. Constraints

- 변경 허용 경로: `.moai/project/codemaps/**`(재생성), `.moai/state/`의 스탬프 부산물, `.moai/reports/t432/**`(증거), `.moai/specs/SPEC-CODEMAPS-REFRESH-001/**`(run-phase의 progress.md §E.2 갱신 포함). 그 외 경로 변경 금지(AC-CMR-007).
- `gate.yaml`, Go 코드, t304 인용 본문: 수정 금지.
- 스탬프: `--commit "$(git merge-base HEAD origin/develop)"` 형식만 허용. bare HEAD 금지(REQ-CMR-005).
- 증거 파일: `.moai/reports/t432/codemaps-accuracy-verification.md` 단일 파일에 (a) 경로 실존 표 / (b) 패키지 대조 / (c) 식별자 hit·miss + docs-truth 스팟체크 / t304 인계 / new-findings 5개 섹션으로 수출. `/tmp` 저장 금지(감사 시점까지 경로가 살아 있어야 함).
- 재생성 실행면: `/moai codemaps --force` (스킬 `.claude/skills/moai/workflows/codemaps.md`). `--area` 부분 생성은 전체 집합 재생성이 목적과 어긋나므로 사용하지 않는다.

## §E. Self-Verification (§F 종결 시 실행)

```bash
# E1 게이트 종결 — codemaps fresh, 타 계층 stale 없음
moai graph check
# E2 스탬프 도달성 — rc=0 기대, 그리고 값이 worktree HEAD와 다름(머지 생존 리비전)을 확인.
# 먼저 SHA를 판독하고 그 출력값을 다음 명령의 <스탬프-SHA> 자리에 직접 넣어 실행한다
# (명령 치환/중괄호 그룹은 worktree 격리 가드가 거부 — 2026-09-02 스모크 실측).
jq -r .commit_sha .moai/project/codemaps/provenance.json
git merge-base --is-ancestor <스탬프-SHA> origin/develop && echo "REACHABLE rc=0"
# E3 증거 파일 존재 + 5 섹션 헤더 확인
grep -c '^## ' .moai/reports/t432/codemaps-accuracy-verification.md
# E4 범위 위생 — 레인 변경 집합이 허용 경로에 한정됨 (tracked=base ad272be20→작업트리, untracked 별도; 커밋 후에도 공허 통과 없음)
git diff --name-only ad272be20 | grep -v -E '^\.moai/(project/codemaps|reports/t432|state|specs/SPEC-CODEMAPS-REFRESH-001)/'
git status --porcelain | grep '^??' | cut -c4- | grep -v -E '^\.moai/(project/codemaps|reports/t432|state|specs/SPEC-CODEMAPS-REFRESH-001)/'
# E5 재측정 일치 — 판독 값이 보고 값과 동일한지 (측정은 명령으로)
moai graph check | grep codemaps
```

기대: E1 codemaps verdict=fresh(value < 40, 기대 0), E2 rc=0, E3 ≥ 5, E4 무출력(측정 창 = base ad272be20 대비 합집합이라 커밋 후에도 공허 통과 없음), E5 E1과 동일 값. 실패 시 해당 마일스톤으로 되돌아간다.

## §F. Milestones

판정 변경 가능성 순(내용 결정 → 검증 판독 → 기계적 절차)으로 검토하고, 실행은 아래 순서대로 M1 → M4 직렬이다.

- **M1 재생성 (High)** — `/moai codemaps --force` 실행, 6문서 전부 재생성 확인(재생성 리포트로 입증). 부분 생성·실패 시 fail-fast, M1 미종결.
- **M2 정확성 검증 (High — 카드의 중심)** — 3개 검증 항목을 기계적 명령으로 수행하고 증거 파일로 수출:
  - (a) 재생성 문서에서 `(internal|pkg|cmd)/` 경로 인용을 추출(grep -o) → 각 경로 `test -d/-f` → 전체 경로 표.
  - (b) `go list ./...` 출력과 modules.md/dependencies.md 패키지 열거 대조 → 누락/유령 불일치 전수 기록, known-6 분류.
  - (c) entry-points.md/data-flow.md 인용 식별자를 명명된 파일·패키지에서 grep → hit/miss 표. docs-truth.md §1 에이전트 카탈로그 vs `.claude/agents/` 나열 bounded 스팟체크 포함.
  - absent/miss는 전부 t304 인계(known-6) 또는 new-findings로 분류 기록. **본문 수정 없음.** 판독 결과 재생성 품질이 구조적으로 부실하다(예: 신규 유령 다수, 실재 패키지 다수 누락)면 실행을 중단하고 범위 질문으로 반환한다.
- **M3 provenance 재스탬프 (Medium)** — M2 증거가 완성된 뒤 `moai graph stamp codemaps --commit "$(git merge-base HEAD origin/develop)"` 실행. provenance.json의 commit_sha·generated_at 갱신 확인.
- **M4 게이트 종결 + 런 리포트 (Medium)** — `moai graph check` 전 계층 판독(§E1), 증거·판정을 progress.md §E.2와 `.moai/reports/t432/`에 기록.

## §G. Anti-Patterns

- **bare-HEAD 재스탬프** — worktree 브랜치 HEAD 스탬프는 squash 머지 시 고아화. REQ-GFR-014 위반이자 본 SPEC 최대 함정.
- **"게이트가 녹색"을 정확성 증거로 대체** — 카드가 금지한 실패 형태 그 자체. M2 증거 없는 M4 종결은 없다.
- **t304 흡수** — 팬텀 인용을 고치는 유혹. 기록만 하고 t304에 넘긴다(범위 규율).
- **임계값 만지기** — value가 임계값 근처라는 이유로 `codemaps_changed_files`를 올리는 것은 게이트 무력화. 보고만.
- **부분 재생성** — `--area` 등으로 일부 문서만 재생성하면 문서 집합 내부 일관성이 깨진다. 전체 재생성 원칙.
- **worktree absent를 실패로 판독** — 신규 worktree에서 mx-index/edges verdict=absent는 예상 상태다. stale과 혼동하지 않는다.
- **복합 중괄호 파이프라인** — worktree 격리 가드가 정적으로 추적할 수 없는 셸 구조(중괄호 그룹 포함)는 실행을 거부당한다(2026-09-02 E4 스모크 실측). 검증 명령은 단순 파이프라인으로 작성한다.

## §H. Cross-References

- spec.md §C REQ-CMR-001~008, acceptance.md §D AC-CMR-001~007
- SPEC-STAMP-REACHABILITY-001(스탬프 규율), SPEC-V3R6-GRAPH-FRESHNESS-001/002(게이트·REQ-GFR-014)
- `.claude/skills/moai/workflows/codemaps.md`(재생성 파이프라인), `internal/cli/graph_stamp.go`(스탬프 CLI)
- 카드 t304(팬텀 패키지 수정 소관)
