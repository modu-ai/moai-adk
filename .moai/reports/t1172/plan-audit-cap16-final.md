# SPEC Review Report: SPEC-CODEX-PREAPPROVAL-PROBE-001

Iteration: 기존 cap16 감사 3회 종료 뒤, 리드가 요청한 새 독립 감사 1회
Verdict: PASS — 문서의 첫 LIVE 계획 게이트와 D1–D4 수정 범위
Overall Score: 0.94

Reasoning context ignored per M1 Context Isolation. 이전 보고서의 결론은 회귀 대상을 식별하는 데만 썼고, 이번 판정은 현재 SPEC 본문에서 추출한 명령과 이번 실행 결과에 둔다.

## Claim

`f496f49e3`에 들어간 AC-CPP-004 개정은 과거 감사의 D4(아직 LIVE에서 참조하지 않은 신규 시작 검사의 원자료 누락)를 거부한다. D1(실패 시작 검사와 LIVE 연결), D2(M1-a 원본 보존), D3(무효 시도 뒤 재반출), D4와 시작 검사 절대 상한 16회의 문서 판정 의미에 이번 감사에서 미해결 blocking 결함을 찾지 못했다. 이는 실제 LIVE 실행 또는 현재 미커밋 Go 하네스의 준비 완료 판정이 아니다.

## Must-Pass Results

- MP-1 PASS — `spec.md:70-110`의 REQ-CPP-001–011은 연속이며 `spec.md:118-128`에서 모두 AC에 연결된다.
- MP-2 PASS — 판단 계층은 `spec.md:70-110`의 REQ 본문이다. 각 REQ가 `When`·`Where`·`The … shall` 형식으로 행동을 기술한다. `acceptance.md`의 Given/When/Then은 별도 검증 계층이다.
- MP-3 PASS — `spec.md:2-13`의 필수 12개 frontmatter 필드와 인용한 semver `version: "0.5.5"`를 확인했다.
- MP-4 N/A — Codex 호스트에 한정된 실험이다.
- MP-5 PASS — `rg -o 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` 결과의 외부 참조 `SPEC-CODEX-AUDIT-READONLY-001`, `SPEC-CODEX-WIRING-001`은 각각 존재하며 `status: completed`였다. 자기 참조는 `status: draft`다.
- MP-6 PASS — `rg -n -F 'syscall' spec.md`는 출력 없이 종료했다.
- MP-7 PASS — `rg -n -F '[NEEDS CLARIFICATION' plan.md acceptance.md`는 출력 없이 종료했다. Tier M에는 `research.md`가 없으며 요구 입력은 `spec.md`·`plan.md`·`acceptance.md`다.
- MP-8 N/A — `acceptance.md`에서 release-blocking으로 분류된 AC가 없다. LIVE AC와 release-blocking 분류는 다른 개념이다.

## Category Scores

| Dimension | Score | 근거 |
|---|---:|---|
| Clarity | 0.75 | `spec.md:80`, `plan.md` §D3, `acceptance.md:95-100,115-164`가 신규 시작 검사·반출 사본·LIVE 연결을 같은 순서로 설명한다. 여러 판정식을 함께 실행해야 재시도 무효 기록까지 닫힌다는 읽기 주의가 남는다. |
| Completeness | 1.00 | Tier M 세 문서, HISTORY/배경/요구사항/AC/구체적인 제외 범위가 있다(`spec.md:24,39,68,130-152`). |
| Testability | 1.00 | AC-CPP-004와 AC-CPP-014의 양성·음성 합성 증거가 아래처럼 구분된다. |
| Traceability | 1.00 | `spec.md:118-128`의 REQ 11개 모두 AC를 가리키며, `acceptance.md`에 AC-CPP-001–014와 AC-CAR-010/011이 있다. |

## Defects Found

No blocking defects found in the requested D1–D4/cap16 delta.

## Regression Check

- D1 RESOLVED — `acceptance.md:117-141`은 신규 시작 검사 자체의 `passed`, 반출 사본, 원자료를 검증하고 `:144-160`은 LIVE의 `startup_id`, 동일 픽스처·두 루트, 선행 종료, 동일 반출 ID/해시를 결속한다. 실패 시작 검사와 루트 불일치 변이가 모두 exit 1이다.
- D2 RESOLVED — `acceptance.md:108-113`은 M1-a 장부 첫 네 행과 원자료 12개의 고정 해시를 검사한다. 실제 원본을 복사한 합성 양성이 이 검사까지 통과했다. `acceptance.md:322`에서도 두 고정 해시를 재확인한다.
- D3 RESOLVED — `acceptance.md:122-126,153-164`가 각 시작 검사와 LIVE를 자기 불변 반출 사본에 묶고 마지막 유효 시도만 최신 매니페스트와 비교한다. 두 팔×두 시도 재반출과 리드 `invalidate` 합성 양성은 `true`, exit 0이다. AC-CPP-004만 독립 실행하면 `invalidate`가 없는 재시도도 `true`지만, `acceptance.md:322`의 AC-CPP-014가 앞 시도의 무효 기록과 시각을 검증한다. 판정 계약은 두 AC를 함께 요구한다.
- D4 RESOLVED — `acceptance.md:115-141`은 모든 신규 시작 검사 행을 LIVE 참조 여부와 무관하게 순회한다. 미참조 car010 행의 원자료 세 파일을 없앤 변이는 `FileNotFoundError`, exit 1이다.
- Cap16 RESOLVED — `acceptance.md:311-322`의 추출 셸에서 원본 네 행+합성 12행은 `true`, exit 0, 17행은 출력 없이 exit 1이다.

## Evidence

기준 작업트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`. 첫 실행 시 `git rev-parse --short HEAD` → `f496f49e3`. 보고 직전 `git rev-parse --short HEAD` → `c2c83f567`; `git rev-parse HEAD:.moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001/acceptance.md`와 `git rev-parse f496f49e3:.moai/specs/SPEC-CODEX-PREAPPROVAL-PROBE-001/acceptance.md`가 모두 `ccc94e2a5913ae85b7789c25acf33417c255ae96`여서 판정식 바이트는 그대로다. Go 파일 두 개는 다른 작업자가 계속 수정 중이었다.

실행 형태: `python3 - <<'PY'` 임시 드라이버가 현재 `acceptance.md`의 `### AC-CPP-004` 첫 ` ```bash ` 블록을 정규식으로 **그대로 추출**해 `bash -c`로 실행했다. `tempfile.TemporaryDirectory()`에 실제 `.moai/reports/t1172`를 복사하고, 원본 M1-a 네 장부 행·원자료 12개는 그대로 둔 채 합성 신규 행·반출 사본·LIVE 행만 보탰다. 각 사례마다 `ledger.json`을 기록하고 종료 코드·stdout·stderr 마지막 줄을 출력했다. 관측 출력 전문:

```text
good exit 0 out 'true' err ''
orphan_raw_present exit 0 out 'true' err ''
orphan_raw_missing exit 1 out '' err "FileNotFoundError: [Errno 2] No such file or directory: '.moai/reports/t1172/startup/attempt-1/car010-1790314003000000000.jsonl'"
failed_startup exit 1 out '' err 'AssertionError'
root_mismatch exit 1 out '' err 'AssertionError'
retry_reexport_with_invalidate exit 0 out 'true' err ''
retry_missing_invalidate exit 0 out 'true' err ''
```

별도 `python3 - <<'PY'` 임시 드라이버가 현재 `acceptance.md:321-323`의 AC-CPP-014 ` ```bash ` 블록을 **그대로 추출**해 `bash -c`로 실행했다. 같은 원본 네 행에 합성 판별 LIVE 두 행을 더하고, 원본 시작 검사 행의 유효한 복제를 붙여 상한 경계를 검사했다. 관측 출력 전문:

```text
positive_base startups 4 exit 0 stdout 'true' stderr ''
positive_cap16 startups 16 exit 0 stdout 'true' stderr ''
negative_cap17 startups 17 exit 1 stdout '' stderr ''
```

구조 검사 관측: `rg -o 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u`의 외부 참조 2개는 모두 `status: completed`; `rg -n -F 'syscall' spec.md`, `rg -n -F '[NEEDS CLARIFICATION' plan.md acceptance.md`, `git diff --check 3549c3e40..HEAD`는 출력이 없었다.

## Baseline-attribution

판정식은 commit `f496f49e3`에서 읽었고, 보고 직전 현재 HEAD `c2c83f567`에서도 동일한 acceptance blob `ccc94e2a5913ae85b7789c25acf33417c255ae96`로 확인했다. M1-a 원본 네 행·원자료 12개와 반출 매니페스트는 이 작업트리의 실제 파일을 임시 디렉터리로 복사했다. 그 밖의 신규 행과 원자료는 합성이다.

## Gaps

- Codex 모델 LIVE 호출을 실행하지 않았다. 실제 수락/거부나 `codex_role_audit` 접근 효과는 미측정이다.
- 동시 작업 중인 Go 하네스의 현재 동작과 첫 LIVE 직전의 preflight 통과는 검증하지 않았다. 이 문서 PASS를 실행 승인 또는 실행 완료로 읽을 수 없다.
- 원래 M1-a 행은 복사해 고정 해시 검사를 통과시켰지만, 이 감사에서 원본 비모델 시작 검사 프로세스를 다시 실행하지 않았다.
- AC-CPP-004 단독은 무효 기록을 검사하지 않는다. AC-CPP-002·014와 함께 집계해야 한다.

## Residual-risk

판정식이 합성 입력에서 의도대로 갈려도 실제 CLI 0.157.0 세션의 응답, 임시 루트 수명, 프로세스 종료 시각과 새 Go 하네스의 장부 기록이 이 스키마와 다를 수 있다. 첫 LIVE 전에는 현재 Go 코드의 독립적인 비모델 검증과 실제 실행 준비 검토가 별도로 필요하다.
