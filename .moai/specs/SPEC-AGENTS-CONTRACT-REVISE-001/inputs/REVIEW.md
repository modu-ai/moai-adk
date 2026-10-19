# AGENTS·템플릿 지침 검토 및 인계

## Claim — 이번에 확인한 범위

고정 커밋 `1cc1c85486b19b70afe2ee83173d10d54f5b467c`의 핵심 문서 3개와 그 안에서 직접 지목한 것으로 해석되는 문서 35개를 끝까지 읽었다. 파일명만 적힌 참조는 로컬 규칙·정책 투영본·배포 템플릿 후보를 모두 포함했다. 간접 참조를 무한히 재귀 탐색한 전수 감사는 아니다. 생성·설치·update·PR·설정의 관련 소스는 필요한 구간을 읽었다. `coverage.json`은 전체 읽기와 구간 읽기를 구분하고, `read-manifest.json`은 실제 출력한 행 구간·SHA256·총 행 수를 담는다.

**제품 파일은 수정하지 않았다.** `proposed.patch`와 `proposed/`는 이 보고서 폴더 안에만 있는 검토용 사본이다. 제품 시험·빌드·init/update·DB/카드 쓰기·워크트리 생성·push/PR·원격 설정 변경을 하지 않았다. 이 산출물의 완료는 읽기 조사와 제안 저장의 완료이며, 정책 적용 또는 CI 전환 완료를 뜻하지 않는다.

현재 문서에는 새 사용자 정책과 다른 카드 PR 금지, 리더 일괄 push, 로컬 병합 후 종료, 리더 직접 실행 예외가 남아 있다. 소스상 PR 생성 경로도 아직 draft 강제가 없고 곧바로 `--auto` merge를 요청한다. 따라서 **문서 정책, 구현 상태, 설치 바이너리 상태를 따로 표시해야 한다.**

최신 사용자 확인은 앞선 수동 병합 정정을 대체한다. **ADK는 연결된 GitHub 사용자 본인이 작성한 비포크·비봇 PR에 한해서 조건부 auto-merge를 허용한다.** 본인 확인은 실제 연결에서 조회하며 제품에 login을 고정하지 않는다. 다른 작성자·fork·bot PR은 자동 활성화 금지이며 사용자 테스트 확인 및 명시 승인이 모두 필요하다. 본인 PR도 최신 SHA의 필수 CI·실제 Codex 리뷰·지적 해소·보호 규칙·필요한 타인 승인을 유지한다. 공개 저장소를 유지하며 본인 PR을 포함해 자동 브랜치 삭제를 금지한다. 이는 로컬 운영 선택이며, 범용 제품의 수동 기본·단계별 opt-in을 바꾸거나 기존 관리 설정을 임의로 변경할 승인이 아니다.

## Baseline-attribution — 세 기준선

| 대상 | 이번 관측 | 해석 |
|---|---|---|
| 로컬 main | `1cc1c85486b19b70afe2ee83173d10d54f5b467c` | 모든 `source/` 스냅숏과 제안 diff의 기준 |
| primary | detached `478285ab4c1fc893b020bb464d14cb69c8422523` | 오래된 공유 작업 사본. 최초 관측에서 수정 14개·미추적 2개가 있었음 |
| 전용 main 작업 트리 | `.claude/worktrees/main`, 최초 `git status --short` 출력 없음 | 그 시점의 clean 관측. primary 변경의 안전성 증거가 아님 |
| 설치 moai | build suffix `1888-g7705b365e2`, built `2026-10-18T10:45:08Z` | 고정 main 소스와 구별. 새 정책 구현을 이 바이너리가 제공한다고 주장하지 않음 |

이 조사자의 ref 재확인에서는 primary와 local main 값이 위와 같았다. 이후 부모가 local main의 `5122060f9e0b12bbfdf154e2fd34fbae1388d86e` 진행을 직접 관측해 전달했다. 전달된 `main-advance-delta.json`을 읽고 31개 변경 경로와 이번 읽기 목록 59개를 대조한 교집합은 0개다. 이는 부모가 수집한 경로 증거의 대조이며, 새 HEAD의 전체 본문·원격 상태를 이 조사자가 다시 검증한 것은 아니다. `parent-evidence.json`에 출처와 해시를 보존했다. 제안 기준은 계속 `1cc1c854…`로 고정한다.

writer는 적용 직전에 최신 main과 자신의 작업 트리·owner/lease를 다시 읽고 이 제안을 재대조해야 한다. 조사 이후 움직인 main에 자동 적용해서는 안 된다. 원격 최신성은 이번 조사에서 조회하지 않았다.

## Evidence — 명령과 관측

원본 명령 배열·종료 코드·출력은 `baseline-observations.json`, `targeted-evidence.json`에 보존했다. 각 본문은 다음 읽기 방식으로 고정했고, 읽기 명령은 `read-manifest.json`의 경로와 구간으로 재현할 수 있다.

```text
python3 /tmp/adk-agents-review-20261019/docs-review/read_source.py <path> <start> <end>
내부 읽기: git show 1cc1c85486b19b70afe2ee83173d10d54f5b467c:<path>
출력: FILE <path> LINES <start>-<end>/<total> 및 행 번호가 붙은 본문
```

대표 관측 원문:

```text
$ git rev-parse HEAD
478285ab4c1fc893b020bb464d14cb69c8422523
$ git branch --show-current
[빈 출력]
$ git rev-parse refs/heads/main
1cc1c85486b19b70afe2ee83173d10d54f5b467c
$ moai version
[moai_cp/20261007_195927] [moai_cp/20261007_195927-1888-g7705b365e2] [built 2026-10-18T10:45:08Z]
```

```text
$ git ls-tree -r --name-only 1cc1c85486b19b70afe2ee83173d10d54f5b467c -- AGENTS.md AGENTS.local.md CLAUDE.md internal/template/templates/AGENTS.md.tmpl internal/template/templates/CLAUDE.md
AGENTS.local.md
AGENTS.md
internal/template/templates/AGENTS.md.tmpl
```

이 조회는 존재하는 세 경로를 양성 대조로 함께 반환했다. 해당 고정 트리에 두 `CLAUDE.md` 경로가 없는 것은 확인했다. 모든 사용자 프로젝트에 그 파일이 없다는 주장은 아니다.

| 핵심 파일 | 읽은 행 | SHA256 |
|---|---|---|
| `AGENTS.md` | 1–367 | `9b48a2afc34f6cc54ce01010cf4a0e1e2bb718270cd110e9b1ed05b3f5c04253` |
| `AGENTS.local.md` | 1–517 | `2857c3cde9efd3bdb1c4905c9f7b6770c1912d9bfffd33167ada9fa16a1d8486` |
| `internal/template/templates/AGENTS.md.tmpl` | 1–371 | `7a40f9b6a8551f64a5bfe845c1e512c320afdf78d137f353789382171db683ad` |

## 확인된 정책 충돌과 제안

아래 행 번호는 모두 고정 원본 기준이다. “충돌”은 읽은 문구와 이번 사용자 정책의 차이이며, 문구만 보고 운영 사고나 런타임 결함이 발생했다고 확정하지 않는다.

| 우선순위 | 근거 | 확인 사항 | 제안·소유 경계 |
|---|---|---|---|
| High | `AGENTS.md:10,34`; `AGENTS.local.md:441–460` | 단순 작업 직접 수행, 위임 대상 부재 등의 직접 실행 예외 | 리더는 조정·감독·증거 검토만 담당. 구현·수리·빌드·시험은 모든 작업 트리에서 금지. 긴급·worker 부재 예외 없음. 기존 diff 보존 후 공식 worker 인계. worker의 승인된 구현 권한과 구분 |
| High | `AGENTS.local.md:182–237`; `.claude/rules/local/repo-local-pr-policy.md:10–14`; `gitflow-lane-protocol.md:83–115` | 로컬 main 병합·리더 일괄 push·카드 PR 금지·병합까지 다음 배차 대기 | 현재 정책을 §4.0 한곳에 두고 짧은 카드 WT → scoped 검사 → draft PR → Codex 리뷰 → 안전 인계로 바꿈. 과거 모델은 효력 없는 기록으로 명시. 실제 의존만 대기 |
| High | `internal/cli/factory_card_pr.go:395–457` | 검증 tip push 뒤 410행 PR 생성 인자에 `--draft` 없음; 444행 `gh pr merge ... --auto ... --match-head-commit` 호출 | t1882 구현 소관. t1884 문서만 바꿔 “이미 draft이며 자동 merge와 분리됨”이라고 쓰면 안 됨. 승인된 PR 권한을 merge 권한으로 확대하지 않음 |
| High | `AGENTS.md:200–205`, 템플릿 202–207 | CodeRabbit 전용 review 완료 계약 | 필요한 Codex 리뷰의 producer·현재 head·완료 결과·근거 URL 계약으로 변경. 요청 댓글·일반 green·이전 SHA·누락·skip·실패를 pass로 취급하지 않음. 실제 check 이름/producer는 CI 담당 결과에 연결. t1874/lane22 계약 문서와 중복 작성 금지 |
| High | `factory_card_pr.go:495–550`; `AGENTS.local.md:211–216` | 소스는 MERGED·head·merge SHA 확인 후 `merged-pr`/closure 경로를 가짐. 문서에는 로컬 병합 종료 및 t1621 미구현 문구가 남음 | PR 제출은 done 아님을 명시. 원격 착지와 완료 증거 후 종료. closure 후처리 실패는 소스 545–550행에서 안내 후 재조정하므로 무조건 종료 성공이라 서술하지 않음 |
| High | `AGENTS.local.md:184,198`; primary 상태 출력 | 전용 main은 clean, primary는 detached이고 다른 변경 보유 | primary-main 정리는 보존·귀속·공식 담당 인계가 선행. checkout/reset/stash로 강제 정렬하지 않음. 이번 제안은 복구 실행 승인이 아님 |
| High | `internal/config/defaults.go:1333,1354,1378–1383,1474,1495–1505` | Git mode 기본 `team`, manual의 자동 push/PR false와 auto-commit true, team draft true/auto-push true 등이 공존 | 범용 수동 기본·단계별 독립 opt-in은 새 요구. 기존 사용자 값과 이 저장소 로컬 factory 값을 일괄 false로 만들지 않음. `ExecutionMode: auto`는 실행 하네스 축, `TodoEnabled:nil`은 안내 기본이므로 자동 카드 발행과 혼동 금지. t1895 구현·시험 필요 |
| Medium | `AGENTS.local.md:70,446`; `codex_contract.go:190–205` | 없는 template `CLAUDE.md` 및 `CLAUDE.md Section 4`를 직접 가리킴. 계약 생성은 absent AGENTS만 쓰고 기존 AGENTS 보존, legacy CLAUDE는 비활성 | template AGENTS 및 실제 delegation/catalog 정의로 pointer 수정. CLAUDE 로컬 개발 세션·GLM 레인을 삭제하는 의미로 확대 금지 |
| Medium | `AGENTS.local.md:50,149`; `.moai/docs/local-dev-guide.md:111–125`; `embed.go:3–8`; `Makefile:34–74` | 문서는 `all:templates`·build가 재생성한다고 설명, 실제는 생성된 per-file allowlist와 명시적 emit/check 분리 | `embed_manifest_gen.go`, 명시적 emit, build의 drift 검사·컴파일을 구분. 고정 명령 수 대신 emitter 입력 집합 사용 |
| Medium | `update-local-file-survival.md:8–29,33–80`; `harness-namespace-doctrine.md:44–47`; `version-management.md:105–109`; `update_reconcile.go:33–166` | 최신 분류 보존 계약과 과거 wipe·일괄 restore 안내가 공존 | 사용자 설정 보존 계약을 현행으로 명시. 과거 복구 명령을 실행 지시에서 분리. 실제 upgrade 보존 성공은 아직 시험하지 않음 |
| Medium | `git-local-workflow-doctrine.md:86–153`; `git-workflow-doctrine.md:253–307,367–375` | 과거 branch-protection 변경 예시, stash/reset 대체, hook 우회와 gate 예외 표현 | 역사 기록임을 명확히 하고 현재 승인 거부의 우회 수단으로 사용하지 못하게 함. 새 권한·보안 설정 변경을 이 검토로 승인하지 않음 |

`proposed.patch`는 11개 문서에 대한 구체적인 최소 이행 제안이다. 과거 기록을 삭제하지 않고 현재 절의 효력을 분리했다. 다만 간접 소비 문서까지 정합된 최종 패치가 아니며, 그대로 전체 적용할 준비가 끝났다는 뜻은 아니다. 아래 소유권·검증 계획을 함께 읽어야 한다.

## 생성·배포·update·하네스 진입 경로

| 단계 | 기계적으로 읽은 경로·행 | 관측한 연결 |
|---|---|---|
| 지침 원본 | `AGENTS.local.md:91–115` | 로컬 C1 / 중립 template C2 / 생성 C3를 구분. 로컬 전용 지침은 범용 template로 복사하지 않음 |
| embed 입력 | `embed.go:3–8,33–35`; `embed_manifest_gen.go:647` 검색 관측 | tracked per-file allowlist가 template AGENTS를 운반하고 `EmbeddedTemplates()`가 `templates` 하위 FS를 반환 |
| build | `Makefile:34–74` | build 선행 drift check; `agents-emit`, `commands-emit`, `embed-manifest`는 명시적 생성 동사. 이번에 실행하지 않음 |
| agentemit | `internal/template/agentemit/emit.go:20–120` 전체 | Markdown walk·frontmatter 파싱·render·권한 계약 검사를 거쳐 TOML 출력. C3 직접 수정 대신 입력 수정 후 emit |
| commandemit | `internal/template/commandemit/emit.go:23–131` 전체 | 명령 source walk·이름 충돌 검사 후 SKILL 출력. 본문은 129행에서 그대로 복사하므로 emitter가 하네스 문구를 자동 수리하지 않음 |
| init | `internal/cli/init.go:744–824,908–916` | 하네스 선택에 맞는 deployer 선택. 사용자 공통 자산 ensure가 프로젝트 실행보다 앞섬 |
| project 배포 | `deployer.go:188–214,244–270`; `deployer_mode.go:33–44` | 공통 skills/agent 루트는 project walk에서 제외. `.tmpl` 렌더·suffix 제거, 기존 파일 provenance 보호 |
| 공통 자산 | `user_asset_phase.go:30–83,117–150`; `internal/userassets/paths.go:35–42` | 사용자 lock과 manifest를 거쳐 HOME의 `.claude/skills`, `.claude/agents`, `.agents/skills`, `.codex/agents`로 설치 |
| 하네스 중립 투영 | `harness_fs.go:130–198` | `.moai/policies`←규칙, `.moai/workflows`←워크플로 대응 및 경로 정규화. tracked 사본과 배포 결과를 같은 것으로 추정하지 않음 |
| update | `update.go:530–604`; `update_reconcile.go:33–166` | 사용자 자산 갱신과 project sync 분리. 실패/skip 시 project 자산 제거 gate, 일반 project 경로는 관리 경로 분류·config 보존 |
| 직접 Codex | `AGENTS.md:17–20,303–317` | 직접 세션은 존재하는 worktree AGENTS.local을 먼저 읽음. 파일 부재와 읽기 실패를 구분 |
| moai Codex 런처 | `codex_launcher.go:170–223` | AGENTS.local 다음 legacy local을 읽고 source header와 developer instructions로 전달. operator override 충돌 거부 |
| 계약 생성 | `codex_contract.go:190–205` | 없는 AGENTS만 생성; 기존 내용 보존; CLAUDE를 생성하거나 읽는 경로가 아님 |
| web 설정 연결점 | `internal/web/projectconfig.go:150–247`; `internal/settings/sectionroute.go:55–119` | 실제 MoAI project config는 `.moai/config/sections`와 공통 settings writer를 씀. workflow nested 값은 YAML patch 경로가 필요. 모두의AI BO가 아님 |

root AGENTS의 project skill 경로와 배포 template의 HOME skill 경로는 다르다. 위 설치 분리를 고려하면 문자 차이만으로 유출이나 잘못된 배포라고 판정할 수 없다. `projection-comparison.json`에서 같은 파일과 실제 내용 차이를 구분했다. 특히 `.moai/policies` 사본의 오래된 CLAUDE pointer·Kanban 언급을 확인했으나, 모든 소비자가 이 tracked 사본을 읽는다는 동적 증거는 없다.

## 의존·인계·실패 격리

`factory_card.go:271–281`의 serial 해제 상태에는 `merge-ready`, `pr-open`, `failed`, `abandoned` 등이 있다. `302–305`는 상태·만료·owner를 함께 본다. `homestate/card_record.go:147–156`에서 expiry가 비어 있으면 만료로 보지 않는다. 따라서 holder 문자열 하나나 PR 존재만 보고 카드·시험 슬롯이 모두 반환됐다고 단정하지 않는다.

문서 계약은 다음처럼 작게 유지한다. 인계에는 카드·base/head·허용 파일·PR·검증 증거·review/repair 담당·다음 동작·lease/resource 상태를 기록한다. 강결합 변경은 같은 책임 안의 단계로 묶고, 결과물을 실제 소비하는 카드만 DAG 의존으로 둔다. 실패한 카드와 그 결과가 필요한 카드만 막는다. 다음 독립 카드 착수는 안전 인계가 조건이며 이전 PR을 done으로 바꾸는 편법이 아니다. 필요한 runtime guard는 기존 카드/lease 경로에서 별도 구현·시험한다.

## 담당 경계와 적용 순서

다음 카드 상태는 부모 조율자가 전달한 정보다. 이번 조사자가 카드 DB나 실제 writer 상태를 재조회한 결과는 아니다.

| 담당 | 이번 인계에서의 경계 |
|---|---|
| t1874 / lane22 | Codex 계약 문서와 AGENTS review 계약 hunk가 중첩. 동일 파일·절은 단독 writer를 정한 뒤 한쪽 결과를 인계 |
| t1882 / lane44 예정 | draft PR, 작성자·fork·bot 및 최신 head gate를 지키는 ADK 조건부 own-PR auto-merge, delivery/handoff runtime. 고정 main의 현재 소스는 아직 새 정책을 충족하지 않음. 특정 login 하드코딩·임의 원격 설정 변경·자동 브랜치 삭제 금지 |
| t1884 | 문서 정합 후속. picked·owner 없음·t1882 뒤라는 전달 상태. 실제 쓰기 배정은 pending |
| t1895 | 범용 manual default, 단계별 독립 opt-in, web→API→runtime 일치·권한/중지/감사 후속 |
| 본 조사자 | read-only 조사와 이 폴더 보고서·제안만. 제품 writer 권한 없음 |

적용 순서는 (1) 부모가 최신 카드·writer를 확인하고 t1874/t1882/t1884 중첩 절을 배정, (2) 공식 worker가 최신 main과 제안 재대조, (3) 먼저 현재/역사·리더 역할·보존 문구 정합, (4) 구현된 t1882 계약을 정확한 CI/check 증거와 연결, (5) t1895는 별도 구현·검증으로 진행한다. Kanban 구현 또는 모드 전환을 이 범위에 넣지 않는다.

## Gaps — 실행으로 확인하지 않은 것

- 새 draft/자동 merge 분리·독립 opt-in·web 유효 설정 계약이 실제 동작하는지 시험하지 않았다. 코드 존재와 시험 이름 존재는 통과 증거가 아니다.
- `go test`, make emit/check/build, init/update, 실제 PR 생성·리뷰·병합, 설치 바이너리 교체를 실행하지 않았다.
- GitHub 앱·권한·required check·실제 Codex producer/리뷰·원격 HEAD·CI 상태를 이번 문서 조사에서 확인하지 않았다. 별도 CI 담당 영수증이 필요하다.
- primary 변경 파일의 내용·소유권·ignored 파일 전체를 조사하거나 복구하지 않았다. 안전한 checkout 가능성을 판정하지 않았다.
- 직접 참조의 glob 디렉터리 전체, 전역 memory의 `lessons.md`, 외부 문서, 직접 참조 문서가 다시 가리키는 모든 문서·스킬·agent 본문은 전수 읽지 않았다. 간접 항목은 `unread-boundary.md`에 기재했다.
- 공급자 제품 기능·가격·공식 API의 최신 사실을 새로 검증한 보고서가 아니다. 읽은 문서의 과거 주장이나 수치를 현재 사실로 재인용하지 않았다.

## Residual-risk — 적용 때 남는 위험

문서만 바꾸고 기존 바이너리를 그대로 쓰면 draft 정책을 쓰면서 실제로는 일반 PR·auto-merge를 요청할 수 있다. 반대로 보존 검증 없이 모든 설정을 manual로 덮으면 이미 승인된 로컬 factory 동작을 잃는다. 긴급/worker 부재 예외를 다른 실행 지침에 남기면 리더 금지 규칙이 다시 약해질 수 있다. 따라서 공식 writer는 `verification-plan.md`의 제한된 검증과 간접 소비 문서 점검을 수행해야 한다. 현재 보고서는 이러한 적용 작업과 원격 완료를 대체하지 않는다.
