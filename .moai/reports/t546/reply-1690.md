# 회신 초안 — GH #1690 (t546 신규 작성)

> 배경: 스윕 중 t515(SPEC-WIN-SMARTPATH-001)가 착지(`52f863f36`, 2026-09-08)해 보류에서 해소로 재분류. t515는 회신 초안 없이 착지해 본 카드가 작성.
> 게시: 리드 승인 후. 대상 제보자: mihaesinbi. 언어: 한국어.

---

mihaesinbi님, 보고 정확했습니다. 결함을 확인했고 수리가 develop 라인에 반영됐습니다. v3.2.0 릴리스에 포함될 예정입니다.

원인은 `settings.json`의 `env.PATH`를 만드는 생성기가 플랫폼을 구분하지 못한 것이었습니다. Windows에서도 POSIX 스타일 디렉터 여섯 개(`/usr/local/bin` 등)를 사용자 경로와 섞어 세미콜론으로 이어 붙였고, 정작 `bash.exe`가 들어 있는 디렉터는 하나도 없었기 때문에 — 제보하신 대로 — exec form 훅(`"command": "bash"`)이 세션 시작 때 전부 실패했습니다. 제보에 첨부된 세 개의 settings 파일에서 관측된 값과 생성기가 배출한 값이 바이트 단위로 일치하는 것까지 확인했습니다.

수리는 PATH 생성 코어에 대상 플랫폼을 주입하는 구조로 바꿨습니다. Windows 분기에서는 사용자 홈 두 항목, `System32`, 그리고 환경변수에서 찾은 Git Bash 후보 디렉터(실제 존재가 확인된 것만)로만 PATH를 구성하고, POSIX 패키지 매니저·시스템 경로는 Windows에서 절대 붙이지 않습니다. 플랫폼별 결과값을 문자열 그대로 고정하는 테스트가 들어가 있어 재발을 잡습니다.

v3.2.0에서 직접 확인해 주시면 감사하겠습니다. 이 이슈는 확인 후 닫겠습니다.

---

> t546 검증 각주 (게시 본문 아님): 착지 근거 — 병합 `52f863f36` 메시지("fixing GH #1690") + develop CHANGELOG SPEC-WIN-SMARTPATH-001 항목(GOOS-blind BuildSmartPATH, POSIX 6디렉터 혼입, `Executable not found in $PATH: "bash"`, buildSmartPATHFor(goos,…) 주입형 코어, windows 분기 전용 구성). t515 리포트 디렉터는 develop에 비어 있어 본 초안이 유일한 회신 재료.
