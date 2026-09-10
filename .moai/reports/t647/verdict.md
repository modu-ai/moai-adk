# t647 — t623 이후 낡은 plan-auditor 카탈로그 해시

card: t647 · A에 준함(생성물 한 줄, 설계 판단 없음) · lane-9
branch: WT-catalog-auditor-hash · base: 로컬 develop `6d228ea19` (워크트리 병합 커밋 `aa3ac13a9`, 트리 동일)
원인: t623 이 `internal/template/templates/.claude/agents/moai/plan-auditor.md` 를 고치고 `internal/template/catalog.yaml` 의 해시를 재생성하지 않았다. t623 검증이 카탈로그 해시 테스트를 돌리지 않아 놓쳤다.

## 1. 재현 — 수리 전 트리에서 실패하는 테스트

### Claim
develop `6d228ea19` 의 카탈로그는 템플릿 파일의 정규화 내용과 맞지 않는 plan-auditor 해시를 적고 있고, `TestManifestHashFormat` 은 정확히 그 한 항목에서 실패한다.

### Evidence
기반 정렬: 새 워크트리(`origin/develop` `d060e0d13` 기준 생성) → `git status --short` 출력 없음 → `git branch -m WT-catalog-auditor-hash` → `git merge --no-ff -m "Merge develop 6d228ea19 as t647 base (card t647)" 6d228ea19` → `merge_exit=0`, `git log -1 --format='%h %p'` → `aa3ac13a9 d060e0d13 6d228ea19`, `git rev-parse --short HEAD^2` → `6d228ea19`. `git rev-parse HEAD^{tree}` = `git rev-parse 6d228ea19^{tree}` = `e906eab95fd1a96cd224f2c3e2a3e13f98fef53f`.

기록된 값: `git show 6d228ea19:internal/template/catalog.yaml | sed -n '142,147p'` → `- name: plan-auditor` / `path: templates/.claude/agents/moai/plan-auditor.md` / `hash: 2403bfb3e0747467ba328cdbfe3767c0801db95702568d5cb606596655d771fb`.

테스트는 원본 바이트가 아니라 **정규화한 내용**을 해시한다(`catalog_tier_audit_test.go:445` `sha256.Sum256(NormalizeForHash(rawContent))`; `catalog_hash_norm.go:31` — CRLF→LF, 줄 끝 공백 제거, 마지막 줄바꿈 정확히 1개). 그래서 원본 sha256 만으로는 낡음이 증명되지 않는다. 둘 다 쟀다:
- 원본: `git show 6d228ea19:internal/template/templates/.claude/agents/moai/plan-auditor.md | shasum -a 256` → `621cb9eea7abe2c649be12d5aa3f494feb7fd525b3f278f94c497b6f9d306298`.
- 정규화(생성기 자신의 계산): `go run ./internal/template/scripts/gen-catalog-hashes.go --dry-run --entry plan-auditor` → `dryrun_exit=0`, `[dry-run] plan-auditor: 621cb9eea7abe2c649be12d5aa3f494feb7fd525b3f278f94c497b6f9d306298`, `[dry-run] catalog.yaml not modified`. 이 파일에서 정규화는 항등이다.
- 대조군: manager-git 기록값 `061fe4fb0dfee5b06a2c67b51835c5512f45f0c56a888fbf786c5d9220099b07` = `manager-git.md` 원본 sha256 `061fe4fb…`.

실패하는 테스트(09:35:39Z–09:35:41Z): `go test ./internal/template -count=1 -run '^TestManifestHashFormat$' -v > red-before-fix.log 2>&1` → `test_exit=1`:
```
catalog_tier_audit_test.go:451: CATALOG_HASH_UNSTABLE: plan-auditor stored hash=2403bfb3e0747467ba328cdbfe3767c0801db95702568d5cb606596655d771fb, computed hash=621cb9eea7abe2c649be12d5aa3f494feb7fd525b3f278f94c497b6f9d306298 (source=.claude/agents/moai/plan-auditor.md)
catalog_tier_audit_test.go:456: audited 45 catalog entries for hash validity
--- FAIL: TestManifestHashFormat (0.01s)
```
감사한 45개 항목 중 `CATALOG_HASH_UNSTABLE` 줄은 정확히 1개다.

### Baseline-attribution
워크트리 `WT-catalog-auditor-hash` HEAD `aa3ac13a9`(트리 = develop `6d228ea19`), 이 실행.

### Gaps
- 재현 자체에는 없음. 수리와 통과 실행은 §2.

## 2. 수리 — 생성기로 한 줄 재생성, 같은 테스트 통과

### Claim
저장소 생성기로 plan-auditor 항목만 재생성하면 `catalog.yaml` 이 정확히 한 줄 바뀌고, 그 뒤 `TestManifestHashFormat` 이 45개 항목 전부에서 통과한다. 해시를 손으로 적지 않았다.

### Evidence
재생성(09:35:54Z): `go run ./internal/template/scripts/gen-catalog-hashes.go --entry plan-auditor > gen-entry.log 2>&1` → `gen_exit=0`:
```
Computing hash for entry "plan-auditor"...
  plan-auditor: 621cb9eea7abe2c649be12d5aa3f494feb7fd525b3f278f94c497b6f9d306298
catalog.yaml updated successfully (12899 bytes)
```
생성기가 `--entry NAME` 을 지원하므로(`gen-catalog-hashes.go:198` 에 플래그 정의) `--all` 은 쓰지 않았다.

diff: `git diff --stat` → `internal/template/catalog.yaml | 2 +-`, `1 file changed, 1 insertion(+), 1 deletion(-)`. `git diff -U0 -- internal/template/catalog.yaml`:
```
@@ -145 +145 @@ catalog:
-              hash: 2403bfb3e0747467ba328cdbfe3767c0801db95702568d5cb606596655d771fb
+              hash: 621cb9eea7abe2c649be12d5aa3f494feb7fd525b3f278f94c497b6f9d306298
```
145행만 바뀌었고, 주석이나 다른 항목은 그대로다.

통과 테스트(09:36:22Z–09:36:24Z): `go test ./internal/template -count=1 -run '^TestManifestHashFormat$' -v > green-after-fix.log 2>&1` → `test_exit=0`:
```
catalog_tier_audit_test.go:456: audited 45 catalog entries for hash validity
--- PASS: TestManifestHashFormat (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.442s
```
`CATALOG_HASH` 줄: 0.

### Baseline-attribution
워크트리 `WT-catalog-auditor-hash` HEAD `aa3ac13a9` + `catalog.yaml` 한 줄 변경, 이 실행.

### Gaps
- `TestManifestHashFormat` 만 돌렸다. `internal/template` 의 나머지 테스트는 돌리지 않았다.
- `make build`·embed-check 는 돌리지 않았다(배치 끝, 리드 몫).
- t623 이 바꾼 템플릿 파일 가운데 plan-auditor 외의 파일을 따로 재확인하지는 않았다. 다만 이 테스트는 카탈로그 45개 항목 전부를 감사하고, 실패는 plan-auditor 하나뿐이었다.

### Residual-risk
- 앞으로도 템플릿을 고치는 카드의 검증에 카탈로그 테스트가 빠지면 같은 종류의 낡은 해시가 다시 남을 수 있다.
