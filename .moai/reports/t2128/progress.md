# t2128 — lane record (lane-28, run to2bbk)

카드: PR #1969 Codex P2 수리(config) — AF 설정 읽기의 secret 검사가 자기 참조 YAML alias(`webhook.routes` 안 `nested: *cycle`)에서 끝없이 재귀해 stack overflow로 죽음. Class B, SPEC 없음. 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2128`, 브랜치 `WT-pr-codex-repair-6`, 기준 HEAD `04d8c27a7e`. 착지는 t2121+rc.3 뒤 카드 PR 경로(리더 배차문).

## 1. 출발점 (2026-11-06)

- 배차: 리더 `uds:/tmp/cc-socks/82930.sock`, card t2128, evidence `.moai/reports/t2128/verdict.md`(audited_sha=HEAD) + `card-review.md`, Codex 단계당 최대 2회(규칙 `d-20261106T080039Z-305c`).
- 지적 원문 위치: `internal/config/agentic_factory_settings.go:434` @04d8c27a7e (PR #1969 리뷰 코멘트).
- 레인 읽기(이 트리, HEAD 04d8c27a7e): `afsHasSecretKey`(파일 429행 근처)는 `yaml.AliasNode`를 만나면 `n.Alias`로 재귀하고, mapping·sequence 자식으로도 재귀하며 방문 기록이 없음. alias가 자기를 포함하는 앵커를 가리키면 종료 조건이 없음 — 지적과 일치(가설, 재현 전).
- 재현은 manager-develop이 서브프로세스 시험으로 먼저 확인(카드 본문 요구).

## 2. 진행 기록

(이하 단계별로 추가)

### 2.1 재현 (RED, HEAD 04d8c27a7e, 수정 전 코드)

- 진입점: `LoadAgenticFactorySettings(projectRoot)` → `readAgenticFactorySettings` → `r.guardSecrets()` → `afsHasSecretKey`(agentic_factory_settings.go:429-447).
- 시험: `internal/config/agentic_factory_settings_cycle_test.go` `TestAgenticFactorySettings_RefCycle`(부모)가 자기 바이너리를 `-test.run=^TestAgenticFactorySettings_RefCycleChild$`로 재실행(환경변수 `MOAI_AFS_REF_CHILD_ROOT`)하고 60초 제한 컨텍스트로 묶음. 최소 YAML: `webhook.routes` 첫 항목이 `&cycle` 앵커 매핑이고 그 안에 `nested: *cycle`.
- 명령: `GOCACHE=<scratch>/gocache-lane28 go test -count=1 -run TestAgenticFactorySettings_RefCycle ./internal/config/` → exit=1. 자식 stderr 머리말(원문):

```
--- FAIL: TestAgenticFactorySettings_RefCycle (3.86s)
    agentic_factory_settings_cycle_test.go:78: loading a self-referencing alias crashed the process (exit status 2):
        === RUN   TestAgenticFactorySettings_RefCycleChild
        runtime: goroutine stack exceeds 1000000000-byte limit
        fatal error: stack overflow
```

- 자식 단독 실행(같은 YAML, 테스트 바이너리 직접 실행) 스택 프레임(원문 발췌): `afsHasSecretKey ... agentic_factory_settings.go:429` → `:434`(AliasNode→n.Alias) → `:439`(매핑 값 재귀) → `:434` → `:439` … 반복, exit=2.

### 2.2 다른 보행 함수 점검 (이 입력 기준)

- `afsLookup`: dotted 경로 세그먼트 수만큼만 순회(유한), alias는 세그먼트당 1회만 따라감 — 순환 입력에서 무한 루프 없음.
- `afsChild`: 비재귀.
- `routes()`: 항목이 `MappingNode`가 아니면 거부, 매핑이면 `repository` 한 키만 조회 — 순환과 무관. 단 `guardSecrets`가 먼저 실행되므로 크래시는 그 단계에서 발생.
- `yaml.Node.Decode`: 스칼라(`enabled`, 정수 값)에만 호출 — 순환 노드를 디코드하지 않음.
- 결론: 이 입력으로 닿는 재귀는 `afsHasSecretKey` 하나뿐. 수정 대상은 그것만.
