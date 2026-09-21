# t339 verdict — SPEC-AGENT-EMIT-LINEAGE-001 본문 정확성 3건(t317 반입 부채) 수리

Card: t339 (G2a, lane-9) · Branch: WT-spec-body-cites · Base: local develop `48c35a4d4` (흡수 시점 develop 팁; live develop이 그 뒤 3커밋 앞센 것은 SPEC 본문 파일과 무교집합 — 창 때 재흡수) · 본문 편집: manager-spec 스폰 2라운드 (레인 검증 아래)

## Claim

t317 plan-audit iter-3이 남긴 SPEC 본문 결함 3건을 수리했다: D10(§B.1 전수 열거가 골든 고정본 3본 누락), D11(doctor 규약 과대 서술), D12(골든 테스트 좌표 오기 + 파일 모호성). 세 건 모두 intake 재측정에서 이 트리에서 여전히 참으로 확인된 뒤 수리됐고, 결함의 도달 범위 안에 있던 §B.2 수치 2건(파일 수 5, 하한 가정 6-7건)도 함께 재계산했다 — 같은 문서 :55가 "위 B.1을 열거로 유지한다"고 자기 규약을 명시하므로 열거-수치 불일치를 남기면 이 SPEC의 자기 교훈(D8 계열)을 재연산하는 형태가 되기 때문.

## Evidence

**Intake 재측정 (레인, 이 트리 `48c35a4d4`, 수리 전):**

1. D10 참: §B.1 표 5행 + "**5건.**" 단락(:23) — 골든 3본 `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` 부재. 3본 모두 존재 실측 (각 7842 bytes). 원 발견 t317 iter3:227.
2. D11 참: plan.md:18 "이 리포의 doctor 항목은 파일 1개에 사는 것이 규약이다", :23 "파일 1개 + 짝 테스트 1개로 살고" — 원 발견 t317 iter3:229 (13개 그룹 중 자기 파일 2건, `doctor.go` 인라인 6건).
3. D12 참: spec.md:88(t317이 :86으로 인용했던 블록쿼트 — 발견 자체가 좌표 드리프트한 사례)와 plan.md:77의 인용 "`golden_test.go:285` 의 `if count != 11`" — 실측 `if`는 `:284`, `want 11`은 `:285` (`internal/template/agentemit/golden_test.go`); `internal/codexadapter/golden_test.go`도 존재해 파일명 모호. plan.md:142의 `:80`(골든 본체 `TestGoldenCommittedArtifactsMatchEmission`)·`:255`(`TestEmbedFSPresenceAndByteEquality`) 인용은 이 트리에서 정확 판정 → 미수정.

**수리 후 레인 독립 검증 (이 실행, 이 트리):**

- `git diff --stat` → 정확히 2 파일 (plan.md +9/-6, spec.md +1/-1), frontmatter·다른 산출물 무변경.
- 낡은 형태 스윕 5종 전부 0히트: "5건", "여섯 번째 편집 파일", "규약이다", "6-7 건", "golden_test.go:285" (양 파일).
- diff 본문 직독 양성 확인: 골든 3행 추가(6·7·8행, 7842 bytes + 측정 SHA 병기), "5건"→"8건", 테스트 소스/고정본 구분 문장 교체, D11 양 지점 두-선례 서술(자기 파일 2개·인라인 6개·측정 SHA, `doctor_hook.go` 근거 제거 — own-file 2건에 미포함 실측 반영), D12 좌표 `internal/template/agentemit/golden_test.go:284-285` + 측정 SHA (plan·spec 양쪽), §B.2 "파일 수 8 / `8 < 5` 거짓 / 밴드 5-15 안(측정: 48c35a4d4)"·하한 가정 "9-10 건" 재기준. 판정 불변(M).

## Baseline-attribution

모든 측정 이 실행·이 트리 (WT-spec-body-cites): intake 3건의 존재·좌표 (수리 전 관측), 스윕 5종 카운트, diff --stat, diff 본문 직독. 수정된 좌표 인용에는 측정 트리 SHA(`48c35a4d4`)가 문서 내에 병기됨 (카드 규칙: 수정 인용에 측정 트리 SHA 기록).

## Gaps (명시적으로 관측하지 않은 것)

- `internal/cli/testdata/` 골든 3본의 **내용**은 재생성·비교하지 않았다(이 카드는 본문 정확성 수리이며 run-phase의 산출물이 아님) — 존재·크기만 관측.
- golangci-lint·go test 미실행 — 변경은 `.moai/specs/` 마크다운 2파일이고 Go 코드 무접촉.
- CI 전체 스위트 판정은 리드 배치 push 소관.
- row 3의 "모든 doctor 항목이 짝 테스트를 갖는다" 서술은 t317이 발견하지 않았고 이 카드가 측정하지 않아 미판정 — 원문 보존.

## Residual-risk

- plan.md:142의 `:80`·`:255` 인용은 이 트리에서 정확했으나 이동 좌표다 — 다음 트리에서 드리프트할 수 있으며 이 카드의 측정 SHA가 그 판독 기준점이 된다.
- live develop이 흡수 후 3커밋 앞섬(비교집합 무교집합) — 창 병합 시 재흡수 절차가 흡수한다.
