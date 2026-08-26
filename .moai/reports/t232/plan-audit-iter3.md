# SPEC Review Report: SPEC-ZONE-REGISTRY-RESYNC-001
Iteration: 3 (Phase 1 Plan Audit Gate 재실행 — iter2 PASS-WITH-DEBT 후 아티팩트 개정으로 artifact-hash 무효화)
Verdict: **FAIL**
Overall Score: **0.825** (Tier M 임계 0.80 — **상회하나 FAIL; 사유는 아래 "FAIL 사유" 참조**)

감사 트리: `.claude/worktrees/t232` @ HEAD `79ee6a106`, 브랜치 `WT-zone-registry-drift`.
델타: `5073c7fbd`(spec v0.5.0 C안 확정 + acceptance GREEN 갱신 + plan §C/M1) + `79ee6a106`(progress §F + guard-failure-scenario §6).
범위: **델타 중심 + 전 아티팩트 채점**(리드 지시). M1 Context Isolation: 저자 추론 무시, 프롬프에 주어진 C안 근거 서술은 증거로 채택하지 않고 전부 재측정했다.

**⚠️ 아티팩트 이동 경고(감사 중 관측)**: 감사 수행 중 `zrr-spec-amend`가 워크트리에서 spec.md·acceptance.md를 계속 편집했다(첫 판독 시 HISTORY 0.5.0 행 2개 → 직후 1행으로 병합, REQ-ZRR-002/007 은퇴 조항 명시, AC-ZRR-002/003 Then 절 "live 97/97" 갱신). 본 보고서의 라인 인용과 판정은 **HEAD `79ee6a106` 커밋 판 + 감사 시점 워크트리 스냅샷** 기준이며, amend 미커밋 변경은 각 finding에 "해소 진행"으로 주석한다. amend가 커밋되면 artifact-hash가 다시 바뀌어 게이트 재실행 대상임을 리드가 인지해야 한다.

**🛑 STOP signal (LEAN — 점수 회귀)**: iter2 0.925 → iter3 0.825. 회귀 원인은 SPEC 구조 결함이 아니라 **C안 계약의 규범 표면 동기화가 불완전**한 것(수정 가능한 문안 결함) + 교차 모델이 발견한 신규 결함 3건이다. 따라서 범위 축소는 오진단이다. 오케스트레이터는 사용자에게 3옵션을 제시해야 한다: (1) 아래 결함 목록대로 1회 더 동기화 수정(권장 — 대부분 이미 amend 진행 중), (2) PASS-with-debt로 run 진입, (3) 계속 반복.

---

## FAIL 사유 (aggregate 0.825 > 0.80 에도 FAIL인 이유)

must-pass 방화벽 7건은 전부 PASS다. FAIL은 **BLOCKING AC 3개의 관측 계약이 문서 표면 간 모순/불가능 상태**이기 때문이다(M3 Testability 러브릭: 모순된 AC는 binary-testable이 아니다; M2: 결함 있음이 기본 가정, 이번엔 반증에 실패).

1. **AC-ZRR-007 평가-수 단언이 두 규범 문서에서 다른 값을 요구** — acceptance.md는 "101×2"(§D 매트릭스 19행)/"그 값이 `101` 이다"(§D.1 110행), guard-failure-scenario.md §6 P4는 "clause 97 / anchor 101". 가드가 어느 쪽을 보고해도 한쪽을 위반 → M2 첫 증거 인용에서 반드시 재작업.
2. **AC-ZRR-005(BLOCKING)의 판정 baseline이 구조적으로 달성 불가** — 실측: `git diff --numstat 294b4b6ab..HEAD -- internal/constitution/validator.go` → `17 / 0`(#1611의 retired 전처리. 보호 3함수·DRIFT 블록 밖이지만 전체 파일 diff는 0이 될 수 없음). run-phase가 매처를 한 줄도 안 건드려도 "diff 라인 0" 기대치 불달성.
3. **plan M2의 권고 구현이 그것이 기계화한다는 AC와 모순** — plan.md:108 `strings.Contains`(boolean) vs acceptance AC-ZRR-002 판정 문단 "적중 **횟수 자체**(boolean 아님)를 집계" + "정확히 1회 적중 / 빈 clause 실패". plan 문언대로 구현하면 AC-ZRR-002/003 미충족 구현이 정당화된다.

교차 모델 수렴이 이 판정을 독립 확인했다(아래 § 교차 모델 수렴).

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o 'REQ-ZRR-[0-9]+' spec.md | sort | uniq -c`: 001–015 각 1회 정의(복수 카운트는 §D.3 추적성 테이블 인용), 결번 0·중복 0. 제시 순서도 현재 전부 오름차순(iter2 N-2 해소 — § Regression Check).
- **[PASS] MP-2 GEARS** — 갱신된 3건 모두 형식 유지: REQ-ZRR-001 Ubiquitous(은퇴 면제는 부연 문장), REQ-ZRR-002 Ubiquitous("for every entry **including retired entries**" — 워크트리 amend 판), REQ-ZRR-007 Event-driven(When…shall + 은퇴 부연). 판정 대상은 spec.md 요구계층만(AC의 Given-When-Then은 검증계층).
- **[PASS] MP-3 YAML frontmatter** — 12 필드 전량 실측: `version: "0.5.0"`(quoted semver), `updated: 2026-08-25`(ISO), `priority: P1`, `phase: "v3.1.3 target"`(릴리즈 라벨 — 금지값 plan/run/sync 아님), `lifecycle: spec-anchored`, `tags` CSV. 거부 alias 0. 옵션 필드 `tier: M`·`era: V3R6`·`related_specs` 적법.
- **[N/A auto-pass] MP-4 언어 중립성** — 단일 언어(이 리포 자체 constitution/Go 도구) SPEC. 16언어 템플릿 배포 대상 아님.
- **[PASS] MP-5 D7 cross-SPEC** — 참조 `SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001`·`SPEC-V3R5-CONSTITUTION-DUAL-001` 모두 존재·`status: completed`(실측 grep). retired/superseded/archived 없음 → 조정 의무 없음.
- **[PASS auto] MP-6 D8 cross-platform** — `syscall` 문자열 spec/plan/acceptance 전부 0건(실측).
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' <SPEC dir>` rc=1, 0건. plan.md 존재(Tier M — research.md 없음은 정상).

---

## Category Scores (0.0–1.0, 러브릭 앵커 기준)

| 차원 | iter2 | iter3 | 밴드 | 근거 |
|-----------|-------|-------|------|------|
| Clarity | 0.95 | **0.85** | 0.75↔1.0 | C안 결정 기록이 3측정+계약 서술로 명확(§1.2:82–88행). 감점: §1.2에 "이 SPEC 은 아직 고르지 않았다"(74행)가 결정 기록 위에 잔류(시제 모순, F5), M2 종료조건 "임의의 한 엔트리"(plan:122) vs AC-ZRR-007 "명시 ID+무작위" 고정의 문언 갈림(C6) |
| Completeness | 0.95 | **0.90** | 0.75↔1.0 | HISTORY v0.5.0 갱신, §1.1 재측정(1ae6e5c36)·§C 재측정(9ba1e308d) 기록, §7 갭 정직 유지, Tier M 아티팩트 구성 준수. iter2 N-1/N-2 해소 확인 |
| Testability | 0.85 | **0.70** | 0.75 밴드 하단 이탈 | carve-out은 결정론적·카운트 가능(강점 — § 델타 판정 ②). 그러나 BLOCKING AC 3개의 판정 가능성 침투: AC-ZRR-007 두 문서 모순(F1), AC-ZRR-005 baseline 불가(C3, 실측 17/0), AC-ZRR-010 관측 갭(C5), plan M2 Contains 모순(C2) |
| Traceability | 0.95 | **0.85** | 0.75↔1.0 | 재측정치 전부 트리 `9ba1e308d` 귀속, RED `294b4b6ab` 역사 고정 명시(acceptance:52,68). 감점: §1.2 근거③의 도구 귀속 부정확(F4 — analyze.py는 retired 처리 없음, 실제 분리는 retired-vs-ac.py), baseline 낡음(C3), 분석 JSON 3개 동일 blob 배치(GLM4) |

**Aggregate = (0.85 + 0.90 + 0.70 + 0.85) / 4 = 0.825**

---

## 델타 판정 — 리드의 5개 질문

### ① AC 약화 여부 — **약화 아님 (범위 축소 정당; 단 Then 절 동기화 지연)**

- AC-ZRR-002/003의 101→97은 **기계적으로 식별 가능한 4개 명명 엔트리**(CONST-V3R2-021..024)의 범위 축소다. 유일 적중·0회·2회 이상·자기참조 4요구는 97에 대해 전부 유지되고(GREEN 문안 "1회 적중 97 / 0회 0 / 2회 이상 0 / 자기참조 0 / 은퇴 clause 면제 4" — acceptance:52), 면제 4건은 은폐가 아니라 **다섯째 버킷으로 가시 보고**된다. 빈 clause 변종 우회(iter2 부채 #1)는 유일 적중 요구로 여전히 차단.
- **"전부 은퇴로 분류" mutant 검사**: GREEN 값이 버킷 합계(97+0+0+0+4=101)까지 고정하므로, 전 엔트리를 은퇴로 분류하는 구현은 다섯째 버킷이 4가 아니게 되어 GREEN 불일치로 걸린다 — acceptance 매트릭스 GREEN 셀 자체가 방어다.
- AC-ZRR-004는 **101/101 그대로 + 은퇴 4건 포함 명시**(acceptance:78) — anchor 커버리지 축소 없음, 오히려 은퇴 anchor의 의미(후계 절 포인터 검증)까지 명문화돼 강화.
- 단, HEAD 기준 AC-ZRR-002/003의 **Given-When-Then 본문 Then은 "101건 전부" 그대로**(F3) — 같은 AC 안에 두 기준 공존. 워크트리 amend가 "live 97/97"로 고치는 것을 diff로 확인(해소 진행).

### ② carve-out 테스트 가능성 — **확인 (결정론 분류기 실재·고정)**

- `internal/constitution/retirement.go:33` `IsRetiredClause` — `[SUPERSEDED` **프리픽스** 매칭(레지스트리 자체 문서 55행이 "marker is a prefix, not a substring"으로 규정). `internal/constitution/validator.go:214` 실측 확인: `retired := !opts.Strict && IsRetiredClause(entry.Clause)`.
- 분류 경계도 테스트로 고정: `retirement_test.go:167` `TestIsRetiredClause` + `marker_regression_check_test.go:20`(shipped 레지스트리 분류 핀). 살아있는 clause가 본문에서 `[SUPERSEDED]`를 **언급만** 하는 경우(레지스트리 675행 사례 — 접두 아니므로 비은퇴)도 프리픽스 규칙이 배제한다.
- 다섯째 버킷 셈 가능성 실측: 레지스트리에서 `[SUPERSEDED` 프리픽스 clause는 253/261/269/277행 4건(CONST-V3R2-021..024) == 101−97.

### ③ 내부 정합성 — **97/101/4는 대체로 일치, 3개 표면 잔류 위반**

- 일치(실측 대조): spec §1.2 결정 기록 · REQ-ZRR-001 · acceptance §D 매트릭스 GREEN(002/003) · §D.1 GREEN 문단 · plan §C v0.5.0 재측정 · plan M1 종료조건(97/97 비은퇴 + anchor 101/101) · guard-scenario §6 P2/P4.
- **위반 3개 표면(F1)**: acceptance AC-ZRR-007(매트릭스 19행 "평가 엔트리 수 101×2" + 본문 110행 "그 값이 `101` 이다 — 두 미러 각각") / plan.md §H 부채②(166행 "clause 검사 101 / anchor 검사 101") / progress.md §E.1(16행 "clause/anchor 각각 101") — guard-scenario §6 P4 "clause 97 / anchor 101"과 갈림.
- §1.2 측정 근거 독립 재검증: `analysis-postmerge.json` 재계산 → total 101 / retired 4 / clause_fail 68(은퇴 4) / anchor_fail 17(은퇴 0) / 은퇴 anchor 전부 `#14-parallel-execution-safeguards`. CLAUDE.md 로컬↔템플릿 `cmp` byte-identical + `## 14. Parallel Execution Safeguards` 양측 153행. 레지스트리 미러 `diff -q` 동일. **spec §1.2의 3측정 전부 이 감사에서 재현됐다.**

### ④ mutant 저항 — **기존 방어 유지, 신규 구멍 2개**

- plan §G mutant 표는 델타가 건드리지 않았고 기존 방어(자기참조/빈 clause/`|| true`/부분 순회) 유지. 전부-은퇴 분류 mutant는 ①에서 확인한 대로 GREEN 버킷 합계가 잡는다.
- **F2(신규)**: guard-scenario §1 R2 = "random entry, ID recorded at run time"(43행) — **비은퇴 제약 없음**. 무작위 추첨이 은퇴 4건에 떨어지면(4/101) clause 변이는 면제 집합에 있어 가드가 초록 → 그 run은 시나리오 실패. §6의 "R1–R4 stand unchanged: `CONST-V3R2-004` is a non-retired entry, so **every mutation** stays inside the checked set"(130–131행)는 R2에 대해 **거짓**(R2 대상은 무작위). acceptance AC-ZRR-007 "무작위로 고른 1건"(111행) 동일 갭. → 리드가 지목한 지점이며, 확인 결과 제약이 없으므로 finding 확정.
- **C5(신규, codex)**: AC-ZRR-010 관측이 "깨진 엔트리 + SKIP=1" 상태만 커버(acceptance:142–144, plan:125). clean-tree + SKIP=1에서 `Skipped`를 성공으로 읽는 구현(리터럴 체크만 계속하는 형태)은 Given 조건에서는 실패를 반환해 AC를 통과한다 — REQ-ZRR-010("skip 자체를 실패로")의 관측 증거가 없다.

### ⑤ 추적성 — **SHA 귀속 성실, 귀속 오류 2건**

- 재측정치 전부 트리 `9ba1e308d` 병기(spec §1.2·plan §C·progress §F·scenario §6), RED는 `294b4b6ab`로 "역사적 측정치로 그대로 둔다" 명시(acceptance:52,68) — 규칙 문서 이동 시 수치 부정합을 예방하는 올바른 귀속.
- **F4**: §1.2 근거③·plan §C가 은퇴 분리 측정을 `analyze.py`로 귀속하나 analyze.py에는 retired 처리가 없다(grep 0건) — 분리는 `retired-vs-ac.py`(analysis-postmerge.json 입력)가 냈다. **값 자체는 이 감사의 독립 재계산으로 참 확인**(③).
- **C3**: AC-ZRR-005의 baseline `294b4b6ab`은 #1611(`bf6083f13`, validator.go +17라인) 이전 — 귀속 시점이 낡아 판정 불가(위 FAIL 사유 2).
- **GLM4**: `analysis-{devrepo,postmerge,repro}.json` 3개 `cmp` byte-identical(실측). 세 트리에서 레지스트리·규칙 문서가 동일해 같은 출력이 나온 것이 자연스러우므로 값의 진실성은 무효화되지 않으나, 파일 배치가 독립 측정처럼 읽히는 것은 개선 여지.

---

## Defects Found (구조화 결함 목록 — 수정 루트)

**D1. F1 — acceptance.md:19,110 / plan.md:166 / progress.md:16** — AC-ZRR-007 평가-수 단언("101×2"/"101")이 guard-scenario §6 P4("clause 97 / anchor 101")와 모순 — Severity: **critical** — Class: **blocking** — 수정: 4개 표면 전부 "clause 97 / anchor 101, 두 미러 각각, 분리 보고"로 갱신. §H 부채②·progress §E.1 표현 동반 갱신.

**D2. F2 — guard-failure-scenario.md:43,130 / acceptance.md:111** — R2 무작위 변이에 비은퇴 제약 없음(4/101 확률로 면제 집합 추첨 → run 무효) + §6 "R1–R4 stand unchanged… every mutation stays inside the checked set" 서술이 R2에 대해 부정확 — Severity: **major** — Class: **blocking** — 수정: R2를 "random **non-retired** entry"로 제한, AC-ZRR-007 "무작위로 고른 1건"에 (비은퇴) 병기, §6 문구를 R1/R3/R4로 한정.

**D3. C2 — plan.md:108** — M2 권고 구현 `strings.Contains`(boolean)가 AC-ZRR-002/003 판정 문단("적중 횟수 자체(boolean 아님)"+유일 적중+빈 clause 실패)과 모순 — Severity: **major** — Class: **blocking** — 수정: "엔트리별 `strings.Count(rawFileContent, clause)` 를 세어 1회 적중/0회/2회 이상/은퇴 면제 버킷을 출력"으로 문안 교체. 빈 clause 실패 조항과의 정합도 명시.

**D4. C3 — acceptance.md:86–92** — AC-ZRR-005(BLOCKING) 판정 baseline `294b4b6ab`이 #1611 이전이라 `git diff 294b4b6ab..HEAD -- internal/constitution/validator.go`가 run-phase 무변경이어도 17라인(실측) — Severity: **major** — Class: **blocking** — 수정: baseline을 #1611 이후 병합 기준(`bf6083f13` 또는 본 워크트리 base)으로 갱신하거나, 판정을 보호 3함수+DRIFT 블록의 추출 diff로 한정(#1611의 17라인은 해당 범위 밖임을 이 감사가 실측 확인 — retirement 전처리·RetiredCount 필드).

**D5. C5 — acceptance.md:142–144 / plan.md:125** — AC-ZRR-010 관측이 변이 주입 상태만 커버 — clean-tree + `MOAI_CONSTITUTION_SKIP_VALIDATE=1`에서의 관측이 없어 skip-성공-무시 구현이 통과 — Severity: **major** — Class: **blocking** — 수정: AC-ZRR-010에 "깨끗한 트리 + SKIP=1 → 가드가 '검증 건너뜀'을 이유로 실패" 관측 1회 추가(변이 주입 관측과 별개).

**D6. C6 — plan.md:119–126** — M2 종료조건이 AC-ZRR-007과 어긋남: "임의의 한 엔트리" 변이(vs 명시 ID+무작위 고정), CI job 결론 관측(`gh pr checks`) 결락 — M2 종료 ≠ AC-ZRR-007 충족으로 run이 끝날 수 있음 — Severity: **minor** — Class: **blocking** — 수정: M2 종료조건 2번을 AC-ZRR-007 문언과 동일시하고 CI 관측 항목을 추가하거나, "판정은 guard-failure-scenario.md §1–§3 을 따른다"로 normative 지정.

**D7. F5 — spec.md:74** — §1.2에 "가능한 해소는 셋이며, **이 SPEC 은 아직 고르지 않았다**"가 결정 기록(82행 "C안 채택")과 같은 절에 잔류 — 시제 모순 — Severity: **minor** — Class: **blocking**(문서 정합성) — 수정: 해당 문장을 "v0.5.0에서 C안으로 확정했다(아래)"로 교체. (amende 진행 여부 미확인)

**D8. F4 — spec.md:86 / plan.md:48** — 은퇴 분리 측정의 도구 귀속이 `analyze.py`로 부정확(retired 처리 없음; 실제는 `retired-vs-ac.py`) — Severity: **minor** — Class: **optional** — 수정: 귀속을 `retired-vs-ac.py`(입력 analysis-postmerge.json)로 정정. 값은 이 감사에서 독립 재검증 완료.

**D9. F3 — acceptance.md:45,62** — (HEAD 기준) AC-ZRR-002/003 Then 절 "101건 전부"가 같은 AC의 GREEN 문단(97)과 이중 기준 — Severity: **minor** — Class: **blocking** — **상태: 워크트리 amend에서 "live 97/97"로 갱신되는 것을 diff로 확인(해소 진행)** — 커밋 시 최종 확인 필요.

**D10. GLM4 — .moai/reports/t232/analysis-*.json 3개** — byte-identical blob이 서로 다른 측정인 것처럼 배치 — Severity: **minor** — Class: **optional** — 수정: "세 트리에서 분석 입력(레지스트리·규칙 문서)이 동일해 같은 출력"임을 주석하는 것이 정직.

**범위 밖 기록(VERDICT 무관 — 보조 스크립트 결함, codex/glm 발견, 이 감사에서 존재만 확인)**:
- watch-review.sh:22,25 — `TIMEOUT` 인자가 검증 없이 산술 평가 `$(( SECONDS + TIMEOUT ))`에 들어감(배열 첨자 확장으로 임의 명령 실행 가능 — codex 재현). 로컬 전용 보고 보조 도구라 SPEC 범위 밖이나 **run-phase에서 이 스크립트를 재사용한다면 정수 검증(`[[ $TIMEOUT =~ ^[0-9]+$ ]]`) 추가 권장**.
- divergence.py:9 — `diff` rc 2(파일 부재)를 DIFF와 동일 취급 — SAME/DIFF 판정 신뢰성 흠. 로컬 보조 도구.
- analyze.py:24 — 값 내 따옴표 파싱 아티팩트(CONST-V3R5-09, findings.md 자각 사항 — spec §2.1에 문서화돼 있음).

---

## Regression Check (iter2 이월 항목)

| iter2 항목 | 판정 | 근거 |
|---|---|---|
| 부채 #1 빈 clause→유일 적중 | **RESOLVED**(v0.4.0) | AC-ZRR-002 "정확히 1회 적중" 명시 + plan M1 사다리. 이번 델타에서 유지 확인(97 기준으로도 그대로) |
| 부채 #2 CONST-V3R2-004 근접 재지정 | **UNRESOLVED (설계대로)** | sync 리뷰어 판정 이관 — run/sync 단계 항목. plan §H 유지. 결함 아님 |
| 부채 #3 평가 수 이중 카운트 | **WORSENED → D1** | §H ②가 "각각 101"로 잔류한 채 option C(97/101)와 갈림 — 이번 델타가 만든 모순(D1)으로 승격 |
| N-1 §D8 붕 뜬 참조 | **RESOLVED** | plan §D8이 현재 `REQ-ZRR-012`를 가리키고 REQ-ZRR-012가 실제 slug 선언 의무(spec:198) — 정확 |
| N-2 REQ-015 제시 순서 | **RESOLVED** | 현재 본문 001→015 전부 오름차순(180–204행 실측) |

---

## 교차 모델 수렴 (audit_multi, multi 모드: claude+codex required / glm advisory)

- **claude(required): pass** — 본 보고서의 판정 앵커
- **codex(required): fail** — 6 findings. 이 감사에서 검증한 결과: #1(카운트 모순)=수용·D1로 확장(§H·progress 포함), #2(Contains)=수용·D3, #3(AC-005 baseline)=수용·D4(실측 17/0), #5(AC-010 관측)=수용·D5, #6(M2 종료조건)=수용·D6, #4(watch-review.sh)=범위 밖 기록
- **glm(advisory): pass** — 4 findings: 97/101 정합성(=D1 수렴), divergence.py·analyze.py 파서(범위 밖 기록), JSON 3개 동일(D10)
- **overall_verdict: fail / disagreement_flag: true**
- **residual_risk_note (verbatim)**: `cross-model disagreement (advisory, NOT a block): pass=[claude(required), glm(advisory)] fail=[codex(required)]`

필독: codex의 required-FAIL은 이 감사가 독립 검증 후 5/6 수용함으로써 단순 불일치가 아니라 **확정 결함 집합**이 됐다. 본 FAIL은 수렴 결과가 아니라 검증된 결함에 근거한 이 감사 자체의 판정이며, 수렴이 이를 보강한다.

---

## Recommendation

**FAIL — run-phase 진입 불가. 아래 1차 수정 후 커밋하면 artifact-hash가 다시 바뀌므로 게이트는 delta 재실행(사실상 iter4)으로 재판정한다. Tier M 재감사 한도(2)는 iter2에서 소진됐으므로 이번 재실행은 리드(game re-run) 판단 아래 진행된 것으로 기록한다.**

1. **D1 우선** — AC-ZRR-007 매트릭스·본문을 "clause 97 / anchor 101, 두 미러 각각 분리 보고"로, plan §H ②·progress §E.1 동반 갱신. guard-scenario §6 P4와 한 문장으로 일치시킬 것.
2. **D2** — R2 "random non-retired entry" 제약 + AC-ZRR-007 무작위 대상 (비은퇴) 병기 + §6 stand-unchanged 문구 R1/R3/R4 한정.
3. **D3** — plan M2 리터럴 체크를 `strings.Count` 기반 횟수 버킷 출력으로 교체.
4. **D4** — AC-ZRR-005 baseline 갱신(#1611 이후) 또는 판정 범위를 보호 블록 추출 diff로 한정.
5. **D5** — AC-ZRR-010에 clean-tree + SKIP 관측 추가.
6. D6–D10은 같은 수정 회차에서 함께(전부 문안 수준).
7. 현재 워크트리의 미커밋 amend(spec/acceptance)는 F3 계열을 이미 해소 중 — 위 항목을 같은 커밋에 반영한 뒤 **커밋 시점의 최종 판을 게이트에 제출**할 것.

다행인 점: 이번 델타의 본체(범위 축소의 정당성·결정론 분류기·측정 재현)는 전부 이 감사에서 독립 재검증됐고 침해 없음. 결함은 전부 "확정된 C안 계약을 나머지 규범 표면에 동기화하는 작업이 덜 끝난 것"이다 — amend 진행 방향이 이미 옳다.

---

## 이 감사가 하지 않은 것 (Gaps)

- **AC를 실행하지 않았다** — 가드는 아직 없다(구현 전). 판정은 전부 문안·계약의 정합성과 mutant 폐쇄력.
- **amend 최종판을 판정하지 않았다** — 감사 중 워크트리가 계속 이동했다. D9는 해소 진행으로 기록했으나 커밋된 최종판에 대한 재확인은 게이트 재실행의 몫.
- **watch-review.sh 인젝션 재현은 codex 보고에 의존** — 이 감사는 25행 산술 평가 구조만 육안 확인(위 범위 밖 기록). 재현 명령은 직접 실행하지 않았다.
- **fresh-init AC-ZRR-001/013 RED는 선행 감사 관측을 승계** — 재실행하지 않음(iter1/2와 동일).
- **교차 모델 1차 호출 실패** — 첫 `audit_multi` 호출이 anchor 누락으로 거부됐고 재호출로 성공. 백엔드 판정은 2차 호출 결과 기준.
