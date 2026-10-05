# Card t847 Verdict — todo 서브프로세스 테스트의 stdout 전용 파싱 수리 (Class B)

- Date: 2026-09-14 · Lane: lane-4 · Branch: `WT-todo-stdout-parse` @ 19d81d48a (base 로컬 develop 93ae49ce7, unpushed)
- Class B — 재현 3/3 → 수리 → 판정 기준 3종 통과

## Claim

t705 수리로 stderr에 켜진 임시-기원 안내문이 서브프로세스 결합 출력에 섞여 3개 테스트가 깨지던 것을, 테스트만 고쳐 수리했다(제품 코드·todo.go 안내문·t705 동작 무변경). 판정 기준 3종(-race·TestTodo 전량·lint 0) 전부 통과.

## 원인 (측정으로 확정)

- 동시성 테스트 2건: `cmd.CombinedOutput()`을 id 파싱에 사용 → 안내문 산문 어휘("temporary", "the", "todo:")가 발급 id로 계수됨 — 실측 `distinct issued ids printed = 11, want 8` / `= 7, want 4`
- history 테스트: `errOut == ""` 단언과 안내문 줄 충돌 — 실측 `history t9999 stderr = "moai todo: the launch directory is inside the temporary root ..."`

## Evidence

| 단계 | 결과 | 증거 |
|---|---|---|
| RED | 3/3 FAIL 관측 | `go test -count=1 -run '^(TestTodoAddPick_ConcurrentProcesses\|TestTodoConcurrentAdd_8Processes\|TestTodoHistoryDisclosesPreArchiveQueue)$' ./internal/cli/ -v` → exit 1, --- FAIL 3 |
| 수리 | 테스트 전용: 결합 출력 → stdout/stderr 분리(stdout만 파싱, stderr는 실패 메시지로), history는 안내문 줄 필터 헬퍼 `withoutTempOriginAdvisory` | 커밋 19d81d48a (2파일, +41/−9) |
| GREEN | 3/3 PASS | 동일 셀렉터 exit 0, --- PASS 3 / --- FAIL 0 |
| -race | ok | `go test -race ...` → ok 11.986s |
| TestTodo 전량 | exit 0, FAIL 0 | `go test -count=1 -run 'TestTodo' ./internal/cli/` → ok 418.557s |
| lint·vet·gofmt | 0 issues / ok / clean | `golangci-lint run ./internal/cli/`, `go vet`, `gofmt -l` |
| 결합 출력 전수 | 잔여 2곳 = git 픽스처 헬퍼(131·208행) — todo 서브프로세스 파싱 아님 | `grep -n CombinedOutput internal/cli/todo_test.go` |

## Baseline-attribution

모든 수치는 lane-4 세션에서 이 워크트리(base 로컬 develop 93ae49ce7)에서 실행한 명령의 관측값.

## Gaps

- internal/cli 전체(셀렉터 무제한) 미실행 — TestTodo 셀렉터가 이 변경의 영역 전부; 전체 판정은 CI 몫
- t9999 외 다른 "stderr 공백" 단언의 존재 여부는 TestTodo 전량 PASS로 커버(없음이 확인된 것과 동치)

## Residual-risk

- 안내문 서두 문구(`"moai todo: the launch directory is inside the temporary root"`)가 제품에서 바뀌면 필터가 다시 어긋남 — 의도된 결합(t705 문안과 한 몸), 변경 시 함께 고칠 것
