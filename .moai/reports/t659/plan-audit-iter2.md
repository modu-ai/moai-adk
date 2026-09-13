# SPEC 감사 보고서: SPEC-CON-AMEND-APPLY-001

Iteration: 2/3
Verdict: FAIL
Overall Score: 0.84 (Tier L 합격선 0.85 미달, 1회차 0.78 대비 +0.06)

- 감사 대상 트리: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t659`, 브랜치 `WT-amend-apply`, HEAD `0085766cdfda3e071cce1a4788d2014ab83d8c5d` (`git rev-parse --show-toplevel` / `git branch --show-current` / `git rev-parse HEAD` 로 확인, 지시와 일치). SPEC 파일의 마지막 변경은 `564c370b5`(0.1.6)이고, 그 뒤 `0085766cd` 는 판정서와 lint 증거 두 파일만 바꿨다(`git diff --stat 564c370b5 HEAD` → 3 files, 모두 `.moai/reports/t659/`).
- 대상 개정: `spec.md` `version: "0.1.6"`, `tier: L`, `status: draft`.
- 읽은 것: `plan-audit-iter1.md`, SPEC 여섯 파일 전부, `git diff 76144d40a HEAD` 의 SPEC 폴더 부분, 판정서 §7–§16, 코드 `internal/constitution/{pipeline,loader,rule}.go`, `internal/cli/constitution.go`, `internal/spec/lint.go`, `internal/spec/lint_test.go`, `internal/constitution/registry_sync_test.go` 머리, 규칙 `spec-workflow.md` § SPEC Complexity Tier, `verification-completeness.md` §1–§2.
- 위임문에 담긴 판정서 요약은 범위 지정으로만 썼다. 작성자 추론 맥락은 받지 않았다(Reasoning context ignored per M1 Context Isolation).
- 교차 모델 감사 MCP 도구는 이 세션의 도구 목록에 없어 부르지 않았다. 판정은 이 감사자 단독이다.
- 실행하지 않은 것: `go test`, `go build`, `moai todo`. 아래 코드 동작 주장과 뮤턴트 생존·사망 판단은 모두 **코드 판독**이며 실행 관측이 아니다.
- 생산 코드 무변경 재확인: `git diff --stat 5a066994b HEAD -- internal/constitution internal/cli/constitution.go internal/spec internal/cli/doctor.go` → 출력 없음, exit 0. 대조 `git diff --stat 5a066994b HEAD` → `22 files changed, 2209 insertions(+)`.

## 판정 요약

1회차 결함 16건은 모두 해결됐다. 운영자 결정 D4 도 여섯 파일에 일관되게 들어갔고, 보존 행 두 개(`loader_unchanged`, `internal/spec` 보존 실행)는 선언된 뮤턴트 M-20 (iii)이 실제로 죽인다(코드 판독). 필수 항목 7개도 모두 통과한다.

그런데 이번 회차의 뮤턴트 전수 점검에서 1회차가 놓친 결함이 새로 나왔다. 가장 무거운 것(N1)은 오류 문자열에 숫자가 들어 있는지를 보는 단언들이다. 오류에 함께 찍히는 `t.TempDir()` 경로가 이미 그 숫자를 품고 있어서, 문구 그대로 구현하면 M-2 는 AC-CAA-003 에서 확실히 살아남는다. M-15 의 두 변형도 살아남을 수 있다. 여기에 macOS 임시 디렉터리 링크 때문에 M-23 의 선언 킬 행이 darwin 에서 판별력을 잃는 문제(N2)와, 등록부 경로 검사 순서가 REQ-CAA-020 의 "읽기도 닿지 않는다"와 어긋나는 문제(N3)가 있다. 점수는 0.84 로 합격선에 0.01 모자라고, blocking 결함 셋(major 1, minor 2)이 남는다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `grep -oE '^- \*\*REQ-CAA-[0-9]+' spec.md` → 21줄, 각 ID 1회. 여섯 파일 전체 구별 ID `REQ-CAA-*` 21개. 빈 번호·중복 없음.
- [PASS] MP-2 GEARS 형식 (요구사항 층에 대해 판정): 21개 REQ 라벨이 `event-driven` 13 · `ubiquitous` 5 · `unwanted` 2 · `state-driven` 1 이고, 본문이 라벨에 맞는 `When`/`While`/보편형/`shall not` 형태다(예: spec.md:L110 REQ-CAA-008 "**When** the reader encounters …", L138 REQ-CAA-012 "**While** `Execute` runs in dry-run mode"). `event-detected` 0회. REQ-CAA-020(L94)·REQ-CAA-021(L98)은 보편형 문장과 `When` 문장을 한 항목에 묶은 복합형이며 GEARS 복합 절로 인정한다. AC 층의 Given-When-Then 은 이 항목에서 채점하지 않았다.
- [PASS] MP-3 YAML frontmatter: spec.md:L2–L13 에 12개 필수 필드가 모두 있다 — `version: "0.1.6"`(인용), `status: draft`, `created`/`updated` `2026-09-11`, `priority: P1`, `lifecycle: spec-anchored`, `tags` 쉼표 문자열. 거부 별칭 없음. `tier: L`, `related_specs` 는 선택 필드.
- [N/A] MP-4 언어 중립성: 템플릿이 아닌 Go 내부 패키지(`internal/constitution`, `internal/cli`) 변경이다.
- [PASS] MP-5 D7 교차 SPEC: 본문 참조는 `SPEC-V3R2-CON-002`(자기 ID 제외) 하나, `status: implemented`. retired/superseded/archived 아님. BLOCKING 없음.
- [PASS] MP-6 D8 교차 플랫폼: `grep -c syscall spec.md` → 0.
- [PASS] MP-7 확인 요청 표식: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → 출력 없음, exit 1.

## Tier 와 예산

| 항목 | 계수(이번 실행) | Tier L 상한 | 판정 |
|---|---|---|---|
| 요구사항 | 21 (정의 줄 21, 여섯 파일 구별 ID 21) | 25 | 범위 안 |
| 인수 조건 | 25 (`^### AC-CAA-` 25, 여섯 파일 구별 ID 25, 최댓값 AC-CAA-025) | 25 | **상한과 같음**, 여유 0 |
| 뮤턴트 | 29 (`^\| M-` 행 29, 중복 ID 0) | 상한 없음 | — |

두 상한은 따로 적용되며 둘 다 넘지 않는다. 합격선은 `spec-workflow.md` 표의 Tier L 값 0.85 다. 아래 권고는 모두 기존 AC 안의 문구·픽스처 수정으로 처리할 수 있어 AC 번호를 늘리지 않는다.

## Category Scores

| 차원 | 점수 | 기준 구간 | 근거 |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 사이 | D9·D10·D16 이 풀렸다(spec.md:L98 락 예외, acceptance.md:L224 사례별 부분 문자열, L109·L364 진입점). 남은 모호성: 숫자 단언이 오류 안 경로 숫자와 구별되지 않음(N1), 등록부 경로 검사를 적재 전에 하는지 후에 하는지(N3, plan.md:L97 "as soon as the registry is loaded"), CLI 사례의 경로 형태 미지정(N2). |
| Completeness | 0.90 | 0.75–1.0 사이 | 필수 섹션과 Tier L 산출물 5+1개가 모두 있다. D4 영향 범위가 호출자 일곱 곳까지 열거됐다(spec.md:L158, plan.md:L20, R-8 L141). 백업·임시 쓰기 실패 주입은 §E.4 Gap 으로 선언됐으나 리드 승인이 아직 없다(spec.md:L170). |
| Testability | 0.75 | 0.75 | D1–D3 이 해결돼 뮤턴트 킬 경로가 대부분 판별력을 갖는다. 그러나 여러 AC 의 숫자 단언이 문구 그대로는 경로 숫자로 통과해 M-2·M-15 가 살아남을 수 있고(N1), M-23 의 선언 킬 행이 darwin 에서는 판별하지 못한다(N2). |
| Traceability | 0.85 | 0.75–1.0 사이 | 21 REQ 모두 AC 가 있고(acceptance.md §D.0 L55–L79), AC 가 가리키는 REQ 는 모두 존재한다. REQ-CAA-020 의 "no read … shall reach another directory tree"(spec.md:L94)는 어느 AC 도 관측하지 않고, REQ-CAA-010 의 백업·임시 쓰기 실패 복원은 선언된 Gap 이다. |

산술 평균 0.8375 → 0.84. 합격선 0.85 미달. 1회차 0.78 대비 +0.06(명확성 +0.10, 완결성 +0.05, 검증 가능성 +0.05, 추적성 +0.05). 점수가 올랐으므로 STOP 신호는 내지 않는다.

## Regression Check — 1회차 D1–D16

| ID | 판정 | 근거 |
|---|---|---|
| D1 M-5b 사멸 불가 | RESOLVED | acceptance.md:L192 주입기에 rename-then-fail 모드, L195 호출 횟수와 "목적지가 실제로 생겼음" 기록, L198–L205 표에 `first_rename_applied`·`third_rename_applied`·`third_rename_log_absent`(모두 rename-then-fail), L210 킬 맵, L389 M-5b 두 변형. 판독: 변형 (i)(성공 보고된 rename 만 복원)은 세 행에서 이미 교체된 파일을 되돌리지 않아 빨갛고, (ii)(부재 기록 파일 미삭제)는 `third_rename_log_absent` 에서 로그가 남아 경로 집합 단언이 빨갛다. |
| D2 CLI 사례 판별력 없음 | RESOLVED | acceptance.md:L351 — `B/other` 사본의 대상 clause 를 다르게 두고, 오류에 `amendment failed`(`internal/cli/constitution.go:544`)도 `clause mismatch`(`:529`)도 없음을 단언. L407 M-20 (ii)(CLI 만 검사 생략). 판독: 검사 없는 CLI 는 `LoadRegistry` 가 상대 경로를 검사하지 않으므로(`loader.go:82`) 사본을 읽고 `:529` 에서 멈춰 빨갛다. 두 줄 번호를 이번에 `grep -n` 으로 확인했다. |
| D3 링크 해석 지점 누락 | RESOLVED (단 N2 참고) | acceptance.md:L340 `symlinked_registry`, L341 `symlinked_file`, L410 M-23 지점별 세 변형. Linux 에서는 각 변형이 해당 행으로 죽는다(판독). darwin 에서는 선언 행이 아닌 `in_root_control` 로 죽는다 — N2. |
| D4 공유 로더 영향 범위 | RESOLVED (운영자 결정) | 아래 "D4 인코딩" 절. |
| D5 AC-007 RED 셀 | RESOLVED | acceptance.md:L33 "both-forms assertion … RED — the untagged decoder ignores `rule_id` and reads `RuleID = B` — baseline-first", L150. 판독: `AmendmentLog` 에 태그가 없어(`amendment.go:192-219`, 1회차 확인) 기본 키 `ruleid` 만 읽힌다. 옳은 이유. |
| D6 AC-005(b) 잘못된 이유 | RESOLVED | acceptance.md:L125 흐름 매핑 `- {id: …, clause: "…"}` — 로더는 clause 를 디코드하고, 규칙 파일은 현재 clause 1회·새 clause 0회라 재작성 단계 전 검사를 모두 통과한다. spec.md:L88 REQ-CAA-004 가 "error naming the registry path" 로 바뀌었고 L127 이 등록부 경로를 단언한다. 오늘의 스텁 오류는 원문 경로를 담아(`pipeline.go:259`) 등록부 경로 단언과 구별된다. |
| D7 복원 단계 미검증 | RESOLVED (요구된 수정 이행, 리드 승인 대기) | acceptance.md:L200 `first_rename_applied`, L203 `third_rename_applied` 추가. 백업 쓰기·임시 쓰기 실패는 spec.md:L170 에 "not yet approved" Gap 으로 선언. 1회차 권고(1번째 rename 사례 추가 + 나머지 Gap)와 일치한다. Gap 승인 여부는 리드 몫이며 이 감사는 Gap 선언 자체를 결함으로 보지 않는다. |
| D8 AC 가 REQ 보다 엄격 | RESOLVED | spec.md:L134 REQ-CAA-018 "shall remove every temporary file it created". acceptance.md:L293 과 일치. |
| D9 락 쓰기 충돌 | RESOLVED | spec.md:L98 "the lock file a real-mode run creates before it loads the registry is the one exception". `pipeline.go:58` 락 선취와 일치. |
| D10 "same kind" 비이진 | RESOLVED (단 N1 참고) | acceptance.md:L224 사례 (b)–(f)마다 부분 문자열 지정, real 모드 동일 부분 문자열 요구. 숫자 부분 문자열의 판별력은 N1. |
| D11 중복 개수 핀 | RESOLVED | acceptance.md:L363 개수 핀을 반복하지 않고 `wantRegistryEntries`(`registry_sync_test.go:49`, 패키지 `constitution_test` — 이번에 확인)를 가리킴, 대상 항목은 규칙으로 선택, L365 개수는 사본의 `- id:` 줄 수와 비교. |
| D12 RED 표시 혼재 | RESOLVED | acceptance.md:L21 공통 규칙(판독 예측, `54ca2e3b6`, 관측은 baseline 커밋), L25 표 머리, L150·L160·L284·L308·L328·L357·L377 상세 문단 머리 "predicted from code reading at `54ca2e3b6`", plan.md:L95·L97·L104. 표시 없는 "must FAIL"/"RED" 단정 0건(`grep -nE 'must FAIL|is RED|…'` 전수 확인). |
| D13 환경 변수 dry-run | RESOLVED | spec.md:L140 REQ-CAA-013 이 `--dry-run` 플래그만 묶고 환경 변수 경로를 §F 로 넘김(L206). |
| D14 GEARS 라벨 | RESOLVED | `event-detected` 0회, REQ-CAA-008(L110)·009(L126) `When` 형, spec.md:L152 §E.1 식별자 예외 명시. |
| D15 작성 트리 누락 | RESOLVED | spec.md:L32, plan.md:L3 에 0.1.4 `699bedd7c`, 0.1.5 `54ca2e3b6`, 0.1.6 `193136a6a` 기록. |
| D16 진입점 미지정 | RESOLVED | acceptance.md:L109 AC-003 `Execute(dryRun=false)`, L364 AC-025 `LoadRegistry("B/link/…", "B/link")`. |

1회차 결함 중 미해결로 남은 것은 없다. 따라서 "이전 회차 미해결 = 자동 FAIL" 규칙은 적용되지 않는다. 이번 FAIL 은 새 결함과 점수 때문이다.

## D4 인코딩

운영자 결정: 검사는 amend 경로(파이프라인 적용 단계와 CLI `amend`)에서만 돌고, `LoadRegistry` 와 amend 가 아닌 호출자 다섯 곳은 현재 동작을 유지한다.

| 산출물 | 위치 | 일치 여부 |
|---|---|---|
| spec 요구사항 | L94 REQ-CAA-020 근거 문단, L98 REQ-CAA-021 "`LoadRegistry` shall not perform it and shall keep its present behaviour, including its present refusal of an absolute registry path" | 일치 |
| spec 제약·범위 | L74 amend 경로 정의, L158 §E.2 호출자 일곱 곳(2 amend + 5 비-amend), L202 §F 새 제외 항목, L219 §G 해결 기록 | 일치 |
| plan | L19–L20 §B, L42 사전 점검, L67 검증 범위에 `internal/spec` 선택자 추가, L97 M2 "do not move the check into the loader", L98 Exit, L141 R-8 (a)(b) | 일치 |
| design | L31 §B, §C.2 표, §C.3 "amend path only (D4)", §C.4, §C.5 기각안 "Put the check inside `LoadRegistry`" | 일치 |
| acceptance | L310 AC-022 관계 문단, L316·L326 AC-023, L350 `loader_unchanged`, L369 보존 실행, L407 M-20 (iii), L413 | 일치 |

호출자 일곱 곳을 이번에 다시 쟀다: `grep -rn --include='*.go' 'LoadRegistry(' internal cmd pkg | grep -v _test.go` → 8줄(정의 1 + 호출 7). `internal/spec/lint.go:114`, `internal/cli/constitution.go:70`·`:160`·`:511`, `internal/cli/doctor.go:683`, `internal/constitution/validator.go:183`, `internal/constitution/pipeline.go:67`. 판정서 §16.2 의 "5곳" 정정과 SPEC 서술이 모두 이 측정과 맞는다.

**보존 행의 사멸 가능성(코드 판독).**

- `loader_unchanged` — 오늘 초록: `relative_registry` 는 상대 경로라 `loader.go:82` 검사를 건너뛰고 읽는다. `absolute_file` 은 로더가 `file:` 을 검사하지 않고 존재 여부만 본다(`loader.go:141-149`). `symlinked_registry` 는 링크를 풀지 않은 문자열로 `filepath.Rel` 을 계산해 안쪽으로 판정한다. `absolute_escape` 는 `loader.go:84-86` 이 거부하며 오류에 `escapes project dir` 가 들어 있다(`:86`). M-20 (iii)(검사를 로더 안으로)에서는 앞의 셋이 거부되어 빨갛다. 부분 이동(등록부 경로만, 또는 `file:` 만 로더로)도 셋 중 적어도 하나로 죽는다. 공허하지 않다.
- `internal/spec` 보존 실행 — 오늘 초록: `testRegistryPath()` 가 저장소 등록부를 찾으면 상대 경로를 돌려주고(`lint_test.go:18-25`), 로더는 상대 경로를 검사하지 않는다. 등록부를 못 찾으면 `""` → 레지스트리 nil → `ZoneRegistryRule.Check` 가 바로 nil 을 돌려(`lint.go:1442-1444`) 테스트가 오늘 **빨갛다**. 따라서 빈 결과로 초록이 되는 공허 통과는 없다. M-20 (iii)에서는 등록부 경로가 패키지 디렉터리 기준 저장소 `.claude/…` 로, 루트 `testdata` 가 `internal/spec/testdata` 로 풀려 거부되고, `NewLinter` 가 오류를 삼켜(`lint.go:114-118`) finding 이 사라져 빨갛다. 공허하지 않다.

D4 인코딩은 일관되고, 보존 행 두 개는 선언 뮤턴트로 죽는다. 남는 흠은 N4(보존 실행의 `<plan-phase HEAD>` 미고정)와 N6 의 경미한 서술이다.

## 뮤턴트 전수 점검 (29개, 변형 포함)

| 뮤턴트 | 선언 킬 사례 | 판독 결과 |
|---|---|---|
| M-1 | AC-002 (b) | 죽음 — 두 번 출현을 치환하면 성공이 반환되어 오류 단언이 빨갛다 |
| **M-2** | AC-003 | **문구 그대로면 생존(N1)** — 정규화 개수 2 의 오류도 경로의 `…/001/` 때문에 `0` 을 담는다 |
| M-3a, M-3b | AC-004 | 죽음 — 줄 차이 ≠ 1 / 따옴표·역슬래시 왕복 실패 |
| M-4 | AC-005 (a) | 죽음 — 잘린 연속 줄이 재파싱 없이 rename 된다 |
| M-5a | AC-012 실패 행 5개 | 죽음 |
| M-5b (i)(ii) | AC-012 지정 행 | 죽음(D1 해결) |
| M-5c | AC-013 | 죽음 — 순서 기록 |
| M-6, M-7, M-8a, M-8b, M-9, M-12 | AC-007, 008, 010/011, 010(c)/018, 006, 009 | 죽음 |
| M-10 | AC-014 (b)–(f) | 죽음 — (b)–(e) 가 성공으로 바뀐다. (f) 는 REQ-CAA-017 검사로 막히므로 M-10 이 건드리지 않을 수 있다(N6) |
| M-11a, M-11b | AC-012 `no_fault_clean`, AC-014 스냅숏 | 죽음 |
| M-13 | AC-015 | 죽음 |
| M-14 | AC-017 | 죽음 — 두 번째 실행에서 환경 변수를 따르는 테스트는 실제 등록부로 향하고, 루트 밖이라 거부되어 두 실행 결과가 갈린다 |
| **M-15** | AC-018 | 변형 (ii)(키 누락)는 죽음. **변형 (i)(줄 번호 누락)·(iii)(블록 기준 줄)은 문구 그대로면 경로 숫자에 따라 생존 가능(N1)** |
| M-16, M-17, M-18 | AC-019, 020, 021 | 죽음 |
| M-19 | AC-022; AC-023 발산 행 | 죽음 — `Execute` 가 `P` 등록부를 읽어 dry-run 성공 또는 `P` 에 적용 |
| M-20 (i) | AC-024 `relative_env_escape`·`symlinked_registry`·CLI | 죽음 — 상대 경로는 로더가 검사하지 않고, 링크 경로는 로더의 풀리지 않은 비교가 안쪽으로 본다 |
| M-20 (ii) | AC-024 CLI | 죽음 — `clause mismatch` 로 빠진다 |
| M-20 (iii) | `loader_unchanged` 셋, `internal/spec` 보존 실행 | 죽음(위 절) |
| M-21 `file:`·로그 | AC-024 해당 행 | 죽음 |
| M-22 (i) | `sibling_prefix_file` | 죽음 — 두 쪽 모두 해석된 경로라 플랫폼 무관 |
| M-22 (ii) | `dotdot_file` | "the root" 가 해석 전 루트면 죽음. 해석된 루트면 darwin 에서 선언 행으로는 안 죽고 `in_root_control` 로 죽음(N2) |
| M-23 (i)(ii)(iii) | 각 `symlinked_*` 행 | Linux 에서 죽음. **darwin 에서는 선언 행이 판별하지 못하고 `in_root_control` 로만 죽음(N2)** |
| M-24 | `in_root_control/symlinked_root`, AC-025 `dry_run` | 죽음 — `dry_run` 이관은 옳다. 등록부 경로가 `B/link/…` 로 들어오고(REQ-CAA-019 세 번째 우선순위) 후보만 `B/root/…` 로 풀리면 거부된다 |

1회차 이후 옮겨진 담당도 확인했다. AC-023 발산 행을 M-20 (i)에서 M-19 로 옮긴 것은 옳다: D4 로 남은 로더의 절대 경로 거부(`loader.go:82-86`)가 `Q/…` 를 여전히 막으므로 M-20 (i)로는 그 행이 초록으로 남는다. M-24 의 AC-025 킬을 `load` 에서 `dry_run` 으로 옮긴 것도 옳다: `load` 는 `LoadRegistry` 를 직접 부르고, D4 아래에서 로더에는 해석 검사가 없다.

원리상 사멸 불가능한 뮤턴트는 없다. 다만 M-2 는 AC 문구 그대로 구현하면 선언 사례로 죽지 않는다(N1). M-23 세 변형은 darwin 에서 선언 사례가 판별력을 잃는다(N2). acceptance.md:L415 "Every mutant and every variant names at least one subtest that only a correct implementation passes" 는 이 두 조건에서 성립하지 않는다.

## 두 칸 채택 판정 (verification-completeness §2)

- 모든 RED 서술에 판독 예측 표시가 있다(D12). 관측된 RED-now 는 아직 없으며, plan.md §F 의 baseline 커밋에서 관측으로 바뀌도록 설계됐다(§2.3 순서 증인과 맞음).
- 0.1.5·0.1.6 에서 바뀐 RED 셀을 다시 판독했다. AC-007 양식 공존 단언(오늘 빨강, 옳은 이유), AC-024 CLI 사례(오늘 `clause mismatch` 로 빨강, 옳은 이유), AC-024 `loader_unchanged`·AC-025 `load`·`dry_run`·보존 실행(오늘 초록, 보존 가드), AC-025 `real`(스텁 오류로 빨강, M5 에서 뒤집힘). 잘못된 이유의 빨강이나 영원한 빨강으로 판정된 것은 없다.
- 보존 가드의 RED 칸: `loader_unchanged` 와 보존 실행은 M-20 (iii), AC-025 `dry_run` 은 M-24 다. `load` 하위 사례와 `loader_unchanged/absolute_escape` 는 선언된 RED 칸이 없는 증인이다(N6, optional).
- 잠재적 잘못된 이유의 빨강 하나: 보존 실행의 "source unchanged" 단언이 기준 커밋을 고정하지 않아, develop 흡수가 `internal/spec/lint.go` 를 바꾸면 이 작업과 무관하게 빨개진다(N4). 현재 로컬 develop 에는 병합 기점 `ed71054d3` 이후 두 파일 변경이 없다(`git log --oneline ed71054d3..develop -- internal/spec/lint.go internal/spec/lint_test.go` → 출력 없음). 대조로 `internal/spec/` 전체는 5커밋이 바뀌었다.
- green 경로 칸: plan.md M1–M7 의 Exit 줄이 모든 AC 를 넘기며, D4 로 추가된 두 보존 행은 M2 Exit(L98)에 들어 있다.

## 산출물 일관성

- 계수: spec HISTORY L24, progress L12, 판정서 §16.2 모두 21 · 25 · 29 로 이번 계수와 같다. 버전: spec frontmatter `"0.1.6"`, progress L8, plan L3, design 머리말(0.1.6 과 D4 언급)이 일치한다. research.md 는 0.1.4 이후 바뀌지 않았고, progress L9 가 그 사실을 적었다.
- 부속 산출물 `^status:` 줄 수: plan 0 · acceptance 0 · design 0 · research 0 · progress 0.
- 판결 밖의 새 요구사항·결정·범위: 없음. 0.1.6 에서 추가된 것은 D4 운영자 결정(판정서 §16.2)의 인코딩, 보존 행 두 개, M-20 (iii), 담당 이관뿐이고 ID 는 늘지 않았다. §F 새 제외 항목(L200–L202)은 D4 의 "후속 카드 후보" 문구를 그대로 옮긴 것이다.
- lint: `lint-0.1.6.txt` 가 트리 `564c370b5` 에서 `0 error(s), 0 warning(s)`, `LINT_EXIT=0` 을 기록한다(INFO 1건은 ownership 트레일러). 판정 바이너리는 `ed71054d3-dirty`(판정서 §16.3). SPEC 파일은 그 뒤 바뀌지 않았다. `-dirty` 내용은 이 감사도 관측하지 않았다(Gap).

## Defects Found (structured defect-list)

N1. TST-NUMERIC-TOKEN-VACUOUS — acceptance.md:L100 (AC-002 "the count (`0` / `2`)"), L110 (AC-003 "the exact-match count `0`"), L224 (AC-014 (b) `2`, (e) `1`), L259–L261 (AC-018 "the decimal line number L1 … L2"), L269 (AC-019 count `1`); §D.2 L384 (M-2), L402 (M-15) — 이 단언들은 숫자가 오류 문자열 안에 "있는지"만 본다. 그런데 같은 오류가 `t.TempDir()` 아래 경로를 담고, 그 경로는 `…/001` 과 이 기계의 `TMPDIR=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/`(이번 실행 `echo $TMPDIR`)가 품은 숫자 `0 1 2 3 4 7 8` 을 이미 포함한다. 문구 그대로의 부분 문자열 검사라면 M-2(정규화 뒤 개수 2)는 AC-003 에서 확실히 살아남는다. M-15 변형 (i) 줄 번호 누락과 (iii) 블록 기준 줄도 L1·L2 가 경로에 있는 숫자면 살아남는다. §2 뮤턴트 탐침에서 요구사항을 어기며 통과하는 뮤턴트를 쓸 수 있다. — Severity: major — Class: blocking — Required fix: 공통 규칙(acceptance.md:L15–L21)에 한 줄을 넣는다. "개수·줄 번호 단언은 경로와 섞일 수 없는 형태로 검사한다 — 오류 문자열에서 픽스처 경로를 모두 지운 나머지에 대해, 그리고 숫자는 단어 경계(또는 오류가 노출하는 구조화된 필드)로 비교한다." AC-003 은 "개수 `0` 이 있고 `2` 는 없음"을, AC-018 은 "줄 번호 L1 이 있고 블록 기준 번호는 없음"을 같은 규칙으로 단언하게 한다. AC 번호는 늘지 않는다.

N2. TST-DARWIN-TEMPDIR-SYMLINK — acceptance.md:L332, L363 (`B = t.TempDir()`), L359, L409 (M-22 (ii) "against the root", 해석 여부 미지정), L410 (M-23), L351 (CLI 사례 "containing the offending path", 형태 미지정), L415 — macOS 에서는 `/var` 가 `private/var` 로 가는 링크다(이번 실행 `ls -ld /var` → `/var -> private/var`). 따라서 `B` 자체가 링크 아래에 있다. M-23 은 "후보는 풀지 않고 루트는 푼다"이므로, darwin 에서는 해당 지점의 **모든** 후보가 `/var/…` 대 `/private/var/…` 로 밖이라 판정된다. `symlinked_registry`·`symlinked_file`·`symlinked_log` 는 그 뮤턴트에서도 거부되어(오류가 해석 전 형태의 경로를 담아 단언 통과) 초록으로 남고, 뮤턴트는 선언되지 않은 `in_root_control` 에서만 죽는다. M-22 (ii) 도 루트를 해석하는 읽기에서는 같다. 뮤턴트는 여전히 죽지만, 이 레인이 도는 darwin 에서 E2 표(plan.md L77)가 "선언 행 초록"을 기록하게 되고, L415 의 기준이 거짓이 된다. plan.md:L140 R-7 은 macOS 임시 디렉터리 링크를 알고 있으나 M-24 쪽만 다룬다. — Severity: minor — Class: blocking — Required fix: AC-024·AC-025 픽스처 첫 줄을 "`B` 는 `filepath.EvalSymlinks(t.TempDir())` 로 풀어 둔다(의도적 링크 `B/link`, `P/linkdir`, `P/linked`, `P/.moai/research` 만 링크로 남긴다)"로 바꾼다. M-22 (ii) 의 비교 대상 루트를 "해석 전 `projectDir`"로 명시한다. CLI 사례도 "정리된 절대 형태 또는 링크 해석 형태(테스트가 둘 다 계산)"로 맞춘다.

N3. CN-REGISTRY-READ-ORDER — spec.md:L94 (REQ-CAA-020 "no read or write of the apply step shall reach another directory tree") vs plan.md:L97 ("for the registry path and for every entry's joined `file:` as soon as the registry is loaded"), design.md:L67 표("at registry load"), L31 — plan 문구대로 등록부 경로 검사를 `LoadRegistry` **뒤에** 두면, 상대 경로 탈출(`relative_env_escape`)과 링크 탈출(`symlinked_registry`)에서 로더가 루트 밖 등록부를 이미 읽은 다음에 거부한다. 쓰기는 없으므로 모든 AC 가 초록이고, REQ-CAA-020 의 "읽기도 닿지 않는다"만 어긴다. 이 순서의 뮤턴트는 어떤 AC 로도 죽지 않는다. — Severity: minor — Class: blocking — Required fix: plan.md M2 와 design.md §B 2단계·§C.2 표에 "등록부 경로 검사는 `LoadRegistry` 가 파일을 읽기 **전에**, `file:` 검사는 적재 **직후에**"를 적는다. 읽기 순서는 스냅숏으로 관측할 수 없으므로 §E.4 에 "등록부 경로 검사가 읽기에 앞선다는 순서는 코드 리뷰로만 확인된다"는 Gap 을 한 줄 선언한다. 요구사항을 새로 만들지 않고 기존 REQ-CAA-020 의 문구를 계획에 맞춘다.

N4. VC-UNPINNED-BASE-WRONG-REASON — acceptance.md:L369 (`git diff --stat <plan-phase HEAD> HEAD -- internal/spec/lint_test.go internal/spec/lint.go` prints nothing) — 기준 커밋이 자리표시자로 남아 있다. run 단계에서 develop 을 흡수하면(§4.1 흐름) 다른 카드가 바꾼 `internal/spec/lint.go` 가 이 diff 에 잡혀, 이 작업과 무관한 이유로 빨개질 수 있다(잘못된 이유의 빨강). 이동 참조 판정 규칙(verification-claim-integrity §2.1)으로 보면 R2(pre-flight 고정)가 맞는 자리다. — Severity: minor — Class: optional — Required fix: plan.md §C.1 에서 `BASELINE_SHA=$(git rev-parse HEAD)` 를 기록하게 하고, 단언을 "이 카드의 커밋이 두 파일을 건드리지 않았다 — `git log --no-merges --format=%H $BASELINE_SHA..HEAD -- internal/spec/lint.go internal/spec/lint_test.go` 출력 없음"으로 바꾼다.

N5. PLAN-PREFLIGHT-COUNT-MISMATCH — plan.md:L42 — 명령 `grep -rn --include='*.go' 'LoadRegistry(' internal cmd pkg` 은 테스트 파일까지 세어 이번 실행에서 34줄을 출력한다. 주석은 "expect 7 call sites outside _test.go plus the definition"이고, §C 는 "stop and report on any mismatch"다. 명령을 글자 그대로 실행하면 사전 점검이 불필요하게 멈춘다. `| grep -v _test.go` 를 붙이면 8줄이다(이번 실행). — Severity: minor — Class: optional — Required fix: 명령 끝에 `| /usr/bin/grep -v _test.go` 를 붙이고 기대값을 "8 lines (definition + 7 call sites)"로 적는다.

N6. ACC-MINOR-MAP-DRIFT — acceptance.md:L408 vs L359; L396 vs L224; L377; L350 — (a) §D.2 M-21 행은 `file:` 변형의 킬 사례에서 `symlinked_file` 을 빠뜨렸는데, L359 는 그 행을 적는다(판독상 M-21 `file:` 변형은 `symlinked_file` 도 죽인다). (b) M-10 행은 AC-014 (b)–(f) 를 적지만, (f) 의 오류는 REQ-CAA-017 검사에서 나오므로 "적용 검증 생략" 뮤턴트가 (f) 를 바꾸지 않을 수 있다. 뮤턴트는 (b)–(e) 로 죽는다. (c) L377 "no containment mutant reaches it"(`load`)는 부정확하다. M-20 (iii) 은 `load` 에 닿지만 초록으로 남는다. (d) `load` 와 `loader_unchanged/absolute_escape` 는 선언된 RED 칸 없는 증인이다. — Severity: minor — Class: optional — Required fix: M-21 행에 `symlinked_file` 추가, M-10 행을 "(b)–(e)"로, L377 을 "no declared mutant turns it RED"로 고친다. (d)는 그대로 둬도 되며, 원하면 `absolute_escape` 에 "로더의 절대 경로 거부 제거" 뮤턴트를 M-20 (iv) 로 적는다.

## Recommendation

FAIL (0.84 < 0.85, blocking 3건). manager-spec 은 새 요구사항·AC 번호 없이 다음 순서로 고친다(AC 25 상한 유지).

1. N1 — acceptance.md 공통 규칙에 "경로를 지운 오류 문자열에 대해, 단어 경계로 숫자를 비교" 규칙을 넣고, AC-003·AC-018 에 반대 숫자(정규화 개수 `2`, 블록 기준 줄)가 **없음**을 단언하게 한다(L15–L21, L110, L259–L261).
2. N2 — AC-024·AC-025 의 `B` 를 `EvalSymlinks(t.TempDir())` 로 고정하고, M-22 (ii) 의 루트를 "해석 전 `projectDir`"로, CLI 사례의 경로 형태를 두 형태 허용으로 명시한다(L332, L351, L363, L409).
3. N3 — plan.md M2(L97)와 design.md §B·§C.2(L31, L67)에 등록부 경로 검사를 적재 전에 둔다고 적고, 읽기 순서는 관측 불가 Gap 으로 §E.4 에 한 줄 선언한다.
4. N4–N6 은 optional 이다. N5 는 한 줄 수정이라 같은 개정에 넣는 편이 싸다.

다음 회차(iteration 3, Tier L 상한)는 N1–N6 에 한정한 델타 감사와 D1–D16 회귀 확인으로 진행한다. 3회차가 FAIL 이면 규칙상 사용자 에스컬레이션(부채를 안고 통과 / 범위 축소 / 명시적 연장)으로 넘어간다.
