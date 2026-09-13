# t649 구현 인계 기록

## Claim

현재 작업은 **부분 구현과 로컬 검증 완료, 실사용 인수 전** 상태다. 카드 완료·배포·PR·통합을 주장하지 않는다.

- 공통 launcher의 시작 모델을 제공자에 결합하고 다른 제공자의 명시 모델을 거절한다. 세션 catalog·modelPicker·availableModels·네 슬롯·fallback을 해당 제공자에 한정한다.
- cc/glm 생산 진입은 `unifiedLaunchDefault → newNativeGatewayBinding → newGatewaySessionBinding`으로 연결했다. GPT 기존 생산 binding을 공통 생성기로 재사용하며 Claude·GLM 대화 namespace는 별도로 둔다. Claude OAuth Bearer와 MoAI 세션 헤더를 분리한다.
- GPT reasoning의 최종 item 원문을 기존 opaque envelope에 보존한다. 응답에서 digest에 묶인 tool ID를 만들고 후속 요청에서 원래 call ID와 reasoning 순서를 복원한다. native 경로는 출력 상한 내에서 완료 응답을 모은 뒤 Messages SSE의 start/delta/stop으로 전달한다.
- receipt는 비공개 store·SessionID·FamilyID·안정된 계정 digest·검증된 모델 계열에 결합된다. refresh는 같은 계정의 scope를 보존한다. 계정 교체, 다른 모델 계열, carrier 제거, 무단 prefix, 게시 실패를 거절한다. 명시적 fork는 승인된 receipt 사본과 FamilyID를 유지한다.
- native 대화 재개는 이미 등록된 UUID의 비공개 namespace 안에서만 정확한 JSONL 파일을 찾는다. 관측된 `sessionId`/`cwd` 형태를 지원하며 오류·synthetic·미완료 assistant는 완료로 등록하지 않는다. 마지막 성공 모델을 재개 시 전달하고 `--continue`와 정확한 `--resume`의 중복 전달을 제거했다.
- 기본 행에 임의 `[1m]` ID가 붙지 않도록 `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`을 전달한다. GPT는 `ENABLE_TOOL_SEARCH=false`로 전체 MCP schema 노출 경로를 선택한다.

## Evidence

RED는 implementation-evidence/의 `*-red.txt`에 실행 당시 출력으로 보존했다. 컴파일 이전의 새 API 시험도 포함되며, 새 런타임 동작의 실패 출력과 구별한다. 대표 관측:

```text
TestGatewayProviderContractRejectsForeignModel: claude/gpt/glm accepted foreign model
TestGatewayProviderContractPickerAndSlots: missing private modelPicker; OPUS/SONNET/HAIKU/FABLE slot not bound
TestReasoningCarrierPreservesOriginalItemBytes: opaque item bytes were reserialized
TestReasoningWithoutPublicCompletionFailsClosed: reasoning-only output reported success
TestNativeCompletedTranscriptIsDiscoveredOnExactResume: native transcript is incomplete
TestGatewayExactResumeDoesNotAlsoPassContinue: resume session failed: ambiguous continue invoked
```

직접 Claude를 실행한 네트워크 미사용 UI probe는 `http://127.0.0.1:9`, 가짜 세션 인증, 격리 config, finally 프로세스 그룹 정리를 사용했다. 이는 MoAI 경유 UI 실험이 아니다.

```text
python3 /tmp/t649-picker-no1m.py gpt
1.Default(recommended)Usethedefaultmodel(currentlyGPT-6AstraPROBE)
⎿SetmodeltoGPT-6AstraPROBEforthissessiononly
PROBE_PROCESS_CLEANED True
```

공식 문서의 환경 변수 계약도 확인했다. `ENABLE_TOOL_SEARCH=false`는 전체 MCP 도구를 미리 로드하며, `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`은 1M 선택 항목을 비활성화한다.

- https://code.claude.com/docs/en/mcp
- https://code.claude.com/docs/en/env-vars

로컬 시험과 정적 검사 원문 출력은 `implementation-evidence/`의 `launch-green.txt`, `gateway-final.txt`, `race.txt`, `native-race-final.txt`, `vet-final.txt`을 참조한다. 마지막 파일들은 최종 검증 종료 후 복사한다.

## Baseline-attribution

```text
git rev-parse --short HEAD
81c1d58f9
git branch --show-current
WT-unified-gateway
```

기준은 기존 미커밋 변경이 다수 있던 moai-proxy-unified 작업 트리에 이 담당자의 변경을 더한 상태다. 기존 변경을 초기화하지 않았으며 커밋·push·PR·병합을 수행하지 않았다. 기존 profile/notice 회귀 시험은 공유 `unifiedLaunchWithGateway(..., nil)` 부분을 검증하도록 연결점을 명시했다. 실제 생산 진입은 별도 default/provider 시험과 앞으로 수행할 실사용 시험의 대상이다.

## Gaps

1. 실제 MoAI launcher에서 세 picker·GPT-6 답변·Read 도구 후속·새 프로세스 resume를 실행하지 않았다. 유료 live 요청은 root 담당으로 지정되어 있었다.
2. managed 설정의 모든 출처(server/MDM/registry/file)와 충돌을 생산 진입 전에 판정하는 기능은 아직 없다. 서버 catalog 제한은 적용되지만 이 항목을 완료로 볼 수 없다.
3. GPT 전체 MCP schema 노출은 공식 계약과 env 시험만 확인했다. 실제 MCP 도구가 완전히 노출되는지는 별도 실행이 필요하다.
4. native 성공 JSONL 시험은 실제 실패 대화에서 확인한 필드 구조와 Messages 완료 필드를 이용한 fixture다. 실제 성공 transcript·모델 선택 기록·resume/fork 조합을 대조해야 한다. native child가 저장하는 public message 분할·phase와 opaque carrier의 보존도 실사용에서 확인해야 한다.
5. 독립 의미 감사와 Windows GitHub CI 실행을 수행하지 않았다. 통합 브랜치의 저장소 전체 CI 판정은 **PENDING**이며 대응 CI run ID는 아직 없다.
6. 수정 파일 85% 기준 전체 충족을 주장하지 않는다. 새 replay_scope/receipt_history/reasoning/native 파일은 별도 측정에서 대체로 89~98%였으나 기존 conversation/family.go는 약 74%다. CLI 수정 파일별 coverage도 추가 측정이 필요하다.

## Residual-risk

- reasoning 검증 뒤 SSE를 보내므로 첫 표시가 늦어질 수 있다. 지연은 측정하지 않았다.
- Claude의 기본 1M 항목을 비활성화하여 정확한 등록 ID를 유지한다. 실제 effective context/압축 시점은 추가 확인 대상이다.
- 계열이 다른 GPT opaque 이력은 새 대화가 필요하다. 오류는 현재 adapter의 일반 오류 envelope로 축약될 수 있으므로 사용자 안내 개선도 검토해야 한다.
- native namespace 격리가 기존 사용자 설정·인증·MCP 구성을 충분히 보존하는지 실제 cc/glm/GPT 실행으로 확인해야 한다.
