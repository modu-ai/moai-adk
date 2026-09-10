# SPEC Review Report: SPEC-GO-TOOLCHAIN-SEC-002

카드: t610 · 트리: `.claude/worktrees/t610` (브랜치 `WT-go-1266`, HEAD `d3b7d438d`) · 감사일: 2026-09-10
Iteration: 1/1 (Tier S 상한 — `.moai/config/sections/harness.yaml:76` `S: 1`)
Verdict: **FAIL** (blocking 3건 — 모두 수정 범위가 작음)
Overall Score: 0.80 (4개 차원의 조화평균, Tier S PASS 임계 0.75)

작성자의 추론 맥락은 M1 Context Isolation에 따라 무시했다. 판정 근거는 SPEC 산출물 4개(spec.md·plan.md·acceptance.md·progress.md), `.moai/reports/t610/baseline/`, 그리고 트리를 직접 잰 결과뿐이다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `REQ-GTS2-001`~`008`이 spec.md:90-117에 빈 번호·중복 없이 순서대로 있다.
- [PASS] MP-2 GEARS 형식 (**요구사항 층 REQ-XXX 기준**): 001·007 Ubiquitous `shall`, 002·003 `shall not`, 004·005·006·008 `When … shall`. AC는 Given-When-Then 검증 층이므로 여기서 채점하지 않았다(Group 4에서 채점). 경미한 점 하나: REQ-005의 "may appear"(spec.md:106)가 규범 문장에 조동사로 들어가 있다(N15).
- [PASS] MP-3 frontmatter: 12개 필드가 모두 있고 타입도 맞다(spec.md:2-13, `version: "0.1.0"` 따옴표 포함, `phase: "v3.2.0 target"`은 금지된 단계명이 아님). `moai spec lint .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002` → `✓ No findings`, exit 0.
- [N/A] MP-4 언어 중립성: 이 저장소 자체의 Go 툴체인만 다루는 SPEC이고 템플릿이나 다언어 도구는 범위 밖이다.
- [PASS] MP-5 D7: 참조된 SPEC은 `SPEC-GO-TOOLCHAIN-SEC-001` 하나이고, 그 `status: completed`(retired/superseded/archived 아님) → BLOCKING 없음.
- [PASS] MP-6 D8: `grep -rn syscall` exit 1(매치 0) → 자동 PASS.
- [PASS] MP-7 clarification 게이트: `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002/` exit 1. research.md는 없음(Tier S).

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | 사실 번호 교차참조 오류(plan.md:70,71,87), REQ-008의 "listed in § A" 참조 대상 없음(spec.md:116), 게이트 범위 불일치(spec.md:123 vs acceptance.md:19). 모두 합리적인 엔지니어라면 같게 해석할 수 있는 수준 |
| Completeness | 1.0 | 1.0 | HISTORY(spec.md:20), WHY §B(:76), WHAT §A/§C, REQ §C(:88), AC §D + acceptance.md, `### Out of Scope — …` H3 4개에 모두 `-` 항목 있음(:146-171) |
| Testability | 0.75 | 0.75 | AC-003·AC-007은 develop 흡수 후 엉뚱한 이유로 red가 됨(B1), AC-007은 뮤턴트를 쓸 수 있음(B2), AC-005의 RED는 버전이 아니라 파일 부재 때문(N1) |
| Traceability | 0.75 | 0.75 | 모든 REQ에 AC가 연결되지만, REQ-007의 "unpiped exit code"는 AC 셀이 아니라 D.3(acceptance.md:154)에만 있어 간접 연결이다 |

## Findings

심각도 표기: blocking = 판정 재검토 전에 수정 필요 / non-blocking = 오케스트레이터 재량(M6).

| ID | 심각도 | 위치 | 결함 | 증거 | 수정안 |
|----|--------|------|------|------|--------|
| B1 | **blocking** (major) | acceptance.md:12, :16, :120-121 · plan.md:169-170 | AC-003(`git diff --numstat d3b7d438d -- go.mod`)과 AC-007(`git diff --name-only d3b7d438d -- …`)은 기준 커밋이 고정돼 있다. 그런데 이 SPEC이 인용하는 통합 절차(`gitflow-lane-protocol.md`, CLAUDE.local.md §4.1 규율 5)는 병합 전에 develop을 흡수하고 **병합 트리에서 다시 재라**고 요구한다. 흡수하는 순간 두 diff에 다른 레인의 변경이 섞여, 이 작업과 무관한 이유로 red가 된다. M4에도 재측정·재고정 단계가 없다. | `git rev-list --count --left-right origin/develop...d3b7d438d` → `0	28`. 기준 트리 자체가 원격보다 28커밋 앞서 있고 develop은 계속 움직인다. dependabot의 go.mod `require` 범프(`7374b183e`)가 흡수로 들어오면 AC-003의 `1	1`도 깨진다. | AC-003과 AC-007을 커밋 후 three-dot 형식으로 판정한다: `git diff --numstat develop...HEAD -- go.mod` (한 번의 호출로 끝나는 형식). M4에 "흡수 후 병합 트리에서 AC-001~007을 다시 재고, 병합 전 증거는 재사용하지 않는다"를 추가하고, RED 셀 고정 규칙(흡수 시 다시 고정)을 명시한다. |
| B2 | **blocking** (major) | acceptance.md:16 · spec.md:97-99 · spec.md:56-57 | REQ-003은 "any … Go source file"을 금지하고 §A.2는 release config를 버전 고정 지점으로 든다. 그런데 AC-007이 보는 경로는 `.github Makefile cmd internal pkg`뿐이다. 그 밖의 tracked Go 소스와 `.goreleaser.yml`을 고쳐도 AC-007은 green으로 남는다 — 뮤턴트를 쓸 수 있으므로 기준이 너무 얕다(verification-completeness §2). | `git ls-files '*.go' \| grep -v -E '^(cmd\|internal\|pkg)/'` → `scripts/convert-nextra-to-hextra/main.go`, `.moai/scripts/lint-skip-cleanup.go`, `.moai/scripts/status-drift-cleanup.go` 등. `.goreleaser.yml:12`도 경로 밖. | 트리 전체의 여집합으로 판정한다: `git diff --name-only develop...HEAD -- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610'` → 빈 출력. B1과 한 번에 고칠 수 있다. |
| B3 | **blocking** (major) | plan.md:44-46, :95-98, :105-107 · acceptance.md:15 | #80927("ReadHeaderTimeout remains active after unencrypted HTTP/2 handoff")이 "internal/web의 `http.Server`로 원리상 도달 가능"하다는 전제, 그리고 "internal/web 테스트가 그 경로의 run-phase 검사"라는 주장이 트리 증거와 맞지 않는다. 암호화하지 않은 HTTP/2를 켜는 설정이 저장소 어디에도 없다. AC 괄호 안의 서술이 존재하지 않는 커버리지를 주장하는 셈이다(VCI §1.1 surface 4: 권고의 전제). | `grep -rn -E 'ReadHeaderTimeout\|h2c\|UnencryptedHTTP2\|Protocols\|http2' internal/web` → `server.go:230 ReadHeaderTimeout` 하나. 저장소 전체에서 `UnencryptedHTTP2\|x/net/http2/h2c` 0건. `server.go:228-231`의 `http.Server{Handler, ReadHeaderTimeout}`에는 `Protocols` 필드가 없다. (Go의 unencrypted HTTP/2 기본값이 off라는 점은 이번 감사에서 재지 않았다 — Gaps 참조.) | D2와 AC-006의 문구를 고친다: "internal/web은 unencrypted HTTP/2를 설정하지 않는다(d3b7d438d 기준 grep 0건). 따라서 #80927은 현재 구성으로는 도달하지 않으며, internal/web 테스트는 net/http 일반 회귀 가드일 뿐 #80927을 검사하지 않는다." internal/web을 선택 목록에 두는 것 자체는 유지해도 된다. |
| N1 | non-blocking (major) | acceptance.md:67-71 (E-06) | release-blocking인 AC-005의 RED가 "바이너리 없음"에서 나왔다. 버전 문자열로 판별되는 쪽(go1.26.4 빌드 → `bin/moai: go1.26.4`)은 가정일 뿐 관측되지 않았다(verification-completeness §1.1 / §2의 "right-reason red"). | 이번 감사에서 대리 관측: `go version /Users/goos/go/bin/moai` → `/Users/goos/go/bin/moai: go1.26.4`. 현재 툴체인으로 빌드한 바이너리가 판별 문자열을 실제로 낸다는 것은 확인됐지만, 이 트리에서 빌드한 산출물은 아니다. | E-06에 버전 쪽 관측(위 명령과 출력, 출처 명시)을 붙이거나, 버전 쪽이 plan-phase에서 미관측임을 E-06에 명시한다. |
| N2 | non-blocking (minor) | acceptance.md:59-65 (E-05) | E-05의 gap("invocation string not persisted")은 이제 `commands.md`가 닫았다. 게다가 기준선 "auto" 셀은 GOTOOLCHAIN을 비운 것이 아니라 `GOTOOLCHAIN=auto`를 **명시**하고 돌렸다. 실행 주체도 E-05는 "lane lead", commands.md:4는 "lane-8"로 다르다. | commands.md:11 `GOTOOLCHAIN=$tc GOMAXPROCS=2 timeout 900 govulncheck -C $W ./...`. 이번 측정: `go env GOTOOLCHAIN` → `auto`, GOENV 파일에서 `grep GOTOOLCHAIN` exit 1(`go env -w` 없음). 이 머신에서는 unset과 auto가 동일하다. | E-05가 commands.md를 인용하게 하고, gap을 해소 처리하고, 판정용 명령 형식을 고정한다. 판정 증거 파일에 `go env GOTOOLCHAIN GOMOD GOVERSION` 출력을 함께 남기면 "no override"가 관측 가능해진다. |
| N3 | non-blocking (minor) | acceptance.md:10, :35 (AC-001/E-01) | `go version`의 결과는 cwd의 go.mod에 따라 달라지는데, 실행 디렉터리가 고정돼 있지 않다. primary 체크아웃에서 돌리면 go1.26.4, `/tmp`에서 돌리면 go1.26.0이 나온다. | commands.md:24 (`/private/tmp`에서 go1.26.0). `go env GOMOD` → 워크트리 go.mod 경로. | `go -C <worktree> version`을 쓰거나 `go env GOMOD`를 같은 증거에 기록한다. |
| N4 | non-blocking (minor) | plan.md:70, :71, :87 | "baseline fact 4/5/6" 교차참조가 spec.md §A 번호와 맞지 않는다. 자동 전환은 fact 3, 다운로드 로그는 번호가 없고, fact 6은 존재하지 않는다(§A는 1-5). | spec.md:51-74 | spec.md §A 번호 또는 증거 파일명(`goversion-go1.26.8.txt`, `govulncheck-*.log`)으로 바꿔 적는다. |
| N5 | non-blocking (minor) | spec.md:116-117 (REQ-008) | "listed in § A"라고 했지만 §A에는 문서 언급 목록도, `doc-1264-refs.txt` 인용도 없다. 목록은 plan.md:171-173과 E-08에 있다. | spec.md:26-74 | "listed in plan.md § F M5 / acceptance.md E-08"로 고친다. |
| N6 | non-blocking (minor) | spec.md:123-124 · spec.md:100-102 · acceptance.md:19-21 | 게이트 범위가 문서마다 다르다. spec §D는 "gates every other AC", REQ-004는 "before any other AC is judged"인데 acceptance의 Gating 절은 AC-004/005만 명시한다. | 인용한 줄들 | 한쪽으로 통일한다(AC-002~007 전부를 게이트 대상으로, 또는 004/005만). |
| N7 | non-blocking (minor) | commands.md:33 · plan.md:171-173 | 문서 스윕의 `--include=…`가 생략 표기로 남아 범위를 재현할 수 없다. tracked 트리를 직접 스윕하면 `internal/cli/testdata/tuxiu/postm4/update.{nocolor,notty,tty}.stdout.golden:2`에 `go1.26.4` 3건이 더 있고, SPEC은 이를 처분하지 않았다. 다만 이 golden들은 golden끼리만 비교하고, 해당 줄(◆ MoAI-ADK identity band)은 `tuxIsPresentation`이 걸러내므로(tuxiu_characterization_test.go:83-85, :139-140) 범프로 깨지지는 않는다. "5 mentions / 4 docs"는 정확하다. | `git grep -n -F '1.26.4' -- . ':!.moai/specs' ':!.moai/reports'` → go.mod:3, 문서 5줄, golden 3줄. untracked 0건. 4개 문서의 `1.26.8` 기준 개수는 각 0(AC-008의 target 쪽 RED도 성립). | Out of Scope에 한 줄 추가: "frozen capture golden 3건의 `go1.26.4`는 의도적으로 두며, identity band 줄은 비교에서 제외된다". |
| N8 | non-blocking (minor) | spec.md:60-62 · spec.md:74 | fact 3의 "`/opt/homebrew/bin/go` is go1.26.0" 근거로 `goversion-auto.txt`를 들었지만 그 파일에는 go1.26.4만 있다(1.26.0은 commands.md:24 주석에만 있음). fact 5는 "Go download index"를 근거로 들지만 그 출력은 저장되지 않았다. | `goversion-auto.txt` = `go version go1.26.4 darwin/arm64`. `go-release-history.html`에는 `id="go1.27.1"`이 있다(grep 2건) → fact 5는 이 파일로 대체 인용할 수 있다. | 인용을 commands.md:24 / `go-release-history.html`로 바꾼다. |
| N9 | non-blocking (minor) | plan.md:93-97 | D2가 #81113(debug/elf PPC)을 "플랫폼" 논리로 배제하는데, debug/elf는 호스트 플랫폼과 무관하게 ELF를 읽는 패키지다. 실제로 도달하지 않는 이유는 import가 없기 때문이다. 결론은 맞고 논거가 틀렸다. | `grep -rln '"debug/elf"' cmd internal pkg` exit 1. `.goreleaser.yml:13-19` goos = linux/darwin/windows (netbsd·openbsd 미출하 → #80827·#80889 배제는 타당). | "#81113: debug/elf를 import하지 않음(grep 0건)"으로 근거를 고친다. |
| N10 | non-blocking (optional) | spec.md:148-150 | Out of Scope에 든 "2 in packages you import"는 이번 범프 자체로 사라진다(표준 라이브러리). 범위 밖이라는 서술이 부정확하다. | `govulncheck-go1.26.6.log` / `-go1.26.8.log`: "0 vulnerabilities in packages you import and 3 … in modules you require" | "범프 후 0이 되며, 남는 것은 modules 3건"으로 고친다. |
| N11 | non-blocking (optional) | acceptance.md:13 (AC-004) | 한 번의 실행 안에서 패키지·모듈 그래프가 실제로 스캔됐는지 보여 주는 대조가 AC-004에 없다(툴체인 대조는 기준선에서만 했다). | 기준선 go1.26.6/8 로그 모두 "3 vulnerabilities in modules you require" 줄이 있다. | 선택 사항: 판정 출력에 "vulnerabilities in modules you require" 줄이 있어야 한다는 조건을 추가한다(advisory DB가 갱신되면 수치는 바뀔 수 있으므로 존재만 확인). |
| N12 | non-blocking (optional) | acceptance.md:14 · :154 | REQ-007의 "exit code recorded unpiped"가 AC 셀에 없고 D.3에만 있다. | acceptance.md:14 셀 문구 | AC-005 셀에 "each judged exit code in `<file>.exit`, unpiped"를 넣는다. |
| N13 | non-blocking (optional) | acceptance.md:73-77 · Makefile:34-36 | E-07(green)은 `make build` 전에 쟀다. build는 `templ-generate`(internal/web)와 `gen-catalog-hashes.go --all`을 실행하는데, 둘 중 하나가 tracked 산출물을 바꾸면 AC-007이 이 작업과 무관한 이유로 red가 된다. | Makefile:34 `build: agents-emit-check commands-emit-check templ-generate`, :35 `go run ./internal/template/scripts/gen-catalog-hashes.go --all` | AC-007은 M3의 `make build` 직후에 판정한다고 명시하고, 생성 파일에 diff가 생기면 어떻게 처분할지 적는다. |
| N14 | non-blocking (optional) | plan.md:130-133 | D3 대안을 기각한 근거("go-version-file does not both honour identically")가 미검증 주장이다. 선례 plan.md:80은 "setup-go … likewise honors the `toolchain`"이라고 쓴다. | `.moai/specs/SPEC-GO-TOOLCHAIN-SEC-001/plan.md:80` | 기각 근거를 선례 §E.2.1의 관측("updates to go.mod needed" 실패)으로 한정한다. |
| N15 | non-blocking (minor) | spec.md:105-106 | REQ-005의 "may appear"가 규범 문장 안에 있다(RQ-5). | spec.md:106 | "shall not appear"로 고친다. |

## 5-Section Evidence

### Claim
1. Must-pass 7개 중 6개는 PASS, MP-4는 N/A다.
2. B1: AC-003과 AC-007은 develop을 흡수한 트리에서 작업과 무관한 이유로 red가 되고, M4에는 재측정 단계가 없다.
3. B2: AC-007의 경로 범위가 REQ-003의 금지 범위보다 좁다.
4. B3: #80927에 도달한다는 전제와 "internal/web 테스트가 그 경로를 검사한다"는 서술을 트리 증거가 뒷받침하지 않는다.
5. D2의 1.26.6→1.26.8 차이 서술은 릴리스 기록 및 마일스톤 JSON과 일치하며, SPEC은 #80927의 영향을 **측정했다고 주장하지 않는다**(plan.md:105-107에 Gap으로 명시).
6. "5 mentions / 4 docs"는 정확하고, REQ·AC 사이에 개수 모순은 없다.
7. run-phase 제약(internal/cli 슬롯 승인, 로컬 전체 스위트 금지, 파일 출력 + unpiped exit, internal/web 포함)이 plan.md:42-53에 모두 있다. 범위 밖 목록(비호출 모듈 취약점 3건, go1.27, 이미 출하한 바이너리, SEC-02)도 spec.md:146-171에 있다. 카드 id t610은 spec.md:13 `card-t610`, :24, plan.md:3, progress.md:5에서 추적된다.

### Evidence (이번 실행에서 돌린 명령과 그대로의 출력)
- `moai spec lint .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002` → `✓ No findings — all SPEC documents are valid`, `lint exit=0`
- `mcp__moai__spec_audit(project_root=<worktree>, filter_spec=SPEC-GO-TOOLCHAIN-SEC-002)` → `{"total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[{"finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}]}`
- `git rev-list --count --left-right origin/develop...d3b7d438d` → `0	28`
- `git grep -n -F '1.26.4' -- . ':!.moai/specs' ':!.moai/reports'` → `.moai/project/codemaps/modules.md:6`, `.moai/project/codemaps/overview.md:6`, `.moai/project/product.md:244`, `.moai/project/product.md:300`, `.moai/project/structure.md:132`, `go.mod:3:go 1.26.4`, `internal/cli/testdata/tuxiu/postm4/update.{nocolor,notty,tty}.stdout.golden:2`
- `git grep -n -E 'go1\.26|1\.26\.[0-9]' -- .github Makefile .goreleaser.yml scripts` → 출력 없음, `exit=1`
- `git ls-files '*.go' | /usr/bin/grep -v -E '^(cmd|internal|pkg)/'` → `scripts/convert-nextra-to-hextra/main.go`, `.moai/scripts/lint-skip-cleanup.go`, `.moai/scripts/status-drift-cleanup.go`, `.moai/reports/…` (발췌)
- `/usr/bin/grep -rn -E 'ReadHeaderTimeout|h2c|UnencryptedHTTP2|Protocols|http2' internal/web --include='*.go'` → `internal/web/server.go:230:		ReadHeaderTimeout: 10 * time.Second,` (1건)
- 저장소 전체 `UnencryptedHTTP2|golang.org/x/net/http2/h2c` → 0건. `"debug/elf"` in cmd/internal/pkg → `elf exit=1`
- release history 추출: `go1.26.7 (released 2026-08-19) includes fixes to the net/http package.` / `go1.26.8 (released 2026-09-01) includes fixes to cgo, the compiler, the runtime, and the debug/elf and os packages.` / `go1.26.6 (released 2026-08-13) includes security fixes to the go command, and the crypto/tls, encoding/asn1, encoding/xml, html/template, net, net/http, and net/url packages, …`
- `go1.26.7-milestone-issues.json` → `[{"number":80927,…}]`. `go1.26.8-milestone-issues.json` → 81152, 81113, 80889, 80851, 80827 (D2 서술과 번호·주제 일치)
- `go env GOTOOLCHAIN GOENV GOMOD` → `auto` / `/Users/goos/Library/Application Support/go/env` / `…/worktrees/t610/go.mod`. GOENV 파일 `grep GOTOOLCHAIN` → `exit=1`
- `go version /Users/goos/go/bin/moai` → `/Users/goos/go/bin/moai: go1.26.4`
- `/usr/bin/grep -c -F '1.26.8'` 4개 문서 → 각 `0`
- `.moai/specs/SPEC-GO-TOOLCHAIN-SEC-001/spec.md:5` → `status: completed`. 같은 SPEC `progress.md:59` `## §E.2.1 Run-phase Decision — toolchain directive dropped`, `:67` `go build ./... / govulncheck ./... FAIL with go: updates to go.mod needed`
- `.moai/config/sections/harness.yaml:76` → `S: 1   # Tier S …: single-pass audit, no iteration 2+`

### Baseline-attribution
전부 이번 실행에서, 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610`(HEAD `d3b7d438d`, `git status --short` = `?? .moai/reports/t610/`, `?? .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002/`)를 대상으로 쟀다. 기준선 파일(`govulncheck-*.log`, `goversion-*.txt`, `*-refs.txt`, JSON, HTML)은 lane-8이 같은 트리에서 만든 산출물이며, 이번 감사에서는 내용을 읽기만 했고 다시 실행하지 않았다(지시에 따라 govulncheck·go test·make build 금지).

### Gaps (이번 감사에서 관측하지 않은 것)
- govulncheck, `make build`, `go test`, `go vet`은 실행하지 않았다(지시에 따른 제약). AC-004/005/006의 green 경로는 미관측이다.
- Go의 `http.Server` unencrypted HTTP/2 기본값이 off라는 점은 이번 실행에서 재지 않았다. B3의 트리 측 근거(설정 0건)만 관측했다.
- #80927의 실제 upstream diff는 읽지 않았다(SPEC의 gap과 같음).
- 교차 모델 감사(`audit_multi` / codex / glm)는 호출하지 않았다. `.moai/config/sections/`에서 `audit_model`을 grep한 결과 0건이라 설정 요청이 없었다.
- `make build`의 생성 단계(templ, catalog hashes)가 tracked 파일을 바꾸는지는 관측하지 않았다(N13).
- `go-version-file`이 `toolchain` 지시자를 어떻게 다루는지는 확인하지 않았다(N14).

### Residual-risk
- advisory DB는 실시간으로 갱신되므로, 범프 후 판정 시점에 새 affecting 항목이 나올 수 있다(EC-2가 대응).
- develop이 병합 창 전에 더 움직이면 B1을 고친 뒤에도 흡수 충돌·재측정 비용이 남는다. D4의 단독 병합 창은 CI 귀속만 보호한다.
- internal/cli 테스트는 로컬 선택에서 빠지므로, 툴체인에 민감한 CLI 회귀는 `origin/develop` CI에서야 드러난다.
- go.mod가 `go 1.26.8`이 되면, `GOTOOLCHAIN=local`이면서 구버전 Go를 쓰는 환경은 즉시 실패한다(Scenario 3, 의도된 잔여 위험).

## Recommendation

1. **B1**: acceptance.md:12와 :16의 명령을 `develop...HEAD` three-dot 형식으로 바꾸고(커밋 후 판정), plan.md:169-170의 M4에 "흡수 후 병합 트리에서 AC-001~007 재측정, 병합 전 증거 재사용 금지, RED 셀 재고정"을 추가한다.
2. **B2**: AC-007을 트리 전체 여집합 형식(`-- . ':!go.mod' ':!<SPEC dir>' ':!.moai/reports/t610'`)으로 바꾸고, E-07을 같은 형식으로 다시 잰다.
3. **B3**: plan.md:44-46, :95-98, :105-107과 acceptance.md:15에서 "#80927 reachable in principle"과 "internal/web tests are the run-phase check for that path"를 삭제하고, "internal/web은 unencrypted HTTP/2 미설정(grep 0건) → #80927 미도달, internal/web 테스트는 일반 회귀 가드"로 바꾼다.
4. non-blocking 중에서는 N1(AC-005의 버전 쪽 RED 관측)과 N2(E-05 gap 해소, 판정 명령 고정, `go env GOTOOLCHAIN` 기록)를 같이 고치기를 권한다. 나머지는 재량이다.

반복 계약: Tier S 상한은 1회(`harness.yaml:76`)이므로 이 감사자가 iteration 2를 수행할 수 없다. 오케스트레이터는 (a) blocking 3건을 수정한 뒤 명시적 override로 결함 차이만 범위로 한 재감사를 요청하거나, (b) PASS-with-debt로 기록할지를 사용자 게이트에서 정해야 한다. B1~B3는 모두 문구·명령 몇 줄 수정이라 SPEC 범위를 줄일 필요는 없다.
