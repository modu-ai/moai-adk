# t1472 판정서 — SPEC status 전이 안내 ↔ 무상태 lint 정렬 (GitHub #1736)

- card: t1472 · class B (SPEC 없음) · cycle_type: tdd
- branch: `WT-spec-status-doc-align` (기점: 로컬 develop `b5815ca80`, fast-forward 흡수)
- commits: `e575e3048` (RED 가드 단독) → `d8ea058d7` (GREEN 문서 정렬)
- 측정 트리: `d8ea058d7^{tree}` = `1a010c584f405164263a3941f6a2cb53a48a7b5d` (테스트 측정 후 편집 없음, `git status` 무출력)

## Claim

1. 상태 전이 안내 4개 문서(로컬·템플릿 각 1벌)가 plan.md / acceptance.md 에 `status:` 를 쓰라고 지시하던 문장을 모두 lint 와 같은 규칙으로 고쳤다: status 는 spec.md 프런트매터와(있으면) progress.md status 줄에만, plan.md / acceptance.md 는 `updated:` 갱신만.
2. 새 규칙은 만들지 않았다. progress.md 를 상태 보유 파일로 남긴 근거는 코드다 — lint 의 통제 집합에서 제외(`statelessArtifacts = plan/acceptance/design/research`, `internal/spec/lint_artifact_status.go:86`)이고, `closer.go` 의 `rewriteProgressStatusCompleted`(:482)가 progress.md 의 첫 `status:` 줄을 `completed` 로 고쳐 쓴다.
3. plan.md / acceptance.md 의 `updated:` 쓰기는 유지했다 — 해당 lint 는 status 축만 다루고 다른 프런트매터 필드는 명시적으로 허용한다(`lint_artifact_status.go` 헤더 "THE AXIS IS STATUS, NOT FRONTMATTER").
4. 리더 지목 3개 파일 외에 같은 모순의 4번째 사례 `manager-spec.md`(:4, :199 "all 4 plan-phase files ... spec.md + plan.md + acceptance.md + progress.md")를 함께 고쳤다. 같은 규칙의 사례라 범위 안으로 판단했고, 가드도 이 파일을 포함한다.

## Evidence

RED (`e575e3048`, 문서 수정 전):

```
$ go test -count=1 -run TestSPECStatusGuidanceMatchesStatelessLint ./internal/template/
--- FAIL: TestSPECStatusGuidanceMatchesStatelessLint (0.00s)
    ... template rules/moai/development/spec-frontmatter-schema.md:137: status guidance names plan.md/acceptance.md without the statelessness carve-out: ...
    ... template agents/moai/manager-develop.md: names plan.md/acceptance.md as status carriers via "all 4 plan-phase" ...
    ... template agents/moai/manager-docs.md: names plan.md/acceptance.md as status carriers via "ALL 4 SPEC artifacts" ...
    ... template agents/moai/manager-spec.md:199: status guidance names plan.md/acceptance.md without the statelessness carve-out: ...
FAIL	github.com/modu-ai/moai-adk/internal/template	0.467s
$ grep -c "spec_status_stateless_guidance_test.go:" red.txt
40
```

GREEN (`d8ea058d7` 트리):

```
$ go test -count=1 -run TestSPECStatusGuidanceMatchesStatelessLint -v ./internal/template/
--- PASS: TestSPECStatusGuidanceMatchesStatelessLint (0.03s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.473s

$ go test -count=1 -timeout 30m ./internal/template/... ./internal/spec/...   # exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	896.388s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.799s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.148s
ok  	github.com/modu-ai/moai-adk/internal/spec	1155.746s

$ make agents-emit   # exit=0
AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.524s
$ make build         # exit=0 — catalog.yaml 해시 3건 갱신(manager-spec/develop/docs)
$ go vet ./internal/template/   # vet-ok
$ gofmt -l internal/template/spec_status_stateless_guidance_test.go   # 무출력
```

가드 술어: 줄이 plan.md/acceptance.md 를 언급하고 status 축 토큰(`status:`, 백틱 `status`, `→`)을 담으면서 무상태 예외 표지("stateless", "carry no `status:`")가 없으면 위반. 별도로 "all 4 …" 열거·`plan.md + acceptance.md + progress.md` 열거를 금지. 양성 대조: schema 사본에 `### Artifact Statelessness` 절이 없으면 실패(빈 입력으로 음성 스캔이 공허 통과하지 않게).

## Baseline-attribution

- 모든 측정은 이 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-ada139d0d67afe8ca`), RED 는 `e575e3048` 트리, GREEN 은 커밋 `d8ea058d7` 과 동일 트리(측정 후 편집 0)에서 이번 실행으로 관측.
- 기존 문서 고정 테스트 조사: `grep -rlnE "on the M1 commit start|Forbidden ownership crossings|single sync commit|records no per-artifact|ONLY status transition" --include='*_test.go'` → `internal/factorylane/merge_test.go`(무관 픽스처 문자열), `internal/template/docs_delegation_lane_flow_test.go`("single sync commit" — 다른 룰 파일 대상, 이번 수정이 지우지 않음). 옛 문장을 인용하는 고정 테스트는 없었고, 수정할 테스트도 없었다.

## Gaps

- 전체 스위트는 돌리지 않았다(로컬 규율 — 판정은 develop push 의 CI 몫, 현재 PENDING). `internal/template`·`internal/spec`·`agentemit` 패키지만 측정.
- `golangci-lint` 미실행(테스트 파일 1개 추가, `go vet`·`gofmt` 만 관측).
- 가드는 이 4개 문서만 스캔한다. 같은 모순이 다른 템플릿 문서(sync/run 워크플로 스킬 등)에 있는지는 grep 2종(`status.{0,80}(plan.md|acceptance.md)`, "all 4 … artifacts/files/frontmatter")으로 템플릿 `.claude` 트리를 훑어 이 4개 파일 외 적중 0 을 관측했을 뿐, 다른 문구 형태는 미측정.
- `make embed-check`(설치 바이너리 임베드 축) 미실행.

## Residual-risk

- 가드의 예외 표지는 문자열 기반이라, 같은 줄에 "stateless" 를 쓰면서 동시에 status 쓰기를 허용하는 문장은 통과한다.
- progress.md 의 status 줄은 프런트매터일 수도, §E.3 본문일 수도 있다(코퍼스에 둘 다 존재). 안내는 "where present, the progress.md status line" 으로 closer 동작을 그대로 따랐고, progress.md 의 status 위치 자체를 정하는 규칙은 이 카드 범위 밖이다.
- 로컬 C1 사본은 해당 문장만 같은 블록으로 미러했다(스크립트로 블록 유일성 확인 후 치환). C1↔C2 의 다른 의도적 분기는 건드리지 않았다.
- 원격 미푸시: push·develop 병합·GitHub 코멘트는 하지 않았다. CI 판정 PENDING.
