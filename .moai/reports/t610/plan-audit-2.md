# SPEC Review Report: SPEC-GO-TOOLCHAIN-SEC-002

카드: t610 · 트리: `.claude/worktrees/t610` (브랜치 `WT-go-1266`, HEAD `d3b7d438d`) · 감사일: 2026-09-10
Iteration: 2 (결함 범위 한정 재감사. Tier S 상한은 1회(`harness.yaml:76`)이며, 이번 회차는 레인 오케스트레이터가 명시적으로 override한 것이다)
대상 개정: spec.md `version: "0.1.1"`
Verdict: **PASS**
Blocking: **0**
Overall Score: 0.86 (4개 차원 조화평균, Tier S 임계 0.75, iteration 1은 0.80 → 하락 없음, STOP 신호 없음)

작성자 추론 맥락은 M1 Context Isolation에 따라 무시했다. 개정본이 주장하는 사실은 모두 가설로 두고 다시 쟀다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호: `REQ-GTS2-001`~`008`이 spec.md:98-126에 빈 번호·중복 없이 순서대로 있다.
- [PASS] MP-2 GEARS (**요구사항 층 REQ-XXX 기준**): 001·007은 Ubiquitous `shall`, 002·003은 `shall not`, 004·005·006·008은 `When … shall`. iteration 1의 N15(REQ-005의 "may")는 spec.md:114에서 "None … **shall** appear"로 바뀌었다. 조동사 약어 grep(`may|should|appropriate|reasonable|adequate`) 매치 0. AC는 Given-When-Then 검증 층이라 여기서 채점하지 않았다.
- [PASS] MP-3 frontmatter: 필수 12개 필드가 모두 있고(+`tier: S`, 13줄 매치) `version: "0.1.1"`은 따옴표로 감쌌다. `moai spec lint` → `✓ No findings`, exit 0.
- [N/A] MP-4 언어 중립성: 이 저장소 자체의 Go 툴체인만 다루고 템플릿에는 닿지 않는다.
- [PASS] MP-5 D7: 참조 SPEC은 `SPEC-GO-TOOLCHAIN-SEC-001` 하나이고 `status: completed`다 → BLOCKING 없음.
- [PASS] MP-6 D8: SPEC 디렉터리 `syscall` grep exit 1 → 자동 PASS.
- [PASS] MP-7 clarification 게이트: `[NEEDS CLARIFICATION` grep exit 1. research.md는 없다(Tier S).

## Category Scores

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | 0.75 | 사실 번호 교차참조가 이제 맞는다(plan.md:80 "fact 3", :97 "fact 4" ↔ spec.md:64,69). 남은 모호함: E-03 exit 필드 오기(R1), D.0b의 재사용 조건이 새 실행 순서에서 도달 불가(R2), D.0b.4의 커버리지 과대 서술(R4) |
| Completeness | 1.0 | 1.0 | HISTORY spec.md:20-25, WHY §B, REQ §C, AC §D + acceptance.md, `### Out of Scope — …` H3 5개(spec.md:160-195), frozen golden 처분 추가(:178-185) |
| Testability | 0.75 | 0.75 | diff AC가 merge-base 형식과 트리 전체 여집합으로 바뀌어 뮤턴트가 제거됐다(acceptance.md:186-187). AC-008이 흡수 뒤 재측정 대상에서 빠진 점(R3) 하나가 남는다 |
| Traceability | 1.0 | 1.0 | REQ-007이 AC-004/005 셀에 직접 연결됐다(acceptance.md:17-18). 8개 REQ 모두 AC가 있고 고아 AC는 없다 |

## 결함별 해소 현황 (iteration 1 → 2)

| ID | 판정 | 근거 (이번 실행에서 다시 잰 것) |
|----|------|------|
| B1 | **해소** | AC-003 `git diff --numstat develop...HEAD -- go.mod`(acceptance.md:16), AC-007도 같은 형식(:20, E-07 :149). 근거는 §D.0a(:33-58), M4 흡수 뒤 재측정은 §D.0b(:60-77)와 plan.md:215-224. 실측: `git rev-parse --verify develop` → `d1b61005d2…`, `git merge-base develop HEAD` → `d3b7d438d2…`, `git merge-base --is-ancestor d3b7d438d develop` 오류 없음. 3-dot 전체 트리 diff는 **빈 출력**, 2-dot `git diff --shortstat develop HEAD`는 `4 files changed, 63 insertions(+), 268 deletions(-)`(`.claude/settings.json`, `.moai/config/sections/tool-policy.yaml`, `.moai/reports/t609/verdict.md`, `…/settings.json.tmpl`). `git rev-list --count HEAD..develop` → `4`. 3-dot 형식이 develop 이동을 흡수하지 않는다는 점이 현재 트리에서 성립한다 |
| B2 | **해소** | AC-007이 `-- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610'` 여집합으로 판정한다(acceptance.md:149). E-07의 "29 paths" 주장 재측정: `git ls-files '*.go' ':!cmd/**' ':!internal/**' ':!pkg/**'` → 29줄(`scripts/i18n-validator/*`, `test/integration/harness/*`, `.moai/scripts/*` 포함). `.goreleaser.yml`도 여집합 안에 있다. 뮤턴트 기록은 :186-187 |
| B3 | **해소** | "#80927 도달 가능 / internal/web 테스트가 그 경로를 검사" 서술을 철회하고 "일반 net/http 회귀 가드"로 바꿨다(plan.md:49-55, :115-125; acceptance.md:19). 실측: `git grep -n -E 'UnencryptedHTTP2\|h2c\|x/net/http2' -- '*.go' go.mod` → 빈 출력. 같은 `'*.go'` pathspec 양성 대조 `git grep -n 'ReadHeaderTimeout' -- '*.go'` → `internal/web/server.go:230` 1건(도구가 작동함). 저장소 전체 `-w 'Protocols\|SetUnencryptedHTTP2\|TLSNextProto'` → 빈 출력. Go 기본값은 재지 않았다는 점을 Gap으로 명시했다(plan.md:132-135). 단, plan.md:119의 "exit 1"은 이 하네스가 non-zero exit를 표시하지 않아 독립 확인이 안 됐다(Gaps 참조). stdout은 확인했다 |
| N1 | 해소 | E-06에 버전 쪽 대리 관측과 출처가 들어갔다(acceptance.md:139-146). 재측정 `go version /Users/goos/go/bin/moai` → `/Users/goos/go/bin/moai: go1.26.4` 일치 |
| N2 | 해소 | E-05가 commands.md:8-16을 인용하고 gap 해소를 표시했으며(:125-132), `GOTOOLCHAIN=auto` 명시와 unset이 같다는 점도 적었다. 판정 명령이 `go -C … env GOTOOLCHAIN GOMOD`를 함께 기록하도록 AC-004 셀(:17)과 plan.md:63-66에 고정됐다 |
| N3 | 해소 | AC-001이 `go -C <worktree-root> version`(:14). E-01 재측정: `go -C …/t610 version` → `go version go1.26.4 darwin/arm64`, `env GOTOOLCHAIN GOMOD GOVERSION` → `auto` / `…/t610/go.mod` / `go1.26.4`. 원장(:88-96)과 일치 |
| N4 | 해소 | plan.md:80 "fact 3"(로컬 툴체인 = spec.md:64), :97 "fact 4"(govulncheck 표 = spec.md:69) |
| N5 | 해소 | REQ-008이 "listed in plan.md § F M5 and acceptance.md E-08"(spec.md:124-126) |
| N6 | 해소 | 게이트 대상이 spec.md:132-133, REQ-004 :108-110, acceptance.md:23-26, plan.md:182-184에서 모두 "AC-GTS2-002..007"로 통일됐다 |
| N7 | 해소 (잔여 있음) | Out of Scope에 frozen golden 3건 처분이 들어갔다(spec.md:178-185). `git grep -n -F '1.26.4' -- . ':!.moai/specs' ':!.moai/reports'` 재측정 → 문서 5줄 + `go.mod:3` + golden 3줄로 서술과 일치. `tuxiu_characterization_test.go:83-85` 인용도 일치한다. 잔여: 기준선 기록 commands.md:33의 `--include=…` 생략 표기는 그대로다(증거 파일 쪽 문제이며 SPEC 판정에는 영향 없음) |
| N8 | 해소 | fact 3은 commands.md:22-24를, fact 5는 `go-release-history.html`을 인용한다(spec.md:66-68, :80-82). 재측정: html에 `id="go1.26.6"`, `id="go1.26.7"`, `id="go1.26.8"`, `id="go1.27.1"`가 각 1건 |
| N9 | 해소 | plan.md:108-110 "import가 없음"으로 근거를 바꿨다. `git grep -n '"debug/elf"' -- '*.go'` → 빈 출력(양성 대조는 B3와 같은 pathspec 형태). `.goreleaser.yml:12` `CGO_ENABLED=0`, :13-16 `goos: linux/darwin/windows` 재확인 |
| N10 | 해소 | spec.md:162-165 |
| N11 | 해소 | AC-004 셀에 "modules you require" 존재 대조가 들어갔다(:17). 빈 스캔 뮤턴트도 기록됐다(:176-178) |
| N12 | 해소 | AC-004/005 셀에 "exit code in `<file>.exit` captured unpiped"(:17-18) |
| N13 | 해소 | M3가 `make build` 전후 `git status --porcelain --untracked-files=no`를 기록하고, 재생성된 tracked 파일은 커밋하지 않는다(plan.md:208-214, acceptance.md:55-58). "tracked `*_templ.go` 8개" 재측정: `git ls-files 'internal/web/*_templ.go'` → 8줄. Makefile:34-36 `build: … templ-generate` + `gen-catalog-hashes.go --all` 확인 |
| N14 | 해소 (Gap으로 기록) | plan.md:163-165에 Gap으로 남겼고, D3 기각 근거(:158-161)는 더 이상 그 미검증 사실에 기대지 않는다 |
| N15 | 해소 | spec.md:114 |

## 개정 결정 3건에 대한 판정

**(a) 실행 순서 M1→M2→M3→M5→M4 — 채택 가능.** 워크트리 트리의 CLAUDE.local.md:393 "sync는 병합 **전에** 워크트리 안에서 끝낸다"와 일치하고, plan.md:199-200, :215, :225, acceptance.md:263-272(Run/Sync/Integration 게이트 분리)가 서로 모순 없이 같은 순서를 말한다. 다만 이 순서 변경이 기존 문장 두 곳을 낡게 만들었다(R2, R5).

**(b) M4에서만 AC-007 제외 목록에 sync 경로 5개 추가 — 기능은 하지만 서술이 과대하다.** 제외 목록 자체는 정확하다(acceptance.md:76. REQ-008 대상 4개 문서 + CHANGELOG.md). 그 밖의 경로를 건드리는 sync 커밋은 M4 여집합에서 여전히 드러난다. 그러나 D.0b.4의 "AC-GTS2-008 judges those paths separately"는 맞지 않는다. AC-008은 4개 문서의 버전 토큰 **줄 수**만 세고, CHANGELOG.md 내용은 어떤 AC도 판정하지 않는다. 예를 들어 sync 커밋이 product.md:244의 LOC 통계를 지우거나 CHANGELOG의 이전 항목을 삭제해도 AC-007(M4)과 AC-008이 모두 green으로 남는다. 이는 REQ 위반이 아니므로(REQ-003은 CI·Makefile·Go 소스만 금지한다) §2의 REQ 뮤턴트는 아니다. 하지만 plan.md:46-48의 scope guard 중 이 5개 경로 부분은 SPEC 안에서 강제되지 않는다 → R4 (non-blocking).

**(c) RED 셀은 base에 고정하고 흡수 뒤에는 green만 tree id와 함께 재측정 — 타당하다. develop...HEAD 형식은 현재 트리에서 올바르게 평가된다.** 근거:
- 현재 상태: 로컬 develop `d1b61005d`는 HEAD보다 4커밋 앞이고 merge-base는 `d3b7d438d`다. 3-dot diff는 빈 출력, 2-dot은 4개 파일. develop이 base 이후 go.mod·sync 경로 5개·`.goreleaser.yml`을 건드리지 않았다(`git diff --numstat develop HEAD -- go.mod CHANGELOG.md <4 docs> .goreleaser.yml` → 빈 출력). `go 1.26.4`는 `develop`, `origin/develop`, `origin/main` 세 곳 모두 그대로다(`git grep -n '^go ' <ref> -- go.mod`). 다른 레인이 이미 범프한 중복 작업은 없다.
- 흡수 뒤: `git merge develop` 뒤 merge-base는 흡수한 develop tip이 되고, 그 뒤 develop이 전진해도 새 tip이 흡수 tip을 조상으로 가지므로 merge-base는 그대로다. 이 부분은 git merge-base 정의에 따른 추론이며 시뮬레이션은 하지 않았다(Gaps).
- RED를 다시 고정하지 않아도 되는 이유: develop 쪽 변경이 AC-004를 대신 green으로 만드는 경우(예: 다른 레인의 directive 범프)는 AC-003의 `1	1	go.mod`와 AC-002가 판별한다. 그 경우 카드 쪽 diff가 `1 1`이 아니거나 병합 충돌이 난다.
- 잔여 위험: primary 체크아웃의 **작업 사본** CLAUDE.local.md:357은 흡수 대상을 `git merge origin/develop`으로 적고 있다(워크트리 tracked 사본 :389는 로컬 develop). 레인이 그 사본을 따르면 D.0a 조건 1이 깨진다. 현재 로컬 develop이 origin/develop보다 32커밋 앞서므로(`git rev-list --count --left-right origin/develop...develop` → `0	32`) merge-base는 origin/develop tip이 되어 여전히 카드 쪽만 보인다. origin이 앞서는 경우에는 AC-007이 red가 되어 **드러나는 방향**으로 실패한다(조용한 green이 아니다) → R7 (optional).

## Defects Found (신규)

R1. E-03 exit 필드가 재현되지 않음 — acceptance.md:108-110 — 원장은 `grep -c '^toolchain' go.mod`의 exit를 `0`으로 적고 "shell grep은 ugrep 래퍼라 0, GNU grep이면 1"이라고 설명한다. 같은 명령을 `$?`로 재면 셸 함수 grep과 `/usr/bin/grep` 모두 **exit=1**이다. 이 하네스의 Bash 표시는 이 명령의 non-zero exit를 드러내지 않는다(단독 실행 시 `0`만 보이고 exit 표시 없음). 표시를 exit 0으로 읽은 것으로 보인다. 판정은 stdout 기준이라 verdict는 뒤집히지 않지만, RED 셀 4요소 중 exit 코드가 관측과 다르다(verification-completeness.md §2.1) — Severity: minor — Class: optional(권장) — Required fix: E-03 exit를 `1`로 고치고 괄호 설명을 삭제한다. 같은 표시 문제가 plan.md:108·:119와 progress.md:15의 "exit 1" 기록에도 있을 수 있으니, run-phase는 REQ-007 형식(`<cmd> > f 2>&1; echo $? > f.exit`)으로만 exit를 남긴다고 명시한다.

R2. 새 순서에서 D.0b 재사용 조건이 도달 불가 — acceptance.md:68-70, plan.md:221-222 — "absorbed `HEAD^{tree}` equals the tree M3 recorded, which happens when absorption brought nothing new". M5 sync 커밋이 M3 뒤, M4 앞에서 tracked 파일 6개 이상을 바꾸므로 흡수가 아무것도 가져오지 않아도 트리는 M3 트리와 같아질 수 없다. 늘 재측정하는 쪽이라 안전하게 실패하지만, 문장이 말하는 동치는 거짓이다 — Severity: minor — Class: optional — Required fix: 비교 기준을 "M5 sync 커밋 직후 기록한 tree id"로 바꾸거나 재사용 조항을 삭제한다.

R3. AC-008이 M4 재측정에서 빠짐 — acceptance.md:65, :243-244, plan.md:194 — M4 재측정 범위는 AC-001..007뿐이다. 그런데 M4의 AC-007 제외 목록은 4개 문서를 통째로 허용하고, `git merge develop`은 그 문서들에 다른 레인의 변경(새 `1.26.4` 언급, 충돌 해결 hunk)을 들여올 수 있다. 흡수 뒤 REQ-008 성립을 확인하는 절차가 없다 — Severity: minor — Class: optional(권장) — Required fix: D.0b.1, D.3, plan.md § E.7에 AC-008을 추가한다.

R4. D.0b.4의 "AC-GTS2-008 judges those paths separately" 과대 서술 — acceptance.md:71-73 — AC-008은 4개 문서의 버전 토큰 줄 수만 판정하고, CHANGELOG.md는 어떤 AC도 읽지 않는다. 5개 경로의 비의도 변경은 SPEC 게이트를 통과한다(판정 (b) 참조) — Severity: minor — Class: optional — Required fix: 문장을 "AC-008은 4개 문서의 버전 토큰을 판정하고, 그 밖의 내용과 CHANGELOG.md는 sync-auditor가 판정한다"로 고친다. 또는 M5에서 흡수 전에 `git diff --numstat <M3 커밋> <sync 커밋> -- <4 docs> CHANGELOG.md`를 판정한다(product.md `2	2`, 나머지 3개 `1	1`, CHANGELOG.md 삭제 열 `0`).

R5. `completed` 전이 후 M4 실패 처리가 비어 있음 — acceptance.md:234-236(EC-5), :267-272 — 새 순서에서 status는 sync 커밋에서 `completed`가 된 뒤 M4 재측정과 CI 판정을 맞는다. EC-5는 blocker 반환만 규정하고 이미 `completed`인 SPEC의 처분(amendment 전이 등)을 말하지 않는다. 레인 규칙과는 양립한다 — Severity: minor — Class: optional — Required fix: EC-5에 "progress.md에 두 tree id 기록, status 정정이 필요하면 manager-spec의 `completed → in-progress (amendment)` 경로"를 한 줄 추가한다.

R6. AC-006 셀의 CI 시점 표기 — acceptance.md:19 — "M3, re-measured at M4 → … CI on the `origin/develop` push green". 그 CI는 M4 병합 뒤 리드의 push에서야 돈다. D.4:260에 따로 적혀 있어 판정에는 영향이 없다 — Severity: minor — Class: optional — Required fix: CI 항목을 "리드 push 후" 시점으로 분리 표기한다.

R7. CLAUDE.local.md 줄 인용이 트리에 따라 다른 내용을 가리킴 — plan.md:218, :254, acceptance.md:48, :72 — `:389`/`:393`은 카드 워크트리의 tracked 사본에서만 해당 규칙이다. primary 체크아웃 작업 사본에서는 :357/:361이고, :357은 `git merge origin/develop`이라 D.0a 조건 1과 반대로 읽힌다 — Severity: minor — Class: optional — Required fix: 줄 번호 대신 규칙 문구("흡수 대상은 **로컬** `develop`")를 인용하거나 트리(`d3b7d438d`)를 명시한다.

(Blocking 결함 없음. iteration 1의 B1–B3는 모두 해소됐다.)

## 5-Section Evidence

### Claim
1. B1·B2·B3 모두 해소됐고, N1–N15 중 14건은 해소, N14는 Gap으로 기록돼 해소 처리했다(N7은 증거 파일 쪽 잔여 1건).
2. Must-pass 6개 PASS, MP-4 N/A.
3. `develop...HEAD` 형식은 로컬 develop이 `d1b61005d`로 전진한 현재 트리에서 올바르게 평가된다(3-dot 빈 출력, 2-dot 4개 파일).
4. 개정이 새로 만든 결함은 모두 non-blocking이다(R1–R7). 이 중 R2·R3·R5는 순서 변경 (a)에서, R4는 결정 (b)에서 나왔다.
5. 판정: PASS (Tier S 임계 0.75 대비 0.86).

### Evidence (이번 실행의 명령과 출력 그대로)
- `moai spec lint .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002` → `✓ No findings — all SPEC documents are valid`, `lint exit=0`
- `mcp__moai__spec_audit(project_root=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610, filter_spec=SPEC-GO-TOOLCHAIN-SEC-002)` → `{"total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[{"spec_id":"SPEC-GO-TOOLCHAIN-SEC-002","era":"V3R6","finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}]}`
- `git rev-parse --verify develop` → `d1b61005d20967fdbd970ec7ec734c6d14f29dc3`
- `git merge-base develop HEAD` → `d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1`
- `git merge-base --is-ancestor d3b7d438d develop` → 오류 출력 없음
- `git diff --name-only develop...HEAD` → 빈 출력
- `git diff --shortstat develop HEAD` → ` 4 files changed, 63 insertions(+), 268 deletions(-)`
- `git diff --name-only develop HEAD` → `.claude/settings.json`, `.moai/config/sections/tool-policy.yaml`, `.moai/reports/t609/verdict.md`, `internal/template/templates/.claude/settings.json.tmpl`
- `git rev-list --count HEAD..develop` → `4`
- `git diff --numstat develop HEAD -- go.mod CHANGELOG.md .moai/project/product.md .moai/project/structure.md .moai/project/codemaps/overview.md .moai/project/codemaps/modules.md .goreleaser.yml` → 빈 출력
- `git diff --numstat develop...HEAD -- go.mod` → 빈 출력
- `git grep -n '^go ' develop -- go.mod` → `develop:go.mod:3:go 1.26.4`. `origin/develop` → `origin/develop:go.mod:3:go 1.26.4`. `origin/main` → `origin/main:go.mod:3:go 1.26.4`
- `git rev-list --count --left-right origin/develop...develop` → `0	32`
- `git grep -n -E 'UnencryptedHTTP2|h2c|x/net/http2' -- '*.go' go.mod` → 빈 출력
- `git grep -n 'ReadHeaderTimeout' -- '*.go'` → `internal/web/server.go:230:		ReadHeaderTimeout: 10 * time.Second,`
- `git grep -n -E 'ReadHeaderTimeout|Protocols' -- internal/web` → 같은 1줄
- `git grep -n -w -E 'Protocols|SetUnencryptedHTTP2|TLSNextProto' -- '*.go'` → 빈 출력
- `git grep -n '"debug/elf"' -- '*.go'` → 빈 출력. `git grep -l '"debug/' -- '*.go'` → 빈 출력
- `git grep -n zzqq_absent_token_qqzz -- go.mod` → 빈 출력, 하네스에 exit 표시 없음(no-match exit 표시 방식 대조)
- `grep -c '^toolchain' go.mod; printf 'grep exit=%s\n' $?` → `0` / `grep exit=1`. `/usr/bin/grep -c '^toolchain' go.mod; … $?` → `0` / `bsd grep exit=1`. `type grep` → `grep is a shell function from …/shell-snapshots/snapshot-zsh-….sh`. `type git` → `git is /usr/bin/git`
- `git ls-files '*.go' ':!cmd/**' ':!internal/**' ':!pkg/**'` → 29줄
- `git ls-files 'internal/web/*_templ.go'` → 8줄
- `git grep -n -F '1.26.4' -- . ':!.moai/specs' ':!.moai/reports'` → `.moai/project/codemaps/modules.md:6`, `codemaps/overview.md:6`, `product.md:244`, `product.md:300`, `structure.md:132`, `go.mod:3`, `internal/cli/testdata/tuxiu/postm4/update.{nocolor,notty,tty}.stdout.golden:2`
- 4개 문서 `/usr/bin/grep -c '1\.26\.4'` / `'1\.26\.8'` → product.md `2`/`0`, structure.md `1`/`0`, overview.md `1`/`0`, modules.md `1`/`0`
- `go -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610 version` → `go version go1.26.4 darwin/arm64`
- `go -C … env GOTOOLCHAIN GOMOD GOVERSION` → `auto` / `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610/go.mod` / `go1.26.4`
- `go version /Users/goos/go/bin/moai` → `/Users/goos/go/bin/moai: go1.26.4`
- `sed -n '10,20p' .goreleaser.yml` → :12 `- CGO_ENABLED=0`, :13 `goos:`, :14-16 `linux`/`darwin`/`windows`
- `sed -n '80,86p' internal/cli/tuxiu_characterization_test.go` → :83 `if strings.Contains(l, "◆") && strings.Contains(l, "MoAI-ADK") { // NEW identity band`
- CLAUDE.local.md(워크트리 tracked 사본) `grep -n` → :389 "`git merge develop` 흡수(대상은 **로컬** `develop` …)", :393 "sync는 병합 **전에** 워크트리 안에서 끝낸다". primary 작업 사본 → :357 "`git merge origin/develop` 흡수", :361 같은 sync 문구
- 기준선 `go1.26.7-milestone-issues.json` → `80927 … ReadHeaderTimeout remains active after unencrypted HTTP/2 handoff`. `go1.26.8-…json` → 81152, 81113, 80889, 80851, 80827
- `/usr/bin/grep -rn 'audit_model' .moai/config/sections/` → `exit=1`

### Baseline-attribution
모두 이번 실행에서 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610`(HEAD `d3b7d438d`, `git status --short` = `?? .moai/reports/t610/`, `?? .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002/`)를 대상으로 쟀다. ref 값(`develop` = `d1b61005d`, `origin/develop`, `origin/main`)은 이 실행 시점의 로컬 ref를 읽은 것이며 `git fetch`는 하지 않았다. 기준선 파일(`.moai/reports/t610/baseline/*`)은 읽기만 했다.

### Gaps (이번 감사에서 관측하지 않은 것)
- 이 하네스의 Bash 도구는 `git grep`·`grep`의 no-match non-zero exit를 표시하지 않고, worktree 가드는 `git` 복합 명령(`; echo $?`)을 거부한다. 그래서 plan.md:108·:119와 progress.md:15의 "exit 1"은 **stdout만 확인했고 exit는 독립 확인하지 못했다**. grep 계열 exit는 셸 `$?`로 잰 R1의 1건뿐이다.
- `git merge develop` 흡수 뒤의 merge-base 동작은 시뮬레이션하지 않았다(객체/커밋 생성이 read-only 지시에 어긋남). 판정 (c)의 흡수 뒤 부분은 git merge-base 정의에 기댄 추론이다.
- govulncheck, `make build`, `go test`, `go vet`은 지시에 따라 실행하지 않았다. AC-004/005/006의 green 경로는 관측하지 않았다.
- Go `http.Server`의 unencrypted HTTP/2 기본값과 #80927 upstream diff는 읽지 않았다(SPEC의 Gap과 같음).
- 교차 모델 감사(`audit_multi`/codex/glm)는 호출하지 않았다. `.moai/config/sections/`에 `audit_model` 설정이 없다(grep exit 1).
- N9 음성 결과의 양성 대조는 같은 `'*.go'` pathspec 형태(`ReadHeaderTimeout`)로만 했고, 따옴표가 들어간 import 문자열 형태로는 따로 대조하지 않았다.

### Residual-risk
- advisory DB가 실시간으로 갱신되므로 run-phase 시점에 새 affecting 항목이 나올 수 있다(EC-2가 대응).
- 레인이 primary 작업 사본의 CLAUDE.local.md를 따라 `origin/develop`을 흡수하면 D.0a 조건 1이 깨진다. 실패 방향은 AC-007 red라서 드러난다(R7).
- sync 커밋이 5개 경로에서 버전 토큰 외의 내용을 바꿔도 SPEC 게이트는 잡지 못한다(R4). 그 몫은 sync-auditor에 남는다.
- `completed`가 된 뒤 M4 재측정이 실패하면 SPEC 상태와 실제가 어긋난 채 blocker가 올라간다(R5).
- internal/cli 테스트는 로컬 선택에서 빠지므로 툴체인에 민감한 CLI 회귀는 `origin/develop` CI에서야 드러난다.

## Recommendation

PASS 근거(must-pass별):
- MP-1: spec.md:98-126, 001~008 연속.
- MP-2: 요구사항 층 8개 모두 GEARS 형식(spec.md:98, :101, :105, :108, :112, :116, :120, :124).
- MP-3: 12개 필드 + `tier`, lint exit 0.
- MP-5/6/7: SEC-001 `completed`, `syscall` 0건, clarification 마커 0건.
- blocking 3건 해소는 위 표의 재측정 명령과 출력으로 확인했다.

run 진입 전에 고치기를 권하는 non-blocking 항목(오케스트레이터 재량, M6):
1. **R3**: D.0b 재측정 목록에 AC-008을 추가한다. 흡수가 4개 문서를 건드릴 수 있고, M4 제외 목록이 그 경로를 통째로 허용하기 때문이다.
2. **R1**: E-03 exit를 `1`로 정정하고, exit는 REQ-007 파일 형식으로만 기록한다고 명시한다.
3. **R2·R4**: D.0b.3 비교 기준 tree를 "M5 후 tree"로 바꾸고, D.0b.4의 "AC-008 judges those paths separately" 문구를 실제 범위로 줄인다.
4. R5·R6·R7은 재량이다.

반복 계약: 이번 회차는 Tier S 상한(1회)을 넘는 명시적 override 재감사다. PASS이므로 추가 회차는 필요 없다. Implementation Kickoff Approval 사람 게이트는 이 PASS와 무관하게 필수다.
