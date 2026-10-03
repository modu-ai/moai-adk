---
id: SPEC-PREFIX-DIET-001
title: "plan — 세션 시작 prefix 다이어트 2단계"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
---

# plan.md — SPEC-PREFIX-DIET-001

마일스톤은 **결정 위험이 큰 순서**로 둔다 — 값을 정하는 데 운영자 판단이 필요하거나 의미 보존 판정이 걸린 항목을 앞에, 기계적 축약을 뒤에 둔다. 시간 추정은 쓰지 않는다.

## §A. 맥락

- 작업 위치: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1450`, 브랜치 `WT-prefix-diet-stage2`, 앵커 `5d5ff1aae`(실행 시작 때 `git rev-parse --short HEAD` 로 재확인).
- 이 트리는 카드 t1449 의 설정 변경을 이미 병합했고(`5d5ff1aae`), 그 위에 쌓는다.
- 범위는 카드 t1450 의 (b) 출력 스타일 본문, (c) 스킬 목록 예산과 에이전트 설명 상한뿐이다. 제외 목록은 `spec.md` §F 가 구속한다.
- 근거 수치(첫 턴 154,219 / 143,186 / 147,025)와 그 한계는 `spec.md` §A.1, §H 에 있다. 이 plan 의 숫자 중 `spec.md` §A.2 의 실측을 제외한 것은 모두 **초안**이며 M0 에서 확정한다.

## §B. 결정 목록 (결정 위험 순)

| # | 결정 | 누가 | 언제 | 되돌림 비용 |
|---|---|---|---|---|
| D1 | `skillListingBudgetFraction` 값 — 0.02 유지, 0.01, 또는 중간값. 측정 근거는 `spec.md` §A.1 의 −7,194(0.01)와 M1 의 발견 측정 | 운영자 | M1 시작 | 낮음(설정 한 줄) — 단, 사용자 가시 동작(스킬 발견)이 바뀐다 |
| D2 | `moai.md` §6 `Session Boundary Handoff` 를 동결에 포함할지 — 이 SPEC 의 기본은 포함(보수적 해석, `spec.md` §F) | 리더 | plan-audit 전 | 낮음 |
| D3 | 출력 스타일 압축 재작성을 구속 원장 조건으로 허용하는지 — `SPEC-ALWAYS-LOADED-BUDGET-001` 의 판정은 그 SPEC 한정이라 이 SPEC 에 대한 판정은 미확인 | 운영자 | plan-audit 전 | 중간(판정 없이는 `rewrite` 행을 만들 수 없고 `dropped`(rationale·example)만 남는다) |
| D4 | 파일별 축약 목표와 에이전트 설명 상한(초안: `moai-easy.md` ≤ 21,000 · `moai-learn.md` ≤ 20,000 · `moai.md` ≤ 45,000 UTF-16, 에이전트 설명 합 ≤ 9,500·단일 ≤ 1,200) | 리더 | M0 끝 | 낮음 |
| D5 | 로컬 에이전트 `description:` 이 템플릿과 이미 다른 2개(`manager-git`, `manager-spec`)를 어느 쪽 문구로 맞출지 | 리더 | M5 시작 | 중간 |

초안 목표는 상한 측정(−11,033 토큰, 본문 전체 제거)의 일부만 노린다는 뜻이며, 근거는 없고 M0 의 원장 하한 측정이 대체한다.

## §C. 사전 점검

- 앵커 재확인: `git rev-parse --short HEAD` → `5d5ff1aae`, `git branch --show-current` → `WT-prefix-diet-stage2`.
- 현 기준 테스트 상태 기록(M0, 편집 전): `go test ./internal/template/ ./internal/config/ -count=1` 를 파일로 리다이렉트하고 종료 코드와 `FAIL` 이름 목록을 `progress.md` §E.2 에 축자 기록한다. 이후 모든 회귀 판정은 이 기록과의 차이로 한다(앵커에서 이미 빨간 테스트는 이 카드 책임이 아니다).
- 크기 측정 명령(재사용):

```bash
python3 -c "import sys;t=open(sys.argv[1],encoding='utf-8').read();print(len(t.encode('utf-16-le'))//2)" <file>
```

- 에이전트 `description:` 블록 추출(크기 측정과 로컬↔템플릿 대조에 재사용, 이 plan 이 앵커에서 돌린 식과 같다):

```bash
python3 -c "import re,sys;t=open(sys.argv[1],encoding='utf-8').read().split('---')[1];m=re.search(r'^description:(.*?)(?=^\S)',t,re.S|re.M);print(m.group(1) if m else '')" <file>
```

- 무거운 측정(`claude -p`)은 실계정 호출이다. 한 조건당 3회로 제한하고 한 번에 한 조건씩 직렬로 돌린다(병렬은 서로의 prompt cache 에 영향을 준다).

## §D. 제약과 위험

- **prompt cache 무효화**: 출력 스타일·설정·에이전트 설명은 세션 prefix 에 실린다. 이 파일들의 편집은 편집 이후 세션 전체의 cache 를 한 번 다시 쓰게 만든다. 착지는 리더의 일괄 push 배치 경계에서 하고, 다른 레인이 도는 트리의 `.claude/` 로컬 사본을 중간에 바꾸지 않는다(`cache-aware-execution.md` directive 3).
- **Codex·다른 하네스**: `.claude/output-styles/`, `.claude/settings.json`, `.claude/agents/` 는 Claude 쪽 기구다. Codex 는 `AGENTS.md` 를 읽고 이 SPEC 은 그 파일을 건드리지 않으므로 영향 밖이라고 읽었다(측정 아님 — `spec.md` §H).
- **의미 손실**: 압축 재작성이 `[HARD]` 의무의 주체·강도·예외를 깎을 위험. 완화 = 구속 원장(AC-PFD-003·004).
- **t1469 와의 충돌**: 카드 t1469 가 세션 핸드오프 규칙을 다시 쓰면 동결 단위 고정물과 렌더 절이 함께 바뀐다. 고정물 갱신은 그 카드의 커밋이 맡는다(이 SPEC 은 고정물을 앵커 상태로 둔다).
- **측정 잡음**: 단일 실행 수치는 분산을 모른다. 이 SPEC 은 조건당 3회 실행으로 분산을 처음 잰다(REQ-PFD-013). 3회 범위가 마일스톤의 기대 감소폭보다 넓으면 그 마일스톤의 토큰 효과는 "측정 불가"로 보고하고 크기 지표(UTF-16)만 주장한다.
- **스킬 발견 품질**: `skillListingBudgetFraction` 을 낮추면 목록의 스킬 설명이 잘려 발견이 나빠질 수 있다. 이 SPEC 의 대리 지표(스킬 이름 나열 수)는 이름 누락만 잡고 설명 품질 저하는 못 잡는다 — 한계를 `progress.md` 에 같이 적는다.
- **에이전트 사본 불일치**: 로컬 에이전트 본문은 12개 중 10개가 템플릿과 다르다(`spec.md` §A.2). 이 SPEC 이 만든 것이 아니므로 본문은 건드리지 않는다.

## §E. 자기 검증 (실행 단계가 보고할 것)

각 마일스톤 보고는 Claim / Evidence(명령+축자 출력) / Baseline-attribution(이 실행, 이 트리, HEAD SHA) / Gaps / Residual-risk 5개 절을 쓴다(`verification-claim-integrity.md` §3). 측정 항목마다 명령, 축자 출력, exit 코드, HEAD SHA 를 적고, 빠진 것은 Gap 으로 보고한다.

## §F. 마일스톤

### M0 — 측정 기준선과 원장 앞면 (편집 없음, 결정 입력)

1. 앵커 재현: `acceptance.md` §D.2 프로토콜로 첫 턴 토큰 3회 측정. 3회 중 어느 것도 154,219 의 ±1% 안에 없으면 편집 전에 리더에게 보고(REQ-PFD-014).
2. 출력 스타일 3개의 크기(UTF-16)와 `[HARD]` 줄 수 기록, 원장 앞면 생성(단위 추출기 + `before_text` 채움). 원장 앞면은 **편집 커밋보다 앞선 별도 커밋**으로 착지한다 — 전/후 순서를 증명할 수 있는 것은 커밋 그래프뿐이다(`verification-claim-integrity.md` §2.3). 같은 커밋에 구현과 기준선을 섞지 않는다.
3. 에이전트 설명 크기 재측정과 로컬↔템플릿 `description:` 블록 `diff` 기록(D5 입력).
4. `SPEC-ALWAYS-LOADED-BUDGET-001` 의 단위 추출기가 이 트리에 있는지 확인한다. 있으면 재사용하고(복제하지 않는다 — 단순성 사다리 2단), 없으면 이 SPEC 의 테스트 헬퍼로 `internal/template` 안에 둔다.
5. 위 측정으로 D4 의 목표를 리더가 확정하고 `progress.md` §E.2 에 기록한다.

### M1 — 스킬 목록 예산 (결정 위험 최고)

1. D1 운영자 판정을 `progress.md` §E.2 에 기록한다(판정이 없으면 이 마일스톤은 키를 그대로 두고 측정만 한다 — REQ-PFD-009).
2. `acceptance.md` §D.2 의 스킬 발견 프로브를 변경 전(0.02)과 후보값에서 3회씩 측정해 기록한다(REQ-PFD-010).
3. `internal/template/templates/.claude/settings.json.tmpl` 414행을 먼저 고치고 `make build`, 그다음 로컬 `.claude/settings.json` 410행을 키 단위로 맞춘다. 파일 전체 `diff` 는 쓰지 않는다(`spec.md` §A.3).
4. 첫 턴 토큰 3회 측정.

### M2 — 출력 스타일 `moai-easy.md` (배포 기본 스타일, 측정 가능한 효과)

1. RED: 예산 테스트·원장 테스트·동결 테스트·로컬라이제이션 표 패리티 테스트를 만든다. `moai-easy.md` 예산 상수를 D4 목표로 낮춘다(나머지 두 파일 상수는 앵커 크기로 둔다). 목표 미달이라 `TestOutputStylesCharBudget/moai-easy` 가 FAIL 하는 출력을 RED 로 기록한다.
2. GREEN: 구속 원장에 행을 채우며 본문을 줄인다. `dropped` 는 `rationale`·`example` 에만, 생존 단위 참조와 함께. 후보 — §10 배너 예시(§7 견본의 반복), §12 FAQ, §13 교육 철학, §14·§15 의 중복 안내. 이 후보는 초안이며 원장이 판정한다.
3. 템플릿 편집 → `make build` → 로컬 사본 복사 → `diff -rq` 무출력.
4. 첫 턴 토큰 3회 측정(기본 스타일 사용자에게 효과가 드러나는 유일한 마일스톤).

### M3 — `moai-learn.md`

M2 와 같은 RED/GREEN 순서. 효과 측정은 `--settings` 에 `outputStyle` 을 `MoAI-Learn` 으로 지정해 앞/뒤를 같은 조건으로 잰다(기본 사용자 영향 없음을 기록).

### M4 — `moai.md` (가장 크고, 핸드오프 동결 단위 두 개를 품은 파일)

M2 와 같은 순서. 동결 단위 두 개(§6, §8)는 해시가 앵커와 같아야 한다. 동결 단위 앞뒤 경계(`---` 구분선, `## 7. Temp File Hygiene` 제목)는 건드리지 않는다. 효과 측정은 `outputStyle` 을 `MoAI` 로 지정한다.

### M5 — 에이전트 설명 상한 (가치 가장 낮고 기계적)

1. D5 판정 반영. 로컬이 템플릿보다 새로운 문구를 가졌으면 템플릿에 먼저 옮긴다.
2. RED: `TestAgentDescriptionBudget`(합계·단일 상한 상수, D4)를 만들어 앵커에서 FAIL 하는 출력을 기록.
3. GREEN: 큰 설명부터(`manager-lead` 2,182 · `manager-docs` 1,552 · `manager-develop` 1,094 · `manager-spec` 1,022) 줄인다. 역할·트리거·`NOT for:` 는 남긴다(REQ-PFD-012).
4. 편집한 에이전트의 `description:` 블록만 로컬 사본과 같게 맞춘다.
5. 첫 턴 토큰 3회 측정. 효과가 3회 범위 안이면 "측정 불가"로 쓴다(REQ-PFD-013 의 분산 규칙).

### M6 — 마감 측정과 기록

최종 트리에서 §D.2 프로토콜로 3회 측정하고, `progress.md` §E.2 에 마일스톤별 전/후 표를 기록한다. 패키지 단위 회귀(`go test ./internal/template/ ./internal/config/ -count=1`)를 앵커 기록과 대조하고 `make build` 를 돌린다. `acceptance.md` 의 모든 AC 를 명령 그대로 재실행해 축자 출력을 기록한다.

## §G. 안티패턴

- 같은 커밋에 원장 앞면(기준선)과 본문 편집을 섞기.
- `-run` 패턴이 0개를 고르고도 `ok` 로 보이는 출력을 초록으로 읽기 — 이름 붙은 `--- PASS: <이름> ` 줄만 근거로 한다.
- `outputStyle=default` 상한을 달성값처럼 인용하기.
- 로컬 사본을 먼저 고치고 템플릿을 나중에 맞추기(Template-First 위반).
- `git add -A` / `git add .` — 명시 경로로만 스테이징한다.

## §H. 상호 참조

- `spec.md` §A(증거·소스 오브 트루스) · §C(요구사항) · §F(제외) · §H(Gaps)
- `acceptance.md` §D(AC 매트릭스) · §D.2(측정 프로토콜)
- `.claude/rules/moai/workflow/cache-aware-execution.md` directive 3 — 세션 로드 파일 편집은 배치 끝에
- `.claude/rules/moai/development/verification-completeness.md` §2 — RED-now/초록 경로 쌍 규율
- `SPEC-ALWAYS-LOADED-BUDGET-001`(브랜치 `WT-always-loaded-budget`) — 구속 원장 방식의 선례
