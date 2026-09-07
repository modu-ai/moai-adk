# Sync-Audit — SPEC-WEB-WRITE-SAFETY-001 (card t517)

- 감사자: sync-auditor (독립 판정 — 실행 레인과 별개 판정 주체)
- 감사 시점: 2026-09-07, 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t517` (branch `WT-web-write-safety`, HEAD `43f9ab202`)
- 검토 범위: `911d9bbcc..43f9ab202` (base `2031ccf7f`) — M1-M3 측정 재현 + RED-first 테스트, M4 쓰기-안전 수리, M5 가드 채택 + 뮤턴트, 커버리지 보강, sync 아티팩트, sync_commit_sha 백필
- 렌즈: --security, --deep (리드 지정) + sync 위생

---

## Evaluation Report

SPEC: SPEC-WEB-WRITE-SAFETY-001
Overall Verdict: **FAIL** (차단 결함 1건 — 하단 F1. 수리 후 결함 델타 재감사 대상)

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 0.75 | **FAIL** (차단 F1) | AC-WWS-001..008 전부 PASS — `go test -count=1 ./internal/settings/... ./internal/web/... ./internal/config/...` → `ok` ×7 (exit 0, 본 런 본 트리). RED 로그 5가드 FAIL(`2031ccf7f`) + 뮤턴트-B 본 감사 독립 재실행 → RED 확인. 단, 수리 (i)(b)가 default-on-when-absent bool의 명시 OFF 저장을 무음 스킵 — F1 참조 |
| Security (25%) | 1.00 | PASS | 원자 reject가 모든 쓰기에 선행(handlers.go:434-462, fail-closed), lineSplice 재파싱 검증이 의심 시 전량 폴백(yamlpatch.go:184-210), 섹션명은 스키마 통제(path traversal 없음), 시크릿 무접촉. OWASP Top 10 적용 — 신규 보안 발견 0 |
| Craft (20%) | 0.75 | **FAIL** (프로필 하드 기준) | 프로필 하드 기준 "Coverage < 85% = Craft FAIL": yamlpatch 81.1%, web 66.8% (패키지 단위, 본 런 실측 `go test -cover`). settings 90.2% ✓. 신규 함수는 83-90% (ApplySchemaEdits 88.2, lineSplice 86.9, PatchFile 88.5, applyTypedEdits 83.8, readSeamScalar 87.0, handleSave 90.2, parseSchemaForm 79.5). web 미달은 기존 baseline 귀속(본 SPEC은 테스트만 추가), yamlpatch 미달은 기존 error-path(atomicWrite 59.1%) 중심 — delta 귀속 결손 아님. `golangci-lint run --timeout=2m ./internal/settings/... ./internal/web/...` → `0 issues.` |
| Consistency (15%) | 0.75 | PASS | 기존 게이트·주석·파일 관례 부합, @MX:ANCHOR fan_in 실측 5호출점/3파일 일치, frontmatter 전환 순수(status 1행). 사소한 이탈: progress.md 기록-코드 불일치 2건(F2, F3) |

가중 평균 81.3 / 조화 평균 ≈ 0.80 — 수치는 통과대이나, 차단 결함(F1)과 Craft 하드 기준에 따라 **FAIL** 판정한다(`.moai/config/evaluator-profiles/default.md` Hard Thresholds + finding-consumption 규율: correctness에 영향하는 발견은 blocking).

### Findings

- **F1 [blocking] [blocking]** `internal/settings/sectionapply.go:64-70` — 수리 (i)(b)의 "absent IS false for a bool key" 가정은 **default-on-when-absent bool에서 거짓**이고, 그 결과 사용자의 명시적 OFF 저장이 무음으로 스킵된다. 피해 필드: `workflow.todo.enabled`(`TodoEnabled()` absent⇒**enabled**, `internal/config/todo_enabled.go:29-34`; 템플릿이 todo 블록을 배포하지 않아 거의 모든 사용자가 해당 상태)와 `mcp.tools.<name>.enabled` 24종(`readMCPToolEnablement` fail-OPEN all-enabled, `internal/cli/mcp_server.go:527-566`). 연쇄: `SchemaCurrentValues`는 부재 키를 `""`로 반환(sectionvalues.go:52) → 콘솔은 absent를 **OFF로 렌더**(본 감사 하베스트 실측: `workflow.todo.enabled=''` 제출)하면서 런타임은 enabled → 사용자가 OFF 저장 → `edits="false"` → `readSeamScalar` ok=false → 스킵 → **아무것도 기록되지 않고 런타임은 enabled 유지**. 콘솔이 보여주는 상태로 영구 수렴 불가(우회는 ON 저장 후 OFF 저장의 2단계 — 문서화 없음). 스키마 자체가 극성을 문서화한다: "this key is default-ON, so the console's 'absent' rendering means enabled"(`schema_sections.go:360-365`). 커밋된 테스트는 default-off 필드(gate)만 고정 — default-on 극성 무보호. **수리 방향**: 스킵 판정을 유효 기본값 인지로 — `FieldDef`에 유효 기본값(또는 absent 해석)을 명시하고, absent && 제출값 == 유효 기본값 → 스킵, absent && 제출값 ≠ 유효 기본값 → 기록. + RED-first 테스트 2건: (a) absent `todo.enabled` + `"false"` 제출은 `enabled: false`를 기록해야 함(현재 RED), (b) absent `todo.enabled` + `"true"` 제출은 스킵해야 함(키 생성 방지 — 현재는 기록함, 이건 본래 무해하지만 M1(d) 서명과 정합).
- **F2 [minor] [F1-class repair]** `.moai/specs/SPEC-WEB-WRITE-SAFETY-001/progress.md` §E.2 M4 표 (iv)행 + §E.2 E1 행렬 AC-WWS-006 — "동의 중복은 통과, 불일치 중복은 reject"로 기술했으나 커밋된 코드는 **동의 여부 무관 모든 중복을 reject**(`schemaform.go:325-332`, `len(vals) > 1` 무조건)이고, 커밋된 테스트도 동일 값 2회 제출의 reject를 고정한다(`TestParseSchemaFormDuplicateCompanionRejected`: `"1"`×2 → reject 기대). CHANGELOG 서술("rejects a form name submitted more than once")은 코드와 일치 — progress.md가 유일한 이탈. 수리: progress.md 두 지점을 "모든 중복 reject"로 정정.
- **F3 [minor] [F1-class repair]** `progress.md` §E.4 `b12_self_test_b_ac_count_match` — "acceptance.md 8건 == CHANGELOG 엔트리 인용 8건"으로 기록했으나, 기록된 동일 명령(`grep -oE 'AC-WWS-[0-9]+' | sort -u`)의 재측정 결과는 acceptance.md **8** / CHANGELOG **1** (엔트리가 범위 표기 `AC-WWS-001..008`로 인용 — 토큰 1개). 실체는 만족(엔트리 1건, SPEC-ID 중복 0 — `grep -c` = 1, 8 AC 전체가 범위로 지칭)이나 기록된 측정 등식은 성립하지 않는다. 수리: CHANGELOG 엔트리에 AC를 열거하거나 §E.4 기록을 실측 형태로 정정.
- **F4 [advisory] [optional]** `progress.md` §E.2 M1(d)/M2 — workflow 부수 손상의 기제 귀속 "중복 폼값 첫-값 채택"은 커밋된 도구 구조상 성립하지 않는다. 본 감사 실측: 렌더 페이지의 동명 name 54건 전부 radio 세그먼트 그룹(브라우저는 그룹당 checked 1개만 제출), 커밋된 `build_post.py`를 동일 페이지에 재실행하면 121필드 **중복 이름 0건**. "false 폴백" 귀속은 지지됨(OFF 라디오가 `name=''` 제출 → bool 파서가 "false" → absent-key upsert가 키 생성 — 하베스트 28-29행 실측). `effort: high→값 소실` 2건의 기제는 재관측 불가(선결함 트리 없음)로 미해결. REQ-WWS-006 가드 자체는 스키마 요구로 정당하나 기제 서사는 정정 대상.
- **F5 [advisory] [optional]** (codex 2차 소견 중 검증된 것) typed 섹션 하나의 실변경 시 `Save()`가 전 typed 파일을 재기록 — 본 리포 코퍼스에서는 byte-동일 round-trip으로 무해(M1(d) 실측), 그러나 주석·미모델링 키를 가진 typed 형제 파일에서는 손실 가능(Save()의 기존 의미론 — SPEC §3이 typed-path 재설계를 범위 밖으로 흡수). 후속 카드 후보.
- **F6 [advisory] [optional]** yamlpatch upsert 경로는 여전히 전체 재직렬화 폴백 → 빈 줄을 가진 파일에 신규 키를 만드는 저장은 REQ-WWS-005 충실도를 충족하지 못함(문서화된 폴백; AC-WWS-005의 When은 기존 스칼라 교체만 커버 — AC 자체는 PASS).
- **F7 [optional]** 중복 reject가 스키마 필드에만 적용됨 — `__profile`/프로필/nested/agentfm/glmkey 파서는 여전히 첫 값 채택(handlers.go:361 등). SPEC §3.1은 schemaform.go 시설로 범위 한정했으므로 AC 위반은 아니며, `ParseForm()` 직후 중앙 검사가 후속 경화 후보.
- **F8 [optional]** Craft 하드 기준의 귀속 명시 — 본 항목은 F1과 별도로 리드 판단용: yamlpatch/web의 패키지 커버리지 미달은 delta가 만든 것이 아니다(본 SPEC은 테스트만 추가; web 66.8%는 기존 baseline). base 트리 직접 측정은 불가했음(Gaps 참조).

### Per-Lens Outcomes

- **--security**: PASS — reject 경로는 fail-closed(모든 검증기가 첫 쓰기 전 실행, 400 렌더 후 return), splice 재파싱 검증이 손상 파일 기록을 차단(불일치 시 폴백), 폼값→YAML 스칼라 주입은 재파싱+인코더 폴백으로 경계됨, 경로는 스키마 통제 섹션명만 사용. 신규 보안 결함 0.
- **--deep**: 5항목 중 4항목 독립 검증 성립 — (1) 값-불변 조기 종료 경로 추적 완료(seam skip/typed 게이트/Save 스킵), (2) typed DeepEqual은 값 복사 스냅샷 비교로 전 필드 커버(양성 통제 2건이 실변경 생존 증명), (3) splice는 커버 케이스에서 byte 보존(테스트 + 뮤턴트-B 재실행으로 확인), upsert 폴백의 영향 반경은 문서화됨, (4) AC-WWS-004 양성 통제 2건(설정레이어+전체 경로)이 gate 비실행과 구별됨, (5) MUTANT-B 독립 재실행 → 두 PatchFile 가드 즉시 RED(빈 줄 2→1, O1 서명). MUTANT-A/C는 커밋된 RED 로그의 선결함 FAIL이 동등 상태를 커버. **단 deep 추적이 F1을 발견** — 값-불변 게이트의 absent 분기가 default-on 극성에서 "변경이지만 스킵"으로 반전됨.
- **sync 위생**: close 커밋이 마지막 코드 쓰기 ✓(`git diff --name-only e94d2f2b4..43f9ab202` → progress.md만), B12 중복 엔트리 0 ✓ / 파일 경로 존재 ✓ / AC 계수 기록 불일치(F3), frontmatter 전환은 sync 커밋 탑승 + 본문 무변경 ✓(`git show e94d2f2b4` spec.md diff = status 1행), D5 괄호 존속은 감사 재가 범위, MX @MX:ANCHOR fan_in 5호출점/3파일 실측 일치 ✓.

### Gaps (관측하지 못한 것)

- 수정 바이너리로의 **실물 서버** M1(d) 절차 재실행은 하지 않았다 — in-process 전체 경로 테스트(`TestHandleSave*`)가 동일 코드 경로를 커버하고 본 감사의 오버레이 계측(렌더 125이름/54 radio 그룹, 하베스트 121필드/중복 0)을 수행했으나, 루프백 서버 기동~POST~diff의 실물 재귀는 미관측.
- base 트리(`2031ccf7f`)의 커버리지 미측정 — web 66.8%의 baseline 귀속은 구조적 추정(테스트 추가만 있음)이지 측정이 아니다.
- M1(d)의 effort 스칼라 소실 기제 미해결 — 선결함 트리 재관측 없이는 판별 불가(F4).
- codex 백엔드는 binary lag(e79c010b8 < HEAD) 상태에서 실행됐고 그 FAIL 소견 4건 중 F1(=codex-1)만 본 감사가 독립 검증해 채택했다. codex-2/3/4는 본 감사가 코드로 검증해 advisory로 강등했다. glm 백엔드 inconclusive(대상 미해석).
- `origin/develop` CI 전체 스위트 판정은 미관측 — develop 병합·push는 리드 소관이며 그 판정은 본 보고서 밖이다.

### Residual Risk

- F1 수리 후에도 absent default-on bool의 **렌더 극성**(부재 ⇒ OFF 표시 vs 런타임 enabled)은 기존 표시-런타임 발산으로 남는다 — 후속 카드로 렌더를 유효 기본값으로 시딩하는 정렬 권장.
- 중복-reject 가드의 안전성은 "렌더 페이지가 radio 그룹 외 중복 name을 갖지 않는다"는 현재 구조에 의존한다 — 렌더 유일성 고정 테스트(렌더→name 카운트)가 없으면 향후 템플릿 변경이 정상 저장을 reject할 수 있다.
- upsert 폴백의 충실도 갭(F6)과 typed 형제 재기록(F5)은 문서화된 한계로 잔존한다.

---

## 판정 요약

8개 AC는 모두 실증적으로 PASS이고 수리 5건 자체는 올바르게 설계·검증됐다. 그러나 수리 (i)(b)가 **default-on-when-absent bool의 명시적 변경을 무음으로 삼키는 새 정확성 결함**을 도입했고(콘솔의 가장 일반적 상태 — todo 블록 부재 — 에서 토글이 무효화됨), 이것은 "수리가 원래 결함과 같은 형태(사용자 의도와 무관한 디스크 상태)"를 부분적으로 재도입한 것이다. 차단 결함 1건의 결함 델타 수리 + progress.md 기록 정정 2건 후, 그 델타 범위의 재감사로 PASS 판정이 가능하다.
