# t541 통합 창 재측정 기록

창 보유: lane-8 (`moai integration acquire --name lane-8`). 리드 지명 후 acquire.
워크트리 `.claude/worktrees/t541`, 브랜치 `WT-auditor-order-conflict`.

## 흡수

- 흡수 대상: 로컬 `develop` `92bf71523bf5d179e2e025baf9b8674ae3826fe9` (원격 아님).
- 흡수 전 브랜치 HEAD: `f028ca9a8`.
- 명령: `git merge --no-ff develop` → 출력 `../window-merge.log`, `Auto-merging internal/template/catalog.yaml` / `Merge made by the 'ort' strategy.` / exit 0.
- catalog.yaml 은 자동 병합됐고 충돌은 없었다. 그래서 생성기는 쓰기 대신 dry-run 으로만 확인했다.
- 흡수 병합 커밋: `e8ae13298b1bb3ea1297cf6c51075d1410241c4d`, 부모 `f028ca9a8` · `92bf71523`, 트리 `08711d7442c9cfa49f5053f28410198fabef1312`.
- 병합 직후 `git rev-parse -q --verify MERGE_HEAD` → 출력 없음, exit 1.

## 병합 트리 재측정

| 검사 | 파일 | exit | 판독 |
|---|---|---|---|
| 해시 dry-run `gen-catalog-hashes.go --entry plan-auditor --dry-run` | `gen-entry-dryrun.log` | 0 | 계산값 `dea6916d…112d` 가 catalog.yaml 145행과 같다 → 재생성해도 차이 0 |
| `make agents-emit-check` | `agents-emit-check.log` | 0 | `ok .../internal/template/agentemit` |
| `go test -count=1 ./internal/template/` (선택자 없음) | `template-test.txt` | 0 | `ok .../internal/template 59.688s` |
| `go test ./internal/cli/ -count=1 -v -run 'TestPlanAuditOrder_'` | `cli-order-test.txt` / `.exit` | 0 | `=== RUN` 15, `--- PASS` 15, `--- FAIL` 0 |

internal/cli 사전 확인: `pgrep -fl 'internal/cli|cli\.test'` 출력 없음 exit 1 (다른 internal/cli 컴파일 0). 대조 `pgrep -f claude | wc -l` → 44. internal/cli 는 창 승인 범위 1회만 돌렸고 전체 스위트는 돌리지 않았다.

## 변경 범위

`CARD_BASE=$(git merge-base develop HEAD)` → `92bf71523…`. `git diff --name-only CARD_BASE..HEAD` → 46개 파일.
증거 디렉터리 `.moai/reports/t541/` 밖은 5개다.

- `.claude/agents/moai/plan-auditor.md`
- `internal/cli/plan_audit_order_conflict_test.go`
- `internal/template/catalog.yaml`
- `internal/template/templates/.claude/agents/moai/plan-auditor.md`
- `internal/template/templates/.codex/agents/moai/plan-auditor.toml`

## 미검증

- 재측정은 이 기록을 커밋하기 전 트리(흡수 병합 커밋 `e8ae13298`)에서 했다. 이 기록 커밋은 `.moai/reports/t541/` 아래 파일만 더하므로 코드·템플릿 트리는 같다. 이 주장의 근거는 커밋 후 `git diff --name-only e8ae13298..HEAD` 이다.
- develop 쪽에서 들어온 core/git·spec 변경의 자체 테스트는 이 카드 범위가 아니라 돌리지 않았다.
