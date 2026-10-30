# SPEC-AGENTIC-FACTORY-LOCAL-001 — 조사 메모 (0.2.0)

짧은 안내 메모다. 깊은 코드 조사는 하지 않았다.

## 읽은 것

- 사용자 결정 기록(리더의 옮겨 적음): `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/user-decision-local-adk-20261030.md` 전문.
- 리더의 구현 계약: `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md` 전문(§1~§7).
- 기준 커밋 `37b67f6d39`에서 읽기 전용 git 명령으로 확인한 것:
  - `git log -1 --format='%h %s' 37b67f6d39` → "37b67f6d39 merge(lane-18): F6 transitions outside the table are refused (5b164ccd14, includes 338c7c33fb) (Issue #1948)"
  - `git ls-tree -d --name-only 37b67f6d39 internal/agenticfactory/ internal/agenticfactory/monitor/ internal/web/` — `ledger`·`mergeq`·`monitor/flow` 없음, 퇴역 대상 `forge`·`mirror`·`gitref`·`effect/ctxledger`·`effect/ghintake`·`effect/labels`·`identity`·`intake`·`resident` 있음.
  - `git grep -n 'CheckMove' 37b67f6d39 -- internal` → `internal/agenticfactory/core/move.go:28` 정의, `claim/claim.go:478` 호출.
  - `git ls-tree --name-only 37b67f6d39 internal/agenticfactory/claim/presence.go` → 있음.
  - 시험 이름 21개 부재(acceptance.md §B).
- (0.2.0) 같은 계약을 다시 읽었다: §1~§10 전문, sha256 `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb`(spec §A.4). 위 0.1.0 줄의 "전문(§1~§7)"은 그때 읽은 범위의 기록이다. 그때 파일에 §8~§10이 있었는지는 옛 판이 없어 확인할 수 없다.

## 읽지 않은 것

- 사용자 결정 원문(리더 세션 기록), 리더 보고 v2(artifact).
- 퇴역 대상 패키지의 본문과 그것을 가져오는 곳의 목록(import 그래프). 퇴역 순서의 실제 비용은 조사하지 않았다.
- `core.CheckMove`의 F6 강화 내용(계약 I6이 말하는 "explicit composite list, park exits single-edge, E20 needs caller evidence"). 이 작성자는 정의 위치만 보았다.
- `claim/presence.go`의 본문.
- 옛 원장(backlog.db, factory.db)과 claim JSON 문서의 실제 형태. importer 대응표는 쓰지 않았다.
- 시험은 하나도 돌리지 않았다.
