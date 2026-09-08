# SPEC Review Report: SPEC-DOCS-LOCALE-PARITY-REPAIR-001 (card t538)

- Iteration: **2/2** (Tier M ceiling — 최종 iteration)
- Verdict: **PASS**
- Overall Score: **0.95** (조화평균) — Tier M 문턱 0.80 통과. iter1 0.71 → iter2 0.95 (상승 — STOP 트리거 없음)
- 재감사 범위: iter1 결함 delta (D1-D6, F1-F6) + 경량 회귀 스윕 — 전수 재감사 아님 (Retry Loop Contract delta scope)
- 감사 도구: 전 수치 `/usr/bin/grep`·`wc -l` 실측, 저자 주장 무신뢰 재측정

## Must-Pass Results (회귀 포함)

- **[PASS] MP-1 REQ 연속성** — REQ-001~013 유지, 재번호 없음 (spec.md:86-136).
- **[PASS] MP-2 GEARS (REQ 계층)** — 13건 형식 유지. F3 각하 기록(HISTORY:34)은 iter1 MP-2 통과 판정과 정합 — 복합절 유지가 판정과 모순되지 않는다.
- **[PASS] MP-3 프론트매터** — 12 필드 유지, version `"0.1.0"` → `"0.2.0"` 인용 semver 정상 (spec.md:4).
- **[N/A] MP-4** — 전 회차와 동일 (프로그래밍-언어 도구 SPEC 아님).
- **[PASS] MP-5 D7** — 참조 3건 상태 불변 (`completed` ×3, 재확인 불요 — 회귀 스윕에서 SPEC 참조 목록 변동 없음 확인).
- **[PASS] MP-6 D8** — `syscall` 0건 불변.
- **[PASS] MP-7** — plan.md `[NEEDS CLARIFICATION` 0건 불변. research.md 부재 (Tier M — N/A 사유 동일).

## Per-Fix 판정 표 (기계 재측정 — 전부 이번 실행, 트리 `babbcc017`)

| 항목 | 수정 내용 (인용) | 재측정 | 판정 |
|---|---|---|---|
| **D1** | AC-006 이중 셀: 블록-한정 행두 앵커 `/usr/bin/grep -cE '^moai doctor (permission\|sandbox)'` 0→2 + 전체 파일 2→4, 표행 불변 논리 명시 (acceptance.md:15, :52-56) | 앵커: ko **2** / en **2** / ja **0** / zh **0** — 전체 파일: ko 4 / en 4 / ja **2** / zh **2**. ja·zh 의 앵커 0 vs 전체 2 = 표행 2건이 정확히 분리됨 — AC 가 기술한 산술 두 cell 모두 실측 일치 | **PASS** |
| **D2** | AC-005 에 호스트-OS 문단 전용 토큰 grep 병합 (acceptance.md:46-50) + Given 이 헤딩-패리티 무감각 사실을 정직하게 기록 | ko `호스트 OS 규칙` = **1** / ja `ホスト OS ルール` = **1** (baseline 유지 — 무변경 증명 앵커 정상), en `host OS rule` = **0** (RED-now cell 정확) | **PASS** |
| **D3** | plan.md AC 색인 재정렬 (:41 §E 6라벨, :62 M1 → AC-001~005+AC-009, :69 M2 → AC-006+AC-007, :78 M3 → AC-009~011) | 전 라벨 acceptance.md 실번호와 대조 — 6라벨 전부 일치, M1/M2/M3 인용 범위 전부 실재하는 AC 를 가리킴 | **PASS** |
| **D4** | REQ-012 zh 토큰 `原生桌面` 계열 + 근거 명시 (spec.md:132), edge 1·5 갱신 (acceptance.md:88, :92) | `原生桌面` = **1행** (zh:170) / `桌面原生` = **0** — 페이지 기존 용어와 정합 | **PASS** |
| **D5** | 「RED-now AC 4건」→「6건」 (acceptance.md:104) | 6건 = 나열 6 ID = progress.md `red_now_acs` 6 (:22) — 3면 일치 | **PASS** |
| **D6** | 배치 경로 `[Unreleased] → ### Added` 최상단으로 3면 정정 + `### Docs` 부재 사실 명시 (spec.md:124, acceptance.md:84, plan.md:77) | CHANGELOG 실구조 (:8 → :10, t535 항목 :12 직하)와 일치 — 수리 커밋이 CHANGELOG 를 건드리지 않아 실측 불변 | **PASS** |
| **F1** | AC-010 `--quiet` 제거 → WARN/ERROR 행 계수 + exit 코드 (acceptance.md:78-79, 사유 기록) | 명령-기대 정합 (계수는 카운트로, exit 별도 — edge 4 와 정합) | **PASS** |
| **F2** | §C G1 서술 정밀화 — 「지연 동안 생존」 (spec.md:70) + 측정 파일 정정 부기 (plan-phase.md:70) | 두 표면 모두 확인 — verbatim-문장 계대 과장 제거, 결함 본질 서술 유지 | **PASS** |
| **F3** | 각하 기록 (HISTORY:34) | 기록 존재, MP-2 판정과 정합 | **PASS (각하 유효)** |
| **F4** | 각하 기록 (HISTORY:35) + 측정 파일 정정 부기 (plan-phase.md:69) | **판정 뒤집힘 — 감사자 측 오귀속** (아래 감사자 정정 절 참조). 초판 64행이 옳았음: 워크트리(`bce6d7e08` 내용) `wc -l` = **64 ×4** 재측정, iter1 의 63 은 primary 체크아웃(main 트리)에서 나온 값 | **PASS (각하 유효 — 초판이 정확)** |
| **F5** | 각하 기록 + plan.md §G 경고 불릿 추가 (plan.md:89 — G4 파생자가 `.moai/reports/...` 서술 스타일 모방 금지) | 기록 + 대응 절차 모두 존재 | **PASS (각하 유효)** |
| **F6** | 각하 기록 (HISTORY:37) — 전제는 배차 지시문 기재, 감사자 find-0 실측 명시, run 진입 시 리드가 러너 실물 확인 | 기록 존재 — 미검증 전제를 미검증이라 명시한 정직한 처리 | **PASS (각하 유효)** |

## zh 지연-토큰 절차 판정 (코디네이터 지정 판단 사항)

**run 페이즈에서 기계 점검 가능 — 홀(hole) 아님.** 절차는 「파생 → 토큰 확정 → 기록 → grep ≥1」 4단으로 문서화돼 있고(acceptance.md:49), 자기-참조 공허 패턴이 아니다. 토큰 선택이 두 독립 앵커로 구속되기 때문이다: (1) REQ-012·edge 5 가 `原生桌面 계열` 패밀리를 zh:170 기존 용어 실측으로 고정 — 파생자가 임의 토큰을 못 만든다, (2) 확정 토큰의 기록 의무 — sync-audit 이 기록된 토큰 vs 페이지 내용을 재검증 가능. 잔여: 확정 구구(宿主/主机 OS 규칙 대응 어구)의 번역 적절성에 대한 외부 오라클은 없다 — 존재·패밀리 정합·기록 3요건만 판정 가능하다. 이 잔여는 plan 페이즈에서 토큰을 선고정했던 iter1 D4 의 실수를 되풀이하지 않기 위한 구조상 필연이며, 결함이 아닌 한계로 기록한다.

## 감사자 정정 (iter1 F4 — 기록 바로잡음)

iter1 보고서의 F4 「plan-phase.md:12 가 64행 주장, 실측 63행」은 **감사자 측 오귀속이었다**. iter1 의 63 은 primary 체크아웃(`main` 트리) 경로에서 측정된 값이고, 감사 대상 트리인 `bce6d7e08`(워크트리)에서는 `wc -l` = 64 ×4 가 실측된다(본 회차 직접 재측정 — worktree 64×4/256 total vs primary ko 63). 저자의 정정 부기(plan-phase.md:67-70)가 옳게 내 값을 반박했고, HISTORY:35 의 각하는 유효하다. VCI §2 baseline 귀속 규정을 감사자가 위반한 사례로 본 기록에 남긴다.

## Regression Sweep (경량)

- **회귀 없음 확인**: 모든 RED-now baseline 재고정 — en/zh deferral 1/1, bold-괄호 ko/en/ja/zh = 1/0/1/1, SVG0 ×4 = 0, `原生桌面` 1/`桌面原生` 0 — iter1 값과 동일.
- **변경 오염 없음**: `babbcc017` 은 SPEC 아티팩트 4종만(acceptance 29행, plan 11행, progress 1행, spec 21행 — `git show --stat` 실측), `c2f169791` 은 plan-phase.md +5행. `git diff --stat bce6d7e08 -- docs-site CHANGELOG.md` 빈 출력 — 문서 표면 무변경 (plan 페이즈 편집 금지 OOS 준수).
- **Must-Pass 보존**: §B.1 양 dated 정정 (spec.md:47-48), §B.2 시대 귀속 [HARD] (:50-52), G4 1행 근거 (plan.md:73), REQ-013 러너 금지 (:136), OOS H3 ×4 — 전부 불변.
- **장부 정합**: AC 인벤토리 11건 불변(번호 변경 없음 — 라벨·값만 정정), 마일스톤 M1/M2/M3 그룹핑 유지, progress.md §E.1-§E.4 구조 intact + Phase Log iter1 수리 행 기록 (:44).

## Category Scores (0.0-1.0, 러브릭 앵커)

| Dimension | Score | 근거 |
|-----------|-------|------|
| Clarity | 0.90 | REQ-012 가 페이지 자체 용어와 자기-정합(근거 실측 인용, spec.md:132), REQ-010 실존 경로(:124). 잔여: REQ-005 「mo-e2e.md 가 아니라」 기법(:104) — 회의적 독자가 일관되게 풀 수 있는 경미 모호 |
| Completeness | 1.0 | 전 절 + HISTORY 가 감사 인용·결함별 처분까지 기록(:25-37) — 아티팩트 셋 완비 |
| Testability | 0.90 | 11 AC 전부 기술된 그대로 이진 판정 가능(D1 이중 셀 산술 실증, D2 앵커 실측). 잔여: zh 토큰 외부 오라클 부재(구조상 한계, 기록됨), M3 최종-게이트 인용 범위(AC-009~011)가 산문의 「4-로케일 패리티」(AC-005/008, 각 자기 마일스톤에서 검증)를 포함하지 않는 표기 여유 — 관측, 결함 아님 |
| Traceability | 1.0 | REQ-005 가 AC-005 호스트-OS 토큰으로 커버 — 13 REQ ↔ 11 AC 전부 직접 매핑 |

**Overall: 4/(1/0.90 + 1/1.0 + 1/0.90 + 1/1.0) = 0.947 → 0.95 ≥ 0.80 (Tier M) → PASS**

## Defects Found

차단 결함 없음. 선택 관측 2건(수리 불요):

- **F-obs-1** — plan.md:78 — M3 최종-게이트 산문의 「4-로케일 패리티」에 해당하는 AC-005·AC-008 이 인용 범위 「AC-009~AC-011」 밖이다. 각 AC 는 소속 마일스톤(M1/M3)에서 검증되므로 누락 아님 — 표기 여유만. — Class: optional.
- **F-obs-2** — zh 호스트-OS 토큰의 번역-적절성 외부 오라클 부재 (위 zh 절차 판정의 잔여). run 페이즈 확정 토큰 기록으로 정밀화됨. — Class: optional.

## Recommendation

plan 페이즈 종료. run 페이즈 진행 권고:

1. 본 verdict(PASS 0.95)는 Tier M 문턱 통과 — run 게이트 skip-eligible 3조건 중 2조건(판정·점수) 충족. 3조건(아티팩트 해시 불변)은 이후 SPEC 아티팩트 무변경 시 성립 — 오케스트레이터가 run 위임 프롬프트 Section A 에 skip 결정과 3조건 충족 여부를 기록할 것 (spec-workflow.md § skip policy).
2. Implementation Kickoff Approval (HUMAN GATE) 은 skip 과 무관하게 계속 필수 — plan.md §C 마지막 미체크 항목.
3. REQ-013 각하 기록대로 run 진입 시 리드가 `hns-oss-docs-run` 러너 실물 확인으로 전제를 확정할 것.
4. run 페이즈에서 zh 호스트-OS 토큰 확정 시 acceptance.md §D.5 의 절차대로 기록 — sync-audit 재검증 가능하게.

## Evidence-Bearing Report (5-섹션)

**Claim**: iter1 결함 6건(D1-D6) 전부 + F1·F2 수리를 기계 재측정으로 확인했고, F3-F6 각하 기록이 HISTORY 에 존재하며, 회귀 없음. 반면 iter1 F4 는 감사자 측 트리-오귀속으로 판정 뒤집혔다(초판 64행이 옳음). 최종 판정 PASS 0.95 — Tier M 문턱 통과.

**Evidence**: 핵심 원시 출력 — `/usr/bin/grep -cE '^moai doctor (permission|sandbox)'` doctor.md ×4 = **ko 2 / en 2 / ja 0 / zh 0**; `/usr/bin/grep -c "moai doctor permission\|moai doctor sandbox"` ×4 = **4/4/2/2** (AC-006 양 셀 산술 실증); `호스트 OS 규칙` ko = **1**, `ホスト OS ルール` ja = **1**, `host OS rule` en = **0** (AC-005 토큰 셀 실증); `原生桌面` zh = **1행** / `桌面原生` = **0** (D4); deferral 1/1, bold-괄호 1/0/1/1, SVG0 0×4 (회귀 없음); `wc -l` codex-dual-harness worktree = **64×4 (256 total)** vs primary ko = **63** (F4 판정 뒤집힘 실증); `git show --stat babbcc017` = SPEC 4파일, `c2f169791` = plan-phase.md 1파일 +5행; `git diff --stat bce6d7e08 -- docs-site CHANGELOG.md` = 빈 출력.

**Baseline-attribution**: (this run, this tree) — defect-delta 재감사, worktree `.claude/worktrees/t538` @ **`babbcc017`** (브랜치 `WT-docs-v313-locales`), 이전 감사 기준점 `86c8023b6`(iter1 보고서) — 수리 커밋 `babbcc017` + 측정 부기 `c2f169791` 상대. 문서 표면은 base `bce6d7e08` 내용과 동일함을 diff-빈 출력으로 소속 확인. F4 비교 측정의 primary 경로 값은 `main` 트리 소속임을 명시한다(감사 대상 아님 — 대조 전용).

**Gaps**: (1) `hns-oss-docs-run` 러너 본체는 이 세션에서도 미도달 — 전제 확정은 각하 기록대로 run 진입 시 리드 몫으로 이관. (2) hugo 빌드 미실행 — AC-010 은 run 페이즈 산출물 판정. (3) zh 호스트-OS 토큰은 파생 시점까지 확정 불가(문서화된 절차로 이관 — 판정은 run+sync 페이즈). (4) D7 3건의 `status:` 재판독은 생략 — 수리 커밋이 참조 목록을 건드리지 않았음을 파일 범위 실측으로 대체했다(재판독 생략 사유).

**Residual-risk**: (1) zh 토큰 번역 적절성 외부 오라클 부재 — 절차 준수만 검증 가능. (2) skip-eligible 캐시는 아티팩트 해시에 민감 — 이 보고서 이후 SPEC 아티팩트를 한 글자라도 고치면 재감사가 강제된다(해시 무효화). (3) F-obs-1 의 표기 여유는 run 페이즈 M3 게이트에서 AC-005/008 누계 인용으로 자연 해소 가능 — 미처리 시 sync-audit 재지적 여지.
