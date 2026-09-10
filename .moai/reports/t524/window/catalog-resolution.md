# t524 통합 창 — catalog.yaml 충돌 해소 기록

- 흡수: 카드 브랜치 HEAD `446b40668` 에서 `git merge --no-ff develop` 실행. 로컬 develop tip 은 `4b82591cb`.
- 충돌: `internal/template/catalog.yaml` 한 곳, 훙크 1개(142-150행 plan-auditor 해시)
  - HEAD 쪽: `d7be4a4a8695459dc5d80d76543a7091476a6423f28ffbf542ab30032fe711b3`
  - develop 쪽: `621cb9eea7abe2c649be12d5aa3f494feb7fd525b3f278f94c497b6f9d306298`
- develop 쪽 값의 출처: 흡수 델타(`446b40668...4b82591cb`)에서 이 파일은 두 해시 줄만 바뀌었다. `manager-develop` 과 `plan-auditor` 다. `plan-auditor.md` 템플릿 자체는 델타에 없다.
- 흡수 트리의 템플릿 해시: `shasum -a 256 internal/template/templates/.claude/agents/moai/plan-auditor.md` → `d7be4a4a…`

## 해소 방법 — 어느 쪽도 고르지 않고 생성기로 결정

1. 생성기는 `yaml.Unmarshal` 로 파일을 읽는다(`gen-catalog-hashes.go` 216행). 충돌 표지가 있으면 읽지 못한다. 그래서 파싱 가능한 입력으로 `git show :3:internal/template/catalog.yaml` 을 썼다. 이 blob `9445d53b1` 은 `develop:internal/template/catalog.yaml` 과 같다.
2. `go run ./internal/template/scripts/gen-catalog-hashes.go --entry plan-auditor` 를 실행했다. exit 0, 출력 `plan-auditor: d7be4a4a…`. 로그는 `gen-entry.log`.
3. 확인:
   - `git diff develop -- internal/template/catalog.yaml` 결과는 plan-auditor 해시 한 줄(`621cb9ee…` → `d7be4a4a…`)뿐이다.
   - 충돌 표지는 0개다. 대조군으로 `hash:` 줄 45개를 셌다.

최종 값은 흡수 트리의 템플릿 바이트가 결정했다. develop 쪽 값은 수정 전 문서의 해시라서 이 트리에서는 낡은 값이다. 입력 파일의 나머지 내용(`manager-develop` 해시 등)은 원래 충돌 없이 병합된 부분과 같다.
