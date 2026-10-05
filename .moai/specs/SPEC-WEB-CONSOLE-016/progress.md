# SPEC-WEB-CONSOLE-016 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set; `design.md` / `research.md` not emitted — Tier L only).
- Status on creation: `draft`. No production code written in plan phase.
- Requirement IDs: REQ-WC16-001 .. REQ-WC16-011. Acceptance IDs: AC-WC16-001 .. AC-WC16-011.
- Evidence base: `.moai/reports/t1049/findings.md` + `internal/web/partial_apply_repro_test.go` (commit `a52ce60f0`), measured on tree `WT-partial-apply` @ `d8304b49a`. No figure was carried over from the t1043 tree.
- Open clarifications blocking Implementation Kickoff Approval: three, enumerated in `plan.md` §F.1.
- Run-phase entry is NOT approved. Axis (4) changes user-visible behaviour and is the operator's decision at the Kickoff gate.

## §E.2 Run-phase Evidence

**이 run-phase 는 SPEC 을 구현하지 않았다.** 운영자 판정(2026-09-21, 리드 경유)으로 카드 t1049 의 범위가 **조사 전용**으로 좁혀졌다 — 산출물은 (1) 재현 계측기 (2) 세 영속화 층 각각의 원자성 측정값 둘뿐이고, 트랜잭션 설계·구현은 카드 밖이다. 따라서 `spec.md` 의 `status` 는 **`draft` 로 남는다**: 이 SPEC 이 서술하는 수리는 착수되지 않았고, 향후 수리 판정의 입력으로 대기한다. 구현하지 않은 작업에 `in-progress → implemented` 이력을 붙이지 않는다.

### 산출물 (1) 재현 계측기 — 착지

`internal/web/partial_apply_repro_test.go` (커밋 `a52ce60f0`). 주입 가능한 6단계 각각에서 앞선 단계가 이미 디스크에 있음을 관측. 7·9단계는 패키지 함수라 주입 seam 이 없어 미측정(§6-1, 부재 아님).

### 산출물 (2) 층별 자체 원자성 — 측정 완료

계측기: `internal/settings/layer_atomicity_probe_test.go`. 실패 주입 방식은 **대상 경로를 비어 있지 않은 디렉터리로 치환**하는 것이다 — 읽기 전용 파일로는 부족하다(두 층이 temp+rename 이라 디렉터리가 쓰기 가능하면 rename 이 성공해 프로브가 조용히 성공을 측정한다).

| 층 | 다중 파일 원자성 | 실측 |
|---|---|---|
| L1 profile store (`WritePreferences`) | **해당 없음** — `os.WriteFile` 한 번, 파일 하나 | 실패한 쓰기가 **기존 내용을 파괴하지 않는다**(0444 차단, 이전 내용 그대로). temp+rename 은 아니다 |
| L2 config manager (`ConfigManager.Save`) | **없음** | `language.yaml` 차단 시 `Save` 가 실패하고 **`user.yaml` 은 이미 디스크에 있다** |
| L3 yamlpatch (`ApplySchemaEdits`) | **없음** (단일 `PatchFile` 은 temp+rename 으로 파일 단위 원자적) | `workflow` 차단 시 루프가 실패하고 **`feedback.yaml` 은 이미 패치돼 있다** |

**결론**: 세 층 모두 **파일 단위까지만** 원자적이고(L1 은 그조차 아니다), 여러 파일을 도는 경로는 전부 **첫 실패에서 early return 하는 루프**다. `handleSave` 의 9단계 부분 적용은 그 층에 고유한 성질이 아니라 **각 층에서 한 번 더 되풀이되는 모양**이다.

- 증거: `.moai/reports/t1049/layer_atomicity.log` (EXIT 0) · 회귀 `.moai/reports/t1049/regress.log` (settings/web/profile/config 4패키지 EXIT 0)
- 세 프로브 모두 **양성 대조 동반**. L1·L3 은 대조가 판별 불가를 검출하면 `t.Skip` 으로 멈추도록 만들었고, 실제로 L3 의 첫 시도가 그렇게 멈춰 아래 부수 발견을 끌어냈다.

### 부수 발견 — 스키마의 `PersistSeam` 과 실제 seam-writable 이 다르다

`harness` 섹션은 스키마상 `Persist.Kind == PersistSeam` 인데 `RouteForSection("harness") != RouteSeam` 이라 쓰기가 거부된다(`section "harness" is not seam-writable`). **값이 바뀌지 않는 제출에서는 이 불일치가 드러나지 않는다** — 쓰기 자체가 시도되지 않아 라우트 검사에 도달하지 않기 때문이다. 프로브가 값을 뒤집자 발화했다. 이 카드의 범위가 아니므로 처분하지 않고 기록만 남긴다.

### 재현하지 않은 것

agent-6 의 워크트리 `probe-layer-atomicity` 실행 관측(`llm.yaml` 차단 시 셋이 이미 변경)은 **본 레인이 재현하지 않았다.** 다만 L2 프로브가 이 트리에서 같은 성질을 독립적으로 관측했다 — 차단 대상이 다르고(`language.yaml`) 관측 파일도 다르다(`user.yaml`).

## §E.3 Run-phase Audit-Ready Signal

- run_status: `investigation-complete` (구현 없음 — 위 §E.2 의 범위 축소 참조)
- 커밋: `a52ce60f0`(계측기 1) · `bf44a147d`(SPEC) · `1bc32e875`(측정값 반영) + 본 커밋(계측기 2 + 본 기록)
- 프로덕션 코드 변경: **0줄**. 추가된 것은 테스트 파일 2개와 SPEC/증거 문서뿐이다.
- 남은 Gap(부재 아님): §6-1 7·9단계 실패 거동(파일시스템 탐침 필요) · §6-3 계측 대리값(seam 호출 관측) · §6-4 실패 경로 nested 렌더 · 브라우저 표시.
- 수리 여부 판정은 **리드/운영자 소관**이며 이 카드가 선취하지 않는다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
