---
id: SPEC-V3R6-UPDATE-NOISE-001
title: "SPEC-V3R6-UPDATE-NOISE-001 — Acceptance Criteria"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-22
author: manager-spec
priority: P2
phase: "v3.6.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "v3r6, ux, update, noise-suppression, idempotency, state-file, acceptance"
tier: S
---

# SPEC-V3R6-UPDATE-NOISE-001 — Acceptance Criteria

본 문서는 `./spec.md` 의 REQ-UN-001~011 에 대응하는 12개 binary ACs 와 각 AC 의 verification 명령 + 기대 출력을 정의한다. 모든 검증은 `t.TempDir()` 기반 단위 테스트 또는 실 환경 (`~/moai/mo.ai.kr`) smoke 로 binary PASS/FAIL 판정 가능하다.

---

## AC 일람

| AC ID | REQ 매핑 | 한 줄 요약 | Verification 방법 |
|-------|---------|----------|------------------|
| AC-UN-001 | REQ-UN-001 | ack ledger 스키마 생성 | `cat .moai/state/reserved-acknowledged.json` |
| AC-UN-002 | REQ-UN-002 | 2회차 reserved warning silent | `moai update` ×2 + stderr grep |
| AC-UN-003 | REQ-UN-004 | hash drift 시 재출현 | 파일 수정 후 `moai update` |
| AC-UN-004 | REQ-UN-006 | merge-history 스키마 생성 | `cat .moai/cache/merge-history.json` |
| AC-UN-005 | REQ-UN-007 | 1·2회차 fallback silent | `moai update` ×2 + stderr grep |
| AC-UN-006 | REQ-UN-008 | 3회차 advisory 한 줄 | `moai update` ×3 + stderr grep |
| AC-UN-007 | REQ-UN-005 + REQ-UN-010 | `--verbose` 모두 재출현 | `moai update --verbose` |
| AC-UN-008 | REQ-UN-009 | merge 성공 시 counter reset | mergeYAML3Way success path |
| AC-UN-009 | REQ-UN-011 | 손상 ledger 자동 복구 | 손상 JSON 작성 후 `moai update` |
| AC-UN-010 | REQ-UN-001 + REQ-UN-006 | state 파일 `.gitignore` 보호 | `git check-ignore` |
| AC-UN-011 | REQ-UN-001~005 unit | reserved_ack_test.go PASS | `go test -run TestReservedAck` |
| AC-UN-012 | REQ-UN-006~010 unit | merge_history_test.go PASS | `go test -run TestMergeHistory` |

총 12 ACs. 모두 PASS / FAIL binary.

---

## AC 상세 — Verification commands + expected outputs

### AC-UN-001 — Reserved ack ledger 스키마 생성

**REQ mapping**: REQ-UN-001 (ledger 파일 스키마)

**Verification command** (실 환경):

```bash
cd ~/moai/mo.ai.kr
moai update
cat .moai/state/reserved-acknowledged.json
```

**Expected output** (예시 — sha256 / 타임스탬프는 동적):

```json
{
  "tokens.json": {
    "sha256": "<64-hex-char hash>",
    "acknowledged_at": "2026-05-23T<HH:MM:SS>Z"
  },
  "components.json": {
    "sha256": "<64-hex-char hash>",
    "acknowledged_at": "2026-05-23T<HH:MM:SS>Z"
  }
}
```

**PASS 조건**: JSON parse 가능 + 각 entry 가 `sha256` (64-hex-char string) + `acknowledged_at` (RFC3339 string) 키를 가진다.

**Programmatic check**:

```bash
jq -e '. | type == "object" and (to_entries | all(.value | has("sha256") and has("acknowledged_at")))' \
  ~/moai/mo.ai.kr/.moai/state/reserved-acknowledged.json
```

---

### AC-UN-002 — 2회차 reserved warning silent

**REQ mapping**: REQ-UN-002 (ack 조회 분기)

**Verification command**:

```bash
cd ~/moai/mo.ai.kr
moai update 2>/tmp/update1.stderr
moai update 2>/tmp/update2.stderr
grep -c "warning: reserved filename" /tmp/update2.stderr
```

**Expected output** (stdout): `0`

**Additional check** (첫 회차는 정상 emit 했는지):

```bash
grep -c "warning: reserved filename" /tmp/update1.stderr
```

기대: `>= 2` (tokens.json + components.json 각각)

**PASS 조건**: `/tmp/update2.stderr` 에 "warning: reserved filename" 문자열이 **0회** 등장 AND `/tmp/update1.stderr` 에는 **2회 이상** 등장.

---

### AC-UN-003 — Hash drift 시 warning 재출현

**REQ mapping**: REQ-UN-004 (hash mismatch trigger)

**Verification command**:

```bash
cd ~/moai/mo.ai.kr
moai update 2>/dev/null  # ack baseline
echo "// modified $(date)" >> .moai/design/tokens.json
moai update 2>/tmp/update_drift.stderr
grep "warning: reserved filename: tokens.json" /tmp/update_drift.stderr
```

**Expected output** (stdout):

```
warning: reserved filename: tokens.json (preserved; rename to use canonical templates)
```

**Additional check** (ledger 의 sha256 이 갱신되었는지):

```bash
# 이전 sha256 캡처 후 비교 (테스트 환경에서)
jq -r '.["tokens.json"].sha256' .moai/state/reserved-acknowledged.json
```

기대: 수정 전 sha256 과 다른 값.

**PASS 조건**: drift 후 update 에서 warning 1회 emit + ledger sha256 이 갱신된다.

---

### AC-UN-004 — Merge-history 스키마 생성

**REQ mapping**: REQ-UN-006 (ledger 파일 스키마)

**Verification command**:

```bash
cd ~/moai/mo.ai.kr
moai update  # quality.yaml 3-way fallback 가정
cat .moai/cache/merge-history.json
```

**Expected output** (예시):

```json
{
  "quality.yaml": {
    "fallback_count": 1,
    "last_failed_at": "2026-05-23T<HH:MM:SS>Z"
  }
}
```

**Programmatic check**:

```bash
jq -e '. | type == "object" and (to_entries | all(.value | has("fallback_count") and has("last_failed_at")))' \
  ~/moai/mo.ai.kr/.moai/cache/merge-history.json
```

**PASS 조건**: JSON parse 가능 + 각 entry 가 `fallback_count` (int) + `last_failed_at` (RFC3339 string).

---

### AC-UN-005 — 1·2회차 fallback silent

**REQ mapping**: REQ-UN-007 (silent until threshold)

**Setup recipe** (재현 환경 준비 — `/tmp/test-project-noise` 가 존재하지 않으면 다음과 같이 부트스트랩):

```bash
rm -rf /tmp/test-project-noise
moai init /tmp/test-project-noise --yes
# Deep-edit quality.yaml so every `moai update` triggers a 3-way merge failure:
printf '\n# Deep user modifications to force 3-way merge fallback:\n' >> /tmp/test-project-noise/.moai/config/sections/quality.yaml
printf 'custom_keys:\n  unmergeable_field: "user-only-value"\n' >> /tmp/test-project-noise/.moai/config/sections/quality.yaml
```

대안: 본 AC 는 단위 검증으로도 충분히 커버된다 — `go test -run TestMergeHistory_FirstTwoFallbacksSilent ./internal/cli/` (AC-UN-012 와 중첩 검증).

**Verification command** (테스트 환경에서 quality.yaml 을 의도적으로 깊이 수정 → 매 update 마다 3-way fallback 발동 보장):

```bash
cd /tmp/test-project-noise
moai update 2>/tmp/fb1.stderr
moai update 2>/tmp/fb2.stderr
grep -c "3-way merge failed\|falling back to 2-way\|hint:" /tmp/fb1.stderr
grep -c "3-way merge failed\|falling back to 2-way\|hint:" /tmp/fb2.stderr
```

**Expected output** (stdout): `0` 과 `0`

**PASS 조건**: 첫 두 회차 모두 stderr 에 fallback 관련 키워드 (`3-way merge failed`, `falling back to 2-way`, `hint:`) 가 **0회** 등장.

---

### AC-UN-006 — 3회차 advisory 한 줄

**REQ mapping**: REQ-UN-008 (threshold-triggered advisory)

**Verification command**:

```bash
cd /tmp/test-project-noise
moai update 2>/dev/null  # fb1
moai update 2>/dev/null  # fb2
moai update 2>/tmp/fb3.stderr
grep "^hint: 'moai update -c' to resync templates for quality.yaml$" /tmp/fb3.stderr
```

**Expected output** (stdout):

```
hint: 'moai update -c' to resync templates for quality.yaml
```

**Additional check** (정확히 한 줄):

```bash
grep -c "^hint: 'moai update -c'" /tmp/fb3.stderr
```

기대: `1`

**Additional check** (4회차 silent — once-per-relPath):

```bash
moai update 2>/tmp/fb4.stderr
grep -c "^hint: 'moai update -c'" /tmp/fb4.stderr
```

기대: `0`

**PASS 조건**: 3회차에서 advisory **정확히 1줄** + 4회차에서 **0줄** + wording 이 exact match (소문자 `hint`, single quote 로 둘러싼 `moai update -c`, `to resync templates for <relPath>`).

---

### AC-UN-007 — `--verbose` 모든 warning 재출현

**REQ mapping**: REQ-UN-005 (verbose bypass reserved) + REQ-UN-010 (verbose bypass merge)

**Verification command**:

```bash
cd ~/moai/mo.ai.kr
moai update 2>/dev/null  # 모든 ack baseline
moai update --verbose 2>/tmp/verbose.stderr
grep -c "warning: reserved filename" /tmp/verbose.stderr
grep -c "3-way merge failed\|falling back to 2-way" /tmp/verbose.stderr
```

**Expected output** (stdout 1): `>= 2` (tokens.json + components.json 양쪽 재출현)

**Expected output** (stdout 2): `>= 1` (quality.yaml 또는 다른 3-way fallback 파일)

**PASS 조건**: `--verbose` 실행 시 두 종류 warning 이 모두 stderr 에 다시 등장. Ack ledger 와 merge-history 는 비-verbose 와 동일하게 갱신 (REQ-UN-005·010 명시).

---

### AC-UN-008 — Merge 성공 시 counter reset

**REQ mapping**: REQ-UN-009 (success path counter reset)

**Verification command** (unit test 우선 — 실 환경 재현 어려움):

```bash
cd /Users/goos/MoAI/moai-adk-go
go test -run TestMergeHistory_SuccessResetsCounter ./internal/cli/ -v
```

**Expected test output**:

```
=== RUN   TestMergeHistory_SuccessResetsCounter
--- PASS: TestMergeHistory_SuccessResetsCounter (0.00s)
PASS
```

**Test scenario** (간략):

1. `recordMergeFallback(_, "config.yaml", false, false, &buf)` 2회 호출 — `FallbackCount = 2` 확인
2. `recordMergeFallback(_, "config.yaml", true, false, &buf)` 1회 호출 — `FallbackCount = 0` 으로 reset 확인

**PASS 조건**: `TestMergeHistory_SuccessResetsCounter` PASS.

---

### AC-UN-009 — 손상 ledger 자동 복구

**REQ mapping**: REQ-UN-011 (corruption recovery)

**Verification command**:

```bash
cd ~/moai/mo.ai.kr
moai update 2>/dev/null  # baseline ledger 생성
echo "this is not valid JSON {{{" > .moai/state/reserved-acknowledged.json
moai update 2>/tmp/corrupt.stderr
echo "exit_code=$?"
jq -e 'type == "object"' .moai/state/reserved-acknowledged.json
```

**Expected output**:

- `moai update` exit code: `0` (success)
- 손상 ledger 는 빈 객체 `{}` 또는 정상 entries 로 재초기화
- `jq -e 'type == "object"'` exit code: `0`

**Additional check** (warning 이 재출현해야 함 — 손상은 fail-safe 으로 "지나친 경고" 방향):

```bash
grep -c "warning: reserved filename" /tmp/corrupt.stderr
```

기대: `>= 2` (모든 reserved 충돌 재경고).

**PASS 조건**: update 가 정상 종료 + ledger 가 valid JSON 으로 재구성 + reserved warning 재출현 (fail-safe 검증).

---

### AC-UN-010 — `.gitignore` 보호 검증

**REQ mapping**: REQ-UN-001 + REQ-UN-006 (state file gitignore — Risk R3 mitigation)

**Verification command**:

```bash
cd /Users/goos/MoAI/moai-adk-go
git check-ignore -v .moai/state/reserved-acknowledged.json
git check-ignore -v .moai/cache/merge-history.json
```

**Expected output**:

```
.gitignore:<line>:<pattern>	.moai/state/reserved-acknowledged.json
.gitignore:<line>:<pattern>	.moai/cache/merge-history.json
```

(각 명령이 exit code 0 으로 종료하고 `.gitignore` 또는 `internal/template/templates/.gitignore` 의 매칭 패턴을 출력)

**Additional check** (template baseline 미러):

```bash
grep -nE '^\.moai/(state|cache)/' internal/template/templates/.gitignore
```

기대: 두 패턴 모두 매치 (또는 더 넓은 `.moai/state/`, `.moai/cache/` 디렉토리 패턴).

**PASS 조건**: 양쪽 `.gitignore` (로컬 + template) 에 두 state 파일이 ignore 패턴으로 등재되어 있다.

---

### AC-UN-011 — `TestReservedAck` 테스트 슈트 PASS

**REQ mapping**: REQ-UN-001~005 + REQ-UN-011 (unit test 통합)

**Verification command**:

```bash
cd /Users/goos/MoAI/moai-adk-go
go test -run TestReservedAck ./internal/cli/ -v -count=1
```

**Expected output**:

```
=== RUN   TestReservedAck_FirstOccurrenceEmitsWarning
--- PASS: TestReservedAck_FirstOccurrenceEmitsWarning (0.00s)
=== RUN   TestReservedAck_SecondCallSilent
--- PASS: TestReservedAck_SecondCallSilent (0.00s)
=== RUN   TestReservedAck_HashDriftReemits
--- PASS: TestReservedAck_HashDriftReemits (0.00s)
=== RUN   TestReservedAck_VerboseBypass
--- PASS: TestReservedAck_VerboseBypass (0.00s)
=== RUN   TestReservedAck_CorruptedLedgerRecovers
--- PASS: TestReservedAck_CorruptedLedgerRecovers (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.0XX s
```

**PASS 조건**: 5개 sub-test 모두 PASS, race condition 없음, coverage 회귀 없음.

---

### AC-UN-012 — `TestMergeHistory` 테스트 슈트 PASS

**REQ mapping**: REQ-UN-006~010 (unit test 통합)

**Verification command**:

```bash
cd /Users/goos/MoAI/moai-adk-go
go test -run TestMergeHistory ./internal/cli/ -v -count=1
```

**Expected output**:

```
=== RUN   TestMergeHistory_FirstTwoFallbacksSilent
--- PASS: TestMergeHistory_FirstTwoFallbacksSilent (0.00s)
=== RUN   TestMergeHistory_ThirdFallbackEmitsAdvisory
--- PASS: TestMergeHistory_ThirdFallbackEmitsAdvisory (0.00s)
=== RUN   TestMergeHistory_FourthFallbackSilent
--- PASS: TestMergeHistory_FourthFallbackSilent (0.00s)
=== RUN   TestMergeHistory_SuccessResetsCounter
--- PASS: TestMergeHistory_SuccessResetsCounter (0.00s)
=== RUN   TestMergeHistory_VerboseBypass
--- PASS: TestMergeHistory_VerboseBypass (0.00s)
=== RUN   TestMergeHistory_CorruptedLedgerRecovers
--- PASS: TestMergeHistory_CorruptedLedgerRecovers (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.0XX s
```

**PASS 조건**: 6개 sub-test 모두 PASS.

---

## Output gating audit (보조 검증)

다음 명령은 AC 범위는 아니지만 구현 완성도 확인용 sanity check 으로 권장된다:

```bash
# 기존 warning 문자열이 unconditional 경로에서 사라지고 verbose 분기로 이동했는지
git grep -n "falling back to 2-way" internal/cli/update.go
git grep -n "warning: reserved filename" internal/cli/design_folder.go
```

**Expected**: 각 grep 결과의 호출 컨텍스트가 verbose 분기 또는 advisory 임계 분기 내부로 이동.

```bash
# state 파일 디렉토리가 실제 생성되는지
ls -la ~/moai/mo.ai.kr/.moai/state/ ~/moai/mo.ai.kr/.moai/cache/
```

**Expected**: 두 디렉토리 모두 존재, ledger JSON 파일이 0755 부모 디렉토리에 0644 mode 로 존재.

---

## Definition of Done

본 SPEC 의 완료 조건은 다음 모두 만족이다:

- [ ] 12개 AC (AC-UN-001 ~ AC-UN-012) 모두 PASS
- [ ] `go test ./internal/cli/...` 전체 PASS, race detector PASS
- [ ] `golangci-lint run --timeout=2m` 0 NEW issues
- [ ] Cross-platform `GOOS=windows go build ./...` + `GOOS=darwin go build ./...` PASS
- [ ] 사용자 실 환경 (`~/moai/mo.ai.kr`) 에서 2회 연속 `moai update` 실행 시 reserved warning 출력 0개
- [ ] `.moai/state/` + `.moai/cache/` 가 양쪽 `.gitignore` 에 등재
- [ ] spec.md · plan.md · acceptance.md frontmatter `status: implemented`, `version: 0.2.0` 갱신
- [ ] progress.md 작성 (M0~M5 milestone evidence + AC verification 출력 캡처)

---

## Out of Scope

### Out of Scope — 본 acceptance 범위 외 검증

- SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 verification (별도 SPEC AC)
- SPEC-V3R6-UPDATE-PROGRESS-001 verification (별도 SPEC AC)
- Reserved filename 자동 정정 / rename 자동화 (user data preservation 의무)
- ack ledger 의 cross-machine 동기화 (.moai/state/ 는 local-only)
- 3-way merge 자체 알고리즘 개선 (mergeYAML3Way 동작 불변)

---

End of acceptance.md
