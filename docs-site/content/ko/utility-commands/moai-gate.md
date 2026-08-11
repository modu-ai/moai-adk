---
title: /moai gate
weight: 70
draft: false
---

커밋 전 품질을 빠르게 검증하는 **경량 게이트** 명령어입니다. 린트·포맷·타입 검사·테스트를 **병렬로** 실행하여 대부분의 프로젝트에서 30초 이내에 완료됩니다. 검사 결과는 공유 진단 스냅샷에 남기 때문에, 뒤따르는 run-phase 사전 게이트나 sync 게이트가 같은 검사를 다시 돌리지 않고 기록된 결과를 재사용합니다.

{{< callout type="info" >}}
**한 줄 요약**: `/moai gate`는 "커밋 전 빠른 검문소" 입니다. 4가지 검사 (린트·포맷·타입·테스트) 를 동시에 돌려 통과/실패를 바로 알려 줍니다. 전체 코드 리뷰나 커버리지 분석은 건너뛰고 판정만 빠르게 냅니다.
{{< /callout >}}

{{< callout type="info" >}}
**슬래시 커맨드**: Claude Code에서 `/moai:gate`를 입력하면 이 명령어를 바로 실행할 수 있습니다. `/moai`만 입력하면 사용 가능한 모든 서브커맨드 목록이 표시됩니다.
{{< /callout >}}

## 개요

커밋하기 전에 "지금 상태가 깨끗한가?"를 확인하고 싶을 때 사용합니다. `/moai review` (심층 4관점 리뷰) 나 sync 파이프라인의 전체 품질 검사와 달리, `/moai gate`는 **빠른 통과/실패 판정**만 제공합니다.

| 워크플로우 | 범위 | 속도 | 사용 시점 |
|-----------|------|------|-----------|
| `/moai gate` | 린트 + 포맷 + 타입 + 테스트 | 빠름 (<30초) | 매 커밋 전 |
| `/moai review` | 4관점 심층 코드 리뷰 | 중간 (2-5분) | PR 전, 설계 리뷰 |
| sync 품질 검사 | 전체 품질 + 코드 리뷰 + 커버리지 | 느림 (5-10분) | sync 파이프라인 일부 |

## 사용법

```bash
# 전체 검사
> /moai gate

# 린트·포맷 자동 수정
> /moai gate --fix

# 스테이징된 파일만 검사
> /moai gate --staged

# 특정 파일만 검사
> /moai gate --file src/auth/service.go
```

## 지원 플래그

| 플래그 | 설명 | 예시 |
|-------|------|------|
| `--fix` | 린트·포맷 이슈 자동 수정 (기본값: 보고만) | `/moai gate --fix` |
| `--staged` | `git diff --staged` 파일만 검사 (테스트는 항상 전체 실행) | `/moai gate --staged` |
| `--file PATH` | 특정 파일만 검사 | `/moai gate --file src/api.go` |
| `--fresh` | fresh 모드 강제 — 공유 진단 스냅샷 재사용을 모두 끄고 모든 검사를 새로 실행 | `/moai gate --fresh` |

## 실행 과정

```mermaid
flowchart TD
    Start["/moai gate 실행"] --> Detect["1단계: 언어 감지<br/>(마커 파일 우선순위)"]
    Detect --> Snap["1B단계: 공유 스냅샷 소비<br/>(신선하면 재사용, --fresh면 건너뜀)"]
    Snap --> Parallel

    subgraph Parallel["2단계: 병렬 검사"]
        C1["Lint<br/>스타일·미사용 import"]
        C2["Format<br/>포맷 검증"]
        C3["Type<br/>정적 타입 분석"]
        C4["Test<br/>테스트 스위트"]
    end

    Parallel --> Report["3단계: 결과 보고<br/>(PASS/FAIL/WARN 표)"]
    Report --> Next["4단계: 다음 단계<br/>(실패 시 --fix / /moai fix / 무시)"]
```

### 1단계: 언어 감지

마커 파일을 우선순위대로 살펴 (먼저 걸리는 마커를 따릅니다) 언어별 도구 체인을 고릅니다. 16개 지원 언어를 똑같이 다루며, 예를 들어 Go라면 `go vet`·`golangci-lint`·`go test -race`가, Python이라면 `ruff`·`mypy`·`pytest`가 돌아갑니다. 알아볼 수 있는 마커가 없으면 언어별 검사를 건너뛰고 "unknown language"로 보고합니다.

### 1B단계: 공유 진단 스냅샷 소비

검사를 돌리기 전에, 지금 작업 트리에 대한 공유 진단 스냅샷이 있는지 먼저 봅니다. 아직 유효한 스냅샷(키 일치 + TTL 이내, 기본 10분)이 이 게이트가 돌릴 검사 항목을 모두 담고 있으면, 다시 돌리는 대신 기록된 결과를 재사용하고 보고서에 `Test | PASS (snapshot)`처럼 표시합니다. 오래된 스냅샷은 증거로 인용하지 않고 다시 돌립니다. `--fresh` 모드에서는 이 단계를 통째로 건너뜁니다.

### 2단계: 병렬 검사

4가지 검사를 백그라운드로 동시에 실행합니다.

| 검사 | 대상 | `--fix` 동작 |
|------|------|--------------|
| **Lint** | 스타일 위반, 미사용 import, 데드 코드 | 자동 수정 가능 항목 교정 |
| **Format** | 포맷되지 않은 파일 | 자동 포맷 |
| **Type** | 타입 오류, 누락된 어노테이션 | 자동 수정 없음 (수동 개입) |
| **Test** | 테스트 실패 | 자동 수정 없음 (원인 조사 필요) |

개별 검사 타임아웃은 60초, 전체 게이트 타임아웃은 90초입니다. 시간을 넘기면 WARNING으로 보고하되 막지는 않습니다. 재사용이 아니라 실제로 돌린 검사 결과는 `.moai/state/verify/` 아래 공유 스냅샷 스토어에 남습니다. 작업 트리가 그대로인 한, 뒤따르는 쪽(run-phase 사전 리뷰 게이트, sync 사전 게이트, stop-goal 평가기)이 이 결과를 그대로 가져다 씁니다.

### 3단계: 결과 보고

```
## Quality Gate: PASS
| Check  | Status | Time  |
|--------|--------|-------|
| Lint   | PASS   | 2.1s  |
| Format | PASS   | 0.8s  |
| Type   | PASS   | 3.2s  |
| Test   | PASS   | 12.4s |
Total: 18.5s
```

### 4단계: 다음 단계

모두 통과하면 커밋해도 좋다는 메시지를 띄웁니다. `--fix` 없이 실패했다면 `AskUserQuestion`으로 세 가지를 제시합니다. 자동 수정 (`--fix` 재실행, 권장), `/moai fix` (더 깊이 해결), 무시하고 진행입니다. `--fix`를 돌린 뒤에도 남은 이슈 (타입 오류·테스트 실패) 는 직접 들여다보는 편이 낫습니다.

## 다른 명령어와의 관계

`/moai gate`는 확인만 하고 파일에는 손대지 않는 **가벼운 검문소**입니다 (`--fix`를 줄 때만 린트·포맷을 바로잡습니다). 더 깊이 손봐야 하면 `/moai fix` (단발) 나 `/moai loop` (반복) 로 넘어가고, PR 전 종합 리뷰는 `/moai review`로 합니다. `--fresh` 모드는 `/moai loop`의 독립 최종 검증 패스가 자기 자신을 근거로 삼지 않는 증거를 얻으려고 이 게이트를 부를 때 씁니다.

## 관련 문서

- [/moai fix - 일회성 자동 수정](/ko/utility-commands/moai-fix)
- [/moai loop - 반복 수정 루프](/ko/utility-commands/moai-loop)
- [TRUST 5 품질 시스템](/ko/core-concepts/trust-5)
