# A2 — ast-grep Integration Audit

## Executive Summary

- **LOC audited**: 6,261 (production: 2,043 / test: 4,218)
- **ast-grep version pinned**: 없음 (installed: 0.40.5, no minimum declared)
- **Overall health**: 6/10
- **Top 5 issues**:
  1. `Rule` 구조체에 `Metadata`/`Note` 필드 부재 — YAML 규칙의 `metadata:` 섹션(OWASP/CWE)이 파싱 후 소실됨
  2. `analyzer.go` 루프 내 `context.WithTimeout` + `cancel()` 이 `defer` 없이 호출되어 에러 반환 경로에서 context 누수 잔존
  3. `detectSGVersion()` 미구현 — SARIF `tool.driver.version`이 항상 `"unknown"`
  4. `security/ast_grep.go` ScanMultiple 고루틴 폭주 — 10,000 파일 입력 시 동시 10,000 고루틴 생성 (세마포어 없음)
  5. 16개 언어 중 11개 언어 규칙 디렉토리가 `.gitkeep`만 존재 — 언어 중립성 선언이 사실상 미이행
- **Verdict**: REFACTOR

---

## File Inventory

| File | LOC | Purpose |
|------|-----|---------|
| `internal/astgrep/analyzer.go` | 559 | SGAnalyzer: pattern scan/replace, file walk |
| `internal/astgrep/models.go` | 79 | Match, ScanResult, Rule 등 공유 타입 |
| `internal/astgrep/rules.go` | 208 | RuleLoader: YAML 파일 로드, 멀티-doc 지원 |
| `internal/astgrep/sarif.go` | 214 | Finding → SARIF 2.1.0 변환 |
| `internal/astgrep/scanner.go` | 411 | Scanner (통합): sgconfig.yml / per-rule scan |
| `internal/astgrep/analyzer_test.go` | 1,211 | SGAnalyzer 단위 테스트 |
| `internal/astgrep/rules_test.go` | 239 | RuleLoader 단위 테스트 |
| `internal/astgrep/sarif_test.go` | 316 | SARIF 변환 단위 테스트 |
| `internal/astgrep/scanner_test.go` | 611 | Scanner 단위 테스트 |
| `internal/cli/astgrep.go` | 238 | `moai ast-grep` Cobra CLI 커맨드 |
| `internal/cli/astgrep_test.go` | 182 | CLI 통합 테스트 |
| `internal/cli/astgrep_filter_test.go` | 58 | filterByLang 단위 테스트 |
| `internal/cli/astgrep_output_test.go` | 204 | 출력 포매터 단위 테스트 |
| `internal/hook/quality/astgrep_gate.go` | 254 | V1/V2 quality gate hook |
| `internal/hook/quality/astgrep_gate_test.go` | 270 | V1 gate 단위 테스트 |
| `internal/hook/quality/astgrep_gate_v2_test.go` | 357 | V2 gate 단위 테스트 |
| `internal/hook/security/ast_grep.go` | 378 | PostToolUse 보안 스캐너 |
| `internal/hook/security/ast_grep_test.go` | 472 | 보안 스캐너 단위 테스트 |
| `.moai/config/astgrep-rules/sgconfig.yml` | 24 | sgconfig 루트 |
| `.moai/config/astgrep-rules/go/*.yml` | ~100 | Go 전용 규칙 5개 파일 |
| `.moai/config/astgrep-rules/security/*.yml` | ~90 | 보안 규칙 4개 파일 |

---

## D1 — CLI Integration

### 1-1. Subprocess 호출

`scanner.go`와 `analyzer.go` 두 곳에서 각각 독립적으로 `os/exec` subprocess를 호출한다.

**`scanner.go`** (Scanner 통합 구현): `exec.CommandContext`와 명시적 context timeout 사용. `scanWithConfig`는 `cmd.Stdout`/`cmd.Stderr`를 분리하고 sg의 non-zero exit을 `_ = cmd.Run()`으로 명시적으로 무시한다(`scanner.go:284`). `runSingleRule`은 `defer cancel()`로 context 누수를 방지하며 `@MX:WARN` 태그가 이유를 문서화한다(`scanner.go:302`).

**`analyzer.go`** (SGAnalyzer): `CommandExecutor` 인터페이스를 통해 추상화되어 테스트 이중(double)이 가능하다. 그러나 루프 안에서 `context.WithTimeout` 후 `cancel()`을 `defer` 없이 직접 호출한다(`analyzer.go:242, 298`). 이 패턴은 `return nil, fmt.Errorf(...)` 경로에서 `cancel`이 호출되지 않아 context 누수가 발생한다.

```go
// analyzer.go:242 — ISSUE: cancel() before error return path
ctx, cancel := context.WithTimeout(ctx, SGTimeout)
args := ...
output, err := a.executor.Execute(ctx, a.workDir, "sg", args...)
cancel()         // ← 에러 return 전 cancel이지만, Execute가 패닉하면 누수
if err != nil {
    return nil, fmt.Errorf(...)
}
```

권장 수정: `cancel := func(){}; defer cancel()` 패턴으로 교체하거나 `runSingleRule`처럼 `defer cancel()`을 사용한다.

### 1-2. 출력 파싱

sg의 JSON 출력 파싱은 `parseSGFindings`(scanner.go)와 `parseSGOutput`(analyzer.go) 두 함수가 중복 구현되어 있다. 두 함수는 동일한 `[]sgMatch` 구조를 파싱하지만 이름이 다르고 패키지 분리로 인해 공유되지 않는다. 향후 sg가 출력 형식을 변경하면 두 곳을 모두 수정해야 한다.

sg JSON 형식에서 0-indexed line을 1-indexed로 변환하는 `+1` 로직이 `scanner.go:401`과 `analyzer.go:122`에 중복된다.

`security/ast_grep.go`는 세 번째 독립 파서(`parseASTGrepJSON`, `parseASTGrepRegex`)를 구현한다. 특히 `range` 객체 형식과 flat 형식을 모두 지원하는 방어적 파싱이 포함되어 있어 robustness는 높으나 코드 중복이 심화된다.

### 1-3. 에러 처리 (binary missing / wrong version)

**Binary missing**: `Scanner.Scan`은 `isSGAvailable()` 체크 후 warn+skip (`scanner.go:232-236`). `SGAnalyzer.IsSGAvailable`은 결과를 mutex로 캐시한다 (`analyzer.go:135-149`). 두 구현 모두 graceful degradation을 지원한다.

**Version verification**: 최소 버전 선언 및 런타임 검증이 없다. `detectSGVersion()` (`cli/astgrep.go:193`)은 `// 미래 구현을 위한 플레이스홀더`라는 주석과 함께 항상 `"unknown"`을 반환한다. `security/ast_grep.go:GetVersion`은 버전을 캐시하지만 최소 버전 제약을 검증하지 않는다.

**버전 관련 리스크**: ast-grep 0.31+ 이전에는 `sg scan --json` 출력 형식이 달랐다. 사용 환경(0.40.5 설치됨)은 현재 정상이지만, CI/CD나 다른 개발자 머신에서 구버전 sg가 사용될 경우 파싱 실패가 조용히 발생한다.

---

## D2 — Pattern/Rule Management

### 2-1. 규칙 저장 및 로딩

규칙은 `.moai/config/astgrep-rules/`에 언어별 서브디렉토리 구조로 저장된다. sgconfig.yml이 존재하면 config 기반 스캔(sg native rule engine), 없으면 `RuleLoader.LoadFromDir`로 재귀 로딩 후 `sg run --pattern` 방식으로 개별 실행한다.

`RuleLoader.LoadFromDir`는 `filepath.WalkDir`로 `.yml`/`.yaml` 파일을 재귀 탐색하며, `loadFileSkipOnError`로 파싱 실패를 격리한다(`rules.go:109-139`). 멀티-doc YAML을 `---` 구분자로 분리하여 개별 파싱하는 방식은 yaml.v3 디코더 상태 복구 불가 문제를 잘 해결했다(`rules.go:186-208`).

### 2-2. Rule 구조체의 Metadata/Note 필드 부재 (중요 버그)

YAML 규칙 파일에는 `metadata:` 섹션(OWASP, CWE 참조)과 `note:` 필드가 사용된다:

```yaml
# security/injection.yml:8
metadata:
  owasp: "A03:2021 - Injection"
  cwe: "CWE-89"
note: "사용자 입력을 SQL 쿼리에 직접 삽입하지 마세요..."
```

그러나 `models.go`의 `Rule` 구조체(`models.go:72-79`)에는 `Metadata map[string]string`과 `Note string` 필드가 없다. `scanner.go`의 `Finding` 구조체에는 있지만(`scanner.go:65-67`), `scanWithRules` 경로에서 Rule로부터 Finding으로의 메타데이터 전파 코드가 없다(`scanner.go:361-374`). 결과적으로:

- `metadata.owasp`, `metadata.cwe`, `note` 값이 SARIF `properties`와 Finding에 전달되지 않는다
- 5개 보안 규칙 파일의 OWASP/CWE 분류 정보가 사용자에게 노출되지 않는다
- SARIF 출력의 `ruleProperties`가 항상 비어있다

### 2-3. 언어 감지

`analyzer.go:DetectLanguage`는 `foundation.DefaultRegistry.ByExtension`을 통해 언어를 감지하며(`analyzer.go:155-169`), 16개 언어 레지스트리에 연동된다. `security/ast_grep.go`는 독립적인 `extensionToLanguage` 맵을 빌드하여 15개 언어를 지원한다(`ast_grep.go:350-358`). 두 구현이 동기화되지 않으면 지원 언어 목록이 분기될 수 있다.

### 2-4. 사용자 정의 규칙

`--rules-dir` 플래그를 통해 임의 경로 지정이 가능하다(`cli/astgrep.go:60`). ValidateBinary가 SGBinary를 검증하지만 RulesDir에 대한 경로 검증은 없다. 악의적 YAML이 있을 경우 YAML 파서는 노출 위험이 있으나 ast-grep 자체가 규칙 실행을 샌드박스한다.

---

## D3 — Security Hook Integration

### 3-1. 탐지 범위

`security/` 규칙 디렉토리의 4개 파일이 커버하는 취약점:

| 규칙 파일 | 탐지 패턴 | OWASP |
|-----------|----------|-------|
| `injection.yml` | SQL Injection (fmt.Sprintf), Command Injection (exec.Command sh -c), Path Traversal | A03:2021 |
| `secrets.yml` | 하드코딩 API Key (sk- 접두사), JWT Signing Key | A07:2021 |
| `crypto.yml` | MD5 사용, InsecureSkipVerify | A02:2021 |
| `web.yml` | XSS (template.HTML), Log Injection (log.Printf), CSRF (POST handler without token) | A01/A03/A09 |

모두 Go 언어 전용이며, 범용 보안 규칙(다른 언어용)은 부재한다.

### 3-2. False-positive 리스크

**높음**:
- `go-error-not-wrapped` (`error-handling.yml`): `return $ERR` 패턴은 정상적인 sentinel error 반환(`return nil`, `return io.EOF`)까지 매칭할 수 있다.
- `go-channel-send-no-select` (`concurrency.yml`): `$CHAN <- $VALUE`는 모든 채널 전송을 탐지하므로 의도적인 blocking send에도 경고가 발생한다.
- `sec-csrf-no-token-check` (`web.yml`): GET 핸들러도 `func $HANDLER(w http.ResponseWriter, r *http.Request)` 패턴과 일치한다.
- `sec-log-injection-unsanitized`: `log.Printf($FORMAT, $$$ARGS)`는 fmt.Sprintf를 래핑한 구조적으로 유사한 모든 호출을 탐지한다.

**측정 불가**: 현재 false-positive rate를 측정하는 메커니즘이 없다.

### 3-3. 위반 처리

`security/ast_grep.go`의 `Scan` 메서드는 결과를 반환만 하고 직접 차단하지 않는다. 차단 로직은 hook 호출자(`internal/hook/security/`) 쪽에서 결정한다. 보안 스캐너 자체는 찾기(detection)만 수행한다.

---

## D4 — Quality Gate Integration

### 4-1. V1/V2 구분

`astgrep_gate.go`에 두 함수가 공존한다:

- **`RunAstGrepGate` (V1)**: 직접 `exec.LookPath("sg")`, `astgrep.RuleLoader.LoadFromDirectory`, `runSGConfig`/`runSGRule` 호출. 자체 subprocess 관리 로직 내장.
- **`RunAstGrepGateV2` (V2)**: `astgrep.NewScanner` 통합 구현에 위임. 코드가 훨씬 단순하고 `scanner.go`의 검증/캐싱/타임아웃 로직을 재사용한다.

V2는 V1의 개선판으로 설계되었으며 `TestRunAstGrepGate_V1_V2_Equivalence` 테스트가 두 구현의 행동 동등성을 검증한다(`astgrep_gate_v2_test.go:297`). V1은 마이그레이션 완료 후 제거 예정이나 현재 두 함수가 동시 존재한다.

**V1만 사용하는 기능**: `LoadFromDirectory`(non-recursive 로딩) — V1에서는 `loader.LoadFromDirectory(rulesDir)`로 flat 로딩한다(`astgrep_gate.go:125`). V2에서는 `Scanner.Scan`을 통해 `LoadFromDir`(recursive)를 사용한다. 서브디렉토리 규칙이 있을 때 V1과 V2의 규칙 로딩 범위가 다르다.

### 4-2. 게이트 트리거

`PreToolUse` 또는 `PostToolUse` hook에서 호출되는 것으로 추정되나, 이 감사 범위 내 파일에서는 hook 연결 코드가 보이지 않는다(`internal/hook/quality/` 디렉토리 외부).

### 4-3. Bypass 메커니즘

`WarnOnlyMode: true`로 설정하면 error severity 발견이 있어도 차단하지 않는다(`astgrep_gate.go:186`). `BlockOnError: false`도 동일 효과다. 두 플래그가 모두 설정될 경우 논리가 중복되나 행동은 동일하다.

### 4-4. 리포팅 형식

gate 실패 시 출력: `"quality gate failed: ast-grep domain rules\n\n{파일:줄: [규칙ID] 메시지 (심각도)}"` 형식의 문자열. JSON이나 구조화된 형식은 아니며 파싱이 어렵다.

---

## D5 — Performance

### 5-1. Subprocess 비용

`scanWithRules` 경로는 규칙 수만큼 sg subprocess를 spawn한다. 현재 Go 규칙 5개라면 5회 spawn. sgconfig.yml 경로는 단일 spawn이므로 현재 config의 sgconfig.yml 사용은 올바른 선택이다.

### 5-2. 캐싱

**바이너리 가용성**: `SGAnalyzer.IsSGAvailable`은 mutex로 결과를 캐시한다(`analyzer.go:136-149`). `security/astGrepScanner.IsAvailable`도 동일(`ast_grep.go:37-55`). `Scanner`(scanner.go)는 캐시 없이 매번 `exec.LookPath`를 호출한다.

**파일 해시 캐싱**: 없음. 동일 파일을 여러 번 스캔해도 결과를 재사용하지 않는다. 대규모 리포지토리에서 반복 스캔 비용이 선형적으로 증가한다.

**규칙 로딩**: `Scanner.Scan` 호출마다 `RuleLoader.LoadFromDir`을 새로 실행하여 디스크에서 YAML을 재파싱한다(`scanner.go:254-257`). sgconfig.yml 경로에서는 이 비용이 없다.

### 5-3. ScanMultiple 고루틴 폭주

`security/ast_grep.go:ScanMultiple`은 파일 수만큼 고루틴을 무제한 생성한다(`ast_grep.go:186-213`):

```go
// ast_grep.go:186 — ISSUE: no goroutine limit
for i, fp := range filePaths {
    wg.Add(1)
    go func(idx int, path string) {  // ← 10,000 files = 10,000 goroutines
        defer wg.Done()
        result, err := s.Scan(ctx, path, configPath)
        ...
    }(i, fp)
}
```

`@MX:WARN` 태그가 이를 명시하지만(`ast_grep.go:172-173`) 실제 세마포어나 worker pool이 없다. 10,000 파일 입력 시 sg subprocess 10,000개가 동시에 spawn되어 OS 리소스 고갈 가능성이 있다.

### 5-4. 대규모 리포지토리

`SGAnalyzer.ScanProject`는 `filepath.Walk`로 전체 디렉토리를 탐색하며 sg를 파일별로 순차 실행한다(`analyzer.go:393-434`). 10,000개 파일에서 파일당 50ms면 총 8분이 소요된다. timeout이 없고 context 전파만 있어 timeout 제어가 상위에 의존한다.

---

## D6 — Testing Quality

### 6-1. 실제 sg 바이너리 호출 여부

**`internal/astgrep/`**: 테스트가 `exec_test.go` 패키지로 `package astgrep_test`로 작성되어 외부 패키지 테스트다. `SGAnalyzer` 테스트는 `mockExecutor`로 실제 subprocess를 모킹한다(`analyzer_test.go:14-49`). `Scanner` 테스트는 실제 sg 없이 환경(빈 rules dir)으로 graceful degradation을 검증한다.

**`internal/hook/security/`**: sg가 없으면 `t.Skip()` 패턴을 사용하여 CI에서 sg가 없을 때 건너뛴다(`ast_grep_test.go:69`). 실제 sg 호출 테스트는 `if !scanner.IsAvailable() { t.Skip(...) }` 가드가 있다.

### 6-2. 모킹 전략

`SGAnalyzer`는 `CommandExecutor` 인터페이스를 통해 완전한 모킹이 가능하다(`analyzer.go:38-39`). `WithCommandExecutor` 옵션으로 테스트 시 주입한다(`analyzer.go:65-69`). `Scanner`는 인터페이스 추상화 없이 직접 `exec.CommandContext`를 호출하므로 단위 테스트에서 실제 sg가 없으면 graceful degradation path만 검증 가능하다.

### 6-3. 에러 경로 커버리지

커버된 에러 경로:
- sg 바이너리 없음: 검증됨
- rules dir 없음: 검증됨
- 잘못된 YAML: 검증됨(`scanner_test.go:183`)
- context 취소: 검증됨(`scanner_test.go:231`)
- JSON 파싱 실패: 검증됨(`astgrep_gate_test.go:203`)

미커버 에러 경로:
- sg가 malformed JSON을 반환하는 경우 (sg 버전 불일치 시나리오)
- timeout 만료 후 부분적 결과 처리
- 매우 큰 JSON 출력 (메모리 압박 시나리오)

### 6-4. Golden file 테스트

SARIF 출력 형식 안정성 검증을 위한 golden file 테스트가 없다. `sarif_test.go`는 구조적 필드를 검증하지만 전체 JSON 스냅샷을 보존하지 않는다. sg 출력 형식 변경 시 감지가 어렵다.

---

## D7 — Language Neutrality

### 7-1. 16개 언어 지원 현황

`sgconfig.yml`에는 16개 언어(+ security, utils) 디렉토리가 선언되어 있으나:

| 언어 | 상태 |
|------|------|
| go | 실 규칙 5개 파일 |
| security (cross-lang) | 실 규칙 4개 파일 |
| cpp, csharp, elixir, flutter, java, javascript, kotlin, php, python, r, ruby, rust, scala, swift, typescript | `.gitkeep` 만 존재 |

15개 언어가 규칙 없이 선언만 되어 있다. sgconfig.yml이 이 빈 디렉토리를 참조하지만 실제 규칙이 없으면 해당 언어의 품질/보안 게이트는 동작하지 않는다.

### 7-2. Security 훅의 언어 지원

`internal/hook/security/ast_grep.go`의 `supportedLanguages`는 15개 언어를 지원하지만 Flutter/Dart(`.dart`)와 R(`.r`)가 없다(`ast_grep.go:331-347`). 16개 언어 지원 약속 대비 2개 누락.

### 7-3. 언어 감지 불일치

`analyzer.go:DetectLanguage`는 `foundation.DefaultRegistry`를 통해 감지하고, `security/ast_grep.go`는 독립 `extensionToLanguage` 맵을 사용한다. 두 레지스트리가 다른 확장자 세트를 지원하면 동일 파일에 대해 다른 언어가 감지될 수 있다.

### 7-4. 우아한 저하

언어별 규칙이 없으면 해당 언어 파일에 대해 빈 결과가 반환된다. 경고나 "규칙 없음" 표시가 없어 사용자는 스캔이 실패했는지 규칙이 없는지 구분할 수 없다.

---

## Detailed Recommendations

### REC-01 (Priority High): Rule 구조체에 Metadata/Note 추가

`models.go`의 `Rule` 구조체에 누락된 필드를 추가하고, `scanWithRules`에서 Finding으로 전파한다:

```go
// models.go — Rule 구조체
type Rule struct {
    ID       string            `json:"id" yaml:"id"`
    Language string            `json:"language" yaml:"language"`
    Severity string            `json:"severity" yaml:"severity"`
    Message  string            `json:"message" yaml:"message"`
    Pattern  string            `json:"pattern" yaml:"pattern"`
    Fix      string            `json:"fix,omitempty" yaml:"fix,omitempty"`
    Note     string            `json:"note,omitempty" yaml:"note,omitempty"`     // 추가
    Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"` // 추가
}
```

`scanner.go:scanWithRules`의 메타데이터 주입 루프에 전파 코드 추가:
```go
findings[i].Note = rule.Note
if rule.Metadata != nil && findings[i].Metadata == nil {
    findings[i].Metadata = rule.Metadata
}
```

### REC-02 (Priority High): ScanMultiple에 Semaphore 추가

`security/ast_grep.go:ScanMultiple`에 최대 동시 goroutine 수를 제한하는 세마포어를 추가한다:

```go
const maxConcurrentScans = 16

sem := make(chan struct{}, maxConcurrentScans)
for i, fp := range filePaths {
    wg.Add(1)
    go func(idx int, path string) {
        sem <- struct{}{}
        defer func() { <-sem }()
        defer wg.Done()
        ...
    }(i, fp)
}
```

### REC-03 (Priority High): detectSGVersion 구현

`cli/astgrep.go:detectSGVersion`을 실제 구현한다:

```go
func detectSGVersion() string {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, "sg", "--version")
    out, err := cmd.Output()
    if err != nil {
        return "unknown"
    }
    return strings.TrimSpace(string(out))
}
```

`Scanner` 구조체에도 버전 캐싱 메서드를 추가하여 SARIF 출력의 `tool.driver.version` 필드를 채운다.

### REC-04 (Priority Medium): analyzer.go 루프 내 context 누수 수정

`analyzer.go:242`와 `analyzer.go:298`의 루프 내 `ctx, cancel :=` 패턴을 `scanner.go:runSingleRule`처럼 `defer cancel()`로 교체한다:

```go
// Before:
ctx, cancel := context.WithTimeout(ctx, SGTimeout)
output, err := a.executor.Execute(ctx, ...)
cancel()

// After: 별도 함수로 추출하거나
func (a *SGAnalyzer) executeWithTimeout(ctx context.Context, timeout time.Duration, ...) {
    scanCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    ...
}
```

### REC-05 (Priority Medium): JSON 파싱 함수 통합

`scanner.go:parseSGFindings`, `analyzer.go:parseSGOutput`, `security/ast_grep.go:parseASTGrepJSON` 세 함수를 하나로 통합한다. `internal/astgrep` 패키지에 공개 함수로 정의하고 다른 패키지가 임포트하도록 한다. 0-indexed to 1-indexed 변환도 단일 위치에서 수행된다.

### REC-06 (Priority Medium): V1 게이트 제거 일정 수립

`RunAstGrepGate` (V1)과 `RunAstGrepGateV2` (V2)가 동시 존재한다. V1/V2 동등성 테스트가 통과하는 지금 V1 제거가 안전하다. 다만 V1의 `LoadFromDirectory`(non-recursive)와 V2의 `LoadFromDir`(recursive) 간 의도적 차이가 있다면 V2가 V1을 완전히 대체하는지 확인 후 제거한다.

### REC-07 (Priority Medium): 최소 버전 선언 및 런타임 검증

`DefaultScannerConfig` 또는 `Scanner` 초기화 시 sg 최소 버전(예: 0.25.0)을 선언하고, 첫 스캔 시 `sg --version` 출력과 비교하여 경고를 발생시킨다. 현재 시스템 sg 0.40.5는 충분하지만 다른 환경 보증이 없다.

### REC-08 (Priority Low): SARIF URI 형식 수정

`sarif.go:toFileURI`는 절대 경로에 `file://` 스킴을 붙이지 않는다. SARIF 2.1.0 스펙에서 `artifactLocation.uri`는 URI여야 하므로, 절대 경로 입력 시 `"file://" + path`로 변환하는 것이 스펙에 부합한다:

```go
func toFileURI(path string) string {
    if path == "" || strings.HasPrefix(path, "file://") {
        return path
    }
    if filepath.IsAbs(path) {
        return "file://" + filepath.ToSlash(path)
    }
    return path
}
```

---

## Open Questions

1. **V1 게이트 사용처**: `RunAstGrepGate`(V1)를 직접 호출하는 hook handler가 있는가? 이 감사 범위 외 파일에서 확인 필요.
2. **Rule.Metadata 파싱 테스트**: YAML metadata 섹션 파싱이 현재 동작하는지(필드 없어서 silently 무시) rules_test.go에서 검증된 바 없음. 의도된 미구현인가?
3. **security/ 규칙 Go 전용 이유**: security 규칙이 Go만 있는 것은 현 프로젝트가 Go로 작성되어 자기 점검 목적인가, 아니면 다른 언어 보안 규칙 추가 계획이 있는가?
4. **ScanProject 사용처**: `SGAnalyzer.ScanProject`는 파일별 sg 호출로 매우 느릴 수 있다. 실제 호출 경로가 있는가? sgconfig.yml 기반 단일 sweep이 더 효율적이다.

---

## References (file:line)

- `internal/astgrep/models.go:72` — Rule 구조체 (Metadata, Note 필드 없음)
- `internal/astgrep/scanner.go:43-73` — Finding 구조체 (Metadata, Note 있음)
- `internal/astgrep/scanner.go:232-236` — sg 미존재 graceful degradation
- `internal/astgrep/scanner.go:302-306` — runSingleRule defer cancel() (@MX:WARN)
- `internal/astgrep/scanner.go:361-374` — scanWithRules 메타데이터 주입 (Metadata 전파 없음)
- `internal/astgrep/analyzer.go:242` — 루프 내 context.WithTimeout cancel() 누수
- `internal/astgrep/analyzer.go:298` — Replace 루프 동일 패턴
- `internal/astgrep/rules.go:109-139` — LoadFromDir 재귀 로딩
- `internal/astgrep/rules.go:186-208` — splitYAMLDocs 멀티-doc 지원
- `internal/astgrep/sarif.go:197-206` — toFileURI 미완성 구현
- `internal/cli/astgrep.go:193-198` — detectSGVersion 미구현 placeholder
- `internal/hook/quality/astgrep_gate.go:20-58` — RunAstGrepGateV2
- `internal/hook/quality/astgrep_gate.go:107-191` — RunAstGrepGate (V1)
- `internal/hook/quality/astgrep_gate.go:125` — V1: LoadFromDirectory (non-recursive)
- `internal/hook/security/ast_grep.go:172-213` — ScanMultiple 고루틴 폭주
- `internal/hook/security/ast_grep.go:331-347` — supportedLanguages (flutter/R 누락)
- `.moai/config/astgrep-rules/sgconfig.yml:1-24` — 16개 언어 선언 (11개 빈 디렉토리)
- `.moai/config/astgrep-rules/security/injection.yml:8-13` — metadata (YAML에 있으나 파싱 불가)
- `.moai/config/astgrep-rules/go/error-handling.yml:11` — go-error-not-wrapped false-positive 위험
- `.moai/config/astgrep-rules/security/web.yml:24-35` — sec-csrf-no-token-check 과도한 패턴
