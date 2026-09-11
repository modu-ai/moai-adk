# t524 통합 창 — 흡수와 재측정

- 창: `moai integration acquire --name lane-8` 로 잡았다(리드 지명 후). 로컬 develop tip 은 `4b82591cb` 로 리드가 준 값과 같았다.
- 흡수 병합: `7a621d7cdebf699ae4397774609f58aab2cd5be2`, 트리 `e0c00c63b6e142a311c90286946f94f9c2d658c3`
  - 부모 1: 카드 브랜치 `446b40668`
  - 부모 2: develop `4b82591cb`
- 충돌: `catalog.yaml` 해시 한 줄. 생성기로 해소했고 과정은 `catalog-resolution.md` 에 있다.
- 흡수 델타(`446b40668...4b82591cb`)에 `go.mod`·`go.sum` 은 없었다. 리드 메시지의 전제와 다르지만, 재승인된 테스트는 그대로 실행했다.

## 트리 `e0c00c63` 에서 잰 것

| 측정 | 명령 | 결과 | 증거 |
|---|---|---|---|
| template 패키지 전체 | `go test -count=1 -timeout 580s ./internal/template/` (선택자 없음) | `go_test_exit=0`, `ok … 35.155s` | `template-test.{txt,exit}` |
| 방출 드리프트 | `make agents-emit-check` | exit 0 | `agents-emit-check.txt` |
| 사전 확인 | `ps -axo pid,args` + 판별식 | 외부 internal/cli 컴파일 0 / 대조군 claude 20 | `precheck-cli.txt` |
| 동사 테스트(리드 1회 재승인) | `go test -count=1 -v -run '^TestPlanAudit(D7|D8|Traceability)' ./internal/cli/` | `go_test_exit=0`, PASS 19 · FAIL 0 | `cli-verb-test.{txt,exit}` |

`ps-before-cli.txt` 원본 덤프는 슬롯 때와 같은 이유로 커밋하지 않는다. 다른 세션 명령줄 인자가 섞일 수 있어서다.
