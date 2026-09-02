# SPEC-CODEX-SKILL-NEUTRAL-001 — sync-phase 독립 감사 (sync-auditor)

- 카드: t196 · 워크트리: `.claude/worktrees/t196` · 브랜치: `WT-codex-skill-neutral`
- 감사 대상 커밋: `b7a87b934` (sync-phase 3-phase close) · run base: `2c18091d1`
- 감사자: sync-auditor (독립, 회의적 입장 — 모든 주장을 자체 재측정으로 검증)
- 감사 일자: 2026-09-01
- cross-model 2차 의견 (`mcp__moai__audit_multi`): **생략** — 백엔드 rate limiting (API 429)으로 감사 재개 시 리드가 fan-out 생략을 지시. fail-open에 따른 명시적 생략이지 조용한 누락이 아니다. 본 감사의 모든 판정은 in-session 독립 재측정에 근거한다.

---

## 판정: PASS-WITH-DEBT (0.89, Tier M 임계 0.80)

| 차원 | 점수 | 근거 요약 |
|---|---|---|
| Functionality | 0.90 | 13개 AC 중 12개를 감사자가 독립 재측정해 PASS 확인. AC-CSN-010은 실질 취지(런처·미러 무변경)는 유지되나 판정 명령이 AC 본문과 불일치(결함 MAJOR-1) |
| Security | 0.95 | 결속표 문구가 검증 우회·사용자 직접 질문 유도로 읽히지 않음. 금지 토큰 0건. 배포 TOML 무변경 (golden + 직접 diff 이중 확인) |
| Craft | 0.85 | 자기참조 수치 규율·셀렉터 매치 수 명기·pre/post 쌍 관행 준수는 우수. 단 표면 집합 교체 미기록 + .log 증거 전량 미추적이 기록 공예의 부 |
| Consistency | 0.88 | CHANGELOG·progress·spec·커밋 본문의 핵심 수치 12건 전부 감사자 재측정과 일치. 사소한 스테일 2건(MINOR-1·2) |

---

## 1. Claim (주장 — 검증한 것)

1. 축 B: 템플릿 스킬 트리 `CLAUDE_SKILL_DIR` census **0**. 회귀 가드 `TestSkillTreeHasNoClaudeSkillDirToken`이 정확한 이름으로 실재하며(`internal/template/skill_dir_token_guard_test.go:41`), embedded FS를 훑고 **0파일 walk 시 가드 자체가 실패**한다(`:82-84`).
2. 축 B: 규칙 트리 잔존 집합 = **정확히 `{skill-authoring.md:219}` 1줄**, Claude-Code-only 한정자가 덧붙여진 사실 기술. 규범 문장 0건.
3. 축 B: 대체된 실행 인자 자리의 치환 경로가 실제로 해석·실행됨 — 감사자가 `check-svg.mjs`를 치환 경로로 직접 실행해 **린트 진단 5건(SVG002/060/062/063/061) 출력 + exit 1(린트 판정)**을 관측. 대조(빈 전개 형태 `node /scripts/...`)는 모듈 로더 throw, 진단 0건 — 두 상태가 명확히 갈림.
4. 축 A: `agents-codex.yaml`에 11번째 클래스 `question-channel` — 매핑 행(`:83`) + `documented-drop` 처분 행(`:144-145`)이 연속 편집으로 들어갔고, 매니페스트 로더의 강제(`manifest.go:114-125`: disposition 검증 + rationale 비면 error + 매핑→처분 행 부재 시 error)를 통과. `tool_classes` 값 집합 **11개** 재측정.
5. 축 A: `AGENTS.md` 결속표 — 3행 {question-channel, task-list, design-sync}, 서두에 부재 파생 기준 명시("a row exists only where a harness driving this contract lacks the capability"), 3열 완전 충전, 질문 채널 행 (c) 칸이 blocker 반환을 지시하며 직접 질문을 금지. 두 사본 `cmp` exit **0**.
6. 축 A 예산: `TestAlwaysLoadedTokenBudget` (비캐시 재실행) → **75,935 tokens / budget 76,000 / headroom 65 / 18 entries**. `TestCodexContractByteCeiling` → **14,774 B × 2사본 / ceiling 24,576 / headroom 9,802**. +545B = 14,774−14,229 산술 일치.
7. AC-CSN-012: acceptance.md `:149-155` **축자 명령** 재실행 → **0 / 0 / 2 / 0 / 0** — 기대값과 정확히 일치. `skill-authoring.md`의 2건이 `:45`·`:89` 프론트매터 예시 날짜와 정확히 일치. 양성 대조(같은 명령, spec.md) → **58** (0 아님 — 정규식 생존).
8. AC-CSN-009 규율: 0매치 셀렉터 `-run TestSkillDirToken`이 `no tests to run` + PASS + exit 0을 내는 것을 재현 — 판정 기록들이 정확한 이름 + 1매치를 명기한 것이 유효.
9. AC-CSN-011: `:226`·`:301` 규범 문장이 채택 설계(루트 기준 상대 경로) 지시로 교체, `:386` 예시의 Absolute-Path-OK 열이 `YES` → `NO — use project-root-relative`로 뒤집힘 — diff로 직접 확인.
10. 3-scoped-guard 배터리 (감사자 비캐시 재실행): `TestRuleTemplateMirrorDrift` **1매치 / 서브테스트 8건 전원 PASS**, `TestSkillTreeHasNoClaudeSkillDirToken` **1매치 / census 0 across 255 files**, `TestGoldenCommittedArtifactsMatchEmission` **1매치 PASS** (agentemit 패키지 전체도 `ok 4.717s`). `TestNoAskUserQuestionInSubagents` **PASS**.
11. Golden 테스트 실재성: `golden_test.go`가 `templates/.codex/agents/moai/`의 커밋된 TOML 11개를 실제 emission과 비교하는 드리프트 가드임을 본문 판독으로 확인. 실제 `git diff 2c18091d1..HEAD -- templates/.codex/` = **빈 출력** — yaml 매핑 추가가 배포물을 바꾸지 않음을 이중 확인. `AskUserQuestion`은 프론트매터 `tools:` CSV에 없음(본문 산문 3곳뿐 — plan-auditor:149, super-advisor:68, sync-auditor:141 재확인).
12. sync close: 커밋 subject에 전체 SPEC-ID 1회 + close 표현("3-phase close") 존재. sync 커밋 자체의 spec.md 변경 = `status: draft → completed` **단 1줄**. status 필드 역사(bb14186d1 +draft → 91f80cb6f 무변경 → b7a87b934 completed)로 **in-progress가 역사상 존재한 적 없음** 실측 확인. `sync_commit_sha: pending-backfill-sync` — 가짜 SHA가 아닌 규약 placeholder(D3 면제). 백필 커밋은 아직 미생성(HEAD가 sync 커밋 — 예고된 후속 작업).
13. CHANGELOG: AC 개수 **13** (acceptance.md와 일치), 인용 경로 실재, "Not demonstrated here" 4항목 정확 — `go test ./...` 미실행 명시, **미푸시 브랜치라 CI 판정 부재** 명시(거짓 CI 주장 없음), dogfood 어긋남 unconfirmed 명시, 범위 밖 잔여 보고 명시.
14. 일관성: census 0(템플릿 스킬) / 1(템플릿 전체) / 46(로컬 dogfood) / 3(로컬 규칙 잔존) — CHANGELOG·progress·커밋 본문 수치와 감사자 재측정 전부 일치.

## 2. Evidence (증거 — 명령 + 축자 출력)

각 Claim의 명령은 감사자가 이 트리(HEAD `b7a87b934`)에서 직접 실행했으며 출력을 관측했다. 핵심 축자:

```
$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills | wc -l
0

$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/rules
internal/template/templates/.claude/rules/moai/development/skill-authoring.md:219:| `${CLAUDE_SKILL_DIR}` | Absolute path to the skill's own directory. Claude Code only — a harness that does not export it expands the reference to nothing | v2.1.69 |

$ node .claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs /tmp/t196-audit-probe.svg
.../t196-audit-probe.svg:1:1  error  SVG002  root <svg> has no viewBox attribute
(이하 SVG060/062/063/061 — 5 errors, 0 warnings)
---exit: 1        ← 린트 판정 (경로 해석 성공의 증거)
$ node /scripts/check-svg.mjs /tmp/t196-audit-probe.svg
node:internal/modules/cjs/loader:1228  throw err;   ← 대조: 진단 0건

$ grep -cE '<AC-CSN-012 정규식>' AGENTS.md …/AGENTS.md …/skill-authoring.md …/worktree-integration.md internal/template/agentemit/agents-codex.yaml
AGENTS.md:0 / templates AGENTS.md:0 / skill-authoring.md:2 / worktree-integration.md:0 / agents-codex.yaml:0

$ go test -count=1 ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v
token_budget_guard_test.go:69: always-loaded surface = 75935 tokens (budget 76000, headroom 65, 18 entries)
--- PASS

$ go test -count=1 ./internal/config/ -run 'TestCodexContractByteCeiling' -v
AGENTS.md = 14774 bytes (ceiling 24576, headroom 9802)   × 2 사본
--- PASS

$ go test -count=1 ./internal/template/ -run 'TestSkillTreeHasNoClaudeSkillDirToken' -v
census: 0 lines carrying CLAUDE_SKILL_DIR across 255 files walked under .claude/skills
--- PASS

$ go test -count=1 ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -v   → 1매치, 서브테스트 8건 전원 PASS
$ go test -count=1 ./internal/template/agentemit/ -run 'TestGoldenCommittedArtifactsMatchEmission' -v   → 1매치 PASS
$ go test -count=1 ./internal/template/agentemit/   → ok 4.717s
$ go test -count=1 ./internal/template/ -run 'TestNoAskUserQuestionInSubagents' -v   → PASS
$ go vet ./internal/template/ ./internal/config/ ./internal/template/agentemit/   → exit 0

$ git diff --stat 2c18091d1..HEAD -- internal/cli/codex_launcher.go internal/template/skill_mirror.go internal/template/agentemit/
 internal/template/agentemit/agents-codex.yaml | 12 ++++++++++++      ← 비어있지 않음 (MAJOR-1)
$ git diff --stat 2c18091d1..HEAD -- internal/cli/codex_launcher.go internal/template/skill_mirror.go internal/template/renderer.go
(출력 없음, exit 0)                                                    ← progress.md가 돌린 교체 명령

$ git ls-files .moai/reports/t196/ | wc -l   → 14   (.log 0건 — MAJOR-2)
$ git check-ignore -v .moai/reports/t196/codex-behavior.log
.gitignore:106:*.log	.moai/reports/t196/codex-behavior.log

$ git log -p 2c18091d1..HEAD -- …/spec.md | grep -E '^(COMMIT|[+-]status:)'
bb14186d1: +status: draft → 91f80cb6f: (무변경) → b7a87b934: -draft +completed

$ go test -count=1 ./internal/template/ -run 'TestSkillDirToken' -v
testing: warning: no tests to run   → PASS, exit 0 (0매치 셀렉터의 초록 재현)
```

기타 축자: `git diff --stat 2c18091d1..HEAD -- templates/.codex/` 빈 출력 · `cmp AGENTS.md internal/template/templates/AGENTS.md` exit 0 · 양성 대조 58 · census 46(로컬 스킬)/1(템플릿 전체)/3(로컬 규칙).

## 3. Baseline-attribution (baseline 귀속)

- 모든 측정은 **이 감사 실행 중, 이 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t196`), HEAD `b7a87b934`** 에서 수행. 캐시 무효화가 필요한 go test는 전부 `-count=1`로 재실행 (캐시 히트 1건을 이후 `-count=1`로 재확인).
- AC-CSN-012·AC-CSN-009·AC-CSN-010 명령은 acceptance.md/progress.md가 지정한 축자 형태를 그대로 사용 (AC-CSN-010은 원본 명령과 교체 명령을 **둘 다** 실행해 대조).
- 사전값(46/6/50/4줄, 75,799 tokens, 14,229 B)은 착수 시점 측정이라 감사자가 재측정 불가 — `2c18091d1` 시점 기록에 의존. 다만 사후값 + 산술 차분(+545B = 14,774−14,229, 201 tokens = 804 B)의 정합으로 간접 검증했다.

## 4. 결함 목록

### MAJOR-1 — AC-CSN-010 판정이 AC 본문의 명령이 아닌 교체 명령으로 내려졌고, 교체가 기록되지 않았다
- acceptance.md `:108-110`의 AC-CSN-010 표면은 `codex_launcher.go` · `skill_mirror.go` · **`agentemit/`** 다. progress.md §E.4 sync 재판정은 이 중 `agentemit/`를 빼고 **`renderer.go`를 새로 넣은** 다른 명령을 돌려 빈 출력을 근거로 PASS를 기록했다. 감사자가 **원본 명령을 착수-base `2c18091d1..HEAD`로 재실행하면 비어 있지 않다** — `agents-codex.yaml +12`.
- base SHA 교체는 AC의 [HARD]가 지시한 대로 progress.md에 기록됐지만, **표면 집합 교체는 한 글자도 기록되지 않았다** — 이 AC 자신이 못박은 "조용히 교체하지 않는다"의 정신을 표면 축에서 위반. 커밋 본문의 "all three out-of-scope surfaces"는 어느 3표면인지 명시하지 않아 두 집합(acceptance 원본 vs progress 교체)이 갈린 채 모호하게 남는다(MINOR-2와 결합).
- **완화 사실**: 실질 위반은 없다 — yaml 매니페스트 편집은 v0.3.1 §B.D7(리드 판정)이 명시적으로 합법화한 범위 확장이고, EmitAll 항등 변환·배포 TOML은 golden + 직접 diff로 무변경 확인. REQ-CSN-014(런처)는 유지. 따라서 이것은 "거짓 PASS"가 아니라 **AC 본문의 진화(범위 확장 반영 실패)를 판정 기록이 감춘 기록 무결성 결함**이다. 이 SPEC이 plan-phase에서 사고의 원인으로 이름 붙인 바로 그 형태("조용한 교체")가 판정 기록에 재발한 것이라 가중한다.
- 권고 조치: progress.md §E.4에 원본 명령의 출력(비어있지 않음)과 교체 사유·교체 시점을 명시 기록. acceptance.md AC-CSN-010의 표면 갱신(agents-codex.yaml 제외, 또는 "매니페스트 데이터 파일 제외 — §B.D7")은 다음 SPEC 본문 편집 배치(또는 백필 커밋 이후 후속 카드)의 몫.

### MAJOR-2 — progress.md가 인용하는 모든 .log 증거 파일이 저장소에 들어가지 않는다
- `git ls-files .moai/reports/t196/` → **14건 전부 .md/.sh/.py/.txt, .log 0건**. `.gitignore:106`의 `*.log` 규칙이 `m1-probe-{a,a2,b}.log` · `m4-guard-red-*.log`·`m4-guard-green-*.log`(AC-CSN-009의 RED/GREEN 관측) · `m3-ac012-post.log` · `m3-cwd-sweep.log` · `m3-path-resolution.log` · `m3-hard-site-execution.log` · `codex-behavior.log`(AC-CSN-001이 지명하는 경로) · `axis-a-verdicts.log`(축 A 5건 증거)를 전부 배제했다. `git check-ignore -v`로 규칙 출처 확인.
- progress.md `:187`은 "`.moai/reports/t196/` 증거 디렉터는 progress 갱신 커밋에 함께 스테이징한다"고 적었으나 .log에 대해서는 **실제로 성립하지 않았다** — 디렉터 단위 add가 ignored 파일을 조용히 걸렀다. 브랜치가 아직 미푸시이고 워크트리가 증거의 유일본이므로, develop 병합 후 워크트리 폐기 시 **모든 .log 증거가 영구 유실**되고 progress.md·CHANGELOG의 인용 경로가 미래 독자에게 단절된다 (`agent-common-protocol.md` § Evidence persistence — cited path must still resolve at audit time —의 예측 결함).
- **완화 사실**: AC-CSN-009 RED의 핵심 정체성 관측(sha256 before/after 쌍, census 0→1→0, exit 1→0)은 progress.md 본문과 sync 커밋 메시지에 요약 기록돼 있고, AC-CSN-001의 축자 발췌+판정 문장은 progress.md와 CHANGELOG에 담겨 있다. 즉 판정이 증거 없는 것은 아니나, 축자 전문의 보존 계층이 사라진다.
- 권고 조치: (a) `.gitignore`에 `!*.log` 예외를 `.moai/reports/**` 범위로 추가 후 증거를 커밋, 또는 (b) 핵심 .log 5건(AC-CSN-001/009/012 관련)을 `.md`/`.txt`로 변환 커밋. 어느 쪽이든 **병합 전**에 수행해야 효력이 있다.

### MINOR-1 — spec.md §A.5의 `tool_classes` 값 집합 "10개"는 현재 11개다
- v0.3.1 시점(question-channel 추가 전) 측정값이지만 시점 라벨 없이 "이 트리 실측"으로 적혀 있어, 지금 재측정하면 11이 나온다. §B.D7이 "11번째 클래스"로 부르므로 문서 내부 논리는 정합이나, §A.5의 측정 시점 표기가 없어 미래 재측정 독자가 모순으로 읽을 수 있다.

### MINOR-2 — "all three out-of-scope surfaces"의 집합이 문서마다 다르다
- acceptance.md 원본 3표면 = {런처, 미러, agentemit/}, progress.md §E.4 = {런처, 미러, renderer}. sync 커밋 본문과 CHANGELOG는 집합을 명시하지 않아 어느 쪽인지 알 수 없다. MAJOR-1의 기록 성격상 병행 수리 대상.

### MINOR-3 — `m3-govet.log`가 0바이트다
- go vet 성공 시 출력이 없는 것이 자연스럽지만 exit 코드 기록이 없어 파일 자체로는 증거 능력이 없다. 감사자가 `go vet` 재실행으로 exit 0을 독립 관측해 결손을 메웠다.

### INFO — 이미 정직하게 기록된 항목 (결함 아님, 확인 기록)
- dogfood census 46 vs 템플릿 0 어긋남 — 바이너리 지연 가설, unconfirmed로 명시 ✓
- 로컬 `.claude/rules/moai/development/skill-authoring.md` `:226`·`:301` 규범 잔존 — 범위 밖 보고 ✓ (감사자 재측정 3줄 일치)
- draft→in-progress 스킵 — §E.4 기록, 역사 실측으로 확인 ✓
- `sync_commit_sha` 백필 미생성 — 예고된 후속 작업 (HEAD가 sync 커밋) ✓

## 5. Gaps (감사자가 관측하지 않은 것)

- **run-phase 사전값 6건**(census 46/6/50, 규칙 4줄, 예산 75,799, 바이트 14,229)의 원시 재측정 — 시점이 지난 측정이라 불가. 산술·사후 정합으로 간접 검증한 것이 전부다.
- **AC-CSN-001의 원 관측 재실행** — codex 세션 프로브(m1-probe-a/a2)를 감사자가 재실행하지 않았다. 발췌가 출처 로그와 일치함(세션 ID + 축자 문구)까지 확인했고, 편찬 파일이 "재실행 없음, 귀속 리드"를 명시하므로 AC의 요건(관측 기록 존재 + 판정 문장)은 충족으로 판단했다.
- **`m3-path-resolution.log`의 34경로 전수 독립 재현** — spot-check(3실행 자리 + 3스크립트 실재 + 1실행 관측)로 대체했다.
- **windows/linux 교차 플랫폼** — progress.md §E.3이 `not-observed`로 정직 선언. 감사자도 관측하지 않았다.
- **cross-model 감사** — 리드 지시로 생략(본문 서두 기재).
- **`go test ./...` 전체 스위트** — 로컬 전체 스위트 금지 규율에 따라 미실행. 전 패키지 판정은 CI 몫이며 브랜치가 미푸시라 CI 판정도 아직 없다.

## 6. Residual-risk (잔여 위험)

- **MAJOR-2의 유실은 아직 발생하지 않았다** — 병합 전에 증거 커밋이 이뤄지면 소멸하는 위험이다. 반대로 병합 창이 먼저 열리면 progress.md의 인용 절반(경로)이 죽은 좌표가 된다. 병합 순서가 이 결함의 실현 여부를 결정한다.
- 결속표의 **실효성**(코덱스가 표를 실제로 읽고 blocker를 반환하는가)은 이 SPEC이 애초에 판정하지 않기로 한 것(acceptance.md §D.3(1)) — 대조 실행 하네스가 갖춰질 때까지 미판정 상태가 지속된다.
- 로컬 `.claude/skills/**`(dogfood 46줄)은 설치 바이너리가 갱신될 때까지 템플릿과 어긋난 채다 — `moai update` 실행 시점에 어긋남이 해소되는지는 관측 전이다.
- spec.md §A.5의 스테일 수치(MINOR-1)와 AC-CSN-010 본문의 스테일 표면(MAJOR-1 후반부)은 다음 본문 편집 배치 전까지 남는다 — 지금 감사가 기록으로 남겼으므로 조용한 결함은 아니다.

---

## 판정 근거 요약

핵심 기능(census 0, 규범 문장 0, 결속표 3행 부재 파생, 예산 65/9,802 여유, 3-scoped-guard 배터리)은 **감사자의 독립 비캐시 재측정으로 전부 재현**됐고, 수치 일관성 12건 전부 일치하며, 보안 렌즈(결속표 문구·금지 토큰·배포 TOML 무변경)에서 결함이 없다. 반면 AC-CSN-010의 판정 명령 교체가 기록 없이 이뤄졌고(MAJOR-1), 인용된 .log 증거 전량이 gitignore로 저장소 밖에 있다(MAJOR-2) — 둘 다 판정의 참/거짓을 바꾸지는 않으나 이 SPEC이 스스로 세운 기록 무결성 기준(§D.4, "조용히 교체하지 않는다")에 걸린다. 기능·보안·정합은 유지되고 결함이 전부 기록·보존 계층이므로 **PASS-WITH-DEBT 0.89** (Tier M 임계 0.80 상회, 부채 2건 MAJOR + 3건 MINOR 명시).

🗿 MoAI · sync-auditor · 2026-09-01
