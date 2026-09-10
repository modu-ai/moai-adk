# SPEC 감사 보고서: SPEC-CODEX-PARTIAL-WIRING-001

- **Iteration**: 2 / 2 (Tier M 상한 — `.moai/config/sections/harness.yaml:77` `plan_audit_tier_ceilings.M: 2`)
- **판정**: **FAIL**
- **종합 점수**: **0.84** (Tier M 임계 0.80 — `spec-workflow.md:141`)
- **감사 트리**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499`, branch `WT-codex-partial-wiring`, HEAD `d69b38320`, base `ace1c5440` (`git merge-base --is-ancestor ace1c5440… HEAD` → rc=0)
- **작성자 추론 맥락 배제**: M1 Context Isolation에 따라 무시했다. 판정은 커밋된 산출물 4종과 소스를 직접 읽어 내렸다.

> **점수는 임계를 넘겼는데 판정이 FAIL인 이유**: 네 축 평균 0.84는 0.80을 통과한다. FAIL은 점수가 아니라
> `verification-completeness.md` §2의 [HARD] **뮤턴트 관문** 조항에서 나온다 — "채택 전에 그 기준을 만족하면서
> 요구를 위반하는 뮤턴트를 써 보라. 쓸 수 있으면 그 기준은 채택하기에 너무 얕다." AC-CPW-002에 대해 그런
> 뮤턴트를 **두 개** 실제로 구성했다(D1·D2). AC-CPW-002는 릴리스 게이트 6건 중 하나이고, 하필 iter1의 D3
> 수리를 지고 있는 기준이다. 수리가 이름만 들어간 것은 아니지만, 토큰 한 개 폭으로만 막혀 있다.

## 1. Must-Pass 결과 (7/7 통과)

| # | 항목 | 판정 | 근거 |
|---|---|---|---|
| MP-1 | REQ 번호 일관성 | PASS | `spec.md` §C 표 행 10개 = REQ-CPW-001…010, 결번·중복 없음, 3자리 패딩 일관 (`grep -c '^\| REQ-CPW-' spec.md` → `10`; `grep -o 'REQ-CPW-[0-9]\{3\}'` + `sort -u` → 001…010) |
| MP-2 | GEARS 형식 | PASS | 요구 계층 10건 전부가 다섯 패턴 중 하나로 서술되고 각 행이 유형을 자기 선언한다 — Ubiquitous(001·005·008), Event-driven(002·003), State-driven(006, "…인 동안"), Unwanted(004·007·009·010, "…해서는 안 된다(SHALL NOT)"). **검증 계층(AC-CPW-001…009의 Given-When-Then)은 이 기준의 대상이 아니며 §4에서 채점했다.** |
| MP-3 | YAML frontmatter | PASS | 정본 12필드 전부 존재·타입 적합(`version: "0.2.0"` 인용 문자열, `created`/`updated` ISO, `status: draft`, `priority: P2`, `lifecycle: spec-anchored`, `tags` 쉼표 구분 문자열). 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `tier: M` / `related_specs` 는 스키마 외 추가 필드 |
| MP-4 | §22 언어 중립성 | N/A | 대상이 Go CLI 검사 1본(`internal/cli/doctor_codex.go`)이며 템플릿·다언어 표면이 아니다 |
| MP-5 | D7 교차-SPEC 정합 | PASS | 참조 3건 전부 실재하며 상태 `completed` — SPEC-CODEX-WIRING-001 / SPEC-CODEX-SIDECAR-GUARD-001 / SPEC-INIT-HARNESS-PROMPT-001. retired·superseded·archived 0건 |
| MP-6 | D8 크로스플랫폼 | PASS | `grep -rn 'syscall' .moai/specs/SPEC-CODEX-PARTIAL-WIRING-001/` → 무출력(rc=1). 자동 통과 |
| MP-7 | 해명 관문 | PASS | `grep -rn 'NEEDS CLARIFICATION'` → 무출력(rc=1). Tier M이라 `research.md`는 없음 |

REQ/AC 예산: Tier M 상한 16/16 대비 10/9 — 여유.

## 2. 축별 점수

| 축 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.80 | 0.75~1.0 | 세 상태 모델(`unwired`/`half-wired`/`wired`)이 REQ-CPW-001에 명시적으로 정의되고 각 요구가 주체를 밝힌다. 감점 하나: REQ-CPW-003의 `CheckOK` 유지가 **무조건**으로 적혀 있는데, 그 갈래가 지나갈 `doctor_codex.go:189`의 사용자 계층 훑기(`codexStaleSkillFinding`)와의 상호작용이 어디에도 서술돼 있지 않다 — 합리적인 구현자 둘이 서로 다르게 풀 수 있고 사용자에게 보이는 결과가 갈린다(D2) |
| Completeness | 0.88 | 0.75~1.0 | HISTORY·WHY(§B)·WHAT(§A/§C)·HOW(§D)·요구·수용·범위 밖 전부 존재. 범위 밖은 `### Out of Scope — <주제>` H3 3개에 구체 항목 불릿. 0.2.0 이력이 좌표 오기를 지우지 않고 정정 절로 남긴 것은 규율에 맞다. 감점: §A.4의 "보존해야 하는 계약 3종"이 변경 지점에서 실제로 흔들릴 수 있는 **네 번째** 기존 동작(사용자 계층 훑기)을 세지 않았다 |
| Testability | 0.73 | 0.50~0.75 경계 | iter1 0.60 대비 실질 개선(RED-now 5셀·뮤턴트 3→5·반쪽 경로를 실제로 밟는 폭 시험 신설). 감점은 D1~D6에 모여 있다: 기준 하나가 뮤턴트 둘을 통과시키고, RED 셀의 명령 형태가 `verification-completeness.md` §2.1의 단일 호출 형식을 벗어나며, 인용된 stdout이 캐시·시간 토큰을 포함해 바이트 재현이 되지 않는다 |
| Traceability | 0.95 | 0.75~1.0 | **양방향 직접 측정**: 매트릭스 매핑 열만을 대상으로 REQ별로 세었을 때 REQ-CPW-001…010 전부 ≥1행(001:2, 002:1, 003:1, 004:2, 005:1, 006:2, 007:1, 008:1, 009:1, 010:1). 고아 AC 0건, 존재하지 않는 REQ를 가리키는 AC 0건. iter1이 놓친 축약 표기(`REQ-CPW-001, 002`)는 매트릭스에서 사라졌고(`acceptance.md:20-28` 전부 전체 ID), 잔존은 `progress.md:19`의 인용 문장 한 곳뿐 — 그 문장은 누락 사실을 기술한 산문이므로 매핑이 아니다. D5 재매핑(AC-CPW-008 → REQ-CPW-010)도 주체가 맞다 |

산술 평균 = (0.80 + 0.88 + 0.73 + 0.95) / 4 = **0.84**.

**점수 회귀 없음**: iter1 0.76 → iter2 0.84. STOP 상신 조건(iter(N+1) < iter(N))에 해당하지 않는다.

## 3. 결함 (Defects)

> 분류: **blocking** = SPEC의 정확성·내부 정합, 또는 이 문서가 실제로 내건 기준에 영향.
> **optional** = 그 외(더 풍부할 수 있었던 것, 취향).

### D1 — AC-CPW-002가 "지시문 우회" 뮤턴트를 통과시킨다 · blocking · major

`acceptance.md` AC-CPW-002 Then (d) — 단언이 **상수 토큰** `initCodexAdvice`(= `run moai init --agent codex`) 부재로 적혀 있다. 다음 뮤턴트를 실제로 구성했다:

```
half-wired × codex 부재 갈래:
  Status  = uikit.CheckOK
  Message = "codex agent definitions present, no wiring files — `moai init --agent codex` wires them"
```

아홉 기준 통과 여부를 한 항목씩 대조한 결과: (a) `CheckOK` ✓ / (b) agent 정의 존재 진술 ✓ / (c) `claude-only` 0회 ✓ / **(d) `run moai init --agent codex`를 부분문자열로 포함하지 않는다** — `run ` 접두가 없다 — ✓ / AC-CPW-007 폭 ~72 rune ✓ / AC-CPW-001·003·004·005·006·008 무관 ✓ / AC-CPW-009 뮤턴트 5종 전부 여전히 RED ✓.

즉 **Codex를 쓸 생각이 없는 사용자의 모든 프로젝트에 실행 불가능한 지시를 상시로 거는 구현이 아홉 기준을 전부 만족한다.** 이것은 REQ-CPW-009가 금지하려던 바로 그 동작이고 `spec.md` §D-1의 "잔소리는 늘리지 않는다"를 정면으로 어긴다. iter1의 D3(단언 자체가 없었음)보다 좁아졌을 뿐 같은 계열이다.

**요구되는 수리**: (d)를 토큰이 아니라 **성질**로 다시 쓴다 — 부재 갈래 Message는 `initCodexAdvice`도, 부분문자열 `moai init`도 담지 않는다. 더해 M4의 짝으로 **M4′**(부재 갈래에 `run ` 없는 패러프레이즈 지시문을 심는다 → 새 단언이 RED)를 뮤턴트 표에 추가한다. 새 단언의 red를 관측하지 않으면 이 수리도 같은 자리에 머문다.

### D2 — AC-CPW-002의 Given이 사용자 계층 설정 환경을 통제하지 않는다 · blocking · major

`acceptance.md` AC-CPW-002 Given — `codexWiringLookPath`만 고정하고 `stubCodexHome`을 고정하지 않는다. 이 파일의 **기존 시험은 예외 없이 전부** 고정한다(`internal/cli/doctor_codex_test.go` 안 `stubCodexHome(` 호출 30여 곳: :154, :196, :249, :273, :291, :308, :357, :377, :401, :447, :521, :534, :550, :566, :580 …). 관례가 아니라 기준 쪽이 빠뜨렸다.

이 구멍으로 두 번째 뮤턴트가 통과한다:

```
half-wired × codex 부재 갈래를 조기 반환시키지 않고 본 경로로 흘려보낸다
→ internal/cli/doctor_codex.go:189 의 codexStaleSkillFinding() 이 실행된다
```

깨끗한 CI 머신에서는 `~/.codex/config.toml`이 없어 `ok=false`(fail-open) → 소견 0건 → 시험은 초록. 그러나 사용자 계층에 낡은 skill 등록이 남은 실제 머신에서는 소견이 생겨 `problems`가 비지 않고 **Status가 `CheckWarn`으로 뒤집힌다** — REQ-CPW-003의 "`CheckOK`로 유지한다"를 운영 환경에서만 위반하며, 어떤 기준도 그것을 잡지 못한다. 반대로 구현자가 강제로 OK를 박으면 진짜 소견을 조용히 삼킨다.

배경 사실 두 개는 확인했다. (1) 오늘 이 집단은 `!wired && !codexInstalled` 조기 반환으로 침묵하므로(`internal/cli/doctor_codex.go:96-103`) 조기 반환을 유지하는 구현에는 회귀가 없다. (2) `internal/cli/doctor_codex.go:186-188`의 주석은 이 훑기가 "Codex가 관여할 때만 도달한다"고 적고 있는데, half-wired × codex 부재가 그 전제를 깨는 첫 사례다. **SPEC은 이 갈래에서 사용자 계층 훑기가 도는지 마는지를 한 줄도 말하지 않는다.**

**요구되는 수리**: Given에 "낡은 `[[skills.config]]` 항목을 하나 이상 선언한 home을 `stubCodexHome`으로 고정"을 넣고, Then에 그 상태에서도 `CheckOK`임을 단언한다. 뮤턴트 **M6**(부재 갈래를 본 경로로 흘려보낸다 → RED)를 추가한다. 아울러 `spec.md` §A.4에 이 훑기를 네 번째 보존 계약으로 세우거나, REQ-CPW-003에 "이 갈래에서 사용자 계층 훑기는 도달하지 않는다"를 명시한다.

### D3 — `moai spec lint` 무결 주장이 공허한 초록이다 · blocking · minor

`progress.md`가 plan-audit-ready 증거로 lint 무결을 든다(리드 브리핑에도 "`moai spec lint` exit 0 with `✓ No findings`"로 올라왔다). 측정했다:

- 전 저장소 실행: `moai spec lint > /tmp/t499-lint.txt 2>&1` → exit 0, `0 error(s), 4338 warning(s)`
- 이 SPEC을 이름으로 잡는 행: `grep -c 'SPEC-CODEX-PARTIAL-WIRING-001' /tmp/t499-lint.txt` → **0**

여기까지는 주장과 맞다. 그러나 **왜** 0인지가 문제다. lint의 REQ 수집기는 목록 항목 줄만 잡는다(`internal/spec/lint_req_widen.go:59` `reqLineWidePattern` = `^\s*[-*]\s+…`). 이 SPEC의 요구는 전부 **표 행**이다:

```
목록형 REQ 정의 줄(파서가 수집하는 형태): 0
표 행 REQ 정의 줄(이 SPEC이 실제로 쓰는 형태): 10
```

`doc.REQs`가 비면 `CoverageRule.Check`는 첫 줄에서 `return nil` 한다(`internal/spec/lint.go:907-909`). 즉 `CoverageIncomplete`·`ModalityMalformed`·`InvalidREQID`·`DuplicateREQID` 네 규칙이 **이 SPEC을 한 번도 방문하지 않았다.** 대조군으로 확인: SPEC-ZONE-REGISTRY-RESYNC-001은 목록형 REQ 줄 15개를 갖고 lint가 그 SPEC에 대해 정확히 15건의 `CoverageIncomplete`를 냈다(REQ-ZRR-001…015). 파서는 잘 돈다 — 이 SPEC의 형태를 못 볼 뿐이다.

`verification-completeness.md` §1.1이 이름 붙인 **훑은 집합이 빈 통과**이며, 판정 전에 훑은 개수를 세라는 의무가 지켜지지 않았다. iter1도 같은 침묵을 추적성의 방증으로 인용했다(`.moai/reports/t499/plan-audit-iter1.md:212`) — 순환이다.

**요구되는 수리**: `progress.md`의 lint 문장에 훑은 집합을 붙인다 — "lint 전 저장소 실행에서 이 SPEC을 이름으로 잡는 행 0건. 단 이 SPEC의 요구는 표 행 형태라 lint의 REQ 기반 4규칙은 이 SPEC을 평가하지 않았다(수집된 REQ 0건). 커버리지는 lint가 아니라 직접 측정으로 성립한다." **SPEC의 요구 표기를 목록형으로 바꾸라는 뜻이 아니다** — 표 형태는 이 저장소의 통용 서식이고 감사 루브릭도 그것을 요구하지 않는다. 고쳐야 하는 것은 증거 문장이지 SPEC이 아니다.

### D4 — D1 수리가 새로 심은 좌표 오기 2건 · blocking · minor

`acceptance.md` §D.1 AC-CPW-007의 "왜 기존 두 시험만으로는 안 되는가" 문단 — "두 시험 모두 `checkCodexWiring(t.TempDir(), …)`를 호출한다(`internal/cli/doctor_codex_test.go:394`, `:441`)".

실측: 두 호출의 실제 좌표는 **:404**와 **:450**이다.

```
$ grep -n 'checkCodexWiring(' internal/cli/doctor_codex_test.go
…
404:	check := checkCodexWiring(t.TempDir(), false)
450:		check := checkCodexWiring(t.TempDir(), verbose)
```

인용한 :394 / :441은 각각 `for i := 0; i < 49; i++` 줄과 그 근처다. **주장 자체는 참이다** — 두 시험 모두 빈 `t.TempDir()`을 넘겨 `unwired` 경로만 밟으므로 반쪽 경로의 폭을 관측하지 못한다는 D1 논거는 그대로 성립한다. 틀린 것은 좌표뿐이다.

이것은 iter1의 D6(§A.3 좌표 `:105`→`:92`)과 **같은 계열의 결함이 수리 라운드에서 재발한 것**이다. 결론이 옳아도 논거의 좌표는 틀릴 수 있고, 커밋된 기록의 한 줄 오류는 done 이후에 고칠 계기가 없다.

**요구되는 수리**: `:394` → `:404`, `:441` → `:450`. 정정 사실을 지우지 말고 §A.3의 D6 정정과 같은 방식으로 남긴다.

### D5 — RED-now 셀의 명령 형태가 §2.1의 단일 호출 형식을 벗어난다 · blocking · minor

`verification-completeness.md` §2.1은 릴리스 차단 기준의 RED-now 셀에 네 요소를 함께 요구하며, **명령**에 대해 이렇게 적는다: "단일 호출로 완결되는 읽기 전용 셸 호출. 파이프, 리다이렉션, `&&`, `;` 연쇄, 서브셸은 이 형식 밖이다." 그리고 **종료 코드는 별도 필드**여야 하는 이유를 "단일 호출 형식이 `; echo $?`를 금지하기 때문"이라고 명시한다.

다섯 셀은 전부 `… > /tmp/t499-redN.txt 2>&1; echo $?` 형태로, 리다이렉션과 `;` 연쇄를 둘 다 쓴다. 같은 §2.1은 형식 밖 인용에 대해 **판정 불가 처분**(릴리스 차단 자격 상실 → 회귀 가드로 강등)을 지정한다. 문자 그대로 적용하면 릴리스 게이트 6건이 1건(AC-CPW-009)으로 줄어든다.

다만 §2.1 자신이 "의무는 어휘가 아니라 구조에 대한 것 — 명령이 있는가, stdout이 있는가, 종료 코드가 있는가, SHA가 고정돼 있는가"라고도 적는다. 그 구조 검사로는 네 요소가 전부 있고 읽을 수 있다. 그리고 이 문서가 리다이렉션을 쓰는 이유는 자의적이지 않다 — 머리의 [HARD] 절이 zsh가 `PIPESTATUS`를 안 채워 빈 종료 코드가 또 하나의 공허한 초록이 되는 것을 막으려는 것이다. **두 규율이 실제로 충돌하며, 나는 강등을 집행하지 않고 재작성을 권고한다.**

**요구되는 수리**: 셀당 세 줄로 분해한다 — 명령 `go test ./internal/cli/ -run TestX`(단일 호출, 리다이렉션 없음) / 인용 stdout / **종료 코드를 자기 필드로** `exit: 0`. 머리의 zsh [HARD] 절은 "종료 코드를 어떻게 읽었는가"에 대한 주석으로 남기고 인용 명령에서는 뺀다. 문서 머리의 SHA 고정은 그대로 네 번째 요소를 만족한다.

### D6 — 인용된 stdout이 바이트 재현되지 않는다 · optional · minor

다섯 셀의 인용 stdout은 캐시/시간 토큰을 포함한다(`(cached)` 1건, `0.939s`·`0.686s`·`0.691s`·`0.683s` 4건). 내가 같은 트리에서 다섯 명령을 그대로 재실행한 결과는 **다섯 개 전부 `(cached)`** 였다. 판정(exit 0 + `[no tests to run]`)은 동일하게 재현됐지만, "verbatim"을 바이트 비교로 읽는 독자에게는 불일치로 보인다.

**권고**: 인용을 판정에 실린 부분(`ok … [no tests to run]`)으로 한정하거나, 시간·캐시 토큰이 실행마다 달라진다는 한 줄을 붙인다. 강제 사항으로 보지 않는다.

## 4. 판단 (결함 아님)

명시적으로 결함과 분리한다. 이 넷은 내가 다르게 썼을 지점이지 잘못이 아니다.

1. **AC-CPW-001(b)·AC-CPW-002(b)의 "진술한다"** — 단언할 토큰을 지정하지 않는다. 그러나 `plan.md` §B-3이 문구를 run-phase의 사람 결정으로 명시적으로 유보했으므로, 여기서 정확한 문자열을 못 박으면 그 유보와 충돌한다. 현 상태가 타당한 절충이다.
2. **REQ-CPW-009가 Message만 구속한다** — Detail에 지시문을 싣는 구현은 요구를 어기지 않는다. Detail은 `--verbose`에서만 렌더되므로 사용자가 선택한 표면이고, 잔소리로 보지 않는다.
3. **REQ-CPW-005가 판별식에 개수 금지라는 HOW를 담는다** — 요구에 구현 세부가 섞였다고 읽을 여지가 있으나, §D-3이 근거(템플릿 편집 한 번이면 11이 바뀐다)를 남겼고 실질은 견고성 요구다.
4. **§F는 릴리스 게이트를 6건, `acceptance.md` §D.4 첫 항목은 5건으로 센다** — 후자가 AC-CPW-009를 별도 항목으로 뺀 것이므로 실질은 일치한다(5 + 뮤턴트 관문 1 = 6). 혼동 소지는 있으나 모순은 아니다.

## 5. 리드가 물은 네 가지에 대한 답

### (1) D3 수리는 구멍을 막았는가 — **절반만**

REQ-CPW-009는 실재하고 AC-CPW-002 (d)에 단언이 붙었다. iter1이 지목한 뮤턴트(부재 갈래에 `initCodexAdvice`를 그대로 붙이는 M4)는 이제 확실히 RED가 된다. 그러나 **패러프레이즈 한 번으로 빠져나간다**(D1). 요구는 성질로 쓰였는데(조치 지시문 금지) 단언은 토큰으로 쓰였고, 그 간극이 뮤턴트가 사는 자리다.

### (2) 다섯 RED-now 셀은 건전한가 — **정직하지만 §2의 의미에서 RED는 아니다**

네 요소는 전부 있다(명령·stdout·종료 코드·문서 머리 SHA 고정, 그리고 `git status --short -- internal/` 무출력으로 관측 시점 트리 청결까지). 다섯 개를 직접 재실행해 **exit 0 + `[no tests to run]`을 전부 재현**했고, 별도로 `grep -rn 'func TestCheckCodexWiring_HalfWired' internal/cli/` → 무출력(rc=1)으로 다섯 시험의 부재를 직접 확인했다.

`ok … [no tests to run]` exit 0을 RED로 읽는 것이 §1.1이 경고한 빈 훑기 위험을 다른 모자를 쓰고 되풀이하는 것인가 — **아니다.** §1.1은 빈 훑기를 **통과로 읽는 것**을 금지하며, 판별 증거로 러너의 `[no tests to run]` 토큰을 지목한다. 이 문서는 그 토큰을 인용하고 "이 `ok`는 통과가 아니다"라고 명시하며, GREEN 경로에 **`[no tests to run]` 부재 확인**을 의무로 박았다. §2.1의 사고 사례(셀이 트리만 고정하고 출력을 안 실어 "한 번도 하지 않은 측정"을 가리켰던 SPEC)와 갈리는 지점이 정확히 여기다.

다만 정확히 말하면 이것은 **RED 관측이 아니라 계측기 부재 관측**이다. §2의 RED-now가 하는 일(공허한 기준과 의미 있는 기준을 가르는 것)을 시험 부재는 하지 못한다. 그 일은 전부 AC-CPW-009의 뮤턴트 5종이 지고 있고, 그것은 아직 실행되지 않은 run-phase 의무다. 문서도 그렇게 적고 있다(§D.3 말미). 남는 흠은 형식 쪽이다(D5·D6).

### (3) 6게이트 / 3가드 분할은 정직한가 — **정직하다**

AC-CPW-003·004·008은 셋 다 순수 보존 기준이다: 003은 기존 세 시험 무수정 통과, 004는 골든 3본 무재생성, 008은 `internal/codexwiring` 무접촉. 구현 전에도 초록이고 구현 후에도 초록이어야 하므로 원리상 RED-now를 가질 수 없다. **RED-now를 면하려고 가드로 내린 기준은 하나도 없다.** 반대 방향의 확인도 했다 — 게이트 6건 중 RED-now를 못 가질 이유가 있는 것은 없다.

부수 확인: AC-CPW-006이 근거로 삼은 전제("종료 코드는 `CheckFail` 수만 센다 — `CheckWarn`에는 둔감")를 소스에서 검증했다. `internal/cli/doctor.go:127-134` `countFailedChecks`가 `CheckFail`만 세고, `:139-148` `doctorExitStatus`가 "Warn-only runs stay exit 0"을 명시한다. 전제는 참이며, 따라서 REQ-CPW-002의 Warn이 REQ-CPW-007의 종료 코드 절을 어기지 않는다.

### (4) M4·M5는 실제로 RED가 되는가 — **된다**

- **M4**(부재 갈래 Message 끝에 `— run moai init --agent codex`) → AC-CPW-002 (d)의 `initCodexAdvice` 부재 단언에 정면으로 걸린다. RED.
- **M5**(반쪽 갈래 Message 200 rune) → 신규 `..._HalfWiredMessageWidthStaysInBand`가 113 상한을 단언하므로 RED. 구현이 `joinCodexSummaries`를 거치더라도 마찬가지다: 반쪽 소견이 선두 단일 요약이면 `joinCodexSummaries`의 "선두 요약 단독 초과는 통째로 방출" 예외(`internal/cli/doctor_codex.go:223-225, 234-246`)를 타 200 rune가 그대로 나온다. 절단으로 우연히 통과할 길이 없다.
- M1·M2·M3도 대응 시험이 실재하는 한 RED가 된다. 특히 M2는 AC-CPW-005의 **빈 디렉터리 케이스**에 전적으로 의존하며, 문서가 그 의존을 명시한 것은 옳다.

다만 M4는 D1의 패러프레이즈 뮤턴트를 잡지 못한다 — M4′가 필요한 이유다.

## 6. iter1이 놓친 두 가지 — 지금 상태와, 검사 자체에 대한 평가

### (a) 축약 REQ 표기 — **수리 확인됨**

`acceptance.md:20-28` 매트릭스 아홉 행이 전부 전체 ID를 쓴다. 매핑 열만을 대상으로 REQ별로 센 결과 10건 전부 ≥1행(§2 Traceability 행에 수치 기재). 축약형 잔존은 `progress.md:19` 한 곳이며 그것은 누락 사실을 기술하는 인용문이지 매핑이 아니다.

### (b) iter1의 추적성 검사는 그것을 잡았어야 하는가 — **잡았어야 했고, 검사가 약했다. 그것도 두 겹으로.**

iter1은 "REQ-CPW-001…008 전부가 매트릭스에 나타난다"고 단언했다(`plan-audit-iter1.md:212`). 당시 매트릭스가 `REQ-CPW-001, 002, 004` 꼴이었다면 전체 ID 대조는 002·003·004를 **고아로 보고했어야** 한다. 그러지 않았다는 것은 대조가 전체 ID 단위로 이뤄지지 않았다는 뜻이다 — 표를 읽고 사람 눈으로 납득한 것이지 측정한 것이 아니다. 검사는 돌았고(그래서 D5 매핑 오류를 잡았다) 입도가 틀렸다.

두 번째 겹이 더 나쁘다. iter1은 같은 문장에서 **`moai spec lint`의 `CoverageIncomplete` 침묵을 방증으로 인용했다.** D3에서 보였듯 그 침묵은 lint가 이 SPEC의 REQ를 하나도 수집하지 못해서 생긴 것이다. 즉 방증이 아니라 같은 맹점을 가진 두 번째 관측이었고, 독립 확인처럼 보이는 순환이었다. 부재를 증거로 삼되 그 부재가 도달 실패인지 판정인지를 가르지 않은 것이다.

**검사에 대한 처방**(이 카드 밖, 감사 절차 쪽): 추적성은 (i) **매핑 열만**을 대상으로 (ii) **전체 ID**로 (iii) REQ별로 세는 명령 한 줄로 측정하고, 그 출력을 판정서에 싣는다. 그리고 **lint의 침묵은 절대 방증으로 쓰지 않는다** — 쓰려면 먼저 그 SPEC에서 lint가 REQ를 몇 개 수집했는지를 세야 하며, 그 수가 0이면 lint는 그 축에 대해 아무 말도 하지 않은 것이다.

lint 쪽 후속 후보에 대한 레인의 관측(2)에는 동의하되 **원인을 정정한다**: 표기 축약이 아니라 **정의 줄 형태**가 원인이다. `CoverageRule`은 `maps REQ-…` 선언이 있는 acceptance.md만 커버리지로 인정하고(`internal/spec/lint_coverage_sibling.go:104-118`), 그 이전에 `doc.REQs`가 비면 아예 반환한다. 표 형태로 요구를 쓰는 SPEC은 REQ 기반 4규칙 전체의 사각지대다. 이 저장소에서 표 형태는 드물지 않으므로 별도 카드 값어치가 있다.

## 7. 회귀 점검 (iter1 지적 7건)

| iter1 | 상태 | 근거 |
|---|---|---|
| D1 — AC-CPW-007이 반쪽 경로를 못 밟는다 | **RESOLVED(단, D4)** | 신규 `..._HalfWiredMessageWidthStaysInBand`가 게이트를 지고 기존 2종은 회귀 가드로 강등. 논거는 실측으로 참 — 두 기존 시험은 :404·:450에서 빈 `t.TempDir()`을 넘긴다. 인용 좌표만 틀렸다(D4) |
| D2 — 게이트에 RED-now 셀 부재 | **RESOLVED(단, D5·D6)** | 5셀 신설, 다섯 개 전부 내가 재실행해 재현. SHA 고정은 문서 머리, 관측 시점 `internal/` 청결까지 기재. 명령 형태와 stdout 재현성에 흠 |
| D3 — 부재 갈래 지시문 금지 요구·단언 부재 | **PARTIAL** | REQ-CPW-009 신설 + AC-CPW-002 (d) 단언은 확인. 그러나 패러프레이즈 뮤턴트가 통과한다(D1) |
| D4 — AC-CPW-001에 `claude-only` 0회 단언 | **RESOLVED** | (e)에 존재하며 Message + Detail 양쪽을 잰다 |
| D5 — AC-CPW-008 매핑 오류 | **RESOLVED** | REQ-CPW-010 신설, AC-CPW-008 → REQ-CPW-010 재매핑(`acceptance.md:27`), 정정 이력 명시. 새 요구의 주체가 `codexwiring`이라 매핑이 맞다 |
| D6 — §A.3 좌표 `:105` | **RESOLVED** | `grep -n 'wired := hooksErr' internal/cli/doctor_codex.go` → `92:`. §A.3이 `:92`로 정정됐고 오기 사실을 지우지 않았다 |
| D7(optional) — §D-1의 보호 대상 구분 | **RESOLVED** | `spec.md` §D-1 0.2.0 보강 문단. 논지도 맞다 |

**정체 결함 없음**: 두 라운드에 걸쳐 변하지 않은 지적은 없다. iter1 지적 7건 중 5건 완결, 1건 부분, 1건은 논거 좌표만 잔존.

## 8. 권고

이번이 Tier M 상한(2회)의 마지막 iteration이므로, FAIL은 자동으로 iter3을 부르지 않는다. 리드가 세 갈래 중 하나를 고른다.

**A. 결함 수리 후 재감사(override로 iter3 허용)** — 권고. 수리는 다섯 곳, 대부분 `acceptance.md`와 `progress.md`이며 `spec.md`는 §A.4 한 문단만 손대면 된다:

1. AC-CPW-002 (d)를 토큰이 아닌 성질로 재작성(부분문자열 `moai init` 금지) + 뮤턴트 **M4′** 추가 — D1
2. AC-CPW-002 Given에 `stubCodexHome`(낡은 항목 ≥1) 고정 추가, Then에 그 상태에서도 `CheckOK` 단언 + 뮤턴트 **M6** 추가 + `spec.md` §A.4에 사용자 계층 훑기를 네 번째 보존 계약으로 추가 — D2
3. `progress.md`의 lint 문장에 훑은 집합을 붙여 재작성 — D3
4. `acceptance.md`의 `:394`/`:441` → `:404`/`:450` — D4
5. 다섯 RED-now 셀을 명령 / stdout / `exit:` 세 필드로 분해 — D5

**B. PASS-with-debt** — D1·D2를 run-phase에 명시된 부채로 넘긴다. 그 경우 run-phase 진입 조건으로 **M4′와 M6의 RED 관측을 DoD에 못박아야** 한다. 그것 없이 넘기면 두 뮤턴트가 그대로 착지할 수 있고, 그때 사용자에게 보이는 결과는 iter1이 막으려던 바로 그것이다.

**C. 범위 축소** — 이 SPEC에는 권하지 않는다. 반경(소스 1본 + 시험 1본)이 이미 작고, 결함은 범위가 아니라 기준의 깊이에 있다.

## 9. PASS로 갈 경우 run-phase로 넘어가는 잔여 위험

점수만 보면 통과이므로, 리드가 B를 택할 경우를 위해 명시한다.

1. **패러프레이즈 지시문**(D1) — 구현자가 "run "을 뺀 지시문을 부재 갈래에 실으면 아홉 기준을 전부 통과하며 착지한다. 사용자가 보는 결과: Codex를 안 쓰는 모든 프로젝트의 `moai doctor`에 실행 불가능한 안내가 상시 노출.
2. **사용자 계층 훑기 상호작용**(D2) — 부재 갈래를 본 경로로 흘려보내는 구현은 CI에서 초록이고 사용자 머신에서만 `CheckWarn`으로 뒤집힌다. 반대로 강제 OK 구현은 진짜 소견을 삼킨다. 어느 쪽도 시험이 잡지 못한다.
3. **뮤턴트 5종이 아직 미실행** — 이 SPEC의 반증 능력은 전부 run-phase의 M3 마일스톤에 예치돼 있다. 다섯 RED-now 셀은 계측기 부재를 보였을 뿐 기준의 결속력을 보이지 않았다. M1이 RED가 되지 않으면 신규 시험이 판별식을 잡고 있지 않다는 뜻이며, 그 상태에서는 여섯 게이트 전부가 공허하다.
4. **문구 결정이 run-phase에 열려 있다**(`plan.md` §B-3) — 의도된 유보이나, 위 1의 위험과 같은 자리에 있다. 문구를 정하는 순간이 D1이 실현되거나 회피되는 순간이다.
5. **`phase: "v3.1.5 target"`** — 브랜치·태그를 세어 값이 관측과 일치함은 확인됐으나 로드맵 확인은 아니다(`plan.md` §I). 스키마·era 분류 어디에도 영향이 없어 감사 축에서는 무해하다.

---

**감사자**: plan-auditor (iteration 2/2, Tier M)
**감사 일자**: 2026-09-07
**측정 트리**: `d69b38320` (base `ace1c5440`), `git status --short -- internal/` 무출력
