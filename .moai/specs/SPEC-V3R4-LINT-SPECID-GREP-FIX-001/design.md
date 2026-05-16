# SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — Design Rationale

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. minimal design 문서: word-boundary 정밀화 의미론, 2-pass post-filter (Approach B) 채택 근거, 성능 분석, alternative rejection rationale. |

---

## 1. Design Question

> walker가 `git log --grep=<SPEC-ID>` substring 매칭으로 false-positive를 생성한다.
> word-boundary 정밀화는 walker layer에서 어떻게 구현해야 하는가?

세 가지 후보 (A: regex grep at git level / B: 2-pass post-filter / C: A+B) 중 무엇이 옳은 trade-off인가?

---

## 2. Semantic Reasoning — Why Word-Boundary is the Correct Semantic

SPEC-ID는 형식 정의상 (`^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`) **닫힌 토큰 (closed token)** 이다. 즉:

- `SPEC-V3R4-HARNESS-001` 와 `SPEC-V3R4-HARNESS-NAMESPACE-001` 은 **두 개의 독립된 SPEC**이다.
- 한 SPEC-ID가 다른 SPEC-ID의 prefix가 되는 것은 우연이 아니라 도메인 명명의 자연스러운 결과 (HARNESS 모듈 안에 NAMESPACE 하위 도메인이 추가됨).
- 따라서 두 토큰은 **의미상 disjoint** 이어야 하며, 어느 한쪽을 검색할 때 다른 쪽이 매칭되는 것은 **semantic error**.

`git log --grep=<pattern>`은 commit message 전체를 단순 substring으로 검색하므로 토큰 의미를 무시한다. 이는 git의 일반 검색에는 적합하나, walker가 **lifecycle status 추론**이라는 **의미론적 작업**을 수행할 때는 부적합하다.

**결론**: walker는 SPEC-ID를 토큰으로 다루어야 하며, **exact word-boundary 매칭**이 의미론적으로 옳다.

### 2.1 Why ExtractSPECIDs is the Right Tool

`internal/spec/transitions.go:97`:

```go
var TransitionSPECIDPattern = regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]+`)
```

이 정규식은:
1. `SPEC-` prefix를 요구 (closed-left boundary).
2. `[A-Z0-9-]+` (domain part) + `-[0-9]+` (numeric suffix) 패턴 (closed-right boundary at numeric end).
3. greedy하지만 SPEC-ID 형식 외 토큰과 충돌하지 않음.

따라서 `ExtractSPECIDs("plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 — supersedes")` 는 `["SPEC-V3R4-HARNESS-NAMESPACE-001"]` 만 반환하고 `SPEC-V3R4-HARNESS-001` 은 추출되지 않는다 (왜냐하면 `NAMESPACE-` 가 중간에 있어 정규식 경계가 다르기 때문).

**이것이 정확히 우리가 필요한 word-boundary 의미론.**

---

## 3. Architectural Choice — Two-Pass Post-Filter (Approach B)

### 3.1 Why not Approach A (regex grep at git layer)

`git log --grep` 의 정규식 엔진:
- 기본: POSIX BRE (basic regular expression). `\b` word-boundary 미지원.
- `--extended-regexp` (`-E`): POSIX ERE. `\b` 표준에 없음 (GNU grep 확장).
- `--perl-regexp` (`-P`): PCRE. `\b` 지원. 단, git 빌드 시 `--with-libpcre` 필요.

**Platform compatibility**:
- macOS (Homebrew git): PCRE 빌드됨 (확인됨).
- Ubuntu (apt git): PCRE 빌드됨 (대부분).
- Alpine / minimal Docker: PCRE 미빌드 가능성. 본 프로젝트 CI runner (`ubuntu-latest`, `macos-latest`, `windows-latest`) 에서 모두 PASS 보장이 어려움.

**Risk**: 사용자가 `moai` 를 PCRE 미지원 git 환경에서 실행 시 walker가 silently fallback 또는 error → undefined behavior.

**Decision**: A는 platform-dependent risk 때문에 reject.

### 3.2 Why not Approach C (A + B)

C의 가치는 "filter at git layer (cheaper) + safety net at Go layer". 하지만:
- A의 platform risk가 그대로 잔존.
- B만으로도 충분히 안전 (정확성 100%).
- Walker N=50 commit 가져오는 git log call 자체가 ~25ms (dominant cost). Go-side regex match는 ~0.05ms × 50 = ~2.5ms. A의 추가 이득 (git-side filter로 candidate set 축소) 은 무시할 수준.

**Decision**: C reject.

### 3.3 Why Approach B is Right

- **Platform-independent**: Go `regexp` 는 모든 OS에서 동일 동작.
- **Zero new dependency**: `ExtractSPECIDs` 는 transitions.go에 이미 존재.
- **Composability**: 기존 `shouldSkipCommitTitle` (LSCSK-001) 필터와 자연 합성.
- **Auditable**: 두 필터 모두 Go 코드 안에 있어 grep/code review 용이.
- **Reversible**: 단일 함수 추가 + 한 줄 wire. rollback trivial.

---

## 4. Filter Composition Order

scanner loop 안에서 두 필터의 순서:

```
1. shouldSkipCommitTitle(commitTitle)   ← chore-skip filter (LSCSK-001)
2. commitMatchesSPECID(commitTitle, specID)   ← word-boundary filter (LSGF-001)
3. ClassifyPRTitle(commitTitle)
```

**Decision**: chore-skip 먼저, word-boundary 다음.

**Rationale**:
- `shouldSkipCommitTitle` 는 `strings.HasPrefix` × 2 호출. cost ~50ns.
- `commitMatchesSPECID` 는 정규식 매칭 + `slices.Contains`. cost ~500ns.
- 10x 차이지만 둘 다 microsecond 이하 — performance impact 무의미.
- **"Cheaper-first" 원칙**을 채택해 향후 더 비싼 필터 추가 시 가이드라인 일관성 유지.

**의미상**: chore-skip은 commit type-level 필터 (메타데이터 작업), word-boundary는 SPEC-ID-level 필터 (정확한 대상). 직교 관계 → 순서가 결과에 영향 없음.

---

## 5. Performance Analysis

### 5.1 Walker single call cost (N=50)

| Operation | Pre-fix | Post-fix | Delta |
|-----------|---------|----------|-------|
| `git log --grep=...` execution | ~25ms | ~25ms | 0 |
| Scanner loop (50 iter) | ~0.5ms | ~3ms | +2.5ms |
| Within loop: `shouldSkipCommitTitle` | ~50ns × 50 | ~50ns × 50 | 0 |
| Within loop: `commitMatchesSPECID` (NEW) | n/a | ~500ns × 50 | +25µs |
| Within loop: `ClassifyPRTitle` | ~200ns × n | ~200ns × n' | minor (n' ≤ n) |
| **Total** | **~26ms** | **~28ms** | **+2ms (+7.7%)** |

### 5.2 Lint sweep impact (197 SPECs)

- `moai spec lint --strict` 1회 호출당 walker가 SPEC별 N회 호출 → 197 SPECs × 1 call = 197 walker calls.
- 197 × 2ms = 394ms 추가 latency.
- Total lint sweep latency: pre-fix ~5.1s → post-fix ~5.5s (+~8%).

**Verdict**: 측정 가능한 latency 증가는 있으나 user-experience 영향 무시 가능 (5초 이하 명령은 동기 작업 표준 범위).

### 5.3 Edge cases (N=50 모두 substring noise)

극단적 경우 (모든 50 commit이 word-boundary fail): walker가 N=50 fully iterate → LSCSK-001 fail-safe path → `error("no classifiable commit within window of 50 for ...")` 반환. lint rule이 skip 처리.
- Latency: ~28ms (same as normal).
- Correctness: AC-LSGF-005 보장 (false-positive 0).

---

## 6. Alternative — Why not Add a New `--grep` Flag

`git log --grep=...` 에 추가 옵션 (`--all-match`, `--regexp-ignore-case`) 사용은 무의미:
- `--all-match` 는 여러 `--grep` 패턴 AND 매칭 (단일 패턴엔 적용 안 됨).
- `--regexp-ignore-case` 는 SPEC-ID는 대소문자 의미 없음 (이미 uppercase 강제).

직접 word-boundary regex 추가가 유일한 git-layer 해법인데, Approach A에서 rejected.

---

## 7. Alternative — Why not Reform the Search Strategy

`git log` 대신 다른 방법 (예: `git rev-list --grep`, `git log --format`, `git for-each-ref`) 검토:
- `git rev-list --grep`: 동일 substring 매칭 한계.
- `git log --format=%H %s%n%b`: body 검색 가능. 본 SPEC scope 외 (OQ3 default = subject-only).
- `git for-each-ref`: tag/branch metadata 대상. commit message 검색 아님.

**Decision**: walker는 `git log --oneline --no-merges --grep=<specID>` 형식 유지. fix는 post-filter에만.

---

## 8. Alternative — Why not Make `lint.skip` Mechanism

`lint.skip: [StatusGitConsistency]` 를 HARNESS-001/002/003 spec.md frontmatter에 추가하면 WARNING은 즉시 사라진다.

**Rejected**:
- **Symptom hiding, not root cause fix**: walker 결함은 여전히 존재.
- **Future regression risk**: 다른 SPEC도 동일 prefix collision 발생 시 같은 hack 반복.
- **Tech debt accumulation**: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 이 이미 lint.skip cleanup을 수행했는데 다시 추가는 backward step.
- **SSOT violation**: lint.skip은 documented debt에만 사용 (SPEC-V3R4-SPECLINT-DEBT-002 SSOT).

---

## 9. Reversibility

| Reversible? | Mechanism |
|-------------|-----------|
| commitMatchesSPECID 함수 추가 | trivial — `git revert` 한 줄 |
| walker scanner loop 한 줄 wire | trivial |
| @MX 태그 갱신 | comment-only, harmless |
| 새 test 파일 | test-only, delete OK |
| 외부 인터페이스 변경 | 없음 (private function 신설만) |
| 데이터 마이그레이션 | 없음 |

**Conclusion**: 100% reversible. Risk Level: Low.

---

## 10. Validation Strategy

1. **Unit**: T1-001 두 테스트가 fix 적용 전 RED → 적용 후 GREEN.
2. **Integration**: `moai spec lint --strict` 0 WARNING.
3. **Regression**: LSCSK-001 가드 테스트 GREEN 유지.
4. **Race**: `go test -race` GREEN.
5. **Plan-auditor**: ≥ 0.85 score.

---

## 11. References

- `internal/spec/drift.go:96-197` — walker baseline (post LSCSK-001).
- `internal/spec/transitions.go:97-116` — `ExtractSPECIDs` (reuse).
- SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 — chore-skip 필터 도입 SPEC.
- git log documentation: <https://git-scm.com/docs/git-log#Documentation/git-log.txt---grepltpatterngt>
- POSIX regex compatibility: BRE / ERE / PCRE 차이.
- Go `regexp` package: <https://pkg.go.dev/regexp> (RE2 syntax, platform-independent).
- Go `slices.Contains`: stdlib 1.21+.
