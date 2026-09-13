# t566 통합 창 기록 — 흡수 트리 재측정

- 카드: t566 · 레인: lane-6 · 브랜치: `WT-fold-unit-guard`
- 창 지명: 리드, t660 release 뒤. 승인된 계획: `TestCodemapsFoldJudgments` 스코프 1회 + 실제 트리 동사 + `./internal/template/` 무선택자 + `commands-emit-check`·`agents-emit-check`.
- 도구: `go version go1.26.8 darwin/arm64`

## 1. Claim

1. 로컬 develop `d669090f8` 는 흡수 시점에 최신이었다(원격 대비 뒤처짐 0).
2. 흡수 병합 `79d001c07`(트리 `43f9d43b`)은 t566 이 만든 파일을 하나도 바꾸지 않았고, `catalog.yaml` 자동 병합은 develop 쪽 `plan-auditor` 해시 한 줄만 들여왔다.
3. 그 트리에서 승인된 네 검사가 모두 통과하고, 실제 트리 동사는 표지 없이 `COLLECTED: fold=5 omission=15` 를 낸다.

## 2. Evidence

### 2.1 창 획득과 develop 최신성

```
moai integration acquire --name lane-6
→ release-integration window acquired by e6946f98-39aa-4df4-9a3f-3b9caed0ab4d on WT-fold-unit-guard
git --no-optional-locks status --porcelain --untracked-files=all   → (출력 없음)
git fetch -q origin develop                                        → FETCH_EXIT=0
git rev-parse origin/develop develop                               → 987eb7e4021ab7a74baef33242307087a74238ae / d669090f87516ded9fd75b35a80341a410806071
git rev-list --count --left-right origin/develop...develop         → 0	26
```

### 2.2 흡수 (`absorb-merge.txt`, `absorb-changed-files.txt`)

```
git merge --no-edit develop   → MERGE_EXIT=0
  Auto-merging internal/template/catalog.yaml
  Merge made by the 'ort' strategy.
git rev-parse HEAD HEAD^{tree} HEAD^1 HEAD^2
→ 79d001c0784b6e300cace1e6152eee34e05ae089
  43f9d43b49374df0b81d0ce11eea741dd4ab3d1f
  30b43f79cb717853c39723be9b962424b808efb9
  d669090f87516ded9fd75b35a80341a410806071
git diff --name-only 30b43f79c HEAD | wc -l   → 276
```

### 2.3 흡수 델타 판단

t566 소유 경로:

```
git diff --stat 30b43f79c HEAD -- internal/template/templates/.claude/skills/moai/workflows/codemaps.md .claude/skills/moai/workflows/codemaps.md internal/cli/codemaps_fold_judgments_test.go .moai/project/codemaps .moai/reports/t566
→ (출력 없음)
대조군: git diff --stat 30b43f79c HEAD -- .claude/agents/moai/plan-auditor.md
→ .claude/agents/moai/plan-auditor.md | 109 +++++++++++++++++++++++++++++++++++-
cmp internal/template/templates/.claude/skills/moai/workflows/codemaps.md .claude/skills/moai/workflows/codemaps.md → CMP_EXIT=0
```

`catalog.yaml` 흡수 diff 전문:

```
@@ -142,7 +142,7 @@ catalog:
             - name: plan-auditor
               tier: core
               path: templates/.claude/agents/moai/plan-auditor.md
-              hash: d7be4a4a8695459dc5d80d76543a7091476a6423f28ffbf542ab30032fe711b3
+              hash: dea6916d4289d16a1448053c24479f9ecba87e72aa14fd00879114a05a97112d
```

t566 의 moai 항목 해시 줄은 흡수로 바뀌지 않았다. 두 변경이 서로 다른 항목이라 자동 병합이 겹치지 않았고, 해시가 실제 파일과 맞는지는 2.6 무선택자 실행이 판정한다.

흡수로 들어온 코드: `internal/cli`(update 계열, spec_view, plan_audit_order_conflict_test, tool_policy_test), `internal/spec`(중복 AC id), `internal/web`(템플 화면·Host 게이트 테스트), `internal/config/loader_git_mode.go`, 템플릿 `plan-auditor` 정의.

### 2.4 `internal/cli` 스코프 (`cli-fold.txt`, 사전 확인 `slot-precheck-cli.txt`: 2026-09-10T18:34:22Z, load 19.59, 다른 go test 0, 관측자 대조군 1)

```
unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -count=1 -v -timeout 600s -run 'TestCodemapsFoldJudgments'
```

```
--- PASS: TestCodemapsFoldJudgments_VerbIdenticalAcrossCopies (0.00s)
--- PASS: TestCodemapsFoldJudgments (0.00s)
    --- PASS: TestCodemapsFoldJudgments/gap_missing_judgments (0.01s)
    --- PASS: TestCodemapsFoldJudgments/gap_no_documents (0.01s)
    --- PASS: TestCodemapsFoldJudgments/red_fold_prose (0.01s)
    --- PASS: TestCodemapsFoldJudgments/gap_malformed_line (0.01s)
    --- PASS: TestCodemapsFoldJudgments/red_uncovered_omission (0.01s)
    --- PASS: TestCodemapsFoldJudgments/gap_empty_judgments (0.01s)
    --- PASS: TestCodemapsFoldJudgments/green_normal_state (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.990s
EXIT=0
```

### 2.5 실제 트리 동사 (`../verb/window-real-green.txt`, 도구 `../tools/run_verb.py`)

```
python3 .moai/reports/t566/tools/run_verb.py window-real-green
doc: internal/template/templates/.claude/skills/moai/workflows/codemaps.md
--- stdout ---
COLLECTED: fold=5 omission=15
--- stderr ---
EXIT=0
```

`FOLD-PROSE` 등 위반 표지 줄은 0이다. 창 전 기록 `../verb/real-green.txt` 와 출력이 같다.

### 2.6 `./internal/template/` 무선택자 (`template-noselector.txt`, 사전 확인 `slot-precheck-template.txt`: 2026-09-10T18:34:57Z, load 14.21, 다른 go test 0)

```
unset … && go test ./internal/template/ -count=1 -v -timeout 600s
top-PASS 363 top-FAIL 0 any-FAIL 0 SKIP 15
--- PASS: TestManifestHashFormat (0.09s)
ok  	github.com/modu-ai/moai-adk/internal/template	30.056s
EXIT=0
```

건너뛴 15건: `TestBuildSmartPATH_WSL2`, `TestRootLevelCommandsThinPattern`, `TestEmbeddedTemplates_LLMConfig`, `TestEmbeddedTemplates_Announcements`, 그리고 `TestNoOrphanedManager{TDD,DDD}Reference` 의 보관 에이전트 파일 하위 11건.

### 2.7 방출 검사 (`commands-emit-check.txt`, `agents-emit-check.txt`)

```
make commands-emit-check → ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.343s / EXIT=0
make agents-emit-check   → ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.418s / EXIT=0
```

## 3. Baseline-attribution

흡수 병합 `79d001c07`, 트리 `43f9d43b49374df0b81d0ce11eea741dd4ab3d1f`. 네 검사는 직렬로 실행했다(cli → 동사 → template → commands-emit → agents-emit). 검사 사이 추적 파일 쓰기는 이 `window/` 증거 파일과 `verb/window-real-green.txt` 뿐이다.

## 4. Gaps

- 창에서 뮤턴트는 다시 돌리지 않았다. t566 소유 파일이 흡수로 바뀌지 않았으므로(2.3) 창 전 뮤턴트 기록(`../verb/`, `../verdict.md` §7)의 대상이 그대로라는 판독에 기댄다.
- `internal/cli` 패키지 전체는 실행하지 않았다(스코프 승인 1회).
- 무선택자 템플릿 실행의 SKIP 15건은 이 창에서 판정되지 않았다.
- 흡수로 들어온 다른 카드 코드(`internal/web`, `internal/spec`, update 계열)의 검증은 각 카드의 몫이다.

## 5. Residual-risk

- develop 병합 트리는 이 기록을 담은 증거 커밋의 트리와 같아야 한다. 병합 뒤 `git rev-parse <merge>^{tree}` 로 대조해 리드 보고에 싣는다.
- 병합 이후 로컬 develop 에 다른 레인이 더 병합하면 이 재측정은 그 커밋들을 포함하지 않는다.
