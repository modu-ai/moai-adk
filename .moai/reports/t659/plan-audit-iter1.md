# SPEC 감사 보고서: SPEC-CON-AMEND-APPLY-001

Iteration: 1/3
Verdict: FAIL
Overall Score: 0.78 (Tier L 합격선 0.85 미달)

- 감사 대상 트리: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t659`, 브랜치 `WT-amend-apply`, HEAD `76144d40ac000097ccac954ffd7fa79bd432f5e4` (`git rev-parse --show-toplevel` / `git branch --show-current` / `git rev-parse HEAD` 로 확인, 지시와 일치)
- 대상 개정: `spec.md` `version: "0.1.4"`, `tier: L`
- 읽은 산출물: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, `.moai/reports/t659/verdict.md` 전체, 코드 `internal/constitution/{pipeline,loader,evolution_log,amendment,rate_limiter,human_oversight,contradiction}.go`, `pipeline_test.go`, `evolution_log_test.go`, `internal/cli/constitution.go`, 그리고 영향 범위 확인을 위해 `internal/spec/lint.go`, `internal/spec/lint_test.go`, `internal/cli/spec_lint.go`, `internal/cli/doctor.go`, `internal/constitution/validator.go`, `registry_sync_test.go`
- 위임문의 판정서·규칙 인용은 감사 범위 지정으로만 취급했다. 작성자 추론 맥락은 받지 않았다(M1 Context Isolation).
- 교차 모델 감사 MCP 도구(`mcp__moai__audit_multi` 등)는 이 세션에 없어 호출하지 않았다. 판정은 이 감사자 단독이다.
- 실행하지 않은 것: `go test`, `go build`, `moai todo`. 아래 코드 동작 주장은 모두 **코드 판독**이며, 실행 관측이 아니다.

## 판정 요약

필수 항목 7개는 모두 통과했다. 그러나 점수가 0.85 에 못 미치고, 합격 여부와 무관하게 먼저 고쳐야 할 blocking 결함이 4건(major) 있다. 세 건은 뮤턴트 탐침에서 나왔다. 선언된 뮤턴트 하나는 원리상 죽일 수 없고(D1), 인수 조건을 통과하면서 요구사항을 어기는 뮤턴트를 두 개 작성할 수 있었다(D2·D3). 나머지 한 건(D4)은 공유 로더를 바꾸는 영향 범위를 계획이 보지 못해, 다른 패키지의 기존 테스트가 깨질 가능성이다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `spec.md` L75–L143 에 `REQ-CAA-001` … `REQ-CAA-021` 이 각 1회씩, 빈 번호·중복 없음(`grep -oE '^- \*\*REQ-CAA-[0-9]+'` 결과 21개 모두 1회). 번호가 문서 안에서 순서대로 놓이지는 않지만(016·017이 §D.1·§D.2 중간에 삽입) 연속성과 유일성은 성립한다.
- [PASS] MP-2 GEARS 형식 (요구사항 층에 대해 판정): 21개 REQ 모두 `shall` 과 When/While/Where/보편형/`shall not` 중 하나로 서술된다(예: L75 "**When** the apply step updates …, the apply step shall …", L105 "shall not drop", L135 "**While** `Execute` runs in dry-run mode"). AC 층의 Given-When-Then 은 이 항목에서 채점하지 않았다. 경미 사항은 D14.
- [PASS] MP-3 YAML frontmatter: L2–L13 에 12개 필수 필드가 모두 있고 타입이 맞다(`version: "0.1.4"` 인용, `created`/`updated` ISO 날짜, `priority: P1`, `phase: "v3.2.0 target"` 는 생애주기 토큰이 아님, `lifecycle: spec-anchored`, `tags` 는 쉼표 문자열). 거부 별칭(`created_at` 등) 없음. `related_specs` 는 스키마 밖 필드지만 금지 대상이 아니다.
- [N/A] MP-4 언어 중립성: 템플릿 대상이 아닌 Go 내부 패키지(`internal/constitution`, `internal/cli`) 변경이다.
- [PASS] MP-5 D7 교차 SPEC: 본문 참조는 `SPEC-V3R2-CON-002` 하나, 그 `status: implemented`(retired/superseded/archived 아님). BLOCKING 없음.
- [PASS] MP-6 D8 교차 플랫폼: `grep -c syscall spec.md` → 0. 자동 통과.
- [PASS] MP-7 확인 요청 표식: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → 출력 없음, exit 1.

## Tier 와 예산

| 항목 | 계수(이번 실행) | Tier L 상한 | 판정 |
|---|---|---|---|
| 요구사항 | 21 (`grep -ohE 'REQ-CAA-[0-9]+' *.md \| sort -u \| wc -l` → 21) | 25 | 범위 안 |
| 인수 조건 | 25 (같은 방식 → 25, `^### AC-CAA-` 헤딩 25) | 25 | **상한과 같음** — 범위 안, 여유 0 |
| 뮤턴트 | 29 (`^\| M-` 행 29) | 상한 없음 | — |

상한은 두 계수에 각각 따로 적용되며 둘 다 넘지 않는다. 다만 인수 조건은 상한에 붙어 있다. 아래 수정 권고는 새 AC 번호를 만들지 않고 기존 AC 의 행·사례로 넣도록 제안한다.

## Category Scores

| 차원 | 점수 | 기준 구간 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 대부분 단일 해석. 다만 AC-CAA-014 "an error of the same kind"(L208)는 이진 판정이 아니고, REQ-CAA-021 "before any backup, temporary file, or write"(spec L95)는 실제 모드에서 먼저 쓰는 락 파일과 충돌하며, AC-CAA-003 "When the apply step validates"(L104)는 호출 지점을 정하지 않는다. |
| Completeness | 0.85 | 0.75–1.0 사이 | 필수 섹션(HISTORY L20, 문제 §A L32, 목표 §B L43, 요구사항 §D L69, Out of Scope `### Out of Scope — …` L170 이하)과 Tier L 산출물 5+1개가 모두 있다. 그러나 공유 로더 변경의 호출자 영향(D4)이 계획·위험·검증 범위 어디에도 없다. |
| Testability | 0.70 | 0.50–0.75 사이 | 대부분의 AC 는 사전 스냅숏·호출 횟수·대조군을 갖춘 강한 형태다. 그러나 선언 뮤턴트 M-5b 는 죽일 수 없고(D1), 요구사항을 어기며 통과하는 뮤턴트 둘(D2·D3), 잘못된 이유로 빨간 하위 사례 하나(D6)가 있다. |
| Traceability | 0.80 | 0.75 | ID 수준에서는 21 REQ 모두 AC 가 있고 AC 가 가리키는 REQ 는 모두 존재한다(acceptance.md §D.0 L50–L76). 그러나 REQ-CAA-021 의 링크 해석 조항은 세 지점 중 한 곳만, REQ-CAA-010 의 백업·임시 쓰기 실패 복원과 REQ-CAA-013 의 환경 변수 경로는 어떤 AC 도 확인하지 않는다. |

산술 평균 0.775 → 0.78. 합격선 0.85 미달.

## 판결 준수 (Ruling fidelity)

| 판결 | 인코딩 | 확인 |
|---|---|---|
| Q1 정확히 1회·정규화 금지·anchor 좁히기 금지 | REQ-CAA-001/002, AC-001/002/003, M-1/M-2 | 일치 |
| Q2 한 줄 치환 + 재파싱 검증, 재직렬화 기각 | REQ-CAA-003/004, AC-004/005, M-3a/3b/4 | 일치. 단 AC-005(b)는 D6 |
| Q3 snake_case 태그·구 키 호환·zone 등록부 형식·사람 작성 항목 읽기 | REQ-CAA-005…009, AC-006…011 | 일치. AC-007 RED 셀은 D5 |
| Q4 백업→임시→원문·등록부·로그 순 rename→전부 복원, 2·3번째 rename 실패 주입, dry-run 검증 실행 | REQ-CAA-010/011/012, AC-012/013/014 | 순서·주입 지점 일치. 로그 부재 복원 검증은 D1 |
| Q5 범위 분리 | spec §F L170–L176 | 일치 |
| §7 CLI 슬롯 미부여 | §E.3, AC-015 | 일치. "home seam" 전제 정정(L162)은 사실 정정이지 판결 뒤집기가 아니다 |
| G1 fail-closed, 파일·줄·키 | REQ-CAA-009, AC-018, M-15 | 일치 |
| G2 새 clause 0회 | REQ-CAA-016, AC-019, M-16 | 일치 |
| G3 `Execute` 안 Before 검사 | REQ-CAA-017, AC-020, M-17 | 일치 |
| G4 백업 보존·경로 반환·실패 주입 1개 | REQ-CAA-018, AC-021, M-18 | 일치 |
| G5 CLI 는 dry-run 만 | §E.4 | 일치 |
| 범위 추가 (a) 단일 해석기, AC 1개 | REQ-CAA-019, AC-022, M-19 | 일치 |
| G6 (ii) 로더 거부가 경계 | REQ-CAA-020, AC-023, M-20 | 일치 |
| G7 (A) Clean+Abs+링크 해석+루트 접두 비교를 세 지점에 적용, 행 4개+실제 등록부 회귀, 뮤턴트 둘 | REQ-CAA-021, AC-024/025, M-20…M-24 | 세 지점 적용은 요구사항에 있으나 **링크 해석은 로그 지점에서만 검증**(D3). 판결의 "심볼릭 링크 해석을 셋 모두에" 가 AC 층에서 반쪽이다 |
| §12.5 수용 확장 2건 | AC-024 CLI 사례, M-22 두 변형·M-21 통합 | 반영됨. 단 CLI 사례는 판별력이 없다(D2) |

design.md 가 "판정서 문구가 아니라 SPEC 인코딩에서 왔다"고 표시한 세 가지:

1. **경로 구분자 경계** — 판결 §11 두 번째 뮤턴트가 `/root-evil` 이 `/root` 접두를 통과하는 모양을 결함으로 명시하므로, 구분자 경계는 판결이 함축한 규칙이다. 일치.
2. **양쪽 해석 + 가장 가까운 존재 상위 경로** — 판결은 `filepath.Abs`·심볼릭 링크 해석을 요구한다. 링크 해석 함수는 존재하지 않는 경로에서 실패하므로, 아직 없는 로그를 판정하려면 가장 가까운 존재 상위 경로 규칙이 필요하다. 루트도 해석하지 않으면 macOS 임시 디렉터리처럼 링크로 닿는 루트가 거부된다. 판결의 의도를 실행 가능하게 만든 세부로, 판결과 모순되지 않는다.
3. **새 clause 0회 검사를 먼저** — G2 는 순서를 정하지 않는다. 둘 다 실패하는 픽스처는 어느 AC 에도 없어 관측 가능한 차이가 없다(AC-019(b)는 현재 clause 1회·새 clause 1회라 순서와 무관). 요구사항에 들어가지 않은 plan 수준 선택이며 모순 없음.

## 두 칸 채택 판정 (verification-completeness §2, §2.1)

- **RED 셀의 성격.** acceptance.md §D 표 머리(L20)는 "RED at `ff11e752f` (predicted from code reading unless cited)" 로 예측임을 밝힌다. 인용된 관측은 AC-006·008·009·010·011 이 기대는 판정서 §2.1–§2.3 뿐이다. 그런데 AC 상세 본문은 예측 표시 없이 단정형으로 적힌 곳이 섞여 있다(AC-020 L268 "must FAIL", AC-022 L292, AC-023 L310, AC-024 L336 탈출 행). AC-024 CLI 사례만 "predicted RED from code reading" 이라고 명시한다. 따라서 **어느 AC 도 plan 시점에 RED-now 가 관측되지 않았다.** 채택은 plan.md §F 의 baseline-first 커밋(RED 출력을 생산 변경 앞 커밋에 기록)으로 완결되도록 설계돼 있고, 이는 §2.3 순서 증인 규율과 맞는다. §2.1 의 네 요소(명령·stdout·exit·SHA)는 어느 AC 에도 없으나, SPEC 이 release-blocking 으로 분류한 AC 가 없으므로 §2.1 의무 위반은 아니다. 표시 혼재는 D12.
- **green 경로 셀.** 표에는 없지만 plan.md §F 각 마일스톤 Exit 줄이 AC 를 이름으로 넘긴다(M1→006…011·018, M2→022·023 일부·024 탈출 행, M3→001…005·019 함수 수준, M4→012·013·021, M5→001·002·020·016·023·024·025, M6→014·015, M7→017). 모든 AC 에 뒤집는 마일스톤이 있다.
- **AC 별 RED 이유 판독(코드 판독, 트리 `76144d40a`; research §A 가 `5a066994b`→`699bedd7c` 코드 무변경을 기록했고, 이번에 `git diff --stat 5a066994b HEAD -- internal/constitution internal/cli/constitution.go` 출력 없음으로 다시 확인):**

| AC | 표의 RED 주장 | 판독 결과 |
|---|---|---|
| 001, 002, 004, 019 | 스텁 오류 | 옳은 이유. `pipeline.go:259`·`:266` 스텁 |
| 003 | 스텁 오류 | 옳은 이유. M-2 는 정규화 시 개수 2 로 여전히 실패하므로, AC 가 **개수 0** 을 단언해야만 죽는다 — L105 가 그렇게 적었다 |
| 005 | 스텁 오류 | (a)(c)는 옳음. **(b)는 D6** |
| 006, 008, 009, 010, 011, 018 | 인용 관측 | 옳은 이유 |
| **007** | "none expected" | **틀림(D5).** 두 번째 단언(`rule_id: A` + `ruleid: B` → `A`)은 태그 없는 디코더가 `rule_id` 를 무시하고 `ruleid: B` 를 읽으므로 오늘 빨갛다 |
| 012, 013, 021 | 컴파일 실패, 뮤턴트가 RED | 정직한 표시. 다만 M-5b 는 D1 |
| 014 | dry-run 이 두 번 출현 픽스처에서 성공 | 옳은 이유(`pipeline.go:133-137`) |
| 015 | CLI 가 성공 출력 | 옳은 이유(`internal/cli/constitution.go` dry-run 분기가 `Dry-run success` 출력) |
| 016 | grep 적중 | 옳은 이유, 대조군 있음 |
| 017 | 불변 가드, M-14 가 RED | 받아들일 만함 |
| 020 | dry-run 성공 | 옳은 이유. real 은 스텁 오류에 규칙 ID 가 없고 게이트 대역이 호출됨 |
| 022, 023 | `Execute` 가 환경 변수 무시 | 옳은 이유(`pipeline.go:66`) |
| 024 | 탈출 행 무검사 성공/스텁 | 옳은 이유(`loader.go:82` 절대 경로만, `pipeline.go:192-195` 무검사 join). CLI 사례는 RED 는 옳지만 green 이 판별력 없음(D2) |
| 025 | load·dry_run 오늘 초록, real 스텁 | 옳다. load 는 등록부 경로를 `B/link` 경유로 넘길 때만 오늘 초록이다(`B/root` 형태면 오늘 `loader.go:84-86` 이 거부) — 경로 형태 미지정은 D16 |

불가능 방향(영원히 빨강)으로 판정한 AC 는 없다. 다만 DoD(L398)의 "29개 뮤턴트 전부 RED 관측"은 D1 때문에 **충족 불가능**하다.

## 뮤턴트 탐침 (containment check · 세 파일 원자적 적용)

작성 가능했던 뮤턴트:

1. **M-5b 두 변형 모두 생존 (D1).** AC-012 `third_rename_log_absent` 의 주입기는 "call N 에서 실패하고 그 외에는 `os.Rename` 에 위임"(L185)한다. 3번째 rename(로그)이 위임 없이 실패하면 로그는 애초에 생기지 않으므로, "부재로 기록된 로그를 지우지 않는" 복원도 "로그 경로 없음" 을 통과한다. "이미 rename 된 파일만 복원" 변형은 rename 실패에 대해서는 **올바른 동작**이라 어떤 rename 실패 사례로도 죽지 않는다. 세 번째 rename 이 마지막이므로, REQ-CAA-010 의 "부재 파일 제거" 는 rename 이 실제로 적용된 뒤 실패를 보고하는 경우에만 도달한다.
2. **CLI 는 검사 없이 적재하고 `Execute` 에만 기대는 뮤턴트 (D2).** AC-024 CLI 사례 픽스처는 `B/other` 에 `P` 등록부의 **바이트 사본**을 둔다. 검사 없는 CLI 는 사본을 읽고, 규칙을 찾고, `--before` 가 일치하고, `Execute` 를 부른다. 그러면 `Execute` 의 검사가 거부해 `amendment failed: … <offending path>` 를 돌려주고, success 줄은 출력되지 않으며, `B` 는 바뀌지 않는다. AC 의 세 단언을 모두 통과하지만, REQ-CAA-021 "the CLI's registry validation and `Execute` shall use that same check" 는 어긴다.
3. **링크 해석을 로그 지점에만 두는 뮤턴트 (D3).** 등록부 경로와 `file:` 지점은 Clean+Abs+구분자 경계만 쓰고, 로그 지점만 링크를 해석한다. AC-024 의 링크 행은 `symlinked_log`(로그 지점)와 `in_root_control/symlinked_root`(루트 링크, 링크 해석 없이도 양쪽이 같은 문자열로 시작해 통과) 뿐이므로 모두 초록이다. AC-025 도 루트 링크뿐이다. 그러나 `file: linked/rule.md` 에서 `P/linked → B/outside` 이면 검사가 안쪽으로 판정하고 **루트 밖 파일에 쓴다** — 판결 §11 이 막으려던 경로 순회 그 자체다.

탐침했으나 죽는 뮤턴트(판별력 확인): 검사를 target 항목에만 적용(`absolute_file/non_target` 가 죽임), Layer 1 뒤로 검사 이동("no gate double was called"), dry-run 에서 로그 검사 생략(`symlinked_log` dry-run), 복원을 전진 rename 경로로 우회(AC-012 호출 횟수 = N), 복원 실패 시 백업 일부 삭제(AC-021 경로별 sha256), 검사 전면 제거(M-20/M-21), 구분자 없는 접두 비교(M-22(i)), Clean 없는 이어 붙이기(M-22(ii)), 후보만 해석(M-24).

## 산출물 일관성

- 계수: spec.md HISTORY L24, progress.md §E.1 L12, 판정서 §14.1 모두 21 · 25 · 29, 이번 계수와 일치. 버전 `0.1.4` 는 spec frontmatter, progress L8·L9, design/research 머리말에서 일치.
- 부속 산출물 `^status:` 줄 수: plan 0 · acceptance 0 · design 0 · research 0 · progress 0 (이번 실행).
- design.md 는 새 요구사항·결정·범위를 들이지 않는다. §B 의 검사 순서는 REQ-CAA-012/017/019/020/021 의 재서술이고, 규칙 밖 세부 세 가지는 출처를 표시했다. §K 가 run 단계 결정을 명시한다.
- research.md 는 머리말 L5–L9 에서 Executed / Code reading / This revision 을 구분하고, 각 절 제목에 표시했다(§B·§C·§D·§E·§F·§I executed, §G·§H code reading, §J 혼합 표시). §D 의 "Two consequences … neither executed" 처럼 판독 결과를 관측과 분리한 서술도 있다. 이 축은 통과다.
- 경미한 불일치: spec.md L30 과 plan.md L3 의 작성 트리 목록은 0.1.3(`578afca87`)에서 끝나고, 0.1.4 작성 트리 `699bedd7c`(research §A)는 없다(D15). `resolveRegistryPath` 줄 범위는 spec L41 `144-155`, design/research `144-154` — 실제 함수는 144–154(이번 `grep -n` 확인). 사소함.

## 제외 항목과 Gap

- §F 제외 항목은 H3 주제별로 구체적이고, 판결 Q5·§8.1·G6(i)·G7(B)과 대응한다. 숨긴 필수 작업은 보이지 않는다.
- 선언된 Gap:
  - 비-dry-run CLI 경로(§E.4, G5 승인). 적용 자체는 `Execute` 수준 AC 가 덮고, 남는 것은 CLI 래퍼의 출력·종료 코드뿐이라는 서술이 정확하다.
  - Windows 링크 skip(AC-024 L334, AC-025 L353). skip 은 PASS 가 아니라 Gap 으로 기록한다고 명시했다. 다만 D3 가 고쳐지지 않으면 링크 검증은 어느 플랫폼에서도 로그 지점에만 존재한다.
  - lint 바이너리 `-dirty`(research §M). 판정서 §14.2 가 기록했다.
- **선언되지 않은 Gap:** 공유 로더 변경의 호출자 영향(D4). 이것은 숨은 작업이다.

## Defects Found

D1. TST-MUTANT-UNKILLABLE — acceptance.md:L185–L191, L367 (AC-CAA-012 `third_rename_log_absent`, M-5b) — 3번째 rename 이 위임 없이 실패하면 로그가 생기지 않아, "부재 기록 파일 미삭제" 복원도 초록이다. "rename 된 파일만 복원" 변형은 rename 실패에 대해 올바른 동작이라 어떤 사례로도 죽지 않는다. DoD L398 "29개 전부 RED" 가 충족 불가능해진다. — Severity: major — Class: blocking — Required fix: `third_rename_log_absent` 의 주입기를 "call 3 에서 `os.Rename` 에 **위임한 뒤** 오류를 반환"(적용됐지만 실패로 보고)으로 바꿔 로그가 실제로 생긴 뒤 복원이 제거해야만 통과하게 한다. M-5b 를 "부재로 기록된 파일을 복원 중 제거하지 않음" 한 변형으로 다시 쓰고, "rename 된 파일만 복원" 변형은 삭제하거나 위임 후 실패 사례에서 죽는 형태로 재정의한다. 새 AC 번호 없이 AC-012 하위 사례로 처리한다.

D2. TST-CLI-CASE-NONDISCRIMINATING — acceptance.md:L318, L330 (AC-CAA-024 CLI 사례, REQ-CAA-021 "CLI 와 Execute 가 같은 검사") — `B/other` 가 `P` 등록부의 바이트 사본이라, 검사 없이 적재하는 CLI 도 `Execute` 의 거부로 같은 오류·같은 출력·같은 스냅숏을 낸다. 요구사항을 어기는 뮤턴트가 통과한다. — Severity: major — Class: blocking — Required fix: CLI 사례 픽스처의 `B/other` 사본에서 대상 항목 clause 를 다르게 두어, 검사 없는 CLI 는 `clause mismatch` 경로(오류에 탈출 경로가 없음)로 빠지게 한다. 동시에 "오류가 `amendment failed` 로 감싸지지 않았다"(파이프라인에 도달하지 않음)를 단언한다. 이 뮤턴트("CLI 적재 지점 검사 제거")를 M-20 의 CLI 변형으로 §D.2 에 적는다.

D3. TST-SYMLINK-SITE-COVERAGE — acceptance.md:L316–L323, L388 (AC-CAA-024 행 표, M-23) · spec.md:L95 (REQ-CAA-021) · verdict.md §11 — 링크 해석을 검증하는 탈출 행이 로그 지점(`symlinked_log`) 하나뿐이다. 등록부 경로·`file:` 지점에서 링크를 해석하지 않는 구현이 모든 AC 를 통과하면서 링크 디렉터리를 통한 루트 밖 쓰기를 허용한다. — Severity: major — Class: blocking — Required fix: AC-024 에 `symlinked_file` 행(`P/linked` → `B/outside`, 대상 `file: linked/rule.md`, 해당 파일에 현재 clause 1회·새 clause 0회)을 추가하고, 가능하면 `symlinked_registry` 행(`MOAI_CONSTITUTION_REGISTRY` 가 `P/linkdir/zone-registry.md`, `P/linkdir` → `B/other`)도 추가한다. M-23 을 지점별 변형(등록부 경로·`file:`·로그)으로 나눈다. 새 AC 번호 없이 AC-024 행으로 넣어 AC 25 상한을 유지한다.

D4. PLAN-SHARED-LOADER-BLAST-RADIUS — spec.md:L155 (§E.2 "Its containment check … is replaced by the one check of REQ-CAA-021"), plan.md:L65 (§D 검증 범위), L139 (R-8) — `LoadRegistry` 의 비-테스트 호출자는 `internal/cli/constitution.go:70`·`:160`·`:511`, `internal/cli/doctor.go:683`, `internal/constitution/validator.go:183`, `internal/spec/lint.go:114`, `internal/constitution/pipeline.go:67` 이다. R-8 은 `list`·`guard` 만 들고, 검증 범위는 `internal/constitution` 과 CLI 선택자로 한정해 `internal/spec`·doctor 를 뺀다. 코드 판독상 `internal/spec/lint_test.go:18-25` 의 `testRegistryPath()` 는 상대 경로 `../../.claude/rules/moai/core/zone-registry.md` 를 쓰고, `TestLinter_AC08_DanglingRuleReference`(`lint_test.go:218-240`)는 `BaseDir: testdataDir`(`"testdata"`)를 넘긴다. 이 등록부는 `internal/spec/testdata` 밖이므로 REQ-CAA-021 검사가 거부하고, `NewLinter` 는 오류를 삼켜(`lint.go:114-118`) 레지스트리 없이 진행한다. 그러면 `DanglingRuleReference` 단언이 빨개질 것으로 **예측**된다(실행하지 않음). 기존 동작에서 이 상대 경로는 절대 경로 검사를 건너뛰어 통과한다(`loader.go:82`). — Severity: major — Class: blocking — Required fix: (1) spec §E.2 또는 plan §G 에 `LoadRegistry` 호출자 전체를 열거하고, 호출자별로 REQ-CAA-021 거부가 의도인지 적는다. (2) `internal/spec` 과 `internal/cli` doctor 관련 테스트를 plan §D 검증 범위와 M2 Exit 에 넣는다. (3) `internal/spec` lint 테스트의 등록부/`BaseDir` 조합을 어떻게 다룰지(픽스처 변경인지, 검사를 amend 경로에만 둘지)는 판결 §11 "등록부 적재 오류" 의 적용 범위 문제이므로, SPEC 이 스스로 정하지 말고 리드 판단으로 올린다. 이 결정은 AC 번호를 늘리지 않고 AC-025 또는 기존 회귀 가드로 담을 수 있다.

D5. TST-RED-CELL-WRONG — acceptance.md:L28, L141 (AC-CAA-007) — 표는 "none expected — invariant guard" 로 적었으나, "`rule_id: A` 와 `ruleid: B` 를 함께 가진 항목이 `A` 를 읽는다" 단언은 태그 없는 디코더(`amendment.go:192-219`)가 `ruleid` 만 읽으므로 오늘 빨갛다(코드 판독). baseline-first 로 지정되지 않아 run 단계 RED 커밋에서 빠진다. — Severity: minor — Class: blocking — Required fix: AC-007 행을 "두 형식 공존 단언은 오늘 RED(`RuleID = B`) — baseline-first, 구 키 단독 단언은 오늘 초록이며 RED 셀은 M-6" 으로 고친다.

D6. TST-WRONG-REASON-SUBCASE — acceptance.md:L120–L122 (AC-CAA-005(b)) · spec.md:L85 (REQ-CAA-004 "no single `clause:` line") — `clause:` 줄이 없는 항목은 로더가 `Clause=""` 로 디코드한다. 그러면 REQ-CAA-017(Before≠"")이나 REQ-CAA-002(빈 문자열 개수 ≠ 1), `AmendmentLog.Validate` 가 먼저 거부한다. "any error" 단언은 REQ-CAA-004 의 해당 조항에 닿지 않고도 통과한다. — Severity: minor — Class: blocking — Required fix: (b) 픽스처를 디코드는 성공해 clause 값이 존재하지만 대상 항목에 단일 `clause:` 줄이 없는 형태(예: flow mapping `- {id: …, clause: "…", …}`)로 바꾸고, 오류 문구가 등록부 재작성 단계를 가리키는지 단언한다. 불가능하면 (b)를 삭제하고 REQ-CAA-004 의 해당 절을 REQ 에서 뺀다.

D7. TRC-UNVERIFIED-RESTORE-STEPS — spec.md:L127 (REQ-CAA-010 "a backup, a temporary write, or any of the three renames") · spec.md:L129 (REQ-CAA-011 seam 은 rename 과 복원만) — 백업·임시 쓰기 실패와 1번째 rename 실패 시 복원·정리는 주입 수단도 AC 도 없다. 판결 Q4 는 2·3번째 rename 주입만 요구했다. — Severity: minor — Class: optional — Required fix: REQ-CAA-010 문장을 검증 대상 단계로 좁히거나, 1번째 rename 실패(비용이 작음)를 AC-012 하위 사례로 추가하고 나머지는 §E.4 Gap 으로 적는다.

D8. TRC-AC-STRICTER-THAN-REQ — acceptance.md:L277 (AC-CAA-021 "no temporary file remains") · spec.md:L131 (REQ-CAA-018) — REQ-CAA-018 은 복원 실패 시 임시 파일 제거를 요구하지 않고, REQ-CAA-010 은 "completed restore" 에만 요구한다. AC 가 요구사항에 없는 성질을 단언한다(design §H.3 에만 있음). — Severity: minor — Class: optional — Required fix: REQ-CAA-018 에 "shall remove every temporary file it created" 를 넣거나 AC 의 해당 단언을 뺀다.

D9. CLR-LOCK-WRITE — spec.md:L95 (REQ-CAA-021 "before any backup, temporary file, or write") vs L87·L91 ("a lock it acquired shall be released"), `pipeline.go:58`(적재 전 락 쓰기) — 실제 모드에서 락 파일은 검사 전에 쓰인다. 문자 그대로 읽으면 요구사항끼리 충돌한다. — Severity: minor — Class: optional — Required fix: "before any backup, temporary file, or write to the three files (the lock file excepted)" 로 명시한다.

D10. CLR-NONBINARY-PHRASE — acceptance.md:L208 (AC-CAA-014 "an error of the same kind a real apply returns") — "same kind" 는 판정자 해석이 필요하다. — Severity: minor — Class: optional — Required fix: 사례 (b)–(f) 별로 단언할 오류 부분 문자열(경로·개수·규칙 ID 등)을 적거나, 같은 픽스처에 대한 real 모드 오류 문자열과의 동등성으로 정의한다.

D11. TST-DUPLICATE-COUNT-PIN — acceptance.md:L342–L344, L351 (AC-CAA-025 "101 `file:` lines", "101 entries") — `registry_sync_test.go:49` `wantRegistryEntries = 101` 가 이미 같은 수를 고정하고 "의도적 증가는 같은 변경에서 갱신" 규율을 둔다. 두 번째 리터럴 핀은 등록부 증가 때 이 작업과 무관한 이유로 빨개진다. 실제 등록부의 특정 live 항목을 대상으로 고르는 것도 같은 방식으로 깨지기 쉽다. — Severity: minor — Class: optional — Required fix: 기존 상수를 재사용하거나 복사본에서 센 값과 비교하고, 절대 경로 0·`..` 0 만 리터럴로 단언한다. 대상 항목은 ID 고정 대신 "첫 번째 live Evolvable 항목" 같은 규칙으로 고른다.

D12. VC-RED-LABEL-MIXED — acceptance.md:L20 vs L268, L292, L310, L336 — 표 머리는 예측이라 밝히지만 상세 본문 여러 곳이 표시 없이 "must FAIL"·"RED" 로 단정한다. 관측된 RED-now 는 없다. — Severity: minor — Class: optional — Required fix: 각 AC 의 Baseline-first 문단 첫머리에 "Predicted from code reading at `<SHA>`" 를 붙이고, 관측으로 바뀌는 시점은 run 단계 baseline 커밋임을 적는다.

D13. TRC-ENV-DRYRUN-UNVERIFIED — spec.md:L137 (REQ-CAA-013 "or `MOAI_CONSTITUTION_DRY_RUN=true`") · acceptance.md:L216 — AC-015 는 `runConstitutionAmend(…, true)` 를 직접 불러, 환경 변수로 dry-run 이 켜지는 경로(`newConstitutionAmendCmd` 에서 읽음)는 확인하지 않는다. — Severity: minor — Class: optional — Required fix: REQ-CAA-013 에서 환경 변수 괄호를 빼거나, 명령 구성 수준 사례를 AC-015 하위 사례로 추가한다.

D14. RQ-GEARS-LABEL — spec.md:L77 등 "(event-detected)" 7곳, L107 REQ-CAA-008 "(capability gate)" — "event-detected" 는 다섯 GEARS 패턴 이름이 아니다(본문은 event-driven 형태라 MP-2 는 통과). REQ-CAA-008 의 `Where` 조건은 기능 플래그·정적 설정이 아니라 데이터 내용이다. 또 §E.1(L149)은 함수 이름을 run 단계 결정이라 하지만, REQ-CAA-014 는 테스트 함수 이름 다섯 개를, REQ-CAA-017 은 `Execute` 를 지명한다. — Severity: minor — Class: optional — Required fix: 라벨을 `event-driven` 으로 통일하고, REQ-CAA-008 을 "When the reader encounters a fenced yaml block …" 로 바꾼다. §E.1 에 "기존 공개 식별자와 교체 대상 테스트 이름은 예외" 를 적는다.

D15. CN-AUTHOR-TREE-OMITTED — spec.md:L30, plan.md:L3 — 작성 트리 목록이 0.1.3 에서 끝나고 0.1.4 작성 트리 `699bedd7c`(research.md §A L23)가 없다. — Severity: minor — Class: optional — Required fix: 두 머리말에 "0.1.4 authored on `699bedd7c` (verdict §13 added, no code change)" 를 덧붙인다.

D16. CLR-ENTRYPOINT-UNSPECIFIED — acceptance.md:L104 (AC-CAA-003 "When the apply step validates"), L343 (AC-CAA-025 `load` 의 등록부 경로 형태) — AC-003 은 `Execute` 인지 함수 수준인지 정하지 않는다. AC-025 `load` 는 `B/link/…` 경유인지 `B/root/…` 인지 정하지 않는데, 오늘 코드에서 `B/root` 형태는 `loader.go:84-86` 이 거부하므로 RED 예측이 형태에 따라 뒤집힌다. — Severity: minor — Class: optional — Required fix: AC-003 은 `Execute(dryRun=false)` 로, AC-025 `load` 는 `LoadRegistry("B/link/.claude/rules/moai/core/zone-registry.md", "B/link")` 로 명시한다.

## Recommendation

FAIL. manager-spec 은 새 요구사항·AC 번호를 만들지 않고 아래 순서로 고친다(AC 25 상한 유지).

1. D1 — AC-012 `third_rename_log_absent` 주입기를 "위임 후 실패" 로 바꾸고 M-5b 를 다시 정의한다(acceptance.md L185–L191, L367).
2. D2 — AC-024 CLI 사례의 `B/other` 사본 clause 를 다르게 두고 "파이프라인 미도달" 을 단언한다. M-20 에 CLI 변형을 추가한다(L318, L330, L385).
3. D3 — AC-024 에 `symlinked_file` 행(가능하면 `symlinked_registry` 행)을 추가하고 M-23 을 지점별로 나눈다(L316–L323, L388).
4. D4 — `LoadRegistry` 호출자 영향을 spec §E.2·plan §G 에 열거하고, `internal/spec`·doctor 테스트를 plan §D·M2 Exit 검증 범위에 넣는다. `internal/spec` lint 테스트의 상대 등록부 경로를 거부할지는 **리드에게 판단을 요청**한다(판결 §11 적용 범위 문제). 이 항목은 SPEC 작성자가 혼자 결정하지 않는다.
5. D5·D6 — AC-007 RED 셀을 고치고, AC-005(b) 픽스처를 REQ-CAA-004 조항에 실제로 닿는 형태로 바꾸거나 해당 조항을 정리한다.
6. optional(D7–D16)은 오케스트레이터 재량이다. D9·D10·D16 은 한 줄 수정이라 같은 개정에 넣을 만하다.

재감사(iteration 2)는 위 결함 목록에 한정한 델타 감사와 D1–D16 회귀 확인으로 진행한다.
