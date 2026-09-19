---
id: SPEC-INIT-DEPLOY-EXIT-001
title: "moai init 템플릿 배포 실패를 치명적 오류로 승격 — exit 0 오보고 제거"
version: "0.1.0"
status: completed
created: 2026-09-18
updated: 2026-09-18
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/core/project, internal/cli"
lifecycle: spec-anchored
tags: "init, exit-code, template-deploy, cli, card-t931"
tier: M
---

# SPEC-INIT-DEPLOY-EXIT-001 — moai init 템플릿 배포 실패의 종료 코드 승격

- 카드: **t931**
- Tier: M · Class: C · cycle_type: tdd

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-09-18 | 0.1.0 | 최초 작성 (카드 t931) | manager-spec |

---

## §A 배경 — 관측된 문제

`moai init`은 템플릿 배포가 실패해도 종료 코드 0으로 끝난다. 사용자는 초록색 "MoAI project initialized" 카드를 읽지만, 계약 파일은 조용히 빠져 있다.

### A.1 실측 대조표

측정 조건: 이 워크트리 HEAD `3ff2e5dda`에서 `make build`로 빌드한 바이너리를 `./bin/moai init <dir> --llm claude --non-interactive`로 실행.

| 조건 | 종료 코드 | 배포 파일 수 | AGENTS.md |
|---|---|---|---|
| 템플릿 무수정 (양성 대조) | 0 | 596 | 존재 |
| `AGENTS.md.tmpl`에 `$REPO` 토큰 주입 | 0 | 595 | **부재** |
| `.claude/hooks/moai/handle-agent-hook.sh.tmpl`에 `$REPO` 토큰 주입 | 0 | **77** | **부재** |

세 번째 행이 심각도를 결정한다. 배포 순회 초반에서 렌더 오류가 나면 **519개 파일이 통째로 누락**되는데도 종료 코드는 0이다.

두 실패 실행 모두 stderr에는 경고가 나왔다.

```
1 warning(s) during init:
  1. template deployment: deploy templates: template render "<path>": template: unexpanded dynamic token detected: found "$REPO"
```

즉 정보 자체는 있었으나, **종료 코드와 성공 카드가 그 정보와 반대되는 신호를 냈다.**

### A.2 확인된 원인 지점

| 위치 | 하는 일 |
|---|---|
| `internal/core/project/initializer.go:208-212` | `deployTemplates` 오류를 `result.Warnings`에 추가하고 `// Template deployment is non-fatal; record warning` 주석을 단 뒤 `Init`이 nil을 반환 |
| `internal/cli/init.go:905` | `result.Warnings`를 경고 수집기에 넣은 다음 `buildInitSuccessCard(...)`를 **무조건** 출력. `init.go` 어디에도 경고를 근거로 한 종료 코드 분기가 없음 |
| `internal/template/deployer.go:160-291` | 임베드 FS 순회 중 렌더 오류가 WalkDir 콜백에서 반환되며 남은 순회를 중단 — 519개 파일 손실의 발생 지점 |

### A.3 두 가지 관문 질문에 대한 확인된 답

**(a) 경고 처리가 설계 결정인가 — 아니다.** 기록된 근거가 없다. 이 동작을 주장하는 것은 코드 주석 하나이고, `internal/core/project/initializer_test.go:597-633`의 `TestInit_WithDeployerError`가 그것을 고정하고 있을 뿐이다. `.moai/specs/SPEC-INIT-001/`과 `.moai/specs/SPEC-CLI-TUX-V3-002/`를 조회했으나 이 동작을 요구하는 SPEC 요구사항은 없다.

**(b) 실패해도 되는 템플릿과 중단해야 하는 템플릿을 가르는 기준이 있는가 — 없다.** `internal/template/deployer.go`에 `required` / `critical` / `mandatory`는 **0회** 등장한다.

따라서 카드 자신의 결정 규칙이 적용된다 — 종료 코드를 0이 아닌 값으로 바꾼다.

---

## §B 요구사항 (GEARS)

### REQ-IDE-001 — 배포 실패의 치명화

**When** `moai init`이 템플릿 배포 단계에서 오류를 만나면, `moai init`은 0이 아닌 종료 코드로 종료해야 한다(shall).

### REQ-IDE-002 — 성공 카드 억제

**While** 템플릿 배포가 실패한 상태에서, `moai init`은 초기화 성공 카드("MoAI project initialized")를 출력해서는 안 된다(shall not).

### REQ-IDE-003 — 실패 표면

**When** 템플릿 배포 실패로 init이 중단되면, `moai init`은 실패한 템플릿을 지목하고 프로젝트 트리가 불완전하다는 사실을 밝히는 오류 표면을 stderr에 출력해야 한다(shall).

### REQ-IDE-004 — 경고 요약과 채널 규율 보존

`moai init`은 기존 경고 요약 stderr 출력 형식을 유지해야 하며(shall), stdout에는 어떤 경고·오류 텍스트도 실어서는 안 된다(shall not).

### REQ-IDE-005 — 다른 일곱 경고 지점의 불변

**Where** 오류가 템플릿 배포 이외의 저하 지점에서 발생한 경우, `moai init`은 종전대로 그것을 경고로 기록하고 종료 코드 0을 유지해야 한다(shall).

### REQ-IDE-006 — 경로 간 일관성

**Where** init이 `--force` 재초기화 경로 또는 codex 전용(`--llm gpt`) 경로로 실행된 경우에도, 템플릿 배포 실패에 대한 종료 코드와 카드 억제 동작은 기본 경로와 동일해야 한다(shall).

### REQ-IDE-007 — 기존 계약 테스트의 반전

`TestInit_WithDeployerError`는 `Init()`이 배포 실패 시 오류를 반환한다는 새 계약을 검증하도록 갱신되어야 한다(shall).

---

## §C 범위 밖 (Out of Scope)

### Out of Scope — 나머지 일곱 개 경고 저하 지점

`initializer.go`는 여덟 곳(줄 210, 216, 236, 255, 272, 293, 304, 396)에서 오류를 경고로 낮춘다. 이 SPEC은 **템플릿 배포 부류(줄 210) 하나만** 치명화한다.

- 줄 216 — config 생성 fallback: 경고 유지
- 줄 236 — report config 기록: 경고 유지
- 줄 255 — page-3 config 기록: 경고 유지
- 줄 272 — workflow 토글 기록: 경고 유지
- 줄 293 — manifest 초기화: 경고 유지
- 줄 304 — shell 환경 설정: 경고 유지 (`git` 바이너리 부재가 init을 실패시켜서는 안 된다)
- 줄 396 — skill-mirror 통지: 경고 유지

"경고가 하나라도 있으면 종료 코드를 0이 아닌 값으로" 같은 일괄 규칙은 **기각한다.** 이유는 셸 설정 지점이 직접 보여준다 — 사용자 머신에 특정 도구가 없다는 환경 사실이 프로젝트 초기화 자체의 실패로 보고되면, 종료 코드는 다시 신뢰할 수 없는 신호가 된다. 지금 문제는 "경고를 무시한다"가 아니라 "계약 파일이 없는데 있다고 말한다"이므로, 치명화 대상은 산출물 완전성을 무너뜨리는 부류로 한정한다.

### Out of Scope — 배포 순회의 부분 실패 복구

`deployer.go`가 렌더 오류 시 남은 순회를 중단해 519개 파일을 잃는 동작 자체는 이 SPEC이 고치지 않는다. 이 SPEC은 그 결과를 **정직하게 보고**하게 만들 뿐이다. 순회 재개나 필수/선택 템플릿 분류 기준 도입은 별도 작업이다.

### Out of Scope — 롤백과 정리

실패 시 이미 배포된 부분 트리를 되돌리는 동작은 범위 밖이다. 부분 트리는 남고, 사용자는 오류 표면을 통해 그 사실을 통보받는다.

### Out of Scope — e2e 시나리오 확장

`e2e/cli/tux3_journeys.sh`의 J1·J1b는 정상 init만 돌리므로 계속 종료 코드 0을 기대하며, 이 SPEC은 그 스크립트를 수정하지 않는다.

---

## §D 소비자 영향 (측정치)

`moai init`의 종료 코드를 소비하는 자동화는 `e2e/cli/tux3_journeys.sh` 하나뿐이다 — J1(약 105행)이 `RC -eq 0`을, J1b도 `RC -eq 0`을 단언한다. 둘 다 정상 init을 돌리므로 변경 후에도 종료 코드 0을 유지한다. 저장소의 나머지 약 30개 `moai init` 언급은 전부 README / docs-site 산문이다.

> 이 수치는 카드 배차 시점에 측정된 값이며 run 단계에서 재측정하지 않는다.

---

## §E 인수 기준

`acceptance.md` 참조.
