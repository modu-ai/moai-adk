# Run Verification — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001

## AC-LSCSK-001 Verification: moai spec lint --strict

### Before Fix (main HEAD `bdcb57f8d`, installed moai binary)

```
WARNING   StatusGitConsistency  .../SPEC-UTIL-001/spec.md         1  SPEC SPEC-UTIL-001 frontmatter status 'implemented' disagrees with git-implied status 'in-progress'
WARNING   StatusGitConsistency  .../SPEC-V3R2-CON-001/spec.md     1  SPEC SPEC-V3R2-CON-001 frontmatter status 'implemented' disagrees with git-implied status 'in-progress'
WARNING   StatusGitConsistency  .../SPEC-V3R2-CON-002/spec.md     1  SPEC SPEC-V3R2-CON-002 frontmatter status 'implemented' disagrees with git-implied status 'in-progress'
WARNING   StatusGitConsistency  .../SPEC-V3R2-CON-003/spec.md     1  SPEC SPEC-V3R2-CON-003 frontmatter status 'implemented' disagrees with git-implied status 'in-progress'
WARNING   StatusGitConsistency  .../SPEC-V3R2-RT-001/spec.md      1  SPEC SPEC-V3R2-RT-001 frontmatter status 'implemented' disagrees with git-implied status 'in-progress'
WARNING   StatusGitConsistency  .../SPEC-V3R2-SPC-003/spec.md     1  SPEC SPEC-V3R2-SPC-003 frontmatter status 'implemented' disagrees with git-implied status 'in-progress'
WARNING   StatusGitConsistency  .../SPEC-V3R4-HARNESS-003/spec.md 1  SPEC SPEC-V3R4-HARNESS-003 frontmatter status 'completed' disagrees with git-implied status 'in-progress'

Total: 7 WARNINGs (StatusGitConsistency) for the 7 targeted SPECs
```

**원인**: `drift.go::getGitImpliedStatus` 가 `git log -1` 로 단일 최신 commit만 가져와, PR #930 sweep commit (`chore(spec): status drift 11건 sweep + lint-skip 등록`)을 채택 → `ClassifyPRTitle` 이 빈 status 반환 → `"in-progress"` fallback.

### After Fix (this worktree, binary built from HEAD `go build -o /tmp/moai-test ./cmd/moai`)

```
(no output for the 7 targeted SPECs)
```

**결과**: 7건 모두 0 WARNING. Walker filter가 chore(spec) commit을 건너뛰고 이전 impl/feat/sync commit의 status를 정확히 채택.

**명령 실행**:
```bash
cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001
/tmp/moai-test spec lint --strict 2>&1 | grep -E "SPEC-UTIL-001|SPEC-V3R2-CON-001|SPEC-V3R2-CON-002|SPEC-V3R2-CON-003|SPEC-V3R2-RT-001|SPEC-V3R2-SPC-003|SPEC-V3R4-HARNESS-003"
# 출력 없음 → 0 WARNING
```

### 잔존 WARNINGs (out of scope)

이 SPEC의 fix 후에도 다른 SPEC들에서 unrelated StatusGitConsistency WARNINGs가 잔존한다.
이는 본 SPEC scope 밖의 별도 status drift이며, 제거 대상이 아니다:

- SPEC-GATE-001, SPEC-SLQG-001, SPEC-UTIL-002: `in-progress` vs `implemented` drift
- SPEC-V3R2-HRN-002: `in-progress` vs `implemented` drift
- SPEC-V3R2-ORC-002, SPEC-V3R2-ORC-004, SPEC-V3R2-RT-002/003/006: `in-progress` vs `planned` drift
- SPEC-V3R4-HOOK-HARDEN-001: `in-progress` vs `implemented` drift
- SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 자체: 이 worktree에서 아직 feat commit이 없어 `plan(spec)` commit이 latest → `planned`. main 머지 후 해소될 예정.
- SPEC-V3R4-SPECLINT-DEBT-001: 동일 이유

이 WARNINGs는 lint.skip 회피책이 없는 SPECs의 genuine drift이며, 별도 status sweep 또는 sync-phase에서 처리된다.

## AC-LSCSK-002/003/004/005 Test Results

```
go test ./internal/spec/... -race -count=1
ok  github.com/modu-ai/moai-adk/internal/spec  ~4s

테스트 목록:
- TestGetGitImpliedStatus_ChoreSkip/sweep_commit_hides_real_impl         PASS  (AC-LSCSK-005a + AC-LSCSK-002)
- TestGetGitImpliedStatus_ChoreSkip/chore_precedes_feat_walker_returns_implemented  PASS  (AC-LSCSK-005b)
- TestGetGitImpliedStatus_ChoreSkip/only_chore_commits_returns_error     PASS  (AC-LSCSK-005c + AC-LSCSK-004)
- TestGetGitImpliedStatus_ChoreSkip/latest_is_real_impl_control_case     PASS  (AC-LSCSK-005d)
- TestShouldSkipCommitTitle_ChorePattern/[6 cases]                       PASS
- TestGetGitImpliedStatus_WalkerDepthBoundary                            PASS  (AC-LSCSK-004)
- TestClassifyPRTitle_ChoreSpecUnchanged/[3 cases]                       PASS  (AC-LSCSK-003)
```
