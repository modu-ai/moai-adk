# t1959 progress — SPEC-less merge-ready 경로 수리 (GH #1913)

- Class: B-유사 직접 수리(parallelizable, 리더 배정 10-25) · run tn8rpj · worktree t1930 바인딩 확인(`factory next --card t1959 --run tn8rpj` → `t1959 stage=- worktree=t1930`)
- 브랜치: WT-merge-ready-specless @ origin/main 53fbc2d206에서 생성

## 결함 (GH #1913 재현)

`moai factory merge ready`는 SPEC-less(Class B) 카드를 CLI 게이트에서 전부 거부:
- `internal/cli/factory_merge.go:77-79` — 빈 `--spec`에 하드 에러 (`--spec is required`)
- `:279` — `MarkFlagRequired("spec")`로 플래그 생략 자체가 cobra 거부
반면 엔진은 SPEC-less 경로를 이미 보유(`internal/factorylane/merge.go:126,182` `SpecDir==""` → `merge_specless.go`가 `.moai/reports/<card>/verdict.md`로 조건 (a) 판정) — 동사에서 도달 불가능했을 뿐.

## 수리 (커밋 c1c9c26d98, 2 files +90/−23)

1. `--spec` 생략 또는 카드 id를 `--spec`으로 전달 → SPEC-less 경로 라우팅(`specID=""` → SpecDir 빈 문자열)
2. 빈 specID의 `filepath.Join` 함정 명시 차단(빈 문자열 유지 — specs 디렉터리 아님)
3. SPEC 보유 카드는 progress.md 게이트 유지 — SPEC 경로·타 게이트(conflict-free·tree-identity·re-measure·window) 전부 불변
4. `MarkFlagRequired("spec")` 제거 + 플래그 헬프에 SPEC-less 계약 기술
5. 테스트 신계약: 구 git-flow `--spec is required` 핀 → 성공 회귀 2건(spec 생략·card id, cleared+window 획득) + 거부 유지(판정 파일 부재 → 조건 (a) 거부·창 미획득)

## 검증 (레인 직접, 이 트리·이 커밋)

- `gofmt -l` 두 파일: 클린 · `go build ./internal/cli/`: exit 0
- `go test ./internal/cli/ -run 'Specless|MergeReady' -count=1` → `ok ... 21.569s`
- `go test ./internal/cli/ -run 'TestFactoryMergeReadySpecLess' -count=1 -v` → `spec_omitted` PASS · `spec_is_card_id` PASS · `TestFactoryMergeReadySpecLessWithoutVerdictRefused` PASS (ok 4.291s)
- `go vet ./internal/cli/`: exit 0 · `golangci-lint run ./internal/cli/...`: `0 issues`

## 착지

- push: `WT-merge-ready-specless` → origin(신규 브랜치, non-force)
- PR: **#1915 draft**(base main, 제목 `[t1959] fix(factory): reach merge ready for SPEC-less Class B cards`)
- card-review: `.moai/reports/t1959/card-review.md` — codex CLI 폴백(MCP 도구 부재, fallback 사유 기재) **verdict pass** @ c1c9c26d9891845fb1c11fb1a3df1a26ccb6b4ce, tree_digest 261ab9e326f0f6f2, codex-cli 0.163.0, findings 없음
- 독립 감사: sync-auditor 스폰(head 결속 1회) — 판정 `.moai/reports/t1959/verdict.md` 대기 중. N1 하네스 결함(수령 가드 재앵커) 재발 시 1회 mint 후 Gaps 기재·판정 진행 지시 완료.
- 병합·이슈 close: 리더(이슈 한글 코멘트와 함께 CI+감사 PASS 뒤)

decision record: decided_by=lane-7 (t1959 repaired, verified, landed as draft PR #1915) evidence_refs=git show c1c9c26d98, ok 21.569s, card-review.md pass, PR 1915 ladder_path=leader directive — awaiting head-bound audit verdict, then leader reply
