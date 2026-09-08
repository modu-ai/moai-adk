# SPEC-CLAUDELOCAL-PUSH-MODEL-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-08
tier: S
artifacts: [spec.md, plan.md, acceptance.md]
card: t531
branch: WT-claudelocal-push-model
base: bce6d7e08
premise_source: .moai/reports/t531/premise-remeasure.md
```

## §E.2 Run-phase Evidence

> 모든 부재 주장은 `/usr/bin/grep` 으로 쟀다 — 이 셸의 `grep` 은 조용히 건너뛰는 ugrep 래퍼다.
> 범위의 왼쪽 끝은 읽는 시점에 재유도했다: `CARD_BASE=$(git merge-base origin/develop HEAD)`.

### 재유도된 baseline

```
$ CARD_BASE=$(git merge-base origin/develop HEAD); echo "CARD_BASE=$CARD_BASE"
CARD_BASE=bce6d7e083208097960c88deac11c1365ad900bc

$ git rev-parse --show-toplevel; git branch --show-current; git rev-parse --short HEAD
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t531
WT-claudelocal-push-model
a833b5658
```

수리 전 `CLAUDE.local.md` 는 `git show origin/develop:CLAUDE.local.md` 와 바이트 동일(`diff -q` rc=0), 734줄이었다.

### AC-CLPM-001 — 흡수 대상 정정 (대조군 동반)

```
$ sed -n '/^### §4.1/,/^## 5\./p' CLAUDE.local.md > /tmp/s41-after.md
$ /usr/bin/grep -c 'merge origin/develop' /tmp/s41-after.md
0

$ git show "$CARD_BASE":CLAUDE.local.md | sed -n '/^### §4.1/,/^## 5\./p' | /usr/bin/grep -c 'merge origin/develop'
2

$ /usr/bin/grep -c 'merge origin/develop' CLAUDE.local.md
0
```

after 0 · base 대조군 2(≥1) → 프로브 판별력 성립. 파일 전체에서도 0.

### AC-CLPM-002 — 정본 선택 규칙

```
$ /usr/bin/grep -n '분기하는 트리' CLAUDE.local.md
13:### 0.1 [HARD] 정본은 레인이 분기하는 트리의 사본이다
15:**판별식은 「레인이 분기하는 트리가 지배한다」이고, 현재 그 트리는 `develop` 이다.** 카드 워크트리가 `develop` 에서 나오므로 `develop` 의 사본이 레인이 실제로 읽는 문서이며, 그것이 정본이다.
```

적중 2행. 15행이 `develop` 을 지배 사본으로 지목한다. 같은 §0.1 이 날짜·최신성 판별식을 명시적으로 기각한다.

### AC-CLPM-003 — ` M CLAUDE.local.md` 표식 (3팔 + 뮤턴트)

```
$ /usr/bin/grep -c 'M CLAUDE.local.md' CLAUDE.local.md
2

$ /usr/bin/grep -c 'git restore CLAUDE.local.md' CLAUDE.local.md
1

$ /usr/bin/grep -n -B2 -A2 'git restore CLAUDE.local.md' CLAUDE.local.md > /tmp/ac003-window.txt; wc -l < /tmp/ac003-window.txt
       5

$ /usr/bin/grep -cE '하지 마라|하지 않는다|금지|회귀|되돌리면' /tmp/ac003-window.txt
2

$ /usr/bin/grep -cE '실행하라|실행한다|정리하려면|하면 된다|권장' /tmp/ac003-window.txt
0
```

팔1 = 2(≥1) · 창 대조군 5행(≥1, 창이 산다) · 팔2 = 2(≥1) · 팔3 = 0.
`git restore CLAUDE.local.md` 는 파일 전체에 **정확히 1회**다 — 2회 이상이면 `grep -B2 -A2` 가 창들을 이어 붙여 팔3 이 지켜야 할 영역이 넓어진다.

#### 뮤턴트 M-003 (실행 결과 — 이 AC 의 유일한 판별 증거)

원본은 건드리지 않고 `/tmp/ac003-mutant.md` 사본에서만 금지 문면을 지시 문면으로 뒤집었다:

```
$ /usr/bin/grep -n -B2 -A2 'git restore CLAUDE.local.md' /tmp/ac003-mutant.md
33-> **[HARD] 이 표식은 정리 대상이다.**
34->
35:> 정리하려면 `git restore CLAUDE.local.md` 를 실행하라.
36->
37-> 표식이 사라지고 워킹 사본이 main 판으로 맞춰진다.

$ /usr/bin/grep -cE '하지 마라|하지 않는다|금지|회귀|되돌리면' /tmp/ac003-mutant-window.txt
0

$ /usr/bin/grep -cE '실행하라|실행한다|정리하려면|하면 된다|권장' /tmp/ac003-mutant-window.txt
1
```

뮤턴트에서 팔2 가 2→0 으로 떨어지고 팔3 이 0→1 로 올랐다 — **AC-CLPM-003 이 뮤턴트에서 FAIL 한다.**
프로브가 살아 있음이 성립한다.

절차 후 원본 무변경을 sha256 으로 확인했다:

```
$ shasum -a 256 CLAUDE.local.md   # 뮤턴트 절차 전후 동일
de5f4d0fb642832478818d44c5b61674a63d7bc1da89311509cdb36ef4e12596
```

뮤턴트 파일은 폐기했다(`rm -f /tmp/ac003-mutant.md /tmp/ac003-mutant-window.txt`; 뒤이은 `ls /tmp/ac003-mutant*` 는 `no matches found`).

### AC-CLPM-004 — main 커밋본 폐기 기록

```
$ /usr/bin/grep -n '폐기' CLAUDE.local.md | /usr/bin/grep -i 'main'
25:### 0.3 [HARD] `main` 의 커밋본은 폐기된 제3의 모델이다
27:**`main` 에 커밋돼 있는 이 파일의 사본은 폐기된 모델이며 인용 대상이 아니다.** 그 판은 「`develop` 을 원격에 올리지 않고 카드마다 `main` 으로 PR 을 낸다」는 체제를 서술하는데, 현행 체제(§4.1)와 정면으로 다르다. 다음 사람이 `main` 사본을 정본으로 집는 것이 2026-09-07 실패의 재현이다.
31:primary 체크아웃이 `main` 에 체크아웃돼 있는 동안 `git status` 는 `M CLAUDE.local.md` 를 **영구적으로, 설계대로** 보여준다. 워킹 사본이 develop 판이고 `main` 의 커밋본은 §0.3 의 폐기 모델이므로, main 대비로는 언제나 modified 로 읽힌다. 사본이 또 갈라진 것이 아니다.
```

적중 3행 — 세 행 모두 `폐기` 와 `main` 을 **한 줄 안에서** 동시에 운반한다(줄 걸침이면 이 프로브는 잡지 못한다). 27행이 main 사본을 인용 금지 대상으로 지목한다.

### AC-CLPM-005 — 레인-push 지시 미재도입 (판별 대조군 동반)

```
$ /usr/bin/grep -c 'push origin develop' CLAUDE.local.md
2

$ /usr/bin/grep -c 'push origin develop' "$BACKUP"
3
```

수리본 2건, 백업(변종 2) 대조군 3건 → 프로브가 살아 있다. 수리본의 2건은 각각:

- §4.1 레인 의무 bullet — 「리드가 창 밖에서 … 일괄로 실행하는 `git push origin develop`이며, **레인은 그 push의 주체가 아니다**」 (리드 일괄 절차를 서술하는 문맥)
- §4.1 「[HARD] 리드 develop 일괄 push」 코드 블록 안의 실제 명령

둘 다 리드 일괄 절차 블록 안이며, 이 카드는 새 출현을 추가하지 않았다(base 2 → after 2).

### AC-CLPM-006 — 미커밋 사본 인용 금지 (일반 규칙)

```
$ /usr/bin/grep -n '미커밋' CLAUDE.local.md
19:### 0.2 [HARD] 미커밋 워킹 사본은 정본으로 인용할 수 없다
21:**어느 브랜치에도 커밋된 적 없는 미커밋 워킹 사본은 정본이 아니며, 정본으로 인용될 수 없다.** 이것은 이 파일에 한정된 규칙이 아니라 인용 일반의 규칙이다 — 이력에 없는 텍스트는 다른 사람이 같은 것을 읽었는지 확인할 방법이 없고, 저자도 시점도 복구되지 않는다.
```

적중 2행. 21행이 「이 파일에 한정된 규칙이 아니라 인용 일반의 규칙」이라고 명시해 특정 파일이 아닌 인용 일반을 구속한다.

### AC-CLPM-007 — Go 코드 0 변경 (대조군 동반)

```
$ git diff --name-only "$CARD_BASE"..HEAD -- '*.go' | wc -l
       0

$ git diff --name-only "$CARD_BASE"..HEAD | wc -l
       5

$ git diff --name-only "$CARD_BASE"..HEAD
.moai/reports/t531/premise-remeasure.md
.moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/acceptance.md
.moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/plan.md
.moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/progress.md
.moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/spec.md
```

`.go` 0 · 전체 대조군 5(≥1, 범위가 산다). 커밋 범위는 plan-phase 산출물뿐이고, run-phase 편집은 아직 워킹트리에 있다(레인이 커밋한다).

```
$ git status --short   # run-phase 종료 시점 재측정
 M .moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/plan.md
 M .moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/progress.md
 M CLAUDE.local.md
```

`plan.md` 의 워킹 수정은 **이 run-phase 가 만든 것이 아니다** — 배차문이 이미 「milestones §F now as `### M1`..`### M5`」로 서술한, 배차 전 밀스톤 제목 형식 변경이다. 이 세션은 `plan.md` 를 열어 편집하지 않았다.

### AC-CLPM-008 — 백업 무결성 (분석하지 않음)

```
$ BACKUP=/Users/goos/MoAI/moai-adk-go/.moai/reports/t531/CLAUDE.local.md.primary-uncommitted-backup-20260908
$ shasum -a 256 "$BACKUP"
23f8427589b739c705c8b17def0409c187af92ce5f27ee8bb87037a8376c9a7a  /Users/goos/MoAI/moai-adk-go/.moai/reports/t531/CLAUDE.local.md.primary-uncommitted-backup-20260908

$ wc -l < "$BACKUP"
     660

$ /usr/bin/grep -c 'push origin develop' "$BACKUP"
3
```

sha256 일치 · 660줄 · AC-CLPM-005 대조군 3건. 백업은 **열거나 diff 하지 않았다** — 서명과 대조군 수치만 쟀다(`spec.md §F` 범위 밖).

### 수리 요약 (밀스톤별)

| 밀스톤 | 무엇을 바꿨나 | 대상 |
|---|---|---|
| M1 | §0.1 정본 선택 규칙(분기 트리 판별식 + 날짜/최신성 기각) · §0.2 미커밋 사본 인용 금지(일반 규칙) | `CLAUDE.local.md` 신설 §0 |
| M2 | §0.4 ` M CLAUDE.local.md` 의도된 상태 + 되돌림 명령을 금지 문맥 안에 1회 배치 | 같은 절 |
| M3 | §0.3 `main` 커밋본 = 폐기된 제3의 모델, 인용 금지 | 같은 절 |
| M4 | §4.1 흡수 대상 `git merge origin/develop` → `git merge develop`(로컬) 2곳 + 이유 1줄 | `CLAUDE.local.md` §4.1 |
| M5 | 백업 무결성 read-only 확인 + 검증 배치 실행 | 증거(이 절) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: 4af14489a
run_status: audit-ready
card: t531
branch: WT-claudelocal-push-model
card_base: bce6d7e083208097960c88deac11c1365ad900bc
head_at_measurement: a833b5658

ac_pass_count: 8
ac_fail_count: 0
ac_matrix:
  AC-CLPM-001: {status: PASS, measured: 0, control: 2, control_kind: "base §4.1 hits"}
  AC-CLPM-002: {status: PASS, measured: 2, control: n/a, control_kind: "존재형 — 대조군 없음"}
  AC-CLPM-003: {status: PASS, arm1: 2, arm2: 2, arm3: 0, window_lines: 5, mutant: FAIL-as-required}
  AC-CLPM-004: {status: PASS, measured: 3, control: n/a, control_kind: "존재형 — 대조군 없음"}
  AC-CLPM-005: {status: PASS, measured: 2, control: 3, control_kind: "백업(변종 2) 지문"}
  AC-CLPM-006: {status: PASS, measured: 2, control: n/a, control_kind: "존재형 — 대조군 없음"}
  AC-CLPM-007: {status: PASS, measured: 0, control: 5, control_kind: "범위 내 전체 변경 파일 수"}
  AC-CLPM-008: {status: PASS, measured: "sha256 일치 + 660줄", control: 3, control_kind: "백업 push origin develop 건수"}

mutant_m003:
  executed: true
  arm2_before: 2
  arm2_mutant: 0
  arm3_before: 0
  arm3_mutant: 1
  verdict: "뮤턴트에서 AC-CLPM-003 FAIL — 프로브 생존 확인"
  original_unchanged: true
  original_sha256: de5f4d0fb642832478818d44c5b61674a63d7bc1da89311509cdb36ef4e12596
  mutant_files_removed: true

zero_controls: none
  # 0 을 낸 대조군 없음. 모든 대조군이 ≥1 을 냈으므로 「측정 불가」 보고 대상 없음.

preserve_list_post_run_count: 1   # CLAUDE.local.md 한 파일만 수정
total_run_phase_files: 2          # CLAUDE.local.md + progress.md(이 파일)
go_files_changed: 0
new_warnings_or_lints_introduced: none-applicable   # Go 코드 0 변경, 문서 전용 카드
cross_platform_build:
  applicable: false
  reason: "Go 코드 0 변경 — 빌드 대상 없음"
test_suite:
  applicable: false
  reason: "Go 코드 0 변경. 전체 스위트 로컬 실행은 이 저장소에서 금지(CLAUDE.local.md §4)"
l44_pre_commit_fetch: not-performed    # 이 세션은 커밋·push 하지 않는다(레인 소관)
l44_post_push_fetch: not-performed
m1_to_mN_commit_strategy: "run-phase 편집은 워킹트리에 남긴다 — 커밋은 레인 세션이 t531 을 운반하는 메시지로 수행"

open_gaps_preserved: 4   # spec.md §G 의 4건 미폐쇄 — 이 run-phase 는 어느 것도 닫지 않았다
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-08
sync_commit_sha: b7344d957
  # backfill 완료. sync 커밋 `b7344d957`
  # (docs(SPEC-CLAUDELOCAL-PUSH-MODEL-001): sync-phase artifacts (t531)) 가 착지한 뒤
  # 후속 커밋에서 채웠다 — 커밋은 자기 해시를 인용할 수 없기 때문이다
  # (spec-frontmatter-schema.md § SHA placeholder backfill exemption). 남은 빚 없음.
sync_status: audit-ready
card: t531
branch: WT-claudelocal-push-model
head_at_sync_authoring: 132e752bf
  # sync 저작을 시작한 시점의 HEAD. 이 블록의 대부분이 이 트리에서 쓰였다.
head_at_last_edit: b7344d957
  # 마지막 편집 시점의 HEAD — 편집 직전에 재판독했다. 이 값은 sync 커밋 자신이며,
  # 위 sync_commit_sha 의 backfill 이 이 트리에서 이뤄졌다.
  # 이전 값들은 드리프트가 아니라 기록이다: §E.4 의 대부분은 `132e752bf` 에서,
  # ownership_transition_lint 블록은 증거-정정 커밋 `4362ba52f` 착지 후 그 트리에서 쓰였다.

b12_self_test_a:
  name: "CHANGELOG 중복 방지 사전 grep"
  cmd: "/usr/bin/grep -c 'CLAUDELOCAL-PUSH-MODEL' CHANGELOG.md"
  before: 0
  verdict: PASS   # 0 이므로 append 가 허용됐다. >=1 이었다면 blocker 보고로 멈췄다
b12_self_test_b:
  name: "AC 수 일치"
  cmd: "/usr/bin/grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l"
  measured: 8
  changelog_claims: 8
  source: acceptance.md   # progress.md 가 아니다
  verdict: PASS   # 0 이 아니므로 공허한 비교가 아니다
b12_self_test_c:
  name: "CHANGELOG 이 부르는 경로 실재 확인"
  paths_checked: 4
  missing: 0
  paths: [CLAUDE.local.md, .moai/reports/t531/premise-remeasure.md,
          .moai/reports/t531/run-phase-evidence.md,
          .moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/spec.md]
  verdict: PASS

changelog_entry_position: "[Unreleased] › ### Fixed 의 첫 항목 (CHANGELOG.md:375)"
changelog_language: en
  # 이 파일의 기존 항목이 전부 영문이라 파일 관례를 따랐다. 이 SPEC 의 산출물 산문은 한국어다.

frontmatter_status_transitions:
  spec.md: "in-progress → completed — status 와 updated 두 키만"
  plan.md: "해당 없음 — status: 키가 없다 (/usr/bin/grep -c '^status:' = 0)"
  acceptance.md: "해당 없음 — 같음 (0)"
  progress.md: "해당 없음 — frontmatter 블록 자체가 없다"

docs_surface:
  readme: not-applicable
  docs_site: not-applicable
  reason: |
    CLAUDE.local.md 는 배포되지 않는 유지자 전용 파일이다(§2 Local-Only).
    사용자 프로젝트에 닿지 않으므로 동기화할 사용자 대면 문서 표면이 없다.

char_budget:
  file: CLAUDE.local.md
  ceiling: 40000
  base_at_bce6d7e08: 40085
  after: 41769
  this_card_delta: 1684
  measurement: "세 좌표를 각각 재유도했다 — python3 len(read()) 문자수. wc -c(바이트)가 아니다."
  attribution: |
    base 가 이미 85자 초과였다 — 초과는 이 카드보다 앞선다.
    이 카드가 더한 1,684자는 AC-CLPM-002/003/004/006 이 문면에 있으라고 요구하는 문장들이다.
    줄이는 것은 요구를 지우는 것이지 만족시키는 것이 아니다.
  ruling: "리드 판정 2026-09-08 — 이 카드의 블로커가 아니다."

ownership_transition_lint:
  cmd: "moai spec lint .moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/spec.md"
  output: "✓ No findings — all SPEC documents are valid"
  findings: 0
  verdict: unmeasured
    # 통과가 아니다. 이 커밋들에 규칙이 적용되지 않았다는 뜻이다.
  rule_registered: |
    등록돼 있고 실행된다 — 판정 커밋의 defaultRules() 에 &OwnershipTransitionRule{} 가 있다
    (91d25bc61:internal/spec/lint.go:139).
  why_zero: |
    규칙은 커밋에 `Authored-By-Agent:` 트레일러가 없으면 **조용히** nil 을 반환한다
    (91d25bc61:internal/spec/lint_ownership.go:414):

        if rec.AuthoredByAgent == "" {
            return nil
        }

    이 브랜치의 커밋에는 그 트레일러가 하나도 없다. 그러므로 0 은
    「전이가 검사를 통과했다」가 아니라 「검사가 이 커밋들에 적용되지 않았다」이다.
  trailer_measurement:
    cmd: "git log --format='%h |%(trailers:key=Authored-By-Agent,valueonly)|' -5"
    output: "4362ba52f || · 132e752bf || · b9571a231 || · 9cf0d625f || · 4af14489a ||"
    carriers: 0
  not_the_git_unreachable_path: |
    git 미도달 경로는 Info finding 을 낸다. `moai spec lint --json` 이 `[]` 를 반환했으므로
    그 경로에 걸린 것이 아니다.
  tool_provenance: |
    좌표가 둘이다. 판정한 바이너리는 v3.2.0-rc.1 list-744-g91d25bc61 — develop `91d25bc61` 빌드이고,
    primary 체크아웃은 main `7ad9f8534` 에 있다. 두 커밋의 lint_ownership.go 블롭이 서로 다르다.
    이 트리에서 재측정한 대응:
      91d25bc61 → ef598d5c7c54bfbc88b60ee0d4fd09ab41821a3e
      7ad9f8534 → a53e9e24106a12315c7d536ab3005b4157ead2c7
    위의 file:line 인용은 판정 커밋(`91d25bc61`)의 블롭에서 읽었다.
  residual_1: |
    manager-develop 이 열어 둔 소유권 위험은 확인되지도 반증되지도 않았다 —
    이 계측기로는 미측정이다.
  residual_2: |
    이 저장소의 커밋은 `Authored-By-Agent:` 트레일러를 달지 않는다. 즉 OwnershipTransitionRule 은
    여기서 구조적으로 침묵한다. 이 카드의 범위 밖이지만, 이 규칙의 초록을 소유권 준수의 근거로
    인용하는 모든 자리가 같은 간극 위에 서 있다.

canary_compliance_check:
  applicable: false
  reason: "이 SPEC 은 자기 sync 가 시험하는 전향 정책을 정의하지 않는다."

open_gaps_preserved: 4
  # spec.md §G 의 4건. sync-phase 도 어느 것도 닫지 않았고, §A.1 의 [추론] 표지도 그대로 둔다.

sync_scope: |
  CHANGELOG.md + 이 파일 §E.4 + spec.md frontmatter, 3파일.
  CLAUDE.local.md 는 이 카드의 산출물이고 sync 에서 건드리지 않았다(diff 에 나타나지 않는다).
commit_performed: false   # 커밋·스테이징·push 는 레인 소관이다. 편집은 워킹트리에 남긴다
```

### 개정 이후 재close 재측정 (2026-09-08)

위 블록은 **첫 close** 의 기록이고 그대로 둔다. 이 블록은 제자리 개정(`0.1.0 → 0.1.1`) 이후
재close 시점에 **다시 잰** 값이다. 덧붙이는 것이지 고쳐 쓰는 것이 아니다 — §E.3 도 §E.4 의 첫
블록도 바이트 그대로다. 특히 §E.3 의 `ac_matrix` 는 `head_at_measurement: a833b5658` 을 달고
있어 그 커밋에서 참이었던 귀속된 run-phase 기록이며, 덮어쓰면 귀속된 baseline 을 파괴한다.

```yaml
re_close_at: 2026-09-08
re_close_commit_sha: pending-backfill
  # 커밋은 자기 해시를 인용할 수 없다. 레인이 다음 커밋에서 채운다
  # (spec-frontmatter-schema.md § SHA placeholder backfill exemption).
  # 비워 두지 않는 이유: SPEC 이 completed 로 읽히는 순간 아무것도 수리를 예약하지 않으므로,
  # 빈 칸은 갚아야 할 빚을 기록하지 못한다.
sync_commit_sha_note: |
  위 블록의 `sync_commit_sha: b7344d957` 은 **첫 close** 를 기록하며 그대로 둔다.
  이 재close 의 SHA 는 위 `re_close_commit_sha` 가 별도로 운반한다.
head_at_remeasurement: 35377526a
card_base: bce6d7e083208097960c88deac11c1365ad900bc
  # `git merge-base origin/develop HEAD` 로 읽는 시점에 재유도했다.

ac_matrix_remeasured:
  # 이 세션에서 직접 재실행한 값이다. §E.3 에서 옮겨 적지 않았다.
  AC-CLPM-001: {status: PASS, slice: 0, file: 0, control: 2, control_kind: "base §4.1 hits"}
  AC-CLPM-002: {status: PASS, measured: 2, control: n/a, control_kind: "존재형 — 대조군 없음"}
  AC-CLPM-003: {status: PASS, arm1: 3, arm2: 2, arm3: 0, window_lines: 5, mutant: FAIL-as-required}
    # arm1 이 2 → 3 으로 오른 것은 수리가 §2.3 에 ` M CLAUDE.local.md` 한 줄을 더한 결과다.
    # §E.3 의 `arm1: 2` 는 a833b5658 에서 참이었다 — 스테일이 아니라 다른 커밋의 기록이다.
  AC-CLPM-004: {status: PASS, measured: 3, control: n/a, control_kind: "존재형 — 대조군 없음"}
  AC-CLPM-005: {status: PASS, measured: 2, control: 3, control_kind: "백업(변종 2) 지문 — 선행 감사 관측"}
  AC-CLPM-006: {status: PASS, measured: 2, control: n/a, control_kind: "존재형 — 대조군 없음"}
  AC-CLPM-007: {status: PASS, measured: 0, control: 10, control_kind: "범위 내 전체 변경 파일 수"}
    # 대조군이 §E.3 의 5, 델타 감사의 9 와 다르다. 셋 다 맞다 — 서로 다른 HEAD 를 쟀다.
    # 5 = a833b5658(plan-phase 산출물만), 9 = 71839912e, 10 = 35377526a(verdict-delta.md 추가분).
  AC-CLPM-008: {status: unmeasured, reason: "백업 파일 접근이 이 재close 의 범위 밖으로 지정됐다"}

mutant_m003_rerun:
  executed: true
  arm2_before: 2
  arm2_mutant: 0
  arm3_before: 0
  arm3_mutant: 1
  verdict: "뮤턴트에서 AC-CLPM-003 FAIL — 프로브 생존 재확인"
  original_unchanged: true
  original_sha256_before: a0919df2c7e7e82a541838d3d607f55cf6c1d976213785260b6b2e92aeb5c5fa
  original_sha256_after: a0919df2c7e7e82a541838d3d607f55cf6c1d976213785260b6b2e92aeb5c5fa
    # 변이 전후 동일. 변이는 /tmp 사본에서만 했고 사본은 폐기했다.
  mutant_files_removed: true
  sha256_differs_from_e3: |
    §E.3 은 `de5f4d0f…` 를 기록한다. 이 값과 다른 것은 드리프트가 아니다 —
    수리 커밋이 CLAUDE.local.md 를 바꿨으므로 다른 내용의 다른 지문이다.

delta_audit_outcome:
  report: .moai/reports/t531/verdict-delta.md
  f1: CLOSED   # §2.3 문자-독해 함정 — §0.4 의 표식 한 건을 기대 baseline 으로 명시
  f2: CLOSED   # §4.1 흡수 논거 — 앞섬·뒤처짐 양방향 + 흡수 전 최신화 + §11 위임
  f3: RAISED-AND-REPAIRED
    # 개정이 형제 산출물 둘(CHANGELOG.md, progress.md)을 자기 범위에서 배제한 결함.
    # 수리 커밋 35377526a 가 spec.md §Amendments 에 재close 갱신 대상을 명시했고,
    # 이 재close 가 그 명시대로 CHANGELOG 항목과 이 블록을 현행화했다.
  f4_to_f7: 범위 밖   # 아래 residual_risk 참조

residual_risk:
  f4: "`CLAUDE.local.md:227` — §0.4 는 표식의 영구성을 「primary 가 main 에 있는 동안」으로 조건 지었는데 §2.3 은 그 조건절을 떨어뜨렸다. [Low][optional]"
  f5: "`CLAUDE.local.md:227` 끝문장 — 「다른 파일이 함께 수정돼 있으면 먼저 정리한다」가 안전한 경로를 지목하지 않는다. 문자적 독해가 primary 에서 금지된 `git restore` 로 갈 수 있다. [Low][optional]"
  f6: "`:389`(「본다」) vs `:404`(「최신화한다」) — 판정과 실행의 어긋남. 둘 다 §11 을 가리켜 실질 위험은 없다. [Low][optional]"
  f7: "`:381`(이름 포인터) vs `:389`/`:405`(번호 포인터 §11) — 같은 절을 두 형태로 가리킨다. 번호는 절 번호가 바뀌면 조용히 어긋난다. [Low][optional]"
  scope_note: |
    F4~F7 은 의도적으로 범위 밖이다. 이 넷을 흡수하려고 범위를 넓히면 F3 이 만든 함정
    — 재close 가 자기 범위를 넘어 번지는 모양 — 을 다시 짓는 것이 된다. 잔여 위험으로만 남긴다.

open_gaps_preserved: 4
  # spec.md §G 의 4건. 재close 도 어느 것도 닫지 않았고 §A.1 의 [추론] 표지도 그대로다.

re_close_scope: |
  CHANGELOG.md(항목 1건 제자리 수정) + 이 파일 §E.4(이 블록 추가) + spec.md frontmatter
  (`status: in-progress → completed`), 3파일. CLAUDE.local.md 는 산출물이고 건드리지 않았다.
commit_performed: false   # 커밋·스테이징은 레인 소관이다
```
