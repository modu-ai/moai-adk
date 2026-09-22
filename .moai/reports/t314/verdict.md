# t314 — CI 가드 트리거 축 감사 (조사 단계 판정)

- card: t314
- base: origin/develop `48eb945df` (reflog: Created from origin/develop)
- tree: `.claude/worktrees/t314`, branch `WT-neutrality-guard-trigger`
- scope: `.github/workflows/**` 18개 전수, git-flow 전환(`11216d13f`) 이후 축
- landed: origin/develop `cd43dc928` (`23f12b3a0` 의 no-ff 병합)

## Claim

1. 카드가 전제한 "template-neutrality-check 가 develop 경로에서 발화하지 않는다"는 **거짓**이다.
2. 진짜 결함은 두 갈래다 — (a) 차단력이 0으로 강등된 것, (b) 실제로 develop 에서 한 번도 돌지 않은 가드가 **다른 두 개** 있다는 것.

## Evidence

### E1 — 신경성 가드는 develop push 에서 실제로 돈다

`gh run list --branch develop --workflow "Template Neutrality Check"`:

```
33066542787	48eb945df	success
33066049570	3809d1d36	success
33064065092	7ed6edb3e	success
```

`gh run view 33066542787 --json jobs`: 모든 step `success` — skipped 아님.

```
Set up job=success | checkout=success | Set up Go=success |
Run TestTemplateNeutralityAudit (isolated)=success |
Run TestTemplateNoInternalContentLeak (isolated)=success |
Run internal-content leak guard, strict tier (isolated)=success
```

공허하지 않음 — 실행 로그에 서브테스트 7개가 실제로 선택돼 통과:
`C1-macos-bias-path`, `C2-bare-narrative-v3r`, `C4-feedback-memory-ref`,
`C5-claude-local-ref`, `C6-pr-number-ref`, `C9-natural-language-canonical-form`,
그리고 `TestTemplateNeutralityAuditC8Preserve`.
(로그의 `no tests to run` 은 형제 패키지 `internal/template/agentemit` 것으로, 이름 선택자가
0개를 고른 초록이 아니다.)

### E2 — 잡기도 한다 (뮤턴트)

`internal/template/templates/CLAUDE.md` 에 C1 위반 1줄을 심고 로컬 실행:

```
--- FAIL: TestTemplateNeutralityAudit/C1-macos-bias-path (0.06s)
    TEMPLATE_NEUTRALITY_VIOLATION: class=C1-macos-bias-path file=CLAUDE.md
```

원복 후 `git status --short` 무출력.

### E3 — 그러나 아무것도 막지 않는다

```
$ gh api repos/modu-ai/moai-adk/branches/develop/protection
{"message":"Branch not protected", "status":"404"}
```

카드 PR 폐지로 레인 작업은 develop 에 직접 병합되므로, push 트리거가 도는 시점은
**이미 머지된 뒤**다. 보호도 없으니 빨간 체크가 아무도 멈추지 못한다.
사전 차단 → 사후 탐지로 강등된 상태.

단, `pull_request: branches:[main]` 필터 덕에 develop→main 릴리스 PR 에서는
main 도달 전 사전 차단이 살아 있다. 현재 열린 PR 은 0건.

### E4 — 실제로 꺼진 가드 2건

| 워크플로 | 트리거 | develop 실행 이력 | 전환 이후 변경 |
|---|---|---|---|
| `docs-i18n-check.yml` | `push: [main]` 만 / PR 은 브랜치 무필터 | **0건** | `docs-site/content` 4개 파일 |
| `spec-lint.yml` | `push` 트리거 **없음** (PR only) | **0건** | `.moai/specs` 50개 파일 |

PR 트리거는 브랜치 무필터라 형태상 넓지만, 카드 PR 이 폐지돼 **걸릴 PR 자체가 없다.**
`git rev-list --count 11216d13f..origin/develop` = 156 커밋.

### E5 — 같은 축 나머지 분류

| 형태 | 워크플로 | 판정 |
|---|---|---|
| push `[main, develop]` + PR `[main]` | `ci.yml`, `codeql.yml`, `template-neutrality-check.yaml`, `test-install.yml` | 사후 커버 있음 / 사전 차단 없음 |
| push `[main, develop]` + PR 무필터 paths | `lsel-leak-guard.yaml` | **모범 형태** |
| push `[main]` + PR 무필터 | `docs-i18n-check.yml` | 구멍 (E4) |
| PR only | `spec-lint.yml` | 구멍 (E4) |
| PR `[main]` only | `release-pr-multi-os.yml` | 설계상 정상 (릴리스 PR 은 main 대상) |
| push `[main]` only | `label-sync.yml`, `release-drafter.yml` | 정상 (레포 전역/릴리스 초안) |
| PR 무필터 + push `[main, develop]` | `graph-freshness.yml` | t294 짝 — 너무 넓음. develop 에서 현재 계속 failure |
| 이벤트 기반 (브랜치 축 없음) | `claude.yml`, `community.yml`, `auto-merge.yml`, `review-quality-gate.yml`, `release-drafter-cleanup.yml`, `release.yml` | 해당 없음 |

### E6 — 교차축 위험 1건

`spec-status-auto-sync.yml` 은 `pull_request: types:[closed]` 를 **베이스 무필터**로 받고
`contents: write` 로 `git push origin main` 을 한다. develop 대상 PR 이 닫히면
main 으로 미는 경로가 열려 있다. 실측은 아직 안 함(전환 이후 PR 0건이라 발화 이력 없음).

## Baseline-attribution

모든 수치는 이 실행에서, 이 트리(`48eb945df`)에 대고 측정. `gh run list` / `gh run view`
/ `gh api .../protection` / `git diff --name-only 11216d13f..origin/develop` / 로컬 뮤턴트 1회.

## Gaps

세 건이며, 셋의 성격이 서로 다르다. 특히 3번은 **결함이 아니라 관측 대기**다.

1. **미검증 — 릴리스 PR 필터.** develop→main 릴리스 PR 에서
   `pull_request: branches:[main]` 가 실제로 걸리는지는 **열린 PR 이 0건이라 관측하지
   못했다.** 설정 판독일 뿐 실측이 아니다.
2. **미검증 가설 — `spec-status-auto-sync` (E6).** develop 대상 PR 이 닫힐 때 실제로
   main 에 push 하는지 실행 이력으로 확인하지 못했다 (전환 이후 PR 0건이라 발화 이력 없음).
   카드 **t325** 소관.
3. **관측 대기 — 이 카드가 배선한 두 트리거의 첫 발화.** `spec-lint` 와
   `docs-i18n-check` 가 develop push 에서 실제로 도는 것을 아직 보지 못했다. 착지 push
   (`cd43dc928`) 의 diff 가 `.github/workflows/**` 와 `.moai/reports/**` 뿐이라 두
   워크플로의 `paths:` 필터에 걸리지 않았기 때문이다 — **필터가 의도대로 동작한 결과이지
   결함이 아니다.** 즉 "배선했다"까지 증명됐고 "돈다"는 아직이다.

   운영자 판정(2026-08-27): **기회주의 관측으로 간다.** `.moai/specs/**` 또는
   `docs-site/content/**` 를 건드리는 카드가 develop 에 들어가는 순간이 첫 관측 기회이며,
   그때 해당 레인이 `gh run list --branch=develop` 로 실행 이력을 확인한다.
   워크플로 자기 경로를 push paths 에 넣는 자가 발화 방식은 **채택하지 않는다** —
   워크플로 파일을 건드릴 때마다 develop 에서 린트가 도는 부작용 때문이다.

범위 밖: `graph-freshness` 의 develop 상태는 이 카드에서 조사하지 않았다 (t322/t294 소관).
착지 시점 실측으로 `52c693327`·`d2cba5e21` 에서 success, `48eb945df` 가 마지막 failure.

## 조치 (이 카드에서 실제로 바꾼 것)

`lsel-leak-guard.yaml` 의 형태를 모범으로 삼아 E4 의 구멍 2개만 막았다.

| 파일 | 변경 |
|---|---|
| `spec-lint.yml` | `push: branches:[main, develop]`, `paths: .moai/specs/**` 추가 (기존 `pull_request` 유지) |
| `docs-i18n-check.yml` | `push.branches` 에 `develop` 추가 + 스테일 주석 1줄(`push to main` → `push to main/develop`) 교정 |

`git diff --stat` = 2 files, +14 −1.

### 부작용 실측 — develop 을 빨갛게 만들지 않는다

- `go run ./cmd/moai spec lint --strict` → **rc=0**, `0 error(s), 64 warning(s)`.
  (파이프 없이 실행해 exit code 를 직접 읽었다.)
- `docs-i18n-check` 의 push 경로는 `strict=false` 이고 실행 step 이
  `continue-on-error: true` 이므로 job 이 실패로 끝나지 않는다. PR 코멘트 step 은
  `github.event_name == 'pull_request'` 로 막혀 있어 push 실행에서 발화하지 않는다.
- 두 파일 모두 YAML 파싱 확인. 파싱된 트리거:
  `spec-lint → {pull_request, push:[main, develop]}`,
  `docs-i18n-check → {pull_request, push:[main, develop], workflow_dispatch}`.

### 바꾸지 않기로 한 것 — `pull_request: branches:[main]` 은 그대로 둔다

`ci.yml` · `codeql.yml` · `template-neutrality-check.yaml` · `test-install.yml` 의
좁은 PR 필터는 **결함이 아니라 릴리스 PR 게이트**다. 카드 PR 이 폐지된 지금
이 트리거가 걸릴 유일한 자리가 develop→main 릴리스 PR 이고, 그 PR 은 main 을 대상으로
하므로 필터가 정확히 맞는다. 무필터로 넓히면 걸릴 PR 이 없어 얻는 것은 없고,
앞으로 생길 임의 브랜치 PR 마다 발화하는 노이즈만 는다.

## 남은 판단 — 리드/사용자 몫

- **E3 (차단력)**: develop 브랜치 보호를 켜서 사후 탐지를 다시 차단으로 승격할지.
  워크플로 파일이 아니라 레포 설정이라 이 카드 범위 밖이다.
- **E6 (`spec-status-auto-sync` 교차축)**: 별도 카드 권고.
- **`graph-freshness` develop 적색**: t294 소관.

## Residual-risk

- 사후 탐지가 살아 있어도 develop 이 무보호인 한 누출은 "빨간 줄이 남은 채로" 계속 쌓일 수 있다.
- 릴리스 PR 시점에 한꺼번에 터지면 원인 커밋 귀속이 156 커밋 범위로 넓어진다.
