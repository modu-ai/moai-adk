# t651 AS-2 hybrid registry 구현 인계

## Claim

초기 완전 정의 도구의 native 목록을 고정하고, ToolSearch 뒤 인증된 Claude snapshot에서 전체 정의가 확인된 후발 도구만 dispatcher 경로로 승인하는 `internal/codextools` 패키지를 구현했다. 패키지는 도구 실행이나 App Server RPC를 수행하지 않는다. 이 결과는 hybrid 경로의 등록·인수 검증·호출 승인 계약에 관한 로컬 증거이며, Claude Code 화면에서의 실제 왕복 완료는 아니다.

## Evidence

마지막 실행 명령과 원출력:

```text
go test -race ./internal/codextools -count=1 -timeout=60s -coverprofile=.moai/reports/t651/coverage.out
ok  github.com/modu-ai/moai-adk/internal/codextools 1.666s coverage: 87.5% of statements
```

`go vet ./internal/codextools`: exit 0, 출력 없음.

`GOOS=windows GOARCH=amd64 go test ./internal/codextools -c -o /tmp/moai-t651-codextools.test.exe`: exit 0, 출력 없음. Windows 런타임 PASS를 의미하지 않는다.

수정 전 RED는 red.log, snapshot-red.log, protocol-red.log, protocol-behavior-red.log, revocation-red.log, binding-red.log에 보존했다. 독립 검토에서 지적한 실제 행동 실패를 포함한다.

```text
--- FAIL: TestRevocationBeforeInvocation
    revoked tool returned executable invocation
--- FAIL: TestNativeToolWireDiscriminator
    installed DynamicToolSpec requires function discriminator
--- FAIL: TestSnapshotRevocationAndDuplicateDiscovery
    unchanged repeated discovery rejected invalid or unauthorized hybrid tool operation
```

최종 테스트는 해당 실패를 해소했다. 제거된 도구는 실행 가능한 Invocation 반환 전에 거절하며, wire schema에는 `type:function`을 포함한다. 동일한 참조의 후속 재검색은 같은 정의일 때 허용하지만, 같은 발견 batch 안의 중복 참조는 거절한다.

## Baseline-attribution

워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t651`, 브랜치 `WT-gateway-appserver-tools`, HEAD `dad5e9a9e`. 시작 전 fetch 및 origin/main...HEAD 결과 `0 3303`. AS2 계획 감사는 이전 트리의 appserver-redesign/hybrid-plan-audit.md에서 PASS 1.00, blocker 0을 읽고 구현했다.

이번 담당자가 작성한 코드 파일:

- internal/codextools/json.go
- internal/codextools/registry.go
- internal/codextools/registry_test.go

기존 간접 의존성 `github.com/santhosh-tekuri/jsonschema/v6 v6.0.2`를 재사용했다. go.mod/go.sum, CLI, AS-1 패키지는 변경하지 않았다. 오케스트레이터가 작성한 integration-source-snapshot.json은 별도 소유 증거이며 이 담당자가 수정하지 않았다.

## API와 동작 계약

- `New(Binding, initialDefinitions)`는 account/conversation binding을 요구한다. ThreadID는 아직 server가 배정하지 않았을 수 있다.
- `NativeTools()`를 thread/start의 dynamicTools에 사용하고, 응답의 실제 thread ID로 `BindThread(id)`를 한 번 호출한다. binding 전 Discover/Begin/Complete는 거절한다. 이미 승인된 thread에 대한 New도 가능하다.
- `Discover(binding, typedReferences, authenticatedCurrentSnapshot)`는 전체 정의와 참조를 결합한다. 모델 텍스트나 이름만으로 도구를 등록하지 않는다. 같은 정의를 다시 발견하는 것은 idempotent이며 등록 epoch를 늘리지 않는다. 다른 정의로 덮어쓰기는 거절한다.
- `Begin(binding, call, authenticatedCurrentSnapshot)`는 현재 snapshot에도 같은 도구 정의가 허용되어 있는지 확인한 뒤 전체 schema로 인수를 검증한다. 성공하면 원래 Claude 도구 이름, 정규화한 인수, schema digest, 발견 epoch를 반환하고 call ID를 예약한다.
- `Complete(binding, turnID, callID, authenticatedCurrentSnapshot)`는 같은 소유 대화/thread/turn/call과 고정 정의를 확인하여 결과를 한 번 소비한다. 실제 RPC 응답 송신과 pending 영속화는 후속 AS-3 책임이다.

초기 도구는 native 경로로만 승인하고 후발 도구는 dispatcher로만 승인한다. dispatcher 재귀와 예약 이름/alias 충돌을 거절한다. `DeferredToolPlaceholder`는 실제 포착한 빈 object schema sentinel일 때만 제외하며, 일반 `defer_loading=true` 도구를 일괄 제외하지 않는다. 초기 완전 정의는 모델에 native 함수로 공개한다.

JSON Schema는 기본 Draft2020-12, 명시 dialect는 Draft7(http URI+#), 2019-09, 2020-12를 지원한다. 기존 validator를 이용하여 required·타입·enum·중첩·additionalProperties 등을 검사한다. 로컬 문서 $ref는 해석하며 외부 URL/file loader는 항상 거절한다. format 등 annotation의 의미는 선택한 dialect와 라이브러리 기본 정책을 따른다.

JSON 입력은 중복 key, 깊이 64 초과, schema 256 KiB 초과, 인수 1 MiB 초과 등을 거절한다. object key 정렬은 canonical encoding에 사용하지만 배열 순서와 숫자 표기는 유지한다. 도구 수 1024, snapshot 합계 8 MiB, call 이력 65536의 상한이 있다. 상한을 넘으면 새 승인을 거절하고 자동 thread 생성/재생은 하지 않는다.

## 여섯 음성군의 로컬 판정

| 그룹 | 검증한 로컬 행동 |
|---|---|
| 1 이름 충돌 | 중복 정의·dispatcher·native alias 예약 이름 등록 거절 |
| 2 참조만 존재 | 전체 schema 없는 참조 등록과 미등록 도구 Invocation 반환 거절 |
| 3 schema 변경 | 기존 등록 덮어쓰기 및 pending 결과에 다른 정의 적용 거절 |
| 4 인수·schema 검증 | required·타입·enum·중첩·additionalProperties·미지원 dialect·외부 $ref·중복 key 위반 거절 |
| 5 소유권·중복 | 다른 conversation binding·다른 turn 결과·같은 call ID·중복 완료 거절 |
| 6 경로·수명 | native의 dispatcher 우회·dispatcher 재귀 거절; 발견 전후 고정 native 목록 동일; 서로 다른 정상 call 두 개 승인 |

거절 사례는 Invocation 반환 또는 결과 소비가 0임을 판정한다. 실제 Claude 도구 실행 횟수가 0이라는 제품 E2E 증거로 바꾸어 표현하지 않는다. 패키지에 thread/start/fork/resume 호출 기능이 없으며 발견은 네트워크를 발생시키지 않는다.

## Gaps

실제 Claude ToolSearch→발견 도구→tool_result→최종 답변, App Server thread/turn 연결, HTTP 경계 유지, 이미지/실패 도구 결과, pending crash 복구는 실행하지 않았다. 새 Registry가 받는 snapshot과 Binding의 인증 출처는 후속 HTTP adapter가 보장해야 한다. 전체 SPEC 완료나 카드 운영 완료 판정은 발급하지 않는다.

Windows runtime은 GitHub CI 대기다. LSP를 실행하지 않았다. 87.5%는 현재 플랫폼 패키지 statement coverage이며 모든 파일 개별 85%를 주장하지 않는다. repository-wide verdict는 integration branch CI 소관이며 PENDING이다. 커밋·설치·push·PR를 하지 않았다.

## Residual-risk

schema compile·검증은 라이브러리에 위임하며 context 취소 API를 제공하지 않는다. 입력 크기·깊이 제한은 있으나 복잡한 합성 schema의 CPU 상한까지 실증하지 않았다. 현재 digest는 전체 typed definition을 canonicalize하므로 설명·defer 플래그 변경도 보수적으로 다른 정의로 본다. 동일 사용자 intent라도 정의가 바뀌면 명시적으로 거절할 수 있으며, 호환성 완화를 자동으로 넣지 않았다. 후속 adapter가 인증되지 않은 model payload를 snapshot으로 넘기면 이 패키지만으로 진위를 판별할 수 없다.
