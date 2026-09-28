# LSEL 드레인 운영 (지역 자가진화 루프)

> 복원 2026-08-26 (SPEC-LSEL-DRAIN-STALL-001 M2). 이 섹션이 과거 진행 문서들이 참조하던 "§28" 앵커의 실체다 — 3주 드레인 정지(2026-08-04~25) 당시 이 섹션이 소실돼 있었다. 정본은 `.claude/skills/hns-lsel-curator/SKILL.md` § Durable operations.

### [HARD] 내구 트리거 — 세션 시작 배선

드레인은 어떤 Claude 세션의 생사에도 의존하지 않는다. `.claude/settings.local.json` `.hooks.SessionStart`에 2개 lsel 항목이 배선돼 있다(각 `"timeout": 30`): (1) `session_drain.sh` 래퍼 — 배타잠금 → 기존 clusters.json 무조건 보존(`clusters-history/`) → `drain.sh` 실행 → 1행 상태 → fail-open. (2) `backlog_check.sh` advisory — 읽지 않은 백로그가 임계(기본 25, `LSEL_BACKLOG_THRESHOLD`) 초과면 system-reminder 발화. **이 배선이 settings.local.json(로컬 전용)에 있는 것은 의도적이다** — tracked `settings.json`은 `moai update`가 통째 재배포해 배선이 매 업데이트 유실된다(§2.3). 검증: `jq '.hooks.SessionStart' .claude/settings.local.json`.

### [HARD] 모든 드레인은 래퍼 경로로

`drain.sh`를 **직접** 호출하지 않는다 — 직접 호출은 호출-전 보존을 우회해 스테이징된 후보를 조용히 유실시킨다(drain.sh는 드레인 경로든 no-op 경로든 clusters.json을 덮어쓴다). 항상 `session_drain.sh --inbox .moai/lessons-inbox.jsonl --state-dir .moai/state/lsel`.

### PROPOSE는 archived 사본 판독

세션 시작마다 드레인이 도는 체제에서 live `clusters.json`은 휘발성이다(no-op 드레인조차 `candidates: []`로 덮어쓴다). 후보 제안(PROPOSE)은 `.moai/state/lsel/clusters-history/`의 사본(최신순)을 읽는다. 검증 레시피·mutant guard 포함 전체 절차는 SKILL.md § Verification.

### 인박스 유용성 범위 (경계 선언 — anchor: `.moai/docs/learning-channel-scope.md`)

`.moai/lessons-inbox.jsonl`은 실패 이벤트 스텁만 기록한다 — 배선된 2패밀리 `tool_failure:<tool>:<sig>`와 `test_fail:<pkg>:`뿐이다. 도구 실패와 테스트 실패 어느 쪽으로도 나타나지 않는 결함 계열(공허 초록, 판정 전 skip, 빈 결과집합 통과, 형제 수리 누락, 이동 ref 고정, 스테일 값 인용)은 이 인박스에 담기지 않으며, 그 학습 채널은 인간 매개 루프다 — 레인 발견 → 리드 판정 → auto-memory `feedback_*.md` + `MEMORY.md` 기록. 측정 구성과 dated baseline은 anchor doc에만 두고 산문에 수치를 두지 않는다.

---
