# Card t655 Verdict — worktree automation keys wired to behavior

**Card**: t655 — `workflow.worktree.auto_create` / `auto_merge` 선언-only 결함 수리 (Class C, Tier M)
**SPEC**: SPEC-WORKTREE-KEY-WIRING-001 (cycle_type ddd)
**Lane**: lane-4 (Factory Mode, session b22dbd78) · orchestration: manager-spec → plan-auditor → manager-develop → manager-docs → sync-auditor, all depth-1 leaf spawns
**Base SHA**: `1d150a27d` (= origin/develop at lane start, fetched immediately before worktree creation — per lead dispatch instruction)
**Final HEAD**: `195489beb` (branch `WT-worktree-keys-wiring`, 8 commits ahead of base, unpushed, tree clean)

## Claim

카드의 5개 범위 요건이 전부 구현·검증돼 develop 병합 대기 상태다:
1. `auto_merge` live reader 신설 (세션 종료 시 통합 창 acquire → 로컬 `--no-ff` merge → release)
2. 통합 대상 키는 신설 없이 기존 `git_strategy.<mode>.develop_branch` 재사용 (SPEC 설계 정제, 리드 보고·내 재검증 완료 — `loader_integration_branch.go:34`, card t449)
3. push 금지 유지 — zero-push seam 단정 (실구현 git 명령 화이트리스트: rev-parse/rev-list/merge/abort 4종)
4. 템플릿 기본 false 유지 + fail-safe (OFF 시 모든 세션-종료 관측치가 변경 전과 바이트 동일)
5. `shipped_key_inventory.yaml` (auto_merge D→W, deprecate_after 제거) + reader 분류 테스트 갱신

추가: `auto_create`는 wording 진실화 노선 (허위 "is auto-creating" 문구 제거 — 카드 전제 "제로 리더"는 트리와 모순, worktree_advisory.go가 실재 읽음을 감사 단계에서 정정).

## Evidence

- **plan-audit**: iteration 1 FAIL 0.75 (D1~D3 기계적 결함) → 수리 → **iteration 2 PASS 1.00** (회귀 0, 설계 결정 3종 생존) — `.moai/reports/t655/plan-audit.md`
- **run**: 14/14 AC PASS — `progress.md` §E.2에 판정 명령+출력 기록; `go test ./internal/cli/ -count=1` 2차 최종 `ok 2392.715s` 0 FAIL (run-1의 유일 FAIL은 카드 diff 밖 `TestHookWrapper_LargeStdin` 벽시계 플레이크, 단독 158.9ms PASS로 분리 입증); go build/GOOS=windows/vet/golangci-lint 0 issues; 커밋 `a0ae8a746`→`33c4c35e8`+백필, 전부 카드 id 포함
- **sync**: `313d7feb5` (CHANGELOG Added 첫 엔트리 + spec frontmatter completed + §E.4) + `195489beb` (sync_commit_sha 백필); 3 files +46/−2 마크다운 전용; docs-site·README 무변경 (키 문서 페이지 4-locale 부재 grep 실측)
- **sync-audit**: **PASS, 조화평균 9.47/10** — Functionality 10 · Security 9 · Craft 9 · Consistency 10; 감사자 자체 재실행: `TestAutoMerge|TestWorktreeAdvisoryTruthful` 16 함수·30 서브테스트 PASS `ok 3.461s`, AC-WKW-012 재판정 PASS — `.moai/reports/t655/sync-audit.md`
- **lane 스팟 체크** (진행 관리자 직접 측정): 커밋 8개 실재(`git log 1d150a27d..HEAD`), 트리 클린(`git status --short` 빈 출력), `AutoMergeNoticePrefix` 상수 실재(automerge.go:46), CHANGELOG Added 엔트리 실재(32행), `run_final_commit_sha: 33c4c35e8` 기록 실재(progress.md:143)

## Baseline-attribution

모든 수치는 이 워크트리(`.claude/worktrees/t655`), 브랜치 `WT-worktree-keys-wiring`, base `1d150a27d`에서 이 배치(2026-09-12) 동안 실제 실행한 명령의 출력이다. plan-audit/sync-audit 점수는 각 감사자가 이 트리에서 직접 재판정한 결과이고, run 수치는 manager-develop이 §E.2에 기록한 것을 sync-auditor가 표본 재실행(5 AC 셀렉터)으로 교차 확인했다.

## Gaps

- `internal/cli` 전량 스위트는 감사 단계에서 재실행하지 않았다 — heavy 슬롯 규약 + CI가 풀 스위트 판정 소관. §E.2의 `2392.715s` 기록은 귀속 근거로만 인용.
- 템플릿 코멘트 갱신에 따른 `make build`는 run 단계에서 수행됐으나(catalog 무변동 확인), sync 이후 재빌드는 없다(마크다운 전용 diff — 빌드 산물 불변).
- develop 병합·push는 미수행(리드 창 지명 전 금지) — 원격 CI 판정은 병합 후의 일이다.

## Residual-risk

- sync-audit F1 [minor]: `develop_branch` 값을 git argv에 넣을 때 `--` 구분자 부재 — 현재 fail-safe로 무해(오류 rev 식은 skip+notice), 선택적 하드닝 후보.
- F2 [minor]: zero-push 단정이 seam-log 의존 — 엔진 본체에 직접 exec.Command를 넣는 미래 변경은 우회 가능(현재 4종 화이트리스트 + fixture 원격 부재가 보완). 정적 소스 스캔 테스트가 후속 후보.
- run-1 플레이크(`TestHookWrapper_LargeStdin_DoesNotExceedTimeout`)는 이 카드 밖 기존 부하 민감 테스트 — develop 트리에서 간헐 재현 가능성은 이 카드 소관이 아님.
- 자동 병합은 origin/develop 흡수·재측정을 자동화하지 않는다(카드의 LOCAL ONLY 스코프) — 국소 develop이 origin보다 뒤처질 수 있고, merge SHA는 통지에 실려 리드 배치 push 흐름이 읽는다 (design.md §1.3 한계 기록).
