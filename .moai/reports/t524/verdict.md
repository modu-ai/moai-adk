# t524 verdict — plan-auditor 추적성 검사: 축약 REQ 표기와 도구 침묵 인용

- 카드: t524 (Class B)
- 워크트리 `.claude/worktrees/t524`, 브랜치 `WT-auditor-req-shorthand`
- 코드 커밋:
  - `cf125c769` — 동사와 인용 규율
  - `7d355be1b` — Go 테스트
- 증거 커밋:
  - `fcb392deb` — 재현
  - `212cf013c` — 픽스처 실행과 뮤턴트
  - `55c49f791` — internal/cli 슬롯
- 기반: 로컬 develop `6d228ea19` 흡수 병합 `a6c9e406c`. 통합 창에서 최신 로컬 develop 을 다시 흡수해야 한다. 아직 흡수하지 않았다.
- 미푸시. 레인은 push 하지 않는다.

## Claim

1. plan-auditor Group 4 에 bash 검증 동사가 들어갔다. 이 동사는 REQ 정의와 AC 매핑을 줄머리 표지 `COLLECTED:` / `UNCOVERED:` / `ORPHAN:` / `GAP:` 로 판정한다. 템플릿 사본과 로컬 사본의 동사는 바이트 동일하다.
2. 축약 표기(`REQ-X-001, 002`)는 같은 칸, 같은 목록 안에서만 앞선 완전 토큰의 숫자 꼬리로 전개된다. 칸을 넘거나 산문 속 맨 숫자는 전개되지 않는다.
3. 동사는 제목형(`### REQ-…`), 목록형, 표 행 정의를 모두 수집한다. 재현 픽스처 D의 진짜 누락 `REQ-FIXD-003` 을 `UNCOVERED` 로 드러낸다. 정의가 0개면 통과가 아니라 `GAP` 이다.
4. 인용 규율 조항이 문서에 들어갔다. 다른 도구의 침묵은 그 SPEC에서 해당 축을 N>0 수집했음을 보인 뒤에만 방증으로 인용할 수 있다. 0이거나 보이지 않았으면 "그 도구는 이 축에 대해 말하지 않았다"로 기록한다.
5. 위 1~4는 Go 테스트 10개가 고정한다. 문서 뮤턴트 두 개(m1·m2)가 각각 해당 테스트를 RED로 만든다. 반대 방향(과잉 전개)은 bash 수준 뮤턴트 m3로 확인했다. 칸 분할·쉼표 목록 연속·뒤따르는 단어에서 멈춤, 이 세 경계를 지우면 픽스처 E의 진짜 누락 두 건이 사라진다. 이 방향은 기존 `TestPlanAuditTraceability_BareNumbersOutsideListDoNotExpand` 가 지킨다. 그 테스트는 픽스처 E와 같은 AC 줄을 입력으로 쓰고, 바로 그 두 `UNCOVERED` 줄이 있어야 통과한다.
6. `catalog.yaml` 의 plan-auditor 해시는 생성기로 재계산했다. 수정한 템플릿의 shasum `d7be4a4a…` 와 일치한다. `.codex` 방출본은 `make agents-emit` 으로 재생성했다.

## Evidence

| 주장 | 명령과 관측 | 경로 |
|---|---|---|
| 2층 결함의 재현 | 트리에서 빌드한 lint 로 픽스처 A~D 를 실행했다. D(제목형 정의 + 실제 누락)에서 추적성 경고 0건이었고, A·C 에서는 거짓 경고가 났다. | `repro.md`, `repro/lint-fix{A..D}.txt` |
| 동사 동작 (bash) | 픽스처별 결과: A·B·C `COLLECTED: 3`, D `UNCOVERED: REQ-FIXD-003`, E `UNCOVERED: REQ-FIXE-002/003`, S `UNCOVERED: REQ-FIXS-003`, Z `GAP: 0 …`, U `GAP: … not readable` | `run/fixture-results*`, `run/run-fixtures.sh` |
| 뮤턴트 (bash) | m1 을 A 에 적용하면 `UNCOVERED: REQ-FIXA-002`, m2 를 D 에 적용하면 `GAP` | `run/mutants/` |
| go vet | `go vet ./internal/cli/` → `go_vet_exit=0`, 출력 0바이트 | `slot/vet-internal-cli.{txt,exit}` |
| Go 테스트 GREEN | `-run '^TestPlanAudit(D7|D8|Traceability)'` → exit 0, PASS 19, FAIL 0 | `slot/test-green.{txt,exit}` |
| m1 RED | exit 1. FAIL 3건: `ShorthandMappingIsCovered`, `HeadingDefinitionsExposeRealGap`, `InlineACsWithoutAcceptance` | `slot/test-m1.{txt,exit}` |
| m2 RED | exit 1. FAIL 1건: `HeadingDefinitionsExposeRealGap` (`COLLECTED: 0` + `GAP`) | `slot/test-m2.{txt,exit}` |
| m3 과잉 전개 (bash) | 커밋본 동사를 추출한 대조군(`run/verb-template.sh` 와 `cmp` 0)과 m3를 `run-fixtures.sh` 로 픽스처 E에 돌렸다. 대조군은 `COLLECTED: 3`, `UNCOVERED: REQ-FIXE-002`, `UNCOVERED: REQ-FIXE-003`. m3는 `COLLECTED: 3` 한 줄뿐이다. 둘 다 exit 0. | `run/mutants/m3-no-expansion-boundary.sh`, `run/mutants/m3-out/`, `run/mutants/m3-control-{verb.sh,out/}` |
| 복원 | m1·m2 각각 복원한 뒤 `cmp` 를 돌렸다. 저장본 대비와 `git show HEAD:` 대비 모두, 두 사본 모두 exit 0이었다. 최종 GREEN 은 PASS 19, FAIL 0. | `slot/summary.md`, `slot/test-final-green.{txt,exit}` |
| 방출·카탈로그 | 복원 후 `make agents-emit-check` exit 0이었다. `git status --short` 로 두 사본, `catalog.yaml`, `plan-auditor.toml` 을 봤고 출력이 없었다. | `slot/agents-emit-check-after-restore.txt`, `run/agents-emit-check` |

## Baseline-attribution

- 코드와 문서 측정 트리는 `fcb392deb` 이다. 그 뒤 커밋 `55c49f791` 은 증거 파일만 더한다.
- Go 테스트는 이 트리에서 `unset MOAI_PROJECT_DIR && timeout 590 go test -count=1 …` 로 실행했다. 실행 전마다 외부 internal/cli 컴파일이 0건임을 쟀다(대조군 claude 20).
- 재현용 lint 는 이 트리에서 빌드한 바이너리를 경로로 직접 호출했다. 설치본은 쓰지 않았다. 스탬프가 없어 계기 좌표는 당시 HEAD `a6c9e406c` 로 귀속한다.
- 통합 창에서 develop 을 흡수한 뒤의 트리는 아직 측정하지 않았다. 병합 전 측정을 병합 후 근거로 재사용하지 않는다.

## Gaps

- **LLM 판정 자체는 Go 로 재현할 수 없다.** 테스트가 고정하는 것은 동사(bash/awk) 출력과 문서 조항의 존재까지다. plan-auditor 에이전트가 실제 감사에서 이 동사를 실행하고, 인용 규율을 지키고, PASS/FAIL 을 옳게 내는지는 관측하지 않았다.
- **awk 이식성은 macOS 에서만 봤다.** BSD awk 계열에서만 실행했다. gawk/mawk 와 Windows(git-bash) 동작은 재지 않았다. CI 매트릭스가 첫 관측이 된다.
- **과잉 전개 방향의 Go 수준 RED는 재지 않았다.** bash 수준에서는 m3가 픽스처 E의 두 누락을 가리는 것을 확인했다. 하지만 m3를 문서 사본에 넣고 `BareNumbersOutsideListDoNotExpand` 가 실제로 RED가 되는지는 돌리지 않았다(리드 결정: 추가 Go 슬롯 없음). 그 테스트가 이 방향을 지킨다는 판단은 입력 AC 줄이 같고 판정 문자열이 m3가 지운 바로 그 두 줄이라는 대응에 기댄다. m1 에서도 그 테스트는 PASS 했다. 전개를 없애는 뮤턴트로는 이 방향을 볼 수 없기 때문이다.
- **도구 쪽 축(spec lint 의 표 매핑·제목형 정의·축약 표기 수집)은 이 카드에서 고치지 않았다.** 리드 결정에 따라 표 매핑은 t561, 나머지 도구 축은 별도 카드 소관이다.
- **코퍼스 실측이 없다.** 기존 SPEC 에 동사를 돌려 거짓 양성·음성이 몇 건인지 전수로 재지 않았다. 실제 SPEC 1건(`SPEC-CODEX-PARTIAL-WIRING-001`)에서 교차 SPEC REQ 인용이 `ORPHAN` 으로 나오는 것만 확인했고, 이는 ORPHAN 설명문에 반영했다.
- **머신 전체 `ps` 원본 덤프는 커밋하지 않았다.** 다른 세션의 명령줄 인자가 섞일 수 있어서다. 사전 확인 계수는 그 덤프에서 이 슬롯에 잰 값이지만, 추적 경로로 재검증할 수 없다.
- **develop 최신본을 흡수한 트리의 재측정은 아직이다.** 통합 창에서 할 일이다. 그때 `catalog.yaml` 이 충돌하면 흡수된 트리의 plan-auditor.md 로 생성기를 다시 돌려 해결한다. 그 뒤 `./internal/template/` 을 선택자 없이 다시 잰다.

## Residual-risk

- **AC 줄의 우연한 REQ 언급이 거짓 통과를 만들 수 있다.** 동사는 AC 토큰이 있는 줄의 REQ 토큰을 모두 매핑으로 센다. "이 AC는 REQ-X-003 과 무관하다" 같은 부정 문장도 커버로 읽힌다. 감사자의 눈이 남는 방어선이다.
- **다른 줄에 걸친 매핑은 놓친다.** AC 토큰과 REQ 목록이 서로 다른 줄에 있으면 매핑으로 읽히지 않는다. 결과는 `UNCOVERED` 과잉 경고 방향이다.
- **표지 형식에 판정이 걸려 있다.** 테스트는 줄머리 표지로 판정한다. 누군가 동사의 출력 문구를 바꾸면 테스트가 먼저 깨지므로 조용한 퇴행은 아니다. 다만 에이전트 본문의 설명문과 동사가 따로 고쳐질 위험은 남는다. 사본 동일성 테스트는 동사 블록만 비교한다.
