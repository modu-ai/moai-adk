# SPEC-SEC-HARDEN-005 — Design

> design.md는 Tier M 표준에서 통상 생략되지만, **신규 의존성(`mvdan.cc/sh`)을 보안-임계 permission 게이트에 도입**하는 결정이므로 통합 설계를 명문화한다. plan-auditor + run-phase manager-develop이 정확히 같은 통합을 구현하도록 한다.

## D.1 §F.1 — `mvdan.cc/sh/v3/syntax` 통합 설계

### D.1.1 의존성 결정

- **추가**: `mvdan.cc/sh/v3` (현재 go.mod 부재 — `grep mvdan.cc/sh go.mod go.sum` → 0). pure-Go, POSIX/Bash/mksh 파서. High reputation(Context7 `/mvdan/sh`, 496 snippets). cross-platform 안전(NFR-SEC5-002).
- **거부된 대안 (SEC-HARDEN-002 §F.1 명시)**: `$`-blacklist 어휘 확장. `$HOME`/`${HOME}`/`TestX$`(9 정상 샘플 중 4)를 false-deny → invariant 위반. 출하 금지(REQ-SEC5-003).
- `go mod tidy` 후 `go.sum` 갱신. transitive 의존성 최소(파서만 import, interp/expand 미사용).

### D.1.2 파서 호출 (Parser invocation)

```
// 의사코드 — run-phase 구체화. hasUnquotedShellSeparator 인접 사적 헬퍼.
import "mvdan.cc/sh/v3/syntax"

func hasIFSWordSplit(remainder string) bool {
    parser := syntax.NewParser(syntax.Variant(syntax.LangBash))
    file, err := parser.Parse(strings.NewReader(remainder), "")
    if err != nil {
        return true // REQ-SEC5-004: parse 실패 → fail-closed(word-split 있다고 간주 → DENY)
    }
    // AST walk → IFS 기반 word-split 또는 다중 명령 탐지 (D.1.3)
    ...
}
```

- **파서 변형**: `syntax.Variant(syntax.LangBash)` — `${IFS}` 파라미터 확장은 Bash/POSIX 공통. Bash 변형이 가장 관대(over-parse가 아닌 정확 인식).
- **입력 범위**: 파서는 prefix remainder(`input[len(prefix):]`)만 받는다. 전체 input이 아닌 remainder를 파싱해야 prefix 매칭 후의 위험 부분만 평가한다. 단, remainder 단독은 leading 공백/구문상 불완전할 수 있으므로 run-phase에서 **전체 input을 파싱하고 첫 CallExpr의 인자 구조를 검사**하는 방식 vs **remainder 파싱** 중 선택 — 권장은 전체 input 파싱(구문적으로 완전, prefix는 첫 명령의 일부) 후 statement/word 구조 분석. (open question OQ-1 참조.)
- **인스턴스 생명주기**: `syntax.Parser`는 stateful·non-concurrent-safe. `Matches`는 동시 호출 가능(permission resolver hot-path)하므로 **per-call `NewParser`** 권장(가장 안전) 또는 `sync.Pool`. package-level 단일 인스턴스 공유는 race 위험 → 금지. (OQ-2.)

### D.1.3 AST walk 전략 (IFS word-split → DENY 매핑)

`mvdan.cc/sh/v3/syntax` AST 구조(WebFetch pkg.go.dev 확인):

- `*File.Stmts []*Stmt` — top-level 문장들. **2개 이상이면** 다중 명령 → DENY(이미 lexical이 `;`/newline 일부 잡지만 파서가 정확).
- `*Stmt.Cmd Command` — `*CallExpr`(단일 명령) | `*BinaryCmd`(`&&`/`||`/`|`/`|&`) | 기타. `*BinaryCmd`이면 command-chain → DENY.
- `*CallExpr.Args []*Word` — 단일 명령의 인자 워드들.
- `*Word.Parts []WordPart` — `*Lit` | `*DblQuoted` | `*SglQuoted` | `*ParamExp` | `*CmdSubst` | ...
- `*ParamExp.Param.Value` — 파라미터 이름. `Short bool`(`$a` vs `${a}`). quoted 여부는 부모 `*DblQuoted`/`*SglQuoted` 포함으로 판정.

**DENY 판정 규칙** (any-of → word-split 위험 → non-match):

1. `len(File.Stmts) > 1` → 다중 문장(DENY).
2. `Stmt.Cmd`가 `*BinaryCmd` → `&&`/`||`/`|` 체인(DENY). (lexical과 중복이지만 파서가 SSOT.)
3. `Stmt.Cmd`가 `*CallExpr`이고 그 인자 워드 중 **DblQuoted/SglQuoted 밖의 `*ParamExp`로 `IFS` 참조**가 존재 → IFS-driven word-split(DENY). `${IFS}`는 unquoted 컨텍스트에서 공백 확장 → 인접 토큰을 분리시킴.
4. (보강) `*CmdSubst`(`$(...)`)/`*ProcSubst` 존재 → DENY(lexical `$(`와 중복, 파서가 정밀).

**핵심 IFS 탐지**: `syntax.Walk(file, fn)`로 모든 노드 순회하되 `*SglQuoted`/`*DblQuoted` 하위 진입 시 quoted 플래그를 세워, **unquoted 컨텍스트의 `*ParamExp`에서 `Param.Value == "IFS"`** 를 찾으면 DENY. (Walk는 quoted 노드 하위도 방문하므로 quoted 컨텍스트 식별이 필수 — false-positive 방지: `echo "${IFS}"` 같은 quoted IFS는 word-split 안 일으킴.)

### D.1.4 ALLOW 보존 (no false-deny, REQ-SEC5-006)

- `go test -race ./...` → 단일 CallExpr, IFS 없음 → ALLOW.
- `go test $HOME/x` → `*ParamExp` Param.Value=="HOME"(≠IFS) → IFS 규칙 미발동 → ALLOW.
- `go test ${HOME}` → 동일, ALLOW.
- `go test TestX$` → trailing `$`는 리터럴(파서가 `*Lit` 또는 불완전 ParamExp 처리) → IFS 아님 → **ALLOW (RESOLVED, OQ-3)**. **결정**: trailing-`$`는 반드시 ALLOW. REQ-SEC5-004 fail-closed(parse-fail→DENY)는 **genuine malformed shell**에만 적용되며 trailing-literal-`$`는 malformed로 분류하지 않는다. `mvdan.cc/sh`가 lone trailing `$`를 parse-error로 반환하는 경우, generic fail-closed DENY **이전에** trailing-literal-`$` 케이스를 special-case ALLOW 처리한다(REQ-SEC5-006 우선). run-phase M1에서 SEC-HARDEN-002 9-sample 전부(특히 `go test TestX$`)를 ALLOW-유지 RED 케이스로 먼저 고정한다(AC-SEC5-005가 이 경계를 over-deny 회귀 케이스로 검증).
- `echo "${IFS}"`(quoted) → quoted 컨텍스트 IFS → word-split 없음 → ALLOW(quoted-IFS는 분리 안 됨).

### D.1.5 lexical guard와의 관계 (REQ-SEC5-005)

- 기존 `hasUnquotedShellSeparator`(stack.go:172) **유지**. 새 IFS 검사는 **추가 레이어**(OR 결합): `Matches`의 `:*` 브랜치는 `!hasUnquotedShellSeparator(rem) && !hasIFSWordSplit(rem)` 둘 다 통과해야 ALLOW.
- 실행 순서 권장: lexical 먼저(빠른 경로, separator 있으면 즉시 DENY) → 통과 시 파서 검사. lexical이 대부분의 명백 케이스를 cheap하게 거른 뒤 파서는 IFS류 미묘 케이스만 처리.
- 둘 다 fail-closed 방향(ambiguous → DENY)으로 일관.

## D.2 §F.2 — update env-trust allowlist 설계

### D.2.1 검증 지점 (deps.go EnsureUpdate)

`internal/cli/deps.go` `EnsureUpdate()` L260-285 env-read 블록 내부/인접:

```
// 의사코드 — run-phase 구체화.
updateSource := os.Getenv(config.EnvUpdateSource)

if updateSource == "local" {
    releasesDir := os.Getenv(config.EnvReleasesDir)
    if !isLocalPath(releasesDir) { // REQ-SEC5-009: URL-shaped 거부
        return fmt.Errorf("MOAI_RELEASES_DIR must be a local path, not a URL: %q", releasesDir)
    }
    ... // 기존 local checker 구성
}

apiURL := os.Getenv(config.EnvUpdateURL)
if apiURL != "" {
    if err := validateUpdateURL(apiURL); err != nil { // REQ-SEC5-007/008
        return fmt.Errorf("MOAI_UPDATE_URL rejected: %w", err)
    }
}
// apiURL == "" 인 default 경로는 githubReleasesURL/githubLatestReleaseURL 사용 → 검증 불필요
// (이들은 신뢰된 컴파일 상수, REQ-SEC5-010 no-regression)
```

### D.2.2 allowlist 정책

- **scheme allowlist**: `{"https"}`. `http`/`file`/`ftp` 등 거부(REQ-SEC5-008).
- **host allowlist**: `{"api.github.com"}` 최소셋. `githubReleasesURL`(deps.go:32)의 host와 일치. `const` 추출(NFR-SEC5-004) — `internal/cli` 사적 const 또는 `internal/config`.
- **검증 함수** `validateUpdateURL(raw string) error`: `url.Parse` → `u.Scheme != "https"` 거부 → `u.Hostname()`이 allowlist 미포함 거부. fail-closed.
- **local path 판정** `isLocalPath(s string) bool`: `url.Parse(s)`의 scheme이 비어있고(`u.Scheme == ""`) URL-shaped(`http`/`https` 등 prefix)가 아니면 로컬 경로로 간주. 보수적으로 scheme 존재 시 거부.

### D.2.3 no-regression 경계 (REQ-SEC5-010)

- env 미설정 → `apiURL == ""` → 컴파일 상수 경로(검증 skip) → 기존 동작 100% 보존.
- `MOAI_UPDATE_URL=https://api.github.com/...` (정상 override) → scheme=https + host=api.github.com → 통과.
- 검증은 **deps 초기화 경계까지만**. `internal/update` 다운로드/무결성 로직은 미변경(spec.md §F.3).

## D.3 §F.3 — TOCTOU OPTIONAL godoc (비요구)

- `restoreTargetContained`(update.go:2150)/`parentChainContained`(update.go:2216)/`runMXScan`(file_changed.go:130) godoc에 check-vs-use race 윈도 1-2줄 note(OPTIONAL).
- 문안 예: "Containment is checked at decision time; a concurrent adversarial process could in principle race the check against the subsequent write/read. This TOCTOU window is out of scope under the offline single-process threat model (`moai update` is a single process on the user's own machine), per SEC-HARDEN-003/004 §F.1 precedent."
- **코드 동작 변경 없음**. AC 게이트 아님(OPT-SEC5-001).

## D.4 Open Questions (run-phase 결정)

- **OQ-1**: §F.1 파서 입력 — remainder 단독 파싱 vs 전체 input 파싱 후 첫 명령 구조 분석. 권장: 전체 input 파싱(구문 완전성). RED 케이스로 둘 다 검증.
- **OQ-2**: 파서 인스턴스 — per-call `NewParser`(안전, 권장) vs `sync.Pool`(hot-path 성능). 동시성 안전이 우선; 성능 회귀 측정 후 결정.
- **OQ-3 (RESOLVED)**: `TestX$`/trailing-`$` 거동 — **결정: 반드시 ALLOW**. REQ-SEC5-004 fail-closed는 genuine malformed shell 한정이며, trailing-literal-`$`가 parse-error를 유발하면 generic DENY **이전에** special-case ALLOW로 처리한다(REQ-SEC5-006 우선, D.1.4 참조). AC-SEC5-005가 이 경계를 over-deny 회귀 케이스로 고정한다. (OQ-1/OQ-2는 순수 구현 세부로 run-phase 잔존.)
