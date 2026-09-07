# 회신 초안 — GH #1658 (t546 정식화)

> 출처: t511 `.moai/reports/t511/reporter-replies.md` 초안 1 — 본 카드에서 착지 사실 재검증 후 정식화.
> 게시: 리드 승인 후. 대상 제보자: jjjh7401. 언어: 한국어.

---

jjjh7401님, 보고 주셔서 감사합니다. 말씀하신 두 방향이 모두 수리되어 develop 라인에 반영됐고, 다음 릴리스 v3.2.0에 포함될 예정입니다.

1. **플래그 순서 우회(너무 좁은 방향)** — 기존 정규식은 플래그 묶음을 문자 그대로 비교했기 때문에 `rm -fr /`처럼 두 글자만 뒤바꿔도 통과했습니다. 이제 정규식 매칭 대신 명령어를 토크나이즈한 뒤 실제 삭제 대상 경로를 판정하는 구조적 체크로 보호 대상을 판단합니다. `rm -rf /`, `rm -fr /`, `rm -r -f /`, `rm --force --recursive /` 등 셸에서 동등한 모든 형태가 동일하게 차단됩니다.

2. **데이터 언급 오탐(너무 넓은 방향)** — 기존 방식은 명령어 전체 텍스트를 스캔했기 때문에 `echo "rm -rf /"`처럼 위험한 문자열을 **출력하거나 저장만 하는** 명령까지 거부했습니다. 이것이 바로 이 결함 자체를 문서화·테스트·보고하기 어려웠던 원인이었습니다. 이제 명령어 실행이 실제로 위험한 지만 판정하므로, 데이터로서의 언급(heredoc 본문 포함)은 허용됩니다.

한 가지 남았던 잔여 문제 — 빌트인 정규식을 걷어낸 뒤에도 배포되는 `security.yaml` 템플릿의 extras 목록에 같은 결함의 잔여 사본이 남아 같은 오탐을 냈던 부분 — 도 이번 정리에서 함께 제거했습니다(배포 템플릿·로컬 설정·테스트 픽스처 세 곳 모두). 이 이슈는 리포터님께서 v3.2.0 릴리스에서 직접 확인해 주신 뒤에 닫겠습니다.

---

> t546 검증 각주 (게시 본문 아님): 구조적 가드 전환은 develop의 `internal/hook/dangerous_removal.go:9`("replacing the six `rm -rf <target>` regexes") + `tokenizeSegment` 인용 처리(:88-90)로, extras 3본 제거는 t511 병합 stat(`.moai/config/sections/security.yaml`, `internal/settings/testdata/sections/security.yaml`, `internal/template/templates/.moai/config/sections/security.yaml` 각 -1)으로 확인.
