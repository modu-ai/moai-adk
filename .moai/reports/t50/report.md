# t50 — internal/astgrep 경로 드리프트 잔여 2건 수정

- 카드: t50 (칸반 배치 tjv7iy round 3, cluster C, run 레인)
- 워크트리: `.claude/worktrees/agent-a05f5722891df4994` (branch `WT-t50`, base = `release/v3.1.1` @ `051a2fa94`)
- 대상: 결함 (1) `internal/astgrep/rule_seed_test.go` — 은닉 스킵, 결함 (2) `internal/astgrep/scanner.go:36` 기본값이 존재하지 않는 경로

## 0. 선행 판단 (lead 요구 사항)

**채택: 옵션 (b)의 완결 변형 — `gate.yaml ast_grep_gate.rules_dir`를 유일한 SSOT로 두고 코드 레벨 기본값 폴백을 제거하되, 템플릿 `gate.yaml`이 명시적 값(`.moai/config/astgrep-rules`)을 갖도록 편집 + `make build` 재임베드.**

해결 순서는 단일 규칙으로 통일됐다:

```
--rules-dir 플래그 (> 우선)  →  gate.yaml ast_grep_gate.rules_dir  →  빈 값 = "미설정" (0 룰)
```

**기각: 옵션 (a) — 코드 기본값을 새 경로(`.moai/astgrep-rules`)로 이동.** 배포 사용자의 룰셋은 템플릿이 **구 경로**에 배포한다(`internal/template/templates/.moai/config/astgrep-rules/`, `moai update`가 `.moai/config`를 통째로 wipe 후 재배포 — CLAUDE.local.md §2.3). 기본값이 새 경로를 가리키면 out-of-box `moai ast-grep`·어드바이저리 게이트가 룰 0개를 **조용히** 반환한다 — 하드 제약 위반이다. 이를 구제할 변형(템플릿 배포 경로 자체를 `.moai/astgrep-rules`로 이전)은 아키텍처적으로 기각: 해당 경로는 `CleanMoaiManagedPaths`의 wipe 뿌리 **밖**이라 `moai update`가 룰을 더 이상 정리·재배포하지 못해 고아 룰이 누적되고(2026-08-15 결함류 그대로 재현), 로컬 전용 dogfood 네임스페이스와 충돌한다.

**옵션 (b) 문언 그대로도 불충분했던 이유:** 템플릿 `gate.yaml`이 `rules_dir: ""`인 상태에서 폴백만 제거하면 템플릿 사용자가 여전히 룰 0개로 조용히 강등된다. 그래서 "SSOT 통합 + 템플릿 gate.yaml 명시적 값"이 최소 일관 변형이며, run 레인 프롬프트가 예시로 예견한 경로 그대로다.

## 1. 주장 (Claim)

1. 결함 (1): `rule_seed_test.go`의 5개 언어 케이스(ruby/php/elixir/csharp/kotlin)는 **한 번도 채워진 적 없는** 디렉터리를 가리켰고, `LoadFromDir`이 존재하지 않는 디렉터리에 `([]Rule{}, nil)`을 반환하므로(REQ-ASTG-UPG-011) 전 서브테스트가 수년째 `t.Skipf`로 통과해왔다. 해당 언어 룰은 어디에도 존재하지 않아 폐기됐고, 생존 표면은 `.moai/astgrep-rules/{go,security}`(추적 파일, 템플릿 사본과 byte-identical — `diff -r` 실측)뿐이다. 테스트를 이 표면에 재구조화하면 스킵 없이 실제 커버리지가 회복된다.
2. 결함 (2): `DefaultScannerConfig().RulesDir`의 하드코딩 `.moai/config/astgrep-rules`는 이 저장소에 존재하지 않는 경로라, `--rules-dir` 미지정 `moai ast-grep`이 룰 0개로 동작했다(CLAUDE.local.md §2.2가 이미 문서화한 결함). SSOT 통합(코드 폴백 제거 + CLI가 gate.yaml 해석 + 템플릿 gate.yaml 명시값)으로 dogfood CLI는 26개 룰을 찾고, 템플릿 사용자는 배포 룰을 계속 찾는다.

## 2. 증거 (Evidence) — 명령 + verbatim 출력

### 2-1. 결함 (1) 재현 — 수정 전 스킵 (RED 전 단계)

```
$ go test ./internal/astgrep/ -run 'TestRuleSeed' -count=1 -v
    rule_seed_test.go:121: rules dir .../.moai/config/astgrep-rules/ruby not populated (expected >=3 rules); skipped pending SPEC-UTIL-002 rule seeding
    ... (php/elixir/csharp/kotlin 동일) ...
--- PASS: TestRuleSeed (0.03s)
    --- SKIP: TestRuleSeed/ruby (0.00s)
    --- SKIP: TestRuleSeed/kotlin (0.00s)
    --- SKIP: TestRuleSeed/csharp (0.00s)
    --- SKIP: TestRuleSeed/elixir (0.00s)
    --- SKIP: TestRuleSeed/php (0.00s)
```

### 2-2. 결함 (2) RED — 기본값·폴백 핀 테스트 (수정 대상을 정확히 지목)

```
$ go test ./internal/astgrep/ -run 'TestDefaultScannerConfig_Fields' -count=1
    scanner_test.go:547: DefaultScannerConfig().RulesDir = ".moai/config/astgrep-rules", want empty (resolved by callers from gate.yaml)

$ go test ./internal/hook/quality/ -run 'TestDefaultAstGrepGateConfig|TestAstGrepGate_GateConfigIntegration' -count=1
    astgrep_gate_test.go:20: RulesDir: want empty (gate.yaml is the SSOT), got ".moai/config/astgrep-rules"
    astgrep_gate_test.go:42: AstGrepGate.RulesDir: want empty (gate.yaml is the SSOT), got ".moai/config/astgrep-rules"

$ go test ./internal/hook/ -run 'TestPreToolHandler_LoadGateConfig' -count=1
        pre_tool_test.go:1372: AstGrepGate.RulesDir default = ".moai/config/astgrep-rules", want empty

$ go vet ./internal/cli/
    internal/cli/astgrep_rulesdir_test.go:33:13: undefined: resolveRulesDir  (신규 API RED)
```

추가 RED(2차): 프로젝트-루트 상대 결합 규칙을 핀한 뒤:

```
$ go test ./internal/cli/ -run 'TestResolveRulesDir' -count=1
        astgrep_rulesdir_test.go:49: resolveRulesDir("", dir) = ".moai/astgrep-rules", want "/var/folders/.../TestResolveRulesDir.../001/.moai/astgrep-rules"
```

(게이트가 `RunAstGrepGateV2`에서 `filepath.Join(projectDir, cfg.RulesDir)`로 해석하는 것과 CLI가 일치해야 한다고 판정 — 절대경로는 join하지 않는 가드 포함, `filepath.Join("/a","/b")="/a/b"` 오염 방지.)

### 2-3. GREEN — 수정 후 동일 테스트

```
$ go test ./internal/astgrep/ -run 'TestDefaultScannerConfig_Fields|TestRuleSeed' -count=1 -v
--- PASS: TestDefaultScannerConfig_Fields (0.00s)
--- PASS: TestRuleSeed (0.02s)
    --- PASS: TestRuleSeed/go (0.00s)
    --- PASS: TestRuleSeed/security (0.00s)
    --- PASS: TestRuleSeed/fixtures-scan (0.07s)        <- 스킵 0건

$ go test ./internal/cli/ -run 'TestResolveRulesDir|TestAstGrepCmdRulesDirFlagNoDefault' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli

$ go test ./internal/hook/quality/ ... / ./internal/hook/ -run (LoadGateConfig 계열)
ok  (각각)
```

### 2-4. 바이너리 end-to-end (하드 제약 검증, `make build` 후 `bin/moai`)

dogfood 체크아웃 — 카드가 명명한 결함이 수정됨 (수정 전에는 룰 0개):

```
$ ./bin/moai ast-grep --dry ./
rules to apply (26):
  [warning] go-goroutine-without-context - ... (go)
  ...
EXIT:0
```

템플릿 사용자 시뮬레이션 — 템플릿 gate.yaml(명시값) + 템플릿 배포 룰만 있는 임시 프로젝트:

```
$ CLAUDE_PROJECT_DIR=/tmp/t50-user ./bin/moai ast-grep --dry ./
rules to apply (26): ...  EXIT:0
```

미설정 프로젝트 — 조용한 0이 아니라 명시적 안내:

```
$ CLAUDE_PROJECT_DIR=/tmp/t50-empty ./bin/moai ast-grep --dry ./
ast-grep: no rules directory configured — set ast_grep_gate.rules_dir in .moai/config/sections/gate.yaml or pass --rules-dir; scanning with 0 rules.
no rules to apply
EXIT:0
```

### 2-5. 픽스처 설계의 사전 실측 (sg 0.40.5, 테스트에 인코딩 전)

```
$ sg scan --config .moai/astgrep-rules/sgconfig.yml --json <fixture>
valid.go      -> 0 findings
violation.go  -> 3 findings ['go-interface-empty-not-any', 'sec-log-injection-unsanitized', 'sec-weak-hash-md5']
suppressed.go -> 0 findings   (주: 억제 마커가 대상 코드와 인접해야 함 — prose에 마커 문구가 들어가면 unused-suppression 진단 3건 발생, 실측으로 배치 확정)
```

## 3. Baseline 귀속

- base: `release/v3.1.1` @ `051a2fa94` (merge-fast-forward, `git merge-base --is-ancestor release/v3.1.1 HEAD` 통과). 병합 직전 release가 전진했으나(동시 레인 CLEAN-HOME M1) 해당 1커밋도 반영 완료.
- 전체 패키지 런(카드 대상, serial, `-count=1`):
  - `go test ./internal/astgrep/` -> ok 2.287s
  - `go test ./internal/hook/quality/` -> ok 3.730s
  - `go test ./internal/hook/` -> ok 37.015s
  - `MOAI_SKIP_LIVE_CODEX=1 go test ./internal/cli/` -> ok 281.627s
  - `go test ./internal/template/` -> ok 37.423s (`make build` 카탈로그 재생성 후)
- `gofmt -l` (수정 11개 .go 파일) -> 빈 목록. `go vet ./internal/astgrep/ ./internal/cli/ ./internal/hook/ ./internal/hook/quality/` -> 무출력.

## 4. 미검증 (Gaps)

- **live-codex 테스트 1건 실패 (본 변경과 무관함을 근거로 보고)**: `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` — 실제 codex 백엔드 판정에 의존하는 LIVE 테스트로, 테스트 본문 주석 자체가 "codex returned an inconclusive/pass verdict on this fixture (a real result — report it)"를 규정. `git diff --stat HEAD | grep -c codex` = 0 (내 diff에 codex 파일 0건)이므로 파급 없음. 옵트아웃 실행(`MOAI_SKIP_LIVE_CODEX=1`)에서 패키지 전체 ok.
- `internal/hook/security/rules.go:40-41`의 sgconfig 탐색 체인은 구 경로를 하드코딩하나 **템플릿 배포 경로를 가리키므로 배포 사용자에게 여전히 정확**하다. 이 저장소 dogfood sgconfig는 이 체인에 애초에 없었다(수정 전후 동일, 별개 하위시스템 — 카드 범위 밖).
- 템플릿 사용자 시뮬레이션은 `moai init` 실행이 아니라 템플릿 파일 복사 + `CLAUDE_PROJECT_DIR` 지정으로 재현했다(개발 체크아웃에서 init 실행 금지 규율).
- `moai gate` CLI 전체 동작(스캔 실행)은 코드 경로가 `mapConfigGateToQuality` 제외 불변이라 단위 테스트로만 검증(게이트 실행 E2E 미실시).
- 발견된 **룰셋 선재 결함(본 카드 범위 밖, 후속 카드 후보)**: `sec-template-injection-html` 룰(`security/web.yml`)이 실제 `template.HTML(u)` 코드에 sg 0.40.5에서 0매치다(-p 직접·파일 기반 양쪽 실측). 위반 픽스처는 이 룰을 의도적으로 배제하고 실측 발화 3룰로 구성했다.

## 5. 잔여 위험 (Residual-risk)

- **혼합 배포 전이 상태**: 구 템플릿(`rules_dir: ""`) + 신규 바이너리 조합에서는 gate.yaml이 빈 값으로 흘러 어드바이저리 게이트가 룰 없이 통과한다. `moai update` 한 번으로 자치되는 1-릴리스 전이 상태다(바이너리와 템플릿은 릴리스 단위로 함께 배포 — #1265 전례와 동일 취급). 발견 가능성은 CLI 측 stderr 안내로 완화.
- dogfood `moai update` 후에는 템플릿 gate.yaml이 `.moai/config/astgrep-rules`(배포 룰 — 작동하는 기본값)로 되돌린다. dogfood 룰셋 재연결은 로컬 gate.yaml 복원 필요 — 파일 내 주석에 이제 이 규칙이 문서화돼 있다(§2.3 wipe 결함은 별도 카드).
- 스킵 회복된 seed 테스트가 저장소 추적 룰셋(`.moai/astgrep-rules`)에 하드 의존한다. 룰셋 축소(디렉터리당 <3)는 이제 테스트 실패로 표면화된다 — 의도된 변화다.
- go/ 12룰에 metadata(owasp/cwe)가 없는 비대칭은 현행 설계로 테스트에 문서화만 했다(가짜 매핑 제조 금지). security/ 14룰에만 metadata 불변식을 걸었다.

## 부록 — 변경 파일 목록

| 파일 | 변경 |
|---|---|
| `internal/astgrep/scanner.go` | `DefaultScannerConfig().RulesDir` -> `""` (+규약 주석) |
| `internal/astgrep/rule_seed_test.go` | 5개 언어 케이스 -> 생존 표면(go/security 로더 불변식 + 공유 Go 픽스처 sgconfig 스캔) 재구조화 |
| `internal/astgrep/testdata/fixtures/go/{valid,violation,suppressed}.go` | 신규 픽스처 (sg 0.40.5 실측 발화 검증) |
| `internal/astgrep/scanner_test.go` | 기본값 핀 갱신 ("" 기대) |
| `internal/cli/gate.go` | `mapConfigGateToQuality` 빈값->구경로 폴백 제거 |
| `internal/cli/astgrep.go` | `--rules-dir` 기본 "" + `resolveRulesDir`(플래그 > gate.yaml > 빈, 상대경로 프로젝트 루트 결합) + 미설정 stderr 안내 |
| `internal/cli/astedit.go` | `--rules-dir` 기본 "" + 룰 모드에서 동일 해석 |
| `internal/cli/astgrep_rulesdir_test.go` | 신규 — 해석 규칙·플래그 기본값 핀 |
| `internal/hook/pre_tool.go` | 미러 폴백 제거 (gate.go와 대칭) |
| `internal/hook/pre_tool_test.go` | 폴백 핀 -> "" 기대 |
| `internal/hook/quality/astgrep_gate.go` | `DefaultAstGrepGateConfig().RulesDir` -> `""` |
| `internal/hook/quality/astgrep_gate_test.go` | 기본값 핀 -> "" 기대 |
| `internal/template/templates/.moai/config/sections/gate.yaml` | `rules_dir: ".moai/config/astgrep-rules"` 명시 (+`make build` 재임베드) |
| `.moai/config/sections/gate.yaml` | 주석 갱신 (SSOT 규칙·update 후 재연결 안내; 값 불변) |
