# t649 Claude [1m] 런처 오류 수정

- Claim: 알려진 Claude Opus 5·Sonnet 5의 `[1m]` 표기를 카탈로그에 등록하고 upstream 기본 ID로 해석한다. Claude 실행에서 1M 비활성화 값을 강제로 주입하지 않는다. GPT·GLM의 등록되지 않은 접미사는 거절한다.
- Evidence: 설치 전 실제 CLI 격리 실행 `moai cc -p default -f lane-2`는 exit 1, `Gateway model not registered: claude-opus-5[1m].`이었다. 수정본과 설치 후 같은 실행은 exit 0, `CLAUDE_FIXTURE_LAUNCHED`이며 자식 argv의 model은 `claude-opus-5[1m]`, name은 `lane-2`이다. 원문은 cli-fixture.json과 installed-cli-fixture.json 참조.
- Evidence: `go test -p 1 ./internal/cli ./internal/gateway -run 'TestGateway|TestNativeGateway|TestCatalog|TestAnthropic' -count=1 -timeout=120s`의 이 실행 출력:

```text
ok  github.com/modu-ai/moai-adk/internal/cli      1.859s
ok  github.com/modu-ai/moai-adk/internal/gateway  0.503s
```

- Baseline-attribution: WT-unified-gateway, HEAD 81c1d58f9cf7045594ee61d5e4ff380948ce9eba의 기존 미커밋 작업에 최소 수정. 정확한 소스 차이는 hotfix.patch, source-diff.json, 새 gateway_context_suffix_test.go에 있다. 설치 경로는 /Users/goos/go/bin/moai, 빌드 ID는 v3.2.0-rc.8-t649-context-qualifier-a34c7e157d1a. 바이너리 SHA-256은 495ac493deb33d909f91888bd1b89112d13bdb8b026059ef08818dc224a3cc39. 원본 백업과 설치 읽기 확인은 build.json 참조.
- Gaps: 실제 mo.ai.kr 프로젝트에서 사용자의 Claude 대화 성공, 1M 실사용 용량, Windows 실행은 이 수정에서 검증하지 않았다. 대역 CLI는 실행 경로와 인수 전달을 검증한다. Codex App Server 전환 완료를 뜻하지 않는다.
- Residual-risk: Claude 계정 권한, upstream 서비스 상태에 따른 응답은 별도이다. 원래 프로젝트 설정은 변경하지 않았다. 최초 설치 후 확인 스크립트의 고정 argv 인덱스 오류는 플래그 검색으로 고쳐 저장된 실제 실행 출력을 재검증했다.
