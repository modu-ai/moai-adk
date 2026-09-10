# t572 Plan-Audit — SPEC-OWNERSHIP-SILENCE-001

- 감사자: plan-auditor (독립 적대적, 단일 모델 — dispatch 지시에 따라 audit_multi 미사용)
- 일자: 2026-09-08
- 측정 트리: `.claude/worktrees/t572` @ `3ac58b5a1` (branch `WT-ownership-lint-silent`, origin/develop tip 기반)
- 기준: Tier M PASS 임계 0.80
- 성격: 구현 전 감사 — 본 SPEC 대상 코드는 아직 존재하지 않는다(워크트리 상태: `.moai/specs/SPEC-OWNERSHIP-SILENCE-001/`와 `.moai/reports/t572/`만 untracked, 그 외 clean).

---

## 판정 (Verdict)

**PASS** — 종합 점수 **0.94** (임계 0.80 통과). must-fix 0건, should-fix 1건, advisory 4건.

**PASS-with-debt 1건 (F1)**: AC-OWN-004의 RED 관측 운반체가 acceptance 텍스트에 명명돼 있지 않다. 왜 수용 가능한가: plan §G m2 뮤턴트 + AC-OWN-005의 4단계 변이 절차(주입 → 스위트 판정 → verbatim FAIL 기록 → 원복 확인)가 이 AC의 적색을 4요소 규율대로 관측하도록 **이미 의무화**돼 있어, 결함의 실질은 문서상 관측점 기재 위치의 부재다. 구현 설계나 검증 설계의 결함이 아니다.

---

## 재검증된 증거 (이번 실행, 이 트리)

아래는 전부 이번 감사 실행에서 직접 관측한 출력이다 (baseline-attribution: this run, tree `3ac58b5a1`).

**코드 좌표 — 전수 일치:**

| SPEC 주장 | 실측 | 판정 |
|---|---|---|
| strict 승급 Warning 한정 (`lint.go:62`) | `if r.Strict && f.Severity == SeverityWarning && !f.Advisory {` — verbatim | 일치 |
| 룰 등록 (`lint.go:147`) | `&OwnershipTransitionRule{}` | 일치 |
| ID 패턴 (`lint.go:1131`) | `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` — 다중 세그먼트 합법 | 일치 |
| REQ 패턴 (`lint.go:651`) | `^REQ-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)?-\d{3}(?:-\d{3})?$` — REQ-OWN-001..010 전부 적합 | 일치 |
| 피의자 분기 (`lint_ownership.go:409-416`) | 주석 :409-413(M4 AC-LSG-004 가드) + `if rec.AuthoredByAgent == "" { return nil }` :414-416 | 일치 |
| 보존 지점 | `:400-402`(rec==nil), `:405-407`(ownerNone), `:419-422`(미인식 행위자) — 각각 조용한 nil | 일치 |
| 형제 발견 | `:375-385` Skipped(Info), `:388-399` Unreachable(Info), `:424-` Invalid(Warning) | 일치 |
| (b) 기각 근거 | `:108-111` NOTE verbatim: "subject-prefix 분류는 더 이상 production Check() 경로에서 사용되지 않는다… 본 함수는 …회귀 테스트를 위해 유지된다"; `commitOwnerKind`(:112) 실재 | 일치 |
| trailer enum (`:160-172`) | manager-spec/manager-develop/manager-docs/orchestrator-direct 만 인식, **manager-git 미포함** → ownerNone | 일치 |
| 낡은 주석 | `:179` "…subject prefix 분류 경로로 fallback한다" + `:367` "2. fallback: …" — **둘 다 실제로 거짓**(:414는 fallback 없이 nil 반환) — REQ-OWN-009의 stale 판정 자체가 검증됨 | 일치 |
| 창 크기 (`drift.go:302`) | `const gitLogWindowSize = 50` | 일치 |
| CI (`spec-lint.yml:58`) | `go run ./cmd/moai spec lint --strict`, `fetch-depth: 0`(:40), paths 트리거 `.moai/specs/**`+`internal/spec/**`, push 트리거 main+develop | 일치 |
| advisory 이디엄 (`lint.go:1056`) | CoverageRule "Severity is `warning` with `Advisory: true` set at the EMISSION SITE" | 일치 |
| 테스트 | `trailer_absent_silent_skip` :550-576, fake 주입 `getOwnershipTransitionRunner` :192 | 일치 |

**측정값 재현 —**

```
$ git log --all --format='%(trailers:key=Authored-By-Agent,valueonly)' | /usr/bin/grep -c '[^[:space:]]'
104
$ git log --all --format='...' --date=short | awk ... (최신 3건)
a20fba05f	2026-09-03	docs(SPEC-LANE-PUSH-DOC-001): M2 run-phase evidence export (t463)	manager-develop
a30edfe098	2026-09-03	fix(SPEC-LANE-PUSH-DOC-001): M1 lane push actor sentence repair (t463)	manager-develop
b1bcce4f4	2026-09-01	fix(SPEC-STATUS-TRANSITION-VALIDITY-001): memoize the status-transition history walk (t376)	manager-develop
```

- **104건·최신 2026-09-03 — baseline과 정확히 일치.** §1.1 배차 전제 정정과 §3 (c) 채택의 재료는 유효하다. 단, 총 커밋 분모는 재측정 시점 **11,936**으로 이동(공유 저장소의 다른 레인 병합으로 ref가 진행 — F2 참조).
- 쌍둥이 규칙 문서: `cmp -s` exit 0, 양측 256행 — plan §C.3의 "256행 동일" 재확인.
- "subject prefix" 문구: 양 사본 각 1건 실재(`spec-frontmatter-schema.md:198`) — 문서 드리프트 주장 실재이자 AC-OWN-006(3) 부재-가드의 **RED-now 관측 가능** 상태(현재 grep -c = 1 → 목표 0).
- `internal/template/templates/.codex/rules` 디렉터리 부재 — "규칙 문서에는 .codex 쌍둥이가 없다, agents-emit 불필요"(plan §D.4) 주장 **참**.
- `manager-develop.md:185`(로컬)·`:186`(미러)·toml 1히트 — §2.2 불접촉 대상의 좌표 실재. 해당 문장은 트레일러-존재-불일치 경로가 살아있는 한 (c) 채택 후에도 참 — SPEC의 판단 검증됨.
- 무음-단언 형제 인벤토리: 테스트 파일 내 `AuthoredByAgent: ""` 리터럴 1건(:556), 테이블 행 12건 전부 비어있지 않은 트레일러 — REQ-OWN-006의 "그 외 동일 단언 전부" 스코프가 정확히 1개 테스트로 확정. 스윕 누락 위험 없음.

---

## 차원별 점수

| 차원 | 점수 | 근거 요약 |
|---|---|---|
| 1. Frontmatter·GEARS 형식 | 0.97 | 12 정규 필드 전부(spec.md:2-13), ID·REQ 패턴 코드 대조 적합, `phase: "v3.2.0"` 릴리스 라벨(금지 단계명 아님), tier M ↔ 3 산출물 정확(디렉터에 design/research 없음), H3 "Out of Scope —" 관례 준수 |
| 2. 증거 건전성 (§3 판정) | 0.95 | 3개 처분 전부 측정에서 도출·재검증됨. (b) 기각이 코드의 기록된 결정(:108-111)에 근거. §1.1 전제 정정 기록됨. (F2: 고정 안 된 분모 수치) |
| 3. 검증 완결성 (two-cell + 뮤턴트) | 0.85 | AC-OWN-001 모범적(4요소 + wrong-reason 배제 + RED 사유 명시). §G 뮤턴트 표 + 생존 메모 양호. **F1: AC-OWN-004 RED-now 셀 부재** |
| 4. 스코프 규율·Non-goals | 0.98 | 보존 3지점·형제 발견의 좌표 실측 확인, manager-develop 불접촉 논리 검증, agents-emit 불필요 참, 낡은 주석의 stale 판정 자체가 검증됨, 템플릿 중립 의무 명기 |
| 5. 계층 간 일관성 | 0.90 | REQ-OWN-010 ↔ plan §D.6 ↔ acceptance §D.2.2 행위자 3값·관측점 일치. (F3: 등급 기록 위치) |
| 6. AC 관측 가능성 | 0.92 | 8개 AC 전부 커맨드+출력 형태로 관측 가능. 메시지 5요소 단언 + m3 픽스처 의존이 수록됨. (F4·F5: 표기 미세 불일치) |
| 7. 검증 스코핑·규율 | 0.97 | AC-OWN-007 스코프드 패키지 한정, 전체 스위트 금지를 명시("부재가 아니라 의도다"), 상속 적색 분리 절차(§C.4 + AC-OWN-007 축 분리) 존재, 뮤턴트 원복 확인 의무화 |
| 8. 산출물 위생 | 0.96 | plan.md·acceptance.md 프론트매터 없음(상태축 무상태성 준수), progress.md §E.1 신호 존재(pending-plan-audit — 판정 전 상태로 정확), 12 필드 + tier/era 옵션만 사용 |

**종합: 0.94** (산술 평균). 임계 0.80 통과.

---

## Findings

### F1 (should-fix) — AC-OWN-004가 release-blocking 등급인데 RED-now 셀이 없다

- **위치**: acceptance.md:15(매트릭스 등급 "High (release-blocking)"), acceptance.md:69-80(AC 상세 — green-path 판정만 존재)
- **내용**: verification-completeness §2는 채택된 AC에 RED-now 셀 + green-path 셀의 쌍을 요구하고, §2.1은 RED-now가 관측 불가능한 AC에 대해 release-blocking 자격 상실 → regression-guard 재분류를 정한다. AC-OWN-004의 주제(새 Info 발견의 strict 무영향)는 구현 전 트리에 존재하지 않는 코드이므로 pre-fix 4요소 RED가 구성 불가능하다 — 컴파일 실패 RED는 wrong-reason red다. acceptance는 이 관측이 **어디서**(뮤턴트 주입 시점) 어떤 형태로 기록되는지 명명하지 않는다.
- **왜 must-fix가 아닌가**: plan §G m2(등급 Warning 승급 뮤턴트 → strict 테스트가 `HasErrors()==true` 뒤집힘을 잡음)가 이 AC의 유일한 유효 판정자로 지정돼 있고, AC-OWN-005의 4단계 절차(주입 → FAIL verbatim → 원복 `git diff` 확인)가 그 적색을 4요소 규율로 **강제 관측**한다. 즉 적색은 반드시 관측된다 — 결함은 기재 위치뿐이다.
- **수리 제안** (run-phase 진입 전 manager-spec 또는 진행 중 acceptance 수정): AC-OWN-004 상세에 한 줄 추가 — "RED-now 셀은 m2 뮤턴트 주입 시점에 관측하며, 커맨드·verbatim FAIL 출력·종료 코드·트리 SHA 4요소를 AC-OWN-005의 증거 파일에 함께 기록한다(구현 전 트리에는 주제 코드가 존재하지 않아 pre-fix RED는 구성 불가 — verification-completeness §2.1의 undecidable 성격을 인지한 기재)". 또는 등급을 release-blocking에서 regression-guard(뮤턴트 검증형)로 재분류.

### F2 (advisory) — 이동하는 분모 수치가 트리 SHA 없이 본문에 박혀 있다

- **위치**: spec.md:46, :53("전체 11,917 건 중 104 건"), spec.md:246("815 디렉터")
- **관측**: 이번 재측정에서 총 커밋 **11,936**, SPEC 디렉터 **816** — 공유 저장소의 다른 레인 병합으로 ref가 진행 중이다. 하중을 지는 수치(보유 104건·최신 2026-09-03)는 정확히 재현됐고 안정적이다.
- **제안**: §C.1 재측정 절차 실행 시 분모를 그 시점에 다시 세거나, 본문의 분모 수치에 "측정 시점 3ac58b5a1" 귀속을 붙인다. 판정에는 영향 없음.

### F3 (advisory) — REQ-OWN-010의 SHOULD 등급이 acceptance에만 기록돼 있다

- **위치**: spec.md:195-198(REQ-OWN-010 — 선언형 문구, 등급 미기재) vs acceptance.md:133-134("REQ-OWN-010 은 SHOULD 등급", 미이행 시 FAIL 아님)
- **내용**: 세 계층(spec/plan/acceptance)이 행위자 3값과 관측점(`git log --format='%(trailers:...)' -1`)에 대해 일치하므로 모순은 없다. 다만 등급 지정이 AC 계층에만 살아 있어 spec.md 본문만 읽는 독자에게는 MUST로 읽힌다. REQ 자리에 등급을 명기하거나, acceptance가 등급 SSOT임을 REQ 쪽에서 역참조하면 계층 간 격차가 닫힌다.

### F4 (advisory) — AC-OWN-006(3) 부재-가드의 판정 표기가 종료 코드와 충돌한다

- **위치**: acceptance.md:100(`grep -c "subject prefix"` 양 사본에서 0)
- **내용**: `grep -c`는 매치 0줄에서 **출력 0 + 종료 코드 1**을 낸다. "0"을 종료 코드로 읽는 관찰자는 통과를 실패로 오독한다. 판정 기준을 "출력된 계수가 0"(종료 코드 무시)으로 명기하면 애매함이 사라진다. 현재 문구는 1건/사본으로 RED-now가 관측 가능한 상태(위 증거 절 확인)이므로 가드 자체는 공허하지 않다.

### F5 (advisory) — m3 관측 조건인 `(none) → draft` 픽스처가 plan M1/M2에 명명돼 있지 않다

- **위치**: plan.md:99-101(M1 — 단수 테스트 서술), plan.md:145(§G m3 — "((none) → draft 전환 픽스처에서만 관측 가능 — 픽스처 선택이 판정을 만든다"), acceptance.md:51(AC-OWN-002 판정 (1) — 픽스처 2종 의무화)
- **내용**: 2-픽스처 요구는 acceptance에 있고 plan §G는 그 의존을 정확히 인지하고 있으나, M1/M2.1의 지시 문장은 단수로 읽힌다. §G 자신이 "픽스처 선택이 판정을 만든다"고 쓴 만큼, M2.1 지시에 "픽스처 2종(prev 보유 + `(none) → draft`)"을 명시하면 런 페이즈가 AC 해석에 의존하지 않는다.

(참고, 계수 불산입): AC-OWN-001 예시 테스트명 `TestOwnershipTransitionRule_TrailerAbsent`(acceptance.md:32)과 plan M2.2의 대체 형태 `trailer_absent_emits_unmeasured`(plan.md:114-115)의 표기 불일치 — 양쪽 다 "예시/형태"로 명시돼 있어 런 페이즈 혼란 위험은 낮다. F5와 함께 M2 편집 시 정렬 권고.

---

## 감사 각도별 판정 요약 (dispatch 9개 각도)

1. **Frontmatter**: 12 필드 전부 + 옵션(tier/era)만 사용. ID `SPEC-OWNERSHIP-SILENCE-001` — lint.go:1131 패턴 적합(다중 세그먼트 합법). `phase: "v3.2.0"` 릴리스 라벨(금지 단계명 아님). tier M ↔ 3 산출물 일치. **통과**
2. **§3 판정의 증거 건전성**: (c)는 104건·최신 2026-09-03 재측정 + lint.go:62 verbatim으로 뒷받침. (b) 기각은 :108-111 NOTE + commitOwnerKind 회귀-전용 실재로 근거. (a)는 스코프 논리 일관. §1.1 전제 정정 spec.md:59-69 + progress.md + baseline 3중 기록. **통과**
3. **Two-cell 규율**: AC-OWN-001 모범적(단일 호출 커맨드/verbatim stdout/종료코드 독립 필드+`; echo $?` 금지/트리 SHA·브랜치명 금지 + wrong-reason 배제). 뮤턴트 표 m1/m2/m3 + 생존 why-acceptable 메모 존재. **F1 하나로 should-fix** — 전체 판정에는 영향 없음.
4. **스코프 규율**: 보존 3지점 좌표 실측 일치, manager-develop 인용 불접촉(그리고 그 문장이 (c) 후에도 참임을 코드로 확인), agents-emit 불필요(`.codex/rules` 부재 실측), 쌍둥이 2벌 편집 + `cmp -s` 유지, 낡은 주석 2곳이 실제로 stale임을 검증 — comment-only same-diff 요구 정당. **통과**
5. **계층 간 개정 스윕**: REQ-OWN-010 ↔ plan §D.6 ↔ acceptance §D.2.2 동일 서술. (F3 등급 위치만 advisory.) **통과**
6. **메시지 5요소**: AC-OWN-002가 5요소 전부(빈 prev는 "(none)" 리터럴 포함)를 단언 대상으로 명시, m3가 `(none) → draft` 픽스처에서만 관측됨을 §G가 명시, acceptance가 픽스처 2종 의무화. **통과** (F5 표기 권고)
7. **strict 안전 테스트 설계**: lint.go:62 조건 verbatim 미러 + 신규 코드↔Invalid(Warning) 별개 코드·별개 등급 단언(REQ-OWN-005 관측점) 존재. **통과**
8. **검증 스코핑**: 전체 스위트 금지를 §6 non-goal + AC-OWN-007("부재가 아니라 의도다")로 이중 명시, 상속 적색 분리 절차(§C.4, t577 축) 존재, touched-package 스코프(internal/spec + internal/template) 정확. **통과**
9. **GEARS 형식**: REQ 패턴 코드 대조 적합, AC-OWN-001..008 ↔ 배차 AC-1..8 1:1, §E.1 신호 존재(판정 전 pending 상태로 정확), plan/acceptance 무상태성(프론트매터 자체 없음 — status 축 준수). **통과**

---

## Gaps (이번 감사에서 관측하지 않은 것)

- `go run ./cmd/moai spec lint --strict` baseline 재실행(rc=1·4,698 warnings) — 레인이 기록한 실행을 baseline 귀속으로 수용. 승급 메커니즘은 소스 수준(lint.go:62 verbatim)으로 검증했으므로 판정에 영향 없음.
- CI spec-lint 적색 귀속(3ac58b5a1·91d25bc61·eefbaf65f, gh run list) — dispatch·baseline 기록을 수용, 재조회하지 않음. 상속 적색은 본 카드 판정 축 밖(§6 non-goal)이므로 게이트 무관.
- unmeasured 발화의 per-SPEC 볼륨 — spec §7이 스스로 Gap으로 선언, 구현 후 실측이 설계된 순서다.
- audit_multi 교차모델 수렴 — dispatch 지시에 따라 단일 모델 판정으로 수행.

## Residual-risk

- develop의 `lint.go`는 t518(axes) 축으로 계속 변동 중 — 좌표는 병합 시점 §C.1 재측정이 유일한 방어이며 plan이 이미 그렇게 규정했다. 본 감사의 좌표 일치는 3ac58b5a1 시점 유효다.
- 공유 저장소 ref 진행으로 인용 수치(분모류)는 계속 이동한다 — F2의 권고(재측정 또는 시점 귀속)가 이 위험의 관리 경로다.

## 다음 단계 권고

1. **게이트 개방** — run-phase 진행(Implementation Kickoff Approval은 본 판정과 무관하게 별도 인간 게이트로 유지).
2. **F1 수리** — run-phase 위임문에 "AC-OWN-004의 RED 관측은 m2 뮤턴트 시점에 4요소로 기록(AC-OWN-005 증거 파일에 동봉)" 한 줄을 싣거나, manager-spec이 acceptance.md AC-OWN-004에 동일 문장을 추가. 어느 쪽이든 산출물 수정 없이 위임문 주입으로 충분.
3. F2-F5는 run/sync 진행 중 소화 가능한 advisory — 별도 재감사를 강제하지 않는다.
