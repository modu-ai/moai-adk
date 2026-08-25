# SPEC-VERIFICATION-COMPLETENESS-001 — plan

> Tier M · Class C (plan → run → sync) · 카드 t261. 본 문서의 결정은 가역성 내림차순으로 배치했다(M1의 이름·구조·스코프가 가장 바뀔 가능성이 높은 결정).

## §A. Context

### §A.1 배경과 산출 축

카드 t241 본문에 2026-08-24~25 라인 8개가 누적한 하네스 규칙 6건을 `.claude/rules/` 에 착지한다. 산출 축: (a) 6규칙의 룰 파일화, (b) 착지의 always-loaded 예산 영향 실측. 측정 기반선은 본 워크트리 HEAD `32d2221fa` 에서 plan-phase 실측 완료(spec.md §A).

### §A.2 핵심 설계 결정 (D1–D6)

| # | 결정 | 내용과 근거 |
|---|------|-------------|
| D1 | 파일 이름·위치 | `.claude/rules/moai/development/verification-completeness.md`. development/ 는 작성-규율 파일의 집합지(spec-frontmatter-schema, rule-authoring 등). core/ 는 always-loaded 헌법 파일 전용 — 넣으면 예산 축 (b)가 자기모순이 된다. 파일명은 쌍둥이 `core/verification-claim-integrity.md` 와 짝을 이룬다(주장 무결성 ↔ 검증 완결성). base-0 확인(A-5). |
| D2 | 단일 축 구조 (리드 제안 채택) | 6평행 항목이 아니라 완결성 단일 축 중심 구조. 규칙 1·3·4는 "검증물은 알려진 입력에서 실패가 관측될 때만 완결"이라는 한 축의 세 측면(관측-완결 / 3부 체크 명세 / 2셀 RED·녹색 채택 규율)이고, 규칙 5(교차-계층 소탕)와 6(SHA 고정)은 종속 섹션, 부속 정리(audit-verification, self-application)는 관련 규칙 아래에 건다. 근거: 카드가 통합하려는 규율을 6평행 항목이 재파편화한다; 6규칙 모두 단일 축 없이는 각자 다른 실패 형태로 우회된다. |
| D3 | `paths:` 스코프 값 | `**/.moai/specs/**,**/.claude/rules/**,internal/template/templates/.claude/rules/**,internal/hook/**,scripts/**,**/.moai/astgrep-rules/**,**/.moai/hooks/**` — 검증물 작성 맥락(SPEC 산출물, 룰 파일 양쪽, 훅·게이트 코드, 점검 스크립트, 룰셋)에만 적재. 콤마 구분·공백 없음. 점-디렉터리 경로는 rule-authoring.md 선례대로 `**/` 접두(plan-audit D10 — 중첩 `.claude`/`.moai` 매칭; `**/.claude/rules/**`는 템플릿 트리의 `.claude/rules/`까지 포괄하되 가독을 위해 명시 병기), repo-루트 고정 경로(`internal/**`, `scripts/**`)는 spec-frontmatter-schema.md 선례(`internal/spec/**`)대로 bare. always-loaded 비용 0. |
| D4 | 템플릿 미러 방식 | **옵션 B 채택 — 템플릿 미러 + 완전 중립화.** (i) 6규칙 내용은 범용 규율(모든 사용자 프로젝트의 AC/체크 작성에 유효)이므로 배포 가치가 있고, (ii) CLAUDE.local.md §2.3 에 따라 `.claude/rules/moai/**` 안의 local-only 파일은 `moai update` 시 와이프되어 불안정하며, `.claude/rules/local/` 로 피하면 Template-First [HARD] 위반이다. **로컬 판 == 템플릿 판 바이트 동일**(`cmp` rc=0)을 채택 — 기존 선례(verification-claim-integrity.md: 로컬 판만 근거행 실음, A-6)가 안고 있는 §2.3 경로-충돌 덮어쓰기 위험을 본 파일은 처음부터 피한다. 카드 ID·날짜·SHA 근거는 로컬 추적 SSOT인 본 plan.md §A.4에 둔다. 정정된 A-6 실측(템플릿 판이 로컬 판보다 86B 큰 통째 재작성 분기, diff 36행)은 이 채택을 강화한다 — 선례의 분기 규모가 클수록 바이트 동일의 봉쇄 가치가 커진다. |
| D5 | zone-registry 비적용 | §ID Allocation Policy 는 등록 대상을 4개 헌법 원천 파일(고정 순서: CLAUDE.md → moai-constitution.md → agent-common-protocol.md → design/constitution.md)로 한정한다. development/ 룰 파일의 [ZONE:Evolvable][HARD] 조항은 등록 대상 밖 → **등록 불필요(비적용), 근거와 함께 기록(REQ-VC-008)**. 향후 조항이 헌법 원천으로 승격되면 그때 CONST-V3R5-NNN 병렬 네임스페이스로 등록한다(명시적 deferral). |
| D6 | 예산 측정 절차 | 계측기 = CMD-3(단일 awk, frontmatter 한정 — worktree-guard 수용 확인됨). 절차: (1) 전-값은 이미 SHA `32d2221fa` 에 고정(14파일/179,081B, A-1~A-3, 대조 2건 관측). (2) run-phase M3에서 같은 명령을 run HEAD에서 재실행(재인용 금지 — 규칙 6 자기적용). (3) 판정: 신규 파일이 열거에 없고 카운트 14 유지(델타 발생 시 명명된 외부 파일로 귀속). (4) rule-authoring 의무 (a)는 스코프 채택으로 비발화 — 그 사실 자체를 측정 기록과 함께 서술. |

### §A.3 룰 파일 아웃라인 (M1 작성 지침 — 영어 본문)

```
# Verification Completeness
> 문서 헤더: loading-scope 노트 + 단일 축 선언문
  "a verification artifact (check, gate, acceptance criterion, rule, assertion)
   is incomplete until its failure has been observed on a known input"

1. The completion axis (rules 1+3)
   1.1 Observed-failure completion  — creating the check is not completion;
       observing it fail on a failing input is. Failure form: report-not-verdict
       (a probe printed rc yet returned the last echo's rc=0; a citation sweep
       printed unverified citations while exiting 0, truncated by head;
       an AC criterion "calls it but never uses the result" stalling rounds).
       Observed instance (evidence footnote, not a rule): a selector that
       swept nothing still prints ok — a zero-match test-name regex, a
       zero-hit grep read as pass. Counting what was swept is the same
       completion act as observing the failure.
   1.2 The three-part check spec (rule 3) — every check states together:
       (a) WHEN it must run to be meaningful (a check at a structurally
       always-green moment proves nothing), (b) the INPUT that turns it red
       (unfinished until that failure is actually observed), (c) the failure's
       REACHABILITY (who sees the red: log level, trace, exit code).
       Defect forms: (a)-missing = always-green check; (c)-missing = red logged
       at debug level, invisible in production and trace.
2. Two-cell adoption discipline (rule 4)
   RED-now on the pre-implementation tree AND green path (which milestone flips
   it, what the passing output becomes) as a pair — and RED alone is not
   enough: the criterion must be RED FOR THE RIGHT REASON, and the RED cell
   must say why it is red. Three failure directions:
     - vacuous criteria (green today; caught by BASE-RED),
     - impossible criteria (red today and forever; sail through BASE-RED),
     - wrong-reason red (red at arrival AND after implementation because of
       pre-existing files this work never touches — indistinguishable from
       impossible criteria unless you ask WHY it is red).
   Green-path disqualification: a green path that runs through "someone fixes
   the unrelated files", or that no change can flip, disqualifies the
   criterion.
   Mutant probe (rule 2): before adopting an AC, try to write a mutant that
   satisfies the AC while violating its REQ — if writable, the AC is too
   shallow. Rule-pairing corollary: invalid-cases-only rule passes an
   all-matching mutant; valid-cases-only rule passes a nothing-matching one.
3. Cross-layer revision sweep (rule 5)
   The layer a rule constrains is its blind spot. When revising a criterion,
   sweep the REQ/plan items it cites in the same pass — a revision does not
   end in one file. Defect forms: rescoped AC with its REQ still scoped to the
   whole tree (false at arrival); REQ demanding X while the new AC asserts X
   does not change — with the plan instructing the forbidden side.
4. Evidence pinning (rule 6)
   Invariant assertions (byte-unchanged / PRESERVE / absence) pin the tree SHA
   where the evidence was collected, never a moving branch name — an upstream
   advance silently falsifies the assertion. Corollaries: never re-cite a
   measured divergence without re-measuring; on rebase, re-measure and re-pin.
   Discriminator: pinning is NOT unconditional — ask whether a moving ref
   flipping the claim is a true signal about the subject or spurious red from
   unrelated upstream work; provenance-style statements about the mainline
   itself keep the moving ref (pinning would weaken them). Without this
   discriminator the next person pins everything mechanically.
5. Corollaries
   - Audit-verification: to verify a fix landed, do not grep for the fixed
     form — run the two command forms directly and observe them diverge
     ("the cheapest modification that passes is one that satisfies the grep").
   - Self-application: a rule's own text must comply with the rule (a warning
     about un-timestamped numbers must not state its own number in the
     present tense).
```

각 **규칙 단위**(§1.1, §1.2, §2의 규칙 4와 돌연변이 탐침 각각, §3, §4 — 총 6행)은 `> Evidence:` 인용행(중립 결함 형태 서술, §25.1 generic prose 클래스)으로 마무리한다 — 6행이 6규칙과 1:1 대응한다. §5 부속 정리는 근거를 본문 간단 인용으로 실되 `> Evidence:` 마커 수(=6)에는 포함하지 않는다. 규칙 파일 어디에도 카드 ID·내부 날짜·SHA·SPEC ID가 등장하지 않는다(REQ-VC-002/005).

### §A.4 근거 행렬 (Provenance Matrix — 내부 추적 SSOT, 템플릿 미배포)

| 규칙 | 원천 카드/반복 | 관측일 | 관측된 사건 (규칙 파일 인라인 근거의 원문) |
|------|---------------|--------|-------------------------------------------|
| VC-1 (관측-완결) | t197 iter4/6/7 (lane 기여) | 2026-08-24 | probe.sh의 run()이 rc를 출력하고 마지막 echo의 rc(항상 0)를 반환해 실패를 은폐 → iter6에서 run_rc로 수정(주입 테스트 rc 1 관측) → iter7의 신규 citation-sweep.sh가 동일 결함 재현(비교 없는 출력·무조건 rc 0; rc=0인 채 미검증 인용 3건 잔존, head로 잘림). 2라운드 전에 고친 결함이 새 도구에서 재발. 형태: "체크를 만들 때 판정이 아니라 보고서를 만든다". AC 측 동형: INIT-001이 3라운드 0.75에서 정체 — 피드백 전부 "호출은 하되 결과를 안 씀 / 주입은 하지만 동사에 안 묶음 / 거부는 하지만 쓰기 선행을 안 증명". 근거 문서 `.moai/reports/t197/procedure-defect.md` 는 본 트리에 없음(A-8, t197 미머지) → 요약 인라인. |
| VC-3 (3부 체크 명세) | t217 M3 (lane-7); t230 N1 | 2026-08-24 | t217 M3: 앞 블록에 묻히지 않도록 만든 tail 실행이 자기 실패를 slog.Debug로 기록 → 프로덕션에서 무음, 트레이스에 부재(도달성은 봤고 관측가능성은 없음 — (c) 결측). t230 N1: "릴리즈 순서를 체크로 제약하라"는 지시가 구조적으로 항상-녹색인 시점(M2 끝)에 도는 체크를 생산 — 실패할 수 없는 체크는 통과해도 아무것도 증명하지 않음(지시측 결함, (a) 결측). |
| VC-4 (2셀 RED·녹색) | t228 iter2 E1/E2 (lane-1) | 2026-08-24 | E1: 배포된 sgconfig.yml에서 sg test가 룰 테스트 0개인 트리에서 'ok. 0 passed; 0 failed' EXIT=0 출력 → 아무것도 검사 안 했는데 AC 통과. E2: `grep -c astgrep internal/template/catalog.yaml` = 0 → 올바른 작업으로는 절대 통과 불가한(엄격이 아닌 불가능) 기준. 자기강화 수정: BASE-RED는 공허 방향만 잡는다 — E2는 오늘 빨강이고 구현 후에도 빨강이라 BASE-RED를 통과해 버림 → 기준마다 2셀(RED-현재 셀 + 녹색-경로 셀). |
| VC-2 (돌연변이 탐침) | **원천: t197 7-라운드 블록 제안 (b)**; 룰 쌍 원리 관측: t228 iter2 (lane-1) | 2026-08-24 | 귀속 정정(plan-audit D2): 탐침 자체는 t197 블록의 제안 규칙 (b) — "새 AC를 쓰면… mutant를 하나 적어본다 — 적히면 AC가 얕다". AC측 근거: INIT-001 3라운드 정체 피드백("호출은 하되 결과를 안 씀" 등 — 돌연변이가 통과하는 얕은 AC의 실례). t228 lane-1은 룰 쌍 원리(무효-only→전부-매칭 룰 통과, 유효-only→무매칭 룰 통과)의 관측원천이다. |
| VC-5 (교차-계층 소탕) | t228 iter3 (lane-1) | 2026-08-24 | AC-019 재스코프하는 동안 REQ-021은 여전히 템플릿 트리 전체 스코프(도착 시 거짓: SPEC이 건드리지 않는 파일에 SPEC-ID 22건+날짜 90건). 직접 모순쌍: REQ-018(catalog.yaml 재생성 요구) vs AC-019(make build가 이를 바꾸지 않는다고 단정) — plan.md이 금지된 쪽을 지시. |
| VC-6 (SHA 고정) | t228 iter4 N1 (lane-1) | 2026-08-25 | plan.md의 PRESERVE 증명이 `git diff --stat origin/main -- internal/hook/` 사용 — 작업 0인데 origin/main이 10커밋 진행(f7eec06c7)되어 18파일 변경/2,825 삭제 출력. 고정 SHA로 재측정 시 빈 출력. 같은 문서의 다른 인용은 전부 SHA 고정, 이것만 브랜치명. 같은 세션에서 lead와 lane이 259커밋 낡은 divergence를 재측정 없이 재인용. 판별자(t228 iter5): 이동 ref가 주장을 뒤집는 것이 주제에 대한 진짜 신호인지, 무관한 상류 작업의 허위 빨강인지 물어라 — 메인라인 자체에 대한 근거-서술은 origin/main 유지(고정이 오히려 약화). |
| 부속: audit-verification | t228 iter5 라운드 | 2026-08-25 | 수정 착지 확인에 수정된 형태를 grep하지 말 것 — 두 명령 형태를 직접 실행해 발산을 관측하라("grep을 만족하는 가장 싼 수정이 결함을 고치는 수정이 아니다"). |
| 부속: self-application | t228 N5 | 2026-08-25 | 이동-ref 숫자를 경고하는 코멘트가 자기 숫자를 시제 없이 현재형으로 서술 — 경고가 경고하는 바를 그대로 수행. 규칙/경고 문장 자체가 그 규칙을 준수하는지 검사하라. |
| 각주: counting (규칙 아님) | lanes 9·11 (독립) | 2026-08-25 | 좁은 -run 정규식이 0개 테스트 선택에 rc 0; 13개 명시 함수명 + -v 서브테스트 카운트로 FAIL 노출; escape된 -run 형태 0선택 후 ok 출력. 동형: grep 0행=통과 AC, 단언 앞 t.Skip, 0룰 sg test. 채택 제외 근거(plan-audit D7 재인용): t261의 스코핑('이 여섯을 규칙 파일로 옮기고' — 6건 한정) + t241의 규정('전부 "무엇을 실제로 훑었는지 세지 않으면 통과가 무의미하다"는 한 규칙의 사례다') — 축 §1.1의 근거 각주로만. |

### §A.5 예측 장부 (Prediction Ledger — t241 주의 반영)

t241: "제안 규칙의 효과는 미검증 예측이므로 Lessons Protocol의 prediction/verified 짝으로 기록 후 다음 라운드 확인 필요." 아래 6줄이 그 prediction 기록이다 — 규칙이 작동한다는 주장은 랜딩 시점까지 미검증 예측이며, 다음 라운드(또는 차기 감사)에서 각 행에 verified: true|false를 부기한다(moai-constitution §Lessons Protocol의 harness-edit prediction/verified 규율과 동형).

| 규칙 | prediction — 규칙이 작동한다 = | 다음 라운드 확인 |
|---|---|---|
| VC-1 | 신규 체크 승인물 중 '출력만 내고 rc·비교를 반환하지 않는' 도구 0건 | 라운드 내 신규 체크 diff 판독(rc 반환 구조) |
| VC-2 | 'mutant가 쓰여 AC 무효화' 감사 지적 0건 | plan-audit shallow-AC 지적 수 |
| VC-3 | 구조적으로 항상-녹색 시점의 체크 승인 0건 | 게이트 제안 리뷰 기록 |
| VC-4 | vacuous·impossible·wrong-reason-RED AC 채택 0건 | plan-audit AC 지적 유형 집계 |
| VC-5 | AC 재스코프 후 해당 REQ/plan 미수정 지적 0건 | 감사 traceability 지적 수 |
| VC-6 | 이동 ref 불변 주장·재측정 없는 재인용 0건 | 감사 moving-ref 지적 수 |

## §B. Known Issues (위임 프롬프트 주입 항목)

1. **t197 근거 문서 부재** — `.moai/reports/t197/procedure-defect.md` 없음(A-8). 룰 파일은 포인터가 아니라 중립 요약을 인라인으로 실어야 한다(REQ-VC-002).
2. **naive 계측 명령의 이론적 허위-제외** — 본문 행이 정확히 `paths:`로 시작하는 파일이 있으면 naive grep이 잘못 제외한다. 현재 트리에서는 미발생(A-2)이나 계측기는 CMD-3(awk, frontmatter 한정)을 정본으로 한다.
3. **워크트리 guard의 명령 복잡도 거부** — 본 세션에서 복합 루프/다중 리다이렉트 명령 2건이 거부됨. run-phase 검증 명령은 단일 파이프라인 형태로(CMD-3 형식은 수용 확인됨).
4. **zsh `=` 확장** — `echo ===` 가 zsh equals-expansion으로 실패함. 스크립트/검증 인용문에서 `===` 리터럴 금지.
5. **기존 로컬↔템플릿 분기 선례** — verification-claim-integrity.md 등은 로컬 판만 근거행을 실어 §2.3 덮어쓰기 위험을 안고 있다(A-6). 본 SPEC은 바이트 동일로 이 위험을 상속하지 않는다. run-phase는 이 선례를 따라 근거행을 로컬 판에 몰래 추가하지 말 것.

## §C. Pre-flight (run-phase 시작 전)

각 항목 독립 실행, 관측을 §E.2 첫 행에 기록:

1. `git rev-parse --short HEAD && git branch --show-current` → 워크트리/브랜치 재확인(규칙 6: 기반선 SHA `32d2221fa` 와 다른 HEAD여도 무방 — 재측정이 규율).
2. CMD-3 재실행 → 14파일/179,081B 재확인(전-값과 불일치 시 명명된 외부 파일로 귀속 기록 후 진행).
3. `grep -rn 'verification-completeness' .claude/rules internal/template/templates` → 여전히 0행(사전 조건).
4. `.claude/rules/moai/development/rule-authoring.md` 및 spec.md §A 재독.

## §D. Constraints — PRESERVE 목록 (무변경)

- 기존 82개 룰 파일 전체 바이트 무변경 (`git diff --name-status 32d2221fa -- .claude/rules/moai` 결과가 신규 파일 1개의 `A` 행만 허용 — pinned-SHA diff, 이동 브랜치 금지).
- `internal/template/templates/.claude/rules/moai/` 기존 19개 development 파일 무변경 (동일 판정, `A` 1개만).
- `internal/template/catalog.yaml` 무변경 (A-7: 룰 파일 비목록화 확인됨).
- `.claude/rules/moai/core/zone-registry.md` 무변경 (D5 비적용).
- 기존 SPEC 디렉터리·`.moai/docs/` 무변경. 본 SPEC 디렉터리 외 `.moai/` 쓰기 없음.

## §E. Self-Verification (run-phase 보고 의무)

run-phase 완료 보고는 acceptance.md §D AC 전수에 대해 (i) 실행 명령, (ii) 관측 출력 verbatim, (iii) 측정 HEAD SHA 를 progress.md §E.2 에 증거 행으로 기록한다(verification-claim-integrity 5-절 형식: Claim/Evidence/Baseline-attribution/Gaps/Residual-risk). 최소 필수 증거: CMD-3 실행 출력(run HEAD), 대조 2건, `cmp` rc, 중립성 grep 카운트, `make build` 종료 코드, `git diff --name-status 32d2221fa -- .claude/rules internal/template/templates` 출력.

## §F. Milestones (우선순위 순, 시간 추정 없음)

### M1 — Priority High: 룰 파일 작성 (로컬 트리)

가장 가역성이 낮은 결정(이름·구조·스코프·중립 문체)이 모두 여기 산출물에 굳는다. 리드/운영자 검토는 이 마일스톤 diff에 집중한다.

1. `.claude/rules/moai/development/verification-completeness.md` 작성 — §A.3 아웃라인대로: 단일 축 구조, 6규칙 [ZONE:Evolvable][HARD] 조항, 규칙별 `> Evidence:` 중립 근거행 6행(§2는 규칙 4·돌연변이 2행 — §A.3 마감 지침), 부속 정리, D3의 `paths:` frontmatter. 영어 본문.
2. 자기점검: 규칙 파일 본문에 금지 토큰(카드 ID, 2026-08, 7-8자리 sha, `SPEC-`) 0건 — 문서화된 절차상 `grep -cE` 로 관측.
3. AC-VC-001/002/006 녹색 전환 대상.

### M2 — Priority High: 템플릿 미러 + 재임베드 + 중립성 스캔

1. 동일 내용을 `internal/template/templates/.claude/rules/moai/development/verification-completeness.md` 에 작성 → `cmp` rc=0 (D4: 바이트 동일, 근거행 로컬 몰래추가 금지 — §B.5).
2. `make build` 실행 → 종료 코드 0 관측.
3. 템플릿 판 중립성 스캔(acceptance.md CMD-N) → 0건 관측. §25.3 수동 5항목 체크리스트(C1–C5) 통과.
4. AC-VC-004/005 녹색 전환 대상.

### M3 — Priority Medium: 예산 실측 + 자기적용 감사

1. Pre-flight §C.2 대로 CMD-3을 run HEAD에서 재실행 → 14파일, 신규 파일 열거 제외, 바이트 합 179,081 유지(또는 외부 귀속) 관측.
2. rule-authoring 의무 (a) 비발화 서술 + 측정 기록(카드 요구 축 (b)의 증명).
3. 자기적용 감사: acceptance.md §D 전수 재독 — 델타/불변 AC마다 SHA 고정 기반선 + 녹색 경로 쌍 상존 확인(AC-VC-008).
4. §E.2 증거 행 전수 기록, progress.md §E.3 신호.

## §G. Anti-Patterns

- 6평행 항목 구조(D2 기각) — 카드가 통합하려는 규율을 재파편화한다.
- 카드 ID·날짜·SHA를 템플릿 판(= 로컬 판)에 실는 것 — §25.1 금지, CI 중립성 가드 위반.
- 근거행을 "t197 참고" 식 포인터로 쓰는 것 — 문서가 없다(A-8); 요약이 인라인이어야 한다.
- naive grep을 계측 정본으로 쓰는 것 — 허위-제외 입력이 이론상 존재(§B.2).
- 예산 AC의 기반선을 origin/main 등 이동 ref에 두는 것 — 규칙 6 자기적용 위반; SHA `32d2221fa` 고정 + 재측정.
- core/ 에 배치해 always-loaded에 넣는 것 — 축 (b) 자기모순.
- 기존 선례를 따라 로컬 판에만 근거행 추가(로컬≠템플릿 분기) — §2.3 덮어쓰기 위험 상속(§B.5).
- 근거 없는 규칙 — "다음 사람이 무시한다"(카드 경고). 반대로 근거를 규칙보다 길게 쓰는 것도 실패 형태: 인라인 근거는 결함 형태 요약에 그친다.

## §H. Cross-References

- 카드 t261 / t241 (근거 원천) — dispatch의 card 필드로 소급.
- `.claude/rules/moai/core/verification-claim-integrity.md` — 쌍둥이 규칙(claim 축).
- `.claude/rules/moai/development/rule-authoring.md` — always-loaded 비용 의무 (a)~(d).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — paths 스코프 문법 선례 + SPEC 스키마 SSOT.
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1~§25.3 — 중립성 클래스/치환/수동 체크리스트.
- `.claude/rules/moai/core/zone-registry.md` §ID Allocation Policy — D5 근거.
- CLAUDE.local.md §2(Template-First), §2.3(managed-roots 와이프 — D4 동기).
- spec.md §A(기반선) / §D(REQ-VC-001..008) / acceptance.md §D(AC 매트릭스).
