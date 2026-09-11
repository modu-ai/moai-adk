# sync-audit — card t659, SPEC-CON-AMEND-APPLY-001 (Tier L, 0.1.8)

- 감사자: sync-auditor (독립 평가, 평가 프로필 `default`, flat 가중 채점)
- 측정 트리: `WT-amend-apply` HEAD `f0ab0ec17930627316a3b36faf08f0ff156566e8`, 작업 트리 clean (감사 전후 `git status --porcelain` 무출력)
- 툴체인: `go version go1.26.8 darwin/arm64`, `golangci-lint has version 2.10.1`
- 카드 diff: `fe8cc9875..HEAD` (first-parent, 범위 안 병합 커밋 0개 — `git log --merges --oneline fe8cc9875..HEAD` 무출력)

## 판정

**PASS (권고).** blocking 결함 0건. 필수 통과 차원(Functionality, Security) 모두 임계값 충족. 아래 결함 9건은 모두 optional이다.

| 차원 | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | AC 테스트 22개 + 가드 테스트 2개를 이 트리에서 재실행: top-level 24 RUN / 24 PASS / 0 FAIL / 0 SKIP, 전체 RUN 102줄. AC-CAA-016 grep 재실행. CLI 셀 4개는 슬롯 증거(`eec1c334c`) 인용 — 그 뒤 Go 변경은 주석 3줄뿐 |
| Security (25%) | 85/100 | PASS | Critical/High 0. 경로 격리 검사가 심볼릭 링크를 해석하고 경로 구분자 경계에서 비교함. Low 3건(F1–F3) |
| Craft (20%) | 88/100 | PASS | 커버리지 88.3% (≥85), vet 0, golangci-lint 0 issues, 뮤턴트 42 + 가드 5 모두 kill(증거 인용). MX 누락 2건(F5, F6) |
| Consistency (15%) | 88/100 | PASS | 파일명·오류 래핑·경로 `%s` 규칙 준수. CHANGELOG 사실 오류 1건(F4) |

가중 평균 88.9, 조화 평균 88.2.

## 리드 지정 검토 항목

### 1. SPEC 밖 가드 G-A, G-B — 올바르고, 테스트되며, 어떤 AC도 깨지 않는다

- **G-A** (`internal/constitution/pipeline.go:98-100`): REQ-CAA-017 검사 바로 뒤, Layer 1 앞에서 두 모드 모두 빈 `After`를 거부한다. `TestExecute_EmptyAfter_Rejected`(`apply_test.go:1177`)는 오류 문자열·규칙 ID·게이트 호출 0회·트리 스냅숏·락 해제를 모두 단언하며 이번 실행에서 PASS. 순서상 REQ-CAA-020/021(격리 검사)과 REQ-CAA-017이 먼저 오므로 해당 AC의 오류 우선순위는 바뀌지 않는다. CLI도 `--after ""`를 이미 거부하므로(`internal/cli/constitution.go:476-478`) G-A는 CLI 밖 호출자를 위한 방어선이다. 한계는 F8 참고(공백만 있는 `After`는 통과).
- **G-B** (`pipeline.go:105-108`): 격리 검사와 REQ-CAA-017, G-A 다음, Layer 1 앞. `sameFile`은 파일 I/O 없이 `filepath.Abs` 결과를 비교한다. 검사 대상 경로는 `prepareApply`가 실제로 쓰는 것과 같은 `ruleFilePath(projectDir, file)` 결합이고, 레지스트리 경로와 로그 경로도 `Execute`가 `prepareApply`에 넘기는 문자열 그대로다(`pipeline.go:77, 82, 105, 166`). `TestExecute_RuleFileIsRegistryOrLog_Rejected`(`apply_test.go:1217`)는 두 반쪽을 각각 검사하고, 거부 없는 대조 케이스를 두며, 게이트 호출 0회를 단언한다. 이번 실행에서 6개 서브테스트 모두 PASS. RED 커밋(`0893611ad`)이 수정 커밋(`5801ebda0`)보다 앞서므로 순서도 커밋 그래프로 입증된다.
- AC 22개 테스트와 패키지 전체가 두 가드가 들어간 트리에서 통과하므로 두 가드는 어떤 AC도 깨지 않는다. REQ-CAA-012(드라이런이 실제 적용과 같은 오류를 반환)도 유지된다 — 두 가드 모두 모드와 무관하게 같은 자리에서 실행된다.

### 2. 동작 변경 — 코드로 확인됨

아직 없는 evolution log를 `file:`로 가리키는 규칙 항목의 흐름:

1. `LoadAmendRegistry`의 항목별 격리 검사가 가장 가까운 기존 조상으로 경로를 판정하므로 통과한다(`registry_path.go:108-134`). `LoadRegistry`는 해당 항목을 orphan 경고로만 처리하며 오류를 내지 않는다(`loader.go:145-149`).
2. G-B는 문자열만 비교하고 파일 존재 여부를 보지 않으므로 Layer 1 전에 G-B 메시지로 거부한다.
3. 이동 전(`ae58f0383`)에는 같은 비교가 `prepareApply` 안에서 `readForChange("rule file", …, false)` **다음**에 있었다. 그 호출이 `os.Stat` 오류(`no such file`)로 먼저 실패했으며, 그것도 Layer 5 승인 **이후**였다.

→ 리드 서술과 일치. 덧붙이면 이전 오류는 승인 뒤에 났고, 드라이런에서도 같은 순서였다. 이 경로를 덮는 테스트는 없다(선언된 대로).

### 3. 잔여 위험 서술의 정확성

| 선언된 위험 | 판정 |
|---|---|
| `sameFile`이 심볼릭 링크를 해석하지 않음 | 정확하다. 다만 **불완전하다**: 대소문자 무시 파일시스템 별칭과 하드 링크 별칭도 같은 이유로 놓친다(F1, 미선언) |
| backup/temporary write 복원 Gap(운영자 승인 공란) | 정확하다. `commitChanges`의 `fail()`이 두 단계 모두에서 `restoreAll`을 호출함을 코드로 확인했다(`apply_commit.go:113-124`). 다만 이 경로는 결함 주입 없이도 자연 조건으로 도달할 수 있다(F7) |
| Windows 로컬 미실행 | 정확하다. `GOOS=windows GOARCH=amd64 go vet ./internal/constitution/` exit 0을 이번에 재측정(컴파일만) |

미선언 위험: F1(대소문자·하드 링크 별칭), F2(승인 대기 중 레지스트리 TOCTOU), F3(트리 안 심볼릭 링크 대상이 링크째 교체됨), F7(`.moai/research/` 부재 시 임시 쓰기 실패), F9(거부된 실제 모드 실행이 락 디렉터리를 남길 수 있음, §F 범위 밖).

## Findings (구조화 결함 목록)

모든 항목은 **optional**이다. blocking은 없다.

- **F1** [Low][optional][신뢰도: 메커니즘 높음 / 결과 중간] `internal/constitution/pipeline.go:267-272` — `sameFile`은 경로 문자열만 비교하므로 대소문자 무시 파일시스템(이 머신의 APFS, Windows NTFS)의 별칭과 하드 링크 별칭을 놓친다. 실측: `ls -l .claude/rules/moai/core/ZONE-REGISTRY.md` → `-rw-r--r--@ 1 goos staff 35135 … ZONE-REGISTRY.md`(해석됨). 이 경우 격리 검사도 통과하므로 G-B만이 방어선이다. 코드로 추적한 결과(미관측): 레지스트리 rename이 규칙 파일 rename보다 뒤에 오므로 최종 레지스트리는 정상이다. 대신 원본 규칙 파일 갱신이 조용히 사라지고, 로그 별칭도 같은 식으로 사라진다. 심볼릭 링크 별칭이면 링크가 일반 파일로 바뀐다. 필요한 수정: 두 경로가 모두 존재하면 `os.Stat` + `os.SameFile`로 비교하고, 그렇지 않으면 `resolveExisting` 결과끼리 비교한다. 잔여 위험 문구에 대소문자·하드 링크 별칭을 추가한다.
- **F2** [Low][optional][신뢰도 중간, 코드 판독] `internal/constitution/apply_transform.go:75-137`, `pipeline.go:88-97, 166` — REQ-CAA-017 검사는 게이트 전에 읽은 레지스트리 기준이다. `prepareApply`는 Layer 5 승인 후 레지스트리를 **다시** 읽는데, `rewriteRegistryClause`는 새로 읽은 대상 항목이 아직 `Before`를 갖고 있는지 확인하지 않는다. 승인 대기 중에 amender가 아닌 누군가가 대상 clause를 고치면 그 수정이 덮어써지고, 로그의 `clause_before`는 낡은 값이 된다(락은 amender끼리만 직렬화한다). 원본 규칙 파일은 옛 clause 정확히 1회 검사로 보호된다. SPEC 요구사항 위반은 아니다. 필요한 수정: `rewriteRegistryClause`에서 재작성 전 디코딩한 대상 clause가 `rule.Clause`와 같지 않으면 거부한다.
- **F3** [Low][optional][정보] `internal/constitution/apply_commit.go:127-131` — 트리 안의 심볼릭 링크를 가리키는 대상(규칙 파일·레지스트리·로그)에 rename하면 링크 자체가 일반 파일로 교체되고 링크 대상은 옛 바이트를 유지한다. 실측: `find .claude/rules .moai/research -type l | wc -l` → `0`이므로 현재는 잠재 위험이다. 필요한 수정: 잔여 위험에 기록하거나, 대상이 심볼릭 링크이면 거부한다.
- **F4** [Low][optional][신뢰도 높음] `CHANGELOG.md:12` — "a dry-run reported success without reading any file"는 사실이 아니다. 기준선 `fe8cc9875:internal/constitution/pipeline.go:66-67`의 드라이런은 `LoadRegistry`로 레지스트리를 읽었고, Layer 4는 로그를 읽었다(`RateLimiter.Admit`). 필요한 수정: "without running any apply-step validation"(적용 단계 검증을 전혀 하지 않고)으로 고친다.
- **F5** [Low][optional][신뢰도 높음] `internal/constitution/evolution_log.go:25-26` — `LoadEvolutionLogs`의 @MX:ANCHOR/REASON이 이미 삭제된 `applyAmendment`를 호출자로 적고 있다. 프로덕션 호출자는 2곳이다(`rate_limiter.go:42`, `evolution_log.go:81`). 적용 단계는 `parseEvolutionLog`를 직접 호출한다(`pipeline.go:245`). 필요한 수정: REASON을 현재 호출자로 고치거나, 프로토콜에 따라 @MX:NOTE로 강등한다.
- **F6** [Low][optional][신뢰도: Execute 높음 / 나머지는 인용] `pipeline.go:64` (`Execute`), `apply_transform.go:75` (`rewriteRegistryClause`), `evolution_log_reader.go:94` (`evolutionLogBlocks`), `evolution_log_reader.go:271` (`entry`) — 복잡도 ≥15인데 @MX:WARN이 없다(SHOULD 수준). manager-docs 측정값 22/18/17/17을 인용한다. 나는 gocyclo를 실행하지 않았다(허용 명령 밖). `Execute`의 분기점을 손으로 세면 21개(+1 = 22)로 인용값과 일치한다. 필요한 수정: 네 함수에 @MX:WARN + @MX:REASON을 붙인다.
- **F7** [Info][optional][신뢰도 중간, 코드 판독] `apply_commit.go:51-63` — `writeSibling`은 대상 디렉터리를 만들지 않는다. `<projectDir>/.moai/research/`가 없고 락이 다른 곳에 있으면 첫 개정이 "temporary write" 단계에서 실패하고, 승인된 공란(미주입) 복원 경로를 탄다. CLI 경로에서는 기본 락(`.moai/research/.amendment.lock`, cwd=projectDir)이 `MkdirAll`로 디렉터리를 만들어 주므로 해당하지 않는다. 템플릿에 `.moai/research`는 없다(`ls internal/template/templates/.moai/research` → No such file or directory). 필요한 수정(선택): 로그 디렉터리를 만들고 복원 시 제거하거나, 이 경로를 잔여 위험으로 기록한다.
- **F8** [Info][optional] `pipeline.go:98` — 공백만 있는 `After`는 G-A를 통과하고, Layer 5 승인 **후** REQ-CAA-016에서야 거부된다(공백은 원본에 거의 항상 있음). 더 넓게 보면 REQ-CAA-001/002/004/016 실패는 SPEC 설계상 모두 승인 뒤에 드러난다. 리드가 G-B를 옮긴 근거("승인 뒤 거부는 승인을 헛되게 한다")가 이 경우에도 그대로 적용된다. SPEC과 충돌하지는 않는다. 권고: 후속 카드 후보로 기록한다(드라이런 선행 안내, 또는 순수 검증의 게이트 전 이동).
- **F9** [Info][optional] `pipeline.go:291` — 실제 모드의 `acquireLock`은 `MkdirAll`로 락 디렉터리를 만들고, 락 해제 때 디렉터리는 지우지 않는다. 거부된 실행이 경로 하나(`.moai/research/`)를 남길 수 있어 REQ-CAA-020의 "adding no path"와 긴장 관계에 있다. 기존 동작이며 §F(기본 락 경로)의 범위 밖이다. 테스트는 락을 픽스처 밖에 두므로 관측되지 않는다.

## 기타 확인 항목

- **상태 전이**: `git show 1f9946188 -- …/spec.md`의 diff는 frontmatter 두 줄(`status: in-progress` → `completed`, `updated: 2026-09-11` → `2026-09-12`)뿐이고 본문 변경은 0줄이다. `128a5ea52..HEAD` 범위에서 spec.md 변경은 이 4줄(+2/−2)이 전부다. 커밋 트레일러는 `Authored-By-Agent: manager-docs`.
- **PRESERVE**: `git diff --stat fe8cc9875 HEAD -- internal/spec/ internal/constitution/loader.go internal/constitution/rate_limiter.go internal/constitution/validator.go internal/cli/doctor.go` → 무출력. `LoadRegistry`는 카드 diff에서 바뀌지 않았다. 범위 안 병합 커밋이 없으므로 트리 diff가 곧 결론이다.
- **AC-CAA-025 보존 실행**: `TestLinter_AC08_DanglingRuleReference` PASS(재실행).
- **AC-CAA-016**: `internal/constitution/` 전체에서 `not yet implemented` 0건. 같은 패턴으로 `internal/` 전체를 훑으면 다른 패키지 적중이 나오므로 대조군이 성립한다. 스텁 테스트 이름 4개는 `pipeline_test.go:150, 317, 364`의 은퇴 주석에만 남아 있다.
- **AC-CAA-017**: 이 셸에서 `CLAUDE_PROJECT_DIR`, `MOAI_CONSTITUTION_REGISTRY`는 설정되지 않았다(스크럽 상태와 동등). 실행 후 실제 파일 sha256: 레지스트리 `f7707b1d…cf4be21`, 로그 `f5735051…d92469f0`으로 progress 기록값과 같다. primary 체크아웃의 두 파일 mtime(2026-09-09, 2026-05-13)은 이번 실행보다 이르다. `internal/constitution/.moai`와 `internal/cli/.moai`는 없다.
- **CHANGELOG**: F4를 뺀 나머지 서술(3파일 원자 적용, 순서, 복원, 드라이런 무쓰기, 공유 resolver·격리 검사, 가드 2개, 동작 변경, 25/25·42+5 뮤턴트·88.3%·잔여 위험)은 코드와 증거에 부합한다.

## 5-섹션 증거

### Claim

1. AC-CAA-001…025의 테스트 기반 셀은 이 트리에서 모두 PASS다.
2. `internal/constitution` 커버리지 88.3%, vet·lint 무결함.
3. `LoadRegistry`와 비-amend 호출자는 카드 diff에서 바뀌지 않았다.
4. G-A와 G-B는 올바르고 테스트되며 AC를 깨지 않는다. 리드가 서술한 동작 변경은 코드와 일치한다.
5. sync 린트의 판정 빌드 지연(`ed71054d3`)은 이 SPEC의 린트 결과에 영향을 줄 수 없다(코드 판독 경계, 아래 Gaps).

### Evidence

```
$ go test ./internal/constitution/ -run '^(TestApply_ExactlyOnce_Success|…22 AC tests…|TestExecute_EmptyAfter_Rejected|TestExecute_RuleFileIsRegistryOrLog_Rejected)$' -count=1 -v
toprun=24 toppass=24 fail=0 skip=0 allrun=102
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.617s

$ go test ./internal/constitution/ -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.824s	coverage: 88.3% of statements

$ go vet ./internal/constitution/          → vet_exit=0 (무출력)
$ golangci-lint run ./internal/constitution/...
0 issues.
$ GOOS=windows GOARCH=amd64 go vet ./internal/constitution/   → win_vet_exit=0

$ go test ./internal/spec/ -run '^TestLinter_AC08_DanglingRuleReference$' -count=1 -v
--- PASS: TestLinter_AC08_DanglingRuleReference (0.48s)
ok  	github.com/modu-ai/moai-adk/internal/spec	0.847s

$ git diff --stat fe8cc9875 HEAD -- internal/spec/ internal/constitution/loader.go internal/constitution/rate_limiter.go internal/constitution/validator.go internal/cli/doctor.go
(무출력)

$ git diff --stat eec1c334c HEAD -- internal/ cmd/ pkg/ go.mod go.sum
 internal/constitution/registry_path.go | 3 +++      ← ruleFilePath 위 @MX 주석 3줄뿐

$ git merge-base --is-ancestor ed71054d3 fe8cc9875   → exit 0
$ grep -niE '^[[:space:]]*#{2,}.*(acceptance|success criteria|ac matrix|ac summary|수락 기준|인수 기준|검수 기준|수용 기준|성공 기준)' …/spec.md
exit=1   (대조: grep -ciE 'acceptance' …/spec.md → 8)
```

### Baseline-attribution

- 모든 측정의 대상은 이 실행의 트리 `f0ab0ec17`이며, 결과는 위에 붙인 그대로다.
- CLI 셀(AC-CAA-015, AC-CAA-017 CLI 절반, AC-CAA-022 CLI, AC-CAA-024 CLI 케이스)과 CLI 뮤턴트 3개는 `.moai/reports/t659/run/slot/summary.md`(트리 `eec1c334c`)를 인용한다. 그 뒤 Go 변경은 `registry_path.go`의 주석 3줄뿐이므로 코드 기준으로 동일 트리다. 리드 지시에 따라 `internal/cli` 테스트는 재실행하지 않았다.
- 뮤턴트 42 + 가드 5의 kill 결과는 `.moai/reports/t659/run/slot/remeasure/`, `run/guards/`를 인용한다. 재실행하지 않았다.
- sync 린트 결과는 `.moai/reports/t659/sync/lint.txt`를 인용한다. 판정 빌드는 `ed71054d3`(트리의 엄격한 조상)이고, 그 기록 자체에 지연이 명시돼 있다.

### Gaps

- **sync 린트가 트리 빌드로 재실행되지 않음**(§2.2). `ed71054d3` 이후 트리에 들어온 `internal/spec` 변경은 카드가 아니라 기준선(`fe8cc9875`)에서 온 것이다. 그중 동작을 바꾸는 부분은 `parser.go`의 `findACSectionStart`/`extractACLines`(AC 섹션 앵커)뿐이고, `lint_coverage_sibling.go`는 주석 1줄 변경이다. 린트는 이 함수들을 `ParseAcceptanceCriteria(spec.md body)`(`lint.go:693`)로만 거친다. spec.md에는 옛 술어(`##` + "acceptance")와 새 술어(레벨 ≥2 + 어휘 목록) 어느 쪽에도 걸리는 제목이 한 줄도 없다(위 grep exit 1, 대조 8). 따라서 두 빌드 모두 앵커 −1을 반환해 결과가 같다 — **영향이 없다고 판단하며, 근거는 코드 판독과 grep이다.** 트리에서 빌드한 `moai spec lint` 실행 결과는 없다.
- CLI 테스트·뮤턴트는 이번 감사에서 재실행하지 않았다(인용만).
- gocyclo를 재측정하지 않았다(F6은 인용 + `Execute` 수기 계수).
- F1, F2, F3, F7의 결과 서술은 코드 판독이며 관측하지 않았다.
- AC-CAA-017의 세션 환경(`CLAUDE_PROJECT_DIR` 설정) 실행은 레인 증거(`slot/s4-*`)를 인용한다.
- Windows 테스트 실행은 없다(vet 컴파일만).

### Residual-risk

- F1–F3: 비정상 레지스트리 항목(대소문자·하드 링크·심볼릭 링크 별칭)이나 승인 대기 중 외부 수정이 있으면 원본 규칙 파일 갱신이 사라지거나 동시 수정이 덮어써질 수 있다. 실제 트리에서는 `.claude/rules`·`.moai/research` 아래 심볼릭 링크가 0개이고, 실제 레지스트리의 `file:` 101줄 중 `zone-registry` 또는 `evolution-log`를 가리키는 줄이 대소문자 무시로도 0줄이다(`grep -ciE '^\s*file:.*(zone-registry|evolution-log)'` → `0`, 대조 `grep -cE '^\s*file:'` → `101`).
- 승인된 공란: backup/temporary write 실패 복원은 결함 주입 없이 코드 판독으로만 확인됐다. F7 조건에서 자연 도달할 수 있다.
- 실제 적용 경로의 CLI 래퍼(G5)는 `Execute` 수준에서만 덮인다.
- Windows 판정은 CI 몫이다.

## Delta audit (e5a18feae..b36f50c4c)

- 감사자: sync-auditor (독립 평가, 평가 프로필 `default`, flat 가중 채점)
- 측정 트리: `WT-amend-apply` HEAD `b36f50c4c`, 감사 시작 시 작업 트리 clean (`git status --porcelain` 무출력)
- 범위: 커밋 5개 — `e8d16eaee`(F5/F6), `2496053a8`(F1 RED), `bca8cf96a`(F1 수정), `393a4af32`(F1 기록), `b36f50c4c`(F4 + 정합화). `internal/cli` 변경 0 (`git diff --stat e5a18feae..b36f50c4c -- internal/cli` 무출력)
- 툴체인: `golangci-lint has version 2.10.1`, Go는 이전 감사와 같은 go1.26.8 darwin/arm64

### 판정

**PASS (권고).** 델타에 blocking 결함은 없다. F1·F4·F5·F6 모두 닫혔다. 새 결함 2건(D1, D2)은 둘 다 optional이다.

| 차원 | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality (40%) | 94/100 | PASS | 가드 테스트 3개 재실행: 서브테스트 18개 전부 PASS, SKIP 0 (case_variant 포함). 패키지 전체 `ok`. RED 커밋이 수정 커밋보다 앞서며 프로덕션 코드를 담지 않음 |
| Security (25%) | 90/100 | PASS | F1 해소: 하드 링크·대소문자 변형·심볼릭 링크 별칭을 Layer 1 전에 거부. 로그 쪽 별칭은 코드상 같은 경로로 막히지만 테스트는 없음(D2) |
| Craft (20%) | 89/100 | PASS | 커버리지 88.4%, vet 0, win vet 0, lint 0 issues. 기록된 뮤턴트 7/7 kill. 내가 새로 넣은 뮤턴트 1개는 생존(D2) |
| Consistency (15%) | 88/100 | PASS | MX 태그 수치 실측 일치. 호출부 주석 한 줄이 F1 이후 사실과 어긋남(D1). CHANGELOG가 로그 별칭까지 테스트로 귀속(D2) |

가중 평균 91.1, 조화 평균 90.2.

### 커밋별 확인

**`e8d16eaee` (F5/F6) — 주석만 바뀌었고 수치도 맞다.**
- 비주석 줄 변경 0: `git diff -U0 e5a18feae e8d16eaee -- internal/`에서 `+`/`-` 줄 중 `//`로 시작하지 않는 줄을 걸러내면 grep exit 1(무출력), 주석 변경 줄은 16줄.
- 복잡도: 임시 설정(`gocyclo`, min-complexity 15)으로 golangci-lint를 돌린 결과 `(*Pipeline).Execute` 22, `rewriteRegistryClause` 18, `evolutionLogBlocks` 17, `(*logEntryDecoder).entry` 17 — 태그 수치와 모두 같다. 이전 감사에서 인용에 그쳤던 값이 이번에 실측으로 확인됐다.
- fan_in: `LoadEvolutionLogs(`의 비테스트 호출부는 `evolution_log.go:81`(MarkRolledBack), `rate_limiter.go:42`(Admit) 두 곳이다. 적용 단계는 `parseEvolutionLog`를 직접 부른다(`pipeline.go:248`). REASON의 "production fan_in 2"는 사실이다. 임계값 미만인데도 ANCHOR를 유지한 것은 "pending a demotion decision"으로 명시돼 있어 허용 범위로 본다.

**`2496053a8` → `bca8cf96a` (F1) — 순서, 정확성, 판별력 모두 성립한다.**
- 순서: `bca8cf96a`의 부모가 `2496053a8`이고, `2496053a8`이 바꾼 파일은 `internal/constitution/apply_test.go`와 `.moai/reports/t659/run/guards/f1-red.txt` 둘뿐이다. `f1-red.txt`에는 별칭 서브테스트 6개가 설정 단계가 아니라 `got nil`로 실패한 기록이 남아 있다.
- 정확성: `sameFile`(`pipeline.go:277-286`)은 두 경로가 모두 stat되면 `os.SameFile`로 판정하고, 그렇지 않으면 이전과 같은 `filepath.Abs` 비교로 떨어진다. `os.Stat`은 링크를 따라가므로 심볼릭 링크 별칭도 잡힌다. Stat 오류를 거부가 아니라 폴백으로 처리한 선택은 옳다. Stat 실패를 거부로 바꾸면 아직 없는 로그를 가리키는 경로를 가드가 식별 판정 없이 거부하게 되고, 권한 오류 같은 일시적 실패가 가드 메시지로 둔갑한다. 폴백으로 빠진 항목은 이름이 가리키는 파일이 없으므로, 게이트 이후 규칙 파일 읽기(`readForChange`)에서 거부되며 아무것도 쓰지 않는다(코드 판독, M-F1-fallback 관측과 일치).
- 판별력: 별칭 서브테스트는 `sameFile`과 무관하게 별칭이 레지스트리 바이트로 읽히는지 먼저 확인하므로, 설정 실패가 거부로 위장할 수 없다. 기록된 M-F1-samefile(식별 분기 제거)은 별칭 서브테스트 6개에서, M-F1-fallback은 absent_log_fallback에서 kill됐다. 뮤턴트 전후 `pipeline.go` sha256 `1ef48dcc…2467`은 현재 트리 값과 같다.
- 스팟 뮤턴트(직접 실행): 로그 쪽 절반만 옛 비교로 되돌렸다(`sameFile(rulePath, evolutionLogPath)` → `filepath.Abs` 문자열 비교). **패키지 전체가 통과해 생존했다**(D2). 레지스트리 절반과 로그 절반이 같은 함수를 쓰므로 현재 코드의 동작은 옳다. 다만 로그 별칭 거부를 직접 증명하는 테스트는 없다. 복원: `pipeline.go` sha256이 복원 후 `1ef48dcc…2467`, `git status --porcelain` 무출력.
- **F1은 닫혔다.** 이전 F1이 요구한 식별 비교(`os.SameFile`)가 들어갔고 잔여 위험 문구도 갱신됐다. 부재 경로 폴백은 제안했던 `resolveExisting` 비교 대신 `filepath.Abs` 비교를 택했다. 부재 파일의 별칭은 어차피 게이트 뒤 읽기에서 거부되므로 결과는 같고, 이 한계는 잔여 위험에 적혀 있다.

**progress §E.3 `residual_risk` — 정확하며, 사소한 누락 하나가 있다.**
- 항목 1(stat 불가 시 Abs 폴백): 정확하다. 매달린 심볼릭 링크, 아직 없는 로그의 대소문자 변형 모두 코드 흐름과 일치한다.
- 항목 2(게이트 전 1회 판정, 승인 대기 중 재확인 없음): 정확하다. `sameFile` 호출부는 `pipeline.go:109` 하나이고 `prepareApply`는 다시 부르지 않는다.
- 항목 3(대소문자 서브테스트는 대소문자 무시 FS에서만 실행): 정확하다. 이번 실행에서도 SKIP 없이 PASS였다.
- 누락: 로그 별칭(하드 링크·대소문자 변형·심볼릭 링크로 로그를 가리키는 `file:`)은 코드로만 막히고 테스트가 없다(D2). 이 사실이 잔여 위험에 없다.

**`b36f50c4c` (F4 + 정합화) — 바뀐 주장을 하나씩 대조했다.**
- F4: 기준선 `fe8cc9875:internal/constitution/pipeline.go`의 드라이런은 Layer 5(121행) 직후 133행 `if dryRun {`에서 반환하며, 적용 단계(`updateSourceFile`, 196행)에 도달하지 않는다. "returned success right after the gates without running any apply-step check"는 사실이다. **F4는 닫혔다.**
- 가드 뮤턴트 수 "7": 기존 5개(M-GA, M-GB, M-GB-reg, M-GB-log, M-GB-always) 재실행과 신규 2개(M-F1-samefile, M-F1-fallback)를 합친 수로, `mutants-f1-summary.txt`의 KILLED 7행과 일치한다.
- 커버리지 "88.4%": 이번 실행 결과와 `guards/cover-f1.txt` 모두 `coverage: 88.4% of statements`다.
- G-B 문구와 동작 변경의 "absent_log_fallback이 새 거부를 덮는다": 해당 서브테스트는 G-B 메시지와 게이트 호출 0회를 단언하며 PASS다. 사실이다.
- 알려진 한계 3줄: §E.3과 일치한다(위 확인 참고).
- `sync_lint_gap`: §2.2에 따라 설치 빌드 `ed71054d3`의 판정을 트리 근거로 쓰지 않는다고 명시했다. 이전 감사의 Gaps 서술과 같다.
- CHANGELOG의 "a hard-link, case-variant, or symbolic-link alias of the registry or the log is rejected as well (`TestExecute_RuleFileAliasOfRegistryOrLog_Rejected`)": 동작 자체는 코드상 사실이다. 그러나 괄호 속 테스트는 레지스트리 별칭만 실행하므로 로그 쪽까지 테스트로 귀속하는 것은 과장이다(D2).

### 회귀 확인

- `go test ./internal/constitution/ -count=1 -cover` → `ok`, 88.4%
- `go vet ./internal/constitution/` exit 0, `GOOS=windows GOARCH=amd64 go vet ./internal/constitution/` exit 0
- `golangci-lint run ./internal/constitution/...` → `0 issues.`
- `TestLinter_AC08_DanglingRuleReference` PASS (AC-CAA-025 보존)
- `internal/cli` 테스트는 지시에 따라 실행하지 않았다. 델타가 `internal/cli`를 건드리지 않음은 확인했다.

### 결함 상태

| ID | 이전 | 현재 | 근거 |
|---|---|---|---|
| F1 | Low, optional | **닫힘** | `bca8cf96a`, 별칭 테스트 서브테스트 8개 PASS, 뮤턴트 kill. 로그 쪽 테스트 공백은 D2로 분리 |
| F4 | Low, optional | **닫힘** | CHANGELOG 문구가 기준선 코드와 일치 |
| F5 | Low, optional | **닫힘** | REASON이 실측 호출부 2곳과 일치 |
| F6 | Low, optional | **닫힘** | WARN 4개, 복잡도 실측 일치 |
| F2, F3, F7, F8, F9 | optional | 변화 없음(델타 범위 밖) | 델타가 해당 코드를 건드리지 않음 |

### 새 결함 (구조화 결함 목록)

- **D1** [Low][optional][신뢰도 높음] `internal/constitution/pipeline.go:107` — G-B 호출부 주석 "it needs only the three paths, no file I/O"가 F1 이후 사실이 아니다. 이제 `sameFile`은 두 경로를 `os.Stat`한다. 이 주석은 `5801ebda0`에서 들어왔고, F1은 `sameFile`과 그 주석만 고쳤다. 필요한 수정: "it stats the three paths and reads no file content" 정도로 바꾸거나, 판정 방식은 `sameFile` 주석에 맡기고 해당 구절을 지운다.
- **D2** [Low][optional][신뢰도 높음, 뮤턴트 관측] `internal/constitution/apply_test.go:1279-1339`, `CHANGELOG.md:12` — `TestExecute_RuleFileAliasOfRegistryOrLog_Rejected`는 레지스트리 별칭만 만든다. 로그 쪽 절반을 옛 문자열 비교로 되돌린 뮤턴트가 패키지 전체를 통과했다. 동작은 공유 함수 덕분에 옳지만, CHANGELOG는 로그 별칭 거부까지 이 테스트로 귀속하고 §E.3 잔여 위험에도 이 공백이 없다. 필요한 수정(택1): 로그를 만들고 `rules/alias.md`를 로그의 하드 링크(또는 심볼릭 링크)로 만드는 `hardlink_log` 서브테스트를 추가한다. 또는 CHANGELOG 괄호를 "(registry aliases; the log half shares the same comparison)"처럼 좁히고 잔여 위험에 한 줄을 넣는다.

### 5-섹션 증거 (델타)

#### Claim

1. `e8d16eaee`는 주석만 바꿨고, 태그의 복잡도(22/18/17/17)와 fan_in(2)은 실측과 같다.
2. F1 RED 커밋은 수정 커밋보다 앞서고 프로덕션 코드를 담지 않는다. 수정은 올바르고, 테스트는 레지스트리 별칭을 판별한다.
3. F4 문구는 기준선 코드와 일치하고, `b36f50c4c`의 수치(가드 뮤턴트 7개, 88.4%)와 한계 서술은 증거와 일치한다.
4. 델타는 패키지 테스트·vet·lint·AC-CAA-025 보존 테스트를 깨지 않았다.

#### Evidence

```
$ git diff -U0 e5a18feae e8d16eaee -- internal/ | grep -E '^[+-]' | grep -vE '^(\+\+\+|---) ' | grep -vE '^[+-][[:space:]]*//'
noncomment_exit=1        (주석 변경 줄: 16)

$ golangci-lint run --config <scratchpad>/gocyclo.yml ./internal/constitution/...   (gocyclo only, min-complexity 15)
apply_transform.go:78:1: cyclomatic complexity 18 of func `rewriteRegistryClause` is high (> 15) (gocyclo)
evolution_log_reader.go:97:1: cyclomatic complexity 17 of func `evolutionLogBlocks` is high (> 15) (gocyclo)
evolution_log_reader.go:277:1: cyclomatic complexity 17 of func `(*logEntryDecoder).entry` is high (> 15) (gocyclo)
pipeline.go:67:1: cyclomatic complexity 22 of func `(*Pipeline).Execute` is high (> 15) (gocyclo)

$ grep -rn 'LoadEvolutionLogs(' --include='*.go' internal cmd pkg | grep -v '_test.go'
internal/constitution/evolution_log.go:27:func LoadEvolutionLogs(path string) ([]AmendmentLog, error) {
internal/constitution/evolution_log.go:81:	logs, err := LoadEvolutionLogs(path)
internal/constitution/rate_limiter.go:42:	logs, err := LoadEvolutionLogs(evolutionLogPath)

$ git show --stat 2496053a8   → parents=e8d16eaee; .moai/reports/t659/run/guards/f1-red.txt, internal/constitution/apply_test.go
$ git show --stat bca8cf96a   → parents=2496053a8; internal/constitution/pipeline.go | 13 ++++++++++++-

$ go test ./internal/constitution/ -run '^(TestExecute_RuleFileAliasOfRegistryOrLog_Rejected|TestExecute_RuleFileIsRegistryOrLog_Rejected|TestExecute_EmptyAfter_Rejected)$' -count=1 -v
--- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected (0.05s)
    --- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/hardlink_registry/dry_run (0.00s)
    --- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/hardlink_registry/real (0.00s)
    --- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/case_variant_registry/dry_run (0.00s)
    --- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/case_variant_registry/real (0.01s)
    --- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/symlink_registry/dry_run (0.02s)
    --- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/symlink_registry/real (0.00s)
    --- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/absent_log_fallback/dry_run (0.00s)
    --- PASS: TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/absent_log_fallback/real (0.00s)
(TestExecute_EmptyAfter_Rejected 서브테스트 2개, TestExecute_RuleFileIsRegistryOrLog_Rejected 서브테스트 6개도 PASS)
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.429s

$ shasum -a 256 internal/constitution/pipeline.go   → 1ef48dcc…2467 (뮤턴트 전)
[스팟 뮤턴트: 로그 절반을 filepath.Abs 문자열 비교로 교체]
$ go test ./internal/constitution/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.832s     ← 생존
$ shasum -a 256 internal/constitution/pipeline.go   → 1ef48dcc…2467 (복원 후); git status --porcelain 무출력

$ git show fe8cc9875:internal/constitution/pipeline.go | grep -n -iE 'dryRun|updateSourceFile'
121:	approved, err := p.HumanOversight.Approve(proposal, dryRun)
133:	if dryRun {
196:	if err := updateSourceFile(sourceFilePath, rule.Anchor, proposal.After); err != nil {

$ go test ./internal/constitution/ -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.807s	coverage: 88.4% of statements
$ go vet ./internal/constitution/ → vet_exit=0 ; GOOS=windows GOARCH=amd64 go vet ./internal/constitution/ → win_vet_exit=0
$ golangci-lint run ./internal/constitution/...
0 issues.
$ go test ./internal/spec/ -run '^TestLinter_AC08_DanglingRuleReference$' -count=1 -v
--- PASS: TestLinter_AC08_DanglingRuleReference (0.46s)
ok  	github.com/modu-ai/moai-adk/internal/spec	0.818s
$ git diff --stat e5a18feae..b36f50c4c -- internal/cli
(무출력)
```

#### Baseline-attribution

- 위 측정은 모두 이번 실행에서 트리 `b36f50c4c`를 대상으로 했으며, 출력은 붙인 그대로다.
- 기록된 뮤턴트 7개의 결과는 `.moai/reports/t659/run/guards/mutants-f1-summary.txt`와 `M-*-f1.txt`를 인용했다(트리 `bca8cf96a`; 그 뒤 `393a4af32`, `b36f50c4c`는 문서·증거만 바꿨으므로 Go 코드는 같다). 직접 실행한 뮤턴트는 1개다.
- RED 출력은 `f1-red.txt`를 인용했고, 커밋 순서는 커밋 그래프로 직접 확인했다.

#### Gaps

- `internal/cli` 테스트는 실행하지 않았다(지시; 델타 무관 확인).
- 로그 별칭 거부는 테스트가 없다. 코드 판독과 공유 함수로만 확인했다(D2).
- 승인 대기 중 별칭 생성·제거(§E.3 항목 2)는 관측하지 않았다.
- Windows 실행은 없다(vet 컴파일만). NTFS에서의 `os.SameFile` 동작 판정은 CI 몫이다.
- 트리에서 빌드한 바이너리로 `moai spec lint`를 돌리지 않았다(§2.2 Gap, 이전과 같음).
- ENOENT가 아닌 Stat 오류(권한 오류 등)의 폴백 경로는 테스트도 관측도 없다.

#### Residual-risk

- 로그 별칭 경로가 나중에 레지스트리 경로와 다른 비교로 갈라지면 그 회귀를 잡을 테스트가 없다(D2).
- 부재 파일의 별칭은 G-B를 통과하고 게이트 뒤에 거부된다. 쓰기는 없지만 승인 뒤 거부되는 경우가 남는다(§E.3 항목 1, 선언됨).
- 이전 감사의 F2, F3, F7, F8, F9 잔여 위험은 그대로다.
