auditor-model: claude-sonnet-5-5[1m]

verdict: PASS-WITH-DEBT
audited_sha: 1f4b052b305b5e5d0efa8d60646e0a8b2cd155fc

# sync-audit 판정 — SPEC-CC-ULTRACODE-TOGGLE-001 (카드 t1416, Tier M)

감사 대상: 워크트리 `.moai/worktrees/t1416`, 브랜치 `WT-ultracode-toggle-wording`, HEAD `1f4b052b3`(전체 SHA는 위 기계 줄), 카드 기점 `c50da9c2f`. 감사 시작 시 `git status --short`는 비어 있었다(깨끗한 트리).

종합 점수 92/100, 판정 PASS-WITH-DEBT. 블로킹 결함은 없다. 부채는 모두 optional로 분류했고, 가장 무거운 것은 F1(설정 키 경로의 버전 한정 표현이 업스트림 문장보다 강하다)이다.

## 1. 차원 점수 (sync-auditor 척도)

| 차원 (가중) | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | AC-001..AC-010 핀 전부 재실행 통과, AC-011만 lane 기록 인용(§3 미검증). 업스트림 원문 대조 결과 내용은 사실과 맞고 F1만 표현이 과하다 |
| Security (25%) | 96/100 | PASS | 산문 변경뿐. 추가 줄에 URL 0, 비밀 패턴 0, 의존성 매니페스트 diff 0. OWASP 해당 표면 없음 |
| Craft (20%) | 88/100 | PASS | spec lint 0 finding, hugo 경고 0, 증거 기록 충실. 핀이 어휘 보존형 오구현을 못 막는 부분(F4), 낡은 문장(F6), ko/ja 표현(F2) 감점 |
| Consistency (15%) | 90/100 | PASS | 미러 바이트 동일, 4개 로케일 구조 동일, 헤딩 수 불변. 상태 전환 시점 편차(F7), ko 콜아웃 불일치(F3) 감점 |

종합: 가중 조화평균 91.8, 가중 산술평균 91.9, 보고 값 92. must-pass 방화벽(Functionality와 Security 각각 단독 통과)은 둘 다 충족. 기본 프로파일(`.moai/config/evaluator-profiles/default.md`)은 sync 전용 수치 임계를 두지 않고 must-pass만 둔다. 브리프가 말한 Tier M 임계는 `spec-workflow.md` § SPEC Complexity Tier의 0.80으로 읽었고, 92/100은 이를 넘는다. `harness.yaml`에 `evaluator_mode: hierarchical`이 없어(`grep` 결과 `evaluator_mode`는 `final-pass`와 `per-sprint`뿐) 평면 가중 모드로 채점했다.

## 2. AC 표 (AC-001..AC-011)

"변이 통과"는 §4 변이 실험에서 그 AC의 핀을 모두 통과하면서 요구사항을 어기는 오구현을 만들 수 있었는지를 뜻한다.

| AC | 판정 | 재관측 값 (경로별) | 변이 통과 |
|---|---|---|---|
| AC-001 규칙 원본 결합 제거와 정본 문구 | PASS | R1s 0,0(exit 1); R1c 1,1; R1b 0,0(기점 트리에서는 1); R2 R3 R4 1,1 | 가능: RS-M1 RS-M3 RS-M4 (§4) |
| AC-002 끄는 경로, 실행 플래그 예외, 현재 세션, 설정 키 | PASS | R12 0,0(기점 1); R5 R6 R7 R8 R10 R11 1,1; R9 1,1 | 가능: RS-M2 RS-M7 |
| AC-003 미러 동일 | PASS | `cmp RS RM` 출력 없음, exit 0 | 한쪽만 고치면 깨짐(핀 설계상) |
| AC-004 workflows 행 4개 로케일 | PASS | W1 W2 W3 W4 W8 각 ko,en,ja,zh 1; W5 W6 각 0; W7 D2 각 1; W10 각 1 | 가능: WF-M1 WF-M2 WF-M3 WF-M6 (§4). WF-M4 WF-M5 WF-M7은 핀이 잡음 |
| AC-005 multi-llm 주석 | PASS | M1 0,0,0,0(exit 1); M2 1,1,1,1 | 변이 미시도(핀이 줄 단위 정규식) |
| AC-006 ultracode-workflows 페이지 | PASS | U1 0,0,0,0; U1c 4,0,0,0; U2 0 전부; U3 1 전부; U4 0; U5 1 | 가능(산문형 재서술은 U1c 밖) — 선언된 잔여 |
| AC-007 commands 페이지 | PASS | C1* 0 전부; C2en C2ko 0; C3 en 2, ko 2, ja 1, zh 1 | 변이 미시도 |
| AC-008 헤딩 수 | PASS | H-WF 11×4, H-ML 13×4, H-UW 24/20/20/20, H-CM 23/23/18/18, 베이스라인 파일 diff 없음 | — |
| AC-009 출구 게이트 | PASS | hugo exit 0, `WARN|ERROR` 0줄(exit 1), URL 금지 grep 출력 없음(exit 1), Mermaid 방향 grep 출력 없음(exit 1) | — |
| AC-010 범위와 slider | PASS | R13 0,0; W11 16개 파일 전부 0; 범위 밖 파일 `git diff --stat` 출력 없음 | 단어 `slider`를 피한 지속성 주장은 통과(RS-M5) |
| AC-011 템플릿 embed | 미재실행(인용) | `make build`와 `go test ./internal/template/`는 progress.md §E.2 인용. 독립 부분 관측: 이 트리에서 빌드된 `bin/moai`에 새 문구 1회, 옛 문구 0회 | — |

AC-011은 회귀 방지용 AC이며 릴리스 차단이 아니다. lane 기록은 첫 `go test` 실행이 기본 10분 제한에 걸린 TOOL_FAILURE였고 `-timeout 45m` 재실행에서 통과했다고 적는다. 이 감사는 그 두 실행을 재현하지 않았다(§3 미검증).

## 3. 5-섹션 증거 블록

### 3.1 주장 (Claim)

C1. 규칙 원본(L111)과 미러는 `/effort ultracode`를 독립 on/off 토글로 서술하고 effort 레벨을 바꾸지 않는다고 말하며, `xhigh`는 실행 플래그 절에 한 번만 나온다.
C2. 4개 로케일의 workflows 표 행 4개, multi-llm 주석 4개, ultracode-workflows 페이지 4개, commands 페이지 4개가 같은 사실을 담고 구조(열 수, 헤딩 수)를 유지한다.
C3. 편집된 문장의 사실은 업스트림 1차 자료와 맞는다. 단 설정 키 경로의 버전 표현(F1)은 원문보다 강하다.
C4. 슬라이더 지속성(OQ-1)을 편집 텍스트가 주장하지 않는다.
C5. 범위는 SPEC 변경 지도와 SPEC 디렉터리, `.moai/reports/t1416/`로 한정된다. CHANGELOG는 건드리지 않았고 그것은 결함이 아니다.
C6. 자율 Kickoff의 보상 통제(결정 기록)는 §9.1 조건을 실제 증거로 뒷받침한다.

### 3.2 증거 (Evidence) — 명령과 원문 출력

표기: `RS`는 `.claude/rules/moai/workflow/dynamic-workflows.md`, `RM`은 `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md`. (A)와 (C)의 W1·W5 블록은 명령 줄의 경로 변수만 줄인 원문 줄이다. 그 밖의 블록은 줄을 합치거나 경로를 로케일 이름으로 줄이고 JSON을 발췌한 전사다. 전체 원문은 머신 로컬 스크래치에 있다(§3.5 손실 인지). 모든 CJK 핀은 `LC_ALL=en_US.UTF-8`(출력 첫 줄에 `LC_ALL=en_US.UTF-8`로 찍힘)에서 실행했다. 이 머신의 `grep`은 `/usr/bin/grep`, `grep (BSD grep, GNU compatible) 2.6.0-FreeBSD`이며 SPEC 장부가 전제한 ugrep이 아니다. lane 기록(ugrep)과 이 재측정(BSD grep)이 같은 값을 냈다.

**(A) 규칙 원본과 미러, HEAD 1f4b052b3**

```
[R1s] grep -c -E -- 'xhigh.*xhigh' $RS $RM
.claude/rules/moai/workflow/dynamic-workflows.md:0
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
exit=1
[R1c] grep -c -F -- xhigh $RS $RM
.claude/rules/moai/workflow/dynamic-workflows.md:1
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:1
exit=0
[R1b] grep -c -F -- 'combines `xhigh` reasoning' $RS $RM
.claude/rules/moai/workflow/dynamic-workflows.md:0
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
exit=1
[R2] grep -c -F -- 'leaves the effort level unchanged' $RS $RM
.claude/rules/moai/workflow/dynamic-workflows.md:1
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:1
exit=0
[R7] grep -c -F -- 'starts the session at `xhigh`' $RS $RM
.claude/rules/moai/workflow/dynamic-workflows.md:1
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:1
exit=0
[R11] grep -c -E -- 'v2.1.284.*"ultracode": true|"ultracode": true.*v2.1.284' $RS $RM
.claude/rules/moai/workflow/dynamic-workflows.md:1
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:1
exit=0
[R12] grep -c -F -- 'step back with `/effort high`' $RS $RM
.claude/rules/moai/workflow/dynamic-workflows.md:0
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
exit=1
[R13] grep -c -i -E -- slider $RS $RM
.claude/rules/moai/workflow/dynamic-workflows.md:0
internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md:0
exit=1
[X-cmp] cmp $RS $RM
exit=0
```

나머지 R 핀(RS와 RM 순서, 둘 다 같은 값, 모두 exit 0): R3 `independent on/off toggle` 1,1; R4 `v2.1.284` 1,1; R5 `/effort ultracode off` 1,1; R6 `--effort ultracode` 1,1; R8 `"ultracode": true` 1,1; R9 `ultrathink.` 1,1; R10 `current session` 1,1.

**(B) 기점 트리 c50da9c2f에서 RED 재관측** (`git show c50da9c2f:<파일>`을 스크래치로 내려받아 같은 핀을 실행)

```
[R1b-base] 1  exit=0        [R2-base] 0  exit=1
[R7-base]  0  exit=1        [R12-base] 1 exit=0
[W1-base] ko:0 en:0 ja:0 zh:0  exit=1
[W4-base] ko:0 en:0 ja:0 zh:0  exit=1
[W6-base] ko:1 en:1 ja:1 zh:1  exit=0
```

기점에서 붉고 HEAD에서 푸르므로 핀은 이 작업이 뒤집는다.

**(C) docs-site 핵심 핀**

```
[W1] grep -c -E -- '--effort ultracode[^|.。;；,，、]*xhigh' <workflows.md ko,en,ja,zh>
docs-site/content/ko/claude-code/agentic/workflows.md:1
docs-site/content/en/claude-code/agentic/workflows.md:1
docs-site/content/ja/claude-code/agentic/workflows.md:1
docs-site/content/zh/claude-code/agentic/workflows.md:1
exit=0
[W5] grep -c -E -- 'xhigh.*xhigh|xhigh.*--effort ultracode' <workflows.md ko,en,ja,zh>
docs-site/content/ko/claude-code/agentic/workflows.md:0
docs-site/content/en/claude-code/agentic/workflows.md:0
docs-site/content/ja/claude-code/agentic/workflows.md:0
docs-site/content/zh/claude-code/agentic/workflows.md:0
exit=1
[W1 매칭 절, grep -o]
--effort ultracode` 실행 플래그는 세션을 `xhigh
--effort ultracode` launch flag also starts the session at `xhigh
--effort ultracode` 起動フラグはセッションを `xhigh
--effort ultracode` 启动参数会同时让会话以 `xhigh
[W1 LC_ALL=C]  ko: 0 exit=1 / ja: 0 exit=1 / zh: 0 exit=1
[M1] ^/effort ultracode # .*xhigh  ko:0 en:0 ja:0 zh:0  exit=1
[U1c] xhigh 포함 줄  ko:4 en:0 ja:0 zh:0  exit=0
[C3] ko:2 en:2 ja:1 zh:1 (각 exit 0)
```

W2 W3 W4 W8 W7 D2 W10은 로케일마다 1(exit 0), W6 W11은 전부 0(exit 1)이었다. 같은 환경에서 U1 0×4, U2 0×4, U3 1×4, U4 0, U5 1, C1* 0, C2* 0, M2 1×4, 헤딩 수는 위 AC 표의 값이다.

**(D) D1과 D2의 독립 핀** (SPEC 핀 밖에서 내가 추가한 것, 각 로케일 파일에서 표 행 한 줄에 대해 `^\| \`/effort ultracode\` \|.*<단어>`)

```
D1 토글 단어: ko(토글) 1  en(toggle) 1  ja(トグル) 1  zh(开关) 1
D1 레벨 불변: ko(바뀌지 않) 1  en(leaves the effort level unchanged) 1  ja(変わらず) 1  zh(不会改变) 1
D2 두 열 유지: ^\| `/effort ultracode` \|[^|]*\|$  ko:1 en:1 ja:1 zh:1
```

**(E) 변이 실험** (스크래치 복사본에서만, 워크트리 미접촉. 규칙 변이는 실제 L111 불릿 위에서, 행 변이는 실제 en 행 위에서 만듦)

```
RS delivered (control)                         PASS (all release-blocking pins green)
RS-M1 appended coupling w/o token xhigh        PASS (all release-blocking pins green)
RS-M2 flag clause rewritten as general coupling PASS (all release-blocking pins green)
RS-M3 wrong intro version v2.1.248             PASS (all release-blocking pins green)
RS-M4 negated 'leaves ... unchanged'           PASS (all release-blocking pins green)
RS-M5 slider persistence claim w/o word slider PASS (all release-blocking pins green)
RS-M6 slider persistence claim with word slider FAIL (tripped: R13)
RS-M7 inverted 'different level does not turn off' PASS (all release-blocking pins green)
WF-en delivered (control)                      PASS (all release-blocking pins green)
WF-M1 drop 'effort level unchanged' (D1)       PASS (all release-blocking pins green)
WF-M2 drop 'toggle' wording (D1)               PASS (all release-blocking pins green)
WF-M3 comma-free relocation (G3 class)         PASS (all release-blocking pins green)
WF-M4 wrong version on key line                FAIL (tripped: W3,W8)
WF-M5 third column added                       FAIL (tripped: D2)
WF-M6 command-form persistence claim flipped   PASS (all release-blocking pins green)
WF-M7 off route replaced w/o literal           FAIL (tripped: W4)
```

행 변이의 핀 평가는 표 행 한 줄에 대해 정규식으로 했다(실제 핀은 파일 전체에 대한 `grep -c`이지만 W 핀이 걸리는 줄은 그 행뿐이라 결과가 같다). RS-M2는 `xhigh`가 플래그 절 자리에 남은 채 "Ultracode always starts the session at `xhigh`"로 일반 결합을 되살린 변이이고, RS-M3은 도입 버전을 `v2.1.248`로 틀리게 쓴 변이다(키 줄에 `v2.1.284`가 남아 R4가 계속 1줄).

**(F) 업스트림 1차 자료 원문** (2026-10-02 curl 후 태그 제거. 스크래치에서 읽었고 아래에 결정 문장만 인용한다)

- `https://code.claude.com/docs/en/model-config` HTTP 200, 948998 바이트. "Turning ultracode on or off with /effort or the ultracode setting leaves the effort level unchanged. The --effort ultracode flag and the Agent SDK effortLevel: "ultracode" value turn it on and also set the level to xhigh . Picking a level in the /effort slider or the /model picker leaves ultracode as it was."
- 같은 페이지: "The /effort ultracode off form, the slider toggle, and keeping ultracode on at effort levels other than xhigh require Claude Code v2.1.284 or later. Before v2.1.284, turning on ultracode set the session to xhigh effort, picking another level turned it off, and an effort cap below xhigh made it unavailable."
- 같은 페이지: "/effort : run /effort ultracode to turn it on for the current session or /effort ultracode off to turn it off."
- `https://code.claude.com/docs/en/workflows` HTTP 200. "/effort ultracode lasts for the current session; to have every session start with it, set the ultracode setting. Turn it off with /effort ultracode off when you return to routine work."
- `https://code.claude.com/docs/en/settings-reference` HTTP 200, `ultracode` 항목: "The key doesn't change the session's effort level: ultracode runs at whichever level the session uses. Claude Code reads this key but never writes it: /effort ultracode turns ultracode on for the current session only." 그리고 "This and the /effort ultracode off form require Claude Code v2.1.284 or later. Before v2.1.284, ultracode: true ran the session at xhigh effort, and an effort cap below xhigh kept ultracode off."
- `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` HTTP 200, `## 2.1.284` 절(파일 342행 시작) 406행: "Changed Ultracode into its own toggle in `/effort` (Tab, or `/effort ultracode [on|off]`): it no longer forces xhigh effort and stays on at any effort level".

**(G) 빌드와 도구 출력**

```
hugo --minify --gc --destination <scratch>   (cd docs-site)   hugo_exit=0
grep -c -E 'WARN|ERROR' <hugo.log>           0   exit=1
 Pages        │ 189 │ 187 │ 187 │ 187
./bin/moai spec lint SPEC-CC-ULTRACODE-TOGGLE-001
✓ No findings — all SPEC documents are valid
lint_exit=0
./bin/moai spec lint SPEC-CC-ULTRACODE-TOGGLE-001 --json --strict
[]
lint_json_exit=0
./bin/moai spec audit --base-dir <worktree> --filter-spec SPEC-CC-ULTRACODE-TOGGLE-001 --json   (JSON 발췌, 줄을 합쳐 옮김)
 "total_specs": 1, "grandfathered": 0, "modern_era_clean": 1,
 "finding_type": "EraAutoDetected", "severity": "INFO",
 "heuristic_matched": "H-4 (§E.2 + §E.4 + sync_commit_sha)"
audit_exit=0
grep -a -o -F 'leaves the effort level unchanged' bin/moai | wc -l   1
grep -a -o -F 'combines `xhigh` reasoning' bin/moai | wc -l          0
```

렌더된 hugo 출력에서 `/effort ultracode off` 문자열이 든 페이지는 10개(ko/en/ja/zh의 ultracode-workflows와 workflows, ko와 en의 commands)였다. ja/zh commands는 설계상 이 문자열이 없다.

**(H) 범위와 방화벽**

```
git diff --stat c50da9c2f HEAD -- <session-handoff 3종> <moai.md> CHANGELOG.md <appendix 2종> docs-site/.locale-parity-baseline
(출력 없음)
git diff --name-only c50da9c2f HEAD
 → 규칙 파일 2, docs-site 16, .moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/ 4파일, .moai/reports/t1416/ 4파일. 그 밖의 경로 없음
git diff --stat abf8ba627 HEAD -- internal cmd go.mod go.sum Makefile
(출력 없음)
추가 줄(docs.diff 19줄) 스캔: http(s):// 0줄, 헤딩 줄 0줄, 이모지 0줄, 비밀 패턴 0줄
```

**(I) 커밋 이력과 소유권**

```
git log c50da9c2f..HEAD (12개) — 본문에 `Authored-By-Agent:` 줄이 있는 커밋
 e7dbdeef3 manager-spec | 0e7b6af5b plan-auditor | d7e97e0f4 manager-spec | ff7b64f6e plan-auditor
 3d44379b8 manager-spec | 2ffcbcf7e plan-auditor | abf8ba627 manager-develop | be5d56b8c manager-develop
 90d5bac45 manager-develop | 211905e37 manager-docs | 1f4b052b3 manager-docs
 0e7087faa (결정 기록, 줄 없음 — 오케스트레이터 직접)
git log --format='%h|%(trailers:key=Authored-By-Agent,valueonly)' → 12개 모두 값이 비어 있음
git log -G'^status:' c50da9c2f..HEAD -- spec.md → 211905e37, 90d5bac45, e7dbdeef3
git diff --stat 3d44379b8 0e7087faa -- .moai/specs   (출력 없음)
git diff --stat 3d44379b8 HEAD -- .moai/specs
 progress.md | 210 +++++-  /  spec.md | 2 +-
```

`Authored-By-Agent:` 줄 다음에 빈 줄과 `🗿 MoAI` 문단이 와서 git의 trailer 파서는 이를 trailer로 보지 않는다(`%(trailers:...)` 값이 비어 있는 이유). 반면 `internal/spec/lint_ownership.go:256`의 `authoredByAgentLine` 정규식은 본문 어디서든 줄 단위로 읽으므로 린트 쪽 소유권 규칙에는 영향이 없다.

### 3.3 Baseline 귀속 (Baseline-attribution)

- 측정 트리: HEAD `1f4b052b305b5e5d0efa8d60646e0a8b2cd155fc`, 이 실행에서 이 워크트리를 대상으로 측정.
- 도구 빌드 귀속(verification-claim-integrity §2.2): `moai spec lint`와 `moai spec audit`는 설치본이 아니라 워크트리의 `./bin/moai`로 돌렸다. 이 바이너리의 `version` 출력은 `moai_cp/20260925_122548-1904-gabf8ba627-dirty   built 2026-10-02T06:01:52Z`이고 `git merge-base --is-ancestor abf8ba627 HEAD`는 exit 0이다. `abf8ba627`에서 HEAD까지 `internal cmd go.mod go.sum Makefile`의 diff가 비어 있어 Go 표면은 HEAD와 같다. 빌드 시점의 `dirty`가 무엇이었는지는 이 감사가 풀지 못했다(§3.4). lane 기록이 적은 설치본(`c50da9c2f`)은 근거로 쓰지 않았다.
- hugo: `/opt/homebrew/bin/hugo`, 출력은 스크래치.
- 업스트림 페이지와 changelog: 2026-10-02에 curl로 이 실행에서 받았다.
- 기점 RED는 `c50da9c2f`의 `git show`로 이 실행에서 추출했다. SPEC 장부의 RED 값을 옮겨 쓰지 않았다.

### 3.4 미검증 (Gaps)

1. `make build`와 `go test ./internal/template/...`는 이 감사에서 실행하지 않았다(브리프 지시와 머신 부하). progress.md §E.2의 결과(첫 실행 기본 10분 제한 초과, `-timeout 45m` 메인 패키지 재실행 통과 349.908s, 하위 패키지는 첫 실행 통과)는 인용일 뿐이다. 내가 관측한 것은 바이너리 안의 문구 존재(새 문구 1, 옛 문구 0)뿐이다. 비용: `grep -a -o`로 74 MB 파일 한 번 읽기.
2. 크로스 백엔드 감사(`audit_multi`, `claude_audit`, `codex_audit`, `glm_audit`)는 실행하지 않았다. 브리프가 요청하지 않았고, `.moai/config/sections`에서 `audit_model`을 찾지 못했으며(`grep` 출력 없음), `workflow.audit.gates`의 명시 설정도 없다. 따라서 영수증을 발급받지 않았다.
3. ja와 zh 문장의 모국어 판단은 하지 못했다. 비모국어 독자로서 읽었다(§5 F2). ko는 한국어 모국어 수준으로 읽었으나 `moai-domain-humanize` 통과 검사는 돌리지 않았다.
4. OQ-1(슬라이더 토글의 세션 간 지속성)은 라이브 관측이 필요해 관측하지 않았다. 편집 텍스트가 이를 주장하지 않는다는 것만 확인했다(R13, W11, 그리고 슬라이더 단어를 피한 표현은 RS-M5처럼 핀을 통과하므로 diff를 눈으로 읽어 확인: 슬라이더 지속성 주장 문장 없음).
5. 렌더된 페이지를 브라우저로 열어 보지 않았다. hugo 출력의 문자열 존재만 확인했다.
6. `OwnershipTransitionRule`의 양성 대조는 만들지 않았다. `moai spec lint --json --strict`가 `[]`를 냈지만, 이 빈 배열이 "규칙이 작동해 문제없음"인지 "규칙이 침묵"인지는 구분하지 못한다.
7. `bin/moai`의 `-dirty` 표시 원인은 확인하지 못했다. lane 기록(progress.md §E.2)은 그 시점에 커밋되지 않은 파일이 docs-site 편집뿐이었다고 적는다.
8. 거부된 명령 두 건. (a) 세션 첫 Bash 호출 `cd <worktree> && git rev-parse ... && ...`: 워크트리 가드가 git 복합형이라 거부했다. 단순 명령으로 쪼개 다시 실행했다. (b) 업스트림 4개를 받는 `cd ... && for ... curl ... raw.githubusercontent.com/...` 반복문: 같은 가드가 거부했다(URL의 `githubusercontent` 때문으로 추정, 미확인). 절대 경로를 쓴 개별 curl 4개로 다시 실행했다. 두 경우 모두 거부를 읽기나 추정으로 대체하지 않고 쪼갠 명령의 실측 출력을 사용했다.
9. 이 감사의 `grep`은 BSD grep이다. SPEC 장부가 전제한 ugrep과 구현이 다르며, 두 구현의 값이 일치한 것은 lane 기록과 내 재측정의 대조로만 알 수 있다(내가 ugrep을 따로 돌리지는 않았다).
10. plan-audit 보고서(iter1..iter3)는 다시 감사하지 않았다. 결정 기록이 인용한 사실만 대조했다(§6).

### 3.5 잔여 위험 (Residual-risk)

- 핀은 어휘에 묶여 있어 §3.2(E)의 변이 중 RS-M1..M5, RS-M7, WF-M1..M3, WF-M6이 통과한다. 전달된 텍스트에는 해당 문장이 없다는 것을 diff 전체 읽기로 확인했지만, 이후 수정이 같은 핀을 통과하면서 틀릴 수 있다.
- 업스트림 문서가 바뀌면 설정 키 경로의 버전 한정(F1)과 `--effort ultracode`의 `xhigh` 예외가 낡는다.
- 스크래치 로그(`pins.out`, hugo 로그, 변이 스크립트)는 머신 로컬이며 인용 대상으로 쓰지 않았다. 결정 줄은 이 파일에 옮겼다. 스크래치에 남긴 전체 출력은 이 워크트리가 폐기되거나 `/tmp`가 정리되면 사라진다(손실 인지).
- 이 파일은 워크트리에만 있다. 리더가 읽기 전에는 워크트리를 폐기하면 안 된다.

## 4. 사전 부채 처분 표 (D1-D4와 선언된 잔여)

| 항목 | 처분 | 근거 |
|---|---|---|
| D1 행마다 "토글"과 "레벨 불변"을 말하는가 | 해소 | §3.2(D): 4개 로케일 모두 토글 단어와 레벨 불변 문구가 표 행 한 줄 안에 있다(각 1). diff 4개 행을 직접 읽어 확인. 이 부분은 SPEC 핀이 없어 WF-M1 WF-M2 변이가 핀을 통과한다는 점이 부채의 구조적 원인이다 |
| D2 행이 두 열 형태를 유지하는가 | 해소 | D2 핀 1×4, 헤더 `\| 항목 \| 설명 \|` 등 두 열 헤더가 diff 맥락에 보인다. 세 번째 열을 넣는 변이(WF-M5)는 D2 핀이 잡는다 |
| D3 W1이 UTF-8 로케일에 의존하는가, 절 안에 구두점이 없는가 | 해소(사실 확인) | `LC_ALL=en_US.UTF-8`에서 1×4, `LC_ALL=C`에서 ko ja zh가 0(§3.2(C)). 매칭된 절 4개는 `| . 。 ; ； , ， 、`를 포함하지 않는다. 핀은 로케일 의존이라 "어느 로케일로 돌렸는가"를 항상 함께 적어야 한다 |
| D4 spec.md L50의 SHA와 HISTORY 순서 | 비차단 부채(F5) | `git diff --stat c50da9c2f ff7b64f6e`와 `... 0e7b6af5b` 모두 `.claude internal docs-site`에서 출력이 없다. 두 SHA가 가리키는 트리에서 범위 파일이 동일하므로 어떤 측정도 영향받지 않는다. 고칠 소유자는 manager-spec(비전환 본문 정정)이며 plan-audit 반복을 다시 열 가치는 없다 |
| 선언된 잔여: `xhigh` 토큰 없이 결합을 말하는 문장 | 해소 | 18개 편집 파일의 diff 전체를 읽었다. 추가·변경 줄 중 ultracode를 어떤 effort 레벨에 묶는 문장은 4개 로케일 어디에도 없다. 추가로 `grep -rn -E 'ultracode.*xhigh|xhigh.*ultracode'`를 docs-site, `.claude`, 템플릿, README 4종에 돌려 co-occurrence가 규칙 원본·미러, workflows 행 4개, 그리고 SPEC이 범위 밖으로 둔 `model-policy.md`(ultrathink 설명과 `/effort` 구문 나열)와 `commands.md` 구문표뿐임을 확인했다 |

## 5. 결함 목록 (구조화 defect-list)

모든 항목은 optional이다. 블로킹으로 분류할 결함은 없었다.

- F1 [minor] [optional] `.claude/rules/moai/workflow/dynamic-workflows.md:111`(와 미러 111), `docs-site/content/zh/claude-code/agentic/workflows.md:105`, 그리고 en/ja/ko 행(`en:105`, `ja:105`, `ko:113`), ultracode-workflows 페이지(`en:115`, `ja:112`, `zh:112`, `ko:70`) — 설정 키 경로의 버전 표현이 업스트림 문장보다 강하다. 규칙 원본은 "requires Claude Code `v2.1.284` or later", zh 행은 "需要 v2.1.284 或更高版本"이라고 쓰고, en/ja/ko는 "(v2.1.284 or later)" 식으로 키 사용 가능 시점처럼 읽히게 쓴다. 업스트림(settings-reference `ultracode` 항목, model-config)은 키가 v2.1.284 이전에도 있었고 그때는 세션을 `xhigh`로 돌렸다고 적는다. v2.1.284가 요구하는 것은 키 자체가 아니라 effort 레벨 불변과 `/effort ultracode off`다. 규칙 원본은 뒤 절 "earlier versions treat that key differently"로 일부 바로잡지만 docs 행에는 그 절이 없다. 근거: model-config가 v2.1.284 필요 목록으로 "The /effort ultracode off form, the slider toggle, and keeping ultracode on at effort levels other than xhigh"만 꼽는다(§3.2(F)). SPEC이 이 표현을 못박았기 때문에(REQ-003, DEC-1, REQ-011, 핀 R11 W3) 구현은 SPEC에 충실하며 optional로 분류했다. 필요한 수정: "since v2.1.284 the key leaves the effort level unchanged; earlier versions ran the session at `xhigh`" 형태로 SPEC 본문(DEC-1과 R11/W3 핀)을 먼저 고친 뒤 10곳(규칙·미러·행 4·UW 4)에 반영. 소유자는 manager-spec이므로 후속 카드 후보이다. 영향 범위: 구버전 사용자가 키를 쓸 수 없다고 오해할 수 있는 정도.
- F2 [minor] [optional] `docs-site/content/ko/claude-code/agentic/workflows.md:113` — "켜기와 끄기를 따로 정하는 토글입니다". 원문 "independent on/off toggle"의 독립성은 effort 레벨에서의 독립인데, 한국어 문장은 켜기와 끄기가 따로라는 뜻으로 읽힌다. 고칠 방향: "effort 레벨과 별개로 켜고 끄는 토글입니다". `ja:105`의 "オンとオフを独立して切り替えるトグル"도 같은 구조로 보이나 비모국어 독자의 판단이다. `ja:105` "毎回のセッションをオンで始めるには"의 "毎回"가 자연스러운지는 확신하지 못한다("すべてのセッションをオンの状態で開始するには"이 무난해 보임, 미확정). zh 문장 3곳은 어색한 점을 찾지 못했으나 비모국어 판단이다.
- F3 [minor] [optional] `docs-site/content/ko/advanced/ultracode-workflows.md:83` — 콜아웃은 "새 세션에서 직접 다시 켜야 합니다"라고만 말하는데, 같은 페이지 `:70`은 이제 설정 키로 매 세션을 켠 채 시작할 수 있다고 말한다. 규칙 원본은 "(or rely on the settings key)"로 이 균형을 맞췄다. REQ-007이 콜아웃 보존을 요구하므로 문장 보존이 우선이고, 설정 키를 한 구절로 덧붙이는 정도가 가능하다. 영향: 한 페이지 안 약한 모순.
- F4 [minor] [optional] 수용 핀의 한계(전달된 텍스트의 결함은 아님). §3.2(E)에서 RS-M1 RS-M2 RS-M3 RS-M4 RS-M5 RS-M7, WF-M1 WF-M2 WF-M3 WF-M6이 모든 릴리스 차단 핀을 통과했다. SPEC이 선언한 잔여("`xhigh` 없이 말한 결합")보다 넓은 것은 RS-M2(플래그 절 자리에서 `xhigh` 결합을 일반화), RS-M3(도입 버전 오기), RS-M7(주장 뒤집기), WF-M6(명령형 지속성 주장 뒤집기)이다. 필요한 수정은 이 카드에는 없다. 이후 비슷한 SPEC에서는 도입 버전을 `since Claude Code .v2.1.284` 절에 묶는 핀과 `does not turn it off` 같은 부정문 핀을 추가한다.
- F5 [minor] [optional] `.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/spec.md:50`(`0e7b6af5b`)와 `acceptance.md:5`, `:115`(`ff7b64f6e`) 불일치, 그리고 `spec.md:21-23`의 HISTORY가 v0.1.0, v0.1.2, v0.1.1 순서(D4 그대로). 측정에는 영향이 없다는 것을 §4에서 확인했다. 수정은 manager-spec 소관이고 반드시 필요하지 않다.
- F6 [minor] [optional] `.moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/progress.md:7` — §E.1 `plan_phase_notes`가 "Final audit (iteration 3) pending."으로 끝난다. 계획 시점에는 맞았으나 iteration 3이 PASS-WITH-DEBT 0.89로 끝났으므로 지금 읽으면 낡았다. 이력 기록이라 고치지 않아도 되며, 고친다면 sync 소유자(manager-docs)가 "pending" 뒤에 iteration 3 결과를 덧붙이는 정도이다.
- F7 [minor] [optional] 상태 전환 시점과 소유 표기. Ownership Matrix는 `draft -> in-progress`를 "run 첫 커밋(M1)"에 두는데 실제 전환은 네 번째 run 커밋 `90d5bac45`(M4 증거 기록)에 실렸다. 첫 run 커밋 `abf8ba627`과 `be5d56b8c` 시점에는 `status: draft`였다(`git log -G'^status:'` 출력으로 확인). lane 지시에 따른 것이고 소유자 표기(`manager-develop`)는 맞으며, `in-progress -> completed`는 매트릭스대로 단일 sync 커밋 `211905e37`(`manager-docs`)에 실렸다. run 단계는 worktree 격리 우회책으로 `general-purpose` 스폰이 manager-develop 역할을 맡았으므로 `Authored-By-Agent`는 역할 표기이며 런타임이 검증한 에이전트 유형이 아니다(자기 증언). 비차단으로 판단했다: 문서화된 우회책이고 린트가 위반을 내지 않는다.
- F8 [info] [optional] 이 카드가 CHANGELOG를 바꾸지 않은 것은 결함이 아니다. SPEC 변경 지도(§3)는 CHANGELOG를 포함하지 않고 §4는 이력 항목을 범위 밖으로 둔다. AC-010은 `CHANGELOG.md`의 `git diff --stat`가 비어야 한다고까지 요구하므로 `[Unreleased]` 한 줄을 넣었다면 오히려 AC-010을 어겼을 것이다. 사용자에게 보이는 규칙·문서 문구 수정이라 항목을 쓸지는 리더 결정이며, 쓰려면 별도 카드나 SPEC 개정이 필요하다.
- F9 [info] [optional] 자율 Kickoff 처리에서 정리해 둘 교리 한 점. `plan-auditor.md`의 Retry Loop Contract(L691-695)는 iter3 이후 상한에서 "사용자에게 세 선택지로 에스컬레이션"한다고 적는데, 결정 기록은 같은 자리에서 오케스트레이터가 선택지 1(PASS-with-debt)을 스스로 골랐다. 이를 §9.1 자율 전환의 결정 사다리가 흡수한다고 읽는 것이 자연스럽지만(자율 형태는 결정 기록을 보상 통제로 둔다), 두 규칙 문장이 같은 자리에서 다른 행동을 지시하는지는 이 감사가 판정하지 않았다. 리더가 교리 정리 여부를 결정한다.

## 6. 과정 증거 판정 (자율 Kickoff 보상 통제)

`.moai/reports/t1416/decision-records.md`를 `auto-semantics.md` §9.1과 §10에 대조했다. 모두 이 실행에서 다시 확인한 값이다.

| §9.1 조건 | 결정 기록의 주장 | 재확인 |
|---|---|---|
| 독립 plan-audit 판정이 FAIL/INCONCLUSIVE가 아님 | iter3 PASS-WITH-DEBT 0.89, Tier M 임계 0.80 | `plan-audit-iter3.md`에서 `auditor-model: claude-sonnet-5-5[1m]`, `verdict: PASS-WITH-DEBT`, `audited_sha: 3d44379b86dc8d66ee3d9d38f40e29d5a095d186`, "Overall Score: 0.89" 줄 확인 |
| plan 단계가 audit-ready를 기록 | progress.md §E.1 `plan_status: audit-ready` | progress.md 5행에서 확인 |
| plan-artifact 해시가 판정 이후 불변 | `git diff --stat 3d44379b8 HEAD -- .moai/specs`가 비었음 | 결정 시점 커밋 `0e7087faa`에 대해 `git diff --stat 3d44379b8 0e7087faa -- .moai/specs`가 출력 없음(재현). 현재 HEAD까지의 diff는 progress.md와 spec.md뿐이고 spec.md 변경은 `status: completed` 한 줄이다. `git diff 3d44379b8 HEAD -- spec.md plan.md acceptance.md`는 spec.md frontmatter 한 줄 외 본문 변경이 없고 plan.md와 acceptance.md는 변경 없음 |
| 열린 블로커 없음 | D1(주요 급 핀 공백), D2(경미) 등 | iter3 보고서의 부채 목록과 일치 |
| keep-set 없음 | 환경 불가·운영자 보유·외부 공유계 비가역 작업 없음 | 이후 단계(통합, push)는 리더 게이트이므로 맞음 |
| §10 한 줄 형식 | `decision record: decided_by=... evidence_refs=... ladder_path=...` | 3행이 세 필드를 이 순서로 한 줄에 담음 |

**PASS-WITH-DEBT를 §9.1의 "PASS"로 보는 것은 방어 가능하다고 판정한다.** 이유: (1) §9.1의 하드 블록은 FAIL과 INCONCLUSIVE이고, `plan-auditor.md`가 `verdict: <PASS|PASS-WITH-DEBT|FAIL>` 세 토큰을 정의하며 iter3 상한 규칙이 PASS-with-debt를 정식 출구로 둔다. (2) 점수 0.89가 Tier M 임계 0.80을 넘고 필수 기준 9개가 모두 통과했다. (3) 부채 5개 중 핵심 둘(D1 D2)이 실제로 run과 sync에서 손으로 확인되었고 내가 다시 해소를 재현했다. (4) 건너뜀 적격성을 계산하는 `internal/runtime`에는 PASS-WITH-DEBT 처리가 없고(`grep -rln -i 'pass-with-debt' internal/runtime` 출력 없음), `internal/spec/closer.go`와 `internal/cli/spec_close.go`의 언급은 SPEC close가 AC 판정 셀을 읽는 전제 조건이며 이 SPEC의 progress.md에는 그런 셀 표기가 없다(`grep -n -E 'PASS-WITH-DEBT|\*\*PASS'` 출력 없음). 그래서 기계적 충돌은 관측되지 않았다. 결정 기록은 자기 증언이지만 그 증언을 이 감사가 모두 재현했으므로 보상 통제가 작동했다. 남는 점은 F9이다.

소유권 판정: 위 F7과 §3.2(I) 참조. `in-progress -> completed`는 매트릭스와 일치하고 린트가 위반을 내지 않는다. 오케스트레이터가 직접 커밋한 것은 감사 판정 세 건(`plan-auditor` 줄 포함)과 결정 기록 한 건(줄 없음)뿐이라는 브리프의 설명과 로그가 일치한다.

## 7. 도구 출처와 린트 인용 조건

이 보고서의 `spec lint`와 `spec audit` 결과는 §3.3에 적은 `./bin/moai`(커밋 `abf8ba627`, HEAD와 Go 표면 동일)가 만든 값이다. 설치본이 HEAD보다 뒤처졌으므로 설치본의 린트는 근거로 쓰지 않았다. 다만 바이너리의 `-dirty` 원인을 확정하지 못했으므로(§3.4 7) 린트 통과는 "HEAD와 같은 Go 표면에서 규칙이 아무 finding도 내지 않았다"는 뜻으로만 읽는다.

## 8. 실행한 명령 목록과 반복 이력

이 감사는 이 카드의 첫 sync 감사다(재감사 이력 없음). 앞선 plan-audit 3회(0.75 FAIL, 0.86 FAIL, 0.89 PASS-WITH-DEBT)는 `.moai/reports/t1416/plan-audit*.md`에 있다.

실행한 명령 군: `git rev-parse`와 `git status --short`, `git diff --stat`/`--name-only`, `git log`, `git show c50da9c2f:<파일>`; `bash pins.sh`(위 핀 전체, 스크래치); `python3 mutants.py`(변이 16개); `bash red_and_scan.sh`(기점 RED, 추가 줄 스캔); `curl`(업스트림 4개) 후 `python3 strip.py`; `cd docs-site && hugo --minify --gc --destination <스크래치>`; `./bin/moai spec lint`/`spec audit`/`version`; `grep -a -o`(바이너리 embed); `grep -rn`(잔여 결합 스캔). 쓰기는 이 파일 하나뿐이고 워크트리의 소스·SPEC 파일은 수정하지 않았다.
