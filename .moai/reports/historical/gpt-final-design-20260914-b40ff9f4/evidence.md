# 최종 검토안 v3 — 실행 근거

## Claim

현재 develop의 명칭 적중 전수 목록을 확보하고, 기존 런처·모델 슬롯 계약을 선택 테스트했다. 새 설계 구현을 검증한 것은 아니다.

## Baseline-attribution

`git branch --show-current && git rev-parse --short HEAD` 출력:

```text
develop
15f3eacd7
```

`git log -4 --oneline` 출력:

```text
15f3eacd7 Merge branch 'WT-gateway-launchers' into develop (t654)
0804a58e2 Merge branch 'develop' into WT-gateway-launchers
9462756bd docs(SPEC-MOAI-GATEWAY-001): t654 card verdict — AS-5 window closed, 6 window-waiting gaps (t654)
ce4a3c6ed docs(SPEC-MOAI-GATEWAY-001): sync-phase t654 — audit-ready, CHANGELOG deferred to deploy-gate window
```

이전 `c9ceff175` 보고서의 test 시간을 현재 baseline 결과로 옮기지 않았다. 이번 작업은 보고서 고유 디렉터리만 추가했다.

## Evidence

### 전수 검색

먼저 `rg --json --hidden -i 'kanban|칸반' ...`을 실행했지만 대형 HTML/보고서로 출력이 잘렸다. 이어 tracked 파일 전체를 기준으로 재실행했다. 잘린 두 출력은 최종 집계에 사용하지 않았다.

재현 명령: develop root에서 `node reports/gpt-final-design-20260914-b40ff9f4/scan-names.js`.

스캐너는 `git ls-files -z`의 모든 파일을 읽고 바이너리(NUL 포함)와 symlink를 따로 집계한다. 텍스트를 줄 단위 `/kanban|칸반/i`로 검사하고 경로도 따로 검사한다. 전체 출력은 [inventory.json](inventory.json)에 baseline metadata를 덧붙여 보존했다.

원출력 JSON의 집계 필드:

```json
{"tracked":18834,"textFiles":18565,"missing":[],"symlinks":[],"filesWithHits":1728,"pathMatches":262,"matchedLines":13901}
```

바이너리 269개. 내용/경로 union 1,733개. [inventory.md](inventory.md)는 모든 적중 파일과 줄 번호를 생략 없이 담는다.

### 현재 CLI 선택 테스트

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-gpt-factory-research-cache go test ./internal/cli -run 'TestGPTLaunchPreservesCommonEntry|TestGatewayProviderContractPickerAndSlots|TestGatewayLaunchAssemblyCarriesAuthDisplay|TestGatewayLaunchTransportGateControl|TestParseFactoryFlag' -count=1 -v -timeout 90s
```

exit 0. 패키지 원출력:

```text
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.000s
```

verbose 출력에서 `TestGatewayLaunchAssemblyCarriesAuthDisplay`의 gpt/claude/glm, `TestGatewayLaunchTransportGateControl`, `TestGatewayProviderContractPickerAndSlots`, `TestGPTLaunchPreservesCommonEntry`, `TestParseFactoryFlag` 및 pass-through 경계의 PASS를 확인했다. 테스트 이름 목록은 요약이며 새 설계의 실계정 성공 증거가 아니다.

### 현재 소스 경로 검사

`rg -n 'func productionGateway|PolicyGPTNative|NewCodexBroker|NewAppServerAdapter' internal/cli --glob '*.go' --glob '!*_test.go'` 출력:

```text
internal/cli/gateway_product_binding.go:198:func productionGatewayHandlerFactory(raw json.RawMessage) (http.Handler, error) {
internal/cli/gateway_product_binding.go:213:        policy = translate.PolicyGPTNative
```

들여쓰기는 표시를 위해 공백으로 정규화했다. 별도 `sed -n '248,270p' internal/cli/gateway_product_binding.go`에서 `return auth.CodexBroker{Executable: executable, Timeout: 10 * time.Minute}, nil`을 읽었다. 이 검색의 범위는 해당 CLI non-test source이며 저장소 모든 동적 경로에 대한 형식 증명은 아니다.

`rg --files internal | rg '^internal/(orchestration|workflow|todo|factory|execution)/'`에서 기존 `internal/workflow` 여섯 파일을 확인했다. 새 이름 `internal/orchestration`과 충돌하는 tracked 파일은 이 검사에서 나오지 않았다.

### 최신 공식 문서

- [Claude 모델 설정](https://code.claude.com/docs/en/model-config): 네 alias env와 main 선택·allowlist 경계.
- [Codex App Server](https://developers.openai.com/codex/app-server/): dynamicTools와 item/tool/call의 experimental 상태.
- [Claude gateway](https://code.claude.com/docs/en/llm-gateway): non-Claude routing 비지원 경계.

문서를 이번 실행에서 열고 해당 항목을 다시 검색했다. 실제 모델 호출이나 로그인은 하지 않았다.

## Gaps

- tracked 텍스트 적중의 전수 목록이지 모든 적중의 의미·외부 소비자 전수 검증이 아니다.
- 바이너리 내용, 다른 worktree, git 역사 전체, untracked runtime, 사용자 홈 배포 복사본은 제외했다.
- 새 이름·새 mapping·모드 폐기·DB migration·실제 다중 lane은 미구현/미실행이다.
- 기존 카드에 기록된 모든 테스트·라이브 Gap을 이번에 재실행하지 않았다.
- 초기 대형 검색 출력 잘림과 잘못 가정한 두 파일 경로/glob 검색 실패는 발견 단계의 도구 오류였고 완료 근거에 포함하지 않았다.

## Residual-risk

동일 문자열도 실제 실행 계약, legacy reader, 역사 식별자에서 의미가 다르다. 전수 치환을 구현 계획으로 오인하면 복구 경로와 증거를 손상시킬 수 있다. 예외 목록과 소비자 검증을 거친 뒤 변경한다.

## v3 추가 근거 — 지정 문서와 HTML 검사

### 문서 범위

[documentation-review.md](documentation-review.md)에 지정 5개 원문과 모든 하위 목차를 기록했다. 직접 링크 102개는 앵커별 중복 제거이며 FULL 21, SECTION 9, NOT_READ 71, GAP 1이다. 페이지 수나 읽기 완료율로 환산하지 않는다. 연결 문서 전체를 재귀적으로 읽은 것은 아니다.

한국어 settings-reference는 curl exit 56 두 번과 웹 open 실패가 있었다. 영어 공식 `https://code.claude.com/docs/en/settings-reference.md`는 exit 0으로 가져왔고 modelPicker 절을 별도로 끝까지 읽었다. 이 fallback으로 project/local 설정에서 modelPicker가 무시되는 점을 확인했다. 대형 출력 몇 건의 잘림은 FULL 판독 근거로 삼지 않았다.

### 실행 버전

`codex --version` 원출력:

```text
codex-cli 0.154.0
```

`claude --version` 원출력:

```text
2.1.270 (Claude Code)
```

이전 단계에서 같은 Codex 버전으로 생성한 `/tmp/moai-gpt-factory-schema-5fWMMB` JSON을 이번에 Node로 읽어 key와 description을 검사했다. 새 schema 생성·실제 RPC 검증은 이번 v3에서 하지 않았다. ThreadStartParams에 instructionSources 없음, ThreadResumeParams에 dynamicTools 없음, multiAgentMode가 ignored라는 description을 관찰했다. 현재 모델 권한이나 native 도구 차단 성공을 의미하지 않는다.

### HTML 생성·readback

`pandoc report.md --standalone --template=report.template.html --output=report.html`을 절대 경로로 실행했다. exit 0.

`wc -c <report.html>` 원출력:

```text
   65009 /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop/reports/gpt-final-design-20260914-b40ff9f4/report.html
```

Node로 HTML href의 로컬 경로를 해석하여 파일 존재 여부를 검사했다. 원출력:

```json
{"htmlBytes":65009,"localLinks":19,"missing":[],"containsFourModels":true,"unexpandedPandocVariable":false}
```

### 브라우저 검사

`agent-browser --session gpt-design-v3 --allow-file-access open file:///Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop/reports/gpt-final-design-20260914-b40ff9f4/report.html` 성공.

`agent-browser --session gpt-design-v3 set viewport 1440 1000` 후 DOM eval 결과(JSON 문자열을 decode한 값):

```json
{"title":"moai gpt 재설계 v3 — Gateway × Codex App Server","width":1440,"scrollWidth":1440,"sections":13,"toc":13,"brokenAnchors":[],"mermaidSVG":true,"fallbackVisible":false}
```

`agent-browser --session gpt-design-v3 set viewport 390 844` 후 DOM eval 결과:

```json
{"width":390,"scrollWidth":390,"tables":13,"tableScrollRegions":13,"bodyOverflow":false}
```

[desktop.png](desktop.png), [mobile.png](mobile.png)를 캡처하고 실제 이미지를 열어 상단 레이아웃을 확인했다. 각 표의 내부 가로 스크롤을 허용하고 페이지 전체의 가로 넘침은 없었다. 전체 페이지의 모든 픽셀과 모든 외부 링크를 시각 검사한 것은 아니다.

목차의 6번 실행부 링크를 실제 클릭해 해당 fragment URL로 이동한 것을 확인했다. 최초 위치 검사 스크립트는 percent-encoded hash를 querySelector에 넣어 SyntaxError가 났다. getElementById(decodeURIComponent(...))로 수정하여 목표 heading의 존재를 확인했다. 로딩 직후 top 좌표는 최종 스크롤 위치 인수 근거로 삼지 않았다. 페이지 자체의 오류로 귀속하지 않는다.

`open <report.html>` exit 0으로 기본 브라우저 열기를 요청했다. 사용자가 실제로 읽었음을 뜻하지 않는다. 자동화용 named browser session은 `agent-browser --session gpt-design-v3 close`로 종료했고 원출력은 `✓ Browser closed`다.

### 마지막 기준 재확인

`git -C <develop> diff --stat` 출력 없음. 이어 `git -C <develop> rev-parse --short HEAD`, `git -C <develop> branch --show-current` 원출력:

```text
15f3eacd7
develop
```

tracked 제품 파일 변경은 없다. 새 보고서와 캡처는 untracked이며 기존 다른 보고서 디렉터리를 삭제·수정하지 않았다. primary checkout의 브랜치를 전환하지 않았다. commit/push/로그인/유료 호출/Factory 실행/DB migration은 하지 않았다.

## 모델 직접 매핑 승인 반영 — develop 4056f69e1

### Claim

별도 모델 티어 열을 삭제하고 네 별칭의 직접 매핑과 독립 effort 정책을 MD·HTML에 반영했다. 제품 코드 구현 완료를 뜻하지 않는다.

### Baseline-attribution

수정 및 검증 경로는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop`이다. `git branch --show-current`는 `develop`, `git rev-parse --short HEAD`는 `4056f69e1`이었다. 이전 조사·테스트는 15f3eacd7 기준으로 보존했다.

### Evidence

해당 develop worktree에서 실행한 렌더링 명령(exit 0):

```sh
pandoc reports/gpt-final-design-20260914-b40ff9f4/report.md --standalone --template=reports/gpt-final-design-20260914-b40ff9f4/report.template.html --output=reports/gpt-final-design-20260914-b40ff9f4/report.html
```

`node -`로 MD·HTML을 readback하고 assert/strict로 매핑 4행·3열, 각 모델 ID, 구 티어 열 부재, effort 분리 문구, 로컬 href 파일 존재, 120KB 제한을 검사했다. 원출력(exit 0):

```json
{"mappingRows":4,"columns":3,"obsoleteTierColumn":false,"effortIndependent":true,"htmlBytes":67032,"localLinks":19,"missingLinks":[]}
```

최초 검사에서 Lite 부분문자열을 차단하는 정규식이 기존 라이브러리 비교의 LiteLLM까지 잡아 실패했다. 단어 경계를 지정하여 재검사했고 해당 비교 문단은 보존했다. 이는 검사 스크립트의 범위 오류이지 제품 결함이 아니다.

`open <report.html>` exit 0. `git diff --stat` 출력 없음. 마지막 HEAD도 `4056f69e1`이었다.

### Gaps

이번 수정 후 새 브라우저 스크린샷·실계정 호출·제품 테스트는 실행하지 않았다. 기존 스크린샷과 1.000s 테스트 결과는 이전 판본·이전 HEAD의 근거다. 모델 매핑 승인을 전체 gateway 재설계와 상태 이관의 포괄 승인으로 확대하지 않았다.

### Residual-risk

모델별 실제 effort 허용값과 부모 effort 상속 계약은 제품 구현 단계에서 검증해야 한다. 이번 확인은 설계 문서의 일관성 검사다.
