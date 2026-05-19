# Git Workflow Doctrine — Enhanced GitHub Flow

> Externalized verbatim from CLAUDE.local.md §18 on 2026-05-20 (v2.20.0-rc1 release-readiness consolidation). Original section authored v2.14.0 onward.

---


v2.14.0 릴리스 이후 공식 채택. Gitflow 대안 비교 분석 결과 (v2.14 후 세션) 이 프로젝트의 1-2명 팀 + 2일/릴리스 cadence + 단일 CLI 배포 환경에서 Gitflow의 `develop` 이중 관리 부담을 상회하는 장점이 부재. 현재 de-facto 패턴을 유지하되 **formalization + automation**으로 개선.

### §18.0 [HARD] 운영 원칙 — 5가지 즉시 개선 Framework

이 프로젝트는 **다음 5가지 개선을 공식 운영 방침으로 채택**한다. 모든 git 작업은 이 5개 축을 준수해야 한다.

| # | 개선 항목 | 현재 구현 | 참조 |
|---|----------|-----------|------|
| 1 | **Branch protection rule 강제** (`main` + `release/*`) | ⏳ `gh api` 명령어 준비 완료, admin 적용 대기 | §18.7 |
| 2 | **Label 3축 체계** (`type:` / `priority:` / `area:`) + `status:` 보조축 | ✅ `.github/labels.yml` 정의 완료, 25 labels 추가됨 | §18.6 |
| 3 | **Merge strategy 명시** (release = merge commit, feature = squash) | ✅ `.github/PULL_REQUEST_TEMPLATE.md` + `git-strategy.yaml` | §18.3 |
| 4 | **Release Drafter로 CHANGELOG 자동화** | ✅ `.github/release-drafter.yml` + workflow 구성 완료 | §18.9 |
| 5 | **Hotfix 브랜치 명명 공식화** (`hotfix/vX.Y.Z-*`) | ✅ 스크립트 `--hotfix` 플래그 + `Makefile release-hotfix` target | §18.5 |

### §18.0.1 운영 방침 (매 PR/Release마다 점검)

**모든 PR 작성자**:
1. 브랜치 명명 §18.2 준수 (11 prefix 중 선택)
2. PR에 type/priority/area 3축 label 부착 (필수)
3. PR template 내 "Merge Strategy" 체크박스 선택 (`--squash` vs `--merge`)
4. Commit 메시지 Conventional Commits 형식 (type:scope) — Release Drafter 자동 분류용

**모든 릴리스 담당자**:
1. Release Drafter draft 확인 → `CHANGELOG.md`에 영문+한국어 섹션 작성
2. `./scripts/release.sh vX.Y.Z` 또는 `make release V=vX.Y.Z` 실행 (수동 tag push 금지)
3. Hotfix는 `./scripts/release.sh vX.Y.Z --hotfix`
4. GoReleaser workflow 완료 후 GitHub Release 5 플랫폼 assets 확인
5. Post-release: docs-site 4개국어 reference sync (별도 PR) — §17 규칙 준수

**모든 리뷰어**:
1. PR의 merge strategy가 브랜치 유형과 일치하는지 확인 (release → `--merge`, feature → `--squash`)
2. 3축 label 부착 여부 확인 (type 축 최소 1, area 축 최소 1, priority 축 최소 1)
3. CI all-green 확인 (Lint / Test ubuntu/macos/windows / Build 5 / CodeQL)
4. v2.14.0 Case Study (§18.11) 재발 방지 체크 (release squash 금지, stacked PR base 전환)

**금지 사항** (§18.10 전체 위반 금지):
- ❌ `main` 직접 push
- ❌ `develop` 브랜치 생성 (Gitflow 패턴)
- ❌ Release PR squash merge (history 손실)
- ❌ `--rebase` merge 전략
- ❌ 수동 `gh release create` (GoReleaser와 충돌)
- ❌ 브랜치 명명 관례 위반

### §18.1 브랜치 구조

```
main ──●──●──●──●──●──  (protected, 항상 배포 가능, tags 부착)
       ↑  ↑     ↑
       │  │     release/vX.Y.Z (며칠 이내)
       │  │
       │  feat/SPEC-XXX-description (1-3일)
       │  fix/issue-NNN-description (1일)
       │  docs/topic-description (1일)
       │  chore/task-description (1일)
       │  plan/vX.Y-topic (SPEC 초안 모음)
       │
       hotfix/vX.Y.Z-description (main의 tag에서 분기)
```

### §18.2 [HARD] 브랜치 명명 규칙

| 접두사 | 용도 | 수명 | 예시 |
|--------|------|------|------|
| `feat/SPEC-XXX-*` | 기능 개발 (SPEC 기반) | 1-3일 | `feat/SPEC-AUTH-001-oauth2` |
| `fix/issue-NNN-*` | 이슈 기반 버그 수정 | 1일 | `fix/issue-683-race-condition` |
| `fix/*` | 이슈 없는 버그 수정 | 1일 | `fix/ci-lint-errcheck` |
| `docs/*` | 문서 작업 | 1일 | `docs/v2.14-release-notes` |
| `chore/*` | 유지보수 (의존성, cleanup) | 1일 | `chore/bump-go-1.26` |
| `ci/*` | CI/GitHub Actions 변경 | 1일 | `ci/add-windows-runner` |
| `refactor/*` | 리팩토링 (동작 불변) | 1-2일 | `refactor/validator-cleanup` |
| `release/vX.Y.Z` | 릴리스 준비 (CHANGELOG, version) | 1-3일 | `release/v2.14.0` |
| `hotfix/vX.Y.Z-*` | 프로덕션 긴급 수정 | 1일 이내 | `hotfix/v2.14.1-import-crash` |
| `plan/vX.Y-*` | SPEC draft 모음 | 1-3일 | `plan/v2.15-util-perf-backlog` |
| `audit/*` | 보안/품질 감사 | 1-3일 | `audit/askuserquestion-propagation` |

### §18.3 [HARD] Merge Strategy

각 브랜치 유형별 머지 방식 **반드시** 준수:

| 머지 유형 | 전략 | gh 명령어 | 이유 |
|-----------|------|-----------|------|
| feature/fix/chore/docs → main | **squash** | `gh pr merge N --squash` | WIP commit 정리, 1 PR = 1 main commit |
| release/* → main | **merge commit** ⭐ | `gh pr merge N --merge` | 릴리스 마일스톤 + 개별 SPEC commit 보존 |
| hotfix/* → main | **merge commit** | `gh pr merge N --merge` | 긴급 변경 이력 명확성 |
| plan/* → main | **squash** | `gh pr merge N --squash` | 초안 commit 정리 |
| dependabot/* → main | **squash** | (auto-merge) | 단순 버전 bump |

**공통 규칙**:
- [HARD] `--delete-branch=true` 기본 (머지 후 head branch 자동 삭제)
- [HARD] `--rebase`는 금지 (main history linear가 아니어도 OK, merge commit 보존 중요)
- [HARD] Force push to `main` 금지 (branch protection으로 차단)

### §18.4 Release Cadence 공식화

| 타입 | 기준 | 주기 | 브랜치 |
|------|------|------|--------|
| **Patch (vX.Y.Z)** | 버그 수정만 | 필요 시 즉시 | `fix/*` 여러 개 직접 main → tag bump |
| **Minor (vX.Y.0)** | SPEC cluster (2-4 SPECs) 또는 주요 feature | 1-2주 | `release/vX.Y.0` 경유 |
| **Major (vX.0.0)** | Breaking changes | 3-6개월 | `release/vX.0.0` + migration guide |
| **Hotfix (vX.Y.Z+1)** | 프로덕션 긴급 수정 | 24h 이내 | `hotfix/vX.Y.Z+1-*` → main |

### §18.5 Hotfix Workflow

프로덕션에서 발견된 긴급 이슈 처리:

```bash
# 1. 최신 production tag에서 분기
git fetch origin --tags
git checkout -b hotfix/v2.14.1-crash-on-startup v2.14.0

# 2. 수정 + 테스트 (로컬)
# ...

# 3. Commit + push
git commit -m "fix(hotfix): crash on startup when config missing (#NNN)"
git push -u origin hotfix/v2.14.1-crash-on-startup

# 4. PR 생성 (base: main)
gh pr create --base main --title "hotfix(v2.14.1): ..." --body "..."

# 5. CI green 확인 후 merge commit
gh pr merge <PR> --merge --delete-branch

# 6. Tag 생성 + push (로컬 릴리스 스크립트 사용)
./scripts/release.sh v2.14.1 "Hotfix release"
```

### §18.6 Label 3축 체계

모든 Issue/PR은 다음 3축 label 부착:

**축 1: type (필수)**
- `type:feature` — 새 기능
- `type:fix` — 버그 수정
- `type:docs` — 문서만 변경
- `type:chore` — 유지보수 (의존성, 빌드 등)
- `type:ci` — CI/CD 변경
- `type:refactor` — 리팩토링
- `type:security` — 보안 이슈/수정
- `type:test` — 테스트 추가/개선

**축 2: priority (필수)**
- `priority:P0` — Critical (즉시 대응)
- `priority:P1` — High (당일)
- `priority:P2` — Medium (이번 sprint)
- `priority:P3` — Low (backlog)
- `priority:P4` — Icebox (장기 보류)

**축 3: area (필수)**
- `area:mx` — MX validator
- `area:astgrep` — ast-grep integration
- `area:lsp` — LSP subsystem
- `area:hooks` — Hook system
- `area:docs-site` — docs-site (Hugo)
- `area:templates` — internal/template/templates/
- `area:cli` — CLI 명령어
- `area:config` — .moai/config/*
- `area:workflow` — SPEC workflow
- `area:ci` — GitHub Actions
- `area:security` — 보안 관련
- `area:deps` — 외부 의존성

**축 4: status (선택)** — 진행 상태 추적
- `status:in-progress` · `status:review` · `status:blocked` · `status:needs-info`

### §18.7 Branch Protection Rule (GitHub)

> **[Updated 2026-05-17 — SPEC-V3R4-CI-FASTTRACK-001]** Required status checks reduced from 6 to 4 items.
> (B) 결정 rationale: 1인 개발, macOS 환경, 5-6분+ PR wait 비현실적.
> `Test (macos-latest)` / `Test (windows-latest)` 제거. Tier 2 릴리즈 PR 풀 매트릭스로 이전.

#### 3-Tier CI Philosophy (SPEC-V3R4-CI-FASTTRACK-001)

| Tier | 언제 | 대상 | 상태 |
|------|------|------|------|
| **Tier 1 (per-PR fast)** | 모든 PR | `ubuntu-latest` Go test + Lint + `linux/amd64` Build + CodeQL | **Required** (branch protection) |
| **Tier 2 (release PR 풀 매트릭스)** | `release/*` branch PR + workflow_dispatch | macOS + Windows + ubuntu 풀 매트릭스 (`.github/workflows/release-pr-multi-os.yml`) | Informational (NOT required) |
| **Tier 3 (수동 override)** | `workflow_dispatch` | release-pr-multi-os.yml 수동 트리거 | Informational |

**Tier 2 trigger 근거 (user directive 2026-05-17)**: "macOS/Windows 검증은 릴리즈 PR 때 처리르 하는게 맞지 않을까?" — release branch PR 시 회귀가 just-in-time 으로 가시화되어 동일 PR에서 수정 가능. nightly cron (async) / release tag (post-merge) 보다 우수.

**Required status checks (4 items, post-(B) baseline)**:
- `Lint`
- `Test (ubuntu-latest)`
- `Build (linux/amd64)`
- `CodeQL`

**Removed from required checks (2026-05-17 (B) 결정)**:
- `Test (macos-latest)` — Tier 2 (release PR) 로 이전
- `Test (windows-latest)` — Tier 2 (release PR) 로 이전

[HARD] `main` 브랜치 보호 설정 (admin 실행 필요):

```bash
gh api -X PUT /repos/modu-ai/moai-adk/branches/main/protection \
  --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["Lint", "Test (ubuntu-latest)", "Build (linux/amd64)", "CodeQL"]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_linear_history": false,
  "required_conversation_resolution": true
}
EOF
```

[HARD] `release/*` 패턴 보호 (feature freeze 기간 안정성):

```bash
gh api -X PUT /repos/modu-ai/moai-adk/branches/release%2F*/protection \
  --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["Lint", "Test (ubuntu-latest)", "Build (linux/amd64)"]
  },
  "allow_force_pushes": false,
  "allow_deletions": false
}
EOF
```

**Cross-references**: `feedback_worktree_autonomous`, lessons #18/#19.

#### Local pre-push verification (lefthook)

[HARD] Tier 1 fast-track 패턴을 보완하기 위해 push 전 로컬 검증을 강제. `lefthook.yml` 이 repo root 에 존재하며 `pre-push` hook 에서 `make preflight` 를 실행 (golangci-lint --fast + go test -race -short + go build). 신규 메인테이너 1회 설치:

```bash
brew install lefthook
lefthook install      # .git/hooks/pre-push 자동 등록
```

검증: `make preflight` 직접 실행 시 1-2분 내 통과 → push 후 CI 실패율 95%+ 감소. 임시 우회 (긴급 hotfix 등): `LEFTHOOK=0 git push ...`. SPEC-V3R4-CI-FASTTRACK-001 REQ-CIFT-005 + lessons #19 (1인 개발 CI 3-tier 패턴).

### §18.8 Release 프로세스

[HARD] 릴리스는 `scripts/release.sh` 스크립트 경유. 수동 tag push 금지 (GoReleaser 자동 릴리스와 충돌 가능).

**Minor/Major Release (release 브랜치 경유)**:

```bash
# 1. release 브랜치 생성 + 작업
git checkout -b release/v2.15.0 main
# CHANGELOG.md, SPEC status 업데이트, docs-site update 등

# 2. PR + review + merge (merge commit 전략)
gh pr create --base main --title "release: v2.15.0" --body "..."
gh pr merge <PR> --merge --delete-branch

# 3. Tag + release (로컬 스크립트)
git checkout main && git pull
./scripts/release.sh v2.15.0
# → tag 생성, push, GoReleaser 자동 실행, GitHub Release 생성
```

**Patch Release (fix 직접 main)**:

```bash
# 1. fix 브랜치에서 수정 + PR + squash merge
# 2. main pull + tag + push (release 스크립트)
./scripts/release.sh v2.14.1
```

### §18.9 자동화 도구

**구성 완료된 도구 (Enhanced GitHub Flow 공식 인프라)**:

| 도구 | 역할 | 트리거 | 설정 파일 |
|------|------|--------|-----------|
| **GoReleaser** | Tag push 시 5 플랫폼 바이너리 빌드 + GitHub Release 생성 | `push: tags: v*` | `.goreleaser.yml` + `.github/workflows/release.yml` |
| **Release Drafter** ⭐ | PR merge 시 next release draft 자동 업데이트 (CHANGELOG 작성 보조) | `push: branches: main`, `pull_request: [opened, synchronize, edited]` | `.github/release-drafter.yml` + `.github/workflows/release-drafter.yml` |
| **Dependabot** | Go modules + GitHub Actions 자동 버전 업데이트 PR | 주간 | `.github/dependabot.yml` |
| **Auto-merge** | Dependabot PR CI pass 시 자동 머지 (squash) | PR + CI success | `.github/workflows/auto-merge.yml` |
| **Labeler** | PR 파일 패턴 기반 자동 라벨 부착 (area 축 추론) | PR opened/synchronized | `.github/labeler.yml` |
| **Release Drafter autolabeler** | PR title/branch/files 기반 type 축 자동 라벨 | 위 Release Drafter와 통합 | `.github/release-drafter.yml` § `autolabeler` |

**Release Drafter ↔ GoReleaser 역할 분담** (공존 설계):
- **Release Drafter** = "다음 릴리스 preview" 자동 축적 → 인간이 `CHANGELOG.md`에 영문+한국어로 정제
- **GoReleaser** = tag push 시 final release 생성 (commit 기반 changelog 포함). Release Drafter draft와 무관하게 tag 기준 독립 운영
- **실제 workflow**: PR merge → Release Drafter가 draft 축적 → 릴리스 시 draft 확인 → `CHANGELOG.md` 업데이트 → `./scripts/release.sh` → GoReleaser final release

**Release Drafter Version Resolver** (자동 SemVer 추정):
- `breaking` / `type:breaking` label → major bump
- `type:feature` label → minor bump
- `type:fix` / `type:security` / `type:performance` / 기타 → patch bump

**추가 도구 (v2.15+ 검토)**:
- **EndBug/label-sync**: `.github/labels.yml`을 GitHub과 주기적 동기화 (현재 수동)
- **Commitlint GitHub Action**: Conventional Commits 형식 CI 강제
- **Changesets (대안)**: CHANGELOG 관리 자동화 심화 — 현재 Release Drafter로 충분

### §18.10 [HARD] 공식 위반 금지 사항

- ❌ `main`에 직접 push (반드시 PR 경유)
- ❌ `develop` 브랜치 생성 (Gitflow 패턴 금지)
- ❌ release PR squash merge (merge commit 필수 — 개별 SPEC commit 보존)
- ❌ tag 수동 push 없이 `gh release create` (GoReleaser와 충돌)
- ❌ branch prefix 관례 위반 (`feat/SPEC-*` 대신 `add-something` 등)
- ❌ `--rebase` merge 전략 (history rewrite 방지)
- ❌ CI fail 상태 PR merge (admin override 필요, 최소화)

### §18.11 이전 세션 학습 (v2.14.0 Case Study)

v2.14.0 릴리스 과정에서 다음 문제 발생 → v2.15부터 방지:

1. **Squash merge로 release history 손실**
   - PR #703 (`release/v2.14.0 → main`)을 squash로 머지하여 9개 의미 있는 commit (UTIL-001/002/003 + review fix + CI fix + ETXTBSY)이 단일 commit으로 압축됨
   - **교훈**: release → main은 항상 `--merge` 사용

2. **Stacked PR close 후 reopen 실패**
   - PR #704 (base: release/v2.14.0)가 PR #703 merge 과정에서 auto-close됨. `gh pr reopen` 실패 → 새 PR #705 생성으로 우회
   - **교훈**: stacked PR은 parent merge 전 base를 main으로 미리 전환하는 것이 안전

3. **CI Ubuntu 환경 의존성 누락**
   - `sg` command가 Ubuntu의 `util-linux` `newgrp` symlink와 충돌
   - `TestRuleSeed` skip 로직이 `LookPath`만으로 판정하여 false positive
   - **교훈**: 환경 의존 테스트는 **command signature 검증** 필수 (예: `sg --version` 출력의 "ast-grep" 확인)

4. **Pre-existing flaky test가 릴리스 block**
   - `TestSupervisor_NonZeroExit`의 ETXTBSY race가 Ubuntu CI에서 간헐 실패
   - **교훈**: flaky 테스트는 CLAUDE.local.md에 기록 + retry 로직 즉시 적용 (ETXTBSY mitigation pattern 참조)

---

### §18.12 Branch Origin Decision Protocol (BODP)

SPEC-V3R3-CI-AUTONOMY-001 Wave 7 (T8) 도입 — 신규 SPEC plan 또는 worktree 생성 시 base branch 결정을 표준화. 새 슬래시/CLI 명령어 ZERO 원칙 — 기존 3개 entry point에 BODP 게이트만 삽입.

#### 알고리즘 (3-Signal Evaluation)

`internal/bodp/relatedness.go` `Check()` 함수가 다음 3개 시그널을 평가한다:

| 시그널 | 출처 | 의미 |
|-------|------|------|
| Signal A | SPEC frontmatter `depends_on` 매칭 + diff path overlap | 코드 의존성 |
| Signal B | `git status --porcelain` 에서 `.moai/specs/<NewSpecID>/` 매칭 | 작업 트리 co-location |
| Signal C | `gh pr list --head <currentBranch> --state open` ≥ 1 | 현재 브랜치 open PR head |

#### Decision Matrix (8 rows)

`internal/bodp/relatedness.go` `applyMatrix()` — SignalB 우선순위 dominates A/C:

| ¬a ¬b ¬c | → main      @ origin/main |
| a  ¬b ¬c | → stacked   @ currentBranch |
| ¬a  b ¬c | → continue  @ "" |
| ¬a ¬b  c | → stacked   @ currentBranch |
| a   b ¬c | → continue  @ "" (b dominates) |
| a  ¬b  c | → stacked   @ currentBranch |
| ¬a  b  c | → continue  @ "" (b dominates) |
| a   b  c | → continue  @ "" (b dominates) |

SignalC positive 시 Rationale 에 `parent-merge gotcha 주의: §18.11 Case Study 참조` suffix 자동 추가 (REQ-CIAUT-047b).

#### 3 Invocation Paths (Verbatim)

1. **`/moai plan --branch`** (skill body) — Phase 3.0 BODP Gate → AskUserQuestion → manager-git delegation with `base=<chosenBase>`. Skill: `.claude/skills/moai/workflows/plan.md` Phase 3.0.

2. **`/moai plan --worktree`** (skill body) — Phase 3.0 BODP Gate → AskUserQuestion → `moai worktree new <SPEC-ID> --base <chosenBase>`.

3. **`moai worktree new <SPEC-ID>`** (CLI) — `--base` (default `origin/main`) + `--from-current` (HEAD) flags. **AskUserQuestion 호출 절대 금지** (orchestrator-only HARD per agent-common-protocol). signal collection 만 수행 후 audit trail 기록.

#### Audit Trail

`.moai/branches/decisions/<normalized-branch-name>.md` (slash → dash 정규화). 각 BODP 결정마다 markdown frontmatter + body (Signals/Decision/Executed sections) 기록.

#### Off-Protocol Reminder

`moai status` 끝에서 `internal/cli/status.go` `emitOffProtocolReminder()` 호출. 4-skip-condition: `MOAI_NO_BODP_REMINDER=1` env, main/master branch, audit trail 존재, audit dir 부재 (false-positive 방지).

#### Out of Scope (Wave 7)

- Audit trail concurrent write race (orchestrator 단일 세션 가정).
- Reminder 빈도 제한 / 자동 mute pattern (env var 1회 비활성화만).
- Untracked content 복구 (W5 동일 — 경로만 기록).
- Lint custom rule for AskUserQuestion 정적 검사 (단일 import grep 만).

---

