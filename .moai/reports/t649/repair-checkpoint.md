# t649 F1·F2 및 GPT medium 수리

## Claim

F1 공개 message 경계·phase 복원, F2 최종 송신 원문 보존, 관측된 GPT medium 정책 오류를 수정했다. 실제 GPT 응답·도구·재개·managed 정책 인수 및 rc.8 배포 완료는 주장하지 않는다.

- reasoning이 포함된 출력은 `moai_opaque_v2_`/version 2 envelope에 public_items의 output_index/type/blocks/phase/phase_null을 함께 결합한다. 공개 텍스트를 중복 저장하지 않는다. 같은 message의 여러 part와 서로 다른 message를 구별한다. absent/null/commentary/final_answer를 구별한다. V1은 기존 의미로 계속 검증한다.
- metadata schema는 개수·인덱스·타입·phase를 제한하고 canonical 재직렬화 대조로 미지 필드·비정규 표현을 거절한다. 기존 receipt 권한을 그대로 요구하며 phase만 바꾼 carrier도 거절하는 실 store 시험을 추가했다.
- 최종 endpoint 정책 변환은 검증된 RawMessage 값을 직접 보존하는 정렬 object encoder를 사용한다. max_output_tokens 제거와 stream 강제 이후에도 synthetic reasoning item 바이트가 동일함을 실제 adapter와 로컬 TLS upstream으로 확인했다.
- 명시 medium을 GPT reasoning.effort=medium으로 전달한다. 누락은 채우지 않고 high 및 기존 거절 계약을 유지한다. native Anthropic의 effort 범위는 확대하지 않는다.

## Evidence

`repair-evidence/`에 실행 원문과 파일 digest를 보존했다. repair-red.txt의 세 회귀는 경계 유실, reasoning 위치 불일치, 최종 송신 바이트 유실로 모두 FAIL이었다. null-red.txt는 명시 null phase 유실, medium-red.txt는 unsupported effort로 FAIL이었다. 아래는 최종 실행이다.

```text
go test ./internal/gateway/... -race -count=1 -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/gateway 9.498s
ok  github.com/modu-ai/moai-adk/internal/gateway/auth 6.440s
ok  github.com/modu-ai/moai-adk/internal/gateway/conversation 1.546s
ok  github.com/modu-ai/moai-adk/internal/gateway/opaque 2.420s
ok  github.com/modu-ai/moai-adk/internal/gateway/receipt 3.189s
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 2.542s

go test ./internal/cli -run '^(TestGateway|TestNativeProvider|TestGPT)' -count=1 -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/cli 1.498s

go vet ./internal/gateway/...
exit 0; stdout/stderr empty

go build -o /tmp/moai-t649-repaired ./cmd/moai
exit 0; stdout/stderr empty
```

정확한 원문 공백·탭까지 포함한 출력은 repair-race-final.txt와 repair-cli.txt를 참조한다. race 첫 실행의 한 시험은 유효한 `[public, reasoning]` layout을 잘못 거절해야 한다고 작성한 시험 오류로 실패했다. 그 잘못된 변형만 제거한 후 전체 race를 다시 실행했다. 생산 불변식을 완화한 것이 아니다.

## Baseline-attribution

실행 트리는 moai-proxy-unified, branch WT-unified-gateway, HEAD 81c1d58f9의 기존 dirty 변경을 포함한다. 최종 파일 digest는 repair-evidence/baseline.json이다. 바이너리는 `/tmp/moai-t649-repaired`, SHA256 `5b259dbbaee35c325d3ff8dd959d7192ff9a6ba91edb2dabab5a18bbc8a792d7`이다. 커밋·push·PR·병합·설치는 수행하지 않았다.

## Gaps

- live 호출은 root/E2E 담당이 별도 실행한다. 이 담당자는 유료 요청을 실행하지 않았다. diagnostic2/3의 root 관측은 요청 변환의 unsupported effort 및 effort_medium이며 upstream 오류로 단정하지 않는다.
- managed effective 설정 preflight는 아직 미구현이다. 단순 managed 파일 검사로 모든 MDM/server 출처를 검증했다고 주장할 수 없다.
- reasoning 없는 응답의 공개 경계·phase 보존은 이번 bounded F1 범위 밖이다. metadata-only envelope는 만들지 않는다.
- 수정 파일 전체 coverage 85% 충족을 재측정하지 않았다. 저장소 전체 통합 CI 판정은 PENDING, 대응 CI run ID 없음. Windows CI 역시 이 작업에서 실행하지 않았다.

## Residual-risk

모형 시험 통과는 실제 Claude native transcript의 보존을 증명하지 않는다. 실제 답변·도구 왕복·새 프로세스 resume/fork 및 managed 충돌 인수가 필요하다. 임시 진단은 Go overlay로만 빌드했고 생산 소스에 T649_DIAG 문자열이 없음을 rg 무일치(exit 1)로 확인했다.
