# PR #1870 native review 수리 근거

작업: t1913 / ci-repair-20261022. 역할: manager-develop 구현 담당. 독립 감사가 아니다.
측정 경로: `/workspace/moai-adk/.moai/worktrees/cloud-hook-repair`,
branch `WT-cloud-hook-repair`.

## Claim

실제 native review 다섯 항목을 대상으로 잘못된 receipt 시간, 서로 다른 payload
프로젝트 경로, YAML 별칭 정책에 대한 회귀를 추가했다. wrapper는 불확실한 경우
기존 handler로 위임한다. chain-event는 수정된 v1.0.5 요구사항에 따라 동기 등록을
유지하며, 두 관찰 hook의 async 등록과 이 handler의 additionalContext 출력 정책은
유지한다. upstream은 async의 두 메시지 필드를 모두 다음 turn의 Claude에 전달한다.

## Evidence

원문은 이 tree의 `work/ci-repair-20261022/` 아래 보존했다.

- 수정 전 실제 source awk 실행은 `2026-10-21T23:49:08bogus`를
  `fresh pass binding reviewer`로 잘못 판독했다 (`pr1870-parser-red-probes.json`).
- 같은 source의 payload probe는 `cwd=/current, project_dir=/other`에서 `/current`만
  추출했다. YAML alias는 `*skip`, merge는 빈 값으로 읽어 skip 분기를 통과했다
  (`pr1870-policy-red-probes.json`).
- 첫 RED 명령:
  `go test -p 1 ./internal/cli ./internal/template -run '^(TestCodexEntryFastpath_Review.*|TestCodexEntryFastpath_SingleKickPerBucket|TestHookSyncAsyncRegistrationFlags)$' -count=1 -timeout 30m`.
  exit 1, `FAIL github.com/modu-ai/moai-adk/internal/cli 52.749s`,
  `FAIL github.com/modu-ai/moai-adk/internal/template 0.040s`.
  잘못된 시간 12개, 혼합 경로 2개, YAML 정책 3개가 실제 실패했다.
  live/rendered 둘 다 async inventory=14, want13 및 chain-event async=true, wantfalse였다
  (`pr1870-native-review-red.log`, `.exit`).
- 새 시간 회귀가 CI의 Codex 설치 여부와 무관하게 실행되도록 테스트 전용
  reviewer `--version` 실행파일을 PATH 경계에 배치했다. 이 fixture는 실제
  RecordReceipt/Load, ConfigDigest, ToolVersion 검증 경로를 그대로 사용하고,
  실제 리뷰나 감사 결과를 생성하지 않는다.
  추가 RED 명령:
  `go test -p 1 ./internal/cli -run '^TestCodexEntryFastpath_ReviewRejectsMalformedTimestamp$' -count=1 -v -timeout 30m`.
  exit 1, `FAIL github.com/modu-ai/moai-adk/internal/cli 38.515s`.
  16개 모두 실행했고 skip 없이 invalid12 FAIL, valid4 PASS를 관측했다
  (`pr1870-hermetic-red.log`, `.exit`).
- 공식 문서 `https://code.claude.com/docs/en/hooks.md`를 정상 HTTP200으로 읽었다.
  4045행은 비대화형 teardown의 async 취소, 4052행은 additionalContext와
  systemMessage의 다음 turn 전달 및 두 필드의 사용자 UI 비표시를 설명한다
  (`pr1870-official-hooks.md`, `.headers`, `pr1870-official-hooks-http.txt`).
- 수정 source의 실제 receipt awk 14개(잘못된 시간10/유효 시간4) 및 live async13,
  양쪽 wrapper/doc byte parity, bash syntax, live JSON parse, git diff check는
  exit0이었다 (`pr1870-parser-green-probes.json`). 전체 수정 검증은 아래 후속
  실행 결과에 귀속하며 이 개별 probe만으로 전체 성공을 주장하지 않는다.

- `GOFLAGS="$GOFLAGS -p=1" make -j1 build`: exit0. 기존 `-mod=readonly`는 유지했다.
  agentemit0.031s, commandemit0.007s, toolpolicy0.022s, embedemit0.041s를 통과했고,
  공식 templ 생성(변경0), catalog70개 hash 재계산 및 binary build가 완료됐다.
  catalog의 tracked diff는 없다 (`pr1870-make-build.log`, `.exit`).
- 첫 전체 CLI 영향 검증:
  `go test -p 1 ./internal/cli -run '^(TestCodexEntryFastpath.*|TestHookEntryParity.*)$' -count=1 -v -timeout 30m`.
  exit1, `FAIL github.com/modu-ai/moai-adk/internal/cli 31.289s`.
  새 시간16개, 혼합경로3개, YAML3개와 기존 receipt/key/binding/marker/clock은 모두
  통과했다. SingleKick3개도 전부 실행·통과했고 skip은0이었다. 유일 실패는 기존
  `TestHookEntryParity_ChainEventRegistered`의 옛 async=true 기대였다. v1.0.5의 동기
  내구성 계약에 맞게 정확한 async=false 검사로 변경했고, 등록 수1/이벤트/timeout5
  검사는 유지했다. 이 새 test 수정 이후 동일 전체 CLI 범위를 다시 측정한다
  (`pr1870-cli-green.log`, `.exit`; 첫 source11개와 후속 source12개 hash는
  `pr1870-green-source-hashes.json`, `pr1870-green-source-hashes-v2.json`).

- 최종 CLI 명령은 바로 위와 동일하며 새 동기 등록 assertion을 반영한 tree에서
  exit0, `ok github.com/modu-ai/moai-adk/internal/cli 25.817s`였다. 실제 verbose
  이름 기준 15개 top-level 함수와 68개 하위 사례, 총83 RUN/83 PASS, SKIP0/FAIL0이다.
  시간16개 모두 non-skip PASS, SingleKick3개도 PASS다
  (`pr1870-cli-green-v2.log`, `.exit`, `pr1870-cli-green-counts.json`).
- template/handler 영향 명령:
  `go test -p 1 ./internal/template ./internal/hook -run '^(TestTemplateNeutrality.*|TestHookSyncAsyncRegistrationFlags|TestHookWrapper.*|TestNoHardcodedAbsolutePath_HookWrappers|TestCatalogHash.*|TestManifestHashFormat|TestBindingLedgerIntegrity|TestPostToolUseFailureHandler.*|TestChainEventHook.*)$' -count=1 -v -timeout 30m`.
  exit0, `ok github.com/modu-ai/moai-adk/internal/template 6.093s`,
  `ok github.com/modu-ai/moai-adk/internal/hook 0.500s`.
  template는 10개 top-level/36 RUN·PASS, hook은 22개 top-level/53 RUN·PASS,
  양쪽 SKIP0/FAIL0이었다. 양 surface의 정확한 async13/sync chain, template neutrality,
  catalog/manifest/binding ledger, 실제 completion-edge 저장 및 handler 메시지를 검사했다
  (`pr1870-template-hook-green.log`, `.exit`, `pr1870-template-hook-counts.json`).
- `go vet -p 1 ./internal/cli ./internal/template ./internal/hook`: exit0, 출력 없음
  (`pr1870-vet.log`, `.exit`). 위 세 scoped 패키지 실행의 총합은 47개 top-level과
  125개 하위 사례, 172 RUN/172 PASS이며 skip/fail은 모두0이다.
- commit 전 정상 원격 경로 `git fetch origin main`: exit0.
  `origin/HEAD=origin/main`, `git rev-list --count --left-right origin/main...HEAD`
  출력 `0 38`, 최신 main은 `bf0a35f10e423d8f0768e637b74a58e143a5d7d6`였다.
  원본 #1867 `5f1567dc02f609185b77fd51bbc32e70b31d9cd0`, common670 및 최신 main의
  HEAD ancestry는 각각 exit0이었다 (`pr1870-precommit-fetch.*`, `pr1870-ancestry.json`).

## Baseline-attribution

RED 기준은 `ee80624c2029c619b63da18e369f71ded79d2ac7` 위의 새 회귀 검사이며
production source는 두 RED 영수증 회수 전까지 변경하지 않았다. SPEC 세 본문은
manager-spec가 작성했고 이 구현 담당은 수정하지 않았다.

## Gaps

새 수정 head의 외부 required CI, 코드/보안 리뷰, 독립 감사와 merge는 미관측이다.
push/PR 변경은 이 구현 담당이 수행하지 않았으며 게시 담당이 후속 검증을 수행한다.
repository 전체 판정은 integration branch의 실제 CI가 소유하며 PENDING이다.
별도 LSP 진단은 측정하지 않았다. 이 문서는 구현 근거이며 독립 감사가 아니다.

기존 head08b의 실제 Ubuntu/Race CI는 SingleKickPerBucket에서 kick0/argv empty로
실패했다. 현재 Cloud의 source 수정 전 Go RED와 두 entry shell probe에서는
이 증상이 재현되지 않았다. 시간 구간 경계가 이 CI 실패의 원인이라고 주장하지 않는다.
같은 bucket이라는 회귀의 전제는 test-only date 실행파일로 고정했고, forward0/kick정확1
기대는 유지·강화했다. 재발하면 두 wrapper stderr와 synthetic Git Trace2를
실패 출력에 포함하도록 진단을 추가했다. 근거 없는 반복 실행은 하지 않았다.

## Residual-risk

shell이 지원하지 않는 날짜/zone/YAML 모양은 기존 handler에 위임하므로 cold start가
추가될 수 있다. receipt key/config/tool/session 결합, TTL, FAIL 전달과 marker 정책을
완화하지 않았다. Windows 및 실제 Claude 비대화형 teardown 실행은 여기서 관측하지
않았고, vendor의 teardown 계약과 등록 검사/기존 completion-edge 회귀로 검증 범위를
명시한다. 과거 완료·감사 자료를 이 수정의 새 감사 결과로 취급하지 않는다.
