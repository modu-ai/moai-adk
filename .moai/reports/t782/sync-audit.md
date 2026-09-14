# SPEC-GIT-PROC-SAFE-001 — Sync-Audit Report (card t782)

- Auditor: sync-auditor (independent, fresh-judgment)
- Date: 2026-09-14
- Audited tree: `WT-gitproc-audit` @ `4c180aa39` (worktree `.claude/worktrees/t782`)
- Baseline attribution: 모든 명령은 이번 감사 실행에서 본 워크트리(HEAD `4c180aa39`)에 대해 직접 실행, 출력은 원문 그대로 인용. 대조 baseline은 develop `c9ceff175` (범위 `c9ceff175..4c180aa39`).

## Evaluation Report

SPEC: SPEC-GIT-PROC-SAFE-001
Overall Verdict: **PASS** (harmonic mean 92/100, must-pass Functionality + Security 양차 통과)

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | 8개 AC 전수 자체 재검증 — 전부 GREEN (본 파일 §AC 재검증 참조). 핵심 원문: `cmp` AC-11 행 → `LOCAL_158_IDENTICAL` / `TMPL_156_IDENTICAL`; 금지명령 grep → 위반 문맥 0건 |
| Security (25%) | 90/100 | PASS | 재작성된 절차 텍스트에 primary-checkout 변이·파괴적 복구·비소유자 push 지시 0건. 신규 위험 신규 도입 없음. 범위 밖 잔존 인출은 전건 계획된 blocker 행으로 상향됨(§ Findings F4-F7) |
| Craft (20%) | 93/100 | PASS | `moai spec lint SPEC-GIT-PROC-SAFE-001` → `0 error(s), 2 warning(s)` (baseline 동일). `go test ./internal/template/agentemit/` → `ok ... 0.509s`. 전이 커밋 trailer 검증: M1 `Authored-By-Agent: manager-develop`, sync `Authored-By-Agent: manager-docs`. `sync_commit_sha: "d8748f618"` == 실제 sync 커밋 |
| Consistency (15%) | 92/100 | PASS | spec-workflow.md 양 사본 HEAD 바이트 동일(diff exit 0). manager-git.md 사본 차이 = 문서화된 frontmatter fork 1곳뿐. delivery.md 사본 fork 94행 — baseline 94행과 동일 세트(본 SPEC 미접촉 증명). manager-git↔spec-workflow 모델 상호 일치. codex emit = `.md` 원본과 일치(agentemit test ok) |

### AC 재검증 (감사자 독자 판정 — §D 매트릭스 8행 전부)

| AC | 판정 | 감사자 검증 방법 + 원문 증거 |
|----|------|------|
| AC-GP-01a | PASS | 현재 텍스트 직접 판독 + `grep -n 'reset --hard\|checkout main\|switch -c' .claude/agents/moai/manager-git.md` → 44행(범위 밖 선존재 Checkpoint 롤백), 129행(금지 문장 "never a recovery path") 2적중, 위반 문맥 0. GREEN 4절 전부 충족: 워크트리 한정 서술·모드조건부 승격(Phase C PR/git-flow 분기)·Phase D 소멸 문장 확인 |
| AC-GP-01b | PASS | `git diff c9ceff175..HEAD -- .claude/rules/moai/workflow/spec-workflow.md` 원문 판독 — Route B 사전조건이 launcher-entered worktree 모델로, Step 4 closure가 `git checkout main`/`git reset --hard origin/main`/`git pull origin main` 제거 + `git rev-list --count --left-right origin/<base>...<worktree-branch>` + 신규 post-condition으로 재작성. 템플릿 diff와 hunks 바이트 동일(blob `e099362ed..8c4bcea5a` 양측 동일) |
| AC-GP-01c | PASS | 옵션 행 신규 텍스트: `main_late_branch: commits accumulate on the worktree's own branch, promoted at integration time — PR ... or integration-window merge ...` — `reset --hard` 문구 없음 |
| AC-SX-01 | PASS | 현재 delivery.md 362-390행 직접 판독: 중복 3블록 제거 + 위임 문장 확인. 보존 4요소 전건 존재 — "Only applies when a PR was created in Step 3.2." / Post-Merge Automatic Cleanup(`workflow.worktree.auto_cleanup`) / 단일 기준 문장("The single criterion is the `--auto-merge` opt-in defined in `manager-git.md` § PR Auto-Merge") / merge_method 해석 1행 |
| AC-AC11-01 | PASS | `cmp <(sed -n '158p' base) <(sed -n '158p' head)` → `LOCAL_158_IDENTICAL`, 템플릿 156행 → `TMPL_156_IDENTICAL`. diff hunks가 158행 미접촉. 문구 일치: `run git fetch first and wait until it completes, then run git rev-list --count --left-right`. AC-11 편집 커밋 0건 |
| AC-XREF-01 | PASS | manager-git.md 131행 → spec-workflow Step 1/Step 4Late-branch 라벨 실존 확인. spec-workflow.md 60행 → `### Late-Branch Invocation Pattern` 앵커 실존. 끊긴 앵커 0건 |
| AC-MIRROR-01 | PASS | 6파일 전체 diff 측정: manager-git 38행×2, spec-workflow 10행×2, delivery 26행×2 + emit 38행 — 전부 발견 건 범위. 사본 간 hunk 본문 바이트 동일(오프셋만 상이: manager-git 2행=frontmatter fork, delivery 25행=선존재 fork). delivery.md 사본 fork 세트 baseline=HEAD 94행 동일 — 선존재 분기(frontmatter description, lint/크로스빌드, TRACE PROBE, Route A/B 산문, Step 3.2 300/303행) 무손상. diff 신규 행 SPEC-ID 도입 0건(`git diff .. -- internal/template/templates/ \| grep '^+' \| grep -o 'SPEC-...'` → 공집합) |
| AC-TN-01 | PASS(편집 범위 판독) | `rules/local` 3개 템플릿 사본 grep → 0적중. diff 신규 행 언어편향 토큰 → 0적중. 신규 SPEC ID 도입 → 0건. 주의: 파일 전체 literal 판독 시 선존재 `SPEC-AUDIT-SNAPSHOT-001`(템플릿 spec-workflow.md 329/345행)이 패턴 적중 — AC-MIRROR-01 보호 대상 선존재 행이라 본 SPEC 판정과 무관(F1 참조) |

### 기계 검증 배치 (본 실행 원문 출력)

```
$ moai spec lint SPEC-GIT-PROC-SAFE-001   → 0 error(s), 2 warning(s)
  (WARNING MovingRefUnpinned acceptance.md:39 / research.md:29 — baseline 선존재 2건과 동일)
$ go test ./internal/template/agentemit/  → ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.509s
$ git show 4c180aa39 --stat               → progress.md 1 file changed, 1 insertion(+), 1 deletion(-)
  (-sync_commit_sha: pending-backfill / +sync_commit_sha: "d8748f618")
$ git show d8748f618 -- spec.md           → status: in-progress → completed 단일 행 (본문 편집 0)
```

### Findings (전건 optional — blocking 0건)

- F1 [Low] [optional] acceptance.md AC-TN-01 — AC 문안이 "내부 SPEC ID 패턴(SPEC-GIT-PROC-SAFE-001 및 동류) 적중 0건"으로 전파 파일 literal 판독을 요구하는데, 템플릿 spec-workflow.md 329/345행의 선존재 `SPEC-AUDIT-SNAPSHOT-001`이 패턴에 적중한다. AC-MIRROR-01이 같은 행의 불변을 요구하므로 두 AC는 literal 판독에서 상호 충돌한다. 본 감사는 편집-범위 판독(신규 도입 0건)으로 통과 판정했다 — 이것이 두 AC가 양립하는 유일한 판독이다. Required fix: 없음(본 SPEC). 후속 AC 작성 시 "편집으로 새로 도입된" 한정자를 (1)(2)절에도 명시할 것.
- F2 [Low] [optional] CHANGELOG.md:28 — "8 live acceptance criteria (AC-01, AC-11, AC-AC11-01, AC-GP-01, ...)" 열거가 발견 건 ID(AC-01, AC-11)와 정식 AC ID를 혼용하고, "AC-11:" 문장이 Step 4 closure 재작성(AC-GP-01b 소관)을 AC-11 라벨로 서술한다. 실제 AC-11 폐쇄 대상은 manager-git.md § Synchronization 158행(무편집 검증 전용)이다. 개수 8은 정확하고 실질 주장은 거짓이 없다. Required fix: 후속 문서 기회 수리 시 AC-11 문장을 "verification-only closure of the fetch→rev-list ordering finding"으로 정정.
- F3 [Low] [optional] progress.md §E.4 말미 산문 — "`sync_commit_sha` is the D3-convention placeholder `pending-backfill` ... backfilled by the orchestrator" 문장이 백필 커밋(4c180aa39) 이후 YAML 필드(`"d8748f618"`)와 모순되게 잔존. 기계 판독면(YAML)은 정확하다. Required fix: 기회 수리 시 산문 1문장 갱신.
- F4 [Medium] [optional — 운영자 후속 카드 후보] delivery.md Step 3.3.5(양 사본, 보존 영역): `git checkout {main_branch} && git pull origin {main_branch}` / `git checkout develop && git pull origin develop` — AC-01이 제거한 것과 동일 클래스의 checkout 지시가 문맥 한정 없이 잔존(어느 체크아웃에서 실행할지 불명확). 본 SPEC 범위 밖(AC-MIRROR-01 보호), §E.2 blocker 행으로 이미 상향됨. 잔존 3건 중 가장 강한 후속 카드 후보 — 본 SPEC의 진단이 옳았다는 반증이 아니라 미완 스윕 범위의 증거.
- F5 [Low] [optional] manager-git.md 44행 Checkpoint System `Rollback: git reset --hard [checkpoint-tag]` — 워크트리 문맥 한정자 없음(선존재, 범위 밖). 자기 워크트리 내용 롤백 문맥이므로 AGENTS.md §2 위반으로 판정하지 않는다.
- F6 [Low] [optional] `.claude/rules/moai/workflow/delivery-policy.md`(로컬 전용, t614) git-flow 절 "it is still pushed by `manager-git`" — 2026-09-02 리드 일괄 push 지시(gitflow-lane-protocol.md §4)와 선존재 모순. 신규 manager-git.md Phase C("lead/operator concern")는 최신 측에 정렬 — delivery-policy.md가 이 절 한정 스테일. 본 SPEC이 만든 모순 아님(09-02부터 존재), 6파일 범위 밖.
- F7 [Low] [optional] spec-assembly.md:376 Late-branch 산문(템플릿 사본 동일) — 6파일 밖, 무편집 확인(diff stat 부재), blocker 행 기록됨. 운영자 판정 대상 유지.

### Gaps (감사자가 관측하지 못한 것)

- 병합 후 CI 녹색(acceptance.md DoD 4번째 항목) — 로컬 develop 병합·origin push가 리드 소관으로 아래 미실행 상태라 본 감사 시점에 관측 불가. 문서 전용 SPEC이므로 lint/빌드 게이트 중심 판정은 로컬에서 대체 관측함(lint 0 error, agentemit ok).
- plan-audit 보고서(.moai/reports/t782/plan-audit.md)의 내용 재감사는 본 sync-audit 범위 밖(plan-phase 판정은 이mehr 별도 게이트로 PASS 9.4/10 기록).

### Residual-risk

- 6파일 밖 late-branch 모델 잔존(F4/F5/F7)은 운영자가 후속 카드로 소화하기 전까지 교리 내 잔존 모순으로 남는다 — SPEC 위험 절이 예측한 바로 그 상태이며, blocker 상향이라는 설계된 경로로 기록되어 있다.
- F6의 delivery-policy.md 스테일 절은 레인이 delivery-policy를 먼저 읽는 세션에서 push 소관을 혼동시킬 수 있다(현행 운영 규칙은 gitflow-lane-protocol §4가 우선하는 구조).

### Recommendations

1. F4(Step 3.3.5 checkout 산문)를 후속 카드로 발행 — 본 SPEC과 동일 수리 클래스, 영향 파일은 delivery.md 양 사본.
2. F6 delivery-policy.md git-flow push 절을 09-02 리드 일괄 push 모델로 갱신하는 문서 정리 카드 검토.
3. F2/F3은 다음 문서 접촉 시 기회 수리.
