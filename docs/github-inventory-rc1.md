# RC1 GitHub 설정·작업 목록 대조 (00-B, #1986)

`modu-ai/moai-adk`의 GitHub 측 상태를 GET 요청만으로 기록한 결과다. 정본 데이터는 아래 산출물이고, 이 문서는 요약과 사용자 결정 대기 항목(HOLD)을 정리한다. 수치는 모두 수집 시각 `2026-11-08T07:56:10Z` 기준이며, 이후 다른 세션이 label이나 Issue를 바꾸면 낡은 값이 된다.

## 산출물

| 경로 | 내용 |
| --- | --- |
| `.moai-cli/manifests/github-settings-inventory.json` | labels(축 분류·동의어 묶음), org issue types, milestones, Issue·PR 집계, sub-issue·dependency, manifest 대조, 양식·label 파일 sha256, 요청 로그 |
| `.moai-cli/manifests/github-duplicate-map.json` | 열린 Issue의 논리 키 매핑, 닫힌 Issue 중복 후보(사람 판단용 `candidate`), 후보 규칙과 관측값 |
| `.moai-cli/manifests/github-forms.sha256` | 양식 8개 경로의 sha256(`shasum -a 256 -c` 형식) |
| `tests/github_inventory_collect.py` | 수집기(표준 라이브러리만 사용, GET 전용, `python3 -I`로 실행 가능) |
| `tests/github_inventory_test.py` | 오프라인 테스트 17개(네트워크·`gh` 호출 없음) |
| `tests/fixtures/github/*.json` | 기록·편집된 원격 응답(필드 화이트리스트만 남김) |
| `tests/fixtures/github/forms-primary.sha256` | 기본 checkout에서 기록한 양식·label 파일 sha256 |

## 수집 방법

- 원격 모드: `python3 -I tests/github_inventory_collect.py --repo modu-ai/moai-adk --record tests/fixtures/github --out .moai-cli/manifests --forms-sha256 tests/fixtures/github/forms-primary.sha256`
- 오프라인 모드: `--repo … --record …` 대신 `--fixtures tests/fixtures/github`. 산출물은 fixture만의 순수 함수라서, 같은 원격 상태를 두 번 기록하면 `fetched_at`만 다르다.
- 모든 원격 호출은 `gh api -X GET`으로 고정되고, 요청 본문을 보내거나 메서드를 바꾸는 인자(`-f`, `-F`, `--input`, `--method` 등)는 실행 전에 거부된다. 이번 수집의 요청 46건은 모두 GET이었다(`requests` 필드).
- Issue 목록은 `issues?state=all` 한 번으로 받고 `pull_request` 키가 있는 항목을 PR로 분리했다. 합계 2,000건(Issue 464 + PR 1,536)이 우연히 정수라 상한에 걸린 것인지 따로 확인했는데, search API 집계(`is:issue` 464, `is:pr` 1,536)와 일치했다.
- API 버전: `X-GitHub-Api-Version` 헤더를 보내지 않았다(gh 2.92.0 기본). 이 상태에서도 `sub_issues_summary`, `issue_dependencies_summary`, `type` 필드가 응답에 있었다. 버전 고정 여부는 HOLD-00B-2로 남긴다.
- Projects v2: 토큰 scope에 `read:project`가 없고 `has_projects=false`라서 `projects: unavailable(read:project scope missing)`로 기록했다. 실패로 처리하지 않았다.

## 결과 요약

| 항목 | 값 |
| --- | --- |
| labels | 94개 (설계 당시 "약 100개") |
| label 축 | type 9 · priority 5 · status 9 · area 14 · host 2 · legacy-synonym 26 · functional 7 · unknown 22 |
| org issue types | Task(23014082)·Bug(23014083)·Feature(23014084), Epic 없음 |
| milestone | #1 `2026.11.0-rc.1` open 44 / closed 0 |
| Issue | 열림 44(#2016 04-D 추가 반영) · 닫힘 420 (completed 388, not_planned 31, duplicate 1) · Issue type 설정 0건 |
| PR | 전체 1,536 (열림 0, merge 1,324) |
| sub-issue | 42쌍 — manifest `links`(linked 42)와 차이 0 |
| dependency(blocked_by) | 51쌍 — manifest `dependencies`(linked 51)와 차이 0 |
| 열린 Issue 매핑 | 44건 모두 manifest 논리 키로 매핑, 비RC1 0 |
| 닫힌 Issue 중복 후보 | 0건 |

### 중복 후보가 0건인 근거

후보 규칙은 세 가지다: 제목에 `SPEC-RC1-…` 포함, 제목에 RC1 논리 키(`EP-0N`, `NN-X`) 포함, 앞쪽 `[..]` 꼬리표를 뗀 제목 토큰의 Jaccard 유사도 0.5 이상. 관측값은 다음과 같다(`observed` 필드).

- 닫힌 Issue 중 아무 SPEC ID든 제목에 있는 것은 36건이지만, 모두 Go 세대 SPEC이고 `SPEC-RC1-`은 0건
- RC1 논리 키를 제목에 가진 닫힌 Issue 0건
- 닫힌 Issue와 RC1 Issue 사이 최대 제목 유사도 0.20 (예: #271 "Windows moai-adk 설치 불가" ↔ #1998 04-A)

그래서 빈 목록은 건너뛴 결과가 아니라 측정한 결과다. 다만 규칙이 제목만 보므로, 본문까지 겹치는 작업(예: Go 세대 `moai update` 결함과 04-D 마이그레이션)은 이 지도에 나오지 않는다. 본문 매칭은 rate limit과 개인정보 노출 때문에 이번 범위에서 뺐다.

## label 분류 규칙

분류표는 수집기 안의 명시 표뿐이고, 표에 없으면 `unknown`으로 남긴다. 정리나 이름 변경은 하지 않았다.

- 이름공간 축: `type:` `priority:` `status:` `area:` `host:`
- 이름공간 없는 상태 label(대응 `status:*` 없음): `ready` `done` `testing` `triage` `investigating` → `status`
- functional: GitHub 기본 label과 Dependabot의 `dependencies`
- 동의어 묶음 19개. 근거가 두 종류다.
  - `issue-body`: #1986 설계 본문이 직접 든 짝 — `type:fix`↔`bug`, `type:feature`↔`feature`·`feature-request`, `type:chore`↔`chore`, `priority:P0`↔`critical`·`priority:critical`, `status:blocked`↔`blocked`, `status:in-progress`↔`in-progress`, `status:review`↔`review`
  - `name-match`: 이름만 보고 제안한 짝(예: `priority:P1`↔`high`, `area:hooks`↔`hooks`). HOLD-00B-1에서 확정하거나 버릴 후보다.
- unknown 22개: `ai-generated` `ai-reviewed` `api` `auth` `automerge` `breaking-change` `coderabbit` `database` `go` `improvement` `infra` `infrastructure` `needs design` `needs security review` `pinned` `planning` `reported` `requires-testing` `sdk` `spec` `stale` `ui`

## 양식 inventory

| 경로 | 추적 | 채택 | sha256(앞 12자) |
| --- | --- | --- | --- |
| `.github/ISSUE_TEMPLATE/moai-feature.yml` | tracked | 채택 | `6ab415cb6d69` |
| `.github/ISSUE_TEMPLATE/moai-bugfix.yml` | tracked | 채택 | `b2a87b7682be` |
| `.github/ISSUE_TEMPLATE/moai-investigation.yml` | tracked | 채택 | `f8f4bf176cd2` |
| `.github/ISSUE_TEMPLATE/agentic-factory-01-parent-issue.yml` | untracked | 미채택-보존 | `30056696d635` |
| `.github/ISSUE_TEMPLATE/agentic-factory-02-child-issue.yml` | untracked | 미채택-보존 | `78e1a62106f9` |
| `.github/ISSUE_TEMPLATE/agentic-factory-03-spec-change.yml` | untracked | 미채택-보존 | `0236f2a2251d` |
| `.github/PULL_REQUEST_TEMPLATE/agentic-factory-11-handoff.md` | untracked | 미채택-보존 | `aa6bc438d509` |
| `.github/workflows/label-sync.yml` | untracked | 미채택-보존 | `8dc3fd482b8c` |

label 파일 3종(`.github/moai/labels.yml` 채택, `.github/labels.yml`·`.github/agentic-factory/labels.proposed.yml` 미채택-보존)도 `label_files`에 같은 방식으로 기록했다. 채택 양식 3종은 `tests/github_assets_test.py`가 검사하는 목록과 같다(테스트로 확인).

untracked 5개는 공유 기본 checkout에만 있고 Issue worktree에는 없다. 그래서 sha256은 worktree에 들어가기 전에 기본 checkout에서 `shasum -a 256`으로 기록해 `forms-primary.sha256`에 담았다. tracked 3종은 작업 당시 worktree에서 다시 계산해 같은 값임을 확인했다. 이 양식들은 00-C(#1974)가 고칠 수 있으므로 상시 테스트로 고정하지 않는다. `shasum -c`(00B-AC-6)는 이 작업 전후 무변경을 보는 일회성 검사이고, 기본 checkout 루트에서 실행해야 8개 모두 검사된다.

## HOLD — 사용자 결정 대기

| ID | 주제 | 이번 작업에서 한 일 | 결정되지 않은 것 |
| --- | --- | --- | --- |
| HOLD-00B-1 | label 체계 | 현황표·축 분류·동의어 후보만 작성, 원격 label 변경 0 | `status:*`·bare 상태 label 사용 여부, 동의어 정리, `type:epic` 또는 org Issue type `Epic`, `area:router/update/forge` 신설, Issue type과 `type:*` label 중 무엇을 쓸지 |
| HOLD-00B-2 | REST API 버전 고정 | 헤더 없이 수집하고 `api_version`에 그 사실과 gh 버전을 기록 | **RC1 임시 결정(사용자, 2026-11-08 18:1x): gh 기본(헤더 미전송) 유지.** `2026-03-10` 고정은 07-A(#2008)에서 다루며, 고정하면 fixture를 다시 기록해야 한다 |
| HOLD-00B-3 | 수집기 위치 | Issue 파일 범위가 지정한 `tests/` Python으로 구현 | **RC1 임시 결정(사용자, 2026-11-08 18:1x): `tests/` Python 유지.** Rust `moai-cli` 명령으로 옮기는 일은 07-A(#2008) 후속 |

함께 남은 연결 HOLD: 00-C(#1974)가 00-B를 기다리는 dependency는 manifest의 `exception_not_linked`대로 원격에 연결되어 있지 않다.

## 위험과 미검증

- 원격 상태는 수집 시각에 묶인다. 소비하는 Issue(00-C, 07-A)는 쓰기 전에 다시 수집해야 한다.
- 이름 기반 동의어 짝은 제안일 뿐이다.
- 중복 후보 규칙은 제목만 본다.
- Issue form이 기본 브랜치에서만 렌더링된다는 점 때문에, 현재 RC 브랜치의 양식 렌더링은 확인하지 않았다(not_run).
