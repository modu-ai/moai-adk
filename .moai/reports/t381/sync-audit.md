# Sync 감사 — SPEC-IGNORED-EVIDENCE-CITATION-001 (카드 t381)

- 감사 트리: `.claude/worktrees/t381`, 브랜치 `WT-ignored-evidence-cite`
- 감사 시점 HEAD: `dc0956728` (작업 트리 clean, 미푸시)
- `origin/develop`: `5928095ea` (감사 중 재확인), merge base `3f03d9c36` (불변)
- 감사 도구: 이 트리에서 `make build`로 새로 만든 `./bin/moai`

---

## 판정

**PASS-WITH-DEBT** — blocking 결함 0건, optional 결함 6건.

| 차원 | 점수 | 판정 | 근거 (실행 명령 + 관측) |
|---|---|---|---|
| Functionality (40%) | 0.95 | PASS | MUST 기준 10개를 전부 직접 재실행해 PASS 확인. AC-IEC-001은 수리 이전 blob에 대해 RED 재현, AC-IEC-007은 병합 트리 존재 확인 + RED 가능성 확인으로 공허하지 않음을 입증 |
| Security (25%) | 0.95 | PASS | 편집 전량이 주석·헤더·푸터. 입력 처리·자격증명·네트워크 표면 없음. 증거 파일에 비밀값 없음(전수 판독) |
| Craft (20%) | 0.82 | PASS | 검증 설계는 강함(three-dot 기준선이 base 이동 3회를 무편집 흡수, `merge-tree --write-tree`로 미흡수 상태에서 흡수 후 검증). 인접성 게이트 미달·AC 개수 자기모순·DoD 자기서술 불일치로 감점 |
| Consistency (15%) | 0.88 | PASS | 전역 SPEC lint `0 error(s), 1096 warning(s)`, 이 SPEC 귀속 findings 0. Conventional Commits 준수. `SyncSHASlotFormat` 일시 위반 1건이 감점 요인 |

**조화평균 = 0.897** (Tier M 임계 0.80). must-pass(Functionality·Security) 양쪽 독립 통과 — 방화벽 발동 없음.

---

## AC 매트릭스 — 감사자가 직접 잰 값

| AC | 명령 | 관측값 | PASS 조건 | 판정 |
|---|---|---|---|---|
| AC-IEC-001 | `grep -l … \| xargs grep -LiE …` | 출력 없음, `exit=0` | 출력이 비어 있을 것 | PASS |
| AC-IEC-001 (동반) | `git grep -c '\.moai/state/verify' -- <5개>` | 4개 파일 각 `1`, `mcp_glm.go`는 목록에서 탈락 | 각 `≤1` | PASS |
| AC-IEC-002 | `grep -c 'still resolve at audit time'` / `grep -c 'Evidence lands in'` | `0` / `1` | `0` 그리고 `1` | PASS |
| AC-IEC-003 | `grep -oE '3667\|3480\|3072\|1024\|1\.02' \| sort -u \| wc -l` | `5` | `5` | PASS |
| AC-IEC-004 | `grep -cE '\.moai/reports/\|\.moai/state/verify' <2개>` | 각 `1` | 각 `≥1` | PASS |
| AC-IEC-005 | glob `1`, 불가 표기 `1` | `1` / `1` | 두 번째 분기(glob 유지 + 명시 표기) | PASS |
| AC-IEC-006 | `git diff --exit-code --stat origin/develop...HEAD -- <8개>` | 빈 출력 `exit=0`; 양성 대조 합계 **12** | `exit=0` + 합계 12 | PASS |
| AC-IEC-007 | 8경로 three-dot diff | 빈 출력 `exit=0` | `exit=0` | PASS |
| AC-IEC-010 | 10개 파일 `ls` / `git check-ignore` / `ls .moai/state/verify` | `exit=0` / `exit=1` / `No such file or directory` | 각각 0 / 1 / 부재 | PASS |
| AC-IEC-011 | `go build` / `go test ./internal/cli/... ./internal/hook/...` | `exit=0` / `exit=0`, `^FAIL` 0줄 | build 0 + FAIL 없음 | PASS |
| AC-IEC-012 | `grep -c 'gitignore:284'` / `grep -ci 'gitignore'` | `0` / `2` | `0` 그리고 `≥1` | PASS |

**10/10 PASS.** 기록된 증거 파일의 값과 재측정값이 전부 일치했다.

### 공허성 검사 (dispatch 요구 2·3항)

**AC-IEC-001은 RED로 갈 수 있다.** `git show origin/develop:<file>`로 수리 이전 blob 3개를 스크래치에 꺼내 같은 패턴 쌍을 걸었더니 3개 전부 위반 목록에 찍혔다(`exit=1`). 게이트가 지금 비어 있는 것은 수리 결과이지 게이트가 아무것도 못 봐서가 아니다.

**AC-IEC-007의 three-dot 형식도 재현했다.** `git merge-tree --write-tree origin/develop HEAD` → `1fe08982d` (develop이 움직여 기록된 `1a1377192`와 다름, 예상된 차이). 그 병합 트리에 대한 8경로 diff는 `exit=0`, 그리고 `git ls-tree 1fe08982d -- internal/template/evidence_citation_guard_test.go` → `100644 blob 9b1970fe32f0…` — 기록과 동일한 blob. 즉 "병합 트리에는 있는데 이 카드 diff에는 없다"가 실제로 성립하며, 양쪽 모두 없어서 통과한 경우가 아니다. 같은 형식을 이 카드가 실제로 고친 `spec.md`에 걸면 `634 insertions(+), exit=1` — 실패할 수 있는 명령이다.

**Carve-out·do-not-touch.** 12줄 carve-out은 파일 단위 diff가 비어 있으므로 줄 단위 불변보다 강하게 성립한다. 양성 대조 합계 12 재확인. t375 소유 8경로도 diff 부재.

**치료법이 SPEC 서술과 맞는지** — 5건을 개별 대조했다. `mcp_glm.go`는 수치 5개가 본문에 남은 채 경로만 삭제(치료 a). `audit_pin_live_test.go`는 참인 절만 남고 거짓 절이 교체됨(치료 d). `evidence_writer_zeroexec_test.go`는 "증거가 아닌 출처 주석"으로 강등되며 미반출 사실을 명시(치료 b). `extract.txt`는 이미 정답 패턴이었고 움직이는 좌표만 제거(갱신이 아니라 제거). html 푸터는 단일 파일 지목 불가 사유를 그 자리에 기록. 전부 일치.

---

## 결함 목록

| # | 심각도 | 분류 | 위치 | 내용 |
|---|---|---|---|---|
| F1 | Medium | optional | `acceptance.md` AC-IEC-001 / `spec.md` REQ-IEC-001 | 요구는 **인접**한 비해결 표기를 강제하는데 기준은 `grep -L`이라 파일 단위다. 합성 파일로 재현: 1행 인용 + 21행 무관한 marker → 게이트 통과. 요구가 기준보다 넓다 |
| F2 | Low | optional | `acceptance.md` 전역 | AC 식별자가 12개인데 게이트는 10개. `AC-IEC-008`/`009`에 은퇴 표기가 없어 기계 계수기가 살아 있는 것으로 읽는다 |
| F3 | Low | optional | `acceptance.md` §F DoD 3번 항목 | "이 SPEC 디렉터리와 증거 디렉터리 외에는 어떤 파일도 수정하지 않는다"가 HEAD에서 거짓이다 — `CHANGELOG.md`가 수정됐다. `progress.md`에 이 조정을 기록한 문장은 없다(`grep 'no other file'` → 무출력) |
| F4 | Low | optional | `acceptance.md` AC-IEC-007 RED 가능성 대조표 | 대조 측정값 `594 insertions(+)`가 이미 스테일하다(현재 `634`). 통과 조건은 아니지만, REQ-IEC-009가 줄 번호에 대해 금지하는 "움직이는 좌표" 하자와 같은 형태다 |
| F5 | Info | optional | `.moai/reports/t381/verify/ac-iec-*.txt` | 명령의 경로 인자가 `<5 in-scope files>`로 축약돼 있다. 출력은 축자이나 명령은 아니다(VCI §3.2). `acceptance.md`에 전체 형태가 있어 모호하지는 않다 |
| F6 | Info | optional | `.moai/reports/t381/` 전반 | 이 카드의 추적 증거가 `.moai/state/verify` 문자열을 **49건** 코퍼스에 새로 더한다(`census-p4.txt` 25, `go-hits.txt` 13, `verify/*.txt` 11). 인용이 아니라 인용된 명령 출력이라 어떤 요구도 위반하지 않지만, §C.7의 프로브 경계 공개는 이 자기 증식을 언급하지 않는다. 같은 프로브가 지금 25가 아니라 **73**을 낸다 |

blocking 0건. F1은 요구 실체를 건드려야 닫히고 운영자가 제외한 축이며, 실측된 marker 거리가 최대 1행이라 잠재 위험이다.

---

## 보고된 3건에 대한 판단

### 1. `SyncSHASlotFormat`과 빈 슬롯 — 처리는 옳았다

소스로 확인했다. `internal/spec/syncsha.go:23` 문법이 `token := SHA | PLACEHOLDER`이고 둘 다 비어 있을 수 없으며, `:115`가 "placeholder는 아직 실제 SHA를 빚진 슬롯"이라고 못 박는다. `internal/spec/lint_syncsha.go:83` `Check`는 형제 `progress.md`를 훑고 severity는 **Warning**이다. 그러니 빈 슬롯은 진짜 위반이고 `pending-backfill`이 정본이다 — 리드 판정과 일치한다.

처리를 판단하면: 최종 상태는 `progress.md:440`이 실제 SHA `4f2eee332`를 담고 있고(주석은 첫 공백 토큰 뒤라 문법상 무시), 빈 슬롯은 정확히 미푸시 커밋 1개 동안만 존재했으며(`4f2eee332` → `dc0956728`), 자가 치유가 예측이 아니라 측정으로 제시됐다. **내가 이 트리에서 직접 잰 종점: `0 error(s), 1096 warning(s)`** — 1096 → 1097 → 1096의 마지막 값이 독립적으로 재현된다. 이 SPEC 귀속 findings는 0건이고, 코퍼스에 남은 `SyncSHASlotFormat` 5건은 다른 SPEC 소관이다.

**잔여는 실재하지만 작다.** 잔여는 착지한 규칙과 어긋나는 관행을 dispatch가 지시했다는 사실 자체이며, 그것은 이 카드가 아니라 dispatch 쪽 부채다. 아티팩트에 남은 흔적은 `progress.md:474` 한 줄과 CHANGELOG의 자진 신고뿐이고, 둘 다 감추지 않고 적었다. 규칙이 Warning인 이상 develop을 붉게 만들 수 없었다는 점도 실측으로 확인된다.

### 2. AC 개수 12 대 10 — 공개는 충분하다

측정: `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` → **12** (`AC-IEC-001`…`012`). 재확인했다.

아티팩트가 **자기 자신에 대해 거짓을 말하는가**가 원래 질문이었고, 답은 아니다. CHANGELOG는 "10 of 10"이라는 헐거운 주장을 피하고 `AC-IEC-001..007, 010..012`라는 명시 범위를 쓴 뒤, 바로 이어서 008/009가 plan-audit에서 비게이팅 §D 구조 검사로 강등된 이유까지 밝힌다. `acceptance.md` §B 매트릭스도 10행이고 그 아래에 강등 사실을 적는다. `progress.md:443`은 한발 더 나가 계수기가 12를 읽는 **기제**(은퇴 표기 부재 → 표기 규약상 '모호' 사례가 아니라서 계수기가 멈추지 않고 12를 낸다)까지 기록하고, 표기 추가는 `completed` 이후 본문 편집이라 manager-docs 권한 밖이므로 리드에 보고한다고 명시한다.

즉 불일치는 **은폐된 것이 아니라 세 곳에서 각각 다른 층위로 공개**돼 있다. 이것은 이 카드의 결함 유형(해결되지 않는 것을 근거인 양 제시)과 반대 방향이다. F2로 남기되 blocking이 아니다.

### 3. `plan.md`/`acceptance.md`에 `status:` 부재 — 스키마에 대고 확인했고, 부재가 **의무**다

보고를 그대로 받지 않고 규칙 소스를 읽었다. `internal/spec/lint_artifact_status.go:86`의 `statelessArtifacts = []string{"plan.md", "acceptance.md", "design.md", "research.md"}`, 그리고 `:5-9`가 `ArtifactStatusFieldForbiddenRule`(SPEC-ARTIFACT-STATELESS-001, REQ-AST-001-004..007)을 **error severity**로 정의한다. 축은 frontmatter 전체가 아니라 status 하나이며(`id`/`title`/`version`/`created`는 규칙 밖), `eraDemotableCodes`에 일부러 없다.

규칙이 공허하지 않은지도 확인했다: `TestArtifactStatus_FiresOnPlanStatus`, `TestArtifactStatus_FiresOnAllFourArtifacts`, `TestArtifactStatus_IgnoresSpecAndProgress` 전부 PASS. 즉 `plan.md`에 `status:`를 넣으면 실제로 error가 난다.

**따라서 "허용된다"가 아니라 "넣으면 error"다.** 카드의 서술이 스키마에 대고 성립한다. 전역 lint의 error 0건이 이를 다시 확인한다.

---

## dispatch에서 실행해 보니 그대로 성립하지 않는 것

1. **`make build`는 `catalog.yaml`을 다시 쓴다.** dispatch는 "필요하면 이 트리에서 빌드하라"고만 했지 쓰기 부작용을 말하지 않았다. 이번에는 재생성 결과가 바이트 동일이라 `git status --short`가 비어 있었지만(확인함), 감사자가 감사 대상 트리를 더럽힐 수 있는 경로다. 감사 지시에 "빌드 후 `git status` 재판독" 한 줄이 있어야 한다.
2. **`.moai/specs/<ID>` 디렉터리를 `moai spec lint`에 넘기면 `ParseFailure` error가 난다** (`is a directory`). 파일 경로를 줘야 한다. "targeted lint" 지시가 디렉터리를 함의하면 가짜 error를 만든다.
3. **전역 `moai spec lint`는 2분보다 오래 걸린다** — 실측 4분 이상. dispatch의 "2분 넘음"은 하한이지 실측 근사가 아니다.
4. **"12줄 carve-out이 그대로다"는 AC-IEC-006이 실제로 재는 것보다 약하다.** 기준은 파일 단위 diff 부재이므로 줄 단위 불변보다 강하다. 결함은 아니고, 표현이 기준을 과소 서술한다.
5. `origin/develop`은 감사 중 다시 움직이지 않았다(`5928095ea` 유지). merge base `3f03d9c36` 불변 — 세 번의 base 이동을 three-dot 기준선이 무편집으로 흡수했다는 카드의 주장은 이번 실행에서도 유지된다.

---

## Gaps — 관측하지 **않은** 것

- **CI 판정 없음.** 이 브랜치는 미푸시라 어떤 워크플로도 돌지 않았다. 전 패키지·크로스 플랫폼(darwin/windows) 판정은 미관측이다. 내가 돌린 것은 `internal/cli/...`와 `internal/hook/...` 두 트리뿐이다.
- **CodeRabbit 미관측.** PR이 없으므로 두 조건(combined status `Review completed` + `Merge Risk:` 접두 일치) 어느 쪽도 확인하지 않았다.
- **흡수 후 실측 아님.** `origin/develop` 병합은 지시대로 하지 않았다. 흡수 후 상태는 `merge-tree --write-tree`로 **계산**한 트리에 대한 판독이며, 실제 병합 트리에서의 빌드·테스트는 재지 않았다.
- **운영자 발언의 진위 미검증.** CHANGELOG가 인용한 운영자 판정("트리를 바꾸면 답이 달라지는 기준은 통과가 아니라 미판정")은 아티팩트 3곳과 **일관**됨만 확인했다. 운영자가 실제로 그렇게 말했는지는 내가 도달할 수 있는 범위 밖이다.
- **1096 → 1097 → 1096 중 앞 두 값 미재현.** 종점 1096만 직접 쟀다. 1096(수리 전)과 1097(빈 슬롯)은 이전 커밋에서 lint를 다시 돌려야 하는데, 커밋 체크아웃은 이 세션에 금지돼 있어 하지 않았다.
- **`.md` 코퍼스(B1)와 `.moai/state/` 비-verify 467줄(B2) 미측정** — 카드가 공개한 대로 넓히지 않았다.

## 잔여 위험

- **F1이 잠재에서 실재로 바뀌는 경로가 열려 있다.** 지금은 marker 4개가 전부 인용 지점에서 0~1행 거리에 있어 안전하지만(실측), 누군가 멀리 떨어진 곳에 marker를 넣어 AC-IEC-001을 "만족"시키는 순간 게이트가 그것을 받는다. 후속 카드가 필요한 재료는 `spec.md` §E.1에 이미 정리돼 있다.
- **C4 지시문이 그대로다.** 에이전트에게 `.moai/state/verify/$MOAI_SESSION_ID/`에 쓰라고 시키는 8줄이 t375 축에 남아 있어 같은 인용이 재생산될 수 있다. 카드가 명시적으로 공개한 미폐쇄 축이다.
- **F6의 자기 증식이 다음 census의 분류 부담을 늘린다.** 73건 중 49건이 이 카드 산출물이며, 다음 감사자가 이를 C2(인용된 출력)로 분류할 근거가 코퍼스 안에 명시돼 있지 않다.
- **통합 시점 재측정이 남아 있다.** develop 흡수 후 병합 트리에서 `internal/cli`·`internal/hook`를 다시 재고, CI 초록을 PR head에서 읽어야 판정이 완결된다. 이 감사는 흡수 전 트리에 대한 것이다.
