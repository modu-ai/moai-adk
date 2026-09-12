# SPEC Review Report: SPEC-MOAI-GATEWAY-001

Iteration: 7 — 승인된 M1 형태 보정과 G6-A1 변경분만 재감사
Verdict: PASS
Overall Score: 1.00 (이번 두 수정의 변경분 점수; Tier L 기준 0.85)

Reasoning context ignored per M1 Context Isolation.

## Claim

0.8.0의 `stream` 키 부재 허용은 보존된 원본 캡처와 일치한다. JSON `false`는 별도 양성 대조군으로 남고 `null`·문자열 `"false"`·`true`는 음성으로 구분된다. G6-A1의 프로세스 env 픽스처는 정리 집합 14키 전부와 별도 `Z_AI_API_KEY`를 포함한다. 두 수정과 직접 회귀에서 차단 결함을 찾지 못했다.

이 PASS는 문서 보정의 판정이다. 제품 인식기·M0 인증·새 모델 LIVE 시험이 통과했다는 뜻이 아니다. 읽기 전에 설정한 실패 가설은 잘못된 truthiness 완화, 양성·음성 대조군 역전, 원본과 다른 서술, 14키 픽스처 누락이었다.

## Must-Pass Results

- **MP-1 PASS:** 현재 REQ 정의 26개가 001~026으로 연속이며 live REQ/AC는 24/24다(E3). 이번 수정은 번호를 추가하거나 제거하지 않았다.
- **MP-2 PASS:** 수정된 REQ-MG-023의 When 의무와 네 조건(`spec.md:675-687`)은 요구사항 계층이다. AC-MG-003·018은 Given-When-Then 검증 계층이며 명시된 값·계수로 판정한다(`acceptance.md:45-77`, `:263-287`).
- **MP-3 PASS:** 현재 frontmatter를 직접 읽었으며 버전은 `"0.8.0"`, 상태는 `draft`, canonical 필드가 유지된다(`spec.md:2-14`). 현재 lint는 E2.
- **MP-4 N/A:** 단일 Go 저장소 코어의 요청 판정·env fixture 수정이다. 범용 언어 template은 변경 범위가 아니다.
- **MP-5·MP-6:** 이번 두 변경에는 형제 SPEC 상태 변경이나 syscall 변경이 없다. iter6에서 판정한 경계를 이번에 다시 전체 감사하지 않았다. 새 cross-SPEC/platform 주장을 PASS 근거로 추가하지 않는다.
- **MP-7 PASS:** 현재 `plan.md`와 `research.md`의 clarification 표지는 각각 0개(E3). M0 INCONCLUSIVE와 향후 LIVE 시험 제한은 명시된 실행 게이트로 남는다.
- **MP-8 N/A:** 두 변경 대상 AC는 release-blocking RED-now 셀을 도입하지 않았다. 인용된 과거 캡처 파일을 read-only로 다시 파싱했으며, LIVE 캡처 명령이나 제품 시험을 재실행한 것으로 취급하지 않는다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 1.00 | 1.00 | `design.md:305-323`: 키 부재·JSON false와 null·문자열·true를 명확히 구분 |
| Completeness | 1.00 | 1.00 | `spec.md:683-687`, `plan.md:105-125`, `research.md:982-999`: 조건·원본 근거·중단 기준·증거 보존이 연결됨 |
| Testability | 1.00 | 1.00 | `acceptance.md:59-77`의 아홉 음성 변형·두 양성 대조군·실제 turn 대조군; `:263-287`의 14키 표지 및 개별 누락 뮤턴트 |
| Traceability | 1.00 | 1.00 | REQ-MG-023→AC-MG-003, REQ-MG-021→AC-MG-018; 조건과 fixture 목록의 일치는 E1·E3 |

변경되지 않은 전체 SPEC의 점수를 새로 계산한 것이 아니다. 조화평균은 1.00이며 iter6 0.92에서 회귀하지 않는다.

## Defects Found

No defects found in the authorized delta.

## Regression Check

- **M1-B1 RESOLVED — 형태 계약:** 원본 `request-003.json`은 stream 키가 없고 나머지 세 조건이 참이다(E1). `spec.md:683-687`, `design.md:305-323`, `acceptance.md:47-48`의 현행 조건은 모두 이 원문을 허용한다.
- **대조군 역전 없음:** 기존 stream 부재 음성 변형은 문자열 false 음성 변형으로 바뀌었다(`acceptance.md:62`). stream 부재와 명시 false는 둘 다 양성 대조군이며 null/true는 계속 음성이다(`:61-75`). 실제 캡처 네 건도 음성 대조군으로 요구한다(`:76-77`).
- **과도한 실측 주장 없음:** `design.md:323`은 명시 false를 이번에 관측한 값이 아닌 기존 계약의 허용값으로 구분한다. E1의 파생 변형 결과를 실제 client가 보낸 값으로 오인하지 않았다.
- **G6-A1 RESOLVED:** `acceptance.md:263-269`의 상속 표지 집합은 14키 전부와 Z_AI_API_KEY다. 기존 14키 목록과 집합 비교 결과 누락 0개다(E3). GLM 저장소 값·슬롯 후주입·순서 판정은 `:270-289`에 유지된다.
- **역사와 현재 구분:** `research.md:984-999`는 0.7.0 당시 BLOCKED 결과와 Sonnet 4.5 출력을 역사적 증거로 보존한다. `plan.md:91-94`는 M0 INCONCLUSIVE를 유지하며 2026-09-11 19:00 Asia/Seoul 이후 Opus 5·Sonnet 5 후속 시험을 요구한다. 옛 보고서의 BLOCKED를 소급 PASS로 고치지 않았다.

## Evidence

### E1 — 원본 다섯 건 재파싱과 조건 대조

이번 감사에서 `python3 - <<'PY'`로 보존된 JSON 파일을 읽고 다음 판정을 적용했다. 이는 명세 조건의 기계적 대조이며 제품 인식기 구현이 아니다.

```python
def match(d):
    return ('stream' not in d or d['stream'] is False) and type(d.get('max_tokens')) is int and d['max_tokens']==1 and isinstance(d.get('messages'),list) and len(d['messages'])==1 and d['messages'][0].get('role')=='user' and ('tools' not in d or d['tools']==[])
```

입력 디렉터리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/`. `m1-raw/request-*.json`을 순회하고 003의 SHA-256 및 stream 한 필드만 바꾼 네 변형을 출력했다. 관측 stdout(exit 0):

```text
request-001.json stream_present=True max_tokens=32000 roles=['user'] match=False
request-002.json stream_present=True max_tokens=32000 roles=['user'] match=False
request-003.json stream_present=False max_tokens=1 roles=['user'] match=True
request-004.json stream_present=True max_tokens=32000 roles=['user'] match=False
request-005.json stream_present=True max_tokens=32000 roles=['user', 'system', 'assistant', 'user', 'system'] match=False
request-003 sha256=0c65416916579d77028e64d1fbbec322055f07b3464e9b3aa1ee0ebe06c133e9
stream_false match=True
stream_null match=False
stream_string_false match=False
stream_true match=False
client_version=2.1.268 (Claude Code) suppression_flags_present=[]
```

### E2 — 구조 lint

현재 WT를 workdir로 지정하여 실행했다.

```text
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
```

Exit 0. 이 바이너리는 이전에 부모 세션이 같은 HEAD에서 빌드했으며 이번 감사에서는 lint만 직접 실행했다.

### E3 — 정리 키 집합·번호·표지·문서 해시

실행한 read-only Python 검사는 AC-MG-018의 정의 목록과 프로세스 env Given 목록에서 backtick으로 감싼 환경변수 이름을 모으고, `{OPUS,SONNET,HAIKU,FABLE}` 표기를 네 키로 확장해 집합을 비교했다. REQ/AC 정의는 `^\*\*REQ-MG-(\d{3})\*\*`와 `^\*\*AC-MG-\d{3}\*\* \(`로 세었다. 관측 stdout(exit 0):

```text
cleanup_keys=14 fixture_keys=15
missing=[] extra=['Z_AI_API_KEY']
REQ_definitions=26 sequential=True
live_REQ=24 live_AC=24
plan.md clarification_count=0
research.md clarification_count=0
spec.md 9825c4826a8465c345a8315592ebc47d82a3cc6e981af93b16b3ea215d52f171
plan.md a31d05c2fded0571931408f212b3d9b478b1574759249c360d29929ad3a3f5f0
acceptance.md 50ba4c5e9a78953689f94072e4f4df999ab92530c28f0ee449699ba0487d1c63
design.md 69dcb37aea386edb2bdef1c6d9fc02ca38b382fb50de66dce88729c086d55ac3
research.md fab7472b60710acaf423f0d36025759e936aaa35f12aaf6dcc6655e5c3c912b7
progress.md f3ade06e14bc8653524d44efde2fbc70e589a566d802e19a7296702ecb8cacb7
```

## Baseline-attribution

`git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse --short HEAD` → `81c1d58f9`, exit 0. 대상 SPEC은 0.8.0이며 여섯 최종 문서 내용은 E3의 해시로 고정했다. 이번 감사에서 LIVE 시험·제품 코드 수정·SPEC 수정·Git 변경은 수행하지 않았다. 새로 작성한 파일은 이 보고서뿐이다.

`git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified check-ignore .moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter7.md` → stdout 빈 출력, exit 1. 보고서 경로는 ignore되지 않는다.

## Gaps

- M0는 INCONCLUSIVE다. 정상 OAuth 응답·refresh·정확한 새 모델 접근 권한을 확인하지 않았다.
- 원본을 이번에 다시 파싱했으나 client를 다시 실행하지 않았다. 이 캡처는 Claude Code 2.1.268·당시 Sonnet 4.5 조건이며 향후 Opus 5·Sonnet 5 시험을 대체하지 않는다.
- 추적 fixture 디렉터리는 아직 없었다. `plan.md:113-118`에 따라 비식별 처리 이력·원본/fixture 해시를 연결해 고정하는 작업은 남아 있다.
- 모든 가능한 비스트리밍 요청과의 구분, 제품 gateway의 3사 연속 전환, 제품 인식기 뮤턴트 시험은 실행하지 않았다.

## Residual-risk

다른 작업 요청이 같은 네 조건을 만족할 가능성은 현재 다섯 캡처로 배제할 수 없다. 그런 요청이 관측되면 M1 중단 조건이 다시 적용된다. 소스 형태 보정 PASS만으로 M1 전체나 인증 게이트를 완료 처리해서는 안 된다.

## Recommendation and Iteration History

iter6 PASS 0.92의 선택 개선 G6-A1은 이번에 해소되었다. 0.7.0 M1 캡처의 BLOCKED 이력은 보존하며 그 원인이던 인식 조건의 문서 보정만 이번 iter7 PASS 1.00으로 판정한다. 추적 fixture와 대조 시험을 준비하는 다음 단계로 진행할 수 있다. M0 양성 판정·LIVE 시험 시각·새 모델 요구는 그대로 유지한다.
