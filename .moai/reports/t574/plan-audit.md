# SPEC Review Report: SPEC-TODO-HOME-TEMP-GUARD-001 — 제자리 개정 0.1.5 (카드 t574)

Iteration: 1/2 (Tier M 상한)
Verdict: PASS
Overall Score: 0.94 (Tier M PASS 기준 0.80)

감사 대상: 개정 커밋 `ac751bc48`(부모 `95ba9deb2`)이 바꾼 SPEC 산출물 diff에 한정한다.
작성자 측 서술(커밋 메시지, repro-summary 주장, 개정 본문의 자기 서술)은 판정 근거로 쓰지 않았다(M1 Context Isolation). 모든 판정은 이 감사에서 직접 실행하고 관측한 출력을 근거로 한다.

---

## Claim

개정 0.1.5는 운영자가 정한 형태(새 AC 없이 AC-THG-006에 양의 방향 절 추가)를 스키마의 개정 규칙에 맞게 수행했다. 추가된 절은 판정 테스트 `TestDefaultTempRoots_Membership`의 단언과 일치하고 이진 판정이 가능하며, 뮤턴트 M-574의 RED 대조는 실제 증거 파일로 뒷받침된다. must-pass 7항목은 모두 PASS 또는 N/A다. 발견된 결함 7건은 모두 minor·optional이며 판정을 뒤집지 않는다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: 개정 diff는 REQ 본문을 건드리지 않는다(`git diff 95ba9deb2 ac751bc48` 추가 행은 frontmatter, HISTORY, `## Amendments`, AC-THG-006 절, 매트릭스 행, progress 블록뿐). spec.md:147-155에 REQ-THG-001..009가 공백·중복 없이 순서대로 있다.
- [PASS] MP-2 GEARS 형식(요구사항 층 기준): REQ 층은 변경 없음(spec.md:147-155, 각 항목에 GEARS 라벨). 새로 추가된 Given/When/Then(acceptance.md:152-156)은 **검증 층의 AC 절**이므로 이 항목의 채점 대상이 아니다.
- [PASS] MP-3 frontmatter: spec.md:2-14에 12개 필드가 모두 있고 타입도 맞다. `version: "0.1.5"`(따옴표 semver, 0.1.4에서 patch 증가), `status: in-progress`(enum), `updated: 2026-09-10`, `amendment_of: SPEC-TODO-HOME-TEMP-GUARD-001`(제자리 개정의 자기 참조). 거부 alias는 없다. 트리 빌드 lint 결과는 무결함(아래 Evidence).
- [N/A] MP-4 언어 중립성: `internal/kanban` 단일 모듈을 다루는 Go 전용 SPEC이고, 템플릿에 묶인 내용이 아니다.
- [PASS] MP-5 D7: 개정이 새로 추가한 SPEC 참조는 자기 ID 하나뿐이며, 폐기·대체·보관 상태 참조가 없다. BLOCKING 없음.
- [PASS] MP-6 D8: `grep -c syscall` 결과 spec/plan/acceptance/progress 모두 0이다.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → plan.md 일치 0건, research.md는 없다(Tier M). 개정 블록도 `open_questions: none`이다(progress.md:35).

### 브리프 must-pass 6항목

1. **절의 판정 가능성과 테스트 일치 — PASS.** internal/kanban/temp_origin_test.go:162-199를 직접 읽었다.
   - `/tmp`·`/var/folders` 리터럴 포함(:169-173)은 AC Then(acceptance.md:154)과 대응한다.
   - `os.TempDir()` 존재만 판정(:179-182)은 Then 단서와 대응한다.
   - `len(got) != 3`(:187-189)은 And 「정확히 3개」(:155)와 대응하고, 중복 허용은 :164의 「중복을 허용한 개수」와 맞다.
   - `slices.Contains(got, "/var/tmp") && os.TempDir() != "/var/tmp"`(:196)는 And의 `/var/tmp` 제외 + `TMPDIR=/var/tmp` 예외(:156)와 대응한다.
   - 스텁도 `TempOriginReason`도 거치지 않는다는 Given(:152)은 테스트 주석 :154-157 및 `defaultTempRoots()`를 직접 호출하는 :163과 맞다.
   - 판정 명령(:158)이 그 테스트를 이름으로 지목한다.
2. **RED 대조가 실재 — PASS(주석 D1).** `.moai/reports/t574/mutant-kanban.txt`에는 `--- FAIL: TestDefaultTempRoots_Membership`와 `fixed temp root "/tmp" missing`·`"/var/folders" missing`·`has 1 members, want 3`가 기록돼 있다. 판정 명령이 지목한 kanban 테스트 8건(SymlinkSpellingEquivalence, FailsOpenOnUnresolvable, ComponentBoundary, TempOriginRefusesHomeQueue, NonTempNonGitKeepsHomeFallback, PureGuardIsSilent, GitBranchUnreachedByGuard, PureFallbackWritesNothing)은 모두 `--- PASS`다. 인용된 문구가 파일 내용과 글자 그대로 일치한다. 측정 트리 `7da398808`은 `ac751bc48`의 조상이고(`git merge-base --is-ancestor` exit 0), 두 커밋 사이 `internal cmd pkg` diff는 비어 있다.
3. **Frontmatter/HISTORY/Amendments — PASS.** 스키마가 요구하는 네 항목이 모두 있다(spec.md:35-61): 직전 completed 판 `0.1.4`, prior_completed_sha `029ab039f`, 근거, 범위와 범위 밖. `git show 029ab039f -- spec.md`에서 `-status: in-progress` / `+status: completed`를 직접 관측했다. 이 커밋의 제목은 `docs(SPEC-TODO-HOME-TEMP-GUARD-001): sync-phase artifacts + 3-phase close (t536)`다. 백필 커밋 `586f26f2a`는 progress.md를 1줄 바꿨다(관측). 개정 커밋 제목은 스키마 행의 `feat(SPEC-{ID}): in-place amendment <rationale-summary>` 패턴과 일치하고, `## Amendments` 배치는 선례 SPEC-DRIFT-CLOSE-BODY-001 등 9건과 같은 H2 형식이다.
4. **progress.md 배치와 파서 토큰 — PASS.** 새 블록은 `## §E.1 Plan-phase Audit-Ready Signal`(progress.md:5) 아래의 `### 개정 0.1.5 — plan-phase 신호`(:13-37)에 있다. Section Map이 manager-spec의 plan 신호에 배정한 자리다. 블록 안에 `§E`, `sync_commit_sha`, `mx_commit_sha` 토큰이 없다. era.go의 `extractProgressField`는 키를 정확히 일치(`m[1] == field`, era.go:247·253)로 비교하므로, `amendment_plan_status`·`prior_completed_sha`는 `plan_status`·`sync_commit_sha`와 충돌하지 않는다.
5. **AC 8건과 §D.0 매핑 — PASS.** `^### AC-` 헤딩은 정확히 8개다(acceptance.md:46, 80, 94, 110, 122, 141, 168, 200). acceptance.md:36은 `- AC-THG-006 maps REQ-THG-002` 그대로이고, 매트릭스 :22의 REQ 칸도 `REQ-THG-002`여서 본문(:141-166)과 어긋나지 않는다.
6. **코드·테스트·규칙·템플릿 무변경 — PASS.** `git diff --stat 95ba9deb2 ac751bc48`의 전체 결과는 `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/` 아래 3개 파일(acceptance.md, progress.md, spec.md)뿐이다.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | 절 본문(acceptance.md:152-156)은 해석이 하나뿐이다. 다만 RED 대조의 「AC가 이름 부른 나머지 kanban 테스트는 전부 PASS」(:162, D1)와 리눅스 셀의 「os.TempDir()은 /tmp」(:164, D6)에 범위 한정어가 빠져 있다 |
| Completeness | 1.0 | 1.0 | 스키마 개정 요소가 전부 있다: frontmatter spec.md:4-8, HISTORY :26, Amendments :35-61(범위 밖 포함), progress 신호 :13-37 |
| Testability | 1.0 | 1.0 | 새 절은 이진 판정이 가능하고 판정 테스트와 1:1로 대응한다(temp_origin_test.go:169-198). 모호한 표현이 없다 |
| Traceability | 1.0 | 1.0 | AC-THG-006은 REQ-THG-002에 매핑되어 있고(:36), REQ-THG-002가 집합 `os.TempDir()`, `/tmp`, `/var/folders`를 명시한다(spec.md:148). `/var/tmp` 제외 근거 spec.md §8(:185)도 실재한다 |

## Evidence

```
$ git rev-parse --show-toplevel ; git branch --show-current ; git rev-parse --short HEAD
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t574
WT-temp-roots-ac
ac751bc48

$ git diff --stat 95ba9deb2 ac751bc48
 .../SPEC-TODO-HOME-TEMP-GUARD-001/acceptance.md    | 22 ++++++++++---
 .../SPEC-TODO-HOME-TEMP-GUARD-001/progress.md      | 26 ++++++++++++++++
 .moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/spec.md  | 36 ++++++++++++++++++++--
 3 files changed, 77 insertions(+), 7 deletions(-)

$ go build -o /tmp/moai-t574-audit ./cmd/moai        → build_exit=0
$ /tmp/moai-t574-audit spec lint SPEC-TODO-HOME-TEMP-GUARD-001   (exit 0)
✓ No findings — all SPEC documents are valid
$ /tmp/moai-t574-audit spec lint --json SPEC-TODO-HOME-TEMP-GUARD-001
[]

$ go version
go version go1.26.8 darwin/arm64
$ go test -count=1 -v -run '^(TestTempOrigin_ComponentBoundary|TestDefaultTempRoots_Membership)$' ./internal/kanban/
=== RUN   TestTempOrigin_ComponentBoundary
=== RUN   TestTempOrigin_ComponentBoundary/production_roots:_/tmpfoo_is_not_inside_/tmp
--- PASS: TestTempOrigin_ComponentBoundary (0.00s)
    --- PASS: TestTempOrigin_ComponentBoundary/production_roots:_/tmpfoo_is_not_inside_/tmp (0.00s)
=== RUN   TestDefaultTempRoots_Membership
    temp_origin_test.go:182: os.TempDir() on this platform: "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/"; full set: ["/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/" "/tmp" "/var/folders"]
--- PASS: TestDefaultTempRoots_Membership (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.441s
(swept: RUN 2건 + 서브테스트 1건, 공허 스윕 아님)

$ git show --format=%H%n%s 029ab039f -- .moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/spec.md
029ab039f59958ee2a35b76a725f205b5b169033
docs(SPEC-TODO-HOME-TEMP-GUARD-001): sync-phase artifacts + 3-phase close (t536)
-status: in-progress
+status: completed

$ git merge-base --is-ancestor {029ab039f|7da398808|1d091087b} ac751bc48   → 셋 다 exit 0
$ git diff --stat 7da398808 ac751bc48 -- internal cmd pkg                   → 출력 없음
$ git log -S'func TestDefaultTempRoots_Membership' --format='%h %s' -- internal/kanban/temp_origin_test.go
1d091087b test(t536): repair sync-audit F1 — guard the production temp-root set (SPEC-TODO-HOME-TEMP-GUARD-001)

$ grep -c syscall spec.md acceptance.md progress.md plan.md   → 모두 0

$ /tmp/moai-t574-audit spec audit --json --filter-spec SPEC-TODO-HOME-TEMP-GUARD-001
{ ... "finding_type": "SyncStatusDrift", "severity": "MUST-FIX",
  "remediation": "moai spec close SPEC-TODO-HOME-TEMP-GUARD-001 --backfill-only",
  "details": { "reason": "§E.2 + §E.4 + sync_commit_sha present (sync complete) but status != completed",
               "spec_md_status": "in-progress",
               "sync_commit_sha": "029ab039f              # D3 백필 창 상환분 — ..." } }
```

(MCP `spec_audit`(서버 빌드 v3.2.0-rc.5 `84fa4ece4`, `project_root`=이 워크트리)도 같은 결과를 냈다. 판정 근거는 트리 빌드 쪽 출력이다.)

## Baseline-attribution

- 측정 트리: `.claude/worktrees/t574` @ `ac751bc48`(브랜치 `WT-temp-roots-ac`). 이번 실행에서 측정했다.
- lint·audit 판정 빌드: 이 트리에서 빌드한 `/tmp/moai-t574-audit`를 경로로 직접 호출했다(§2.2 도구 출처). 버전 배너는 `moai-adk v3.1.3`으로 찍히는데, 이는 ldflags 없이 빌드했을 때의 기본 문자열이며 소스는 이 트리다.
- 뮤턴트 RED 증거는 작성자가 `7da398808`에서 측정한 커밋 파일(`.moai/reports/t574/mutant-kanban.txt`)이다. 이 감사는 뮤턴트를 **재주입하지 않았다**. 파일 내용과 `7da398808`↔`ac751bc48` 코드 동일성은 직접 확인했다.
- 툴체인: go1.26.8 darwin/arm64.

## Defects Found (structured defect-list)

D1. RED-CLAIM-SCOPE — acceptance.md:162 — 「AC가 이름 부른 나머지 kanban 테스트는 **전부** PASS」가 인용 증거보다 넓다. AC-THG-007 본문(acceptance.md:183)이 이름을 부른 kanban 테스트 `TestAdoptionLandsWhereConsumersRead`(todo_root_contract_test.go:14)와 `TestResolveTodoQueueRootAdopting_AdoptsLocalQueue`(todo_root_test.go:249)는 mutant-kanban.txt에 RUN 줄이 없다. 즉 뮤턴트 아래에서 관측되지 않았다. 증거가 덮는 범위는 **판정 명령이 지목한** kanban 테스트 8건이다. — Severity: minor — Class: optional — Required fix: 문구를 「판정 명령이 이름 부른 나머지 kanban 테스트(8건)는 전부 PASS」로 좁히거나, 두 테스트를 포함해 뮤턴트를 재측정하고 결과를 증거 파일에 덧붙인다.

D2. RED-CELL-COMMAND — acceptance.md:162, .moai/reports/t574/repro-summary.md:24-29 — 뮤턴트 대조의 **실행 명령**이 어디에도 기록되지 않았다(`grep -rn 'go test' .moai/reports/t574`는 판정 명령 제안문 외에 일치 없음). exit 코드는 요약 산문에만 있고 증거 파일의 필드로는 없다. 뮤턴트 소스도 보존되지 않았다(t536은 `mutants/mutant-AUD2-rootset.go.txt`로 보존했다). verification-completeness §2.1이 요구하는 네 요소 중 「명령」이 빠진 셈이다. 출력만으로 재구성할 수는 있어 판정에는 영향이 없다. — Severity: minor — Class: optional — Required fix: run 재측정 때 뮤턴트 diff, 단일 실행 명령, exit 코드를 증거 파일에 필드로 남긴다.

D3. MATRIX-GRADE-DRIFT — acceptance.md:22, 등급 정의 acceptance.md:7 — 매트릭스 행의 RED 칸은 여전히 `/tmpfoo` 한 가지만 적고 M-574를 반영하지 않았다. 등급도 `blocking (RED 관측 전)` 그대로다. 그런데 새 소속 절은 개정 시점 트리에서 이미 GREEN(`repro-tests-develop.txt` PASS, 테스트는 `1d091087b`부터 존재)이라 RED-now 셀을 가질 수 없고, RED는 뮤턴트로만 나온다. 문서 자신의 등급 용어로 보면 invariant-guard 모양이다. 합쳐진 행이 이 성격 차이를 드러내지 않는다. — Severity: minor — Class: optional — Required fix: 운영자 결정(기존 AC 확장)은 그대로 두고, 매트릭스 RED 칸에 「소속 절: M-574 뮤턴트 대조(관측됨), 기준 트리에서는 이미 GREEN」을 한 구절 덧붙인다.

D4. AMENDMENT-WINDOW-DRIFT — progress.md:36(`next:`) — 개정 창 동안 트리 빌드와 MCP 서버 빌드가 모두 `SyncStatusDrift` **MUST-FIX**를 낸다. 제시되는 조치 `moai spec close ... --backfill-only`를 따르면 status가 completed로 돌아가 개정이 run 재측정·sync 재close 없이 조용히 끝난다. 스키마는 개정 중 드리프트 탐지가 다시 켜진다고 명시하므로 개정 자체의 위반은 아니지만, 산출물 어디에도 이 예상 경보가 적혀 있지 않다. — Severity: minor — Class: optional — Required fix: progress 개정 블록의 `next:`에 「개정 창 동안 SyncStatusDrift는 예상된 경보다. `--backfill-only`를 실행하지 않는다」를 덧붙인다. 도구 쪽 개선(`amendment_of` 인식)은 별도 카드 후보다.

D5. LINUX-CELL-CONDITION — acceptance.md:164 — 「리눅스에서 `os.TempDir()`은 `/tmp`」와 「`/tmp`가 두 번 들어간다」는 `TMPDIR`이 비어 있을 때만 참이다. 셀이 미측정이라고 스스로 밝혔으므로 결함은 작다. — Severity: minor — Class: optional — Required fix: 「`TMPDIR`이 비어 있으면」 한정어를 붙인다.

D6. UNANCHORED-SELECTOR — acceptance.md:158 — 판정 명령 `-run 'TestTempOrigin_ComponentBoundary|TestDefaultTempRoots_Membership'`에 앵커가 없다. 현재는 두 함수에만 일치한다(grep 확인). 그러나 한쪽 이름이 바뀌면 나머지만 돌고도 초록으로 끝난다(verification-completeness §1.1 공허 스윕). — Severity: minor — Class: optional — Required fix: `'^(TestTempOrigin_ComponentBoundary|TestDefaultTempRoots_Membership)$'`로 앵커를 걸고, 기대 RUN 수(2)를 함께 적는다.

D7. NO-WHO-TRAILER — `ac751bc48` 커밋 본문 — `completed → in-progress (amendment)` 전이 커밋에 `Authored-By-Agent:` 트레일러가 없어 소유자(manager-spec)를 기계적으로 귀속할 수 없다. 이미 착지한 커밋이므로 되돌려 쓸 수 없다. — Severity: minor — Class: optional — Required fix: 이후 run·sync 재close 커밋에는 트레일러를 붙인다.

(관측, 이번 개정의 범위 밖: progress.md §E.4의 `sync_commit_sha` 줄 끝 주석이 era.go 파서에 값 일부로 읽힌다(audit details 참조). `586f26f2a`에서 들어온 선재 상태이며 이번 개정과 무관하다.)

## Gaps

- 뮤턴트 M-574를 이 감사에서 재주입하지 않았다. RED 대조는 커밋된 증거 파일과 코드 동일성 확인으로만 판정했다.
- 리눅스·윈도우 셀에서 `TestDefaultTempRoots_Membership`를 실행하지 않았다(CI 몫).
- AC-THG-007 본문의 두 kanban 테스트는 뮤턴트 아래에서 관측된 적이 없다(D1).
- 교차 모델 감사(codex/glm)는 호출하지 않았다. 프로젝트의 `audit_model` 설정도 읽지 않았다.
- 개정 범위 밖(REQ 본문, plan.md, 다른 AC)은 이 감사가 판정하지 않았다.

## Residual-risk

- 개정 창에서 `SyncStatusDrift` MUST-FIX가 계속 나오는 동안, 누군가 권고 조치를 기계적으로 실행하면 개정이 재측정 없이 닫힌다(D4).
- 소속 절은 테스트 이름에 기대고 있다. 앵커 없는 선택자와 함께 쓰이면 이름 변경 시 공허 초록이 날 수 있다(D6).
- 뮤턴트 소스와 명령이 보존되지 않아, 뒤에 누군가 대조를 재현할 때 모양이 달라질 수 있다(D2).

## Recommendation

PASS. must-pass 근거는 이렇다.
- MP-1·MP-2: REQ 층 무변경(spec.md:147-155).
- MP-3: frontmatter 12필드와 개정 필드 정상(spec.md:2-14). 트리 빌드 lint 무결함.
- MP-5·MP-6·MP-7: 해당 결함 없음.
- 브리프 6항목: 모두 직접 관측으로 PASS.

결함 D1-D7은 모두 optional이다. 가장 값싼 순서는 D1·D5·D6(문구 몇 곳) → D3(매트릭스 한 구절) → D4(`next:` 한 줄)이고, 모두 manager-spec이 run 재측정 전에 반영할 수 있다. D2·D7은 run·sync 단계 소유자가 다음 커밋에서 챙기면 된다. 판정은 이 감사의 것이며, Implementation Kickoff Approval 게이트를 대신하지 않는다.
