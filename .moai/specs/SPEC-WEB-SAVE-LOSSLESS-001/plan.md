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

D2는 별도 수리: `internal/profile/sync.go:28`의 user 섹션 전체-교체(`cfg.User = models.UserConfig{Name: ...}`)를 "로드된 구조체 복사 + Name 필드만 변경"으로 수정한다. (user 섹션에 seam 쓰기 경로가 없으므로 A의 직접 대상이 아니고, 백스톱만으로는 전체-교체 자체가 막히지 않는다.)

### §A.2 기각 대안과 사유

- **기각: 순수 B(차이-검출 게이트만, typed Save 유지).** 무편집 Save는 막지만 실제 편집 시 편집-대상 섹션 자체가 여전히 재마샬된다 — 해당 파일의 주석·미모델링 키가 매 편집마다 소실된다. 계약 REQ-WSL-002/003("한 필드 편집 = 그 필드만", "미모델링 키·주석 영생존")을 충족하지 못한다. 이슈 제안 (B)를 단독 채택할 수 없는 이유다.
- **기각: 순수 A(seam 전환만, 게이트 없음).** (1) 동치 제출이 스플라이스-동일-바이트 재기록을 유발한다(동작은 동일해도 REQ-WSL-001의 "기록하지 않는다"를 위반 — mtime 포함). (2) user/language 섹션은 편집 FieldDef가 없어 A의 대상이 아니며, `SyncToProjectConfig`/`writeProjectConfig`의 잔여 `Save()` 호출이 여전히 4파일을 재기록한다. (3) 부재-키 빈 제출(D3)은 막지 못한다.
- **기각: yamlpatch에 노드-삭제/이동 기능 추가.** yamlpatch는 명시적으로 "노드 삭제 미지원"(패키지 헤더)이며, 무손실 계약에 삭제는 불필요하다. 범위 팽창이다.

### §A.3 D4 부분-저장 보증의 성취 가능 수준 (명시적 비(非)목표 포함)

- 보증: 폼 검증(EC-2 atomic reject)·사전-플라이트(대상 파일 판독 가능성) 통과 전에는 어떤 파일도 기록 없음 + per-file 원자 쓰기(`atomicWrite`, `saveSection`).
- 비보증: 다중 파일 트랜잭션 원자성. 뒤 단계 실패 시 앞선 기록은 남으며, 응답이 어디까지 기록됐는지 판독 가능하게 한다(REQ-WSL-008 — 현행 `renderErrorPage` 관관 계승). 500-class 부분 저장의 근원이던 파일-부재 seam은 SPEC-SEAM-GREENFIELD-001/002로 이미 소멸(착지 `e365c2d30`) — 본 SPEC은 회귀 가드만 추가한다.

## §B Known Issues

- `quality_extras_enabled` 강제-true(`internal/settings/sectionapply.go:165-167`)는 문서화된 의도 동작이다 — 무손실 대상에서 제외하되, quality 섹션을 seam으로 라우팅한 뒤에도 이 강제가 유지되는지 M1에서 명시적 판정이 필요하다.
- yamlpatch 재직렬화 폴백은 빈 줄을 정규화한다 — TestPatchFileValueInvariantPreservesBytes가 이 분기를 감시한다(C3).

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
- `applyTypedEdits`를 `WriteSectionViaSeam` 호출로 교체. `LoadRaw/SetSection/Save` 잔여 사용 여부를 M1 종료 시점에 명시(잔여 시 백스톱이 맡는다).
- 검증: 한 필드 편집 diff가 그 행만(AC-WSL-002), 미모델링 키·주석 생존(AC-WSL-003).
- **판정 게이트**: `quality_extras_enabled` 강제 동작의 승계 방식 결정(§B 첫 항) — 이 결정이 M1의 착지 조건이다.

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
- 손실-행위 회귀 테스트 정리(REQ-WSL-010): 기존 테스트 중 손실을 단언/허용하는 케이스 목록화 → 무손실 단언으로 교체.

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
