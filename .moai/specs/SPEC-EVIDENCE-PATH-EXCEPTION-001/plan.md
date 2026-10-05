# SPEC-EVIDENCE-PATH-EXCEPTION-001 — 구현 계획

> 마일스톤은 **되돌리기 어려운 결정부터** 배열한다. M1의 `.gitignore` 형태·삽입 위치가 가장
> 바뀌기 쉽고 가장 비싸며(순서가 결과를 정한다), M2의 문언이 그 다음이다. M4~M5의 미러·재생성·
> 번역은 앞선 결정에서 기계적으로 따라 나오고, M6의 기록은 순수 추가다.

## §A 맥락

카드 t1039. 독트린이 `.moai/reports/<card-id>/`를 추적 인용 대상이라 단언하는데 `.gitignore:227`
블랭킷이 그것을 무시한다. 운영자 판정: **블랭킷은 유지, 문장을 좁히고 좁은 예외를 뚫는다.**
셋째 항목으로 `.gitignore:109`의 죽은 negation과 그 주석을 정리한다.

**Tier L** (`spec.md` §0의 판별식 — 수정 표면 **20파일**, REQ 18, AC 23). plan-auditor PASS 임계
0.85, 산출물 5종. *(iteration 2는 16으로 적었다 — 층별 행의 합은 20이며 iteration 3이 정정했다,
`spec.md` §1.3(d). Tier 결론 불변.)*

가장 가까운 실사례는 이 카드 자신의 절차에서 나왔다: iter1 plan-audit 판정서가 규정된 목적지에
쓰였는데 그 목적지가 무시되어 반출되지 않았다(`spec.md` §1.0, 측정 동반).

측정 기록: `research.md`(이 SPEC의 측정 정본) + `.moai/reports/t1039/lane-measurements.md`
(레인 저작, 읽기 전용) + `spec.md` §1. `.gitignore` 규칙 형태의 설계 근거는 `design.md`.

## §B 알려진 문제 (착수 전에 이미 아는 것)

1. **소박한 와일드카드는 누수한다(가설)** — `!reports/*/verdict.md`가 FORBIDDEN인
   `.moai/reports/plan-audit/`의 verdict를 되살릴 수 있다(`spec.md` §5).
2. **깊이 제한(가설)** — 한 단계만 매칭한다. 실재하는 깊이-2 판정서 1건
   (`lead/<batch>/verdict.md`)이 놓인다. 범위 밖으로 결정됐으나 주석에 남겨야 한다
   (REQ-EPE-011 / AC-EPE-021).
3. **[HARD] 1과 2는 픽스처에서 나온 가설이며 증거가 아니다** — 4줄짜리 `.gitignore`가 이
   트리 417줄 파일의 규칙 순서를 재현하지 않았다. 특히 **픽스처에는 무리 D(`:294`)가 없는데,
   제자리에서는 그 규칙이 `plan-audit/verdict.md`의 현재 판정을 내고 있다.** run 단계는
   제자리에서 재측정하고, 어떤 인수조건도 픽스처 결과를 baseline으로 인용하지 않는다
   (REQ-EPE-008). 픽스처의 역할은 「무엇을 재야 하는가」의 범위 지정뿐이다.
4. **폭이 미해소** — REQ-EPE-006은 운영자 소관이고 답이 없다. **plan 단계는 이 축이 열린 채로
   닫힌다.** 결정 의존은 AC-EPE-002·AC-EPE-013 둘뿐이고, REQ-EPE-010은 **참조 의존**(결정하지
   않음)이며 `spec.md` §3.4가 두 등급을 전수 열거한다.
5. **`check-ignore`의 두 함정 + `git status`의 세 번째 함정** — `-v`의 exit 0, 추적 파일 스킵,
   그리고 `--ignored` 없는 `status`가 무시된 경로에 대해 **구조상 침묵**한다는 것
   (`spec.md` §1.0·§1.2, `acceptance.md` §A).
6. **줄수는 트리마다 다르다** — 이 워크트리 417줄, `main` 트리 319줄. 줄번호를 트리 귀속 없이
   옮기면 다른 트리에서 조용히 빗나간다.
7. **docs-site 인벤토리는 로케일 내성이어야 한다** — 영어 토큰만 겨눈 grep은 12파일 중 3파일만
   본다. 이것이 iter1이 이 표면 전체를 놓친 경로다(`spec.md` §1.3(c)).

## §C 사전 점검 (M1 착수 전)

- [ ] `git rev-parse --show-toplevel` 로 작업 트리를 명명하고, 그 트리의 `.gitignore` 줄수를
      기록(이 워크트리 417, `main` 트리 319 — **같은 수치가 아니다**).
- [ ] `git status --porcelain -- .gitignore` 가 비어 있는가. **이 워크트리의 `.gitignore`가
      선행 시도에서 417줄→10줄 스텁으로 파괴된 전례가 있다**(파괴된 것은 primary 체크아웃이
      아니라 **이 워크트리의 사본**이다).
- [ ] `grep -n 'moai/reports' .gitignore` 를 다시 돌려 `spec.md` §1.5의 다섯 무리 표가 이
      시점 트리와 일치하는지 확인. **어긋나면 표를 먼저 고치고 나서 삽입 위치를 정한다.**
- [ ] REQ-EPE-006 폭의 상태 기록: 운영자 확정이 있으면 출처와 함께, 없으면 **미해소**로.
      미해소여도 M1의 나머지(누수 봉쇄·깊이·삽입 위치 측정)는 진행 가능하다 —
      폭은 **되살릴 파일명 목록**만 정하기 때문이다.
- [ ] `git ls-files .moai/reports | wc -l` 기준선 기록(현재 54).

## §D 제약

- **[HARD] `.gitignore` 실험은 제자리에서 하지 않는다.** 규칙 형태를 탐침할 때는 `.gitignore`의
  **사본**을 스크래치로 복사해 별도 트리에서 재거나, 최종 형태를 한 번에 적용하고
  `git status --porcelain -uall`로 검증한다. 「임시로 고쳤다가 되돌린다」는 금지다.
- **[HARD]** 모든 ignore 판별식은 `git check-ignore --no-index -q` 또는
  `git status --porcelain -uall`. `-v` 단독 판정 금지(`-v`는 **어느 규칙이 결정했는가**를 읽는
  용도로만).
- **[HARD]** 양성 대조 없는 무출력/`rc=1`은 근거가 아니다.
- **[HARD]** `.codex/agents/moai/*.toml` 손편집 금지 — `make agents-emit`.
- **[HARD]** 템플릿 측 문언은 `.moai/docs/template-internal-isolation-doctrine.md` §25.1 금지
  클래스를 담지 않는다(SPEC ID·REQ 토큰·내부 날짜·커밋 SHA·감사 인용 금지). **docs-site는 그
  뿌리 밖이므로 이 제약의 대상이 아니다** — 대신 4-로케일 동기화 독트린이 적용된다.
- **[HARD]** docs-site 편집은 네 로케일을 **같은 변경 안에서** 처리한다. 영어를 비-영어 페이지에
  그대로 넣지 않는다.

## §E 자체 검증

각 마일스톤 종료 시 `progress.md` §E.2에 5절 형식(Claim / Evidence / Baseline-attribution /
Gaps / Residual-risk)으로 기록한다. 인용하는 명령은 전부 이 트리, 그 시점 HEAD에서 실행된 것.

## §F 마일스톤

### M1 — `.gitignore` 예외 (되돌리기 가장 비쌈)

가장 먼저 오는 이유: 규칙 **순서**가 결과를 정하고, 삽입 위치가 바뀌면 M2~M5의 어떤 것도
그것을 되돌려 주지 않는다.

1. REQ-EPE-006 폭의 상태를 기록한다. **확정돼 있지 않아도 M1은 진행한다** — 폭은 되살릴
   파일명 목록만 정하고, 아래 2~7(형태·삽입 위치·누수 봉쇄·깊이·주석·충돌)은 폭과 무관하다.
   미해소면 negation 줄을 `verdict.md` 하나로 두고, 확정 시 같은 형태로 줄만 추가한다.
2. 예외 규칙 형태를 결정한다 — 후보와 판별식은 `design.md` §1. 기본은 와일드카드형
   (`!.moai/reports/*/` → `.moai/reports/*/*` → 파일명별 negation), 저장소가 이미 네 번 쓰는
   3단 관용구(무리 C)와 같은 골격이다.
3. **삽입 위치를 측정으로 정한다.** `spec.md` §1.5의 다섯 무리 전량을 놓고 본다. 최소 측정 항목:
   - `:227` 직후에 넣었을 때, **아래에 있는 무리 D(`:294-295`)가 `plan-audit/` 판정을 그대로
     유지하는가** — 착수 전 측정에서 `plan-audit/verdict.md`를 결정하는 것은 `:294`다.
   - `REQ-EPE-007`의 명시적 재-제외가 `:294`와 **중복인가, 아니면 삽입 위치 때문에 필요한가.**
     중복이면 넣지 않는다(무리 D 자체는 건드리지 않는다 — `spec.md` §7 범위 밖).
   - 무리 E(`:376`)는 깊이 1 전용이라 `<card>/verdict.md`에 닿지 않는다(측정 완료). 삽입
     위치 판단의 입력이 아니지만, **닿지 않는다는 사실을 §E.2에 적는다** — 적지 않으면 다음
     사람이 다시 잰다.
4. `plan-audit/` 누수 봉쇄를 검증한다(REQ-EPE-007 / AC-EPE-004). `-v`가 이름 붙인 **줄번호를
   기록**한다 — 삽입 전후로 결정 규칙이 바뀔 수 있다.
5. 깊이가 한 단계임을, 그리고 **놓치는 형태가 `reports/lead/<batch>/verdict.md`임을** 주석에
   적는다(REQ-EPE-011 / AC-EPE-021).
6. 기존 추적 54건과 충돌 없음 확인(REQ-EPE-009).
7. 제자리 재측정과 5절 형식 기록(REQ-EPE-008 / AC-EPE-006).

검증: `git status --porcelain -uall .moai/reports` + 파일별 `check-ignore --no-index -q`,
양성 대조 동반.

### M2 — 독트린 문장 좁히기 (사용자가 읽는 표면)

1. `.claude/rules/moai/core/agent-common-protocol.md` § Evidence export의 "The citation target
   is a tracked path — in this repository `.moai/reports/<card-id>/`" 를 판정서 한정으로 좁힌다
   (REQ-EPE-001).
2. `.claude/rules/moai/core/agent-common-protocol-reference.md` § Evidence export obligation의
   같은 문장 + § Export width를 좁힌다. **리터럴 grep으로 이 파일을 놓치기 쉽다**
   (`a **tracked** path` 볼드) — 반드시 포함한다.
3. 좁히기가 **도달하는** 표면 정정(REQ-EPE-002 / AC-EPE-012). **네** 프로브가 0이 되어야 하고,
   **대체 문언을 §E.2에 인용해야** 한다:
   - `.claude/agents/moai/manager-lead.md:152` — `M<n>.<AC-id>.log`에 붙은 직접 추적 주장
   - `.claude/agents/moai/manager-lead.md:154` — "and only the tracked path does".
     **이 문장이 `:161`의 접기 행(`M<n>-report.md`)을 추적 주장으로 만든다.** `:161` 자체는
     「tracked」라는 낱말을 담지 않으므로 낱말 grep으로는 잡히지 않는다 — iter1의 공허한
     인수조건이 나온 자리다.
   - `.claude/output-styles/moai/moai.md:354, 561` — `<check>.log`에 붙은
     `(tracked; exported before citing)` 주석
   - `.claude/agents/moai/manager-lead.md:63` — **형태 ii(디렉터리 단독)**: "export of the
     deciding lines to the tracked `.moai/reports/<card-id>/`". 파일이 아니라 **디렉터리**를
     추적으로 주석하며, `AC-EPE-012`의 **P-d**가 이 좌표를 겨눈다(iteration 3, D12). 같은
     문장이 템플릿 미러에도 1건 있으므로 M4가 함께 옮긴다.
4. 좁히기를 **견뎌야 하는** 표면 무회귀 확인(REQ-EPE-003 / AC-EPE-013): `plan-auditor.md:601`,
   `sync-auditor.md:108` — 이 둘은 진짜 판정서를 반출하므로 수정 대상이 아니다. 문언이 여전히
   참인지(폭 (A)) 또는 거짓으로 남는지(폭 (B))를 **측정하고 기록**한다.

### M3 — 죽은 negation 철회와 주석 정정

1. `.gitignore:109` `!.moai/reports/**/*.log` 제거(REQ-EPE-012).
2. `:107-108` 주석 정정(REQ-EPE-013) — 왜 제거됐는지 + 로그는 의도적으로 무시된다는 사실.
   「must resolve post-merge」처럼 무언가를 고쳤다고 읽히는 문장을 남기지 않는다.
3. 철회 후에도 M1의 판정서 예외가 깨지지 않음을 재측정(AC-EPE-009).

### M4 — 템플릿 미러와 기계 방출 (기계적)

1. `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
2. `internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md`
3. `internal/template/templates/.claude/output-styles/moai/moai.md`
4. `internal/template/templates/.claude/agents/moai/manager-lead.md`
5. `make agents-emit` → `internal/template/templates/.codex/agents/moai/*.toml` 재생성.
   **손편집 금지**(REQ-EPE-015). 재생성은 조건부가 아니라 **확실히 촉발된다**: 방출된 세
   toml 중 `manager-lead.toml` 하나가 `` tracked `.moai/reports/<card-id>/` `` 문구를 담고
   비-판정서 `M<n>-report.md`를 그렇게 주석한다. `plan-auditor.toml` / `sync-auditor.toml` 은
   같은 경로를 한 번씩 언급하지만 tracked 주장을 하지 않으므로 좁히기를 견딘다.
6. `make build` (`//go:embed all:templates`)
7. 템플릿 중립성 자체 점검(REQ-EPE-016 / AC-EPE-016) — `internal/template/templates/` 아래
   hunk에만 적용.

**검증 주의 — 소스↔소스 검사는 노후한 바이너리를 볼 수 없다.** `make build`의 선행
`agents-emit-check`는 커밋된 `.toml`을 `.claude` 미러의 방출 결과와 **소스 대 소스**로 비교하므로,
이미 빌드된 설치본이 옛 정의를 임베드한 채 도는 상황을 원리상 검출하지 못한다(별도 축은
`make embed-check` / `moai doctor --check "Agent Emit Embed"`). 이 카드는 `.gitignore`와 문서
산문만 만지므로 그 축은 촉발되지 않고 범위 밖이다 — **다만 run 단계가 Go 코드로 범위를 넓히면
이 축을 잊지 말고 다시 본다.** 이것은 요구사항이 아니라 run 단계용 상기다.

### M5 — docs-site 4-로케일 반영 (iteration 2에서 편입)

M2의 문언이 확정된 뒤에만 시작한다 — docs-site는 그 문장의 파생 출판물이다.

1. 인벤토리를 **로케일 내성** grep으로 다시 뽑는다(`acceptance.md` AC-EPE-022의 명령).
   착수 시점 baseline: 12파일 / 20행 / 4로케일 × 3페이지 계열
   (`advanced/{token-budget,agent-guide,manager-lead}.md`).
2. 각 행을 판정한다 — **판정서 파일을 이름 붙인 행은 그대로 두고**, `REQ-EPE-002`의 두 형태
   (i 비-판정서 파일명 / ii 파일명 없는 `<card-id>/` 디렉터리 단독)를 추적 경로로 서술하는
   행만 고친다(REQ-EPE-017).
   착수 시점 실측(iteration 3): 20행이 **S1 8행 + S2 12행으로 남김 없이 갈리고, 판정서 파일을
   이름 붙이는 행은 0행**이다 — 즉 **그대로 둘 행이 없고 20행 전부가 수정 대상**이다.
   S1 8행 = 네 로케일 `manager-lead.md`의 `:122`·`:123`. S2 12행 = 네 로케일의
   `agent-guide.md` 1행 + `token-budget.md` 2행. 줄번호는 로케일마다 다르므로
   (`ko/agent-guide.md`는 `:212`) 줄번호가 아니라 위 두 프로브로 다시 뽑는다.
   [HARD] 플레이스홀더 이름을 리터럴로 겨눈 grep(`M<n>` 등)을 쓰지 않는다 — 로케일이
   플레이스홀더 자체를 번역하므로 **조용히 0을 낸다**(`research.md` §5.4의 정정 기록).
   [HARD] **「남김 없이 갈린다」의 근거는 `8 + 12 = 20`이 아니다 (iteration 4, D16)** — 합은
   중복 1행과 미분류 1행이 함께 있어도 같은 값을 낸다. 근거는 잔차 `L-R1` = 0과 중복
   `L-R2` = 0의 직접 측정이고, 그 둘은 `AC-EPE-022`의 PASS 조건에 들어 있으므로 M5 종료
   시점에 **다시 잰다**. `S1 = 0 AND S2 = 0`만 확인하고 닫지 않는다.
   [HARD] **네 판정식은 `acceptance.md` §B.9의 판정식 원장(fenced 블록)에서 그대로 복사해
   실행한다 (iteration 4, D15)** — 표 셀의 이스케이프본을 복사하면 S1이 구성상 `0`을 내고,
   목표값이 `0`이라 **그 거짓 PASS가 아무 신호도 내지 않는다**.
3. 네 로케일을 **같은 변경 안에서** 고친다(REQ-EPE-018). 각 로케일 본문은 그 로케일의
   자연스러운 원어로 쓰고, 영어 문장을 비-영어 페이지에 넣지 않는다.
4. 패리티와 빌드를 검증한다(AC-EPE-023): 페이지 stem별 변경 파일 수가 **정확히 4**,
   warning-free 빌드.
5. 로케일 확인(비-영어 문장이 원어로 읽히는가)을 §E.2에 **기록**한다 — grep으로 재지 않는
   항목이므로 기록되지 않으면 PASS가 아니다.

**범위 경계**: `.moai/reports`를 언급하지만 추적 주장을 담지 않는 docs-site 페이지(41 − 12 =
29파일)는 손대지 않는다(`spec.md` §7).

### M6 — 선행 SPEC 충돌 기록 (순수 추가)

1. 이 SPEC의 `related_specs`에 `SPEC-EVIDENCE-CITATION-CANON-001` (frontmatter에 이미 있음).
2. `SPEC-EVIDENCE-CITATION-CANON-001/spec.md`:
   - frontmatter에 `partially_superseded_by: [SPEC-EVIDENCE-PATH-EXCEPTION-001]`,
     `related_specs`에 이 SPEC 추가
   - HISTORY에 **추가만 하는** 행 한 개
   - **요구사항 본문은 건드리지 않는다**(REQ-EPE-004 / AC-EPE-018)

## §G 안티패턴 (이 카드에서 실제로 일어났거나 일어나기 쉬운 것)

- **제자리 `.gitignore` 실험.** 전례가 있다: **이 워크트리의** 417줄 사본이 10줄 스텁이 됐고
  (primary 체크아웃이 아니다), 발견된 것은 가드가 아니라 「규칙 줄번호가 두 번의 읽기 사이에
  움직였다」는 관측이었다. 그 사건이 드러낸 「서브에이전트 쓰기 범위를 막는 장치가 없다」는
  축은 **이 SPEC의 범위 밖**이며 여기서 설계하지 않는다(`spec.md` §7).
- **규칙 순서를 부분만 보고 삽입 위치 정하기.** iter1의 분석은 다섯 무리 중 둘만 담았고,
  빠진 `:294`가 지금 `AC-EPE-004`의 대상 경로를 결정하고 있었다. **삽입 지점보다 아래에 있는
  규칙이 결과를 뒤집는다.**
- **낱말 grep으로 「주장이 사라졌는가」를 재기.** `:161`은 `M<n>-report.md`를 추적 인용 경로로
  제시하면서 「tracked」라는 낱말을 담지 않는다. 낱말을 겨눈 판정식은 **착수 전부터 초록**이었다
  (iter1 AC-EPE-012). 주장을 만드는 **문장**을 겨눈다.
- **영어 토큰으로 다국어 표면 세기.** docs-site 12파일 중 영어 토큰에 걸리는 것은 3파일뿐이다.
  0에 가까운 수는 부재가 아니라 **미측정**일 수 있다 — 양성 대조가 그것을 가른다.
- **폭 권고를 결정으로 읽기.** `spec.md` §3.3의 권고는 운영자 결정이 아니다.
- **픽스처 결과를 제자리 결과로 인용하기.** 픽스처 표가 이미 문서 안에 있어서 다시 재는 동기가
  사라진다 — 그래서 이것이 가장 자주 일어난다. 순서가 다르고, 픽스처에는 무리 D가 없다.
- **`-v`로 판정하기.** negation 적중에도 exit 0.
- **`--ignored` 없는 `git status`의 0행을 「문제 없음」으로 읽기.** 무시된 경로에 대해 구조상
  침묵한다(`spec.md` §1.0 실측).
- **추적 파일로 양성 대조 만들기.** `--no-index` 없으면 스킵돼 `rc=1` — 구조상 실패할 수 없는
  대조가 된다.
- **폭 넓히기가 안전해 보이는 착각.** 넓히면 블랭킷이 막으려던 것이 돌아온다.
- **완료 SPEC의 REQ 본문 고치기.** 기록이 사라진다.
- **Tier를 예산에 맞추려고 요구사항 지우기.** 예산 초과는 tier up 하거나 분할하라는 신호이지
  내용을 줄이라는 신호가 아니다(`spec.md` §0).

## §H 교차 참조

- `spec.md` §0(Tier 판별식), §1.0(자기-실사례), §1.5(규칙 순서 다섯 무리), §3(폭 판별식),
  §4(충돌 기록), §5(누수 픽스처)
- `design.md` — `.gitignore` 규칙 형태 후보와 삽입 위치 설계
- `research.md` — 이 SPEC의 측정 정본(명령·출력·양성 대조)
- `.moai/reports/t1039/lane-measurements.md`, `.moai/reports/t1039/plan-audit.md` — 읽기 전용
  (둘 다 `:227`로 무시되어 반출돼 있지 않다)
- `.moai/docs/audit-artifact-convention.md` §Where, §50-53(FORBIDDEN)
- `.moai/docs/docs-site-i18n-rules.md` — 4-로케일 동기화 의무
- `.claude/rules/moai/core/verification-claim-integrity.md` §1, §2
- `SPEC-EVIDENCE-CITATION-CANON-001`
