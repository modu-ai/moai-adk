# acceptance.md — SPEC-WIN-SMARTPATH-001

기계 판정은 spec.md §D의 AC-CWSP-001..008이 원천이다. 본 문서는 Given/When/Then 시나리오와 경계 사례로 그것을 전개한다.

## Scenario 1 — Windows 에서의 init 렌더 (본 결함 시나리오)

- **Given** `buildSmartPATHFor("windows", "C:\\Users\\u", envLookup, stat, false)` 이고 stat 가 `C:\Program Files\Git\bin` 만 true
- **When** 그 반환값이 `settings.json.tmpl` 의 `{{jsonEscape .SmartPATH}}` 로 렌더된다
- **Then** `"PATH"` 값은 `C:\Users\u\.local\bin;C:\Users\u\go\bin;C:\Windows\System32;C:\Program Files\Git\bin` 형태(`;` 조립, Windows 항목만)이고 `/usr/...` 가 하나도 없으며 JSON 유효다

## Scenario 2 — darwin 비회귀

- **Given** M1 시작 시 캡처한 darwin 픽스처(무수정 함수의 실출력, provenance 주석)
- **When** `buildSmartPATHFor("darwin", <fixture home>, envLookup, stat, false)` 실행
- **Then** 반환값이 픽스처와 바이트 동일하고, 기존 `TestBuildSmartPATH_StableAcrossTerminalPATH`·WSL2 스위트가 무수정으로 green

## Scenario 3 — linux 비회귀 + WSL2

- **Given** linux GOOS 입력(wsl2=false / true 각각)
- **When** 생성기 실행
- **Then** wsl2=false 출력은 사전 상태와 바이트 동일, wsl2=true 는 기존 `/mnt/` 캡처·필터 동작 유지(`TestBuildSmartPATH_WSL2` 스위트 무수정 green)

## Edge cases

- **E1 — Git Bash 부재**: 세 후보 모두 stat=false → windows PATH 는 홈 2항목(+SystemRoot)만. POSIX 부재 유지, 빈 값 아님(clean-install 가드 `env.PATH = ""` 회피)
- **E2 — SystemRoot 미설정**: envLookup("SystemRoot") 공란 → System32 항목 생략, 나머지 불변
- **E3 — 후보 중 일부만 존재**: 존재하는 것만 순서대로 채택(후보 나열 순서 유지)
- **E4 — HOME 공란**: 래퍼의 기존 폴백(os.Getenv("HOME")) 유지 — 행동 변화 없음
- **E5 — 경로 내 특수문자**: jsonEscape 렌더 경로가 그대로 검증 대상(AC-CWSP-004)

## Quality gates

- AC-CWSP-005 전수(windows 사례 포함 — empty-sweep 토큰 없음)
- AC-CWSP-006: `GOOS=windows go build ./...` rc0 (smoke) + `go test ./internal/template/... -count=1` ok
- 커버리지: internal/template 패키지 기존 수준 유지(하락 금지)
- 보안: 경로 하드코딩 — hardcoded_path_audit_test 의 4검사 중 3개는 본문 없는 스텁(단언 0, F3 실측)이라 공허 초록으로 취급하지 않는다. 신규 경로 상수의 안전성은 AC-CWSP-002 의 정확-문자열 단언(기계 클래스 경로만, 사용자 절대경로 없음)이 실질 게이트다. 실측 근거: F3 판정(plan-audit iter-1 review-1.md)

## Performance/quality criteria

- 생성기는 init/update 시점 1회 호출 — 프로브 os.Stat 비용은 허용(어드바이저리 체크 캐시 규율상 임계 경로 아님)
- 렌더 결정론 유지: 동일 입력 → 동일 바이트(R-006 정신 계승)
