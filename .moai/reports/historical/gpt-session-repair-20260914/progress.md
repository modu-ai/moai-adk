# moai gpt 세션 복구 검증 기록

최종 로컬 검증 결과와 운영 검증 대기 항목은 [수정 검증 보고서](report.md)에 정리했다. 이 파일은 수정 전 기준 기록을 보존한다.

## 작업 기준

- 요청: rc.11에서 발생한 502의 원인을 실제 로그와 코드로 연결하고, GPT 메인 세션의 도구·서브에이전트·취소/재개·스트리밍·동시 세션을 수정 및 검증한다.
- 기준 커밋: `b45c813493751106cd6619dce60fba1c61e70583`.
- 작업 공간: `.claude/worktrees/gpt-session-repair`, 브랜치 `WT-gpt-session-repair`.
- 세션: `01a09ebc-fbd0-7683-ac83-b9d315338def`.
- 문제 로그의 family: `48a351b8-4a08-48b5-887d-e5ceae71d74b`.
- 기존 develop의 수정 및 보고서는 이 작업 공간에 복사하거나 덮어쓰지 않았다.

## Claim

작업 진행 중이다. 전체 기능 호환 또는 배포 완료를 선언하지 않는다.

## Evidence

수정 전 해당 기준의 새 작업 공간에서 실행:

```sh
go test ./internal/codexapp ./internal/codexbridge ./internal/gateway -count=1
```

관측 출력:

```text
ok  	github.com/modu-ai/moai-adk/internal/codexapp	0.431s
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	12.859s
--- FAIL: TestGPTAppServerLifecycleAndHistoryAttribution (0.59s)
    --- FAIL: TestGPTAppServerLifecycleAndHistoryAttribution/changed-history (0.02s)
        gpt_lifecycle_history_contract_test.go:161: history_changed boundary status=502 generation=0 rpc=0 http=0 redacted-log=0, want 400/0/0/0/0
    --- FAIL: TestGPTAppServerLifecycleAndHistoryAttribution/agent-summary (0.02s)
        gpt_lifecycle_history_contract_test.go:168: agent_summary_untrusted boundary status=502 generation=0 rpc=0 http=0 redacted-log=0, want 400/0/0/0/0
    --- FAIL: TestGPTAppServerLifecycleAndHistoryAttribution/receipt-manifest-mismatch (0.03s)
        gpt_lifecycle_history_contract_test.go:174: receipt_manifest_mismatch boundary status=502 generation=0 rpc=0 http=0 redacted-log=0, want 400/0/0/0/0
    --- FAIL: TestGPTAppServerLifecycleAndHistoryAttribution/resume-duplicate (0.02s)
        gpt_lifecycle_history_contract_test.go:180: resume_duplicate_input boundary status=502 generation=0 rpc=0 http=0 redacted-log=0, want 400/0/0/0/0
    --- FAIL: TestGPTAppServerLifecycleAndHistoryAttribution/structured-rejection-logger (0.02s)
        gpt_lifecycle_history_contract_test.go:196: production ServerConfig RejectionLogger seam available=false settable=false type=<nil>, want injectable structured rejection recorder
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	9.408s
FAIL
```

전체 출력: `/tmp/moai-gpt-repair-baseline.log`.

## Baseline-attribution

위 결과는 rc.11의 소스 커밋을 새 작업 공간에 fast-forward한 뒤, 수정 전에 실행한 결과다. 설치된 바이너리의 실행 결과와 구분한다. 실행 중인 기존 Claude Code/GPT 프로세스는 종료하지 않았다.

## Gaps

- 최초 502의 내부 예외는 기존 서버에서 폐기되어 로그만으로 정확한 예외 분기를 확정할 수 없다.
- 모든 Claude Code 기능의 GPT 동등성은 관측되지 않았다. 개별 기능 테스트와 실사용 검증이 필요하다.
- 실제 GPT 통합 테스트는 기존 rc.11 프로필 점유의 해소가 필요하다.

## Residual-risk

- 인증 파일을 여러 App Server 프로세스가 동시에 갱신하지 않도록 단일 소유권을 유지해야 한다.
- 취소된 도구 실행을 자동 재실행하면 중복 외부 작업이 발생할 수 있다. 취소 확인과 도구 결과 대기를 구분해야 한다.
- 초기 텍스트를 스트리밍해도 성공 종료는 도구 상태의 영속화 또는 실제 turn 완료를 확인한 뒤에만 알릴 수 있다.
