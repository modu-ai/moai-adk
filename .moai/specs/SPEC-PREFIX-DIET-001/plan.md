---
id: SPEC-PREFIX-DIET-001
title: "plan — 세션 시작 prefix 다이어트 2단계"
version: "0.3.0"
created: 2026-10-03
updated: 2026-10-03
---

# plan.md — SPEC-PREFIX-DIET-001

마일스톤은 **결정 위험이 큰 순서**로 둔다 — 값을 정하는 데 운영자·리더 판단이 남은 항목을 앞에, 기계적 축약을 뒤에 둔다. 시간 추정은 쓰지 않는다.

## §A. 맥락

- 작업 위치: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1450`, 브랜치 `WT-prefix-diet-stage2`, 앵커 `5d5ff1aae`(실행 시작 때 `git rev-parse --short HEAD` 로 재확인).
- 이 트리는 카드 t1449 의 설정 변경을 이미 병합했고(`5d5ff1aae`), 그 위에 쌓는다.
- 범위는 카드 t1450 의 (b) 출력 스타일 파일 축약과 (c) 에이전트 설명 상한뿐이다. 스킬 목록 예산 값은 **바꾸지 않는다**(리더 결정). 제외 목록은 `spec.md` §F 가 구속한다.
- 근거 수치와 한계는 `spec.md` §A.1·§H. 이 plan 의 숫자 중 `spec.md` §A.2 의 실측을 제외한 것은 모두 **초안**이며 M0 에서 확정한다.

## §B. 결정 목록 (결정 위험 순)

| # | 결정 | 상태 | 누가 |
|---|---|---|---|
| D1 | `skillListingBudgetFraction` — **변경하지 않음**. `0.01` 측정(−7,194, −4.7%, 단일 실행)은 운영자 결정 항목으로 기록만 한다 | 닫힘(리더) | — |
| D2 | `moai.md` §6 `Session Boundary Handoff` 를 동결에 포함 | 닫힘(리더 확정) | — |
| D3 | `[HARD]` 줄 압축 재작성 — **하지 않음**. `dropped`(rationale·example)/`verbatim` 만 | 닫힘(리더) | — |
| D4 | 파일별 축약 목표(초안, 파일 전체 UTF-16: `moai-easy.md` ≤ 21,000 · `moai-learn.md` ≤ 20,000 · `moai.md` ≤ 45,000)와 에이전트 설명 상한(초안: 합 ≤ 9,500 · 단일 ≤ 1,200). D3 때문에 목표는 M0 의 `rationale`+`example` 합계를 넘을 수 없다(REQ-PFD-002) | 열림 — 비차단 | 리더, M0 끝 |
| D5 | 로컬 에이전트 `description:` 이 템플릿과 이미 다른 2개(`manager-git`, `manager-spec`)를 어느 쪽 문구로 맞출지 | 열림 — 비차단 | 리더, M4 시작 |

초안 목표는 상한 측정(−11,033 토큰, 본문 전체 제거)의 일부만 노린다는 뜻이며, 근거가 없고 M0 의 측정이 대체한다. 재작성이 막혀 있어 `[HARD]` 줄이 많은 파일(`moai.md` 89줄)은 목표에 못 미칠 수 있고, 그 경우 REQ-PFD-002 가 목표를 낮춰 리더에게 보고하게 한다.

## §C. 사전 점검과 도구

- 앵커 재확인: `git rev-parse --short HEAD` → `5d5ff1aae`(plan 커밋 이후 HEAD 는 SPEC 디렉터리만 바꿨다), `git branch --show-current` → `WT-prefix-diet-stage2`.
- 현 기준 테스트 상태 기록(M0, 편집 전): `go test ./internal/template/ ./internal/config/ -count=1` 를 파일로 리다이렉트하고 종료 코드와 `FAIL` 이름 목록을 `progress.md` §E.2 에 축자 기록한다. 이후 모든 회귀 판정은 이 기록과의 차이로 한다.
- 크기 측정 명령(재사용, 파일 전체 UTF-16):

```bash
python3 -c "import sys;t=open(sys.argv[1],encoding='utf-8').read();print(len(t.encode('utf-16-le'))//2)" <file>
```

- 에이전트 `description:` 블록 추출(크기 측정과 로컬↔템플릿 대조에 재사용):

```bash
python3 -c "import re,sys;t=open(sys.argv[1],encoding='utf-8').read().split('---')[1];m=re.search(r'^description:(.*?)(?=^\S)',t,re.S|re.M);print(m.group(1) if m else '')" <file>
```

- **단위 추출기와 구속 토큰 개수 규칙은 이 SPEC 이 정의한다**(`spec.md` §B). `SPEC-ALWAYS-LOADED-BUDGET-001` 의 추출기는 이 트리에 없고(`git show WT-always-loaded-budget:<경로>` 로만 읽을 수 있다) 재사용에 기대지 않는다. 추출기는 `internal/template` 의 이 SPEC 테스트 헬퍼로 구현하고, 같은 규칙(구속 토큰 개수)을 AC-PFD-004 의 독립 `python3` 명령이 따로 구현해 두 구현이 서로를 대조한다.
- **점검기**: `.moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py`(이 plan 에 커밋됨). 실행: 워크트리 루트에서 `python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py <BASE>`. 변경·신규 파일이 허용 표면 밖이면 위반, 출력 스타일은 frontmatter 불변, 에이전트는 본문과 `description:` 외 frontmatter 불변. 허용 표면은 출력 스타일·에이전트·생성 Codex TOML(`internal/template/templates/.codex/agents/moai/*.toml`)·이 SPEC 의 테스트와 고정물·문서뿐이며, **설정 파일 두 개는 허용목록에 없어 어떤 변경이든 위반**이다(키 예외 없음). 허용목록에 새 경로가 필요하면(예: `make build` 가 추적 파일을 갱신) 그 경로와 이유를 `progress.md` 에 적고 점검기를 같은 커밋에서 고친다 — 조용히 넓히지 않는다.
- 무거운 측정(`claude -p`)은 실계정 호출이다. 한 조건당 3회로 제한하고 조건을 직렬로 돌린다.

## §D. 제약과 위험

- **prompt cache 무효화**: 출력 스타일·에이전트 설명은 세션 prefix 에 실린다. 편집은 편집 이후 세션 전체의 cache 를 한 번 다시 쓰게 한다. 착지는 리더의 일괄 push 배치 경계에서 하고, 다른 레인이 도는 트리의 `.claude/` 로컬 사본을 중간에 바꾸지 않는다(`cache-aware-execution.md` directive 3).
- **Codex·다른 하네스**: `.claude/output-styles/`, `.claude/agents/` 는 Claude 쪽 기구다. Codex 는 `AGENTS.md` 를 읽고 이 SPEC 은 그 파일을 건드리지 않으므로 영향 밖이라고 읽었다(측정 아님 — `spec.md` §H).
- **의미 손실**: 재작성을 금지(D3)하고 `dropped` 는 `rationale`·`example` 에만 허용한다. 남는 위험은 작성자가 `normative` 단위를 `rationale` 로 오분류해 지우는 경우이며 기계가 못 잡는다 — `dropped` 행 목록 반출과 sync-audit·리더 검토로 막는다.
- **t1469 와의 충돌**: 카드 t1469 가 세션 핸드오프 규칙을 다시 쓰면 동결 단위 고정물과 렌더 절이 함께 바뀐다. 고정물 갱신은 그 카드의 커밋이 맡는다.
- **측정 잡음**: 단일 실행 수치는 분산을 모른다. 이 SPEC 은 조건당 3회 실행으로 분산을 처음 잰다(REQ-PFD-012). 3회 범위가 기대 감소폭보다 넓으면 그 마일스톤의 토큰 효과는 "측정 불가"로 보고하고 크기 지표(UTF-16)만 주장한다.
- **생성 TOML 드리프트**: 설명 편집 뒤 `make agents-emit` 을 빠뜨리면 `make build` 가 실패한다(재현됨). M5 3b 가 막는다.
- **에이전트 사본 불일치**: 로컬 에이전트 본문은 12개 중 10개가 템플릿과 다르다(`spec.md` §A.2). 이 SPEC 이 만든 것이 아니므로 본문은 건드리지 않는다.
- **범위 점검기 오탐**: develop 을 흡수하면 앵커 기준 점검이 남의 변경을 위반으로 읽는다. 기준 SHA 를 읽는 시점에 `git merge-base develop HEAD` 로 다시 구한다(병합 전 평가 전용).

## §E. 자기 검증 (실행 단계가 보고할 것)

각 마일스톤 보고는 Claim / Evidence(명령+축자 출력) / Baseline-attribution(이 실행, 이 트리, HEAD SHA) / Gaps / Residual-risk 5개 절을 쓴다(`verification-claim-integrity.md` §3). 빠진 것은 Gap 으로 보고한다.

## §F. 마일스톤

### M0 — 측정 기준선과 원장 앞면 (편집 없음, D4 입력)

1. 앵커 재현: `acceptance.md` §D.2 프로토콜로 첫 턴 토큰 3회 측정. 3회 중 어느 것도 154,219 의 ±1% 안에 없으면 편집 전에 리더에게 보고(REQ-PFD-012).
2. 출력 스타일 3개의 파일 전체 UTF-16 크기와 `[HARD]` 줄 수 기록, 구속 토큰 개수 기록, 단위 추출기 구현, 원장 앞면(앵커의 단위별 `before_text`, 종류 분류) 생성. 단위별 `rationale`+`example` 합계(파일별 축약 가능 상한)를 계산해 기록한다. 원장 앞면은 **편집 커밋보다 앞선 별도 커밋**으로 착지한다 — 전/후 순서를 증명할 수 있는 것은 커밋 그래프뿐이다(`verification-claim-integrity.md` §2.3).
3. 에이전트 설명 크기 재측정과 로컬↔템플릿 `description:` 블록 `diff` 기록(D5 입력).
4. D4 목표를 리더가 확정하고 `progress.md` §E.2 에 기록한다.
5. 점검기를 앵커에서 돌려 `surface-guard=PASS` 를 기록한다.

### M1 — 스킬 목록 예산 키 가드 (변경 없음)

1. `0.01` 측정(147,025 vs 154,219, −7,194, −4.7%, 단일 실행)을 `progress.md` §E.2 에 **운영자 결정 항목**으로 기록한다. 값은 바꾸지 않는다.
2. AC-PFD-008 명령으로 두 파일의 키가 `0.02` 이고 설정 파일이 무수정임을 기록한다.

### M2 — 출력 스타일 `moai-easy.md` (배포 기본 스타일, 측정 가능한 효과)

1. RED: 예산 테스트·원장 테스트·동결 테스트·로컬라이제이션 표 패리티 테스트를 만든다. `moai-easy.md` 예산 상수를 D4 목표로 낮춘다(나머지 두 파일 상수는 앵커 크기로 둔다). 목표 미달이라 `TestOutputStylesCharBudget/moai-easy` 가 FAIL 하는 출력을 RED 로 기록한다.
2. GREEN: 원장에 행을 채우며 `rationale`·`example` 단위를 `dropped` 로 지운다(생존 단위 참조 필수). 후보 — §10 배너 예시(§7 견본의 반복), §12 FAQ, §13 교육 철학, §14·§15 의 중복 안내. 이 후보는 초안이며 원장이 판정한다. `[HARD]` 줄은 한 글자도 바꾸지 않는다.
3. 템플릿 편집 → `make build` → 로컬 사본 복사 → `diff -rq` 무출력.
4. 첫 턴 토큰 3회 측정(기본 스타일 사용자에게 효과가 드러나는 유일한 마일스톤). 점검기 PASS 기록.

### M3 — `moai-learn.md`

M2 와 같은 RED/GREEN 순서. 효과 측정은 `--settings` 에 `outputStyle` 을 해당 스타일 이름으로 지정해 앞/뒤를 같은 조건으로 잰다(기본 사용자 영향 없음을 기록).

### M4 — `moai.md` (가장 크고, 핸드오프 동결 단위 두 개를 품은 파일)

M2 와 같은 순서. 동결 단위 두 개(§6, §8)는 해시가 앵커와 같아야 한다. 동결 단위 앞뒤 경계(`---` 구분선, `## 7. Temp File Hygiene` 제목)는 건드리지 않는다. `[HARD]` 줄이 89개라 목표 도달이 어려울 수 있으며 그 경우 REQ-PFD-002 에 따라 목표를 낮춰 보고한다.

### M5 — 에이전트 설명 상한 (가치 가장 낮고 기계적), 이어서 마감 측정

1. D5 판정 반영. 로컬이 템플릿보다 새로운 문구를 가졌으면 템플릿에 먼저 옮긴다.
2. RED: `TestAgentDescriptionBudget`(합계·단일 상한 상수, D4)를 만들어 앵커에서 FAIL 하는 출력을 기록.
3. GREEN: 큰 설명부터(`manager-lead` 2,182 · `manager-docs` 1,552 · `manager-develop` 1,094 · `manager-spec` 1,022) 줄인다. 역할·트리거·`NOT for:` 는 남긴다(REQ-PFD-011). 편집한 에이전트의 `description:` 블록만 로컬 사본과 같게 맞춘다.
3b. **`make agents-emit`**: 설명을 줄이면 `make build` 의 `agents-emit-check` 가 golden 해시 불일치로 실패한다(이 plan 에서 재현 — `spec.md` §H). 설명 편집 뒤 `make agents-emit` 을 돌려 `internal/template/templates/.codex/agents/moai/<이름>.toml` 을 재생성하고 같은 커밋에 포함한다. 생성 파일은 점검기 허용목록(`codex-tomls`)에 있다. 이후 `make agents-emit-check` 와 AC-PFD-016 을 재관측한다. 로컬 `.codex/agents/moai` 는 이 트리에 없다.
4. 마감: 최종 트리에서 §D.2 프로토콜로 3회 측정, 마일스톤별 전/후 표 기록, 패키지 단위 회귀 대조, `make build`, 점검기 최종 PASS 와 양성 대조(AC-PFD-012), `acceptance.md` 의 모든 AC 를 명령 그대로 재실행해 축자 출력 기록. `dropped` 행 목록 반출.

## §G. 안티패턴

- 같은 커밋에 원장 앞면(기준선)과 본문 편집을 섞기.
- `[HARD]` 줄을 "의미를 보존하며" 줄이려는 시도 — 재작성은 이 SPEC 에서 금지다.
- `-run` 패턴이 0개를 고르고도 `ok` 로 보이는 출력을 초록으로 읽기 — 이름 붙은 `--- PASS: <이름> ` 줄만 근거로 한다.
- `outputStyle=default` 상한을 달성값처럼 인용하기. `skillListingBudgetFraction` 을 "손쉬운 절감"으로 슬쩍 바꾸기.
- 로컬 사본을 먼저 고치고 템플릿을 나중에 맞추기(Template-First 위반).
- 점검기 허용목록을 이유 기록 없이 넓히기. `git add -A` / `git add .` — 명시 경로로만 스테이징한다.

## §H. 상호 참조

- `spec.md` §A(증거·소스 오브 트루스) · §B(추출기 명세) · §C(요구사항) · §F(제외) · §H(Gaps)
- `acceptance.md` §D(AC 매트릭스) · §D.1(증거 원장) · §D.2(측정 프로토콜)
- `.claude/rules/moai/workflow/cache-aware-execution.md` directive 3 — 세션 로드 파일 편집은 배치 끝에
- `.claude/rules/moai/development/verification-completeness.md` §2 — RED-now/초록 경로 쌍 규율
- `SPEC-ALWAYS-LOADED-BUDGET-001` — 이 트리에 없음. `git show WT-always-loaded-budget:.moai/specs/SPEC-ALWAYS-LOADED-BUDGET-001/spec.md` 로만 인용(구속 원장 방식의 선례)
