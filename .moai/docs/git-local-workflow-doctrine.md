# Local Git Workflow Doctrine (PR-mandatory 1-person OSS) — extracted from CLAUDE.local.md §23

> Maintainer-local doctrine extracted from CLAUDE.local.md to cut session-launch context (CLAUDE.local.md loads in full at every launch). The matching CLAUDE.local.md section now carries a short stub pointing here. This file is NOT loaded at launch — read it when the topic applies. Subsection numbering is preserved so existing cross-references still resolve.

## 23. Local Git Workflows + Hook Setup (PR-mandatory 1-person OSS)

> **[HARD] POLICY CHANGE (2026-07-20) — Hybrid Trunk main-direct RETIRED.** modu-ai/moai-adk main branch에 `enforce_admins: true` 적용됨 (`gh api`로 live 검증). 이제 admin 포함 **누구도 main에 직접 push 불가**. 모든 변경 (daily Tier S/M commit 포함)은 PR을 경유해야 한다. self-merge는 허용 (`required_approving_review_count: 0`) — 4개 CI status check (Test ubuntu-latest / Lint / Build linux/amd64 / CodeQL) 통과 시 리뷰어 대기 없이 본인이 머지. 종전 "모든 tier main 직진 push 허용" 정책 (2026-05-22 `cd9eead14`)은 폐기(RETIRED)되었으며, 아래 본문에서 superseded로 표시된 문장은 역사적 기록으로만 유지한다.

**변경 전 (2026-05-22 ~ 2026-07-19, RETIRED)**: commit `cd9eead14` (`chore(config)`)로 1인 OSS Hybrid Trunk policy 채택. main 직접 push 허용 + auto_branch/auto_pr 활성.

**변경 후 (2026-07-20 ~)**: `enforce_admins: true` → PR-mandatory. tier는 이제 main-direct 여부가 아니라 PR ceremony 무게(§23.9)에만 영향. 본 섹션은 정책 운영 시 마주치는 6가지 오류/경고 패턴과 처리 절차를 정리한다 (PR merge 후 로컬 동기화 패턴 A4/A5/A6는 PR-mandatory 체제에서 오히려 상시 적용).

### §23.1 pre-push hook (manual setup — local infra)

`.git/hooks/pre-push`는 git infra (local-only). Template 동기화 안 됨. 다른 머신 clone 시 수동 설치 필요.

**현재 (2026-05-22~) — Warn-only + 5s sleep**:
- main 직접 push 시 경고 출력 + 5초 대기 (Ctrl+C로 취소 가능)
- ALLOW_MAIN_PUSH=1 escape hatch 불필요 (차단 모드 폐기)
- 보호 장치 4중: pre-commit hook (enforce) + CI workflows (main push 트리거) + GitHub branch protection (4 status checks) + Conventional Commits / Release Drafter (audit)

> **[2026-07-20 주석]** `enforce_admins: true` 적용 이후 이 warn-only hook은 main direct push에 대해 **redundant** (server-side가 이미 차단). 그러나 harmless — direct push 시도는 서버에서 rejected되고, hook은 로컬에서 5초 경고만 낼 뿐 무해하므로 제거 불필요. 여전히 실수로 `git push origin main`을 시도할 때 즉각적 로컬 피드백을 준다는 부수 이점이 있다. 유지한다.

**다른 머신 manual setup**:

```bash
cat > .git/hooks/pre-push <<'EOF'
#!/bin/bash
while read local_ref local_sha remote_ref remote_sha; do
  if echo "$remote_ref" | grep -qE "refs/heads/main$"; then
    echo "⚠️  main direct push는 enforce_admins:true로 server-side 차단됨 — PR 경유 필요 (redundant 로컬 경고)" >&2
    sleep 5
  fi
done
exit 0
EOF
chmod +x .git/hooks/pre-push
```

**Hook 동작 검증** (dry-run):

```bash
echo "refs/heads/main 0000 refs/heads/main 0000" | .git/hooks/pre-push  # warn + 5s + exit 0
echo "refs/heads/feat/test 0000 refs/heads/feat/test 0000" | .git/hooks/pre-push  # silent + exit 0
```

### §23.2 GitHub branch protection 현황 (modu-ai/moai-adk main)

`gh api repos/modu-ai/moai-adk/branches/main/protection` 조회 결과 (2026-07-20 live 검증):

| 설정 | 값 | 의도 |
|------|------|------|
| `required_status_checks.strict` | `true` | up-to-date 강제 (병합 전 rebase) |
| `required_status_checks.contexts` | 4개 (Test ubuntu / Lint / Build linux/amd64 / CodeQL) | CI 보호 (PR에서 required) |
| `required_pull_request_reviews` | 활성 (0 approvals) | **PR 경유 필수** — 단, self-merge 허용 |
| `required_approving_review_count` | `0` | 1인 OSS — 리뷰어 없이 self-merge |
| `enforce_admins` | **`true`** ⭐ (2026-07-20 변경, 종전 `false`) | **admin 포함 누구도 정책 bypass 불가 → main direct push 완전 차단** |
| `allow_force_pushes` | `false` | history 보호 |
| `allow_deletions` | `false` | branch 삭제 보호 |
| `required_conversation_resolution` | `true` | PR 대화 해결 필수 |
| `required_signatures` | `false` | GPG signing 강제 안함 |

**핵심 (2026-07-20)**: `enforce_admins: true` + `required_pull_request_reviews` (0 approvals) 조합 → **모든 변경은 PR 경유, self-merge 가능**. daily Tier S/M commit도 예외 없음. 종전 `enforce_admins: false` (admin이 main 직접 push로 정책 bypass) 는 폐기.

**tag push는 branch protection 대상 아님**: `vX.Y.Z` 태그 push (`scripts/release.sh`)는 branch protection이 적용되지 않으므로 릴리스 tag flow는 이 변경의 영향을 받지 않는다.

조정 필요 시: `gh api -X PATCH repos/modu-ai/moai-adk/branches/main/protection ...`

### §23.3 운영 패턴 — A4: `gh pr merge --delete-branch` fatal

**증상**: PR admin merge 후 `fatal: Not possible to fast-forward, aborting`

**근본 원인**: gh CLI가 머지 직후 자동 `git pull --ff-only` 시도. 로컬 main이 머지된 PR squash commit과 분기되어 fast-forward 불가.

**핵심**: **실제 머지는 GitHub에서 완료된 상태** (`gh pr view <PR> --json state` → MERGED 확인). 로컬 동기화만 별도 필요.

**처리 절차**:

```bash
gh pr view <PR> --json state,mergedAt    # MERGED 확인
git fetch origin main
git reset --keep origin/main             # --hard 차단 우회 (§23.5)
```

### §23.4 운영 패턴 — A5: `git stash pop` 부분 적용 silent skip

**증상**: `git stash pop`이 일부 파일만 복원 + 나머지 파일 silent skip + "stash entry is kept in case you need it again."

**근본 원인**: stash 파일과 working tree 파일이 충돌하지 않더라도, git이 정책상 일부 적용 후 stash 보존. Silent skip은 표면화 안 됨.

**처리 절차** (명시적 복원):

```bash
git stash show --stat stash@{0}                              # 누락 진단
git checkout stash@{0} -- <missing-path-1> <missing-path-2>  # 명시 복원
git restore --staged <paths>                                 # unstage (필요 시)
git stash drop stash@{0}                                     # cleanup
```

### §23.5 운영 패턴 — A6: `git reset --hard` sandbox 자동 차단

**증상**: Claude Code sandbox에서 `git reset --hard` 명령 자동 거부 (Permission Denied)

**근본 원인**: claude-code sandbox가 destructive 명령 (`--hard`, `--force`, `rm -rf`, …)를 명시적 사용자 권한 없이 차단.

**우회 절차** (--keep equivalent + 안전):

```bash
# 1. dirty preserve
git stash push --include-untracked -m "phase-d $(date -u +%Y%m%dT%H%M%SZ)"

# 2. safe reset (--hard 대신 --keep)
git fetch origin main
git reset --keep origin/main   # local modifications 자동 보호

# 3. stash pop + 누락 명시 복원 (§23.4)
git stash pop || git checkout stash@{0} -- <paths>
```

`--keep`는 `--hard`와 달리 working tree에 commit되지 않은 변경이 있으면 reset 자체를 거부하지만, stash로 working tree가 clean한 상태에서는 `--hard`와 동등 효과.

### §23.6 운영 패턴 — Late-Branch Phase D 2중 보호

orphan commits 보존 + dirty 보존 + reset + stash pop 5단계:

```bash
git branch save-orphan-$(date +%Y-%m-%d) <latest-local-commit>             # 1) orphan 보존
git stash push --include-untracked -m "phase-d-$(date -u +%Y%m%dT%H%M%SZ)" # 2) dirty 보존
git fetch origin main                                                       # 3) 원격 최신
git reset --keep origin/main                                                # 4) 정렬 (§23.5)
git stash pop || git checkout stash@{0} -- <missing-paths>                  # 5) 복원 (§23.4)
```

선례: SPEC-V3R6-HARNESS-RENAME-001 sync (PR #1043) + chore PR #1044 (2026-05-22).

### §23.7 [HARD] 운영 원칙

- [HARD] pre-push hook은 `.git/hooks/`에 위치 — template 동기화 불가, 다른 머신 manual setup 필수
- [HARD] GitHub branch protection 변경은 `gh api -X PATCH` 명시적 수정으로만 (Settings UI 사용 시 audit trail 손실)
- [HARD] `git reset --hard` 대신 `--keep` 사용 (sandbox 안전)
- [HARD] `gh pr merge --delete-branch` 후 fatal 발생 시 `gh pr view --json state` 별도 확인 (실제 머지 여부)
- [HARD] `git stash pop` 결과는 `git status` 별도 검증 필수 (silent skip 가능성)
- [HARD] **(2026-07-20 신규) PR-mandatory: 모든 tier (S/M/L) 변경은 PR 경유.** `enforce_admins: true`로 main direct push는 admin 포함 완전 차단. self-merge 허용 (0 approvals) — 4개 CI status check 통과 시 리뷰어 대기 없이 본인 머지. tier는 main-direct 여부가 아니라 PR ceremony 무게(§23.9)에만 영향.
- [HARD] **`git push origin main` 금지** — 시도 시 server-side rejected. 항상 feat/fix/chore/docs/release 브랜치 → `gh pr create` → CI green → `gh pr merge` 흐름.
- ~~[HARD] 1-person OSS Hybrid Trunk: 모든 tier (S/M/L) main 직진 push 허용~~ **[RETIRED 2026-07-20 — enforce_admins: true 적용으로 무효]** (종전: CI 4 status checks + pre-push hook 5s warn + Conventional Commits + Release Drafter 4중 보호, §23.0 chore commit `cd9eead14`, 2026-05-22 채택. 이 문장은 역사적 기록으로만 유지.)

### §23.9 Tier-based PR Routing (REQ-ATR-020 — SPEC-V3R6-AGENT-TEAM-REBUILD-001; 2026-07-20 PR-mandatory 개정)

[HARD] **(2026-07-20 개정) 모든 tier (S/M/L)는 이제 PR을 경유한다.** `enforce_admins: true` 적용으로 main direct push가 완전히 차단되었으므로, tier는 더 이상 "main-direct vs PR" 을 가르지 않는다. tier가 결정하는 것은 **PR ceremony 무게** (브랜치 수명 / 리뷰 깊이 / 라벨 세밀도 / 풀 CI 매트릭스 여부)뿐이다. 모든 경우 PR 생성·머지는 `manager-git` 서브에이전트가 담당한다.

| Tier / 조건 | 기본 routing | Owner | PR ceremony 무게 |
|------------|-------------|-------|------|
| Tier S (< 300 LOC, < 5 files) | `fix/*`·`chore/*`·`docs/*` 등 단기 브랜치 + `gh pr create` → self-merge (0 approvals) | **manager-git** (commit은 manager-develop/manager-docs) | 경량 — 3축 라벨 최소, Tier 1 CI (4 checks) 통과 즉시 self-merge |
| Tier M (300-1000 LOC, 5-15 files) | `feat/SPEC-XXX` 브랜치 + `gh pr create` → self-merge | **manager-git** | 중간 — 3축 라벨 + PR body 설명, Tier 1 CI |
| Tier L (> 1000 LOC OR > 15 files OR constitutional) | `feat/SPEC-XXX` 브랜치 + `gh pr create` → self-merge | **manager-git** | 무거움 — Late-Branch 4-Phase + 풀 CI 매트릭스 (release PR 시 Tier 2 macOS/Windows) + 상세 리뷰 |
| Explicit `--pr` (any tier) | `feat/SPEC-XXX` 브랜치 + `gh pr create` | **manager-git** | 사용자 명시적 review round 요구 시 (cross-team review, security-sensitive change 등) — Tier 무관 무거운 ceremony 적용 |

> **[RETIRED 2026-07-20]** 종전 이 표의 Tier S/M 행은 "main 직접 push (manager-develop/manager-docs commit 직접 수행)" 이었다. `enforce_admins: true` 로 main-direct가 불가능해지면서 두 행 모두 PR routing으로 통합. `manager-develop`/`manager-docs`는 여전히 commit을 수행하되, push·PR은 `manager-git`이 담당한다 (self-merge 흐름).

**Owner 명시 (REQ-ATR-020 정합)**: 모든 tier에서 PR 생성은 `manager-git` 의 책임이다. `manager-develop` 또는 `manager-docs` 는 commit 만 수행 후 `manager-git` 에게 push + PR 생성을 위임한다. 이는 Anthropic 2026 SRP (Single Responsibility Principle) 정합 — 각 retained agent 가 명확한 phase boundary 를 가진다. (2026-07-20 이전: Tier S/M은 manager-develop이 commit+push를 자체 수행했으나, main-direct push 차단으로 이제 push도 PR 경유.)

**Late-Branch 4-Phase Pattern**: Tier L PR routing 시 `manager-git` 은 `.moai/docs/git-workflow-doctrine.md` §18.3.1 의 Late-Branch 4-Phase 패턴 (A: branch creation / B: commit / C: PR creation / D: Late-Branch closure)을 따른다. Phase D Late-Branch closure 는 PR 머지 후 local main 정렬 의무 — `.claude/agents/moai/manager-git.md` § Late-Branch Invocation Pattern 참조.

**Routing 결정 흐름 (2026-07-20 PR-mandatory)**:
1. 모든 SPEC/변경 → `manager-git` routing으로 PR 생성 (main-direct 불가)
2. SPEC tier 가 L OR 사용자가 `--pr` 명시 → 무거운 ceremony (Late-Branch 4-Phase + 풀 CI 매트릭스)
3. Tier S/M (without `--pr`) → 경량 ceremony (단기 브랜치 + Tier 1 CI 통과 즉시 self-merge) — 여전히 PR 경유이나 리뷰 오버헤드 최소
   > **[RETIRED 2026-07-20]** 종전 item 3은 "그 외 (Tier S/M without `--pr`) → main 직접 push" 였다. `enforce_admins: true`로 무효 — 모든 tier PR 경유.

상위 SPEC 참조:
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/spec.md` REQ-ATR-020 (manager-git PR doctrine reconciliation)
- `.moai/docs/git-workflow-doctrine.md` §18.3.1 [HARD] Tier-based PR Routing (SPEC-V3R6-AGENT-TEAM-REBUILD-001 REQ-ATR-020) — M5 NEW section
- `.claude/agents/moai/manager-git.md` § Late-Branch Invocation Pattern
- `.claude/skills/moai/workflows/sync.md` § Phase Owners (Tier L OR `--pr` 플래그 시 manager-git)

### §23.8 [HARD] Multi-Session Race Mitigation

동일 project root + 동일 memory hash (`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)에서 2개 이상 Claude Code 세션이 동시에 작업할 때 race condition이 빈번 발생. 메모리 공유 + git working tree 공유로 양 세션 모두 같은 paste-ready resume을 보고 같은 SPEC을 동시에 진행 시도.

**선례**: SPEC-V3R6-LEGACY-CLEANUP-001 sync-phase race (2026-05-23) — parallel session이 spec.md frontmatter status `draft → implemented`를 commit `aea0cf7b9`로 별도 push, 내 세션 manager-docs는 "이미 올바른 상태"로 감지 (`git diff` skip). 다행히 conflict 회피했으나 push range mismatch (`aea0cf7b9..19bc873ff` instead of `ccd1fa9cf..19bc873ff`)로 retrospectively 감지.

**완화 정책 4중 (Defense in Depth)**:

1. **[HARD] Pre-spawn fetch obligation**: `.claude/rules/moai/core/agent-common-protocol.md` §Pre-Spawn Sync Check (L1) — implementation Agent spawn 전 `git fetch origin && git rev-list --count --left-right origin/main...HEAD` 의무. `N 0` (origin ahead) 감지 시 STOP + AskUserQuestion (rebase / inspect / abort 3 옵션).

2. **[SHOULD] Multi-session 인지 시 L2/L3 worktree opt-in 권장**: 사용자가 동일 cwd에서 2+ 세션 작업 패턴이면 `/moai plan --worktree` 또는 `moai worktree new SPEC-XXX --base origin/main`으로 SPEC별 working tree 분리. Memory는 여전히 공유되나 git working tree는 분리 → race 원천 차단. CLAUDE.md §14 [SHOULD] worktree advisory + session-handoff.md Block 0 패턴 활용.

3. **[SHOULD] Paste-ready resume 단일 세션 처리 discipline**: 사용자 수동 규율 — paste-ready resume은 1 세션에서만 paste. 다른 세션에서는 별도 SPEC 작업 OR read-only 활동 (`Agent(Explore)` 또는 `Agent(general-purpose)` diagnostic — 과거 `manager-quality` diagnostic은 archived per SPEC-V3R6-AGENT-TEAM-REBUILD-001). Memory hash 공유로 인한 paste-ready 동시 consume 회피.

4. **[INFO] Detection signal**: `git log --oneline` mystery commit 발견 시 (예: 본인이 commit 안 한 SPEC ID commit이 main에 등장) parallel session race 정황. retrospective 감지이지만 향후 sync 전 fetch 필요성 명시.

**Multi-session pattern 판단 기준**:
- 사용자가 명시적으로 2+ terminal/IDE 띄워 사용 중 (예: 한 세션 plan / 다른 세션 review)
- `ps aux | grep claude` 또는 `tmux list-panes` 다중 결과
- mystery commit 1회 이상 발견된 경험 있음

본 정책 적용 시 §23.9 PR-mandatory routing (모든 tier PR 경유, 2026-07-20)은 single/multi-session 공통 기본값. Multi-session 시 사용자는 L2 worktree로 자발적 분리 OR SPEC별 독립 feat 브랜치로 race 원천 차단 (branch 격리가 PR-mandatory 체제에서 자연스러운 완화책). ~~§23.7 일반화 (모든 tier main 직진)는 single-session 환경 기본값~~ **[RETIRED 2026-07-20]**.

선례: SPEC-V3R6-LEGACY-CLEANUP-001 race incident (2026-05-23) + agent-common-protocol.md §Pre-Spawn Sync Check L1 정책 도입.

