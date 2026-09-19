# SPEC-INIT-DEPLOY-EXIT-001 — 인수 기준

- 카드: **t931** · Tier M · cycle_type: tdd

모든 AC는 기계적으로 검증 가능하며, 각 항목은 이진 판정(PASS/FAIL)을 낸다.

---

## §D AC 매트릭스

| AC | 층 | 대응 요구사항 |
|---|---|---|
| AC-IDE-001 | 바이너리 대조 프로브 | REQ-IDE-001, REQ-IDE-002 |
| AC-IDE-002 | 바이너리 대조 프로브(양성 대조) | REQ-IDE-001 |
| AC-IDE-003 | initializer 층 (Go) | REQ-IDE-001, REQ-IDE-007 |
| AC-IDE-004 | initializer 층 (Go) | REQ-IDE-005 |
| AC-IDE-005 | CLI 층 (Go) | REQ-IDE-002, REQ-IDE-003 |
| AC-IDE-006 | CLI 층 (Go) | REQ-IDE-004 |
| AC-IDE-007 | CLI 층 (Go) | REQ-IDE-006 |
| AC-IDE-008 | 회귀 | 소비자 영향 |

---

### AC-IDE-001 — 배포 실패 시 종료 코드와 카드 부재 (바이너리 프로브)

**Given** 이 워크트리의 `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl`에 `$REPO` 토큰을 주입하고 `make build`로 바이너리를 다시 빌드한 상태에서,
**When** `./bin/moai init <빈 임시 디렉터리> --llm claude --non-interactive`를 실행하면,
**Then** 종료 코드가 0이 아니고, 결합된 stdout+stderr 어디에도 문자열 `MoAI project initialized`가 나타나지 않으며, stderr에 실패한 템플릿 경로가 포함된 오류 메시지가 나타난다.

측정 후 주입한 토큰을 원복하고 `git status --porcelain`이 해당 파일에 대해 빈 출력을 내는 것까지 확인해야 이 AC가 완결된다.

---

### AC-IDE-002 — 양성 대조: 정상 init은 변하지 않는다

**Given** 템플릿이 무수정인 상태에서 `make build`로 빌드한 바이너리로,
**When** `./bin/moai init <빈 임시 디렉터리> --llm claude --non-interactive`를 실행하면,
**Then** 종료 코드가 0이고, stderr에 `MoAI project initialized` 카드가 나타나며, 배포 파일 수가 기준선(596)과 일치한다.

이 대조군 없이 AC-IDE-001만 통과하는 것은 근거가 되지 않는다 — 프로브 자체가 항상 실패를 보고하는 고장 상태와 구별되지 않기 때문이다.

---

### AC-IDE-003 — `Init()`이 배포 실패에 오류를 반환한다

**Given** `deployErr`를 반환하도록 구성한 `mockDeployer`로 `Initializer`를 만든 상태에서,
**When** `init.Init(context.Background(), opts)`를 호출하면,
**Then** 반환된 error가 non-nil이고, 그 메시지가 템플릿 배포를 지목한다.

이는 `internal/core/project/initializer_test.go:597`의 `TestInit_WithDeployerError`를 **반전**시키는 것이다. 이 테스트는 현재 `Init()`이 nil을 반환한다고 단언하므로, 갱신은 부수적 정리가 아니라 이 SPEC이 요구하는 필수 변경이다(REQ-IDE-007).

---

### AC-IDE-004 — 나머지 저하 지점은 여전히 경고로 남는다

**Given** 템플릿 배포는 성공하지만 배포 이외 단계(예: manifest 초기화, 셸 설정) 중 하나가 오류를 내는 구성에서,
**When** `init.Init(...)`을 호출하면,
**Then** 반환된 error가 nil이고, `result.Warnings`에 해당 경고가 기록되어 있다.

---

### AC-IDE-005 — CLI가 실패 시 성공 카드를 인쇄하지 않는다

**Given** 템플릿 배포 실패를 일으키는 executor를 주입한 init 커맨드에서,
**When** 커맨드를 실행하면,
**Then** `RunE`가 non-nil error를 반환하고, 캡처된 stderr에 `MoAI project initialized`가 포함되지 않으며, 오류 텍스트가 프로젝트 트리의 불완전성을 밝힌다.

---

### AC-IDE-006 — 경고 요약 보존과 채널 규율

**Given** 템플릿 배포 실패로 init이 중단되는 실행에서,
**When** stdout과 stderr를 분리 캡처하면,
**Then** stderr에 기존 형식의 `N warning(s) during init:` 요약이 여전히 나타나고, stdout에는 경고·오류 텍스트가 전혀 포함되지 않는다.

---

### AC-IDE-007 — `--force` 및 codex 전용 경로의 일관성

**Given** 템플릿 배포 실패를 일으키는 구성에서,
**When** `--force`(이미 초기화된 디렉터리 재초기화)로 한 번, `--llm gpt`(codex 전용)로 한 번 각각 실행하면,
**Then** 두 실행 모두 0이 아닌 종료 코드를 내고 성공 카드를 인쇄하지 않는다.

---

### AC-IDE-008 — 소비자 회귀 부재

**Given** 변경된 바이너리로,
**When** `e2e/cli/tux3_journeys.sh`의 J1 및 J1b 시나리오를 실행하면,
**Then** 두 시나리오 모두 종료 코드 0을 단언하는 기존 검사를 통과한다.

---

## §D.1 경계 사례

- **배포기가 nil인 fallback 경로**: `i.deployer == nil`일 때는 `generateConfigsFallback`가 돌고 이는 경고 경로다. 이 SPEC은 그 분기를 바꾸지 않으므로, 해당 경로가 여전히 경고 + 종료 코드 0임을 확인한다.
- **skill-mirror 통지 동반**: `deployTemplates`는 오류가 나도 mirror 통지 라인을 `result.Warnings`에 먼저 넣는다(`initializer.go` deployTemplates 본문). 오류 반환으로 바뀐 뒤에도 그 통지가 유실되지 않고 경고 요약에 나타나는지 확인한다.
- **부분 트리 잔존**: 실패 후 디렉터리에 부분 배포물이 남는다. 이는 의도된 범위 밖 동작이며(`spec.md` §C), 테스트는 정리를 요구하지 않는다.

---

## §D.2 품질 게이트

- `go vet ./internal/core/project/... ./internal/cli/...` 무결
- `golangci-lint run` 신규 지적 0
- 변경 패키지 테스트 통과 (`go test -count=1 ./internal/core/project/... ./internal/cli/...`) — 전체 스위트 판정은 CI 몫
- `make build` 통과

---

## §D.3 완료 정의 (Definition of Done)

1. AC-IDE-001 ~ AC-IDE-008 전부 PASS, 각 항목에 실행한 명령과 그 출력이 증거로 첨부됨
2. `TestInit_WithDeployerError`가 새 계약으로 갱신되어 있고, 갱신 전 RED 관측이 기록됨
3. 측정용 템플릿 토큰 주입이 전부 원복되고 `git status --porcelain`로 확인됨
4. `spec.md` §C의 일곱 경고 지점이 변경되지 않았음이 diff로 확인됨
