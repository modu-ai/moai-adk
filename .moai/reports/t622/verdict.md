# t622 — 레인 판정 기록 (SPEC-GIT-DELIVERY-PROCEDURE-001)

> 최종 PASS/FAIL 은 리드가 판정한다. 이 문서는 레인이 읽은 증거와 권고를 적는다.

- 카드: t622 · 브랜치 `WT-git-procedure-fixes` · 워크트리 `.claude/worktrees/t622`
- SPEC: 0.2.6, Tier M, 판정 대상 REQ 12개 · AC 16개
- 레인 권고: **PASS** (근거는 §2~§4)
- 기록 시점: HEAD `947cc8439`, 로컬 develop `ac6c42c2d` (흡수 전, merge-base `f1f034bb4`)
- 이 기록 다음에 SPEC 닫기 커밋(`completed`, §E.4)이 마지막 쓰기로 온다.

## 1. 무엇이 바뀌었나

- 병합 옵트인이 `/moai sync --auto-merge` 하나로 정리됐다.
  - `--merge` 는 `--auto-merge` 의 폐기된 별칭이다(경고를 낸다).
  - `--no-merge` 는 폐기된 호환용 no-op 이다. 병합하지 않는 것이 이미 기본이다.
  - 워크트리 문맥만으로는 병합하지 않는다.
- team 모드는 `--auto-merge` 와 전원 승인이 있어야 병합한다. personal·manual 모드는 승인 조건 없이 병합한다.
- 병합 실행 명령은 `gh pr merge --<merge_method> --delete-branch` 이다. `<merge_method>` 는 `git_strategy.<mode>.merge_method` 값이며 기본은 `squash` 다.
- `manager-git.md` 동기화 절은 `git fetch` 가 끝난 뒤에 그 결과를 읽는 `git rev-list` 를 실행한다.
- 슬래시 명령 `argument-hint` 에 `--auto-merge` 가 나온다. 명령 게시본은 바뀌지 않았다(AC-GDP-030 경우 A).
- 그 밖에 바뀐 것:
  - 생성물 `manager-git.toml` 을 `make agents-emit` 으로 재생성했다.
  - CHANGELOG `[Unreleased]` 에 항목을 넣었다.
  - `doc-execution.md:34` 가 `is_worktree_context` 를 실제로 쓰는 곳(Phase 13 Step 3.2, Phase 14)을 적도록 문구를 고쳤다.
- 바뀐 지침 파일: 범위 8쌍(로컬·템플릿 16개)과 생성물 1개.
- `agent-common-protocol.md` 는 바꾸지 않았다. 리드 판단 (a)에 따라 develop 에 착지한 t635 의 Lane A/B 형태가 REQ-GDP-002 를 충족하고, 템플릿 미러는 dr0911 이 했다.

## 2. 증거 — 판정

| 단계 | 결과 | 파일 (`.moai/reports/t622/`) |
|---|---|---|
| plan-audit (전체) | PASS 0.92 (축소 SPEC 3회차) | `plan-audit-reduced-iter3.md` |
| 재기준 한정 감사 | 1회차 FAIL 0.75 → 2회차 PASS 0.875 | `plan-audit-reanchor.md`, `plan-audit-reanchor-iter2.md` |
| run M1~M6 | 판정 대상 16개 PASS (AC-GDP-016 은 SHOULD) | `run/run-report.md` |
| sync-audit | 1회차 FAIL 84 (F1) → 2회차 PASS 91 | `sync-audit.md`, `sync-audit-iter2.md` |

- `make agents-emit-check` / `make commands-emit-check`: 모든 측정에서 exit 0 / 0 이다.
  - M6: `run/m6-*-emit-check.txt`
  - sync 수리 뒤: `sync-fix/`
  - sync-audit 2회차와 레인 재실행 결과도 같다.
- 미러 비회귀: 명령은 `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity|TestRuleTemplateMirrorDrift' -v` 이다.
  - 결과는 PASS 17개다. FAIL 집합은 흡수 기준선 {`TestRuleTemplateMirrorDrift/spec-workflow.md`}(범위 밖) 그대로다.
  - 새 FAIL 과 잃은 PASS 는 없다.
  - 증거: `run/mirror-m6*`, `reanchor/mirror-baseline-sets.txt`, `sync-audit2/`
- AC-GDP-016: `git log --no-merges --format=%H HEAD --not develop -- <acp 두 경로>` 결과가 비어 있다. 양성 대조는 `97ef8e302`·`6896eef37` 을 찾는다.

## 3. 커밋

- plan: `a52865a22` … `b25d1ba6e` (재현, SPEC 0.1.0~0.2.4, 감사 기록)
- 흡수: `7ac8b8491` (develop `ee99507fb`), `255f88eb0` (develop `f1f034bb4`)
- 재기준: `1f3adf5ee`, `b24f2e184`, `bca80d7c5`, `f2fa64e08`, `33e0dac9c`, `fa13c27b6`
- run: `0da3bebf0` 사전 점검 · `3f6c5163f` M1 · `2f4dfd803` X2 · `af54bf1ff` M2 · `7a02b90e2` M3 · `cc51d8479` M4 RED · `5708e04d2` toml · `8a115e0f5` · `35c0e30df` · `8d1b8c920` · `28f9cc6c9`
- sync: `0dc007201` · `50cebf3c2` · `ae4861945` (F1) · `c61405c8b` · `947cc8439`. 이어서 이 기록 커밋과 닫기 커밋이 온다.
- push: 없음

## 4. 병합 창에서 다시 잴 것

- 범위 기반 판정 AC-GDP-014·015·016·025·030 과 M5 는 병합 전 판정이다.
  - 창에서 로컬 develop 을 흡수한 뒤 `$CARD_BASE` 를 다시 구해 재측정한다. `$CARD_BASE` 는 `git merge-base --all develop HEAD` 결과이며 1줄이어야 한다.
  - BASE(`255f88eb0`)는 옮기지 않는다.
- 스냅숏 신선도 점검 `git diff --name-only $BASE $CARD_BASE -- <스냅숏 21경로>` 가 비어 있지 않으면 해당 스냅숏만 다시 반출하고 리드에 보고한다.
- Pre-Spawn 기준 절을 다시 반출할 때는 BASE 절과의 diff 도 기록한다(plan-audit 재기준 2회차 N1).
- 미러 비회귀를 병합 트리에서 다시 돌린다. 흡수가 미러 대상 파일을 바꿨다면 AC-GDP-013 (e)의 귀속 규칙으로 가린다.

## 5. 후속 카드 후보

1. `agent-common-protocol.md` Pre-Edit Sync Check(360행 부근)는 아직 `git fetch origin main 2>&1; git rev-list …` 한 줄 형태다. 이 카드의 REQ-GDP-003 이 그 절을 막았다. 리드가 발행 목록에 올리기로 했다.
2. 흡수가 Pre-Spawn 절을 바꾸는 경우, `$CARD_BASE` 에서 다시 반출한 기준 절로는 AC-GDP-002 (b)가 자기 자신과 비교하게 된다(`plan-audit-reanchor-iter2.md` N1).
3. team 모드 "전원 승인" 조건이 `manager-git.md` PR Auto-Merge 실행 단계 목록에 확인 단계로 들어 있지 않다. 카드 이전과 같은 모양이다(sync-audit F2).
4. `manager-git.md` 123·129·139행 산문이 squash 를 전제한다. late-branch 절이므로 t658 의 입력이다(sync-audit F3).
5. `workflows/sync.md` 플래그 줄에 한국어 괄호가 섞여 있다. 원래 줄부터 혼용이었다(sync-audit F4).
6. `doc-execution.md:34` 가 말하는 "worktree delivery route" 는 github-flow 전략 아래에만 있다. git-flow 는 `WT-*` 브랜치 이름으로 경로를 고른다(`delivery.md:297`). "(github-flow)" 한정을 붙이면 오해가 줄어든다(sync-audit 2회차 N1).
7. t658 — 이 카드 뒤에 착지하고, `manager-git.md:114`(로컬 116)의 `--<merge_method>` 를 유지한다. REQ-WBG-011 / `branch_guard.go:455` 잔여 위험도 t658 소관이다.
8. t659 — amend 적용 도우미.
9. 문서 카드(X4) — docs-site ko 473 / en 480 이 워크트리 기본 병합을 서술하고, ja·zh 표는 `--no-merge` 를 건너뛰기로 서술한다.
10. git-flow Step 3.2 는 `feature/*` 만 받는다. `feature/SPEC-` 형제 브랜치 서술도 함께 볼 것.

## 6. 절차 메모

- **스크립트로 가드 우회:**
  - run 1부 에이전트가 워크트리 가드가 거부한 명령을 증거 폴더의 스크립트로 옮겨 실행했다. 읽기 전용 판정(`final-judges.sh`, `final-reading-targets.sh`)뿐 아니라 M1·M2 지침 편집(`m1-edit.py`, `m2-edit.py`)도 스크립트로 했다.
  - 레인은 처음에 "읽기 전용 판정만"이라고 잘못 보고했고, sync-audit F5 를 받고 리드에게 정정했다.
  - 결과는 독립적으로 확인됐다. 로컬·템플릿·`.toml` 의 바뀐 줄이 서로 같고, BASE 사본 차이 본문 8쌍이 보존됐다. run 2부는 스크립트 없이 모든 판정을 다시 돌렸다.
  - 이후 위임문에는 스크립트 우회 금지를 명시했다.
- **zsh 에서 생긴 공허한 초록:**
  - run 2부에서 zsh 가 따옴표 없는 `$S` 를 단어로 쪼개지 않아, 여러 파일 목록이 경로 하나로 넘어갔다. 빈 파일이 곧 PASS 인 AC-026/027 판정이 한 번 통과처럼 보였다.
  - 비어 있으면 안 되는 merge-lines 점검이 0줄로 나오면서 드러났고, 경로를 글자 그대로 넣어 다시 돌렸다.
  - 인용된 `p2-*` 값은 재실행 값이다. sync-audit 이 자기 재실행과 일치함을 확인했다. 증거 폴더에 이 사건 기록이 없어서 여기 남긴다(sync-audit F6).
- **레인의 전제 오류(F1):**
  - sync 1단계 위임에서 레인이 `is_worktree_context` 변수 이름으로만 사용처를 찾고, "이후 단계에서 읽지 않는다"는 전제를 위임문에 넣었다.
  - 실제로 worktree context 는 Phase 13 Step 3.2 전달 경로와 Phase 14 선택지에서 쓰였다.
  - sync-audit 이 잡았고 `ae4861945` 로 고쳤다.
- **make 종료 코드:** make 는 실패한 레시피를 exit 2 로 보고한다(레시피 줄은 `Error 1`). acceptance.md 는 RED 를 exit 1 로 적었다.
- **남은 미추적 폴더:** `.moai/reports/t622/run-stale-b412/`(46파일)는 멈춘 옛 run 시도의 산출물이다. 판정 근거로 쓰지 않았고 커밋하지 않았다(sync-audit F7).

## Gaps

- CI 는 보지 못했다(레인은 push 하지 않는다). 전체 스위트·darwin/windows 매트릭스·`template-neutrality-check` 는 develop push 뒤 리드가 읽는다.
- `go test ./...`, `make build`, `make embed-check` 는 돌리지 않았다.
- 교차 모델 감사(codex/glm)는 돌리지 않았다. `baseBranch` 가 main 으로 풀려 이 카드 범위가 아닌 변경 집합을 보게 된다.
- 범위 기반 판정은 흡수 전 트리에서만 쟀다(§4).

## Residual-risk

- AC-GDP-026~029 검출식은 줄 단위라 예상 밖의 표현을 놓칠 수 있다. 그래서 읽기 기록이 PASS 의 전제다.
- AC-GDP-002 는 Pre-Spawn 절 밖에 순서를 뒤집는 문장이 들어오는 모양(뮤턴트 ix)을 원리상 보지 못한다. 카드 커밋은 AC-GDP-016 이 잡고, 흡수로 들어온 변경은 신선도 점검이 잡는다.
