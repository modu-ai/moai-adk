# SPEC Review Report: SPEC-CODEMAPS-REFRESH-002 — iteration 2

Iteration: 2/2 (Tier M ceiling)
SPEC version: **0.1.4**
Verdict: **PASS**
Overall Score: **0.88** (Tier M PASS threshold 0.80; iter-1 = 0.74)

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t475` · 브랜치 `WT-codemaps-stale` · HEAD `52f863f3666c9ec754253a06b96ed1fe844f1590`

> **파일 위치 선택**: iter-1 판정서(`plan-audit.md`)를 덮지 않고 형제 파일로 분리했다. `spec-workflow.md` § Report Persistence 의 plan-phase review stream 규약(`plan-audit-iter<N>.md`, 반복마다 한 파일)을 따른다 — 반복 간 판정 대비가 곧 회귀 검사의 근거이므로 이전 판정서는 보존한다.

M1 Context Isolation: 저작자의 추론 맥락은 무시했다. 리드가 전달한 측정 4건도 **주장으로 취급**하고 전부 이 트리에서 독립 재측정했다.

---

## Claim

iter-1 의 blocking 9건(D1~D9)과 Addendum A 의 O1 이 **전부 해소됐다.** 핵심은 D1 이다: 후보 집합이 손 열거에서 **명령이 뱉는 집합**으로 바뀌었고, 나는 SPEC 이 적어 둔 명령을 그대로 실행해 **A층 5 · B층 14 · C층 1 = 20** 을 재현했다 — SPEC 이 주장한 목록과 **원소까지 동일**하다. must-pass 7항목 전원 PASS, 신규 blocking 0건. Tier M 임계 0.80 을 넘는다.

---

## Regression Check — iter-1 결함 9건 + O1

| # | iter-1 결함 | 상태 | 이 트리에서의 증거 |
|---|---|---|---|
| **D1** | 6단위 집합 재현 불가 · `agentemit` 누락 | **RESOLVED** | §A.3(a)의 3우주 명령을 **그대로 실행**: `Zero=48 / A=5 / B=14`. A층 원소가 SPEC 표와 동일(`core/git`, `settings/yamlpatch`, `stateanchor`, `template/agentemit`, `template/commandemit`) — **`agentemit`이 A층에 들어온다**. B층 14행도 SPEC 목록과 원소 동일. C층 `internal/chain`은 `Zero`에 있고 `A`에 없음을 확인(`grep -cx internal/chain /tmp/A.txt` → `0`), 변경 0건(`grep -c 'internal/chain' <diff>` → `0`)이므로 이름 이월이 규칙과 모순되지 않는다 |
| **D2** | 사전 사본이 재생성 뒤에 위치 | **RESOLVED** | `plan.md:117-123` M2.0 이 재생성(M2.1, `:125`)보다 앞이고, `ls … \| wc -l` → 6 이 **M2.1 실행의 차단 선행 조건**으로 명시("출력이 6이 아니면 M2.1을 실행하지 않는다") |
| **D3** | `docs-truth.md`가 생성기 산출물이 아닌데 AC-003이 통과 | **RESOLVED** | REQ-CM2-003 이 생성기 **5문서**로 좁혀지고 `docs-truth.md`를 명시 제외; REQ-CM2-014 + AC-CM2-003a + M2.2 로 손 갱신 경로 분리; AC-CM2-003a 가 **증거 섹션 §③을 §②와 분리**할 것을 요구("합쳐지면 '재생성했으니 docs-truth도 됐다'는 오독이 그대로 통과한다") |
| **D4** | DoD "3항목" vs AC "2항목" 충돌 | **RESOLVED** | 충돌이 반대 방향으로 해소됐다 — REQ-CM2-013 이 관측을 **셋**(①값 거동 ②후보 20 분류 요약 ③잔여 42 목록)으로 늘렸고, AC-CM2-012(`:229-236`)와 §D.3(`:266`)이 **셋으로 일치**한다. 세 자리 전부 재확인 |
| **D5** | 상위 구간 컷오프 부재 · `internal/settings` 누락 | **RESOLVED** | §A.3(b)의 컷오프 명령(`≥6` AND `internal/template/templates/` 제외)을 그대로 실행 → **7구간**, `internal/settings`(6) 포함. SPEC 이 적은 출력과 행 단위로 동일 |
| **D6** | AC-006/008 추출 명령 부재 | **RESOLVED** | REQ-CM2-006 이 `check_citations.go` 3요소(정규식 `:23` · 구두점 절삭 `:35` · **blockquote 면제**)를 명명하고 `citations` 계층을 판정면으로 연결. REQ-CM2-008 이 추출 명령을 본문에 싣고 **"식별자의 정의는 이 명령의 출력이다"**로 못박음 |
| **D7** | AC-005 "변경 없음" 자유 서술 | **RESOLVED** | AC-CM2-005(`:170-171`)가 구간마다 `diff -u`를 명명하고 **"변경 없음" 행에 빈 diff 출력 첨부**를 요구. M2.4(`plan.md:141-145`)가 같은 형식 |
| **D8** | AC-002가 부모 산문 인용 미요구 | **RESOLVED** | AC-CM2-002 조건 6 이 인용을 요구(`fold`는 `<문서>.md:L<n>` + 인용문, `omission`은 검사 범위 명시), 그리고 내가 iter-1 에 쓴 뮤턴트 문장을 그대로 인용해 무엇을 막는지 명시 |
| **D9** | MUST AC 12개 중 9개 RED-now 부재 | **RESOLVED** | RED-1~RED-7b(8셀) + AC 별 인라인 RED. 13개 AC 전부가 RED-now 를 명명하거나 §B.1 에서 **regression-guard 로 분류**한다 — `verification-completeness.md` §2.1 의 undecidable disposition 을 올바르게 적용한 것이며, 억지 RED 를 지어내지 않은 쪽이 옳다 |
| **O1** | §E 도달성 검증이 `/tmp` 중간 파일 판독 | **RESOLVED** | `plan.md:175` 가 `provenance.json` 의 `commit_sha` 직독으로 교체. `/tmp/t475-stamp-rev` 의 유일한 잔존 언급은 `plan.md:171`, **왜 제거했는지 설명하는 문단**이다("실제로 스탬프에 들어간 값이 아니라 들어갔다고 믿는 값을 검사한다") |

**미해결 이월 0건.** 3회 연속 동일 결함(stagnation) 해당 없음.

---

## Evidence

### Must-Pass Results (iter-2 재실행)

| # | 항목 | 판정 | 증거 |
|---|------|------|------|
| MP-1 | REQ 번호 일관성 | **PASS** | 정의행 14개, id 집합 `REQ-CM2-001`~`014` 연속, 중복 0(`grep -oE '^- \*\*REQ-CM2-[0-9]{3}' \| sort \| uniq -d` → 무출력). 접미사 id(`REQ-CM2-003a`)는 `REQ-CM2-014` 로 개번돼 사라졌다 |
| MP-2 | GEARS (요건 층) | **PASS** | 14개 REQ 전부 패턴 적중. 신규 REQ-CM2-014 는 Ubiquitous(`the executor shall refresh it by hand`) ✓. iter-1 O2 가 지적한 REQ-CM2-002/003/005 의 라벨이 `event-driven` 으로 정정됐다 |
| MP-3 | YAML frontmatter | **PASS** | `moai spec lint <spec.md>` → **`0 error(s)`**. 12 정본 필드 전수, `phase: "v3.2.0 target"` 유효 |
| MP-4 | 언어 중립성 | **N/A** | 저장소 내부 문서 SPEC (auto-pass) |
| MP-5 | D7 교차-SPEC | **PASS** | 참조 SPEC 8건 전부 실재 + 전부 `status: completed`. retired/superseded/archived 0건 |
| MP-6 | D8 크로스플랫폼 | **PASS** | `grep -c syscall` → 4파일 전부 `0`. (REQ-CM2-008 추출식이 codemaps 에서 `syscall.Exec` 를 뽑지만 그것은 codemaps 문서의 내용이지 SPEC 본문이 아니다) |
| MP-7 | clarification 게이트 | **PASS** | `grep -rn '\[NEEDS CLARIFICATION'` → exit 1 |

### 후보 규칙 독립 재유도 — SPEC 의 명령을 그대로 실행

리드가 "either of us 를 믿지 말고 SPEC 의 명령에서 재유도하라"고 요청했다. §A.3(a) 코드블록을 **한 글자도 바꾸지 않고** 실행했다:

```
48 /tmp/Zero.txt
 5 /tmp/A.txt
14 /tmp/B.txt
```

A층 5 (SPEC 표와 원소 동일):
```
internal/core/git   internal/settings/yamlpatch   internal/stateanchor
internal/template/agentemit   internal/template/commandemit
```

B층 14 (SPEC 목록과 원소 동일):
```
internal/cli/{codex_skills_disable,codex_skills_prune,doctor_hook_delivery,
              integration_settings_drift,skills,update_mirror_heal}.go
internal/hook/quality/step_git_env.go
internal/kanban/{prlink_landedref,settings_drift}.go
internal/statusline/state_anchor.go
internal/template/{published_skills,skill_mirror_repair}.go
internal/web/{codexmirror,fieldsets_codex_templ}.go
```

C층 `internal/chain` — `Zero` 에 존재, `A` 에 부재, 앵커 대비 변경 0건. **A+B+C = 20** ✓
잔여 산술: `Zero − A − chain` = **42** ✓ (REQ-CM2-013 ③ 의 수와 일치)

**D1 의 핵심 판정: 이 집합은 이제 재현 가능하다.** iter-1 의 결함은 "6이 잘못된 수"가 아니라 "어떤 규칙으로도 나오지 않는 수"였고, 그것이 사라졌다.

### 명령이 주장하는 수치의 재현

| SPEC 주장 | 재측정 | 결과 |
|---|---|---|
| §A.3(a) `Zero 48 / A 5 / B 14 / C 1 = 20` | 코드블록 verbatim 실행 | 동일 ✓ |
| 잔여 42 | `grep -vxF -f A.txt Zero.txt \| grep -vx internal/chain \| wc -l` | `42` ✓ |
| §A.3(b) 컷오프 → 7구간 | 코드블록 verbatim 실행 | 7행, `internal/settings`(6) 포함, 순서까지 동일 ✓ |
| REQ-CM2-008 추출식이 `52f863f36`에서 **10행** | 코드블록 verbatim 실행 | **정확히 10행** — `AddCommand`, `BacklogPathForRoot`, `ExitCoder`, `PreToolUse`, `RunE`, `Shutdown`, `cli.ResolveExitCode`, `hook.EventType`, `mcp.NewTool`, `syscall.Exec` ✓ |
| `go list … \| wc -l` = 136 | 동일 명령 | `136` ✓ |
| `moai graph check` value=64 / anchor `25a3212a9` | `--json` 재실행 | `"value": 64`, `"content_anchor": "25a3212a9…"`, `"content_anchor_source": "working-tree-differs-from-stamp"` ✓ |
| M4 의 `provenance.json` 추출 명령 | verbatim 실행 | `25a3212a93b4c811cbb22e3c0b34d43571fa65b4` — 정상 동작 ✓ |
| M4 의 `graph check --json \| grep content_anchor` | verbatim 실행 | `2` 히트(`content_anchor`, `content_anchor_source`) — **죽은 명령이 아니다** ✓ |

**M4 의 두 검증 명령이 실제로 값을 낸다는 것을 확인한 이유**: 출력이 비는 명령은 "비실행과 성공이 구별되지 않는" 검사가 된다(`verification-completeness.md` §1.3). 둘 다 살아 있다.

### 이동 ref — 리드의 요청대로 재측정했고, 리드의 값과 다르다

리드는 `origin/develop` 이 "now reads `ef10a2524`" 라고 전했다. **이 트리에서 내가 읽은 값은 `9dddac882` 다** — 리드의 측정 이후 또 움직였거나 다른 트리에서 읽힌 값이다. ref 는 움직이므로 이것은 누구의 오류도 아니지만, 기록은 **내가 관측한 값**을 실어야 한다.

그 tip 에서 재측정:

| 항목 | 결과 |
|---|---|
| `git merge-base HEAD origin/develop` | `52f863f3666c9ec754253a06b96ed1fe844f1590` — **여전히 안정** |
| `git diff --name-only 52f863f36 origin/develop -- internal cmd pkg \| wc -l` | `15` |
| 그중 described-worthy | **정확히 2** — `internal/config/defaults.go`, `internal/config/types.go` |

즉 §A.1 의 "15개 파일 / described-worthy 2개" 주장은 **저자가 측정한 적 없는 네 번째 tip 에서도 성립한다**(`dbee3f6cf` → `19cf21408` → `ef10a2524` → `9dddac882`). 인용된 세 ref 전부 이 트리에서 `git cat-file -t` 로 `commit` 으로 해석되므로 **재실행 가능한 주장**이며, 감사 불가 항목이 아니다.

### 날짜 표기 점검 — "현재값처럼 읽히는 수치"는 남아 있는가

리드의 요청 항목. 전수 확인 결과 **남아 있지 않다**:

- `64` — §A.1 이 `[HARD]` 로 "현재값이 아니라 `52f863f36` 에서 취한 날짜 있는 판독" 이라 못박고, run 재측정이 대체함을 규정.
- `20` — AC-CM2-002 가 "**20은 목표가 아니라 그날의 후보 수다** … 기준은 언제나 '같은가'이지 '20인가'가 아니다".
- `7구간` — REQ-CM2-005/AC-CM2-005 가 "실행 시점 출력이 7과 다르면 실측값을 쓰고 차이를 기록한다".
- `136 / 48 / 5 / 14 / 42` — §B.1 이 regression-guard 로 분류하고 **통과로 기록하지 않는다**.
- `62` — 라이브 참조 0건. 세 히트 전부 HISTORY 행이며, 그중 `0.1.3` 행은 `~~취소선~~` + `[SUPERSEDED by 0.1.4]` 로 표시.

---

## Baseline-attribution

모든 측정은 **이 실행에서, 이 트리에 대해** 수행됐다. `git rev-parse --show-toplevel` = `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t475`, `git rev-parse HEAD` = `52f863f3666c9ec754253a06b96ed1fe844f1590`, `git branch --show-current` = `WT-codemaps-stale`, `git rev-parse origin/develop` = `9dddac882`(감사 시점 판독), `git merge-base HEAD origin/develop` = `52f863f36`. `git status --porcelain` = `?? .moai/reports/t475/` + `?? .moai/specs/SPEC-CODEMAPS-REFRESH-002/` 두 항목뿐. 판정 도구는 이 워크트리에서 빌드된 `./bin/moai`(gitignored).

리드가 사전 검증했다고 전한 4건(status draft / REQ 14·AC 13 / git status / O1 종결)은 **전부 독립 재측정했고 전부 일치**한다 — 전달받은 값을 그대로 채택하지 않았다.

---

## Category Scores

| Dimension | iter-1 | iter-2 | Band | Evidence |
|-----------|--------|--------|------|----------|
| Clarity | 0.75 | **0.90** | 0.75~1.0 | 두 선정 집합이 모두 명령이 됐고 내가 재현했다. 62→20 의 이력이 취소선 + SUPERSEDED 로 투명하게 남아 독자가 어떤 지시가 살아 있는지 오독할 수 없다. 잔여: REQ-CM2-001 의 `(Ubiquitous)` 라벨(본문은 "run 시작 시" — event-driven) |
| Completeness | 0.75 | **0.90** | 0.75~1.0 | `docs-truth.md` 가 REQ-CM2-014 / AC-CM2-003a / M2.2 / 증거 §③ 으로 완전히 분리됐고, 계승 결함임을 명시. 잔여 42 가 버려지지 않고 AC-CM2-012 가 집행하는 이관 조건이 됐다 |
| Testability | 0.65 | **0.85** | 0.75~1.0 | MUST AC 전수가 실행 가능한 명령 또는 RED 셀을 갖는다. 저자가 적은 검증 수치(10행 / 7구간 / 5·14)가 전부 재현됐다. AC-CM2-008 이 **0행을 "통과가 아니라 blocker"** 로 규정한 것은 `verification-completeness.md` §1.1(빈 집합 위의 통과)을 정확히 적용한 것이다. 잔여: AC-CM2-002 조건 6 이 인용의 **존재**를 보되 **해석 가능성**은 보지 않는다(O-1) |
| Traceability | 0.80 | **0.85** | 0.75~1.0 | §D.2 가 13 AC ↔ 14 REQ 를 매핑, REQ-CM2-011 의 비대응 사유 명시. 잔여: 매핑이 `CoverageRule` 에 기계적으로 안 읽힌다(O-2) |

산술 평균 = (0.90 + 0.90 + 0.85 + 0.85) / 4 = **0.875 → 0.88**. Tier M 임계 0.80 초과. **점수 회귀 없음**(0.74 → 0.88), STOP 조건 미해당.

---

## Blocking findings

**없다.** iter-1 의 9건 전부 해소됐고 신규 blocking 0건.

---

## Optional findings (리드 재량 — M6 상, 이것으로 FAIL 을 만들지 않는다)

**O-1. AC-CM2-002 조건 6 은 인용의 존재를 보되 해석 가능성을 보지 않는다 — 저자의 자기 신고 잔여에 대한 판정** — `acceptance.md:135`

저자가 스스로 남긴 잔여를 리드가 그대로 전달했다: *"AC-CM2-002 는 모든 후보가 판정되고 인용됐음을 검증하지 판정이 옳은지는 검증하지 않는다."* **그 분업 자체는 옳다** — 책임 질문은 질적 판단이고, 기계가 대신할 수 없으며, D8 의 인용 요구가 사람 리뷰를 가능하게 만든 것이 정확히 올바른 해법이다.

다만 잔여가 저자가 말한 만큼 환원 불가능하지는 않다. 조건 6 은 `<문서>.md:L<n>` + 인용문의 **존재**를 요구할 뿐, 그 좌표에 그 문장이 실제로 있는지는 보지 않는다. 즉 **좌표를 지어낸 인용**이 조건 6 을 통과한다. 그리고 그 한 겹은 기계화가 가능하다 — 인용문이 그 줄에 실제로 있는지 대조하면 된다:

```bash
# 판정 행이 인용한 <문서>.md:L<n> 과 인용문에 대해
sed -n '<n>p' .moai/reports/t475/pre-regen/<문서>.md | /usr/bin/grep -qF "<인용문>"
```

이것은 **판정의 옳음이 아니라 인용의 해석 가능성**을 보는 것이므로 저자의 분업을 무너뜨리지 않는다. 판정의 옳음은 사람 리뷰에 그대로 남는다.

**blocking 으로 올리지 않는 이유**: 좌표를 지어내는 것은 이미 `verification-claim-integrity.md` §1.1 이 금지하는 형태이고, 조건 6 은 그 위에 기계 검사를 한 겹 얹은 것이다. 두 겹이 최적이지만 한 겹이 결함은 아니다. 값이 싸므로 권한다.

**O-2. §D.2 추적성 표가 `CoverageRule` 에 기계적으로 읽히지 않는다** — `acceptance.md:244-258`

`moai spec lint <spec.md>` 는 REQ 14개 전부를 `CoverageIncomplete` 로 보고한다. 원인을 코드에서 확인했다: `CoverageRule`(`internal/spec/lint.go:906`)은 형제 `acceptance.md` 를 **경로로 직접 읽지만**(`lint_coverage_sibling.go:105`), 커버리지 선언 형식으로 `ExtractRequirementMappings`(`internal/spec/ears.go:128`)의 `maps\s+(REQ-…)` 형태만 인정한다. 이 SPEC 의 §D.2 는 마크다운 표(`| AC-CM2-001 | REQ-CM2-001 |`)라 그 형태가 아니다.

수리는 한 줄 형식 변경이다 — 표 대신 `AC-CM2-001 maps REQ-CM2-001` 형태의 목록을 쓰면 0/14 가 14/14 로 바뀐다. 자기 판정을 기계화하는 것이 이 SPEC 의 전체 논지이므로 자기 추적성만 기계 밖에 두는 것은 일관성 간극이다.

**blocking 이 아닌 이유**: 이 룰은 설계상 advisory 이고(그 룰의 주석이 "reports, never gates" 라고 적는다), 코퍼스 규범이 동일하며(`SPEC-CODEMAPS-REFRESH-001`: REQ 8개 중 **8개 전부** 동일 경고), 어떤 AC 도 lint 청결을 요구하지 않는다.

> **iter-1 기록 정정.** iter-1 판정서의 Traceability 행에서 나는 이 경고를 "린터가 `spec.md` 단독 스캔 시 `acceptance.md` 의 AC 를 못 보는 형태" 라고 적었다. **메커니즘 서술이 틀렸다** — 형제 파일은 인자 유무와 무관하게 항상 읽히며, 걸리는 지점은 매핑 **선언 형식**이다. 결론(코퍼스 전반의 저작 관행, SPEC 결함 아님)은 그대로이고 리드가 취할 조치도 달라지지 않지만, 다음 독자가 잘못된 메커니즘 위에서 수리를 시도하지 않도록 기록을 바로잡는다.

**O-3. A층 필터가 테스트 전용 변경을 다른 변경과 동급으로 들인다** — `spec.md:145`

`internal/core/git` 은 앵커 이후 변경이 `status_branch_test.go` **하나뿐**이라 A층에 들어온다. SPEC 은 `.go` 한정을 쓰지 않는 근거로 `agentemit`(`agents-codex.yaml`)과 `core/git`(테스트)을 나란히 든다.

**필터의 방향 자체는 옳다고 판정한다** — 후보 필터는 과대 포함이 안전한 방향이고, described-worthy 술어가 게이트 지표를 정의할 뿐 후보를 정의하지 않는다는 논거는 성립하며, 무엇보다 그 술어를 쓰면 내가 D1 의 물증으로 든 `agentemit` 이 후보에서 빠진다(그 사실 자체가 실측으로 확인된다). 다만 **방출 계약이 담긴 데이터 파일**의 변경과 **테스트 전용** 변경은 "기술되지 않은 역량" 의 신호로서 강도가 다르며, 두 사례를 동렬로 제시한 문장은 그 차이를 지운다. 후보 하나가 늘어날 뿐이므로 비용은 없다.

**O-4. REQ-CM2-001 의 modality 라벨** — `spec.md:299`

`(Ubiquitous)` 라벨이나 본문은 "run 시작 시" 로 event-driven 이다. 패턴 자체는 GEARS 에 적중하므로 MP-2 는 통과한다. iter-1 O2 에서 REQ-CM2-003 은 정정됐고 001 만 남았다.

---

## Gaps — 명시적으로 관측하지 않은 것

1. **`/moai codemaps --force` 를 실행하지 않았다.** 재생성이 실제로 5문서만 건드리는지, omission 후보를 자동 편입하는지는 미관측이다. D3 해소 판정은 스킬의 **선언된 산출 목록**과 SPEC 의 분리 구조에 근거하며 실행 관측이 아니다.
2. **`moai graph stamp codemaps --commit` 를 실행하지 않았다.** 재스탬프 후 `value=0 verdict=fresh` 가 실제로 나오는지는 미관측이다. 코드 독해(`resolveContentAnchor` Rule A)와 `--json` 의 `content_anchor` 판독까지가 내 근거다.
3. **`git fetch` 를 하지 않았다.** `origin/develop = 9dddac882` 는 이 워크트리의 원격 추적 ref 현재값이며, 실제 원격은 이미 더 나아가 있을 수 있다.
4. **20개 후보 각각의 fold/omission 이 무엇이 될지 판정하지 않았다.** 그것은 M1 의 소관이고 plan-phase 감사의 소관이 아니다. 내가 검증한 것은 **집합이 재현되는가**이지 **판정이 옳을 것인가**가 아니다.
5. **잔여 42개의 성질을 검사하지 않았다.** 운영자 결정으로 이 카드 범위 밖이며, REQ-CM2-013 ③ 의 이관 목록이 그 처분이다.
6. **AC-CM2-002 조건 2 의 두 grep 을 실측 대조하지 못했다** — `candidates.txt` 와 판정 표가 아직 존재하지 않는다(정상: RED-4 가 그 부재다). 정규식이 규정된 행 형식과 정합하는지는 형식 독해로만 확인했다.
7. **codex / GLM 백엔드를 호출하지 않았다.** Claude 단독 앵커다.

---

## Residual-risk

1. **게이트 녹색은 여전히 재생성 품질과 무관하다.** 재스탬프만으로 `value=0` 이 된다 — AC-CM2-010 자신이 그 사실을 본문에 적고 AC-CM2-002~008 이 구멍을 막는다고 명시한다. iter-1 과 달리 그 방어층(D6/D7/D8)이 이제 명령을 갖고 있으므로 위험은 실질적으로 낮아졌으나, **구조는 변하지 않았다**.
2. **판정의 질은 여전히 사람이 본다.** 20행이 전부 형식을 만족하면서 얕은 판정일 수 있다. O-1 의 한 겹이 좌표 위조를 막지만, 부모 산문을 정확히 인용하고도 책임 귀속을 잘못 읽는 경우는 어떤 기계 검사도 잡지 못한다 — 리드의 evidence 판독이 최종 층이다.
3. **후보 수는 실행 시점에 달라진다.** 20 은 `52f863f36` 의 값이고 SPEC 이 그것을 명시한다. 다른 레인이 착지하면 A층·B층이 커지며, 그때 판정 부담도 함께 커진다 — Tier M 에서 그 상한이 어디인지는 규정돼 있지 않다.
4. **`merge-base` 와 `HEAD` 가 아직 같다**(둘 다 `52f863f36`). AC-CM2-009 는 이 구간에서 물지 않으며 SPEC 이 그 사실을 본문에 적고 regression-guard 로 읽는다. run 이 커밋을 쌓는 순간 갈라진다.
5. **`docs-truth.md` 의 손 갱신은 새로 만든 의무다.** REFRESH-001 이 같은 문서를 6문서 주장 아래 두고 `completed` 로 닫았으므로, 이 카드가 그 계승 결함을 실제로 닫는지는 run 증거(§③ 분리)에서만 확인된다.

---

## Recommendation

**PASS (0.88).** blocking 0건이므로 run-phase 진입을 막을 사유가 없다.

리드에게 남기는 세 가지:

1. **Implementation Kickoff Approval 은 여전히 필수다.** plan-auditor 의 PASS 는 그 human gate 를 대체하지 않으며 점수와 무관하다(`orchestration-mode-selection.md` 헤더).
2. **O-1(인용 해석 가능성)은 run 전에 넣으면 값이 싸고, run 후에는 넣을 자리가 없다.** 판정 표가 이미 쓰인 뒤에 좌표 대조를 요구하면 재작성이 된다. 넣을지 여부는 리드 재량이되 **시점은 지금**이다.
3. **run 첫 턴의 §C 재측정을 실제 판독으로 읽어라.** 이 감사 중에도 `origin/develop` 이 최소 한 번 움직였고(리드 `ef10a2524` → 내 판독 `9dddac882`), `merge-base` 와 described-worthy 델타는 안정적이지만 `64` 는 언제든 바뀔 수 있다.

운영자 결정 3건은 이번에도 어디에서도 재개하지 않았다 — 임계 40 / `gate.yaml` / `DefaultThresholds()` / 접힘 정책 / Go 코드는 소관 밖 유지, 판별식은 §A.3(a1) 그대로, `tree_root` 는 결함 아님. **20-vs-62 의 범위 결정도 재개하지 않았다** — 운영자가 인과 근거로 20 을 채택했고, 내가 검증한 것은 그 결정의 옳고 그름이 아니라 (a) 규칙이 실제로 20 을 뱉는가, (b) 좁힘을 정당화하는 이관 조건(잔여 42 목록)이 실재하고 AC 가 집행하는가 — **둘 다 성립한다.**

정책 규칙 적용 기록: 이 감사는 `.claude/rules/moai/development/verification-completeness.md` §1.1(빈 집합 위의 통과) · §1.3(비실행과 성공의 구별 불가) · §2(두 셀 채택 규율) · §2.1(RED-now 4요소 + undecidable disposition) · §4(증거 고정), 그리고 `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3·4 와 §2(baseline 귀속)를 적용해 판정했다.
