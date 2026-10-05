# t363 판정 — concurrency 재키잉: 하지 않는다 (DROPPED)

- 카드: t363 · `concurrency.group` 이 `github.ref` 키라 push 런과 PR 런이 서로를 취소하지 못한다
- 레인: lane-11 · 트리: `.claude/worktrees/t363` · 브랜치: `WT-concurrency-keying`
- 기준선: `origin/develop` = `18ba3cddb` (워크트리 HEAD 동일, 측정 시점 clean)
- 판정: **DROPPED** — 카드가 스스로 정한 대가 비교 규칙을 적용한 결과, 현행 유지가 옳다

---

## Claim

1. 카드가 서술한 **기제**(push 는 `refs/heads/<b>`, PR 은 `refs/pull/N/merge` 라 `${{ github.workflow }}-${{ github.ref }}` 키가 갈린다)는 GitHub 문서 의미론상 참이다. 다만 이 저장소의 런 기록으로 재현하지는 못했다(§Gaps G1).
2. 카드가 서술한 **대가**(“릴리스 때 develop→main PR 당 중복 런 1건”)는 현행 체제에서 **발생하지 않는다**. 실측 표본에서 push/pull_request 중복 쌍 0건.
3. 중복의 전제는 “push 필터에 든 브랜치를 head 로 하는 열린 PR”이고, 그 브랜치 집합은 저장소 전체에서 `{main, develop}` 뿐이다. head=develop 인 PR 은 **2025-11-16 #228 이 마지막**이며, 현행 릴리스 체인(`develop → release/vX.Y.Z → PR → main`)은 그런 PR 을 만들지 않는다.
4. 따라서 재키잉의 대가(브랜치보호 필수 PR 런이 push 런에 취소될 위험, 비결정적)가 현행 유지의 대가(0건)보다 크다. 카드의 결정 규칙 — “재키잉의 대가가 그보다 크면 하지 않는 것이 맞다” — 이 그대로 적용된다.

## Evidence

### E1 — concurrency 블록 전수 (18개 중 13개 보유)

```
$ grep -rn -A3 '^concurrency:' .github/workflows/
```

`${{ github.workflow }}-${{ github.ref }}` 형태 6건: ci.yml:36 · graph-freshness.yml:18 · spec-lint.yml:22 ·
lsel-leak-guard.yaml:20 · template-neutrality-check.yaml:38 · test-install.yml:33.

고유 접두사 + `github.ref` 4건: release.yml:19 · codeql.yml:13 · release-pr-multi-os.yml:19 · release-drafter.yml:20.

ref 를 안 쓰는 3건: review-quality-gate.yml:13(PR 번호) · auto-merge.yml:15(PR 번호) · spec-status-auto-sync.yml:19(고정 문자열).

concurrency 부재 5건: claude.yml · community.yml · docs-i18n-check.yml · label-sync.yml · release-drafter-cleanup.yml.

### E2 — push 필터 브랜치 집합은 `{main, develop}` 뿐

```
$ grep -rn -A3 '^  push:' .github/workflows/ | grep -E 'branches|- '
```

전 워크플로의 `push.branches` 가 `[main, develop]` 또는 `[main]` 단독(label-sync). `release.yml` 은 `tags: v*` 만.
**`release/*` 는 어떤 push 필터에도 없다** — 릴리스 브랜치 push 는 push 런을 아예 일으키지 않는다. 현행 릴리스 PR(#1685, head=`release/v3.1.4`)이 중복을 못 만드는 구조적 이유가 이것이다.

### E3 — 중복 쌍 0건 (런 1000건, 6일 창)

```
$ gh run list --limit 1000 --json name,event,headSha,createdAt > /tmp/runs.json
$ jq -r '[.[].createdAt]|"sample n=\(length) window: \(min) .. \(max)"' /tmp/runs.json
sample n=1000 window: 2026-08-27T14:17:53Z .. 2026-09-02T05:38:55Z

$ jq -r 'group_by(.headSha + "|" + .name)
  | map(select((map(.event)|unique|length) > 1
        and (map(.event)|any(. == "push"))
        and (map(.event)|any(. == "pull_request"))))
  | if length == 0 then "NONE — zero push/pull_request duplicate pairs" else ... end' /tmp/runs.json
NONE — zero push/pull_request duplicate pairs
```

같은 head 에서 반복 발화한 것은 두 종뿐이고 둘 다 이벤트 구동 봇이라 중복이 아니다:
`Auto Merge` (event=workflow_run, n=36) · `Review Quality Gate` (event=check_run, n=80).

이벤트 분포(표본 1000): push 162 · check_run 80 · workflow_run 36 · pull_request 18 ·
pull_request_target 2 · schedule 1 · issue_comment 1.

### E4 — develop push 1건당 6런은 팬아웃이지 중복이 아니다

`18ba3cddb` (2026-09-02T05:28:57Z) 에서 발화한 런 6건, 전부 `event=push`, 워크플로 이름 중복 0:
CI · CodeQL · SPEC Lint · Template Neutrality Check · lsel-leak-guard · Graph Freshness.
직전 두 push(`ad272be20`, `09bf452c0`)도 같은 모양이다. 리드가 오늘 관측한 “6개 워크플로 동시 발화”는 이 팬아웃이며, 이 카드가 다루는 이중 런과는 다른 현상이다.

### E5 — head=develop 인 PR 은 9개월 반 전이 마지막

```
$ gh pr list --state all --head develop --json number,baseRefName,state,createdAt --limit 30
```

최신 항목 #228, `createdAt 2025-11-16T00:15:54Z`. 그 이후 0건.
현행 열린 PR 3건의 head 는 `release/v3.1.4`(#1685) · `dependabot/go_modules/...`(#1684) · `fix/heavy-gate-nested-toolchain`(#1681) — 어느 것도 `{main, develop}` 에 들지 않는다.
main 으로 머지된 최근 15건의 head 도 전부 `WT-*` / `chore/*` 형태다.

### E6 — 저장소는 이미 PR merge-ref 의미론을 인지하고 조치한 전례가 있다

`.github/workflows/release-drafter.yml:11` 주석: “pull_request 트리거 제거 (2026-05): GitHub Releases API 가 PR ref(refs/pull/N/merge)를 …”.
기제 인식 자체는 새롭지 않고, 이미 한 번 다른 축에서 조치됐다.

## Baseline-attribution

- 정적 측정: 워크트리 `.claude/worktrees/t363` @ `18ba3cddb`, `git status --short` 무출력(clean). E1·E2 의 grep 은 전부 이 트리에서 이번 실행.
- 런 측정: `gh run list --limit 1000` / `gh pr list` / `gh api repos/modu-ai/moai-adk/actions/runs/33548459311`, 전부 2026-09-02 이 세션에서 읽기 전용 실행. CI 를 요청하지 않았다(`gh run rerun` · workflow dispatch 0회).
- `run_attempt` 직독 표본: run 33548459311 = `{event: pull_request, head_branch: release/v3.1.4, head_sha: 10948d057…, run_attempt: 1}`. 재시도 은닉 없음.
- git-flow 전환 커밋 `11216d13f` (2026-08-27) — push 트리거가 `[main, develop]` 로 넓어진 시점.

## Gaps — 관측하지 않은 것

- **G1. 기제를 이 저장소의 런 기록으로 재현하지 못했다.** push/PR 중복 쌍이 0건이라, “키가 갈려 취소가 안 된다”를 이 트리의 관측으로 보이는 것은 원리상 불가능하다. 기제 주장은 GitHub 의 `github.ref` 문서 의미론에 근거하며 관측 근거가 아니다.
- **G2. head=develop 시절(2025-10~11)의 런은 판독하지 못했다.** Actions 런 보존 기간(기본 90일)을 넘겨 조회 대상이 아니다. 그 시기에 중복이 실제로 났는지는 미확인이다.
- **G3. 재키잉의 위험을 실측하지 않았다.** “push 런이 필수 PR 런을 취소해 cancelled 로 남긴다”는 카드가 미측정으로 남긴 그대로다. 재현하려면 head=develop PR 을 실제로 열어야 하고, 그것은 CI 를 요청하는 행위라 이 레인의 범위 밖이다(배차문 [HARD]).
- **G4. 표본 창이 6일이다.** `gh run list` 1000건 상한에 걸려 그 이전 런은 보지 않았다. 다만 E5 의 PR 목록은 전 기간을 훑었고, 중복의 전제가 되는 PR 이 9개월 반 없었다는 사실이 창의 좁음을 보완한다.
- **G5. concurrency 부재 5건은 이 카드의 범위 밖으로 두었다.** 중복 발화가 아니라 “취소 자체가 없음”이라 다른 축이다. 카드가 물려준 “PR 브랜치 필터 없는 7개” 관측도 같은 이유로 재측정하지 않았다.

## Residual-risk

- **R1. 휴면이지 소멸이 아니다.** head=`develop`(또는 `main`)인 PR 을 다시 여는 순간 E3 의 0건은 무효가 된다. 릴리스 체인이 바뀌면 이 판정도 다시 재야 한다.
- **R2. 재키잉을 해도 중복이 낭비가 아닐 수 있다.** push 런은 develop 팁을, PR 런은 `refs/pull/N/merge`(develop 을 main 에 합친 트리)를 잰다. main 이 갈라져 있으면 두 트리는 다르고, 두 번째 런은 첫 번째가 재지 않는 것을 잰다. 재키잉에 반대하는 두 번째 독립 근거이며, GitHub 의 merge-ref 의미론에 근거한 추론이지 이 트리의 관측이 아니다.
- **R3. 이 판정은 코드를 바꾸지 않았다.** 워크플로 파일 무수정이라 회귀 위험이 0인 동시에, 어떤 개선도 착지하지 않았다.

## 권고

재키잉하지 않는다. 카드를 DROPPED 로 닫되 R1 을 조건으로 남긴다 — **head 가 `main` 또는 `develop` 인 PR 을 다시 여는 체제로 돌아가면 이 판정을 재측정할 것.**
