# t623 — plan-auditor D7·D8 예제 스크립트 판정 기록

card: t623 · class B · Tier S~M · lane-9
branch: WT-auditor-d7d8-scripts · base: 로컬 develop `1e207c3ff`
출처: `.moai/reports/instruction-audit-20260910-01a089f6/report.md` AC-02 · AC-03 (보고서 기준 main `2213871af`)

## 1. develop 재현 (전제 재측정)

### Claim
보고서의 두 결함은 develop `1e207c3ff` 에서도 그대로 재현된다. AC-02 는 거짓 양성(명시적 조정이 있는 정상 참조를 BLOCKING), AC-03 은 거짓 음성(다른 섹션의 `//go:build` 가 신규 섹션의 누락을 가림)이다. 로컬 사본과 템플릿 사본의 두 스크립트는 바이트 동일하다.

### Evidence
위치(보고서 줄번호는 main 기준이라 문구로 찾음): `git grep -n -E "D7|D8" 1e207c3ff -- .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md` → 두 사본 모두 `### Group 7: Cross-SPEC Reconciliation (D7)` 336행, `### Group 8: Cross-Platform Discipline (D8)` 367행.

두 사본 차이: `git diff 1e207c3ff:.claude/agents/moai/plan-auditor.md 1e207c3ff:internal/template/templates/.claude/agents/moai/plan-auditor.md` → D7-1 의 예시 SPEC-ID 한 줄(템플릿은 중립화된 예시 ID)과 `## Retry Loop Contract` 본문만 다르다. D7·D8 스크립트 블록은 차이 없음.

스크립트 추출: 각 사본의 `### Group 7:`·`### Group 8:` 뒤 첫 ` ```bash ` 블록을 awk 로 그대로 뽑고, 자리표시자 `<new-spec.md>` 만 `new-spec.md` 로 바꿨다(`repro/d7-local.sh`, `d7-tmpl.sh`, `d8-local.sh`, `d8-tmpl.sh`). `cmp d7-local.sh d7-tmpl.sh` → `d7_cmp_exit=0`, `cmp d8-local.sh d8-tmpl.sh` → `d8_cmp_exit=0`.

픽스처(`repro/`, 각 픽스처 디렉터리를 작업 디렉터리로 두고 `sh <script>` 실행, 출력은 `out-<fixture>.log`):

| 픽스처 | 내용 | 현재 스크립트 출력 | 판정 |
|---|---|---|---|
| `d7-fp` | superseded SPEC 을 `## Reconciliation` 문단에서 "supersedes ... absorbed" 로 명시 조정 | `BLOCKING: SPEC-FIXTURE-OLD-001 has status=superseded but is referenced without reconciliation`, `exit=0` | **AC-02 거짓 양성 재현** |
| `d7-tp` | 같은 superseded SPEC 을 조정 문구 없이 참조 | `BLOCKING: ...`, `exit=0` | 옳은 BLOCKING (반대 방향 대조군) |
| `d7-live` | status=implemented SPEC 참조 | BLOCKING 없음, `exit=0` | 스크립트가 침묵할 수 있음을 보이는 대조군 |
| `d8-fn` | 1절에 `//go:build !windows`(syscall 없음), 2절에 `syscall.Kqueue`(태그 없음) | 출력 없음, `exit=0` | **AC-03 거짓 음성 재현** |
| `d8-tn` | 같은 절에 `syscall.Kqueue` + `//go:build darwin` | 출력 없음, `exit=0` | 옳은 통과 (반대 방향 대조군) |
| `d8-tp` | 태그 없이 `syscall.Kqueue` 만 | `BLOCKING: SPEC references syscall but no //go:build constraint or EXCL justification`, `exit=0` | 스크립트가 BLOCKING 을 낼 수 있음을 보이는 대조군 |

`d8-fn` 의 무출력이 "스크립트가 아무것도 못 냄" 이 아니라는 것은 같은 스크립트가 `d8-tp` 에서 BLOCKING 을 낸 것으로 선다. `d7-fp` 의 BLOCKING 이 "무엇이든 BLOCKING" 이 아니라는 것은 `d7-live` 의 침묵으로 선다.

### Baseline-attribution
워크트리 `WT-auditor-d7d8-scripts` HEAD `1e207c3ff`, 이 실행.

### Gaps
- 부수 관측: 세 D7 픽스처 모두 새 SPEC 자신의 ID(`SPEC-FIXTURE-NEW-001`)에 대해 `SHOULD: referenced SPEC ... not found` 를 냈다. 픽스처의 새 SPEC 이 `.moai/specs/` 아래에 없기 때문에 생긴 것이며, 실제 감사에서는 새 SPEC 이 자기 경로에 있어 해소되는지는 관측하지 않았다. 이 카드 범위 밖.
- 워크트리 생성 직후 첫 `git merge --ff-only 1e207c3ff` 가 `exit=1` 로 실패했다(출력을 버렸기 때문에 원인 미관측). 곧이은 재시도는 `ff_exit=0` 으로 성공했고 `git merge-base --is-ancestor d060e0d13 1e207c3ff` 는 `exit=0` 이었다.

### Residual-risk
픽스처는 결함의 최소 형태다. 실제 SPEC 본문의 다른 형태(코드 펜스 안의 `syscall`, 제목 없는 문서 등)는 여기서 재지 않았다.

## 2. 수리 초안 — 픽스처 사전 실행 (문서 반영 전)

### Claim
초안 스크립트(`repro/d7-fixed.sh`, `repro/d8-fixed.sh`)는 두 결함을 닫고, 반대 방향 대조군을 망가뜨리지 않는다.

### Evidence
초안의 요지:
- D7: 스크립트가 BLOCKING 을 직접 내지 않는다. retired·superseded·archived 참조마다 `REVIEW:` 한 줄을 내고, 그 SID 가 조정 키워드(revers·supersed·absorb·carve-out)와 같은 문단에 있으면 그 문단을 `reconciliation candidate` 로 함께 출력한다. BLOCKING 결정은 감사자가 그 문단을 읽고 내린다.
- D8: 제목(`^#+ `) 단위 섹션마다 `syscall` 과 `//go:build`·`cross-platform exemption`·`EXCL.*syscall` 을 따로 본다. 같은 섹션에 태그가 없으면 섹션 제목을 붙여 BLOCKING. 입력을 읽을 수 없으면 `GAP:`.

실행 방식은 §1 과 같다(픽스처 디렉터리에서 `sh <script>`, 출력은 `repro/fixed-<fixture>.log`).

| 픽스처 | 초안 출력 | 판정 |
|---|---|---|
| `d7-fp` | `REVIEW: SPEC-FIXTURE-OLD-001 has status=superseded — confirm explicit reconciliation ...` + `reconciliation candidate (paragraph 4): This SPEC supersedes SPEC-FIXTURE-OLD-001: its eviction requirement is absorbed here, and the old SPEC stays retired.`, `exit=0` | BLOCKING 없음 — **AC-02 거짓 양성 닫힘** |
| `d7-tp` | `REVIEW: ...` 만, 후보 문단 없음, `exit=0` | 침묵하지 않음 — 반대 방향(조정 없는 참조) 유지 |
| `d7-live` | REVIEW 없음, `exit=0` | 대조군 유지 |
| `d8-fn` | `BLOCKING: section "## 2. New file watcher" references syscall but carries no //go:build constraint or EXCL justification`, `exit=0` | **AC-03 거짓 음성 닫힘** |
| `d8-tn` | 출력 없음, `exit=0` | 반대 방향(같은 섹션 태그) 통과 유지 |
| `d8-tp` | `BLOCKING: section "## 1. New file watcher" ...`, `exit=0` | 대조군 유지 |
| `d8-gap` (`new-spec.md` 없음) | `GAP: new-spec.md is not readable — D8 was not observed`, `exit=0` | 미관측이 통과로 읽히지 않음 |

세 D7 픽스처의 `SHOULD: ... SPEC-FIXTURE-NEW-001 not found` 는 §1 과 같은 픽스처 부산물이다.

### Baseline-attribution
워크트리 `WT-auditor-d7d8-scripts` HEAD `1e207c3ff`, 문서 미수정 상태, 이 실행.

### Gaps
- 초안은 아직 문서에 반영하지 않았다. 반영 후 두 사본에서 다시 추출해 같은 픽스처로 재실행해야 한다.
- 리드 판정 대기 두 건: (1) D7 스크립트가 BLOCKING 을 내지 않게 되면서 MP-5 산문을 "REVIEW 줄을 읽은 감사자가 BLOCKING 을 낸다" 로 맞추는 방향, (2) C3 사본 `internal/template/templates/.codex/agents/moai/plan-auditor.toml` 에도 같은 스크립트가 들어 있어(D8 제목 363행, grep 381행) `make agents-emit` 재생성이 필요하다는 점.
- D7 키워드 목록이 놓치는 조정 표현(예: "replaces")은 후보 문단으로 뜨지 않는다. 그 경우에도 REVIEW 줄은 남으므로 감사자가 읽어서 판정한다 — 스크립트가 자동 BLOCKING 을 내지 않는 이유다.

### Residual-risk
- D8 섹션 경계는 마크다운 제목만 본다. 코드 펜스 안의 `#` 로 시작하는 줄은 제목으로 오인될 수 있다.
- awk 의 `RS = ""` 문단 모드와 `tolower` 는 POSIX awk 기능이며 macOS 기본 awk 에서만 실행했다.

## 3. 문서 반영 — 템플릿 먼저, 로컬 손편집, C3 재생성

### Claim
리드 판정(결정 1: 스크립트는 REVIEW/candidate/GAP 만, BLOCKING 은 auditor 가 읽고 판단 / 결정 2: 워크트리 안 `make agents-emit` 1회 허용)대로 두 사본을 고쳤고, 두 사본에서 다시 추출한 스크립트가 §2 에서 검증한 초안과 바이트 동일하며 7개 픽스처에서 같은 결과를 낸다. C3 는 재생성분만 바뀌었고 방출 검사가 통과한다.

### Evidence
수정 위치(두 사본 동일 4곳, 템플릿 먼저 → 로컬 손편집, cp 없음):
1. MP-5: "Group 7 스크립트는 `REVIEW:` 후보만 내고 BLOCKING 은 auditor 가 문맥을 읽고 낸다" 한 문장 추가.
2. D7-4: "— otherwise BLOCKING" 뒤에 "decided by the auditor after reading the script's `REVIEW:` output, never by the script".
3. D7 스크립트 교체 + 양방향 문단: "`reconciliation candidate` 없는 `REVIEW:` 는 자동 BLOCKING 이 아니다(읽어서 조정이 없을 때만 BLOCKING, 키워드 밖 표현도 조정이다) / candidate 문단은 자동 통과가 아니다(읽어서 실제로 조정하지 않으면 BLOCKING)".
4. D8 스크립트 교체(제목 단위 섹션 판정 + `GAP:`) + 섹션 범위를 두는 이유와 GAP 을 통과로 읽지 말라는 문단.

두 사본 차이(수정 후): `diff .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md > copies-diff-after.txt` → `diff_exit=1`, 헝크 `338c338`, `463c463`, `465,466d464` — 수정 전과 같은 기존 차이 2곳(D7-1 예시 ID, Retry Loop)뿐. 이번 수정은 두 사본에 같은 바이트로 들어갔다.

재추출 비교(§1 과 같은 awk 추출 + 자리표시자 치환):
```
d7_copies_cmp=0
d8_copies_cmp=0
d7_vs_draft_cmp=0
d8_vs_draft_cmp=0
```
재추출본 줄 수: `post-d7-local.sh` 18, `post-d8-local.sh` 16.

C3 재생성: `make agents-emit > agents-emit.log 2>&1` → `emit_exit=0`, 로그 끝 `ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.417s`. `git status --short -- internal/template/templates/.codex` → ` M internal/template/templates/.codex/agents/moai/plan-auditor.toml` 한 파일. `git diff --stat` → `29 insertions(+), 9 deletions(-)`, 헝크 위치 142(MP-5)·339(D7-4)·346~367(D7)·379~405(D8) — 수정한 영역뿐. C3 손편집 없음.
방출 검사(읽기 전용): `make agents-emit-check > agents-emit-check.log 2>&1` → `check_exit=0`, 로그 끝 `ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.419s`.
대조군: `git grep -c -F 'Surface cross-SPEC reconciliation candidates'` → 로컬 .md 1, C3 .toml 1 — 새 문구가 C3 에 실제로 실렸다.

재추출 스크립트로 픽스처 재실행(템플릿 사본 7건, 로컬 사본 7건, 각각 단독 명령; 출력 `repro/post-tmpl-*.log`, `repro/post-local-*.log`):

| 픽스처 | 템플릿·로컬 사본 출력(동일) |
|---|---|
| `d7-fp` | `REVIEW: ... superseded — confirm explicit reconciliation ...` + `reconciliation candidate (paragraph 4): This SPEC supersedes SPEC-FIXTURE-OLD-001: ...`, BLOCKING 없음, `exit=0` |
| `d7-tp` | `REVIEW: ...` 만, `exit=0` |
| `d7-live` | REVIEW 없음, `exit=0` |
| `d8-fn` | `BLOCKING: section "## 2. New file watcher" ...`, `exit=0` |
| `d8-tn` | 출력 없음, `exit=0` |
| `d8-tp` | `BLOCKING: section "## 1. New file watcher" ...`, `exit=0` |
| `d8-gap` | `GAP: new-spec.md is not readable — D8 was not observed`, `exit=0` |

### Baseline-attribution
워크트리 `WT-auditor-d7d8-scripts` HEAD `955a66792` + 미커밋 수정(두 .md, C3 .toml), 이 실행. `make build`·embed-check 는 돌리지 않았다(배치 끝 리드 몫).

### Gaps
- **기존 테스트가 옛 계약을 고정한다.** `internal/cli/plan_audit_d7_d8_test.go` 는 D7·D8 스크립트를 문서에서 읽지 않고 테스트 코드 안에 옛 스크립트 사본을 박아 두었다(87행·142행 D7, 184~188·226~230행 D8). 그래서 문서가 바뀌어도 이 테스트는 계속 초록이고, 동시에 `TestPlanAuditD7_RetiredSPECConflict` 는 "조정 없는 retired 참조 → 스크립트가 BLOCKING" 이라는 **폐기된 계약**을 맞다고 주장한다. 주석의 문서 경로(`.claude/agents/meta/plan-auditor.md`)도 현존하지 않는 경로다. 이 테스트는 `internal/cli` 소속이라 고치거나 돌리려면 리드 슬롯이 필요하다 — 리드 판정 대기.
- D7 후보 문단 판독(auditor 가 읽고 BLOCKING 을 내는 층)은 사람·모델 판단이라 기계적으로 검증하지 않았다.

### Residual-risk
- 테스트를 고치지 않으면, 누군가 문서 스크립트를 옛 형태로 되돌려도 어떤 테스트도 붉어지지 않는다.

## 4. 테스트 재작성 — 리드 판정 (A), 슬롯 1회차

### Claim
`internal/cli/plan_audit_d7_d8_test.go` 를 문서에서 스크립트를 추출해 실행하도록 다시 썼다. 옛 문서(`1e207c3ff`)에 대해 붉어지는 칸은 측정 전에 적어 둔 4칸과 정확히 같다. 현재 문서에 대해서는 2칸이 붉었는데, 원인은 문서가 아니라 테스트 단언의 결함이었고 수정했다(수정본은 아직 미컴파일).

### Evidence
재작성 요지: 테스트 9개(기존 이름 4 + 신규 5). 스크립트는 템플릿 사본의 `### Group 7:`/`### Group 8:` 아래 첫 ` ```bash ` 블록에서 매 실행마다 읽는다(제목이나 블록이 없으면 실패). 두 사본 스크립트 동일성도 테스트로 고정. 주석의 존재하지 않는 경로(`.claude/agents/meta/...`)를 정정.

측정 전 예측(기록): 옛 문서에서 FAIL = {RetiredSPECConflict, ReconciledReferenceIsNotBlocking, BuildTagInOtherSectionDoesNotCover, UnreadableSpecIsGap}, PASS = 나머지 5개.

슬롯 1회차 절차:
- 09:00:58Z 실행 파일 이름 기준 비교기 `ps -eo pid,etime,comm,args | awk '$3 ~ /(^|\/)(go|vet|cli\.test)$/ && /internal\/cli|cli\.test/'`. 양성 대조: 합성 행 3개 중 `go test ./internal/cli/` 행과 `cli.test` 행 2개 매치, `awk internal/cli` 행 제외. 실측 출력 없음.
- 09:01:04Z–09:01:19Z `go test -c -o /tmp/t623-lane9/cli.test ./internal/cli` → `compile_exit=0` (컴파일 1회).
- `TestMain` 이 작업 디렉터리에 의존하지 않음을 확인(`main_test.go:213` — 프로필 디렉터리 임시화, 명령 트리 예열, 작업 디렉터리 `.moai` 잔여물 검사뿐).
- 같은 바이너리로 두 번 실행: (1) 워크트리 `internal/cli` 에서, (2) `/tmp/t623-lane9/oldtree/internal/cli` 에서 — 그 트리에는 `git show 1e207c3ff:<경로>` 로 꺼낸 두 사본만 둔다(`tmpl_show_exit=0`, `local_show_exit=0`, 43888 / 44840 바이트).

옛 문서 실행(`test-old-doc-1e207c3ff.log`, `exit=1`): FAIL 4 = UnreadableSpecIsGap · BuildTagInOtherSectionDoesNotCover · ReconciledReferenceIsNotBlocking · RetiredSPECConflict, PASS 5 = VerbsIdenticalAcrossCopies · MissingBuildTag · WithBuildTagPasses · MissingReferencedSPEC · LiveReferenceIsSilent. 예측과 일치. 결함이 실제 출력으로 드러났다: GAP 칸 `got: grep: new-spec.md: No such file or directory`, 다른 섹션 칸 `got: `(빈 출력).

현재 문서 실행(`test-current-doc.log`, `exit=1`): PASS 7, FAIL 2 = ReconciledReferenceIsNotBlocking(169행) · RetiredSPECConflict(143행). 로그 전문(37~38·41행)을 보면 문서 출력은 기대대로다 — `REVIEW: ... before emitting BLOCKING` 과 `reconciliation candidate (paragraph 4): ...`. REVIEW 줄 문구 자체에 `BLOCKING` 이 들어 있어 `strings.Contains(out, "BLOCKING")` 이 잘못 적중했다. 같은 테스트의 "REVIEW 줄 있음"·"후보 문단 있음" 단언은 통과했다.

수정: 줄 시작 표지로 판정하는 `hasFinding(out, "BLOCKING:")` / `hasFinding(out, "REVIEW:")` 로 교체(일괄 치환). 확인: 옛 부분문자열 검사 잔존 `git grep` → 매치 없음(`leftover_grep_exit=1`), 대조로 `hasFinding(out, ` 8줄. `gofmt -l` → 출력 없음(`gofmt_exit=0`).

### Baseline-attribution
워크트리 `WT-auditor-d7d8-scripts` HEAD `57240649b` + 미커밋 테스트 재작성, 이 실행. 옛 문서 트리는 `1e207c3ff` 블롭.

### Gaps
- 수정한 단언으로는 아직 컴파일·실행하지 않았다 — 리드에게 슬롯 2회차 요청 중.
- 옛 문서 칸의 FAIL 이 수정 후에도 같은 4칸일 것이라는 판단은 옛 출력이 `BLOCKING:` 로 줄을 시작한다는 로그 판독에 근거한 추론이며, 재실행으로 확인해야 한다.

### Residual-risk
- CI 의 Windows 러너에 `bash` 가 없으면 이 테스트는 실패한다. 재작성 전 테스트도 `bash -c` 를 썼으므로 새로 생긴 조건은 아니다.

## 5. 슬롯 2회차 — 수정한 단언으로 재측정 (리드 추가 승인, 마지막 슬롯)

### Claim
줄머리 표지 판정으로 고친 테스트는 현재 문서에서 9개 전부 통과하고, 옛 문서(`1e207c3ff`)에서는 측정 전 기록한 4칸만 실패한다. REVIEW 줄 속의 `BLOCKING` 낱말은 더 이상 BLOCKING 발견으로 세지지 않는다 — 이번 실행 로그의 해당 줄로 확인된다.

### Evidence
슬롯 2 전 추가 변경: D7 두 테스트(RetiredSPECConflict·ReconciledReferenceIsNotBlocking)에 스크립트 출력을 남기는 `t.Logf("D7 output:\n%s", out)` 한 줄씩(새 테스트 아님 — 통과한 테스트의 출력을 로그에 싣기 위함). `gofmt -l` → 출력 없음(`gofmt_exit=0`).

절차:
- 09:03:57Z 실행 파일 이름 비교기(1회차와 같은 awk). 양성 대조 합성 행 2개 매치·`awk` 행 제외, 실측 출력 없음.
- 09:04:03Z–09:04:13Z `go test -c -o /tmp/t623-lane9/cli2.test ./internal/cli` → `compile_exit=0` (컴파일 1회).
- 같은 바이너리로 현재 문서 트리 / 옛 문서 트리 두 번 실행.

현재 문서(`test2-current-doc.log`, `exit=0`, 47행): 최상위 `--- PASS` 9개 = VerbsIdenticalAcrossCopies · UnreadableSpecIsGap · WithBuildTagPasses · BuildTagInOtherSectionDoesNotCover · MissingBuildTag · MissingReferencedSPEC · LiveReferenceIsSilent · RetiredSPECConflict · ReconciledReferenceIsNotBlocking. `--- FAIL`·`^FAIL` 0건. 마지막 줄 `PASS`.
REVIEW 줄의 `BLOCKING` 낱말이 발견으로 세지지 않는다는 증거(같은 로그):
```
36:         REVIEW: SPEC-FIXTURE-OLD-001 has status=retired — confirm explicit reconciliation in the same section or paragraph before emitting BLOCKING
38:         REVIEW: SPEC-FIXTURE-OLD-001 has status=superseded — confirm explicit reconciliation in the same section or paragraph before emitting BLOCKING
39:           reconciliation candidate (paragraph 4): This SPEC supersedes SPEC-FIXTURE-OLD-001: its eviction requirement is absorbed here.
45: --- PASS: TestPlanAuditD7_RetiredSPECConflict (0.04s)
46: --- PASS: TestPlanAuditD7_ReconciledReferenceIsNotBlocking (0.04s)
```
두 테스트는 출력에 `BLOCKING` 낱말이 있는데도 "BLOCKING 발견 없음" 단언을 통과했다.

옛 문서(`test2-old-doc-1e207c3ff.log`, `exit=1`, 57행): FAIL 4 = ReconciledReferenceIsNotBlocking · RetiredSPECConflict · BuildTagInOtherSectionDoesNotCover · UnreadableSpecIsGap — 정렬 비교로 측정 전 예측 집합과 일치. PASS 5 = VerbsIdenticalAcrossCopies · MissingReferencedSPEC · MissingBuildTag · WithBuildTagPasses · LiveReferenceIsSilent. 옛 스크립트는 실제로 `BLOCKING:` 로 시작하는 줄을 냈고(로그 40·42·44행 superseded, 51·53·55행 retired) 새 판정이 그것을 잡았다.

### Baseline-attribution
워크트리 `WT-auditor-d7d8-scripts` HEAD `57240649b` + 미커밋 테스트 재작성, 이 실행. 옛 문서 트리는 `1e207c3ff` 블롭.

### Gaps
- `internal/cli` 패키지 전체, `golangci-lint`/`go vet`, `GOOS=windows` 교차 빌드는 돌리지 않았다(리드 규칙: 레인은 영향 범위만).
- `make build`·embed-check 는 배치 끝 리드 몫이라 돌리지 않았다.

### Residual-risk
- 테스트는 D7 후보 문단을 "보여 주는지"까지만 고정한다. 그 문단을 읽고 BLOCKING 을 내는 auditor 판단은 기계로 검증하지 않는다.
