# cc234-align VCI Evidence — CC 2.1.234 teammate-model 문서 정렬 (Tier 1)

Branch: WT-cc234-align @ base 52c5f7ab3 (origin/release/v3.1.1)
Date: 2026-08-18
Session: db221a6c-e73f-4806-b60e-bc00af9ab6fa (run lane)

## 1. Claim (주장)

CC 2.1.234가 `/config`의 "Default teammate model" 설정을 제거하고 팀메이트가 리더 모델을
기본 상속하도록 변경함에 따라, moai가 모든 배포 사용자에게 발송하던 낡은 주장
("`/model`은 상속되지 않는다") 8개 표면을 현재 사실로 정정했다:

1–2. `CLAUDE.md:151` + `internal/template/templates/CLAUDE.md:151` (byte-identical pair) —
   제약 목록 내 구문 교체: "`/model` inherited from the leader by default since CC 2.1.234
   (a spawn-named model overrides; effort inherited since v2.1.186)"
3–4. `.claude/rules/moai/workflow/orchestration-mode-selection.md:120` + template mirror —
   "`/model` IS inherited from the leader by default since CC 2.1.234 (the former Default
   teammate model `/config` setting was removed; a spawn-named model overrides; effort
   inheritance unchanged since v2.1.186)"
5–8. `docs-site/content/{en,ko,ja,zh}/claude-code/agentic/agent-teams.md` — 낡은 두 문장
   문단을 v2.1.234 상속 서술로 교체 (en·ko는 release-update 전문가 계획서 초안 그대로,
   ja·zh는 동일 문장 순서·버전 식별자 verbatim 번역; zh는 기존에 누락돼 있던 effort
   상속 문장을 포함한 전체 문단으로 교체해 4로케일 내용 일치).

런타임 영향 0 (전문가 조사 실증: `defaultTeammateModel|teammateModel` 코드 0히트 — 본
워크트리에서 재확인 없음, 조사 산물 근거; 아래 Gaps 명시). Tier 2 선택 항목(statusline
GitLab 뱃지·usage-limit 자동계속·CLAUDE_CODE_PROJECT_DIR_NAME)은 계획서 권고대로 별도
docs 번들 카드로 이연. 조사 산물의 선택적 GLM 문장(`glm --team` 리더의 glm-5.3 전파)은
**실측 없는 실행 거동 주장**이라 증거 규율상 미포함 (VCI §1.1 surface 3·4).

## 2. Evidence (증거)

### 미러 쌍 동일성 (편집 후)

```
$ diff CLAUDE.md internal/template/templates/CLAUDE.md && echo "CLAUDE-pair identical"
CLAUDE-pair identical
$ diff .claude/rules/moai/workflow/orchestration-mode-selection.md \
    internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md \
    && echo "rule-pair identical"
rule-pair identical
```

### 낡은 주장 grep 게이트

```
$ grep -rn "Default teammate model" CLAUDE.md internal/template/templates/ .claude/rules/ docs-site/content/ \
    | grep -v "was removed\|제거되\|削除され\|已被移除"
stale-gate-exit: 1        # 무매치 — 남은 "Default teammate model" 언급은 전부 "제거되었음을 서술하는" 올바른 역사적 언급뿐
```

### Template-First 주기

```
$ make build
catalog.yaml updated successfully (12652 bytes)
go build -ldflags "... -X ...version.Commit=52c5f7ab3 ..." -o bin/moai ./cmd/moai
```
catalog.yaml은 스킬 해시 카탈로그라 rules/CLAUDE.md 변경 후 내용 불변 (git M 없음 — 정상).

### 게이트 테스트

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...
ok  github.com/modu-ai/moai-adk/internal/template  28.214s
$ go test ./internal/config/ -run "^TestAlwaysLoadedTokenBudget$|^TestAlwaysLoadedSurfaceEnumeration|^TestAlwaysLoadedTokenBudget_OverBudgetFails" -count=1 -v
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
--- PASS: TestAlwaysLoadedSurfaceEnumeration (0.01s)
--- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails (0.00s)
```
(미러 패리티·중립성 — "CC 2.1.234"/"v2.1.186"는 업스트림 버전 참조[C1 허용 클래스], SPEC ID·내부 날짜 없음 — leak-strict 테스트가 기계적으로 확인)

### docs-site

```
$ hugo --source docs-site --minify --gc 2>&1 | grep -E "WARN|ERROR"
warn-exit: 1        # 무매치 = 경고 0
$ grep -c '^#\{2,\}' .../agent-teams.md (en/ko/ja/zh)
15 / 14 / 15 / 15   # 기존 발산 상태 그대로 (문단 교체, 헤딩 불변; baseline 등재 페이지)
```

### 변경 세트

`git status --porcelain` → 정확히 8 ` M` (CLAUDE.md ×2, rule ×2, agent-teams ×4). 외부 0.

## 3. Baseline-attribution (baseline 귀속)

전부 본 워크트리(`.claude/worktrees/cc234-align`, HEAD 52c5f7ab3 + 작업 diff)에서 이번 실행으로
측정. 8개 표면의 편집 전 상태(행 내용·쌍 동일성·문단 원문)는 편집 직전 sed/grep/Read로 직독 확인.
업스트림 변경 내용은 CC 2.1.234 공식 릴리스 노트 인용 (release-update 전문가 산물
tier1-findings.md 수록분).

## 4. Gaps (미검증)

- 런타임 0히트 주장(`defaultTeammateModel|teammateModel` 부재)은 전문가 조사 산물 인용이며
  본 세션에서 재grep하지 않음 — 조사 워크트리(다른 base)에서 실증됨. 템플릿 누출 테스트 통과가
  간접 정합성을 뒷받침.
- CC 2.1.234 실제 런타임에서의 팀메이트 모델 상속 거동(문서 주장의 실측)은 본 카드에서
  수행하지 않음 — 문서는 업스트림 릴리스 노트 서술을 반영.
- ja·zh 번역의 네이티브 검수(모국어 심사) 미실시 — 작성자(오케스트레이터) 자체 검토만.
- agent-teams.md의 ko 14 vs 타 로케일 15 헤딩 발산은 기존 상태(baseline 등재)로 유지 — 본
  카드 범위 밖 (B5류 후속 파생 카드 소관).

## 5. Residual-risk (잔여 위험)

- "CC 2.1.234" 이후 업스트림이 상속 semantics를 다시 바꾸면 동일 8표면이 재정렬 대상이 됨 —
  release-update 하네스의 다음 스윕이 잡는 구조.
- 선택 GLM 문장 미포함 결정: `moai glm --team`에서 리더 모델(glm-5.3)이 팀메이트에 실제로
  전파되는지 실측이 없어 넣지 않았음 — 실측 후 docs 번들 카드에 한 문장 추가 가치 있음.
- CLAUDE.md 성장 +85B·룰 +93B (1,000B 의무 임계 미만) — always-loaded 예산 래칫 영향 미미.
