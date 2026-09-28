# t1292 — MoAI L1 워크트리 경로 전환

## Claim

MoAI가 만드는 L1 워크트리의 위치를 `.moai/worktrees`로 바꾸고, `moai update`에 기존 Git 등록 트리의 안전한 이전 경로를 추가했다. Codex `-w`는 기존 트리만 연다. 코드와 테스트는 `develop`에 병합돼 CI가 통과했다.

## Evidence

```text
$ go test ./internal/cli ./internal/cli/worktree -run '^(TestLegacyWorktreeMigration|TestRunUpdateMigratesRegisteredLegacyWorktree|TestUpdateWorktreeMigration|TestWorktreeNew_|TestCodexWorktree|TestNew_)' -count=1 -timeout 180s
ok  	github.com/modu-ai/moai-adk/internal/cli	15.195s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	0.381s
$ gh run view 36367483094 --json headSha,status,conclusion,jobs --jq '{headSha,status,conclusion,failed:[.jobs[]|select(.conclusion=="failure")|.name]}'
{"conclusion":"success","failed":[],"headSha":"cee197917b83aff5b04761e13d5a6328179cfee1","status":"completed"}
```

통합 커밋: `a7190891d`(기능), `2dbf4321b`(CI 테스트 경로 정정). 테스트는 잠금·활성 세션·프로세스 CWD·대상 충돌을 건너뛰고, 수정·미추적·무시 파일 및 Git HEAD·브랜치·등록 경로를 보존하는 경우를 포함한다.

## Baseline-attribution

로컬 테스트는 `WT-neutral-worktree-root`의 `5298dac278f299070b25ebcf4e2ebda34d78b780`에서 이번 실행에 측정했다. CI 판정은 두 통합 커밋을 포함한 `develop` `cee197917b83aff5b04761e13d5a6328179cfee1`의 run `36367483094`다.

## Gaps

공유 primary 체크아웃의 실제 워크트리들은 이 카드에서 일괄 이동하지 않았다. `moai update`의 자동 이전은 사용자가 update를 실행할 때 안전 조건을 만족하는 항목에만 적용된다. 설치된 전역 `moai` 바이너리는 이 변경을 반영한 빌드로 교체하지 않았다.

## Residual-risk

오래된 `.claude/worktrees` 중 활성 세션이나 프로세스가 사용하는 트리는 원위치에 남고 후속 update에서 재시도해야 한다. Codex 앱 자체 관리 트리와 HOME 아래 L2 트리는 이 이전의 대상이 아니다.
