# SPEC-SPEC-LINT-ID-ARG-001 — 진행 기록

카드: t518 (축 3) · 브랜치: WT-spec-lint-axes · base: 0b1e27877
형제: SPEC-SPEC-LINT-BLIND-AXES-001 (축 1·2, 반경 분리 — `internal/spec` 소관)

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (v0.3.0에서 S → M 재분류 — 아래 수리 2회차). 아티팩트 3개(`spec.md` + `plan.md` + `acceptance.md`) + 이 `progress.md`. ~~v0.2.0까지: Tier S, 아티팩트 2개, AC는 spec.md §D 인라인~~
- SPEC ID 정규식 검사: `PASS` (Bash 실행, 출력 인용은 완료 보고 참조)
- 요구사항 8건(REQ-SLI-001..008) · 수용 기준 **기준 8개 / 식별자 11개**(`001a/b/c`, `004a/b`가 각각 한 기준) — 기준 단위가 Tier S 천장 8과 같다. S/M 경계 신호로 기록. 셈법은 spec.md §D.0-6이 정본
- baseline 귀속: `moai-adk v3.2.0-rc.0` @ `0b1e27877`, 파이프 없는 rc 판독. 대조쌍 증거 `.moai/reports/t518/id-arg-control-pair.txt`
  - ID 형태 rc=1 / `ParseFailure` 1건 · 경로 형태(대조군) rc=0 / warning 1건 · 디렉터리 형태 rc=1 · 부재 ID rc=1(ID 형태와 출력 동일) · invalid-argument 통로 rc=3 실재
- 코퍼스 총량 4,346은 **시점 재유도값**이며 확정값이 아니다(두 SPEC 저작 이후 재측정 4,366; 증분 +20 = 이 SPEC 8 + 형제 12)

### 수리 1회차 (v0.2.0, 2026-09-07)

- 판정문: `.moai/reports/t518/plan-audit-id-arg.md` — FAIL 0.56 (Tier S 임계 0.75), critical 2 + major 4 + minor 5 + optional 2
- 닫은 blocking 결함: D-1(REQ-SLI-006/AC-SLI-006 쌍 재작성 — 도달 불가능한 `ValidateSpecID` 단언을 앵커 불변식으로 교체) · D-2(`specIDPattern` 동명 충돌 명시 + import 불가 확정 + 새 이름 지정 + 어긋남 감지 대상 명시) · D-3(§G #4 기전 정정, 세 번째 원인 추가, 거부는 유지) · D-4(기준 디렉터리 → `findProjectRootFn`, `detectBaseDir`는 인자 없는 스캔 전용, AC-SLI-001c 신설) · D-5(REQ-SLI-007용 AC-SLI-006 신설 + 뮤턴트) · D-6(11개 AC 식별자 전부 REQ 인용 + §D.3 전수 대응표) · D-7(반경 격리 AC에 base SHA·명령 2개·뮤턴트, 파이프 금지 명시) · D-8(004a/004b/007에 뮤턴트, 004a 채취 창 M1 고정) · D-9(선택지 3 기각 근거를 「대조군」으로 교체, 소비자 주장은 미검증 라벨)
- 함께 처리한 minor/optional: D-10(코퍼스 수치 재유도 표기) · D-11(AC 셈법 §D.0-6) · D-12(REQ-SLI-003 라벨 `Where`→`When`) · D-13(후속 카드 t528 지명 — 발행 사실은 리드 지시 전달분이며 이 트리에서 큐 미확인)
- **AC 재번호 발생**: 옛 004/005 → 004a/004b(한 기준의 두 항목), 옛 006 → 005, 신설 → 006(도움말), 옛 007 → 007, 옛 008 → 008. 사유: 도움말 AC 신설로 기준 단위가 9가 되어 Tier S 천장 8을 넘었고, 대조군 무변형 두 항목이 같은 REQ의 양면이므로 한 기준으로 묶었다
- 수집기 정규식 재유도: 이 spec.md의 REQ 정의 행 **8건 전부 수집**(목록 형식 유지 확인)
- **운영자 결재 대기 1건**(spec.md §H): 앵커된 모양 정규식의 거처 — α 로컬 사본 / β `internal/cli/specid` 공유 boundary / γ 탐침만. **Implementation Kickoff Approval 전에 닫혀야 한다.**

### 수리 2회차 (v0.3.0, 2026-09-07)

- 판정문: `.moai/reports/t518/plan-audit-iter2.md` — FAIL 0.71 (Tier S 임계 0.75), critical 1 + major 2 + minor 1 + optional 1
- **N-3 처분: Tier S → Tier M 재분류.** v0.2.0의 「기준 단위 8」은 `001a/b/c`와 `004a/b` 묶음으로 만든 값이고, **그 묶음의 동기가 성질이 아니라 천장이었음을 v0.2.0의 이 파일이 스스로 적었다**(아래 회차 기록 참조). 묶음을 철회하면 기준은 **11개**이며 어떤 자연스러운 셈법으로도 Tier S 천장 8을 넘는다. 예산을 완화하는 대신 tier를 올렸다 — 임계 0.75 → **0.80**, 아티팩트 3개(`acceptance.md` 신설, AC 이관). 셈법 정본은 `acceptance.md §A.6`. **기각한 대안**: 기준을 8로 줄이는 범위 축소 — 잘라야 할 세 후보 중 둘(도움말 계약·디렉터리 형태)이 감사 요구와 이견 기록으로 들어온 항목이라, 예산에 맞추려 되돌리는 것은 사후 회계의 거울상이다
- **N-1 닫음(critical)**: AC-SLI-005 픽스처를 `sub/SPEC-FIX-001/spec.md` → **`SPEC-A-1`**로 교체. 옛 픽스처는 `/`를 담아 판별식의 **경로 신호 선행 배제**에 먼저 걸렸고, 그래서 정규식을 앵커 없는 것으로 바꾸는 뮤턴트가 **판정에 닿지 못해 AC가 GREEN을 유지했다** — REQ-SLI-006의 유일한 가드가 공허했다. 새 픽스처는 3자리가 아니므로 앵커된 패턴은 기각·앵커 없는 패턴은 수용해 **판단이 실제로 정규식에서 일어난다.** 「`/`를 담은 픽스처로는 이 불변식을 잴 수 없다」는 구조적 사실을 `acceptance.md §A.1`에 명시
- **N-2 닫음(major)**: AC-SLI-008 명령 ①의 귀속 범위를 `0b1e27877..HEAD` 전체 → **이 SPEC-ID를 담은 커밋들**로 한정. 형제 SPEC과 브랜치를 공유하므로 옛 범위에서는 **형제의 착지가 이 카드를 FAIL시켰다**(wrong-reason red). 전제(모든 커밋이 SPEC-ID를 담음)를 커밋 수 대조로 확인하는 절차를 함께 넣고, 성립하지 않던 판정 문구도 고쳤다
- **N-5 닫음(optional)**: 선택지 (i)(`ValidateSpecID`를 판별식 앞에 두기)가 **왜 불가능한지**를 §B에 기록 — `/`를 거부하므로 먼저 돌리면 경로 형태 대조군이 전멸한다. 형제 `view`에서는 같은 배치가 옳다(받는 인자 모양이 다르다)
- **N-4(optional, 가설)**: 미처분. `detectBaseDir`가 두 경로 모두에서 linter에 설정되는지는 `internal/spec` linter 본문을 읽어야 하고 **이 회차에서도 읽지 않았다 — 미검증으로 남긴다.** 요구 자체는 흔들리지 않는다
- 요구사항 8건 유지 · 수용 기준 **식별자 11 = 기준 11**(묶음 철회) — 둘 다 Tier M 천장 16 이하
- 재측정(파이프 없음): `moai spec lint <이 spec.md>` → rc=0, `0 error(s), 8 warning(s)`, 전부 `CoverageIncomplete`. 코퍼스 총량 **4,367**(증분 +21 = 이 SPEC 8 + 형제 13)
- **남은 운영자 결재 1건**(spec.md §H): 앵커된 모양 정규식의 거처 — α / β / γ. **Implementation Kickoff Approval 전에 닫혀야 한다.** γ를 고르면 REQ-SLI-006과 AC-SLI-005가 함께 다시 쓰여야 하고, **판별식 순서를 뒤집는 결정이 나오면 AC-SLI-005의 픽스처 선택도 재검토 대상이다**(§A.1)

### 수리 3회차 (v0.4.0, 2026-09-07)

- 판정문: `.moai/reports/t518/plan-audit-iter3.md` — **PASS 0.86**(Tier M 임계 0.80), PASS-with-debt. blocking 3건 중 이 SPEC 소관 1건(I-1, **major**)을 상환했다
- **닫은 blocking — I-1(major): AC-SLI-008의 전제 확인이 양방향으로 성립하지 않았다.** v0.3.0은 전제(「이 카드의 커밋 메시지가 모두 이 SPEC-ID를 담는다」)를 **커밋 수 대조**로 확인했다 — `--grep` 목록의 개수 vs `git log --oneline base..HEAD -- internal/cli/`의 개수. 두 방향 모두 무너진다:
  - **거짓 경보**: 문서 전용 커밋(spec/plan/acceptance/progress 갱신)은 SPEC-ID를 담지만 `internal/cli/`를 만지지 않아 `--grep` 쪽에만 있다 → 누락이 0건인데 경보가 뜬다. 이 카드는 문서 수리를 세 번 했으므로 발생은 사실상 확정이다
  - **거짓 통과(더 나쁜 쪽)**: 문서 전용 커밋 1개와 SPEC-ID 누락 `internal/cli/` 커밋 1개가 **상쇄되어 두 수가 같아진다** → 확인이 통과하고, 그 누락 커밋은 명령 ①의 목록에도 없다. **명령 ①이 전제에 기대는 바로 그 사각지대가 확인을 통과한다**
  - 원인은 하나다: 개수 대조는 두 집합의 **크기**만 보므로 반대 방향의 두 차이가 상쇄된다
- **처분: 개수 대조 폐기 → 「귀속되지 않은 커밋 집합의 직접 판독」.** `git log --format='%H %s' 0b1e27877..HEAD` 출력의 각 줄을 읽어 **두 SPEC-ID 중 어느 것도 담지 않은 줄**을 모으고, 그 **잔여 집합이 비어 있음**을 확인한다(형제와 브랜치를 공유하므로 두 ID의 합집합이 옳은 모집단이다). 비어 있지 않으면 각 커밋의 `git show --name-only --format=` 출력을 그대로 인용해 `internal/spec/` 무접촉을 보인다 — 미판정 커밋이 하나라도 남아 있는 동안 통과가 아니다. **크기를 비교하지 않으므로 상쇄될 것이 없다.** 「파이프 안에서 세지 않는다」 유지
- **기각한 대안과 그 대가를 함께 기록했다**: `--grep` 한정을 떼고 전 범위에서 `internal/spec/` 접촉을 보는 자족적 명령 — 확인할 전제가 사라져 가장 단순하나, **N-2가 그대로 되살아난다**(형제의 정당한 착지로 이 카드가 아무 잘못 없이 FAIL). 상시 빨강은 침묵과 같은 자리에 도착하므로 기각
- **문서에 「통과할 수 있으면서 위험이 그대로인 검사는 검사가 없는 것보다 나쁘다」를 [HARD]로 명시했다** — 검사가 없으면 사각지대는 사각지대로 남지만, 통과하는 검사는 **점검했다는 기록**을 남겨 다음 독자가 다시 볼 이유를 없앤다. **원문을 지우지 않고 반례를 병기**했다
- §C 완료 정의 체크박스(옛 「커밋 수 대조로 SPEC-ID 누락 커밋이 0건」)도 새 절차로 갱신 — 완료 정의가 그 절차를 게이트로 삼고 있었으므로 함께 고치지 않으면 폐기된 절차가 반드시 실행된다
- **근거의 성질**: 집합 논증이며 실행 관측이 아니다. 현재 `0b1e27877..HEAD` 범위의 커밋은 **0건**(HEAD == base)이므로 오늘 발화시켜 볼 수 없다. 다만 **반례의 존재는 논증만으로 확정된다** — 미검증으로 남기지 않고 이렇게 이름 붙인다
- **감사 판정 반영 — Tier M 재분류는 자체 근거로 정당하다.** 「천장 회피」가 아니라는 판정을 받았고, 그 근거가 천장이 아니라 실질임을 `§G #5`에 네 항목으로 기록했다(아티팩트 규정 일치 · 셈법이 `acceptance.md §B` 소제목 구조에 반영 · 범위 축소 기각 논거가 옳음 · **임계가 올라가는 쪽을 골라 회피 동기와 정반대**)
- **감사가 남긴 optional 2건은 미처리**: I-2(§H 미해결 1에 형제와 대칭인 명명 게이트가 없음 — 다만 Implementation Kickoff Approval이 실재하는 필수 게이트이므로 차단이 없는 것은 아니다) · I-3(`acceptance.md §A.1/§A.2` 제목의 버전 라벨 중의성). 부재가 아니라 **미처리로 기록**한다. N-4도 여전히 가설이며 이번 회차에도 `internal/spec` linter 본문을 읽지 않았다
- 요구사항 **8건** · 수용 기준 식별자 **11건**(기준 11) — v0.3.0에서 변동 없음. 재유도(파이프 없음): `moai spec lint .moai/specs/SPEC-SPEC-LINT-ID-ARG-001/spec.md` → rc=**0**, `0 error(s), 8 warning(s)`
- **남은 운영자 결재 1건**(spec.md §H): 앵커된 모양 정규식의 거처 α / β / γ — **여전히 OPEN.** 형제 SPEC의 `BLIND-AXES-DISCRIMINATOR-GATE`가 v0.4.0에 닫혔으나 **이 결정에는 아무것도 함의하지 않는다**
- **판정문의 do-not-disturb 항목 무이동 확인**: N-1의 픽스처 `SPEC-A-1`과 그 선택 근거 미변경(감사의 대안 `SPEC--1`은 앵커 없는 쪽에서도 매치되지 않아 뮤턴트가 발화하지 못했을 것 — **저자의 기각이 옳았다는 근처-실패 기록**을 §A.1에 남겼다) · 명령 ① 자체(`--grep` + 커밋별 `git show --name-only`)는 옳으므로 **손대지 않았다**(무너진 것은 전제 확인 절차뿐이다) · 명령 ②의 `-- internal/cli/` 한정 미변경

### 미해결 1 결재 (v0.5.0, 2026-09-07)

- **판정문 없음 — 이 회차는 감사 부채 상환이 아니라 운영자 결재의 착지다.** iteration-3(`PASS 0.86`)에서 이 SPEC 소관 blocking(I-1)은 v0.4.0에 닫혔고, 남아 있던 것은 **운영자 결재 1건(§H 미해결 1)**뿐이었다
- **[운영자 결재 완료] §H 미해결 1 = 후보 `β`(기존 leaf 패키지 `internal/cli/specid`에 모양 검사 함수 추가).** α(로컬 사본)·γ(탐침만) 기각. 반영한 자리: `spec.md §H` 「미해결 1 — 결재 완료」(근거 전문, 원문 병기) · `spec.md §C` REQ-SLI-009 신설 · `acceptance.md §B` 기준 12·13(AC-SLI-009 · 010) · `plan.md §B` D4 결재 배너 + `§C` M3 + `§F` 안티패턴 4건
- **채택 논거는 반대 논거의 전제가 성립하지 않는다는 것이다.** `plan.md` D4가 세운 「소비자 하나짜리 패키지를 만들지 않는다」는 기준에 β는 **닿지 않는다** — β는 패키지를 만드는 안이 아니라 **이미 있는 패키지에 함수를 더하는 안**이고, 그 package doc(`internal/cli/specid/specid.go:1-8`)이 스스로를 「모든 CLI SPEC-ID 경계에서 사용하는 단일 검증 헬퍼」로 선언하고 있다
- **[HARD] 실측 정정 — 소비자는 셋이 아니라 다섯이다. 카드가 묻지 않은 발견이고, 논거를 강화하는 동시에 반경을 넓힌다.** `spec.md §H`와 `plan.md` D4는 「소비자 이미 셋(view/status/close)」이라 적었다. 재유도(`grep -rn 'specid\.ValidateSpecID' internal/`, 테스트 제외): `internal/cli/spec_view.go:46` · `internal/cli/spec_status.go:72` · `internal/cli/spec_close.go:108` · `internal/kanban/board_store.go:268,360` · `internal/kanban/status_read.go:109` = **다섯 호출부, 두 패키지.** 「셋」은 **`internal/cli` 안에서만** 참이었다. ① β 논거는 **강해진다**(한 패키지의 사설 헬퍼가 아니라 두 패키지가 공유하는 경계이며, package doc의 선언이 실제로 지켜지고 있다). ② 반경 판단은 **넓어진다**(`internal/kanban`도 의미가 넓어진 패키지를 import하는 쪽이 된다). **조용히 「셋」으로 두지 않고 정정을 병기했다**
- **[HARD] 결재와 함께 결정된 두 의무를 산문이 아니라 검증 가능한 기준으로 세웠다** — REQ-SLI-009(요구) + AC-SLI-009 · AC-SLI-010(기준):
  - **AC-SLI-009 — 기존 소비자 무변형을 「증명」한다.** 「오늘 아무도 모양 검사를 쓰지 않으므로 함수를 더하는 것은 순수 추가다」는 그럴듯하지만 **관측된 적이 없다 — 가설이지 측정이 아니다.** 갈래 ②가 함수 추가 **전/후** 두 트리에서 `go test ./internal/cli/... ./internal/kanban/...`를 돌린 두 출력의 축자 인용을 요구하고, **다섯 경로가 그 출력에서 실제로 실행됐음**까지 요구한다(셀렉터가 0건을 골라도 초록은 초록으로 보이므로 훑은 집합이 비어 있지 않음을 먼저 확인한다). **증거가 없으면 FAIL**로 쓰였다 — 「자명하다」는 통과 사유가 아니다. 뮤턴트도 둘이다: 거처를 α로 옮기면 갈래 ①이 RED, `ValidateSpecID` 거부 조건을 한 글자 바꾸면 갈래 ②가 RED(무변형 단언은 부재 단언이므로 변형이 실제로 잡힘을 보여야 비공허하다)
  - **AC-SLI-010 — 두 정규식 병존의 처분을 기록하고 드리프트를 가드한다.** 갈래 ①은 처분·근거·위험 이름 **셋 전부**를 요구하고(하나라도 없으면 FAIL — 「통합하지 않았다」만 적고 이유가 없으면 다음 저자가 판단을 재구성할 수 없다), 갈래 ②는 `spec_status.go:18`의 리터럴 `SPEC-[A-Z0-9-]+-[0-9]+`(v0.5.0 관측값) 바이트 동일 동결을 요구한다. **못 잡는 것도 적었다** — 이 가드는 리터럴만 보므로 그 심볼을 **호출하는 쪽**의 의미 변화는 잡지 못하고, 그 축은 AC-SLI-005의 뮤턴트가 덮는다
- **[HARD] `spec_status.go:18`은 통합하지 않는다 — 판단과 세 가지 이유를 기록했다.** ① 그 정규식은 **자기 자리에서 옳다**(용도가 git 커밋 메시지 스캔이고, 그 용도에는 앵커 없음이 맞다 — 통합은 고치는 것이 아니라 부수는 것이다). ② **반경이 이 카드의 것이 아니다**(`spec status`의 스캔 동작을 바꾸는 것은 REQ-SLI-004가 동결한 대조군 바깥의 세 번째 축을 여는 일이고, 그러면 어떤 측정도 귀속되지 않는다). ③ **위험의 모양이 다르다**(남는 위험은 「뜻이 다른 두 정규식이 공존한다」이지 「어느 하나가 틀렸다」가 아니다). 남는 위험에는 이름을 붙였다 — **동명-이의 재사용 드리프트** — 그리고 세 장치가 막는다: 이름이 다름(REQ-SLI-006) · AC-SLI-005 뮤턴트 · AC-SLI-010 리터럴 동결
- **수치 이동 — 옛 값을 지우지 않고 이동으로 적었다.** REQ **8 → 9** · AC 식별자 **11 → 13**(기준 11 → 13) · 이 SPEC의 lint 경고 **8 → 9**. 경고 수는 REQ 수를 따라간다(전부 `CoverageIncomplete`, REQ당 1건). 형제 SPEC도 같은 회차에 13 → 14로 움직였으므로 `spec.md §A`의 형제 인용도 함께 갱신했다. **묶음을 새로 만들어 수를 줄이지 않았다** — 이 파일의 규칙 6이 v0.2.0의 묶음을 「천장을 맞추기 위한 사후 회계」로 철회했으므로 같은 일을 반대 방향으로도 하지 않는다. 요구 9 · 기준 13 둘 다 Tier M 천장 16 이하
- **실측 재유도(파이프 없음, 이 트리)**: 수집기 정규식 grep → **9** · `grep -oE 'AC-SLI-[0-9]+[a-z]?' acceptance.md | sort -u | wc -l` → **13** · `moai spec lint .moai/specs/SPEC-SPEC-LINT-ID-ARG-001/spec.md` → rc=**0**, `0 error(s), 9 warning(s)`
- **[HARD] v0.4.0 자기 보고 간극 — 유지한다. 버전 번호가 담지 않은 변경을 주장하게 두지 않는다.** v0.4.0에서 **`plan.md`가 내용 변경 없이 버전만 올랐다** — 그 회차의 blocking(I-1)은 `acceptance.md`와 `spec.md`에만 닿았다. 이 사실을 지우지 않고 `spec.md` HISTORY의 v0.5.0 행에 다시 실었다. **v0.5.0에서는 네 아티팩트 전부에 내용 변경이 있다**(아래 항목) — 같은 간극이 반복되지 않았음을 확인해 적는다
- **이번 회차에 변경한 아티팩트**: `spec.md`(HISTORY · §A 경고 수 · §C REQ-SLI-009 · §D 포인터 · §F · §G #4·#5 · §H 결재) · `plan.md`(§B D4 결재 배너 + 반경 · §C M3 · §E 자가 검증 · §F 안티패턴 4건) · `acceptance.md`(§A 규칙 6·인용 주석 · §B 기준 12·13 신설 · §C 완료 정의 · §D 대응표) · `progress.md`(이 절). **버전만 오르고 내용이 그대로인 파일은 없다**
- **남은 운영자 결재 0건.** 형제 SPEC의 미해결 2도 같은 회차에 닫혔다. **Implementation Kickoff Approval은 여전히 별개의 미통과 게이트**다
- **판정문의 do-not-disturb 항목 무이동 확인**: `AC-SLI-005`의 픽스처 `SPEC-A-1`과 그 선택 근거 **미변경**(뮤턴트가 α·β 어느 거처에서도 발화하므로 거처 결재가 이 기준을 흔들지 않는다 — 그 사실을 `spec.md §H`와 `acceptance.md §D` 주석에 명시했다) · 감사 대안 `SPEC--1`이 두 패턴 모두에서 기각돼 뮤턴트가 발화하지 못했을 것이라는 **근처-실패 기록(§A.1) 미변경** · I-1의 「귀속되지 않은 커밋 집합 직접 판독」 절차와 기각한 대안(N-2 재발) **미변경** · 명령 ①·②와 `-- internal/cli/` 한정 **미변경** · Tier M 재분류의 실질 근거 4항(§G #5) **미변경**(수치 이동 주석만 병기) · 바이트 동일 판독기·BINLAG 열거는 형제 SPEC 소관으로 이 카드가 건드리지 않았다

## §E.2 Run-phase Evidence

### M-B1 — 세 인자 모양의 해석 (plan §C M1-M5를 한 커밋으로 묶어 착지)

**[HARD] 이 절은 소급 기록이다.** 구현은 커밋 `2cfccd4eb`(`feat(SPEC-SPEC-LINT-ID-ARG-001): M-B1 resolve three argument shapes for spec lint (t518)`)로 이미 착지했으나 그 회차에 §E.2/§E.3이 쓰이지 않았다. 아래에서 **이 실행에서 직접 관측한 것**과 **커밋된 증거 파일을 다시 읽은 것**을 구분해 적는다. 재판독은 새 측정이 아니다 — 출처 파일 경로를 항상 함께 적는다.

- 기록 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518` · 브랜치 `WT-spec-lint-axes` · HEAD `a4fbaeb82` (이 절을 쓸 때 `git rev-parse --show-toplevel` / `--short HEAD` 로 확인)
- 구현 커밋: `2cfccd4eb` — `internal/cli/spec_lint.go`(+129) · `internal/cli/specid/specid.go`(신규) · `internal/cli/spec_lint_test.go`(451줄) · `internal/cli/specid/specid_shape_test.go` + 증거 15파일(`.moai/reports/t518/mb1-*`)
- **[HARD] 줄 인용은 트리와 SHA에 매인다.** 아래 모든 `파일:줄` 인용은 트리 `.claude/worktrees/t518` · HEAD `a4fbaeb82` 기준이다. 다른 트리나 이후 커밋에서 같은 `파일:줄`은 다른 것을 가리킬 수 있다.

#### 마일스톤 대조 — 계획 M1-M5 vs 실제 착지

| plan §C | 계획된 산출물 | 실제 착지 | 근거 |
|---|---|---|---|
| M1 | 대조쌍 하네스 + RED + 수리 전 동작 기록 | 착지 (커밋 `2cfccd4eb` 안) | `mb1-red.txt` · `mb1-before-tests.txt` · `mb1-before-head.txt`(`97f9012…`) |
| M2 | 판별식 + 해석기 GREEN | 착지 | `internal/cli/spec_lint.go` · `mb1-green-cli.txt` |
| M3 | 진단 + 앵커 불변식 + β 거처 + 전/후 무변형 | 착지 | `internal/cli/specid/specid.go` · `mb1-green-specid.txt` · `mb1-after-tests.txt` |
| M4 | 도움말 세 모양 명시 | 착지 | `mb1-green-cli.txt` `TestSpecLintHelp_NamesThreeArgumentShapes` PASS |
| M5 | 혼합 인자 + 반경 격리 두 명령 | **부분 착지** — 혼합 인자는 착지, **반경 격리 두 명령은 이 실행에서 처음 실행됐다** | `mb1-mutants.txt` M8 · 본 절 주 ③ |

다섯 마일스톤이 **한 커밋**으로 눌렸다. 커밋 단위가 마일스톤 단위와 어긋나면 어느 변경이 어느 기준을 만족시켰는지가 커밋 이력에서 복원되지 않는다 — 그것을 이 표가 대신한다.

#### AC 판정표 (식별자 13개 전수)

판정 근거의 성질을 열로 나눴다. `RED` 열의 「없음」은 결함이 아니라 **부재 가드**라는 뜻이다 — 「이 동작이 바뀌면 안 된다」류의 기준은 수리 전에도 초록이므로 RED-now 로 채택 판정할 수 없고, 뮤턴트가 유일한 비공허성 증거다.

| AC | 검증 수단 | RED (`mb1-red.txt`) | GREEN | 뮤턴트 | 이 실행 재측정 | 판정 |
|---|---|---|---|---|---|---|
| AC-SLI-001a | `TestSpecLint_IDArg_ResolvesToSpecMD` (`spec_lint_test.go:167`) | FAIL — `ParseFailure count = 1, want 0` | `mb1-green-cli.txt` PASS | M1 CAUGHT | PASS (0.36s) | **PASS** |
| AC-SLI-001b | `TestSpecLint_IDArg_EqualsPathForm` (`spec_lint_test.go:188`) | FAIL — `exit codes differ: id=1 path=0` | `mb1-green-cli.txt` PASS | M1 CAUGHT | PASS (0.50s) | **PASS** |
| AC-SLI-001c | `TestSpecLint_IDArg_FromSubdirectory` (`spec_lint_test.go:216`) | FAIL — `ParseFailure count = 1, want 0 from a subdirectory` | `mb1-green-cli.txt` PASS | M1 · M9 CAUGHT | PASS (0.24s) | **PASS** |
| AC-SLI-002 | `TestSpecLint_DirectoryArg` (`spec_lint_test.go:233`) | FAIL — `ParseFailure count = 1 for directory form` | PASS | M2 CAUGHT | PASS (0.39s) | **PASS** |
| AC-SLI-003 | `TestSpecLint_UnresolvableID_IsArgumentError` (`spec_lint_test.go:254`) | FAIL — `exit code = 1, want 3` | PASS | M3 · M14 CAUGHT | PASS (0.00s) | **PASS** |
| AC-SLI-004a | `TestSpecLint_PathForm_Unchanged` (`spec_lint_test.go:275`) | FAIL — **다른 이유로**(주 ①) | PASS | **M4 NOT CAUGHT** · M4b CAUGHT | PASS (0.16s) | **PASS (주 ①·②)** |
| AC-SLI-004b | `TestSpecLint_MissingPathArg_StillParseFailure` (`spec_lint_test.go:299`) | **없음 — PASS** (부재 가드) | PASS | M5 CAUGHT | PASS (0.00s) | **PASS (뮤턴트 채택)** |
| AC-SLI-005 | `TestSpecLint_AnchorInvariant_SPEC_A_1` (`spec_lint_test.go:321`) | **없음 — PASS** (부재 가드) | PASS | M6 CAUGHT | PASS (0.00s) | **PASS (뮤턴트 채택)** |
| AC-SLI-006 | `TestSpecLintHelp_NamesThreeArgumentShapes` (`spec_lint_test.go:344`) | FAIL — `help does not mention "SPEC-ID"` | PASS | M7 CAUGHT | PASS (0.00s) | **PASS** |
| AC-SLI-007 | `TestSpecLint_MixedArgs` (`spec_lint_test.go:358`) | FAIL — `ParseFailure count = 1, want 0` | PASS | M8 CAUGHT | PASS (0.36s) | **PASS** |
| AC-SLI-008 | git 명령 2개 (테스트 없음) | 해당 없음 | 해당 없음 | 미발화(주 ④) | **이 실행에서 처음 실행** | **PASS (주 ③·④)** |
| AC-SLI-009 갈래 ① | `TestShapeFunction_HomeAndName` (`spec_lint_test.go:427`) | FAIL — `shape function is not declared in internal/cli/specid/specid.go` | PASS | M10 CAUGHT | PASS (0.00s) | **PASS** |
| AC-SLI-009 갈래 ② | 전/후 두 트리의 `go test` 출력 대조 | 해당 없음 | `mb1-before-tests.txt` / `mb1-after-tests.txt` | M11 CAUGHT | 재현 창 닫힘(주 ⑤) | **PASS-WITH-DEBT (주 ⑤)** |
| AC-SLI-010 갈래 ① | 문서 — `spec.md §H` | 해당 없음 | 해당 없음 | 해당 없음 | 이 실행에서 재판독 | **PASS (주 ⑥)** |
| AC-SLI-010 갈래 ② | `TestSpecStatusIDPattern_LiteralFrozen` (`spec_lint_test.go:411`) | **없음 — PASS** (부재 가드) | PASS | M12 CAUGHT | PASS (0.00s) | **PASS (뮤턴트 채택)** |

**[HARD] 배차문이 제기한 두 의심의 처분 — 하나는 오해였고 하나는 사실이다.**

- **`AC-SLI-001b` 는 실제로 덮여 있다.** 배차문은 「mb1 보고 파일에 나타나지 않는다」고 관측했다. 그 관측 자체는 참이다 — `mb1-*.txt` 는 **테스트 이름**만 싣고 AC 식별자를 싣지 않기 때문이다. AC 식별자는 소스 주석에만 있다(`internal/cli/spec_lint_test.go:186`, 트리 `t518` · HEAD `a4fbaeb82`). 이 실행에서 `/usr/bin/grep -n 'func Test\|AC-SLI' internal/cli/spec_lint_test.go` 로 대응을 재유도했고, 001b ↔ `TestSpecLint_IDArg_EqualsPathForm` 이 성립한다. **보고 파일의 침묵은 기준의 부재가 아니라 그 파일이 다른 축을 적는다는 뜻이었다.**
- **`AC-SLI-008` 은 실제로 미판정이었다.** 테스트가 없고 커밋된 증거 파일도 없다. 이 실행에서 처음 실행했다(주 ③).

##### 주 ① — AC-SLI-004a 의 RED 는 겨냥한 결함이 아니라 픽스처 인구조사 불일치였다

`mb1-red.txt` 축자:

```
    spec_lint_test.go:279: path-form census moved:
         got: INFO/OwnershipTransitionUnreachable,INFO/StatusGitUnreachable,WARNING/CoverageIncomplete,WARNING/FrontmatterInvalid,…,WARNING/MissingExclusions
        want: WARNING/CoverageIncomplete
```

`want` 가 한 항목뿐이었다 — 대조군이 「움직이지 않았다」를 재기 전에 **기준값 자체가 틀려 있었다.** 즉 이 RED 는 수리 대상 결함의 증거가 아니다(wrong-reason red). 그 뒤 기준값이 고쳐져 GREEN 이 됐고, 대조군의 비공허성은 **RED 가 아니라 M4b 뮤턴트**가 세운다. 원 관측을 지우지 않고 이렇게 병기한다.

##### 주 ② — 못 잡은 뮤턴트 M4 를 남긴다

`mb1-mutants.txt` 축자:

```
MUTANT M4 — path-signal pre-exclusion removed, anchored shape intact (AC-SLI-004a)
command: go test ./internal/cli/ -run 'TestSpecLint_PathForm_Unchanged'
mutation landed in source: yes
tests matched by selector: 1
rc=0  => NOT CAUGHT
```

`if hasPathSignal(arg)` 를 `if false` 로 바꿔도 대조군이 초록을 유지한다. **이것은 가드의 결함이 아니라 방어가 두 겹이라는 사실의 관측이다** — 경로 신호 선행 배제가 사라져도 앵커된 모양 검사가 여전히 그 경로를 기각하므로 동작이 바뀌지 않는다. 두 겹을 함께 벗기는 M4b 는 잡힌다(`rc=1 => CAUGHT`, `path form exit code = 3, want the pre-repair 0 or 1`). **못 잡았다는 사실은 가드의 경계를 그리므로 지우지 않는다.**

##### 주 ③ — AC-SLI-008 을 이 실행에서 실행했다 (신규 측정)

트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518` · HEAD `a4fbaeb82`. 파이프 없이, rc 를 각각 읽었다.

**명령 ①** — `git log --format=%H --grep=SPEC-SPEC-LINT-ID-ARG-001 0b1e27877..HEAD` (rc=0):

```
2cfccd4eb304c9929bd108caabe9915004674b2e
5d495bafca56ec82f80a8b040cec56de081e39e0
```

`git show --name-only --format= 2cfccd4eb304c9929bd108caabe9915004674b2e` (rc=0) — 19줄, 그중 코드는 넷:

```
internal/cli/spec_lint.go
internal/cli/spec_lint_test.go
internal/cli/specid/specid.go
internal/cli/specid/specid_shape_test.go
```

(나머지 15줄은 전부 `.moai/reports/t518/mb1-*`.) `git show --name-only --format= 5d495bafca56ec82f80a8b040cec56de081e39e0` (rc=0) — 16줄, 전부 `.moai/reports/t518/` 와 두 SPEC 의 `.moai/specs/…`. **두 출력 어디에도 `internal/spec/` 로 시작하는 줄이 없다.**

**명령 ②** — 새로 도입된 finding 코드 **0개**. 이 카드가 `internal/cli/` 를 만진 커밋은 하나뿐이므로 그 커밋으로 범위를 잡았다(`git diff -U0 2cfccd4eb^ 2cfccd4eb -- internal/cli/`, rc=0, 추가 704줄). 생산 코드 쪽(`spec_lint.go` + `specid/`)에서 `^\+.*"[A-Z][A-Za-z]{3,}"` 매치 **0건**(grep rc=1). finding 코드 문자열이 나타나는 추가 줄 6개는 전부 **테스트 파일**의 `countCode(out, "ParseFailure")` 이며, 기존 코드를 **참조**할 뿐 새 코드를 도입하지 않는다.

**[HARD] 그리고 이 기준의 모집단 정의가 이 트리에서 깨져 있다 — 발견으로 남긴다.** `acceptance.md` 기준 11 의 전제 확인 절차는 「`git log --format='%H %s' 0b1e27877..HEAD` 의 잔여 집합(두 SPEC-ID 중 어느 것도 담지 않은 줄)이 비어 있음」을 요구한다. 그 절차가 쓰인 v0.4.0 시점에는 `HEAD == base` 라 범위가 0건이었다. **지금 그 범위는 285 커밋이다**(`git rev-list --count 0b1e27877..HEAD` → `285`) — 그 사이 이 브랜치가 `origin/develop` 을 흡수했고, 흡수된 남의 카드 커밋 수백 개가 모집단에 들어왔다. 잔여 집합은 결코 비지 않으며, 절차를 문자 그대로 따르면 **남의 커밋 수백 개의 `git show` 를 인용해야** 통과다. 문자 그대로의 절차는 실행 불가능하다.

**실질로 재측정했다.** 옳은 모집단은 「이 카드가 만든 커밋」이며, 그것은 `origin/develop` 에서 도달 불가능한 non-merge 커밋이다. `git log --format='%H %s' --no-merges origin/develop..HEAD` (rc=0) → **8줄**:

```
a4fbaeb824f0688bfb11051900f4f56cc7accde8 docs(SPEC-SPEC-LINT-BLIND-AXES-001): M-A3 third recount reading at the committed state (t518)
b65850dbedc3e12f11352122e0e244c63e9b4638 docs(SPEC-SPEC-LINT-BLIND-AXES-001): M-A3 corpus recount on the merged tree (t518)
fc0540f4156ca11b52b99ac9d760973ae2a036f5 feat(SPEC-SPEC-LINT-BLIND-AXES-001): M-A2b announce table rejection (t518)
2cfccd4eb304c9929bd108caabe9915004674b2e feat(SPEC-SPEC-LINT-ID-ARG-001): M-B1 resolve three argument shapes for spec lint (t518)
97f90128431a81e14247cd964ea0d4dd1bc5c57f docs(SPEC-SPEC-LINT-BLIND-AXES-001): correct the acceptance.md §C fixture defect (t518, v0.6.0)
458fc7ebcec1c95385ffa8c4e3131fea260068a6 feat(SPEC-SPEC-LINT-BLIND-AXES-001): M-A2 announce unjudgeable modality (t518)
6cfcfef00daf53d58b7261f58ec9bfb626effb9d feat(SPEC-SPEC-LINT-BLIND-AXES-001): M-A1 collect table-form REQ definitions (t518)
5d495bafca56ec82f80a8b040cec56de081e39e0 docs(t518): SPEC-SPEC-LINT-BLIND-AXES-001 + SPEC-SPEC-LINT-ID-ARG-001 plan-phase (v0.5.0)
```

**여덟 줄 전부가 두 SPEC-ID 중 하나를 담는다 — 잔여 집합이 비어 있다.** 전제는 성립한다. 크기를 비교하지 않고 집합 자체를 읽었으므로 v0.3.0 이 폐기한 개수 대조의 상쇄 문제는 여기에 없다.

**판정**: AC-SLI-008 **PASS**. 다만 **기준 문구의 모집단 정의(`0b1e27877..HEAD`)는 결함이며 `--no-merges origin/develop..HEAD` 로 고쳐져야 한다** — `acceptance.md` 는 이 에이전트의 소관이 아니므로 §E.3 에 manager-spec 후속(F-B1)으로 올린다.

##### 주 ④ — AC-SLI-008 의 뮤턴트는 발화된 적이 없다

기준이 요구하는 뮤턴트는 「이 카드의 커밋 하나에서 `internal/spec/` 아래 파일을 한 글자 바꾸면 명령 ①의 출력에 나타난다」이다. **이 뮤턴트는 커밋된 증거에도 없고 이 실행에서도 발화시키지 않았다** — 발화시키려면 착지한 커밋을 고쳐 써야 하고, 그것은 공유 브랜치의 이력 재작성이다. **미검증으로 남긴다.** (가드의 논리는 자명에 가깝다 — `git show --name-only` 는 그 커밋이 만진 파일을 전부 낸다 — 그러나 자명함은 이 카드가 통과 사유로 인정하지 않는 것이다.)

##### 주 ⑤ — AC-SLI-009 갈래 ②: 창은 닫혔고, 기준이 요구한 한 항목이 증거에 없다

**창이 닫혔다.** 「함수 추가 **전** 트리」는 `97f90128431a81e14247cd964ea0d4dd1bc5c57f`(`mb1-before-head.txt` 축자)이고, 그 시점의 `go test` 출력은 그 시점에만 채취 가능하다. 이 실행에서 재채취하지 않았다 — 커밋된 채취본을 인용한다.

`mb1-before-tests.txt`(전) 와 `mb1-after-tests.txt`(후) 를 대조하면 **18개 패키지 전부가 양쪽에서 `ok`** 이고 실패 집합이 양쪽 모두 공집합이다. 다섯 호출부가 사는 두 패키지는 양쪽에서 **캐시 없이 실제로 돌았다**:

| 패키지 | 전 (`mb1-before-tests.txt`) | 후 (`mb1-after-tests.txt`) |
|---|---|---|
| `internal/cli` | `ok … 404.484s` | `ok … 482.526s` |
| `internal/kanban` | `ok … 147.176s` | `ok … 145.593s` |

`mb1-after-tests.txt` 의 나머지 16줄은 `(cached)` 다. **캐시된 `ok` 는 과거 어느 시점의 통과를 말하지 그 실행이 일어났음을 말하지 않는다** — 이 카드가 `mb1-verify-uncached-summary.txt` 에 스스로 적은 규율이다. 다섯 호출부가 캐시되지 않은 두 패키지 안에 전부 있다는 점이 그 위험을 이 자리에서만 비껴간다.

**[HARD] 그러나 기준이 명시적으로 요구한 한 항목이 증거에 없다.** AC-SLI-009 갈래 ② 는 「이 다섯 경로가 인용된 출력에서 **실제로 실행됐음**을 함께 보인다 — 셀렉터가 0건을 골라도 초록은 초록으로 보이므로」를 요구한다. 커밋된 두 파일은 **패키지 단위 `ok` 줄만** 싣고 테스트 단위 출력을 싣지 않으므로, `spec_view.go:46` · `spec_status.go:72` · `spec_close.go:108` · `kanban/board_store.go:268,360` · `kanban/status_read.go:109` 의 다섯 호출부가 그 실행에서 지나갔다는 것이 **출력에서 보이지 않는다.** 패키지가 초록이라는 사실은 그 안의 특정 경로가 실행됐다는 사실보다 약하다.

**부분적으로 메우는 것**: 뮤턴트 M11(`ValidateSpecID` 가 `..` 를 더 이상 거부하지 않도록 훼손)이 CAUGHT 이며(`mb1-mutants.txt` — `tests matched by selector: 8`, `specid_test.go:39: ValidateSpecID("SPEC-..-001") = nil, want non-nil error`), 이는 **`ValidateSpecID` 의 거부 조건이 실제로 판정에 닿는다**는 것을 세운다. 그러나 M11 이 잡힌 자리는 `internal/cli/specid` 의 자기 테스트이지 **다섯 소비자 호출부가 아니다.** 무변형 단언의 비공허성은 서고, 「훑은 집합이 비어 있지 않다」는 서지 않는다.

**판정: PASS-WITH-DEBT.** 무변형 자체는 두 출력의 대조로 관측됐고 뮤턴트로 비공허하다. **미상환 부채**: 다섯 호출부의 실행 관측. 창이 닫혀 소급 채취가 불가능하므로, 상환은 후속 카드에서 **후 트리 단독으로** `go test -v -run` 셀렉터가 그 다섯 경로를 지나는 테스트를 실제로 골랐음을 보이는 형태여야 한다(전 트리 없이도 「훑은 집합이 비어 있지 않다」는 잴 수 있다).

##### 주 ⑥ — AC-SLI-010 갈래 ①: 세 요소 전수 확인 (이 실행에서 재판독)

`spec.md §H`(트리 `t518` · HEAD `a4fbaeb82`)에서 셋을 모두 읽었다.

1. **처분**: 「**[HARD] 이 카드는 `spec_status.go:18`을 통합하지 않는다** … 그대로 둔다.」
2. **근거**: 세 항목이 명시돼 있다 — ① 그 정규식은 자기 자리에서 옳다(git 커밋 메시지 스캔 용도이므로 앵커 없는 편이 맞다) ② 반경이 이 카드의 것이 아니다(REQ-SLI-004 가 동결한 대조군 바깥의 세 번째 축) ③ 위험의 모양이 다르다(「뜻이 다른 두 정규식이 있다」이지 「어느 하나가 틀렸다」가 아니다).
3. **위험의 이름**: **동명-이의 재사용 드리프트**.

세 요소가 모두 있으므로 **PASS**.

#### 뮤턴트 표 (14개 — 13 CAUGHT · 1 NOT CAUGHT)

출처: `.moai/reports/t518/mb1-mutants.txt`(커밋 `2cfccd4eb` 에 포함). **이 실행에서 재발화시키지 않았다 — 재판독이다.**

| 뮤턴트 | 훼손 내용 | 겨냥 AC | rc | 결과 |
|---|---|---|---|---|
| M1 | ID 해석기 제거 | 001a/001b/001c | 1 | CAUGHT |
| M2 | 디렉터리 분기 제거 | 002 | 1 | CAUGHT |
| M3 | rc=3 진단 제거 | 003 | 1 | CAUGHT |
| **M4** | 경로 신호 선행 배제만 제거(앵커 유지) | 004a | 0 | **NOT CAUGHT** (주 ②) |
| M4b | 선행 배제 제거 **+** 모양 완화 | 004a | 1 | CAUGHT |
| M5 | 모든 미해결 인자를 rc=3 으로 | 004b | 1 | CAUGHT |
| M6 | 판별식을 `spec_status.go:18` 의 앵커 없는 패턴으로 교체 | 005 | 1 | CAUGHT |
| M7 | 도움말을 `lint [spec.md...]` 로 되돌림 | 006 | 1 | CAUGHT |
| M8 | 첫 인자만 해석 | 007 | 1 | CAUGHT |
| M9 | ID 기준 디렉터리를 `findProjectRootFn` → cwd | 001c | 1 | CAUGHT |
| M10 | 모양 검사를 `spec_lint.go` 로컬 사본(후보 α)으로 이동 | 009 갈래 ① | 1 | CAUGHT |
| M11 | `ValidateSpecID` 가 `..` 를 거부하지 않게 | 009 갈래 ② 비공허성 | 1 | CAUGHT |
| M12 | `spec_status.go:18` 리터럴에 앵커 부착 | 010 갈래 ② | 1 | CAUGHT |
| M13 | 복사된 모양 리터럴을 `\d{3}` → `\d+` 로 확장 | REQ-SLI-006 드리프트 가드 | 1 | CAUGHT |
| M14 | 진단을 stderr 에 쓰지 않음(rc=3 유지) | 003 비공허성 | 1 | CAUGHT |

**발화되지 않은 뮤턴트 1건**: AC-SLI-008 의 `internal/spec/` 접촉 뮤턴트(주 ④).

#### 이 실행의 검증 (신규 측정 · 트리 `t518` · HEAD `a4fbaeb82`)

**① AC 겨냥 12개 테스트 — rc=0.** `go test -count=1 -v -timeout 10m ./internal/cli/ -run '<12개 이름>'` → `go test rc=0`, 전문 `.moai/reports/t518/mb1-rerun-targeted-cli.txt`:

```
--- PASS: TestSpecLint_IDArg_ResolvesToSpecMD (0.36s)
--- PASS: TestSpecLint_IDArg_EqualsPathForm (0.50s)
--- PASS: TestSpecLint_IDArg_FromSubdirectory (0.24s)
--- PASS: TestSpecLint_DirectoryArg (0.39s)
--- PASS: TestSpecLint_UnresolvableID_IsArgumentError (0.00s)
--- PASS: TestSpecLint_PathForm_Unchanged (0.16s)
--- PASS: TestSpecLint_MissingPathArg_StillParseFailure (0.00s)
--- PASS: TestSpecLint_AnchorInvariant_SPEC_A_1 (0.00s)
--- PASS: TestSpecLintHelp_NamesThreeArgumentShapes (0.00s)
--- PASS: TestSpecLint_MixedArgs (0.36s)
--- PASS: TestSpecStatusIDPattern_LiteralFrozen (0.00s)
--- PASS: TestShapeFunction_HomeAndName (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.422s
```

**셀렉터가 12건을 골랐다** — 위 `--- PASS` 12줄이 그 자체로 훑은 집합이 비어 있지 않다는 증거다.

**② `internal/cli/specid` 패키지 — rc=0.** `go test -count=1 -v -timeout 10m ./internal/cli/specid/` → `specid rc=0`, 전문 `.moai/reports/t518/mb1-rerun-specid.txt`:

```
--- PASS: TestHasCanonicalSpecIDShape (0.00s)
--- PASS: TestCanonicalShapeLiteral_MatchesInternalSpecSource (0.00s)
--- PASS: TestValidateSpecID (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli/specid	0.355s
```

**③ 범위 전체 — rc=1, 실패 1건이며 이 카드의 표면이 아니다.** `go test -count=1 -timeout 20m ./internal/cli/... ./internal/kanban/...` → `rc=1`, 전문 `.moai/reports/t518/mb1-rerun-scoped.txt`. 판정 줄 축자:

```
--- FAIL: TestGateCmd_SecondRunWaitsForFirst (5.27s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	489.022s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	8.317s
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	19.405s
ok  	github.com/modu-ai/moai-adk/internal/cli/pr	8.662s
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	11.095s
ok  	github.com/modu-ai/moai-adk/internal/cli/printer	6.193s
ok  	github.com/modu-ai/moai-adk/internal/cli/specid	11.376s
ok  	github.com/modu-ai/moai-adk/internal/cli/taskledger	7.856s
ok  	github.com/modu-ai/moai-adk/internal/cli/uikit	12.229s
ok  	github.com/modu-ai/moai-adk/internal/cli/update	10.377s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	9.450s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/deploy	7.565s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	11.232s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/plan	12.586s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	11.756s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	13.973s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	15.481s
ok  	github.com/modu-ai/moai-adk/internal/kanban	153.586s
FAIL
```

`--- FAIL` 은 **정확히 1줄**이다. 그 실패의 단언 메시지 축자:

```
    gate_lock_cli_test.go:227: the runs' execution windows overlap: second run's first executed step began 2026-09-08T07:07:25.907+09:00, first run's last executed step ended 2026-09-08T07:07:25.907+09:00
```

**두 타임스탬프가 밀리초까지 같다** — 단언이 「겹침」의 경계 조건(동일 시각)을 겹침으로 읽는다. 같은 트리에서 단독 재실행은 통과한다:

```
go test -count=1 -v -timeout 5m ./internal/cli/ -run 'TestGateCmd_SecondRunWaitsForFirst'   → rc=0
--- PASS: TestGateCmd_SecondRunWaitsForFirst (5.42s)
ok  	github.com/modu-ai/moai-adk/internal/cli	6.320s
```

(전문 `.moai/reports/t518/mb1-rerun-gatelock.txt`.)

**[HARD] 이것을 「기존 결함」이라고 단정하지 않는다 — 잰 것만 적는다.** 관측된 것은 셋이다: ① 이 카드의 diff 는 `spec_lint.go` · `spec_lint_test.go` · `specid/` 넷뿐이고 `gate_lock_cli_test.go` 나 게이트 락 경로를 포함하지 않는다(주 ③ 명령 ① 출력) ② 실패 단언이 시각 비교이고 두 값이 같다 ③ 단독 재실행이 통과한다. **이 카드의 착지 이전에 같은 실패가 있었는지는 재지 않았다** — base 트리에서 재현을 시도하지 않았으므로 「기존부터 있었다」는 미검증이다. 판정으로 적을 수 있는 것은 **부하 의존 타이밍 실패이며 이 카드가 만진 표면 밖**이라는 데까지다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: 2cfccd4eb304c9929bd108caabe9915004674b2e   # 구현 착지 커밋. 이 §E.2/§E.3 기록 자체는 후속 backfill 커밋
run_status: implemented-with-debt
ac_pass_count: 12            # 식별자 13개 중 12개 PASS
ac_pass_with_debt_count: 1   # AC-SLI-009 갈래 ② (다섯 호출부 실행 관측 미상환)
ac_fail_count: 0
mutants_total: 14
mutants_caught: 13
mutants_not_caught: 1        # M4 — 방어 두 겹 중 한 겹만 벗긴 변형
mutants_never_fired: 1       # AC-SLI-008 의 internal/spec 접촉 뮤턴트 (주 ④)
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not_performed        # push 하지 않음 (배차 지시)
l44_post_push_fetch: not_applicable
new_warnings_or_lints_introduced: not_measured   # golangci-lint 미실행 — E5 참조
cross_platform_build: not_measured               # GOOS 교차 빌드 미실행 — E2 참조
coverage: not_measured                           # -cover 미실행 — E3 참조
total_run_phase_files: 19                        # 커밋 2cfccd4eb 의 파일 수 (코드 4 + 증거 15)
m1_to_mN_commit_strategy: "plan §C M1-M5 를 단일 커밋 2cfccd4eb 으로 압축. 마일스톤↔커밋 대응은 커밋 이력에서 복원되지 않으며 §E.2 의 마일스톤 대조표가 그것을 대신한다."
verification_tree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518
verification_head: a4fbaeb82
verification_branch: WT-spec-lint-axes
```

### E1 — AC 판정 행렬

§E.2 「AC 판정표」가 정본이다. 요약: 식별자 13개 중 **PASS 12 · PASS-WITH-DEBT 1 · FAIL 0**.

### E2 — 크로스 플랫폼 빌드

**미측정.** 이 실행에서 GOOS 교차 빌드를 돌리지 않았다. 부재를 통과로 읽지 않는다.

### E3 — 커버리지

**미측정.** `-cover` 를 실행하지 않았다.

### E4 — 서브에이전트 경계

이 실행은 서브에이전트를 스폰하지 않았다. 해당 없음.

### E5 — lint

**미측정.** `golangci-lint` 를 실행하지 않았다. (`mb1-rerun-scoped.txt` 안에 `golangci-lint: skipped — none of its config files exist in the project directory` 라는 줄이 있으나 그것은 **게이트 테스트가 낸 출력**이지 이 실행의 lint 측정이 아니다 — 남의 출력을 자기 측정으로 인용하지 않는다.)

### E6 — push 상태

**push 하지 않았다** — 배차 지시. 커밋만 만든다.

### E7 — 상태 전이

`spec.md` frontmatter `status: draft → in-progress`, `updated` 갱신. **이 에이전트가 수행하는 유일한 상태 전이다.** `in-progress → implemented → completed` 는 manager-docs 소관이다. (`plan.md` / `acceptance.md` 는 `status:` 필드를 갖지 않아 전이 대상이 아니다.)

### manager-spec 후속 (이 에이전트의 소관 밖 — 계획 아티팩트)

| # | 아티팩트 | 발견 | 근거 |
|---|---|---|---|
| F-B1 | `acceptance.md` 기준 11 (AC-SLI-008) | 전제 확인의 **모집단 정의가 이 트리에서 실행 불가능**하다. `0b1e27877..HEAD` 는 `origin/develop` 흡수 이후 285 커밋이며 잔여 집합이 결코 비지 않는다. `--no-merges origin/develop..HEAD`(8커밋) 로 고쳐야 한다 | §E.2 주 ③ |
| F-B2 | `acceptance.md` 기준 12 (AC-SLI-009 갈래 ②) | 요구한 「다섯 호출부가 실제로 실행됐음」이 **전 트리 창이 닫힌 뒤에는 소급 충족 불가능**하다. 후 트리 단독으로 잴 수 있는 형태로 고치거나 부채로 명시해야 한다 | §E.2 주 ⑤ |
| F-B3 | `acceptance.md` 기준 8 (AC-SLI-004a) | M4 가 못 잡히는 것이 **방어 두 겹** 때문이라는 사실이 기준 문서에 없다. 다음 독자가 `NOT CAUGHT` 를 가드 결함으로 오독할 수 있다 | §E.2 주 ② |
| F-B4 | `spec.md` §A 또는 `acceptance.md` §D | AC 식별자가 **소스 주석에만** 있고 증거 파일에는 테스트 이름만 남는다. 두 축의 대응표가 아티팩트 어디에도 없어 배차 단계에서 AC-SLI-001b 를 미판정으로 오독하는 일이 실제로 일어났다 | §E.2 「배차문이 제기한 두 의심의 처분」 |

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
