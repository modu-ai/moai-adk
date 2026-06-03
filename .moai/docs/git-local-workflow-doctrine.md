# Local Git Workflow Doctrine (Hybrid Trunk 1-person OSS) — extracted from CLAUDE.local.md §23

> Maintainer-local doctrine extracted from CLAUDE.local.md to cut session-launch context (CLAUDE.local.md loads in full at every launch). The matching CLAUDE.local.md section now carries a short stub pointing here. This file is NOT loaded at launch — read it when the topic applies. Subsection numbering is preserved so existing cross-references still resolve.

## 23. Local Git Workflows + Hook Setup (Hybrid Trunk 1-person OSS)

2026-05-22 commit `cd9eead14` (`chore(config)`)로 1인 OSS Hybrid Trunk policy 채택. main 직접 push 허용 + auto_branch/auto_pr 활성. 본 섹션은 정책 운영 시 마주치는 6가지 오류/경고 패턴과 처리 절차를 정리.

### §23.1 pre-push hook (manual setup — local infra)

`.git/hooks/pre-push`는 git infra (local-only). Template 동기화 안 됨. 다른 머신 clone 시 수동 설치 필요.

**현재 (2026-05-22~) — Warn-only + 5s sleep**:
- main 직접 push 시 경고 출력 + 5초 대기 (Ctrl+C로 취소 가능)
- ALLOW_MAIN_PUSH=1 escape hatch 불필요 (차단 모드 폐기)
- 보호 장치 4중: pre-commit hook (enforce) + CI workflows (main push 트리거) + GitHub branch protection (4 status checks) + Conventional Commits / Release Drafter (audit)

**다른 머신 manual setup**:

```bash
cat > .git/hooks/pre-push <<'EOF'
#!/bin/bash
while read local_ref local_sha remote_ref remote_sha; do
  if echo "$remote_ref" | grep -qE "refs/heads/main$"; then
    echo "⚠️  main 직접 push — Hybrid Trunk (모든 tier 허용) | CI 자동 트리거" >&2
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

`gh api repos/modu-ai/moai-adk/branches/main/protection` 조회 결과 (2026-05-22):

| 설정 | 값 | 의도 |
|------|------|------|
| `required_status_checks.strict` | `true` | up-to-date 강제 (병합 전 rebase) |
| `required_status_checks.contexts` | 4개 (Test ubuntu / Lint / Build linux/amd64 / CodeQL) | CI 보호 (main push에도 작동) |
| `required_approving_review_count` | `0` | 1인 OSS — self admin merge 허용 |
| `enforce_admins` | `false` | admin이 정책 bypass 가능 |
| `allow_force_pushes` | `false` | history 보호 |
| `allow_deletions` | `false` | branch 삭제 보호 |
| `required_conversation_resolution` | `true` | PR 대화 해결 필수 |
| `required_signatures` | `false` | GPG signing 강제 안함 |

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
- [HARD] 1-person OSS Hybrid Trunk: 모든 tier (S/M/L) main 직진 push 허용 — CI 4 status checks + pre-push hook 5s warn + Conventional Commits + Release Drafter 4중 보호 (§23.0 chore commit `cd9eead14`, 2026-05-22 채택). feat 브랜치 + 자동 PR은 사용자가 명시적으로 review round 필요하다고 결정한 경우 (예: cross-team review, security-sensitive change) opt-in으로만 사용

### §23.9 Tier-based PR Routing (REQ-ATR-020 — SPEC-V3R6-AGENT-TEAM-REBUILD-001)

[HARD] **§23.7의 "모든 tier main 직진" 일반화에 대한 Tier-based 예외 명문화.** Hybrid Trunk 1-person OSS 정책의 기본은 모든 tier (S/M/L) main 직진 push이지만, SPEC tier 가 L 이거나 사용자가 명시적으로 `--pr` 플래그를 사용한 경우 `manager-git` 서브에이전트로 PR 생성을 routing한다.

| Tier / 조건 | 기본 routing | Owner | 비고 |
|------------|-------------|-------|------|
| Tier S (< 300 LOC, < 5 files) | main 직접 push | manager-develop / manager-docs (commit 직접 수행) | Hybrid Trunk 기본 — CI 4 status checks + pre-push hook 5s warn |
| Tier M (300-1000 LOC, 5-15 files) | main 직접 push | manager-develop / manager-docs (commit 직접 수행) | Hybrid Trunk 기본 |
| Tier L (> 1000 LOC OR > 15 files OR constitutional) | `feat/SPEC-XXX` 브랜치 + `gh pr create` | **manager-git** | Tier L 또는 사용자 `--pr` 플래그 시 PR routing |
| Tier S/M + 사용자 `--pr` opt-in | `feat/SPEC-XXX` 브랜치 + `gh pr create` | **manager-git** | 사용자 명시적 review round 요구 시 (cross-team review, security-sensitive change 등) |

**Owner 명시 (REQ-ATR-020 정합)**: Tier L OR `--pr` 케이스에서 PR 생성은 `manager-git` 의 책임이다. `manager-develop` 또는 `manager-docs` 는 PR 생성을 직접 수행하지 않으며, commit 만 수행 후 `manager-git` 에게 PR 생성을 위임한다. 이는 Anthropic 2026 SRP (Single Responsibility Principle) 정합 — 각 retained agent 가 명확한 phase boundary 를 가진다.

**Late-Branch 4-Phase Pattern**: Tier L PR routing 시 `manager-git` 은 `.moai/docs/git-workflow-doctrine.md` §18.3.1 의 Late-Branch 4-Phase 패턴 (A: branch creation / B: commit / C: PR creation / D: Late-Branch closure)을 따른다. Phase D Late-Branch closure 는 PR 머지 후 local main 정렬 의무 — `.claude/agents/moai/manager-git.md` § Late-Branch Invocation Pattern 참조.

**Routing 결정 흐름**:
1. SPEC tier 가 L → `manager-git` routing (자동)
2. 사용자가 `/moai sync --pr` 또는 `/moai run --pr` 명시적 사용 → `manager-git` routing
3. 그 외 (Tier S/M without `--pr`) → main 직접 push (manager-develop/manager-docs commit 직접 수행)

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

본 정책 적용 시 §23.7 일반화 (모든 tier main 직진)는 single-session 환경 기본값. Multi-session 시 사용자는 L2 worktree로 자발적 분리 OR feat 브랜치 + PR opt-in 활용.

선례: SPEC-V3R6-LEGACY-CLEANUP-001 race incident (2026-05-23) + agent-common-protocol.md §Pre-Spawn Sync Check L1 정책 도입.

