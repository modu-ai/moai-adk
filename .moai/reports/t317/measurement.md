# t317 — 에이전트 정의 계보 드리프트: 조치안 3갈래 실측

- card: t317
- worktree: `.claude/worktrees/t317` (branch `WT-agent-emit-lineage`)
- base: `48eb945df` — reflog 증거 `worktree-t317@{0}: branch: Created from origin/develop`
- 측정 트리: 위 base, 이 카드의 코드 변경 0인 상태

모든 수치는 이 트리에서 이 실행으로 측정했다. 다른 트리·다른 시점의 값을 옮겨오지 않았다.

---

## Claim 1 — `make build` 는 `agents-emit` 을 부르지 않는다

**판정: 참(카드 전제 유지).**

Evidence — `Makefile`:

```
20: templ-generate: ## Generate *_templ.go from *.templ sources (pure-Go codegen, no Node)
21: 	go run github.com/a-h/templ/cmd/templ generate -path ./internal/web
23: build: templ-generate ## Build the binary
24: 	@go run ./internal/template/scripts/gen-catalog-hashes.go --all
25: 	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/moai
27: agents-emit: ## Regenerate the .codex/agents/moai TOMLs from the neutral .md layer
28: 	AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission
```

`build` 의 선행 타깃은 `templ-generate` 뿐이고, `agents-emit` 은 어떤 타깃의 선행으로도 등장하지 않는다.

부수 관측(같은 파일, 별건): `agents-emit` 이 `.PHONY` 목록(Makefile:16)에 **없다**. 동명의 파일·디렉터리가 생기면 make 가 조용히 건너뛴다. 현재 그런 경로는 없으므로 잠재 결함이다.

---

## Claim 2 — 재생성 누락은 "조용히" 되돌려진다

**판정: 거짓. 이 전제는 실측으로 반증된다.**

드리프트는 **기본 테스트 실행에서 잡힌다.** `golden_test.go` 의 update 분기는 `AGENTEMIT_UPDATE=1` 일 때만 타고, 환경변수 없는 기본 실행은 커밋된 `.toml` 과 방출 결과를 sha256 으로 비교해 불일치를 `t.Errorf` 로 세운다.

Evidence — `internal/template/agentemit/golden_test.go`:

```go
update := os.Getenv("AGENTEMIT_UPDATE") == "1"
...
if update { ...WriteFile...; continue }
committed, err := os.ReadFile(committedTOMLPath(p))
if fmt.Sprintf("%x", sha256.Sum256(committed)) != sum {
    t.Errorf("%s: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing", p)
}
```

Evidence — CI 가 그 기본 실행을 돌린다 (`.github/workflows/`):

```
ci.yml:183:                    go test -coverprofile=coverage.out -covermode=atomic ./...
ci.yml:238:                    go test -race -count=1 ./...
release-pr-multi-os.yml:189:   go test -race -timeout 25m ./...
```

`./...` 는 `internal/template/agentemit` 를 포함하고, CI 는 `AGENTEMIT_UPDATE` 를 설정하지 않는다. 따라서 **C2(.md) 를 고치고 재생성을 빠뜨린 PR 은 CI 에서 빨간불이 된다.** 손편집한 `.toml` 도 같은 검사에 걸린다.

이 반증이 조치안 (c) 의 "CI 로 드리프트 검출" 절반을 이미 충족된 것으로 만든다. (c) 에 남는 것은 편집 절차의 명문화뿐이다.

---

## Claim 3 — 재생성 이후에는 골든이 원리상 빨간불이 안 된다

**판정: 참이지만 결함이 아니다 — update 모드의 정의 그 자체다.**

`AGENTEMIT_UPDATE=1` 은 "커밋본을 방출 결과로 맞춰라"는 지시다. 맞춘 뒤 불일치가 없는 것은 동어반복이지 침묵이 아니다. 판정력은 update 를 걸지 않은 실행이 갖고 있고, 그 실행이 CI 에서 매번 돈다(Claim 2).

즉 (b) "골든이 사후에도 판정 가능하도록 update 분기 수정" 은 **없는 결함을 겨눈다.** update 모드가 사후 판정까지 하게 만들면, 재생성 명령이 재생성을 하고 나서 자기가 한 일을 실패로 보고하는 모순이 된다.

다만 update 분기에 남는 실제 약점 하나는 있다: `WriteFile` 후 되읽어 확인하지 않는다. 부분 쓰기·권한 실패가 아닌 한 문제되지 않으며, 위 모순과는 별개의 작은 사안이다.

---

## 측정 4 — 현재 트리에 드리프트 없음

```
$ go test ./internal/template/agentemit/...
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.419s
```

11개 TOML 전부 방출 결과와 sha256 일치. 이 카드는 존재하는 드리프트를 고치는 카드가 아니라, 드리프트가 **어느 지점에서 늦게 잡히는지**를 다루는 카드다.

---

## 측정 5 — build 는 이미 트리를 변형한다 (조치안 (a) 판정의 핵심)

(a) 에 대한 자연스러운 반론은 "빌드는 결정적·읽기전용이어야 한다"인데, **이 리포의 build 는 이미 읽기전용이 아니다.**

- `templ-generate` → `internal/web` 하위 `*_templ.go` 를 씀
- `gen-catalog-hashes.go --all` → `catalog.yaml` 을 in-place 로 씀
  (`internal/template/scripts/gen-catalog-hashes.go:262: os.WriteFile(*flagCatalogPath, outYAML, 0644)`)

따라서 `agents-emit` 을 `build` 선행에 넣는 것은 **새 범주의 부작용이 아니라 기존 build 계약과 동종**이다. 비용도 작다 — 해당 패키지 전체 테스트가 0.419s.

---

## 측정 6 — `.codex` TOML 은 catalog 무결성 층 밖에 있다

```
$ grep -n "codex" internal/template/catalog.yaml
(출력 없음 — 총 268행)
```

catalog 는 `.codex/agents/moai/*.toml` 을 한 항목도 담지 않는다. 방출물의 유일한 무결성 가드는 agentemit 골든 테스트 하나뿐이고, build 가 계산하는 catalog 해시는 이 방출물에 대해 아무것도 말하지 않는다.

---

## 측정 7 — C1 ↔ C2 는 바이트 동일 관계가 아니다 (결함 아님)

```
$ diff -rq .claude/agents/moai internal/template/templates/.claude/agents/moai
7개 파일 differ (builder-harness, e2e-tester, manager-develop,
                 manager-lead, manager-spec, plan-auditor, super-advisor)
$ diff .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md
1a2
> isolation: worktree
```

차이는 의도된 것(로컬 도그푸드 vs 배포판 중립화)으로 보이며 이 카드 범위 밖이다. **C1↔C2 패리티 가드 부재를 결함으로 주장하지 않는다** — 그 관계가 동일성이 아니기 때문이다. 이 카드가 다루는 계보는 C2 → C3 한 방향뿐이다.

---

## 세 갈래 판정

| 안 | 판정 | 근거 |
|---|---|---|
| (a) `build` 선행에 `agents-emit` 추가 | **부분 채택 — 단, 재생성이 아니라 검증으로** | 측정 5 가 "결정성" 반론을 약화시켰고 비용도 0.4s 다. 그러나 build 가 방출물을 **재생성**하면, 잘못된 손편집을 조용히 덮어 CI 가 볼 것을 없애 버린다. build 는 **검사**해야지 고쳐서는 안 된다 |
| (b) 골든 update 분기 수정 | **기각** | 측정 3. 없는 결함을 겨눈다. update 모드가 사후 판정을 하면 자기모순 |
| (c) 절차 명문화 + CI 드리프트 검출 | **절반 이미 충족 / 나머지 채택** | 측정 2 — CI 검출은 이미 있다. 남는 것은 로컬 편집 절차 명문화뿐이고, 그것만으로는 기계적 가드가 아니다 |

**권고 = (a′) + (c 잔여).** `build` 에 **읽기전용 드리프트 검사**를 선행으로 걸어 로컬에서 즉시 빨간불이 뜨게 하고(현재는 CI 까지 가야 안다), 편집 절차를 문서에 명문화한다. 재생성은 지금처럼 `make agents-emit` 이라는 **명시적 동사**로만 일어난다.

(a′) 가 (a) 보다 나은 이유는 하나다: 재생성은 증거를 지우고, 검증은 증거를 세운다. 손편집한 `.toml` 을 build 가 조용히 덮으면 그 손편집이 있었다는 사실 자체가 사라진다.

---

## Gaps — 관측하지 않은 것

- `make build` 직후 바이너리에 스테일 TOML 이 임베드되는 경로를 **실증하지 않았다**(수정→build→embed 바이트 비교 미실행). 코드 경로상 그렇게 되나, 실측은 run-phase 몫이다.
- CI 가 실제로 이 실패를 낸 이력(과거 빨간불 사례)을 조회하지 않았다.
- 7개 C1↔C2 차이 전부를 읽지 않았다 — manager-develop 1건만 표본으로 읽었다.
- `.PHONY` 누락이 실제 사고를 낸 적이 있는지 조회하지 않았다.

## Residual risk

- CI 가 드리프트를 잡는다는 판정은 워크플로 파일 판독에 근거한다. 러너에서 `AGENTEMIT_UPDATE` 가 다른 경로로 주입될 가능성은 배제하지 못했다(grep 상 등장 0회이나 전 환경 스캔은 아니다).
- 측정 5 는 build 가 이미 쓰기를 한다는 사실을 말할 뿐, 쓰기를 더 늘리는 것이 옳다는 결론까지 자동으로 주지는 않는다. 권고가 (a) 가 아니라 (a′) 인 이유이기도 하다.

---

## 발견 경위 (SPEC 배경에 남길 것)

이 결함은 t301 감사 중 **부산물로** 잡혔다. 감사관이 방출 배선을 겨눠서 찾은 것이 아니라, 어휘 토큰 전역 검색을 하다 `.codex` 사본이 세 번째 사본으로 존재한다는 사실이 눈에 걸린 것이다. 우연이 잡아낸 것이므로 같은 경로로 다시 잡힐 것을 기대할 수 없다 — 이것이 이 카드가 절차가 아니라 기계적 가드를 요구하는 이유다.

---

## 측정 8 — 임베드 FS 테스트는 자기 목적에 대해 공허하다 (뮤테이션 실증)

리드가 전달한 lane-16 축("사용자가 받는 건 `//go:embed all:templates` 로 박힌 임베드 자산인데 검증은 소스를 본다")을 이 패키지에 대고 직접 뮤테이션으로 세웠다.

`golden_test.go` 의 `TestEmbedFSPresenceAndByteEquality` 는 주석에서 "embedded bytes differ from committed (run make build)" 를 자기 실패 메시지로 갖는다. 즉 **make build 를 빠뜨린 상태를 잡겠다고 선언한 테스트다.** 실제로 잡는지 물었다.

뮤턴트: 커밋된 TOML 한 개 끝에 두 줄 주입.

```
$ printf '\n# mutant-t317\n' >> internal/template/templates/.codex/agents/moai/manager-git.toml
```

관측:

```
$ go test ./internal/template/agentemit/... -run TestEmbedFSPresenceAndByteEquality -count=1 -v
=== RUN   TestEmbedFSPresenceAndByteEquality
--- PASS: TestEmbedFSPresenceAndByteEquality (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.420s
```

**뮤턴트 생존.** 대조군으로 같은 뮤턴트에 골든 본체를 걸면 죽는다:

```
$ go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission -count=1
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
    golden_test.go:109: .codex/agents/moai/manager-git.toml: committed artifact differs from emission (sha256 mismatch)
FAIL
```

원복 및 원복 확인:

```
$ git checkout -- internal/template/templates/.codex/agents/moai/manager-git.toml
$ git status --short
?? .moai/reports/t317/          ← 이 카드가 만든 증거 디렉터리뿐
$ go test ./internal/template/agentemit/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.249s
```

**원인**: `go test` 는 테스트 바이너리를 매번 새로 컴파일하고, `//go:embed all:templates` 는 그 컴파일 시점의 **커밋된 파일 그대로** 읽어 들인다. 그래서 이 테스트는 "임베드 바이트 vs 커밋 바이트" 를 비교하는데 양변이 같은 원본에서 함께 갱신된다 — 동어반복이다. 스테일 임베드는 **미리 빌드된 바이너리**에만 존재하고, `go test` 는 그런 바이너리를 결코 보지 않는다.

**결과적으로 이 패키지에서 임베드 축을 판정하는 테스트는 하나도 없다.** 실패 메시지가 "run make build" 라고 말하지만, 그 상황을 만들어 낼 수 있는 실행 경로가 테스트 안에 없다.

### 이것이 세 갈래 판정에 미치는 영향

리드의 지적대로 두 발견은 한 쌍이다.

- 측정 1: 재생성이 build 에 안 걸려 있다 → 빠뜨리기 쉽다
- 측정 8: 임베드 층을 보는 눈이 없다 → 빠뜨린 결과가 배포 자산에서 관측되지 않는다

다만 측정 2 가 이 쌍의 파괴력을 한정한다: **소스 층 골든이 CI 에서 드리프트를 잡으므로**, 커밋된 `.toml` 이 스테일한 채 머지되는 경로는 막혀 있다. 남는 진짜 창은 **로컬에서 빌드해 쓰는 바이너리**다 — `make build` 로 만든 `bin/moai` 는 커밋된(=정확한) TOML 을 담으므로, 실은 `.md` 를 고치고 재생성하지 않은 상태에서 build 하면 **스테일 TOML 이 그대로 임베드된 바이너리**가 나오고 로컬에서는 아무 신호가 없다. CI 가 그 PR 을 빨간불로 세우기 전까지 그 바이너리로 작업이 진행된다.

따라서 (b) 를 "update 분기 수정"이 아니라 **"판정 지점을 임베드 층으로 옮기기"** 로 재해석하면 값이 있다 — 그러나 위 원인 때문에 `go test` 안에서는 그 판정을 세울 수 없다. 판정 지점은 **빌드 산출물(`bin/moai`)에 대고** 잡아야 한다. 이는 테스트가 아니라 build 이후 검사(또는 CI 단계)의 형태를 요구한다.

### 갱신된 권고

**(a′) 읽기전용 드리프트 검사를 `build` 선행에 + (b′) 판정 지점을 빌드 산출물로 + (c 잔여) 절차 명문화.**

- (a′) `build` 전에 `AGENTEMIT_UPDATE` 없이 골든을 돌려 소스 층 드리프트를 로컬에서 즉시 세운다. 재생성은 여전히 `make agents-emit` 이라는 명시적 동사로만.
- (b′) 임베드 축은 소스 비교로 성립하지 않는다(측정 8). 성립시키려면 `bin/moai` 가 실제로 담은 바이트를 꺼내 커밋본과 대조해야 한다.
- (c) CI 검출은 이미 있다(측정 2). 남는 것은 "C2 를 고쳤으면 `make agents-emit`" 을 편집 절차에 명문화하는 것.

### 측정 8 의 Gaps

- 뮤턴트를 TOML 한 개에만 심었다. 11개 전부에 대해 같은 생존을 확인하지는 않았다(기전상 동일할 것이나 미측정).
- `bin/moai` 를 실제로 빌드해 스테일 임베드를 꺼내 보이는 실증은 하지 않았다 — run-phase 몫이다.
- 이 공허성이 `agentemit` 밖의 다른 임베드 검증 테스트에도 있는지는 조사하지 않았다.

---

## 측정 9 — 추출 경로 독립 재현 (감사관 근거의 자기 확인)

plan.md M1 의 1순위 추출 경로는 감사관이 실행해 세운 것이고, SPEC 은 그 근거 위에서 폴백 2순위를 **삭제**했다. 폴백을 지운 판단이 남의 측정 하나에만 기대게 두지 않기 위해 이 세션에서 직접 재현했다.

```console
$ moai init <scratchpad>/t317verify --non-interactive
exit=0
$ find <scratchpad>/t317verify -path '*/.codex/*' -name '*.toml' | wc -l
      11
$ diff -rq <scratchpad>/t317verify/.codex/agents/moai internal/template/templates/.codex/agents/moai
Files <scratchpad>/t317verify/.codex/agents/moai/manager-develop.toml and
      internal/template/templates/.codex/agents/moai/manager-develop.toml differ
```

세 가지가 감사관 결과와 일치한다.

1. `.codex` 는 **기본 배포 대상**이다 — `--agent codex` 도 `--all` 도 주지 않았고 11건이 나왔다.
2. 배포 경로는 바이트를 보존한다 — 11건 중 10건 완전 일치.
3. 유일한 불일치 `manager-develop.toml` 이 **양성 검출**이다. 설치된 바이너리가 이 트리의 커밋본과 다른 `.codex` 바이트를 싣고 있음을, 소스↔소스 비교를 전혀 쓰지 않고 관측했다.

즉 REQ-AEL-004 가 겨눈 상태는 가정이 아니라 **지금 이 머신에 존재한다**. 폴백 2순위 삭제는 이제 두 독립 실행 위에 서 있다.

스크래치 디렉터리는 저장소 트리 밖(세션 scratchpad)에 만들었고 확인 후 제거했다. 워크트리 `git status` 는 이 카드의 미추적 디렉터리 2개 외에 변화가 없다.

---

## 측정 10 — `moai doctor --check` 개별 호출 실증

D1 결정(`moai doctor` 편입)의 근거 하나는 "항목이 `--check` 로 개별 호출된다"였는데, 저작자는 이를 플래그 존재와 기존 항목의 자기 문서화로만 세우고 **실행하지는 않았다**고 잔여 위험에 적었다. 이 세션에서 실행해 닫는다.

```console
$ moai doctor --check "MCP Server Version"
  ○ MCP Server Version
  ✓ MCP Server Version
    STATUS  CHECK               MESSAGE
    ok      MCP Server Version  no running moai MCP server recorded
    1 ok, 0 warn, 0 fail
   Pass 1    Warn 0    Fail 0
exit=0
```

`1 ok, 0 warn, 0 fail` 과 `Pass 1` 이 **필터가 실제로 한 항목만 남겼음**을 말한다(전체 실행이면 수십 건이 나온다). 따라서 D1 근거 2의 "자동 트리거와 스크립트 가능한 동사를 동시에 제공한다"는 주장은 이제 관측에 근거한다.

부수 관측: 이 실행의 판정이 `ok — no running moai MCP server recorded` 다. 즉 대조할 대상이 없을 때 이 선례 항목은 **OK 로 처리**한다. 신설할 임베드 검사는 D3 수리로 **정반대**를 요구한다 — 판정 대상 바이너리가 없으면 실패로 종료. 두 항목이 같은 `doctor` 안에서 부재를 다르게 다루는 것이므로, run-phase 가 항목을 등록할 때 이 비대칭을 의도된 것으로 명시하는 편이 좋다(임베드 검사는 부재를 "비교 0건 통과"로 흘리면 공허해지기 때문이다).

---

## 측정 11 — run-phase 결과 오케스트레이터 독립 재현

run-phase 가 AC 7/7 을 보고했다. 보고를 근거로 삼지 않고 이 카드의 핵심 주장 — "새 검사는 측정 8이 죽이지 못한 뮤턴트를 죽인다" — 를 직접 재현했다.

같은 뮤턴트를 심고 두 가지를 나란히 돌렸다.

```console
$ printf '\n# verify-t317-orchestrator\n' >> internal/template/templates/.codex/agents/moai/manager-git.toml

$ go test ./internal/template/agentemit/... -run TestEmbedFSPresenceAndByteEquality -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.459s      ← 생존 (여전히 공허)

$ make embed-check
    fail    Agent Emit Embed  moai embeds stale agent-emit artifacts (11/11 compared): manager-git.toml
    0 ok, 0 warn, 1 fail
   Pass 0    Warn 0    Fail 1
exit status 1                                                              ← 사망
```

**측정 8의 결론이 뒤집혔다.** 같은 뮤턴트에 대해 기존 임베드 테스트는 여전히 통과하고, 신설 검사는 실패한다. 그리고 `11/11 compared` 가 함께 나온다 — D3 이 요구한 기수 보고가 실제로 붙어 있어 부분 추출이 통과로 새지 않는다.

`make build` 도 컴파일에 닿기 전에 멈춘다:

```console
$ make build
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
    golden_test.go:109: .codex/agents/moai/manager-git.toml: committed artifact differs from emission (sha256 mismatch)
agent-emit drift: committed .codex/agents/moai/*.toml differ from the .md source layer — run `make agents-emit`
make: *** [agents-emit-check] Error 1
```

`go build` 행이 출력에 없다 — `agents-emit-check` 가 `build` 의 **첫 번째** 선행이라 컴파일 단계에 도달하지 않는다(`Makefile:23 build: agents-emit-check templ-generate`). 그리고 이 검사는 재생성하지 않는다: 실패 메시지가 `make agents-emit` 을 **부르라고 안내**할 뿐 스스로 고치지 않는다.

원복과 원복 확인:

```console
$ git checkout -- internal/template/templates/.codex/agents/moai/manager-git.toml
$ git status --short
(무출력 — 트리 청결)
$ make embed-check
   Pass 1    Warn 0    Fail 0
```

`.PHONY` 도 닫혔다 — `Makefile:16` 에 `agents-emit agents-emit-check embed-check` 세 이름이 모두 들어 있다.

### LSP 유령 중복 진단 (결함 아님)

이 검증 중 편집기 진단이 `internal/cli` 테스트에 중복 선언(`sameDir`, `snapshotTree`)을 보고했다. 실제 결함이 아니다.

```console
$ go vet ./internal/cli/...
(무출력 — 테스트 포함 컴파일 통과)
$ grep -rn "func sameDir\|func snapshotTree" internal/cli/
internal/cli/todo_queue_root_test.go:61:func sameDir(a, b string) bool {
internal/cli/clean_home_test.go:24:func snapshotTree(t *testing.T, root string) map[string]string {
$ grep -rn "func sameDir\|func snapshotTree" <primary>/internal/cli/
<primary>/internal/cli/todo_queue_root_test.go:61:func sameDir(a, b string) bool {
<primary>/internal/cli/clean_home_test.go:24:func snapshotTree(t *testing.T, root string) map[string]string {
```

각 트리에서 한 번씩만 선언돼 있다. 워크트리와 primary 가 같은 패키지 경로를 공유해 언어 서버가 둘을 겹쳐 색인하며 만든 유령이고, `go vet` 이 판정 권한을 갖는다. 워크트리 병행 작업에서 재발할 형태라 적어 둔다.
