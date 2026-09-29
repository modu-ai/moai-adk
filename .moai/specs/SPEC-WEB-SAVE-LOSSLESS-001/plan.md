# SPEC-WEB-SAVE-LOSSLESS-001 Implementation Plan

> 카드 t1314 · GitHub issue #1731 · branch `WT-web-save-lossless` (develop `2b1233b13` 기점) · Tier M

## §A Design Decision (1급 항목 — 본 카드가 확정하는 설계)

### §A.1 채택: A+B 결합 + Save 더티-게이트 백스톱 (3층 방어)

무손실 계약(REQ-WSL-001/002/003)은 단일 기제로 충족되지 않는다. 세 층을 맞댄다:

| 층 | 내용 | 막는 결함 |
|---|---|---|
| **A — typed 편집의 seam 라우팅** | `applyTypedEdits`의 `LoadRaw → SetSection → Save` 경로를, `FieldDef.Persist.Key` → yamlpatch 경로 매핑에 따라 `settings.WriteSectionViaSeam`(`internal/settings/sectionwrite.go:62`)으로 교체한다. 기존 스칼라 교체는 `lineSplice`(`internal/settings/yamlpatch/yamlpatch.go:145`)로 대상 행만 재작성 — 주석·미모델링 키·키 순서 원문 보존. upsert(부재 키 신설)만 기존 재직렬화 폴백(C3 한계 승계). | D1의 편집-대상 섹션 손실 |
| **B — 차이-게이트 확장** | (i) seam no-op 게이트에 "부재 키 + `""` 제출 → skip" 분기를 추가한다(`internal/settings/sectionapply.go:68-80` — 현행은 존재-키 동치와 부재-bool만 검사). (ii) "제출 == 영속 → 무기록" 불변식을 전 경로에서 유지한다(기존 REQ-WWS-003 게이트 승계). | D3 (workflow.yaml 빈 키 추가), 무편집 재기록 |
| **백스톱 — Save 더티-게이트 확장** | `ConfigManager.Save`가 user/language/quality/llm 4개를 무조건 재마샬하는 것(`internal/config/manager.go:215-225,263`)을 git-strategy/git-convention 선례(`manager.go:236-260`)와 동일한 `dirty-or-absent` 게이트로 전환한다. seam 라우팅 후에도 `Save()`를 호출하는 잔여 경로(`SyncToProjectConfig` user/language, `writeProjectConfig`, `WriteProjectNestedConfig`)가 미변형 섹션을 재기록하지 못 하는 방어선. | D1의 미편집 섹션 3개 손실 |

D2는 별도 수리(**개정 — plan-audit iter-1 F1**): 초안의 "구조체 복사 + Name 필드만 변경"은 공허한 수리다 — `models.UserConfig`가 `name` 하나만 모델링하므로(`pkg/models/config.go:32-37`) 구조체를 어떻게 조립하든 재마샬 시점(`saveSection`은 원문 병합 없이 `yaml.Marshal`만 수행 — `internal/config/manager.go:462-470`)에 미모델링 키·주석이 소실된다. user 이름 동기화를 **행 단위 치환으로 전환**한다: `SyncToProjectConfig`의 user 섹션 `SetSection` + 전체-교체(`internal/profile/sync.go:27-33`)를 폐기하고, 이름 변경 시 user.yaml의 `name:` 스칼라만 seam 라인-스플라이스(`PatchFile`/`WriteSectionViaSeam`)로 치환한다. user 섹션의 seam 허용 확장(F4)이 전제다.

### §A.2 기각 대안과 사유

- **기각: 순수 B(차이-검출 게이트만, typed Save 유지).** 무편집 Save는 막지만 실제 편집 시 편집-대상 섹션 자체가 여전히 재마샬된다 — 해당 파일의 주석·미모델링 키가 매 편집마다 소실된다. 계약 REQ-WSL-002/003("한 필드 편집 = 그 필드만", "미모델링 키·주석 영생존")을 충족하지 못한다. 이슈 제안 (B)를 단독 채택할 수 없는 이유다.
- **기각: 순수 A(seam 전환만, 게이트 없음).** (1) 동치 제출이 스플라이스-동일-바이트 재기록을 유발한다(동작은 동일해도 REQ-WSL-001의 "기록하지 않는다"를 위반 — mtime 포함). (2) user/language 섹션은 편집 FieldDef가 없어 A의 대상이 아니며, `SyncToProjectConfig`/`writeProjectConfig`의 잔여 `Save()` 호출이 여전히 4파일을 재기록한다. (3) 부재-키 빈 제출(D3)은 막지 못한다.
- **기각: yamlpatch에 노드-삭제/이동 기능 추가.** yamlpatch는 명시적으로 "노드 삭제 미지원"(패키지 헤더)이며, 무손실 계약에 삭제는 불필요하다. 범위 팽창이다.

### §A.3 D4 부분-저장 보증의 성취 가능 수준 (명시적 비(非)목표 포함)

- 보증: 폼 검증(EC-2 atomic reject)·사전-플라이트(대상 파일 판독 가능성) 통과 전에는 어떤 파일도 기록 없음 + per-file 원자 쓰기(`atomicWrite`, `saveSection`).
- 비보증: 다중 파일 트랜잭션 원자성. 뒤 단계 실패 시 앞선 기록은 남으며, 응답이 어디까지 기록됐는지 판독 가능하게 한다(REQ-WSL-008 — 현행 `renderErrorPage` 관행 계승). 500-class 부분 저장의 근원이던 파일-부재 seam은 SPEC-SEAM-GREENFIELD-001/002로 이미 소멸(착지 `e365c2d30`) — 본 SPEC은 회귀 가드만 추가한다.

### §A.4 plan-audit iter-1 결정·판정 기록 (2026-09-29 — 감사 판정은 구속 입력으로 반영)

| # | 결정 | 근거·반영 위치 |
|---|---|---|
| Q1 = F2 | save-time `quality_extras_enabled` 강제-true를 **폐기(RETIRE)**한다. 유지 대안은 기각 — (a) 강제 분기(`internal/settings/sectionapply.go:165-167`)는 M1의 seam 라우팅 교체와 함께 소멸하며, seam 경로에 재현하면 미제출 키를 기록해 REQ-WSL-002 위반이고, (b) 템플릿에 absent 키를 기록하는 것은 구조적으로 AC-WSL-002를 깬다. 마이그레이션(구형 false 잔존 설정의 true 전환)이 필요하면 loader/init 경로의 별도 과제다 — 본 SPEC 밖. | plan §B 해소, M1 판정 게이트 → 실행 항목 전환, M4 전환 목록에 `TestApplySchemaEditsForcesQualityExtrasTrue`(`internal/settings/schema_sections_test.go:285-290`) 포함 |
| Q2 | EmptySubmits 경계는 본문 그대로 **승인**: 키 **존재** 시 `""` 기록(삭제 시맨틱 — `TestCrossSessionEmptySubmitsRoundTrip` 계승), 키 **부재** 시 no-op(D3 수리). | AC-WSL-004 변이 (a)/(b)에 고정 — SPEC 텍스트 변경 없음 |
| Q3 | 부분-저장 보증 수준은 본문 그대로 **승인**: 다중 파일 트랜잭션 원자성은 비목표, 실패 시 판독 가능한 진행 상태 응답. | REQ-WSL-006/007/008 + AC-WSL-006 — SPEC 텍스트 변경 없음 |

## §B Known Issues

- ~~`quality_extras_enabled` 강제-true 유지 여부 미판정~~ — **해소(§A.4 Q1)**: 폐기 결정 확정. 잔여 작업은 M4의 테스트 전환뿐이다.
- yamlpatch 재직렬화 폴백은 빈 줄을 정규화한다 — TestPatchFileValueInvariantPreservesBytes가 이 분기를 감시한다(C3). AC-WSL-002는 이 한계를 변이-별 술어로 분리해 반영한다(plan-audit F3).

## §C Pre-flight

1. `git rev-parse --show-toplevel` → 본 워크트리 확인, `git status --porcelain` → clean 출발.
2. 영향 패키지 기준선 측정: `go test ./internal/settings/... ./internal/config/... ./internal/profile/... ./internal/web/...` (전체 스위트 금지 — CLAUDE.local.md §4).
3. `moai web` 수동 재현 1회(이슈 재현 절차 — scratch project + 미모델링 키·주석 시드)로 D1/D3 현행 재확인 — 보고서 관측을 본 트리에서 1회 독립 측정.

## §D Constraints

- 영속화 규약 경로 강제(C1), `FieldDef` SSOT(C2), yamlpatch 폴백 한계(C3) — spec.md §4 그대로.
- 템플릿 중립성: 본 SPEC의 코드 변경은 `internal/`이므로 템플릿 미러 이슈 없음.
- 커밋 규율: 카드 id + SPEC id 명기, pathspec 스테이징, 커밋 직전 HEAD/브랜치 재판독.

## §E Self-Verification

- [ ] AC-WSL-001..008 전 행이 acceptance.md의 Given-When-Then과 1:1 대응
- [ ] C1 위반(grep: web 패키지 내 `yaml.Marshal`/`os.WriteFile` 신규 호출) 0건
- [ ] REQ-WSL-010: 손실 행위를 단언하던 기존 테스트 목록화 + 전환 증거
- [ ] 영향 패키지 4개 `go test` 통과 출력 인용
- [ ] `golangci-lint run` (CI 판 버전) 통과

## §F Milestones (우선순위 기반 — 변경-가능성 높은 결정 먼저)

### M1 (High) — typed 섹션 편집의 yamlpatch seam 라우팅 [설계 핵심, 가장 변경 가능성 높음]

- `FieldDef.Persist`에 typed 필드의 yamlpatch 경로 노출(또는 `Persist.Key` → dot-path 변환기) — git_strategy(mode, worktree_base_branch, {manual,personal,team}.hooks.pre_push/merge_method), llm(glm.models.{high,medium,low,fable}, glm.effort.*), quality(4개 bool) 전 표면 매핑.
- **seam 허용 확장(plan-audit F4)**: `WriteSectionViaSeam`의 RouteSeam 가드(`internal/settings/sectionwrite.go:56-64`)와 `sectionRootKeys` 맵이 현재 typed 섹션을 "not seam-writable"로 거부한다 — seam-라우팅 대상 섹션(git_strategy/llm/quality)과 D2의 user를 허용 목록에 추가한다.
- `applyTypedEdits`를 `WriteSectionViaSeam` 호출로 교체. `LoadRaw/SetSection/Save` 잔여 사용 여부를 M1 종료 시점에 명시(잔여 시 백스톱이 맡는다).
- D2: `SyncToProjectConfig`의 user 전체-교체를 폐기하고 `name:` 행 스플라이스로 전환(§A.1 개정).
- 검증: 한 필드 편집 diff가 그 행만(AC-WSL-002 line-splice 변이), 미모델링 키·주석 생존(AC-WSL-003), user 이름 편집 행 고립(AC-WSL-005).
- **판정 게이트(해소 — §A.4 Q1)**: `quality_extras_enabled` save-time 강제는 **폐기**한다 — M1의 교체가 `applyTypedEdits:165-167` 분기를 재현하지 않음으로써 실행된다. 마이그레이션이 필요해지면 loader/init 경로 별도 과제로 발의한다.

### M2 (High) — 차이-게이트 확장 [D3 수리]

- seam no-op 게이트에 부재-키 + `""` 제출 skip 분기 추가(`internal/settings/sectionapply.go:68-80` 인접).
- EmptySubmits 필드의 "삭제" 시맨틱 보존 확인: 키가 **존재**하는 상태에서의 `""` 제출은 계속 `""`로 기록(설계 의도 — crosssession 라운드트립 테스트 `TestCrossSessionEmptySubmitsRoundTrip` 유지)되고, **부재** 상태에서만 no-op다. 두 분기의 AC-WSL-004 변이 테스트.
- 검증: workflow.yaml audit pin 빈 키 추가 소멸.

### M3 (Medium) — Save 더티-게이트 확장 + user 전체-교체 수리 [백스톱]

- `ConfigManager.Save`의 user/language/quality/llm을 `dirty-or-absent` 게이트로 전환(`internal/config/manager.go:215-225,263`). `SetSection`에서 더티 플래그 설정, `Save` 성공 시 초기화 — git-strategy 선례 복제.
- `internal/profile/sync.go:28` 전체-교체를 로드-복사 + Name만 변경으로 수정.
- greenfield(파일 부재) 생성은 게이트의 absent 분기로 유지 — 초기화 흐름 회귀 없음 확인.
- 검증: 잔여 `Save()` 호출 경로에서도 미편집 섹션 4파일 불변(AC-WSL-001 백스톱 변이).

### M4 (Medium) — 부분-저장 사전-플라이트 + 회귀 가드 [D4]

- `ApplySchemaEdits`/`WriteSectionViaSeam` 계약에 "첫 기록 전 대상 파일 판독 가능성 검증" 추가(PatchFile은 이미 부재-허용이므로 판독 불능은 파손 파일 케이스만).
- 회귀 가드 테스트: 파일-부재 seam 저장이 200으로 완결(e365c2d30 착지 고정), 검증-실패 시 디스크 무변경(EC-2 고정), 기록 실패 시 오류 응답에 진행 상태 표기(REQ-WSL-008).
- 손실-행위 회귀 테스트 정리(REQ-WSL-010): 기존 테스트 중 손실을 단언/허용하는 케이스 목록화 → 무손실 단언으로 교체. **§A.4 Q1 전환 포함**: `TestApplySchemaEditsForcesQualityExtrasTrue`(`internal/settings/schema_sections_test.go:285-290`)의 강제 단언을 "강제 분기 부재" 단언으로 교체한다(quality 편집 Save가 미제출 `quality_extras_enabled` 키를 기록하지 않음).

### M5 (Medium) — 재검증·마감

- 영향 패키지 4개 `go test`, `golangci-lint run`(CI 판 버전), 이슈 재현 절차 재실행으로 AC-WSL-001(전 섹션 byte-identical) 실측.
- `moai spec lint SPEC-WEB-SAVE-LOSSLESS-001` 0 findings 유지.

## §G Anti-Patterns

- 웹 레이어에서 YAML 직접 marshal/쓰기(REQ-WC3-008 위반) — 금지.
- "diff가 깨끗해 보이니 괜찮다" — 동치 제출의 스플라이스 재기록도 REQ-WSL-001 위반이다(mtime 포함 무기록).
- 폴백 재직렬화 경로를 테스트 없이 남겨두기 — upsert 폴백이 빈 줄을 정규화하는 한계는 기존 invariant 테스트로 고정돼 있다(C3).
- git-strategy 더티-게이트를 "참고"만 하고 플래그 초기화(EC-3)를 빠뜨리는 것 — 선례의 초기화 계약까지 복제한다.

## §H Cross-References

- spec.md §1.3 결함-근거 표, §A 설계 결정 본문.
- acceptance.md §D AC 매트릭스.
- `.moai/reports/t1314/` — plan-audit 보고 착지 예정 경로(본 단계에서 생성하지 않음).
