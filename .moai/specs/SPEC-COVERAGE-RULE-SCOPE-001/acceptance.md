# SPEC-COVERAGE-RULE-SCOPE-001 — 인수 기준

측정 기준 트리: `ee50984ab` (워크트리 `.claude/worktrees/t362`, 브랜치 `WT-coverage-rule-scope`). 아래 모든 기준은 **그 트리에서 빌드한 바이너리**로 검증한다. 설치본 `~/go/bin/moai`(v3.1.2, `343399d2f`)는 다른 트리이므로 근거가 되지 않는다.

## §A AC 목록

### 결함 ② — REQ 파서 (본체)

**AC-CRS-001-001** *(maps REQ-CRS-001-001)*
- **Given** 코퍼스에서 관측된 여섯 가지 REQ ID 형태(`REQ-HOOK-001`, `REQ-WF001-001`, `REQ-VNRN-RT-001-001`, `REQ-HRN-FND-001`, `REQ-TUX1-001`, `REQ-WC01-001`)를 각각 한 줄씩 담은 표 기반 Go 단위 테스트가 있을 때
- **When** 넓힌 `parseREQs`를 각 줄에 적용하면
- **Then** 여섯 형태 모두 REQ 정의 줄로 인식되고, 추출된 ID가 입력 토큰과 문자열로 일치한다.

**AC-CRS-001-002** *(maps REQ-CRS-001-001, REQ-CRS-001-004)*
- **Given** 넓힌 파서가 들어간 바이너리를 `ee50984ab` 기준 트리에서 빌드했을 때
- **When** 전 코퍼스에 대해 REQ 수집량을 측정하면
- **Then** REQ 정의 줄을 갖는 `spec.md` 수가 현행 15에서 유의하게 증가하고, 그 실측값이 증거로 기록된다. 계획 단계 추정치 741 / 62는 대조 참고값으로만 인용하며 baseline으로 쓰지 않는다.
- **왜 실제 Go 파서여야 하는가**: Python 근사의 커버리지 술어가 `extractACLines`보다 **넓다**. 근사는 `maps REQ-…`를 문서 어디에 있든 커버로 세고, Go 파서는 AC 절 안의 출현만 읽는다. 따라서 **근사는 자기 오차를 스스로 한정하지 못한다** — 재측정은 근사를 정교하게 만드는 일이 아니라 술어를 바꾸는 일이다.
- **방향 단언(검증 가능해야 함)**: 위 오차원 때문에 741은 **하한**으로 행동할 것으로 예측한다. 실측이 741 이상이면 예측대로이고, 미만이면 예측이 틀린 것이므로 그 사실을 그대로 기록한다. "추정과 유사"라는 서술은 방향을 세지 않은 채로는 쓰지 않는다.
- **세지 말아야 할 것**: AC 절에 도달하지 못하는 16개 SPEC은 AC 토큰이 0이라 이미 741 안에 계산돼 있다(plan.md §B 관측 3). 이를 별도 증분으로 다시 더하면 이중 계상이다.

**AC-CRS-001-003** *(maps REQ-CRS-001-002)*
- **Given** 넓힌 `reqLinePattern`이 적용된 상태에서
- **When** 전 코퍼스 lint를 실행하고 `InvalidREQID` 코드를 계수하면
- **Then** 그 건수가 0이거나, 0이 아닌 경우 각 건이 실제 규약 위반임을 개별 근거와 함께 제시한다. 넓힌 추출 때문에 생긴 대량 오탐이 남아 있어서는 안 된다.

**AC-CRS-001-004** *(maps REQ-CRS-001-005)*
- **Given** `parseREQs`가 인식해야 할 형태의 REQ 줄만 담긴 픽스처가 있을 때
- **When** 파서를 적용하면
- **Then** 반환된 REQ 슬라이스의 길이가 픽스처의 줄 수와 같다. 뮤테이션 확인: 패턴에서 한 분절 허용을 제거하면 이 테스트가 RED가 된다(공허 통과가 아님을 보인다).

**AC-CRS-001-005** *(maps REQ-CRS-001-003)*
- **Given** 최종 심각도 안(§D의 A / B / C 중 채택안)이 적용된 상태에서
- **When** 병합 대상 head에 대해 전 코퍼스 lint와 CI를 실행하면
- **Then** 미해소 `error` 등급 finding이 0건이며, CI가 조용한 head에서 완주해 GREEN을 낸다. 취소된 실행은 판정 근거가 아니다.

### 결함 ① — `acceptance.md` 판독 (잠복)

**AC-CRS-001-006a** *(maps REQ-CRS-001-006, REQ-CRS-001-008)* — 회귀 쌍의 앞쪽
- **Given** REQ 한 줄은 `spec.md`에 있고 그 REQ를 참조하는 AC는 형제 `acceptance.md`에만 있는 관행 모양 픽스처(현행 재현에서 `CoverageIncomplete`로 실패했던 것과 같은 배치)가 있을 때
- **When** 수리된 바이너리로 그 SPEC에 lint를 실행하면
- **Then** `CoverageIncomplete`가 발화하지 않고 종료코드가 0이다.

**AC-CRS-001-006b** *(maps REQ-CRS-001-007)* — 회귀 쌍의 뒤쪽 (필수)
- **Given** REQ 한 줄이 `spec.md`에 있고 그 REQ를 참조하는 AC가 `spec.md`에도 `acceptance.md`에도 **없는** 픽스처가 있을 때
- **When** 같은 바이너리로 lint를 실행하면
- **Then** `CoverageIncomplete`가 여전히 발화한다(**발화 여부로 판정한다** — A안 채택으로 종료코드는 양쪽 모두 0이므로 rc는 판정 기준이 아니다).
- **판정 기준이 rc에서 발화 여부로 옮겨간 이유**: §D에서 A안(`warning` + 발화 지점 `Advisory: true`)이 채택되면서 이 기준이 원래 쓰던 기제가 사라졌다. `internal/spec/lint.go:61`은 `if r.Strict && f.Severity == SeverityWarning && !f.Advisory`이므로, 자문 등급 warning은 `--strict`에서도 error로 올라가지 않는다. 따라서 이 픽스처는 plain에서도 `--strict`에서도 rc=0으로 끝나며, rc를 판정에 쓰면 **어떤 올바른 구현으로도 이 기준을 통과시킬 수 없다.** 기준의 의도(규칙이 여전히 돌고, 회귀 쌍의 두 결과가 갈린다)는 그대로이고 판정 수단만 finding 발화로 옮긴다.
- **왜 쌍이어야 하는가**: 앞쪽만 관측하면 "규칙을 껐다"와 "규칙이 올바르게 읽는다"를 구분할 수 없다. 두 결과가 갈려야만 수리가 확인된다. 판정이 rc에서 발화 여부로 옮겨간 것은 바로 이 구분을 유지하기 위해서다 — rc는 두 경우에 같은 값을 내므로 더 이상 갈라 세지 못한다.

**AC-CRS-001-007** *(maps REQ-CRS-001-006)*
- **Given** `acceptance.md`가 존재하지 않는 Tier S SPEC 픽스처가 있을 때
- **When** lint를 실행하면
- **Then** 파일 부재로 인한 오류나 패닉 없이, spec.md 인라인 AC만으로 커버리지를 판정한다.

**AC-CRS-001-008** *(maps REQ-CRS-001-008)*
- **Given** 수리가 착지한 뒤
- **When** 코퍼스에서 Tier M/L SPEC의 `spec.md`를 검사하면
- **Then** AC 중복 기재를 새로 요구하는 변경이 없다. `acceptance.md`가 AC의 SSOT라는 기존 규약이 그대로 유지된다.

## §B 품질 게이트

- `go vet ./internal/spec/...` 통과 (vet은 컴파일만 증명함에 유의).
- `go test ./internal/spec/...` 통과. 새 테스트는 뮤테이션으로 RED를 먼저 확립한 뒤 GREEN을 보인다.
- `golangci-lint run` 신규 지적 0건.
- 검증 범위는 건드린 패키지로 한정한다. 전 패키지 판정은 CI 몫이며, 로컬 전체 스위트는 돌리지 않는다.

## §C 완료 정의 (Definition of Done)

1. AC-CRS-001-001 ~ -008 전부 PASS, 각각 실행한 명령과 그 출력이 증거로 남는다.
2. M1 실측값(넓힌 파서의 실제 REQ 수집량)이 기록되고, 추정치 741과의 차이가 서술된다.
3. §D 심각도 안의 채택 결과가 plan.md에 반영되고, `[NEEDS CLARIFICATION]` 표식 두 건이 모두 해소된다.
4. 병합 대상 head에서 CI가 완주해 GREEN이며, 그 head가 조용했음(다른 커밋이 끼어들지 않았음)을 보인다.
5. 미검증 항목(전달값 "13건", `MissingExclusions` 가설, Tier별 발현 차이)이 Gaps로 남아 있고 검증된 것처럼 서술되지 않는다.

## §D 미검증 사항 (Gaps)

- 전달값 "13건"의 출처 — 관측되지 않음. 이 트리 측정값은 0.
- `MissingExclusions` 숫자 접두 가설 — **반증됨**(원인은 H3 하위 표제 부재). 다만 코퍼스 24건이 모두 같은 원인인지는 개별 미확인.
- AC 절 가로채기(세 번째 협소성)의 코퍼스 파급 건수 — 미측정. AC-CRS-001-002가 M1에서 센다.
- Tier L(89) / Tier S(113) SPEC의 결함 ① 발현 차이 — 미측정. 15/702 분할은 Tier가 아니라 REQ 줄 존재 여부로 갈렸다.
- 741 추정치의 실측 대응값 — M1에서 얻는다.
