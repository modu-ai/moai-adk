# SPEC 감사 보고서: SPEC-CON-AMEND-APPLY-001

Iteration: 3/3 (Tier L 상한, 최종 회차)
Verdict: PASS
Overall Score: 0.89 (Tier L 합격선 0.85 충족, 2회차 0.84 대비 +0.05)

- 감사 대상 트리: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t659`, 브랜치 `WT-amend-apply`, HEAD `e39168140d499307fd3daf4ed2a19a0f6ae1b039`. `git rev-parse --show-toplevel` / `git branch --show-current` / `git rev-parse HEAD` 로 확인했고 지시와 일치한다.
- 대상 개정: `spec.md` `version: "0.1.7"`, `tier: L`, `status: draft`. SPEC 파일의 마지막 변경은 `183bba916` 이고, `git diff --stat 183bba916 HEAD -- .moai/specs` 는 출력 없이 exit 0 이다(뒤 커밋 `e39168140` 은 `.moai/reports/t659/` 세 파일만 바꿨다).
- 감사 범위: 2회차 권고대로 N1–N6 에 한정한 델타 감사, 1회차 D1–D16 회귀 확인, 위임문이 지정한 세 가지 판정(N4 `--first-parent`, N6(d) M-20 (iv), N1 숫자 규칙), 0.1.7 이 건드린 행과 M-2·M-15·M-22·M-23 의 뮤턴트 재점검, 필수 항목 재확인.
- 주 변경 보기: `git diff 1fe4a0289 183bba916 -- .moai/specs/SPEC-CON-AMEND-APPLY-001` (5 files, +58 −49). 판정은 diff 가 아니라 파일 본문을 읽고 내렸다.
- 작성자 추론 맥락은 받지 않았다(Reasoning context ignored per M1 Context Isolation). 교차 모델 감사 MCP 는 부르지 않았고 판정은 이 감사자 단독이다.
- 실행하지 않은 것: `go test`, `go build`, `moai todo`. 뮤턴트 생존·사망과 코드 동작에 관한 주장은 모두 **코드·문서 판독**이다. N4 의 git 동작을 임시 저장소에서 재현하려 했으나 워크트리 격리 가드가 워크트리 밖 git 호출을 거부해 실행하지 못했다. 그 판정도 git 문서와 이력 단순화 규칙에 대한 판독이다.
- 생산 코드 무변경 재확인: `git diff --stat 5a066994b HEAD -- internal/constitution internal/cli/constitution.go internal/spec internal/cli/doctor.go` → 출력 없음, exit 0. 대조 `git diff --stat 5a066994b HEAD | tail -1` → `25 files changed, 2415 insertions(+)`.

## 판정 요약

2회차 결함 N1–N6 은 모두 해결됐다. 가장 무거웠던 N1 은 공통 규칙에 "경로를 지운 오류 문자열에서, 앞뒤에 숫자가 없는 온전한 수로 비교"라는 규칙이 들어가면서 풀렸고, M-2 와 M-15 의 세 변형은 이제 선언된 사례에서 죽는다. N2 는 AC-CAA-024·025 의 기준 디렉터리를 `filepath.EvalSymlinks(t.TempDir())` 로 고정해 darwin 에서도 M-23 세 변형이 자기 행에서 죽는다. N3 은 계획·설계에 순서를 적고 관측 불가를 Gap 으로 선언해 풀렸다. 1회차 D1–D16 은 하나도 되돌아가지 않았다.

작성자가 권고를 넘어선 두 선택은 모두 옳다. 특히 N4 의 `--first-parent` 는 2회차 권고 명령 자체의 결함을 바로잡은 것이다 — 2회차가 제시한 `--no-merges` 단독 형태는 develop 흡수로 들어온 다른 카드의 커밋을 여전히 나열한다. 이 점은 아래 N4 판정에서 이 감사자의 오류로 기록한다.

새 결함은 네 건이 나왔으나 모두 optional 이다. 필수 항목 7개가 통과하고 점수가 0.89 로 합격선을 넘으므로 PASS 다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `grep -cE '^- \*\*REQ-CAA-[0-9]+' spec.md` → `21`, 같은 패턴의 `sort | uniq -d` → 출력 없음. 여섯 파일 전체 구별 ID `REQ-CAA-*` 21개. 빈 번호·중복 없음.
- [PASS] MP-2 GEARS 형식 (요구사항 층에 대해 판정): 0.1.7 은 REQ 본문을 바꾸지 않았다(`spec.md` 변경은 HISTORY 한 행, L33 작성 트리 문장, §E.4 Gap 한 항목뿐). 2회차에서 확인한 라벨 분포(event-driven 13 · ubiquitous 5 · unwanted 2 · state-driven 1)와 본문 형태가 유지된다. 예: spec.md:L95 REQ-CAA-020 "The registry, … shall all lie inside the same project root" + "**When** the registry path … lies outside", L99 REQ-CAA-021 "**When** the check finds any of the following outside". AC 층의 Given-When-Then 은 이 항목에서 채점하지 않았다.
- [PASS] MP-3 YAML frontmatter: spec.md:L2–L13 에 12개 필수 필드가 모두 있다 — `version: "0.1.7"`(인용), `status: draft`, `created`/`updated` `2026-09-11`, `priority: P1`, `phase: "v3.2.0 target"`(수명 단계명 아님), `lifecycle: spec-anchored`, `tags` 쉼표 문자열. 거부 별칭 없음. `tier: L`, `related_specs` 는 선택 필드.
- [N/A] MP-4 언어 중립성: 템플릿이 아닌 Go 내부 패키지(`internal/constitution`, `internal/cli`) 변경이다.
- [PASS] MP-5 D7 교차 SPEC: `grep -ohE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → 자기 ID 와 `SPEC-V3R2-CON-002` 둘. 후자의 `^status:` → `implemented`. BLOCKING 없음.
- [PASS] MP-6 D8 교차 플랫폼: `grep -c syscall spec.md` → `0`.
- [PASS] MP-7 확인 요청 표식: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → 출력 없음, exit 1.

## Tier 와 예산

| 항목 | 계수(이번 실행) | Tier L 상한 | 판정 |
|---|---|---|---|
| 요구사항 | 21 (정의 줄 21, 여섯 파일 구별 ID 21) | 25 | 범위 안 |
| 인수 조건 | 25 (`^### AC-CAA-` 25, 여섯 파일 구별 ID 25, 최댓값 AC-CAA-025) | 25 | 상한과 같음, 여유 0 |
| 뮤턴트 | 29 (`^\| M-` 행 29, 중복 ID 0) | 상한 없음 | — |

부속 산출물 `^status:` 줄 수: plan 0 · acceptance 0 · design 0 · research 0 · progress 0. 합격선은 `spec-workflow.md` § SPEC Complexity Tier 표의 Tier L 값 0.85 다. 0.1.7 은 ID 를 하나도 늘리거나 바꾸지 않았다(spec HISTORY 0.1.7 행, progress.md L12 "unchanged in 0.1.7").

## Category Scores

| 차원 | 점수 | 기준 구간 | 근거 |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.0 사이 | N1(acceptance.md:L22 숫자 규칙), N2(L335·L366 기준 디렉터리, L412 M-22 (ii) "unresolved `projectDir` as passed"), N3(plan.md M2, design.md §B 2단계·§C.2 L68) 의 모호성이 풀렸다. 남은 것: AC-CAA-023 L319 "which fires first is a run-phase choice" 가 0.1.7 에서 고정된 순서와 어긋나고(X1), REQ-CAA-020 의 "read" 가 경로 해석의 메타데이터 조회까지 포함하는지 적혀 있지 않다(X2). 둘 다 구현을 갈라놓을 정도는 아니다. |
| Completeness | 0.90 | 0.75–1.0 사이 | 필수 섹션과 Tier L 산출물이 모두 있다. §E.4 에 Gap 셋이 선언돼 있다(spec.md:L170 G5 승인됨, L171 백업·임시 쓰기 실패 "not yet approved", L172 읽기 순서). 둘째 Gap 은 리드 승인이 아직 없다. |
| Testability | 0.85 | 0.75–1.0 사이 | 숫자 단언이 경로 숫자에서 분리됐고(L22), 반대 숫자 부재 단언이 AC-CAA-003(L113)·018(L262–L264)에 들어갔다. darwin 에서도 M-23 이 선언 행에서 죽는다(L335, L418). 남은 한계: 읽기 순서는 어떤 AC 도 관측하지 못하며(L172, 선언됨), AC-CAA-018 의 "다른 곳에 온전한 수로 나타나지 않음" 조건은 구현의 오류 문구가 정해진 뒤에야 확정된다(아래 뮤턴트 절). |
| Traceability | 0.90 | 0.75–1.0 사이 | 21 REQ 모두 AC 가 있고(acceptance.md §D.0 L58–L82), AC 가 가리키는 REQ 는 모두 존재한다. REQ-CAA-020 의 "no read … shall reach another directory tree" 는 2회차에 어느 AC 에도 연결되지 않았으나, 이제 §E.4 L172 에 코드 리뷰 전용으로 명시됐다. REQ-CAA-021 의 "keeps its present refusal" 절은 M-20 (iv) 로 RED 칸을 얻었다. |

산술 평균 0.8875 → 0.89. 합격선 0.85 충족. 2회차 0.84 대비 +0.05(명확성 +0.05, 완결성 ±0, 검증 가능성 +0.10, 추적성 +0.05). 점수가 올랐으므로 STOP 신호는 내지 않는다.

## Regression Check — 1회차 D1–D16

0.1.7 이 acceptance.md 에서 바꾼 곳은 L15–L16, L22, L53–L54, L103, L113, L227, L260–L264, L272, L335, L354, L360, L362, L366, L372, L380, L387, L399, L405, L410–L412, L418 이다(`git diff -U0 1fe4a0289 183bba916 -- …/acceptance.md | grep '^@@'`). 각 결함과 겹치는 곳을 읽었다.

| ID | 판정 | 근거 |
|---|---|---|
| D1 M-5b 사멸 불가 | RESOLVED 유지 | AC-CAA-012(L195–L213)는 0.1.7 에서 바뀌지 않았다. rename-then-fail 모드, 호출 횟수 단언, 다섯 실패 행, 킬 맵 그대로. |
| D2 CLI 사례 판별력 | RESOLVED 유지 | L354 는 경로 형태만 넓혔다("cleaned absolute form or its symbolic-link-resolved form … `B` is the resolved base"). `B/other` 사본의 clause 를 다르게 두는 설정, `amendment failed`(`internal/cli/constitution.go:544`)·`clause mismatch`(`:529`) 부재 단언은 남아 있다. 두 줄 번호는 이번에 `sed -n '525,546p'` 로 다시 확인했다. |
| D3 링크 해석 지점 | RESOLVED 유지, 보강 | `symlinked_registry`·`symlinked_file`·`symlinked_log` 행(L343–L345)과 M-23 세 변형(L413) 유지. 2회차의 darwin 한계(N2)도 풀렸다. |
| D4 공유 로더 영향 범위 | RESOLVED 유지 | REQ-CAA-021(L99) 문구 불변. plan.md §C.1 명령이 이제 테스트 파일을 거른다(N5). 이번 실행 `… \| /usr/bin/grep -v _test.go \| wc -l` → `8`, 거르지 않으면 `34`. |
| D5 AC-007 RED 칸 | RESOLVED 유지 | L36, L153 불변. |
| D6 AC-005(b) | RESOLVED 유지 | L128–L130 불변. |
| D7 복원 단계 | RESOLVED 유지 | AC-CAA-012 불변. 백업·임시 쓰기 Gap(spec.md:L171)은 여전히 리드 판단 대기로 선언돼 있다. |
| D8 AC 가 REQ 보다 엄격 | RESOLVED 유지 | REQ-CAA-018·AC-CAA-021(L289–L298) 불변. |
| D9 락 쓰기 예외 | RESOLVED 유지 | REQ-CAA-021(L99) "the lock file a real-mode run creates before it loads the registry is the one exception" 불변. |
| D10 "same kind" | RESOLVED 유지, 보강 | L227 사례별 부분 문자열 유지. (b)·(e) 의 숫자가 이제 숫자 규칙 아래 비교된다. |
| D11 개수 핀 중복 | RESOLVED 유지 | L366 이 여전히 `wantRegistryEntries`(`registry_sync_test.go:49`, 패키지 `constitution_test` — 이번에 `grep -n` 으로 확인)를 가리키고, L368 은 `- id:` 줄 수와 비교한다. |
| D12 RED 표시 | RESOLVED 유지 | 0.1.7 에서 새로 들어간 문장 중 RED 를 단정하는 것은 L360·L380 뿐이고 둘 다 "predicted from code reading at `54ca2e3b6`" 머리를 달고 있다. L362·L410 의 "turns RED" 는 뮤턴트 킬 예측이며 2회차에도 같은 형식이었다. |
| D13 환경 변수 dry-run | RESOLVED 유지 | REQ-CAA-013 불변. |
| D14 GEARS 라벨 | RESOLVED 유지 | REQ 본문 불변. |
| D15 작성 트리 | RESOLVED 유지 | spec.md:L33 에 0.1.7 이 `1fe4a0289` 에서 작성됐다고 적혀 있다. 이번 실행 `git diff --stat 564c370b5 1fe4a0289` → 4 files(`lint-0.1.6.txt`, `lint-binary-0.1.6.txt`, `plan-audit-iter2.md`, `verdict.md`), 서술과 일치. `git diff --stat 5a066994b 1fe4a0289 -- internal/constitution internal/cli/constitution.go internal/spec` → 출력 없음, 서술과 일치. |
| D16 진입점 | RESOLVED 유지 | L109 `Execute(dryRun=false)`, L367 `LoadRegistry("B/link/…", "B/link")` 불변. |

회귀 없음.

## N1–N6 해결 판정

| ID | 판정 | 근거 |
|---|---|---|
| N1 숫자 단언 공허 | RESOLVED | acceptance.md:L22 공통 규칙 — 먼저 AC 가 지정한 경로를 단언하고, 테스트가 만든 `t.TempDir()` 아래 경로를 정리된 절대 형태와 링크 해석 형태 모두 오류에서 지운 뒤, 숫자는 앞뒤에 다른 숫자가 없는 온전한 수로(예 `(^\|[^0-9])2([^0-9]\|$)`) 또는 구조화된 필드로 읽는다. 픽스처의 다른 숫자열(entry `id`, clause 문구)은 단언 대상 숫자와 다르게 유지한다. 적용: AC-CAA-002 L103, AC-CAA-003 L113("the exact-match count `0` is present and the whitespace-normalized count `2` is absent"), AC-CAA-014 L227, AC-CAA-018 L260–L264(두 번째 로그에 3줄 산문 오프셋 추가, 상대 줄 번호 부재 단언, "none occurring as a whole number elsewhere in the path-stripped error (the digits of `EVO-X-001` and `LEARN-20260911-009` included)"), AC-CAA-019 L272. M-2 행 L387, M-15 행 L405 가 이 단언을 인용한다. 판정은 아래 "N1 숫자 규칙" 절. |
| N2 darwin 임시 디렉터리 링크 | RESOLVED | AC-CAA-024 L335 "a base `B` equal to `filepath.EvalSymlinks(t.TempDir())` — resolved, so no directory on the way to `B` is itself a symbolic link … the only symbolic links in the fixture are … `P/linkdir`, `P/linked`, `P/.moai/research`, and `B/link`". AC-CAA-025 L366 동일. M-22 (ii) L412 "against the unresolved `projectDir` as passed". CLI 사례 L354 두 형태 허용. plan.md R-7(L141) 이 M-23·M-22 (ii) 까지 다룬다. design.md 에 같은 문장 추가(§C 끝 목록). L418 이 darwin 과 Linux 를 명시한다. |
| N3 등록부 읽기 순서 | RESOLVED | plan.md M2 "in both callers on the registry path before `LoadRegistry` reads the file — so a refused registry is never read … then on every entry's joined `file:` immediately after the registry is loaded … The order is observable by code review only". design.md §B 2단계("containment check on the registry path, before LoadRegistry reads it … a refused registry is never read"), §C.2 L68. spec.md §E.4 L172 Gap 선언. 2회차 요구 수정과 일치한다. 이 수정이 AC-CAA-023 의 한 괄호문을 낡게 만든 점은 X1. |
| N4 기준 커밋 미고정 | RESOLVED (권고보다 나은 형태) | plan.md §C.1 `BASELINE_SHA=$(git rev-parse HEAD)   # before the first run-phase commit; record the resolved value in progress.md §E.2`. AC-CAA-025 L372 `git log --first-parent --no-merges --format=%H $BASELINE_SHA..HEAD -- internal/spec/lint.go internal/spec/lint_test.go` 가 비어야 한다. `verification-claim-integrity.md` §2.1 의 R2(pre-flight 고정)에 해당한다. 판정은 아래 "N4 `--first-parent`" 절. |
| N5 사전 점검 개수 | RESOLVED | plan.md §C.1 명령에 `\| /usr/bin/grep -v _test.go` 가 붙고 기대값이 "8 lines (definition + 7 call sites)". 이번 실행 → `8`. |
| N6 킬 맵 표류 | RESOLVED | (a) M-21 `file:` 변형에 `symlinked_file` 추가(L411). (b) M-10 이 "(b)–(e)"로 좁혀지고 (f) 가 REQ-CAA-017 검사로 막힌다는 이유를 적었다(L399). (c) L54·L380 "no declared mutant turns it RED"(`load`). (d) M-20 (iv) 추가(L410), L53·L360·L362 에 `absolute_escape` 의 RED 칸으로 연결. 판정은 아래 "N6(d) M-20 (iv)" 절. |

### N4 `--first-parent` 판정 (문서·이력 단순화 규칙 판독)

작성자의 판단이 옳고, 2회차 권고 명령은 틀렸다.

- **`--no-merges` 단독은 흡수된 커밋을 걸러내지 못한다.** 카드 브랜치에서 `git merge develop` 으로 만든 병합 커밋 M 을 생각하자. 첫 부모는 카드 tip, 둘째 부모는 develop tip 이다. `$BASELINE_SHA..HEAD` 범위에는 둘째 부모 쪽에서만 닿는 develop 커밋이 들어 있다. 경로 한정 로그의 기본 이력 단순화는 M 이 한 부모와 해당 경로에서 같으면(TREESAME) 그 부모만 따라간다. 카드가 `lint.go` 를 건드리지 않았다면 M 의 `lint.go` 는 둘째 부모와 같으므로 git 은 둘째 부모 쪽을 걷고, 다른 카드의 `lint.go` 커밋을 출력한다. `--no-merges` 는 M 자체만 뺄 뿐이다. 2회차 N4 가 권한 명령은 그래서 이 AC 와 무관한 이유로 빨개질 수 있었다. 이 감사자의 오류로 기록한다.
- **`--first-parent` 는 그 커밋을 뺀다.** 병합 커밋을 만나면 첫 부모만 따라가므로 둘째 부모 쪽 커밋은 범위에 들어오지 않는다. 첫 부모 사슬의 병합 커밋은 첫 부모와 비교되어 흡수분이 차이로 보이지만, `--no-merges` 가 그것을 출력에서 뺀다. 두 옵션의 조합이라야 흡수분이 사라진다.
- **카드 자신의 커밋은 남는다.** 카드 브랜치가 선형이라면 run-phase 커밋은 모두 첫 부모 사슬 위에 있고, 경로를 건드린 커밋은 첫 부모와의 차이로 출력된다. `$BASELINE_SHA` 를 기록한 뒤 첫 커밋이 흡수 병합이어도 M 의 첫 부모가 `$BASELINE_SHA` 이므로 사슬은 거기서 끝난다.
- **이 조합이 가리는 것(optional X3 로 기록).** (1) 병합 커밋 안에서 이루어진 편집 — 충돌 해결이나 수동 편집으로 `lint.go` 를 바꾼 흡수 병합은 `--no-merges` 에 가려진다. 카드가 첫 부모 쪽에서 `lint.go` 를 건드리지 않았다면 그 파일에 충돌이 날 수 없으므로, 남는 것은 의도적 수동 편집뿐이다. (2) 카드 자신의 작업이 둘째 부모로 들어오는 경우 — 카드 안의 보조 브랜치나 격리 워크트리 결과를 비-fast-forward 병합으로 합치면 그 커밋은 둘째 부모 쪽에 있어 보이지 않는다. 이 저장소의 레인 절차(manager-develop 이 카드 워크트리에서 직접 커밋)에서는 둘 다 일어날 가능성이 낮다. 이 단언은 보조 증인이고, 주 증인은 `TestLinter_AC08_DanglingRuleReference` 통과 자체다.
- **저장소 로컬 규칙과의 차이.** `.claude/rules/local/gitflow-lane-protocol.md` §8 [HARD] 는 "이 카드가 무엇을 바꿨는가"를 흡수한 ref 와의 merge-base 부터(`CARD_BASE=$(git merge-base develop HEAD)`) 재라고 한다. `git diff "$CARD_BASE"..HEAD -- <두 파일>` 은 위 (1)(2)를 모두 잡는다(순 트리 차이이므로). 병합 뒤에는 공허해지는 한계가 있지만 이 AC 는 병합 전 run-phase 트리에서 도므로 해당하지 않는다. SPEC 의 형태는 흡수 문제를 올바르게 풀지만 그 규칙이 정한 판정식과 다르다.

### N6(d) M-20 (iv) 판정 (코드 판독)

- **선언된 하위 사례로 죽는다.** M-20 (iv)는 `LoadRegistry` 의 절대 경로 거부(`internal/constitution/loader.go:82-87` — 이번에 `sed -n '70,95p'` 로 확인, `:86` 에 `"registry path %q escapes project dir %q"`)를 없앤다. `loader_unchanged/absolute_escape` 는 `LoadRegistry("B/other/.claude/rules/moai/core/zone-registry.md", P)` 를 부르고 `escapes project dir` 를 담은 오류를 기대한다(L353). 뮤턴트 아래에서 로더는 `B/other` 의 사본을 읽어 오류 없이 돌려주므로 단언이 빨갛다. `relative_registry` 픽스처가 `B/other` 에 등록부 사본을 두므로 "파일이 없어서" 나는 다른 오류로 빨개질 여지도 없다.
- **D4 와 충돌하지 않는다.** D4 는 "`LoadRegistry` 는 검사를 받지 않고 현재 동작을 유지한다"이고, REQ-CAA-021(L99)은 이를 "including its present refusal of an absolute registry path outside `projectDir`"로 명문화한다. M-20 (iv)는 바로 그 절을 어기는 뮤턴트이고, `absolute_escape` 는 그 절의 RED 칸이다. 2회차에 이 절에는 RED 칸이 없었다. 로더에 해석 검사를 넣으라는 뜻이 아니며, 로더의 현재 동작을 고정하는 보존 가드다.
- **부수 확인.** (iv) 아래에서 amend 경로의 검사는 그대로이므로 AC-CAA-023 의 발산 행은 초록으로 남는다. 로더 거부가 사라져도 amend 경로가 스스로 막는다는 사실을 보여 주는 셈이며, (iv)가 그 행을 킬 대상으로 주장하지 않는 것도 옳다.

### N1 숫자 규칙 판정 (코드·문서 판독)

- **M-2 는 darwin·Linux 모두에서 AC-CAA-003 으로 죽는다.** 공백 정규화 뒤 개수는 2 다. 오류에서 규칙 파일 경로(두 형태)를 지우면 `/var/folders/…/001/…` 나 `/tmp/…/001/…` 의 숫자가 사라지고, 남은 문자열에 온전한 수 `2` 가 있고 `0` 이 없으므로 L113 의 두 단언이 모두 빨갛다. 정규화 방식이 달라 개수가 1 이 되면 치환이 진행되어 성공이 반환되므로 오류 단언으로 죽는다.
- **M-15 세 변형은 AC-CAA-018 로 죽는다.** (i) 줄 번호를 빼면 경로를 지운 문자열에 L1·L2 가 없다. (ii) 키를 빼면 `timestamp`/`approved_at` 부분 문자열이 없다. (iii) 블록·세그먼트 상대 줄을 쓰면 L1·L2 가 없고 상대 번호가 있다. 첫 로그는 산문 5줄, 둘째 로그는 산문 3줄 오프셋이 있어 파일 줄과 상대 줄이 갈린다. ID 의 숫자열 `001`, `20260911`, `009` 는 온전한 수 규칙상 한 자리 L 값과 겹치지 않는다(앞뒤가 숫자다).
- **구현 가능하고 공허하지 않다.** 테스트는 자기가 만든 경로를 알고 있고(모든 사례가 쓰기 전 실패라 백업·임시 파일 경로가 오류에 끼지 않는다), 정리 형태와 해석 형태를 모두 지운다. 한 가지 주의점: 구현이 `time.Parse` 오류를 감싸면 레이아웃 `2006-01-02T15:04:05Z07:00` 의 숫자열이 오류에 남는다. 그중 온전한 수로 걸리는 것은 `2006`, `15` 등 두 자리 이상 값이고, 한 자리 값은 `01`·`02`·`05` 처럼 앞에 0 이 붙어 걸리지 않는다. L264 의 "none occurring as a whole number elsewhere in the path-stripped error" 조건이 이것까지 덮지만, 이 조건은 구현의 문구가 정해진 뒤에야 확인된다. run-phase 는 경로를 지운 문자열에서 L1·L2 가 정확히 한 번 나오는지를 전제 단언으로 두면 된다. 규칙의 결함이 아니라 적용 시점의 문제라서 결함으로 올리지 않는다.

## 뮤턴트 재점검 (0.1.7 이 건드린 행 + M-2·M-15·M-22·M-23, darwin 기준, 코드 판독)

| 뮤턴트 | 선언 킬 사례 | 판독 결과 |
|---|---|---|
| M-2 | AC-CAA-003 | 죽음 — 경로를 지운 오류에 `2` 있음, `0` 없음 |
| M-10 | AC-CAA-014 (b)–(e) | 죽음 — (b) 두 번 출현, (c) 연속 줄(로더는 여러 줄 스칼라를 정상 파싱하므로 재작성 단계에서만 걸림), (d) 규칙 파일 없음(로더는 존재 여부를 경고로만 본다, `loader.go:141-149`), (e) 새 clause 존재 — 넷 모두 적용 검증에서만 잡히므로 dry-run 이 성공한다 |
| M-15 (i)(ii)(iii) | AC-CAA-018 각 변형 | 죽음 — 위 절 |
| M-20 (i) | AC-CAA-024 `relative_env_escape`, `symlinked_registry`, CLI | 죽음 — 기준 디렉터리가 해석돼도 결론 불변(로더는 상대 경로를 검사하지 않고, `P/linkdir/…` 은 문자열상 안쪽) |
| M-20 (ii) | CLI 사례 | 죽음 — `clause mismatch` 로 빠진다 |
| M-20 (iii) | `loader_unchanged` 셋, `internal/spec` 보존 실행 | 죽음 — 2회차 판독 유지. `lint_test.go:18-25`, `:218-222`, `lint.go:112-118` 을 이번에 다시 읽었다 |
| M-20 (iv) | `loader_unchanged/absolute_escape` | 죽음 — 위 절 |
| M-21 `file:` / 로그 | `absolute_file`(두 사례), `dotdot_file`, `sibling_prefix_file`, `symlinked_file` / `symlinked_log` | 죽음 — `symlinked_file` 은 `P/linked/rule.md` 가 `B/outside/rule.md` 에 닿아 거기에 쓰므로 스냅숏이 바뀐다 |
| M-22 (i) | `sibling_prefix_file` | 죽음 — `B/root-evil/rule.md` 가 문자열 접두사로 통과해 그 파일에 쓴다 |
| M-22 (ii) | `dotdot_file` | 죽음 — `P/../outside/rule.md` 는 정리하지 않은 문자열로 `P` 로 시작해 통과하고, 적용이 `B/outside/rule.md` 에 쓴다. `P` 가 이미 해석된 경로라 "해석 전 `projectDir`" 명시와 결론이 같다 |
| M-23 (i) | `symlinked_registry` | **darwin 에서도 죽음** — 루트 `P` 에 링크가 없으므로 해석하지 않은 후보 `P/linkdir/zone-registry.md` 가 안쪽으로 판정된다. 로더도 `filepath.Rel` 로 안쪽이라 보고 `B/other` 의 사본을 읽는다. dry-run 은 성공, real 은 링크를 통해 `B/other` 를 고쳐 스냅숏이 바뀐다 |
| M-23 (ii) | `symlinked_file` | darwin 에서도 죽음 — 같은 이유로 `B/outside/rule.md` 에 쓴다 |
| M-23 (iii) | `symlinked_log` | darwin 에서도 죽음 — 로그가 `B/outside-research` 에 쓰인다 |
| M-24 | `in_root_control/symlinked_root`, AC-CAA-025 `dry_run` | 죽음 — 의도적 링크 `B/link` 만 남았으므로 후보는 `B/root/…`, 루트는 `B/link` 로 비교되어 거부된다. 기준 디렉터리 해석이 킬 경로를 바꾸지 않는다 |

acceptance.md:L418 의 "Every mutant and every variant names at least one subtest that only a correct implementation passes, on darwin as on Linux" 는 재점검한 행 전부에서 성립한다(판독). 재점검하지 않은 행(M-1, M-3a/b, M-4, M-5a/b/c, M-6–M-9, M-11a/b, M-12–M-14, M-16–M-19)은 0.1.7 이 바꾸지 않았고, 2회차 판독에서 모두 죽음으로 판정했다. 선언된 한계는 그대로다 — 링크 생성을 거부하는 플랫폼에서는 M-23·M-24 가 Gap 이 된다(L358, L378).

## Defects Found (structured defect-list)

X1. CN-STALE-ORDER-PARENTHETICAL — acceptance.md:L319 (AC-CAA-023 "which fires first is a run-phase choice") vs plan.md M2, design.md §B 2단계·§C.2 L68 — 0.1.7 은 amend 경로에서 등록부 경로 검사를 `LoadRegistry` 호출 전에 두도록 정했다. 그러면 AC-CAA-023 의 절대 경로에서는 항상 검사가 먼저 오류를 낸다. 괄호문의 "어느 쪽이 먼저인지는 run-phase 선택"은 이제 계획과 맞지 않는다. 다만 이 AC 는 경로만 단언하고 문구는 단언하지 않으므로 테스트 결과는 달라지지 않는다. spec.md:L97 의 "either may produce the error" 는 요구사항 층 서술로서 여전히 참이다(로더도 읽기 전에 거부하므로 어느 순서든 REQ-CAA-020 을 만족한다). — Severity: minor — Class: optional — Required fix: L319 괄호문을 "the wording around the path is not asserted: the check runs first on the amend path (plan.md M2), and the loader's retained absolute-only refusal would refuse the same path"처럼 바꾼다. 문구 단언을 추가하지 않는다.

X2. CL-READ-SCOPE — spec.md:L95 (REQ-CAA-020 "no read or write of the apply step shall reach another directory tree") vs L99 (REQ-CAA-021 "has its symbolic links resolved"), L172 — 링크를 해석하면 경로 구성 요소를 `lstat` 하게 되고, 탈출하는 후보(`P/linkdir/zone-registry.md`, `../other/…`)라면 다른 트리의 메타데이터를 조회한다. "read" 가 파일 내용을 여는 것만 뜻한다면 두 요구사항은 양립하고, 메타데이터 조회까지 포함한다면 서로 모순이다. §E.4 L172 의 "reading a file changes none of them" 은 전자로 읽히고, 합리적인 엔지니어도 REQ-CAA-021 이 해석을 명시하므로 전자로 구현할 것이다. — Severity: minor — Class: optional — Required fix: REQ-CAA-020 근거 문단이나 §E.4 L172 에 "read means opening a file's contents; the metadata lookups REQ-CAA-021's symbolic-link resolution performs are not reads" 한 문장을 넣는다. 새 REQ·AC 는 필요 없다.

X3. VC-FIRST-PARENT-BLIND-SPOTS — acceptance.md:L372, plan.md §C.1 — 위 N4 판정 절의 (1)(2): 병합 커밋 안의 편집과 둘째 부모로 들어온 카드 자신의 커밋은 `--first-parent --no-merges` 에 보이지 않는다. 또 이 형태는 `.claude/rules/local/gitflow-lane-protocol.md` §8 [HARD] 의 merge-base 판정식과 다르다. 흡수로 인한 잘못된 이유의 빨강은 올바르게 막고, 이 저장소 절차에서 두 사각지대가 생길 가능성은 낮으며, 주 증인은 테스트 통과다. — Severity: minor — Class: optional — Required fix(택일): (a) 그대로 두고 L372 에 "commits carried by a merge's second parent and edits made inside a merge commit are not listed; the test run itself is the primary witness"를 한 줄 적는다. (b) 로컬 규칙에 맞춰 단언을 `CARD_BASE=$(git merge-base develop HEAD)` 뒤 `git diff --stat "$CARD_BASE"..HEAD -- internal/spec/lint.go internal/spec/lint_test.go` 출력 없음으로 바꾸고, 대조군으로 `git diff --name-only "$CARD_BASE"..HEAD \| wc -l` 이 1 이상임을 요구한다(병합 전 평가 전용).

X4. RR-WINDOWS-QUOTED-PATH — acceptance.md:L22, L319; `internal/constitution/loader.go:86` — 로더는 경로를 `%q` 로 찍는다. Windows 에서 `%q` 는 역슬래시를 두 번 쓰므로, 구현이 같은 서식을 쓰면 "오류가 경로를 담는다" 단언과 숫자 규칙의 경로 삭제가 Windows CI 에서 맞지 않는다. 숫자 규칙은 먼저 경로 존재를 단언하므로 공허하게 통과하지 않고 크게 실패한다. SPEC 의 뮤턴트 주장은 darwin·Linux 로 한정돼 있다. — Severity: minor — Class: optional — Required fix: 필요하면 공통 규칙에 "errors format paths with `%s`/`%v`, not `%q`, so the path assertions hold on Windows" 를 적거나, run-phase 가 처리하도록 두고 Residual-risk 로만 남긴다.

blocking 결함 없음.

## Evidence-bearing 요약

- **Claim**: 0.1.7 은 N1–N6 을 해결했고 D1–D16 에 회귀가 없으며, 필수 항목 7개를 통과하고 0.89 로 Tier L 합격선을 넘는다.
- **Evidence**: 이번 실행의 명령과 출력 — REQ 정의 `21`, AC 머리 `25`, 뮤턴트 행 `29`, 중복 0; 부속 `^status:` 전부 `0`; `NEEDS CLARIFICATION` exit 1; `syscall` `0`; `LoadRegistry(` 비테스트 `8`; 생산 코드 diff 출력 없음 exit 0; `git diff --stat 183bba916 HEAD -- .moai/specs` 출력 없음; `lint-0.1.7.txt` `0 error(s), 0 warning(s)` `LINT_EXIT=0`(INFO 1건은 ownership 트레일러).
- **Baseline-attribution**: 워크트리 `t659`, HEAD `e39168140`, SPEC 트리 `183bba916`. lint 증거의 판정 바이너리는 `lint-binary-0.1.7.txt` 기준 `v3.2.0-rc.7 … ged71054d3-dirty`.
- **Gaps**: `go test` 미실행 — 모든 뮤턴트·RED 주장은 판독 예측이다. N4 의 git 동작은 워크트리 가드 때문에 임시 저장소 재현을 못 했다. lint 바이너리의 `-dirty` 내용은 관측하지 않았다. 재점검하지 않은 뮤턴트 행은 2회차 판독에 기댄다.
- **Residual-risk**: 읽기 순서(L172)와 백업·임시 쓰기 실패 복원(L171)은 어떤 AC 도 관측하지 않는다. AC-CAA-018 의 숫자 격리 조건은 구현 문구가 정해진 뒤 확정된다. Windows 에서 경로 서식이 맞지 않을 수 있다(X4).

## Recommendation

PASS. 근거는 다음과 같다.

- MP-1: 정의 줄 21개, 여섯 파일 구별 ID 21개, 중복 0.
- MP-2: REQ 본문이 0.1.7 에서 바뀌지 않았고 GEARS 형태를 유지한다(spec.md:L95, L99 등).
- MP-3: 12개 필드 모두 있고 형식이 맞다(spec.md:L2–L13, `version: "0.1.7"`).
- MP-5·MP-6·MP-7: BLOCKING 없음, `syscall` 0, 표식 0.
- 2회차 blocking 셋(N1·N2·N3)이 모두 해결됐고 새 결함 넷은 모두 optional 이다.

run-phase 로 넘기기 전에 orchestrator 가 재량으로 처리할 것:

1. X1 은 한 문장 수정이라 run-phase 전에 고치는 편이 싸다.
2. X3 은 (a) 한 줄 서술 또는 (b) 로컬 규칙식 교체 중 하나를 고른다. (b)를 고르면 AC-CAA-025 의 보존 단언이 병합 전 평가 전용이라는 점도 함께 적는다.
3. X2·X4 는 선택 사항이다.
4. spec.md:L171 의 백업·임시 쓰기 실패 Gap 은 여전히 "not yet approved" 다. 이 감사는 Gap 선언을 결함으로 보지 않지만, 승인은 리드의 몫이며 Implementation Kickoff Approval 전에 결정돼야 한다.

이 PASS 는 Implementation Kickoff Approval(plan→run 사용자 승인)을 대신하지 않는다.
