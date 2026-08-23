# t170 sync-phase 보고서 — SPEC-FEEDBACK-AUTO-SUBMIT-001

> 5-section 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
> 범위: **브랜치 위 sync 만.** push · PR · 릴리스 워크트리 진입 · 머지는 리드 보류 대상이며 이 레인은 sync 커밋에서 멈춘다.

## 1. Claim (주장)

1. `CHANGELOG.md` `[Unreleased]` 에 이 SPEC 의 항목이 섹션의 기존 어투(영어 설명 산문)로 들어갔고, 강제력을 과장하지 않았다 — "convention the skill body follows, not a sandbox".
2. docs-site 4로케일(`ko`/`en`/`ja`/`zh`)의 `utility-commands/moai-feedback.md` 는 run-phase M8(`a6682a007`)이 이미 착지시켰고, sync 는 **다시 쓰지 않고 패리티만 관측**했다.
3. README 는 변경이 **필요 없다** — 근거는 실행한 grep 이지 가정이 아니다.
4. `progress.md §E.4` 를 감사관이 읽는 입력으로 채웠다(sync 가 바꾼 것 · 커밋 SHA · 미검증 · 외부 블로커 t189).
5. 3-phase close 를 단일 sync 커밋에 실었다. **전이는 `spec.md` 한 파일에서만** 일어난다 — 나머지 3종은 frontmatter 자체가 없다.
6. 문서·CHANGELOG 전용 변경이 `internal/config` · `internal/template` 를 움직이지 않았다.

## 2. Evidence (증거 — 실행한 명령과 그 출력)

### 2.1 SPEC 감사 (close 전)

```
$ moai spec audit --json   # cwd = 워크트리 루트
exit=0
total_specs: 640, grandfathered: 278, modern_era_clean: 361
MUST-FIX: 1건 → SPEC-CODEX-SKILLS-CANONICAL-001 (SyncStatusDrift, 타 SPEC·선재)
SPEC-FEEDBACK-AUTO-SUBMIT-001 의 drift finding: 0건
```

### 2.2 테스트

```
$ go test ./internal/config/... ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/config	9.752s
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	0.681s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	(cached)
ok  	github.com/modu-ai/moai-adk/internal/template	52.516s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	(cached)
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
```

### 2.3 4로케일 패리티

```
$ grep -c auto_submit docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md
ko:2  ja:2  en:2  zh:2

$ grep -c '^### ' docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md
en:14  ko:14  zh:14  ja:14
```

새 절은 4로케일 모두 존재(`제출 전 확인` / `Confirming Before Submission` / 일·중 대응절). Mermaid 다이어그램은 이번 sync 에서 건드리지 않았고, M8 이 추가한 것은 표준 markdown 문단 + YAML 코드블록뿐이다(shortcode · Mermaid · 이모지 없음).

### 2.4 README 무변경 근거

```
$ grep -n feedback README.md README.ko.md README.ja.md README.zh.md
→ 4파일 각각 3곳: 서브커맨드 나열 2 + 이슈 링크 1. 설정 키를 다루는 문단 없음.

$ grep -n "config/sections\|auto_submit" README.md
574:### `.moai/config/sections/`
593:| `state.yaml` | …
→ 설정 섹션 표는 "편집하게 되는" 6개 + v3.1.1 추가 4개만 싣고 `feedback.yaml` 을 포함하지 않는다.
```

### 2.5 B12 CHANGELOG 방출 규율 3종

```
(a) 중복 방지: $ git show HEAD:CHANGELOG.md | grep -c 'SPEC-FEEDBACK-AUTO-SUBMIT-001' → 0
(b) AC 개수: $ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 24
             CHANGELOG 주장 24 와 일치. 0 이 아니므로 공허한 비교가 아니다.
(c) 경로 확인: $ ls …/spec.md …/plan.md …/design.md → 3개 모두 존재
```

### 2.6 close 후 상태

커밋 직후 `git status --porcelain` 무출력, `moai spec audit --json` 재실행 결과는 아래 §3 에 귀속과 함께 적는다.

## 3. Baseline-attribution (baseline 귀속)

- 트리: 워크트리 `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`.
- sync 착수 시 HEAD: `cdff7f315` (run-phase M9 기록 커밋). base: `3210da7d3`.
- 위 §2 의 모든 측정은 **이 트리에서 이번 회차에 직접 실행**한 것이다. run-phase §E.2/§E.3 의 수치를 옮겨 적은 것이 아니다.
- `moai spec audit` 은 **CLI** 로 실행했다. `mcp__moai__spec_audit` 은 `project_root` 를 넘겼는데도 primary 체크아웃을 감사한 것으로 관측됐다(§4-1).

## 4. Gaps (미검증 — 명시)

1. **`mcp__moai__spec_audit` 의 `project_root` 미작동.** `project_root=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170` 을 넘겨 호출했으나 결과가 `total_specs: 627` 이었고 `SPEC-FEEDBACK-AUTO-SUBMIT-001` 이 출력에 **0회** 등장했다. 이 SPEC 디렉터리는 워크트리에만 있고 primary(`/Users/goos/MoAI/moai-adk-go/.moai/specs/`, 628개)에는 없다 — 즉 primary 를 감사했다는 뜻이다. 워크트리 cwd 의 CLI 는 `total_specs: 640` 으로 이 SPEC 을 포함한다. 원인 규명은 이 카드 범위 밖이며 **별도 카드감**이다. (이 관측 자체는 결함 주장이 아니라 "MCP 경로로는 이 트리를 감사하지 못했다"는 사실 기록이다.)
2. **Hugo 빌드 미실행.** docs-site 렌더를 확인하지 않았다. 이번 sync 는 docs-site 파일을 편집하지 않았으므로 새 위험을 만들지는 않았다.
3. **전체 스위트 미실행.** `go test ./...` 를 로컬에서 돌리지 않았다(CLAUDE.local.md §4). 전 패키지 판정은 통합 후 PR CI 몫.
4. **브라우저 확인 없음.** 웹 콘솔 feedback 섹션은 파일 내용으로만 관측했다(§E.3 미검증 3 그대로).
5. **외부 블로커 t189 는 손대지 않았다.** `internal/cli/agentlint` `TestConstitutionCrossReference` 는 base `3210da7d3` 에서도 붉고 귀속 커밋은 `243eb07ef`(t82 M4). lane-8 소관이며 이 sync 에서 재실행하지도, 고치지도 않았다.
6. **push · PR · 머지 없음.** 리드 보류 대상.
7. **`sync_commit_sha` 는 플레이스홀더.** 커밋이 자기 해시를 알 수 없으므로 `pending-backfill-SPEC-FEEDBACK-AUTO-SUBMIT-001` 로 두고 후속 커밋에서 백필해야 한다.

## 5. Residual-risk (잔여 위험)

- **백필 부채.** 플레이스홀더가 백필되지 않으면 감사 이력에 "sync 는 끝났는데 해시가 없다"가 남는다. 다만 `status` 를 같은 커밋에서 `completed` 까지 올렸으므로 `SyncStatusDrift` MUST-FIX 형태(형제 `SPEC-CODEX-SKILLS-CANONICAL-001` 사례)로는 뜨지 않는다.
- **§E.3 의 누적 잔여 위험 28건은 sync 로 해소되지 않는다.** 그중 문서 표면에 직접 걸리는 둘: **#26** docs 문구 ↔ 스킬 본문 사이에 기계적 패리티 검사가 없어 스킬 본문이 바뀌면 문서가 조용히 어긋난다. **#28** docs 에 스크러버 서사를 새로 열지 않았으므로 마스킹 규칙 전체를 문서에서 찾는 사용자는 이 페이지에서 찾지 못한다.
- **강제력 문구의 내구성.** CHANGELOG 는 "convention, not sandbox" 로 정직하게 적었으나, 이후 누군가 릴리스 노트나 마케팅 문구를 요약하면서 "마스킹을 강제한다"로 줄일 여지가 있다. `plan.md` AP-12 · `design.md` §1 · `progress.md` §E.3 · 이 보고서가 같은 문구를 네 곳에서 반복하는 이유다.
- **형제 SPEC 병합.** `SPEC-TODO-ENABLE-FLAG-001` 과 파일 9종을 공유한다(§E.3 잔여 22·23). 통합 순서에 따라 개수 고정 테스트 4건을 형제 쪽에서 다시 조정해야 할 수 있다.
