# t523 — lane verdict (미러 스킬 12디렉터리 71건 — 지시 중립화 + 전제 표본 확인)

card: t523 (Class C, Tier M)
worktree: .claude/worktrees/t523
branch: WT-mirror-skill-neutrality
base: 7e1859c28 (로컬 develop 착수 시점; 이후 develop의 t521·t522 병합은 본 카드 편집 파일과 무관)
measured by: lane-8 orchestrator, directly

## Claim

1. **12개 디렉터리 71건 전수 분류 완료** — 지시 13건 중립화, 산문·하네스 메타데이터 58건 보존.
   분류표와 근거는 `.moai/reports/t523/classification.md`.
2. **중립화 어휘는 착지된 규약을 따랐다** — `templates/AGENTS.md` 결속표의 question-channel 행과
   `agents-codex.yaml` tool_classes 어휘(REQ-CBN-005, 새 어휘 0건). 문체는 상위 SPEC이 11본
   에이전트 본문에 착지시킨 "through the harness's `<class>` capability" 형식.
3. **§B.6 선행 조건(표본 확인)을 수행했고, 전제가 부분 반증됐다** — moai-foundation-cc는
   「전부 산문」 성립 / moai 디렉터리(413건)는 워크플로 파일에 실제 행위 지시 다수 확인.
   이 카드 범위(12개/71건)는 운영자 확정대로 수행했고, moai 디렉터리 전수는 후속 판정 사항으로
   리드·운영자에게 넘긴다.
4. **행동 보존** — 제목·예시 서술·오케스트레이터 주어 문장·kanban-foreman의 `disallowed-tools:`
   프론트매터(제거 시 Claude 동작 변경)는 전부 원문 보존. 산문 문면 불변(REQ-CBN-010 논리).

## Evidence

- 분포 재측정(이 트리): `moai` 413 · foundation-cc 120 · foundation-core 105 · 나머지 12개 합계 71.
  t497 기준선(719/639/71) 대비 moai 1건 감소 외 일치.
- 표본 확인: cc 5줄 전부 산문 / core 3줄 산문(지시 인접 교육 콘텐츠) / moai 15줄 중 실제 지시
  다수(loop.md:194, mode-detection.md:64, phase-execution.md:361, harness.md:190, codemaps.md:234 등).
- 편집 후 잔여 AskUserQuestion(12개 디렉터리): 17건 — 전부 분류상 보존 대상(제목 3, 예시 서술 5,
  meta-harness Builder 묘사 3+주석, workflow-spec 오케스트레이터 주어 2, kanban-foreman 메타데이터·설명 2,
  workflow-spec SKILL.md 2). 지시형은 0건.
- `make build`: catalog.yaml 재생성(스킬 해시 갱신, 같은 커밋에 포함) + 빌드 성공.
- `go test ./internal/template/... -count=1`: ok 33.6s (agentemit 0.8s, commandemit 1.0s 포함 전부 ok).

## Baseline-attribution

모든 수치와 출력은 이번 런, 워크트리 `.claude/worktrees/t523`(WT-mirror-skill-neutrality @
7e1859c28 + 본 카드 커밋)에서 수집. 편집 파일은 4개 스킬 파일이며 t521(의사 Go)·t522(판정서)
병합과 겹치지 않는다.

## Gaps

- 상위 3본 638건은 표본 23줄만 분류했다. moai 디렉터리의 지시 전수(추정 수십 건)는 본 카드
  범위 밖 — 후속 카드 후보로 보고.
- codex 런타임이 개정된 문면을 실제로 읽었을 때의 거동 변화는 프로브하지 않았다(상위 SPEC도
  미측정 영역으로 기록).
- `moai/SKILL.md:8`류 프론트매터 `allowed-tools:`의 미러 노출 적합성은 본 카드의 분류 축 밖.

## Residual-risk

- 개정 문면("the harness's `question-channel` capability")은 결속표가 로드된 하네스에서만
  완전하게 해석된다. AGENTS.md 없이 스킬 단독으로 읽는 독자에게는 능력 이름만 남는다 —
  결속표가 모든 MoAI 프로젝트에 함께 배포되므로 이 저장소의 배포 형태에서는 닫히는 위험.
- moai 디렉터리 지시 중립화가 후속 카드로 진행되면 본 카드의 문체를 그대로 따르게 되며,
  그때 워크플로 파일의 지시 밀도(샘플상 높음)만큼 편집량이 커진다.
